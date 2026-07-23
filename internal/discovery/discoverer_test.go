package discovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"project-dash/internal/fs"
	"project-dash/internal/logging"
)

// Mock implementations for testing

type MockConfigProvider struct {
	roots        []string
	ignoredPaths []string
}

func (m *MockConfigProvider) GetProjectRoots() ([]string, error) {
	return m.roots, nil
}

func (m *MockConfigProvider) GetIgnoredPaths() []string {
	return m.ignoredPaths
}

type MockProjectStore struct {
	projects    []*Project
	upsertErr   error
	upsertDelay time.Duration // Optional delay for upsert operations
}

func (m *MockProjectStore) UpsertProject(project *Project) error {
	if m.upsertDelay > 0 {
		time.Sleep(m.upsertDelay)
	}
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.projects = append(m.projects, project)
	return nil
}

func TestNewDiscoverer(t *testing.T) {
	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{}
	store := &MockProjectStore{}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	if discoverer == nil {
		t.Error("expected discoverer to be created")
	}

	if discoverer.maxDepth != 0 {
		t.Errorf("expected maxDepth 0, got %d", discoverer.maxDepth)
	}
}

func TestDiscoverer_DiscoverProjects_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository
	repoPath := filepath.Join(tmpDir, "testrepo", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{tmpDir},
		ignoredPaths: []string{},
	}
	store := &MockProjectStore{}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	ctx := context.Background()
	result, err := discoverer.DiscoverProjects(ctx)

	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}

	if result.Discovered != 1 {
		t.Errorf("expected 1 discovered project, got %d", result.Discovered)
	}

	if len(store.projects) != 1 {
		t.Errorf("expected 1 project in store, got %d", len(store.projects))
	}

	// Verify project details
	project := store.projects[0]
	if project.Name != "testrepo" {
		t.Errorf("expected project name 'testrepo', got '%s'", project.Name)
	}

	if project.RootPath != filepath.Join(tmpDir, "testrepo") {
		t.Errorf("expected root path '%s', got '%s'", filepath.Join(tmpDir, "testrepo"), project.RootPath)
	}

	if project.RepositoryType != "unknown" {
		t.Errorf("expected repository type 'regular', got '%s'", project.RepositoryType)
	}

	if project.ID == "" {
		t.Error("expected project ID to be set")
	}
}

func TestDiscoverer_DiscoverProjects_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple repositories to ensure discovery takes some time
	for i := 0; i < 5; i++ {
		repoPath := filepath.Join(tmpDir, fmt.Sprintf("testrepo%d", i), ".git")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo %d: %v", i, err)
		}
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{tmpDir},
		ignoredPaths: []string{},
	}
	store := &MockProjectStore{
		upsertDelay: 20 * time.Millisecond, // Add delay for each project
	}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	ctx := context.Background()

	// Start first discovery in background
	firstDone := make(chan error, 1)
	go func() {
		_, err := discoverer.DiscoverProjects(ctx)
		firstDone <- err
	}()

	// Give the first discovery a moment to start
	time.Sleep(5 * time.Millisecond)

	// Try to start second discovery - should fail immediately
	secondDone := make(chan error, 1)
	go func() {
		_, err := discoverer.DiscoverProjects(ctx)
		secondDone <- err
	}()

	// The concurrent call should return quickly with an error
	select {
	case err := <-secondDone:
		if err == nil {
			t.Error("expected concurrent discovery to fail")
		}
		if _, ok := err.(*ConcurrentDiscoveryError); !ok {
			t.Errorf("expected ConcurrentDiscoveryError, got %T", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("concurrent discovery should have failed quickly, but it didn't return in time")
	}

	// Wait for first discovery to complete
	if err := <-firstDone; err != nil {
		t.Fatalf("first discovery failed: %v", err)
	}
}

func TestDiscoverer_DiscoverProjects_MultipleRoots(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repositories in multiple roots
	root1 := filepath.Join(tmpDir, "root1")
	root2 := filepath.Join(tmpDir, "root2")

	for _, root := range []string{root1, root2} {
		repoPath := filepath.Join(root, "repo", ".git")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo in %s: %v", root, err)
		}
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{root1, root2},
		ignoredPaths: []string{},
	}
	store := &MockProjectStore{}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	ctx := context.Background()
	result, err := discoverer.DiscoverProjects(ctx)

	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}

	if result.Discovered != 2 {
		t.Errorf("expected 2 discovered projects, got %d", result.Discovered)
	}

	if len(result.RootStats) != 2 {
		t.Errorf("expected 2 root stats, got %d", len(result.RootStats))
	}

	// Check that both roots have stats
	for _, root := range []string{root1, root2} {
		stat, exists := result.RootStats[root]
		if !exists {
			t.Errorf("expected stat for root %s", root)
		}
		if stat.Discovered != 1 {
			t.Errorf("expected 1 discovered in root %s, got %d", root, stat.Discovered)
		}
	}
}

