# Portfolio Context Pack for Epic 1 - Project Foundation

**Generated:** 2026-07-22
**Purpose:** Comprehensive context for all subagents working on Epic 1 implementation
**Version:** Based on specification v1.0

---

## Project Overview

### Vision Statement
Portfolio is a local-first project inventory and knowledge platform that enables developers and AI coding agents to understand an entire software portfolio. It is **not** a project management tool — it is a portfolio awareness system.

### Core Problem Solved
Developers accumulate dozens or hundreds of repositories and lose visibility into:
- What they have built
- What each project does
- Which projects are active
- What can be reused
- How projects relate to one another

### Product Principles
- **Local-first** — All data remains on user's machine
- **Deterministic engine** — Repeatable operations only
- **AI-native integrations** — Primary interface through AI coding agents
- **Zero ongoing maintenance** — Install → Initialize → Forget
- **Read-only dashboard** — Visualization only, no modifications
- **Existing folder structure respected** — Works with user's current setup

### User Journey
```text
Install → Initialize → Forget
```

After initialization:
- **AI coding agent** = primary interface
- **Dashboard** = primary exploration interface  
- **CLI** = administrative tasks only (init, doctor, upgrades)

---

## Technical Architecture

### System Philosophy
> **The Engine knows. The AI Agent thinks.**

The engine is deterministic. AI agents provide reasoning. The dashboard visualizes knowledge.

### Core Components

**Portfolio Engine (Go)**
- Responsibilities: Project discovery, metadata extraction, documentation indexing, knowledge storage, search, MCP server, HTTP API
- Key Principle: Never performs AI reasoning

**AI Coding Agent**
- Responsibilities: Understanding user intent, investigating repositories, producing semantic knowledge, comparing projects, maintaining analyses
- Key Principle: Never owns project data

**Agent Integration**
- Responsibilities: MCP registration, agent instructions/skills, installation, validation, upgrades
- First integration: Claude Code
- Future: Codex CLI, OpenCode, Cursor, other MCP-compatible agents

**Dashboard (Read-only)**
- Capabilities: Browse projects, search, filter, view metadata/analyses/relationships, statistics
- Key Principle: Never invokes AI, never modifies repositories

**CLI (Administrative)**
- Commands: init, config, doctor, upgrade, integration
- Key Principle: Not intended for day-to-day interaction

### Knowledge Flow
```text
Repository → Discovery → Deterministic Metadata → Knowledge Store → AI Analysis (optional) → Semantic Knowledge → Dashboard/AI Responses
```

Projects become useful immediately after discovery and richer after AI analysis.

### Interfaces
Both HTTP API and MCP operate on the same service layer and knowledge model:
```text
Portfolio Engine → HTTP API (Dashboard) & MCP Server (AI Agents)
```

---

## Technology Stack

### Core Technologies

**Portfolio Engine: Go**
- Rationale: Systems-level work (filesystem, git metadata, database) benefits from Go's standard library, static compilation, cross-platform support

**Local Knowledge Store: SQLite**
- Rationale: Local-first architecture requires embedded persistence with zero operational overhead
- Design: Store facts, compute indicators, never duplicate deterministic metadata, versionable analyses

**AI Agent Integration: MCP (Model Context Protocol)**
- Rationale: Standardized interface for AI coding agents
- Design: Small, composable tools, stateless where possible, deterministic outputs

**Interfaces:**
- RESTful HTTP API for dashboard
- Git for repository discovery and metadata extraction

### Dependencies
- Minimize dependencies (per engineering guidelines)
- Dependency management policies defined during Epic 1 implementation

---

## Engineering Principles

### Core Principles

**1. Engine Knows, Agent Thinks**
- Engine: Discover repositories, extract metadata, index documentation, store/retrieve knowledge, search
- AI agents: Summaries, feature extraction, architecture understanding, project comparisons, relationship discovery
- **Never move semantic reasoning into the engine**

