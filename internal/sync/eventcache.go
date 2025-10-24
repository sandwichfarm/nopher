package sync

import (
	"sync"
)

// EventCache is a simple LRU cache for tracking recent event IDs
// Used for fast deduplication without hitting the database
type EventCache struct {
	cache    map[string]struct{}
	keys     []string
	maxSize  int
	position int
	mu       sync.RWMutex
}

// NewEventCache creates a new event cache with the given max size
func NewEventCache(maxSize int) *EventCache {
	return &EventCache{
		cache:   make(map[string]struct{}, maxSize),
		keys:    make([]string, maxSize),
		maxSize: maxSize,
	}
}

// Add adds an event ID to the cache
func (c *EventCache) Add(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already exists, nothing to do
	if _, exists := c.cache[eventID]; exists {
		return
	}

	// If we've reached max size, evict the oldest entry
	if len(c.cache) >= c.maxSize {
		oldKey := c.keys[c.position]
		if oldKey != "" {
			delete(c.cache, oldKey)
		}
	}

	// Add new entry
	c.cache[eventID] = struct{}{}
	c.keys[c.position] = eventID

	// Move to next position (circular buffer)
	c.position = (c.position + 1) % c.maxSize
}

// Contains checks if an event ID is in the cache
func (c *EventCache) Contains(eventID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.cache[eventID]
	return exists
}

// Size returns the current number of entries in the cache
func (c *EventCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear removes all entries from the cache
func (c *EventCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]struct{}, c.maxSize)
	c.keys = make([]string, c.maxSize)
	c.position = 0
}
