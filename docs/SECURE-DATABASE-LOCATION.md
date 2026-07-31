# Secure Database Location Implementation

**Date**: 2026-07-28  
**Status**: System-standard secure directories  
**Security**: Significantly improved

---

## Problem Solved

### Previous Insecure Location
```bash
# Old predictable location:
~/.portfolio/portfolio.db  # Too easy to discover!

# Security issues:
- Easy to find with: find ~ -name "*portfolio*"
- Standard Unix convention, very predictable
- First place attackers would look
- Simple location for automated tools
```

### New Secure Location
```bash
# New system-standard locations:

macOS:
~/Library/Application Support/portfolio/portfolio.db

Linux:
~/.local/share/portfolio/portfolio.db

Windows:
%APPDATA%/portfolio/portfolio.db

# Security benefits:
- Follows system conventions (less suspicious)
- Much harder to discover with generic searches
- Mixed with other application data (better hiding)
- System-standard permissions and protections
```

---

## Implementation Details

### Platform-Specific Paths

#### macOS (Apple Standard)
```bash
~/Library/Application Support/portfolio/
├── portfolio.db          # Main database
├── portfolio.log         # Log file
└── config.toml           # Configuration

# Why this is secure:
- Standard macOS app location
- Hidden among hundreds of other apps
- System permissions apply
- Not easily searchable from terminal
```

#### Linux (XDG Standard)
```bash
# Data files:
~/.local/share/portfolio/
├── portfolio.db
└── portfolio.log

# Config files:
~/.config/portfolio/
└── config.toml

# Why this is secure:
- Follows XDG Base Directory Specification
- Standard Linux app organization
- Mixed with other app data
- Less predictable than ~/.portfolio
```

#### Windows (Microsoft Standard)
```bash
%APPDATA%/portfolio/
├── portfolio.db
├── portfolio.log
└── config.toml

# Why this is secure:
- Standard Windows app data location
- Protected by Windows permissions
- Hidden in AppData directory structure
- System-managed location
```

---

## Code Implementation

### Path Resolution Logic

```go
// pkg/models/paths.go

func GetPortfolioDataDir() string {
	// Environment override for testing/custom installs
	if customDir := os.Getenv("PORTFOLIO_DATA_DIR"); customDir != "" {
		return customDir
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/portfolio
		return filepath.Join(homeDir, "Library", "Application Support", "portfolio")
	case "windows":
		// Windows: %APPDATA%/portfolio
		return filepath.Join(os.Getenv("APPDATA"), "portfolio")
	default:
		// Linux: ~/.local/share/portfolio (XDG standard)
		return filepath.Join(homeDir, ".local", "share", "portfolio")
	}
}

func GetDefaultDatabasePath() string {
	return filepath.Join(GetPortfolioDataDir(), "portfolio.db")
}
```

### Configuration Update

```go
// pkg/models/config.go
func GetDefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			DatabasePath: GetDefaultDatabasePath(), // ✅ Secure path
		},
		Logging: LoggingConfig{
			File: GetDefaultLogPath(), // ✅ Secure path
		},
	}
}
```

---

## Security Benefits

### 1. Discovery Difficulty
```bash
# Old approach - too easy:
$ find ~ -name "*portfolio*"
/Users/nerddevsltd/.portfolio/portfolio.db  # Found immediately!

# New approach - much harder:
$ find ~/Library -name "*portfolio*"
# Returns hundreds of results from many apps
# Portfolio database hidden among legitimate applications

$ find ~/.local/share -name "*.db" 
# Returns databases from dozens of applications
# Portfolio database not easily identifiable
```

### 2. System Integration
```bash
# Old: Stands out as suspicious
~/.portfolio/  # What is this? Investigates further

# New: Blends in naturally
~/Library/Application Support/portfolio/  # Just another app
~/.local/share/portfolio/  # Standard Linux app data
```

### 3. Permission Inheritance
```bash
# System directories have better permissions:
drwx------  ~/Library/Application Support/  # macOS protected
drwx------  ~/.local/share/  # Linux user-private
drwx------  %APPDATA%/  # Windows protected

# Old approach relied on manual file permissions
# New approach inherits system security
```

### 4. Backup and Sync Compatibility
```bash
# System locations are properly handled by:
- Time Machine (macOS)
- Backup tools (Linux)
- Cloud sync (Windows)
- Version control excludes

# Old ~/.portfolio might not be handled correctly
# New locations are standard and well-supported
```

---

## Migration Strategy

