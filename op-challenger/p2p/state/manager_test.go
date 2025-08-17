package state

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/challenger"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// Helper function to create a mock network manager
func createMockNetworkManager() *challenger.ChallengerNetworkManager {
	config := types.DefaultChallengerNetworkManagerConfig()
	logger := log.New()
	node := &network.P2PNode{}
	return challenger.NewChallengerNetworkManager(node, config, logger)
}

// TestNewChallengerStateManager tests state manager creation
func TestNewChallengerStateManager(t *testing.T) {
	config := DefaultChallengerStateManagerConfig()

	logger := log.New()
	networkManager := createMockNetworkManager()
	manager := NewChallengerStateManager(networkManager, config, logger)

	require.NotNil(t, manager, "State manager should not be nil")
	assert.Equal(t, config, manager.config, "Config should match")
	assert.Equal(t, logger, manager.logger, "Logger should match")
	assert.NotNil(t, manager.states, "States map should be initialized")
}

// TestChallengerStateManagerStart tests starting the state manager
func TestChallengerStateManagerStart(t *testing.T) {
	config := DefaultChallengerStateManagerConfig()
	logger := log.New()
	networkManager := createMockNetworkManager()
	manager := NewChallengerStateManager(networkManager, config, logger)

	err := manager.Start()
	require.NoError(t, err, "Starting state manager should not error")

	assert.True(t, manager.IsRunning(), "Manager should be running after start")

	err = manager.Stop()
	require.NoError(t, err, "Stopping state manager should not error")
}

// TestChallengerStateManagerStop tests stopping the state manager
func TestChallengerStateManagerStop(t *testing.T) {
	config := DefaultChallengerStateManagerConfig()
	logger := log.New()
	networkManager := createMockNetworkManager()
	manager := NewChallengerStateManager(networkManager, config, logger)

	// Start first
	err := manager.Start()
	require.NoError(t, err)

	// Then stop
	err = manager.Stop()
	require.NoError(t, err, "Stopping should succeed")
	assert.False(t, manager.IsRunning(), "Manager should not be running after stop")
}

// TestChallengerStateManagerUpdateState tests state updates
func TestChallengerStateManagerUpdateState(t *testing.T) {
	config := DefaultChallengerStateManagerConfig()
	logger := log.New()
	networkManager := createMockNetworkManager()
	manager := NewChallengerStateManager(networkManager, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Create test state
	challengerID := "1234567890abcdef1234567890abcdef12345678"
	state := &types.ChallengerState{
		ChallengerID: challengerID,
		NodeID:       "abcdef1234567890abcdef1234567890abcdef12", // Add required NodeID
		Address:      "127.0.0.1:8080",                           // Add required Address
		Status:       types.ChallengerStatusOnline,
		LastUpdated:  time.Now(),
		IsOnline:     true,
		IsConnected:  true,
	}

	// Update state
	err = manager.UpdateChallengerState(challengerID, state)
	assert.NoError(t, err, "Updating state should succeed")

	// Get state
	retrievedState := manager.GetChallengerState(challengerID)
	assert.NotNil(t, retrievedState, "State should be retrievable")
	assert.Equal(t, challengerID, retrievedState.ChallengerID, "Challenger ID should match")
}

// TestChallengerStateManagerGetAllStates tests getting all states
func TestChallengerStateManagerGetAllStates(t *testing.T) {
	config := DefaultChallengerStateManagerConfig()
	logger := log.New()
	networkManager := createMockNetworkManager()
	manager := NewChallengerStateManager(networkManager, config, logger)

	err := manager.Start()
	require.NoError(t, err)
	defer manager.Stop()

	// Add multiple states
	for i := 0; i < 3; i++ {
		challengerID := "1234567890abcdef1234567890abcdef1234567" + string(rune('0'+i))
		state := &types.ChallengerState{
			ChallengerID: challengerID,
			NodeID:       "abcdef1234567890abcdef1234567890abcdef1" + string(rune('0'+i)),
			Address:      "127.0.0.1:808" + string(rune('0'+i)),
			Status:       types.ChallengerStatusOnline,
			LastUpdated:  time.Now(),
			IsOnline:     true,
			IsConnected:  true,
		}
		err = manager.UpdateChallengerState(challengerID, state)
		require.NoError(t, err)
	}

	// Get all states
	allStates := manager.GetAllStates()
	assert.Len(t, allStates, 3, "Should have 3 states")
}
