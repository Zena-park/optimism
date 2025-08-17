package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
)

// StateSynchronizer handles state synchronization between challenger nodes (Phase 1: basic functionality)
type StateSynchronizer struct {
	stateManager *ChallengerStateManager // Reference to state manager
	logger       log.Logger              // Logger

	// Synchronization state
	isRunning bool               // Whether synchronizer is running
	ctx       context.Context    // Context
	cancel    context.CancelFunc // Cancel function

	// Sync requests and responses
	pendingRequests map[string]*syncRequestEntry // Pending sync requests
	mu              sync.RWMutex                 // Concurrency control

	// Phase 1: Basic sync tracking
	lastSyncTime time.Time // Last synchronization time
	syncCount    int       // Number of synchronizations performed
	errorCount   int       // Number of sync errors
}

// syncRequestEntry represents a pending synchronization request
type syncRequestEntry struct {
	Request      *SyncRequest                           `json:"request"`
	ResponseChan chan map[string]*types.ChallengerState `json:"-"`
	CreatedAt    time.Time                              `json:"created_at"`
	Timeout      time.Duration                          `json:"timeout"`
}

// NewStateSynchronizer creates a new state synchronizer
func NewStateSynchronizer(stateManager *ChallengerStateManager, logger log.Logger) *StateSynchronizer {
	ctx, cancel := context.WithCancel(context.Background())

	return &StateSynchronizer{
		stateManager:    stateManager,
		logger:          logger,
		isRunning:       false,
		ctx:             ctx,
		cancel:          cancel,
		pendingRequests: make(map[string]*syncRequestEntry),
		mu:              sync.RWMutex{},
	}
}

// Start starts the state synchronizer
func (ss *StateSynchronizer) Start() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.isRunning {
		return fmt.Errorf("state synchronizer is already running")
	}

	ss.logger.Info("Starting state synchronizer")

	// Phase 1: Basic synchronizer - no background tasks yet
	// In Phase 2-3, we could add periodic sync requests, conflict resolution, etc.

	ss.isRunning = true
	ss.lastSyncTime = time.Now()

	ss.logger.Info("State synchronizer started")
	return nil
}

// Stop stops the state synchronizer
func (ss *StateSynchronizer) Stop() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if !ss.isRunning {
		return
	}

	ss.logger.Info("Stopping state synchronizer")

	// Cancel context
	ss.cancel()

	// Cancel pending requests
	for _, entry := range ss.pendingRequests {
		close(entry.ResponseChan)
	}
	ss.pendingRequests = make(map[string]*syncRequestEntry)

	ss.isRunning = false
	ss.logger.Info("State synchronizer stopped")
}

