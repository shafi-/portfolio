
# ADR.md

# Architecture Decision Records

This document contains incremental architecture decisions for Portfolio.

---

## ADR-013: Agent Integrations are First-Class Components

**Status:** Accepted

### Context

The Portfolio Engine is designed to be AI-agent agnostic. While Claude Code is the
first supported agent, the architecture should support additional coding agents
without changing the engine.

Examples include:

- Claude Code
- Codex CLI
- OpenCode
- Cursor
- Future MCP-compatible agents

### Decision

Agent-specific behavior shall be implemented as installable integrations rather
than embedded into the engine.

An integration is responsible for:

- Registering the MCP server
- Installing agent-specific skills or instructions
- Validating the installation
- Upgrading the integration
- Removing the integration

The engine exposes deterministic capabilities only and has no knowledge of a
specific AI agent.

### Consequences

Positive:

- Engine remains agent-agnostic.
- New AI agents can be supported independently.
- Agent-specific prompting and workflows evolve separately from the engine.

Negative:

- Each supported agent requires its own integration package.

---

## ADR-014: Install → Initialize → Forget

**Status:** Accepted

### Context

Developer tools often require users to remember commands and perform ongoing
maintenance.

Portfolio aims to be invisible infrastructure rather than another tool that
demands attention.

### Decision

Portfolio follows a simple lifecycle:

```text
Install
    ↓
Initialize
    ↓
Forget
```

The CLI exists primarily for:

- Initialization
- Diagnostics
- Upgrades
- Integration management

After initialization, the primary interaction is through an AI coding agent.

### Consequences

Positive:

- Minimal user friction.
- Natural AI-first workflow.
- Reduced operational complexity for users.

Negative:

- Greater emphasis on robust agent integrations.

---

## ADR-015: KnowledgeModel is the Canonical Source of Truth

**Status:** Accepted

### Context

As the platform grows, multiple interfaces (database, MCP, HTTP API, dashboard,
and AI agents) require a consistent understanding of Portfolio concepts.

### Decision

`KnowledgeModel.md` is the canonical definition of the domain model.

`PlatformSpecification.md` defines how that model is implemented.

All implementations must derive from these documents rather than redefining
entities independently.

### Consequences

Positive:

- Consistent domain language.
- Easier evolution of the platform.
- Reduced duplication across documentation and code.

Negative:

- Changes to the domain model require synchronized updates to downstream
specifications.

---

## ADR-016: Official Methods Only for Agent Integrations

**Status:** Accepted

### Context

Portfolio integrations need to register MCP servers with various AI coding agents.
Different agents have different approaches to MCP server registration:

- **Claude Code**: Provides official CLI commands (`claude mcp add/remove/get`)
- **OpenCode**: Partial support (remote servers only via CLI, local requires config editing)
- **Cline**: No CLI support (requires manual config editing)

Initial implementation attempted direct config file manipulation, which created several problems:

1. **Fragility**: Config formats change between tool versions
2. **Maintenance burden**: Breaking changes require immediate fixes
3. **User trust**: Direct file editing feels unsafe
4. **No official support**: Tools don't guarantee config format stability

### Decision

All Portfolio integrations MUST use official tool methods for MCP server registration.

**Requirements:**

1. **Use official CLI**: When available, integrations must use official CLI commands
2. **No direct config editing**: Production code must never directly edit agent config files
3. **Transparent fallback**: For tools without official methods, provide unsafe scripts with warnings
4. **Documentation first**: Manual setup documentation takes precedence over automation

**Implementation Approach:**

| Tool Status | Integration Approach |
|-------------|---------------------|
| Official CLI exists | ✅ Create automated integration using official commands |
| Partial official support | ⚠️ Document limitations, provide unsafe scripts with warnings |
| No official support | ❌ Provide manual setup docs and unsafe scripts only |

**For Tools Without Official Methods:**

- Document manual setup in `docs/integration-guideline.md`
- Create `scripts/unsafe-<tool>-integration.sh` with:
  - Clear warnings that script is unofficial and unsafe
  - User consent requirements
  - Automatic backups before changes
  - Fully visible and reviewable code
- Never embed config manipulation in production code

### Consequences

**Positive:**

- Integrations remain stable across tool updates
- Users trust the integration process
- Reduced maintenance burden
- Clear user expectations

**Negative:**

- Some tools require manual setup instead of full automation
- Users must understand risks for unsafe scripts
- Cannot provide "perfect" automation for every tool

**Implementation Notes:**

- Claude Code integration uses `claude mcp add/remove/get` commands
- OpenCode has `opencode mcp add` but only for remote servers
- Cline requires manual `~/.cline/mcp.json` editing
- All unsafe scripts live in `scripts/` with clear README documentation

---

## ADR-017: Deterministic Importance Signals

**Status:** Accepted

### Context

Projects can be ranked by "importance" without any AI analysis — but only if the
Engine persists enough deterministic signal. Before this decision, `metadata` was
`null` for every project: extractors existed and were tested, but nothing in the
production scan path called them, and the indexer's only metadata write was a
two-field `INSERT OR REPLACE` that clobbered every other column.

