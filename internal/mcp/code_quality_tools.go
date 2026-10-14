package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// codeQualityTools returns tools for code complexity, test coverage, and technical debt analysis
func (s *Server) codeQualityTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("getCodeComplexity",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("path", mcp.Description("Specific file or directory path to analyze (optional)")),
				mcp.WithString("metric", mcp.Description("Complexity metric: cyclomatic, lines, functions, all (default: all)")),
			),
			Handler: s.handleGetCodeComplexity,
		},
		{
			Tool: mcp.NewTool("getTestCoverage",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithBoolean("detailed", mcp.Description("Include detailed per-file coverage (default: false)")),
			),
			Handler: s.handleGetTestCoverage,
		},
		{
			Tool: mcp.NewTool("getTechnicalDebt",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("category", mcp.Description("Debt category: complexity, duplication, issues, all (default: all)")),
			),
			Handler: s.handleGetTechnicalDebt,
		},
		{
			Tool: mcp.NewTool("analyzeCodeSmells",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("path", mcp.Description("Specific file or directory path to analyze (optional)")),
			),
			Handler: s.handleAnalyzeCodeSmells,
		},
		{
			Tool: mcp.NewTool("getCodeMetrics",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("path", mcp.Description("Specific file or directory path (optional)")),
			),
			Handler: s.handleGetCodeMetrics,
		},
	}
}

