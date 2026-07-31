package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Portfolio Engine Initialization")
	fmt.Println("==============================")
	fmt.Println()

	// Check for existing configuration
	loader := config.NewLoader("")
	existingCfg, configErr := loader.Load()

	isReinit := configErr == nil && existingCfg != nil && len(existingCfg.Discovery.ProjectRoots) > 0

	var projectRoots []string
	var databasePath string
	var isDefaultDB bool
	var logLevel string

	if isReinit {
		// Re-init: show existing roots, offer to add more
		fmt.Println("Existing configuration found.")
		fmt.Println("\nCurrent project roots:")
		for _, root := range existingCfg.Discovery.ProjectRoots {
			fmt.Printf("  - %s\n", root)
		}

		fmt.Print("\nAdd more project roots? (y/N): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			handleInitError(err, "Failed to read input")
			return
		}
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			newRoots, err := promptProjectRoots(reader)
			if err != nil {
				handleInitError(err, "Failed to get project roots")
				return
			}
			projectRoots = append(existingCfg.Discovery.ProjectRoots, newRoots...)
		} else {
			projectRoots = existingCfg.Discovery.ProjectRoots
		}

		databasePath = existingCfg.General.DatabasePath
		logLevel = existingCfg.Logging.Level
	} else {
		// Fresh init
		var err error
		projectRoots, err = promptProjectRoots(reader)
		if err != nil {
			handleInitError(err, "Failed to get project roots")
			return
		}

		databasePath, isDefaultDB, err = promptDatabasePath(reader)
		if err != nil {
			handleInitError(err, "Failed to get database path")
			return
		}

		logLevel, err = promptLogLevel(reader)
		if err != nil {
			handleInitError(err, "Failed to get log level")
			return
		}
	}

	// Step: Confirmation
	if !confirmConfiguration(projectRoots, databasePath, isDefaultDB, logLevel, isReinit, reader) {
		fmt.Println("\nInitialization cancelled.")
		return
	}

	// Step: Create/update configuration
	fmt.Println("\nSaving configuration...")

	cfg := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: databasePath,
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: projectRoots,
			IgnoredPaths: existingCfg.Discovery.IgnoredPaths,
		},
		Logging: models.LoggingConfig{
			Level: logLevel,
			File:  existingCfg.Logging.File,
		},
		Dashboard: existingCfg.Dashboard,
	}

	if existingCfg != nil && configErr == nil {
		cfg.Discovery.IgnoredPaths = existingCfg.Discovery.IgnoredPaths
		cfg.Logging.File = existingCfg.Logging.File
	} else {
		cfg.Discovery.IgnoredPaths = models.DefaultIgnoredPaths()
		cfg.Logging.File = models.GetDefaultLogPath()
	}

	if err := config.EnsureConfigDir(); err != nil {
		handleInitError(err, "Failed to create config directory")
		return
	}

	if err := loader.Save(cfg); err != nil {
		handleInitError(err, "Failed to save configuration")
		return
	}

	fmt.Printf("✓ Configuration saved: %s\n", models.GetConfigPath())

	// Step: Initialize database (skip if already exists)
	dbExists := false
	if _, err := os.Stat(cfg.General.DatabasePath); err == nil {
		dbExists = true
	}

	if dbExists {
		fmt.Printf("✓ Database already exists: %s\n", cfg.General.DatabasePath)
	} else {
		fmt.Println("Initializing database...")

		db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
		if err != nil {
			handleInitError(err, "Failed to initialize database")
			return
		}
		defer db.Close()

		if err := db.Connect(); err != nil {
			handleInitError(err, "Failed to connect to database")
			return
		}

		if err := db.Initialize(); err != nil {
			handleInitError(err, "Failed to initialize database schema")
			return
		}

		fmt.Printf("✓ Database initialized: %s\n", cfg.General.DatabasePath)
	}

	fmt.Println("\n✓ Portfolio Engine initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'portfolio status' to verify installation")
	fmt.Println("  2. Run 'portfolio doctor' for system diagnostics")
	fmt.Println("  3. Start discovering projects in your configured roots")
}

func promptProjectRoots(reader *bufio.Reader) ([]string, error) {
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

func promptDatabasePath(reader *bufio.Reader) (string, bool, error) {
	defaultPath := models.GetDefaultDatabasePath()

	fmt.Print("\nEnter database path (or press Enter for default): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultPath, true, nil
	}

	return input, false, nil
}

func promptLogLevel(reader *bufio.Reader) (string, error) {
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

func confirmConfiguration(roots []string, dbPath string, isDefaultDB bool, logLevel string, isReinit bool, reader *bufio.Reader) bool {
	fmt.Println("\nConfiguration Summary:")
	fmt.Println("=======================")
	if isReinit {
		fmt.Println("Mode: Re-init (updating existing configuration)")
	}
	fmt.Println("Project Roots:")
	for _, root := range roots {
		fmt.Printf("  - %s\n", root)
	}
	if isDefaultDB {
		fmt.Printf("\nDatabase: default\n")
	} else {
		fmt.Printf("\nDatabase: %s\n", dbPath)
	}
	fmt.Printf("Log Level: %s\n", logLevel)

	fmt.Print("\nProceed with initialization? (y/N): ")
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
