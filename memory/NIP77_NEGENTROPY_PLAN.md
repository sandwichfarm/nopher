# NIP-77 Negentropy Integration Plan

## Overview
Plan for integrating NIP-77 Negentropy protocol into nophr's sync engine for more efficient event synchronization.

## Date
2025-10-24

---

## What is NIP-77 Negentropy?

### Problem Statement
Current nophr sync uses traditional REQ-based queries:
- **Bandwidth inefficient**: Fetches all events since last cursor
- **Redundant transfers**: Re-fetches events we already have
- **No deduplication**: Each sync iteration may transfer duplicates

### Solution: Negentropy Set Reconciliation
NIP-77 uses **Range-Based Set Reconciliation** to efficiently sync event sets:
- **Bandwidth efficient**: Only transfers IDs of missing events
- **Bidirectional**: Discovers what both sides need
- **Logarithmic rounds**: Typically syncs in 2-4 message exchanges

### How It Works
1. Both sides build a negentropy vector from their local events
2. Vectors are compared to find differences
3. Only missing event IDs are exchanged
4. Actual events are fetched separately

### Expected Benefits
- **-50-80% bandwidth** for incremental syncs (most events already present)
- **-60-70% sync time** (fewer round trips)
- **Better for mobile/low-bandwidth** environments
- **Scales well** with large event sets

---

## Critical Constraint: Relay Compatibility

**⚠️ NOT ALL RELAYS SUPPORT NIP-77**

### Relay Support Detection
Before attempting negentropy sync, we must:

1. **Check relay capabilities** (via NIP-11 relay information document)
2. **Attempt NEG-OPEN** and handle errors gracefully
3. **Fall back to traditional REQ** if unsupported

### Error Handling
Relays may respond with:
- `NEG-ERR` with "unsupported" reason code
- `NOTICE` with error message
- Connection closure
- Timeout (no response)

**All of these must trigger fallback to REQ-based sync.**

---

## Integration Strategy

### Phase 1: Relay Capability Detection

**Files to modify**:
- `internal/nostr/client.go` (add capability detection)
- `internal/storage/storage.go` (track relay capabilities)

**Implementation**:
```go
// RelayCapabilities tracks what features a relay supports
type RelayCapabilities struct {
    URL                string
    SupportsNegentropy bool
    LastChecked        time.Time
    CheckedVersion     string // NIP-11 software version
}

// Check relay capabilities via NIP-11
func (c *Client) GetRelayCapabilities(ctx context.Context, url string) (*RelayCapabilities, error)

// Try negentropy handshake with timeout
func (c *Client) TestNegentropySupport(ctx context.Context, url string) bool
```

**Strategy**:
- Check capabilities on first connection to relay
- Cache results in storage (avoid re-checking every sync)
- Re-check after 7 days or if relay version changes
- Fallback to REQ immediately if unsupported

---

### Phase 2: Negentropy Sync Implementation

**Files to modify**:
- `internal/sync/engine.go` (add negentropy sync method)
- `internal/sync/negentropy.go` (NEW - wrapper around nip77)

**New file: `internal/sync/negentropy.go`**:
```go
package sync

import (
    "context"
    "github.com/nbd-wtf/go-nostr"
    "github.com/nbd-wtf/go-nostr/nip77"
)

// NegentropySync performs efficient sync using NIP-77 if supported
// Returns (success, error) where success=false triggers REQ fallback
func (e *Engine) NegentropySync(
    ctx context.Context,
    relay string,
    filter nostr.Filter,
) (bool, error) {
    // Check if relay supports negentropy
    if !e.relaySupportsNegentropy(relay) {
        return false, nil // Fallback to REQ
    }

    // Wrap storage as RelayStore for nip77
    store := &NegentropyStore{storage: e.storage, filter: filter}

    // Attempt negentropy sync (DOWN direction - fetch missing events)
    err := nip77.NegentropySync(ctx, store, relay, filter, nip77.Down)
    if err != nil {
        // Check if error is "unsupported" - update capabilities and fallback
        if isUnsupportedError(err) {
            e.markRelayDoesNotSupportNegentropy(relay)
            return false, nil
        }
        return false, err
    }

    return true, nil
}

// NegentropyStore adapts nophr's storage to nostr.RelayStore interface
type NegentropyStore struct {
    storage *storage.Storage
    filter  nostr.Filter
}

func (s *NegentropyStore) QueryEvents(ctx context.Context, filter nostr.Filter) ([]nostr.Event, error) {
    return s.storage.QueryEvents(ctx, filter)
}

func (s *NegentropyStore) Publish(ctx context.Context, event nostr.Event) error {
    return s.storage.StoreEvent(ctx, &event)
}
```

**Integration in engine.go**:
```go
func (e *Engine) syncRelay(relay string, filters []nostr.Filter) error {
    // Try negentropy sync first
    for _, filter := range filters {
        success, err := e.NegentropySync(e.ctx, relay, filter)
        if err != nil {
            return err // Hard error
        }
        if success {
            continue // Negentropy worked, next filter
        }

        // Fallback to traditional REQ-based sync
        fmt.Printf("[SYNC] Relay %s doesn't support negentropy, using REQ\n", relay)
        e.subscribeRelay(relay, []nostr.Filter{filter})
    }

    return nil
}
```

