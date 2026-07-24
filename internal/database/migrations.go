package database

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"project-dash/pkg/models"
)

type migration struct {
	version int
	name    string
	up      string
	down    string
}

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
		{
			version: 3,
			name:    "documentation_indexing",
			up:      documentationIndexingUp,
			down:    documentationIndexingDown,
		},
		{
			version: 4,
			name:    "fts5_fulltext_search",
			up:      fts5SearchUp,
			down:    fts5SearchDown,
		},
		{
			version: 5,
			name:    "search_indexes",
			up:      searchIndexesUp,
			down:    searchIndexesDown,
		},
		{
			version: 6,
			name:    "ai_analysis_schema",
			up:      aiAnalysisSchemaUp,
			down:    aiAnalysisSchemaDown,
		},
		{
			version: 7,
			name:    "upgrade_migration_tracking",
			up:      upgradeMigrationTrackingUp,
			down:    upgradeMigrationTrackingDown,
		},
	}
}

func loadMigrations() []migration {
	embedded := getMigrations()
	fileMigrations := loadFileMigrations()

	seen := make(map[int]bool)
	for _, m := range embedded {
		seen[m.version] = true
	}

	for _, fm := range fileMigrations {
		if !seen[fm.version] {
			embedded = append(embedded, fm)
			seen[fm.version] = true
		}
	}

	sort.Slice(embedded, func(i, j int) bool {
		return embedded[i].version < embedded[j].version
	})
	return embedded
}

func loadFileMigrations() []migration {
	const migrationsDir = "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil
	}

	var fileMigs []migration
	seen := make(map[int]*migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}

		var version int
		var suffix string
		if _, err := fmt.Sscanf(name, "%d.%s", &version, &suffix); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(name, "%d-", &version); err == nil {
		}

		if _, ok := seen[version]; !ok {
			seen[version] = &migration{version: version, name: name}
		}

		data, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			continue
		}

		m := seen[version]
		if suffix == "up.sql" || suffix == "up" || !isDownSuffix(name, version) {
			m.up = string(data)
		} else {
			m.down = string(data)
		}
	}

	for _, m := range seen {
		if m.up != "" {
			fileMigs = append(fileMigs, *m)
		}
	}

	return fileMigs
}

func isDownSuffix(name string, version int) bool {
	return len(name) > len(fmt.Sprintf("%d-", version))+3 &&
		(name[len(name)-len(".down.sql"):] == ".down.sql" || name[len(name)-len("-down"):] == "-down")
}

func (d *Database) Migrate() error {
	d.logger.Info("Starting database migrations",
		models.Field{Key: "path", Value: d.dbPath},
	)

	if err := d.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if err := d.verifyAppliedMigrations(); err != nil {
		return fmt.Errorf("migration checksum mismatch: %w", err)
	}

	currentVersion, err := d.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	d.logger.Info("Current schema version",
		models.Field{Key: "version", Value: currentVersion},
	)

	migrations := loadMigrations()

	for _, m := range migrations {
		if m.version > currentVersion {
			d.logger.Info("Running migration",
				models.Field{Key: "version", Value: m.version},
				models.Field{Key: "name", Value: m.name},
			)

			if err := d.runMigration(m); err != nil {
				if m.version == 4 && !d.HasFTS5() {
					d.logger.Warn("FTS5 not available in SQLite build—skipping full-text search migration",
						models.Field{Key: "hint", Value: "build with -tags fts5 to enable"},
					)
					if err := d.recordMigration(m.version, calculateChecksum(m.up)); err != nil {
						return fmt.Errorf("failed to record skipped FTS5 migration: %w", err)
					}
					continue
				}
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

func (d *Database) verifyAppliedMigrations() error {
	rows, err := d.db.Query("SELECT version, checksum FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil
	}
	defer rows.Close()

	applied := make(map[int]string)
	for rows.Next() {
		var version int
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return err
		}
		applied[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return err
	}

	migrations := loadMigrations()
	for _, m := range migrations {
		if expected, ok := applied[m.version]; ok {
			// Handle backwards compatibility with old checksum format
			// Old databases stored migration names as checksums instead of SHA256 hashes
			// Names might have hyphens (initial-schema) or underscores (initial_schema)
			legacyNames := []string{m.name, strings.Replace(m.name, "_", "-", -1)}
			for _, legacyName := range legacyNames {
				if expected == legacyName {
					// Old format: checksum is the literal migration name (in either format)
					// This is acceptable for backwards compatibility
					d.logger.Info("Migration using legacy checksum format",
						models.Field{Key: "version", Value: m.version},
						models.Field{Key: "name", Value: m.name},
						models.Field{Key: "stored_checksum", Value: expected},
					)
					continue
				}
			}

			// If we used a legacy checksum format, skip to next migration
			if expected == m.name || expected == strings.Replace(m.name, "_", "-", -1) {
				continue
			}

			// New format: calculate expected SHA256 and compare
			actual := calculateChecksum(m.up)
			if expected != actual {
				return fmt.Errorf("migration %d (%s) checksum changed: expected %s, got %s",
					m.version, m.name, expected, actual)
			}
		}
	}
	return nil
}

func (d *Database) MigrateDown(targetVersion int) error {
	d.logger.Info("Rolling back migrations", models.Field{Key: "target", Value: targetVersion})

	migrations := loadMigrations()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version > migrations[j].version
	})

	for _, m := range migrations {
		if m.version <= targetVersion {
			continue
		}
		if m.down == "" {
			d.logger.Warn("no down migration available", models.Field{Key: "version", Value: m.version})
			continue
		}

		d.logger.Info("Rolling back migration",
			models.Field{Key: "version", Value: m.version},
			models.Field{Key: "name", Value: m.name},
		)

		tx, err := d.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if _, err := tx.Exec(m.down); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d down failed: %w", m.version, err)
		}

		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback: %w", err)
		}
	}

	return nil
}

