# Phase 1: 백업 완료 거래 풀 시스템

## 🎯 목표

다중 시퀀서 환경에서 거래 순서 일관성을 보장하는 백업 완료 거래 풀 시스템을 구축합니다.

## 📋 작업 개요

### 핵심 기능
- 메인 시퀀서가 거래를 완료 처리한 후 백업 시퀀서와 동기화
- 백업 시퀀서가 메인 시퀀서의 거래 순서를 정확히 복제
- 장애 발생 시 백업 시퀀서가 즉시 활성화
- 거래 순서 일관성 보장 및 중복 처리 방지

### 기술적 요구사항
- 실시간 거래 동기화 (지연 시간 < 1초)
- 장애 감지 및 자동 전환 (전환 시간 < 5초)
- 거래 무결성 검증
- 메모리 효율적인 거래 풀 관리

## 🏗️ 아키텍처 설계

### 시스템 구성
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   메인 시퀀서    │    │   백업 시퀀서    │    │   백업 시퀀서    │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ 거래 처리    │ │    │ │ 거래 풀     │ │    │ │ 거래 풀     │ │
│ │ 엔진        │ │    │ │ 동기화      │ │    │ │ 동기화      │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│         │       │    │         │       │    │         │       │
│         ▼       │    │         ▼       │    │         ▼       │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ 완료 거래    │ │───▶│ │ 백업 풀     │ │───▶│ │ 백업 풀     │ │
│ │ 브로드캐스트 │ │    │ │ 수신        │ │    │ │ 수신        │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 데이터 흐름
1. **거래 수신**: 메인 시퀀서가 L2 거래를 수신
2. **거래 처리**: 메인 시퀀서가 거래를 처리하고 블록에 포함
3. **완료 브로드캐스트**: 처리 완료된 거래를 백업 시퀀서들에게 브로드캐스트
4. **풀 동기화**: 백업 시퀀서들이 완료된 거래를 로컬 풀에 저장
5. **순서 보장**: 백업 시퀀서들이 메인 시퀀서의 거래 순서를 정확히 복제

## 📁 구현 파일 구조

```
op-node/rollup/sequencing/
├── backup_pool/
│   ├── pool.go              # 백업 거래 풀 핵심 구현
│   ├── sync.go              # 동기화 메커니즘
│   ├── broadcast.go         # 브로드캐스트 시스템
│   ├── validation.go        # 거래 검증 로직
│   └── pool_test.go         # 단위 테스트
├── backup_sequencer/
│   ├── sequencer.go         # 백업 시퀀서 구현
│   ├── activation.go        # 활성화 메커니즘
│   ├── health.go            # 헬스 체크
│   └── sequencer_test.go    # 테스트
└── integration/
    ├── conductor.go         # Conductor 통합
    ├── metrics.go           # 메트릭 수집
    └── config.go            # 설정 관리
```

## 🔧 구현 단계

### Step 1: 백업 거래 풀 핵심 구현 (1주)

#### 1.1 풀 데이터 구조 설계
```go
// pool.go
type BackupTransactionPool struct {
    mu          sync.RWMutex
    transactions map[common.Hash]*BackupTransaction
    order       []common.Hash
    maxSize     int
    metrics     *PoolMetrics
}

type BackupTransaction struct {
    Hash      common.Hash
    Data      []byte
    Timestamp time.Time
    Sequence  uint64
    Status    TransactionStatus
}
```

#### 1.2 풀 관리 기능 구현
- 거래 추가/제거
- 순서 보장 메커니즘
- 메모리 관리 (LRU 캐시)
- 중복 거래 방지

#### 1.3 동기화 프로토콜 구현
```go
// sync.go
type SyncProtocol struct {
    pool       *BackupTransactionPool
    broadcast  *BroadcastService
    validator  *TransactionValidator
}

func (sp *SyncProtocol) SyncTransaction(tx *BackupTransaction) error
func (sp *SyncProtocol) ValidateOrder(sequence uint64) error
```

### Step 2: 브로드캐스트 시스템 (1주)

#### 2.1 메인 시퀀서 브로드캐스트
```go
// broadcast.go
type BroadcastService struct {
    peers      map[string]*Peer
    pool       *BackupTransactionPool
    transport  Transport
}

func (bs *BroadcastService) BroadcastCompleted(tx *BackupTransaction) error
func (bs *BroadcastService) AddPeer(peer *Peer) error
```

#### 2.2 백업 시퀀서 수신
```go
func (bs *BroadcastService) HandleBroadcast(msg *BroadcastMessage) error
func (bs *BroadcastService) ValidateBroadcast(msg *BroadcastMessage) error
```

