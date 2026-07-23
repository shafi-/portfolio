# Epic 2 — Project Discovery: Architecture Document

**Milestone:** 1 — Core Engine
**Status:** Draft
**Documents:** KnowledgeModel.md, PlatformSpecification.md, Guideline.md
**Requirements:** .requirements/epic-02-requirements.md
**Tasks:** docs/tasks/epic-02-discovery.md

---

## 1. Package Structure

```
internal/
├── config/
│   └── provider.go          ← ConfigProvider interface (Epic 1 provides impl)
├── discovery/
│   ├── discoverer.go        ← Orchestrator: coordinates walk + detect + persist
│   ├── walker.go            ← Filesystem walker (ignore patterns, symlink safety)
│   ├── detector.go          ← Git repo detection (.git dir/file, bare repos)
│   ├── marker.go            ← Project type marker detection
│   └── types.go             ← DiscoveryResult, WalkEvent, shared enums
├── store/
│   └── project_store.go     ← ProjectStore interface (Epic 1 provides impl)
├── model/
│   └── project.go           ← Project entity (shared across all interfaces)
└── fs/
    └── filesystem.go        ← Filesystem interface (enables testability)
```

All packages live under `internal/` — no external consumers of discovery internals.

---

## 2. Component Interfaces

Every dependency is an interface. No package imports concrete implementations of its peers.

### 2.1 Filesystem Abstraction

```go
// internal/fs/filesystem.go
type Filesystem interface {
    ReadDir(path string) ([]os.DirEntry, error)
    Lstat(path string) (os.FileInfo, error)
    Stat(path string) (os.FileInfo, error)
    Open(name string) (io.ReadCloser, error)
    ReadFile(path string) ([]byte, error)
}
```

Default implementation uses `os` package. Test implementation uses `testing/fstest.MapFS`.

### 2.2 Config Provider

```go
// internal/config/provider.go
type Provider interface {
    GetProjectRoots() ([]string, error)
    GetIgnorePatterns() ([]string, error)
    IsGitignoreEnabled() bool
}
```

Epic 1 provides the concrete implementation. Discovery only consumes the interface.

### 2.3 Project Store

```go
// internal/store/project_store.go
type ProjectStore interface {
    UpsertProject(ctx context.Context, p *model.Project) (*model.Project, error)
    ListProjects(ctx context.Context) ([]*model.Project, error)
    GetProject(ctx context.Context, id uuid.UUID) (*model.Project, error)
    GetProjectByRootPath(ctx context.Context, rootPath string) (*model.Project, error)
}
```

Epic 1 provides the concrete SQLite implementation. Key operation: `UpsertProject` — looks up by `root_path`, updates if exists, inserts if not.

### 2.4 Discovery Orchestrator

```go
// internal/discovery/discoverer.go
type UUIDProvider interface {
    New() uuid.UUID
}

type Discoverer struct {
    config   config.Provider
    store    store.ProjectStore
    fs       fs.Filesystem
    marker   MarkerDetector
    walker   *Walker
    uuid     UUIDProvider
    mu       sync.Mutex  // prevents concurrent discovery invocations
}

func NewDiscoverer(cfg config.Provider, st store.ProjectStore, f fs.Filesystem, u UUIDProvider) *Discoverer

func (d *Discoverer) DiscoverProjects(ctx context.Context) (*DiscoveryResult, error)
```

### 2.5 Marker Detector

```go
// internal/discovery/marker.go
type MarkerDetector interface {
    DetectMarkers(ctx context.Context, rootPath string) ([]string, error)
}

type markerDetector struct {
    fs fs.Filesystem
}
```

Returns slice of type strings: `["node"]`, `["go"]`, `["node", "go"]`, or `["unknown"]`.

### 2.6 Walker

