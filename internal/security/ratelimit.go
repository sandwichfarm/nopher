package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate     int           // Requests per window
	window   time.Duration // Time window
	buckets  map[string]*bucket
	mu       sync.RWMutex
	cleanupInterval time.Duration
	stopCleanup chan struct{}
}

// bucket represents a token bucket for a single client
type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:            rate,
		window:          window,
		buckets:         make(map[string]*bucket),
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request from client is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.RLock()
	b, exists := rl.buckets[clientID]
	rl.mu.RUnlock()

	if !exists {
		// Create new bucket
		b = &bucket{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}

		rl.mu.Lock()
		rl.buckets[clientID] = b
		rl.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	if elapsed >= rl.window {
		// Full refill
		b.tokens = rl.rate
		b.lastRefill = now
	} else {
		// Partial refill
		tokensToAdd := int(float64(rl.rate) * (float64(elapsed) / float64(rl.window)))
		b.tokens += tokensToAdd
		if b.tokens > rl.rate {
			b.tokens = rl.rate
		}
		if tokensToAdd > 0 {
			b.lastRefill = now
		}
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// Reset resets the rate limit for a client
func (rl *RateLimiter) Reset(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, clientID)
}

// GetLimit returns the current limit for a client
func (rl *RateLimiter) GetLimit(clientID string) (remaining int, resetTime time.Time) {
	rl.mu.RLock()
	b, exists := rl.buckets[clientID]
	rl.mu.RUnlock()

	if !exists {
		return rl.rate, time.Now()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.tokens, b.lastRefill.Add(rl.window)
}

// cleanupLoop periodically removes old buckets
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCleanup:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

// cleanup removes buckets that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-2 * rl.window)

	for clientID, b := range rl.buckets {
		b.mu.Lock()
		if b.lastRefill.Before(cutoff) {
			delete(rl.buckets, clientID)
		}
		b.mu.Unlock()
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	close(rl.stopCleanup)
}

// MultiRateLimiter manages multiple rate limiters for different purposes
type MultiRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewMultiRateLimiter creates a new multi rate limiter
func NewMultiRateLimiter() *MultiRateLimiter {
	return &MultiRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

// AddLimiter adds a named rate limiter
func (mrl *MultiRateLimiter) AddLimiter(name string, limiter *RateLimiter) {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	mrl.limiters[name] = limiter
}

// Allow checks if a request is allowed for a specific limiter
func (mrl *MultiRateLimiter) Allow(limiterName, clientID string) bool {
	mrl.mu.RLock()
	limiter, exists := mrl.limiters[limiterName]
	mrl.mu.RUnlock()

	if !exists {
		// No limiter configured, allow by default
		return true
	}

	return limiter.Allow(clientID)
}

// Close closes all rate limiters
func (mrl *MultiRateLimiter) Close() {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	for _, limiter := range mrl.limiters {
		limiter.Close()
	}
}

// RateLimitMiddleware wraps a handler with rate limiting
type RateLimitMiddleware struct {
	limiter      *RateLimiter
	getClientID  func(ctx context.Context) string
	onLimitExceeded func(ctx context.Context, clientID string) error
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(
	limiter *RateLimiter,
	getClientID func(ctx context.Context) string,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:     limiter,
		getClientID: getClientID,
		onLimitExceeded: func(ctx context.Context, clientID string) error {
			return fmt.Errorf("rate limit exceeded for client: %s", clientID)
		},
	}
}

// SetOnLimitExceeded sets the callback for when limit is exceeded
func (rlm *RateLimitMiddleware) SetOnLimitExceeded(fn func(ctx context.Context, clientID string) error) {
	rlm.onLimitExceeded = fn
}

// Check checks if the request is allowed
func (rlm *RateLimitMiddleware) Check(ctx context.Context) error {
	clientID := rlm.getClientID(ctx)

	if !rlm.limiter.Allow(clientID) {
		return rlm.onLimitExceeded(ctx, clientID)
	}

	return nil
}

// PerIPRateLimiter creates a rate limiter that limits by IP address
func PerIPRateLimiter(rate int, window time.Duration) *RateLimitMiddleware {
	limiter := NewRateLimiter(rate, window)

	return NewRateLimitMiddleware(
		limiter,
		func(ctx context.Context) string {
			// Extract IP from context
			// This would be set by the protocol server
			if ip, ok := ctx.Value("client_ip").(string); ok {
				return ip
			}
			return "unknown"
		},
	)
}

// RateLimitConfig contains rate limit configuration
type RateLimitConfig struct {
	Enabled        bool
	RequestsPerMin int
	BurstSize      int
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 60,
		BurstSize:      10,
	}
}

// NewRateLimiterFromConfig creates a rate limiter from config
func NewRateLimiterFromConfig(cfg *RateLimitConfig) *RateLimiter {
	if !cfg.Enabled {
		// Return a no-op rate limiter
		return nil
	}

	return NewRateLimiter(cfg.RequestsPerMin, time.Minute)
}
