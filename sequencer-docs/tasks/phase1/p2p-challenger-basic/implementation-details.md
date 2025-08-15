# P2P 챌린저 네트워크 - 구현 세부사항

## 📋 개요

이 문서는 P2P 챌린저 네트워크의 구체적인 구현 방법과 기술적 세부사항을 제공합니다. 각 컴포넌트의 상세한 설계와 구현 방안을 단계별로 설명합니다.

## 🏗️ 1. P2P 노드 관리 시스템

### 1.1 P2P 노드 구조 설계

#### 핵심 데이터 구조
```go
// op-challenger/p2p/network/node.go
type P2PNode struct {
    // 기본 정보
    id          string              // 노드 고유 ID (공개키 기반)
    address     string              // 노드 주소 (IP:Port)
    publicKey   *ecdsa.PublicKey    // 공개키
    privateKey  *ecdsa.PrivateKey   // 개인키

    // 네트워크 관리
    peers       map[string]*Peer    // 연결된 피어들
    discovery   *DiscoveryService   // 노드 발견 서비스
    transport   Transport           // 네트워크 전송 계층

    // 상태 관리
    status      NodeStatus          // 노드 상태
    lastSeen    time.Time           // 마지막 활동 시간
    reputation  float64             // 평판 점수

    // 동시성 제어
    mu          sync.RWMutex        // 읽기/쓰기 뮤텍스
    ctx         context.Context     // 컨텍스트
    cancel      context.CancelFunc  // 취소 함수
}

type Peer struct {
    ID          string              // 피어 ID
    Address     string              // 피어 주소
    PublicKey   *ecdsa.PublicKey    // 피어 공개키
    Status      PeerStatus          // 피어 상태
    Reputation  float64             // 피어 평판
    LastSeen    time.Time           // 마지막 활동 시간
    Connection  *Connection         // 연결 객체
}

type NodeStatus int

const (
    NodeStatusStarting NodeStatus = iota
    NodeStatusRunning
    NodeStatusStopping
    NodeStatusStopped
    NodeStatusError
)
```

#### 노드 초기화 및 시작
```go
// op-challenger/p2p/network/node.go
func NewP2PNode(config *P2PConfig) (*P2PNode, error) {
    // 개인키 생성 또는 로드
    privateKey, err := loadOrGeneratePrivateKey(config.PrivateKeyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load private key: %w", err)
    }

    // 공개키 추출
    publicKey := privateKey.Public().(*ecdsa.PublicKey)

    // 노드 ID 생성 (공개키 해시)
    nodeID := generateNodeID(publicKey)

    // 네트워크 전송 계층 초기화
    transport, err := NewTransport(config.TransportConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create transport: %w", err)
    }

    // 노드 발견 서비스 초기화
    discovery := NewDiscoveryService(config.DiscoveryConfig)

    ctx, cancel := context.WithCancel(context.Background())

    return &P2PNode{
        id:          nodeID,
        address:     config.Address,
        publicKey:   publicKey,
        privateKey:  privateKey,
        peers:       make(map[string]*Peer),
        discovery:   discovery,
        transport:   transport,
        status:      NodeStatusStarting,
        reputation:  1.0, // 초기 평판
        ctx:         ctx,
        cancel:      cancel,
    }, nil
}

func (n *P2PNode) Start() error {
    n.mu.Lock()
    defer n.mu.Unlock()

    if n.status != NodeStatusStarting {
        return fmt.Errorf("node is not in starting status")
    }

    // 전송 계층 시작
    if err := n.transport.Start(n.address); err != nil {
        return fmt.Errorf("failed to start transport: %w", err)
    }

    // 노드 발견 서비스 시작
    if err := n.discovery.Start(n); err != nil {
        return fmt.Errorf("failed to start discovery: %w", err)
    }

    // 연결 수락 루틴 시작
    go n.acceptConnections()

    // 피어 관리 루틴 시작
    go n.managePeers()

    n.status = NodeStatusRunning
    n.logger.Info("P2P node started", "id", n.id, "address", n.address)

    return nil
}
```

### 1.2 자동 노드 발견 (DHT 기반)

