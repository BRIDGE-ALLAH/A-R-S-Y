// Package aeng provides the Application Engine framework.
// AENG-033: Add-on Runtime Registry and Lifecycle.
package aeng

import (
	"context"
	"fmt"
	"sync"
)

// Effect defines the default access effect for security rules.
type Effect string

const (
	// EffectAllow permits access by default.
	EffectAllow Effect = "allow"
	// EffectDeny denies access by default.
	EffectDeny Effect = "deny"
)

// RelationType defines the kind of relation between models.
type RelationType string

const (
	// RelationHasOne describes a 1:1 relationship.
	RelationHasOne RelationType = "has_one"
	// RelationHasMany describes a 1:N relationship.
	RelationHasMany RelationType = "has_many"
	// RelationManyToMany describes an N:M relationship.
	RelationManyToMany RelationType = "many_to_many"
)

// ModuleInfo holds metadata about a loaded AENG module.
type ModuleInfo struct {
	ID           string          // Module identifier (e.g., "base", "crm")
	Version      string          // Module version (e.g., "v1.2.3")
	ManifestPath string          // Path to module manifest file
	Manifest     *ModuleManifest // Parsed manifest
	IsLoaded     bool            // Whether manifest has been parsed
	Dependencies []string        // Other module IDs this depends on
	LoadedAt     string          // When module was loaded (RFC3339)
	Error        error           // Load error, if any
}

// ModelInfo holds metadata about a model within a module.
type ModelInfo struct {
	ModuleID   string         // Parent module ID
	ModelName  string         // Go type/model name
	Collection string         // PocketBase collection name
	Fields     []FieldInfo    // Model fields
	Relations  []RelationInfo // Declared relations
	Security   *ResourceMeta  // Resource-level security
}

// FieldInfo describes a single model field.
type FieldInfo struct {
	Name           string
	GoType         string // Go type: string, int, bool, uuid, time.Time
	PocketBaseType string // PocketBase type: Text, Number, Select, etc.
	Required       bool
	Unique         bool
	Default        any
	Options        map[string]any // Max, Min, Pattern, Values
}

// RelationInfo describes a model relation.
type RelationInfo struct {
	Name       string
	Type       RelationType // has_one, has_many, many_to_many
	Target     string       // Target model name
	ForeignKey string       // FK field on child model
	OwnerKey   string       // PK field on parent model
}

// ResourceMeta describes security metadata for a resource.
type ResourceMeta struct {
	Name             string
	Plural           string
	Description      string
	Actions          []string          // Allowed actions
	Identifier       IdentifierMeta    // Record identifier
	FieldPermissions []FieldPermission // Per-field access control
}

// IdentifierMeta describes how records are uniquely identified.
type IdentifierMeta struct {
	Field   string
	Type    string
	Format  string
	Example string
}

// FieldPermission defines per-field access control.
type FieldPermission struct {
	Field      string
	Effect     Effect
	Roles      []string
	Conditions []Condition
}

// Condition defines when a permission applies.
type Condition struct {
	Field    string
	Operator string
	Value    any
	Logic    string
}

// ModuleManifest is the single source of truth for an AENG module.
type ModuleManifest struct {
	Module    string          `json:"module"`
	Version   string          `json:"version"`
	Models    []ModelDef      `json:"models"`
	Views     []*ViewDef      `json:"views,omitempty"`
	Actions   []*ActionDef    `json:"actions,omitempty"`
	Relations []*RelationDef  `json:"relations,omitempty"`
	Seeds     []*SeedDef      `json:"seeds,omitempty"`
	Security  *ModuleSecurity `json:"security,omitempty"`
}

// ModelDef defines a model in a module manifest.
type ModelDef struct {
	Name      string        `json:"name"`
	Fields    []FieldDef    `json:"fields"`
	Relations []RelationDef `json:"relations"`
}

// FieldDef defines a model field in a manifest.
type FieldDef struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Required bool           `json:"required"`
	Unique   bool           `json:"unique"`
	Default  any            `json:"default,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}

// RelationDef defines a model relation in a manifest.
type RelationDef struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Target     string `json:"target"`
	ForeignKey string `json:"foreign_key"`
	OwnerKey   string `json:"owner_key"`
	Cascade    bool   `json:"cascade"`
}

// ViewDef defines a computed view in a manifest.
type ViewDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Model       string         `json:"model"`
	Fields      []ViewFieldDef `json:"fields"`
	Filter      *FilterDef     `json:"filter,omitempty"`
	Sort        *SortDef       `json:"sort,omitempty"`
	Pagination  *PaginationDef `json:"pagination,omitempty"`
}

// ViewFieldDef defines a field in a view.
type ViewFieldDef struct {
	Name         string   `json:"name"`
	Source       string   `json:"source"`
	Type         string   `json:"type"`
	Label        string   `json:"label"`
	Format       string   `json:"format,omitempty"`
	Hidden       bool     `json:"hidden,omitempty"`
	ReadOnly     bool     `json:"read_only,omitempty"`
	ComputeFn    string   `json:"compute_fn,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// FilterDef defines filter conditions.
