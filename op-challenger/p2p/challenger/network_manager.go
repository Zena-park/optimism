package challenger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
)

// ChallengerNetworkManager manages all challenger-related operations in P2P network
type ChallengerNetworkManager struct {
	// Basic information
	nodeID string           // Current node ID
	node   *network.P2PNode // P2P node reference

	// Challenger management
	challengers map[string]*types.ChallengerInfo // Known challengers
	localInfo   *types.ChallengerInfo            // Local challenger information

	// Service management
	discovery *ChallengerDiscovery // Challenger discovery service
	registry  *ChallengerRegistry  // Challenger registration service
	monitor   *ChallengerMonitor   // Challenger monitoring (basic Phase 1 features)

	// Synchronization and state management
	mu     sync.RWMutex       // Concurrency control
	ctx    context.Context    // Context
	cancel context.CancelFunc // Cancel function

	// Configuration
	config *types.ChallengerNetworkManagerConfig // Configuration
	logger log.Logger                            // Logger

	// Phase 1: Basic state tracking
	isRunning       bool         // Whether the manager is running
	lastHeartbeat   time.Time    // Last heartbeat time
	heartbeatTicker *time.Ticker // Heartbeat ticker
}

// NewChallengerNetworkManager creates a new challenger network manager
func NewChallengerNetworkManager(node *network.P2PNode, config *types.ChallengerNetworkManagerConfig, logger log.Logger) *ChallengerNetworkManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &ChallengerNetworkManager{
		nodeID:      node.GetID(),
		node:        node,
		challengers: make(map[string]*types.ChallengerInfo),
		mu:          sync.RWMutex{},
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
		logger:      logger,
		isRunning:   false,
	}

	// Initialize services
	manager.discovery = NewChallengerDiscovery(manager, logger)
	manager.registry = NewChallengerRegistry(manager, logger)
	manager.monitor = NewChallengerMonitor(manager, logger)

	return manager
}

// Start starts the challenger network manager
func (cnm *ChallengerNetworkManager) Start() error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	if cnm.isRunning {
		return fmt.Errorf("challenger network manager is already running")
	}

	cnm.logger.Info("Starting challenger network manager", "node_id", cnm.nodeID)

	// Start services
	if err := cnm.discovery.Start(); err != nil {
		return fmt.Errorf("failed to start challenger discovery: %w", err)
	}

	if err := cnm.registry.Start(); err != nil {
		return fmt.Errorf("failed to start challenger registry: %w", err)
	}

	if err := cnm.monitor.Start(); err != nil {
		return fmt.Errorf("failed to start challenger monitor: %w", err)
	}

	// Start heartbeat
	cnm.heartbeatTicker = time.NewTicker(cnm.config.HeartbeatInterval)
	go cnm.heartbeatLoop()

	cnm.isRunning = true
	cnm.lastHeartbeat = time.Now()

	cnm.logger.Info("Challenger network manager started successfully")
	return nil
}

// Stop stops the challenger network manager
func (cnm *ChallengerNetworkManager) Stop() error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	if !cnm.isRunning {
		return nil
	}

	cnm.logger.Info("Stopping challenger network manager")

	// Stop heartbeat
	if cnm.heartbeatTicker != nil {
		cnm.heartbeatTicker.Stop()
	}

	// Stop services
	cnm.monitor.Stop()
	cnm.registry.Stop()
	cnm.discovery.Stop()

	// Cancel context
	cnm.cancel()

	cnm.isRunning = false
	cnm.logger.Info("Challenger network manager stopped")
	return nil
}

