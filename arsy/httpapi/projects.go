package httpapi

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/deployments"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type projectDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

type applicationDTO struct {
	ID             string `json:"id"`
	ProjectID      string `json:"project_id"`
	ServerID       string `json:"server_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	LifecycleState string `json:"lifecycle_state"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type createApplicationRequest struct {
	ProjectID       string `json:"project_id"`
	ServerID        string `json:"server_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	ComposeManifest string `json:"compose_manifest"`
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateApplicationRequest struct {
	Name            *string `json:"name"`
	ComposeManifest *string `json:"compose_manifest"`
}

func listProjects(e *core.RequestEvent) error {
	records, err := e.App.FindAllRecords("arsy_projects")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	items := make([]projectDTO, 0, len(records))
	for _, r := range records {
		items = append(items, projectDTO{
			ID:          r.Id,
			Name:        r.GetString("name"),
			Slug:        r.GetString("slug"),
			Description: r.GetString("description"),
			Created:     r.GetDateTime("created").String(),
			Updated:     r.GetDateTime("updated").String(),
		})
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items})
}

func getProject(e *core.RequestEvent) error {
	record, err := e.App.FindRecordById("arsy_projects", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "project not found"})
	}
	return e.JSON(http.StatusOK, projectDTOFromRecord(record))
}

func updateProject(e *core.RequestEvent) error {
	var req updateProjectRequest
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	record, err := e.App.FindRecordById("arsy_projects", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "project not found"})
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		record.Set("name", name)
	}
	if req.Description != "" || req.Name != "" {
		record.Set("description", strings.TrimSpace(req.Description))
	}
	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	return e.JSON(http.StatusOK, projectDTOFromRecord(record))
}

func deleteProject(e *core.RequestEvent) error {
	record, err := e.App.FindRecordById("arsy_projects", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "project not found"})
	}
	apps, err := e.App.FindRecordsByFilter("arsy_applications", "project = {:project}", "", 1, 0, map[string]any{"project": record.Id})
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	if len(apps) > 0 {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "project still has applications; remove them first"})
	}
	if err := e.App.Delete(record); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "deleted", "project_id": record.Id})
}

func createProject(e *core.RequestEvent) error {
	var req createProjectRequest
	if err := decodeServerRequest(e, &req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
	if req.Name == "" {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "name is required"})
	}
	if req.Slug == "" {
		req.Slug = generateSlug(req.Name)
	}
	if !slugPattern.MatchString(req.Slug) {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid slug"})
	}

	collection, err := e.App.FindCollectionByNameOrId("arsy_projects")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	record := core.NewRecord(collection)
	record.Set("name", req.Name)
	record.Set("slug", req.Slug)
	record.Set("description", req.Description)

	if err := e.App.Save(record); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
	}

	return e.JSON(http.StatusCreated, projectDTOFromRecord(record))
}

func listApplications(e *core.RequestEvent) error {
	projectID := e.Request.URL.Query().Get("project_id")
	var records []*core.Record
	var err error
	if projectID != "" {
		records, err = e.App.FindRecordsByFilter("arsy_applications", "project = {:project}", "-created", 100, 0, map[string]any{"project": projectID})
	} else {
		records, err = e.App.FindAllRecords("arsy_applications")
	}
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}

	items := make([]applicationDTO, 0, len(records))
	for _, r := range records {
		items = append(items, applicationDTOFromRecord(r))
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items})
}

func getApplication(e *core.RequestEvent) error {
	record, err := e.App.FindRecordById("arsy_applications", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
	}
	return e.JSON(http.StatusOK, applicationDTOFromRecord(record))
}

func updateApplication(cfg config.Config) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var req updateApplicationRequest
		if err := decodeServerRequest(e, &req); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}
		record, err := e.App.FindRecordById("arsy_applications", e.Request.PathValue("id"))
		if err != nil {
			return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
		}
		active, err := deployments.HasActiveDeployment(e.App, record.Id)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}
		if active {
			return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "cannot update application while a deployment is active"})
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "name must not be empty"})
			}
			record.Set("name", name)
		}
		if req.ComposeManifest != nil {
			manifest := strings.TrimSpace(*req.ComposeManifest)
			if manifest == "" {
				return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "compose_manifest must not be empty"})
			}
			ciphertext, encErr := deployments.EncryptManifest(manifest, cfg.SecretKeyEnv)
			if encErr != nil {
				return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "unable to encrypt application manifest; verify the secret key"})
			}
			record.Set("manifest_ciphertext", ciphertext)
			record.Set("compose_manifest", "")
		}
		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}
		return e.JSON(http.StatusOK, applicationDTOFromRecord(record))
	}
}

func deleteApplication(e *core.RequestEvent) error {
	record, err := e.App.FindRecordById("arsy_applications", e.Request.PathValue("id"))
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]any{"status": "error", "message": "application not found"})
	}
	active, err := deployments.HasActiveDeployment(e.App, record.Id)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	if active {
		return e.JSON(http.StatusConflict, map[string]any{"status": "error", "message": "cannot delete application while a deployment is active"})
	}
	if err := e.App.Delete(record); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
	}
	return e.JSON(http.StatusOK, map[string]any{"status": "deleted", "application_id": record.Id})
}

func createApplication(cfg config.Config) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var req createApplicationRequest
		if err := decodeServerRequest(e, &req); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
		if req.Name == "" || req.ProjectID == "" || req.ServerID == "" || req.ComposeManifest == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "all fields are required"})
		}
		if req.Slug == "" {
			req.Slug = generateSlug(req.Name)
		}
		if !slugPattern.MatchString(req.Slug) {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "invalid slug"})
		}

		collection, err := e.App.FindCollectionByNameOrId("arsy_applications")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
		}

		record := core.NewRecord(collection)
		record.Set("project", req.ProjectID)
		record.Set("server", req.ServerID)
		record.Set("name", req.Name)
		record.Set("slug", req.Slug)
		record.Set("lifecycle_state", "draft")

		ciphertext, encErr := deployments.EncryptManifest(req.ComposeManifest, cfg.SecretKeyEnv)
		if encErr != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "unable to encrypt application manifest; verify the secret key"})
		}
		record.Set("manifest_ciphertext", ciphertext)
		record.Set("compose_manifest", "")

		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}

		return e.JSON(http.StatusCreated, applicationDTOFromRecord(record))
	}
}

func registerProjectAndAppRoutes(api *router.RouterGroup[*core.RequestEvent], cfg config.Config) {
	api.GET("/projects", requireSuperuser(listProjects))
	api.POST("/projects", requireSuperuser(createProject))
	api.GET("/projects/{id}", requireSuperuser(getProject))
	api.PATCH("/projects/{id}", requireSuperuser(updateProject))
	api.DELETE("/projects/{id}", requireSuperuser(deleteProject))
	api.GET("/applications", requireSuperuser(listApplications))
	api.POST("/applications", requireSuperuser(createApplication(cfg)))
	api.GET("/applications/{id}", requireSuperuser(getApplication))
	api.PATCH("/applications/{id}", requireSuperuser(updateApplication(cfg)))
	api.DELETE("/applications/{id}", requireSuperuser(deleteApplication))
}

func projectDTOFromRecord(r *core.Record) projectDTO {
	return projectDTO{
		ID:          r.Id,
		Name:        r.GetString("name"),
		Slug:        r.GetString("slug"),
		Description: r.GetString("description"),
		Created:     r.GetDateTime("created").String(),
		Updated:     r.GetDateTime("updated").String(),
	}
}

func applicationDTOFromRecord(r *core.Record) applicationDTO {
	return applicationDTO{
		ID:             r.Id,
		ProjectID:      r.GetString("project"),
		ServerID:       r.GetString("server"),
		Name:           r.GetString("name"),
		Slug:           r.GetString("slug"),
		LifecycleState: r.GetString("lifecycle_state"),
		Created:        r.GetDateTime("created").String(),
		Updated:        r.GetDateTime("updated").String(),
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = regexp.MustCompile("[^a-z0-9]+").ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
