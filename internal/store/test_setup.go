package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type testStore struct {
	db            *sql.DB
	projects      *ProjectStore
	metadata      *MetadataStore
	documents     *DocumentStore
	analyses      *AnalysisStore
	features      *FeatureStore
	technologies  *TechnologyStore
	relationships *RelationshipStore
	dependencies  *DependencyStore
	configuration *ConfigurationStore
}

func setupTestStore(t *testing.T) *testStore {
	t.Helper()

	tempFile := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite3", tempFile)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	logger := zap.NewNop()
	migrateTestDB(t, db)

	return &testStore{
		db:            db,
		projects:      NewProjectStore(db, logger),
		metadata:      NewMetadataStore(db, logger),
		documents:     NewDocumentStore(db, logger),
		analyses:      NewAnalysisStore(db, logger),
		features:      NewFeatureStore(db, logger),
		technologies:  NewTechnologyStore(db, logger),
		relationships: NewRelationshipStore(db, logger),
		dependencies:  NewDependencyStore(db, logger),
		configuration: NewConfigurationStore(db, logger),
	}
}

func cleanupTestStore(t *testing.T, store *testStore) {
	t.Helper()
	store.db.Close()
}

func migrateTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	// Run initial schema
	_, err := db.Exec(`
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
			language_summary TEXT,
			framework_summary TEXT,
			dependency_summary TEXT,
			documentation_hash TEXT,
			last_scan_at TIMESTAMP,
			last_modified_at TIMESTAMP,
			commit_count INTEGER DEFAULT 0,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
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
		
		CREATE TABLE IF NOT EXISTS analyses (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			analyzer TEXT NOT NULL,
			analyzed_git_head TEXT NOT NULL,
			analyzed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			summary TEXT,
			purpose TEXT,
			architecture TEXT,
			maturity TEXT,
			strengths TEXT,
			weaknesses TEXT,
			reusable_components TEXT,
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
			implementation_status TEXT DEFAULT 'planned',
			feature_architecture TEXT,
			pattern TEXT,
			FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
		);
		
		CREATE TABLE IF NOT EXISTS technologies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			category TEXT
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
			value TEXT,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS dependencies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id TEXT NOT NULL,
			name TEXT NOT NULL,
			manager TEXT NOT NULL,
			scope TEXT NOT NULL DEFAULT 'prod',
			version TEXT NOT NULL DEFAULT '',
			version_type TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, name, manager),
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Run AI analysis schema migration
	_, err = db.Exec(`
		ALTER TABLE analyses ADD COLUMN maturity TEXT;
		ALTER TABLE analyses ADD COLUMN strengths TEXT;
		ALTER TABLE analyses ADD COLUMN weaknesses TEXT;
		ALTER TABLE analyses ADD COLUMN reusable_components TEXT;
	`)
	if err != nil {
		// Ignore errors for ALTER TABLE - columns might already exist
	}

	// Deterministic importance signals (mirrors migration v8: metadataExtrasUp).
	_, err = db.Exec(`
		ALTER TABLE metadata ADD COLUMN first_commit_at TIMESTAMP;
		ALTER TABLE metadata ADD COLUMN commit_velocity_90d INTEGER DEFAULT 0;
		ALTER TABLE metadata ADD COLUMN contributor_count INTEGER DEFAULT 0;
		ALTER TABLE metadata ADD COLUMN tag_count INTEGER DEFAULT 0;
		ALTER TABLE metadata ADD COLUMN remote_url TEXT;
		ALTER TABLE metadata ADD COLUMN is_published INTEGER DEFAULT 0;
		ALTER TABLE metadata ADD COLUMN maturity_score INTEGER DEFAULT 0;
		ALTER TABLE metadata ADD COLUMN maturity_indicators TEXT;
		ALTER TABLE metadata ADD COLUMN capabilities_summary TEXT;
	`)
	if err != nil {
		// Ignore errors for ALTER TABLE - columns might already exist
	}

	// Create unique index for analyses
	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_analyses_project_analyzer ON analyses(project_id, analyzer);
	`)
	if err != nil {
		t.Fatalf("failed to create analysis index: %v", err)
	}
}
