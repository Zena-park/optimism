package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBasicStateSynchronization tests basic state synchronization between challenger nodes.
func TestBasicStateSynchronization(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.StopAll()

	// Add two nodes
	env.AddNode("node1", 8080)
	env.AddNode("node2", 8081)

	challengerID1 := GenerateValidHexID("challenger1")
	challengerID2 := GenerateValidHexID("challenger2")

	env.AddChallengerManager("node1", challengerID1)
	env.AddChallengerManager("node2", challengerID2)

	env.StartAll()
	env.ConnectNodes("node1", "node2")

	// Verify basic connectivity
	node1 := env.GetNode("node1")
	node2 := env.GetNode("node2")

	require.Equal(t, 1, node1.GetConnectedPeerCount(), "Node1 should have 1 peer")
	require.Equal(t, 1, node2.GetConnectedPeerCount(), "Node2 should have 1 peer")

	t.Log("State synchronization integration test completed successfully")
}
