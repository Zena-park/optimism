# Phase 1: 기본 P2P 챌린저 네트워크

## 🎯 목표

완전한 분산화된 챌린저 네트워크의 기반을 구축하여 중앙 권한 없이 챌린저들이 상호 관리할 수 있는 시스템을 개발합니다.

## 📋 작업 개요

### 핵심 기능
- P2P 네트워크 기반 챌린저 발견 및 연결
- 챌린저 간 상태 동기화 및 정보 공유
- 온체인 검증 우선 원칙으로 다수결 최소화
- 기본적인 챌린저 평판 시스템

### 기술적 요구사항
- P2P 네트워크 자동 발견 (DHT 기반)
- 실시간 챌린저 상태 동기화
- 온체인 검증 결과 공유
- 기본 보안 및 인증 메커니즘

## 🏗️ 아키텍처 설계

### 시스템 구성
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   챌린저 A      │    │   챌린저 B      │    │   챌린저 C      │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ P2P 노드    │ │◄──►│ │ P2P 노드    │ │◄──►│ │ P2P 노드    │ │
│ │ 관리        │ │    │ │ 관리        │ │    │ │ 관리        │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│         │       │    │         │       │    │         │       │
│         ▼       │    │         ▼       │    │         ▼       │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ 챌린저      │ │    │ │ 챌린저      │ │    │ │ 챌린저      │ │
│ │ 엔진        │ │    │ │ 엔진        │ │    │ │ 엔진        │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 데이터 흐름
1. **노드 발견**: DHT를 통한 챌린저 노드 자동 발견
2. **연결 수립**: P2P 프로토콜을 통한 직접 연결
3. **상태 동기화**: 챌린저 상태 및 검증 결과 공유
4. **정보 교환**: 온체인 검증 결과 및 평판 정보 교환

## 📁 구현 파일 구조

```
op-challenger/p2p/
├── network/
│   ├── node.go              # P2P 노드 핵심 구현
│   ├── discovery.go         # DHT 기반 노드 발견
│   ├── connection.go        # 연결 관리
│   └── protocol.go          # P2P 프로토콜
├── challenger/
│   ├── engine.go            # 챌린저 엔진
│   ├── state.go             # 상태 관리
│   ├── validation.go        # 검증 로직
│   └── reputation.go        # 평판 시스템
├── sync/
│   ├── sync.go              # 상태 동기화
│   ├── message.go           # 메시지 정의
│   └── handler.go           # 메시지 핸들러
└── security/
    ├── auth.go              # 인증 메커니즘
    ├── encryption.go        # 암호화
    └── verification.go      # 검증
```

## 🔧 구현 단계

### Step 1: P2P 네트워크 기반 구축 (1주)

#### 1.1 P2P 노드 구현
```go
// node.go
type P2PNode struct {
    id          string
    address     string
    peers       map[string]*Peer
    discovery   *DiscoveryService
    challenger  *ChallengerEngine
    sync        *StateSync
}

type Peer struct {
    ID      string
    Address string
    Status  PeerStatus
    Reputation float64
}
```

#### 1.2 DHT 기반 노드 발견
```go
// discovery.go
type DiscoveryService struct {
    dht         *DHT
    nodeID      string
    bootstrap   []string
}

func (ds *DiscoveryService) StartDiscovery() error
func (ds *DiscoveryService) FindPeers() ([]*Peer, error)
func (ds *DiscoveryService) AnnouncePresence() error
```

#### 1.3 연결 관리
```go
// connection.go
type ConnectionManager struct {
    node        *P2PNode
    connections map[string]*Connection
    listener    net.Listener
}

func (cm *ConnectionManager) Connect(peer *Peer) error
func (cm *ConnectionManager) AcceptConnections() error
func (cm *ConnectionManager) HandleConnection(conn net.Conn) error
```

### Step 2: 챌린저 엔진 통합 (1주)

#### 2.1 챌린저 엔진 확장
```go
// engine.go
type ChallengerEngine struct {
    p2pNode     *P2PNode
    state       *ChallengerState
    validator   *Validator
    reputation  *ReputationSystem
}

func (ce *ChallengerEngine) StartP2P() error
func (ce *ChallengerEngine) ShareValidation(result *ValidationResult) error
func (ce *ChallengerEngine) GetPeerStatus() map[string]*PeerStatus
```

#### 2.2 상태 관리
```go
// state.go
type ChallengerState struct {
    NodeID          string
    LastValidation  time.Time
    ActiveGames     []common.Hash
    Reputation      float64
    Peers           map[string]*PeerInfo
}

type PeerInfo struct {
    ID              string
    LastSeen        time.Time
    ValidationCount int
    Reputation      float64
}
```

### Step 3: 상태 동기화 시스템 (1주)

