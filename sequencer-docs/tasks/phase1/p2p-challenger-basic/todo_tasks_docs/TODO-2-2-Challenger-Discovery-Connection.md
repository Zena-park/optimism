# TODO 2.2: 챌린저 발견 및 연결

## 📋 작업 개요
- **작업명**: 챌린저 발견 및 연결
- **상태**: 🔄 진행중
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
기존 discovery 시스템을 확장하여 챌린저 특화 발견 메커니즘을 구현한다. 네트워크 부트스트래핑, 자동 연결/재연결, 연결 품질 모니터링, 연결 풀 관리 시스템을 포함한다.

## 📐 구현 요구사항

### 1. 챌린저 발견 프로토콜 구현
- 기존 discovery 시스템 확장
- 챌린저 특화 발견 메커니즘
- 네트워크 부트스트래핑

### 2. 연결 관리 시스템
- 자동 연결 및 재연결
- 연결 품질 모니터링
- 연결 풀 관리

## 🏗️ 시스템 설계

### 1. 챌린저 발견 시스템 (ChallengerDiscovery)

#### **1.1 확장된 ChallengerDiscovery 구조체**
```go
// ChallengerDiscovery handles challenger-specific discovery and networking
type ChallengerDiscovery struct {
    // 기본 정보
    nodeID          string
    challengerID    string
    node           *network.P2PNode
    challengerNode *ChallengerP2PNode

    // 발견 관리
    baseDiscovery  *network.DiscoveryService  // 기존 discovery 서비스
    dht           *ChallengerDHT              // 챌린저 특화 DHT

    // 챌린저 정보 관리
    knownChallengers    map[string]*ChallengerInfo     // 알려진 챌린저들
    challengerNodes     map[string]*network.NodeInfo   // 챌린저 노드 정보
    bootstrapNodes      []string                       // 부트스트랩 챌린저들

    // 발견 상태 관리
    discoveryState      DiscoveryState
    lastDiscovery       time.Time
    discoveryErrors     uint64

    // 연결 관리
    connectionManager   *ChallengerConnectionManager

    // 설정 및 제어
    config             *ChallengerDiscoveryConfig
    running            bool

    // 동시성 제어
    mu                 sync.RWMutex
    ctx                context.Context
    cancel             context.CancelFunc

    // 이벤트 및 로깅
    eventBus           *EventBus
    logger             log.Logger
}

// DiscoveryState represents the current state of discovery process
type DiscoveryState int

const (
    DiscoveryStateIdle       DiscoveryState = iota // 유휴 상태
    DiscoveryStateDiscovering                      // 발견 중
    DiscoveryStateConnecting                       // 연결 중
    DiscoveryStateError                            // 오류 상태
)

// ChallengerDiscoveryConfig contains configuration for challenger discovery
type ChallengerDiscoveryConfig struct {
    // 발견 설정
    DiscoveryInterval    time.Duration `json:"discovery_interval"`     // 발견 간격 (기본: 30초)
    DiscoveryTimeout     time.Duration `json:"discovery_timeout"`      // 발견 타임아웃 (기본: 10초)
    MaxDiscoveryRetries  int           `json:"max_discovery_retries"`   // 최대 재시도 (기본: 3)

    // 부트스트랩 설정
    BootstrapNodes       []string      `json:"bootstrap_nodes"`         // 부트스트랩 노드들
    BootstrapTimeout     time.Duration `json:"bootstrap_timeout"`       // 부트스트랩 타임아웃
    MinBootstrapPeers    int           `json:"min_bootstrap_peers"`     // 최소 부트스트랩 피어 수

    // 연결 설정
    MaxPeers             int           `json:"max_peers"`               // 최대 피어 수 (기본: 50)
    MinPeers             int           `json:"min_peers"`               // 최소 피어 수 (기본: 5)
    ConnectionTimeout    time.Duration `json:"connection_timeout"`      // 연결 타임아웃
    ReconnectInterval    time.Duration `json:"reconnect_interval"`      // 재연결 간격

    // DHT 설정
    DHTEnabled           bool          `json:"dht_enabled"`             // DHT 활성화 여부
    DHTBucketSize        int           `json:"dht_bucket_size"`         // DHT 버킷 크기 (기본: 20)
    DHTAlpha             int           `json:"dht_alpha"`               // DHT 알파 값 (기본: 3)

    // 필터링 설정
    EnableReputation     bool          `json:"enable_reputation"`       // 평판 기반 필터링
    MinReputationScore   float64       `json:"min_reputation_score"`    // 최소 평판 점수
    EnableStakeFilter    bool          `json:"enable_stake_filter"`     // 스테이킹 기반 필터링
    MinStakeAmount       uint64        `json:"min_stake_amount"`        // 최소 스테이킹 금액
}
```

