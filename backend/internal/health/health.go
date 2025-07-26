package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Checker defines an interface for health checks
type Checker interface {
	HealthCheck(ctx context.Context) error
	Name() string
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp"`
}

// HealthHandler creates a health check handler that accepts multiple checkers
func HealthHandler(checkers ...Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)
		overallStatus := "healthy"

		// Run all health checks
		for _, checker := range checkers {
			checkName := checker.Name()
			if err := checker.HealthCheck(ctx); err != nil {
				checks[checkName] = "unhealthy"
				overallStatus = "unhealthy"
			} else {
				checks[checkName] = "healthy"
			}
		}

		// Set appropriate status code
		if overallStatus == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		response := HealthResponse{
			Status:    overallStatus,
			Checks:    checks,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
