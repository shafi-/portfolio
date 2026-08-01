package main

import (
	"fmt"
	"os"
	"runtime"

	"project-dash/internal/cli"
	"project-dash/internal/errors"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

func main() {
	// Global panic handler for safe error recovery
	defer func() {
		if r := recover(); r != nil {
			// Log panic to error file for debugging
			logger := logging.GetGlobalLogger()
			if logger != nil {
				logger.LogErrorToFile("Panic recovered", fmt.Errorf("%v", r),
					models.Field{Key: "stack_trace", Value: string(getStackTrace())})
			}
			fmt.Fprintf(os.Stderr, "An internal error occurred.\n\nRun 'portfolio doctor' for diagnostics, or check error.log for details.\n")
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
		// Error file captures ERROR+ level with full stack traces.
		logger, err = logging.NewLoggerWithFiles(level, logConfig.Format, logConfig.File, logConfig.ErrorFile, os.Stderr)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logging system\n")
		os.Exit(1)
	}

	logging.SetGlobalLogger(logger)
	logger.Info("Portfolio Engine starting")

	// Execute CLI
	if err := cli.Execute(); err != nil {
		// Log detailed error to error.log
		logger.LogErrorToFile("CLI execution failed", err,
			models.Field{Key: "command", Value: os.Args})

		// Show user-friendly message on stderr with guidance
		if safeErr, ok := err.(*errors.SafeError); ok {
			fmt.Fprintf(os.Stderr, "%s\n\nRun 'portfolio doctor' for diagnostics\n", safeErr.UserMessage)
		} else {
			// For non-safe errors, show sanitized message
			sanitized := errors.SanitizeError(err)
			if sanitized == "An error occurred" {
				fmt.Fprintf(os.Stderr, "An unexpected error occurred.\n\nRun 'portfolio doctor' for diagnostics\n")
			} else {
				fmt.Fprintf(os.Stderr, "%s\n\nRun 'portfolio doctor' for diagnostics\n", sanitized)
			}
		}
		os.Exit(1)
	}

	// Ensure logs are flushed
	_ = logger.Sync()
}

// getStackTrace returns the current stack trace for panic logging
func getStackTrace() []byte {
	// Use runtime to get stack trace
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