#### DHT 구현
```go
// op-challenger/p2p/network/discovery.go
type DiscoveryService struct {
    dht         *DHT                // 분산 해시 테이블
    nodeID      string              // 현재 노드 ID
    bootstrap   []string            // 부트스트랩 노드들
    knownNodes  map[string]*NodeInfo // 알려진 노드들
    mu          sync.RWMutex

    // 설정
    config      DiscoveryConfig
    logger      log.Logger
}

type DHT struct {
    buckets     [160][]*NodeInfo    // Kademlia 스타일 버킷
    nodeID      string              // 현재 노드 ID
    k           int                 // 버킷 크기 (보통 20)
    alpha       int                 // 병렬 요청 수 (보통 3)
}

type NodeInfo struct {
    ID          string
    Address     string
    PublicKey   *ecdsa.PublicKey
    LastSeen    time.Time
    Distance    *big.Int            // XOR 거리
}

// Kademlia 스타일 노드 발견
func (ds *DiscoveryService) FindPeers(targetID string) ([]*NodeInfo, error) {
    ds.mu.Lock()
    defer ds.mu.Unlock()

    // 1. 로컬 버킷에서 가장 가까운 k개 노드 찾기
    closest := ds.dht.FindClosest(targetID, ds.config.K)

    // 2. 병렬로 FIND_NODE 요청 전송
    var found []*NodeInfo
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, node := range closest {
        wg.Add(1)
        go func(n *NodeInfo) {
            defer wg.Done()

            // FIND_NODE RPC 호출
            peers, err := ds.sendFindNode(n, targetID)
            if err != nil {
                ds.logger.Debug("Failed to find node", "node", n.ID, "error", err)
                return
            }

            mu.Lock()
            found = append(found, peers...)
            mu.Unlock()
        }(node)
    }

    wg.Wait()

    // 3. 결과 정렬 및 중복 제거
    unique := ds.removeDuplicates(found)
    sort.Slice(unique, func(i, j int) bool {
        return unique[i].Distance.Cmp(unique[j].Distance) < 0
    })

    return unique[:min(len(unique), ds.config.K)], nil
}

// 노드 등록 (자신을 네트워크에 알림)
func (ds *DiscoveryService) AnnouncePresence() error {
    // 1. 부트스트랩 노드들에게 자신 등록
    for _, bootstrap := range ds.bootstrap {
        if err := ds.announceToNode(bootstrap); err != nil {
            ds.logger.Debug("Failed to announce to bootstrap", "node", bootstrap, "error", err)
        }
    }

    // 2. 알려진 노드들에게 자신 등록
    ds.mu.RLock()
    knownNodes := make([]*NodeInfo, 0, len(ds.knownNodes))
    for _, node := range ds.knownNodes {
        knownNodes = append(knownNodes, node)
    }
    ds.mu.RUnlock()

    for _, node := range knownNodes {
        if err := ds.announceToNode(node.Address); err != nil {
            ds.logger.Debug("Failed to announce to known node", "node", node.ID, "error", err)
        }
    }

    return nil
}
```

#### RPC 메시지 정의
```go
// op-challenger/p2p/network/messages.go
type Message struct {
    Type      MessageType `json:"type"`
    From      string      `json:"from"`
    To        string      `json:"to,omitempty"`      // 빈 값이면 브로드캐스트
    Payload   []byte      `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
    Signature []byte      `json:"signature"`
    Nonce     uint64      `json:"nonce"`             // 재생 공격 방지
}

type MessageType int

const (
    MessageTypePing MessageType = iota
    MessageTypePong
    MessageTypeFindNode
    MessageTypeFindNodeResponse
    MessageTypeAnnounce
    MessageTypeGameState
    MessageTypeValidationResult
    MessageTypeReputationUpdate
)

// FIND_NODE RPC
type FindNodeRequest struct {
    TargetID string `json:"target_id"`
}

type FindNodeResponse struct {
    Nodes []*NodeInfo `json:"nodes"`
}

// ANNOUNCE RPC
type AnnounceRequest struct {
    NodeInfo *NodeInfo `json:"node_info"`
}

type AnnounceResponse struct {
    Success bool `json:"success"`
}
```

### 1.3 연결 관리

#### 연결 수립 및 관리
```go
// op-challenger/p2p/network/connection.go
type ConnectionManager struct {
    node        *P2PNode
    connections map[string]*Connection
    listener    net.Listener
    mu          sync.RWMutex

    // 연결 풀 관리
    maxConnections int
    connectionPool chan struct{}
}

