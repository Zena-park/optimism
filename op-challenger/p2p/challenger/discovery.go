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

// ChallengerDiscovery handles challenger node discovery
type ChallengerDiscovery struct {
	manager *ChallengerNetworkManager // Reference to parent manager
	logger  log.Logger                // Logger

	// Discovery state
	isRunning       bool               // Whether discovery is running
	discoveryTicker *time.Ticker       // Discovery ticker
	ctx             context.Context    // Context
	cancel          context.CancelFunc // Cancel function

	// Discovery data
	bootstrapNodes  []string                     // Bootstrap challenger nodes
	pendingRequests map[string]*discoveryRequest // Pending discovery requests
	mu              sync.RWMutex                 // Concurrency control

	// Phase 1: Basic discovery tracking
	lastDiscovery  time.Time // Last discovery time
	discoveryCount int       // Number of discoveries performed
}

// discoveryRequest represents a pending discovery request
type discoveryRequest struct {
	RequestID     string                       `json:"request_id"`
	ChallengerID  string                       `json:"challenger_id"`
	MaxPeers      int                          `json:"max_peers"`
	RequiredRoles types.ChallengerRole         `json:"required_roles"`
	Timestamp     time.Time                    `json:"timestamp"`
	ResponseChan  chan []*types.ChallengerInfo `json:"-"`
}

// NewChallengerDiscovery creates a new challenger discovery service
func NewChallengerDiscovery(manager *ChallengerNetworkManager, logger log.Logger) *ChallengerDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	return &ChallengerDiscovery{
		manager:         manager,
		logger:          logger,
		isRunning:       false,
		pendingRequests: make(map[string]*discoveryRequest),
		ctx:             ctx,
		cancel:          cancel,
		bootstrapNodes:  manager.config.BootstrapNodes,
	}
}

// Start starts the challenger discovery service
func (cd *ChallengerDiscovery) Start() error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if cd.isRunning {
		return fmt.Errorf("challenger discovery is already running")
	}

	cd.logger.Info("Starting challenger discovery service")

	// Start discovery loop
	cd.discoveryTicker = time.NewTicker(cd.manager.config.DiscoveryInterval)
	go cd.discoveryLoop()

	cd.isRunning = true
	cd.lastDiscovery = time.Now()

	cd.logger.Info("Challenger discovery service started")
	return nil
}

// Stop stops the challenger discovery service
func (cd *ChallengerDiscovery) Stop() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if !cd.isRunning {
		return
	}

	cd.logger.Info("Stopping challenger discovery service")

	// Stop discovery ticker
	if cd.discoveryTicker != nil {
		cd.discoveryTicker.Stop()
	}

	// Cancel context
	cd.cancel()

	// Cancel pending requests
	for _, req := range cd.pendingRequests {
		close(req.ResponseChan)
	}
	cd.pendingRequests = make(map[string]*discoveryRequest)

	cd.isRunning = false
	cd.logger.Info("Challenger discovery service stopped")
}

// DiscoverChallengers discovers challenger nodes
func (cd *ChallengerDiscovery) DiscoverChallengers(maxPeers int, requiredRoles types.ChallengerRole) ([]*types.ChallengerInfo, error) {
	cd.mu.RLock()
	if !cd.isRunning {
		cd.mu.RUnlock()
		return nil, fmt.Errorf("challenger discovery is not running")
	}
	cd.mu.RUnlock()

	// Generate request ID
	requestID := utils.GenerateRequestID()

	// Create discovery request
	request := &discoveryRequest{
		RequestID:     requestID,
		ChallengerID:  cd.manager.nodeID,
		MaxPeers:      maxPeers,
		RequiredRoles: requiredRoles,
		Timestamp:     time.Now(),
		ResponseChan:  make(chan []*types.ChallengerInfo, 1),
	}

	// Add to pending requests
	cd.mu.Lock()
	cd.pendingRequests[requestID] = request
	cd.mu.Unlock()

	// Phase 1: Simple discovery from known nodes
	go cd.performDiscovery(request)

	// Wait for response with timeout
	select {
	case result := <-request.ResponseChan:
		cd.mu.Lock()
		delete(cd.pendingRequests, requestID)
		cd.mu.Unlock()
		return result, nil
	case <-time.After(10 * time.Second): // 10 second timeout
		cd.mu.Lock()
		delete(cd.pendingRequests, requestID)
		cd.mu.Unlock()
		close(request.ResponseChan)
		return nil, fmt.Errorf("discovery request timed out")
	}
}

