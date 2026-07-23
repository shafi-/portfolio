# Epic 4 — Documentation Indexing: Architecture

**Milestone:** 1 — Core Engine
**Status:** Draft
**Version:** 0.1

---

## 1. Component Design

### 1.1 Package Structure

```
internal/
  indexer/
    indexer.go            # Indexer — public API, orchestrates per-project indexing
    runner.go             # IndexRunner — per-project indexing pipeline
    reader.go             # DocReader — file reading, hashing, size limiting
    discoverer.go         # DocDiscoverer — finds files by kind across project
    fts.go                # FTSManager — FTS5 query (index synced via Epic 5 triggers)
    dedup.go              # DedupEngine — content_hash comparison, upsert logic
    cleanup.go            # OrphanCleaner — removes stale document rows
    types.go              # IndexResult, IndexStats, DocFile values

pkg/
  models/
    document.go           # Document, DocumentKind types
    metadata.go           # Metadata fields (documentation_hash, git_head)
```

### 1.2 Public API (Indexer)

```go
type Indexer struct {
    db     *sql.DB         // SQLite connection
    logger *slog.Logger
    cfg    *IndexerConfig
}

func NewIndexer(db *sql.DB, logger *slog.Logger, cfg *IndexerConfig) *Indexer

// IndexProject indexes a single project. Idempotent, transactional.
func (idx *Indexer) IndexProject(ctx context.Context, projectID uuid.UUID, rootPath string) (*IndexResult, error)

// IndexAll indexes all discovered projects. Wraps IndexProject per project.
func (idx *Indexer) IndexAll(ctx context.Context) (map[uuid.UUID]*IndexResult, error)

// Search performs a full-text search across all indexed documents.
// Supports phrase queries and Boolean operators (AND, OR, NOT).
// Returns ranked results with project context.
func (idx *Indexer) Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error)

type SearchResult struct {
    ID          string       `json:"id"`
    ProjectID   string       `json:"project_id"`
    Path        string       `json:"path"`
    Kind        DocumentKind `json:"kind"`
    Content     string       `json:"content"`
    ContentHash string       `json:"content_hash"`
    IndexedAt   string       `json:"indexed_at"`
    Rank        float64      `json:"rank"`
}
```

### 1.3 Internal Components

#### DocDiscoverer

```
DocDiscoverer
  .FindREADME(rootPath)     → []DocFile (0 or 1)
  .FindDocs(rootPath)       → []DocFile (recursive, .md/.rst/.txt/.adoc)
  .FindADRs(rootPath)       → []DocFile (docs/adr/, .adr/, adr/)
  .FindCHANGELOG(rootPath)  → []DocFile (0-3: CHANGELOG, CHANGES, HISTORY)
```

- Each method returns a slice of `DocFile` values (path relative, absolute path, kind).
- All paths case-insensitive for discovery (readdir + strings.EqualFold).
- Symlinks not followed (lstat, skip if mode&os.ModeSymlink).
- Binary detection: read first 512 bytes, check for null byte.
- `.gitignore` respect: use `git check-ignore` or a `.gitignore`-aware walker for `docs/` scanning.

#### DocReader

```go
type DocReader struct {
    maxFileSize int64 // default 1MB (1048576)
}

func (r *DocReader) Read(path string) (content string, contentHash string, err error)
```

- Reads file, computes SHA-256, truncates at `maxFileSize`.
- Non-UTF8: store raw bytes as BLOB.
- Returns content as string, hex-encoded SHA-256, and any error.

#### DedupEngine

```go
type DedupEngine struct {
    db *sql.DB
}

// Resolve returns action: skip | update | insert
func (d *DedupEngine) Resolve(projectID uuid.UUID, path string, contentHash string) (DedupAction, error)
```

Implements the dedup table from requirements:

| stored.content_hash | action |
|---------------------|--------|
| == file hash | skip |
| != file hash | update content, hash, indexed_at |
| not found | insert new row |

#### FTSManager

```go
type FTSManager struct {
    db *sql.DB
}

// Search queries the FTS5 virtual table (synced automatically via Epic 5 triggers).
// FTS sync is handled by SQLite triggers on the documents table — no manual rebuild needed.
func (f *FTSManager) Search(ctx context.Context, query string, limit, offset int) ([]SearchResult, error)
```

