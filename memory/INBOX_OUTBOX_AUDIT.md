# Inbox/Outbox Model Audit - Issues Found

## Date
2025-10-24

---

## Executive Summary

❌ **Critical Issues Found** in inbox/outbox (NIP-65) implementation

### Issues Discovered
1. **GetRelaysForPubkey() logic is backwards** - Queries READ relays first when it should query WRITE relays
2. **Mentions/interactions not using inbox model** - Querying all relays instead of our inbox
3. **Zap calculations are correct** ✅

---

## Background: NIP-65 Inbox/Outbox Model

### Terminology
- **Outbox (WRITE relays)**: Where a user **publishes** their content
- **Inbox (READ relays)**: Where a user **receives** interactions/notifications

### Correct Usage
| Action | Query Location | Reason |
|--------|---------------|---------|
| Read someone's posts | Their WRITE relays (outbox) | That's where they publish |
| Find replies TO us | OUR READ relays (inbox) | That's where others send replies |
| Find mentions OF us | OUR READ relays (inbox) | That's where others send mentions |
| Find reactions TO us | OUR READ relays (inbox) | That's where others send reactions |
| Find zaps TO us | OUR READ relays (inbox) | That's where others send zap receipts |

---

## Issue #1: GetRelaysForPubkey() Logic is Backwards

### Current Code
**File**: `internal/nostr/discovery.go:165-183`

```go
// GetRelaysForPubkey returns the relay URLs where a pubkey can be found
// Prefers read relays, but falls back to write relays if needed
func (d *Discovery) GetRelaysForPubkey(ctx context.Context, pubkey string) ([]string, error) {
    // First try read relays ❌ WRONG
    relays, err := d.storage.GetReadRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get read relays: %w", err)
    }

    if len(relays) > 0 {
        return relays, nil
    }

    // Fall back to write relays (they might publish there too) ❌ BACKWARDS
    relays, err = d.storage.GetWriteRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get write relays: %w", err)
    }

    return relays, nil
}
```

### Problem
- This function is used in `engine.go:600` to find relays for syncing **authors' posts**
- Posts are published to WRITE relays (outbox)
- **We're checking READ relays (inbox) first** - this is backwards!

### Where It's Used
**File**: `internal/sync/engine.go:600`
```go
for _, author := range authors {
    relays, err := e.discovery.GetRelaysForPubkey(e.ctx, author)
    // These should be the author's WRITE relays, not READ relays
}
```

### Impact
- **Low to Medium** - Most users specify relays without read/write markers
- When no marker specified, both `can_read=1` and `can_write=1` are set
- Only affects users who properly specify read/write markers in their NIP-65 lists
- But for those users, we're querying the wrong relays

### Correct Implementation
```go
// GetRelaysForPubkey returns the relay URLs where a pubkey PUBLISHES content
// Prefers write relays (outbox), but falls back to read relays if needed
func (d *Discovery) GetRelaysForPubkey(ctx context.Context, pubkey string) ([]string, error) {
    // First try write relays (outbox) - where they publish
    relays, err := d.storage.GetWriteRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get write relays: %w", err)
    }

    if len(relays) > 0 {
        return relays, nil
    }

    // Fall back to read relays as backup
    relays, err = d.storage.GetReadRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get read relays: %w", err)
    }

    return relays, nil
}
```

---

## Issue #2: Mentions/Interactions Not Using Inbox Model

### Current Code
**File**: `internal/sync/engine.go:342-349`

```go
// Add mention filter if configured
if e.config.Sync.Scope.IncludeDirectMentions {
    mentionFilter := e.filterBuilder.BuildMentionFilter(ownerPubkey, since)
    filters = append(filters, mentionFilter)
    fmt.Printf("[SYNC]   Added mention filter (total: %d filters)\n", len(filters))
}

// Try negentropy sync first, fall back to REQ if unsupported
go e.syncRelayWithFallback(relay, filters)
```

### Problem
- Mention filter (`#p` tag with our pubkey) is querying events that mention US
- These should be queried from OUR inbox (READ relays)
- **Currently querying from all authors' relays** (same relays as their posts)
- This mixes outbox queries (their posts) with inbox queries (mentions to us)

### Where Mentions Should Be Queried
According to inbox/outbox model:
- **Owner's READ relays (inbox)**: Where others send mentions/replies/reactions/zaps TO us
- **Not authors' WRITE relays**: Those are for reading THEIR posts