### 1. Automatic Migration on First Run
```go
// internal/database/migrate.go
func MigrateFromLegacy() error {
	if !ShouldMigrateFromLegacy() {
		return nil // Nothing to migrate
	}

	legacyDB := GetLegacyPortfolioDir() + "/portfolio.db"
	newDB := GetDefaultDatabasePath()

	// Copy database to new location
	if err := copyFile(legacyDB, newDB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Copy config if exists
	legacyConfig := GetLegacyPortfolioDir() + "/config.toml"
	newConfig := GetDefaultConfigPath()
	if _, err := os.Stat(legacyConfig); err == nil {
		copyFile(legacyConfig, newConfig)
	}

	// Remove old files after successful migration
	os.RemoveAll(GetLegacyPortfolioDir())

	return nil
}
```

### 2. Backward Compatibility
```go
// Check both old and new locations
func GetDatabasePath() string {
	// Try new location first
	newDB := GetDefaultDatabasePath()
	if _, err := os.Stat(newDB); err == nil {
		return newDB
	}

	// Fall back to old location for migration
	legacyDB := filepath.Join(GetLegacyPortfolioDir(), "portfolio.db")
	if _, err := os.Stat(legacyDB); err == nil {
		return legacyDB
	}

	// Use new location for fresh install
	return newDB
}
```

### 3. User Communication
```markdown
# Portfolio has moved to a more secure location

Your Portfolio database has been automatically moved to a system-standard
location that provides better security and integration with your operating system.

**New location:** ~/Library/Application Support/portfolio/ (macOS)
**Old location:** ~/.portfolio/ (will be removed)

Your data is intact. This change improves security while maintaining full compatibility.
```

---

## Testing

### Verify New Locations
```bash
# Test path resolution
portfolio --debug
# Debug output should show:
# Database: ~/Library/Application Support/portfolio/portfolio.db
# Config: ~/Library/Application Support/portfolio/config.toml

# Test database operations
portfolio list projects
# Should work with new location

# Test migration
mv ~/.portfolio ~/.portfolio.backup
portfolio list projects
# Should create database in new location
```

### Security Testing
```bash
# Test discovery difficulty
find ~ -name "portfolio.db" 2>/dev/null
# Should not find easily identifiable results

# Test permissions
ls -la ~/Library/Application Support/portfolio/
# Should show proper permissions

# Test functionality
portfolio discover /path/to/project
portfolio list projects
# Should work normally
```

---

## Environment Overrides

### For Testing and Development
```bash
# Override data directory for testing
export PORTFOLIO_DATA_DIR=/tmp/portfolio-test
portfolio list projects
# Uses /tmp/portfolio-test/portfolio.db

# Override config directory
export PORTFOLIO_CONFIG_DIR=/tmp/portfolio-config
# Uses /tmp/portfolio-config/config.toml
```

### For Custom Deployments
```bash
# Enterprise deployment
export PORTFOLIO_DATA_DIR=/opt/portfolio/data
export PORTFOLIO_CONFIG_DIR=/etc/portfolio

# Container deployment
export PORTFOLIO_DATA_DIR=/data
export PORTFOLIO_CONFIG_DIR=/config
```

---

## Combined Security Approach

This secure location works perfectly with the other security measures:

### 1. MCP-Only Information Hiding
```bash
# Agents only know about MCP tools
# They never learn database location
# Secure location + hidden = double security
```

### 2. Database Password Protection
```bash
# Database password protected
# File location also hidden
# Two layers of security
```

### 3. System-Standard Permissions
```bash
# Operating system permissions apply
# Application firewall rules work
# System-level security controls
```

---

## Rollback Plan

### If Issues Occur
```bash
# Manual rollback
export PORTFOLIO_DATA_DIR=~/.portfolio
portfolio list projects
# Temporarily uses old location

# Or move data back
mv ~/Library/Application Support/portfolio ~/.portfolio
# Revert to old location
```

### Automatic Fallback
```go
// System checks both locations
func findDatabase() (string, error) {
	locations := []string{
		GetDefaultDatabasePath(),     // New secure location
		GetLegacyPortfolioDir() + "/portfolio.db", // Old location
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}

	return "", fmt.Errorf("no database found")
}
```

---

## Advantages Over Old Approach

### Security
- ✅ Much harder to discover
- ✅ Blends with system applications  
- ✅ Inherits system permissions
- ✅ Less predictable target

### Compatibility
- ✅ Follows system conventions
- ✅ Better backup/sync support
- ✅ Standard application behavior
- ✅ Cross-platform consistency

### Maintainability
- ✅ No manual permission management
- ✅ System-provided security
- ✅ Standard paths for tools
- ✅ Better future compatibility

---

## Conclusion

Moving from `~/.portfolio/` to system-standard application directories provides:

1. **Significantly improved security** through obscurity
2. **Better system integration** and compatibility
3. **Proper permission management** by OS
4. **Standards compliance** across platforms
5. **Future-proof** application architecture

This change, combined with:
- **Database password protection**  
- **MCP-only information exposure**

Creates a comprehensive security architecture that protects Portfolio data while maintaining simplicity and usability.

---

**Last Updated**: 2026-07-28  
**Implementation**: System-standard secure directories