#### OrphanCleaner

```go
type OrphanCleaner struct {
    db *sql.DB
}

// Clean removes document rows for files that no longer exist on disk.
// Accepts the set of currently-valid relative paths; deletes the rest.
func (c *OrphanCleaner) Clean(ctx context.Context, projectID uuid.UUID, validPaths []string) error
```

### 1.4 IndexRunner (Orchestration)

```
IndexRunner.Run(ctx, projectID, rootPath) → *IndexResult

  Sequence:
  1. ResolvePaths(rootPath) → PathSet
  2. docs := discoverer.FindREADME(rootPath)
           + discoverer.FindDocs(rootPath)
           + discoverer.FindADRs(rootPath)
           + discoverer.FindCHANGELOG(rootPath)
  3. BEGIN TRANSACTION
  4. For each doc in docs:
       a. reader.Read(doc.absPath) → content, hash
       b. dedup.Resolve(projectID, doc.relPath, hash) → action
       c. If action == skip: continue
       d. If action == insert: INSERT INTO documents
       e. If action == update: UPDATE documents SET content, content_hash, indexed_at
   5. Cleaner.Clean(ctx, projectID, validPaths)
   // FTS sync is automatic via Epic 5 SQLite triggers on documents table
   6. sortedHashes := sort.Strings(collect all content_hashes from documents for this project)
  8. documentationHash := SHA-256 hex of strings.Join(sortedHashes, "")
  9. UPDATE metadata SET documentation_hash = ?, git_head = ?, last_scan_at = NOW
 10. COMMIT
 11. Return IndexResult{...}
```

---

## 2. Schema Changes (FTS5 Tables)

### 2.1 Existing documents Table (Epic 1.5)

```sql
CREATE TABLE documents (
    id            TEXT PRIMARY KEY,       -- UUID
    project_id    TEXT NOT NULL,          -- FK → projects.id
    path          TEXT NOT NULL,          -- relative to project root
    kind          TEXT NOT NULL,          -- 'README', 'DOC', 'ADR', 'CHANGELOG'
    content       TEXT NOT NULL,
    content_hash  TEXT NOT NULL,          -- SHA-256 hex
    indexed_at    TEXT NOT NULL,          -- ISO 8601 UTC
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, path)
);
```

### 2.2 FTS5 Virtual Table (Defined in Epic 5 Migration)

The FTS5 virtual table, triggers, and index maintenance are owned by **Epic 5 (Knowledge Store)**.
The FTS table uses `content=documents` content-sync mode with SQLite triggers:

```sql
-- Defined in Epic 5 migration (001_initial_schema.up.sql):
CREATE VIRTUAL TABLE documents_fts USING fts5(
    content,
    tokenize='unicode61 remove_diacritics 2',
    content=documents,
    content_rowid=rowid
);

-- Triggers keep FTS in sync on insert, update, delete:
--   AFTER INSERT → INSERT INTO documents_fts
--   AFTER UPDATE → DELETE old + INSERT new
--   AFTER DELETE → DELETE FROM documents_fts
-- (Full trigger definitions in Epic 5 architecture §3.11)
```

**Why not full rebuild?** DedupEngine ensures most index runs are incremental
(few changed files). SQLite triggers provide automatic sync with zero application
logic. Full rebuild on every index is redundant and conflicts with trigger-based sync.

### 2.3 Corruption Recovery

On startup, Epic 5 verifies FTS integrity by comparing row counts and content
hashes. If corruption is detected, it triggers a rebuild. Epic 4 does not
initiate rebuilds — it writes to `documents` and lets triggers handle FTS sync.

### 2.4 Search Query Template

```sql
SELECT
    d.id,
    d.project_id,
    d.path,
    d.kind,
    d.content,
    d.content_hash,
    d.indexed_at,
    bm25(documents_fts) AS rank
FROM documents_fts
JOIN documents d ON documents_fts.rowid = d.rowid
WHERE documents_fts MATCH ?
ORDER BY rank
LIMIT ? OFFSET ?;
```

### 2.5 metadata.documentation_hash Update

