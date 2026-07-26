package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type FeatureStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewFeatureStore(db *sql.DB, logger *zap.Logger) *FeatureStore {
	return &FeatureStore{db: db, logger: logger}
}

// featureColumns is f.-qualified because some readers JOIN analyses (which also
// has id/name) — an unqualified projection is ambiguous there. Every SELECT
// below aliases the features table as f.
const featureColumns = "f.id, f.analysis_id, f.name, f.description, f.confidence, f.implementation_status, f.feature_architecture, f.pattern"

func (s *FeatureStore) CreateFeature(f *models.Feature) error {
	return s.create(s.db, f)
}

func (s *FeatureStore) CreateFeatureTx(tx *sql.Tx, f *models.Feature) error {
	return s.create(tx, f)
}

func (s *FeatureStore) create(q Querier, f *models.Feature) error {
	query := `
		INSERT INTO features (id, analysis_id, name, description, confidence, implementation_status, feature_architecture, pattern)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(context.Background(), query,
		f.ID, f.AnalysisID, f.Name, f.Description, f.Confidence,
		f.ImplementationStatus, f.FeatureArchitecture, f.Pattern,
	)
	if err != nil {
		return fmt.Errorf("failed to create feature: %w", err)
	}
	return nil
}

func (s *FeatureStore) GetFeature(id string) (*models.Feature, error) {
	query := `SELECT ` + featureColumns + ` FROM features f WHERE f.id = ?`
	f := &models.Feature{}
	err := s.db.QueryRow(query, id).Scan(
		&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence,
		&f.ImplementationStatus, &f.FeatureArchitecture, &f.Pattern,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}
	return f, nil
}

// GetByAnalysisAndName returns the feature for the given analysis + name, or
// (nil, nil) when none exists. It backs the storeFeature upsert: an agent
// deep-diving a feature calls storeFeature again with the same project/analyzer/
// name, and this lookup decides update-vs-create. There is intentionally no
// UNIQUE(analysis_id, name) constraint yet, so pre-existing duplicates collapse
// to the first row (LIMIT 1); a constraint + backfill is a future hardening.
func (s *FeatureStore) GetByAnalysisAndName(analysisID, name string) (*models.Feature, error) {
	query := `SELECT ` + featureColumns + ` FROM features f WHERE f.analysis_id = ? AND f.name = ? LIMIT 1`
	f := &models.Feature{}
	err := s.db.QueryRow(query, analysisID, name).Scan(
		&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence,
		&f.ImplementationStatus, &f.FeatureArchitecture, &f.Pattern,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feature by name: %w", err)
	}
	return f, nil
}

// UpdateFeature overwrites every column of the feature identified by f.ID. The
// handler is responsible for merge semantics — it reads the existing row,
// overlays only the fields the caller supplied, and writes the merged struct
// back — so an empty Tier-3 call never blanks stored Tier-2 facts.
func (s *FeatureStore) UpdateFeature(f *models.Feature) error {
	query := `
		UPDATE features
		SET analysis_id = ?, name = ?, description = ?, confidence = ?,
		    implementation_status = ?, feature_architecture = ?, pattern = ?
		WHERE id = ?`
	_, err := s.db.ExecContext(context.Background(), query,
		f.AnalysisID, f.Name, f.Description, f.Confidence,
		f.ImplementationStatus, f.FeatureArchitecture, f.Pattern,
		f.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update feature: %w", err)
	}
	return nil
}

func (s *FeatureStore) ListByAnalysis(analysisID string) ([]*models.Feature, error) {
	query := `SELECT ` + featureColumns + ` FROM features f WHERE f.analysis_id = ? ORDER BY f.name`
	rows, err := s.db.Query(query, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}
	defer rows.Close()

	var features []*models.Feature
	for rows.Next() {
		f := &models.Feature{}
		if err := rows.Scan(
			&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence,
			&f.ImplementationStatus, &f.FeatureArchitecture, &f.Pattern,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature row: %w", err)
		}
		features = append(features, f)
	}
	return features, rows.Err()
}

func (s *FeatureStore) ListByProject(projectID string) ([]*models.Feature, error) {
	query := `SELECT ` + featureColumns + ` FROM features f
		JOIN analyses a ON a.id = f.analysis_id
		WHERE a.project_id = ?
		ORDER BY f.name`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project features: %w", err)
	}
	defer rows.Close()

	var features []*models.Feature
	for rows.Next() {
		f := &models.Feature{}
		if err := rows.Scan(
			&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence,
			&f.ImplementationStatus, &f.FeatureArchitecture, &f.Pattern,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature row: %w", err)
		}
		features = append(features, f)
	}
	return features, rows.Err()
}

func (s *FeatureStore) SearchFeatures(opts FeatureSearchOptions) ([]*models.Feature, error) {
	// Conditions are built with plain positional `?` placeholders (the style used
	// across every other store); args are appended in the same order, so the
	// driver binds them left-to-right. No numbered placeholders needed.
	var conditions []string
	var args []interface{}

	if opts.ProjectID != "" {
		conditions = append(conditions, "a.project_id = ?")
		args = append(args, opts.ProjectID)
	}
	if opts.ImplementationStatus != "" {
		conditions = append(conditions, "f.implementation_status = ?")
		args = append(args, opts.ImplementationStatus)
	}
	if opts.Pattern != "" {
		conditions = append(conditions, "f.pattern LIKE ?")
		args = append(args, "%"+opts.Pattern+"%")
	}
	if opts.Query != "" {
		q := "%" + opts.Query + "%"
		conditions = append(conditions, "(f.name LIKE ? OR f.description LIKE ? OR f.feature_architecture LIKE ?)")
		args = append(args, q, q, q)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT ` + featureColumns + ` FROM features f
		JOIN analyses a ON a.id = f.analysis_id` + where + ` ORDER BY f.name`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search features: %w", err)
	}
	defer rows.Close()

	var features []*models.Feature
	for rows.Next() {
		f := &models.Feature{}
		if err := rows.Scan(
			&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence,
			&f.ImplementationStatus, &f.FeatureArchitecture, &f.Pattern,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature row: %w", err)
		}
		features = append(features, f)
	}
	return features, rows.Err()
}

func (s *FeatureStore) DeleteAllForAnalysis(analysisID string) error {
	_, err := s.db.Exec("DELETE FROM features WHERE analysis_id = ?", analysisID)
	if err != nil {
		return fmt.Errorf("failed to delete features for analysis: %w", err)
	}
	return nil
}

func (s *FeatureStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec(`DELETE FROM features WHERE analysis_id IN (SELECT id FROM analyses WHERE project_id = ?)`, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete features for project: %w", err)
	}
	return nil
}

type FeatureSearchOptions struct {
	ProjectID            string
	ImplementationStatus string
	Pattern              string
	Query                string
}
