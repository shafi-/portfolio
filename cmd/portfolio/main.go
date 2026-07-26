package main

import (
	"fmt"
	"os"

	"project-dash/internal/cli"
	"project-dash/internal/logging"
	"project-dash/internal/version"
	"project-dash/pkg/models"
)

func main() {
	// Initialize logging first
	logConfig := logging.LoadConfigFromEnv()
	if err := logConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logging configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if running in MCP mode - use stderr to keep stdout clean for JSON-RPC
	isMCPMode := len(os.Args) > 1 && os.Args[1] == "mcp"

	var logger *logging.Logger
	var err error
	if isMCPMode {
		logger, err = logging.NewStderrLogger(logConfig.Level, logConfig.Format)
	} else {
		logger, err = logging.NewLogger(logConfig.Level, logConfig.Format)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logging.SetGlobalLogger(logger)
	logger.Info("Portfolio Engine starting",
		models.Field{Key: "version", Value: version.Version()},
	)

	// Execute CLI
	if err := cli.Execute(); err != nil {
		logger.Error("CLI execution failed",
			models.Field{Key: "error", Value: err},
		)
		os.Exit(1)
	}

	// Ensure logs are flushed
	_ = logger.Sync()
}
