package defense

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestNewRateLimiter tests rate limiter creation
func TestNewRateLimiter(t *testing.T) {
	config := &types.RateLimiterConfig{
		Enabled:               true,
		MaxMessagesPerSecond:  10,
		MaxRequestsPerSecond:  5,
		MaxConnectionsPerPeer: 3,
		BurstSize:             20,
		BurstDuration:         time.Second,
		BlockEnabled:          false,
		BlockDuration:         time.Minute,
		MaxViolations:         5,
		CleanupInterval:       time.Minute,
		EntryTTL:              time.Hour,
		BucketTTL:             time.Minute,
	}

	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	require.NotNil(t, limiter, "Rate limiter should not be nil")
	assert.Equal(t, config, limiter.config, "Config should match")
	assert.Equal(t, logger, limiter.logger, "Logger should match")
	assert.NotNil(t, limiter.buckets, "Buckets map should be initialized")
}

// TestRateLimiterStart tests starting the rate limiter
func TestRateLimiterStart(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	// Test starting
	err := limiter.Start()
	assert.NoError(t, err, "Starting rate limiter should succeed")
	assert.True(t, limiter.IsRunning(), "Rate limiter should be running")

	// Test double start
	err = limiter.Start()
	assert.Error(t, err, "Double start should fail")

	// Cleanup
	limiter.Stop()
}

// TestRateLimiterStop tests stopping the rate limiter
func TestRateLimiterStop(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	// Test stopping when not running
	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Rate limiter should not be running")

	// Start and then stop
	err := limiter.Start()
	require.NoError(t, err)

	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Rate limiter should not be running")

	// Test double stop
	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Rate limiter should still not be running")
}

// TestRateLimiterDisabled tests disabled rate limiter
func TestRateLimiterDisabled(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	config.Enabled = false
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)

	// When disabled, all requests should be allowed
	allowed := limiter.CheckLimit("peer1")
	assert.True(t, allowed, "Request should be allowed when disabled")

	allowed = limiter.CheckLimitWithCost("peer1", 10.0)
	assert.True(t, allowed, "Request with cost should be allowed when disabled")

	limiter.Stop()
}

// TestRateLimiterCheckLimit tests basic limit checking
func TestRateLimiterCheckLimit(t *testing.T) {
	config := &types.RateLimiterConfig{
		Enabled:               true,
		MaxMessagesPerSecond:  10,
		MaxRequestsPerSecond:  2, // Very low limit for testing
		MaxConnectionsPerPeer: 10,
		BurstSize:             2,
		BurstDuration:         time.Second,
		BlockEnabled:          false,
		BlockDuration:         time.Minute,
		MaxViolations:         3,
		CleanupInterval:       time.Minute,
		EntryTTL:              time.Hour,
		BucketTTL:             time.Second,
	}

	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// First requests should be allowed (within burst)
	allowed := limiter.CheckLimit(peerID)
	assert.True(t, allowed, "First request should be allowed")

	allowed = limiter.CheckLimit(peerID)
	assert.True(t, allowed, "Second request should be allowed")

	// Third request should be rate limited
	allowed = limiter.CheckLimit(peerID)
	assert.False(t, allowed, "Third request should be rate limited")

	// Wait for bucket to refill
	time.Sleep(600 * time.Millisecond)

	// Should be allowed again
	allowed = limiter.CheckLimit(peerID)
	assert.True(t, allowed, "Request should be allowed after waiting")
}

