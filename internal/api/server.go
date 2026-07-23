package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"project-dash/internal/logging"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

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
	if dbOK {
		s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
	}

	status := "healthy"
	if !dbOK {
		status = "unhealthy"
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":             status,
		"database_connected": dbOK,
		"project_count":      projectCount,
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
