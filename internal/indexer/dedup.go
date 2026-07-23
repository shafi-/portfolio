package indexer

import (
	"database/sql"
	"fmt"
)

type DedupEngine struct {
	db *sql.DB
}

func NewDedupEngine(db *sql.DB) *DedupEngine {
	return &DedupEngine{db: db}
}

func (d *DedupEngine) Resolve(projectID, path, contentHash string) (DedupAction, error) {
	var storedHash string
	err := d.db.QueryRow(
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