// handleGetCodeComplexity calculates code complexity metrics
func (s *Server) handleGetCodeComplexity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	targetPath, _ := args["path"].(string)
	metric := "all"
	if m, ok := args["metric"].(string); ok && m != "" {
		metric = m
	}

	// Get files to analyze
	files := []string{}
	basePath := project.RootPath
	if targetPath != "" {
		basePath = filepath.Join(project.RootPath, targetPath)
	}

	// Walk the directory to find code files
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip common directories to ignore
			if s.shouldSkip(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		// Only include code files based on extension
		if s.isCodeFile(path) {
			relPath, _ := filepath.Rel(project.RootPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	// Calculate complexity metrics
	complexityResults := map[string]interface{}{
		"project_id":     projectID,
		"project_name":   project.Name,
		"base_path":      targetPath,
		"files_analyzed": len(files),
		"total_files":    len(files),
	}

	if metric == "all" || metric == "cyclomatic" {
		cyclomaticData := s.calculateCyclomaticComplexity(project.RootPath, files)
		complexityResults["cyclomatic_complexity"] = cyclomaticData
	}

	if metric == "all" || metric == "lines" {
		linesData := s.calculateLineMetrics(project.RootPath, files)
		complexityResults["line_metrics"] = linesData
	}

	if metric == "all" || metric == "functions" {
		functionData := s.analyzeFunctions(project.RootPath, files)
		complexityResults["function_analysis"] = functionData
	}

	return mcp.NewToolResultJSON(complexityResults)
}

// handleGetTestCoverage retrieves test coverage information
func (s *Server) handleGetTestCoverage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	detailed := false
	if d, ok := args["detailed"].(bool); ok {
		detailed = d
	}

	// Try different coverage tools based on project type
	metadata, err := s.metadata.GetMetadata(projectID)
	if err != nil {
		return mcp.NewToolResultError("failed to get project metadata"), nil
	}

	coverageResults := map[string]interface{}{
		"project_id":         projectID,
		"project_name":       project.Name,
		"coverage_available": false,
		"tools_checked":      []string{},
	}

	// Try Go coverage
	if metadata.LanguageSummary != "" && strings.Contains(metadata.LanguageSummary, "Go") {
		goCoverage, err := s.getGoCoverage(project.RootPath, detailed)
		if err == nil && goCoverage != nil {
			coverageResults["coverage_available"] = true
			coverageResults["go_coverage"] = goCoverage
			coverageResults["tools_checked"] = append(coverageResults["tools_checked"].([]string), "go")
		}
	}

	// Try JavaScript/TypeScript coverage
	if metadata.LanguageSummary != "" && (strings.Contains(metadata.LanguageSummary, "JavaScript") || strings.Contains(metadata.LanguageSummary, "TypeScript")) {
		jsCoverage, err := s.getJSCoverage(project.RootPath, detailed)
		if err == nil && jsCoverage != nil {
			coverageResults["coverage_available"] = true
			coverageResults["javascript_coverage"] = jsCoverage
			coverageResults["tools_checked"] = append(coverageResults["tools_checked"].([]string), "javascript")
		}
	}

	// Try Python coverage
	if metadata.LanguageSummary != "" && strings.Contains(metadata.LanguageSummary, "Python") {
		pyCoverage, err := s.getPythonCoverage(project.RootPath, detailed)
		if err == nil && pyCoverage != nil {
			coverageResults["coverage_available"] = true
			coverageResults["python_coverage"] = pyCoverage
			coverageResults["tools_checked"] = append(coverageResults["tools_checked"].([]string), "python")
		}
	}

	// If no coverage found, check for test files as fallback
	if !coverageResults["coverage_available"].(bool) {
		testFiles := s.findTestFiles(project.RootPath)
		coverageResults["test_files_found"] = len(testFiles)
		coverageResults["test_files"] = testFiles
	}

	return mcp.NewToolResultJSON(coverageResults)
}

// handleGetTechnicalDebt analyzes technical debt indicators
func (s *Server) handleGetTechnicalDebt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	category := "all"
	if c, ok := args["category"].(string); ok && c != "" {
		category = c
	}

	debtResults := map[string]interface{}{
		"project_id":       projectID,
		"project_name":     project.Name,
		"total_debt_score": 0,
		"debt_categories":  map[string]interface{}{},
	}

	// Get all code files
	files := []string{}
	err = filepath.Walk(project.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if s.shouldSkip(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if s.isCodeFile(path) {
			relPath, _ := filepath.Rel(project.RootPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	if category == "all" || category == "complexity" {
		complexityDebt := s.analyzeComplexityDebt(project.RootPath, files)
		debtResults["debt_categories"].(map[string]interface{})["complexity"] = complexityDebt
		debtResults["total_debt_score"] = debtResults["total_debt_score"].(int) + complexityDebt["score"].(int)
	}

	if category == "all" || category == "duplication" {
		duplicationDebt := s.analyzeDuplicationDebt(project.RootPath, files)
		debtResults["debt_categories"].(map[string]interface{})["duplication"] = duplicationDebt
		debtResults["total_debt_score"] = debtResults["total_debt_score"].(int) + duplicationDebt["score"].(int)
	}

	if category == "all" || category == "issues" {
		issuesDebt := s.analyzeCodeIssues(project.RootPath, files)
		debtResults["debt_categories"].(map[string]interface{})["issues"] = issuesDebt
		debtResults["total_debt_score"] = debtResults["total_debt_score"].(int) + issuesDebt["score"].(int)
	}

	return mcp.NewToolResultJSON(debtResults)
}

// handleAnalyzeCodeSmells detects code smells and anti-patterns
func (s *Server) handleAnalyzeCodeSmells(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	targetPath, _ := args["path"].(string)

	// Get code files
	files := []string{}
	basePath := project.RootPath
	if targetPath != "" {
		basePath = filepath.Join(project.RootPath, targetPath)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if s.shouldSkip(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if s.isCodeFile(path) {
			relPath, _ := filepath.Rel(project.RootPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	// Analyze code smells
	smells := s.detectCodeSmells(project.RootPath, files)

	result := map[string]interface{}{
		"project_id":     projectID,
		"project_name":   project.Name,
		"target_path":    targetPath,
		"files_analyzed": len(files),
		"code_smells":    smells,
		"total_smells":   len(smells),
	}

	return mcp.NewToolResultJSON(result)
}

// handleGetCodeMetrics provides comprehensive code metrics
func (s *Server) handleGetCodeMetrics(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	targetPath, _ := args["path"].(string)

	// Get code files
	files := []string{}
	basePath := project.RootPath
	if targetPath != "" {
		basePath = filepath.Join(project.RootPath, targetPath)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if s.shouldSkip(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if s.isCodeFile(path) {
			relPath, _ := filepath.Rel(project.RootPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	// Calculate comprehensive metrics
	metrics := s.calculateComprehensiveMetrics(project.RootPath, files)

	result := map[string]interface{}{
		"project_id":     projectID,
		"project_name":   project.Name,
		"target_path":    targetPath,
		"files_analyzed": len(files),
		"metrics":        metrics,
	}

	return mcp.NewToolResultJSON(result)
}

// Helper functions for code quality analysis

func (s *Server) isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExtensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".jsx":   true,
		".tsx":   true,
		".py":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".rb":    true,
		".php":   true,
		".rs":    true,
		".kt":    true,
		".swift": true,
		".scala": true,
	}
	return codeExtensions[ext]
}

func (s *Server) calculateCyclomaticComplexity(rootPath string, files []string) map[string]interface{} {
	totalComplexity := 0
	maxComplexity := 0
	complexityByFile := make(map[string]int)

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		complexity := s.calculateFileComplexity(string(content))
		totalComplexity += complexity
		if complexity > maxComplexity {
			maxComplexity = complexity
		}
		complexityByFile[file] = complexity
	}

	averageComplexity := 0.0
	if len(files) > 0 {
		averageComplexity = float64(totalComplexity) / float64(len(files))
	}

	return map[string]interface{}{
		"total_complexity":   totalComplexity,
		"average_complexity": averageComplexity,
		"max_complexity":     maxComplexity,
		"files_count":        len(files),
		"complexity_by_file": complexityByFile,
	}
}

func (s *Server) calculateFileComplexity(content string) int {
	// Simple cyclomatic complexity estimation
	// Count decision points: if, else, for, while, case, catch, etc.
	decisionPatterns := []string{
		`\bif\b`, `\belse\b`, `\bfor\b`, `\bwhile\b`, `\bcase\b`,
		`\bcatch\b`, `\bswitch\b`, `\b&&\b`, `\b\|\|\b`, `\?`, `try`,
	}

	complexity := 1 // Base complexity
	for _, pattern := range decisionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringIndex(content, -1)
		complexity += len(matches)
	}

	return complexity
}

func (s *Server) calculateLineMetrics(rootPath string, files []string) map[string]interface{} {
	totalLines := 0
	totalCodeLines := 0
	totalCommentLines := 0
	totalBlankLines := 0
	linesByFile := make(map[string]map[string]int)

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		codeLines := 0
		commentLines := 0
		blankLines := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				blankLines++
			} else if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
				commentLines++
			} else {
				codeLines++
			}
		}

		totalLines += len(lines)
		totalCodeLines += codeLines
		totalCommentLines += commentLines
		totalBlankLines += blankLines

		linesByFile[file] = map[string]int{
			"total":    len(lines),
			"code":     codeLines,
			"comments": commentLines,
			"blank":    blankLines,
		}
	}

	return map[string]interface{}{
		"total_lines":   totalLines,
		"code_lines":    totalCodeLines,
		"comment_lines": totalCommentLines,
		"blank_lines":   totalBlankLines,
		"files_count":   len(files),
		"lines_by_file": linesByFile,
	}
}

func (s *Server) analyzeFunctions(rootPath string, files []string) map[string]interface{} {
	totalFunctions := 0
	functionsByFile := make(map[string]int)

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Count function definitions based on file extension
		functionCount := s.countFunctionsInFile(file, string(content))
		totalFunctions += functionCount
		functionsByFile[file] = functionCount
	}

	return map[string]interface{}{
		"total_functions":      totalFunctions,
		"files_with_functions": len(functionsByFile),
		"functions_by_file":    functionsByFile,
	}
}

func (s *Server) countFunctionsInFile(filename string, content string) int {
	ext := strings.ToLower(filepath.Ext(filename))
	functionCount := 0

	switch ext {
	case ".go":
		// Count function declarations
		re := regexp.MustCompile(`func\s+\w+\s*\(`)
		matches := re.FindAllStringIndex(content, -1)
		functionCount = len(matches)
	case ".py":
		// Count function definitions
		re := regexp.MustCompile(`def\s+\w+\s*\(`)
		matches := re.FindAllStringIndex(content, -1)
		functionCount = len(matches)
	case ".js", ".ts", ".jsx", ".tsx":
		// Count function declarations and arrow functions
		re := regexp.MustCompile(`function\s+\w+\s*\(|\w+\s*:\s*\(.*\)\s*=>|const\s+\w+\s*=\s*\(.*\)\s*=>`)
		matches := re.FindAllStringIndex(content, -1)
		functionCount = len(matches)
	case ".java":
		// Count method declarations
		re := regexp.MustCompile(`(?:public|private|protected)?\s*(?:static\s+)?(?:\w+\s+)+(\w+)\s*\([^)]*\)\s*(?:throws\s+[\w\s]+)?\s*\{`)
		matches := re.FindAllStringIndex(content, -1)
		functionCount = len(matches)
	default:
		// Generic function detection
		re := regexp.MustCompile(`(func|function|def|method)\s+\w+`)
		matches := re.FindAllStringIndex(content, -1)
		functionCount = len(matches)
	}

	return functionCount
}

func (s *Server) getGoCoverage(rootPath string, detailed bool) (map[string]interface{}, error) {
	// Check if coverage files exist
	coverageFile := filepath.Join(rootPath, "coverage.out")
	if _, err := s.osFS.Stat(coverageFile); err != nil {
		// Try to generate coverage
		cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-covermode=atomic")
		cmd.Dir = rootPath
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to generate go coverage: %v", err)
		}
	}

	content, err := s.osFS.ReadFile(coverageFile)
	if err != nil {
		return nil, err
	}

	// Parse coverage output
	lines := strings.Split(string(content), "\n")
	totalCoverage := 0.0
	fileCount := 0
	fileCoverage := make(map[string]float64)

	for _, line := range lines {
		if strings.HasPrefix(line, "mode:") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 {
			filePath := parts[0]
			var covered, total int
			fmt.Sscanf(parts[1], "%d", &covered)
			fmt.Sscanf(parts[2], "%d", &total)

			if total > 0 {
				coverage := (float64(covered) / float64(total)) * 100.0
				fileCoverage[filePath] = coverage
				totalCoverage += coverage
				fileCount++
			}
		}
	}

	averageCoverage := 0.0
	if fileCount > 0 {
		averageCoverage = totalCoverage / float64(fileCount)
	}

	result := map[string]interface{}{
		"overall_coverage": averageCoverage,
		"files_covered":    fileCount,
	}

	if detailed {
		result["file_coverage"] = fileCoverage
	}

	return result, nil
}

func (s *Server) getJSCoverage(rootPath string, detailed bool) (map[string]interface{}, error) {
	// Check for common JS coverage files
	coverageFiles := []string{"coverage/coverage-final.json", "coverage-summary.json", "istanbul-coverage.json"}

	for _, covFile := range coverageFiles {
		_, err := s.osFS.ReadFile(filepath.Join(rootPath, covFile))
		if err == nil {
			// Parse coverage data (simplified)
			return map[string]interface{}{
				"coverage_file":      covFile,
				"coverage_available": true,
				"detailed_analysis":  "Coverage file found but detailed parsing not implemented",
			}, nil
		}
	}

	return nil, fmt.Errorf("no JavaScript coverage files found")
}

func (s *Server) getPythonCoverage(rootPath string, detailed bool) (map[string]interface{}, error) {
	// Check for common Python coverage files
	coverageFiles := []string{".coverage", "coverage.xml", ".coverage.xml"}

	for _, covFile := range coverageFiles {
		_, err := s.osFS.Stat(filepath.Join(rootPath, covFile))
		if err == nil {
			return map[string]interface{}{
				"coverage_file":      covFile,
				"coverage_available": true,
				"detailed_analysis":  "Coverage file found but detailed parsing not implemented",
			}, nil
		}
	}

	return nil, fmt.Errorf("no Python coverage files found")
}

func (s *Server) findTestFiles(rootPath string) []string {
	testFiles := []string{}
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if s.shouldSkip(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for test file patterns
		baseName := strings.ToLower(info.Name())
		if strings.Contains(baseName, "test") || strings.Contains(baseName, "spec") {
			if s.isCodeFile(path) {
				relPath, _ := filepath.Rel(rootPath, path)
				testFiles = append(testFiles, relPath)
			}
		}
		return nil
	})
	if err != nil {
		return testFiles
	}
	return testFiles
}

func (s *Server) analyzeComplexityDebt(rootPath string, files []string) map[string]interface{} {
	complexityData := s.calculateCyclomaticComplexity(rootPath, files)
	averageComplexity := complexityData["average_complexity"].(float64)
	maxComplexity := complexityData["max_complexity"].(int)

	// Calculate debt score based on complexity thresholds
	debtScore := 0
	highComplexityFiles := 0

	averageThreshold := 10.0
	maxThreshold := 20

	if averageComplexity > averageThreshold {
		debtScore += int(averageComplexity - averageThreshold)
	}
	if maxComplexity > maxThreshold {
		debtScore += maxComplexity - maxThreshold
		highComplexityFiles++
	}

	return map[string]interface{}{
		"score":                 debtScore,
		"average_complexity":    averageComplexity,
		"max_complexity":        maxComplexity,
		"high_complexity_files": highComplexityFiles,
		"description":           "Cyclomatic complexity debt indicates potentially hard-to-maintain code",
	}
}

func (s *Server) analyzeDuplicationDebt(rootPath string, files []string) map[string]interface{} {
	// Simplified duplication detection
	// In a real implementation, this would use more sophisticated algorithms
	duplicateLines := 0
	totalLines := 0

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		lineFrequency := make(map[string]int)

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 20 { // Only consider non-trivial lines
				lineFrequency[trimmed]++
				totalLines++
			}
		}

		// Count duplicated lines
		for _, count := range lineFrequency {
			if count > 1 {
				duplicateLines += count - 1
			}
		}
	}

	duplicationRatio := 0.0
	if totalLines > 0 {
		duplicationRatio = float64(duplicateLines) / float64(totalLines)
	}

	debtScore := int(duplicationRatio * 100)

	return map[string]interface{}{
		"score":                debtScore,
		"duplicate_lines":      duplicateLines,
		"total_lines_analyzed": totalLines,
		"duplication_ratio":    duplicationRatio,
		"description":          "Code duplication debt indicates maintenance risks and potential for refactoring",
	}
}

