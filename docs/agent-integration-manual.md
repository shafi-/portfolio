<!-- DO NOT EDIT — regenerate via: portfolio manual --write -->
# Portfolio — Agent Integration Manual

This is the canonical reference for connecting any MCP-capable AI coding agent
to Portfolio. It is rendered from Portfolio's shared skill template, so it
always matches the tools Portfolio actually exposes — it never drifts.

Supported agents with automated setup:

    portfolio install claude      # Claude Code
    portfolio install opencode    # OpenCode

For any other agent (Cursor, Zed, Continue, Roo, ...), follow the steps below.

This file is generated. Regenerate it with:  portfolio manual --write

## 1. Connect to the Portfolio MCP server

Portfolio runs an MCP server over stdio. Configure your agent to launch it as a
local stdio MCP server using the command:

    portfolio mcp

If your agent config requires an absolute path, use the path to your installed
portfolio binary followed by "mcp". The server needs no other arguments.

## 2. Load the Portfolio skill

Give your agent the skill text below so it knows Portfolio's tools and the
three-tier knowledge protocol. When the skill stores analysis or features it
records an analyzer identity; replace <your-agent> below with a stable name for
your agent (for example cursor or zed) so Portfolio can attribute the work.

---

## Tools

### health
Check if Portfolio is running.

### discoverProjects
Scan configured directories for new projects.

### listProjects
List all known projects.

### getProject
Get details + metadata for a project.
Call `getProject(id: "<project-id>")`

### searchProjects
Search across project names, analyses, features, technologies.
Call `searchProjects(query: "react")`

### searchDocumentation
Search within project documentation.
Call `searchDocumentation(query: "auth")`

### listProjectFiles
List files in a project directory with recursive depth control.
Params: project_id, path (relative, default root), max_depth (default 5)

### getFileContent
Read a file's content. Blocks sensitive files (.env, keys, .git config).
Params: project_id, path (relative)

### searchFiles
Find files whose relative path matches a regex and return them (with content by default). Use to locate a feature's files by name, e.g. `searchFiles(projectId, pattern: "auth")` or `pattern: "payment.*handler"`.
Params: project_id, pattern (regex), max_results (default 20), include_content (default true)

### getProjectStructure
Aggregated project overview: file tree + metadata + key file contents.
Params: project_id, include_content (bool)

### getDependencies
List parsed dependencies with versions and package managers.
Params: project_id

### getAnalysis
Get semantic analysis for a project — includes maturity, strengths, weaknesses.
Call `getAnalysis(projectId: "<id>")`

### storeAnalysis
Store Tier 2 analysis — project-level understanding.
Params: projectId, analyzer, summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes, raw_json

### listProjectsNeedingAnalysis
Find projects missing or with outdated analysis.

### storeFeature
Store a feature (Tier 2 or 3). Upserts by (project, analyzer, name): re-calling
with the same name updates the existing feature instead of creating a duplicate.
For Tier 3, include implementation_status, feature_architecture, pattern.
Params: project_id, analyzer, name, description, confidence, implementation_status, feature_architecture, pattern

### listFeatures
List all features for a project.

### searchFeatures
Search features by project, status, pattern, or text.
Params: project_id, query, implementation_status, pattern

### storeTechnology / tagProjectWithTechnology / listTechnologies / listProjectTechnologies / searchByTechnology
Manage and search technologies across the portfolio.

### listRelationships / storeRelationship
Manage project relationships (similar, evolution, shared feature/tech, reuses component).

### getConfiguration / updateConfiguration
View/update Portfolio configuration.

## Analyzer Identity

Always set `analyzer: "<your-agent>"` for storeAnalysis and storeFeature.

## Three-Tier Knowledge Protocol

Portfolio uses three tiers of knowledge. Each tier builds on the previous.

### Tier 1: Project Context (Engine-owned, deterministic)

Automatically populated by the engine:
- Discovery + metadata (language, framework, git info)
- Documentation indexing
- Technology detection

**What you do**: Check before analyzing. Use `getProject`, `searchDocumentation`.

### Tier 2: Analysis + Features (Agent — project-level)

One investigation pass produces both analysis and feature list.

**Workflow:**
1. `getProject` — check existing state
2. Investigate source code (read key files)
3. `storeAnalysis` — purpose, architecture, strengths, weaknesses, reusable_components
4. `storeFeature` for each feature found — name, description, confidence

**Guidelines:**
- Extract features as named capabilities ("User Authentication", "Search", "Dashboard")
- Confidence reflects how certain you are the feature exists (0.5=likely, 0.8=confirmed, 1.0=verified in code)
- Strengths/weaknesses are project-level, not feature-level
- Reusable_components describes libraries, utilities, patterns worth reusing

### Tier 3: Feature Deep Dive (Agent — per-feature)

For features that warrant deeper understanding, update with:
- `implementation_status`: planned|partial|complete|mature|deprecated
- `feature_architecture`: how the feature is implemented (components, data flow)
- `pattern`: architectural patterns used (MVVM, Repository, Event-driven)

**When to deep dive:**
- The user asks "how does X work?"
- You're building something related and need to understand the pattern
- The feature is core to the project's value

**How:**
- No second investigation pass — reuse the Tier 2 investigation
- Call `storeFeature` again with the SAME project_id, analyzer, and name, plus
  the Tier 3 fields. Portfolio upserts by (project, analyzer, name): the
  existing feature is found and updated — no duplicate is created.
- Fields you omit are preserved. Empty values do NOT overwrite stored facts, so
  a Tier 3 call (which omits description) keeps the Tier 2 description intact.
  Pass `confidence` only if you want to revise it.

## Important Notes

- **Workflow**: `health` → `discoverProjects` → `listProjects` → investigate → store
- **Never Edit Repositories**: Portfolio is read-only
- **Prefer Existing Knowledge**: Check existing analysis before re-analyzing
- **Single Investigation Pass**: Read source code once; extract all tiers from that pass
- **Analyzer Field**: Always `"<your-agent>"` — this identifies your analysis across sessions

## Example Workflows

1. "What projects do I have?"
   → `listProjects()`

2. "Analyze this project"
   → `getProject(id)` → investigate code → `storeAnalysis(...)` → `storeFeature(...)` for each feature

3. "Find projects using React"
   → `searchProjects(query: "react")`

4. "What features are partially implemented?"
   → `searchFeatures(implementation_status: "partial")`

5. "How does authentication work in my projects?"
   → `searchFeatures(query: "auth")` → `searchFeatures(pattern: "JWT")`

6. "Deep dive on User Auth feature"
   → `storeFeature(project_id, analyzer: "<your-agent>", name: "User Authentication", implementation_status: "complete", feature_architecture: "JWT middleware + session store", pattern: "Middleware-based auth")`

7. "What changed in my portfolio?"
   → `discoverProjects()` → `listProjectsNeedingAnalysis()` → analyze outdated projects

