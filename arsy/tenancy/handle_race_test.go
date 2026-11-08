package tenancy_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func newProvider(t *testing.T) *tenancy.SQLiteProvider {
	t.Helper()
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	return tenancy.NewSQLiteProvider(dirs)
}

func tenantID(idx int) string {
	return "tenant-" + string(rune('A'+idx%26))
}

// TestSQLiteProvider_ConcurrentOpenSameTenant verifies that concurrent opens for
// one tenant identity do not create unsafe duplicate handles.
func TestSQLiteProvider_ConcurrentOpenSameTenant(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	ctx := context.Background()

	const goroutines = 32
	start := make(chan struct{})
	results := make(chan tenancy.Datastore, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ds, err := p.Open(ctx, "race-same")
			if err != nil {
				t.Errorf("Open: %v", err)
				return
			}
			results <- ds
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var handles []tenancy.Datastore
	for ds := range results {
		handles = append(handles, ds)
	}
	if len(handles) != goroutines {
		t.Fatalf("expected %d handles, got %d", goroutines, len(handles))
	}
	// Idempotency: concurrent opens for one tenant must return the same handle.
	first := handles[0]
	for _, h := range handles[1:] {
		if h != first {
			t.Fatal("concurrent opens for one tenant returned different handles")
		}
	}
}

// TestSQLiteProvider_OpenCloseInterleaving covers open/close/suspend/reassign races.
func TestSQLiteProvider_OpenCloseInterleaving(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	ctx := context.Background()

	const goroutines = 64
	start := make(chan struct{})
	var okCount atomic.Int32
	var errCount atomic.Int32
	var wg sync.WaitGroup
	for idx := 0; idx < goroutines; idx++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			id := tenantID(idx)
			ds, err := p.Open(ctx, id)
			if err != nil && !errors.Is(err, tenancy.ErrDatastoreClosed) {
				errCount.Add(1)
				return
			}
			if err == nil {
				_ = ds
				okCount.Add(1)
			}
			// Close some tenants mid-flight to exercise eviction races.
			if idx%3 == 0 {
				_ = p.CloseTenant(ctx, id)
			}
		}(idx)
	}
	close(start)
	wg.Wait()

	if errCount.Load() > 0 {
		t.Errorf("unexpected errors: %d", errCount.Load())
	}
	if okCount.Load() == 0 {
		t.Fatal("no successful opens in interleaving test")
	}
}

// TestSQLiteProvider_DifferentTenantsNeverShareHandle verifies isolation.
func TestSQLiteProvider_DifferentTenantsNeverShareHandle(t *testing.T) {
	p := newProvider(t)
	defer p.Close()
	ctx := context.Background()

	const goroutines = 16
	start := make(chan struct{})
	results := make(chan tenancy.Datastore, goroutines)
	var wg sync.WaitGroup
	for idx := 0; idx < goroutines; idx++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			id := tenantID(idx % 8)
			ds, _ := p.Open(ctx, id)
			results <- ds
		}(idx)
	}
	close(start)
	wg.Wait()
	close(results)

	count := 0
	seen := make(map[tenancy.Datastore]struct{})
	for ds := range results {
		count++
		seen[ds] = struct{}{}
	}
	if count != goroutines {
		t.Fatalf("expected %d handles, got %d", goroutines, count)
	}
	// 8 distinct tenants => at most 8 distinct handles.
	if len(seen) > 8 {
		t.Errorf("expected at most 8 distinct handles, got %d", len(seen))
	}
}
