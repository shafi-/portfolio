package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"project-dash/internal/logging"
)

// Integration tests for listProjectsNeedingAnalysis tool
// These tests verify the complete data flow from database to MCP response

func TestListProjectsNeedingAnalysis_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}
	server := New(cfg)

	t.Run("response structure validation", func(t *testing.T) {
		// Create a project with stale analysis
		project := createTestProject(t, db)
		analyzedAt := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project.ID, "old123", analyzedAt)
		createTestMetadata(t, db, project.ID, "current456")

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		result, err := server.handleListProjectsNeedingAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		// Verify response is not nil and not error
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.IsError {
			t.Fatal("expected non-error response")
		}
		if len(result.Content) == 0 {
			t.Fatal("expected content in response")
		}

		// Verify content type
		content, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatalf("expected TextContent, got %T", result.Content[0])
		}

		// Verify JSON is valid
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		// Verify required top-level fields
		requiredFields := []string{"no_analysis", "stale_analysis", "counts"}
		for _, field := range requiredFields {
			if _, ok := response[field]; !ok {
				t.Errorf("missing required field: %s", field)
			}
		}

		// Verify counts structure
		counts, ok := response["counts"].(map[string]interface{})
		if !ok {
			t.Fatal("counts is not a map")
		}

		requiredCountFields := []string{"no_analysis", "stale_analysis", "total"}
		for _, field := range requiredCountFields {
			if _, ok := counts[field]; !ok {
				t.Errorf("missing required count field: %s", field)
			}
		}
	})

	t.Run("no_analysis structure validation", func(t *testing.T) {
		db.DB().Exec("DELETE FROM projects")
		db.DB().Exec("DELETE FROM metadata")

		project := createTestProject(t, db)
		createTestMetadata(t, db, project.ID, "git123")

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		result, err := server.handleListProjectsNeedingAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		var response map[string]interface{}
		if content, ok := result.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
		}

		noAnalysis := response["no_analysis"].([]interface{})
		if len(noAnalysis) == 0 {
			t.Fatal("expected at least one no_analysis project")
		}

		// Verify structure of no_analysis entry
		projectData, ok := noAnalysis[0].(map[string]interface{})
		if !ok {
			t.Fatal("no_analysis entry is not a map")
		}

		requiredFields := []string{"id", "name", "path"}
		for _, field := range requiredFields {
			if _, ok := projectData[field]; !ok {
				t.Errorf("missing required field in no_analysis: %s", field)
			}
		}

		// Verify field types
		if _, ok := projectData["id"].(string); !ok {
			t.Error("id field is not a string")
		}
		if _, ok := projectData["name"].(string); !ok {
			t.Error("name field is not a string")
		}
		if _, ok := projectData["path"].(string); !ok {
			t.Error("path field is not a string")
		}

		// Verify no extra fields
		expectedFieldCount := 3
		if len(projectData) != expectedFieldCount {
			t.Errorf("expected %d fields in no_analysis entry, got %d", expectedFieldCount, len(projectData))
		}
	})

	t.Run("stale_analysis structure validation", func(t *testing.T) {
		db.DB().Exec("DELETE FROM projects")
		db.DB().Exec("DELETE FROM metadata")
		db.DB().Exec("DELETE FROM analyses")

		project := createTestProject(t, db)
		analyzedAt := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project.ID, "old123", analyzedAt)
		createTestMetadata(t, db, project.ID, "new456")

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		result, err := server.handleListProjectsNeedingAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		var response map[string]interface{}
		if content, ok := result.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
		}

		staleAnalysis := response["stale_analysis"].([]interface{})
		if len(staleAnalysis) == 0 {
			t.Fatal("expected at least one stale_analysis project")
		}

		// Verify structure of stale_analysis entry
		projectData, ok := staleAnalysis[0].(map[string]interface{})
		if !ok {
			t.Fatal("stale_analysis entry is not a map")
		}

		requiredFields := []string{"id", "name", "path", "analyzed_at", "analyzed_git_head", "current_git_head"}
		for _, field := range requiredFields {
			if _, ok := projectData[field]; !ok {
				t.Errorf("missing required field in stale_analysis: %s", field)
			}
		}

		// Verify field types
		if _, ok := projectData["id"].(string); !ok {
			t.Error("id field is not a string")
		}
		if _, ok := projectData["name"].(string); !ok {
			t.Error("name field is not a string")
		}
		if _, ok := projectData["path"].(string); !ok {
			t.Error("path field is not a string")
		}
		if _, ok := projectData["analyzed_at"].(string); !ok {
			t.Error("analyzed_at field is not a string")
		}
		if _, ok := projectData["analyzed_git_head"].(string); !ok {
			t.Error("analyzed_git_head field is not a string")
		}
		if _, ok := projectData["current_git_head"].(string); !ok {
			t.Error("current_git_head field is not a string")
		}

		// Verify git heads are different
		analyzedGitHead := projectData["analyzed_git_head"].(string)
		currentGitHead := projectData["current_git_head"].(string)
		if analyzedGitHead == currentGitHead {
			t.Error("expected different git heads for stale analysis")
		}
	})

	t.Run("counts accuracy validation", func(t *testing.T) {
		db.DB().Exec("DELETE FROM projects")
		db.DB().Exec("DELETE FROM metadata")
		db.DB().Exec("DELETE FROM analyses")

		// Create multiple projects
		project1 := createTestProject(t, db)
		project2 := createTestProject(t, db)
		project3 := createTestProject(t, db)
		_ = createTestProject(t, db)

		// Project 1: No analysis
		createTestMetadata(t, db, project1.ID, "git1")

		// Project 2: Stale analysis
		analyzedAt2 := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project2.ID, "old2", analyzedAt2)
		createTestMetadata(t, db, project2.ID, "new2")

		// Project 3: Up-to-date analysis
		analyzedAt3 := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project3.ID, "same3", analyzedAt3)
		createTestMetadata(t, db, project3.ID, "same3")

		// Project 4: No analysis, no metadata
		// (no metadata)

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		result, err := server.handleListProjectsNeedingAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		var response map[string]interface{}
		if content, ok := result.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
		}

		noAnalysis := response["no_analysis"].([]interface{})
		staleAnalysis := response["stale_analysis"].([]interface{})
		counts := response["counts"].(map[string]interface{})

		// Verify counts match actual array lengths
		noAnalysisCount := int(counts["no_analysis"].(float64))
		staleAnalysisCount := int(counts["stale_analysis"].(float64))
		totalCount := int(counts["total"].(float64))

		if noAnalysisCount != len(noAnalysis) {
			t.Errorf("no_analysis count mismatch: counts says %d, array has %d",
				noAnalysisCount, len(noAnalysis))
		}
		if staleAnalysisCount != len(staleAnalysis) {
			t.Errorf("stale_analysis count mismatch: counts says %d, array has %d",
				staleAnalysisCount, len(staleAnalysis))
		}
		if totalCount != (len(noAnalysis) + len(staleAnalysis)) {
			t.Errorf("total count mismatch: counts says %d, actual total is %d",
				totalCount, len(noAnalysis)+len(staleAnalysis))
		}

		// Verify expected counts
		// project1 (no analysis) + project4 (no metadata) = 2 no_analysis
		// project2 (stale) = 1 stale_analysis
		// project3 (up-to-date) = excluded
		expectedNoAnalysis := 2
		expectedStaleAnalysis := 1
		expectedTotal := 3

		if noAnalysisCount != expectedNoAnalysis {
			t.Errorf("no_analysis count: expected %d, got %d", expectedNoAnalysis, noAnalysisCount)
		}
		if staleAnalysisCount != expectedStaleAnalysis {
			t.Errorf("stale_analysis count: expected %d, got %d", expectedStaleAnalysis, staleAnalysisCount)
		}
		if totalCount != expectedTotal {
			t.Errorf("total count: expected %d, got %d", expectedTotal, totalCount)
		}
	})

	t.Run("correctness: git head comparison", func(t *testing.T) {
		db.DB().Exec("DELETE FROM projects")
		db.DB().Exec("DELETE FROM metadata")
		db.DB().Exec("DELETE FROM analyses")

		// Create project with multiple analyses
		project := createTestProject(t, db)

		// Create old analysis
		oldAnalyzedAt := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project.ID, "oldCommit", oldAnalyzedAt)

		// Create newer analysis (different analyzer to avoid UNIQUE constraint)
		newAnalyzedAt := time.Now().Add(-12 * time.Hour).UTC().Format(time.RFC3339)
		createTestAnalysis(t, db, project.ID, "newCommit", newAnalyzedAt, "test-analyzer-2")

		// Current git head is different from both
		createTestMetadata(t, db, project.ID, "currentCommit")

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		result, err := server.handleListProjectsNeedingAnalysis(context.Background(), req)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		var response map[string]interface{}
		if content, ok := result.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
		}

		staleAnalysis := response["stale_analysis"].([]interface{})
		if len(staleAnalysis) != 1 {
			t.Fatalf("expected 1 stale_analysis project, got %d", len(staleAnalysis))
		}

		// Should compare against the latest analysis (most recent analyzed_at)
		projectData := staleAnalysis[0].(map[string]interface{})
		analyzedGitHead := projectData["analyzed_git_head"].(string)

		// Should be "newCommit" (latest), not "oldCommit"
		if analyzedGitHead != "newCommit" {
			t.Errorf("expected analyzed_git_head to be 'newCommit' (latest), got '%s'", analyzedGitHead)
		}

		currentGitHead := projectData["current_git_head"].(string)
		if currentGitHead != "currentCommit" {
			t.Errorf("expected current_git_head 'currentCommit', got '%s'", currentGitHead)
		}
	})
}

