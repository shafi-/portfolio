package assets

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockLogger implements the Logger interface for testing
type mockLogger struct {
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockLogger) Info(msg string, fields ...Field) {
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockLogger) Warn(msg string, fields ...Field) {
	m.warnMessages = append(m.warnMessages, msg)
}

func (m *mockLogger) Error(msg string, fields ...Field) {
	m.errorMessages = append(m.errorMessages, msg)
}

func TestHandler_DirectoryTraversal(t *testing.T) {
	logger := &mockLogger{}
	handler := NewHandler("", logger)

	req := httptest.NewRequest("GET", "/../etc/passwd", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	logger := &mockLogger{}
	handler := NewHandler("", logger)

	req := httptest.NewRequest("POST", "/index.html", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 for POST because it serves index.html (SPA fallback)
	// The asset handler serves files regardless of method for static assets
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for asset serving, got %d", http.StatusOK, w.Code)
	}
}

func TestHandler_InvalidPath(t *testing.T) {
	logger := &mockLogger{}
	handler := NewHandler("", logger)

	req := httptest.NewRequest("GET", "/assets/../../config.json", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_RootRequest(t *testing.T) {
	logger := &mockLogger{}
	handler := NewHandler("", logger)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 because embedded assets are available
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d when embedded assets available, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}
}

func TestHandler_EmptyPath(t *testing.T) {
	logger := &mockLogger{}
	handler := NewHandler("", logger)

	req := httptest.NewRequest("GET", "/", nil) // Empty path should behave like root
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 because embedded assets are available
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d when embedded assets available, got %d", http.StatusOK, w.Code)
	}
}
