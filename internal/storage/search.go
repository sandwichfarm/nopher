package storage

import (
	"context"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// QueryEventsWithSearch performs a NIP-50 compliant search when the Search field is present
// Falls back to regular QueryEvents if no search term is provided
func (s *Storage) QueryEventsWithSearch(ctx context.Context, filter nostr.Filter) ([]*nostr.Event, error) {
	// If no search term, use regular query
	if filter.Search == "" {
		return s.QueryEvents(ctx, filter)
	}

	// Get all matching events (without search filter first)
	// Store the search term and clear it temporarily
	searchTerm := filter.Search
	filter.Search = ""

	events, err := s.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter results by search term
	searchLower := strings.ToLower(searchTerm)
	var results []*nostr.Event

	for _, event := range events {
		if matchesSearch(event, searchLower) {
			results = append(results, event)
		}
	}

	// Rank by relevance (simple relevance scoring)
	rankByRelevance(results, searchLower)

	// Apply limit after relevance sorting (per NIP-50)
	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

// matchesSearch checks if an event matches the search term
func matchesSearch(event *nostr.Event, searchLower string) bool {
	// Search in content (primary field per NIP-50)
	if strings.Contains(strings.ToLower(event.Content), searchLower) {
		return true
	}

	// For profiles (kind 0), also search in parsed metadata
	// This provides better UX for profile searches
	if event.Kind == 0 {
		// Simple check in raw JSON - could be enhanced with proper parsing
		return strings.Contains(strings.ToLower(event.Content), searchLower)
	}

	return false
}

// rankByRelevance sorts events by search relevance
// Higher score = more relevant, appears first
func rankByRelevance(events []*nostr.Event, searchLower string) {
	// Calculate scores
	scores := make([]int, len(events))
	for i, event := range events {
		scores[i] = calculateRelevance(event, searchLower)
	}

	// Bubble sort by score (descending)
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if scores[j] > scores[i] {
				events[i], events[j] = events[j], events[i]
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
}

// calculateRelevance scores an event's relevance to search term
func calculateRelevance(event *nostr.Event, searchLower string) int {
	score := 0
	contentLower := strings.ToLower(event.Content)

	// Exact phrase match = highest score
	if contentLower == searchLower {
		score += 100
	} else if strings.Contains(contentLower, searchLower) {
		score += 50
	}

	// Count word matches
	searchWords := strings.Fields(searchLower)
	for _, word := range searchWords {
		if strings.Contains(contentLower, word) {
			score += 10
		}
	}

	// Bonus for shorter content (more focused)
	if len(event.Content) < 500 {
		score += 5
	}

	// Bonus for profiles (kind 0)
	if event.Kind == 0 {
		score += 10
	}

	return score
}
