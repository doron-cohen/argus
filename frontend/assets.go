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

//go:embed index.html dist/*
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
    return assets
}

// RouteConfig controls how static assets and client routes are matched
type RouteConfig struct {
    StaticPrefixes []string
    APIPrefixes    []string
    ClientRoutes   []string
}

type staticHandler struct {
    config RouteConfig
    distFS fs.FS
}

func mustCreateDistFS() fs.FS {
    distFS, err := fs.Sub(assets, "dist")
    if err != nil {
        panic(err)
    }
    return distFS
}

// NewStaticHandler builds a handler using the provided routing configuration
func NewStaticHandler(config RouteConfig) http.Handler {
    return &staticHandler{
        config: config,
        distFS: mustCreateDistFS(),
    }
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

func (h *staticHandler) isStaticAsset(path string) bool {
    if filepath.Ext(path) != "" {
        return true
    }
    for _, prefix := range h.config.StaticPrefixes {
        if strings.HasPrefix(path, prefix) {
            return true
        }
    }
    return false
}

func (h *staticHandler) isClientRoute(path string) bool {
    for _, prefix := range h.config.APIPrefixes {
        if strings.HasPrefix(path, prefix) {
            return false
        }
    }
    for _, pattern := range h.config.ClientRoutes {
        if h.matchesPattern(pattern, path) {
            return true
        }
    }
    return false
}

func (h *staticHandler) matchesPattern(pattern, path string) bool {
    if pattern == "" {
        return false
    }
    // Exact match
    if pattern == path {
        return true
    }
    // Trailing wildcard support: "/components/*" matches "/components" and any deeper path
    if strings.HasSuffix(pattern, "/*") {
        base := strings.TrimSuffix(pattern, "/*")
        if path == base || strings.HasPrefix(path, base+"/") {
            return true
        }
    }
    return false
}

func (h *staticHandler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) error {
    file, err := h.distFS.Open(filePath)
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := file.Close(); closeErr != nil {
            slog.Error("Failed to close file", "error", closeErr)
        }
    }()

    readSeeker, ok := file.(io.ReadSeeker)
    if !ok {
        return errors.New("file does not implement io.ReadSeeker")
    }

    applyCacheHeaders(w, filePath)
    http.ServeContent(w, r, filepath.Base(filePath), time.Now(), readSeeker)
    return nil
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    requestedPath := r.URL.Path

    // Serve index.html for configured client routes
    if requestedPath == "/" || h.isClientRoute(requestedPath) {
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
        return
    }

    // Serve static asset files from /dist/
    if strings.HasPrefix(requestedPath, "/dist/") {
        trimmed := strings.TrimPrefix(requestedPath, "/dist/")
        if trimmed == "" || strings.HasSuffix(requestedPath, "/") {
            http.NotFound(w, r)
            return
        }
        if err := h.serveFile(w, r, trimmed); err != nil {
            http.NotFound(w, r)
        }
        return
    }

    // If it's a file-like path (has extension) but not under /dist/, do not serve it
    if h.isStaticAsset(requestedPath) {
        http.NotFound(w, r)
        return
    }

    // Fallback: serve index.html for any other non-API route
    if h.isClientRoute(requestedPath) {
        file, err := assets.Open("index.html")
        if err != nil {
            http.NotFound(w, r)
            return
        }
        readSeeker, ok := file.(io.ReadSeeker)
        if !ok {
            _ = file.Close()
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Cache-Control", "public, max-age=300")
        http.ServeContent(w, r, "index.html", time.Now(), readSeeker)
        _ = file.Close()
        return
    }

    http.NotFound(w, r)
}

// Handler returns an http.Handler that serves the embedded frontend assets with default configuration
func Handler() http.Handler {
    // Default configuration mirrors existing behavior and adds client routes support
    config := RouteConfig{
        StaticPrefixes: []string{"/dist/"},
        APIPrefixes:    []string{"/api/"},
        ClientRoutes: []string{
            "/",
            "/components",
            "/components/*",
            "/settings",
            "/sync",
        },
    }
    return NewStaticHandler(config)
}

// HandlerWithPrefix returns an http.Handler that serves the embedded frontend assets
// with an optional path prefix (e.g., "/static/")
func HandlerWithPrefix(prefix string) http.Handler {
    if prefix == "" {
        return Handler()
    }

    baseHandler := Handler()
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.HasPrefix(r.URL.Path, prefix) {
            http.NotFound(w, r)
            return
        }
        // Strip prefix and delegate
        original := r.URL.Path
        stripped := strings.TrimPrefix(r.URL.Path, prefix)

        // Special-case: allow accessing index.html directly under the prefix
        if stripped == "index.html" || stripped == "/index.html" {
            file, err := assets.Open("index.html")
            if err != nil {
                http.NotFound(w, r)
                return
            }
            readSeeker, ok := file.(io.ReadSeeker)
            if !ok {
                _ = file.Close()
                http.Error(w, "Internal server error", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.Header().Set("Cache-Control", "public, max-age=300")
            http.ServeContent(w, r, "index.html", time.Now(), readSeeker)
            _ = file.Close()
            return
        }

        r.URL.Path = stripped
        if r.URL.Path == "" || !strings.HasPrefix(original, prefix) {
            r.URL.Path = "/"
        }
        baseHandler.ServeHTTP(w, r)
    })
}
