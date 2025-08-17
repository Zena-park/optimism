# TODO 1.2: 챌린저 관리 시스템 설계

## 📋 작업 개요
- **작업명**: 챌린저 관리 시스템 설계
- **상태**: ✅ 완료
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
기존 P2P 인프라를 기반으로 챌린저 특화 관리 시스템을 설계한다. 챌린저 식별, 등록, 상태 관리, 발견 및 연결 시스템을 포함한다.

## 📐 설계 요구사항

### 1. 챌린저 식별 및 등록 시스템
- 챌린저 고유 ID 생성 방식
- 챌린저 메타데이터 구조 정의
- 챌린저 상태 관리 (온라인/오프라인/활성/비활성)

### 2. 챌린저 발견 및 연결 시스템
- 기존 discovery 시스템 확장
- 챌린저 네트워크 토폴로지 설계
- 연결 유지 및 관리 전략

## 🏗️ 시스템 설계

### 1. 챌린저 메타데이터 구조

#### **1.1 ChallengerInfo 구조체**
```go
// ChallengerInfo represents detailed information about a challenger
type ChallengerInfo struct {
    // 기본 식별 정보
    ID          string             `json:"id"`           // 챌린저 고유 ID
    NodeID      string             `json:"node_id"`      // P2P 노드 ID
    Address     string             `json:"address"`      // 네트워크 주소
    PublicKey   *ecdsa.PublicKey   `json:"public_key"`   // 공개키

        // 챌린저 특화 정보
    Role        ChallengerRole     `json:"role"`         // 챌린저 역할 (향후 백업 시퀀서 지원)
    Status      ChallengerStatus   `json:"status"`       // 현재 상태
    Reputation  ReputationScore    `json:"reputation"`   // 평판 점수

    // 스테이킹 정보
    StakeAmount uint64             `json:"stake_amount"` // 스테이킹 금액 (Wei)
    StakeStatus StakeStatus        `json:"stake_status"` // 스테이킹 상태

    // 활동 정보
    LastSeen       time.Time       `json:"last_seen"`        // 마지막 활동 시간
    LastValidation time.Time       `json:"last_validation"`  // 마지막 검증 시간
    ActiveGames    []string        `json:"active_games"`     // 참여 중인 게임들

    // 성과 정보
    TotalTests     uint64          `json:"total_tests"`      // 총 테스트 수
    PassedTests    uint64          `json:"passed_tests"`     // 통과한 테스트 수
    FailedTests    uint64          `json:"failed_tests"`     // 실패한 테스트 수
    ResponseTime   time.Duration   `json:"response_time"`    // 평균 응답 시간

    // 네트워크 정보
    ConnectedPeers []string        `json:"connected_peers"`  // 연결된 피어들
    NetworkLatency time.Duration   `json:"network_latency"`  // 네트워크 지연시간

    // 메타데이터
    Version    string             `json:"version"`          // 챌린저 버전
    Timestamp  time.Time          `json:"timestamp"`        // 정보 업데이트 시간
}
```

#### **1.2 챌린저 역할 및 상태 정의**
```go
// ChallengerRole represents the role capabilities of a challenger
type ChallengerRole int

const (
    ChallengerRoleChallenger ChallengerRole = 1 << iota // 챌린저 역할
    ChallengerRoleSequencer                             // 시퀀서 역할 (백업 시퀀서)
)

// 역할 조합 (비트마스크 방식으로 다중 역할 지원)
const (
    ChallengerRoleChallengerOnly = ChallengerRoleChallenger                        // 챌린저만 (현재 Phase 1)
    ChallengerRoleSequencerOnly  = ChallengerRoleSequencer                        // 시퀀서만 (향후 확장)
    ChallengerRoleBoth          = ChallengerRoleChallenger | ChallengerRoleSequencer // 둘 다 (Phase 2+ 백업 시퀀서)
)

// HasRole checks if challenger has specific role
func (role ChallengerRole) HasRole(checkRole ChallengerRole) bool {
    return role&checkRole != 0
}

// ChallengerStatus represents the current status of a challenger
type ChallengerStatus int

const (
    ChallengerStatusOffline   ChallengerStatus = iota // 오프라인
    ChallengerStatusOnline                            // 온라인
    ChallengerStatusActive                            // 활성 (게임 참여중)
    ChallengerStatusIdle                              // 유휴 상태
    ChallengerStatusSuspended                         // 일시 정지
    ChallengerStatusBanned                            // 차단됨
)

// StakeStatus represents the staking status
type StakeStatus int

const (
    StakeStatusNone      StakeStatus = iota // 스테이킹 없음
    StakeStatusPending                      // 스테이킹 대기중
    StakeStatusActive                       // 스테이킹 활성
    StakeStatusSlashing                     // 슬래싱 중
    StakeStatusWithdrawing                  // 인출 중
)

// ReputationScore represents reputation scoring
type ReputationScore struct {
    Overall    float64   `json:"overall"`     // 전체 평판 점수 (0-100)
    Accuracy   float64   `json:"accuracy"`    // 정확도 점수
    Speed      float64   `json:"speed"`       // 응답 속도 점수
    Reliability float64  `json:"reliability"` // 신뢰성 점수
    LastUpdate time.Time `json:"last_update"` // 마지막 업데이트
}
```

