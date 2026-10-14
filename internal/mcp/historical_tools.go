package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// historicalTools returns tools for git history analysis and evolution tracking
func (s *Server) historicalTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("getGitHistory",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("file_path", mcp.Description("Specific file path to get history for (optional)")),
				mcp.WithNumber("limit", mcp.Description("Maximum number of commits to return (default: 50, max: 500)")),
				mcp.WithString("since", mcp.Description("Only show commits since this date (RFC3339 format)")),
			),
			Handler: s.handleGetGitHistory,
		},
		{
			Tool: mcp.NewTool("getCommitDiff",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("commit_hash", mcp.Required(), mcp.Description("Git commit hash")),
				mcp.WithString("file_path", mcp.Description("Specific file path to get diff for (optional)")),
			),
			Handler: s.handleGetCommitDiff,
		},
		{
			Tool: mcp.NewTool("getFileEvolution",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("File path to analyze evolution")),
				mcp.WithNumber("limit", mcp.Description("Maximum number of commits to analyze (default: 20)")),
			),
			Handler: s.handleGetFileEvolution,
		},
		{
			Tool: mcp.NewTool("analyzeCommitPatterns",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("time_range", mcp.Description("Time range: 30d, 90d, 180d, 365d (default: 90d)")),
				mcp.WithString("author", mcp.Description("Filter commits by author (optional)")),
			),
			Handler: s.handleAnalyzeCommitPatterns,
		},
		{
			Tool: mcp.NewTool("getProjectTimeline",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("aspect", mcp.Description("Timeline aspect: commits, releases, contributors, all (default: all)")),
			),
			Handler: s.handleGetProjectTimeline,
		},
	}
}

