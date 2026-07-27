# Portfolio Release Scripts

Automated release scripts for Portfolio project.

## Quick Release

```bash
# Simple one-command release
./scripts/make-release.sh v0.2.0
```

## Advanced Release Script

```bash
# Full featured release script with options
./scripts/release.sh v0.2.0
```

### Options

- `--skip-tests` - Skip test execution (use with caution)
- `--dry-run` - Show what would be done without executing
- `--push-only` - Only push existing release branch

### Examples

```bash
# Standard release
./scripts/release.sh v0.2.0

# Dry run to see what would happen
./scripts/release.sh v0.2.0 --dry-run

# Skip tests for quick iteration
./scripts/release.sh v0.2.0 --skip-tests

# Just push existing branch
./scripts/release.sh v0.2.0 --push-only
```

## What the Script Does

1. **Validates version format** (must be like v1.2.3)
2. **Checks current git status** (ensures on main branch)
3. **Updates version numbers** in:
   - `internal/version/version.go`
   - `USER_MANUAL.md`
4. **Runs full test suite** (unless skipped)
5. **Builds binary** to verify compilation
6. **Creates release branch** `release-v1.2.3`
7. **Commits changes** with descriptive message
8. **Pushes to GitHub**
9. **Creates Pull Request** automatically (if gh CLI available)

## Requirements

- **Git** - For branch management
- **Go 1.22+** - For building and testing
- **GitHub CLI (optional)** - For automatic PR creation

### Install GitHub CLI (optional)

```bash
# macOS
brew install gh

# Linux
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
```

## Release Workflow

```bash
# 1. Run the release script
./scripts/release.sh v0.2.0

# 2. Review the created PR
# Link will be shown in output

# 3. Merge the PR
# This triggers the automated release workflow

# 4. Verify release on GitHub
# Check https://github.com/shafi-/portfolio/releases
```

## Safety Features

- **Dry-run mode** - Preview changes without executing
- **Version validation** - Ensures proper semantic versioning
- **Test requirements** - Runs tests by default
- **Build verification** - Ensures binary builds correctly
- **Git status checks** - Prevents accidental commits
- **Error handling** - Exits on any failure with clear messages

## Troubleshooting

### Script fails with permission denied
```bash
chmod +x scripts/release.sh
chmod +x scripts/make-release.sh
```

### Tests fail
```bash
# Run tests manually to see detailed output
go test -race ./...

# Fix issues, then run release script
./scripts/release.sh v0.2.0
```

### Build fails
```bash
# Try building manually
go build -o portfolio ./cmd/portfolio
./portfolio --version
```

### GitHub CLI not available
```bash
# Script will work without gh CLI
# You'll just need to create PR manually
# Link will be provided in output
```

## Next Steps After Release

1. **Verify GitHub release** was created successfully
2. **Test installation** on clean system
3. **Update documentation** if needed
4. **Announce release** on GitHub Discussions
5. **Merge release branch** back to main

---

For detailed release process, see `docs/RELEASE_PROCESS.md`
