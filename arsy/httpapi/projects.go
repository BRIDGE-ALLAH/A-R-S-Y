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

	return e.JSON(http.StatusCreated, projectDTO{
		ID:          record.Id,
		Name:        record.GetString("name"),
		Slug:        record.GetString("slug"),
		Description: record.GetString("description"),
		Created:     record.GetDateTime("created").String(),
		Updated:     record.GetDateTime("updated").String(),
	})
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
		items = append(items, applicationDTO{
			ID:             r.Id,
			ProjectID:      r.GetString("project"),
			ServerID:       r.GetString("server"),
			Name:           r.GetString("name"),
			Slug:           r.GetString("slug"),
			LifecycleState: r.GetString("lifecycle_state"),
			Created:        r.GetDateTime("created").String(),
			Updated:        r.GetDateTime("updated").String(),
		})
	}
	return e.JSON(http.StatusOK, map[string]any{"items": items})
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
		if encErr == nil {
			record.Set("manifest_ciphertext", ciphertext)
			record.Set("compose_manifest", "")
		} else {
			record.Set("compose_manifest", req.ComposeManifest)
		}

		if err := e.App.Save(record); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": err.Error()})
		}

		return e.JSON(http.StatusCreated, applicationDTO{
			ID:             record.Id,
			ProjectID:      record.GetString("project"),
			ServerID:       record.GetString("server"),
			Name:           record.GetString("name"),
			Slug:           record.GetString("slug"),
			LifecycleState: record.GetString("lifecycle_state"),
			Created:        record.GetDateTime("created").String(),
			Updated:        record.GetDateTime("updated").String(),
		})
	}
}

func registerProjectAndAppRoutes(api *router.RouterGroup[*core.RequestEvent], cfg config.Config) {
	api.GET("/projects", requireSuperuser(listProjects))
	api.POST("/projects", requireSuperuser(createProject))
	api.GET("/applications", requireSuperuser(listApplications))
	api.POST("/applications", requireSuperuser(createApplication(cfg)))
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = regexp.MustCompile("[^a-z0-9]+").ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
