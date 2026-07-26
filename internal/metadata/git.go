package metadata

import (
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type GitResult struct {
	GitHead           *string
	DefaultBranch     *string
	LastCommitAt      *time.Time
	LastModifiedAt    *time.Time
	CommitCount       int
	FirstCommitAt     *time.Time
	CommitVelocity90d int
	ContributorCount  int
	TagCount          int
	RemoteURL         *string
	IsPublished       bool
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

	// First (oldest) commit timestamp — `--reverse` makes the oldest commit first.
	firstCommitStr := firstLine(runGit(root, "log", "--reverse", "--format=%ct", "HEAD"))
	if firstCommitStr != "" {
		if ts, err := strconv.ParseInt(firstCommitStr, 10, 64); err == nil {
			t := time.Unix(ts, 0)
			result.FirstCommitAt = &t
		}
	}

	// Commits in the last 90 days — activity / "is this cared about right now".
	velocityStr := runGit(root, "rev-list", "--count", "--since=90 days ago", "HEAD")
	if velocityStr != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(velocityStr)); err == nil {
			result.CommitVelocity90d = v
		}
	}

	// Unique contributor count via author emails.
	result.ContributorCount = countUniqueLines(runGit(root, "log", "--format=%ae", "HEAD"))

	// Tag count — release maturity.
	result.TagCount = countNonEmptyLines(runGit(root, "tag", "--list"))

	// Remote URL + published flag (host is a known public forge).
	remoteURL := runGit(root, "remote", "get-url", "origin")
	if remoteURL != "" {
		result.RemoteURL = &remoteURL
		result.IsPublished = isPublishedRemote(remoteURL)
	}

	lastModifiedAt := getLastModifiedAt(root)
	result.LastModifiedAt = lastModifiedAt

	return result, nil
}

// firstLine returns the first non-empty line of s.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// countUniqueLines counts non-empty unique lines (case-insensitive).
func countUniqueLines(s string) int {
	seen := make(map[string]struct{})
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			seen[strings.ToLower(line)] = struct{}{}
		}
	}
	return len(seen)
}

// countNonEmptyLines counts non-empty lines.
func countNonEmptyLines(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// knownPublicForges are remote hosts that imply a project is published.
var knownPublicForges = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"bitbucket.org": true,
	"codeberg.org":  true,
	"git.sr.ht":     true,
}

// isPublishedRemote reports whether the remote URL points at a known public forge.
func isPublishedRemote(remote string) bool {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return false
	}
	// SSH form: git@github.com:owner/repo.git
	if strings.HasPrefix(remote, "git@") {
		if i := strings.Index(remote, "@"); i != -1 {
			host := remote[i+1:]
			if j := strings.Index(host, ":"); j != -1 {
				host = host[:j]
			}
			return knownPublicForges[strings.ToLower(host)]
		}
	}
	// HTTPS form: https://github.com/owner/repo.git
	if u, err := url.Parse(remote); err == nil && u.Host != "" {
		return knownPublicForges[strings.ToLower(u.Host)]
	}
	return false
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
