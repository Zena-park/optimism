# Optimism Conductor 시스템 개요

## 목차
1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [핵심 컴포넌트](#핵심-컴포넌트)
4. [리더 선출 메커니즘](#리더-선출-메커니즘)
5. [상태 관리](#상태-관리)
6. [장애 대응](#장애-대응)
7. [API 및 인터페이스](#api-및-인터페이스)
8. [설정 및 운영](#설정-및-운영)
9. [실제 사용 예시](#실제-사용-예시)

## 개요

### Conductor란?
Optimism Conductor는 **다중 시퀀서 환경에서 리더 선출과 장애 대응을 관리하는 핵심 시스템**입니다. Raft 합의 프로토콜을 기반으로 하여 안정적이고 신뢰할 수 있는 다중 시퀀서 운영을 보장합니다.

### 주요 기능
- **리더 선출**: Raft 합의 기반 안정적인 리더 선출
- **자동 장애 대응**: 시퀀서 장애 시 자동 리더십 전환
- **상태 모니터링**: 실시간 시퀀서 상태 추적
- **클러스터 관리**: 동적 노드 추가/제거
- **고가용성**: 무중단 서비스 보장

### 핵심 개념
- **리더**: 현재 활성 시퀀서 (블록 생성 권한 보유)
- **팔로워**: 백업 시퀀서 (리더 대기 상태)
- **클러스터**: 여러 Conductor 노드로 구성된 그룹
- **합의**: Raft 프로토콜을 통한 리더 선출 및 상태 동기화

## 아키텍처

### 전체 구조
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Conductor 1   │    │   Conductor 2   │    │   Conductor 3   │
│  (Raft Node)    │◄──►│  (Raft Node)    │◄──►│  (Raft Node)    │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Raft      │ │    │ │   Raft      │ │    │ │   Raft      │ │
│ │ Consensus   │ │    │ │ Consensus   │ │    │ │ Consensus   │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Health      │ │    │ │ Health      │ │    │ │ Health      │ │
│ │ Monitor     │ │    │ │ Monitor     │ │    │ │ Monitor     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ RPC Server  │ │    │ │ RPC Server  │ │    │ │ RPC Server  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────┬───────────┴───────────┬───────────┘
                     │                       │
         ┌─────────────────────────────────────────────────┐
         │              op-node (시퀀서)                   │
         │                                                 │
         │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
         │ │ Sequencer 1 │ │ Sequencer 2 │ │ Sequencer 3 │ │
         │ │ (Leader)    │ │ (Follower)  │ │ (Follower)  │ │
         │ └─────────────┘ └─────────────┘ └─────────────┘ │
         └─────────────────────────────────────────────────┘
```

### 컴포넌트 구성
1. **Raft Consensus**: 리더 선출 및 상태 동기화
2. **Health Monitor**: 시퀀서 상태 모니터링
3. **RPC Server**: API 제공 및 통신
4. **State Manager**: 상태 관리 및 전환 로직
5. **Metrics**: 성능 지표 수집

## 핵심 컴포넌트

### 1. Raft Consensus

#### Raft 합의 구현
```go
// op-conductor/consensus/raft.go
type RaftConsensus struct {
    log       log.Logger
    rollupCfg *rollup.Config

    serverID raft.ServerID
    r        *raft.Raft

    transport *raft.NetworkTransport
    advertisedAddr string

    unsafeTracker *unsafeHeadTracker
}

// 리더 선출 상태 확인
func (rc *RaftConsensus) Leader() bool {
    return rc.r.State() == raft.Leader
}

// 리더 정보 조회
func (rc *RaftConsensus) LeaderWithID() *ServerInfo {
    addr, id := rc.r.LeaderWithID()
    return &ServerInfo{
        ID:       string(id),
        Addr:     string(addr),
        Suffrage: Voter,
    }
}

// 리더십 전환
func (rc *RaftConsensus) TransferLeader() error {
    return rc.r.LeadershipTransfer().Error()
}
```

#### 클러스터 멤버십 관리
```go
// op-conductor/consensus/iface.go
type ClusterMembership struct {
    Servers []ServerInfo `json:"servers"`
    Version uint64       `json:"version"`
}

type ServerInfo struct {
    ID       string         `json:"id"`
    Addr     string         `json:"addr"`
    Suffrage ServerSuffrage `json:"suffrage"`
}

type ServerSuffrage int

const (
    Voter    ServerSuffrage = iota // 투표권 보유 (리더 선출 가능)
    Nonvoter                       // 투표권 없음 (리더 선출 불가)
)
```

### 2. Health Monitor

#### 상태 모니터링
```go
// op-conductor/health/health.go
type HealthMonitor struct {
    sequencerClient client.SequencerControl
    log             log.Logger

    healthUpdateCh chan error
    stopCh         chan struct{}
}

// 시퀀서 상태 확인
func (hm *HealthMonitor) checkSequencerHealth() error {
    // 1. 네트워크 연결 확인
    if err := hm.checkConnectivity(); err != nil {
        return fmt.Errorf("connectivity check failed: %w", err)
    }

    // 2. 블록 생성 상태 확인
    if err := hm.checkBlockGeneration(); err != nil {
        return fmt.Errorf("block generation check failed: %w", err)
    }

    // 3. 동기화 상태 확인
    if err := hm.checkSyncStatus(); err != nil {
        return fmt.Errorf("sync status check failed: %w", err)
    }

    return nil
}
```

### 3. State Manager

#### 상태 관리 로직
```go
// op-conductor/conductor/service.go
type OpConductor struct {
    leader         atomic.Bool
    leaderOverride atomic.Bool
    seqActive      atomic.Bool
    healthy        atomic.Bool

    healthUpdateCh <-chan error
    leaderUpdateCh <-chan bool

    ctrl client.SequencerControl
    cons consensus.Consensus
    hmon health.HealthMonitor
}

// 상태 기반 액션 실행
func (oc *OpConductor) action() {
    status := NewState(oc.leader.Load(), oc.healthy.Load(), oc.seqActive.Load())

    switch {
    case status.leader && status.healthy && !status.active:
        // 리더이지만 시퀀서가 비활성화된 경우 → 시퀀서 시작
        err = oc.startSequencer()

    case status.leader && !status.healthy && status.active:
        // 리더이지만 비정상이고 시퀀서가 활성화된 경우 → 리더십 전환
        err = oc.transferLeader()

    case !status.leader && status.healthy && status.active:
        // 팔로워이지만 시퀀서가 활성화된 경우 → 시퀀서 중지
        err = oc.stopSequencer()

    case status.leader && status.healthy && status.active:
        // 정상 리더 상태 → 아무것도 하지 않음
    }
}
```

## 리더 선출 메커니즘

### 1. Raft 합의 프로토콜

#### 기본 원리
Raft는 **리더 기반 합의 알고리즘**으로, 다음과 같은 특징을 가집니다:

- **단일 리더**: 언제든지 하나의 노드만 리더
- **리더 선출**: 투표를 통한 리더 선출
- **로그 복제**: 리더가 로그를 팔로워들에게 복제
- **안전성**: 로그가 커밋되면 모든 노드에 적용

#### 선출 과정
```go
// 1. 초기화 시 Bootstrap 설정
raftConsensusConfig := &consensus.RaftConsensusConfig{
    ServerID: c.cfg.RaftServerID,
    Bootstrap: c.cfg.RaftBootstrap, // 첫 번째 노드만 true
    ListenAddr: c.cfg.ConsensusAddr,
    ListenPort: c.cfg.ConsensusPort,
}

// 2. 리더 선출
cons, err := consensus.NewRaftConsensus(log, raftConsensusConfig)
if err != nil {
    if !errors.Is(err, raft.ErrCantBootstrap) {
        return errors.Wrap(err, "failed to create raft consensus")
    }
}

// 3. 리더십 상태 모니터링
oc.leaderUpdateCh = oc.cons.LeaderCh()
```

### 2. 클러스터 구성

#### 노드 추가
```go
// 투표권 있는 노드 추가 (리더 선출 가능)
func (oc *OpConductor) AddServerAsVoter(ctx context.Context, id string, addr string, version uint64) error {
    return oc.cons.AddVoter(id, addr, version)
}

// 투표권 없는 노드 추가 (리더 선출 불가)
func (oc *OpConductor) AddServerAsNonvoter(ctx context.Context, id string, addr string, version uint64) error {
    return oc.cons.AddNonVoter(id, addr, version)
}
```

#### 노드 제거
```go
// 노드 제거 (투표권 여부와 관계없이)
func (oc *OpConductor) RemoveServer(ctx context.Context, id string, version uint64) error {
    return oc.cons.RemoveServer(id, version)
}
```

### 3. 리더십 전환

#### 자동 전환
```go
// 장애 감지 시 자동 리더십 전환
func (oc *OpConductor) transferLeader() error {
    oc.log.Info("transferring leadership", "server", oc.cons.ServerID())
    err := oc.cons.TransferLeader()
    oc.metrics.RecordLeaderTransfer(err == nil)

    if err == nil {
        oc.leader.Store(false)
        return nil
    }

    switch {
    case errors.Is(err, raft.ErrNotLeader):
        oc.log.Warn("cannot transfer leadership since current server is not the leader")
        return nil
    default:
        oc.log.Error("failed to transfer leadership", "err", err)
        return err
    }
}
```

#### 수동 전환
```go
// 특정 서버로 리더십 전환
func (oc *OpConductor) TransferLeaderToServer(ctx context.Context, id string, addr string) error {
    return oc.cons.TransferLeaderTo(id, addr)
}
```

## 상태 관리

### 1. 상태 정의

#### 상태 구조
```go
type state struct {
    leader, healthy, active bool
}

func NewState(leader, healthy, active bool) *state {
    return &state{
        leader:  leader,
        healthy: healthy,
        active:  active,
    }
}
```

#### 상태 조합 (8가지 경우)
```go
// 8가지 상태 조합에 대한 처리
switch {
case !status.leader && !status.healthy && !status.active:
    // 팔로워, 비정상, 시퀀서 비활성화 → 에러 로그만

case !status.leader && !status.healthy && status.active:
    // 팔로워, 비정상, 시퀀서 활성화 → 시퀀서 중지

case !status.leader && status.healthy && !status.active:
    // 팔로워, 정상, 시퀀서 비활성화 → 정상 상태

case !status.leader && status.healthy && status.active:
    // 팔로워, 정상, 시퀀서 활성화 → 시퀀서 중지

case status.leader && !status.healthy && !status.active:
    // 리더, 비정상, 시퀀서 비활성화 → 리더십 전환 또는 시퀀서 시작

case status.leader && !status.healthy && status.active:
    // 리더, 비정상, 시퀀서 활성화 → 리더십 전환

case status.leader && status.healthy && !status.active:
    // 리더, 정상, 시퀀서 비활성화 → 시퀀서 시작

case status.leader && status.healthy && status.active:
    // 리더, 정상, 시퀀서 활성화 → 정상 상태
}
```

### 2. 상태 전환

#### 상태 변경 감지
```go
// 리더십 변경 감지
func (oc *OpConductor) handleLeaderUpdate(leader bool) {
    oc.log.Info("Leadership status changed", "server", oc.cons.ServerID(), "leader", leader)
    oc.leader.Store(leader)
    oc.queueAction()
}

// 헬스 상태 변경 감지
func (oc *OpConductor) handleHealthUpdate(hcerr error) {
    healthy := hcerr == nil
    if !healthy {
        oc.log.Error("Sequencer is unhealthy", "server", oc.cons.ServerID(), "err", hcerr)
        oc.queueAction()
    }

    if old := oc.healthy.Swap(healthy); old != healthy {
        oc.log.Info("Health state changed", "old", old, "new", healthy)
        oc.queueAction()
    }

    oc.hcerr = hcerr
}
```

## 장애 대응

### 1. 장애 감지

#### 헬스 체크
```go
// 시퀀서 헬스 체크
func (hm *HealthMonitor) checkSequencerHealth() error {
    // 1. 네트워크 연결 확인
    if err := hm.checkConnectivity(); err != nil {
        return fmt.Errorf("connectivity check failed: %w", err)
    }

    // 2. 블록 생성 상태 확인
    if err := hm.checkBlockGeneration(); err != nil {
        return fmt.Errorf("block generation check failed: %w", err)
    }

    // 3. 동기화 상태 확인
    if err := hm.checkSyncStatus(); err != nil {
        return fmt.Errorf("sync status check failed: %w", err)
    }

    return nil
}

// 블록 생성 상태 확인
func (hm *HealthMonitor) checkBlockGeneration() error {
    latestBlock, err := hm.sequencerClient.LatestBlock()
    if err != nil {
        return fmt.Errorf("failed to get latest block: %w", err)
    }

    // 30초 내에 블록이 생성되었는지 확인
    blockTime := time.Unix(int64(latestBlock.Time()), 0)
    if time.Since(blockTime) > 30*time.Second {
        return fmt.Errorf("block generation is too slow: last block at %v", blockTime)
    }

    return nil
}
```

### 2. 자동 복구

#### 장애 시나리오별 대응
```go
// 시나리오 1: 리더가 비정상이고 시퀀서가 비활성화된 경우
case status.leader && !status.healthy && !status.active:
    // 1. 이전에 팔로워였고 시퀀서가 비활성화된 상태에서 리더가 된 경우
    if !oc.prevState.leader && !oc.prevState.active &&
       !errors.Is(oc.hcerr, health.ErrSequencerConnectionDown) {
        // 시퀀서 시작 시도
        err = oc.startSequencer()
        if err != nil {
            oc.log.Error("failed to start sequencer, transferring leadership instead", "err", err)
        } else {
            break
        }
    }

    // 2. 다른 경우에는 리더십 전환
    err = oc.transferLeader()

// 시나리오 2: 리더가 비정상이고 시퀀서가 활성화된 경우
case status.leader && !status.healthy && status.active:
    // 리더십 전환 시도
    err = oc.transferLeader()
```

### 3. 재해 복구

#### 리더 오버라이드
```go
// 재해 상황에서 리더 오버라이드
func (oc *OpConductor) OverrideLeader(override bool) {
    oc.leaderOverride.Store(override)
}

func (oc *OpConductor) Leader(ctx context.Context) bool {
    return oc.LeaderOverridden() || oc.cons.Leader()
}
```

## API 및 인터페이스

### 1. RPC API

#### 주요 API 엔드포인트
```go
// op-conductor/rpc/api.go
type APIBackend struct {
    conductor *OpConductor
    log       log.Logger
}

// 리더십 상태 확인
func (api *APIBackend) Leader(ctx context.Context) (bool, error) {
    return api.conductor.Leader(ctx), nil
}

// 리더 정보 조회
func (api *APIBackend) LeaderWithID(ctx context.Context) (*consensus.ServerInfo, error) {
    return api.conductor.LeaderWithID(ctx), nil
}

// 클러스터 멤버십 조회
func (api *APIBackend) ClusterMembership(ctx context.Context) (*consensus.ClusterMembership, error) {
    return api.conductor.ClusterMembership(ctx)
}

// 시퀀서 상태 확인
func (api *APIBackend) SequencerHealthy(ctx context.Context) bool {
    return api.conductor.SequencerHealthy(ctx)
}
```

### 2. 클라이언트 인터페이스

#### op-node에서의 사용
```go
// op-node/node/conductor.go
type ConductorClient struct {
    cfg     *config.Config
    metrics *metrics.Metrics
    log     log.Logger

    apiClient locks.RWValue[*conductorRpc.APIClient]
    overrideLeader atomic.Bool
}

// 리더십 확인
func (c *ConductorClient) Leader(ctx context.Context) (bool, error) {
    if c.overrideLeader.Load() {
        return true, nil
    }

    if err := c.initialize(ctx); err != nil {
        return false, err
    }

    ctx, cancel := context.WithTimeout(ctx, c.cfg.ConductorRpcTimeout)
    defer cancel()

    isLeader, err := retry.Do(ctx, 2, retry.Fixed(50*time.Millisecond), func() (bool, error) {
        result, err := c.apiClient.Get().Leader(ctx)
        if err != nil {
            c.log.Error("Failed to check conductor for leadership", "err", err)
        }
        return result, err
    })

    return isLeader, err
}
```

## 설정 및 운영

### 1. 설정 파일

#### Conductor 설정
```yaml
# conductor.yaml
conductor:
  # Raft 합의 설정
  raft:
    server_id: "conductor-1"
    bootstrap: true  # 첫 번째 노드만 true
    listen_addr: "0.0.0.0"
    listen_port: 8081
    storage_dir: "/var/lib/conductor/raft"

    # 성능 튜닝
    heartbeat_timeout: "1s"
    leader_lease_timeout: "500ms"
    snapshot_interval: "10s"
    snapshot_threshold: 1000
    trailing_logs: 1000

  # RPC 설정
  rpc:
    listen_addr: "0.0.0.0"
    listen_port: 8080
    enable_proxy: true

  # 메트릭스 설정
  metrics:
    enabled: true
    listen_addr: "0.0.0.0"
    listen_port: 9090

  # 헬스 체크 설정
  health:
    check_interval: "5s"
    timeout: "3s"

  # 시퀀서 연결 설정
  sequencer:
    execution_rpc: "http://localhost:8545"
    node_rpc: "http://localhost:8547"
```

### 2. 실행 명령

#### Conductor 실행
```bash
# Conductor 서비스 시작
op-conductor \
  --config=conductor.yaml \
  --raft.server-id=conductor-1 \
  --raft.bootstrap=true \
  --raft.listen-addr=0.0.0.0 \
  --raft.listen-port=8081 \
  --rpc.listen-addr=0.0.0.0 \
  --rpc.listen-port=8080 \
  --metrics.enabled=true \
  --metrics.listen-port=9090
```

#### op-node에서 Conductor 사용
```bash
# Conductor와 연동하여 op-node 실행
op-node \
  --conductor.enabled=true \
  --conductor.rpc=http://conductor:8080 \
  --conductor.rpc-timeout=5s \
  --sequencer.enabled=true \
  --l1=http://l1:8545 \
  --l2=http://l2:8545
```

### 3. 모니터링

#### 메트릭스
```go
// op-conductor/metrics/metrics.go
type Metricer interface {
    RecordInfo(version string)
    RecordUp()
    RecordLeaderTransfer(success bool)
    RecordStateChange(leader, healthy, active bool)
    RecordLoopExecutionTime(duration float64)
    RecordUnsafePayloadCommit()
}
```

#### 주요 지표
- **리더십 전환 횟수**: `conductor_leader_transfer_total`
- **상태 변경 횟수**: `conductor_state_change_total`
- **루프 실행 시간**: `conductor_loop_execution_time_seconds`
- **안전하지 않은 페이로드 커밋**: `conductor_unsafe_payload_commit_total`

## 실제 사용 예시

### 1. 3노드 클러스터 구성

#### 노드 1 (Bootstrap)
```bash
# 첫 번째 노드 (Bootstrap)
op-conductor \
  --raft.server-id=conductor-1 \
  --raft.bootstrap=true \
  --raft.listen-addr=0.0.0.0 \
  --raft.listen-port=8081 \
  --rpc.listen-addr=0.0.0.0 \
  --rpc.listen-port=8080
```

#### 노드 2, 3 (Follower)
```bash
# 두 번째 노드
op-conductor \
  --raft.server-id=conductor-2 \
  --raft.bootstrap=false \
  --raft.listen-addr=0.0.0.0 \
  --raft.listen-port=8082 \
  --rpc.listen-addr=0.0.0.0 \
  --rpc.listen-port=8083

# 세 번째 노드
op-conductor \
  --raft.server-id=conductor-3 \
  --raft.bootstrap=false \
  --raft.listen-addr=0.0.0.0 \
  --raft.listen-port=8084 \
  --rpc.listen-addr=0.0.0.0 \
  --rpc.listen-port=8085
```

### 2. 클러스터 관리

#### 노드 추가
```bash
# 투표권 있는 노드 추가
curl -X POST http://conductor-1:8080/conductor/add-voter \
  -H "Content-Type: application/json" \
  -d '{"id": "conductor-4", "addr": "conductor-4:8086", "version": 1}'

# 투표권 없는 노드 추가
curl -X POST http://conductor-1:8080/conductor/add-nonvoter \
  -H "Content-Type: application/json" \
  -d '{"id": "conductor-5", "addr": "conductor-5:8087", "version": 1}'
```

#### 리더십 전환
```bash
# 자동 리더십 전환
curl -X POST http://conductor-1:8080/conductor/transfer-leader

# 특정 노드로 리더십 전환
curl -X POST http://conductor-1:8080/conductor/transfer-leader-to \
  -H "Content-Type: application/json" \
  -d '{"id": "conductor-2", "addr": "conductor-2:8083"}'
```

### 3. 상태 확인

#### 클러스터 상태 조회
```bash
# 현재 리더 확인
curl http://conductor-1:8080/conductor/leader

# 클러스터 멤버십 확인
curl http://conductor-1:8080/conductor/cluster-membership

# 시퀀서 상태 확인
curl http://conductor-1:8080/conductor/sequencer-healthy
```

### 4. 장애 시나리오

#### 리더 장애 시나리오
```go
// 1. 리더 노드 장애 발생
// 2. Health Monitor가 장애 감지
// 3. Raft 합의로 새로운 리더 선출
// 4. 새 리더가 시퀀서 활성화
// 5. 이전 리더 복구 시 팔로워로 전환

// 실제 로그 예시
2024-01-01T10:00:00Z INFO Leadership status changed server=conductor-1 leader=false
2024-01-01T10:00:01Z INFO Leadership status changed server=conductor-2 leader=true
2024-01-01T10:00:02Z INFO state changed prev_state="leader: false, healthy: true, active: false" new_state="leader: true, healthy: true, active: true"
```

## 결론

Optimism Conductor는 **완전한 다중 시퀀서 지원 시스템**으로, 다음과 같은 특징을 가집니다:

### 강점
1. **안정성**: Raft 합의 기반으로 안정적인 리더 선출
2. **자동화**: 장애 감지 및 자동 복구
3. **확장성**: 동적 노드 추가/제거 지원
4. **모니터링**: 상세한 메트릭스 및 로깅
5. **운영성**: 쉬운 설정 및 관리

### 활용 방안
1. **기존 시스템 활용**: 현재 Conductor를 그대로 활용
2. **추가 기능 구현**: 백업 완료 거래 풀 등 추가 기능 구현
3. **운영 최적화**: 모니터링 및 알림 시스템 구축

이러한 강력한 기반 위에 **백업 완료 거래 풀 시스템**을 추가하면, 완벽한 다중 시퀀서 환경을 구축할 수 있습니다.

## 현재 시스템의 백업 시퀀서 역할

### 1. **기본 역할: 완전 비활성화**

현재 Optimism에서 **팔로워 시퀀서는 완전히 비활성화된 상태**로 운영됩니다:

```go
// op-conductor/conductor/service.go
case !status.leader && status.healthy && !status.active:
    // normal follower, do nothing
case !status.leader && status.healthy && status.active:
    // stop sequencer, this happens when current server steps down as leader.
    err = oc.stopSequencer()
```

### 2. **상태별 역할**

#### **A. 정상 팔로워 상태**
```go
// [팔로워, 정상, 시퀀서 비활성화] → 아무것도 하지 않음
case !status.leader && status.healthy && !status.active:
    // normal follower, do nothing
```

**역할:**
- **블록 생성 안함**: 시퀀서가 완전히 중지됨
- **동기화만 수행**: L1/L2 블록 동기화만 수행
- **대기 상태**: 리더가 될 때까지 대기

#### **B. 시퀀서 중지 강제**
```go
// [팔로워, 정상, 시퀀서 활성화] → 시퀀서 중지
case !status.leader && status.healthy && status.active:
    // stop sequencer, this happens when current server steps down as leader.
    err = oc.stopSequencer()
```

**역할:**
- **강제 중지**: 리더에서 팔로워로 전환 시 시퀀서 강제 중지
- **상태 정리**: 활성화된 시퀀서 상태를 정리

### 3. **시퀀서 중지 메커니즘**

#### **중지 과정**
```go
func (oc *OpConductor) stopSequencer() error {
    oc.log.Info("stopping sequencer",
        "server", oc.cons.ServerID(),
        "leader", oc.leader.Load(),
        "healthy", oc.healthy.Load(),
        "active", oc.seqActive.Load())

    // 시퀀서 중지 호출
    latestHead, err := oc.ctrl.StopSequencer(oc.shutdownCtx)
    if err == nil {
        oc.log.Info("stopped sequencer", "latestHead", latestHead)
    }

    oc.seqActive.Store(false)
    return nil
}
```

#### **중지의 안전성**
```go
// 주석에서 설명하는 안전성 보장
// StopSequencer is called after conductor loses leadership. In the event that
// the StopSequencer call fails, it actually has little real consequences because the sequencer
// cant produce a block and gossip / commit it to the raft log (requires leadership).
```

### 4. **리더 전환 시나리오**

#### **정상 전환**
```go
// 1. 리더가 팔로워로 전환
// 2. 시퀀서 자동 중지
// 3. 새로운 리더 선출
// 4. 새 리더가 시퀀서 시작
```

#### **장애 전환**
```go
// 1. 리더 장애 감지
// 2. Raft 합의로 새 리더 선출
// 3. 새 리더가 시퀀서 시작
// 4. 이전 리더 복구 시 팔로워로 전환
```

### 5. **현재 시스템의 한계**

#### **A. 거래 풀 동기화 없음**
- **독립적 거래 풀**: 각 시퀀서가 독립적인 거래 풀 보유
- **동기화 부재**: 팔로워가 거래 풀을 동기화하지 않음
- **순서 불일치**: 리더 전환 시 거래 순서가 달라질 수 있음

#### **B. 블록 생성 중단**
- **완전 중지**: 팔로워는 블록 생성을 완전히 중지
- **전환 지연**: 리더 전환 시 블록 생성이 일시적으로 중단
- **동기화 시간**: 새 리더가 동기화하는 동안 블록 생성 지연

### 6. **개선 방향**

#### **현재 Optimism vs 개선된 시스템**

**현재 Optimism:**
```
리더: [시퀀서 활성화] → 블록 생성
팔로워: [시퀀서 비활성화] → 블록 생성 안함
```

**개선된 시스템 (백업 완료 거래 풀):**
```
리더: [시퀀서 활성화] → 백업 완료 거래로 블록 생성
팔로워: [시퀀서 활성화] → 백업 완료 거래로 블록 생성
```

#### **개선 효과**
1. **거래 순서 보장**: 메인과 백업이 동일한 거래 순서로 블록 생성
2. **연속성 보장**: 리더 전환 시에도 블록 생성 중단 없음
3. **일관성 향상**: 모든 시퀀서가 동일한 상태로 동작
4. **안정성 증대**: 장애 상황에서도 일관된 서비스 제공

### 7. **실제 운영 시나리오**

#### **현재 시스템의 리더 전환**
```go
// 1. 리더 장애 발생
// 2. 새 리더 선출 (5-10초 소요)
// 3. 새 리더 시퀀서 시작 (추가 5-10초 소요)
// 4. 블록 생성 재개
// 총 중단 시간: 10-20초
```

#### **개선된 시스템의 리더 전환**
```go
// 1. 리더 장애 발생
// 2. 새 리더 선출 (5-10초 소요)
// 3. 백업 완료 거래로 즉시 블록 생성 재개
// 총 중단 시간: 5-10초 (50% 단축)
```

## 결론

현재 Optimism Conductor는 **완전한 다중 시퀀서 지원 시스템**을 제공하지만, **백업 시퀀서의 역할이 제한적**입니다.

### 현재 시스템의 특징
1. **안정성**: Raft 합의 기반으로 안정적인 리더 선출
2. **자동화**: 장애 감지 및 자동 복구
3. **한계**: 팔로워 시퀀서의 완전 비활성화로 인한 블록 생성 중단

### 개선 방향
1. **백업 완료 거래 풀 시스템** 추가
2. **모든 시퀀서 활성화**로 연속성 보장
3. **거래 순서 일관성** 확보
4. **장애 대응 시간 단축**

이러한 개선을 통해 **더욱 안정적이고 일관된 다중 시퀀서 환경**을 구축할 수 있습니다.
