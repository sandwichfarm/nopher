# Configuration Reference

**Status:** ✅ VERIFIED

Complete reference for Nopher's YAML configuration file.

## Overview

Nopher uses YAML for configuration with environment variable overrides for secrets. Configuration is validated on startup.

**Generate example configuration:**
```bash
nopher init > nopher.yaml
```

**Load configuration:**
```bash
nopher --config nopher.yaml
```

## Configuration Sections

- [site](#site) - Site metadata
- [identity](#identity) - Your Nostr identity
- [protocols](#protocols) - Protocol server settings
- [relays](#relays) - Seed relays and policies
- [discovery](#discovery) - Relay discovery (NIP-65)
- [sync](#sync) - Event synchronization scope
- [inbox](#inbox) - Interaction aggregation
- [outbox](#outbox) - Publishing settings
- [storage](#storage) - Database backend
- [rendering](#rendering) - Protocol-specific rendering
- [caching](#caching) - Response caching
- [logging](#logging) - Logging configuration
- [layout](#layout) - Custom sections and pages
- [security](#security) - Security features (deny lists, rate limiting, validation)

---

## site

Site metadata displayed in protocol responses.

```yaml
site:
  title: "My Notes"
  description: "Personal Nostr gopherhole"
  operator: "Alice"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | Site name shown in menus/headers |
| `description` | string | Yes | Brief site description |
| `operator` | string | Yes | Your name or handle |

**Example:**
```yaml
site:
  title: "Alice's Nostr Archive"
  description: "Notes, articles, and interactions from Nostr"
  operator: "Alice (@alice)"
```

---

## identity

Your Nostr identity (public and private keys).

```yaml
identity:
  npub: "npub1..." # Your Nostr public key (REQUIRED)
  # nsec is NEVER in config - use NOPHER_NSEC env var
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `npub` | string | **Yes** | Your Nostr public key (npub1...) |
| `nsec` | string | No | **NEVER IN FILE** - Set via `NOPHER_NSEC` env var |

**Security:**
- ✅ `npub` goes in config file (public key, safe to share)
- ❌ `nsec` NEVER in config file (private key, keep secret!)
- ✅ Set `nsec` via environment: `export NOPHER_NSEC="nsec1..."`

**Get your npub:**
- From any Nostr client (profile settings)
- Must start with `npub1`

**Example:**
```yaml
identity:
  npub: "npub1a2b3c4d5e6f7g8h9i0j..."
```

```bash
# Set private key for publishing (optional, future feature)
export NOPHER_NSEC="nsec1x2y3z4..."
```

---

## protocols

Enable/disable protocol servers and configure ports.

```yaml
protocols:
  gopher:
    enabled: true
    host: "gopher.example.com"
    port: 70
    bind: "0.0.0.0"
  gemini:
    enabled: true
    host: "gemini.example.com"
    port: 1965
    bind: "0.0.0.0"
    tls:
      cert_path: "./certs/cert.pem"
      key_path: "./certs/key.pem"
      auto_generate: true
  finger:
    enabled: true
    port: 79
    bind: "0.0.0.0"
    max_users: 100
```

### protocols.gopher

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable Gopher server |
| `host` | string | `localhost` | Hostname for gopher:// URLs |
| `port` | int | `70` | TCP port (RFC 1436 standard) |
| `bind` | string | `0.0.0.0` | Interface to bind to |

**Notes:**
- Port 70 requires root/sudo on most systems
- Use `127.0.0.1` to bind only to localhost
- `host` used in gophermap links

### protocols.gemini

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable Gemini server |
| `host` | string | `localhost` | Hostname for gemini:// URLs |
| `port` | int | `1965` | TLS port (Gemini standard) |
| `bind` | string | `0.0.0.0` | Interface to bind to |
| `tls.cert_path` | string | `./certs/cert.pem` | Path to TLS certificate |
| `tls.key_path` | string | `./certs/key.pem` | Path to TLS private key |
| `tls.auto_generate` | bool | `true` | Generate self-signed cert if missing |

**TLS Certificates:**
- If `auto_generate: true` and cert files missing, creates self-signed cert
- For production, use proper TLS cert (Let's Encrypt, etc.)
- Self-signed certs require TOFU (Trust On First Use) in Gemini clients

**Generate cert manually:**
```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem \
  -out certs/cert.pem -days 365 -nodes -subj "/CN=gemini.example.com"
```

### protocols.finger

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable Finger server |
| `port` | int | `79` | TCP port (RFC 742 standard) |
| `bind` | string | `0.0.0.0` | Interface to bind to |
| `max_users` | int | `100` | Max users queryable (owner + followed) |

**Notes:**
- Port 79 requires root/sudo
- `max_users` limits which followed users are fingerable

---

## relays

Seed relays and connection policies.

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
```

### relays.seeds

**Type:** Array of strings (WebSocket URLs)

Seed relays used for initial relay discovery. After startup, Nopher discovers additional relays via NIP-65 (kind 10002).

**Requirements:**
- Must start with `wss://` (TLS) or `ws://` (unencrypted)
- At least one seed required
- Choose reliable, well-connected relays

**Popular seed relays:**
```yaml
seeds:
  - "wss://relay.damus.io"        # Large, reliable
  - "wss://relay.nostr.band"      # Aggregator, good coverage
  - "wss://nos.lol"               # Fast, popular
  - "wss://relay.snort.social"    # Well-maintained
  - "wss://nostr.wine"            # Paid, quality
```

### relays.policy

Connection behavior and limits.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `connect_timeout_ms` | int | `5000` | Connection timeout (milliseconds) |
| `max_concurrent_subs` | int | `8` | Max concurrent subscriptions per relay |
| `backoff_ms` | int[] | `[500, 1500, 5000]` | Retry backoff schedule (ms) |

**Backoff behavior:**
- First retry: 500ms delay
- Second retry: 1500ms delay
- Third+ retry: 5000ms delay
- Prevents hammering unavailable relays

---

## discovery

Relay discovery settings using NIP-65 (kind 10002).

```yaml
discovery:
  refresh_seconds: 900
  use_owner_hints: true
  use_author_hints: true
  fallback_to_seeds: true
  max_relays_per_author: 8
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `refresh_seconds` | int | `900` | How often to refresh kind 10002 (15 min) |
| `use_owner_hints` | bool | `true` | Use owner's relay hints for owner data |
| `use_author_hints` | bool | `true` | Use authors' relay hints for their data |
| `fallback_to_seeds` | bool | `true` | Use seeds if hints missing/stale |
| `max_relays_per_author` | int | `8` | Safety cap per author |

**How it works:**
1. Fetch kind 10002 from seed relays (owner + followed users)
2. Parse relay hints (read/write tags)
3. Connect to discovered relays for targeted queries
4. Refresh periodically to catch relay changes

**Status:** 🟡 IMPLEMENTED (code in internal/nostr/discovery.go)

---

## sync

Event synchronization scope and retention.

```yaml
sync:
  enabled: true  # Enable/disable sync engine
  kinds: [0, 1, 3, 6, 7, 9735, 30023, 10002]
  scope:
    mode: "foaf"
    depth: 2
    include_direct_mentions: true
    include_threads_of_mine: true
    max_authors: 5000
    allowlist_pubkeys: []
    denylist_pubkeys: []
  retention:
    keep_days: 365
    prune_on_start: true
```

### sync.enabled

**Type:** Boolean
**Default:** `true`

Enable or disable the sync engine.

```yaml
sync:
  enabled: true   # Sync engine runs, pulls events from relays
  # enabled: false  # Sync engine disabled, no new events synced
```

**When disabled:**
- No events are synced from remote relays
- Only serves existing events from database
- Useful for read-only deployments or maintenance

**When enabled:**
- Sync engine starts and connects to relays
- Events are pulled based on scope configuration
- Relay discovery runs periodically

**Status:** ✅ VERIFIED (integrated in main.go)

### sync.kinds

**Type:** Array of integers

Nostr event kinds to synchronize.

| Kind | Description | Purpose |
|------|-------------|---------|
| `0` | Profile (metadata) | User info, names, avatars |
| `1` | Short note | Text posts |
| `3` | Contacts (follows) | Social graph |
| `6` | Repost | Shares/boosts |
| `7` | Reaction | Likes, emoji reactions |
| `9735` | Zap receipt | Lightning tips |
| `30023` | Long-form article | Blog posts |
| `10002` | Relay hints (NIP-65) | Relay discovery |

**Add more kinds:**
```yaml
kinds: [0, 1, 3, 6, 7, 9735, 30023, 10002, 30311]  # Add NIP-89 app handlers
```

### sync.scope

Controls whose events to sync.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `foaf` | Sync mode (see below) |
| `depth` | int | `2` | FOAF depth (when mode=foaf) |
| `include_direct_mentions` | bool | `true` | Include events mentioning you |
| `include_threads_of_mine` | bool | `true` | Include threads you participated in |
| `max_authors` | int | `5000` | Safety cap on total authors |
| `allowlist_pubkeys` | string[] | `[]` | Always include these pubkeys |
| `denylist_pubkeys` | string[] | `[]` | Never include these pubkeys |

**Sync modes:**

| Mode | Description | Authors Synced |
|------|-------------|----------------|
| `self` | Only your events | 1 (you) |
| `following` | You + who you follow | 1 + kind 3 count |
| `mutual` | You + mutual follows | 1 + bidirectional follows |
| `foaf` | Friend-of-a-friend | Grows exponentially by depth |

**FOAF depth examples:**
- `depth: 1` = You + following (same as `following` mode)
- `depth: 2` = You + following + their follows (2nd degree)
- `depth: 3` = 3rd degree connections (can be huge!)

**Recommendations:**
- Start with `mode: following` or `mode: mutual`
- Use `mode: foaf` with `depth: 2` cautiously (may sync thousands)
- Set `max_authors` to prevent runaway sync
- Use `denylist_pubkeys` for spam accounts

### sync.retention

Data retention and pruning.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `keep_days` | int | `365` | Keep events newer than N days |
| `prune_on_start` | bool | `true` | Prune old events at startup |

**Pruning behavior:**
- Events older than `keep_days` are deleted
- Kind 0 (profiles) and kind 3 (follows) never pruned
- Replaceable events (kind 10002, 30023) keep only latest

**Status:** ✅ VERIFIED (code in internal/sync/)

### sync.retention.advanced (Phase 17)

**Advanced configurable retention system** - sophisticated, multi-dimensional retention rules.

```yaml
sync:
  retention:
    keep_days: 365
    prune_on_start: true

    advanced:
      enabled: false              # Must explicitly enable
      mode: "rules"               # rules|caps

      evaluation:
        on_ingest: true           # Evaluate new events immediately
        re_eval_interval_hours: 168  # Re-evaluate weekly
        batch_size: 1000

      global_caps:
        max_total_events: 1000000
        max_storage_mb: 5000
        max_events_per_kind:
          1: 100000               # Max 100k notes
          30023: 10000            # Max 10k articles

      rules:
        - name: "protect_owner"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true          # Never delete

        - name: "close_network"
          priority: 800
          conditions:
            social_distance_max: 1
            kinds: [1, 30023]
          action:
            retain_days: 365

        - name: "default"
          priority: 100
          conditions:
            all: true
          action:
            retain_days: 90
```

**Key features:**

| Feature | Description |
|---------|-------------|
| **Rule-based** | Define retention rules with conditions and priorities |
| **Multi-dimensional** | Filter by kind, author, social distance, interaction count, etc. |
| **Cap enforcement** | Hard limits on total events, storage, per-kind counts |
| **Score-based pruning** | When caps exceeded, delete lowest-scored events first |
| **Protected events** | Mark events that should never be deleted |
| **Incremental evaluation** | Evaluate on ingestion + periodic re-evaluation |

**Condition types (gates):**
- `author_is_owner` - Event is from owner
- `social_distance_max` - FOAF distance ≤ N
- `kinds` - Event kind matches list
- `min_interactions` - Has at least N replies/reactions/zaps
- `age_days_max` - Event age ≤ N days
- `content_length_min` - Content ≥ N chars
- `is_thread_root` - Is root of thread
- `has_replies` - Has at least one reply

**Action types:**
- `retain: true` - Never delete (protected)
- `retain_days: N` - Keep for N days
- `retain: false` - Eligible for deletion

**Priority:**
- Higher priority rules match first
- If multiple rules match, highest priority wins
- Default rule (lowest priority) catches all

**Backward compatibility:**
- If `advanced.enabled: false`, uses simple `keep_days` only
- Invalid advanced config falls back to simple mode with warning
- Simple mode remains fully functional

**See also:** [memory/PHASE_17_RETENTION.md](../memory/PHASE_17_RETENTION.md) for complete specification

**Status:** 📋 PLANNED (Phase 17)

---

## inbox

Inbox aggregation of replies, reactions, and zaps.

```yaml
inbox:
  include_replies: true
  include_reactions: true
  include_zaps: true
  group_by_thread: true
  collapse_reposts: true
  noise_filters:
    min_zap_sats: 1
    allowed_reaction_chars: ["+"]
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `include_replies` | bool | `true` | Show replies to your notes |
| `include_reactions` | bool | `true` | Show kind 7 reactions |
| `include_zaps` | bool | `true` | Show kind 9735 zaps |
| `group_by_thread` | bool | `true` | Group inbox by thread root |
| `collapse_reposts` | bool | `true` | Collapse multiple reposts |
| `noise_filters.min_zap_sats` | int | `1` | Minimum zap amount to show |
| `noise_filters.allowed_reaction_chars` | string[] | `["+"]` | Filter reactions (e.g., only "+") |

**Noise filtering:**
- Filter out tiny zaps: `min_zap_sats: 100` (0.1 sat minimum)
- Allow only specific reactions: `allowed_reaction_chars: ["+", "❤️", "🔥"]`
- Prevent spam/unwanted reactions

**Status:** ✅ VERIFIED (aggregates code tested in internal/aggregates/)

---

## outbox

Publishing settings (future feature).

```yaml
outbox:
  publish:
    notes: true
    reactions: false
    zaps: false
  draft_dir: "./content"
  auto_sign: false
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publish.notes` | bool | `true` | Publish notes (kind 1) |
| `publish.reactions` | bool | `false` | Publish reactions (kind 7) |
| `publish.zaps` | bool | `false` | Publish zaps (kind 9735) |
| `draft_dir` | string | `./content` | Directory for draft notes |
| `auto_sign` | bool | `false` | Auto-sign with nsec |

**Status:** 📋 PLANNED (Phase 13)

---

## storage

Database backend configuration.

```yaml
storage:
  driver: "sqlite"
  sqlite_path: "./data/nopher.db"
  lmdb_path: "./data/nopher.lmdb"
  lmdb_max_size_mb: 10240
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `driver` | string | `sqlite` | Database backend (`sqlite` or `lmdb`) |
| `sqlite_path` | string | `./data/nopher.db` | SQLite database file path |
| `lmdb_path` | string | `./data/nopher.lmdb` | LMDB database directory |
| `lmdb_max_size_mb` | int | `10240` | LMDB max size (MB) - 10GB default |

**Choosing a backend:**

| Feature | SQLite | LMDB |
|---------|--------|------|
| File format | Single .db file | Directory with data files |
| Setup | Zero config | Zero config |
| Performance | Good for <100K events | Excellent for millions |
| Concurrency | Limited writes | Excellent |
| Backups | Copy .db file | Copy directory |
| Best for | Personal use | High-volume streaming |

**Recommendations:**
- **Start with SQLite** - simpler, sufficient for most users
- **Switch to LMDB** if you sync >100K events or need high write throughput
- Both use Khatru eventstore interface (see [storage.md](storage.md))

**Status:** ✅ VERIFIED (both backends implemented in internal/storage/)

---

## rendering

Protocol-specific rendering options.

```yaml
rendering:
  gopher:
    max_line_length: 70
    show_timestamps: true
    date_format: "2006-01-02 15:04 MST"
    thread_indent: "  "
  gemini:
    max_line_length: 80
    show_timestamps: true
    emoji: true
  finger:
    plan_source: "kind_0"
    recent_notes_count: 5
```

### rendering.gopher

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_line_length` | int | `70` | Wrap text at N characters |
| `show_timestamps` | bool | `true` | Show event timestamps |
| `date_format` | string | `2006-01-02 15:04 MST` | Go time format string |
| `thread_indent` | string | `"  "` | Indent string for replies |

**Gopher conventions:**
- 70 chars is traditional (old terminal width)
- Plain ASCII, no ANSI colors
- Minimal formatting

### rendering.gemini

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_line_length` | int | `80` | Wrap text at N characters |
| `show_timestamps` | bool | `true` | Show event timestamps |
| `emoji` | bool | `true` | Allow emoji in gemtext |

**Gemini conventions:**
- 80 chars common but not required
- UTF-8 supported (emoji OK)
- Gemtext format (headings, links, quotes)

### rendering.finger

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plan_source` | string | `kind_0` | Source for .plan field (`kind_0` or `kind_1`) |
| `recent_notes_count` | int | `5` | Number of recent notes to show |

**Plan source:**
- `kind_0`: Use profile "about" field as .plan
- `kind_1`: Use most recent note as .plan

**Status:** 🟡 IMPLEMENTED (code in internal/markdown/, internal/gopher/, etc.)

---

## caching

Response caching configuration for dramatic performance improvements.

```yaml
caching:
  enabled: true
  engine: "memory"  # or "redis"
  redis_url: ""  # Set via NOPHER_REDIS_URL env var
  max_size_mb: 100  # Memory cache size limit
  default_ttl_seconds: 300
  cleanup_interval_seconds: 60

  ttl:
    sections:
      notes: 60        # 1 minute
      comments: 30     # 30 seconds
      articles: 300    # 5 minutes
      interactions: 10 # 10 seconds
    render:
      gopher_menu: 300      # 5 minutes
      gemini_page: 300      # 5 minutes
      finger_response: 60   # 1 minute
      kind_1: 86400         # 24 hours
      kind_30023: 604800    # 7 days
      kind_0: 3600          # 1 hour
      kind_3: 600           # 10 minutes

  aggregates:
    enabled: true
    update_on_ingest: true
    reconciler_interval_seconds: 900  # 15 minutes
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Master switch for caching |
| `engine` | string | `memory` | Cache backend (`memory` or `redis`) |
| `redis_url` | string | `""` | Redis URL (via `NOPHER_REDIS_URL` env) |
| `max_size_mb` | int | `100` | Memory cache size limit (MB) |
| `default_ttl_seconds` | int | `300` | Default cache TTL (5 minutes) |
| `cleanup_interval_seconds` | int | `60` | Expired entry cleanup interval |
| `ttl.sections.*` | int | varies | Section cache TTLs (seconds) |
| `ttl.render.*` | int | varies | Render cache TTLs (seconds) |
| `aggregates.enabled` | bool | `true` | Cache aggregate computations |
| `aggregates.update_on_ingest` | bool | `true` | Update on new events |
| `aggregates.reconciler_interval_seconds` | int | `900` | Reconcile drift (15 min) |

### caching.enabled

**Type:** Boolean
**Default:** `true`

Enable or disable the caching layer.

```yaml
caching:
  enabled: true   # Caching active, responses cached
  # enabled: false  # No caching, always regenerate responses
```

**Performance Impact:**
- **Enabled**: 10-100x faster responses, 80-95% reduction in database queries
- **Disabled**: Always regenerates responses, higher latency and CPU usage

### caching.engine

**Type:** String (`memory` or `redis`)
**Default:** `memory`

Cache backend engine.

**Memory Cache:**
```yaml
caching:
  engine: "memory"
  max_size_mb: 100
```

- Thread-safe in-memory cache
- LRU eviction when size limit reached
- Automatic cleanup of expired entries
- Best for single-instance deployments
- No external dependencies

**Redis Cache:**
```yaml
caching:
  engine: "redis"
  redis_url: "redis://localhost:6379/0"
```

- Distributed cache across multiple instances
- Persistent across restarts
- Better memory management
- Built-in clustering support
- Requires external Redis server

**When to use Redis:**
- Running multiple Nopher instances
- Need persistent cache across restarts
- Limited memory on host
- Want shared cache for load balancing

### Cache Invalidation

Cache entries are automatically invalidated when relevant events are ingested:

| Event Kind | Invalidates |
|------------|-------------|
| Kind 0 (Profile) | Profile cache, kind0 cache |
| Kind 1 (Note) | Notes section cache |
| Kind 3 (Contacts) | Kind3 cache |
| Kind 7 (Reaction) | Parent event aggregates |
| Kind 9735 (Zap) | Parent event aggregates |

**Manual Invalidation:**
Cache is cleared when:
- Configuration changes
- Sync scope changes
- Manual server restart

### Cache Keys

Cache uses hierarchical keys:

```
gopher:/path/to/selector        - Gopher response
gemini:/path?query=test         - Gemini response
finger:username                 - Finger response
event:event123:gopher:text      - Event rendering
section:notes:gemini:p2         - Section page
thread:root123:gopher           - Thread rendering
profile:pubkey123:gemini        - Profile page
aggregate:event123              - Interaction counts
kind0:pubkey123                 - Profile metadata
kind3:pubkey123                 - Contact list
```

**Pattern Matching** (for bulk operations):
```
gopher:*                  - All Gopher responses
event:event123:*          - All renderings of event
profile:pubkey123:*       - All profile renderings
```

### TTL Strategy

**Short TTL (10-60s):** Live/changing content
- Interactions, inbox, recent notes

**Medium TTL (300-600s):** Semi-static content
- Sections, menus, navigation

**Long TTL (hours/days):** Immutable content
- Old events, profiles, articles

### Cache Statistics

Monitor cache performance:

```
Cache Statistics:
  Hits: 950
  Misses: 50
  Hit Rate: 95%
  Keys: 150
  Size: 12.3 MB / 100 MB
  Evictions: 5
  Avg Get Time: 0.3ms
  Avg Set Time: 0.5ms
```

**Target Metrics:**
- Hit Rate: > 80%
- Avg Get Time: < 1ms (memory), < 5ms (Redis)
- Evictions: Low (increase max_size_mb if high)

### Redis Configuration

**Environment Variable:**
```bash
export NOPHER_REDIS_URL="redis://localhost:6379/0"
```

**Redis URL Format:**
```
redis://[user:password@]host:port[/database]
```

**Examples:**
```bash
# Local Redis, no auth
export NOPHER_REDIS_URL="redis://localhost:6379/0"

# Remote Redis with password
export NOPHER_REDIS_URL="redis://:mypassword@redis.example.com:6379/0"

# Redis with username and password
export NOPHER_REDIS_URL="redis://user:pass@redis.example.com:6379/0"
```

**See also:** [deployment.md](deployment.md#redis-setup) for Redis installation and configuration.

**Status:** ✅ VERIFIED (Phase 10 complete - implemented in internal/cache/)

---

## logging

Logging configuration.

```yaml
logging:
  level: "info"  # debug|info|warn|error
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | `info` | Log level |

**Log levels:**
- `debug`: Verbose (all events, queries, connections)
- `info`: Normal (startup, errors, important events)
- `warn`: Warnings only
- `error`: Errors only

**Example:**
```bash
# Debug mode for troubleshooting
NOPHER_LOG_LEVEL=debug nopher --config nopher.yaml
```

**Status:** ✅ VERIFIED (validated in internal/config/config.go)

---

## layout

Custom sections and page layouts for organizing and presenting content.

```yaml
layout:
  sections:
    notes:
      title: "Recent Notes"
      description: "Latest short-form posts"
      filters:
        kinds: [1]
        limit: 20
      sort_by: "created_at"
      sort_order: "desc"
      show_dates: true
      show_authors: true

    articles:
      title: "Articles"
      description: "Long-form content"
      filters:
        kinds: [30023]
        limit: 10
      sort_by: "published_at"
      sort_order: "desc"

    inbox:
      title: "Inbox"
      description: "Your interactions"
      filters:
        tags:
          p: ["${OWNER_PUBKEY}"]
        kinds: [1, 7, 9735]
        limit: 50
      sort_by: "created_at"
      sort_order: "desc"
```

### layout.sections

Define custom sections for organizing events by kind, author, tags, time range, and other criteria.

**Section structure:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | - | Display title for the section |
| `description` | string | - | Section description |
| `filters` | object | - | Filter criteria (see below) |
| `sort_by` | string | `created_at` | Sort field: `created_at`, `reactions`, `zaps`, `replies` |
| `sort_order` | string | `desc` | Sort order: `asc` or `desc` |
| `limit` | int | `20` | Events per page |
| `show_dates` | bool | `true` | Display event timestamps |
| `show_authors` | bool | `true` | Display author names |
| `group_by` | string | - | Grouping: `day`, `week`, `month`, `year`, `author`, `kind` |

**Filter options:**

```yaml
filters:
  kinds: [1, 30023]                    # Event kinds to include
  authors: ["pubkey1", "pubkey2"]      # Filter by pubkeys
  tags:
    p: ["pubkey"]                      # Filter by tags (p, e, etc.)
    e: ["eventid"]
  since: 1704067200                    # Unix timestamp (start)
  until: 1735689600                    # Unix timestamp (end)
  limit: 20                            # Max events
  scope: "following"                   # self, following, mutual, foaf, all
```

**Default sections:**

Nopher provides built-in sections:
- **notes**: Short-form notes (kind 1)
- **articles**: Long-form articles (kind 30023)
- **reactions**: Reactions (kind 7)
- **zaps**: Zap receipts (kind 9735)
- **inbox**: Mentions and replies to owner
- **outbox**: Owner's published content

**Time-based filtering:**

```yaml
filters:
  kinds: [1]
  since: ${LAST_7_DAYS}    # Built-in time ranges
  # LAST_7_DAYS, LAST_30_DAYS, THIS_WEEK, THIS_MONTH, THIS_YEAR
```

**Archive support:**

Sections automatically generate time-based archives:
- `/archive/notes/2025/10` - October 2025 notes
- `/archive/articles/2025` - All 2025 articles
- Monthly calendar views with event counts

### layout.sections Examples

**Popular posts section:**
```yaml
sections:
  popular:
    title: "Popular Posts"
    description: "Most liked and zapped"
    filters:
      kinds: [1, 30023]
      limit: 10
    sort_by: "reactions"
    sort_order: "desc"
```

**Recent from following:**
```yaml
sections:
  timeline:
    title: "Timeline"
    description: "Recent posts from people you follow"
    filters:
      kinds: [1]
      scope: "following"
      limit: 50
    sort_by: "created_at"
    sort_order: "desc"
```

**Thread view:**
```yaml
sections:
  thread:
    title: "Thread"
    description: "Conversation thread"
    filters:
      tags:
        e: ["${ROOT_EVENT_ID}"]
      kinds: [1]
    sort_by: "created_at"
    sort_order: "asc"
    group_by: "day"
```

**Monthly archive:**
```yaml
sections:
  monthly:
    title: "This Month"
    description: "Posts from this month"
    filters:
      kinds: [1]
      since: ${THIS_MONTH_START}
      until: ${THIS_MONTH_END}
    sort_by: "created_at"
    sort_order: "desc"
```

### Pagination

Sections support automatic pagination:
- Page 1: `/section/notes`
- Page 2: `/section/notes/2`
- Page 3: `/section/notes/3`

Navigation includes:
- Previous/Next page links
- Page numbers
- Total pages and items

### Archives

Sections generate archives automatically:
- List archives: `/archive/notes`
- Monthly view: `/archive/notes/2025/10`
- Daily view: `/archive/notes/2025/10/24`
- Calendar: Monthly calendar with event counts per day

**Status:** ✅ VERIFIED (Phase 11 complete - implemented in internal/sections/)

---

## security

Security features including deny lists, rate limiting, input validation, and content filtering.

```yaml
security:
  # Deny list configuration
  denylist:
    enabled: true
    pubkeys:
      - "deadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678"
      - "cafebabe1234567890abcdef1234567890abcdef1234567890abcdef12345678"

  # Content filtering
  content_filter:
    enabled: true
    banned_words:
      - "spam"
      - "scam"
    case_sensitive: false

  # Rate limiting
  ratelimit:
    enabled: true
    global:
      requests_per_minute: 60
      burst_size: 10
    per_protocol:
      gopher:
        requests_per_minute: 30
        burst_size: 5
      gemini:
        requests_per_minute: 60
        burst_size: 10
      finger:
        requests_per_minute: 20
        burst_size: 3

  # Input validation
  validation:
    enabled: true
    max_selector_length: 1024
    max_query_length: 2048
    max_path_length: 2048
    strict_mode: true

  # Security policy
  policy:
    allow_anonymous: true
    require_nip05: false
    block_tor: false
    block_vpn: false
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `denylist.enabled` | bool | `true` | Enable deny list filtering |
| `denylist.pubkeys` | []string | `[]` | Blocked pubkeys (hex format) |
| `content_filter.enabled` | bool | `true` | Enable content filtering |
| `content_filter.banned_words` | []string | `[]` | List of banned words |
| `content_filter.case_sensitive` | bool | `false` | Case-sensitive matching |
| `ratelimit.enabled` | bool | `true` | Enable rate limiting |
| `ratelimit.global.requests_per_minute` | int | `60` | Global rate limit |
| `ratelimit.global.burst_size` | int | `10` | Burst allowance |
| `ratelimit.per_protocol.*` | object | - | Per-protocol rate limits |
| `validation.enabled` | bool | `true` | Enable input validation |
| `validation.max_selector_length` | int | `1024` | Max Gopher selector length |
| `validation.max_query_length` | int | `2048` | Max Gemini query length |
| `validation.max_path_length` | int | `2048` | Max path length |
| `validation.strict_mode` | bool | `true` | Strict validation mode |
| `policy.allow_anonymous` | bool | `true` | Allow anonymous access |
| `policy.require_nip05` | bool | `false` | Require NIP-05 verification |

### security.denylist

Block specific Nostr pubkeys from appearing in your gateway.

**Pubkey format:** Full 64-character hex pubkey (not npub)

**Usage:**
```yaml
denylist:
  enabled: true
  pubkeys:
    - "deadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678"
```

**Dynamic management:**
- Pubkeys can be added/removed at runtime
- Thread-safe concurrent access
- Applies to all protocol servers

**What gets blocked:**
- Events authored by blocked pubkeys
- Profile information
- All interactions (replies, reactions, zaps) from blocked pubkeys

### security.content_filter

Filter events based on content patterns and banned words.

**Configuration:**
```yaml
content_filter:
  enabled: true
  banned_words:
    - "spam"
    - "phishing"
    - "scam"
  case_sensitive: false
```

**Behavior:**
- Checks event content for banned words
- Can be case-sensitive or insensitive
- Combines with deny list for comprehensive filtering
- Does not modify content, only filters visibility

### security.ratelimit

Prevent abuse with token bucket rate limiting.

**Algorithm:**
- Each client gets a bucket with N tokens
- Each request consumes 1 token
- Tokens refill over time (requests_per_minute / 60 per second)
- When bucket empty, requests are denied until refill

**Global rate limit:**
```yaml
ratelimit:
  enabled: true
  global:
    requests_per_minute: 60  # 1 request per second average
    burst_size: 10           # Allow bursts up to 10 requests
```

**Per-protocol limits:**
```yaml
ratelimit:
  per_protocol:
    gopher:
      requests_per_minute: 30  # Slower for Gopher
    gemini:
      requests_per_minute: 60  # Normal for Gemini
    finger:
      requests_per_minute: 20  # Slowest for Finger
```

**Client identification:**
- By IP address
- Shared across protocols unless per-protocol limits set
- Old client buckets automatically cleaned up

**Response when limited:**
- Gopher: Returns error message
- Gemini: Returns 44 status (slow down)
- Finger: Closes connection

### security.validation

Validate and sanitize all user input to prevent injection attacks.

**Protections against:**
- **CRLF injection**: `\r\n` characters removed
- **Null byte injection**: `\x00` characters removed
- **Directory traversal**: `..` sequences blocked
- **XSS attacks**: Script tags and dangerous HTML removed
- **Length limits**: Enforces maximum input lengths

**Strict mode:**
```yaml
validation:
  enabled: true
  strict_mode: true  # Reject invalid input
  # strict_mode: false  # Sanitize invalid input
```

**Strict mode true:** Rejects requests with invalid characters
**Strict mode false:** Attempts to sanitize and continues

**What gets validated:**
- Gopher selectors
- Gemini paths and queries
- Finger usernames
- Pubkeys and event IDs
- URLs and references

### security.policy

Security policy settings.

```yaml
policy:
  allow_anonymous: true      # Allow access without authentication
  require_nip05: false       # Require NIP-05 verification
  block_tor: false           # Block Tor exit nodes
  block_vpn: false           # Block known VPN IPs
```

**Note:** Authentication features are planned for future phases.

### Security Best Practices

1. **Enable all security features** in production
2. **Use strict validation mode** to catch attacks early
3. **Set appropriate rate limits** based on your capacity
4. **Regularly review deny list** for new abusive pubkeys
5. **Monitor logs** for suspicious activity
6. **Keep banned words list updated** for your community standards
7. **Never commit secrets** to configuration files (use environment variables)

### Security Monitoring

Monitor these metrics:
- Rate limit hits per client
- Validation failures
- Deny list blocks
- Content filter matches

**See also:** [SECURITY.md](SECURITY.md) for comprehensive security guide

**Status:** ✅ VERIFIED (Phase 14 complete - implemented in internal/security/)

---

## Environment Variable Overrides

Any configuration value can be overridden with `NOPHER_*` environment variables.

**Important overrides:**

| Variable | Overrides | Example |
|----------|-----------|---------|
| `NOPHER_NSEC` | `identity.nsec` | `nsec1abc...` |
| `NOPHER_REDIS_URL` | `caching.redis_url` | `redis://localhost:6379` |

**Example:**
```bash
export NOPHER_NSEC="nsec1..."
export NOPHER_REDIS_URL="redis://localhost:6379"
nopher --config nopher.yaml
```

---

## Validation

Configuration is validated on startup. Common errors:

| Error | Fix |
|-------|-----|
| `identity.npub is required` | Set `identity.npub` in config |
| `identity.npub must start with 'npub1'` | Use valid npub (check format) |
| `at least one protocol must be enabled` | Enable at least one of gopher/gemini/finger |
| `port must be between 1 and 65535` | Fix port number |
| `relay seed must start with ws:// or wss://` | Fix relay URL format |
| `invalid sync mode` | Use: `self`, `following`, `mutual`, or `foaf` |
| `invalid storage driver` | Use: `sqlite` or `lmdb` |
| `invalid log level` | Use: `debug`, `info`, `warn`, or `error` |

**Validate config without starting servers:**
```bash
nopher --config nopher.yaml --validate  # (future feature)
```

For now, config is validated on startup. Check output for validation errors.

---

## Complete Example

See [configs/nopher.example.yaml](../configs/nopher.example.yaml) for a complete, commented example configuration.

---

**Next:** [Storage Guide](storage.md) | [Protocol Servers](protocols.md) | [Getting Started](getting-started.md)
