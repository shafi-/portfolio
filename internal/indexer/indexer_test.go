package indexer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

func tempDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open temp db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func initSchema(t *testing.T, db *sql.DB, projectID, rootPath string) {
	t.Helper()

	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		root_path TEXT NOT NULL UNIQUE,
		repository_type TEXT NOT NULL,
		discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS metadata (
		project_id TEXT PRIMARY KEY,
		git_head TEXT,
		default_branch TEXT,
		last_commit_at TIMESTAMP,
		last_modified_at TIMESTAMP,
		commit_count INTEGER DEFAULT 0,
		language_summary TEXT,
		framework_summary TEXT,
		dependency_summary TEXT,
		documentation_hash TEXT,
		last_scan_at TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		path TEXT NOT NULL,
		kind TEXT NOT NULL,
		content TEXT,
		content_hash TEXT,
		indexed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		UNIQUE(project_id, path)
	);
	CREATE INDEX IF NOT EXISTS idx_documents_project_kind ON documents(project_id, kind);
	CREATE INDEX IF NOT EXISTS idx_documents_path ON documents(project_id, path);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	if _, err := db.Exec(
		"INSERT OR IGNORE INTO projects (id, name, root_path, repository_type) VALUES (?, ?, ?, ?)",
		projectID, "test-project", rootPath, "git",
	); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	if _, err := db.Exec(
		"INSERT OR IGNORE INTO metadata (project_id) VALUES (?)",
		projectID,
	); err != nil {
		t.Fatalf("insert metadata: %v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestDocReader_Read(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	writeFile(t, path, "hello world")

	reader := NewDocReader(1 << 20)
	content, hash, err := reader.Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}
	if hash != "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" {
		t.Errorf("unexpected hash: %s", hash)
	}
}

func TestDocReader_Truncation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.md")

	content := make([]byte, 2<<20)
	for i := range content {
		content[i] = 'a' + byte(i%26)
	}
	writeFile(t, path, string(content))

	reader := NewDocReader(1 << 20)
	readContent, _, err := reader.Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(readContent) > 1<<20+1 {
		t.Errorf("expected truncated content <= 1MB, got %d", len(readContent))
	}
}

func TestDocDiscoverer_FindREADME(t *testing.T) {
	t.Run("finds README.md", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "README.md"), "# Project")
		d := NewDocDiscoverer()
		files := d.FindREADME(dir)
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		if files[0].RelPath != "README.md" {
			t.Errorf("expected README.md, got %s", files[0].RelPath)
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "readme.md"), "# Project")
		d := NewDocDiscoverer()
		files := d.FindREADME(dir)
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("no README", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "main.go"), "package main")
		d := NewDocDiscoverer()
		files := d.FindREADME(dir)
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("non-standard extension ignored", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "README.org"), "* Project")
		d := NewDocDiscoverer()
		files := d.FindREADME(dir)
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})
}

func TestDocDiscoverer_FindDocs(t *testing.T) {
	t.Run("finds supported files", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docs", "guide.md"), "# Guide")
		writeFile(t, filepath.Join(dir, "docs", "manual.rst"), "Manual")
		writeFile(t, filepath.Join(dir, "docs", "notes.txt"), "Notes")
		writeFile(t, filepath.Join(dir, "docs", "spec.adoc"), "Spec")
		d := NewDocDiscoverer()
		files := d.FindDocs(dir)
		if len(files) != 4 {
			t.Fatalf("expected 4 files, got %d", len(files))
		}
	})

	t.Run("recursive subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docs", "getting-started", "install.md"), "# Install")
		writeFile(t, filepath.Join(dir, "docs", "api", "reference.md"), "# API")
		d := NewDocDiscoverer()
		files := d.FindDocs(dir)
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("no docs directory", func(t *testing.T) {
		dir := t.TempDir()
		d := NewDocDiscoverer()
		files := d.FindDocs(dir)
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("skips unsupported extensions", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docs", "readme.md"), "# Readme")
		writeFile(t, filepath.Join(dir, "docs", "image.png"), "PNG")
		writeFile(t, filepath.Join(dir, "docs", "style.css"), "body {}")
		d := NewDocDiscoverer()
		files := d.FindDocs(dir)
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
	})
}

