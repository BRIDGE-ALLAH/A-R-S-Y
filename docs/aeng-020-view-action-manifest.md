# AENG-020: Define Metadata-Driven View and Action Manifest

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-019

## Purpose
Define the metadata-driven manifest for views and actions that AENG modules use to declare computed fields, business actions, and their contracts. This enables declarative UI generation and runtime action execution.

---

## 1. View Manifest

### 1.1 View Definition
```go
type View struct {
    Name       string          // View name, e.g. "company_summary"
    Description string          // Human-readable description
    Model      string          // Source model/collection
    Fields     []ViewField     // Computed/selected fields
    Filter     *Filter         // Optional default filter
    Sort       *Sort           // Optional default sort
    Pagination *Pagination     // Optional pagination config
}

// ViewField represents a computed or selected field
type ViewField struct {
    Name       string   // Field name in the view
    Source     string   // Source field or expression
    Type       string   // Go type for rendering
    Label      string   // Display label
    Format     string   // Formatting (date, currency, percentage)
    Hidden     bool     // Should not be shown in UI
    ReadOnly   bool     // Cannot be edited
    ComputeFn  string   // Function name for computation
    Dependencies []string // Other fields this depends on
}
```

### 1.2 Example: Company Summary View
```go
var CompanySummaryView = &View{
    Name:        "company_summary",
    Description: "Summary view for company listings",
    Model:       "companies",
    Fields: []ViewField{
        {
            Name:        "id",
            Source:      "id",
            Type:        "uuid",
            Label:       "ID",
            Hidden:      false,
        },
        {
            Name:        "name",
            Source:      "name",
            Type:        "string",
            Label:       "Name",
            Hidden:      false,
        },
        {
            Name:        "plan",
            Source:      "plan",
            Type:        "string",
            Label:       "Plan",
            Hidden:      false,
        },
        {
            Name:        "employee_count",
            Source:      `SELECT count(*) FROM employees WHERE company_id = companies.id`,
            Type:        "integer",
            Label:       "Employee Count",
            Hidden:      false,
            ComputeFn:   "computeEmployeeCount",
        },
        {
            Name:        "created_formatted",
            Source:      "created",
            Type:        "string",
            Label:       "Created",
            Hidden:      false,
            Format:      "date",
            Hidden:      false,
        },
    },
    Filter: &Filter{
        // Default filter: only active companies
        Conditions: []Condition{
            {Field: "status", Operator: "eq", Value: "active"},
        },
    },
    Pagination: &Pagination{
        DefaultPage:  1,
        DefaultLimit: 20,
        MaxLimit:     100,
    },
}
```

### 1.3 Filter Condition
```go
type Condition struct {
    Field      string   // Field name
    Operator   string   // "eq", "neq", "gt", "lt", "gte", "lte", "contains", "starts_with", "ends_with"
    Value      any      // Value to compare
    Logic      string   // "and", "or" for combining conditions
}
```

### 1.4 Sort Configuration
```go
type Sort struct {
    Field string   // Field to sort by
    Order string   // "asc" or "desc"
}
```

### 1.5 Pagination Configuration
```go
type Pagination struct {
    DefaultPage   int   // Default page number
    DefaultLimit int  // Default items per page
    MaxLimit    int  // Maximum items per page allowed
    TotalField  string // Field name for total count (if different)
}
```

---

## 2. Action Manifest

### 2.1 Action Definition
```go
type Action struct {
    Name        string          // Action name, e.g. "archive", "publish"
    Description string          // Human-readable description
    Group       string          // UI group, e.g. "actions", "utilities"
    Icon        string          // Icon name/URL
    Confirmation  *Confirmation   // Confirmation dialog config
    Access      *AccessControl  // Who can execute this action
    Input       *ActionInput    // Required input parameters
    Output      *ActionOutput   // Returned output
    ComputeFn   string          // Server-side computation function name
    Mutates     bool            // Does this mutate state?
    ReadOnly    bool            // Does this only read?
}

// Confirmation dialog config
type Confirmation struct {
    Title       string   // Dialog title
    Message     string   // Dialog message
    ConfirmLabel string  // Confirm button label (default: "Confirm")
    CancelLabel  string   // Cancel button label (default: "Cancel")
}

// Action input parameter
type ActionInput struct {
    Fields map[string]any   // Input fields/parameters
    Required []string         // Required fields
    Validation ValidationConfig // Validation rules
}

// Action output result
type ActionOutput struct {
    Success  bool     `json:"success"`
    Message  string   `json:"message,omitempty"`
    Data     any      `json:"data,omitempty"`
}
```