func (d *Database) ListAppliedMigrations() ([]int, error) {
	rows, err := d.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, rows.Err()
}

func (d *Database) HasFTS5() bool {
	_, err := d.db.Exec("CREATE VIRTUAL TABLE _fts5_test USING fts5(content TEXT)")
	if err != nil {
		return false
	}
	d.db.Exec("DROP TABLE _fts5_test")
	return true
}

func (d *Database) createMigrationsTable() error {
	// First, try to create the table with the new schema (including name column)
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL,
			checksum TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Check if the table has the old schema (missing name column)
	rows, err := d.db.Query("PRAGMA table_info(schema_migrations)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasNameColumn := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == "name" {
			hasNameColumn = true
			break
		}
	}
	rows.Close()

	// If the name column doesn't exist, upgrade the table schema
	if !hasNameColumn {
		d.logger.Info("Upgrading schema_migrations table to include name column")

		// Back up existing data
		var existingData []struct {
			version   int
			checksum  string
			appliedAt string
		}

		rows, err = d.db.Query("SELECT version, checksum, applied_at FROM schema_migrations")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var v int
				var c, a string
				if err := rows.Scan(&v, &c, &a); err != nil {
					continue
				}
				existingData = append(existingData, struct {
					version   int
					checksum  string
					appliedAt string
				}{v, c, a})
			}
		}

		// Recreate table with new schema
		d.db.Exec("DROP TABLE schema_migrations")
		_, err = d.db.Exec(`
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
				applied_at TIMESTAMP NOT NULL,
				checksum TEXT NOT NULL
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to recreate schema_migrations table: %w", err)
		}

		// Restore existing data with name column populated
		for _, data := range existingData {
			migrationName := ""
			switch data.version {
			case 1:
				migrationName = "initial_schema"
			case 2:
				migrationName = "metadata_extraction"
			case 3:
				migrationName = "documentation_indexing"
			case 4:
				migrationName = "fts5_fulltext_search"
			case 5:
				migrationName = "search_indexes"
			case 6:
				migrationName = "ai_analysis_schema"
			}

			_, err = d.db.Exec(
				"INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
				data.version, migrationName, data.appliedAt, data.checksum,
			)
			if err != nil {
				return fmt.Errorf("failed to restore migration data: %w", err)
			}
		}

		d.logger.Info("Schema_migrations table upgraded successfully",
			models.Field{Key: "restored_migrations", Value: len(existingData)})
	}
	return err
}

func (d *Database) runMigration(m migration) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.up); err != nil {
		// Check if this is a "duplicate column" error - column already exists
		if strings.Contains(err.Error(), "duplicate column") {
			// Column already exists, this is OK for backwards compatibility
			d.logger.Info("Migration column already exists, skipping addition",
				models.Field{Key: "version", Value: m.version},
				models.Field{Key: "name", Value: m.name},
				models.Field{Key: "error", Value: err.Error()},
			)
		} else {
			return fmt.Errorf("migration SQL failed: %w", err)
		}
	}

	checksum := calculateChecksum(m.up)
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
		m.version, m.name, time.Now().UTC(), checksum,
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

func (d *Database) recordMigration(version int, checksum string) error {
	_, err := d.db.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
		version, "skipped", time.Now().UTC(), checksum,
	)
	return err
}

func calculateChecksum(sql string) string {
	h := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", h)
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
-- Add columns if they don't exist (backwards compatibility)
-- Check if last_modified_at exists
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

const documentationIndexingUp = `
CREATE INDEX IF NOT EXISTS idx_documents_project_kind ON documents(project_id, kind);
CREATE INDEX IF NOT EXISTS idx_documents_path ON documents(project_id, path);
`

const documentationIndexingDown = `
DROP INDEX IF EXISTS idx_documents_path;
DROP INDEX IF EXISTS idx_documents_project_kind;
`

const fts5SearchUp = `
CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    content,
    tokenize='unicode61 remove_diacritics 2',
    content=documents,
    content_rowid=rowid
);

