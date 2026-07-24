package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AnalysisService orchestrates analysis operations
type AnalysisService struct {
	store           AnalysisStore
	schemaValidator *SchemaValidator
	staleDetector   *StaleDetector
	logger          *zap.Logger
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(store AnalysisStore, logger *zap.Logger) *AnalysisService {
	validator, err := NewSchemaValidator()
	if err != nil {
		logger.Fatal("failed to create schema validator", zap.Error(err))
	}

	return &AnalysisService{
		store:           store,
		schemaValidator: validator,
		staleDetector:   NewStaleDetector(store),
		logger:          logger,
	}
}

// StoreAnalysis stores or updates an analysis for a project
func (s *AnalysisService) StoreAnalysis(ctx context.Context, projectID uuid.UUID, input AnalysisInput) (*Analysis, error) {
	logger := s.logger.With(zap.String("project_id", projectID.String()))
	logger.Info("storing analysis", zap.String("analyzer", input.Analyzer))

	// Step 1: Validate project exists
	exists, err := s.store.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if !exists {
		return nil, ErrProjectNotFound
	}

	// Step 2: Validate schema
	if err := s.schemaValidator.Validate(input); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRelationType, err)
	}

	// Step 3: Serialize to raw_json
	rawJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis: %w", err)
	}

	// Step 4: Check for existing analysis
	existing, err := s.store.GetAnalysisByAnalyzer(ctx, projectID, input.Analyzer)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing analysis: %w", err)
	}

	// Step 5: Prepare analysis object
	now := time.Now().UTC()
	analysis := &Analysis{
		ID:                 uuid.New().String(),
		ProjectID:          projectID.String(),
		Analyzer:           input.Analyzer,
		AnalyzedGitHead:    input.AnalyzedGitHead,
		AnalyzedAt:         input.AnalyzedAt,
		Summary:            input.Summary,
		Purpose:            input.Purpose,
		Architecture:       input.Architecture,
		Maturity:           input.Maturity,
		Strengths:          input.Strengths,
		Weaknesses:         input.Weaknesses,
		ReusableComponents: input.ReusableComponents,
		Notes:              input.Notes,
		RawJSON:            rawJSON,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Step 6: Transaction
	if existing != nil {
		analysis.ID = existing.ID
		analysis.CreatedAt = existing.CreatedAt

		// Update existing analysis
		if err := s.store.UpdateAnalysis(ctx, analysis); err != nil {
			return nil, fmt.Errorf("failed to update analysis: %w", err)
		}

		// Delete old features
		if err := s.store.DeleteFeaturesByAnalysisID(ctx, uuid.MustParse(existing.ID)); err != nil {
			return nil, fmt.Errorf("failed to delete old features: %w", err)
		}
	} else {
		// Create new analysis
		if err := s.store.CreateAnalysis(ctx, analysis); err != nil {
			return nil, fmt.Errorf("failed to create analysis: %w", err)
		}
	}

	// Step 7: Insert features
	if len(input.Features) > 0 {
		features := make([]Feature, len(input.Features))
		for i, f := range input.Features {
			features[i] = Feature{
				ID:          uuid.New().String(),
				AnalysisID:  analysis.ID,
				Name:        f.Name,
				Description: f.Description,
				Confidence:  f.Confidence,
				CreatedAt:   now,
			}
		}
		if err := s.store.CreateFeatures(ctx, features); err != nil {
			return nil, fmt.Errorf("failed to create features: %w", err)
		}
		analysis.Features = features
	}

	logger.Info("analysis stored successfully", zap.String("analysis_id", analysis.ID))
	return analysis, nil
}

// GetAnalysis retrieves the most recent analysis for a project
func (s *AnalysisService) GetAnalysis(ctx context.Context, projectID uuid.UUID) (*Analysis, error) {
	logger := s.logger.With(zap.String("project_id", projectID.String()))
	logger.Debug("getting analysis")

	analysis, err := s.store.GetAnalysis(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	if analysis == nil {
		return nil, nil
	}

	logger.Debug("analysis retrieved", zap.String("analysis_id", analysis.ID))
	return analysis, nil
}

// GetAnalysisByAnalyzer retrieves an analysis for a project by analyzer
func (s *AnalysisService) GetAnalysisByAnalyzer(ctx context.Context, projectID uuid.UUID, analyzer string) (*Analysis, error) {
	logger := s.logger.With(
		zap.String("project_id", projectID.String()),
		zap.String("analyzer", analyzer),
	)
	logger.Debug("getting analysis by analyzer")

	analysis, err := s.store.GetAnalysisByAnalyzer(ctx, projectID, analyzer)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis by analyzer: %w", err)
	}

	if analysis == nil {
		return nil, nil
	}

	logger.Debug("analysis retrieved", zap.String("analysis_id", analysis.ID))
	return analysis, nil
}

// ListProjectsNeedingAnalysis lists all projects that need analysis
func (s *AnalysisService) ListProjectsNeedingAnalysis(ctx context.Context) ([]NeedsAnalysisResult, error) {
	s.logger.Debug("listing projects needing analysis")

	results, err := s.staleDetector.ListNeedingAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects needing analysis: %w", err)
	}

	s.logger.Debug("projects needing analysis", zap.Int("count", len(results)))
	return results, nil
}
