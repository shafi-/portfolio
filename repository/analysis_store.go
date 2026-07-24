package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"project-dash/internal/analysis"
)

// SQLiteAnalysisStore implements AnalysisStore for SQLite
type SQLiteAnalysisStore struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLiteAnalysisStore creates a new SQLite analysis store
func NewSQLiteAnalysisStore(db *sql.DB, logger *zap.Logger) *SQLiteAnalysisStore {
	return &SQLiteAnalysisStore{
		db:     db,
		logger: logger,
	}
}

// ProjectExists checks if a project exists in the database
func (s *SQLiteAnalysisStore) ProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, queryProjectExists, projectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}
	return exists, nil
}

// CreateAnalysis creates a new analysis record
func (s *SQLiteAnalysisStore) CreateAnalysis(ctx context.Context, analysis *analysis.Analysis) error {
	strengthsJSON, err := json.Marshal(analysis.Strengths)
	if err != nil {
		return fmt.Errorf("failed to marshal strengths: %w", err)
	}

	weaknessesJSON, err := json.Marshal(analysis.Weaknesses)
	if err != nil {
		return fmt.Errorf("failed to marshal weaknesses: %w", err)
	}

	reusableJSON, err := json.Marshal(analysis.ReusableComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal reusable_components: %w", err)
	}

	_, err = s.db.ExecContext(ctx, queryCreateAnalysis,
		analysis.ID,
		analysis.ProjectID,
		analysis.Analyzer,
		analysis.AnalyzedGitHead,
		analysis.AnalyzedAt.Format(time.RFC3339),
		analysis.Summary,
		analysis.Purpose,
		analysis.Architecture,
		analysis.Maturity,
		string(strengthsJSON),
		string(weaknessesJSON),
		string(reusableJSON),
		analysis.Notes,
		string(analysis.RawJSON),
		analysis.CreatedAt.Format(time.RFC3339),
		analysis.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to insert analysis: %w", err)
	}

	s.logger.Debug("analysis created",
		zap.String("id", analysis.ID),
		zap.String("project_id", analysis.ProjectID),
	)

	return nil
}