```go
// internal/discovery/walker.go
type WalkEvent int

const (
    EventEnterDir  WalkEvent = iota  // entering a directory (before children)
    EventFoundRepo                    // directory contains .git
    EventError                        // error during walk
    EventSkipped                      // directory skipped (permissions, ignore)
)

type WalkCallback func(ctx context.Context, path string, event WalkEvent, err error) error

// Returning a non-nil error from WalkCallback aborts the walk immediately.

type Walker struct {
    fs            fs.Filesystem
    ignoreMatcher IgnoreMatcher
    visitedInodes map[InodeKey]struct{}  // symlink loop detection
}

func NewWalker(f fs.Filesystem, ignoreMatcher IgnoreMatcher) *Walker

func (w *Walker) Walk(ctx context.Context, root string, fn WalkCallback) error
```

Internal (unexported) types:

```go
type InodeKey struct {
    Dev  uint64
    Inode uint64
}
```

---

## 3. Data Model

### 3.1 Project Entity

```go
// internal/model/project.go
package model

import (
    "time"
    "github.com/google/uuid"
)

type Project struct {
    ID              uuid.UUID `db:"id"`
    Name            string    `db:"name"`             // populated from filepath.Base(RootPath)
    RootPath        string    `db:"root_path"`        // unique key for upsert
    RepositoryType  string    `db:"repository_type"`   // "node,go" | "unknown" | etc.
    DiscoveredAt    time.Time `db:"discovered_at"`
    UpdatedAt       time.Time `db:"updated_at"`
}
```

### 3.2 Database Schema (reproduced from PlatformSpecification.md)

```sql
CREATE TABLE projects (
    id              TEXT PRIMARY KEY,       -- UUID as string
    name            TEXT NOT NULL,
    root_path       TEXT NOT NULL UNIQUE,   -- unique for idempotent upsert
    repository_type TEXT NOT NULL DEFAULT 'unknown',
    discovered_at   TEXT NOT NULL,          -- ISO 8601
    updated_at      TEXT NOT NULL           -- ISO 8601
);

CREATE INDEX idx_projects_root_path ON projects(root_path);
```

No schema changes beyond PlatformSpecification.md. This epic populates the `projects` table.

### 3.3 Derived Indicators (Compute, Don't Store)

| Indicator | Formula |
|-----------|---------|
| `needs_analysis` | `metadata.last_scan_at IS NULL` |
| `analysis_outdated` | `metadata.last_scan_at < metadata.last_commit_at` |
| `documentation_changed` | `current_hash != metadata.documentation_hash` |

Not stored. Computed on query by downstream consumers.

---

## 4. Key Flows

### 4.1 Initiating Discovery (CLI + MCP)

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌───────────┐
│ CLI/MCP  │     │  Discoverer  │     │  Store       │     │   Config  │
│ (caller) │     │              │     │              │     │  Provider │
└────┬─────┘     └──────┬───────┘     └──────┬────────┘     └─────┬─────┘
     │                  │                     │                   │
     │  discoverProjects│                     │                   │
     │─────────────────>│                     │                   │
     │                  │  GetProjectRoots()  │                   │
     │                  │────────────────────────────────────────>│
     │                  │                     │                   │
     │                  │  []string, error    │                   │
     │                  │<────────────────────────────────────────│
     │                  │                     │                   │
     │                  │  Validate roots     │                   │
     │                  │  (exists, dir,      │                   │
     │                  │   accessible)       │                   │
     │                  │                     │                   │
     │                  │  for each root:     │                   │
     │                  │    Walker.Walk()    │                   │
     │                  │    ────┐            │                   │
     │                  │        │ EventFoundRepo(path)          │
     │                  │    <───┘            │                   │
     │                  │                     │                   │
     │                  │  DetectMarkers(path)│                   │
     │                  │  ────┐              │                   │
     │                  │      │ []string     │                   │
     │                  │  <───┘              │                   │
     │                  │                     │                   │
     │                  │  UpsertProject(p)   │                   │
     │                  │────────────────────>│                   │
     │                  │                     │                   │
     │  DiscoveryResult │                     │                   │
     │<─────────────────│                     │                   │
 ```

After each repo is detected, the Discoverer:
1. Calls `DetectMarkers` to identify project type markers
2. Checks if the repo is bare (has `HEAD` + `objects/` without `.git` dir) via the detector
3. Computes `repository_type`:
   - Bare repo with markers → `"bare-git,go"` (bare-git prepended to sorted marker types)
   - Bare repo with no markers → `"bare-git"` (overrides "unknown")
   - Non-bare repo → marker types as-is

`DiscoverProjects` is guarded by a mutex — concurrent calls return an error immediately.

### 4.2 Walker Internal Flow

```
Walker.Walk(root)
  │
  ├─ Validate root (exists, is dir)
  ├─ Create pathQueue: [root]
  │
  └─ for each dir in pathQueue:
       │
       ├─ Check ignore patterns
       │   └─ matched → log DEBUG, EventSkipped, continue
       │
       ├─ ReadDir(dir) → entries
       │   └─ error → EventError, continue next dir
       │
       ├─ Scan entries for:
       │   ├─ .git          → EventFoundRepo
       │   ├─ subdirectories → add to pathQueue (for recursion)
       │   └─ symlinks       → stat, check inode cache
       │       └─ seen before → skip (break symlink loop)
       │
       └─ Call WalkCallback per event
