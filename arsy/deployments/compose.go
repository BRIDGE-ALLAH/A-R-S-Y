package deployments

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	serviceNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`)
	volumeNamePattern  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}$`)
	environmentKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// ValidationError identifies the exact Compose field that violates MVP-1 policy.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ComposeManifest is the validated, non-secret structural representation of
// an MVP-1 Compose file. Environment values are deliberately not retained.
type ComposeManifest struct {
	Services map[string]ComposeService
	Volumes  map[string]struct{}
	Networks map[string]struct{}
}

type ComposeService struct {
	Image           string
	PublishedPorts  []string
	EnvironmentKeys []string
	Volumes         []string
	Networks        []string
	Restart         string
}

// ValidateCompose parses YAML structurally and enforces the supported MVP-1
// subset. It never interprets user input as a shell command.
func ValidateCompose(input string) (ComposeManifest, error) {
	var root map[string]any
	if err := yaml.Unmarshal([]byte(input), &root); err != nil {
		return ComposeManifest{}, &ValidationError{Field: "manifest", Message: "must be valid YAML"}
	}
	if len(root) == 0 {
		return ComposeManifest{}, &ValidationError{Field: "manifest", Message: "must not be empty"}
	}
	for key := range root {
		switch key {
		case "services", "volumes", "networks", "name", "version":
		default:
			return ComposeManifest{}, &ValidationError{Field: key, Message: "is not supported in MVP-1"}
		}
	}

	manifest := ComposeManifest{
		Services: make(map[string]ComposeService),
		Volumes:  make(map[string]struct{}),
		Networks: make(map[string]struct{}),
	}
	if err := collectResourceNames(root["volumes"], "volumes", volumeNamePattern, manifest.Volumes); err != nil {
		return ComposeManifest{}, err
	}
	if err := collectResourceNames(root["networks"], "networks", volumeNamePattern, manifest.Networks); err != nil {
		return ComposeManifest{}, err
	}

	rawServices, ok := root["services"].(map[string]any)
	if !ok || len(rawServices) == 0 {
		return ComposeManifest{}, &ValidationError{Field: "services", Message: "must contain at least one service"}
	}
	for name, rawService := range rawServices {
		field := "services." + name
		if !serviceNamePattern.MatchString(name) {
			return ComposeManifest{}, &ValidationError{Field: field, Message: "has an invalid service name"}
		}
		fields, ok := rawService.(map[string]any)
		if !ok {
			return ComposeManifest{}, &ValidationError{Field: field, Message: "must be a mapping"}
		}
		for key, value := range fields {
			switch key {
			case "image", "ports", "environment", "volumes", "networks", "restart":
			case "build":
				return ComposeManifest{}, &ValidationError{Field: field + ".build", Message: "is not supported; use an existing image"}
			case "privileged":
				if enabled, _ := value.(bool); enabled {
					return ComposeManifest{}, &ValidationError{Field: field + ".privileged", Message: "must not be enabled"}
				}
			case "network_mode":
				if value == "host" {
					return ComposeManifest{}, &ValidationError{Field: field + ".network_mode", Message: "host networking is not allowed"}
				}
				return ComposeManifest{}, &ValidationError{Field: field + ".network_mode", Message: "is not supported in MVP-1"}
			case "pid", "ipc":
				if value == "host" {
					return ComposeManifest{}, &ValidationError{Field: field + "." + key, Message: "host namespaces are not allowed"}
				}
				return ComposeManifest{}, &ValidationError{Field: field + "." + key, Message: "is not supported in MVP-1"}
			case "devices":
				return ComposeManifest{}, &ValidationError{Field: field + ".devices", Message: "device access is not allowed"}
			default:
				return ComposeManifest{}, &ValidationError{Field: field + "." + key, Message: "is not supported in MVP-1"}
			}
		}

		image, ok := fields["image"].(string)
		if !ok || strings.TrimSpace(image) == "" || strings.ContainsAny(image, "\r\n\t ") {
			return ComposeManifest{}, &ValidationError{Field: field + ".image", Message: "must be a non-empty image reference"}
		}
		service := ComposeService{Image: strings.TrimSpace(image)}
		if service.PublishedPorts, err = parsePublishedPorts(fields["ports"], field+".ports"); err != nil {
			return ComposeManifest{}, err
		}
		if service.EnvironmentKeys, err = parseEnvironmentKeys(fields["environment"], field+".environment"); err != nil {
			return ComposeManifest{}, err
		}
		if service.Volumes, err = parseServiceVolumes(fields["volumes"], field+".volumes", manifest.Volumes); err != nil {
			return ComposeManifest{}, err
		}
		if service.Networks, err = parseServiceNetworks(fields["networks"], field+".networks", manifest.Networks); err != nil {
			return ComposeManifest{}, err
		}
		if rawRestart, exists := fields["restart"]; exists {
			restart, ok := rawRestart.(string)
			if !ok || !validRestartPolicy(restart) {
				return ComposeManifest{}, &ValidationError{Field: field + ".restart", Message: "must be no, always, on-failure, or unless-stopped"}
			}
			service.Restart = restart
		}
		manifest.Services[name] = service
	}
	return manifest, nil
}

