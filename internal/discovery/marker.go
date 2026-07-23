package discovery

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MarkerDetector detects project type markers in a repository
type MarkerDetector struct {
	fs MarkerFS
}

// MarkerFS interface for marker detection operations
type MarkerFS interface {
	Lstat(path string) (os.FileInfo, error)
}

// NewMarkerDetector creates a new MarkerDetector
func NewMarkerDetector(fs MarkerFS) *MarkerDetector {
	return &MarkerDetector{fs: fs}
}

// marker represents a project type marker file and its corresponding type
type marker struct {
	file        string
	projectType string
}

// markers defines all supported project type markers
var markers = []marker{
	{"package.json", "node"},
	{"go.mod", "go"},
	{"requirements.txt", "python"},
	{"pyproject.toml", "python"},
	{"Cargo.toml", "rust"},
	{"pom.xml", "java"},
}

// DetectProjectType detects project type markers at the root of a repository
// Returns a comma-separated string of detected types (e.g., "node" or "node,go")
// Returns "unknown" if no recognized markers are found
func (m *MarkerDetector) DetectProjectType(dirPath string) string {
	// Use the slice version for deduplication and consistency
	detectedTypes := m.DetectProjectTypeSlice(dirPath)

	// If no markers found, return "unknown"
	if len(detectedTypes) == 0 {
		return "unknown"
	}

	// Join with comma for polyglot projects
	return strings.Join(detectedTypes, ",")
}

// DetectProjectTypeSlice detects project type markers and returns them as a slice
// This is useful for testing or when you need the individual types
func (m *MarkerDetector) DetectProjectTypeSlice(dirPath string) []string {
	var detectedTypes []string

	// Check each marker file
	seenTypes := make(map[string]bool)
	for _, marker := range markers {
		markerPath := filepath.Join(dirPath, marker.file)
		if _, err := m.fs.Lstat(markerPath); err == nil {
			// Marker file exists
			if !seenTypes[marker.projectType] {
				detectedTypes = append(detectedTypes, marker.projectType)
				seenTypes[marker.projectType] = true
			}
		}
	}

	// Sort to ensure consistent ordering
	sort.Strings(detectedTypes)

	return detectedTypes
}
