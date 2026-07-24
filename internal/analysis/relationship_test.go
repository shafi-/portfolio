package analysis

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestRelationshipService_StoreRelationship_Success(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID] = true

	confidence := 0.9
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationship, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Both are todo apps", &confidence)
	require.NoError(t, err)
	assert.NotNil(t, relationship)
	assert.Equal(t, "Similar", relationship.Type)
	assert.Equal(t, sourceProjectID.String(), relationship.SourceProject)
	assert.Equal(t, targetProjectID.String(), relationship.TargetProject)
	assert.Equal(t, "Both are todo apps", relationship.Description)
	assert.Equal(t, &confidence, relationship.Confidence)
}

func TestRelationshipService_StoreRelationship_UpdateExisting(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID] = true

	// Create existing relationship
	existingRelationship := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProjectID.String(),
		TargetProject: targetProjectID.String(),
		Type:          "Similar",
		Description:   "Old description",
		Confidence:    floatPtr(0.9),
	}
	mockStore.relationships[existingRelationship.ID] = existingRelationship

	newConfidence := 0.8
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationship, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "New description", &newConfidence)
	require.NoError(t, err)
	assert.NotNil(t, relationship)
	assert.Equal(t, "New description", relationship.Description)
	assert.Equal(t, &newConfidence, relationship.Confidence)
}

func TestRelationshipService_StoreRelationship_InvalidType(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID] = true

	confidence := 0.9
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	_, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "InvalidType", "Test description", &confidence)
	assert.Error(t, err)
}

func TestRelationshipService_StoreRelationship_InvalidConfidence(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID] = true

	invalidConfidence := 1.5
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	_, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Test description", &invalidConfidence)
	assert.Error(t, err)
}

func TestRelationshipService_StoreRelationship_NonExistentSourceProject(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[targetProjectID] = true // Only target exists

	confidence := 0.9
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	_, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Test description", &confidence)
	assert.Error(t, err)
}

func TestRelationshipService_StoreRelationship_NonExistentTargetProject(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true // Only source exists

	confidence := 0.9
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	_, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Test description", &confidence)
	assert.Error(t, err)
}

func TestRelationshipService_ListRelationships_AsSource(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID1 := uuid.New()
	targetProjectID2 := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID1] = true
	mockStore.projects[targetProjectID2] = true

	// Create relationships
	rel1 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProjectID.String(),
		TargetProject: targetProjectID1.String(),
		Type:          "Similar",
	}
	rel2 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProjectID.String(),
		TargetProject: targetProjectID2.String(),
		Type:          "Evolution",
	}
	mockStore.relationships[rel1.ID] = rel1
	mockStore.relationships[rel2.ID] = rel2

	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationships, err := service.ListRelationships(context.Background(), sourceProjectID)
	require.NoError(t, err)
	assert.Len(t, relationships, 2)
}

func TestRelationshipService_ListRelationships_AsTarget(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID1 := uuid.New()
	sourceProjectID2 := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID1] = true
	mockStore.projects[sourceProjectID2] = true
	mockStore.projects[targetProjectID] = true

	// Create relationships where target is the project
	rel1 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProjectID1.String(),
		TargetProject: targetProjectID.String(),
		Type:          "Similar",
	}
	rel2 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProjectID2.String(),
		TargetProject: targetProjectID.String(),
		Type:          "Evolution",
	}
	mockStore.relationships[rel1.ID] = rel1
	mockStore.relationships[rel2.ID] = rel2

	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationships, err := service.ListRelationships(context.Background(), targetProjectID)
	require.NoError(t, err)
	assert.Len(t, relationships, 2)
}

