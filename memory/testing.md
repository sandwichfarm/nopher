Testing Plan

Unit
- NIP parsers: NIP-10 (threading), NIP-25 (reactions), NIP-57 (zaps), NIP-65 (relay hints), NIP-19 (bech32 ids).
- Filters and section query compiler to SQL.
- Relay hint store and freshness logic.

Integration
- Seed bootstrap to dynamic relay discovery for owner and authors.
- FOAF/mutual graph expansion with caps and allow/deny lists.
- Sync cursors and replaceable kind refresh across restarts.
- Pruning behavior with diagnostics.

End-to-End
- Run against a small set of public relays; verify sections populate and archives render.
- Static export correctness (links, feeds, archives).
- Load tests for author caps and per-section limits.
