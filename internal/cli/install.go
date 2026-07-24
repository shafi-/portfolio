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

var installCmd = &cobra.Command{
	Use:   "install <target>",
	Short: "Install components",
	Long: `Install components for Portfolio.

Supported targets:
  claude     Claude Code integration`,
	Args: cobra.ExactArgs(1),
	Run:  runInstall,
}

var installClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Install Claude Code integration",
	Long: `Install Claude Code integration for Portfolio.

This command registers the Portfolio MCP server with Claude Code,
installs agent-specific skills, and validates the installation.`,
	Example: `  portfolio install claude
  portfolio install claude --force`,
	Args: cobra.NoArgs,
	Run:  runInstallClaude,
}

var forceInstall bool

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installClaudeCmd)

	installClaudeCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "force reinstall even if already installed")
}

func runInstall(cmd *cobra.Command, args []string) {
	if args[0] == "claude" {
		runInstallClaude(cmd, []string{})
		return
	}
	fmt.Printf("Error: Unknown install target '%s'\n", args[0])
	fmt.Println("Supported targets: claude")
	os.Exit(1)
}

func runInstallClaude(cmd *cobra.Command, args []string) {
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
		fmt.Printf("Error: failed to create Claude integration: %v\n\nTo fix: Check Claude Code is installed: https://claude.ai/download\n", err)
		os.Exit(1)
	}
	manager.RegisterIntegration(claudeIntegration)

	opts := integration.InstallOptions{
		Force: forceInstall,
	}

	result, err := manager.Install(ctx, "claude", opts)
	if err != nil {
		logger.Error("installation failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: installation failed: %v\n\nDiagnostics: %s\nTo fix: Check Claude Code is installed: https://claude.ai/download\n", err, "")
		os.Exit(1)
	}

	fmt.Printf("✓ Claude Code integration installed successfully\n")
	fmt.Printf("  Version: %s\n", result.Version)
	fmt.Printf("  Agent Type: %s\n", result.AgentType)
	fmt.Printf("  Installed At: %s\n", result.InstalledAt)
}
