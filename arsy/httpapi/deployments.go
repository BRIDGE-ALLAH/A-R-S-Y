package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/deployments"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

type deploymentRecord struct {
	DeploymentID         string `json:"deployment_id"`
	ApplicationID        string `json:"application_id"`
	State                string `json:"state"`
	PreviousDeploymentID string `json:"previous_deployment_id,omitempty"`
	Log                  string `json:"log,omitempty"`
	StartedAt            string `json:"started_at,omitempty"`
	FinishedAt           string `json:"finished_at,omitempty"`
}

type createDeploymentRequest struct {
	ApplicationID  string `json:"application_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

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

func createDeployment(cfg config.Config) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var req createDeploymentRequest
		if err := decodeServerRequest(e, &req); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}
		req.ApplicationID = strings.TrimSpace(req.ApplicationID)
		req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
		if req.ApplicationID == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "application_id is required"})
		}
		if req.IdempotencyKey != "" {
			existing, err := deployments.FindDeploymentByIdempotencyKey(e.App, req.IdempotencyKey)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			}
			if existing != nil {
				existingAppID := existing.GetString("application")
				if existingAppID == req.ApplicationID {
					response := map[string]any{
						"deployment_id": existing.GetString("deployment_id"),
						"status":        existing.GetString("status"),
					}
					if req.IdempotencyKey != "" {
						response["idempotency_key"] = req.IdempotencyKey
					}
					if prevID := existing.GetString("previous_deployment_id"); prevID != "" {
						response["previous_deployment_id"] = prevID
					}
					return e.JSON(http.StatusOK, response)
				}
			}
		}
		application, err := e.App.FindRecordById("arsy_applications", req.ApplicationID)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
		}
		server, err := e.App.FindRecordById("arsy_servers", application.GetString("server"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "server not found"})
		}
		if server.GetBool("host_key_revoked") {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "host key revoked; cannot deploy"})
		}
		active, err := deployments.HasActiveDeployment(e.App, req.ApplicationID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		if active {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "deployment in progress for this application"})
		}

		manifestPlaintext := application.GetString("compose_manifest")
		if ciphertext := application.GetString("manifest_ciphertext"); ciphertext != "" {
			decrypted, decErr := deployments.DecryptManifest(ciphertext, cfg.SecretKeyEnv)
			if decErr != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to decrypt application manifest"})
			}
			manifestPlaintext = decrypted
		}
		if _, err := deployments.ValidateCompose(manifestPlaintext); err != nil {
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
		if req.IdempotencyKey != "" {
			record.Set("idempotency_key", req.IdempotencyKey)
		}
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
		if req.IdempotencyKey != "" {
			response["idempotency_key"] = req.IdempotencyKey
		}
		if previousDeploymentID != "" {
			response["previous_deployment_id"] = previousDeploymentID
		}
		return e.JSON(http.StatusCreated, response)
	}
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

	return e.JSON(http.StatusOK, map[string]any{"items": result})
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
	switch deployments.State(state) {
	case deployments.StateQueued, deployments.StatePreparing, deployments.StateUploading, deployments.StateValidating, deployments.StateDeploying, deployments.StateVerifying:
	default:
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "deployment is not running"})
	}

	if _, err := deployments.UpdateDeploymentState(e.App, record.Id, deployments.StateCancelled, "Cancelled by operator request."); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	return e.JSON(http.StatusOK, map[string]any{"status": "cancelled", "deployment_id": id})
}

func redeployDeployment(cfg config.Config) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		if id == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
		}
		record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
		}
		if record.GetString("status") != string(deployments.StateSucceeded) && record.GetString("status") != string(deployments.StateFailed) && record.GetString("status") != string(deployments.StateCancelled) {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "can only redeploy from a terminal deployment"})
		}
		application, err := e.App.FindRecordById("arsy_applications", record.GetString("application"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
		}
		server, err := e.App.FindRecordById("arsy_servers", application.GetString("server"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "server not found"})
		}
		if server.GetBool("host_key_revoked") {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "host key revoked; cannot deploy"})
		}
		active, err := deployments.HasActiveDeployment(e.App, application.Id)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		if active {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "deployment in progress for this application"})
		}
		manifestPlaintext := application.GetString("compose_manifest")
		if ciphertext := application.GetString("manifest_ciphertext"); ciphertext != "" {
			decrypted, decErr := deployments.DecryptManifest(ciphertext, cfg.SecretKeyEnv)
			if decErr != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to decrypt application manifest"})
			}
			manifestPlaintext = decrypted
		}
		if _, err := deployments.ValidateCompose(manifestPlaintext); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "application Compose manifest is invalid"})
		}
		deploymentID, err := deployments.NewDeploymentID()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		collection, err := e.App.FindCollectionByNameOrId("arsy_deployments")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		queued := core.NewRecord(collection)
		queued.Set("application", application.Id)
		queued.Set("deployment_id", deploymentID)
		queued.Set("status", string(deployments.StateQueued))
		queued.Set("previous_deployment_id", record.GetString("deployment_id"))
		if err := e.App.Save(queued); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		return e.JSON(http.StatusCreated, map[string]any{
			"status":               "queued",
			"deployment_id":        deploymentID,
			"application_id":       application.Id,
			"source_deployment_id": record.GetString("deployment_id"),
			// Kept for contract compatibility with deployment records.
			"previous_deployment_id": record.GetString("deployment_id"),
		})
	}
}

func rollbackDeployment(cfg config.Config) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		if id == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "deployment ID required"})
		}
		record, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", id)
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "deployment not found"})
		}
		if record.GetString("status") != string(deployments.StateSucceeded) && record.GetString("status") != string(deployments.StateFailed) && record.GetString("status") != string(deployments.StateCancelled) {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "can only rollback from a terminal deployment"})
		}
		prevID := record.GetString("previous_deployment_id")
		if prevID == "" {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "no previous deployment to rollback to"})
		}
		prev, err := e.App.FindFirstRecordByData("arsy_deployments", "deployment_id", prevID)
		if err != nil || prev == nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "previous deployment not found"})
		}
		application, err := e.App.FindRecordById("arsy_applications", record.GetString("application"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
		}
		server, err := e.App.FindRecordById("arsy_servers", application.GetString("server"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "server not found"})
		}
		if server.GetBool("host_key_revoked") {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "host key revoked; cannot deploy"})
		}
		active, err := deployments.HasActiveDeployment(e.App, application.Id)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		if active {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "deployment in progress for this application"})
		}
		snapshot, err := deployments.LoadDeploymentManifest(e.App, prevID, cfg.SecretKeyEnv)
		if err != nil {
			if errors.Is(err, deployments.ErrSnapshotUnavailable) {
				return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "previous deployment has no stored manifest snapshot; cannot roll back"})
			}
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to load previous deployment manifest snapshot"})
		}
		if err := deployments.ApplyManifestSnapshot(e.App, application.Id, snapshot, cfg.SecretKeyEnv); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "failed to restore previous deployment manifest"})
		}
		deploymentID, err := deployments.NewDeploymentID()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		collection, err := e.App.FindCollectionByNameOrId("arsy_deployments")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		queued := core.NewRecord(collection)
		queued.Set("application", application.Id)
		queued.Set("deployment_id", deploymentID)
		queued.Set("status", string(deployments.StateQueued))
		queued.Set("previous_deployment_id", prevID)
		if err := e.App.Save(queued); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		return e.JSON(http.StatusCreated, map[string]any{
			"status":                 "queued",
			"deployment_id":          deploymentID,
			"application_id":         application.Id,
			"rollback_to":            prevID,
			"previous_deployment_id": prevID,
		})
	}
}

func registerDeploymentRoutes(e *core.ServeEvent, api *router.RouterGroup[*core.RequestEvent], cfg config.Config) {
	api.POST("/deployments", requireSuperuser(createDeployment(cfg)))
	api.GET("/deployments", requireSuperuser(listDeployments))
	api.GET("/deployments/{id}", requireSuperuser(getDeployment))
	api.GET("/deployments/{id}/progress", requireSuperuser(getDeploymentProgress))
	api.GET("/deployments/{id}/logs", requireSuperuser(getDeploymentLogs))
	api.POST("/deployments/{id}/stop", requireSuperuser(stopDeployment))
	api.POST("/deployments/{id}/redeploy", requireSuperuser(redeployDeployment(cfg)))
	api.POST("/deployments/{id}/rollback", requireSuperuser(rollbackDeployment(cfg)))
}
