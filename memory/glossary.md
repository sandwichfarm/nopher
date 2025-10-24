Glossary

Owner
- The Nostr identity (npub) whose notes/articles the site showcases. Single-tenant by default.

Operator
- The person deploying and configuring the site. Often the same as Owner.

npub / nsec / note id (NIP-19)
- bech32 encodings for public key (npub), secret key (nsec), and event id (note...). Secrets are env-only.

Relay
- Nostr server speaking NIP-01. May be used for reading and/or writing.

Seed Relays
- Operator-configured relays used only to find kinds 0, 3, and 10002 for bootstrap and refresh.
- Not the permanent read/write set.

Relay Hints (NIP-65, kind 10002)
- Event listing preferred read/write relays for a pubkey. Parsed into per-author active relay sets.
- Freshness = created_at of the latest 10002 for that pubkey.

Kinds (examples)
- 0: Metadata (profile)
- 1: Notes (short-form)
- 3: Contacts (follows graph)
- 6: Reposts
- 7: Reactions
- 9735: Zaps
- 30023: Long-form content (NIP‑23)
- 10002: Relay List Metadata (NIP‑65)

Replaceable / Parameterized Replaceable
- Replaceable: latest event per (kind, pubkey). Example: 0, 3, 10002.
- Parameterized replaceable: latest event per (kind, pubkey, parameter). Example: 30023 with d-tag.

FOAF / Mutual
- FOAF: Friends-of-a-friend expansion from the owner to depth N via kind 3 follows.
- Mutual: Reciprocal follow between owner and an author.

Inbox / Outbox
- Outbox: content authored by the owner (notes, articles, reposts, reactions, zaps if enabled).
- Inbox: content targeting the owner or the owner's events (replies, reactions, zaps, mentions).

Refs (Relationships)
- Extracted edges between events and pubkeys: root/reply (NIP-10), mention (#p), reference (#e), reaction, zap, repost.

Graph
- Owner-centric social graph built from kind 3. Used for scope: self/following/mutual/FOAF.

Discovery
- Process of building per-pubkey active relay sets from NIP-65 (10002) using seed relays for bootstrap.

Sync Engine
- Subscribes to read relays with filters (authors, kinds, since cursors), deduplicates, ingests, and updates derived state.

Aggregates
- Derived counters per event: reply_count, reaction_total (+ per char), zap_sats_total, last_interaction_at.

Sections / Templates / Pages
- Section: query-driven view over stored data; has filters, transforms, protocol-specific rendering.
- Template: rendering style for protocol (gophermap menu, gemtext page, finger response).
- Page: composition of sections served via Gopher selectors, Gemini URLs, or Finger queries.

Gopher (RFC 1436)
- Internet protocol from 1991 for hierarchical document distribution via menus and text files.
- Uses selectors to navigate; served on port 70; no encryption by default.
- Item types: 0 (text file), 1 (submenu/directory), i (informational line), 3 (error).

Gophermap
- Menu file format for Gopher servers; each line defines an item with type, display text, selector, host, port.
- Example: "0About this site\t/about.txt\tgopher.example.com\t70"

Gemini (gemini://)
- Lightweight internet protocol emphasizing privacy and simplicity; released 2019.
- Uses gemtext markup (simpler than Markdown); TLS required; served on port 1965.
- Trust-On-First-Use (TOFU) certificate model; supports user input via status codes.

Gemtext
- Text/gemini markup format: headings (# ## ###), links (=> URL text), quotes (>), preformatted (```), lists (*).
- Simpler than Markdown; designed for readability and machine parsing.

Finger (RFC 742)
- User information protocol from 1977; queries user details on port 79.
- Query format: username@host; response is plain text (profile, .plan, recent activity).
- .plan: traditional Unix file containing user's current status or bio.

Caching
- Render cache (per event/protocol), section-result cache, protocol response cache. Namespaced by config hash.
- TTLs per protocol (gophermaps, gemtext, finger responses); serve stale if source unavailable.

Retention / Pruning
- Time-based or cap-based removal of old events to control storage and scope explosion.

Cursors
- Per-relay, per-kind since timestamps to avoid replaying old events; refreshed for replaceable kinds periodically.

Config Hash / Query Hash / Template Hash
- Short hashes used to namespace cache keys, identify queries, and version render templates.
