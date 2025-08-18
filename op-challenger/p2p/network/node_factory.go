package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// NodeType represents the type of P2P node to create
type NodeType string

const (
	NodeTypeLibP2P NodeType = "libp2p"
	NodeTypeTCP    NodeType = "tcp"
)

// NewP2PNodeWithType creates a new P2P node of the specified type
func NewP2PNodeWithType(nodeType NodeType, config interface{}, logger log.Logger) (P2PNodeInterface, error) {
	switch nodeType {
	case NodeTypeLibP2P:
		libp2pConfig, ok := config.(*LibP2PNodeConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for libp2p node")
		}
		return NewLibP2PNode(libp2pConfig, logger)
	case NodeTypeTCP:
		tcpConfig, ok := config.(*P2PConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for TCP node")
		}
		return NewP2PNode(tcpConfig, logger)
	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}
}

// P2PNodeInterface defines the common interface for all P2P node types
type P2PNodeInterface interface {
	Start() error
	Stop() error
	GetID() string
	GetAddress() string
	GetStatus() NodeStatus
	BroadcastMessage(msg *Message) error
	
	// Challenger functionality
	EnableChallenger(challengerID string) error
	DisableChallenger()
	IsChallenger() bool
}

// CreateDefaultLibP2PConfig creates a default configuration for LibP2P node
func CreateDefaultLibP2PConfig(listenAddrs []string, bootstrapPeers []string) *LibP2PNodeConfig {
	// Generate a new private key
	privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		// If key generation fails, we'll let NewLibP2PNode handle it
		privKey = nil
	}
	
	return &LibP2PNodeConfig{
		ListenAddrs:     listenAddrs,
		PrivateKey:      privKey,
		BootstrapPeers:  bootstrapPeers,
		MaxPeers:        100,
		TransportConfig: DefaultTransportConfig(),
		DiscoveryConfig: DefaultDiscoveryConfigFactory(),
	}
}

// CreateDefaultTCPConfig creates a default configuration for TCP node
func CreateDefaultTCPConfig(address string, bootstrapNodes []string) *P2PConfig {
	return &P2PConfig{
		Address:         address,
		BootstrapNodes:  bootstrapNodes,
		MaxPeers:        100,
		DiscoveryPort:   0, // Use default
		TransportConfig: DefaultTransportConfig(),
		DiscoveryConfig: DefaultDiscoveryConfigFactory(),
	}
}

// NewP2PNodeLibP2P creates a new libp2p-based P2P node with sensible defaults
func NewP2PNodeLibP2P(listenAddrs []string, bootstrapPeers []string, logger log.Logger) (*LibP2PNode, error) {
	config := CreateDefaultLibP2PConfig(listenAddrs, bootstrapPeers)
	return NewLibP2PNode(config, logger)
}

// NewP2PNodeTCP creates a new TCP-based P2P node with sensible defaults
func NewP2PNodeTCP(address string, bootstrapNodes []string, logger log.Logger) (*P2PNode, error) {
	config := CreateDefaultTCPConfig(address, bootstrapNodes)
	return NewP2PNode(config, logger)
}

// DefaultDiscoveryConfigFactory returns default discovery configuration
func DefaultDiscoveryConfigFactory() DiscoveryConfig {
	return DiscoveryConfig{
		K:     20, // Kademlia bucket size
		Alpha: 3,  // Parallel requests
	}
}

// ConvertLibP2PToLegacy converts a LibP2PNode to legacy P2PNode interface
// This is a temporary adapter for backward compatibility
func ConvertLibP2PToLegacy(libp2pNode *LibP2PNode) *P2PNodeAdapter {
	return &P2PNodeAdapter{
		libp2pNode: libp2pNode,
	}
}

// P2PNodeAdapter adapts LibP2PNode to legacy P2PNode interface
type P2PNodeAdapter struct {
	libp2pNode *LibP2PNode
}

// Legacy interface methods
func (a *P2PNodeAdapter) Start() error {
	return a.libp2pNode.Start()
}

func (a *P2PNodeAdapter) Stop() error {
	return a.libp2pNode.Stop()
}

func (a *P2PNodeAdapter) GetID() string {
	return a.libp2pNode.GetID()
}

func (a *P2PNodeAdapter) GetAddress() string {
	return a.libp2pNode.GetAddress()
}

func (a *P2PNodeAdapter) GetStatus() NodeStatus {
	return a.libp2pNode.GetStatus()
}

func (a *P2PNodeAdapter) BroadcastMessage(msg *Message) error {
	return a.libp2pNode.BroadcastMessage(msg)
}

func (a *P2PNodeAdapter) EnableChallenger(challengerID string) error {
	return a.libp2pNode.EnableChallenger(challengerID)
}

func (a *P2PNodeAdapter) DisableChallenger() {
	a.libp2pNode.DisableChallenger()
}

func (a *P2PNodeAdapter) IsChallenger() bool {
	return a.libp2pNode.IsChallenger()
}

// Additional methods that might be needed for compatibility
func (a *P2PNodeAdapter) GetTransport() Transport {
	return a.libp2pNode.GetTransport()
}

func (a *P2PNodeAdapter) GetPeers() map[string]*Peer {
	return a.libp2pNode.GetPeers()
}

func (a *P2PNodeAdapter) AddPeer(peer *Peer) error {
	return a.libp2pNode.AddPeer(peer)
}

func (a *P2PNodeAdapter) RemovePeer(peerID string) error {
	return a.libp2pNode.RemovePeer(peerID)
}

func (a *P2PNodeAdapter) ConnectAndHandshake(address string) (*Peer, error) {
	return a.libp2pNode.ConnectAndHandshake(address)
}

func (a *P2PNodeAdapter) SendMessage(peer *Peer, message *Message) error {
	return a.libp2pNode.SendMessage(peer, message)
}

// GetLibP2PNode returns the underlying LibP2PNode
func (a *P2PNodeAdapter) GetLibP2PNode() *LibP2PNode {
	return a.libp2pNode
}