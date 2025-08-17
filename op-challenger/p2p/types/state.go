package types

import (
	"time"
)

// ChallengerState represents basic state of a challenger (Phase 1 version)
type ChallengerState struct {
	// 기본 정보
	ChallengerID string `json:"challenger_id"`
	NodeID       string `json:"node_id"`
	Address      string `json:"address"`

	// 기본 상태 정보 (Phase 1)
	Status      ChallengerStatus `json:"status"`       // 현재 상태
	IsOnline    bool             `json:"is_online"`    // 온라인 여부
	IsConnected bool             `json:"is_connected"` // 연결 여부

	// 기본 연결 정보 (Phase 1)
	ConnectionInfo *ConnectionInfo `json:"connection_info"`

	// 메타데이터
	Version     string    `json:"version"`      // 상태 버전
	LastUpdated time.Time `json:"last_updated"` // 마지막 업데이트
	Sequence    uint64    `json:"sequence"`     // 업데이트 시퀀스 번호
}

// ConnectionInfo represents basic connection information
type ConnectionInfo struct {
	// 연결 상태
	ConnectedAt    time.Time     `json:"connected_at"`    // 연결 시작 시간
	LastSeen       time.Time     `json:"last_seen"`       // 마지막 확인 시간
	UptimeDuration time.Duration `json:"uptime_duration"` // 업타임

	// 네트워크 정보
	PeerCount      int           `json:"peer_count"`      // 연결된 피어 수
	NetworkLatency time.Duration `json:"network_latency"` // 네트워크 지연시간

	// 연결 품질 (Phase 1 기본)
	ConnectionQuality ConnectionQuality `json:"connection_quality"` // 연결 품질
}

// ConnectionQuality represents the quality of network connection
type ConnectionQuality int

const (
	ConnectionQualityUnknown   ConnectionQuality = iota // 알 수 없음
	ConnectionQualityPoor                               // 나쁨
	ConnectionQualityFair                               // 보통
	ConnectionQualityGood                               // 좋음
	ConnectionQualityExcellent                          // 매우 좋음
)

// String returns string representation of connection quality
func (cq ConnectionQuality) String() string {
	switch cq {
	case ConnectionQualityPoor:
		return "poor"
	case ConnectionQualityFair:
		return "fair"
	case ConnectionQualityGood:
		return "good"
	case ConnectionQualityExcellent:
		return "excellent"
	default:
		return "unknown"
	}
}

// NewChallengerState creates a new ChallengerState with basic information
func NewChallengerState(challengerID, nodeID, address string) *ChallengerState {
	now := time.Now()
	return &ChallengerState{
		ChallengerID: challengerID,
		NodeID:       nodeID,
		Address:      address,
		Status:       ChallengerStatusOffline,
		IsOnline:     false,
		IsConnected:  false,
		ConnectionInfo: &ConnectionInfo{
			ConnectedAt:       time.Time{},
			LastSeen:          now,
			UptimeDuration:    0,
			PeerCount:         0,
			NetworkLatency:    0,
			ConnectionQuality: ConnectionQualityUnknown,
		},
		Version:     "1.0.0",
		LastUpdated: now,
		Sequence:    0,
	}
}

// UpdateStatus updates the challenger status
func (cs *ChallengerState) UpdateStatus(status ChallengerStatus) {
	cs.Status = status
	cs.LastUpdated = time.Now()
	cs.Sequence++

	// 상태에 따른 온라인/연결 상태 업데이트
	switch status {
	case ChallengerStatusOnline:
		cs.IsOnline = true
		cs.IsConnected = true
		if cs.ConnectionInfo.ConnectedAt.IsZero() {
			cs.ConnectionInfo.ConnectedAt = time.Now()
		}
	case ChallengerStatusOffline:
		cs.IsOnline = false
		cs.IsConnected = false
		cs.ConnectionInfo.ConnectedAt = time.Time{}
		cs.ConnectionInfo.UptimeDuration = 0
	}

	cs.ConnectionInfo.LastSeen = time.Now()
}

// SetOnline marks the challenger as online and connected
func (cs *ChallengerState) SetOnline() {
	cs.UpdateStatus(ChallengerStatusOnline)
}

