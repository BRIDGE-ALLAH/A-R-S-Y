// Package aeng provides the Application Engine framework.
// AENG-030: Static Go Model Persistence Mapper.
package aeng

import (
	"fmt"
	"sort"
	"strings"
)

// GoModelMapper generates type-safe Go model code from a module manifest.
type GoModelMapper struct {
	registry *ModelRegistry
}

// NewGoModelMapper creates a mapper backed by a registry.
func NewGoModelMapper(registry *ModelRegistry) *GoModelMapper {
	return &GoModelMapper{registry: registry}
}

// GenerateModel emits a Go source file for all models in a module manifest.
func (m *GoModelMapper) GenerateModel(moduleID string, manifest *ModuleManifest) (string, error) {
	if manifest == nil {
		return "", fmt.Errorf("aeng: module %s: nil manifest", moduleID)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("package %s\n\n", moduleID))
	b.WriteString(`import (
	"database/sql"
	"time"
)
`)

	models := make([]ModelDef, 0, len(manifest.Models))
	models = append(models, manifest.Models...)
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	for _, model := range models {
		b.WriteString(m.generateStruct(model))
		b.WriteString("\n")
		b.WriteString(m.generateRelationMethods(model))
		b.WriteString("\n")
		b.WriteString(m.generateQueryBuilder(model))
		b.WriteString("\n")
		b.WriteString(m.generateService(model))
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (m *GoModelMapper) generateStruct(model ModelDef) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("// %s represents a %s in the system\n", model.Name, model.Name))
	b.WriteString(fmt.Sprintf("type %s struct {\n", model.Name))
	for _, f := range model.Fields {
		b.WriteString(fmt.Sprintf("    %s %s `db:\"%s\" json:\"%s,omitempty\"`\n",
			goFieldName(f.Name), goTypeFromPB(f.Type), f.Name, strings.ToLower(f.Name)))
	}
	b.WriteString("}\n")
	return b.String()
}

func goFieldName(name string) string {
	if name == "" {
		return "ID"
	}
	runes := []rune(name)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] -= 32
	}
	return string(runes)
}

