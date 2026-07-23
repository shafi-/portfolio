package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type AnalysisStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewAnalysisStore(db *sql.DB, logger *zap.Logger) *AnalysisStore {
	return &AnalysisStore{db: db, logger: logger}
}

func (s *AnalysisStore) CreateAnalysis(a *models.Analysis) error {
	return s.create(s.db, a)
}

func (s *AnalysisStore) CreateAnalysisTx(tx *sql.Tx, a *models.Analysis) error {
	return s.create(tx, a)
}

func (s *AnalysisStore) create(q Querier, a *models.Analysis) error {
	query := `
		INSERT INTO analyses (id, project_id, analyzer, analyzed_git_head, analyzed_at,
		                      summary, purpose, architecture, notes, raw_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(context.Background(), query,
		a.ID, a.ProjectID, a.Analyzer, a.AnalyzedGitHead, a.AnalyzedAt,
		a.Summary, a.Purpose, a.Architecture, a.Notes, a.RawJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create analysis: %w", err)
	}
	return nil
}

func (s *AnalysisStore) GetAnalysis(id string) (*models.Analysis, error) {
	query := `
		SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
		       summary, purpose, architecture, notes, raw_json
		FROM analyses WHERE id = ?
	`
	a := &models.Analysis{}
	err := s.db.QueryRow(query, id).Scan(
		&a.ID, &a.ProjectID, &a.Analyzer, &a.AnalyzedGitHead, &a.AnalyzedAt,
		&a.Summary, &a.Purpose, &a.Architecture, &a.Notes, &a.RawJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}
	return a, nil
}

func (s *AnalysisStore) ListAnalyses(projectID string) ([]*models.Analysis, error) {
	query := `
		SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
		       summary, purpose, architecture, notes, raw_json
		FROM analyses WHERE project_id = ? ORDER BY analyzed_at DESC
	`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*models.Analysis
	for rows.Next() {
		a := &models.Analysis{}
		if err := rows.Scan(
			&a.ID, &a.ProjectID, &a.Analyzer, &a.AnalyzedGitHead, &a.AnalyzedAt,
			&a.Summary, &a.Purpose, &a.Architecture, &a.Notes, &a.RawJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan analysis row: %w", err)
		}
		analyses = append(analyses, a)
	}
	return analyses, rows.Err()
}

func (s *AnalysisStore) DeleteAnalysis(id string) error {
	_, err := s.db.Exec("DELETE FROM analyses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}
	return nil
}

func (s *AnalysisStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec("DELETE FROM analyses WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete analyses for project: %w", err)
	}
	return nil
}
