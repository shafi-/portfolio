#!/bin/bash
set -e

# Portfolio One-Click Release Script
# This script automates the entire release process

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERSION_FILE="$PROJECT_ROOT/internal/version/version.go"

echo -e "${BLUE}🚀 Portfolio Release Automation${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Function to display usage
usage() {
    echo "Usage: $0 <version> [options]"
    echo ""
    echo "Arguments:"
    echo "  version          Version to release (e.g., v0.2.0)"
    echo ""
    echo "Options:"
    echo "  --skip-tests     Skip test execution"
    echo "  --dry-run        Show what would be done without executing"
    echo "  --push-only      Only push existing release branch"
    echo ""
    echo "Example:"
    echo "  $0 v0.2.0"
    exit 1
}

# Function to extract current version
get_current_version() {
    grep -o 'version = "[^"]*"' "$VERSION_FILE" | sed 's/version = "\(.*\)"/\1/'
}

# Function to update version in files
update_version() {
    local new_version="$1"

    echo -e "${YELLOW}📝 Updating version to $new_version...${NC}"

    if [ "$DRY_RUN" = true ]; then
        echo "Would update version in $VERSION_FILE"
        echo "Would update version in USER_MANUAL.md"
        return
    fi

    # Update version.go
    sed -i.bak "s/version = \".*\"/version = \"$new_version\"/" "$VERSION_FILE"
    rm -f "$VERSION_FILE.bak"

    # Update USER_MANUAL.md version header
    sed -i.bak "s/Version: [0-9.]*/Version: $new_version/" "$PROJECT_ROOT/USER_MANUAL.md"
    rm -f "$PROJECT_ROOT/USER_MANUAL.md.bak"

    echo -e "${GREEN}✅ Version updated${NC}"
}

# Function to run tests
run_tests() {
    echo -e "${YELLOW}🧪 Running tests...${NC}"

    cd "$PROJECT_ROOT"

    if [ "$DRY_RUN" = true ]; then
        echo "Would run: go test -race ./..."
        echo "Would run: go vet ./..."
        return
    fi

    go test -race ./... || {
        echo -e "${RED}❌ Tests failed${NC}"
        exit 1
    }

    go vet ./... || {
        echo -e "${RED}❌ Vet failed${NC}"
        exit 1
    }

    echo -e "${GREEN}✅ All tests passing${NC}"
}

# Function to build binary
build_binary() {
    echo -e "${YELLOW}🔨 Building release binary...${NC}"

    cd "$PROJECT_ROOT"

    if [ "$DRY_RUN" = true ]; then
        echo "Would run: go build -o portfolio ./cmd/portfolio"
        return
    fi

    go build -o portfolio ./cmd/portfolio || {
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
    }

    # Verify version output
    VERSION_OUTPUT="./portfolio --version"
    if ! $VERSION_OUTPUT | grep -q "$VERSION"; then
        echo -e "${RED}❌ Version mismatch${NC}"
        echo "Expected: $VERSION"
        echo "Got: $($VERSION_OUTPUT)"
        rm -f portfolio
        exit 1
    fi

    rm -f portfolio
    echo -e "${GREEN}✅ Build successful${NC}"
}

# Function to create release branch
create_release_branch() {
    local version="$1"
    local branch_name="release-$version"

    echo -e "${YELLOW}🌿 Creating release branch: $branch_name${NC}"

    cd "$PROJECT_ROOT"

    # Check if branch already exists
    if git show-ref --verify --quiet refs/heads/"$branch_name"; then
        echo -e "${YELLOW}⚠️  Branch $branch_name already exists${NC}"
        read -p "Use existing branch? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}❌ Aborted${NC}"
            exit 1
        fi
        git checkout "$branch_name"
    else
        if [ "$DRY_RUN" = true ]; then
            echo "Would create branch: $branch_name"
        else
            git checkout -b "$branch_name" || {
                echo -e "${RED}❌ Failed to create branch${NC}"
                exit 1
            }
        fi
    fi

    echo -e "${GREEN}✅ Release branch ready${NC}"
}

# Function to commit changes
commit_changes() {
    local version="$1"

    echo -e "${YELLOW}📦 Committing release changes...${NC}"

    cd "$PROJECT_ROOT"

    if [ "$DRY_RUN" = true ]; then
        echo "Would commit: release: prepare $version"
        return
    fi

    # Check if there are changes to commit
    if git diff --quiet && git diff --cached --quiet; then
        echo -e "${YELLOW}⚠️  No changes to commit${NC}"
        return
    fi

    git add internal/version/version.go USER_MANUAL.md
    git commit -m "release: prepare $version" || {
        echo -e "${RED}❌ Commit failed${NC}"
        exit 1
    }

    echo -e "${GREEN}✅ Changes committed${NC}"
}

