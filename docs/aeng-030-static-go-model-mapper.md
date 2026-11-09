# AENG-030: Implement Static Go Model Persistence Mapper

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-001

## Purpose
Implement a static Go model persistence mapper that generates type-safe Go structs from PocketBase collection schemas, enabling compile-time field access and relation handling for AENG modules.

---

## 1. Design Goals

### 1.1 Requirements
| Goal | Priority | Description |
|------|----------|-------------|
| Type-safe field access | P0 | `model.Company.Name` compiles; `model.Company.Phone` errors if field doesn't exist |
| Multi-tenant routing | P0 | Per-tenant database selection at runtime |
| PocketBase compatibility | P0 | Generated code works with PocketBase API |
| No runtime reflection | P1 | Avoid `reflect` package for performance |
| Codegen from manifest | P1 | Single source of truth in Go manifest |
| Relations as struct embeds | P2 | `Company.Dept()` compiles |
| Seed data integration | P2 | Generated models work with seed runner |

### 1.2 Non-Goals
- Dynamic schema at runtime (PocketBase handles this)
- Full ORM features (migrations, relations handled separately)
- Database agnostic (SQLite/PocketBase focus)

---

## 2. Manifest-Driven Approach

### 2.1 Model Definition in Manifest
```go
type Model struct {
    Name        string   // Go type name + PocketBase collection name
    Collection  string   // PocketBase collection name
    TableName   string   // SQLite table name (for multi-tenant)
    Fields      []Field  // Model fields
    Relations   []Relation // Declared relations
}

// Field defines a model field with type metadata
type Field struct {
    Name      string
    GoType    string   // e.g. "string", "int", "uuid", "bool"
    PocketBaseType string // e.g. "Text", "Number", "Select"
    Required  bool
    Unique    bool
    Default   any
    Options   map[string]any // Max, Min, Pattern, Values
}

// Relation defines a relation to another model
type Relation struct {
    Name       string
    Type       string   // "has_one", "has_many", "many_to_many"
    Target     string   // Target model Go name
    ForeignKey string   // FK field on child model
    OwnerKey   string   // PK field on parent model
}
```

### 2.2 Example Manifest
```go
var CompanyModel = &Model{
    Name:      "Company",
    Collection: "companies",
    Fields: []Field{
        {Name: "ID", GoType: "uuid", PocketBaseType: "Text", Required: true, Unique: true},
        {Name: "Name", GoType: "string", PocketBaseType: "Text", Required: true},
        {Name: "Plan", GoType: "string", PocketBaseType: "Text"},
        {Name: "Status", GoType: "string", PocketBaseType: "Select", Options: map[string]any{"values": []string{"active", "suspended", "archived"}}},
        {Name: "CreatedAt", GoType: "time.Time", PocketBaseType: "Autodate"},
    },
    Relations: []Relation{
        {Name: "Departments", Type: "has_many", Target: "Department", ForeignKey: "company_id", OwnerKey: "id"},
    },
}
```

---

## 2. Code Generator

### 2.1 Generate `models_gen.go`
```go
package base

// Company represents a company in the system
type Company struct {
    ID        uuid.UUID    `db:"id" json:"id"`
    Name      string       `db:"name" json:"name"`
    Plan      *string      `db:"plan" json:"plan,omitempty"`
    Status    string       `db:"status" json:"status"`
    CreatedAt time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

// Department represents a department linked to a company
type Department struct {
    ID        uuid.UUID    `db:"id" json:"id"`
    Name      string       `db:"name" json:"name"`
    CompanyID uuid.UUID    `db:"company_id" json:"company_id"`
    Company   *Company     `db:"-" json:"company,omitempty"` // embedded relation
}

// CompanyService provides type-safe operations on Company model
type CompanyService struct {
    db     *sql.DB
    registry *ModelRegistry
}
```

