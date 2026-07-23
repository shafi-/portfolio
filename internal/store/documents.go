package store

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type DocumentStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDocumentStore(db *sql.DB, logger *zap.Logger) *DocumentStore {
	return &DocumentStore{db: db, logger: logger}
}

func (s *DocumentStore) UpsertDocument(doc *models.Document) error {
	query := `
		INSERT INTO documents (id, project_id, path, kind, content, content_hash, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, path) DO UPDATE SET
			content = excluded.content,
			content_hash = excluded.content_hash,
			indexed_at = excluded.indexed_at
	`
	_, err := s.db.Exec(query,
		doc.ID, doc.ProjectID, doc.Path, doc.Kind,
		doc.Content, doc.ContentHash, doc.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert document: %w", err)
	}
	return nil
}

func (s *DocumentStore) GetDocument(projectID, path string) (*models.Document, error) {
	query := `
		SELECT id, project_id, path, kind, content, content_hash, indexed_at
		FROM documents WHERE project_id = ? AND path = ?
	`
	d := &models.Document{}
	err := s.db.QueryRow(query, projectID, path).Scan(
		&d.ID, &d.ProjectID, &d.Path, &d.Kind,
		&d.Content, &d.ContentHash, &d.IndexedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return d, nil
}

func (s *DocumentStore) DeleteDocuments(projectID string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare("DELETE FROM documents WHERE project_id = ? AND path = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range paths {
		if _, err := stmt.Exec(projectID, p); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", p, err)
		}
	}

	return tx.Commit()
}

func (s *DocumentStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec("DELETE FROM documents WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete documents for project: %w", err)
	}
	return nil
}

func (s *DocumentStore) ListDocuments(projectID string) ([]*models.Document, error) {
	query := `
		SELECT id, project_id, path, kind, content, content_hash, indexed_at
		FROM documents WHERE project_id = ?
		ORDER BY kind, path
	`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		d := &models.Document{}
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Path, &d.Kind,
			&d.Content, &d.ContentHash, &d.IndexedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *DocumentStore) StoreDB() *sql.DB {
	return s.db
}

func (s *DocumentStore) CountByKind(projectID string) (map[string]int, error) {
	query := `SELECT kind, COUNT(*) FROM documents WHERE project_id = ? GROUP BY kind`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents by kind: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count row: %w", err)
		}
		counts[kind] = count
	}
	return counts, rows.Err()
}
