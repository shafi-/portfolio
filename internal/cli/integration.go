package cli

import (
	"context"
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

var (
	forceInstall bool
)

var integrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Manage agent integrations",
	Long: `Manage agent integrations for Portfolio.

Integrations connect AI coding agents (like Claude Code) to the Portfolio
Engine via MCP. Each integration is an installable package that handles
agent-specific configuration and setup.`,
}

var installCmd = &cobra.Command{
	Use:   "install <agent>",
	Short: "Install an agent integration",
	Long: `Install an agent integration for Portfolio.

Supported agents:
  claude     Claude Code integration

The install command registers the Portfolio MCP server with the agent,
installs agent-specific skills, and validates the installation.`,
	Args: cobra.ExactArgs(1),
	Run:  runInstall,
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <agent>",
	Short: "Uninstall an agent integration",
	Long: `Uninstall an agent integration from Portfolio.

Supported agents:
  claude     Claude Code integration

The uninstall command removes agent-specific configuration and skills
while preserving all project data in the Portfolio database.`,
	Args: cobra.ExactArgs(1),
	Run:  runUninstall,
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <agent>",
	Short: "Upgrade an agent integration",
	Long: `Upgrade an agent integration to the latest version.

Supported agents:
  claude     Claude Code integration

The upgrade command updates agent configuration and skills to the latest
version while preserving user settings.`,
	Args: cobra.ExactArgs(1),
	Run:  runUpgrade,
}

var integrationDoctorCmd = &cobra.Command{
	Use:   "doctor [agent]",
	Short: "Check integration health",
	Long: `Check the health of installed integrations.

If an agent name is specified, only that integration is checked.
If no agent is specified, all installed integrations are checked.

Supported agents:
  claude     Claude Code integration`,
	Run: runIntegrationDoctor,
}

func init() {
	rootCmd.AddCommand(integrationCmd)
	integrationCmd.AddCommand(installCmd)
	integrationCmd.AddCommand(uninstallCmd)
	integrationCmd.AddCommand(upgradeCmd)
	integrationCmd.AddCommand(integrationDoctorCmd)

	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "force reinstall even if already installed")
}

func runInstall(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	ctx := context.Background()

	agent := args[0]

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	mcpClient := integration.NewStdioMCPClient(binaryPath, []string{"mcp"})
	store := integration.NewDatabaseStore(db.DB(), logger.Zap())

	manager := integration.NewManager(store, mcpClient, logger.Zap(), version)

	var agentIntegration integration.Integration
	switch agent {
	case "claude":
		agentIntegration, err = claude.New(store, mcpClient, logger.Zap())
		if err != nil {
			logger.Error("failed to create Claude integration", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
		manager.RegisterIntegration(agentIntegration)
	default:
		fmt.Printf("Error: Unknown agent '%s'\n", agent)
		fmt.Println("Supported agents: claude")
		os.Exit(1)
	}

	opts := integration.InstallOptions{
		Force: forceInstall,
	}

	meta, err := manager.Install(ctx, agent, opts)
	if err != nil {
		logger.Error("installation failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Integration installed: %s (v%s)\n", meta.Name, meta.Version)
	fmt.Printf("  Agent Type: %s\n", meta.AgentType)
	fmt.Printf("  Installed At: %s\n", meta.InstalledAt)
}

func runUninstall(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	ctx := context.Background()

	agent := args[0]

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	mcpClient := integration.NewStdioMCPClient(binaryPath, []string{"mcp"})
	store := integration.NewDatabaseStore(db.DB(), logger.Zap())

	manager := integration.NewManager(store, mcpClient, logger.Zap(), version)

	var agentIntegration integration.Integration
	switch agent {
	case "claude":
		agentIntegration, err = claude.New(store, mcpClient, logger.Zap())
		if err != nil {
			logger.Error("failed to create Claude integration", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
		manager.RegisterIntegration(agentIntegration)
	default:
		fmt.Printf("Error: Unknown agent '%s'\n", agent)
		fmt.Println("Supported agents: claude")
		os.Exit(1)
	}

	if err := manager.Remove(ctx, agent); err != nil {
		logger.Error("uninstallation failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Integration uninstalled: %s\n", agent)
	fmt.Println("  Project data preserved in database")
}

func runUpgrade(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	ctx := context.Background()

	agent := args[0]

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	mcpClient := integration.NewStdioMCPClient(binaryPath, []string{"mcp"})
	store := integration.NewDatabaseStore(db.DB(), logger.Zap())

	manager := integration.NewManager(store, mcpClient, logger.Zap(), version)

	var agentIntegration integration.Integration
	switch agent {
	case "claude":
		agentIntegration, err = claude.New(store, mcpClient, logger.Zap())
		if err != nil {
			logger.Error("failed to create Claude integration", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
		manager.RegisterIntegration(agentIntegration)
	default:
		fmt.Printf("Error: Unknown agent '%s'\n", agent)
		fmt.Println("Supported agents: claude")
		os.Exit(1)
	}

	opts := integration.UpgradeOptions{
		TargetVersion: "1.0.0",
	}

	result, err := manager.Upgrade(ctx, agent, opts)
	if err != nil {
		logger.Error("upgrade failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if result.Version == "1.0.0" {
		fmt.Printf("✓ Integration already up to date: %s (v%s)\n", agent, result.Version)
	} else {
		fmt.Printf("✓ Integration upgraded: %s (v%s)\n", agent, result.Version)
		fmt.Printf("  Updated at: %s\n", result.UpdatedAt)
	}
}

func runIntegrationDoctor(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()
	ctx := context.Background()

	agent := ""
	if len(args) > 0 {
		agent = args[0]
	}

	loader := config.NewLoader(cfgFile)
	cfg, err := loader.Load()
	if err != nil {
		logger.Error("failed to load config", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to create database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	if err := db.Connect(); err != nil {
		logger.Error("failed to connect to database", models.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		logger.Error("failed to initialize database", models.Field{Key: "error", Value: err})
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
		os.Exit(1)
	}
	manager.RegisterIntegration(claudeIntegration)

	result, err := manager.Doctor(ctx, agent, false)
	if err != nil {
		logger.Error("doctor check failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if result.Passed {
		fmt.Println("✓ All integration checks passed")
		os.Exit(0)
	}

	fmt.Println("✗ Integration health check failed")
	fmt.Println()

	for _, check := range result.Checks {
		if check.Passed {
			fmt.Printf("  ✓ %s: %s\n", check.Name, check.Message)
		} else {
			fmt.Printf("  ✗ %s: %s\n", check.Name, check.Message)
			if check.Remediation != "" {
				fmt.Printf("    → %s\n", check.Remediation)
			}
		}
	}

	os.Exit(1)
}