```sql
-- After all documents indexed for a project, compute an aggregate hash
-- over all document content_hashes (sorted, joined, SHA-256 of the concatenation)
-- Also store git_head for staleness detection
UPDATE metadata SET
    documentation_hash = ?,
    git_head = ?,
    last_scan_at = ?
WHERE project_id = ?;
```

### 2.6 Indexes

```sql
-- For faster kind-filtered queries
CREATE INDEX IF NOT EXISTS idx_documents_project_kind ON documents(project_id, kind);

-- For orphan cleanup
CREATE INDEX IF NOT EXISTS idx_documents_path ON documents(project_id, path);
```

---

## 3. Sequence Diagrams

### 3.1 IndexProject

```
┌──────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌────────┐    ┌───────────┐
│ Caller   │    │  Indexer     │    │ DocDiscoverer│    │  DocReader   │    │ Dedup  │    │ SQLite   │
│ (MCP/CLI)│    │              │    │              │    │              │    │ Engine │    │          │
└────┬─────┘    └──────┬───────┘    └──────┬───────────┘    └──────┬──────┘    └───┬────┘    └────┬─────┘
     │                 │                    │                       │              │             │
     │ IndexProject    │                    │                       │              │             │
     │ (project, path) │                    │                       │              │             │
     │────────────────▶│                    │                       │              │             │
     │                 │  FindREADME(path)  │                       │              │             │
     │                 │───────────────────▶│                       │              │             │
     │                 │  []DocFile         │                       │              │             │
     │                 │◀───────────────────│                       │              │             │
     │                 │  FindDocs(path)    │                       │              │             │
     │                 │───────────────────▶│                       │              │             │
     │                 │  []DocFile         │                       │              │             │
     │                 │◀───────────────────│                       │              │             │
     │                 │  FindADRs(path)    │                       │              │             │
     │                 │───────────────────▶│                       │              │             │
     │                 │  []DocFile         │                       │              │             │
     │                 │◀───────────────────│                       │              │             │
     │                 │  FindCHANGELOG(path)                      │              │             │
     │                 │───────────────────▶│                       │              │             │
     │                 │  []DocFile         │                       │              │             │
     │                 │◀───────────────────│                       │              │             │
     │                 │                    │                       │              │             │
     │                 │  ── begin tx ──────│───────────────────────│──────────────│────────────▶│
     │                 │                    │                       │              │             │
     │                 │  for each doc:     │                       │              │             │
     │                 │  Read(absPath)     │                       │              │             │
     │                 │───────────────────────────────────────────▶│              │             │
     │                 │  content, hash     │                       │              │             │
     │                 │◀───────────────────────────────────────────│              │             │
     │                 │  Resolve(proj,path,hash)                   │              │             │
     │                 │──────────────────────────────────────────────────────────▶│             │
     │                 │  skip/update/insert                       │              │             │
     │                 │◀──────────────────────────────────────────────────────────│             │
     │                 │  [if update/insert]                        │              │             │
     │                 │  UPSERT documents  │                       │              │             │
     │                 │───────────────────────────────────────────────────────────────────────▶│
     │                 │  ok               │                       │              │             │
     │                 │◀───────────────────────────────────────────────────────────────────────│
     │                 │                    │                       │              │             │
     │                 │  Clean(validPaths) │                       │              │             │
     │                 │───────────────────────────────────────────────────────────────────────▶│
     │                 │  ok               │                       │              │             │
     │                 │◀───────────────────────────────────────────────────────────────────────│
     │                 │                    │                       │              │             │
     │                 │  FTS Rebuild       │                       │              │             │
     │                 │───────────────────────────────────────────────────────────────────────▶│
     │                 │  ok               │                       │              │             │
     │                 │◀───────────────────────────────────────────────────────────────────────│
     │                 │                    │                       │              │             │
     │                 │  Update metadata   │                       │              │             │
     │                 │───────────────────────────────────────────────────────────────────────▶│
     │                 │  ok               │                       │              │             │
     │                 │◀───────────────────────────────────────────────────────────────────────│
     │                 │  ── commit tx ────│───────────────────────│──────────────│────────────▶│
     │                 │                    │                       │              │             │
     │  IndexResult    │                    │                       │              │             │
     │◀────────────────│                    │                       │              │             │
```

