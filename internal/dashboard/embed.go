package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"project-dash/pkg/models"
)

//go:embed dist
var dashboardFS embed.FS

// GetEmbeddedFS returns the embedded dashboard filesystem
func GetEmbeddedFS() (fs.FS, error) {
	sub, err := fs.Sub(dashboardFS, "dist")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// HasEmbeddedAssets checks if dashboard assets are embedded
func HasEmbeddedAssets() bool {
	fs, err := GetEmbeddedFS()
	if err != nil {
		return false
	}

	// Try to open index.html to verify assets exist
	_, err = fs.Open("index.html")
	return err == nil
}

// FileServer returns an http.FileServer for the embedded dashboard
func FileServer() (http.Handler, error) {
	sub, err := fs.Sub(dashboardFS, "dist")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(sub)), nil
}

// DefaultDashboardEmbedPath is the path used for sub filesystem
const DefaultDashboardEmbedPath = models.DefaultDashboardEmbedPath
