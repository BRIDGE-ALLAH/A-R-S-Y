package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	_ "modernc.org/sqlite"
)

// DefaultRegistry registers starter engine models shipped with the runtime.
// Models are table-backed; schema application is handled by ApplySchema.
func DefaultRegistry() (*registry.InMemoryRegistry, error) {
	reg := registry.New()
	models := []registry.Descriptor{
		{
			ID:          "base.company",
			Label:       "Company",
			TableName:   "base_company",
			Version:     "1.0.0",
			OwnerModule: "base",
			Source:      registry.SourceStaticGo,
		},
	}
	for _, model := range models {
		if err := reg.Register(model); err != nil && !errors.Is(err, registry.ErrDuplicateModel) {
			return nil, err
		}
	}
	return reg, nil
}

// ApplySchema creates idempotent table structures for registered models in the
// tenant database identified by tenantID under dirs.TenantsRoot.
func ApplySchema(ctx context.Context, dirs config.BaaSDirectories, tenantID string, reg registry.ModelRegistry) error {
	if ctx == nil {
		return errors.New("bootstrap: context is required")
	}
	if reg == nil {
		return errors.New("bootstrap: registry is required")
	}
	if tenantID == "" || filepath.Base(tenantID) != tenantID {
		return tenancy.ErrInvalidTenantID
	}
	dbPath, err := dirs.TenantDBPath(tenantID)
	if err != nil {
		return err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	for _, desc := range reg.Snapshot() {
		if err := applyModelTable(ctx, db, desc); err != nil {
			return err
		}
	}
	return nil
}

var identPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func applyModelTable(ctx context.Context, db *sql.DB, desc registry.Descriptor) error {
	table := desc.TableName
	if !identPattern.MatchString(table) {
		return fmt.Errorf("bootstrap: invalid table name %q", table)
	}
	create := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q (id TEXT PRIMARY KEY NOT NULL)`, table)
	if _, err := db.ExecContext(ctx, create); err != nil {
		return err
	}
	// Starter schema columns for shipped models; dynamic field-driven schema is
	// intentionally deferred. Existing data is preserved.
	switch string(desc.ID) {
	case "base.company":
		for _, stmt := range []string{
			fmt.Sprintf(`ALTER TABLE %q ADD COLUMN name TEXT`, table),
			fmt.Sprintf(`ALTER TABLE %q ADD COLUMN code TEXT`, table),
			fmt.Sprintf(`ALTER TABLE %q ADD COLUMN active TEXT`, table),
		} {
			if _, err := db.ExecContext(ctx, stmt); err != nil && !isDuplicateColumn(err) {
				return err
			}
		}
	}
	return nil
}

func isDuplicateColumn(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate column")
}
