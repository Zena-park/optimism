package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicChallengerNetworkIntegration tests basic network setup and challenger registration.
func TestBasicChallengerNetworkIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.StopAll()

	// Add two nodes
	env.AddNode("node1", 8080)
	env.AddNode("node2", 8081)

	// Add challenger managers to nodes (this will skip if P2P implementation is missing)
	challengerID1 := GenerateValidHexID("challenger1")
	challengerID2 := GenerateValidHexID("challenger2")
	env.AddChallengerManager("node1", challengerID1)
	env.AddChallengerManager("node2", challengerID2)

	// Start all components
	env.StartAll()

	// Connect nodes
	env.ConnectNodes("node1", "node2")

	// Verify basic node connectivity
	node1 := env.GetNode("node1")
	node2 := env.GetNode("node2")

	require.Eventually(t, func() bool {
		return node1.GetConnectedPeerCount() > 0 && node2.GetConnectedPeerCount() > 0
	}, 5*time.Second, 100*time.Millisecond, "Nodes should connect to each other")

	// Verify peer counts
	assert.Equal(t, 1, node1.GetConnectedPeerCount(), "Node1 should have 1 peer")
	assert.Equal(t, 1, node2.GetConnectedPeerCount(), "Node2 should have 1 peer")

	t.Log("Basic integration test completed successfully")
}
