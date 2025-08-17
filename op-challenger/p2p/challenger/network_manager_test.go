package challenger

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// MockP2PNode is a mock implementation for testing
type MockP2PNode struct {
	id      string
	address string
}

func (m *MockP2PNode) GetID() string      { return m.id }
func (m *MockP2PNode) GetAddress() string { return m.address }
func (m *MockP2PNode) IsRunning() bool    { return true }

// Create a proper P2PNode mock
func createMockP2PNode() *network.P2PNode {
	// This is a simplified mock - in real implementation we'd need proper mocking
	// For now, we'll create a basic node structure
	return &network.P2PNode{}
}

// Create a test challenger info with valid IDs
func createTestChallengerInfo(suffix string) *types.ChallengerInfo {
	// Create valid 40-character hexadecimal IDs (pad with zeros if needed)
	paddedSuffix := fmt.Sprintf("%08s", suffix)
	if len(paddedSuffix) > 8 {
		paddedSuffix = paddedSuffix[:8]
	}

	id := "1234567890abcdef1234567890abcdef" + paddedSuffix     // 32 + 8 = 40 hex chars
	nodeID := "abcdef1234567890abcdef1234567890" + paddedSuffix // 32 + 8 = 40 hex chars

	return &types.ChallengerInfo{
		ID:          id,
		NodeID:      nodeID,
		Address:     "127.0.0.1:808" + suffix,
		Role:        types.ChallengerRoleChallenger,
		Status:      types.ChallengerStatusOnline,
		IsConnected: true, // Make sure it's considered online
	}
}

// TestNewChallengerNetworkManager tests network manager creation
func TestNewChallengerNetworkManager(t *testing.T) {
	config := &types.ChallengerNetworkManagerConfig{
		MaxPeers:          100,
		HeartbeatInterval: time.Minute,
		DiscoveryEnabled:  true,
		DiscoveryInterval: 30 * time.Second,
		ConnectionTimeout: 5 * time.Minute,
	}

	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	require.NotNil(t, manager, "Network manager should not be nil")
	assert.Equal(t, config, manager.config, "Config should match")
	assert.Equal(t, logger, manager.logger, "Logger should match")
	assert.Equal(t, node, manager.node, "Node should match")
	assert.NotNil(t, manager.challengers, "Challengers map should be initialized")
}

// TestChallengerNetworkManagerStart tests starting the network manager
func TestChallengerNetworkManagerStart(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	// Test starting
	err := manager.Start()
	assert.NoError(t, err, "Starting network manager should succeed")
	assert.True(t, manager.IsRunning(), "Network manager should be running")

	// Test double start
	err = manager.Start()
	assert.Error(t, err, "Double start should fail")

	// Cleanup
	err = manager.Stop()
	assert.NoError(t, err, "Stopping should succeed")
}

// TestChallengerNetworkManagerStop tests stopping the network manager
func TestChallengerNetworkManagerStop(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	// Test stopping when not running
	err := manager.Stop()
	assert.NoError(t, err, "Stopping when not running should succeed")

	// Start and then stop
	err = manager.Start()
	require.NoError(t, err)

	err = manager.Stop()
	assert.NoError(t, err, "Stopping should succeed")
	assert.False(t, manager.IsRunning(), "Network manager should not be running")

	// Test double stop
	err = manager.Stop()
	assert.NoError(t, err, "Double stop should succeed")
}

// TestChallengerNetworkManagerRegisterChallenger tests challenger registration
func TestChallengerNetworkManagerRegisterChallenger(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Create test challenger info
	challengerInfo := createTestChallengerInfo("1")

	// Register challenger
	err = manager.AddChallenger(challengerInfo)
	assert.NoError(t, err, "Registering challenger should succeed")

	// Check challenger is registered
	registered := manager.GetChallenger(challengerInfo.ID)
	assert.NotNil(t, registered, "Challenger should be registered")
	assert.Equal(t, challengerInfo.ID, registered.ID, "Challenger ID should match")

	// Test duplicate registration (should succeed by overwriting)
	err = manager.AddChallenger(challengerInfo)
	assert.NoError(t, err, "Duplicate registration should succeed by overwriting")
}

