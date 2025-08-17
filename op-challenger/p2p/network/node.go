package network

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/utils"
	"github.com/ethereum/go-ethereum/log"
)

// NodeStatus represents the current status of a P2P node
type NodeStatus int

const (
	NodeStatusStarting NodeStatus = iota
	NodeStatusRunning
	NodeStatusStopping
	NodeStatusStopped
	NodeStatusError
)

// P2PNode represents a P2P network node
type P2PNode struct {
	// Basic information
	id         string            // Node unique ID (public key based)
	address    string            // Node address (IP:Port)
	publicKey  *ecdsa.PublicKey  // Public key
	privateKey *ecdsa.PrivateKey // Private key

	// Network management
	peers     map[string]*Peer  // Connected peers
	discovery *DiscoveryService // Node discovery service
	transport Transport         // Network transport layer

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

	// Logging
	logger log.Logger
}

// EnableChallenger enables challenger functionality for this node
func (n *P2PNode) EnableChallenger(challengerID string) error {
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
	n.challengerInfo = types.NewChallengerInfo(challengerID, n.id, n.address, n.publicKey)
	n.isChallenger = true
	n.challengerPeers = make(map[string]*types.ChallengerInfo)

	n.logger.Info("Challenger functionality enabled", "challenger_id", challengerID)
	return nil
}

// DisableChallenger disables challenger functionality for this node
func (n *P2PNode) DisableChallenger() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger {
		return
	}

	n.challengerInfo = nil
	n.isChallenger = false
	n.challengerPeers = nil

	n.logger.Info("Challenger functionality disabled")
}

// IsChallenger returns whether this node is a challenger
func (n *P2PNode) IsChallenger() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isChallenger
}

// GetChallengerInfo returns the challenger information
func (n *P2PNode) GetChallengerInfo() *types.ChallengerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return nil
	}

	// Return a copy to prevent external modification
	info := *n.challengerInfo
	return &info
}

// UpdateChallengerStatus updates the challenger status
func (n *P2PNode) UpdateChallengerStatus(status types.ChallengerStatus) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return fmt.Errorf("node is not a challenger")
	}

	n.challengerInfo.UpdateStatus(status)
	n.lastSeen = time.Now()

	n.logger.Debug("Challenger status updated", "status", status.String())
	return nil
}

// SetChallengerConnected marks the challenger as connected
func (n *P2PNode) SetChallengerConnected() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return fmt.Errorf("node is not a challenger")
	}

	n.challengerInfo.SetConnected()
	n.lastSeen = time.Now()

	n.logger.Debug("Challenger marked as connected")
	return nil
}

// SetChallengerDisconnected marks the challenger as disconnected
func (n *P2PNode) SetChallengerDisconnected() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return fmt.Errorf("node is not a challenger")
	}

	n.challengerInfo.SetDisconnected()
	n.lastSeen = time.Now()

	n.logger.Debug("Challenger marked as disconnected")
	return nil
}

// UpdateChallengerNetworkInfo updates challenger network information
func (n *P2PNode) UpdateChallengerNetworkInfo(peerCount int, latency time.Duration) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger || n.challengerInfo == nil {
		return fmt.Errorf("node is not a challenger")
	}

	n.challengerInfo.UpdatePeerCount(peerCount)
	n.challengerInfo.UpdateNetworkLatency(latency)
	n.lastSeen = time.Now()

	n.logger.Debug("Challenger network info updated",
		"peer_count", peerCount, "latency", latency)
	return nil
}

// AddChallengerPeer adds a challenger peer to the known peers list
func (n *P2PNode) AddChallengerPeer(challengerInfo *types.ChallengerInfo) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger {
		return fmt.Errorf("node is not a challenger")
	}

	if challengerInfo == nil {
		return fmt.Errorf("challenger info cannot be nil")
	}

	// Validate challenger info
	if err := utils.ValidateChallengerID(challengerInfo.ID); err != nil {
		return fmt.Errorf("invalid challenger peer ID: %w", err)
	}

	if err := utils.ValidateNetworkAddress(challengerInfo.Address); err != nil {
		return fmt.Errorf("invalid challenger peer address: %w", err)
	}

	// Add to known challenger peers
	n.challengerPeers[challengerInfo.ID] = challengerInfo

	n.logger.Debug("Challenger peer added",
		"challenger_id", challengerInfo.ID, "address", challengerInfo.Address)
	return nil
}

// RemoveChallengerPeer removes a challenger peer from the known peers list
func (n *P2PNode) RemoveChallengerPeer(challengerID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger {
		return
	}

	delete(n.challengerPeers, challengerID)
	n.logger.Debug("Challenger peer removed", "challenger_id", challengerID)
}

