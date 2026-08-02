package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func (m *Manager) Validate(ctx context.Context, name string) (*ValidationResult, error) {
	m.logger.Info("Validating integration", zap.String("name", name))

	_, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		if err == ErrNotFound {
			return &ValidationResult{
				Passed: false,
				Checks: []ValidationCheck{
					{
						Name:    "integration_installed",
						Passed:  false,
						Message: fmt.Sprintf("Integration '%s' is not installed. Run `portfolio install %s` first.", name, name),
					},
				},
			}, nil
		}
		return nil, NewError(ErrCodeStoreUnavailable, "Failed to load integration metadata", err)
	}

	integration, err := m.getIntegration(name)
	if err != nil {
		return nil, err
	}

	result, err := integration.Validate(ctx)
	if err != nil {
		m.logger.Error("Integration validation failed", zap.String("name", name), zap.Error(err))
		return nil, NewError(ErrCodeNotInstalled, fmt.Sprintf("Validation failed for integration '%s'", name), err)
	}

	m.logger.Info("Integration validation completed",
		zap.String("name", name),
		zap.Bool("passed", result.Passed))

	return result, nil
}

func (m *Manager) Doctor(ctx context.Context, name string, fix bool) (*ValidationResult, error) {
	m.logger.Info("Running doctor on integration", zap.String("name", name), zap.Bool("fix", fix))

	if name == "" {
		metas, err := m.store.ListIntegrations(ctx)
		if err != nil {
			return nil, NewError(ErrCodeStoreUnavailable, "Failed to list integrations", err)
		}

		if len(metas) == 0 {
			return &ValidationResult{
				Passed: true,
				Checks: []ValidationCheck{
					{
						Name:    "no_integrations",
						Passed:  true,
						Message: "No integrations installed",
					},
				},
			}, nil
		}

		allPassed := true
		allChecks := []ValidationCheck{}
		for _, meta := range metas {
			result, err := m.Doctor(ctx, meta.Name, fix)
			if err != nil {
				m.logger.Error("Doctor check failed", zap.String("name", meta.Name), zap.Error(err))
				allPassed = false
				continue
			}
			if !result.Passed {
				allPassed = false
			}
			allChecks = append(allChecks, result.Checks...)
		}

		return &ValidationResult{
			Passed: allPassed,
			Checks: allChecks,
		}, nil
	}

	return m.validateAndFix(ctx, name, fix)
}

func (m *Manager) validateAndFix(ctx context.Context, name string, fix bool) (*ValidationResult, error) {
	result, err := m.Validate(ctx, name)
	if err != nil {
		return nil, err
	}

	if !fix {
		return result, nil
	}

	fixedResult := &ValidationResult{
		Passed: true,
		Checks: []ValidationCheck{},
	}

	for _, check := range result.Checks {
		if check.Passed {
			fixedResult.Checks = append(fixedResult.Checks, check)
			continue
		}

		if check.SelfHealable {
			m.logger.Info("Attempting self-heal", zap.String("check", check.Name))
			fixed, healCheck := m.applySelfHeal(ctx, name, check)
			fixedResult.Checks = append(fixedResult.Checks, healCheck)
			if !fixed {
				fixedResult.Passed = false
			}
		} else {
			fixedResult.Checks = append(fixedResult.Checks, check)
			fixedResult.Passed = false
		}
	}

	return fixedResult, nil
}

func (m *Manager) applySelfHeal(ctx context.Context, name string, check ValidationCheck) (bool, ValidationCheck) {
	switch check.Name {
	case "mcp_server_reachable":
		return m.tryRestartMCP(ctx)
	case "config_file_exists":
		return m.tryRecreateConfig(ctx, name)
	case "integration_directory_exists":
		return m.tryRecreateDirectory(ctx, name)
	default:
		return false, check
	}
}

func (m *Manager) tryRestartMCP(ctx context.Context) (bool, ValidationCheck) {
	m.logger.Info("Attempting to restart MCP server")

	err := m.mcp.Health(ctx)
	if err == nil {
		return true, ValidationCheck{
			Name:    "mcp_server_reachable",
			Passed:  true,
			Message: "MCP server is now responding",
		}
	}

	return false, ValidationCheck{
		Name:         "mcp_server_reachable",
		Passed:       false,
		Message:      "Failed to restart MCP server: " + err.Error(),
		Remediation:  "Start manually: `portfolio mcp start`",
		SelfHealable: true,
	}
}

