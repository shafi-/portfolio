package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// mockLogger implements the Logger interface for testing
type mockSearchLogger struct {
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockSearchLogger) Info(msg string, fields ...Field) {
	m.infoMessages = append(m.infoMessages, msg)
}

func (m *mockSearchLogger) Error(msg string, fields ...Field) {
	m.errorMessages = append(m.errorMessages, msg)
}

func (m *mockSearchLogger) Warn(msg string, fields ...Field) {
	m.warnMessages = append(m.warnMessages, msg)
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			root_path TEXT NOT NULL,
			repository_type TEXT NOT NULL,
			discovered_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE documents (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			path TEXT NOT NULL,
			kind TEXT NOT NULL,
			content TEXT,
			discovered_at TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);

		CREATE TABLE project_technologies (
			project_id TEXT NOT NULL,
			technology TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);

		CREATE TABLE project_frameworks (
			project_id TEXT NOT NULL,
			framework TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		);

		CREATE INDEX idx_projects_name ON projects(name);
		CREATE INDEX idx_documents_content ON documents(content);
		CREATE INDEX idx_project_technologies ON project_technologies(technology);
		CREATE INDEX idx_project_frameworks ON project_frameworks(framework);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func seedTestData(db *sql.DB) error {
	now := time.Now().Format(time.RFC3339)

	// Insert projects
	projects := []struct {
		ID   string
		Name string
		Path string
		Type string
	}{
		{"1", "React Dashboard", "/path1", "git"},
		{"2", "Vue Admin Panel", "/path2", "git"},
		{"3", "Go API Server", "/path3", "git"},
	}

	for _, p := range projects {
		_, err := db.Exec(`
			INSERT INTO projects (id, name, root_path, repository_type, discovered_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, p.ID, p.Name, p.Path, p.Type, now, now)
		if err != nil {
			return err
		}
	}

	// Insert documents
	documents := []struct {
		ID        string
		ProjectID string
		Path      string
		Kind      string
		Content   string
	}{
		{"1", "1", "README.md", "README", "This React dashboard uses modern React for UI"},
		{"2", "2", "README.md", "README", "Vue admin panel built with Vue.js"},
		{"3", "3", "README.md", "README", "Go API server with REST endpoints"},
	}

	for _, d := range documents {
		_, err := db.Exec(`
			INSERT INTO documents (id, project_id, path, kind, content, discovered_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, d.ID, d.ProjectID, d.Path, d.Kind, d.Content, now)
		if err != nil {
			return err
		}
	}

	// Insert technologies
	_, err := db.Exec(`INSERT INTO project_technologies (project_id, technology) VALUES (?, ?)`, "1", "React")
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO project_technologies (project_id, technology) VALUES (?, ?)`, "2", "Vue")
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO project_technologies (project_id, technology) VALUES (?, ?)`, "3", "Go")
	if err != nil {
		return err
	}

	// Insert frameworks
	_, err = db.Exec(`INSERT INTO project_frameworks (project_id, framework) VALUES (?, ?)`, "1", "Next.js")
	if err != nil {
		return err
	}

	return nil
}

func TestSearchHandler_BasicSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := seedTestData(db)
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=react", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestSearchHandler_MissingQueryParameter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_EmptyQueryParameter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_NoMatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := seedTestData(db)
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=zzzznonexistentzzzz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSearchHandler_InvalidDateFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=test&from=01-15-2024", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_PaginationDefaults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := seedTestData(db)
	if err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=project", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check that pagination defaults are applied
	// (This would require parsing the JSON response to fully validate)
}

func TestSearchHandler_InvalidPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=test&page=-1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_NonNumericPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("GET", "/search?q=test&page=abc", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSearchHandler_MethodNotAllowed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := &mockSearchLogger{}
	handler := NewSearchHandler(db, logger)

	req := httptest.NewRequest("POST", "/search?q=test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	allowHeader := w.Header().Get("Allow")
	if allowHeader != "GET, OPTIONS" {
		t.Errorf("Expected Allow header 'GET, OPTIONS', got %s", allowHeader)
	}
}