**2. Deterministic by Default**
- Every engine operation produces repeatable results for the same input
- Avoid heuristics that require an LLM

**3. Store Facts, Compute Indicators**
- **Store immutable facts:** git HEAD, documentation hash, timestamps
- **Compute indicators:** Needs analysis, analysis outdated, recently modified
- **Do not store derived state unless necessary**

**4. Local First**
- Repositories remain on user's machine
- Knowledge remains local
- Cloud services are optional extensions, not requirements

**5. Capabilities over Workflows**
- Expose small, composable capabilities: discoverProjects, listProjects, searchProjects, storeAnalysis
- Engine should not implement high-level workflows; AI agents compose them

**6. AI is Optional**
- Portfolio provides value immediately after deterministic discovery
- AI enriches portfolio but is not required for core functionality

**7. Dashboard is Read-only**
- Dashboard visualizes knowledge only
- Must not: invoke AI, modify repositories, perform analysis

**8. CLI is Administrative**
- CLI exists for: initialization, upgrades, diagnostics, integration management
- Users interact primarily through AI coding agent

**9. Agent Agnostic**
- Engine must never depend on specific AI assistant
- Agent-specific behavior belongs in installable integrations

**10. Single Knowledge Model**
- Every interface operates on same canonical model: Database, MCP, HTTP API, Dashboard, AI agents
- No interface should invent its own representation

### Coding Guidelines
- Prefer composition over inheritance
- Keep packages cohesive
- Avoid global state
- Design interfaces around capabilities
- Keep public APIs stable
- Minimize dependencies
- Write tests for deterministic logic

---

## Knowledge Model Summary

### Core Entities

**Project**
- Identity: id (UUID), name, root_path, repository_type, discovered_at
- Metadata (Engine): git information, languages, frameworks, dependencies, documentation, hashes, timestamps
- Analysis (Agent): Optional semantic understanding

