# TODO 2.1: 챌린저 노드 확장

## 📋 작업 개요
- **작업명**: 챌린저 노드 확장
- **상태**: ✅ 완료
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
기존 P2PNode 구조체를 확장하여 챌린저 특화 기능을 추가한다. 챌린저 메타데이터 관리, 상태 추적, 고유 ID 생성, 공개키 기반 인증 시스템을 구현한다.

## 📐 구현 요구사항

### 1. 기존 P2PNode 구조체 확장
- 챌린저 특화 기능 추가
- 챌린저 메타데이터 관리
- 상태 추적 및 모니터링

### 2. 챌린저 식별 시스템 구현
- 고유 챌린저 ID 생성
- 공개키 기반 인증
- 챌린저 서명 및 검증

## 🏗️ 구현 설계

### 1. 확장된 P2PNode 구조체

#### **1.1 ChallengerP2PNode 구조체**
```go
// ChallengerP2PNode extends P2PNode with challenger-specific functionality
type ChallengerP2PNode struct {
    // 기존 P2PNode 임베딩
    *network.P2PNode

    // 챌린저 네트워크 관리
    challengerNetwork *ChallengerNetwork

    // 챌린저 특화 정보
    challengerInfo    *ChallengerInfo
    challengerID      string

    // 인증 및 보안
    challengerSigner  *ChallengerSigner
    validator         *ChallengerValidator

    // 상태 관리
    statusTracker     *StatusTracker
    healthMonitor     *HealthMonitor

    // 메트릭스 및 모니터링
    metrics          *ChallengerMetrics

    // 설정 및 로깅
    config           *ChallengerNodeConfig
    logger           log.Logger

    // 동시성 제어
    mu               sync.RWMutex
    ctx              context.Context
    cancel           context.CancelFunc
}
```

#### **1.2 ChallengerNodeConfig**
```go
// ChallengerNodeConfig contains configuration for challenger node
type ChallengerNodeConfig struct {
    // 기본 설정
    NodeName          string        `json:"node_name"`           // 노드 이름
    ChallengerRole    ChallengerRole `json:"challenger_role"`    // 챌린저 역할

    // 네트워크 설정
    P2PConfig         *network.P2PConfig         `json:"p2p_config"`         // P2P 설정
    NetworkConfig     *ChallengerNetworkConfig   `json:"network_config"`     // 챌린저 네트워크 설정

    // 보안 설정
    PrivateKeyPath    string        `json:"private_key_path"`    // 개인키 파일 경로
    RequireAuth       bool          `json:"require_auth"`        // 인증 필수 여부
    SignatureRequired bool          `json:"signature_required"`  // 서명 필수 여부

    // 모니터링 설정
    MetricsEnabled    bool          `json:"metrics_enabled"`     // 메트릭스 활성화
    HealthCheckInterval time.Duration `json:"health_check_interval"` // 헬스체크 간격
    StatusUpdateInterval time.Duration `json:"status_update_interval"` // 상태 업데이트 간격

    // 성능 설정
    MaxConcurrentRequests int       `json:"max_concurrent_requests"` // 최대 동시 요청 수
    RequestTimeout        time.Duration `json:"request_timeout"`      // 요청 타임아웃

    // 스테이킹 설정
    StakeAmount       uint64        `json:"stake_amount"`        // 스테이킹 금액
    StakeRequired     bool          `json:"stake_required"`      // 스테이킹 필수 여부
}
```

### 2. 챌린저 식별 시스템

#### **2.1 ChallengerIDGenerator**
```go
// ChallengerIDGenerator generates unique challenger IDs
type ChallengerIDGenerator struct {
    nodeID    string
    publicKey *ecdsa.PublicKey
    timestamp time.Time
}

// GenerateChallengerID generates a unique challenger ID
func (gen *ChallengerIDGenerator) GenerateChallengerID() string {
    // 공개키 + 노드ID + 타임스탬프 기반 해시 생성
    data := fmt.Sprintf("%s-%s-%d",
        gen.nodeID,
        hex.EncodeToString(crypto.FromECDSAPub(gen.publicKey)),
        gen.timestamp.Unix())

    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("challenger-%s", hex.EncodeToString(hash[:16]))
}

// ValidateChallengerID validates challenger ID format
func ValidateChallengerID(challengerID string) bool {
    // challenger-[32자리 hex] 형식 검증
    pattern := `^challenger-[a-f0-9]{32}$`
    matched, _ := regexp.MatchString(pattern, challengerID)
    return matched
}
```

