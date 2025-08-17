package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
)

// DiscoveryService handles node discovery in the P2P network
type DiscoveryService struct {
	dht        *DHT                 // Distributed hash table
	nodeID     string               // Current node ID
	bootstrap  []string             // Bootstrap nodes
	knownNodes map[string]*NodeInfo // Known nodes
	mu         sync.RWMutex

	// Challenger-specific discovery (Phase 1 extension)
	challengerNodes map[string]*types.ChallengerInfo // Known challenger nodes
	challengerMu    sync.RWMutex                     // Separate mutex for challenger data

	// Configuration
	config DiscoveryConfig
	logger log.Logger
	node   *P2PNode // Reference to parent node
}

// DHT represents a distributed hash table
type DHT struct {
	buckets [160][]*NodeInfo // Kademlia style buckets
	nodeID  string           // Current node ID
	k       int              // Bucket size (usually 20)
	alpha   int              // Parallel requests (usually 3)
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(config DiscoveryConfig, logger log.Logger) *DiscoveryService {
	return &DiscoveryService{
		dht: &DHT{
			buckets: [160][]*NodeInfo{},
			k:       config.K,
			alpha:   config.Alpha,
		},
		bootstrap:       config.Bootstrap,
		knownNodes:      make(map[string]*NodeInfo),
		challengerNodes: make(map[string]*types.ChallengerInfo), // Phase 1: challenger discovery
		config:          config,
		logger:          logger,
	}
}

// Start starts the discovery service
func (ds *DiscoveryService) Start(node *P2PNode) error {
	ds.node = node
	ds.nodeID = node.GetID()
	ds.dht.nodeID = ds.nodeID

	// Start periodic discovery
	go ds.periodicDiscovery()

	// Announce presence to bootstrap nodes
	go ds.AnnouncePresence()

	// 즉시 첫 번째 Discovery 실행
	go func() {
		time.Sleep(1 * time.Second) // 노드 시작 후 1초 대기
		ds.performDiscovery()
	}()

	ds.logger.Info("Discovery service started", "node_id", ds.nodeID)
	return nil
}

// Stop stops the discovery service
func (ds *DiscoveryService) Stop() error {
	ds.logger.Info("Discovery service stopped")
	return nil
}

// FindPeers finds peers in the network
func (ds *DiscoveryService) FindPeers(targetID string) ([]*NodeInfo, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// 부트스트랩 노드에 연결 시도
	for _, bootstrapAddr := range ds.bootstrap {
		if bootstrapAddr == ds.node.GetAddress() {
			continue // 자신이 부트스트랩 노드인 경우 스킵
		}

		// 부트스트랩 노드에 연결
		ds.logger.Info("Connecting to bootstrap node", "addr", bootstrapAddr)
		conn, err := ds.node.transport.Connect(bootstrapAddr)
		if err != nil {
			ds.logger.Debug("Failed to connect to bootstrap node", "addr", bootstrapAddr, "error", err)
			continue
		}

		// 핸드셰이크 수행
		peer, err := ds.node.PerformHandshake(conn)
		if err != nil {
			ds.logger.Debug("Handshake failed with bootstrap node", "addr", bootstrapAddr, "error", err)
			conn.Close()
			continue
		}

		// 피어 추가
		if err := ds.node.AddPeer(peer); err != nil {
			ds.logger.Debug("Failed to add bootstrap peer", "addr", bootstrapAddr, "error", err)
			conn.Close()
			continue
		}

		// 부트스트랩 노드의 피어 정보 요청
		ds.logger.Info("Requesting peer info from bootstrap node", "addr", bootstrapAddr)
		peerInfos := ds.node.GetPeers()
		for _, peerInfo := range peerInfos {
			if peerInfo.Address != ds.node.GetAddress() {
				ds.logger.Info("Found new peer from bootstrap node", "addr", peerInfo.Address)
				// 새로운 피어에 연결 시도
				conn, err := ds.node.transport.Connect(peerInfo.Address)
				if err != nil {
					ds.logger.Debug("Failed to connect to peer", "addr", peerInfo.Address, "error", err)
					continue
				}

				// 핸드셰이크 수행
				newPeer, err := ds.node.PerformHandshake(conn)
				if err != nil {
					ds.logger.Debug("Handshake failed with peer", "addr", peerInfo.Address, "error", err)
					conn.Close()
					continue
				}

				// 피어 추가
				if err := ds.node.AddPeer(newPeer); err != nil {
					ds.logger.Debug("Failed to add peer", "addr", peerInfo.Address, "error", err)
					conn.Close()
					continue
				}
			}
		}
	}

	// 알려진 노드 반환
	nodes := make([]*NodeInfo, 0, len(ds.knownNodes))
	for _, node := range ds.knownNodes {
		nodes = append(nodes, node)
	}

	ds.logger.Info("Known peers", "count", len(nodes))
	return nodes, nil
}

// AnnouncePresence announces the node's presence to the network
func (ds *DiscoveryService) AnnouncePresence() error {
	// TODO: Implement announcement to bootstrap nodes
	ds.logger.Debug("Announcing presence to network")
	return nil
}

// periodicDiscovery performs periodic discovery of new nodes
func (ds *DiscoveryService) periodicDiscovery() {
	ticker := time.NewTicker(5 * time.Second) // Discover every 5 seconds for testing
	defer ticker.Stop()

	for {
		select {
		case <-ds.node.ctx.Done():
			return
		case <-ticker.C:
			ds.performDiscovery()
		}
	}
}

// performDiscovery performs a discovery round
func (ds *DiscoveryService) performDiscovery() {
	ds.logger.Info("Performing discovery round", "node_id", ds.nodeID, "bootstrap_nodes", ds.bootstrap)

	// 부트스트랩 노드에 연결 시도
	for _, bootstrapAddr := range ds.bootstrap {
		if bootstrapAddr == ds.node.GetAddress() {
			continue // 자신이 부트스트랩 노드인 경우 스킵
		}

		// 이미 알고 있는 노드인지 확인
		known := false
		ds.mu.RLock()
		for _, node := range ds.knownNodes {
			if node.Address == bootstrapAddr {
				known = true
				break
			}
		}
		ds.mu.RUnlock()

		if known {
			continue
		}

		// 부트스트랩 노드에 연결
		ds.logger.Info("Attempting to connect to bootstrap node", "addr", bootstrapAddr)
		conn, err := ds.node.transport.Connect(bootstrapAddr)
		if err != nil {
			ds.logger.Error("Failed to connect to bootstrap node", "addr", bootstrapAddr, "error", err)
			continue
		}
		ds.logger.Info("TCP connection established with bootstrap node", "addr", bootstrapAddr)

		// 핸드셰이크 수행
		ds.logger.Info("Starting handshake with bootstrap node", "addr", bootstrapAddr)
		peer, err := ds.node.PerformHandshake(conn)
		if err != nil {
			ds.logger.Error("Handshake failed with bootstrap node", "addr", bootstrapAddr, "error", err)
			conn.Close()
			continue
		}
		ds.logger.Info("Handshake completed with bootstrap node", "addr", bootstrapAddr, "peer_id", peer.ID)

		// 피어 추가
		ds.logger.Info("Adding bootstrap node as peer", "addr", bootstrapAddr, "peer_id", peer.ID)
		if err := ds.node.AddPeer(peer); err != nil {
			ds.logger.Error("Failed to add bootstrap peer", "addr", bootstrapAddr, "error", err)
			conn.Close()
			continue
		}

		// 노드 정보 추가
		ds.mu.Lock()
		ds.knownNodes[peer.ID] = &NodeInfo{
			ID:        peer.ID,
			Address:   peer.Address,
			PublicKey: peer.PublicKey,
			LastSeen:  time.Now(),
		}
		ds.mu.Unlock()

		ds.logger.Info("Successfully connected and added bootstrap node", "addr", bootstrapAddr, "peer_id", peer.ID)
	}
}

// addNode adds a node to the known nodes list (internal use)
func (ds *DiscoveryService) AddNode(nodeInfo *NodeInfo) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.knownNodes[nodeInfo.ID] = nodeInfo
	ds.logger.Debug("Added node to known nodes", "id", nodeInfo.ID, "address", nodeInfo.Address)

	return nil
}

