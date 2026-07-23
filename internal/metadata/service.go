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
	store     Store
	depStore  DependencyStore
	projects  ProjectProvider
	walker    *FileWalker
	langMap   LanguageMap
	fwMarkers []FrameworkMarker
	logger    *zap.Logger
}

func NewService(
	store Store,
	depStore DependencyStore,
	projects ProjectProvider,
	logger *zap.Logger,
) *Service {
	return &Service{
		store:     store,
		depStore:  depStore,
		projects:  projects,
		walker:    NewFileWalker(logger),
		langMap:   DefaultLanguageMap(),
		fwMarkers: DefaultFrameworkMarkers(),
		logger:    logger,
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

	m := &models.Metadata{
		ProjectID:  projectID,
		LastScanAt: time.Now().UTC().Format(time.RFC3339),
	}

	if gitResult.GitHead != nil {
		m.GitHead = *gitResult.GitHead
	}
	if gitResult.DefaultBranch != nil {
		m.DefaultBranch = *gitResult.DefaultBranch
	}
	if gitResult.LastCommitAt != nil {
		m.LastCommitAt = gitResult.LastCommitAt.Format(time.RFC3339)
	}
	if gitResult.LastModifiedAt != nil {
		m.LastModifiedAt = gitResult.LastModifiedAt.Format(time.RFC3339)
	}
	m.CommitCount = gitResult.CommitCount

	return s.store.UpsertMetadata(m)
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

	gitResult, _ := ExtractGitMetadata(rootPath)

	langSummary, _ := DetectLanguages(rootPath, s.walker, s.langMap)

	fwSummary, _ := DetectFrameworks(rootPath, s.walker, s.fwMarkers)

	deps, depSummary, _ := DetectDependencies(rootPath, s.walker)

	m := &models.Metadata{
		ProjectID:  projectID,
		LastScanAt: time.Now().UTC().Format(time.RFC3339),
	}

	if gitResult != nil {
		if gitResult.GitHead != nil {
			m.GitHead = *gitResult.GitHead
		}
		if gitResult.DefaultBranch != nil {
			m.DefaultBranch = *gitResult.DefaultBranch
		}
		if gitResult.LastCommitAt != nil {
			m.LastCommitAt = gitResult.LastCommitAt.Format(time.RFC3339)
		}
		if gitResult.LastModifiedAt != nil {
			m.LastModifiedAt = gitResult.LastModifiedAt.Format(time.RFC3339)
		}
		m.CommitCount = gitResult.CommitCount
	}
	if langSummary != nil {
		m.LanguageSummary = *langSummary
	}
	if fwSummary != nil {
		m.FrameworkSummary = *fwSummary
	}
	if depSummary != nil {
		m.DependencySummary = *depSummary
	}

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
