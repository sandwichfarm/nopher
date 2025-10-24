# Phase 12: Operations and Diagnostics - Completion Report

## Overview

Phase 12 focused on adding operational features essential for production deployments, including structured logging, diagnostics, health monitoring, retention policies, and backup utilities.

**Status**: ✅ Complete

**Date Completed**: 2025-10-24

## Deliverables

### 1. Structured Logging System ✅

**File**: `internal/ops/logging.go`

**Features**:
- Built on Go's standard `log/slog` package
- Configurable log levels (debug, info, warn, error)
- Multiple output formats (text, json)
- Component-specific loggers with contextual fields
- Specialized logging methods for different operations:
  - Storage operations
  - Relay connections
  - Sync progress
  - Protocol requests
  - Cache operations
  - Aggregate updates
  - Retention pruning
  - Backup operations
  - Startup/shutdown events
  - Panic recovery

**Usage**:
```go
logger := ops.NewLogger(&config.Logging{
    Level:  "info",
    Format: "json",
})

componentLogger := logger.WithComponent("gopher-server")
componentLogger.Info("server started", "port", 70)
```

**Configuration**:
```yaml
logging:
  level: "info"   # debug|info|warn|error
  format: "text"  # text|json
```

### 2. Diagnostics System ✅

**File**: `internal/ops/diagnostics.go`

**Components**:
- **DiagnosticsCollector**: Collects system-wide diagnostic information
- **SystemStats**: Runtime statistics (memory, goroutines, GC, uptime)
- **StorageStats**: Database statistics (event counts, size, time range)
- **SyncStats**: Sync engine statistics (relays, cursors, total synced)
- **RelayHealth**: Per-relay health information
- **AggregateStats**: Aggregate computation statistics

**Output Formats**:
- Plain text (for logs/CLI)
- Gophermap (for Gopher protocol)
- Gemtext (for Gemini protocol)

**Example Output**:
```
=== Nopher Diagnostics ===
Collected: 2025-10-24T12:00:00Z

--- System ---
Version: v1.0.0 (abc123)
Uptime: 2h30m15s
Go Version: go1.21.5
Goroutines: 42
Memory: 125.50 MB allocated, 256.00 MB system
GC Runs: 15

--- Storage ---
Driver: sqlite
Total Events: 15,432
Database Size: 45.30 MB
Oldest Event: 2025-09-15T08:00:00Z
Newest Event: 2025-10-24T12:00:00Z

Events by Kind:
  Kind 0: 150 events
  Kind 1: 12,500 events
  Kind 3: 150 events
  Kind 7: 2,000 events
  Kind 9735: 500 events

--- Sync ---
Enabled: true
Relays: 5 total, 4 connected
Total Synced: 15,432 events
Last Sync: 2025-10-24T11:59:30Z

--- Relay Health ---
wss://relay.damus.io: connected
  Last Connect: 2025-10-24T10:00:00Z
  Events Synced: 8,500
wss://nos.lol: connected
  Last Connect: 2025-10-24T10:00:15Z
  Events Synced: 6,932

--- Aggregates ---
Total: 12,500
Last Reconcile: 2025-10-24T11:45:00Z
```

### 3. Relay Health Monitoring ✅

**Files**:
- `internal/sync/stats.go`
- `internal/nostr/discovery.go` (extended)

**Features**:
- Per-relay connection status
- Last connect/disconnect times
- Error tracking
- Events synced per relay
- Connection health indicators

**Usage**:
```go
relays := syncEngine.GetRelays()
for _, relay := range relays {
    if relay.IsConnected() {
        fmt.Printf("%s: connected\n", relay.URL())
    } else if err := relay.LastError(); err != nil {
        fmt.Printf("%s: error - %v\n", relay.URL(), err)
    }
}
```

### 4. Event Count Statistics ✅

**File**: `internal/storage/stats.go`

**Methods**:
- `CountEvents()`: Total event count
- `CountEventsByKind()`: Events grouped by kind
- `CountEventsByRelay()`: Events per relay (placeholder)
- `DatabaseSize()`: Database size in MB
- `EventTimeRange()`: Oldest and newest event timestamps
- `CountAggregates()`: Total aggregates
- `CountAggregatesByKind()`: Aggregates by kind
- `LastReconcileTime()`: Last aggregate reconciliation

