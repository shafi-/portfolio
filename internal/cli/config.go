package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// normalizePath performs comprehensive path normalization with robust edge case handling
func normalizePath(path string) (string, error) {
	// Check for empty path
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Expand home directory if present
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Clean the path (removes redundant separators, . , .. etc)
	// This handles trailing slashes and other path issues
	path = filepath.Clean(path)

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("cannot convert to absolute path: %w", err)
		}
		path = abs
	}

	return path, nil
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Portfolio configuration",
	Long: `Manage Portfolio Engine configuration settings.

This command provides subcommands for managing project roots, ignored paths,
and other configuration settings stored in the Portfolio configuration file.

Available Subcommands:
  • set-root    - Add a project root directory for discovery
  • remove-root - Remove a project root directory from configuration
  • list-roots  - List all configured project roots with validation status

Common Workflows:
  1. Set up your first project root:
     portfolio config set-root ~/Projects
     portfolio config list-roots

  2. Add multiple workspace directories:
     portfolio config set-root ~/work/frontend
     portfolio config set-root ~/work/backend
     portfolio config set-root ~/work/devops

  3. Review and clean up configuration:
     portfolio config list-roots
     portfolio config remove-root ~/old-workspace

Path Normalization:
  All paths are automatically normalized to handle:
  • Trailing slashes: ~/Projects/ → ~/Projects
  • Relative paths: ./projects → /full/path/to/projects
  • Home directory expansion: ~ → /home/user
  • Multiple/redundant slashes: /home//user → /home/user

Examples:
  # Show help for config subcommands
  portfolio config set-root --help
  portfolio config list-roots --help

  # Common configuration workflow
  portfolio config set-root ~/Projects
  portfolio config list-roots`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand provided, show help
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(setRootCmd)
	configCmd.AddCommand(removeRootCmd)
	configCmd.AddCommand(listRootsCmd)
}

// setRootCmd represents the config set-root command
var setRootCmd = &cobra.Command{
	Use:   "set-root <path>",
	Short: "Add a project root directory",
	Long: `Add a directory path to the list of project roots for discovery.

The path must exist, be accessible, and be a directory. The path is automatically
normalized to handle trailing slashes, relative paths, and home directory expansion.
If the path is already in the project roots list, this command does nothing.

Common Use Cases:
  • Add your main projects directory for discovery
  • Add multiple workspace directories for different project types
  • Include nested directories for specific project categories

Path Normalization:
  • Trailing slashes are automatically handled: ~/Projects/ → ~/Projects
  • Relative paths are converted to absolute: ./projects → /full/path/to/projects
  • Home directory expansion works: ~/Developer → /home/user/Developer
  • Multiple slashes are handled: /home//user/Projects → /home/user/Projects

Examples:
  # Add home projects directory
  portfolio config set-root ~/Projects

  # Add workspace directory with trailing slash
  portfolio config set-root /home/user/workspace/

  # Add relative path (gets converted to absolute)
  portfolio config set-root ./projects

  # Add specific workspace
  portfolio config set-root /home/user/Developer/frontend`,
	Args: cobra.ExactArgs(1),
	Run:  runSetRoot,
}

// removeRootCmd represents the config remove-root command
var removeRootCmd = &cobra.Command{
	Use:   "remove-root <path>",
	Short: "Remove a project root directory",
	Long: `Remove a directory path from the list of project roots.

The path is normalized before matching (handles trailing slashes, relative paths,
home directory expansion). If the path is not found in the project roots list,
this command does nothing but displays a message. At least one project root must
remain configured.

Common Use Cases:
  • Remove a workspace directory that's no longer needed
  • Clean up duplicate or outdated project roots
  • Reorganize your project discovery layout

Path Matching:
  • Paths are normalized before comparison
  • Must match exactly (after normalization) to be removed
  • Cannot remove the last remaining project root

Examples:
  # Remove home projects directory
  portfolio config remove-root ~/Projects

  # Remove specific workspace (with trailing slash handling)
  portfolio config remove-root /home/user/workspace/

  # Try removing a path that doesn't exist (shows helpful message)
  portfolio config remove-root /nonexistent/path

Notes:
  • Use 'portfolio config list-roots' to see current roots
  • Add new roots first before removing the last one
  • Case-sensitive matching on all platforms`,
	Args: cobra.ExactArgs(1),
	Run:  runRemoveRoot,
}