### 3.2 Search

```
┌─────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Caller  │    │   Indexer    │    │  FTSManager  │    │   SQLite     │
└────┬────┘    └──────┬───────┘    └──────┬────────┘    └──────┬──────┘
     │                │                    │                    │
     │ Search(q)      │                    │                    │
     │───────────────▶│                    │                    │
     │                │  Search(q,lim,off) │                    │
     │                │───────────────────▶│                    │
     │                │                    │ SELECT FROM        │
     │                │                    │ documents_fts      │
     │                │                    │ JOIN documents      │
     │                │                    │ WHERE MATCH ?       │
     │                │                    │────────────────────▶│
     │                │                    │    ranked results   │
     │                │                    │◀────────────────────│
     │                │  []SearchResult    │                    │
     │                │◀───────────────────│                    │
     │ []SearchResult │                    │                    │
     │◀───────────────│                    │                    │
```

### 3.3 IndexAll (Portfolio-Wide)

```
IndexAll(ctx):
  for each project in projects:
    result, err := idx.IndexProject(ctx, project.ID, project.RootPath)
    collect result
  return map[projectID]result
```

No parallelism in v1 (serial per project to keep memory predictable). Future optimization: goroutine per project with semaphore limit.

---

## 4. Error Handling

### 4.1 Error Types

```go
type IndexError struct {
    Code    string   // machine-readable error code
    Message string   // human-readable description
    Cause   error    // wrapped error (optional)
    Project uuid.UUID // project context (optional)
    Path    string   // file path context (optional)
}
```

### 4.2 Error Code Catalog

| Code | When | Handling |
|------|------|----------|
| `DB_CONNECTION_FAILED` | SQLite open/ping fails | Return error; caller retries |
| `TX_BEGIN_FAILED` | Cannot start transaction | Return error; caller retries |
| `TX_COMMIT_FAILED` | Commit fails after successful work | Log; index may be stale; retry on next call |
| `READ_FAILED` | File read error (permissions, I/O) | Log at error level; skip file; continue |
| `HASH_FAILED` | SHA-256 computation fails | Log; skip file (should not happen) |
| `FTS_BUILD_FAILED` | FTS5 rebuild query fails | Rollback transaction; return error |
| `ORPHAN_CLEAN_FAILED` | DELETE of orphaned rows fails | Log; continue (non-fatal) |
| `MISSING_PROJECT` | project_id not found in projects table | Return error |
| `SEARCH_PARSE_FAILED` | FTS5 query syntax error | Return empty results with error detail |
| `ENCODING_FAILURE` | Cannot decode file as valid UTF-8 | Store raw bytes as BLOB; continue |
| `GITIGNORE_EVAL_FAILED` | git check-ignore call fails | Log warning; index the file anyway (safe but may index more than expected) |

### 4.3 Edge Case Matrix

| Scenario | Detection | Action |
|----------|-----------|--------|
| README >1MB | `stat.Size() > maxFileSize` | Truncate to `maxFileSize` bytes; log debug |
| Binary file in `docs/` | First 512 bytes contain null byte | Skip; log debug with filename |
| Symlink | `lstat().Mode()&os.ModeSymlink != 0` | Do not follow; skip |
| Directory depth >50 | Depth counter exceeds limit | Log warning; skip subtree |
| Non-UTF8 content | `utf8.ValidString()` fails | Store as BLOB; log debug |
| Concurrent index of same project | Mutex per projectID | Block second caller; return result after first completes |
| gitignore eval failure | `exec.Command("git", "check-ignore")` fails | Log warn; index file anyway |
| Deleted file between runs | `row exists but file missing` | `Cleaner.Clean()` deletes orphaned row |
| Changed file between runs | `content_hash != stored hash` | Update content, hash, indexed_at |
| FTS5 not available | `SELECT * FROM pragma_compile_options` lacks ENABLE_FTS5 | Fall back to LIKE scan; log warning |
| docs/ contains 10k+ files | `os.ReadDir` with batch iterator | Process in batches of 100; yield via `runtime.Gosched()` |

---

## 5. Test Strategy

### 5.1 Test Levels

#### Unit Tests (∼80% coverage target for indexer logic)

