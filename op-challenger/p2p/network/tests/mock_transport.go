package network

import (
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/network"
)

// MockTransport implements the Transport interface for testing
type MockTransport struct {
	connections map[string]*MockConnection
	mu          sync.RWMutex
	started     bool
	address     string
}

// MockConnection implements the network.Connection interface for testing
type MockConnection struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	closed     bool
	mu         sync.RWMutex
	readCount  int
	writeCount int
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		connections: make(map[string]*MockConnection),
	}
}

func (mt *MockTransport) Start(address string) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.started = true
	mt.address = address
	return nil
}

func (mt *MockTransport) Stop() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.started = false
	for _, conn := range mt.connections {
		conn.Close()
	}
	return nil
}

func (mt *MockTransport) Connect(address string) (network.Connection, error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if !mt.started {
		return nil, fmt.Errorf("transport not started")
	}

	// Create mock connection
	conn := &MockConnection{
		localAddr:  mockAddr(mt.address),
		remoteAddr: mockAddr(address),
	}

	mt.connections[address] = conn
	return conn, nil
}

func (mt *MockTransport) Accept() (network.Connection, error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if !mt.started {
		return nil, fmt.Errorf("transport not started")
	}

	// In real tests, we would wait for incoming connections
	// For now, just return error to simulate no incoming connections
	return nil, fmt.Errorf("no incoming connections")
}

func (mc *MockConnection) Read() ([]byte, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return nil, fmt.Errorf("connection closed")
	}

	mc.readCount++

	// 핸드셰이크 시뮬레이션
	if mc.readCount == 1 {
		// 고정된 유효한 P256 공개키 사용 (테스트용)
		// 이 값은 실제 P256 곡선 위의 유효한 점입니다
		validPublicKeyHex := "04" + // 압축되지 않은 점 표시
			"6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296" + // X 좌표
			"4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5" // Y 좌표

		// hex를 바이트로 변환
		publicKeyBytes := make([]byte, 65)
		for i := 0; i < 65; i++ {
			var b byte
			fmt.Sscanf(validPublicKeyHex[i*2:i*2+2], "%02x", &b)
			publicKeyBytes[i] = b
		}

		encodedKey := base64.StdEncoding.EncodeToString(publicKeyBytes)
		response := fmt.Sprintf(`{"accepted":true,"publicKey":"%s","version":"1.0.0"}`, encodedKey)
		return []byte(response), nil
	}

	// 이후 Read는 빈 데이터 반환
	return []byte{}, nil
}

func (mc *MockConnection) Write(data []byte) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.closed {
		return fmt.Errorf("connection closed")
	}

	mc.writeCount++
	// Write는 단순히 성공 처리 (실제 데이터 전송은 시뮬레이션)
	return nil
}

func (mc *MockConnection) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.closed = true
	return nil
}

func (mc *MockConnection) LocalAddr() net.Addr {
	return mc.localAddr
}

func (mc *MockConnection) RemoteAddr() net.Addr {
	return mc.remoteAddr
}

type mockAddr string

func (ma mockAddr) Network() string { return "mock" }
func (ma mockAddr) String() string  { return string(ma) }

func (mc *MockConnection) SetDeadline(t time.Time) error {
	return nil // 테스트용이므로 데드라인 설정은 무시
}

func (mc *MockConnection) SetReadDeadline(t time.Time) error {
	return nil // 테스트용이므로 데드라인 설정은 무시
}

func (mc *MockConnection) SetWriteDeadline(t time.Time) error {
	return nil // 테스트용이므로 데드라인 설정은 무시
}
