package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	// Protocol IDs for different message types
	ChallengerProtocolV1    = protocol.ID("/optimism/challenger/1.0.0")
	DiscoveryProtocolV1     = protocol.ID("/optimism/discovery/1.0.0")
	HeartbeatProtocolV1     = protocol.ID("/optimism/heartbeat/1.0.0")
	StateUpdateProtocolV1   = protocol.ID("/optimism/state-update/1.0.0")
	
	// Stream timeouts
	StreamReadTimeout  = 30 * time.Second
	StreamWriteTimeout = 10 * time.Second
	StreamIdleTimeout  = 2 * time.Minute
)

// MessageHandler interface for handling different message types
type MessageHandler interface {
	HandleMessage(msg *Message, stream network.Stream) error
	GetSupportedMessageTypes() []MessageType
}

// LibP2PMessageRouter handles message routing and protocol management
type LibP2PMessageRouter struct {
	host     host.Host
	logger   log.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	
	// Message handling
	handlers map[MessageType]MessageHandler
	mu       sync.RWMutex
	
	// Stream management
	activeStreams map[peer.ID]map[protocol.ID]network.Stream
	streamsMu     sync.RWMutex
	
	// Message queues for async processing
	incomingMessages chan *MessageWrapper
	outgoingMessages chan *MessageWrapper
	
	// Statistics
	messagesSent     int64
	messagesReceived int64
	statsMu          sync.RWMutex
}

// MessageWrapper wraps a message with additional metadata
type MessageWrapper struct {
	Message  *Message
	Stream   network.Stream
	PeerID   peer.ID
	Protocol protocol.ID
}

// NewLibP2PMessageRouter creates a new message router
func NewLibP2PMessageRouter(host host.Host, logger log.Logger) *LibP2PMessageRouter {
	ctx, cancel := context.WithCancel(context.Background())
	
	router := &LibP2PMessageRouter{
		host:             host,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
		handlers:         make(map[MessageType]MessageHandler),
		activeStreams:    make(map[peer.ID]map[protocol.ID]network.Stream),
		incomingMessages: make(chan *MessageWrapper, 1000),
		outgoingMessages: make(chan *MessageWrapper, 1000),
	}
	
	return router
}

// Start starts the message router
func (mr *LibP2PMessageRouter) Start() error {
	mr.logger.Info("Starting libp2p message router")
	
	// Set up protocol handlers
	mr.host.SetStreamHandler(ChallengerProtocolV1, mr.handleChallengerProtocol)
	mr.host.SetStreamHandler(DiscoveryProtocolV1, mr.handleDiscoveryProtocol)
	mr.host.SetStreamHandler(HeartbeatProtocolV1, mr.handleHeartbeatProtocol)
	mr.host.SetStreamHandler(StateUpdateProtocolV1, mr.handleStateUpdateProtocol)
	
	// Start message processing workers
	go mr.processIncomingMessages()
	go mr.processOutgoingMessages()
	go mr.cleanupIdleStreams()
	
	mr.logger.Info("LibP2P message router started successfully")
	return nil
}

// Stop stops the message router
func (mr *LibP2PMessageRouter) Stop() error {
	mr.logger.Info("Stopping libp2p message router")
	
	// Close all active streams
	mr.streamsMu.Lock()
	for peerID, protocols := range mr.activeStreams {
		for protocolID, stream := range protocols {
			mr.logger.Debug("Closing stream", "peer", peerID, "protocol", protocolID)
			stream.Close()
		}
	}
	mr.activeStreams = make(map[peer.ID]map[protocol.ID]network.Stream)
	mr.streamsMu.Unlock()
	
	// Cancel context
	mr.cancel()
	
	mr.logger.Info("LibP2P message router stopped")
	return nil
}

// RegisterHandler registers a message handler for specific message types
func (mr *LibP2PMessageRouter) RegisterHandler(handler MessageHandler) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	
	for _, msgType := range handler.GetSupportedMessageTypes() {
		mr.handlers[msgType] = handler
		mr.logger.Debug("Registered message handler", "message_type", msgType)
	}
}

