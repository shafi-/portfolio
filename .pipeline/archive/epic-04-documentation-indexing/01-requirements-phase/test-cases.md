# Epic 4 — Documentation Indexing: Test Cases

**Version:** 1.0
**Milestone:** 1 — Core Engine
**Coverage:** Stories 4.1–4.5

---

## Story 4.1 — Index README

### TC-4.1-1: Index README.md (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-1 |
| **Title** | Index a standard README.md file |
| **Description** | Verify that a project with a README.md at its root is indexed with kind="README" and all required fields are populated. |
| **Preconditions** | A discovered project exists with a README.md file containing text content. Database schema for `documents` table exists. |
| **Steps** | 1. Call `indexDocumentation(project_id)` for the project. 2. Query the `documents` table for kind="README". 3. Verify the returned record fields. |
| **Expected Result** | A single row with: `kind="README"`, `path="README.md"`, non-empty `content`, non-null `content_hash` (SHA-256 hex), non-null `indexed_at` (UTC ISO-8601), valid `project_id`. |
| **Story** | 4.1 |

### TC-4.1-2: Index README.rst (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-2 |
| **Title** | Index a README.rst file |
| **Description** | Verify that a README.rst file is detected and indexed correctly. |
| **Preconditions** | Project has README.rst (no README.md). |
| **Steps** | 1. Index the project. 2. Query documents for kind="README". |
| **Expected Result** | One row with `kind="README"`, `path="README.rst"`. |
| **Story** | 4.1 |

### TC-4.1-3: Index README.txt (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-3 |
| **Title** | Index a README.txt file |
| **Description** | Verify that a plain-text README.txt is indexed. |
| **Preconditions** | Project has README.txt (only). |
| **Steps** | 1. Index the project. 2. Query documents for kind="README". |
| **Expected Result** | One row with `kind="README"`, `path="README.txt"`. |
| **Story** | 4.1 |

### TC-4.1-4: Index Case-Insensitive README Variants

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-4 |
| **Title** | Index case-insensitive readme file variants |
| **Description** | Verify that readme.md, README.TXT, Readme.rst, etc. are all detected regardless of case. |
| **Preconditions** | Project has `readme.md` (lowercase). |
| **Steps** | 1. Index the project. 2. Query documents for kind="README". |
| **Expected Result** | README is found and indexed. Run again with `README.TXT` — file is found. Run again with `Readme.rst` — file is found. |
| **Story** | 4.1 |

### TC-4.1-5: No README File Exists

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-5 |
| **Title** | Project without any README file |
| **Description** | Verify that a project with no README variant produces no error and zero README documents. |
| **Preconditions** | Project exists with no README.* or readme.* files. |
| **Steps** | 1. Index the project. 2. Query documents for kind="README". 3. Check error return value. |
| **Expected Result** | Zero rows with kind="README". No error raised. Indexing completes successfully. |
| **Story** | 4.1 |

### TC-4.1-6: README Larger Than 1MB

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-6 |
| **Title** | Truncate README content exceeding 1MB |
| **Description** | Verify that a README file larger than 1MB is truncated at the 1MB boundary without error or crash. |
| **Preconditions** | Project has a README.md file >1,048,576 bytes (e.g., generated with large repeated content). |
| **Steps** | 1. Index the project. 2. Query the README document's content and content_hash. |
| **Expected Result** | `len(content) <= 1,048,576`. Content is the first 1MB of the file. No panic or OOM. A debug-level log message is emitted about truncation. |
| **Story** | 4.1 |

### TC-4.1-7: Idempotent Re-index of Unchanged README

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-7 |
| **Title** | Re-indexing unchanged README is a no-op |
| **Description** | Verify that indexing the same project twice with an unchanged README does not update `indexed_at`. |
| **Preconditions** | Project has been indexed once. README content is unchanged. |
| **Steps** | 1. Record the `indexed_at` and `content_hash` from the first index. 2. Wait 1 second (clock resolution). 3. Call `indexDocumentation` again. 4. Query the README row. |
| **Expected Result** | `content_hash` matches. `indexed_at` is unchanged (same timestamp). No new row is inserted. |
| **Story** | 4.1 |

### TC-4.1-8: README Updated Between Index Runs

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-8 |
| **Title** | Re-index detects changed README content |
| **Description** | Verify that modifying the README between index runs updates content, content_hash, and indexed_at. |
| **Preconditions** | Project was indexed with README content V1. README is now modified to content V2. |
| **Steps** | 1. Record the old `content_hash` and `indexed_at`. 2. Re-index the project. 3. Query the README row. |
| **Expected Result** | `content_hash` differs from V1. `content` matches V2. `indexed_at` is newer than V1 timestamp. Exactly one row remains (no duplicate). |
| **Story** | 4.1 |

### TC-4.1-9: README with Non-Standard Extension

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-9 |
| **Title** | README with non-standard extension is ignored |
| **Description** | Verify that a README file with an extension not in the supported list (e.g., README.org, README.html) is not indexed. |
| **Preconditions** | Project contains README.org (no supported README variant present). |
| **Steps** | 1. Index the project. 2. Query documents for kind="README". |
| **Expected Result** | Zero rows with kind="README". No error raised. |
| **Story** | 4.1 |

