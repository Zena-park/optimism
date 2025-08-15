# op-challenger 기존 기능 상세 분석

## 개요

op-challenger는 Optimism의 Fault Proof 시스템에서 핵심적인 역할을 하는 컴포넌트입니다. 이 문서는 op-challenger의 기존 기능과 모니터링 시스템을 상세히 분석합니다.

## 1. 아키텍처 개요

### 1.1 주요 컴포넌트

```
op-challenger/
├── game/                    # 게임 관련 핵심 로직
│   ├── service.go          # 메인 서비스
│   ├── monitor.go          # 모니터링 시스템
│   ├── scheduler/          # 게임 스케줄링
│   ├── fault/              # Fault Proof 게임
│   └── registry/           # 게임 타입 레지스트리
├── config/                 # 설정 관리
├── metrics/                # 메트릭스 수집
├── sender/                 # 트랜잭션 전송
└── p2p/                    # P2P 네트워킹 (새로 추가)
```

### 1.2 서비스 구조

```go
type Service struct {
    logger          log.Logger
    metrics         metrics.Metricer
    monitor         *gameMonitor
    sched           *scheduler.Scheduler
    faultGamesCloser fault.CloseFunc
    preimages       *keccak.LargePreimageScheduler
    txMgr           *txmgr.SimpleTxManager
    txSender        *sender.TxSender
    systemClock     clock.Clock
    l1Clock         *clock.SimpleClock
    claimants       []common.Address
    claimer         *claims.BondClaimScheduler
    factoryContract *contracts.DisputeGameFactoryContract
    registry        *registry.GameTypeRegistry
    oracles         *registry.OracleRegistry
    l1Client        *ethclient.Client
    pollClient      client.RPC
    attentionTrigger *attention.AttentionTrigger  // RAT 시스템
}
```

## 2. 모니터링 시스템 상세 분석

### 2.1 L1 블록 헤드 모니터링

#### 2.1.1 모니터링 메커니즘

```go
// gameMonitor 구조체
type gameMonitor struct {
    logger       log.Logger
    clock        RWClock
    source       gameSource
    scheduler    gameScheduler
    preimages    preimageScheduler
    gameWindow   time.Duration
    claimer      claimer
    allowedGames []common.Address
    l1HeadsSub   ethereum.Subscription
    l1Source     *headSource
    runState     sync.Mutex
}
```

#### 2.1.2 L1 헤드 구독 프로세스

1. **구독 시작**: `StartMonitoring()` 메서드에서 L1 헤드 구독 시작
2. **이벤트 리스닝**: `eth_subscribe`를 통해 새로운 L1 블록 헤드 이벤트 수신
3. **자동 재구독**: 연결이 끊어지면 10초 후 자동으로 재구독
4. **콜백 처리**: 새로운 헤드가 도착하면 `onNewL1Head` 콜백 실행

```go
func (m *gameMonitor) onNewL1Head(ctx context.Context, sig eth.L1BlockRef) {
    m.clock.SetTime(sig.Time)
    if err := m.progressGames(ctx, sig.Hash, sig.Number); err != nil {
        m.logger.Error("Failed to progress games", "err", err)
    }
    if err := m.preimages.Schedule(sig.Hash, sig.Number); err != nil {
        m.logger.Error("Failed to validate large preimages", "err", err)
    }
}
```

### 2.2 Dispute Game 모니터링

#### 2.2.1 게임 소스 인터페이스

```go
type gameSource interface {
    GetGamesAtOrAfter(ctx context.Context, blockHash common.Hash, earliestTimestamp uint64) ([]types.GameMetadata, error)
}
```

#### 2.2.2 게임 진행 프로세스

1. **게임 조회**: `GetGamesAtOrAfter`로 게임 윈도우 내의 모든 게임 조회
2. **필터링**: 허용된 게임만 선택 (GameAllowlist 기반)
3. **스케줄링**: 게임 스케줄러에 게임 등록
4. **보증금 청구**: 게임 종료 후 보증금 자동 청구

