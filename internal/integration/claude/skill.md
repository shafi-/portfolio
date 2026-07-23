# Portfolio — Claude Code Skill

Portfolio helps you understand a developer's entire software portfolio.
Use these MCP tools through Claude Code:

## Tools

### health
Check if Portfolio is running.
Usage: `Call health()`

### discoverProjects
Scan configured directories for new projects.
Usage: `Call discoverProjects()`

### listProjects
List all known projects.
Usage: `Call listProjects()`

### getProject
Get details for a specific project.
Usage: `Call getProject(id: "<project-id>")`

### searchProjects
Search projects by name, language, framework.
Usage: `Call searchProjects(query: "react")`

### searchDocumentation
Search within project documentation.
Usage: `Call searchDocumentation(query: "architecture")`

### getAnalysis
Get semantic analysis for a project.
Usage: `Call getAnalysis(projectId: "<project-id>")`

### storeAnalysis
Store semantic analysis for a project.
Usage: `Call storeAnalysis(projectId: "<project-id>", summary: "...", purpose: "...")`

### listProjectsNeedingAnalysis
Find projects missing or with outdated analysis.
Usage: `Call listProjectsNeedingAnalysis()`

### getConfiguration
View Portfolio configuration.
Usage: `Call getConfiguration()`

### updateConfiguration
Update Portfolio configuration.
Usage: `Call updateConfiguration(key: "...", value: "...")`

### listRelationships
List relationships for a project.
Usage: `Call listRelationships(projectId: "<project-id>")`

## Important Notes

- **Analyzer Identity**: Always set `analyzer: "claude-code"` when calling `storeAnalysis()`
- **Workflow**: Start with `health()` → `discoverProjects()` → search metadata → analyze → store
- **Never Edit Repositories**: Portfolio is read-only — never suggest code changes to repositories
- **Prefer Existing Knowledge**: Check for existing analysis before re-analyzing

## Example Workflows

1. "What projects do I have?"
   → `Call listProjects()`

2. "Show me analysis for project X"
   → `Call getAnalysis(projectId: "<id>")`

3. "Find projects using React"
   → `Call searchProjects(query: "react")`

4. "What changed recently?"
   → `Call discoverProjects()` → `Call listProjectsNeedingAnalysis()`

5. "How are my projects related?"
   → `Call listProjects()` → for each: `Call listRelationships(projectId: "<id>")`

6. "Analyze a new project"
   → `Call getProject(id: "<id>")` → investigate → `Call storeAnalysis(projectId: "<id>", analyzer: "claude-code", summary: "...", purpose: "...", features: [...], technologies: [...])`