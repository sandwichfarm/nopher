package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sandwich/nopher/internal/config"
	internalnostr "github.com/sandwich/nopher/internal/nostr"
	"github.com/sandwich/nopher/internal/storage"
)

// Engine manages the synchronization of events from Nostr relays
type Engine struct {
	config        *config.Config
	storage       *storage.Storage
	nostrClient   *internalnostr.Client
	discovery     *internalnostr.Discovery
	filterBuilder *FilterBuilder
	graph         *Graph
	cursors       *CursorManager

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Channels for coordination
	eventChan chan *nostr.Event
}

// New creates a new sync engine
func New(ctx context.Context, cfg *config.Config, st *storage.Storage, client *internalnostr.Client) *Engine {
	engineCtx, cancel := context.WithCancel(ctx)

	discovery := internalnostr.NewDiscovery(client, st)
	filterBuilder := NewFilterBuilder(&cfg.Sync)
	graph := NewGraph(st, &cfg.Sync.Scope)
	cursors := NewCursorManager(st)

	return &Engine{
		config:        cfg,
		storage:       st,
		nostrClient:   client,
		discovery:     discovery,
		filterBuilder: filterBuilder,
		graph:         graph,
		cursors:       cursors,
		ctx:           engineCtx,
		cancel:        cancel,
		eventChan:     make(chan *nostr.Event, 1000),
	}
}

// Start begins the sync process
func (e *Engine) Start() error {
	// Bootstrap from seed relays
	if err := e.bootstrap(); err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Start event ingestion worker
	e.wg.Add(1)
	go e.ingestEvents()

	// Start continuous sync
	e.wg.Add(1)
	go e.continuousSync()

	// Start periodic refresh of replaceables
	e.wg.Add(1)
	go e.periodicRefresh()

	return nil
}

// Stop gracefully stops the sync engine
func (e *Engine) Stop() {
	e.cancel()
	close(e.eventChan)
	e.wg.Wait()
}

// bootstrap performs initial discovery and graph building
func (e *Engine) bootstrap() error {
	ownerPubkey := e.config.Identity.Npub

	// Step 1: Fetch owner's profile, contacts, and relay hints from seeds
	if err := e.discovery.BootstrapFromSeeds(e.ctx, ownerPubkey); err != nil {
		return fmt.Errorf("failed to bootstrap from seeds: %w", err)
	}

	// Step 2: Fetch owner's contact list (kind 3) to build initial graph
	seedRelays := e.nostrClient.GetSeedRelays()
	filter := nostr.Filter{
		Kinds:   []int{3},
		Authors: []string{ownerPubkey},
		Limit:   1,
	}

	events, err := e.nostrClient.FetchEvents(e.ctx, seedRelays, filter)
	if err != nil {
		return fmt.Errorf("failed to fetch contact list: %w", err)
	}

	if len(events) > 0 {
		// Process the contact list to build the graph
		if err := e.graph.ProcessContactList(e.ctx, events[0], ownerPubkey); err != nil {
			return fmt.Errorf("failed to process contact list: %w", err)
		}
	}

	// Step 3: Get authors in scope
	authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
	if err != nil {
		return fmt.Errorf("failed to get authors in scope: %w", err)
	}

	// Step 4: Discover relay hints for all authors in scope
	ownerRelays, err := e.discovery.GetRelaysForPubkey(e.ctx, ownerPubkey)
	if err != nil || len(ownerRelays) == 0 {
		ownerRelays = seedRelays // Fallback to seeds
	}

	if err := e.discovery.DiscoverRelayHintsForPubkeys(e.ctx, authors, ownerRelays); err != nil {
		return fmt.Errorf("failed to discover relay hints: %w", err)
	}

	return nil
}

// continuousSync runs the main sync loop
func (e *Engine) continuousSync() {
	defer e.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.syncOnce(); err != nil {
				// Log error but continue
				fmt.Printf("Sync error: %v\n", err)
			}
		}
	}
}

// syncOnce performs a single sync iteration
func (e *Engine) syncOnce() error {
	ownerPubkey := e.config.Identity.Npub

	// Get authors in scope
	authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
	if err != nil {
		return fmt.Errorf("failed to get authors: %w", err)
	}

	// Get relays to sync from
	relays := e.getActiveRelays(authors)
	if len(relays) == 0 {
		return fmt.Errorf("no active relays")
	}

	// Build filters with cursors
	kinds := e.filterBuilder.GetConfiguredKinds()
	for _, relay := range relays {
		// Get since cursor for this relay
		since, err := e.cursors.GetSinceCursorForRelay(e.ctx, relay, kinds)
		if err != nil {
			continue
		}

		// Build filters
		filters := e.filterBuilder.BuildFilters(authors, since)

		// Add mention filter if configured
		if e.config.Sync.Scope.IncludeDirectMentions {
			mentionFilter := e.filterBuilder.BuildMentionFilter(ownerPubkey, since)
			filters = append(filters, mentionFilter)
		}

		// Subscribe and collect events
		go e.subscribeRelay(relay, filters)
	}

	return nil
}

