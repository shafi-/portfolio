# Epic 2 — Discovery: Requirements Document

**Milestone:** 1 — Core Engine
**Status:** Draft
**Total Estimate:** ~17 days

---

## 1. Feature Overview

Epic 2 implements **project discovery** — the first deterministic operation of the Portfolio Engine. Given configured root directories, the engine recursively scans the filesystem to find all Git repositories, identifies common project types (Node, Go, Python, Rust, Java), and persists Project records to the local knowledge store (SQLite).

Discovery is purely deterministic — no semantic reasoning, no AI calls. It produces filesystem facts (paths, timestamps, marker files) that downstream epics (Metadata Extraction, Documentation Indexing) consume.

**Key principle:** Engine Knows, Agent Thinks. Discovery belongs entirely in the Engine.

**User journey segment:** `Initialize (epic 1) → Discover (epic 2) → ...`

---

## 2. Functional Requirements

### 2.1 Configure Project Roots

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-01 | The system MUST provide a CLI command (`portfolio init`) that prompts the user for project root directories | P0 |
| FR-02 | The system MUST persist project root paths to a configuration file | P0 |
| FR-03 | The system MUST support specifying multiple root directories | P0 |
| FR-04 | The system MUST validate that configured root paths exist and are accessible before persisting | P0 |
| FR-05 | The system MUST allow reconfiguration of project roots via a CLI command | P1 |
| FR-06 | The system MUST read root directories from the configuration file during discovery | P0 |

### 2.2 Recursive Project Discovery

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-07 | The system MUST walk directory trees from each configured root | P0 |
| FR-08 | The system MUST detect Git repositories by checking for the presence of a `.git` subdirectory | P0 |
| FR-09 | For each discovered repository, the system MUST create a Project record with: id (UUID), name (directory name), root_path (absolute path), repository_type (initially empty or "git"), discovered_at (timestamp) | P0 |
| FR-10 | The system MUST handle permission errors gracefully without aborting the full discovery (skip inaccessible paths, continue scanning) | P0 |
| FR-11 | The system MUST report discovery results: total projects found, errors encountered | P0 |
| FR-12 | The system MUST be idempotent — re-running discovery updates existing Project records rather than duplicating them (keyed by root_path) | P0 |

### 2.3 Support Nested Folders

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-13 | When a subdirectory within a discovered repo contains its own `.git` directory, the system MUST create a separate Project record for that nested repo | P0 |
| FR-14 | The system MUST continue descending into subdirectories even when a parent contains a Git repo (handling monorepo/service structures) | P0 |
| FR-15 | The system MUST NOT impose an artificial depth limit on recursion; only filesystem constraints apply | P0 |

### 2.4 Detect Common Project Types

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-16 | The system MUST detect project type markers at the root of each discovered repository | P0 |
| FR-17 | The system MUST recognize these markers: `package.json` (Node), `go.mod` (Go), `requirements.txt` or `pyproject.toml` (Python), `Cargo.toml` (Rust), `pom.xml` (Java) | P0 |
| FR-18 | The system MUST set the `repository_type` field on the Project record based on detected markers (e.g., "node", "go", "python", "rust", "java") | P0 |
| FR-19 | The system MUST support multiple markers per project (polyglot detection — e.g., both `package.json` and `go.mod`) | P1 |
| FR-20 | When no known marker is found, the system MUST leave `repository_type` as "unknown" (not fail) | P0 |

### 2.5 Ignore Generated Directories

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-21 | The system MUST skip the following generated directories during discovery: `node_modules/`, `vendor/`, `.venv/`, `target/`, `build/`, `dist/` | P0 |
| FR-22 | The system MUST respect `.gitignore` patterns within discovered repositories to skip additional directories | P1 |
| FR-23 | The system MUST support configurable ignore patterns in the configuration file | P1 |
| FR-24 | The system MUST log skipped directories at DEBUG level (not visible at default log level) | P0 |

---

## 3. Non-Functional Requirements

| ID | Requirement | Category |
|----|-------------|----------|
| NFR-01 | Discovery MUST NOT modify any repository content (read-only) | Correctness |
| NFR-02 | Discovery MUST be deterministic — same filesystem state produces same results | Reliability |
| NFR-03 | Discovery MUST NOT make network calls or contact cloud services | Privacy/Local-First |
| NFR-04 | The system MUST handle large root directories (100,000+ files) without excessive memory growth | Performance |
| NFR-05 | Discovery MUST respect filesystem boundaries (do not follow symlinks outside configured roots by default) | Safety |
| NFR-06 | The system MUST handle SIGINT/SIGTERM gracefully — return partial results (projects discovered before interruption) alongside cancellation error | Resilience |
| NFR-08 | The system MUST guard discovery with a mutex — concurrent calls return an error immediately | Correctness |
| NFR-07 | All discovery logic MUST be testable without a real filesystem (use interface abstraction for filesystem ops) | Testability |

