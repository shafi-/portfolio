package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type DependencyStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDependencyStore(db *sql.DB, logger *zap.Logger) *DependencyStore {
	return &DependencyStore{db: db, logger: logger}
}

func (s *DependencyStore) InsertDependencies(deps []models.Dependency) error {
	return s.insert(s.db, deps)
}

func (s *DependencyStore) InsertDependenciesTx(tx *sql.Tx, deps []models.Dependency) error {
	return s.insert(tx, deps)
}

func (s *DependencyStore) insert(q Querier, deps []models.Dependency) error {
	for _, d := range deps {
		_, err := q.ExecContext(context.Background(),
			"INSERT OR IGNORE INTO dependencies (project_id, name, manager) VALUES (?, ?, ?)",
			d.ProjectID, d.Name, d.Manager,
		)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %s/%s: %w", d.Name, d.Manager, err)
		}
	}
	return nil
}

func (s *DependencyStore) ReplaceDependencies(projectID string, deps []models.Dependency) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.replaceTx(tx, projectID, deps); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DependencyStore) ReplaceDependenciesTx(tx *sql.Tx, projectID string, deps []models.Dependency) error {
	return s.replaceTx(tx, projectID, deps)
}

func (s *DependencyStore) replaceTx(tx *sql.Tx, projectID string, deps []models.Dependency) error {
	if _, err := tx.Exec("DELETE FROM dependencies WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("failed to clear dependencies: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO dependencies (project_id, name, manager) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range deps {
		if _, err := stmt.Exec(projectID, d.Name, d.Manager); err != nil {
			return fmt.Errorf("failed to insert dependency %s/%s: %w", d.Name, d.Manager, err)
		}
	}
	return nil
}

func (s *DependencyStore) ListDependencies(projectID string) ([]models.Dependency, error) {
	rows, err := s.db.Query("SELECT project_id, name, manager FROM dependencies WHERE project_id = ? ORDER BY name", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list dependencies: %w", err)
	}
	defer rows.Close()

	var deps []models.Dependency
	for rows.Next() {
		var d models.Dependency
		if err := rows.Scan(&d.ProjectID, &d.Name, &d.Manager); err != nil {
			return nil, fmt.Errorf("failed to scan dependency row: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

func (s *DependencyStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec("DELETE FROM dependencies WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete dependencies for project: %w", err)
	}
	return nil
}
