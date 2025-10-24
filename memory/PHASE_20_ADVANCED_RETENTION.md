Phase 20: Advanced Configurable Retention

Status: FUTURE FEATURE - Extends Phase 12 simple retention

Overview

This phase adds advanced, configurable retention capabilities to the existing nophr implementation. The current simple retention system (`keep_days`) remains fully functional. This is an additive feature that does not modify existing infrastructure.

Goal

Implement a sophisticated, multi-dimensional retention system that allows fine-grained control over which events to keep based on configurable rules and priority.

Why This is Safe to Add

✅ **Backward Compatible** - Existing `sync.retention.keep_days` unchanged
✅ **Additive Only** - New table, new package, no modifications to existing code
✅ **Opt-In** - Must explicitly enable `advanced.enabled: true`
✅ **Independent** - Retention logic is isolated from sync, storage, and protocol servers
✅ **Fallback Safe** - Invalid config falls back to simple mode

Dependencies

**Completed Phases Required:**
- Phase 2: Storage Layer (Khatru + custom tables) ✅
- Phase 4: Sync Engine (social graph computation) ✅
- Phase 5: Aggregates (interaction counts) ✅
- Phase 12: Operations (existing pruning) ✅

**Does NOT Depend On:**
- Protocol servers (Gopher, Gemini, Finger)
- Markdown conversion
- Caching
- Publisher

Deliverables

1. Configuration Schema Extension
   - Add `sync.retention.advanced` section to config
   - Parse and validate retention rules
   - Maintain backward compatibility with simple `keep_days`

2. Database Schema Addition
   - Add `retention_metadata` table (new, does not modify existing tables)
   - Indexes for efficient querying
   - Migration that safely adds table to existing databases

3. Rule Evaluation Engine
   - Parse retention rules from config
   - Evaluate conditions (gates) against events
   - Priority-based rule matching
   - Score calculation for events

4. Condition Evaluators (Gates)
   - Time-based gates (age, date ranges)
   - Size-based gates (content length, tag count)
   - Quantity-based gates (event counts per kind/author)
   - Kind-based gates (specific kinds, categories)
   - Social distance gates (FOAF, mutual, specific pubkeys)
   - Reference-based gates (replies to owner, interaction counts)
   - Logical operators (AND, OR, NOT)

5. Enhanced Pruning System
   - Use retention_metadata for pruning decisions
   - Fallback to simple keep_days if advanced disabled
   - Cap enforcement (max events, max storage, per-kind limits)
   - Score-based deletion when caps exceeded
   - Protected event handling

6. Incremental Evaluation
   - Evaluate new events on ingestion
   - Re-evaluation scheduler for existing events
   - Batch processing to avoid performance impact
   - Config change detection and re-evaluation

7. Diagnostics and Observability
   - Retention statistics endpoint/page
   - Per-rule event counts
   - Storage utilization vs caps
   - Pruning reports (what/why deleted)
   - Logging for retention actions

8. Testing
   - Unit tests for each gate type
   - Rule evaluation tests
   - Priority ordering tests
   - Cap enforcement tests
   - Backward compatibility tests
   - Performance tests (evaluate 100k events)

Completion Criteria

✅ **Configuration:**
- [ ] Advanced retention config parses correctly
- [ ] Invalid config falls back to simple mode with warning
- [ ] Simple mode (keep_days only) still works unchanged

✅ **Database:**
- [ ] retention_metadata table created via migration
- [ ] Migration works on existing databases without data loss
- [ ] Indexes optimize queries

✅ **Rule Evaluation:**
- [ ] All gate types evaluate correctly
- [ ] Priority ordering works
- [ ] First-match semantics correct
- [ ] Logical operators (AND/OR/NOT) work

✅ **Pruning:**
- [ ] Events deleted based on retention_metadata
- [ ] Protected events never deleted
- [ ] Simple mode fallback works
- [ ] Cap enforcement deletes lowest-priority events first

