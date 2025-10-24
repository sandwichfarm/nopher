package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/sandwich/nophr/internal/config"
	internalnostr "github.com/sandwich/nophr/internal/nostr"
	"github.com/sandwich/nophr/internal/storage"
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

	// Phase 20: Optional retention evaluation callback
	evaluateRetention func(context.Context, *nostr.Event) error
}

// New creates a new sync engine (legacy signature for compatibility)
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

// NewEngine creates a new sync engine with storage and config only
func NewEngine(st *storage.Storage, cfg *config.Config) *Engine {
	ctx := context.Background()
	engineCtx, cancel := context.WithCancel(ctx)

	// Create nostr client
	nostrClient := internalnostr.New(ctx, &cfg.Relays)

	discovery := internalnostr.NewDiscovery(nostrClient, st)
	filterBuilder := NewFilterBuilder(&cfg.Sync)
	graph := NewGraph(st, &cfg.Sync.Scope)
	cursors := NewCursorManager(st)

	return &Engine{
		config:        cfg,
		storage:       st,
		nostrClient:   nostrClient,
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

// SetRetentionEvaluator sets the retention evaluation callback (Phase 20)
func (e *Engine) SetRetentionEvaluator(fn func(context.Context, *nostr.Event) error) {
	e.evaluateRetention = fn
}

// getOwnerPubkey decodes the npub to hex pubkey
func (e *Engine) getOwnerPubkey() (string, error) {
	if _, hex, err := nip19.Decode(e.config.Identity.Npub); err != nil {
		return "", fmt.Errorf("failed to decode npub: %w", err)
	} else {
		return hex.(string), nil
	}
}

// bootstrap performs initial discovery and graph building
func (e *Engine) bootstrap() error {
	fmt.Printf("[SYNC] Starting bootstrap process...\n")
	ownerPubkey, err := e.getOwnerPubkey()
	if err != nil {
		return err
	}
	fmt.Printf("[SYNC] Owner pubkey (hex): %s\n", ownerPubkey)

	// Step 1: Fetch owner's profile, contacts, and relay hints from seeds
	fmt.Printf("[SYNC] Step 1: Bootstrapping from seed relays...\n")
	if err := e.discovery.BootstrapFromSeeds(e.ctx, ownerPubkey); err != nil {
		return fmt.Errorf("failed to bootstrap from seeds: %w", err)
	}
	fmt.Printf("[SYNC] ✓ Bootstrap from seeds complete\n")

	// Step 2: Fetch owner's contact list (kind 3) to build initial graph
	seedRelays := e.nostrClient.GetSeedRelays()
	fmt.Printf("[SYNC] Step 2: Fetching contact list from %d seed relays\n", len(seedRelays))
	for i, relay := range seedRelays {
		fmt.Printf("[SYNC]   Seed relay %d: %s\n", i+1, relay)
	}

	filter := nostr.Filter{
		Kinds:   []int{3},
		Authors: []string{ownerPubkey},
		Limit:   1,
	}

	events, err := e.nostrClient.FetchEvents(e.ctx, seedRelays, filter)
	if err != nil {
		return fmt.Errorf("failed to fetch contact list: %w", err)
	}
	fmt.Printf("[SYNC] Fetched %d contact list events\n", len(events))

	if len(events) > 0 {
		// Process the contact list to build the graph
		fmt.Printf("[SYNC] Processing contact list (event ID: %s)\n", events[0].ID)
		if err := e.graph.ProcessContactList(e.ctx, events[0], ownerPubkey); err != nil {
			return fmt.Errorf("failed to process contact list: %w", err)
		}
		fmt.Printf("[SYNC] ✓ Contact list processed\n")
	} else {
		fmt.Printf("[SYNC] ⚠ No contact list found - will sync owner events only\n")
	}

	// Step 3: Get authors in scope
	fmt.Printf("[SYNC] Step 3: Getting authors in scope...\n")
	authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
	if err != nil {
		return fmt.Errorf("failed to get authors in scope: %w", err)
	}
	fmt.Printf("[SYNC] Authors in scope: %d\n", len(authors))
	if len(authors) <= 5 {
		for i, author := range authors {
			fmt.Printf("[SYNC]   Author %d: %s\n", i+1, author[:16]+"...")
		}
	} else {
		fmt.Printf("[SYNC]   (First 5 authors shown)\n")
		for i := 0; i < 5; i++ {
			fmt.Printf("[SYNC]   Author %d: %s\n", i+1, authors[i][:16]+"...")
		}
	}

	// Step 4: Discover relay hints for all authors in scope
	fmt.Printf("[SYNC] Step 4: Discovering relay hints...\n")
	ownerRelays, err := e.discovery.GetRelaysForPubkey(e.ctx, ownerPubkey)
	if err != nil || len(ownerRelays) == 0 {
		ownerRelays = seedRelays // Fallback to seeds
		fmt.Printf("[SYNC] Using seed relays as fallback (%d relays)\n", len(ownerRelays))
	} else {
		fmt.Printf("[SYNC] Using owner's relays (%d relays)\n", len(ownerRelays))
	}

	if err := e.discovery.DiscoverRelayHintsForPubkeys(e.ctx, authors, ownerRelays); err != nil {
		return fmt.Errorf("failed to discover relay hints: %w", err)
	}
	fmt.Printf("[SYNC] ✓ Relay hints discovered\n")
	fmt.Printf("[SYNC] ✓ Bootstrap complete!\n\n")

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
	fmt.Printf("[SYNC] Starting sync iteration...\n")
	ownerPubkey, err := e.getOwnerPubkey()
	if err != nil {
		return err
	}

	// Get authors in scope
	authors, err := e.graph.GetAuthorsInScope(e.ctx, ownerPubkey)
	if err != nil {
		return fmt.Errorf("failed to get authors: %w", err)
	}
	fmt.Printf("[SYNC] Syncing for %d authors\n", len(authors))

	// Get relays to sync from
	relays := e.getActiveRelays(authors)
	if len(relays) == 0 {
		fmt.Printf("[SYNC] ⚠ No active relays found!\n")
		return fmt.Errorf("no active relays")
	}
	fmt.Printf("[SYNC] Active relays: %d\n", len(relays))

	// Build filters with cursors
	kinds := e.filterBuilder.GetConfiguredKinds()
	fmt.Printf("[SYNC] Configured event kinds: %v\n", kinds)

	for i, relay := range relays {
		fmt.Printf("[SYNC] Processing relay %d/%d: %s\n", i+1, len(relays), relay)

		// Get since cursor for this relay
		since, err := e.cursors.GetSinceCursorForRelay(e.ctx, relay, kinds)
		if err != nil {
			fmt.Printf("[SYNC]   ⚠ Failed to get cursor: %v\n", err)
			continue
		}
		if since > 0 {
			fmt.Printf("[SYNC]   Since cursor: %d (%s)\n", since, time.Unix(int64(since), 0).Format(time.RFC3339))
		} else {
			fmt.Printf("[SYNC]   Since cursor: 0 (fetching all history)\n")
		}

		// Build filters
		filters := e.filterBuilder.BuildFilters(authors, since)
		fmt.Printf("[SYNC]   Built %d filters\n", len(filters))

		// Add mention filter if configured
		if e.config.Sync.Scope.IncludeDirectMentions {
			mentionFilter := e.filterBuilder.BuildMentionFilter(ownerPubkey, since)
			filters = append(filters, mentionFilter)
			fmt.Printf("[SYNC]   Added mention filter (total: %d filters)\n", len(filters))
		}

		// Subscribe and collect events
		fmt.Printf("[SYNC]   Subscribing to relay with %d filters...\n", len(filters))
		go e.subscribeRelay(relay, filters)
	}

	fmt.Printf("[SYNC] ✓ Sync iteration dispatched\n\n")
	return nil
}

// subscribeRelay subscribes to a relay with the given filters
func (e *Engine) subscribeRelay(relay string, filters []nostr.Filter) {
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	fmt.Printf("[SYNC] Subscribing to %s...\n", relay)
	eventChan := e.nostrClient.SubscribeEvents(ctx, []string{relay}, filters)

	eventCount := 0
	for event := range eventChan {
		eventCount++
		if eventCount == 1 {
			fmt.Printf("[SYNC] ✓ Receiving events from %s\n", relay)
		}
		select {
		case e.eventChan <- event:
		case <-e.ctx.Done():
			fmt.Printf("[SYNC] Subscription to %s cancelled (context done)\n", relay)
			return
		}
	}

	if eventCount > 0 {
		fmt.Printf("[SYNC] ✓ Received %d events from %s\n", eventCount, relay)
	} else {
		fmt.Printf("[SYNC] No events received from %s\n", relay)
	}
}

// ingestEvents processes events from the event channel
func (e *Engine) ingestEvents() {
	defer e.wg.Done()

	fmt.Printf("[SYNC] Event ingestion worker started\n")
	eventCount := 0

	for event := range e.eventChan {
		eventCount++
		if eventCount%10 == 1 {
			fmt.Printf("[SYNC] Processing event %d (kind %d, author: %s)\n", eventCount, event.Kind, event.PubKey[:16]+"...")
		}

		if err := e.processEvent(event); err != nil {
			// Log error but continue
			fmt.Printf("[SYNC] ⚠ Event processing error: %v\n", err)
		}
	}

	fmt.Printf("[SYNC] Event ingestion worker stopped (processed %d events)\n", eventCount)
}

// processEvent handles a single event
func (e *Engine) processEvent(event *nostr.Event) error {
	// Store event in Khatru
	if err := e.storage.StoreEvent(e.ctx, event); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}
	fmt.Printf("[SYNC]   ✓ Stored event %s (kind %d)\n", event.ID[:16]+"...", event.Kind)

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

	// Phase 20: Evaluate retention if enabled
	if e.evaluateRetention != nil {
		if err := e.evaluateRetention(e.ctx, event); err != nil {
			// Log error but don't fail the entire event processing
			fmt.Printf("[SYNC]   ⚠ Retention evaluation error: %v\n", err)
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
	ownerPubkey, err := e.getOwnerPubkey()
	if err != nil {
		return err
	}

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

	// Fallback to seed relays if no relays discovered
	if len(relays) == 0 {
		fmt.Printf("[SYNC] No relay hints found, falling back to seed relays\n")
		relays = e.nostrClient.GetSeedRelays()
	} else {
		// Also include seed relays as backup
		fmt.Printf("[SYNC] Adding seed relays as backup to discovered relays\n")
		seedRelays := e.nostrClient.GetSeedRelays()
		for _, seed := range seedRelays {
			if !relaySet[seed] {
				relays = append(relays, seed)
			}
		}
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
