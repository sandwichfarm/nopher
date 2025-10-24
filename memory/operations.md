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
- Scheduled pruning according to sync.retention; report skipped/pruned counts.