✅ **Performance:**
- [ ] Can evaluate 100k events in <10 seconds
- [ ] Incremental evaluation on ingestion <10ms per event
- [ ] Pruning 10k events completes in <5 seconds

✅ **Observability:**
- [ ] Diagnostics show retention statistics
- [ ] Logging captures retention actions
- [ ] Pruning reports explain deletions

✅ **Backward Compatibility:**
- [ ] Existing databases upgrade cleanly
- [ ] Simple retention mode unaffected
- [ ] No breaking changes to existing APIs

Implementation Plan

Step 1: Configuration Schema (Non-Breaking)

**Files to Create:**
- `internal/config/retention.go` - Advanced retention config structs
- `internal/config/retention_test.go` - Config parsing tests

**Changes to Existing Files:**
- `internal/config/config.go` - Add `Advanced *AdvancedRetention` field to existing `Retention` struct
- Preserve existing `KeepDays` and `PruneOnStart` fields
- Make `Advanced` optional (nil if not configured)

**Example:**
```go
// Existing struct (unchanged)
type Retention struct {
    KeepDays     int  `yaml:"keep_days"`
    PruneOnStart bool `yaml:"prune_on_start"`
    // New field (additive)
    Advanced     *AdvancedRetention `yaml:"advanced,omitempty"`
}

// New struct
type AdvancedRetention struct {
    Enabled    bool          `yaml:"enabled"`
    Mode       string        `yaml:"mode"` // "rules" or "caps"
    GlobalCaps GlobalCaps    `yaml:"global_caps"`
    Rules      []RetentionRule `yaml:"rules"`
    Evaluation EvalConfig    `yaml:"evaluation"`
}
```

Step 2: Database Migration (Additive Only)

**Files to Create:**
- `internal/storage/migrations/017_retention_metadata.sql` - New table DDL
- `internal/storage/retention_metadata.go` - CRUD operations for retention_metadata

**Migration:**
```sql
-- Only adds new table, does not modify existing tables
CREATE TABLE IF NOT EXISTS retention_metadata (
  event_id TEXT PRIMARY KEY,
  rule_name TEXT NOT NULL,
  rule_priority INTEGER NOT NULL,
  retain_until INTEGER,
  last_evaluated_at INTEGER NOT NULL,
  score INTEGER,
  protected BOOLEAN DEFAULT 0,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE INDEX idx_retention_metadata_retain_until ON retention_metadata(retain_until);
CREATE INDEX idx_retention_metadata_score ON retention_metadata(score);
CREATE INDEX idx_retention_metadata_protected ON retention_metadata(protected);
```

**Safe Migration:**
- Use `CREATE TABLE IF NOT EXISTS` (idempotent)
- Does not alter existing tables
- Foreign key ensures cleanup if event deleted
- Can be rolled back by dropping table

Step 3: Rule Evaluation Engine

**Files to Create:**
- `internal/retention/engine.go` - Rule evaluation orchestrator
- `internal/retention/rules.go` - Rule definition and matching
- `internal/retention/conditions.go` - Condition evaluation (gates)
- `internal/retention/actions.go` - Action application
- `internal/retention/score.go` - Event scoring for caps
- `internal/retention/engine_test.go` - Engine tests

**Interface (Isolated from Existing Code):**
```go
package retention

type Engine struct {
    config  *config.AdvancedRetention
    storage Storage  // Interface to storage layer
    graph   SocialGraph  // Interface to graph_nodes
}

func NewEngine(cfg *config.AdvancedRetention, storage Storage, graph SocialGraph) *Engine

func (e *Engine) EvaluateEvent(event *Event) (*RetentionDecision, error)

func (e *Engine) EvaluateBatch(events []*Event) ([]*RetentionDecision, error)

type RetentionDecision struct {
    EventID      string
    RuleName     string
    RulePriority int
    RetainUntil  *time.Time  // nil = forever
    Protected    bool
    Score        int
}
```

