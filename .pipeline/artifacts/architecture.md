# Epic 3 — Metadata Extraction: Architecture

**Milestone:** 1 — Core Engine
**Status:** Draft
**Date:** 2026-07-22

---

## Table of Contents

1. [Component Design](#1-component-design)
2. [Schema Changes](#2-schema-changes)
3. [Sequence Diagrams](#3-sequence-diagrams)
4. [Error Handling](#4-error-handling)
5. [Test Strategy](#5-test-strategy)
6. [Implementation Order](#6-implementation-order)

---

## 1. Component Design

### 1.1 Package Structure

```
internal/
  metadata/
    service.go              # MetadataService orchestrator
    git.go                  # Story 3.1 — extractGitMetadata
    languages.go            # Story 3.2 — detectLanguages
    frameworks.go           # Story 3.3 — detectFrameworks
    dependencies.go         # Story 3.4 — detectDependencies

    dochash.go              # Story 3.6 — computeDocHash
    walk.go                 # Shared: filtered file tree walker
    config.go               # Shared: language/framework mapping config
    metadata_test.go
    git_test.go
    languages_test.go
    frameworks_test.go
    dependencies_test.go

    walk_test.go

internal/
  store/
    metadata.go             # metadata table CRUD
    dependencies.go         # dependencies table CRUD

pkg/
  models/
    metadata.go             # Metadata struct + dependency struct
```

### 1.2 Core Types

```go
// pkg/models/metadata.go

package models

import "time"

type Metadata struct {
    ProjectID        string     `json:"project_id"`
    GitHead          *string    `json:"git_head"`
    DefaultBranch    *string    `json:"default_branch"`
    LastCommitAt     *time.Time `json:"last_commit_at"`
    LastModifiedAt   *time.Time `json:"last_modified_at"`
    CommitCount      int        `json:"commit_count"`
    LanguageSummary  *string    `json:"language_summary"`
    FrameworkSummary *string    `json:"framework_summary"`
    DependencySummary *string   `json:"dependency_summary"`
    DocumentationHash *string   `json:"documentation_hash"`

    LastScanAt       time.Time  `json:"last_scan_at"`
}

type Dependency struct {
    ProjectID string `json:"project_id"`
    Name      string `json:"name"`
    Manager   string `json:"manager"` // e.g. "npm", "go_mod", "pip"
}
```

### 1.3 MetadataService

```go
// internal/metadata/service.go

package metadata

type Service struct {
    store  Store
    walker FileWalker
    logger *zap.Logger
}

type Store interface {
    UpsertMetadata(m *models.Metadata) error
    GetMetadata(projectID string) (*models.Metadata, error)
    InsertDependencies(deps []models.Dependency) error
    ReplaceDependencies(projectID string, deps []models.Dependency) error
}

type FileWalker interface {
    Walk(root string, fn func(path string, info os.FileInfo) error) error
    WalkWithConfig(root string, cfg WalkConfig, fn func(path string, info os.FileInfo) error) error
}
```

### 1.4 Capability Functions

Each story maps to a single exported function. Functions accept a project root path and return their specific output. Functions are independent — no shared state, no pipeline.

| Capability | Signature | Output |
|---|---|---|
| `ExtractGitMetadata` | `func(root string) (*GitResult, error)` | `default_branch`, `git_head`, `last_commit_at`, `last_modified_at`, `commit_count` |
| `DetectLanguages` | `func(root string, walker FileWalker, langMap LanguageMap) (*string, error)` | `language_summary` (comma-sep, prevalence-sorted) |
| `DetectFrameworks` | `func(root string, walker FileWalker, fwMap FrameworkMap) (*string, error)` | `framework_summary` (comma-sep) |
| `DetectDependencies` | `func(root string, walker FileWalker) ([]models.Dependency, *string, error)` | dependency list + `dependency_summary` (top 10) |
| `ComputeDocHash` | `func(docPaths []string) (*string, error)` | `documentation_hash` (SHA-256 hash-of-hashes) |

### 1.5 Walk Service

```go
// internal/metadata/walk.go

type WalkConfig struct {
    IgnoredDirs   []string   // skip these directory basenames
    MaxFiles      int        // safety limit, 0 = unlimited
    FollowSymlink bool       // false = skip symlinks outside root
}

type FileWalker struct {
    logger *zap.Logger
}

func (w *FileWalker) Walk(root string, fn WalkFn) error
func (w *FileWalker) WalkWithConfig(root string, cfg WalkConfig, fn WalkFn) error
```

Walk uses `filepath.WalkDir` with early pruning: when a directory basename matches an entry in `IgnoredDirs`, `fs.SkipDir` is returned immediately. This ensures ignored dirs are never descended into (NFR-5: Disk respect).

### 1.6 Config for Mappings

```go
// internal/metadata/config.go

type LanguageMap struct {
    Extensions map[string]string  // ".go" -> "Go", ".ts" -> "TypeScript"
}

type FrameworkMap struct {
    Markers []FrameworkMarker     // ordered list of detection rules
}

type FrameworkMarker struct {
    Name       string             // "React"
    Manifest   string             // "package.json"
    Pattern    string             // "react" (dependency name or key)
    Ecosystem  string             // "npm"
}
```

Language mappings and framework markers are loaded at startup from a default set embedded in the binary (via `//go:embed`), with optional user overrides via TOML config.

Default language mapping is embedded in `internal/metadata/languages_data.go` as a Go map (top 10 common extensions: .go, .ts/.tsx, .js/.jsx, .py, .rs, .java, .rb, .c/.h, .cs, .swift). The user can override in `~/.portfolio/config.toml`:

```toml
[metadata.languages]
".foo" = "FooLang"
```

Framework mappings are embedded in `internal/metadata/frameworks_data.go`. User overrides:

```toml
[metadata.frameworks]
"MyFramework" = { manifest = "my.json", pattern = "my-framework" }
```

### 1.7 Git Metadata Extraction

Implemented via `os/exec` calling the `git` CLI (not libgit2):

```go
// internal/metadata/git.go

func ExtractGitMetadata(root string) (*GitResult, error)

type GitResult struct {
    GitHead        *string
    DefaultBranch  *string
    LastCommitAt   *time.Time
    LastModifiedAt *time.Time
    CommitCount    int
}
```

**Commands used:**

| Field | Git Command | Edge Cases |
|---|---|---|
| `git_head` | `git rev-parse HEAD` | Fails on empty repo → NULL |
| `default_branch` | `git symbolic-ref refs/remotes/origin/HEAD` → strip refs/remotes/origin/ prefix. Fallback: `git rev-parse --abbrev-ref HEAD` | No remote → local HEAD ref |
| `last_commit_at` | `git log -1 --format=%ct HEAD` (Unix timestamp) | Empty repo → NULL |
| `last_modified_at` | `git status --porcelain` → check for unstaged; then `git log -1 --format=%ct` for last commit time | Unstaged changes → use `stat` on modified files |
| `commit_count` | `git rev-list --count HEAD` | Empty repo → 0 |

Bare repo detection: check if `root/.git` is missing but `root/HEAD` and `root/objects` exist → `last_modified_at` = NULL, skip language/framework detection.

### 1.8 Language Detection

```
Walk(root, skip=vendor,node_modules,.git,build,dist)
  → for each file: extract extension
  → map extension → language name (via LanguageMap)
  → count files per language
  → sort by count (descending)
  → join names: "Go, TypeScript, Shell"
```

Zero code files → `language_summary` = NULL.

### 1.9 Framework Detection

```
Walk(root, skip=vendor,node_modules,.git,build,dist)
  → for each known manifest file (package.json, go.mod, etc.):
    → parse manifest
    → for each FrameworkMarker with matching manifest:
      → check if pattern exists in dependencies/devDependencies
      → if match → add to detected set
  → join detected framework names: "React, Vite, Express"
```

### 1.10 Dependency Detection

```
Walk(root, skip=ignored dirs)
  → for each manifest file:
    → parse by type (JSON, TOMT, requirements.txt line parsing)
    → extract direct dependency names (NOT devDependencies for npm unless also in deps)
    → deduplicate across manifests
  → sort by occurrence across manifests (most referenced first)
  → top 10 → "express, react, lodash, ..."
  → store ALL deps in dependencies table (F-3.4.5)
```

Supported manifest formats:

| Manifest | Parser | Ecosystem |
|---|---|---|
| `package.json` | Go `encoding/json` | npm/yarn/pnpm |
| `go.mod` | Custom line parser | Go |
| `requirements.txt` | Custom line parser | pip |
| `pyproject.toml` | TOML parser | Poetry |
| `Cargo.toml` | TOML parser | Rust/Cargo |
| `Gemfile` | Custom line parser | Ruby/Bundler |
| `pom.xml` | XML parser | Maven |
| `build.gradle` | Custom line parser | Gradle |

### 1.11 Documentation Hash

```go
// internal/metadata/dochash.go

func ComputeDocHash(docContents [][]byte) string {
    // Sort doc paths for determinism (Epic 4 provides sorted paths)
    // For each doc: sha256(content)
    // Combine: sha256(hash1 + hash2 + ... + hashN)
    // Return hex string
}
```

Hashes raw bytes (non-UTF-8 safe). Deterministic by construction. No docs → returns empty string → stored as NULL.

### 1.13 Orchestrator Behavior

`MetadataService.ExtractAll(projectID string) error`:

1. Load project from store (get `root_path`, check existence, handle EC-12)
2. Execute each capability function independently
3. Accumulate partial results — if one capability fails, log warning, continue
4. Assemble `Metadata` struct from results
5. `store.UpsertMetadata(&metadata)` — single SQL UPDATE with all fields
6. If Story 3.4 succeeded: `store.ReplaceDependencies(projectID, deps)`
7. Return nil (partial failures are logged, not propagated)

---

## 2. Schema Changes

### 2.1 Extended `metadata` Table

The schema from PlatformSpecification.md is extended with a field from story 3.1 (commit_count):

```sql
CREATE TABLE IF NOT EXISTS metadata (
    project_id        TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,

    -- Story 3.1 — Git Metadata
    git_head          TEXT,
    default_branch    TEXT,
    last_commit_at    TEXT,            -- ISO 8601 timestamp
    last_modified_at  TEXT,            -- ISO 8601 timestamp
    commit_count      INTEGER DEFAULT 0,

    -- Story 3.2 — Language Detection
    language_summary  TEXT,            -- comma-separated, e.g. "Go, TypeScript, Shell"

    -- Story 3.3 — Framework Detection
    framework_summary TEXT,            -- comma-separated, e.g. "React, Express"

    -- Story 3.4 — Dependency Detection
    dependency_summary TEXT,           -- comma-separated top 10, e.g. "express, react, lodash"

    -- Story 3.6 — Documentation Hash
    documentation_hash TEXT,           -- SHA-256 hex

    -- Tracking
    last_scan_at      TEXT             -- ISO 8601 timestamp
);
```

### 2.2 New `dependencies` Table

Stores ALL direct dependencies (not just top 10), supporting downstream Epic 13 (relationship analysis).

```sql
CREATE TABLE IF NOT EXISTS dependencies (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,             -- dependency package name
    manager     TEXT NOT NULL,             -- ecosystem: "npm", "go_mod", "pip", "cargo", "bundler", "maven", "gradle"
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),

    UNIQUE(project_id, name, manager)
);

CREATE INDEX idx_dependencies_project_id ON dependencies(project_id);
CREATE INDEX idx_dependencies_name ON dependencies(name);
```

---

## 3. Sequence Diagrams

### 3.1 Full Metadata Extraction

```
Caller (CLI/MCP)         MetadataService         Git          Walk        Frameworks    Store
        |                       |                  |            |            |            |
        |--- ExtractAll(id) --->|                  |            |            |            |
        |                       |--- getProject--->|            |            |            |
        |                       |<-- rootPath -----|            |            |            |
        |                       |                  |            |            |            |
        |                       |-- E1: gitHead -->|            |            |            |
        |                       |<-- GitResult ----|            |            |            |
        |                       |                  |            |            |            |
        |                       |-- E2: langs ---->|-- Walk --->|            |            |
        |                       |<-- langSummary <-|<-- files --|            |            |
        |                       |                  |            |            |            |
        |                       |-- E3: fwks ---->|-- Walk --->|            |            |
        |                       |                  |            |-- parse -->|            |
        |                       |<-- fwSummary ---|            |            |            |
        |                       |                  |            |            |            |
        |                       |-- E4: deps ---->|-- Walk --->|            |            |
        |                       |                  |            |-- parse -->|            |
        |                       |<-- [deps], top10|            |            |            |
        |                       |                  |            |            |            |
        |                       |-- E5: docHash -->|            |            |            |
        |                       |<-- hash --------|            |            |            |
        |                       |                  |            |            |            |
        |                       |--- UpsertMeta ---------------------------->|            |
        |                       |--- ReplaceDeps ---------------------------->|            |
        |<-- OK ----------------|                  |            |            |            |
```

### 3.2 Partial Re-scan (e.g. Story 3.2 Only)

```
Caller                     MetadataService         Walk          Store
  |--- DetectLanguages ---->|                       |              |
  |                         |--- Walk(root) ------->|              |
  |                         |<-- files -------------|              |
  |                         | (count extensions)    |              |
  |                         |--- UpsertMeta(lang) -->|              |
  |<-- langSummary ---------|                       |              |
```

### 3.3 Partial Failure (Story 3.3 Fails)

```
Caller                     MetadataService         Frameworks    Store
  |--- ExtractAll(id) ---->|                       |              |
  |                         |-- Git OK ----------->|              |
  |                         |-- Langs OK --------->|              |
  |                         |                       |              |
  |                         |-- Frameworks -------->|              |
  |                         |    (corrupted        |              |
  |                         |     go.mod)          |              |
  |                         |<-- error ------------|              |
  |                         |  log WARN "framework |              |
  |                         |  detection failed"   |              |
  |                         |                       |              |
  |                         |-- Deps OK ---------->|              |
  |                         |                       |              |
  |                         |-- UpsertMeta -------->|              |
  |                         |  (git+lang+deps      |              |
  |                         |   +frameworks)        |              |
  |<-- OK (partial) --------|                       |              |
```

### 3.4 Empty Repository

```
Caller                     MetadataService         Git            Store
  |--- ExtractAll(id) ---->|                       |                |
  |                         |-- git rev-parse ---->|                |
  |                         |<-- error ------------|                |
  |                         |  (fatal: bad revision HEAD)          |
  |                         |                       |                |
  |                         |-- is-empty-repo? ---->|                |
  |                         |  git rev-list --count HEAD           |
  |                         |<-- error ------------|                |
  |                         |                       |                |
  |                         | Result:                              |
  |                         |   git_head=NULL                       |
  |                         |   commit_count=0                      |
  |                         |   last_commit_at=NULL                 |
  |                         |   default_branch=origin_HEAD or main  |
  |                         |   lang/stats/deps/fw = (skip)         |
  |                         |                       |                |
  |                         |-- UpsertMeta(NULLs) ->|                |
  |<-- OK ------------------|                       |                |
```

---

## 4. Error Handling

### 4.1 Error Classification

| Category | Example | Behavior |
|---|---|---|
| **Skip-and-continue** | Permission denied on file | Log WARN, skip file, continue walk |
| **Capability failure** | Corrupted manifest JSON | Log WARN, return nil for that field, continue other capabilities |
| **Project not found** | Deleted/moved repo | Log WARN, skip project entirely, return nil |
| **Fatal** | DB write failure, config parse failure | Return error, abort entire operation |

### 4.2 Edge Case Handling

| EC | Handler |
|---|---|
| **EC-1 Empty repo** | `gitHead`: git command fails → set NULL. `commitCount`: 0. Other capabilities: walk repo with no files → NULLs/zeros. |
| **EC-2 Bare repo** | Check for `.git` being absent but `HEAD` present. Skip language/framework/deps/stats. `lastModifiedAt` = NULL. |
| **EC-3 Detached HEAD** | `git rev-parse HEAD` still works. `defaultBranch` resolved from `refs/remotes/origin/HEAD` not local HEAD. |
| **EC-4 No recognized language** | Walk produces files but no extensions match → `languageSummary` = NULL. `codeFiles` = 0, `locEstimate` = 0. |
| **EC-5 Unknown extensions** | Files count toward `totalFiles` but not `codeFiles`. Not included in `languageSummary`. |
| **EC-6 Polyglot 20+** | No truncation in summary. Sorted by extension count descending. All names included. |
| **EC-7 Multi-framework** | No limit on framework count. Union of all matches across manifests. |
| **EC-8 Monorepo** | Walk finds multiple manifests at different depths. Each is parsed independently. Dependencies deduplicated by name across manifests. |
| **EC-9 Corrupted manifest** | JSON/TOML/XML parse error → log WARN with path, skip that file, continue. |
| **EC-10 >100k files** | `WalkConfig.MaxFiles` set high. LOC estimation uses sampling: count lines in every Nth file (N=ceil(totalFiles/10000)). |
| **EC-11 No doc files** | `ComputeDocHash([])` returns empty string → stored as NULL. |
| **EC-12 Deleted/moved repo** | `os.Stat(rootPath)` fails → log WARN "project directory not found, skipping", return nil. |
| **EC-13 Symlink escape** | Walk: `filepath.EvalSymlinks` on symlink targets. If symlink target is outside `root`, skip. |
| **EC-14 Non-UTF-8 docs** | Read as `[]byte`, hash raw bytes, not decoded string. |
| **EC-15 No remote** | `git symbolic-ref refs/remotes/origin/HEAD` fails → fallback to `git rev-parse --abbrev-ref HEAD` for local branch. |
| **EC-16 Permission error** | Walk: `os.Stat` / `os.Open` returns permission error → log WARN, `return nil` for the single file, continue walk. |

### 4.3 Logging Conventions

```
[WARN] [metadata] skipped inaccessible file: /path/to/file (permission denied)
[WARN] [metadata] corrupt manifest, skipping: /path/to/package.json (invalid JSON at line 42)
[WARN] [metadata] framework detection failed for project abc-123: no manifest files found
[INFO] [metadata] extracted metadata for project abc-123 (6/6 capabilities OK)
[INFO] [metadata] extracted metadata for project abc-123 (5/6 capabilities, frameworks skipped)
```

---

## 5. Test Strategy

### 5.1 Unit Tests

| Package | Tests | Approach |
|---|---|---|
| `internal/metadata/git_test.go` | 5 | Use `testdata/` git repos with known states: normal, empty, bare, detached HEAD, no remote |
| `internal/metadata/languages_test.go` | 4 | Use synthetic file trees. Assert correct prevalence sorting, unknown ext handling, empty project |
| `internal/metadata/frameworks_test.go` | 4 | Use manifest fixtures (package.json with react, go.mod with gin, etc.) |
| `internal/metadata/dependencies_test.go` | 6 | One per manifest format + monorepo + corrupted manifest + empty |
| `internal/metadata/dochash_test.go` | 3 | Known content → known hash, determinism, empty input |
| `internal/metadata/walk_test.go` | 4 | Ignored dir skipping, symlink handling, permission error, max files limit |

### 5.2 Integration Tests

| Test | Description |
|---|---|
| `TestFullExtraction` | Create real git repo with files, run all 6 stories, verify DB state |
| `TestPartialReScan` | Extract, change one file, re-run single capability, verify only that field updates |
| `TestIdempotency` | Run extraction twice on same repo, verify all fields match (except last_scan_at) |
| `TestConcurrentExtraction` | Extract metadata for two projects in parallel, verify no cross-contamination |

### 5.3 Test Fixtures

```
testdata/
  repos/
    normal/              # Real git repo: Go + JS files, frameworks, deps
    empty/               # git init with no commits
    bare.git/            # git init --bare
    detached/            # git checkout --detached
    no-remote/           # Local-only repo
    polyglot/            # 5+ languages, 3+ frameworks
    monorepo/            # Multiple manifests in subdirs
    corrupt-manifest/    # Invalid package.json
    >100k/               # (generated at test time if needed)

  fixtures/
    package.json
    go.mod
    Cargo.toml
    requirements.txt
    pyproject.toml
    Gemfile
    pom.xml
    build.gradle
```

### 5.4 Testing Principles

- Git-backed test repos are actual git repos (created via shell or `go-git`), not mock files
- Language/framework detection tested against real manifests, not stubs
- Walk tests use real filesystem, not in-memory
- Permission tests: `chmod 000` + defer `chmod 644`
- Determinism assertions: run twice, assert field equality
- Performance regression tests: use `testing.B` benchmarks for each capability

---

## 6. Implementation Order

```
Story 3.1 ───────────────────────────────────────── 3 days
  ├── internal/metadata/git.go
  ├── internal/metadata/walk.go              (shared dependency)
  ├── internal/store/metadata.go             (UpsertMetadata)
  ├── store schema migration (add columns)
  └── internal/metadata/git_test.go

Story 3.2 ───────────────────────────────────────── 3 days
  ├── internal/metadata/languages.go
  ├── internal/metadata/config.go            (LanguageMap)
  ├── internal/metadata/languages_data.go    (embedded map)
  └── internal/metadata/languages_test.go

Story 3.3 ───────────────────────────────────────── 3 days
  ├── internal/metadata/frameworks.go
  ├── internal/metadata/frameworks_data.go   (embedded markers)
  └── internal/metadata/frameworks_test.go

Story 3.4 ───────────────────────────────────────── 3 days
  ├── internal/metadata/dependencies.go
  ├── internal/store/dependencies.go         (ReplaceDependencies)
  ├── dependencies table migration
  └── internal/metadata/dependencies_test.go

Story 3.6 ───────────────────────────────────────── 1 day
  ├── internal/metadata/dochash.go
  └── internal/metadata/dochash_test.go

Service assembly ────────────────────────────────── 2 days
  ├── internal/metadata/service.go           (orchestrator)
  ├── internal/metadata/service_test.go
  └── Integration tests
```

**Total: ~15 days** (within Epic 3 estimate of ~15 days)

---

## 7. MCP Tool Mapping

Each capability is exposed as a separate MCP tool (principle: capabilities over workflows):

| MCP Tool | Maps To | Returns |
|---|---|---|
| `extractGitMetadata(projectId)` | Story 3.1 | Git fields |
| `detectLanguages(projectId)` | Story 3.2 | `language_summary` |
| `detectFrameworks(projectId)` | Story 3.3 | `framework_summary` |
| `detectDependencies(projectId)` | Story 3.4 | `dependency_summary` + dependency list |
| `computeDocumentationHash(projectId)` | Story 3.6 | `documentation_hash` |
| `extractMetadata(projectId)` | Aggregate | All metadata fields (writes to store) |

A caller (CLI agent or AI agent) can invoke individual capabilities or the full aggregate. The engine does NOT mandate a pipeline — agents compose workflows.