#### 3.1 동기화 프로토콜
```go
// sync.go
type StateSync struct {
    node        *P2PNode
    state       *ChallengerState
    peers       map[string]*Peer
}

func (ss *StateSync) SyncWithPeer(peer *Peer) error
func (ss *StateSync) BroadcastState() error
func (ss *StateSync) HandleStateUpdate(update *StateUpdate) error
```

#### 3.2 메시지 정의
```go
// message.go
type Message struct {
    Type      MessageType
    From      string
    To        string
    Payload   []byte
    Timestamp time.Time
    Signature []byte
}

type MessageType int

const (
    MessageTypeStateUpdate MessageType = iota
    MessageTypeValidationResult
    MessageTypeReputationUpdate
    MessageTypePeerDiscovery
)
```

### Step 4: 기본 보안 및 평판 시스템 (1주)

#### 4.1 인증 메커니즘
```go
// auth.go
type AuthManager struct {
    privateKey *ecdsa.PrivateKey
    publicKey  *ecdsa.PublicKey
    peers      map[string]*PeerAuth
}

func (am *AuthManager) SignMessage(message []byte) ([]byte, error)
func (am *AuthManager) VerifyMessage(message []byte, signature []byte, publicKey *ecdsa.PublicKey) error
```

#### 4.2 평판 시스템
```go
// reputation.go
type ReputationSystem struct {
    nodeID      string
    reputation  float64
    peers       map[string]*PeerReputation
}

type PeerReputation struct {
    PeerID      string
    Score       float64
    Validations int
    LastUpdate  time.Time
}

func (rs *ReputationSystem) UpdateReputation(peerID string, validation *ValidationResult) error
func (rs *ReputationSystem) GetReputation(peerID string) float64
```

## 🧪 테스트 전략

### 단위 테스트
```bash
# P2P 네트워크 테스트
go test ./op-challenger/p2p/network -v

# 챌린저 엔진 테스트
go test ./op-challenger/p2p/challenger -v

# 동기화 테스트
go test ./op-challenger/p2p/sync -v
```

### 통합 테스트
```bash
# 전체 P2P 시스템 테스트
go test ./op-challenger/p2p -v

# 다중 노드 시뮬레이션
go test ./op-challenger/p2p -run TestMultiNode -v
```

### 네트워크 테스트
```bash
# 실제 네트워크 환경 테스트
cd op-e2e
go test -run TestP2PChallengerNetwork -v
```

## 📊 성능 지표

### 목표 성능
- **노드 발견 시간**: < 30초
- **상태 동기화 지연**: < 5초
- **연결 수립 시간**: < 10초
- **메시지 전송 지연**: < 1초

### 모니터링 메트릭
```go
type P2PMetrics struct {
    ActivePeers       prometheus.Gauge
    DiscoveryTime     prometheus.Histogram
    SyncLatency       prometheus.Histogram
    MessageRate       prometheus.Counter
    ReputationScore   prometheus.Gauge
}
```

## 🚀 배포 및 운영

### 개발 환경 설정
```bash
# P2P 챌린저 기능 활성화
export ENABLE_P2P_CHALLENGER=true
export P2P_PORT=8080
export BOOTSTRAP_NODES="node1:8080,node2:8080"

# op-challenger 실행
go run ./op-challenger/cmd/main.go --p2p.enabled=true
```

### 운영 환경 설정
```yaml
# op-challenger.yaml
p2p:
  enabled: true
  port: 8080
  bootstrap_nodes:
    - "challenger-1:8080"
    - "challenger-2:8080"
  discovery_timeout: "30s"
  sync_interval: "5s"
```

## 🔍 디버깅 및 문제 해결

### 일반적인 문제
1. **노드 발견 실패**: 부트스트랩 노드 설정 확인
2. **연결 실패**: 방화벽 및 포트 설정 확인
3. **동기화 지연**: 네트워크 대역폭 확인

### 로그 분석
```bash
# P2P 네트워크 로그 확인
grep "p2p" op-challenger.log

# 노드 발견 로그 확인
grep "discovery" op-challenger.log | tail -20
```

## 📚 참고 자료

- [optimism-challenger-systems.md](../../../optimism-challenger-systems.md)
- [optimism-challenger-management.md](../../../optimism-challenger-management.md)
- [backup-sequencer-challenger-analysis.md](../../../backup-sequencer-challenger-analysis.md)

## ✅ 완료 체크리스트

- [ ] P2P 노드 핵심 구현
- [ ] DHT 기반 노드 발견 구현
- [ ] 연결 관리 시스템 구현
- [ ] 챌린저 엔진 P2P 통합
- [ ] 상태 동기화 시스템 구현
- [ ] 기본 보안 메커니즘 구현
- [ ] 평판 시스템 구현
- [ ] 단위 테스트 작성
- [ ] 통합 테스트 작성
- [ ] 네트워크 테스트 수행
- [ ] 문서화 완료
- [ ] 코드 리뷰 완료

---

**다음 단계**: [Phase 2 - 백업 시퀀서 + 챌린저 통합](../../phase2/backup-challenger-integration/README.md)
