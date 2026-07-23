package store

import (
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
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO dependencies (project_id, name, manager)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range deps {
		if _, err := stmt.Exec(d.ProjectID, d.Name, d.Manager); err != nil {
			return fmt.Errorf("failed to insert dependency %s/%s: %w", d.Name, d.Manager, err)
		}
	}

	return tx.Commit()
}

func (s *DependencyStore) ReplaceDependencies(projectID string, deps []models.Dependency) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("DELETE FROM dependencies WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("failed to clear dependencies: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO dependencies (project_id, name, manager)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range deps {
		if _, err := stmt.Exec(projectID, d.Name, d.Manager); err != nil {
			return fmt.Errorf("failed to insert dependency %s/%s: %w", d.Name, d.Manager, err)
		}
	}

	return tx.Commit()
}
