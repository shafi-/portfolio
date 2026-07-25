package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// codeContentTools returns tools for accessing project code content
func (s *Server) codeContentTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("listProjectFiles",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("path", mcp.Description("Relative path to list (default: root)")),
				mcp.WithNumber("max_depth", mcp.Description("Maximum directory depth (default: 5)")),
			),
			Handler: s.handleListProjectFiles,
		},
		{
			Tool: mcp.NewTool("getFileContent",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("path", mcp.Required(), mcp.Description("Relative file path")),
			),
			Handler: s.handleGetFileContent,
		},
		{
			Tool: mcp.NewTool("getProjectStructure",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithBoolean("include_content", mcp.Description("Include file content for key files (default: false)")),
			),
			Handler: s.handleGetProjectStructure,
		},
		{
			Tool: mcp.NewTool("getDependencies",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleGetDependencies,
		},
	}
}

// handleListProjectFiles returns file tree for a project
func (s *Server) handleListProjectFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	path := ""
	if p, ok := args["path"].(string); ok {
		path = p
	}

	maxDepth := 5
	if d, ok := args["max_depth"].(float64); ok {
		maxDepth = int(d)
	}

	// Get project
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Build full path
	fullPath := filepath.Join(project.RootPath, path)

	// Check path exists
	if _, err := s.osFS.Stat(fullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path not found: %v", err)), nil
	}

	// List files
	files, err := s.listFilesRecursively(fullPath, "", maxDepth, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list files: %v", err)), nil
	}

	result := map[string]interface{}{
		"project_id": projectID,
		"path":       path,
		"files":      files,
		"count":      len(files),
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetFileContent returns file content
func (s *Server) handleGetFileContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	path, _ := args["path"].(string)
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	// Get project
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Build full path
	fullPath := filepath.Join(project.RootPath, path)

	// Read file using os.ReadFile since fs interface doesn't have ReadFile
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	// Get file info
	info, _ := s.osFS.Stat(fullPath)

	result := map[string]interface{}{
		"project_id": projectID,
		"path":       path,
		"content":    string(content),
		"size":       info.Size(),
		"modified":   info.ModTime(),
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetProjectStructure returns aggregated project structure
func (s *Server) handleGetProjectStructure(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	includeContent := false
	if ic, ok := args["include_content"].(bool); ok {
		includeContent = ic
	}

	// Get project
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Get metadata
	metadata, err := s.metadata.GetMetadata(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("metadata not found: %v", err)), nil
	}

	// List files
	files, err := s.listFilesRecursively(project.RootPath, "", 10, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list files: %v", err)), nil
	}

	// Build structure
	structure := map[string]interface{}{
		"project_id": projectID,
		"name":       project.Name,
		"type":       project.RepositoryType,
		"root_path":  project.RootPath,
		"files":      files,
		"file_count": len(files),
	}

	// Add key file contents if requested
	if includeContent {
		keyFiles := s.getKeyFiles(project.RootPath, files)
		contents := make(map[string]string)

		for _, keyFile := range keyFiles {
			fullPath := filepath.Join(project.RootPath, keyFile)
			if content, err := os.ReadFile(fullPath); err == nil {
				// Truncate large files
				if len(content) > 10000 {
					contents[keyFile] = string(content[:10000]) + "\n... (truncated)"
				} else {
					contents[keyFile] = string(content)
				}
			}
		}

		structure["key_file_contents"] = contents
	}

	// Add metadata if available
	if metadata != nil {
		structure["metadata"] = map[string]interface{}{
			"languages":        metadata.LanguageSummary,
			"frameworks":       metadata.FrameworkSummary,
			"package_managers": metadata.DependencySummary,
		}
	}

	return mcp.NewToolResultJSON(structure)
}

// handleGetDependencies returns parsed dependencies
func (s *Server) handleGetDependencies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	// Get dependencies from store
	dependencies, err := s.dependencies.ListDependencies(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get dependencies: %v", err)), nil
	}

	// Get metadata for language info
	metadata, err := s.metadata.GetMetadata(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("metadata not found: %v", err)), nil
	}

	result := map[string]interface{}{
		"project_id":   projectID,
		"dependencies": dependencies,
		"count":        len(dependencies),
		"languages":    []string{},
	}

	if metadata != nil {
		result["languages"] = metadata.LanguageSummary
		result["repository_type"] = "" // Add from project if needed
	}

	return mcp.NewToolResultJSON(result)
}

// Helper: list files recursively with depth limit
func (s *Server) listFilesRecursively(basePath, relPath string, maxDepth, currentDepth int) ([]map[string]interface{}, error) {
	if currentDepth >= maxDepth {
		return []map[string]interface{}{}, nil
	}

	fullPath := filepath.Join(basePath, relPath)

	entries, err := s.osFS.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []map[string]interface{}

	for _, entry := range entries {
		name := entry.Name()

		// Skip common directories to ignore
		if s.shouldSkip(name) {
			continue
		}

		entryRelPath := filepath.Join(relPath, name)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := map[string]interface{}{
			"path":   entryRelPath,
			"name":   name,
			"size":   info.Size(),
			"mode":   info.Mode().String(),
			"is_dir": info.IsDir(),
		}

		if info.IsDir() {
			// Recurse into subdirectory
			subFiles, err := s.listFilesRecursively(basePath, entryRelPath, maxDepth, currentDepth+1)
			if err == nil {
				fileInfo["children"] = subFiles
			}
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// Helper: check if path should be skipped
func (s *Server) shouldSkip(name string) bool {
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"target":       true,
		"build":        true,
		"dist":         true,
		"bin":          true,
		"obj":          true,
		".vscode":      true,
		".idea":        true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		"env":          true,
		".env":         true,
		".DS_Store":    true,
	}

	return skipDirs[name]
}

// Helper: get key files for analysis
func (s *Server) getKeyFiles(rootPath string, files []map[string]interface{}) []string {
	keyFiles := []string{}
	keyPatterns := []string{
		"README.md",
		"package.json",
		"go.mod",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
		"requirements.txt",
		"Gemfile",
		"composer.json",
		"main.go",
		"main.rs",
		"app.py",
		"index.js",
		"Dockerfile",
		"docker-compose.yml",
		".gitignore",
		"LICENSE",
	}

	// Check for key files in file list
	for _, file := range files {
		path, ok := file["path"].(string)
		if !ok {
			continue
		}

		for _, pattern := range keyPatterns {
			if strings.HasSuffix(path, pattern) || path == pattern {
				keyFiles = append(keyFiles, path)
				break
			}
		}
	}

	return keyFiles
}
