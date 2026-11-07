package servers

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
)

var sshUsernamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*[$]?$`)

// ConnectionInput contains the fields needed to establish an SSH connection.
type ConnectionInput struct {
	Host       string
	Port       int
	Username   string
	PrivateKey []byte
}

// ValidateConnectionInput checks connection data and returns a normalized host.
// DNS names are lowercased; IP literals and zone identifiers are preserved.
func ValidateConnectionInput(input ConnectionInput) (string, error) {
	host := strings.TrimSpace(input.Host)
	if host == "" || strings.ContainsAny(host, " \t\r\n/@\\") {
		return "", fmt.Errorf("host must be a valid hostname or IP address")
	}
	if strings.ContainsRune(host, '\x00') {
		return "", fmt.Errorf("host must be a valid hostname or IP address")
	}
	if ip := net.ParseIP(host); ip == nil {
		if strings.Contains(host, ":") {
			return "", fmt.Errorf("host must be a valid hostname or IP address")
		}
		if err := validateDNSName(host); err != nil {
			return "", err
		}
		host = strings.ToLower(host)
	}
	if input.Port < 1 || input.Port > 65535 {
		return "", fmt.Errorf("SSH port must be between 1 and 65535")
	}
	if !sshUsernamePattern.MatchString(input.Username) || len(input.Username) > 64 {
		return "", fmt.Errorf("SSH username has an invalid format")
	}
	if _, err := ssh.ParsePrivateKey(input.PrivateKey); err != nil {
		return "", fmt.Errorf("SSH private key is invalid or unsupported")
	}
	return host, nil
}

func validateDNSName(host string) error {
	if len(host) > 253 || strings.HasSuffix(host, ".") {
		return fmt.Errorf("host must be a valid hostname or IP address")
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("host must be a valid hostname or IP address")
		}
		for _, char := range label {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-') {
				return fmt.Errorf("host must be a valid hostname or IP address")
			}
		}
	}
	return nil
}
