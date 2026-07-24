package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaleDetector_IsOutdated_GitHeadMismatch(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	currentGitHead := "def456"
	mockStore.gitHeads[testProjectID] = &currentGitHead

	analysis := &Analysis{
		ID:              uuid.New().String(),
		ProjectID:       testProjectID.String(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
	}

	detector := NewStaleDetector(mockStore)
	outdated, err := detector.IsOutdated(context.Background(), analysis)
	require.NoError(t, err)
	assert.True(t, outdated)
}

func TestStaleDetector_IsOutdated_GitHeadMatch(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	currentGitHead := "abc123"
	mockStore.gitHeads[testProjectID] = &currentGitHead

	analysis := &Analysis{
		ID:              uuid.New().String(),
		ProjectID:       testProjectID.String(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
	}

	detector := NewStaleDetector(mockStore)
	outdated, err := detector.IsOutdated(context.Background(), analysis)
	require.NoError(t, err)
	assert.False(t, outdated)
}

func TestStaleDetector_IsOutdated_NULLGitHead(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	// No git head set (nil)

	analysis := &Analysis{
		ID:              uuid.New().String(),
		ProjectID:       testProjectID.String(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
	}

	detector := NewStaleDetector(mockStore)
	outdated, err := detector.IsOutdated(context.Background(), analysis)
	require.NoError(t, err)
	assert.True(t, outdated)
}

func TestAnalysisService_ListProjectsNeedingAnalysis_Unanalyzed(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID1 := uuid.New()
	testProjectID2 := uuid.New()

	// Create one analyzed project and one unanalyzed project
	mockStore.projects[testProjectID1] = true
	mockStore.projects[testProjectID2] = true

	analyzedAnalysis := &Analysis{
		ID:              uuid.New().String(),
		ProjectID:       testProjectID1.String(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
		AnalyzedAt:      time.Now().UTC(),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	mockStore.analyses[analyzedAnalysis.ID] = analyzedAnalysis

	currentGitHead := "abc123"
	mockStore.gitHeads[testProjectID1] = &currentGitHead

	// Note: This test uses the actual SQL query, which requires a real database
	// For now, we'll skip this test as it needs integration testing with a database
	t.Skip("Requires integration test with real database")
}

func TestAnalysisService_ListProjectsNeedingAnalysis_Outdated(t *testing.T) {
	// Note: This test requires a real database to properly test the SQL query
	t.Skip("Requires integration test with real database")
}

func TestAnalysisService_ListProjectsNeedingAnalysis_Empty(t *testing.T) {
	// Note: This test requires a real database to properly test the SQL query
	t.Skip("Requires integration test with real database")
}