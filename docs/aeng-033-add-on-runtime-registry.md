# AENG-033: Implement Add-On Runtime Registry and Lifecycle

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-001, AENG-019

## Purpose
Implement the runtime registry that tracks all loaded AENG modules, their models, views, actions, and security metadata at application startup. This enables module discovery, dependency resolution, and runtime access.

---

## 1. Registry Structure

### 1.1 Top-Level Registry
```go
type ModelRegistry struct {
    mu              sync.RWMutex
    modules         map[string]*ModuleInfo
    modelByName     map[string]*ModelInfo
    actionByName    map[string]*ActionInfo
    securityByModule map[string]*ModuleSecurity
}
```

### 1.2 Module Information
```go
type ModuleInfo struct {
    ID          string          // Module identifier (e.g., "base", "crm")
    Version      string          // Module version (e.g., "v1.2.3")
    ManifestPath  string          // Path to module manifest file
    Manifest      *ModuleManifest // Parsed manifest
    IsLoaded      bool            // Whether manifest has been parsed
    Dependencies  []string        // Other module IDs this depends on
    LoadedAt      time.Time       // When module was loaded
    Error         error           // Load error, if any
}
```

### 1.3 Model Information
```go
type ModelInfo struct {
    ModuleID     string          // Parent module
    ModelName    string          // Go type/model name
    Collection   string          // PocketBase collection name
    Fields       []FieldInfo     // Model fields
    Relations    []RelationInfo  // Declared relations
    Security     *ResourceMeta   // Resource-level security
}

// Field information for a model
type FieldInfo struct {
    Name      string
    GoType    string
    PocketBaseType string
    Required  bool
    Unique    bool
}

// Relation information
type RelationInfo struct {
    Name      string
    Type      string // "has_one", "has_many", "many_to_many"
    Target    string // Target model name
    FK        string // Foreign key field
    OwnerKey  string // Owner key field
}
```

### 1.4 Action Information
```go
type ActionInfo struct {
    ModuleID    string
    ActionName  string // e.g. "archive", "publish"
    ActionDef   *Action // Full action definition
    IsLoaded    bool
    Error       error
}
```

### 1.5 Security Information
```go
type ModuleSecurity struct {
    Module      string
    Version     string
    Resources   []ResourceMeta
    Roles       []RoleMeta
    Actions     []ActionSecurity
}
```

---

## 2. Registry Operations

### 2.1 Load Module Manifest
```go
func (r *ModelRegistry) LoadModule(ctx context.Context, moduleID string) error {
    // 1. Find manifest file for module
    manifestPath := findManifestPath(moduleID)
    if manifestPath == "" {
        return fmt.Errorf("manifest not found for module %s", moduleID)
    }
    
    // 2. Read and parse manifest
    data, err := os.ReadFile(manifestPath)
    if err != nil {
        return err
    }
    
    var manifest ModuleManifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return err
    }
    
    // 3. Validate manifest
    if err := validateManifest(&manifest); err != nil {
        return err
    }
    
    // 4. Store in registry
    r.mu.Lock()
    r.modules[moduleID] = &ModuleInfo{
        ID:          moduleID,
        Version:     manifest.Version,
        ManifestPath: manifestPath,
        Manifest:    &manifest,
        LoadedAt:    time.Now(),
    }
    r.mu.Unlock()
    
    // 5. Index models and actions
    r.indexModels(&manifest)
    r.indexActions(&manifest)
    r.indexSecurity(&manifest)
    
    return nil
}
```

### 2.2 Index Models
```go
func (r *ModelRegistry) indexModels(m *ModuleManifest) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for _, model := range m.Models {
        r.modelByName[model.Name] = &ModelInfo{
            ModuleID:     m.ID,
            ModelName:    model.Name,
            Collection:   model.Collection,
            Fields:       indexFields(model.Fields),
            Relations:    indexRelations(model.Relations),
            Security:     indexResourceMeta(m.Security, model.Name),
        }
    }
}
```

### 2.3 Index Actions
```go
func (r *ModelRegistry) indexActions(m *ModuleManifest) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for _, action := range m.Actions {
        r.actionByName[action.Name] = &ActionInfo{
            ModuleID: m.ID,
            ActionName: action.Name,
            ActionDef:  &action,
        }
    }
}
```

### 2.4 Index Security
```go
func (r *ModelRegistry) indexSecurity(m *ModuleManifest) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if m.Security != nil {
        r.securityByModule[m.ID] = m.Security
    }
}
```

---

## 3. Runtime Lookup

### 3.1 Get Model by Name
```go
func (r *ModelRegistry) GetModel(name string) (*ModelInfo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    m, ok := r.modelByName[name]
    return m, ok
}
```

### 3.2 Get Action by Name
```go
func (r *ModelRegistry) GetAction(name string) (*ActionInfo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    a, ok := r.actionByName[name]
    return a, ok
}
```

### 3.3 Get Module by ID
```go
func (r *ModelRegistry) GetModule(id string) (*ModuleInfo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    m, ok := r.modules[id]
    return m, ok
}
```

### 3.4 Get Security for Module
```go
func (r *ModelRegistry) GetSecurity(moduleID string) (*ModuleSecurity, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    s, ok := r.securityByModule[moduleID]
    return s, ok
}
```

---

## 4. Module Lifecycle

### 4.1 Module Load Sequence
```
1. Application startup
2. ModelRegistry.Initialize() - load all built-in modules
3. For each add-on module:
   a. ModuleRegistry.LoadModule(ctx, moduleID)
   b. ValidateManifest(manifest)
   c. Register module routes (if HTTP API)
   d. Register module migrations
   e. Run seeds (if configured)
4. All modules loaded - registry ready for queries
```

