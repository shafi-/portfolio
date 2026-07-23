package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type ProjectStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewProjectStore(db *sql.DB, logger *zap.Logger) *ProjectStore {
	return &ProjectStore{db: db, logger: logger}
}

func (s *ProjectStore) UpsertProject(p *models.Project) error {
	return s.upsert(s.db, p)
}

func (s *ProjectStore) UpsertProjectTx(tx *sql.Tx, p *models.Project) error {
	return s.upsert(tx, p)
}

func (s *ProjectStore) upsert(q Querier, p *models.Project) error {
	query := `
		INSERT INTO projects (id, name, root_path, repository_type, discovered_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(root_path) DO UPDATE SET
			name = excluded.name,
			repository_type = excluded.repository_type,
			updated_at = excluded.updated_at
	`
	_, err := q.ExecContext(context.Background(), query,
		p.ID, p.Name, p.RootPath, p.RepositoryType, p.DiscoveredAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert project: %w", err)
	}
	return nil
}

func (s *ProjectStore) GetProject(id string) (*models.Project, error) {
	return s.get(s.db, id)
}

func (s *ProjectStore) GetProjectTx(tx *sql.Tx, id string) (*models.Project, error) {
	return s.get(tx, id)
}

func (s *ProjectStore) get(q Querier, id string) (*models.Project, error) {
	query := `
		SELECT id, name, root_path, repository_type, discovered_at, updated_at
		FROM projects WHERE id = ?
	`
	p := &models.Project{}
	err := q.QueryRowContext(context.Background(), query, id).Scan(
		&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return p, nil
}

func (s *ProjectStore) GetProjectByRoot(rootPath string) (*models.Project, error) {
	query := `
		SELECT id, name, root_path, repository_type, discovered_at, updated_at
		FROM projects WHERE root_path = ?
	`
	p := &models.Project{}
	err := s.db.QueryRow(query, rootPath).Scan(
		&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project by root: %w", err)
	}
	return p, nil
}

func (s *ProjectStore) ListProjects() ([]*models.Project, error) {
	query := `SELECT id, name, root_path, repository_type, discovered_at, updated_at FROM projects ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *ProjectStore) DeleteProject(id string) error {
	_, err := s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}
