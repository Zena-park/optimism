package network

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	libp2ppeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
)

// LibP2PNode represents a libp2p-based P2P network node
type LibP2PNode struct {
	// Basic information
	id         string            // Node unique ID (derived from libp2p peer ID)
	host       host.Host         // LibP2P host
	privateKey crypto.PrivKey    // LibP2P private key
	publicKey  crypto.PubKey     // LibP2P public key

	// Network management
	transport   *LibP2PTransport          // Network transport layer (wrapper)
	discovery   *LibP2PDiscoveryService   // Node discovery service
	messageRouter *LibP2PMessageRouter    // Message routing
	
	// Legacy peer management (for compatibility)
	peers     map[string]*Peer  // Connected peers (legacy format)
	peersMu   sync.RWMutex      // Peers mutex

	// State management
	status     NodeStatus // Node status
	lastSeen   time.Time  // Last activity time
	reputation float64    // Reputation score

	// Challenger-specific information (Phase 1 extension)
	challengerInfo  *types.ChallengerInfo            // Challenger metadata
	isChallenger    bool                             // Whether this node is a challenger
	challengerPeers map[string]*types.ChallengerInfo // Known challenger peers

	// Concurrency control
	mu     sync.RWMutex       // Read/write mutex
	ctx    context.Context    // Context
	cancel context.CancelFunc // Cancel function

	// Configuration
	config *LibP2PNodeConfig // Configuration
	logger log.Logger        // Logger
}

// LibP2PNodeConfig contains configuration for LibP2P node
type LibP2PNodeConfig struct {
	ListenAddrs    []string     // Multiaddresses to listen on
	PrivateKey     crypto.PrivKey // Private key for the host (optional)
	BootstrapPeers []string     // Bootstrap peer multiaddresses
	MaxPeers       int          // Maximum number of peers
	
	// Transport configuration
	TransportConfig TransportConfig
	
	// Discovery configuration
	DiscoveryConfig DiscoveryConfig
}

// NewLibP2PNode creates a new libp2p-based P2P node
func NewLibP2PNode(config *LibP2PNodeConfig, logger log.Logger) (*LibP2PNode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Generate private key if not provided
	var privKey crypto.PrivKey
	var err error
	if config.PrivateKey != nil {
		privKey = config.PrivateKey
	} else {
		privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
	}
	
	// Extract public key
	pubKey := privKey.GetPublic()
	
	// Convert string addresses to multiaddrs
	var listenAddrs []multiaddr.Multiaddr
	for _, addr := range config.ListenAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("invalid listen address %s: %w", addr, err)
		}
		listenAddrs = append(listenAddrs, maddr)
	}
	
	// Create libp2p host
	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	
	// Generate node ID from peer ID
	nodeID := h.ID().String()
	
	// Create transport wrapper
	transportConfig := LibP2PTransportConfig{
		TransportConfig: config.TransportConfig,
		ListenAddrs:     config.ListenAddrs,
		PrivateKey:      privKey,
		BootstrapPeers:  config.BootstrapPeers,
	}
	
	transport, err := NewLibP2PTransport(transportConfig, logger)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}
	
	// Cast to LibP2PTransport
	libp2pTransport, ok := transport.(*LibP2PTransport)
	if !ok {
		h.Close()
		cancel()
		return nil, fmt.Errorf("transport is not LibP2PTransport")
	}
	
	// Create discovery service
	discovery, err := NewLibP2PDiscoveryService(h, config.DiscoveryConfig, logger)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create discovery service: %w", err)
	}
	
	// Create message router
	messageRouter := NewLibP2PMessageRouter(h, logger)
	
	node := &LibP2PNode{
		id:              nodeID,
		host:            h,
		privateKey:      privKey,
		publicKey:       pubKey,
		transport:       libp2pTransport,
		discovery:       discovery,
		messageRouter:   messageRouter,
		peers:           make(map[string]*Peer),
		status:          NodeStatusStarting,
		reputation:      1.0, // Initial reputation
		challengerPeers: make(map[string]*types.ChallengerInfo),
		ctx:             ctx,
		cancel:          cancel,
		config:          config,
		logger:          logger,
	}
	
	// Register basic message handler
	basicHandler := NewBasicChallengerMessageHandler(nodeID, logger)
	messageRouter.RegisterHandler(basicHandler)
	
	return node, nil
}

// Start starts the libp2p P2P node
func (n *LibP2PNode) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status != NodeStatusStarting {
		return fmt.Errorf("node is not in starting status")
	}

	n.logger.Info("Starting LibP2P P2P node", "id", n.id, "addrs", n.host.Addrs())

	// Start message router
	if err := n.messageRouter.Start(); err != nil {
		return fmt.Errorf("failed to start message router: %w", err)
	}

	// Start discovery service
	if err := n.discovery.Start(); err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}

	// Start peer management routine
	go n.managePeers()

	n.status = NodeStatusRunning
	n.logger.Info("LibP2P P2P node started successfully", "id", n.id)

	return nil
}

