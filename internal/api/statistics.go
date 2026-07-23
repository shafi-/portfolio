package api

import "net/http"

type statistics struct {
	TotalProjects         int            `json:"total_projects"`
	ProjectsWithMetadata  int            `json:"projects_with_metadata"`
	ProjectsWithAnalysis  int            `json:"projects_with_analysis"`
	ProjectsWithDocuments int            `json:"projects_with_documents"`
	TotalDocuments        int            `json:"total_documents"`
	TotalDependencies     int            `json:"total_dependencies"`
	TotalRelationships    int            `json:"total_relationships"`
	LanguageCounts        map[string]int `json:"language_counts"`
	FrameworkCounts       map[string]int `json:"framework_counts"`
	TechnologyCounts      map[string]int `json:"technology_counts"`
	DocumentKindCounts    map[string]int `json:"document_kind_counts"`
}

func (s *Server) handleStatistics(w http.ResponseWriter, r *http.Request) {
	stats := statistics{
		LanguageCounts:     make(map[string]int),
		FrameworkCounts:    make(map[string]int),
		TechnologyCounts:   make(map[string]int),
		DocumentKindCounts: make(map[string]int),
	}

	s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&stats.TotalProjects)
	s.db.QueryRow("SELECT COUNT(*) FROM metadata").Scan(&stats.ProjectsWithMetadata)
	s.db.QueryRow("SELECT COUNT(DISTINCT project_id) FROM analyses").Scan(&stats.ProjectsWithAnalysis)
	s.db.QueryRow("SELECT COUNT(DISTINCT project_id) FROM documents").Scan(&stats.ProjectsWithDocuments)
	s.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&stats.TotalDocuments)
	s.db.QueryRow("SELECT COUNT(*) FROM dependencies").Scan(&stats.TotalDependencies)
	s.db.QueryRow("SELECT COUNT(*) FROM relationships").Scan(&stats.TotalRelationships)

	langRows, err := s.db.Query("SELECT language_summary, COUNT(*) FROM metadata WHERE language_summary IS NOT NULL GROUP BY language_summary")
	if err == nil {
		defer langRows.Close()
		for langRows.Next() {
			var lang string
			var count int
			langRows.Scan(&lang, &count)
			stats.LanguageCounts[lang] = count
		}
	}

	fwRows, err := s.db.Query("SELECT framework_summary, COUNT(*) FROM metadata WHERE framework_summary IS NOT NULL GROUP BY framework_summary")
	if err == nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var fw string
			var count int
			fwRows.Scan(&fw, &count)
			stats.FrameworkCounts[fw] = count
		}
	}

	techRows, err := s.db.Query("SELECT name, category FROM technologies")
	if err == nil {
		defer techRows.Close()
		for techRows.Next() {
			var name, category string
			techRows.Scan(&name, &category)
			stats.TechnologyCounts[name]++
		}
	}

	docRows, err := s.db.Query("SELECT kind, COUNT(*) FROM documents GROUP BY kind")
	if err == nil {
		defer docRows.Close()
		for docRows.Next() {
			var kind string
			var count int
			docRows.Scan(&kind, &count)
			stats.DocumentKindCounts[kind] = count
		}
	}

	s.writeJSON(w, http.StatusOK, stats)
}
