package storage

import (
	"context"
	"fmt"
)

// RelayHint represents a relay hint from NIP-65 (kind 10002)
type RelayHint struct {
	Pubkey          string
	Relay           string
	CanRead         bool
	CanWrite        bool
	Freshness       int64
	LastSeenEventID string
}

// SaveRelayHint stores or updates a relay hint
func (s *Storage) SaveRelayHint(ctx context.Context, hint *RelayHint) error {
	query := `
		INSERT INTO relay_hints (pubkey, relay, can_read, can_write, freshness, last_seen_event_id)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(pubkey, relay) DO UPDATE SET
			can_read = excluded.can_read,
			can_write = excluded.can_write,
			freshness = excluded.freshness,
			last_seen_event_id = excluded.last_seen_event_id
		WHERE excluded.freshness > freshness
	`

	canRead := 0
	if hint.CanRead {
		canRead = 1
	}
	canWrite := 0
	if hint.CanWrite {
		canWrite = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		hint.Pubkey, hint.Relay, canRead, canWrite, hint.Freshness, hint.LastSeenEventID)
	if err != nil {
		return fmt.Errorf("failed to save relay hint: %w", err)
	}

	return nil
}

// GetRelayHints retrieves relay hints for a given pubkey
func (s *Storage) GetRelayHints(ctx context.Context, pubkey string) ([]*RelayHint, error) {
	query := `
		SELECT pubkey, relay, can_read, can_write, freshness, last_seen_event_id
		FROM relay_hints
		WHERE pubkey = ?
		ORDER BY freshness DESC
	`

	rows, err := s.db.QueryContext(ctx, query, pubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query relay hints: %w", err)
	}
	defer rows.Close()

	var hints []*RelayHint
	for rows.Next() {
		var hint RelayHint
		var canRead, canWrite int

		if err := rows.Scan(
			&hint.Pubkey, &hint.Relay, &canRead, &canWrite,
			&hint.Freshness, &hint.LastSeenEventID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relay hint: %w", err)
		}

		hint.CanRead = canRead == 1
		hint.CanWrite = canWrite == 1
		hints = append(hints, &hint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return hints, nil
}

// GetWriteRelays returns the write relays for a given pubkey
func (s *Storage) GetWriteRelays(ctx context.Context, pubkey string) ([]string, error) {
	query := `
		SELECT relay
		FROM relay_hints
		WHERE pubkey = ? AND can_write = 1
		ORDER BY freshness DESC
	`

	rows, err := s.db.QueryContext(ctx, query, pubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query write relays: %w", err)
	}
	defer rows.Close()

	var relays []string
	for rows.Next() {
		var relay string
		if err := rows.Scan(&relay); err != nil {
			return nil, fmt.Errorf("failed to scan relay: %w", err)
		}
		relays = append(relays, relay)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return relays, nil
}

// GetReadRelays returns the read relays for a given pubkey
func (s *Storage) GetReadRelays(ctx context.Context, pubkey string) ([]string, error) {
	query := `
		SELECT relay
		FROM relay_hints
		WHERE pubkey = ? AND can_read = 1
		ORDER BY freshness DESC
	`

	rows, err := s.db.QueryContext(ctx, query, pubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query read relays: %w", err)
	}
	defer rows.Close()

	var relays []string
	for rows.Next() {
		var relay string
		if err := rows.Scan(&relay); err != nil {
			return nil, fmt.Errorf("failed to scan relay: %w", err)
		}
		relays = append(relays, relay)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return relays, nil
}

// DeleteRelayHints removes all relay hints for a given pubkey
func (s *Storage) DeleteRelayHints(ctx context.Context, pubkey string) error {
	query := `DELETE FROM relay_hints WHERE pubkey = ?`
	_, err := s.db.ExecContext(ctx, query, pubkey)
	if err != nil {
		return fmt.Errorf("failed to delete relay hints: %w", err)
	}
	return nil
}
