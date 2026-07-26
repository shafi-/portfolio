
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

**Status:** Accepted (amended by ADR-021)

### Context

Portfolio integrations need to register MCP servers with various AI coding agents.
Different agents have different approaches to MCP server registration:

- **Claude Code**: Provides official CLI commands (`claude mcp add/remove/get`)
- **OpenCode**: No local-stdio CLI, but its config file (`~/.config/opencode/opencode.json`) is the officially schema-documented method for local servers — see ADR-021
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
2. **No fragile/undocumented config editing**: Production code must not edit agent config files by guesswork. Editing a tool's *officially schema-documented* config file — one the tool publishes a `$schema` for and documents as its config surface — is permitted, because that file IS the tool's official method. Blind/undocumented edits remain forbidden. See ADR-021.
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
- OpenCode: local-stdio servers are registered by writing its schema-documented `opencode.json` (the official method per ADR-021); `opencode mcp add` covers remote servers only
- Cline requires manual `~/.cline/mcp.json` editing (no official method)
- Unsafe scripts for tools without any official method live in `scripts/` with clear README documentation

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
- The hard-coded switch over manifest filenames was replaced by a generic
  map-dispatch loop; adding an ecosystem no longer touches the dispatcher.
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

## ADR-019: Store Declared Dependency Version as a Literal Fact

**Status:** Accepted

### Context

`dependencies` rows stored only `name`, `manager`, and `scope`; every parser
discarded the declared version (npm even had it in hand and dropped it). Users
want to identify projects on a given version and to judge whether a dependency is
"outdated." The design question was what to store: the raw declared spec (e.g.
`^4.0.0`, `~> 7.0`), a normalized bare version, or a split of the two.

### Decision

Store the declared version as **two** columns on `dependencies`:

- `version` — the value with its operator stripped (`4.0.0`, `v1.2.3`, `2.28.0`).
- `version_type` — the literal constraint kind: the leading operator verbatim
  (`^`, `~`, `~>`, `>=`, `<=`, `==`, `!=`, `=`, `>`, `<`), or `exact` for a bare
  pin, `range` for a compound spec (`>=1.0,<2.0`), `any` for `*`/`latest`/x-ranges,
  or `""` when unknown.

The decomposition is **literal, not semantic.** Cargo's implicit caret
(`serde = "1.0"`) is recorded as `exact` because the manifest pins it literally;
interpreting it as a caret is the AI agent's job. A single shared helper,
`parseVersionSpec(raw)` in `internal/metadata/version.go`, centralizes the
operator/kind logic; each ecosystem parser extracts its raw spec string and feeds
it in. The dedupe identity is unchanged (`name|manager|scope`, prod wins the
`INSERT OR IGNORE`), so the version rides along on the surviving production row.

### Consequences

