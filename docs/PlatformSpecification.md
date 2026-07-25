
# PlatformSpecification.md

Version: 0.1

> This specification complements KnowledgeModel.md and defines the implementation
> contracts for the Portfolio platform.

==================================================================
0. CONFIGURATION FILE FORMAT
==================================================================

Portfolio stores configuration in ~/.portfolio/config.toml (TOML format).

Structure:

```toml
[general]
database_path = "/path/to/portfolio.db"

[discovery]
project_roots = ["/path/to/projects", "/another/path"]
ignored_paths = ["node_modules", ".git", "vendor", "build", "dist", "target", "bin"]

[logging]
level = "INFO"  # DEBUG, INFO, WARN, ERROR

[dashboard]
enabled = false
port = 8080
```

Field Naming Convention:
- All TOML field names use snake_case (e.g., `database_path`, `project_roots`, `ignored_paths`)
- This differs from Go struct field names which use camelCase
- TOML parsers map snake_case to camelCase automatically

Required Fields:
- general.database_path: Path to SQLite database file
- discovery.project_roots: Array of project root directories (at least one required)

Optional Fields:
- discovery.ignored_paths: Directories to skip during discovery (defaults shown above)
- logging.level: Log level (default: INFO)
- dashboard.*: Dashboard configuration (not yet implemented)

==================================================================
1. DATABASE SCHEMA
==================================================================

Core Tables

projects
- id (UUID, PK)
- name
- root_path
- repository_type
- discovered_at
- updated_at

metadata
- project_id (FK)
- git_head
- default_branch
- last_commit_at
- last_modified_at
- commit_count (INTEGER, default 0)
- language_summary
- framework_summary
- dependency_summary
- documentation_hash
- last_scan_at

documents
- id
- project_id
- path
- kind (README, ADR, DOC, CHANGELOG, ...)
- content
- content_hash
- indexed_at

analyses
- id
- project_id
- analyzer
- analyzed_git_head
- analyzed_at
- summary
- purpose
- architecture
- maturity
- strengths
- weaknesses
- reusable_components
- notes
- raw_json

features
- id
- analysis_id
- name
- description
- confidence

technologies
- id
- name
- category

project_technologies
- project_id
- technology_id

relationships
- id
- source_project
- target_project
- type
- description
- confidence

configuration
- key
- value

dependencies
- id (INTEGER, PK)
- project_id (FK)
- name
- manager (npm, go_mod, pip, cargo, bundler, maven, gradle)
- created_at
- UNIQUE(project_id, name, manager)

Design Principles

- Store facts.
- Compute indicators.
- Never duplicate deterministic metadata.
- Analyses are versionable.

==================================================================
2. MCP TOOL SPECIFICATION
==================================================================

Discovery

- health()
- discoverProjects()
- listProjects()
- getProject(id)

Search

- searchProjects(query)
- searchDocumentation(query)

Analysis

- getAnalysis(projectId)
- storeAnalysis(projectId, analysis)
- listProjectsNeedingAnalysis()

Configuration

- getConfiguration()
- updateConfiguration()

Relationships

- listRelationships(projectId)

Rules

- Small composable tools
- Stateless where possible
- Deterministic outputs

==================================================================
3. HTTP API
==================================================================

GET /health

GET /projects
GET /projects/{id}

GET /projects/{id}/documents
GET /projects/{id}/analysis

GET /search?q=

GET /relationships
GET /relationships/{projectId}

GET /statistics

GET /configuration

PATCH /configuration

Dashboard consumes only HTTP.

Agents consume MCP.

==================================================================
4. AI AGENT SPECIFICATION
==================================================================

Responsibilities

- Discover projects
- Detect changes
- Decide when analysis is needed
- Investigate repositories
- Produce semantic knowledge
- Persist analyses

Recommended workflow

User
↓

Portfolio question
↓

health()

↓

discoverProjects()

↓

Search metadata

↓

If semantic knowledge missing or outdated:
    Investigate repository
    Produce analysis
    storeAnalysis()

↓

Answer user

Never edit repositories.

Prefer existing knowledge before re-analysis.

==================================================================
7. DASHBOARD SPECIFICATION
==================================================================

Read-only.

Pages

1. Portfolio Overview
- counts
- activity
- technologies

2. Project List
- search
- filters
- sorting

3. Project Detail
- metadata
- documentation
- analysis
- relationships

4. Relationship Explorer

5. Statistics

Never invokes AI.

Never modifies knowledge.

==================================================================
8. SYNCHRONIZATION WITH KNOWLEDGEMODEL
==================================================================

KnowledgeModel.md defines concepts.

This document defines implementation.

Mapping

Project -> projects
Metadata -> metadata
Dependency -> dependencies
Documentation -> documents
Analysis -> analyses
Feature -> features
Technology -> technologies
Relationship -> relationships

Every API and MCP tool exchanges these canonical entities.

==================================================================
6. AGENT INTEGRATION FRAMEWORK
==================================================================

Integration Interface

Every agent integration implements the Integration interface:

```go
type Integration interface {
    Name() string
    AgentType() string
    Install(ctx, opts) (*InstallResult, error)
    Validate(ctx) (*ValidationResult, error)
    Upgrade(ctx, opts) (*UpgradeResult, error)
    Remove(ctx) error
}
```

Integration Responsibilities

- Register MCP server with agent using official CLI commands only
- Install agent-specific skills/prompts
- Validate installation (MCP reachable, tools available, configs present)
- Upgrade integration (re-register MCP, update skills)
- Remove integration (unregister MCP, remove skills, clean metadata)

Official Methods Requirement

Per ADR-016, all integrations MUST use official tool methods for MCP server registration:

| Tool | Official Method | Status |
|------|---------------|--------|
| Claude Code | `claude mcp add/remove/get` | ✅ Required |
| OpenCode | `opencode mcp add --url` (remote only) | ⚠️ Partial |
| Cline | No CLI method | ❌ Manual setup only |

Direct config file manipulation is FORBIDDEN in production code.

For tools without official CLI support:
- Document manual setup in docs/integration-guideline.md
- Provide transparent unsafe-*.sh scripts with explicit warnings
- Never embed unofficial config editing in integration code

CLI Commands

- `portfolio install <integration>` — Install integration
- `portfolio validate <integration>` — Validate installation  
- `portfolio upgrade <integration>` — Upgrade to new version
- `portfolio uninstall <integration>` — Remove integration

Integration Storage

- Integration metadata stored in database (not on filesystem)
- Tracks: name, version, agent_type, status, installed_at, last_validated_at
- Manager handles backup/rollback; integration handles agent-specific operations

Validation Checks

Standard checks every integration must provide:

1. Agent installed (binary in PATH)
2. Integration metadata exists
3. Config file accessible
4. MCP server entry registered (via official CLI)
5. MCP server reachable (health check)
6. Tools available (test calls)
7. Skill file present (if applicable)

==================================================================
9. IMPLEMENTATION ORDER
==================================================================

1. Database
2. Discovery
3. Metadata extraction
4. Documentation indexing
5. Search
6. HTTP API
7. MCP
8. Agent integration
9. Dashboard
10. Portfolio intelligence

==================================================================
Guiding Principle

The Engine owns deterministic knowledge.

AI Agents own semantic understanding.

The Dashboard visualizes.

The CLI bootstraps.

Everything is built around the Knowledge Model.
