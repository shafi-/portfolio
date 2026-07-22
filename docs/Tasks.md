# Tasks.md

# Portfolio Implementation Roadmap

Version: 1.0

This roadmap is organized into milestones, epics, and user stories.

---

# Milestone 1 — Core Engine

## Epic 1 — Project Foundation

- Bootstrap Go project
- Configuration system
- Logging
- CLI framework
- SQLite initialization

## Epic 2 — Discovery

Stories:

- Configure project roots
- Recursive project discovery
- Support nested folders
- Detect Git repositories
- Detect common project types
- Ignore generated directories

## Epic 3 — Metadata Extraction

Stories:

- Extract Git metadata
- Detect languages
- Detect frameworks
- Detect dependencies
- Compute project statistics
- Compute documentation hashes

## Epic 4 — Documentation Indexing

Stories:

- Index README
- Index docs/
- Index ADRs
- Index CHANGELOG
- Full-text search indexing

## Epic 5 — Knowledge Store

Stories:

- Database schema
- Repository layer
- Migrations
- Search indexes

## Epic 6 — HTTP API

Stories:

- Health endpoint
- Projects API
- Search API
- Configuration API
- Statistics API

## Epic 7 — MCP Server

Stories:

- MCP server
- Discovery tools
- Search tools
- Analysis storage tools
- Configuration tools

---

# Milestone 2 — Agent Integration

## Epic 8 — Integration Framework

Stories:

- Integration abstraction
- Installation framework
- Validation
- Upgrade mechanism

## Epic 9 — Claude Code Integration

Stories:

- Install MCP
- Install Portfolio skill
- Verify integration
- Update integration

## Epic 10 — AI Analysis

Stories:

- Analysis schema
- Persist analyses
- Detect stale analyses
- Relationship persistence

---

# Milestone 3 — Dashboard

## Epic 11 — Dashboard Backend

Stories:

- Asset serving
- REST integration
- Search endpoints

## Epic 12 — Dashboard Frontend

Stories:

- Portfolio overview
- Project list
- Project details
- Relationship explorer
- Statistics

Dashboard remains read-only.

---

# Milestone 4 — Portfolio Intelligence

## Epic 13 — Relationships

Stories:

- Relationship model
- Relationship queries
- Visualization support

## Epic 14 — Insights

Stories:

- Technology summaries
- Portfolio health
- Reusable implementation discovery
- Activity timeline

---

# Future Milestones

- Codex CLI integration
- OpenCode integration
- Cursor integration
- Plugin system
- Team workspaces
- Portfolio snapshots
- Cloud sync (optional)

---

# Definition of Done

A story is complete when:

- Acceptance criteria are satisfied.
- Tests are added for deterministic logic.
- Documentation is updated where applicable.
- KnowledgeModel.md remains consistent.
- PlatformSpecification.md remains consistent.
- Architectural guidelines are respected.