#### **2.2 ChallengerSigner**
```go
// ChallengerSigner handles challenger message signing and verification
type ChallengerSigner struct {
    privateKey   *ecdsa.PrivateKey
    publicKey    *ecdsa.PublicKey
    challengerID string

    mu           sync.RWMutex
}

// NewChallengerSigner creates a new challenger signer
func NewChallengerSigner(privateKey *ecdsa.PrivateKey, challengerID string) *ChallengerSigner {
    return &ChallengerSigner{
        privateKey:   privateKey,
        publicKey:    &privateKey.PublicKey,
        challengerID: challengerID,
    }
}

// SignMessage signs a message with challenger's private key
func (cs *ChallengerSigner) SignMessage(data []byte) ([]byte, error) {
    cs.mu.RLock()
    defer cs.mu.RUnlock()

    hash := crypto.Keccak256Hash(data)
    signature, err := crypto.Sign(hash.Bytes(), cs.privateKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign message: %w", err)
    }

    return signature, nil
}

// VerifySignature verifies a signature against challenger's public key
func (cs *ChallengerSigner) VerifySignature(data []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
    hash := crypto.Keccak256Hash(data)

    // Signature에서 recovery ID 제거 (마지막 바이트)
    if len(signature) == 65 {
        signature = signature[:64]
    }

    return crypto.VerifySignature(
        crypto.FromECDSAPub(publicKey),
        hash.Bytes(),
        signature,
    )
}

// GetPublicKey returns challenger's public key
func (cs *ChallengerSigner) GetPublicKey() *ecdsa.PublicKey {
    cs.mu.RLock()
    defer cs.mu.RUnlock()
    return cs.publicKey
}

// GetChallengerID returns challenger ID
func (cs *ChallengerSigner) GetChallengerID() string {
    return cs.challengerID
}
```

