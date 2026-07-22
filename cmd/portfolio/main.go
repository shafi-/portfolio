package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/nerddevsltd/portfolio/internal/config"
	"github.com/nerddevsltd/portfolio/internal/database"
	"github.com/nerddevsltd/portfolio/internal/logging"
	"github.com/nerddevsltd/portfolio/pkg/models"
)

var (
	// Version is the application version
	Version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Portfolio - Local-first project inventory and knowledge platform",
	Long: `Portfolio is a local-first project inventory and knowledge platform
that enables developers and AI coding agents to understand an entire
software portfolio.`,
	Version: Version,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Portfolio configuration",
	Long:  `Initialize Portfolio with default configuration and prompt for project roots.`,
	RunE: runInit,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Portfolio engine status",
	Long:  `Display the current status of the Portfolio engine including database health and project count.`,
	RunE: runStatus,
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics",
	Long:  `Run diagnostics to check Portfolio configuration, database access, and system health.`,
	RunE: runDoctor,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doctorCmd)
}

func main() {
	// Initialize logging for CLI output
	logConfig := logging.LoadConfigFromEnv()
	if err := logConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logging configuration: %v\n", err)
		os.Exit(1)
	}

	logger, err := logging.NewLogger(logConfig.Level, logConfig.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Store logger globally for use in commands
	if err := logging.InitializeGlobalLogger(logConfig.Level, logConfig.Format); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize global logger: %v\n", err)
		os.Exit(1)
	}

	// Ensure logs are flushed on exit
	defer logger.Sync()

	// Execute CLI
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command failed", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
}

// runInit initializes Portfolio configuration
func runInit(cmd *cobra.Command, args []string) error {
	logger := logging.GetGlobalLogger()
	logger.Info("Initializing Portfolio", models.Field{Key: "version", Value: Version})

	// Create config loader
	loader := config.NewLoader("")

	// Load or create default config
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Prompt for project roots
	fmt.Println("\nPortfolio Configuration")
	fmt.Println("====================")
	fmt.Printf("\nDatabase: %s\n", cfg.General.DatabasePath)
	fmt.Println("\nCurrent project roots:")

	if len(cfg.Discovery.ProjectRoots) == 0 {
		fmt.Println("  (none configured)")
	} else {
		for _, root := range cfg.Discovery.ProjectRoots {
			fmt.Printf("  - %s\n", root)
		}
	}

	fmt.Println("\nProject roots will be scanned for Git repositories.")
	fmt.Println("Enter paths one at a time, or leave empty to finish.")

	for {
		var path string
		fmt.Print("\nAdd project root (or press Enter to finish): ")
		fmt.Scanln(&path)

		if path == "" {
			break
		}

		// Validate path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Error: Path '%s' does not exist.\n", path)
			continue
		}

		cfg.Discovery.ProjectRoots = append(cfg.Discovery.ProjectRoots, path)
		fmt.Printf("Added: %s\n", path)
	}

	// Save configuration
	if err := loader.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", models.GetConfigPath())
	fmt.Println("\nPortfolio initialized successfully!")
	fmt.Println("Run 'portfolio status' to verify the setup.")
	fmt.Println("Run 'portfolio discover' to scan for projects (when implemented).")

	return nil
}

// runStatus shows Portfolio engine status
func runStatus(cmd *cobra.Command, args []string) error {
	logger := logging.GetGlobalLogger()

	fmt.Println("\nPortfolio Engine Status")
	fmt.Println("=======================")

	// Load configuration
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check database health
	db, err := database.New(cfg.General.DatabasePath)
	if err != nil {
		logger.Error("Failed to open database", models.Field{Key: "error", Value: err})
		fmt.Printf("\nDatabase: ERROR - %v\n", err)
		return err
	}
	defer db.Close()

	// Get migration status
	migrations, err := db.GetMigrationStatus()
	if err != nil {
		logger.Error("Failed to get migration status", models.Field{Key: "error", Value: err})
	} else {
		fmt.Printf("\nDatabase: %s\n", cfg.General.DatabasePath)
		fmt.Printf("  Migrations applied: %d\n", len(migrations))
		fmt.Printf("  Health: %s\n", map[bool]string{true: "HEALTHY", false: "UNHEALTHY"}[db.IsHealthy()])
	}

	// Show project roots
	fmt.Printf("\nProject Roots: %d configured\n", len(cfg.Discovery.ProjectRoots))
	for _, root := range cfg.Discovery.ProjectRoots {
		fmt.Printf("  - %s\n", root)
	}

	// TODO: Add project count when discovery is implemented
	fmt.Println("\nProjects: Discovery not yet implemented")
	logFormat := logging.GetLogFormatFromEnv()
	fmt.Printf("\nLogging: %s (%s format)\n", cfg.Logging.Level, logFormat)

	return nil
}

// runDoctor runs diagnostics
func runDoctor(cmd *cobra.Command, args []string) error {
	logger := logging.GetGlobalLogger()

	fmt.Println("\nPortfolio Diagnostics")
	fmt.Println("====================")

	allHealthy := true

	// Check 1: Configuration file
	fmt.Print("\n[1/4] Configuration... ")
	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("  Failed to load configuration: %v\n", err)
		allHealthy = false
	} else {
		fmt.Println("OK")
		fmt.Printf("  Config path: %s\n", models.GetConfigPath())
		fmt.Printf("  Database: %s\n", cfg.General.DatabasePath)
		fmt.Printf("  Project roots: %d\n", len(cfg.Discovery.ProjectRoots))
	}

	// Check 2: Database access
	fmt.Print("\n[2/4] Database... ")
	if cfg != nil {
		db, err := database.New(cfg.General.DatabasePath)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Printf("  Failed to open database: %v\n", err)
			allHealthy = false
		} else {
			defer db.Close()
			if db.IsHealthy() {
				fmt.Println("OK")
				migrations, _ := db.GetMigrationStatus()
				fmt.Printf("  Migrations: %d applied\n", len(migrations))
			} else {
				fmt.Println("UNHEALTHY")
				allHealthy = false
			}
		}
	}

	// Check 3: Logging system
	fmt.Print("\n[3/4] Logging... ")
	logConfig := logging.LoadConfigFromEnv()
	if err := logConfig.Validate(); err != nil {
		fmt.Println("ERROR")
		fmt.Printf("  Invalid logging config: %v\n", err)
		allHealthy = false
	} else {
		fmt.Println("OK")
		logger.Info("Diagnostic log test successful")
		fmt.Printf("  Level: %s, Format: %s\n", logConfig.Level, logConfig.Format)
	}

	// Check 4: Project roots accessibility
	fmt.Print("\n[4/4] Project roots... ")
	if cfg != nil {
		accessibleCount := 0
		for _, root := range cfg.Discovery.ProjectRoots {
			if _, err := os.Stat(root); err == nil {
				accessibleCount++
			} else {
				fmt.Printf("\n  Warning: '%s' is not accessible\n", root)
			}
		}
		if accessibleCount == len(cfg.Discovery.ProjectRoots) {
			fmt.Println("OK")
		} else {
			fmt.Printf("WARNING (%d/%d accessible)\n", accessibleCount, len(cfg.Discovery.ProjectRoots))
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	if allHealthy {
		fmt.Println("Status: All systems operational")
		logger.Info("Diagnostics passed", models.Field{Key: "result", Value: "healthy"})
	} else {
		fmt.Println("Status: Issues detected - please review above")
		logger.Warn("Diagnostics found issues", models.Field{Key: "result", Value: "unhealthy"})
	}

	return nil
}
