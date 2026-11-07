package httpapi

import (
	"net/http"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/deployments"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

type deploymentRecord struct {
	DeploymentID         string `json:"deployment_id"`
	ApplicationID        string `json:"application_id"`
	ServerID             string `json:"server_id"`
	State                string `json:"state"`
	PreviousDeploymentID string `json:"previous_deployment_id,omitempty"`
	Log                  string `json:"log,omitempty"`
	StartedAt            string `json:"started_at,omitempty"`
	FinishedAt           string `json:"finished_at,omitempty"`
}

type createDeploymentRequest struct {
	ApplicationID string `json:"application_id"`
}

// deploymentProgressSteps maps every deployment record state to a stable
// percent and human-readable step for the progress endpoint.
var deploymentProgressSteps = map[string]struct {
	Percent int
	Step    string
}{
	string(deployments.StateQueued):     {Percent: 0, Step: "Queued"},
	string(deployments.StatePreparing):  {Percent: 10, Step: "Preparing deployment"},
	string(deployments.StateUploading):  {Percent: 30, Step: "Uploading Compose manifest"},
	string(deployments.StateValidating): {Percent: 50, Step: "Validating on target host"},
	string(deployments.StateDeploying):  {Percent: 70, Step: "Deploying Compose services"},
	string(deployments.StateVerifying):  {Percent: 90, Step: "Verifying service state"},
	string(deployments.StateSucceeded):  {Percent: 100, Step: "Succeeded"},
	string(deployments.StateFailed):     {Percent: 0, Step: "Failed"},
	string(deployments.StateCancelled):  {Percent: 0, Step: "Cancelled"},
}

func createDeployment(e *core.RequestEvent) error {
	var req createDeploymentRequest
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	req.ApplicationID = strings.TrimSpace(req.ApplicationID)
	if req.ApplicationID == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "application_id is required"})
	}
	application, err := e.App.FindRecordById("arsy_applications", req.ApplicationID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
	}
	if _, err := deployments.ValidateCompose(application.GetString("compose_manifest")); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "application Compose manifest is invalid"})
	}
	previousDeploymentID := ""
	previous, err := e.App.FindRecordsByFilter(
		"arsy_deployments",
		"application = {:application} && status = 'succeeded'",
		"-created",
		1, 0,
		map[string]any{"application": application.Id},
	)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	if len(previous) > 0 {
		previousDeploymentID = previous[0].GetString("deployment_id")
	}
	deploymentID, err := deployments.NewDeploymentID()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	collection, err := e.App.FindCollectionByNameOrId("arsy_deployments")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	record := core.NewRecord(collection)
	record.Set("application", application.Id)
	record.Set("deployment_id", deploymentID)
	record.Set("status", string(deployments.StateQueued))
	record.Set("previous_deployment_id", previousDeploymentID)
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	response := map[string]any{
		"deployment_id":  deploymentID,
		"application_id": application.Id,
		"status":         string(deployments.StateQueued),
	}
	if previousDeploymentID != "" {
		response["previous_deployment_id"] = previousDeploymentID
	}
	return e.JSON(http.StatusCreated, response)
}

func getDeploymentProgress(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}
	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}
	state := record.GetString("status")
	step := deploymentProgressSteps[state].Step
	if step == "" {
		step = state
	}
	return e.JSON(http.StatusOK, map[string]any{
		"deployment_id": id,
		"state":         state,
		"percent":       deploymentProgressSteps[state].Percent,
		"step":          step,
	})
}

func listDeployments(e *core.RequestEvent) error {
	appID := e.Request.URL.Query().Get("application_id")

	var records []*core.Record
	var err error
	if appID != "" {
		records, err = e.App.FindRecordsByFilter(
			"arsy_deployments",
			"application = {:appID}",
			"-started_at",
			100, 0,
			map[string]any{"appID": appID},
		)
	} else {
		records, err = e.App.FindRecordsByFilter(
			"arsy_deployments",
			"",
			"-started_at",
			100, 0,
			nil,
		)
	}
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	result := make([]deploymentRecord, 0, len(records))
	for _, r := range records {
		rec := deploymentRecord{
			DeploymentID:         r.GetString("deployment_id"),
			ApplicationID:        r.GetString("application"),
			ServerID:             r.GetString("server_id"),
			State:                r.GetString("status"),
			PreviousDeploymentID: r.GetString("previous_deployment_id"),
			Log:                  r.GetString("summary"),
		}
		if r.GetDateTime("started_at").String() != "" {
			rec.StartedAt = r.GetDateTime("started_at").String()
		}
		if r.GetDateTime("finished_at").String() != "" {
			rec.FinishedAt = r.GetDateTime("finished_at").String()
		}
		result = append(result, rec)
	}

	return e.JSON(http.StatusOK, result)
}