// SynchronizeStates performs state synchronization with other nodes
func (ss *StateSynchronizer) SynchronizeStates() error {
	ss.mu.RLock()
	if !ss.isRunning {
		ss.mu.RUnlock()
		return fmt.Errorf("state synchronizer is not running")
	}
	ss.mu.RUnlock()

	ss.mu.Lock()
	ss.lastSyncTime = time.Now()
	ss.syncCount++
	ss.mu.Unlock()

	// Phase 1: Basic synchronization - sync with network manager only
	// In Phase 2-3, this would include P2P sync requests to other nodes

	networkManager := ss.stateManager.GetNetworkManager()
	if networkManager == nil {
		return fmt.Errorf("network manager not available")
	}

	// Get all known challengers from network manager
	challengers := networkManager.GetAllChallengers()
	syncedCount := 0

	for _, challengerInfo := range challengers {
		// Create or update state from challenger info
		existingState := ss.stateManager.GetChallengerState(challengerInfo.ID)

		// Create updated state
		updatedState := &types.ChallengerState{
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

		// Update state if it's new or different
		if existingState == nil || ss.shouldUpdateState(existingState, updatedState) {
			if err := ss.stateManager.UpdateChallengerState(challengerInfo.ID, updatedState); err != nil {
				ss.logger.Debug("Failed to update challenger state during sync",
					"challenger_id", challengerInfo.ID, "error", err)
				ss.mu.Lock()
				ss.errorCount++
				ss.mu.Unlock()
			} else {
				syncedCount++
			}
		}
	}

	ss.logger.Debug("State synchronization completed",
		"synced_count", syncedCount,
		"total_challengers", len(challengers))
	return nil
}

// shouldUpdateState determines if a state should be updated
func (ss *StateSynchronizer) shouldUpdateState(existing, new *types.ChallengerState) bool {
	// Phase 1: Simple update logic

	// Update if status changed
	if existing.Status != new.Status {
		return true
	}

	// Update if online status changed
	if existing.IsOnline != new.IsOnline {
		return true
	}

	// Update if connection info changed significantly
	if existing.ConnectionInfo.PeerCount != new.ConnectionInfo.PeerCount {
		return true
	}

	// Update if it's been a while since last update (prevent stale data)
	if time.Since(existing.LastUpdated) > 2*time.Minute {
		return true
	}

	return false
}

// RequestStateSync requests state synchronization for specific challengers
func (ss *StateSynchronizer) RequestStateSync(challengerIDs []string) (map[string]*types.ChallengerState, error) {
	ss.mu.RLock()
	if !ss.isRunning {
		ss.mu.RUnlock()
		return nil, fmt.Errorf("state synchronizer is not running")
	}
	ss.mu.RUnlock()

	// Generate request ID
	requestID := utils.GenerateRequestID()

	// Create sync request
	request := &SyncRequest{
		RequestID:     requestID,
		RequesterID:   ss.getRequesterID(),
		ChallengerIDs: challengerIDs,
		Timestamp:     time.Now(),
	}

	// Create request entry
	entry := &syncRequestEntry{
		Request:      request,
		ResponseChan: make(chan map[string]*types.ChallengerState, 1),
		CreatedAt:    time.Now(),
		Timeout:      10 * time.Second, // 10 second timeout
	}

	// Add to pending requests
	ss.mu.Lock()
	ss.pendingRequests[requestID] = entry
	ss.mu.Unlock()

	// Phase 1: Immediate local sync (no P2P requests yet)
	go ss.processLocalSyncRequest(entry)

	// Wait for response with timeout
	select {
	case result := <-entry.ResponseChan:
		ss.mu.Lock()
		delete(ss.pendingRequests, requestID)
		ss.mu.Unlock()
		return result, nil
	case <-time.After(entry.Timeout):
		ss.mu.Lock()
		delete(ss.pendingRequests, requestID)
		ss.mu.Unlock()
		close(entry.ResponseChan)
		return nil, fmt.Errorf("sync request timed out")
	}
}

// processLocalSyncRequest processes a sync request locally (Phase 1)
func (ss *StateSynchronizer) processLocalSyncRequest(entry *syncRequestEntry) {
	defer func() {
		if r := recover(); r != nil {
			ss.logger.Error("Local sync request panic recovered", "error", r)
			close(entry.ResponseChan)
		}
	}()

	request := entry.Request
	result := make(map[string]*types.ChallengerState)

	if len(request.ChallengerIDs) == 0 {
		// Return all states
		result = ss.stateManager.GetAllStates()
	} else {
		// Return specific states
		for _, challengerID := range request.ChallengerIDs {
			if state := ss.stateManager.GetChallengerState(challengerID); state != nil {
				result[challengerID] = state
			}
		}
	}

	// Send response
	select {
	case entry.ResponseChan <- result:
	case <-ss.ctx.Done():
	}
}

// getRequesterID returns the ID of the requester (local challenger)
func (ss *StateSynchronizer) getRequesterID() string {
	localState := ss.stateManager.GetLocalState()
	if localState != nil {
		return localState.ChallengerID
	}

	// Fallback: use network manager node ID
	networkManager := ss.stateManager.GetNetworkManager()
	if networkManager != nil {
		if challengerInfo := networkManager.GetLocalChallengerInfo(); challengerInfo != nil {
			return challengerInfo.ID
		}
	}

	return "unknown"
}

// ProcessStateUpdate processes an incoming state update
func (ss *StateSynchronizer) ProcessStateUpdate(update *StateUpdate) error {
	ss.mu.RLock()
	if !ss.isRunning {
		ss.mu.RUnlock()
		return fmt.Errorf("state synchronizer is not running")
	}
	ss.mu.RUnlock()

	if update == nil {
		return fmt.Errorf("state update cannot be nil")
	}

	// Validate update
	if err := ss.validateStateUpdate(update); err != nil {
		return fmt.Errorf("invalid state update: %w", err)
	}

	// Apply update to state manager
	if err := ss.stateManager.UpdateChallengerState(update.ChallengerID, update.State); err != nil {
		return fmt.Errorf("failed to apply state update: %w", err)
	}

	ss.logger.Debug("State update processed",
		"challenger_id", update.ChallengerID,
		"update_type", update.UpdateType.String())
	return nil
}

// validateStateUpdate validates a state update (Phase 1: basic validation)
func (ss *StateSynchronizer) validateStateUpdate(update *StateUpdate) error {
	if update.ChallengerID == "" {
		return fmt.Errorf("challenger ID cannot be empty")
	}

	if update.State == nil {
		return fmt.Errorf("state cannot be nil")
	}

	if update.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	// Check if update is too old (prevent replay attacks)
	if time.Since(update.Timestamp) > 5*time.Minute {
		return fmt.Errorf("update is too old")
	}

	// Phase 1: Basic validation only
	// In Phase 2-3, we would add signature verification, etc.
	return nil
}

// CreateStateUpdate creates a state update from current local state
func (ss *StateSynchronizer) CreateStateUpdate(updateType StateUpdateType) (*StateUpdate, error) {
	localState := ss.stateManager.GetLocalState()
	if localState == nil {
		return nil, fmt.Errorf("no local state available")
	}

	update := &StateUpdate{
		ChallengerID: localState.ChallengerID,
		State:        localState,
		UpdateType:   updateType,
		Timestamp:    time.Now(),
		// Signature: nil, // Phase 1: No signature yet
	}

	return update, nil
}

// GetSyncStats returns synchronization statistics
func (ss *StateSynchronizer) GetSyncStats() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return map[string]interface{}{
		"is_running":       ss.isRunning,
		"sync_count":       ss.syncCount,
		"error_count":      ss.errorCount,
		"last_sync_time":   ss.lastSyncTime,
		"pending_requests": len(ss.pendingRequests),
	}
}

// CleanupStaleRequests removes stale sync requests
func (ss *StateSynchronizer) CleanupStaleRequests() int {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	removed := 0
	now := time.Now()

	for requestID, entry := range ss.pendingRequests {
		if now.Sub(entry.CreatedAt) > entry.Timeout {
			close(entry.ResponseChan)
			delete(ss.pendingRequests, requestID)
			removed++
		}
	}

	if removed > 0 {
		ss.logger.Debug("Stale sync requests cleaned up", "removed_count", removed)
	}

	return removed
}

// IsRunning returns whether the synchronizer is running
func (ss *StateSynchronizer) IsRunning() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.isRunning
}
