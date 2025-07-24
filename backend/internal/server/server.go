package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/doron-cohen/argus/backend/api"
	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/health"
	"github.com/doron-cohen/argus/backend/internal/storage"
	"github.com/doron-cohen/argus/backend/sync"
	"github.com/go-chi/chi/v5"
)

func Start(cfg config.Config) (stop func(), err error) {
	mux := chi.NewRouter()

	// Mount healthz
	mux.Get("/healthz", health.HealthHandler)

	// Connect to PostgreSQL using storage.ConnectAndMigrate
	dsn := cfg.Storage.DSN()
	repo, dberr := storage.ConnectAndMigrate(context.Background(), dsn)
	if dberr != nil {
		slog.Error("Failed to connect or migrate database", "error", dberr)
		return nil, dberr
	}

	// Mount OpenAPI-generated handlers under /api
	mux.Mount("/api", api.Handler(api.NewAPIServer(repo)))

	// Initialize sync service (always create, but may not start if no sources configured)
	// Cast to sync.Repository interface since storage.Repository implements it
	syncService := sync.NewService(repo, cfg.Sync)
	syncCtx, syncCancel := context.WithCancel(context.Background())

	// Start sync service (will log warning and return if no sources configured)
	go syncService.StartPeriodicSync(syncCtx)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		slog.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	stop = func() {
		syncCancel() // Stop sync goroutines
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed", "error", err)
		}
	}

	return stop, nil
}
