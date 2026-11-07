package servers

import (
	"errors"
	"strings"
)

type ComponentStatus struct {
	Ready   bool   `json:"ready"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Version string `json:"version,omitempty"`
}

type ReadinessSummary struct {
	Ready   bool            `json:"ready"`
	SSH     ComponentStatus `json:"ssh"`
	HostKey ComponentStatus `json:"host_key"`
	OS      ComponentStatus `json:"os"`
	Docker  ComponentStatus `json:"docker"`
	Compose ComponentStatus `json:"compose"`
}

// ParseLinuxFacts accepts the stable two-line output from uname -s && uname -m.
func ParseLinuxFacts(output string) (string, string, error) {
	fields := strings.Fields(output)
	if len(fields) != 2 || fields[0] != "Linux" {
		return "", "", errors.New("managed host must report Linux and a supported architecture")
	}
	architecture := fields[1]
	switch architecture {
	case "x86_64", "amd64":
		architecture = "amd64"
	case "aarch64", "arm64":
		architecture = "arm64"
	default:
		return "", "", errors.New("managed host architecture is not supported")
	}
	return "linux", architecture, nil
}

func DockerReadiness(result CommandResult) ComponentStatus {
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) != "" {
		return ComponentStatus{Ready: true, Code: "ready", Message: "Docker Engine is available", Version: strings.TrimSpace(result.Stdout)}
	}
	message := strings.ToLower(result.Stderr + " " + result.Stdout)
	switch {
	case strings.Contains(message, "permission denied"), strings.Contains(message, "permission to access the docker"):
		return ComponentStatus{Code: "permission_denied", Message: "SSH user cannot access Docker; grant access explicitly and retry"}
	case strings.Contains(message, "cannot connect"), strings.Contains(message, "daemon is not running"), strings.Contains(message, "is the docker daemon running"):
		return ComponentStatus{Code: "daemon_unavailable", Message: "Docker daemon is unavailable"}
	case strings.Contains(message, "not found"), strings.Contains(message, "command not found"):
		return ComponentStatus{Code: "docker_missing", Message: "Docker Engine is not installed or not on PATH"}
	default:
		return ComponentStatus{Code: "docker_unavailable", Message: "Docker Engine readiness could not be confirmed"}
	}
}

func ComposeReadiness(result CommandResult) ComponentStatus {
	if result.ExitCode == 0 {
		version := strings.TrimSpace(result.Stdout)
		if version != "" {
			return ComponentStatus{Ready: true, Code: "ready", Message: "Docker Compose plugin is available", Version: version}
		}
	}
	message := strings.ToLower(result.Stderr + " " + result.Stdout)
	if strings.Contains(message, "not a docker command") || strings.Contains(message, "unknown command: compose") || strings.Contains(message, "command not found") {
		return ComponentStatus{Code: "compose_missing", Message: "Docker Compose plugin is not installed"}
	}
	return ComponentStatus{Code: "compose_unavailable", Message: "Docker Compose plugin readiness could not be confirmed"}
}

func BuildReadinessSummary(sshStatus, hostKeyStatus, osStatus, dockerStatus, composeStatus ComponentStatus) ReadinessSummary {
	return ReadinessSummary{
		Ready:   sshStatus.Ready && hostKeyStatus.Ready && osStatus.Ready && dockerStatus.Ready && composeStatus.Ready,
		SSH:     sshStatus,
		HostKey: hostKeyStatus,
		OS:      osStatus,
		Docker:  dockerStatus,
		Compose: composeStatus,
	}
}