// handleGetGitHistory retrieves git commit history
func (s *Server) handleGetGitHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	filePath, _ := args["file_path"].(string)
	limit := 50
	if l, ok := args["limit"].(float64); ok && l >= 1 && l <= 500 {
		limit = int(l)
	}
	since, _ := args["since"].(string)

	// Build git log command
	cmdArgs := []string{"log", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso-strict"}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since", since)
	}
	if limit > 0 {
		cmdArgs = append(cmdArgs, "-n", fmt.Sprintf("%d", limit))
	}
	if filePath != "" {
		cmdArgs = append(cmdArgs, "--", filePath)
	}

	cmd := s.gitCommand(project.RootPath, cmdArgs...)

	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("git command failed: %v", err)), nil
	}

	commits := s.parseGitLogOutput(output)

	result := map[string]interface{}{
		"project_id":   projectID,
		"project_name": project.Name,
		"commits":      commits,
		"count":        len(commits),
	}

	if filePath != "" {
		result["file_path"] = filePath
	}
	if since != "" {
		result["since"] = since
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetCommitDiff retrieves commit diff
func (s *Server) handleGetCommitDiff(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	commitHash, _ := args["commit_hash"].(string)
	if projectID == "" || commitHash == "" {
		return mcp.NewToolResultError("project_id and commit_hash are required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	filePath, _ := args["file_path"].(string)

	// Build git show command
	cmdArgs := []string{"show", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso-strict"}
	if filePath != "" {
		cmdArgs = append(cmdArgs, commitHash, "--", filePath)
	} else {
		cmdArgs = append(cmdArgs, commitHash)
	}

	cmd := s.gitCommand(project.RootPath, cmdArgs...)

	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("git command failed: %v", err)), nil
	}

	// Parse the output
	lines := strings.Split(output, "\n")
	var commitInfo map[string]interface{}
	var diffContent string
	var inDiff bool

	for _, line := range lines {
		if strings.HasPrefix(line, "diff ") {
			inDiff = true
			diffContent += line + "\n"
			continue
		}

		if inDiff {
			diffContent += line + "\n"
		} else if strings.Contains(line, "|") {
			// This is commit info
			parts := strings.SplitN(line, "|", 5)
			if len(parts) >= 5 {
				commitInfo = map[string]interface{}{
					"hash":         parts[0],
					"author":       parts[1],
					"author_email": parts[2],
					"date":         parts[3],
					"message":      parts[4],
				}
			}
		}
	}

	result := map[string]interface{}{
		"project_id":   projectID,
		"project_name": project.Name,
		"commit":       commitInfo,
		"diff":         diffContent,
	}

	if filePath != "" {
		result["file_path"] = filePath
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetFileEvolution analyzes file evolution over time
func (s *Server) handleGetFileEvolution(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	filePath, _ := args["file_path"].(string)
	if projectID == "" || filePath == "" {
		return mcp.NewToolResultError("project_id and file_path are required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok && l >= 1 && l <= 100 {
		limit = int(l)
	}

	// Get git log for this file
	cmdArgs := []string{"log", "--pretty=format:%H|%an|%ad", "--date=iso-strict", "-n", fmt.Sprintf("%d", limit), "--", filePath}
	cmd := s.gitCommand(project.RootPath, cmdArgs...)

	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("git command failed: %v", err)), nil
	}

	// Parse the log output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	evolution := make([]map[string]interface{}, 0)

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}

		commitHash := parts[0]
		author := parts[1]
		date := parts[2]

		// Get file size at this commit
		sizeCmd := s.gitCommand(project.RootPath, "ls-tree", "-l", commitHash, filePath)
		sizeOutput, err := s.runGitCommand(ctx, sizeCmd)

		var fileSize int64 = 0
		if err == nil && sizeOutput != "" {
			parts := strings.Fields(sizeOutput)
			if len(parts) >= 4 {
				fmt.Sscanf(parts[3], "%d", &fileSize)
			}
		}

		evolution = append(evolution, map[string]interface{}{
			"commit_hash": commitHash,
			"author":      author,
			"date":        date,
			"file_size":   fileSize,
		})
	}

	// Calculate change statistics
	changeStats := s.calculateEvolutionStats(evolution)

	result := map[string]interface{}{
		"project_id":   projectID,
		"project_name": project.Name,
		"file_path":    filePath,
		"evolution":    evolution,
		"commit_count": len(evolution),
		"statistics":   changeStats,
	}

	return mcp.NewToolResultJSON(result)
}

// handleAnalyzeCommitPatterns analyzes commit patterns over time
func (s *Server) handleAnalyzeCommitPatterns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	timeRange := "90d"
	if tr, ok := args["time_range"].(string); ok && tr != "" {
		timeRange = tr
	}
	author, _ := args["author"].(string)

	// Build git log command with specific format
	cmdArgs := []string{"log", "--pretty=format:%an|%ad", "--date=format:%Y-%m-%d", fmt.Sprintf("--since=%s ago", timeRange)}
	if author != "" {
		cmdArgs = append(cmdArgs, "--author", author)
	}

	cmd := s.gitCommand(project.RootPath, cmdArgs...)

	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("git command failed: %v", err)), nil
	}

	// Analyze patterns
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commitsByAuthor := make(map[string]int)
	commitsByDate := make(map[string]int)
	totalCommits := 0

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}

		commitAuthor := parts[0]
		commitDate := parts[1]

		commitsByAuthor[commitAuthor]++
		commitsByDate[commitDate]++
		totalCommits++
	}

	// Calculate daily average
	dateCount := len(commitsByDate)
	dailyAverage := 0.0
	if dateCount > 0 {
		dailyAverage = float64(totalCommits) / float64(dateCount)
	}

	result := map[string]interface{}{
		"project_id":        projectID,
		"project_name":      project.Name,
		"time_range":        timeRange,
		"total_commits":     totalCommits,
		"daily_average":     dailyAverage,
		"commits_by_author": commitsByAuthor,
		"commits_by_date":   commitsByDate,
		"active_authors":    len(commitsByAuthor),
		"days_with_commits": dateCount,
	}

	if author != "" {
		result["filter_author"] = author
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetProjectTimeline provides project timeline analysis
func (s *Server) handleGetProjectTimeline(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	aspect := "all"
	if a, ok := args["aspect"].(string); ok && a != "" {
		aspect = a
	}

	timeline := map[string]interface{}{
		"project_id":   projectID,
		"project_name": project.Name,
	}

	// Get metadata for basic timeline info
	metadata, err := s.metadata.GetMetadata(projectID)
	if err == nil && metadata != nil {
		timeline["first_commit"] = metadata.FirstCommitAt
		timeline["last_commit"] = metadata.LastCommitAt
		timeline["total_commits"] = metadata.CommitCount
		timeline["contributor_count"] = metadata.ContributorCount
		timeline["tag_count"] = metadata.TagCount
	}

	// Add commits timeline
	if aspect == "all" || aspect == "commits" {
		commitsTimeline, err := s.getCommitsTimeline(project.RootPath, 50)
		if err == nil {
			timeline["commits"] = commitsTimeline
		}
	}

	// Add releases timeline
	if aspect == "all" || aspect == "releases" {
		releasesTimeline, err := s.getReleasesTimeline(project.RootPath)
		if err == nil {
			timeline["releases"] = releasesTimeline
		}
	}

	// Add contributors timeline
	if aspect == "all" || aspect == "contributors" {
		contributorsTimeline, err := s.getContributorsTimeline(project.RootPath)
		if err == nil {
			timeline["contributors"] = contributorsTimeline
		}
	}

	return mcp.NewToolResultJSON(timeline)
}

// Helper functions for historical analysis

func (s *Server) parseGitLogOutput(output string) []map[string]interface{} {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commits := make([]map[string]interface{}, 0)

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		commit := map[string]interface{}{
			"hash":         parts[0],
			"author":       parts[1],
			"author_email": parts[2],
			"date":         parts[3],
			"message":      parts[4],
		}

		commits = append(commits, commit)
	}

	return commits
}

