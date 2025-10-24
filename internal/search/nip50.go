package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// Relay interface defines the minimal relay operations needed for search
type Relay interface {
	QuerySync(ctx context.Context, filter nostr.Filter) ([]*nostr.Event, error)
}

// NIP50Engine provides NIP-50 compliant search capabilities
type NIP50Engine struct {
	relay Relay
}

// NewNIP50Engine creates a new NIP-50 search engine
func NewNIP50Engine(relay Relay) *NIP50Engine {
	return &NIP50Engine{relay: relay}
}

// Search performs a NIP-50 compliant search query
// The search field is passed directly to the relay which handles the actual search implementation
func (e *NIP50Engine) Search(ctx context.Context, searchText string, opts ...SearchOption) ([]*nostr.Event, error) {
	if searchText == "" {
		return nil, fmt.Errorf("search text cannot be empty")
	}

	// Build filter with search field (NIP-50)
	filter := nostr.Filter{
		Search: searchText,
		Limit:  100, // Default limit
	}

	// Apply options
	for _, opt := range opts {
		opt(&filter)
	}

	// Query the relay
	// The relay is responsible for:
	// 1. Matching against content field (and optionally other fields)
	// 2. Ranking results by relevance
	// 3. Applying the limit after relevance sorting
	events, err := e.relay.QuerySync(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}

	return events, nil
}

// SearchOption allows configuring the search filter
type SearchOption func(*nostr.Filter)

// WithKinds restricts search to specific event kinds
func WithKinds(kinds ...int) SearchOption {
	return func(f *nostr.Filter) {
		f.Kinds = kinds
	}
}

// WithAuthors restricts search to specific authors
func WithAuthors(pubkeys ...string) SearchOption {
	return func(f *nostr.Filter) {
		f.Authors = pubkeys
	}
}

// WithLimit sets the maximum number of results
func WithLimit(limit int) SearchOption {
	return func(f *nostr.Filter) {
		f.Limit = limit
	}
}

// WithSince sets the minimum timestamp
func WithSince(since nostr.Timestamp) SearchOption {
	return func(f *nostr.Filter) {
		f.Since = &since
	}
}

// WithUntil sets the maximum timestamp
func WithUntil(until nostr.Timestamp) SearchOption {
	return func(f *nostr.Filter) {
		f.Until = &until
	}
}

// ParseSearchQuery provides helper functionality for common search patterns
// Returns appropriate search options based on query analysis
func ParseSearchQuery(query string) (searchText string, opts []SearchOption) {
	query = strings.TrimSpace(query)

	// Check for kind-specific search (e.g., "kind:1 bitcoin")
	if strings.Contains(query, "kind:") {
		parts := strings.Fields(query)
		var kinds []int
		var textParts []string

		for _, part := range parts {
			if strings.HasPrefix(part, "kind:") {
				kindStr := strings.TrimPrefix(part, "kind:")
				var kind int
				fmt.Sscanf(kindStr, "%d", &kind)
				if kind > 0 {
					kinds = append(kinds, kind)
				}
			} else {
				textParts = append(textParts, part)
			}
		}

		if len(kinds) > 0 {
			opts = append(opts, WithKinds(kinds...))
		}
		searchText = strings.Join(textParts, " ")
	} else {
		searchText = query
	}

	return searchText, opts
}
