package mcp

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/mark3labs/mcp-go/mcp"
)

// maxFileContentBytes caps the size of a single file returned by the
// code-content tools, bounding memory use and MCP message size.
const maxFileContentBytes = 1 << 20 // 1 MiB

// maxListDepth caps how deep listProjectFiles / getProjectStructure recurse, so
// an agent-supplied depth cannot force an unbounded traversal of the tree.
const maxListDepth = 20

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

// resolveProjectPath validates an agent-supplied relative path against the
// project root, defeating ".." traversal and symlink escape, and returns the
// resolved absolute path inside root. It errors if the path escapes root or does
// not exist. An absolute rel is collapsed under root by filepath.Join, so it is
// contained too — only ".." segments and out-of-root symlinks are rejected.
func resolveProjectPath(root, rel string) (string, error) {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("invalid project root: %w", err)
	}
	// Resolve the root itself (it may contain symlinks) for an accurate prefix
	// comparison; fall back to the cleaned root if it cannot be resolved.
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		resolvedRoot = absRoot
	}

	// filepath.Join cleans "..", collapsing traversal; EvalSymlinks then resolves
	// the real target so a symlink planted inside the repo cannot escape root.
	candidate, err := filepath.EvalSymlinks(filepath.Join(absRoot, rel))
	if err != nil {
		return "", fmt.Errorf("path not found: %w", err)
	}
	if candidate != resolvedRoot && !isWithinPath(candidate, resolvedRoot) {
		return "", fmt.Errorf("path escapes project root: %q", rel)
	}
	return candidate, nil
}

// isWithinPath reports whether p is a strict descendant of root (both absolute,
// clean, symlink-resolved). Callers handle the p == root case separately.
func isWithinPath(p, root string) bool {
	prefix := root
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(p, prefix)
}

// isSensitiveFile reports whether relPath (relative to the project root) names a
// file whose content the code-content tools must never return — credentials,
// private keys, and VCS/agent config that commonly embeds secrets. Existence may
// still be listed; only content is withheld.
func isSensitiveFile(relPath string) bool {
	rel := filepath.ToSlash(filepath.Clean(relPath))
	base := path.Base(rel)

	// Never expose anything under .git (config may carry http.extraHeader tokens).
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".git" {
			return true
		}
	}
	if strings.HasPrefix(base, ".env") {
		return true
	}
	switch base {
	case ".npmrc", ".pypirc", ".netrc", ".gitconfig", ".dockercfg",
		"credentials", "credentials.json":
		return true
	}
	// SSH private keys (public ".pub" siblings are over-blocked for safety).
	for _, k := range []string{"id_rsa", "id_ecdsa", "id_ed25519", "id_dsa"} {
		if strings.HasPrefix(base, k) {
			return true
		}
	}
	// TLS / keystore private material.
	for _, ext := range []string{".pem", ".key", ".pfx", ".keystore", ".p12"} {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}
	if strings.Contains(rel, ".aws/credentials") {
		return true
	}
	return false
}

// truncateContent returns content as a string, rune-safely truncated to max
// bytes with a sentinel appended. Slicing on a raw byte boundary would split a
// multi-byte UTF-8 rune and emit invalid JSON.
func truncateContent(content []byte, max int) string {
	if len(content) <= max {
		return string(content)
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(content[cut]) {
		cut--
	}
	return string(content[:cut]) + "\n... (truncated)"
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
	if d, ok := args["max_depth"].(float64); ok && d >= 1 {
		maxDepth = int(d)
	}
	if maxDepth > maxListDepth {
		maxDepth = maxListDepth
	}

	// Get project
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Resolve + contain the path (defeats traversal and symlink escape).
	fullPath, err := resolveProjectPath(project.RootPath, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	info, err := s.osFS.Stat(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("path not found: %v", err)), nil
	}
	if !info.IsDir() {
		return mcp.NewToolResultError("path is not a directory"), nil
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

	relPath, _ := args["path"].(string)
	if relPath == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	// Get project
	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Resolve + contain the path (defeats traversal and symlink escape).
	fullPath, err := resolveProjectPath(project.RootPath, relPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Never expose credentials/keys/VCS config to the agent.
	if isSensitiveFile(relPath) {
		return mcp.NewToolResultError("access to this file is not permitted"), nil
	}

	info, err := s.osFS.Stat(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stat file: %v", err)), nil
	}
	if info.IsDir() {
		return mcp.NewToolResultError("path is a directory"), nil
	}
	if info.Size() > maxFileContentBytes {
		return mcp.NewToolResultError(fmt.Sprintf(
			"file is too large (%d bytes; limit %d)", info.Size(), maxFileContentBytes)), nil
	}

	// Read file using os.ReadFile since the fs interface doesn't have ReadFile.
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	result := map[string]interface{}{
		"project_id": projectID,
		"path":       relPath,
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
			// Each key file is resolved + contained and screened for secrets
			// before reading, same as getFileContent.
			fullPath, err := resolveProjectPath(project.RootPath, keyFile)
			if err != nil || isSensitiveFile(keyFile) {
				continue
			}
			info, err := s.osFS.Stat(fullPath)
			if err != nil || info.IsDir() || info.Size() > maxFileContentBytes {
				continue
			}
			if content, err := os.ReadFile(fullPath); err == nil {
				contents[keyFile] = truncateContent(content, 10000)
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
