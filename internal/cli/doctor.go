package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"project-dash/internal/config"
	"project-dash/internal/database"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [target]",
	Short: "Run diagnostics and health checks",
	Long: `Perform diagnostics to identify and resolve issues.

Without arguments, runs system diagnostics:
- Configuration file accessibility and validity
- Database file accessibility and integrity
- Project roots accessibility
- File permissions
- Disk space availability
- Go environment and dependencies

With a target, runs integration-specific diagnostics:
  claude     Claude Code integration health
  opencode   OpenCode integration health`,
	Args: cobra.MaximumNArgs(1),
	Run:  runDoctor,
}

var doctorClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Check Claude Code integration health",
	Long: `Check the health of the Claude Code integration.

This validates the integration status, MCP server connectivity,
configuration files, and skills installation.`,
	Example: `  portfolio doctor claude`,
	Args:    cobra.NoArgs,
	Run:     runDoctorClaude,
}

var doctorOpencodeCmd = &cobra.Command{
	Use:   "opencode",
	Short: "Check OpenCode integration health",
	Long: `Check the health of the OpenCode integration.

This validates the integration status, MCP server connectivity,
the opencode.json config entry, and the installed skill.`,
	Example: `  portfolio doctor opencode`,
	Args:    cobra.NoArgs,
	Run:     runDoctorOpencode,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.AddCommand(doctorClaudeCmd)
	doctorCmd.AddCommand(doctorOpencodeCmd)
}

func runDoctor(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "claude":
			runDoctorClaude(cmd, []string{})
			return
		case "opencode":
			runDoctorOpencode(cmd, []string{})
			return
		default:
			fmt.Printf("Error: Unknown doctor target '%s'\n", args[0])
			fmt.Println("Supported targets: claude, opencode")
			os.Exit(1)
		}
	}

	logger := logging.GetGlobalLogger()

	fmt.Println("Portfolio Engine Diagnostics")
	fmt.Println("=============================")
	fmt.Println()

	allPassed := true

	if !checkConfigFile(logger) {
		allPassed = false
	}

	if !checkDatabase(logger) {
		allPassed = false
	}

	if !checkProjectRoots(logger) {
		allPassed = false
	}

	if !checkFilePermissions(logger) {
		allPassed = false
	}

	if !checkDiskSpace(logger) {
		allPassed = false
	}

	if !checkGoEnvironment(logger) {
		allPassed = false
	}

	fmt.Println()

	if allPassed {
		fmt.Println("✓ All checks passed")
		os.Exit(0)
	} else {
		fmt.Println("✗ Some checks failed - run 'portfolio doctor' again after fixing issues")
		os.Exit(1)
	}
}

func checkConfigFile(logger *logging.Logger) bool {
	fmt.Println("Configuration Check:")

	configPath := models.GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Config file not found: %s\n", configPath)
		fmt.Printf("    Action: Run 'portfolio init' to create configuration\n")
		return false
	}

	if _, err := os.ReadFile(configPath); err != nil {
		fmt.Printf("  ✗ Config file not readable: %s\n", configPath)
		fmt.Printf("    Action: Check file permissions\n")
		return false
	}

	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Printf("  ✗ Config file invalid: %s\n", configPath)
		fmt.Printf("    Error: %v\n", err)
		return false
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Printf("  ✗ Config validation failed: %s\n", configPath)
		fmt.Printf("    Error: %v\n", err)
		return false
	}

	fmt.Printf("  ✓ Config file accessible: %s\n", configPath)
	fmt.Printf("  ✓ Config file valid: TOML parses correctly\n")
	fmt.Printf("  ✓ Project roots configured: %d\n", len(cfg.Discovery.ProjectRoots))

	return true
}

func checkDatabase(logger *logging.Logger) bool {
	fmt.Println("\nDatabase Check:")

	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		fmt.Printf("  ✗ Cannot load configuration to get database path\n")
		return false
	}

	dbPath := cfg.General.DatabasePath

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Database file not found: %s\n", dbPath)
		fmt.Printf("    Action: Run 'portfolio init' to create database\n")
		return false
	}

	db, err := database.NewDatabase(dbPath, logger)
	if err != nil {
		fmt.Printf("  ✗ Database not accessible: %s\n", dbPath)
		fmt.Printf("    Error: %v\n", err)
		return false
	}

	if err := db.Connect(); err != nil {
		fmt.Printf("  ✗ Database not accessible: %s\n", dbPath)
		fmt.Printf("    Error: %v\n", err)
		return false
	}
	defer db.Close()

	version, err := db.GetSchemaVersion()
	if err != nil {
		fmt.Printf("  ✗ Cannot determine schema version\n")
		return false
	}

	tableCount, err := db.GetTableCount()
	if err != nil {
		fmt.Printf("  ✗ Cannot validate schema\n")
		return false
	}

	fmt.Printf("  ✓ Database accessible: %s\n", dbPath)
	fmt.Printf("  ✓ Schema version: %d\n", version)
	fmt.Printf("  ✓ Tables present: %d/10\n", tableCount)

	return true
}

