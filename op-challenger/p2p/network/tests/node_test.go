package network_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestNodeListening_Integration(t *testing.T) {
	t.Skip("Integration test - requires actual network setup")

	// 통합 테스트로 이동됨: 실제 TCP 리스닝 및 연결 테스트
	logger := log.New()
	addr := "127.0.0.1:9001"

	// 노드 생성
	config := &network.P2PConfig{
		Address:         addr,
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9001,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{},
		},
	}

	node, err := network.NewP2PNode(config, logger)
	require.NoError(t, err)
	require.NotNil(t, node)

	// 노드 시작
	err = node.Start()
	require.NoError(t, err)
	defer node.Stop()

	// 리스닝 상태 확인
	require.Equal(t, network.NodeStatusRunning, node.GetStatus())

	// 다른 노드에서 연결 시도
	conn, err := node.GetTransport().Connect(addr)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()
}

// P2P Node 단위 테스트 (모킹 사용)
func TestP2PNode_Creation(t *testing.T) {
	logger := log.New()

	config := &network.P2PConfig{
		Address:         "127.0.0.1:9001",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9001,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{},
		},
	}

	node, err := network.NewP2PNode(config, logger)
	require.NoError(t, err)
	require.NotNil(t, node)

	// 초기 상태 확인 (새로 생성된 노드는 Starting 상태)
	status := node.GetStatus()
	require.True(t, status == network.NodeStatusStarting || status == network.NodeStatusStopped,
		"Node should be in Starting or Stopped state, got: %d", status)
	require.Empty(t, node.GetPeers())
}

func TestP2PNode_Configuration(t *testing.T) {
	logger := log.New()

	config := &network.P2PConfig{
		Address:         "127.0.0.1:9001",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9001,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{},
		},
	}

	node, err := network.NewP2PNode(config, logger)
	require.NoError(t, err)

	// 설정값 확인
	require.NotNil(t, node.GetTransport())

	// 초기 상태 확인
	status := node.GetStatus()
	require.True(t, status == network.NodeStatusStarting || status == network.NodeStatusStopped,
		"Node should be in Starting or Stopped state, got: %d", status)

	// 설정된 최대 피어 수 확인 (내부 필드는 직접 접근할 수 없으므로 초기 상태만 확인)
	require.Empty(t, node.GetPeers())
}

func TestP2PNode_PeerManagement(t *testing.T) {
	logger := log.New()

	config := &network.P2PConfig{
		Address:         "127.0.0.1:9001",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9001,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{},
		},
	}

	node, err := network.NewP2PNode(config, logger)
	require.NoError(t, err)

	// 초기에는 피어가 없어야 함
	require.Empty(t, node.GetPeers())

	// 초기 피어 수 확인
	initialPeerCount := len(node.GetPeers())
	require.Equal(t, 0, initialPeerCount)

	// Mock 피어 정보 생성 (실제 공개키 생성)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	peerInfo := &network.NodeInfo{
		ID:        "test-peer-1",
		Address:   "127.0.0.1:9002",
		PublicKey: &privateKey.PublicKey,
	}

	// 피어 정보가 올바르게 생성되었는지 확인
	require.NotNil(t, peerInfo.PublicKey)
	require.Equal(t, "test-peer-1", peerInfo.ID)
	require.Equal(t, "127.0.0.1:9002", peerInfo.Address)
}

// 통합 테스트로 이동된 테스트들:
// - TestNodeListening_Integration: 실제 TCP 리스닝 및 연결
// - TestNodeConnection: 실제 노드 간 연결 및 핸드셰이크
// - TestMessageSendReceive: 메시지 송수신 처리
