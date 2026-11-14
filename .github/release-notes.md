## 🎉 Portfolio {{ .Version }}

<!-- Maintainer: only update the Highlights section each release.
     The rest is stable; both placeholders are substituted by CI. -->

### ✨ Highlights

- fix(claude): MCP installer handles entries registered in multiple scopes, registers at user scope, and reports actionable install errors instead of a generic failure
- fix(ci): release workflow builds the embedded dashboard frontend before running Go tests (`//go:embed dist` needs it in CI)
- fix(ci): release notes are generated outside the repo tree (goreleaser dirty-state check) and the version placeholder is substituted correctly

**Full change list:** https://github.com/shafi-/portfolio/compare/{{ .PrevVersion }}...{{ .Version }}

### 📦 Installation

```bash
# One-command install (recommended)
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash

# Manual download
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/
```

**Note:** Homebrew installation has been discontinued. Please use the curl script for installation.

### 🧪 Quality

- `go test -race ./...` passing on CI
- `go vet` clean
- Multi-platform builds verified by GoReleaser (darwin, linux, windows)

---

**Upgrade from previous version:** `portfolio upgrade`

**Issues and feedback:** https://github.com/shafi-/portfolio/issues
