package cache

import (
	"context"
	"time"
)

// Cache defines the interface for all cache implementations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, bool, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from cache
	Clear(ctx context.Context) error

	// Has checks if a key exists in cache
	Has(ctx context.Context, key string) (bool, error)

	// Stats returns cache statistics
	Stats(ctx context.Context) (*Stats, error)

	// Close closes the cache connection
	Close() error
}

// Stats contains cache statistics
type Stats struct {
	Hits          int64
	Misses        int64
	Keys          int64
	SizeBytes     int64
	Evictions     int64
	HitRate       float64
	AvgGetTimeMs  float64
	AvgSetTimeMs  float64
}

// Entry represents a cache entry with metadata
type Entry struct {
	Key        string
	Value      []byte
	Size       int64
	CreatedAt  time.Time
	ExpiresAt  time.Time
	AccessedAt time.Time
	HitCount   int64
}

// IsExpired checks if an entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// TTL returns the time until expiration
func (e *Entry) TTL() time.Duration {
	return time.Until(e.ExpiresAt)
}

// Config contains cache configuration
type Config struct {
	Enabled      bool
	Engine       string // "memory" or "redis"
	RedisURL     string
	MaxSize      int64  // Maximum cache size in bytes
	DefaultTTL   time.Duration
	CleanupInterval time.Duration
}

// Option is a functional option for cache configuration
type Option func(*Config)

// WithEngine sets the cache engine
func WithEngine(engine string) Option {
	return func(c *Config) {
		c.Engine = engine
	}
}

// WithRedisURL sets the Redis URL
func WithRedisURL(url string) Option {
	return func(c *Config) {
		c.RedisURL = url
	}
}

// WithMaxSize sets the maximum cache size
func WithMaxSize(size int64) Option {
	return func(c *Config) {
		c.MaxSize = size
	}
}

// WithDefaultTTL sets the default TTL
func WithDefaultTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.DefaultTTL = ttl
	}
}

// WithCleanupInterval sets the cleanup interval
func WithCleanupInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.CleanupInterval = interval
	}
}

// DefaultConfig returns a default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		Engine:          "memory",
		MaxSize:         100 * 1024 * 1024, // 100MB
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

// New creates a new cache instance based on the engine type
func New(config *Config) (Cache, error) {
	if !config.Enabled {
		return NewNullCache(), nil
	}

	switch config.Engine {
	case "memory":
		return NewMemoryCache(config), nil
	case "redis":
		return NewRedisCache(config)
	default:
		// Default to memory cache
		return NewMemoryCache(config), nil
	}
}

// NullCache is a no-op cache implementation
type NullCache struct{}

// NewNullCache creates a new null cache
func NewNullCache() *NullCache {
	return &NullCache{}
}

// Get always returns a miss
func (n *NullCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return nil, false, nil
}

// Set does nothing
func (n *NullCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (n *NullCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Clear does nothing
func (n *NullCache) Clear(ctx context.Context) error {
	return nil
}

// Has always returns false
func (n *NullCache) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Stats returns empty stats
func (n *NullCache) Stats(ctx context.Context) (*Stats, error) {
	return &Stats{}, nil
}

// Close does nothing
func (n *NullCache) Close() error {
	return nil
}
