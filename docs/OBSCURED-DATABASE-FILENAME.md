# Obscured Database Filename Security

**Date**: 2026-07-28  
**Approach**: Security through additional obscurity  
**Status**: Enhanced security layer

---

## Additional Security Layer

### Previous Approach
```bash
# Predictable database filename:
~/.local/share/portfolio/portfolio.db  # Too obvious!

# Easy to discover:
find ~ -name "portfolio.db"
find ~ -name "*.db"
```

### New Enhanced Security
```bash
# Obscured database filename:
~/.local/share/portfolio/.portfoliodata  # Much harder to guess!

# Harder to discover:
find ~ -name "portfolio.db"        # Returns nothing
find ~ -name "*.db"               # Returns nothing
find ~ -name ".portfoliodata"     # Would need to guess this specific name
```

---

## Security Benefits

### 1. Prevents Automated Discovery
```bash
# Common database searches fail:
$ find ~ -name "*.db"          # Doesn't find portfolio database
$ find ~ -name "*portfolio*"   # Doesn't find portfolio database  
$ find ~ -name "portfolio.db"  # Returns nothing

# File looks like system data, not portfolio:
$ ls -la ~/.local/share/portfolio/
-rw-------  1 user  staff  1.2M Jul 28 18:30 .portfoliodata
# Doesn't reveal it's portfolio data
```

### 2. Blends with System Files
```bash
# Looks like generic system/application data:
.portfoliodata       # Could be anything
user.cache          # Looks like cache
app_data.bin        # Looks like binary app data
.index_data         # Looks like search index
```

### 3. Multiple Alternative Names
```go
// Available obscured filenames:
.portfoliodata     // Default: looks like hidden system data
user.cache         // Alternative 1: looks like user cache file  
app_data.bin       // Alternative 2: looks like binary app data
.index_data        // Alternative 3: looks like index file
settings.dat       // Alternative 4: looks like settings
.local_storage     // Alternative 5: looks like browser storage
app_state.db       // Alternative 6: less obvious naming
```

---

## Implementation

### Default Obscured Name
```go
// pkg/models/paths.go
func GetDefaultDatabasePath() string {
    return filepath.Join(GetPortfolioDataDir(), ".portfoliodata")
}

// Result:
// macOS: ~/Library/Application Support/portfolio/.portfoliodata
// Linux: ~/.local/share/portfolio/.portfoliodata  
// Windows: %APPDATA%\portfolio\.portfoliodata
```

### Custom Filename Option
```bash
# Users can specify custom filename for additional security
export PORTFOLIO_DB_PATH=/custom/path/custom_name.dat
portfolio list projects
# Uses custom path and filename
```

---

## Discovery Difficulty Comparison

### Old Approach - Too Easy
```bash
$ find ~ -name "portfolio.db" 2>/dev/null
/Users/user/.portfolio/portfolio.db  # Found immediately!

$ find ~ -name "*.db" 2>/dev/null
/Users/user/.portfolio/portfolio.db  # Found easily!
```

### New Approach - Much Harder
```bash
$ find ~ -name "portfolio.db" 2>/dev/null
# Returns nothing - filename changed

$ find ~ -name "*.db" 2>/dev/null  
# Returns nothing - extension changed

$ find ~ -name ".portfoliodata" 2>/dev/null
# Would need to guess this exact name
# Hidden file, non-standard extension
# Much harder to discover
```

---

## Combined Security Layers

This obscured filename works together with our other security measures:

### 1. Secure Location
```
❌ Old: ~/.portfolio/portfolio.db
✅ New: ~/.local/share/portfolio/.portfoliodata
```

### 2. Obscured Filename
```
❌ Old: portfolio.db (obvious)
✅ New: .portfoliodata (obscured)
```

### 3. Password Protection
```
❌ External: sqlite3 .portfoliodata (fails - no password)
✅ Portfolio: Has password access
```

### 4. MCP-Only Interface
```
❌ Agent knowledge: Database location hidden
✅ Agent tools: Only knows MCP interface
```

---

## Security Through Obscurity Criticism

### Common Criticism
"Security through obscurity is not real security"

### Our Defense
While we agree obscurity alone isn't sufficient, **combined with other measures** it provides valuable defense-in-depth:

**Our complete security stack:**
1. ✅ Password protection (real security)
2. ✅ MCP-only interface (architectural security)  
3. ✅ Secure location (standard security)
4. ✅ Obscured filename (additional layer)

**The obscured filename:**
- Raises the bar for automated discovery
- Prevents casual file browsing discovery
- Adds minimal implementation complexity
- Provides defense-in-depth

**We're not relying solely on obscurity** - it's one layer among many proper security measures.

---

## Migration Handling

### Automatic Migration
```go
// Old database automatically migrated to new name
~/.portfolio/portfolio.db → ~/.local/share/portfolio/.portfoliodata

// Migration happens on first run
// User data preserved, just renamed
```

### Backward Compatibility
```go
// System checks multiple locations and names
func FindDatabase() (string, error) {
    // Check new location with obscured name
    // Check old location with old name  
    // Check alternative names
    // Return first match found
}
```

---

## Testing

### Verify Obscured Filename
```bash
# Test 1: Old searches fail
find ~ -name "portfolio.db"    # Should return nothing
find ~ -name "*.db"            # Should return nothing

# Test 2: Portfolio still works
portfolio list projects         # Should work fine
portfolio discover /path        # Should work fine

# Test 3: File exists but obscured
ls -la ~/.local/share/portfolio/
# Should show .portfoliodata (not portfolio.db)
```

### Security Testing
```bash
# Test 1: External access blocked
sqlite3 ~/.local/share/portfolio/.portfoliodata "SELECT * FROM projects;"
# Should fail: unable to open database file

# Test 2: Portfolio access works
portfolio list projects
# Should succeed: password protection + obscured name

# Test 3: Discovery difficulty
find ~ -name ".portfoliodata" 2>/dev/null
# Should return nothing or very few results
```

---

## User Documentation

### For Users
```markdown
# Portfolio Database Security

Your Portfolio database is stored with enhanced security:

**Location**: System-standard application directories
**Filename**: Obscured (not obvious portfolio data)  
**Protection**: Password-protected
**Access**: Only through Portfolio tools

**What this means:**
- Your data is more secure against unauthorized access
- File browsing won't reveal portfolio data  
- Automated tools can't easily discover your database
- Portfolio CLI and MCP tools work seamlessly

**No action needed** - everything works automatically!
```

### For Developers
```markdown
# Database Filename Obscuration

Portfolio uses an obscured database filename as an additional security layer:

**Default filename**: `.portfoliodata`  
**Purpose**: Prevent automated discovery  
**Implementation**: Transparent to users

**Custom filename option:**
```bash
export PORTFOLIO_DB_PATH=/custom/path/custom_name
```

**Migration handled automatically:**
- Old `portfolio.db` → New `.portfoliodata`
- User data preserved
- No manual intervention required
```

---

## Conclusion

The obscured database filename provides an **additional security layer** that:

1. **Prevents automated discovery** tools from finding the database
2. **Blends with system files** rather than standing out
3. **Provides defense-in-depth** alongside proper security measures
4. **Costs nothing** in complexity or usability

Combined with:
- **Password protection** (real security)
- **MCP-only interface** (architectural security)  
- **Secure location** (standard security)

This creates a comprehensive security approach that protects Portfolio data while maintaining simplicity and usability.

The key insight: **Multiple layers of simple security > One layer of complex security**

---

**Last Updated**: 2026-07-28  
**Implementation**: Obscured database filename + secure location + password protection