```

Walker uses BFS (queue-based) rather than recursion to avoid stack overflow on deep trees. `pathQueue` is a simple slice used as a FIFO.

### 4.3 Idempotent Upsert Flow

`Project.Name` is populated from `filepath.Base(project.RootPath)` (directory basename) before upsert.

```
UpsertProject(ctx, project)
  │
  ├─ existing := GetByRootPath(ctx, project.RootPath)
  │
  ├─ if existing != nil:
  │   ├─ project.ID = existing.ID           // preserve original UUID
  │   ├─ project.Name unchanged             // preserve original name on re-discovery
  │   ├─ project.DiscoveredAt = now         // update timestamps
  │   ├─ project.UpdatedAt = now
  │   ├─ UPDATE projects SET ... WHERE id = existing.ID
  │   └─ return updated record
  │
  └─ if existing == nil:
      ├─ project.ID = d.uuid.New()          // via injected UUIDProvider
      ├─ project.DiscoveredAt = now
      ├─ project.UpdatedAt = now
      ├─ INSERT INTO projects (...)
      └─ return new record
```

### 4.4 Marker Detection Flow

```
DetectMarkers(ctx, rootPath)
  │
  ├─ knownMarkers := map[string]string{
  │     "package.json":      "node",
  │     "go.mod":            "go",
  │     "requirements.txt":  "python",
  │     "pyproject.toml":    "python",
  │     "Cargo.toml":        "rust",
  │     "pom.xml":           "java",
  │   }
  │
  ├─ detected := []string{}
  │
  ├─ for markerFile, typeName := range knownMarkers:
  │   ├─ path := filepath.Join(rootPath, markerFile)
  │   ├─ exists := fs.Stat(path) == nil
  │   ├─ if exists → append(detected, typeName)
  │
  ├─ if len(detected) == 0 → detected = ["unknown"]
  │
  └─ return detected, nil
```

Marker detection checks file existence only (not content). Sorting of detected types is alphabetically stable for determinism.

### 4.5 Ignore Pattern Matching

```go
type IgnoreMatcher struct {
    builtIn     []string  // node_modules, vendor, .venv, target, build, dist
    configured  []string  // from config file
    gitignore   []string  // loaded from .gitignore if enabled
    fs          fs.Filesystem
}

func (m *IgnoreMatcher) Matches(absPath string) bool {
    base := filepath.Base(absPath)
    // Check built-in patterns
    for _, p := range m.builtIn {
        if base == p { return true }
    }
    // Check configured patterns
    for _, p := range m.configured {
        if matched, _ := filepath.Match(p, base); matched { return true }
    }
    // Check .gitignore
    for _, p := range m.gitignore {
        if matched, _ := filepath.Match(p, base); matched { return true }
    }
    return false
}
```

`.gitignore` patterns are loaded lazily per repository root (not per visited directory) for performance. Only the root `.gitignore` of each discovered repo is consulted. This is configurable via `IsGitignoreEnabled()` — when it returns `false`, `.gitignore` loading is skipped entirely.

---

## 5. Symlink Safety

Symlink loop detection uses a map of `(device, inode)` pairs tracked per walk root.

```go
type walkState struct {
    visited map[InodeKey]struct{}
}

