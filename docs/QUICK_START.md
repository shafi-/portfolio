# Portfolio Quick Start Guide

Get Portfolio up and running in under 5 minutes with this streamlined installation process.

## 🚀 One-Command Install (Recommended)

The fastest way to get Portfolio running:

```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
portfolio init
```

That's it! Portfolio is now installed and initialized.

## 📋 What Happens During Installation

The install script:
1. **Detects your platform** (macOS/ Linux, Intel/ ARM)
2. **Downloads the latest binary** from GitHub releases
3. **Installs to `/usr/local/bin`** (requires sudo)
4. **Verifies installation** by running `portfolio --version`
5. **Shows you next steps** to get started

## ✅ Verify Installation

```bash
# Check Portfolio is working
portfolio --version

# Expected output: portfolio version 0.2.0 (commit: dev, built: unknown)
```

## 🎯 Quick Start Commands

```bash
# Initialize Portfolio (creates database, sets defaults)
portfolio init

# Discover your projects automatically
portfolio discover

# Check Portfolio status and health
portfolio status

# Run diagnostics if needed
portfolio doctor
```

## 🔧 Alternative Installation Methods

### Method 2: Manual Binary Download

```bash
# Download for your platform (macOS Intel shown)
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-amd64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/

# Verify
portfolio --version
```

### Method 3: Build from Source

```bash
# Clone repository
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# Build binary
go build -o portfolio ./cmd/portfolio

# Install to PATH
sudo mv portfolio /usr/local/bin/
```

### Method 4: Homebrew (macOS)

```bash
brew tap shafi-/portfolio
brew install portfolio
```

## 🎨 First-Time Setup

### 1. Initialize Portfolio

```bash
portfolio init
```

This creates:
- `~/.portfolio/portfolio.db` (SQLite database)
- `~/.portfolio/config.toml` (configuration file)
- `~/.portfolio/portfolio.log` (log file)

### 2. Discover Projects

```bash
# Discover projects in your home directory
portfolio discover

# Or specify custom directories
portfolio config set roots /path/to/projects
portfolio discover
```

### 3. Verify Projects

```bash
# List discovered projects
portfolio list

# Check a specific project
portfolio get <project-id>
```

## 🤖 AI Agent Integration

### Claude Code (Recommended)

```bash
# Install MCP server
portfolio install claude

# Install Portfolio skill
portfolio skill install

# Verify integration
portfolio doctor claude
```

### OpenCode (Beta)

```bash
# Install OpenCode integration
portfolio install opencode

# Verify integration
portfolio doctor opencode
```

## 📚 Next Steps

1. **Read the User Manual:** `portfolio manual`
2. **Explore the MCP tools:** 26 tools available for AI agents
3. **Check documentation:** https://github.com/shafi-/portfolio
4. **Join discussions:** https://github.com/shafi-/portfolio/discussions

## 🏗️ Core Concepts

- **Install → Initialize → Forget:** Portfolio runs automatically, keeping your project knowledge current
- **AI-First Interface:** Primary interaction through AI coding agents via MCP
- **Local-First:** All data stays on your machine
- **Deterministic Engine:** Repeatable project discovery and metadata extraction
- **AI-Enriched:** Semantic analysis by AI agents

## 🎯 Common Use Cases

### "What authentication mechanisms are used across my projects?"

```bash
# Via Claude Code
User: What authentication mechanisms are used in my portfolio?

Claude: [Searches projects using Portfolio tools]
Your portfolio has 4 projects using authentication:
1. user-auth-service - JWT tokens with refresh rotation
2. api-gateway - OAuth 2.0 with PKCE
3. mobile-app - Biometric auth + JWT
4. admin-panel - Session-based auth
```

### "Find all projects using TypeScript"

```bash
# Via search
portfolio search "TypeScript"
```

### "What needs analysis?"

```bash
# Via MCP tools
# Claude Code automatically identifies projects needing analysis
portfolio analyze --check-stale
```

## 🔍 Troubleshooting

### Installation Issues

**Permission denied:**
```bash
# Run with sudo
sudo bash install.sh
```

**Platform not supported:**
```bash
# Build from source instead
git clone https://github.com/shafi-/portfolio.git
cd portfolio && go build -o portfolio ./cmd/portfolio
```

### First Run Issues

**Database error:**
```bash
# Reinitialize
rm ~/.portfolio/portfolio.db
portfolio init
```

**No projects found:**
```bash
# Check your roots
portfolio config list roots

# Add your projects directory
portfolio config set roots ~/projects
portfolio discover
```

## 📖 Learn More

- **Full Documentation:** https://github.com/shafi-/portfolio
- **Release Notes:** Check releases page on GitHub
- **Contributing:** See CONTRIBUTING.md
- **Architecture:** See docs/Architecture.md

## 🎉 You're Ready!

Portfolio is now installed and ready to help you and your AI agents understand your entire software portfolio. Try asking Claude Code about your projects:

```bash
# In Claude Code
"What projects do I have?"
"What technologies are used across my portfolio?"
"Analyze my authentication systems"
```

---

**Portfolio Version:** 0.2.0  
**Repository:** https://github.com/shafi-/portfolio  
**License:** MIT
