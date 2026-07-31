package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/logging"
)

// setupTestEnvironment sets up environment for testing
func setupTestEnvironment(t *testing.T) func() {
	// Set test database key for password protection
	oldKey := os.Getenv("PORTFOLIO_DB_KEY")
	os.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

	return func() {
		// Restore original value
		if oldKey != "" {
			os.Setenv("PORTFOLIO_DB_KEY", oldKey)
		} else {
			os.Unsetenv("PORTFOLIO_DB_KEY")
		}
	}
}

func TestNewDatabase(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	cleanup := setupTestEnvironment(t)
	defer cleanup()

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

	if version < 2 {
		t.Errorf("Expected schema version >= 2 (consolidated schema fully applied), got %d", version)
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
	cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

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

func TestMigrationConsolidatedSchema(t *testing.T) {
	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

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

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrations failed: %v", err)
	}

	// Re-running Migrate must be a no-op: the version check in Migrate() skips
	// migrations already recorded, and the consolidated schema uses IF NOT EXISTS.
	if err := db.Migrate(); err != nil {
		t.Fatalf("Re-running migrations failed (not idempotent): %v", err)
	}

	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if version != 4 {
		t.Errorf("schema version: got %d, want 4 (initial_schema + fts5_fulltext_search + tier3_feature_extras + cv_portfolio)", version)
	}

	metaCols := tableColumns(t, db, "metadata")
	for _, c := range []string{
		"first_commit_at", "commit_velocity_90d", "contributor_count", "tag_count",
		"remote_url", "is_published", "maturity_score", "maturity_indicators",
		"capabilities_summary",
	} {
		if !metaCols[c] {
			t.Errorf("metadata column %q missing after migration", c)
		}
	}

	if depCols := tableColumns(t, db, "dependencies"); !depCols["scope"] {
		t.Error("dependencies column \"scope\" missing after migration")
	}
}

func tableColumns(t *testing.T, db *Database, table string) map[string]bool {
	t.Helper()
	rows, err := db.DB().Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma row: %v", err)
		}
		out[name] = true
	}
	return out
}

