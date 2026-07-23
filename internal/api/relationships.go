package api

import "net/http"

type relationshipResponse struct {
	SourceProject string  `json:"source_project"`
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