// removeNode removes a node from the known nodes list (internal use)
func (ds *DiscoveryService) RemoveNode(nodeID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.knownNodes, nodeID)
	ds.logger.Debug("Removed node from known nodes", "id", nodeID)

	return nil
}

// GetKnownNodes returns a copy of known nodes
func (ds *DiscoveryService) GetKnownNodes() map[string]*NodeInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	nodes := make(map[string]*NodeInfo)
	for id, node := range ds.knownNodes {
		nodes[id] = node
	}

	return nodes
}

// DHT methods (simplified for now)

// FindClosest finds the k closest nodes to the target
func (d *DHT) FindClosest(targetID string, k int) []*NodeInfo {
	// TODO: Implement Kademlia closest node finding
	return []*NodeInfo{}
}

// AddNode adds a node to the DHT
func (d *DHT) AddNode(nodeInfo *NodeInfo) error {
	// TODO: Implement DHT node addition
	return nil
}

// RemoveNode removes a node from the DHT
func (d *DHT) RemoveNode(nodeID string) error {
	// TODO: Implement DHT node removal
	return nil
}

// Phase 1 Challenger Discovery Methods

// AddChallengerNode adds a challenger node to the discovery service
func (ds *DiscoveryService) AddChallengerNode(challengerInfo *types.ChallengerInfo) error {
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

	ds.logger.Debug("Challenger node added to discovery",
		"challenger_id", challengerInfo.ID, "address", challengerInfo.Address)
	return nil
}