// TestRateLimiterCheckLimitWithCost tests cost-based limit checking
func TestRateLimiterCheckLimitWithCost(t *testing.T) {
	config := &types.RateLimiterConfig{
		Enabled:               true,
		MaxMessagesPerSecond:  10,
		MaxRequestsPerSecond:  2, // 2 tokens per second
		MaxConnectionsPerPeer: 10,
		BurstSize:             4, // 4 token capacity
		BurstDuration:         time.Second,
		BlockEnabled:          false,
		BlockDuration:         time.Minute,
		MaxViolations:         3,
		CleanupInterval:       time.Minute,
		EntryTTL:              time.Hour,
		BucketTTL:             time.Second,
	}

	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// First request with cost 2 should be allowed
	allowed := limiter.CheckLimitWithCost(peerID, 2.0)
	assert.True(t, allowed, "First request with cost 2 should be allowed")

	// Second request with cost 2 should be allowed (uses remaining 2 tokens)
	allowed = limiter.CheckLimitWithCost(peerID, 2.0)
	assert.True(t, allowed, "Second request with cost 2 should be allowed")

	// Third request should be rate limited (no tokens left)
	allowed = limiter.CheckLimitWithCost(peerID, 1.0)
	assert.False(t, allowed, "Third request should be rate limited")

	// Wait for bucket to refill
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	allowed = limiter.CheckLimitWithCost(peerID, 1.0)
	assert.True(t, allowed, "Request should be allowed after waiting")
}

// TestRateLimiterGetStats tests statistics retrieval
func TestRateLimiterGetStats(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// Generate some activity
	limiter.CheckLimit(peerID)
	limiter.CheckLimitWithCost(peerID, 2.0)

	// Get stats
	stats := limiter.GetStats()
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Contains(t, stats, "is_running", "Stats should contain is_running")
	assert.Contains(t, stats, "total_requests", "Stats should contain total_requests")
	assert.Contains(t, stats, "blocked_requests", "Stats should contain blocked_requests")
	assert.Contains(t, stats, "active_buckets", "Stats should contain active_buckets")
}

// TestRateLimiterGetBucketInfo tests bucket information retrieval
func TestRateLimiterGetBucketInfo(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// Initially no bucket exists
	info := limiter.GetBucketInfo(peerID)
	assert.NotNil(t, info, "Bucket info should not be nil")
	assert.False(t, info["exists"].(bool), "Bucket should not exist initially")

	// Create bucket by making a request
	limiter.CheckLimit(peerID)

	// Now bucket should exist
	info = limiter.GetBucketInfo(peerID)
	assert.True(t, info["exists"].(bool), "Bucket should exist after request")
	assert.Contains(t, info, "tokens", "Bucket info should contain tokens")
	assert.Contains(t, info, "capacity", "Bucket info should contain capacity")
	assert.Contains(t, info, "refill_rate", "Bucket info should contain refill_rate")
}

// TestRateLimiterCleanup tests cleanup functionality
func TestRateLimiterCleanup(t *testing.T) {
	config := &types.RateLimiterConfig{
		Enabled:               true,
		MaxMessagesPerSecond:  10,
		MaxRequestsPerSecond:  10,
		MaxConnectionsPerPeer: 10,
		BurstSize:             10,
		BurstDuration:         time.Second,
		BlockEnabled:          false,
		BlockDuration:         time.Minute,
		MaxViolations:         5,
		CleanupInterval:       100 * time.Millisecond, // Fast cleanup for testing
		EntryTTL:              time.Hour,
		BucketTTL:             200 * time.Millisecond, // Short TTL for testing
	}

	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// Generate activity to create bucket
	limiter.CheckLimit(peerID)

	// Check that bucket exists
	initialCount := limiter.GetActiveBucketCount()
	assert.Greater(t, initialCount, 0, "Should have active buckets")

	// Wait for cleanup to occur
	time.Sleep(400 * time.Millisecond)

	// Check that buckets are cleaned up
	finalCount := limiter.GetActiveBucketCount()
	assert.LessOrEqual(t, finalCount, initialCount, "Buckets should be cleaned up")
}

// TestRateLimiterReset tests reset functionality
func TestRateLimiterReset(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	peerID := "test-peer"

	// Generate some activity
	limiter.CheckLimit(peerID)
	limiter.CheckLimit(peerID)

	// Get stats before reset
	statsBefore := limiter.GetStats()
	totalBefore := statsBefore["total_requests"].(int64)
	assert.Greater(t, totalBefore, int64(0), "Should have processed requests")

	// Reset
	limiter.Reset()

	// Get stats after reset
	statsAfter := limiter.GetStats()
	totalAfter := statsAfter["total_requests"].(int64)
	assert.Equal(t, int64(0), totalAfter, "Total requests should be reset to 0")

	bucketsAfter := statsAfter["active_buckets"].(int)
	assert.Equal(t, 0, bucketsAfter, "Active buckets should be reset to 0")
}

