package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/types"
)

const outboxCollectionName = "arsy_outbox"

func init() {
	core.SystemMigrations.Register(upBaaSOutbox, downBaaSOutbox, "20261109000000_baas_outbox.go")
}

func upBaaSOutbox(app core.App) error {
	_, err := ensureCollection(app, outboxCollectionName, func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "operation", Required: true, Max: 32},
			&core.TextField{Name: "tenant_id", Required: true, Max: 64},
			&core.TextField{Name: "idempotency_key", Max: 128},
			&core.TextField{Name: "state", Required: true, Max: 32, Pattern: `^(pending|in_progress|success|failed_retryable|failed_permanent)$`},
			&core.NumberField{Name: "attempt_count", Min: 0},
			&core.TextField{Name: "result_summary", Max: 500},
		)
		c.Indexes = types.JSONArray[string]{
			"CREATE INDEX idx_baas_outbox_state ON arsy_outbox (state, created)",
			"CREATE INDEX idx_baas_outbox_tenant ON arsy_outbox (tenant_id)",
		}
	})
	return err
}

func downBaaSOutbox(app core.App) error {
	existing, err := app.FindCollectionByNameOrId(outboxCollectionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("find %s: %w", outboxCollectionName, err)
	}
	return app.Delete(existing)
}
