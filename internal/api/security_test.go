package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"project-dash/internal/logging"
)

// TestHealthUnhealthyDB verifies the health endpoint returns 503 with an
// "unhealthy" status when the database is unreachable (Issue #5 regression).
func TestHealthUnhealthyDB(t *testing.T) {
	db := tempDB(t)
	db.Close() // break the connection

	logger, _ := logging.NewLogger("ERROR", "console")
	s := NewServer(db, logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for broken DB, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "unhealthy" {
		t.Errorf("expected status unhealthy, got %v", resp["status"])
	}
	if resp["database_connected"] != false {
		t.Errorf("expected database_connected=false, got %v", resp["database_connected"])
	}
}

// TestSQLInjectionPayloads verifies SQL injection payloads via the search
// endpoint never cause server errors or data leakage (Issue #12).
func TestSQLInjectionPayloads(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "victim-project", "/tmp/victim")

	payloads := []string{
		"'; DROP TABLE projects--",
		"' OR '1'='1",
		"admin'--",
		"' UNION SELECT id, name FROM projects--",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/search?q="+url.QueryEscape(payload), nil)
			s.Handler().ServeHTTP(w, r)

			if w.Code >= 500 {
				t.Errorf("payload caused server error: status %d, body: %s", w.Code, w.Body.String())
			}
			// Payload must be treated as a literal search term, not executed
			if w.Code == http.StatusOK {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				for _, r := range resp["results"].([]interface{}) {
					name := r.(map[string]interface{})["name"]
					if name == "victim-project" && strings.Contains(strings.ToUpper(payload), "UNION") {
						t.Errorf("UNION payload leaked project data: %v", name)
					}
				}
			}
		})
	}

	// Verify the projects table still exists and works after all payloads
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects", nil)
	s.Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("projects endpoint broken after injection payloads: status %d", w.Code)
	}
}

// TestXSSPayloads verifies XSS payloads are safely JSON-encoded in responses.
func TestXSSPayloads(t *testing.T) {
	s := newTestServer(t)

	payloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
	}

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/search?q="+url.QueryEscape(payload), nil)
			s.Handler().ServeHTTP(w, r)

			body := w.Body.String()
			if strings.Contains(body, "<script>") || strings.Contains(body, "onerror=") {
				t.Errorf("XSS payload not escaped in response body: %s", body)
			}
		})
	}
}

// TestSearchInputValidation verifies query length limits (Issue #10).
func TestSearchInputValidation(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q="+strings.Repeat("a", 10000), nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for over-length query, got %d", w.Code)
	}
}

// TestStructuredErrorResponses verifies the error response format from
// Issue #21: {error, code, type} with proper categorization.
func TestStructuredErrorResponses(t *testing.T) {
	s := newTestServer(t)

	tests := []struct {
		name     string
		path     string
		status   int
		wantCode string
		wantType string
	}{
		{"not found", "/projects/does-not-exist", http.StatusNotFound, "NOT_FOUND", "resource_error"},
		{"invalid input", "/search", http.StatusBadRequest, "INVALID_INPUT", "validation_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.path, nil)
			s.Handler().ServeHTTP(w, r)

			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode ErrorResponse: %v", err)
			}
			if resp.Error == "" {
				t.Error("expected non-empty error message")
			}
			if resp.Code != tt.wantCode {
				t.Errorf("expected code %q, got %q", tt.wantCode, resp.Code)
			}
			if resp.Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, resp.Type)
			}
		})
	}
}

// TestTechnologiesEndpoint verifies the /technologies endpoint (Issue #22).
func TestTechnologiesEndpoint(t *testing.T) {
	s := newTestServer(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/technologies", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var technologies []interface{}
	if err := json.NewDecoder(w.Body).Decode(&technologies); err != nil {
		t.Fatalf("decode response as array: %v (body: %s)", err, w.Body.String())
	}
	if technologies == nil {
		t.Error("expected non-nil technologies array (empty list should serialize as [])")
	}
}