// SendMessage sends a message to a specific peer
func (mr *LibP2PMessageRouter) SendMessage(peerID peer.ID, msg *Message) error {
	protocolID := mr.getProtocolForMessageType(msg.Type)
	
	// Get or create stream
	stream, err := mr.getOrCreateStream(peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}
	
	// Queue message for sending
	wrapper := &MessageWrapper{
		Message:  msg,
		Stream:   stream,
		PeerID:   peerID,
		Protocol: protocolID,
	}
	
	select {
	case mr.outgoingMessages <- wrapper:
		return nil
	case <-mr.ctx.Done():
		return fmt.Errorf("message router stopped")
	default:
		return fmt.Errorf("outgoing message queue full")
	}
}

// BroadcastMessage broadcasts a message to all connected peers
func (mr *LibP2PMessageRouter) BroadcastMessage(msg *Message) error {
	connectedPeers := mr.host.Network().Peers()
	
	var errors []error
	for _, peerID := range connectedPeers {
		if err := mr.SendMessage(peerID, msg); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", peerID, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}
	
	return nil
}

// getProtocolForMessageType returns the protocol ID for a message type
func (mr *LibP2PMessageRouter) getProtocolForMessageType(msgType MessageType) protocol.ID {
	switch msgType {
	case MessageTypeChallengerHello, MessageTypeChallengerGoodbye, 
		 MessageTypeChallengerRegistration, MessageTypeChallengerStatus,
		 MessageTypeChallengerChallenge:
		return ChallengerProtocolV1
	case MessageTypeFindNode, MessageTypeFindNodeResponse,
		 MessageTypeChallengerPeerDiscovery, MessageTypeChallengerPeerResponse:
		return DiscoveryProtocolV1
	case MessageTypePing, MessageTypePong, MessageTypeChallengerHeartbeat:
		return HeartbeatProtocolV1
	case MessageTypeStateCommitment, MessageTypeStateSyncRequest, 
		 MessageTypeStateSyncResponse, MessageTypeChallengerStateUpdate:
		return StateUpdateProtocolV1
	default:
		return ChallengerProtocolV1 // Default to challenger protocol
	}
}

// getOrCreateStream gets an existing stream or creates a new one
func (mr *LibP2PMessageRouter) getOrCreateStream(peerID peer.ID, protocolID protocol.ID) (network.Stream, error) {
	mr.streamsMu.Lock()
	defer mr.streamsMu.Unlock()
	
	// Check if stream already exists
	if peerStreams, exists := mr.activeStreams[peerID]; exists {
		if stream, exists := peerStreams[protocolID]; exists {
			// Check if stream is still valid
			if stream.Stat().Direction != network.DirUnknown {
				return stream, nil
			} else {
				// Stream is invalid, remove it
				delete(peerStreams, protocolID)
			}
		}
	}
	
	// Create new stream
	ctx, cancel := context.WithTimeout(mr.ctx, 30*time.Second)
	defer cancel()
	
	stream, err := mr.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	
	// Store the stream
	if mr.activeStreams[peerID] == nil {
		mr.activeStreams[peerID] = make(map[protocol.ID]network.Stream)
	}
	mr.activeStreams[peerID][protocolID] = stream
	
	mr.logger.Debug("Created new stream", "peer", peerID, "protocol", protocolID)
	return stream, nil
}

// Protocol handlers

func (mr *LibP2PMessageRouter) handleChallengerProtocol(stream network.Stream) {
	mr.handleIncomingStream(stream, ChallengerProtocolV1)
}

func (mr *LibP2PMessageRouter) handleDiscoveryProtocol(stream network.Stream) {
	mr.handleIncomingStream(stream, DiscoveryProtocolV1)
}

func (mr *LibP2PMessageRouter) handleHeartbeatProtocol(stream network.Stream) {
	mr.handleIncomingStream(stream, HeartbeatProtocolV1)
}

func (mr *LibP2PMessageRouter) handleStateUpdateProtocol(stream network.Stream) {
	mr.handleIncomingStream(stream, StateUpdateProtocolV1)
}

// handleIncomingStream handles incoming streams for all protocols
func (mr *LibP2PMessageRouter) handleIncomingStream(stream network.Stream, protocolID protocol.ID) {
	peerID := stream.Conn().RemotePeer()
	mr.logger.Debug("Handling incoming stream", "peer", peerID, "protocol", protocolID)
	
	// Set stream deadlines
	stream.SetDeadline(time.Now().Add(StreamIdleTimeout))
	
	// Store the stream
	mr.streamsMu.Lock()
	if mr.activeStreams[peerID] == nil {
		mr.activeStreams[peerID] = make(map[protocol.ID]network.Stream)
	}
	mr.activeStreams[peerID][protocolID] = stream
	mr.streamsMu.Unlock()
	
	// Read messages from the stream
	reader := bufio.NewReader(stream)
	for {
		// Set read deadline
		stream.SetReadDeadline(time.Now().Add(StreamReadTimeout))
		
		// Read message length
		var msgLen uint32
		if err := json.NewDecoder(reader).Decode(&msgLen); err != nil {
			if err != io.EOF {
				mr.logger.Debug("Failed to read message length", "peer", peerID, "error", err)
			}
			break
		}
		
		// Read message data
		msgData := make([]byte, msgLen)
		if _, err := io.ReadFull(reader, msgData); err != nil {
			mr.logger.Debug("Failed to read message data", "peer", peerID, "error", err)
			break
		}
		
		// Deserialize message
		msg, err := DeserializeMessage(msgData)
		if err != nil {
			mr.logger.Warn("Failed to deserialize message", "peer", peerID, "error", err)
			continue
		}
		
		// Queue message for processing
		wrapper := &MessageWrapper{
			Message:  msg,
			Stream:   stream,
			PeerID:   peerID,
			Protocol: protocolID,
		}
		
		select {
		case mr.incomingMessages <- wrapper:
			// Message queued successfully
		case <-mr.ctx.Done():
			return
		default:
			mr.logger.Warn("Incoming message queue full, dropping message", "peer", peerID)
		}
	}
	
	// Clean up stream
	mr.streamsMu.Lock()
	if peerStreams, exists := mr.activeStreams[peerID]; exists {
		delete(peerStreams, protocolID)
		if len(peerStreams) == 0 {
			delete(mr.activeStreams, peerID)
		}
	}
	mr.streamsMu.Unlock()
	
	stream.Close()
	mr.logger.Debug("Closed incoming stream", "peer", peerID, "protocol", protocolID)
}

// processIncomingMessages processes incoming messages
func (mr *LibP2PMessageRouter) processIncomingMessages() {
	for {
		select {
		case wrapper := <-mr.incomingMessages:
			mr.handleIncomingMessage(wrapper)
		case <-mr.ctx.Done():
			return
		}
	}
}

// processOutgoingMessages processes outgoing messages
func (mr *LibP2PMessageRouter) processOutgoingMessages() {
	for {
		select {
		case wrapper := <-mr.outgoingMessages:
			mr.handleOutgoingMessage(wrapper)
		case <-mr.ctx.Done():
			return
		}
	}
}

// handleIncomingMessage handles an incoming message
func (mr *LibP2PMessageRouter) handleIncomingMessage(wrapper *MessageWrapper) {
	mr.statsMu.Lock()
	mr.messagesReceived++
	mr.statsMu.Unlock()
	
	mr.logger.Debug("Processing incoming message", 
		"type", wrapper.Message.Type, "from", wrapper.PeerID)
	
	// Find handler for message type
	mr.mu.RLock()
	handler, exists := mr.handlers[wrapper.Message.Type]
	mr.mu.RUnlock()
	
	if !exists {
		mr.logger.Debug("No handler for message type", "type", wrapper.Message.Type)
		return
	}
	
	// Handle the message
	if err := handler.HandleMessage(wrapper.Message, wrapper.Stream); err != nil {
		mr.logger.Warn("Failed to handle message", 
			"type", wrapper.Message.Type, "from", wrapper.PeerID, "error", err)
	}
}

// handleOutgoingMessage handles an outgoing message
func (mr *LibP2PMessageRouter) handleOutgoingMessage(wrapper *MessageWrapper) {
	mr.statsMu.Lock()
	mr.messagesSent++
	mr.statsMu.Unlock()
	
	mr.logger.Debug("Sending message", 
		"type", wrapper.Message.Type, "to", wrapper.PeerID)
	
	// Serialize message
	msgData, err := wrapper.Message.Serialize()
	if err != nil {
		mr.logger.Error("Failed to serialize message", "error", err)
		return
	}
	
	// Set write deadline
	wrapper.Stream.SetWriteDeadline(time.Now().Add(StreamWriteTimeout))
	
	// Create writer
	writer := bufio.NewWriter(wrapper.Stream)
	
	// Write message length
	msgLen := uint32(len(msgData))
	if err := json.NewEncoder(writer).Encode(msgLen); err != nil {
		mr.logger.Error("Failed to write message length", "error", err)
		return
	}
	
	// Write message data
	if _, err := writer.Write(msgData); err != nil {
		mr.logger.Error("Failed to write message data", "error", err)
		return
	}
	
	// Flush writer
	if err := writer.Flush(); err != nil {
		mr.logger.Error("Failed to flush message", "error", err)
		return
	}
	
	mr.logger.Debug("Message sent successfully", 
		"type", wrapper.Message.Type, "to", wrapper.PeerID)
}

// cleanupIdleStreams periodically cleans up idle streams
func (mr *LibP2PMessageRouter) cleanupIdleStreams() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mr.performStreamCleanup()
		case <-mr.ctx.Done():
			return
		}
	}
}

