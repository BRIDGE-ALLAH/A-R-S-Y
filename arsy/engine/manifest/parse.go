package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ParseFile reads and validates a manifest from a JSON file.
func ParseFile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return Parse(data)
}

// Parse validates a manifest from JSON bytes.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest json: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// LoadModuleManifest looks for manifest.json in a module directory.
func LoadModuleManifest(dir string) (*Manifest, error) {
	return ParseFile(filepath.Join(dir, "manifest.json"))
}