### TC-4.1-10: README is Empty File

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-10 |
| **Title** | Index an empty README file |
| **Description** | Verify that an empty README.md (0 bytes) is indexed without error. |
| **Preconditions** | Project has an empty README.md (0 bytes). |
| **Steps** | 1. Index the project. 2. Query the README row. |
| **Expected Result** | Row exists with `kind="README"`, `content=""`, `content_hash` is the SHA-256 of an empty string (`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`). |
| **Story** | 4.1 |

### TC-4.1-11: README is a Symlink (Security)

| Field | Value |
|-------|-------|
| **ID** | TC-4.1-11 |
| **Title** | README that is a symlink is not followed |
| **Description** | Verify that if README.md is a symlink pointing outside the project, it is not followed (read-only constraint). |
| **Preconditions** | Project has README.md as a symlink to `/etc/passwd` or another out-of-repo file. |
| **Steps** | 1. Index the project. 2. Check document records. |
| **Expected Result** | README is not indexed or the content is from the symlink target after verifying it's within the repo. If outside repo, skip. No security breach. |
| **Story** | 4.1 |

---

## Story 4.2 — Index docs/ Directory

### TC-4.2-1: Index All Supported File Types in docs/

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-1 |
| **Title** | Index .md, .rst, .txt, .adoc files in docs/ |
| **Description** | Verify that files with each supported extension in the docs/ directory are indexed with kind="DOC". |
| **Preconditions** | Project has `docs/` containing: guide.md, manual.rst, notes.txt, specification.adoc. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | Four rows with kind="DOC". Each has correct `path` (relative to project root), content, and content_hash. |
| **Story** | 4.2 |

### TC-4.2-2: Recursive Subdirectory Indexing

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-2 |
| **Title** | Recursively index files in nested subdirectories |
| **Description** | Verify that files inside subdirectories of docs/ are indexed recursively. |
| **Preconditions** | Project has `docs/` with structure: `docs/getting-started/install.md`, `docs/guides/advanced/deployment.md`, `docs/api/reference/v2/changelog.md`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | All 3 files are indexed with correct relative paths. Paths include subdirectory structure (e.g., `docs/getting-started/install.md`). |
| **Story** | 4.2 |

### TC-4.2-3: Skip Binary Files in docs/

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-3 |
| **Title** | Binary files in docs/ are skipped |
| **Description** | Verify that binary files (e.g., .png, .ico, .pdf) are detected and skipped during indexing. |
| **Preconditions** | Project has `docs/` with: readme.md (text), logo.png (binary), diagram.pdf (binary), script.bin (binary). |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". 3. Check logs for skip messages. |
| **Expected Result** | Only readme.md is indexed. Binary files have no document rows. Debug-level log entries exist for each skipped binary file with the file path. |
| **Story** | 4.2 |

### TC-4.2-4: No docs/ Directory Exists

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-4 |
| **Title** | Project without docs/ directory |
| **Description** | Verify that absence of docs/ is handled gracefully as a no-op. |
| **Preconditions** | Project has no `docs/` directory. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". 3. Check error return. |
| **Expected Result** | Zero rows with kind="DOC". No error raised. Indexing completes successfully. |
| **Story** | 4.2 |

### TC-4.2-5: gitignored Files in docs/ Are Excluded

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-5 |
| **Title** | Respect .gitignore patterns for docs/ files |
| **Description** | Verify that files matching .gitignore patterns in docs/ are excluded from indexing. |
| **Preconditions** | Project has `docs/` with: `api.md` and `internal.md`. `.gitignore` contains `docs/internal.md`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | Only `api.md` is indexed. `internal.md` is excluded. |
| **Story** | 4.2 |

### TC-4.2-6: Files with Unsupported Extensions Are Skipped

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-6 |
| **Title** | Skip files with unsupported extensions in docs/ |
| **Description** | Verify that files with extensions not in .md, .rst, .txt, .adoc are not indexed from docs/. |
| **Preconditions** | Project has `docs/` with: readme.md, notes.txt, image.png, style.css, script.js, data.json. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | Only `readme.md` and `notes.txt` are indexed. .png, .css, .js, .json are skipped. |
| **Story** | 4.2 |

### TC-4.2-7: Symlink in docs/ Is Not Followed

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-7 |
| **Title** | Symlinks inside docs/ are not followed |
| **Description** | Verify that symbolic links in docs/ are not followed to prevent infinite loops and out-of-repo access. |
| **Preconditions** | Project has `docs/` containing: `guide.md`, a symlink `external.md -> ../../outside-repo/secret.md`, and a symlink `loop/ -> ../docs/` (self-referential). |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | Only `guide.md` is indexed. Symlink targets are not indexed regardless of whether they point in or out of the repo. |
| **Story** | 4.2 |

### TC-4.2-8: Very Deep Directory Nesting

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-8 |
| **Title** | Max recursion depth of 50 levels |
| **Description** | Verify that recursion into docs/ subdirectories is limited to a max depth of 50 levels to prevent stack overflow. |
| **Preconditions** | Project has `docs/` with nested subdirectories >50 levels deep (e.g., `docs/a/b/c/...`). A file exists at depth 51. A file exists at depth 49. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". |
| **Expected Result** | The file at depth 49 is indexed. The file at depth 51 is skipped. No panic/stack overflow occurs. |
| **Story** | 4.2 |

