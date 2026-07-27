# Portfolio CLI & MCP UAT Plan

## Overview
User Acceptance Testing plan for Portfolio Engine CLI and MCP interface. Tests end-to-end functionality without dashboard.

---

## Pre-Test Setup

### Prerequisites
- [ ] Go installed and working
- [ ] Portfolio binary built (`go build`)
- [ ] Test environment with sample git repositories
- [ ] Claude Code installed (for MCP integration tests)

### Test Data Preparation
- [ ] Create test project directory with 3-5 git repos
- [ ] Mix of repo types (bare git, GitHub, GitLab)
- [ ] At least one repo with README.md
- [ ] At least one repo with code files

---

## Phase 1: CLI UAT

### 1.1 Initialization (`portfolio init`)

**Test Case 1.1.1: First-time initialization**
```bash
# Run
portfolio init

# Expected
- Interactive prompt for project roots
- Interactive prompt for database path
- Interactive prompt for log level
- Confirmation summary shown
- After confirmation: config file created, database initialized
```

**Verify:**
- [ ] Config file exists at `~/.portfolio/config.toml`
- [ ] Database file created at specified path
- [ ] `portfolio status` shows "Running"

**Test Case 1.1.2: Re-initialization protection**
```bash
# Run
portfolio init

# Expected
- Error or warning that already initialized
- Suggestion to use `portfolio config` commands instead
```

---

### 1.2 Status Check (`portfolio status`)

**Test Case 1.2.1: Healthy system status**
```bash
# Run
portfolio status

# Expected Output
- Configuration: ✓ Accessible (/path/to/config.toml)
- Database: ✓ Accessible (/path/to/portfolio.db)
- Projects Discovered: 0
- Last Discovery: Never
- Engine Status: Running
```

**Verify:**
- [ ] All checks show ✓
- [ ] Paths are correct absolute paths
- [ ] Project count is 0 (before discovery)

**Test Case 1.2.2: Status with projects**
```bash
# After running discovery
portfolio status

# Expected
- Projects Discovered: > 0
- Last Discovery: timestamp
```

---

### 1.3 Configuration Management (`portfolio config`)

**Test Case 1.3.1: List project roots**
```bash
# Run
portfolio config list-roots

# Expected
- List of configured roots
- Each with ✓ or ✗ status indicator
- Total count shown
```

**Verify:**
- [ ] All roots shown
- [ ] Status indicators accurate (test by deleting a root path)

**Test Case 1.3.2: Add project root**
```bash
# Run
portfolio config set-root /path/to/projects

# Expected
- Success message: "✓ Added project root: /path/to/projects"
- Total count incremented
```

**Edge Cases:**
- [ ] Add duplicate path → "already configured" message
- [ ] Add non-existent path → error message
- [ ] Add file instead of directory → error message
- [ ] Add path with trailing slash → normalized

**Test Case 1.3.3: Remove project root**
```bash
# Run
portfolio config remove-root /path/to/projects

# Expected
- Success message: "✓ Removed project root"
- Total count decremented
```

**Edge Cases:**
- [ ] Remove non-existent path → "not found" message
- [ ] Remove last root → error, at least one required
- [ ] Path normalization works

---

### 1.4 Diagnostics (`portfolio doctor`)

**Test Case 1.4.1: Full system diagnostics**
```bash
# Run
portfolio doctor

# Expected Sections
- Configuration Check
- Database Check
- Project Roots Check
- File Permissions Check
- Disk Space Check
- System Check

# Verify each section
- [ ] All checks pass with ✓
- [ ] Actionable remediation for failures
```

**Test Case 1.4.2: Induce failures**
```bash
# Test failure handling
# 1. Corrupt config file
# 2. Delete database file
# 3. Make config unreadable (chmod 000)

# Expected
- Clear ✗ indicators
- Specific error messages
- Remediation suggestions
- Exit code 1
```

**Test Case 1.4.3: Claude integration diagnostics**
```bash
# After installing Claude integration
portfolio doctor claude

# Expected
- Integration status check
- MCP server connectivity check
- Configuration file validation
- Skills installation check
- All pass with ✓
```

---

### 1.5 HTTP API Server (`portfolio serve`)

**Test Case 1.5.1: Start server**
```bash
# Run
portfolio serve --port 8080

# Expected
- Message: "Portfolio API server listening on :8080"
- Server runs until SIGINT/SIGTERM

# While running, test endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/projects
```