// RemoveChallengerNode removes a challenger node from the discovery service
func (ds *DiscoveryService) RemoveChallengerNode(challengerID string) {
	ds.challengerMu.Lock()
	defer ds.challengerMu.Unlock()

	delete(ds.challengerNodes, challengerID)
	ds.logger.Debug("Challenger node removed from discovery", "challenger_id", challengerID)
}

// GetChallengerNode returns information about a specific challenger node
func (ds *DiscoveryService) GetChallengerNode(challengerID string) *types.ChallengerInfo {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()

	info, exists := ds.challengerNodes[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	infoCopy := *info
	return &infoCopy
}

// GetAllChallengerNodes returns all known challenger nodes
func (ds *DiscoveryService) GetAllChallengerNodes() map[string]*types.ChallengerInfo {
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

// FindChallengerNodes finds challenger nodes matching the specified criteria
func (ds *DiscoveryService) FindChallengerNodes(role types.ChallengerRole, maxResults int) []*types.ChallengerInfo {
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

// FindClosestChallengerNodes finds the closest challenger nodes to a target ID
func (ds *DiscoveryService) FindClosestChallengerNodes(targetID string, maxResults int) []*types.ChallengerInfo {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()

	// For Phase 1, we'll use a simple approach
	// In Phase 2-3, this could be enhanced with proper distance calculation
	var results []*types.ChallengerInfo
	count := 0

	for _, info := range ds.challengerNodes {
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

	ds.logger.Debug("Found closest challenger nodes",
		"target_id", targetID, "max_results", maxResults, "found", len(results))
	return results
}

// UpdateChallengerNodeStatus updates the status of a challenger node
func (ds *DiscoveryService) UpdateChallengerNodeStatus(challengerID string, status types.ChallengerStatus) error {
	ds.challengerMu.Lock()
	defer ds.challengerMu.Unlock()

	info, exists := ds.challengerNodes[challengerID]
	if !exists {
		return fmt.Errorf("challenger node not found: %s", challengerID)
	}

	info.UpdateStatus(status)
	ds.logger.Debug("Challenger node status updated",
		"challenger_id", challengerID, "status", status.String())
	return nil
}

// CleanupStaleChallengerNodes removes stale challenger nodes
func (ds *DiscoveryService) CleanupStaleChallengerNodes(staleDuration time.Duration) int {
	ds.challengerMu.Lock()
	defer ds.challengerMu.Unlock()

	removed := 0
	for id, info := range ds.challengerNodes {
		if info.IsStale(staleDuration) {
			delete(ds.challengerNodes, id)
			removed++
			ds.logger.Debug("Stale challenger node removed", "challenger_id", id)
		}
	}

	if removed > 0 {
		ds.logger.Info("Stale challenger nodes cleaned up", "removed_count", removed)
	}

	return removed
}

// GetChallengerNodeCount returns the number of known challenger nodes
func (ds *DiscoveryService) GetChallengerNodeCount() int {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()
	return len(ds.challengerNodes)
}

// GetOnlineChallengerNodeCount returns the number of online challenger nodes
func (ds *DiscoveryService) GetOnlineChallengerNodeCount() int {
	ds.challengerMu.RLock()
	defer ds.challengerMu.RUnlock()

	count := 0
	for _, info := range ds.challengerNodes {
		if info.IsOnline() {
			count++
		}
	}

	return count
}
