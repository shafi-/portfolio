package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"project-dash/internal/integration"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var installCmd = &cobra.Command{
	Use:   "install <target>",
	Short: "Install components",
	Long: `Install components for Portfolio.

Supported targets:
  claude     Claude Code integration
  opencode   OpenCode integration

Other AI agents (e.g. Cline) have no official automated install.
Run 'portfolio install <agent>' for guidance, or run 'portfolio manual'.`,
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

var installOpencodeCmd = &cobra.Command{
	Use:   "opencode",
	Short: "Install OpenCode integration",
	Long: `Install OpenCode integration for Portfolio.

This command registers the Portfolio MCP server in OpenCode's config
(~/.config/opencode/opencode.json), installs the Portfolio skill, and
validates the installation. OpenCode must be installed first.`,
	Example: `  portfolio install opencode
  portfolio install opencode --force`,
	Args: cobra.NoArgs,
	Run:  runInstallOpencode,
}

var forceInstall bool

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installClaudeCmd)
	installCmd.AddCommand(installOpencodeCmd)

	installClaudeCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "force reinstall even if already installed")
	installOpencodeCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "force reinstall even if already installed")
}

func runInstall(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "claude":
		runInstallClaude(cmd, []string{})
	case "opencode":
		runInstallOpencode(cmd, []string{})
	default:
		suggestAgentIntegration("install", args[0])
		os.Exit(1)
	}
}

func runInstallClaude(cmd *cobra.Command, args []string) {
	runInstallIntegration(cmd, "claude")
}

func runInstallOpencode(cmd *cobra.Command, args []string) {
	runInstallIntegration(cmd, "opencode")
}

// runInstallIntegration installs the named integration via the shared manager.
func runInstallIntegration(cmd *cobra.Command, name string) {
	logger := logging.GetGlobalLogger()
	ctx := cmd.Context()

	im, err := setupIntegrationManager(logger)
	if err != nil {
		logger.Error(err.Error(), models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer im.db.Close()

	opts := integration.InstallOptions{
		Force: forceInstall,
	}

	result, err := im.manager.Install(ctx, name, opts)
	if err != nil {
		logger.Error("installation failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: installation failed: %v\n", err)
		os.Exit(1)
	}

	display := integrationDisplayName(name)
	fmt.Printf("✓ %s integration installed successfully\n", display)
	fmt.Printf("  Version: %s\n", result.Version)
	fmt.Printf("  Agent Type: %s\n", result.AgentType)
	fmt.Printf("  Installed At: %s\n", result.InstalledAt)
}
