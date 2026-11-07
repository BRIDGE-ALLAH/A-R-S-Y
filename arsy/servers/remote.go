package servers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// RemoteOperation is a closed set of fixed, read-only diagnostics. HTTP input
// cannot provide arbitrary shell commands.
type RemoteOperation string

const (
	RemoteOSFacts         RemoteOperation = "os_facts"
	RemoteDockerVersion   RemoteOperation = "docker_version"
	RemoteComposeVersion  RemoteOperation = "compose_version"
	RemoteComposeValidate RemoteOperation = "compose_validate"
)

type CommandResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Truncated bool
}

type Executor struct {
	Timeout        time.Duration
	MaxOutputBytes int
}

func (e Executor) Execute(ctx context.Context, client *ssh.Client, operation RemoteOperation, secrets ...string) (CommandResult, error) {
	if ctx == nil {
		return CommandResult{}, errors.New("SSH context is required")
	}
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	if client == nil {
		return CommandResult{}, errors.New("SSH client is required")
	}
	command, err := commandForOperation(operation)
	if err != nil {
		return CommandResult{}, err
	}
	return e.executeCommand(ctx, client, operation, command, secrets...)
}

func (e Executor) ValidateComposeManifest(ctx context.Context, client *ssh.Client, manifestPath string) (CommandResult, error) {
	command, err := composeValidationCommand(manifestPath)
	if err != nil {
		return CommandResult{}, err
	}
	return e.executeCommand(ctx, client, RemoteComposeValidate, command)
}

func (e Executor) executeCommand(ctx context.Context, client *ssh.Client, operation RemoteOperation, command string, secrets ...string) (CommandResult, error) {
	if ctx == nil {
		return CommandResult{}, errors.New("SSH context is required")
	}
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	if client == nil {
		return CommandResult{}, errors.New("SSH client is required")
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	maxOutput := e.MaxOutputBytes
	if maxOutput <= 0 {
		maxOutput = 64 * 1024
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	session, err := client.NewSession()
	if err != nil {
		return CommandResult{}, fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()
	stopCancellation := context.AfterFunc(runCtx, func() { _ = session.Close() })
	defer stopCancellation()

	stdout := &boundedBuffer{limit: maxOutput}
	stderr := &boundedBuffer{limit: maxOutput}
	session.Stdout = stdout
	session.Stderr = stderr
	runErr := session.Run(command)

	result := CommandResult{
		Stdout:    redact(strings.TrimSpace(stdout.String()), secrets...),
		Stderr:    redact(strings.TrimSpace(stderr.String()), secrets...),
		ExitCode:  0,
		Truncated: stdout.truncated || stderr.truncated,
	}
	if runErr == nil {
		return result, nil
	}
	if runCtx.Err() != nil {
		return result, runCtx.Err()
	}
	var exitErr *ssh.ExitError
	if errors.As(runErr, &exitErr) {
		result.ExitCode = exitErr.ExitStatus()
		return result, nil
	}
	return result, fmt.Errorf("execute fixed SSH operation %q: %w", operation, runErr)
}

var remoteManifestPathPattern = regexp.MustCompile("^/opt/arsy/apps/[a-z0-9]+(?:-[a-z0-9]+)*/releases/[a-f0-9]{32}/compose\\.yaml$")

func composeValidationCommand(manifestPath string) (string, error) {
	if !remoteManifestPathPattern.MatchString(manifestPath) {
		return "", errors.New("remote Compose manifest path is not an A-R-S-Y release path")
	}
	return "docker compose -f " + shellQuote(manifestPath) + " config --quiet", nil
}

func commandForOperation(operation RemoteOperation) (string, error) {
	switch operation {
	case RemoteOSFacts:
		return "uname -s && uname -m", nil
	case RemoteDockerVersion:
		return "docker version --format '{{.Server.Version}}'", nil
	case RemoteComposeVersion:
		return "docker compose version --short", nil
	case RemoteComposeValidate:
		return "", fmt.Errorf("operation %q requires a generated deployment manifest path", operation)
	default:
		return "", fmt.Errorf("unsupported remote operation %q", operation)
	}
}

type boundedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(value []byte) (int, error) {
	originalLength := len(value)
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = b.truncated || originalLength > 0
		return originalLength, nil
	}
	if len(value) > remaining {
		value = value[:remaining]
		b.truncated = true
	}
	_, err := b.buffer.Write(value)
	return originalLength, err
}

func (b *boundedBuffer) String() string {
	return b.buffer.String()
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func redact(value string, secrets ...string) string {
	return RedactSecrets(value, secrets...)
}
