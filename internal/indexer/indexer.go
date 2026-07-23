package indexer

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

type ProjectLister interface {
	ListProjects() ([]*models.Project, error)
}

type Indexer struct {
	runner   *IndexRunner
	projects ProjectLister
	logger   *zap.Logger
}

func NewIndexer(db *sql.DB, logger *zap.Logger) *Indexer {
	docStore := store.NewDocumentStore(db, logger)
	metaStore := store.NewMetadataStore(db, logger)

	return &Indexer{
		runner: NewIndexRunner(docStore, metaStore, logger),
		logger: logger,
	}
}

func (idx *Indexer) WithProjectLister(l ProjectLister) *Indexer {
	idx.projects = l
	return idx
}

func (idx *Indexer) IndexProject(ctx context.Context, projectID, rootPath string) (*IndexResult, error) {
	return idx.runner.Run(ctx, projectID, rootPath)
}

func (idx *Indexer) IndexAll(ctx context.Context) (map[string]*IndexResult, error) {
	if idx.projects == nil {
		return nil, nil
	}

	projects, err := idx.projects.ListProjects()
	if err != nil {
		return nil, err
	}

	results := make(map[string]*IndexResult, len(projects))
	for _, p := range projects {
		r, err := idx.IndexProject(ctx, p.ID, p.RootPath)
		if err != nil {
			idx.logger.Error("index project failed", zap.String("project", p.ID), zap.Error(err))
			continue
		}
		results[p.ID] = r
	}

	return results, nil
}

func (idx *Indexer) Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error) {
	return idx.runner.fts.Search(ctx, query, limit, offset)
}
