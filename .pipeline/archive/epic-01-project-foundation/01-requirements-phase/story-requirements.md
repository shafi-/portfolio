# Requirements: Story 1.1 — Bootstrap Go Project

**Epic:** 1 — Project Foundation  
**Story:** 1.1 — Bootstrap Go Project  
**Version:** 1.0  
**Status:** Draft  

---

## Overview

These requirements define the initial project setup for the Portfolio Engine, a greenfield Go project that will discover and analyze projects in user-defined directories. The bootstrap must establish a foundation that supports the engine's deterministic operations, local-first architecture, and future integration with AI agents through MCP and HTTP APIs.

---

## Core Requirements

### REQ-1.1.1: Go Module Initialization [MUST HAVE]

The project must be initialized as a proper Go module following Go community conventions.

**Acceptance Criteria:**
- Go module file (`go.mod`) exists at repository root
- Module name follows Go naming conventions (reverse DNS notation)
- Go version specified is stable and appropriate for production use
- Module name supports import paths for all planned packages (cmd, internal, pkg)
- No implicit dependencies or undeclared imports

**Verification:**
```bash
go mod verify
go list -m all
go build ./...
```

---

### REQ-1.1.2: Standard Project Structure [MUST HAVE]

The project must follow standard Go project layout conventions to support maintainability and tooling expectations.

**Acceptance Criteria:**
- `cmd/` directory exists for main applications
- `internal/` directory exists for private application code
- `pkg/` directory exists for public library code (if applicable)
- `docs/` directory exists for project documentation (already exists)
- `.git/` directory indicates proper git initialization
- No source files exist at repository root

**Directory Structure:**
```
portfolio-tool/
├── cmd/             # Main applications (portfolio-cli, portfolio-engine)
├── internal/        # Private application code
├── pkg/             # Public libraries (future use)
├── docs/            # Project documentation (existing)
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── .gitignore
```

**Verification:**
```bash
tree -L 1 -d
test -d cmd && test -d internal && test -d pkg
```

---

### REQ-1.1.3: Git Configuration [MUST HAVE]

The repository must have appropriate git configuration for Go development.

**Acceptance Criteria:**
- `.gitignore` file exists and is properly configured for Go projects
- `.gitignore` includes standard Go exclusions (binaries, cache, test files)
- `.gitignore` includes project-specific exclusions (SQLite databases, config files)
- `.gitignore` supports common Go IDEs and tools
- Repository is initialized with proper git metadata

**Required .gitignore entries:**
```goignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDE-specific files
.vscode/
.idea/
*.swp
*.swo
*~

# Portfolio-specific data
*.db
*.sqlite
*.sqlite3
.config/
portfolio/

# OS-specific files
.DS_Store
Thumbs.db
```

**Verification:**
```bash
git status
git check-ignore bin/portfolio
```

---

### REQ-1.1.4: LICENSE File [MUST HAVE]

The project must include a properly formatted LICENSE file to establish usage terms and comply with open-source best practices.

**Acceptance Criteria:**
- LICENSE file exists at repository root
- License type is specified and appropriate for the project goals
- License includes copyright notice
- License includes standard license text
- Year in copyright notice is current

**Ambiguity:** 
- **Question:** What license type should be used for the Portfolio Engine?
- **Suggested Default:** MIT License (permissive, widely used in Go projects, suitable for local-first infrastructure)
- **Alternatives:** Apache 2.0 (patent protection), BSD 3-Clause (similar to MIT), GPL (copyleft)

**Verification:**
```bash
test -f LICENSE
head -n 5 LICENSE
```

---

### REQ-1.1.5: README Documentation [MUST HAVE]

The project must have comprehensive README documentation that enables developers to understand and build the project.

**Acceptance Criteria:**
- README.md exists at repository root
- README includes project name and brief description
- README includes installation instructions
- README includes build instructions
- README includes run instructions for the CLI
- README includes project status and goals alignment
- README references authoritative documents (PRD, Architecture, etc.)
- README includes contribution guidelines (can reference external docs)

**Required README Sections:**
1. **Project Name & Description** — What is Portfolio?
2. **Quick Start** — Installation and basic usage
3. **Development Setup** — Building and running from source
4. **Documentation Links** — Links to PRD, Architecture, etc.
5. **Project Status** — Current development phase
6. **License** — Reference to LICENSE file

