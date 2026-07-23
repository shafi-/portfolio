package database

import (
	"fmt"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"project-dash/pkg/models"
)

// migration represents a database migration
type migration struct {
	version int
	name    string
	up      string
	down    string
}

// getMigrations returns all available migrations
func getMigrations() []migration {
	return []migration{
		{
			version: 1,
			name:    "initial_schema",
			up:      initialSchemaUp,
			down:    initialSchemaDown,
		},
		{
			version: 2,
			name:    "metadata_extraction",
			up:      metadataExtractionUp,
			down:    metadataExtractionDown,
		},
	}
}

// Migrate runs pending migrations
func (d *Database) Migrate() error {
	d.logger.Info("Starting database migrations",
		models.Field{Key: "path", Value: d.dbPath},
	)

	// Create migrations table if it doesn't exist
	if err := d.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := d.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	d.logger.Info("Current schema version",
		models.Field{Key: "version", Value: currentVersion},
	)

	// Get all migrations
	migrations := getMigrations()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Run pending migrations
	for _, m := range migrations {
		if m.version > currentVersion {
			d.logger.Info("Running migration",
				models.Field{Key: "version", Value: m.version},
				models.Field{Key: "name", Value: m.name},
			)

			if err := d.runMigration(m); err != nil {
				return fmt.Errorf("migration %d failed: %w", m.version, err)
			}

			d.logger.Info("Migration completed",
				models.Field{Key: "version", Value: m.version},
			)
		}
	}

	d.logger.Info("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the schema_migrations table
func (d *Database) createMigrationsTable() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL,
			checksum TEXT NOT NULL
		);
	`)
	return err
}

// runMigration executes a single migration
func (d *Database) runMigration(m migration) error {
	// Start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Execute migration
	if _, err := tx.Exec(m.up); err != nil {
		return fmt.Errorf("migration SQL failed: %w", err)
	}

	// Record migration
	checksum := calculateChecksum(m.up)
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
		m.version, m.name, time.Now().UTC(), checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// calculateChecksum calculates a simple checksum for migration SQL
func calculateChecksum(sql string) string {
	// Simple checksum - can be improved with proper hash
	return fmt.Sprintf("%x", len(sql)+int(sql[0]))
}

// initial schema migration SQL
const initialSchemaUp = `
-- Projects table
CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	root_path TEXT NOT NULL UNIQUE,
	repository_type TEXT NOT NULL,
	discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Metadata table
CREATE TABLE IF NOT EXISTS metadata (
	project_id TEXT PRIMARY KEY,
	git_head TEXT,
	default_branch TEXT,
	last_commit_at TIMESTAMP,
	language_summary TEXT,
	framework_summary TEXT,
	dependency_summary TEXT,
	documentation_hash TEXT,
	last_scan_at TIMESTAMP,
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Documents table
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

-- Analyses table
CREATE TABLE IF NOT EXISTS analyses (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL,
	analyzer TEXT NOT NULL,
	analyzed_git_head TEXT NOT NULL,
	analyzed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	summary TEXT,
	purpose TEXT,
	architecture TEXT,
	notes TEXT,
	raw_json TEXT,
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Features table
CREATE TABLE IF NOT EXISTS features (
	id TEXT PRIMARY KEY,
	analysis_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	confidence REAL,
	FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
);

-- Technologies table
CREATE TABLE IF NOT EXISTS technologies (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	category TEXT
);

-- Project Technologies relationship table
CREATE TABLE IF NOT EXISTS project_technologies (
	project_id TEXT NOT NULL,
	technology_id TEXT NOT NULL,
	PRIMARY KEY (project_id, technology_id),
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (technology_id) REFERENCES technologies(id) ON DELETE CASCADE
);

-- Relationships table
CREATE TABLE IF NOT EXISTS relationships (
	id TEXT PRIMARY KEY,
	source_project TEXT NOT NULL,
	target_project TEXT NOT NULL,
	type TEXT NOT NULL,
	description TEXT,
	confidence REAL,
	FOREIGN KEY (source_project) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (target_project) REFERENCES projects(id) ON DELETE CASCADE
);

-- Configuration table
CREATE TABLE IF NOT EXISTS configuration (
	key TEXT PRIMARY KEY,
	value TEXT,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_projects_root_path ON projects(root_path);
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_analyses_project_id ON analyses(project_id);
CREATE INDEX IF NOT EXISTS idx_features_analysis_id ON features(analysis_id);
CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_project);
CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_project);
`

const metadataExtractionUp = `
ALTER TABLE metadata ADD COLUMN last_modified_at TIMESTAMP;
ALTER TABLE metadata ADD COLUMN commit_count INTEGER DEFAULT 0;
CREATE TABLE IF NOT EXISTS dependencies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	manager TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(project_id, name, manager)
);
CREATE INDEX IF NOT EXISTS idx_dependencies_project_id ON dependencies(project_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_name ON dependencies(name);
`

const metadataExtractionDown = `
DROP INDEX IF EXISTS idx_dependencies_name;
DROP INDEX IF EXISTS idx_dependencies_project_id;
DROP TABLE IF EXISTS dependencies;
`

const initialSchemaDown = `
DROP INDEX IF EXISTS idx_relationships_target;
DROP INDEX IF EXISTS idx_relationships_source;
DROP INDEX IF EXISTS idx_features_analysis_id;
DROP INDEX IF EXISTS idx_analyses_project_id;
DROP INDEX IF EXISTS idx_documents_project_id;
DROP INDEX IF EXISTS idx_projects_root_path;
DROP TABLE IF EXISTS configuration;
DROP TABLE IF EXISTS relationships;
DROP TABLE IF EXISTS project_technologies;
DROP TABLE IF EXISTS technologies;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS analyses;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS projects;
`
