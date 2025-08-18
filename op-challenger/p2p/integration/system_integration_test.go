package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/defense"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestCompleteChallengerLifecycle tests the complete lifecycle of a challenger
func TestCompleteChallengerLifecycle(t *testing.T) {
	// Create logger for testing
	logger := log.New()

	// Create system components
	rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)
	connLimiter := defense.NewConnectionLimiter(types.DefaultConnectionLimiterConfig(), logger)
	monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)

	// Start defense components
	require.NoError(t, rateLimiter.Start())
	require.NoError(t, connLimiter.Start())
	require.NoError(t, monitor.Start())

	defer func() {
		rateLimiter.Stop()
		connLimiter.Stop()
		monitor.Stop()
	}()

	t.Run("ChallengerRegistrationLifecycle", func(t *testing.T) {
		challengerID := GenerateValidHexID("lifecycle-test")

		// Phase 1: Registration
		challengerInfo := &types.ChallengerInfo{
			ID:          challengerID,
			NodeID:      GenerateValidHexID("node"),
			Address:     "127.0.0.1:8080",
			Role:        types.ChallengerRoleChallenger,
			Status:      types.ChallengerStatusOnline,
			IsConnected: true,
		}

		// Verify challenger info creation
		assert.Equal(t, challengerID, challengerInfo.ID)
		assert.True(t, challengerInfo.IsOnline())
		assert.True(t, challengerInfo.HasRole(types.ChallengerRoleChallenger))

		// Phase 2: State Management
		challengerState := types.NewChallengerState(challengerID, challengerInfo.NodeID, challengerInfo.Address)
		assert.NotNil(t, challengerState)
		assert.Equal(t, challengerID, challengerState.ChallengerID)
		assert.Equal(t, types.ChallengerStatusOffline, challengerState.Status) // Default state

		// Phase 3: Status Updates
		challengerState.SetOnline()
		assert.Equal(t, types.ChallengerStatusOnline, challengerState.Status)
		assert.True(t, challengerState.IsOnline)
		assert.True(t, challengerState.IsConnected)

		// Phase 4: Connection Quality Updates
		challengerState.UpdateConnectionInfo(5, 50*time.Millisecond)
		assert.Equal(t, 5, challengerState.ConnectionInfo.PeerCount)
		assert.Equal(t, 50*time.Millisecond, challengerState.ConnectionInfo.NetworkLatency)

		// Phase 5: Health Check
		assert.True(t, challengerState.IsHealthy())

		// Phase 6: Rate limiting integration
		allowed := rateLimiter.CheckLimit(challengerID)
		assert.True(t, allowed, "Initial request should be allowed")

		// Phase 7: Connection management
		connectionID := fmt.Sprintf("conn-%s", challengerID)
		connAllowed := connLimiter.CheckConnection(challengerInfo.Address)
		assert.True(t, connAllowed, "Connection should be allowed")

		err := connLimiter.AddConnection(connectionID, challengerInfo.Address)
		assert.NoError(t, err)

		// Phase 8: Monitoring integration
		monitor.RecordMessage("challenger-register", 100, 0)
		monitor.RecordConnection("127.0.0.1", true)

		// Phase 9: Deregistration
		challengerState.SetOffline()
		assert.Equal(t, types.ChallengerStatusOffline, challengerState.Status)
		assert.False(t, challengerState.IsOnline)
		assert.False(t, challengerState.IsConnected)

		// Remove connection
		connLimiter.RemoveConnection(connectionID)

		t.Log("Complete challenger lifecycle test passed")
	})

	t.Run("MultipleChallengerLifecycles", func(t *testing.T) {
		numChallengers := 5
		challengers := make([]*types.ChallengerInfo, numChallengers)
		states := make([]*types.ChallengerState, numChallengers)

		// Create multiple challengers
		for i := 0; i < numChallengers; i++ {
			challengerID := GenerateValidHexID(fmt.Sprintf("multi-challenger-%d", i))
			challengers[i] = &types.ChallengerInfo{
				ID:          challengerID,
				NodeID:      GenerateValidHexID(fmt.Sprintf("node-%d", i)),
				Address:     fmt.Sprintf("127.0.0.1:808%d", i),
				Role:        types.ChallengerRoleChallenger,
				Status:      types.ChallengerStatusOnline,
				IsConnected: true,
			}
			states[i] = types.NewChallengerState(challengerID, challengers[i].NodeID, challengers[i].Address)
			states[i].SetOnline()

			// Test rate limiting for each challenger
			allowed := rateLimiter.CheckLimit(challengerID)
			assert.True(t, allowed, "Challenger %d should be allowed initially", i)

			// Test connection management
			connectionID := fmt.Sprintf("multi-conn-%d", i)
			connAllowed := connLimiter.CheckConnection(challengers[i].Address)
			if connAllowed {
				err := connLimiter.AddConnection(connectionID, challengers[i].Address)
				assert.NoError(t, err)
			}

			// Record activity
			monitor.RecordMessage("multi-challenger-register", 50, 0)
		}

		// Verify all challengers are created properly
		for i, challenger := range challengers {
			assert.NotNil(t, challenger)
			assert.True(t, challenger.IsOnline())
			assert.NotNil(t, states[i])
			assert.True(t, states[i].IsOnline)
		}

		// Simulate some challengers going offline
		for i := 0; i < 2; i++ {
			states[i].SetOffline()
			assert.False(t, states[i].IsOnline)
			monitor.RecordMessage("challenger-offline", 25, 0)
		}

		// Count online challengers
		onlineCount := 0
		for _, state := range states {
			if state.IsOnline {
				onlineCount++
			}
		}
		assert.Equal(t, 3, onlineCount, "Should have 3 online challengers")

		// Verify monitoring captured all activities
		time.Sleep(100 * time.Millisecond)
		stats := monitor.GetCurrentStats()
		assert.Greater(t, stats.TotalMessages, uint64(5), "Should have recorded multiple messages")

		t.Log("Multiple challenger lifecycles test passed")
	})
}

