package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/logging"
)

func TestNewDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if db == nil {
		t.Fatal("Database is nil")
	}

	if db.dbPath != dbPath {
		t.Errorf("Expected dbPath %s, got %s", dbPath, db.dbPath)
	}
}

func TestDatabaseConnection(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Test connection
	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !db.IsConnected() {
		t.Error("Database should be connected")
	}

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// Test close
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	if db.IsConnected() {
		t.Error("Database should not be connected after close")
	}
}

func TestDatabaseInitialization(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Initialize database
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Validate schema
	if err := db.ValidateSchema(); err != nil {
		t.Errorf("Schema validation failed: %v", err)
	}

	// Check schema version
	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Errorf("Failed to get schema version: %v", err)
	}

	if version < 3 {
		t.Errorf("Expected schema version >= 3, got %d", version)
	}

	// Check table count
	tableCount, err := db.GetTableCount()
	if err != nil {
		t.Errorf("Failed to get table count: %v", err)
	}

	if tableCount < 10 {
		t.Errorf("Expected at least 10 tables, got %d", tableCount)
	}
}

func TestMigrationSystem(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrations failed: %v", err)
	}

	// Verify database is accessible and working
	if err := db.Ping(); err != nil {
		t.Errorf("Database not accessible after migrations: %v", err)
	}

	// Check schema version is set
	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Errorf("Failed to get schema version: %v", err)
	}

	if version == 0 {
		t.Error("Schema version should be > 0 after migrations")
	}
}

func TestSchemaValidation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test validation with valid schema
	if err := db.ValidateSchema(); err != nil {
		t.Errorf("Valid schema should pass validation: %v", err)
	}
}

func TestDatabasePermissions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	logger, _ := logging.NewLogger("INFO", "console")

	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := db.Connect(); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Check file permissions
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Errorf("Cannot stat database file: %v", err)
	}

	// On Unix-like systems, check for secure permissions
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		t.Logf("Warning: Database file has loose permissions: %o", mode)
	}
}

func initBenchDB(b *testing.B) *Database {
	b.Helper()
	dir := b.TempDir()
	logger, _ := logging.NewLogger("ERROR", "console")
	db, err := NewDatabase(filepath.Join(dir, "bench.db"), logger)
	if err != nil {
		b.Fatalf("NewDatabase: %v", err)
	}
	if err := db.Connect(); err != nil {
		b.Fatalf("Connect: %v", err)
	}
	if err := db.Initialize(); err != nil {
		b.Fatalf("Initialize: %v", err)
	}

	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("proj-%d", i)
		name := fmt.Sprintf("project-%d", i)
		db.db.Exec("INSERT OR IGNORE INTO projects (id, name, root_path, repository_type) VALUES (?, ?, ?, ?)",
			id, name, "/tmp/"+id, "git")
		db.db.Exec("INSERT OR IGNORE INTO metadata (project_id, language_summary, framework_summary) VALUES (?, ?, ?)",
			id, "Go", "Gin")
	}

	db.db.Exec("INSERT OR IGNORE INTO documents (id, project_id, path, kind, content) VALUES (?, ?, ?, ?, ?)",
		"doc-bench", "proj-0", "README.md", "README", "Benchmark content for FTS search")
	return db
}

func BenchmarkProjectSearch(b *testing.B) {
	db := initBenchDB(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		row := db.db.QueryRow("SELECT COUNT(*) FROM projects WHERE name LIKE ?", "project-%")
		var count int
		row.Scan(&count)
	}
}

func BenchmarkMetadataFilter(b *testing.B) {
	db := initBenchDB(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		row := db.db.QueryRow("SELECT COUNT(*) FROM metadata WHERE language_summary = ?", "Go")
		var count int
		row.Scan(&count)
	}
}

func BenchmarkDocumentKindLookup(b *testing.B) {
	db := initBenchDB(b)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		row := db.db.QueryRow("SELECT COUNT(*) FROM documents WHERE kind = ?", "README")
		var count int
		row.Scan(&count)
	}
}
