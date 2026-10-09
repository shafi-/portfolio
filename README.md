# Portfolio

> **Local-first project inventory and knowledge platform for developers and AI coding agents**

Portfolio automatically discovers, catalogs, and enables searching across your entire software portfolio—so you never lose track of what you've built.

---

## 🎯 For Users: Just Use Portfolio

**Quick Install (Recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
portfolio init
```

**That's it!** Portfolio is now ready to:
- 📡 Automatically discover all your Git repositories
- 📚 Index documentation across every project
- 🔍 Enable instant search through your entire codebase
- 🤖 Power AI coding agents with 26 MCP tools

**What You Need to Know:**
- ✅ **Zero maintenance** - set it once, it keeps itself updated
- ✅ **100% local** - your code never leaves your machine  
- ✅ **Works offline** - no internet required
- ✅ **AI-ready** - deep integration with Claude Code and other AI agents

**Get Started:**
```bash
portfolio discover          # Find all your projects
portfolio status            # See what's discovered
portfolio install claude    # Add AI integration
```

**Documentation:**
- 📖 [User Manual](USER_MANUAL.md) - Complete reference
- 🚀 [Quick Start Guide](docs/QUICK_START.md) - Up and running in 5 minutes
- 🆘 [Troubleshooting](USER_MANUAL.md#troubleshooting) - Common issues and solutions

---

## 🛠️ For Contributors: Work with Source Code

**Development Prerequisites:**
- Go 1.21+ 
- Git

**Setup:**
```bash
# Clone repository
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# Run tests
go test ./...

# Build development binary
go build ./cmd/portfolio

# Run locally
./portfolio --help
```

**Project Structure:**
```
portfolio/
├── cmd/portfolio/        # CLI entry point
├── internal/            # Core application code
│   ├── config/         # Configuration management
│   ├── database/       # SQLite database layer
│   ├── api/            # HTTP API server
│   ├── mcp/            # MCP server (AI integration)
│   └── cli/            # CLI commands
├── pkg/models/         # Shared data models
└── dashboard/          # Web dashboard (React)
```

**Contributing Guidelines:**
- 📋 [Engineering Principles](docs/Guideline.md) - Development philosophy
- 🏗️ [Platform Specification](docs/PlatformSpecification.md) - Implementation contracts
- 🔧 [Release Process](docs/RELEASE_PROCESS.md) - How to make releases
- 📖 [Architecture Decisions](docs/ADR.md) - Design rationale

**Development Commands:**
```bash
# Run tests with race detection
go test -race ./...

# Run with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint
go vet ./...

# Build release binary
go build -ldflags="-s -w" ./cmd/portfolio
```

---

## 📦 Installation Methods

**Method 1: One-Command Install (Recommended)**
```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
```

**Method 2: Manual Binary Download**
```bash
# Download for your platform
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/
```

**Method 3: Build from Source**
```bash
git clone https://github.com/shafi-/portfolio.git
cd portfolio
go build ./cmd/portfolio
sudo mv portfolio /usr/local/bin/
```

---

## ✨ Key Features

**For Users:**
- 🔍 **Automatic Discovery** - Finds all Git repositories across your projects
- 📊 **Smart Metadata** - Extracts languages, frameworks, dependencies
- 📚 **Documentation Search** - Search across all project docs instantly
- 🤖 **AI Integration** - 26 MCP tools for Claude Code, OpenCode, etc.
- 💾 **Local-First** - All data stays on your machine
- ⚡ **Zero Maintenance** - Set once, runs forever

**For Developers:**
- 🏗️ **Modular Architecture** - Clean separation of concerns
- 🧪 **Comprehensive Tests** - High test coverage with race detection
- 📖 **Well-Documented** - Extensive specs and guides
- 🔧 **Developer Friendly** - Easy to build, test, and contribute
- 🚀 **Performance** - Optimized for large portfolios (100+ projects)

---

## 🎓 Documentation

**📚 Documentation Website:** [https://shafi-.github.io/portfolio/](https://shafi-.github.io/portfolio/) — the full Laravel-style docs site (source in [`website/`](website/))

**User Documentation:**
- [User Manual](USER_MANUAL.md) - Complete reference guide
- [Quick Start Guide](docs/QUICK_START.md) - Get started in 5 minutes
- [Feature List](PUBLIC_FEATURE_LIST.md) - What's included

**Developer Documentation:**
- [Knowledge Model](docs/KnowledgeModel.md) - Core data structures
- [Platform Specification](docs/PlatformSpecification.md) - API contracts
- [Product Requirements](docs/PRD.md) - Vision and roadmap
- [Engineering Guidelines](docs/Guideline.md) - Development principles

**Integration Documentation:**
- [Agent Integration Manual](docs/agent-integration-manual.md) - MCP integration
- [MCP Agent Guide](docs/MCP-AGENT-GUIDE.md) - Deep technical guide

**Release Documentation:**
- [Release Process](docs/RELEASE_PROCESS.md) - Release workflow
- [Release Guide](RELEASE_GUIDE.md) - One-click releases

---

## 🚀 Quick Start

**1. Install Portfolio:**
```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
```

**2. Initialize and Discover:**
```bash
portfolio init          # Set up database
portfolio discover      # Find your projects
```

**3. Check Status:**
```bash
portfolio status        # See what's discovered
portfolio doctor        # Run health check
```

**4. Add AI Integration (Optional):**
```bash
portfolio install claude    # Claude Code integration
portfolio doctor claude     # Verify setup
```

---

## 🤝 Contributing

We welcome contributions! Please see our [Engineering Guidelines](docs/Guideline.md) for development philosophy and [Platform Specification](docs/PlatformSpecification.md) for implementation details.

**Development Workflow:**
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Ensure all tests pass: `go test -race ./...`
5. Submit a pull request

**Areas for Contribution:**
- 🎨 CLI/UX improvements
- 📚 Documentation enhancements  
- 🧪 Test coverage expansion
- 🐛 Bug fixes and optimizations
- 🚀 Performance improvements

---

## 📈 Status

**Current Version:** v0.3.4  
**Implementation Status:** Production Ready  
**Milestones 1-3:** Complete (Core Engine, Agent Integration, Dashboard)  
**Milestone 4:** In Planning (Portfolio Intelligence)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🆘 Support & Community

**🔗 Resources:**
- **Repository:** [https://github.com/shafi-/portfolio](https://github.com/shafi-/portfolio)
- **Issues:** [https://github.com/shafi-/portfolio/issues](https://github.com/shafi-/portfolio/issues)
- **Discussions:** [https://github.com/shafi-/portfolio/discussions](https://github.com/shafi-/portfolio/discussions)

**🚨 Getting Help:**
- Check the [User Manual](USER_MANUAL.md) for common issues
- Run `portfolio doctor` for diagnostics
- Search existing [GitHub Issues](https://github.com/shafi-/portfolio/issues)
- Ask a question in [GitHub Discussions](https://github.com/shafi-/portfolio/discussions)

---

*Portfolio helps developers understand their entire software portfolio while providing AI agents with the context they need to assist effectively.*