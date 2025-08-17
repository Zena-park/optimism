package network

import (
	"encoding/json"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// MessageType represents the type of P2P message
type MessageType int

const (
	MessageTypePing MessageType = iota
	MessageTypePong
	MessageTypeFindNode
	MessageTypeFindNodeResponse
	MessageTypeAnnounce
	MessageTypeGameState
	MessageTypeValidationResult
	MessageTypeReputationUpdate
	// RAT (Randomized Attention Test) 관련 메시지
	MessageTypeAttentionTestTriggered
	MessageTypeAttentionTestResponse
	MessageTypeAttentionTestResult
	MessageTypeChallengerRegistration
	MessageTypeChallengerStatus
	MessageTypeChallengerChallenge
	MessageTypeStateCommitment
	MessageTypeStateSyncRequest
	MessageTypeStateSyncResponse
	// Phase 1 basic challenger messages
	MessageTypeChallengerHello
	MessageTypeChallengerGoodbye
	MessageTypeChallengerHeartbeat
	MessageTypeChallengerStateUpdate
	MessageTypeChallengerPeerDiscovery
	MessageTypeChallengerPeerResponse
)

// Message represents a P2P network message
type Message struct {
	Type      MessageType `json:"type"`
	From      string      `json:"from"`
	To        string      `json:"to,omitempty"` // Empty for broadcast
	Payload   []byte      `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	Signature []byte      `json:"signature"`
	Nonce     uint64      `json:"nonce"` // Prevent replay attacks
}

// FindNodeRequest represents a FIND_NODE RPC request
type FindNodeRequest struct {
	TargetID string `json:"target_id"`
}

// FindNodeResponse represents a FIND_NODE RPC response
type FindNodeResponse struct {
	Nodes []*NodeInfo `json:"nodes"`
}

// AnnounceRequest represents an ANNOUNCE RPC request
type AnnounceRequest struct {
	NodeInfo *NodeInfo `json:"node_info"`
}

// AnnounceResponse represents an ANNOUNCE RPC response
type AnnounceResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

// HandshakeRequest represents a connection handshake request
type HandshakeRequest struct {
	NodeID    string `json:"node_id"`
	PublicKey []byte `json:"public_key"` // Serialized public key
	Version   string `json:"version"`
}

// HandshakeResponse represents a connection handshake response
type HandshakeResponse struct {
	Accepted  bool   `json:"accepted"`
	Reason    string `json:"reason,omitempty"`
	PublicKey []byte `json:"public_key"` // Serialized public key
	Version   string `json:"version"`
}

// GameStateMessage represents a game state update message
type GameStateMessage struct {
	GameHash  string    `json:"game_hash"`
	GameInfo  *GameInfo `json:"game_info"`
	Timestamp time.Time `json:"timestamp"`
}

// ValidationResultMessage represents a validation result message
type ValidationResultMessage struct {
	GameHash   string    `json:"game_hash"`
	Success    bool      `json:"success"`
	Action     string    `json:"action"`
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
}

// ReputationUpdateMessage represents a reputation update message
type ReputationUpdateMessage struct {
	PeerID    string    `json:"peer_id"`
	OldScore  float64   `json:"old_score"`
	NewScore  float64   `json:"new_score"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// RAT (Randomized Attention Test) 관련 메시지 구조체들

// AttentionTestTriggeredMessage represents an attention test trigger notification
type AttentionTestTriggeredMessage struct {
	TestID           string    `json:"test_id"`
	StateRoot        string    `json:"state_root"`
	L2BlockNumber    uint64    `json:"l2_block_number"`
	TargetChallenger string    `json:"target_challenger"`
	ResponseWindow   uint64    `json:"response_window"` // seconds
	Timestamp        time.Time `json:"timestamp"`
}

// AttentionTestResponseMessage represents a challenger's response to attention test
type AttentionTestResponseMessage struct {
	TestID     string    `json:"test_id"`
	Challenger string    `json:"challenger"`
	LeftChild  string    `json:"left_child"`  // 왼쪽 자식 해시
	RightChild string    `json:"right_child"` // 오른쪽 자식 해시
	Timestamp  time.Time `json:"timestamp"`
}

// AttentionTestResultMessage represents the result of an attention test
type AttentionTestResultMessage struct {
	TestID         string    `json:"test_id"`
	Challenger     string    `json:"challenger"`
	Passed         bool      `json:"passed"`
	Reason         string    `json:"reason,omitempty"`
	PenaltyApplied bool      `json:"penalty_applied"`
	Timestamp      time.Time `json:"timestamp"`
}

// ChallengerRegistrationMessage represents challenger registration announcement
type ChallengerRegistrationMessage struct {
	ChallengerID string    `json:"challenger_id"`
	Address      string    `json:"address"`
	Deposit      uint64    `json:"deposit"`
	PublicKey    []byte    `json:"public_key"`
	Timestamp    time.Time `json:"timestamp"`
}

// ChallengerStatusMessage represents challenger online/offline status
type ChallengerStatusMessage struct {
	ChallengerID string    `json:"challenger_id"`
	Status       string    `json:"status"` // "online", "offline", "active", "inactive"
	LastSeen     time.Time `json:"last_seen"`
	Timestamp    time.Time `json:"timestamp"`
}

// ChallengerChallengeMessage represents a challenge to a challenger
type ChallengerChallengeMessage struct {
	ChallengeID      string    `json:"challenge_id"`
	TargetChallenger string    `json:"target_challenger"`
	ChallengeType    string    `json:"challenge_type"` // "attention_test", "fraud_proof", etc.
	StateRoot        string    `json:"state_root,omitempty"`
	L2BlockNumber    uint64    `json:"l2_block_number,omitempty"`
	ResponseWindow   uint64    `json:"response_window"` // seconds
	Timestamp        time.Time `json:"timestamp"`
}

// StateCommitmentMessage represents L2 state commitment sharing
type StateCommitmentMessage struct {
	StateRoot     string    `json:"state_root"`
	L2BlockNumber uint64    `json:"l2_block_number"`
	Proposer      string    `json:"proposer"`
	TxCount       uint64    `json:"tx_count"`
	GasUsed       uint64    `json:"gas_used"`
	Timestamp     time.Time `json:"timestamp"`
}

// StateSyncRequestMessage represents a request for state synchronization
type StateSyncRequestMessage struct {
	RequestID       string    `json:"request_id"`
	FromBlockNumber uint64    `json:"from_block_number"`
	ToBlockNumber   uint64    `json:"to_block_number"`
	StateRoot       string    `json:"state_root,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

// StateSyncResponseMessage represents a response to state sync request
type StateSyncResponseMessage struct {
	RequestID     string                   `json:"request_id"`
	StateData     []StateCommitmentMessage `json:"state_data"`
	Complete      bool                     `json:"complete"`
	NextRequestID string                   `json:"next_request_id,omitempty"`
	Timestamp     time.Time                `json:"timestamp"`
}

// Serialize serializes a message to JSON
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage deserializes a JSON message
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// IsBroadcast returns true if the message is a broadcast message
func (m *Message) IsBroadcast() bool {
	return m.To == ""
}

// IsValid checks if the message is valid
func (m *Message) IsValid() bool {
	if m.From == "" {
		return false
	}
	if m.Timestamp.IsZero() {
		return false
	}
	if m.Nonce == 0 {
		return false
	}
	return true
}

// Phase 1 Challenger Message Structures

// ChallengerHelloMessage represents a challenger introduction message
type ChallengerHelloMessage struct {
	ChallengerInfo *types.ChallengerInfo `json:"challenger_info"`
	Timestamp      time.Time             `json:"timestamp"`
}

// ChallengerGoodbyeMessage represents a challenger leaving message
type ChallengerGoodbyeMessage struct {
	ChallengerID string    `json:"challenger_id"`
	Reason       string    `json:"reason,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ChallengerHeartbeatMessage represents a challenger heartbeat message
type ChallengerHeartbeatMessage struct {
	ChallengerID   string                 `json:"challenger_id"`
	Status         types.ChallengerStatus `json:"status"`
	PeerCount      int                    `json:"peer_count"`
	NetworkLatency time.Duration          `json:"network_latency"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ChallengerStateUpdateMessage represents a challenger state update message
type ChallengerStateUpdateMessage struct {
	ChallengerID string             `json:"challenger_id"`
	StateUpdate  *types.StateUpdate `json:"state_update"`
	Timestamp    time.Time          `json:"timestamp"`
}

// ChallengerPeerDiscoveryMessage represents a challenger peer discovery request
type ChallengerPeerDiscoveryMessage struct {
	ChallengerID  string               `json:"challenger_id"`
	RequestID     string               `json:"request_id"`
	MaxPeers      int                  `json:"max_peers"`
	RequiredRoles types.ChallengerRole `json:"required_roles,omitempty"`
	Timestamp     time.Time            `json:"timestamp"`
}

// ChallengerPeerResponseMessage represents a response to challenger peer discovery
type ChallengerPeerResponseMessage struct {
	RequestID    string                  `json:"request_id"`
	ChallengerID string                  `json:"challenger_id"`
	KnownPeers   []*types.ChallengerInfo `json:"known_peers"`
	Timestamp    time.Time               `json:"timestamp"`
}

// CreateChallengerMessage creates a new challenger message with the given type and payload
func CreateChallengerMessage(msgType MessageType, from, to string, payload interface{}) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		From:      from,
		To:        to,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
		Nonce:     uint64(time.Now().UnixNano()), // Simple nonce generation
	}, nil
}

// ParseChallengerMessage parses a challenger message payload into the specified type
func ParseChallengerMessage(msg *Message, target interface{}) error {
	return json.Unmarshal(msg.Payload, target)
}

// IsChallengerMessage returns true if the message is a challenger-specific message
func IsChallengerMessage(msgType MessageType) bool {
	switch msgType {
	case MessageTypeChallengerHello,
		MessageTypeChallengerGoodbye,
		MessageTypeChallengerHeartbeat,
		MessageTypeChallengerStateUpdate,
		MessageTypeChallengerPeerDiscovery,
		MessageTypeChallengerPeerResponse,
		MessageTypeChallengerRegistration,
		MessageTypeChallengerStatus,
		MessageTypeChallengerChallenge:
		return true
	default:
		return false
	}
}

// GetAge returns the age of the message
func (m *Message) GetAge() time.Duration {
	return time.Since(m.Timestamp)
}

// IsExpired checks if the message is expired (older than maxAge)
func (m *Message) IsExpired(maxAge time.Duration) bool {
	return m.GetAge() > maxAge
}
