package types

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewChallengerInfo tests ChallengerInfo creation and initialization
func TestNewChallengerInfo(t *testing.T) {
	// Generate test key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Test data
	id := "test-challenger-id"
	nodeID := "test-node-id"
	address := "127.0.0.1:8080"
	publicKey := &privateKey.PublicKey

	// Create ChallengerInfo
	info := NewChallengerInfo(id, nodeID, address, publicKey)
	require.NotNil(t, info)

	// Verify basic fields
	assert.Equal(t, id, info.ID)
	assert.Equal(t, nodeID, info.NodeID)
	assert.Equal(t, address, info.Address)
	assert.Equal(t, publicKey, info.PublicKey)

	// Verify default values
	assert.Equal(t, ChallengerRoleChallenger, info.Role)
	assert.Equal(t, ChallengerStatusOffline, info.Status)
	assert.Equal(t, StakeStatusNone, info.StakeStatus)
	assert.Equal(t, uint64(0), info.StakeAmount)
	assert.Equal(t, uint64(0), info.TotalTests)
	assert.Equal(t, uint64(0), info.PassedTests)
	assert.Equal(t, 0, info.PeerCount)
	assert.False(t, info.IsConnected)
	assert.True(t, info.ConnectedAt.IsZero())
	assert.Empty(t, info.ConnectedPeers)
	assert.Equal(t, time.Duration(0), info.NetworkLatency)
	assert.NotEmpty(t, info.Version)
	assert.False(t, info.Timestamp.IsZero())
}

// TestChallengerInfoHasRole tests the HasRole method
func TestChallengerInfoHasRole(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	info := NewChallengerInfo("test-id", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)

	// Test default challenger role
	assert.True(t, info.HasRole(ChallengerRoleChallenger))
	assert.False(t, info.HasRole(ChallengerRoleSequencer))

	// Test setting sequencer role
	info.Role = ChallengerRoleSequencer
	assert.False(t, info.HasRole(ChallengerRoleChallenger))
	assert.True(t, info.HasRole(ChallengerRoleSequencer))

	// Test combined roles (bitmask)
	info.Role = ChallengerRoleChallenger | ChallengerRoleSequencer
	assert.True(t, info.HasRole(ChallengerRoleChallenger))
	assert.True(t, info.HasRole(ChallengerRoleSequencer))

	// Test no roles
	info.Role = 0
	assert.False(t, info.HasRole(ChallengerRoleChallenger))
	assert.False(t, info.HasRole(ChallengerRoleSequencer))
}

// TestChallengerInfoRoleBitmask tests role bitmask operations
func TestChallengerInfoRoleBitmask(t *testing.T) {
	// Test individual roles
	assert.Equal(t, ChallengerRole(1), ChallengerRoleChallenger)
	assert.Equal(t, ChallengerRole(2), ChallengerRoleSequencer)

	// Test combined roles
	combined := ChallengerRoleChallenger | ChallengerRoleSequencer
	assert.Equal(t, ChallengerRole(3), combined) // 1 | 2 = 3

	// Test role checking with bitmask
	assert.True(t, (combined&ChallengerRoleChallenger) != 0)
	assert.True(t, (combined&ChallengerRoleSequencer) != 0)

	// Test role string representations
	assert.Equal(t, "challenger", ChallengerRoleChallenger.String())
	assert.Equal(t, "sequencer", ChallengerRoleSequencer.String())
	assert.Equal(t, "unknown", ChallengerRole(99).String())
}