func TestRelationshipService_ListRelationships_BothDirections(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	projectA := uuid.New()
	projectB := uuid.New()
	projectC := uuid.New()
	mockStore.projects[projectA] = true
	mockStore.projects[projectB] = true
	mockStore.projects[projectC] = true

	// Create relationships in both directions
	rel1 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: projectA.String(),
		TargetProject: projectB.String(),
		Type:          "Similar",
	}
	rel2 := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: projectC.String(),
		TargetProject: projectA.String(),
		Type:          "Evolution",
	}
	mockStore.relationships[rel1.ID] = rel1
	mockStore.relationships[rel2.ID] = rel2

	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationships, err := service.ListRelationships(context.Background(), projectA)
	require.NoError(t, err)
	assert.Len(t, relationships, 2)
}

func TestRelationshipService_ListRelationships_Empty(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	projectID := uuid.New()
	mockStore.projects[projectID] = true

	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	relationships, err := service.ListRelationships(context.Background(), projectID)
	require.NoError(t, err)
	assert.Len(t, relationships, 0)
}

func TestRelationshipService_DeleteRelationship(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	projectID := uuid.New()
	mockStore.projects[projectID] = true

	// Create relationship
	relationshipID := uuid.New()
	relationship := &Relationship{
		ID:            relationshipID.String(),
		SourceProject: projectID.String(),
		TargetProject: uuid.New().String(),
		Type:          "Similar",
	}
	mockStore.relationships[relationshipID.String()] = relationship

	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	err := service.DeleteRelationship(context.Background(), relationshipID)
	require.NoError(t, err)

	// Verify relationship is deleted
	_, exists := mockStore.relationships[relationshipID.String()]
	assert.False(t, exists)
}

func TestRelationshipService_DeleteRelationship_NonExistent(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	projectID := uuid.New()
	mockStore.projects[projectID] = true

	nonExistentID := uuid.New()
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	err := service.DeleteRelationship(context.Background(), nonExistentID)
	// Should not error, just be a no-op
	require.NoError(t, err)
}

func TestRelationshipService_ValidateRelationshipType_AllowedTypes(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	allowedTypes := []string{"Similar", "Evolution", "Shared Feature", "Shared Technology", "Reuses Component"}
	for _, relType := range allowedTypes {
		err := service.ValidateRelationshipType(relType)
		assert.NoError(t, err, "Type should be allowed: %s", relType)
	}
}

func TestRelationshipService_ValidateConfidence_ValidRange(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	// Valid values
	validValues := []float64{0.0, 0.5, 1.0}
	for _, val := range validValues {
		confidence := val
		err := service.ValidateConfidence(&confidence)
		assert.NoError(t, err, "Confidence should be valid: %f", val)
	}

	// Valid nil value
	err := service.ValidateConfidence(nil)
	assert.NoError(t, err, "Nil confidence should be valid")
}

func TestRelationshipService_ValidateConfidence_InvalidRange(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	// Invalid values
	invalidValues := []float64{-0.5, -1.0, 1.5, 2.0}
	for _, val := range invalidValues {
		confidence := val
		err := service.ValidateConfidence(&confidence)
		assert.Error(t, err, "Confidence should be invalid: %f", val)
	}
}

func TestRelationshipService_DuplicateHandling(t *testing.T) {
	mockStore := NewMockAnalysisStore()
	sourceProjectID := uuid.New()
	targetProjectID := uuid.New()
	mockStore.projects[sourceProjectID] = true
	mockStore.projects[targetProjectID] = true

	confidence := 0.9
	logger := zaptest.NewLogger(t)
	service := NewRelationshipService(mockStore, logger)

	// Store first relationship
	relationship1, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Description 1", &confidence)
	require.NoError(t, err)
	assert.NotNil(t, relationship1)

	// Store duplicate relationship (same source, target, type)
	newConfidence := 0.8
	relationship2, err := service.StoreRelationship(context.Background(), sourceProjectID, targetProjectID, "Similar", "Description 2", &newConfidence)
	require.NoError(t, err)
	assert.NotNil(t, relationship2)

	// Should be the same relationship (updated)
	assert.Equal(t, relationship1.ID, relationship2.ID)
	assert.Equal(t, "Description 2", relationship2.Description)
}
