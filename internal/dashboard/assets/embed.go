package assets

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed dashboard/dist/index.html
var dashboardFS embed.FS

// DashboardFS returns the embedded dashboard filesystem
func DashboardFS() (fs.FS, error) {
	sub, err := fs.Sub(dashboardFS, "dashboard/dist")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// HasEmbeddedAssets returns true if embedded assets are available
func HasEmbeddedAssets() bool {
	// Check if we have the index.html file
	_, err := DashboardFS()
	if err != nil {
		return false
	}

	// Try to open index.html to verify assets exist
	f, err := dashboardFS.Open("dashboard/dist/index.html")
	if err != nil {
		return false
	}
	defer f.Close()

	// Try to stat to verify it's a valid file
	_, err = f.Stat()
	return err == nil
}

// EnsureDashboardDir ensures the dashboard directory exists for external serving
func EnsureDashboardDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, 0755)
}
