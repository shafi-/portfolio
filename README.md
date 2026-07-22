# Portfolio Engine

Portfolio is a local-first project inventory and knowledge platform that enables developers and AI coding agents to understand an entire software portfolio.

## Prerequisites

- Go 1.21 or higher
- Git (for project discovery)

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/nerddevsltd/portfolio.git
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

# Check system status
portfolio status

# Run diagnostics
portfolio doctor
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

- [Knowledge Model](docs/KnowledgeModel.md)
- [Platform Specification](docs/PlatformSpecification.md)
- [Product Requirements](docs/PRD.md)
- [Engineering Guidelines](docs/Guideline.md)

## License

MIT License - see [LICENSE](LICENSE) file for details