type Connection struct {
    id       string
    peer     *Peer
    conn     net.Conn
    encoder  *json.Encoder
    decoder  *json.Decoder
    status   ConnectionStatus
    lastPing time.Time

    // 메시지 처리
    sendQueue    chan *Message
    receiveQueue chan *Message
    ctx          context.Context
    cancel       context.CancelFunc
}

func (cm *ConnectionManager) Connect(peer *Peer) error {
    // 연결 풀 확인
    select {
    case cm.connectionPool <- struct{}{}:
        defer func() {
            <-cm.connectionPool
        }()
    default:
        return fmt.Errorf("connection pool full")
    }

    // TCP 연결 수립
    conn, err := net.DialTimeout("tcp", peer.Address, 10*time.Second)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", peer.Address, err)
    }

    // 핸드셰이크 수행
    if err := cm.performHandshake(conn, peer); err != nil {
        conn.Close()
        return fmt.Errorf("handshake failed: %w", err)
    }

    // 연결 객체 생성
    connection := cm.createConnection(conn, peer)

    // 연결 등록
    cm.mu.Lock()
    cm.connections[peer.ID] = connection
    cm.mu.Unlock()

    // 메시지 처리 루틴 시작
    go connection.handleMessages()

    return nil
}

func (cm *ConnectionManager) performHandshake(conn net.Conn, peer *Peer) error {
    // 1. 클라이언트 -> 서버: 연결 요청
    handshake := &HandshakeRequest{
        NodeID:    cm.node.id,
        PublicKey: cm.node.publicKey,
        Version:   "1.0.0",
    }

    if err := json.NewEncoder(conn).Encode(handshake); err != nil {
        return fmt.Errorf("failed to send handshake: %w", err)
    }

    // 2. 서버 -> 클라이언트: 연결 응답
    var response HandshakeResponse
    if err := json.NewDecoder(conn).Decode(&response); err != nil {
        return fmt.Errorf("failed to receive handshake response: %w", err)
    }

    if !response.Accepted {
        return fmt.Errorf("handshake rejected: %s", response.Reason)
    }

    // 3. 피어 정보 업데이트
    peer.PublicKey = response.PublicKey

    return nil
}
```

## 🔄 2. 챌린저 간 상태 동기화

### 2.1 상태 동기화 프로토콜

#### 상태 정의
```go
// op-challenger/p2p/sync/state.go
type ChallengerState struct {
    NodeID          string                    `json:"node_id"`
    LastValidation  time.Time                 `json:"last_validation"`
    ActiveGames     map[common.Hash]*GameInfo `json:"active_games"`
    Reputation      float64                   `json:"reputation"`
    Peers           map[string]*PeerInfo      `json:"peers"`
    Version         uint64                    `json:"version"`           // 상태 버전
    LastUpdate      time.Time                 `json:"last_update"`
}

type GameInfo struct {
    GameAddress     common.Hash `json:"game_address"`
    Status          GameStatus  `json:"status"`
    LastMove        time.Time   `json:"last_move"`
    CurrentClaim    uint64      `json:"current_claim"`
    MyRole          PlayerRole  `json:"my_role"`
    LastValidation  time.Time   `json:"last_validation"`
}

type PeerInfo struct {
    ID              string    `json:"id"`
    LastSeen        time.Time `json:"last_seen"`
    ValidationCount int       `json:"validation_count"`
    Reputation      float64   `json:"reputation"`
    ActiveGames     int       `json:"active_games"`
}
```

#### 동기화 메커니즘
```go
// op-challenger/p2p/sync/sync.go
type StateSync struct {
    node        *P2PNode
    state       *ChallengerState
    peers       map[string]*Peer
    mu          sync.RWMutex

    // 동기화 설정
    syncInterval time.Duration
    batchSize    int
    timeout      time.Duration
}

func (ss *StateSync) SyncWithPeer(peer *Peer) error {
    // 1. 피어의 최신 상태 요청
    request := &StateRequest{
        FromVersion: ss.state.Version,
        GameFilter:  ss.getActiveGameHashes(),
    }

    response, err := ss.sendStateRequest(peer, request)
    if err != nil {
        return fmt.Errorf("failed to request state from peer: %w", err)
    }

    // 2. 상태 병합
    if err := ss.mergeState(response.State); err != nil {
        return fmt.Errorf("failed to merge state: %w", err)
    }

    // 3. 로컬 상태를 피어에게 전송
    if err := ss.sendStateToPeer(peer, ss.state); err != nil {
        return fmt.Errorf("failed to send state to peer: %w", err)
    }

    return nil
}