func (s *Server) calculateEvolutionStats(evolution []map[string]interface{}) map[string]interface{} {
	if len(evolution) == 0 {
		return map[string]interface{}{
			"total_changes": 0,
			"size_variance": 0,
		}
	}

	totalSize := int64(0)
	minSize := int64(-1)
	maxSize := int64(0)

	for _, commit := range evolution {
		size := int64(0)
		if sz, ok := commit["file_size"].(int64); ok {
			size = sz
		}

		totalSize += size
		if minSize < 0 || size < minSize {
			minSize = size
		}
		if size > maxSize {
			maxSize = size
		}
	}

	averageSize := int64(0)
	if len(evolution) > 0 {
		averageSize = totalSize / int64(len(evolution))
	}

	return map[string]interface{}{
		"total_changes": len(evolution),
		"average_size":  averageSize,
		"min_size":      minSize,
		"max_size":      maxSize,
		"size_variance": maxSize - minSize,
	}
}

func (s *Server) gitCommand(rootPath string, args ...string) []string {
	return append([]string{"-C", rootPath, "git"}, args...)
}

func (s *Server) runGitCommand(ctx context.Context, cmd []string) (string, error) {
	if len(cmd) < 3 {
		return "", fmt.Errorf("invalid git command")
	}

	// cmd is expected to be ["-C", "rootPath", "git", ...args]
	execCmd := exec.CommandContext(ctx, cmd[2], cmd[3:]...)
	execCmd.Dir = cmd[1]

	// Set environment to ensure consistent git behavior
	execCmd.Env = append(execCmd.Env,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_DATE=",
		"GIT_COMMITTER_DATE=",
	)

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

func (s *Server) getCommitsTimeline(rootPath string, limit int) ([]map[string]interface{}, error) {
	// Get recent commits timeline
	cmdArgs := []string{"log", "--pretty=format:%H|%an|%ad|%s", "--date=iso-strict", "-n", fmt.Sprintf("%d", limit)}
	cmd := s.gitCommand(rootPath, cmdArgs...)

	ctx := context.Background()
	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return s.parseGitLogOutput(output), nil
}

func (s *Server) getReleasesTimeline(rootPath string) ([]map[string]interface{}, error) {
	// Get tags/releases timeline
	cmdArgs := []string{"tag", "--sort=-creatordate", "--format=%(refname:short)|%(creatordate:iso-strict)|%(subject)"}
	cmd := s.gitCommand(rootPath, cmdArgs...)

	ctx := context.Background()
	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	releases := make([]map[string]interface{}, 0)

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}

		release := map[string]interface{}{
			"tag":  parts[0],
			"date": parts[1],
		}

		if len(parts) >= 3 {
			release["subject"] = parts[2]
		}

		releases = append(releases, release)
	}

	return releases, nil
}

func (s *Server) getContributorsTimeline(rootPath string) ([]map[string]interface{}, error) {
	// Get contributors timeline with commit counts
	cmdArgs := []string{"log", "--pretty=format:%an|%ae", "--date=iso-strict"}
	cmd := s.gitCommand(rootPath, cmdArgs...)

	ctx := context.Background()
	output, err := s.runGitCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	contributors := make(map[string]map[string]int)

	for _, line := range lines {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 1 {
			continue
		}

		author := parts[0]

		key := author
		if contributors[key] == nil {
			contributors[key] = make(map[string]int)
			contributors[key]["commits"] = 0
		}
		contributors[key]["commits"]++
	}

	result := make([]map[string]interface{}, 0)
	for author, stats := range contributors {
		result = append(result, map[string]interface{}{
			"author":       author,
			"commit_count": stats["commits"],
		})
	}

	return result, nil
}