func (s *Server) analyzeCodeIssues(rootPath string, files []string) map[string]interface{} {
	issueCount := 0
	longLines := 0
	magicNumbers := 0
	largeFunctions := 0

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		functionLines := 0

		for _, line := range lines {
			// Check for overly long lines (>120 characters)
			if len(line) > 120 {
				longLines++
				issueCount++
			}

			// Check for magic numbers (standalone numeric literals)
			if matched, _ := regexp.MatchString(`\b\d{3,}\b`, line); matched {
				magicNumbers++
				issueCount++
			}

			// Track function size (simplified)
			if strings.Contains(line, "func") || strings.Contains(line, "function") || strings.Contains(line, "def ") {
				if functionLines > 50 {
					largeFunctions++
					issueCount++
				}
				functionLines = 0
			}
			functionLines++
		}
	}

	return map[string]interface{}{
		"score":           issueCount,
		"total_issues":    issueCount,
		"long_lines":      longLines,
		"magic_numbers":   magicNumbers,
		"large_functions": largeFunctions,
		"description":     "Code issues debt indicates maintainability and readability concerns",
	}
}

func (s *Server) detectCodeSmells(rootPath string, files []string) []map[string]interface{} {
	smells := []map[string]interface{}{}

	for _, file := range files {
		fullPath := filepath.Join(rootPath, file)
		content, err := s.osFS.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Detect various code smells
		contentStr := string(content)

		// Long methods
		lines := strings.Split(contentStr, "\n")
		if len(lines) > 300 {
			smells = append(smells, map[string]interface{}{
				"type":        "long_file",
				"file":        file,
				"line_count":  len(lines),
				"severity":    "medium",
				"description": "File is too long and may need refactoring",
			})
		}

		// Deep nesting
		maxNesting := s.calculateMaxNesting(contentStr)
		if maxNesting > 5 {
			smells = append(smells, map[string]interface{}{
				"type":          "deep_nesting",
				"file":          file,
				"nesting_level": maxNesting,
				"severity":      "high",
				"description":   "Deep nesting indicates complex logic that could be simplified",
			})
		}

		// God object detection (many methods)
		functionCount := s.countFunctionsInFile(file, contentStr)
		if functionCount > 20 {
			smells = append(smells, map[string]interface{}{
				"type":           "god_object",
				"file":           file,
				"function_count": functionCount,
				"severity":       "medium",
				"description":    "File may be doing too much and could be split into smaller modules",
			})
		}
	}

	return smells
}