func (m *Manager) tryRecreateConfig(ctx context.Context, name string) (bool, ValidationCheck) {
	m.logger.Info("Attempting to recreate config file", zap.String("name", name))

	meta, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		return false, ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     "Failed to load integration metadata: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	configPath := filepath.Join(meta.InstallPath, "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return false, ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     "Failed to create config directory: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	data, err := MetaToJSON(*meta)
	if err != nil {
		return false, ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     "Failed to serialize config: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return false, ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     "Failed to write config file: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	return true, ValidationCheck{
		Name:    "config_file_exists",
		Passed:  true,
		Message: "Config file recreated",
	}
}

func (m *Manager) tryRecreateDirectory(ctx context.Context, name string) (bool, ValidationCheck) {
	m.logger.Info("Attempting to recreate integration directory", zap.String("name", name))

	meta, err := m.store.GetIntegration(ctx, name)
	if err != nil {
		return false, ValidationCheck{
			Name:        "integration_directory_exists",
			Passed:      false,
			Message:     "Failed to load integration metadata: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	if err := os.MkdirAll(meta.InstallPath, 0755); err != nil {
		return false, ValidationCheck{
			Name:        "integration_directory_exists",
			Passed:      false,
			Message:     "Failed to create directory: " + err.Error(),
			Remediation: "Reinstall integration: `portfolio install " + name + " --force`",
		}
	}

	return true, ValidationCheck{
		Name:    "integration_directory_exists",
		Passed:  true,
		Message: "Integration directory recreated",
	}
}

func CheckMCPTools(ctx context.Context, mcp MCPClient, requiredTools []string) ValidationCheck {
	availableTools, err := mcp.ListTools(ctx)
	if err != nil {
		return ValidationCheck{
			Name:         "mcp_tools_available",
			Passed:       false,
			Message:      "Failed to list MCP tools: " + err.Error(),
			Remediation:  "Start MCP server: `portfolio mcp start`",
			SelfHealable: true,
		}
	}

	missingTools := []string{}
	for _, required := range requiredTools {
		found := false
		for _, available := range availableTools {
			if strings.EqualFold(required, available) {
				found = true
				break
			}
		}
		if !found {
			missingTools = append(missingTools, required)
		}
	}

	if len(missingTools) > 0 {
		return ValidationCheck{
			Name:        "mcp_tools_available",
			Passed:      false,
			Message:     fmt.Sprintf("Missing MCP tools: %s", strings.Join(missingTools, ", ")),
			Remediation: "Reinstall integration: `portfolio install <name> --force`",
		}
	}

	return ValidationCheck{
		Name:    "mcp_tools_available",
		Passed:  true,
		Message: fmt.Sprintf("All %d required MCP tools available", len(requiredTools)),
	}
}

func CheckAgentBinary(ctx context.Context, binaryPath string) ValidationCheck {
	if binaryPath == "" {
		return ValidationCheck{
			Name:        "agent_binary_exists",
			Passed:      false,
			Message:     "Agent binary path not configured",
			Remediation: "Configure agent binary path in integration settings",
		}
	}

	_, err := exec.LookPath(binaryPath)
	if err != nil {
		return ValidationCheck{
			Name:        "agent_binary_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Agent binary not found at %s", binaryPath),
			Remediation: fmt.Sprintf("Install %s or update binary path in integration settings", binaryPath),
		}
	}

	return ValidationCheck{
		Name:    "agent_binary_exists",
		Passed:  true,
		Message: fmt.Sprintf("Agent binary found at %s", binaryPath),
	}
}

func CheckDirectoryExists(path string) ValidationCheck {
	if path == "" {
		return ValidationCheck{
			Name:         "integration_directory_exists",
			Passed:       false,
			Message:      "Integration path not configured",
			Remediation:  "Configure install path: `portfolio install <name> --path <path>`",
			SelfHealable: true,
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ValidationCheck{
				Name:         "integration_directory_exists",
				Passed:       false,
				Message:      fmt.Sprintf("Integration directory does not exist: %s", path),
				Remediation:  fmt.Sprintf("Create directory or reinstall: `portfolio install <name> --force`"),
				SelfHealable: true,
			}
		}
		return ValidationCheck{
			Name:        "integration_directory_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Failed to check directory: %s", err.Error()),
			Remediation: fmt.Sprintf("Check permissions and path: %s", path),
		}
	}

	if !info.IsDir() {
		return ValidationCheck{
			Name:        "integration_directory_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Path exists but is not a directory: %s", path),
			Remediation: fmt.Sprintf("Remove the file and reinstall: `portfolio install <name> --force`"),
		}
	}

	return ValidationCheck{
		Name:    "integration_directory_exists",
		Passed:  true,
		Message: fmt.Sprintf("Integration directory exists at %s", path),
	}
}

func CheckConfigFileExists(configPath string) ValidationCheck {
	if configPath == "" {
		return ValidationCheck{
			Name:         "config_file_exists",
			Passed:       false,
			Message:      "Config file path not configured",
			Remediation:  "Reinstall integration: `portfolio install <name> --force`",
			SelfHealable: true,
		}
	}

	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ValidationCheck{
				Name:         "config_file_exists",
				Passed:       false,
				Message:      fmt.Sprintf("Config file does not exist: %s", configPath),
				Remediation:  fmt.Sprintf("Recreate or reinstall: `portfolio install <name> --force`"),
				SelfHealable: true,
			}
		}
		return ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Failed to check config file: %s", err.Error()),
			Remediation: fmt.Sprintf("Check permissions and path: %s", configPath),
		}
	}

	if info.IsDir() {
		return ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Config path exists but is a directory: %s", configPath),
			Remediation: fmt.Sprintf("Remove the directory and reinstall: `portfolio install <name> --force`"),
		}
	}

	return ValidationCheck{
		Name:    "config_file_exists",
		Passed:  true,
		Message: fmt.Sprintf("Config file exists at %s", configPath),
	}
}
