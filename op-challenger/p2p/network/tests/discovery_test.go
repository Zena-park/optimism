package network

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryService_Basic(t *testing.T) {
	logger := log.New()
	config := network.DiscoveryConfig{
		Bootstrap: []string{"localhost:9001", "localhost:9002"},
		K:         20,
		Alpha:     3,
	}

	ds := network.NewDiscoveryService(config, logger)
	require.NotNil(t, ds, "Discovery service should not be nil")

	// Test configuration
	knownNodes := ds.GetKnownNodes()
	assert.Empty(t, knownNodes, "Known nodes should be empty initially")
}

func setupTestNetwork(t *testing.T) (*network.P2PNode, *network.DiscoveryService, *MockTransport) {
	logger := log.New()
	transport := NewMockTransport()

	// 노드 설정
	nodeConfig := &network.P2PConfig{
		Address:        "127.0.0.1:9001",
		BootstrapNodes: []string{},
		MaxPeers:       10,
		DiscoveryPort:  9001,
		DiscoveryConfig: network.DiscoveryConfig{
			Bootstrap: []string{},
			K:         20,
			Alpha:     3,
		},
	}

	node, err := network.NewP2PNode(nodeConfig, logger)
	require.NoError(t, err)

	// Transport 시작
	err = transport.Start("127.0.0.1:9001")
	require.NoError(t, err)

	node.SetTransport(transport) // 테스트용 transport 설정

	service := network.NewDiscoveryService(nodeConfig.DiscoveryConfig, logger)
	require.NoError(t, service.Start(node))

	return node, service, transport
}

func TestDiscoveryService_NodeDiscovery(t *testing.T) {
	_, service, transport := setupTestNetwork(t)
	defer func() {
		transport.Stop()
		service.Stop()
	}()

	// 새로운 노드 정보 생성
	newNodeInfo := &network.NodeInfo{
		ID:      "test-node-1",
		Address: "127.0.0.1:9002",
	}

	// 노드 정보 추가
	err := service.AddNode(newNodeInfo)
	require.NoError(t, err)

	// 노드가 발견되었는지 확인
	knownNodes := service.GetKnownNodes()
	assert.Len(t, knownNodes, 1, "Should have one known node")
	assert.Equal(t, newNodeInfo, knownNodes[newNodeInfo.ID], "Node info should match")

	// 노드 제거
	err = service.RemoveNode(newNodeInfo.ID)
	require.NoError(t, err)

	// 노드가 제거되었는지 확인
	knownNodes = service.GetKnownNodes()
	assert.Empty(t, knownNodes, "Known nodes should be empty after removal")
}

// 통합 테스트로 이동된 테스트들:
// - TestDiscoveryService_PeerConnection: 실제 피어 연결 및 핸드셰이크
// - TestDiscoveryService_HealthCheck: 피어 상태 확인 및 메시지 처리
// - TestDiscoveryService_MessageHandling: 메시지 송수신 처리
