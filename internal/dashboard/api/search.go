package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Shared error handling functions (defined in configuration.go)

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string
	Technology string
	Framework  string
	FromDate   string
	ToDate     string
	Page       int
	PageSize   int
}

// SearchResponse represents a search response
type SearchResponse struct {
	TotalResults int          `json:"total_results"`
	Page         int          `json:"page"`
	PageSize     int          `json:"page_size"`
	TotalPages   int          `json:"total_pages"`
	Results      []ResultItem `json:"results"`
}

// ResultItem represents a single search result
type ResultItem struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
}

// SearchHandler handles search requests
type SearchHandler struct {
	db     *sql.DB
	logger Logger
}

// Logger interface for search logging
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *sql.DB, logger Logger) *SearchHandler {
	return &SearchHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles the search request
func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only GET method is supported
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		w.Header().Set("Allow", "GET, OPTIONS")
		return
	}

	// Parse and validate request
	req, err := h.parseRequest(r.URL.Query())
	if err != nil {
		h.logger.Warn("invalid search request", Field{Key: "error", Value: err})
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Execute search
	response, err := h.executeSearch(req)
	if err != nil {
		h.logger.Error("search execution failed", Field{Key: "error", Value: err})
		WriteError(w, http.StatusInternalServerError, "search failed")
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseRequest parses and validates the search request
func (h *SearchHandler) parseRequest(q url.Values) (*SearchRequest, error) {
	req := &SearchRequest{
		Page:     1,
		PageSize: 20,
	}

	// Get query parameter (required)
	query := q.Get("q")
	if query == "" {
		return nil, fmt.Errorf("query parameter 'q' is required")
	}
	req.Query = query

	// Parse optional filters
	req.Technology = q.Get("technology")
	req.Framework = q.Get("framework")
	req.FromDate = q.Get("from")
	req.ToDate = q.Get("to")

	// Validate date formats
	if req.FromDate != "" && !isValidISODate(req.FromDate) {
		return nil, fmt.Errorf("invalid date format for 'from' — use ISO 8601")
	}
	if req.ToDate != "" && !isValidISODate(req.ToDate) {
		return nil, fmt.Errorf("invalid date format for 'to' — use ISO 8601")
	}

	// Parse pagination parameters
	if pageStr := q.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return nil, fmt.Errorf("page must be a positive integer")
		}
		req.Page = page
	}

	if pageSizeStr := q.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			return nil, fmt.Errorf("page_size must be a positive integer")
		}
		// Cap page size at 100
		if pageSize > 100 {
			pageSize = 100
		}
		req.PageSize = pageSize
	}

	return req, nil
}

