package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/challenger"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum/go-ethereum/log"
)

// ChallengerStateManager manages the state of challengers in the network (Phase 1: basic functionality)
type ChallengerStateManager struct {
	// Core components
	networkManager *challenger.ChallengerNetworkManager // Reference to network manager
	logger         log.Logger                           // Logger

	// State management
	states     map[string]*types.ChallengerState // Current states of all known challengers
	localState *types.ChallengerState            // Local challenger state
	mu         sync.RWMutex                      // Concurrency control

	// Synchronization
	synchronizer *StateSynchronizer // State synchronization service
	syncTicker   *time.Ticker       // Synchronization ticker
	ctx          context.Context    // Context
	cancel       context.CancelFunc // Cancel function

	// Configuration
	config    *ChallengerStateManagerConfig // Configuration
	isRunning bool                          // Whether the manager is running

	// Phase 1: Basic state tracking
	lastSync      time.Time // Last synchronization time
	syncCount     int       // Number of synchronizations performed
	conflictCount int       // Number of conflicts detected
}

// ChallengerStateManagerConfig contains configuration for state manager
type ChallengerStateManagerConfig struct {
	SyncInterval             time.Duration `json:"sync_interval"`              // Interval between state synchronizations
	StateTimeout             time.Duration `json:"state_timeout"`              // Timeout for stale states
	MaxStates                int           `json:"max_states"`                 // Maximum number of states to track
	EnableConflictResolution bool          `json:"enable_conflict_resolution"` // Enable conflict resolution (Phase 2+)
}

// NewChallengerStateManager creates a new challenger state manager
func NewChallengerStateManager(networkManager *challenger.ChallengerNetworkManager, config *ChallengerStateManagerConfig, logger log.Logger) *ChallengerStateManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &ChallengerStateManager{
		networkManager: networkManager,
		logger:         logger,
		states:         make(map[string]*types.ChallengerState),
		mu:             sync.RWMutex{},
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
		isRunning:      false,
	}

	// Initialize synchronizer
	manager.synchronizer = NewStateSynchronizer(manager, logger)

	return manager
}

// Start starts the state manager
func (csm *ChallengerStateManager) Start() error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if csm.isRunning {
		return fmt.Errorf("challenger state manager is already running")
	}

	csm.logger.Info("Starting challenger state manager")

	// Start synchronizer
	if err := csm.synchronizer.Start(); err != nil {
		return fmt.Errorf("failed to start state synchronizer: %w", err)
	}

	// Initialize local state if this node is a challenger
	if csm.networkManager.IsChallenger() {
		if err := csm.initializeLocalState(); err != nil {
			return fmt.Errorf("failed to initialize local state: %w", err)
		}
	}

	// Start synchronization loop
	csm.syncTicker = time.NewTicker(csm.config.SyncInterval)
	go csm.syncLoop()

	csm.isRunning = true
	csm.lastSync = time.Now()

	csm.logger.Info("Challenger state manager started successfully")
	return nil
}

// Stop stops the state manager
func (csm *ChallengerStateManager) Stop() error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.isRunning {
		return nil
	}

	csm.logger.Info("Stopping challenger state manager")

	// Stop synchronization ticker
	if csm.syncTicker != nil {
		csm.syncTicker.Stop()
	}

	// Stop synchronizer
	csm.synchronizer.Stop()

	// Cancel context
	csm.cancel()

	csm.isRunning = false
	csm.logger.Info("Challenger state manager stopped")
	return nil
}

// initializeLocalState initializes the local challenger state
func (csm *ChallengerStateManager) initializeLocalState() error {
	challengerInfo := csm.networkManager.GetLocalChallengerInfo()
	if challengerInfo == nil {
		return fmt.Errorf("no local challenger info available")
	}

	csm.localState = &types.ChallengerState{
		ChallengerID: challengerInfo.ID,
		NodeID:       challengerInfo.NodeID,
		Address:      challengerInfo.Address,
		Status:       challengerInfo.Status,
		IsOnline:     challengerInfo.IsOnline(),
		IsConnected:  challengerInfo.IsConnected,
		ConnectionInfo: &types.ConnectionInfo{
			ConnectedAt:    challengerInfo.ConnectedAt,
			LastSeen:       challengerInfo.LastSeen,
			PeerCount:      challengerInfo.PeerCount,
			NetworkLatency: challengerInfo.NetworkLatency,
		},
		Version:     "1.0.0",
		LastUpdated: time.Now(),
		Sequence:    1,
	}

	// Add to states map
	csm.states[challengerInfo.ID] = csm.localState

	csm.logger.Info("Local challenger state initialized", "challenger_id", challengerInfo.ID)
	return nil
}