// listRootsCmd represents the config list-roots command
var listRootsCmd = &cobra.Command{
	Use:   "list-roots",
	Short: "List all configured project root directories",
	Long: `Display all configured project root directories with validation status.

Shows the complete list of directories where Portfolio searches for projects,
along with real-time validation of each path's accessibility. Each path is
checked for existence and accessibility to ensure discovery will work properly.

Output Format:
  • ✓ - Path is accessible and valid for discovery
  • ✗ - Path has issues (removed, unreadable, or not a directory)

Common Use Cases:
  • Verify your project roots are configured correctly
  • Check if workspace directories are still accessible
  • Troubleshoot discovery issues
  • Review your current discovery layout

Examples:
  # List all project roots with validation status
  portfolio config list-roots

  # Use with other commands to manage configuration
  portfolio config set-root ~/projects
  portfolio config list-roots

Notes:
  • Paths are displayed in the order they were added
  • Use this command after system changes to verify accessibility
  • Consider removing roots that show ✗ status
  • See 'portfolio discover' to start project discovery`,
	Args: cobra.NoArgs,
	Run:  runListRoots,
}

func runSetRoot(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger().With("config")

	path := args[0]

	// Normalize the path with comprehensive handling
	normalizedPath, err := normalizePath(path)
	if err != nil {
		handleConfigError(err, "path normalization failed")
		return
	}
	path = normalizedPath

	// Validate the path
	if err := validatePathForRoot(path); err != nil {
		handleConfigError(err, "path validation failed")
		return
	}

	provider := config.NewProvider("")

	err = provider.Update(func(cfg *models.Config) error {
		if containsRoot(cfg.Discovery.ProjectRoots, path) {
			return fmt.Errorf("already configured")
		}
		cfg.Discovery.ProjectRoots = append(cfg.Discovery.ProjectRoots, path)
		return nil
	})
	if err != nil {
		if err.Error() == "already configured" {
			fmt.Printf("Path already configured as project root: %s\n", path)
			return
		}
		handleConfigError(err, "failed to save configuration")
		return
	}

	cfg, _ := provider.Load()
	logger.Info("Project root added",
		models.Field{Key: "action", Value: "set-root"},
		models.Field{Key: "path", Value: path},
		models.Field{Key: "total_roots", Value: len(cfg.Discovery.ProjectRoots)},
	)

	fmt.Printf("✓ Added project root: %s\n", path)
	fmt.Printf("Total project roots: %d\n", len(cfg.Discovery.ProjectRoots))
}

func runRemoveRoot(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger().With("config")

	path := args[0]

	// Normalize the path with comprehensive handling
	normalizedPath, err := normalizePath(path)
	if err != nil {
		handleConfigError(err, "path normalization failed")
		return
	}
	path = normalizedPath

	provider := config.NewProvider("")
	cfg, err := provider.Load()
	if err != nil {
		handleConfigError(err, "failed to load configuration")
		return
	}

	// Find and remove the path
	found := false
	newRoots := make([]string, 0, len(cfg.Discovery.ProjectRoots))
	for _, root := range cfg.Discovery.ProjectRoots {
		if root == path {
			found = true
		} else {
			newRoots = append(newRoots, root)
		}
	}

	if !found {
		fmt.Printf("Path not found in project roots: %s\n", path)
		fmt.Printf("Current roots: %d\n", len(cfg.Discovery.ProjectRoots))
		return
	}

	// Check if this would leave no roots
	if len(newRoots) == 0 {
		fmt.Println("Cannot remove the last project root.")
		fmt.Println("At least one project root must be configured.")
		fmt.Println("Use 'portfolio config set-root <path>' to add a new root first.")
		return
	}

	// Update configuration
	cfg.Discovery.ProjectRoots = newRoots

	// Save configuration
	if err := provider.Save(cfg); err != nil {
		handleConfigError(err, "failed to save configuration")
		return
	}

	logger.Info("Project root removed",
		models.Field{Key: "action", Value: "remove-root"},
		models.Field{Key: "path", Value: path},
		models.Field{Key: "total_roots", Value: len(cfg.Discovery.ProjectRoots)},
	)

	fmt.Printf("✓ Removed project root: %s\n", path)
	fmt.Printf("Total project roots: %d\n", len(cfg.Discovery.ProjectRoots))
}