---

### Phase 3: Configuration

**Files to modify**:
- `internal/config/config.go` (add negentropy config)
- `configs/nophr.example.yaml` (document options)

**Config structure**:
```go
type SyncPerformance struct {
    Workers           int  `yaml:"workers"`              // Existing
    UseNegentropy     bool `yaml:"use_negentropy"`       // NEW: Enable NIP-77 (default: true)
    NegentropyFallback bool `yaml:"negentropy_fallback"` // NEW: Fallback to REQ if unsupported (default: true)
}
```

**Configuration example**:
```yaml
sync:
  performance:
    workers: 4
    use_negentropy: true       # Try NIP-77 if relay supports it
    negentropy_fallback: true  # Fall back to REQ if unsupported
```

**Use cases**:
- `use_negentropy: true, negentropy_fallback: true` - **Default**: Try negentropy, fallback to REQ (best of both)
- `use_negentropy: true, negentropy_fallback: false` - **Strict**: Only sync with NIP-77 relays (fail if unsupported)
- `use_negentropy: false, negentropy_fallback: true` - **Disable**: Always use REQ (for debugging/testing)

---

## Implementation Phases

### Phase 1: Capability Detection (Week 1)
**Tasks**:
1. Add `RelayCapabilities` struct to storage
2. Implement `GetRelayCapabilities()` via NIP-11
3. Implement `TestNegentropySupport()` with NEG-OPEN handshake
4. Add capability caching in storage
5. Test with known NIP-77 relays (strfry, others) and non-supporting relays

**Deliverables**:
- `internal/nostr/capabilities.go` (NEW)
- `internal/storage/relay_capabilities.go` (NEW)
- Tests with mock relays (supporting/non-supporting)

**Testing**:
- Test against strfry relay (supports NIP-77)
- Test against older relays (no NIP-77 support)
- Verify fallback behavior

---

### Phase 2: Negentropy Integration (Week 2)
**Tasks**:
1. Create `internal/sync/negentropy.go` wrapper
2. Implement `NegentropyStore` adapter for nophr storage
3. Integrate into `syncRelay()` with fallback logic
4. Add error handling for all failure modes
5. Add logging for negentropy vs REQ decision

**Deliverables**:
- `internal/sync/negentropy.go` (NEW)
- Modified `internal/sync/engine.go`
- Comprehensive error handling

**Testing**:
- Unit tests for `NegentropyStore` adapter
- Integration test: sync with negentropy-supporting relay
- Integration test: fallback to REQ when unsupported
- Integration test: handle mid-sync errors

---

### Phase 3: Configuration & Polish (Week 3)
**Tasks**:
1. Add config options for negentropy
2. Update example configs with documentation
3. Add metrics/logging for negentropy usage
4. Document in memory/operations.md
5. Update SYNC_OPTIMIZATIONS_TIER2.md with negentropy info

**Deliverables**:
- Updated config structs and examples
- Metrics tracking (negentropy vs REQ usage)
- Documentation

**Testing**:
- Test all config combinations
- Verify metrics accuracy
- Performance benchmarks (bandwidth savings)

---

## Expected Performance Impact

### Baseline (Current REQ-based sync)
- Bandwidth: ~100KB per sync iteration (typical)
- Round trips: 1-2 per relay
- Time: 500-1000ms per relay

### With Negentropy (Incremental sync, 90% overlap)
- Bandwidth: ~20-40KB per sync iteration (**-60-80%**)
- Round trips: 2-4 per relay (negentropy protocol)
- Time: 200-400ms per relay (**-50-60%**)

### With Negentropy (Initial sync, no overlap)
- Bandwidth: Similar to REQ (still need all events)
- Round trips: 3-5 per relay (ID discovery + event fetch)
- Time: Similar to REQ (extra rounds may add latency)

### Best Use Cases
✅ **Incremental syncs** - Most events already synced
✅ **Mobile/metered connections** - Bandwidth constrained
✅ **Large event sets** - Many relays/authors
✅ **Continuous sync** - Frequent polling

### Not Beneficial For
❌ **Initial sync** - No events cached yet
❌ **High churn** - Most events are new every time
❌ **Single-author** - Small event set, REQ is fine

---

## Risk Assessment

### Technical Risks

**Risk 1: Storage interface mismatch**
- **Impact**: NegentropyStore adapter may not map cleanly to nophr storage
- **Mitigation**: Prototype adapter first, verify all required methods
- **Fallback**: Use REQ-based sync if adapter is too complex

**Risk 2: Relay compatibility detection false positives**
- **Impact**: Try negentropy on unsupported relay, waste time
- **Mitigation**: Conservative detection (require explicit support signal)
- **Fallback**: Short timeout on NEG-OPEN, quick fallback to REQ

