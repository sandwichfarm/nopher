package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	entries         map[string]*Entry
	mu              sync.RWMutex
	config          *Config
	stats           Stats
	statsMu         sync.RWMutex
	stopCleanup     chan struct{}
	cleanupDone     chan struct{}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config *Config) *MemoryCache {
	mc := &MemoryCache{
		entries:     make(map[string]*Entry),
		config:      config,
		stopCleanup: make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.cleanupLoop()

	return mc
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	start := time.Now()
	defer func() {
		m.recordGetTime(time.Since(start))
	}()

	m.mu.RLock()
	entry, exists := m.entries[key]
	m.mu.RUnlock()

	if !exists {
		m.recordMiss()
		return nil, false, nil
	}

	// Check if expired
	if entry.IsExpired() {
		m.mu.Lock()
		delete(m.entries, key)
		m.mu.Unlock()
		m.recordMiss()
		return nil, false, nil
	}

	// Update access time and hit count
	m.mu.Lock()
	entry.AccessedAt = time.Now()
	entry.HitCount++
	m.mu.Unlock()

	m.recordHit()
	return entry.Value, true, nil
}

// Set stores a value in cache with TTL
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		m.recordSetTime(time.Since(start))
	}()

	if ttl == 0 {
		ttl = m.config.DefaultTTL
	}

	now := time.Now()
	entry := &Entry{
		Key:        key,
		Value:      value,
		Size:       int64(len(value)),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		AccessedAt: now,
		HitCount:   0,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we need to evict entries
	newSize := m.calculateSizeWithoutLock() + entry.Size
	if m.config.MaxSize > 0 && newSize > m.config.MaxSize {
		m.evictLRUWithoutLock(entry.Size)
	}

	m.entries[key] = entry
	return nil
}

// Delete removes a value from cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, key)
	return nil
}

// Clear removes all values from cache
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make(map[string]*Entry)
	return nil
}

// Has checks if a key exists in cache
func (m *MemoryCache) Has(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	entry, exists := m.entries[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if entry.IsExpired() {
		m.mu.Lock()
		delete(m.entries, key)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Stats returns cache statistics
func (m *MemoryCache) Stats(ctx context.Context) (*Stats, error) {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	m.mu.RLock()
	keys := int64(len(m.entries))
	sizeBytes := m.calculateSizeWithoutLock()
	m.mu.RUnlock()

	stats := m.stats
	stats.Keys = keys
	stats.SizeBytes = sizeBytes

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return &stats, nil
}

// Close closes the cache and stops cleanup
func (m *MemoryCache) Close() error {
	close(m.stopCleanup)
	<-m.cleanupDone
	return nil
}

// cleanupLoop periodically removes expired entries
func (m *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()
	defer close(m.cleanupDone)

	for {
		select {
		case <-m.stopCleanup:
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup removes expired entries
func (m *MemoryCache) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, entry := range m.entries {
		if now.After(entry.ExpiresAt) {
			delete(m.entries, key)
		}
	}
}

// evictLRUWithoutLock evicts least recently used entries to make room
// Must be called with m.mu locked
func (m *MemoryCache) evictLRUWithoutLock(needed int64) {
	// Find least recently used entries
	type entryWithKey struct {
		key   string
		entry *Entry
	}

	var entries []entryWithKey
	for key, entry := range m.entries {
		entries = append(entries, entryWithKey{key, entry})
	}

	// Sort by access time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.AccessedAt.After(entries[j].entry.AccessedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict until we have enough space
	freed := int64(0)
	evicted := 0
	for _, e := range entries {
		if freed >= needed {
			break
		}
		freed += e.entry.Size
		delete(m.entries, e.key)
		evicted++
	}

	m.statsMu.Lock()
	m.stats.Evictions += int64(evicted)
	m.statsMu.Unlock()
}

// calculateSizeWithoutLock calculates total cache size
// Must be called with m.mu locked (read or write)
func (m *MemoryCache) calculateSizeWithoutLock() int64 {
	var size int64
	for _, entry := range m.entries {
		size += entry.Size
	}
	return size
}

// recordHit increments hit counter
func (m *MemoryCache) recordHit() {
	m.statsMu.Lock()
	m.stats.Hits++
	m.statsMu.Unlock()
}

// recordMiss increments miss counter
func (m *MemoryCache) recordMiss() {
	m.statsMu.Lock()
	m.stats.Misses++
	m.statsMu.Unlock()
}

// recordGetTime records get operation time
func (m *MemoryCache) recordGetTime(duration time.Duration) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	ms := float64(duration.Microseconds()) / 1000.0
	total := m.stats.Hits + m.stats.Misses

	if total == 0 {
		m.stats.AvgGetTimeMs = ms
	} else {
		// Running average
		m.stats.AvgGetTimeMs = (m.stats.AvgGetTimeMs*float64(total-1) + ms) / float64(total)
	}
}

// recordSetTime records set operation time
func (m *MemoryCache) recordSetTime(duration time.Duration) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	ms := float64(duration.Microseconds()) / 1000.0

	if m.stats.AvgSetTimeMs == 0 {
		m.stats.AvgSetTimeMs = ms
	} else {
		// Simple running average (approximate)
		m.stats.AvgSetTimeMs = (m.stats.AvgSetTimeMs + ms) / 2.0
	}
}
