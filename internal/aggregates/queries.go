package aggregates

import (
	"context"
	"fmt"
	"sort"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/sandwich/nophr/internal/config"
	"github.com/sandwich/nophr/internal/storage"
)

// QueryHelper provides helper methods for inbox/outbox queries
type QueryHelper struct {
	storage *storage.Storage
	config  *config.Config
	manager *Manager
}

// NewQueryHelper creates a new query helper
func NewQueryHelper(st *storage.Storage, cfg *config.Config, mgr *Manager) *QueryHelper {
	return &QueryHelper{
		storage: st,
		config:  cfg,
		manager: mgr,
	}
}

// getOwnerHex decodes the owner's npub to hex pubkey
func (qh *QueryHelper) getOwnerHex() (string, error) {
	if _, hex, err := nip19.Decode(qh.config.Identity.Npub); err != nil {
		return "", fmt.Errorf("failed to decode npub: %w", err)
	} else {
		return hex.(string), nil
	}
}

// GetOutboxNotes returns notes authored by the owner
func (qh *QueryHelper) GetOutboxNotes(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds:   []int{1}, // Notes
		Authors: []string{ownerHex},
		Limit:   limit,
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return qh.enrichEvents(ctx, events)
}

// GetInboxReplies returns replies to the owner's posts or mentions of the owner
func (qh *QueryHelper) GetInboxReplies(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	// Query notes that mention the owner
	filter := nostr.Filter{
		Kinds: []int{1},
		Tags: nostr.TagMap{
			"p": []string{ownerHex},
		},
		Limit: limit * 2, // Get more since we'll filter
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter to only actual replies (not just mentions)
	replies := make([]*nostr.Event, 0)
	for _, event := range events {
		if qh.manager.IsMentioning(ctx, event, ownerHex) {
			replies = append(replies, event)
		}
	}

	// Apply limit
	if len(replies) > limit {
		replies = replies[:limit]
	}

	return qh.enrichEvents(ctx, replies)
}

// GetInboxReactions returns reactions to the owner's posts
func (qh *QueryHelper) GetInboxReactions(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	// First get owner's notes
	ownerNotes, err := qh.GetOutboxNotes(ctx, 100)
	if err != nil {
		return nil, err
	}

	if len(ownerNotes) == 0 {
		return []*EnrichedEvent{}, nil
	}

	// Get IDs of owner's notes
	noteIDs := make([]string, 0, len(ownerNotes))
	for _, note := range ownerNotes {
		noteIDs = append(noteIDs, note.Event.ID)
	}

	// Query reactions to those notes
	filter := nostr.Filter{
		Kinds: []int{7},
		Tags: nostr.TagMap{
			"e": noteIDs,
		},
		Limit: limit,
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return qh.enrichEvents(ctx, events)
}

// GetThreadReplies returns all replies in a thread
func (qh *QueryHelper) GetThreadReplies(ctx context.Context, rootEventID string) ([]*EnrichedEvent, error) {
	filter := nostr.Filter{
		Kinds: []int{1},
		Tags: nostr.TagMap{
			"e": []string{rootEventID},
		},
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return qh.enrichEvents(ctx, events)
}

// GetThreadByEvent returns the full thread for a given event
func (qh *QueryHelper) GetThreadByEvent(ctx context.Context, eventID string) (*ThreadView, error) {
	// Get the event
	filter := nostr.Filter{
		IDs: []string{eventID},
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	event := events[0]

	// Determine root
	rootID, err := qh.manager.GetThreadRoot(ctx, event)
	if err != nil {
		rootID = eventID // Use event itself as root
	}

	// Get root event
	rootFilter := nostr.Filter{
		IDs: []string{rootID},
	}

	rootEvents, err := qh.storage.QueryEvents(ctx, rootFilter)
	if err != nil {
		return nil, err
	}

	var root *nostr.Event
	if len(rootEvents) > 0 {
		root = rootEvents[0]
	} else {
		root = event // Fallback
	}

	// Get all replies in thread
	replies, err := qh.GetThreadReplies(ctx, rootID)
	if err != nil {
		return nil, err
	}

	return &ThreadView{
		Root:    qh.enrichEvent(ctx, root),
		Replies: replies,
	}, nil
}

// enrichEvents adds aggregate data to events
func (qh *QueryHelper) enrichEvents(ctx context.Context, events []*nostr.Event) ([]*EnrichedEvent, error) {
	enriched := make([]*EnrichedEvent, 0, len(events))
	for _, event := range events {
		enriched = append(enriched, qh.enrichEvent(ctx, event))
	}
	return enriched, nil
}

// filterAndSortEvents applies content filtering and sorting based on config
func (qh *QueryHelper) filterAndSortEvents(enriched []*EnrichedEvent, sortMode string) []*EnrichedEvent {
	// Apply content filtering if enabled
	if qh.config.Behavior.ContentFiltering.Enabled {
		filtered := make([]*EnrichedEvent, 0)
		for _, e := range enriched {
			if qh.passesContentFilter(e) {
				filtered = append(filtered, e)
			}
		}
		enriched = filtered
	}

	// Apply sorting
	switch sortMode {
	case "engagement":
		sort.Slice(enriched, func(i, j int) bool {
			return enriched[i].Aggregates.InteractionScore() > enriched[j].Aggregates.InteractionScore()
		})
	case "zaps":
		sort.Slice(enriched, func(i, j int) bool {
			return enriched[i].Aggregates.ZapSatsTotal > enriched[j].Aggregates.ZapSatsTotal
		})
	case "reactions":
		sort.Slice(enriched, func(i, j int) bool {
			return enriched[i].Aggregates.ReactionTotal > enriched[j].Aggregates.ReactionTotal
		})
	case "chronological":
		fallthrough
	default:
		// Already in chronological order from query (newest first)
		// No additional sorting needed
	}

	return enriched
}

// passesContentFilter checks if an event passes content filtering rules
func (qh *QueryHelper) passesContentFilter(e *EnrichedEvent) bool {
	cfg := qh.config.Behavior.ContentFiltering

	// Check minimum reactions
	if cfg.MinReactions > 0 && e.Aggregates.ReactionTotal < cfg.MinReactions {
		return false
	}

	// Check minimum zap sats
	if cfg.MinZapSats > 0 && e.Aggregates.ZapSatsTotal < int64(cfg.MinZapSats) {
		return false
	}

	// Check minimum engagement (combined score)
	if cfg.MinEngagement > 0 && e.Aggregates.InteractionScore() < int64(cfg.MinEngagement) {
		return false
	}

	// Check hide no interactions
	if cfg.HideNoInteractions && !e.Aggregates.HasInteractions() {
		return false
	}

	// Content type filtering would go here if needed
	// For now, we don't filter by content type

	return true
}

// enrichEvent adds aggregate data to a single event
func (qh *QueryHelper) enrichEvent(ctx context.Context, event *nostr.Event) *EnrichedEvent {
	agg, _ := qh.manager.GetEventAggregates(ctx, event.ID)
	if agg == nil {
		agg = &EventAggregates{EventID: event.ID}
	}

	return &EnrichedEvent{
		Event:      event,
		Aggregates: agg,
	}
}

// GetPopularNotes returns notes sorted by interaction score
func (qh *QueryHelper) GetPopularNotes(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	// Get recent notes
	filter := nostr.Filter{
		Kinds: []int{1},
		Limit: limit * 10, // Get more to sort
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	enriched, err := qh.enrichEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	// Sort by interaction score
	sort.Slice(enriched, func(i, j int) bool {
		return enriched[i].Aggregates.InteractionScore() > enriched[j].Aggregates.InteractionScore()
	})

	// Apply limit
	if len(enriched) > limit {
		enriched = enriched[:limit]
	}

	return enriched, nil
}

// EnrichedEvent contains an event with its aggregate data
type EnrichedEvent struct {
	Event      *nostr.Event
	Aggregates *EventAggregates
}

// ThreadView represents a full thread with root and replies
type ThreadView struct {
	Root    *EnrichedEvent
	Replies []*EnrichedEvent
}

// === Public Section-Based Query Methods ===
// These map to user-facing sections as per design docs

// GetNotes returns owner's notes (kind 1, non-replies only)
func (qh *QueryHelper) GetNotes(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	// Get all owner's kind 1 events
	filter := nostr.Filter{
		Kinds:   []int{1},
		Authors: []string{ownerHex},
		Limit:   limit * 2, // Get more since we'll filter out replies
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter out replies - only root notes
	notes := make([]*nostr.Event, 0)
	for _, event := range events {
		threadInfo, err := ParseThreadInfo(event)
		if err != nil {
			continue
		}
		if !threadInfo.IsReply() {
			notes = append(notes, event)
		}
	}

	enriched, err := qh.enrichEvents(ctx, notes)
	if err != nil {
		return nil, err
	}

	// Apply filtering and sorting
	enriched = qh.filterAndSortEvents(enriched, qh.config.Behavior.SortPreferences.Notes)

	// Apply limit after filtering
	if len(enriched) > limit {
		enriched = enriched[:limit]
	}

	return enriched, nil
}

// GetArticles returns owner's long-form articles (kind 30023)
func (qh *QueryHelper) GetArticles(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds:   []int{30023},
		Authors: []string{ownerHex},
		Limit:   limit,
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	enriched, err := qh.enrichEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	// Apply filtering and sorting
	enriched = qh.filterAndSortEvents(enriched, qh.config.Behavior.SortPreferences.Articles)

	// Apply limit after filtering
	if len(enriched) > limit {
		enriched = enriched[:limit]
	}

	return enriched, nil
}

// GetReplies returns replies to owner's content
// This queries for events that mention the owner and are actual replies
func (qh *QueryHelper) GetReplies(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	// Query notes that mention the owner
	filter := nostr.Filter{
		Kinds: []int{1},
		Tags: nostr.TagMap{
			"p": []string{ownerHex},
		},
		Limit: limit * 2, // Get more since we'll filter
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter to only actual replies (have e tags)
	replies := make([]*nostr.Event, 0)
	for _, event := range events {
		threadInfo, err := ParseThreadInfo(event)
		if err != nil {
			continue
		}
		// A reply must have a ReplyToID (e tag)
		if threadInfo.IsReply() && qh.manager.IsMentioning(ctx, event, ownerHex) {
			replies = append(replies, event)
		}
	}

	enriched, err := qh.enrichEvents(ctx, replies)
	if err != nil {
		return nil, err
	}

	// Apply filtering and sorting
	enriched = qh.filterAndSortEvents(enriched, qh.config.Behavior.SortPreferences.Replies)

	// Apply limit after filtering
	if len(enriched) > limit {
		enriched = enriched[:limit]
	}

	return enriched, nil
}

// GetMentions returns posts that mention the owner (including non-reply mentions)
func (qh *QueryHelper) GetMentions(ctx context.Context, limit int) ([]*EnrichedEvent, error) {
	ownerHex, err := qh.getOwnerHex()
	if err != nil {
		return nil, err
	}

	// Query notes that mention the owner
	filter := nostr.Filter{
		Kinds: []int{1},
		Tags: nostr.TagMap{
			"p": []string{ownerHex},
		},
		Limit: limit,
	}

	events, err := qh.storage.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	enriched, err := qh.enrichEvents(ctx, events)
	if err != nil {
		return nil, err
	}

	// Apply filtering and sorting
	enriched = qh.filterAndSortEvents(enriched, qh.config.Behavior.SortPreferences.Mentions)

	// Apply limit after filtering
	if len(enriched) > limit {
		enriched = enriched[:limit]
	}

	// Return all mentions (both replies and non-reply mentions)
	return enriched, nil
}