func TestDatabasePermissions(t *testing.T) {
	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

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

// --- Legacy forward-compatibility (reconciliation) tests ---------------------

// seedLegacyDB writes a pre-consolidation database into db: the full legacy
// table set with legacy shapes (features without the tier-3 columns; metadata
// without the ADR-017 columns; a wrong-shape dependencies table carrying a
// `type` column and TEXT primary key), a schema_migrations log claiming v1..v8
// with bogus checksums and empty names for v7/v8, plus sample rows in projects
// and analyses (with a sentinel summary) and a dangling + duplicate dependency
// row. This mimics the real failing database so the reconciliation path is
// exercised end to end.
func seedLegacyDB(t *testing.T, db *sql.DB) {
	t.Helper()
	const legacySchema = `
CREATE TABLE projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	root_path TEXT NOT NULL UNIQUE,
	repository_type TEXT NOT NULL,
	discovered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE metadata (
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
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE TABLE documents (
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
CREATE TABLE analyses (
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
CREATE TABLE features (
	id TEXT PRIMARY KEY,
	analysis_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	confidence REAL,
	FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
);
CREATE TABLE technologies (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	category TEXT
);
CREATE TABLE project_technologies (
	project_id TEXT NOT NULL,
	technology_id TEXT NOT NULL,
	PRIMARY KEY (project_id, technology_id),
	FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (technology_id) REFERENCES technologies(id) ON DELETE CASCADE
);
CREATE TABLE relationships (
	id TEXT PRIMARY KEY,
	source_project TEXT NOT NULL,
	target_project TEXT NOT NULL,
	type TEXT NOT NULL,
	description TEXT,
	confidence REAL,
	FOREIGN KEY (source_project) REFERENCES projects(id) ON DELETE CASCADE,
	FOREIGN KEY (target_project) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE TABLE configuration (
	key TEXT PRIMARY KEY,
	value TEXT,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Wrong-shape dependencies from an uncommitted pre-history bootstrap: a "type"
-- column and TEXT primary key, with no manager/version_type/created_at/UNIQUE/FK.
CREATE TABLE dependencies (
	id TEXT PRIMARY KEY,
	project_id TEXT,
	name TEXT,
	version TEXT,
	type TEXT,
	scope TEXT DEFAULT 'prod'
);
CREATE TABLE schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TIMESTAMP NOT NULL,
	checksum TEXT NOT NULL
);
`
	mustExec := func(sql string, args ...interface{}) {
		t.Helper()
		if _, err := db.Exec(sql, args...); err != nil {
			t.Fatalf("seed exec failed (%s): %v", sql, err)
		}
	}
	mustExec(legacySchema)

	mustExec(`INSERT INTO projects (id, name, root_path, repository_type) VALUES ('proj-a','Alpha','/tmp/alpha','git')`)
	mustExec(`INSERT INTO projects (id, name, root_path, repository_type) VALUES ('proj-b','Beta','/tmp/beta','git')`)
	mustExec(`INSERT INTO analyses (id, project_id, analyzer, analyzed_git_head, summary) VALUES ('an-1','proj-a','claude','deadbeef','AI-SENTINEL-SUMMARY')`)

	// A real dependency row, a dangling-project_id row, and a duplicate. The
	// rebuild drops all of them — dependencies are re-derived by detection.
	mustExec(`INSERT INTO dependencies (id, project_id, name, version, type, scope) VALUES ('d1','proj-a','react','18.0.0','npm','prod')`)
	mustExec(`INSERT INTO dependencies (id, project_id, name, version, type, scope) VALUES ('d2','proj-missing','lodash','4.0.0','npm','prod')`)
	mustExec(`INSERT INTO dependencies (id, project_id, name, version, type, scope) VALUES ('d3','proj-a','react','18.0.0','npm','dev')`)

	// Legacy migration log: v1..v8, bogus checksums, empty names for v7/v8
	// (mimicking the historical backfill-switch bug).
	for v := 1; v <= 8; v++ {
		name := fmt.Sprintf("legacy-%d", v)
		if v >= 7 {
			name = ""
		}
		mustExec(`INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES (?, ?, CURRENT_TIMESTAMP, ?)`,
			v, name, fmt.Sprintf("legacy-checksum-%d", v))
	}
}

func newLegacyTestDB(t *testing.T) *Database {
	t.Helper()

	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

	dir := t.TempDir()
	logger, _ := logging.NewLogger("INFO", "console")
	db, err := NewDatabase(filepath.Join(dir, "legacy.db"), logger)
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	if err := db.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	seedLegacyDB(t, db.DB())
	return db
}

func TestMigrate_LegacyDBUpgrade(t *testing.T) {
	db := newLegacyTestDB(t)

	if err := db.Initialize(); err != nil {
		t.Fatalf("Initialize on legacy DB failed: %v", err)
	}

	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if version != 4 {
		t.Fatalf("schema version: got %d, want 4", version)
	}

	// schema_migrations is exactly {1,2,3} with current checksums.
	rows, err := db.DB().Query("SELECT version, checksum FROM schema_migrations")
	if err != nil {
		t.Fatalf("query migrations: %v", err)
	}
	got := make(map[int]string)
	for rows.Next() {
		var v int
		var c string
		if err := rows.Scan(&v, &c); err != nil {
			t.Fatalf("scan migration row: %v", err)
		}
		got[v] = c
	}
	if err := rows.Close(); err != nil {
		t.Errorf("close rows: %v", err)
	}
	for _, m := range getMigrations() {
		c, ok := got[m.version]
		if !ok {
			t.Errorf("migration %d missing from schema_migrations after reconcile", m.version)
			continue
		}
		if c != calculateChecksum(m.up) {
			t.Errorf("migration %d checksum not healed: got %s, want %s", m.version, c, calculateChecksum(m.up))
		}
	}
	if len(got) != len(getMigrations()) {
		t.Errorf("schema_migrations row count: got %d, want %d", len(got), len(getMigrations()))
	}

	// Additive columns added.
	featCols := tableColumns(t, db, "features")
	for _, c := range []string{"implementation_status", "feature_architecture", "pattern"} {
		if !featCols[c] {
			t.Errorf("features column %q missing after reconcile", c)
		}
	}
	metaCols := tableColumns(t, db, "metadata")
	for _, c := range []string{
		"first_commit_at", "commit_velocity_90d", "contributor_count", "tag_count",
		"remote_url", "is_published", "maturity_score", "maturity_indicators",
		"capabilities_summary",
	} {
		if !metaCols[c] {
			t.Errorf("metadata column %q missing after reconcile", c)
		}
	}

	// dependencies rebuilt to the consolidated shape, empty.
	depCols := tableColumns(t, db, "dependencies")
	for _, c := range []string{"manager", "version_type", "scope", "created_at"} {
		if !depCols[c] {
			t.Errorf("dependencies column %q missing after rebuild", c)
		}
	}
	if depCols["type"] {
		t.Error("dependencies still has legacy `type` column after rebuild")
	}
	var depCount int
	if err := db.DB().QueryRow("SELECT count(*) FROM dependencies").Scan(&depCount); err != nil {
		t.Fatalf("count dependencies: %v", err)
	}
	if depCount != 0 {
		t.Errorf("dependencies row count after rebuild: got %d, want 0 (re-derived by scan)", depCount)
	}

	// Data preserved: projects + analyses (the AI-generated knowledge).
	var projCount int
	if err := db.DB().QueryRow("SELECT count(*) FROM projects").Scan(&projCount); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projCount != 2 {
		t.Errorf("projects preserved: got %d, want 2", projCount)
	}
	var summary string
	if err := db.DB().QueryRow("SELECT summary FROM analyses WHERE id='an-1'").Scan(&summary); err != nil {
		t.Fatalf("read analysis summary: %v", err)
	}
	if summary != "AI-SENTINEL-SUMMARY" {
		t.Errorf("analysis summary not preserved: got %q", summary)
	}

	if err := db.ValidateSchema(); err != nil {
		t.Errorf("ValidateSchema after reconcile: %v", err)
	}
}

func TestMigrate_LegacyDBUpgrade_Idempotent(t *testing.T) {
	db := newLegacyTestDB(t)

	if err := db.Initialize(); err != nil {
		t.Fatalf("first Initialize: %v", err)
	}
	// Re-initializing a reconciled DB must be a no-op.
	if err := db.Initialize(); err != nil {
		t.Fatalf("second Initialize: %v", err)
	}
	// A direct reconcileLegacyDB call must also be safe (every step a no-op on a
	// reconciled DB, except resetting the baseline log to the same values).
	if err := db.reconcileLegacyDB(); err != nil {
		t.Fatalf("direct reconcileLegacyDB after reconcile: %v", err)
	}

	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if version != 4 {
		t.Fatalf("schema version: got %d, want 4", version)
	}
	depCols := tableColumns(t, db, "dependencies")
	if !depCols["manager"] || depCols["type"] {
		t.Errorf("dependencies shape not stable after repeat: manager=%v type=%v", depCols["manager"], depCols["type"])
	}
	var projCount int
	if err := db.DB().QueryRow("SELECT count(*) FROM projects").Scan(&projCount); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projCount != 2 {
		t.Errorf("projects not preserved across re-init: got %d, want 2", projCount)
	}
	var summary string
	if err := db.DB().QueryRow("SELECT summary FROM analyses WHERE id='an-1'").Scan(&summary); err != nil {
		t.Fatalf("read analysis summary: %v", err)
	}
	if summary != "AI-SENTINEL-SUMMARY" {
		t.Errorf("analysis summary changed across re-init: %q", summary)
	}
}

func TestMigrate_FreshDBUnchanged(t *testing.T) {
	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

	dir := t.TempDir()
	logger, _ := logging.NewLogger("INFO", "console")
	db, err := NewDatabase(filepath.Join(dir, "fresh.db"), logger)
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	if err := db.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Initialize(); err != nil {
		t.Fatalf("Initialize fresh DB: %v", err)
	}
	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion: %v", err)
	}
	if version != 4 {
		t.Fatalf("fresh schema version: got %d, want 4", version)
	}
	if !tableColumns(t, db, "dependencies")["manager"] {
		t.Error("fresh dependencies missing manager column")
	}
	// The fresh path never renames; no leftover table should exist.
	var n int
	if err := db.DB().QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='dependencies_legacy'").Scan(&n); err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if n != 0 {
		t.Error("fresh DB should not have a dependencies_legacy table")
	}
}

func TestMigrate_ChecksumMismatch_Heals(t *testing.T) {
	// Set database key for test environment
	t.Setenv("PORTFOLIO_DB_KEY", "test-database-key-for-testing")

	dir := t.TempDir()
	logger, _ := logging.NewLogger("INFO", "console")
	db, err := NewDatabase(filepath.Join(dir, "tamper.db"), logger)
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	if err := db.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	// Corrupt a recorded checksum — a legacy DB that reused v1's number with
	// different SQL is indistinguishable from this by checksum alone. Migrate
	// must heal it rather than abort.
	if _, err := db.DB().Exec("UPDATE schema_migrations SET checksum='bogus' WHERE version=1"); err != nil {
		t.Fatalf("corrupt checksum: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate after checksum corruption should heal, got: %v", err)
	}

	var got string
	if err := db.DB().QueryRow("SELECT checksum FROM schema_migrations WHERE version=1").Scan(&got); err != nil {
		t.Fatalf("read checksum: %v", err)
	}
	want := calculateChecksum(initialSchemaUp)
	if got != want {
		t.Errorf("checksum not healed: got %s, want %s", got, want)
	}
}
