package indexer

import (
	"context"
	"database/sql"
	"fmt"
)

type DedupEngine struct {
	db *sql.DB
}

func NewDedupEngine(db *sql.DB) *DedupEngine {
	return &DedupEngine{db: db}
}

type hashQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (d *DedupEngine) Resolve(projectID, path, contentHash string) (DedupAction, error) {
	return d.resolve(d.db, projectID, path, contentHash)
}

func (d *DedupEngine) ResolveTx(tx *sql.Tx, projectID, path, contentHash string) (DedupAction, error) {
	return d.resolve(tx, projectID, path, contentHash)
}

func (d *DedupEngine) resolve(q hashQuerier, projectID, path, contentHash string) (DedupAction, error) {
	ctx := context.Background()
	var storedHash string
	err := q.QueryRowContext(ctx,
		"SELECT content_hash FROM documents WHERE project_id = ? AND path = ?",
		projectID, path,
	).Scan(&storedHash)

	if err == sql.ErrNoRows {
		return DedupInsert, nil
	}
	if err != nil {
		return "", fmt.Errorf("resolve dedup: %w", err)
	}

	if storedHash == contentHash {
		return DedupSkip, nil
	}

	return DedupUpdate, nil
}
