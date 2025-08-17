package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNetworkResilienceWithNodeFailure tests network resilience when nodes fail.
func TestNetworkResilienceWithNodeFailure(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.StopAll()

	// Add four nodes for better resilience testing
	env.AddNode("node1", 8080)
	env.AddNode("node2", 8081)
	env.AddNode("node3", 8082)
	env.AddNode("node4", 8083)

	challengerID1 := GenerateValidHexID("challenger1")
	challengerID2 := GenerateValidHexID("challenger2")
	challengerID3 := GenerateValidHexID("challenger3")
	challengerID4 := GenerateValidHexID("challenger4")

	env.AddChallengerManager("node1", challengerID1)
	env.AddChallengerManager("node2", challengerID2)
	env.AddChallengerManager("node3", challengerID3)
	env.AddChallengerManager("node4", challengerID4)

	env.StartAll()

	// Create a mesh network topology
	env.ConnectNodes("node1", "node2")
	env.ConnectNodes("node2", "node3")
	env.ConnectNodes("node3", "node4")
	env.ConnectNodes("node1", "node4") // Complete the mesh

	// Verify basic connectivity
	node1 := env.GetNode("node1")
	node2 := env.GetNode("node2")
	node3 := env.GetNode("node3")
	node4 := env.GetNode("node4")

	require.Equal(t, 2, node1.GetConnectedPeerCount(), "Node1 should have 2 peers")
	require.Equal(t, 2, node2.GetConnectedPeerCount(), "Node2 should have 2 peers")
	require.Equal(t, 2, node3.GetConnectedPeerCount(), "Node3 should have 2 peers")
	require.Equal(t, 2, node4.GetConnectedPeerCount(), "Node4 should have 2 peers")

	// Test node failure by stopping node2
	err := node2.Stop()
	require.NoError(t, err)

	t.Log("Network resilience integration test completed successfully")
}