```go
func (m *gameMonitor) progressGames(ctx context.Context, blockHash common.Hash, blockNumber uint64) error {
    minGameTimestamp := clock.MinCheckedTimestamp(m.clock, m.gameWindow)
    games, err := m.source.GetGamesAtOrAfter(ctx, blockHash, minGameTimestamp)
    if err != nil {
        return fmt.Errorf("failed to load games: %w", err)
    }

    var gamesToPlay []types.GameMetadata
    for _, game := range games {
        if !m.allowedGame(game.Proxy) {
            m.logger.Debug("Skipping game not on allow list", "game", game.Proxy)
            continue
        }
        gamesToPlay = append(gamesToPlay, game)
    }

    // 보증금 청구 스케줄링
    if err := m.claimer.Schedule(blockNumber, gamesToPlay); err != nil {
        return fmt.Errorf("failed to schedule bond claims: %w", err)
    }

    // 게임 스케줄링
    if err := m.scheduler.Schedule(gamesToPlay, blockNumber); errors.Is(err, scheduler.ErrBusy) {
        m.logger.Info("Scheduler still busy with previous update")
    } else if err != nil {
        return fmt.Errorf("failed to schedule games: %w", err)
    }
    return nil
}
```

### 2.3 Large Preimage 모니터링

#### 2.3.1 Preimage 시스템 구조

```go
type preimageScheduler interface {
    Schedule(blockHash common.Hash, blockNumber uint64) error
}
```

#### 2.3.2 Keccak Preimage 처리

1. **Preimage Fetcher**: L1에서 large preimage 데이터 가져오기
2. **Preimage Verifier**: preimage 데이터 검증
3. **Preimage Challenger**: 잘못된 preimage에 대한 챌린지 수행
4. **Large Preimage Scheduler**: preimage 처리 스케줄링

```go
func (s *Service) initLargePreimages() error {
    fetcher := fetcher.NewPreimageFetcher(s.logger, s.l1Client)
    verifier := keccak.NewPreimageVerifier(s.logger, fetcher)
    challenger := keccak.NewPreimageChallenger(s.logger, s.metrics, verifier, s.txSender)
    s.preimages = keccak.NewLargePreimageScheduler(s.logger, s.metrics, s.l1Clock, s.oracles, challenger)
    return nil
}
```

### 2.4 Bond Claim 모니터링

#### 2.4.1 보증금 청구 시스템

```go
type claimer interface {
    Schedule(blockNumber uint64, games []types.GameMetadata) error
}
```

#### 2.4.2 보증금 청구 프로세스

1. **게임 종료 감지**: 게임이 종료되면 자동으로 감지
2. **승리 확인**: 챌린저가 승리한 게임 식별
3. **보증금 청구**: 승리한 게임의 보증금을 자동으로 청구
4. **다중 청구자 지원**: 여러 주소에서 보증금 청구 가능

```go
func (s *Service) initBondClaims() error {
    claimer := claims.NewBondClaimer(s.logger, s.metrics, s.registry.CreateBondContract, s.txSender, s.claimants...)
    s.claimer = claims.NewBondClaimScheduler(s.logger, s.metrics, claimer)
    return nil
}
```

### 2.5 메트릭스 모니터링

#### 2.5.1 수집되는 메트릭스

1. **게임 관련 메트릭스**:
   - 활성 게임 수
   - 게임 진행 상태
   - 게임 종료 시간

2. **트랜잭션 메트릭스**:
   - 대기 중인 트랜잭션 수
   - 트랜잭션 성공/실패율
   - 가스 사용량

3. **성능 메트릭스**:
   - CPU 사용률
   - 메모리 사용량
   - 응답 시간

4. **잔액 메트릭스**:
   - 챌린저 계정 잔액
   - 보증금 잔액

#### 2.5.2 메트릭스 서버