// UpdateAnalysis updates an existing analysis record
func (s *SQLiteAnalysisStore) UpdateAnalysis(ctx context.Context, analysis *analysis.Analysis) error {
	strengthsJSON, err := json.Marshal(analysis.Strengths)
	if err != nil {
		return fmt.Errorf("failed to marshal strengths: %w", err)
	}

	weaknessesJSON, err := json.Marshal(analysis.Weaknesses)
	if err != nil {
		return fmt.Errorf("failed to marshal weaknesses: %w", err)
	}

	reusableJSON, err := json.Marshal(analysis.ReusableComponents)
	if err != nil {
		return fmt.Errorf("failed to marshal reusable_components: %w", err)
	}

	result, err := s.db.ExecContext(ctx, queryUpdateAnalysis,
		analysis.Summary,
		analysis.Purpose,
		analysis.Architecture,
		analysis.Maturity,
		string(strengthsJSON),
		string(weaknessesJSON),
		string(reusableJSON),
		analysis.Notes,
		string(analysis.RawJSON),
		analysis.AnalyzedGitHead,
		analysis.AnalyzedAt.Format(time.RFC3339),
		analysis.UpdatedAt.Format(time.RFC3339),
		analysis.ProjectID,
		analysis.Analyzer,
	)

	if err != nil {
		return fmt.Errorf("failed to update analysis: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	s.logger.Debug("analysis updated",
		zap.String("id", analysis.ID),
		zap.String("project_id", analysis.ProjectID),
		zap.String("analyzer", analysis.Analyzer),
	)

	return nil
}

// GetAnalysis retrieves the most recent analysis for a project
func (s *SQLiteAnalysisStore) GetAnalysis(ctx context.Context, projectID uuid.UUID) (*analysis.Analysis, error) {
	row := s.db.QueryRowContext(ctx, queryGetAnalysis, projectID)

	var analysis analysis.Analysis
	var strengthsJSON, weaknessesJSON, reusableJSON string

	err := row.Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&analysis.Analyzer,
		&analysis.AnalyzedGitHead,
		&analysis.AnalyzedAt,
		&analysis.Summary,
		&analysis.Purpose,
		&analysis.Architecture,
		&analysis.Maturity,
		&strengthsJSON,
		&weaknessesJSON,
		&reusableJSON,
		&analysis.Notes,
		&analysis.RawJSON,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan analysis: %w", err)
	}

	// Unmarshal JSON arrays
	if err := json.Unmarshal([]byte(strengthsJSON), &analysis.Strengths); err != nil {
		return nil, fmt.Errorf("failed to unmarshal strengths: %w", err)
	}

	if err := json.Unmarshal([]byte(weaknessesJSON), &analysis.Weaknesses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weaknesses: %w", err)
	}

	if err := json.Unmarshal([]byte(reusableJSON), &analysis.ReusableComponents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reusable_components: %w", err)
	}

	// Load features
	features, err := s.GetFeaturesByAnalysisID(ctx, uuid.MustParse(analysis.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to load features: %w", err)
	}

	analysis.Features = features

	return &analysis, nil
}

// GetAnalysisByAnalyzer retrieves an analysis for a project by analyzer
func (s *SQLiteAnalysisStore) GetAnalysisByAnalyzer(ctx context.Context, projectID uuid.UUID, analyzer string) (*analysis.Analysis, error) {
	row := s.db.QueryRowContext(ctx, queryGetAnalysisByAnalyzer, projectID, analyzer)

	var analysis analysis.Analysis
	var strengthsJSON, weaknessesJSON, reusableJSON string

	err := row.Scan(
		&analysis.ID,
		&analysis.ProjectID,
		&analysis.Analyzer,
		&analysis.AnalyzedGitHead,
		&analysis.AnalyzedAt,
		&analysis.Summary,
		&analysis.Purpose,
		&analysis.Architecture,
		&analysis.Maturity,
		&strengthsJSON,
		&weaknessesJSON,
		&reusableJSON,
		&analysis.Notes,
		&analysis.RawJSON,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan analysis: %w", err)
	}

	// Unmarshal JSON arrays
	if err := json.Unmarshal([]byte(strengthsJSON), &analysis.Strengths); err != nil {
		return nil, fmt.Errorf("failed to unmarshal strengths: %w", err)
	}

	if err := json.Unmarshal([]byte(weaknessesJSON), &analysis.Weaknesses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weaknesses: %w", err)
	}

	if err := json.Unmarshal([]byte(reusableJSON), &analysis.ReusableComponents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reusable_components: %w", err)
	}

	// Load features
	features, err := s.GetFeaturesByAnalysisID(ctx, uuid.MustParse(analysis.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to load features: %w", err)
	}

	analysis.Features = features

	return &analysis, nil
}

// DeleteAnalysis deletes an analysis for a project by analyzer
func (s *SQLiteAnalysisStore) DeleteAnalysis(ctx context.Context, projectID uuid.UUID, analyzer string) error {
	_, err := s.db.ExecContext(ctx, queryDeleteAnalysis, projectID, analyzer)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}

	s.logger.Debug("analysis deleted",
		zap.String("project_id", projectID.String()),
		zap.String("analyzer", analyzer),
	)

	return nil
}

// CreateFeatures creates multiple feature records
func (s *SQLiteAnalysisStore) CreateFeatures(ctx context.Context, features []analysis.Feature) error {
	if len(features) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, queryCreateFeatures)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, feature := range features {
		_, err := stmt.ExecContext(ctx,
			feature.ID,
			feature.AnalysisID,
			feature.Name,
			feature.Description,
			feature.Confidence,
			feature.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to insert feature: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("features created",
		zap.Int("count", len(features)),
		zap.String("analysis_id", features[0].AnalysisID),
	)

	return nil
}

// DeleteFeaturesByAnalysisID deletes all features for an analysis
func (s *SQLiteAnalysisStore) DeleteFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, queryDeleteFeaturesByAnalysisID, analysisID)
	if err != nil {
		return fmt.Errorf("failed to delete features: %w", err)
	}

	s.logger.Debug("features deleted",
		zap.String("analysis_id", analysisID.String()),
	)

	return nil
}

// GetFeaturesByAnalysisID retrieves all features for an analysis
func (s *SQLiteAnalysisStore) GetFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]analysis.Feature, error) {
	rows, err := s.db.QueryContext(ctx, queryGetFeaturesByAnalysisID, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to query features: %w", err)
	}
	defer rows.Close()

	var features []analysis.Feature
	for rows.Next() {
		var feature analysis.Feature
		if err := rows.Scan(
			&feature.ID,
			&feature.AnalysisID,
			&feature.Name,
			&feature.Description,
			&feature.Confidence,
			&feature.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature: %w", err)
		}
		features = append(features, feature)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating features: %w", err)
	}

	return features, nil
}

// GetGitHeadForProject retrieves the current git HEAD for a project
func (s *SQLiteAnalysisStore) GetGitHeadForProject(ctx context.Context, projectID uuid.UUID) (*string, error) {
	var gitHead sql.NullString
	err := s.db.QueryRowContext(ctx, queryGetGitHeadForProject, projectID).Scan(&gitHead)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get git head: %w", err)
	}

	if !gitHead.Valid {
		return nil, nil
	}

	return &gitHead.String, nil
}

