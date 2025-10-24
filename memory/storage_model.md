Storage Model

Architecture
- Uses Khatru (https://github.com/fiatjaf/khatru) as embedded local Nostr relay for event storage.
- Khatru provides battle-tested event storage, querying, deduplication, and NIP compliance.
- Sync engine pulls events from remote Nostr relays and stores them in the local Khatru instance.
- Protocol renderers (Gopher/Gemini/Finger) query the local Khatru instance for content.
- Written in Go; Khatru is a library that we instantiate and configure, not a separate service.

Database Backend (via eventstore)
- Primary: SQLite (default, single-file, zero-config, excellent for single-tenant)
- Alternative: LMDB (excellent for streaming large numbers of events and migrations)
- Both supported via Khatru's eventstore plugin system (https://github.com/fiatjaf/eventstore)
- No PostgreSQL support (removed from requirements)

What Khatru Handles
- Event storage, indexing, and querying (standard Nostr filters)
- Event validation and signature verification
- Deduplication and replaceable event handling (kinds 0, 3, 10002, 30023, etc.)
- NIP-01 compliance (basic protocol), NIP-10 (threading), NIP-33 (parameterized replaceable)
- WebSocket relay interface (optional; we may use internal API only)

Custom Tables (augment Khatru's event storage)

Since Khatru handles core event storage, we only need additional tables for Nopher-specific features:

- relay_hints (from NIP-65 kind 10002)
  - pubkey TEXT NOT NULL
  - relay TEXT NOT NULL
  - can_read INTEGER NOT NULL    # 0/1
  - can_write INTEGER NOT NULL   # 0/1
  - freshness INTEGER NOT NULL   # created_at of hint event
  - last_seen_event_id TEXT NOT NULL
  - PRIMARY KEY (pubkey, relay)
  Purpose: Track which relays to query for each author; built from kind 10002 events.

- graph_nodes (owner-centric social graph cache)
  - root_pubkey TEXT NOT NULL    # the owner
  - pubkey TEXT NOT NULL
  - depth INTEGER NOT NULL       # FOAF distance from owner
  - mutual INTEGER NOT NULL      # 0/1
  - last_seen INTEGER NOT NULL
  - PRIMARY KEY (root_pubkey, pubkey)
  Purpose: Efficiently determine which authors are in sync scope (following/mutual/FOAF).

- sync_state (cursor tracking per relay/kind)
  - relay TEXT NOT NULL
  - kind INTEGER NOT NULL
  - since INTEGER NOT NULL       # since cursor for subscriptions
  - updated_at INTEGER NOT NULL
  - PRIMARY KEY (relay, kind)
  Purpose: Avoid re-syncing old events; track progress per relay.

- aggregates (interaction rollups)
  - event_id TEXT PRIMARY KEY
  - reply_count INTEGER
  - reaction_total INTEGER
  - reaction_counts_json TEXT    # JSON map char->count
  - zap_sats_total INTEGER
  - last_interaction_at INTEGER
  Purpose: Cache interaction counts for Gopher/Gemini display; computed from refs in events.

Indexes
- relay_hints(pubkey, freshness DESC)
- graph_nodes(root_pubkey, depth, mutual)

Implementation Notes
- Khatru's eventstore handles the events table with its own optimized schema.
- We query Khatru for events using standard Nostr filters (authors, kinds, #e, #p tags).
- Our custom tables are separate SQLite/LMDB tables in the same database or separate file.
- Sync engine writes events via Khatru's StoreEvent; reads via QueryEvents.
- Aggregates computed on-demand or via background reconciler; cached for performance.

Benefits of Using Khatru
- No need to implement event storage, indexing, or replaceable event logic.
- Battle-tested event validation and signature verification.
- NIP compliance out-of-box.
- Future-proof: supports any eventstore backend (SQLite, LMDB, etc.).
- Simpler codebase: focus on sync logic and protocol rendering, not event plumbing.