// subscribeRelay subscribes to a relay with the given filters
func (e *Engine) subscribeRelay(relay string, filters []nostr.Filter) {
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	eventChan := e.nostrClient.SubscribeEvents(ctx, []string{relay}, filters)

	for event := range eventChan {
		select {
		case e.eventChan <- event:
		case <-e.ctx.Done():
			return
		}
	}
}

// ingestEvents processes events from the event channel
func (e *Engine) ingestEvents() {
	defer e.wg.Done()

	for event := range e.eventChan {
		if err := e.processEvent(event); err != nil {
			// Log error but continue
			fmt.Printf("Event processing error: %v\n", err)
		}
	}
}

// processEvent handles a single event
func (e *Engine) processEvent(event *nostr.Event) error {
	// Store event in Khatru
	if err := e.storage.StoreEvent(e.ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Handle special event kinds
	switch event.Kind {
	case 3:
		// Contact list - update graph
		if err := e.graph.ProcessContactList(e.ctx, event, e.config.Identity.Npub); err != nil {
			return fmt.Errorf("failed to process contact list: %w", err)
		}

		// Recompute mutuals
		if err := e.graph.ComputeMutuals(e.ctx, e.config.Identity.Npub); err != nil {
			return fmt.Errorf("failed to compute mutuals: %w", err)
		}

	case 10002:
		// Relay hints - update relay hints
		hints, err := internalnostr.ParseRelayHints(event)
		if err != nil {
			return fmt.Errorf("failed to parse relay hints: %w", err)
		}

		for _, hint := range hints {
			if err := e.storage.SaveRelayHint(e.ctx, hint); err != nil {
				return fmt.Errorf("failed to save relay hint: %w", err)
			}
		}

	case 7:
		// Reaction - update aggregates
		if err := e.updateReactionAggregate(event); err != nil {
			return fmt.Errorf("failed to update reaction: %w", err)
		}

	case 1:
		// Note - check if it's a reply
		if err := e.updateReplyAggregate(event); err != nil {
			return fmt.Errorf("failed to update reply: %w", err)
		}

	case 9735:
		// Zap - update aggregates
		if err := e.updateZapAggregate(event); err != nil {
			return fmt.Errorf("failed to update zap: %w", err)
		}
	}

	return nil
}

// periodicRefresh refreshes replaceable events periodically
func (e *Engine) periodicRefresh() {
	defer e.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.refreshReplaceables(); err != nil {
				fmt.Printf("Refresh error: %v\n", err)
			}
		}
	}
}

// refreshReplaceables refreshes replaceable events (kinds 0, 3, 10002)
func (e *Engine) refreshReplaceables() error {
	ownerPubkey := e.config.Identity.Npub

	// Get authors in scope
	authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
	if err != nil {
		return err
	}

	// Get active relays
	relays := e.getActiveRelays(authors)
	if len(relays) == 0 {
		return fmt.Errorf("no active relays")
	}

	// Build replaceable filter (no since cursor)
	filter := e.filterBuilder.BuildReplaceableFilter(authors)

	// Fetch events
	events, err := e.nostrClient.FetchEvents(e.ctx, relays, filter)
	if err != nil {
		return err
	}

	// Process events
	for _, event := range events {
		if err := e.processEvent(event); err != nil {
			fmt.Printf("Error processing replaceable event: %v\n", err)
		}
	}

	return nil
}

// getActiveRelays returns the list of active relays to sync from
func (e *Engine) getActiveRelays(authors []string) []string {
	relaySet := make(map[string]bool)

	for _, author := range authors {
		relays, err := e.discovery.GetRelaysForPubkey(e.ctx, author)
		if err != nil {
			continue
		}

		for _, relay := range relays {
			relaySet[relay] = true
		}
	}

	// Convert set to slice
	relays := make([]string, 0, len(relaySet))
	for relay := range relaySet {
		relays = append(relays, relay)
	}

	return relays
}

// Helper methods for aggregate updates
func (e *Engine) updateReactionAggregate(event *nostr.Event) error {
	// Find the event being reacted to
	var targetEventID string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			targetEventID = tag[1]
			break
		}
	}

	if targetEventID == "" {
		return nil // No target event
	}

	// Reaction content is the emoji
	reaction := event.Content
	if reaction == "" {
		reaction = "+" // Default like
	}

	return e.storage.IncrementReaction(e.ctx, targetEventID, reaction, int64(event.CreatedAt))
}

func (e *Engine) updateReplyAggregate(event *nostr.Event) error {
	// Check if this is a reply (has e tags)
	var targetEventID string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			targetEventID = tag[1]
			break
		}
	}

	if targetEventID == "" {
		return nil // Not a reply
	}

	return e.storage.IncrementReplyCount(e.ctx, targetEventID, int64(event.CreatedAt))
}

func (e *Engine) updateZapAggregate(event *nostr.Event) error {
	// Parse zap amount from bolt11 invoice
	// This is simplified - real implementation needs to parse the invoice
	var targetEventID string
	var amount int64 = 1000 // Placeholder

	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			targetEventID = tag[1]
			break
		}
	}

	if targetEventID == "" {
		return nil
	}

	return e.storage.AddZapAmount(e.ctx, targetEventID, amount, int64(event.CreatedAt))
}
