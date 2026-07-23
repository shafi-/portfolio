package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"project-dash/pkg/models"
)

type relationshipResponse struct {
	SourceProject string  `json:"source_project"`
	TargetProject string  `json:"target_project"`
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Confidence    float64 `json:"confidence"`
}

type storeRelationshipRequest struct {
	TargetProject string  `json:"target_project"`
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Confidence    float64 `json:"confidence"`
}

func (s *Server) handleListRelationships(w http.ResponseWriter, r *http.Request) {
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

	rels, err := s.relationships.ListRelationships(id)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to list relationships")
		return
	}

	var resp []relationshipResponse
	for _, rel := range rels {
		resp = append(resp, relationshipResponse{
			SourceProject: rel.SourceProject,
			TargetProject: rel.TargetProject,
			Type:          rel.Type,
			Description:   rel.Description,
			Confidence:    rel.Confidence,
		})
	}
	if resp == nil {
		resp = []relationshipResponse{}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":    id,
		"relationships": resp,
	})
}

func (s *Server) handleStoreRelationship(w http.ResponseWriter, r *http.Request) {
	sourceID := r.PathValue("id")

	var req storeRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sourceProj, err := s.projects.GetProject(sourceID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if sourceProj == nil {
		s.writeError(w, http.StatusNotFound, "source project not found")
		return
	}

	targetProj, err := s.projects.GetProject(req.TargetProject)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if targetProj == nil {
		s.writeError(w, http.StatusNotFound, "target project not found")
		return
	}

	relationship := &models.Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceID,
		TargetProject: req.TargetProject,
		Type:          req.Type,
		Description:   req.Description,
		Confidence:    req.Confidence,
	}

	if err := models.ValidateRelationship(relationship); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.relationships.CreateRelationship(relationship); err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to store relationship")
		return
	}

	s.writeJSON(w, http.StatusCreated, relationshipResponse{
		SourceProject: relationship.SourceProject,
		TargetProject: relationship.TargetProject,
		Type:          relationship.Type,
		Description:   relationship.Description,
		Confidence:    relationship.Confidence,
	})
}
