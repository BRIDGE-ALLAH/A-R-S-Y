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