### 4.2 Module Unload (Optional)
```go
func (r *ModelRegistry) UnloadModule(moduleID string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // 1. Check if any requests are using this module
    if !r.isModuleInUse(moduleID) {
        // 2. Remove from registry
        delete(r.modules, moduleID)
        delete(r.modelByName, /* ... */)
        delete(r.actionByName, /* ... */)
        delete(r.securityByModule, moduleID)
        return nil
    }
    
    return fmt.Errorf("cannot unload module: in use by active requests")
}
```

### 4.3 Hot Reload (Future)
- Not required for v1
- Would require rebuilding models, re-indexing, etc.

---

## 5. Initialization Sequence

### 5.1 Built-In Modules
```go
func InitializeBuiltInModules(ctx context.Context, r *ModelRegistry) error {
    // Load base module (always present)
    if err := r.LoadModule(ctx, "base"); err != nil {
        return err
    }
    
    // Load core modules
    for _, modID := range []string{"tenancy", "httpapi", "migrations"} {
        if err := r.LoadModule(ctx, modID); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 5.2 Add-On Module Loading
```go
func LoadAddOnModules(ctx context.Context, r *ModelRegistry, addOnDir string) error {
    // Read all directories in addOnDir
    entries, err := os.ReadDir(addOnDir)
    if err != nil {
        return err
    }
    
    for _, entry := range entries {
        if entry.IsDir() {
            // Each directory is a module
            modID := entry.Name()
            
            // Try to load manifest
            manifestPath := filepath.Join(addOnDir, modID, "manifest.json")
            if _, err := os.Stat(manifestPath); err != nil {
                // No manifest - skip or error
                continue
            }
            
            if err := r.LoadModule(ctx, modID); err != nil {
                // Log error but continue with other modules
                log.Printf("warning: failed to load module %s: %v", modID, err)
                continue
            }
            
            log.Printf("info: loaded add-on module %s", modID)
        }
    }
    
    return nil
}
```

---

## 6. Dependency Resolution

### 5.1 Topological Sort
```go
func (r *ModelRegistry) ResolveDependencies() error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Build dependency graph
    var sorted []string
    visited := make(map[string]bool)
    visiting := make(map[string]bool)
    
    var visit func(moduleID string) error
    visit = func(moduleID string) error {
        if visited[moduleID] {
            return nil // already processed
        }
        if visiting[moduleID] {
            return fmt.Errorf("circular dependency involving %s", moduleID)
        }
        
        visiting[moduleID] = true
        
        // Get module and its dependencies
        mod, ok := r.modules[moduleID]
        if !ok {
            delete(visiting, moduleID)
            return fmt.Errorf("module %s not found", moduleID)
        }
        
        // Visit dependencies first
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
    
    // Visit all loaded modules
    for id := range r.modules {
        if err := visit(id); err != nil {
            return err
        }
    }
    
    // Store sorted order
    r.loadOrder = sorted
    return nil
}
```

### 5.2 Execution Order
```go
func ExecuteInOrder(r *ModelRegistry) error {
    for _, modID := range r.loadOrder {
        mod := r.modules[modID]
        // Execute module initialization
        if err := mod.Initialize(); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 6. Integration with HTTP API

### 6.1 Route Registration
```go
func RegisterModuleRoutes(r *ModelRegistry, api *router.RouterGroup) error {
    // For each loaded module, register its routes
    for _, mod := range r.getSortedModules() {
        // Load module manifest
        manifest := mod.Manifest
        if manifest == nil {
            continue
        }
        
        // Register views/actions routes
        for _, view := range manifest.Views {
            // Register view endpoint
            api.GET(view.Name, requirePermission(view.Name, "read"))
        }
        
        for _, action := range manifest.Actions {
            // Register action endpoint
            api.POST("/actions/"+action.Name, requirePermission(action.Name, "execute"))
        }
    }
    
    return nil
}
```

### 6.2 Security Middleware
```go
func SecurityMiddleware(r *ModelRegistry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Resolve tenant
            tenant := resolveTenant(r)
            
            // 2. Look up model/action from registry
            // 3. Check permissions
            // 4. Continue to handler
        })
    }
}
```

---

## 6. Dependency Management

### 7.1 Module Dependencies
```go
type ModuleDependency struct {
    ModuleID      string // Module that this module depends on
    DependencyType string // "schema", "data", "api", "security"
    Optional      bool   // If true, error if missing; if false, warn
}
```

### 7.2 Dependency Graph
```
base (v1.2.0)
  └─ tenancy (v1.1.0) - depends on base schema
     └─ httpapi (v1.0.0) - depends on tenancy API
        └─ crm (v1.0.0) - depends on base + httpapi
```

### 5.2 Dependency Validation
```go
func ValidateDependencies(r *ModelRegistry) error {
    for _, mod := range r.getSortedModules() {
        for _, dep := range mod.Dependencies {
            if _, ok := r.modules[dep]; !ok {
                return fmt.Errorf("module %s depends on missing module %s", mod.ID, dep)
            }
        }
    }
    return nil
}
```

---

## 7. Open Issues

1. **Manifest format:** JSON vs TOML vs Go struct (Go struct preferred for type safety)
2. **Circular dependency detection:** How strict should we be? Warn vs error?
3. **Lazy vs eager loading:** Load all modules at startup or on-demand?
4. **Module version conflicts:** What if two modules declare same version?
5. **Module removal:** Clean unload vs just disabling?

---

## 6. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-019 Module Security Metadata](aeng-019-module-security-metadata.md)
- [AENG-020 View/Action Manifest](aeng-020-view-action-manifest.md)
- [PocketBase Module System](https://pocketbase.io/docs/modules/)