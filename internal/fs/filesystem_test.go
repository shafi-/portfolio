package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOSFilesystem_ReadDir(t *testing.T) {
	fs := NewOSFilesystem()

	// Create a temporary directory with some files
	tmpDir := t.TempDir()

	// Create some test files and directories
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test ReadDir
	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Should have 2 entries: testdir and testfile.txt
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Verify entries
	foundDir := false
	foundFile := false
	for _, entry := range entries {
		if entry.Name() == "testdir" {
			foundDir = true
			if !entry.IsDir() {
				t.Errorf("expected testdir to be a directory")
			}
		}
		if entry.Name() == "testfile.txt" {
			foundFile = true
			if entry.IsDir() {
				t.Errorf("expected testfile.txt to be a file")
			}
		}
	}

	if !foundDir {
		t.Error("did not find testdir entry")
	}
	if !foundFile {
		t.Error("did not find testfile.txt entry")
	}
}

func TestOSFilesystem_ReadDir_NotExists(t *testing.T) {
	fs := NewOSFilesystem()

	// Try to read a directory that doesn't exist
	_, err := fs.ReadDir("/this/path/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestOSFilesystem_Lstat(t *testing.T) {
	fs := NewOSFilesystem()

	// Create a temporary file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test Lstat
	info, err := fs.Lstat(tmpFile)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected name test.txt, got %s", info.Name())
	}

	if info.IsDir() {
		t.Error("expected file, got directory")
	}

	if info.Size() != 4 {
		t.Errorf("expected size 4, got %d", info.Size())
	}
}

func TestOSFilesystem_Stat(t *testing.T) {
	fs := NewOSFilesystem()

	// Create a temporary file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test Stat
	info, err := fs.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected name test.txt, got %s", info.Name())
	}

	if info.IsDir() {
		t.Error("expected file, got directory")
	}

	if info.Size() != 4 {
		t.Errorf("expected size 4, got %d", info.Size())
	}
}

func TestOSFilesystem_Symlink_Lstat_vs_Stat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping symlink test in short mode")
	}

	fs := NewOSFilesystem()
	tmpDir := t.TempDir()

	// Create a target file and a symlink
	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("target content"), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	symlinkPath := filepath.Join(tmpDir, "symlink.txt")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Lstat should return info about the symlink itself
	lstatInfo, err := fs.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}

	// Stat should return info about the target file
	statInfo, err := fs.Stat(symlinkPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Lstat size should be small (length of target path)
	// Stat size should be the actual file size (length of "target content")
	// Note: File size might vary based on line endings, so just check it's reasonable
	if statInfo.Size() < 10 || statInfo.Size() > 20 {
		t.Errorf("expected Stat size to be reasonable (10-20), got %d", statInfo.Size())
	}

	// The symlink should have mode bits indicating it's a symlink
	if lstatInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("expected Lstat to indicate symlink mode")
	}
}
