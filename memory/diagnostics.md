Diagnostics and Troubleshooting

Quick Triage (order)
- Config load: identity.npub present; relays.seeds non-empty; discovery.use_* flags as intended; scope/caps sane.
- Seed reachability: all seeds connect; can fetch owner's 0/3/10002 within connect_timeout.
- Owner hints: relay_hints rows exist for owner; freshness within last 30d (tunable).
- Author set size: matches expectation for chosen scope; below max_authors cap.
- Active reads: connected read relays > 0; per-relay last_event_at recent; cursors advancing.

Discovery Checks
- Seeds connectivity
  - Confirm WebSocket handshake success and SUB/EOSE for kinds 0/3/10002.
  - Measure latency; log connect->EOSE time; flag >5s as slow.
- Owner bootstrap
  - Latest kind 0/3 present; 10002 exists? If missing, proceed with seeds; schedule rediscovery.
  - Log owner 10002 freshness (created_at) and relay count split read/write.
- Author hints coverage
  - Percentage of in-scope authors with fresh 10002; target >60% (varies by network health).
  - For authors without hints, confirm fallback to seeds and track retry schedule.
- Hint correctness
  - Deduplicate relays; cap to discovery.max_relays_per_author; strip invalid URLs.
  - Prefer wss; downgrade ws only if explicitly allowed.
  - Multiple 10002 events: select highest created_at; store last_seen_event_id.

Sync Engine Checks
- Subscriptions
  - Per-relay filter count within policy.max_concurrent_subs; authors chunk size within relay message limits.
  - Kinds covered per config; replaceable kinds always included for refresh.
- Cursors
  - sync_state[relay,kind].since increases monotonically; no large regressions.
  - Detect future timestamps (clock skew); clamp and warn.
- Throughput and dedupe
  - Ingest rate (events/sec) within CPU/IO budget; duplicates ratio <30% (multi-relay overlap).
  - Signature verification failures near 0; log offenders.
- Replaceables
  - Periodic refresh interval respected (0/3/10002); verify profile/contact/relay updates reflect in UI within refresh window.

Aggregates and Caching
- Aggregates
  - Spot-check random 20 events: recompute reply/reaction/zap counts from refs and compare; delta should be 0.
  - last_interaction_at updates on new interactions; used in sorting interactions.
- Render/Section caches
  - config hash present in key prefixes (v:{cfg}); TTLs align with config.caching.ttl.* values.
  - Section cache hit rate acceptable (>70% for home); render cache hits improve after warmup.
- HTTP validators
  - ETag changes when aggregates change; 304 served when unchanged; Last-Modified aligns with max created_at.

Realtime (SSE)
- Connectivity
  - /live endpoints accept connections; heartbeat arrives at configured interval.
- Content
  - new_reply/reaction/zap events triggered on ingest; corresponding aggregate_update follows quickly.
- Limits
  - max_clients not exceeded; idle clients dropped; bandwidth within target.

Scope and Pruning
- Graph expansion produces expected counts; FOAF depth and mutual toggles work as intended.
- Pruning runs on schedule; reports pruned counts; no visible gaps beyond configured retention.

Common Issues and Fixes
- Seeds unreachable / very slow
  - Add/replace seeds with known watchers; increase connect_timeout_ms; check local firewall.
- No or stale 10002 for owner
  - Keep reading from seeds; publish/update 10002 from normal client; set discovery.fallback_to_seeds=true.
- FOAF explosion (huge author set)
  - Reduce depth; switch to mutual mode; lower max_authors; add denylist; disable include_threads_of_mine temporarily.
- Interactions missing live
  - SSE disabled or blocked by proxy; enable caching.live.sse; ensure reverse proxy supports HTTP/1.1 streaming and disables buffering.
- Reactions hidden unexpectedly
  - noise_filters.allowed_reaction_chars too strict; broaden or disable; check denylist.
- Zaps not counted
  - Missing amounts in receipts; ensure parsing for bolt11; note: cannot verify against LN node unless configured.
- Stuck cursors / no new events
  - Relay returns EOSE without events; since too new; reset since back by small window; verify clock skew.
- Duplicate or flickering replaceables
  - Multiple 0/3/10002 with out-of-order timestamps; always choose highest created_at; add tie-break by event id.

What to Capture in Bug Reports
- Config excerpt (site/relays/discovery/sync/caching/layout, redact secrets)
- Seeds list and which succeed/fail
- Owner 0/3/10002 event ids and created_at timestamps
- Relay hints coverage stats and sample rows
- Active relays list, cursors, last_event_at per relay
- Section cache hit/miss rates and ETag values observed
- SSE connectivity status and sample payloads
