package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nerddevsltd/portfolio/internal/config"
	"github.com/nerddevsltd/portfolio/internal/database"
	"github.com/nerddevsltd/portfolio/internal/logging"
	"github.com/nerddevsltd/portfolio/pkg/models"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Portfolio Engine with interactive setup",
	Long: `Initialize Portfolio Engine by creating configuration file and database.

This command guides you through setting up your Portfolio configuration:
- Project root directories for discovery
- Database file location
- Logging preferences

After initialization, Portfolio will be ready for project discovery and analysis.`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	logger.Info("Starting Portfolio initialization",
		models.Field{Key: "component", Value: "cli"},
	)

	fmt.Println("Portfolio Engine Initialization")
	fmt.Println("==============================")
	fmt.Println()

	// Step 1: Project roots
	projectRoots, err := promptProjectRoots()
	if err != nil {
		handleInitError(err, "Failed to get project roots")
		return
	}

	// Step 2: Database path
	databasePath, err := promptDatabasePath()
	if err != nil {
		handleInitError(err, "Failed to get database path")
		return
	}

	// Step 3: Log level
	logLevel, err := promptLogLevel()
	if err != nil {
		handleInitError(err, "Failed to get log level")
		return
	}

	// Step 4: Confirmation
	if !confirmConfiguration(projectRoots, databasePath, logLevel) {
		fmt.Println("\nInitialization cancelled.")
		return
	}

	// Step 5: Create configuration
	fmt.Println("\nCreating configuration...")

	cfg := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: databasePath,
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: projectRoots,
			IgnoredPaths: []string{
				"node_modules", ".git", "vendor", "build", "dist", "target", "bin",
			},
		},
		Logging: models.LoggingConfig{
			Level: logLevel,
		},
	}

	manager := config.NewManager("")
	if err := manager.CreateDefaultConfig(); err != nil {
		handleInitError(err, "Failed to create configuration")
		return
	}

	// Update config with user values
	loader := config.NewLoader("")
	if err := loader.Save(cfg); err != nil {
		handleInitError(err, "Failed to save configuration")
		return
	}

	fmt.Printf("✓ Configuration created: %s\n", models.GetConfigPath())

	// Step 6: Initialize database
	fmt.Println("Initializing database...")

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		handleInitError(err, "Failed to initialize database")
		return
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		handleInitError(err, "Failed to initialize database schema")
		return
	}

	fmt.Printf("✓ Database initialized: %s\n", cfg.General.DatabasePath)

	fmt.Println("\n✓ Portfolio Engine initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'portfolio status' to verify installation")
	fmt.Println("  2. Run 'portfolio doctor' for system diagnostics")
	fmt.Println("  3. Start discovering projects in your configured roots")
}

func promptProjectRoots() ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	var roots []string

	fmt.Println("Enter project root directories (one per line, empty line to finish):")
	fmt.Println("Example: /home/user/developer or /Users/developer/Projects")

	for {
		fmt.Print("Project root: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			if len(roots) == 0 {
				fmt.Println("At least one project root is required.")
				continue
			}
			break
		}

		// Validate path
		if err := validatePath(input); err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}

		roots = append(roots, input)
		fmt.Printf("Added: %s\n", input)
	}

	return roots, nil
}

func promptDatabasePath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, ".portfolio", "portfolio.db")

	fmt.Printf("\nEnter database path (default: %s):\n", defaultPath)
	fmt.Print("Database path: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultPath, nil
	}

	return input, nil
}

func promptLogLevel() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nEnter log level (default: INFO):")
	fmt.Println("Options: DEBUG, INFO, WARN, ERROR")
	fmt.Print("Log level: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToUpper(input))
	if input == "" {
		return "INFO", nil
	}

	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for _, level := range validLevels {
		if input == level {
			return input, nil
		}
	}

	return "INFO", nil // default to INFO for invalid input
}

func confirmConfiguration(roots []string, dbPath, logLevel string) bool {
	fmt.Println("\nConfiguration Summary:")
	fmt.Println("=======================")
	fmt.Println("Project Roots:")
	for _, root := range roots {
		fmt.Printf("  - %s\n", root)
	}
	fmt.Printf("\nDatabase: %s\n", dbPath)
	fmt.Printf("Log Level: %s\n", logLevel)

	fmt.Print("\nProceed with initialization? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func validatePath(path string) error {
	// Clean the path
	path = filepath.Clean(path)

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	} else if err != nil {
		return fmt.Errorf("cannot access path: %s", path)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

func handleInitError(err error, message string) {
	logger := logging.GetGlobalLogger()
	logger.Error("Initialization failed",
		models.Field{Key: "error", Value: err},
		models.Field{Key: "message", Value: message},
	)

	fmt.Fprintf(os.Stderr, "\nError: %s: %v\n", message, err)
	fmt.Fprintln(os.Stderr, "\nRun 'portfolio doctor' for diagnostics")
}
