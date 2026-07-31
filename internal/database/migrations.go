package database

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// getMigrations returns the embedded baseline migrations.
//
// The schema is consolidated into a single initial migration that creates every
// table, column, and index in its final form, plus the FTS5 full-text-search and
// tier-3 feature-deep-dive migrations. CREATE TABLE / CREATE INDEX IF NOT EXISTS
// make each migration idempotent on a fresh database. FTS5 is a separate
// migration so Migrate() can skip it on SQLite builds compiled without FTS5.
//
// APPEND-ONLY CONTRACT (ADR-022): the migrations below are FROZEN. Their up/down
// SQL is immutable once released — editing it changes the sha256 checksum stored
// in every existing database's schema_migrations and breaks forward
// compatibility (an upgraded binary would refuse to open the DB). A new schema
// change is a NEW entry in this list with the next monotonic version number
// (v4, v5, …), never an edit to initialSchemaUp or any other migration's SQL.
// Pre-consolidation databases are carried forward to this baseline by
// reconcileLegacyDB (see verifyAppliedMigrations).
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
			name:    "fts5_fulltext_search",
			up:      fts5SearchUp,
			down:    fts5SearchDown,
		},
		{
			version: 3,
			name:    "tier3_feature_extras",
			up:      tier3FeatureExtrasUp,
			down:    tier3FeatureExtrasDown,
		},
		{
			version: 4,
			name:    "cv_portfolio",
			up:      cvPortfolioUp,
			down:    cvPortfolioDown,
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

	legacy, err := d.verifyAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to verify applied migrations: %w", err)
	}

	if legacy {
		d.logger.Warn("Legacy migration lineage detected; reconciling schema and resetting migration log to consolidated baseline")
		if err := d.reconcileLegacyDB(); err != nil {
			return fmt.Errorf("failed to reconcile legacy database: %w", err)
		}
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
				// FTS5 is the only optional migration: SQLite may be built
				// without it. Detect by name rather than a version literal so
				// the consolidation above can renumber freely.
				if m.name == "fts5_fulltext_search" && !d.HasFTS5() {
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

// verifyAppliedMigrations inspects the stored migration history and reports
// whether the database is on the current migration lineage.
//
// It returns legacy=true when the stored history does NOT perfectly match the
// current embedded baseline — i.e. any stored version outside the current set
// (a pre-consolidation incremental DB recorded v4..v8) or a checksum mismatch on
// a current-version migration (the consolidation reused v1/v2/v3, so a legacy DB
// that recorded those numbers with different SQL trips this). In that case the
// caller (Migrate) reconciles the schema and resets the log instead of erroring,
// so a legacy database is carried forward in place.
//
// A fresh database (no rows) returns (false, nil) and is built by the apply loop.
// A current consolidated database whose checksums all match returns (false, nil).
//
// This deliberately does NOT hard-error on a checksum mismatch: a legacy DB
// stopped at an early version is indistinguishable from tampering by checksum
// alone, and bricking a local-first database on upgrade is the exact defect this
// fixes. Reconciliation is idempotent and self-healing; genuine structural
// corruption surfaces at runtime instead.
func (d *Database) verifyAppliedMigrations() (legacy bool, err error) {
	rows, err := d.db.Query("SELECT version, checksum FROM schema_migrations")
	if err != nil {
		// schema_migrations is created immediately before this runs, so an error
		// here means there is nothing applied yet to verify.
		return false, nil
	}
	defer rows.Close()

	applied := make(map[int]string)
	for rows.Next() {
		var version int
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return false, err
		}
		applied[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	if len(applied) == 0 {
		return false, nil // fresh database — apply loop builds it
	}

	migrations := getMigrations()
	currentVersions := make(map[int]bool, len(migrations))
	for _, m := range migrations {
		currentVersions[m.version] = true
	}

	// Any stored version outside the current set marks a legacy lineage.
	for v := range applied {
		if !currentVersions[v] {
			return true, nil
		}
	}

	// All stored versions are within the current set; a checksum mismatch on any
	// of them also marks a legacy lineage (consolidation reused v1/v2/v3).
	for _, m := range migrations {
		if stored, ok := applied[m.version]; ok {
			// The FTS5 migration is recorded with the same checksum whether it
			// ran or was skipped, so this comparison holds in both cases.
			if stored != calculateChecksum(m.up) {
				return true, nil
			}
		}
	}

	return false, nil
}

// reconcileLegacyDB brings a pre-consolidation database forward to the current
// baseline in a single atomic transaction: it adds any missing additive columns,
// rebuilds the dependencies table if it has the wrong (legacy bootstrap) shape,
// and resets schema_migrations to the embedded baseline with current checksums.
//
// The whole operation commits together, so on any failure the database is left
// untouched. It is idempotent — safe to run on an already-current or
// partially-reconciled database.
func (d *Database) reconcileLegacyDB() error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin reconciliation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := d.addMissingColumns(tx); err != nil {
		return fmt.Errorf("add missing columns: %w", err)
	}
	if err := d.rebuildDependenciesIfWrongShape(tx); err != nil {
		return fmt.Errorf("rebuild dependencies table: %w", err)
	}
	if err := d.resetMigrationLogToBaseline(tx); err != nil {
		return fmt.Errorf("reset migration log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reconciliation: %w", err)
	}
	return nil
}

// addMissingColumns adds any consolidated columns missing from legacy tables.
// All additions are additive (nullable or with a default), so existing rows are
// preserved. Each ALTER is gated on PRAGMA table_info so it is idempotent —
// SQLite has no ADD COLUMN IF NOT EXISTS.
func (d *Database) addMissingColumns(tx *sql.Tx) error {
	type colSpec struct {
		table, name, def string
	}
	// The consolidated schema's additive columns over the legacy incremental
	// lineage. Mirrors tier3FeatureExtrasUp (features +3) and the ADR-017
	// deterministic signals folded into initialSchemaUp's metadata block.
	specs := []colSpec{
		// tier-3 feature deep-dive (consolidated migration v3 "tier3_feature_extras")
		{"features", "implementation_status", "TEXT DEFAULT 'planned'"},
		{"features", "feature_architecture", "TEXT"},
		{"features", "pattern", "TEXT"},
		// ADR-017 deterministic importance signals on metadata (legacy v8
		// "metadata_extras"); added for pre-v8 databases, skipped for v8.
		{"metadata", "first_commit_at", "TIMESTAMP"},
		{"metadata", "commit_velocity_90d", "INTEGER DEFAULT 0"},
		{"metadata", "contributor_count", "INTEGER DEFAULT 0"},
		{"metadata", "tag_count", "INTEGER DEFAULT 0"},
		{"metadata", "remote_url", "TEXT"},
		{"metadata", "is_published", "INTEGER DEFAULT 0"},
		{"metadata", "maturity_score", "INTEGER DEFAULT 0"},
		{"metadata", "maturity_indicators", "TEXT"},
		{"metadata", "capabilities_summary", "TEXT"},
	}

	for _, s := range specs {
		exists, err := columnExists(tx, s.table, s.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", s.table, s.name, s.def)
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to add %s.%s: %w", s.table, s.name, err)
		}
		d.logger.Info("Reconciliation: added missing column",
			models.Field{Key: "table", Value: s.table},
			models.Field{Key: "column", Value: s.name},
		)
	}
	return nil
}

// rebuildDependenciesIfWrongShape recreates the dependencies table if it lacks
// the consolidated `manager` column. The wrong shape originates from an
// uncommitted pre-history bootstrap (a `type` column + TEXT primary key appear in
// no committed migration); its exact schema varies across installs, so a
// best-effort data copy would be fragile. Dependencies are deterministic facts
// the engine re-extracts (DetectDependencies), so the table is dropped and
// recreated empty — the next scan repopulates it. dependencies is referenced by
// nothing (no foreign key points at dependencies.id), so the DROP is safe under
// foreign_keys=on. If `manager` is already present the table is left untouched.
func (d *Database) rebuildDependenciesIfWrongShape(tx *sql.Tx) error {
	managerPresent, err := columnExists(tx, "dependencies", "manager")
	if err != nil {
		return err
	}
	if managerPresent {
		return nil // already the consolidated shape
	}

	d.logger.Warn("Reconciliation: dependencies table has legacy shape; rebuilding (run a scan to repopulate)")

	if _, err := tx.Exec("DROP TABLE IF EXISTS dependencies;"); err != nil {
		return fmt.Errorf("drop legacy dependencies: %w", err)
	}
	if _, err := tx.Exec(dependenciesTableDef); err != nil {
		return fmt.Errorf("recreate dependencies: %w", err)
	}
	if _, err := tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_dependencies_project_id ON dependencies(project_id);
		CREATE INDEX IF NOT EXISTS idx_dependencies_name ON dependencies(name);
	`); err != nil {
		return fmt.Errorf("recreate dependencies indexes: %w", err)
	}
	return nil
}

// resetMigrationLogToBaseline replaces schema_migrations with the embedded
// baseline (getMigrations — v1/v2/v3) recorded with current checksums, so a
// reconciled database verifies cleanly and GetSchemaVersion reports 3. Uses the
// embedded set only (not loadMigrations) so any future v4+ file migration is
// left for the apply loop to run normally. The FTS5 row is recorded with
// calculateChecksum(fts5SearchUp), matching what recordMigration stores on the
// skip path, so it verifies whether or not FTS5 was ever built.
func (d *Database) resetMigrationLogToBaseline(tx *sql.Tx) error {
	if _, err := tx.Exec("DELETE FROM schema_migrations;"); err != nil {
		return fmt.Errorf("clear migration records: %w", err)
	}
	now := time.Now().UTC()
	for _, m := range getMigrations() {
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, ?, ?)",
			m.version, m.name, now, calculateChecksum(m.up),
		); err != nil {
			return fmt.Errorf("record baseline migration %d: %w", m.version, err)
		}
	}
	return nil
}

// columnExists reports whether a column exists on a table, via PRAGMA table_info.
func columnExists(tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("inspect %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("scan pragma row for %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
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

func (d *Database) runMigration(m migration) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.up); err != nil {
		return fmt.Errorf("migration SQL failed: %w", err)
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

// initialSchemaUp is the consolidated schema. It creates every table, column,
// and index in its final form. All statements are IF NOT EXISTS, so re-running
// Migrate() on an already-current database is a no-op via the version check in
// Migrate() — and even a direct re-run of this migration is harmless.
//
// FROZEN (ADR-022): do NOT edit this SQL. Its sha256 checksum is recorded in
// every existing database's schema_migrations; editing it breaks forward
// compatibility for all of them. A schema change is a new versioned migration
// (v4, v5, …), never an edit here.
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

-- Metadata table. Includes all deterministic importance signals from ADR-017
-- (git investment, maturity, capabilities) as first-class columns rather than
-- later ALTER additions.
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
	last_scan_at TIMESTAMP,
	first_commit_at TIMESTAMP,
	commit_velocity_90d INTEGER DEFAULT 0,
	contributor_count INTEGER DEFAULT 0,
	tag_count INTEGER DEFAULT 0,
	remote_url TEXT,
	is_published INTEGER DEFAULT 0,
	maturity_score INTEGER DEFAULT 0,
	maturity_indicators TEXT,
	capabilities_summary TEXT,
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

-- Analyses table. Includes the AI-analysis extra columns (maturity, strengths,
-- weaknesses, reusable_components) as first-class columns.
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
	maturity TEXT,
	strengths TEXT,
	weaknesses TEXT,
	reusable_components TEXT,
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

-- Dependencies table. The prod/dev scope column (ADR-017 maturity signal) is a
-- first-class column defaulting to "prod". version + version_type record the
-- declared version spec as a literal fact (value + constraint kind); whether it
-- is "outdated" is an agent-computed indicator, not stored. The UNIQUE
-- constraint is (project_id, name, manager) without scope, so a package declared
-- in both dependencies and devDependencies collapses to one row;
-- DetectDependencies sorts prod before dev so the surviving row is deterministic.
CREATE TABLE IF NOT EXISTS dependencies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	manager TEXT NOT NULL,
	scope TEXT NOT NULL DEFAULT 'prod',
	version TEXT NOT NULL DEFAULT '',
	version_type TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(project_id, name, manager)
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_projects_root_path ON projects(root_path);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_documents_project_kind ON documents(project_id, kind);
CREATE INDEX IF NOT EXISTS idx_documents_path ON documents(project_id, path);
CREATE INDEX IF NOT EXISTS idx_documents_kind ON documents(kind);
CREATE INDEX IF NOT EXISTS idx_analyses_project_id ON analyses(project_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_analyses_project_analyzer ON analyses(project_id, analyzer);
CREATE INDEX IF NOT EXISTS idx_features_analysis_id ON features(analysis_id);
CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_project);
CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_project);
CREATE INDEX IF NOT EXISTS idx_metadata_language ON metadata(language_summary);
CREATE INDEX IF NOT EXISTS idx_metadata_framework ON metadata(framework_summary);
CREATE INDEX IF NOT EXISTS idx_dependencies_project_id ON dependencies(project_id);
CREATE INDEX IF NOT EXISTS idx_dependencies_name ON dependencies(name);
`

const initialSchemaDown = `
DROP INDEX IF EXISTS idx_dependencies_name;
DROP INDEX IF EXISTS idx_dependencies_project_id;
DROP INDEX IF EXISTS idx_metadata_framework;
DROP INDEX IF EXISTS idx_metadata_language;
DROP INDEX IF EXISTS idx_relationships_target;
DROP INDEX IF EXISTS idx_relationships_source;
DROP INDEX IF EXISTS idx_features_analysis_id;
DROP INDEX IF EXISTS idx_analyses_project_analyzer;
DROP INDEX IF EXISTS idx_analyses_project_id;
DROP INDEX IF EXISTS idx_documents_kind;
DROP INDEX IF EXISTS idx_documents_path;
DROP INDEX IF EXISTS idx_documents_project_kind;
DROP INDEX IF EXISTS idx_documents_project_id;
DROP INDEX IF EXISTS idx_projects_name;
DROP INDEX IF EXISTS idx_projects_root_path;
DROP TABLE IF EXISTS dependencies;
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

// dependenciesTableDef is the consolidated dependencies DDL, duplicated from the
// dependencies block in initialSchemaUp. It is used by reconcileLegacyDB to
// recreate a legacy-shape dependencies table. It is intentionally duplicated
// rather than concatenated into initialSchemaUp so that initialSchemaUp's bytes
// (and thus its checksum) are never put at risk. Keep this byte-identical to the
// CREATE TABLE block in initialSchemaUp.
const dependenciesTableDef = `CREATE TABLE IF NOT EXISTS dependencies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	manager TEXT NOT NULL,
	scope TEXT NOT NULL DEFAULT 'prod',
	version TEXT NOT NULL DEFAULT '',
	version_type TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(project_id, name, manager)
);`

// fts5SearchUp is FROZEN (ADR-022): editing it changes its checksum and breaks
// every database that recorded it. New schema changes are new versioned migrations.
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

// tier3FeatureExtrasUp is FROZEN (ADR-022): editing it changes its checksum and
// breaks every database that recorded it. New schema changes are new versioned migrations.
const tier3FeatureExtrasUp = `
ALTER TABLE features ADD COLUMN implementation_status TEXT DEFAULT 'planned';
ALTER TABLE features ADD COLUMN feature_architecture TEXT;
ALTER TABLE features ADD COLUMN pattern TEXT;
`

const tier3FeatureExtrasDown = `-- No rollback for ALTER TABLE ADD COLUMN`

// cvPortfolioUp creates tables for the CV Builder feature.
const cvPortfolioUp = `
CREATE TABLE IF NOT EXISTS cv_portfolios (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  user_id TEXT NOT NULL DEFAULT 'default',
  summary TEXT,
  target_roles TEXT,
  industry_focus TEXT,
  preferred_locations TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_experiences (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  company TEXT NOT NULL,
  position TEXT NOT NULL,
  location TEXT,
  start_date TEXT NOT NULL,
  end_date TEXT,
  employment_type TEXT,
  description TEXT,
  key_responsibilities TEXT,
  technologies_used TEXT,
  team_size INTEGER,
  reporting_to TEXT,
  is_current INTEGER DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_achievements (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  experience_id TEXT REFERENCES cv_experiences(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  impact TEXT,
  metrics TEXT,
  skills_used TEXT,
  category TEXT,
  relevance_score REAL DEFAULT 0.5,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_skills (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT,
  proficiency TEXT,
  years_of_experience INTEGER,
  last_used TEXT,
  is_highlight INTEGER DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_education (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  institution TEXT NOT NULL,
  degree TEXT,
  field_of_study TEXT,
  start_date TEXT,
  end_date TEXT,
  gpa REAL,
  honors TEXT,
  relevant_coursework TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_certifications (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  issuer TEXT,
  issue_date TEXT,
  expiry_date TEXT,
  credential_id TEXT,
  credential_url TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS cv_generated (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  template_id TEXT,
  job_description TEXT,
  content TEXT NOT NULL,
  markdown_content TEXT,
  ats_score REAL,
  tailoring_notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_cv_portfolios_user ON cv_portfolios(user_id);
CREATE INDEX IF NOT EXISTS idx_cv_experiences_portfolio ON cv_experiences(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_cv_achievements_portfolio ON cv_achievements(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_cv_achievements_experience ON cv_achievements(experience_id);
CREATE INDEX IF NOT EXISTS idx_cv_skills_portfolio ON cv_skills(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_cv_education_portfolio ON cv_education(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_cv_certifications_portfolio ON cv_certifications(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_cv_generated_portfolio ON cv_generated(portfolio_id);
`

const cvPortfolioDown = `
DROP TABLE IF EXISTS cv_generated;
DROP TABLE IF EXISTS cv_certifications;
DROP TABLE IF EXISTS cv_education;
DROP TABLE IF EXISTS cv_skills;
DROP TABLE IF EXISTS cv_achievements;
DROP TABLE IF EXISTS cv_experiences;
DROP TABLE IF EXISTS cv_portfolios;
`
