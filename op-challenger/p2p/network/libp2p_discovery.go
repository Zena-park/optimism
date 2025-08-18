package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	libp2ppeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
)

// LibP2PDiscoveryService implements discovery using libp2p DHT
type LibP2PDiscoveryService struct {
	host       host.Host
	dht        *dht.IpfsDHT
	logger     log.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	
	// Node management
	knownNodes map[libp2ppeer.ID]*NodeInfo
	nodesMu    sync.RWMutex
	
	// Challenger discovery (Phase 1)
	challengerNodes map[string]*types.ChallengerInfo
	challengerMu    sync.RWMutex
	
	// Configuration
	config         DiscoveryConfig
	bootstrapPeers []libp2ppeer.AddrInfo
	
	// Discovery state
	isBootstrapped bool
	bootstrapMu    sync.RWMutex
}

// NewLibP2PDiscoveryService creates a new libp2p-based discovery service
func NewLibP2PDiscoveryService(host host.Host, config DiscoveryConfig, logger log.Logger) (*LibP2PDiscoveryService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create DHT with default options
	kadDHT, err := dht.New(ctx, host)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	
	// Parse bootstrap peers
	var bootstrapPeers []libp2ppeer.AddrInfo
	for _, addr := range config.Bootstrap {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			logger.Warn("Invalid bootstrap address", "addr", addr, "error", err)
			continue
		}
		
		peerInfo, err := libp2ppeer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logger.Warn("Failed to parse peer info from address", "addr", addr, "error", err)
			continue
		}
		
		bootstrapPeers = append(bootstrapPeers, *peerInfo)
	}
	
	service := &LibP2PDiscoveryService{
		host:            host,
		dht:             kadDHT,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		knownNodes:      make(map[libp2ppeer.ID]*NodeInfo),
		challengerNodes: make(map[string]*types.ChallengerInfo),
		config:          config,
		bootstrapPeers:  bootstrapPeers,
		isBootstrapped:  false,
	}
	
	return service, nil
}

// Start starts the libp2p discovery service
func (ds *LibP2PDiscoveryService) Start() error {
	ds.logger.Info("Starting libp2p discovery service", "peer_id", ds.host.ID())
	
	// Bootstrap the DHT
	if err := ds.bootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}
	
	// Start periodic discovery
	go ds.periodicDiscovery()
	
	// Start peer monitoring
	go ds.monitorPeers()
	
	ds.logger.Info("LibP2P discovery service started successfully")
	return nil
}

// Stop stops the libp2p discovery service
func (ds *LibP2PDiscoveryService) Stop() error {
	ds.logger.Info("Stopping libp2p discovery service")
	
	// Close DHT
	if err := ds.dht.Close(); err != nil {
		ds.logger.Warn("Error closing DHT", "error", err)
	}
	
	// Cancel context
	ds.cancel()
	
	ds.logger.Info("LibP2P discovery service stopped")
	return nil
}

// bootstrap connects to bootstrap peers and bootstraps the DHT
func (ds *LibP2PDiscoveryService) bootstrap() error {
	ds.logger.Info("Bootstrapping DHT", "bootstrap_peers", len(ds.bootstrapPeers))
	
	// Connect to bootstrap peers
	var wg sync.WaitGroup
	for _, peerInfo := range ds.bootstrapPeers {
		wg.Add(1)
		go func(pi libp2ppeer.AddrInfo) {
			defer wg.Done()
			
			ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
			defer cancel()
			
			if err := ds.host.Connect(ctx, pi); err != nil {
				ds.logger.Debug("Failed to connect to bootstrap peer", "peer", pi.ID, "error", err)
			} else {
				ds.logger.Info("Connected to bootstrap peer", "peer", pi.ID)
				ds.addKnownNode(pi.ID)
			}
		}(peerInfo)
	}
	
	wg.Wait()
	
	// Bootstrap the DHT
	if err := ds.dht.Bootstrap(ds.ctx); err != nil {
		return fmt.Errorf("DHT bootstrap failed: %w", err)
	}
	
	ds.bootstrapMu.Lock()
	ds.isBootstrapped = true
	ds.bootstrapMu.Unlock()
	
	ds.logger.Info("DHT bootstrapping completed")
	return nil
}