#### **2.3 ChallengerValidator**
```go
// ChallengerValidator validates challenger authenticity and permissions
type ChallengerValidator struct {
    knownChallengers map[string]*ChallengerInfo
    stakeValidator   *StakeValidator

    mu               sync.RWMutex
    logger           log.Logger
}

// NewChallengerValidator creates a new challenger validator
func NewChallengerValidator(logger log.Logger) *ChallengerValidator {
    return &ChallengerValidator{
        knownChallengers: make(map[string]*ChallengerInfo),
        stakeValidator:   NewStakeValidator(),
        logger:           logger,
    }
}

// ValidateChallenger validates challenger registration and authenticity
func (cv *ChallengerValidator) ValidateChallenger(info *ChallengerInfo) error {
    cv.mu.RLock()
    defer cv.mu.RUnlock()

    // 1. 챌린저 ID 형식 검증
    if !ValidateChallengerID(info.ID) {
        return fmt.Errorf("invalid challenger ID format: %s", info.ID)
    }

    // 2. 공개키 검증
    if info.PublicKey == nil {
        return fmt.Errorf("public key is required")
    }

    // 3. 스테이킹 검증 (필요한 경우)
    if err := cv.stakeValidator.ValidateStake(info); err != nil {
        return fmt.Errorf("stake validation failed: %w", err)
    }

    // 4. 중복 등록 검증
    if existing, exists := cv.knownChallengers[info.ID]; exists {
        if !cv.isSameChallenger(existing, info) {
            return fmt.Errorf("challenger ID already exists with different identity")
        }
    }

    return nil
}

// AuthorizeAction checks if challenger is authorized for specific action
func (cv *ChallengerValidator) AuthorizeAction(challengerID string, action string) error {
    cv.mu.RLock()
    defer cv.mu.RUnlock()

    challenger, exists := cv.knownChallengers[challengerID]
    if !exists {
        return fmt.Errorf("unknown challenger: %s", challengerID)
    }

    // 상태 확인
    if challenger.Status == ChallengerStatusBanned {
        return fmt.Errorf("challenger is banned")
    }

    if challenger.Status == ChallengerStatusSuspended {
        return fmt.Errorf("challenger is suspended")
    }

    // 역할 기반 권한 확인
    if !cv.hasPermission(challenger, action) {
        return fmt.Errorf("insufficient permissions for action: %s", action)
    }

    return nil
}

// AddChallenger adds a validated challenger to known list
func (cv *ChallengerValidator) AddChallenger(info *ChallengerInfo) error {
    if err := cv.ValidateChallenger(info); err != nil {
        return err
    }

    cv.mu.Lock()
    defer cv.mu.Unlock()

    cv.knownChallengers[info.ID] = info
    cv.logger.Info("Challenger added to validator", "id", info.ID, "address", info.Address)

    return nil
}

// RemoveChallenger removes challenger from known list
func (cv *ChallengerValidator) RemoveChallenger(challengerID string) {
    cv.mu.Lock()
    defer cv.mu.Unlock()

    delete(cv.knownChallengers, challengerID)
    cv.logger.Info("Challenger removed from validator", "id", challengerID)
}

// 내부 헬퍼 메서드들
func (cv *ChallengerValidator) isSameChallenger(existing, new *ChallengerInfo) bool {
    return existing.NodeID == new.NodeID &&
           existing.Address == new.Address &&
           crypto.FromECDSAPub(existing.PublicKey) != nil &&
           crypto.FromECDSAPub(new.PublicKey) != nil &&
           bytes.Equal(crypto.FromECDSAPub(existing.PublicKey), crypto.FromECDSAPub(new.PublicKey))
}

func (cv *ChallengerValidator) hasPermission(challenger *ChallengerInfo, action string) bool {
    // 역할 기반 권한 확인 로직
    switch action {
    case "participate_attention_test":
        return challenger.Role.HasRole(ChallengerRoleChallenger)
    case "sequence_backup":
        return challenger.Role.HasRole(ChallengerRoleSequencer)
    default:
        return true // 기본 액션은 모든 챌린저에게 허용
    }
}
```

### 3. 상태 추적 시스템

#### **3.1 StatusTracker**
```go
// StatusTracker tracks challenger status and state changes
type StatusTracker struct {
    currentStatus    ChallengerStatus
    statusHistory    []StatusChange
    lastUpdate       time.Time

    // 상태 변경 알림
    statusListeners  []StatusListener

    mu               sync.RWMutex
    logger           log.Logger
}

// StatusChange represents a status change event
type StatusChange struct {
    From      ChallengerStatus `json:"from"`
    To        ChallengerStatus `json:"to"`
    Reason    string           `json:"reason"`
    Timestamp time.Time        `json:"timestamp"`
}

// StatusListener interface for status change notifications
type StatusListener interface {
    OnStatusChange(change StatusChange)
}

// NewStatusTracker creates a new status tracker
func NewStatusTracker(initialStatus ChallengerStatus, logger log.Logger) *StatusTracker {
    return &StatusTracker{
        currentStatus:   initialStatus,
        statusHistory:   make([]StatusChange, 0),
        lastUpdate:      time.Now(),
        statusListeners: make([]StatusListener, 0),
        logger:          logger,
    }
}

// UpdateStatus updates challenger status
func (st *StatusTracker) UpdateStatus(newStatus ChallengerStatus, reason string) error {
    st.mu.Lock()
    defer st.mu.Unlock()

    if st.currentStatus == newStatus {
        return nil // 상태 변경 없음
    }

    // 상태 변경 기록
    change := StatusChange{
        From:      st.currentStatus,
        To:        newStatus,
        Reason:    reason,
        Timestamp: time.Now(),
    }

    st.statusHistory = append(st.statusHistory, change)
    st.currentStatus = newStatus
    st.lastUpdate = change.Timestamp

    // 상태 변경 알림
    go st.notifyStatusChange(change)

    st.logger.Info("Challenger status updated",
        "from", st.currentStatus, "to", newStatus, "reason", reason)

    return nil
}

// GetCurrentStatus returns current status
func (st *StatusTracker) GetCurrentStatus() ChallengerStatus {
    st.mu.RLock()
    defer st.mu.RUnlock()
    return st.currentStatus
}

// GetStatusHistory returns status change history
func (st *StatusTracker) GetStatusHistory() []StatusChange {
    st.mu.RLock()
    defer st.mu.RUnlock()

    history := make([]StatusChange, len(st.statusHistory))
    copy(history, st.statusHistory)
    return history
}

// AddStatusListener adds a status change listener
func (st *StatusTracker) AddStatusListener(listener StatusListener) {
    st.mu.Lock()
    defer st.mu.Unlock()
    st.statusListeners = append(st.statusListeners, listener)
}

// 내부 메서드
func (st *StatusTracker) notifyStatusChange(change StatusChange) {
    for _, listener := range st.statusListeners {
        go listener.OnStatusChange(change)
    }
}
```