### 2. 챌린저 네트워크 관리자 (ChallengerNetworkManager)

#### **2.1 ChallengerNetworkManager 구조체**
```go
// ChallengerNetworkManager manages all challenger-related operations in P2P network
type ChallengerNetworkManager struct {
    // 기본 정보
    nodeID     string                        // 현재 노드 ID
    node       *network.P2PNode             // P2P 노드 참조

    // 챌린저 관리
    challengers    map[string]*ChallengerInfo // 알려진 챌린저들
    localInfo      *ChallengerInfo            // 로컬 챌린저 정보

    // 서비스 관리
    discovery      *ChallengerDiscovery       // 챌린저 발견 서비스
    registry       *ChallengerRegistry        // 챌린저 등록 서비스
    monitor        *ChallengerMonitor         // 챌린저 모니터링

    // 동기화 및 상태 관리
    mu             sync.RWMutex               // 동시성 제어
    ctx            context.Context            // 컨텍스트
    cancel         context.CancelFunc         // 취소 함수

    // 설정
    config         *ChallengerNetworkManagerConfig   // 설정
    logger         log.Logger                 // 로거
}

// ChallengerNetworkManagerConfig contains configuration for challenger network manager
type ChallengerNetworkManagerConfig struct {
    // 기본 설정
    MinStakeAmount    uint64        `json:"min_stake_amount"`     // 최소 스테이킹 금액
    MaxPeers          int           `json:"max_peers"`            // 최대 피어 수
    HeartbeatInterval time.Duration `json:"heartbeat_interval"`   // 하트비트 간격

    // 평판 설정
    InitialReputation float64       `json:"initial_reputation"`   // 초기 평판 점수
    MinReputation     float64       `json:"min_reputation"`       // 최소 평판 점수

    // 네트워크 설정
    DiscoveryInterval time.Duration `json:"discovery_interval"`   // 발견 간격
    ConnectionTimeout time.Duration `json:"connection_timeout"`   // 연결 타임아웃

    // 보안 설정
    RequireStaking    bool          `json:"require_staking"`      // 스테이킹 필수 여부
    EnableReputation  bool          `json:"enable_reputation"`    // 평판 시스템 활성화
}
```

#### **2.2 주요 기능 메서드**
```go
// NewChallengerNetwork creates a new challenger network
func NewChallengerNetwork(
    node *network.P2PNode,
    config *ChallengerNetworkConfig,
    logger log.Logger,
) (*ChallengerNetwork, error)

// Start starts the challenger network
func (cn *ChallengerNetwork) Start() error

// Stop stops the challenger network
func (cn *ChallengerNetwork) Stop() error

// RegisterChallenger registers a new challenger
func (cn *ChallengerNetwork) RegisterChallenger(info *ChallengerInfo) error

// UnregisterChallenger unregisters a challenger
func (cn *ChallengerNetwork) UnregisterChallenger(challengerID string) error

// GetChallenger retrieves challenger information
func (cn *ChallengerNetwork) GetChallenger(challengerID string) (*ChallengerInfo, error)

// GetAllChallengers returns all known challengers
func (cn *ChallengerNetwork) GetAllChallengers() map[string]*ChallengerInfo

// UpdateChallengerStatus updates challenger status
func (cn *ChallengerNetwork) UpdateChallengerStatus(
    challengerID string,
    status ChallengerStatus,
) error

// UpdateReputation updates challenger reputation
func (cn *ChallengerNetwork) UpdateReputation(
    challengerID string,
    score ReputationScore,
) error

// SelectChallengers selects challengers based on criteria
func (cn *ChallengerNetwork) SelectChallengers(
    criteria *SelectionCriteria,
) ([]*ChallengerInfo, error)
```

### 3. 챌린저 발견 서비스 (ChallengerDiscovery)

#### **3.1 ChallengerDiscovery 구조체**
```go
// ChallengerDiscovery handles challenger discovery and announcement
type ChallengerDiscovery struct {
    network    *ChallengerNetwork     // 챌린저 네트워크 참조
    node       *network.P2PNode      // P2P 노드 참조

    // 발견 관련
    knownNodes map[string]*network.NodeInfo // 알려진 노드들
    bootstrap  []string                     // 부트스트랩 챌린저들

    // 상태 관리
    mu         sync.RWMutex              // 동시성 제어
    running    bool                      // 실행 상태

    config     *DiscoveryConfig          // 설정
    logger     log.Logger                // 로거
}

// 주요 메서드들
func (cd *ChallengerDiscovery) Start() error
func (cd *ChallengerDiscovery) Stop() error
func (cd *ChallengerDiscovery) DiscoverChallengers() ([]*ChallengerInfo, error)
func (cd *ChallengerDiscovery) AnnouncePresence() error
func (cd *ChallengerDiscovery) HandleChallengerAnnouncement(msg *network.Message) error
```

### 4. 챌린저 등록 서비스 (ChallengerRegistry)

