package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
	logger    Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, logger Logger) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		logger:    logger,
	}
}

// ServeHTTP handles the health check request
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only GET method is supported
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		w.Header().Set("Allow", "GET, OPTIONS")
		return
	}

	// Check database connectivity
	dbConnected := true
	if err := h.db.Ping(); err != nil {
		dbConnected = false
		h.logger.Error("database health check failed", Field{Key: "error", Value: err})
	}

	// Calculate uptime
	uptime := time.Since(h.startTime)

	// Determine overall status
	status := "ok"
	if !dbConnected {
		status = "unhealthy"
	}

	response := map[string]interface{}{
		"status":   status,
		"uptime":   uptime.String(),
		"database": "disconnected",
	}

	if dbConnected {
		response["database"] = "connected"
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