// DiscoverChallengersByRole discovers challengers with specific roles
func (cd *ChallengerDiscovery) DiscoverChallengersByRole(role types.ChallengerRole, maxPeers int) ([]*types.ChallengerInfo, error) {
	return cd.DiscoverChallengers(maxPeers, role)
}

// DiscoverAllChallengers discovers all available challengers
func (cd *ChallengerDiscovery) DiscoverAllChallengers(maxPeers int) ([]*types.ChallengerInfo, error) {
	return cd.DiscoverChallengers(maxPeers, 0) // 0 means any role
}

// performDiscovery performs the actual discovery operation
func (cd *ChallengerDiscovery) performDiscovery(request *discoveryRequest) {
	defer func() {
		if r := recover(); r != nil {
			cd.logger.Error("Discovery operation panic recovered", "error", r)
			close(request.ResponseChan)
		}
	}()

	cd.logger.Debug("Performing challenger discovery",
		"request_id", request.RequestID, "max_peers", request.MaxPeers)

	var discoveredChallengers []*types.ChallengerInfo

	// Phase 1: Discover from bootstrap nodes
	bootstrapChallengers := cd.discoverFromBootstrap(request)
	discoveredChallengers = append(discoveredChallengers, bootstrapChallengers...)

	// Phase 1: Discover from P2P network discovery service
	networkChallengers := cd.discoverFromNetwork(request)
	discoveredChallengers = append(discoveredChallengers, networkChallengers...)

	// Phase 1: Discover from known challengers
	knownChallengers := cd.discoverFromKnown(request)
	discoveredChallengers = append(discoveredChallengers, knownChallengers...)

	// Remove duplicates and filter by requirements
	filteredChallengers := cd.filterAndDeduplicateChallengers(discoveredChallengers, request)

	// Limit results
	if request.MaxPeers > 0 && len(filteredChallengers) > request.MaxPeers {
		filteredChallengers = filteredChallengers[:request.MaxPeers]
	}

	cd.logger.Debug("Discovery completed",
		"request_id", request.RequestID, "found", len(filteredChallengers))

	// Send response
	select {
	case request.ResponseChan <- filteredChallengers:
	case <-cd.ctx.Done():
	}
}

// discoverFromBootstrap discovers challengers from bootstrap nodes
func (cd *ChallengerDiscovery) discoverFromBootstrap(request *discoveryRequest) []*types.ChallengerInfo {
	var challengers []*types.ChallengerInfo

	// Phase 1: Simple approach - treat bootstrap nodes as potential challengers
	for _, bootstrapAddr := range cd.bootstrapNodes {
		if !utils.IsValidAddress(bootstrapAddr) {
			continue
		}

		// Create a basic challenger info for bootstrap node
		// In a real implementation, we would query the node to get actual challenger info
		challengerID := utils.GenerateChallengerID(bootstrapAddr, time.Now())
		challengerInfo := &types.ChallengerInfo{
			ID:       challengerID,
			NodeID:   utils.PublicKeyToID(nil), // Would get from actual node
			Address:  bootstrapAddr,
			Role:     types.ChallengerRoleChallenger,
			Status:   types.ChallengerStatusOnline,
			LastSeen: time.Now(),
		}

		challengers = append(challengers, challengerInfo)
	}

	cd.logger.Debug("Discovered challengers from bootstrap", "count", len(challengers))
	return challengers
}

