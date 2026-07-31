# Portfolio Engine

Portfolio is a local-first project inventory and knowledge platform that enables developers and AI coding agents to understand an entire software portfolio.

## Prerequisites

- Go 1.21 or higher
- Git (for project discovery)

## Installation

### Method 1: One-Command Install (Recommended)

```bash
# Install and start in one command
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
portfolio init
```

### Method 2: Binary Release

```bash
# Download latest release for your platform
# macOS (Intel): portfolio-darwin-amd64
# macOS (Apple Silicon): portfolio-darwin-arm64
# Linux (Intel): portfolio-linux-amd64
# Linux (ARM): portfolio-linux-arm64

curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/

# Verify installation
portfolio --version
```

### Method 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# Build the CLI
go build ./cmd/portfolio

# (Optional) Install to system path
go install ./cmd/portfolio
```

## Quick Start

```bash
# Initialize Portfolio
portfolio init

# Discover your projects automatically
portfolio discover

# Check system status
portfolio status

# Run diagnostics if needed
portfolio doctor
```

## Key Features

- **🔍 Automatic Discovery**: Finds all your Git repositories automatically
- **📊 Metadata Extraction**: Extracts languages, frameworks, and dependencies
- **📚 Documentation Indexing**: Search across all project documentation
- **🤖 AI Agent Integration**: 26 MCP tools for Claude Code, OpenCode, and other AI agents
- **💾 Local-First**: All data stays on your machine
- **🔧 CLI Administration**: Simple commands for management and diagnostics

## AI Agent Integration

```bash
# Claude Code integration
portfolio install claude

# OpenCode integration  
portfolio install opencode

# Verify integration
portfolio doctor claude
```

## Development

### Project Structure

```
portfolio/
├── cmd/portfolio/        # CLI entry point
├── internal/
│   ├── config/          # Configuration system
│   ├── database/        # SQLite database
│   ├── logging/         # Structured logging
│   └── cli/             # CLI commands
└── pkg/models/          # Shared data structures
```

### Development Setup

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build for development
go build ./cmd/portfolio

# Run development binary
./portfolio --help
```

## Documentation

Full documentation available in [docs/](docs/)

**Getting Started:**
- [Quick Start Guide](docs/QUICK_START.md) - Get up and running in 5 minutes
- [User Manual](USER_MANUAL.md) - Complete reference guide
- [Homebrew Setup](docs/HOMEBREW_SETUP.md) - macOS package management

**Technical Documentation:**
- [Knowledge Model](docs/KnowledgeModel.md) - Canonical domain model
- [Platform Specification](docs/PlatformSpecification.md) - Implementation contracts
- [Product Requirements](docs/PRD.md) - Vision and goals
- [Engineering Guidelines](docs/Guideline.md) - Development principles

**Development:**
- [Release Process](docs/RELEASE_PROCESS.md) - How to make releases
- [One-Click Release](RELEASE_GUIDE.md) - Release automation guide

## License

MIT License - see [LICENSE](LICENSE) file for details

---

## Installation Help

Need help installing? Choose your preferred method:

**💻 One-Command (Recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
```

**🍺 macOS with Homebrew:**
```bash
brew tap shafi-/portfolio && brew install portfolio
```

**🔨 Manual Binary Download:**
```bash
# Visit: https://github.com/shafi-/portfolio/releases/latest
```

**📖 Full Guide:** See [Quick Start Guide](docs/QUICK_START.md)

---

**Repository:** [https://github.com/shafi-/portfolio](https://github.com/shafi-/portfolio)  
**Issues:** [https://github.com/shafi-/portfolio/issues](https://github.com/shafi-/portfolio/issues)  
**Discussions:** [https://github.com/shafi-/portfolio/discussions](https://github.com/shafi-/portfolio/discussions)
