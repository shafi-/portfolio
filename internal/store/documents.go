package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type DocumentStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDocumentStore(db *sql.DB, logger *zap.Logger) *DocumentStore {
	return &DocumentStore{db: db, logger: logger}
}

func (s *DocumentStore) StoreDB() *sql.DB {
	return s.db
}

func (s *DocumentStore) UpsertDocument(doc *models.Document) error {
	return s.upsert(s.db, doc)
}

func (s *DocumentStore) UpsertDocumentTx(tx *sql.Tx, doc *models.Document) error {
	return s.upsert(tx, doc)
}

func (s *DocumentStore) upsert(q Querier, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, project_id, path, kind, content, content_hash, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, path) DO UPDATE SET
			content = excluded.content,
			content_hash = excluded.content_hash,
			indexed_at = excluded.indexed_at
	`
	ctx := context.Background()
	_, err := q.ExecContext(ctx, query,
		doc.ID, doc.ProjectID, doc.Path, doc.Kind,
		doc.Content, doc.ContentHash, doc.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert document: %w", err)
	}
	return nil
}

func (s *DocumentStore) GetDocument(projectID, path string) (*models.Document, error) {
	return s.get(s.db, projectID, path)
}

func (s *DocumentStore) GetDocumentTx(tx *sql.Tx, projectID, path string) (*models.Document, error) {
	return s.get(tx, projectID, path)
}

func (s *DocumentStore) get(q Querier, projectID, path string) (*models.Document, error) {
	ctx := context.Background()
	query := `
		SELECT id, project_id, path, kind, content, content_hash, indexed_at
		FROM documents WHERE project_id = ? AND path = ?
	`
	d := &models.Document{}
	err := q.QueryRowContext(ctx, query, projectID, path).Scan(
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
	return s.deleteDocs(s.db, projectID, paths)
}

func (s *DocumentStore) DeleteDocumentsTx(tx *sql.Tx, projectID string, paths []string) error {
	return s.deleteDocs(tx, projectID, paths)
}

func (s *DocumentStore) deleteDocs(q Querier, projectID string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	ctx := context.Background()
	for _, p := range paths {
		if _, err := q.ExecContext(ctx, "DELETE FROM documents WHERE project_id = ? AND path = ?", projectID, p); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", p, err)
		}
	}
	return nil
}

func (s *DocumentStore) DeleteAllForProject(projectID string) error {
	_, err := s.db.Exec("DELETE FROM documents WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete documents for project: %w", err)
	}
	return nil
}

func (s *DocumentStore) ListDocuments(projectID string) ([]*models.Document, error) {
	return s.list(s.db, projectID)
}

func (s *DocumentStore) ListDocumentsTx(tx *sql.Tx, projectID string) ([]*models.Document, error) {
	return s.list(tx, projectID)
}

func (s *DocumentStore) list(q Querier, projectID string) ([]*models.Document, error) {
	ctx := context.Background()
	query := `
		SELECT id, project_id, path, kind, content, content_hash, indexed_at
		FROM documents WHERE project_id = ?
		ORDER BY kind, path
	`
	rows, err := q.QueryContext(ctx, query, projectID)
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

func (s *DocumentStore) CountByKind(projectID string) (map[string]int, error) {
	ctx := context.Background()
	query := `SELECT kind, COUNT(*) FROM documents WHERE project_id = ? GROUP BY kind`
	rows, err := s.db.QueryContext(ctx, query, projectID)
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
