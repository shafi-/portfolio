package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"project-dash/internal/version"
	"project-dash/pkg/models"
)

var (
	// CLI flags
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Portfolio Engine - Local-first project inventory and knowledge platform",
	Long: `Portfolio is a local-first project inventory and knowledge platform that
enables developers and AI coding agents to understand an entire software portfolio.

The Portfolio Engine provides deterministic project discovery, metadata extraction,
and knowledge storage while maintaining clear separation between engine operations
and AI agent reasoning.

Primary interface through AI coding agents. CLI exists for administrative tasks only.`,
	Version: version.Full(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes flags and configuration
func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: <data-dir>/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show diagnostic logging (hidden by default; honors the configured [logging] level)")
	rootCmd.Flags().Bool("toggle", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Skip logging here - will be handled by individual commands
	// This prevents logs from going to stdout in MCP mode

	if cfgFile != "" {
		// Config file from flag
	} else {
		// Use the canonical config path from models package
		cfgFile = models.GetConfigPath()
	}

	// Note: Verbose logging is handled by individual commands
}

// GetRootCommand returns the root command for testing
func GetRootCommand() *cobra.Command {
	return rootCmd
}

// GenerateDocs generates markdown documentation for CLI commands
func GenerateDocs(dir string) error {
	return doc.GenMarkdownTree(rootCmd, dir)
}

// HasVerboseFlag reports whether args contain the --verbose/-v flag. It is used
// before cobra parses flags (in main, where the logger is constructed) to pick
// the initial log level: without --verbose the engine runs quiet (ERROR level),
// with it the configured level applies. The scan is intentionally shallow — it
// stops at "--" (end of flags) and does not try to model every subcommand's
// flag set, only the global persistent --verbose.
func HasVerboseFlag(args []string) bool {
	for _, a := range args {
		if a == "--" {
			break
		}
		switch {
		case a == "--verbose" || strings.HasPrefix(a, "--verbose="):
			return a != "--verbose=false"
		case a == "-v":
			return true
		case len(a) > 1 && a[0] == '-' && a[1] != '-' && strings.ContainsRune(a, 'v'):
			// Short-flag group containing v, e.g. -vv or -hv.
			return true
		}
	}
	return false
}
