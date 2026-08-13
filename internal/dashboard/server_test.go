package dashboard

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAssetServer_EmbeddedMode(t *testing.T) {
	// Test that embedded mode can be created (even if assets aren't available)
	server, err := NewAssetServer(ModeEmbedded, "")
	if err != nil {
		// This might fail if no embedded assets are available, which is okay for tests
		t.Skipf("Embedded mode not available: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.mode != ModeEmbedded {
		t.Errorf("Expected mode %s, got %s", ModeEmbedded, server.mode)
	}
}

func TestAssetServer_ExternalMode(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file
	testFile := tempDir + "/test.html"
	os.WriteFile(testFile, []byte("<html>test</html>"), 0644)

	server, err := NewAssetServer(ModeExternal, tempDir)
	if err != nil {
		t.Fatalf("Failed to create external asset server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.mode != ModeExternal {
		t.Errorf("Expected mode %s, got %s", ModeExternal, server.mode)
	}

	// Test serving a file
	req := httptest.NewRequest("GET", "/test.html", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAssetServer_ExternalMode_MissingPath(t *testing.T) {
	_, err := NewAssetServer(ModeExternal, "")
	if err == nil {
		t.Error("Expected error for missing external path")
	}

	_, err = NewAssetServer(ModeExternal, "/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent external path")
	}
}

func TestAssetServer_ServeHTTP_EmbeddedUnavailable(t *testing.T) {
	server := &AssetServer{
		mode:     ModeEmbedded,
		embedded: nil, // Simulate unavailable embedded assets
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestAssetServer_ServeHTTP_ExternalUnavailable(t *testing.T) {
	server := &AssetServer{
		mode:     ModeExternal,
		external: nil, // Simulate unavailable external assets
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}