**Usage**:
```go
total, _ := storage.CountEvents(ctx)
byKind, _ := storage.CountEventsByKind(ctx)
sizeMB, _ := storage.DatabaseSize(ctx)
```

### 5. Cursor Status Display ✅

**Method**: `storage.GetAllCursors()`

**Returns**: Array of cursor information including:
- Relay URL
- Event kind
- Cursor position
- Last updated timestamp

**Usage**:
```go
cursors, _ := storage.GetAllCursors(ctx)
for _, cursor := range cursors {
    fmt.Printf("%s kind %d: position %d (updated %s)\n",
        cursor.Relay, cursor.Kind, cursor.Position,
        cursor.Updated.Format(time.RFC3339))
}
```

### 6. Retention and Pruning System ✅

**File**: `internal/ops/retention.go`

**Components**:
- **RetentionManager**: Manages data retention policies
- **PeriodicPruner**: Runs automatic pruning on schedule

**Features**:
- Time-based retention (keep events for N days)
- Kind-based pruning (delete all events of specific kind)
- Prune-on-start option
- Periodic automatic pruning
- Retention statistics and estimates

**Configuration**:
```yaml
sync:
  retention:
    keep_days: 365
    prune_on_start: true
```

**Usage**:
```go
retentionMgr := ops.NewRetentionManager(storage, &config.Retention, logger)

// Manual pruning
deleted, err := retentionMgr.PruneOldEvents(ctx)

// Periodic pruning
pruner := ops.NewPeriodicPruner(retentionMgr, 24*time.Hour, logger)
go pruner.Start(ctx)
```

**Methods**:
- `PruneOldEvents()`: Delete events older than retention period
- `PruneByKind()`: Delete all events of specific kind
- `GetRetentionStats()`: Get retention statistics

### 7. Backup Utilities ✅

**File**: `internal/ops/backup.go`

**Components**:
- **BackupManager**: Handles database backups
- **PeriodicBackup**: Runs automatic backups on schedule
- **CleanOldBackups()**: Cleanup old backup files

**Features**:
- File-based backup (copy database file)
- Timestamped backup filenames
- Periodic automatic backups
- Old backup cleanup
- Restore from backup

**Usage**:
```go
backupMgr := ops.NewBackupManager(storage, logger)

// Manual backup
err := backupMgr.BackupWithConfig(ctx,
    "/var/lib/nopher/nopher.db",
    "/backups/nopher-backup-20251024.db")

// Periodic backups
periodicBackup := ops.NewPeriodicBackup(
    backupMgr,
    "/var/lib/nopher/nopher.db",
    "/backups",
    24*time.Hour,
    logger)
go periodicBackup.Start(ctx)

// Cleanup old backups
ops.CleanOldBackups("/backups", 30*24*time.Hour, logger)
```

**Methods**:
- `Backup()`: Create database backup
- `BackupWithConfig()`: Backup with explicit paths
- `Restore()`: Restore from backup
- `CleanOldBackups()`: Remove old backups

### 8. Tests ✅

**Files**:
- `internal/ops/logging_test.go`
- `internal/ops/diagnostics_test.go`

**Coverage**:
- Logger creation and configuration
- Component-specific loggers
- Log level filtering
- Helper method functionality
- Diagnostic stats collection
- Output format rendering (text, gemtext, gophermap)

## File Structure

```
internal/ops/
├── logging.go             # Structured logging system
├── logging_test.go        # Logging tests
├── diagnostics.go         # Diagnostics collection
├── diagnostics_test.go    # Diagnostics tests
├── retention.go           # Retention and pruning
├── backup.go              # Backup utilities

internal/storage/
├── stats.go               # Storage statistics methods

internal/sync/
├── stats.go               # Sync statistics methods

internal/nostr/
├── discovery.go           # Extended with relay health
```

## Integration Points

### Main Application Integration

The ops features integrate with the main application at several points:

