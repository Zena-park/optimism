package network_test

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestPeerSharing_Integration(t *testing.T) {
	t.Skip("Integration test - requires actual network setup")

	// 통합 테스트로 이동됨: 실제 피어 간 정보 공유 테스트
	logger := log.New()

	// 3개의 노드 생성
	node1, err := network.NewP2PNode(&network.P2PConfig{
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
	}, logger)
	require.NoError(t, err)

	node2, err := network.NewP2PNode(&network.P2PConfig{
		Address:         "127.0.0.1:9002",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9002,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{"127.0.0.1:9001"},
		},
	}, logger)
	require.NoError(t, err)

	node3, err := network.NewP2PNode(&network.P2PConfig{
		Address:         "127.0.0.1:9003",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9003,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{"127.0.0.1:9001"},
		},
	}, logger)
	require.NoError(t, err)

	// 노드들 시작
	err = node1.Start()
	require.NoError(t, err)
	defer node1.Stop()

	err = node2.Start()
	require.NoError(t, err)
	defer node2.Stop()

	err = node3.Start()
	require.NoError(t, err)
	defer node3.Stop()

	// node2를 node1에 연결 및 핸드셰이크 수행
	peer2, err := node2.ConnectAndHandshake(node1.GetAddress())
	require.NoError(t, err)
	err = node2.AddPeer(peer2)
	require.NoError(t, err)

	// node1의 피어 목록 확인
	peers1 := node1.GetPeers()
	require.Equal(t, 1, len(peers1), "node1 should have 1 peer")

	// node3를 node1에 연결 및 핸드셰이크 수행
	peer3, err := node3.ConnectAndHandshake(node1.GetAddress())
	require.NoError(t, err)
	err = node3.AddPeer(peer3)
	require.NoError(t, err)

	// node1의 피어 목록 다시 확인
	peers1 = node1.GetPeers()
	require.Equal(t, 2, len(peers1), "node1 should have 2 peers")

	// node1의 피어 정보 요청
	nodes, err := node2.GetDiscovery().FindPeers("")
	require.NoError(t, err)
	require.NotEmpty(t, nodes, "should get peer info from node1")

	// node2가 node3를 발견했는지 확인
	time.Sleep(1 * time.Second) // 피어 정보 공유 대기
	peers2 := node2.GetPeers()
	require.Equal(t, 2, len(peers2), "node2 should discover node3 through node1")

	// 각 노드의 피어 정보 출력
	t.Logf("Node1 peers: %d", len(node1.GetPeers()))
	t.Logf("Node2 peers: %d", len(node2.GetPeers()))
	t.Logf("Node3 peers: %d", len(node3.GetPeers()))

	// 모든 노드가 서로를 알고 있는지 확인
	require.Equal(t, 2, len(node1.GetPeers()), "node1 should know about node2 and node3")
	require.Equal(t, 2, len(node2.GetPeers()), "node2 should know about node1 and node3")
	require.Equal(t, 2, len(node3.GetPeers()), "node3 should know about node1 and node2")
}

func TestPeerSharingWithDisconnection_Integration(t *testing.T) {
	t.Skip("Integration test - requires actual network setup")

	// 통합 테스트로 이동됨: 실제 피어 연결 해제 및 재연결 테스트
	logger := log.New()

	// 3개의 노드 생성 (설정은 위와 동일)
	node1, err := network.NewP2PNode(&network.P2PConfig{
		Address:         "127.0.0.1:9011",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9011,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{},
		},
	}, logger)
	require.NoError(t, err)

	node2, err := network.NewP2PNode(&network.P2PConfig{
		Address:         "127.0.0.1:9012",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9012,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{"127.0.0.1:9011"},
		},
	}, logger)
	require.NoError(t, err)

	node3, err := network.NewP2PNode(&network.P2PConfig{
		Address:         "127.0.0.1:9013",
		PrivateKeyPath:  "",
		MaxPeers:        10,
		DiscoveryPort:   9013,
		TransportConfig: network.DefaultTransportConfig(),
		DiscoveryConfig: network.DiscoveryConfig{
			K:         20,
			Alpha:     3,
			Timeout:   5 * time.Second,
			Bootstrap: []string{"127.0.0.1:9011"},
		},
	}, logger)
	require.NoError(t, err)

	// 노드들 시작
	err = node1.Start()
	require.NoError(t, err)
	defer node1.Stop()

	err = node2.Start()
	require.NoError(t, err)
	defer node2.Stop()

	err = node3.Start()
	require.NoError(t, err)
	defer node3.Stop()

	// 모든 노드 연결 및 핸드셰이크 수행
	peer2, err := node2.ConnectAndHandshake(node1.GetAddress())
	require.NoError(t, err)
	err = node2.AddPeer(peer2)
	require.NoError(t, err)

	peer3, err := node3.ConnectAndHandshake(node1.GetAddress())
	require.NoError(t, err)
	err = node3.AddPeer(peer3)
	require.NoError(t, err)

	// 피어 정보 공유 대기
	time.Sleep(1 * time.Second)

	// 초기 상태 확인
	require.Equal(t, 2, len(node1.GetPeers()), "node1 should have 2 peers initially")
	require.Equal(t, 2, len(node2.GetPeers()), "node2 should have 2 peers initially")
	require.Equal(t, 2, len(node3.GetPeers()), "node3 should have 2 peers initially")

	// node2 연결 해제
	for _, peer := range node2.GetPeers() {
		err = node2.RemovePeer(peer.ID)
		require.NoError(t, err)
	}

	// 연결 해제 후 상태 확인
	require.Equal(t, 0, len(node2.GetPeers()), "node2 should have no peers after disconnection")
	require.Equal(t, 1, len(node1.GetPeers()), "node1 should have 1 peer after node2 disconnection")
	require.Equal(t, 1, len(node3.GetPeers()), "node3 should have 1 peer after node2 disconnection")

	// node2 재연결 및 핸드셰이크 수행
	peer2New, err := node2.ConnectAndHandshake(node1.GetAddress())
	require.NoError(t, err)
	err = node2.AddPeer(peer2New)
	require.NoError(t, err)

	// 피어 정보 공유 대기
	time.Sleep(1 * time.Second)

	// 재연결 후 상태 확인
	require.Equal(t, 2, len(node1.GetPeers()), "node1 should have 2 peers after reconnection")
	require.Equal(t, 2, len(node2.GetPeers()), "node2 should have 2 peers after reconnection")
	require.Equal(t, 2, len(node3.GetPeers()), "node3 should have 2 peers after reconnection")
}
