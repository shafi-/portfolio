## 🎉 Portfolio {{ .Tag }} Release

### 🚀 Major Features

**One-Command Installation**
- `curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash`
- Automatic platform detection (macOS/Linux, Intel/ARM)
- Smart error handling and user guidance

**Homebrew Support (macOS)**
- `brew tap shafi-/portfolio && brew install portfolio`
- Official binaries with verification
- Easy updates via `brew upgrade portfolio`

**Multi-Platform Support**
- macOS (Intel/ARM)
- Linux (Intel/ARM)
- Automatic platform detection

### 📦 Installation

```bash
# One-command install (recommended)
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash

# Homebrew (macOS)
brew tap shafi-/portfolio && brew install portfolio

# Manual download
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/
```

### 📚 Documentation
- Quick Start Guide - Get started in 5 minutes
- Homebrew Setup Guide - macOS installation
- Release Process - Team workflow documentation

### 🧪 Quality
- All tests passing
- Build verification successful
- Multi-platform support

---

**Upgrade from previous version:** `portfolio upgrade`

**Full Release Notes:** https://github.com/shafi-/portfolio/releases/{{ .Tag }}