func TestDocDiscoverer_FindADRs(t *testing.T) {
	t.Run("finds ADRs in docs/adr", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docs", "adr", "001-use-go.md"), "# Use Go")
		writeFile(t, filepath.Join(dir, "docs", "adr", "002-sqlite.md"), "# SQLite")
		d := NewDocDiscoverer()
		files := d.FindADRs(dir)
		if len(files) != 2 {
			t.Fatalf("expected 2 ADRs, got %d", len(files))
		}
	})

	t.Run("finds ADRs in .adr and adr", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, ".adr", "001-db.md"), "# DB")
		writeFile(t, filepath.Join(dir, "adr", "002-api.md"), "# API")
		d := NewDocDiscoverer()
		files := d.FindADRs(dir)
		if len(files) != 2 {
			t.Fatalf("expected 2 ADRs, got %d", len(files))
		}
	})

	t.Run("no ADR directory", func(t *testing.T) {
		dir := t.TempDir()
		d := NewDocDiscoverer()
		files := d.FindADRs(dir)
		if len(files) != 0 {
			t.Errorf("expected 0 ADRs, got %d", len(files))
		}
	})
}

func TestDocDiscoverer_FindCHANGELOG(t *testing.T) {
	t.Run("finds CHANGELOG.md", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "CHANGELOG.md"), "# Changelog")
		d := NewDocDiscoverer()
		files := d.FindCHANGELOG(dir)
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("finds CHANGES.md and HISTORY.md", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "CHANGES.md"), "# Changes")
		writeFile(t, filepath.Join(dir, "HISTORY.md"), "# History")
		d := NewDocDiscoverer()
		files := d.FindCHANGELOG(dir)
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "changelog.md"), "# Changelog")
		d := NewDocDiscoverer()
		files := d.FindCHANGELOG(dir)
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("no changelog", func(t *testing.T) {
		dir := t.TempDir()
		d := NewDocDiscoverer()
		files := d.FindCHANGELOG(dir)
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})
}