**Verification:**
```bash
test -f README.md
grep -q "Portfolio" README.md
grep -q "Installation" README.md
grep -q "Build" README.md
```

---

### REQ-1.1.6: Build Configuration [SHOULD HAVE]

The project should support standard Go build commands and tooling.

**Acceptance Criteria:**
- `go build ./cmd/...` successfully builds all applications
- `go test ./...` can run tests (even if none exist yet)
- `go vet ./...` can check code
- `go fmt ./...` can format code
- Build produces executables in expected locations

**Verification:**
```bash
go build -v ./...
go test -v ./...
go vet ./...
```

---

### REQ-1.1.7: Development Environment [SHOULD HAVE]

The project should provide clear guidance on development environment setup.

**Acceptance Criteria:**
- README specifies minimum Go version required
- README lists any required development tools
- README explains how to set up the development environment
- README provides troubleshooting steps for common issues

**Ambiguity:**
- **Question:** What is the minimum Go version requirement?
- **Suggested Default:** Go 1.21+ (recent stable version with good standard library support)
- **Rationale:** Balance between modern features and broad compatibility

---

### REQ-1.1.8: Package Organization Planning [COULD HAVE]

The project structure should anticipate future package organization needs.

**Acceptance Criteria:**
- `internal/` directory structure anticipates major components:
  - `internal/config/` — Configuration management
  - `internal/database/` — SQLite operations
  - `internal/discovery/` — Project discovery
  - `internal/metadata/` — Metadata extraction
  - `internal/docs/` — Documentation indexing
  - `internal/search/` — Search functionality
  - `internal/api/` — HTTP API
  - `internal/mcp/` — MCP server
- `cmd/` directory anticipates main applications:
  - `cmd/portfolio/` — Main CLI application
- Package names follow Go conventions (lowercase, no underscores)

**Verification:**
```bash
ls -la internal/ # Expected: empty but planned structure documented
ls -la cmd/      # Expected: empty but planned structure documented
```

---

## Architecture Alignment Requirements

### REQ-1.1.9: Alignment with Go Best Practices [MUST HAVE]

The bootstrap must align with Go community best practices and idioms.

**Acceptance Criteria:**
- Project follows standard Go project layout
- Module naming follows Go conventions
- Package naming follows Go conventions (lowercase, single words when possible)
- No unnecessary subdirectories at bootstrap stage
- Preparation for future dependency management policies (per Guideline.md)

---

### REQ-1.1.10: Support for Planned Components [MUST HAVE]

The project structure must support the planned components from Architecture.md.

**Acceptance Criteria:**
- Structure supports Portfolio Engine (Go implementation)
- Structure supports CLI interface (cmd/portfolio)
- Structure supports future MCP integration (internal/mcp)
- Structure supports future HTTP API (internal/api)
- Structure supports local SQLite knowledge store (internal/database)

**Planned Components (from Architecture.md):**
- Portfolio Engine — Core deterministic operations
- MCP Interface — AI agent integration
- HTTP API — Dashboard communication
- CLI — Administrative interface
- Local Knowledge Store — SQLite persistence

---

### REQ-1.1.11: Minimal Dependencies Principle [SHOULD HAVE]

The bootstrap should follow the "Minimize Dependencies" principle from Guideline.md.

**Acceptance Criteria:**
- No unnecessary dependencies in go.mod
- Prefer standard library over third-party packages when possible
- Document reasons for any dependencies added during bootstrap
- Prepare for dependency management policies in future stories

**Ambiguity:**
- **Question:** What initial dependencies are acceptable for bootstrap?
- **Suggested Default:** None (pure Go standard library only)
- **Rationale:** Dependencies should be added in specific stories when needed (e.g., Story 1.3 for logging, Story 1.4 for CLI framework)

---

## Testing & Quality Requirements

### REQ-1.1.12: Test Framework Setup [COULD HAVE]

The project could establish testing infrastructure during bootstrap.

**Acceptance Criteria:**
- `go test` works without errors (even with no tests)
- Table-driven test pattern documented for future use
- Test naming conventions documented
- Setup for integration tests anticipated

**Ambiguity:**
- **Question:** Should testing infrastructure be established now or deferred to first feature implementation?
- **Suggested Default:** Defer to Story 1.2 (Configuration) or first feature implementation
- **Rationale:** Bootstrap focuses on structure; testing infrastructure adds value when writing actual tests

