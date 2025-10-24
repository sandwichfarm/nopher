Advanced Configurable Retention System

Overview

This document describes the advanced retention system for nophr, allowing fine-grained control over which events to keep based on multiple configurable criteria (gates). This extends the existing simple retention system without replacing it.

Design Goals

1. **Backward Compatible** - Existing `sync.retention.keep_days` continues to work
2. **Composable Gates** - Multiple retention rules can be combined with AND/OR logic
3. **Priority-Based** - Protect important events (owner, close friends, certain kinds)
4. **Multi-Dimensional** - Control by time, size, quantity, kind, social distance, specific pubkeys
5. **Configurable** - All rules defined in configuration, no hard-coded policies
6. **Observable** - Detailed logging of what gets pruned and why

Retention Architecture

Existing Simple Retention (Unchanged)
```yaml
sync:
  retention:
    keep_days: 365           # Keep events newer than 365 days
    prune_on_start: true     # Run pruning on startup
```

This continues to work as-is. If only `keep_days` is set, behavior is unchanged.

New Advanced Retention
```yaml
sync:
  retention:
    # Simple retention (backward compatible)
    keep_days: 365
    prune_on_start: true

    # Advanced retention system (optional)
    advanced:
      enabled: false         # Must explicitly enable
      mode: "rules"          # rules|caps

      # Global caps (apply after rules)
      global_caps:
        max_total_events: 1000000      # Hard limit on total events
        max_storage_mb: 5000           # Hard limit on database size
        max_events_per_kind:
          1: 100000                    # Max 100k kind 1 (notes)
          30023: 10000                 # Max 10k kind 30023 (articles)

      # Retention rules (evaluated in order)
      rules:
        - name: "protect_owner"
          description: "Never delete owner's content"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true

        - name: "protect_close_friends"
          description: "Keep all content from close friends indefinitely"
          priority: 900
          conditions:
            social_distance_max: 1     # Direct follows only
            author_in_list: []         # Or specific npubs
          action:
            retain: true

        - name: "priority_pubkeys"
          description: "Keep priority users' notes for 2 years"
          priority: 800
          conditions:
            author_in_list:
              - "npub1priorityuser1..."
              - "npub1priorityuser2..."
            kinds: [1, 30023]
          action:
            retain_days: 730

        - name: "important_kinds"
          description: "Keep articles and long-form longer than notes"
          priority: 700
          conditions:
            kinds: [30023]
          action:
            retain_days: 1095         # 3 years

        - name: "close_network_notes"
          description: "Keep mutual/2-hop friends' notes for 1 year"
          priority: 600
          conditions:
            social_distance_max: 2
            kinds: [1]
          action:
            retain_days: 365

        - name: "interactions_on_mine"
          description: "Keep reactions/zaps on my posts for 6 months"
          priority: 500
          conditions:
            kinds: [7, 9735]
            references_owner_events: true
          action:
            retain_days: 180

        - name: "default_ephemeral"
          description: "Default: keep other events for 90 days"
          priority: 100
          conditions:
            all: true                 # Catch-all
          action:
            retain_days: 90
```

Retention Gates (Conditions)

Time-Based Gates
```yaml
conditions:
  created_after: "2024-01-01T00:00:00Z"  # ISO 8601 timestamp
  created_before: "2025-01-01T00:00:00Z"
  age_days_max: 90                       # Max age in days
  age_days_min: 7                        # Min age in days
```

Size-Based Gates
```yaml
conditions:
  content_size_max: 10240                # Max content length in bytes
  content_size_min: 10                   # Min content length
  tags_count_max: 50                     # Max number of tags
```

Quantity-Based Gates
```yaml
conditions:
  kind_count_max:
    1: 100000                            # Max 100k kind 1 events
  author_event_count_max: 1000           # Max events per author
  author_event_count_min: 10             # Min events per author
```

Kind-Based Gates
```yaml
conditions:
  kinds: [1, 30023]                      # Match specific kinds
  kinds_exclude: [7, 9735]               # Exclude specific kinds
  kind_category: "ephemeral"             # ephemeral|replaceable|parameterized|regular
```

Social Distance Gates
```yaml
conditions:
  social_distance_max: 2                 # FOAF distance from owner (0=owner, 1=following, 2=FOAF)
  social_distance_min: 1                 # Minimum distance
  author_is_owner: true                  # Event by owner
  author_is_following: true              # Direct follow
  author_is_mutual: true                 # Mutual follow
  author_in_list:                        # Specific pubkeys
    - "npub1..."
    - "npub2..."
  author_not_in_list:                    # Exclude pubkeys
    - "npub3..."
```

Reference-Based Gates
```yaml
conditions:
  references_owner_events: true          # References any owner event (replies, reactions, zaps)
  references_event_ids:                  # References specific events
    - "event_id_1"
    - "event_id_2"
  is_root_post: true                     # Is root of thread (no reply tags)
  is_reply: true                         # Is a reply
  has_replies: true                      # Has at least one reply
  reply_count_min: 5                     # Min reply count
  reaction_count_min: 10                 # Min reaction count
  zap_sats_min: 1000                     # Min zap amount
```