### 2.2 Example: Company Archive Action
```go
var CompanyArchiveAction = &Action{
    Name:        "archive",
    Description: "Soft-archive a company (set status to archived)",
    Group:       "actions",
    Icon:        "archive",
    Confirmation: &Confirmation{
        Title:       "Archive Company",
        Message:     "Are you sure you want to archive this company?",
        ConfirmLabel: "Archive",
        CancelLabel: "Cancel",
    },
    Access: &AccessControl{
        Roles: []string{"admin", "manager"},
    },
    Input: &ActionInput{
        Fields: map[string]any{
            "confirm": true, // from confirmation dialog
        },
        Required: []string{"confirm"},
        Validation: ValidationConfig{
            FieldRules: map[string]ValidationRule{
                "confirm": {Required: true, OneOf: []bool{true}},
            },
        },
    },
    Output: &ActionOutput{
        Success: true,
        Message: "Company archived successfully",
    },
    Mutates:   true,
    ReadOnly:  false,
    ComputeFn: "computeCompanyArchive",
}
```

### 2.2 Action Input Validation
```go
type ValidationRule struct {
    Required   bool
    OneOf      []bool     // Allow only these values
    Min        any        // Minimum value (for numbers/length)
    Max        any        // Maximum value (for numbers/length)
    Pattern    string     // Regex pattern
    Format     string     // "email", "date", "uuid"
}

// Example validation for archive action
ValidationRule{
    Required: true,
    OneOf:    []bool{true}, // confirm must be true
}
```

### 2.3 Action Compute Function
```go
func computeCompanyArchive(ctx context.Context, input map[string]any, tenant *core.Record) (ActionOutput, error) {
    // 1. Validate input
    if !input["confirm"].(bool) {
        return ActionOutput{
            Success: false,
            Message: "Confirmation required",
        }, nil
    }
    
    // 2. Set status to archived
    tenant.Set("status", "archived")
    if err := tenant.Save(); err != nil {
        return ActionOutput{
            Success: false,
            Message: "Failed to archive company: " + err.Error(),
        }, nil
    }
    
    return ActionOutput{
        Success: true,
        Message: "Company archived successfully",
    }, nil
}
```

---

## 2.3 Action Execution API

### 2.3.1 Programmatic Execution
```go
func ExecuteAction(ctx context.Context, actionName string, input map[string]any, tenant *core.Record) (ActionOutput, error) {
    // 1. Look up action definition
    action := lookupAction(actionName)
    if action == nil {
        return ActionOutput{}, fmt.Errorf("action %s not found", actionName)
    }
    
    // 2. Check permissions
    if !checkActionPermission(ctx, action, tenant) {
        return ActionOutput{
            Success: false,
            Message: "Permission denied",
        }, nil
    }
    
    // 3. Execute compute function
    fn := action.ComputeFn
    if fn != "" {
        // Resolve and call the function
        f := getFunction(fn)
        return f(ctx, input, tenant)
    }
    
    // 4. Default: return success with message
    return ActionOutput{
        Success: true,
        Message: "Action executed",
    }, nil
}
```

### 2.3.2 HTTP Action Endpoint
```go
func CompanyActionHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Parse action name from path
    actionName := strings.TrimSuffix(r.URL.Path, "/archive")
    
    // 2. Parse input from body
    var input map[string]any
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "invalid input", http.StatusBadRequest)
        return
    }
    
    // 3. Execute action
    output, err := ExecuteAction(r.Context(), actionName, input, tenant)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 4. Return result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(output)
}
```

### 2.3.3 Example HTTP Request
```
POST /actions/companies/archive
Content-Type: application/json

{
    "confirm": true
}
```

### 2.3.4 Example HTTP Response
```json
{
    "success": true,
    "message": "Company archived successfully"
}
```

---

## 3. View + Action Integration

### 3.1 Combined Module Manifest
```go
type ModuleManifest struct {
    Module      string          `json:"module"`
    Version     string          `json:"version"`
    Security    *ModuleSecurity `json:"security,omitempty"`
    Models      []Model         `json:"models"`
    Views       []*View         `json:"views,omitempty"`
    Actions     []*Action       `json:"actions,omitempty"`
    Relations   []*Relation     `json:"relations,omitempty"`
    Seeds       []*Seed         `json:"seeds,omitempty"`
}
```

### 3.2 UI Generation from Manifest
| Component | Generated from |
|-----------|----------------|
| Column display | View.Fields |
| Filter UI | View.Filter.Conditions |
| Sort UI | View.Sort |
| Pagination controls | View.Pagination |
| Action buttons | Action.Group, Action.Icon, Action.Name |
| Confirmation dialogs | Action.Confirmation |
| Permission checks | Action.Access.Roles, Resource.FieldPermissions |

