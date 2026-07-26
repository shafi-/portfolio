package metadata_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
	"project-dash/internal/metadata"
	"project-dash/pkg/models"
)

type mockStore struct {
	meta map[string]*models.Metadata
}

func newMockStore() *mockStore {
	return &mockStore{meta: make(map[string]*models.Metadata)}
}

func (s *mockStore) UpsertMetadata(m *models.Metadata) error {
	s.meta[m.ProjectID] = m
	return nil
}

func (s *mockStore) GetMetadata(projectID string) (*models.Metadata, error) {
	m, ok := s.meta[projectID]
	if !ok {
		return nil, nil
	}
	return m, nil
}

type mockDependencyStore struct{}

func (s *mockDependencyStore) ReplaceDependencies(projectID string, deps []models.Dependency) error {
	return nil
}

type mockProjectProvider struct {
	projects map[string]*models.Project
}

func newMockProjectProvider() *mockProjectProvider {
	return &mockProjectProvider{projects: make(map[string]*models.Project)}
}

func (p *mockProjectProvider) GetProject(id string) (*models.Project, error) {
	proj, ok := p.projects[id]
	if !ok {
		return nil, nil
	}
	return proj, nil
}

func TestService_ExtractAll(t *testing.T) {
	dir := t.TempDir()

	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "config", "user.email", "test@test.com")
	mustRun(t, dir, "git", "config", "user.name", "Test")

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.ts"), []byte("const x = 1"), 0644)

	pkg := `{"dependencies": {"react": "^18.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "initial commit")

	logger, _ := zap.NewDevelopment()
	store := newMockStore()
	depStore := &mockDependencyStore{}
	projects := newMockProjectProvider()

	projectID := "test-123"
	projects.projects[projectID] = &models.Project{
		ID:       projectID,
		Name:     "test-proj",
		RootPath: dir,
	}

	svc := metadata.NewService(store, depStore, projects, logger)

	m, err := svc.ExtractAll(projectID)
	if err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil metadata")
	}

	if m.GitHead == "" {
		t.Error("expected git_head to be set")
	}
	if m.CommitCount == 0 {
		t.Error("expected commit_count > 0")
	}
	if m.LanguageSummary == "" {
		t.Error("expected language_summary to be set")
	}
	if m.DependencySummary == "" {
		t.Error("expected dependency_summary to be set")
	}
	if m.LastScanAt == "" {
		t.Error("expected last_scan_at to be set")
	}
	if _, err := time.Parse(time.RFC3339, m.LastScanAt); err != nil {
		t.Errorf("last_scan_at not valid RFC3339: %v", err)
	}
}

func TestService_ExtractAll_DeletedProject(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	store := newMockStore()
	depStore := &mockDependencyStore{}
	projects := newMockProjectProvider()

	projectID := "deleted-proj"
	projects.projects[projectID] = &models.Project{
		ID:       projectID,
		Name:     "deleted",
		RootPath: "/nonexistent/path/" + projectID,
	}

	svc := metadata.NewService(store, depStore, projects, logger)

	m, err := svc.ExtractAll(projectID)
	if err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}
	if m != nil {
		t.Error("expected nil metadata for deleted project")
	}
}

func TestService_ExtractAll_UnknownProject(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	store := newMockStore()
	depStore := &mockDependencyStore{}
	projects := newMockProjectProvider()

	svc := metadata.NewService(store, depStore, projects, logger)

	m, err := svc.ExtractAll("unknown")
	if err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}
	if m != nil {
		t.Error("expected nil metadata for unknown project")
	}
}

// TestService_Extract exercises the pure extractor (no DB, no project lookup).
// It is the single source of truth that ExtractAll and the indexer both build on.
func TestService_Extract(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# proj"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	pkg := `{"dependencies": {"pg": "^8.0.0"}, "devDependencies": {"jest": "^29.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "init")

	// store/depStore/projects are nil: Extract never consults them.
	svc := metadata.NewService(nil, nil, nil, zap.NewNop())
	m, deps, err := svc.Extract(dir, nil)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if m == nil {
		t.Fatal("expected metadata, got nil")
	}

	if m.CommitCount != 1 {
		t.Errorf("commit_count: got %d, want 1", m.CommitCount)
	}
	if m.CommitVelocity90d != 1 {
		t.Errorf("commit_velocity_90d: got %d, want 1", m.CommitVelocity90d)
	}
	if m.MaturityScore < 1 {
		t.Errorf("expected non-zero maturity score, got %d", m.MaturityScore)
	}
	if m.CapabilitiesSummary != "database" {
		t.Errorf("capabilities_summary: got %q, want database", m.CapabilitiesSummary)
	}
	if m.LanguageSummary == "" {
		t.Error("expected language summary to be populated for main.go")
	}
	if m.ProjectID != "" {
		t.Errorf("expected empty project id from pure Extract, got %q", m.ProjectID)
	}

	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
	scopes := make(map[string]string, len(deps))
	for _, d := range deps {
		scopes[d.Name] = d.Scope
	}
	if scopes["pg"] != "prod" {
		t.Errorf("pg scope: got %q, want prod", scopes["pg"])
	}
	if scopes["jest"] != "dev" {
		t.Errorf("jest scope: got %q, want dev", scopes["jest"])
	}
}

func TestService_Extract_NonexistentRoot(t *testing.T) {
	svc := metadata.NewService(nil, nil, nil, zap.NewNop())
	m, _, err := svc.Extract("/this/path/does/not/exist/12345", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent root, got nil")
	}
	if m != nil {
		t.Errorf("expected nil metadata on error, got %+v", m)
	}
}
