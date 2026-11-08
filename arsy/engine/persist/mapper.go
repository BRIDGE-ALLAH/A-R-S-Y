package persist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ganigeorgiev.com/pocketbase/dbx"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/query"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/query/compiler"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
)

// Store persists validated engine records into the current application DB.
// It is intentionally thin: query validation/compilation live in engine/query,
// and execution uses existing dbx/sql patterns.
type Store struct {
	registry registry.ModelRegistry
	db       dbx.Builder
}

func New(reg registry.ModelRegistry, db dbx.Builder) (*Store, error) {
	if reg == nil {
		return nil, errors.New("persist: model registry is required")
	}
	if db == nil {
		return nil, errors.New("persist: db is required")
	}
	return &Store{registry: reg, db: db}, nil
}

func (s *Store) describe(modelID string) (registry.Descriptor, map[string]struct{}, error) {
	desc, err := s.registry.Lookup(registry.ModelID(modelID))
	if err != nil {
		if errors.Is(err, registry.ErrUnknownModel) {
			return registry.Descriptor{}, nil, &ErrUnknownModel{ModelID: modelID}
		}
		return registry.Descriptor{}, nil, err
	}
	allowed := map[string]struct{}{
		"id": {},
	}
	if desc.TableName == "" || !isSafeIdentifier(desc.TableName) {
		return registry.Descriptor{}, nil, fmt.Errorf("persist: model %s has invalid table mapping", modelID)
	}
	return desc, allowed, nil
}

func (s *Store) Create(ctx context.Context, modelID string, values map[string]any) (string, error) {
	if ctx == nil {
		return "", errors.New("persist: context is required")
	}
	desc, _, err := s.describe(modelID)
	if err != nil {
		return "", err
	}
	if values == nil {
		return "", &ErrUnknownField{ModelID: modelID, Field: "<nil>"}
	}
	cols := make([]string, 0, len(values)+1)
	args := make([]any, 0, len(values)+1)
	for k, v := range values {
		if !isSafeIdentifier(k) || isReservedWriteField(k) {
			return "", &ErrUnknownField{ModelID: modelID, Field: k}
		}
		cols = append(cols, quoteIdent(k))
		args = append(args, v)
	}
	if len(cols) == 0 {
		return "", &ErrUnknownField{ModelID: modelID, Field: "<empty>"}
	}
	sort.SliceStable(cols, func(i, j int) bool { return cols[i] < cols[j] })
	sortValuesByQuotedColumns(cols, args, values)
	ph := make([]string, len(cols))
	for i := range ph {
		ph[i] = fmt.Sprintf("{:v%d}", i)
	}
	params := dbx.Params{}
	for i, v := range args {
		params[fmt.Sprintf("v%d", i)] = v
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", quoteIdent(desc.TableName), strings.Join(cols, ", "), strings.Join(ph, ", "))
	if _, err := s.db.NewQuery(sql).WithContext(ctx).Bind(params).Execute(); err != nil {
		return "", mapDBError(modelID, err)
	}
	id, _ := values["id"].(string)
	if id == "" {
		id = fmt.Sprint(values["id"])
	}
	return id, nil
}

func (s *Store) Get(ctx context.Context, modelID, id string) (map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("persist: context is required")
	}
	desc, _, err := s.describe(modelID)
	if err != nil {
		return nil, err
	}
	row := dbx.NullStringMap{}
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = {:id} LIMIT 1", quoteIdent(desc.TableName), quoteIdent("id"))
	err = s.db.NewQuery(sql).WithContext(ctx).Bind(dbx.Params{"id": id}).One(row)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil, &ErrNotFound{ModelID: modelID, RecordID: id}
	}
	if err != nil {
		return nil, mapDBError(modelID, err)
	}
	return normalizeRow(row), nil
}

