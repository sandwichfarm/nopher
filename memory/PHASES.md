nophr Implementation Phases

Overview

This document breaks down nophr implementation into discrete phases. Each phase builds on the previous one and produces a working, testable increment.

Phases are designed to be:
- Independent where possible (can be worked on in parallel)
- Testable (each phase has clear completion criteria)
- Incremental (each adds value to the project)
- Documented (changes to memory/ are tracked)

Phase 0: Project Bootstrap

Goal: Set up project structure, tooling, and CI/CD pipeline.

Deliverables:
- Go module initialized
- Directory structure created
- Makefile with common tasks
- Local build/test/lint scripts
- GitHub Actions workflows (test, lint, build, release, docker)
- GoReleaser configuration
- Dockerfile and docker-compose.yml
- Example configuration files
- Basic README and contributing guide

Completion Criteria:
- `make build` produces a binary
- `make test` runs (even if no tests yet)
- `make lint` passes
- GitHub Actions workflows are functional
- Can create snapshot release locally

Dependencies: None

Files to Create:
- go.mod, go.sum
- cmd/nophr/main.go (minimal)
- Makefile
- scripts/test.sh, scripts/lint.sh, scripts/build.sh
- .github/workflows/*.yml
- .goreleaser.yml
- Dockerfile
- docker-compose.yml
- configs/nophr.example.yaml
- README.md, CONTRIBUTING.md

Memory Updates: None (bootstrap)

Phase 1: Configuration System

Goal: Implement configuration loading with validation and env overrides.

Deliverables:
- Config struct matching memory/configuration.md schema
- YAML parsing with validation
- Environment variable override support (NOPHER_*)
- Config initialization command (nophr init)
- Embedded example config via //go:embed
- Unit tests for config loading

Completion Criteria:
- Can load config from file
- Can override any config value via env var
- Validation catches invalid configs
- `nophr init` generates valid config
- Tests cover all config sections

Dependencies: Phase 0

Files to Create:
- internal/config/config.go
- internal/config/config_test.go
- internal/config/validation.go
- internal/config/env.go
- configs/nophr.example.yaml (detailed)

Memory Updates:
- configuration.md (if schema changes discovered)

Phase 2: Storage Layer (Khatru Integration)

Goal: Set up Khatru with eventstore backend and custom tables.

Deliverables:
- Khatru relay instance initialization
- SQLite eventstore configuration
- LMDB eventstore configuration (optional)
- Custom tables: relay_hints, graph_nodes, sync_state, aggregates
- Database migrations
- Storage interface abstraction
- Unit tests for storage operations

Completion Criteria:
- Can initialize Khatru with SQLite backend
- Can store and query events via Khatru
- Custom tables are created
- Can switch between SQLite and LMDB via config
- Storage tests pass

Dependencies: Phase 1

Files to Create:
- internal/storage/storage.go
- internal/storage/khatru.go
- internal/storage/migrations.go
- internal/storage/relay_hints.go
- internal/storage/graph_nodes.go
- internal/storage/sync_state.go
- internal/storage/aggregates.go
- internal/storage/storage_test.go

Memory Updates:
- storage_model.md (if schema changes needed)

Phase 3: Nostr Client and Relay Discovery

Goal: Connect to Nostr relays and discover relay hints (NIP-65).

Deliverables:
- Nostr relay connection manager
- WebSocket client for remote relays
- Seed relay bootstrap
- NIP-65 (kind 10002) parsing
- Relay hints persistence
- Connection pool management
- Retry/backoff logic
- Unit and integration tests

Completion Criteria:
- Can connect to seed relays
- Can fetch owner's kind 0, 3, 10002
- Parses and stores relay hints
- Handles connection failures gracefully
- Tests verify relay discovery

Dependencies: Phase 2

Files to Create:
- internal/nostr/client.go
- internal/nostr/relay.go
- internal/nostr/discovery.go
- internal/nostr/nip65.go
- internal/nostr/client_test.go

Memory Updates:
- relay_discovery.md (if discovery logic changes)
- sequence_seed_discovery_sync.md (if flow changes)

Phase 4: Sync Engine

Goal: Synchronize events from remote relays to local Khatru instance.

Deliverables:
- Subscription manager for remote relays
- Filter builder based on scope config
- Cursor tracking (sync_state table)
- Event ingestion pipeline to Khatru
- Social graph computation (kind 3 follows)
- FOAF/mutual calculation
- Scope enforcement (caps, allow/deny lists)
- Background sync process
- Tests for sync logic

Completion Criteria:
- Can subscribe to remote relays with filters
- Events are stored in local Khatru instance
- Cursors prevent re-syncing old events
- Social graph is computed correctly
- Scope limits are enforced
- Sync can run continuously

Dependencies: Phase 3

Files to Create:
- internal/sync/engine.go
- internal/sync/subscriptions.go
- internal/sync/filters.go
- internal/sync/graph.go
- internal/sync/scope.go
- internal/sync/cursors.go
- internal/sync/engine_test.go

Memory Updates:
- sync_scope.md (if scope logic changes)
- sequence_seed_discovery_sync.md (if sync flow changes)

Phase 5: Aggregates and Threading

Goal: Compute interaction rollups and thread relationships.

Deliverables:
- NIP-10 thread resolution (root/reply)
- Reaction counting (kind 7)
- Zap parsing and sum (kind 9735)
- Aggregates table computation
- Reconciler for periodic recalculation
- Queries for inbox/outbox views
- Tests for aggregation logic

Completion Criteria:
- Threads are correctly resolved
- Reaction counts are accurate
- Zap amounts are parsed and summed
- Aggregates update on new events
- Reconciler corrects drift
- Inbox/outbox queries work

Dependencies: Phase 4

Files to Create:
- internal/aggregates/aggregates.go
- internal/aggregates/threading.go
- internal/aggregates/reactions.go
- internal/aggregates/zaps.go
- internal/aggregates/reconciler.go
- internal/aggregates/aggregates_test.go

Memory Updates:
- inbox_outbox.md (if aggregation changes)

Phase 6: Markdown Conversion

Goal: Convert markdown to protocol-specific formats.

Deliverables:
- Markdown parser (AST-based)
- Gopher renderer (plain text)
- Gemini renderer (gemtext)
- Finger renderer (compact)
- Configurable rendering options
- Tests with sample markdown

Completion Criteria:
- Markdown parses to AST
- Gopher output is readable plain text
- Gemini output is valid gemtext
- Finger output is compact
- All markdown syntax is handled
- Tests cover edge cases

Dependencies: Phase 1

Files to Create:
- internal/markdown/parser.go
- internal/markdown/renderer.go
- internal/markdown/gopher.go
- internal/markdown/gemini.go
- internal/markdown/finger.go
- internal/markdown/markdown_test.go

Memory Updates:
- markdown_conversion.md (if rendering logic changes)

Phase 7: Gopher Server

Goal: Implement Gopher protocol server (RFC 1436).

Deliverables:
- TCP server on port 70
- Selector routing
- Gophermap generation
- Text file rendering
- Event querying from Khatru
- Thread display with indentation
- Archive navigation
- Diagnostics page
- Tests for Gopher protocol

Completion Criteria:
- Gopher server listens on configured port
- Selectors route correctly
- Gophermaps are valid
- Events render as plain text
- Can navigate threads
- Works with Gopher clients (lynx, VF-1)

Dependencies: Phase 5, Phase 6

Files to Create:
- internal/gopher/server.go
- internal/gopher/router.go
- internal/gopher/gophermap.go
- internal/gopher/renderer.go
- internal/gopher/sections.go
- internal/gopher/server_test.go

Memory Updates:
- ui_export.md (if Gopher rendering changes)

Phase 8: Gemini Server

Goal: Implement Gemini protocol server (gemini://).

Deliverables:
- TLS server on port 1965
- Certificate generation (self-signed)
- URL routing
- Gemtext rendering
- Input handling (status 10)
- Event querying from Khatru
- Thread navigation
- Archive pages
- Diagnostics page
- Tests for Gemini protocol

Completion Criteria:
- Gemini server listens with TLS
- URLs route correctly
- Gemtext is valid
- Input queries work
- Can navigate threads
- Works with Gemini clients (amfora, lagrange)

Dependencies: Phase 5, Phase 6

Files to Create:
- internal/gemini/server.go
- internal/gemini/router.go
- internal/gemini/gemtext.go
- internal/gemini/renderer.go
- internal/gemini/tls.go
- internal/gemini/sections.go
- internal/gemini/server_test.go

Memory Updates:
- ui_export.md (if Gemini rendering changes)

Phase 9: Finger Server

Goal: Implement Finger protocol server (RFC 742).

Deliverables:
- TCP server on port 79
- Query parsing (user@host)
- Profile formatting from kind 0
- Recent notes display
- .plan field from about
- Interaction counts
- Tests for Finger protocol

Completion Criteria:
- Finger server listens on configured port
- Queries parse correctly
- Profile info displays
- Recent notes shown
- Works with finger clients

Dependencies: Phase 5, Phase 6

Files to Create:
- internal/finger/server.go
- internal/finger/query.go
- internal/finger/renderer.go
- internal/finger/server_test.go

Memory Updates:
- ui_export.md (if Finger rendering changes)

Phase 10: Caching Layer

Goal: Implement caching for protocol responses.

Deliverables:
- In-memory cache implementation
- Redis cache implementation (optional)
- Per-protocol TTL configuration
- Cache key generation (hash-based)
- Invalidation on new events
- Cache warming strategies
- Tests for cache behavior

Completion Criteria:
- Responses are cached per TTL
- Cache hits reduce database queries
- Invalidation works correctly
- Can switch between memory and Redis
- Tests verify cache correctness

Dependencies: Phase 7, 8, 9

Files to Create:
- internal/cache/cache.go
- internal/cache/memory.go
- internal/cache/redis.go
- internal/cache/keys.go
- internal/cache/cache_test.go

Memory Updates:
- caching.md (if caching strategy changes)
- cache_keys_hashing.md (if key generation changes)

Phase 11: Sections and Layouts

Goal: Implement configurable sections and page layouts.

Deliverables:
- Section definition schema
- Filter query builder
- Default sections (notes, articles, inbox, outbox)
- Archive generation (by month/year)
- Page composition
- Tests for sections

Completion Criteria:
- Sections are configurable
- Filters query correctly
- Default sections work
- Archives generate
- All protocols support sections

Dependencies: Phase 7, 8, 9

Files to Create:
- internal/sections/sections.go
- internal/sections/filters.go
- internal/sections/archives.go
- internal/sections/sections_test.go

Memory Updates:
- layouts_sections.md (if section schema changes)

Phase 12: Operations and Diagnostics

Goal: Add logging, diagnostics, and operational features.

Deliverables:
- Structured logging
- Diagnostics page (via Gopher/Gemini)
- Relay health monitoring
- Cursor status display
- Event count statistics
- Pruning and retention
- Backup utilities
- Tests for ops features

Completion Criteria:
- Logs are structured and configurable
- Diagnostics show system status
- Retention policies work
- Can backup/restore database

Dependencies: Phase 4, 10

Files to Create:
- internal/ops/logging.go
- internal/ops/diagnostics.go
- internal/ops/health.go
- internal/ops/retention.go
- internal/ops/backup.go
- internal/ops/ops_test.go

Memory Updates:
- diagnostics.md (if diagnostic features change)
- operations.md (if ops procedures change)

Phase 13: Publisher (Optional)

Goal: Publish events to Nostr relays.

Deliverables:
- Event signing with nsec
- Publisher to write relays
- Relay health checks
- Retry/backoff logic
- Draft management
- Tests for publishing

Completion Criteria:
- Can sign events with nsec
- Events publish to write relays
- Failed publishes retry
- Health checks prevent bad relays

Dependencies: Phase 3

Files to Create:
- internal/publisher/publisher.go
- internal/publisher/signer.go
- internal/publisher/health.go
- internal/publisher/publisher_test.go

Memory Updates:
- inbox_outbox.md (if publishing logic changes)

Phase 14: Security and Privacy

Goal: Implement security hardening and privacy features.

Deliverables:
- Deny-list enforcement
- Secret handling (nsec env-only)
- Rate limiting
- Input validation
- Privilege separation (systemd socket)
- Security tests

Completion Criteria:
- Deny-lists block pubkeys
- Secrets never in logs/files
- Rate limits prevent abuse
- Input is validated
- Can run unprivileged

Dependencies: All server phases

Files to Create:
- internal/security/denylist.go
- internal/security/ratelimit.go
- internal/security/validation.go
- internal/security/security_test.go

Memory Updates:
- security_privacy.md (if security features change)

Phase 15: Testing and Documentation

Goal: Comprehensive testing and user documentation.

Deliverables:
- Unit tests (>80% coverage)
- Integration tests (full flow)
- Protocol compliance tests
- Example configurations
- Deployment guide
- Troubleshooting guide
- Man page

Completion Criteria:
- Test coverage >80%
- All protocols tested with real clients
- Documentation is complete
- Examples work out-of-box

Dependencies: All previous phases

Files to Create:
- test/integration/*_test.go
- test/compliance/*_test.go
- docs/deployment.md
- docs/troubleshooting.md
- docs/examples/*
- nophr.1 (man page)

Memory Updates:
- testing.md (if testing strategy changes)

Phase 16: Distribution and Packaging

Goal: Finalize distribution artifacts and installers.

Deliverables:
- GoReleaser configuration finalized
- Dockerfiles optimized
- Systemd service files
- One-line installer script
- Reverse proxy examples
- Homebrew formula
- Package repository setup

Completion Criteria:
- Can create full release with one command
- Docker images build multi-arch
- Systemd service works
- Installer script works
- Packages install cleanly

Dependencies: Phase 0, all core phases complete

Files to Create:
- .goreleaser.yml (finalized)
- Dockerfile (optimized)
- scripts/systemd/nophr.service
- scripts/install.sh (user installer)
- docs/reverse-proxy/*

Memory Updates:
- distribution.md (if packaging changes)

Summary

Total Phases: 16

Can Start Immediately (no dependencies):
- Phase 0: Project Bootstrap
- Phase 1: Configuration System
- Phase 6: Markdown Conversion (depends on Phase 1 only)

Critical Path:
Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 7/8/9 → Complete

Parallel Opportunities:
- Phases 7, 8, 9 (protocol servers) can be built in parallel
- Phase 6 (markdown) can be built independently
- Phase 10 (caching) can be added to any server
- Phase 13 (publisher) is independent of servers

Minimum Viable Product (MVP):
- Phases 0-5, 7 (Gopher only), 15 (basic tests), 16 (basic packaging)
- This gives: Working Gopher site with Nostr sync

Full Feature Set:
- All 16 phases
- Gives: Gopher + Gemini + Finger with full features

---

## Phase 17: Endpoint Refactoring and Bug Fixes

**Status**: ✅ Complete

**Goal**: Refactor protocol handlers to use new QueryHelper methods and fix critical bugs in thread retrieval and reply filtering.

**Key Changes**:
- Updated Gopher and Gemini handlers to use new section-based query methods (GetNotes, GetArticles, GetReplies, GetMentions)
- Fixed thread retrieval logic to properly handle root events and replies
- Improved reply filtering to distinguish actual replies from mentions
- Added proper error handling for empty result sets
- Consolidated query logic in aggregates package

**Files Modified**:
- `internal/gopher/handler.go` - Updated to use QueryHelper methods
- `internal/gemini/handler.go` - Updated to use QueryHelper methods
- `internal/aggregates/queries.go` - Fixed thread retrieval logic

**Impact**:
- More consistent behavior across protocols
- Better separation of concerns
- Easier to maintain and extend

---

## Phase 18: Display Customization and Content Control

**Status**: ✅ Complete

**Goal**: Provide comprehensive configuration options for display, presentation, and behavior customization.

**Key Features**:

### 1. Display Configuration
- **Feed vs Detail Views**: Separate control for showing interactions in list views vs detail views
- **Selective Interaction Display**: Toggle replies, reactions, and zaps independently
- **Content Limits**: Configurable summary lengths, max content length, thread depth limits
- **Truncation Control**: Custom truncation indicators

### 2. Presentation System
- **Headers/Footers**: Global and per-page headers/footers
- **Content Sources**: Inline content or file-based loading
- **Template Variables**: `{{site.title}}`, `{{date}}`, `{{year}}`, etc.
- **Caching**: 5-minute TTL for loaded content
- **Protocol Separators**: Customizable item and section separators per protocol

### 3. Behavior Configuration
- **Content Filtering**: Min reactions, min zap sats, min engagement thresholds
- **Quality Filters**: Hide notes with no interactions
- **Sort Preferences**: Per-section sorting (chronological, engagement, zaps, reactions)
- **Pagination**: Framework for future pagination support

### 4. Granular Sync Control
- **Kind Filtering**: Boolean flags for each event kind instead of array
- **User-Friendly**: Named flags (profiles, notes, articles, etc.)
- **Backward Compatible**: ToIntSlice() method converts to integer array

**Configuration Schema**:
```yaml
display:
  feed:
    show_interactions: true
    show_reactions: true
    show_zaps: true
    show_replies: true
  detail:
    show_interactions: true
    show_reactions: true
    show_zaps: true
    show_replies: true
    show_thread: true
  limits:
    summary_length: 100
    max_content_length: 5000
    max_thread_depth: 10
    max_replies_in_feed: 3
    truncate_indicator: "..."

presentation:
  headers:
    global:
      enabled: false
      content: ""
      file_path: ""
    per_page: {}
  footers:
    global:
      enabled: false
      content: ""
      file_path: ""
    per_page: {}
  separators:
    item:
      gopher: ""
      gemini: ""
      finger: ""
    section:
      gopher: "---"
      gemini: "---"
      finger: "---"

behavior:
  content_filtering:
    enabled: false
    min_reactions: 0
    min_zap_sats: 0
    min_engagement: 0
    hide_no_interactions: false
  sort_preferences:
    notes: "chronological"
    articles: "chronological"
    replies: "chronological"
    mentions: "chronological"
  pagination:
    enabled: false
    items_per_page: 50
    max_pages: 10
```

**Files Modified**:
- `internal/config/config.go` - Extended with Display, Presentation, Behavior structs
- `internal/config/sync.go` - Changed SyncKinds to struct with ToIntSlice()
- `internal/aggregates/queries.go` - Added filterAndSortEvents and passesContentFilter methods
- `internal/gopher/renderer.go` - Updated to use display config
- `internal/gemini/renderer.go` - Updated to use display config
- `internal/presentation/loader.go` - NEW: Header/footer loading with caching
- `configs/nophr.example.yaml` - Added comprehensive Phase 18 configuration

**Impact**:
- Users have fine-grained control over what content is displayed
- Customizable visual presentation per protocol
- Quality filtering to reduce noise
- Flexible sorting options for different use cases
- Professional appearance with headers/footers

**Template Variables**:
- `{{site.title}}` - Site title from site.title
- `{{site.description}}` - Site description
- `{{site.operator}}` - Operator name
- `{{date}}` - Current date (YYYY-MM-DD)
- `{{datetime}}` - Current date and time
- `{{year}}` - Current year


---

## Phase 19: Profile Enhancement and Search Implementation

**Status**: ✅ Complete

**Goal**: Implement proper profile parsing, enhance diagnostics, and add functional search across all protocols.

**Key Features**:

### 1. Profile Parsing (Kind 0)
Currently profile endpoints show raw data. Parse kind 0 JSON metadata to display:
- Display name
- About/bio
- Picture URL (as text link)
- NIP-05 identifier
- LN address
- Website

### 2. Search Implementation
Currently search returns "not yet implemented". Add full-text search:
- Search across note content (kind 1)
- Search in profile names/bios (kind 0)
- Search by npub/hex pubkey
- Search by note ID (event ID)
- Protocol-appropriate result formatting

### 3. Enhanced Diagnostics
Expand diagnostics endpoint with actual system stats:
- Storage statistics (event counts by kind)
- Relay connection status
- Sync status and progress
- Cache hit/miss rates
- Uptime
- Memory/database size

### 4. Profile Metadata Helpers
Create reusable profile parsing utilities:
- Parse kind 0 JSON safely
- Extract common fields (name, about, picture, etc.)
- Handle missing/malformed data gracefully
- Fallback to npub if no display name

**Deliverables**:

1. **Profile Parser** ✅
   - `internal/nostr/profile.go` - Kind 0 JSON parsing
   - `internal/nostr/profile_test.go` - Comprehensive tests
   - Handles all common kind 0 fields

2. **Profile Rendering** ✅
   - Update `internal/gopher/renderer.go` RenderProfile()
   - Update `internal/gemini/renderer.go` RenderProfile()
   - Update `internal/finger/renderer.go` QueryUser()

3. **Search Implementation** ✅
   - `internal/search/engine.go` - Search logic
   - `internal/search/engine_test.go` - Search tests
   - SQLite FTS (Full-Text Search) or simple LIKE queries
   - Pluggable interface for future improvements

4. **Search Endpoints** ✅
   - Update `internal/gopher/router.go` handleSearch()
   - Update `internal/gemini/router.go` handleSearch()
   - Handle empty queries (prompt for input)
   - Display ranked results

5. **Diagnostics Enhancement** ✅
   - Implement stats in `internal/storage/stats.go`
   - Update `internal/gopher/router.go` handleDiagnostics()
   - Update `internal/gemini/router.go` handleDiagnostics()
   - Real-time data, not just static messages

6. **Tests** ✅
   - Profile parsing tests
   - Search functionality tests
   - Diagnostics accuracy tests
   - Integration tests for all protocols

**Completion Criteria**:

- [x] Profile endpoints display parsed metadata (name, bio, picture link)
- [x] Search returns relevant results for content queries
- [x] Search works by npub/pubkey
- [x] Search works by event ID
- [x] Diagnostics show real statistics
- [x] All unit tests pass
- [x] Integration tests verify functionality
- [x] No regressions in existing features

**Files to Create**:
- `internal/nostr/profile.go`
- `internal/nostr/profile_test.go`
- `internal/search/engine.go`
- `internal/search/engine_test.go`

**Files to Modify**:
- `internal/gopher/renderer.go`
- `internal/gopher/router.go`
- `internal/gemini/renderer.go`
- `internal/gemini/router.go`
- `internal/finger/renderer.go`
- `internal/storage/stats.go`

**Dependencies**: None (Phase 18 complete)

---

## Phase 20: Advanced Configurable Retention

**Status**: 📋 Future Enhancement

**Goal**: Extend the simple retention system (Phase 12) with sophisticated rule-based retention capabilities.

**Overview**:

Phase 12 implemented simple time-based retention (`keep_days` configuration). Phase 20 would add an advanced, multi-dimensional retention system on top of this foundation.

**Proposed Features**:
- **Rule Engine**: Priority-based rule matching with configurable conditions
- **Multi-Dimensional Conditions**:
  - Social distance (FOAF, mutual, specific pubkeys)
  - Engagement thresholds (reactions, zaps, replies)
  - Content characteristics (kind, size, age)
  - Reference-based (replies to owner, mentions)
- **Sophisticated Caps**: Global and per-kind storage limits with score-based pruning
- **Flexible Actions**: Retain forever, retain until date, or delete
- **Protected Events**: Mark important events as never-delete

**Why Not Yet Implemented**:
- Phase 12 simple retention meets current needs
- Would add significant complexity
- Requires careful design to avoid performance impact
- Optional enhancement, not core requirement

**Documentation**: See `memory/PHASE_20_ADVANCED_RETENTION.md` for complete specification.

**Dependencies**: Phase 12 (Simple Retention) - Complete

---

## Summary (Updated)

**Total Phases**: 19 (Complete), 20+ (Future)
**Status**: Core Features Complete

**Phase Breakdown**:
- ✅ Phases 0-19: Complete
- 📋 Phase 20: Advanced Retention (Future Enhancement)
- 🚧 Future: Additional features and optimizations

**Current Status**: All core features implemented and operational

**Architecture Highlights**:
- Multi-protocol support (Gopher, Gemini, Finger)
- Intelligent relay discovery (NIP-65)
- Aggregate computation and caching
- Flexible content filtering and sorting
- Customizable presentation system
