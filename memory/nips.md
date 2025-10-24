NIPs Covered

- **NIP-01**: Basic protocol and event kinds
- **NIP-02**: Contact List (kind 3)
- **NIP-10**: Replies and threading (#e root/reply tags)
- **NIP-18**: Reposts (kind 6)
- **NIP-19**: bech32 encodings (npub, nsec, note, etc.) - ✅ **Entity resolution implemented**
- **NIP-23**: Long-form content (kind 30023)
- **NIP-25**: Reactions (kind 7)
- **NIP-50**: Search capability - ✅ **Implemented**
- **NIP-57**: Zaps (kind 9735)
- **NIP-65**: Relay List Metadata (kind 10002) - Used for inbox/outbox relay discovery

## Implementation Details

### NIP-19: Entity Resolution ✅
- Parse `nostr:` URIs from event content
- Decode npub, nprofile, note, nevent, naddr
- Resolve to human-readable names/titles
- Protocol-specific formatters (Gopher, Gemini)
- Implemented in `internal/entities/resolver.go`

### NIP-50: Search ✅
- Full-text content search
- Uses `Filter.Search` field
- Server-side relevance ranking
- Supports multiple event kinds
- Implemented in `internal/search/nip50.go` and `internal/storage/search.go`

### NIP-65: Relay Discovery ✅
- Parse kind 10002 relay lists
- Extract read/write relay preferences
- Used for inbox/outbox relay selection
- Implemented in relay discovery system

Notes
- Additional NIPs may be adopted as needed
- Some NIPs are partially implemented (focus on core features)
- NIP implementations may evolve over time
