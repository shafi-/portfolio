# TechStack.md

# Portfolio Technology Stack

Version: 0.1 (Draft)

> This document derives from PlatformSpecification.md, Guideline.md, and Tasks.md. When introducing new technologies, first update the canonical specifications.

---

## Core Technologies

### Portfolio Engine

**Language:** Go

The Portfolio Engine performs deterministic operations:
- Project discovery
- Metadata extraction
- Documentation indexing
- Storage and retrieval

**Rationale:** The engine performs systems-level work (filesystem operations, git metadata extraction, database interactions) and benefits from Go's standard library, static compilation, and cross-platform support.

---

### Local Knowledge Store

**Database:** SQLite

SQLite stores the canonical knowledge model:
- Projects and metadata
- Documents and analysis
- Technologies and relationships
- Configuration

**Design Principles:**
- Store facts, compute indicators
- Never duplicate deterministic metadata
- Analyses are versionable

**Rationale:** Local-first architecture requires knowledge to remain on the user's machine. SQLite provides serverless, embedded persistence with zero operational overhead.

---

### AI Agent Integration

**Protocol:** MCP (Model Context Protocol)

MCP exposes composable capabilities:
- Discovery tools
- Search tools
- Analysis storage tools
- Configuration tools

**Design Principles:**
- Small, composable tools
- Stateless where possible
- Deterministic outputs

**Rationale:** MCP enables AI coding agents to interact with Portfolio through a standardized interface.

---

## Interfaces

### HTTP API

RESTful API consumed by the dashboard:
- GET /health
- GET /projects
- GET /projects/{id}
- GET /projects/{id}/documents
- GET /projects/{id}/analysis
- GET /search
- GET /relationships/{projectId}
- GET /statistics
- GET /configuration
- PATCH /configuration

### Dashboard Backend

Asset serving and REST integration for the read-only dashboard.

---

## Technology Categories

### Languages

- **Go** — Portfolio Engine

### Data Storage

- **SQLite** — Local knowledge store

### Protocols

- **MCP** — AI agent integration
- **HTTP/REST** — Dashboard communication
- **Git** — Repository discovery and metadata extraction

### Future Considerations

Per Tasks.md future milestones:
- Additional agent integrations (Codex CLI, OpenCode, Cursor)
- Plugin system
- Cloud sync (optional)

These will be specified when reaching those milestones.

---

## Version Management

When updating technology versions:
1. Update this document
2. Update PlatformSpecification.md
3. Assess impact on existing implementations

---

## Dependencies

The engine must minimize dependencies (per Guideline.md coding guidelines).

Specific dependency management policies will be defined during Epic 1 (Project Foundation) implementation.

---

## Alignment with Principles

This tech stack supports the engineering principles defined in Guideline.md:

- **Engine Knows, Agent Thinks** — Go engine handles deterministic operations only
- **Deterministic by Default** — SQLite ensures repeatable queries
- **Local First** — SQLite keeps knowledge on user's machine
- **Capabilities over Workflows** — MCP exposes small, composable tools
- **Agent Agnostic** — MCP allows any AI agent to integrate