#### **3.2 HealthMonitor**
```go
// HealthMonitor monitors challenger node health
type HealthMonitor struct {
    healthStatus     HealthStatus
    lastHeartbeat    time.Time
    errorCount       uint64
    responseTime     time.Duration

    // 모니터링 설정
    checkInterval    time.Duration
    timeout          time.Duration
    maxErrors        uint64

    // 상태 관리
    running          bool

    mu               sync.RWMutex
    ctx              context.Context
    cancel           context.CancelFunc
    logger           log.Logger
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval, timeout time.Duration, maxErrors uint64, logger log.Logger) *HealthMonitor {
    ctx, cancel := context.WithCancel(context.Background())

    return &HealthMonitor{
        healthStatus:  HealthStatusUnknown,
        checkInterval: checkInterval,
        timeout:       timeout,
        maxErrors:     maxErrors,
        ctx:           ctx,
        cancel:        cancel,
        logger:        logger,
    }
}

// Start starts health monitoring
func (hm *HealthMonitor) Start() error {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    if hm.running {
        return fmt.Errorf("health monitor already running")
    }

    hm.running = true
    go hm.monitorLoop()

    hm.logger.Info("Health monitor started")
    return nil
}

// Stop stops health monitoring
func (hm *HealthMonitor) Stop() error {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    if !hm.running {
        return fmt.Errorf("health monitor not running")
    }

    hm.cancel()
    hm.running = false

    hm.logger.Info("Health monitor stopped")
    return nil
}

// UpdateHeartbeat updates last heartbeat time
func (hm *HealthMonitor) UpdateHeartbeat() {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    hm.lastHeartbeat = time.Now()

    // 에러 카운트 리셋 (정상 하트비트 수신 시)
    if hm.errorCount > 0 {
        hm.errorCount = 0
        hm.updateHealthStatus()
    }
}

// RecordError records an error occurrence
func (hm *HealthMonitor) RecordError() {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    hm.errorCount++
    hm.updateHealthStatus()

    hm.logger.Warn("Health monitor recorded error", "count", hm.errorCount)
}

// GetHealthStatus returns current health status
func (hm *HealthMonitor) GetHealthStatus() HealthStatus {
    hm.mu.RLock()
    defer hm.mu.RUnlock()
    return hm.healthStatus
}

// GetHealthData returns comprehensive health data
func (hm *HealthMonitor) GetHealthData() *HealthData {
    hm.mu.RLock()
    defer hm.mu.RUnlock()

    return &HealthData{
        LastHeartbeat: hm.lastHeartbeat,
        ResponseTime:  hm.responseTime,
        ErrorCount:    hm.errorCount,
        Status:        hm.healthStatus,
    }
}

// 내부 메서드들
func (hm *HealthMonitor) monitorLoop() {
    ticker := time.NewTicker(hm.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-hm.ctx.Done():
            return
        case <-ticker.C:
            hm.performHealthCheck()
        }
    }
}

func (hm *HealthMonitor) performHealthCheck() {
    hm.mu.Lock()
    defer hm.mu.Unlock()

    now := time.Now()
    timeSinceLastHeartbeat := now.Sub(hm.lastHeartbeat)

    // 하트비트 타임아웃 체크
    if timeSinceLastHeartbeat > hm.timeout {
        hm.errorCount++
        hm.logger.Warn("Heartbeat timeout", "duration", timeSinceLastHeartbeat)
    }

    hm.updateHealthStatus()
}

func (hm *HealthMonitor) updateHealthStatus() {
    oldStatus := hm.healthStatus

    if hm.errorCount >= hm.maxErrors {
        hm.healthStatus = HealthStatusUnhealthy
    } else if hm.errorCount > hm.maxErrors/2 {
        hm.healthStatus = HealthStatusDegraded
    } else {
        hm.healthStatus = HealthStatusHealthy
    }

    if oldStatus != hm.healthStatus {
        hm.logger.Info("Health status changed",
            "from", oldStatus, "to", hm.healthStatus, "errors", hm.errorCount)
    }
}
```

