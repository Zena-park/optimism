package defense

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestNewConnectionLimiter tests connection limiter creation
func TestNewConnectionLimiter(t *testing.T) {
	config := &types.ConnectionLimiterConfig{
		Enabled:         true,
		MaxConnections:  100,
		CleanupInterval: time.Minute,
	}

	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	require.NotNil(t, limiter, "Connection limiter should not be nil")
	assert.Equal(t, config, limiter.config, "Config should match")
	assert.Equal(t, logger, limiter.logger, "Logger should match")
}

// TestConnectionLimiterStart tests starting the connection limiter
func TestConnectionLimiterStart(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	// Test starting
	err := limiter.Start()
	assert.NoError(t, err, "Starting connection limiter should succeed")
	assert.True(t, limiter.IsRunning(), "Connection limiter should be running")

	// Test double start
	err = limiter.Start()
	assert.Error(t, err, "Double start should fail")

	// Cleanup
	limiter.Stop()
}

// TestConnectionLimiterStop tests stopping the connection limiter
func TestConnectionLimiterStop(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	// Test stopping when not running
	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Connection limiter should not be running")

	// Start and then stop
	err := limiter.Start()
	require.NoError(t, err)

	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Connection limiter should not be running")

	// Test double stop
	limiter.Stop()
	assert.False(t, limiter.IsRunning(), "Connection limiter should still not be running")
}

// TestConnectionLimiterCheckConnection tests connection checking
func TestConnectionLimiterCheckConnection(t *testing.T) {
	config := &types.ConnectionLimiterConfig{
		Enabled:             true,
		MaxConnections:      2,  // Very low limit for testing
		MaxConnectionsPerIP: 10, // Allow connections per IP
		CleanupInterval:     time.Minute,
	}

	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	// CheckConnection only checks if connection would be allowed
	// It doesn't actually add the connection, so all should be allowed initially
	allowed := limiter.CheckConnection("127.0.0.1:8080")
	if !allowed {
		t.Logf("Connection check failed for 127.0.0.1:8080")
	}
	assert.True(t, allowed, "Connection check should be allowed")

	allowed = limiter.CheckConnection("127.0.0.2:8080")
	if !allowed {
		t.Logf("Connection check failed for 127.0.0.2:8080")
	}
	assert.True(t, allowed, "Connection check should be allowed")

	// Add actual connections to test limits
	err = limiter.AddConnection("conn1", "127.0.0.1:8080")
	assert.NoError(t, err, "Adding connection should succeed")

	err = limiter.AddConnection("conn2", "127.0.0.2:8080")
	assert.NoError(t, err, "Adding connection should succeed")

	// Now check should fail due to max connections
	allowed = limiter.CheckConnection("127.0.0.3:8080")
	assert.False(t, allowed, "Connection check should be blocked when at max")
}

// TestConnectionLimiterAddRemoveConnection tests connection management
func TestConnectionLimiterAddRemoveConnection(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	connectionID := "conn-1"
	remoteAddr := "127.0.0.1:8080"

	// Add connection
	err = limiter.AddConnection(connectionID, remoteAddr)
	assert.NoError(t, err, "Adding connection should succeed")

	// Check connection info
	info := limiter.GetConnectionInfo(connectionID)
	assert.NotNil(t, info, "Connection info should exist")
	assert.Equal(t, "127.0.0.1", info.IP, "IP should be extracted correctly")

	// Update activity
	limiter.UpdateConnectionActivity(connectionID, 100, 200)

	// Get updated info
	info = limiter.GetConnectionInfo(connectionID)
	assert.Equal(t, int64(100), info.BytesSent, "Bytes sent should be updated")
	assert.Equal(t, int64(200), info.BytesReceived, "Bytes received should be updated")

	// Remove connection
	limiter.RemoveConnection(connectionID)

	// Check connection is removed
	info = limiter.GetConnectionInfo(connectionID)
	assert.Nil(t, info, "Connection info should be nil after removal")
}

