package state

import (
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// StateConflict represents a conflict between two challenger states
type StateConflict struct {
	Type          string                 `json:"type"`           // Type of conflict
	ChallengerID  string                 `json:"challenger_id"`  // ID of the challenger with conflict
	ExistingState *types.ChallengerState `json:"existing_state"` // Current state
	NewState      *types.ChallengerState `json:"new_state"`      // Conflicting state
	DetectedAt    time.Time              `json:"detected_at"`    // When the conflict was detected
	Resolved      bool                   `json:"resolved"`       // Whether the conflict has been resolved
	ResolvedAt    time.Time              `json:"resolved_at"`    // When the conflict was resolved
	Resolution    string                 `json:"resolution"`     // How the conflict was resolved
}

// StateUpdate represents a state update message
type StateUpdate struct {
	ChallengerID string                 `json:"challenger_id"` // ID of the challenger
	State        *types.ChallengerState `json:"state"`         // Updated state
	UpdateType   StateUpdateType        `json:"update_type"`   // Type of update
	Timestamp    time.Time              `json:"timestamp"`     // When the update occurred
	Signature    []byte                 `json:"signature"`     // Digital signature (Phase 2+)
}

// StateUpdateType represents the type of state update
type StateUpdateType int

const (
	StateUpdateTypeStatus       StateUpdateType = iota // Status change (online/offline)
	StateUpdateTypeConnection                          // Connection info change
	StateUpdateTypeHeartbeat                           // Heartbeat update
	StateUpdateTypeRegistration                        // Registration/unregistration
)

// String returns string representation of StateUpdateType
func (sut StateUpdateType) String() string {
	switch sut {
	case StateUpdateTypeStatus:
		return "status"
	case StateUpdateTypeConnection:
		return "connection"
	case StateUpdateTypeHeartbeat:
		return "heartbeat"
	case StateUpdateTypeRegistration:
		return "registration"
	default:
		return "unknown"
	}
}

// SyncRequest represents a state synchronization request
type SyncRequest struct {
	RequestID     string    `json:"request_id"`     // Unique request ID
	RequesterID   string    `json:"requester_id"`   // ID of the requester
	ChallengerIDs []string  `json:"challenger_ids"` // IDs of challengers to sync (empty for all)
	Timestamp     time.Time `json:"timestamp"`      // Request timestamp
}

// SyncResponse represents a state synchronization response
type SyncResponse struct {
	RequestID string                            `json:"request_id"` // Request ID this responds to
	States    map[string]*types.ChallengerState `json:"states"`     // Challenger states
	Timestamp time.Time                         `json:"timestamp"`  // Response timestamp
	Complete  bool                              `json:"complete"`   // Whether this is the complete response
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	ConflictID   string                 `json:"conflict_id"`   // Unique conflict ID
	ChallengerID string                 `json:"challenger_id"` // Challenger with conflict
	Method       string                 `json:"method"`        // Resolution method used
	Result       *types.ChallengerState `json:"result"`        // Final resolved state
	ResolvedAt   time.Time              `json:"resolved_at"`   // When resolved
	ResolvedBy   string                 `json:"resolved_by"`   // Who resolved it
}

// DefaultChallengerStateManagerConfig returns default configuration for state manager
func DefaultChallengerStateManagerConfig() *ChallengerStateManagerConfig {
	return &ChallengerStateManagerConfig{
		SyncInterval:             30 * time.Second, // Sync every 30 seconds
		StateTimeout:             5 * time.Minute,  // States timeout after 5 minutes
		MaxStates:                10000,            // Track up to 10,000 states
		EnableConflictResolution: false,            // Phase 1: Disable advanced conflict resolution
	}
}

// Validate validates the state manager configuration
func (config *ChallengerStateManagerConfig) Validate() error {
	if config.SyncInterval <= 0 {
		return fmt.Errorf("sync interval must be positive")
	}

	if config.StateTimeout <= 0 {
		return fmt.Errorf("state timeout must be positive")
	}

	if config.MaxStates <= 0 {
		return fmt.Errorf("max states must be positive")
	}

	return nil
}
