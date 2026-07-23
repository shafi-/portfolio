package indexer

import (
	"database/sql"

	"project-dash/internal/store"
)

type OrphanCleaner struct {
	docStore *store.DocumentStore
}

func NewOrphanCleaner(docStore *store.DocumentStore) *OrphanCleaner {
	return &OrphanCleaner{docStore: docStore}
}

func (c *OrphanCleaner) CleanTx(tx *sql.Tx, projectID string, validPaths []string) error {
	docs, err := c.docStore.ListDocumentsTx(tx, projectID)
	if err != nil {
		return err
	}

	valid := make(map[string]bool, len(validPaths))
	for _, p := range validPaths {
		valid[p] = true
	}

	var orphaned []string
	for _, d := range docs {
		if !valid[d.Path] {
			orphaned = append(orphaned, d.Path)
		}
	}

	if len(orphaned) == 0 {
		return nil
	}

	return c.docStore.DeleteDocumentsTx(tx, projectID, orphaned)
}
