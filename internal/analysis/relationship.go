package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RelationshipService manages relationship operations
type RelationshipService struct {
	store  AnalysisStore
	logger *zap.Logger
}

// NewRelationshipService creates a new relationship service
func NewRelationshipService(store AnalysisStore, logger *zap.Logger) *RelationshipService {
	return &RelationshipService{
		store:  store,
		logger: logger,
	}
}

// allowedRelationshipTypes defines the valid relationship types
var allowedRelationshipTypes = map[string]bool{
	"Similar":           true,
	"Evolution":         true,
	"Shared Feature":    true,
	"Shared Technology": true,
	"Reuses Component":  true,
}

// ValidateRelationshipType validates a relationship type
func (s *RelationshipService) ValidateRelationshipType(relType string) error {
	if !allowedRelationshipTypes[relType] {
		return fmt.Errorf("%w: '%s'. Allowed types: Similar, Evolution, Shared Feature, Shared Technology, Reuses Component", ErrInvalidRelationType, relType)
	}
	return nil
}

// ValidateConfidence validates a confidence value
func (s *RelationshipService) ValidateConfidence(confidence *float64) error {
	if confidence != nil && (*confidence < 0.0 || *confidence > 1.0) {
		return fmt.Errorf("%w: confidence must be in range [0.0, 1.0], got %f", ErrInvalidRelationType, *confidence)
	}
	return nil
}

// StoreRelationship stores or updates a relationship
func (s *RelationshipService) StoreRelationship(ctx context.Context, source, target uuid.UUID, relType, description string, confidence *float64) (*Relationship, error) {
	logger := s.logger.With(
		zap.String("source_project", source.String()),
		zap.String("target_project", target.String()),
		zap.String("type", relType),
	)
	logger.Info("storing relationship")

	// Step 1: Validate relationship type
	if err := s.ValidateRelationshipType(relType); err != nil {
		return nil, err
	}

	// Step 2: Validate confidence
	if err := s.ValidateConfidence(confidence); err != nil {
		return nil, err
	}

	// Step 3: Validate project existence
	sourceExists, err := s.store.ProjectExists(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to check source project existence: %w", err)
	}
	if !sourceExists {
		return nil, ErrProjectNotFound
	}

	targetExists, err := s.store.ProjectExists(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("failed to check target project existence: %w", err)
	}
	if !targetExists {
		return nil, ErrProjectNotFound
	}

	// Step 4: Check for existing relationship
	existing, err := s.store.FindExistingRelationship(ctx, source, target, relType)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing relationship: %w", err)
	}

	now := time.Now().UTC()

	if existing != nil {
		// Update existing relationship
		existing.Description = description
		existing.Confidence = confidence
		existing.UpdatedAt = now

		if err := s.store.UpdateRelationship(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update relationship: %w", err)
		}

		logger.Info("relationship updated", zap.String("relationship_id", existing.ID))
		return existing, nil
	}

	// Step 5: Create new relationship
	relationship := &Relationship{
		ID:            uuid.New().String(),
		SourceProject: source.String(),
		TargetProject: target.String(),
		Type:          relType,
		Description:   description,
		Confidence:    confidence,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.store.CreateRelationship(ctx, relationship); err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	logger.Info("relationship created", zap.String("relationship_id", relationship.ID))
	return relationship, nil
}

// ListRelationships lists all relationships for a project
func (s *RelationshipService) ListRelationships(ctx context.Context, projectID uuid.UUID) ([]Relationship, error) {
	logger := s.logger.With(zap.String("project_id", projectID.String()))
	logger.Debug("listing relationships")

	relationships, err := s.store.ListRelationshipsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}

	logger.Debug("relationships listed", zap.Int("count", len(relationships)))
	return relationships, nil
}

// DeleteRelationship deletes a relationship
func (s *RelationshipService) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	logger := s.logger.With(zap.String("relationship_id", relationshipID.String()))
	logger.Info("deleting relationship")

	if err := s.store.DeleteRelationship(ctx, relationshipID); err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	logger.Info("relationship deleted")
	return nil
}