### TC-4.2-9: Binary File Detection via Null-Byte

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-9 |
| **Title** | Binary files detected via null-byte or MIME |
| **Description** | Verify that files containing null bytes are correctly detected as binary and skipped. |
| **Preconditions** | Project has `docs/` with a text file containing a null byte (e.g., corrupted text), and a genuine binary file. |
| **Steps** | 1. Index the project. 2. Check which files are indexed. |
| **Expected Result** | Files detected as binary (via null-byte check or MIME detection) are skipped. |
| **Story** | 4.2 |

### TC-4.2-10: docs/ Directory with Only Binary Files

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-10 |
| **Title** | docs/ with only binary content |
| **Description** | Verify that a docs/ directory containing only binary/skipped files produces zero DOC records without error. |
| **Preconditions** | Project has `docs/` containing only: image.png, document.pdf. |
| **Steps** | 1. Index the project. 2. Query documents for kind="DOC". 3. Check error. |
| **Expected Result** | Zero DOC rows. No error. |
| **Story** | 4.2 |

### TC-4.2-11: docs/ Directory with 10,000+ Files

| Field | Value |
|-------|-------|
| **ID** | TC-4.2-11 |
| **Title** | Batch processing of extremely large docs/ directory |
| **Description** | Verify that indexing a docs/ directory with >10,000 files processes them in batches of 100 and does not block the event loop. |
| **Preconditions** | Project has docs/ with 12,000 small .md files. |
| **Steps** | 1. Index the project. 2. Measure time and check all files are indexed. 3. Verify memory stays under 100MB. |
| **Expected Result** | All 12,000 files are indexed. Processing completes without OOM. Memory usage <100MB. Files processed in batches. |
| **Story** | 4.2 |

---

## Story 4.3 — Index ADRs

### TC-4.3-1: ADRs in docs/adr/ (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-1 |
| **Title** | Index ADRs from docs/adr/ directory |
| **Description** | Verify that ADR files in the standard `docs/adr/` location are indexed with kind="ADR". |
| **Preconditions** | Project has `docs/adr/` containing: `001-use-postgres.md`, `002-use-graphql.md`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | Two rows with kind="ADR". Each has correct path (e.g., `docs/adr/001-use-postgres.md`), content, and content_hash. |
| **Story** | 4.3 |

### TC-4.3-2: ADRs in .adr/ and adr/ Directories

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-2 |
| **Title** | Index ADRs from .adr/ and adr/ |
| **Description** | Verify that ADR files in alternative locations (.adr/, adr/) are also indexed. |
| **Preconditions** | Project has `.adr/001-db-choice.md` and `adr/002-api-design.md`. No `docs/adr/` directory. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | Both files are indexed with kind="ADR". |
| **Story** | 4.3 |

### TC-4.3-3: ADR Naming Patterns — With and Without NNN- Prefix

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-3 |
| **Title** | Index ADRs with both NNN- prefix and plain names |
| **Description** | Verify that any .md file in an ADR directory is indexed, regardless of NNN- prefix. |
| **Preconditions** | Project has `docs/adr/` containing: `001-decision.md`, `architecture-decision.md` (no prefix), `README.md`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | All three files are indexed with kind="ADR". Both naming patterns (NNN-*.md and *.md) are accepted. |
| **Story** | 4.3 |

### TC-4.3-4: No ADR Directory Exists

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-4 |
| **Title** | Project without any ADR directory |
| **Description** | Verify that absence of docs/adr/, .adr/, and adr/ is handled gracefully. |
| **Preconditions** | Project has none of: docs/adr/, .adr/, adr/. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". 3. Check error. |
| **Expected Result** | Zero ADR rows. No error raised. |
| **Story** | 4.3 |

### TC-4.3-5: ADR Directory Exists but Empty

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-5 |
| **Title** | Empty ADR directory |
| **Description** | Verify that an empty ADR directory produces zero ADR records. |
| **Preconditions** | Project has `docs/adr/` but it is empty. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | Zero ADR rows. No error. |
| **Story** | 4.3 |

### TC-4.3-6: Non-.md Files in ADR Directory

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-6 |
| **Title** | Non-markdown files in ADR directory are excluded |
| **Description** | Verify that only .md files in ADR directories are indexed. |
| **Preconditions** | Project has `docs/adr/` containing: `001-decision.md`, `002-diagram.png`, `notes.txt`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | Only `001-decision.md` is indexed. .png and .txt are skipped. |
| **Story** | 4.3 |

### TC-4.3-7: Multiple ADR Directories Exist

| Field | Value |
|-------|-------|
| **ID** | TC-4.3-7 |
| **Title** | Multiple ADR directories are all scanned |
| **Description** | Verify that if multiple ADR directories exist (docs/adr/, .adr/, adr/), all are scanned and indexed. |
| **Preconditions** | Project has: `docs/adr/001-db.md`, `.adr/002-api.md`, `adr/003-auth.md`. |
| **Steps** | 1. Index the project. 2. Query documents for kind="ADR". |
| **Expected Result** | All three files are indexed with kind="ADR". |
| **Story** | 4.3 |

---

## Story 4.4 — Index CHANGELOG

### TC-4.4-1: Index CHANGELOG.md (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-1 |
| **Title** | Index CHANGELOG.md |
| **Description** | Verify that a CHANGELOG.md at the project root is indexed with kind="CHANGELOG". |
| **Preconditions** | Project has CHANGELOG.md with content. |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". |
| **Expected Result** | One row with `kind="CHANGELOG"`, `path="CHANGELOG.md"`, content and content_hash populated. |
| **Story** | 4.4 |

