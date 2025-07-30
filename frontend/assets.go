package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *.html
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
	return assets
}

// Handler returns an http.Handler that serves the embedded frontend assets
func Handler() http.Handler {
	return http.FileServer(http.FS(assets))
}

// HandlerWithPrefix returns an http.Handler that serves the embedded frontend assets
// with an optional path prefix (e.g., "/static/")
func HandlerWithPrefix(prefix string) http.Handler {
	if prefix == "" {
		return Handler()
	}
	return http.StripPrefix(prefix, Handler())
}