```go
func (s *Service) initMetricsServer(cfg *opmetrics.CLIConfig) error {
    if !cfg.Enabled {
        return nil
    }

    m, ok := s.metrics.(opmetrics.RegistryMetricer)
    if !ok {
        return fmt.Errorf("metrics were enabled, but metricer %T does not expose registry for metrics-server", s.metrics)
    }

    metricsSrv, err := opmetrics.StartServer(m.Registry(), cfg.ListenAddr, cfg.ListenPort)
    if err != nil {
        return fmt.Errorf("failed to start metrics server: %w", err)
    }

    s.metricsSrv = metricsSrv
    s.balanceMetricer = s.metrics.StartBalanceMetrics(s.logger, s.l1Client, s.txSender.From())
    return nil
}
```

## 3. 게임 스케줄링 시스템

### 3.1 스케줄러 구조

```go
type Scheduler struct {
    logger    log.Logger
    metrics   metrics.Metricer
    disk      *diskManager
    maxConcurrency uint
    createPlayer   game.PlayerCreator
    allowInvalidPrestate bool
}
```

### 3.2 게임 진행 프로세스

1. **게임 등록**: 새로운 게임을 스케줄러에 등록
2. **동시성 제어**: `MaxConcurrency` 설정에 따라 동시 처리 게임 수 제한
3. **플레이어 생성**: 각 게임에 대한 플레이어 인스턴스 생성
4. **게임 실행**: 게임 로직에 따라 자동으로 게임 진행

## 4. 설정 시스템

### 4.1 주요 설정 항목

```go
type Config struct {
    L1EthRpc             string           // L1 RPC URL
    L1Beacon             string           // L1 Beacon API URL
    GameFactoryAddress   common.Address   // Dispute Game Factory 주소
    GameAllowlist        []common.Address // 허용된 게임 주소 목록
    GameWindow           time.Duration    // 게임 모니터링 윈도우 (기본 28일)
    Datadir              string           // 데이터 디렉토리
    MaxConcurrency       uint             // 최대 동시 처리 수
    PollInterval         time.Duration    // 폴링 간격 (기본 12초)
    AllowInvalidPrestate bool             // 잘못된 prestate 허용 여부
    AdditionalBondClaimants []common.Address // 추가 보증금 청구자
    TraceTypes           []types.TraceType // 지원하는 트레이스 타입
    MaxPendingTx         uint64           // 최대 대기 트랜잭션 수
}
```

### 4.2 설정 검증

```go
func (c *Config) Check() error {
    if c.L1EthRpc == "" {
        return ErrMissingL1EthRPC
    }
    if c.GameFactoryAddress == (common.Address{}) {
        return ErrMissingGameFactoryAddress
    }
    if c.Datadir == "" {
        return ErrMissingDatadir
    }
    if c.MaxConcurrency == 0 {
        return ErrMaxConcurrencyZero
    }
    if len(c.TraceTypes) == 0 {
        return ErrMissingTraceType
    }
    return nil
}
```

## 5. 트랜잭션 관리

### 5.1 트랜잭션 매니저

```go
type SimpleTxManager struct {
    name     string
    cfg      CLIConfig
    backend  bind.ContractBackend
    l        log.Logger
    metr     metrics.TxMetricer
    chainID  *big.Int
    wallet   *wallet.Wallet
    nonce    *nonce.Manager
    pending  *pending.TxTracker
    resub    *resubmit.Resubmitter
    fees     *fees.Estimator
}
```

### 5.2 트랜잭션 전송 프로세스

1. **가스 추정**: 현재 네트워크 상황에 맞는 가스 가격 추정
2. **논스 관리**: 트랜잭션 논스 자동 관리
3. **재전송**: 실패한 트랜잭션 자동 재전송
4. **대기 트랜잭션 관리**: 최대 대기 트랜잭션 수 제한

## 6. 서비스 생명주기

### 6.1 서비스 시작

