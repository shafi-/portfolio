package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type RelationshipStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewRelationshipStore(db *sql.DB, logger *zap.Logger) *RelationshipStore {
	return &RelationshipStore{db: db, logger: logger}
}

func (s *RelationshipStore) CreateRelationship(r *models.Relationship) error {
	return s.create(s.db, r)
}

func (s *RelationshipStore) CreateRelationshipTx(tx *sql.Tx, r *models.Relationship) error {
	return s.create(tx, r)
}

func (s *RelationshipStore) create(q Querier, r *models.Relationship) error {
	query := `
		INSERT INTO relationships (id, source_project, target_project, type, description, confidence)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(context.Background(), query,
		r.ID, r.SourceProject, r.TargetProject, r.Type, r.Description, r.Confidence,
	)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}
	return nil
}

func (s *RelationshipStore) GetRelationship(id string) (*models.Relationship, error) {
	query := `SELECT id, source_project, target_project, type, description, confidence FROM relationships WHERE id = ?`
	r := &models.Relationship{}
	err := s.db.QueryRow(query, id).Scan(
		&r.ID, &r.SourceProject, &r.TargetProject, &r.Type, &r.Description, &r.Confidence,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}
	return r, nil
}

func (s *RelationshipStore) ListRelationships(projectID string) ([]*models.Relationship, error) {
	query := `
		SELECT id, source_project, target_project, type, description, confidence
		FROM relationships WHERE source_project = ? OR target_project = ?
		ORDER BY type
	`
	rows, err := s.db.Query(query, projectID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*models.Relationship
	for rows.Next() {
		r := &models.Relationship{}
		if err := rows.Scan(
			&r.ID, &r.SourceProject, &r.TargetProject, &r.Type, &r.Description, &r.Confidence,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relationship row: %w", err)
		}
		relationships = append(relationships, r)
	}
	return relationships, rows.Err()
}

func (s *RelationshipStore) DeleteRelationship(id string) error {
	_, err := s.db.Exec("DELETE FROM relationships WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}
	return nil
}

func (s *RelationshipStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec("DELETE FROM relationships WHERE source_project = ? OR target_project = ?", projectID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete relationships for project: %w", err)
	}
	return nil
}