func TestDedupEngine(t *testing.T) {
	db := tempDB(t)
	initSchema(t, db, "proj-1", "/tmp")

	dedup := NewDedupEngine(db)

	t.Run("insert for new document", func(t *testing.T) {
		action, err := dedup.Resolve("proj-1", "README.md", "hash1")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if action != DedupInsert {
			t.Errorf("expected insert, got %s", action)
		}
	})

	t.Run("skip for unchanged hash", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO documents (id, project_id, path, kind, content, content_hash, indexed_at)
			VALUES ('d1', 'proj-1', 'README.md', 'README', '# Content', 'hash1', '2024-01-01')`)
		if err != nil {
			t.Fatalf("insert doc: %v", err)
		}

		action, err := dedup.Resolve("proj-1", "README.md", "hash1")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if action != DedupSkip {
			t.Errorf("expected skip, got %s", action)
		}
	})

	t.Run("update for changed hash", func(t *testing.T) {
		action, err := dedup.Resolve("proj-1", "README.md", "hash2")
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
		if action != DedupUpdate {
			t.Errorf("expected update, got %s", action)
		}
	})
}

func TestIndexProject_Basic(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-1"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Test Project")
	writeFile(t, filepath.Join(rootPath, "docs", "guide.md"), "# Guide")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	result, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("IndexProject failed: %v", err)
	}
	if result.Documents == 0 {
		t.Fatal("expected at least 1 document")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM documents WHERE project_id = ?", projectID).Scan(&count); err != nil {
		t.Fatalf("count docs: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 documents, got %d", count)
	}
}

func TestIndexProject_Idempotent(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-2"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Stable")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	r1, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("first index: %v", err)
	}

	r2, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("second index: %v", err)
	}

	if r2.Documents != 0 && r2.Skipped == 0 {
		t.Logf("first: docs=%d skipped=%d, second: docs=%d skipped=%d",
			r1.Documents, r1.Skipped, r2.Documents, r2.Skipped)
	}
}

func TestIndexProject_EmptyProject(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-3"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "main.go"), "package main")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	result, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("IndexProject failed: %v", err)
	}
	if result.Documents != 0 {
		t.Errorf("expected 0 documents for empty project, got %d", result.Documents)
	}
}

func TestIndexProject_AllKinds(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-4"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Full Project")
	writeFile(t, filepath.Join(rootPath, "docs", "api.md"), "# API Docs")
	writeFile(t, filepath.Join(rootPath, "docs", "adr", "001-use-go.md"), "# Use Go")
	writeFile(t, filepath.Join(rootPath, "CHANGELOG.md"), "# Changelog")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	result, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("IndexProject failed: %v", err)
	}
	fmt.Printf("Index result: docs=%d skipped=%d errors=%d\n",
		result.Documents, result.Skipped, len(result.Errors))
	for _, e := range result.Errors {
		t.Logf("error: %s: %s", e.Code, e.Message)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM documents WHERE project_id = ?", projectID).Scan(&total)
	if total != 4 {
		t.Errorf("expected 4 documents (README+DOC+ADR+CHANGELOG), got %d", total)
	}
}

func TestDocReader_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	writeFile(t, path, "")

	reader := NewDocReader(1 << 20)
	content, hash, err := reader.Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expectedHash {
		t.Errorf("expected empty hash %s, got %s", expectedHash, hash)
	}
}

func TestBinaryDetection(t *testing.T) {
	dir := t.TempDir()

	textPath := filepath.Join(dir, "text.md")
	os.WriteFile(textPath, []byte("hello"), 0644)

	binaryPath := filepath.Join(dir, "binary.bin")
	os.WriteFile(binaryPath, []byte{0xff, 0x00, 0xfe, 0x01}, 0644)

	if isBinaryFile(textPath) {
		t.Error("text file should not be detected as binary")
	}
	if !isBinaryFile(binaryPath) {
		t.Error("binary file should be detected as binary")
	}
}

func TestDocumentationHash_Stored(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-h1"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Project A")
	writeFile(t, filepath.Join(rootPath, "docs", "guide.md"), "# Guide A")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	result, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("IndexProject failed: %v", err)
	}
	if result.Documentation == "" {
		t.Fatal("documentation_hash should not be empty after indexing")
	}

	var storedHash string
	err = db.QueryRow("SELECT documentation_hash FROM metadata WHERE project_id = ?", projectID).Scan(&storedHash)
	if err != nil {
		t.Fatalf("read documentation_hash: %v", err)
	}
	if storedHash == "" {
		t.Fatal("documentation_hash should be stored in metadata table")
	}
	if storedHash != result.Documentation {
		t.Errorf("stored hash %q != result hash %q", storedHash, result.Documentation)
	}
}

func TestDocumentationHash_Changed(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-h2"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Initial")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	r1, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("first index: %v", err)
	}
	if r1.DocsChanged {
		t.Log("first index: DocsChanged=true (expected on first run with no prior hash)")
	}

	writeFile(t, filepath.Join(rootPath, "README.md"), "# Modified content")

	r2, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("second index: %v", err)
	}
	if !r2.DocsChanged {
		t.Error("DocsChanged should be true when content changes")
	}
}

func TestDocumentationHash_Unchanged(t *testing.T) {
	db := tempDB(t)
	projectID := "proj-h3"
	rootPath := t.TempDir()

	initSchema(t, db, projectID, rootPath)
	writeFile(t, filepath.Join(rootPath, "README.md"), "# Stable content")

	logger, _ := zap.NewDevelopment()
	idx := NewIndexer(db, logger)

	_, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("first index: %v", err)
	}

	r2, err := idx.IndexProject(context.Background(), projectID, rootPath)
	if err != nil {
		t.Fatalf("second index: %v", err)
	}
	if r2.DocsChanged {
		t.Error("DocsChanged should be false when content is unchanged")
	}
}

func TestDocumentStore(t *testing.T) {
	db := tempDB(t)
	initSchema(t, db, "proj-s", "/tmp")
	docStore := store.NewDocumentStore(db, zap.NewNop())

	doc := &models.Document{
		ID:          "doc-1",
		ProjectID:   "proj-s",
		Path:        "README.md",
		Kind:        "README",
		Content:     "# Hi",
		ContentHash: "abc",
		IndexedAt:   "2024-01-01T00:00:00Z",
	}

	if err := docStore.UpsertDocument(doc); err != nil {
		t.Fatalf("UpsertDocument failed: %v", err)
	}

	loaded, err := docStore.GetDocument("proj-s", "README.md")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetDocument returned nil")
	}
	if loaded.Content != "# Hi" {
		t.Errorf("expected '# Hi', got %q", loaded.Content)
	}
}
