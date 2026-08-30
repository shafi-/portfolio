package dashboard

import (
	"net/http"
	"strings"
)

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

// DashboardRouter routes requests between API and asset serving
type DashboardRouter struct {
	assetServer *AssetServer
	apiHandler  http.Handler
	corsEnabled bool
}

// NewDashboardRouter creates a new dashboard router
func NewDashboardRouter(assetServer *AssetServer, apiHandler http.Handler, corsEnabled bool) *DashboardRouter {
	return &DashboardRouter{
		assetServer: assetServer,
		apiHandler:  apiHandler,
		corsEnabled: corsEnabled,
	}
}

// isAPIRoute detects if a path is an API route
func (router *DashboardRouter) isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/health") ||
		strings.HasPrefix(path, "/configuration") ||
		strings.HasPrefix(path, "/projects") ||
		strings.HasPrefix(path, "/search") ||
		strings.HasPrefix(path, "/relationships") ||
		strings.HasPrefix(path, "/statistics")
}

// setCorsHeaders sets CORS headers if enabled
func (router *DashboardRouter) setCorsHeaders(w http.ResponseWriter, origin string) {
	if router.corsEnabled {
		// Only allow localhost origins for local-first security
		if origin == "" || isLocalhostOrigin(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}
}

// ServeHTTP implements http.Handler
func (router *DashboardRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		router.setCorsHeaders(w, origin)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Route to API handler
	if router.isAPIRoute(r.URL.Path) {
		router.setCorsHeaders(w, origin)
		router.apiHandler.ServeHTTP(w, r)
		return
	}

	// Serve static assets
	router.assetServer.ServeHTTP(w, r)
}