// RegisterAsChallenger registers this node as a challenger
func (cnm *ChallengerNetworkManager) RegisterAsChallenger(challengerID string, stakeAmount uint64) error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	if cnm.localInfo != nil {
		return fmt.Errorf("node is already registered as a challenger")
	}

	// Validate challenger ID
	if err := utils.ValidateChallengerID(challengerID); err != nil {
		return fmt.Errorf("invalid challenger ID: %w", err)
	}

	// Check minimum stake amount
	if stakeAmount < cnm.config.MinStakeAmount {
		return fmt.Errorf("stake amount %d is below minimum %d", stakeAmount, cnm.config.MinStakeAmount)
	}

	// Enable challenger functionality in the P2P node
	if err := cnm.node.EnableChallenger(challengerID); err != nil {
		return fmt.Errorf("failed to enable challenger functionality: %w", err)
	}

	// Create local challenger info
	cnm.localInfo = types.NewChallengerInfo(challengerID, cnm.nodeID, cnm.node.GetAddress(), cnm.node.GetPublicKey())
	cnm.localInfo.StakeAmount = stakeAmount
	cnm.localInfo.StakeStatus = types.StakeStatusActive
	cnm.localInfo.UpdateStatus(types.ChallengerStatusOnline)

	// Register with the registry service
	if err := cnm.registry.RegisterChallenger(cnm.localInfo); err != nil {
		cnm.localInfo = nil
		cnm.node.DisableChallenger()
		return fmt.Errorf("failed to register challenger: %w", err)
	}

	cnm.logger.Info("Successfully registered as challenger",
		"challenger_id", challengerID, "stake_amount", stakeAmount)
	return nil
}

// UnregisterAsChallenger unregisters this node as a challenger
func (cnm *ChallengerNetworkManager) UnregisterAsChallenger() error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	if cnm.localInfo == nil {
		return fmt.Errorf("node is not registered as a challenger")
	}

	challengerID := cnm.localInfo.ID

	// Unregister from the registry service
	if err := cnm.registry.UnregisterChallenger(challengerID); err != nil {
		cnm.logger.Warn("Failed to unregister from registry service", "error", err)
	}

	// Disable challenger functionality in the P2P node
	cnm.node.DisableChallenger()

	// Clear local info
	cnm.localInfo = nil

	cnm.logger.Info("Successfully unregistered as challenger", "challenger_id", challengerID)
	return nil
}

// IsChallenger returns whether this node is registered as a challenger
func (cnm *ChallengerNetworkManager) IsChallenger() bool {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()
	return cnm.localInfo != nil
}

// GetLocalChallengerInfo returns the local challenger information
func (cnm *ChallengerNetworkManager) GetLocalChallengerInfo() *types.ChallengerInfo {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()

	if cnm.localInfo == nil {
		return nil
	}

	// Return a copy to prevent external modification
	info := *cnm.localInfo
	return &info
}

// AddChallenger adds a challenger to the known challengers list
func (cnm *ChallengerNetworkManager) AddChallenger(challengerInfo *types.ChallengerInfo) error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	if challengerInfo == nil {
		return fmt.Errorf("challenger info cannot be nil")
	}

	// Validate challenger info
	if err := utils.ValidateChallengerID(challengerInfo.ID); err != nil {
		return fmt.Errorf("invalid challenger ID: %w", err)
	}

	if err := utils.ValidateNetworkAddress(challengerInfo.Address); err != nil {
		return fmt.Errorf("invalid challenger address: %w", err)
	}

	// Add to known challengers
	cnm.challengers[challengerInfo.ID] = challengerInfo

	// Also add to P2P node's challenger peers
	if cnm.node.IsChallenger() {
		if err := cnm.node.AddChallengerPeer(challengerInfo); err != nil {
			cnm.logger.Warn("Failed to add challenger peer to P2P node", "error", err)
		}
	}

	cnm.logger.Debug("Challenger added",
		"challenger_id", challengerInfo.ID, "address", challengerInfo.Address)
	return nil
}

// RemoveChallenger removes a challenger from the known challengers list
func (cnm *ChallengerNetworkManager) RemoveChallenger(challengerID string) {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	delete(cnm.challengers, challengerID)

	// Also remove from P2P node's challenger peers
	if cnm.node.IsChallenger() {
		cnm.node.RemoveChallengerPeer(challengerID)
	}

	cnm.logger.Debug("Challenger removed", "challenger_id", challengerID)
}

