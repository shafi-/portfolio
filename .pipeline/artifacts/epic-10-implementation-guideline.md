# Epic 10 — AI Analysis: Implementation Guidelines

**Version:** 1.0  
**Milestone:** 2 — Agent Integration  
**Status:** Ready for Implementation  
**Total Size:** 4M (~12-20 days)

---

## 1. Technical Standards

### 1.1 Go Version and Tooling

- **Go Version:** Go 1.21 or higher
- **Module:** `github.com/nerddevsltd/portfolio-tool`
- **Testing:** Standard Go testing package with testify for assertions
- **Logging:** `go.uber.org/zap` (structured logging, from Epic 0)
- **Database:** SQLite via `github.com/mattn/go-sqlite3` (from Epic 5)
- **UUID:** `github.com/google/uuid` (from Epic 0)
- **JSON Schema:** `github.com/xeipuuv/gojsonschema` for validation
- **MCP Protocol:** `github.com/mark3labs/mcp-go` (from Epic 7)

### 1.2 Code Conventions

#### Package Structure
- Follow standard Go project layout
- `internal/` packages are private to the application
- `pkg/` packages can be imported by other modules
- Use short, lowercase package names
- Avoid package name redundancy in type names

#### Naming Conventions
```go
// Types: PascalCase
type AnalysisService struct { }
type Analysis struct { }

// Interfaces: CamelCase, no "I" prefix
type AnalysisStore interface { }

// Functions: PascalCase for exported, camelCase for private
func StoreAnalysis(ctx context.Context, ...) (*Analysis, error) { }
func validateInput(input string) error { }

// Constants: PascalCase for exported, camelCase for private
const ErrProjectNotFound = "PROJECT_NOT_FOUND"
const maxAnalysisSize = 10 * 1024 * 1024

// Errors: Wrap with context
return fmt.Errorf("failed to store analysis: %w", err)
```

#### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to validate project existence: %w", err)
}

// Use custom error types for domain-specific errors
type AnalysisError struct {
    Code    string
    Message string
    Cause   error
}

// Return nil for "not found" scenarios, not errors
func GetAnalysis(...) (*Analysis, error) {
    if notFound {
        return nil, nil  // Graceful, not error
    }
}
```

#### Context Usage
```go
// Accept context as first parameter
func StoreAnalysis(ctx context.Context, ...) error { }

// Pass context through all layers
func (s *AnalysisService) StoreAnalysis(ctx context.Context, ...) {
    if err := s.store.CreateAnalysis(ctx, analysis); err != nil {
        return fmt.Errorf("store failed: %w", err)
    }
}

// Respect context cancellation
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // Continue work
}
```

#### Database Operations
```go
// Use transactions for multi-step operations
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback()

// Perform operations
if err := s.createAnalysis(tx, analysis); err != nil {
    return err
}

// Commit on success
if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

#### Logging
```go
// Use structured logging
logger.Info("storing analysis",
    zap.String("project_id", projectID.String()),
    zap.String("analyzer", analyzer),
    zap.Int("feature_count", len(features)),
)

// Debug logging for detailed information
logger.Debug("analysis validation result",
    zap.Bool("valid", valid),
    zap.Strings("errors", errors),
)

// Error logging with context
logger.Error("failed to store analysis",
    zap.String("project_id", projectID.String()),
    zap.Error(err),
)
```

### 1.3 Database Conventions

#### Query Patterns
```go
// Use parameterized queries (no string concatenation)
const queryGetAnalysis = `
    SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
           summary, purpose, architecture, maturity, strengths, weaknesses,
           reusable_components, notes, raw_json, created_at, updated_at
    FROM analyses
    WHERE project_id = ?
    ORDER BY analyzed_at DESC
    LIMIT 1
`

// Scan into structs
row := s.db.QueryRowContext(ctx, queryGetAnalysis, projectID)
var analysis Analysis
if err := row.Scan(&analysis.ID, &analysis.ProjectID, ...); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil  // Not found, graceful
    }
    return nil, fmt.Errorf("failed to scan analysis: %w", err)
}
```

#### Transaction Management
```go
// Always defer rollback
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer func() {
    if p := recover(); p != nil {
        tx.Rollback()
        panic(p)  // Re-throw after rollback
    }
}()

// Commit explicitly
if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit: %w", err)
}
```

#### Connection Management
```go
// Use connection pooling (SQLite default)
// Set reasonable limits
db.SetMaxOpenConns(1)      // SQLite writes are serialized
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(0)   // Connections live forever
```

---

## 2. Package Organization

### 2.1 Package Map

```
internal/
  analysis/
    ├── service.go              # AnalysisService: orchestration layer
    ├── service_test.go         # Service unit tests
    ├── schema_validator.go     # JSON schema validation
    ├── schema_validator_test.go    # Validation tests
    ├── stale_detector.go       # Stale detection logic
    ├── stale_detector_test.go  # Stale detection tests
    ├── relationship.go         # Relationship service
    ├── relationship_test.go    # Relationship tests
    ├── types.go                # Domain types
    ├── errors.go               # Analysis-specific error types
    └── repository.go           # AnalysisStore interface definition

  repository/
    ├── analysis_store.go       # SQLite implementation of AnalysisStore
    ├── analysis_store_test.go  # Store tests with in-memory SQLite
    └── queries.go              # SQL queries for analyses/features/relationships

  mcp/
    └── tools/
        ├── analysis.go         # Analysis MCP tools (store, get, list needing)
        ├── analysis_test.go    # Analysis tool tests
        ├── relationships.go    # Relationship MCP tools (list, store, delete)
        └── relationships_test.go   # Relationship tool tests

pkg/models/
  └── analysis.go               # Public domain model types

migrations/
  └── 001_add_analysis_tables.sql    # Database migration
```

### 2.2 Package Dependencies

```
internal/analysis
  ├── repository (AnalysisStore interface)
  ├── go.uber.org/zap (logging)
  └── github.com/xeipuuv/gojsonschema (validation)

internal/repository
  ├── database/sql (standard library)
  ├── github.com/mattn/go-sqlite3 (SQLite driver)
  └── go.uber.org/zap (logging)

internal/mcp/tools
  ├── internal/analysis (service layer)
  ├── internal/repository (store layer)
  └── github.com/mark3labs/mcp-go (MCP protocol)

pkg/models
  └── github.com/google/uuid (UUID types)
```

### 2.3 Package Responsibilities

#### `internal/analysis`
- **Purpose:** Service layer for analysis operations
- **Responsibilities:**
  - Orchestrate analysis storage and retrieval
  - Validate analysis inputs against JSON schema
  - Compute staleness indicators
  - Manage relationship operations
  - Convert between domain types and storage models
- **Dependencies:** Repository interface, logger, schema validator