// UpdateLocalState updates the local challenger state
func (csm *ChallengerStateManager) UpdateLocalState(status types.ChallengerStatus) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.isRunning {
		return fmt.Errorf("state manager is not running")
	}

	if csm.localState == nil {
		return fmt.Errorf("no local state available")
	}

	// Update local state
	csm.localState.Status = status
	csm.localState.LastUpdated = time.Now()
	csm.localState.IsOnline = (status == types.ChallengerStatusOnline)
	csm.localState.ConnectionInfo.LastSeen = time.Now()
	csm.localState.Sequence++

	// Update network manager
	if err := csm.networkManager.UpdateChallengerStatus(csm.localState.ChallengerID, status); err != nil {
		csm.logger.Warn("Failed to update network manager status", "error", err)
	}

	csm.logger.Debug("Local challenger state updated",
		"challenger_id", csm.localState.ChallengerID,
		"status", status.String())
	return nil
}

// GetLocalState returns the local challenger state
func (csm *ChallengerStateManager) GetLocalState() *types.ChallengerState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	if csm.localState == nil {
		return nil
	}

	// Return a copy to prevent external modification
	stateCopy := csm.localState.Clone()
	return stateCopy
}

// UpdateChallengerState updates the state of a challenger
func (csm *ChallengerStateManager) UpdateChallengerState(challengerID string, state *types.ChallengerState) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if !csm.isRunning {
		return fmt.Errorf("state manager is not running")
	}

	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	// Validate state
	if err := csm.validateState(state); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	// Check for conflicts (Phase 1: basic conflict detection)
	existingState, exists := csm.states[challengerID]
	if exists {
		if conflict := csm.detectConflict(existingState, state); conflict != nil {
			csm.conflictCount++
			csm.logger.Warn("State conflict detected",
				"challenger_id", challengerID,
				"conflict", conflict.Type)

			// Phase 1: Simple conflict resolution - use newer state
			if state.LastUpdated.After(existingState.LastUpdated) {
				csm.states[challengerID] = state.Clone()
				csm.logger.Debug("Conflict resolved using newer state", "challenger_id", challengerID)
			}
			return nil
		}
	}

	// Update state
	csm.states[challengerID] = state.Clone()

	csm.logger.Debug("Challenger state updated",
		"challenger_id", challengerID,
		"status", state.Status.String())
	return nil
}

// GetChallengerState returns the state of a specific challenger
func (csm *ChallengerStateManager) GetChallengerState(challengerID string) *types.ChallengerState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	state, exists := csm.states[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	return state.Clone()
}

// GetAllStates returns all challenger states
func (csm *ChallengerStateManager) GetAllStates() map[string]*types.ChallengerState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	states := make(map[string]*types.ChallengerState)
	for id, state := range csm.states {
		states[id] = state.Clone()
	}

	return states
}

// GetOnlineChallengers returns a list of online challengers
func (csm *ChallengerStateManager) GetOnlineChallengers() []string {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	var online []string
	for id, state := range csm.states {
		if state.IsOnline {
			online = append(online, id)
		}
	}

	return online
}

// GetOfflineChallengers returns a list of offline challengers
func (csm *ChallengerStateManager) GetOfflineChallengers() []string {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	var offline []string
	for id, state := range csm.states {
		if !state.IsOnline {
			offline = append(offline, id)
		}
	}

	return offline
}

// RemoveChallengerState removes a challenger state
func (csm *ChallengerStateManager) RemoveChallengerState(challengerID string) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	delete(csm.states, challengerID)
	csm.logger.Debug("Challenger state removed", "challenger_id", challengerID)
}

// CleanupStaleStates removes stale challenger states
func (csm *ChallengerStateManager) CleanupStaleStates() int {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	removed := 0
	now := time.Now()

	for id, state := range csm.states {
		// Skip local state
		if csm.localState != nil && id == csm.localState.ChallengerID {
			continue
		}

		// Check if state is stale
		if now.Sub(state.LastUpdated) > csm.config.StateTimeout {
			delete(csm.states, id)
			removed++
			csm.logger.Debug("Stale challenger state removed", "challenger_id", id)
		}
	}

	if removed > 0 {
		csm.logger.Info("Stale challenger states cleaned up", "removed_count", removed)
	}

	return removed
}

