package tests

import (
	"context"
	"testing"
)

func TestDockerEnabledFlagRespected(t *testing.T) {
	if !IsDockerEnabled() {
		t.Skip("Docker testing not enabled (set ARSY_ENABLE_DOCKER_TESTS=true)")
	}
}

func TestDockerHelperCleanupIsSafe(t *testing.T) {
	if !IsDockerEnabled() {
		t.Skip("Docker testing not enabled")
	}
	ctx := context.Background()
	helper := &DockerTestHelper{}
	if err := helper.Cleanup(ctx); err != nil {
		t.Fatalf("cleanup on nil container: %v", err)
	}
	helper.ContainerID = ""
	if err := helper.Cleanup(ctx); err != nil {
		t.Fatalf("cleanup on empty container: %v", err)
	}
}