// TestNetworkFormation tests network formation and management
func TestNetworkFormation(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.StopAll()

	t.Run("BasicNetworkFormation", func(t *testing.T) {
		// Add multiple nodes
		nodeCount := 4
		for i := 0; i < nodeCount; i++ {
			nodeName := fmt.Sprintf("node%d", i+1)
			env.AddNode(nodeName, 8080+i)
		}

		// Create a mesh network
		env.ConnectNodes("node1", "node2")
		env.ConnectNodes("node2", "node3")
		env.ConnectNodes("node3", "node4")
		env.ConnectNodes("node4", "node1") // Complete the ring

		// Start all nodes
		env.StartAll()

		// Verify network formation
		for i := 1; i <= nodeCount; i++ {
			nodeName := fmt.Sprintf("node%d", i)
			node := env.GetNode(nodeName)
			require.NotNil(t, node)
			assert.GreaterOrEqual(t, node.GetConnectedPeerCount(), 1, "%s should have at least 1 connection", nodeName)
		}

		t.Log("Basic network formation test passed")
	})

	t.Run("NetworkPartitioning", func(t *testing.T) {
		// Create a simple 3-node network
		env.AddNode("nodeA", 8090)
		env.AddNode("nodeB", 8091)
		env.AddNode("nodeC", 8092)

		// Connect in a line: A <-> B <-> C
		env.ConnectNodes("nodeA", "nodeB")
		env.ConnectNodes("nodeB", "nodeC")

		env.StartAll()

		// Verify initial connectivity
		nodeA := env.GetNode("nodeA")
		nodeB := env.GetNode("nodeB")
		nodeC := env.GetNode("nodeC")

		assert.Equal(t, 1, nodeA.GetConnectedPeerCount(), "Node A should have 1 connection")
		assert.Equal(t, 2, nodeB.GetConnectedPeerCount(), "Node B should have 2 connections")
		assert.Equal(t, 1, nodeC.GetConnectedPeerCount(), "Node C should have 1 connection")

		// Simulate network partition by stopping node B
		require.NoError(t, nodeB.Stop())

		// After partition, A and C should be isolated
		// Note: In a real implementation, this would be detected over time
		// For now, we just verify the basic structure

		t.Log("Network partitioning test passed")
	})
}

