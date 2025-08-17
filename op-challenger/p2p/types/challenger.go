package types

import (
	"crypto/ecdsa"
	"time"
)

// ChallengerRole represents the role capabilities of a challenger
type ChallengerRole int

const (
	ChallengerRoleChallenger ChallengerRole = 1 << iota // 챌린저 역할
	ChallengerRoleSequencer                             // 시퀀서 역할 (백업 시퀀서)
)

// 역할 조합 (비트마스크 방식으로 다중 역할 지원)
const (
	ChallengerRoleChallengerOnly = ChallengerRoleChallenger                           // 챌린저만 (현재 Phase 1)
	ChallengerRoleSequencerOnly  = ChallengerRoleSequencer                            // 시퀀서만 (향후 확장)
	ChallengerRoleBoth           = ChallengerRoleChallenger | ChallengerRoleSequencer // 둘 다 (Phase 2+ 백업 시퀀서)
)

// HasRole checks if challenger has specific role
func (role ChallengerRole) HasRole(checkRole ChallengerRole) bool {
	return role&checkRole != 0
}

// String returns string representation of the role
func (role ChallengerRole) String() string {
	switch role {
	case ChallengerRoleChallengerOnly:
		return "challenger"
	case ChallengerRoleSequencerOnly:
		return "sequencer"
	case ChallengerRoleBoth:
		return "challenger+sequencer"
	default:
		return "unknown"
	}
}

// ChallengerStatus represents the current status of a challenger
type ChallengerStatus int

const (
	ChallengerStatusOffline   ChallengerStatus = iota // 오프라인
	ChallengerStatusOnline                            // 온라인
	ChallengerStatusActive                            // 활성 (Phase 2+에서 사용)
	ChallengerStatusIdle                              // 유휴 (Phase 2+에서 사용)
	ChallengerStatusSuspended                         // 일시 정지 (Phase 2+에서 사용)
	ChallengerStatusBanned                            // 차단됨 (Phase 2+에서 사용)
)

// String returns string representation of the status
func (status ChallengerStatus) String() string {
	switch status {
	case ChallengerStatusOffline:
		return "offline"
	case ChallengerStatusOnline:
		return "online"
	case ChallengerStatusActive:
		return "active"
	case ChallengerStatusIdle:
		return "idle"
	case ChallengerStatusSuspended:
		return "suspended"
	case ChallengerStatusBanned:
		return "banned"
	default:
		return "unknown"
	}
}

// StakeStatus represents the staking status of a challenger
type StakeStatus int

const (
	StakeStatusNone        StakeStatus = iota // No stake
	StakeStatusPending                        // Stake pending
	StakeStatusActive                         // Stake active
	StakeStatusSlashing                       // Being slashed (Phase 2+)
	StakeStatusWithdrawing                    // Withdrawing stake
)

// String returns string representation of the stake status
func (status StakeStatus) String() string {
	switch status {
	case StakeStatusNone:
		return "none"
	case StakeStatusPending:
		return "pending"
	case StakeStatusActive:
		return "active"
	case StakeStatusSlashing:
		return "slashing"
	case StakeStatusWithdrawing:
		return "withdrawing"
	default:
		return "unknown"
	}
}