func (ss *StateSync) mergeState(peerState *ChallengerState) error {
    ss.mu.Lock()
    defer ss.mu.Unlock()

    // 버전 충돌 해결
    if peerState.Version <= ss.state.Version {
        return nil // 이미 최신 상태
    }

    // 게임 상태 병합
    for gameHash, peerGame := range peerState.ActiveGames {
        localGame, exists := ss.state.ActiveGames[gameHash]

        if !exists {
            // 새로운 게임 추가
            ss.state.ActiveGames[gameHash] = peerGame
        } else {
            // 기존 게임 업데이트 (최신 정보 우선)
            if peerGame.LastUpdate.After(localGame.LastUpdate) {
                ss.state.ActiveGames[gameHash] = peerGame
            }
        }
    }

    // 피어 정보 업데이트
    for peerID, peerInfo := range peerState.Peers {
        ss.state.Peers[peerID] = peerInfo
    }

    // 상태 버전 업데이트
    ss.state.Version = max(ss.state.Version, peerState.Version) + 1
    ss.state.LastUpdate = time.Now()

    return nil
}
```

### 2.2 게임 정보 공유

#### 게임 상태 브로드캐스트
```go
// op-challenger/p2p/sync/broadcast.go
type BroadcastService struct {
    node        *P2PNode
    state       *ChallengerState
    peers       map[string]*Peer

    // 브로드캐스트 설정
    broadcastInterval time.Duration
    maxRetries        int
}

func (bs *BroadcastService) BroadcastGameState(gameHash common.Hash, gameInfo *GameInfo) error {
    message := &Message{
        Type:      MessageTypeGameState,
        From:      bs.node.id,
        Payload:   bs.serializeGameState(gameHash, gameInfo),
        Timestamp: time.Now(),
        Nonce:     bs.generateNonce(),
    }

    // 메시지 서명
    signature, err := bs.node.signMessage(message)
    if err != nil {
        return fmt.Errorf("failed to sign message: %w", err)
    }
    message.Signature = signature

    // 모든 피어에게 브로드캐스트
    var wg sync.WaitGroup
    var errors []error
    var mu sync.Mutex

    for _, peer := range bs.peers {
        wg.Add(1)
        go func(p *Peer) {
            defer wg.Done()

            if err := bs.sendMessage(p, message); err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("failed to send to %s: %w", p.ID, err))
                mu.Unlock()
            }
        }(peer)
    }

    wg.Wait()

    if len(errors) > 0 {
        return fmt.Errorf("broadcast errors: %v", errors)
    }

    return nil
}

func (bs *BroadcastService) HandleGameStateMessage(message *Message) error {
    // 메시지 검증
    if err := bs.validateMessage(message); err != nil {
        return fmt.Errorf("invalid message: %w", err)
    }

    // 게임 상태 역직렬화
    gameHash, gameInfo, err := bs.deserializeGameState(message.Payload)
    if err != nil {
        return fmt.Errorf("failed to deserialize game state: %w", err)
    }

    // 로컬 상태 업데이트
    bs.state.mu.Lock()
    bs.state.ActiveGames[gameHash] = gameInfo
    bs.state.Version++
    bs.state.LastUpdate = time.Now()
    bs.state.mu.Unlock()

    // 다른 피어들에게 전달 (Gossip 프로토콜)
    return bs.gossipMessage(message)
}
```

## 🤝 3. 협력적 게임 해결

### 3.1 협력적 의사결정

#### 협력 에이전트 구조
```go
// op-challenger/p2p/challenger/cooperative.go
type CooperativeAgent struct {
    localAgent  *game.Agent
    peerAgents  map[string]*PeerAgent
    consensus   *ConsensusEngine

    // 협력 설정
    minConsensus int           // 최소 합의 수
    timeout      time.Duration // 합의 타임아웃
}

type PeerAgent struct {
    peerID       string
    reputation   float64
    lastResponse time.Time
    responses    map[common.Hash]*ValidationResponse
}

type ValidationResponse struct {
    GameHash     common.Hash `json:"game_hash"`
    Action       types.Action `json:"action"`
    Confidence   float64     `json:"confidence"`
    Reasoning    string      `json:"reasoning"`
    Timestamp    time.Time   `json:"timestamp"`
}