Combined Gates (Logical Operators)
```yaml
conditions:
  and:                                   # All must match
    - kinds: [1]
    - social_distance_max: 1
  or:                                    # Any must match
    - author_is_owner: true
    - author_is_mutual: true
  not:                                   # Must NOT match
    - kinds: [7]
```

Retention Actions

Retain Permanently
```yaml
action:
  retain: true                           # Never delete
```

Retain for Duration
```yaml
action:
  retain_days: 365                       # Keep for 365 days from created_at
  retain_until: "2026-01-01T00:00:00Z"   # Keep until specific date
```

Delete Immediately
```yaml
action:
  delete: true                           # Delete on next prune run
  delete_after_days: 30                  # Grace period before deletion
```

Archive (Future Enhancement)
```yaml
action:
  archive: true                          # Move to archive storage
  archive_after_days: 180               # Archive after 180 days
```

Rule Evaluation Logic

Priority and Order
1. Rules are sorted by priority (highest first)
2. For each event, rules are evaluated in priority order
3. First matching rule determines the action
4. If no rule matches, event is deleted (safe default)

Evaluation Process
```
For each event:
  1. Check each rule in priority order
  2. Evaluate all conditions in the rule
  3. If all conditions match:
     - Apply action
     - Stop evaluation (first match wins)
  4. If no rules match:
     - Apply default action (delete or use fallback rule)
```

Global Caps (Applied After Rules)
Even if a rule says "retain forever", global caps apply:
- `max_total_events`: Hard limit on total event count
- `max_storage_mb`: Hard limit on database size
- `max_events_per_kind`: Per-kind caps

When caps are exceeded:
1. Events are sorted by priority (rule priority × event score)
2. Lowest priority events are deleted first
3. Deletion continues until under cap

Event Scoring (for cap enforcement)
```
score = rule_priority × (
  + (is_owner_content ? 1000 : 0)
  + (social_distance_weight × 100)
  + (age_weight × 10)
  + (interaction_weight × 5)
)

social_distance_weight = max(0, 10 - social_distance)
age_weight = max(0, 10 - (age_days / 30))
interaction_weight = min(10, reply_count + reaction_count/10 + zap_sats/1000)
```

Configuration Examples

Minimal (Backward Compatible)
```yaml
sync:
  retention:
    keep_days: 365
    prune_on_start: true
```

Conservative (Keep Everything Important)
```yaml
sync:
  retention:
    keep_days: 90  # Fallback for simple mode
    advanced:
      enabled: true
      mode: "rules"

      global_caps:
        max_total_events: 5000000
        max_storage_mb: 20000

      rules:
        - name: "protect_owner"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true

        - name: "protect_network"
          priority: 900
          conditions:
            social_distance_max: 2
          action:
            retain_days: 730

        - name: "default"
          priority: 100
          conditions:
            all: true
          action:
            retain_days: 90
```

Aggressive (Limited Storage)
```yaml
sync:
  retention:
    advanced:
      enabled: true
      mode: "rules"

      global_caps:
        max_total_events: 100000
        max_storage_mb: 500
        max_events_per_kind:
          1: 50000
          7: 10000
          9735: 5000

      rules:
        - name: "owner_only"
          priority: 1000
          conditions:
            author_is_owner: true
          action:
            retain: true

        - name: "direct_follows_recent"
          priority: 800
          conditions:
            social_distance_max: 1
            kinds: [1, 30023]
          action:
            retain_days: 180

        - name: "interactions_on_mine"
          priority: 700
          conditions:
            references_owner_events: true
            kinds: [7, 9735]
          action:
            retain_days: 90

        - name: "popular_posts"
          priority: 600
          conditions:
            or:
              - reply_count_min: 10
              - zap_sats_min: 10000
          action:
            retain_days: 365

        - name: "ephemeral_default"
          priority: 100
          conditions:
            all: true
          action:
            retain_days: 30
```

Priority Pubkeys Focus
```yaml
sync:
  retention:
    advanced:
      enabled: true

      rules:
        - name: "priority_users_unlimited"
          priority: 1000
          conditions:
            author_in_list:
              - "npub1important1..."
              - "npub1important2..."
              - "npub1important3..."
          action:
            retain: true

        - name: "priority_users_extended"
          priority: 900
          conditions:
            author_in_list:
              - "npub1extended1..."
              - "npub1extended2..."
          action:
            retain_days: 730

        - name: "owner"
          priority: 950
          conditions:
            author_is_owner: true
          action:
            retain: true

        - name: "everyone_else"
          priority: 100
          conditions:
            all: true
          action:
            retain_days: 90
```

Implementation Details

Database Schema Changes (Additions Only)

No changes to existing tables. Add new table for retention metadata:

```sql
CREATE TABLE IF NOT EXISTS retention_metadata (
  event_id TEXT PRIMARY KEY,
  rule_name TEXT NOT NULL,
  rule_priority INTEGER NOT NULL,
  retain_until INTEGER,        -- Unix timestamp or NULL for permanent
  last_evaluated_at INTEGER NOT NULL,
  score INTEGER,               -- Calculated score for cap enforcement
  protected BOOLEAN DEFAULT 0, -- Cannot be deleted by caps
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE INDEX idx_retention_metadata_retain_until ON retention_metadata(retain_until);
CREATE INDEX idx_retention_metadata_score ON retention_metadata(score);
CREATE INDEX idx_retention_metadata_protected ON retention_metadata(protected);
```

