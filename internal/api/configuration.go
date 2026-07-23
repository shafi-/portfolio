package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	configs, err := s.configuration.List()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to list configuration")
		return
	}

	cfg := make(map[string]string)
	for _, c := range configs {
		cfg[c.Key] = c.Value
	}

	s.writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handlePatchConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	for key, value := range updates {
		if key == "" {
			s.writeError(w, http.StatusBadRequest, "empty configuration key")
			return
		}
		if err := s.configuration.Set(key, value); err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to update configuration")
			return
		}
	}

	configs, err := s.configuration.List()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to list configuration")
		return
	}

	cfg := make(map[string]string)
	for _, c := range configs {
		cfg[c.Key] = c.Value
	}
	s.writeJSON(w, http.StatusOK, cfg)
}
