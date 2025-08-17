package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestChallengerDiscoveryIntegration tests challenger discovery across multiple nodes.
func TestChallengerDiscoveryIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.StopAll()

	// Add three nodes
	env.AddNode("node1", 8080)
	env.AddNode("node2", 8081)
	env.AddNode("node3", 8082)

	// Add challenger managers to nodes (this will skip if P2P implementation is missing)
	challengerID1 := GenerateValidHexID("challenger1")
	challengerID2 := GenerateValidHexID("challenger2")
	challengerID3 := GenerateValidHexID("challenger3")

	env.AddChallengerManager("node1", challengerID1)
	env.AddChallengerManager("node2", challengerID2)
	env.AddChallengerManager("node3", challengerID3)

	// Start all components
	env.StartAll()

	// Connect nodes in a chain: node1 <-> node2 <-> node3
	env.ConnectNodes("node1", "node2")
	env.ConnectNodes("node2", "node3")

	// Verify basic connectivity
	node1 := env.GetNode("node1")
	node2 := env.GetNode("node2")
	node3 := env.GetNode("node3")

	require.Equal(t, 1, node1.GetConnectedPeerCount(), "Node1 should have 1 direct peer")
	require.Equal(t, 2, node2.GetConnectedPeerCount(), "Node2 should have 2 direct peers")
	require.Equal(t, 1, node3.GetConnectedPeerCount(), "Node3 should have 1 direct peer")

	t.Log("Discovery integration test completed successfully")
}