func runListRoots(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger().With("config")

	provider := config.NewProvider("")
	cfg, err := provider.Load()
	if err != nil {
		handleConfigError(err, "failed to load configuration")
		return
	}

	fmt.Println("Portfolio Project Roots")
	fmt.Println("======================")
	fmt.Println()

	if len(cfg.Discovery.ProjectRoots) == 0 {
		fmt.Println("No project roots configured.")
		fmt.Println("Use 'portfolio config set-root <path>' to add a project root.")
		return
	}

	for i, root := range cfg.Discovery.ProjectRoots {
		// Check if path is still accessible
		status := "✓"
		if err := validatePathForRoot(root); err != nil {
			status = "✗"
		}

		fmt.Printf("%d. %s %s\n", i+1, status, root)
	}

	fmt.Println()
	fmt.Printf("Total: %d project root(s)\n", len(cfg.Discovery.ProjectRoots))

	logger.Info("Project roots listed",
		models.Field{Key: "action", Value: "list-roots"},
		models.Field{Key: "count", Value: len(cfg.Discovery.ProjectRoots)},
	)
}

// validatePathForRoot validates a path for use as a project root
func validatePathForRoot(path string) error {
	// Check if path is empty
	if path == "" {
		return &ConfigPathError{
			Path:    path,
			Message: "path cannot be empty",
		}
	}

	// Clean the path
	path = filepath.Clean(path)

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return &ConfigPathError{
				Path:    path,
				Message: "cannot expand home directory",
				Cause:   err,
			}
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &ConfigPathError{
			Path:    path,
			Message: "path does not exist",
		}
	} else if err != nil {
		return &ConfigPathError{
			Path:    path,
			Message: "cannot access path",
			Cause:   err,
		}
	}

	// Check if it's a directory
	if !info.IsDir() {
		return &ConfigPathError{
			Path:    path,
			Message: "path is not a directory",
		}
	}

	// Check if readable
	if _, err := os.ReadDir(path); err != nil {
		return &ConfigPathError{
			Path:    path,
			Message: "path is not readable",
			Cause:   err,
		}
	}

	return nil
}

// containsRoot checks if a path exists in the roots list
func containsRoot(roots []string, path string) bool {
	for _, root := range roots {
		if root == path {
			return true
		}
	}
	return false
}

// handleConfigError handles configuration command errors
func handleConfigError(err error, message string) {
	logger := logging.GetGlobalLogger()

	// Log detailed error to error.log
	logger.LogErrorToFile("Config command failed", err,
		models.Field{Key: "message", Value: message})

	// Show user-friendly message on stderr
	fmt.Fprintf(os.Stderr, "\nError: %s\n", message)
	fmt.Fprintln(os.Stderr, "\nRun 'portfolio doctor' for diagnostics")
}

// ConfigPathError represents a path validation error
type ConfigPathError struct {
	Path    string
	Message string
	Cause   error
}

func (e *ConfigPathError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Message, e.Path, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Path)
}

// Unwrap returns the underlying cause
func (e *ConfigPathError) Unwrap() error {
	return e.Cause
}