**Documentation**
- Engine-extracted: README.md, docs/*, ADRs, CHANGELOG, DESIGN, ARCHITECTURE
- Stored as searchable documents without interpretation

**Analysis** (AI Agent only)
- Fields: summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes
- Metadata: analyzed_at, analyzed_git_head, analyzer

**Feature**
- Capability implemented by project (Authentication, JWT, Payments, REST API, etc.)
- Belongs to analyses, can become shared entities

**Technology**
- Normalized reference (Go, Gin, PostgreSQL, React, Docker)
- Used for filtering and relationships

**Relationship**
- Connection between projects (Similar, Evolution, Shared Feature, Shared Technology, Reuses Component)
- Primarily generated by AI agents

### Deterministic vs Semantic

**Engine Stores:**
- filesystem facts, git facts, documentation, metadata, technologies, timestamps

**Agent Stores:**
- summaries, purpose, features, architecture, recommendations, relationships

### Derived Indicators
**Store:** git_head, last_scan, last_analysis, documentation_hash
**Compute:** analysis_available, needs_analysis, analysis_outdated, documentation_changed

---

## Platform Specification Summary

### Database Schema (SQLite)

**Core Tables:**
- projects: id, name, root_path, repository_type, discovered_at, updated_at
- metadata: project_id, git_head, default_branch, last_commit_at, language_summary, framework_summary, dependency_summary, documentation_hash, last_scan_at
- documents: id, project_id, path, kind (README, ADR, DOC, CHANGELOG), content, content_hash, indexed_at
- analyses: id, project_id, analyzer, analyzed_git_head, analyzed_at, summary, purpose, architecture, notes, raw_json
- features: id, analysis_id, name, description, confidence
- technologies: id, name, category
- project_technologies: project_id, technology_id
- relationships: id, source_project, target_project, type, description, confidence
- configuration: key, value

**Design Principles:**
- Store facts, compute indicators
- Never duplicate deterministic metadata
- Analyses are versionable

### MCP Tools Specification

**Discovery:**
- health(), discoverProjects(), listProjects(), getProject(id)

**Search:**
- searchProjects(query), searchDocumentation(query)

**Analysis:**
- getAnalysis(projectId), storeAnalysis(projectId, analysis), listProjectsNeedingAnalysis()

**Configuration:**
- getConfiguration(), updateConfiguration()

**Relationships:**
- listRelationships(projectId)

### HTTP API
- GET /health, /projects, /projects/{id}, /projects/{id}/documents, /projects/{id}/analysis
- GET /search?q=, /relationships/{projectId}, /statistics, /configuration
- PATCH /configuration

### AI Agent Workflow
```text
User → Portfolio question → health() → discoverProjects() → Search metadata → 
If semantic knowledge missing: Investigate repository → Produce analysis → storeAnalysis() → Answer user
```

### Dashboard Specification
- **Pages:** Portfolio Overview, Project List, Project Detail, Relationship Explorer, Statistics
- **Constraints:** Never invokes AI, never modifies knowledge

### Implementation Order
1. Database → 2. Discovery → 3. Metadata extraction → 4. Documentation indexing → 5. Search → 6. HTTP API → 7. MCP → 8. Agent integration → 9. Dashboard → 10. Portfolio intelligence

---

## Epic 1 - Project Foundation Requirements

### Epic Overview
**Milestone:** 1 — Core Engine  
**Status:** todo  
**Total Size:** 1M + 2S = ~8 days

### Story Breakdown

**Story 1.1: Bootstrap Go Project (S)**
- Status: todo, Blocked by: None
- Acceptance Criteria:
  - Go module initialized with appropriate name
  - Standard project structure: `cmd/`, `internal/`, `pkg/`
  - `.gitignore` configured for Go
  - LICENSE file present
  - README with build and run instructions

**Story 1.2: Configuration System (M)**
- Status: todo, Blocked by: 1.1
- Acceptance Criteria:
  - Configuration file format defined (TOML per Go conventions)
  - Configuration schema: project roots, ignored paths, database path
  - Configuration loading with defaults
  - Configuration validation on startup
  - Error handling for missing or invalid config
- Technical Context:
  - Config location: `~/.portfolio/config.toml`
  - Stores list of directories to scan for projects

**Story 1.3: Logging Framework (S)**
- Status: todo, Blocked by: 1.1
- Acceptance Criteria:
  - Structured logging implementation (e.g., zap, zerolog)
  - Log levels: DEBUG, INFO, WARN, ERROR
  - Log output to stdout
  - Configurable log level via environment variable

**Story 1.4: CLI Framework (M)**
- Status: todo, Blocked by: 1.1, 1.3
- Acceptance Criteria:
  - CLI framework (e.g., cobra)
  - Subcommands: `init`, `status`, `doctor`
  - `init` prompts for project roots and creates config
  - `status` shows engine health and project count
  - `doctor` runs diagnostics (config check, database access)
- Constraints: CLI is for administration only

**Story 1.5: SQLite Initialization (M)**
- Status: todo, Blocked by: 1.2
- Acceptance Criteria:
  - Database file created at configured path
  - Connection management with proper closing
  - Basic schema validation on startup
  - Migration system in place
  - Database creation with all tables from PlatformSpecification.md
- Technical Context:
  - Tables: projects, metadata, documents, analyses, features, technologies, project_technologies, relationships, configuration

### Dependency Graph
```
Story 1.1 (Bootstrap) → 1.2 (Config) → 1.5 (SQLite)
Story 1.1 (Bootstrap) → 1.3 (Logging) → 1.4 (CLI)
```

**Can Start:** Story 1.1 (no dependencies)

---

## Architecture Decision Records

### ADR-013: Agent Integrations are First-Class Components
**Status:** Accepted
Agent-specific behavior implemented as installable integrations rather than embedded in engine. Integration responsibilities: MCP registration, agent-specific skills, validation, upgrades, removal. Engine remains agent-agnostic.

### ADR-014: Install → Initialize → Forget
**Status:** Accepted
Portfolio follows simple lifecycle: Install → Initialize → Forget. CLI exists primarily for initialization, diagnostics, upgrades, integration management. Primary interaction through AI coding agent.

### ADR-015: KnowledgeModel is the Canonical Source of Truth
**Status:** Accepted
`KnowledgeModel.md` = canonical domain model definition. `PlatformSpecification.md` = implementation contracts. All implementations derive from these documents.

---

## UX Guidelines Summary

### Core Principle
**Dashboard is Read-Only** — visualizes knowledge only, must NOT invoke AI, modify repositories, perform analysis, or modify knowledge.

### Dashboard Pages
1. **Portfolio Overview:** Counts, activity, technologies
2. **Project List:** Search, filters, sorting
3. **Project Detail:** Metadata, documentation, analysis, relationships
4. **Relationship Explorer:** Visual representation, navigation
5. **Statistics:** Technology distribution, maturity, timelines

### Design Principles
- **Knowledge Visualization:** Quick scanning, pattern recognition, relationship discovery
- **Progressive Enhancement:** Base layer (deterministic metadata), enhancement layer (semantic analysis)
- **Single Knowledge Model:** All views operate on canonical KnowledgeModel.md entities

### User Experience Patterns
- **Search-First Navigation:** Fast, rich local search
- **Non-Destructive Interaction:** Links navigate, filters refine, sorting reorders
- **Consistent Entity Views:** Same structure for every project

---

## Development Guidelines

### Project Status
This is a **specification repository** — no source code exists yet. Repository contains planning documents defining product vision, architecture, and implementation roadmap.

### Documentation Hierarchy
1. **KnowledgeModel.md** — Canonical domain model
2. **PlatformSpecification.md** — Implementation contracts
3. **PRD.md** — Product requirements
4. **Guideline.md** — Engineering principles
5. **Tasks** — Implementation roadmap

### When Model Changes
1. Update KnowledgeModel.md
2. Update PlatformSpecification.md
3. Update implementation

### Testing Philosophy
- Tests required for deterministic logic
- Semantic understanding by AI agents is not tested as deterministic
- Focus on engine functionality, not AI reasoning

### Development Workflow
Repository supports devflow development pipeline for systematic feature implementation:
- `devflow:documentation-readiness` — Checks documentation completeness
- `devflow:devflow` — Executes full requirements→merge pipeline

---

## Go Project Structure Requirements

Based on Epic 1.1 requirements and Go best practices:

### Standard Structure
```
portfolio-tool/
├── cmd/                    # Executable applications
│   └── portfolio/         # Main CLI application
├── internal/              # Private application code
│   ├── config/            # Configuration system
│   ├── logging/           # Logging framework
│   ├── database/          # SQLite initialization
│   └── cli/               # CLI commands
├── pkg/                   # Public library code
├── docs/                  # Documentation (already exists)
├── .gitignore            # Go-specific ignores
├── LICENSE               # License file
├── README.md             # Build/run instructions
└── go.mod                # Go module definition
```

### Module Naming Convention
- Module name should reflect import path
- Consider using `github.com/` or generic `portfolio-tool` format

### Configuration System
- Format: TOML (per Go conventions)
- Location: `~/.portfolio/config.toml`
- Schema: project roots, ignored paths, database path
- Must support defaults, validation, error handling

### Logging Requirements
- Structured logging (zap or zerolog recommended)
- Levels: DEBUG, INFO, WARN, ERROR
- Output: stdout
- Configuration: Environment variable for log level

### CLI Framework
- Framework: cobra recommended
- Commands: init, status, doctor
- Administrative only (not primary interface)

### Database Requirements
- SQLite with proper connection management
- Migration system from day one
- All tables from PlatformSpecification.md
- Schema validation on startup

---

## Key Technical Constraints

### Database Constraints
- Use SQLite (local-first requirement)
- Store facts, compute indicators
- Never duplicate deterministic metadata
- Analyses must be versionable

### API Constraints
- Both HTTP and MCP operate on same service layer
- Single knowledge model across all interfaces
- HTTP for dashboard, MCP for AI agents

### Logging Constraints
- Structured logging only
- Configurable via environment variable
- Output to stdout

### Configuration Constraints
- TOML format (Go convention)
- Must support defaults
- Validation on startup
- Proper error handling

### CLI Constraints
- Administrative tasks only
- Not primary user interface
- Should follow Go CLI best practices

---

## Success Criteria for Epic 1

A successful Epic 1 implementation enables:
1. Properly structured Go project following standard conventions
2. Configuration system that can be loaded, validated, and used
3. Structured logging that provides visibility into engine behavior
4. CLI interface for initialization and administration
5. SQLite database with proper schema and migration system

### Definition of Done
1. All acceptance criteria satisfied
2. Tests added for deterministic logic (config loading, schema validation, migration)
3. Documentation updated if domain model/APIs change
4. Architectural guidelines from Guideline.md respected
5. No regressions in existing functionality

---

## Related Documentation References

### Authoritative Documents
| Document | Purpose | Location |
|----------|---------|----------|
| KnowledgeModel.md | Canonical domain model | /Users/nerddevsltd/Projects/portfolio-tool/docs/KnowledgeModel.md |
| PlatformSpecification.md | Implementation contracts | /Users/nerddevsltd/Projects/portfolio-tool/docs/PlatformSpecification.md |
| PRD.md | Product requirements | /Users/nerddevsltd/Projects/portfolio-tool/docs/PRD.md |
| Guideline.md | Engineering principles | /Users/nerddevsltd/Projects/portfolio-tool/docs/Guideline.md |
| TechStack.md | Technology stack | /Users/nerddevsltd/Projects/portfolio-tool/docs/TechStack.md |
| UXGuidelines.md | Dashboard UX patterns | /Users/nerddevsltd/Projects/portfolio-tool/docs/UXGuidelines.md |
| ADR.md | Architecture decisions | /Users/nerddevsltd/Projects/portfolio-tool/docs/ADR.md |

### Task Documentation
| Document | Purpose | Location |
|----------|---------|----------|
| tasks/index.md | Implementation roadmap | /Users/nerddevsltd/Projects/portfolio-tool/docs/tasks/index.md |
| tasks/epic-01-project-foundation.md | Epic 1 detailed stories | /Users/nerddevsltd/Projects/portfolio-tool/docs/tasks/epic-01-project-foundation.md |

### Quick Reference Commands
- For detailed technical specs: Reference PlatformSpecification.md and KnowledgeModel.md
- For engineering principles: Reference Guideline.md
- For Epic-specific requirements: Reference tasks/epic-01-project-foundation.md
- For technical decisions: Reference ADR.md

---

## Notes for Subagents

### This Context Pack
- Contains condensed summaries from all authoritative documentation
- Focuses on Epic 1 - Project Foundation implementation
- References original documents for detailed specifications
- Single source of truth for all Epic 1 development work

### When Working on Epic 1
1. Always reference this context pack first
2. Consult original documents for detailed specifications
3. Follow engineering principles from Guideline.md
4. Ensure all acceptance criteria are met
5. Write tests for deterministic logic
6. Update documentation if domain model changes

### Key Reminders
- **Engine knows, Agent thinks** — Never move semantic reasoning into engine
- **Local-first** — All data stays on user's machine  
- **Deterministic by default** — All operations repeatable
- **Store facts, compute indicators** — Don't store derived state
- **AI is optional** — Core functionality works without AI
- **Dashboard is read-only** — Never invoke AI or modify data
- **Single knowledge model** — All interfaces use same entities

---

**End of Context Pack for Epic 1 - Project Foundation**