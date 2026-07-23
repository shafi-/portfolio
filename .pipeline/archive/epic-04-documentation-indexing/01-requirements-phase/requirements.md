# Epic 4 — Documentation Indexing: Requirements

**Milestone:** 1 — Core Engine
**Status:** Draft
**Version:** 0.1

---

## Feature Overview

Documentation Indexing is the fourth capability of the Portfolio Engine. It extracts and stores project documentation files into the local SQLite knowledge store as searchable documents, without performing any semantic interpretation. The engine discovers, reads, hashes, and persists documentation content from known locations within each discovered project. A full-text search index (SQLite FTS5) is built over all indexed content to enable cross-portfolio search.

This is a purely deterministic operation: the same project at the same git HEAD produces the same index. No AI, no heuristics, no interpretation.

**Lifecycle placement:** After metadata extraction (Epic 2.2) and database layer (Epic 1.5) are complete. Before search (Epic 5), HTTP API (Epic 6), and MCP (Epic 7).

---

## Requirements

### Functional Requirements

| ID | Requirement | Story |
|----|------------|-------|
| FR-1 | Engine shall find and read the project's README file. Must detect README.md, README.rst, README.txt, README (no extension), case-insensitive in both stem and extension (e.g., readme.md, README.TXT, Readme.rst, README all match). | 4.1 |
| FR-2 | Engine shall store each README as a record in the `documents` table with fields: `project_id`, `path` (relative to project root), `kind="README"`, `content`, `content_hash` (SHA-256 of content), `indexed_at` (UTC timestamp). | 4.1 |
| FR-3 | Engine shall truncate or stream README content exceeding 1MB; must not fail or OOM on large files. | 4.1 |
| FR-3.1 | Engine shall apply a 1MB per-file size limit to all indexed documents (DOC, ADR, CHANGELOG). Files exceeding this limit shall be truncated at the 1MB boundary. | All |
| FR-4 | Engine shall recursively scan the `docs/` directory (relative to project root) for documentation files. | 4.2 |
| FR-5 | Engine shall index files with extensions: `.md`, `.rst`, `.txt`, `.adoc`. Kind shall be `"DOC"`. | 4.2 |
| FR-6 | Engine shall detect and skip binary files during docs/ scanning (identify by MIME type or null-byte detection). | 4.2 |
| FR-7 | Engine shall find ADRs in standard locations: `docs/adr/`, `.adr/`, `adr/` (relative to project root). | 4.3 |
| FR-8 | Engine shall recognize ADR files by naming patterns: `NNN-*.md` or any `*.md`. Kind shall be `"ADR"`. | 4.3 |
| FR-9 | Engine shall find changelog/history files: CHANGELOG.md, CHANGES.md, HISTORY.md (case-insensitive). Kind shall be `"CHANGELOG"`. | 4.4 |
| FR-10 | Engine shall create and maintain a SQLite FTS5 virtual table over `documents.content`. | 4.5 |
| FR-10.1 | FTS5 tokenizer shall be `unicode61` (baseline). The tokenizer choice must be documented and consistent across all rebuilds. | 4.5 |
| FR-10.2 | FTS index shall be rebuilt from scratch (delete and repopulate) after each indexing operation. Incremental sync is NOT supported — full rebuild ensures consistency. | 4.5 |
| FR-11 | Engine shall support phrase queries and Boolean operators (AND, OR, NOT) in FTS search. | 4.5 |
| FR-12 | Engine shall return ranked search results with project context (project name, path, document kind). | 4.5 |
| FR-12.1 | Search results shall be limited to 50 results by default, with a configurable maximum of 200. Pagination via offset parameter shall be supported. | 4.5 |
| FR-13 | Engine shall support cross-portfolio search across all indexed documents. | 4.5 |
| FR-14 | Engine shall use `content_hash` as deduplication key; re-indexing the same content shall be idempotent. | All |
| FR-15 | Engine shall update `indexed_at` when content_hash changes (document updated). | All |
| FR-16 | Engine shall update `metadata.documentation_hash` on the project after indexing completes. | All |
| FR-17 | Engine shall respect `.gitignore` — files matching gitignore patterns should not be indexed when scanning `docs/`. This does NOT apply to explicitly targeted locations (README root, ADR directories, CHANGELOG root) — those are scanned regardless of gitignore. | 4.2 |
| FR-18 | Engine shall store the project's current `git_head` (commit SHA) in the metadata table after successful indexing, enabling staleness detection without re-hashing. | All |

