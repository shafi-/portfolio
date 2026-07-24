package analysis

import (
	"context"
	"github.com/google/uuid"
)

// AnalysisStore defines the interface for analysis persistence operations
type AnalysisStore interface {
	// Project existence check
	ProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error)

	// Analysis CRUD
	CreateAnalysis(ctx context.Context, analysis *Analysis) error
	UpdateAnalysis(ctx context.Context, analysis *Analysis) error
	GetAnalysis(ctx context.Context, projectID uuid.UUID) (*Analysis, error)
	GetAnalysisByAnalyzer(ctx context.Context, projectID uuid.UUID, analyzer string) (*Analysis, error)
	DeleteAnalysis(ctx context.Context, projectID uuid.UUID, analyzer string) error

	// Features CRUD
	CreateFeatures(ctx context.Context, features []Feature) error
	DeleteFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) error
	GetFeaturesByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]Feature, error)

	// Stale detection
	GetGitHeadForProject(ctx context.Context, projectID uuid.UUID) (*string, error)
	ListAllAnalyses(ctx context.Context) ([]Analysis, error)
	ListProjectsNeedingAnalysis(ctx context.Context) ([]NeedsAnalysisResult, error)

	// Relationships
	CreateRelationship(ctx context.Context, rel *Relationship) error
	UpdateRelationship(ctx context.Context, rel *Relationship) error
	GetRelationship(ctx context.Context, id uuid.UUID) (*Relationship, error)
	ListRelationshipsByProject(ctx context.Context, projectID uuid.UUID) ([]Relationship, error)
	DeleteRelationship(ctx context.Context, id uuid.UUID) error
	FindExistingRelationship(ctx context.Context, source, target uuid.UUID, relType string) (*Relationship, error)
}