### TC-4.4-2: Index CHANGES.md (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-2 |
| **Title** | Index CHANGES.md |
| **Description** | Verify that a CHANGES.md file is detected and indexed. |
| **Preconditions** | Project has CHANGES.md (no CHANGELOG.md). |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". |
| **Expected Result** | One row with `kind="CHANGELOG"`, `path="CHANGES.md"`. |
| **Story** | 4.4 |

### TC-4.4-3: Index HISTORY.md (Happy Path)

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-3 |
| **Title** | Index HISTORY.md |
| **Description** | Verify that a HISTORY.md file is detected and indexed. |
| **Preconditions** | Project has HISTORY.md (no CHANGELOG.md or CHANGES.md). |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". |
| **Expected Result** | One row with `kind="CHANGELOG"`, `path="HISTORY.md"`. |
| **Story** | 4.4 |

### TC-4.4-4: Multiple Changelog Files Exist

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-4 |
| **Title** | Index multiple changelog files |
| **Description** | Verify that if multiple changelog files coexist, all are indexed. |
| **Preconditions** | Project has CHANGELOG.md, CHANGES.md, and HISTORY.md all present. |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". |
| **Expected Result** | Three rows with kind="CHANGELOG". Each file is indexed independently. |
| **Story** | 4.4 |

### TC-4.4-5: No Changelog File Exists

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-5 |
| **Title** | Project without any changelog file |
| **Description** | Verify that absence of changelog files is not an error. |
| **Preconditions** | Project has no CHANGELOG.md, CHANGES.md, or HISTORY.md. |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". 3. Check error. |
| **Expected Result** | Zero CHANGELOG rows. No error raised. |
| **Story** | 4.4 |

### TC-4.4-6: Case-Insensitive Changelog Detection

| Field | Value |
|-------|-------|
| **ID** | TC-4.4-6 |
| **Title** | Case-insensitive changelog file detection |
| **Description** | Verify that changelog files are detected case-insensitively. |
| **Preconditions** | Project has `changelog.md` (lowercase) and `History.md` (mixed case). |
| **Steps** | 1. Index the project. 2. Query documents for kind="CHANGELOG". |
| **Expected Result** | Both files are indexed with kind="CHANGELOG". |
| **Story** | 4.4 |

---

## Story 4.5 — Full-Text Search Indexing

### TC-4.5-1: FTS5 Virtual Table Exists After Indexing

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-1 |
| **Title** | FTS5 virtual table is created |
| **Description** | Verify that after indexing, an FTS5 virtual table exists on the documents content. |
| **Preconditions** | At least one project has been indexed. |
| **Steps** | 1. Execute `SELECT count(*) FROM documents_fts`. |
| **Expected Result** | Query succeeds without error. FTS5 table exists and is populated. |
| **Story** | 4.5 |

### TC-4.5-2: Phrase Query Returns Exact Matches

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-2 |
| **Title** | Phrase query matches exact phrases |
| **Description** | Verify that a phrase query (e.g., `"error handling"`) returns only documents containing that exact phrase. |
| **Preconditions** | Documents exist containing the phrase "error handling" and documents containing only "error" or "handling" separately. |
| **Steps** | 1. Execute search with query: `"error handling"`. 2. Examine results. |
| **Expected Result** | Only documents containing the exact adjacent phrase "error handling" are returned. Documents with "error" and "handling" in separate locations are excluded. |
| **Story** | 4.5 |

### TC-4.5-3: Boolean AND Operator

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-3 |
| **Title** | Boolean AND operator works |
| **Description** | Verify that `foo AND bar` returns only documents containing both terms. |
| **Preconditions** | Doc A contains "foo" and "bar", Doc B contains only "foo", Doc C contains only "bar". |
| **Steps** | 1. Execute search: `foo AND bar`. 2. Examine results. |
| **Expected Result** | Only Doc A is returned. Doc B and Doc C are excluded. |
| **Story** | 4.5 |

### TC-4.5-4: Boolean OR Operator

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-4 |
| **Title** | Boolean OR operator works |
| **Description** | Verify that `foo OR bar` returns documents containing either term. |
| **Preconditions** | Doc A contains "foo", Doc B contains "bar", Doc C contains neither. |
| **Steps** | 1. Execute search: `foo OR bar`. 2. Examine results. |
| **Expected Result** | Doc A and Doc B are returned. Doc C is excluded. |
| **Story** | 4.5 |

### TC-4.5-5: Boolean NOT Operator

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-5 |
| **Title** | Boolean NOT operator works |
| **Description** | Verify that `foo NOT bar` returns documents containing "foo" but excluding those that also contain "bar". |
| **Preconditions** | Doc A contains "foo", Doc B contains "foo" and "bar", Doc C contains "bar". |
| **Steps** | 1. Execute search: `foo NOT bar`. 2. Examine results. |
| **Expected Result** | Only Doc A is returned. Doc B and Doc C are excluded. |
| **Story** | 4.5 |

### TC-4.5-6: Ranked Results (BM25)

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-6 |
| **Title** | Search results are ranked by relevance |
| **Description** | Verify that results are ordered by relevance score (BM25 or equivalent), with most relevant first. |
| **Preconditions** | Doc A mentions "deployment" 20 times. Doc B mentions "deployment" once. |
| **Steps** | 1. Execute search: `deployment`. 2. Examine result order and rank scores. |
| **Expected Result** | Doc A appears before Doc B. Each result has a rank/score field. Order is descending by relevance. |
| **Story** | 4.5 |

