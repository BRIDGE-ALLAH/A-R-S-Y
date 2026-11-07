package diagnostics

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectReportsReadySetupWithoutSecretValue(t *testing.T) {
	secret := "never-print-this-secret"
	t.Setenv("ARSY_DOCTOR_TEST_KEY", secret)

	status, err := inspect(t.TempDir(), "1.2.3", "ARSY_DOCTOR_TEST_KEY", func() error { return nil })
	if err != nil {
		t.Fatalf("inspect() error = %v", err)
	}
	if status.Version != "1.2.3" || status.GoVersion == "" || status.GOOS == "" || status.GOARCH == "" {
		t.Fatalf("inspect() returned incomplete runtime information: %#v", status)
	}
	if status.DataDirectory != "accessible" || status.Database != "ready" || status.SecretKey != "configured" {
		t.Fatalf("inspect() = %#v", status)
	}
	if strings.Contains(strings.Join([]string{status.Version, status.GoVersion, status.DataDirectory, status.Database, status.SecretKey}, " "), secret) {
		t.Fatal("doctor status exposed the secret value")
	}
}

func TestInspectReturnsErrorForBrokenLocalSetup(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing")
	status, err := inspect(missingDir, "1.2.3", "", func() error {
		return errors.New("private database detail")
	})
	if err == nil {
		t.Fatal("expected broken local setup to return an error")
	}
	if status.DataDirectory != "unavailable" || status.Database != "not ready" {
		t.Fatalf("inspect() status = %#v", status)
	}
	if strings.Contains(err.Error(), "private database detail") {
		t.Fatal("doctor error exposed the database error detail")
	}
}

func TestCheckDataDirectoryRequiresWriteAccess(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkDataDirectory(file); err == nil {
		t.Fatal("expected non-directory data path to fail")
	}
}