func TestDiscoverer_DiscoverProjects_Ignores(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository in node_modules
	nodeModulesRepo := filepath.Join(tmpDir, "project", "node_modules", "dep", ".git")
	if err := os.MkdirAll(nodeModulesRepo, 0755); err != nil {
		t.Fatalf("failed to create node_modules repo: %v", err)
	}

	// Create a regular repository
	regularRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(regularRepo, 0755); err != nil {
		t.Fatalf("failed to create regular repo: %v", err)
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{tmpDir},
		ignoredPaths: []string{"node_modules"},
	}
	store := &MockProjectStore{}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	ctx := context.Background()
	result, err := discoverer.DiscoverProjects(ctx)

	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}

	// Should only discover the regular repo, not the one in node_modules
	if result.Discovered != 1 {
		t.Errorf("expected 1 discovered project, got %d", result.Discovered)
	}

	if len(store.projects) != 1 {
		t.Errorf("expected 1 project in store, got %d", len(store.projects))
	}

	if store.projects[0].Name != "project" {
		t.Errorf("expected project name 'project', got '%s'", store.projects[0].Name)
	}
}

func TestDiscoverer_DiscoverProjects_StoreError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository
	repoPath := filepath.Join(tmpDir, "testrepo", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{tmpDir},
		ignoredPaths: []string{},
	}
	store := &MockProjectStore{
		upsertErr: errors.New("store error"),
	}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	ctx := context.Background()
	result, err := discoverer.DiscoverProjects(ctx)

	// Discovery should not fail, but should report errors
	if err != nil {
		t.Fatalf("DiscoverProjects should not fail: %v", err)
	}

	// Should have 0 discovered (since upsert failed)
	if result.Discovered != 0 {
		t.Errorf("expected 0 discovered projects due to store error, got %d", result.Discovered)
	}

	// Should have at least one error
	if len(result.Errors) == 0 {
		t.Error("expected errors to be reported")
	}
}

func TestDiscoverer_IsRunning(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository to ensure discovery takes some time
	repoPath := filepath.Join(tmpDir, "testrepo", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{
		roots:        []string{tmpDir},
		ignoredPaths: []string{},
	}
	store := &MockProjectStore{
		upsertDelay: 100 * time.Millisecond, // Add delay to ensure discovery takes time
	}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	// Initially not running
	if discoverer.IsRunning() {
		t.Error("expected IsRunning to be false initially")
	}

	discoveryDone := make(chan struct{})

	// Start discovery in background
	go func() {
		discoverer.DiscoverProjects(context.Background())
		close(discoveryDone)
	}()

	// Poll for a short period to check if discovery is running
	// This is more reliable than a single sleep
	wasRunning := false
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Millisecond)
		if discoverer.IsRunning() {
			wasRunning = true
			break
		}
	}

	if !wasRunning {
		t.Error("expected IsRunning to be true during discovery")
	}

	// Wait for discovery to complete
	<-discoveryDone

	// Should not be running anymore
	if discoverer.IsRunning() {
		t.Error("expected IsRunning to be false after discovery")
	}
}

func TestDiscoverer_createProject(t *testing.T) {
	osFS := fs.NewOSFilesystem()
	config := &MockConfigProvider{}
	store := &MockProjectStore{}
	logger, _ := logging.NewLogger("INFO", "console")

	discoverer := NewDiscoverer(osFS, config, store, logger, 0)

	// Create a repository to test with
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "testrepo")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	project := discoverer.createProject(repoPath)

	if project.Name != "testrepo" {
		t.Errorf("expected name 'testrepo', got '%s'", project.Name)
	}

	if project.RootPath != repoPath {
		t.Errorf("expected root path '%s', got '%s'", repoPath, project.RootPath)
	}

	if project.RepositoryType != "unknown" {
		t.Errorf("expected repository type 'regular', got '%s'", project.RepositoryType)
	}

	if project.ID == "" {
		t.Error("expected project ID to be set")
	}

	if project.DiscoveredAt.IsZero() {
		t.Error("expected DiscoveredAt to be set")
	}

	// Check that DiscoveredAt is in UTC
	if project.DiscoveredAt.Location() != time.UTC {
		t.Error("expected DiscoveredAt to be in UTC")
	}
}