func (w *Walker) walkDir(ctx context.Context, dir string, state *walkState, fn WalkCallback) error {
    info, err := w.fs.Lstat(dir)
    if err != nil { return err }

    if info.Mode()&os.ModeSymlink != 0 {
        key := InodeKey{Dev: uint64(info.Sys().Dev), Inode: uint64(info.Sys().Ino)}
        if _, seen := state.visited[key]; seen {
            return nil  // skip, already visited
        }
        state.visited[key] = struct{}{}
    }

    // ... continue walking
}
```

Note: On macOS, `Stat` follows symlinks, so we must use `Lstat` for the actual link detection. The implementation should use `os.Lstat` for entries and only follow symlinks that point to directories that haven't been visited.

---

## 6. Error Handling Strategy

### 6.1 Error Classification

| Category | Examples | Action |
|----------|----------|--------|
| **Fatal** | DB connection failure, config unreadable | Abort discovery, return error |
| **Root error** | Root missing, root is file, permission denied on root | Skip root, add to errors list, continue other roots |
| **Walk error** | Permission denied on subdirectory, broken symlink | Log warning, skip entry, continue walk |
| **Record error** | DB upsert fails for one project | Log error, skip that project, continue walk |
| **Detection error** | Cannot read marker file | Log DEBUG, default to "unknown" |

### 6.2 Error Accumulation

```go
type DiscoveryResult struct {
    Discovered   int
    Updated      int
    Errors       []DiscoveryError
}