// FindPeers finds peers in the network using DHT
func (ds *LibP2PDiscoveryService) FindPeers(targetID string) ([]*NodeInfo, error) {
	ds.bootstrapMu.RLock()
	bootstrapped := ds.isBootstrapped
	ds.bootstrapMu.RUnlock()
	
	if !bootstrapped {
		return nil, fmt.Errorf("DHT not bootstrapped yet")
	}
	
	// Convert targetID to libp2ppeer.ID if possible
	var peers []libp2ppeer.ID
	
	// Get connected peers
	connectedPeers := ds.host.Network().Peers()
	peers = append(peers, connectedPeers...)
	
	// Find closest peers using DHT
	ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel()
	
	closestPeers, err := ds.dht.GetClosestPeers(ctx, targetID)
	if err != nil {
		ds.logger.Debug("Failed to get closest peers from DHT", "target", targetID, "error", err)
	} else {
		peers = append(peers, closestPeers...)
	}
	
	// Convert peers to NodeInfo
	var nodes []*NodeInfo
	for _, peerID := range peers {
		if peerID == ds.host.ID() {
			continue // Skip self
		}
		
		peerInfo := ds.host.Peerstore().PeerInfo(peerID)
		if len(peerInfo.Addrs) == 0 {
			continue // Skip peers without addresses
		}
		
		nodeInfo := &NodeInfo{
			ID:       peerID.String(),
			Address:  peerInfo.Addrs[0].String(), // Use first address
			LastSeen: time.Now(),
		}
		
		nodes = append(nodes, nodeInfo)
		ds.addKnownNode(peerID)
	}
	
	ds.logger.Debug("Found peers via DHT", "target", targetID, "count", len(nodes))
	return nodes, nil
}

// AnnouncePresence announces the node's presence to the DHT network
func (ds *LibP2PDiscoveryService) AnnouncePresence() error {
	ds.bootstrapMu.RLock()
	bootstrapped := ds.isBootstrapped
	ds.bootstrapMu.RUnlock()
	
	if !bootstrapped {
		return fmt.Errorf("DHT not bootstrapped yet")
	}
	
	// Announce our presence by providing our peer info to the DHT
	ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel()
	
	// Put our own peer info into the DHT
	key := fmt.Sprintf("/challenger/node/%s", ds.host.ID().String())
	value := ds.host.ID().String()
	
	if err := ds.dht.PutValue(ctx, key, []byte(value)); err != nil {
		return fmt.Errorf("failed to announce presence: %w", err)
	}
	
	ds.logger.Debug("Announced presence to DHT", "key", key)
	return nil
}

// periodicDiscovery performs periodic discovery operations
func (ds *LibP2PDiscoveryService) periodicDiscovery() {
	ticker := time.NewTicker(30 * time.Second) // Discover every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.performDiscovery()
		}
	}
}

// performDiscovery performs a discovery round
func (ds *LibP2PDiscoveryService) performDiscovery() {
	ds.bootstrapMu.RLock()
	bootstrapped := ds.isBootstrapped
	ds.bootstrapMu.RUnlock()
	
	if !bootstrapped {
		ds.logger.Debug("Skipping discovery - DHT not bootstrapped")
		return
	}
	
	// Find random peers
	ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel()
	
	// Generate a random key for discovery
	randomKey := fmt.Sprintf("discovery-%d", time.Now().UnixNano())
	
	closestPeers, err := ds.dht.GetClosestPeers(ctx, randomKey)
	if err != nil {
		ds.logger.Debug("Discovery round failed", "error", err)
		return
	}
	
	peerCount := 0
	for _, peerID := range closestPeers {
		if peerID != ds.host.ID() {
			ds.addKnownNode(peerID)
			peerCount++
		}
	}
	
	// Also announce our presence
	ds.AnnouncePresence()
	
	ds.logger.Debug("Discovery round completed", "new_peers", peerCount)
}

// monitorPeers monitors connected peers and updates known nodes
func (ds *LibP2PDiscoveryService) monitorPeers() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.updateConnectedPeers()
		}
	}
}

// updateConnectedPeers updates the list of connected peers
func (ds *LibP2PDiscoveryService) updateConnectedPeers() {
	connectedPeers := ds.host.Network().Peers()
	
	for _, peerID := range connectedPeers {
		ds.addKnownNode(peerID)
	}
	
	// Remove disconnected peers
	ds.nodesMu.Lock()
	for peerID, nodeInfo := range ds.knownNodes {
		// Check if peer is still connected
		if ds.host.Network().Connectedness(peerID) == network.NotConnected {
			// Remove if not seen for too long
			if time.Since(nodeInfo.LastSeen) > 5*time.Minute {
				delete(ds.knownNodes, peerID)
				ds.logger.Debug("Removed stale peer", "peer", peerID)
			}
		} else {
			// Update last seen for connected peers
			nodeInfo.LastSeen = time.Now()
		}
	}
	ds.nodesMu.Unlock()
}

// addKnownNode adds a peer to the known nodes list
func (ds *LibP2PDiscoveryService) addKnownNode(peerID libp2ppeer.ID) {
	ds.nodesMu.Lock()
	defer ds.nodesMu.Unlock()
	
	if _, exists := ds.knownNodes[peerID]; exists {
		// Update last seen
		ds.knownNodes[peerID].LastSeen = time.Now()
		return
	}
	
	peerInfo := ds.host.Peerstore().PeerInfo(peerID)
	address := ""
	if len(peerInfo.Addrs) > 0 {
		address = peerInfo.Addrs[0].String()
	}
	
	ds.knownNodes[peerID] = &NodeInfo{
		ID:       peerID.String(),
		Address:  address,
		LastSeen: time.Now(),
	}
	
	ds.logger.Debug("Added known node", "peer", peerID, "address", address)
}

