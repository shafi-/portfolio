package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <target>",
	Short: "Uninstall components",
	Long: `Uninstall components from Portfolio.

Supported targets:
  claude     Claude Code integration
  opencode   OpenCode integration`,
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

var uninstallOpencodeCmd = &cobra.Command{
	Use:   "opencode",
	Short: "Uninstall OpenCode integration",
	Long: `Uninstall OpenCode integration from Portfolio.

This command removes the Portfolio MCP entry from OpenCode's config
and the installed skill, while preserving all project data.`,
	Example: `  portfolio uninstall opencode`,
	Args:    cobra.NoArgs,
	Run:     runUninstallOpencode,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.AddCommand(uninstallClaudeCmd)
	uninstallCmd.AddCommand(uninstallOpencodeCmd)
}

func runUninstall(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "claude":
		runUninstallClaude(cmd, []string{})
	case "opencode":
		runUninstallOpencode(cmd, []string{})
	default:
		suggestAgentIntegration("uninstall", args[0])
		os.Exit(1)
	}
}

func runUninstallClaude(cmd *cobra.Command, args []string) { runUninstallIntegration(cmd, "claude") }
func runUninstallOpencode(cmd *cobra.Command, args []string) {
	runUninstallIntegration(cmd, "opencode")
}

func runUninstallIntegration(cmd *cobra.Command, name string) {
	logger := logging.GetGlobalLogger()
	ctx := cmd.Context()

	im, err := setupIntegrationManager(logger)
	if err != nil {
		logger.LogErrorToFile("Failed to setup integration manager", err)
		fmt.Fprintf(os.Stderr, "Error: could not initialize integration manager\n\nRun 'portfolio doctor' for diagnostics\n")
		os.Exit(1)
	}
	defer im.db.Close()

	if err := im.manager.Remove(ctx, name); err != nil {
		logger.LogErrorToFile("Uninstallation failed", err,
			models.Field{Key: "integration", Value: name})
		display := integrationDisplayName(name)
		fmt.Fprintf(os.Stderr, "Error: failed to uninstall %s integration\n\nRun 'portfolio doctor %s' for diagnostics\n", display, name)
		os.Exit(1)
	}

	display := integrationDisplayName(name)
	fmt.Printf("✓ %s integration uninstalled successfully\n", display)
	fmt.Println("  Project data preserved in database")
}
