# AENG-009: Define Model Relation Metadata Contracts

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-001, AENG-006

## Purpose
Define the metadata contract for model relations that AENG modules use to describe has-one, has-many, and many-to-many relationships between collections. This enables type-safe relation loading and validation.

---

## 1. Relation Types

### 1.1 Has-One (1:1)
```
Parent Collection: Companies
Child Collection: CompanySettings

Fields:
  - Companies: id, name, ...
  - CompanySettings: id, company_id, config_key, config_value
```

| Aspect | Contract |
|--------|----------|
| Foreign Key | `company_id` on CompanySettings table |
| Go Type | `CompanySettings.Company() *Company` |
| Loading | Eager (join) or Lazy (separate query) |
| Cascade Delete | Optional (set in manifest) |

### 1.2 Has-Many (1:N)
```
Parent Collection: Users
Child Collection: UserProfiles

Fields:
  - Users: id, name, email, ...
  - UserProfiles: id, user_id, bio, avatar
```

| Aspect | Contract |
|--------|----------|
| Foreign Key | `user_id` on UserProfiles table |
| Go Type | `User.Profiles() []UserProfile` |
| Loading | Lazy (default) or Eager with `LoadOptions` |
| Max Limit | Configurable via `LoadOptions` |

### 1.3 Many-to-Many (N:M)
```
Parent Collection: Projects
Child Collection: Tags

Junction Table: project_tag
  - project_id, tag_id (composite primary key)
```

| Aspect | Contract |
|--------|----------|
| Junction Table | Required: `project_tag(project_id, tag_id)` |
| Go Type | `Project.Tags() []Tag` and `Tag.Projects() []Project` |
| Loading | Lazy (separate queries for both sides) |
| Relation Fields | Additional columns allowed (e.g., `role`) |

---

## 2. Metadata Contract Definition

### 2.1 Relation Field in Model Manifest
```go
type Relation struct {
    Name        string           // e.g. "employees", "tags"
    Type        RelationType     // HasOne, HasMany, ManyToMany
    Model       string           // Target model name
    ForeignKey  string           // Field name on child, e.g. "company_id"
    OwnerKey    string           // Field name on parent, e.g. "id"
    CascadeDelete bool            // Set to true to cascade delete
    SelectOnly  bool             // If true, only select fields, don't load relations
}

// RelationType defines the kind of relation
type RelationType string

const (
    RelationHasOne    RelationType = "has_one"
    RelationHasMany   RelationType = "has_many"
    RelationManyToMany RelationType = "many_to_many"
)
```

### 2.2 Generated Accessors

```go
// HasOne: Company -> Department
func (c *Company) Department() (*Department, error) { ... }

// HasMany -> Users -> Posts
func (u *Users) Posts() ([]*Post, error) { ... }

// ManyToMany -> Projects <-> Tags
func (p *Project) Tags() ([]*Tag, error) { ... }
func (t *Tag) Projects() ([]*Project, error) { ... }
```

### 2.2 Relation Loading Options
```go
type LoadOptions struct {
    // Eager load specified relations
    Relations []string
    // Limit per relation
    Limits map[string]int
    // Filter per relation
    Filters map[string]Filter
}
```

---

## 3. Validation Rules

### 3.1 Relation Contract Validation
| Rule | Description |
|------|-------------|
| FK Existence | Foreign key field must exist on child model |
| Owner Key Existence | Owner key field must exist on parent model |
| Type Consistency | FK type must match Owner key type |
| Cascade Safety | Cascade delete must not violate other relations |

### 3.2 Cross-Model Validation
- No circular relations allowed
- Relation names must be unique within a model
- Foreign keys must be indexed (or created automatically by migration)

---

## 3. Example: Full Manifest Snippet

