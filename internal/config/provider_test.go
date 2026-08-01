package config

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/pkg/models"
)

func TestNewProvider_DefaultPath(t *testing.T) {
	p := NewProvider("")
	expected := models.GetConfigPath()
	if p.ConfigPath() != expected {
		t.Errorf("expected %s, got %s", expected, p.ConfigPath())
	}
}

func TestNewProvider_ExplicitPath(t *testing.T) {
	p := NewProvider("/tmp/test-config/config.toml")
	if p.ConfigPath() != "/tmp/test-config/config.toml" {
		t.Errorf("unexpected config path: %s", p.ConfigPath())
	}
}

func TestProvider_Load_CreatesDefault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	p := NewProvider(configPath)
	cfg, err := p.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to be created on first Load()")
	}

	if cfg.General.DatabasePath == "" {
		t.Error("expected default DatabasePath to be set")
	}
}

func TestProvider_Load_CachesResult(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	p := NewProvider(configPath)
	cfg1, _ := p.Load()
	cfg2, _ := p.Load()

	if cfg1 != cfg2 {
		t.Error("expected Load() to return cached result")
	}
}

func TestProvider_Save_UpdatesCache(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	p := NewProvider(configPath)
	p.Load()

	cfg := models.GetDefaultConfig()
	cfg.Discovery.ProjectRoots = []string{"/tmp/projects"}
	if err := p.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, _ := p.Load()
	if len(got.Discovery.ProjectRoots) != 1 || got.Discovery.ProjectRoots[0] != "/tmp/projects" {
		t.Errorf("expected saved roots to persist, got %v", got.Discovery.ProjectRoots)
	}
}

func TestProvider_Update_ModifiesAndSaves(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	p := NewProvider(configPath)
	p.Load()

	err := p.Update(func(cfg *models.Config) error {
		cfg.Discovery.ProjectRoots = append(cfg.Discovery.ProjectRoots, "/new/root")
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, _ := p.Load()
	if len(got.Discovery.ProjectRoots) != 1 || got.Discovery.ProjectRoots[0] != "/new/root" {
		t.Errorf("expected updated roots, got %v", got.Discovery.ProjectRoots)
	}
}

func TestProvider_Invalidate_ForcesReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	p := NewProvider(configPath)
	p.Load()

	// Mutate on disk directly.
	p.Invalidate()
	cfg := models.GetDefaultConfig()
	cfg.Discovery.ProjectRoots = []string{"disk-root"}
	p.Save(cfg)

	p.Invalidate()
	got, _ := p.Load()
	if len(got.Discovery.ProjectRoots) != 1 || got.Discovery.ProjectRoots[0] != "disk-root" {
		t.Errorf("expected reloaded roots, got %v", got.Discovery.ProjectRoots)
	}
}
