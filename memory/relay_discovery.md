Relay Discovery (Seed-Only Bootstrap)

Goal
- Let operators change their personal relays without updating server config.
- Use configured seed relays only to discover where to fetch data, not as the permanent read/write set.

Mechanism
- Bootstrap from seeds: fetch latest kinds 0, 3, and 10002 for the owner (and later, for in-scope authors).
- Parse kind 10002 (NIP-65) relay hints to build dynamic per-pubkey relay sets:
  - read relays: where to subscribe for that author's events
  - write relays: where to publish (used only if publishing is enabled)
- Refresh periodically (e.g., every 15 minutes) and on significant updates.
- Fallback to seeds when hints are missing or stale.

Algorithm
1) From relays.seeds, query latest 0/3/10002 for the owner npub.
2) Store parsed NIP-65 hints in relay_hints(pubkey, relay, read/write, freshness).
3) Build the active read set for the owner from hints; use these read relays for owner data.
4) When expanding to other authors (per sync.scope policy), get their 10002 using:
   - Their own read relays if known; otherwise try seeds.
5) Maintain cursors per relay and kind; refresh replaceable kinds (0/3/10002) regularly.
6) If no hints found for an author, temporarily use seeds and mark as pending.

Seed Relays
- Provide example seed list in config. These are not hard-coded; operators may change them.
- Seeds should be chosen for good coverage of 0/3/10002. Some public relays specialize in these kinds.

Notes
- Discovery is per-pubkey: different authors can have different active relay sets.
- Respect max_relays_per_author to avoid excessive connections.
- Do not persist secrets in discovery state.
