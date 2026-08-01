package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Portfolio Engine system status",
	Long: `Display the current status of the Portfolio Engine including:
- Configuration file location and validity
- Database accessibility and project count
- Last discovery timestamp
- System health indicators`,
	Run: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	fmt.Println("Portfolio Engine Status")
	fmt.Println("======================")
	fmt.Println()

	// Check configuration
	configStatus := checkConfiguration(logger)
	fmt.Printf("Configuration: %s\n", configStatus)

	// Check database
	dbStatus, projectCount, lastDiscovery := checkDatabaseStatus(logger)
	fmt.Printf("Database: %s\n", dbStatus)

	if projectCount >= 0 {
		fmt.Printf("Projects Discovered: %d\n", projectCount)
	}

	if !lastDiscovery.IsZero() {
		fmt.Printf("Last Discovery: %s\n", lastDiscovery.Format(time.RFC3339))
	} else {
		fmt.Println("Last Discovery: Never")
	}

	fmt.Println()

	// Determine overall status
	overallStatus := "Running"
	if !strings.HasPrefix(configStatus, "✓") || !strings.HasPrefix(dbStatus, "✓") {
		overallStatus = "Degraded"
	}

	fmt.Printf("Engine Status: %s\n", overallStatus)
}

func checkConfiguration(logger *logging.Logger) string {
	provider := config.NewProvider("")

	if _, err := os.Stat(provider.ConfigPath()); os.IsNotExist(err) {
		logger.Warn("Configuration file not found",
			models.Field{Key: "path", Value: provider.ConfigPath()},
		)
		return "✗ Not found (run 'portfolio init')"
	}

	// Try to load configuration
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load configuration",
			models.Field{Key: "error", Value: err},
		)
		return "✗ Invalid (run 'portfolio doctor')"
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		logger.Error("Configuration validation failed",
			models.Field{Key: "error", Value: err},
		)
		return "✗ Invalid (run 'portfolio doctor')"
	}

	logger.Info("Configuration loaded successfully",
		models.Field{Key: "path", Value: provider.ConfigPath()},
		models.Field{Key: "project_roots", Value: len(cfg.Discovery.ProjectRoots)},
	)

	return fmt.Sprintf("✓ Accessible (%s)", provider.ConfigPath())
}

func checkDatabaseStatus(logger *logging.Logger) (string, int, time.Time) {
	provider := config.NewProvider("")
	cfg, err := provider.Load()
	if err != nil {
		logger.Error("Failed to load configuration for database check",
			models.Field{Key: "error", Value: err},
		)
		return "✗ Configuration error", -1, time.Time{}
	}

	dbPath := cfg.General.DatabasePath

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Warn("Database file not found",
			models.Field{Key: "path", Value: dbPath},
		)
		return "✗ Not found (run 'portfolio init')", -1, time.Time{}
	}

	// Try to connect to database
	db, err := database.NewDatabase(dbPath, logger)
	if err != nil {
		logger.Error("Failed to create database object",
			models.Field{Key: "error", Value: err},
		)
		return "✗ Inaccessible", -1, time.Time{}
	}

	// Actually connect to the database
	if err := db.Connect(); err != nil {
		logger.Error("Failed to connect to database",
			models.Field{Key: "error", Value: err},
		)
		return "✗ Inaccessible", -1, time.Time{}
	}
	defer db.Close()

	// Get project count
	projectCount, err := db.GetProjectCount()
	if err != nil {
		logger.Warn("Failed to get project count",
			models.Field{Key: "error", Value: err},
		)
		return "✓ Accessible (count unavailable)", 0, time.Time{}
	}

	// Get last discovery time
	lastDiscovery, err := db.GetLastDiscoveryTime()
	if err != nil {
		logger.Warn("Failed to get last discovery time",
			models.Field{Key: "error", Value: err},
		)
	}

	logger.Info("Database status retrieved",
		models.Field{Key: "path", Value: dbPath},
		models.Field{Key: "project_count", Value: projectCount},
		models.Field{Key: "last_discovery", Value: lastDiscovery},
	)

	return fmt.Sprintf("✓ Accessible (%s)", dbPath), projectCount, lastDiscovery
}
