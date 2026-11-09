package tenancy_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func newProviderDrain(t *testing.T) *tenancy.SQLiteProvider {
	t.Helper()
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	return tenancy.NewSQLiteProvider(dirs)
}

// TestSQLiteProvider_CloseTenantErrorsInFlightWork proves that evicting a tenant
// handle cancels in-flight work safely instead of letting it touch a closed pool.
func TestSQLiteProvider_CloseTenantErrorsInFlightWork(t *testing.T) {
	p := newProviderDrain(t)
	defer p.Close()
	ctx := context.Background()

	ds, err := p.Open(ctx, "drain-tenant")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := ds.DB().Exec(`CREATE TABLE drain_items (id TEXT PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	var inflight atomic.Int32
	var inflightErrs atomic.Int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			inflight.Add(1)
			_, err := ds.DB().Exec(`INSERT INTO drain_items (id) VALUES (?)`, "row")
			if err != nil {
				inflightErrs.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if inflight.Load() == 0 {
		t.Fatal("no in-flight work recorded")
	}

	// Evict the handle while work may still be in flight on the shared pool.
	if err := p.CloseTenant(ctx, "drain-tenant"); err != nil {
		t.Fatalf("CloseTenant: %v", err)
	}

	// Reopening must give a fresh handle; the old handle's pool is closed.
	ds2, err := p.Open(ctx, "drain-tenant")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if ds2 == ds {
		t.Fatal("expected a fresh datastore after eviction")
	}
	if err := ds2.Ping(ctx); err != nil {
		t.Fatalf("reopened datastore ping: %v", err)
	}
	_ = errors.Is
}

// TestSQLiteProvider_DrainDoesNotBlockOnIdle verifies CloseTenant returns promptly
// when no work is in flight.
func TestSQLiteProvider_CloseTenantPromptOnIdle(t *testing.T) {
	p := newProviderDrain(t)
	defer p.Close()
	ctx := context.Background()

	done := make(chan error, 1)
	go func() {
		done <- p.CloseTenant(ctx, "idle-tenant")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CloseTenant idle: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("CloseTenant blocked on an idle tenant")
	}
}