**Risk 3: Negentropy library bugs**
- **Impact**: Crashes, incorrect sync, data loss
- **Mitigation**: Thorough testing, fallback to REQ on errors
- **Fallback**: Config option to disable negentropy entirely

### User Experience Risks

**Risk 4: Confusing behavior with mixed relay support**
- **Impact**: Some relays sync fast (negentropy), others slow (REQ)
- **Mitigation**: Clear logging: "[SYNC] Using negentropy for wss://relay1" vs "[SYNC] Using REQ for wss://relay2"
- **Fallback**: None needed, this is expected behavior

**Risk 5: Increased memory usage**
- **Impact**: Negentropy builds vectors from all local events
- **Mitigation**: Limit negentropy to filters with <10K events
- **Fallback**: Use REQ for large filters

---

## Testing Strategy

### Unit Tests
- `NegentropyStore` adapter methods
- Capability detection logic
- Error handling for all failure modes
- Config parsing and defaults

### Integration Tests
- Sync with NIP-77-supporting relay (e.g., local strfry instance)
- Sync with non-supporting relay (mock or older relay)
- Fallback behavior when relay returns NEG-ERR
- Mixed relay support (some with, some without NIP-77)

### Performance Tests
- Bandwidth comparison: negentropy vs REQ
- Latency comparison: time to sync
- Memory usage during negentropy sync
- Verify savings with different overlap percentages (10%, 50%, 90%)

### Compatibility Tests
- Test against known relay implementations:
  - strfry (supports NIP-77)
  - nostream (check support status)
  - relay.nostr.band (check support status)
  - Custom mock relay (controlled behavior)

---

## Implementation Checklist

### Phase 1: Capability Detection
- [ ] Add `RelayCapabilities` struct to storage schema
- [ ] Implement NIP-11 relay info fetching
- [ ] Implement NEG-OPEN handshake test
- [ ] Add capability caching (7-day expiry)
- [ ] Test with supporting and non-supporting relays

### Phase 2: Negentropy Integration
- [ ] Create `internal/sync/negentropy.go`
- [ ] Implement `NegentropyStore` adapter
- [ ] Integrate into `syncRelay()` with fallback
- [ ] Add comprehensive error handling
- [ ] Add logging for negentropy usage

### Phase 3: Configuration & Documentation
- [ ] Add config options (`use_negentropy`, `negentropy_fallback`)
- [ ] Update example configs
- [ ] Add metrics tracking
- [ ] Update documentation (operations.md, SYNC_OPTIMIZATIONS_TIER2.md)
- [ ] Add troubleshooting guide

### Testing
- [ ] Unit tests for all new code
- [ ] Integration tests with real relays
- [ ] Performance benchmarks
- [ ] Compatibility matrix across relay types

---

## Alternative Approaches Considered

### Alternative 1: Always use negentropy, no fallback
**Pros**: Simpler code, forces relay adoption
**Cons**: Breaks sync for non-supporting relays
**Decision**: ❌ Rejected - Too disruptive, reduces relay compatibility

### Alternative 2: Detect support via NIP-11 only (no handshake test)
**Pros**: Faster capability detection
**Cons**: Relays may report incorrect support, no verification
**Decision**: ❌ Rejected - Not reliable enough, handshake test is safer

### Alternative 3: Use negentropy as separate optional sync mode
**Pros**: Keeps existing sync unchanged, safer
**Cons**: Duplicates sync logic, users must choose mode
**Decision**: ❌ Rejected - Automatic fallback is better UX

---

## Success Criteria

### Must Have
✅ Negentropy sync works with supporting relays
✅ Automatic fallback to REQ for non-supporting relays
✅ No breaking changes to existing sync behavior
✅ Config option to disable negentropy entirely
✅ Clear logging of which sync method is used

### Should Have
✅ 50-80% bandwidth reduction for incremental syncs
✅ 40-60% latency reduction for incremental syncs
✅ Capability detection cached to avoid repeated checks
✅ Comprehensive error handling
✅ Documentation and examples

### Nice to Have
- Metrics dashboard showing negentropy vs REQ usage
- Per-relay statistics (bandwidth saved, sync time)
- Automatic retry logic for failed negentropy syncs
- User notification when all relays support negentropy

---

## Next Steps

**Review this plan and decide**:
1. Should we proceed with NIP-77 integration?
2. Is the phased approach acceptable (3 weeks)?
3. Are there any additional concerns or requirements?
4. Should negentropy be enabled by default (`use_negentropy: true`)?

**If approved, I will**:
1. Start with Phase 1 (Capability Detection)
2. Implement and test each phase incrementally
3. Document progress and any issues encountered
4. Provide benchmarks comparing negentropy vs REQ performance

---

## References

- **NIP-77 Specification**: https://github.com/nostr-protocol/nips/blob/master/77.md
- **go-nostr nip77 package**: https://github.com/nbd-wtf/go-nostr/tree/master/nip77
- **Negentropy protocol**: https://github.com/hoytech/negentropy
- **strfry relay** (supports NIP-77): https://github.com/hoytech/strfry
