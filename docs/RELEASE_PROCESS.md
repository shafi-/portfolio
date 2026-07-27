# Portfolio Release Process

## Overview

Portfolio uses a **release branch** workflow where releases are created by merging a PR into a `release-*` branch, which automatically triggers the release workflow.

## Release Workflow

### 1. Create Release Branch

```bash
# Create release branch from main
git checkout main
git pull
git checkout -b release-v0.2.0
```

### 2. Update Version Numbers

Update these files with the new version number:

- `internal/version/version.go` - Update `version = "0.2.0"`
- `USER_MANUAL.md` - Update version header and add release notes
- Any other version-specific documentation

### 3. Test Build Locally

```bash
# Run full test suite
go test -race ./...
go vet ./...

# Build binary
go build -o portfolio ./cmd/portfolio

# Test version output
./portfolio --version
```

### 4. Push Release Branch

```bash
git add internal/version/version.go USER_MANUAL.md
git commit -m "release: prepare v0.2.0 release"
git push origin release-v0.2.0
```

### 5. Create and Merge PR

1. Create PR: `release-v0.2.0` → `release-v0.2.0`
2. Include release notes in PR description
3. Get approval and merge

### 6. Automatic Release

Once the PR is merged to the release branch, the GitHub Actions workflow will:

1. Run full test suite
2. Build release binary
3. Create GitHub release with assets
4. Generate release notes from commits

### 7. Post-Release Tasks

After successful release:

1. **Merge release branch to main:**
   ```bash
   git checkout main
   git merge release-v0.2.0
   git push origin main
   ```

2. **Update documentation links** (if needed)

3. **Install and test release:**
   ```bash
   curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
   portfolio --version
   ```

4. **Announce release** (GitHub Discussions, README badges, etc.)

## Manual Release Trigger

If needed, you can manually trigger a release:

```bash
# Via GitHub Actions UI
1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Select branch and enter version (e.g., v0.2.0)
4. Run workflow
```

## Rollback Procedure

If something goes wrong:

1. **Delete the GitHub release** (if created)
2. **Create new release branch** with incremented version (v0.2.1)
3. **Fix issues** and follow normal release process
4. **Document issues** in release notes

## Version Naming

Use semantic versioning:
- **v0.1.0** - Initial release
- **v0.2.0** - Feature additions
- **v0.2.1** - Bug fixes
- **v1.0.0** - Major milestone

## Release Checklist

- [ ] All tests passing
- [ ] Version numbers updated
- [ ] Release notes prepared
- [ ] Documentation updated
- [ ] Release branch pushed
- [ ] PR created and approved
- [ ] Release workflow triggered successfully
- [ ] Release assets verified
- [ ] One-click install tested
- [ ] Main branch updated
- [ ] Announcement posted

## Troubleshooting

### Workflow fails on test
- Check test logs in GitHub Actions
- Fix issues in release branch or create new branch

### Release assets incorrect
- Delete GitHub release
- Fix version numbers
- Re-run workflow or create new release branch

### Version mismatch
- Ensure `internal/version/version.go` is updated
- Check release branch name matches version
