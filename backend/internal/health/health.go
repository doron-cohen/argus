package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/doron-cohen/argus/backend/internal/storage"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

func HealthHandler(repo *storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check database connectivity
		dbStatus := "healthy"
		if err := repo.DB.WithContext(ctx).Raw("SELECT 1").Error; err != nil {
			dbStatus = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		response := HealthResponse{
			Status:    "ok",
			Database:  dbStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