### Step 3: 백업 시퀀서 활성화 (1주)

#### 3.1 장애 감지
```go
// activation.go
type ActivationManager struct {
    healthChecker *HealthChecker
    pool          *BackupTransactionPool
    sequencer     *BackupSequencer
}

func (am *ActivationManager) MonitorMainSequencer() error
func (am *ActivationManager) ActivateBackup() error
```

#### 3.2 역할 전환
```go
func (am *ActivationManager) SwitchToMain() error
func (am *ActivationManager) SyncWithPool() error
```

### Step 4: Conductor 통합 (1주)

#### 4.1 Conductor 인터페이스 확장
```go
// conductor.go
type BackupPoolConductor struct {
    conductor  *conductor.Conductor
    pool       *BackupTransactionPool
    sequencer  *BackupSequencer
}

func (bpc *BackupPoolConductor) RegisterBackupSequencer() error
func (bpc *BackupPoolConductor) HandleRoleChange(role SequencerRole) error
```

#### 4.2 설정 및 메트릭
```go
// config.go
type BackupPoolConfig struct {
    MaxPoolSize    int           `json:"max_pool_size"`
    SyncTimeout    time.Duration `json:"sync_timeout"`
    BroadcastPeers []string      `json:"broadcast_peers"`
}
```

## 🧪 테스트 전략

### 단위 테스트
```bash
# 풀 기능 테스트
go test ./op-node/rollup/sequencing/backup_pool -v

# 동기화 테스트
go test ./op-node/rollup/sequencing/backup_pool -run TestSync -v

# 브로드캐스트 테스트
go test ./op-node/rollup/sequencing/backup_pool -run TestBroadcast -v
```

### 통합 테스트
```bash
# 전체 시스템 테스트
go test ./op-node/rollup/sequencing/integration -v

# 성능 테스트
go test ./op-node/rollup/sequencing/integration -run TestPerformance -v
```

### E2E 테스트
```bash
# 실제 네트워크 환경 테스트
cd op-e2e
go test -run TestBackupPoolE2E -v
```

## 📊 성능 지표

### 목표 성능
- **동기화 지연**: < 1초
- **전환 시간**: < 5초
- **메모리 사용량**: < 1GB (10,000 거래 기준)
- **CPU 사용률**: < 10% (정상 상태)

### 모니터링 메트릭
```go
type PoolMetrics struct {
    SyncLatency    prometheus.Histogram
    PoolSize       prometheus.Gauge
    BroadcastRate  prometheus.Counter
    ErrorRate      prometheus.Counter
}
```

## 🚀 배포 및 운영

### 개발 환경 설정
```bash
# 백업 풀 기능 활성화
export ENABLE_BACKUP_POOL=true
export BACKUP_POOL_SIZE=10000
export SYNC_TIMEOUT=1s

# op-node 실행
go run ./op-node/cmd/main.go --backup-pool.enabled=true
```

### 운영 환경 설정
```yaml
# op-node.yaml
backup_pool:
  enabled: true
  max_pool_size: 10000
  sync_timeout: "1s"
  broadcast_peers:
    - "backup-sequencer-1:8080"
    - "backup-sequencer-2:8080"
```

## 🔍 디버깅 및 문제 해결

### 일반적인 문제
1. **동기화 지연**: 네트워크 대역폭 확인
2. **메모리 누수**: 풀 크기 제한 확인
3. **전환 실패**: 헬스 체크 설정 확인

### 로그 분석
```bash
# 백업 풀 로그 확인
grep "backup_pool" op-node.log

# 동기화 상태 확인
grep "sync" op-node.log | tail -20
```

## 📚 참고 자료

- [backup-complete-transaction-pool.md](../../../backup-complete-transaction-pool.md)
- [optimism-conductor-overview.md](../../../optimism-conductor-overview.md)
- [sequencer-overview.md](../../../sequencer-overview.md)

## ✅ 완료 체크리스트

- [ ] 백업 거래 풀 데이터 구조 구현
- [ ] 동기화 프로토콜 구현
- [ ] 브로드캐스트 시스템 구현
- [ ] 백업 시퀀서 활성화 메커니즘 구현
- [ ] Conductor 통합 구현
- [ ] 단위 테스트 작성
- [ ] 통합 테스트 작성
- [ ] 성능 테스트 수행
- [ ] 문서화 완료
- [ ] 코드 리뷰 완료

---

**다음 단계**: [기본 P2P 챌린저 네트워크](../p2p-challenger-basic/README.md)