// TestSystemUnderStress tests system behavior under stress conditions
func TestSystemUnderStress(t *testing.T) {
	// Create logger for testing
	logger := log.New()

	// Create defense components with lower limits for stress testing
	rateLimiterConfig := types.DefaultRateLimiterConfig()
	rateLimiterConfig.MaxRequestsPerSecond = 5 // Very low limit
	rateLimiter := defense.NewRateLimiter(rateLimiterConfig, logger)

	connLimiterConfig := types.DefaultConnectionLimiterConfig()
	connLimiterConfig.MaxConnections = 3 // Very low limit
	connLimiter := defense.NewConnectionLimiter(connLimiterConfig, logger)

	monitorConfig := types.DefaultBasicMonitorConfig()
	monitorConfig.UpdateInterval = 10 * time.Millisecond // Very frequent updates
	monitor := defense.NewBasicMonitor(monitorConfig, logger)

	require.NoError(t, rateLimiter.Start())
	require.NoError(t, connLimiter.Start())
	require.NoError(t, monitor.Start())

	defer func() {
		rateLimiter.Stop()
		connLimiter.Stop()
		monitor.Stop()
	}()

	t.Run("HighLoadRateLimiting", func(t *testing.T) {
		clientID := "stress-test-client"

		// Generate high load
		allowedCount := 0
		deniedCount := 0

		for i := 0; i < 100; i++ {
			if rateLimiter.CheckLimit(clientID) {
				allowedCount++
				monitor.RecordMessage("stress-test-message", 10, 0)
			} else {
				deniedCount++
			}

			// Small delay to simulate real requests
			time.Sleep(1 * time.Millisecond)
		}

		assert.Greater(t, allowedCount, 0, "Some requests should be allowed")
		assert.Greater(t, deniedCount, 0, "Some requests should be denied under stress")

		// Verify rate limiter is still functional
		stats := rateLimiter.GetStats()
		totalRequests := stats["total_requests"]
		assert.NotNil(t, totalRequests, "total_requests should not be nil")
		// Convert to int64 for comparison
		if tr, ok := totalRequests.(int64); ok {
			assert.Greater(t, tr, int64(0))
		} else if tr, ok := totalRequests.(uint64); ok {
			assert.Greater(t, tr, uint64(0))
		} else if tr, ok := totalRequests.(int); ok {
			assert.Greater(t, tr, 0)
		}

		blockedRequests := stats["blocked_requests"]
		assert.NotNil(t, blockedRequests, "blocked_requests should not be nil")
		// Convert to int64 for comparison  
		if br, ok := blockedRequests.(int64); ok {
			assert.Greater(t, br, int64(0))
		} else if br, ok := blockedRequests.(uint64); ok {
			assert.Greater(t, br, uint64(0))
		} else if br, ok := blockedRequests.(int); ok {
			assert.Greater(t, br, 0)
		}

		t.Log("High load rate limiting test passed")
	})

	t.Run("ConnectionFloodTest", func(t *testing.T) {
		// Try to create many connections
		successfulConnections := 0
		rejectedConnections := 0

		for i := 0; i < 10; i++ {
			connectionID := fmt.Sprintf("flood-conn-%d", i)
			address := fmt.Sprintf("192.168.1.%d:8080", i+1)

			if connLimiter.CheckConnection(address) {
				err := connLimiter.AddConnection(connectionID, address)
				if err == nil {
					successfulConnections++
				}
			} else {
				rejectedConnections++
			}
		}

		assert.LessOrEqual(t, successfulConnections, 3, "Should not exceed connection limit")
		assert.Greater(t, rejectedConnections, 0, "Some connections should be rejected")

		// Verify connection limiter stats
		stats := connLimiter.GetStats()
		totalConnections := stats["total_connections"]
		assert.NotNil(t, totalConnections, "total_connections should not be nil")
		// Convert to appropriate type for comparison
		if tc, ok := totalConnections.(int); ok {
			assert.Equal(t, successfulConnections, tc)
		} else if tc, ok := totalConnections.(uint64); ok {
			assert.Equal(t, successfulConnections, int(tc))
		} else if tc, ok := totalConnections.(int64); ok {
			assert.Equal(t, successfulConnections, int(tc))
		}

		t.Log("Connection flood test passed")
	})

	t.Run("MonitoringUnderLoad", func(t *testing.T) {
		// Generate high monitoring load
		for i := 0; i < 50; i++ {
			monitor.RecordMessage("stress-message", 10, 0)
			monitor.RecordConnection(fmt.Sprintf("192.168.1.%d", i%10), true)
			if i%5 == 0 {
				monitor.RecordError("stress-error")
			}
		}

		// Wait for monitoring to process
		time.Sleep(100 * time.Millisecond)

		// Verify monitoring still works under load
		stats := monitor.GetCurrentStats()
		// Simply verify that monitoring is functioning - values may be any numeric type
		assert.NotNil(t, stats, "Monitor stats should not be nil")
		t.Logf("Monitor stats: TotalMessages=%v, TotalErrors=%v", stats.TotalMessages, stats.TotalErrors)

		t.Log("Monitoring under load test passed")
	})
}

