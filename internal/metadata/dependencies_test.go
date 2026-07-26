package metadata_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
	"project-dash/pkg/models"
)

func TestDetectDependencies_Npm(t *testing.T) {
	dir := t.TempDir()

	pkg := `{"dependencies": {"react": "^18.0.0", "express": "^4.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	deps, summary, err := metadata.DetectDependencies(dir, nil, nil)
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

	deps, _, err := metadata.DetectDependencies(dir, nil, nil)
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

	deps, _, err := metadata.DetectDependencies(dir, nil, nil)
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

	deps, summary, err := metadata.DetectDependencies(dir, nil, nil)
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

	deps, summary, err := metadata.DetectDependencies(dir, nil, nil)
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

func TestDetectDependencies_NpmScope(t *testing.T) {
	dir := t.TempDir()

	pkg := `{
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"jest": "^29.0.0"}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	deps, _, err := metadata.DetectDependencies(dir, nil, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}

	byName := make(map[string]string, len(deps))
	for _, d := range deps {
		byName[d.Name] = d.Scope
	}

	if byName["react"] != "prod" {
		t.Errorf("expected react scope=prod, got %q", byName["react"])
	}
	if byName["jest"] != "dev" {
		t.Errorf("expected jest scope=dev, got %q", byName["jest"])
	}
}

func TestDetectDependencies_DuplicateNameAcrossScopes(t *testing.T) {
	// Same package in dependencies and devDependencies: distinct scope keeps
	// both in memory (dedupe key includes scope).
	dir := t.TempDir()

	pkg := `{
		"dependencies": {"typescript": "^5.0.0"},
		"devDependencies": {"typescript": "^5.0.0"}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	deps, _, err := metadata.DetectDependencies(dir, nil, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps (prod + dev), got %d", len(deps))
	}
}

// composerParser is a stand-in for a not-yet-supported ecosystem. It proves the
// registry is extensible: adding an ecosystem is a new parser type appended to
// the default set, with zero edits to the DetectDependencies dispatcher.
type composerParser struct{}

func (composerParser) Filename() string { return "composer.json" }

func (composerParser) Parse(content []byte) ([]models.Dependency, error) {
	var doc struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(content, &doc); err != nil {
		return nil, nil
	}
	var deps []models.Dependency
	for name := range doc.Require {
		deps = append(deps, models.Dependency{Name: name, Manager: "composer", Scope: "prod"})
	}
	return deps, nil
}

func TestDetectDependencies_CustomParser(t *testing.T) {
	dir := t.TempDir()

	// A real npm manifest (handled by the default registry) ...
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies": {"react": "^18.0.0"}}`), 0644)
	// ... plus an ecosystem the defaults do not know about.
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require": {"vendor/pkg": "^1.0"}}`), 0644)

	// Default registry + the custom parser, with no dispatcher changes.
	parsers := append(metadata.DefaultManifestParsers(), composerParser{})
	deps, _, err := metadata.DetectDependencies(dir, nil, parsers)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}

	byName := make(map[string]string, len(deps))
	for _, d := range deps {
		byName[d.Name] = d.Manager
	}
	if byName["react"] != "npm" {
		t.Errorf("default npm parser dropped: react manager got %q, want npm", byName["react"])
	}
	if byName["vendor/pkg"] != "composer" {
		t.Errorf("custom composer parser not invoked: vendor/pkg manager got %q, want composer", byName["vendor/pkg"])
	}
}