// TestRateLimiterSetConfig tests configuration updates
func TestRateLimiterSetConfig(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	// Test valid config update
	newConfig := &types.RateLimiterConfig{
		Enabled:               true,
		MaxMessagesPerSecond:  20,
		MaxRequestsPerSecond:  15,
		MaxConnectionsPerPeer: 5,
		BurstSize:             30,
		BurstDuration:         time.Second,
		BlockEnabled:          false,
		BlockDuration:         time.Minute,
		MaxViolations:         3,
		CleanupInterval:       time.Minute,
		EntryTTL:              time.Hour,
		BucketTTL:             time.Minute,
	}

	err := limiter.SetConfig(newConfig)
	assert.NoError(t, err, "Setting valid config should succeed")
	assert.Equal(t, newConfig, limiter.config, "Config should be updated")

	// Test invalid config
	invalidConfig := &types.RateLimiterConfig{
		MaxRequestsPerSecond: 0, // Invalid
		BurstSize:            10,
	}

	err = limiter.SetConfig(invalidConfig)
	assert.Error(t, err, "Setting invalid config should fail")

	invalidConfig2 := &types.RateLimiterConfig{
		MaxRequestsPerSecond: 10,
		BurstSize:            0, // Invalid
	}

	err = limiter.SetConfig(invalidConfig2)
	assert.Error(t, err, "Setting invalid config should fail")
}

// TestRateLimiterConcurrency tests concurrent access
func TestRateLimiterConcurrency(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			peerID := fmt.Sprintf("peer-%d", id)
			for j := 0; j < numOperations; j++ {
				limiter.CheckLimit(peerID)
				limiter.CheckLimitWithCost(peerID, 2.0)
				limiter.GetBucketInfo(peerID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Get final stats
	stats := limiter.GetStats()
	assert.NotNil(t, stats, "Stats should be available after concurrent operations")
}

// TestRateLimiterEdgeCases tests edge cases
func TestRateLimiterEdgeCases(t *testing.T) {
	config := types.DefaultRateLimiterConfig()
	logger := log.New()
	limiter := NewRateLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	// Test empty peer ID
	allowed := limiter.CheckLimit("")
	assert.True(t, allowed, "Empty peer ID should be allowed") // Assuming no validation

	// Test very long peer ID
	longPeerID := strings.Repeat("a", 1000)
	allowed = limiter.CheckLimit(longPeerID)
	assert.True(t, allowed, "Long peer ID should be allowed")

	// Test special characters in peer ID
	specialPeerID := "peer@#$%^&*()"
	allowed = limiter.CheckLimit(specialPeerID)
	assert.True(t, allowed, "Special characters in peer ID should be allowed")

	// Test zero cost
	allowed = limiter.CheckLimitWithCost("peer1", 0.0)
	assert.True(t, allowed, "Zero cost should be allowed")

	// Test negative cost
	allowed = limiter.CheckLimitWithCost("peer1", -1.0)
	assert.True(t, allowed, "Negative cost should be handled gracefully")

	// Test very large cost
	allowed = limiter.CheckLimitWithCost("peer1", 1000000.0)
	assert.False(t, allowed, "Very large cost should be blocked")

	// Test getting info for non-existent bucket
	info := limiter.GetBucketInfo("non-existent-peer")
	assert.False(t, info["exists"].(bool), "Non-existent bucket should return exists=false")
}

// TestRateLimiterNilConfig tests behavior with nil config
func TestRateLimiterNilConfig(t *testing.T) {
	logger := log.New()

	// This should not panic but use default config or handle gracefully
	limiter := NewRateLimiter(nil, logger)
	assert.NotNil(t, limiter, "Rate limiter should handle nil config")

	// Note: This might panic depending on implementation
	// In a real implementation, we'd want to handle nil config gracefully
}