---

## 4. Edge Cases & Error Handling

### 4.1 Filesystem Errors

| Edge Case | Handling |
|-----------|----------|
| Permission denied on directory | Log warning, skip directory, continue recursion on siblings |
| Broken symlink | Detect and skip, log at DEBUG level |
| Circular symlinks (symlink loop) | Track visited inode/dev pairs per root; skip on revisit |
| Very deep directory tree (>1000 levels) | Let OS/hardware limit apply; avoid artificial cutoff |
| Path contains non-UTF-8 characters | Handle as raw bytes, encode as valid UTF-8 or skip with warning |
| Root directory deleted between config and scan | Validate roots at start of discovery; skip missing roots with error |

### 4.2 Repository Edge Cases

| Edge Case | Handling |
|-----------|----------|
| `.git` is a file (worktree) rather than a directory | Detect file-based `.git` references (read first line for `gitdir:`); treat as valid repo |
| Bare Git repository (no working tree) | Detect `HEAD` file + `objects/` dir without `.git` subdirectory; mark as type "bare-git" |
| Nested repo inside another repo's ignored directory | Check ignore patterns before recursing into nested repos |
| Git repository with no commits yet | Still create a Project record; git metadata fields remain empty |
| Repository with broken `.git` directory | Do not fail; create record with partial metadata; log warning |

### 4.3 Project Type Edge Cases

| Edge Case | Handling |
|-----------|----------|
| Marker file exists but is empty/directory | Treat as present (file existence is the signal, not content) |
| Multiple `package.json` files in subdirectories | Only check repo root for type detection; subdirectory markers ignored |
| Unknown marker file (e.g., `Brewfile`, `Gemfile`) | `repository_type` set to "unknown"; no error |
| Mixed/polyglot project | Set `repository_type` to comma-separated or array of detected types |

### 4.4 Configuration Edge Cases

| Edge Case | Handling |
|-----------|----------|
| Root path is a file, not a directory | Validate and reject with error message |
| No root directories configured | Discovery returns empty set; log warning at WARN level |
| Duplicate root directories | Deduplicate at load time |
| Home directory (`~`) in path | Expand `~` to absolute path before validation |
| Concurrent discovery calls | Mutex guard on Discoverer — second call returns "discovery in progress" error |

---

## 5. Acceptance Criteria

Derived from Epic 2 stories:

### Story 2.1 — Configure Project Roots

| ID | Criterion |
|----|-----------|
| AC-01 | `portfolio init` prompts user to enter project root directories (initial and reconfiguration) |
| AC-02 | Entered paths are persisted to the configuration file |
| AC-03 | Multiple root directories can be configured |
| AC-04 | Non-existent or inaccessible paths are rejected with a clear error |
| AC-05 | Configured roots are readable by the `discoverProjects` operation |

### Story 2.2 — Recursive Project Discovery

| ID | Criterion |
|----|-----------|
| AC-06 | `discoverProjects` walks all configured root directories recursively |
| AC-07 | Any directory containing a `.git` subdirectory is detected as a project |
| AC-08 | Each discovered repo produces a Project record with id, name, root_path, repository_type, discovered_at |
| AC-09 | Permission errors on individual directories do not halt discovery |
| AC-10 | The operation returns the total count of discovered projects and a list of errors |

### Story 2.3 — Support Nested Folders

| ID | Criterion |
|----|-----------|
| AC-11 | When a subdirectory of a repo is itself a repo, both parent and nested repo are recorded as separate projects |
| AC-12 | Monorepo structures (e.g., `services/auth/`, `services/api/` each with `.git`) are fully discovered |
| AC-13 | No predefined recursion depth limit |

### Story 2.4 — Detect Common Project Types

| ID | Criterion |
|----|-----------|
| AC-14 | `package.json` → `repository_type` includes "node" |
| AC-15 | `go.mod` → `repository_type` includes "go" |
| AC-16 | `requirements.txt` or `pyproject.toml` → `repository_type` includes "python" |
| AC-17 | `Cargo.toml` → `repository_type` includes "rust" |
| AC-18 | `pom.xml` → `repository_type` includes "java" |
| AC-19 | Multiple markers present → `repository_type` reflects all detected types |
| AC-20 | No markers found → `repository_type` remains "unknown" |

