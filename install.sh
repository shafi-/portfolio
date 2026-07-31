#!/bin/bash
set -e

# Portfolio One-Click Installer
# This script downloads and installs the latest Portfolio release

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Installing Portfolio...${NC}"

# Check for existing Homebrew installation
if command -v brew >/dev/null 2>&1 && brew list portfolio 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Warning: Portfolio is already installed via Homebrew${NC}"
    echo -e "${YELLOW}   This may cause version conflicts. Consider removing:${NC}"
    echo -e "${YELLOW}   brew uninstall portfolio && brew untap shafi-/portfolio${NC}"
    echo ""
    read -p "Continue with installation? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Installation cancelled${NC}"
        exit 1
    fi
    echo ""
fi

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Darwin)
        BINARY_NAME="portfolio-darwin"
        ;;
    Linux)
        BINARY_NAME="portfolio-linux"
        ;;
    *)
        echo -e "${RED}Error: Unsupported OS: $OS${NC}"
        echo "Portfolio currently supports macOS and Linux"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        BINARY_NAME="${BINARY_NAME}-amd64"
        ;;
    arm64|aarch64)
        BINARY_NAME="${BINARY_NAME}-arm64"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Download latest release
RELEASE_URL="https://github.com/shafi-/portfolio/releases/latest/download/${BINARY_NAME}"
echo -e "${YELLOW}📥 Downloading Portfolio from: $RELEASE_URL${NC}"

curl -fsSL "$RELEASE_URL" -o /tmp/portfolio || {
    echo -e "${RED}Error: Failed to download Portfolio${NC}"
    echo "Please check your internet connection and try again"
    exit 1
}

# Make executable
chmod +x /tmp/portfolio

# Install to /usr/local/bin
echo -e "${YELLOW}📦 Installing to /usr/local/bin/portfolio${NC}"
sudo mv /tmp/portfolio /usr/local/bin/portfolio || {
    echo -e "${RED}Error: Failed to install to /usr/local/bin${NC}"
    echo "You may need to run this script with sudo privileges"
    exit 1
}

# Verify installation
echo -e "${YELLOW}✅ Verifying installation...${NC}"
if portfolio --version &>/dev/null; then
    VERSION=$(portfolio --version)
    echo -e "${GREEN}✨ Portfolio installed successfully!${NC}"
    echo -e "${GREEN}Version: $VERSION${NC}"
    echo ""
    echo -e "${YELLOW}🎯 Quick Start:${NC}"
    echo "  1. Initialize Portfolio:  ${GREEN}portfolio init${NC}"
    echo "  2. Discover projects:    ${GREEN}portfolio discover${NC}"
    echo "  3. Check status:         ${GREEN}portfolio status${NC}"
    echo ""
    echo -e "${YELLOW}📖 For more information:${NC}"
    echo "  Documentation: https://github.com/shafi-/portfolio"
    echo "  User Manual:   ${GREEN}portfolio manual${NC}"
else
    echo -e "${RED}Error: Installation verification failed${NC}"
    exit 1
fi
