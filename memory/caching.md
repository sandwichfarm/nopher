Caching Strategy

Goals
- Speed up page and feed rendering while keeping interactions (replies, reactions, zaps) feeling live.
- Avoid heavy recomputation for common queries (sections) and expensive transforms (markdown rendering).
- Keep behavior predictable under eventual consistency across relays.

Data Freshness Classes
- Stable content
  - Notes (kind 1): immutable; safe to cache aggressively once indexed.
  - Articles (kind 30023): parameterized replaceable by d-tag; treat as versions per d; cache, invalidate on newer same-d.
- Semi-stable metadata
  - Profile (kind 0), Contacts (kind 3), Relay hints (kind 10002): rarely change; long TTL; refresh periodically and on replace.
- Highly dynamic interactions
  - Replies (kind 1 replies), Reactions (kind 7), Zaps (9735): short TTL, prefer live updates via SSE; cache counts as aggregates.

Layers
- Authoritative store (DB)
  - SQLite tables hold events, refs, relay hints, graph, sync state, and aggregates.
- Aggregates (persistent derived state)
  - Maintain counters per event: reply_count, reaction_count_total, reaction_count_by_char, zap_sats_total, last_interaction_at.
  - Update on ingest; periodic reconcile jobs for missed updates; cheap to query per page.
- In-memory cache (default) / Redis (optional)
  - Render cache: pre-rendered content for events (e.g., markdown to HTML for 30023/1).
  - Section results: lists of event ids for section queries (home page slots, archives pages).
  - Feed payloads: serialized RSS/JSON feeds per section.
- HTTP cache
  - ETag/Last-Modified on dynamic endpoints; Cache-Control with max-age and stale-while-revalidate.
  - Static export served with long-lived immutable caching and hashed filenames.

Invalidation and Refresh
- Event ingest
  - Stable content: invalidate render cache for that event id; if section queries affected (owner outbox), invalidate those section keys.
  - Interactions: bump aggregates and push live updates; no need to invalidate base page cache immediately.
- Replaceable kinds
  - Kind 0/3/10002: invalidate profile/contact/relay-hints dependent caches on replacement.
  - Kind 30023: invalidate render cache and section entries for that d-key.
- Config/theme changes
  - Version cache namespace by config+theme hash; any change flips namespace and avoids stale mix.
- Time-based expiry
  - Section TTLs (e.g., notes 60s, articles 300s, interactions 10s) trigger background revalidation (SWR) on access.

Live Updates (overlay pattern)
- Serve cached HTML for pages and sections quickly.
- Open SSE channel (or polling fallback) to stream interaction deltas and updated aggregates.
- Update visible counters (replies, reactions, zaps) and append new interaction items client-side.
- Thread pages use the same overlay to append new replies.

Suggested Defaults (overridable)
- Section TTLs
  - notes: 60s
  - comments: 30s
  - articles: 300s
  - interactions: 10s (or use SSE only)
- Render cache TTLs
  - kind 1: 24h
  - kind 30023: 7d (invalidate on replacement)
  - kind 0: 1h
  - kind 3: 10m
- Feeds TTL: 5m with ETag; SWR 15m

Keys and Namespacing
- Cache key format
  - render:event:{event_id}
  - render:param:{kind}:{pubkey}:{d}
  - section:{section_id}:{page}:{query_hash}
  - feed:{section_id}:{format}
  - config_version: computed short hash of relevant config and theme settings
- Namespacing: prefix all keys with v:{config_version}: to isolate caches per configuration.

Aggregates Table (overview)
- aggregates
  - event_id TEXT PRIMARY KEY
  - reply_count INTEGER
  - reaction_total INTEGER
  - reaction_counts_json TEXT   # map char->count
  - zap_sats_total INTEGER
  - last_interaction_at INTEGER
- Updated on ingest; reconciler verifies counts periodically.

Section Query Cache
- Store results as ordered event id arrays per section+page.
- On ingest of owner-authored content, invalidate only affected sections (e.g., notes, articles).
- On interactions, keep section caches; overlay will refresh visible counters and append to interactions section.

HTTP Semantics
- ETag generated from section result hash + aggregates versions.
- Last-Modified derived from max(created_at) among included items.
- Cache-Control: max-age per section; stale-while-revalidate for smoothness.

Redis Option
- For higher traffic or multiple instances, swap in Redis for shared cache.
- Configurable via caching.engine and REDIS_URL in env.

Static Export + Live Overlay
- Export pages with embedded section payloads (frozen at export time).
- JS optionally connects to SSE endpoints to show fresh interactions without full regen.
- If strict static desired, disable live overlay; interactions will be as-of export time.

Failure Modes and Safety
- On cache miss or cache store failure, fall back to DB queries.
- TTLs prevent permanent staleness; aggregates ensure we do not compute heavy counts per request.
- Diagnostics surface hit/miss rates and invalidation counts.

Open Questions
- How aggressively to pre-render markdown for kind 1 vs on-demand?
- Expose per-section cache controls in the UI? (e.g., disable cache for a given section id)
- CDN integration: optional surrogate keys for fast invalidation per section.