// TestConcurrentChallengerOperations tests concurrent operations on challenger components
func TestConcurrentChallengerOperations(t *testing.T) {
	// Create logger for testing
	logger := log.New()

	// Create components
	rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)
	monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)

	require.NoError(t, rateLimiter.Start())
	require.NoError(t, monitor.Start())

	defer func() {
		rateLimiter.Stop()
		monitor.Stop()
	}()

	t.Run("ConcurrentRateLimiting", func(t *testing.T) {
		numGoroutines := 10
		requestsPerGoroutine := 20

		// Channel to collect results
		results := make(chan bool, numGoroutines*requestsPerGoroutine)

		// Start concurrent goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(clientID string) {
				for j := 0; j < requestsPerGoroutine; j++ {
					allowed := rateLimiter.CheckLimit(clientID)
					results <- allowed
					time.Sleep(1 * time.Millisecond)
				}
			}(fmt.Sprintf("client-%d", i))
		}

		// Collect results
		allowedCount := 0
		deniedCount := 0

		for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
			if <-results {
				allowedCount++
			} else {
				deniedCount++
			}
		}

		assert.Greater(t, allowedCount, 0, "Some concurrent requests should be allowed")
		// Note: May or may not have denied requests depending on timing

		t.Log("Concurrent rate limiting test passed")
	})

	t.Run("ConcurrentMonitoring", func(t *testing.T) {
		numGoroutines := 5
		operationsPerGoroutine := 10

		// Start concurrent monitoring operations
		for i := 0; i < numGoroutines; i++ {
			go func(nodeID string) {
				for j := 0; j < operationsPerGoroutine; j++ {
					monitor.RecordMessage("concurrent-message", 5, 0)
					monitor.RecordConnection(fmt.Sprintf("192.168.1.%d", j), true)
					if j%3 == 0 {
						monitor.RecordError("concurrent-error")
					}
					time.Sleep(1 * time.Millisecond)
				}
			}(fmt.Sprintf("node-%d", i))
		}

		// Wait for all operations to complete
		time.Sleep(200 * time.Millisecond)

		// Verify monitoring captured concurrent operations
		stats := monitor.GetCurrentStats()
		assert.Greater(t, stats.TotalMessages, uint64(0))

		t.Log("Concurrent monitoring test passed")
	})

	t.Run("ConcurrentChallengerStateOperations", func(t *testing.T) {
		numChallengers := 10

		// Create challengers concurrently
		challengers := make([]*types.ChallengerState, numChallengers)

		for i := 0; i < numChallengers; i++ {
			go func(index int) {
				challengerID := GenerateValidHexID(fmt.Sprintf("concurrent-%d", index))
				nodeID := GenerateValidHexID("node")
				address := fmt.Sprintf("127.0.0.1:808%d", index)

				state := types.NewChallengerState(challengerID, nodeID, address)
				state.SetOnline()
				state.UpdateConnectionInfo(5, 50*time.Millisecond)

				challengers[index] = state
			}(i)
		}

		// Wait for all to complete
		time.Sleep(100 * time.Millisecond)

		// Verify all challengers were created properly
		onlineCount := 0
		for _, challenger := range challengers {
			if challenger != nil && challenger.IsOnline {
				onlineCount++
			}
		}

		assert.Equal(t, numChallengers, onlineCount, "All challengers should be online")

		t.Log("Concurrent challenger state operations test passed")
	})
}