// ChallengerInfo represents basic information about a challenger (Phase 1 version)
type ChallengerInfo struct {
	// 기본 식별 정보
	ID        string           `json:"id"`         // 챌린저 고유 ID
	NodeID    string           `json:"node_id"`    // P2P 노드 ID
	Address   string           `json:"address"`    // 네트워크 주소
	PublicKey *ecdsa.PublicKey `json:"public_key"` // 공개키

	// 챌린저 특화 정보
	Role   ChallengerRole   `json:"role"`   // 챌린저 역할 (향후 백업 시퀀서 지원)
	Status ChallengerStatus `json:"status"` // 현재 상태

	// 스테이킹 정보 (Phase 1: 기본 존재 여부만)
	StakeAmount uint64      `json:"stake_amount"` // 스테이킹 금액 (Wei)
	StakeStatus StakeStatus `json:"stake_status"` // 스테이킹 상태

	// 기본 활동 정보 (Phase 1)
	LastSeen    time.Time `json:"last_seen"`    // 마지막 활동 시간
	ConnectedAt time.Time `json:"connected_at"` // 연결 시작 시간
	IsConnected bool      `json:"is_connected"` // 현재 연결 상태

	// 성과 정보 (Phase 1: 기본)
	TotalTests  uint64 `json:"total_tests"`  // 총 참여한 테스트 수
	PassedTests uint64 `json:"passed_tests"` // 통과한 테스트 수

	// 네트워크 정보 (Phase 1)
	ConnectedPeers []string      `json:"connected_peers"` // 연결된 피어들
	PeerCount      int           `json:"peer_count"`      // 연결된 피어 수
	NetworkLatency time.Duration `json:"network_latency"` // 네트워크 지연시간

	// 메타데이터
	Version   string    `json:"version"`   // 챌린저 버전
	Timestamp time.Time `json:"timestamp"` // 정보 업데이트 시간
}

// NewChallengerInfo creates a new ChallengerInfo with basic information
func NewChallengerInfo(id, nodeID, address string, publicKey *ecdsa.PublicKey) *ChallengerInfo {
	now := time.Now()
	return &ChallengerInfo{
		ID:             id,
		NodeID:         nodeID,
		Address:        address,
		PublicKey:      publicKey,
		Role:           ChallengerRoleChallenger, // Default role
		Status:         ChallengerStatusOffline,  // 초기 상태는 오프라인
		StakeAmount:    0,
		StakeStatus:    StakeStatusNone,
		LastSeen:       now,
		ConnectedAt:    time.Time{}, // 연결되지 않은 상태
		IsConnected:    false,
		TotalTests:     0,
		PassedTests:    0,
		ConnectedPeers: make([]string, 0),
		PeerCount:      0,
		NetworkLatency: 0,
		Version:        "1.0.0", // Phase 1 버전
		Timestamp:      now,
	}
}

// UpdateStatus updates the challenger status and timestamp
func (ci *ChallengerInfo) UpdateStatus(status ChallengerStatus) {
	ci.Status = status
	ci.LastSeen = time.Now()
	ci.Timestamp = time.Now()
}

// SetConnected marks the challenger as connected
func (ci *ChallengerInfo) SetConnected() {
	ci.IsConnected = true
	ci.ConnectedAt = time.Now()
	ci.UpdateStatus(ChallengerStatusOnline)
}

// SetDisconnected marks the challenger as disconnected
func (ci *ChallengerInfo) SetDisconnected() {
	ci.IsConnected = false
	ci.ConnectedAt = time.Time{}
	ci.UpdateStatus(ChallengerStatusOffline)
}

// UpdatePeerCount updates the number of connected peers
func (ci *ChallengerInfo) UpdatePeerCount(count int) {
	ci.PeerCount = count
	ci.Timestamp = time.Now()
}

// UpdateNetworkLatency updates the network latency
func (ci *ChallengerInfo) UpdateNetworkLatency(latency time.Duration) {
	ci.NetworkLatency = latency
	ci.Timestamp = time.Now()
}

// IsOnline returns true if the challenger is currently online
func (ci *ChallengerInfo) IsOnline() bool {
	return ci.Status == ChallengerStatusOnline && ci.IsConnected
}

// GetUptimeDuration returns how long the challenger has been connected
func (ci *ChallengerInfo) GetUptimeDuration() time.Duration {
	if !ci.IsConnected || ci.ConnectedAt.IsZero() {
		return 0
	}
	return time.Since(ci.ConnectedAt)
}

// IsStale checks if the challenger info is stale (not updated recently)
func (ci *ChallengerInfo) IsStale(staleDuration time.Duration) bool {
	return time.Since(ci.Timestamp) > staleDuration
}

// HasRole checks if the challenger has the specified role
func (ci *ChallengerInfo) HasRole(role ChallengerRole) bool {
	return (ci.Role & role) != 0
}
