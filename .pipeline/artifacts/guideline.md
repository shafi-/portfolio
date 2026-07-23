# Epic 4 — Documentation Indexing: Implementation Guideline

**Reference:** `.architecture/epic-04-architecture.md`, `.requirements/epic-04-requirements.md`, `docs/tasks/epic-04-documentation-indexing.md`

---

## 1. Technical Standards

### Language & Runtime
- Go 1.22.6, module `github.com/nerddevsltd/portfolio`
- SQLite FTS5 for full-text search (via `modernc.org/sqlite` or crawshaw)
- File discovery via `filepath.Walk` with ignore patterns

### Package Organization
```
internal/indexer/
├── readme.go        — README discovery and indexing
├── docs_dir.go      — docs/ directory recursion
├── adr.go           — ADR discovery
├── changelog.go     — CHANGELOG discovery
├── search.go        — FTS5 index building and query
└── types.go         — Document types, constants
```

### Code Conventions
- **Document kind enum:** README, DOC, ADR, CHANGELOG (string constants in types.go)
- **Content storage:** Store full content in `documents.content`. Large files (>1MB) truncated with warning.
- **Hashing:** SHA-256 of content for change detection.
- **FTS5:** Create FTS virtual table that syncs with documents table via triggers.
- **Error handling:** Missing README/docs/ADRs is not an error — log at DEBUG.

### Key Design Decisions
- README detection: case-insensitive basename matching (`README.md`, `README.rst`, `README.txt`, `README`)
- docs/ recursion: respect `.gitignore` patterns, skip binary files
- ADR directories: `docs/adr/`, `.adr/`, `adr/` — recognize `NNN-*.md` patterns
- CHANGELOG: `CHANGELOG.md`, `CHANGES.md`, `HISTORY.md` (case-insensitive)
- FTS5: SQLite built-in, no external search engine needed

## 2. Story Implementation Order

### Story 4.1 — Index README
**Files to create:** `internal/indexer/readme.go`

Find README files at repo root (not recursive). Store in documents table with kind="README".

### Story 4.2 — Index docs/ Directory
**Files to create:** `internal/indexer/docs_dir.go`

Recursively scan `docs/` directory. Supported formats: `.md`, `.rst`, `.txt`, `.adoc`. Skip binary files (check MIME or first bytes). Store kind="DOC".

### Story 4.3 — Index ADRs
**Files to create:** `internal/indexer/adr.go`

Scan well-known ADR directories. Store kind="ADR".

### Story 4.4 — Index CHANGELOG
**Files to create:** `internal/indexer/changelog.go`

Find CHANGELOG files at repo root. Store kind="CHANGELOG".

### Story 4.5 — Full-Text Search Indexing
**Files to create:** `internal/indexer/search.go`

Create FTS5 virtual table:
```sql
CREATE VIRTUAL TABLE documents_fts USING fts5(
    content,
    content=documents,
    content_rowid=rowid
);
```

Create triggers to keep FTS in sync with documents table. Support phrase queries and Boolean operators. Return ranked results.

## 3. Database Schema

```sql
CREATE TABLE documents (
    id           TEXT PRIMARY KEY,
    project_id   TEXT NOT NULL REFERENCES projects(id),
    path         TEXT NOT NULL,
    kind         TEXT NOT NULL,  -- README, DOC, ADR, CHANGELOG
    content      TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    indexed_at   TEXT NOT NULL,
    UNIQUE(project_id, path)
);
```

## 4. Testing Strategy

### Unit Tests
- README: various naming conventions, missing README, very large README
- docs/: nested directories, binary files, no docs/ directory
- ADRs: naming patterns NNN-*, various directories
- CHANGELOG: different filenames, missing changelog
- FTS5: insert documents, search with phrases, verify ranking

### Integration Tests
- Full indexing on a project with all document types
- Re-indexing detects changed content via hash
- Search across multiple projects
- Edge case: empty project (no documents)

## 5. Build & Verification

```bash
go build ./cmd/portfolio
go test ./internal/indexer/... -v -cover
```

## 6. Quality Gates

- [ ] All tests pass: `go test ./...`
- [ ] FTS5 search returns ranked results <500ms for 1000+ documents
- [ ] Indexing handles 1MB+ READMEs without OOM
- [ ] Binary files in docs/ are skipped (not indexed)
- [ ] Missing documentation is never an error
- [ ] All acceptance criteria from `.requirements/epic-04-requirements.md` pass
