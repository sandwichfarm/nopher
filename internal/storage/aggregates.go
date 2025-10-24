package storage

import (
	"context"
	"encoding/json"
	"fmt"
)

// Aggregate represents interaction rollups for an event
type Aggregate struct {
	EventID           string
	ReplyCount        int
	ReactionTotal     int
	ReactionCounts    map[string]int
	ZapSatsTotal      int64
	LastInteractionAt int64
}

// SaveAggregate stores or updates an aggregate
func (s *Storage) SaveAggregate(ctx context.Context, agg *Aggregate) error {
	reactionCountsJSON, err := json.Marshal(agg.ReactionCounts)
	if err != nil {
		return fmt.Errorf("failed to marshal reaction counts: %w", err)
	}

	query := `
		INSERT INTO aggregates (
			event_id, reply_count, reaction_total, reaction_counts_json,
			zap_sats_total, last_interaction_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(event_id) DO UPDATE SET
			reply_count = excluded.reply_count,
			reaction_total = excluded.reaction_total,
			reaction_counts_json = excluded.reaction_counts_json,
			zap_sats_total = excluded.zap_sats_total,
			last_interaction_at = excluded.last_interaction_at
	`

	_, err = s.db.ExecContext(ctx, query,
		agg.EventID, agg.ReplyCount, agg.ReactionTotal, string(reactionCountsJSON),
		agg.ZapSatsTotal, agg.LastInteractionAt)
	if err != nil {
		return fmt.Errorf("failed to save aggregate: %w", err)
	}

	return nil
}

// GetAggregate retrieves an aggregate for a given event ID
func (s *Storage) GetAggregate(ctx context.Context, eventID string) (*Aggregate, error) {
	query := `
		SELECT event_id, reply_count, reaction_total, reaction_counts_json,
		       zap_sats_total, last_interaction_at
		FROM aggregates
		WHERE event_id = ?
	`

	var agg Aggregate
	var reactionCountsJSON string

	err := s.db.QueryRowContext(ctx, query, eventID).Scan(
		&agg.EventID, &agg.ReplyCount, &agg.ReactionTotal, &reactionCountsJSON,
		&agg.ZapSatsTotal, &agg.LastInteractionAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate: %w", err)
	}

	if reactionCountsJSON != "" {
		if err := json.Unmarshal([]byte(reactionCountsJSON), &agg.ReactionCounts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reaction counts: %w", err)
		}
	} else {
		agg.ReactionCounts = make(map[string]int)
	}

	return &agg, nil
}

