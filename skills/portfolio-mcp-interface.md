---
name: portfolio-mcp-interface
description: Portfolio MCP interface for AI coding assistants
metadata:
  type: agent-guide
  audience: ai-agents
---

# Portfolio MCP Interface for AI Agents

## CRITICAL: Always Use MCP Tools

You are an AI coding assistant with access to Portfolio MCP tools. **You must ALWAYS use these MCP tools for any Portfolio operations.**

### ❌ NEVER Do This:
- Never attempt to access the Portfolio database directly
- Never use the Portfolio CLI unless explicitly requested by the user
- Never attempt to discover or access Portfolio files
- Never bypass the MCP interface
- Never try to find, read, or modify Portfolio's internal files

### ✅ ALWAYS Do This:
- Use the `list-projects` MCP tool to discover projects
- Use the `get-project` MCP tool to get project details
- Use the `search-projects` MCP tool to find projects by technology
- Use the `discover-project` MCP tool to add new projects
- Use the `get-project-analysis` MCP tool to understand project insights

## Available MCP Tools

### list-projects
Lists all projects in the user's portfolio.

**When to use:** Start here to understand the user's portfolio context.

**Example:**
```
User: "What projects do I have?"
You: [Uses list-projects] "I can see you have 5 projects in your portfolio: webapp, api-service, mobile-app, data-pipeline, and portfolio-tool itself."
```

### get-project  
Gets detailed information about a specific project.

**When to use:** When you need detailed project information for development work.

**Example:**
```
User: "Tell me about the webapp project"
You: [Uses get-project for webapp] "The webapp project uses React, Node.js, and PostgreSQL. It was last updated 2 days ago and includes features like REST API, web dashboard, and authentication."
```

### search-projects
Searches projects by technology, language, or framework.

**When to use:** When the user asks about projects using specific technologies.

**Example:**
```
User: "What projects use React?"
You: [Uses search-projects with query "React"] "You have 2 projects using React: webapp and mobile-app."
```

### discover-project
Discovers and adds a new project to the portfolio.

**When to use:** When the user mentions a new project to analyze.

**Example:**
```
User: "I just started working on a new project in ~/projects/new-tool"
You: [Uses discover-project with path ~/projects/new-tool] "I've discovered your new project. It appears to be a Go project with Makefile and go.mod files."
```

### get-project-analysis
Gets AI analysis and insights about a specific project.

**When to use:** When you need deeper understanding of project architecture, features, or relationships.

**Example:**
```
User: "What do you know about the architecture of my api-service?"
You: [Uses get-project-analysis for api-service] "The api-service follows a microservices architecture with 3 main services: authentication, data processing, and API gateway. It uses Redis for caching..."
```

### getProjectAnalyzerPrompt
Gets the analysis prompt with investigation workflow and storeAnalysis schema.

**When to use:** Before analyzing a project — provides the complete workflow and output schema.

**Example:**
```
User: "Analyze my api-service project"
You: [Uses getProjectAnalyzerPrompt] → follows workflow → stores analysis and features
```

## Why MCP Tools Matter

The MCP interface provides:
- **Semantic understanding** of project structure and relationships
- **Proper access control** for project data
- **Consistent operations** across different project types
- **Optimized queries** for portfolio information
- **Standardized interface** that works across all AI assistants

## Example Workflows

### Understanding User's Development Portfolio
1. User asks: "What technologies am I working with?"
2. You: Use `list-projects` to get overview, then analyze technologies across all projects
3. You: "Based on your portfolio, you're working with React, Node.js, Go, PostgreSQL, Redis, and Python across your 5 projects."

### Starting Work on Existing Project
1. User asks: "I need to add a new feature to the webapp"
2. You: Use `get-project` for webapp to understand current structure
3. You: Use `get-project-analysis` for webapp to understand architecture
4. You: "I see the webapp uses React with a component-based architecture. Where would you like to add the new feature?"

### Adding New Project to Portfolio
1. User mentions: "I'm working on a new Python project in ~/projects/ai-tool"
2. You: Use `discover-project` with path ~/projects/ai-tool
3. You: Use `get-project-analysis` to understand the new project
4. You: "I've discovered your new ai-tool project. It appears to be a Python project with AI/ML dependencies. Would you like me to analyze its structure?"

### Technology-Specific Questions
1. User asks: "Which projects should I update to the latest React version?"
2. You: Use `search-projects` with query "React"
3. You: "You have 2 React projects: webapp and mobile-app. Both would benefit from updating to the latest version."

## Important Reminders

**The MCP tools are the ONLY interface you should use for Portfolio operations.**

- Portfolio's database is protected and cannot be accessed directly
- The Portfolio CLI exists but should only be used when the user explicitly requests it
- All file discovery and project analysis should happen through MCP tools
- Never try to bypass the MCP interface for any reason

## Error Handling

If you encounter issues with MCP tools:
1. Report the issue clearly to the user
2. Suggest they check their Portfolio configuration
3. Never try to work around the MCP interface
4. The MCP tools provide all the functionality needed

## Summary

**ALWAYS use MCP tools. They are the correct, supported interface for Portfolio operations.**

When in doubt, start with `list-projects` to understand the user's portfolio, then use the specific tools you need for the task at hand.