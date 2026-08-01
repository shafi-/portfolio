package main

import (
	"fmt"
	"os"

	"project-dash/internal/cli"
	"project-dash/internal/errors"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

func main() {
	// Global panic handler for safe error recovery
	defer func() {
		if err := errors.SafePanicHandler(); err != nil {
			fmt.Fprintf(os.Stderr, "Internal error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Initialize logging first
	logConfig := logging.LoadConfigFromEnv()
	if err := logConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", errors.SanitizeError(err))
		os.Exit(1)
	}

	// Check if running in MCP mode - use stderr to keep stdout clean for JSON-RPC
	isMCPMode := len(os.Args) > 1 && os.Args[1] == "mcp"

	var logger *logging.Logger
	var err error
	if isMCPMode {
		logger, err = logging.NewStderrLogger(logConfig.Level, logConfig.Format)
	} else {
		// Engine logs are noisy on the happy path. Default to ERROR so a
		// successful command prints only its own output; --verbose restores the
		// configured level. Cobra parses flags inside cli.Execute() (after the
		// logger is built), so read --verbose from os.Args here.
		level := logConfig.Level
		if !cli.HasVerboseFlag(os.Args) {
			level = "ERROR"
		}

		// Log file always captures INFO level for debugging (issue attachment).
		// File path comes from config (default: <data-dir>/portfolio.log).
		logger, err = logging.NewLoggerWithFile(level, logConfig.Format, logConfig.File, os.Stdout)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging system\n")
		os.Exit(1)
	}

	logging.SetGlobalLogger(logger)
	logger.Info("Portfolio Engine starting")

	// Execute CLI
	if err := cli.Execute(); err != nil {
		// Check if it's already a safe error
		if safeErr, ok := err.(*errors.SafeError); ok {
			logger.Error("CLI execution failed",
				models.Field{Key: "error", Value: safeErr.UserMessage},
				models.Field{Key: "request_id", Value: safeErr.RequestID},
			)
		} else {
			logger.Error("CLI execution failed",
				models.Field{Key: "error", Value: errors.SanitizeError(err)},
			)
		}
		os.Exit(1)
	}

	// Ensure logs are flushed
	_ = logger.Sync()
}
