package models

// Project represents a discovered project
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	RootPath       string `json:"root_path"`
	RepositoryType string `json:"repository_type"`
	DiscoveredAt   string `json:"discovered_at"`
	UpdatedAt      string `json:"updated_at"`
}

// Metadata represents project metadata
type Metadata struct {
	ProjectID         string `json:"project_id"`
	GitHead           string `json:"git_head"`
	DefaultBranch     string `json:"default_branch"`
	LastCommitAt      string `json:"last_commit_at"`
	LastModifiedAt    string `json:"last_modified_at"`
	CommitCount       int    `json:"commit_count"`
	LanguageSummary   string `json:"language_summary"`
	FrameworkSummary  string `json:"framework_summary"`
	DependencySummary string `json:"dependency_summary"`
	DocumentationHash string `json:"documentation_hash"`
	LastScanAt        string `json:"last_scan_at"`
	// Deterministic importance signals (no LLM required).
	FirstCommitAt       string `json:"first_commit_at"`
	CommitVelocity90d   int    `json:"commit_velocity_90d"`
	ContributorCount    int    `json:"contributor_count"`
	TagCount            int    `json:"tag_count"`
	RemoteURL           string `json:"remote_url"`
	IsPublished         bool   `json:"is_published"`
	MaturityScore       int    `json:"maturity_score"`
	MaturityIndicators  string `json:"maturity_indicators"`
	CapabilitiesSummary string `json:"capabilities_summary"`
}

// Document represents indexed documentation
type Document struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Path        string `json:"path"`
	Kind        string `json:"kind"`
	Content     string `json:"content"`
	ContentHash string `json:"content_hash"`
	IndexedAt   string `json:"indexed_at"`
}

// Analysis represents AI analysis results
type Analysis struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	Analyzer           string `json:"analyzer"`
	AnalyzedGitHead    string `json:"analyzed_git_head"`
	AnalyzedAt         string `json:"analyzed_at"`
	Summary            string `json:"summary"`
	Purpose            string `json:"purpose"`
	Architecture       string `json:"architecture"`
	Maturity           string `json:"maturity"`
	Strengths          string `json:"strengths"`
	Weaknesses         string `json:"weaknesses"`
	ReusableComponents string `json:"reusable_components"`
	Notes              string `json:"notes"`
	RawJSON            string `json:"raw_json"`
}

// Feature represents extracted features
type Feature struct {
	ID          string  `json:"id"`
	AnalysisID  string  `json:"analysis_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// Technology represents technology reference
type Technology struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// ProjectTechnology represents project-technology relationship
type ProjectTechnology struct {
	ProjectID    string `json:"project_id"`
	TechnologyID string `json:"technology_id"`
}

// Relationship represents inter-project relationships
type Relationship struct {
	ID            string  `json:"id"`
	SourceProject string  `json:"source_project"`
	TargetProject string  `json:"target_project"`
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Confidence    float64 `json:"confidence"`
}

// Dependency represents a project dependency
type Dependency struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Manager     string `json:"manager"`
	Scope       string `json:"scope"`        // "prod" (default) or "dev"
	Version     string `json:"version"`      // declared version value (operator stripped)
	VersionType string `json:"version_type"` // constraint kind: ^/~/~>/>=/==/exact/range/any, "" if unknown
}

// Configuration represents system configuration
type Configuration struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}
