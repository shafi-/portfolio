package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type MockAnalysisStore struct {
	projects      map[uuid.UUID]bool
	analyses      map[string]*Analysis
	features      map[string][]Feature
	gitHeads      map[uuid.UUID]*string
	relationships map[string]*Relationship
}

func NewMockAnalysisStore() *MockAnalysisStore {
	return &MockAnalysisStore{
		projects:      make(map[uuid.UUID]bool),
		analyses:      make(map[string]*Analysis),
		features:      make(map[string][]Feature),
		gitHeads:      make(map[uuid.UUID]*string),
		relationships: make(map[string]*Relationship),
	}
}

func (m *MockAnalysisStore) ProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	return m.projects[projectID], nil
}

func (m *MockAnalysisStore) CreateAnalysis(ctx context.Context, analysis *Analysis) error {
	m.analyses[analysis.ID] = analysis
	return nil
}

func (m *MockAnalysisStore) UpdateAnalysis(ctx context.Context, analysis *Analysis) error {
	m.analyses[analysis.ID] = analysis
	return nil
}

func (m *MockAnalysisStore) GetAnalysis(ctx context.Context, projectID uuid.UUID) (*Analysis, error) {
	for _, analysis := range m.analyses {
		if analysis.ProjectID == projectID.String() {
			return analysis, nil
		}
	}
	return nil, nil
}

func (m *MockAnalysisStore) GetAnalysisByAnalyzer(ctx context.Context, projectID uuid.UUID, analyzer string) (*Analysis, error) {
	for _, analysis := range m.analyses {
		if analysis.ProjectID == projectID.String() && analysis.Analyzer == analyzer {
			return analysis, nil
		}
	}
	return nil, nil
}

func (m *MockAnalysisStore) DeleteAnalysis(ctx context.Context, projectID uuid.UUID, analyzer string) error {
	for id, analysis := range m.analyses {
		if analysis.ProjectID == projectID.String() && analysis.Analyzer == analyzer {
			delete(m.analyses, id)
			delete(m.features, id)
			return nil
		}
	}
	return nil
}

func (m *MockAnalysisStore) CreateFeatures(ctx context.Context, features []Feature) error {
	for _, feature := range features {
		m.features[feature.AnalysisID] = append(m.features[feature.AnalysisID], feature)
	}
	return nil
}

func (m *MockAnalysisStore) DeleteFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) error {
	delete(m.features, analysisID.String())
	return nil
}

func (m *MockAnalysisStore) GetFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]Feature, error) {
	return m.features[analysisID.String()], nil
}

func (m *MockAnalysisStore) GetGitHeadForProject(ctx context.Context, projectID uuid.UUID) (*string, error) {
	return m.gitHeads[projectID], nil
}

func (m *MockAnalysisStore) ListAllAnalyses(ctx context.Context) ([]Analysis, error) {
	var analyses []Analysis
	for _, analysis := range m.analyses {
		analyses = append(analyses, *analysis)
	}
	return analyses, nil
}

func (m *MockAnalysisStore) CreateRelationship(ctx context.Context, rel *Relationship) error {
	m.relationships[rel.ID] = rel
	return nil
}

func (m *MockAnalysisStore) UpdateRelationship(ctx context.Context, rel *Relationship) error {
	m.relationships[rel.ID] = rel
	return nil
}

func (m *MockAnalysisStore) GetRelationship(ctx context.Context, id uuid.UUID) (*Relationship, error) {
	return m.relationships[id.String()], nil
}

func (m *MockAnalysisStore) ListRelationshipsByProject(ctx context.Context, projectID uuid.UUID) ([]Relationship, error) {
	var relationships []Relationship
	for _, rel := range m.relationships {
		if rel.SourceProject == projectID.String() || rel.TargetProject == projectID.String() {
			relationships = append(relationships, *rel)
		}
	}
	return relationships, nil
}

func (m *MockAnalysisStore) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	delete(m.relationships, id.String())
	return nil
}

func (m *MockAnalysisStore) FindExistingRelationship(ctx context.Context, source, target uuid.UUID, relType string) (*Relationship, error) {
	for _, rel := range m.relationships {
		if rel.SourceProject == source.String() && rel.TargetProject == target.String() && rel.Type == relType {
			return rel, nil
		}
	}
	return nil, nil
}

func (m *MockAnalysisStore) ListProjectsNeedingAnalysis(ctx context.Context) ([]NeedsAnalysisResult, error) {
	// Simple mock implementation - returns empty list for now
	return []NeedsAnalysisResult{}, nil
}

