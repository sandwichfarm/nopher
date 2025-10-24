package storage

import (
	"context"
	"fmt"
)

// GraphNode represents a node in the social graph
type GraphNode struct {
	RootPubkey string
	Pubkey     string
	Depth      int
	Mutual     bool
	LastSeen   int64
}

// SaveGraphNode stores or updates a graph node
func (s *Storage) SaveGraphNode(ctx context.Context, node *GraphNode) error {
	query := `
		INSERT INTO graph_nodes (root_pubkey, pubkey, depth, mutual, last_seen)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(root_pubkey, pubkey) DO UPDATE SET
			depth = excluded.depth,
			mutual = excluded.mutual,
			last_seen = excluded.last_seen
	`

	mutual := 0
	if node.Mutual {
		mutual = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		node.RootPubkey, node.Pubkey, node.Depth, mutual, node.LastSeen)
	if err != nil {
		return fmt.Errorf("failed to save graph node: %w", err)
	}

	return nil
}

// GetGraphNodes retrieves graph nodes for a given root pubkey
func (s *Storage) GetGraphNodes(ctx context.Context, rootPubkey string, maxDepth int) ([]*GraphNode, error) {
	query := `
		SELECT root_pubkey, pubkey, depth, mutual, last_seen
		FROM graph_nodes
		WHERE root_pubkey = ? AND depth <= ?
		ORDER BY depth, pubkey
	`

	rows, err := s.db.QueryContext(ctx, query, rootPubkey, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to query graph nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*GraphNode
	for rows.Next() {
		var node GraphNode
		var mutual int

		if err := rows.Scan(
			&node.RootPubkey, &node.Pubkey, &node.Depth, &mutual, &node.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan graph node: %w", err)
		}

		node.Mutual = mutual == 1
		nodes = append(nodes, &node)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return nodes, nil
}

// GetFollowingPubkeys returns the pubkeys being followed by the root
func (s *Storage) GetFollowingPubkeys(ctx context.Context, rootPubkey string) ([]string, error) {
	query := `
		SELECT pubkey
		FROM graph_nodes
		WHERE root_pubkey = ? AND depth = 1
		ORDER BY pubkey
	`

	rows, err := s.db.QueryContext(ctx, query, rootPubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query following pubkeys: %w", err)
	}
	defer rows.Close()

	var pubkeys []string
	for rows.Next() {
		var pubkey string
		if err := rows.Scan(&pubkey); err != nil {
			return nil, fmt.Errorf("failed to scan pubkey: %w", err)
		}
		pubkeys = append(pubkeys, pubkey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return pubkeys, nil
}

// GetMutualPubkeys returns the pubkeys with mutual follows
func (s *Storage) GetMutualPubkeys(ctx context.Context, rootPubkey string) ([]string, error) {
	query := `
		SELECT pubkey
		FROM graph_nodes
		WHERE root_pubkey = ? AND depth = 1 AND mutual = 1
		ORDER BY pubkey
	`

	rows, err := s.db.QueryContext(ctx, query, rootPubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to query mutual pubkeys: %w", err)
	}
	defer rows.Close()

	var pubkeys []string
	for rows.Next() {
		var pubkey string
		if err := rows.Scan(&pubkey); err != nil {
			return nil, fmt.Errorf("failed to scan pubkey: %w", err)
		}
		pubkeys = append(pubkeys, pubkey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return pubkeys, nil
}

// DeleteGraphNodes removes all graph nodes for a given root pubkey
func (s *Storage) DeleteGraphNodes(ctx context.Context, rootPubkey string) error {
	query := `DELETE FROM graph_nodes WHERE root_pubkey = ?`
	_, err := s.db.ExecContext(ctx, query, rootPubkey)
	if err != nil {
		return fmt.Errorf("failed to delete graph nodes: %w", err)
	}
	return nil
}
