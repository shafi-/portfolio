package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	defaultInstallPath = ".portfolio/integrations"
	integrationMetaKey = "integration:%s:meta"
)

type Manager struct {
	store         Store
	mcp           MCPClient
	logger        *zap.Logger
	integrations  map[string]Integration
	engineVersion string
}

func NewManager(store Store, mcpClient MCPClient, logger *zap.Logger, engineVersion string) *Manager {
	return &Manager{
		store:         store,
		mcp:           mcpClient,
		logger:        logger,
		integrations:  make(map[string]Integration),
		engineVersion: engineVersion,
	}
}

func (m *Manager) RegisterIntegration(integration Integration) {
	m.integrations[integration.Name()] = integration
}

func (m *Manager) getIntegration(name string) (Integration, error) {
	integration, ok := m.integrations[name]
	if !ok {
		return nil, NewError(ErrCodeNotFound, fmt.Sprintf("Integration '%s' not found", name), nil)
	}
	return integration, nil
}

func (m *Manager) Install(ctx context.Context, name string, opts InstallOptions) (*IntegrationMeta, error) {
	integration, err := m.getIntegration(name)
	if err != nil {
		return nil, err
	}

	existingMeta, err := m.store.GetIntegration(ctx, name)
	if err == nil && existingMeta != nil {
		if existingMeta.Status == StatusInstalled {
			if opts.Force {
				m.logger.Info("Reinstalling integration", zap.String("name", name))
			} else {
				return nil, NewError(ErrCodeAlreadyInstalled, fmt.Sprintf("Integration '%s' is already installed. Use `upgrade` or `uninstall` first.", name), nil)
			}
		}
	} else if err != ErrNotFound {
		return nil, NewError(ErrCodeStoreUnavailable, "Knowledge store unavailable", err)
	}

	installPath := opts.InstallPath
	if installPath == "" {
		installPath = filepath.Join(defaultInstallPath, name)
	}

	m.logger.Info("Installing integration", zap.String("name", name), zap.String("path", installPath))

	result, err := integration.Install(ctx, opts)
	if err != nil {
		return nil, NewError(ErrCodeNotInstalled, fmt.Sprintf("Installation failed for integration '%s'", name), err)
	}

	meta := result.Meta
	meta.InstallPath = installPath
	meta.Status = StatusInstalled
	meta.InstalledAt = time.Now().Format(time.RFC3339)
	meta.UpdatedAt = meta.InstalledAt

	if err := m.store.SaveIntegration(ctx, meta); err != nil {
		return nil, NewError(ErrCodeStoreUnavailable, "Failed to save integration metadata", err)
	}

	m.logger.Info("Integration installed successfully", zap.String("name", name), zap.String("version", meta.Version))

	return &meta, nil
}

func (m *Manager) List(ctx context.Context) ([]IntegrationMeta, error) {
	metas, err := m.store.ListIntegrations(ctx)
	if err != nil {
		return nil, NewError(ErrCodeStoreUnavailable, "Failed to list integrations", err)
	}
	return metas, nil
}

func (m *Manager) Get(ctx context.Context, name string) (*IntegrationMeta, error) {
	meta, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		return nil, NewError(ErrCodeNotFound, fmt.Sprintf("Integration '%s' not found", name), err)
	}
	return meta, nil
}

func GetMetaKey(name string) string {
	return fmt.Sprintf(integrationMetaKey, name)
}

func MetaToJSON(meta IntegrationMeta) ([]byte, error) {
	return json.Marshal(meta)
}

func MetaFromJSON(data []byte) (IntegrationMeta, error) {
	var meta IntegrationMeta
	err := json.Unmarshal(data, &meta)
	return meta, err
}