#### **1.2 챌린저 특화 DHT (ChallengerDHT)**
```go
// ChallengerDHT implements a challenger-specific distributed hash table
type ChallengerDHT struct {
    // 기본 DHT 정보
    nodeID         string
    challengerID   string
    buckets       [160][]*ChallengerNodeInfo  // 160-bit 키 공간용 버킷들

    // 설정
    k             int                         // 버킷 크기 (기본: 20)
    alpha         int                         // 병렬 요청 수 (기본: 3)

    // 라우팅 테이블 관리
    routingTable  *ChallengerRoutingTable

    // 쿼리 관리
    activeQueries map[string]*DHTQuery
    queryTimeout  time.Duration

    // 통계 및 메트릭스
    stats         *DHTStats

    mu            sync.RWMutex
    logger        log.Logger
}

// ChallengerNodeInfo extends NodeInfo with challenger-specific information
type ChallengerNodeInfo struct {
    *network.NodeInfo                    // 기본 노드 정보
    ChallengerID      string            `json:"challenger_id"`
    ChallengerInfo    *ChallengerInfo   `json:"challenger_info"`
    LastSeen          time.Time         `json:"last_seen"`
    ResponseTime      time.Duration     `json:"response_time"`
    FailureCount      int               `json:"failure_count"`
}

// ChallengerRoutingTable manages routing information for challenger DHT
type ChallengerRoutingTable struct {
    localNodeID   string
    buckets      [160]*KBucket
    mu           sync.RWMutex
}

// KBucket represents a k-bucket in the DHT
type KBucket struct {
    nodes        []*ChallengerNodeInfo
    lastUpdated  time.Time
    maxSize      int
    mu           sync.RWMutex
}

// DHTQuery represents an active DHT query
type DHTQuery struct {
    QueryID      string
    Target       string
    QueryType    DHTQueryType
    StartTime    time.Time
    Timeout      time.Duration
    Results      []*ChallengerNodeInfo
    Contacted    map[string]bool
    Responses    int
    mu           sync.RWMutex
}

// DHTQueryType represents the type of DHT query
type DHTQueryType int

const (
    DHTQueryTypeFindNode       DHTQueryType = iota
    DHTQueryTypeFindChallenger
    DHTQueryTypeStore
    DHTQueryTypePing
)

// DHTStats contains DHT statistics
type DHTStats struct {
    TotalQueries     uint64
    SuccessfulQueries uint64
    FailedQueries    uint64
    AverageLatency   time.Duration
    ActiveNodes      int
    RoutingTableSize int
    LastUpdated      time.Time
}
```

### 2. 연결 관리 시스템

