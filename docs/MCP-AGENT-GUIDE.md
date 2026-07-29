# Portfolio MCP Tools - Agent Guide

**For AI Coding Assistants**

Portfolio provides your project data through these MCP tools. Use these tools to understand the user's development portfolio and project context.

## Available Tools

### `list-projects`
List all projects in the user's portfolio.

**Usage:**
```json
{
  "tool": "list-projects",
  "parameters": {
    "filter": "optional filter criteria (optional)"
  }
}
```

**Returns:**
```json
{
  "projects": [
    {
      "id": "project-1",
      "name": "My Web App",
      "path": "/Users/developer/projects/webapp",
      "type": "git"
    },
    {
      "id": "project-2", 
      "name": "API Service",
      "path": "/Users/developer/projects/api",
      "type": "git"
    }
  ]
}
```

**When to use:**
- Start here to understand the user's portfolio
- Get overview of all projects
- Find specific project by name or type

---

### `get-project`
Get detailed information about a specific project.

**Usage:**
```json
{
  "tool": "get-project",
  "parameters": {
    "project_id": "project-1",
    "include_analysis": true
  }
}
```

**Returns:**
```json
{
  "id": "project-1",
  "name": "My Web App",
  "path": "/Users/developer/projects/webapp",
  "technologies": ["React", "Node.js", "PostgreSQL"],
  "features": ["REST API", "Web Dashboard", "Authentication"],
  "metadata": {
    "language": "JavaScript",
    "framework": "React",
    "last_updated": "2026-07-28"
  }
}
```

**When to use:**
- Need detailed information about a specific project
- Understanding project tech stack
- Learning about project features
- Getting project context for development work

---

### `search-projects`
Search projects by technology, language, or framework.

**Usage:**
```json
{
  "tool": "search-projects",
  "parameters": {
    "query": "python and django"
  }
}
```

**Returns:**
```json
{
  "query": "python and django",
  "results": [
    {
      "id": "project-3",
      "name": "Django API",
      "relevance": 0.95
    }
  ]
}
```

**When to use:**
- Find projects with specific technologies
- Search for capabilities across portfolio
- Find projects matching technical criteria

---

### `discover-project`
Discover and add a new project to the portfolio.

**Usage:**
```json
{
  "tool": "discover-project",
  "parameters": {
    "path": "/Users/developer/projects/new-project"
  }
}
```

**Returns:**
```json
{
  "id": "new-project-1",
  "name": "New Project",
  "path": "/Users/developer/projects/new-project",
  "technologies": ["Go", "React"],
  "status": "discovered"
}
```

**When to use:**
- User wants to add a project to their portfolio
- Analyze a new project structure
- Extract metadata from project files

---

## Best Practices

### 1. Start with `list-projects`
Always begin by listing projects to understand the user's portfolio context.

### 2. Use `get-project` for context
When working on a specific project, get detailed information first.

### 3. Search before creating
Use `search-projects` to check if similar capabilities already exist.

### 4. Discover when needed
Use `discover-project` when user mentions a new project.

## Example Workflows

### Understanding User's Development Portfolio
```bash
# Step 1: List all projects
list-projects

# Step 2: Get details about relevant projects  
get-project(project_id="webapp")

# Step 3: Search for specific capabilities
search-projects(query="microservices")
```

### Starting Work on Existing Project
```bash
# Step 1: Find the project
list-projects(filter="webapp")

# Step 2: Get project details
get-project(project_id="webapp", include_analysis=true)

# Step 3: Understand tech stack
# (from get-project results)
```

### Adding New Project to Portfolio
```bash
# Step 1: Discover the project
discover-project(path="/path/to/new-project")

# Step 2: Verify it was added
list-projects()
```

## Important Notes

- **Portfolio provides project data through these MCP tools only**
- **No direct database or file access needed or supported**
- **All project information is available through these semantic tools**
- **Tools are designed to be intuitive and comprehensive**

## Tool Philosophy

Portfolio follows the principle that AI assistants should work at the semantic level:
- **You**: Understand and reason about projects
- **Portfolio**: Provide project data and handle storage
- **MCP Interface**: Clean separation of concerns

This approach keeps your focus on helping the user with their development work, while Portfolio handles the data management.

---

**Need help?** Start with `list-projects` to explore what's available in the user's portfolio!
