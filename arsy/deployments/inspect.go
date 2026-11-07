package deployments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
)

// ContainerStatus is the normalized per-service view of `docker compose ps`.
type ContainerStatus struct {
	Service string `json:"service"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Health  string `json:"health"`
	Ports   string `json:"ports"`
}

// DeploymentInspection is the read-only result of inspecting a stored
// deployment and its live Compose services.
type DeploymentInspection struct {
	DeploymentID         string            `json:"deployment_id"`
	ApplicationSlug      string            `json:"application_slug"`
	Status               string            `json:"status"`
	Summary              string            `json:"summary"`
	ManifestChecksum     string            `json:"manifest_checksum"`
	PreviousDeploymentID string            `json:"previous_deployment_id"`
	StartedAt            string            `json:"started_at"`
	FinishedAt           string            `json:"finished_at"`
	Overall              string            `json:"overall"`
	Services             []ContainerStatus `json:"services"`
}

type composePSContainer struct {
	Service    string               `json:"Service"`
	Image      string               `json:"Image"`
	State      string               `json:"State"`
	Health     string               `json:"Health"`
	Ports      string               `json:"Ports"`
	Publishers []composePSPublisher `json:"Publishers"`
}

type composePSPublisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
}

// ParseComposePS accepts Docker Compose JSON array output or
// newline-delimited JSON and returns normalized per-service statuses.
// Empty output, no attributable services, and invalid JSON are errors.
func ParseComposePS(output string) ([]ContainerStatus, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, errors.New("Compose inspection returned no containers")
	}
	entries := make([]composePSContainer, 0)
	if strings.HasPrefix(output, "[") {
		if err := json.Unmarshal([]byte(output), &entries); err != nil {
			return nil, errors.New("Compose inspection returned invalid JSON")
		}
	} else {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var entry composePSContainer
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				return nil, errors.New("Compose inspection returned invalid JSON")
			}
			entries = append(entries, entry)
		}
	}
	services := make([]ContainerStatus, 0, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.Service) == "" && strings.TrimSpace(entry.State) == "" {
			continue
		}
		services = append(services, ContainerStatus{
			Service: entry.Service,
			Image:   strings.TrimSpace(entry.Image),
			State:   normalizeServiceState(entry.State, entry.Health),
			Health:  strings.ToLower(strings.TrimSpace(entry.Health)),
			Ports:   containerPorts(entry),
		})
	}
	if len(services) == 0 {
		return nil, errors.New("Compose inspection returned no containers")
	}
	return services, nil
}

// normalizeServiceState maps raw Compose states to the closed inspection
// vocabulary. An unhealthy health check overrides a running container, and a
// service row without a container state is reported as missing.
func normalizeServiceState(state, health string) string {
	state = strings.ToLower(strings.TrimSpace(state))
	health = strings.ToLower(strings.TrimSpace(health))
	if health == "unhealthy" {
		return "unhealthy"
	}
	switch state {
	case "running", "exited", "restarting", "paused", "created", "dead":
		return state
	case "":
		return "missing"
	default:
		return state
	}
}

// containerPorts prefers structured publisher entries and falls back to the
// legacy `Ports` string emitted by older Docker Compose versions.
func containerPorts(entry composePSContainer) string {
	published := make([]string, 0, len(entry.Publishers))
	for _, publisher := range entry.Publishers {
		if publisher.PublishedPort == 0 {
			continue
		}
		host := publisher.URL
		if host == "" {
			host = "0.0.0.0"
		}
		published = append(published, fmt.Sprintf("%s:%d->%d", host, publisher.PublishedPort, publisher.TargetPort))
	}
	if len(published) > 0 {
		return strings.Join(published, ", ")
	}
	return strings.TrimSpace(entry.Ports)
}

// MissingServices returns the sorted manifest services that have no parsed
// container entry.
func MissingServices(manifest ComposeManifest, services []ContainerStatus) []string {
	present := make(map[string]struct{}, len(services))
	for _, service := range services {
		present[service.Service] = struct{}{}
	}
	missing := make([]string, 0, len(manifest.Services))
	for name := range manifest.Services {
		if _, ok := present[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// overallServiceState collapses per-service statuses into a single verdict.
func overallServiceState(services []ContainerStatus) string {
	if len(services) == 0 {
		return "down"
	}
	overall := "healthy"
	for _, service := range services {
		switch service.State {
		case "running":
		case "exited", "missing", "dead":
			return "down"
		default:
			overall = "degraded"
		}
	}
	return overall
}

// InspectDeploymentStatus retrieves the stored deployment record and inspects
// its deployed Compose services over SSH without mutating any state.
func InspectDeploymentStatus(ctx context.Context, app core.App, deploymentID, secretKeyEnv string) (DeploymentInspection, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if app == nil {
		return DeploymentInspection{}, errors.New("PocketBase app is required")
	}
	if !deploymentIDPattern.MatchString(deploymentID) {
		return DeploymentInspection{}, errors.New("deployment ID is invalid")
	}
	deployment, err := app.FindFirstRecordByData("arsy_deployments", "deployment_id", deploymentID)
	if err != nil {
		return DeploymentInspection{}, fmt.Errorf("deployment not found: %w", err)
	}
	application, err := app.FindRecordById("arsy_applications", deployment.GetString("application"))
	if err != nil {
		return DeploymentInspection{}, fmt.Errorf("load deployment application: %w", err)
	}
	inspection := DeploymentInspection{
		DeploymentID:         deployment.GetString("deployment_id"),
		ApplicationSlug:      application.GetString("slug"),
		Status:               deployment.GetString("status"),
		Summary:              deployment.GetString("summary"),
		ManifestChecksum:     deployment.GetString("manifest_checksum"),
		PreviousDeploymentID: deployment.GetString("previous_deployment_id"),
		StartedAt:            deployment.GetDateTime("started_at").String(),
		FinishedAt:           deployment.GetDateTime("finished_at").String(),
		Services:             []ContainerStatus{},
	}
	server, err := app.FindRecordById("arsy_servers", application.GetString("server"))
	if err != nil {
		return DeploymentInspection{}, fmt.Errorf("load application server: %w", err)
	}
	pinnedFingerprint := server.GetString("host_key_fingerprint")
	if pinnedFingerprint == "" {
		return DeploymentInspection{}, errors.New("target server has no confirmed SSH host-key fingerprint")
	}
	privateKey, err := servers.LoadPrivateKey(server, secretKeyEnv)
	if err != nil {
		return DeploymentInspection{}, errors.New("stored SSH credential could not be decrypted")
	}
	defer clear(privateKey)
	address := net.JoinHostPort(server.GetString("host"), strconv.Itoa(server.GetInt("ssh_port")))
	sshCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	client, err := servers.NewSSHClient(sshCtx, address, server.GetString("ssh_username"), privateKey, pinnedFingerprint, 10*time.Second)
	cancel()
	if err != nil {
		return DeploymentInspection{}, errors.New("SSH authentication or pinned host-key verification failed")
	}
	defer client.Close()

	manifestPath, err := ManifestPath(application.GetString("slug"), deploymentID)
	if err != nil {
		return DeploymentInspection{}, err
	}
	executor := servers.Executor{Timeout: 45 * time.Second, MaxOutputBytes: 32 * 1024}
	inspectionResult, err := executor.InspectComposeManifest(ctx, client, manifestPath)
	if err != nil || inspectionResult.ExitCode != 0 {
		return DeploymentInspection{}, errors.New("unable to inspect deployed Compose services")
	}
	manifest, err := ValidateCompose(application.GetString("compose_manifest"))
	if err != nil {
		return DeploymentInspection{}, fmt.Errorf("application Compose manifest is invalid: %w", err)
	}
	var services []ContainerStatus
	if strings.TrimSpace(inspectionResult.Stdout) != "" {
		services, err = ParseComposePS(inspectionResult.Stdout)
		if err != nil {
			return DeploymentInspection{}, err
		}
	}
	if services == nil {
		services = []ContainerStatus{}
	}
	for _, name := range MissingServices(manifest, services) {
		services = append(services, ContainerStatus{Service: name, State: "missing"})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Service < services[j].Service })
	inspection.Services = services
	inspection.Overall = overallServiceState(services)
	return inspection, nil
}
