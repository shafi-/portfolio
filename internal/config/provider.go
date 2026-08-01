package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"project-dash/pkg/models"
)

// Provider is the single source of truth for configuration paths and access.
// It centralizes path resolution and config lifecycle, preventing inconsistencies
// where different commands resolve to different config files.
type Provider struct {
	configPath string
	mu         sync.RWMutex
	config     *models.Config
}

// NewProvider creates a config provider. If configPath is empty, the canonical
// default path is used (~/Library/Application Support/com.portfolio.cli/config.toml
// on macOS). An explicit path is useful for the --config flag or tests.
func NewProvider(configPath string) *Provider {
	if configPath == "" {
		configPath = models.GetConfigPath()
	}
	return &Provider{configPath: configPath}
}

// ConfigPath returns the resolved config file path.
func (p *Provider) ConfigPath() string {
	return p.configPath
}

// DataDir returns the portfolio data directory (same dir as config on macOS).
func (p *Provider) DataDir() string {
	return filepath.Dir(p.configPath)
}

// DatabasePath returns the default database path derived from the config directory.
func (p *Provider) DatabasePath() string {
	return models.GetDefaultDatabasePath()
}

// LogPath returns the default log file path derived from the config directory.
func (p *Provider) LogPath() string {
	return models.GetDefaultLogPath()
}

// IntegrationsDir returns the default directory for integration files.
func (p *Provider) IntegrationsDir() string {
	return models.GetDefaultIntegrationsDir()
}

// IntegrationPath returns the path for a specific integration's files.
func (p *Provider) IntegrationPath(name string) string {
	return filepath.Join(p.IntegrationsDir(), name)
}

// Load reads and parses the configuration file. Returns a cached copy on
// subsequent calls within the same Provider instance. If the file does not
// exist, a default config is created and persisted.
func (p *Provider) Load() (*models.Config, error) {
	p.mu.RLock()
	if p.config != nil {
		defer p.mu.RUnlock()
		return p.config, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if p.config != nil {
		return p.config, nil
	}

	cfg, err := p.loadFromDisk()
	if err != nil {
		return nil, err
	}

	p.config = cfg
	return cfg, nil
}

// Save writes the given config to disk and updates the in-memory cache.
func (p *Provider) Save(cfg *models.Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.saveToDisk(cfg); err != nil {
		return err
	}

	p.config = cfg
	return nil
}

// Update modifies the config via a callback and persists the result.
// The callback receives a deep copy; only the returned (possibly modified)
// config is saved.
func (p *Provider) Update(fn func(*models.Config) error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cfg, err := p.loadFromDisk()
	if err != nil {
		return err
	}

	// Deep copy so the callback cannot corrupt cached state.
	clone := deepCopyConfig(cfg)

	if err := fn(&clone); err != nil {
		return err
	}

	if err := p.saveToDisk(&clone); err != nil {
		return err
	}

	p.config = &clone
	return nil
}

// Invalidate forces the next Load() to re-read from disk.
func (p *Provider) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = nil
}

// loadFromDisk reads and parses the config file. Creates defaults if missing.
func (p *Provider) loadFromDisk() (*models.Config, error) {
	if _, err := os.Stat(p.configPath); os.IsNotExist(err) {
		defaultConfig := models.GetDefaultConfig()
		if err := p.saveToDisk(defaultConfig); err != nil {
			return nil, &ConfigError{
				Message: fmt.Sprintf("Cannot create default config: %s", p.configPath),
				Cause:   err,
			}
		}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(p.configPath)
	if err != nil {
		return nil, &ConfigError{
			Message: fmt.Sprintf("Cannot read config file: %s", p.configPath),
			Cause:   err,
		}
	}

	config := models.GetDefaultConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, &ConfigError{
			Message: fmt.Sprintf("Invalid TOML syntax in config file: %s", p.configPath),
			Cause:   err,
		}
	}

	return config, nil
}

// saveToDisk writes the config to disk, creating parent directories as needed.
func (p *Provider) saveToDisk(cfg *models.Config) error {
	dir := filepath.Dir(p.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(p.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func deepCopyConfig(src *models.Config) models.Config {
	dst := *src
	dst.Discovery.ProjectRoots = make([]string, len(src.Discovery.ProjectRoots))
	copy(dst.Discovery.ProjectRoots, src.Discovery.ProjectRoots)
	dst.Discovery.IgnoredPaths = make([]string, len(src.Discovery.IgnoredPaths))
	copy(dst.Discovery.IgnoredPaths, src.Discovery.IgnoredPaths)
	dst.Dashboard.AllowedOrigins = make([]string, len(src.Dashboard.AllowedOrigins))
	copy(dst.Dashboard.AllowedOrigins, src.Dashboard.AllowedOrigins)
	return dst
}
