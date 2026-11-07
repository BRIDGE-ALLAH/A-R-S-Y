package servers

import (
	"strings"
	"testing"
)

func TestCommandForOperationIsClosedAndFixed(t *testing.T) {
	tests := []struct {
		operation RemoteOperation
		want      string
	}{
		{RemoteOSFacts, "uname -s && uname -m"},
		{RemoteDockerVersion, "docker version --format '{{.Server.Version}}'"},
		{RemoteComposeVersion, "docker compose version --short"},
	}
	for _, tt := range tests {
		got, err := commandForOperation(tt.operation)
		if err != nil || got != tt.want {
			t.Fatalf("commandForOperation(%q) = %q, %v; want %q", tt.operation, got, err, tt.want)
		}
	}
	if _, err := commandForOperation(RemoteOperation("echo unsafe")); err == nil {
		t.Fatal("arbitrary command was accepted")
	}
}

func TestComposeValidationCommandUsesOnlyGeneratedReleasePaths(t *testing.T) {
	manifestPath := "/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml"
	got, err := composeValidationCommand(manifestPath)
	want := "docker compose --project-name 'demo-app' -f '/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml' config --quiet"
	if err != nil || got != want {
		t.Fatalf("composeValidationCommand() = %q, %v; want %q", got, err, want)
	}
	for action, want := range map[string]string{
		"config --quiet":   "docker compose --project-name 'demo-app' -f '/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml' config --quiet",
		"up -d":            "docker compose --project-name 'demo-app' -f '/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml' up -d",
		"ps --format json": "docker compose --project-name 'demo-app' -f '/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml' ps --format json",
		"down":             "docker compose --project-name 'demo-app' -f '/opt/arsy/apps/demo-app/releases/0123456789abcdef0123456789abcdef/compose.yaml' down",
	} {
		got, err := composeManifestCommand(manifestPath, action)
		if err != nil || got != want {
			t.Fatalf("composeManifestCommand(%q) = %q, %v; want %q", action, got, err, want)
		}
	}
	for _, invalid := range []string{
		"/tmp/compose.yaml",
		"/opt/arsy/apps/../etc/releases/0123456789abcdef0123456789abcdef/compose.yaml",
		"/opt/arsy/apps/demo/releases/../../etc/compose.yaml",
	} {
		if _, err := composeValidationCommand(invalid); err == nil {
			t.Fatalf("unsafe manifest path %q was accepted", invalid)
		}
	}
	for _, operation := range []RemoteOperation{RemoteComposeValidate, RemoteComposeDeploy, RemoteComposeInspect, RemoteComposeStop} {
		if _, err := commandForOperation(operation); err == nil {
			t.Fatalf("dynamic operation %q must require a validated manifest path", operation)
		}
	}
}

func TestBoundedBufferTruncatesWithoutFailingWriter(t *testing.T) {
	buffer := &boundedBuffer{limit: 5}
	n, err := buffer.Write([]byte("abcdefgh"))
	if err != nil || n != 8 {
		t.Fatalf("Write() = (%d, %v), want (8, nil)", n, err)
	}
	if buffer.String() != "abcde" || !buffer.truncated {
		t.Fatalf("buffer = %q, truncated = %v", buffer.String(), buffer.truncated)
	}
}

func TestExecutorRedactionHelper(t *testing.T) {
	got := redact("failed with token secret-value", "secret-value")
	if strings.Contains(got, "secret-value") {
		t.Fatalf("secret was not redacted: %q", got)
	}
}