**Verify:**
- [ ] Health endpoint returns 200
- [ ] Projects endpoint returns JSON array
- [ ] Server stops gracefully on Ctrl+C

**Test Case 1.5.2: Custom port**
```bash
portfolio serve --port 3000

# Expected
- Server starts on :3000
```

---

### 1.6 Integration Management

**Test Case 1.6.1: Install Claude integration**
```bash
# Run
portfolio install claude

# Expected
- Integration registered with database
- MCP server configured in Claude settings
- Skills installed
- Success message with version info
```

**Verify:**
- [ ] `portfolio doctor claude` passes
- [ ] Claude Code can connect to MCP server
- [ ] MCP tools available in Claude Code

**Test Case 1.6.2: Force reinstall**
```bash
portfolio install claude --force

# Expected
- Reinstalls even if already installed
- No errors
```

**Test Case 1.6.3: Upgrade Claude integration**
```bash
portfolio upgrade claude

# Expected
- If up-to-date: "already up to date"
- If upgrade available: performs upgrade
- Shows version changes
```

**Test Case 1.6.4: Uninstall Claude integration**
```bash
portfolio uninstall claude

# Expected
- Integration removed
- MCP config removed from Claude Code
- Skills removed
- Project data preserved in database
```

---

## Phase 2: MCP Server UAT

### 2.1 MCP Server Startup

**Test Case 2.1.1: Start MCP server**
```bash
# Run (stdio mode)
portfolio mcp

# Expected
- Server starts listening on stdin/stdout
- No output to stdout (MCP protocol only)
- Logs to stderr
```

**Test Case 2.1.2: MCP initialization**
```bash
# Send MCP initialize request via Claude Code
# Expected
- Server responds with capabilities
- Lists available tools
- Server info returned
```

---

### 2.2 Discovery Tools

**Test Case 2.2.1: MCP tool - discoverProjects**
```bash
# Via Claude Code, call discoverProjects

# Expected Response
{
  "discovered": <number>,
  "error_count": <number>,
  "roots_checked": <number>
}
```

**Verify:**
- [ ] Projects added to database
- [ ] `portfolio status` shows updated count
- [ ] Discovery skips ignored paths (.git, node_modules)

**Test Case 2.2.2: Empty project roots**
```bash
# Configure no project roots
# Call discoverProjects

# Expected
- Error: "no project roots configured"
```

---

### 2.3 Project Query Tools

**Test Case 2.3.1: MCP tool - listProjects**
```bash
# Call listProjects

# Expected
- Array of all projects
- Each project: id, name, root_path, repository_type, discovered_at, updated_at
```

**Verify:**
- [ ] All discovered projects listed
- [ ] Fields are complete
- [ ] Timestamps valid ISO 8601

**Test Case 2.3.2: MCP tool - getProject**
```bash
# Call getProject with project ID

# Expected
{
  "project": { <project details> },
  "metadata": { <metadata: git HEAD, languages, frameworks, dependencies> }
}
```

**Edge Cases:**
- [ ] Invalid ID → "project not found"
- [ ] Empty ID → "id is required"

**Test Case 2.3.3: MCP tool - searchProjects**
```bash
# Call searchProjects with query

# Expected
{
  "results": [<matching projects>],
  "query": "<search term>",
  "count": <number>
}
```

**Verify:**
- [ ] Search matches project names
- [ ] Returns partial matches
- [ ] Limited to 50 results
- [ ] Results sorted by name

**Test Case 2.3.4: MCP tool - searchDocumentation**
```bash
# Call searchDocumentation with query

# Expected
{
  "results": [
    {
      "id": <doc ID>,
      "project_id": <project ID>,
      "project": <project name>,
      "path": <file path>,
      "kind": <README, CONTRIBUTING, etc>,
      "content": <preview>
    }
  ],
  "query": "<search term>",
  "count": <number>
}
```

**Verify:**
- [ ] Searches documentation content
- [ ] Returns content preview (first 500 chars)
- [ ] Limited to 50 results

---

### 2.4 Analysis Tools

**Test Case 2.4.1: MCP tool - getAnalysis**
```bash
# Call getAnalysis with project_id

# Expected
{
  "analyses": [<analysis objects>],
  "count": <number>
}
```

