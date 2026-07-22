package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nerddevsltd/portfolio/internal/config"
	"github.com/nerddevsltd/portfolio/internal/database"
	"github.com/nerddevsltd/portfolio/internal/logging"
	"github.com/nerddevsltd/portfolio/pkg/models"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system diagnostics and health checks",
	Long: `Perform comprehensive system diagnostics to identify and resolve issues.

Diagnostic checks include:
- Configuration file accessibility and validity
- Database file accessibility and integrity
- Project roots accessibility
- File permissions
- Disk space availability
- Go environment and dependencies`,
	Run: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) {
	logger := logging.GetGlobalLogger()

	fmt.Println("Portfolio Engine Diagnostics")
	fmt.Println("=============================")
	fmt.Println()

	allPassed := true

	// Check 1: Configuration
	if !checkConfigFile(logger) {
		allPassed = false
	}

	// Check 2: Database
	if !checkDatabase(logger) {
		allPassed = false
	}

	// Check 3: Project Roots
	if !checkProjectRoots(logger) {
		allPassed = false
	}

	// Check 4: File Permissions
	if !checkFilePermissions(logger) {
		allPassed = false
	}

	// Check 5: Disk Space
	if !checkDiskSpace(logger) {
		allPassed = false
	}

	// Check 6: Go Environment
	if !checkGoEnvironment(logger) {
		allPassed = false
	}

	fmt.Println()

	// Exit with appropriate code
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

	// Check file existence
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Config file not found: %s\n", configPath)
		fmt.Printf("    Action: Run 'portfolio init' to create configuration\n")
		return false
	}

	// Check file readability
	if _, err := os.ReadFile(configPath); err != nil {
		fmt.Printf("  ✗ Config file not readable: %s\n", configPath)
		fmt.Printf("    Action: Check file permissions\n")
		return false
	}

	// Check TOML validity
	manager := config.NewManager(configPath)
	cfg, err := manager.LoadConfig()
	if err != nil {
		fmt.Printf("  ✗ Config file invalid: %s\n", configPath)
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

	// Load configuration to get database path
	manager := config.NewManager("")
	cfg, err := manager.LoadConfig()
	if err != nil {
		fmt.Printf("  ✗ Cannot load configuration to get database path\n")
		return false
	}

	dbPath := cfg.General.DatabasePath

	// Check file existence
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Database file not found: %s\n", dbPath)
		fmt.Printf("    Action: Run 'portfolio init' to create database\n")
		return false
	}

	// Check file accessibility
	db, err := database.NewDatabase(dbPath, logger)
	if err != nil {
		fmt.Printf("  ✗ Database not accessible: %s\n", dbPath)
		fmt.Printf("    Error: %v\n", err)
		return false
	}
	defer db.Close()

	// Check schema version
	version, err := db.GetSchemaVersion()
	if err != nil {
		fmt.Printf("  ✗ Cannot determine schema version\n")
		return false
	}

	// Check table count
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

	manager := config.NewManager("")
	cfg, err := manager.LoadConfig()
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

	// Check config file permissions
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

	// Get disk usage (Unix-specific)
	if runtime.GOOS != "windows" {
		cmd := exec.Command("df", "-h", path)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("  ⚠ Cannot check disk space\n")
			return true // Non-critical
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

	// Check Go version
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  ✗ Go not found in PATH\n")
		return false
	}

	goVersion := strings.TrimSpace(string(output))
	fmt.Printf("  ✓ Go version: %s\n", goVersion)

	// Check dependencies
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
	// Clean the path
	root = filepath.Clean(root)

	// Check if path exists
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	} else if err != nil {
		return fmt.Errorf("cannot access path")
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	// Check if readable
	if _, err := os.ReadDir(root); err != nil {
		return fmt.Errorf("not readable")
	}

	return nil
}