type DiscoveryError struct {
    Path    string
    Err     error
    Severity string  // "warn" | "error"
}
```

Errors never abort the full discovery unless they are fatal (config, DB). Partial results are returned with accumulated errors.

### 6.3 Interruption Safety

Each project upsert is an individual `INSERT OR UPDATE` statement. No multi-project transaction. If discovery is interrupted:

- Completed projects remain persisted
- Unprocessed roots are simply missing — re-running is safe (idempotent)
- No cleanup needed

OS signal handling (SIGINT/SIGTERM) cancels the walker context via `signal.NotifyContext`, causing the in-progress walk to return a cancellation error. The Discoverer catches this and returns any projects upserted before the signal as partial results, alongside the cancellation error.

### 6.4 Edge Case Handling Matrix

| Edge Case | Detection | Action |
|-----------|-----------|--------|
| Permission denied | `os.ReadDir` returns `Permission` error | Log WARN, skip dir, continue siblings |
| Broken symlink | `os.Lstat` succeeds but target missing | Log DEBUG, skip entry |
| Circular symlink | Inode already in `visited` set | Skip silently |
| `.git` is a file (worktree) | `os.Stat` returns file, not dir | Read first line for `gitdir:`, treat as valid |
| Bare repo | Has `HEAD` + `objects/` but no `.git` dir | Only matches if top-level root with these markers |
| Git repo with no commits | `.git` dir exists with `HEAD` but no commit objects | Discovered normally, `repository_type` computed from markers; no special handling needed |
| Broken `.git` directory | `.git` exists but is unreadable/corrupt | Log WARN, create record with partial fields, continue discovery |
| Very deep tree | OS/hardware limit | No artificial cutoff; let OS error propagate |
| Non-UTF-8 path | ReadDir returns raw bytes | Check with `unicode/utf8.ValidString`; if invalid, attempt encoding fallback or skip with WARN (portable across Go 1.21+) |
| Root deleted between config and walk | `os.Stat` on root fails | Add to errors, skip root, continue other roots |
| No roots configured | Config returns empty slice | Return empty result, log WARN |
| Duplicate roots | Config returns `["/a", "/a"]` | Deduplicate at load time |

---

## 7. Performance Considerations

| Concern | Strategy |
|---------|----------|
| 100k+ files | Use `os.ReadDir` (no `lstat` per entry) instead of `filepath.Walk`; only stat dirs |
| Memory growth | BFS queue bounded by directory depth, not total file count |
| Ignore patterns | Check before `ReadDir` — skip entire subtrees at the directory level |
| Symlink tracking | `visited` map cleared per root, not persisted |
| Marker detection | Only check repo roots (not every directory) — O(n) where n = discovered repos, not files |

The custom walker avoids `filepath.Walk` because `Walk` calls `os.Lstat` on every file. For 100k+ files this is expensive. Instead, use `os.ReadDir` which returns `os.DirEntry` without stat'ing, then call `Stat` only for subdirectories we need to recurse into.

---

## 8. Test Strategy

### 8.1 Unit Tests (Deterministic, No Filesystem, No DB)

| Package | What to Test | Approach |
|---------|-------------|----------|
| `discovery/walker` | Walk calls callback for repos, skips ignored dirs, handles permission errors, detects symlink loops | Use `testing/fstest.MapFS` as mock filesystem |
| `discovery/detector` | Detects `.git` dir, `.git` file (worktree), bare repo, no `.git` | MapFS with various git structures |
| `discovery/marker` | Detects all 6 marker types, polyglot, unknown, empty marker files | MapFS with marker files |
| `discovery/discoverer` | Orchestration: invokes walker, calls marker detector, calls store upsert | Mock all 3 dependencies |

The `Discoverer` accepts an injected `UUIDProvider` interface (see §2.4). In unit tests, a mock provider returns deterministic UUIDs (e.g., `uuid.MustParse("00000000-0000-0000-0000-000000000001")`), enabling predictable assertions on created `Project.ID` fields.

### 8.2 Integration Tests (Real Filesystem, Real SQLite)

| Scenario | Setup | Verify |
|----------|-------|--------|
| Full discovery | Temp dir with multiple repos | All repos discovered, correct fields |
| Idempotency | Run discovery twice | Same number of records, timestamps updated, no duplicates |
| Nested repos | Repo inside repo's subdirectory | Both parent and child have records |
| Permission errors | `chmod 000` on one subtree | Discovery continues, error reported |
| Ignore patterns | `node_modules` with `.git` inside | Skipped, logged at DEBUG |
| No repos | Empty directory | Zero discovered, no errors |
| Unknown type | Repo with no marker files | `repository_type = "unknown"` |
| Polyglot | Repo with `package.json` + `go.mod` | `repository_type = "go,node"` |
| Bare repo | Init with `--bare` | Detected, marked `bare-git` |

### 8.3 What NOT to Test

- AI agent behavior (not deterministic)
- HTTP API or MCP layer (covered by Epics 6 and 7)
- Dashboard rendering (Epic 11/12)
- Configuration file format details (Epic 1)

### 8.4 Acceptance Criteria Mapping

| Test | ACs Covered |
|------|------------|
| Discovery creates correct Project records | AC-06, AC-07, AC-08 |
| Permission errors don't halt | AC-09 |
| DISCOVERY result contains counts + errors | AC-10 |
| Nested repos both recorded | AC-11, AC-12 |
| No depth limit | AC-13 |
| Each marker → correct type | AC-14 through AC-18 |
| Multiple markers combined | AC-19 |
| No markers → "unknown" | AC-20 |
| Built-in dirs skipped | AC-21 |
| Custom ignores work | AC-23 |
| Skipped dirs logged at DEBUG | AC-24 |
| Idempotent re-run | (implicit from FR-12) |

---

## 9. Implementation Order

### Phase 1: Story 2.1 — Configure Project Roots (S, ~2 days)

**Depends on:** Epic 1.2 (config), 1.4 (DB schema)

Tasks:
1. Define `config.Provider` interface in `internal/config/provider.go`
2. Implement config file read/write for project roots
3. Add `portfolio init` prompt for root directories
4. Add validation: path exists, is dir, is accessible
5. Add `portfolio config set-root` and `portfolio config list-roots` commands
6. Write unit tests for config validation
7. Write integration test: persist roots, read them back

Deliverable: User can configure root directories via CLI.

### Phase 2: Story 2.2 — Recursive Project Discovery (L, ~6 days)

**Depends on:** 2.1, Epic 1.5 (filesystem utilities)

Tasks:
1. Implement `fs.Filesystem` interface + `osFilesystem` in `internal/fs/filesystem.go`
2. Implement `ProjectStore` interface + SQLite impl in `internal/store/project_store.go`
3. Implement git detector in `internal/discovery/detector.go`
4. Implement core walker in `internal/discovery/walker.go` (basic walk, no ignores yet)
5. Implement `Discoverer` orchestrator in `internal/discovery/discoverer.go`
6. Write unit tests with mock FS for walker + detector
7. Write integration test: discover real repos in temp dir
8. Write idempotency test: discover twice, verify upsert behavior

Deliverable: Engine can discover all Git repos from configured roots and persist them.

### Phase 3: Story 2.5 — Ignore Generated Directories (S, ~2 days)

**Depends on:** 2.2

Tasks:
1. Implement `IgnoreMatcher` with built-in patterns in `internal/discovery/walker.go`
2. Wire ignore checks into walker's `ReadDir` loop
3. Add configured pattern support from config
4. Add `.gitignore` support (load from each repo root)
5. Add DEBUG logging for skipped directories
6. Write unit tests: ignore patterns match expected dirs
7. Write integration test: repo inside `node_modules` is skipped

Deliverable: Common generated directories are skipped during discovery.

### Phase 4: Story 2.4 — Detect Common Project Types (M, ~3 days)

**Depends on:** 2.2

Tasks:
1. Implement `MarkerDetector` in `internal/discovery/marker.go`
2. Add all 6 marker checks: `package.json`, `go.mod`, `requirements.txt`, `pyproject.toml`, `Cargo.toml`, `pom.xml`
3. Handle polyglot: multiple markers → sorted, comma-separated string
4. Handle unknown: no markers → `"unknown"`
5. Wire into discoverer (call after walker finds repo)
6. Write unit tests for each marker type + combinations
7. Write integration test: create repos with various markers

Deliverable: Discovered projects have accurate `repository_type` field.

### Phase 5: Story 2.3 — Support Nested Folders (M, ~3 days)

**Depends on:** 2.2 (walker already continues recursion), 2.5 (ignore patterns prevent false positives)

Note: Much of this is already handled if the walker in Phase 2 recursively descends. The story tasks are:

1. Verify walker continues descending after finding `.git` at parent level
2. Add test: repo at `/root` and `/root/services/auth/.git` → both discovered
3. Handle monorepo edge case: what if parent has no `.git` but subdirectory does? (already works)
4. Symlink safety: ensure we don't follow into nested symlinked repos
5. Write integration test: multi-level monorepo structure

Deliverable: Nested repositories are correctly discovered as separate projects.

---

## 10. Dependency Graph

```
Phase 1 (2.1)
  │
  ▼
