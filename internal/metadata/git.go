package metadata

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type GitResult struct {
	GitHead        *string
	DefaultBranch  *string
	LastCommitAt   *time.Time
	LastModifiedAt *time.Time
	CommitCount    int
}

func ExtractGitMetadata(root string) (*GitResult, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	result := &GitResult{}

	gitHead := runGit(root, "rev-parse", "HEAD")
	if gitHead != "" {
		result.GitHead = &gitHead
	}

	defaultBranch := getDefaultBranch(root)
	result.DefaultBranch = defaultBranch

	lastCommitAt := runGit(root, "log", "-1", "--format=%ct", "HEAD")
	if lastCommitAt != "" {
		ts, err := strconv.ParseInt(strings.TrimSpace(lastCommitAt), 10, 64)
		if err == nil {
			t := time.Unix(ts, 0)
			result.LastCommitAt = &t
		}
	}

	commitCountStr := runGit(root, "rev-list", "--count", "HEAD")
	if commitCountStr != "" {
		count, err := strconv.Atoi(strings.TrimSpace(commitCountStr))
		if err == nil {
			result.CommitCount = count
		}
	}

	lastModifiedAt := getLastModifiedAt(root)
	result.LastModifiedAt = lastModifiedAt

	return result, nil
}

func runGit(root string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getDefaultBranch(root string) *string {
	branch := runGit(root, "symbolic-ref", "refs/remotes/origin/HEAD")
	if branch != "" {
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		return &branch
	}

	branch = runGit(root, "rev-parse", "--abbrev-ref", "HEAD")
	if branch != "" && branch != "HEAD" {
		return &branch
	}

	return nil
}

func getLastModifiedAt(root string) *time.Time {
	lastCommitStr := runGit(root, "log", "-1", "--format=%ct", "HEAD")
	if lastCommitStr == "" {
		return nil
	}

	ts, err := strconv.ParseInt(strings.TrimSpace(lastCommitStr), 10, 64)
	if err != nil {
		return nil
	}
	lastCommitTime := time.Unix(ts, 0)

	status := runGit(root, "status", "--porcelain")
	if status == "" {
		return &lastCommitTime
	}

	lines := strings.Split(status, "\n")
	var maxModTime time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		filePath := parts[len(parts)-1]
		fullPath := root + "/" + filePath

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		if modTime.After(maxModTime) {
			maxModTime = modTime
		}
	}

	if maxModTime.After(lastCommitTime) {
		return &maxModTime
	}
	return &lastCommitTime
}
