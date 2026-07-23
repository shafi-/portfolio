package indexer

import (
	"context"
	"fmt"

	"project-dash/internal/store"
)

type OrphanCleaner struct {
	docStore *store.DocumentStore
}

func NewOrphanCleaner(docStore *store.DocumentStore) *OrphanCleaner {
	return &OrphanCleaner{docStore: docStore}
}

func (c *OrphanCleaner) Clean(ctx context.Context, projectID string, validPaths []string) error {
	valid := make(map[string]bool, len(validPaths))
	for _, p := range validPaths {
		valid[p] = true
	}

	docs, err := c.docStore.ListDocuments(projectID)
	if err != nil {
		return fmt.Errorf("list documents for cleanup: %w", err)
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

	return c.docStore.DeleteDocuments(projectID, orphaned)
}