// ListAllAnalyses retrieves all analyses
func (s *SQLiteAnalysisStore) ListAllAnalyses(ctx context.Context) ([]analysis.Analysis, error) {
	rows, err := s.db.QueryContext(ctx, queryListAllAnalyses)
	if err != nil {
		return nil, fmt.Errorf("failed to query analyses: %w", err)
	}
	defer rows.Close()

	var analyses []analysis.Analysis
	for rows.Next() {
		var a analysis.Analysis
		var strengthsJSON, weaknessesJSON, reusableJSON string

		if err := rows.Scan(
			&a.ID,
			&a.ProjectID,
			&a.Analyzer,
			&a.AnalyzedGitHead,
			&a.AnalyzedAt,
			&a.Summary,
			&a.Purpose,
			&a.Architecture,
			&a.Maturity,
			&strengthsJSON,
			&weaknessesJSON,
			&reusableJSON,
			&a.Notes,
			&a.RawJSON,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}

		// Unmarshal JSON arrays
		if err := json.Unmarshal([]byte(strengthsJSON), &a.Strengths); err != nil {
			return nil, fmt.Errorf("failed to unmarshal strengths: %w", err)
		}

		if err := json.Unmarshal([]byte(weaknessesJSON), &a.Weaknesses); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weaknesses: %w", err)
		}

		if err := json.Unmarshal([]byte(reusableJSON), &a.ReusableComponents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reusable_components: %w", err)
		}

		analyses = append(analyses, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating analyses: %w", err)
	}

	return analyses, nil
}

// CreateRelationship creates a new relationship
func (s *SQLiteAnalysisStore) CreateRelationship(ctx context.Context, rel *analysis.Relationship) error {
	_, err := s.db.ExecContext(ctx, queryCreateRelationship,
		rel.ID,
		rel.SourceProject,
		rel.TargetProject,
		rel.Type,
		rel.Description,
		rel.Confidence,
		rel.CreatedAt.Format(time.RFC3339),
		rel.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	s.logger.Debug("relationship created",
		zap.String("id", rel.ID),
		zap.String("source_project", rel.SourceProject),
		zap.String("target_project", rel.TargetProject),
	)

	return nil
}

// UpdateRelationship updates an existing relationship
func (s *SQLiteAnalysisStore) UpdateRelationship(ctx context.Context, rel *analysis.Relationship) error {
	result, err := s.db.ExecContext(ctx, queryUpdateRelationship,
		rel.Description,
		rel.Confidence,
		rel.UpdatedAt.Format(time.RFC3339),
		rel.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	s.logger.Debug("relationship updated",
		zap.String("id", rel.ID),
	)

	return nil
}

// GetRelationship retrieves a relationship by ID
func (s *SQLiteAnalysisStore) GetRelationship(ctx context.Context, id uuid.UUID) (*analysis.Relationship, error) {
	row := s.db.QueryRowContext(ctx, queryGetRelationship, id)

	var rel analysis.Relationship
	err := row.Scan(
		&rel.ID,
		&rel.SourceProject,
		&rel.TargetProject,
		&rel.Type,
		&rel.Description,
		&rel.Confidence,
		&rel.CreatedAt,
		&rel.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan relationship: %w", err)
	}

	return &rel, nil
}

// ListRelationshipsByProject retrieves all relationships for a project
func (s *SQLiteAnalysisStore) ListRelationshipsByProject(ctx context.Context, projectID uuid.UUID) ([]analysis.Relationship, error) {
	rows, err := s.db.QueryContext(ctx, queryListRelationshipsByProject, projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer rows.Close()

	var relationships []analysis.Relationship
	for rows.Next() {
		var rel analysis.Relationship
		if err := rows.Scan(
			&rel.ID,
			&rel.SourceProject,
			&rel.TargetProject,
			&rel.Type,
			&rel.Description,
			&rel.Confidence,
			&rel.CreatedAt,
			&rel.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relationships: %w", err)
	}

	return relationships, nil
}

// DeleteRelationship deletes a relationship
func (s *SQLiteAnalysisStore) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, queryDeleteRelationship, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	s.logger.Debug("relationship deleted",
		zap.String("id", id.String()),
	)

	return nil
}

// FindExistingRelationship finds an existing relationship by source, target, and type
func (s *SQLiteAnalysisStore) FindExistingRelationship(ctx context.Context, source, target uuid.UUID, relType string) (*analysis.Relationship, error) {
	row := s.db.QueryRowContext(ctx, queryFindExistingRelationship, source, target, relType)

	var rel analysis.Relationship
	err := row.Scan(
		&rel.ID,
		&rel.SourceProject,
		&rel.TargetProject,
		&rel.Type,
		&rel.Description,
		&rel.Confidence,
		&rel.CreatedAt,
		&rel.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan relationship: %w", err)
	}

	return &rel, nil
}