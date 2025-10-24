# Architecture Overview

**Status:** Technical deep-dive for developers and contributors

Complete architectural overview of Nopher's design, components, and implementation.

## Executive Summary

Nopher is a **personal Nostr gateway** that serves content via legacy internet protocols (Gopher, Gemini, Finger). It acts as a bridge between the modern Nostr protocol and classic protocols from the 1980s-90s.

**Key design principles:**
1. **Config-first** - Everything customizable via YAML
2. **Single-tenant** - Optimized for one operator
3. **Embedded storage** - Uses Khatru as library, not separate service
4. **Protocol agnostic** - Serve same content via multiple protocols
5. **Pull model** - Sync from remote relays, serve locally

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Nostr Network                            │
│  (Remote relays: wss://relay.damus.io, etc.)               │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ WebSocket subscriptions
                          │ (filtered by scope/kinds)
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                   Sync Engine                               │
│  - Relay discovery (NIP-65)                                 │
│  - Social graph (follows/FOAF)                              │
│  - Cursor management                                        │
│  - Event ingestion                                          │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ StoreEvent()
                          ↓
┌─────────────────────────────────────────────────────────────┐
│              Khatru (Embedded Nostr Relay)                  │
│  - Event storage & indexing                                 │
│  - Signature verification                                   │
│  - Replaceable event handling                               │
│  - NIP compliance (01, 10, 33, etc.)                        │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ eventstore interface
                          ↓
┌─────────────────────────────────────────────────────────────┐
│           Database Backend (SQLite or LMDB)                 │
│  - Events (managed by Khatru)                               │
│  - Custom tables:                                           │
│    • relay_hints                                            │
│    • graph_nodes                                            │
│    • sync_state                                             │
│    • aggregates                                             │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          │ QueryEvents()
                          ↓
┌─────────────────────────────────────────────────────────────┐
│              Content Query & Aggregation                    │
│  - Section filters                                          │
│  - Thread resolution                                        │
│  - Interaction rollups                                      │
│  - Caching layer                                            │
└───────────┬─────────────┬──────────────┬────────────────────┘
            │             │              │
            ↓             ↓              ↓
┌─────────────┐  ┌──────────────┐  ┌─────────────┐
│   Gopher    │  │    Gemini    │  │   Finger    │
│  Renderer   │  │   Renderer   │  │  Renderer   │
│             │  │              │  │             │
│ Markdown→   │  │ Markdown→    │  │ Markdown→   │
│ Plain Text  │  │ Gemtext      │  │ Compact     │
└──────┬──────┘  └──────┬───────┘  └──────┬──────┘
       │                │                 │
       ↓                ↓                 ↓
┌──────────────────────────────────────────────────────────────┐
│                   Protocol Servers                           │
│  - Gopher: port 70 (RFC 1436)                                │
│  - Gemini: port 1965 (TLS/TOFU)                              │
│  - Finger: port 79 (RFC 742)                                 │
└──────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### 1. Configuration System

**Location:** `internal/config/`

**Purpose:** Load, validate, and apply configuration.

**Key files:**
- `config.go` - Struct definitions, loader
- `config_test.go` - Validation tests
- `example.yaml` - Embedded example config

**Features:**
- YAML parsing with validation
- Environment variable overrides (`NOPHER_*`)
- Defaults for all options
- Secrets via env only (never in files)

**Configuration flow:**
```
config.yaml
    ↓
Load() → Unmarshal YAML
    ↓
applyEnvOverrides() → Apply NOPHER_* env vars
    ↓
Validate() → Check required fields, formats
    ↓
*Config → Pass to components
```

**Status:** ✅ Phase 1 complete

---

### 2. Storage Layer

**Location:** `internal/storage/`

**Purpose:** Persist events and Nopher-specific data.

**Architecture:**
```
┌─────────────────────────┐
│  Storage Interface      │
│  (storage.go)           │
└───────┬─────────────────┘
        │
    ┌───┴───┐
    ↓       ↓
┌──────┐ ┌──────┐
│SQLite│ │ LMDB │
└──────┘ └──────┘
    ↓       ↓
┌──────────────┐
│   Khatru     │
│ (eventstore) │
└──────────────┘
```

**Key files:**
- `storage.go` - Interface, factory
- `sqlite.go` - SQLite implementation
- `lmdb.go` - LMDB implementation
- `migrations.go` - Schema creation
- `relay_hints.go`, `graph_nodes.go`, `sync_state.go`, `aggregates.go` - Custom tables

**What Khatru handles:**
- Event storage (events table)
- Querying (Nostr filters)
- Signature verification
- Replaceable events logic

**What Nopher adds:**
- relay_hints (NIP-65 data)
- graph_nodes (social graph cache)
- sync_state (cursors per relay/kind)
- aggregates (interaction rollups)

**Status:** ✅ Phase 2 complete

---

### 3. Nostr Client

**Location:** `internal/nostr/`

**Purpose:** Connect to remote Nostr relays.

**Key files:**
- `client.go` - WebSocket client pool
- `relay.go` - Per-relay connection management
- `discovery.go` - NIP-65 relay discovery
- `relay_hints.go` - Relay hint parsing

**Connection pool:**
```
┌───────────────────────────┐
│   Relay Connection Pool   │
│                           │
│  ┌─────┐  ┌─────┐        │
│  │Relay│  │Relay│  ...   │
│  │  1  │  │  2  │        │
│  └─────┘  └─────┘        │
└───────────────────────────┘
```

**Per-relay features:**
- WebSocket connection
- Subscription management
- Backoff/retry logic
- Health tracking

**Status:** ✅ Phase 3 complete

---

### 4. Sync Engine

**Location:** `internal/sync/`

**Purpose:** Pull events from remote relays, store locally.

**Key files:**
- `engine.go` - Main sync orchestration
- `filters.go` - Build Nostr filters from scope
- `graph.go` - Social graph computation
- `cursors.go` - Cursor tracking
- `scope.go` - Scope enforcement (self/following/mutual/foaf)

**Sync flow:**
```
1. Discovery
   └→ Fetch kind 10002 from seeds
   └→ Parse relay hints
   └→ Build relay pool

2. Graph Computation
   └→ Fetch kind 3 (contacts)
   └→ Compute depth, mutual
   └→ Populate graph_nodes

3. Filter Building
   └→ Per-kind filters
   └→ Per-author filters (from graph)
   └→ Apply scope modifiers

4. Subscription
   └→ Subscribe to each relay with filters
   └→ Use cursors (since timestamps)

5. Ingestion
   └→ Receive events via WebSocket
   └→ Validate & store (Khatru)
   └→ Update cursors
   └→ Trigger aggregates
```

**Status:** ✅ Phase 4 complete (integrated in main.go)

---

### 5. Aggregates

**Location:** `internal/aggregates/`

**Purpose:** Count interactions (replies, reactions, zaps).

**Key files:**
- `aggregates.go` - Main aggregates manager
- `threading.go` - NIP-10 thread resolution
- `reactions.go` - Reaction counting
- `zaps.go` - Zap parsing and sum
- `reconciler.go` - Periodic recount
- `queries.go` - Helper queries

**Aggregate computation:**
```
Event (kind 1)
    ↓
Find referencing events:
  - kind 1 with #e → replies
  - kind 7 with #e → reactions
  - kind 9735 with #e → zaps
    ↓
Count and sum
    ↓
Store in aggregates table
```

**Update strategies:**
1. **On ingestion** - Update immediately when new interaction arrives
2. **Reconciler** - Periodic full recount (detect drift)

**Status:** ✅ Phase 5 complete (tests passing)

---

### 6. Markdown Conversion

**Location:** `internal/markdown/`

**Purpose:** Convert Nostr content (often markdown) to protocol-specific formats.

**Key files:**
- `parser.go` - Markdown AST parsing
- `gopher.go` - Markdown → plain text
- `gemini.go` - Markdown → gemtext
- `finger.go` - Markdown → stripped compact

**Conversion pipeline:**
```
Markdown content
    ↓
Parse to AST
    ↓
┌───────┬────────┬─────────┐
│       │        │         │
↓       ↓        ↓         ↓
Gopher  Gemini  Finger  (other)
(plain) (gemtext)(strip)
```

**Gopher transformations:**
- Headings → UPPERCASE or underline
- Bold → UPPERCASE or **keep**
- Links → `text <url>`
- Code → indent or separators
- Wrap at 70 chars

**Gemini transformations:**
- Headings → `# ## ###`
- Links → `=> url text` (separate lines)
- Lists → `* item`
- Code → `` ``` ... ``` ``
- Wrap at 80 chars (optional)

**Finger transformations:**
- Strip all markdown syntax
- Preserve bare URLs optionally
- Truncate to ~500 chars

**Status:** ✅ Phase 6 complete (tests present)

---

### 7. Protocol Servers

**Location:** `internal/gopher/`, `internal/gemini/`, `internal/finger/`

#### Gopher Server

**Files:**
- `server.go` - TCP listener, connection handler
- `router.go` - Selector routing
- `gophermap.go` - Menu generation
- `renderer.go` - Event rendering

**Request flow:**
```
Client connects to port 70
    ↓
Send selector (e.g., "/notes")
    ↓
Router matches selector
    ↓
Query events from storage
    ↓
Render gophermap or text
    ↓
Send response
    ↓
Close connection
```

**Status:** 🟡 Phase 7 implemented

#### Gemini Server

**Files:**
- `server.go` - TLS listener, connection handler
- `router.go` - URL routing
- `renderer.go` - Gemtext rendering
- `protocol.go` - Gemini protocol helpers
- `tls.go` - TLS cert management

**Request flow:**
```
Client connects to port 1965 (TLS)
    ↓
Send URL (e.g., "gemini://host/notes")
    ↓
Router matches path
    ↓
Query events from storage
    ↓
Render gemtext
    ↓
Send response with status code
    ↓
Close connection
```

**Status:** 🟡 Phase 8 implemented

#### Finger Server

**Files:**
- `server.go` - TCP listener, connection handler
- `handler.go` - Query parsing, user lookup
- `renderer.go` - Finger response formatting

**Request flow:**
```
Client connects to port 79
    ↓
Send query (e.g., "npub1abc@host")
    ↓
Parse username
    ↓
Query profile (kind 0) + recent notes
    ↓
Format finger response
    ↓
Send response
    ↓
Close connection
```

**Status:** 🟡 Phase 9 implemented

---

### 8. Caching Layer (Planned)

**Location:** `internal/cache/` (not yet created)

**Purpose:** Cache rendered responses, reduce database queries.

**Planned architecture:**
```
┌─────────────────────────┐
│   Cache Interface       │
│   (cache.go)            │
└───────┬─────────────────┘
        │
    ┌───┴───┐
    ↓       ↓
┌──────┐ ┌──────┐
│Memory│ │Redis │
└──────┘ └──────┘
```

**Cache keys:**
- Protocol + path + timestamp range → rendered response
- Event ID → rendered event
- Section query → event list

**TTL strategy:**
- Short (10-60s): Live content (inbox, interactions)
- Medium (300-600s): Sections, menus
- Long (hours/days): Immutable (old events, profiles)

**Invalidation:**
- On new event ingestion (if affects cached content)
- On aggregate updates
- Manual invalidation

**Status:** 📋 Phase 10 planned (no code yet)

---

## Code Organization

```
nopher/
├── cmd/nopher/              # Main application entry point
│   └── main.go              # CLI, server startup
│
├── internal/                # Private application code
│   ├── config/              # Configuration
│   │   ├── config.go
│   │   ├── config_test.go
│   │   └── example.yaml
│   │
│   ├── storage/             # Storage layer
│   │   ├── storage.go       # Interface
│   │   ├── sqlite.go        # SQLite backend
│   │   ├── lmdb.go          # LMDB backend
│   │   ├── migrations.go    # Schema
│   │   ├── relay_hints.go   # Custom table
│   │   ├── graph_nodes.go   # Custom table
│   │   ├── sync_state.go    # Custom table
│   │   └── aggregates.go    # Custom table
│   │
│   ├── nostr/               # Nostr client
│   │   ├── client.go        # WebSocket pool
│   │   ├── relay.go         # Per-relay connection
│   │   ├── discovery.go     # NIP-65
│   │   └── relay_hints.go   # Hint parsing
│   │
│   ├── sync/                # Sync engine
│   │   ├── engine.go        # Main orchestration
│   │   ├── filters.go       # Filter builder
│   │   ├── graph.go         # Social graph
│   │   ├── cursors.go       # Cursor tracking
│   │   └── scope.go         # Scope enforcement
│   │
│   ├── aggregates/          # Aggregates
│   │   ├── aggregates.go    # Manager
│   │   ├── threading.go     # NIP-10
│   │   ├── reactions.go     # Reactions
│   │   ├── zaps.go          # Zaps
│   │   ├── reconciler.go    # Periodic recount
│   │   └── queries.go       # Helpers
│   │
│   ├── markdown/            # Markdown conversion
│   │   ├── parser.go        # AST parsing
│   │   ├── gopher.go        # → plain text
│   │   ├── gemini.go        # → gemtext
│   │   └── finger.go        # → stripped
│   │
│   ├── gopher/              # Gopher protocol
│   │   ├── server.go        # TCP server
│   │   ├── router.go        # Selector routing
│   │   ├── gophermap.go     # Menu generation
│   │   └── renderer.go      # Event rendering
│   │
│   ├── gemini/              # Gemini protocol
│   │   ├── server.go        # TLS server
│   │   ├── router.go        # URL routing
│   │   ├── renderer.go      # Gemtext rendering
│   │   ├── protocol.go      # Protocol helpers
│   │   └── tls.go           # TLS management
│   │
│   └── finger/              # Finger protocol
│       ├── server.go        # TCP server
│       ├── handler.go       # Query parsing
│       └── renderer.go      # Response formatting
│
├── configs/                 # Example configurations
│   └── nopher.example.yaml
│
├── memory/                  # Design documentation
│   ├── README.md
│   ├── PHASES.md
│   ├── architecture.md
│   └── ...
│
├── docs/                    # User documentation
│   ├── getting-started.md
│   ├── configuration.md
│   ├── storage.md
│   ├── protocols.md
│   ├── nostr-integration.md
│   ├── architecture.md      # (this file)
│   ├── deployment.md
│   └── troubleshooting.md
│
├── scripts/                 # Build and CI scripts
│   ├── build.sh
│   ├── test.sh
│   └── lint.sh
│
├── Makefile                 # Build automation
├── go.mod                   # Go dependencies
├── README.md                # Project overview
├── CONTRIBUTING.md          # Contribution guidelines
└── AGENTS.md                # Agent/contributor instructions
```

---

## Technology Stack

### Language

**Go 1.23+**

**Why Go:**
- Excellent concurrency (goroutines for multiple protocols)
- Strong networking libraries
- Cross-platform compilation
- Great performance
- Mature ecosystem

### Core Dependencies

**Khatru** - Nostr relay framework
- https://github.com/fiatjaf/khatru
- Provides event storage, validation, querying
- Pluggable eventstore backends

**eventstore** - Database adapters
- https://github.com/fiatjaf/eventstore
- SQLite and LMDB implementations
- Used by Khatru

**go-nostr** - Nostr protocol
- https://github.com/nbd-wtf/go-nostr
- WebSocket client, event signing/verification
- NIP implementations

**gopkg.in/yaml.v3** - YAML parsing
- Configuration file parsing

### Why Khatru?

**Benefits:**
1. Battle-tested event storage (production Nostr relays)
2. NIP compliance out-of-box
3. Automatic signature verification
4. Pluggable backends (SQLite/LMDB)
5. Simpler codebase (no event plumbing)
6. Future-proof (protocol evolution)

**What Khatru handles:**
- Event storage, indexing, deduplication
- Replaceable event logic
- Event validation
- Querying via Nostr filters
- WebSocket relay interface (optional)

**What Nopher adds:**
- Sync engine (pull from remote → push to Khatru)
- Relay discovery and social graph
- Aggregates (interaction counts)
- Protocol servers (Gopher/Gemini/Finger)
- Markdown conversion and rendering

---

## Data Flow Examples

### Example 1: Syncing a New Note

```
1. User publishes note (kind 1) on remote relay
   │
2. Nopher sync engine subscribes to relay
   │
3. Relay sends EVENT message
   │
4. Sync engine receives event
   │
5. Sync engine calls khatru.StoreEvent(event)
   │
6. Khatru validates signature
   │
7. Khatru checks for duplicates
   │
8. Khatru stores event in SQLite/LMDB
   │
9. Sync engine updates sync_state cursor
   │
10. If event is reply/reaction/zap:
    └→ Aggregates manager updates aggregates table
   │
11. Event now queryable via protocols
```

### Example 2: Serving a Gopher Menu

```
1. Client connects to port 70
   │
2. Client sends selector: "/notes"
   │
3. Gopher server receives request
   │
4. Router matches "/notes" → notes section
   │
5. Query Khatru: filter={kinds:[1], authors:[owner]}
   │
6. Khatru returns events from eventstore
   │
7. Query aggregates for interaction counts
   │
8. Render gophermap:
   │  - Line per event
   │  - Include metadata (date, reactions, zaps)
   │
9. Send gophermap to client
   │
10. Close connection
```

### Example 3: Rendering a Gemini Article

```
1. Client connects to port 1965 (TLS)
   │
2. Client sends: gemini://host/article/xyz
   │
3. Gemini server receives request
   │
4. Router matches "/article/xyz" → article view
   │
5. Query Khatru: filter={kinds:[30023], ids:[xyz]}
   │
6. Khatru returns article event
   │
7. Parse markdown content
   │
8. Convert markdown → gemtext:
   │  - Headings → # ##
   │  - Links → => url text
   │  - Code → ``` ... ```
   │
9. Query aggregates for replies/reactions
   │
10. Append interaction summary
   │
11. Send gemtext response with status 20
   │
12. Close connection
```

---

## Concurrency Model

Nopher uses Go's goroutines for concurrency:

**Protocol servers:**
- Each protocol server runs in own goroutine
- Each client connection handled in separate goroutine
- Concurrent connections: 1000+ per protocol (lightweight)

**Sync engine:**
- One goroutine per relay subscription
- Separate goroutines for cursor updates, graph computation
- Background reconciler goroutine (aggregate recount)

**Example:**
```go
// Main goroutine
func main() {
    // Start protocol servers (3 goroutines)
    go gopherServer.Start()
    go geminiServer.Start()
    go fingerServer.Start()

    // Start sync engine (N goroutines for N relays)
    go syncEngine.Start()

    // Wait for shutdown signal
    <-sigChan
}

// Per-protocol server
func (s *Server) Start() {
    for {
        conn := listener.Accept()
        go s.handleConnection(conn)  // New goroutine per connection
    }
}
```

---

## Testing Strategy

**Unit tests:**
- Per-package unit tests
- Mock interfaces (storage, relay clients)
- Test coverage target: >80%

**Integration tests:**
- Full flow tests (sync → storage → render)
- Test with real Khatru instance
- Mock remote relays

**Protocol compliance tests:**
- RFC 1436 (Gopher) compliance
- Gemini spec compliance
- RFC 742/1288 (Finger) compliance

**Test files:**
- `*_test.go` - Unit tests alongside source
- `test/integration/` - Integration tests
- `test/compliance/` - Protocol tests

**Current status:**
- Unit tests: 19.4% coverage (growing)
- Integration tests: Planned (Phase 15)
- Compliance tests: Planned (Phase 15)

---

## Security Considerations

### Secrets Management

**nsec (private key):**
- NEVER in config files
- Only via `NOPHER_NSEC` environment variable
- Never logged, never serialized

**Redis URL:**
- Via `NOPHER_REDIS_URL` environment variable
- Keep out of config files if contains password

### Port Binding

**Ports <1024 require root:**
- Gopher: 70
- Finger: 79

**Mitigation:**
- Systemd socket activation (recommended)
- Port forwarding (iptables)
- Run on higher ports (testing only)

### TLS (Gemini)

**Certificate validation:**
- Gemini uses TOFU (Trust On First Use)
- Client stores certificate fingerprint on first connect
- Subsequent connects verify against stored fingerprint

**Self-signed vs. proper certs:**
- Self-signed: OK for personal use
- Let's Encrypt: Recommended for production

### Input Validation

**Selectors/URLs:**
- Validate format, length
- Prevent directory traversal (../)
- Sanitize user input

**Event content:**
- Khatru validates signatures
- Markdown parsing should be safe (no execution)

### Rate Limiting

**Planned (Phase 14):**
- Per-IP rate limits
- Per-protocol limits
- Configurable thresholds

**Current workaround:**
- Firewall rules (iptables, fail2ban)

---

## Performance Characteristics

### Resource Usage

**Memory:**
- Base: ~50MB
- Per protocol server: ~10MB
- Per relay connection: ~1MB
- Per cached response: varies (KB-MB)
- Total typical: 100-200MB

**Disk:**
- SQLite database: ~1KB per event
- 100K events: ~100MB
- LMDB: similar, but pre-allocated

**CPU:**
- Idle: <1%
- Sync active: 5-10% (depends on relay count)
- Serving requests: <1% per connection

### Scalability

**Single-tenant design:**
- Optimized for one owner
- Supports thousands of followed users (with caps)
- Handles millions of events (with LMDB)

**Concurrent connections:**
- Gopher: 1000+ (lightweight, <10KB per connection)
- Gemini: 1000+ (TLS overhead, ~50KB per connection)
- Finger: 1000+ (very lightweight, ~5KB per connection)

**Bottlenecks:**
- Database queries (mitigated by caching)
- Markdown rendering (CPU-bound)
- Aggregate computation (mitigated by caching + reconciler)

---

## Future Enhancements

### Phase 10: Caching

- In-memory cache (default)
- Redis cache (optional)
- Invalidation on new events
- Configurable TTLs

### Phase 11: Sections and Layouts

- Custom section definitions
- Configurable page layouts
- Filter-based sections
- Archive generation

### Phase 12: Operations and Diagnostics

- Structured logging
- Diagnostics page (via protocols)
- Relay health monitoring
- Event count statistics
- Pruning and retention

### Phase 13: Publisher

- Sign events with nsec
- Publish to write relays
- Retry/backoff logic
- Draft management

### Phase 14: Security

- Deny-list enforcement
- Rate limiting
- Input validation
- Privilege separation

### Phase 15: Testing

- >80% unit test coverage
- Integration tests
- Protocol compliance tests

### Phase 16: Distribution

- Optimized builds
- Docker images (multi-arch)
- Systemd service files
- ✅ One-line installer script (completed - `scripts/install.sh`)
- ✅ Enhanced Docker Compose (completed - with Redis, Caddy options)
- ✅ Reverse proxy examples (completed - nginx, Caddy configs)

### Phase 17: Advanced Retention

- Rule-based retention system
- Multi-dimensional conditions (kind, author, social distance, interactions)
- Global caps enforcement (max events, storage, per-kind limits)
- Score-based pruning (when caps exceeded)
- Protected events (never delete)
- Incremental evaluation (on ingestion + periodic)

**Status:** 📋 Planned - Full specification available in [memory/PHASE_17_RETENTION.md](../memory/PHASE_17_RETENTION.md)

---

## Design Decisions

### Why Khatru?

**Alternatives considered:**
- Custom event storage (too much work)
- Direct database access (no NIP compliance)
- Separate relay (unnecessary complexity)

**Khatru wins:**
- Battle-tested, production-ready
- NIP compliance out-of-box
- Pluggable backends
- Active development
- Go ecosystem

### Why SQLite?

**Alternatives considered:**
- PostgreSQL (too heavy for single-tenant)
- LMDB (better for high-volume, but more complex)
- Badger/LevelDB (less mature in Nostr ecosystem)

**SQLite wins for default:**
- Zero configuration
- Single file (easy backups)
- Sufficient for most users
- Mature, stable

**LMDB available for:**
- High-volume use cases
- Better write performance
- Millions of events

### Why Embedded?

**Alternatives considered:**
- Separate Nostr relay (Khatru as service)
- Client-server architecture

**Embedded wins:**
- Simpler deployment (one binary)
- No network overhead
- Direct API access (faster)
- Easier to reason about

### Why Three Protocols?

**Why not just Gopher?**
- Different audiences (Gopher purists, Gemini fans, Finger users)
- Showcase Nostr content in multiple contexts
- Educational (protocol comparison)

**Why not HTTP?**
- Nostr already has HTTP gateways (njump, etc.)
- Focus on underserved protocols
- Minimalist philosophy

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for contributor guidelines.

For AI agents working on this project, see [AGENTS.md](../AGENTS.md).

**Code style:**
- Go standard formatting (`gofmt`)
- Linting with `golangci-lint`
- Keep files <500 lines (see AGENTS.md for guidelines)
- DRY (Don't Repeat Yourself)
- Clear package boundaries

**Pull requests:**
- Write tests for new code
- Update docs if behavior changes
- Follow existing patterns
- Keep PRs focused (one feature/fix per PR)

---

## References

**Design documentation:**
- [memory/architecture.md](../memory/architecture.md) - Original design
- [memory/PHASES.md](../memory/PHASES.md) - Implementation roadmap
- [memory/storage_model.md](../memory/storage_model.md) - Storage design
- [memory/sync_scope.md](../memory/sync_scope.md) - Sync design
- [memory/ui_export.md](../memory/ui_export.md) - Protocol rendering

**External:**
- Khatru: https://github.com/fiatjaf/khatru
- eventstore: https://github.com/fiatjaf/eventstore
- go-nostr: https://github.com/nbd-wtf/go-nostr
- Nostr NIPs: https://github.com/nostr-protocol/nips

---

**Next:** [Deployment Guide](deployment.md) | [Getting Started](getting-started.md) | [Configuration Reference](configuration.md)