**Verify:**
- [ ] Returns all analyses for project
- [ ] Empty array if no analyses
- [ ] Includes summary, purpose, architecture, etc.

**Test Case 2.4.2: MCP tool - storeAnalysis**
```bash
# Call storeAnalysis with:
# - project_id (required)
# - analyzer (required)
# - analyzed_git_head (optional, defaults to current)
# - summary, purpose, architecture, maturity, strengths, weaknesses,
#   reusable_components, notes, raw_json (all optional)

# Expected
{
  "id": <new analysis ID>,
  "project_id": <project ID>,
  "analyzed_at": <timestamp>
}
```

**Verify:**
- [ ] Analysis persisted to database
- [ ] UUID generated
- [ ] Validation errors for invalid raw_json
- [ ] getAnalysis returns new analysis

**Test Case 2.4.3: MCP tool - listProjectsNeedingAnalysis**
```bash
# Call listProjectsNeedingAnalysis

# Expected
{
  "no_analysis": [
    {"id": "...", "name": "...", "path": "..."}
  ],
  "stale_analysis": [
    {"id": "...", "name": "...", "path": "...",
     "analyzed_at": "...", "analyzed_git_head": "...", "current_git_head": "..."}
  ],
  "counts": {
    "no_analysis": <number>,
    "stale_analysis": <number>,
    "total": <number>
  }
}
```

**Verify:**
- [ ] Projects without any analysis appear in `no_analysis` array
- [ ] Projects with stale analysis (git HEAD changed) appear in `stale_analysis` array
- [ ] `stale_analysis` entries include analyzed_at, analyzed_git_head, and current_git_head
- [ ] Projects with up-to-date analysis excluded from both arrays
- [ ] `counts` fields match array lengths

---

### 2.5 Configuration Tools

**Test Case 2.5.1: MCP tool - getConfiguration**
```bash
# Call getConfiguration

# Expected
{
  "key1": "value1",
  "key2": "value2",
  ...
}
```

**Verify:**
- [ ] Returns all configuration key-value pairs
- [ ] Empty object if no config

**Test Case 2.5.2: MCP tool - updateConfiguration**
```bash
# Call updateConfiguration with key and value

# Expected
{
  "key": "<key>",
  "value": "<value>",
  "status": "updated"
}
```

**Verify:**
- [ ] Configuration persisted
- [ ] getConfiguration returns updated value
- [ ] Error for empty key

---

### 2.6 Relationship Tools

**Test Case 2.6.1: MCP tool - listRelationships**
```bash
# Call listRelationships with project_id

# Expected
- Array of relationships for project
- Each: id, source_project, target_project, type, description, confidence
```

**Verify:**
- [ ] All relationships returned
- [ ] Empty array if none

**Test Case 2.6.2: MCP tool - storeRelationship**
```bash
# Call storeRelationship with:
# - source_project (required)
# - target_project (required)
# - type (required: Similar, Evolution, Shared Feature, Shared Technology, Reuses Component)
# - description (optional)
# - confidence (optional, 0-1)

# Expected
{
  "id": <new relationship ID>,
  "source_project": <source ID>,
  "target_project": <target ID>,
  "type": <relationship type>
}
```

**Verify:**
- [ ] Relationship persisted
- [ ] UUID generated
- [ ] Error for invalid type
- [ ] Error for non-existent projects

---

### 2.7 Health Check

**Test Case 2.7.1: MCP tool - health**
```bash
# Call health

# Expected
{
  "status": "healthy",
  "database_connected": true,
  "project_count": <number>
}
```

**Verify:**
- [ ] Status is "healthy" when database connected
- [ ] Status is "unhealthy" when database disconnected
- [ ] project_count accurate

---

## Phase 3: Integration Testing

### 3.1 Claude Code Integration

**Test Case 3.1.1: End-to-end Claude Code workflow**
```bash
# In Claude Code, after integration installed
1. Ask: "What projects do I have?"
   - Claude calls listProjects
   - Responds with project list

2. Ask: "Tell me about the <project-name> project"
   - Claude calls getProject
   - Responds with project details and metadata

3. Ask: "Search for projects containing 'auth'"
   - Claude calls searchProjects
   - Responds with matching projects

4. Ask: "What documentation mentions 'deployment'?"
   - Claude calls searchDocumentation
   - Responds with relevant docs

5. Ask: "Analyze the <project-name> project"
   - Claude calls getProject, then storeAnalysis
   - Analysis persisted
```

