package storage

import (
	"context"
	"fmt"
)

// runMigrations creates the custom tables for nophr
func (s *Storage) runMigrations(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	migrations := []string{
		// relay_hints: Track which relays to use for each author (from NIP-65)
		`CREATE TABLE IF NOT EXISTS relay_hints (
			pubkey TEXT NOT NULL,
			relay TEXT NOT NULL,
			can_read INTEGER NOT NULL DEFAULT 1,
			can_write INTEGER NOT NULL DEFAULT 1,
			freshness INTEGER NOT NULL,
			last_seen_event_id TEXT NOT NULL,
			PRIMARY KEY (pubkey, relay)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_relay_hints_pubkey_freshness
		 ON relay_hints(pubkey, freshness DESC)`,

		// graph_nodes: Owner-centric social graph cache
		`CREATE TABLE IF NOT EXISTS graph_nodes (
			root_pubkey TEXT NOT NULL,
			pubkey TEXT NOT NULL,
			depth INTEGER NOT NULL,
			mutual INTEGER NOT NULL DEFAULT 0,
			last_seen INTEGER NOT NULL,
			PRIMARY KEY (root_pubkey, pubkey)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_nodes_root_depth_mutual
		 ON graph_nodes(root_pubkey, depth, mutual)`,

		// sync_state: Cursor tracking per relay/kind
		`CREATE TABLE IF NOT EXISTS sync_state (
			relay TEXT NOT NULL,
			kind INTEGER NOT NULL,
			since INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (relay, kind)
		)`,

		// aggregates: Interaction rollups (reply counts, reactions, zaps)
		`CREATE TABLE IF NOT EXISTS aggregates (
			event_id TEXT PRIMARY KEY,
			reply_count INTEGER NOT NULL DEFAULT 0,
			reaction_total INTEGER NOT NULL DEFAULT 0,
			reaction_counts_json TEXT,
			zap_sats_total INTEGER NOT NULL DEFAULT 0,
			last_interaction_at INTEGER NOT NULL
		)`,
	}

	for i, migration := range migrations {
		if _, err := s.db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}
