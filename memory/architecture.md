Architecture Overview

High-Level Design

Nopher bridges the Nostr protocol with classic internet protocols (Gopher/Gemini/Finger).

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

Technology Stack

Language: Go
- Mature concurrency (goroutines)
- Strong networking libraries
- Cross-platform compilation
- Excellent performance

Core Dependencies:
- Khatru: Nostr relay framework (https://github.com/fiatjaf/khatru)
  - Provides event storage, validation, querying
  - Pluggable eventstore backends
- eventstore: Database adapters (https://github.com/fiatjaf/eventstore)
  - SQLite: default, single-file, zero-config
  - LMDB: alternative, excellent for high-volume streaming
- nostr-sdk or websocket client: Connect to remote Nostr relays
- Markdown parser: Convert Nostr content to protocol-specific formats

Why Khatru?

Benefits:
1. Battle-tested event storage - used by production Nostr relays
2. NIP compliance out-of-box - handles replaceable events, threading, etc.
3. Signature verification - automatic validation of Nostr events
4. Pluggable backends - SQLite/LMDB/others via eventstore
5. Simpler codebase - no need to implement event plumbing
6. Future-proof - maintains compatibility with Nostr protocol evolution
7. Go ecosystem - matches our language choice

What Khatru Handles:
- Event storage, indexing, deduplication
- Replaceable event logic (kinds 0, 3, 10002, 30023, etc.)
- Event validation and signature verification
- Querying via standard Nostr filters
- WebSocket relay interface (optional; we may use internal API only)

What We Implement:
- Sync engine (pull from remote relays → push to local Khatru)
- Relay discovery and social graph computation
- Aggregates (interaction counts)
- Protocol servers (Gopher, Gemini, Finger)
- Markdown conversion and rendering
- Caching and performance optimization

Component Breakdown

1. Sync Engine
- Connects to remote Nostr relays via WebSocket
- Subscribes with filters (authors, kinds, since cursors)
- Streams events to local Khatru instance
- Manages relay_hints table (NIP-65 discovery)
- Computes graph_nodes (social graph/FOAF)
- Tracks sync_state (cursors per relay/kind)

2. Khatru Instance (Embedded)
- Initialized with SQLite or LMDB eventstore
- Receives events via StoreEvent()
- Provides events via QueryEvents(filter)
- Handles all event validation and storage logic
- Not exposed externally (private relay for Nopher only)

3. Query Layer
- Queries Khatru for events matching section filters
- Resolves threads (NIP-10 root/reply relationships)
- Computes aggregates (reply counts, reactions, zaps)
- Caches results in memory or Redis

4. Renderers
- GopherRenderer: Events → plain text + gophermaps
  - Markdown → plain text (headings, links, code)
  - Line wrapping at 70 chars
  - Thread indentation
- GeminiRenderer: Events → gemtext
  - Markdown → gemtext (headings, links, quotes, code)
  - Extract inline links to separate lines
  - Optional wrapping at 80 chars
- FingerRenderer: Events → compact plain text
  - Strip all markdown
  - Truncate to ~500 chars
  - Profile + recent notes

5. Protocol Servers
- Gopher: TCP listener on port 70
  - Selector routing (/, /notes, /event/id, etc.)
  - Serve gophermaps (menus) or text files
- Gemini: TLS listener on port 1965
  - URL routing
  - Serve gemtext responses
  - Handle input queries (status 10)
  - Self-signed cert or custom cert
- Finger: TCP listener on port 79
  - Query parsing (user@host)
  - Serve user info (owner + followed users)

Data Flow Examples

Example 1: Syncing a New Note

1. Sync engine subscribes to author's write relays (from relay_hints)
2. Remote relay sends EVENT message with kind 1 note
3. Sync engine calls khatru.StoreEvent(event)
4. Khatru validates signature, checks for duplicates
5. Khatru stores event in SQLite/LMDB via eventstore
6. Sync engine updates sync_state cursor for relay/kind
7. If event is a reply, sync engine updates aggregates table

Example 2: Serving a Gopher Menu

1. Client connects to port 70, sends selector "/notes"
2. Gopher server queries Khatru: filter={kinds:[1], authors:[owner]}
3. Khatru returns events from eventstore
4. Query layer fetches aggregates (reaction/zap counts)
5. GopherRenderer converts events to gophermap format
6. Cache layer stores rendered gophermap (TTL 300s)
7. Server sends gophermap to client

Example 3: Rendering a Gemini Article

1. Client connects to port 1965 (TLS), requests gemini://host/article/xyz
2. Gemini server queries Khatru: filter={kinds:[30023], ids:[xyz]}
3. Khatru returns long-form article event
4. GeminiRenderer converts markdown content to gemtext
5. Renderer adds thread links (replies to this article)
6. Cache layer stores rendered gemtext (TTL 300s)
7. Server sends gemtext response with status 20

Example 4: Finger Query

1. Client connects to port 79, sends "npub1abc@gopher.example.com"
2. Finger server looks up npub1abc in graph_nodes (is it owner or followed?)
3. Queries Khatru: filter={kinds:[0], authors:[npub1abc]} for profile
4. Queries Khatru: filter={kinds:[1], authors:[npub1abc], limit:5} for recent notes
5. FingerRenderer formats as plain text (.plan + notes)
6. Server sends finger response to client

Configuration Philosophy

Everything is configurable via YAML + environment variable overrides:
- Which protocols to run (all, some, or one)
- Which database backend (SQLite or LMDB)
- Sync scope (self/following/mutual/FOAF depth N)
- Rendering options (line length, date format, markdown style)
- Caching TTLs per content type
- Relay policies (timeouts, max connections, backoff)

Secrets (nsec, DB credentials) are environment-only, never in config files.

Deployment Scenarios

Single Binary:
- Compile to single static binary with embedded Khatru
- Run on VPS, homelab, or localhost
- No external dependencies (besides chosen DB file)

Multi-Protocol:
- Run all three servers simultaneously on one host
- Gopher on port 70, Gemini on 1965, Finger on 79
- Share same Khatru instance and database

Minimal Resources:
- SQLite backend: <100MB disk for typical personal use
- <50MB RAM with conservative caching
- Single-core CPU sufficient for personal traffic

Scalability Notes

Single-tenant design:
- Optimized for one owner's content
- Supports thousands of followed users (with caps)
- LMDB backend scales to millions of events if needed

Performance:
- Khatru handles event storage efficiently
- Caching layer reduces database queries
- Protocol servers are lightweight (no HTTP overhead)
- Go's concurrency handles multiple protocol servers in one process

Limitations (by design):
- Not a multi-tenant platform
- Not optimized for public relay usage
- Protocols are read-mostly (publishing is optional)
- No real-time updates in protocols (Gopher/Gemini/Finger are pull-based)

Future Considerations

- Optional publishing: Sign and publish notes via Khatru to remote relays
- Protocol extensions: Gopher+ for metadata, Gemini subscriptions
- Advanced caching: Pre-render popular pages, stale-while-revalidate
- Monitoring: Prometheus metrics for relay health, cache hit rates
- Backup/export: Periodic snapshots of SQLite/LMDB database
- Migration tools: Import/export between SQLite and LMDB