```
internal/indexer/
  indexer_test.go     — IndexProject happy path, error propagation
  runner_test.go      — orchestration sequencing, tx lifecycle
  reader_test.go      — encoding, truncation, hashing
  discoverer_test.go  — file discovery patterns, missing dirs, case sensitivity
  dedup_test.go       — content_hash comparison, upsert decision matrix
  cleanup_test.go     — orphan detection, dry-run vs actual delete
  fts_test.go         — rebuild, search queries, phrase + boolean operators
```

Test techniques:
- **filesystem fixtures**: `t.TempDir()` with known file layouts for discoverer/reader tests
- **table-driven tests**: for dedup decision matrix (skip/insert/update)
- **golden files**: for expected FTS query results

#### Integration Tests (∼60% coverage)

```
internal/indexer/
  indexer_integration_test.go  — full pipeline: discover → read → dedup → upsert → fts
```

- Use in-memory SQLite (`:memory:`) for fast setup/teardown.
- Pre-populate `projects` and `metadata` tables.
- Run full `IndexProject` against a temp directory fixture.
- Verify `documents` rows, FTS query results, orphan cleanup.

#### Behavior-Driven Tests (Acceptance)

Map acceptance criteria to test cases:

| AC ID | Test Name | Coverage |
|-------|-----------|----------|
| AC-4.1-1 | `TestIndexREADME_ReadmeMD` | discoverer + runner |
| AC-4.1-3 | `TestIndexREADME_CaseInsensitive` | discoverer |
| AC-4.1-5 | `TestIndexREADME_Truncation` | reader |
| AC-4.1-6 | `TestIndexREADME_Idempotent` | dedup |
| AC-4.2-1 | `TestIndexDocs_SupportedFormats` | discoverer |
| AC-4.2-3 | `TestIndexDocs_SkipBinary` | discoverer |
| AC-4.2-5 | `TestIndexDocs_RespectGitignore` | discoverer (with mock exec) |
| AC-4.3-1 | `TestIndexADRs_StandardPaths` | discoverer |
| AC-4.4-1 | `TestIndexCHANGELOG_Variants` | discoverer |
| AC-4.5-1 | `TestFTS_VirtualTableExists` | fts |
| AC-4.5-2 | `TestFTS_PhraseQuery` | fts |
| AC-4.5-3 | `TestFTS_BooleanOperators` | fts |
| AC-4.5-4 | `TestFTS_RankedResults` | fts |
| AC-4.5-5 | `TestFTS_ProjectContext` | fts + integration |
| AC-4.5-6 | `TestFTS_CrossPortfolio` | integration |

### 5.2 Test Fixtures

```
internal/indexer/testdata/
  project-with-readme/
    README.md                      # simple readme
  project-case-insensitive/
    Readme.MD                      # case variant
  project-large-readme/
    README.md                      # >1MB (generated in test setup)
  project-with-docs/
    docs/
      getting-started.md
      api/
        reference.rst
      guide.txt
      changelog.adoc
      binary-file.bin              # binary (null byte at offset 3)
  project-no-readme/
    main.go                        # no readme at all
  project-with-adrs/
    docs/adr/
      001-use-go.md
      002-adopt-sqlite.md
    .adr/
      template.md
    adr/
      other.md
  project-gitignored/
    README.md
    docs/
      internal.md                  # gitignored
    .gitignore                     # internal.md
```

### 5.3 Benchmarks

| Benchmark | File Count | Target |
|-----------|-----------|--------|
| `BenchmarkIndexProject_Small` | 10 files | <500ms hot |
| `BenchmarkIndexProject_Medium` | 100 files | <2s cold |
| `BenchmarkIndexProject_Large` | 500 files | <10s cold |
| `BenchmarkFTS_Search` | 10k documents | <100ms per query |
| `BenchmarkDedup_NoChanges` | 100 files | <200ms (no-op) |

---

## 6. Implementation Order

### Phase 1: Foundation (Days 1-3)

**Story 4.1 — Index README**

