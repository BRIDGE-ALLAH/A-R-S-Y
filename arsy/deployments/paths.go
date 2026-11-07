package deployments

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

var (
	applicationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	deploymentIDPattern   = regexp.MustCompile(`^[a-f0-9]{32}$`)
)

const remoteApplicationRoot = "/opt/arsy/apps"

// NewDeploymentID creates an unpredictable identifier without accepting
// deployment IDs from a web request.
func NewDeploymentID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate deployment ID: %w", err)
	}
	return hex.EncodeToString(value), nil
}

// ReleasePath builds the only supported remote release path from validated
// application slugs and internally generated deployment IDs.
func ReleasePath(applicationSlug, deploymentID string) (string, error) {
	if !applicationSlugPattern.MatchString(applicationSlug) {
		return "", errors.New("invalid application slug")
	}
	if !deploymentIDPattern.MatchString(deploymentID) {
		return "", errors.New("invalid deployment ID")
	}
	return fmt.Sprintf("%s/%s/releases/%s", remoteApplicationRoot, applicationSlug, deploymentID), nil
}

// ManifestPath returns the fixed Compose manifest location within a release.
func ManifestPath(applicationSlug, deploymentID string) (string, error) {
	releasePath, err := ReleasePath(applicationSlug, deploymentID)
	if err != nil {
		return "", err
	}
	return releasePath + "/compose.yaml", nil
}
