# Nopher

**Nostr to Gopher/Gemini/Finger Gateway**

Nopher is a personal gateway that serves your Nostr content via legacy internet protocols: Gopher (RFC 1436), Gemini, and Finger (RFC 742).

## Overview

- **Single-tenant** by default - shows one operator's notes and articles from Nostr
- **Config-first** - everything configurable via file and env overrides
- **Protocol servers** - Gopher, Gemini, and Finger simultaneously
- **Inbox/Outbox model** - aggregates replies, reactions, and zaps from Nostr
- **Smart relay discovery** - uses NIP-65 (kind 10002) for dynamic relay hints
- **Controlled sync scope** - sync self/following/mutual/FOAF with caps and deny lists
- **Embedded storage** - uses Khatru relay with SQLite or LMDB backend
- **Protocol-specific rendering** - gopher menus, gemini gemtext, finger responses

## Status

🚧 **Phase 0 Complete - Bootstrap Phase**

The project structure is in place and ready for implementation:
- ✅ Go module initialized
- ✅ Directory structure created
- ✅ Build/test/lint scripts ready
- ✅ CI/CD pipelines configured
- ✅ Docker support ready
- ✅ Example configuration available

Next: Phase 1 (Configuration System)

See `memory/PHASES.md` for the complete implementation roadmap.

## Quick Start

### Installation

#### Download Binary

```bash
# Coming soon - releases not yet available
# curl -fsSL https://get.nopher.io | sh
```

#### Build from Source

```bash
# Clone repository
git clone https://github.com/sandwich/nopher.git
cd nopher

# Build
make build

# Run
./dist/nopher --version
```

### Configuration

```bash
# Copy example config
cp configs/nopher.example.yaml nopher.yaml

# Edit with your npub and seed relays
vim nopher.yaml

# Set your nsec (NEVER in config file!)
export NOPHER_NSEC="nsec1..."

# Run
./dist/nopher --config nopher.yaml
```

### Docker

```bash
# Build image
make docker

# Or use docker-compose
docker-compose up -d
```

## Development

### Prerequisites

- Go 1.23 or later
- Make
- golangci-lint (for linting)

### Local Development

```bash
# Run all checks
make check

# Run tests
make test

# Run linters
make lint

# Build binary
make build

# Run locally
make dev
```

### Project Structure

```
nopher/
├── cmd/nopher/          # Main application
├── internal/            # Private application code
├── pkg/                 # Public libraries
├── configs/             # Example configurations
├── scripts/             # Build and CI scripts
├── memory/              # Design documentation
├── docs/                # User documentation
└── test/                # Integration tests
```

## Architecture

Nopher follows a config-first philosophy with clear separation of concerns:

- **Storage Layer** - Khatru relay with SQLite/LMDB
- **Sync Engine** - Discovers and syncs from Nostr relays
- **Protocol Servers** - Gopher (port 70), Gemini (port 1965), Finger (port 79)
- **Rendering** - Protocol-specific content transformation
- **Caching** - In-memory or Redis for performance

For detailed architecture, see `memory/architecture.md`.

## Documentation

- `memory/PHASES.md` - Implementation phases and roadmap
- `memory/configuration.md` - Configuration reference
- `memory/architecture.md` - System architecture
- `AGENTS.md` - Guidelines for AI agents and contributors

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

For AI agents working on this project, please read [AGENTS.md](AGENTS.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Khatru](https://github.com/fiatjaf/khatru) - Go relay framework
- [go-nostr](https://github.com/nbd-wtf/go-nostr) - Nostr protocol implementation

## Support

- Issues: https://github.com/sandwich/nopher/issues
- Discussions: https://github.com/sandwich/nopher/discussions

---

Built with ❤️ for the Nostr and legacy internet communities.
