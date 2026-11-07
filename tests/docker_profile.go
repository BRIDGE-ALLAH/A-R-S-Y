package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DockerTestHelper provides opt-in Docker integration for local testing.
type DockerTestHelper struct {
	ContainerID string
	Port        string
}

// IsDockerEnabled returns true if Docker testing is opt-in enabled.
func IsDockerEnabled() bool {
	return os.Getenv("ARSY_ENABLE_DOCKER_TESTS") == "true"
}

// SpinUpTestContainer creates a disposable test container.
// Returns helper and port, or error if Docker unavailable.
func SpinUpTestContainer(ctx context.Context, image string) (*DockerTestHelper, error) {
	if !IsDockerEnabled() {
		return nil, fmt.Errorf("Docker testing not enabled; set ARSY_ENABLE_DOCKER_TESTS=true")
	}

	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "--rm", "-p", "0:5432", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker run: %w (%s)", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return nil, fmt.Errorf("docker run returned empty container ID")
	}

	return &DockerTestHelper{ContainerID: containerID, Port: "5432"}, nil
}

// Cleanup removes the container. Safe to call multiple times.
func (h *DockerTestHelper) Cleanup(ctx context.Context) error {
	if h == nil || h.ContainerID == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, "docker", "stop", h.ContainerID)
	return cmd.Run()
}
