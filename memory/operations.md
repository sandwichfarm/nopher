Operations and Observability

Logging
- Levels: error, warn, info, debug, trace (configurable).
- Structured logs with context: relay, kind, pubkey, counts, durations.

Diagnostics
- Status page: connected relays, cursors per relay/kind, relay hints age, authors in-scope, pruning stats.
- Health checks: relay reachability, DB writable, disk space.

Backups
- SQLite file backups; optional vacuum and integrity check.

Config Reload
- Optional SIGHUP or file watcher to reload config without restart (best-effort, safe-only fields).

Retention Jobs
- Simple mode: Scheduled pruning according to sync.retention.keep_days; report skipped/pruned counts.
- Advanced mode (optional, Phase 17): Multi-dimensional retention with configurable rules, caps, and priorities.
  - Rule-based retention: Evaluate events against priority-ordered rules with various gates (time, size, kind, social distance, references).
  - Global caps: Enforce max_total_events, max_storage_mb, max_events_per_kind.
  - Protected events: Owner content and high-priority events never deleted by caps.
  - Incremental evaluation: Evaluate new events on ingestion, re-evaluate periodically.
  - Detailed diagnostics: Per-rule statistics, storage utilization, pruning reports.
  - See memory/retention_advanced.md for full specification.