#### `internal/repository`
- **Purpose:** Data access layer for SQLite
- **Responsibilities:**
  - Execute SQL queries
  - Map database rows to domain types
  - Manage transactions
  - Handle connection lifecycle
- **Dependencies:** Database driver, logger

#### `internal/mcp/tools`
- **Purpose:** MCP tool implementations
- **Responsibilities:**
  - Register MCP tools with registry
  - Parse and validate MCP requests
  - Call service layer methods
  - Format responses as MCP protocol messages
- **Dependencies:** Service layer, MCP framework

#### `pkg/models`
- **Purpose:** Public domain model types
- **Responsibilities:**
  - Define domain types shared across packages
  - Provide serialization/deserialization methods
- **Dependencies:** None (pure data types)

---

## 3. Implementation Order

### 3.1 Story 10.1 — Analysis Schema (2 days)

#### Phase 1: Foundation Types
1. **File:** `internal/analysis/types.go`
   - Define `Analysis`, `AnalysisInput`, `Feature`, `FeatureInput` types
   - Define `Relationship`, `NeedsAnalysisResult` types
   - Add JSON tags for serialization
   - Add validation tags if using struct validator

2. **File:** `pkg/models/analysis.go`
   - Export public domain model types
   - Import from `internal/analysis/types.go` or re-export
   - Document public API

3. **File:** `internal/analysis/errors.go`
   - Define `Error` struct with `Code`, `Message`, `Cause`
   - Define error constants: `ErrProjectNotFound`, `ErrAnalysisNotFound`, etc.
   - Implement `Error()` and `Unwrap()` methods
   - Define error code constants

4. **File:** `internal/analysis/repository.go`
   - Define `AnalysisStore` interface with all required methods
   - Document each method with context and return semantics

#### Phase 2: Schema Validation
5. **File:** `internal/analysis/schema_validator.go`
   - Implement `SchemaValidator` struct
   - Define JSON schema as constant
   - Implement `Validate()` method
   - Return detailed validation errors

6. **File:** `internal/analysis/schema_validator_test.go`
   - Test valid analysis payload passes
   - Test missing required fields fail
   - Test invalid field types fail
   - Test array field validation
   - Test ISO 8601 timestamp validation
   - Test additional fields preserved
   - Test empty optional fields

#### Verification
- All unit tests pass
- Schema validation correctly accepts valid payloads
- Schema validation correctly rejects invalid payloads with detailed errors
- Code coverage >= 90% for schema_validator.go

---

### 3.2 Story 10.2 — Persist Analyses (3 days)

#### Phase 1: Database Schema
7. **File:** `migrations/001_add_analysis_tables.sql`
   - Create `analyses` table with all columns
   - Create `features` table with all columns
   - Add foreign key constraints with CASCADE
   - Add uniqueness constraint: `UNIQUE(project_id, analyzer)`
   - Create indexes on `project_id`, `analyzer`, `analyzed_at`
   - Create indexes on `features.analysis_id`, `features.name`
   - Test migration in local environment

#### Phase 2: Repository Implementation
8. **File:** `repository/queries.go`
   - Define SQL queries as constants
   - Include: INSERT analysis, UPDATE analysis, SELECT analysis, DELETE features
   - Document each query's purpose and parameters

9. **File:** `repository/analysis_store.go`
   - Implement `SQLiteAnalysisStore` struct
   - Implement `CreateAnalysis()` method
   - Implement `UpdateAnalysis()` method
   - Implement `GetAnalysis()` method
   - Implement `GetAnalysisByAnalyzer()` method
   - Implement `ProjectExists()` method
   - Implement `CreateFeatures()` method
   - Implement `DeleteFeaturesByAnalysisID()` method
   - Implement `GetFeaturesByAnalysisID()` method
   - Use transactions for multi-step operations

10. **File:** `repository/analysis_store_test.go`
    - Test create new analysis
    - Test update existing analysis
    - Test get analysis by project ID
    - Test get analysis by analyzer
    - Test project existence check
    - Test feature creation and retrieval
    - Test feature deletion on analysis update
    - Test foreign key constraints
    - Test cascade deletes
    - Use in-memory SQLite for tests

#### Phase 3: Service Layer
11. **File:** `internal/analysis/service.go`
    - Implement `AnalysisService` struct
    - Implement `StoreAnalysis()` method:
      - Validate project exists
      - Validate analysis schema
      - Serialize to raw_json
      - Check for existing analysis
      - Start transaction
      - Delete old features (if update)
      - Insert/update analysis
      - Insert new features
      - Commit transaction
    - Implement `GetAnalysis()` method
    - Implement `GetAnalysisByAnalyzer()` method
    - Add comprehensive logging
    - Handle all error cases

12. **File:** `internal/analysis/service_test.go`
    - Test store analysis (new)
    - Test store analysis (update)
    - Test store analysis (invalid project)
    - Test store analysis (schema validation error)
    - Test get analysis (exists)
    - Test get analysis (not found)
    - Test get analysis (multiple analyzers)
    - Test feature handling
    - Mock repository for isolated testing

#### Phase 4: MCP Tools
13. **File:** `internal/mcp/tools/analysis.go`
    - Implement `RegisterAnalysisTools()` function
    - Register `storeAnalysis` tool with schema
    - Register `getAnalysis` tool with schema
    - Implement `HandleStoreAnalysis()` handler:
      - Parse request
      - Call service.StoreAnalysis()
      - Format response
    - Implement `HandleGetAnalysis()` handler:
      - Parse request
      - Call service.GetAnalysis()
      - Format response

14. **File:** `internal/mcp/tools/analysis_test.go`
    - Test MCP tool registration
    - Test store analysis tool (success)
    - Test store analysis tool (validation error)
    - Test get analysis tool (exists)
    - Test get analysis tool (not found)
    - Test input parsing
    - Test response formatting
    - Mock service for isolated testing

#### Verification
- All unit tests pass
- All integration tests pass
- Database migration applies successfully
- MCP tools register correctly
- Analysis storage and retrieval work end-to-end
- Code coverage >= 80% for analysis package
- Performance targets met (<100ms for store, <50ms for get)

---

### 3.3 Story 10.3 — Detect Stale Analyses (2 days)

#### Phase 1: Stale Detection Logic
15. **File:** `internal/analysis/stale_detector.go`
    - Implement `StaleDetector` struct
    - Implement `IsOutdated()` method:
      - Get current git_head from metadata
      - Compare with analyzed_git_head
      - Handle NULL git_head cases
    - Implement `ListNeedingAnalysis()` method:
      - Execute single LEFT JOIN query
      - Compute reason for each project
      - Return results

