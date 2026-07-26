package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"project-dash/internal/integration"
	"project-dash/internal/integration/claude"
	"project-dash/internal/integration/opencode"
	"project-dash/internal/logging"
	"project-dash/internal/version"
	"project-dash/pkg/models"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <target>",
	Short: "Upgrade components",
	Long: `Upgrade components for Portfolio.

Supported targets:
  claude     Claude Code integration
  opencode   OpenCode integration`,
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

var upgradeOpencodeCmd = &cobra.Command{
	Use:   "opencode",
	Short: "Upgrade OpenCode integration",
	Long: `Upgrade OpenCode integration to the latest version.

This command refreshes the OpenCode config entry and skill to the
latest version while preserving user settings.`,
	Example: `  portfolio upgrade opencode`,
	Args:    cobra.NoArgs,
	Run:     runUpgradeOpencode,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.AddCommand(upgradeClaudeCmd)
	upgradeCmd.AddCommand(upgradeOpencodeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "claude":
		runUpgradeClaude(cmd, []string{})
	case "opencode":
		runUpgradeOpencode(cmd, []string{})
	default:
		suggestAgentIntegration("upgrade", args[0])
		os.Exit(1)
	}
}

func runUpgradeClaude(cmd *cobra.Command, args []string)   { runUpgradeIntegration(cmd, "claude") }
func runUpgradeOpencode(cmd *cobra.Command, args []string) { runUpgradeIntegration(cmd, "opencode") }

func runUpgradeIntegration(cmd *cobra.Command, name string) {
	logger := logging.GetGlobalLogger()
	ctx := cmd.Context()

	im, err := setupIntegrationManager(logger)
	if err != nil {
		logger.Error(err.Error(), models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer im.db.Close()

	previousMeta, err := im.manager.Get(ctx, name)
	if err != nil {
		logger.Error("failed to get current integration", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: failed to get current integration: %v\n", err)
		os.Exit(1)
	}

	previousVersion := previousMeta.Version

	// Target the integration's own latest version constant.
	targetVersion := claude.Version
	if name == opencode.Name {
		targetVersion = opencode.Version
	}

	opts := integration.UpgradeOptions{
		TargetVersion: targetVersion,
		EngineVersion: version.Version(),
	}

	result, err := im.manager.Upgrade(ctx, name, opts)
	if err != nil {
		logger.Error("upgrade failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: upgrade failed: %v\n\nTo fix: Check integration status with 'portfolio doctor %s'\n", err, name)
		os.Exit(1)
	}

	display := integrationDisplayName(name)
	if result.Version == previousVersion {
		fmt.Printf("✓ %s integration already up to date (v%s)\n", display, result.Version)
	} else {
		fmt.Printf("✓ %s integration upgraded successfully\n", display)
		fmt.Printf("  Previous Version: %s\n", previousVersion)
		fmt.Printf("  New Version: %s\n", result.Version)
		fmt.Printf("  Updated At: %s\n", result.UpdatedAt)
	}
}
