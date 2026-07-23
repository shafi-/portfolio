package discovery

import (
	"strings"
	"testing"
	"testing/fstest"

	"project-dash/internal/fs"
)

// TestMarkerDetector_DetectProjectType tests project type detection for all marker types
func TestMarkerDetector_DetectProjectType(t *testing.T) {
	tests := []struct {
		name           string
		markerFiles    map[string]string
		expectedResult string
	}{
		{
			name:           "no markers",
			markerFiles:    map[string]string{},
			expectedResult: "unknown",
		},
		{
			name: "package.json → node",
			markerFiles: map[string]string{
				"package.json": "{}",
			},
			expectedResult: "node",
		},
		{
			name: "go.mod → go",
			markerFiles: map[string]string{
				"go.mod": "module example",
			},
			expectedResult: "go",
		},
		{
			name: "requirements.txt → python",
			markerFiles: map[string]string{
				"requirements.txt": "requests==2.28.0",
			},
			expectedResult: "python",
		},
		{
			name: "pyproject.toml → python",
			markerFiles: map[string]string{
				"pyproject.toml": "[tool.poetry]",
			},
			expectedResult: "python",
		},
		{
			name: "Cargo.toml → rust",
			markerFiles: map[string]string{
				"Cargo.toml": "[package]",
			},
			expectedResult: "rust",
		},
		{
			name: "pom.xml → java",
			markerFiles: map[string]string{
				"pom.xml": "<project></project>",
			},
			expectedResult: "java",
		},
		{
			name: "polyglot node and go",
			markerFiles: map[string]string{
				"package.json": "{}",
				"go.mod":       "module example",
			},
			expectedResult: "go,node", // alphabetically sorted
		},
		{
			name: "polyglot python and rust",
			markerFiles: map[string]string{
				"requirements.txt": "requests==2.28.0",
				"Cargo.toml":       "[package]",
			},
			expectedResult: "python,rust",
		},
		{
			name: "polyglot all types",
			markerFiles: map[string]string{
				"package.json":     "{}",
				"go.mod":           "module example",
				"requirements.txt": "requests==2.28.0",
				"pyproject.toml":   "[tool.poetry]",
				"Cargo.toml":       "[package]",
				"pom.xml":          "<project></project>",
			},
			expectedResult: "go,java,node,python,rust", // alphabetically sorted, python appears once
		},
		{
			name: "both python markers",
			markerFiles: map[string]string{
				"requirements.txt": "requests==2.28.0",
				"pyproject.toml":   "[tool.poetry]",
			},
			expectedResult: "python", // python should appear only once
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test filesystem with marker files
			fsys := fstest.MapFS{}
			for filename, content := range tt.markerFiles {
				fsys[filename] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			// Create adapter for the filesystem
			fsAdapter := &fs.MapFSAdapter{MapFS: fsys}

			// Create marker detector
			detector := NewMarkerDetector(fsAdapter)

			// Detect project type
			result := detector.DetectProjectType(".")

			// Check result
			if result != tt.expectedResult {
				t.Errorf("DetectProjectType() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestMarkerDetector_DetectProjectTypeSlice tests the slice version of project type detection
func TestMarkerDetector_DetectProjectTypeSlice(t *testing.T) {
	tests := []struct {
		name           string
		markerFiles    map[string]string
		expectedResult []string
	}{
		{
			name:           "no markers",
			markerFiles:    map[string]string{},
			expectedResult: []string{},
		},
		{
			name: "package.json → node",
			markerFiles: map[string]string{
				"package.json": "{}",
			},
			expectedResult: []string{"node"},
		},
		{
			name: "go.mod → go",
			markerFiles: map[string]string{
				"go.mod": "module example",
			},
			expectedResult: []string{"go"},
		},
		{
			name: "polyglot node and go",
			markerFiles: map[string]string{
				"package.json": "{}",
				"go.mod":       "module example",
			},
			expectedResult: []string{"go", "node"},
		},
		{
			name: "polyglot all types",
			markerFiles: map[string]string{
				"package.json":     "{}",
				"go.mod":           "module example",
				"requirements.txt": "requests==2.28.0",
				"pyproject.toml":   "[tool.poetry]",
				"Cargo.toml":       "[package]",
				"pom.xml":          "<project></project>",
			},
			expectedResult: []string{"go", "java", "node", "python", "rust"}, // sorted, python appears once
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test filesystem with marker files
			fsys := fstest.MapFS{}
			for filename, content := range tt.markerFiles {
				fsys[filename] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			// Create adapter for the filesystem
			fsAdapter := &fs.MapFSAdapter{MapFS: fsys}

			// Create marker detector
			detector := NewMarkerDetector(fsAdapter)

			// Detect project type
			result := detector.DetectProjectTypeSlice(".")

			// Check result length
			if len(result) != len(tt.expectedResult) {
				t.Errorf("DetectProjectTypeSlice() length = %v, want %v", len(result), len(tt.expectedResult))
			}

			// Check each element
			for i, expected := range tt.expectedResult {
				if i >= len(result) || result[i] != expected {
					t.Errorf("DetectProjectTypeSlice()[%d] = %v, want %v", i, result[i], expected)
				}
			}
		})
	}
}

// TestMarkerDetector_AcceptanceCriteria tests all acceptance criteria from Story 2.4
func TestMarkerDetector_AcceptanceCriteria(t *testing.T) {
	// AC-14: package.json → repository_type includes "node"
	t.Run("AC-14_package_json_to_node", func(t *testing.T) {
		fsys := fstest.MapFS{
			"package.json": {Data: []byte("{}")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "node" {
			t.Errorf("AC-14 failed: package.json should result in 'node', got '%s'", result)
		}
	})

	// AC-15: go.mod → repository_type includes "go"
	t.Run("AC-15_go_mod_to_go", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go.mod": {Data: []byte("module example")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "go" {
			t.Errorf("AC-15 failed: go.mod should result in 'go', got '%s'", result)
		}
	})

	// AC-16: requirements.txt → repository_type includes "python"
	t.Run("AC-16_requirements_txt_to_python", func(t *testing.T) {
		fsys := fstest.MapFS{
			"requirements.txt": {Data: []byte("requests==2.28.0")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "python" {
			t.Errorf("AC-16 failed: requirements.txt should result in 'python', got '%s'", result)
		}
	})

	// AC-16: pyproject.toml → repository_type includes "python"
	t.Run("AC-16_pyproject_toml_to_python", func(t *testing.T) {
		fsys := fstest.MapFS{
			"pyproject.toml": {Data: []byte("[tool.poetry]")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "python" {
			t.Errorf("AC-16 failed: pyproject.toml should result in 'python', got '%s'", result)
		}
	})

	// AC-17: Cargo.toml → repository_type includes "rust"
	t.Run("AC-17_cargo_toml_to_rust", func(t *testing.T) {
		fsys := fstest.MapFS{
			"Cargo.toml": {Data: []byte("[package]")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "rust" {
			t.Errorf("AC-17 failed: Cargo.toml should result in 'rust', got '%s'", result)
		}
	})

	// AC-18: pom.xml → repository_type includes "java"
	t.Run("AC-18_pom_xml_to_java", func(t *testing.T) {
		fsys := fstest.MapFS{
			"pom.xml": {Data: []byte("<project></project>")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "java" {
			t.Errorf("AC-18 failed: pom.xml should result in 'java', got '%s'", result)
		}
	})

	// AC-19: Multiple markers present → repository_type reflects all detected types
	t.Run("AC-19_polyglot_detection", func(t *testing.T) {
		fsys := fstest.MapFS{
			"package.json":     {Data: []byte("{}")},
			"go.mod":           {Data: []byte("module example")},
			"requirements.txt": {Data: []byte("requests==2.28.0")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		expected := "go,node,python" // alphabetically sorted
		if result != expected {
			t.Errorf("AC-19 failed: multiple markers should result in '%s', got '%s'", expected, result)
		}
	})

	// AC-20: No markers found → repository_type remains "unknown"
	t.Run("AC-20_no_markers_unknown", func(t *testing.T) {
		fsys := fstest.MapFS{
			"README.md": {Data: []byte("# My Project")},
		}
		fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
		detector := NewMarkerDetector(fsAdapter)

		result := detector.DetectProjectType(".")
		if result != "unknown" {
			t.Errorf("AC-20 failed: no markers should result in 'unknown', got '%s'", result)
		}
	})
}

// TestMarkerDetector_FileExistenceOnly tests that marker detection is based on file existence only
func TestMarkerDetector_FileExistenceOnly(t *testing.T) {
	tests := []struct {
		name        string
		markerFiles map[string]string
		expectType  string
	}{
		{
			name: "empty package.json",
			markerFiles: map[string]string{
				"package.json": "", // empty file
			},
			expectType: "node",
		},
		{
			name: "package.json with invalid JSON",
			markerFiles: map[string]string{
				"package.json": "not valid json",
			},
			expectType: "node",
		},
		{
			name: "package.json directory",
			markerFiles: map[string]string{
				"package.json/subdir/file.txt": "content",
			},
			expectType: "node", // directory counts as present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test filesystem with marker files
			fsys := fstest.MapFS{}
			for filename, content := range tt.markerFiles {
				// Check if this is a directory (contains /)
				if idx := strings.Index(filename, "/"); idx != -1 && idx < len(filename)-1 {
					// This is a directory structure
					dirName := filename[:idx]
					fsys[dirName] = &fstest.MapFile{Mode: 040000} // directory mode
					// Add the file inside
					fsys[filename] = &fstest.MapFile{Data: []byte(content)}
				} else {
					// Regular file
					fsys[filename] = &fstest.MapFile{
						Data: []byte(content),
					}
				}
			}

			// Create adapter for the filesystem
			fsAdapter := &fs.MapFSAdapter{MapFS: fsys}

			// Create marker detector
			detector := NewMarkerDetector(fsAdapter)

			// Detect project type
			result := detector.DetectProjectType(".")

			// Check result
			if result != tt.expectType {
				t.Errorf("DetectProjectType() = %v, want %v", result, tt.expectType)
			}
		})
	}
}

// TestMarkerDetector_PolyglotDeduplication tests that duplicate markers are handled correctly
func TestMarkerDetector_PolyglotDeduplication(t *testing.T) {
	// Create a test filesystem with both python markers
	fsys := fstest.MapFS{
		"requirements.txt": {Data: []byte("requests==2.28.0")},
		"pyproject.toml":   {Data: []byte("[tool.poetry]")},
	}

	fsAdapter := &fs.MapFSAdapter{MapFS: fsys}
	detector := NewMarkerDetector(fsAdapter)

	// Test string version
	result := detector.DetectProjectType(".")
	if result != "python" {
		t.Errorf("DetectProjectType() = %v, want 'python' (deduplicated)", result)
	}

	// Test slice version
	resultSlice := detector.DetectProjectTypeSlice(".")
	if len(resultSlice) != 1 {
		t.Errorf("DetectProjectTypeSlice() length = %v, want 1 (deduplicated)", len(resultSlice))
	}
	if resultSlice[0] != "python" {
		t.Errorf("DetectProjectTypeSlice()[0] = %v, want 'python'", resultSlice[0])
	}
}
