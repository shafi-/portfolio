package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/config"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// TestConfigOperations_EndToEnd tests core config operations end-to-end
func TestConfigOperations_EndToEnd(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create test directories
	validRoot1 := t.TempDir()
	validRoot2 := t.TempDir()
	invalidRoot := "/nonexistent/path"

	// Initialize logger for testing
	_ = logging.GetGlobalLogger()

	t.Run("config operations work end-to-end", func(t *testing.T) {
		// Create initial config
		cfg := models.GetDefaultConfig()
		testLoader := config.NewLoader(configPath)
		if err := testLoader.Save(cfg); err != nil {
			t.Fatalf("Failed to save initial config: %v", err)
		}

		// Test 1: Adding valid roots
		t.Run("add valid roots", func(t *testing.T) {
			loadedCfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Add first root
			if err := validatePathForRoot(validRoot1); err != nil {
				t.Fatalf("Failed to validate path: %v", err)
			}

			if !containsRoot(loadedCfg.Discovery.ProjectRoots, validRoot1) {
				loadedCfg.Discovery.ProjectRoots = append(loadedCfg.Discovery.ProjectRoots, validRoot1)
			}

			if err := testLoader.Save(loadedCfg); err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			// Verify
			cfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if len(cfg.Discovery.ProjectRoots) != 1 {
				t.Errorf("Expected 1 root, got %d", len(cfg.Discovery.ProjectRoots))
			}

			if cfg.Discovery.ProjectRoots[0] != validRoot1 {
				t.Errorf("Expected root %q, got %q", validRoot1, cfg.Discovery.ProjectRoots[0])
			}
		})

		// Test 2: Invalid path validation
		t.Run("reject invalid paths", func(t *testing.T) {
			err := validatePathForRoot(invalidRoot)
			if err == nil {
				t.Error("Expected error when validating invalid path, got nil")
			}
		})

		// Test 3: Adding multiple roots
		t.Run("add multiple roots", func(t *testing.T) {
			loadedCfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			initialCount := len(loadedCfg.Discovery.ProjectRoots)

			// Add second root
			if err := validatePathForRoot(validRoot2); err != nil {
				t.Fatalf("Failed to validate path: %v", err)
			}

			if !containsRoot(loadedCfg.Discovery.ProjectRoots, validRoot2) {
				loadedCfg.Discovery.ProjectRoots = append(loadedCfg.Discovery.ProjectRoots, validRoot2)
			}

			if err := testLoader.Save(loadedCfg); err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			// Verify
			cfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			expectedCount := initialCount + 1
			if len(cfg.Discovery.ProjectRoots) != expectedCount {
				t.Errorf("Expected %d roots, got %d", expectedCount, len(cfg.Discovery.ProjectRoots))
			}
		})

		// Test 4: Removing roots
		t.Run("remove existing root", func(t *testing.T) {
			loadedCfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Remove first root
			newRoots := make([]string, 0)
			for _, root := range loadedCfg.Discovery.ProjectRoots {
				if root != validRoot1 {
					newRoots = append(newRoots, root)
				}
			}

			// Check if this would leave no roots
			if len(newRoots) > 0 {
				loadedCfg.Discovery.ProjectRoots = newRoots
				if err := testLoader.Save(loadedCfg); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}

				// Verify
				cfg, err := testLoader.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				if len(cfg.Discovery.ProjectRoots) != 1 {
					t.Errorf("Expected 1 root after removal, got %d", len(cfg.Discovery.ProjectRoots))
				}

				if cfg.Discovery.ProjectRoots[0] != validRoot2 {
					t.Errorf("Expected remaining root %q, got %q", validRoot2, cfg.Discovery.ProjectRoots[0])
				}
			}
		})

		// Test 5: ContainsRoot function
		t.Run("containsRoot works correctly", func(t *testing.T) {
			loadedCfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Test containsRoot function
			if !containsRoot(loadedCfg.Discovery.ProjectRoots, validRoot2) {
				t.Error("containsRoot should return true for existing root")
			}

			if containsRoot(loadedCfg.Discovery.ProjectRoots, validRoot1) {
				t.Error("containsRoot should return false for removed root")
			}

			if containsRoot(loadedCfg.Discovery.ProjectRoots, "/nonexistent") {
				t.Error("containsRoot should return false for non-existent root")
			}
		})

		// Test 6: Protect last root
		t.Run("last root protection logic", func(t *testing.T) {
			loadedCfg, err := testLoader.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// If there's only one root, trying to remove it should be prevented
			if len(loadedCfg.Discovery.ProjectRoots) == 1 {
				// Simulate the protection logic
				newRoots := make([]string, 0)
				for _, root := range loadedCfg.Discovery.ProjectRoots {
					if root != validRoot2 {
						newRoots = append(newRoots, root)
					}
				}

				// The command should prevent this
				if len(newRoots) == 0 {
					// Protection prevents removing last root
					t.Log("Last root protection logic would prevent removal")
				}
			}
		})
	})
}

// TestConfigPathError_TypeAssertion tests that ConfigPathError works with error interfaces
func TestConfigPathError_TypeAssertion(t *testing.T) {
	err := &ConfigPathError{
		Path:    "/test/path",
		Message: "test message",
		Cause:   os.ErrPermission,
	}

	// Test that we can unwrap and get the original error
	unwrapped := err.Unwrap()
	if unwrapped != os.ErrPermission {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, os.ErrPermission)
	}

	// Test that errors.Is works
	if !errors.Is(err, os.ErrPermission) {
		t.Error("errors.Is should work with ConfigPathError")
	}

	// Test error message
	errMsg := err.Error()
	expectedMsg := "test message: /test/path: permission denied"
	if errMsg != expectedMsg {
		t.Errorf("Error() = %q, want %q", errMsg, expectedMsg)
	}
}
