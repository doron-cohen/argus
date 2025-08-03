package frontend

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

//go:embed index.html dist/*
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
	return assets
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
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					// Log the error but don't fail the request if response already written
					// Note: We can't call http.Error here as the response may already be written
					slog.Error("Failed to close file", "error", closeErr)
				}
			}()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", time.Now(), file.(io.ReadSeeker))
			return
		}

		// For dist files, strip the /dist/ prefix and serve from dist directory
		if strings.HasPrefix(r.URL.Path, "/dist/") {
			http.StripPrefix("/dist/", http.FileServer(http.FS(distFS))).ServeHTTP(w, r)
			return
		}

		// For all other paths, serve from dist directory
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
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					// Log the error but don't fail the request if response already written
					// Note: We can't call http.Error here as the response may already be written
					slog.Error("Failed to close file", "error", closeErr)
				}
			}()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", time.Now(), file.(io.ReadSeeker))
			return
		}

		// For all other paths, serve from dist directory
		r.URL.Path = path
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
	})
}
