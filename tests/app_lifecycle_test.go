package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestNewTestAppUsesIsolatedTemporaryDataDir(t *testing.T) {
	app, err := NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp() error: %v", err)
	}
	dataDir := app.DataDir()
	if dataDir == "" {
		t.Fatal("expected non-empty data dir")
	}
	if filepath.Clean(dataDir) == filepath.Clean(filepath.Join("tests", "data")) {
		t.Fatalf("expected cloned temp data dir, got source-like path: %s", dataDir)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Fatalf("expected temp data dir to exist: %v", err)
	}
	if !app.IsBootstrapped() {
		t.Fatal("expected bootstrapped app")
	}
	if app.DB() == nil || app.AuxDB() == nil {
		t.Fatal("expected DB and AuxDB")
	}
	if _, err := app.DB().NewQuery("Select 1").Execute(); err != nil {
		t.Fatalf("main db query failed: %v", err)
	}
	if _, err := app.AuxDB().NewQuery("Select 1").Execute(); err != nil {
		t.Fatalf("aux db query failed: %v", err)
	}

	app.Cleanup()
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("expected temp data dir removed after cleanup, stat err=%v", err)
	}
}

func TestNewTestAppWithConfigClonesProvidedDataDir(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(currentFile), "data")

	app, err := NewTestAppWithConfig(core.BaseAppConfig{
		DataDir:       src,
		EncryptionEnv: "pb_test_env",
	})
	if err != nil {
		t.Fatalf("NewTestAppWithConfig() error: %v", err)
	}
	defer func() {
		dataDir := app.DataDir()
		app.Cleanup()
		if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
			t.Fatalf("expected cleanup to remove cloned data dir, stat err=%v", err)
		}
	}()

	if filepath.Clean(app.DataDir()) == filepath.Clean(src) {
		t.Fatalf("expected cloned data dir, got source path: %s", app.DataDir())
	}
}

func TestRunAllMigrationsAndLocalRunnerRegistration(t *testing.T) {
	app, err := NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp() error: %v", err)
	}
	defer app.Cleanup()

	if err := app.RunAllMigrations(); err != nil {
		t.Fatalf("RunAllMigrations() error: %v", err)
	}

	executed := false
	list := core.MigrationsList{}
	list.Register(func(_ core.App) error {
		executed = true
		return nil
	}, nil, "9999999999_test_lifecycle_local.go")

	runner := core.NewMigrationsRunner(app, list)
	applied, err := runner.Up()
	if err != nil {
		t.Fatalf("runner.Up() error: %v", err)
	}
	if !executed {
		t.Fatal("expected locally registered migration to execute")
	}
	if len(applied) != 1 || applied[0] != "9999999999_test_lifecycle_local.go" {
		t.Fatalf("unexpected applied migrations: %#v", applied)
	}

	appliedAgain, err := runner.Up()
	if err != nil {
		t.Fatalf("runner.Up() second run error: %v", err)
	}
	if len(appliedAgain) != 0 {
		t.Fatalf("expected no migrations on second run, got %#v", appliedAgain)
	}
}