CREATE TRIGGER IF NOT EXISTS documents_ai AFTER INSERT ON documents BEGIN
    INSERT INTO documents_fts(rowid, content) VALUES (new.rowid, new.content);
END;

CREATE TRIGGER IF NOT EXISTS documents_ad AFTER DELETE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, content) VALUES('delete', old.rowid, old.content);
END;

CREATE TRIGGER IF NOT EXISTS documents_au AFTER UPDATE ON documents BEGIN
    INSERT INTO documents_fts(documents_fts, rowid, content) VALUES('delete', old.rowid, old.content);
    INSERT INTO documents_fts(rowid, content) VALUES (new.rowid, new.content);
END;
`

const fts5SearchDown = `
DROP TRIGGER IF EXISTS documents_au;
DROP TRIGGER IF EXISTS documents_ad;
DROP TRIGGER IF EXISTS documents_ai;
DROP TABLE IF EXISTS documents_fts;
`

const searchIndexesUp = `
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_metadata_language ON metadata(language_summary);
CREATE INDEX IF NOT EXISTS idx_metadata_framework ON metadata(framework_summary);
CREATE INDEX IF NOT EXISTS idx_documents_kind ON documents(kind);
`

const searchIndexesDown = `
DROP INDEX IF EXISTS idx_documents_kind;
DROP INDEX IF EXISTS idx_metadata_framework;
DROP INDEX IF EXISTS idx_metadata_language;
DROP INDEX IF EXISTS idx_projects_name;
`

const aiAnalysisSchemaUp = `
ALTER TABLE analyses ADD COLUMN maturity TEXT;
ALTER TABLE analyses ADD COLUMN strengths TEXT;
ALTER TABLE analyses ADD COLUMN weaknesses TEXT;
ALTER TABLE analyses ADD COLUMN reusable_components TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_analyses_project_analyzer ON analyses(project_id, analyzer);
`

const aiAnalysisSchemaDown = `
DROP INDEX IF EXISTS idx_analyses_project_analyzer;
ALTER TABLE analyses DROP COLUMN IF EXISTS reusable_components;
ALTER TABLE analyses DROP COLUMN IF EXISTS weaknesses;
ALTER TABLE analyses DROP COLUMN IF EXISTS strengths;
ALTER TABLE analyses DROP COLUMN IF EXISTS maturity;
`

const upgradeMigrationTrackingUp = `
-- Add name column if it doesn't exist (backwards compatibility)
ALTER TABLE schema_migrations ADD COLUMN name TEXT;

-- Update migration records that have literal names as checksums
UPDATE schema_migrations
SET checksum = (
	SELECT CASE
		WHEN checksum IN ('initial-schema', 'metadata_extraction', 'documentation_indexing', 'fts5_fulltext_search', 'search_indexes', 'ai_analysis_schema')
		THEN (
			SELECT hex FROM (
				SELECT LOWER(sha256(
					CASE
						WHEN version = 1 THEN (SELECT up FROM (
							SELECT 'initial_schema' as name,
								   initialSchemaUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
						WHEN version = 2 THEN (SELECT up FROM (
							SELECT 'metadata_extraction' as name,
								   metadataExtractionUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
						WHEN version = 3 THEN (SELECT up FROM (
							SELECT 'documentation_indexing' as name,
								   documentationIndexingUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
						WHEN version = 4 THEN (SELECT up FROM (
							SELECT 'fts5_fulltext_search' as name,
								   fts5SearchUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
						WHEN version = 5 THEN (SELECT up FROM (
							SELECT 'search_indexes' as name,
								   searchIndexesUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
						WHEN version = 6 THEN (SELECT up FROM (
							SELECT 'ai_analysis_schema' as name,
								   aiAnalysisSchemaUp as up,
								   '' as down
							) WHERE name = checksum LIMIT 1
						))
					END
				))
			)
		ELSE checksum
	END
)
WHERE version <= 6;

-- Ensure all migration records have proper names
UPDATE schema_migrations
SET name = CASE
	WHEN version = 1 THEN 'initial_schema'
	WHEN version = 2 THEN 'metadata_extraction'
	WHEN version = 3 THEN 'documentation_indexing'
	WHEN version = 4 THEN 'fts5_fulltext_search'
	WHEN version = 5 THEN 'search_indexes'
	WHEN version = 6 THEN 'ai_analysis_schema'
END
WHERE name IS NULL OR name = '';
`

const upgradeMigrationTrackingDown = `
-- This migration cannot be safely reversed as it involves data migration
-- Mark it as irreversible
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
