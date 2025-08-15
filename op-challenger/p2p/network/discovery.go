package network

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// DiscoveryService handles node discovery in the P2P network
type DiscoveryService struct {
	dht        *DHT                 // Distributed hash table
	nodeID     string               // Current node ID
	bootstrap  []string             // Bootstrap nodes
	knownNodes map[string]*NodeInfo // Known nodes
	mu         sync.RWMutex

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
		bootstrap:  config.Bootstrap,
		knownNodes: make(map[string]*NodeInfo),
		config:     config,
		logger:     logger,
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