// Stop stops the libp2p P2P node
func (n *LibP2PNode) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status != NodeStatusRunning {
		return fmt.Errorf("node is not running")
	}

	n.status = NodeStatusStopping
	n.logger.Info("Stopping LibP2P P2P node")

	// Stop message router
	if err := n.messageRouter.Stop(); err != nil {
		n.logger.Warn("Failed to stop message router", "error", err)
	}

	// Stop discovery service
	if err := n.discovery.Stop(); err != nil {
		n.logger.Warn("Failed to stop discovery service", "error", err)
	}

	// Close the host
	if err := n.host.Close(); err != nil {
		n.logger.Warn("Failed to close libp2p host", "error", err)
	}

	// Cancel context
	n.cancel()

	n.status = NodeStatusStopped
	n.logger.Info("LibP2P P2P node stopped")

	return nil
}

// GetID returns the node ID
func (n *LibP2PNode) GetID() string {
	return n.id
}

// GetAddress returns the node's primary address
func (n *LibP2PNode) GetAddress() string {
	addrs := n.host.Addrs()
	if len(addrs) > 0 {
		return addrs[0].String()
	}
	return ""
}

// GetPublicKey returns the node public key (converted to ECDSA for compatibility)
func (n *LibP2PNode) GetPublicKey() *ecdsa.PublicKey {
	// This is a compatibility method
	// In a real implementation, you might want to store the original ECDSA key
	// For now, return nil and handle this in the calling code
	return nil
}

// GetHost returns the libp2p host
func (n *LibP2PNode) GetHost() host.Host {
	return n.host
}

// GetTransport returns the transport layer (for compatibility)
func (n *LibP2PNode) GetTransport() Transport {
	return n.transport
}

// GetDiscovery returns the discovery service (for compatibility)
func (n *LibP2PNode) GetDiscovery() *DiscoveryService {
	// Return a wrapper or adapter for compatibility
	return nil
}

// GetLibP2PDiscovery returns the libp2p discovery service
func (n *LibP2PNode) GetLibP2PDiscovery() *LibP2PDiscoveryService {
	return n.discovery
}

// GetMessageRouter returns the message router
func (n *LibP2PNode) GetMessageRouter() *LibP2PMessageRouter {
	return n.messageRouter
}

// ConnectAndHandshake connects to a peer (libp2p handles handshaking automatically)
func (n *LibP2PNode) ConnectAndHandshake(address string) (*Peer, error) {
	// Parse the multiaddress
	maddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return nil, fmt.Errorf("invalid multiaddress %s: %w", address, err)
	}

	// Extract peer info
	peerInfo, err := libp2ppeer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract peer info: %w", err)
	}

	// Connect to the peer
	ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
	defer cancel()

	if err := n.host.Connect(ctx, *peerInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create peer object for compatibility
	peer := &Peer{
		ID:         peerInfo.ID.String(),
		Address:    address,
		Status:     PeerStatusConnected,
		Reputation: 1.0,
		LastSeen:   time.Now(),
		// Connection is handled by libp2p streams
	}

	n.AddPeer(peer)
	return peer, nil
}

// BroadcastMessage broadcasts a message to all connected peers
func (n *LibP2PNode) BroadcastMessage(msg *Message) error {
	return n.messageRouter.BroadcastMessage(msg)
}

// SendMessage sends a message to a specific peer
func (n *LibP2PNode) SendMessage(peerObj *Peer, message *Message) error {
	peerID, err := libp2ppeer.Decode(peerObj.ID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	return n.messageRouter.SendMessage(peerID, message)
}

// SendMessageToPeer sends a message to a peer by peer ID
func (n *LibP2PNode) SendMessageToPeer(peerID libp2ppeer.ID, message *Message) error {
	return n.messageRouter.SendMessage(peerID, message)
}

// GetStatus returns the current node status
func (n *LibP2PNode) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

// GetPeers returns a copy of connected peers (legacy format)
func (n *LibP2PNode) GetPeers() map[string]*Peer {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()

	peers := make(map[string]*Peer)
	for id, peer := range n.peers {
		peers[id] = peer
	}
	return peers
}

// GetConnectedPeers returns libp2p connected peers
func (n *LibP2PNode) GetConnectedPeers() []libp2ppeer.ID {
	return n.host.Network().Peers()
}

// AddPeer adds a new peer to the node (legacy compatibility)
func (n *LibP2PNode) AddPeer(peer *Peer) error {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if n.status != NodeStatusRunning {
		return fmt.Errorf("node is not running")
	}

	// Check if peer already exists
	if _, exists := n.peers[peer.ID]; exists {
		return fmt.Errorf("peer %s already exists", peer.ID)
	}

	// Check maximum peers limit
	if len(n.peers) >= n.config.MaxPeers {
		return fmt.Errorf("maximum number of peers reached")
	}

	n.peers[peer.ID] = peer
	n.logger.Info("Peer added", "id", peer.ID, "address", peer.Address)

	return nil
}

// RemovePeer removes a peer from the node (legacy compatibility)
func (n *LibP2PNode) RemovePeer(peerID string) error {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	_, exists := n.peers[peerID]
	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	// Close libp2p connection if exists
	if libp2pPeerID, err := libp2ppeer.Decode(peerID); err == nil {
		if err := n.host.Network().ClosePeer(libp2pPeerID); err != nil {
			n.logger.Debug("Failed to close libp2p connection", "peer", peerID, "error", err)
		}
	}

	delete(n.peers, peerID)
	n.logger.Info("Peer removed", "id", peerID)

	return nil
}

// managePeers manages peer connections and updates legacy peer list
func (n *LibP2PNode) managePeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.updatePeerList()
		}
	}
}