This violated two guiding principles: "Engine Knows, Agent Thinks" (the Engine
should surface all the deterministic facts it can) and "Store Facts, Compute
Indicators" (persist immutable facts; derive rankings at read time).

### Decision

Add a set of deterministic, LLM-free signals and make the Engine actually
populate them:

- **Git investment:** `first_commit_at`, `commit_velocity_90d`, `contributor_count`,
  `tag_count`, `remote_url`, `is_published`.
- **Dependency scope:** `scope` (`prod`/`dev`) on each dependency row, capturing
  `devDependencies`.
- **Capabilities:** `capabilities_summary` — categories (database, auth, payments,
  queue, orm, search, container, orchestration, caching, monitoring) derived from
  dependency names.
- **Maturity:** `maturity_score` + `maturity_indicators` (JSON) from a weighted
  file-presence scoreboard (README, LICENSE, CHANGELOG, CONTRIBUTING, SECURITY,
  Dockerfile, CI, test config, linter, `tsconfig.json`, `docs/`, Makefile, …).

Engineering changes that implement this:

1. The indexer becomes the **single correct metadata writer**: it calls the pure
   `metadata.Service.Extract`, attaches `ProjectID` + `DocumentationHash` +
   `LastScanAt`, and issues one `UpsertMetadata` + `ReplaceDependencies`. The old
   clobbering partial upsert is removed.
2. A new `portfolio scan` command activates the previously-dormant scan path.
3. LOC-by-language is **intentionally excluded** as noisy and low-signal.

### Consequences

**Positive:**

- `getProject` (MCP/REST) and `portfolio projects get` (CLI) return real
  investment/maturity/capability facts with zero AI cost.
- Downstream ranking can be a **computed indicator** (read time), not a stored
  one — fully aligned with "Store Facts, Compute Indicators."

**Negative:**

- The `dependencies` unique constraint is `(project_id, name, manager)` and does
  not include `scope`; a package declared in both `dependencies` and
  `devDependencies` collapses to one row (`INSERT OR IGNORE`). This is accepted as
  a rare edge case.
- More git invocations per scan (first commit, velocity, contributors, tags,
  remote). All use cached, single-purpose `git` calls and tolerate failure.

---

## ADR-018: Per-Ecosystem Manifest Parser Registry

**Status:** Accepted

### Context

Dependency extraction (`internal/metadata/dependencies.go`) was implemented as a
single dispatcher with a hard-coded `switch` over manifest filenames, each case
calling a bespoke `parseX(path)` free function. Adding an ecosystem required
editing the central switch and growing an already-large file. This also diverged
from the catalog Pattern A used everywhere else in the metadata package
(`frameworks_data.go`, `capabilities_data.go`, `maturity_data.go`): a data type,
a `default…` registry slice, and a defensive-copy `Default…()` accessor.

### Decision

Dependency extraction is now registry-driven. Each ecosystem is a stateless
value type implementing a `ManifestParser` interface:

```go
type ManifestParser interface {
    Filename() string
    Parse(content []byte) ([]models.Dependency, error)
}
```

`DetectDependencies(root, walker, parsers)` is a thin dispatcher: it builds a
filename → parser map from the registry, and for each walked manifest reads the
bytes once and delegates to `Parse`. Parsers live in one file per ecosystem
(`dependencies_npm.go`, `dependencies_go.go`, `dependencies_python.go`,
`dependencies_rust.go`, `dependencies_ruby.go`, `dependencies_jvm.go`).

`Parse` receives the manifest **bytes**, not the path. This keeps every parser
pure (no filesystem access) and unit-testable in isolation; `os.ReadFile` happens
once, in the dispatcher.

Adding an ecosystem is now: create one file with a parser type, append one entry
to `defaultManifestParsers`. No dispatcher edits.

This mirrors the existing catalog Pattern A and keeps `DetectDependencies`
injectable (the parser set is a parameter, defaulting to
`DefaultManifestParsers()`), consistent with `DetectFrameworks` /
`DetectCapabilities` / `DetectMaturity`.

### Consequences

**Positive:**

- Adding an ecosystem is local and low-risk (one file + one registry line).
- Parsers are focused, individually unit-testable, and easy to enrich.
- The dispatcher shrank from ~90 lines of branching to a generic loop.
- Consistency with the rest of the metadata package's registry pattern.

**Negative:**

- `Parse(content)` carries no path context; a future parser that needs the
  directory (e.g. npm workspaces reading nested manifests) would require widening
  the interface to `Parse(path string, content []byte)`. This is a deliberate
  YAGNI trade-off, documented on the interface.
- One more concept (the interface) for contributors to learn — mitigated by
  matching an already-established pattern in the package.

No domain model, schema, store, or interface (MCP/REST/CLI) changed: the
`manager` values, dedupe key (`name|manager|scope`), prod/dev scope, sort order,
and summary are identical to before.

