-- Migration 001: Add analysis tables
-- Story 10.2: Persist Analyses

-- Create analyses table
CREATE TABLE IF NOT EXISTS analyses (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    analyzer TEXT NOT NULL,
    analyzed_git_head TEXT NOT NULL,
    analyzed_at TEXT NOT NULL,
    summary TEXT,
    purpose TEXT,
    architecture TEXT,
    maturity TEXT,
    strengths TEXT,              -- JSON array
    weaknesses TEXT,             -- JSON array
    reusable_components TEXT,    -- JSON array
    notes TEXT,
    raw_json TEXT,               -- Full JSON payload
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, analyzer)
);

-- Create indexes for analyses table
CREATE INDEX IF NOT EXISTS idx_analyses_project_id ON analyses(project_id);
CREATE INDEX IF NOT EXISTS idx_analyses_analyzer ON analyses(analyzer);
CREATE INDEX IF NOT EXISTS idx_analyses_composite ON analyses(project_id, analyzer);
CREATE INDEX IF NOT EXISTS idx_analyses_analyzed_at ON analyses(analyzed_at);

-- Create features table
CREATE TABLE IF NOT EXISTS features (
    id TEXT PRIMARY KEY,
    analysis_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    confidence REAL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
);

-- Create indexes for features table
CREATE INDEX IF NOT EXISTS idx_features_analysis_id ON features(analysis_id);
CREATE INDEX IF NOT EXISTS idx_features_name ON features(name);

-- Create relationships table
CREATE TABLE IF NOT EXISTS relationships (
    id TEXT PRIMARY KEY,
    source_project TEXT NOT NULL,
    target_project TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    confidence REAL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (source_project) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (target_project) REFERENCES projects(id) ON DELETE CASCADE,
    CHECK(type IN ('Similar', 'Evolution', 'Shared Feature', 'Shared Technology', 'Reuses Component')),
    CHECK(confidence IS NULL OR (confidence >= 0.0 AND confidence <= 1.0)),
    UNIQUE(source_project, target_project, type)
);

-- Create indexes for relationships table
CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_project);
CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_project);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(type);
CREATE INDEX IF NOT EXISTS idx_relationships_composite ON relationships(source_project, target_project, type);