// GetChallenger returns information about a specific challenger
func (cnm *ChallengerNetworkManager) GetChallenger(challengerID string) *types.ChallengerInfo {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()

	info, exists := cnm.challengers[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	infoCopy := *info
	return &infoCopy
}

// GetAllChallengers returns all known challengers
func (cnm *ChallengerNetworkManager) GetAllChallengers() map[string]*types.ChallengerInfo {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()

	// Return a copy to prevent external modification
	challengers := make(map[string]*types.ChallengerInfo)
	for id, info := range cnm.challengers {
		infoCopy := *info
		challengers[id] = &infoCopy
	}

	return challengers
}

// GetChallengerCount returns the number of known challengers
func (cnm *ChallengerNetworkManager) GetChallengerCount() int {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()
	return len(cnm.challengers)
}

// GetOnlineChallengerCount returns the number of online challengers
func (cnm *ChallengerNetworkManager) GetOnlineChallengerCount() int {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()

	count := 0
	for _, info := range cnm.challengers {
		if info.IsOnline() {
			count++
		}
	}

	return count
}

// UpdateChallengerStatus updates the status of a challenger
func (cnm *ChallengerNetworkManager) UpdateChallengerStatus(challengerID string, status types.ChallengerStatus) error {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	info, exists := cnm.challengers[challengerID]
	if !exists {
		return fmt.Errorf("challenger not found: %s", challengerID)
	}

	info.UpdateStatus(status)
	cnm.logger.Debug("Challenger status updated",
		"challenger_id", challengerID, "status", status.String())
	return nil
}

// CleanupStaleChallengers removes stale challengers
func (cnm *ChallengerNetworkManager) CleanupStaleChallengers(staleDuration time.Duration) int {
	cnm.mu.Lock()
	defer cnm.mu.Unlock()

	removed := 0
	for id, info := range cnm.challengers {
		if info.IsStale(staleDuration) {
			delete(cnm.challengers, id)

			// Also remove from P2P node's challenger peers
			if cnm.node.IsChallenger() {
				cnm.node.RemoveChallengerPeer(id)
			}

			removed++
			cnm.logger.Debug("Stale challenger removed", "challenger_id", id)
		}
	}

	if removed > 0 {
		cnm.logger.Info("Stale challengers cleaned up", "removed_count", removed)
	}

	return removed
}

// GetP2PNode returns the underlying P2P node
func (cnm *ChallengerNetworkManager) GetP2PNode() *network.P2PNode {
	return cnm.node
}

// GetDiscoveryService returns the challenger discovery service
func (cnm *ChallengerNetworkManager) GetDiscoveryService() *ChallengerDiscovery {
	return cnm.discovery
}

// GetRegistryService returns the challenger registry service
func (cnm *ChallengerNetworkManager) GetRegistryService() *ChallengerRegistry {
	return cnm.registry
}

// GetMonitorService returns the challenger monitor service
func (cnm *ChallengerNetworkManager) GetMonitorService() *ChallengerMonitor {
	return cnm.monitor
}

// heartbeatLoop runs the heartbeat loop
func (cnm *ChallengerNetworkManager) heartbeatLoop() {
	defer func() {
		if r := recover(); r != nil {
			cnm.logger.Error("Heartbeat loop panic recovered", "error", r)
		}
	}()

	for {
		select {
		case <-cnm.ctx.Done():
			return
		case <-cnm.heartbeatTicker.C:
			cnm.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat message to peers
func (cnm *ChallengerNetworkManager) sendHeartbeat() {
	cnm.mu.RLock()
	isChallenger := cnm.localInfo != nil
	cnm.mu.RUnlock()

	if !isChallenger {
		return
	}

	cnm.mu.Lock()
	cnm.lastHeartbeat = time.Now()

	// Update local challenger info
	if cnm.localInfo != nil {
		cnm.localInfo.LastSeen = time.Now()
		cnm.localInfo.PeerCount = cnm.node.GetChallengerPeerCount()
	}
	cnm.mu.Unlock()

	// Phase 1: Basic heartbeat - just update timestamp
	// In Phase 2-3, this could include sending actual heartbeat messages to peers
	cnm.logger.Debug("Heartbeat sent", "timestamp", time.Now())
}

// IsRunning returns whether the manager is running
func (cnm *ChallengerNetworkManager) IsRunning() bool {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()
	return cnm.isRunning
}

// GetLastHeartbeat returns the last heartbeat time
func (cnm *ChallengerNetworkManager) GetLastHeartbeat() time.Time {
	cnm.mu.RLock()
	defer cnm.mu.RUnlock()
	return cnm.lastHeartbeat
}
