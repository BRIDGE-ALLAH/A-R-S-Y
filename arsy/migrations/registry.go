package migrations

// registry.go - Version registry and migration tracking

type MigrationVersion struct {
	Version  string
	File     string
	Up       func(app interface{}) error
	Down     func(app interface{}) error
	Applied  bool
}

// Registry maintains the in-memory registry of discovered migrations
// and tracks which versions have been applied.
type Registry struct {
	migrations      map[string]MigrationVersion
	appliedVersions map[string]bool
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	return &Registry{
		migrations:      make(map[string]MigrationVersion),
		appliedVersions: make(map[string]bool),
	}
}

// Register adds a migration to the registry
func (r *Registry) Register(version string, file string, up func(app interface{}) error, down func(app interface{}) error) {
	r.migrations[version] = MigrationVersion{
		Version:  version,
		File:     file,
		Up:       up,
		Down:     down,
		Applied:  false,
	}
}

// IsApplied checks if a migration version has been applied
func (r *Registry) IsApplied(version string) bool {
	return r.appliedVersions[version]
}

// MarkApplied marks a migration version as applied
func (r *Registry) MarkApplied(version string) {
	r.appliedVersions[version] = true
	if mv, ok := r.migrations[version]; ok {
		mv.Applied = true
		r.migrations[version] = mv
	}
}

// GetLastApplied returns the most recently applied version
func (r *Registry) GetLastApplied() string {
	var lastVersion string
	for version := range r.appliedVersions {
		if lastVersion == "" || version > lastVersion {
			lastVersion = version
		}
	}
	return lastVersion
}

// LoadAll returns all registered migrations
func (r *Registry) LoadAll() []MigrationVersion {
	var result []MigrationVersion
	for _, mv := range r.migrations {
		result = append(result, mv)
	}
	return result
}

