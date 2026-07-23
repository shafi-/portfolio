package main

import (
	"fmt"
	"os"

	"project-dash/internal/cli"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

func main() {
	// Initialize logging first
	logConfig := logging.LoadConfigFromEnv()
	if err := logConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logging configuration: %v\n", err)
		os.Exit(1)
	}

	if err := logging.InitializeGlobalLogger(logConfig.Level, logConfig.Format); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger := logging.GetGlobalLogger()
	logger.Info("Portfolio Engine starting",
		models.Field{Key: "version", Value: "0.1.0"},
	)

	// Execute CLI
	if err := cli.Execute(); err != nil {
		logger.Error("CLI execution failed",
			models.Field{Key: "error", Value: err},
		)
		os.Exit(1)
	}

	// Ensure logs are flushed
	logger.Sync()
}