16. **File:** `repository/analysis_store.go` (additions)
    - Add `GetGitHeadForProject()` method
    - Add `ListAllAnalyses()` method
    - Optimize query with proper indexes

17. **File:** `internal/analysis/stale_detector_test.go`
    - Test is outdated (git head mismatch)
    - Test is current (git head match)
    - Test NULL git_head handling
    - Test list needing analysis (unanalyzed)
    - Test list needing analysis (outdated)
    - Test list needing analysis (empty)
    - Test list needing analysis (all current)
    - Test single query execution (verify no N+1)

18. **File:** `repository/analysis_store_test.go` (additions)
    - Test get git head for project
    - Test list all analyses
    - Test query performance with 500 projects

#### Phase 2: MCP Tool
19. **File:** `internal/mcp/tools/analysis.go` (additions)
    - Register `listProjectsNeedingAnalysis` tool
    - Implement `HandleListNeedingAnalysis()` handler:
      - Call stale_detector.ListNeedingAnalysis()
      - Format response

20. **File:** `internal/mcp/tools/analysis_test.go` (additions)
    - Test list needing analysis tool
    - Test response format
    - Test empty result handling
    - Mock stale detector for isolated testing

#### Phase 3: Performance Testing
21. **File:** `internal/analysis/stale_detector_bench_test.go`
    - Benchmark `ListNeedingAnalysis()` with 500 projects
    - Verify single query execution
    - Measure query time

#### Verification
- All unit tests pass
- Stale detection correctly identifies outdated analyses
- `listProjectsNeedingAnalysis()` returns both unanalyzed and outdated projects
- Single query execution verified (no N+1 problem)
- Performance target met (<200ms for 500 projects)
- Indicator not persisted (verified by schema inspection)

---

### 3.4 Story 10.4 — Relationship Persistence (2 days)

#### Phase 1: Database Schema
22. **File:** `migrations/001_add_analysis_tables.sql` (additions)
    - Create `relationships` table
    - Add foreign key constraints with CASCADE
    - Add CHECK constraint for relationship types
    - Add CHECK constraint for confidence range
    - Add uniqueness constraint: `UNIQUE(source_project, target_project, type)`
    - Create indexes on `source_project`, `target_project`, `type`
    - Test migration

#### Phase 2: Repository Implementation
23. **File:** `repository/queries.go` (additions)
    - Define relationship SQL queries
    - Include: INSERT/UPDATE, SELECT by project, DELETE, SELECT by source/target/type

24. **File:** `repository/analysis_store.go` (additions)
    - Implement `CreateRelationship()` method
    - Implement `UpdateRelationship()` method
    - Implement `GetRelationship()` method
    - Implement `ListRelationshipsByProject()` method
    - Implement `DeleteRelationship()` method
    - Implement `FindExistingRelationship()` method
    - Use INSERT OR REPLACE for deduplication

25. **File:** `repository/analysis_store_test.go` (additions)
    - Test create relationship
    - Test update relationship (deduplication)
    - Test list relationships (source)
    - Test list relationships (target)
    - Test list relationships (both directions)
    - Test delete relationship
    - Test foreign key constraints
    - Test CHECK constraints (type, confidence)
    - Test uniqueness constraint

#### Phase 3: Service Layer
26. **File:** `internal/analysis/relationship.go`
    - Implement `RelationshipService` struct
    - Implement `StoreRelationship()` method:
      - Validate project IDs exist
      - Validate relationship type
      - Validate confidence range
      - Check for existing relationship
      - Insert or update relationship
    - Implement `ListRelationships()` method
    - Implement `DeleteRelationship()` method
    - Implement `ValidateRelationshipType()` method
    - Define allowed relationship types map
    - Add comprehensive logging

27. **File:** `internal/analysis/relationship_test.go`
    - Test store relationship (new)
    - Test store relationship (update)
    - Test store relationship (invalid type)
    - Test store relationship (invalid confidence)
    - Test store relationship (non-existent project)
    - Test list relationships (source)
    - Test list relationships (target)
    - Test list relationships (both directions)
    - Test list relationships (empty)
    - Test delete relationship
    - Test duplicate handling
    - Mock repository for isolated testing

#### Phase 4: MCP Tools
28. **File:** `internal/mcp/tools/relationships.go`
    - Implement `RegisterRelationshipTools()` function
    - Register `listRelationships` tool with schema
    - Register `storeRelationship` tool with schema
    - Register `deleteRelationship` tool with schema
    - Implement handlers for each tool:
      - Parse request
      - Call service methods
      - Format response

29. **File:** `internal/mcp/tools/relationships_test.go`
    - Test MCP tool registration
    - Test list relationships tool
    - Test store relationship tool (success)
    - Test store relationship tool (validation errors)
    - Test delete relationship tool
    - Test input parsing
    - Test response formatting
    - Mock service for isolated testing

#### Verification
- All unit tests pass
- All integration tests pass
- Relationship CRUD operations work end-to-end
- Relationship type validation enforced
- Confidence range validation enforced
- Foreign key constraints enforced
- Deduplication works correctly
- Code coverage >= 80% for relationship package
- Performance target met (<100ms for list relationships)

---

## 4. Code Patterns to Follow

### 4.1 Service Layer Pattern