Pruning Process

1. **Rule Evaluation Phase**
   - For each event without retention_metadata or needing re-evaluation
   - Evaluate rules in priority order
   - Store result in retention_metadata table
   - Mark protected events

2. **Expiration Phase**
   - Find events where retain_until < NOW
   - Exclude protected events
   - Delete events and update statistics

3. **Cap Enforcement Phase**
   - Check if global caps exceeded
   - If yes: Query events by score (lowest first)
   - Delete until under cap
   - Respect protected flag

4. **Cleanup Phase**
   - Vacuum database if significant deletions
   - Update diagnostics with pruning stats
   - Log retention report

Incremental Evaluation

To avoid re-evaluating all events on every run:
- Only evaluate new events (no retention_metadata)
- Re-evaluate if rules change (detect config hash change)
- Re-evaluate periodically (configurable interval, e.g., weekly)

```yaml
sync:
  retention:
    advanced:
      evaluation:
        on_ingest: true                    # Evaluate when event first stored
        re_eval_interval_hours: 168        # Re-evaluate weekly
        batch_size: 1000                   # Process in batches
```

Diagnostics and Observability

Retention Statistics
```
Retention Statistics:
  Total Events: 456,789
  Protected Events: 12,345 (owner: 5,000, rules: 7,345)
  Expiring Soon (7d): 3,456
  Under Retention: 441,098

  By Rule:
    protect_owner: 5,000 events (permanent)
    protect_close_friends: 7,345 events (permanent)
    important_kinds: 12,456 events (avg 456 days remaining)
    default_ephemeral: 416,297 events (avg 45 days remaining)

  Storage:
    Database Size: 2,345 MB / 5,000 MB (47%)
    Total Events: 456,789 / 1,000,000 (46%)

  Last Prune:
    Timestamp: 2025-01-15 10:30:00 UTC
    Events Deleted: 1,234
    Space Freed: 45 MB
    Duration: 2.3s
```

Logging
```
[INFO] Retention: Evaluating 1,234 new events
[INFO] Retention: Applied rule 'protect_owner' to 45 events
[INFO] Retention: Applied rule 'default_ephemeral' to 989 events
[INFO] Retention: Marked 1,234 events for retention
[INFO] Retention: Pruning expired events (567 candidates)
[INFO] Retention: Deleted 567 events, freed 23 MB
[WARN] Retention: Global cap 'max_total_events' exceeded, enforcing
[INFO] Retention: Cap enforcement deleted 100 lowest-priority events
```

Configuration Structure Summary

```yaml
sync:
  retention:
    # Simple mode (backward compatible)
    keep_days: 365
    prune_on_start: true

    # Advanced mode
    advanced:
      enabled: false
      mode: "rules"  # or "caps" (caps only, no rules)

      # Evaluation settings
      evaluation:
        on_ingest: true
        re_eval_interval_hours: 168
        batch_size: 1000

      # Global caps (hard limits)
      global_caps:
        max_total_events: 1000000
        max_storage_mb: 5000
        max_events_per_kind:
          1: 100000
          30023: 10000

      # Retention rules (priority-ordered)
      rules:
        - name: "rule_name"
          description: "Human-readable description"
          priority: 1000
          conditions:
            # Various gates (see above)
          action:
            # Retention action (see above)
```

Migration Path

Phase 1: Add retention_metadata table
- Add new table to schema
- Does not affect existing functionality

Phase 2: Implement rule evaluation engine
- Parse advanced config
- Evaluate rules on new events
- Store in retention_metadata
- Existing pruning still uses keep_days

Phase 3: Implement advanced pruning
- Use retention_metadata for pruning decisions
- Fallback to keep_days if no retention_metadata

Phase 4: Implement cap enforcement
- Add cap checking
- Score-based deletion

Phase 5: Optimization
- Incremental evaluation
- Background re-evaluation
- Performance tuning

Testing Strategy

Unit Tests
- Rule parsing and validation
- Condition evaluation (each gate type)
- Action application
- Score calculation
- Priority ordering

Integration Tests
- Full rule evaluation on test events
- Pruning with multiple rules
- Cap enforcement
- Mixed simple/advanced retention

Performance Tests
- Evaluate 100k events
- Prune 10k events
- Cap enforcement on 1M events
- Database query performance

Backward Compatibility

1. **Existing config continues to work**
   - If `advanced.enabled: false`, only `keep_days` is used
   - Existing pruning logic unchanged

2. **Graceful degradation**
   - If advanced config is invalid, fall back to simple mode
   - Log warning but don't fail

3. **No schema breaking changes**
   - Only additive changes (new table)
   - Existing tables unchanged

4. **API compatibility**
   - Retention metadata is internal implementation detail
   - No changes to external APIs or protocol servers
