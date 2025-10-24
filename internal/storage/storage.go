package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nophr/internal/config"
)

// Storage provides the main storage interface for nophr
type Storage struct {
	relay  *khatru.Relay
	db     *sql.DB
	config *config.Storage
}

// New creates a new Storage instance with the given configuration
func New(ctx context.Context, cfg *config.Storage) (*Storage, error) {
	s := &Storage{
		config: cfg,
	}

	// Initialize the appropriate backend
	switch cfg.Driver {
	case "sqlite":
		if err := s.initSQLite(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize SQLite: %w", err)
		}
	case "lmdb":
		if err := s.initLMDB(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize LMDB: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}

	// Run migrations for custom tables
	if err := s.runMigrations(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Relay returns the underlying Khatru relay instance
func (s *Storage) Relay() *khatru.Relay {
	return s.relay
}

// DB returns the underlying database connection (for custom tables)
func (s *Storage) DB() *sql.DB {
	return s.db
}

// StoreEvent stores an event in the Khatru relay
func (s *Storage) StoreEvent(ctx context.Context, event *nostr.Event) error {
	if s.relay == nil {
		return fmt.Errorf("relay not initialized")
	}

	// Call all StoreEvent handlers
	for _, handler := range s.relay.StoreEvent {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}
	}

	return nil
}

// QueryEvents queries events from the Khatru relay using Nostr filters
func (s *Storage) QueryEvents(ctx context.Context, filter nostr.Filter) ([]*nostr.Event, error) {
	if s.relay == nil {
		return nil, fmt.Errorf("relay not initialized")
	}

	// Use the first QueryEvents handler (eventstore)
	if len(s.relay.QueryEvents) == 0 {
		return nil, fmt.Errorf("no query handlers configured")
	}

	ch, err := s.relay.QueryEvents[0](ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	// Collect events from channel
	var events []*nostr.Event
	for event := range ch {
		events = append(events, event)
	}

	return events, nil
}

// QuerySync is a synchronous query adapter (implements search.Relay interface)
func (s *Storage) QuerySync(ctx context.Context, filter nostr.Filter) ([]*nostr.Event, error) {
	// Use QueryEventsWithSearch to support NIP-50
	return s.QueryEventsWithSearch(ctx, filter)
}

// Close closes the storage connections
func (s *Storage) Close() error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	return nil
}
