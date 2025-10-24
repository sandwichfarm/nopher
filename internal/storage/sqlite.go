package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fiatjaf/eventstore/sqlite3"
	"github.com/fiatjaf/khatru"
	_ "github.com/mattn/go-sqlite3"
)

// initSQLite initializes the SQLite backend with Khatru
func (s *Storage) initSQLite(ctx context.Context) error {
	// Ensure the directory exists
	dbPath := s.config.SQLitePath
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Initialize SQLite eventstore for Khatru
	db := &sqlite3.SQLite3Backend{
		DatabaseURL: dbPath,
	}

	if err := db.Init(); err != nil {
		return fmt.Errorf("failed to initialize SQLite eventstore: %w", err)
	}

	// Create Khatru relay instance
	relay := khatru.NewRelay()
	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)

	s.relay = relay

	// Open a separate connection for custom tables
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database for custom tables: %w", err)
	}

	// Enable foreign keys and optimize for performance
	if _, err := sqlDB.ExecContext(ctx, `
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = -64000;
		PRAGMA temp_store = MEMORY;
	`); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to configure SQLite: %w", err)
	}

	s.db = sqlDB
	return nil
}
