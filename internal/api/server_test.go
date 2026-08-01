package api

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
	"project-dash/internal/logging"
)

func tempDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open temp db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func initSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		root_path TEXT NOT NULL UNIQUE,
		repository_type TEXT NOT NULL,
		discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS metadata (
		project_id TEXT PRIMARY KEY,
		git_head TEXT, default_branch TEXT, last_commit_at TIMESTAMP,
		last_modified_at TIMESTAMP, commit_count INTEGER DEFAULT 0,
		language_summary TEXT, framework_summary TEXT,
		dependency_summary TEXT, documentation_hash TEXT, last_scan_at TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY, project_id TEXT NOT NULL, path TEXT NOT NULL,
		kind TEXT NOT NULL, content TEXT, content_hash TEXT,
		indexed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS analyses (
		id TEXT PRIMARY KEY, project_id TEXT NOT NULL, analyzer TEXT NOT NULL,
		analyzed_git_head TEXT NOT NULL, analyzed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		summary TEXT, purpose TEXT, architecture TEXT, notes TEXT, raw_json TEXT
	);
	CREATE TABLE IF NOT EXISTS features (
		id TEXT PRIMARY KEY, analysis_id TEXT NOT NULL, name TEXT NOT NULL,
		description TEXT, confidence REAL
	);
	CREATE TABLE IF NOT EXISTS technologies (
		id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, category TEXT
	);
	CREATE TABLE IF NOT EXISTS project_technologies (
		project_id TEXT NOT NULL, technology_id TEXT NOT NULL,
		PRIMARY KEY (project_id, technology_id)
	);
	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY, source_project TEXT NOT NULL, target_project TEXT NOT NULL,
		type TEXT NOT NULL, description TEXT, confidence REAL
	);
	CREATE TABLE IF NOT EXISTS configuration (
		key TEXT PRIMARY KEY, value TEXT, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("init schema: %v", err)
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	db := tempDB(t)
	initSchema(t, db)
	logger, _ := logging.NewLogger("ERROR", "console")
	return NewServer(db, logger)
}

func seedProject(t *testing.T, s *Server, id, name, rootPath string) {
	t.Helper()
	_, err := s.db.Exec(
		"INSERT INTO projects (id, name, root_path, repository_type) VALUES (?, ?, ?, ?)",
		id, name, rootPath, "git",
	)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func TestHealth(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "healthy" {
		t.Errorf("expected healthy, got %v", resp["status"])
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	s := newTestServer(t)
	s.db.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.Handler().ServeHTTP(w, r)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["database_connected"] != false {
		t.Error("expected database_connected to be false")
	}
}

func TestListProjects_Empty(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	projects := resp["projects"].([]interface{})
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestListProjects(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "alpha", "/tmp/alpha")
	seedProject(t, s, "p2", "beta", "/tmp/beta")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects", nil)
	s.Handler().ServeHTTP(w, r)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	projects := resp["projects"].([]interface{})
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestListProjects_Search(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "alpha-project", "/tmp/alpha")
	seedProject(t, s, "p2", "beta-app", "/tmp/beta")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects?q=alpha", nil)
	s.Handler().ServeHTTP(w, r)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	projects := resp["projects"].([]interface{})
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
}

func TestGetProject_Found(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "test-project", "/tmp/test")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects/p1", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["name"] != "test-project" {
		t.Errorf("expected test-project, got %v", resp["name"])
	}
}

func TestGetProject_NotFound(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/projects/nonexistent", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSearch(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "my-project", "/tmp/my")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=project", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	results := resp["results"].([]interface{})
	found := false
	for _, r := range results {
		item := r.(map[string]interface{})
		if item["type"] == "project" {
			found = true
		}
	}
	if !found {
		t.Error("expected project in search results")
	}
}

func TestSearch_NoQuery(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetConfig(t *testing.T) {
	s := newTestServer(t)
	s.db.Exec("INSERT INTO configuration (key, value, updated_at) VALUES (?, ?, datetime('now'))", "test_key", "test_value")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/configuration", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["test_key"] != "test_value" {
		t.Errorf("expected test_value, got %v", resp["test_key"])
	}
}

func TestPatchConfig(t *testing.T) {
	s := newTestServer(t)
	body := `{"custom_key": "custom_value"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PATCH", "/configuration", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["custom_key"] != "custom_value" {
		t.Errorf("expected custom_value, got %v", resp["custom_key"])
	}
}

func TestPatchConfig_InvalidBody(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PATCH", "/configuration", strings.NewReader("not-json"))
	r.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(w, r)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatistics(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "proj-a", "/tmp/a")
	seedProject(t, s, "p2", "proj-b", "/tmp/b")
	s.db.Exec("INSERT INTO metadata (project_id) VALUES ('p1')")
	s.db.Exec("INSERT INTO metadata (project_id) VALUES ('p2')")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/statistics", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["total_projects"].(float64) != 2 {
		t.Errorf("expected 2 projects, got %v", resp["total_projects"])
	}
}

func TestRelationships_ProjectNotFound(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/relationships/nonexistent", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRelationships_Empty(t *testing.T) {
	s := newTestServer(t)
	seedProject(t, s, "p1", "test", "/tmp/test")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/relationships/p1", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	rels := resp["relationships"].([]interface{})
	if len(rels) != 0 {
		t.Errorf("expected empty relationships, got %d", len(rels))
	}
}

func TestCORS(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/health", nil)
	s.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200 for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header")
	}
}
