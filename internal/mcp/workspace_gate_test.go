package mcp

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// gitIn runs a git command inside dir, failing the test on error.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v (%s)", args, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

// initRepoProject creates a test project whose root is a real temp git repo
// and returns the project and its current HEAD.
func initRepoProject(t *testing.T, db *database.Database, name string) (*models.Project, string) {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "checkout", "-q", "-b", "main")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	head := gitIn(t, dir, "rev-parse", "HEAD")

	project := createTestProject(t, db)
	// Point the stored project at the real repo.
	if _, err := db.DB().Exec(`UPDATE projects SET root_path = ?, name = ? WHERE id = ?`, dir, name, project.ID); err != nil {
		t.Fatalf("failed to point project at temp repo: %v", err)
	}
	project.RootPath = dir
	project.Name = name
	return project, head
}

func newTestServer(t *testing.T, db *database.Database) *Server {
	t.Helper()
	logger, _ := logging.NewLogger("INFO", "console")
	return New(&Config{DB: db.DB(), Logger: logger, Roots: []string{}})
}

func toolJSON(t *testing.T, result *mcp.CallToolResult) map[string]interface{} {
	t.Helper()
	if result == nil || result.IsError {
		t.Fatalf("expected non-error result, got %+v", result)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v (%s)", err, content.Text)
	}
	return response
}

func TestHandleGetAnalysis_IncludesFreshness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	server := newTestServer(t, db)

	project, head := initRepoProject(t, db, "orders-service")
	// Analysis made against the initial HEAD; repo has since moved on.
	createTestAnalysis(t, db, project.ID, head, time.Now().Add(-24*time.Hour).UTC().Format(time.RFC3339))
	createTestMetadata(t, db, project.ID, head)
	gitIn(t, project.RootPath, "commit", "-q", "--allow-empty", "-m", "newer work")

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"project_id": project.ID},
	}}
	result, err := server.handleGetAnalysis(context.Background(), req)
	if err != nil {
		t.Fatalf("handleGetAnalysis failed: %v", err)
	}

	response := toolJSON(t, result)
	freshness, ok := response["freshness"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected freshness object in response, got: %v", response["freshness"])
	}
	if freshness["status"] != "stale" {
		t.Fatalf("expected stale status, got %v", freshness["status"])
	}
	if _, ok := freshness["commits_behind"]; !ok {
		t.Fatalf("expected commits_behind for stale analysis, got %v", freshness)
	}
}

func TestHandleGetAnalysis_FreshWhenHeadMatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	server := newTestServer(t, db)

	project, head := initRepoProject(t, db, "payments-service")
	createTestAnalysis(t, db, project.ID, head, time.Now().UTC().Format(time.RFC3339))
	createTestMetadata(t, db, project.ID, head)

	req := mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"project_id": project.ID},
	}}
	result, err := server.handleGetAnalysis(context.Background(), req)
	if err != nil {
		t.Fatalf("handleGetAnalysis failed: %v", err)
	}

	freshness := toolJSON(t, result)["freshness"].(map[string]interface{})
	if freshness["status"] != "fresh" {
		t.Fatalf("expected fresh status, got %v", freshness["status"])
	}
}

func TestHandleGetProjectAnalyzerPrompt_WithAndWithoutProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	server := newTestServer(t, db)

	// Without project_id: the unchanged static prompt.
	req := mcp.CallToolRequest{Params: mcp.CallToolParams{}}
	result, err := server.handleGetProjectAnalyzerPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("prompt without project failed: %v", err)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if strings.Contains(text, "FRESHNESS:") || strings.Contains(text, "PREVIOUS ANALYSIS") {
		t.Fatal("static prompt must not contain per-project sections")
	}

	// With project_id: freshness banner + previous analysis seed + user gate.
	project, head := initRepoProject(t, db, "inventory-service")
	createTestAnalysis(t, db, project.ID, head, time.Now().UTC().Format(time.RFC3339))
	createTestMetadata(t, db, project.ID, head)
	gitIn(t, project.RootPath, "commit", "-q", "--allow-empty", "-m", "drift")

	req = mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"project_id": project.ID},
	}}
	result, err = server.handleGetProjectAnalyzerPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("prompt with project failed: %v", err)
	}
	text = result.Content[0].(mcp.TextContent).Text

	for _, want := range []string{"FRESHNESS:", "STALE", "PREVIOUS ANALYSIS", "inventory-service", "ask whether to refresh"} {
		if !strings.Contains(text, want) {
			t.Fatalf("prompt missing %q in:\n%s", want, text)
		}
	}

	// Unknown project → tool error.
	req = mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"project_id": "does-not-exist"},
	}}
	result, err = server.handleGetProjectAnalyzerPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("prompt with unknown project failed: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for unknown project")
	}
}

func TestHandleWorkspaceTools(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	server := newTestServer(t, db)

	// listWorkspaces on empty store.
	result, err := server.handleListWorkspaces(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{}})
	if err != nil {
		t.Fatalf("handleListWorkspaces failed: %v", err)
	}
	response := toolJSON(t, result)
	if response["count"].(float64) != 0 {
		t.Fatalf("expected 0 workspaces, got %v", response["count"])
	}

	// Create workspace + members through the store (management is CLI-side).
	projectA, _ := initRepoProject(t, db, "orders-service")
	if _, err := db.DB().Exec(
		`INSERT INTO workspaces (id, name, description, created_at, updated_at) VALUES ('ws-1', 'commerce', 'Commerce services', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("failed to insert workspace: %v", err)
	}
	if _, err := db.DB().Exec(
		`INSERT INTO workspace_projects (workspace_id, project_id) VALUES ('ws-1', ?)`, projectA.ID); err != nil {
		t.Fatalf("failed to insert membership: %v", err)
	}

	// listWorkspaces shows the member count.
	result, err = server.handleListWorkspaces(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{}})
	if err != nil {
		t.Fatalf("handleListWorkspaces failed: %v", err)
	}
	response = toolJSON(t, result)
	if response["count"].(float64) != 1 {
		t.Fatalf("expected 1 workspace, got %v", response["count"])
	}

	// getWorkspace by name returns members with freshness.
	result, err = server.handleGetWorkspace(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"workspace": "commerce"},
	}})
	if err != nil {
		t.Fatalf("handleGetWorkspace failed: %v", err)
	}
	response = toolJSON(t, result)
	if response["name"] != "commerce" {
		t.Fatalf("expected workspace commerce, got %v", response["name"])
	}
	members, ok := response["members"].([]interface{})
	if !ok || len(members) != 1 {
		t.Fatalf("expected 1 member, got %v", response["members"])
	}
	member := members[0].(map[string]interface{})
	if member["name"] != "orders-service" {
		t.Fatalf("unexpected member: %v", member)
	}
	if _, ok := member["analysis"].(map[string]interface{}); !ok {
		t.Fatalf("member missing analysis freshness: %v", member)
	}

	// getWorkspace with unknown name is a tool error.
	result, err = server.handleGetWorkspace(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]interface{}{"workspace": "ghost"},
	}})
	if err != nil {
		t.Fatalf("handleGetWorkspace(ghost) failed: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for unknown workspace")
	}
}
