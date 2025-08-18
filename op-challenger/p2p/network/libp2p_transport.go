package network

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// P2P protocol ID for challenger network
	ChallengerProtocolID = protocol.ID("/optimism/challenger/1.0.0")
)

// LibP2PTransport implements Transport interface using libp2p
type LibP2PTransport struct {
	host        host.Host
	config      TransportConfig
	logger      log.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	
	// Connection management
	connections map[peer.ID]*LibP2PConnection
	connMu      sync.RWMutex
	
	// Stream handling
	incomingStreams chan network.Stream
	streamHandler   network.StreamHandler
}

// LibP2PConnection wraps a libp2p stream to implement Connection interface
type LibP2PConnection struct {
	stream     network.Stream
	peerID     peer.ID
	remoteAddr net.Addr
	localAddr  net.Addr
	config     TransportConfig
	mu         sync.Mutex
}

// LibP2PTransportConfig contains libp2p-specific configuration
type LibP2PTransportConfig struct {
	TransportConfig
	ListenAddrs    []string // Multiaddresses to listen on
	PrivateKey     crypto.PrivKey // Private key for the host
	BootstrapPeers []string // Bootstrap peer multiaddresses
}

// NewLibP2PTransport creates a new libp2p-based transport
func NewLibP2PTransport(config LibP2PTransportConfig, logger log.Logger) (Transport, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
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
	
	// Create libp2p host options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	}
	
	// Add private key if provided
	if config.PrivateKey != nil {
		opts = append(opts, libp2p.Identity(config.PrivateKey))
	}
	
	// Create the libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	
	transport := &LibP2PTransport{
		host:            h,
		config:          config.TransportConfig,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		connections:     make(map[peer.ID]*LibP2PConnection),
		incomingStreams: make(chan network.Stream, 100),
	}
	
	// Set up stream handler
	transport.streamHandler = transport.handleIncomingStream
	
	return transport, nil
}

// Start starts the libp2p transport
func (t *LibP2PTransport) Start(address string) error {
	t.logger.Info("Starting libp2p transport", "peer_id", t.host.ID(), "addrs", t.host.Addrs())
	
	// Set stream handler for our protocol
	t.host.SetStreamHandler(ChallengerProtocolID, t.streamHandler)
	
	t.logger.Info("LibP2P transport started successfully", 
		"peer_id", t.host.ID().String(),
		"listen_addrs", t.host.Addrs())
	
	return nil
}

// Stop stops the libp2p transport
func (t *LibP2PTransport) Stop() error {
	t.logger.Info("Stopping libp2p transport")
	
	// Close all connections
	t.connMu.Lock()
	for _, conn := range t.connections {
		conn.Close()
	}
	t.connections = make(map[peer.ID]*LibP2PConnection)
	t.connMu.Unlock()
	
	// Close the host
	if err := t.host.Close(); err != nil {
		t.logger.Warn("Error closing libp2p host", "error", err)
	}
	
	// Cancel context
	t.cancel()
	
	t.logger.Info("LibP2P transport stopped")
	return nil
}

// Accept accepts a new connection (waits for incoming streams)
func (t *LibP2PTransport) Accept() (Connection, error) {
	select {
	case stream := <-t.incomingStreams:
		conn := t.wrapStream(stream)
		t.addConnection(stream.Conn().RemotePeer(), conn)
		return conn, nil
	case <-t.ctx.Done():
		return nil, fmt.Errorf("transport context cancelled")
	}
}

// Connect connects to a remote peer
func (t *LibP2PTransport) Connect(address string) (Connection, error) {
	// Parse the multiaddress
	maddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return nil, fmt.Errorf("invalid multiaddress %s: %w", address, err)
	}
	
	// Extract peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract peer info from %s: %w", address, err)
	}
	
	t.logger.Debug("Connecting to peer", "peer_id", peerInfo.ID, "addrs", peerInfo.Addrs)
	
	// Connect to the peer
	ctx, cancel := context.WithTimeout(t.ctx, time.Duration(t.config.ReadTimeout))
	defer cancel()
	
	if err := t.host.Connect(ctx, *peerInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", peerInfo.ID, err)
	}
	
	// Open a stream to the peer
	stream, err := t.host.NewStream(ctx, peerInfo.ID, ChallengerProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream to peer %s: %w", peerInfo.ID, err)
	}
	
	conn := t.wrapStream(stream)
	t.addConnection(peerInfo.ID, conn)
	
	t.logger.Info("Successfully connected to peer", "peer_id", peerInfo.ID)
	return conn, nil
}