func checkProjectRoots(logger *logging.Logger) bool {
	fmt.Println("\nProject Roots Check:")

	loader := config.NewLoader("")
	cfg, err := loader.Load()
	if err != nil {
		fmt.Printf("  ✗ Cannot load configuration to check project roots\n")
		return false
	}

	if len(cfg.Discovery.ProjectRoots) == 0 {
		fmt.Printf("  ⚠ No project roots configured\n")
		fmt.Printf("    Action: Run 'portfolio init' to configure project roots\n")
		return false
	}

	allAccessible := true
	for i, root := range cfg.Discovery.ProjectRoots {
		if err := validateProjectRoot(root); err != nil {
			fmt.Printf("  ✗ %s: %s\n", root, err)
			allAccessible = false
		} else {
			fmt.Printf("  ✓ %s: Accessible\n", root)
		}

		if i > 5 {
			fmt.Printf("  ... (%d more roots)\n", len(cfg.Discovery.ProjectRoots)-i-1)
			break
		}
	}

	return allAccessible
}

func checkFilePermissions(logger *logging.Logger) bool {
	fmt.Println("\nFile Permissions Check:")

	configPath := models.GetConfigPath()

	info, err := os.Stat(configPath)
	if err != nil {
		fmt.Printf("  ✗ Cannot check config file permissions\n")
		return false
	}

	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Printf("  ⚠ Config file has loose permissions: %o\n", mode)
		fmt.Printf("    Action: chmod 600 %s\n", configPath)
		return false
	}

	fmt.Printf("  ✓ Config file permissions secure: %o\n", mode)

	return true
}

func checkDiskSpace(logger *logging.Logger) bool {
	fmt.Println("\nDisk Space Check:")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("  ✗ Cannot determine home directory\n")
		return false
	}

	var path string
	if runtime.GOOS == "windows" {
		path = filepath.VolumeName(homeDir) + "\\"
	} else {
		path = homeDir
	}

	if runtime.GOOS != "windows" {
		cmd := exec.Command("df", "-h", path)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("  ⚠ Cannot check disk space\n")
			return true
		}

		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 3 {
				available := fields[3]
				fmt.Printf("  ✓ Disk space available: %s\n", available)
				return true
			}
		}
	}

	fmt.Printf("  ✓ Disk space check skipped (platform-specific)\n")
	return true
}

func checkGoEnvironment(logger *logging.Logger) bool {
	fmt.Println("\nSystem Check:")

	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  ✗ Go not found in PATH\n")
		return false
	}

	goVersion := strings.TrimSpace(string(output))
	fmt.Printf("  ✓ Go version: %s\n", goVersion)

	cmd = exec.Command("go", "list", "-m", "all")
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("  ✗ Cannot list Go dependencies\n")
		return false
	}

	deps := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Printf("  ✓ Dependencies: %d present\n", len(deps))

	return true
}

func validateProjectRoot(root string) error {
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	} else if err != nil {
		return fmt.Errorf("cannot access path")
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	if _, err := os.ReadDir(root); err != nil {
		return fmt.Errorf("not readable")
	}

	return nil
}

func runDoctorClaude(cmd *cobra.Command, args []string)   { runDoctorIntegration(cmd, "claude") }
func runDoctorOpencode(cmd *cobra.Command, args []string) { runDoctorIntegration(cmd, "opencode") }

func runDoctorIntegration(cmd *cobra.Command, name string) {
	logger := logging.GetGlobalLogger()
	ctx := cmd.Context()

	im, err := setupIntegrationManager(logger)
	if err != nil {
		logger.Error(err.Error(), models.Field{Key: "error", Value: err})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer im.db.Close()

	result, err := im.manager.Doctor(ctx, name, false)
	if err != nil {
		logger.Error("doctor check failed", models.Field{Key: "error", Value: err})
		fmt.Printf("Error: doctor check failed: %v\n", err)
		os.Exit(1)
	}

	display := integrationDisplayName(name)
	if result.Passed {
		fmt.Printf("✓ All %s integration checks passed\n", display)
		return
	}

	fmt.Printf("✗ %s integration health check failed\n", display)
	fmt.Println()

	for _, check := range result.Checks {
		if check.Passed {
			fmt.Printf("  ✓ %s: %s\n", check.Name, check.Message)
		} else {
			fmt.Printf("  ✗ %s: %s\n", check.Name, check.Message)
			if check.Remediation != "" {
				fmt.Printf("    → %s\n", check.Remediation)
			}
		}
	}

	os.Exit(1)
}
