package api

import (
	"database/sql"
	"net/http"
	"strings"
)

type searchResult struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Name    string      `json:"name,omitempty"`
	Path    string      `json:"path,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Content string      `json:"content,omitempty"`
	Rank    float64     `json:"rank"`
	Project *searchProj `json:"project,omitempty"`
}

type searchProj struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		s.writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	var results []searchResult

	projectRows, err := s.db.Query(
		"SELECT id, name FROM projects WHERE name LIKE ? ORDER BY name LIMIT 50",
		"%"+q+"%",
	)
	if err == nil {
		defer projectRows.Close()
		for projectRows.Next() {
			var id, name string
			if err := projectRows.Scan(&id, &name); err != nil {
				continue
			}
			results = append(results, searchResult{
				Type: "project", ID: id, Name: name,
			})
		}
	}

	ftsQuery := s.buildFTSQuery(q)
	if ftsQuery != "" {
		docRows, err := s.db.Query(
			`SELECT d.id, d.project_id, d.path, d.kind, d.content, p.name
			 FROM documents d JOIN projects p ON p.id = d.project_id
			 WHERE d.content LIKE ? ORDER BY d.kind LIMIT 50`,
			"%"+q+"%",
		)
		if err == nil {
			defer docRows.Close()
			for docRows.Next() {
				var id, projectID, path, kind, content, projName string
				if err := docRows.Scan(&id, &projectID, &path, &kind, &content, &projName); err != nil {
					continue
				}
				results = append(results, searchResult{
					Type:    "document",
					ID:      id,
					Path:    path,
					Kind:    kind,
					Content: truncateContent(content),
					Project: &searchProj{ID: projectID, Name: projName},
				})
			}
		}
	}

	if results == nil {
		results = []searchResult{}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   q,
	})
}

func (s *Server) buildFTSQuery(q string) string {
	if !s.hasFTS5() {
		return ""
	}
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return ""
	}
	var ftsParts []string
	for _, p := range parts {
		ftsParts = append(ftsParts, p+`*`)
	}
	return strings.Join(ftsParts, " AND ")
}

func (s *Server) hasFTS5() bool {
	var tableName string
	err := s.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='documents_fts'").Scan(&tableName)
	return err == nil
}

func truncateContent(content string) string {
	if len(content) > 200 {
		return content[:200] + "..."
	}
	return content
}

// ensure rows are closed properly
func closeRows(rows *sql.Rows) {
	rows.Close()
}
