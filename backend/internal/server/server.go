package server

import (
	"log/slog"
	"net/http"
)

func StartServer() {
	slog.Info("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Server failed", "error", err)
	}
}
