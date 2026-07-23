package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"project-dash/internal/metadata"
)

func TestWalk_IgnoresDirs(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	os.WriteFile(filepath.Join(dir, "vendor", "lib.go"), []byte("package lib"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	logger, _ := zap.NewDevelopment()
	walker := metadata.NewFileWalker(logger)

	var files []string
	err := walker.Walk(dir, func(path string, info os.FileInfo) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	for _, f := range files {
		if filepath.Base(f) == "lib.go" {
			t.Errorf("expected vendor/lib.go to be skipped, but it was walked")
		}
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestWalk_MaxFiles(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 100; i++ {
		name := filepath.Join(dir, "f"+filepath.Base("/a/"+string(rune(65+i)))+".txt")
		os.WriteFile(name, []byte("content"), 0644)
	}

	logger, _ := zap.NewDevelopment()
	walker := metadata.NewFileWalker(logger)

	var count int
	cfg := metadata.WalkConfig{
		IgnoredDirs:   []string{"vendor", "node_modules", ".git"},
		MaxFiles:      10,
		FollowSymlink: false,
	}

	err := walker.WalkWithConfig(dir, cfg, func(path string, info os.FileInfo) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}
	if count != 10 {
		t.Errorf("expected 10 files, got %d", count)
	}
}