## 🔄 통합 및 확장

### 1. ChallengerP2PNode 구현

#### **1.1 생성자 및 초기화**
```go
// NewChallengerP2PNode creates a new challenger P2P node
func NewChallengerP2PNode(config *ChallengerNodeConfig, logger log.Logger) (*ChallengerP2PNode, error) {
    // 기존 P2PNode 생성
    p2pNode, err := network.NewP2PNode(config.P2PConfig, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create P2P node: %w", err)
    }

    // 챌린저 ID 생성
    idGenerator := &ChallengerIDGenerator{
        nodeID:    p2pNode.GetID(),
        publicKey: p2pNode.GetPublicKey(),
        timestamp: time.Now(),
    }
    challengerID := idGenerator.GenerateChallengerID()

    // 챌린저 서명자 생성
    signer := NewChallengerSigner(p2pNode.GetPrivateKey(), challengerID)

    // 챌린저 검증자 생성
    validator := NewChallengerValidator(logger)

    // 상태 추적자 생성
    statusTracker := NewStatusTracker(ChallengerStatusOffline, logger)

    // 헬스 모니터 생성
    healthMonitor := NewHealthMonitor(
        config.HealthCheckInterval,
        config.RequestTimeout,
        10, // maxErrors
        logger,
    )

    // 챌린저 정보 생성
    challengerInfo := &ChallengerInfo{
        ID:        challengerID,
        NodeID:    p2pNode.GetID(),
        Address:   p2pNode.GetAddress(),
        PublicKey: p2pNode.GetPublicKey(),
        Role:      config.ChallengerRole,
        Status:    ChallengerStatusOffline,
        Reputation: ReputationScore{
            Overall:     50.0, // 초기 평판
            Accuracy:    50.0,
            Speed:       50.0,
            Reliability: 50.0,
            LastUpdate:  time.Now(),
        },
        StakeAmount: config.StakeAmount,
        StakeStatus: StakeStatusNone,
        LastSeen:    time.Now(),
        Version:     "1.0.0",
        Timestamp:   time.Now(),
    }

    ctx, cancel := context.WithCancel(context.Background())

    node := &ChallengerP2PNode{
        P2PNode:           p2pNode,
        challengerInfo:    challengerInfo,
        challengerID:      challengerID,
        challengerSigner:  signer,
        validator:         validator,
        statusTracker:     statusTracker,
        healthMonitor:     healthMonitor,
        config:            config,
        logger:            logger,
        ctx:               ctx,
        cancel:            cancel,
    }

    // 챌린저 네트워크 생성
    challengerNetwork, err := NewChallengerNetwork(p2pNode, config.NetworkConfig, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create challenger network: %w", err)
    }
    node.challengerNetwork = challengerNetwork

    return node, nil
}
```

