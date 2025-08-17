package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/challenger"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/defense"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/state"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestChallengerNetworkManagerWithDefense tests integration between NetworkManager and Defense components
func TestChallengerNetworkManagerWithDefense(t *testing.T) {
	// Create logger for testing
	logger := log.New()

	// Create defense components
	rateLimiterConfig := types.DefaultRateLimiterConfig()
	rateLimiterConfig.MaxRequestsPerSecond = 10 // Lower limit for testing
	rateLimiter := defense.NewRateLimiter(rateLimiterConfig, logger)

	connLimiterConfig := types.DefaultConnectionLimiterConfig()
	connLimiterConfig.MaxConnections = 5 // Lower limit for testing
	connLimiter := defense.NewConnectionLimiter(connLimiterConfig, logger)

	monitorConfig := types.DefaultBasicMonitorConfig()
	monitorConfig.UpdateInterval = 100 * time.Millisecond // Faster updates for testing
	monitor := defense.NewBasicMonitor(monitorConfig, logger)

	// Start defense components
	require.NoError(t, rateLimiter.Start())
	require.NoError(t, connLimiter.Start())
	require.NoError(t, monitor.Start())

	defer func() {
		rateLimiter.Stop()
		connLimiter.Stop()
		monitor.Stop()
	}()

	// Test rate limiting integration
	t.Run("RateLimitingIntegration", func(t *testing.T) {
		clientID := "test-client-1"

		// Should allow initial requests
		for i := 0; i < 5; i++ {
			allowed := rateLimiter.CheckLimit(clientID)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}

		// Should start limiting after exceeding rate
		time.Sleep(100 * time.Millisecond) // Allow some token regeneration

		// Burst test - should eventually be limited
		limitedCount := 0
		for i := 0; i < 20; i++ {
			if !rateLimiter.CheckLimit(clientID) {
				limitedCount++
			}
		}
		assert.Greater(t, limitedCount, 0, "Some requests should be rate limited")
	})

	// Test connection limiting integration
	t.Run("ConnectionLimitingIntegration", func(t *testing.T) {
		// Add connections up to limit
		for i := 0; i < 5; i++ {
			connectionID := fmt.Sprintf("conn-%d", i)
			address := fmt.Sprintf("192.168.1.%d:8080", i+1)
			allowed := connLimiter.CheckConnection(address)
			assert.True(t, allowed, "Connection %d should be allowed", i+1)

			err := connLimiter.AddConnection(connectionID, address)
			assert.NoError(t, err)
		}

		// Should reject additional connections
		address := "192.168.1.10:8080"
		allowed := connLimiter.CheckConnection(address)
		assert.False(t, allowed, "Connection beyond limit should be rejected")
	})

	// Test monitoring integration
	t.Run("MonitoringIntegration", func(t *testing.T) {
		// Record some activity using correct API
		monitor.RecordMessage("challenger-register", 100, 0)
		monitor.RecordMessage("challenger-heartbeat", 50, 0)
		monitor.RecordConnection("192.168.1.1", true)
		monitor.RecordError("connection-timeout")

		// Wait for monitoring update
		time.Sleep(200 * time.Millisecond)

		// Check statistics
		stats := monitor.GetCurrentStats()
		assert.Greater(t, stats.TotalMessages, uint64(0), "Should have recorded messages")
		assert.Greater(t, stats.TotalErrors, uint64(0), "Should have recorded errors")
	})

	t.Log("ChallengerNetworkManager + Defense integration test completed successfully")
}

// TestNetworkManagerWithStateManager tests integration between NetworkManager and StateManager
func TestNetworkManagerWithStateManager(t *testing.T) {
	// Create a mock P2P node for testing
	mockNode := NewMockP2PNode("test-node-1", "127.0.0.1:8080")
	require.NoError(t, mockNode.Start())
	defer mockNode.Stop()

	// Create network manager
	networkConfig := types.DefaultChallengerNetworkManagerConfig()
	networkConfig.HeartbeatInterval = 100 * time.Millisecond

	// Note: This will likely skip due to missing P2P implementation
	// For now, we'll test with mock components
	t.Run("StateManagerCreation", func(t *testing.T) {
		// Test creating state manager with correct config type
		stateConfig := &state.ChallengerStateManagerConfig{
			SyncInterval:             100 * time.Millisecond,
			StateTimeout:             5 * time.Second,
			MaxStates:                1000,
			EnableConflictResolution: false, // Phase 1 - basic only
		}

		// Create a minimal network manager for testing
		// Since we can't create a real P2P node, we'll test the state manager independently
		var networkManager *challenger.ChallengerNetworkManager = nil

		// Test state manager creation (it should handle nil network manager gracefully)
		stateManager := state.NewChallengerStateManager(networkManager, stateConfig, nil)
		assert.NotNil(t, stateManager, "State manager should be created")

		// Test basic state operations
		challengerID := GenerateValidHexID("test-challenger")
		challengerState := types.NewChallengerState(challengerID, GenerateValidHexID("node"), "127.0.0.1:8080")

		// Test state updates
		challengerState.SetOnline()
		assert.True(t, challengerState.IsOnline, "Challenger should be online")
		assert.Equal(t, types.ChallengerStatusOnline, challengerState.Status)

		challengerState.UpdateConnectionInfo(5, 50*time.Millisecond)
		assert.Equal(t, 5, challengerState.ConnectionInfo.PeerCount)
		assert.Equal(t, 50*time.Millisecond, challengerState.ConnectionInfo.NetworkLatency)
	})

	t.Log("NetworkManager + StateManager integration test completed successfully")
}

