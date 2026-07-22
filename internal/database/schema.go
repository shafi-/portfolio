package database

import (
	"fmt"
)

// createSchema creates all database tables
func (d *Database) createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		root_path TEXT NOT NULL UNIQUE,
		repository_type TEXT NOT NULL,
		discovered_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS metadata (
		project_id TEXT NOT NULL PRIMARY KEY,
		git_head TEXT,
		default_branch TEXT,
		last_commit_at TEXT,
		last_modified_at TEXT,
		language_summary TEXT,
		framework_summary TEXT,
		dependency_summary TEXT,
		documentation_hash TEXT,
		last_scan_at TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		path TEXT NOT NULL,
		kind TEXT NOT NULL,
		content TEXT,
		content_hash TEXT NOT NULL,
		indexed_at TEXT NOT NULL,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS analyses (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		analyzer TEXT NOT NULL,
		analyzed_git_head TEXT NOT NULL,
		analyzed_at TEXT NOT NULL,
		summary TEXT,
		purpose TEXT,
		architecture TEXT,
		notes TEXT,
		raw_json TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS features (
		id TEXT PRIMARY KEY,
		analysis_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		confidence REAL,
		FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS technologies (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		category TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS project_technologies (
		project_id TEXT NOT NULL,
		technology_id TEXT NOT NULL,
		PRIMARY KEY (project_id, technology_id),
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (technology_id) REFERENCES technologies(id) ON DELETE CASCADE
	);

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

	CREATE TABLE IF NOT EXISTS configuration (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		checksum TEXT NOT NULL,
		applied_at TEXT NOT NULL
	);

	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
	CREATE INDEX IF NOT EXISTS idx_documents_kind ON documents(kind);
	CREATE INDEX IF NOT EXISTS idx_analyses_project_id ON analyses(project_id);
	CREATE INDEX IF NOT EXISTS idx_features_analysis_id ON features(analysis_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_project);
	CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_project);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}
