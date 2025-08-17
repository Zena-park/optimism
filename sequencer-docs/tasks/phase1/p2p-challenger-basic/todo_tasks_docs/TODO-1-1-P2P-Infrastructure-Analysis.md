# TODO 1.1: 기존 P2P 인프라 분석

## 📋 작업 개요
- **작업명**: 기존 P2P 인프라 분석
- **상태**: ✅ 완료
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
`op-challenger/p2p/` 패키지의 기존 인프라를 분석하여 RAT(Randomized Attention Test) 구현을 위한 기반 구조를 파악한다.

## 🔍 분석 결과

### 1. 기존 P2P 시스템 구조

#### **1.1 P2PNode 구조체** (`op-challenger/p2p/network/node.go`)
```go
type P2PNode struct {
    // 기본 정보
    id         string            // 노드 고유 ID (공개키 기반)
    address    string            // 노드 주소 (IP:Port)
    publicKey  *ecdsa.PublicKey  // 공개키
    privateKey *ecdsa.PrivateKey // 개인키

    // 네트워크 관리
    peers     map[string]*Peer  // 연결된 피어들
    discovery *DiscoveryService // 노드 발견 서비스
    transport Transport         // 네트워크 전송 계층

    // 상태 관리
    status     NodeStatus // 노드 상태
    lastSeen   time.Time  // 마지막 활동 시간
    reputation float64    // 평판 점수
}
```

**주요 기능:**
- ✅ 노드 시작/중지 관리
- ✅ 피어 연결 및 관리 (최대 100개 피어)
- ✅ 메시지 브로드캐스팅
- ✅ 핸드셰이크 프로토콜
- ✅ 헬스체크 (30초마다)

#### **1.2 DiscoveryService** (`op-challenger/p2p/network/discovery.go`)
```go
type DiscoveryService struct {
    dht        *DHT                 // 분산 해시 테이블
    nodeID     string               // 현재 노드 ID
    bootstrap  []string             // 부트스트랩 노드
    knownNodes map[string]*NodeInfo // 알려진 노드들
}
```

**주요 기능:**
- ✅ DHT 기반 노드 발견
- ✅ 부트스트랩 노드 연결
- ✅ 주기적 노드 발견 (5초마다)
- ✅ 노드 정보 관리

#### **1.3 Transport Layer** (`op-challenger/p2p/network/transport.go`)
```go
type Transport interface {
    Start(address string) error
    Stop() error
    Accept() (Connection, error)
    Connect(address string) (Connection, error)
}
```

**주요 기능:**
- ✅ TCP 기반 전송 계층
- ✅ 연결 타임아웃 관리
- ✅ 버퍼 크기 설정 (4096 bytes)

### 2. 메시지 시스템 분석

#### **2.1 이미 정의된 RAT 관련 메시지** (`op-challenger/p2p/network/messages.go`)

**🎉 발견: RAT 관련 메시지 타입들이 이미 정의되어 있음!**

```go
// RAT (Randomized Attention Test) 관련 메시지
MessageTypeAttentionTestTriggered   // 어텐션 테스트 트리거
MessageTypeAttentionTestResponse    // 어텐션 테스트 응답
MessageTypeAttentionTestResult      // 어텐션 테스트 결과
MessageTypeChallengerRegistration   // 챌린저 등록
MessageTypeChallengerStatus         // 챌린저 상태
MessageTypeChallengerChallenge      // 챌린저 도전
```

**주요 메시지 구조체:**
```go
// 어텐션 테스트 트리거 메시지
type AttentionTestTriggeredMessage struct {
    TestID           string    `json:"test_id"`
    StateRoot        string    `json:"state_root"`
    L2BlockNumber    uint64    `json:"l2_block_number"`
    TargetChallenger string    `json:"target_challenger"`
    ResponseWindow   uint64    `json:"response_window"`
    Timestamp        time.Time `json:"timestamp"`
}

// 어텐션 테스트 응답 메시지
type AttentionTestResponseMessage struct {
    TestID     string    `json:"test_id"`
    Challenger string    `json:"challenger"`
    LeftChild  string    `json:"left_child"`
    RightChild string    `json:"right_child"`
    Timestamp  time.Time `json:"timestamp"`
}

// 챌린저 등록 메시지
type ChallengerRegistrationMessage struct {
    ChallengerID string    `json:"challenger_id"`
    Address      string    `json:"address"`
    Deposit      uint64    `json:"deposit"`
    PublicKey    []byte    `json:"public_key"`
    Timestamp    time.Time `json:"timestamp"`
}
```

