package api

import (
	"net/http"
	"strconv"
	"strings"

	"project-dash/pkg/models"
)

type projectResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	RootPath       string              `json:"root_path"`
	RepositoryType string              `json:"repository_type"`
	DiscoveredAt   string              `json:"discovered_at"`
	UpdatedAt      string              `json:"updated_at"`
	Metadata       *models.Metadata     `json:"metadata,omitempty"`
	Documents      []*models.Document   `json:"documents,omitempty"`
	Analyses       []*models.Analysis   `json:"analyses,omitempty"`
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projects.ListProjects()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	if projects == nil {
		projects = []*models.Project{}
	}

	search := r.URL.Query().Get("q")
	if search != "" {
		var filtered []*models.Project
		search = strings.ToLower(search)
		for _, p := range projects {
			if strings.Contains(strings.ToLower(p.Name), search) {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "name"
	}
	if sortBy == "name" {
		for i := 0; i < len(projects); i++ {
			for j := i + 1; j < len(projects); j++ {
				if projects[i].Name > projects[j].Name {
					projects[i], projects[j] = projects[j], projects[i]
				}
			}
		}
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 0
	offset := 0
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	total := len(projects)
	if offset > total {
		offset = total
	}
	if limit > 0 && offset+limit < total {
		projects = projects[offset : offset+limit]
	} else {
		projects = projects[offset:]
	}

	var responses []projectResponse
	for _, p := range projects {
		meta, _ := s.metadata.GetMetadata(p.ID)
		responses = append(responses, projectResponse{
			ID:             p.ID,
			Name:           p.Name,
			RootPath:       p.RootPath,
			RepositoryType: p.RepositoryType,
			DiscoveredAt:   p.DiscoveredAt,
			UpdatedAt:      p.UpdatedAt,
			Metadata:       meta,
		})
	}

	if responses == nil {
		responses = []projectResponse{}
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": responses,
		"total":    total,
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	project, err := s.projects.GetProject(id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if project == nil {
		s.writeError(w, http.StatusNotFound, "project not found")
		return
	}

	meta, _ := s.metadata.GetMetadata(id)
	docs, _ := s.documents.ListDocuments(id)
	analyses, _ := s.analyses.ListAnalyses(id)

	resp := projectResponse{
		ID:             project.ID,
		Name:           project.Name,
		RootPath:       project.RootPath,
		RepositoryType: project.RepositoryType,
		DiscoveredAt:   project.DiscoveredAt,
		UpdatedAt:      project.UpdatedAt,
		Metadata:       meta,
		Documents:      docs,
		Analyses:       analyses,
	}

	s.writeJSON(w, http.StatusOK, resp)
}