#### **4.1 ChallengerRegistry 구조체**
```go
// ChallengerRegistry handles challenger registration and validation
type ChallengerRegistry struct {
    network     *ChallengerNetwork       // 챌린저 네트워크 참조

    // 등록 관리
    registrations map[string]*RegistrationRequest // 등록 요청들
    validators    []RegistrationValidator          // 등록 검증자들

    // 상태 관리
    mu          sync.RWMutex                      // 동시성 제어

    config      *RegistryConfig                   // 설정
    logger      log.Logger                        // 로거
}

// RegistrationRequest represents a challenger registration request
type RegistrationRequest struct {
    ChallengerInfo *ChallengerInfo `json:"challenger_info"`
    Signature      []byte          `json:"signature"`
    Timestamp      time.Time       `json:"timestamp"`
    Status         RequestStatus   `json:"status"`
}

// RequestStatus represents registration request status
type RequestStatus int

const (
    RequestStatusPending   RequestStatus = iota
    RequestStatusApproved
    RequestStatusRejected
    RequestStatusExpired
)
```

### 5. 챌린저 모니터링 (ChallengerMonitor)

#### **5.1 ChallengerMonitor 구조체**
```go
// ChallengerMonitor monitors challenger health and performance
type ChallengerMonitor struct {
    network     *ChallengerNetwork    // 챌린저 네트워크 참조

    // 모니터링 데이터
    healthData  map[string]*HealthData // 헬스 데이터
    metrics     *MonitoringMetrics     // 모니터링 메트릭스

    // 상태 관리
    mu          sync.RWMutex           // 동시성 제어
    running     bool                   // 실행 상태

    config      *MonitorConfig         // 설정
    logger      log.Logger             // 로거
}

// HealthData represents challenger health information
type HealthData struct {
    ChallengerID   string        `json:"challenger_id"`
    LastHeartbeat  time.Time     `json:"last_heartbeat"`
    ResponseTime   time.Duration `json:"response_time"`
    ErrorCount     uint64        `json:"error_count"`
    Status         HealthStatus  `json:"status"`
}

// HealthStatus represents health status
type HealthStatus int

const (
    HealthStatusHealthy   HealthStatus = iota
    HealthStatusDegraded
    HealthStatusUnhealthy
    HealthStatusUnknown
)
```

## 🔄 시스템 통합 설계

### 1. 기존 P2P 시스템과의 통합
```go
// 기존 P2PNode 확장
func (node *P2PNode) SetChallengerNetwork(cn *ChallengerNetwork) {
    node.challengerNetwork = cn
}

// 기존 메시지 핸들링 확장
func (node *P2PNode) handleChallengerMessage(msg *Message) error {
    if node.challengerNetwork != nil {
        return node.challengerNetwork.HandleMessage(msg)
    }
    return nil
}
```

### 2. 메시지 라우팅
```go
// ChallengerMessageRouter handles challenger-specific messages
type ChallengerMessageRouter struct {
    network *ChallengerNetwork
}

func (cmr *ChallengerMessageRouter) RouteMessage(msg *network.Message) error {
    switch msg.Type {
    case network.MessageTypeChallengerRegistration:
        return cmr.handleRegistration(msg)
    case network.MessageTypeChallengerStatus:
        return cmr.handleStatusUpdate(msg)
    case network.MessageTypeAttentionTestTriggered:
        return cmr.handleAttentionTest(msg)
    // ... 기타 메시지 타입들
    }
    return nil
}
```

## 📊 성능 및 확장성 고려사항

### 1. 메모리 사용량 최적화
- 챌린저 정보 캐싱 전략
- LRU 캐시를 통한 메모리 관리
- 주기적 가비지 컬렉션

### 2. 네트워크 효율성
- 배치 메시지 처리
- 압축을 통한 대역폭 절약
- 적응형 하트비트 간격

### 3. 확장성 설계
- 수평적 확장 지원
- 샤딩을 통한 부하 분산
- 동적 로드 밸런싱

## ✅ 설계 체크리스트
- [x] 챌린저 메타데이터 구조 정의
- [x] ChallengerNetwork 구조체 설계
- [x] ChallengerDiscovery 서비스 설계
- [x] ChallengerRegistry 서비스 설계
- [x] ChallengerMonitor 서비스 설계
- [x] 기존 P2P 시스템과의 통합 방안
- [x] 메시지 라우팅 시스템 설계
- [x] 성능 및 확장성 고려사항 검토

## 🚀 다음 단계
1. **TODO 2.1**: 챌린저 노드 확장 - 실제 구현 시작
2. **TODO 2.2**: 챌린저 발견 및 연결 - 구체적 구현
3. **TODO 2.3**: 챌린저 상태 관리 - 모니터링 시스템 구현

## 📝 설계 결론
챌린저 관리 시스템은 기존 P2P 인프라를 최대한 활용하면서도 챌린저 특화 기능을 제공하도록 설계되었습니다. 모듈화된 구조로 각 컴포넌트가 독립적으로 개발 및 테스트될 수 있으며, 확장성과 성능을 고려한 설계입니다.