func collectResourceNames(raw any, field string, pattern *regexp.Regexp, destination map[string]struct{}) error {
	if raw == nil {
		return nil
	}
	values, ok := raw.(map[string]any)
	if !ok {
		return &ValidationError{Field: field, Message: "must be a mapping"}
	}
	for name := range values {
		if !pattern.MatchString(name) {
			return &ValidationError{Field: field + "." + name, Message: "has an invalid resource name"}
		}
		destination[name] = struct{}{}
	}
	return nil
}

func parsePublishedPorts(raw any, field string) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, &ValidationError{Field: field, Message: "must be a list of published port mappings"}
	}
	result := make([]string, 0, len(values))
	for index, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "must use host:container port notation"}
		}
		protocol := "tcp"
		parts := strings.Split(text, "/")
		if len(parts) > 2 {
			return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "has an invalid protocol suffix"}
		}
		if len(parts) == 2 {
			protocol = parts[1]
			if protocol != "tcp" && protocol != "udp" {
				return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "protocol must be tcp or udp"}
			}
		}
		ports := strings.Split(parts[0], ":")
		if len(ports) == 3 {
			if net.ParseIP(ports[0]) == nil {
				return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "host IP must be a valid IPv4 address"}
			}
			ports = ports[1:]
		}
		if len(ports) != 2 || !validPort(ports[0]) || !validPort(ports[1]) {
			return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "must contain valid host and container ports"}
		}
		result = append(result, ports[0]+":"+ports[1]+"/"+protocol)
	}
	return result, nil
}

func validPort(value string) bool {
	port, err := strconv.Atoi(value)
	return err == nil && port >= 1 && port <= 65535
}

func parseEnvironmentKeys(raw any, field string) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	keys := make([]string, 0)
	switch values := raw.(type) {
	case map[string]any:
		for key := range values {
			if !environmentKeyPattern.MatchString(key) {
				return nil, &ValidationError{Field: field + "." + key, Message: "has an invalid environment variable name"}
			}
			keys = append(keys, key)
		}
	case []any:
		for index, value := range values {
			entry, ok := value.(string)
			if !ok {
				return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "must be KEY=VALUE"}
			}
			key, _, found := strings.Cut(entry, "=")
			if !found || !environmentKeyPattern.MatchString(key) {
				return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "must be KEY=VALUE with a valid key"}
			}
			keys = append(keys, key)
		}
	default:
		return nil, &ValidationError{Field: field, Message: "must be a mapping or KEY=VALUE list"}
	}
	return keys, nil
}

func parseServiceVolumes(raw any, field string, declared map[string]struct{}) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, &ValidationError{Field: field, Message: "must be a list of named-volume mappings"}
	}
	result := make([]string, 0, len(values))
	for index, value := range values {
		entry, ok := value.(string)
		entryField := fmt.Sprintf("%s[%d]", field, index)
		if !ok {
			return nil, &ValidationError{Field: entryField, Message: "only short named-volume syntax is supported"}
		}
		if strings.Contains(entry, "docker.sock") {
			return nil, &ValidationError{Field: entryField, Message: "Docker socket mounts are not allowed"}
		}
		parts := strings.Split(entry, ":")
		if len(parts) < 2 || len(parts) > 3 || !volumeNamePattern.MatchString(parts[0]) || !strings.HasPrefix(parts[1], "/") {
			return nil, &ValidationError{Field: entryField, Message: "must use a declared named volume and an absolute container path"}
		}
		if len(parts) == 3 && parts[2] != "ro" && parts[2] != "rw" {
			return nil, &ValidationError{Field: entryField, Message: "volume mode must be ro or rw"}
		}
		if _, ok := declared[parts[0]]; !ok {
			return nil, &ValidationError{Field: entryField, Message: "references an undeclared named volume"}
		}
		result = append(result, entry)
	}
	return result, nil
}

func parseServiceNetworks(raw any, field string, declared map[string]struct{}) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	names := make([]string, 0)
	switch values := raw.(type) {
	case []any:
		for index, value := range values {
			name, ok := value.(string)
			if !ok {
				return nil, &ValidationError{Field: fmt.Sprintf("%s[%d]", field, index), Message: "must be a network name"}
			}
			names = append(names, name)
		}
	case map[string]any:
		for name := range values {
			names = append(names, name)
		}
	default:
		return nil, &ValidationError{Field: field, Message: "must be a list or mapping of declared networks"}
	}
	for _, name := range names {
		if _, ok := declared[name]; !ok {
			return nil, &ValidationError{Field: field + "." + name, Message: "references an undeclared network"}
		}
	}
	return names, nil
}

func validRestartPolicy(value string) bool {
	switch value {
	case "no", "always", "on-failure", "unless-stopped":
		return true
	default:
		return false
	}
}