### Story 2.5 — Ignore Generated Directories

| ID | Criterion |
|----|-----------|
| AC-21 | `node_modules/`, `vendor/`, `.venv/`, `target/`, `build/`, `dist/` are not treated as projects |
| AC-22 | `.gitignore` patterns are optionally respected during discovery |
| AC-23 | Custom ignore patterns from configuration are applied |
| AC-24 | Skipped directories are logged at DEBUG level |

---

## 6. Data Flow

### 6.1 Discovery Flow

```
User runs: portfolio init (or discoverProjects via MCP)
    │
    ▼
┌─────────────────────────────┐
│  Read configuration file    │
│  (root directories, ignore  │
│   patterns, .gitignore ref) │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  Validate root directories  │
│  - path exists?             │
│  - path accessible?         │
│  - is directory?            │
└─────────────┬───────────────┘
              │
              ▼ (for each root)
┌─────────────────────────────┐
│  Walk directory tree        │
│  (respect ignore patterns)  │
└─────────────┬───────────────┘
              │
              ▼ (for each entry)
┌─────────────────────────────┐
│  Is .git present?           │
│  ┌──No──► Skip (continue)   │
│  │                          │
│  └──Yes──► ┌────────────────┐
│            │ Detect markers │
│            │ (package.json, │
│            │  go.mod, etc.) │
│            └───────┬────────┘
│                    ▼
│            ┌────────────────┐
│            │ Upsert Project │
│            │ record to DB   │
│            │ (key: root_path│
│            └───────┬────────┘
│                    │
│                    ▼
│            ┌────────────────┐
│            │ Continue       │
│            │ recursion into │
│            │ subdirectories │
│            └────────────────┘
              │
              ▼
┌─────────────────────────────┐
│  Return result:             │
│  - discovered: int          │
│  - errors: []error          │
└─────────────────────────────┘
```

### 6.2 Idempotency Strategy

```
discoverProjects()
    │
    ├── If Project with root_path exists in DB:
    │       ├── Update: discovered_at, repository_type
    │       └── (do not duplicate)
    │
    └── If Project with root_path does NOT exist:
            ├── INSERT new Project record
            └── (generate UUID)
```

### 6.3 Data Produced

```
Project {
    id:            UUID       ← generated by engine
    name:          string     ← directory basename
    root_path:     string     ← absolute filesystem path
    repository_type: string   ← "node" | "go" | "python" | "rust" | "java" | "unknown" | "node,go" etc.
    discovered_at: timestamp  ← when discovery ran
    updated_at:    timestamp  ← when record was last updated
}
```

---

## 7. Dependencies

### 7.1 Epic Dependencies

| Dependency | Required For | Notes |
|-----------|-------------|-------|
| Epic 1.2 — Configuration | Reading/writing project roots from config file | Story 2.1 blocked until this exists |
| Epic 1.4 — Database Schema | `projects` table must exist before persisting records | Story 2.2 blocked until migrations are ready |
| Epic 1.5 — Filesystem Utilities | Walking directory trees, checking `.git` presence | Story 2.2 blocked until reusable FS layer exists |

### 7.2 Technology Dependencies

| Dependency | Purpose |
|-----------|---------|
| Go 1.21+ | Engine language |
| SQLite (via Go driver, e.g., `modernc.org/sqlite`) | Knowledge store persistence |
| `os.ReadDir` / `os.Stat` / `filepath.Walk` (Go stdlib) | Filesystem traversal |
| Google UUID package (`github.com/google/uuid`) | Project ID generation |
| Standard Go `testing` + `testing/fstest` | Filesystem-abstraction testing |

### 7.3 Downstream Consumers

| Consumer | What It Needs From This Epic |
|----------|------------------------------|
| Epic 3 — Metadata Extraction | Project records with root_path to read git metadata |
| Epic 4 — Documentation Indexing | Project records with root_path to find README, ADRs, etc. |
| MCP `listProjects()` / `getProject(id)` | Populated projects table |
| HTTP API `GET /projects` | Populated projects table |
| Dashboard | Project listings |

### 7.4 MCP Tools Enabled by This Epic

- `health()` — basic readiness check (no dependency on discovery)
- `discoverProjects()` — triggers the full discovery flow
- `listProjects()` — returns all discovered Project records
- `getProject(id)` — returns a single Project record