func (s *Store) Update(ctx context.Context, modelID, id string, values map[string]any) error {
	if ctx == nil {
		return errors.New("persist: context is required")
	}
	desc, _, err := s.describe(modelID)
	if err != nil {
		return err
	}
	if values == nil || len(values) == 0 {
		return &ErrUnknownField{ModelID: modelID, Field: "<empty>"}
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if !isSafeIdentifier(k) || isReservedWriteField(k) || k == "id" {
			return &ErrUnknownField{ModelID: modelID, Field: k}
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sets := make([]string, 0, len(keys)+1)
	params := dbx.Params{"id": id}
	for i, k := range keys {
		ph := fmt.Sprintf("v%d", i)
		sets = append(sets, fmt.Sprintf("%s = {:%s}", quoteIdent(k), ph))
		params[ph] = values[k]
	}
	if _, ok := values["updated"]; !ok {
		withUpdatedSets := append(append([]string(nil), sets...), fmt.Sprintf("%s = {:updated}", quoteIdent("updated")))
		withUpdatedParams := dbx.Params{"id": id, "updated": time.Now().UTC().Format(time.RFC3339Nano)}
		for k, v := range params {
			if k == "id" {
				continue
			}
			withUpdatedParams[k] = v
		}
		sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s = {:id}", quoteIdent(desc.TableName), strings.Join(withUpdatedSets, ", "), quoteIdent("id"))
		res, err := s.db.NewQuery(sql).WithContext(ctx).Bind(withUpdatedParams).Execute()
		if err == nil {
			affected, _ := res.RowsAffected()
			if affected == 0 {
				return &ErrNotFound{ModelID: modelID, RecordID: id}
			}
			return nil
		}
		if !isMissingUpdatedColumn(err) {
			return mapDBError(modelID, err)
		}
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s = {:id}", quoteIdent(desc.TableName), strings.Join(sets, ", "), quoteIdent("id"))
	res, err := s.db.NewQuery(sql).WithContext(ctx).Bind(params).Execute()
	if err != nil {
		return mapDBError(modelID, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return &ErrNotFound{ModelID: modelID, RecordID: id}
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, modelID, id string) error {
	if ctx == nil {
		return errors.New("persist: context is required")
	}
	desc, _, err := s.describe(modelID)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s = {:id}", quoteIdent(desc.TableName), quoteIdent("id"))
	res, err := s.db.NewQuery(sql).WithContext(ctx).Bind(dbx.Params{"id": id}).Execute()
	if err != nil {
		return mapDBError(modelID, err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return &ErrNotFound{ModelID: modelID, RecordID: id}
	}
	return nil
}

func (s *Store) Query(ctx context.Context, modelID string, q *query.Query, allowedFields []string) ([]map[string]any, error) {
	if ctx == nil {
		return nil, errors.New("persist: context is required")
	}
	desc, _, err := s.describe(modelID)
	if err != nil {
		return nil, err
	}
	if q == nil {
		q = &query.Query{}
	}
	fields := q.Select
	if len(allowedFields) == 0 && len(fields) == 0 {
		fields = []string{"*"}
	}
	if len(allowedFields) == 0 {
		return nil, &ErrUnknownField{ModelID: modelID, Field: "select"}
	}
	schema := allowedSchema(allowedFields)
	qq := *q
	qq.Select = fields
	if err := qq.ValidateWithOptions(schema, query.ValidateOptions{}); err != nil {
		if errors.Is(err, query.ErrUnknownField) {
			return nil, &ErrUnknownField{ModelID: modelID, Field: "select/filter/sort"}
		}
		return nil, err
	}
	plan, err := compiler.Compile(&qq, schema, compiler.Options{})
	if err != nil {
		return nil, err
	}
	sql := strings.Replace(plan.SQL, " FROM target", " FROM "+quoteIdent(desc.TableName), 1)
	params := dbx.Params{}
	for i, v := range plan.Args {
		params[fmt.Sprintf("p%d", i+1)] = normalizeArg(v)
	}
	// compiler uses $n placeholders; dbx uses named params for portability in this codebase.
	for i := len(plan.Args); i >= 1; i-- {
		sql = strings.ReplaceAll(sql, fmt.Sprintf("$%d", i), fmt.Sprintf("{:p%d}", i))
	}
	rows := []dbx.NullStringMap{}
	if err := s.db.NewQuery(sql).WithContext(ctx).Bind(params).All(&rows); err != nil {
		return nil, mapDBError(modelID, err)
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, normalizeRow(row))
	}
	return out, nil
}

func normalizeArg(v any) any {
	switch raw := v.(type) {
	case json.RawMessage:
		if len(raw) == 0 {
			return nil
		}
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err == nil {
			return decoded
		}
		return string(raw)
	case []byte:
		if len(raw) == 0 {
			return nil
		}
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err == nil {
			return decoded
		}
		return string(raw)
	default:
		return v
	}
}

type allowedSchema []string

func (a allowedSchema) HasField(name string) bool {
	if name == "" {
		return false
	}
	if len(a) == 0 {
		return isSafeIdentifier(name)
	}
	for _, f := range a {
		if f == name {
			return true
		}
	}
	return false
}

func isSafeIdentifier(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for i, c := range name {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			continue
		}
		if i > 0 && c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}

func isReservedWriteField(name string) bool {
	switch strings.ToLower(name) {
	case "created", "rowid":
		return true
	default:
		return false
	}
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func normalizeRow(row dbx.NullStringMap) map[string]any {
	out := make(map[string]any, len(row))
	for k, v := range row {
		if v.Valid {
			out[k] = v.String
		} else {
			out[k] = nil
		}
	}
	return out
}

func sortValuesByQuotedColumns(cols []string, args []any, values map[string]any) {
	for i, col := range cols {
		name := strings.Trim(col, `"`)
		args[i] = values[name]
	}
}

func mapDBError(modelID string, err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "unique constraint"), strings.Contains(msg, "constraint failed: unique"), strings.Contains(msg, "duplicate"):
		return &ErrConstraintViolation{ModelID: modelID, Constraint: "unique"}
	case strings.Contains(msg, "no such table"), strings.Contains(msg, "unknown column"), strings.Contains(msg, "has no column"), strings.Contains(msg, "invalid field name"):
		return &ErrUnknownField{ModelID: modelID, Field: "schema"}
	default:
		return err
	}
}

func isMissingUpdatedColumn(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such column") && strings.Contains(msg, "updated")
}