#### **2.1 ChallengerConnectionManager**
```go
// ChallengerConnectionManager manages connections to other challengers
type ChallengerConnectionManager struct {
    // 기본 정보
    nodeID         string
    challengerID   string
    node          *ChallengerP2PNode

    // 연결 관리
    connections    map[string]*ChallengerConnection  // 활성 연결들
    pending        map[string]*PendingConnection     // 대기 중인 연결들
    blacklist      map[string]*BlacklistEntry       // 차단된 노드들

    // 연결 풀 관리
    connectionPool *ConnectionPool

    // 품질 모니터링
    qualityMonitor *ConnectionQualityMonitor

    // 재연결 관리
    reconnectQueue *ReconnectQueue

    // 설정
    config         *ConnectionManagerConfig

    // 상태 관리
    running        bool
    stats          *ConnectionStats

    // 동시성 제어
    mu             sync.RWMutex
    ctx            context.Context
    cancel         context.CancelFunc

    // 이벤트 및 로깅
    eventBus       *EventBus
    logger         log.Logger
}

// ChallengerConnection represents a connection to another challenger
type ChallengerConnection struct {
    // 연결 정보
    ConnectionID     string
    ChallengerID     string
    RemoteAddress    string
    LocalAddress     string

    // 네트워크 연결
    Connection       network.Connection

    // 상태 정보
    Status           ConnectionStatus
    EstablishedAt    time.Time
    LastActivity     time.Time

    // 품질 메트릭스
    Latency          time.Duration
    PacketLoss       float64
    Bandwidth        uint64
    ErrorCount       uint64

    // 챌린저 정보
    RemoteChallengerInfo *ChallengerInfo

    // 보안 정보
    TLSState         *tls.ConnectionState
    Authenticated    bool

    mu               sync.RWMutex
}

// ConnectionStatus represents the status of a connection
type ConnectionStatus int

const (
    ConnectionStatusConnecting   ConnectionStatus = iota
    ConnectionStatusConnected
    ConnectionStatusAuthenticated
    ConnectionStatusDisconnecting
    ConnectionStatusDisconnected
    ConnectionStatusError
)

// PendingConnection represents a connection attempt in progress
type PendingConnection struct {
    ChallengerID  string
    Address       string
    StartTime     time.Time
    Timeout       time.Duration
    Retries       int
    LastError     error
}

// BlacklistEntry represents a blacklisted node
type BlacklistEntry struct {
    ChallengerID  string
    Address       string
    Reason        string
    BlacklistedAt time.Time
    ExpiresAt     time.Time
    Permanent     bool
}

// ConnectionManagerConfig contains configuration for connection manager
type ConnectionManagerConfig struct {
    // 연결 제한
    MaxConnections       int           `json:"max_connections"`        // 최대 연결 수
    MinConnections       int           `json:"min_connections"`        // 최소 연결 수
    MaxPendingConnections int          `json:"max_pending_connections"` // 최대 대기 연결 수

    // 타임아웃 설정
    ConnectionTimeout    time.Duration `json:"connection_timeout"`     // 연결 타임아웃
    HandshakeTimeout     time.Duration `json:"handshake_timeout"`      // 핸드셰이크 타임아웃
    IdleTimeout          time.Duration `json:"idle_timeout"`           // 유휴 타임아웃

    // 재연결 설정
    ReconnectEnabled     bool          `json:"reconnect_enabled"`      // 재연결 활성화
    ReconnectInterval    time.Duration `json:"reconnect_interval"`     // 재연결 간격
    MaxReconnectRetries  int           `json:"max_reconnect_retries"`  // 최대 재연결 시도
    BackoffMultiplier    float64       `json:"backoff_multiplier"`     // 백오프 배수

    // 품질 모니터링
    QualityCheckEnabled  bool          `json:"quality_check_enabled"`  // 품질 체크 활성화
    QualityCheckInterval time.Duration `json:"quality_check_interval"` // 품질 체크 간격
    MinLatency           time.Duration `json:"min_latency"`            // 최소 지연시간
    MaxLatency           time.Duration `json:"max_latency"`            // 최대 지연시간
    MaxPacketLoss        float64       `json:"max_packet_loss"`        // 최대 패킷 손실률

    // 보안 설정
    TLSEnabled           bool          `json:"tls_enabled"`            // TLS 활성화
    RequireAuthentication bool         `json:"require_authentication"` // 인증 필수
    BlacklistEnabled     bool          `json:"blacklist_enabled"`      // 블랙리스트 활성화
    BlacklistDuration    time.Duration `json:"blacklist_duration"`     // 블랙리스트 기간
}
```

