# Portfolio User Manual

Version: v0.3.4v0.3.3
Last Updated: July 26, 2026

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Core Commands](#core-commands)
6. [AI Agent Integration](#ai-agent-integration)
7. [API Usage](#api-usage)
8. [Troubleshooting](#troubleshooting)
9. [Advanced Usage](#advanced-usage)

---

## Introduction

Portfolio is a **local-first project inventory and knowledge platform** that enables developers and AI coding agents to understand an entire software portfolio.

**What Portfolio Does:**
- Automatically discovers all your Git repositories
- Extracts project metadata (languages, frameworks, dependencies)
- Indexes documentation for searchable access
- Stores semantic knowledge about your projects
- Enables AI agents to answer questions about your portfolio

**What Portfolio Is Not:**
- Project management software
- Issue tracking system
- Documentation authoring tool
- CI/CD platform
- Source control system

**Key Principle:** Install → Initialize → Forget. Portfolio runs in the background, keeping your portfolio knowledge current without requiring ongoing maintenance.

---

## Installation

### System Requirements

- **Operating System:** macOS, Linux (Windows support planned)
- **Git:** Version 2.0 or higher
- **Go:** Version 1.22.6 or higher (for building from source)
- **Disk Space:** ~50MB for installation, plus metadata storage per project

### Installation Methods

#### Method 1: One-Command Install (Recommended)

```bash
# Install and start in one command
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
portfolio init
```

#### Method 2: Binary Release

```bash
# Download latest release for your platform
# macOS (Intel): portfolio-darwin-amd64
# macOS (Apple Silicon): portfolio-darwin-arm64  
# Linux (Intel): portfolio-linux-amd64
# Linux (ARM): portfolio-linux-arm64

curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-amd64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/

# Verify installation
portfolio --version
```

#### Method 3: Build from Source

```bash
# Clone repository
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# Build binary
go build -o portfolio ./cmd/portfolio

# Install to PATH
sudo mv portfolio /usr/local/bin/
```

#### Method 3: Homebrew (macOS)

```bash
brew tap shafi-/portfolio
brew install portfolio
```

---

## Quick Start

### 1. Initialize Portfolio

```bash
# Initialize with default settings
portfolio init

# Initialize with custom database location
portfolio init --db-path ~/.portfolio/portfolio.db

# Initialize with specific project roots
portfolio init --roots ~/Projects ~/src
```

### 2. Discover Projects

```bash
# Discover all projects in configured roots
portfolio discover

# Discover with verbose output
portfolio discover --verbose
```

### 3. Explore Your Portfolio

```bash
# List all discovered projects
portfolio list projects

# Search for projects
portfolio search projects "authentication"
portfolio search documentation "API"

# Get project details
portfolio get project <project-id>
```

### 4. Check Health

```bash
# Check portfolio health
portfolio health

# Check database integrity
portfolio health --database

# Check MCP server status
portfolio health --mcp
```

---

## Configuration

### Configuration File

Portfolio stores configuration in `~/.portfolio/config.toml`:

```toml
[general]
# Database file path (default: ~/.portfolio/portfolio.db)
database_path = "/Users/username/.portfolio/portfolio.db"

# Log level: debug, info, warn, error (default: info)
log_level = "info"

[discovery]
# Project root directories to scan
roots = [
    "/Users/username/Projects",
    "/Users/username/src"
]

# Ignored directory names
ignore_dirs = ["node_modules", ".git", "vendor", "build", "dist"]

# Maximum depth for directory scanning (default: 5)
max_depth = 5

[mcp]
# MCP server settings
host = "localhost"
port = 3000

[ai]
# AI agent settings
analyzer = "claude-code"  # Default analyzer for AI analysis
```

### Environment Variables

```bash
# Override default config location
export PORTFOLIO_CONFIG="/custom/config.toml"

# Set custom database location
export PORTFOLIO_DB="/custom/portfolio.db"

# Enable verbose logging
export PORTFOLIO_VERBOSE="true"
```

### Configuration Commands

```bash
# Edit configuration
portfolio config edit

# Show current configuration
portfolio config show

# Reset to defaults
portfolio config reset

# Validate configuration
portfolio config validate
```

---

## Core Commands

### Discovery Commands

#### `portfolio discover`

Discover projects in configured root directories:

```bash
# Standard discovery
portfolio discover

# Verbose mode
portfolio discover --verbose
```

**What it does:**
- Scans configured root directories for Git repositories
- Extracts git metadata (commits, branches, HEAD)
- Detects programming languages
- Identifies frameworks and dependencies
- Stores discovered projects in database

### Project Management Commands

#### `portfolio list projects`

List all discovered projects:

```bash
# List all projects
portfolio list projects

# Filter by technology
portfolio list projects --filter language=go
portfolio list projects --filter framework=react

# Sort by last modified
portfolio list projects --sort updated

# Limit results
portfolio list projects --limit 10
```

#### `portfolio get project`

Get detailed project information:

```bash
# Get project details
portfolio get project <project-id>

# Include metadata
portfolio get project <project-id> --metadata

# Include documentation
portfolio get project <project-id> --docs
```

**Output includes:**
- Project name, path, repository type
- Git metadata (commits, branches, HEAD)
- Languages and frameworks used
- Dependencies
- Documentation files
- Analysis (if available)

#### `portfolio search projects`

Search projects by various criteria:

```bash
# Search by name
portfolio search projects "authentication"

# Search by language
portfolio search projects --language python

# Search by framework
portfolio search projects --framework django

# Search by dependency
portfolio search projects --dependency postgresql
```

#### `portfolio search documentation`

Search across all project documentation:

```bash
# Search documentation
portfolio search documentation "API endpoint"

# Search in specific project
portfolio search documentation --project <project-id> "authentication"

# Search with context (lines before/after)
portfolio search documentation --context 2 "OAuth"
```

### Diagnostic Commands

#### `portfolio health`

Check system health:

```bash
# Basic health check
portfolio health

# Comprehensive health check
portfolio health --full

# Check specific components
portfolio health --database
portfolio health --mcp
portfolio health --discovery
```

**Health checks include:**
- Database connectivity and integrity
- MCP server status (if running)
- Project metadata consistency
- Documentation index status
- Recent operation logs

---

## AI Agent Integration

Portfolio ships first-class integrations for two AI coding agents, with the same
capabilities (install / verify / upgrade / remove). Both install a local stdio MCP
server and a Portfolio skill rendered from one shared template.

### Claude Code

```bash
portfolio install claude          # register MCP server + install skill
portfolio doctor claude           # verify
portfolio upgrade claude          # upgrade
portfolio uninstall claude        # remove
portfolio install claude --force  # force reinstall
```

### OpenCode

OpenCode has no local-stdio MCP CLI, so Portfolio writes its schema-documented
config file (`~/.config/opencode/opencode.json`) — the official method per
**ADR-021**. The install is an idempotent read-merge-write that preserves your
other settings and any sibling MCP servers, and installs the Portfolio skill to
`~/.config/opencode/skills/portfolio/SKILL.md`.

```bash
portfolio install opencode        # register MCP server + install skill
portfolio doctor opencode         # verify
portfolio upgrade opencode        # upgrade
portfolio uninstall opencode      # remove (preserves other servers)
```

### Other agents

Portfolio works with any MCP-capable agent (Cursor, Zed, Continue, Roo, …). There
is no automated installer for these, but the connection is identical: point your
agent at the MCP server command `portfolio mcp`. For the canonical, always
up-to-date integration reference (the rendered skill plus connection details),
run `portfolio manual` or read `docs/agent-integration-manual.md`.

### What every integration installs

1. **MCP server registration** — Portfolio is registered as a local stdio MCP
   server (Claude Code via `claude mcp add`; OpenCode via its config file).
2. **Portfolio skill** — rendered from a single shared template (`SKILL_COMMON`)
   that documents every tool and the three-tier knowledge protocol. The skill
   records an `analyzer` identity so Portfolio can attribute analyses and
   features to the agent that produced them.
3. **Validation** — `portfolio doctor <agent>` verifies the server is registered,
   the skill is present, and all tools respond.

### Available MCP Tools

Every integrated agent can call the same **26** Portfolio tools:

**Discovery (4)**
- `health()` — check Portfolio health
- `discoverProjects()` — discover git repositories
- `listProjects()` — list all known projects
- `getProject(project_id)` — full project details (deterministic facts)

**Search (2)**
- `searchProjects(query)` — search projects by text
- `searchDocumentation(query)` — search indexed documentation

**Configuration (2)**
- `getConfiguration()` — read engine configuration
- `updateConfiguration(config)` — update engine configuration

**Project & code content (5)**
- `getProjectStructure(project_id)` — project file tree
- `listProjectFiles(project_id)` — list project files
- `getFileContent(project_id, path)` — read a single file (sensitive files are blocked)
- `searchFiles(query)` — search across project files
- `getDependencies(project_id)` — declared dependencies and versions

**Analysis (3)**
- `getAnalysis(project_id)` — read stored analysis
- `storeAnalysis(analysis)` — store AI analysis (Tier 2)
- `listProjectsNeedingAnalysis()` — projects whose analysis is stale or missing

**Features (3)**
- `listFeatures(project_id)` — list a project's features
- `searchFeatures(query | implementation_status | pattern)` — search features
- `storeFeature(...)` — create/update a feature (Tier 2/3, upsert by name)

**Technologies (5)**
- `listTechnologies()` — list known technologies
- `searchByTechnology(technology)` — projects using a technology
- `listProjectTechnologies(project_id)` — technologies in a project
- `storeTechnology(...)` — record a technology
- `tagProjectWithTechnology(project_id, technology)` — tag a project

**Relationships (2)**
- `listRelationships(project_id)` — list project relationships
- `storeRelationship(relationship)` — store a relationship

### Example Session

Once integrated, an agent can:

```text
User: What authentication mechanisms are used in my portfolio?

Claude: Let me search your portfolio for authentication implementations.
[Searches projects, returns results]

Your portfolio has 4 projects using authentication:
1. user-auth-service - JWT tokens with refresh rotation
2. api-gateway - OAuth 2.0 with PKCE
3. mobile-app - Biometric auth + JWT
4. admin-panel - Session-based auth

Would you like me to analyze the security of any of these implementations?
```

---

## API Usage

Portfolio provides an HTTP API for dashboard integration and custom tooling.

### Starting the HTTP Server

```bash
# Start API server (default port 8080)
portfolio api serve

# Start on custom port
portfolio api serve --port 3000

# Start with specific host
portfolio api serve --host 0.0.0.0 --port 3000
```

### API Endpoints

#### Projects

```bash
# List all projects
GET /api/projects
GET /api/projects?limit=10&offset=0

# Get project details
GET /api/projects/{id}

# Get project metadata
GET /api/projects/{id}/metadata

# Search projects
GET /api/projects/search?q=authentication
```

#### Documentation

```bash
# Search documentation
GET /api/documentation/search?q=API

# List project documentation
GET /api/projects/{id}/documentation
```

#### Analysis

```bash
# Get project analysis
GET /api/projects/{id}/analysis?analyzer=claude-code

# List projects needing analysis
GET /api/projects/analysis/needs-analysis
```

#### Relationships

```bash
# List project relationships
GET /api/projects/{id}/relationships
```

#### Health

```bash
# Health check
GET /api/health
```

### Example API Usage

```bash
# List all projects
curl http://localhost:8080/api/projects

# Get specific project
curl http://localhost:8080/api/projects/{project-id}

# Search projects
curl http://localhost:8080/api/projects/search?q=go

# Health check
curl http://localhost:8080/api/health
```

### API Response Format

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "user-auth-service",
    "root_path": "/Users/username/Projects/user-auth-service",
    "repository_type": "git",
    "discovered_at": "2026-07-24T00:00:00Z",
    "updated_at": "2026-07-24T10:30:00Z"
  }
}
```

---

## Troubleshooting

### Common Issues

#### Issue: "No projects discovered"

**Solution:**
```bash
# Check configured roots
portfolio config show | grep roots

# Add project roots
portfolio set-root /path/to/projects

# Run discovery with verbose output
portfolio discover --verbose
```

#### Issue: "Database locked" or "database busy"

**Solution:**
```bash
# Close any running MCP server
portfolio mcp stop

# Check for other portfolio processes
ps aux | grep portfolio

# Restart portfolio
portfolio health
```

#### Issue: "Claude Code not connecting"

**Solution:**
```bash
# Verify installation
portfolio doctor claude

# Check Claude Code configuration
# Look for portfolio entry in Claude's MCP servers config

# Reinstall integration
portfolio uninstall claude
portfolio install claude
```

#### Issue: "Slow discovery performance"

**Solution:**
```bash
# Reduce scan depth
portfolio config set discovery.max_depth 3

# Exclude more directories
portfolio config set discovery.ignore_dirs '["node_modules",".git","vendor","build","dist","target"]'

# Run discovery again
portfolio discover
```

### Diagnostic Commands

```bash
# Full health check
portfolio health --full

# Check database integrity
portfolio health --database

# View recent logs
portfolio logs --tail 50

# Export diagnostic info
portfolio doctor --export diagnostic-info.json
```

### Getting Help

```bash
# Show help
portfolio --help

# Show help for specific command
portfolio discover --help
portfolio install --help

# Get version info
portfolio --version
```

---

## Advanced Usage

### Custom Root Directories

```bash
# Add single root
portfolio set-root /path/to/projects

# Add multiple roots
portfolio set-root /path/to/projects1 /path/to/projects2

# Remove root
portfolio remove-root /path/to/projects

# List all roots
portfolio list-roots
```

### MCP Server Configuration

Portfolio includes a built-in MCP server for AI agent integration:

```bash
# Start MCP server (stdio mode)
portfolio mcp

# Start with custom configuration
portfolio mcp --config /custom/mcp-config.json

# Test MCP server connection
portfolio mcp --test
```

### Database Management

```bash
# Database location
portfolio db location

# Database statistics
portfolio db stats

# Database integrity check
portfolio db check

# Backup database
portfolio db backup /path/to/backup.db

# Restore database
portfolio db restore /path/to/backup.db
```

### Performance Tuning

```bash
# Increase scan depth for large projects
portfolio config set discovery.max_depth 8

# Enable parallel processing (experimental)
portfolio config set discovery.parallel true

# Adjust cache size
portfolio config set cache.size 1000
```

### Batch Operations

```bash
# Recalculate all metadata
portfolio metadata refresh --all

# Reindex all documentation
portfolio docs reindex --all

# Analyze all unanalyzed projects
portfolio analyze batch --unanalyzed
```

### Integration Management

```bash
# List installed integrations
portfolio integration list

# Show integration status
portfolio integration status claude

# Upgrade integration
portfolio integration upgrade claude

# Remove integration
portfolio integration remove claude
```

---

## Best Practices

### Project Organization

1. **Use consistent project structure** - Helps Portfolio identify similar patterns
2. **Keep meaningful README files** - Portfolio indexes these for search
3. **Use standard Git branching** - Portfolio tracks branch information
4. **Maintain clean commit history** - Improves metadata quality

### Regular Maintenance

1. **Run discovery periodically** - Keep project metadata current
2. **Run health checks** - Ensure database integrity
3. **Update roots** - Add new project directories as needed
4. **Review analysis freshness** - Update stale analyses

### AI Agent Usage

1. **Let Claude Code manage analysis** - AI agents decide when to analyze
2. **Ask specific questions** - Portfolio excels at targeted queries
3. **Explore relationships** - Ask about project connections and patterns
4. **Leverage semantic knowledge** - Use stored analysis for deep understanding

### Performance Tips

1. **Limit scan depth** - Reduces discovery time for large project trees
2. **Exclude unnecessary directories** - Ignore node_modules, vendor, etc.
3. **Use search instead of listing** - Search is faster than full listing for large portfolios
4. **Cache frequently accessed data** - Portfolio caches query results

---

## Security Considerations

### Data Privacy

- **Local-First by Design** - All data stays on your machine
- **No Cloud Dependencies** - No data is sent to external services
- **AI Agnostic** - Works with any AI agent, no lock-in

### Access Control

Portfolio currently has no built-in access control as it is designed for single-user local use. Future versions may include:
- User authentication
- Role-based access control
- Audit logging

### Sensitive Data

Portfolio stores:
- Project paths and metadata (non-sensitive)
- Git commit hashes (non-sensitive)
- File paths and sizes (non-sensitive)
- AI-generated analysis (potentially sensitive if it contains project details)

Portfolio does NOT store:
- Source code
- Credentials
- Personal information
- Proprietary algorithms

---

## FAQ

**Q: Can Portfolio handle private repositories?**  
A: Yes, Portfolio can scan any Git repository it has file system access to, regardless of whether it's public or private.

**Q: Does Portfolio work with monorepos?**  
A: Yes, Portfolio can discover and analyze projects within monorepos, treating each subdirectory as a separate project if it contains a Git repository.

**Q: How much disk space does Portfolio use?**  
A: Approximately 50MB for the application plus 1-5MB per project for metadata and documentation indexes.

**Q: Can I use Portfolio with multiple AI agents?**  
A: Yes, Portfolio is agent-agnostic. You can integrate it with Claude Code, Codex CLI, or any other AI agent that supports MCP.

**Q: Does Portfolio require internet access?**  
A: No, Portfolio works entirely offline. It only requires internet access when you want to install or upgrade the binary.

**Q: How often should I run discovery?**  
A: Run discovery whenever you add new projects or make significant changes to existing projects. Portfolio does not automatically track file system changes.

**Q: Can Portfolio analyze non-code projects?**  
A: Portfolio can discover any Git repository, but its metadata extraction is optimized for software projects. Documentation analysis works with any text-based files.

**Q: What happens if I delete a project?**  
A: Portfolio will show the project as "not found" in the file system but retains its metadata in the database. You can remove the stale project data with `portfolio remove-project <id>`.

---

## Getting Support

### Documentation

- **Full Documentation:** `/docs/` directory in repository
- **API Reference:** `docs/PlatformSpecification.md`
- **Knowledge Model:** `docs/KnowledgeModel.md`
- **Architecture Decisions:** `docs/ADR.md`

### Reporting Issues

Report bugs or feature requests at:  
https://github.com/shafi-/portfolio/issues

### Community

- **GitHub Discussions:** https://github.com/shafi-/portfolio/discussions
- **Documentation Updates:** Pull requests welcome

---

## Version History

**v0.2.0** (July 26, 2026)
- Official OpenCode integration (`portfolio install opencode`)
- `portfolio manual` command + committed `docs/agent-integration-manual.md`
- Deterministic dependency versions; feature deep-dive fields; `searchFeatures`
- ADR-021 (schema-documented config files are official methods)

**v0.1.0** (July 24, 2026)
- Initial release
- Core discovery and metadata extraction
- Documentation indexing
- MCP server and HTTP API
- Claude Code integration
- AI analysis support

---

**End of User Manual**

For the most up-to-date information, please check the repository documentation or run `portfolio --help`.
