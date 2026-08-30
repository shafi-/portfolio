package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
	Type  string `json:"type,omitempty"`
}

func isLocalhostOrigin(origin string) bool {
	localPatterns := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"http://[::1]",
		"https://[::1]",
	}

	for _, pattern := range localPatterns {
		if strings.HasPrefix(origin, pattern) {
			return true
		}
	}
	return false
}

type Server struct {
	projects      *store.ProjectStore
	metadata      *store.MetadataStore
	documents     *store.DocumentStore
	analyses      *store.AnalysisStore
	features      *store.FeatureStore
	technologies  *store.TechnologyStore
	relationships *store.RelationshipStore
	dependencies  *store.DependencyStore
	configuration *store.ConfigurationStore
	db            *sql.DB
	logger        *logging.Logger
}

func NewServer(db *sql.DB, logger *logging.Logger) *Server {
	zapLogger := logger.Zap()
	return &Server{
		projects:      store.NewProjectStore(db, zapLogger),
		metadata:      store.NewMetadataStore(db, zapLogger),
		documents:     store.NewDocumentStore(db, zapLogger),
		analyses:      store.NewAnalysisStore(db, zapLogger),
		features:      store.NewFeatureStore(db, zapLogger),
		technologies:  store.NewTechnologyStore(db, zapLogger),
		relationships: store.NewRelationshipStore(db, zapLogger),
		dependencies:  store.NewDependencyStore(db, zapLogger),
		configuration: store.NewConfigurationStore(db, zapLogger),
		db:            db,
		logger:        logger,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("GET /projects", s.handleListProjects)
	mux.HandleFunc("GET /projects/{id}", s.handleGetProject)
	mux.HandleFunc("GET /projects/{id}/analysis", s.handleGetAnalysis)

	mux.HandleFunc("GET /search", s.handleSearch)

	mux.HandleFunc("GET /configuration", s.handleGetConfig)
	mux.HandleFunc("PATCH /configuration", s.handlePatchConfig)

	mux.HandleFunc("GET /statistics", s.handleStatistics)
	mux.HandleFunc("GET /technologies", s.handleTechnologies)

	mux.HandleFunc("GET /relationships", s.handleListAllRelationships)
	mux.HandleFunc("GET /relationships/{id}", s.handleListRelationships)
	mux.HandleFunc("POST /relationships/{id}", s.handleStoreRelationship)

	return withCORS(withLogger(mux, s.logger))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	dbOK := true
	if err := s.db.Ping(); err != nil {
		dbOK = false
	}

	projectCount := 0
	var countErr error
	if dbOK {
		countErr = s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
		if countErr != nil {
			dbOK = false
		}
	}

	status := "healthy"
	if !dbOK {
		status = "unhealthy"
	}

	response := map[string]interface{}{
		"status":             status,
		"database_connected": dbOK,
		"project_count":      projectCount,
	}
	if countErr != nil {
		response["error"] = "database_query_failed"
	}

	code := http.StatusOK
	if !dbOK {
		code = http.StatusServiceUnavailable
	}
	s.writeJSON(w, code, response)
}

func (s *Server) handleTechnologies(w http.ResponseWriter, r *http.Request) {
	technologies, err := s.technologies.ListTechnologies()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to fetch technologies")
		s.logger.Error("failed to fetch technologies", models.Field{Key: "error", Value: err})
		return
	}

	if technologies == nil {
		technologies = []*models.Technology{}
	}
	s.writeJSON(w, http.StatusOK, technologies)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("failed to encode JSON", models.Field{Key: "error", Value: err})
		return err
	}
	return nil
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	response := ErrorResponse{
		Error: message,
		Type:  "api_error",
	}

	// Add error code based on status
	switch status {
	case http.StatusNotFound:
		response.Code = "NOT_FOUND"
		response.Type = "resource_error"
	case http.StatusBadRequest:
		response.Code = "INVALID_INPUT"
		response.Type = "validation_error"
	}

	s.writeJSON(w, status, response)
}

func withLogger(next http.Handler, logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request",
			models.Field{Key: "method", Value: r.Method},
			models.Field{Key: "path", Value: r.URL.Path},
			models.Field{Key: "duration", Value: time.Since(start).String()},
		)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only allow localhost origins for local-first security
		if origin == "" || isLocalhostOrigin(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Handle OPTIONS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
