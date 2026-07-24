package mcp

import (
	"context"
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
					"project_id":        project.ID,
					"analyzer":          "test-analyzer",
					"summary":           "Test summary",
					"purpose":           "Test purpose",
					"architecture":      "Test architecture",
					"analyzed_at":       time.Now().UTC().Format(time.RFC3339),
					"analyzed_git_head": "abc123",
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
			t.Fatalf("storeAnalysis should not error for valid data, got error result")
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
