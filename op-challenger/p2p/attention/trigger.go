package attention

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// DefaultTriggerRatio is the default trigger ratio for challengers
// Based on personalization: 10% trigger ratio (batch number % 100 < 10)
const DefaultTriggerRatio = 0.10

// TriggerConfig contains configuration for the attention trigger
type TriggerConfig struct {
	// Trigger ratio (percentage for batch number based triggering)
	// Example: 0.10 = 10% trigger ratio (batch number % 100 < 10)
	TriggerRatio float64
	// Minimum number of challengers required for consensus
	MinConsensusThreshold int
	// Timeout for consensus collection
	ConsensusTimeout time.Duration
	// Block time in seconds (12 seconds for Optimism)
	BlockTime time.Duration
	// L2 batch submission contract address
	BatchSubmissionAddress common.Address
	// Logger for the trigger
	Logger log.Logger
}

// TriggerResult represents the result of a trigger decision
type TriggerResult struct {
	ShouldTrigger    bool
	TargetChallenger string
	BlockHash        common.Hash
	StateRoot        common.Hash
	TriggerTime      time.Time
	Consensus        bool
	Votes            int
	TotalVotes       int
}

// AttentionTrigger handles distributed attention test triggering
type AttentionTrigger struct {
	config       *TriggerConfig
	challengerID string
	peers        map[string]bool // active peers
	mu           sync.RWMutex
	logger       log.Logger
	// RAT system monitoring
	monitoring bool
	monitorCtx context.Context
}

// NewAttentionTrigger creates a new attention trigger instance
func NewAttentionTrigger(config *TriggerConfig, challengerID string) *AttentionTrigger {
	if config == nil {
		config = &TriggerConfig{
			TriggerRatio:           DefaultTriggerRatio,
			MinConsensusThreshold:  51, // 51% consensus required
			ConsensusTimeout:       30 * time.Second,
			BlockTime:              12 * time.Second,
			BatchSubmissionAddress: common.Address{},
			Logger:                 log.New(),
		}
	}

	return &AttentionTrigger{
		config:       config,
		challengerID: challengerID,
		peers:        make(map[string]bool),
		logger:       config.Logger,
		monitoring:   false,
		monitorCtx:   nil,
	}
}

// ShouldTriggerAttentionTest determines if an attention test should be triggered
// Uses batch number based trigger ratio calculation
func (at *AttentionTrigger) ShouldTriggerAttentionTest(blockHash common.Hash, stateRoot common.Hash, blockNumber uint64, batchIndex uint64) bool {
	// Simple ratio calculation: batch number % 100 < trigger ratio
	return float64(batchIndex%100) < at.config.TriggerRatio*100
}

// ShouldTriggerManualTest determines if a manual trigger should be executed
// Called by operator to manually trigger attention test for specific challenger
func (at *AttentionTrigger) ShouldTriggerManualTest(targetChallengerID string) bool {
	// Check if target challenger is in active peers
	at.mu.RLock()
	defer at.mu.RUnlock()

	// Return true if target challenger is active
	return at.peers[targetChallengerID]
}

// TriggerManualTest manually triggers an attention test for a specific challenger
// This is called by operator when manual intervention is needed
func (at *AttentionTrigger) TriggerManualTest(targetChallengerID string, blockHash common.Hash, stateRoot common.Hash, blockNumber uint64, batchIndex uint64) *TriggerResult {
	// Validate target challenger
	if !at.ShouldTriggerManualTest(targetChallengerID) {
		return &TriggerResult{
			ShouldTrigger:    false,
			TargetChallenger: targetChallengerID,
			BlockHash:        blockHash,
			StateRoot:        stateRoot,
			TriggerTime:      time.Now(),
			Consensus:        false,
			Votes:            0,
			TotalVotes:       0,
		}
	}

	// Create trigger result for manual test
	return &TriggerResult{
		ShouldTrigger:    true,
		TargetChallenger: targetChallengerID,
		BlockHash:        blockHash,
		StateRoot:        stateRoot,
		TriggerTime:      time.Now(),
		Consensus:        true, // Manual triggers are always considered consensus
		Votes:            1,
		TotalVotes:       1,
	}
}

