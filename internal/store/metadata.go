package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (s *MetadataStore) upsertMetadata(q Querier, m *models.Metadata) error {
	ctx := context.Background()
	query := `
		INSERT OR REPLACE INTO metadata (
			project_id, git_head, default_branch, last_commit_at,
			last_modified_at, commit_count, language_summary,
			framework_summary, dependency_summary, documentation_hash, last_scan_at,
			first_commit_at, commit_velocity_90d, contributor_count, tag_count,
			remote_url, is_published, maturity_score, maturity_indicators, capabilities_summary
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := q.ExecContext(ctx, query,
		m.ProjectID, nullIfEmpty(m.GitHead), nullIfEmpty(m.DefaultBranch),
		nullIfEmpty(m.LastCommitAt), nullIfEmpty(m.LastModifiedAt),
		m.CommitCount, nullIfEmpty(m.LanguageSummary),
		nullIfEmpty(m.FrameworkSummary), nullIfEmpty(m.DependencySummary),
		nullIfEmpty(m.DocumentationHash), nullIfEmpty(m.LastScanAt),
		nullIfEmpty(m.FirstCommitAt), m.CommitVelocity90d, m.ContributorCount,
		m.TagCount, nullIfEmpty(m.RemoteURL), boolToInt(m.IsPublished),
		m.MaturityScore, nullIfEmpty(m.MaturityIndicators), nullIfEmpty(m.CapabilitiesSummary),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert metadata: %w", err)
	}
	return nil
}

func (s *MetadataStore) GetMetadata(projectID string) (*models.Metadata, error) {
	return s.getMetadata(s.db, projectID)
}

func (s *MetadataStore) GetMetadataTx(tx *sql.Tx, projectID string) (*models.Metadata, error) {
	return s.getMetadata(tx, projectID)
}

func (s *MetadataStore) getMetadata(q Querier, projectID string) (*models.Metadata, error) {
	query := `
		SELECT project_id, git_head, default_branch, last_commit_at,
		       last_modified_at, commit_count, language_summary,
		       framework_summary, dependency_summary, documentation_hash, last_scan_at,
		       first_commit_at, commit_velocity_90d, contributor_count, tag_count,
		       remote_url, is_published, maturity_score, maturity_indicators, capabilities_summary
		FROM metadata WHERE project_id = ?
	`
	m := &models.Metadata{}
	var gitHead, defaultBranch, lastCommitAt, lastModifiedAt *string
	var languageSummary, frameworkSummary, dependencySummary, documentationHash, lastScanAt *string
	var firstCommitAt, remoteURL, maturityIndicators, capabilitiesSummary *string
	var isPublished int

	err := q.QueryRowContext(context.Background(), query, projectID).Scan(
		&m.ProjectID, &gitHead, &defaultBranch, &lastCommitAt,
		&lastModifiedAt, &m.CommitCount, &languageSummary,
		&frameworkSummary, &dependencySummary, &documentationHash, &lastScanAt,
		&firstCommitAt, &m.CommitVelocity90d, &m.ContributorCount, &m.TagCount,
		&remoteURL, &isPublished, &m.MaturityScore, &maturityIndicators, &capabilitiesSummary,
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
	if firstCommitAt != nil {
		m.FirstCommitAt = *firstCommitAt
	}
	if remoteURL != nil {
		m.RemoteURL = *remoteURL
	}
	m.IsPublished = isPublished != 0
	if maturityIndicators != nil {
		m.MaturityIndicators = *maturityIndicators
	}
	if capabilitiesSummary != nil {
		m.CapabilitiesSummary = *capabilitiesSummary
	}
	return m, nil
}

func (s *MetadataStore) GetMetadataBatch(projectIDs []string) (map[string]*models.Metadata, error) {
	if len(projectIDs) == 0 {
		return make(map[string]*models.Metadata), nil
	}

	ctx := context.Background()
	// Build placeholder string for IN clause: (?,?,?)
	placeholders := make([]string, len(projectIDs))
	args := make([]interface{}, len(projectIDs))
	for i, id := range projectIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT project_id, git_head, default_branch, last_commit_at,
		       last_modified_at, commit_count, language_summary,
		       framework_summary, dependency_summary, documentation_hash, last_scan_at,
		       first_commit_at, commit_velocity_90d, contributor_count, tag_count,
		       remote_url, is_published, maturity_score, maturity_indicators, capabilities_summary
		FROM metadata WHERE project_id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata batch: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*models.Metadata)
	for rows.Next() {
		m := &models.Metadata{}
		var gitHead, defaultBranch, lastCommitAt, lastModifiedAt *string
		var languageSummary, frameworkSummary, dependencySummary, documentationHash, lastScanAt *string
		var firstCommitAt, remoteURL, maturityIndicators, capabilitiesSummary *string
		var isPublished int

		err := rows.Scan(
			&m.ProjectID, &gitHead, &defaultBranch, &lastCommitAt,
			&lastModifiedAt, &m.CommitCount, &languageSummary,
			&frameworkSummary, &dependencySummary, &documentationHash, &lastScanAt,
			&firstCommitAt, &m.CommitVelocity90d, &m.ContributorCount, &m.TagCount,
			&remoteURL, &isPublished, &m.MaturityScore, &maturityIndicators, &capabilitiesSummary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata row: %w", err)
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
		if firstCommitAt != nil {
			m.FirstCommitAt = *firstCommitAt
		}
		if remoteURL != nil {
			m.RemoteURL = *remoteURL
		}
		m.IsPublished = isPublished != 0
		if maturityIndicators != nil {
			m.MaturityIndicators = *maturityIndicators
		}
		if capabilitiesSummary != nil {
			m.CapabilitiesSummary = *capabilitiesSummary
		}

		result[m.ProjectID] = m
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata rows: %w", err)
	}

	return result, nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