### 2.3 Generated Query Builder
```go
// CompanyQuery is a type-safe query for Company
type CompanyQuery struct {
    db     *sql.DB
    filter map[string]any
    limit  int
    offset int
}

// FilterByStatus filters companies by status
func (q *CompanyQuery) FilterByStatus(status string) *CompanyQuery {
    q.filter["status"] = status
    return q
}

// All returns all companies matching the current filters
func (q *CompanyQuery) All(ctx context.Context) ([]*Company, error) {
    // SELECT * FROM companies WHERE ... LIMIT/OFFSET
    rows, err := q.db.QueryContext(ctx, q.buildQuery())
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*Company
    for rows.Next() {
        c := &Company{}
        if err := rows.Scan(
            &c.ID, &c.Name, &c.Plan, &c.Status, &c.CreatedAt, &c.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        results = append(results, c)
    }
    return results, rows.Err()
}

// FilterByStatus is a convenience method
func (q *CompanyQuery) FilterByStatus(status string) *CompanyQuery {
    q.filter["status"] = status
    return q
}
```

---

## 3. Relation Loading

### 3.1 Has-One Relation Method
```go
// Company returns the associated department, loading from DB if needed
func (c *Company) Department(db *sql.DB) (*Department, error) {
    ds := sql.DB{Must: mustStmt(`SELECT id, name, company_id FROM departments WHERE company_id = ? LIMIT 1`, c.ID)}
    ds.QueryRow()
    // Scan into Department
    var d Department
    if err := ds.Scan(&d.ID, &d.Name, &d.CompanyID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil // no department
        }
        return nil, err
    }
    return &d, nil
}
```

### 3.2 Has-Many Relation Method
```go
// Departments returns all departments for this company
func (c *Company) Departments(db *sql.DB) ([]*Department, error) {
    rows, err := db.Query(`SELECT id, name, company_id FROM departments WHERE company_id = ?`, c.ID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*Department
    for rows.Next() {
        d := &Department{}
        if err := rows.Scan(&d.ID, &d.Name, &d.CompanyID); err != nil {
            return nil, err
        }
        results = append(results, d)
    }
    return results, rows.Err()
}
```

### 3.3 Eager Loading in Query Builder
```go
// CompanyQuery with eager loading
func (q *CompanyQuery) WithDepartments() *CompanyQuery {
    q.relations = append(q.relations, "departments")
    return q
}

func (q *CompanyQuery) All(ctx context.Context) ([]*Company, error) {
    companies, err := q.baseQuery(ctx)
    if err != nil {
        return nil, err
    }
    
    // Load relations if specified
    if len(q.relations) > 0 {
        for _, r := range q.relations {
            // Load each relation for each company
            for _, c := range companies {
                // ... load relation
            }
        }
    }
    
    return companies, nil
}
```

---

## 4. Multi-Tenant Integration

### 4.1 Tenant-Aware Model Access
```go
// Get tenant database for a given tenant ID
func getTenantDB(tenantID string) *sql.DB {
    // Return the tenant's SQLite database connection
    // Cached in tenant provider
    return tenantProviders[tenantID].DB()
}

// Company service per tenant
func GetCompanyService(tenantID string) *CompanyService {
    db := getTenantDB(tenantID)
    return &CompanyService{
        db: db,
    }
}
```

### 4.2 Tenant-Scoped Query
```go
// Get companies for a specific tenant only
func (q *CompanyQuery) ForTenant(tenantID string) *CompanyQuery {
    q.filter["tenant_id"] = tenantID
    return q
}
```

---

## 4.3 Seed Data with Generated Models
```go
func initCompanySeed(db *sql.DB) error {
    // Insert using generated struct
    c := &Company{
        ID:        uuid.New(),
        Name:      "Acme Corp",
        Plan:      strPtr("pro"),
        Status:    "active",
    }
    
    if err := db.Create(c); err != nil {
        return err
    }
    return nil
}
```

---

## 4.4 Codegen Pipeline

### 4.1 go:generate Templates
```
go:generate go run ./cmd/codegen generate ./docs/manifests/base/
```

