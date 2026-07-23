package store

import (
	"testing"

	"project-dash/pkg/models"
)

func TestAnalysisStore_CreateAnalysis(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project := &models.Project{
		ID:             "proj-1",
		Name:           "Test Project",
		RootPath:       "/test/path",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	analysis := &models.Analysis{
		ID:                 "analysis-1",
		ProjectID:          "proj-1",
		Analyzer:           "test-analyzer",
		AnalyzedGitHead:    "abc123",
		AnalyzedAt:         "2024-01-01T12:00:00Z",
		Summary:            "Test summary",
		Purpose:            "Test purpose",
		Architecture:       "Test architecture",
		Maturity:           "Production",
		Strengths:          "Test strengths",
		Weaknesses:         "Test weaknesses",
		ReusableComponents: "Test components",
		Notes:              "Test notes",
		RawJSON:            `{"test": "data"}`,
	}

	if err := store.analyses.CreateAnalysis(analysis); err != nil {
		t.Fatalf("failed to create analysis: %v", err)
	}

	retrieved, err := store.analyses.GetAnalysis("analysis-1")
	if err != nil {
		t.Fatalf("failed to get analysis: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected analysis to be retrieved, got nil")
	}

	if retrieved.ProjectID != analysis.ProjectID {
		t.Errorf("expected project_id %s, got %s", analysis.ProjectID, retrieved.ProjectID)
	}

	if retrieved.Analyzer != analysis.Analyzer {
		t.Errorf("expected analyzer %s, got %s", analysis.Analyzer, retrieved.Analyzer)
	}

	if retrieved.Maturity != analysis.Maturity {
		t.Errorf("expected maturity %s, got %s", analysis.Maturity, retrieved.Maturity)
	}

	if retrieved.Strengths != analysis.Strengths {
		t.Errorf("expected strengths %s, got %s", analysis.Strengths, retrieved.Strengths)
	}
}

func TestAnalysisStore_OverwriteByAnalyzer(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project := &models.Project{
		ID:             "proj-1",
		Name:           "Test Project",
		RootPath:       "/test/path",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	analysis1 := &models.Analysis{
		ID:                 "analysis-1",
		ProjectID:          "proj-1",
		Analyzer:           "test-analyzer",
		AnalyzedGitHead:    "abc123",
		AnalyzedAt:         "2024-01-01T12:00:00Z",
		Summary:            "Original summary",
		Purpose:            "Original purpose",
		Architecture:       "Original architecture",
		Maturity:           "Development",
		Strengths:          "Original strengths",
		Weaknesses:         "Original weaknesses",
		ReusableComponents: "Original components",
		Notes:              "Original notes",
		RawJSON:            `{"test": "original"}`,
	}

	if err := store.analyses.CreateAnalysis(analysis1); err != nil {
		t.Fatalf("failed to create first analysis: %v", err)
	}

	analysis2 := &models.Analysis{
		ID:                 "analysis-2",
		ProjectID:          "proj-1",
		Analyzer:           "test-analyzer",
		AnalyzedGitHead:    "def456",
		AnalyzedAt:         "2024-01-02T12:00:00Z",
		Summary:            "Updated summary",
		Purpose:            "Updated purpose",
		Architecture:       "Updated architecture",
		Maturity:           "Production",
		Strengths:          "Updated strengths",
		Weaknesses:         "Updated weaknesses",
		ReusableComponents: "Updated components",
		Notes:              "Updated notes",
		RawJSON:            `{"test": "updated"}`,
	}

	if err := store.analyses.CreateAnalysis(analysis2); err != nil {
		t.Fatalf("failed to create second analysis: %v", err)
	}

	retrieved, err := store.analyses.GetAnalysisByProjectAndAnalyzer("proj-1", "test-analyzer")
	if err != nil {
		t.Fatalf("failed to get analysis by project and analyzer: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected analysis to be retrieved, got nil")
	}

	if retrieved.Summary != "Updated summary" {
		t.Errorf("expected summary 'Updated summary', got '%s'", retrieved.Summary)
	}

	if retrieved.Maturity != "Production" {
		t.Errorf("expected maturity 'Production', got '%s'", retrieved.Maturity)
	}

	analyses, err := store.analyses.ListAnalyses("proj-1")
	if err != nil {
		t.Fatalf("failed to list analyses: %v", err)
	}

	if len(analyses) != 1 {
		t.Errorf("expected 1 analysis after overwrite, got %d", len(analyses))
	}
}

func TestAnalysisStore_GetAnalysisByProjectAndAnalyzer(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project := &models.Project{
		ID:             "proj-1",
		Name:           "Test Project",
		RootPath:       "/test/path",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	analysis := &models.Analysis{
		ID:                 "analysis-1",
		ProjectID:          "proj-1",
		Analyzer:           "test-analyzer",
		AnalyzedGitHead:    "abc123",
		AnalyzedAt:         "2024-01-01T12:00:00Z",
		Summary:            "Test summary",
		Purpose:            "Test purpose",
		Architecture:       "Test architecture",
		Maturity:           "Production",
		Strengths:          "Test strengths",
		Weaknesses:         "Test weaknesses",
		ReusableComponents: "Test components",
		Notes:              "Test notes",
		RawJSON:            `{"test": "data"}`,
	}

	if err := store.analyses.CreateAnalysis(analysis); err != nil {
		t.Fatalf("failed to create analysis: %v", err)
	}

	retrieved, err := store.analyses.GetAnalysisByProjectAndAnalyzer("proj-1", "test-analyzer")
	if err != nil {
		t.Fatalf("failed to get analysis: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected analysis to be retrieved, got nil")
	}

	if retrieved.ID != "analysis-1" {
		t.Errorf("expected ID 'analysis-1', got '%s'", retrieved.ID)
	}

	nonExistent, err := store.analyses.GetAnalysisByProjectAndAnalyzer("proj-1", "non-existent")
	if err != nil {
		t.Fatalf("failed to query non-existent analysis: %v", err)
	}

	if nonExistent != nil {
		t.Error("expected nil for non-existent analysis, got value")
	}
}

func TestAnalysisStore_ListAnalyses(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project := &models.Project{
		ID:             "proj-1",
		Name:           "Test Project",
		RootPath:       "/test/path",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	analyses := []*models.Analysis{
		{
			ID:                 "analysis-1",
			ProjectID:          "proj-1",
			Analyzer:           "analyzer-1",
			AnalyzedGitHead:    "abc123",
			AnalyzedAt:         "2024-01-03T12:00:00Z",
			Summary:            "Summary 1",
			Purpose:            "Purpose 1",
			Architecture:       "Architecture 1",
			Maturity:           "Production",
			Strengths:          "Strengths 1",
			Weaknesses:         "Weaknesses 1",
			ReusableComponents: "Components 1",
			Notes:              "Notes 1",
			RawJSON:            `{"test": "1"}`,
		},
		{
			ID:                 "analysis-2",
			ProjectID:          "proj-1",
			Analyzer:           "analyzer-2",
			AnalyzedGitHead:    "def456",
			AnalyzedAt:         "2024-01-02T12:00:00Z",
			Summary:            "Summary 2",
			Purpose:            "Purpose 2",
			Architecture:       "Architecture 2",
			Maturity:           "Development",
			Strengths:          "Strengths 2",
			Weaknesses:         "Weaknesses 2",
			ReusableComponents: "Components 2",
			Notes:              "Notes 2",
			RawJSON:            `{"test": "2"}`,
		},
	}

	for _, analysis := range analyses {
		if err := store.analyses.CreateAnalysis(analysis); err != nil {
			t.Fatalf("failed to create analysis: %v", err)
		}
	}

	retrieved, err := store.analyses.ListAnalyses("proj-1")
	if err != nil {
		t.Fatalf("failed to list analyses: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 analyses, got %d", len(retrieved))
	}

	if retrieved[0].Analyzer != "analyzer-1" {
		t.Errorf("expected first analyzer to be 'analyzer-1', got '%s'", retrieved[0].Analyzer)
	}
}

func TestAnalysisStore_DeleteAllForProject(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project := &models.Project{
		ID:             "proj-1",
		Name:           "Test Project",
		RootPath:       "/test/path",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	analysis := &models.Analysis{
		ID:                 "analysis-1",
		ProjectID:          "proj-1",
		Analyzer:           "test-analyzer",
		AnalyzedGitHead:    "abc123",
		AnalyzedAt:         "2024-01-01T12:00:00Z",
		Summary:            "Test summary",
		Purpose:            "Test purpose",
		Architecture:       "Test architecture",
		Maturity:           "Production",
		Strengths:          "Test strengths",
		Weaknesses:         "Test weaknesses",
		ReusableComponents: "Test components",
		Notes:              "Test notes",
		RawJSON:            `{"test": "data"}`,
	}

	if err := store.analyses.CreateAnalysis(analysis); err != nil {
		t.Fatalf("failed to create analysis: %v", err)
	}

	if err := store.analyses.DeleteAllForProject("proj-1"); err != nil {
		t.Fatalf("failed to delete analyses for project: %v", err)
	}

	retrieved, err := store.analyses.ListAnalyses("proj-1")
	if err != nil {
		t.Fatalf("failed to list analyses after deletion: %v", err)
	}

	if len(retrieved) != 0 {
		t.Errorf("expected 0 analyses after deletion, got %d", len(retrieved))
	}
}