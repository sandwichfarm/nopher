Nopher: Config-First Personal Nostr-to-Gopher/Gemini/Finger Gateway

Overview
- Serves Nostr content via Gopher (RFC 1436), Gemini (gemini://), and Finger (RFC 742) protocols only.
- Single-tenant by default; shows one operator's notes and articles from Nostr.
- Everything configurable via file and env overrides; no hard-coded relays.
- Inbox/Outbox model to aggregate replies, reactions, and zaps from Nostr.
- Seed-relay bootstrap; dynamic relay discovery via NIP-65 (kind 10002).
- Controlled synchronization scope (self/following/mutual/FOAF depth) with caps and allow/deny lists.
- Embedded Khatru relay for event storage; SQLite or LMDB backend (no PostgreSQL).
- Protocol-specific rendering: Gopher menus/text, Gemini gemtext, Finger user info.
- Composable layouts: sections defined by filters; sensible defaults and archives.
- Privacy and safety: env-only secrets, pruning, deny-lists, and diagnostics.

Status
- This directory documents the design decisions and plan. Implementation has not started yet.

Implementation Guide:
- PHASES.md (phased implementation plan with deliverables and completion criteria)
- ../AGENTS.md (instructions for AI agents working on this project)

Architecture and Design:
- architecture.md (system design and component overview)
- glossary.md (terminology reference)

Development Process:
- cicd.md (CI/CD pipeline and release automation)
- distribution.md (installation and packaging strategy)
- testing.md (testing strategy)

Core Systems:
- storage_model.md (Khatru integration and database schema)
- configuration.md (config system and options)
- sequence_seed_discovery_sync.md (Nostr sync flow)

Features:
- ui_export.md (Gopher/Gemini/Finger protocol rendering)
- markdown_conversion.md (markdown to protocol conversion)
- layouts_sections.md (configurable sections and pages)
- inbox_outbox.md (interaction aggregation)
- caching.md (caching strategy)

Nostr Integration:
- relay_discovery.md (NIP-65 relay discovery)
- sync_scope.md (social graph and scope control)
- nips.md (Nostr protocol specifications)

Operations:
- operations.md (operational procedures)
- diagnostics.md (monitoring and diagnostics)
- security_privacy.md (security and privacy features)
- retention_advanced.md (Phase 17: advanced configurable retention system)

