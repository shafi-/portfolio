# Homebrew Installation for Portfolio

Portfolio is available via Homebrew for macOS users.

## Quick Install

```bash
brew tap shafi-/portfolio
brew install portfolio
```

## What This Does

1. **Adds Portfolio tap** to your Homebrew installation
2. **Installs the latest Portfolio binary**
3. **Sets up proper symlinks** in your PATH
4. **Enables easy updates** via `brew upgrade`

## Verify Installation

```bash
portfolio --version
# Expected: portfolio version 0.2.0 (commit: <hash>, built: <date>)
```

## Update Portfolio

```bash
brew upgrade portfolio
```

## Uninstall Portfolio

```bash
brew uninstall portfolio
brew untap shafi-/portfolio
```

## How It Works

The Homebrew formula:
- Downloads the official GitHub release binary
- Verifies checksums for security
- Installs to `/usr/local/Cellar/portfolio/`
- Symlinks to `/usr/local/bin/portfolio`

## Advantages of Homebrew Installation

- **Official binaries** from GitHub releases
- **Checksum verification** for security
- **Easy updates** with single command
- **Dependency management** handled by Homebrew
- **Clean uninstall** with no leftover files

## Troubleshooting

### Tap not found
```bash
# If the tap doesn't exist yet, you can install from source:
brew install shafi-/portfolio/portfolio
```

### Permission issues
```bash
# If you get permission errors, fix Homebrew permissions:
sudo chown -R $(whoami) /usr/local/Cellar
sudo chown -R $(whoami) /usr/local/bin
```

### Version mismatch
```bash
# Force upgrade to latest:
brew uninstall portfolio
brew install shafi-/portfolio/portfolio
```

## Alternative: Install from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# Install via Homebrew from source
brew install --HEAD ./portfolio.rb
```

## Compared to One-Click Install

**Homebrew:**
```bash
brew tap shafi-/portfolio
brew install portfolio
```
- Requires Homebrew installed
- Easy updates and management
- Official binaries with checksums

**One-Click Install:**
```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
```
- No prerequisites
- Direct download
- Manual updates needed

## Automatic Updates

To keep Portfolio updated automatically:

```bash
# Add to your cron or launchd
brew upgrade portfolio
```

Or use a tool like `brew-upgrade-all` to keep all packages updated.

---

**For more installation methods, see the [User Manual](../USER_MANUAL.md)**