// GetKnownNodes returns a copy of known nodes
func (ds *LibP2PDiscoveryService) GetKnownNodes() map[string]*NodeInfo {
	ds.nodesMu.RLock()
	defer ds.nodesMu.RUnlock()
	
	nodes := make(map[string]*NodeInfo)
	for peerID, nodeInfo := range ds.knownNodes {
		// Create a copy
		nodeCopy := *nodeInfo
		nodes[peerID.String()] = &nodeCopy
	}
	
	return nodes
}

// AddNode adds a node to the known nodes list
func (ds *LibP2PDiscoveryService) AddNode(nodeInfo *NodeInfo) error {
	peerID, err := libp2ppeer.Decode(nodeInfo.ID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}
	
	ds.addKnownNode(peerID)
	return nil
}

// RemoveNode removes a node from the known nodes list
func (ds *LibP2PDiscoveryService) RemoveNode(nodeID string) error {
	peerID, err := libp2ppeer.Decode(nodeID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}
	
	ds.nodesMu.Lock()
	defer ds.nodesMu.Unlock()
	
	delete(ds.knownNodes, peerID)
	ds.logger.Debug("Removed node", "peer", peerID)
	
	return nil
}

// GetDHT returns the underlying DHT instance
func (ds *LibP2PDiscoveryService) GetDHT() routing.Routing {
	return ds.dht
}

// IsBootstrapped returns whether the DHT is bootstrapped
func (ds *LibP2PDiscoveryService) IsBootstrapped() bool {
	ds.bootstrapMu.RLock()
	defer ds.bootstrapMu.RUnlock()
	return ds.isBootstrapped
}

// Challenger Discovery Methods (Phase 1)

// AddChallengerNode adds a challenger node to the discovery service
func (ds *LibP2PDiscoveryService) AddChallengerNode(challengerInfo *types.ChallengerInfo) error {
	ds.challengerMu.Lock()
	defer ds.challengerMu.Unlock()
	
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
	
	// Add to known challenger nodes
	ds.challengerNodes[challengerInfo.ID] = challengerInfo
	
	// Also announce to DHT
	go ds.announceChallengerNode(challengerInfo)
	
	ds.logger.Debug("Challenger node added to discovery",
		"challenger_id", challengerInfo.ID, "address", challengerInfo.Address)
	return nil
}

// announceChallengerNode announces a challenger node to the DHT
func (ds *LibP2PDiscoveryService) announceChallengerNode(challengerInfo *types.ChallengerInfo) {
	ds.bootstrapMu.RLock()
	bootstrapped := ds.isBootstrapped
	ds.bootstrapMu.RUnlock()
	
	if !bootstrapped {
		return
	}
	
	ctx, cancel := context.WithTimeout(ds.ctx, 30*time.Second)
	defer cancel()
	
	// Put challenger info into DHT
	key := fmt.Sprintf("/challenger/node/%s", challengerInfo.ID)
	value := fmt.Sprintf("%s|%s", challengerInfo.NodeID, challengerInfo.Address)
	
	if err := ds.dht.PutValue(ctx, key, []byte(value)); err != nil {
		ds.logger.Debug("Failed to announce challenger node to DHT", 
			"challenger_id", challengerInfo.ID, "error", err)
	} else {
		ds.logger.Debug("Announced challenger node to DHT", 
			"challenger_id", challengerInfo.ID, "key", key)
	}
}

// FindChallengerNodes finds challenger nodes using DHT
func (ds *LibP2PDiscoveryService) FindChallengerNodes(role types.ChallengerRole, maxResults int) []*types.ChallengerInfo {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()
	
	var results []*types.ChallengerInfo
	count := 0
	
	for _, info := range ds.challengerNodes {
		// Check if the challenger has the required role
		if role != 0 && !info.HasRole(role) {
			continue
		}
		
		// Check if challenger is online
		if !info.IsOnline() {
			continue
		}
		
		// Add to results
		infoCopy := *info
		results = append(results, &infoCopy)
		count++
		
		// Stop if we've reached max results
		if maxResults > 0 && count >= maxResults {
			break
		}
	}
	
	ds.logger.Debug("Found challenger nodes",
		"role", role, "max_results", maxResults, "found", len(results))
	return results
}

// GetAllChallengerNodes returns all known challenger nodes
func (ds *LibP2PDiscoveryService) GetAllChallengerNodes() map[string]*types.ChallengerInfo {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()
	
	// Return a copy to prevent external modification
	nodes := make(map[string]*types.ChallengerInfo)
	for id, info := range ds.challengerNodes {
		infoCopy := *info
		nodes[id] = &infoCopy
	}
	
	return nodes
}