package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// Driver returns the storage driver name
func (s *Storage) Driver() string {
	return s.config.Driver
}

// CountEvents returns the total number of events in storage
func (s *Storage) CountEvents(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM event"

	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// CountEventsByKind returns event counts grouped by kind
func (s *Storage) CountEventsByKind(ctx context.Context) (map[int]int64, error) {
	counts := make(map[int]int64)

	query := "SELECT kind, COUNT(*) FROM event GROUP BY kind"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query event counts by kind: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var kind int
		var count int64
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		counts[kind] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return counts, nil
}

// DatabaseSize returns the database size in MB
func (s *Storage) DatabaseSize(ctx context.Context) (float64, error) {
	var path string

	switch s.config.Driver {
	case "sqlite":
		path = s.config.SQLitePath
	case "lmdb":
		path = s.config.LMDBPath
	default:
		return 0, fmt.Errorf("unsupported driver: %s", s.config.Driver)
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat database file: %w", err)
	}

	sizeMB := float64(info.Size()) / 1024 / 1024
	return sizeMB, nil
}

// EventTimeRange returns the oldest and newest event timestamps
func (s *Storage) EventTimeRange(ctx context.Context) (*time.Time, *time.Time, error) {
	var oldestUnix, newestUnix sql.NullInt64

	query := "SELECT MIN(created_at), MAX(created_at) FROM event"
	err := s.db.QueryRowContext(ctx, query).Scan(&oldestUnix, &newestUnix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query event time range: %w", err)
	}

	var oldest, newest *time.Time

	if oldestUnix.Valid {
		t := time.Unix(oldestUnix.Int64, 0)
		oldest = &t
	}

	if newestUnix.Valid {
		t := time.Unix(newestUnix.Int64, 0)
		newest = &t
	}

	return oldest, newest, nil
}

// CountEventsByRelay returns the number of events synced from a specific relay
func (s *Storage) CountEventsByRelay(ctx context.Context, relayURL string) (int64, error) {
	var count int64

	// This requires tracking relay source in sync_state or a separate table
	// For now, return 0 as placeholder
	// TODO: Implement relay tracking in sync engine

	return count, nil
}

// CursorInfo represents cursor information
type CursorInfo struct {
	Relay    string
	Kind     int
	Position int64
	Updated  time.Time
}

// GetAllCursors returns all cursor information
func (s *Storage) GetAllCursors(ctx context.Context) ([]CursorInfo, error) {
	var cursors []CursorInfo

	query := `
		SELECT relay, kind, cursor, updated_at
		FROM sync_state
		ORDER BY relay, kind
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cursors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c CursorInfo
		var updatedUnix int64

		if err := rows.Scan(&c.Relay, &c.Kind, &c.Position, &updatedUnix); err != nil {
			return nil, fmt.Errorf("failed to scan cursor: %w", err)
		}

		c.Updated = time.Unix(updatedUnix, 0)
		cursors = append(cursors, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cursors: %w", err)
	}

	return cursors, nil
}

// CountAggregates returns the total number of aggregates
func (s *Storage) CountAggregates(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM aggregates"

	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count aggregates: %w", err)
	}

	return count, nil
}

// CountAggregatesByKind returns aggregate counts grouped by kind
func (s *Storage) CountAggregatesByKind(ctx context.Context) (map[int]int64, error) {
	counts := make(map[int]int64)

	query := "SELECT kind, COUNT(*) FROM aggregates GROUP BY kind"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregate counts by kind: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var kind int
		var count int64
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		counts[kind] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return counts, nil
}

// LastReconcileTime returns the last time aggregates were reconciled
func (s *Storage) LastReconcileTime(ctx context.Context) (*time.Time, error) {
	var lastReconcileUnix sql.NullInt64

	query := "SELECT MAX(last_reconciled_at) FROM aggregates"
	err := s.db.QueryRowContext(ctx, query).Scan(&lastReconcileUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to query last reconcile time: %w", err)
	}

	if !lastReconcileUnix.Valid {
		return nil, nil
	}

	t := time.Unix(lastReconcileUnix.Int64, 0)
	return &t, nil
}

// GetEventByID retrieves an event by its ID
func (s *Storage) GetEventByID(ctx context.Context, id string) (*nostr.Event, error) {
	filter := nostr.Filter{
		IDs: []string{id},
	}

	events, err := s.QueryEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("event not found: %s", id)
	}

	return events[0], nil
}

// GetEventsByAuthor retrieves events by author pubkey
func (s *Storage) GetEventsByAuthor(ctx context.Context, pubkey string, limit int) ([]*nostr.Event, error) {
	filter := nostr.Filter{
		Authors: []string{pubkey},
		Limit:   limit,
	}

	return s.QueryEvents(ctx, filter)
}

// GetEventsByKind retrieves events by kind
func (s *Storage) GetEventsByKind(ctx context.Context, kind int, limit int) ([]*nostr.Event, error) {
	filter := nostr.Filter{
		Kinds: []int{kind},
		Limit: limit,
	}

	return s.QueryEvents(ctx, filter)
}

// DeleteEventsBefore deletes events created before the given timestamp
func (s *Storage) DeleteEventsBefore(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM event WHERE created_at < ?",
		before.Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to delete events: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return deleted, nil
}

// DeleteEventsByKind deletes all events of a specific kind
func (s *Storage) DeleteEventsByKind(ctx context.Context, kind int) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM event WHERE kind = ?",
		kind)
	if err != nil {
		return 0, fmt.Errorf("failed to delete events: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return deleted, nil
}