// updatePeerList updates the legacy peer list based on libp2p connections
func (n *LibP2PNode) updatePeerList() {
	connectedPeers := n.host.Network().Peers()
	
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	// Add new connected peers
	for _, peerID := range connectedPeers {
		peerIDStr := peerID.String()
		if _, exists := n.peers[peerIDStr]; !exists {
			peerInfo := n.host.Peerstore().PeerInfo(peerID)
			address := ""
			if len(peerInfo.Addrs) > 0 {
				address = peerInfo.Addrs[0].String()
			}

			peer := &Peer{
				ID:         peerIDStr,
				Address:    address,
				Status:     PeerStatusConnected,
				Reputation: 1.0,
				LastSeen:   time.Now(),
			}

			n.peers[peerIDStr] = peer
			n.logger.Debug("Added connected peer to legacy list", "peer", peerIDStr)
		} else {
			// Update last seen for existing peers
			n.peers[peerIDStr].LastSeen = time.Now()
		}
	}

	// Remove disconnected peers
	for peerIDStr := range n.peers {
		if peerID, err := libp2ppeer.Decode(peerIDStr); err != nil {
			delete(n.peers, peerIDStr)
			continue
		} else if n.host.Network().Connectedness(peerID) == network.NotConnected {
			delete(n.peers, peerIDStr)
			n.logger.Debug("Removed disconnected peer from legacy list", "peer", peerIDStr)
		}
	}
}

// Challenger-specific methods (Phase 1 extension)

// EnableChallenger enables challenger functionality for this node
func (n *LibP2PNode) EnableChallenger(challengerID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isChallenger {
		return fmt.Errorf("node is already a challenger")
	}

	// Validate challenger ID
	if err := utils.ValidateChallengerID(challengerID); err != nil {
		return fmt.Errorf("invalid challenger ID: %w", err)
	}

	// Create challenger info
	n.challengerInfo = types.NewChallengerInfo(challengerID, n.id, n.GetAddress(), nil)
	n.isChallenger = true

	n.logger.Info("Challenger functionality enabled", "challenger_id", challengerID)
	return nil
}

// DisableChallenger disables challenger functionality for this node
func (n *LibP2PNode) DisableChallenger() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger {
		return
	}

	n.challengerInfo = nil
	n.isChallenger = false
	n.challengerPeers = make(map[string]*types.ChallengerInfo)

	n.logger.Info("Challenger functionality disabled")
}

// IsChallenger returns whether this node is a challenger
func (n *LibP2PNode) IsChallenger() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isChallenger
}

// GetChallengerInfo returns the challenger information
func (n *LibP2PNode) GetChallengerInfo() *types.ChallengerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return nil
	}

	// Return a copy to prevent external modification
	info := *n.challengerInfo
	return &info
}

// AddChallengerPeer adds a challenger peer to the known peers list
func (n *LibP2PNode) AddChallengerPeer(challengerInfo *types.ChallengerInfo) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if challengerInfo == nil {
		return fmt.Errorf("challenger info cannot be nil")
	}

	// Add to known challenger peers
	n.challengerPeers[challengerInfo.ID] = challengerInfo

	n.logger.Debug("Challenger peer added",
		"challenger_id", challengerInfo.ID, "address", challengerInfo.Address)
	return nil
}

// RemoveChallengerPeer removes a challenger peer from the known peers list
func (n *LibP2PNode) RemoveChallengerPeer(challengerID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.challengerPeers, challengerID)
	n.logger.Debug("Challenger peer removed", "challenger_id", challengerID)
}

// GetChallengerPeerCount returns the number of known challenger peers
func (n *LibP2PNode) GetChallengerPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.challengerPeers)
}

// PerformHandshake is a compatibility method (libp2p handles handshaking)
func (n *LibP2PNode) PerformHandshake(conn Connection) (*Peer, error) {
	// This is a compatibility method for the old interface
	// In libp2p, handshaking is handled automatically
	
	// Extract peer information from the connection
	// This is a simplified implementation
	remoteAddr := conn.RemoteAddr().String()
	
	peer := &Peer{
		ID:         fmt.Sprintf("peer-%d", time.Now().UnixNano()),
		Address:    remoteAddr,
		Status:     PeerStatusConnected,
		Reputation: 1.0,
		LastSeen:   time.Now(),
		Connection: conn,
	}
	
	return peer, nil
}