### TC-4.5-7: Results Include Project Context

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-7 |
| **Title** | Search results include project name, path, and kind |
| **Description** | Verify that search results contain project name, document path, and document kind alongside content. |
| **Preconditions** | Multiple projects indexed with various document kinds. |
| **Steps** | 1. Execute any search that returns results. 2. Examine result fields. |
| **Expected Result** | Each result includes: `project_name`, `project_path`, `document_path`, `kind`. |
| **Story** | 4.5 |

### TC-4.5-8: Cross-Portfolio Search

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-8 |
| **Title** | Search across multiple projects |
| **Description** | Verify that search queries return results across all indexed projects in the portfolio. |
| **Preconditions** | Project A and Project B both have documents containing the term "authentication". |
| **Steps** | 1. Execute search: `authentication`. 2. Examine results. |
| **Expected Result** | Results include documents from both Project A and Project B. Different `project_id` values are present in the result set. |
| **Story** | 4.5 |

### TC-4.5-9: Empty Result Set

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-9 |
| **Title** | Search with no matching results returns empty set |
| **Description** | Verify that a search query matching nothing returns an empty array, not an error. |
| **Preconditions** | Projects are indexed but none contain the term "xyznonexistent". |
| **Steps** | 1. Execute search: `xyznonexistent`. 2. Examine response. |
| **Expected Result** | Empty array/result set. No error returned. |
| **Story** | 4.5 |

### TC-4.5-10: Search Query Sanitization

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-10 |
| **Title** | Search handles special characters and injection attempts |
| **Description** | Verify that search queries with special FTS5 syntax characters or SQL injection attempts are handled safely. |
| **Preconditions** | Documents are indexed. |
| **Steps** | 1. Execute search: `'; DROP TABLE documents; --`. 2. Execute search: `*`. 3. Execute search: empty string `""`. |
| **Expected Result** | All queries complete without error or database corruption. Malformed queries return empty results or are sanitized. Empty query may return all documents or be rejected gracefully. |
| **Story** | 4.5 |

### TC-4.5-11: FTS5 Index Rebuilt on Re-index

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-11 |
| **Title** | FTS5 index is rebuilt when documents change |
| **Description** | Verify that when documents are added, removed, or updated between index runs, the FTS5 index reflects the new state. |
| **Preconditions** | First index of Project A created FTS5 entries. Between runs: a DOC is added, a DOC is deleted, a README is modified. |
| **Steps** | 1. Re-index Project A. 2. Search for content from the deleted document. 3. Search for content from the new document. 4. Search for content from the modified document. |
| **Expected Result** | Deleted document content no longer appears in search results. New document content appears. Modified document shows new content. |
| **Story** | 4.5 |

### TC-4.5-12: FTS5 Index Size Constraint

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-12 |
| **Title** | FTS5 index does not exceed 2x raw content size |
| **Description** | Verify that the FTS5 virtual table storage does not exceed 2x the raw content size (NFR-6). |
| **Preconditions** | A project with known total raw content size (e.g., 1MB of documentation text) is fully indexed. |
| **Steps** | 1. Measure raw content size from `documents` table. 2. Measure FTS5 index size from SQLite database file or `dbstat`. 3. Compare. |
| **Expected Result** | FTS5 storage <= 2x raw content size. |
| **Story** | 4.5 |

### TC-4.5-13: Search Performance Under Load

| Field | Value |
|-------|-------|
| **ID** | TC-4.5-13 |
| **Title** | Search latency under 500ms for 500 projects |
| **Description** | Verify that search queries return results in <500ms for a portfolio of up to 500 projects (NFR-7). |
| **Preconditions** | 500 projects are indexed, each with documentation content. |
| **Steps** | 1. Execute a search query matching multiple projects. 2. Measure response time. |
| **Expected Result** | Query returns in <500ms. Results are ranked and include project context. |
| **Story** | 4.5 |

---

## Integration & Cross-Cutting Test Cases

### TC-INT-1: Full End-to-End Indexing Pipeline

| Field | Value |
|-------|-------|
| **ID** | TC-INT-1 |
| **Title** | Complete indexing of a real-world project with all document kinds |
| **Description** | Verify that indexing a project with README, docs/, ADRs, and CHANGELOG produces correct records for all kinds. |
| **Preconditions** | A discovered project has: README.md, docs/guide.md, docs/adr/001-decision.md, and CHANGELOG.md. |
| **Steps** | 1. Index the project. 2. Query documents by kind. 3. Verify counts. |
| **Expected Result** | 1 README, 1 DOC, 1 ADR, 1 CHANGELOG = 4 total document rows. Each has correct kind, path, content, content_hash. |
| **Story** | 4.1–4.5 |

### TC-INT-2: Idempotent Re-index of Unchanged Project

| Field | Value |
|-------|-------|
| **ID** | TC-INT-2 |
| **Title** | Full re-index of unchanged project is a no-op |
| **Description** | Verify that running the full indexing pipeline twice on an unchanged project produces identical results. |
| **Preconditions** | Project fully indexed. No files changed. |
| **Steps** | 1. Record all document rows and indexed_at values. 2. Re-index. 3. Query all documents again. |
| **Expected Result** | Same number of rows. Same content_hash values. indexed_at is not updated. FTS5 queries return identical results. |
| **Story** | All |

### TC-INT-3: File Deleted Between Index Runs

