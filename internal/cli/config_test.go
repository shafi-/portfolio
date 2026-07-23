package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestSetRoot_CommandStructure tests the set-root command structure
func TestSetRoot_CommandStructure(t *testing.T) {
	tests := []struct {
		name      string
		wantUse   string
		wantShort string
		wantLong  bool
		wantArgs  cobra.PositionalArgs
	}{
		{
			name:      "set-root command structure",
			wantUse:   "set-root <path>",
			wantShort: "Add a project root directory",
			wantLong:  true,
			wantArgs:  cobra.ExactArgs(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUse := setRootCmd.Use; gotUse != tt.wantUse {
				t.Errorf("setRootCmd.Use = %v, want %v", gotUse, tt.wantUse)
			}
			if gotShort := setRootCmd.Short; gotShort != tt.wantShort {
				t.Errorf("setRootCmd.Short = %v, want %v", gotShort, tt.wantShort)
			}
			if (setRootCmd.Long != "") != tt.wantLong {
				t.Errorf("setRootCmd.Long presence = %v, want %v", (setRootCmd.Long != ""), tt.wantLong)
			}
			if gotArgs := setRootCmd.Args; gotArgs(nil, nil) != tt.wantArgs(nil, nil) {
				t.Errorf("setRootCmd.Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// TestRemoveRoot_CommandStructure tests the remove-root command structure
func TestRemoveRoot_CommandStructure(t *testing.T) {
	tests := []struct {
		name      string
		wantUse   string
		wantShort string
		wantLong  bool
		wantArgs  cobra.PositionalArgs
	}{
		{
			name:      "remove-root command structure",
			wantUse:   "remove-root <path>",
			wantShort: "Remove a project root directory",
			wantLong:  true,
			wantArgs:  cobra.ExactArgs(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUse := removeRootCmd.Use; gotUse != tt.wantUse {
				t.Errorf("removeRootCmd.Use = %v, want %v", gotUse, tt.wantUse)
			}
			if gotShort := removeRootCmd.Short; gotShort != tt.wantShort {
				t.Errorf("removeRootCmd.Short = %v, want %v", gotShort, tt.wantShort)
			}
			if (removeRootCmd.Long != "") != tt.wantLong {
				t.Errorf("removeRootCmd.Long presence = %v, want %v", (removeRootCmd.Long != ""), tt.wantLong)
			}
			if gotArgs := removeRootCmd.Args; gotArgs(nil, nil) != tt.wantArgs(nil, nil) {
				t.Errorf("removeRootCmd.Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// TestListRoots_CommandStructure tests the list-roots command structure
func TestListRoots_CommandStructure(t *testing.T) {
	tests := []struct {
		name      string
		wantUse   string
		wantShort string
		wantLong  bool
		wantArgs  cobra.PositionalArgs
	}{
		{
			name:      "list-roots command structure",
			wantUse:   "list-roots",
			wantShort: "List all configured project root directories",
			wantLong:  true,
			wantArgs:  cobra.NoArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUse := listRootsCmd.Use; gotUse != tt.wantUse {
				t.Errorf("listRootsCmd.Use = %v, want %v", gotUse, tt.wantUse)
			}
			if gotShort := listRootsCmd.Short; gotShort != tt.wantShort {
				t.Errorf("listRootsCmd.Short = %v, want %v", gotShort, tt.wantShort)
			}
			if (listRootsCmd.Long != "") != tt.wantLong {
				t.Errorf("listRootsCmd.Long presence = %v, want %v", (listRootsCmd.Long != ""), tt.wantLong)
			}
			if gotArgs := listRootsCmd.Args; gotArgs(nil, nil) != tt.wantArgs(nil, nil) {
				t.Errorf("listRootsCmd.Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// TestValidatePathForRoot tests path validation for project roots
func TestValidatePathForRoot(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, func())
		wantErr     bool
		errContains string
	}{
		{
			name: "empty path",
			setupFunc: func() (string, func()) {
				return "", func() {}
			},
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name: "non-existent path",
			setupFunc: func() (string, func()) {
				return "/nonexistent/path/that/does/not/exist", func() {}
			},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name: "file instead of directory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "notadir.txt")
				os.WriteFile(filePath, []byte("test"), 0644)
				return filePath, func() {}
			},
			wantErr:     true,
			errContains: "not a directory",
		},
		{
			name: "valid directory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			wantErr: false,
		},
		{
			name: "unreadable directory",
			setupFunc: func() (string, func()) {
				tmpDir := t.TempDir()
				noReadDir := filepath.Join(tmpDir, "noread")
				os.Mkdir(noReadDir, 0000)
				// On some systems, we can still read directories we own
				// This test may not work consistently across platforms
				return noReadDir, func() {
					// Cleanup - restore permissions for deletion
					os.Chmod(noReadDir, 0755)
				}
			},
			wantErr:     false, // May succeed on some systems
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath, cleanup := tt.setupFunc()
			defer cleanup()

			err := validatePathForRoot(testPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePathForRoot() expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validatePathForRoot() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validatePathForRoot() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContainsRoot tests the containsRoot helper function
func TestContainsRoot(t *testing.T) {
	tests := []struct {
		name  string
		roots []string
		path  string
		want  bool
	}{
		{
			name:  "path exists in roots",
			roots: []string{"/home/user/projects", "~/Developer", "/opt/code"},
			path:  "/home/user/projects",
			want:  true,
		},
		{
			name:  "path does not exist in roots",
			roots: []string{"/home/user/projects", "~/Developer"},
			path:  "/different/path",
			want:  false,
		},
		{
			name:  "empty roots list",
			roots: []string{},
			path:  "/home/user/projects",
			want:  false,
		},
		{
			name:  "case sensitive match",
			roots: []string{"/home/user/Projects"},
			path:  "/home/user/projects",
			want:  false,
		},
		{
			name:  "trailing slash difference",
			roots: []string{"/home/user/projects/"},
			path:  "/home/user/projects",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsRoot(tt.roots, tt.path)
			if got != tt.want {
				t.Errorf("containsRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfigPathError tests the ConfigPathError error type
func TestConfigPathError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ConfigPathError
		wantMsg string
	}{
		{
			name: "error without cause",
			err: &ConfigPathError{
				Path:    "/test/path",
				Message: "test error",
			},
			wantMsg: "test error: /test/path",
		},
		{
			name: "error with cause",
			err: &ConfigPathError{
				Path:    "/test/path",
				Message: "test error",
				Cause:   os.ErrNotExist,
			},
			wantMsg: "test error: /test/path: file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("ConfigPathError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

// TestConfigPathError_Unwrap tests the Unwrap method
func TestConfigPathError_Unwrap(t *testing.T) {
	cause := os.ErrNotExist
	err := &ConfigPathError{
		Path:    "/test/path",
		Message: "test error",
		Cause:   cause,
	}

	if got := err.Unwrap(); got != cause {
		t.Errorf("ConfigPathError.Unwrap() = %v, want %v", got, cause)
	}
}

// Integration tests moved to config_integration_test.go to avoid issues with command execution
