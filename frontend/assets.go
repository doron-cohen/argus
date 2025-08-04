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
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
	case ".map":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
	default:
		w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
	}
}

// Handler returns an http.Handler that serves the embedded frontend assets
func Handler() http.Handler {
	// Create a sub-filesystem for dist files
	distFS, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}

	// Create a custom file server that serves index.html from root and dist files from their paths
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting root path, serve index.html
		if r.URL.Path == "/" {
			file, err := assets.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}

			var responseWritten bool
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					// Log the error but don't fail the request if response already written
					slog.Error("Failed to close file", "error", closeErr)
					if !responseWritten {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					}
				}
			}()

			// Type-safe assertion with proper error handling
			readSeeker, ok := file.(io.ReadSeeker)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				responseWritten = true
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Cache HTML for 5 minutes
			w.Header().Set("Cache-Control", "public, max-age=300")
			http.ServeContent(w, r, "index.html", time.Now(), readSeeker)
			responseWritten = true
			return
		}

		// For dist files, strip the /dist/ prefix and serve from dist directory
		if strings.HasPrefix(r.URL.Path, "/dist/") {
			// Add cache headers for static assets
			applyCacheHeaders(w, r.URL.Path)
			http.StripPrefix("/dist/", http.FileServer(http.FS(distFS))).ServeHTTP(w, r)
			return
		}

		// For all other paths, serve from dist directory with cache headers
		applyCacheHeaders(w, r.URL.Path)
		http.StripPrefix("/", http.FileServer(http.FS(distFS))).ServeHTTP(w, r)
	})
}

// HandlerWithPrefix returns an http.Handler that serves the embedded frontend assets
// with an optional path prefix (e.g., "/static/")
func HandlerWithPrefix(prefix string) http.Handler {
	if prefix == "" {
		return Handler()
	}

	// Create a sub-filesystem for dist files
	distFS, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path starts with the prefix
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}

		// Remove the prefix from the path
		path := r.URL.Path[len(prefix):]

		// If requesting root path (after prefix removal), serve index.html
		if path == "/" || path == "" {
			file, err := assets.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}

			var responseWritten bool
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					// Log the error but don't fail the request if response already written
					slog.Error("Failed to close file", "error", closeErr)
					if !responseWritten {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
					}
				}
			}()

			// Type-safe assertion with proper error handling
			readSeeker, ok := file.(io.ReadSeeker)
			if !ok {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				responseWritten = true
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Cache HTML for 5 minutes
			w.Header().Set("Cache-Control", "public, max-age=300")
			http.ServeContent(w, r, "index.html", time.Now(), readSeeker)
			responseWritten = true
			return
		}

		// For all other paths, serve from dist directory with cache headers
		applyCacheHeaders(w, path)
		r.URL.Path = path
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
	})
}