// handleIncomingStream handles incoming streams
func (t *LibP2PTransport) handleIncomingStream(stream network.Stream) {
	t.logger.Debug("Received incoming stream", "peer_id", stream.Conn().RemotePeer())
	
	select {
	case t.incomingStreams <- stream:
		// Stream queued for acceptance
	default:
		t.logger.Warn("Incoming stream buffer full, closing stream", "peer_id", stream.Conn().RemotePeer())
		stream.Close()
	}
}

// wrapStream wraps a libp2p stream into our Connection interface
func (t *LibP2PTransport) wrapStream(stream network.Stream) *LibP2PConnection {
	peerID := stream.Conn().RemotePeer()
	
	// Create mock addresses for compatibility
	remoteAddr := &P2PAddr{
		PeerID: peerID,
		Addr:   stream.Conn().RemoteMultiaddr().String(),
	}
	localAddr := &P2PAddr{
		PeerID: t.host.ID(),
		Addr:   stream.Conn().LocalMultiaddr().String(),
	}
	
	return &LibP2PConnection{
		stream:     stream,
		peerID:     peerID,
		remoteAddr: remoteAddr,
		localAddr:  localAddr,
		config:     t.config,
	}
}

// addConnection adds a connection to the connections map
func (t *LibP2PTransport) addConnection(peerID peer.ID, conn *LibP2PConnection) {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	t.connections[peerID] = conn
}

// removeConnection removes a connection from the connections map
func (t *LibP2PTransport) removeConnection(peerID peer.ID) {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	delete(t.connections, peerID)
}

// GetHost returns the underlying libp2p host
func (t *LibP2PTransport) GetHost() host.Host {
	return t.host
}

// GetConnections returns all active connections
func (t *LibP2PTransport) GetConnections() map[peer.ID]*LibP2PConnection {
	t.connMu.RLock()
	defer t.connMu.RUnlock()
	
	connections := make(map[peer.ID]*LibP2PConnection)
	for id, conn := range t.connections {
		connections[id] = conn
	}
	return connections
}

// LibP2PConnection methods

// Read reads data from the stream
func (c *LibP2PConnection) Read() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Set read deadline
	if err := c.stream.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}
	
	buffer := make([]byte, c.config.BufferSize)
	n, err := c.stream.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}
	
	return buffer[:n], nil
}

// Write writes data to the stream
func (c *LibP2PConnection) Write(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Set write deadline
	if err := c.stream.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}
	
	_, err := c.stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}
	
	return nil
}

// Close closes the stream
func (c *LibP2PConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.stream.Close()
}

// RemoteAddr returns the remote address
func (c *LibP2PConnection) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// LocalAddr returns the local address
func (c *LibP2PConnection) LocalAddr() net.Addr {
	return c.localAddr
}

// SetDeadline sets read and write deadlines
func (c *LibP2PConnection) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if err := c.stream.SetReadDeadline(t); err != nil {
		return err
	}
	return c.stream.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline
func (c *LibP2PConnection) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (c *LibP2PConnection) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.stream.SetWriteDeadline(t)
}

// GetPeerID returns the peer ID
func (c *LibP2PConnection) GetPeerID() peer.ID {
	return c.peerID
}

// GetStream returns the underlying stream
func (c *LibP2PConnection) GetStream() network.Stream {
	return c.stream
}

// P2PAddr implements net.Addr for P2P addresses
type P2PAddr struct {
	PeerID peer.ID
	Addr   string
}

func (a *P2PAddr) Network() string {
	return "p2p"
}

func (a *P2PAddr) String() string {
	return a.Addr
}