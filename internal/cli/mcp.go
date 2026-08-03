package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/internal/mcp"
	"project-dash/pkg/models"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP stdio server for AI agent integration",
	Long: `Start the Portfolio MCP server on stdio transport.

The MCP server provides tools for AI coding agents to discover, search,
and analyze projects in the portfolio. Uses stdin/stdout for communication
per the MCP specification.`,
	Run: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) {
	// Use stderr logger for MCP to keep stdout clean for JSON-RPC messages
	logger, err := logging.NewStderrLogger("INFO", "console")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Set global logger to stderr logger for any components that use GetGlobalLogger
	logging.SetGlobalLogger(logger)

	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
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

	mcpCfg := &mcp.Config{
		DB:     db.DB(),
		Logger: logger,
		Roots:  cfg.Discovery.ProjectRoots,
	}

	srv := mcp.New(mcpCfg)

	logger.Info("MCP server starting on stdio")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := srv.Serve(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("MCP server stopped")
		} else {
			logger.Error("MCP server error", models.Field{Key: "error", Value: err})
			os.Exit(1)
		}
	}
}