// performStreamCleanup performs stream cleanup
func (mr *LibP2PMessageRouter) performStreamCleanup() {
	mr.streamsMu.Lock()
	defer mr.streamsMu.Unlock()
	
	now := time.Now()
	for peerID, protocols := range mr.activeStreams {
		for protocolID, stream := range protocols {
			// Check if stream is still valid
			if stream.Stat().Direction == network.DirUnknown {
				mr.logger.Debug("Cleaning up invalid stream", "peer", peerID, "protocol", protocolID)
				delete(protocols, protocolID)
				stream.Close()
			} else {
				// Reset deadline for active streams
				stream.SetDeadline(now.Add(StreamIdleTimeout))
			}
		}
		
		// Remove peer if no protocols left
		if len(protocols) == 0 {
			delete(mr.activeStreams, peerID)
		}
	}
}

// GetStats returns message router statistics
func (mr *LibP2PMessageRouter) GetStats() map[string]interface{} {
	mr.statsMu.RLock()
	sent := mr.messagesSent
	received := mr.messagesReceived
	mr.statsMu.RUnlock()
	
	mr.streamsMu.RLock()
	activeStreamCount := 0
	for _, protocols := range mr.activeStreams {
		activeStreamCount += len(protocols)
	}
	mr.streamsMu.RUnlock()
	
	return map[string]interface{}{
		"messages_sent":      sent,
		"messages_received":  received,
		"active_streams":     activeStreamCount,
		"active_peers":       len(mr.activeStreams),
		"queue_incoming":     len(mr.incomingMessages),
		"queue_outgoing":     len(mr.outgoingMessages),
	}
}

