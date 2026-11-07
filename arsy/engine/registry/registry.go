package registry

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// ModelID is the stable, namespaced identifier of an application model.
// Example: "base.company".
type ModelID string

// ModuleID identifies the add-on module that owns a model.
type ModuleID string

// ModelSource identifies how a descriptor entered the runtime. Both sources
// use exactly the same validation and registration contract.
type ModelSource string

const (
	SourceStaticGo ModelSource = "static_go"
	SourceMetadata ModelSource = "metadata"
)

// Descriptor is the runtime identity and storage mapping for one model.
// Field and relation metadata are intentionally added by later engine layers.
type Descriptor struct {
	ID          ModelID
	Label       string
	TableName   string
	Version     string
	OwnerModule ModuleID
	Source      ModelSource
}

// ModelRegistry is the minimal contract consumed by application-engine
// components. Implementations must return stable snapshots and be safe for
// concurrent registration and lookup.
type ModelRegistry interface {
	Register(Descriptor) error
	Lookup(ModelID) (Descriptor, error)
	Snapshot() []Descriptor
}

// InMemoryRegistry stores immutable descriptor values in memory. Its zero
// value is ready for use. It is not a process-wide singleton; callers own its
// lifecycle and may create one per application runtime.
type InMemoryRegistry struct {
	mu     sync.RWMutex
	models map[ModelID]Descriptor
}

// New creates an empty model registry.
func New() *InMemoryRegistry {
	return &InMemoryRegistry{models: make(map[ModelID]Descriptor)}
}

var (
	modelIDPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
	moduleIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*(\.[a-z][a-z0-9_-]*)*$`)
	tablePattern    = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	versionPattern  = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
)

func registryError(id ModelID, cause error, detail string) error {
	return &RegistrationError{ModelID: id, Cause: cause, Detail: detail}
}

func validateDescriptor(d Descriptor) (Descriptor, error) {
	if !modelIDPattern.MatchString(string(d.ID)) {
		return d, registryError(d.ID, ErrInvalidModelID, "expected a namespaced identifier such as base.company")
	}
	if strings.TrimSpace(d.Label) == "" {
		return d, registryError(d.ID, ErrInvalidLabel, "label must not be empty")
	}
	d.Label = strings.TrimSpace(d.Label)
	if !tablePattern.MatchString(d.TableName) {
		return d, registryError(d.ID, ErrInvalidTableName, "table name must be a lowercase SQL identifier")
	}
	if d.Version == "" || !versionPattern.MatchString(d.Version) {
		return d, registryError(d.ID, ErrInvalidVersion, "version must use MAJOR.MINOR.PATCH without leading zeroes")
	}
	if strings.TrimSpace(string(d.OwnerModule)) == "" {
		return d, registryError(d.ID, ErrMissingOwnerModule, "owner module is required")
	}
	if !moduleIDPattern.MatchString(string(d.OwnerModule)) {
		return d, registryError(d.ID, ErrInvalidModuleID, "owner module must be a lowercase module identifier")
	}
	switch d.Source {
	case SourceStaticGo, SourceMetadata:
	default:
		return d, registryError(d.ID, ErrInvalidSource, "source must be static_go or metadata")
	}
	return d, nil
}

// Register validates and registers one descriptor. A model ID can only be
// registered once. Re-registering the same ID with another version is reported
// separately so callers can distinguish duplicate loading from an incompatible
// runtime definition.
func (r *InMemoryRegistry) Register(input Descriptor) error {
	d, err := validateDescriptor(input)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.models == nil {
		r.models = make(map[ModelID]Descriptor)
	}
	if existing, ok := r.models[d.ID]; ok {
		if existing.Version != d.Version {
			return registryError(d.ID, ErrModelVersionConflict,
				fmt.Sprintf("registered version %s; attempted version %s", existing.Version, d.Version))
		}
		return registryError(d.ID, ErrDuplicateModel, "model ID is already registered")
	}
	r.models[d.ID] = d
	return nil
}

// Lookup returns a descriptor by stable model ID.
func (r *InMemoryRegistry) Lookup(id ModelID) (Descriptor, error) {
	if !modelIDPattern.MatchString(string(id)) {
		return Descriptor{}, registryError(id, ErrInvalidModelID, "expected a namespaced identifier such as base.company")
	}

	r.mu.RLock()
	d, ok := r.models[id]
	r.mu.RUnlock()
	if !ok {
		return Descriptor{}, registryError(id, ErrUnknownModel, "model is not registered")
	}
	return d, nil
}

// Snapshot returns a detached, deterministically ordered copy of all models.
// The returned slice can be modified by the caller without affecting the
// registry.
func (r *InMemoryRegistry) Snapshot() []Descriptor {
	r.mu.RLock()
	out := make([]Descriptor, 0, len(r.models))
	for _, d := range r.models {
		out = append(out, d)
	}
	r.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Ensure the concrete implementation satisfies the public contract.
var _ ModelRegistry = (*InMemoryRegistry)(nil)

