package api

import (
	"net/http"

	"project-dash/pkg/models"
)

func (s *Server) handleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	project, err := s.projects.GetProject(projectID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if project == nil {
		s.writeError(w, http.StatusNotFound, "project not found")
		return
	}

	analyses, err := s.analyses.ListAnalyses(projectID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to get analyses")
		return
	}

	if analyses == nil {
		analyses = []*models.Analysis{}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id": projectID,
		"analyses":   analyses,
		"count":      len(analyses),
	})
}