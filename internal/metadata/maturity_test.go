package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
)

func TestDetectMaturity_Basic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# proj"), 0644)
	os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch"), 0644)
	os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)
	os.WriteFile(filepath.Join(dir, ".github", "workflows", "ci.yml"), []byte("on: push"), 0644)

	score, indicators, err := metadata.DetectMaturity(dir, nil)
	if err != nil {
		t.Fatalf("DetectMaturity failed: %v", err)
	}

	// readme(1) + license(1) + dockerfile(1) + ci(2) = 5
	if score != 5 {
		t.Errorf("expected maturity score 5, got %d", score)
	}
	want := `{"ci":true,"dockerfile":true,"license":true,"readme":true}`
	if indicators != want {
		t.Errorf("expected indicators %s, got %s", want, indicators)
	}
}

func TestDetectMaturity_Empty(t *testing.T) {
	dir := t.TempDir()
	score, indicators, err := metadata.DetectMaturity(dir, nil)
	if err != nil {
		t.Fatalf("DetectMaturity failed: %v", err)
	}
	if score != 0 {
		t.Errorf("expected score 0, got %d", score)
	}
	if indicators != "" {
		t.Errorf("expected empty indicators, got %s", indicators)
	}
}

func TestDetectMaturity_TypescriptConfig(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# proj"), 0644)

	score, indicators, err := metadata.DetectMaturity(dir, nil)
	if err != nil {
		t.Fatalf("DetectMaturity failed: %v", err)
	}
	// typescript(1) + readme(1) = 2
	if score != 2 {
		t.Errorf("expected maturity score 2, got %d", score)
	}
	want := `{"readme":true,"typescript":true}`
	if indicators != want {
		t.Errorf("expected indicators %s, got %s", want, indicators)
	}
}

func TestDetectMaturity_DocsDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)
	os.WriteFile(filepath.Join(dir, "docs", "intro.md"), []byte("# intro"), 0644)

	score, _, err := metadata.DetectMaturity(dir, nil)
	if err != nil {
		t.Fatalf("DetectMaturity failed: %v", err)
	}
	if score != 2 { // docs weight 2
		t.Errorf("expected maturity score 2, got %d", score)
	}
}
