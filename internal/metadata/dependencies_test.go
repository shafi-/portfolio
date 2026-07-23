package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
)

func TestDetectDependencies_Npm(t *testing.T) {
	dir := t.TempDir()

	pkg := `{"dependencies": {"react": "^18.0.0", "express": "^4.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	deps, summary, err := metadata.DetectDependencies(dir, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	if deps == nil {
		t.Fatal("expected non-nil deps")
	}
	if len(deps) != 2 {
		t.Errorf("expected 2 deps, got %d", len(deps))
	}
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
}

func TestDetectDependencies_GoMod(t *testing.T) {
	dir := t.TempDir()

	gomod := `module example
go 1.21
require (
	github.com/gin-gonic/gin v1.9.0
	github.com/spf13/cobra v1.8.0
)
`
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644)

	deps, _, err := metadata.DetectDependencies(dir, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("expected 2 deps, got %d", len(deps))
	}
}

func TestDetectDependencies_Monorepo(t *testing.T) {
	dir := t.TempDir()

	pkg1 := `{"dependencies": {"react": "^18.0.0"}}`
	pkg2 := `{"dependencies": {"lodash": "^4.17.0"}}`

	os.MkdirAll(filepath.Join(dir, "packages", "web"), 0755)
	os.MkdirAll(filepath.Join(dir, "packages", "lib"), 0755)
	os.WriteFile(filepath.Join(dir, "packages", "web", "package.json"), []byte(pkg1), 0644)
	os.WriteFile(filepath.Join(dir, "packages", "lib", "package.json"), []byte(pkg2), 0644)

	deps, _, err := metadata.DetectDependencies(dir, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("expected 2 deps (deduplicated), got %d", len(deps))
	}
}

func TestDetectDependencies_NoManifests(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	deps, summary, err := metadata.DetectDependencies(dir, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	if deps != nil {
		t.Errorf("expected nil deps for no manifests, got %d items", len(deps))
	}
	if summary != nil {
		t.Errorf("expected nil summary for no manifests, got %q", *summary)
	}
}

func TestDetectDependencies_CorruptedManifest(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "package.json"), []byte("invalid json"), 0644)

	deps, summary, err := metadata.DetectDependencies(dir, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	if deps != nil {
		t.Errorf("expected nil deps for corrupted manifest, got %d items", len(deps))
	}
	if summary != nil {
		t.Errorf("expected nil summary for corrupted manifest, got %q", *summary)
	}
}
