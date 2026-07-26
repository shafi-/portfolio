package metadata

import (
	"os"
	"time"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type Store interface {
	UpsertMetadata(m *models.Metadata) error
	GetMetadata(projectID string) (*models.Metadata, error)
}

type DependencyStore interface {
	ReplaceDependencies(projectID string, deps []models.Dependency) error
}

type ProjectProvider interface {
	GetProject(id string) (*models.Project, error)
}

type Service struct {
	store      Store
	depStore   DependencyStore
	projects   ProjectProvider
	walker     *FileWalker
	langMap    LanguageMap
	fwMarkers  []FrameworkMarker
	depParsers []ManifestParser
	logger     *zap.Logger
}

func NewService(
	store Store,
	depStore DependencyStore,
	projects ProjectProvider,
	logger *zap.Logger,
) *Service {
	return &Service{
		store:      store,
		depStore:   depStore,
		projects:   projects,
		walker:     NewFileWalker(logger),
		langMap:    DefaultLanguageMap(),
		fwMarkers:  DefaultFrameworkMarkers(),
		depParsers: DefaultManifestParsers(),
		logger:     logger,
	}
}

func (s *Service) ExtractGitMetadata(projectID string) error {
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return nil
	}

	gitResult, err := ExtractGitMetadata(project.RootPath)
	if err != nil {
		s.logger.Warn("git metadata extraction failed", zap.String("project", projectID), zap.Error(err))
		return nil
	}

	// Read-modify-write: avoid INSERT OR REPLACE clobbering other metadata columns.
	existing, err := s.store.GetMetadata(projectID)
	if err != nil || existing == nil {
		existing = &models.Metadata{ProjectID: projectID}
	}
	applyGitMetadata(existing, gitResult)
	existing.LastScanAt = time.Now().UTC().Format(time.RFC3339)

	return s.store.UpsertMetadata(existing)
}

// applyGitMetadata copies git facts from a GitResult onto a Metadata struct.
func applyGitMetadata(m *models.Metadata, g *GitResult) {
	if g == nil {
		return
	}
	if g.GitHead != nil {
		m.GitHead = *g.GitHead
	}
	if g.DefaultBranch != nil {
		m.DefaultBranch = *g.DefaultBranch
	}
	if g.LastCommitAt != nil {
		m.LastCommitAt = g.LastCommitAt.Format(time.RFC3339)
	}
	if g.LastModifiedAt != nil {
		m.LastModifiedAt = g.LastModifiedAt.Format(time.RFC3339)
	}
	m.CommitCount = g.CommitCount
	if g.FirstCommitAt != nil {
		m.FirstCommitAt = g.FirstCommitAt.Format(time.RFC3339)
	}
	m.CommitVelocity90d = g.CommitVelocity90d
	m.ContributorCount = g.ContributorCount
	m.TagCount = g.TagCount
	if g.RemoteURL != nil {
		m.RemoteURL = *g.RemoteURL
	}
	m.IsPublished = g.IsPublished
}

func (s *Service) DetectLanguages(projectID string) error {
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return nil
	}

	langSummary, err := DetectLanguages(project.RootPath, s.walker, s.langMap)
	if err != nil {
		s.logger.Warn("language detection failed", zap.String("project", projectID), zap.Error(err))
		return nil
	}

	existing, err := s.store.GetMetadata(projectID)
	if err != nil || existing == nil {
		existing = &models.Metadata{ProjectID: projectID}
	}
	if langSummary != nil {
		existing.LanguageSummary = *langSummary
	}
	existing.LastScanAt = time.Now().UTC().Format(time.RFC3339)

	return s.store.UpsertMetadata(existing)
}

func (s *Service) DetectFrameworks(projectID string) error {
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return nil
	}

	fwSummary, err := DetectFrameworks(project.RootPath, s.walker, s.fwMarkers)
	if err != nil {
		s.logger.Warn("framework detection failed", zap.String("project", projectID), zap.Error(err))
		return nil
	}

	existing, err := s.store.GetMetadata(projectID)
	if err != nil || existing == nil {
		existing = &models.Metadata{ProjectID: projectID}
	}
	if fwSummary != nil {
		existing.FrameworkSummary = *fwSummary
	}
	existing.LastScanAt = time.Now().UTC().Format(time.RFC3339)

	return s.store.UpsertMetadata(existing)
}

// Extract runs all deterministic extractors against rootPath and returns an
// assembled Metadata (without ProjectID) plus the parsed dependencies. It does
// not touch the database — callers own persistence. This is the single source
// of truth for extraction; individual extractor errors are tolerated (best
// effort), mirroring the prior behavior of ExtractAll.
func (s *Service) Extract(rootPath string) (*models.Metadata, []models.Dependency, error) {
	if _, err := os.Stat(rootPath); err != nil {
		return nil, nil, err
	}

	gitResult, _ := ExtractGitMetadata(rootPath)
	langSummary, _ := DetectLanguages(rootPath, s.walker, s.langMap)
	fwSummary, _ := DetectFrameworks(rootPath, s.walker, s.fwMarkers)
	deps, depSummary, _ := DetectDependencies(rootPath, s.walker, s.depParsers)

	m := &models.Metadata{}
	applyGitMetadata(m, gitResult)
	if langSummary != nil {
		m.LanguageSummary = *langSummary
	}
	if fwSummary != nil {
		m.FrameworkSummary = *fwSummary
	}
	if depSummary != nil {
		m.DependencySummary = *depSummary
	}

	if caps, _ := DetectCapabilities(deps, nil); caps != nil {
		m.CapabilitiesSummary = *caps
	}

	score, indicators, _ := DetectMaturity(rootPath, nil)
	m.MaturityScore = score
	m.MaturityIndicators = indicators

	return m, deps, nil
}

func (s *Service) ExtractAll(projectID string) (*models.Metadata, error) {
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, nil
	}

	rootPath := project.RootPath
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		s.logger.Warn("project directory not found, skipping", zap.String("project", projectID), zap.String("path", rootPath))
		return nil, nil
	}

	m, deps, err := s.Extract(rootPath)
	if err != nil {
		s.logger.Warn("metadata extraction failed", zap.String("project", projectID), zap.Error(err))
		return nil, nil
	}
	if m == nil {
		return nil, nil
	}

	m.ProjectID = projectID
	m.LastScanAt = time.Now().UTC().Format(time.RFC3339)

	if err := s.store.UpsertMetadata(m); err != nil {
		s.logger.Warn("failed to upsert metadata", zap.String("project", projectID), zap.Error(err))
	}

	if deps != nil {
		for i := range deps {
			deps[i].ProjectID = projectID
		}
		if err := s.depStore.ReplaceDependencies(projectID, deps); err != nil {
			s.logger.Warn("failed to replace dependencies", zap.String("project", projectID), zap.Error(err))
		}
	}

	return m, nil
}
