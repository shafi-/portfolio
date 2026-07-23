package store

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type MetadataStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMetadataStore(db *sql.DB, logger *zap.Logger) *MetadataStore {
	return &MetadataStore{db: db, logger: logger}
}

func (s *MetadataStore) UpsertMetadata(m *models.Metadata) error {
	return s.upsertMetadata(s.db, m)
}

func (s *MetadataStore) UpsertMetadataTx(tx *sql.Tx, m *models.Metadata) error {
	return s.upsertMetadata(tx, m)
}

type metaExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (s *MetadataStore) upsertMetadata(q metaExecer, m *models.Metadata) error {
	ctx := context.Background()
	query := `
		INSERT OR REPLACE INTO metadata (
			project_id, git_head, default_branch, last_commit_at,
			last_modified_at, commit_count, language_summary,
			framework_summary, dependency_summary, documentation_hash, last_scan_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(ctx, query,
		m.ProjectID, nullIfEmpty(m.GitHead), nullIfEmpty(m.DefaultBranch),
		nullIfEmpty(m.LastCommitAt), nullIfEmpty(m.LastModifiedAt),
		m.CommitCount, nullIfEmpty(m.LanguageSummary),
		nullIfEmpty(m.FrameworkSummary), nullIfEmpty(m.DependencySummary),
		nullIfEmpty(m.DocumentationHash), nullIfEmpty(m.LastScanAt),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert metadata: %w", err)
	}
	return nil
}

func (s *MetadataStore) GetMetadata(projectID string) (*models.Metadata, error) {
	query := `
		SELECT project_id, git_head, default_branch, last_commit_at,
		       last_modified_at, commit_count, language_summary,
		       framework_summary, dependency_summary, documentation_hash, last_scan_at
		FROM metadata WHERE project_id = ?
	`
	m := &models.Metadata{}
	var gitHead, defaultBranch, lastCommitAt, lastModifiedAt *string
	var languageSummary, frameworkSummary, dependencySummary, documentationHash, lastScanAt *string

	err := s.db.QueryRow(query, projectID).Scan(
		&m.ProjectID, &gitHead, &defaultBranch, &lastCommitAt,
		&lastModifiedAt, &m.CommitCount, &languageSummary,
		&frameworkSummary, &dependencySummary, &documentationHash, &lastScanAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if gitHead != nil {
		m.GitHead = *gitHead
	}
	if defaultBranch != nil {
		m.DefaultBranch = *defaultBranch
	}
	if lastCommitAt != nil {
		m.LastCommitAt = *lastCommitAt
	}
	if lastModifiedAt != nil {
		m.LastModifiedAt = *lastModifiedAt
	}
	if languageSummary != nil {
		m.LanguageSummary = *languageSummary
	}
	if frameworkSummary != nil {
		m.FrameworkSummary = *frameworkSummary
	}
	if dependencySummary != nil {
		m.DependencySummary = *dependencySummary
	}
	if documentationHash != nil {
		m.DocumentationHash = *documentationHash
	}
	if lastScanAt != nil {
		m.LastScanAt = *lastScanAt
	}
	return m, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