Both the value and the constraint kind are independently queryable ("who pins
exactly 4.0.0", "who uses caret ranges"). "Is this outdated?" remains a
**semantic indicator** — it needs a registry/latest-version lookup plus semver
comparison — and so stays agent-side, consistent with "Engine Knows, Agent
Thinks" and "Store Facts, Compute Indicators." The `dependencies` table gains two
NOT NULL DEFAULT '' columns. Compound ranges and x-ranges have no single value,
so `version` holds the whole declared string under kind `range`/`any`. Reading
lockfiles (`package-lock.json`, `Cargo.lock`, `go.sum`) for *resolved* versions is
deliberately out of scope; the declared spec is the fact stored here.

## ADR-020: Three-Tier Knowledge Model and Feature Upsert-by-Name

**Status:** Accepted

### Context

Analysis knowledge had two layers: Tier 1 (engine-owned, deterministic) and
Tier 2 (agent-owned analysis + features). In practice an agent needs a third,
optional layer — a per-feature deep dive ("how is auth implemented?", "what
pattern does search use?") — without re-running the whole investigation. The
`features` table stored only `name`, `description`, `confidence`, so there was
nowhere to put implementation status, architecture, or pattern. Worse, the
`storeFeature` MCP tool always created a new row (fresh UUID), so a deep-dive
"update" actually produced a duplicate feature with empty Tier-2 fields — the
documented workflow did not work.

### Decision

1. **Add three Tier-3 columns** to `features` (migration `tier3_feature_extras`,
   v3): `implementation_status` (`planned|partial|complete|mature|deprecated`,
   default `planned`), `feature_architecture` (how it is implemented), and
   `pattern` (architectural patterns).

2. **Upsert `storeFeature` by (project, analyzer, name).** The handler resolves
   the analysis for the analyzer, then looks up a feature by
   `(analysis_id, name)` via `FeatureStore.GetByAnalysisAndName`. If present, it
   **merges** the caller's supplied fields onto the stored row and calls
   `UpdateFeature`; otherwise it creates. Merge semantics: empty strings mean
   "leave as-is" (so a Tier-3 call omitting `description` never blanks the
   stored Tier-2 description), and `confidence` is applied only when explicitly
   passed. This makes a single investigation pass reusable across tiers.

3. **No `UNIQUE(analysis_id, name)` constraint yet.** The lookup is read-then-
   write. The store is single-writer (the indexer for deterministic facts; the
   agent for analyses/features, low concurrency), so the TOCTOU window is
   acceptable. A unique constraint + backfill of any pre-existing duplicates is
   recorded as future hardening; `GetByAnalysisAndName` collapses duplicates to
   the first row (`LIMIT 1`).

### Consequences

The deep-dive workflow now enriches instead of duplicating, and the three-tier
model (engine context → project analysis+features → per-feature deep dive) is the
documented agent contract in the Claude Code skill. `searchFeatures` queries the
new columns (status, pattern, free text across name/description/architecture).
`featureColumns` is `f.`-qualified because the search/list-by-project readers
JOIN `analyses`, which also has `id`/`name` — an unqualified projection was
ambiguous. A second call with the same name is an update (returns `updated: true`
vs `created: true`), so callers can distinguish the two. Stale Tier-3 fields are
preserved unless explicitly overwritten; agents that want to clear a field must
pass the new value rather than rely on omission.

---

## ADR-021: Schema-Documented Config Files Are Official Methods

**Status:** Accepted
**Amends:** ADR-016

### Context

ADR-016 forbade direct config-file editing in production code to avoid fragility
and "guesswork" edits. The OpenCode integration needs to register a **local**
(stdio) MCP server, but OpenCode ships no local-stdio MCP CLI — `opencode mcp
add` supports remote servers only. OpenCode's only documented way to register a
local server is to write its config file, `~/.config/opencode/opencode.json`,
which it declares stable by publishing `$schema: https://opencode.ai/config.json`
and documenting the `mcp.<name>` object shape.

This is categorically different from the fragile edits ADR-016 warned against:
the file is the tool's intended, versioned config surface, not an internal
representation we reverse-engineered.

### Decision

Writing a tool's **officially schema-documented** config file is an *official
method*, permitted in production code. It is held to the same standard as an
official CLI. Requirements for any such integration:

1. The target file must be publicly schema-documented (e.g. a published `$schema`)
   by the tool itself.
2. The write must be a **read-merge-write** that preserves all unrelated keys
   (other servers, user settings) and an existing `$schema`.
3. The write must be **atomic** (temp file + rename) so the config is never left
   half-written on failure.
4. The integration must be **idempotent** — re-running install/upgrade produces
   the same result without duplicating entries.
5. Remove must delete only our entry and leave siblings intact.

Blind/undocumented config edits remain forbidden by ADR-016.

### Consequences

**Positive:**

- OpenCode gets a first-class, automated integration (`portfolio install
  opencode`) on equal footing with Claude Code.
- The policy now distinguishes *official config surfaces* from *fragile hacks*,
  so ADR-016's intent (stability, trust) is preserved without blocking
  legitimate official methods.

**Negative:**

- We own correctness of the merge/atomic-write logic; it is covered by tests
  (`internal/integration/opencode/mcp_config_test.go`).
- If OpenCode renames or restructures the documented config, the integration
  needs an update — the same maintenance reality as any official CLI whose
  commands change.

**Implementation Notes:**

- `internal/integration/opencode/mcp_config.go` performs the read-merge-write via
  `map[string]json.RawMessage` (preserves unrelated top-level keys) and a
  per-entry `mcp` map (preserves sibling servers).
- The entry written is `mcp.portfolio = {type:"local", command:[<binary>,"mcp"],
  enabled:true}`.
- The superseded `scripts/unsafe-opencode-integration.sh` is removed; Cline
  (no official method) keeps its unsafe script.

