package challenger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
)

// ChallengerRegistry handles challenger registration and validation
type ChallengerRegistry struct {
	manager *ChallengerNetworkManager // Reference to parent manager
	logger  log.Logger                // Logger

	// Registry state
	isRunning bool               // Whether registry is running
	ctx       context.Context    // Context
	cancel    context.CancelFunc // Cancel function

	// Registration data
	registeredChallengers map[string]*registrationEntry // Registered challengers
	mu                    sync.RWMutex                  // Concurrency control

	// Phase 1: Basic registration tracking
	registrationCount int       // Number of registrations processed
	lastRegistration  time.Time // Last registration time
}

// registrationEntry represents a challenger registration entry
type registrationEntry struct {
	ChallengerInfo  *types.ChallengerInfo `json:"challenger_info"`
	RegisteredAt    time.Time             `json:"registered_at"`
	LastValidated   time.Time             `json:"last_validated"`
	ValidationCount int                   `json:"validation_count"`
	IsValid         bool                  `json:"is_valid"`
}

// NewChallengerRegistry creates a new challenger registry service
func NewChallengerRegistry(manager *ChallengerNetworkManager, logger log.Logger) *ChallengerRegistry {
	ctx, cancel := context.WithCancel(context.Background())

	return &ChallengerRegistry{
		manager:               manager,
		logger:                logger,
		isRunning:             false,
		registeredChallengers: make(map[string]*registrationEntry),
		ctx:                   ctx,
		cancel:                cancel,
	}
}

// Start starts the challenger registry service
func (cr *ChallengerRegistry) Start() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.isRunning {
		return fmt.Errorf("challenger registry is already running")
	}

	cr.logger.Info("Starting challenger registry service")

	// Phase 1: Basic registry - no background validation yet
	// In Phase 2-3, we could add periodic validation, cleanup, etc.

	cr.isRunning = true
	cr.logger.Info("Challenger registry service started")
	return nil
}

// Stop stops the challenger registry service
func (cr *ChallengerRegistry) Stop() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.isRunning {
		return
	}

	cr.logger.Info("Stopping challenger registry service")

	// Cancel context
	cr.cancel()

	cr.isRunning = false
	cr.logger.Info("Challenger registry service stopped")
}

// RegisterChallenger registers a challenger
func (cr *ChallengerRegistry) RegisterChallenger(challengerInfo *types.ChallengerInfo) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.isRunning {
		return fmt.Errorf("challenger registry is not running")
	}

	if challengerInfo == nil {
		return fmt.Errorf("challenger info cannot be nil")
	}

	// Validate challenger info
	if err := cr.validateChallengerInfo(challengerInfo); err != nil {
		return fmt.Errorf("challenger validation failed: %w", err)
	}

	// Check if already registered
	if _, exists := cr.registeredChallengers[challengerInfo.ID]; exists {
		return fmt.Errorf("challenger already registered: %s", challengerInfo.ID)
	}

	// Create registration entry
	entry := &registrationEntry{
		ChallengerInfo:  challengerInfo,
		RegisteredAt:    time.Now(),
		LastValidated:   time.Now(),
		ValidationCount: 1,
		IsValid:         true,
	}

	// Add to registry
	cr.registeredChallengers[challengerInfo.ID] = entry
	cr.registrationCount++
	cr.lastRegistration = time.Now()

	cr.logger.Info("Challenger registered",
		"challenger_id", challengerInfo.ID,
		"address", challengerInfo.Address,
		"stake_amount", challengerInfo.StakeAmount)
	return nil
}

// UnregisterChallenger unregisters a challenger
func (cr *ChallengerRegistry) UnregisterChallenger(challengerID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.isRunning {
		return fmt.Errorf("challenger registry is not running")
	}

	// Check if registered
	entry, exists := cr.registeredChallengers[challengerID]
	if !exists {
		return fmt.Errorf("challenger not registered: %s", challengerID)
	}

	// Remove from registry
	delete(cr.registeredChallengers, challengerID)

	cr.logger.Info("Challenger unregistered",
		"challenger_id", challengerID,
		"was_registered_at", entry.RegisteredAt)
	return nil
}

// IsRegistered checks if a challenger is registered
func (cr *ChallengerRegistry) IsRegistered(challengerID string) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	entry, exists := cr.registeredChallengers[challengerID]
	return exists && entry.IsValid
}

// GetRegisteredChallenger returns information about a registered challenger
func (cr *ChallengerRegistry) GetRegisteredChallenger(challengerID string) *types.ChallengerInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	entry, exists := cr.registeredChallengers[challengerID]
	if !exists || !entry.IsValid {
		return nil
	}

	// Return a copy to prevent external modification
	info := *entry.ChallengerInfo
	return &info
}

// GetAllRegisteredChallengers returns all registered challengers
func (cr *ChallengerRegistry) GetAllRegisteredChallengers() map[string]*types.ChallengerInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	challengers := make(map[string]*types.ChallengerInfo)
	for id, entry := range cr.registeredChallengers {
		if entry.IsValid {
			// Return a copy to prevent external modification
			info := *entry.ChallengerInfo
			challengers[id] = &info
		}
	}

	return challengers
}