**Verify:**
- [ ] Each MCP call succeeds
- [ ] Claude Code provides coherent responses
- [ ] No MCP protocol errors

**Test Case 3.1.2: MCP server resilience**
```bash
# Test error handling
- Call tool with missing required parameters
- Call with invalid project ID
- Call with invalid relationship type

# Expected
- Clear error messages
- No server crashes
- Server continues processing
```

---

## Phase 4: Edge Cases & Error Handling

### 4.1 CLI Edge Cases

**Test Case 4.1.1: Missing config file**
```bash
rm ~/.portfolio/config.toml
portfolio status

# Expected
- Error: "run 'portfolio init'"
```

**Test Case 4.1.2: Corrupted database**
```bash
# Corrupt SQLite database
portfolio status

# Expected
- Clear error message
- Suggestion to run diagnostics
```

**Test Case 4.1.3: Permission denied**
```bash
chmod 000 ~/.portfolio/config.toml
portfolio config list-roots

# Expected
- Permission error
- Remediation suggestion
```

**Test Case 4.1.4: Disk space exhaustion**
```bash
# Fill disk, then run discovery
# Expected
- Graceful error
- No partial writes
```

---

### 4.2 MCP Edge Cases

**Test Case 4.2.1: Concurrent requests**
```bash
# Send multiple MCP requests simultaneously
# Expected
- All processed correctly
- No race conditions
```

**Test Case 4.2.2: Malformed MCP protocol**
```bash
# Send invalid JSON
# Expected
- Protocol error response
- Server continues running
```

**Test Case 4.2.3: Database disconnection mid-session**
```bash
# Start MCP server
# Delete database file
# Make request

# Expected
- Error response
- Suggestion to check connectivity
```

---

## Phase 5: Performance & Scalability

### 5.1 CLI Performance

**Test Case 5.1.1: Large project discovery**
```bash
# Configure root with 100+ projects
portfolio mcp
# Call discoverProjects

# Expected
- Completes in reasonable time (< 30s for 100 projects)
- Progress feedback (if any)
- All projects discovered
```

**Test Case 5.1.2: Large project listing**
```bash
# With 1000+ projects
# Call listProjects

# Expected
- Returns all projects
- Response size manageable
```

---

### 5.2 MCP Performance

**Test Case 5.2.1: Search performance**
```bash
# With 1000+ projects and documents
# Call searchProjects and searchDocumentation

# Expected
- Results in < 1s
- Limited result set (50 items)
```

**Test Case 5.2.2: Analysis retrieval**
```bash
# With project having 100+ analyses
# Call getAnalysis

# Expected
- Returns all analyses
- Response time reasonable
```

---

## Phase 6: Data Integrity

### 6.1 Database Integrity

**Test Case 6.1.1: Concurrent writes**
```bash
# Run multiple discoverProjects simultaneously
# Expected
- No database corruption
- All projects recorded
```

**Test Case 6.1.2: Schema validation**
```bash
# Check database schema
sqlite3 portfolio.db ".schema"

# Expected
- All tables present
- Foreign keys enforced
- Indexes present
```

---

## Test Execution Checklist

### Pre-UAT
- [ ] Backup existing Portfolio data
- [ ] Note current schema version
- [ ] Document test environment

### During UAT
- [ ] Log each test case result
- [ ] Note deviations from expected
- [ ] Capture error messages
- [ ] Record timing for performance tests

### Post-UAT
- [ ] Document all failures
- [ ] Categorize issues (critical, major, minor)
- [ ] Verify rollback procedures work
- [ ] Clean up test data

---

## Exit Criteria

UAT considered complete when:

1. **All critical test cases pass** (90%+ pass rate)
2. **No data loss scenarios found**
3. **Error handling clear and actionable**
4. **Performance within acceptable bounds**
5. **Claude Code integration works end-to-end**

---

## Issue Reporting Template

For each failure found:

```
[Test Case X.X.X: <name>]

Steps to Reproduce:
1.
2.
3.

Expected:
<expected behavior>

Actual:
<actual behavior>

Severity: [Critical/Major/Minor]

Environment:
- OS:
- Go version:
- Portfolio version:
```

---

## Sign-off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| UAT Lead | | | |
| Engineering | | | |
| Product | | | |