| Field | Value |
|-------|-------|
| **ID** | TC-INT-3 |
| **Title** | Orphaned document records are cleaned up on re-index |
| **Description** | Verify that when a file is deleted from disk, its document record is removed on the next index run. |
| **Preconditions** | Project indexed with 3 docs. One .md file is then deleted from docs/. |
| **Steps** | 1. Re-index the project. 2. Query documents. |
| **Expected Result** | Only 2 document rows remain. The deleted file's row is removed. FTS5 no longer contains content from the deleted file. |
| **Story** | All |

### TC-INT-4: Content Hash Deduplication

| Field | Value |
|-------|-------|
| **ID** | TC-INT-4 |
| **Title** | Identical content produces duplicate hash — idempotent |
| **Description** | Verify that content_hash deduplication prevents duplicate rows. |
| **Preconditions** | Project indexed. Files unchanged. |
| **Steps** | 1. Check row count. 2. Re-index. 3. Check row count. |
| **Expected Result** | Row count is identical. No duplicate rows. `content_hash` values match between runs. |
| **Story** | All |

### TC-INT-5: Content Hash Changes on Update

| Field | Value |
|-------|-------|
| **ID** | TC-INT-5 |
| **Title** | Modified file gets new content_hash and updated indexed_at |
| **Description** | Verify that a modified file generates a new content_hash and updates indexed_at. |
| **Preconditions** | Project indexed with baseline content. A file in docs/ is edited. |
| **Steps** | 1. Record old hash and timestamp. 2. Re-index. 3. Query updated row. |
| **Expected Result** | New content_hash differs from old. indexed_at is newer. Content reflects new file content. |
| **Story** | All |

### TC-INT-6: metadata.documentation_hash Is Updated

| Field | Value |
|-------|-------|
| **ID** | TC-INT-6 |
| **Title** | Project metadata hash is updated after indexing |
| **Description** | Verify that `metadata.documentation_hash` is computed and stored after indexing completes. |
| **Preconditions** | Project exists with metadata row (from Epic 2.2). Documentation not yet indexed. |
| **Steps** | 1. Index the project. 2. Query metadata.documentation_hash. 3. Modify a doc and re-index. 4. Query again. |
| **Expected Result** | After first index: `documentation_hash` is non-null. After re-index with changes: hash differs. Hash is a deterministic combination of all document content_hashes. |
| **Story** | All |

### TC-INT-7: Database Transaction Failure During Indexing

| Field | Value |
|-------|-------|
| **ID** | TC-INT-7 |
| **Title** | Partial index on transaction failure rolls back |
| **Description** | Verify that if a database transaction fails mid-indexing, no partial state persists. |
| **Preconditions** | Project with multiple files. Simulate a database write failure (e.g., disk full, constraint violation). |
| **Steps** | 1. Index the project with simulated failure partway through. 2. Query documents and FTS5. |
| **Expected Result** | Transaction rolls back. No partial document records exist. FTS5 index reflects pre-index state. Error is propagated to caller. |
| **Story** | All |

### TC-INT-8: Concurrency Lock Per Project

| Field | Value |
|-------|-------|
| **ID** | TC-INT-8 |
| **Title** | Concurrent indexing of same project is serialized |
| **Description** | Verify that two simultaneous indexDocumentation calls for the same project are serialized via mutex/lock. |
| **Preconditions** | Project exists. Two concurrent callers invoke indexDocumentation. |
| **Steps** | 1. Start two index calls simultaneously for the same project_id. 2. Observe behavior. |
| **Expected Result** | Second caller either waits for the first to complete, or immediately receives "already indexing" response. No corruption occurs. |
| **Story** | All |

### TC-INT-9: Concurrent Indexing of Different Projects

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9 |
| **Title** | Concurrent indexing of different projects succeeds |
| **Description** | Verify that indexing two different projects concurrently works without interference. |
| **Preconditions** | Project A and Project B both exist and are not yet indexed. |
| **Steps** | 1. Start indexing Project A and Project B simultaneously. 2. Query documents for both. |
| **Expected Result** | Both index operations complete successfully. Documents for each project are correctly stored and attributed. No cross-contamination. |
| **Story** | All |

### TC-INT-10: Repository Path Changes Between Runs

| Field | Value |
|-------|-------|
| **ID** | TC-INT-10 |
| **Title** | Re-indexing after repository path change |
| **Description** | Verify that if a repository is moved on disk, re-discovery and re-indexing works correctly. |
| **Preconditions** | Project was indexed at path `/old/path/project`. Repository moved to `/new/path/project`. |
| **Steps** | 1. Re-discover projects. 2. Re-index the project at new path. 3. Query documents. |
| **Expected Result** | Documents are keyed to the correct (new) project_id or matched via root_path. Old orphaned records may remain (or are cleaned up per cleanup strategy). |
| **Story** | All |

### TC-INT-11: Unicode and Non-UTF8 File Content

| Field | Value |
|-------|-------|
| **ID** | TC-INT-11 |
| **Title** | Index files with unicode and non-UTF8 content |
| **Description** | Verify that files containing unicode (UTF-8) and non-UTF8 encoded text are handled without error. |
| **Preconditions** | Project has: `docs/unicode.md` containing UTF-8 text with emoji and CJK characters, `docs/iso8859.txt` encoded in ISO-8859-1, `docs/binary.bin` with raw bytes. |
| **Steps** | 1. Index the project. 2. Query documents. |
| **Expected Result** | `unicode.md` is indexed as text with content preserved. `iso8859.txt` is either detected as non-UTF8 and stored as BLOB, or transcoded. `binary.bin` is detected as binary and skipped. No crash or data corruption. |
| **Story** | All |

