package deployments

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/bridge-allah/a-r-s-y/core"
)

const maxDeploymentLogMessageRunes = 5000

// AppendDeploymentLog persists a bounded operator-facing deployment event.
// Callers must pass already-redacted command output.
func AppendDeploymentLog(app core.App, deploymentID, step, level, message string) error {
	if app == nil {
		return errors.New("PocketBase app is required")
	}
	if strings.TrimSpace(deploymentID) == "" {
		return errors.New("deployment ID is required")
	}
	step = strings.TrimSpace(step)
	if step == "" || len([]rune(step)) > 80 {
		return errors.New("deployment log step must be between 1 and 80 characters")
	}
	switch level {
	case "info", "warning", "error":
	default:
		return errors.New("deployment log level must be info, warning, or error")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return errors.New("deployment log message is required")
	}
	if !utf8.ValidString(message) {
		return errors.New("deployment log message must be valid UTF-8")
	}
	runes := []rune(message)
	if len(runes) > maxDeploymentLogMessageRunes {
		message = string(runes[:maxDeploymentLogMessageRunes])
	}

	collection, err := app.FindCachedCollectionByNameOrId("arsy_deployment_logs")
	if err != nil {
		return errors.New("deployment log storage is unavailable")
	}
	record := core.NewRecord(collection)
	record.Set("deployment", deploymentID)
	record.Set("step", step)
	record.Set("level", level)
	record.Set("message", message)
	if err := app.Save(record); err != nil {
		return errors.New("unable to persist deployment log")
	}
	return nil
}
