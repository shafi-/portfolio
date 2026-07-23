package store

import (
	"context"
	"database/sql"
	"fmt"

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

func (s *FeatureStore) CreateFeature(f *models.Feature) error {
	return s.create(s.db, f)
}

func (s *FeatureStore) CreateFeatureTx(tx *sql.Tx, f *models.Feature) error {
	return s.create(tx, f)
}

func (s *FeatureStore) create(q Querier, f *models.Feature) error {
	query := `
		INSERT INTO features (id, analysis_id, name, description, confidence)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(context.Background(), query,
		f.ID, f.AnalysisID, f.Name, f.Description, f.Confidence,
	)
	if err != nil {
		return fmt.Errorf("failed to create feature: %w", err)
	}
	return nil
}

func (s *FeatureStore) GetFeature(id string) (*models.Feature, error) {
	query := `SELECT id, analysis_id, name, description, confidence FROM features WHERE id = ?`
	f := &models.Feature{}
	err := s.db.QueryRow(query, id).Scan(&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}
	return f, nil
}

func (s *FeatureStore) ListByAnalysis(analysisID string) ([]*models.Feature, error) {
	query := `SELECT id, analysis_id, name, description, confidence FROM features WHERE analysis_id = ? ORDER BY name`
	rows, err := s.db.Query(query, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to list features: %w", err)
	}
	defer rows.Close()

	var features []*models.Feature
	for rows.Next() {
		f := &models.Feature{}
		if err := rows.Scan(&f.ID, &f.AnalysisID, &f.Name, &f.Description, &f.Confidence); err != nil {
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
