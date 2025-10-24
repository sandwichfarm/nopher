package storage

import (
	"context"
	"fmt"
	"time"
)

// SyncState represents the synchronization cursor for a relay/kind pair
type SyncState struct {
	Relay     string
	Kind      int
	Since     int64
	UpdatedAt int64
}

// SaveSyncState stores or updates a sync state cursor
func (s *Storage) SaveSyncState(ctx context.Context, state *SyncState) error {
	query := `
		INSERT INTO sync_state (relay, kind, since, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(relay, kind) DO UPDATE SET
			since = excluded.since,
			updated_at = excluded.updated_at
	`

	_, err := s.db.ExecContext(ctx, query,
		state.Relay, state.Kind, state.Since, state.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save sync state: %w", err)
	}

	return nil
}

// GetSyncState retrieves the sync state for a relay/kind pair
func (s *Storage) GetSyncState(ctx context.Context, relay string, kind int) (*SyncState, error) {
	query := `
		SELECT relay, kind, since, updated_at
		FROM sync_state
		WHERE relay = ? AND kind = ?
	`

	var state SyncState
	err := s.db.QueryRowContext(ctx, query, relay, kind).Scan(
		&state.Relay, &state.Kind, &state.Since, &state.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}

	return &state, nil
}

// UpdateSyncCursor updates the since cursor for a relay/kind pair
func (s *Storage) UpdateSyncCursor(ctx context.Context, relay string, kind int, since int64) error {
	state := &SyncState{
		Relay:     relay,
		Kind:      kind,
		Since:     since,
		UpdatedAt: time.Now().Unix(),
	}
	return s.SaveSyncState(ctx, state)
}

// GetAllSyncStates retrieves all sync states
func (s *Storage) GetAllSyncStates(ctx context.Context) ([]*SyncState, error) {
	query := `
		SELECT relay, kind, since, updated_at
		FROM sync_state
		ORDER BY relay, kind
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync states: %w", err)
	}
	defer rows.Close()

	var states []*SyncState
	for rows.Next() {
		var state SyncState
		if err := rows.Scan(
			&state.Relay, &state.Kind, &state.Since, &state.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sync state: %w", err)
		}
		states = append(states, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return states, nil
}

// DeleteSyncState removes sync state for a relay/kind pair
func (s *Storage) DeleteSyncState(ctx context.Context, relay string, kind int) error {
	query := `DELETE FROM sync_state WHERE relay = ? AND kind = ?`
	_, err := s.db.ExecContext(ctx, query, relay, kind)
	if err != nil {
		return fmt.Errorf("failed to delete sync state: %w", err)
	}
	return nil
}