# Function to push release branch
push_release_branch() {
    local version="$1"
    local branch_name="release-$version"

    echo -e "${YELLOW}📤 Pushing release branch...${NC}"

    cd "$PROJECT_ROOT"

    if [ "$DRY_RUN" = true ]; then
        echo "Would push branch: $branch_name"
        return
    fi

    git push origin "$branch_name" || {
        echo -e "${RED}❌ Push failed${NC}"
        exit 1
    }

    echo -e "${GREEN}✅ Release branch pushed${NC}"
}

# Function to create GitHub PR
create_pr() {
    local version="$1"
    local branch_name="release-$version"

    echo -e "${YELLOW}🎯 Creating GitHub PR...${NC}"

    if [ "$DRY_RUN" = true ]; then
        echo "Would create PR: $branch_name → $branch_name"
        return
    fi

    # Check if gh CLI is available
    if ! command -v gh &> /dev/null; then
        echo -e "${YELLOW}⚠️  GitHub CLI not found${NC}"
        echo "Please create PR manually at:"
        echo "https://github.com/shafi-/portfolio/compare/$branch_name"
        return
    fi

    # Create PR using gh CLI
    gh pr create \
        --base "$branch_name" \
        --head "$branch_name" \
        --title "Release $version" \
        --body "Auto-generated release PR for $version

## Changes
- Version bump to $version
- Release preparation
- Automated testing and build verification

## Checklist
- [ ] All tests passing
- [ ] Build verification successful
- [ ] Documentation updated
- [ ] Ready for release

Merging this PR will trigger the automated release workflow." || {
        echo -e "${YELLOW}⚠️  PR creation failed${NC}"
        echo "Please create PR manually at:"
        echo "https://github.com/shafi-/portfolio/compare/$branch_name"
        return
    }

    echo -e "${GREEN}✅ PR created successfully${NC}"
}

# Function to show release summary
show_summary() {
    local version="$1"
    local branch_name="release-$version"

    echo ""
    echo -e "${BLUE}📋 Release Summary${NC}"
    echo -e "${BLUE}=================${NC}"
    echo ""
    echo "Version: $version"
    echo "Branch: $branch_name"
    echo "Repository: shafi-/portfolio"
    echo ""
    echo "Next steps:"
    echo "1. Review PR: https://github.com/shafi-/portfolio/compare/$branch_name"
    echo "2. Merge PR to trigger automated release"
    echo "3. Verify release assets on GitHub"
    echo "4. Test installation: curl -fsSL https://raw.githubusercontent.com/shafi-/portfolio/main/install.sh | bash"
    echo ""
    echo -e "${GREEN}🎉 Release preparation complete!${NC}"
}

# Parse arguments
VERSION=""
SKIP_TESTS=false
DRY_RUN=false
PUSH_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --push-only)
            PUSH_ONLY=true
            shift
            ;;
        -*)
            echo "Unknown option: $1"
            usage
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

# Validate version
if [ -z "$VERSION" ]; then
    echo -e "${RED}❌ Version argument required${NC}"
    usage
fi

if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ Invalid version format: $VERSION${NC}"
    echo "Version must be in format: v1.2.3"
    exit 1
fi

# Show dry run mode
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}🔍 DRY RUN MODE - No actual changes will be made${NC}"
    echo ""
fi

# Main release workflow
main() {
    local current_version
    current_version=$(get_current_version)

    echo -e "${BLUE}Current version: $current_version${NC}"
    echo -e "${BLUE}Release version: $VERSION${NC}"
    echo ""

    if [ "$PUSH_ONLY" = false ]; then
        # Ensure we're on main branch
        echo -e "${YELLOW}📍 Checking branch...${NC}"
        cd "$PROJECT_ROOT"

        if [ "$(git branch --show-current)" != "main" ]; then
            echo -e "${YELLOW}⚠️  Not on main branch${NC}"
            read -p "Switch to main branch? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                git checkout main || {
                    echo -e "${RED}❌ Failed to checkout main${NC}"
                    exit 1
                }
            else
                echo -e "${RED}❌ Aborted${NC}"
                exit 1
            fi
        fi

        # Pull latest changes
        echo -e "${YELLOW}📥 Pulling latest changes...${NC}"
        if [ "$DRY_RUN" = false ]; then
            git pull origin main || {
                echo -e "${RED}❌ Failed to pull latest changes${NC}"
                exit 1
            }
        fi

        # Update version
        update_version "$VERSION"

        # Run tests
        if [ "$SKIP_TESTS" = false ]; then
            run_tests
        else
            echo -e "${YELLOW}⚠️  Skipping tests${NC}"
        fi

        # Build verification
        build_binary

        # Create release branch
        create_release_branch "$VERSION"

        # Commit changes
        commit_changes "$VERSION"
    fi

    # Push release branch
    push_release_branch "$VERSION"

    # Create PR
    if [ "$PUSH_ONLY" = false ]; then
        create_pr "$VERSION"
    fi

    # Show summary
    show_summary "$VERSION"
}

# Run main workflow
main