// isValidISODate checks if a string is a valid ISO 8601 date
func isValidISODate(dateStr string) bool {
	// Accept YYYY-MM-DD format
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !re.MatchString(dateStr) {
		return false
	}
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// executeSearch performs the actual search
func (h *SearchHandler) executeSearch(req *SearchRequest) (*SearchResponse, error) {
	// Get total count
	total, err := h.countResults(req)
	if err != nil {
		return nil, fmt.Errorf("failed to count results: %w", err)
	}

	// Get paginated results
	results, err := h.fetchResults(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	// Calculate pagination metadata
	totalPages := (total + req.PageSize - 1) / req.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &SearchResponse{
		TotalResults: total,
		Page:         req.Page,
		PageSize:     req.PageSize,
		TotalPages:   totalPages,
		Results:      results,
	}, nil
}

// countResults counts total matching results
func (h *SearchHandler) countResults(req *SearchRequest) (int, error) {
	query := "SELECT COUNT(*) FROM (SELECT 1 FROM projects WHERE name LIKE ?"
	args := []interface{}{"%" + req.Query + "%"}

	// Add documents to the count
	query += " UNION SELECT 1 FROM documents d JOIN projects p ON p.id = d.project_id WHERE d.content LIKE ?"
	args = append(args, "%"+req.Query+"%")

	query += ")"

	var total int
	err := h.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// fetchResults fetches paginated search results
func (h *SearchHandler) fetchResults(req *SearchRequest) ([]ResultItem, error) {
	var results []ResultItem

	// Search projects
	projectResults, err := h.searchProjects(req)
	if err == nil {
		results = append(results, projectResults...)
	}

	// Search documents
	documentResults, err := h.searchDocuments(req)
	if err == nil {
		results = append(results, documentResults...)
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	start := offset
	end := offset + req.PageSize

	if start >= len(results) {
		return []ResultItem{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// searchProjects searches for matching projects
func (h *SearchHandler) searchProjects(req *SearchRequest) ([]ResultItem, error) {
	query := `SELECT id, name FROM projects WHERE name LIKE ?`
	args := []interface{}{"%" + req.Query + "%"}

	// Apply filters
	if req.Technology != "" {
		query += " AND id IN (SELECT project_id FROM project_technologies WHERE technology = ?)"
		args = append(args, req.Technology)
	}
	if req.Framework != "" {
		query += " AND id IN (SELECT project_id FROM project_frameworks WHERE framework = ?)"
		args = append(args, req.Framework)
	}
	if req.FromDate != "" {
		query += " AND discovered_at >= ?"
		args = append(args, req.FromDate)
	}
	if req.ToDate != "" {
		query += " AND discovered_at <= ?"
		args = append(args, req.ToDate)
	}

	query += " ORDER BY name LIMIT 100"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ResultItem
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		results = append(results, ResultItem{
			Type:    "project",
			ID:      id,
			Name:    name,
			Snippet: h.highlightText(name, req.Query),
		})
	}

	return results, nil
}

// searchDocuments searches for matching documents
func (h *SearchHandler) searchDocuments(req *SearchRequest) ([]ResultItem, error) {
	query := `SELECT d.id, d.project_id, p.name as project_name, d.path, d.kind, d.content
			  FROM documents d
			  JOIN projects p ON p.id = d.project_id
			  WHERE d.content LIKE ?`
	args := []interface{}{"%" + req.Query + "%"}

	// Apply date filters to documents
	if req.FromDate != "" {
		query += " AND d.discovered_at >= ?"
		args = append(args, req.FromDate)
	}
	if req.ToDate != "" {
		query += " AND d.discovered_at <= ?"
		args = append(args, req.ToDate)
	}

	// Apply project-level filters
	if req.Technology != "" {
		query += " AND d.project_id IN (SELECT project_id FROM project_technologies WHERE technology = ?)"
		args = append(args, req.Technology)
	}
	if req.Framework != "" {
		query += " AND d.project_id IN (SELECT project_id FROM project_frameworks WHERE framework = ?)"
		args = append(args, req.Framework)
	}

	query += " ORDER BY d.kind LIMIT 100"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ResultItem
	for rows.Next() {
		var id, projectID, projectName, path, kind, content string
		if err := rows.Scan(&id, &projectID, &projectName, &path, &kind, &content); err != nil {
			continue
		}
		results = append(results, ResultItem{
			Type:        "document",
			ID:          id,
			ProjectID:   projectID,
			ProjectName: projectName,
			Path:        path,
			Kind:        kind,
			Snippet:     h.generateSnippet(content, req.Query),
		})
	}

	return results, nil
}

// highlightText highlights query terms in text
func (h *SearchHandler) highlightText(text, query string) string {
	// Convert query to regex
	escapedQuery := regexp.QuoteMeta(query)
	re := regexp.MustCompile(`(?i)` + escapedQuery)
	return re.ReplaceAllString(text, `<mark>$0</mark>`)
}

// generateSnippet generates a highlighted snippet from content
func (h *SearchHandler) generateSnippet(content, query string) string {
	// Find first occurrence of query in content
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		// Fallback to first 200 chars
		if len(content) > 200 {
			return content[:200] + "..."
		}
		return content
	}

	// Generate context around the match (50 chars before and after)
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 50
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]

	// Add ellipsis if we truncated
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	// Highlight the query term
	return h.highlightText(snippet, query)
}

// The WriteError and errorCodeForStatus functions are defined in configuration.go
// to avoid duplication across multiple handlers
