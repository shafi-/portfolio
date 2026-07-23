# Context Pack: Epic 3 — Metadata Extraction

_Gathered 2026-07-22. Token-optimized for agent consumption._

---

## Project Identity

- **Product:** Portfolio — local-first project inventory & knowledge platform
- **Principle:** "Engine Knows, Agent Thinks" — engine is deterministic only
- **Lifecycle:** Install → Initialize → Forget
- **Primary interface:** AI coding agent (MCP). Dashboard = exploration. CLI = admin.

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Engine | Go |
| Store | SQLite |
| Agent protocol | MCP |
| API | HTTP/REST |

---

## Epic 3 Scope

**Milestone 1 — Core Engine.** 6 stories, total ~18 days.

### Stories & Data Flow

```
Story 3.1 (Git) ──► 3.2 (Languages) ──┬──► 3.3 (Frameworks) ──► 3.4 (Dependencies)
                                      └──► 3.6 (Doc Hashes) ──► blocked by Epic 4
```

**Can start when:** Blockers 2.2 (Discovery inserts projects), 1.5 (DB schema) done.

### Story Details

| Story | Size | Key Output | Edge Cases |
|-------|------|-----------|------------|
| **3.1** Git Metadata | L | `default_branch`, `git_head`, `last_commit_at`, `last_modified_at`, `commit_count` | Empty repo, bare repo, detached HEAD |
| **3.2** Detect Languages | M | `language_summary` (e.g. "Go, TypeScript") | Polyglot; ignore vendor/node_modules/generated; top 10 built-in extensions, configurable |
| **3.3** Detect Frameworks | M | `framework_summary` | Multi-framework; extensible mapping |
| **3.4** Detect Dependencies | M | `dependency_summary` (top 10) | Per-language: package.json, go.mod, requirements.txt/pyproject.toml, Cargo.toml, Gemfile, pom.xml/build.gradle |
| **3.6** Doc Hashes | M | `documentation_hash` (SHA-256) | Blocked on Epic 4 doc discovery |

---

## Relevant Schema (PlatformSpecification.md + Architecture)

```sql
metadata (
  project_id        FK,
  git_head,
  default_branch,
  last_commit_at,
  last_modified_at,
  commit_count      INTEGER,
  language_summary,
  framework_summary,
  dependency_summary,
  documentation_hash,
  last_scan_at
)

dependencies (
  project_id  FK,
  name        TEXT,
  manager     TEXT,   -- "npm", "go_mod", "pip", "cargo", etc.
  UNIQUE(project_id, name, manager)
)
```

---

## Principles to Enforce

1. **Deterministic** — same input = same output. No LLM/heuristics.
2. **Store facts, compute indicators** — store `documentation_hash`; compute `documentation_changed`.
3. **Capabilities over workflows** — expose small tools, not orchestrated pipelines.
4. **Never duplicate** — metadata in one place, no synthetic fields.
5. **Ignore generated** — vendor/, node_modules/, build/, .git/ always excluded.

---

## Output Fields Reference

| Field | Stored In | Source Story | Type |
|-------|-----------|-------------|------|
| `default_branch` | metadata | 3.1 | string |
| `git_head` | metadata | 3.1 | string (SHA) |
| `last_commit_at` | metadata | 3.1 | timestamp |
| `last_modified_at` | metadata | 3.1 | timestamp |
| `commit_count` | metadata | 3.1 | integer |
| `language_summary` | metadata | 3.2 | comma-separated string |
| `framework_summary` | metadata | 3.3 | comma-separated string |
| `dependency_summary` | metadata | 3.4 | string (top 10 names) |
| `documentation_hash` | metadata | 3.6 | SHA-256 hex string |

---

## Composition with Prior Epics

- **Epic 1** — provides DB schema (`projects`, `metadata` tables) and CLI skeleton
- **Epic 2** — `discoverProjects()` inserts project rows; **Epic 3 fills in metadata** for those rows
- **Epic 4** — doc discovery (parallel track); Story 3.6 hashes its output
- **Epic 5** — search indexes what Epic 3 produces

---

## ADRs in Effect

- **ADR-013:** Agent integrations = first-class components (engine stays agnostic)
- **ADR-014:** Install → Initialize → Forget (after init, engine runs passively)
- **ADR-015:** KnowledgeModel.md = canonical source (schema derives from it)