// GetActiveStreams returns information about active streams
func (mr *LibP2PMessageRouter) GetActiveStreams() map[peer.ID][]protocol.ID {
	mr.streamsMu.RLock()
	defer mr.streamsMu.RUnlock()
	
	result := make(map[peer.ID][]protocol.ID)
	for peerID, protocols := range mr.activeStreams {
		var protocolList []protocol.ID
		for protocolID := range protocols {
			protocolList = append(protocolList, protocolID)
		}
		result[peerID] = protocolList
	}
	
	return result
}

// Default message handlers

// BasicChallengerMessageHandler provides basic handling for challenger messages
type BasicChallengerMessageHandler struct {
	logger log.Logger
	nodeID string
}

// NewBasicChallengerMessageHandler creates a new basic challenger message handler
func NewBasicChallengerMessageHandler(nodeID string, logger log.Logger) *BasicChallengerMessageHandler {
	return &BasicChallengerMessageHandler{
		logger: logger,
		nodeID: nodeID,
	}
}

// HandleMessage handles challenger messages
func (h *BasicChallengerMessageHandler) HandleMessage(msg *Message, stream network.Stream) error {
	switch msg.Type {
	case MessageTypePing:
		return h.handlePing(msg, stream)
	case MessageTypePong:
		return h.handlePong(msg, stream)
	case MessageTypeChallengerHello:
		return h.handleChallengerHello(msg, stream)
	case MessageTypeChallengerHeartbeat:
		return h.handleChallengerHeartbeat(msg, stream)
	default:
		h.logger.Debug("Unhandled message type", "type", msg.Type)
		return nil
	}
}