### Non-Functional Requirements

| ID | Requirement |
|----|------------|
| NFR-1 | Indexing a project with <100 files shall complete in <2s (cold) / <500ms (hot, no changes). |
| NFR-2 | Memory usage shall not exceed 100MB for any single project indexing operation. |
| NFR-3 | All indexing operations shall be deterministic — same project at same git HEAD (stored via FR-18) produces identical `content_hash` values for all documents. |
| NFR-4 | Engine shall never modify files in the repository; indexing is read-only. |
| NFR-5 | Indexing shall be interrupt-safe — partially written index state must not corrupt the database. Use transactions. |
| NFR-6 | FTS5 index size must not exceed 2x the raw content size. |
| NFR-7 | Search queries must return results in <500ms for a portfolio of up to 500 projects. |
| NFR-8 | All indexing operations must be idempotent; running index twice on an unchanged project is a no-op. |

---

## Edge Cases & Error Handling

| Edge Case | Handling |
|-----------|----------|
| No README file exists | Skip silently; no error, no warning. |
| README >1MB | Truncate content at 1MB boundary; log a debug message. |
| README with non-standard extension | Only the specified patterns are matched; no heuristic guessing. Case-insensitive matching applies to both the stem and extension (e.g., README.TXT matches via readme.txt → TXT matches txt). |
| `docs/` directory does not exist | Skip silently; no error. |
| Binary file in `docs/` | Detect and skip; log at debug level with file path. |
| Symlink in `docs/` | Do not follow symlinks (prevent infinite loops and out-of-repo access). |
| Nested directory >10 levels deep in `docs/` | Honor recursion but set a max depth of 50 to prevent stack overflow. |
| No ADR directory exists | Skip silently; no error. |
| ADR file without NNN- prefix | Still index with kind="ADR" as long as it's .md in an ADR directory. |
| No CHANGELOG file exists | Skip silently; no error. |
| Multiple changelog files found | Index all matching files (rare but valid). |
| File changes between index runs | `content_hash` differs; update content and `indexed_at`. E.g., re-index after `git pull`. |
| Database transaction failure during indexing | Rollback; error propagates to caller; no partial state persisted. |
| FTS5 index out of sync with documents table | On re-index, truncate FTS table and rebuild from documents table. |
| Repository path changes between runs | Project is re-discovered; documents re-keyed by new `project_id` or matched via `root_path`. |
| File is deleted between index runs | Orphaned document record; clean up on next index (delete from documents where file no longer exists). |
| Unicode/non-UTF8 file content | Read as bytes; store as text; if encoding detection fails, store raw bytes as BLOB. |
| Extremely large docs/ directory (>10,000 files) | Process in batches of 100; yield between batches to avoid blocking the event loop. |
| Concurrency: two index operations on same project | Serialize per-project via mutex/lock; second caller waits or returns "already indexing". |

---

## Acceptance Criteria

Derived from epic stories, mapped to verifiable outcomes.

### Story 4.1 — Index README

| AC ID | Criterion | Verification |
|-------|-----------|--------------|
| AC-4.1-1 | Given a project with README.md, when indexed, a document record exists with kind="README". | Query `documents` table; expect 1 row with kind="README". |
| AC-4.1-2 | Given a project with README.rst, when indexed, a document record exists with kind="README". | Same as above; file name differs. |
| AC-4.1-3 | Given a project with a README file of any case variant (readme.md, README.TXT), when indexed, the document is found. | Case-insensitive match confirmed. |
| AC-4.1-4 | Given a project with no README file, when indexed, no error is raised and 0 README documents are created. | Error-free run; 0 rows with kind="README". |
| AC-4.1-5 | Given a README file >1MB, when indexed, content is truncated at 1MB boundary without error. | Content length <= 1,048,576 bytes. |
| AC-4.1-6 | Given a previously indexed README with unchanged content, when re-indexed, `indexed_at` is not updated (idempotent). | Timestamps match. |

### Story 4.2 — Index docs/ Directory

| AC ID | Criterion | Verification |
|-------|-----------|--------------|
| AC-4.2-1 | Given a project with docs/ containing .md, .rst, .txt, .adoc files, when indexed, all are stored with kind="DOC". | Query matching rows. |
| AC-4.2-2 | Given a project with nested subdirectories in docs/, when indexed, files in subdirectories are recursively indexed. | Files from subdirectories present in documents. |
| AC-4.2-3 | Given a project with a binary file in docs/, when indexed, the binary file is skipped. | No document record for the binary file. |
| AC-4.2-4 | Given a project with no docs/ directory, when indexed, no error is raised. | Error-free run. |
| AC-4.2-5 | Given a project with .gitignore'd files in docs/, when indexed, ignored files are excluded. | Ignored files absent from documents. |