**Does Not Modify:**
- Existing storage interface
- Existing sync engine
- Existing pruning (extended, not replaced)

Step 4: Condition Evaluators (Gates)

**Files to Create:**
- `internal/retention/conditions/time.go` - Time-based gates
- `internal/retention/conditions/size.go` - Size-based gates
- `internal/retention/conditions/quantity.go` - Quantity-based gates
- `internal/retention/conditions/kind.go` - Kind-based gates
- `internal/retention/conditions/social.go` - Social distance gates
- `internal/retention/conditions/reference.go` - Reference-based gates
- `internal/retention/conditions/logical.go` - AND/OR/NOT operators
- `internal/retention/conditions/conditions_test.go` - Gate tests

**Each Gate Interface:**
```go
type Condition interface {
    Evaluate(ctx *EvalContext) (bool, error)
}

type EvalContext struct {
    Event       *Event
    Storage     Storage
    SocialGraph SocialGraph
    Aggregates  *Aggregates
}
```

Step 5: Enhanced Pruning (Extends Existing)

**Files to Modify (Carefully):**
- `internal/ops/retention.go` - Extend existing pruning logic

**Changes:**
```go
// Existing function signature (unchanged)
func (r *RetentionManager) Prune() error

// Implementation (extended, not replaced)
func (r *RetentionManager) Prune() error {
    // Check if advanced retention enabled
    if r.config.Retention.Advanced != nil && r.config.Retention.Advanced.Enabled {
        return r.pruneAdvanced()
    }

    // Existing simple pruning (unchanged)
    return r.pruneSimple()
}

// New function (additive)
func (r *RetentionManager) pruneAdvanced() error {
    // 1. Query retention_metadata for expired events
    // 2. Delete expired events
    // 3. Check global caps
    // 4. If caps exceeded, delete by score
    // 5. Update statistics
}

// Existing function (unchanged)
func (r *RetentionManager) pruneSimple() error {
    // Existing implementation unchanged
}
```

**Safety:**
- Existing `pruneSimple()` untouched
- New path only taken if advanced enabled
- Fallback to simple on any error in advanced

Step 6: Integration Points

**Sync Engine Integration (Minimal):**

Add hook to evaluate new events on ingestion:

```go
// In internal/sync/engine.go
func (s *SyncEngine) processEvent(event *Event) error {
    // Existing: Store in Khatru
    err := s.storage.StoreEvent(event)
    if err != nil {
        return err
    }

    // New: Evaluate retention (if advanced enabled)
    if s.retentionEngine != nil {
        decision, err := s.retentionEngine.EvaluateEvent(event)
        if err != nil {
            log.Warn("retention evaluation failed", "error", err)
            // Continue anyway, don't fail ingestion
        } else {
            err = s.storage.StoreRetentionDecision(decision)
            // Errors logged but don't stop ingestion
        }
    }

    // Existing: Update aggregates, etc.
    return s.updateAggregates(event)
}
```

**Key Point:** Retention evaluation is non-blocking. Failures don't stop sync.

Step 7: Diagnostics Extension

**Files to Modify:**
- `internal/ops/diagnostics.go` - Add retention stats section

**Addition:**
```go
// Add to existing DiagnosticsInfo struct
type DiagnosticsInfo struct {
    // Existing fields unchanged
    Relays        []RelayStatus
    Cursors       []CursorStatus
    Authors       int

    // New field (additive)
    Retention     *RetentionStats `json:"retention,omitempty"`
}

type RetentionStats struct {
    Enabled           bool
    TotalEvents       int
    ProtectedEvents   int
    ExpiringWithin7d  int
    ByRule           map[string]int
    StorageMB        float64
    StorageCapMB     float64
    LastPruneAt      time.Time
    LastPruneDeleted int
}
```

Step 8: Testing

**Files to Create:**
- `internal/retention/engine_test.go` - Engine tests
- `internal/retention/conditions/*_test.go` - Gate tests
- `test/integration/retention_test.go` - Integration tests
- `test/performance/retention_bench_test.go` - Performance benchmarks

