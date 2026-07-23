package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
)

func TestDetectLanguages_Basic(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.ts"), []byte("const x = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "lib.py"), []byte("print(1)"), 0644)

	langMap := metadata.DefaultLanguageMap()
	result, err := metadata.DetectLanguages(dir, nil, langMap)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	expected := "Go, Python, TypeScript"
	if *result != expected {
		t.Errorf("expected %q, got %q", expected, *result)
	}
}

func TestDetectLanguages_Polyglot(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package b"), 0644)
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("const a = 1"), 0644)
	os.WriteFile(filepath.Join(dir, "a.py"), []byte("print(1)"), 0644)

	langMap := metadata.DefaultLanguageMap()
	result, err := metadata.DetectLanguages(dir, nil, langMap)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if *result != "Go, Python, TypeScript" {
		t.Errorf("expected Go first (most files), got %q", *result)
	}
}

func TestDetectLanguages_IgnoresVendor(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "dep.go"), []byte("package dep"), 0644)

	langMap := metadata.DefaultLanguageMap()
	result, err := metadata.DetectLanguages(dir, nil, langMap)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result != "Go" {
		t.Errorf("expected only Go (vendor excluded), got %q", *result)
	}
}

func TestDetectLanguages_EmptyProject(t *testing.T) {
	dir := t.TempDir()

	langMap := metadata.DefaultLanguageMap()
	result, err := metadata.DetectLanguages(dir, nil, langMap)
	if err != nil {
		t.Fatalf("DetectLanguages failed: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for empty project, got %q", *result)
	}
}