#### **2.2 ConnectionPool**
```go
// ConnectionPool manages a pool of reusable connections
type ConnectionPool struct {
    // 풀 관리
    available    []*ChallengerConnection
    inUse        map[string]*ChallengerConnection

    // 설정
    maxSize      int
    minSize      int
    idleTimeout  time.Duration

    // 통계
    stats        *PoolStats

    mu           sync.RWMutex
    logger       log.Logger
}

// PoolStats contains connection pool statistics
type PoolStats struct {
    TotalConnections     int
    AvailableConnections int
    InUseConnections     int
    CreatedConnections   uint64
    ClosedConnections    uint64
    ReuseCount          uint64
    LastCleanup         time.Time
}

// Get retrieves a connection from the pool
func (cp *ConnectionPool) Get(challengerID string) (*ChallengerConnection, error) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    // 기존 연결 확인
    if conn, exists := cp.inUse[challengerID]; exists {
        return conn, nil
    }

    // 사용 가능한 연결 찾기
    for i, conn := range cp.available {
        if conn.ChallengerID == challengerID && cp.isConnectionValid(conn) {
            // 풀에서 제거하고 사용 중으로 이동
            cp.available = append(cp.available[:i], cp.available[i+1:]...)
            cp.inUse[challengerID] = conn
            cp.stats.ReuseCount++
            return conn, nil
        }
    }

    return nil, fmt.Errorf("no available connection for challenger %s", challengerID)
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn *ChallengerConnection) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    // 사용 중에서 제거
    delete(cp.inUse, conn.ChallengerID)

    // 연결이 유효한지 확인
    if !cp.isConnectionValid(conn) {
        conn.Connection.Close()
        cp.stats.ClosedConnections++
        return nil
    }

    // 풀 크기 확인
    if len(cp.available) >= cp.maxSize {
        // 가장 오래된 연결 제거
        oldest := cp.available[0]
        cp.available = cp.available[1:]
        oldest.Connection.Close()
        cp.stats.ClosedConnections++
    }

    // 풀에 추가
    cp.available = append(cp.available, conn)
    return nil
}

// isConnectionValid checks if a connection is still valid
func (cp *ConnectionPool) isConnectionValid(conn *ChallengerConnection) bool {
    if conn.Status != ConnectionStatusConnected && conn.Status != ConnectionStatusAuthenticated {
        return false
    }

    if time.Since(conn.LastActivity) > cp.idleTimeout {
        return false
    }

    // 연결 상태 확인 (ping 등)
    return true
}

// Cleanup removes invalid connections from the pool
func (cp *ConnectionPool) Cleanup() {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    var validConnections []*ChallengerConnection

    for _, conn := range cp.available {
        if cp.isConnectionValid(conn) {
            validConnections = append(validConnections, conn)
        } else {
            conn.Connection.Close()
            cp.stats.ClosedConnections++
        }
    }

    cp.available = validConnections
    cp.stats.LastCleanup = time.Now()
}
```