### Story 4.3 — Index ADRs

| AC ID | Criterion | Verification |
|-------|-----------|--------------|
| AC-4.3-1 | Given ADRs in docs/adr/, when indexed, they are stored with kind="ADR". | Rows with kind="ADR". |
| AC-4.3-2 | Given ADRs in .adr/ or adr/, when indexed, they are stored with kind="ADR". | Same check for alternative paths. |
| AC-4.3-3 | Given ADRs named 001-decision.md and plain.md, both are indexed. | Both patterns accepted. |
| AC-4.3-4 | Given a project with no ADR directory, when indexed, no error is raised. | Error-free run. |

### Story 4.4 — Index CHANGELOG

| AC ID | Criterion | Verification |
|-------|-----------|--------------|
| AC-4.4-1 | Given CHANGELOG.md, CHANGES.md, or HISTORY.md, when indexed, they are stored with kind="CHANGELOG". | Rows with kind="CHANGELOG". |
| AC-4.4-2 | Given a project with no changelog file, when indexed, no error is raised. | Error-free run. |

### Story 4.5 — Full-Text Search Indexing

| AC ID | Criterion | Verification |
|-------|-----------|--------------|
| AC-4.5-1 | Given indexed documents, an FTS5 virtual table exists on `documents.content`. | `SELECT count(*) FROM documents_fts` succeeds. |
| AC-4.5-2 | Given a phrase query ("error handling"), results match the exact phrase. | Results contain documents with the phrase. |
| AC-4.5-3 | Given Boolean operators (AND, OR, NOT), results respect the boolean logic. | Query "foo AND bar" returns intersection of matches. |
| AC-4.5-4 | Given a search query, results are ranked by relevance (BM25 or equivalent). | Results ordered by rank score descending. |
| AC-4.5-5 | Given a search query, results include project name, document path, and kind. | Result set includes joined project context. |
| AC-4.5-6 | Given documents from multiple projects, search returns cross-portfolio results. | Results include multiple project_ids. |
| AC-4.5-7 | Given a search with zero matching results, empty result set is returned (no error). | Empty array, not an error. |

---

## Data Flow

```
┌──────────────────────────────────────────────────────────┐
│                     Portfolio Engine                      │
│                                                           │
│  ┌──────────────┐    ┌──────────────────┐                 │
│  │ Discovered   │    │  IndexRunner      │                 │
│  │ Project      │───▶│  (per project)    │                 │
│  │ (root_path)  │    │                   │                 │
│  └──────────────┘    │  1. Read README   │                 │
│                      │  2. Scan docs/    │                 │
│                      │  3. Find ADRs     │                 │
│                      │  4. Find CHANGELOG│                 │
│                      │  5. Hash content  │                 │
│                      │  6. Upsert into   │                 │
│                      │     documents table│                │
│                      │  7. Rebuild FTS   │                 │
│                      └────────┬──────────┘                 │
│                               │                            │
│                               ▼                            │
│  ┌──────────────────────────────────────────────────┐      │
│  │                SQLite Knowledge Store             │      │
│  │                                                    │      │
│  │  ┌─────────────┐   ┌─────────────────────────┐    │      │
│  │  │ documents   │   │ documents_fts (FTS5)     │    │      │
│  │  ├─────────────┤   ├─────────────────────────┤    │      │
│  │  │ id          │   │ content MATCH ?          │    │      │
│  │  │ project_id  │   │ rank                     │    │      │
│  │  │ path        │   └─────────────────────────┘    │      │
│  │  │ kind        │                                  │      │
│  │  │ content     │   ┌─────────────────────────┐    │      │
│  │  │ content_hash│   │ metadata                │    │      │
│  │  │ indexed_at  │   │ ─────────────────────   │    │      │
│  │  └─────────────┘   │ documentation_hash      │    │      │
│  │                    │ last_scan_at             │    │      │
│  │                    └─────────────────────────┘    │      │
│  └──────────────────────────────────────────────────┘      │
│                                                           │
└──────────────────────────────────────────────────────────┘
         │
         ▼
┌────────────────────┐   ┌──────────────────────────────┐
│ searchDocumentation │   │ GET /projects/{id}/documents  │
│ (MCP tool)          │   │ GET /search?q= (HTTP API)    │
│                     │   │                              │
│ Query → FTS5 →      │   │ Query → FTS5 → ranked results│
│ ranked results      │   │ with project context         │
└────────────────────┘   └──────────────────────────────┘
```

