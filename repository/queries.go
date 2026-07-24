package repository

import "fmt"

const (
	// Analysis queries
	queryCreateAnalysis = `
		INSERT INTO analyses (
			id, project_id, analyzer, analyzed_git_head, analyzed_at,
			summary, purpose, architecture, maturity, strengths, weaknesses,
			reusable_components, notes, raw_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	queryUpdateAnalysis = `
		UPDATE analyses 
		SET summary = ?, purpose = ?, architecture = ?, maturity = ?, 
		    strengths = ?, weaknesses = ?, reusable_components = ?, notes = ?, 
		    raw_json = ?, analyzed_git_head = ?, analyzed_at = ?, updated_at = ?
		WHERE project_id = ? AND analyzer = ?
	`

	queryGetAnalysis = `
		SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
		       summary, purpose, architecture, maturity, strengths, weaknesses,
		       reusable_components, notes, raw_json, created_at, updated_at
		FROM analyses
		WHERE project_id = ?
		ORDER BY analyzed_at DESC
		LIMIT 1
	`

	queryGetAnalysisByAnalyzer = `
		SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
		       summary, purpose, architecture, maturity, strengths, weaknesses,
		       reusable_components, notes, raw_json, created_at, updated_at
		FROM analyses
		WHERE project_id = ? AND analyzer = ?
	`

	queryDeleteAnalysis = `
		DELETE FROM analyses
		WHERE project_id = ? AND analyzer = ?
	`

	// Feature queries
	queryCreateFeatures = `
		INSERT INTO features (id, analysis_id, name, description, confidence, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	queryDeleteFeaturesByAnalysisID = `
		DELETE FROM features
		WHERE analysis_id = ?
	`

	queryGetFeaturesByAnalysisID = `
		SELECT id, analysis_id, name, description, confidence, created_at
		FROM features
		WHERE analysis_id = ?
		ORDER BY name
	`

	// Project and metadata queries
	queryProjectExists = `
		SELECT EXISTS(
			SELECT 1 FROM projects WHERE id = ?
		)
	`

	queryGetGitHeadForProject = `
		SELECT git_head
		FROM metadata
		WHERE project_id = ?
	`

	queryListAllAnalyses = `
		SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
		       summary, purpose, architecture, maturity, strengths, weaknesses,
		       reusable_components, notes, raw_json, created_at, updated_at
		FROM analyses
		ORDER BY analyzed_at DESC
	`

	// Relationship queries
	queryCreateRelationship = `
		INSERT INTO relationships (id, source_project, target_project, type, description, confidence, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	queryUpdateRelationship = `
		UPDATE relationships 
		SET description = ?, confidence = ?, updated_at = ?
		WHERE id = ?
	`

	queryGetRelationship = `
		SELECT id, source_project, target_project, type, description, confidence, created_at, updated_at
		FROM relationships
		WHERE id = ?
	`

	queryListRelationshipsByProject = `
		SELECT id, source_project, target_project, type, description, confidence, created_at, updated_at
		FROM relationships
		WHERE source_project = ? OR target_project = ?
		ORDER BY type, created_at
	`

	queryDeleteRelationship = `
		DELETE FROM relationships
		WHERE id = ?
	`

	queryFindExistingRelationship = `
		SELECT id, source_project, target_project, type, description, confidence, created_at, updated_at
		FROM relationships
		WHERE source_project = ? AND target_project = ? AND type = ?
	`

	// Stale detection query
	queryListProjectsNeedingAnalysis = `
		SELECT p.id, p.name,
		       CASE 
		         WHEN a.id IS NULL THEN 'unanalyzed'
		         WHEN m.git_head IS NULL OR m.git_head != a.analyzed_git_head THEN 'outdated'
		       END as reason
		FROM projects p
		LEFT JOIN metadata m ON m.project_id = p.id
		LEFT JOIN analyses a ON a.project_id = p.id
		WHERE a.id IS NULL 
		   OR m.git_head IS NULL 
		   OR m.git_head != a.analyzed_git_head
	`
)

// queryListProjectsNeedingAnalysis is defined above but we need to add it to the set
// Since it's already defined, we just add it to the list of available queries
var _ = fmt.Sprintf(queryListProjectsNeedingAnalysis)