### 3.2 Example UI Integration
```html
<!-- Company list page -->
<div class="company-list">
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Plan</th>
                <th>Employees</th>
                <th>Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range companies}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Plan}}</td>
                <td>{{.EmployeeCount}}</td>
                <td>
                    {{if can "admin"}}
                    <button class="action-btn" data-action="archive">Archive</button>
                    {{end}}
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
    
    <!-- Pagination -->
    <div class="pagination">
        <button>{{.Page - 1}}</button>
        <span>{{.Page}}/{{.TotalPages}}</span>
        <button>{{.Page + 1}}</button>
    </div>
</div>
```

---

## 4. Action Permission Check

### 4.1 Permission Logic
```go
func checkActionPermission(ctx context.Context, action *Action, tenant *core.Record) bool {
    // 1. Superuser bypass
    if tenant.App.HasSuperuserAuth() {
        return true
    }
    
    // 2. Check roles from action access
    userRole := getUserRole(ctx)
    allowed := false
    for _, role := range action.Access.Roles {
        if userRole == role || contains(role, "superuser") {
            allowed = true
            break
        }
    }
    
    // 3. Check field-level permissions from resource security
    if allowed {
        for _, fp := range action.Permissions {
            if !checkFieldPermission(ctx, fp, tenant) {
                allowed = false
                break
            }
        }
    }
    
    return allowed
}
```

### 4.2 Field Permission Check
```go
func checkFieldPermission(ctx context.Context, fp FieldPermission, tenant *core.Record) bool {
    // Evaluate conditions
    if len(fp.Conditions) == 0 {
        return fp.Effect == EffectAllow
    }
    
    for _, cond := range fp.Conditions {
        // Evaluate condition against tenant data
        met := evaluateCondition(ctx, cond, tenant)
        if fp.Effect == EffectAllow && !met {
            return false // condition not met, deny
        }
        if fp.Effect == EffectDeny && met {
            return false // condition met, deny
        }
    }
    
    return fp.Effect == EffectAllow
}
```

---

## 5. Validation

### 5.1 View + Action Manifest Validation
```go
func ValidateManifest(m *ModuleManifest) error {
    var errors []string
    
    // 1. Check all view source fields exist in model
    for _, v := range m.Views {
        for _, f := range v.Fields {
            if !fieldExistsInModel(f.Source, m.Models) {
                errors = append(errors, fmt.Sprintf("view %s: field %s not found in model", v.Name, f.Source))
            }
        }
    }
    
    // 2. Check all action input fields validate
    for _, a := range m.Actions {
        if a.Input != nil {
            if err := validateActionInput(a.Input); err != nil {
                errors = append(errors, fmt.Sprintf("action %s: %v", a.Name, err))
            }
        }
        
        // 3. Check action compute function exists
        if action.ComputeFn != "" && !functionExists(action.ComputeFn) {
            errors = append(errors, fmt.Sprintf("action %s: compute function %s not found", a.Name, action.ComputeFn))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("manifest validation failed: %v", errors)
    }
    
    return nil
}
```

---

## 6. Example: Full Module Manifest

```go
var BaseModule = &ModuleManifest{
    Module:   "base",
    Version:  "v1.0.0",
    Security: CompanySecurity,
    Models: []Model{...},
    Views: []*View{CompanySummaryView},
    Actions: []*Action{CompanyArchiveAction, CompanyReactivateAction},
    Seeds:    []*Seed{DefaultSeed},
}
```

### 6.1 Generated UI from This Manifest
```
Company List Page:
- Columns: Name, Plan, Employee Count (from View)
- Filters: Status dropdown (from View.Filter)
- Sort: By Name asc/desc (from View.Sort)
- Pagination: 20 per page (from View.Pagination)
- Actions: Archive button (visible for admin/manager) (from Action)
- Confirmation: "Are you sure?" modal (from Action.Confirmation)
```

---

## 7. Open Issues

1. **View-SQL translation:** How are View.Filter conditions translated to PocketBase queries?
2. **Action side effects:** What happens to related records when an action runs? (cascade, manual?)
3. **Action versioning:** Should actions be versioned like migrations?
4. **Bulk actions:** Can actions operate on multiple records at once?
5. **Action history:** Should action executions be logged/auditable?

---

## 8. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-019 Module Security Metadata](aeng-019-module-security-metadata.md)
- [PocketBase Actions](https://pocketbase.io/docs/actions/)
- [Vue/React component design patterns](https://vuejs.org/guide/)