### 4.2 Codegen Outputs
| File | Purpose |
|------|---------|
| `models_gen.go` | Type-safe structs |
| `queries_gen.go` | Query builders |
| `relations_gen.go` | Relation methods |
| `seed_gen.go` | Seed functions |
| `registry_gen.go` | Model registry |

### 4.2 Template Example
```go
// Codegen template for model struct
type {{ .Name }} struct {
{{ range .Fields }}{{ .Name }} {{ .GoType }} `db:"{{ .Name }}" json:"{{ .Name }}{{ end}}`
```

### 4.3 Generated Output
```go
type Company struct {
    ID        uuid.UUID    `db:"id" json:"id"`
    Name      string       `db:"name" json:"name"`
    Plan      *string      `db:"plan" json:"plan,omitempty"`
    Status    string       `db:"status" json:"status"`
    CreatedAt time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}
```

---

## 5. Relation Methods Auto-Generation

### 5.1 Has-One Auto-Generate
```go
// Department returns the company's department
func (c *Company) Department(db *sql.DB) (*Department, error) {
    var d Department
    err := db.QueryRow(`SELECT id, name, company_id FROM departments WHERE company_id = ?`, c.ID).
        Scan(&d.ID, &d.Name, &d.CompanyID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &d, nil
}
```

### 5.2 Has-Many Auto-Generate
```go
// Departments returns all departments for this company
func (c *Company) Departments(db *sql.DB) ([]*Department, error) {
    rows, err := db.Query(`SELECT id, name, company_id FROM departments WHERE company_id = ?`, c.ID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*Department
    for rows.Next() {
        d := Department{}
        if err := rows.Scan(&d.ID, &d.Name, &d.CompanyID); err != nil {
            return nil, err
        }
        results = append(results, &d)
    }
    return results, rows.Err()
}
```

### 5.3 Many-to-Many Auto-Generate
```go
// Tags returns all tags for this project
func (p *Project) Tags(db *sql.DB) ([]*Tag, error) {
    rows, err := db.Query(`SELECT t.id, t.name FROM tags t JOIN project_tag pt ON t.id = pt.tag_id WHERE pt.project_id = ?`, p.ID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*Tag
    for rows.Next() {
        t := Tag{}
        if err := rows.Scan(&t.ID, &t.Name); err != nil {
            return nil, err
        }
        results = append(results, &t)
    }
    return results, rows.Err()
}
```

---

## 6. Integration with Migration Executor

### 6.1 Migration Creates Tables from Manifest
```go
func upCreateCompaniesTable(db *sql.DB, manifest *Model) error {
    // Generate SQL from manifest fields
    cols := []string{"id TEXT PRIMARY KEY"}
    for _, f := range manifest.Fields {
        col := fmt.Sprintf(`%s %s`, f.Name, goTypeToSQL(f.GoType))
        if f.Unique {
            col += " UNIQUE"
        }
        if f.Required {
            col += " NOT NULL"
        }
        cols = append(cols, col)
    }
    
    createTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)`, manifest.Collection, strings.Join(cols, ", "))
    _, err := db.Exec(createTableSQL)
    return err
}
```

### 6.2 Migration Drops Tables
```go
func downDropTable(db *sql.DB, manifest *Model) error {
    _, err := db.Exec(`DROP TABLE IF EXISTS ` + manifest.Collection)
    return err
}
```

---

## 7. Open Issues

1. **UUID generation:** Should IDs be generated by Go (uuid.New()) or PocketBase (auto-generated 21-char string)?
2. **Nullability:** How to handle optional fields in Go structs vs PocketBase (which always has a value)?
3. **JSON tags:** PocketBase stores JSON differently; how to map?
4. **Composite keys:** PocketBase uses string IDs; how to handle composite primary keys?
5. **Relation caching:** Should relation results be cached per request?

---

## 6. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-001 PocketBase Capabilities](aeng-001-pocketbase-capabilities.md)
- [PocketBase Go SDK](https://pkg.go.dev/github.com/pocketbase/pocketbase/core)
- [sqlc](https://sqlc.dev/) - for codegen inspiration