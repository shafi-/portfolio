package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// ConfigHandler handles configuration requests
type ConfigHandler struct {
	config      Config
	configStore ConfigStore
	logger      Logger
}

// Config represents the configuration interface
type Config interface {
	GetDashboardPort() int
	GetDashboardHost() string
	GetDashboardAssetsPath() string
	GetAllowedOrigins() []string
	SetDashboardPort(port int) error
	SetDashboardHost(host string) error
	SetDashboardAssetsPath(path string) error
	SetAllowedOrigins(origins []string) error
	ToMap() map[string]interface{}
	FromMap(data map[string]interface{}) error
}

// ConfigStore handles configuration persistence
type ConfigStore interface {
	Save(config Config) error
	Load() (Config, error)
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(config Config, store ConfigStore, logger Logger) *ConfigHandler {
	return &ConfigHandler{
		config:      config,
		configStore: store,
		logger:      logger,
	}
}

// ServeHTTP handles configuration requests
func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetConfig(w, r)
	case http.MethodPatch:
		h.handlePatchConfig(w, r)
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		w.Header().Set("Allow", "GET, PATCH, OPTIONS")
	}
}

// handleGetConfig handles GET /configuration
func (h *ConfigHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	response := h.config.ToMap()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handlePatchConfig handles PATCH /configuration
func (h *ConfigHandler) handlePatchConfig(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.logger.Warn("invalid configuration update", Field{Key: "error", Value: err})
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate updates
	if err := h.validateUpdates(updates); err != nil {
		h.logger.Warn("configuration validation failed", Field{Key: "error", Value: err})
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Apply updates
	if err := h.config.FromMap(updates); err != nil {
		h.logger.Error("failed to apply configuration updates", Field{Key: "error", Value: err})
		WriteError(w, http.StatusInternalServerError, "failed to apply updates")
		return
	}

	// Save configuration
	if err := h.configStore.Save(h.config); err != nil {
		h.logger.Error("failed to save configuration", Field{Key: "error", Value: err})
		WriteError(w, http.StatusInternalServerError, "failed to save configuration")
		return
	}

	// Return updated configuration
	response := h.config.ToMap()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// validateUpdates validates configuration updates
func (h *ConfigHandler) validateUpdates(updates map[string]interface{}) error {
	// Define allowed configuration keys and their types
	allowedFields := map[string]string{
		"dashboard.port":            "int",
		"dashboard.host":            "string",
		"dashboard.assets_path":     "string",
		"dashboard.allowed_origins": "[]string",
	}

	// Check for unknown keys
	for key := range updates {
		if _, ok := allowedFields[key]; !ok {
			return fmt.Errorf("unknown configuration key: %s", key)
		}
	}

	// Validate types
	for key, value := range updates {
		expectedType, ok := allowedFields[key]
		if !ok {
			continue
		}

		if expectedType == "int" {
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("invalid type for field '%s': expected integer", key)
			}
		} else if expectedType == "string" {
			if _, ok := value.(string); !ok {
				return fmt.Errorf("invalid type for field '%s': expected string", key)
			}
		} else if expectedType == "[]string" {
			if !isStringArray(value) {
				return fmt.Errorf("invalid type for field '%s': expected array of strings", key)
			}
		}
	}

	return nil
}

// isStringArray checks if a value is an array of strings
func isStringArray(value interface{}) bool {
	if value == nil {
		return false
	}

	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i).Interface()
		if _, ok := elem.(string); !ok {
			return false
		}
	}

	return true
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   errorCodeForStatus(status),
		"message": message,
	})
}

// errorCodeForStatus maps HTTP status codes to error codes
func errorCodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusInternalServerError:
		return "internal_error"
	default:
		return "internal_error"
	}
}
