package claude

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func detectPaths() (ClaudeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ClaudeConfig{}, err
	}

	configPath, err := detectConfigPath(homeDir)
	if err != nil {
		return ClaudeConfig{}, err
	}

	skillsDir := detectSkillsDir(homeDir)

	binaryPath, err := detectBinaryPath()
	if err != nil {
		return ClaudeConfig{}, err
	}

	return ClaudeConfig{
		InstallPath: detectInstallPath(),
		ConfigPath:  configPath,
		SkillsDir:   skillsDir,
		BinaryPath:  binaryPath,
	}, nil
}

func detectConfigPath(homeDir string) (string, error) {
	candidates := []string{
		filepath.Join(homeDir, ".claude", "settings.json"),
		filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append([]string{filepath.Join(xdg, "claude", "settings.json")}, candidates...)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	defaultPath := filepath.Join(homeDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(defaultPath), 0755); err != nil {
		return "", err
	}

	return defaultPath, nil
}

func detectSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
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
	return filepath.Join(".portfolio", "integrations", "claude")
}

func isClaudeInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (c *ClaudeConfig) isDarwin() bool {
	return runtime.GOOS == "darwin"
}

func (c *ClaudeConfig) isLinux() bool {
	return runtime.GOOS == "linux"
}

func (c *ClaudeConfig) isWindows() bool {
	return runtime.GOOS == "windows"
}