### Impact
- **High** - We're likely missing many mentions/interactions
- If someone mentions us and sends to our inbox relays, but we're not querying those relays for mentions
- We only find mentions if they happen to also be on the relays where we're reading posts

### Current Flow (Incorrect)
```
For each relay in (authors' outbox relays):
    Query:
    1. Authors' posts (✅ correct - outbox)
    2. Mentions of us (❌ wrong - should be our inbox)
```

### Correct Flow
```
Step 1: Query authors' posts
For each author in scope:
    Get author's WRITE relays (outbox)
    Query their posts from outbox

Step 2: Query interactions to us (separate!)
Get OUR READ relays (inbox)
Query:
    - Mentions of us (#p tag)
    - Replies to our posts (#e tag)
    - Reactions to our posts (kind 7)
    - Zaps to our posts (kind 9735)
```

---

## Issue #3: No Separate Inbox Query

### Current Architecture
```go
syncOnce() {
    for each relay in (authors' relays):
        Build filters:
        - Authors' posts
        - Mentions of us (if enabled)

        Sync from relay
}
```

### Problem
- No separate step to query OUR inbox
- All queries bundled together and sent to authors' outbox relays

### What's Missing
```go
syncOnce() {
    // Step 1: Query authors' posts from their outbox
    for each author in scope:
        relays = author.GetWriteRelays()  // ✅ Outbox
        Query author's posts

    // Step 2: Query OUR inbox for interactions ❌ MISSING
    ownerInboxRelays = owner.GetReadRelays()  // Our inbox
    Query:
        - #p mentions of us
        - #e replies to our posts
        - kind 7 reactions to our posts
        - kind 9735 zaps to our posts
}
```

---

## Issue #4: Zap Calculations

### Status
✅ **Zap calculations are CORRECT**

### Code Review
**File**: `internal/storage/aggregates.go:182-193`

```go
func (s *Storage) AddZapAmount(ctx context.Context, eventID string, sats int64, interactionAt int64) error {
    query := `
        INSERT INTO aggregates (event_id, reply_count, reaction_total, zap_sats_total, last_interaction_at)
        VALUES (?, 0, 0, ?, ?)
        ON CONFLICT(event_id) DO UPDATE SET
            zap_sats_total = zap_sats_total + excluded.zap_sats_total,  ✅ Cumulative
            last_interaction_at = MAX(last_interaction_at, excluded.last_interaction_at)
    `
    // ...
}
```

### Verification
- ✅ Uses `zap_sats_total + excluded.zap_sats_total` for cumulative total
- ✅ Parses bolt11 invoices correctly (internal/aggregates/zaps.go:128-164)
- ✅ Batch processing working (internal/storage/aggregates.go:243-274)
- ✅ Test coverage confirms correctness (internal/storage/storage_test.go:394-404)

---

## Impact Assessment

### High Impact
1. **Missing interactions** - We're likely missing many mentions, replies, reactions, and zaps
   - Only found if they happen to be on relays we query for posts
   - Not systematically querying our inbox

2. **Inefficient relay usage** - Querying wrong relays
   - Wasting bandwidth on relays that don't have our interactions
   - Not querying relays that DO have our interactions

### Medium Impact
1. **GetRelaysForPubkey backwards logic**
   - Only affects users who properly mark read/write in their relay lists
   - Most users don't specify markers (defaults to both)
   - But for proper NIP-65 users, we're querying wrong relays

### Low Impact
1. **Zap calculations** - ✅ Working correctly

---

## Recommended Fixes

### Fix #1: Correct GetRelaysForPubkey()

**File**: `internal/nostr/discovery.go`

```go
// GetOutboxRelays returns where a pubkey PUBLISHES content (write relays)
func (d *Discovery) GetOutboxRelays(ctx context.Context, pubkey string) ([]string, error) {
    // Try write relays first (outbox)
    relays, err := d.storage.GetWriteRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get write relays: %w", err)
    }

    if len(relays) > 0 {
        return relays, nil
    }

    // Fall back to read relays
    return d.storage.GetReadRelays(ctx, pubkey)
}

// GetInboxRelays returns where a pubkey RECEIVES interactions (read relays)
func (d *Discovery) GetInboxRelays(ctx context.Context, pubkey string) ([]string, error) {
    // Try read relays first (inbox)
    relays, err := d.storage.GetReadRelays(ctx, pubkey)
    if err != nil {
        return nil, fmt.Errorf("failed to get read relays: %w", err)
    }

    if len(relays) > 0 {
        return relays, nil
    }

    // Fall back to write relays
    return d.storage.GetWriteRelays(ctx, pubkey)
}
```

