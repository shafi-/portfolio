package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// DefaultCORSConfig returns default CORS settings for local development
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}
}

// CORSMiddleware creates a CORS middleware
func CORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range config.AllowedOrigins {
		allowedOriginsMap[origin] = true
	}

	allowedMethodsMap := make(map[string]bool)
	for _, method := range config.AllowedMethods {
		allowedMethodsMap[method] = true
	}

	allowedHeadersMap := make(map[string]bool)
	for _, header := range config.AllowedHeaders {
		allowedHeadersMap[strings.ToLower(header)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && allowedOriginsMap[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set allowed methods
			if len(config.AllowedMethods) > 0 {
				methodsHeader := strings.Join(config.AllowedMethods, ", ")
				w.Header().Set("Access-Control-Allow-Methods", methodsHeader)
			}

			// Set allowed headers
			if len(config.AllowedHeaders) > 0 {
				headersHeader := strings.Join(config.AllowedHeaders, ", ")
				w.Header().Set("Access-Control-Allow-Headers", headersHeader)
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