```go
func (s *Service) Start(ctx context.Context) error {
    s.logger.Info("starting scheduler")
    s.sched.Start(ctx)
    s.claimer.Start(ctx)
    s.preimages.Start(ctx)

    s.logger.Info("starting monitoring")
    s.monitor.StartMonitoring()

    // RAT 모니터링 시작
    if s.attentionTrigger != nil {
        s.logger.Info("starting RAT monitoring")
        if err := s.attentionTrigger.StartRATMonitoring(ctx, s.config.L1EthRpc); err != nil {
            s.logger.Error("failed to start RAT monitoring", "error", err)
            return fmt.Errorf("failed to start RAT monitoring: %w", err)
        }
        s.logger.Info("RAT monitoring started successfully")
    }

    s.logger.Info("challenger game service start completed")
    return nil
}
```

### 6.2 서비스 종료

```go
func (s *Service) Stop(ctx context.Context) error {
    s.logger.Info("stopping challenger game service")

    if s.faultGamesCloser != nil {
        s.faultGamesCloser()
    }

    if s.monitor != nil {
        s.monitor.StopMonitoring()
    }

    if s.sched != nil {
        s.sched.Close()
    }

    if s.claimer != nil {
        s.claimer.Stop()
    }

    if s.preimages != nil {
        s.preimages.Stop()
    }

    if s.balanceMetricer != nil {
        s.balanceMetricer.Close()
    }

    if s.metricsSrv != nil {
        s.metricsSrv.Stop(ctx)
    }

    if s.pprofService != nil {
        s.pprofService.Stop()
    }

    s.stopped.Store(true)
    return nil
}
```

## 7. 모니터링 데이터 흐름

### 7.1 전체 데이터 흐름

```
L1 블록 헤드 → gameMonitor.onNewL1Head()
    ↓
progressGames() → GetGamesAtOrAfter()
    ↓
게임 필터링 → allowedGame()
    ↓
스케줄링 → scheduler.Schedule()
    ↓
게임 실행 → Player.ProgressGame()
    ↓
결과 처리 → Bond Claim / Metrics Update
```

### 7.2 모니터링 주기

1. **L1 헤드 구독**: 실시간 이벤트 기반
2. **게임 진행**: 새로운 L1 블록마다 실행
3. **Preimage 검증**: 새로운 L1 블록마다 실행
4. **보증금 청구**: 게임 종료 시 자동 실행
5. **메트릭스 수집**: 지속적으로 수집

## 8. 성능 특성

### 8.1 동시성 제어

- **MaxConcurrency**: 동시에 처리할 수 있는 게임 수 제한
- **게임 윈도우**: 28일 이내의 게임만 모니터링
- **폴링 간격**: 12초마다 L1 상태 확인

### 8.2 리소스 사용량

- **메모리**: 게임 상태와 트랜잭션 데이터 저장
- **CPU**: 게임 진행 로직 실행
- **네트워크**: L1 RPC 호출 및 트랜잭션 전송

### 8.3 장애 복구

- **자동 재구독**: L1 구독 연결 끊김 시 자동 복구
- **트랜잭션 재전송**: 실패한 트랜잭션 자동 재시도
- **게임 재시작**: 게임 진행 실패 시 재시도

## 9. 보안 고려사항

### 9.1 게임 필터링

- **GameAllowlist**: 허용된 게임만 처리
- **Prestate 검증**: 잘못된 prestate 거부 (설정 가능)

### 9.2 트랜잭션 보안

- **가스 추정**: 동적 가스 가격 추정
- **논스 관리**: 트랜잭션 순서 보장
- **재전송 제한**: 무한 재전송 방지

## 10. 결론

op-challenger는 Optimism 네트워크의 보안을 담당하는 핵심 컴포넌트로, 다음과 같은 특징을 가집니다:

1. **실시간 모니터링**: L1 블록 헤드 기반 실시간 감시
2. **자동화된 게임 진행**: Dispute Game 자동 처리
3. **다중 보안 계층**: Preimage 검증, Bond Claim 등
4. **확장 가능한 아키텍처**: 새로운 게임 타입 지원
5. **견고한 장애 복구**: 자동 재연결 및 재시도
6. **포괄적인 모니터링**: 메트릭스 및 로깅

이러한 기능들을 통해 Optimism 네트워크의 무결성과 보안을 지속적으로 보장합니다.