#### **1.2 주요 메서드들**
```go
// Start starts the challenger P2P node
func (cn *ChallengerP2PNode) Start() error {
    cn.mu.Lock()
    defer cn.mu.Unlock()

    // 기존 P2P 노드 시작
    if err := cn.P2PNode.Start(); err != nil {
        return fmt.Errorf("failed to start P2P node: %w", err)
    }

    // 헬스 모니터 시작
    if err := cn.healthMonitor.Start(); err != nil {
        return fmt.Errorf("failed to start health monitor: %w", err)
    }

    // 챌린저 네트워크 시작
    if err := cn.challengerNetwork.Start(); err != nil {
        return fmt.Errorf("failed to start challenger network: %w", err)
    }

    // 상태를 온라인으로 변경
    cn.statusTracker.UpdateStatus(ChallengerStatusOnline, "node started")

    // 자신을 챌린저 네트워크에 등록
    if err := cn.challengerNetwork.RegisterChallenger(cn.challengerInfo); err != nil {
        cn.logger.Warn("Failed to register self to challenger network", "error", err)
    }

    cn.logger.Info("Challenger P2P node started", "id", cn.challengerID)
    return nil
}

// Stop stops the challenger P2P node
func (cn *ChallengerP2PNode) Stop() error {
    cn.mu.Lock()
    defer cn.mu.Unlock()

    // 상태를 오프라인으로 변경
    cn.statusTracker.UpdateStatus(ChallengerStatusOffline, "node stopping")

    // 챌린저 네트워크에서 자신을 제거
    cn.challengerNetwork.UnregisterChallenger(cn.challengerID)

    // 챌린저 네트워크 중지
    if err := cn.challengerNetwork.Stop(); err != nil {
        cn.logger.Warn("Failed to stop challenger network", "error", err)
    }

    // 헬스 모니터 중지
    if err := cn.healthMonitor.Stop(); err != nil {
        cn.logger.Warn("Failed to stop health monitor", "error", err)
    }

    // 기존 P2P 노드 중지
    if err := cn.P2PNode.Stop(); err != nil {
        return fmt.Errorf("failed to stop P2P node: %w", err)
    }

    cn.cancel()
    cn.logger.Info("Challenger P2P node stopped")
    return nil
}

// GetChallengerInfo returns challenger information
func (cn *ChallengerP2PNode) GetChallengerInfo() *ChallengerInfo {
    cn.mu.RLock()
    defer cn.mu.RUnlock()
    return cn.challengerInfo
}

// GetChallengerID returns challenger ID
func (cn *ChallengerP2PNode) GetChallengerID() string {
    return cn.challengerID
}

// UpdateChallengerStatus updates challenger status
func (cn *ChallengerP2PNode) UpdateChallengerStatus(status ChallengerStatus, reason string) error {
    cn.mu.Lock()
    defer cn.mu.Unlock()

    if err := cn.statusTracker.UpdateStatus(status, reason); err != nil {
        return err
    }

    cn.challengerInfo.Status = status
    cn.challengerInfo.Timestamp = time.Now()

    // 챌린저 네트워크에 상태 변경 알림
    return cn.challengerNetwork.UpdateChallengerStatus(cn.challengerID, status)
}

// SignMessage signs a message with challenger's private key
func (cn *ChallengerP2PNode) SignMessage(data []byte) ([]byte, error) {
    return cn.challengerSigner.SignMessage(data)
}

// VerifyMessage verifies a message signature
func (cn *ChallengerP2PNode) VerifyMessage(data []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
    return cn.challengerSigner.VerifySignature(data, signature, publicKey)
}

// GetHealthStatus returns current health status
func (cn *ChallengerP2PNode) GetHealthStatus() HealthStatus {
    return cn.healthMonitor.GetHealthStatus()
}
```

## ✅ 구현 체크리스트
- [x] ChallengerP2PNode 구조체 설계
- [x] ChallengerIDGenerator 구현
- [x] ChallengerSigner 구현 (서명/검증)
- [x] ChallengerValidator 구현 (인증/권한)
- [x] StatusTracker 구현 (상태 추적)
- [x] HealthMonitor 구현 (헬스 모니터링)
- [x] 기존 P2PNode와의 통합 방안
- [x] 생성자 및 주요 메서드 구현

## 🚀 다음 단계
1. **실제 코드 구현**: Go 파일로 구현
2. **단위 테스트 작성**: 각 컴포넌트별 테스트
3. **통합 테스트**: 전체 시스템 연동 테스트

## 📝 구현 결론
기존 P2PNode를 확장한 ChallengerP2PNode는 챌린저 특화 기능을 제공하면서도 기존 P2P 인프라와 완벽하게 통합됩니다. 모듈화된 설계로 각 컴포넌트가 독립적으로 테스트 가능하며, 확장성과 유지보수성을 고려한 구조입니다.