func TestAnalysisService_StoreAnalysis_Success(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	input := AnalysisInput{
		Summary:         "Test summary",
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Name:       "Authentication",
				Confidence: floatPtr(0.95),
			},
		},
	}

	analysis, err := service.StoreAnalysis(context.Background(), testProjectID, input)
	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, testProjectID.String(), analysis.ProjectID)
	assert.Equal(t, input.Summary, analysis.Summary)
	assert.Equal(t, input.Analyzer, analysis.Analyzer)
	assert.Len(t, analysis.Features, 1)
	assert.Equal(t, "Authentication", analysis.Features[0].Name)
}

func TestAnalysisService_StoreAnalysis_UpdateExisting(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	// Create existing analysis
	existingAnalysis := &Analysis{
		ID:         uuid.New().String(),
		ProjectID:  testProjectID.String(),
		Analyzer:   "test-analyzer",
		Summary:    "Old summary",
		AnalyzedAt: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	mockStore.analyses[existingAnalysis.ID] = existingAnalysis
	mockStore.features[existingAnalysis.ID] = []Feature{
		{
			ID:         uuid.New().String(),
			AnalysisID: existingAnalysis.ID,
			Name:       "Old Feature",
		},
	}

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	input := AnalysisInput{
		Summary:         "New summary",
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Name:       "New Feature",
				Confidence: floatPtr(0.90),
			},
		},
	}

	analysis, err := service.StoreAnalysis(context.Background(), testProjectID, input)
	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "New summary", analysis.Summary)
	assert.Len(t, analysis.Features, 1)
	assert.Equal(t, "New Feature", analysis.Features[0].Name)
}

func TestAnalysisService_StoreAnalysis_InvalidProject(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	input := AnalysisInput{
		Summary:         "Test summary",
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
	}

	_, err := service.StoreAnalysis(context.Background(), testProjectID, input)
	assert.Error(t, err)
	assert.Equal(t, ErrProjectNotFound, err)
}

func TestAnalysisService_StoreAnalysis_ValidationError(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	// Missing required field
	input := AnalysisInput{
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "test-analyzer",
	}

	_, err := service.StoreAnalysis(context.Background(), testProjectID, input)
	assert.Error(t, err)
}

func TestAnalysisService_GetAnalysis_Exists(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	// Create existing analysis
	existingAnalysis := &Analysis{
		ID:         uuid.New().String(),
		ProjectID:  testProjectID.String(),
		Analyzer:   "test-analyzer",
		Summary:    "Test summary",
		AnalyzedAt: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	mockStore.analyses[existingAnalysis.ID] = existingAnalysis

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	analysis, err := service.GetAnalysis(context.Background(), testProjectID)
	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "Test summary", analysis.Summary)
}

func TestAnalysisService_GetAnalysis_NotExists(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	analysis, err := service.GetAnalysis(context.Background(), testProjectID)
	require.NoError(t, err)
	assert.Nil(t, analysis)
}

func TestAnalysisService_GetAnalysisByAnalyzer_Exists(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	// Create existing analysis
	existingAnalysis := &Analysis{
		ID:         uuid.New().String(),
		ProjectID:  testProjectID.String(),
		Analyzer:   "test-analyzer",
		Summary:    "Test summary",
		AnalyzedAt: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	mockStore.analyses[existingAnalysis.ID] = existingAnalysis

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	analysis, err := service.GetAnalysisByAnalyzer(context.Background(), testProjectID, "test-analyzer")
	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "test-analyzer", analysis.Analyzer)
}

func TestAnalysisService_GetAnalysisByAnalyzer_NotExists(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	analysis, err := service.GetAnalysisByAnalyzer(context.Background(), testProjectID, "test-analyzer")
	require.NoError(t, err)
	assert.Nil(t, analysis)
}

func TestAnalysisService_MultipleAnalyzersPerProject(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	testProjectID := uuid.New()
	mockStore.projects[testProjectID] = true

	logger := zaptest.NewLogger(t)
	service := NewAnalysisService(mockStore, logger)

	input1 := AnalysisInput{
		Summary:         "Analysis 1",
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "analyzer-1",
	}

	input2 := AnalysisInput{
		Summary:         "Analysis 2",
		Purpose:         "Test purpose",
		Architecture:    "Test architecture",
		AnalyzedAt:      time.Now().UTC(),
		AnalyzedGitHead: "abc123",
		Analyzer:        "analyzer-2",
	}

	analysis1, err := service.StoreAnalysis(context.Background(), testProjectID, input1)
	require.NoError(t, err)
	assert.NotNil(t, analysis1)

	analysis2, err := service.StoreAnalysis(context.Background(), testProjectID, input2)
	require.NoError(t, err)
	assert.NotNil(t, analysis2)

	assert.Equal(t, "analyzer-1", analysis1.Analyzer)
	assert.Equal(t, "analyzer-2", analysis2.Analyzer)
	assert.NotEqual(t, analysis1.ID, analysis2.ID)
}