## 11. Q&A

### Q1: L1 헤드 구독하는 이유는 무엇인가요?

#### A1: 실시간 Fault Proof 게임 감지

L1 헤드 구독의 가장 중요한 이유는 **새로운 Dispute Game이 생성되는 것을 실시간으로 감지**하기 위함입니다.

```go
func (m *gameMonitor) onNewL1Head(ctx context.Context, sig eth.L1BlockRef) {
    m.clock.SetTime(sig.Time)
    if err := m.progressGames(ctx, sig.Hash, sig.Number); err != nil {
        m.logger.Error("Failed to progress games", "err", err)
    }
    if err := m.preimages.Schedule(sig.Hash, sig.Number); err != nil {
        m.logger.Error("Failed to validate large preimages", "err", err)
    }
}
```

**왜 실시간이 중요한가?**
- Optimism에서 잘못된 상태 전환이 발생하면 즉시 챌린지해야 함
- 게임이 생성된 후 일정 시간 내에 응답하지 않으면 자동으로 패배
- **시간이 곧 보안**이기 때문에 지연은 치명적

#### A2: 블록 해시 기반 게임 조회

```go
func (f *DisputeGameFactoryContract) GetGamesAtOrAfter(ctx context.Context, blockHash common.Hash, earliestTimestamp uint64) ([]types.GameMetadata, error) {
    count, err := f.GetGameCount(ctx, blockHash)
    // ... 게임 조회 로직
}
```

**블록 해시가 필요한 이유:**
- 특정 L1 블록 시점의 게임 상태를 정확히 조회
- 포크나 리오그가 발생해도 일관된 상태 보장
- 게임 진행 시 정확한 블록 컨텍스트 필요

#### A3: 시간 기반 게임 윈도우 관리

```go
func (m *gameMonitor) progressGames(ctx context.Context, blockHash common.Hash, blockNumber uint64) error {
    minGameTimestamp := clock.MinCheckedTimestamp(m.clock, m.gameWindow)
    games, err := m.source.GetGamesAtOrAfter(ctx, blockHash, minGameTimestamp)
    // ...
}
```

**게임 윈도우의 의미:**
- **기본 28일**: 게임의 최대 지속 시간 (16일 게임 + 7일 WETH 출금 지연 + 5일 버퍼)
- 오래된 게임은 자동으로 제외하여 리소스 절약
- 새로운 L1 블록마다 윈도우 업데이트

#### A4: Preimage 검증 타이밍

```go
if err := m.preimages.Schedule(sig.Hash, sig.Number); err != nil {
    m.logger.Error("Failed to validate large preimages", "err", err)
}
```

**Preimage 검증이 L1 헤드와 연동되는 이유:**
- Large preimage는 L1에 제출되는 데이터
- 새로운 L1 블록마다 새로운 preimage 제출 가능
- 즉시 검증하여 잘못된 preimage 챌린지

#### A5: 보증금 청구 타이밍

```go
if err := m.claimer.Schedule(blockNumber, gamesToPlay); err != nil {
    return fmt.Errorf("failed to schedule bond claims: %w", err)
}
```

**보증금 청구가 L1 헤드와 연동되는 이유:**
- 게임이 종료되면 즉시 보증금 청구 가능
- L1 블록마다 게임 상태 확인
- 승리한 게임의 보증금을 빠르게 회수

### Q2: 폴링 vs 구독, 왜 구독을 선택했나요?

#### A2: 폴링 방식의 문제점

```go
// 폴링 방식 (비효율적)
for {
    latestBlock := getLatestBlock()
    if latestBlock > lastProcessedBlock {
        processNewBlock(latestBlock)
        lastProcessedBlock = latestBlock
    }
    time.Sleep(12 * time.Second)  // 지연 발생
}
```

#### A2: 구독 방식의 장점

```go
// 구독 방식 (실시간)
m.l1HeadsSub = event.ResubscribeErr(time.Second*10, m.resubscribeFunction())
```

