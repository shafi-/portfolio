# 🚀 One-Click Release Guide

## Super Quick Release

```bash
./scripts/make-release.sh v0.2.0
```

That's it! The script handles everything automatically.

## What Happens

1. ✅ Validates version format
2. ✅ Updates version numbers in code
3. ✅ Runs all tests
4. ✅ Builds release binary
5. ✅ Creates release branch
6. ✅ Commits all changes
7. ✅ Pushes to GitHub
8. ✅ Creates Pull Request
9. ✅ Shows you the PR link to merge

## Advanced Options

```bash
# Preview what would happen
./scripts/release.sh v0.2.0 --dry-run

# Skip tests (use with caution)
./scripts/release.sh v0.2.0 --skip-tests

# Just push existing branch
./scripts/release.sh v0.2.0 --push-only
```

## Requirements

- Git
- Go 1.22+
- GitHub CLI (optional, for automatic PR creation)

## Example Output

```
🚀 Portfolio Release Automation
======================================

Current version: 0.2.0
Release version: v0.2.0

📍 Checking branch...
📥 Pulling latest changes...
📝 Updating version to v0.2.0...
🧪 Running tests...
✅ All tests passing
🔨 Building release binary...
✅ Build successful
🌿 Creating release branch: release-v0.2.0
📦 Committing release changes...
✅ Changes committed
📤 Pushing release branch...
✅ Release branch pushed
🎯 Creating GitHub PR...
✅ PR created successfully

📋 Release Summary
=================
Version: v0.2.0
Branch: release-v0.2.0
Repository: shafi-/portfolio

Next steps:
1. Review PR: https://github.com/shafi-/portfolio/compare/release-v0.2.0
2. Merge PR to trigger automated release
3. Verify release assets on GitHub
4. Test installation

🎉 Release preparation complete!
```

## Safety Features

- ✅ Version format validation
- ✅ Git status checks
- ✅ Test suite execution
- ✅ Build verification
- ✅ Error handling
- ✅ Dry-run mode

## Next Steps

1. **Click the PR link** shown in output
2. **Review and merge** the PR
3. **Watch GitHub Actions** create the release
4. **Test the release** on a clean system

## Troubleshooting

**Not on main branch?** Script will prompt to switch to main.

**Tests failing?** Fix issues first, then run release script.

**No GitHub CLI?** Script works without it - you'll create PR manually.

**Need to cancel?** Just Ctrl+C - no changes are made until final push.

## Manual Fallback

If automation fails, see `docs/RELEASE_PROCESS.md` for manual steps.

---

**Ready to release?** Just run: `./scripts/make-release.sh v0.2.0`