// GetAggregates retrieves aggregates for multiple event IDs
func (s *Storage) GetAggregates(ctx context.Context, eventIDs []string) (map[string]*Aggregate, error) {
	if len(eventIDs) == 0 {
		return make(map[string]*Aggregate), nil
	}

	// Build placeholders for the IN clause
	placeholders := ""
	args := make([]interface{}, len(eventIDs))
	for i, id := range eventIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT event_id, reply_count, reaction_total, reaction_counts_json,
		       zap_sats_total, last_interaction_at
		FROM aggregates
		WHERE event_id IN (%s)
	`, placeholders)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregates: %w", err)
	}
	defer rows.Close()

	aggregates := make(map[string]*Aggregate)
	for rows.Next() {
		var agg Aggregate
		var reactionCountsJSON string

		if err := rows.Scan(
			&agg.EventID, &agg.ReplyCount, &agg.ReactionTotal, &reactionCountsJSON,
			&agg.ZapSatsTotal, &agg.LastInteractionAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan aggregate: %w", err)
		}

		if reactionCountsJSON != "" {
			if err := json.Unmarshal([]byte(reactionCountsJSON), &agg.ReactionCounts); err != nil {
				return nil, fmt.Errorf("failed to unmarshal reaction counts: %w", err)
			}
		} else {
			agg.ReactionCounts = make(map[string]int)
		}

		aggregates[agg.EventID] = &agg
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return aggregates, nil
}

// IncrementReplyCount increments the reply count for an event
func (s *Storage) IncrementReplyCount(ctx context.Context, eventID string, interactionAt int64) error {
	query := `
		INSERT INTO aggregates (event_id, reply_count, reaction_total, zap_sats_total, last_interaction_at)
		VALUES (?, 1, 0, 0, ?)
		ON CONFLICT(event_id) DO UPDATE SET
			reply_count = reply_count + 1,
			last_interaction_at = MAX(last_interaction_at, excluded.last_interaction_at)
	`

	_, err := s.db.ExecContext(ctx, query, eventID, interactionAt)
	if err != nil {
		return fmt.Errorf("failed to increment reply count: %w", err)
	}

	return nil
}

// IncrementReaction increments the reaction count for an event
func (s *Storage) IncrementReaction(ctx context.Context, eventID string, reaction string, interactionAt int64) error {
	// Get current aggregate
	agg, err := s.GetAggregate(ctx, eventID)
	if err != nil {
		// Create new aggregate
		agg = &Aggregate{
			EventID:           eventID,
			ReactionCounts:    make(map[string]int),
			LastInteractionAt: interactionAt,
		}
	}

	// Increment reaction count
	agg.ReactionCounts[reaction]++
	agg.ReactionTotal++
	if interactionAt > agg.LastInteractionAt {
		agg.LastInteractionAt = interactionAt
	}

	return s.SaveAggregate(ctx, agg)
}

// AddZapAmount adds zap sats to an event's aggregate
func (s *Storage) AddZapAmount(ctx context.Context, eventID string, sats int64, interactionAt int64) error {
	query := `
		INSERT INTO aggregates (event_id, reply_count, reaction_total, zap_sats_total, last_interaction_at)
		VALUES (?, 0, 0, ?, ?)
		ON CONFLICT(event_id) DO UPDATE SET
			zap_sats_total = zap_sats_total + excluded.zap_sats_total,
			last_interaction_at = MAX(last_interaction_at, excluded.last_interaction_at)
	`

	_, err := s.db.ExecContext(ctx, query, eventID, sats, interactionAt)
	if err != nil {
		return fmt.Errorf("failed to add zap amount: %w", err)
	}

	return nil
}

// DeleteAggregate removes an aggregate
func (s *Storage) DeleteAggregate(ctx context.Context, eventID string) error {
	query := `DELETE FROM aggregates WHERE event_id = ?`
	_, err := s.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete aggregate: %w", err)
	}
	return nil
}

// BatchIncrementReplies increments reply counts for multiple events (Performance optimization)
func (s *Storage) BatchIncrementReplies(ctx context.Context, updates map[string]int64) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO aggregates (event_id, reply_count, reaction_total, zap_sats_total, last_interaction_at)
		VALUES (?, 1, 0, 0, ?)
		ON CONFLICT(event_id) DO UPDATE SET
			reply_count = reply_count + 1,
			last_interaction_at = MAX(last_interaction_at, excluded.last_interaction_at)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for eventID, interactionAt := range updates {
		if _, err := stmt.ExecContext(ctx, eventID, interactionAt); err != nil {
			return fmt.Errorf("failed to increment reply for %s: %w", eventID, err)
		}
	}

	return tx.Commit()
}

// BatchAddZaps adds zap amounts for multiple events (Performance optimization)
func (s *Storage) BatchAddZaps(ctx context.Context, updates map[string]struct {
	Sats          int64
	InteractionAt int64
}) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO aggregates (event_id, reply_count, reaction_total, zap_sats_total, last_interaction_at)
		VALUES (?, 0, 0, ?, ?)
		ON CONFLICT(event_id) DO UPDATE SET
			zap_sats_total = zap_sats_total + excluded.zap_sats_total,
			last_interaction_at = MAX(last_interaction_at, excluded.last_interaction_at)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for eventID, update := range updates {
		if _, err := stmt.ExecContext(ctx, eventID, update.Sats, update.InteractionAt); err != nil {
			return fmt.Errorf("failed to add zap for %s: %w", eventID, err)
		}
	}

	return tx.Commit()
}

// BatchIncrementReactions increments reaction counts for multiple events (Performance optimization)
func (s *Storage) BatchIncrementReactions(ctx context.Context, updates map[string]map[string]int64) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for eventID, reactions := range updates {
		// Get current aggregate
		agg, err := s.GetAggregate(ctx, eventID)
		if err != nil {
			// Create new aggregate
			agg = &Aggregate{
				EventID:        eventID,
				ReactionCounts: make(map[string]int),
			}
		}

		// Increment reaction counts
		for reaction, interactionAt := range reactions {
			agg.ReactionCounts[reaction]++
			agg.ReactionTotal++
			if interactionAt > agg.LastInteractionAt {
				agg.LastInteractionAt = interactionAt
			}
		}

		// Save updated aggregate
		if err := s.SaveAggregate(ctx, agg); err != nil {
			return fmt.Errorf("failed to save aggregate for %s: %w", eventID, err)
		}
	}

	return tx.Commit()
}
