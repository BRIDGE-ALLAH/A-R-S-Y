package deployments

import (
	"strings"
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestAppendDeploymentLogPersistsBoundedEntries(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "log-test", StateQueued)

	message := strings.Repeat("a", maxDeploymentLogMessageRunes+50)
	if err := AppendDeploymentLog(app, deployment.Id, "upload", "info", message); err != nil {
		t.Fatal(err)
	}
	logs, err := app.FindRecordsByFilter("arsy_deployment_logs", `deployment = {:deployment}`, "created", 10, 0, map[string]any{"deployment": deployment.Id})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || len([]rune(logs[0].GetString("message"))) != maxDeploymentLogMessageRunes {
		t.Fatalf("persisted logs = %d, message length = %d", len(logs), len([]rune(logs[0].GetString("message"))))
	}
}

func TestAppendDeploymentLogRejectsInvalidEntries(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	for _, test := range []struct {
		step, level, message string
	}{
		{"", "info", "message"},
		{"upload", "debug", "message"},
		{"upload", "info", ""},
	} {
		if err := AppendDeploymentLog(app, "deployment-id", test.step, test.level, test.message); err == nil {
			t.Fatalf("invalid log entry was accepted: %#v", test)
		}
	}
}
