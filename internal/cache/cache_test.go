package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	config := DefaultConfig()
	config.CleanupInterval = 100 * time.Millisecond

	cache := NewMemoryCache(config)
	defer cache.Close()

	testCacheOperations(t, cache)
}

func TestNullCache(t *testing.T) {
	cache := NewNullCache()
	defer cache.Close()

	ctx := context.Background()

	// All operations should succeed but do nothing
	err := cache.Set(ctx, "key", []byte("value"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, hit, err := cache.Get(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hit {
		t.Error("expected cache miss for null cache")
	}
	if val != nil {
		t.Error("expected nil value for null cache")
	}
}

func testCacheOperations(t *testing.T, cache Cache) {
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := cache.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Fatalf("failed to set: %v", err)
		}

		got, hit, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if !hit {
			t.Error("expected cache hit")
		}
		if string(got) != string(value) {
			t.Errorf("expected %s, got %s", value, got)
		}
	})

	t.Run("Get Miss", func(t *testing.T) {
		_, hit, err := cache.Get(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hit {
			t.Error("expected cache miss")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete-test"
		value := []byte("value")

		cache.Set(ctx, key, value, time.Minute)
		err := cache.Delete(ctx, key)
		if err != nil {
			t.Fatalf("failed to delete: %v", err)
		}

		_, hit, _ := cache.Get(ctx, key)
		if hit {
			t.Error("expected cache miss after delete")
		}
	})

	t.Run("Has", func(t *testing.T) {
		key := "has-test"
		value := []byte("value")

		cache.Set(ctx, key, value, time.Minute)

		has, err := cache.Has(ctx, key)
		if err != nil {
			t.Fatalf("failed to check has: %v", err)
		}
		if !has {
			t.Error("expected key to exist")
		}

		cache.Delete(ctx, key)

		has, err = cache.Has(ctx, key)
		if err != nil {
			t.Fatalf("failed to check has: %v", err)
		}
		if has {
			t.Error("expected key to not exist")
		}
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		key := "ttl-test"
		value := []byte("value")

		// Set with very short TTL
		cache.Set(ctx, key, value, 50*time.Millisecond)

		// Should exist immediately
		_, hit, _ := cache.Get(ctx, key)
		if !hit {
			t.Error("expected cache hit immediately after set")
		}

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Should be expired
		_, hit, _ = cache.Get(ctx, key)
		if hit {
			t.Error("expected cache miss after TTL expiration")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		// Add multiple keys
		for i := 0; i < 10; i++ {
			key := string(rune('a' + i))
			cache.Set(ctx, key, []byte("value"), time.Minute)
		}

		err := cache.Clear(ctx)
		if err != nil {
			t.Fatalf("failed to clear: %v", err)
		}

		// All keys should be gone
		for i := 0; i < 10; i++ {
			key := string(rune('a' + i))
			_, hit, _ := cache.Get(ctx, key)
			if hit {
				t.Errorf("expected key %s to be cleared", key)
			}
		}
	})

	t.Run("Stats", func(t *testing.T) {
		// Reset cache
		cache.Clear(ctx)

		// Generate some hits and misses
		cache.Set(ctx, "stats-key", []byte("value"), time.Minute)

		cache.Get(ctx, "stats-key")     // hit
		cache.Get(ctx, "stats-key")     // hit
		cache.Get(ctx, "nonexistent-1") // miss
		cache.Get(ctx, "nonexistent-2") // miss

		stats, err := cache.Stats(ctx)
		if err != nil {
			t.Fatalf("failed to get stats: %v", err)
		}

		if stats.Hits < 2 {
			t.Errorf("expected at least 2 hits, got %d", stats.Hits)
		}

		if stats.Misses < 2 {
			t.Errorf("expected at least 2 misses, got %d", stats.Misses)
		}

		if stats.HitRate == 0 {
			t.Error("expected non-zero hit rate")
		}
	})
}

func TestMemoryCacheEviction(t *testing.T) {
	config := DefaultConfig()
	config.MaxSize = 100 // Very small cache
	config.CleanupInterval = 1 * time.Second

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Fill cache beyond capacity
	for i := 0; i < 20; i++ {
		key := string(rune('a' + i))
		value := make([]byte, 10)
		cache.Set(ctx, key, value, time.Minute)
	}

	stats, _ := cache.Stats(ctx)

	// Should have evicted some entries
	if stats.Evictions == 0 {
		t.Error("expected some evictions")
	}

	// Size should be under max
	if stats.SizeBytes > config.MaxSize {
		t.Errorf("cache size %d exceeds max %d", stats.SizeBytes, config.MaxSize)
	}
}

func TestMemoryCacheCleanup(t *testing.T) {
	config := DefaultConfig()
	config.CleanupInterval = 50 * time.Millisecond

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()

	// Add expired entries
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		cache.Set(ctx, key, []byte("value"), 10*time.Millisecond)
	}

	// Wait for expiration + cleanup
	time.Sleep(100 * time.Millisecond)

	// All should be cleaned up
	stats, _ := cache.Stats(ctx)
	if stats.Keys > 0 {
		t.Errorf("expected 0 keys after cleanup, got %d", stats.Keys)
	}
}