### Step-by-Step Data Flow

1. **Trigger:** Engine receives a request to index a project (via MCP tool `indexDocumentation`, or as part of `discoverProjects` post-processing).
2. **Resolve paths:** Relative to `project.root_path`, compute candidate paths for README, docs/, adr/, CHANGELOG.
3. **README:** Check existence of README variants (case-insensitive). Read file, compute SHA-256 hash. If hash differs from stored `content_hash`, upsert into `documents`. If content >1MB, truncate before hashing/storing.
4. **docs/:** List directory contents recursively. Filter by extension. Skip binary files via MIME/null-byte check. Skip gitignored files. For each file: read, hash, upsert.
5. **ADRs:** For each known ADR directory (docs/adr/, .adr/, adr/): list *.md files, hash, upsert with kind="ADR".
6. **CHANGELOG:** Check for CHANGELOG.md, CHANGES.md, HISTORY.md. Hash, upsert with kind="CHANGELOG".
7. **Cleanup:** Delete `documents` rows for files that no longer exist on disk (comparing indexed paths against current filesystem state).
8. **FTS rebuild:** Delete and repopulate `documents_fts` from `documents` (full rebuild only — no incremental sync). Transactional.
9. **Metadata update:** Set `metadata.documentation_hash` to a combined hash of all document hashes. Set `metadata.last_scan_at` to now.
10. **Return:** Summary of indexed documents (counts per kind, total bytes, duration).

### Deduplication Logic

```
if stored.content_hash == file.content_hash:
    skip (no change)
elif stored.content_hash != file.content_hash:
    update content, content_hash, indexed_at
else (no stored row):
    insert new row
```

---

## Dependencies

### Internal Dependencies

| Dependency | Epic/Story | Why |
|------------|-----------|-----|
| Database schema (documents table, metadata table) | Epic 1.5 | Tables must exist before indexing can write. |
| Project discovery (project_id, root_path) | Epic 2.1 | Indexing operates per discovered project. |
| Metadata extraction (metadata row exists, git tracking) | Epic 3 | Indexing updates metadata.documentation_hash; git HEAD used for staleness detection. |
| Search API (searchDocumentation MCP tool) | Epic 5 | Consumes the FTS5 index built in 4.5. |
| HTTP API (GET /projects/{id}/documents, GET /search) | Epic 6 | Exposes indexed documents to dashboard and external consumers. |

### External Dependencies

| Dependency | Version | Why |
|------------|---------|-----|
| Go standard library `os`, `path/filepath`, `crypto/sha256` | stdlib | File system traversal, path handling, content hashing. |
| SQLite via Go driver (e.g., `modernc.org/sqlite` or `mattn/go-sqlite3`) | latest stable | FTS5 virtual table support required. |
| `mattn/go-sqlite3` FTS5 enabled build tag | — | FTS5 is not always compiled in by default; must verify build tag. |

### Implementation Order Within Epic

```
4.1 Index README ──▶ 4.2 Index docs/ ──▶ 4.3 Index ADRs
      │                                    (blocked by 4.2)
      └──▶ 4.4 Index CHANGELOG
             │
             ▼
          4.5 Full-Text Search Indexing
             (blocked by 4.1-4.4)
```

Stories 4.1 and 4.4 can be implemented after 4.1's README scanning framework is built, since both are single-file patterns. Story 4.3 requires the recursive directory traversal from 4.2. Story 4.5 requires all document kinds to be indexed first.

### Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| FTS5 not available in SQLite build | Cannot build search index | Use Go FTS5 driver or fall back to LIKE search. Detect at startup. |
| Very large monorepo with thousands of docs | Slow indexing, memory pressure | Batch processing; streaming reads; configurable max file count per project. |
| Encoded/non-UTF8 text files | Content corruption in DB | Detect encoding; store as BLOB if not valid UTF-8; skip truly binary files. |
| Symlink loops in docs/ | Infinite recursion | Max depth limit (50); do not follow symlinks. |
| Race condition: repo modified during indexing | Stale or inconsistent index | Use git HEAD snapshot; document that index reflects a point-in-time state. |