// GetChallengerPeers returns a copy of known challenger peers
func (n *P2PNode) GetChallengerPeers() map[string]*types.ChallengerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isChallenger {
		return nil
	}

	// Return a copy to prevent external modification
	peers := make(map[string]*types.ChallengerInfo)
	for id, info := range n.challengerPeers {
		infoCopy := *info
		peers[id] = &infoCopy
	}

	return peers
}

// GetChallengerPeer returns information about a specific challenger peer
func (n *P2PNode) GetChallengerPeer(challengerID string) *types.ChallengerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isChallenger {
		return nil
	}

	info, exists := n.challengerPeers[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	infoCopy := *info
	return &infoCopy
}

// GetChallengerPeerCount returns the number of known challenger peers
func (n *P2PNode) GetChallengerPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isChallenger {
		return 0
	}

	return len(n.challengerPeers)
}

// CleanupStaleChallengerPeers removes stale challenger peers
func (n *P2PNode) CleanupStaleChallengerPeers(staleDuration time.Duration) int {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isChallenger {
		return 0
	}

	removed := 0
	for id, info := range n.challengerPeers {
		if info.IsStale(staleDuration) {
			delete(n.challengerPeers, id)
			removed++
			n.logger.Debug("Stale challenger peer removed", "challenger_id", id)
		}
	}

	if removed > 0 {
		n.logger.Info("Stale challenger peers cleaned up", "removed_count", removed)
	}

	return removed
}

// Peer represents a connected peer node
type Peer struct {
	ID         string           // Peer ID
	Address    string           // Peer address
	PublicKey  *ecdsa.PublicKey // Peer public key
	Status     PeerStatus       // Peer status
	Reputation float64          // Peer reputation
	LastSeen   time.Time        // Last activity time
	Connection Connection       // Connection object
}

// PeerStatus represents the status of a peer connection
type PeerStatus int

const (
	PeerStatusConnecting PeerStatus = iota
	PeerStatusConnected
	PeerStatusDisconnected
	PeerStatusError
)

// P2PConfig contains configuration for P2P node
type P2PConfig struct {
	Address         string   // Node address (IP:Port)
	PrivateKeyPath  string   // Path to private key file
	BootstrapNodes  []string // Bootstrap nodes for initial connection
	MaxPeers        int      // Maximum number of peers
	DiscoveryPort   int      // Discovery service port
	TransportConfig TransportConfig
	DiscoveryConfig DiscoveryConfig
}

// NewP2PNode creates a new P2P node
func NewP2PNode(config *P2PConfig, logger log.Logger) (*P2PNode, error) {
	// Generate or load private key
	privateKey, err := loadOrGeneratePrivateKey(config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Extract public key
	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	// Generate node ID (hash of public key)
	nodeID := generateNodeID(publicKey)

	// Initialize network transport layer
	transport, err := NewTransport(config.TransportConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Initialize node discovery service
	discovery := NewDiscoveryService(config.DiscoveryConfig, logger)

	ctx, cancel := context.WithCancel(context.Background())

	return &P2PNode{
		id:         nodeID,
		address:    config.Address,
		publicKey:  publicKey,
		privateKey: privateKey,
		peers:      make(map[string]*Peer),
		discovery:  discovery,
		transport:  transport,
		status:     NodeStatusStarting,
		reputation: 1.0, // Initial reputation
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}, nil
}

// Start starts the P2P node
func (n *P2PNode) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status != NodeStatusStarting {
		return fmt.Errorf("node is not in starting status")
	}

	// Start transport layer
	if err := n.transport.Start(n.address); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Start node discovery service
	if err := n.discovery.Start(n); err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}

	// Start connection acceptance routine
	go n.acceptConnections()

	// Start peer management routine
	go n.managePeers()

	n.status = NodeStatusRunning
	n.logger.Info("P2P node started", "id", n.id, "address", n.address)

	return nil
}

// Stop stops the P2P node
func (n *P2PNode) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status != NodeStatusRunning {
		return fmt.Errorf("node is not running")
	}

	n.status = NodeStatusStopping

	// Cancel context to stop all goroutines
	n.cancel()

	// Stop discovery service
	if err := n.discovery.Stop(); err != nil {
		n.logger.Warn("Failed to stop discovery service", "error", err)
	}

	// Stop transport layer
	if err := n.transport.Stop(); err != nil {
		n.logger.Warn("Failed to stop transport", "error", err)
	}

	// Close all peer connections
	for _, peer := range n.peers {
		if peer.Connection != nil {
			peer.Connection.Close()
		}
	}

	n.status = NodeStatusStopped
	n.logger.Info("P2P node stopped")

	return nil
}

// GetID returns the node ID
func (n *P2PNode) GetID() string {
	return n.id
}

