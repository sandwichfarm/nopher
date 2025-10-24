# Nostr Integration Guide

**Status:** ✅ VERIFIED (Integrated and working)

Complete guide to how nophr integrates with Nostr: relay discovery, event synchronization, social graph computation, and interaction aggregation.

## Overview

nophr acts as a personal Nostr archive, syncing events from remote relays to a local storage layer, then serving them via Gopher/Gemini/Finger protocols.

**Data flow:**
```
Remote Nostr Relays
        ↓
  (WebSocket subscriptions)
        ↓
    Sync Engine
        ↓
  Local Storage (Khatru)
        ↓
 Protocol Servers (Gopher/Gemini/Finger)
```

**Key components:**
1. **Relay Discovery** - Find relays via NIP-65 (kind 10002)
2. **Sync Engine** - Pull events matching scope/filters
3. **Social Graph** - Compute follows/mutuals/FOAF
4. **Aggregates** - Count replies, reactions, zaps
5. **Threading** - Resolve NIP-10 thread relationships

---

## Table of Contents

- [Relay Discovery](#relay-discovery) - NIP-65 dynamic relay hints
- [Sync Engine](#sync-engine) - Event synchronization
- [Sync Scope](#sync-scope) - Control whose events to sync
- [Social Graph](#social-graph) - Follows, mutuals, FOAF
- [Aggregates](#aggregates) - Interaction counting
- [Threading](#threading) - NIP-10 thread resolution
- [Retention](#retention) - Data pruning

---

## Relay Discovery

**NIP:** [NIP-65](https://github.com/nostr-protocol/nips/blob/master/65.md)
**Status:** 🟡 IMPLEMENTED (`internal/nostr/discovery.go`)

nophr uses **dynamic relay discovery** via kind 10002 events. This allows you to change your Nostr relays without updating nophr's config.

### How It Works

1. **Seed relays** (from config) are used for initial bootstrap
2. Fetch kind 10002 (relay hints) for owner and followed users
3. Parse relay hints to build per-pubkey relay sets
4. Subscribe to discovered relays for targeted queries
5. Refresh periodically to catch relay changes

### Configuration

```yaml
relays:
  seeds:
    - "wss://relay.damus.io"
    - "wss://relay.nostr.band"
    - "wss://nos.lol"
  policy:
    connect_timeout_ms: 5000
    max_concurrent_subs: 8
    backoff_ms: [500, 1500, 5000]

discovery:
  refresh_seconds: 900          # Refresh every 15 min
  use_owner_hints: true          # Use owner's kind 10002
  use_author_hints: true         # Use followed users' kind 10002
  fallback_to_seeds: true        # Use seeds if hints missing
  max_relays_per_author: 8      # Safety cap per author
```

### Seed Relays

**Purpose:** Bootstrap relay discovery.

**Characteristics:**
- Used only for initial discovery (not permanent)
- Should have good coverage of kinds 0, 3, 10002
- Configurable (not hard-coded)

**Popular seed choices:**
```yaml
seeds:
  - "wss://relay.damus.io"       # Large, reliable
  - "wss://relay.nostr.band"     # Aggregator, good coverage
  - "wss://nos.lol"              # Fast, well-connected
  - "wss://relay.snort.social"   # Popular
  - "wss://nostr.wine"           # Paid, quality
```

### Kind 10002 (Relay Hints)

**Format:**
```json
{
  "kind": 10002,
  "tags": [
    ["r", "wss://relay.damus.io"],
    ["r", "wss://nos.lol", "read"],
    ["r", "wss://relay.primal.net", "write"]
  ],
  "content": "",
  ...
}
```

**Tag format:** `["r", "<url>", "<usage>"]`
- No usage marker = read + write
- `"read"` = read only
- `"write"` = write only

**Parsing:**
- Read relays: used to fetch that author's events
- Write relays: used when publishing (future feature)

### relay_hints Table

Discovered relays are stored:

```sql
CREATE TABLE relay_hints (
  pubkey TEXT NOT NULL,
  relay TEXT NOT NULL,
  can_read INTEGER NOT NULL,    -- 0 or 1
  can_write INTEGER NOT NULL,   -- 0 or 1
  freshness INTEGER NOT NULL,   -- created_at of hint event
  last_seen_event_id TEXT NOT NULL,
  PRIMARY KEY (pubkey, relay)
);
```

**Query relay hints:**
```bash
sqlite3 ./data/nophr.db "SELECT * FROM relay_hints WHERE pubkey = 'hex_pubkey';"
```

### Refresh Strategy

**Periodic refresh:**
- Every `discovery.refresh_seconds` (default: 900 = 15 min)
- Fetches latest kind 10002 for all in-scope authors
- Updates `relay_hints` table

**Triggered refresh:**
- On sync errors (relay unreachable)
- When discovering new authors (follows)

### Fallback Behavior

If relay hints are missing/stale:
1. Use seed relays as fallback
2. Mark author as "pending discovery"
3. Retry discovery on next refresh

---

## Sync Engine

**Status:** ✅ VERIFIED (Integrated in cmd/nophr/main.go, controlled by `sync.enabled`)

The sync engine pulls events from remote Nostr relays and stores them locally.

**Configuration:**
```yaml
sync:
  enabled: true  # Enable/disable sync engine
```

When enabled, the sync engine:
- Starts automatically on nophr startup
- Connects to relays based on discovery
- Syncs events matching configured scope
- Updates cursors to track progress

### Architecture

```
┌──────────────────────┐
│ Remote Relay Pool    │
│  (WebSocket clients) │
└──────────┬───────────┘
           │ Subscribe(filters)
           ↓
┌──────────────────────┐
│   Filter Builder     │
│  (scope + kinds)     │
└──────────┬───────────┘
           │ Built filters
           ↓
┌──────────────────────┐
│   Event Ingestion    │
│  (validation, dedup) │
└──────────┬───────────┘
           │ StoreEvent()
           ↓
┌──────────────────────┐
│  Local Storage       │
│  (Khatru + custom)   │
└──────────────────────┘
```

### Subscription Model

**Per-relay, per-kind subscriptions:**
- Separate subscription for each (relay, kind) pair
- Enables fine-grained cursor tracking
- Prevents re-fetching old events

**Subscription format:**
```json
{
  "authors": ["<hex_pubkey>", ...],
  "kinds": [1],
  "since": 1698765432,
  "limit": 1000
}
```

### Cursor Tracking

**sync_state Table:**
```sql
CREATE TABLE sync_state (
  relay TEXT NOT NULL,
  kind INTEGER NOT NULL,
  since INTEGER NOT NULL,       -- Latest event timestamp
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (relay, kind)
);
```

**Purpose:**
- Track progress per (relay, kind)
- Resume sync after restart
- Avoid re-syncing old events

**Cursor update:**
- After each batch: update `since` to latest event timestamp
- Next subscription uses new `since` value

### Event Ingestion Pipeline

1. **Receive event** from relay (WebSocket)
2. **Validate** - signature, format (Khatru handles this)
3. **Deduplicate** - check if already stored (Khatru)
4. **Store** - write to Khatru eventstore
5. **Update cursors** - update `sync_state`
6. **Trigger aggregates** - update interaction counts (if configured)

### Replaceable Events

**Replaceable kinds:** 0, 3, 10002, 30023 (parameterized)

**Handling:**
- Khatru automatically keeps only latest per (pubkey, kind)
- Parameterized (kind 30023): keeps latest per (pubkey, kind, d-tag)
- Old versions automatically replaced

**Refresh strategy:**
- Periodic re-fetch to catch updates
- Ignores cursors for replaceable kinds

---

## Sync Scope

**Status:** 🟡 IMPLEMENTED (`internal/sync/scope.go`)

Control whose events to synchronize.

### Sync Modes

```yaml
sync:
  scope:
    mode: "foaf"   # self | following | mutual | foaf
    depth: 2       # FOAF depth (when mode=foaf)
```

| Mode | Description | Authors Synced |
|------|-------------|----------------|
| **self** | Only your events | 1 (you) |
| **following** | You + who you follow | 1 + contacts count |
| **mutual** | You + bidirectional follows | 1 + mutual follows |
| **foaf** | Friend-of-a-friend | Exponential by depth |

### FOAF (Friend-of-a-Friend)

**Depth examples:**
- **Depth 1:** You + following (same as `following` mode)
- **Depth 2:** You + following + their follows (2nd degree)
- **Depth 3:** 3rd degree connections (can be huge!)

**Configuration:**
```yaml
sync:
  scope:
    mode: "foaf"
    depth: 2
    max_authors: 5000   # Safety cap
```

**Warning:** FOAF grows exponentially. Use `max_authors` cap!

**Example growth:**
- You follow: 100
- Each follows: 100
- Depth 2: 100 + (100 * 100) = 10,100 authors (capped at 5,000)

### Modifiers

```yaml
sync:
  scope:
    include_direct_mentions: true      # Events mentioning you
    include_threads_of_mine: true      # Full threads under your posts
    allowlist_pubkeys: ["npub1...", "npub2..."]
    denylist_pubkeys: ["npub3...", "npub4..."]
    max_authors: 5000
```

| Modifier | Effect |
|----------|--------|
| `include_direct_mentions` | Always include events with `#p` tag matching you |
| `include_threads_of_mine` | Include all replies to your events (regardless of author) |
| `allowlist_pubkeys` | Always include these pubkeys (bypass mode) |
| `denylist_pubkeys` | Never include these pubkeys (spam/block) |
| `max_authors` | Stop expansion when cap reached |

### Event Kinds

```yaml
sync:
  kinds: [0, 1, 3, 6, 7, 9735, 30023, 10002]
```

| Kind | Description | Purpose |
|------|-------------|---------|
| 0 | Profile (metadata) | User info, names, avatars |
| 1 | Short note | Text posts |
| 3 | Contacts (follows) | Social graph |
| 6 | Repost | Shares/boosts |
| 7 | Reaction | Likes, emoji reactions |
| 9735 | Zap receipt | Lightning tips |
| 30023 | Long-form article | Blog posts |
| 10002 | Relay hints | Relay discovery |

**Add more kinds:**
```yaml
kinds: [0, 1, 3, 6, 7, 9735, 30023, 10002, 30311]  # Add app handlers
```

---

## Social Graph

**Status:** 🟡 IMPLEMENTED (`internal/sync/graph.go`)

nophr computes the social graph from kind 3 (contacts) events.

### graph_nodes Table

```sql
CREATE TABLE graph_nodes (
  root_pubkey TEXT NOT NULL,    -- Owner
  pubkey TEXT NOT NULL,
  depth INTEGER NOT NULL,       -- FOAF distance
  mutual INTEGER NOT NULL,      -- 0 or 1 (bidirectional follow)
  last_seen INTEGER NOT NULL,
  PRIMARY KEY (root_pubkey, pubkey)
);
```

**Columns:**
- `root_pubkey` - The owner (you)
- `pubkey` - Discovered author
- `depth` - Distance from owner (1 = direct follow, 2 = FOAF, etc.)
- `mutual` - Whether follow is bidirectional
- `last_seen` - Timestamp of last update

### Graph Computation

**Algorithm:**
1. Fetch owner's kind 3 (contacts list)
2. For each contact, set `depth = 1`
3. For each contact, fetch their kind 3
4. Check if they follow owner back → set `mutual = 1`
5. If mode=foaf, recurse to depth N

**Mutual detection:**
```
Owner follows Alice:  owner → alice
Alice follows Owner:  alice → owner
→ mutual = 1
```

### Query Graph

**All followers:**
```bash
sqlite3 ./data/nophr.db "SELECT pubkey, depth, mutual FROM graph_nodes WHERE root_pubkey = 'owner_hex';"
```

**Only mutuals:**
```bash
sqlite3 ./data/nophr.db "SELECT pubkey FROM graph_nodes WHERE root_pubkey = 'owner_hex' AND mutual = 1;"
```

### Performance

**Graph size:**
- following mode: ~100-1,000 nodes
- mutual mode: ~50-500 nodes
- foaf depth=2: ~5,000-10,000 nodes (with cap)

**Refresh frequency:**
- When owner updates kind 3
- Periodically (e.g., hourly)

---

## Aggregates

**Status:** ✅ VERIFIED (`internal/aggregates/aggregates.go`)

Interaction aggregation: replies, reactions, zaps.

### aggregates Table

```sql
CREATE TABLE aggregates (
  event_id TEXT PRIMARY KEY,
  reply_count INTEGER,
  reaction_total INTEGER,
  reaction_counts_json TEXT,    -- JSON: {"char": count}
  zap_sats_total INTEGER,
  last_interaction_at INTEGER
);
```

### Computation

**Replies (kind 1):**
- Find events with `#e` tag pointing to `event_id`
- Use NIP-10 markers (`reply`, `root`) if present
- Count replies per event

**Reactions (kind 7):**
- Find events with `#e` tag pointing to `event_id` and `kind=7`
- Parse `content` field (emoji, "+", etc.)
- Count per reaction type

**Zaps (kind 9735):**
- Find events with `#e` tag pointing to `event_id` and `kind=9735`
- Parse bolt11 invoice in `description` tag
- Sum satoshi amounts

### Update Strategy

```yaml
caching:
  aggregates:
    enabled: true
    update_on_ingest: true              # Update on new events
    reconciler_interval_seconds: 900    # Periodic reconciliation
```

**On ingestion:**
- When new reply/reaction/zap arrives
- Update aggregate for referenced event
- Fast, immediate update

**Reconciler:**
- Periodic full recount (detect drift)
- Runs every `reconciler_interval_seconds`
- Corrects any inconsistencies

### Example

**Event:**
```
event_id: abc123
```

**Interactions:**
- 5 replies (kind 1 with #e = abc123)
- 12 reactions: 10x "+", 2x "❤️"
- 3 zaps: 10,000 + 5,000 + 6,000 = 21,000 sats

**Aggregate:**
```sql
INSERT INTO aggregates VALUES (
  'abc123',
  5,                     -- reply_count
  12,                    -- reaction_total
  '{"+": 10, "❤️": 2}',  -- reaction_counts_json
  21000,                 -- zap_sats_total
  1698765500             -- last_interaction_at
);
```

---

## Threading

**NIP:** [NIP-10](https://github.com/nostr-protocol/nips/blob/master/10.md)
**Status:** ✅ VERIFIED (`internal/aggregates/threading.go`)

Thread resolution from NIP-10 tags.

### NIP-10 Format

**Marked format (recommended):**
```json
{
  "kind": 1,
  "tags": [
    ["e", "<root_id>", "", "root"],
    ["e", "<reply_id>", "", "reply"]
  ],
  "content": "This is a reply"
}
```

**Positional format (deprecated but supported):**
```json
{
  "kind": 1,
  "tags": [
    ["e", "<root_id>"],         // First = root
    ["e", "<parent_id>"],       // Last = parent
    ["e", "<mention_id>"]       // Middle = mention
  ]
}
```

### Thread Resolution

**Algorithm:**
1. Parse all `#e` tags in event
2. Look for `root` marker → root event
3. Look for `reply` marker → parent event
4. If no markers, use positional:
   - First `#e` = root
   - Last `#e` = parent
5. Store in thread hierarchy

### Display

**Gopher (indented):**
```
1Root post	/event/root_id	host	70
1  Reply by Bob	/event/reply1_id	host	70
1    Reply by Carol	/event/reply2_id	host	70
1  Reply by Dave	/event/reply3_id	host	70
```

**Gemini (links with context):**
```gemtext
## Root post

=> /event/reply1_id Reply by Bob
=> /event/reply2_id   Reply by Carol (nested)
=> /event/reply3_id Reply by Dave
```

---

## Retention

**Status:** 🟡 IMPLEMENTED (basic), 🚧 IN PROGRESS (advanced - Phase 17)

Data pruning to manage disk space.

### Basic Retention

```yaml
sync:
  retention:
    keep_days: 365          # Keep events newer than N days
    prune_on_start: true    # Prune at startup
```

**What gets pruned:**
- Events older than `keep_days`
- Except: kind 0 (profiles), kind 3 (follows) - never pruned
- Replaceable events: only latest kept anyway

**Manual pruning:**
```bash
# Future feature
nophr --config nophr.yaml --prune
```

### Advanced Retention (Phase 17)

**Planned features:**
- Per-kind retention rules
- Per-author retention rules
- Score-based pruning (keep popular events longer)
- Protected events (never delete)

See [memory/PHASE_17_RETENTION.md](../memory/PHASE_17_RETENTION.md) for details.

---

## Inbox/Outbox

**Inbox:** 🟡 IMPLEMENTED (aggregates)
**Outbox:** 📋 PLANNED (Phase 13 - publishing)

### Inbox

Your inbox shows interactions with your content:
- Replies to your notes
- Reactions to your content
- Zaps you received

**Configuration:**
```yaml
inbox:
  include_replies: true
  include_reactions: true
  include_zaps: true
  group_by_thread: true
  collapse_reposts: true
  noise_filters:
    min_zap_sats: 1                     # Minimum zap amount
    allowed_reaction_chars: ["+"]       # Filter reactions
```

**Noise filtering:**
```yaml
noise_filters:
  min_zap_sats: 100         # Ignore tiny zaps
  allowed_reaction_chars:   # Only these reactions
    - "+"
    - "❤️"
    - "🔥"
```

### Outbox (Publishing - Future)

**Planned:** Publish events from nophr to Nostr relays.

```yaml
outbox:
  publish:
    notes: true
    reactions: false
    zaps: false
  draft_dir: "./content"
  auto_sign: false
```

**Status:** Phase 13 (not yet implemented)

---

## Monitoring

### Check Sync Status

**Event count:**
```bash
sqlite3 ./data/nophr.db "SELECT COUNT(*) FROM events;"
```

**Events by kind:**
```bash
sqlite3 ./data/nophr.db "SELECT kind, COUNT(*) FROM events GROUP BY kind;"
```

**Sync cursors:**
```bash
sqlite3 ./data/nophr.db "SELECT * FROM sync_state ORDER BY updated_at DESC LIMIT 10;"
```

**Graph size:**
```bash
sqlite3 ./data/nophr.db "SELECT COUNT(*) FROM graph_nodes;"
```

**Relay hints:**
```bash
sqlite3 ./data/nophr.db "SELECT pubkey, COUNT(*) FROM relay_hints GROUP BY pubkey LIMIT 10;"
```

### Diagnostics Page

**Access via protocols:**
- Gopher: `/diagnostics`
- Gemini: `/diagnostics`

**Shows:**
- Relay health (connected, errors)
- Sync state (cursors, last update)
- Event counts per kind
- Author counts by depth

**Status:** 📋 PLANNED (Phase 12)

---

## Troubleshooting

### No events syncing

**Check:**
1. Verify `npub` in config is correct
2. Check seed relays are reachable
3. Verify `sync.kinds` includes desired kinds
4. Check `sync.scope.mode` isn't too restrictive

**Debug:**
```bash
# Check if kind 10002 was fetched
sqlite3 ./data/nophr.db "SELECT * FROM events WHERE kind = 10002;"

# Check relay hints
sqlite3 ./data/nophr.db "SELECT * FROM relay_hints;"
```

### Slow sync

**Causes:**
- Too many authors (FOAF depth too high)
- Too many relays
- Slow seed relays

**Fixes:**
- Reduce `sync.scope.depth`
- Lower `sync.scope.max_authors`
- Use fewer, faster seed relays
- Increase `relays.policy.max_concurrent_subs`

### Missing interactions

**Check:**
- `sync.kinds` includes 7 (reactions), 9735 (zaps)
- `inbox.include_replies/reactions/zaps` enabled
- Events are within `retention.keep_days`

### Relay connection errors

**Check logs:**
```bash
journalctl -u nophr -f
```

**Common issues:**
- Relay down/unreachable
- Firewall blocking WebSocket (wss://)
- Invalid relay URL

---

## Implementation Details

**Code locations:**
- Relay discovery: `internal/nostr/discovery.go`, `internal/nostr/relay_hints.go`
- Sync engine: `internal/sync/engine.go`, `internal/sync/filters.go`
- Social graph: `internal/sync/graph.go`
- Aggregates: `internal/aggregates/aggregates.go`, `internal/aggregates/reconciler.go`
- Threading: `internal/aggregates/threading.go`
- Cursors: `internal/sync/cursors.go`

**Design docs:**
- [memory/relay_discovery.md](../memory/relay_discovery.md)
- [memory/sync_scope.md](../memory/sync_scope.md)
- [memory/inbox_outbox.md](../memory/inbox_outbox.md)
- [memory/sequence_seed_discovery_sync.md](../memory/sequence_seed_discovery_sync.md)

---

## References

- **NIP-01** (Basic protocol): https://github.com/nostr-protocol/nips/blob/master/01.md
- **NIP-10** (Threading): https://github.com/nostr-protocol/nips/blob/master/10.md
- **NIP-65** (Relay hints): https://github.com/nostr-protocol/nips/blob/master/65.md
- **Nostr NIPs**: https://github.com/nostr-protocol/nips

---

**Next:** [Architecture Overview](architecture.md) | [Deployment Guide](deployment.md) | [Configuration Reference](configuration.md)
