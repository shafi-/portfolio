package opencode

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"project-dash/pkg/models"
)

func detectPaths() (OpenCodeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return OpenCodeConfig{}, err
	}

	configPath := detectConfigPath(homeDir)
	skillsDir := detectSkillsDir(homeDir)

	binaryPath, err := detectBinaryPath()
	if err != nil {
		return OpenCodeConfig{}, err
	}

	return OpenCodeConfig{
		InstallPath: detectInstallPath(),
		ConfigPath:  configPath,
		SkillsDir:   skillsDir,
		BinaryPath:  binaryPath,
	}, nil
}

// configRoot returns the opencode config directory, honouring XDG_CONFIG_HOME.
func configRoot(homeDir string) string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "opencode")
	}
	return filepath.Join(homeDir, ".config", "opencode")
}

// detectConfigPath returns the path to opencode.json. The file may not exist
// yet; it is created on first install.
func detectConfigPath(homeDir string) string {
	return filepath.Join(configRoot(homeDir), "opencode.json")
}

// detectSkillsDir returns the directory opencode loads skills from:
// <configRoot>/skills/<name>/SKILL.md.
func detectSkillsDir(homeDir string) string {
	return filepath.Join(configRoot(homeDir), "skills")
}

func detectBinaryPath() (string, error) {
	// The binary is already running — resolve its own path instead of
	// searching PATH, which fails for ./portfolio or `go run`.
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs, nil
		}
		return exe, nil
	}
	// Fallback: a Portfolio install on PATH (e.g. `go install`).
	if path, err := exec.LookPath("portfolio"); err == nil {
		return path, nil
	}
	return "", errors.New("could not locate the portfolio binary")
}

func detectInstallPath() string {
	return filepath.Join(models.GetDefaultIntegrationsDir(), "opencode")
}

func isOpenCodeInstalled() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}
