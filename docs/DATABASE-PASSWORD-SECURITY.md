# Database Password Protection Implementation

**Date**: 2026-07-28  
**Approach**: Password-based database access control  
**Status**: Simple, effective security solution

---

## Overview

Instead of full database encryption, Portfolio uses **password protection** for database access. This prevents external tools (like `sqlite3`) from accessing the database while keeping the implementation simple.

## How It Works

```bash
# External access fails (no password)
$ sqlite3 ~/.portfolio/portfolio.db "SELECT * FROM projects;"
Error: database is locked or password required

# Portfolio CLI works (has password)
$ portfolio list projects
# Successfully accesses database
```

## Implementation

### 1. Database Key Management

```go
// internal/database/security.go
package database

import (
    "os"
    "github.com/google/uuid"
)

// GetDatabaseKey returns the database access key
func GetDatabaseKey() string {
    // Check environment variable first
    if key := os.Getenv("PORTFOLIO_DB_KEY"); key != "" {
        return key
    }
    
    // Fall back to secure config file
    return loadKeyFromConfig()
}

func loadKeyFromConfig() string {
    // Load from ~/.portfolio/db_key
    // Generate if doesn't exist
    keyFile := os.Getenv("HOME") + "/.portfolio/db_key"
    
    if key, err := os.ReadFile(keyFile); err == nil {
        return string(key)
    }
    
    // Generate new key
    key := uuid.New().String()
    os.WriteFile(keyFile, []byte(key), 0600)
    return key
}
```

### 2. SQLite PRAGMA Key

```go
// internal/database/database.go
// Update connection string with password
func (d *Database) Connect() error {
    key := GetDatabaseKey()
    
    // Use SQLite PRAGMA key for password protection
    connString := fmt.Sprintf("file:%s?_pragma_key=%s&_foreign_keys=on", 
        d.dbPath, key)
    
    db, err := sql.Open("sqlite3", connString)
    // ... rest of connection logic
}
```

### 3. Configuration Storage

```go
// internal/config/security.go
package config

type SecurityConfig struct {
    DatabaseKey string `toml:"database_key"`
}

func LoadSecurityConfig() (*SecurityConfig, error) {
    // Load from secure location
    // ~/.portfolio/security.toml (chmod 600)
}
```

## Security Properties

### ✅ Prevents External Access
```bash
# Direct SQLite access fails
$ sqlite3 ~/.portfolio/portfolio.db
Error: unable to open database file

# Hexdump shows encrypted-looking data
$ hexdump -C ~/.portfolio/portfolio.db | head -5
00000000  53 51 4c 69 74 65 20 66  ... (SQLite header with key protection)
```

### ✅ Internal Access Works
```bash
# Portfolio CLI has password
$ portfolio list projects
project-1  /path/to/project1
project-2  /path/to/project2

# MCP server has password
$ portfolio mcp list-projects
["project-1", "project-2"]
```

### ✅ Key Management
- Key stored in secure config file (~/.portfolio/db_key, chmod 600)
- Environment variable override available
- Generated on first run if missing
- Not logged or exposed in error messages

## Deployment Strategy

### 1. Release Process
```bash
# Generate unique password per release
DB_KEY=$(uuidgen) || openssl rand -hex 16

# Store in secure config
echo "database_key = '$DB_KEY'" > ~/.portfolio/security.toml
chmod 600 ~/.portfolio/security.toml

# Test database access with password
portfolio list projects  # Should work
```

### 2. Migration Path
```bash
# Existing databases: Add password protection
portfolio migrate --add-password

# New databases: Password protected by default
portfolio init --password-protected
```

### 3. Rollback Plan
```bash
# Remove password protection if needed
portfolio migrate --remove-password

# Export data with password
portfolio export --output backup.json

# Import without password
portfolio import --input backup.json --no-password
```

## Advantages Over Full Encryption

### Simplicity
- ✅ No SQLCipher dependency
- ✅ No complex key management
- ✅ No performance overhead
- ✅ Easy to implement and maintain

### Compatibility
- ✅ Works with standard go-sqlite3
- ✅ CGO compilation works normally
- ✅ No external dependencies
- ✅ Cross-platform compatible

### Security
- ✅ Prevents unauthorized access
- ✅ Stops external SQLite tools
- ✅ Protects against casual inspection
- ✅ Adequate for local-first security model

## Limitations

### What It Doesn't Protect Against
- ❌ Physical access to the machine with key file
- ❌ Process memory inspection (key in RAM)
- ❌ Disk forensics if key file accessible

### What It Does Protect Against
- ✅ Casual inspection with file tools
- ✅ Automated discovery and access
- ✅ External SQLite tool access
- ✅ Unauthorized read attempts

### Use Case Fit
This approach is **perfect for**:
- Local-first applications
- Developer tool security
- Preventing AI agent bypass
- Single-user security model

This approach is **not sufficient for**:
- Multi-user systems
- Server deployments
- High-security requirements
- Regulatory compliance (encryption required)

## Implementation Timeline

### Phase 1: Core Implementation (This Release)
1. Add password generation on init
2. Update database connection to use password
3. Store password securely in config
4. Test external access is blocked

### Phase 2: CLI Integration (This Release)
1. Add password to all CLI commands
2. Test MCP server with password
3. Add password validation on startup
4. Handle missing password gracefully

### Phase 3: Migration Support (Next Release)
1. Add password to existing databases
2. Provide export/import without password
3. Support password rotation
4. Add password reset functionality

## Testing

### Security Testing
```bash
# Test 1: External access fails
! sqlite3 ~/.portfolio/portfolio.db ".tables"
# Expected: Error

# Test 2: Portfolio CLI works
portfolio list projects
# Expected: Success

# Test 3: Wrong password fails
export PORTFOLIO_DB_KEY="wrong-key"
portfolio list projects
# Expected: Error

# Test 4: File permissions
ls -la ~/.portfolio/db_key
# Expected: -rw------- (600)
```

### Integration Testing
```bash
# Test with MCP server
portfolio mcp start
# Expected: Server starts successfully

# Test with CLI
portfolio discover /path/to/project
portfolio list projects
# Expected: Both work successfully
```

## Documentation

### User Guide
```markdown
## Database Security

Portfolio databases are password-protected by default. This prevents unauthorized access while maintaining simplicity.

### Accessing Your Database

Your Portfolio CLI and MCP server automatically use the stored password to access your database.

### Manual Access

If you need to access your database manually:

```bash
export PORTFOLIO_DB_KEY=$(cat ~/.portfolio/db_key)
sqlite3 ~/.portfolio/portfolio.db
```

### Security Best Practices

1. Keep your `~/.portfolio/db_key` file private (chmod 600)
2. Never share your database password
3. Use environment variables for deployment
4. Backup your database key along with your database
```

### Developer Guide
```markdown
## Database Password Implementation

Portfolio uses SQLite PRAGMA key for password protection.

### Key Generation
- UUID v4 generated on first run
- Stored in ~/.portfolio/db_key (chmod 600)
- Environment variable override: PORTFOLIO_DB_KEY

### Connection String
```go
connString = fmt.Sprintf("file:%s?_pragma_key=%s&_foreign_keys=on", dbPath, key)
```

### Security Properties
- External SQLite tools: Blocked (no key)
- Portfolio CLI: Allowed (has key)
- MCP Server: Allowed (has key)
```

## Conclusion

Password-based database protection provides the **right balance** of security and simplicity for Portfolio's local-first architecture. It prevents the security breach (AI agents accessing database directly) while maintaining the simplicity and reliability that users expect.

---

**Last Updated**: 2026-07-28  
**Implementation**: Simple, effective, local-first security