// TestConnectionLimiterGetStats tests statistics retrieval
func TestConnectionLimiterGetStats(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	// Add some connections
	limiter.AddConnection("conn1", "127.0.0.1:8080")
	limiter.AddConnection("conn2", "127.0.0.2:8080")

	// Get stats
	stats := limiter.GetStats()
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Contains(t, stats, "total_connections", "Stats should contain total_connections")
	assert.Contains(t, stats, "config", "Stats should contain config")

	configMap, ok := stats["config"].(map[string]interface{})
	assert.True(t, ok, "Config should be a map")
	assert.Contains(t, configMap, "max_connections", "Config should contain max_connections")

	// Get IP stats
	ipStats := limiter.GetIPStats()
	assert.NotNil(t, ipStats, "IP stats should not be nil")
}

// TestConnectionLimiterGetConnectionsByIP tests IP-based connection retrieval
func TestConnectionLimiterGetConnectionsByIP(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	ip := "127.0.0.1"

	// Add connections for same IP
	limiter.AddConnection("conn1", ip+":8080")
	limiter.AddConnection("conn2", ip+":8081")
	limiter.AddConnection("conn3", "127.0.0.2:8080")

	// Get connections by IP
	connections := limiter.GetConnectionsByIP(ip)
	assert.Len(t, connections, 2, "Should have 2 connections for the IP")

	// Get all connections
	allConnections := limiter.GetAllConnections()
	assert.Len(t, allConnections, 3, "Should have 3 total connections")
}

// TestConnectionLimiterReset tests reset functionality
func TestConnectionLimiterReset(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	// Add connections
	limiter.AddConnection("conn1", "127.0.0.1:8080")
	limiter.AddConnection("conn2", "127.0.0.2:8080")

	// Check initial count
	initialCount := limiter.GetTotalConnections()
	assert.Equal(t, 2, initialCount, "Should have 2 connections initially")

	// Reset
	limiter.Reset()

	// Check count after reset
	finalCount := limiter.GetTotalConnections()
	assert.Equal(t, 0, finalCount, "Should have 0 connections after reset")
}

// TestConnectionLimiterConcurrency tests concurrent access
func TestConnectionLimiterConcurrency(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	const numGoroutines = 10
	const numOperations = 50

	done := make(chan bool, numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				connectionID := fmt.Sprintf("conn-%d-%d", id, j)
				remoteAddr := fmt.Sprintf("127.0.%d.%d:8080", id%255, j%255)

				limiter.CheckConnection(remoteAddr)
				limiter.AddConnection(connectionID, remoteAddr)
				limiter.UpdateConnectionActivity(connectionID, 100, 200)
				limiter.GetConnectionInfo(connectionID)
				limiter.RemoveConnection(connectionID)
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

// TestConnectionLimiterEdgeCases tests edge cases
func TestConnectionLimiterEdgeCases(t *testing.T) {
	config := types.DefaultConnectionLimiterConfig()
	logger := log.New()
	limiter := NewConnectionLimiter(config, logger)

	err := limiter.Start()
	require.NoError(t, err)
	defer limiter.Stop()

	// Test empty connection ID (actually allowed)
	err = limiter.AddConnection("", "127.0.0.1:8080")
	assert.NoError(t, err, "Empty connection ID is allowed")

	// Test empty remote address
	err = limiter.AddConnection("conn1", "")
	assert.Error(t, err, "Empty remote address should be rejected")

	// Test invalid remote address
	err = limiter.AddConnection("conn2", "invalid-address")
	assert.Error(t, err, "Invalid remote address should be rejected")

	// Test removing non-existent connection
	limiter.RemoveConnection("non-existent")
	// Should not panic

	// Test getting info for non-existent connection
	info := limiter.GetConnectionInfo("non-existent")
	assert.Nil(t, info, "Non-existent connection should return nil")

	// Test updating activity for non-existent connection
	limiter.UpdateConnectionActivity("non-existent", 100, 200)
	// Should not panic

	// Test duplicate connection ID
	err = limiter.AddConnection("conn1", "127.0.0.1:8080")
	assert.NoError(t, err, "First connection should succeed")

	err = limiter.AddConnection("conn1", "127.0.0.2:8080")
	assert.Error(t, err, "Duplicate connection ID should be rejected")
}

// TestConnectionLimiterNilConfig tests behavior with nil config
func TestConnectionLimiterNilConfig(t *testing.T) {
	logger := log.New()

	// This should not panic but use default config
	limiter := NewConnectionLimiter(nil, logger)
	assert.NotNil(t, limiter, "Connection limiter should handle nil config")

	// Should be able to start with default config
	err := limiter.Start()
	assert.NoError(t, err, "Should start with default config")

	limiter.Stop()
}