**구독의 장점:**
1. **즉시성**: 새로운 블록이 생성되면 즉시 알림
2. **효율성**: 불필요한 폴링 없음
3. **자동 복구**: 연결 끊김 시 자동 재구독
4. **리소스 절약**: CPU 및 네트워크 사용량 최적화

### Q3: 실제 시나리오에서 어떻게 동작하나요?

#### A3: 시나리오 1 - 잘못된 상태 전환 감지

```
1. L2에서 잘못된 상태 전환 발생
2. L1에 Dispute Game 생성 트랜잭션 제출
3. L1 블록에 포함됨
4. op-challenger가 즉시 새로운 L1 헤드 감지
5. 새로운 게임 조회 및 스케줄링
6. 즉시 챌린지 시작
```

#### A3: 시나리오 2 - 게임 진행

```
1. 새로운 L1 블록 생성
2. op-challenger가 L1 헤드 업데이트 감지
3. 모든 활성 게임의 상태 확인
4. 응답이 필요한 게임에 대해 즉시 응답
5. 게임 종료 시 보증금 청구
```

### Q4: L1 헤드 구독이 Optimism 보안에 미치는 영향은?

#### A4: 핵심 보안 메커니즘

L1 헤드 구독은 **Optimism의 보안을 실시간으로 보장**하기 위한 핵심 메커니즘입니다:

1. **보안**: 잘못된 상태 전환을 즉시 감지하고 챌린지
2. **효율성**: 불필요한 폴링 없이 실시간 이벤트 기반 처리
3. **신뢰성**: 자동 재구독으로 연결 안정성 보장
4. **정확성**: 블록 해시 기반으로 일관된 상태 조회

이러한 이유로 op-challenger는 폴링이 아닌 **구독 방식**을 선택하여 Optimism 네트워크의 무결성을 실시간으로 보호합니다.

### Q5: L1에 Dispute Game 생성 트랜잭션을 제출하는 주체는 누구인가요?

#### A5: Proposer (제안자)

L1에 Dispute Game 생성 트랜잭션을 제출하는 주체는 **op-proposer**입니다.

> **📖 관련 문서**: [Optimism 보안 모델: 배치 제출 vs Dispute Game](../optimism-security-model.md)
>
> 이 문서에서는 왜 Sequencer의 배치 제출과 Proposer의 Dispute Game 제출이 모두 필요한지, 그리고 각각의 보안상 중요성을 자세히 설명합니다.

```go
// op-proposer/proposer/driver.go
func (l *L2OutputSubmitter) sendTransaction(ctx context.Context, output source.Proposal) error {
    l.Log.Info("Proposing output root", "output", output.Root, "block", output.SequenceNum)
    var receipt *types.Receipt
    if l.Cfg.DisputeGameFactoryAddr != nil {
        candidate, err := l.ProposeL2OutputDGFTxCandidate(ctx, output)
        if err != nil {
            return err
        }
        receipt, err = l.Txmgr.Send(ctx, candidate)
        if err != nil {
            return err
        }
    }
    // ...
}
```

#### A5: Proposer의 역할

1. **L2 상태 모니터링**: L2 체인의 새로운 블록을 지속적으로 모니터링
2. **Output Root 생성**: L2 블록의 상태 루트를 계산하여 output root 생성
3. **Dispute Game 생성**: 새로운 output root에 대해 Dispute Game 생성 트랜잭션 제출
4. **정기적 제출**: 설정된 간격(ProposalInterval)에 따라 자동으로 제출

#### A5: Dispute Game 생성 조건

```go
func (l *L2OutputSubmitter) FetchDGFOutput(ctx context.Context) (source.Proposal, bool, error) {
    cutoff := time.Now().Add(-l.Cfg.ProposalInterval)
    proposedRecently, proposalTime, claim, err := l.dgfContract.HasProposedSince(ctx, l.Txmgr.From(), cutoff, l.Cfg.DisputeGameType)

    if proposedRecently {
        l.Log.Debug("Duration since last game not past proposal interval", "duration", time.Since(proposalTime))
        return source.Proposal{}, false, nil
    }

    // 새로운 output root가 이전과 다른 경우에만 제출
    if claim == output.Root {
        l.Log.Debug("Skipping proposal: output root unchanged since last proposed game")
        return source.Proposal{}, false, nil
    }

    return output, true, nil
}
```