// discoverFromNetwork discovers challengers from P2P network
func (cd *ChallengerDiscovery) discoverFromNetwork(request *discoveryRequest) []*types.ChallengerInfo {
	// Get P2P node's discovery service
	p2pNode := cd.manager.GetP2PNode()
	if p2pNode == nil {
		return nil
	}

	discoveryService := p2pNode.GetDiscoveryService()
	if discoveryService == nil {
		return nil
	}

	// Get challenger nodes from network discovery
	challengerNodes := discoveryService.GetAllChallengerNodes()

	var challengers []*types.ChallengerInfo
	for _, info := range challengerNodes {
		// Create a copy to prevent modification
		infoCopy := *info
		challengers = append(challengers, &infoCopy)
	}

	cd.logger.Debug("Discovered challengers from network", "count", len(challengers))
	return challengers
}

// discoverFromKnown discovers challengers from known challengers list
func (cd *ChallengerDiscovery) discoverFromKnown(request *discoveryRequest) []*types.ChallengerInfo {
	// Get known challengers from manager
	knownChallengers := cd.manager.GetAllChallengers()

	var challengers []*types.ChallengerInfo
	for _, info := range knownChallengers {
		// Create a copy to prevent modification
		infoCopy := *info
		challengers = append(challengers, &infoCopy)
	}

	cd.logger.Debug("Discovered challengers from known list", "count", len(challengers))
	return challengers
}

// filterAndDeduplicateChallengers filters and removes duplicate challengers
func (cd *ChallengerDiscovery) filterAndDeduplicateChallengers(challengers []*types.ChallengerInfo, request *discoveryRequest) []*types.ChallengerInfo {
	seen := make(map[string]bool)
	var filtered []*types.ChallengerInfo

	for _, challenger := range challengers {
		// Skip if already seen
		if seen[challenger.ID] {
			continue
		}

		// Skip if not online
		if !challenger.IsOnline() {
			continue
		}

		// Skip if doesn't have required role
		if request.RequiredRoles != 0 && !challenger.HasRole(request.RequiredRoles) {
			continue
		}

		// Skip self
		if challenger.ID == cd.manager.nodeID {
			continue
		}

		seen[challenger.ID] = true
		filtered = append(filtered, challenger)
	}

	return filtered
}

// discoveryLoop runs the periodic discovery loop
func (cd *ChallengerDiscovery) discoveryLoop() {
	defer func() {
		if r := recover(); r != nil {
			cd.logger.Error("Discovery loop panic recovered", "error", r)
		}
	}()

	for {
		select {
		case <-cd.ctx.Done():
			return
		case <-cd.discoveryTicker.C:
			cd.performPeriodicDiscovery()
		}
	}
}

// performPeriodicDiscovery performs periodic discovery
func (cd *ChallengerDiscovery) performPeriodicDiscovery() {
	cd.mu.Lock()
	cd.lastDiscovery = time.Now()
	cd.discoveryCount++
	cd.mu.Unlock()

	// Phase 1: Basic periodic discovery
	// Discover a small number of challengers periodically
	challengers, err := cd.DiscoverAllChallengers(10) // Discover up to 10 challengers
	if err != nil {
		cd.logger.Warn("Periodic discovery failed", "error", err)
		return
	}

	// Add discovered challengers to manager
	for _, challenger := range challengers {
		if err := cd.manager.AddChallenger(challenger); err != nil {
			cd.logger.Debug("Failed to add discovered challenger", "challenger_id", challenger.ID, "error", err)
		}
	}

	cd.logger.Debug("Periodic discovery completed", "discovered", len(challengers))
}

// GetDiscoveryStats returns discovery statistics
func (cd *ChallengerDiscovery) GetDiscoveryStats() map[string]interface{} {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	return map[string]interface{}{
		"is_running":       cd.isRunning,
		"last_discovery":   cd.lastDiscovery,
		"discovery_count":  cd.discoveryCount,
		"pending_requests": len(cd.pendingRequests),
	}
}

// IsRunning returns whether the discovery service is running
func (cd *ChallengerDiscovery) IsRunning() bool {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return cd.isRunning
}