// GetRegisteredChallengersByRole returns registered challengers with specific roles
func (cr *ChallengerRegistry) GetRegisteredChallengersByRole(role types.ChallengerRole) map[string]*types.ChallengerInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	challengers := make(map[string]*types.ChallengerInfo)
	for id, entry := range cr.registeredChallengers {
		if entry.IsValid && entry.ChallengerInfo.HasRole(role) {
			// Return a copy to prevent external modification
			info := *entry.ChallengerInfo
			challengers[id] = &info
		}
	}

	return challengers
}

// UpdateChallengerInfo updates challenger information
func (cr *ChallengerRegistry) UpdateChallengerInfo(challengerID string, challengerInfo *types.ChallengerInfo) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.isRunning {
		return fmt.Errorf("challenger registry is not running")
	}

	// Check if registered
	entry, exists := cr.registeredChallengers[challengerID]
	if !exists {
		return fmt.Errorf("challenger not registered: %s", challengerID)
	}

	// Validate new challenger info
	if err := cr.validateChallengerInfo(challengerInfo); err != nil {
		return fmt.Errorf("challenger validation failed: %w", err)
	}

	// Update entry
	entry.ChallengerInfo = challengerInfo
	entry.LastValidated = time.Now()
	entry.ValidationCount++

	cr.logger.Debug("Challenger info updated", "challenger_id", challengerID)
	return nil
}

// ValidateChallenger validates a challenger's current state
func (cr *ChallengerRegistry) ValidateChallenger(challengerID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.isRunning {
		return fmt.Errorf("challenger registry is not running")
	}

	// Check if registered
	entry, exists := cr.registeredChallengers[challengerID]
	if !exists {
		return fmt.Errorf("challenger not registered: %s", challengerID)
	}

	// Validate challenger info
	if err := cr.validateChallengerInfo(entry.ChallengerInfo); err != nil {
		entry.IsValid = false
		cr.logger.Warn("Challenger validation failed", "challenger_id", challengerID, "error", err)
		return err
	}

	// Update validation info
	entry.LastValidated = time.Now()
	entry.ValidationCount++
	entry.IsValid = true

	cr.logger.Debug("Challenger validated", "challenger_id", challengerID)
	return nil
}

// validateChallengerInfo validates challenger information
func (cr *ChallengerRegistry) validateChallengerInfo(challengerInfo *types.ChallengerInfo) error {
	// Validate challenger ID
	if err := utils.ValidateChallengerID(challengerInfo.ID); err != nil {
		return fmt.Errorf("invalid challenger ID: %w", err)
	}

	// Validate node ID
	if err := utils.ValidateNodeID(challengerInfo.NodeID); err != nil {
		return fmt.Errorf("invalid node ID: %w", err)
	}

	// Validate network address
	if err := utils.ValidateNetworkAddress(challengerInfo.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Validate public key
	if challengerInfo.PublicKey == nil {
		return fmt.Errorf("public key cannot be nil")
	}

	// Validate stake amount (Phase 1: basic check)
	if challengerInfo.StakeAmount < cr.manager.config.MinStakeAmount {
		return fmt.Errorf("stake amount %d is below minimum %d",
			challengerInfo.StakeAmount, cr.manager.config.MinStakeAmount)
	}

	// Validate role
	if challengerInfo.Role == 0 {
		return fmt.Errorf("challenger role cannot be empty")
	}

	// Phase 1: Basic validation - more sophisticated validation in Phase 2-3
	return nil
}

// GetRegistrationEntry returns the registration entry for a challenger
func (cr *ChallengerRegistry) GetRegistrationEntry(challengerID string) *registrationEntry {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	entry, exists := cr.registeredChallengers[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	entryCopy := *entry
	infoCopy := *entry.ChallengerInfo
	entryCopy.ChallengerInfo = &infoCopy

	return &entryCopy
}

// GetRegistrationStats returns registration statistics
func (cr *ChallengerRegistry) GetRegistrationStats() map[string]interface{} {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	validCount := 0
	invalidCount := 0

	for _, entry := range cr.registeredChallengers {
		if entry.IsValid {
			validCount++
		} else {
			invalidCount++
		}
	}

	return map[string]interface{}{
		"is_running":          cr.isRunning,
		"total_registered":    len(cr.registeredChallengers),
		"valid_challengers":   validCount,
		"invalid_challengers": invalidCount,
		"registration_count":  cr.registrationCount,
		"last_registration":   cr.lastRegistration,
	}
}

// CleanupInvalidChallengers removes invalid challengers from the registry
func (cr *ChallengerRegistry) CleanupInvalidChallengers() int {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	removed := 0
	for id, entry := range cr.registeredChallengers {
		if !entry.IsValid {
			delete(cr.registeredChallengers, id)
			removed++
			cr.logger.Debug("Invalid challenger removed from registry", "challenger_id", id)
		}
	}

	if removed > 0 {
		cr.logger.Info("Invalid challengers cleaned up from registry", "removed_count", removed)
	}

	return removed
}

// GetRegisteredChallengerCount returns the number of registered challengers
func (cr *ChallengerRegistry) GetRegisteredChallengerCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	validCount := 0
	for _, entry := range cr.registeredChallengers {
		if entry.IsValid {
			validCount++
		}
	}

	return validCount
}

// IsRunning returns whether the registry service is running
func (cr *ChallengerRegistry) IsRunning() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.isRunning
}
