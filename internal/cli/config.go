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

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Portfolio configuration",
	Long: `Manage Portfolio Engine configuration settings.

This command provides subcommands for managing project roots, ignored paths,
and other configuration settings stored in the Portfolio configuration file.`,
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

The path must exist, be accessible, and be a directory. If the path is already
in the project roots list, this command does nothing.

Example:
  portfolio config set-root /home/user/Projects
  portfolio config set-root ~/Developer`,
	Args: cobra.ExactArgs(1),
	Run:  runSetRoot,
}

// removeRootCmd represents the config remove-root command
var removeRootCmd = &cobra.Command{
	Use:   "remove-root <path>",
	Short: "Remove a project root directory",
	Long: `Remove a directory path from the list of project roots.

The path must match exactly (including case) a path already in the project
roots list. If the path is not found, this command does nothing.

Example:
  portfolio config remove-root /home/user/Projects`,
	Args: cobra.ExactArgs(1),
	Run:  runRemoveRoot,
}

// listRootsCmd represents the config list-roots command
var listRootsCmd = &cobra.Command{
	Use:   "list-roots",
	Short: "List all configured project root directories",
	Long: `Display all configured project root directories.

Shows the complete list of directories where Portfolio searches for projects.
Each path is validated for existence and accessibility when configured.

Example:
  portfolio config list-roots`,
	Args: cobra.NoArgs,
	Run:  runListRoots,
}

func runSetRoot(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger().With("config")

	path := args[0]

	// Clean the path
	path = filepath.Clean(path)

	// Validate the path
	if err := validatePathForRoot(path); err != nil {
		handleConfigError(err, "path validation failed")
		return
	}

	// Load current configuration
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		handleConfigError(err, "failed to load configuration")
		return
	}

	// Check if path already exists in roots
	if containsRoot(cfg.Discovery.ProjectRoots, path) {
		fmt.Printf("Path already configured as project root: %s\n", path)
		return
	}

	// Add the path
	cfg.Discovery.ProjectRoots = append(cfg.Discovery.ProjectRoots, path)

	// Save configuration
	if err := loader.Save(cfg); err != nil {
		handleConfigError(err, "failed to save configuration")
		return
	}

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

	// Clean the path
	path = filepath.Clean(path)

	// Load current configuration
	loader := config.NewLoader("")
	cfg, err := loader.Load()
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
	if err := loader.Save(cfg); err != nil {
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

	// Load configuration
	loader := config.NewLoader("")
	cfg, err := loader.Load()
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
	logger.Error("Config command failed",
		models.Field{Key: "error", Value: err},
		models.Field{Key: "message", Value: message},
	)

	fmt.Fprintf(os.Stderr, "\nError: %s: %v\n", message, err)
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
