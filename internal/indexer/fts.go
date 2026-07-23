package indexer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type FTSManager struct {
	db      *sql.DB
	hasFTS5 bool
}

func NewFTSManager(db *sql.DB) *FTSManager {
	m := &FTSManager{db: db}
	m.hasFTS5 = m.checkFTS5()
	return m
}

func (f *FTSManager) checkFTS5() bool {
	_, err := f.db.Exec("CREATE VIRTUAL TABLE _fts_check USING fts5(content TEXT)")
	if err != nil {
		return false
	}
	f.db.Exec("DROP TABLE _fts_check")
	return true
}

func (f *FTSManager) HasFTS5() bool {
	return f.hasFTS5
}

func (f *FTSManager) Rebuild(ctx context.Context) error {
	if !f.hasFTS5 {
		return nil
	}

	tx, err := f.db.Begin()
	if err != nil {
		return fmt.Errorf("begin fts rebuild tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM documents_fts"); err != nil {
		return fmt.Errorf("clear fts table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO documents_fts(rowid, content)
		SELECT rowid, content FROM documents
	`); err != nil {
		return fmt.Errorf("populate fts table: %w", err)
	}

	return tx.Commit()
}

func (f *FTSManager) Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	if f.hasFTS5 {
		return f.fts5Search(ctx, query, limit, offset)
	}
	return f.likeSearch(ctx, query, limit, offset)
}

func (f *FTSManager) fts5Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error) {
	sqlQuery := `
		SELECT d.id, d.project_id, d.path, d.kind,
		       d.content, d.content_hash, d.indexed_at,
		       bm25(documents_fts) AS rank
		FROM documents_fts
		JOIN documents d ON documents_fts.rowid = d.rowid
		WHERE documents_fts MATCH ?
		ORDER BY rank
		LIMIT ? OFFSET ?
	`

	rows, err := f.db.QueryContext(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fts5 search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.Path, &r.Kind,
			&r.Content, &r.ContentHash, &r.IndexedAt, &r.Rank); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (f *FTSManager) likeSearch(ctx context.Context, query string, limit, offset int) ([]SearchResult, error) {
	likePattern := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"

	sqlQuery := `
		SELECT id, project_id, path, kind,
		       content, content_hash, indexed_at,
		       0.0 AS rank
		FROM documents
		WHERE content LIKE ? ESCAPE '\\'
		ORDER BY kind, path
		LIMIT ? OFFSET ?
	`

	rows, err := f.db.QueryContext(ctx, sqlQuery, likePattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("like search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.Path, &r.Kind,
			&r.Content, &r.ContentHash, &r.IndexedAt, &r.Rank); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
