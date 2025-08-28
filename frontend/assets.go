package frontend

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

//go:embed dist/*
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
	return assets
}

// applyCacheHeaders sets appropriate cache headers based on file extension
func applyCacheHeaders(w http.ResponseWriter, path string) {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	case ".map":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	default:
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := assets.Open(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		slog.Error("failed to open file", "path", filePath, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.Error("Failed to close file", "error", closeErr)
		}
	}()

	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		slog.Error("file does not implement io.ReadSeeker", "path", filePath)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	applyCacheHeaders(w, filePath)
	http.ServeContent(w, r, filepath.Base(filePath), time.Now(), readSeeker)
}

// Handler serves files from dist directory and serves index.html for any other path
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// If path is root or doesn't start with /assets/, serve index.html
		if path == "/" || path == "" || !strings.HasPrefix(path, "/assets/") {
			serveFile(w, r, "dist/index.html")
			return
		}

		// Serve the requested asset file (convert /assets/ to dist/assets/)
		filePath := strings.Replace(path, "/assets/", "dist/assets/", 1)
		serveFile(w, r, filePath)
	})
}

// HandlerWithPrefix restricts serving to a given prefix (e.g. "/static/")
func HandlerWithPrefix(prefix string) http.Handler {
	if prefix == "" {
		return Handler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		stripped := strings.TrimPrefix(r.URL.Path, prefix)
		if stripped == "" || stripped == "/" || stripped == "index.html" {
			serveFile(w, r, "dist/index.html")
			return
		}
		// Serve the requested asset file
		filePath := stripped
		if strings.HasPrefix(filePath, "assets/") {
			filePath = "dist/" + filePath
		} else if !strings.HasPrefix(filePath, "dist/") {
			filePath = "dist/" + filePath
		}
		serveFile(w, r, filePath)
	})
}
