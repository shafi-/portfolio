package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type TechnologyStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTechnologyStore(db *sql.DB, logger *zap.Logger) *TechnologyStore {
	return &TechnologyStore{db: db, logger: logger}
}

func (s *TechnologyStore) CreateTechnology(t *models.Technology) error {
	return s.create(s.db, t)
}

func (s *TechnologyStore) CreateTechnologyTx(tx *sql.Tx, t *models.Technology) error {
	return s.create(tx, t)
}

func (s *TechnologyStore) create(q Querier, t *models.Technology) error {
	query := `INSERT INTO technologies (id, name, category) VALUES (?, ?, ?)`
	_, err := q.ExecContext(context.Background(), query, t.ID, t.Name, t.Category)
	if err != nil {
		return fmt.Errorf("failed to create technology: %w", err)
	}
	return nil
}

func (s *TechnologyStore) GetTechnology(id string) (*models.Technology, error) {
	query := `SELECT id, name, category FROM technologies WHERE id = ?`
	t := &models.Technology{}
	err := s.db.QueryRow(query, id).Scan(&t.ID, &t.Name, &t.Category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get technology: %w", err)
	}
	return t, nil
}

func (s *TechnologyStore) GetByName(name string) (*models.Technology, error) {
	query := `SELECT id, name, category FROM technologies WHERE name = ?`
	t := &models.Technology{}
	err := s.db.QueryRow(query, name).Scan(&t.ID, &t.Name, &t.Category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get technology by name: %w", err)
	}
	return t, nil
}

func (s *TechnologyStore) ListTechnologies() ([]*models.Technology, error) {
	query := `SELECT id, name, category FROM technologies ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list technologies: %w", err)
	}
	defer rows.Close()

	var technologies []*models.Technology
	for rows.Next() {
		t := &models.Technology{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Category); err != nil {
			return nil, fmt.Errorf("failed to scan technology row: %w", err)
		}
		technologies = append(technologies, t)
	}
	return technologies, rows.Err()
}

func (s *TechnologyStore) DeleteTechnology(id string) error {
	_, err := s.db.Exec("DELETE FROM technologies WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete technology: %w", err)
	}
	return nil
}

func (s *TechnologyStore) AddProjectTechnology(pt models.ProjectTechnology) error {
	return s.addPT(s.db, pt)
}

func (s *TechnologyStore) AddProjectTechnologyTx(tx *sql.Tx, pt models.ProjectTechnology) error {
	return s.addPT(tx, pt)
}

func (s *TechnologyStore) addPT(q Querier, pt models.ProjectTechnology) error {
	query := `INSERT OR IGNORE INTO project_technologies (project_id, technology_id) VALUES (?, ?)`
	_, err := q.ExecContext(context.Background(), query, pt.ProjectID, pt.TechnologyID)
	if err != nil {
		return fmt.Errorf("failed to add project technology: %w", err)
	}
	return nil
}

func (s *TechnologyStore) ListProjectTechnologies(projectID string) ([]*models.ProjectTechnology, error) {
	query := `SELECT project_id, technology_id FROM project_technologies WHERE project_id = ?`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project technologies: %w", err)
	}
	defer rows.Close()

	var pts []*models.ProjectTechnology
	for rows.Next() {
		pt := &models.ProjectTechnology{}
		if err := rows.Scan(&pt.ProjectID, &pt.TechnologyID); err != nil {
			return nil, fmt.Errorf("failed to scan project technology row: %w", err)
		}
		pts = append(pts, pt)
	}
	return pts, rows.Err()
}