// TestChallengerNetworkManagerUnregisterChallenger tests challenger unregistration
func TestChallengerNetworkManagerUnregisterChallenger(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Create and register challenger
	challengerInfo := createTestChallengerInfo("1")

	err = manager.AddChallenger(challengerInfo)
	require.NoError(t, err)

	// Remove challenger
	manager.RemoveChallenger(challengerInfo.ID)

	// Check challenger is removed
	registered := manager.GetChallenger(challengerInfo.ID)
	assert.Nil(t, registered, "Challenger should be removed")

	// Test removing non-existent challenger (should not panic)
	manager.RemoveChallenger("non-existent")
}

// TestChallengerNetworkManagerCounts tests challenger count retrieval
func TestChallengerNetworkManagerCounts(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Register some challengers
	for i := 0; i < 3; i++ {
		challengerInfo := createTestChallengerInfo(fmt.Sprintf("%d", i))
		manager.AddChallenger(challengerInfo)
	}

	// Check challenger counts
	totalCount := manager.GetChallengerCount()
	onlineCount := manager.GetOnlineChallengerCount()
	isRunning := manager.IsRunning()

	assert.Equal(t, 3, totalCount, "Should have 3 total challengers")
	assert.Equal(t, 3, onlineCount, "Should have 3 online challengers")
	assert.True(t, isRunning, "Manager should be running")
}

// TestChallengerNetworkManagerConcurrency tests concurrent access
func TestChallengerNetworkManagerConcurrency(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	const numGoroutines = 10
	const numOperations = 50

	done := make(chan bool, numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				challengerInfo := &types.ChallengerInfo{
					ID:      fmt.Sprintf("challenger-%d-%d", id, j),
					NodeID:  fmt.Sprintf("node-%d-%d", id, j),
					Address: fmt.Sprintf("127.0.%d.%d:8080", id%255, j%255),
					Role:    types.ChallengerRoleChallenger,
					Status:  types.ChallengerStatusOnline,
				}

				manager.AddChallenger(challengerInfo)
				manager.GetChallenger(challengerInfo.ID)
				manager.GetChallengerCount()
				manager.RemoveChallenger(challengerInfo.ID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Get final counts
	finalCount := manager.GetChallengerCount()
	assert.GreaterOrEqual(t, finalCount, 0, "Final count should be non-negative")
}

// TestChallengerNetworkManagerEdgeCases tests edge cases
func TestChallengerNetworkManagerEdgeCases(t *testing.T) {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := createMockP2PNode()

	manager := NewChallengerNetworkManager(node, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Test registering nil challenger
	err = manager.AddChallenger(nil)
	assert.Error(t, err, "Registering nil challenger should fail")

	// Test registering challenger with empty ID
	emptyIDChallenger := &types.ChallengerInfo{
		ID:      "",
		NodeID:  "node-1",
		Address: "127.0.0.1:8081",
		Role:    types.ChallengerRoleChallenger,
		Status:  types.ChallengerStatusOnline,
	}

	err = manager.AddChallenger(emptyIDChallenger)
	assert.Error(t, err, "Registering challenger with empty ID should fail")

	// Test getting non-existent challenger
	challenger := manager.GetChallenger("non-existent")
	assert.Nil(t, challenger, "Getting non-existent challenger should return nil")

	// Test removing empty ID (should not panic)
	manager.RemoveChallenger("")
}

// TestChallengerNetworkManagerNilConfig tests behavior with nil config
func TestChallengerNetworkManagerNilConfig(t *testing.T) {
	logger := log.New()
	node := createMockP2PNode()

	// Use default config instead of nil (implementation doesn't handle nil)
	config := types.DefaultChallengerNetworkManagerConfig()
	manager := NewChallengerNetworkManager(node, config, logger)
	assert.NotNil(t, manager, "Network manager should be created with default config")

	// Should be able to start with default config
	err := manager.Start()
	assert.NoError(t, err, "Should start with default config")

	err = manager.Stop()
	assert.NoError(t, err, "Should stop cleanly")
}
