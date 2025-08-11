package frontend

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

//go:embed index.html dist/*
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
	return assets
}

// applyCacheHeaders sets appropriate cache headers based on file extension
func applyCacheHeaders(w http.ResponseWriter, path string) {
	ext := filepath.Ext(path)
	switch ext {
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

func serveIndex(w http.ResponseWriter, r *http.Request) {
	file, err := assets.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var responseWritten bool
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.Error("Failed to close file", "error", closeErr)
			if !responseWritten {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}
	}()
	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		responseWritten = true
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeContent(w, r, "index.html", time.Now(), readSeeker)
	responseWritten = true
}

func serveDistFile(w http.ResponseWriter, r *http.Request, path string) {
	// Create a sub-filesystem for dist files
	distFS, err := fs.Sub(assets, "dist")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	filePath := strings.TrimPrefix(path, "/dist/")
	if filePath == "" || strings.HasSuffix(path, "/") {
		http.NotFound(w, r)
		return
	}
	file, err := distFS.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.Error("Failed to close file", "error", closeErr)
		}
	}()
	readSeeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	applyCacheHeaders(w, path)
	http.ServeContent(w, r, filepath.Base(filePath), time.Now(), readSeeker)
}

// Handler serves /dist/* as static assets and serves index.html for any other path
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/dist/") {
			serveDistFile(w, r, r.URL.Path)
			return
		}
		// Everything else is a client route (SPA)
		serveIndex(w, r)
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
		if stripped == "" || stripped == "/" || stripped == "index.html" || stripped == "/index.html" {
			serveIndex(w, r)
			return
		}
		if strings.HasPrefix(stripped, "dist/") || strings.HasPrefix(stripped, "/dist/") {
			// Reconstruct a path starting with /dist/
			path := stripped
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			serveDistFile(w, r, path)
			return
		}
		// Any other path under the prefix is treated as client route
		serveIndex(w, r)
	})
}
