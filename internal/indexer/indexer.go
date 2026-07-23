package indexer

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
	"project-dash/internal/store"
)

type Indexer struct {
	runner *IndexRunner
	logger *zap.Logger
}

func NewIndexer(db *sql.DB, logger *zap.Logger) *Indexer {
	docStore := store.NewDocumentStore(db, logger)
	metaStore := store.NewMetadataStore(db, logger)

	return &Indexer{
		runner: NewIndexRunner(docStore, metaStore, logger),
		logger: logger,
	}
}

func (idx *Indexer) IndexProject(ctx context.Context, projectID, rootPath string) (*IndexResult, error) {
	return idx.runner.Run(ctx, projectID, rootPath)
}

func (idx *Indexer) Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error) {
	return idx.runner.fts.Search(ctx, query, limit, offset)
}