```go
// Service orchestration with validation, logging, error handling
func (s *AnalysisService) StoreAnalysis(ctx context.Context, projectID uuid.UUID, input AnalysisInput) (*Analysis, error) {
    // Log operation start
    logger := s.logger.With(zap.String("project_id", projectID.String()))
    logger.Info("storing analysis", zap.String("analyzer", input.Analyzer))

    // Step 1: Validate project exists
    exists, err := s.store.ProjectExists(ctx, projectID)
    if err != nil {
        return nil, fmt.Errorf("failed to check project existence: %w", err)
    }
    if !exists {
        return nil, ErrProjectNotFound
    }

    // Step 2: Validate schema
    if err := s.schemaValidator.Validate(input); err != nil {
        return nil, fmt.Errorf("%w: %v", ErrCodeSchemaValidation, err)
    }

    // Step 3: Serialize to raw_json
    rawJSON, err := json.Marshal(input)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal analysis: %w", err)
    }

    // Step 4: Check for existing analysis
    existing, err := s.store.GetAnalysisByAnalyzer(ctx, projectID, input.Analyzer)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("failed to check for existing analysis: %w", err)
    }

    // Step 5: Prepare analysis object
    analysis := &Analysis{
        ID:                 uuid.New(),
        ProjectID:          projectID,
        Analyzer:           input.Analyzer,
        AnalyzedGitHead:    input.AnalyzedGitHead,
        AnalyzedAt:         input.AnalyzedAt,
        Summary:            input.Summary,
        Purpose:            input.Purpose,
        Architecture:       input.Architecture,
        Maturity:           input.Maturity,
        Strengths:          input.Strengths,
        Weaknesses:         input.Weaknesses,
        ReusableComponents: input.ReusableComponents,
        Notes:              input.Notes,
        RawJSON:            rawJSON,
        CreatedAt:          time.Now().UTC(),
        UpdatedAt:          time.Now().UTC(),
    }

    // Step 6: Transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Step 7: Insert or update analysis
    if existing != nil {
        if err := s.store.UpdateAnalysis(ctx, analysis); err != nil {
            return nil, fmt.Errorf("failed to update analysis: %w", err)
        }
        // Delete old features
        if err := s.store.DeleteFeaturesByAnalysisID(ctx, existing.ID); err != nil {
            return nil, fmt.Errorf("failed to delete old features: %w", err)
        }
    } else {
        if err := s.store.CreateAnalysis(ctx, analysis); err != nil {
            return nil, fmt.Errorf("failed to create analysis: %w", err)
        }
    }

    // Step 8: Insert features
    if len(input.Features) > 0 {
        features := make([]Feature, len(input.Features))
        for i, f := range input.Features {
            features[i] = Feature{
                ID:          uuid.New(),
                AnalysisID:  analysis.ID,
                Name:        f.Name,
                Description: f.Description,
                Confidence:  f.Confidence,
                CreatedAt:   time.Now().UTC(),
            }
        }
        if err := s.store.CreateFeatures(ctx, features); err != nil {
            return nil, fmt.Errorf("failed to create features: %w", err)
        }
    }

    // Step 9: Commit transaction
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    logger.Info("analysis stored successfully", zap.String("analysis_id", analysis.ID.String()))
    return analysis, nil
}
```

### 4.2 Repository Pattern

```go
// Repository with parameterized queries and error handling
func (s *SQLiteAnalysisStore) CreateAnalysis(ctx context.Context, analysis *Analysis) error {
    const query = `
        INSERT INTO analyses (
            id, project_id, analyzer, analyzed_git_head, analyzed_at,
            summary, purpose, architecture, maturity, strengths, weaknesses,
            reusable_components, notes, raw_json, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

    strengthsJSON, err := json.Marshal(analysis.Strengths)
    if err != nil {
        return fmt.Errorf("failed to marshal strengths: %w", err)
    }

    weaknessesJSON, err := json.Marshal(analysis.Weaknesses)
    if err != nil {
        return fmt.Errorf("failed to marshal weaknesses: %w", err)
    }

    reusableJSON, err := json.Marshal(analysis.ReusableComponents)
    if err != nil {
        return fmt.Errorf("failed to marshal reusable_components: %w", err)
    }

    _, err = s.db.ExecContext(ctx, query,
        analysis.ID,
        analysis.ProjectID,
        analysis.Analyzer,
        analysis.AnalyzedGitHead,
        analysis.AnalyzedAt.Format(time.RFC3339),
        analysis.Summary,
        analysis.Purpose,
        analysis.Architecture,
        analysis.Maturity,
        string(strengthsJSON),
        string(weaknessesJSON),
        string(reusableJSON),
        analysis.Notes,
        string(analysis.RawJSON),
        analysis.CreatedAt.Format(time.RFC3339),
        analysis.UpdatedAt.Format(time.RFC3339),
    )

    if err != nil {
        return fmt.Errorf("failed to insert analysis: %w", err)
    }

    s.logger.Debug("analysis created",
        zap.String("id", analysis.ID.String()),
        zap.String("project_id", analysis.ProjectID.String()),
    )

    return nil
}