func (s *Server) calculateMaxNesting(content string) int {
	maxNesting := 0
	currentNesting := 0

	for _, char := range content {
		if char == '{' || char == '(' {
			currentNesting++
			if currentNesting > maxNesting {
				maxNesting = currentNesting
			}
		} else if char == '}' || char == ')' {
			currentNesting--
		}
	}

	return maxNesting
}

func (s *Server) calculateComprehensiveMetrics(rootPath string, files []string) map[string]interface{} {
	metrics := make(map[string]interface{})

	// Line metrics
	lineMetrics := s.calculateLineMetrics(rootPath, files)
	metrics["line_metrics"] = lineMetrics

	// Complexity metrics
	complexityMetrics := s.calculateCyclomaticComplexity(rootPath, files)
	metrics["complexity_metrics"] = complexityMetrics

	// Function metrics
	functionMetrics := s.analyzeFunctions(rootPath, files)
	metrics["function_metrics"] = functionMetrics

	// Calculate overall quality score
	qualityScore := s.calculateQualityScore(lineMetrics, complexityMetrics, functionMetrics)
	metrics["quality_score"] = qualityScore

	return metrics
}

func (s *Server) calculateQualityScore(lineMetrics, complexityMetrics, functionMetrics map[string]interface{}) map[string]interface{} {
	score := 100.0

	// Deduct for high complexity
	if avgComplexity, ok := complexityMetrics["average_complexity"].(float64); ok {
		if avgComplexity > 10 {
			score -= (avgComplexity - 10) * 2
		}
	}

	// Deduct for poor comment ratio
	totalLines := lineMetrics["total_lines"].(int)
	commentLines := lineMetrics["comment_lines"].(int)
	if totalLines > 0 {
		commentRatio := float64(commentLines) / float64(totalLines)
		if commentRatio < 0.1 { // Less than 10% comments
			score -= 20
		} else if commentRatio < 0.2 {
			score -= 10
		}
	}

	// Ensure score stays within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	grade := "A"
	if score < 60 {
		grade = "F"
	} else if score < 70 {
		grade = "D"
	} else if score < 80 {
		grade = "C"
	} else if score < 90 {
		grade = "B"
	}

	return map[string]interface{}{
		"score":       score,
		"grade":       grade,
		"description": fmt.Sprintf("Overall code quality score based on complexity, documentation, and maintainability metrics"),
	}
}
