Implementation Plan

1) Foundation
- Stack: Go + Khatru (relay framework) + eventstore (SQLite or LMDB backend).
- Implement config loader with schema validation and env overrides.
- Initialize embedded Khatru relay instance with chosen eventstore backend.

2) Storage and Schema
- Configure Khatru with SQLite or LMDB eventstore for event storage.
- Create custom tables for relay_hints, graph_nodes, sync_state, aggregates.
- Khatru handles events table, indexing, and replaceable event logic.

3) Relay Discovery (Nostr)
- Seed bootstrap for kinds 0/3/10002; parse NIP-65; persist relay hints; periodic refresh.

4) Sync Engine and Graph (Nostr)
- Relay manager; connect to remote relays via nostr-sdk or websocket client.
- Subscribe with filters; stream events to local Khatru instance via StoreEvent.
- Per-relay/kind cursors in sync_state table; kind-3 parsing; FOAF/mutual computation; caps and diagnostics.

5) Inbox/Outbox Aggregation
- Query Khatru for events using standard Nostr filters.
- Threading (NIP-10); reactions and zaps rollups; compute aggregates table.
- Build queries for sections and feeds from Khatru's event storage.

6) Layouts and Sections
- Section/filter schema; default layout; archive and feeds; page composition.

7) Protocol Servers
- Gopher server (RFC 1436): menu/text rendering on port 70; selector routing; gophermap generation.
- Gemini server (gemini://): TLS with self-signed/TOFU certs on port 1965; gemtext rendering; input handling.
- Finger server (RFC 742): user info on port 79; query handling for owner and followed users.

8) Rendering Engines
- Query events from Khatru using Nostr filters; cache results.
- Gopher: convert Nostr events to plain text and gophermaps; thread display; selector-based navigation.
- Gemini: convert Nostr events to gemtext with links; support input for search/filters.
- Finger: format profile and recent activity as finger responses.
- Markdown conversion: implement parsers for Gopher (plain text), Gemini (gemtext), Finger (compact).

9) Publisher (Optional)
- Sign/publish outbox to Nostr relays; relay health checks; retry/backoff.

10) Operations and Safety
- Logging, status page (via Gopher/Gemini), pruning and retention; backups.

11) Testing and Docs
- Unit/integration/e2e; example config; deployment guide; protocol compliance testing; troubleshooting.

12) CI/CD and Automation
- Create local scripts for test, lint, build (scripts/*.sh).
- Set up Makefile for common development tasks.
- Configure GitHub Actions workflows (test, lint, build, release, docker).
- Set up GoReleaser for automated releases on tag push.
- Configure changelog generation from conventional commits.
- Set up GPG signing for releases.
- Configure multi-arch Docker builds.

13) Distribution and Packaging
- Configure GoReleaser for multi-platform releases (Linux, macOS, BSD).
- Create Dockerfile with multi-stage build and multi-arch support.
- Package for Homebrew, APT, RPM repositories.
- Create one-line installer script (curl | sh).
- Write systemd service files and reverse proxy examples.
- Embed default configs and templates using //go:embed.
- Generate checksums and GPG signatures for releases.

17) Advanced Configurable Retention (NEW FEATURE)
- Multi-dimensional retention with configurable rules and priorities.
- Retention gates: time, size, quantity, kind, social distance, references.
- Global caps: max_total_events, max_storage_mb, per-kind limits.
- Priority-based rule evaluation with first-match semantics.
- Protected events (owner, high-priority) never deleted by caps.
- Score-based deletion when caps exceeded.
- Incremental evaluation: on ingestion and periodic re-evaluation.
- New retention_metadata table (additive, doesn't modify existing schema).
- Backward compatible: simple keep_days retention unchanged.
- Opt-in: advanced.enabled must be explicitly set to true.
- Detailed diagnostics: per-rule stats, storage utilization, pruning reports.
- See memory/retention_advanced.md and memory/PHASE20_COMPLETION.md for full specification.