### 3. 활용 가능한 기존 인프라

#### **3.1 즉시 활용 가능한 기능들**
- ✅ **P2P 네트워크 기반 구조**: 기존 노드 연결 및 관리 시스템
- ✅ **메시지 시스템**: RAT 관련 메시지 타입들이 이미 정의됨
- ✅ **피어 관리**: 자동 연결/재연결, 헬스체크
- ✅ **보안**: ECDSA 기반 공개키/개인키 시스템
- ✅ **평판 시스템 기반**: 기본 평판 점수 관리

#### **3.2 확장이 필요한 영역**
- 🔧 **챌린저 특화 메타데이터**: 챌린저 역할, 스테이킹 정보 등
- 🔧 **어텐션 테스트 로직**: 실제 테스트 실행 및 검증 로직
- 🔧 **평판 시스템 고도화**: 어텐션 테스트 결과 기반 평판 업데이트
- 🔧 **DDoS 방어**: 레이트 리미팅, 스테이킹 기반 인증

## 📊 기존 인프라 평가

| 영역 | 기존 구현 수준 | 활용도 | 추가 작업 필요도 |
|------|-------------|--------|----------------|
| **P2P 네트워크** | 🟢 완성도 높음 | 🟢 높음 | 🟡 중간 |
| **메시지 시스템** | 🟢 RAT 메시지 정의됨 | 🟢 높음 | 🟡 중간 |
| **노드 발견** | 🟢 DHT 기반 구현 | 🟢 높음 | 🟡 중간 |
| **보안** | 🟡 기본 구현 | 🟡 중간 | 🔴 높음 |
| **평판 시스템** | 🟡 기본 구조만 | 🟡 중간 | 🔴 높음 |

## 🚀 다음 단계 권장사항

### 1. 우선순위 높음
1. **챌린저 관리 시스템 설계** (TODO 1.2)
   - 기존 P2PNode 확장하여 챌린저 특화 기능 추가
   - 챌린저 메타데이터 구조 정의

2. **어텐션 테스트 실행 로직**
   - 기존 메시지 구조를 활용한 실제 테스트 로직 구현

### 2. 우선순위 중간
1. **평판 시스템 고도화**
   - 어텐션 테스트 결과 기반 평판 업데이트
   - 담합 탐지 알고리즘

2. **DDoS 방어 시스템**
   - 레이트 리미팅
   - 스테이킹 기반 인증

## 🔗 관련 파일들
- `op-challenger/p2p/network/node.go` - P2P 노드 구현
- `op-challenger/p2p/network/discovery.go` - 노드 발견 서비스
- `op-challenger/p2p/network/messages.go` - 메시지 시스템 (RAT 메시지 포함)
- `op-challenger/p2p/network/transport.go` - 전송 계층
- `op-challenger/p2p/network/types.go` - 기본 타입 정의

## ✅ 완료 체크리스트
- [x] `op-challenger/p2p/network/` 패키지 분석
- [x] `node.go`의 `P2PNode` 구조 이해
- [x] `discovery.go`의 노드 발견 메커니즘 분석
- [x] `messages.go`의 메시지 시스템 분석 (RAT 메시지 발견!)
- [x] `transport.go`의 전송 계층 분석
- [x] `op-node/p2p/` 시스템과의 연관성 파악
- [x] 기존 libp2p 기반 인프라 활용 방안 검토

## 📝 결론
기존 P2P 인프라가 예상보다 잘 구축되어 있으며, 특히 RAT 관련 메시지 타입들이 이미 정의되어 있어서 구현 작업이 크게 단순화될 것으로 예상됩니다. 다음 단계에서는 이 기존 인프라를 활용하여 챌린저 관리 시스템을 설계하고 구현하면 됩니다.