// SelectRandomChallenger selects a random challenger based on deterministic randomness
func (at *AttentionTrigger) SelectRandomChallenger(blockHash common.Hash, stateRoot common.Hash, blockNumber uint64, batchIndex uint64) string {
	at.mu.RLock()
	defer at.mu.RUnlock()

	// Get list of active peers
	activePeers := make([]string, 0, len(at.peers))
	for peerID, active := range at.peers {
		if active {
			activePeers = append(activePeers, peerID)
		}
	}

	if len(activePeers) == 0 {
		return ""
	}

	// Create deterministic seed for selection using batch index
	seed := at.createDeterministicSeed(blockHash, stateRoot, blockNumber, batchIndex)

	// Select challenger based on seed
	index := int(seed % uint64(len(activePeers)))
	return activePeers[index]
}

// UpdatePeerList adds or removes a peer from the available challengers list
func (at *AttentionTrigger) UpdatePeerList(peerID string, active bool) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.peers[peerID] = active
	if active {
		at.logger.Debug("Challenger added to attention trigger", "peerID", peerID)
	} else {
		at.logger.Debug("Challenger removed from attention trigger", "peerID", peerID)
	}
}

// GetAvailableChallengerCount returns the number of available challengers
func (at *AttentionTrigger) GetAvailableChallengerCount() int {
	at.mu.RLock()
	defer at.mu.RUnlock()

	count := 0
	for _, active := range at.peers {
		if active {
			count++
		}
	}
	return count
}

// createDeterministicSeed creates a deterministic seed from block hash, state root, challenger ID, block number, and batch index
// Each challenger gets a different seed for the same block and batch, ensuring independent selection
func (at *AttentionTrigger) createDeterministicSeed(blockHash common.Hash, stateRoot common.Hash, blockNumber uint64, batchIndex uint64) uint64 {
	// Combine block hash + state root + challenger ID + block number + batch index
	// This ensures each challenger has different selection for the same block and batch
	data := fmt.Sprintf("%s%s%s%d%d", blockHash.Hex(), stateRoot.Hex(), at.challengerID, blockNumber, batchIndex)

	// Create SHA256 hash
	hash := sha256.Sum256([]byte(data))

	// Convert first 8 bytes to uint64
	var seed uint64
	for i := 0; i < 8; i++ {
		seed = seed<<8 + uint64(hash[i])
	}
	return seed
}

// GetConfig returns the current configuration
func (at *AttentionTrigger) GetConfig() *TriggerConfig {
	return at.config
}

// SetConfig updates the configuration
func (at *AttentionTrigger) SetConfig(config *TriggerConfig) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.config = config
	at.logger = config.Logger
}

// GetChallengerID returns the challenger ID
func (at *AttentionTrigger) GetChallengerID() string {
	return at.challengerID
}

// StartRATMonitoring starts the RAT system using existing challenger's L1 monitoring
// This function integrates with the existing gameMonitor.onNewL1Head() system
func (at *AttentionTrigger) StartRATMonitoring(ctx context.Context, l1RPCEndpoint string) error {
	at.logger.Info("Starting RAT system with existing L1 monitoring", "endpoint", l1RPCEndpoint)

	// RAT system will be triggered from existing gameMonitor.onNewL1Head()
	// No need for separate L2 batch monitoring - reuse existing infrastructure
	at.monitoring = true
	at.monitorCtx = ctx

	at.logger.Info("RAT system integrated with existing L1 head monitoring")
	return nil
}

// StopRATMonitoring stops the RAT system
func (at *AttentionTrigger) StopRATMonitoring() {
	if !at.monitoring {
		return
	}

	at.monitoring = false
	at.logger.Info("Stopped RAT system")
}

// ProcessL1Block is called from existing gameMonitor.onNewL1Head()
// This integrates RAT triggering with existing L1 monitoring
func (at *AttentionTrigger) ProcessL1Block(blockHash common.Hash, stateRoot common.Hash, blockNumber uint64) {
	if !at.monitoring {
		return
	}

	// Use block number as batch index for trigger calculation
	// This is a simplified approach - in real implementation,
	// we would extract actual batch information from the block
	batchIndex := blockNumber

	// Check if attention test should be triggered
	if at.ShouldTriggerAttentionTest(blockHash, stateRoot, blockNumber, batchIndex) {
		// Select target challenger
		targetChallenger := at.SelectRandomChallenger(blockHash, stateRoot, blockNumber, batchIndex)

		if targetChallenger != "" {
			at.logger.Info("Triggering attention test from L1 block",
				"batchIndex", batchIndex,
				"targetChallenger", targetChallenger,
				"blockNumber", blockNumber,
				"blockHash", blockHash.Hex())

			// TODO: Broadcast attention test to P2P network
			// This will be implemented in TODO 1.4: P2P 네트워크 통합
		}
	}
}
