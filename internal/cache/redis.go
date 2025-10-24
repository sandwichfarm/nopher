package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client *redis.Client
	config *Config
	stats  Stats
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config *Config) (*RedisCache, error) {
	if config.RedisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		config: config,
	}, nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	start := time.Now()
	defer func() {
		r.recordGetTime(time.Since(start))
	}()

	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		r.recordMiss()
		return nil, false, nil
	}
	if err != nil {
		r.recordMiss()
		return nil, false, fmt.Errorf("redis get failed: %w", err)
	}

	r.recordHit()
	return val, true, nil
}

// Set stores a value in cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		r.recordSetTime(time.Since(start))
	}()

	if ttl == 0 {
		ttl = r.config.DefaultTTL
	}

	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Clear removes all values from cache
func (r *RedisCache) Clear(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flush failed: %w", err)
	}
	return nil
}

// Has checks if a key exists in cache
func (r *RedisCache) Has(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return count > 0, nil
}

// Stats returns cache statistics
func (r *RedisCache) Stats(ctx context.Context) (*Stats, error) {
	// Get Redis INFO stats
	info, err := r.client.Info(ctx, "stats", "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("redis info failed: %w", err)
	}

	// Parse Redis info for basic stats
	// This is a simplified version - full parsing would be more complex
	stats := r.stats

	// Get key count
	dbSize, err := r.client.DBSize(ctx).Result()
	if err == nil {
		stats.Keys = dbSize
	}

	// Calculate hit rate from local tracking
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	// Note: Size calculation would require scanning all keys,
	// which is expensive. Left as 0 for Redis implementation.
	_ = info // Info string could be parsed for more stats

	return &stats, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// recordHit increments hit counter
func (r *RedisCache) recordHit() {
	r.stats.Hits++
}

// recordMiss increments miss counter
func (r *RedisCache) recordMiss() {
	r.stats.Misses++
}

// recordGetTime records get operation time
func (r *RedisCache) recordGetTime(duration time.Duration) {
	ms := float64(duration.Microseconds()) / 1000.0
	total := r.stats.Hits + r.stats.Misses

	if total == 0 {
		r.stats.AvgGetTimeMs = ms
	} else {
		r.stats.AvgGetTimeMs = (r.stats.AvgGetTimeMs*float64(total-1) + ms) / float64(total)
	}
}

// recordSetTime records set operation time
func (r *RedisCache) recordSetTime(duration time.Duration) {
	ms := float64(duration.Microseconds()) / 1000.0

	if r.stats.AvgSetTimeMs == 0 {
		r.stats.AvgSetTimeMs = ms
	} else {
		r.stats.AvgSetTimeMs = (r.stats.AvgSetTimeMs + ms) / 2.0
	}
}
