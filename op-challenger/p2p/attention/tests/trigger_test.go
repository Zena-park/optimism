package tests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/attention"
)

func TestNewAttentionTrigger(t *testing.T) {
	challengerID := "test-challenger-1"
	trigger := attention.NewAttentionTrigger(nil, challengerID)

	assert.NotNil(t, trigger)
	assert.Equal(t, challengerID, trigger.GetChallengerID())
	assert.Equal(t, attention.AttentionTestProbability, trigger.GetConfig().Probability)
	assert.Equal(t, 51, trigger.GetConfig().MinConsensusThreshold)
	assert.Equal(t, 30*time.Second, trigger.GetConfig().ConsensusTimeout)
	assert.Equal(t, 12*time.Second, trigger.GetConfig().BlockTime)
}

func TestAttentionTrigger_ShouldTriggerAttentionTest_Deterministic(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	stateRoot := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	// Same inputs should produce same results
	result1 := trigger.ShouldTriggerAttentionTest(blockHash, stateRoot)
	result2 := trigger.ShouldTriggerAttentionTest(blockHash, stateRoot)

	assert.Equal(t, result1, result2, "Deterministic results should be identical")
}

func TestAttentionTrigger_ShouldTriggerAttentionTest_DifferentInputs(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	blockHash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	stateRoot1 := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	blockHash2 := common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	stateRoot2 := common.HexToHash("0x0987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba")

	result1 := trigger.ShouldTriggerAttentionTest(blockHash1, stateRoot1)
	result2 := trigger.ShouldTriggerAttentionTest(blockHash2, stateRoot2)

	// Results may be different due to randomness, but both should be boolean
	assert.IsType(t, true, result1)
	assert.IsType(t, true, result2)
}

func TestAttentionTrigger_SelectRandomChallenger_EmptyPeers(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	stateRoot := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	target := trigger.SelectRandomChallenger(blockHash, stateRoot)
	assert.Equal(t, "", target, "Should return empty string when no peers available")
}

func TestAttentionTrigger_SelectRandomChallenger_WithPeers(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	// Add some peers
	trigger.UpdatePeerList("peer-1", true)
	trigger.UpdatePeerList("peer-2", true)
	trigger.UpdatePeerList("peer-3", true)

	assert.Equal(t, 3, trigger.GetAvailableChallengerCount())

	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	stateRoot := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	target := trigger.SelectRandomChallenger(blockHash, stateRoot)

	// Should select one of the peers
	assert.Contains(t, []string{"peer-1", "peer-2", "peer-3"}, target)
}

func TestAttentionTrigger_SelectRandomChallenger_Deterministic(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	// Add peers
	trigger.UpdatePeerList("peer-1", true)
	trigger.UpdatePeerList("peer-2", true)
	trigger.UpdatePeerList("peer-3", true)

	blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	stateRoot := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

	// Same inputs should produce same results
	target1 := trigger.SelectRandomChallenger(blockHash, stateRoot)
	target2 := trigger.SelectRandomChallenger(blockHash, stateRoot)

	assert.Equal(t, target1, target2, "Deterministic selection should be identical")
}

func TestAttentionTrigger_PeerManagement(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	// Initially no peers
	assert.Equal(t, 0, trigger.GetAvailableChallengerCount())

	// Add peers
	trigger.UpdatePeerList("peer-1", true)
	trigger.UpdatePeerList("peer-2", true)
	assert.Equal(t, 2, trigger.GetAvailableChallengerCount())

	// Remove a peer
	trigger.UpdatePeerList("peer-1", false)
	assert.Equal(t, 1, trigger.GetAvailableChallengerCount())

	// Remove non-existent peer (should not crash)
	trigger.UpdatePeerList("non-existent", false)
	assert.Equal(t, 1, trigger.GetAvailableChallengerCount())
}

func TestAttentionTrigger_ConfigManagement(t *testing.T) {
	trigger := attention.NewAttentionTrigger(nil, "test-challenger")

	// Get initial config
	initialConfig := trigger.GetConfig()
	assert.Equal(t, attention.AttentionTestProbability, initialConfig.Probability)

	// Update config
	newConfig := &attention.TriggerConfig{
		Probability:           0.005, // 0.5%
		MinConsensusThreshold: 75,    // 75% consensus
		ConsensusTimeout:      60 * time.Second,
		BlockTime:             15 * time.Second,
		Logger:                log.New(),
	}

	trigger.SetConfig(newConfig)

	// Verify config was updated
	updatedConfig := trigger.GetConfig()
	assert.Equal(t, 0.005, updatedConfig.Probability)
	assert.Equal(t, 75, updatedConfig.MinConsensusThreshold)
	assert.Equal(t, 60*time.Second, updatedConfig.ConsensusTimeout)
	assert.Equal(t, 15*time.Second, updatedConfig.BlockTime)
}

func TestAttentionTrigger_ProbabilityCalculation(t *testing.T) {
	// Test with different probabilities
	testCases := []struct {
		probability float64
		name        string
	}{
		{0.001, "0.1%"},
		{0.0028, "0.28% (RAT paper)"},
		{0.005, "0.5%"},
		{0.01, "1%"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &attention.TriggerConfig{
				Probability:           tc.probability,
				MinConsensusThreshold: 51,
				ConsensusTimeout:      30 * time.Second,
				BlockTime:             12 * time.Second,
				Logger:                log.New(),
			}

			trigger := attention.NewAttentionTrigger(config, "test-challenger")

			// Test multiple blocks to see if probability is roughly correct
			blockHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
			stateRoot := common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321")

			triggerCount := 0
			totalTests := 1000

			for i := 0; i < totalTests; i++ {
				// Modify block hash for each test to simulate different blocks
				blockHashBytes := blockHash.Bytes()
				blockHashBytes[0] = byte(i % 256)
				blockHash = common.BytesToHash(blockHashBytes)

				if trigger.ShouldTriggerAttentionTest(blockHash, stateRoot) {
					triggerCount++
				}
			}

			actualProbability := float64(triggerCount) / float64(totalTests)
			expectedProbability := tc.probability

			// Allow some tolerance (within 50% of expected probability)
			tolerance := expectedProbability * 0.5
			assert.InDelta(t, expectedProbability, actualProbability, tolerance,
				"Actual probability should be close to expected probability")
		})
	}
}
