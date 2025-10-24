package security

import (
	"context"
	"fmt"
	"sync"

	"github.com/nbd-wtf/go-nostr"
)

// DenyList manages blocked pubkeys and content filtering
type DenyList struct {
	pubkeys map[string]bool
	mu      sync.RWMutex
}

// NewDenyList creates a new deny list
func NewDenyList(pubkeys []string) *DenyList {
	dl := &DenyList{
		pubkeys: make(map[string]bool),
	}

	for _, pubkey := range pubkeys {
		dl.pubkeys[pubkey] = true
	}

	return dl
}

// IsPubkeyDenied checks if a pubkey is on the deny list
func (dl *DenyList) IsPubkeyDenied(pubkey string) bool {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return dl.pubkeys[pubkey]
}

// IsEventDenied checks if an event should be denied
func (dl *DenyList) IsEventDenied(event *nostr.Event) bool {
	return dl.IsPubkeyDenied(event.PubKey)
}

// AddPubkey adds a pubkey to the deny list
func (dl *DenyList) AddPubkey(pubkey string) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.pubkeys[pubkey] = true
}

// RemovePubkey removes a pubkey from the deny list
func (dl *DenyList) RemovePubkey(pubkey string) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	delete(dl.pubkeys, pubkey)
}

// ListDeniedPubkeys returns all denied pubkeys
func (dl *DenyList) ListDeniedPubkeys() []string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	pubkeys := make([]string, 0, len(dl.pubkeys))
	for pubkey := range dl.pubkeys {
		pubkeys = append(pubkeys, pubkey)
	}

	return pubkeys
}

// Count returns the number of denied pubkeys
func (dl *DenyList) Count() int {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	return len(dl.pubkeys)
}

// Clear removes all pubkeys from the deny list
func (dl *DenyList) Clear() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.pubkeys = make(map[string]bool)
}

// FilterEvents filters out denied events from a slice
func (dl *DenyList) FilterEvents(events []*nostr.Event) []*nostr.Event {
	if len(dl.pubkeys) == 0 {
		return events
	}

	filtered := make([]*nostr.Event, 0, len(events))
	for _, event := range events {
		if !dl.IsEventDenied(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// ContentFilter handles content-based filtering
type ContentFilter struct {
	bannedWords []string
	mu          sync.RWMutex
}

// NewContentFilter creates a new content filter
func NewContentFilter(bannedWords []string) *ContentFilter {
	return &ContentFilter{
		bannedWords: bannedWords,
	}
}

// ContainsBannedContent checks if content contains banned words
func (cf *ContentFilter) ContainsBannedContent(content string) bool {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	// Simple substring matching
	// Production version would use regex and case-insensitive matching
	for _, word := range cf.bannedWords {
		if contains(content, word) {
			return true
		}
	}

	return false
}

// IsEventFiltered checks if an event should be filtered
func (cf *ContentFilter) IsEventFiltered(event *nostr.Event) bool {
	return cf.ContainsBannedContent(event.Content)
}

// AddBannedWord adds a word to the banned list
func (cf *ContentFilter) AddBannedWord(word string) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	cf.bannedWords = append(cf.bannedWords, word)
}

// CombinedFilter combines deny list and content filter
type CombinedFilter struct {
	denyList      *DenyList
	contentFilter *ContentFilter
}

// NewCombinedFilter creates a combined filter
func NewCombinedFilter(denyList *DenyList, contentFilter *ContentFilter) *CombinedFilter {
	return &CombinedFilter{
		denyList:      denyList,
		contentFilter: contentFilter,
	}
}

// IsEventAllowed checks if an event passes all filters
func (cf *CombinedFilter) IsEventAllowed(event *nostr.Event) bool {
	if cf.denyList != nil && cf.denyList.IsEventDenied(event) {
		return false
	}

	if cf.contentFilter != nil && cf.contentFilter.IsEventFiltered(event) {
		return false
	}

	return true
}

// FilterEvents filters a slice of events
func (cf *CombinedFilter) FilterEvents(events []*nostr.Event) []*nostr.Event {
	filtered := make([]*nostr.Event, 0, len(events))

	for _, event := range events {
		if cf.IsEventAllowed(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// SecurityPolicy defines security rules
type SecurityPolicy struct {
	DenyListPubkeys []string
	BannedWords     []string
	AllowAnonymous  bool
	RequireNIP05    bool
}

// Enforcer enforces security policies
type Enforcer struct {
	policy        *SecurityPolicy
	denyList      *DenyList
	contentFilter *ContentFilter
	filter        *CombinedFilter
}

// NewEnforcer creates a new security enforcer
func NewEnforcer(policy *SecurityPolicy) *Enforcer {
	denyList := NewDenyList(policy.DenyListPubkeys)
	contentFilter := NewContentFilter(policy.BannedWords)
	filter := NewCombinedFilter(denyList, contentFilter)

	return &Enforcer{
		policy:        policy,
		denyList:      denyList,
		contentFilter: contentFilter,
		filter:        filter,
	}
}

// EnforceEvent checks if an event is allowed
func (e *Enforcer) EnforceEvent(ctx context.Context, event *nostr.Event) error {
	if !e.filter.IsEventAllowed(event) {
		return fmt.Errorf("event denied by security policy")
	}

	return nil
}

// EnforceEvents filters a list of events
func (e *Enforcer) EnforceEvents(ctx context.Context, events []*nostr.Event) []*nostr.Event {
	return e.filter.FilterEvents(events)
}

// GetDenyList returns the deny list
func (e *Enforcer) GetDenyList() *DenyList {
	return e.denyList
}

// GetContentFilter returns the content filter
func (e *Enforcer) GetContentFilter() *ContentFilter {
	return e.contentFilter
}

// Helper function for simple string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
