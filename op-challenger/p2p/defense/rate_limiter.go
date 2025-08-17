package defense

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum/go-ethereum/log"
)

// RateLimiter implements basic rate limiting for P2P messages (Phase 1: simple token bucket)
type RateLimiter struct {
	// Configuration
	config *types.RateLimiterConfig // Rate limiter configuration
	logger log.Logger               // Logger

	// Rate limiting data
	buckets map[string]*TokenBucket // Rate limiting buckets per peer/IP
	mu      sync.RWMutex            // Concurrency control

	// Statistics
	totalRequests   int64     // Total requests processed
	blockedRequests int64     // Total requests blocked
	lastCleanup     time.Time // Last cleanup time

	// State
	isRunning bool // Whether the rate limiter is running
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     float64    // Current number of tokens
	capacity   float64    // Maximum capacity (burst size)
	refillRate float64    // Tokens per second
	lastRefill time.Time  // Last refill time
	mu         sync.Mutex // Bucket-specific mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *types.RateLimiterConfig, logger log.Logger) *RateLimiter {
	return &RateLimiter{
		config:      config,
		logger:      logger,
		buckets:     make(map[string]*TokenBucket),
		mu:          sync.RWMutex{},
		lastCleanup: time.Now(),
		isRunning:   false,
	}
}

// Start starts the rate limiter
func (rl *RateLimiter) Start() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.isRunning {
		return fmt.Errorf("rate limiter is already running")
	}

	rl.logger.Info("Starting rate limiter",
		"max_requests_per_second", rl.config.MaxRequestsPerSecond,
		"burst_size", rl.config.BurstSize)

	rl.isRunning = true
	rl.lastCleanup = time.Now()

	// Start cleanup routine
	go rl.cleanupRoutine()

	rl.logger.Info("Rate limiter started successfully")
	return nil
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.isRunning {
		return
	}

	rl.logger.Info("Stopping rate limiter")
	rl.isRunning = false
	rl.logger.Info("Rate limiter stopped")
}

// CheckLimit checks if a request from the given identifier is within rate limits
func (rl *RateLimiter) CheckLimit(identifier string) bool {
	rl.mu.RLock()
	if !rl.isRunning {
		rl.mu.RUnlock()
		return true // Allow all requests if not running
	}
	rl.mu.RUnlock()

	// Get or create token bucket for this identifier
	bucket := rl.getOrCreateBucket(identifier)

	// Check if request is allowed
	allowed := bucket.consume(1.0)

	// Update statistics
	rl.mu.Lock()
	rl.totalRequests++
	if !allowed {
		rl.blockedRequests++
	}
	rl.mu.Unlock()

	if !allowed {
		rl.logger.Debug("Request rate limited", "identifier", identifier)
	}

	return allowed
}

// CheckLimitWithCost checks if a request with specific cost is within rate limits
func (rl *RateLimiter) CheckLimitWithCost(identifier string, cost float64) bool {
	rl.mu.RLock()
	if !rl.isRunning {
		rl.mu.RUnlock()
		return true // Allow all requests if not running
	}
	rl.mu.RUnlock()

	// Get or create token bucket for this identifier
	bucket := rl.getOrCreateBucket(identifier)

	// Check if request is allowed
	allowed := bucket.consume(cost)

	// Update statistics
	rl.mu.Lock()
	rl.totalRequests++
	if !allowed {
		rl.blockedRequests++
	}
	rl.mu.Unlock()

	if !allowed {
		rl.logger.Debug("Request rate limited", "identifier", identifier, "cost", cost)
	}

	return allowed
}

// getOrCreateBucket gets or creates a token bucket for the given identifier
func (rl *RateLimiter) getOrCreateBucket(identifier string) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[identifier]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[identifier]; exists {
		return bucket
	}

	bucket = &TokenBucket{
		tokens:     float64(rl.config.BurstSize), // Start with full bucket
		capacity:   float64(rl.config.BurstSize),
		refillRate: float64(rl.config.MaxRequestsPerSecond),
		lastRefill: time.Now(),
	}

	rl.buckets[identifier] = bucket
	return bucket
}

// consume attempts to consume tokens from the bucket
func (tb *TokenBucket) consume(tokens float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap at capacity
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

// cleanupRoutine periodically cleans up old buckets
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !rl.isRunning {
				return
			}
			rl.cleanup()
		}
	}
}

// cleanup removes old unused buckets
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.config.BucketTTL)
	removed := 0

	for identifier, bucket := range rl.buckets {
		bucket.mu.Lock()
		lastUsed := bucket.lastRefill
		bucket.mu.Unlock()

		if lastUsed.Before(cutoff) {
			delete(rl.buckets, identifier)
			removed++
		}
	}

	rl.lastCleanup = now

	if removed > 0 {
		rl.logger.Debug("Cleaned up old rate limit buckets", "removed", removed)
	}
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	blockedPercentage := 0.0
	if rl.totalRequests > 0 {
		blockedPercentage = float64(rl.blockedRequests) / float64(rl.totalRequests) * 100
	}

	return map[string]interface{}{
		"is_running":         rl.isRunning,
		"total_requests":     rl.totalRequests,
		"blocked_requests":   rl.blockedRequests,
		"blocked_percentage": blockedPercentage,
		"active_buckets":     len(rl.buckets),
		"last_cleanup":       rl.lastCleanup,
		"config": map[string]interface{}{
			"max_requests_per_second": rl.config.MaxRequestsPerSecond,
			"burst_size":              rl.config.BurstSize,
			"cleanup_interval":        rl.config.CleanupInterval,
			"bucket_ttl":              rl.config.BucketTTL,
		},
	}
}

// GetBucketInfo returns information about a specific bucket
func (rl *RateLimiter) GetBucketInfo(identifier string) map[string]interface{} {
	rl.mu.RLock()
	bucket, exists := rl.buckets[identifier]
	rl.mu.RUnlock()

	if !exists {
		return map[string]interface{}{
			"exists": false,
		}
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return map[string]interface{}{
		"exists":      true,
		"tokens":      bucket.tokens,
		"capacity":    bucket.capacity,
		"refill_rate": bucket.refillRate,
		"last_refill": bucket.lastRefill,
	}
}

// Reset resets the rate limiter statistics
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.totalRequests = 0
	rl.blockedRequests = 0
	rl.buckets = make(map[string]*TokenBucket)
	rl.lastCleanup = time.Now()

	rl.logger.Info("Rate limiter statistics reset")
}

// IsRunning returns whether the rate limiter is running
func (rl *RateLimiter) IsRunning() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.isRunning
}

// GetActiveBucketCount returns the number of active buckets
func (rl *RateLimiter) GetActiveBucketCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.buckets)
}

// SetConfig updates the rate limiter configuration
func (rl *RateLimiter) SetConfig(config *types.RateLimiterConfig) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if config.MaxRequestsPerSecond <= 0 {
		return fmt.Errorf("max requests per second must be positive")
	}

	if config.BurstSize <= 0 {
		return fmt.Errorf("burst size must be positive")
	}

	rl.config = config
	rl.logger.Info("Rate limiter configuration updated",
		"max_requests_per_second", config.MaxRequestsPerSecond,
		"burst_size", config.BurstSize)

	return nil
}