#### A5: 생성 프로세스

1. **L2 블록 확인**: 현재 L2 블록 번호 조회
2. **Output Root 계산**: 해당 블록의 상태 루트 기반 output root 생성
3. **중복 확인**: 이전에 제출된 output root와 동일한지 확인
4. **시간 간격 확인**: 마지막 제출 이후 충분한 시간이 지났는지 확인
5. **트랜잭션 제출**: DisputeGameFactory의 `create` 함수 호출

#### A5: 보안 메커니즘

- **보증금 요구**: Dispute Game 생성 시 필수 보증금 지불
- **중복 방지**: 동일한 output root에 대한 중복 게임 생성 방지
- **시간 제한**: 제안 간격을 통한 과도한 게임 생성 방지
- **권한 검증**: 특정 주소만 게임 생성 가능 (PermissionedDisputeGame의 경우)

#### A5: op-challenger와의 관계

```
Proposer (op-proposer) → L1에 Dispute Game 생성 → op-challenger가 L1 헤드 구독으로 감지 → 즉시 챌린지 시작
```

이러한 구조를 통해:
- **Proposer**: L2 상태를 L1에 정기적으로 제출
- **Challenger**: 잘못된 제출을 즉시 감지하고 챌린지
- **상호 견제**: Proposer와 Challenger가 서로를 감시하여 시스템 무결성 보장

### Q6: 왜 배치 제출과 Dispute Game 제출이 모두 필요한가요?

#### A6: 보안 모델의 완성

Sequencer의 배치 제출과 Proposer의 Dispute Game 제출은 각각 다른 목적을 가지고 있으며, 둘 다 Optimism의 완전한 보안을 위해 필요합니다.

> **📖 상세 분석**: [Optimism 보안 모델: 배치 제출 vs Dispute Game](../optimism-security-model.md)
>
> 이 문서에서는 다음 내용을 자세히 다룹니다:
> - 배치 제출만으로는 부족한 이유
> - Dispute Game의 필요성과 역할
> - 비용 vs 보안의 트레이드오프
> - 실제 위험 시나리오 분석
> - 다른 L2 솔루션과의 비교

## 12. 챌린저 개선 제안

### 12.1 현재 검증 방식의 한계

#### 12.1.1 현재 검증 방식의 한계

현재 op-challenger는 **상태 루트 비교**를 통해 배치 오류를 감지합니다:

**현재 검증 과정:**
- Dispute Game의 상태 루트와 L2 실행 결과의 상태 루트를 비교
- **L1 배치정보를 직접 확인하지 않고**, 배치 데이터가 L2 상태에 미친 영향을 통해 검증
- 배치 오류가 있으면 → 잘못된 상태 루트 계산 → 상태 루트 불일치 감지

#### 12.1.2 현재 방식의 한계

현재 방식의 주요 한계는 **L1 배치정보를 직접 확인하지 않는다**는 점입니다:

**핵심 문제:**
- 배치 데이터 자체의 오류를 즉시 감지할 수 없음
- 상태 루트 불일치만으로는 정확한 오류 원인 파악이 어려움
- 디버깅 시 배치 데이터 → L2 상태 변화 → 상태 루트 계산 과정을 역추적해야 함

### 12.2 제안: 직접 배치 데이터 비교 기능

#### 12.2.1 새로운 검증 아키텍처

제안하는 개선된 검증 시스템은 **이중 검증 구조**를 가집니다:

**기존 검증 (상태 루트 기반):**
- Dispute Game의 상태 루트와 L2 실행 결과의 상태 루트를 비교
- 상태 무결성 검증을 담당
- 기존의 모든 기능을 그대로 유지

