package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

func (m *Manager) Upgrade(ctx context.Context, name string, opts UpgradeOptions) (*IntegrationMeta, error) {
	m.logger.Info("Upgrading integration", zap.String("name", name), zap.String("targetVersion", opts.TargetVersion))

	meta, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		if err == ErrNotFound {
			return nil, NewError(ErrCodeNotInstalled, fmt.Sprintf("Integration '%s' is not installed. Cannot upgrade.", name), nil)
		}
		return nil, NewError(ErrCodeStoreUnavailable, "Failed to load integration metadata", err)
	}

	integration, err := m.getIntegration(name)
	if err != nil {
		return nil, err
	}

	if opts.TargetVersion == "" {
		return nil, NewError(ErrCodeIncompatible, "Target version must be specified for upgrade", nil)
	}

	currentVersion, err := semver.NewVersion(meta.Version)
	if err != nil {
		return nil, NewError(ErrCodeIncompatible, fmt.Sprintf("Invalid current version '%s': %v", meta.Version, err), nil)
	}

	targetVersion, err := semver.NewVersion(opts.TargetVersion)
	if err != nil {
		return nil, NewError(ErrCodeIncompatible, fmt.Sprintf("Invalid target version '%s': %v", opts.TargetVersion, err), nil)
	}

	if targetVersion.LessThan(currentVersion) {
		return nil, NewError(ErrCodeIncompatible, fmt.Sprintf("Downgrade not supported: current %s, target %s", meta.Version, opts.TargetVersion), nil)
	}

	if targetVersion.Equal(currentVersion) {
		m.logger.Info("Integration already at target version",
			zap.String("name", name),
			zap.String("version", meta.Version))
		return meta, nil
	}

	backupPath := filepath.Join(meta.InstallPath, "backup")
	if err := m.createBackup(ctx, meta.InstallPath, backupPath); err != nil {
		return nil, NewError(ErrCodeUpgradeFailed, "Failed to create backup before upgrade", err)
	}

	defer func() {
		if err != nil {
			m.logger.Warn("Upgrade failed, attempting rollback", zap.String("name", name))
			if rollbackErr := m.restoreBackup(ctx, meta.InstallPath, backupPath, meta.Version); rollbackErr != nil {
				m.logger.Error("Rollback failed",
					zap.String("name", name),
					zap.Error(rollbackErr))
			}
		}
	}()

	upgradeResult, err := integration.Upgrade(ctx, opts)
	if err != nil {
		m.logger.Error("Integration upgrade failed",
			zap.String("name", name),
			zap.Error(err))
		return nil, NewError(ErrCodeUpgradeFailed, fmt.Sprintf("Upgrade failed for integration '%s'", name), err)
	}

	m.logger.Info("Validating upgraded integration", zap.String("name", name))

	validationResult, err := integration.Validate(ctx)
	if err != nil {
		m.logger.Error("Post-upgrade validation failed",
			zap.String("name", name),
			zap.Error(err))
		return nil, NewError(ErrCodeUpgradeFailed, "Post-upgrade validation failed", err)
	}

	if !validationResult.Passed {
		m.logger.Error("Post-upgrade validation checks failed",
			zap.String("name", name),
			zap.Int("failed_checks", len(validationResult.Checks)))
		return nil, NewError(ErrCodeUpgradeFailed, fmt.Sprintf("Post-upgrade validation failed: %d checks failed", len(validationResult.Checks)), nil)
	}

	updatedMeta := *meta
	updatedMeta.Version = opts.TargetVersion
	updatedMeta.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := m.store.SaveIntegration(ctx, updatedMeta); err != nil {
		m.logger.Error("Failed to save upgraded integration metadata",
			zap.String("name", name),
			zap.Error(err))
		return nil, NewError(ErrCodeStoreUnavailable, "Failed to save integration metadata", err)
	}

	m.cleanupBackup(backupPath)

	m.logger.Info("Integration upgraded successfully",
		zap.String("name", name),
		zap.String("old_version", upgradeResult.PreviousVersion),
		zap.String("new_version", upgradeResult.NewVersion))

	return &updatedMeta, nil
}

func (m *Manager) Remove(ctx context.Context, name string) error {
	m.logger.Info("Removing integration", zap.String("name", name))

	meta, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		if err == ErrNotFound {
			m.logger.Info("Integration not installed, nothing to remove", zap.String("name", name))
			return nil
		}
		return NewError(ErrCodeStoreUnavailable, "Failed to load integration metadata", err)
	}

	integration, err := m.getIntegration(name)
	if err != nil {
		return err
	}

	if err := integration.Remove(ctx); err != nil {
		m.logger.Error("Integration removal failed",
			zap.String("name", name),
			zap.Error(err))
		return NewError(ErrCodeUpgradeFailed, fmt.Sprintf("Removal failed for integration '%s'", name), err)
	}

	if meta.InstallPath != "" {
		m.logger.Info("Cleaning up integration directory", zap.String("path", meta.InstallPath))
		if err := os.RemoveAll(meta.InstallPath); err != nil {
			m.logger.Warn("Failed to remove integration directory",
				zap.String("path", meta.InstallPath),
				zap.Error(err))
		}
	}

	if err := m.store.DeleteIntegration(ctx, name); err != nil {
		return NewError(ErrCodeStoreUnavailable, "Failed to delete integration metadata", err)
	}

	m.logger.Info("Integration removed successfully", zap.String("name", name))

	return nil
}

func (m *Manager) createBackup(ctx context.Context, sourcePath, backupPath string) error {
	m.logger.Info("Creating backup", zap.String("source", sourcePath), zap.String("backup", backupPath))

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(backupPath, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return err
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, data, info.Mode()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	m.logger.Info("Backup created successfully", zap.String("backup", backupPath))

	return nil
}

func (m *Manager) restoreBackup(ctx context.Context, targetPath, backupPath, version string) error {
	m.logger.Info("Restoring from backup", zap.String("target", targetPath), zap.String("backup", backupPath))

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup does not exist: %s", backupPath)
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to clean target directory: %w", err)
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == backupPath {
			return nil
		}

		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetPath, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return err
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, data, info.Mode()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	m.logger.Info("Backup restored successfully", zap.String("target", targetPath))

	return nil
}

func (m *Manager) cleanupBackup(backupPath string) {
	m.logger.Info("Cleaning up backup", zap.String("backup", backupPath))

	if err := os.RemoveAll(backupPath); err != nil {
		m.logger.Warn("Failed to remove backup directory",
			zap.String("backup", backupPath),
			zap.Error(err))
	}
}