```go
type ModuleManifest struct {
    Models []Model `json:"models"`
}

type Model struct {
    Name string `json:"name"`
    Fields []Field `json:"fields"`
    Relations []Relation `json:"relations"`
}

type Field struct {
    Name      string   `json:"name"`
    Type    string   `json:"type"`     // text, number, bool, relation, etc.
    Required bool     `json:"required"`
    Unique   bool     `json:"unique"`
}

type Relation struct {
    Name      string   `json:"name"`       // e.g. "department", "posts", "tags"
    Type      string   `json:"type"`       // "has_one", "has_many", "many_to_many"
    Target    string   `json:"target"`     // target model name
    ForeignKey string  `json:"foreign_key"` // e.g. "company_id"
    OwnerKey  string   `json:"owner_key"`  // e.g. "id"
    Cascade   bool     `json:"cascade"`    // true to cascade delete
}
```

---

## 4. Migration Integration

### 4.1 Generated Migration SQL

```go
// HasOne: Add company_id to departments
func up_Migration_Base() error {
    _, err := db.Exec(`ALTER TABLE departments ADD COLUMN company_id INTEGER`)
    if err != nil { return err }
    _, err = db.Exec(`CREATE INDEX idx_departments_company_id ON departments(company_id)`)
    return err
}

func down_Migration_Base() error {
    _, err = db.Exec(`DROP INDEX idx_departments_company_id`)
    if err != nil { return err }
    _, err = db.Exec(`ALTER TABLE departments DROP COLUMN company_id`)
    return err
}
```

### 4.2 Many-to-Many Junction Table
```go
func up_Migration_Base() error {
    _, err := db.Exec(`CREATE TABLE project_tag (
        project_id INTEGER NOT NULL,
        tag_id INTEGER NOT NULL,
        PRIMARY KEY (project_id, tag_id)
    )`)
    if err != nil { return err }
    _, err = db.Exec(`CREATE INDEX idx_project_tag_project_id ON project_tag(project_id)`)
    if err != nil { return err }
    _, err = db.Exec(`CREATE INDEX idx_project_tag_tag_id ON project_tag(tag_id)`)
    return err
}

func down_Migration_Base() error {
    _, err = db.Exec(`DROP INDEX idx_project_tag_tag_id`)
    if err != nil { return err }
    _, err = db.Exec(`DROP INDEX idx_project_tag_project_id`)
    if err != nil { return err }
    _, err = db.Exec(`DROP TABLE project_tag`)
    return err
}
```

---

## 5. Relation Loading API

### 5.1 Eager Load
```go
// Load company with department
company, err := model.Load(ctx, "company", "1", LoadOptions{
    Relations: []string{"department"},
})

// Load user with profiles and posts
user, err := model.Load(ctx, "user", "1", LoadOptions{
    Relations: []string{"profiles", "posts"},
    Limits: map[string]int{
        "profiles": 1,
        "posts": 10,
    },
})
```

### 5.2 Lazy Load (Default)
```go
// User already loaded, then:
posts, err := user.LoadRelated(ctx, "posts", LoadOptions{
    Filters: map[string]Filter{
        "company_id": {"Op": "=", "Val: companyID},
    },
})
```

---

## 6. Security Integration

### 6.1 Relation-Level Authorization
```go
// Only allow viewing relations if user has access
func (u *User) CanViewProfiles(ctx context.Context) bool {
    return auth.HasRole(ctx, "admin") || auth.IsOwner(ctx, u.ID)
}
```

### 6.2 Relation Security Metadata
```go
type RelationSecurity struct {
    Name      string
    AllowRead bool   // default: true
    AllowWrite bool  // default: false
    Roles     []string // roles allowed, empty = superuser only
}
```

RelationSecurity would be part of the model manifest and consumed by the API layer.

---

## 7. Open Issues

1. **Relation naming conflict:** What if two models have relation with same name?
2. **Deep relations:** Nested `LoadOptions.Relations` with multiple levels?
3. **Relation caching:** Cache loaded relations per request?
4. **Bulk relation loading:** Load relations for multiple records at once?

---

## 8. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-001 PocketBase Capabilities](aeng-001-pocketbase-capabilities.md)
- [PocketBase Relations](https://pocketbase.io/docs/relations/)