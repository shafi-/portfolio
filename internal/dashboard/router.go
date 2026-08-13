package dashboard

import (
	"net/http"
	"strings"
)

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
func (router *DashboardRouter) setCorsHeaders(w http.ResponseWriter) {
	if router.corsEnabled {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

// ServeHTTP implements http.Handler
func (router *DashboardRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS preflight
	if r.Method == "OPTIONS" {
		router.setCorsHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Route to API handler
	if router.isAPIRoute(r.URL.Path) {
		router.setCorsHeaders(w)
		router.apiHandler.ServeHTTP(w, r)
		return
	}

	// Serve static assets
	router.assetServer.ServeHTTP(w, r)
}
