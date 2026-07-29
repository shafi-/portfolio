# MCP-Only Security Architecture

**Date**: 2026-07-28  
**Approach**: Security by Design through Information Hiding  
**Status**: Recommended architectural approach

---

## Core Principle

**AI agents should ONLY know about Portfolio MCP tools for accessing project information.**

By hiding database location, file structure, and direct access methods from agents, we create natural security - agents will use MCP tools because that's the only interface they know about.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│              AI Agent (Limited Knowledge)                    │
│                                                              │
│  Agent KNOWS:                                               │
│  ✅ Portfolio MCP tools (list-projects, get-project, etc.) │
│  ✅ Portfolio CLI commands                                 │
│  ✅ Project-level operations                                │
│                                                              │
│  Agent DOESN'T KNOW:                                       │
│  ❌ Database location (~/.portfolio/portfolio.db)           │
│  ❌ Configuration file location                             │
│  ❌ Direct access methods (sqlite3, file reading)          │
│  ❌ Database schema or structure                            │
│  ❌ Internal file organization                              │
└────────────────────┬────────────────────────────────────────┘
                     │ Natural MCP Usage
                     │ (Only option agent knows)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Portfolio MCP Server                            │
│  - Handles all project data requests                       │
│  - Enforces access boundaries                              │
│  - Manages database internally                             │
│  - Provides semantic API                                    │
└────────────────────┬────────────────────────────────────────┘
                     │ Internal Database Access
                     │ (Hidden from agents)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│         Portfolio Database (Password Protected)              │
│  - Internal implementation detail                          │
│  - Password protected                                      │
│  - Never exposed to agents                                 │
└─────────────────────────────────────────────────────────────┘
```

## Security by Design

### Information Hiding Strategy

```go
// internal/mcp/tools.go
// MCP tools expose ONLY semantic operations

// ✅ GOOD: Expose semantic operations
mcp.RegisterTool("list-projects", "List all projects in portfolio")
mcp.RegisterTool("get-project", "Get detailed information about a project")
mcp.RegisterTool("search-projects", "Search projects by criteria")

// ❌ BAD: Never expose internal details
// mcp.RegisterTool("get-database-path", "Get database file location")
// mcp.RegisterTool("raw-sql-query", "Execute raw SQL on database")
// mcp.RegisterTool("read-config-file", "Read Portfolio configuration")
```

### Agent Interface Design

```json
{
  "tools": [
    {
      "name": "list-projects",
      "description": "List all projects in the user's portfolio",
      "parameters": {
        "filter": "Optional filter criteria"
      }
    },
    {
      "name": "get-project", 
      "description": "Get detailed information about a specific project",
      "parameters": {
        "project_id": "Project identifier",
        "include_analysis": "Whether to include AI analysis"
      }
    }
  ],
  "no_access": [
    "database_location",
    "file_system_access", 
    "raw_data_access",
    "configuration_details"
  ]
}
```

## Implementation

### 1. MCP Tool Registration

```go
// internal/mcp/server.go
package mcp

import (
    "context"
    "github.com/mark3labs/mcp-go/mcp"
)

// RegisterProjectTools registers ONLY high-level project tools
func (s *Server) RegisterProjectTools() error {
    
    // Project discovery and listing
    s.tools["list-projects"] = mcp.Tool{
        Name:        "list-projects",
        Description: "List all projects in the user's portfolio. Returns project names, paths, and basic metadata.",
        Parameters: map[string]interface{}{
            "filter": map[string]interface{}{
                "type":        "string",
                "description": "Optional filter for project type or status",
            },
        },
    }
    
    // Project details
    s.tools["get-project"] = mcp.Tool{
        Name:        "get-project",
        Description: "Get detailed information about a specific project including metadata, technologies, and analysis results.",
        Parameters: map[string]interface{}{
            "project_id": map[string]interface{}{
                "type":        "string", 
                "description": "Project identifier or name",
            },
            "include_analysis": map[string]interface{}{
                "type":        "boolean",
                "description": "Whether to include AI-generated analysis",
            },
        },
    }
    
    // Project search
    s.tools["search-projects"] = mcp.Tool{
        Name:        "search-projects",
        Description: "Search projects by technology, language, framework, or custom criteria.",
        Parameters: map[string]interface{}{
            "query": map[string]interface{}{
                "type":        "string",
                "description": "Search query for finding matching projects",
            },
        },
    }
    
    // NOTE: Intentionally NOT exposing:
    // - database_path
    // - config_location  
    // - raw_sql_access
    // - file_system_operations
    
    return nil
}
```

### 2. Tool Implementation

```go
// internal/mcp/handlers.go
package mcp

import (
    "context"
    "database/sql"
)

// ListProjects handles list-projects tool calls
func (s *Server) ListProjects(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Agent gets ONLY semantic information
    projects, err := s.store.ListProjects()
    if err != nil {
        return nil, err
    }
    
    // Return semantic project data, NOT database details
    result := make([]map[string]interface{}, len(projects))
    for i, p := range projects {
        result[i] = map[string]interface{}{
            "id":          p.ID,
            "name":        p.Name,
            "path":        p.RootPath,     // Project path, NOT database path
            "type":        p.RepositoryType,
            "technologies": p.Technologies, // High-level info
            // NO database_path, schema, or internal details
        }
    }
    
    return result, nil
}

// GetProject handles get-project tool calls  
func (s *Server) GetProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    projectID := params["project_id"].(string)
    
    // Get project data through store, NOT database access
    project, err := s.store.GetProject(projectID)
    if err != nil {
        return nil, err
    }
    
    // Return semantic information
    return map[string]interface{}{
        "id":           project.ID,
        "name":         project.Name,
        "path":         project.RootPath,
        "metadata":     project.Metadata,
        "technologies": project.Technologies,
        "features":     project.Features,
        "analysis":     project.Analysis, // If requested
        // NO database location or internal structure
    }, nil
}
```

### 3. Agent Context Management

```go
// internal/mcp/context.go
package mcp

