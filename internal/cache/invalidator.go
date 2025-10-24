package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// Invalidator handles cache invalidation
type Invalidator struct {
	cache Cache
}

// NewInvalidator creates a new cache invalidator
func NewInvalidator(cache Cache) *Invalidator {
	return &Invalidator{
		cache: cache,
	}
}

// InvalidateEvent invalidates cache entries related to an event
func (inv *Invalidator) InvalidateEvent(ctx context.Context, event *nostr.Event) error {
	// Get invalidation patterns for this event
	patterns := InvalidationPatterns(event.ID, event.Kind, event.PubKey)

	// Invalidate each pattern
	for _, pattern := range patterns {
		if err := inv.InvalidatePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// InvalidatePattern invalidates all keys matching a pattern
func (inv *Invalidator) InvalidatePattern(ctx context.Context, pattern string) error {
	// For patterns with wildcards, we need to handle differently
	// based on the cache implementation

	if !strings.Contains(pattern, "*") {
		// Simple key, just delete it
		return inv.cache.Delete(ctx, pattern)
	}

	// For wildcard patterns, we need pattern-based deletion
	// This is only efficiently supported by some cache implementations

	switch c := inv.cache.(type) {
	case *MemoryCache:
		return inv.invalidateMemoryPattern(ctx, c, pattern)
	case *RedisCache:
		return inv.invalidateRedisPattern(ctx, c, pattern)
	default:
		// For other implementations, can't efficiently handle patterns
		// Just log and continue
		return nil
	}
}

// invalidateMemoryPattern invalidates memory cache keys matching pattern
func (inv *Invalidator) invalidateMemoryPattern(ctx context.Context, mc *MemoryCache, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Convert glob pattern to regex-like matching
	prefix := strings.TrimSuffix(pattern, "*")

	for key := range mc.entries {
		if strings.HasPrefix(key, prefix) {
			delete(mc.entries, key)
		}
	}

	return nil
}

// invalidateRedisPattern invalidates Redis keys matching pattern
func (inv *Invalidator) invalidateRedisPattern(ctx context.Context, rc *RedisCache, pattern string) error {
	// Use Redis SCAN to find matching keys
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()

	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		// Delete in batches of 100
		if len(keys) >= 100 {
			if err := rc.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
			keys = keys[:0]
		}
	}

	// Delete remaining keys
	if len(keys) > 0 {
		if err := rc.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	return nil
}

// InvalidateGopher invalidates all Gopher cache entries
func (inv *Invalidator) InvalidateGopher(ctx context.Context) error {
	return inv.InvalidatePattern(ctx, GopherPattern())
}

// InvalidateGemini invalidates all Gemini cache entries
func (inv *Invalidator) InvalidateGemini(ctx context.Context) error {
	return inv.InvalidatePattern(ctx, GeminiPattern())
}

// InvalidateFinger invalidates all Finger cache entries
func (inv *Invalidator) InvalidateFinger(ctx context.Context) error {
	return inv.InvalidatePattern(ctx, FingerPattern())
}

// InvalidateAll invalidates all cache entries
func (inv *Invalidator) InvalidateAll(ctx context.Context) error {
	return inv.cache.Clear(ctx)
}

// InvalidateProfile invalidates profile cache for a pubkey
func (inv *Invalidator) InvalidateProfile(ctx context.Context, pubkey string) error {
	patterns := []string{
		Kind0Key(pubkey),
		ProfilePattern(pubkey),
	}

	for _, pattern := range patterns {
		if err := inv.InvalidatePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateSection invalidates a specific section
func (inv *Invalidator) InvalidateSection(ctx context.Context, sectionName string) error {
	// Invalidate all pages of this section across all protocols
	patterns := []string{
		SectionKey(sectionName, "gopher", 0),
		SectionKey(sectionName, "gemini", 0),
	}

	for _, pattern := range patterns {
		// This will invalidate page 0, which should trigger re-render
		// For production, might want to invalidate all pages
		if err := inv.InvalidatePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// OnEventIngested is called when a new event is ingested
// This automatically invalidates relevant cache entries
func (inv *Invalidator) OnEventIngested(ctx context.Context, event *nostr.Event) error {
	return inv.InvalidateEvent(ctx, event)
}

// Warmer handles cache warming (pre-populating cache)
type Warmer struct {
	cache Cache
}

// NewWarmer creates a new cache warmer
func NewWarmer(cache Cache) *Warmer {
	return &Warmer{
		cache: cache,
	}
}

// WarmGopherHome pre-populates the Gopher home page
func (w *Warmer) WarmGopherHome(ctx context.Context, content []byte, ttl time.Duration) error {
	key := GopherKey("/")
	return w.cache.Set(ctx, key, content, ttl)
}

// WarmGeminiHome pre-populates the Gemini home page
func (w *Warmer) WarmGeminiHome(ctx context.Context, content []byte, ttl time.Duration) error {
	key := GeminiKey("/", "")
	return w.cache.Set(ctx, key, content, ttl)
}

// WarmProfile pre-populates a profile
func (w *Warmer) WarmProfile(ctx context.Context, pubkey string, protocol string, content []byte, ttl time.Duration) error {
	key := ProfileKey(pubkey, protocol)
	return w.cache.Set(ctx, key, content, ttl)
}