func (s *SQLiteAnalysisStore) GetAnalysis(ctx context.Context, projectID uuid.UUID) (*Analysis, error) {
    const query = `
        SELECT id, project_id, analyzer, analyzed_git_head, analyzed_at,
               summary, purpose, architecture, maturity, strengths, weaknesses,
               reusable_components, notes, raw_json, created_at, updated_at
        FROM analyses
        WHERE project_id = ?
        ORDER BY analyzed_at DESC
        LIMIT 1
    `

    row := s.db.QueryRowContext(ctx, query, projectID)

    var analysis Analysis
    var strengthsJSON, weaknessesJSON, reusableJSON string

    err := row.Scan(
        &analysis.ID,
        &analysis.ProjectID,
        &analysis.Analyzer,
        &analysis.AnalyzedGitHead,
        &analysis.AnalyzedAt,
        &analysis.Summary,
        &analysis.Purpose,
        &analysis.Architecture,
        &analysis.Maturity,
        &strengthsJSON,
        &weaknessesJSON,
        &reusableJSON,
        &analysis.Notes,
        &analysis.RawJSON,
        &analysis.CreatedAt,
        &analysis.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil  // Not found, graceful
        }
        return nil, fmt.Errorf("failed to scan analysis: %w", err)
    }

    // Unmarshal JSON arrays
    if err := json.Unmarshal([]byte(strengthsJSON), &analysis.Strengths); err != nil {
        return nil, fmt.Errorf("failed to unmarshal strengths: %w", err)
    }

    if err := json.Unmarshal([]byte(weaknessesJSON), &analysis.Weaknesses); err != nil {
        return nil, fmt.Errorf("failed to unmarshal weaknesses: %w", err)
    }

    if err := json.Unmarshal([]byte(reusableJSON), &analysis.ReusableComponents); err != nil {
        return nil, fmt.Errorf("failed to unmarshal reusable_components: %w", err)
    }

    return &analysis, nil
}
```

### 4.3 Schema Validation Pattern

```go
// JSON schema validation with detailed errors
func (v *SchemaValidator) Validate(input interface{}) error {
    // Convert input to JSON
    inputJSON, err := json.Marshal(input)
    if err != nil {
        return fmt.Errorf("failed to marshal input: %w", err)
    }

    // Load JSON document
    documentLoader := gojsonschema.NewBytesLoader(inputJSON)

    // Validate against schema
    result, err := v.schema.Validate(documentLoader)
    if err != nil {
        return fmt.Errorf("schema validation error: %w", err)
    }

    // Check if valid
    if !result.Valid() {
        var errors []string
        for _, desc := range result.Errors() {
            errors = append(errors, desc.Field()+": "+desc.Description())
        }
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

### 4.4 Stale Detection Pattern

```go
// Single query stale detection with LEFT JOIN
func (d *StaleDetector) ListNeedingAnalysis(ctx context.Context) ([]NeedsAnalysisResult, error) {
    const query = `
        SELECT p.id, p.name,
               CASE 
                 WHEN a.id IS NULL THEN 'unanalyzed'
                 WHEN m.git_head IS NULL OR m.git_head != a.analyzed_git_head THEN 'outdated'
               END as reason
        FROM projects p
        LEFT JOIN metadata m ON m.project_id = p.id
        LEFT JOIN analyses a ON a.project_id = p.id
        WHERE a.id IS NULL 
           OR m.git_head IS NULL 
           OR m.git_head != a.analyzed_git_head
    `

    rows, err := d.store.DB().QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query projects needing analysis: %w", err)
    }
    defer rows.Close()

    var results []NeedsAnalysisResult
    for rows.Next() {
        var result NeedsAnalysisResult
        if err := rows.Scan(&result.ProjectID, &result.Name, &result.Reason); err != nil {
            return nil, fmt.Errorf("failed to scan result: %w", err)
        }
        results = append(results, result)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating results: %w", err)
    }

    return results, nil
}
```

### 4.5 MCP Tool Handler Pattern

```go
// MCP tool handler with error mapping
func HandleStoreAnalysis(service *analysis.AnalysisService) mcp.ToolHandler {
    return func(ctx context.Context, request mcp.ToolCallRequest) (interface{}, error) {
        // Parse request
        var input struct {
            ProjectID string       `json:"project_id"`
            Analysis  AnalysisInput `json:"analysis"`
        }

        if err := json.Unmarshal(request.Params.Arguments, &input); err != nil {
            return nil, &mcp.Error{
                Code:    -32602,  // Invalid params
                Message: fmt.Sprintf("failed to parse request: %v", err),
            }
        }

        // Validate project ID format
        projectID, err := uuid.Parse(input.ProjectID)
        if err != nil {
            return nil, &mcp.Error{
                Code:    -32602,
                Message: fmt.Sprintf("invalid project_id format: %v", err),
            }
        }

        // Call service
        storedAnalysis, err := service.StoreAnalysis(ctx, projectID, input.Analysis)
        if err != nil {
            // Map service errors to MCP errors
            var analysisErr *analysis.Error
            if errors.As(err, &analysisErr) {
                switch analysisErr.Code {
                case analysis.ErrCodeProjectNotFound:
                    return nil, &mcp.Error{
                        Code:    -32602,
                        Message: analysisErr.Message,
                    }
                case analysis.ErrCodeSchemaValidation:
                    return nil, &mcp.Error{
                        Code:    -32602,
                        Message: analysisErr.Message,
                        Data:    map[string]interface{}{"errors": analysisErr.Cause},
                    }
                default:
                    return nil, &mcp.Error{
                        Code:    -32603,  // Internal error
                        Message: analysisErr.Message,
                    }
                }
            }
            return nil, &mcp.Error{
                Code:    -32603,
                Message: fmt.Sprintf("internal error: %v", err),
            }
        }

        // Format response
        return map[string]interface{}{
            "analysis_id": storedAnalysis.ID.String(),
        }, nil
    }
}
```

### 4.6 Error Handling Pattern

```go
// Custom error type with code and context
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"-"`
}

func (e *Error) Error() string {
    return e.Message
}

func (e *Error) Unwrap() error {
    return e.Cause
}

// Error constants
var (
    ErrProjectNotFound     = &Error{Code: ErrCodeProjectNotFound, Message: "Project not found"}
    ErrAnalysisNotFound    = &Error{Code: ErrCodeNotFound, Message: "Analysis not found"}
    ErrInvalidRelationType = &Error{Code: ErrCodeInvalidRelationType, Message: "Invalid relationship type"}
)

// Error code constants
const (
    ErrCodeProjectNotFound     = "PROJECT_NOT_FOUND"
    ErrCodeSchemaValidation    = "SCHEMA_VALIDATION_FAILED"
    ErrCodeNotFound            = "ANALYSIS_NOT_FOUND"
    ErrCodeInvalidRelationType = "INVALID_RELATIONSHIP_TYPE"
    ErrCodeInvalidConfidence   = "INVALID_CONFIDENCE_RANGE"
)

// Usage
func (s *AnalysisService) GetAnalysis(ctx context.Context, projectID uuid.UUID) (*Analysis, error) {
    exists, err := s.store.ProjectExists(ctx, projectID)
    if err != nil {
        return nil, fmt.Errorf("failed to check project existence: %w", err)
    }
    if !exists {
        return nil, ErrProjectNotFound
    }

    analysis, err := s.store.GetAnalysis(ctx, projectID)
    if err != nil {
        return nil, fmt.Errorf("failed to get analysis: %w", err)
    }

    if analysis == nil {
        return nil, nil  // Not found, graceful
    }

    return analysis, nil
}
```

---

## 5. Testing Strategy

### 5.1 Unit Testing

#### Coverage Targets
- `internal/analysis` package: 80%+ coverage
- `internal/repository` package: 90%+ coverage
- `internal/mcp/tools` package: 80%+ coverage
- Overall project: 75%+ coverage

#### Test Structure
```go
func TestAnalysisService_StoreAnalysis_Success(t *testing.T) {
    // Arrange
    mockStore := &MockAnalysisStore{
        projects: map[uuid.UUID]bool{testProjectID: true},
        analyses: make(map[string]*Analysis),
    }
    logger := zaptest.NewLogger(t)
    service := NewAnalysisService(mockStore, logger)

    input := AnalysisInput{
        Summary:     "Test summary",
        Purpose:     "Test purpose",
        Architecture: "Test architecture",
        AnalyzedAt:  time.Now().UTC(),
        AnalyzedGitHead: "abc123",
        Analyzer:    "test-analyzer",
    }

    // Act
    analysis, err := service.StoreAnalysis(context.Background(), testProjectID, input)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, analysis)
    assert.Equal(t, testProjectID, analysis.ProjectID)
    assert.Equal(t, input.Summary, analysis.Summary)
    assert.Equal(t, input.Analyzer, analysis.Analyzer)
}
```

#### Mocking Strategy
```go
// Mock repository for service layer tests
type MockAnalysisStore struct {
    projects map[uuid.UUID]bool
    analyses map[string]*Analysis
    mu       sync.RWMutex
}

func (m *MockAnalysisStore) ProjectExists(ctx context.Context, projectID uuid.UUID) (bool, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    exists := m.projects[projectID]
    return exists, nil
}

