package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RetentionMetadata represents retention metadata for an event
type RetentionMetadata struct {
	EventID         string
	RuleName        string
	RulePriority    int
	RetainUntil     *time.Time // nil = retain forever
	LastEvaluatedAt time.Time
	Score           int
	Protected       bool
}

// StoreRetentionMetadata stores or updates retention metadata for an event
func (s *Storage) StoreRetentionMetadata(ctx context.Context, meta *RetentionMetadata) error {
	var retainUntil *int64
	if meta.RetainUntil != nil {
		ts := meta.RetainUntil.Unix()
		retainUntil = &ts
	}

	query := `
		INSERT OR REPLACE INTO retention_metadata
		(event_id, rule_name, rule_priority, retain_until, last_evaluated_at, score, protected)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		meta.EventID,
		meta.RuleName,
		meta.RulePriority,
		retainUntil,
		meta.LastEvaluatedAt.Unix(),
		meta.Score,
		meta.Protected,
	)
	if err != nil {
		return fmt.Errorf("failed to store retention metadata: %w", err)
	}

	return nil
}

// GetRetentionMetadata retrieves retention metadata for an event
func (s *Storage) GetRetentionMetadata(ctx context.Context, eventID string) (*RetentionMetadata, error) {
	query := `
		SELECT event_id, rule_name, rule_priority, retain_until, last_evaluated_at, score, protected
		FROM retention_metadata
		WHERE event_id = ?
	`

	var meta RetentionMetadata
	var retainUntil *int64
	var lastEvaluatedAt int64

	err := s.db.QueryRowContext(ctx, query, eventID).Scan(
		&meta.EventID,
		&meta.RuleName,
		&meta.RulePriority,
		&retainUntil,
		&lastEvaluatedAt,
		&meta.Score,
		&meta.Protected,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No metadata found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get retention metadata: %w", err)
	}

	if retainUntil != nil {
		t := time.Unix(*retainUntil, 0)
		meta.RetainUntil = &t
	}
	meta.LastEvaluatedAt = time.Unix(lastEvaluatedAt, 0)

	return &meta, nil
}

// GetExpiredEvents returns event IDs that have passed their retain_until date
func (s *Storage) GetExpiredEvents(ctx context.Context, limit int) ([]string, error) {
	now := time.Now().Unix()
	query := `
		SELECT event_id
		FROM retention_metadata
		WHERE retain_until IS NOT NULL
		  AND retain_until < ?
		  AND protected = 0
		ORDER BY retain_until ASC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired events: %w", err)
	}
	defer rows.Close()

	var eventIDs []string
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, fmt.Errorf("failed to scan expired event: %w", err)
		}
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, rows.Err()
}

// GetEventsByScore returns events sorted by score (ascending - lowest priority first)
// Used for cap enforcement
func (s *Storage) GetEventsByScore(ctx context.Context, limit int) ([]*RetentionMetadata, error) {
	query := `
		SELECT event_id, rule_name, rule_priority, retain_until, last_evaluated_at, score, protected
		FROM retention_metadata
		WHERE protected = 0
		ORDER BY score ASC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by score: %w", err)
	}
	defer rows.Close()

	var results []*RetentionMetadata
	for rows.Next() {
		var meta RetentionMetadata
		var retainUntil *int64
		var lastEvaluatedAt int64

		err := rows.Scan(
			&meta.EventID,
			&meta.RuleName,
			&meta.RulePriority,
			&retainUntil,
			&lastEvaluatedAt,
			&meta.Score,
			&meta.Protected,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention metadata: %w", err)
		}

		if retainUntil != nil {
			t := time.Unix(*retainUntil, 0)
			meta.RetainUntil = &t
		}
		meta.LastEvaluatedAt = time.Unix(lastEvaluatedAt, 0)

		results = append(results, &meta)
	}

	return results, rows.Err()
}

// GetEventsNeedingEvaluation returns event IDs that don't have retention metadata yet
func (s *Storage) GetEventsNeedingEvaluation(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT e.id
		FROM events e
		LEFT JOIN retention_metadata rm ON e.id = rm.event_id
		WHERE rm.event_id IS NULL
		ORDER BY e.created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events needing evaluation: %w", err)
	}
	defer rows.Close()

	var eventIDs []string
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, fmt.Errorf("failed to scan event ID: %w", err)
		}
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, rows.Err()
}

// GetEventsForReEvaluation returns events that need re-evaluation (older than interval)
func (s *Storage) GetEventsForReEvaluation(ctx context.Context, olderThan time.Time, limit int) ([]string, error) {
	query := `
		SELECT event_id
		FROM retention_metadata
		WHERE last_evaluated_at < ?
		ORDER BY last_evaluated_at ASC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, olderThan.Unix(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for re-evaluation: %w", err)
	}
	defer rows.Close()

	var eventIDs []string
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, fmt.Errorf("failed to scan event ID: %w", err)
		}
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, rows.Err()
}

// CountRetentionStats returns retention statistics
func (s *Storage) CountRetentionStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total events with retention metadata
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM retention_metadata").Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count total retention metadata: %w", err)
	}
	stats["total_events"] = total

	// Protected events
	var protected int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM retention_metadata WHERE protected = 1").Scan(&protected)
	if err != nil {
		return nil, fmt.Errorf("failed to count protected events: %w", err)
	}
	stats["protected_events"] = protected

	// Events expiring within 7 days
	sevenDaysFromNow := time.Now().Add(7 * 24 * time.Hour).Unix()
	var expiringWithin7d int
	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM retention_metadata WHERE retain_until IS NOT NULL AND retain_until < ?",
		sevenDaysFromNow,
	).Scan(&expiringWithin7d)
	if err != nil {
		return nil, fmt.Errorf("failed to count expiring events: %w", err)
	}
	stats["expiring_within_7d"] = expiringWithin7d

	// By rule name
	byRule := make(map[string]int)
	rows, err := s.db.QueryContext(ctx, "SELECT rule_name, COUNT(*) FROM retention_metadata GROUP BY rule_name")
	if err != nil {
		return nil, fmt.Errorf("failed to query events by rule: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ruleName string
		var count int
		if err := rows.Scan(&ruleName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan rule count: %w", err)
		}
		byRule[ruleName] = count
	}
	stats["by_rule"] = byRule

	return stats, nil
}

// DeleteRetentionMetadata removes retention metadata for an event
func (s *Storage) DeleteRetentionMetadata(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM retention_metadata WHERE event_id = ?", eventID)
	if err != nil {
		return fmt.Errorf("failed to delete retention metadata: %w", err)
	}
	return nil
}
