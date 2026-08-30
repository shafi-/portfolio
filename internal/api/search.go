package api

import (
	"net/http"
	"strconv"
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

const maxQueryLength = 500

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	// Validate query length
	if len(q) > maxQueryLength {
		s.writeError(w, http.StatusBadRequest, "query too long")
		return
	}

	// Trim whitespace
	q = strings.TrimSpace(q)

	// Get technology and framework filters first
	technologies := r.URL.Query()["technology"]
	frameworks := r.URL.Query()["framework"]

	// Return 400 if query is missing and no filters are provided
	if q == "" && len(technologies) == 0 && len(frameworks) == 0 {
		s.writeError(w, http.StatusBadRequest, "query parameter is required")
		return
	}

	// Get pagination parameters
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	pageSize := r.URL.Query().Get("page_size")
	if pageSize == "" {
		pageSize = "20"
	}

	var results []searchResult

	// Build base query for projects with JOIN to metadata
	var projectQuery string
	var countQuery string
	var args []interface{}

	// Start with name filter (case-insensitive)
	projectQuery = "SELECT p.id, p.name FROM projects p LEFT JOIN metadata m ON p.id = m.project_id WHERE p.name LIKE ?"
	countQuery = "SELECT COUNT(*) FROM projects p LEFT JOIN metadata m ON p.id = m.project_id WHERE p.name LIKE ?"
	args = append(args, "%"+q+"%")

	// Add technology filter if provided (case-insensitive)
	if len(technologies) > 0 {
		var techConditions []string
		for _, tech := range technologies {
			techConditions = append(techConditions, "LOWER(m.language_summary) LIKE ?")
			args = append(args, "%"+strings.ToLower(tech)+"%")
		}
		if len(techConditions) > 0 {
			projectQuery += " AND (" + strings.Join(techConditions, " OR ") + ")"
			countQuery += " AND (" + strings.Join(techConditions, " OR ") + ")"
		}
	}

	// Add framework filter if provided (case-insensitive)
	if len(frameworks) > 0 {
		var frameworkConditions []string
		for _, framework := range frameworks {
			frameworkConditions = append(frameworkConditions, "LOWER(m.framework_summary) LIKE ?")
			args = append(args, "%"+strings.ToLower(framework)+"%")
		}
		if len(frameworkConditions) > 0 {
			projectQuery += " AND (" + strings.Join(frameworkConditions, " OR ") + ")"
			countQuery += " AND (" + strings.Join(frameworkConditions, " OR ") + ")"
		}
	}

	// Get total count first
	var totalCount int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	countRow := s.db.QueryRow(countQuery, countArgs...)
	if err := countRow.Scan(&totalCount); err != nil {
		totalCount = 0
	}

	// Calculate offset for pagination
	offset := 0
	if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 1 {
		if pageSizeNum, err := strconv.Atoi(pageSize); err == nil && pageSizeNum > 0 {
			offset = (pageNum - 1) * pageSizeNum
		}
	}

	projectQuery += " ORDER BY p.name LIMIT ? OFFSET ?"

	// Add pagination args to query args
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, pageSize, offset)

	projectRows, err := s.db.Query(projectQuery, queryArgs...)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to query projects")
		return
	}
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

	// Only include document search if no technology/framework filters are applied
	if len(technologies) == 0 && len(frameworks) == 0 {
		ftsQuery := s.buildFTSQuery(q)
		if ftsQuery != "" {
			docRows, err := s.db.Query(
				`SELECT d.id, d.project_id, d.path, d.kind, d.content, p.name
				 FROM documents d JOIN projects p ON p.id = d.project_id
				 WHERE d.content LIKE ? ORDER BY d.kind LIMIT 50`,
				"%"+q+"%",
			)
			if err != nil {
				s.writeError(w, http.StatusInternalServerError, "failed to search documents")
				return
			}
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

	// Calculate total pages
	totalPages := 0
	if pageSizeNum, err := strconv.Atoi(pageSize); err == nil && pageSizeNum > 0 {
		totalPages = (totalCount + pageSizeNum - 1) / pageSizeNum
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results":    results,
		"totalCount": totalCount,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
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