// GetAddress returns the node address
func (n *P2PNode) GetAddress() string {
	return n.address
}

// GetPublicKey returns the node public key
func (n *P2PNode) GetPublicKey() *ecdsa.PublicKey {
	return n.publicKey
}

// GetTransport returns the transport layer
func (n *P2PNode) GetTransport() Transport {
	return n.transport
}

// SetTransport sets the transport layer (for testing)
func (n *P2PNode) SetTransport(transport Transport) {
	n.transport = transport
}

// GetDiscovery returns the discovery service
func (n *P2PNode) GetDiscovery() *DiscoveryService {
	return n.discovery
}

// ConnectAndHandshake connects to a peer and performs handshake
func (n *P2PNode) ConnectAndHandshake(address string) (*Peer, error) {
	// Connect to peer
	conn, err := n.transport.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Perform handshake
	peer, err := n.handleHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	return peer, nil
}

// BroadcastMessage broadcasts a message to all connected peers
func (n *P2PNode) BroadcastMessage(msg *Message) error {
	n.mu.RLock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	n.mu.RUnlock()

	// Sign message if not already signed
	if len(msg.Signature) == 0 {
		signature, err := n.signMessage(msg)
		if err != nil {
			return fmt.Errorf("failed to sign message: %w", err)
		}
		msg.Signature = signature
	}

	// Send to all peers in parallel
	var wg sync.WaitGroup
	var errors []error
	var errorsMu sync.Mutex

	for _, peer := range peers {
		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			if err := n.SendMessage(p, msg); err != nil {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("failed to send to %s: %w", p.ID, err))
				errorsMu.Unlock()
			}
		}(peer)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	return nil
}

// GetStatus returns the current node status
func (n *P2PNode) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

// GetPeers returns a copy of connected peers
func (n *P2PNode) GetPeers() map[string]*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make(map[string]*Peer)
	for id, peer := range n.peers {
		peers[id] = peer
	}
	return peers
}

// AddPeer adds a new peer to the node
func (n *P2PNode) AddPeer(peer *Peer) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status != NodeStatusRunning {
		return fmt.Errorf("node is not running")
	}

	// Check if peer already exists
	if _, exists := n.peers[peer.ID]; exists {
		return fmt.Errorf("peer %s already exists", peer.ID)
	}

	// Check maximum peers limit
	if len(n.peers) >= 100 { // TODO: Make this configurable
		return fmt.Errorf("maximum number of peers reached")
	}

	n.peers[peer.ID] = peer
	n.logger.Info("Peer added", "id", peer.ID, "address", peer.Address)

	return nil
}

// RemovePeer removes a peer from the node
func (n *P2PNode) RemovePeer(peerID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	peer, exists := n.peers[peerID]
	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	// Close connection if exists
	if peer.Connection != nil {
		peer.Connection.Close()
	}

	delete(n.peers, peerID)
	n.logger.Info("Peer removed", "id", peerID)

	return nil
}

// acceptConnections handles incoming connections
func (n *P2PNode) acceptConnections() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			// Accept new connection from transport
			conn, err := n.transport.Accept()
			if err != nil {
				n.logger.Debug("Failed to accept connection", "error", err)
				continue
			}

			// Handle new connection in goroutine
			go n.handleNewConnection(conn)
		}
	}
}

// managePeers manages peer connections and health checks
func (n *P2PNode) managePeers() {
	ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.performHealthChecks()
		}
	}
}

// handleNewConnection handles a new incoming connection
func (n *P2PNode) handleNewConnection(conn Connection) {
	// 서버 측 핸드셰이크 수행
	peer, err := n.handleHandshake(conn)
	if err != nil {
		n.logger.Warn("Handshake failed", "error", err)
		conn.Close()
		return
	}

	// Add peer to node
	if err := n.AddPeer(peer); err != nil {
		n.logger.Warn("Failed to add peer", "error", err)
		conn.Close()
		return
	}

	n.logger.Info("New peer connected", "id", peer.ID, "address", peer.Address)
}

// performHealthChecks performs health checks on all peers
func (n *P2PNode) performHealthChecks() {
	n.mu.RLock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	n.mu.RUnlock()

	for _, peer := range peers {
		if err := n.checkPeerHealth(peer); err != nil {
			n.logger.Warn("Peer health check failed", "peer", peer.ID, "error", err)
			n.RemovePeer(peer.ID)
		}
	}
}

// checkPeerHealth checks the health of a specific peer
func (n *P2PNode) checkPeerHealth(peer *Peer) error {
	// Send ping message
	pingMsg := &Message{
		Type:      MessageTypePing,
		From:      n.id,
		To:        peer.ID,
		Timestamp: time.Now(),
		Nonce:     generateNonce(),
	}

	// Sign message
	signature, err := n.signMessage(pingMsg)
	if err != nil {
		return fmt.Errorf("failed to sign ping message: %w", err)
	}
	pingMsg.Signature = signature

	// Send ping
	if err := n.SendMessage(peer, pingMsg); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// Update last seen time
	peer.LastSeen = time.Now()

	return nil
}