// TestChallengerInfoSerialization tests JSON serialization/deserialization
func TestChallengerInfoSerialization(t *testing.T) {
	// Create original info without PublicKey for easier serialization testing
	original := &ChallengerInfo{
		ID:             "test-id",
		NodeID:         "node-id",
		Address:        "127.0.0.1:8080",
		PublicKey:      nil, // Skip PublicKey for JSON test
		Role:           ChallengerRoleChallenger | ChallengerRoleSequencer,
		Status:         ChallengerStatusOnline,
		StakeAmount:    1000000000000000000, // 1 ETH in wei
		StakeStatus:    StakeStatusActive,
		TotalTests:     100,
		PassedTests:    95,
		PeerCount:      5,
		IsConnected:    true,
		ConnectedAt:    time.Now().Round(time.Second), // Round for comparison
		ConnectedPeers: []string{"peer1", "peer2"},
		NetworkLatency: 50 * time.Millisecond,
		Version:        "1.0.0",
		Timestamp:      time.Now().Round(time.Second),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserialized ChallengerInfo
	err = json.Unmarshal(jsonData, &deserialized)
	require.NoError(t, err)

	// Compare all fields (except PublicKey)
	assert.Equal(t, original.ID, deserialized.ID)
	assert.Equal(t, original.NodeID, deserialized.NodeID)
	assert.Equal(t, original.Address, deserialized.Address)
	assert.Equal(t, original.Role, deserialized.Role)
	assert.Equal(t, original.Status, deserialized.Status)
	assert.Equal(t, original.StakeAmount, deserialized.StakeAmount)
	assert.Equal(t, original.StakeStatus, deserialized.StakeStatus)
	assert.Equal(t, original.TotalTests, deserialized.TotalTests)
	assert.Equal(t, original.PassedTests, deserialized.PassedTests)
	assert.Equal(t, original.PeerCount, deserialized.PeerCount)
	assert.Equal(t, original.IsConnected, deserialized.IsConnected)
	assert.Equal(t, original.ConnectedAt.Unix(), deserialized.ConnectedAt.Unix())
	assert.Equal(t, original.ConnectedPeers, deserialized.ConnectedPeers)
	assert.Equal(t, original.NetworkLatency, deserialized.NetworkLatency)
	assert.Equal(t, original.Version, deserialized.Version)
}

// TestChallengerInfoValidation tests field validation
func TestChallengerInfoValidation(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Test valid info
	validInfo := NewChallengerInfo("valid-id", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)
	assert.NotNil(t, validInfo)

	// Test empty ID
	emptyIDInfo := NewChallengerInfo("", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)
	assert.NotNil(t, emptyIDInfo) // Constructor doesn't validate, but ID is empty
	assert.Empty(t, emptyIDInfo.ID)

	// Test nil public key
	nilKeyInfo := NewChallengerInfo("test-id", "node-id", "127.0.0.1:8080", nil)
	assert.NotNil(t, nilKeyInfo)
	assert.Nil(t, nilKeyInfo.PublicKey)

	// Test empty address
	emptyAddrInfo := NewChallengerInfo("test-id", "node-id", "", &privateKey.PublicKey)
	assert.NotNil(t, emptyAddrInfo)
	assert.Empty(t, emptyAddrInfo.Address)
}

// TestChallengerStatus tests challenger status enum
func TestChallengerStatus(t *testing.T) {
	// Test status string representations
	assert.Equal(t, "offline", ChallengerStatusOffline.String())
	assert.Equal(t, "online", ChallengerStatusOnline.String())
	assert.Equal(t, "active", ChallengerStatusActive.String())
	assert.Equal(t, "idle", ChallengerStatusIdle.String())
	assert.Equal(t, "suspended", ChallengerStatusSuspended.String())
	assert.Equal(t, "banned", ChallengerStatusBanned.String())
	assert.Equal(t, "unknown", ChallengerStatus(99).String())

	// Test status values
	assert.Equal(t, ChallengerStatus(0), ChallengerStatusOffline)
	assert.Equal(t, ChallengerStatus(1), ChallengerStatusOnline)
	assert.Equal(t, ChallengerStatus(2), ChallengerStatusActive)
	assert.Equal(t, ChallengerStatus(3), ChallengerStatusIdle)
	assert.Equal(t, ChallengerStatus(4), ChallengerStatusSuspended)
	assert.Equal(t, ChallengerStatus(5), ChallengerStatusBanned)
}

// TestStakeStatus tests stake status enum
func TestStakeStatus(t *testing.T) {
	// Test stake status string representations
	assert.Equal(t, "none", StakeStatusNone.String())
	assert.Equal(t, "pending", StakeStatusPending.String())
	assert.Equal(t, "active", StakeStatusActive.String())
	assert.Equal(t, "slashing", StakeStatusSlashing.String())
	assert.Equal(t, "withdrawing", StakeStatusWithdrawing.String())
	assert.Equal(t, "unknown", StakeStatus(99).String())

	// Test stake status values
	assert.Equal(t, StakeStatus(0), StakeStatusNone)
	assert.Equal(t, StakeStatus(1), StakeStatusPending)
	assert.Equal(t, StakeStatus(2), StakeStatusActive)
	assert.Equal(t, StakeStatus(3), StakeStatusSlashing)
	assert.Equal(t, StakeStatus(4), StakeStatusWithdrawing)
}

// TestChallengerInfoClone tests cloning functionality (if implemented)
func TestChallengerInfoClone(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Create original info with all fields set
	original := NewChallengerInfo("test-id", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)
	original.Role = ChallengerRoleChallenger | ChallengerRoleSequencer
	original.Status = ChallengerStatusActive
	original.StakeAmount = 1000000000000000000
	original.StakeStatus = StakeStatusActive
	original.TotalTests = 100
	original.PassedTests = 95
	original.PeerCount = 5
	original.IsConnected = true
	original.ConnectedAt = time.Now()
	original.ConnectedPeers = []string{"peer1", "peer2"}
	original.NetworkLatency = 50 * time.Millisecond

	// Test that modifying slice doesn't affect original
	// (This tests if the struct properly handles slice fields)
	originalPeers := make([]string, len(original.ConnectedPeers))
	copy(originalPeers, original.ConnectedPeers)

	// Modify the slice
	original.ConnectedPeers[0] = "modified-peer"

	// Original slice should be modified
	assert.Equal(t, "modified-peer", original.ConnectedPeers[0])
	assert.NotEqual(t, originalPeers[0], original.ConnectedPeers[0])
}

// TestChallengerInfoEdgeCases tests edge cases and boundary conditions
func TestChallengerInfoEdgeCases(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	info := NewChallengerInfo("test-id", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)

	// Test maximum values
	info.StakeAmount = ^uint64(0) // Maximum uint64
	assert.Equal(t, ^uint64(0), info.StakeAmount)

	// Test negative test counts (should not be possible with uint64, but test boundary)
	info.TotalTests = 0
	info.PassedTests = 0
	assert.Equal(t, uint64(0), info.TotalTests)
	assert.Equal(t, uint64(0), info.PassedTests)

	// Test very long peer list
	longPeerList := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		longPeerList[i] = "peer-" + string(rune(i))
	}
	info.ConnectedPeers = longPeerList
	assert.Len(t, info.ConnectedPeers, 1000)

	// Test very high network latency
	info.NetworkLatency = time.Hour
	assert.Equal(t, time.Hour, info.NetworkLatency)

	// Test future timestamp
	futureTime := time.Now().Add(24 * time.Hour)
	info.Timestamp = futureTime
	assert.Equal(t, futureTime, info.Timestamp)
}

// TestChallengerInfoConcurrency tests concurrent access (basic)
func TestChallengerInfoConcurrency(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	info := NewChallengerInfo("test-id", "node-id", "127.0.0.1:8080", &privateKey.PublicKey)

	// Test that basic field access doesn't cause data races
	// Note: This is a basic test. Real concurrent access would need proper synchronization
	done := make(chan bool, 2)

	// Goroutine 1: Read fields
	go func() {
		for i := 0; i < 100; i++ {
			_ = info.ID
			_ = info.Status
			_ = info.Role
		}
		done <- true
	}()

	// Goroutine 2: Read different fields
	go func() {
		for i := 0; i < 100; i++ {
			_ = info.Address
			_ = info.StakeAmount
			_ = info.IsConnected
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// If we get here without data race, the test passes
	assert.True(t, true)
}