Phase 2 (2.2)
  │
  ├──────────────┬──────────────┐
  ▼              ▼              ▼
Phase 3 (2.5)  Phase 4 (2.4)  Phase 5 (2.3)
  │
  ├── 2.3 benefits from 2.5 (ignore patterns prevent false positives in monorepo)
  └── 2.4 is independent of 2.3 and 2.5
```

Phase 3, 4, and 5 can be implemented in any order. Recommended: 2.5 first (most value for performance + correctness), then 2.4, then 2.3.

---

## 11. CLI Command Surface

```
portfolio init                          # Interactive: prompts for root directories
                                         # Creates initial config, runs discovery

portfolio config set-root <path>        # Add a root directory
portfolio config remove-root <path>     # Remove a root directory  
portfolio config list-roots             # List configured roots
portfolio config set-ignore <pattern>   # Add custom ignore pattern
portfolio config list-ignore            # List configured ignore patterns

portfolio discover                      # Run discovery (non-interactive)
portfolio discover --verbose            # Run with DEBUG-level logging
portfolio projects list                # List all discovered projects
portfolio projects get <id>            # Get project by ID
```

---

## 12. MCP Tools Enabled (Spec Only — Implemented in Epic 7)

| Tool | What It Does | Data Source |
|------|-------------|-------------|
| `discoverProjects()` | Triggers `Discoverer.DiscoverProjects()` | Returns `DiscoveryResult` |
| `listProjects()` | Calls `store.ListProjects()` | Returns `[]Project` |
| `getProject(id)` | Calls `store.GetProject(id)` | Returns single `Project` |
| `health()` | Returns engine status | Static + DB ping |

These are specified here for interface completeness but implemented in Epic 7 (MCP Server).
