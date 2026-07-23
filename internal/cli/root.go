package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/nerddevsltd/portfolio/internal/logging"
	"github.com/nerddevsltd/portfolio/pkg/models"
)

var (
	// Version information
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"

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
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes flags and configuration
func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.portfolio/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().Bool("toggle", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Get global logger
	logger := logging.GetGlobalLogger()

	if cfgFile != "" {
		// Use config file from the flag
		logger.Info("Using config file from flag",
			models.Field{Key: "config", Value: cfgFile},
		)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory
		cfgFile = fmt.Sprintf("%s/.portfolio/config.toml", home)
	}

	if verbose {
		logger.Info("Verbose mode enabled")
	}
}

// GetRootCommand returns the root command for testing
func GetRootCommand() *cobra.Command {
	return rootCmd
}

// GenerateDocs generates markdown documentation for CLI commands
func GenerateDocs(dir string) error {
	return doc.GenMarkdownTree(rootCmd, dir)
}
