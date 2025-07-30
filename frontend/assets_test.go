package frontend

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssets(t *testing.T) {
	fs := Assets()
	if fs == nil {
		t.Fatal("Assets() returned nil")
	}

	// Test that we can read the index.html file from dist
	file, err := fs.Open("index.html")
	if err != nil {
		t.Fatalf("Failed to open index.html: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Failed to close file: %v", closeErr)
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	if len(content) == 0 {
		t.Error("index.html is empty")
	}

	// Check that it contains expected content
	contentStr := string(content)
	if !contains(contentStr, "Component Catalog") {
		t.Error("index.html does not contain expected 'Component Catalog' content")
	}
	if !contains(contentStr, "data-testid") {
		t.Error("index.html does not contain expected 'data-testid' attributes")
	}
}

func TestHandler(t *testing.T) {
	handler := Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test serving index.html from root path
	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to GET /: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(content) == 0 {
		t.Error("Response body is empty")
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestHandlerWithPrefix(t *testing.T) {
	// Test with empty prefix (should behave like Handler())
	handler := HandlerWithPrefix("")
	if handler == nil {
		t.Fatal("HandlerWithPrefix(\"\") returned nil")
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to GET /: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test with a prefix
	prefixedHandler := HandlerWithPrefix("/static/")
	if prefixedHandler == nil {
		t.Fatal("HandlerWithPrefix(\"/static/\") returned nil")
	}

	prefixedServer := httptest.NewServer(prefixedHandler)
	defer prefixedServer.Close()

	// Should be able to access with the prefix
	resp, err = http.Get(prefixedServer.URL + "/static/")
	if err != nil {
		t.Fatalf("Failed to GET /static/: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Should also be able to access index.html with the prefix
	resp, err = http.Get(prefixedServer.URL + "/static/index.html")
	if err != nil {
		t.Fatalf("Failed to GET /static/index.html: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for /static/index.html, got %d", resp.StatusCode)
	}

	// Should not be able to access without the prefix
	resp, err = http.Get(prefixedServer.URL + "/")
	if err != nil {
		t.Fatalf("Failed to GET /: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandlerRootServing(t *testing.T) {
	handler := Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test serving root path (should serve index.html)
	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to GET /: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(content) == 0 {
		t.Error("Response body is empty")
	}

	contentStr := string(content)
	if !contains(contentStr, "Component Catalog") {
		t.Error("Root path does not serve index.html content")
	}
}

func TestHandlerNotFound(t *testing.T) {
	handler := Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test non-existent file
	resp, err := http.Get(server.URL + "/non-existent.html")
	if err != nil {
		t.Fatalf("Failed to GET /non-existent.html: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestDistFilesServed(t *testing.T) {
	handler := Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test that dist/app.js is served
	resp, err := http.Get(server.URL + "/app.js")
	if err != nil {
		t.Fatalf("Failed to GET /app.js: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for app.js, got %d", resp.StatusCode)
	}

	// Test that dist CSS files are served
	resp, err = http.Get(server.URL + "/app-16c896aca7e8d804.css")
	if err != nil {
		t.Fatalf("Failed to GET CSS file: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for CSS file, got %d", resp.StatusCode)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
