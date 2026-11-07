package servers

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
)

var linuxUsernamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

// ConnectionInput contains the user-supplied connection details accepted by
// the server-registration workflow.
type ConnectionInput struct {
	Host       string
	Port       int
	Username   string
	PrivateKey []byte
}

// ValidateConnectionInput validates and normalizes SSH connection details.
// Hostnames are lowercased and a trailing DNS root dot is removed; IP literals
// are preserved in their canonical net.IP string form.
func ValidateConnectionInput(input ConnectionInput) (ConnectionInput, error) {
	input.Host = strings.TrimSpace(input.Host)
	if input.Host == "" {
		return ConnectionInput{}, errors.New("server host is required")
	}
	if strings.ContainsAny(input.Host, " \t\r\n/@") {
		return ConnectionInput{}, errors.New("server host contains invalid characters")
	}
	if ip := net.ParseIP(input.Host); ip != nil {
		input.Host = ip.String()
	} else {
		host := strings.TrimSuffix(strings.ToLower(input.Host), ".")
		if len(host) == 0 || len(host) > 253 {
			return ConnectionInput{}, errors.New("server hostname must be between 1 and 253 characters")
		}
		for _, label := range strings.Split(host, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return ConnectionInput{}, fmt.Errorf("invalid server hostname label %q", label)
			}
			for _, r := range label {
				if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
					return ConnectionInput{}, fmt.Errorf("invalid character in server hostname label %q", label)
				}
			}
		}
		input.Host = host
	}
	if input.Port < 1 || input.Port > 65535 {
		return ConnectionInput{}, errors.New("SSH port must be between 1 and 65535")
	}
	input.Username = strings.TrimSpace(input.Username)
	if !linuxUsernamePattern.MatchString(input.Username) {
		return ConnectionInput{}, errors.New("SSH username must be a valid Linux account name")
	}
	if err := ValidatePrivateKey(input.PrivateKey); err != nil {
		return ConnectionInput{}, err
	}
	return input, nil
}
