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

// Dependency represents a project dependency
type Dependency struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Manager   string `json:"manager"`
}

// Configuration represents system configuration
type Configuration struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}
