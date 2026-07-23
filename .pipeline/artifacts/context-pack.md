# Context Pack: Epic 4 — Documentation Indexing

## Project Overview

Portfolio is a local-first project inventory and knowledge platform. Enables developers and AI coding agents to understand an entire software portfolio. Not project management — portfolio awareness.

**Lifecycle:** Install → Initialize → Forget. AI agent is primary interface, dashboard is exploration, CLI is admin.

## Architecture

```
AI Agent <-> MCP/HTTP <-> Portfolio Engine <-> SQLite <-> Local Repos
```

- **Engine** owns deterministic ops (discovery, metadata, indexing, search, storage)
- **Agent** owns semantic reasoning (summaries, analysis, relationships)
- **Dashboard** is read-only visualization
- **CLI** is admin only (init, config, doctor, upgrade)

## Tech Stack

| Layer | Choice |
|-------|--------|
| Engine | **Go** (systems work, stdlib, static binary) |
| Store | **SQLite** (local-first, embedded, zero ops) |
| Agent Protocol | **MCP** (composable tools, stateless, deterministic) |
| Dashboard API | HTTP/REST |

## Engineering Principles (Guideline.md)

1. **Engine Knows, Agent Thinks** — never move semantic reasoning into engine
2. **Deterministic by Default** — same input → same output
3. **Store Facts, Compute Indicators** — persist git HEAD/docs hash/timestamps; compute needs_analysis/outdated on the fly
4. **Local First** — repos on user machine, knowledge local, cloud optional
5. **Capabilities over Workflows** — expose small composable tools (discoverProjects, searchDocumentation), agents compose
6. **AI is Optional** — value after deterministic discovery alone
7. **Dashboard is Read-only** — no AI invocation, no modification
8. **Agent Agnostic** — engine never depends on specific AI assistant
9. **Single Knowledge Model** — all interfaces use same canonical model

## Knowledge Model (relevant entities)

**Project** — discovered repo; identity (UUID, name, root_path), metadata (git, langs, frameworks), analysis (agent, optional)

**Documentation** — engine-extracted documents stored searchable without interpretation:
- README.md, docs/*, ADRs, CHANGELOG, DESIGN, ARCHITECTURE
- Fields: id, project_id, path, kind (README/ADR/DOC/CHANGELOG), content, content_hash, indexed_at

**Technology** — normalized refs (Go, React, etc.) for filtering/relationships

## Platform Spec (relevant contracts)

### DB Schema

**documents** table:
- id (PK), project_id (FK→projects), path, kind, content, content_hash, indexed_at

### MCP Tools

Search:
- `searchProjects(query)` — search project metadata
- `searchDocumentation(query)` — search indexed doc content (Epic 4 enables this)

### HTTP API

- `GET /projects/{id}/documents` — list indexed docs for project
- `GET /search?q=` — cross-portfolio search

### Implementation Order

Epic 4 = step 4 in the sequence: DB → Discovery → Metadata → **Docs Indexing** → Search → HTTP API → MCP → Agent → Dashboard → Intelligence

## ADR-013: Agent Integrations are First-Class

Agent-specific behavior (MCP registration, skills) goes in installable integrations, not engine. Engine stays agent-agnostic.

## ADR-014: Install → Initialize → Forget

CLI is init/diag/upgrade only. After init, AI agent drives interaction.

## ADR-015: KnowledgeModel is Canonical Source of Truth

KnowledgeModel.md defines concepts. PlatformSpecification.md defines implementation. All derive from these.

---

# Epic 4: Documentation Indexing

**Milestone:** 1 — Core Engine
**Status:** todo
**Total Size:** 1L + 2M + 2S (~15 days)
**Can Start:** After Epic 2.2 (Metadata Extraction) and Epic 1.5 (DB layer) complete

## Stories

### 4.1 Index README (M, blocked by 2.2, 1.5)

**What:** Find and store README files in documents table.
**AC:**
- Find README.md, README.rst, README.txt, readme.md (case-insensitive)
- Store: project_id, path, kind="README", content, content_hash, indexed_at
- Missing README = no error
- Handle >1MB files (truncate or stream)

### 4.2 Index docs/ Directory (M, blocked by 4.1)

**What:** Recursively index docs/ files.
**AC:**
- Recursive scan of docs/ directory
- kind="DOC" for docs
- Supported formats: .md, .rst, .txt, .adoc
- Skip binary files
- Handle missing docs/ (no error)

### 4.3 Index ADRs (S, blocked by 4.2)

**What:** Find ADRs in standard locations.
**AC:**
- Check: docs/adr/, .adr/, adr/ directories
- Recognize: NNN-*.md or *.md naming
- kind="ADR"
- Handle missing ADRs (no error)

### 4.4 Index CHANGELOG (S, blocked by 4.1)

**What:** Find changelog/history files.
**AC:**
- Find: CHANGELOG.md, CHANGES.md, HISTORY.md
- kind="CHANGELOG"
- Handle missing (no error)

### 4.5 Full-Text Search Indexing (L, blocked by 4.1-4.4)

**What:** Enable full-text search across all indexed documents.
**AC:**
- SQLite FTS5 index on documents.content
- Support: phrase queries, Boolean operators
- Ranked results with project context
- Fast cross-portfolio search
- Enables searchDocumentation(query) MCP tool

## Key Design Constraints

1. **Deterministic** — same repo → same index (content_hash as dedup key)
2. **Respect existing structure** — don't move/copy files, just index
3. **Store facts** — content_hash + indexed_at; compute staleness from hash vs repo
4. **Graceful** — missing files/dirs = skip, not error; binary skip; large file handling
5. **Single model** — documents table follows the canonical schema in PlatformSpecification.md