**새로운 직접 검증 (배치 데이터 기반 - 매우 간단함):**
- **L1 배치정보를 직접 확인**하여 L1 BatchInbox에 제출된 배치 데이터와 L2 체인의 실제 배치 데이터를 **직접 비교**
- **단순한 데이터 비교**만으로 충분
- 복잡한 시뮬레이션이나 계산이 전혀 필요 없음
- 배치 데이터 자체의 오류를 즉시 감지

**통합 검증 결과:**
- 두 검증 결과를 종합하여 최종 판단
- "VALID", "INVALID_STATE", "INVALID_BATCH", "INVALID_BOTH" 중 하나로 분류
- 정확한 오류 원인 파악 가능

**핵심 장점:**
- **구현 간단성**: 단순한 배치 데이터 직접 비교
- **효율성**: 복잡한 계산 없이 즉시 검증 가능
- **정확성**: 배치 데이터 수준에서 정확한 오류 감지

#### 12.2.2 새로운 컴포넌트 구조

**BatchVerifier 컴포넌트:**
- L1과 L2 클라이언트를 통해 배치 데이터에 접근
- BatchInbox 컨트랙트 주소를 통해 L1 배치 데이터 조회
- 메트릭스 수집을 통한 성능 모니터링

**BatchComparisonResult:**
- L1과 L2 배치 데이터의 해시값 비교
- 일치 여부와 불일치 시 상세 차이점 정보
- 검증 수행 시간 기록

#### 12.2.3 구현 방법

**L1 배치 데이터 추출:**
- L1 BatchInbox 컨트랙트에서 특정 블록 번호의 배치 데이터를 조회
- 컨트랙트의 getBatch 함수를 호출하여 배치 데이터 추출
- 배치 데이터의 해시값과 원본 데이터를 함께 저장

**L2 배치 데이터 추출:**
- L2 RPC를 통해 해당 블록 번호의 블록 정보 조회
- 블록에서 트랜잭션 데이터를 추출하여 배치 형태로 변환
- L1 배치와 동일한 형식으로 정규화하여 비교 가능하게 만듦

**직접 비교 수행:**
- L1과 L2 배치 데이터의 해시값을 직접 비교
- 일치하는 경우 즉시 검증 완료
- 불일치하는 경우 상세 분석을 통해 구체적인 차이점 파악
- 차이점의 심각도 수준을 분류하여 우선순위 결정

#### 12.2.4 통합 검증 프로세스

**EnhancedVerifier 구조:**
- 기존 상태 루트 검증기와 새로운 배치 검증기를 모두 포함
- 두 검증 결과를 종합하여 최종 판단
- 로깅과 메트릭스를 통한 모니터링

**검증 프로세스:**
1. **상태 루트 검증**: 기존 방식대로 Dispute Game과 L2 실행 결과의 상태 루트 비교
2. **배치 데이터 검증**: 새로운 방식으로 L1과 L2 배치 데이터 직접 비교
3. **통합 결과 결정**: 두 검증 결과를 종합하여 최종 판단

**결과 분류:**
- **VALID**: 상태 루트와 배치 데이터 모두 일치
- **INVALID_STATE**: 배치 데이터는 일치하지만 상태 루트 불일치 (계산 오류 가능성)
- **INVALID_BATCH**: 상태 루트는 일치하지만 배치 데이터 불일치 (데이터 전송 오류)
- **INVALID_BOTH**: 상태 루트와 배치 데이터 모두 불일치 (심각한 오류)



### 12.3 결론

이 제안은 op-challenger의 검증 능력을 크게 향상시킬 것으로 예상됩니다:

1. **정확성**: 직접 배치 데이터 비교로 더 정확한 오류 감지
2. **디버깅**: 구체적인 오류 원인 파악으로 빠른 문제 해결
3. **신뢰성**: 이중 검증으로 시스템 무결성 강화
4. **확장성**: 새로운 검증 방법의 기반 제공

이러한 개선을 통해 Optimism 네트워크의 보안과 안정성을 한 단계 더 높일 수 있을 것입니다.
