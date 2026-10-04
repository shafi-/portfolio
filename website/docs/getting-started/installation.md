---
title: Installation
description: Install Portfolio with the one-command installer, Homebrew, a prebuilt binary, or from source.
---

# Installation

Portfolio ships as a single static binary with no runtime dependencies — CGO is disabled and SQLite is compiled in. Pick whichever method suits you.

## One-command installer

The fastest way to get started:

```bash
curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash
```

The script detects your OS and architecture, downloads the latest release binary, and installs it to your PATH.

## Homebrew

If you use Homebrew on macOS or Linux, you can install from the project tap:

```bash
brew install shafi-/portfolio/portfolio
```

## Manual binary download

Grab a prebuilt binary from [GitHub Releases](https://github.com/shafi-/portfolio/releases) and put it on your PATH. Binaries are published for `darwin`/`linux` on `amd64`/`arm64`:

```bash
# Example: macOS on Apple Silicon
curl -L https://github.com/shafi-/portfolio/releases/latest/download/portfolio-darwin-arm64 -o portfolio
chmod +x portfolio
sudo mv portfolio /usr/local/bin/
```

## Build from source

Requires [Go 1.21+](https://go.dev/dl/):

```bash
git clone https://github.com/shafi-/portfolio.git
cd portfolio
go build -o portfolio ./cmd/portfolio
```

The dashboard is compiled into the binary, so `go build` is all you need — no Node build step required unless you are [developing the dashboard](#building-the-dashboard-from-source).

### Building the dashboard from source

The React dashboard lives in `dashboard/` and its production build is embedded via `go:embed`:

```bash
npm --prefix dashboard install
npm --prefix dashboard run build
go build -o portfolio ./cmd/portfolio
```

## Verify the installation

```bash
portfolio --version
# portfolio version v0.3.4 (commit: …, built: …)
```

If your shell can't find the binary, make sure the install location is on your `PATH` — see [Troubleshooting](/docs/digging-deeper/troubleshooting).
