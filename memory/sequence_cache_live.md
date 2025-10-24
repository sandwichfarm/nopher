Sequence: Cache and Live Updates

1) Page Request (Home with Notes & Articles)

Browser -> Server: GET /
Server -> Cache: GET v:{cfg}:section:notes:page:1:q:{q}
Cache -> Server: HIT? If miss, continue
Server -> DB: Query for notes per section filter
Server: Store section results in cache with TTL
Server -> Cache: GET v:{cfg}:render:event:{id}:t:{template_h} for visible ids (batch)
Cache -> Server: Return hits; misses are rendered now and stored
Server: Compute ETag from ids + aggregates; set Cache-Control/ETag/Last-Modified
Server -> Browser: 200 HTML (cached fragments + initial aggregates embedded)
Browser -> Server: Open SSE /live/inbox (if enabled)

2) Aggregates Overlay (on page load)
Browser -> Server: GET /live/aggregates?ids=...
Server -> Browser (SSE): initial aggregate_update with current counters

3) New Interaction Arrives (Reaction/Zap/Reply)
Relay -> Indexer: deliver event
Indexer -> DB: insert event; update refs
Indexer -> Aggregates: increment counters for target event; set last_interaction_at
Indexer -> SSE: emit new_reaction/new_zap/new_reply + aggregate_update
Server -> Browser (SSE): push updates
Browser: merge updates into DOM (bump counts, append items)

4) Cache Invalidation
- For stable content (new note/article by owner):
  Indexer: insert event
  Server: invalidate section keys affected (notes/articles); render cache miss on next request
- For interactions:
  Aggregates updated; section caches NOT invalidated immediately
  ETag changes via aggregate overlay; conditional GET yields 200 with fresh ETag or 304 when unchanged

5) Conditional Request Flow
Browser -> Server: GET / (If-None-Match: etag)
Server: recompute etag quickly from cached section ids + aggregates
If match -> 304 Not Modified
Else -> 200 with updated content and new etag

6) Static Export + Live
- Export writes static HTML and feeds from cached sections/renders.
- Optional JS connects to SSE for live overlays on static pages.
- Without SSE, users see content as-of export time; feeds carry ETags for client caching.