---

## Documentation Requirements

### REQ-1.1.13: Code Documentation Preparation [COULD HAVE]

The project could establish code documentation patterns.

**Acceptance Criteria:**
- Godoc comment style documented for future use
- Package documentation guidelines referenced
- Example package documentation provided (even if placeholder)
- Exported function comment conventions established

**Verification:**
```bash
godoc -http=:6060 # Should work once packages exist
```

---

## Security & Privacy Requirements

### REQ-1.1.14: Privacy-First Setup [MUST HAVE]

The bootstrap must support local-first architecture and privacy principles.

**Acceptance Criteria:**
- No cloud services configured or required
- No telemetry or analytics included
- No external network dependencies for basic operations
- .gitignore protects local data (config files, databases)

**Alignment with PRD.md:**
- "Local-first" core principle
- Repositories and knowledge remain on user's machine
- Zero cloud services required

---

## Build & Distribution Requirements

### REQ-1.1.15: Cross-Platform Build Preparation [COULD HAVE]

The project structure should anticipate cross-platform distribution.

**Acceptance Criteria:**
- Platform-agnostic Go code practices documented
- Build commands for multiple platforms documented in README
- Preparation for future release automation noted

**Ambiguity:**
- **Question:** Which platforms should be prioritized for cross-platform support?
- **Suggested Default:** Linux (primary), macOS (secondary), Windows (tertiary)
- **Rationale:** Aligns with typical developer tool usage and Go's strong cross-platform support

---

## Open Questions & Ambiguities

### Question 1: Module Naming
**What should the Go module name be?**
- **Context:** Must support import paths for all planned packages
- **Options:** 
  - `github.com/username/portfolio`
  - `github.com/username/portfolio-tool`
  - Custom domain-based path
- **Recommendation:** Use repository URL as module name (standard Go practice)

### Question 2: Minimum Go Version
**What is the minimum Go version requirement?**
- **Context:** Balance modern features with broad compatibility
- **Options:** Go 1.20, Go 1.21, Go 1.22
- **Recommendation:** Go 1.21+ (recent stable with good library support)

### Question 3: License Type
**What license should be used?**
- **Context:** Affects contribution, distribution, and usage
- **Options:** MIT, Apache 2.0, BSD 3-Clause, GPL
- **Recommendation:** MIT License (permissive, widely used in Go ecosystem)

### Question 4: Initial Dependencies
**Should any dependencies be included during bootstrap?**
- **Context:** Balance between functionality and "minimal dependencies" principle
- **Options:** None, basic logging, basic CLI framework
- **Recommendation:** None — defer to specific stories (1.3 for logging, 1.4 for CLI)

### Question 5: Platform Priority
**Which platforms should be prioritized for build and distribution?**
- **Context:** Affects testing, CI/CD, and release planning
- **Options:** All platforms equally, Linux-first, macOS-first
- **Recommendation:** Linux (primary), macOS (secondary), Windows (tertiary)

---

## Success Criteria

The bootstrap is considered complete when:

1. **Structure:** Standard Go project layout exists (cmd/, internal/, pkg/)
2. **Module:** Go module initialized with appropriate naming
3. **Configuration:** .gitignore protects build artifacts and local data
4. **Legal:** LICENSE file establishes usage terms
5. **Documentation:** README enables developers to build and run the project
6. **Alignment:** Structure supports all planned components from Architecture.md
7. **Verification:** All verification commands pass successfully
8. **Foundation:** Project is ready for Story 1.2 (Configuration System)

---

## Dependencies

**Prerequisites:**
- Git installed and configured
- Go 1.21+ installed
- Write access to repository directory

**Blocked by:** None
**Blocks:** Story 1.2 (Configuration System)

---

## References

- **PRD.md** — Product vision and goals
- **Architecture.md** — System design and components
- **Guideline.md** — Engineering principles (minimize dependencies)
- **PlatformSpecification.md** — Technical implementation contracts
- **TechStack.md** — Go as engine language, SQLite for storage
- **UXGuidelines.md** — Dashboard patterns (not applicable to bootstrap)

---

## Notes

- This bootstrap establishes the foundation for all subsequent development
- The structure must accommodate both CLI and HTTP interfaces
- MCP integration planning should begin immediately after bootstrap
- Future stories will add functionality incrementally on this foundation
- All architectural decisions made during bootstrap should be captured in ADR.md