#### **2.3 ConnectionQualityMonitor**
```go
// ConnectionQualityMonitor monitors connection quality and performance
type ConnectionQualityMonitor struct {
    // 모니터링 대상
    connections    map[string]*ChallengerConnection

    // 품질 메트릭스
    qualityMetrics map[string]*QualityMetrics

    // 설정
    config         *QualityMonitorConfig

    // 상태 관리
    running        bool

    mu             sync.RWMutex
    ctx            context.Context
    cancel         context.CancelFunc
    logger         log.Logger
}

// QualityMetrics contains quality metrics for a connection
type QualityMetrics struct {
    ChallengerID     string

    // 지연시간 메트릭스
    MinLatency       time.Duration
    MaxLatency       time.Duration
    AvgLatency       time.Duration
    LatencyHistory   []time.Duration

    // 패킷 손실 메트릭스
    PacketsSent      uint64
    PacketsReceived  uint64
    PacketLossRate   float64

    // 대역폭 메트릭스
    BytesSent        uint64
    BytesReceived    uint64
    Throughput       uint64

    // 에러 메트릭스
    ErrorCount       uint64
    TimeoutCount     uint64
    LastError        error

    // 시간 정보
    LastUpdated      time.Time
    MonitoringStart  time.Time
}

// QualityMonitorConfig contains configuration for quality monitoring
type QualityMonitorConfig struct {
    CheckInterval     time.Duration `json:"check_interval"`      // 체크 간격
    PingInterval      time.Duration `json:"ping_interval"`       // 핑 간격
    HistorySize       int           `json:"history_size"`        // 히스토리 크기
    QualityThreshold  float64       `json:"quality_threshold"`   // 품질 임계값
    AlertThreshold    float64       `json:"alert_threshold"`     // 알림 임계값
}

// StartMonitoring starts monitoring a connection
func (cqm *ConnectionQualityMonitor) StartMonitoring(conn *ChallengerConnection) {
    cqm.mu.Lock()
    defer cqm.mu.Unlock()

    cqm.connections[conn.ChallengerID] = conn
    cqm.qualityMetrics[conn.ChallengerID] = &QualityMetrics{
        ChallengerID:    conn.ChallengerID,
        LatencyHistory:  make([]time.Duration, 0, cqm.config.HistorySize),
        MonitoringStart: time.Now(),
        LastUpdated:     time.Now(),
    }

    // 모니터링 고루틴 시작
    go cqm.monitorConnection(conn)
}

// StopMonitoring stops monitoring a connection
func (cqm *ConnectionQualityMonitor) StopMonitoring(challengerID string) {
    cqm.mu.Lock()
    defer cqm.mu.Unlock()

    delete(cqm.connections, challengerID)
    delete(cqm.qualityMetrics, challengerID)
}

// monitorConnection monitors a specific connection
func (cqm *ConnectionQualityMonitor) monitorConnection(conn *ChallengerConnection) {
    ticker := time.NewTicker(cqm.config.CheckInterval)
    defer ticker.Stop()

    pingTicker := time.NewTicker(cqm.config.PingInterval)
    defer pingTicker.Stop()

    for {
        select {
        case <-cqm.ctx.Done():
            return
        case <-ticker.C:
            cqm.updateQualityMetrics(conn)
        case <-pingTicker.C:
            cqm.performPing(conn)
        }
    }
}

// performPing performs a ping to measure latency
func (cqm *ConnectionQualityMonitor) performPing(conn *ChallengerConnection) {
    start := time.Now()

    // 핑 메시지 생성 및 전송
    pingMsg := &network.Message{
        Type:      network.MessageTypePing,
        From:      conn.LocalAddress,
        To:        conn.ChallengerID,
        Timestamp: start,
        Nonce:     uint64(start.UnixNano()),
    }

    // 메시지 직렬화 및 전송
    data, err := pingMsg.Serialize()
    if err != nil {
        cqm.recordError(conn.ChallengerID, err)
        return
    }

    err = conn.Connection.Write(data)
    if err != nil {
        cqm.recordError(conn.ChallengerID, err)
        return
    }

    // 응답 대기 (별도 고루틴에서 처리)
    go cqm.waitForPingResponse(conn, start, pingMsg.Nonce)
}

// waitForPingResponse waits for ping response and calculates latency
func (cqm *ConnectionQualityMonitor) waitForPingResponse(conn *ChallengerConnection, start time.Time, nonce uint64) {
    timeout := time.NewTimer(5 * time.Second)
    defer timeout.Stop()

    // 실제 구현에서는 응답 메시지를 기다리는 로직 필요
    // 여기서는 간단히 시뮬레이션
    select {
    case <-timeout.C:
        cqm.recordTimeout(conn.ChallengerID)
    case <-time.After(time.Duration(rand.Intn(100)) * time.Millisecond):
        // 시뮬레이션된 응답
        latency := time.Since(start)
        cqm.recordLatency(conn.ChallengerID, latency)
    }
}

// recordLatency records latency measurement
func (cqm *ConnectionQualityMonitor) recordLatency(challengerID string, latency time.Duration) {
    cqm.mu.Lock()
    defer cqm.mu.Unlock()

    metrics, exists := cqm.qualityMetrics[challengerID]
    if !exists {
        return
    }

    // 지연시간 기록
    if metrics.MinLatency == 0 || latency < metrics.MinLatency {
        metrics.MinLatency = latency
    }
    if latency > metrics.MaxLatency {
        metrics.MaxLatency = latency
    }

    // 히스토리 업데이트
    metrics.LatencyHistory = append(metrics.LatencyHistory, latency)
    if len(metrics.LatencyHistory) > cqm.config.HistorySize {
        metrics.LatencyHistory = metrics.LatencyHistory[1:]
    }

    // 평균 계산
    var total time.Duration
    for _, l := range metrics.LatencyHistory {
        total += l
    }
    metrics.AvgLatency = total / time.Duration(len(metrics.LatencyHistory))

    metrics.LastUpdated = time.Now()
}

// recordError records an error occurrence
func (cqm *ConnectionQualityMonitor) recordError(challengerID string, err error) {
    cqm.mu.Lock()
    defer cqm.mu.Unlock()

    metrics, exists := cqm.qualityMetrics[challengerID]
    if !exists {
        return
    }

    metrics.ErrorCount++
    metrics.LastError = err
    metrics.LastUpdated = time.Now()

    cqm.logger.Warn("Connection error recorded",
        "challenger", challengerID, "error", err, "count", metrics.ErrorCount)
}

// recordTimeout records a timeout occurrence
func (cqm *ConnectionQualityMonitor) recordTimeout(challengerID string) {
    cqm.mu.Lock()
    defer cqm.mu.Unlock()

    metrics, exists := cqm.qualityMetrics[challengerID]
    if !exists {
        return
    }

    metrics.TimeoutCount++
    metrics.LastUpdated = time.Now()

    cqm.logger.Warn("Connection timeout recorded",
        "challenger", challengerID, "count", metrics.TimeoutCount)
}

// GetQualityScore calculates overall quality score for a connection
func (cqm *ConnectionQualityMonitor) GetQualityScore(challengerID string) float64 {
    cqm.mu.RLock()
    defer cqm.mu.RUnlock()

    metrics, exists := cqm.qualityMetrics[challengerID]
    if !exists {
        return 0.0
    }

    // 품질 점수 계산 (0-100)
    score := 100.0

    // 지연시간 페널티
    if metrics.AvgLatency > 100*time.Millisecond {
        score -= 20.0
    } else if metrics.AvgLatency > 50*time.Millisecond {
        score -= 10.0
    }

    // 패킷 손실 페널티
    if metrics.PacketLossRate > 0.05 { // 5%
        score -= 30.0
    } else if metrics.PacketLossRate > 0.01 { // 1%
        score -= 15.0
    }

    // 에러 페널티
    if metrics.ErrorCount > 10 {
        score -= 25.0
    } else if metrics.ErrorCount > 5 {
        score -= 10.0
    }

    // 타임아웃 페널티
    if metrics.TimeoutCount > 5 {
        score -= 15.0
    } else if metrics.TimeoutCount > 2 {
        score -= 5.0
    }

    if score < 0 {
        score = 0
    }

    return score
}
```

