Synchronization Scope and Limits

Author Set Modes
- self: only the owner
- following: owner plus everyone in owner's kind-3 contact list
- mutual: authors with reciprocal follow with owner
- foaf(depth): breadth-first expansion of the social graph to the given depth

Modifiers
- include_direct_mentions: always include events that directly mention the owner via #p
- include_threads_of_mine: include full threads under the owner's posts
- allowlist_pubkeys: authors always included
- denylist_pubkeys: authors always excluded
- max_authors: stop expansion when cap is reached

Kinds and Threads
- Default kinds: 0,1,3,6,7,9735,30023,10002 (configurable)
- Replies identified via NIP-10 (#e with root/reply markers)
- Threads can be limited by depth and item count to prevent explosion

Retention and Pruning
- keep_days: time-based pruning of events
- prune_on_start: optional pruning run at startup
- Diagnostics should report when pruning or caps affect visible results

Filters and Subscriptions
- Build filters per kind and per author set, with since-cursors per relay
- Always include special filters for mentions of the owner and for replies to the owner's events when the related flags are set
- Replaceable kinds (0/3/10002) are refreshed periodically regardless of since-cursors
