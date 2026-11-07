package httpapi

import "net/http"

// readinessStatus reports only whether local startup and the database are ready.
// Database errors are deliberately reduced to a generic status for API callers.
func readinessStatus(bootstrapped bool, checkDatabase func() error) (int, string) {
	if !bootstrapped || checkDatabase == nil {
		return http.StatusServiceUnavailable, "not_ready"
	}
	if err := checkDatabase(); err != nil {
		return http.StatusServiceUnavailable, "not_ready"
	}
	return http.StatusOK, "ready"
}