// validateState validates a challenger state (Phase 1: basic validation)
func (csm *ChallengerStateManager) validateState(state *types.ChallengerState) error {
	if state.ChallengerID == "" {
		return fmt.Errorf("challenger ID cannot be empty")
	}

	if state.NodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if state.LastUpdated.IsZero() {
		return fmt.Errorf("last update time cannot be zero")
	}

	// Phase 1: Basic validation only
	return nil
}

// detectConflict detects conflicts between states (Phase 1: basic conflict detection)
func (csm *ChallengerStateManager) detectConflict(existing, new *types.ChallengerState) *StateConflict {
	// Phase 1: Simple conflict detection
	if existing.Status != new.Status && existing.LastUpdated.Equal(new.LastUpdated) {
		return &StateConflict{
			Type:          "status_mismatch",
			ChallengerID:  existing.ChallengerID,
			ExistingState: existing,
			NewState:      new,
			DetectedAt:    time.Now(),
		}
	}

	// No conflict detected
	return nil
}

// syncLoop runs the periodic synchronization loop
func (csm *ChallengerStateManager) syncLoop() {
	defer func() {
		if r := recover(); r != nil {
			csm.logger.Error("State sync loop panic recovered", "error", r)
		}
	}()

	for {
		select {
		case <-csm.ctx.Done():
			return
		case <-csm.syncTicker.C:
			csm.performSync()
		}
	}
}

// performSync performs periodic state synchronization
func (csm *ChallengerStateManager) performSync() {
	csm.mu.Lock()
	csm.lastSync = time.Now()
	csm.syncCount++
	csm.mu.Unlock()

	// Phase 1: Basic synchronization
	// 1. Cleanup stale states
	removed := csm.CleanupStaleStates()

	// 2. Sync with network manager
	csm.syncWithNetworkManager()

	// 3. Trigger synchronizer
	if err := csm.synchronizer.SynchronizeStates(); err != nil {
		csm.logger.Warn("State synchronization failed", "error", err)
	}

	csm.logger.Debug("State synchronization completed",
		"sync_count", csm.syncCount,
		"removed_stale", removed)
}

// syncWithNetworkManager synchronizes states with network manager
func (csm *ChallengerStateManager) syncWithNetworkManager() {
	// Get all known challengers from network manager
	challengers := csm.networkManager.GetAllChallengers()

	for _, challengerInfo := range challengers {
		// Check if we have state for this challenger
		existingState := csm.GetChallengerState(challengerInfo.ID)
		if existingState == nil {
			// Create new state from challenger info
			newState := &types.ChallengerState{
				ChallengerID: challengerInfo.ID,
				NodeID:       challengerInfo.NodeID,
				Address:      challengerInfo.Address,
				Status:       challengerInfo.Status,
				IsOnline:     challengerInfo.IsOnline(),
				IsConnected:  challengerInfo.IsConnected,
				ConnectionInfo: &types.ConnectionInfo{
					ConnectedAt:    challengerInfo.ConnectedAt,
					LastSeen:       challengerInfo.LastSeen,
					PeerCount:      challengerInfo.PeerCount,
					NetworkLatency: challengerInfo.NetworkLatency,
				},
				Version:     "1.0.0",
				LastUpdated: time.Now(),
				Sequence:    1,
			}

			if err := csm.UpdateChallengerState(challengerInfo.ID, newState); err != nil {
				csm.logger.Debug("Failed to update challenger state from network manager",
					"challenger_id", challengerInfo.ID, "error", err)
			}
		}
	}
}

// GetStateStats returns state management statistics
func (csm *ChallengerStateManager) GetStateStats() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	onlineCount := 0
	offlineCount := 0

	for _, state := range csm.states {
		if state.IsOnline {
			onlineCount++
		} else {
			offlineCount++
		}
	}

	return map[string]interface{}{
		"is_running":      csm.isRunning,
		"total_states":    len(csm.states),
		"online_count":    onlineCount,
		"offline_count":   offlineCount,
		"sync_count":      csm.syncCount,
		"conflict_count":  csm.conflictCount,
		"last_sync":       csm.lastSync,
		"has_local_state": csm.localState != nil,
	}
}

// IsRunning returns whether the state manager is running
func (csm *ChallengerStateManager) IsRunning() bool {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.isRunning
}

// GetNetworkManager returns the network manager reference
func (csm *ChallengerStateManager) GetNetworkManager() *challenger.ChallengerNetworkManager {
	return csm.networkManager
}

// GetSynchronizer returns the state synchronizer
func (csm *ChallengerStateManager) GetSynchronizer() *StateSynchronizer {
	return csm.synchronizer
}
