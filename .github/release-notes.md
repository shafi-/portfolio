## 🎉 Portfolio {{ .Tag }} Release

### 🔒 Security Model Overhaul

**AI Agent Security**
- Comprehensive MCP documentation for AI coding assistants
- Clear guidance: agents must always use MCP tools
- Safe error handling that never exposes internal details
- Database protected from external access via password encryption

**Error Handling**
- New safe error system with user-friendly messages
- No internal paths, stack traces, or implementation details exposed
- Graceful panic recovery with proper logging
- Request ID tracking for debugging

### 🚀 Improvements

**Security Enhancements**
- Database password protection via SQLite PRAGMA key encryption
- Agent-focused MCP documentation prevents misuse
- Comprehensive safe error handling throughout codebase
- Enhanced protection against information disclosure

**Documentation**
- New `skills/portfolio-mcp-interface.md` for AI agents
- Explicit instructions on proper MCP tool usage
- Clear examples and warnings against database direct access
- Complete workflow documentation for agent integration

**Code Quality**
- Refactored error handling across all database operations
- Global panic handler in main.go
- Safe error types with proper message sanitization
- Improved error context and user experience

### 📦 Installation

```bash
# One-command install (recommended)
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash

# Manual download
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/
```

**Note:** Homebrew installation has been discontinued to simplify the release process and eliminate cross-repository complexity. Please use the curl script for installation.

### 📚 Documentation
- Agent Integration Guide - Comprehensive MCP tool documentation
- Security Model - Database protection and safe error handling
- Quick Start Guide - Get started in 5 minutes

### 🧪 Quality
- All security improvements tested
- Safe error handling verified
- Build verification successful
- Multi-platform support maintained

---

**Upgrade from previous version:** `portfolio upgrade`

**Full Release Notes:** https://github.com/shafi-/portfolio/releases/{{ .Tag }}
