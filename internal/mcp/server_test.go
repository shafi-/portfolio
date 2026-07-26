package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/database"
	"project-dash/internal/discovery"
	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

func setupTestDB(t *testing.T) *database.Database {
	t.Helper()
	logger, _ := logging.NewLogger("INFO", "console")
	db, err := database.NewDatabase(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	if err := db.Connect(); err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	if err := db.Initialize(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}
	return db
}

func createTestProject(t *testing.T, db *database.Database) *models.Project {
	t.Helper()
	projectID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	project := &models.Project{
		ID:             projectID,
		Name:           "test-project",
		RootPath:       "/tmp/test-project",
		RepositoryType: "git",
		DiscoveredAt:   now,
		UpdatedAt:      now,
	}

	_, err := db.DB().Exec(
		"INSERT INTO projects (id, name, root_path, repository_type, discovered_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		project.ID, project.Name, project.RootPath, project.RepositoryType, project.DiscoveredAt, project.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	return project
}

func createTestMetadata(t *testing.T, db *database.Database, projectID string) *models.Metadata {
	t.Helper()
	metadata := &models.Metadata{
		ProjectID:         projectID,
		GitHead:           "abc123",
		DefaultBranch:     "main",
		LastCommitAt:      time.Now().UTC().Format(time.RFC3339),
		LastModifiedAt:    time.Now().UTC().Format(time.RFC3339),
		CommitCount:       100,
		LanguageSummary:   "Go:90%, TypeScript:10%",
		FrameworkSummary:  "Cobra, Gin",
		DependencySummary: "github.com/spf13/cobra",
		DocumentationHash: "hash123",
		LastScanAt:        time.Now().UTC().Format(time.RFC3339),
	}

	_, err := db.DB().Exec(
		`INSERT INTO metadata (project_id, git_head, default_branch, last_commit_at, last_modified_at,
		 commit_count, language_summary, framework_summary, dependency_summary, documentation_hash, last_scan_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metadata.ProjectID, metadata.GitHead, metadata.DefaultBranch, metadata.LastCommitAt,
		metadata.LastModifiedAt, metadata.CommitCount, metadata.LanguageSummary,
		metadata.FrameworkSummary, metadata.DependencySummary, metadata.DocumentationHash,
		metadata.LastScanAt,
	)
	if err != nil {
		t.Fatalf("failed to create test metadata: %v", err)
	}

	return metadata
}

func TestHandleHealth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{},
	}

	result, err := server.handleHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("handleHealth failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestHandleDiscoverProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{"/tmp/nonexistent"},
	}

	server := New(cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{},
	}

	result, err := server.handleDiscoverProjects(context.Background(), req)
	if err != nil {
		t.Fatalf("handleDiscoverProjects failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.IsError {
		t.Fatal("discovery should not error on empty roots")
	}
}

func TestHandleListProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	createTestProject(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{},
	}

	result, err := server.handleListProjects(context.Background(), req)
	if err != nil {
		t.Fatalf("handleListProjects failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.IsError {
		t.Fatal("listProjects should not error")
	}
}

func TestHandleGetProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	t.Run("valid project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id": project.ID,
				},
			},
		}

		result, err := server.handleGetProject(context.Background(), req)
		if err != nil {
			t.Fatalf("handleGetProject failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IsError {
			t.Fatal("getProject should not error for valid project")
		}
	})

	t.Run("missing id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result, err := server.handleGetProject(context.Background(), req)
		if err != nil {
			t.Fatalf("handleGetProject failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for missing id")
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"id": "nonexistent",
				},
			},
		}

		result, err := server.handleGetProject(context.Background(), req)
		if err != nil {
			t.Fatalf("handleGetProject failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for nonexistent project")
		}
	})
}

func TestHandleSearchProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	createTestProject(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	t.Run("valid query", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"query": "test",
				},
			},
		}

		result, err := server.handleSearchProjects(context.Background(), req)
		if err != nil {
			t.Fatalf("handleSearchProjects failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IsError {
			t.Fatal("searchProjects should not error for valid query")
		}
	})

	t.Run("missing query", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result, err := server.handleSearchProjects(context.Background(), req)
		if err != nil {
			t.Fatalf("handleSearchProjects failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for missing query")
		}
	})
}

func TestHandleStoreAnalysis(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	t.Run("valid analysis", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
					"analyzer":   "test-analyzer",
					"summary":    "Test summary",
					"purpose":    "Test purpose",
				},
			},
		}

		result, err := server.handleStoreAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleStoreAnalysis failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IsError {
			t.Fatal("storeAnalysis should not error for valid data")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result, err := server.handleStoreAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleStoreAnalysis failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for missing required fields")
		}
	})
}

func TestHandleGetConfiguration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.DB().Exec("INSERT INTO configuration (key, value, updated_at) VALUES (?, ?, datetime('now'))", "test_key", "test_value")
	if err != nil {
		t.Fatalf("failed to insert test configuration: %v", err)
	}

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{},
	}

	result, err := server.handleGetConfiguration(context.Background(), req)
	if err != nil {
		t.Fatalf("handleGetConfiguration failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.IsError {
		t.Fatal("getConfiguration should not error")
	}
}

func TestHandleUpdateConfiguration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	t.Run("valid update", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"key":   "test_key",
					"value": "test_value",
				},
			},
		}

		result, err := server.handleUpdateConfiguration(context.Background(), req)
		if err != nil {
			t.Fatalf("handleUpdateConfiguration failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IsError {
			t.Fatal("updateConfiguration should not error for valid data")
		}
	})

	t.Run("missing key", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"value": "test_value",
				},
			},
		}

		result, err := server.handleUpdateConfiguration(context.Background(), req)
		if err != nil {
			t.Fatalf("handleUpdateConfiguration failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for missing key")
		}
	})
}

func TestHandleListRelationships(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}

	server := New(cfg)

	t.Run("valid project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
				},
			},
		}

		result, err := server.handleListRelationships(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListRelationships failed: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.IsError {
			t.Fatal("listRelationships should not error for valid project")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{},
			},
		}

		result, err := server.handleListRelationships(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListRelationships failed: %v", err)
		}

		if !result.IsError {
			t.Fatal("expected error for missing project_id")
		}
	})
}

func TestDiscoveryStoreAdapter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	adapter := &discoveryStoreAdapter{store: store.NewProjectStore(db.DB(), logger.Zap())}

	now := time.Now()
	discProject := &discovery.Project{
		ID:             "test-id",
		Name:           "test-project",
		RootPath:       "/tmp/test",
		RepositoryType: "git",
		DiscoveredAt:   now,
	}

	err := adapter.UpsertProject(discProject)
	if err != nil {
		t.Fatalf("UpsertProject failed: %v", err)
	}

	project, err := adapter.store.GetProject("test-id")
	if err != nil {
		t.Fatalf("failed to retrieve project: %v", err)
	}

	if project == nil {
		t.Fatal("expected project to be stored")
	}

	if project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got '%s'", project.Name)
	}

	if project.RootPath != "/tmp/test" {
		t.Errorf("expected root path '/tmp/test', got '%s'", project.RootPath)
	}
}

func TestRootsConfigProvider(t *testing.T) {
	roots := []string{"/path1", "/path2"}
	provider := &rootsConfigProvider{roots: roots}

	got, err := provider.GetProjectRoots()
	if err != nil {
		t.Fatalf("GetProjectRoots failed: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(got))
	}

	if got[0] != "/path1" {
		t.Errorf("expected first root '/path1', got '%s'", got[0])
	}

	ignored := provider.GetIgnoredPaths()
	if len(ignored) == 0 {
		t.Fatal("expected non-zero ignored paths")
	}
}

// ---------------------------------------------------------------------------
// Feature tool tests
// ---------------------------------------------------------------------------

func createTestAnalysis(t *testing.T, db *database.Database, projectID string) string {
	t.Helper()
	id := uuid.New().String()
	// Must supply all columns that ListAnalyses scans, even with empty defaults,
	// so SQL NULL never hits a string Scan (which would error).
	_, err := db.DB().Exec(
		`INSERT INTO analyses (id, project_id, analyzer, analyzed_git_head, analyzed_at,
		 summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes, raw_json)
		 VALUES (?, ?, ?, ?, datetime('now'), ?, ?, '', '', '', '', '', '', '')`,
		id, projectID, "test-analyzer", "abc123", "Test summary", "Test purpose",
	)
	if err != nil {
		t.Fatalf("failed to create test analysis: %v", err)
	}
	return id
}

func TestHandleStoreFeature(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)
	createTestAnalysis(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("valid feature", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":  project.ID,
					"analyzer":    "test-analyzer",
					"name":        "user-auth",
					"description": "User authentication with JWT",
					"confidence":  0.95,
				},
			},
		}
		result, err := server.handleStoreFeature(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleStoreFeature(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})

	t.Run("tier3 upsert enriches existing", func(t *testing.T) {
		// First store Tier-2 feature
		req1 := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":  project.ID,
					"analyzer":    "test-analyzer",
					"name":        "tier3-test",
					"description": "Tier 2 description",
				},
			},
		}
		_, err := server.handleStoreFeature(context.Background(), req1)
		if err != nil {
			t.Fatal(err)
		}

		// Now enrich with Tier-3 fields (same name)
		req2 := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":            project.ID,
					"analyzer":              "test-analyzer",
					"name":                  "tier3-test",
					"implementation_status": "complete",
					"feature_architecture":  "JWT middleware",
					"pattern":               "Middleware-based auth",
				},
			},
		}
		result, err := server.handleStoreFeature(context.Background(), req2)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
		// Verify via listFeatures
		listReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		listResult, err := server.handleListFeatures(context.Background(), listReq)
		if err != nil || listResult == nil || listResult.IsError {
			t.Fatal("listFeatures failed")
		}
	})
}

func TestHandleListFeatures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)
	createTestAnalysis(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// Store a feature first
	storeReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"project_id": project.ID,
				"analyzer":   "test-analyzer",
				"name":       "test-feature",
			},
		},
	}
	server.handleStoreFeature(context.Background(), storeReq)

	t.Run("lists features for project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleListFeatures(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleListFeatures(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestHandleSearchFeatures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)
	createTestAnalysis(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// Store features with different patterns
	store := func(name, status, pattern, arch string) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":            project.ID,
					"analyzer":              "test-analyzer",
					"name":                  name,
					"implementation_status": status,
					"pattern":               pattern,
					"feature_architecture":  arch,
				},
			},
		}
		server.handleStoreFeature(context.Background(), req)
	}
	store("auth", "complete", "JWT", "middleware")
	store("search", "partial", "Full-text", "indexer")
	store("dashboard", "planned", "MVC", "controller")

	t.Run("filter by implementation_status", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":            project.ID,
					"implementation_status": "complete",
				},
			},
		}
		result, err := server.handleSearchFeatures(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("filter by pattern", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"pattern": "MVC",
				},
			},
		}
		result, err := server.handleSearchFeatures(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("text search", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"query": "auth",
				},
			},
		}
		result, err := server.handleSearchFeatures(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})
}

// ---------------------------------------------------------------------------
// Technology tool tests
// ---------------------------------------------------------------------------

func TestHandleStoreTechnology(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("create new technology", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"name":     "Flutter",
					"category": "framework",
				},
			},
		}
		result, err := server.handleStoreTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleStoreTechnology(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})

	t.Run("duplicate returns existing", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"name": "Flutter",
				},
			},
		}
		result, err := server.handleStoreTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})
}

func TestHandleTagProjectWithTechnology(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("tag with existing technology", func(t *testing.T) {
		// Create tech first
		techReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"name": "Go", "category": "language"},
			},
		}
		server.handleStoreTechnology(context.Background(), techReq)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":      project.ID,
					"technology_name": "Go",
				},
			},
		}
		result, err := server.handleTagProjectWithTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("tag with new technology auto-creates", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":      project.ID,
					"technology_name": "React",
					"category":        "framework",
				},
			},
		}
		result, err := server.handleTagProjectWithTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleTagProjectWithTechnology(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestHandleListTechnologies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// Store a couple
	server.handleStoreTechnology(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"name": "Go", "category": "language"},
		},
	})
	server.handleStoreTechnology(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"name": "PostgreSQL", "category": "database"},
		},
	})

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{}}
	result, err := server.handleListTechnologies(context.Background(), req)
	if err != nil || result == nil || result.IsError {
		t.Fatal("expected success")
	}
}

func TestHandleListProjectTechnologies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// Tag project with a tech
	server.handleTagProjectWithTechnology(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"project_id":      project.ID,
				"technology_name": "Go",
				"category":        "language",
			},
		},
	})

	t.Run("list for project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleListProjectTechnologies(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleListProjectTechnologies(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestHandleSearchByTechnology(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := createTestProject(t, db)
	createTestMetadata(t, db, project.ID)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// Tag with Go
	server.handleTagProjectWithTechnology(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"project_id":      project.ID,
				"technology_name": "Go",
				"category":        "language",
			},
		},
	})

	t.Run("find projects by technology", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"technology_name": "Go"},
			},
		}
		result, err := server.handleSearchByTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("technology not found", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"technology_name": "Nonexistent"},
			},
		}
		result, err := server.handleSearchByTechnology(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success (empty results)")
		}
	})

	t.Run("missing technology_name", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleSearchByTechnology(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

// ---------------------------------------------------------------------------
// Code content tool tests
// ---------------------------------------------------------------------------

func tempProjectWithFiles(t *testing.T, db *database.Database) (*models.Project, string) {
	t.Helper()
	root := t.TempDir()

	// Create some files
	write := func(name, content string) {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("README.md", "# Test Project\n")
	write("main.go", "package main\nfunc main() {}\n")
	write("go.mod", "module test\n")
	write("internal/app.go", "package app\n")
	write(".env.sample", "DB_HOST=localhost\nDB_PORT=5432\n")
	write(".gitignore", "*.log\n")

	projectID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	project := &models.Project{
		ID:             projectID,
		Name:           "code-test-project",
		RootPath:       root,
		RepositoryType: "git",
		DiscoveredAt:   now,
		UpdatedAt:      now,
	}
	_, err := db.DB().Exec(
		"INSERT INTO projects (id, name, root_path, repository_type, discovered_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		project.ID, project.Name, project.RootPath, project.RepositoryType, project.DiscoveredAt, project.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	return project, root
}

func TestHandleListProjectFiles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project, _ := tempProjectWithFiles(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("list root files", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleListProjectFiles(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleListProjectFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})

	t.Run("nonexistent project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": "nonexistent"},
			},
		}
		result, err := server.handleListProjectFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestHandleGetFileContent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project, _ := tempProjectWithFiles(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("read existing file", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
					"path":       "README.md",
				},
			},
		}
		result, err := server.handleGetFileContent(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run(".env.sample allowed (not sensitive)", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
					"path":       ".env.sample",
				},
			},
		}
		result, err := server.handleGetFileContent(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success for .env.sample")
		}
	})

	t.Run(".env file blocked (sensitive)", func(t *testing.T) {
		// Create a .env file
		if err := os.WriteFile(filepath.Join(project.RootPath, ".env"), []byte("SECRET=xxx"), 0644); err != nil {
			t.Fatal(err)
		}
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
					"path":       ".env",
				},
			},
		}
		result, err := server.handleGetFileContent(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error for .env")
		}
	})

	t.Run("missing path", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleGetFileContent(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})

	t.Run("nonexistent project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": "nonexistent",
					"path":       "README.md",
				},
			},
		}
		result, err := server.handleGetFileContent(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

// searchFileMatch mirrors the JSON shape emitted by handleSearchFiles.
type searchFileMatch struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Content   string `json:"content"`
	Truncated bool   `json:"truncated"`
}

// searchFileResult mirrors the top-level JSON shape emitted by
// handleSearchFiles for use in assertions.
type searchFileResult struct {
	ProjectID string            `json:"project_id"`
	Pattern   string            `json:"pattern"`
	Matches   []searchFileMatch `json:"matches"`
	Count     int               `json:"count"`
	Truncated bool              `json:"truncated"`
	Skipped   int               `json:"skipped"`
}

// callSearchFiles invokes handleSearchFiles and unmarshals its JSON
// result, failing the test on transport-level error.
func callSearchFiles(t *testing.T, srv *Server, args map[string]interface{}) searchFileResult {
	t.Helper()
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	}
	result, err := srv.handleSearchFiles(context.Background(), req)
	if err != nil {
		t.Fatalf("handleSearchFiles returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Fatalf("expected success, got error result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var fc searchFileResult
	if err := json.Unmarshal([]byte(textContent.Text), &fc); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	return fc
}

func TestHandleSearchFiles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project, _ := tempProjectWithFiles(t, db)

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	// tempProjectWithFiles creates: README.md, main.go, go.mod,
	// internal/app.go, .env.sample, .gitignore

	t.Run("match by filename", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id": project.ID,
			"pattern":    "main\\.go",
		})
		if fc.Count != 1 {
			t.Fatalf("expected 1 match, got %d", len(fc.Matches))
		}
		if fc.Matches[0].Path != "main.go" {
			t.Fatalf("expected main.go, got %q", fc.Matches[0].Path)
		}
		if !strings.Contains(fc.Matches[0].Content, "func main") {
			t.Fatalf("expected content with func main, got %q", fc.Matches[0].Content)
		}
	})

	t.Run("match by directory path", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id": project.ID,
			"pattern":    "^internal/",
		})
		if fc.Count != 1 {
			t.Fatalf("expected 1 match under internal/, got %d", len(fc.Matches))
		}
		if fc.Matches[0].Path != "internal/app.go" {
			t.Fatalf("expected internal/app.go, got %q", fc.Matches[0].Path)
		}
	})

	t.Run("regex alternation matches multiple files", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id": project.ID,
			"pattern":    "(main|app)\\.go",
		})
		if fc.Count != 2 {
			t.Fatalf("expected 2 matches, got %d (%v)", len(fc.Matches), pathsOf(fc.Matches))
		}
	})

	t.Run("include_content=false omits content", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id":      project.ID,
			"pattern":         "main\\.go",
			"include_content": false,
		})
		if fc.Count != 1 {
			t.Fatalf("expected 1 match, got %d", len(fc.Matches))
		}
		if fc.Matches[0].Content != "" {
			t.Fatalf("expected empty content, got %q", fc.Matches[0].Content)
		}
	})

	t.Run("sensitive file is skipped but template is allowed", func(t *testing.T) {
		// .env.sample already exists from the fixture; verify it is returned.
		// Add two files that must never be exposed: a real .env (blocked at
		// walk time by shouldSkip) and a private key (blocked at read time by
		// isSensitiveFile, so it counts toward "skipped").
		writeFile := func(name, content string) {
			p := filepath.Join(project.RootPath, name)
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		writeFile(".env", "SECRET=xxx")
		writeFile("deploy.pem", "PRIVATE KEY MATERIAL")
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id": project.ID,
			"pattern":    `^(\.env|deploy\.pem)`,
		})
		got := pathsOf(fc.Matches)
		if !contains(got, ".env.sample") {
			t.Fatalf("expected .env.sample in matches, got %v", got)
		}
		if contains(got, ".env") {
			t.Fatalf(".env must be excluded, got %v", got)
		}
		if contains(got, "deploy.pem") {
			t.Fatalf("deploy.pem must be excluded, got %v", got)
		}
		if fc.Skipped < 1 {
			t.Fatalf("expected at least 1 skipped sensitive file, got %d", fc.Skipped)
		}
	})

	t.Run("max_results cap truncates results", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id":  project.ID,
			"pattern":     ".*",
			"max_results": float64(1),
		})
		if fc.Count != 1 {
			t.Fatalf("expected 1 match, got %d", fc.Count)
		}
		if !fc.Truncated {
			t.Fatal("expected truncated=true when max_results reached")
		}
	})

	t.Run("no matches returns empty list", func(t *testing.T) {
		fc := callSearchFiles(t, server, map[string]interface{}{
			"project_id": project.ID,
			"pattern":    "this_does_not_exist_xyz",
		})
		if fc.Count != 0 {
			t.Fatalf("expected 0 matches, got %d", fc.Count)
		}
		if fc.Matches == nil {
			t.Fatal("expected non-nil matches slice")
		}
	})

	t.Run("invalid regex", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
					"pattern":    "[",
				},
			},
		}
		result, err := server.handleSearchFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error for invalid regex")
		}
	})

	t.Run("missing pattern", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleSearchFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error for missing pattern")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"pattern": "main"},
			},
		}
		result, err := server.handleSearchFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error for missing project_id")
		}
	})

	t.Run("nonexistent project", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": "nonexistent",
					"pattern":    "main",
				},
			},
		}
		result, err := server.handleSearchFiles(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error for nonexistent project")
		}
	})
}

// pathsOf returns the path field of each match, for readable assertions.
func pathsOf(matches []searchFileMatch) []string {
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = m.Path
	}
	return out
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestHandleGetProjectStructure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project, _ := tempProjectWithFiles(t, db)

	_, err := db.DB().Exec(
		`INSERT INTO metadata (project_id, git_head, default_branch, last_commit_at, last_modified_at,
		 commit_count, language_summary, framework_summary, dependency_summary, documentation_hash, last_scan_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		project.ID, "abc", "main", "now", "now", 10, "Go", "none", "none", "hash", "now",
	)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("structure without content", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleGetProjectStructure(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("structure with content", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id":      project.ID,
					"include_content": true,
				},
			},
		}
		result, err := server.handleGetProjectStructure(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleGetProjectStructure(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestHandleGetDependencies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project, _ := tempProjectWithFiles(t, db)

	// Add some dependencies
	_, err := db.DB().Exec(
		`INSERT INTO dependencies (project_id, name, version, manager, scope)
		 VALUES (?, ?, ?, ?, ?)`,
		project.ID, "github.com/spf13/cobra", "v1.8.0", "go-mod", "dependency",
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DB().Exec(
		`INSERT INTO dependencies (project_id, name, version, manager, scope)
		 VALUES (?, ?, ?, ?, ?)`,
		project.ID, "go.uber.org/zap", "v1.27.0", "go-mod", "dependency",
	)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{DB: db.DB(), Logger: logger, Roots: []string{}}
	server := New(cfg)

	t.Run("list dependencies with repository_type", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"project_id": project.ID},
			},
		}
		result, err := server.handleGetDependencies(context.Background(), req)
		if err != nil || result == nil || result.IsError {
			t.Fatal("expected success")
		}
	})

	t.Run("missing project_id", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
		}
		result, err := server.handleGetDependencies(context.Background(), req)
		if err != nil || !result.IsError {
			t.Fatal("expected error")
		}
	})
}

func TestGetStringArg(t *testing.T) {
	args := map[string]interface{}{
		"string":  "value",
		"number":  123,
		"missing": nil,
	}

	if got := getStringArg(args, "string"); got != "value" {
		t.Errorf("expected 'value', got '%s'", got)
	}

	if got := getStringArg(args, "number"); got != "" {
		t.Errorf("expected empty string for non-string, got '%s'", got)
	}

	if got := getStringArg(args, "missing"); got != "" {
		t.Errorf("expected empty string for missing key, got '%s'", got)
	}
}
