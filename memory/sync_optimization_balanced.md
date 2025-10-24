# Balanced Sync Performance Optimization Plan

## Goal
Improve sync throughput **without** sacrificing latency or ballooning memory usage.

## Current Performance Baseline
- Throughput: ~100-200 events/sec
- Latency (event→storage): ~50-100ms
- Memory: ~50-100MB
- DB transactions: 1 per event (~100-200/sec)

## Target Performance (Balanced)
- Throughput: 500-1000 events/sec (**5-10x improvement**)
- Latency: <150ms (**≤50ms increase**)
- Memory: <200MB (**≤2x increase**)
- DB transactions: 20-50/sec (**5-10x reduction**)

---

## Tier 1: Quick Wins (No Trade-offs)

### 1. **Deduplication Before Storage** ✅ IMPLEMENTED
**Impact**: Reduces wasted DB queries by 30-70%
**Trade-offs**: None (small memory for bloom filter)
**Latency**: -10ms (faster, skips duplicate work)
**Memory**: +10MB (bloom filter for 1M events)

**Implementation**:
```go
// Already added: EventExists() method
// Simple approach: check DB before storing
if exists, _ := storage.EventExists(ctx, event.ID); exists {
    return nil // Skip duplicate
}
```

**Next step**: Add in-memory LRU cache (1000 recent IDs) for even faster checks

---

### 2. **Connection Pooling**
**Impact**: Reduces connection overhead
**Trade-offs**: None
**Latency**: -5ms (reuse connections)
**Memory**: Negligible

**Implementation**:
- Set `MaxOpenConns` and `MaxIdleConns` on SQL DB
- Reuse WebSocket connections to relays

---

### 3. **Smart Sync Intervals**
**Impact**: Reduces unnecessary polling
**Trade-offs**: None (actually improves everything)
**Latency**: 0ms
**Memory**: 0MB
**CPU**: -20% (less polling)

**Implementation**:
```go
// Adaptive interval based on activity
if eventsReceived == 0 {
    interval = 30 * time.Second  // Slow when idle
} else if eventsReceived < 100 {
    interval = 10 * time.Second  // Normal
} else {
    interval = 5 * time.Second   // Fast when active
}
```

---

## Tier 2: Small Trade-offs (Recommended)

### 4. **Increase Event Buffer (Moderate)**
**Impact**: Handles burst traffic better
**Trade-offs**: +20MB memory for larger buffer
**Latency**: 0ms (no change)
**Throughput**: +20% (reduces blocking)

**Current**: `eventChan: make(chan *nostr.Event, 1000)`
**Proposed**: `eventChan: make(chan *nostr.Event, 5000)` ← Balanced size

---

### 5. **Async Aggregate Updates**
**Impact**: Don't block event storage on aggregate calculations
**Trade-offs**: Aggregates lag by 100-200ms
**Latency**: -30ms for event storage (huge win!)
**Throughput**: +50% (removes bottleneck)
**Memory**: +20MB (aggregate queue)

**Implementation**:
```go
// Queue aggregate updates, don't process inline
func (e *Engine) processEvent(event *nostr.Event) error {
    // Store event immediately
    if err := e.storage.StoreEvent(e.ctx, event); err != nil {
        return err
    }

    // Queue aggregate update (non-blocking)
    e.queueAggregateUpdate(event)

    return nil
}

// Background worker processes aggregates in batches
func (e *Engine) aggregateWorker() {
    ticker := time.NewTicker(200 * time.Millisecond)
    pending := make(map[string]*AggregateUpdate)

    for {
        select {
        case update := <-e.aggregateChan:
            pending[update.EventID] = update

        case <-ticker.C:
            if len(pending) > 0 {
                e.flushAggregates(pending)
                pending = make(map[string]*AggregateUpdate)
            }
        }
    }
}
```

**Result**: Events appear in DB immediately, stats update 200ms later

---

### 6. **Small Worker Pool (2-4 workers)**
**Impact**: Parallel processing without excessive overhead
**Trade-offs**: Slight ordering complexity
**Latency**: 0ms
**Throughput**: +100-200% (2-4x)
**Memory**: +40MB (4 workers × 10MB each)

**Implementation**:
```go
// Start 4 worker goroutines
for i := 0; i < 4; i++ {
    go e.eventWorker()
}

func (e *Engine) eventWorker() {
    for event := range e.eventChan {
        e.processEvent(event) // Each worker processes independently
    }
}
```

**Ordering**: Events from different authors can be processed in parallel (safe)

---

## Tier 3: Optional Features (User Choice)

### 7. **Configurable Batching** (Off by Default)
**Impact**: Massive throughput for power users
**Trade-offs**: Latency increases significantly
**Config**:
```yaml
sync:
  performance:
    batch_size: 1      # Default: no batching (low latency)
    # batch_size: 50   # Optional: high throughput (adds 100-500ms latency)
    batch_timeout_ms: 100
```

**Use case**: Users syncing hundreds of thousands of events who don't care about latency

---

## Recommended Implementation Order

### Phase 1: Quick Wins (Week 1)
✅ Deduplication check (already done)
□ Add LRU cache for recent event IDs
□ Connection pooling
□ Smart sync intervals

**Expected**: +30-50% throughput, -15ms latency, +10MB memory

---

### Phase 2: Async Optimizations (Week 2)
□ Async aggregate updates
□ Increase buffer to 5000

**Expected**: +70-100% throughput, -30ms latency for events, +40MB memory

---

### Phase 3: Parallelization (Week 3)
□ 4-worker pool
□ Prepared statements for aggregates

**Expected**: +200-300% throughput total, +40MB memory

---

### Phase 4: Optional (Week 4)
□ Configurable batching (opt-in)
□ Bloom filter for deduplication
□ Metrics and monitoring

---

## Comparison: Balanced vs Aggressive

| Metric | Current | Balanced Plan | Aggressive Plan |
|--------|---------|---------------|-----------------|
| Throughput | 100/s | **500-1000/s** | 5000/s |
| Latency | 50ms | **80-120ms** | 300-500ms |
| Memory | 50MB | **150-200MB** | 500MB+ |
| Complexity | Low | **Medium** | High |
| Trade-offs | None | **Minimal** | Significant |

---

## Decision Points

**Choose Balanced Plan if**:
- ✅ You want steady improvements without downsides
- ✅ You care about latency (events appear quickly)
- ✅ You want to keep memory usage reasonable
- ✅ You value code simplicity

**Choose Aggressive Plan if**:
- ❌ You're syncing 100k+ events initially
- ❌ You don't care about 500ms latency
- ❌ You have 1GB+ RAM to spare
- ❌ You're willing to maintain complex batching logic

---

## Next Steps

**Shall we proceed with the Balanced Plan?**

If yes, I'll implement in this order:
1. LRU cache for event deduplication (5 min)
2. Connection pooling (5 min)
3. Smart sync intervals (10 min)
4. Async aggregate updates (30 min)
5. Increase buffer + 4-worker pool (20 min)

**Total time**: ~70 minutes for 5-10x throughput with minimal trade-offs
