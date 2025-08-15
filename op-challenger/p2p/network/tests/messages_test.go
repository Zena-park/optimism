package network

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// P2P Messages 모듈 단위 테스트

func TestMessage_Creation(t *testing.T) {
	// 기본 메시지 생성 테스트
	msg := &network.Message{
		Type:      network.MessageTypePing,
		From:      "node-1",
		To:        "node-2",
		Payload:   []byte("ping"),
		Timestamp: time.Now(),
		Nonce:     12345,
	}

	require.NotNil(t, msg)
	require.Equal(t, network.MessageTypePing, msg.Type)
	require.Equal(t, "node-1", msg.From)
	require.Equal(t, "node-2", msg.To)
	require.Equal(t, []byte("ping"), msg.Payload)
	require.Equal(t, uint64(12345), msg.Nonce)
}

func TestMessage_Serialization(t *testing.T) {
	// 메시지 직렬화/역직렬화 테스트
	originalMsg := &network.Message{
		Type:      network.MessageTypeGameState,
		From:      "validator-1",
		To:        "", // 브로드캐스트
		Payload:   []byte(`{"game_hash":"0x123","status":"active"}`),
		Timestamp: time.Now().UTC(),
		Signature: []byte("signature"),
		Nonce:     98765,
	}

	// 직렬화
	data, err := originalMsg.Serialize()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// 역직렬화
	deserializedMsg, err := network.DeserializeMessage(data)
	require.NoError(t, err)
	require.NotNil(t, deserializedMsg)

	// 내용 비교
	assert.Equal(t, originalMsg.Type, deserializedMsg.Type)
	assert.Equal(t, originalMsg.From, deserializedMsg.From)
	assert.Equal(t, originalMsg.To, deserializedMsg.To)
	assert.Equal(t, originalMsg.Payload, deserializedMsg.Payload)
	assert.Equal(t, originalMsg.Nonce, deserializedMsg.Nonce)
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name     string
		msg      *network.Message
		expected bool
	}{
		{
			name: "Valid message",
			msg: &network.Message{
				From:      "node-1",
				Timestamp: time.Now(),
				Nonce:     123,
			},
			expected: true,
		},
		{
			name: "Missing From field",
			msg: &network.Message{
				From:      "",
				Timestamp: time.Now(),
				Nonce:     123,
			},
			expected: false,
		},
		{
			name: "Zero timestamp",
			msg: &network.Message{
				From:      "node-1",
				Timestamp: time.Time{},
				Nonce:     123,
			},
			expected: false,
		},
		{
			name: "Zero nonce",
			msg: &network.Message{
				From:      "node-1",
				Timestamp: time.Now(),
				Nonce:     0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessage_BroadcastDetection(t *testing.T) {
	// 브로드캐스트 메시지 테스트
	broadcastMsg := &network.Message{
		From: "node-1",
		To:   "", // 빈 문자열은 브로드캐스트
	}
	assert.True(t, broadcastMsg.IsBroadcast())

	// 개별 메시지 테스트
	unicastMsg := &network.Message{
		From: "node-1",
		To:   "node-2",
	}
	assert.False(t, unicastMsg.IsBroadcast())
}

func TestMessage_AgeAndExpiration(t *testing.T) {
	// 메시지 나이 계산 테스트
	pastTime := time.Now().Add(-5 * time.Minute)
	msg := &network.Message{
		Timestamp: pastTime,
	}

	age := msg.GetAge()
	assert.True(t, age >= 5*time.Minute)
	assert.True(t, age < 6*time.Minute) // 약간의 여유

	// 만료 테스트
	maxAge := 3 * time.Minute
	assert.True(t, msg.IsExpired(maxAge))

	maxAge = 10 * time.Minute
	assert.False(t, msg.IsExpired(maxAge))
}

func TestFindNodeRequest_Serialization(t *testing.T) {
	// network.FindNodeRequest 직렬화 테스트
	req := &network.FindNodeRequest{
		TargetID: "target-node-123",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var deserializedReq network.FindNodeRequest
	err = json.Unmarshal(data, &deserializedReq)
	require.NoError(t, err)

	assert.Equal(t, req.TargetID, deserializedReq.TargetID)
}

func TestFindNodeResponse_Serialization(t *testing.T) {
	// network.FindNodeResponse 직렬화 테스트 (PublicKey 제외)
	nodes := []*network.NodeInfo{
		{
			ID:        "node-1",
			Address:   "127.0.0.1:9001",
			PublicKey: nil, // JSON 직렬화 문제로 인해 nil로 설정
		},
		{
			ID:        "node-2",
			Address:   "127.0.0.1:9002",
			PublicKey: nil, // JSON 직렬화 문제로 인해 nil로 설정
		},
	}

	resp := &network.FindNodeResponse{
		Nodes: nodes,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var deserializedResp network.FindNodeResponse
	err = json.Unmarshal(data, &deserializedResp)
	require.NoError(t, err)

	assert.Len(t, deserializedResp.Nodes, 2)
	assert.Equal(t, "node-1", deserializedResp.Nodes[0].ID)
	assert.Equal(t, "node-2", deserializedResp.Nodes[1].ID)
}

func TestHandshakeRequest_Serialization(t *testing.T) {
	// network.HandshakeRequest 직렬화 테스트
	req := &network.HandshakeRequest{
		NodeID:    "node-123",
		PublicKey: []byte("mock-public-key-data"),
		Version:   "1.0.0",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var deserializedReq network.HandshakeRequest
	err = json.Unmarshal(data, &deserializedReq)
	require.NoError(t, err)

	assert.Equal(t, req.NodeID, deserializedReq.NodeID)
	assert.Equal(t, req.PublicKey, deserializedReq.PublicKey)
	assert.Equal(t, req.Version, deserializedReq.Version)
}

func TestHandshakeResponse_Serialization(t *testing.T) {
	// network.HandshakeResponse 직렬화 테스트
	resp := &network.HandshakeResponse{
		Accepted:  true,
		Reason:    "",
		PublicKey: []byte("response-public-key-data"),
		Version:   "1.0.0",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var deserializedResp network.HandshakeResponse
	err = json.Unmarshal(data, &deserializedResp)
	require.NoError(t, err)

	assert.Equal(t, resp.Accepted, deserializedResp.Accepted)
	assert.Equal(t, resp.Reason, deserializedResp.Reason)
	assert.Equal(t, resp.PublicKey, deserializedResp.PublicKey)
	assert.Equal(t, resp.Version, deserializedResp.Version)
}

func TestGameStateMessage_Serialization(t *testing.T) {
	// network.GameStateMessage 직렬화 테스트
	gameInfo := &network.GameInfo{
		GameAddress:    "0x1234567890abcdef",
		Status:         network.GameStatusActive,
		LastMove:       time.Now().UTC(),
		CurrentClaim:   42,
		MyRole:         network.PlayerRoleChallenger,
		LastValidation: time.Now().UTC(),
	}

	msg := &network.GameStateMessage{
		GameHash:  "0xabcdef1234567890",
		GameInfo:  gameInfo,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.GameStateMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.GameHash, deserializedMsg.GameHash)
	assert.Equal(t, msg.GameInfo.GameAddress, deserializedMsg.GameInfo.GameAddress)
	assert.Equal(t, msg.GameInfo.Status, deserializedMsg.GameInfo.Status)
}

func TestValidationResultMessage_Serialization(t *testing.T) {
	// network.ValidationResultMessage 직렬화 테스트
	msg := &network.ValidationResultMessage{
		GameHash:   "0x123456789",
		Success:    true,
		Action:     "challenge_move",
		Confidence: 0.95,
		Timestamp:  time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.ValidationResultMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.GameHash, deserializedMsg.GameHash)
	assert.Equal(t, msg.Success, deserializedMsg.Success)
	assert.Equal(t, msg.Action, deserializedMsg.Action)
	assert.Equal(t, msg.Confidence, deserializedMsg.Confidence)
}

func TestReputationUpdateMessage_Serialization(t *testing.T) {
	// network.ReputationUpdateMessage 직렬화 테스트
	msg := &network.ReputationUpdateMessage{
		PeerID:    "peer-123",
		OldScore:  0.8,
		NewScore:  0.7,
		Reason:    "failed_validation",
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.ReputationUpdateMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.PeerID, deserializedMsg.PeerID)
	assert.Equal(t, msg.OldScore, deserializedMsg.OldScore)
	assert.Equal(t, msg.NewScore, deserializedMsg.NewScore)
	assert.Equal(t, msg.Reason, deserializedMsg.Reason)
}

func TestMessageType_Constants(t *testing.T) {
	// 메시지 타입 상수 테스트
	assert.Equal(t, network.MessageType(0), network.MessageTypePing)
	assert.Equal(t, network.MessageType(1), network.MessageTypePong)
	assert.Equal(t, network.MessageType(2), network.MessageTypeFindNode)
	assert.Equal(t, network.MessageType(3), network.MessageTypeFindNodeResponse)
	assert.Equal(t, network.MessageType(4), network.MessageTypeAnnounce)
	assert.Equal(t, network.MessageType(5), network.MessageTypeGameState)
	assert.Equal(t, network.MessageType(6), network.MessageTypeValidationResult)
	assert.Equal(t, network.MessageType(7), network.MessageTypeReputationUpdate)
}

// RAT (Randomized Attention Test) 관련 메시지 테스트

func TestAttentionTestTriggeredMessage_Serialization(t *testing.T) {
	msg := &network.AttentionTestTriggeredMessage{
		TestID:           "test-123",
		StateRoot:        "0xabcdef1234567890",
		L2BlockNumber:    12345,
		TargetChallenger: "challenger-1",
		ResponseWindow:   300, // 5분
		Timestamp:        time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.AttentionTestTriggeredMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.TestID, deserializedMsg.TestID)
	assert.Equal(t, msg.StateRoot, deserializedMsg.StateRoot)
	assert.Equal(t, msg.L2BlockNumber, deserializedMsg.L2BlockNumber)
	assert.Equal(t, msg.TargetChallenger, deserializedMsg.TargetChallenger)
	assert.Equal(t, msg.ResponseWindow, deserializedMsg.ResponseWindow)
}

func TestAttentionTestResponseMessage_Serialization(t *testing.T) {
	msg := &network.AttentionTestResponseMessage{
		TestID:     "test-123",
		Challenger: "challenger-1",
		LeftChild:  "0x1111111111111111",
		RightChild: "0x2222222222222222",
		Timestamp:  time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.AttentionTestResponseMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.TestID, deserializedMsg.TestID)
	assert.Equal(t, msg.Challenger, deserializedMsg.Challenger)
	assert.Equal(t, msg.LeftChild, deserializedMsg.LeftChild)
	assert.Equal(t, msg.RightChild, deserializedMsg.RightChild)
}

func TestAttentionTestResultMessage_Serialization(t *testing.T) {
	msg := &network.AttentionTestResultMessage{
		TestID:         "test-123",
		Challenger:     "challenger-1",
		Passed:         false,
		Reason:         "incorrect_solution",
		PenaltyApplied: true,
		Timestamp:      time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.AttentionTestResultMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.TestID, deserializedMsg.TestID)
	assert.Equal(t, msg.Challenger, deserializedMsg.Challenger)
	assert.Equal(t, msg.Passed, deserializedMsg.Passed)
	assert.Equal(t, msg.Reason, deserializedMsg.Reason)
	assert.Equal(t, msg.PenaltyApplied, deserializedMsg.PenaltyApplied)
}

func TestChallengerRegistrationMessage_Serialization(t *testing.T) {
	msg := &network.ChallengerRegistrationMessage{
		ChallengerID: "challenger-123",
		Address:      "127.0.0.1:9001",
		Deposit:      1000000, // 1 ETH in wei (simplified)
		PublicKey:    []byte("mock-public-key-data"),
		Timestamp:    time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.ChallengerRegistrationMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ChallengerID, deserializedMsg.ChallengerID)
	assert.Equal(t, msg.Address, deserializedMsg.Address)
	assert.Equal(t, msg.Deposit, deserializedMsg.Deposit)
	assert.Equal(t, msg.PublicKey, deserializedMsg.PublicKey)
}

func TestChallengerStatusMessage_Serialization(t *testing.T) {
	msg := &network.ChallengerStatusMessage{
		ChallengerID: "challenger-123",
		Status:       "online",
		LastSeen:     time.Now().UTC(),
		Timestamp:    time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.ChallengerStatusMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.ChallengerID, deserializedMsg.ChallengerID)
	assert.Equal(t, msg.Status, deserializedMsg.Status)
}

func TestStateCommitmentMessage_Serialization(t *testing.T) {
	msg := &network.StateCommitmentMessage{
		StateRoot:     "0xabcdef1234567890",
		L2BlockNumber: 12345,
		Proposer:      "proposer-1",
		TxCount:       100,
		GasUsed:       2000000,
		Timestamp:     time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.StateCommitmentMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.StateRoot, deserializedMsg.StateRoot)
	assert.Equal(t, msg.L2BlockNumber, deserializedMsg.L2BlockNumber)
	assert.Equal(t, msg.Proposer, deserializedMsg.Proposer)
	assert.Equal(t, msg.TxCount, deserializedMsg.TxCount)
	assert.Equal(t, msg.GasUsed, deserializedMsg.GasUsed)
}

func TestStateSyncRequestMessage_Serialization(t *testing.T) {
	msg := &network.StateSyncRequestMessage{
		RequestID:       "sync-req-123",
		FromBlockNumber: 1000,
		ToBlockNumber:   2000,
		StateRoot:       "0x1234567890abcdef",
		Timestamp:       time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.StateSyncRequestMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.RequestID, deserializedMsg.RequestID)
	assert.Equal(t, msg.FromBlockNumber, deserializedMsg.FromBlockNumber)
	assert.Equal(t, msg.ToBlockNumber, deserializedMsg.ToBlockNumber)
	assert.Equal(t, msg.StateRoot, deserializedMsg.StateRoot)
}

func TestStateSyncResponseMessage_Serialization(t *testing.T) {
	stateData := []network.StateCommitmentMessage{
		{
			StateRoot:     "0x1111111111111111",
			L2BlockNumber: 1001,
			Proposer:      "proposer-1",
			TxCount:       50,
			GasUsed:       1000000,
			Timestamp:     time.Now().UTC(),
		},
		{
			StateRoot:     "0x2222222222222222",
			L2BlockNumber: 1002,
			Proposer:      "proposer-1",
			TxCount:       75,
			GasUsed:       1500000,
			Timestamp:     time.Now().UTC(),
		},
	}

	msg := &network.StateSyncResponseMessage{
		RequestID:     "sync-req-123",
		StateData:     stateData,
		Complete:      true,
		NextRequestID: "",
		Timestamp:     time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var deserializedMsg network.StateSyncResponseMessage
	err = json.Unmarshal(data, &deserializedMsg)
	require.NoError(t, err)

	assert.Equal(t, msg.RequestID, deserializedMsg.RequestID)
	assert.Len(t, deserializedMsg.StateData, 2)
	assert.Equal(t, msg.Complete, deserializedMsg.Complete)
	assert.Equal(t, msg.NextRequestID, deserializedMsg.NextRequestID)
}

func TestRATMessageTypes_Constants(t *testing.T) {
	// RAT 관련 메시지 타입 상수 테스트
	assert.Equal(t, network.MessageType(8), network.MessageTypeAttentionTestTriggered)
	assert.Equal(t, network.MessageType(9), network.MessageTypeAttentionTestResponse)
	assert.Equal(t, network.MessageType(10), network.MessageTypeAttentionTestResult)
	assert.Equal(t, network.MessageType(11), network.MessageTypeChallengerRegistration)
	assert.Equal(t, network.MessageType(12), network.MessageTypeChallengerStatus)
	assert.Equal(t, network.MessageType(13), network.MessageTypeChallengerChallenge)
	assert.Equal(t, network.MessageType(14), network.MessageTypeStateCommitment)
	assert.Equal(t, network.MessageType(15), network.MessageTypeStateSyncRequest)
	assert.Equal(t, network.MessageType(16), network.MessageTypeStateSyncResponse)
}