// Helper functions

// loadOrGeneratePrivateKey loads or generates a private key
func loadOrGeneratePrivateKey(path string) (*ecdsa.PrivateKey, error) {
	// TODO: Implement private key loading from file
	// For now, generate a new key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return privateKey, nil
}

// generateNodeID generates a node ID from public key
func generateNodeID(publicKey *ecdsa.PublicKey) string {
	// Serialize public key
	pubBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)

	// Hash the public key
	hash := sha256.Sum256(pubBytes)

	// Return hex string
	return hex.EncodeToString(hash[:])
}

// generateNonce generates a random nonce
func generateNonce() uint64 {
	// TODO: Implement proper nonce generation
	return uint64(time.Now().UnixNano())
}

// signMessage signs a message with the node's private key
func (n *P2PNode) signMessage(message *Message) ([]byte, error) {
	// TODO: Implement message signing
	return []byte{}, nil
}

// SendMessage sends a message to a peer
func (n *P2PNode) SendMessage(peer *Peer, message *Message) error {
	if peer.Connection == nil {
		return fmt.Errorf("peer %s has no active connection", peer.ID)
	}

	// 메시지 직렬화
	data, err := message.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// 메시지 전송
	if err := peer.Connection.Write(data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// PerformHandshake performs handshake with a new connection
func (n *P2PNode) PerformHandshake(conn Connection) (*Peer, error) {
	// 핸드셰이크 요청 생성
	req := &HandshakeRequest{
		NodeID:    n.id,
		PublicKey: serializePublicKey(n.publicKey),
		Version:   "1.0.0",
	}

	// 요청 전송
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handshake request: %w", err)
	}

	if err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send handshake request: %w", err)
	}

	// 응답 수신
	respData, err := conn.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to receive handshake response: %w", err)
	}

	var resp HandshakeResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handshake response: %w", err)
	}

	if !resp.Accepted {
		return nil, fmt.Errorf("handshake rejected: %s", resp.Reason)
	}

	// 피어 객체 생성
	peer := &Peer{
		ID:         generateNodeID(deserializePublicKey(resp.PublicKey)),
		Address:    conn.RemoteAddr().String(),
		PublicKey:  deserializePublicKey(resp.PublicKey),
		Status:     PeerStatusConnected,
		Reputation: 1.0,
		LastSeen:   time.Now(),
		Connection: conn,
	}

	return peer, nil
}

// serializePublicKey serializes a public key to bytes (internal use)
func serializePublicKey(key *ecdsa.PublicKey) []byte {
	if key == nil {
		return nil
	}
	return elliptic.Marshal(key.Curve, key.X, key.Y)
}

// handleHandshake handles incoming handshake request (internal use)
func (n *P2PNode) handleHandshake(conn Connection) (*Peer, error) {
	// 요청 수신
	reqData, err := conn.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake request: %w", err)
	}

	var req HandshakeRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handshake request: %w", err)
	}

	n.logger.Debug("Received handshake request",
		"from", conn.RemoteAddr(),
		"node_id", req.NodeID,
		"version", req.Version)

	// 버전 확인
	if req.Version != "1.0.0" {
		resp := HandshakeResponse{
			Accepted: false,
			Reason:   "version mismatch",
		}
		respData, _ := json.Marshal(resp)
		conn.Write(respData)
		return nil, fmt.Errorf("version mismatch")
	}

	// 응답 전송
	resp := HandshakeResponse{
		Accepted:  true,
		PublicKey: serializePublicKey(n.publicKey),
		Version:   "1.0.0",
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handshake response: %w", err)
	}

	if err := conn.Write(respData); err != nil {
		return nil, fmt.Errorf("failed to send handshake response: %w", err)
	}

	// 피어 객체 생성
	peer := &Peer{
		ID:         req.NodeID,
		Address:    conn.RemoteAddr().String(),
		PublicKey:  deserializePublicKey(req.PublicKey),
		Status:     PeerStatusConnected,
		Reputation: 1.0,
		LastSeen:   time.Now(),
		Connection: conn,
	}

	return peer, nil
}

// deserializePublicKey deserializes bytes to a public key (internal use)
func deserializePublicKey(data []byte) *ecdsa.PublicKey {
	if len(data) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), data)
	if x == nil {
		return nil
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
}

// GetDiscoveryService returns the discovery service
func (n *P2PNode) GetDiscoveryService() *DiscoveryService {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.discovery
}