### TC-INT-12: Engine Never Modifies Repository Files

| Field | Value |
|-------|-------|
| **ID** | TC-INT-12 |
| **Title** | Indexing is read-only — no repo files modified |
| **Description** | Verify that the indexing process never writes to, creates, or deletes files in the repository (NFR-4). |
| **Preconditions** | A repository with known file count and modification timestamps. |
| **Steps** | 1. Record all file modification timestamps and `git status` of the repo. 2. Index the project. 3. Run `git status` and compare timestamps. |
| **Expected Result** | `git status` shows clean (no changes). All file modification timestamps are unchanged. No new files created in repo. |
| **Story** | All |

### TC-INT-13: Indexing Performance — Cold Start (<2s for <100 files)

| Field | Value |
|-------|-------|
| **ID** | TC-INT-13 |
| **Title** | Cold indexing of project with <100 files completes in <2s |
| **Description** | Verify that indexing a project with fewer than 100 documentation files (cold, no prior index) completes within 2 seconds (NFR-1). |
| **Preconditions** | Project has ~80 documentation files across all kinds. No prior index exists. |
| **Steps** | 1. Measure time for `indexDocumentation()` to return. 2. Repeat 3 times with fresh database. |
| **Expected Result** | Average time < 2000ms. |
| **Story** | All |

### TC-INT-14: Indexing Performance — Hot Start (<500ms for unchanged)

| Field | Value |
|-------|-------|
| **ID** | TC-INT-14 |
| **Title** | Hot re-index of unchanged project completes in <500ms |
| **Description** | Verify that re-indexing an already-indexed project with no changes completes in under 500ms (NFR-1). |
| **Preconditions** | Project fully indexed. No files changed. |
| **Steps** | 1. Measure time for second `indexDocumentation()` call. 2. Repeat 3 times. |
| **Expected Result** | Average time < 500ms. |
| **Story** | All |

### TC-INT-15: Memory Usage Under 100MB

| Field | Value |
|-------|-------|
| **ID** | TC-INT-15 |
| **Title** | Memory usage does not exceed 100MB during indexing |
| **Description** | Verify that indexing a large project does not exceed 100MB of memory (NFR-2). |
| **Preconditions** | Project with 10,000 documentation files of various sizes (total ~50MB content). |
| **Steps** | 1. Index the project while monitoring peak memory usage (e.g., via `runtime.ReadMemStats` or OS tools). |
| **Expected Result** | Peak memory usage < 100MB. |
| **Story** | All |

### TC-INT-16: README Takes Priority Over docs/ README