func (m *MockAnalysisStore) CreateAnalysis(ctx context.Context, analysis *Analysis) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.analyses[analysis.ID.String()] = analysis
    return nil
}
```

#### Table-Driven Tests
```go
func TestSchemaValidator_Validate(t *testing.T) {
    tests := []struct {
        name      string
        input     AnalysisInput
        wantErr   bool
        errCode   string
    }{
        {
            name: "valid analysis",
            input: AnalysisInput{
                Summary:     "Valid summary",
                Purpose:     "Valid purpose",
                Architecture: "Valid architecture",
                AnalyzedAt:  time.Now().UTC(),
                AnalyzedGitHead: "abc123",
                Analyzer:    "test-analyzer",
            },
            wantErr: false,
        },
        {
            name: "missing required field",
            input: AnalysisInput{
                Purpose:     "Valid purpose",
                Architecture: "Valid architecture",
                AnalyzedAt:  time.Now().UTC(),
                AnalyzedGitHead: "abc123",
                Analyzer:    "test-analyzer",
            },
            wantErr: true,
            errCode: ErrCodeSchemaValidation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := NewSchemaValidator()
            err := validator.Validate(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                var analysisErr *Error
                assert.True(t, errors.As(err, &analysisErr))
                assert.Equal(t, tt.errCode, analysisErr.Code)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 5.2 Integration Testing

#### Test Database Setup
```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)

    // Run migrations
    migrationSQL := `
        CREATE TABLE projects (id TEXT PRIMARY KEY, name TEXT, path TEXT);
        CREATE TABLE metadata (project_id TEXT PRIMARY KEY, git_head TEXT);
        CREATE TABLE analyses (...);
        CREATE TABLE features (...);
        CREATE TABLE relationships (...);
    `
    _, err = db.Exec(migrationSQL)
    require.NoError(t, err)

    return db
}
```

#### Full Workflow Test
```go
func TestFullAnalysisWorkflow(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    store := NewSQLiteAnalysisStore(db, zaptest.NewLogger(t))
    service := NewAnalysisService(store, zaptest.NewLogger(t))

    // Create test project
    projectID := uuid.New()
    _, err := db.Exec("INSERT INTO projects (id, name, path) VALUES (?, ?, ?)",
        projectID, "test-project", "/tmp/test")
    require.NoError(t, err)

    // Store analysis
    input := AnalysisInput{
        Summary:     "Test summary",
        Purpose:     "Test purpose",
        Architecture: "Test architecture",
        AnalyzedAt:  time.Now().UTC(),
        AnalyzedGitHead: "abc123",
        Analyzer:    "claude-code",
        Features: []FeatureInput{
            {Name: "Authentication", Confidence: float64Ptr(0.95)},
        },
    }

    analysis, err := service.StoreAnalysis(context.Background(), projectID, input)
    require.NoError(t, err)
    assert.NotNil(t, analysis)

    // Retrieve analysis
    retrieved, err := service.GetAnalysis(context.Background(), projectID)
    require.NoError(t, err)
    assert.NotNil(t, retrieved)
    assert.Equal(t, analysis.ID, retrieved.ID)
    assert.Equal(t, input.Summary, retrieved.Summary)
    assert.Len(t, retrieved.Features, 1)
}
```

### 5.3 Performance Testing

#### Benchmark Tests
```go
func BenchmarkStoreAnalysis(b *testing.B) {
    db := setupTestDB(b)
    store := NewSQLiteAnalysisStore(db, zap.NewNop())
    service := NewAnalysisService(store, zap.NewNop())

    projectID := uuid.New()
    _, _ = db.Exec("INSERT INTO projects (id, name, path) VALUES (?, ?, ?)",
        projectID, "test", "/tmp")

    input := AnalysisInput{
        Summary:     "Test summary",
        Purpose:     "Test purpose",
        Architecture: "Test architecture",
        AnalyzedAt:  time.Now().UTC(),
        AnalyzedGitHead: "abc123",
        Analyzer:    "claude-code",
        Features:    make([]FeatureInput, 10),
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = service.StoreAnalysis(context.Background(), projectID, input)
    }
}
```

#### Performance Targets
- `storeAnalysis()`: < 100ms (typical: 10 features, 2KB text)
- `getAnalysis()`: < 50ms (1000 analyses in database)
- `listProjectsNeedingAnalysis()`: < 200ms (500 projects)
- `listRelationships()`: < 100ms (50 relationships)
- Schema validation: < 10ms

### 5.4 Security Testing

#### Input Validation Tests
```go
func TestSQLInjectionPrevention(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    store := NewSQLiteAnalysisStore(db, zaptest.NewLogger(t))
    service := NewAnalysisService(store, zaptest.NewLogger(t))

    projectID := uuid.New()
    _, _ = db.Exec("INSERT INTO projects (id, name, path) VALUES (?, ?, ?)",
        projectID, "test", "/tmp")

    // Attempt SQL injection
    maliciousInput := AnalysisInput{
        Summary: "'; DROP TABLE analyses; --",
        Purpose:     "Test purpose",
        Architecture: "Test architecture",
        AnalyzedAt:  time.Now().UTC(),
        AnalyzedGitHead: "abc123",
        Analyzer:    "test-analyzer",
    }

    _, err := service.StoreAnalysis(context.Background(), projectID, maliciousInput)
    // Should succeed (content is stored, not executed)
    require.NoError(t, err)

    // Verify table still exists
    _, err = db.Query("SELECT COUNT(*) FROM analyses")
    assert.NoError(t, err)  // Table still exists
}
```

### 5.5 Test Execution

#### Run All Tests
```bash
# Unit tests
go test ./internal/analysis -v -cover

# Integration tests
go test ./internal/analysis -v -tags=integration

# Benchmark tests
go test ./internal/analysis -bench=. -benchmem

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Continuous Integration
```bash
# Run linting
golangci-lint run

# Run tests with race detection
go test ./... -race -count=1

# Check coverage threshold
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{if ($3 < 75.0) exit 1}'
```

---

## 6. Build and Verification Steps

### 6.1 Build Steps

#### 1. Implement Story 10.1
```bash
# Create type definitions
touch internal/analysis/types.go
touch pkg/models/analysis.go
touch internal/analysis/errors.go
touch internal/analysis/repository.go

# Implement schema validator
touch internal/analysis/schema_validator.go
touch internal/analysis/schema_validator_test.go

# Run tests
go test ./internal/analysis -v -run TestSchemaValidator
```

#### 2. Implement Story 10.2
```bash
# Create database migration
touch migrations/001_add_analysis_tables.sql
sqlite3 portfolio.db < migrations/001_add_analysis_tables.sql

# Implement repository layer
touch repository/queries.go
touch repository/analysis_store.go
touch repository/analysis_store_test.go

# Implement service layer
touch internal/analysis/service.go
touch internal/analysis/service_test.go

# Implement MCP tools
touch internal/mcp/tools/analysis.go
touch internal/mcp/tools/analysis_test.go

# Run tests
go test ./internal/analysis ./repository ./internal/mcp/tools -v
```

#### 3. Implement Story 10.3
```bash
# Implement stale detector
touch internal/analysis/stale_detector.go
touch internal/analysis/stale_detector_test.go

# Update repository
# (Add GetGitHeadForProject, ListAllAnalyses)

# Update MCP tools
# (Add listProjectsNeedingAnalysis tool)

# Run tests
go test ./internal/analysis -v -run TestStaleDetector
go test ./internal/analysis -bench=ListNeedingAnalysis
```

#### 4. Implement Story 10.4
```bash
# Update database migration
# (Add relationships table)

# Update repository layer
# (Add relationship CRUD methods)

# Implement relationship service
touch internal/analysis/relationship.go
touch internal/analysis/relationship_test.go

# Implement relationship MCP tools
touch internal/mcp/tools/relationships.go
touch internal/mcp/tools/relationships_test.go

# Run tests
go test ./internal/analysis ./repository ./internal/mcp/tools -v
```

### 6.2 Verification Steps

#### Functional Verification
1. **Schema Validation**
   - Test valid analysis payload is accepted
   - Test invalid payload is rejected with detailed errors
   - Test additional fields preserved in raw_json

2. **Analysis Storage**
   - Test create new analysis
   - Test update existing analysis
   - Test multiple analyzers per project
   - Test feature storage and retrieval

3. **Stale Detection**
   - Test outdated analysis detection
   - Test unanalyzed project detection
   - Test NULL git_head handling
   - Verify single query execution

4. **Relationship Persistence**
   - Test create relationship
   - Test list relationships (both directions)
   - Test delete relationship
   - Test validation (type, confidence, FK)

#### Integration Verification
1. **MCP Tool Workflow**
   - Start MCP server
   - Call storeAnalysis tool
   - Call getAnalysis tool
   - Call listProjectsNeedingAnalysis tool
   - Call listRelationships tool
   - Verify end-to-end flow

2. **Database Integrity**
   - Verify foreign key constraints
   - Verify cascade deletes
   - Verify uniqueness constraints
   - Verify CHECK constraints

3. **Performance Verification**
   - Run benchmark tests
   - Verify all performance targets met
   - Profile slow queries if needed

#### Security Verification
1. **Input Validation**
   - Test SQL injection attempts
   - Test malformed JSON
   - Test invalid UUID format
   - Test invalid relationship types
   - Test confidence range validation

2. **Error Handling**
   - Test database connection loss
   - Test transaction rollback
   - Test concurrent writes
   - Test corrupted data handling

### 6.3 Quality Gates

#### Pre-Commit Checklist
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Code coverage >= 75% overall, 80% for analysis package
- [ ] No linting errors (`golangci-lint run`)
- [ ] No race conditions (`go test -race`)
- [ ] Performance benchmarks meet targets
- [ ] Security tests pass

#### Story Acceptance Checklist

**Story 10.1 — Analysis Schema**
- [ ] JSON schema defined with all required fields
- [ ] Schema validator implemented
- [ ] Valid payloads accepted
- [ ] Invalid payloads rejected with detailed errors
- [ ] raw_json preserves additional fields
- [ ] Unit tests pass (100% coverage)
- [ ] Integration test passes

**Story 10.2 — Persist Analyses**
- [ ] Database schema created
- [ ] Repository implemented
- [ ] Service layer implemented
- [ ] MCP tools implemented
- [ ] Create analysis works
- [ ] Update analysis works
- [ ] Get analysis works
- [ ] Features handled correctly
- [ ] Unit tests pass (80% coverage)
- [ ] Integration tests pass
- [ ] Performance targets met

**Story 10.3 — Detect Stale Analyses**
- [ ] Stale detector implemented
- [ ] Outdated analysis detected
- [ ] Unanalyzed projects detected
- [ ] NULL git_head handled
- [ ] Single query verified
- [ ] MCP tool implemented
- [ ] Unit tests pass (100% coverage)
- [ ] Performance target met (<200ms for 500 projects)

**Story 10.4 — Relationship Persistence**
- [ ] Database schema extended
- [ ] Repository extended
- [ ] Relationship service implemented
- [ ] MCP tools implemented
- [ ] Create relationship works
- [ ] List relationships works
- [ ] Delete relationship works
- [ ] Validation enforced
- [ ] Unit tests pass (80% coverage)
- [ ] Integration tests pass
- [ ] Performance target met (<100ms)

#### Epic Acceptance Checklist
- [ ] All stories accepted
- [ ] All acceptance criteria met
- [ ] All test cases pass (123 tests)
- [ ] Code coverage targets met
- [ ] Performance targets met
- [ ] Security tests pass
- [ ] Documentation updated
- [ ] No blocking issues
- [ ] Ready for integration with Epic 11

---

## 7. Quality Gates

### 7.1 Code Quality

#### Static Analysis
```bash
# Run linter
golangci-lint run --config .golangci.yml

# Expected: No errors, warnings addressed
```

#### Code Coverage
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Expected thresholds:
# - internal/analysis: 80%+
# - internal/repository: 90%+
# - internal/mcp/tools: 80%+
# - Overall: 75%+
```

#### Race Detection
```bash
# Run tests with race detector
go test ./... -race -count=1

# Expected: No race conditions detected
```

### 7.2 Performance Gates

#### Benchmark Results
```bash
# Run benchmarks
go test ./internal/analysis -bench=. -benchmem

# Expected thresholds:
# - BenchmarkStoreAnalysis: < 100ms/op
# - BenchmarkGetAnalysis: < 50ms/op
# - BenchmarkListNeedingAnalysis: < 200ms/op (500 projects)
# - BenchmarkListRelationships: < 100ms/op (50 relationships)
# - BenchmarkSchemaValidation: < 10ms/op
```

#### Query Performance
```bash
# Enable query logging and verify single query execution
# Expected: listProjectsNeedingAnalysis() executes exactly 1 query
```

### 7.3 Security Gates

#### Input Validation
- [ ] All inputs validated before processing
- [ ] SQL injection attempts blocked
- [ ] Malformed JSON rejected
- [ ] Invalid UUID formats rejected
- [ ] Enum values enforced

#### Error Handling
- [ ] Errors never expose sensitive information
- [ ] Database errors wrapped with context
- [ ] Stack traces not exposed in production
- [ ] Transactions rolled back on error

#### Data Integrity
- [ ] Foreign key constraints enforced
- [ ] Cascade deletes tested
- [ ] Uniqueness constraints tested
- [ ] CHECK constraints tested

### 7.4 Functional Gates

#### Schema Validation
- [ ] Valid analysis payloads accepted
- [ ] Invalid payloads rejected with detailed errors
- [ ] All required fields validated
- [ ] ISO 8601 timestamps validated
- [ ] Array fields validated
- [ ] Additional fields preserved

#### Analysis Persistence
- [ ] New analyses created
- [ ] Existing analyses updated
- [ ] Multiple analyzers supported
- [ ] Features stored and retrieved
- [ ] Raw JSON preserved
- [ ] Timestamps accurate

#### Stale Detection
- [ ] Outdated analyses identified
- [ ] Unanalyzed projects identified
- [ ] NULL git_head handled
- [ ] Single query execution
- [ ] Performance target met

#### Relationship Persistence
- [ ] Relationships created
- [ ] Relationships listed (both directions)
- [ ] Relationships deleted
- [ ] Type validation enforced
- [ ] Confidence validation enforced
- [ ] Foreign key constraints enforced
- [ ] Deduplication works

### 7.5 Integration Gates

#### MCP Tools
- [ ] All tools registered
- [ ] Tool schemas correct
- [ ] Handlers parse input correctly
- [ ] Handlers call service methods
- [ ] Handlers format output correctly
- [ ] Errors mapped to MCP error codes
- [ ] End-to-end workflow tested

#### Database Operations
- [ ] Migration applies successfully
- [ ] All queries execute correctly
- [ ] Transactions work correctly
- [ ] Connection pooling configured
- [ ] Cascade deletes work
- [ ] Constraints enforced

#### Service Layer
- [ ] Methods accept context
- [ ] Methods respect context cancellation
- [ ] Errors wrapped with context
- [ ] Logging comprehensive
- [ ] Methods are deterministic

---

## 8. Dependencies and Blockers

### 8.1 External Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `go.uber.org/zap` | 1.24+ | Structured logging |
| `github.com/mattn/go-sqlite3` | 1.14+ | SQLite driver |
| `github.com/google/uuid` | 1.3+ | UUID generation |
| `github.com/xeipuuv/gojsonschema` | 1.2+ | JSON schema validation |
| `github.com/mark3labs/mcp-go` | 0.1+ | MCP protocol |

### 8.2 Internal Dependencies

| Dependency | Epic/Component | Status |
|------------|---------------|--------|
| Epic 2 — Discovery | `projects` table | ✅ Complete |
| Epic 3 — Metadata | `metadata.git_head` | ✅ Complete |
| Epic 5 — Repository | SQLite store pattern | ✅ Complete |
| Epic 7 — MCP Server | Tool registration | ✅ Complete |
| Epic 8 — Integration | `analyzer` identity | ✅ Complete |

### 8.3 Blockers

**Story 10.1:**
- ❌ Blocked by: Story 5.1 (MCP Server implementation)
- ✅ Ready when: MCP server tool registration is available

**Story 10.2:**
- ❌ Blocked by: Story 10.1, Story 7.4 (Documentation indexing)
- ✅ Ready when: Analysis schema defined and documentation indexing complete

**Story 10.3:**
- ❌ Blocked by: Story 3.1, Story 10.2
- ✅ Ready when: Git metadata extraction and analysis persistence complete

**Story 10.4:**
- ❌ Blocked by: Story 10.1
- ✅ Ready when: Analysis schema defined

---

## 9. Risk Mitigation

### 9.1 Technical Risks

#### Risk: Schema Validation Performance
**Mitigation:**
- Compile JSON schema once at startup
- Reuse validator instance across calls
- Benchmark validation performance
- Target: < 10ms per validation

#### Risk: Database Query Performance
**Mitigation:**
- Create proper indexes on foreign keys
- Use LEFT JOIN for stale detection (single query)
- Batch feature inserts
- Benchmark with realistic dataset (500 projects)

#### Risk: Concurrent Write Conflicts
**Mitigation:**
- Rely on SQLite write serialization
- Use UNIQUE constraints for deduplication
- Test with concurrent operations
- Document last-write-wins behavior

### 9.2 Integration Risks

#### Risk: MCP Protocol Compatibility
**Mitigation:**
- Follow MCP protocol specification
- Test with real MCP client
- Validate tool schemas
- Handle protocol errors gracefully

#### Risk: Schema Evolution
**Mitigation:**
- Store raw JSON for flexibility
- Use CHECK constraints for invariants
- Document migration strategy
- Test with future fields in raw_json

### 9.3 Quality Risks

#### Risk: Insufficient Test Coverage
**Mitigation:**
- Target 80%+ coverage
- Include edge cases in tests
- Use table-driven tests
- Continuous integration enforcement

#### Risk: Security Vulnerabilities
**Mitigation:**
- Parameterized queries only
- Input validation at all layers
- Security test suite
- Error message sanitization

---

## 10. Success Criteria

### 10.1 Functional Success
- ✅ AI agents can store analyses via MCP tools
- ✅ AI agents can retrieve analyses via MCP tools
- ✅ AI agents can detect stale analyses
- ✅ AI agents can store relationships
- ✅ Schema validation prevents invalid data
- ✅ Multiple analyzers supported per project
- ✅ Stale detection uses single query

### 10.2 Non-Functional Success
- ✅ Performance targets met (< 100ms store, < 50ms get, < 200ms list)
- ✅ Code coverage targets met (80%+ analysis package)
- ✅ No SQL injection vulnerabilities
- ✅ No race conditions
- ✅ No data corruption
- ✅ Graceful error handling

### 10.3 Integration Success
- ✅ MCP tools registered and functional
- ✅ End-to-end workflow tested
- ✅ Database migrations apply successfully
- ✅ Compatible with Epic 9 (Claude Code integration)
- ✅ Ready for Epic 11 (Agent workflows)

---

## 11. Next Steps

1. **Review and Approve:**
   - Review implementation guidelines with team
   - Get approval on technical standards
   - Confirm dependency blockers resolved

2. **Setup Development Environment:**
   - Ensure Go 1.21+ installed
   - Install dependencies: `go mod tidy`
   - Setup test database: `sqlite3 portfolio.db < migrations/001_add_analysis_tables.sql`

3. **Begin Story 10.1:**
   - Implement type definitions
   - Implement schema validator
   - Write unit tests
   - Verify schema validation works

4. **Continue Story 10.2:**
   - Create database migration
   - Implement repository layer
   - Implement service layer
   - Implement MCP tools
   - Write integration tests

5. **Continue Story 10.3:**
   - Implement stale detector
   - Optimize query performance
   - Implement MCP tool
   - Write performance tests

6. **Continue Story 10.4:**
   - Update database schema
   - Implement relationship service
   - Implement relationship MCP tools
   - Write integration tests

7. **Final Verification:**
   - Run all tests (123 test cases)
   - Verify code coverage
   - Run performance benchmarks
   - Run security tests
   - Complete acceptance checklists

8. **Handoff to Epic 11:**
   - Document MCP tool contracts
   - Provide example workflows
   - Prepare integration guidelines

---

This implementation guideline provides a comprehensive roadmap for implementing Epic 10 — AI Analysis. Following these standards, patterns, and verification steps will ensure a high-quality, performant, and secure implementation that adheres to Portfolio's engineering principles.