func getDeployment(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}

	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}

	rec := deploymentRecord{
		DeploymentID:         record.GetString("deployment_id"),
		ApplicationID:        record.GetString("application"),
		ServerID:             record.GetString("server_id"),
		State:                record.GetString("status"),
		PreviousDeploymentID: record.GetString("previous_deployment_id"),
		Log:                  record.GetString("summary"),
	}
	if record.GetDateTime("started_at").String() != "" {
		rec.StartedAt = record.GetDateTime("started_at").String()
	}
	if record.GetDateTime("finished_at").String() != "" {
		rec.FinishedAt = record.GetDateTime("finished_at").String()
	}

	return e.JSON(http.StatusOK, rec)
}

func getDeploymentLogs(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}

	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}

	log := record.GetString("summary")
	if log == "" {
		log = "(no logs available)"
	}

	const maxLogSize = 1 << 20
	if len(log) > maxLogSize {
		log = "...(truncated)\n" + log[len(log)-maxLogSize:]
	}

	return e.JSON(http.StatusOK, map[string]any{
		"deployment_id": id,
		"log":           log,
		"state":         record.GetString("status"),
	})
}

func stopDeployment(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}

	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}

	state := record.GetString("status")
	if state != string(deployments.StateDeploying) && state != string(deployments.StatePreparing) &&
		state != string(deployments.StateUploading) && state != string(deployments.StateValidating) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "deployment is not running"})
	}

	record.Set("status", string(deployments.StateCancelled))
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	return e.JSON(http.StatusOK, map[string]any{"status": "cancelled", "deployment_id": id})
}

func redeployDeployment(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}

	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}

	if record.GetString("status") != string(deployments.StateSucceeded) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "can only redeploy from succeeded deployments"})
	}

	return e.JSON(http.StatusOK, map[string]any{
		"status":                 "redeploy_requested",
		"deployment_id":          id,
		"application_id":         record.GetString("application"),
		"previous_deployment_id": record.GetString("deployment_id"),
		"message":                "Use POST /api/arsy/v1/deployments with application_id to create a new deployment",
	})
}

func rollbackDeployment(e *core.RequestEvent) error {
	id := e.Request.PathValue("id")
	if id == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
	}

	record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
	}

	if record.GetString("status") != string(deployments.StateSucceeded) {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "can only rollback from succeeded deployments"})
	}

	prevID := record.GetString("previous_deployment_id")
	if prevID == "" {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "no previous deployment to rollback to"})
	}

	prevRecord, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", prevID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "previous deployment not found"})
	}

	return e.JSON(http.StatusOK, map[string]any{
		"status":            "rollback_requested",
		"deployment_id":     id,
		"rollback_to":       prevID,
		"application_id":    record.GetString("application"),
		"previous_manifest": prevRecord.GetString("summary"),
		"message":           "Create a new deployment using the previous manifest",
	})
}

func registerDeploymentRoutes(e *core.ServeEvent, api *router.RouterGroup[*core.RequestEvent]) {
	api.POST("/deployments", requireSuperuser(createDeployment))
	api.GET("/deployments", requireSuperuser(listDeployments))
	api.GET("/deployments/{id}", requireSuperuser(getDeployment))
	api.GET("/deployments/{id}/progress", requireSuperuser(getDeploymentProgress))
	api.GET("/deployments/{id}/logs", requireSuperuser(getDeploymentLogs))
	api.POST("/deployments/{id}/stop", requireSuperuser(stopDeployment))
	api.POST("/deployments/{id}/redeploy", requireSuperuser(redeployDeployment))
	api.POST("/deployments/{id}/rollback", requireSuperuser(rollbackDeployment))
}