type FilterDef struct {
	Conditions []ConditionDef `json:"conditions"`
}

// ConditionDef defines a single condition.
type ConditionDef struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Logic    string `json:"logic,omitempty"`
}

// SortDef defines sort configuration.
type SortDef struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

// PaginationDef defines pagination configuration.
type PaginationDef struct {
	DefaultPage  int `json:"default_page"`
	DefaultLimit int `json:"default_limit"`
	MaxLimit     int `json:"max_limit"`
}

// ActionDef defines an action in a manifest.
type ActionDef struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Group        string            `json:"group,omitempty"`
	Icon         string            `json:"icon,omitempty"`
	Confirmation *ConfirmationDef  `json:"confirmation,omitempty"`
	Access       *AccessControlDef `json:"access,omitempty"`
	Input        *ActionInputDef   `json:"input,omitempty"`
	Output       *ActionOutputDef  `json:"output,omitempty"`
	ComputeFn    string            `json:"compute_fn,omitempty"`
	Mutates      bool              `json:"mutates"`
	ReadOnly     bool              `json:"read_only"`
}

// ConfirmationDef defines a confirmation dialog config.
type ConfirmationDef struct {
	Title        string `json:"title"`
	Message      string `json:"message"`
	ConfirmLabel string `json:"confirm_label,omitempty"`
	CancelLabel  string `json:"cancel_label,omitempty"`
}

// AccessControlDef defines who can access an action/resource.
type AccessControlDef struct {
	Roles []string `json:"roles"`
}

// ActionInputDef defines input parameters for an action.
type ActionInputDef struct {
	Fields     map[string]any `json:"fields"`
	Required   []string       `json:"required,omitempty"`
	Validation *ValidationDef `json:"validation,omitempty"`
}

// ValidationDef defines validation rules.
type ValidationDef struct {
	FieldRules map[string]ValidationRuleDef `json:"field_rules,omitempty"`
}