func (m *GoModelMapper) generateRelationMethods(model ModelDef) string {
	var b strings.Builder
	for _, rel := range model.Relations {
		targetName := goFieldName(rel.Target)
		b.WriteString(fmt.Sprintf("// %s returns the related %s for this %s\n",
			rel.Name, rel.Target, model.Name))
		switch RelationType(rel.Type) {
		case RelationHasOne:
			b.WriteString(fmt.Sprintf("func (c *%s) %s(db *sql.DB) (*%s, error) {\n",
				model.Name, rel.Name, targetName))
			b.WriteString(fmt.Sprintf("    var d %s\n", targetName))
			b.WriteString(fmt.Sprintf("    err := db.QueryRow(\"SELECT * FROM %s WHERE %s = ?\", c.ID).Scan(&d.ID, &d.Name)\n",
				strings.ToLower(rel.Target), rel.ForeignKey))
			b.WriteString("    if err != nil {\n")
			b.WriteString("        if errors.Is(err, sql.ErrNoRows) {\n")
			b.WriteString("            return nil, nil\n")
			b.WriteString("        }\n")
			b.WriteString("        return nil, err\n")
			b.WriteString("    }\n")
			b.WriteString("    return &d, nil\n")
			b.WriteString("}\n")
		case RelationHasMany:
			b.WriteString(fmt.Sprintf("func (c *%s) %s(db *sql.DB) ([]*%s, error) {\n",
				model.Name, rel.Name, targetName))
			b.WriteString(fmt.Sprintf("    rows, err := db.Query(\"SELECT * FROM %s WHERE %s = ?\", c.ID)\n",
				strings.ToLower(rel.Target), rel.ForeignKey))
			b.WriteString("    if err != nil {\n")
			b.WriteString("        return nil, err\n")
			b.WriteString("    }\n")
			b.WriteString("    defer rows.Close()\n")
			b.WriteString(fmt.Sprintf("    var results []*%s\n", targetName))
			b.WriteString("    for rows.Next() {\n")
			b.WriteString(fmt.Sprintf("        d := &%s{}\n", targetName))
			b.WriteString("        if err := rows.Scan(&d.ID, &d.Name); err != nil {\n")
			b.WriteString("            return nil, err\n")
			b.WriteString("        }\n")
			b.WriteString(fmt.Sprintf("        results = append(results, d)\n    }\n"))
			b.WriteString("    return results, rows.Err()\n")
			b.WriteString("}\n")
		case RelationManyToMany:
			b.WriteString(fmt.Sprintf("func (c *%s) %s(db *sql.DB) ([]*%s, error) {\n",
				model.Name, rel.Name, targetName))
			jt := strings.ToLower(model.Name) + "_" + strings.ToLower(rel.Target)
			b.WriteString(fmt.Sprintf("    rows, err := db.Query(\"SELECT t.* FROM %s t JOIN %s jt ON t.id = jt.%s_id WHERE jt.%s_id = ?\", c.ID)\n",
				strings.ToLower(rel.Target), jt, strings.ToLower(rel.Target), strings.ToLower(model.Name)))
			b.WriteString("    if err != nil {\n")
			b.WriteString("        return nil, err\n")
			b.WriteString("    }\n")
			b.WriteString("    defer rows.Close()\n")
			b.WriteString(fmt.Sprintf("    var results []*%s\n", targetName))
			b.WriteString("    for rows.Next() {\n")
			b.WriteString(fmt.Sprintf("        d := &%s{}\n", targetName))
			b.WriteString("        if err := rows.Scan(&d.ID, &d.Name); err != nil {\n")
			b.WriteString("            return nil, err\n")
			b.WriteString("        }\n")
			b.WriteString(fmt.Sprintf("        results = append(results, d)\n    }\n"))
			b.WriteString("    return results, rows.Err()\n")
			b.WriteString("}\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *GoModelMapper) generateQueryBuilder(model ModelDef) string {
	var b strings.Builder
	name := model.Name
	b.WriteString(fmt.Sprintf("// %sQuery is a type-safe query for %s\n", name, name))
	b.WriteString(fmt.Sprintf("type %sQuery struct {\n    db     *sql.DB\n    filter map[string]any\n    limit  int\n    offset int\n}\n\n", name))
	b.WriteString(fmt.Sprintf("func New%sQuery(db *sql.DB) *%sQuery {\n    return &%sQuery{db: db, filter: make(map[string]any)}\n}\n\n", name, name, name))
	b.WriteString(fmt.Sprintf("func (q *%sQuery) Filter(field string, value any) *%sQuery {\n    q.filter[field] = value\n    return q\n}\n\n", name, name))
	b.WriteString(fmt.Sprintf("func (q *%sQuery) Limit(n int) *%sQuery {\n    q.limit = n\n    return q\n}\n\n", name, name))
	b.WriteString(fmt.Sprintf("func (q *%sQuery) Offset(n int) *%sQuery {\n    q.offset = n\n    return q\n}\n\n", name, name))
	b.WriteString(fmt.Sprintf("func (q *%sQuery) All(ctx context.Context) ([]*%s, error) {\n    query, args := q.buildQuery()\n    rows, err := q.db.QueryContext(ctx, query, args...)\n    if err != nil {\n        return nil, err\n    }\n    defer rows.Close()\n    var results []*%s\n    for rows.Next() {\n        m := &%s{}\n        if err := rows.Scan(&m.ID, &m.Name); err != nil {\n            return nil, err\n        }\n        results = append(results, m)\n    }\n    return results, rows.Err()\n}\n\n", name, name, name, name))
	b.WriteString(fmt.Sprintf("func (q *%sQuery) buildQuery() (string, []any) {\n    query := \"SELECT * FROM %%s\"\n    var args []any\n    if len(q.filter) > 0 {\n        var conds []string\n        for k, v := range q.filter {\n            conds = append(conds, k+\" = ?\")\n            args = append(args, v)\n        }\n        query += \" WHERE \" + strings.Join(conds, \" AND \")\n    }\n    if q.limit > 0 {\n        query += \" LIMIT \" + strconv.Itoa(q.limit)\n    }\n    if q.offset > 0 {\n        query += \" OFFSET \" + strconv.Itoa(q.offset)\n    }\n    return fmt.Sprintf(query, args...), args\n}\n", name))
	return b.String()
}

func (m *GoModelMapper) generateService(model ModelDef) string {
	var b strings.Builder
	name := model.Name
	table := strings.ToLower(model.Name)
	b.WriteString(fmt.Sprintf("// %sService provides type-safe operations on %s model\n", name, name))
	b.WriteString(fmt.Sprintf("type %sService struct {\n    db     *sql.DB\n    registry *ModelRegistry\n}\n\n", name))
	b.WriteString(fmt.Sprintf("func New%sService(db *sql.DB, registry *ModelRegistry) *%sService {\n    return &%sService{db: db, registry: registry}\n}\n\n", name, name, name))
	cols := strings.Join(columns(model), ", ")
	ph := strings.Join(placeholders(len(model.Fields)), ", ")
	b.WriteString(fmt.Sprintf("func (s *%sService) Create(ctx context.Context, m *%s) error {\n    query := \"INSERT INTO %s (%s) VALUES (%s)\"\n    _, err := s.db.ExecContext(ctx, query, values(m)...)\n    return err\n}\n\n", name, name, table, cols, ph))
	b.WriteString(fmt.Sprintf("func (s *%sService) GetByID(ctx context.Context, id string) (*%s, error) {\n    var m %s\n    err := s.db.QueryRowContext(ctx, \"SELECT * FROM %s WHERE id = ?\", id).Scan(&m.ID, &m.Name)\n    if err != nil {\n        return nil, err\n    }\n    return &m, nil\n}\n", name, name, name, table))
	return b.String()
}

func columns(model ModelDef) []string {
	cols := make([]string, 0, len(model.Fields))
	for _, f := range model.Fields {
		cols = append(cols, f.Name)
	}
	return cols
}

func placeholders(n int) []string {
	ph := make([]string, n)
	for i := range ph {
		ph[i] = "?"
	}
	return ph
}

func values(instance any) []any {
	return []any{instance}
}