func (ca *CooperativeAgent) CooperativeAct(ctx context.Context, gameHash common.Hash) error {
    // 1. 로컬 의사결정
    localAction, err := ca.localAgent.DecideAction(ctx, gameHash)
    if err != nil {
        return fmt.Errorf("local decision failed: %w", err)
    }

    // 2. 피어들에게 의사결정 요청
    responses, err := ca.requestPeerDecisions(ctx, gameHash)
    if err != nil {
        ca.logger.Warn("Failed to get peer decisions, using local decision", "error", err)
        return ca.localAgent.Act(ctx)
    }

    // 3. 합의 도출
    consensusAction, confidence := ca.reachConsensus(localAction, responses)

    // 4. 합의 결과에 따른 액션 실행
    if confidence >= ca.minConsensus {
        ca.logger.Info("Consensus reached", "action", consensusAction, "confidence", confidence)
        return ca.executeConsensusAction(ctx, consensusAction)
    } else {
        ca.logger.Info("No consensus, using local decision", "confidence", confidence)
        return ca.localAgent.Act(ctx)
    }
}

func (ca *CooperativeAgent) reachConsensus(localAction types.Action, responses []*ValidationResponse) (types.Action, float64) {
    // 액션별 투표 집계
    actionVotes := make(map[types.Action]float64)
    totalWeight := 1.0 // 로컬 에이전트 가중치

    // 로컬 액션 투표
    actionVotes[localAction] += 1.0

    // 피어 응답 집계 (평판 기반 가중치)
    for _, response := range responses {
        peerAgent := ca.peerAgents[response.peerID]
        if peerAgent == nil {
            continue
        }

        weight := peerAgent.reputation
        actionVotes[response.Action] += weight
        totalWeight += weight
    }

    // 최다 득표 액션 찾기
    var bestAction types.Action
    var maxVotes float64

    for action, votes := range actionVotes {
        if votes > maxVotes {
            maxVotes = votes
            bestAction = action
        }
    }

    confidence := maxVotes / totalWeight
    return bestAction, confidence
}
```

### 3.2 중복 작업 방지

#### 작업 분배 시스템
```go
// op-challenger/p2p/challenger/work_distribution.go
type WorkDistributor struct {
    node        *P2PNode
    activeGames map[common.Hash]*GameWork
    peers       map[string]*Peer

    // 분배 설정
    distributionStrategy DistributionStrategy
    loadBalancing       bool
}

type GameWork struct {
    GameHash     common.Hash
    AssignedTo   string
    Status       WorkStatus
    LastUpdate   time.Time
    Priority     int
}

type DistributionStrategy int

const (
    StrategyRoundRobin DistributionStrategy = iota
    StrategyLoadBased
    StrategyReputationBased
    StrategyConsensus
)

func (wd *WorkDistributor) AssignGameWork(gameHash common.Hash) error {
    // 1. 현재 작업 상태 확인
    if work, exists := wd.activeGames[gameHash]; exists {
        if work.Status == WorkStatusInProgress {
            return fmt.Errorf("game work already assigned to %s", work.AssignedTo)
        }
    }

    // 2. 최적 피어 선택
    selectedPeer := wd.selectOptimalPeer(gameHash)
    if selectedPeer == "" {
        return fmt.Errorf("no suitable peer available")
    }

    // 3. 작업 할당
    wd.activeGames[gameHash] = &GameWork{
        GameHash:   gameHash,
        AssignedTo: selectedPeer,
        Status:     WorkStatusInProgress,
        LastUpdate: time.Now(),
        Priority:   wd.calculatePriority(gameHash),
    }

    // 4. 작업 할당 알림
    return wd.notifyWorkAssignment(gameHash, selectedPeer)
}

func (wd *WorkDistributor) selectOptimalPeer(gameHash common.Hash) string {
    switch wd.distributionStrategy {
    case StrategyRoundRobin:
        return wd.roundRobinSelect()
    case StrategyLoadBased:
        return wd.loadBasedSelect()
    case StrategyReputationBased:
        return wd.reputationBasedSelect()
    case StrategyConsensus:
        return wd.consensusSelect(gameHash)
    default:
        return wd.roundRobinSelect()
    }
}