**Test Coverage:**
- Unit tests for each gate type (>80% coverage)
- Rule priority ordering
- First-match semantics
- Logical operators
- Score calculation
- Cap enforcement
- Backward compatibility (simple mode still works)
- Migration safety (existing DB upgrades cleanly)

File Structure

New Files (Additive):
```
internal/
  retention/              # New package
    engine.go             # Rule evaluation orchestrator
    rules.go              # Rule matching
    actions.go            # Action application
    score.go              # Event scoring
    engine_test.go
    conditions/           # New subpackage
      time.go
      size.go
      quantity.go
      kind.go
      social.go
      reference.go
      logical.go
      conditions_test.go
  config/
    retention.go          # New file for advanced config
    retention_test.go
  storage/
    retention_metadata.go # New file for retention_metadata table
    migrations/
      017_retention_metadata.sql  # New migration

test/
  integration/
    retention_test.go     # New integration tests
  performance/
    retention_bench_test.go  # New benchmarks
```

Modified Files (Minimal Changes):
```
internal/
  config/
    config.go             # Add Advanced field to Retention struct
  ops/
    retention.go          # Extend Prune() to call pruneAdvanced()
    diagnostics.go        # Add RetentionStats to DiagnosticsInfo
  sync/
    engine.go             # Add retention evaluation hook (optional)
```

Configuration Migration

Old Config (Still Works):
```yaml
sync:
  retention:
    keep_days: 365
    prune_on_start: true
```

New Config (Opt-In):
```yaml
sync:
  retention:
    keep_days: 365        # Fallback if advanced fails
    prune_on_start: true

    advanced:
      enabled: true       # Must explicitly enable
      mode: "rules"

      global_caps:
        max_total_events: 1000000

      rules:
        - name: "protect_owner"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true
```

Rollout Plan

1. **Deploy Phase 20 Code**
   - New retention package added
   - Config schema extended (backward compatible)
   - Advanced retention disabled by default

2. **Database Migration Runs**
   - retention_metadata table created
   - Existing events unaffected
   - No downtime required

3. **Test Advanced Retention**
   - Enable `advanced.enabled: true` in config
   - Configure rules
   - Monitor diagnostics
   - Verify pruning behavior

4. **Iterate on Rules**
   - Adjust rules based on actual usage
   - Monitor storage utilization
   - Fine-tune caps and priorities

5. **Optional: Disable Simple Mode**
   - Once confident in advanced retention
   - Keep `keep_days` as ultimate fallback

Risks and Mitigations

Risk: Advanced retention breaks existing pruning
✅ **Mitigation:** Advanced is opt-in, simple mode untouched, fallback on errors

Risk: Database migration fails
✅ **Mitigation:** Migration is idempotent (IF NOT EXISTS), doesn't alter existing tables

Risk: Performance impact on sync
✅ **Mitigation:** Evaluation is async, errors don't block, batch processing

Risk: Configuration complexity overwhelms users
✅ **Mitigation:** Simple mode still works, advanced is optional, good defaults provided

Risk: Bugs in cap enforcement delete important events
✅ **Mitigation:** Protected flag prevents deletion, extensive testing, gradual rollout

Success Metrics

- Advanced retention enabled and working in production
- No regressions in simple retention mode
- Database migration completes on existing instances
- Pruning reduces storage as configured
- Diagnostics show retention statistics
- Performance remains acceptable (evaluation <10ms per event)
- Users report satisfied with retention control

Documentation Updates

Update these files:
- `memory/configuration.md` - Add advanced retention config examples
- `memory/operations.md` - Update retention section with advanced features
- `memory/storage_model.md` - Document retention_metadata table
- `memory/PHASES.md` - Add Phase 20 summary

Do NOT change:
- `memory/README.md` - Status unchanged, phases are internal
- User-facing docs - Only document after implementation complete
