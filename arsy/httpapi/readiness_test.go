package httpapi

import (
	"errors"
	"net/http"
	"testing"
)

func TestReadinessStatus(t *testing.T) {
	tests := []struct {
		name         string
		bootstrapped bool
		checkDB      func() error
		wantCode     int
		wantStatus   string
	}{
		{
			name:         "ready",
			bootstrapped: true,
			checkDB:      func() error { return nil },
			wantCode:     http.StatusOK,
			wantStatus:   "ready",
		},
		{
			name:         "not bootstrapped",
			bootstrapped: false,
			checkDB:      func() error { return nil },
			wantCode:     http.StatusServiceUnavailable,
			wantStatus:   "not_ready",
		},
		{
			name:         "database unavailable",
			bootstrapped: true,
			checkDB:      func() error { return errors.New("database unavailable") },
			wantCode:     http.StatusServiceUnavailable,
			wantStatus:   "not_ready",
		},
		{
			name:         "missing database check",
			bootstrapped: true,
			wantCode:     http.StatusServiceUnavailable,
			wantStatus:   "not_ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotStatus := readinessStatus(tt.bootstrapped, tt.checkDB)
			if gotCode != tt.wantCode || gotStatus != tt.wantStatus {
				t.Fatalf("readinessStatus() = (%d, %q), want (%d, %q)", gotCode, gotStatus, tt.wantCode, tt.wantStatus)
			}
		})
	}
}
