Realtime Updates

Transport
- Server-Sent Events (SSE) primary; long-lived GET endpoints that stream JSON lines.
- Polling fallback when SSE disabled or not supported.

Endpoints (examples)
- GET /live/inbox
  - Streams interactions targeting the owner: reactions, zaps, and replies metadata.
- GET /live/thread/{event_id}
  - Streams new replies to the given thread root.
- GET /live/aggregates?ids={comma-separated-event-ids}
  - Streams aggregate deltas for listed events (reply_count, reaction_total, zap_sats_total).

Event Types (payloads)
- new_reply
  - { type: "new_reply", root: "<event_id>", reply: { id, pubkey, created_at, preview } }
- new_reaction
  - { type: "new_reaction", target: "<event_id>", char: "+", pubkey, created_at }
- new_zap
  - { type: "new_zap", target: "<event_id>"|"profile", sats: 21, pubkey, created_at }
- aggregate_update
  - { type: "aggregate_update", id: "<event_id>", reply_count, reaction_total, zap_sats_total }
- heartbeat
  - { type: "heartbeat", now }

Client Behavior
- Open SSE on page load for relevant channels (inbox, visible threads).
- Merge updates into the DOM; bump counters; append items to interaction lists.
- Backoff and retry on disconnect; show subtle "Live paused" indicator if offline.

Server Behavior
- Authenticate channels only if needed (single-tenant public site usually not required).
- Respect caching: base pages may be cached; SSE provides freshness without invalidating caches.
- Heartbeat at configured interval to keep connections alive and detect dead clients.

Limits and Safety
- Cap max connected clients; drop idle connections.
- Deduplicate events across relays before emitting updates.
- Rate-limit outbound updates per client to avoid floods during spikes.
