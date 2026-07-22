package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nerddevsltd/portfolio/internal/logging"
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

	if version != 1 {
		t.Errorf("Expected schema version 1, got %d", version)
	}

	// Check table count
	tableCount, err := db.GetTableCount()
	if err != nil {
		t.Errorf("Failed to get table count: %v", err)
	}

	expectedTables := 10 // 9 data tables + 1 migrations table
	if tableCount < expectedTables {
		t.Errorf("Expected at least %d tables, got %d", expectedTables, tableCount)
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

	// Verify migrations table exists
	result, err := db.ExecuteQuery(
		"SELECT * FROM schema_migrations ORDER BY version",
	)
	if err != nil {
		t.Errorf("Failed to query migrations table: %v", err)
	}

	if len(result.Rows) == 0 {
		t.Error("No migrations recorded")
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