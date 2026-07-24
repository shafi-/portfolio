package analysis

import (
	"time"
)

// Analysis represents a semantic analysis of a project
type Analysis struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"project_id"`
	Analyzer           string    `json:"analyzer"`
	AnalyzedGitHead    string    `json:"analyzed_git_head"`
	AnalyzedAt         time.Time `json:"analyzed_at"`
	Summary            string    `json:"summary"`
	Purpose            string    `json:"purpose"`
	Architecture       string    `json:"architecture"`
	Maturity           string    `json:"maturity"`
	Strengths          []string  `json:"strengths"`
	Weaknesses         []string  `json:"weaknesses"`
	ReusableComponents []string  `json:"reusable_components"`
	Notes              string    `json:"notes"`
	RawJSON            []byte    `json:"raw_json"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Features           []Feature `json:"features,omitempty"`
}

// AnalysisInput represents input for storing an analysis
type AnalysisInput struct {
	Summary            string         `json:"summary" validate:"required"`
	Purpose            string         `json:"purpose" validate:"required"`
	Architecture       string         `json:"architecture" validate:"required"`
	Maturity           string         `json:"maturity"`
	Strengths          []string       `json:"strengths"`
	Weaknesses         []string       `json:"weaknesses"`
	ReusableComponents []string       `json:"reusable_components"`
	Notes              string         `json:"notes"`
	AnalyzedAt         time.Time      `json:"analyzed_at" validate:"required"`
	AnalyzedGitHead    string         `json:"analyzed_git_head" validate:"required"`
	Analyzer           string         `json:"analyzer" validate:"required"`
	Features           []FeatureInput `json:"features"`
}

// Feature represents a capability implemented by a project
type Feature struct {
	ID          string    `json:"id"`
	AnalysisID  string    `json:"analysis_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Confidence  *float64  `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeatureInput represents input for storing a feature
type FeatureInput struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Confidence  *float64 `json:"confidence"`
}

// Relationship represents a connection between two projects
type Relationship struct {
	ID            string    `json:"id"`
	SourceProject string    `json:"source_project"`
	TargetProject string    `json:"target_project"`
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	Confidence    *float64  `json:"confidence"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NeedsAnalysisResult represents a project that needs analysis
type NeedsAnalysisResult struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Reason    string `json:"reason"` // "unanalyzed" or "outdated"
}