// SetOffline marks the challenger as offline and disconnected
func (cs *ChallengerState) SetOffline() {
	cs.UpdateStatus(ChallengerStatusOffline)
}

// UpdateConnectionInfo updates connection-related information
func (cs *ChallengerState) UpdateConnectionInfo(peerCount int, latency time.Duration) {
	cs.ConnectionInfo.PeerCount = peerCount
	cs.ConnectionInfo.NetworkLatency = latency
	cs.ConnectionInfo.LastSeen = time.Now()

	// 업타임 계산
	if cs.IsConnected && !cs.ConnectionInfo.ConnectedAt.IsZero() {
		cs.ConnectionInfo.UptimeDuration = time.Since(cs.ConnectionInfo.ConnectedAt)
	}

	// 연결 품질 계산 (간단한 로직)
	cs.ConnectionInfo.ConnectionQuality = cs.calculateConnectionQuality(latency, peerCount)

	cs.LastUpdated = time.Now()
	cs.Sequence++
}

// calculateConnectionQuality calculates connection quality based on latency and peer count
func (cs *ChallengerState) calculateConnectionQuality(latency time.Duration, peerCount int) ConnectionQuality {
	// 간단한 품질 계산 로직 (Phase 1)
	if latency == 0 || peerCount == 0 {
		return ConnectionQualityUnknown
	}

	// 지연시간 기반 품질 평가
	switch {
	case latency < 50*time.Millisecond && peerCount >= 10:
		return ConnectionQualityExcellent
	case latency < 100*time.Millisecond && peerCount >= 5:
		return ConnectionQualityGood
	case latency < 200*time.Millisecond && peerCount >= 3:
		return ConnectionQualityFair
	default:
		return ConnectionQualityPoor
	}
}

// GetUptime returns the current uptime duration
func (cs *ChallengerState) GetUptime() time.Duration {
	if !cs.IsConnected || cs.ConnectionInfo.ConnectedAt.IsZero() {
		return 0
	}
	return time.Since(cs.ConnectionInfo.ConnectedAt)
}

// IsHealthy returns true if the challenger is in a healthy state
func (cs *ChallengerState) IsHealthy() bool {
	return cs.IsOnline && cs.IsConnected && cs.ConnectionInfo.ConnectionQuality >= ConnectionQualityFair
}

// IsStale checks if the state is stale (not updated recently)
func (cs *ChallengerState) IsStale(staleDuration time.Duration) bool {
	return time.Since(cs.LastUpdated) > staleDuration
}

// Clone creates a deep copy of the ChallengerState
func (cs *ChallengerState) Clone() *ChallengerState {
	clone := *cs
	if cs.ConnectionInfo != nil {
		connInfo := *cs.ConnectionInfo
		clone.ConnectionInfo = &connInfo
	}
	return &clone
}

// StateUpdate represents an update to challenger state
type StateUpdate struct {
	ChallengerID string                 `json:"challenger_id"`
	UpdateType   StateUpdateType        `json:"update_type"`
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	Sequence     uint64                 `json:"sequence"`
}

// StateUpdateType represents the type of state update
type StateUpdateType int

const (
	StateUpdateTypeStatus     StateUpdateType = iota // 상태 업데이트
	StateUpdateTypeConnection                        // 연결 정보 업데이트
	StateUpdateTypeHeartbeat                         // 하트비트
)

// String returns string representation of update type
func (sut StateUpdateType) String() string {
	switch sut {
	case StateUpdateTypeStatus:
		return "status"
	case StateUpdateTypeConnection:
		return "connection"
	case StateUpdateTypeHeartbeat:
		return "heartbeat"
	default:
		return "unknown"
	}
}

// NewStateUpdate creates a new state update
func NewStateUpdate(challengerID string, updateType StateUpdateType, data map[string]interface{}) *StateUpdate {
	return &StateUpdate{
		ChallengerID: challengerID,
		UpdateType:   updateType,
		Data:         data,
		Timestamp:    time.Now(),
		Sequence:     0, // 실제 시퀀스는 상태 관리자에서 할당
	}
}