// GetSupportedMessageTypes returns supported message types
func (h *BasicChallengerMessageHandler) GetSupportedMessageTypes() []MessageType {
	return []MessageType{
		MessageTypePing,
		MessageTypePong,
		MessageTypeChallengerHello,
		MessageTypeChallengerGoodbye,
		MessageTypeChallengerHeartbeat,
		MessageTypeChallengerStateUpdate,
	}
}

func (h *BasicChallengerMessageHandler) handlePing(msg *Message, stream network.Stream) error {
	h.logger.Debug("Received ping", "from", msg.From)
	
	// Create pong response
	pong := &Message{
		Type:      MessageTypePong,
		From:      h.nodeID,
		To:        msg.From,
		Timestamp: time.Now(),
		Nonce:     uint64(time.Now().UnixNano()),
	}
	
	// Send pong back through the same stream
	pongData, err := pong.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize pong: %w", err)
	}
	
	writer := bufio.NewWriter(stream)
	msgLen := uint32(len(pongData))
	if err := json.NewEncoder(writer).Encode(msgLen); err != nil {
		return fmt.Errorf("failed to write pong length: %w", err)
	}
	
	if _, err := writer.Write(pongData); err != nil {
		return fmt.Errorf("failed to write pong data: %w", err)
	}
	
	return writer.Flush()
}

func (h *BasicChallengerMessageHandler) handlePong(msg *Message, stream network.Stream) error {
	h.logger.Debug("Received pong", "from", msg.From)
	return nil
}

func (h *BasicChallengerMessageHandler) handleChallengerHello(msg *Message, stream network.Stream) error {
	var hello ChallengerHelloMessage
	if err := ParseChallengerMessage(msg, &hello); err != nil {
		return fmt.Errorf("failed to parse challenger hello: %w", err)
	}
	
	h.logger.Info("Received challenger hello", 
		"challenger_id", hello.ChallengerInfo.ID,
		"from", msg.From)
	
	// TODO: Add challenger to known challengers
	
	return nil
}

func (h *BasicChallengerMessageHandler) handleChallengerHeartbeat(msg *Message, stream network.Stream) error {
	var heartbeat ChallengerHeartbeatMessage
	if err := ParseChallengerMessage(msg, &heartbeat); err != nil {
		return fmt.Errorf("failed to parse challenger heartbeat: %w", err)
	}
	
	h.logger.Debug("Received challenger heartbeat", 
		"challenger_id", heartbeat.ChallengerID,
		"status", heartbeat.Status,
		"peer_count", heartbeat.PeerCount)
	
	// TODO: Update challenger status and metrics
	
	return nil
}