// AgentContext defines what agents know about Portfolio
type AgentContext struct {
    // Exposed to agents
    AvailableTools []string `json:"available_tools"`
    Capabilities   []string `json:"capabilities"`
    
    // Hidden from agents
    databasePath   string `json:"-"`              // Never exposed
    configPath     string `json:"-"`              // Never exposed  
    internalSchema bool   `json:"-"`              // Never exposed
}

// GetAgentContext returns ONLY what agents should know
func (s *Server) GetAgentContext() *AgentContext {
    return &AgentContext{
        AvailableTools: []string{
            "list-projects",
            "get-project", 
            "search-projects",
            "discover-project",
        },
        Capabilities: []string{
            "Project discovery",
            "Metadata extraction", 
            "Technology identification",
            "Analysis results",
        },
        // Internal fields intentionally omitted
    }
}
```

## Benefits of MCP-Only Architecture

### 1. Natural Security
```bash
# Agent's thought process:
"I need to list projects" 
→ "I have list-projects tool"
→ "Use list-projects tool"
→ ✅ Security maintained

# NOT:
"I need to list projects"
→ "Where's the database?"  
→ "Let me find and read it"
→ ❌ Security breach
```

### 2. Clear Separation of Concerns
```
Agent Layer:        Semantic operations (list, get, search)
MCP Layer:          API enforcement and access control
Engine Layer:       Database operations and persistence
```

### 3. Future Flexibility
```go
// Can change internal implementation without affecting agents
// Agents only know about MCP tools, not database structure

// Example: Switch from SQLite to PostgreSQL
// MCP tools stay the same, only engine implementation changes
```

### 4. Observability and Auditing
```go
// All agent operations go through MCP
// Easy to log, monitor, and audit
func (s *Server) LogAgentAccess(tool string, params map[string]interface{}) {
    s.accessLog <- AccessLog{
        Tool:      tool,
        Timestamp: time.Now(),
        AgentID:   getAgentID(),
        Parameters: params,
    }
}
```

## Documentation for AI Agents

### Agent Guide (What They See)

```markdown
# Portfolio MCP Tools

## Available Tools

### list-projects
List all projects in the user's portfolio.

**Usage:**
```json
{
  "tool": "list-projects",
  "parameters": {
    "filter": "optional filter criteria"
  }
}
```

**Returns:** Project list with names, paths, and metadata

### get-project
Get detailed information about a specific project.

**Usage:**
```json
{
  "tool": "get-project", 
  "parameters": {
    "project_id": "project-name-or-id",
    "include_analysis": true
  }
}
```

**Returns:** Complete project information

### search-projects
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

**Returns:** Matching projects with relevance scores

## Best Practices

1. Always use MCP tools for project data
2. Start with list-projects to understand portfolio
3. Use get-project for detailed information
4. Use search-projects to find specific capabilities

## Note

Portfolio provides your project data through these MCP tools. No direct database or file access needed or supported.
```

## Implementation Checklist

### Phase 1: MCP Tool Design (Current)
- [x] Design semantic MCP tools
- [x] Define tool interfaces and parameters
- [x] Document agent-facing API
- [x] Hide internal implementation details

### Phase 2: Implementation (Current)
- [ ] Implement MCP tool handlers
- [ ] Add password protection to database
- [ ] Test agent workflows use only MCP
- [ ] Verify no database location exposed

### Phase 3: Documentation (Current)
- [ ] Create agent guide
- [ ] Update API documentation
- [ ] Add examples and best practices
- [ ] Document security architecture

### Phase 4: Testing (Current)
- [ ] Test agents use MCP tools naturally
- [ ] Verify no bypass attempts
- [ ] Test access control enforcement
- [ ] Security audit of implementation

## Security Properties

### ✅ What This Prevents
- Agents discovering database location
- Agents accessing database directly
- Agents reading configuration files
- Agents bypassing access controls
- Information leakage about internal structure

### ✅ What This Enables
- Natural security through information hiding
- Clear agent interface boundaries
- Comprehensive access logging
- Flexible internal implementation
- Future-proof architecture

### ✅ Compliance and Best Practices
- Principle of least privilege
- Separation of concerns
- Defense in depth
- Security by design
- Observable and auditable

## Comparison with Other Approaches

### ❌ Approach 1: Full Database Encryption
- Complex implementation
- Performance overhead  
- Key management challenges
- Overkill for local-first

### ❌ Approach 2: Filesystem Sandboxing
- Complex sandbox setup
- Cross-platform compatibility issues
- Performance overhead
- Still exposes database location

### ✅ Approach 3: MCP-Only (RECOMMENDED)
- Simple and elegant
- Natural security by design
- Clear architectural boundaries
- Future-proof and flexible
- No performance overhead

## Conclusion

The MCP-only approach creates **natural security** by design:

1. **Agents only know about MCP tools** → They naturally use them
2. **Database details hidden** → No incentive to find them  
3. **Clear interface boundaries** → Security through architecture
4. **Simple implementation** → Easy to maintain and extend

This aligns perfectly with Portfolio's core principle: **"Engine Knows, Agent Thinks"**

- **Engine**: Database, persistence, internal operations
- **Agent**: Semantic reasoning using MCP tools
- **Boundary**: MCP interface

The result is a secure, maintainable, and future-proof architecture that prevents the security breach while maintaining simplicity.

---

**Last Updated**: 2026-07-28  
**Implementation**: Security by design through information hiding