func (wd *WorkDistributor) reputationBasedSelect() string {
    var bestPeer string
    var bestScore float64

    for peerID, peer := range wd.peers {
        score := peer.Reputation * wd.calculateLoadFactor(peerID)
        if score > bestScore {
            bestScore = score
            bestPeer = peerID
        }
    }

    return bestPeer
}
```

## 📊 4. 평판 시스템

### 4.1 평판 계산 및 관리

#### 평판 시스템 구조
```go
// op-challenger/p2p/reputation/system.go
type ReputationSystem struct {
    nodeID      string
    reputation  float64
    peers       map[string]*PeerReputation

    // 평판 계산 설정
    weights     ReputationWeights
    decayRate   float64
    minReputation float64
    maxReputation float64
}

type PeerReputation struct {
    PeerID      string
    Score       float64
    Validations int
    Failures    int
    LastUpdate  time.Time

    // 상세 기록
    validationHistory []*ValidationRecord
    penaltyHistory    []*PenaltyRecord
}

type ValidationRecord struct {
    GameHash     common.Hash
    Success      bool
    Timestamp    time.Time
    Impact       float64  // 검증 결과의 중요도
}

type ReputationWeights struct {
    RecentValidation   float64 // 최근 검증 가중치
    HistoricalAccuracy float64 // 역사적 정확도 가중치
    ResponseTime       float64 // 응답 시간 가중치
    Cooperation        float64 // 협력도 가중치
}

func (rs *ReputationSystem) UpdateReputation(peerID string, validation *ValidationResult) error {
    peer, exists := rs.peers[peerID]
    if !exists {
        peer = &PeerReputation{
            PeerID: peerID,
            Score:  1.0, // 초기 평판
        }
        rs.peers[peerID] = peer
    }

    // 검증 결과 기록
    record := &ValidationRecord{
        GameHash:  validation.GameHash,
        Success:   validation.Success,
        Timestamp: time.Now(),
        Impact:    rs.calculateImpact(validation),
    }

    peer.validationHistory = append(peer.validationHistory, record)

    // 평판 점수 업데이트
    oldScore := peer.Score
    peer.Score = rs.calculateNewScore(peer, record)

    // 평판 점수 범위 제한
    peer.Score = math.Max(rs.minReputation, math.Min(rs.maxReputation, peer.Score))

    // 통계 업데이트
    if validation.Success {
        peer.Validations++
    } else {
        peer.Failures++
    }

    peer.LastUpdate = time.Now()

    rs.logger.Info("Reputation updated",
        "peer", peerID,
        "old_score", oldScore,
        "new_score", peer.Score,
        "success", validation.Success)

    return nil
}

func (rs *ReputationSystem) calculateNewScore(peer *PeerReputation, record *ValidationRecord) float64 {
    // 1. 최근 검증 가중치 계산
    recentWeight := rs.weights.RecentValidation
    if record.Success {
        recentWeight *= 1.0
    } else {
        recentWeight *= -1.0
    }

    // 2. 역사적 정확도 계산
    historicalAccuracy := rs.calculateHistoricalAccuracy(peer)
    historicalWeight := rs.weights.HistoricalAccuracy * historicalAccuracy

    // 3. 응답 시간 가중치 계산
    responseTimeWeight := rs.calculateResponseTimeWeight(record)

    // 4. 협력도 가중치 계산
    cooperationWeight := rs.calculateCooperationWeight(peer)

    // 5. 종합 점수 계산
    newScore := peer.Score +
                recentWeight +
                historicalWeight +
                responseTimeWeight +
                cooperationWeight

    // 6. 시간에 따른 감쇠 적용
    timeDecay := rs.calculateTimeDecay(peer.LastUpdate)
    newScore *= timeDecay

    return newScore
}

func (rs *ReputationSystem) calculateHistoricalAccuracy(peer *PeerReputation) float64 {
    if len(peer.validationHistory) == 0 {
        return 0.5 // 중립적 초기값
    }

    recentValidations := peer.validationHistory[len(peer.validationHistory)-10:] // 최근 10개
    successCount := 0

    for _, record := range recentValidations {
        if record.Success {
            successCount++
        }
    }

    return float64(successCount) / float64(len(recentValidations))
}
```

### 4.2 악의적 행위 감지

#### 악의적 행위 감지 시스템
```go
// op-challenger/p2p/reputation/detection.go
type MaliciousBehaviorDetector struct {
    reputation  *ReputationSystem
    thresholds  DetectionThresholds
    penalties   map[string]*PenaltyRecord
}