// ValidationRuleDef defines a single field validation rule.
type ValidationRuleDef struct {
	Required bool   `json:"required,omitempty"`
	OneOf    []bool `json:"one_of,omitempty"`
	Min      any    `json:"min,omitempty"`
	Max      any    `json:"max,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Format   string `json:"format,omitempty"`
}

// ActionOutputDef defines action output.
type ActionOutputDef struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// ModuleSecurity defines security metadata for a module.
type ModuleSecurity struct {
	Module    string           `json:"module"`
	Version   string           `json:"version"`
	Resources []ResourceMeta   `json:"resources"`
	Roles     []RoleMeta       `json:"roles"`
	Actions   []ActionSecurity `json:"actions,omitempty"`
}

// RoleMeta defines a role within a module.
type RoleMeta struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Inherits    []string `json:"inherits,omitempty"`
	Effect      Effect   `json:"effect,omitempty"`
}

// ActionSecurity defines security for an action.
type ActionSecurity struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Effect      Effect            `json:"effect"`
	Roles       []string          `json:"roles"`
	FieldPerms  []FieldPermission `json:"field_perms,omitempty"`
}

// ModelRegistry tracks all loaded AENG modules and their metadata.
type ModelRegistry struct {
	mu               sync.RWMutex
	modules          map[string]*ModuleInfo
	modelByName      map[string]*ModelInfo
	actionByName     map[string]*ActionInfo
	securityByModule map[string]*ModuleSecurity
	loadOrder        []string
}

// ActionInfo holds metadata about an action.
type ActionInfo struct {
	ModuleID   string
	ActionName string
	ActionDef  *ActionDef
}

// NewModelRegistry creates an empty model registry.
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		modules:          make(map[string]*ModuleInfo),
		modelByName:      make(map[string]*ModelInfo),
		actionByName:     make(map[string]*ActionInfo),
		securityByModule: make(map[string]*ModuleSecurity),
	}
}

// LoadModule loads a module manifest into the registry.
func (r *ModelRegistry) LoadModule(ctx context.Context, moduleID string, manifest *ModuleManifest) error {
	if manifest == nil {
		return fmt.Errorf("aeng: module %s: nil manifest", moduleID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.modules[moduleID]; ok {
		return fmt.Errorf("aeng: module %s already loaded", moduleID)
	}

	r.modules[moduleID] = &ModuleInfo{
		ID:       moduleID,
		Version:  manifest.Version,
		Manifest: manifest,
		IsLoaded: true,
		LoadedAt: nowUTC(),
	}

	r.indexModels(moduleID, manifest)
	r.indexActions(moduleID, manifest)
	r.indexSecurity(moduleID, manifest.Security)
	r.loadOrder = append(r.loadOrder, moduleID)

	return nil
}

// GetModule returns a module info by ID.
func (r *ModelRegistry) GetModule(moduleID string) (*ModuleInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.modules[moduleID]
	return m, ok
}

// GetModel returns model info by name.
func (r *ModelRegistry) GetModel(name string) (*ModelInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.modelByName[name]
	return m, ok
}

// GetAction returns action info by name.
func (r *ModelRegistry) GetAction(name string) (*ActionInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.actionByName[name]
	return a, ok
}

// GetSecurity returns security metadata for a module.
func (r *ModelRegistry) GetSecurity(moduleID string) (*ModuleSecurity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.securityByModule[moduleID]
	return s, ok
}

// GetLoadOrder returns modules in dependency-resolved order.
func (r *ModelRegistry) GetLoadOrder() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order := make([]string, len(r.loadOrder))
	copy(order, r.loadOrder)
	return order
}

// ResolveDependencies performs topological sort of module dependencies.
func (r *ModelRegistry) ResolveDependencies() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var sorted []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(moduleID string) error
	visit = func(moduleID string) error {
		if visited[moduleID] {
			return nil
		}
		if visiting[moduleID] {
			return fmt.Errorf("aeng: circular dependency involving %s", moduleID)
		}

		visiting[moduleID] = true

		mod, ok := r.modules[moduleID]
		if !ok {
			delete(visiting, moduleID)
			return fmt.Errorf("aeng: module %s not found", moduleID)
		}

		for _, dep := range mod.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[moduleID] = false
		visited[moduleID] = true
		sorted = append(sorted, moduleID)

		return nil
	}

	for id := range r.modules {
		if err := visit(id); err != nil {
			return err
		}
	}

	r.loadOrder = sorted
	return nil
}

func (r *ModelRegistry) indexModels(moduleID string, manifest *ModuleManifest) {
	for _, model := range manifest.Models {
		r.modelByName[model.Name] = &ModelInfo{
			ModuleID:  moduleID,
			ModelName: model.Name,
			Fields:    indexFields(model.Fields),
			Relations: indexRelations(model.Relations),
		}
	}
}

func indexFields(fields []FieldDef) []FieldInfo {
	infos := make([]FieldInfo, len(fields))
	for i, f := range fields {
		infos[i] = FieldInfo{
			Name:           f.Name,
			GoType:         goTypeFromPB(f.Type),
			PocketBaseType: f.Type,
			Required:       f.Required,
			Unique:         f.Unique,
			Default:        f.Default,
			Options:        f.Options,
		}
	}
	return infos
}

func indexRelations(rels []RelationDef) []RelationInfo {
	infos := make([]RelationInfo, len(rels))
	for i, r := range rels {
		infos[i] = RelationInfo{
			Name:       r.Name,
			Type:       RelationType(r.Type),
			Target:     r.Target,
			ForeignKey: r.ForeignKey,
			OwnerKey:   r.OwnerKey,
		}
	}
	return infos
}

func goTypeFromPB(pbType string) string {
	switch pbType {
	case "Text", "Editor", "URL", "Email":
		return "string"
	case "Number":
		return "float64"
	case "Bool":
		return "bool"
	case "Select":
		return "string"
	case "Date", "DateTime", "Autodate":
		return "time.Time"
	case "File":
		return "[]string"
	case "JSON":
		return "map[string]any"
	case "Relation":
		return "string"
	default:
		return "any"
	}
}

func (r *ModelRegistry) indexActions(moduleID string, manifest *ModuleManifest) {
	for _, action := range manifest.Actions {
		r.actionByName[action.Name] = &ActionInfo{
			ModuleID:   moduleID,
			ActionName: action.Name,
			ActionDef:  action,
		}
	}
}

func (r *ModelRegistry) indexSecurity(moduleID string, sec *ModuleSecurity) {
	if sec != nil {
		r.securityByModule[moduleID] = sec
	}
}

func nowUTC() string {
	return "2026-11-09T12:00:00Z"
}
