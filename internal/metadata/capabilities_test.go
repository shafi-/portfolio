package metadata_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"project-dash/internal/metadata"
	"project-dash/pkg/models"
)

func TestDetectCapabilities_FromDependencies(t *testing.T) {
	deps := []models.Dependency{
		{Name: "stripe", Manager: "npm"},
		{Name: "pg", Manager: "npm"},
		{Name: "passport", Manager: "npm"},
		{Name: "lodash", Manager: "npm"}, // no match
	}
	caps, err := metadata.DetectCapabilities(deps, nil)
	if err != nil {
		t.Fatalf("DetectCapabilities failed: %v", err)
	}
	if caps == nil {
		t.Fatal("expected non-nil capabilities")
	}

	got := strings.Split(*caps, ", ")
	want := map[string]bool{"auth": true, "database": true, "payments": true}
	if len(got) != len(want) {
		t.Fatalf("expected %d capabilities, got %q", len(want), *caps)
	}
	for _, c := range got {
		if !want[c] {
			t.Errorf("unexpected capability %q in %q", c, *caps)
		}
	}
}

func TestDetectCapabilities_SortedAndUnique(t *testing.T) {
	// redis + ioredis both map to database -> must be deduplicated; jwt -> auth.
	deps := []models.Dependency{
		{Name: "redis", Manager: "npm"},
		{Name: "ioredis", Manager: "npm"},
		{Name: "jsonwebtoken", Manager: "npm"},
	}
	caps, err := metadata.DetectCapabilities(deps, nil)
	if err != nil {
		t.Fatalf("DetectCapabilities failed: %v", err)
	}
	if caps == nil {
		t.Fatal("expected non-nil capabilities")
	}
	parts := strings.Split(*caps, ", ")
	if !sort.StringsAreSorted(parts) {
		t.Errorf("expected sorted capabilities, got %q", *caps)
	}
	if *caps != "auth, database" {
		t.Errorf("expected 'auth, database', got %q", *caps)
	}
}

func TestDetectCapabilities_None(t *testing.T) {
	deps := []models.Dependency{{Name: "lodash", Manager: "npm"}}
	caps, err := metadata.DetectCapabilities(deps, nil)
	if err != nil {
		t.Fatalf("DetectCapabilities failed: %v", err)
	}
	if caps != nil {
		t.Errorf("expected nil capabilities, got %q", *caps)
	}
}

func TestDetectCapabilities_RealManifest(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"dependencies": {"stripe": "^12.0.0", "pg": "^8.0.0", "passport": "^0.7.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	deps, _, err := metadata.DetectDependencies(dir, nil, nil)
	if err != nil {
		t.Fatalf("DetectDependencies failed: %v", err)
	}
	caps, err := metadata.DetectCapabilities(deps, nil)
	if err != nil {
		t.Fatalf("DetectCapabilities failed: %v", err)
	}
	if caps == nil {
		t.Fatal("expected non-nil capabilities from manifest")
	}
	if *caps != "auth, database, payments" {
		t.Errorf("expected 'auth, database, payments', got %q", *caps)
	}
}