### Fix #2: Separate Inbox/Outbox Sync

**File**: `internal/sync/engine.go`

```go
func (e *Engine) syncOnce() error {
    ownerPubkey, err := e.getOwnerPubkey()
    if err != nil {
        return err
    }

    authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
    if err != nil {
        return err
    }

    // STEP 1: Sync authors' posts from their OUTBOX
    e.syncAuthorsOutbox(authors)

    // STEP 2: Sync interactions to US from OUR INBOX
    if e.config.Sync.Scope.IncludeDirectMentions {
        e.syncOwnerInbox(ownerPubkey)
    }

    return nil
}

func (e *Engine) syncAuthorsOutbox(authors []string) {
    // Get authors' WRITE relays (where they publish)
    for _, author := range authors {
        relays, err := e.discovery.GetOutboxRelays(e.ctx, author)
        // Query their posts from their outbox
    }
}

func (e *Engine) syncOwnerInbox(ownerPubkey string) {
    // Get OUR READ relays (where we receive interactions)
    inboxRelays, err := e.discovery.GetInboxRelays(e.ctx, ownerPubkey)

    // Build inbox filter (mentions, replies, reactions, zaps TO us)
    filter := e.filterBuilder.BuildInboxFilter(ownerPubkey, since)

    // Query from our inbox relays
    for _, relay := range inboxRelays {
        go e.syncRelayWithFallback(relay, []nostr.Filter{filter})
    }
}
```

### Fix #3: Add BuildInboxFilter()

**File**: `internal/sync/filters.go`

```go
// BuildInboxFilter creates a filter for interactions directed at the owner
// Queries: mentions (#p), replies (#e), reactions (kind 7), zaps (kind 9735)
func (fb *FilterBuilder) BuildInboxFilter(ownerPubkey string, since int64) nostr.Filter {
    // Get owner's event IDs for #e tag filtering (replies to our posts)
    // This would require a storage query for our recent event IDs

    filter := nostr.Filter{
        Kinds: []int{1, 6, 7, 9735}, // notes, reposts, reactions, zaps
        Tags: nostr.TagMap{
            "p": []string{ownerPubkey}, // Mentions/interactions to us
        },
    }

    if since > 0 {
        sinceTs := nostr.Timestamp(since)
        filter.Since = &sinceTs
    }

    return filter
}
```

---

## Testing Plan

### Unit Tests
1. Test `GetOutboxRelays()` returns write relays first
2. Test `GetInboxRelays()` returns read relays first
3. Test `BuildInboxFilter()` creates correct filter

### Integration Tests
1. Create test relay with marked read/write relays
2. Verify posts queried from outbox
3. Verify interactions queried from inbox
4. Verify mentions found in inbox but not outbox

### Real-World Validation
1. Test with npub that properly marks relay lists
2. Verify all mentions/reactions/zaps are found
3. Compare before/after interaction counts

---

## Priority

### High Priority (Fix Now)
1. ❌ Fix `GetRelaysForPubkey()` - rename and fix logic
2. ❌ Separate inbox sync from outbox sync

### Medium Priority (Next Sprint)
1. Add comprehensive inbox/outbox tests
2. Add metrics for inbox vs outbox queries

### Low Priority (Future)
1. Optimize relay selection (limit queries)
2. Add caching for relay hints

---

## Backwards Compatibility

### Breaking Changes
None - these are bug fixes that improve correctness

### Migration
No data migration needed - just code changes

### Deployment
- Can deploy immediately
- Will automatically find more interactions
- No user configuration required

---

## Summary

| Component | Status | Issue | Priority |
|-----------|--------|-------|----------|
| GetRelaysForPubkey() | ❌ Backwards | Queries READ first, should query WRITE | High |
| Mention queries | ❌ Wrong relays | Queries outbox, should query inbox | High |
| Inbox sync | ❌ Missing | No separate inbox query step | High |
| Zap calculations | ✅ Correct | Cumulative calculations working | N/A |

### Action Items
1. ✅ Document issues (this file)
2. ⏳ Implement fixes
3. ⏳ Add tests
4. ⏳ Deploy and validate

---

**Status**: ❌ **Critical issues found - fixes required**
**Date**: 2025-10-24
