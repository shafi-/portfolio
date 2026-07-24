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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <target>",
	Short: "Uninstall components",
	Long: `Uninstall components from Portfolio.

Supported targets:
  claude     Claude Code integration`,
	Args: cobra.ExactArgs(1),
	Run:  runUninstall,
}

var uninstallClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Uninstall Claude Code integration",
	Long: `Uninstall Claude Code integration from Portfolio.

This command removes agent-specific configuration and skills
while preserving all project data in the Portfolio database.`,
	Example: `  portfolio uninstall claude`,
	Args:    cobra.NoArgs,
	Run:     runUninstallClaude,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.AddCommand(uninstallClaudeCmd)
}

func runUninstall(cmd *cobra.Command, args []string) {
	if args[0] == "claude" {
		runUninstallClaude(cmd, []string{})
		return
	}
	fmt.Printf("Error: Unknown uninstall target '%s'\n", args[0])
	fmt.Println("Supported targets: claude")
	os.Exit(1)
}

func runUninstallClaude(cmd *cobra.Command, args []string) {
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

	if err := manager.Remove(ctx, "claude"); err != nil {
		logger.Error("uninstallation failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: uninstallation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Claude Code integration uninstalled successfully\n")
	fmt.Println("  Project data preserved in database")
}
