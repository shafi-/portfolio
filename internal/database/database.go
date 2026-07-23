package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// Database implements the DatabaseInterface
type Database struct {
	db        *sql.DB
	dbPath    string
	logger    *logging.Logger
	connected bool
}

// NewDatabase creates a new database instance
func NewDatabase(dbPath string, logger *logging.Logger) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	return &Database{
		dbPath: dbPath,
		logger: logger,
	}, nil
}

// Connect establishes database connection
func (d *Database) Connect() error {
	d.logger.Info("Connecting to database",
		models.Field{Key: "path", Value: d.dbPath},
	)

	// Build connection string
	connString := fmt.Sprintf("file:%s?_foreign_keys=on", d.dbPath)

	// Open database connection
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection settings
	db.SetMaxOpenConns(25)                 // Maximum open connections
	db.SetMaxIdleConns(25)                 // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	db.SetConnMaxIdleTime(1 * time.Minute) // Idle connection timeout

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout for concurrent access
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		db.Close()
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = db
	d.connected = true

	d.logger.Info("Database connected successfully",
		models.Field{Key: "path", Value: d.dbPath},
	)

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if !d.connected {
		return nil
	}

	d.logger.Info("Closing database connection",
		models.Field{Key: "path", Value: d.dbPath},
	)

	if err := d.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	d.connected = false
	return nil
}

// DB returns the underlying sql.DB for use by stores
func (d *Database) DB() *sql.DB {
	return d.db
}

// IsConnected returns connection status
func (d *Database) IsConnected() bool {
	return d.connected && d.db != nil
}

// Ping tests database connectivity
func (d *Database) Ping() error {
	if !d.connected {
		return fmt.Errorf("database not connected")
	}
	return d.db.Ping()
}

// Initialize creates the database schema
func (d *Database) Initialize() error {
	d.logger.Info("Initializing database schema",
		models.Field{Key: "path", Value: d.dbPath},
	)

	// Run migrations
	if err := d.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Validate schema
	if err := d.ValidateSchema(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	d.logger.Info("Database initialized successfully")
	return nil
}

// ValidateSchema checks that all required tables exist
func (d *Database) ValidateSchema() error {
	requiredTables := []string{
		"projects", "metadata", "documents", "analyses",
		"features", "technologies", "project_technologies",
		"relationships", "configuration",
	}

	for _, table := range requiredTables {
		var count int
		err := d.db.QueryRow(
			"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if count == 0 {
			return fmt.Errorf("required table missing: %s", table)
		}
	}

	d.logger.Info("Schema validation passed",
		models.Field{Key: "tables", Value: len(requiredTables)},
	)

	return nil
}

// GetSchemaVersion returns the current schema version
func (d *Database) GetSchemaVersion() (int, error) {
	var version int

	// Check if schema_migrations table exists
	var tableCount int
	err := d.db.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'",
	).Scan(&tableCount)

	if err != nil {
		return 0, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if tableCount == 0 {
		// No migrations table - assume version 0
		return 0, nil
	}

	// Get latest version
	err = d.db.QueryRow(
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// GetTableCount returns the number of tables in the database
func (d *Database) GetTableCount() (int, error) {
	var count int
	err := d.db.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table'",
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get table count: %w", err)
	}

	return count, nil
}

// GetProjectCount returns the number of projects in the database
func (d *Database) GetProjectCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT count(*) FROM projects").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get project count: %w", err)
	}
	return count, nil
}

// GetProject returns a project by ID
func (d *Database) GetProject(id string) (*models.Project, error) {
	var p models.Project
	err := d.db.QueryRow(
		"SELECT id, name, root_path, repository_type, discovered_at, updated_at FROM projects WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &p, nil
}

// ListProjects returns all discovered projects
func (d *Database) ListProjects() ([]*models.Project, error) {
	rows, err := d.db.Query(
		"SELECT id, name, root_path, repository_type, discovered_at, updated_at FROM projects ORDER BY name",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetLastDiscoveryTime returns the last discovery timestamp
func (d *Database) GetLastDiscoveryTime() (time.Time, error) {
	var lastScan time.Time
	err := d.db.QueryRow(
		"SELECT MAX(last_scan_at) FROM metadata",
	).Scan(&lastScan)

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last discovery time: %w", err)
	}

	return lastScan, nil
}