| Field | Value |
|-------|-------|
| **ID** | TC-INT-16 |
| **Title** | Root README is indexed with kind README even if docs/ also has a README |
| **Description** | Verify that the root-level README is always kind="README" even if docs/ contains a file named README.md. |
| **Preconditions** | Project has both `README.md` (root) and `docs/README.md`. |
| **Steps** | 1. Index the project. 2. Query for kind="README". 3. Query for kind="DOC" with path matching docs/README.md. |
| **Expected Result** | Root README is kind="README". docs/README.md is kind="DOC" (because it's in the docs/ directory). Both are indexed. |
| **Story** | 4.1, 4.2 |

### TC-INT-17: Empty Project Indexing

| Field | Value |
|-------|-------|
| **ID** | TC-INT-17 |
| **Title** | Index a project with no documentation files at all |
| **Description** | Verify that indexing a project with zero documentation files (no README, no docs/, no ADRs, no CHANGELOG) completes successfully with zero documents. |
| **Preconditions** | Project exists but has none of: README, docs/, ADR dirs, CHANGELOG files. |
| **Steps** | 1. Index the project. 2. Query all documents for this project. 3. Check error. |
| **Expected Result** | Zero document rows. No error. metadata.documentation_hash is set to a deterministic "empty" value (e.g., hash of empty string). |
| **Story** | All |

### TC-INT-18: Deterministic Hashing — Same Project, Same Result

| Field | Value |
|-------|-------|
| **ID** | TC-INT-18 |
| **Title** | Same project state produces identical content hashes |
| **Description** | Verify determinism: indexing the same project at the same state produces identical content_hash values for all documents (NFR-3). |
| **Preconditions** | A known project with fixed content. |
| **Steps** | 1. Index -> record all hash values. 2. Drop database. 3. Index again -> record all hash values. 4. Compare. |
| **Expected Result** | All content_hash values are identical between runs. |
| **Story** | All |

### TC-INT-19: FTS5 Table Truncated and Rebuilt on Re-index

| Field | Value |
|-------|-------|
| **ID** | TC-INT-19 |
| **Title** | FTS5 index is truncated and rebuilt to stay in sync |
| **Description** | Verify that on re-index, the FTS5 table is truncated and rebuilt from the current documents table to prevent sync drift. |
| **Preconditions** | Project indexed. Directly insert a rogue row into documents_fts (simulating sync drift). |
| **Steps** | 1. Insert a fake row directly into documents_fts. 2. Verify it appears in search. 3. Re-index the project. 4. Search for the fake row's content. |
| **Expected Result** | After re-index, the fake row is gone. FTS5 content matches documents table exactly. |
| **Story** | 4.5 |

### TC-INT-20: Interrupt Safety — Partial Index on Crash

| Field | Value |
|-------|-------|
| **ID** | TC-INT-20 |
| **Title** | In-progress index does not corrupt database on interrupt |
| **Description** | Verify that if the indexing process is interrupted (kill, crash), the database remains in a consistent state (NFR-5). |
| **Preconditions** | Project with 500 documentation files. |
| **Steps** | 1. Begin indexing. 2. Kill the process partway through (e.g., after 50 files). 3. Restart and check database integrity. 4. Run `PRAGMA integrity_check`. |
| **Expected Result** | Database integrity check passes. No partial document state exists (transaction rolled back). Indexing can be cleanly resumed. |
| **Story** | All |

---

## Edge Case Matrix Summary

| # | Edge Case | Covered By |
|---|-----------|------------|
| 1 | No README file | TC-4.1-5 |
| 2 | README >1MB | TC-4.1-6 |
| 3 | README with non-standard extension | TC-4.1-9 |
| 4 | docs/ directory does not exist | TC-4.2-4 |
| 5 | Binary file in docs/ | TC-4.2-3, TC-4.2-9 |
| 6 | Symlink in docs/ | TC-4.2-7 |
| 7 | Nested directory >10 levels deep in docs/ | TC-4.2-8 |
| 8 | No ADR directory | TC-4.3-4 |
| 9 | ADR file without NNN- prefix | TC-4.3-3 |
| 10 | No CHANGELOG file | TC-4.4-5 |
| 11 | Multiple changelog files | TC-4.4-4 |
| 12 | File changes between index runs | TC-4.1-8, TC-INT-5 |
| 13 | Database transaction failure | TC-INT-7 |
| 14 | FTS5 out of sync with documents | TC-INT-19 |
| 15 | Repository path changes | TC-INT-10 |
| 16 | File deleted between runs | TC-INT-3 |
| 17 | Unicode/non-UTF8 content | TC-INT-11 |
| 18 | Extremely large docs/ (>10,000 files) | TC-4.2-11 |
| 19 | Concurrent indexing same project | TC-INT-8 |
| 20 | Concurrent indexing different projects | TC-INT-9 |
| 21 | Empty README file | TC-4.1-10 |
| 22 | README is a symlink | TC-4.1-11 |
| 23 | Case-insensitive detection (all kinds) | TC-4.1-4, TC-4.4-6 |
| 24 | Empty docs/ directory | TC-4.2-10 |
| 25 | Empty ADR directory | TC-4.3-5 |
| 26 | .gitignore exclusion | TC-4.2-5 |
| 27 | Non-.md files in ADR dir | TC-4.3-6 |
| 28 | Search with special characters | TC-4.5-10 |
| 29 | Performance: cold <2s | TC-INT-13 |
| 30 | Performance: hot <500ms | TC-INT-14 |
| 31 | Memory <100MB | TC-INT-15 |
| 32 | FTS5 size <2x raw | TC-4.5-12 |
| 33 | Search latency <500ms for 500 projects | TC-4.5-13 |
| 34 | Deterministic hashing | TC-INT-18 |
| 35 | Read-only repository guarantee | TC-INT-12 |
| 36 | Interrupt safety | TC-INT-20 |

---

## Coverage Verification

| Requirement | Covered By |
|-------------|------------|
| FR-1: README detection (case-insensitive, extensions) | TC-4.1-1, TC-4.1-2, TC-4.1-3, TC-4.1-4 |
| FR-2: Store README with correct fields | TC-4.1-1 |
| FR-3: Truncate README >1MB | TC-4.1-6 |
| FR-4: Recursively scan docs/ | TC-4.2-2 |
| FR-5: Index .md/.rst/.txt/.adoc as DOC | TC-4.2-1 |
| FR-6: Skip binary files | TC-4.2-3, TC-4.2-9 |
| FR-7: Find ADRs in standard locations | TC-4.3-1, TC-4.3-2 |
| FR-8: Recognize ADR naming patterns | TC-4.3-3 |
| FR-9: Find changelog files | TC-4.4-1, TC-4.4-2, TC-4.4-3 |
| FR-10: Create FTS5 virtual table | TC-4.5-1 |
| FR-11: Phrase queries + Boolean operators | TC-4.5-2, TC-4.5-3, TC-4.5-4, TC-4.5-5 |
| FR-12: Ranked results with project context | TC-4.5-6, TC-4.5-7 |
| FR-13: Cross-portfolio search | TC-4.5-8 |
| FR-14: Deduplication via content_hash | TC-INT-4 |
| FR-15: Update indexed_at on content change | TC-4.1-8, TC-INT-5 |
| FR-16: Update metadata.documentation_hash | TC-INT-6 |
| FR-17: Respect .gitignore | TC-4.2-5 |
| NFR-1: Cold <2s / Hot <500ms | TC-INT-13, TC-INT-14 |
| NFR-2: Memory <100MB | TC-INT-15 |
| NFR-3: Deterministic | TC-INT-18 |
| NFR-4: Read-only | TC-INT-12 |
| NFR-5: Interrupt-safe | TC-INT-20 |
| NFR-6: FTS5 size <2x raw | TC-4.5-12 |
| NFR-7: Search <500ms for 500 projects | TC-4.5-13 |
| NFR-8: Idempotent | TC-INT-2 |
