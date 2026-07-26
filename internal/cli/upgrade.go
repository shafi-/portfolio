package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/integration"
	"project-dash/internal/integration/claude"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <target>",
	Short: "Upgrade components",
	Long: `Upgrade components for Portfolio.

Supported targets:
  claude     Claude Code integration`,
	Args: cobra.ExactArgs(1),
	Run:  runUpgrade,
}

var upgradeClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Upgrade Claude Code integration",
	Long: `Upgrade Claude Code integration to the latest version.

This command updates agent configuration and skills to the latest
version while preserving user settings.`,
	Example: `  portfolio upgrade claude`,
	Args:    cobra.NoArgs,
	Run:     runUpgradeClaude,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.AddCommand(upgradeClaudeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) {
	if args[0] == "claude" {
		runUpgradeClaude(cmd, []string{})
		return
	}
	suggestAgentIntegration("upgrade", args[0])
	os.Exit(1)
}

func runUpgradeClaude(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	ctx := cmd.Context()

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to create database: %v\n", err)
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	mcpClient := integration.NewStdioMCPClient(binaryPath, []string{"mcp"})
	store := integration.NewDatabaseStore(db.DB(), logger.Zap())

	manager := integration.NewManager(store, mcpClient, logger.Zap(), version)

	claudeIntegration, err := claude.New(store, mcpClient, logger.Zap())
	if err != nil {
		logger.Error("failed to create Claude integration", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to create Claude integration: %v\n", err)
		os.Exit(1)
	}
	manager.RegisterIntegration(claudeIntegration)

	previousMeta, err := manager.Get(ctx, "claude")
	if err != nil {
		logger.Error("failed to get current integration", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to get current integration: %v\n", err)
		os.Exit(1)
	}

	previousVersion := previousMeta.Version

	opts := integration.UpgradeOptions{
		TargetVersion: "1.0.0",
		EngineVersion: version,
	}

	result, err := manager.Upgrade(ctx, "claude", opts)
	if err != nil {
		logger.Error("upgrade failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: upgrade failed: %v\n\nTo fix: Check integration status with 'portfolio doctor claude'\n", err)
		os.Exit(1)
	}

	if result.Version == previousVersion {
		fmt.Printf("✓ Claude Code integration already up to date (v%s)\n", result.Version)
	} else {
		fmt.Printf("✓ Claude Code integration upgraded successfully\n")
		fmt.Printf("  Previous Version: %s\n", previousVersion)
		fmt.Printf("  New Version: %s\n", result.Version)
		fmt.Printf("  Updated At: %s\n", result.UpdatedAt)
	}
}
