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

**Status:** 🟡 IMPLEMENTED (code in internal/sync/)

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

Response caching configuration.

```yaml
caching:
  enabled: true
  engine: "memory"
  redis_url: ""
  ttl:
    sections:
      notes: 60
      comments: 30
      articles: 300
      interactions: 10
    render:
      gopher_menu: 300
      gemini_page: 300
      finger_response: 60
      kind_1: 86400
      kind_30023: 604800
      kind_0: 3600
      kind_3: 600
  aggregates:
    enabled: true
    update_on_ingest: true
    reconciler_interval_seconds: 900
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Master switch for caching |
| `engine` | string | `memory` | Cache backend (`memory` or `redis`) |
| `redis_url` | string | `""` | Redis URL (via `NOPHER_REDIS_URL` env) |
| `ttl.sections.*` | int | varies | Section cache TTLs (seconds) |
| `ttl.render.*` | int | varies | Render cache TTLs (seconds) |
| `aggregates.enabled` | bool | `true` | Cache aggregate computations |
| `aggregates.update_on_ingest` | bool | `true` | Update on new events |
| `aggregates.reconciler_interval_seconds` | int | `900` | Reconcile drift (15 min) |

**TTL recommendations:**
- Short TTL (10-60s): Live/changing content (interactions, inbox)
- Medium TTL (300-600s): Semi-static (sections, menus)
- Long TTL (hours/days): Immutable (old events, profiles)

**Redis:**
```bash
export NOPHER_REDIS_URL="redis://localhost:6379"
```

**Status:** 📋 PLANNED (Phase 10 - no cache code found yet)

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

Custom sections and page layouts.

```yaml
layout:
  sections: {}
  pages: {}
```

**Status:** 🚧 IN PROGRESS (Phase 11)

See [memory/layouts_sections.md](../memory/layouts_sections.md) for planned schema.

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
