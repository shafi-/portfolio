package cli

import (
	"fmt"
	"os"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/integration"
	"project-dash/internal/integration/claude"
	"project-dash/internal/integration/opencode"
	"project-dash/internal/logging"
	"project-dash/internal/version"
)

// integrationManager bundles a configured Manager with the database handle it
// was built from, so callers can defer Close().
type integrationManager struct {
	manager *integration.Manager
	db      *database.Database
}

// setupIntegrationManager builds a Manager with every supported integration
// (currently Claude Code and OpenCode) registered, using the running binary as
// the stdio MCP server. It centralises the config→database→manager wiring that
// install/uninstall/upgrade/doctor otherwise duplicate. Callers must defer
// im.db.Close().
func setupIntegrationManager(logger *logging.Logger) (*integrationManager, error) {
	provider := config.NewProvider(cfgFile)
	cfg, err := provider.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewDatabase(cfg.General.DatabasePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	if err := db.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	mcpClient := integration.NewStdioMCPClient(binaryPath, []string{"mcp"})
	store := integration.NewDatabaseStore(db.DB(), logger.Zap())
	manager := integration.NewManager(store, mcpClient, logger.Zap(), version.Version())

	claudeIntegration, err := claude.New(store, mcpClient, logger.Zap())
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create Claude integration: %w", err)
	}
	manager.RegisterIntegration(claudeIntegration)

	opencodeIntegration, err := opencode.New(store, mcpClient, logger.Zap())
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create OpenCode integration: %w", err)
	}
	manager.RegisterIntegration(opencodeIntegration)

	return &integrationManager{manager: manager, db: db}, nil
}