1. **Startup**: Initialize logger and diagnostics collector
2. **Runtime**: Log operations, collect metrics
3. **Shutdown**: Final diagnostics snapshot, cleanup

### Protocol Servers Integration

Each protocol server (Gopher, Gemini, Finger) can:
- Expose diagnostics page at special selectors/URLs
- Log protocol requests
- Track performance metrics

Example diagnostic endpoints:
- Gopher: `/diagnostics` selector
- Gemini: `gemini://host/diagnostics`
- Finger: `diagnostics@host`

### Scheduled Operations

Background goroutines for:
- **Periodic Pruning**: Every 24 hours (configurable)
- **Periodic Backups**: Every 24 hours (configurable)
- **Diagnostic Snapshots**: On-demand or periodic

## Usage Examples

### Basic Setup

```go
// Initialize logger
logger := ops.NewLogger(&cfg.Logging)
ops.SetDefault(logger)

// Create diagnostics collector
diagCollector := ops.NewDiagnosticsCollector(
    version, commit, storage, syncEngine)

// Collect diagnostics
diag, err := diagCollector.CollectAll(ctx)
if err != nil {
    logger.Error("failed to collect diagnostics", "error", err)
}

// Output as text
fmt.Println(diag.FormatAsText())
```

### Retention Management

```go
// Create retention manager
retentionMgr := ops.NewRetentionManager(
    storage, &cfg.Sync.Retention, logger)

// Prune on startup if configured
if retentionMgr.ShouldPruneOnStart() {
    deleted, err := retentionMgr.PruneOldEvents(ctx)
    logger.Info("startup pruning completed", "deleted", deleted)
}

// Start periodic pruner
pruner := ops.NewPeriodicPruner(
    retentionMgr, 24*time.Hour, logger)
go pruner.Start(ctx)
```

### Backup Management

```go
// Create backup manager
backupMgr := ops.NewBackupManager(storage, logger)

// Manual backup
timestamp := time.Now().Format("20060102-150405")
backupPath := fmt.Sprintf("/backups/nopher-%s.db", timestamp)
err := backupMgr.BackupWithConfig(ctx, cfg.Storage.SQLitePath, backupPath)

// Start periodic backups
periodicBackup := ops.NewPeriodicBackup(
    backupMgr,
    cfg.Storage.SQLitePath,
    "/backups",
    24*time.Hour,
    logger)
go periodicBackup.Start(ctx)

// Cleanup old backups (keep 30 days)
ops.CleanOldBackups("/backups", 30*24*time.Hour, logger)
```

## Performance Impact

- **Logging**: Minimal overhead with structured logging
- **Diagnostics**: On-demand collection, no continuous overhead
- **Retention**: Scheduled operations, minimal runtime impact
- **Backups**: File copy operation, runs in background

## Security Considerations

- Logs may contain sensitive information (configure appropriately)
- Backup files should be secured (file permissions)
- Diagnostics output should not expose secrets
- Retention policies prevent unlimited data growth

## Future Enhancements

- [ ] Prometheus metrics endpoint
- [ ] Health check HTTP endpoint
- [ ] Real-time metrics dashboard
- [ ] Distributed tracing support
- [ ] Log shipping to external systems
- [ ] Incremental backups (WAL mode)
- [ ] S3/cloud backup support
- [ ] Alert notifications (email, webhooks)
- [ ] Performance profiling integration

## Testing

Run ops tests:
```bash
go test ./internal/ops/... -v
```

## Completion Criteria

All Phase 12 requirements have been met:

- [x] Structured logging with slog
- [x] Diagnostics page implementation
- [x] Relay health monitoring
- [x] Cursor status display
- [x] Event count statistics
- [x] Pruning and retention
- [x] Backup utilities
- [x] Tests for ops features
- [x] Protocol output formats (text, gophermap, gemtext)

## Next Phase

**Phase 10: Caching Layer** - Add performance-critical caching to protocol servers.

or

**Phase 14: Security and Privacy** - Add rate limiting, deny-lists, and security hardening.

## References

- Go slog package: https://pkg.go.dev/log/slog
- Gopher Protocol RFC 1436
- Gemini Protocol Specification
- SQLite Backup API