// Test data consistency across tools
func TestListProjectsNeedingAnalysis_Consistency(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.NewLogger("INFO", "console")
	cfg := &Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  []string{},
	}
	server := New(cfg)

	t.Run("consistency with getAnalysis", func(t *testing.T) {
		db.DB().Exec("DELETE FROM projects")
		db.DB().Exec("DELETE FROM metadata")
		db.DB().Exec("DELETE FROM analyses")

		project := createTestProject(t, db)
		analyzedAt := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		analysis := createTestAnalysis(t, db, project.ID, "old123", analyzedAt)
		createTestMetadata(t, db, project.ID, "new456")

		// Get analyses via getAnalysis
		getAnalysisReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": project.ID,
				},
			},
		}

		getAnalysisResult, err := server.handleGetAnalysis(context.Background(), getAnalysisReq)
		if err != nil {
			t.Fatalf("handleGetAnalysis failed: %v", err)
		}

		var getAnalysisResp map[string]interface{}
		if content, ok := getAnalysisResult.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &getAnalysisResp); err != nil {
				t.Fatalf("failed to parse getAnalysis response: %v", err)
			}
		}

		// Get needing analysis list
		needingAnalysisReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{},
		}

		needingAnalysisResult, err := server.handleListProjectsNeedingAnalysis(context.Background(), needingAnalysisReq)
		if err != nil {
			t.Fatalf("handleListProjectsNeedingAnalysis failed: %v", err)
		}

		var needingAnalysisResp map[string]interface{}
		if content, ok := needingAnalysisResult.Content[0].(mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(content.Text), &needingAnalysisResp); err != nil {
				t.Fatalf("failed to parse listProjectsNeedingAnalysis response: %v", err)
			}
		}

		// Verify project appears in stale_analysis
		staleAnalysis := needingAnalysisResp["stale_analysis"].([]interface{})
		if len(staleAnalysis) != 1 {
			t.Fatalf("expected 1 stale_analysis project, got %d", len(staleAnalysis))
		}

		staleProject := staleAnalysis[0].(map[string]interface{})

		// Verify analyzed_git_head matches what getAnalysis returned
		staleGitHead := staleProject["analyzed_git_head"].(string)

		analyses := getAnalysisResp["analyses"].([]interface{})
		if len(analyses) == 0 {
			t.Fatal("getAnalysis returned no analyses")
		}

		firstAnalysis := analyses[0].(map[string]interface{})
		getAnalysisGitHead := firstAnalysis["analyzed_git_head"].(string)

		if staleGitHead != getAnalysisGitHead {
			t.Errorf("git head mismatch: listProjectsNeedingAnalysis says '%s', getAnalysis says '%s'",
				staleGitHead, getAnalysisGitHead)
		}

		// Verify they match the original analysis
		if staleGitHead != analysis.AnalyzedGitHead {
			t.Errorf("git head doesn't match original analysis: expected '%s', got '%s'",
				analysis.AnalyzedGitHead, staleGitHead)
		}
	})
}
