package models

import (
	"fmt"
)

// Dashboard config methods for the dashboard API handlers

// GetDashboardPort returns the dashboard port
func (c *Config) GetDashboardPort() int {
	return c.Dashboard.Port
}

// GetDashboardHost returns the dashboard host
func (c *Config) GetDashboardHost() string {
	return c.Dashboard.Host
}

// GetDashboardAssetsPath returns the dashboard assets path
func (c *Config) GetDashboardAssetsPath() string {
	return c.Dashboard.AssetPath
}

// GetAllowedOrigins returns the allowed CORS origins
func (c *Config) GetAllowedOrigins() []string {
	return c.Dashboard.AllowedOrigins
}

// SetDashboardPort sets the dashboard port
func (c *Config) SetDashboardPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	c.Dashboard.Port = port
	return nil
}

// SetDashboardHost sets the dashboard host
func (c *Config) SetDashboardHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	c.Dashboard.Host = host
	return nil
}

// SetDashboardAssetsPath sets the dashboard assets path
func (c *Config) SetDashboardAssetsPath(path string) error {
	c.Dashboard.AssetPath = path
	return nil
}

// SetAllowedOrigins sets the allowed CORS origins
func (c *Config) SetAllowedOrigins(origins []string) error {
	c.Dashboard.AllowedOrigins = origins
	return nil
}

// ToMap converts the config to a map for API responses
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"general": map[string]interface{}{
			"database_path": c.General.DatabasePath,
		},
		"discovery": map[string]interface{}{
			"project_roots": c.Discovery.ProjectRoots,
			"ignored_paths": c.Discovery.IgnoredPaths,
		},
		"logging": map[string]interface{}{
			"level": c.Logging.Level,
		},
		"dashboard": map[string]interface{}{
			"host":            c.Dashboard.Host,
			"port":            c.Dashboard.Port,
			"asset_path":      c.Dashboard.AssetPath,
			"allowed_origins": c.Dashboard.AllowedOrigins,
		},
	}
}

// FromMap updates the config from a map (for PATCH requests)
func (c *Config) FromMap(data map[string]interface{}) error {
	for key, value := range data {
		switch key {
		case "dashboard.port":
			if port, ok := value.(float64); ok {
				if err := c.SetDashboardPort(int(port)); err != nil {
					return err
				}
			}
		case "dashboard.host":
			if host, ok := value.(string); ok {
				if err := c.SetDashboardHost(host); err != nil {
					return err
				}
			}
		case "dashboard.assets_path":
			if path, ok := value.(string); ok {
				if err := c.SetDashboardAssetsPath(path); err != nil {
					return err
				}
			}
		case "dashboard.allowed_origins":
			if origins, ok := value.([]interface{}); ok {
				originsStr := make([]string, len(origins))
				for i, origin := range origins {
					if originStr, ok := origin.(string); ok {
						originsStr[i] = originStr
					}
				}
				if err := c.SetAllowedOrigins(originsStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
