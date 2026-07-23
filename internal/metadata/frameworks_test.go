package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
)

func TestDetectFrameworks_React(t *testing.T) {
	dir := t.TempDir()

	pkg := `{"dependencies": {"react": "^18.0.0", "express": "^4.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	markers := metadata.DefaultFrameworkMarkers()
	result, err := metadata.DetectFrameworks(dir, nil, markers)
	if err != nil {
		t.Fatalf("DetectFrameworks failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if *result != "Express, React" {
		t.Errorf("expected 'Express, React' (sorted), got %q", *result)
	}
}

func TestDetectFrameworks_GoGin(t *testing.T) {
	dir := t.TempDir()

	gomod := `module example
go 1.21
require (
	github.com/gin-gonic/gin v1.9.0
)
`
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644)

	markers := metadata.DefaultFrameworkMarkers()
	result, err := metadata.DetectFrameworks(dir, nil, markers)
	if err != nil {
		t.Fatalf("DetectFrameworks failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result != "Gin" {
		t.Errorf("expected 'Gin', got %q", *result)
	}
}

func TestDetectFrameworks_Django(t *testing.T) {
	dir := t.TempDir()

	req := "django==4.2\nrequests==2.31.0\n"
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(req), 0644)

	markers := metadata.DefaultFrameworkMarkers()
	result, err := metadata.DetectFrameworks(dir, nil, markers)
	if err != nil {
		t.Fatalf("DetectFrameworks failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result != "Django" {
		t.Errorf("expected 'Django', got %q", *result)
	}
}

func TestDetectFrameworks_MultiFramework(t *testing.T) {
	dir := t.TempDir()

	pkg := `{"dependencies": {"react": "^18.0.0", "express": "^4.0.0", "next": "^13.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	gomod := `module example
go 1.21
require github.com/gin-gonic/gin v1.9.0
`
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644)

	markers := metadata.DefaultFrameworkMarkers()
	result, err := metadata.DetectFrameworks(dir, nil, markers)
	if err != nil {
		t.Fatalf("DetectFrameworks failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if *result != "Express, Gin, Next.js, React" {
		t.Errorf("expected 'Express, Gin, Next.js, React' (sorted), got %q", *result)
	}
}