1. Define `Document` model in `pkg/models/document.go`:
   ```go
   type DocumentKind string
   const (
       DocKindREADME    DocumentKind = "README"
       DocKindDOC       DocumentKind = "DOC"
       DocKindADR       DocumentKind = "ADR"
       DocKindCHANGELOG DocumentKind = "CHANGELOG"
   )
   type Document struct {
       ID          string       `json:"id"`
       ProjectID   string       `json:"project_id"`
       Path        string       `json:"path"`
       Kind        DocumentKind `json:"kind"`
       Content     string       `json:"content"`
       ContentHash string       `json:"content_hash"`
       IndexedAt   string       `json:"indexed_at"`
   }
   ```
2. Implement `DocDiscoverer.FindREADME()` — case-insensitive glob in repo root.
3. Implement `DocReader.Read()` — read, hash, truncate at 1MB.
4. Implement `DedupEngine` — query content_hash, return action.
5. Implement `IndexProject` with README-only pipeline.
6. **Tests:** discoverer, reader, dedup, README-only integration.

### Phase 2: Directory Scanning (Days 4-7)

**Story 4.2 — Index docs/**

1. Implement `DocDiscoverer.FindDocs()` — recursive `filepath.WalkDir`, filtered by extension.
2. Binary detection (first 512 bytes, null-byte check).
3. Depth limiter (max 50).
4. `.gitignore` respect: shell out to `git check-ignore` or embed `go-gitignore` library.
5. Implement `OrphanCleaner`.
6. Wire docs/ into `IndexRunner`.
7. **Tests:** docs/ discovery, binary skip, gitignore, orphan cleanup.

### Phase 3: ADRs + CHANGELOG (Days 8-10)

**Stories 4.3 + 4.4**

1. `DocDiscoverer.FindADRs()` — check three directories, accept NNN-*.md and *.md.
2. `DocDiscoverer.FindCHANGELOG()` — case-insensitive for known names.
3. Wire both into `IndexRunner`.
4. **Tests:** ADR paths, CHANGELOG variants.

### Phase 4: Full-Text Search (Days 11-14)

**Story 4.5 — FTS5 Indexing**

1. Implement `FTSManager.Rebuild()` — transactional rebuild from `documents`.
2. Implement `FTSManager.Search()` — MATCH with ranked results.
3. FTS5 availability check at startup (compile_options).
4. Query sanitization (FTS5 syntax errors → safe fallback).
5. Wire FTS rebuild into `IndexRunner` (after dedup + cleanup).
6. **Tests:** phrase queries, Boolean operators, ranking, cross-portfolio.
7. **Benchmarks:** search latency, rebuild time.

### Phase 5: Polish (Days 15)

1. `IndexAll()` — iterate all projects, collect results.
2. Metadata update (documentation_hash computation, git_head capture).
3. Per-project mutex for concurrency safety.
4. Error wrapping (`IndexError` types).
5. Logging (structured, slog, per-operation).
6. Full integration test suite.
7. Documentation (update ADRs if needed).

---

## 7. Configuration

Add to `models.Config`:

```go
type IndexerConfig struct {
    MaxFileSize    int  `toml:"max_file_size"`    // default 1048576 (1MB)
    MaxDocsPerBatch int `toml:"max_docs_per_batch"` // default 100, for docs/ scanning
    MaxDepth       int  `toml:"max_depth"`         // default 50
    SearchMaxResults int `toml:"search_max_results"` // default 200
}
```

---

## 8. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| FTS sync strategy | SQLite triggers (Epic 5) | Automatic sync on document insert/update/delete; no application logic; Epic 5 verifies integrity on startup |
| Tokenizer | `unicode61 remove_diacritics 2` | Good baseline for English + code; no Porter stemmer (code terms are not natural language) |
| Parallelism | Serial per project | Predictable memory; parallelism adds complexity; can be added later with semaphore |
| Binary detection | Null-byte in first 512 bytes | Simple, fast, no external dependency; covers 99% of cases |
| gitignore respect | `git check-ignore` via exec | Matches actual git behavior; falls back to index-all on failure |
| Per-file limit | 1MB truncation | Prevents OOM; stream is overkill for 99% of docs; truncation is deterministic |
| documentation_hash | SHA-256 of sorted concatenated content hashes | Deterministic aggregate; fast to compute; single field to compare |
| Orphan cleanup | Full diff after indexing | Guarantees consistency; document-level granularity |