type DetectionThresholds struct {
    MinReputation     float64 // 최소 평판 임계값
    MaxFailureRate    float64 // 최대 실패율
    SuspiciousPattern string  // 의심스러운 패턴
    PenaltyThreshold  float64 // 페널티 적용 임계값
}

type PenaltyRecord struct {
    PeerID      string
    Reason      string
    Severity    PenaltySeverity
    AppliedAt   time.Time
    Duration    time.Duration
    Active      bool
}

func (mbd *MaliciousBehaviorDetector) DetectMaliciousBehavior(peerID string, action *ValidationResult) *PenaltyRecord {
    peer := mbd.reputation.peers[peerID]
    if peer == nil {
        return nil
    }

    // 1. 평판 점수 확인
    if peer.Score < mbd.thresholds.MinReputation {
        return mbd.createPenalty(peerID, "Low reputation", PenaltySeverityMedium)
    }

    // 2. 실패율 확인
    failureRate := float64(peer.Failures) / float64(peer.Validations+peer.Failures)
    if failureRate > mbd.thresholds.MaxFailureRate {
        return mbd.createPenalty(peerID, "High failure rate", PenaltySeverityHigh)
    }

    // 3. 패턴 분석
    if mbd.detectSuspiciousPattern(peer) {
        return mbd.createPenalty(peerID, "Suspicious pattern detected", PenaltySeverityHigh)
    }

    // 4. 일관성 검사
    if mbd.detectInconsistency(peer, action) {
        return mbd.createPenalty(peerID, "Inconsistent behavior", PenaltySeverityMedium)
    }

    return nil
}

func (mbd *MaliciousBehaviorDetector) detectSuspiciousPattern(peer *PeerReputation) bool {
    if len(peer.validationHistory) < 5 {
        return false
    }

    // 최근 5개 검증의 패턴 분석
    recent := peer.validationHistory[len(peer.validationHistory)-5:]

    // 모든 검증이 실패하는 패턴
    allFailures := true
    for _, record := range recent {
        if record.Success {
            allFailures = false
            break
        }
    }

    if allFailures {
        return true
    }

    // 특정 시간대에만 실패하는 패턴
    // (구현 생략)

    return false
}
```

## 🔧 5. 구현 우선순위 및 단계

### Phase 1: 기본 P2P 기능 (4주)

#### Week 1: P2P 노드 관리
- [ ] P2P 노드 구조 구현
- [ ] 기본 네트워크 전송 계층 구현
- [ ] 핸드셰이크 프로토콜 구현
- [ ] 연결 관리 시스템 구현

#### Week 2: 노드 발견 시스템
- [ ] DHT 기반 노드 발견 구현
- [ ] RPC 메시지 시스템 구현
- [ ] 부트스트랩 노드 관리
- [ ] 노드 등록 및 해제

#### Week 3: 상태 동기화
- [ ] 챌린저 상태 정의
- [ ] 상태 동기화 프로토콜 구현
- [ ] 게임 정보 브로드캐스트
- [ ] 충돌 해결 메커니즘

#### Week 4: 기본 평판 시스템
- [ ] 평판 계산 알고리즘 구현
- [ ] 검증 결과 기록 시스템
- [ ] 기본 악의적 행위 감지
- [ ] 통합 테스트

### Phase 2: 고급 기능 (4주)

#### Week 5-6: 협력적 게임 해결
- [ ] 협력적 의사결정 시스템
- [ ] 작업 분배 알고리즘
- [ ] 합의 도출 메커니즘
- [ ] 중복 작업 방지

#### Week 7-8: 고급 평판 및 모니터링
- [ ] 고급 악의적 행위 감지
- [ ] 실시간 모니터링 시스템
- [ ] 성능 최적화
- [ ] 운영 도구 개발

## 📈 6. 성능 목표 및 지표

### 목표 성능
- **노드 발견 시간**: < 30초
- **상태 동기화 지연**: < 5초
- **평판 계산 지연**: < 1초
- **협력적 의사결정**: < 10초
- **네트워크 대역폭**: < 1MB/s (정상 상태)

### 모니터링 지표
- 활성 피어 수
- 평균 응답 시간
- 평판 점수 분포
- 협력 효과 측정
- 악의적 행위 감지율

---

**참고**: 이 문서는 P2P 챌린저 네트워크의 구체적인 구현 세부사항을 제공하며, 개발 과정에서 참조할 수 있는 기술적 가이드입니다.