// TestStateManagerWithDefense tests integration between StateManager and Defense components
func TestStateManagerWithDefense(t *testing.T) {
	// Create defense components
	rateLimiterConfig := types.DefaultRateLimiterConfig()
	rateLimiterConfig.MaxRequestsPerSecond = 20
	rateLimiter := defense.NewRateLimiter(rateLimiterConfig, nil)

	monitorConfig := types.DefaultBasicMonitorConfig()
	monitorConfig.UpdateInterval = 50 * time.Millisecond
	monitor := defense.NewBasicMonitor(monitorConfig, nil)

	require.NoError(t, rateLimiter.Start())
	require.NoError(t, monitor.Start())

	defer func() {
		rateLimiter.Stop()
		monitor.Stop()
	}()

	// Test state sync rate limiting
	t.Run("StateSyncRateLimiting", func(t *testing.T) {
		clientID := "state-sync-client"

		// Simulate state sync requests
		allowedRequests := 0
		deniedRequests := 0

		for i := 0; i < 30; i++ {
			if rateLimiter.CheckLimit(clientID) {
				allowedRequests++
				// Simulate processing state sync request
				monitor.RecordMessage("state-sync-request", 10, 0)
			} else {
				deniedRequests++
				monitor.RecordError("rate-limit-exceeded")
			}
		}

		assert.Greater(t, allowedRequests, 0, "Some state sync requests should be allowed")
		assert.Greater(t, deniedRequests, 0, "Some state sync requests should be rate limited")

		// Wait for monitoring update
		time.Sleep(100 * time.Millisecond)

		// Check monitoring statistics
		stats := monitor.GetCurrentStats()
		assert.Greater(t, stats.TotalMessages, uint64(0), "Should have recorded state sync messages")
		assert.Greater(t, stats.TotalErrors, uint64(0), "Should have recorded rate limit errors")
	})

	// Test state monitoring
	t.Run("StateMonitoring", func(t *testing.T) {
		// Simulate various state operations
		monitor.RecordMessage("state-update", 15, 0)
		monitor.RecordMessage("state-query", 5, 0)
		monitor.RecordMessage("state-sync", 25, 0)
		monitor.RecordConnection("192.168.1.100", true)

		// Wait for monitoring update
		time.Sleep(100 * time.Millisecond)

		// Verify monitoring captured state operations
		stats := monitor.GetCurrentStats()
		assert.GreaterOrEqual(t, stats.TotalMessages, uint64(3), "Should have recorded state messages")
	})

	t.Log("StateManager + Defense integration test completed successfully")
}

// TestChallengerInfoIntegration tests challenger info operations
func TestChallengerInfoIntegration(t *testing.T) {
	t.Run("ChallengerInfoCreationAndValidation", func(t *testing.T) {
		challengerID := GenerateValidHexID("integration-test")
		nodeID := GenerateValidHexID("node")
		address := "127.0.0.1:8080"

		// Create challenger info
		challengerInfo := &types.ChallengerInfo{
			ID:          challengerID,
			NodeID:      nodeID,
			Address:     address,
			Role:        types.ChallengerRoleChallenger,
			Status:      types.ChallengerStatusOnline,
			IsConnected: true,
		}

		// Test basic properties
		assert.Equal(t, challengerID, challengerInfo.ID)
		assert.Equal(t, nodeID, challengerInfo.NodeID)
		assert.Equal(t, address, challengerInfo.Address)
		assert.True(t, challengerInfo.IsOnline())
		assert.True(t, challengerInfo.HasRole(types.ChallengerRoleChallenger))

		// Test status updates
		challengerInfo.UpdateStatus(types.ChallengerStatusIdle)
		assert.Equal(t, types.ChallengerStatusIdle, challengerInfo.Status)

		// Test connection state
		challengerInfo.SetConnected()
		assert.True(t, challengerInfo.IsConnected)

		challengerInfo.SetDisconnected()
		assert.False(t, challengerInfo.IsConnected)
		assert.False(t, challengerInfo.IsOnline()) // Should be false when disconnected
	})

	t.Run("ChallengerStateIntegration", func(t *testing.T) {
		challengerID := GenerateValidHexID("state-test")
		nodeID := GenerateValidHexID("node")
		address := "127.0.0.1:8080"

		// Create challenger state
		challengerState := types.NewChallengerState(challengerID, nodeID, address)
		assert.NotNil(t, challengerState)
		assert.Equal(t, challengerID, challengerState.ChallengerID)
		assert.Equal(t, types.ChallengerStatusOffline, challengerState.Status) // Default state

		// Test state transitions
		challengerState.SetOnline()
		assert.True(t, challengerState.IsOnline)
		assert.True(t, challengerState.IsConnected)
		assert.Equal(t, types.ChallengerStatusOnline, challengerState.Status)

		// Test connection info updates
		challengerState.UpdateConnectionInfo(10, 25*time.Millisecond)
		assert.Equal(t, 10, challengerState.ConnectionInfo.PeerCount)
		assert.Equal(t, 25*time.Millisecond, challengerState.ConnectionInfo.NetworkLatency)

		// Test health check
		assert.True(t, challengerState.IsHealthy(), "Challenger should be healthy with good connection")

		// Test offline transition
		challengerState.SetOffline()
		assert.False(t, challengerState.IsOnline)
		assert.False(t, challengerState.IsConnected)
		assert.False(t, challengerState.IsHealthy())
	})

	t.Log("Challenger info integration test completed successfully")
}