### 3. 통합 및 이벤트 시스템

#### **3.1 EventBus**
```go
// EventBus handles events related to discovery and connections
type EventBus struct {
    subscribers map[EventType][]EventSubscriber
    mu          sync.RWMutex
}

// EventType represents the type of event
type EventType int

const (
    EventTypeChallengerDiscovered EventType = iota
    EventTypeChallengerConnected
    EventTypeChallengerDisconnected
    EventTypeConnectionQualityChanged
    EventTypeDiscoveryError
    EventTypeConnectionError
)

// Event represents a system event
type Event struct {
    Type      EventType
    Data      interface{}
    Timestamp time.Time
}

// EventSubscriber interface for event subscribers
type EventSubscriber interface {
    OnEvent(event Event)
}

// Subscribe subscribes to events of specific type
func (eb *EventBus) Subscribe(eventType EventType, subscriber EventSubscriber) {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    if eb.subscribers == nil {
        eb.subscribers = make(map[EventType][]EventSubscriber)
    }

    eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(eventType EventType, data interface{}) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    event := Event{
        Type:      eventType,
        Data:      data,
        Timestamp: time.Now(),
    }

    subscribers, exists := eb.subscribers[eventType]
    if !exists {
        return
    }

    // 비동기적으로 이벤트 전달
    for _, subscriber := range subscribers {
        go subscriber.OnEvent(event)
    }
}
```

## ✅ 구현 체크리스트
- [x] ChallengerDiscovery 확장 설계
- [x] ChallengerDHT 구현 설계
- [x] ChallengerConnectionManager 설계
- [x] ConnectionPool 구현 설계
- [x] ConnectionQualityMonitor 설계
- [x] EventBus 이벤트 시스템 설계
- [x] 설정 구조체들 정의
- [x] 에러 처리 및 로깅 고려

## 🚀 다음 단계
1. **TODO 2.3**: 챌린저 상태 관리 설계
2. **TODO 2.4**: DDoS 방어 시스템 설계
3. **전체 코드 구현**: 모든 설계를 실제 Go 코드로 구현

## 📝 설계 결론
챌린저 발견 및 연결 시스템은 기존 P2P 인프라를 확장하여 챌린저 특화 기능을 제공합니다. DHT 기반 분산 발견, 품질 기반 연결 관리, 이벤트 기반 시스템 통합을 통해 안정적이고 효율적인 챌린저 네트워크를 구축할 수 있습니다.
