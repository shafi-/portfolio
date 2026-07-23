# Epic 2 — Discovery: Context Pack

## Project Overview

Portfolio is a **local-first project inventory and knowledge platform** that enables developers and AI coding agents to understand an entire software portfolio. It is infrastructure, not project management.

**Lifecycle:** Install → Initialize → Forget  
**Primary interface:** AI coding agent (via MCP)  
**Exploration interface:** Read-only dashboard (via HTTP API)  
**Admin interface:** CLI

**Key architectural split:** Engine does deterministic work; AI agents do semantic reasoning.

---

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Engine language | Go | Filesystem ops, git metadata, DB, deterministic logic |
| Knowledge store | SQLite | Local persistence, serverless, embedded |
| Agent protocol | MCP (Model Context Protocol) | AI agent ↔ Engine interface |
| Dashboard API | HTTP/REST | Read-only dashboard communication |
| VCS | Git | Repository discovery & metadata |

---

## Domain Model (Relevant to Epic 2)

Key entities from KnowledgeModel.md that this epic touches:

**Project** (engine-owned, deterministic)
- id (UUID), name, root_path, repository_type, discovered_at, updated_at
- Metadata: git_head, default_branch, last_commit_at, language_summary, framework_summary, dependency_summary, documentation_hash, last_scan_at

**Technology** (normalized ref, e.g. "Go", "React", "Docker") — used for filtering/relationships

**Derived indicators** (compute, don't store): needs_analysis, analysis_outdated, documentation_changed

**Key rule:** Store facts (git_head, last_scan, timestamps). Compute indicators when needed.

---

## Epic Summary

Epic 2 implements **project discovery** — scanning configured root directories to find all repositories, detect project types, and persist Project records to the knowledge store.

**Milestone:** 1 — Core Engine | **Status:** todo | **Total estimate:** ~17 days

Stories are blocked by Epic 1 completion (stories 1.2 config, 1.4 DB schema, 1.5 file operations).

---

## Stories & Acceptance Criteria

### 2.1 Configure Project Roots (S, blocked by 1.2, 1.4)
- `portfolio init` prompts for project root directories (initial and reconfiguration)
- `portfolio init` is idempotent — re-running reconfigures roots
- Configuration persists to config file
- Support for multiple root directories
- Validation that paths exist and are accessible
- CLI: `portfolio projects list`, `portfolio projects get <id>`, `portfolio discover`

### 2.2 Recursive Project Discovery (L, blocked by 2.1, 1.5)
- Walks directory trees from configured roots
- Detects Git repos by `.git` directory presence
- Creates or updates Project records (keyed by root_path): id, name, root_path, repository_type, discovered_at
- Handles permission errors gracefully
- Reports discovery count and errors
- Guarded by mutex — concurrent calls return error

### 2.3 Support Nested Folders (M, blocked by 2.2)
- Continues recursion when subdirectory contains a repo
- Creates separate Project record per discovered repo
- Handles monorepo structures with nested services
- No depth limit (within filesystem constraints)

### 2.4 Detect Common Project Types (M, blocked by 2.2)
- Detects markers: `package.json` (Node), `go.mod` (Go), `requirements.txt`/`pyproject.toml` (Python), `Cargo.toml` (Rust), `pom.xml` (Java)
- Sets `repository_type` field based on detected markers
- Supports multiple markers per project (polyglot)
- Gracefully handles unknown project types

### 2.5 Ignore Generated Directories (S, blocked by 2.2)
- Skips: `node_modules/`, `vendor/`, `.venv/`, `target/`, `build/`, `dist/`
- Respects `.gitignore` for additional ignores
- Configurable ignore patterns in config file
- Logs skipped directories at DEBUG level

---

## Key Principles to Follow

From Guideline.md — critical for Epic 2 implementation:

1. **Engine Knows, Agent Thinks** — Never move semantic reasoning into the engine. Discovery is purely deterministic (filesystem facts).
2. **Deterministic by Default** — Same input → same output. Avoid heuristics that require LLM.
3. **Store Facts, Compute Indicators** — Persist paths, hashes, timestamps. Compute `needs_analysis` when queried.
4. **Local First** — Everything stays on the user's machine. No cloud calls during discovery.
5. **Capabilities over Workflows** — Expose `discoverProjects`, `listProjects`, `getProject`. Don't build high-level workflows in the engine.
6. **AI is Optional** — Discovered projects are useful immediately (even before any AI analysis).
7. **Single Knowledge Model** — Every interface (DB, MCP, HTTP) uses the same canonical entities from KnowledgeModel.md.
8. **Agent Agnostic** — Engine must not depend on a specific AI assistant.
9. **Respect Existing Project Structures** — Don't modify repos during discovery. Read-only.

**Coding:** Prefer composition, minimize deps, avoid global state, write tests for deterministic logic.

---

## Relevant Dependencies

- **Epic 1 (Project Foundation):** Config management (1.2), DB schema / migrations (1.4), filesystem utilities (1.5) — all required before Epic 2 can start
- **Downstream consumers of this epic:** Epic 3 (Metadata Extraction) reads the Project records created here; Epic 4 (Documentation Indexing) needs discovered paths
- **Database:** Must implement `projects` table per PlatformSpecification.md schema before Story 2.2
- **MCP tools** defined in PlatformSpecification.md that this epic enables: `discoverProjects()`, `listProjects()`, `getProject(id)`, `health()`
