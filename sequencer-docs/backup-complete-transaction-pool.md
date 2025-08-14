# 백업 완료 거래 풀 기반 블록 생성 방식

## 목차
1. [개요](#개요)
2. [아키텍처](#아키텍처)
3. [구현 방식](#구현-방식)
4. [백업 완료 거래 풀](#백업-완료-거래-풀)
5. [시퀀서 수정](#시퀀서-수정)
6. [시퀀서 관리 규칙](#시퀀서-관리-규칙)
7. [백업 동기화 메커니즘](#백업-동기화-메커니즘)
8. [장점 및 특징](#장점-및-특징)
9. [구현 예시](#구현-예시)

## 개요

### 문제 정의
멀티 시퀀서 환경에서 거래 순서를 보장하는 것은 중요한 과제입니다. 기존 방식에서는 각 시퀀서가 독립적인 거래 풀을 가지고 있어, 거래 순서가 일관되지 않을 수 있습니다.

### 해결 방안
**백업 완료 거래 풀 기반 블록 생성 방식**은 메인 노드와 백업 노드 모두 시퀀서가 되지만, **백업이 완료된 거래 풀에서만 블록에 담도록 하는 방식**입니다.

### 핵심 개념
- 메인 노드와 백업 노드 모두 시퀀서 역할 수행
- 거래는 메인 노드에서 먼저 수신
- 백업 노드로 거래 전파 및 백업 완료 확인
- 백업 완료된 거래만 블록에 포함
- 메인과 백업 시퀀서가 동일한 거래 순서 보장

## 아키텍처

### 전체 구조
```
┌─────────────────┐    ┌─────────────────┐
│   메인 노드     │    │   백업 노드     │
│  (시퀀서)       │    │  (시퀀서)       │
│                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ 거래 풀     │ │    │ │ 거래 풀     │ │
│ │ (메인)      │ │    │ │ (백업)      │ │
│ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
         ┌─────────────────────────┐
         │   백업 완료 거래 풀      │
         │   (공유 풀)             │
         │                         │
         │ ┌─────────────────────┐ │
         │ │ 백업 완료된 거래들   │ │
         │ │ (FIFO 순서)         │ │
         │ └─────────────────────┘ │
         └─────────────────────────┘
                     │
         ┌─────────────────────────┐
         │   블록 생성             │
         │   (동일한 거래 순서)    │
         └─────────────────────────┘
```

### 컴포넌트 구성
1. **메인 노드**: 거래 수신 및 메인 거래 풀 관리
2. **백업 노드**: 거래 백업 및 백업 거래 풀 관리
3. **백업 추적기**: 거래 백업 상태 추적
4. **공유 풀**: 백업 완료된 거래들 관리
5. **시퀀서**: 백업 완료 거래로 블록 생성

## 구현 방식

### 1. 거래 풀 백업 상태 관리

백업 완료 거래 풀 시스템의 핵심은 **거래의 백업 상태를 정확히 추적하는 것**입니다. 이를 통해 메인 노드와 백업 노드 간의 거래 동기화 상태를 실시간으로 모니터링하고, 백업이 완료된 거래만 블록에 포함시킬 수 있습니다.

#### 백업 상태 구조

백업 상태는 각 거래의 메인 노드 수신부터 백업 노드 수신까지의 전체 과정을 추적합니다:

- **MainReceived**: 메인 노드가 거래를 처음 수신한 시간
- **BackupReceived**: 백업 노드가 거래를 수신한 시간
- **IsBackedUp**: 백업 완료 여부 (백업 노드 수신 확인)
- **BackupLatency**: 메인 수신부터 백업 수신까지의 지연 시간

이 정보를 통해 백업 지연 시간을 분석하고, 백업 완료 조건을 정확히 판단할 수 있습니다.
```go
type TransactionPoolBackup struct {
    // 메인 노드의 거래 풀
    mainPool *TransactionPool

    // 백업 노드의 거래 풀
    backupPool *TransactionPool

    // 백업 완료된 거래들만 포함하는 공유 풀
    sharedPool *TransactionPool

    // 백업 상태 추적
    backupStatus map[common.Hash]BackupStatus
}

type BackupStatus struct {
    MainReceived    time.Time     // 메인 노드 수신 시간
    BackupReceived  time.Time     // 백업 노드 수신 시간
    IsBackedUp      bool          // 백업 완료 여부
    BackupLatency   time.Duration // 백업 지연 시간
}
```

#### 백업 추적기 구현

백업 추적기는 메인 노드와 백업 노드의 거래 풀 상태를 실시간으로 동기화하는 핵심 컴포넌트입니다. 이 컴포넌트는 다음과 같은 주요 기능을 제공합니다:

1. **거래 백업 상태 추적**: 각 거래의 메인 수신부터 백업 완료까지의 전체 과정을 기록
2. **백업 완료 거래 관리**: 백업이 완료된 거래들을 별도로 관리하여 시퀀서가 쉽게 접근할 수 있도록 함
3. **백업 통계 제공**: 백업 성공률, 평균 지연 시간 등 성능 지표를 실시간으로 계산
4. **동시성 제어**: 여러 시퀀서가 동시에 접근할 때 데이터 일관성을 보장

백업 추적기는 메인 노드와 백업 노드 모두에서 동일한 인터페이스를 제공하여, 어떤 시퀀서가 활성화되어도 동일한 백업 완료 거래 목록을 얻을 수 있도록 합니다.
```go
type TransactionBackupTracker struct {
    mainPool   *TransactionPool
    backupPool *TransactionPool
    sharedPool *TransactionPool

    // 거래 해시별 백업 상태
    backupStatus map[common.Hash]*BackupStatus

    // 백업 완료된 거래들
    backedUpTxs map[common.Hash]*types.Transaction

    mu sync.RWMutex
    log log.Logger
}

// 메인 노드에 거래 추가
func (tbt *TransactionBackupTracker) AddTransactionToMain(tx *types.Transaction) {
    tbt.mu.Lock()
    defer tbt.mu.Unlock()

    txHash := tx.Hash()
    status := &BackupStatus{
        MainReceived: time.Now(),
        IsBackedUp:   false,
    }

    tbt.backupStatus[txHash] = status
    tbt.mainPool.AddTransaction(tx)

    tbt.log.Debug("Transaction added to main pool", "hash", txHash)
}

// 백업 노드에 거래 추가
func (tbt *TransactionBackupTracker) AddTransactionToBackup(tx *types.Transaction) {
    tbt.mu.Lock()
    defer tbt.mu.Unlock()

    txHash := tx.Hash()
    if status, exists := tbt.backupStatus[txHash]; exists {
        status.BackupReceived = time.Now()
        status.IsBackedUp = true
        status.BackupLatency = status.BackupReceived.Sub(status.MainReceived)

        // 백업 완료된 거래를 공유 풀에 추가
        tbt.backedUpTxs[txHash] = tx
        tbt.sharedPool.AddTransaction(tx)

        tbt.log.Info("Transaction backed up",
            "hash", txHash,
            "latency", status.BackupLatency)
    }

    tbt.backupPool.AddTransaction(tx)
}
```

### 2. 백업 완료 거래 선택

백업 완료 거래 선택은 시스템의 핵심 기능 중 하나입니다. 이 기능은 백업이 완료된 거래들 중에서 블록에 포함할 거래들을 선택하는 로직을 담당합니다.

#### 기존 거래 선택 룰 유지

백업 완료 거래 풀에서는 **기존 Optimism의 거래 선택 룰을 그대로 적용**합니다. 이는 시스템 호환성을 유지하면서도 멀티 시퀀서 환경에서 일관된 거래 순서를 보장하기 위함입니다.

**거래 선택 우선순위:**
1. **Gas Price 우선순위**: 높은 Gas Price를 가진 거래가 먼저 선택됩니다. 이는 사용자가 지불한 수수료에 따른 공정한 처리 순서를 보장합니다.
2. **Nonce 순서**: 같은 계정에서 발생한 거래들은 Nonce 순서대로 처리됩니다. 이는 계정의 거래 순서를 보장합니다.
3. **도착 시간 순**: Gas Price와 Nonce가 모두 동일한 경우, 메인 노드에 먼저 도착한 거래가 우선 처리됩니다.

이러한 선택 룰을 통해 기존 Optimism 사용자들이 기대하는 거래 처리 방식과 동일한 경험을 제공하면서도, 백업 완료된 거래만 처리하여 멀티 시퀀서 간의 일관성을 보장합니다.
```go
func (tbt *TransactionBackupTracker) GetBackedUpTransactions() []*types.Transaction {
    tbt.mu.RLock()
    defer tbt.mu.RUnlock()

    var backedUpTxs []*types.Transaction
    for _, tx := range tbt.backedUpTxs {
        backedUpTxs = append(backedUpTxs, tx)
    }

    // 기존 Optimism 거래 선택 룰 적용
    sort.Slice(backedUpTxs, func(i, j int) bool {
        // 1. Gas Price가 높은 거래가 우선
        if backedUpTxs[i].GasPrice().Cmp(backedUpTxs[j].GasPrice()) != 0 {
            return backedUpTxs[i].GasPrice().Cmp(backedUpTxs[j].GasPrice()) > 0
        }
        // 2. Gas Price가 같으면 Nonce 순서
        if backedUpTxs[i].Nonce() != backedUpTxs[j].Nonce() {
            return backedUpTxs[i].Nonce() < backedUpTxs[j].Nonce()
        }
        // 3. Nonce도 같으면 도착 시간 순
        statusI := tbt.backupStatus[backedUpTxs[i].Hash()]
        statusJ := tbt.backupStatus[backedUpTxs[j].Hash()]
        return statusI.MainReceived.Before(statusJ.MainReceived)
    })

    return backedUpTxs
}
```

## 백업 완료 거래 풀

백업 완료 거래 풀은 메인 노드와 백업 노드 간의 거래 동기화를 위한 핵심 컴포넌트입니다. 이 풀은 백업이 완료된 거래들만을 포함하며, 모든 시퀀서가 동일한 거래 목록에 접근할 수 있도록 합니다.

### 1. 공유 풀 인터페이스

공유 풀은 백업 완료된 거래들을 관리하는 중앙 집중식 저장소 역할을 합니다. 이 풀을 통해 메인 시퀀서와 백업 시퀀서가 동일한 거래 목록을 바탕으로 블록을 생성할 수 있습니다.

#### 공유 풀 구조

공유 풀은 다음과 같은 주요 기능을 제공합니다:

- **백업 완료 거래 저장**: 백업 노드에서 수신 확인된 거래들을 안전하게 저장
- **거래 목록 제공**: 시퀀서가 블록 생성 시 사용할 수 있는 거래 목록 제공
- **블록 포함 거래 제거**: 블록에 포함된 거래들을 풀에서 제거하여 중복 처리 방지
- **동시성 제어**: 여러 시퀀서가 동시에 접근할 때 데이터 일관성 보장

이 구조를 통해 메인과 백업 시퀀서가 동일한 거래 순서로 블록을 생성할 수 있습니다.
```go
type SharedTransactionPool struct {
    backedUpTxs map[common.Hash]*types.Transaction
    mu          sync.RWMutex
    log         log.Logger
}

// 백업 완료된 거래 추가
func (stp *SharedTransactionPool) AddBackedUpTransaction(tx *types.Transaction) {
    stp.mu.Lock()
    defer stp.mu.Unlock()

    stp.backedUpTxs[tx.Hash()] = tx
    stp.log.Debug("Added backed up transaction to shared pool", "hash", tx.Hash())
}

// 백업 완료된 거래들 반환
func (stp *SharedTransactionPool) GetBackedUpTransactions() []*types.Transaction {
    stp.mu.RLock()
    defer stp.mu.RUnlock()

    var txs []*types.Transaction
    for _, tx := range stp.backedUpTxs {
        txs = append(txs, tx)
    }

    return txs
}

// 블록에 포함된 거래들 제거
func (stp *SharedTransactionPool) RemoveIncludedTransactions(block *types.Block) {
    stp.mu.Lock()
    defer stp.mu.Unlock()

    for _, tx := range block.Transactions() {
        delete(stp.backedUpTxs, tx.Hash())
    }

    stp.log.Debug("Removed included transactions from shared pool",
        "block", block.Number(), "txs", len(block.Transactions()))
}
```

### 2. 백업 완료 조건

백업 완료 조건은 거래가 백업 노드에서 안전하게 수신되었는지를 판단하는 기준입니다. 이 조건을 만족해야만 해당 거래가 블록에 포함될 수 있습니다.

#### 완료 조건 정의

백업 완료 조건은 다음과 같은 요소들을 종합적으로 고려합니다:

- **백업 노드 수신 확인**: 백업 노드가 실제로 거래를 수신했는지 확인
- **백업 지연 시간 제한**: 메인 노드 수신부터 백업 노드 수신까지의 지연 시간이 허용 범위 내인지 확인
- **백업 노드 상태 확인**: 백업 노드가 정상적으로 운영되고 있는지 확인

이러한 조건들을 통해 백업의 신뢰성을 보장하고, 백업 노드에 문제가 있을 때 메인 노드가 안전하게 운영될 수 있도록 합니다.
```go
type BackupCompletionCriteria struct {
    // 백업 노드 수신 확인
    BackupReceived bool

    // 백업 지연 시간 제한
    MaxBackupLatency time.Duration

    // 백업 노드 상태 확인
    BackupNodeHealthy bool
}

// 백업 완료 여부 확인
func (bcc *BackupCompletionCriteria) IsBackupComplete(txHash common.Hash) bool {
    status := bcc.getBackupStatus(txHash)

    return status.BackupReceived &&
           status.BackupLatency <= bcc.MaxBackupLatency &&
           bcc.BackupNodeHealthy
}
```

## 시퀀서 수정

기존 Optimism 시퀀서를 백업 완료 거래 풀 시스템에 맞게 수정해야 합니다. 이 수정을 통해 시퀀서가 백업 완료된 거래만 블록에 포함하도록 합니다.

### 1. 메인 시퀀서 수정

메인 시퀀서는 기존의 거래 풀에서 거래를 선택하는 대신, 백업 완료 거래 풀에서 거래를 선택하도록 수정됩니다. 이를 통해 백업 노드와의 동기화 상태를 보장할 수 있습니다.

#### 메인 시퀀서 구조

메인 시퀀서는 다음과 같은 추가 컴포넌트를 포함합니다:

- **백업 추적기**: 거래 백업 상태를 추적하는 컴포넌트
- **백업 완료 거래 선택 로직**: 백업이 완료된 거래만 선택하는 로직
- **백업 상태 모니터링**: 백업 성능과 상태를 모니터링하는 기능

이러한 구조를 통해 메인 시퀀서는 백업 노드와의 동기화를 유지하면서도 안정적으로 블록을 생성할 수 있습니다.
```go
type MainSequencer struct {
    *Sequencer
    backupTracker *TransactionBackupTracker
    log           log.Logger
}

// 블록 빌딩 시작
func (ms *MainSequencer) startBuildingBlock() {
    // 기존 코드...

    if !attrs.NoTxPool {
        // 백업 완료된 거래만 선택
        attrs.Transactions = ms.selectBackedUpTransactions()
    }

    // 기존 코드...
}

// 백업 완료된 거래 선택
func (ms *MainSequencer) selectBackedUpTransactions() []eth.Data {
    backedUpTxs := ms.backupTracker.GetBackedUpTransactions()

    var txData []eth.Data
    for _, tx := range backedUpTxs {
        data, err := tx.MarshalBinary()
        if err != nil {
            ms.log.Error("Failed to marshal transaction", "err", err)
            continue
        }
        txData = append(txData, data)
    }

    ms.log.Info("Selected backed up transactions for block",
        "count", len(txData))

    return txData
}
```

### 2. 백업 시퀀서 수정

백업 시퀀서는 메인 시퀀서와 동일한 로직을 사용하여 블록을 생성합니다. 이를 통해 메인 시퀀서가 장애가 발생했을 때 백업 시퀀서가 즉시 대체할 수 있습니다.

#### 백업 시퀀서 구조

백업 시퀀서는 메인 시퀀서와 거의 동일한 구조를 가지지만, 다음과 같은 차이점이 있습니다:

- **백업 역할**: 메인 시퀀서의 백업 역할을 수행
- **동일한 거래 선택 로직**: 메인 시퀀서와 동일한 백업 완료 거래 선택 로직 사용
- **장애 대응**: 메인 시퀀서 장애 시 즉시 활성화될 수 있는 준비 상태 유지

이러한 구조를 통해 백업 시퀀서는 메인 시퀀서와 완전히 동일한 블록을 생성할 수 있으며, 장애 상황에서도 서비스 연속성을 보장할 수 있습니다.
```go
type BackupSequencer struct {
    *Sequencer
    backupTracker *TransactionBackupTracker
    log           log.Logger
}

// 블록 빌딩 시작
func (bs *BackupSequencer) startBuildingBlock() {
    // 기존 코드...

    if !attrs.NoTxPool {
        // 백업 완료된 거래만 선택 (메인과 동일)
        attrs.Transactions = bs.selectBackedUpTransactions()
    }

    // 기존 코드...
}

// 백업 완료된 거래 선택
func (bs *BackupSequencer) selectBackedUpTransactions() []eth.Data {
    // 메인 시퀀서와 동일한 로직
    backedUpTxs := bs.backupTracker.GetBackedUpTransactions()

    var txData []eth.Data
    for _, tx := range backedUpTxs {
        data, err := tx.MarshalBinary()
        if err != nil {
            bs.log.Error("Failed to marshal transaction", "err", err)
            continue
        }
        txData = append(txData, data)
    }

    bs.log.Info("Selected backed up transactions for block",
        "count", len(txData))

    return txData
}
```

## 시퀀서 관리 규칙

멀티 시퀀서 환경에서는 누가 시퀀서가 될 수 있는지, 어떤 규칙으로 블록을 제출할 수 있는지를 명확히 정의해야 합니다. 이를 통해 시스템의 안정성과 공정성을 보장할 수 있습니다.

### 시퀀서 역할 결정 메커니즘

시퀀서가 자신의 역할(메인/백업)을 결정하는 방법은 다음과 같습니다:

#### 1. **Conductor 기반 역할 결정**
```go
type SequencerRoleManager struct {
    conductor    *OpConductor
    nodeAddress  common.Address
    role         SequencerRole
    log          log.Logger
}

type SequencerRole string

const (
    SequencerRoleMain   SequencerRole = "main"
    SequencerRoleBackup SequencerRole = "backup"
    SequencerRoleNone   SequencerRole = "none"
)

// Conductor에서 역할 확인
func (srm *SequencerRoleManager) DetermineRole() SequencerRole {
    // Conductor에 현재 노드의 역할 문의
    role, err := srm.conductor.GetNodeRole(srm.nodeAddress)
    if err != nil {
        srm.log.Error("Failed to get node role from conductor", "err", err)
        return SequencerRoleNone
    }

    srm.role = role
    srm.log.Info("Sequencer role determined", "role", role, "address", srm.nodeAddress)
    return role
}

// 주기적으로 역할 확인
func (srm *SequencerRoleManager) StartRoleMonitoring() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        newRole := srm.DetermineRole()
        if newRole != srm.role {
            srm.log.Info("Sequencer role changed",
                "old_role", srm.role, "new_role", newRole)
            srm.role = newRole
            srm.handleRoleChange(newRole)
        }
    }
}
```

#### 2. **리더 선출 기반 역할 결정**
```go
type LeaderElectionBasedRole struct {
    registry     *SequencerRegistry
    nodeAddress  common.Address
    isLeader     bool
    log          log.Logger
}

// 리더 선출 결과에 따른 역할 결정
func (ler *LeaderElectionBasedRole) DetermineRoleFromLeaderElection() SequencerRole {
    leader := ler.registry.GetCurrentLeader()

    if leader == ler.nodeAddress {
        ler.isLeader = true
        ler.log.Info("This node is elected as leader (main sequencer)")
        return SequencerRoleMain
    } else {
        ler.isLeader = false
        ler.log.Info("This node is backup sequencer", "leader", leader)
        return SequencerRoleBackup
    }
}
```

#### 3. **설정 기반 역할 결정**
```go
type ConfigBasedRole struct {
    config       *SequencerConfig
    role         SequencerRole
    log          log.Logger
}

// 설정 파일에서 역할 확인
func (cbr *ConfigBasedRole) DetermineRoleFromConfig() SequencerRole {
    role := cbr.config.GetSequencerRole()
    cbr.role = role

    cbr.log.Info("Sequencer role determined from config", "role", role)
    return role
}
```

#### 4. **동적 역할 전환**
```go
type DynamicRoleManager struct {
    roleManager  *SequencerRoleManager
    currentRole  SequencerRole
    log          log.Logger
}

// 역할 변경 처리
func (drm *DynamicRoleManager) handleRoleChange(newRole SequencerRole) {
    switch newRole {
    case SequencerRoleMain:
        drm.activateMainSequencer()
    case SequencerRoleBackup:
        drm.activateBackupSequencer()
    case SequencerRoleNone:
        drm.deactivateSequencer()
    }
}

// 메인 시퀀서 활성화
func (drm *DynamicRoleManager) activateMainSequencer() {
    drm.log.Info("Activating main sequencer")

    // 메인 시퀀서 로직 시작
    // - 블록 제출 권한 활성화
    // - 백업 서비스 시작
    // - 모니터링 강화
}

// 백업 시퀀서 활성화
func (drm *DynamicRoleManager) activateBackupSequencer() {
    drm.log.Info("Activating backup sequencer")

    // 백업 시퀀서 로직 시작
    // - 블록 제출 권한 비활성화 (대기 모드)
    // - 백업 서비스만 활성화
    // - 메인 시퀀서 장애 감시
}

### 5. **실행 파라미터 차이**

메인 시퀀서와 백업 시퀀서는 실행 시 다른 파라미터를 사용합니다:

#### 메인 시퀀서 실행 파라미터
```bash
# 메인 시퀀서 실행
op-node \
  --sequencer.enabled=true \
  --sequencer.role=main \
  --sequencer.block-submission-enabled=true \
  --sequencer.backup-service-enabled=true \
  --sequencer.backup-interval=100ms \
  --sequencer.max-backup-latency=2s \
  --sequencer.conductor-endpoint=http://conductor:8080 \
  --sequencer.node-address=0x1234567890123456789012345678901234567890 \
  --sequencer.priority=100 \
  --sequencer.backup-nodes=http://backup1:8545,http://backup2:8545 \
  --sequencer.monitoring-enabled=true \
  --sequencer.metrics-port=9090
```

#### 백업 시퀀서 실행 파라미터
```bash
# 백업 시퀀서 실행
op-node \
  --sequencer.enabled=true \
  --sequencer.role=backup \
  --sequencer.block-submission-enabled=false \
  --sequencer.backup-service-enabled=true \
  --sequencer.backup-interval=100ms \
  --sequencer.max-backup-latency=2s \
  --sequencer.conductor-endpoint=http://conductor:8080 \
  --sequencer.node-address=0x0987654321098765432109876543210987654321 \
  --sequencer.priority=50 \
  --sequencer.main-node=http://main:8545 \
  --sequencer.failover-enabled=true \
  --sequencer.health-check-interval=5s \
  --sequencer.monitoring-enabled=true \
  --sequencer.metrics-port=9091
```

#### 파라미터 설명

**공통 파라미터:**
- `--sequencer.enabled`: 시퀀서 기능 활성화
- `--sequencer.role`: 시퀀서 역할 (main/backup)
- `--sequencer.backup-service-enabled`: 백업 서비스 활성화
- `--sequencer.backup-interval`: 백업 간격
- `--sequencer.max-backup-latency`: 최대 백업 지연 시간
- `--sequencer.conductor-endpoint`: Conductor 서버 주소
- `--sequencer.node-address`: 현재 노드의 주소
- `--sequencer.priority`: 시퀀서 우선순위
- `--sequencer.monitoring-enabled`: 모니터링 활성화

**메인 시퀀서 전용:**
- `--sequencer.block-submission-enabled=true`: 블록 제출 활성화
- `--sequencer.backup-nodes`: 백업 노드 목록

**백업 시퀀서 전용:**
- `--sequencer.block-submission-enabled=false`: 블록 제출 비활성화
- `--sequencer.main-node`: 메인 노드 주소
- `--sequencer.failover-enabled`: 장애 대응 활성화
- `--sequencer.health-check-interval`: 메인 노드 상태 확인 간격

### 6. **메인 시퀀서 발견 메커니즘**

백업 시퀀서가 메인 시퀀서를 찾는 방법은 여러 가지가 있습니다:

#### A. **Conductor 기반 메인 시퀀서 발견**
```go
type MainSequencerDiscovery struct {
    conductor    *OpConductor
    log          log.Logger
}

// Conductor에서 현재 메인 시퀀서 조회
func (msd *MainSequencerDiscovery) DiscoverMainSequencer() (*SequencerInfo, error) {
    // Conductor에 현재 메인 시퀀서 문의
    mainSequencer, err := msd.conductor.GetCurrentMainSequencer()
    if err != nil {
        return nil, fmt.Errorf("failed to get main sequencer from conductor: %w", err)
    }

    msd.log.Info("Main sequencer discovered",
        "address", mainSequencer.Address,
        "endpoint", mainSequencer.Endpoint)

    return mainSequencer, nil
}

// 주기적으로 메인 시퀀서 상태 확인
func (msd *MainSequencerDiscovery) MonitorMainSequencer() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        mainSeq, err := msd.DiscoverMainSequencer()
        if err != nil {
            msd.log.Error("Failed to discover main sequencer", "err", err)
            continue
        }

        // 메인 시퀀서 상태 확인
        if !msd.isMainSequencerHealthy(mainSeq) {
            msd.log.Warn("Main sequencer is unhealthy", "address", mainSeq.Address)
            // 장애 대응 로직 실행
            msd.handleMainSequencerFailure()
        }
    }
}
```

#### B. **리더 선출 기반 메인 시퀀서 발견**
```go
type LeaderBasedDiscovery struct {
    registry     *SequencerRegistry
    log          log.Logger
}

// 리더 선출 결과로 메인 시퀀서 확인
func (lbd *LeaderBasedDiscovery) DiscoverMainSequencerFromLeader() (*SequencerInfo, error) {
    leader := lbd.registry.GetCurrentLeader()
    if leader == (common.Address{}) {
        return nil, fmt.Errorf("no leader elected")
    }

    // 리더 정보 조회
    leaderInfo, exists := lbd.registry.GetSequencer(leader)
    if !exists {
        return nil, fmt.Errorf("leader not found in registry: %s", leader)
    }

    lbd.log.Info("Main sequencer discovered from leader election",
        "leader", leader, "endpoint", leaderInfo.Endpoint)

    return leaderInfo, nil
}
```

#### C. **설정 기반 메인 시퀀서 발견**
```go
type ConfigBasedDiscovery struct {
    config       *SequencerConfig
    log          log.Logger
}

// 설정 파일에서 메인 시퀀서 주소 확인
func (cbd *ConfigBasedDiscovery) DiscoverMainSequencerFromConfig() (*SequencerInfo, error) {
    mainNodeEndpoint := cbd.config.GetMainNodeEndpoint()
    if mainNodeEndpoint == "" {
        return nil, fmt.Errorf("main node endpoint not configured")
    }

    // 메인 노드 연결 테스트
    if !cbd.testConnection(mainNodeEndpoint) {
        return nil, fmt.Errorf("cannot connect to main node: %s", mainNodeEndpoint)
    }

    mainSeq := &SequencerInfo{
        Address:  cbd.config.GetMainNodeAddress(),
        Endpoint: mainNodeEndpoint,
        Status:   SequencerStatusActive,
    }

    cbd.log.Info("Main sequencer discovered from config",
        "endpoint", mainNodeEndpoint)

    return mainSeq, nil
}
```

#### D. **동적 메인 시퀀서 발견**
```go
type DynamicMainSequencerDiscovery struct {
    discoveryMethods []MainSequencerDiscoveryMethod
    currentMain      *SequencerInfo
    log              log.Logger
}

type MainSequencerDiscoveryMethod interface {
    Discover() (*SequencerInfo, error)
    GetPriority() int
}

// 여러 방법으로 메인 시퀀서 발견 시도
func (dmds *DynamicMainSequencerDiscovery) DiscoverMainSequencer() (*SequencerInfo, error) {
    // 우선순위 순으로 발견 방법 시도
    sort.Slice(dmds.discoveryMethods, func(i, j int) bool {
        return dmds.discoveryMethods[i].GetPriority() > dmds.discoveryMethods[j].GetPriority()
    })

    for _, method := range dmds.discoveryMethods {
        mainSeq, err := method.Discover()
        if err != nil {
            dmds.log.Debug("Discovery method failed",
                "method", reflect.TypeOf(method), "err", err)
            continue
        }

        dmds.currentMain = mainSeq
        dmds.log.Info("Main sequencer discovered",
            "method", reflect.TypeOf(method),
            "address", mainSeq.Address)

        return mainSeq, nil
    }

    return nil, fmt.Errorf("all discovery methods failed")
}
```

#### E. **메인 시퀀서 상태 확인**
```go
type MainSequencerHealthChecker struct {
    mainSequencer *SequencerInfo
    log           log.Logger
}

// 메인 시퀀서 상태 확인
func (mshc *MainSequencerHealthChecker) CheckMainSequencerHealth() bool {
    // 1. 네트워크 연결 확인
    if !mshc.checkNetworkConnectivity() {
        mshc.log.Warn("Main sequencer network connectivity failed")
        return false
    }

    // 2. 블록 생성 상태 확인
    if !mshc.checkBlockGeneration() {
        mshc.log.Warn("Main sequencer block generation failed")
        return false
    }

    // 3. 백업 서비스 상태 확인
    if !mshc.checkBackupService() {
        mshc.log.Warn("Main sequencer backup service failed")
        return false
    }

    mshc.log.Debug("Main sequencer health check passed")
    return true
}

// 네트워크 연결 확인
func (mshc *MainSequencerHealthChecker) checkNetworkConnectivity() bool {
    client := ethclient.NewClient(mshc.mainSequencer.Endpoint)
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := client.BlockNumber(ctx)
    return err == nil
}

// 블록 생성 상태 확인
func (mshc *MainSequencerHealthChecker) checkBlockGeneration() bool {
    client := ethclient.NewClient(mshc.mainSequencer.Endpoint)
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 최근 블록 생성 시간 확인
    latestBlock, err := client.BlockByNumber(ctx, nil)
    if err != nil {
        return false
    }

    // 30초 내에 블록이 생성되었는지 확인
    blockTime := time.Unix(int64(latestBlock.Time()), 0)
    return time.Since(blockTime) < 30*time.Second
}
```

#### 설정 파일 예시

**메인 시퀀서 설정 (main-sequencer.yaml):**
```yaml
sequencer:
  enabled: true
  role: "main"
  block_submission_enabled: true
  backup_service_enabled: true
  backup_interval: "100ms"
  max_backup_latency: "2s"
  conductor_endpoint: "http://conductor:8080"
  node_address: "0x1234567890123456789012345678901234567890"
  priority: 100
  backup_nodes:
    - "http://backup1:8545"
    - "http://backup2:8545"
  monitoring:
    enabled: true
    metrics_port: 9090
```

**백업 시퀀서 설정 (backup-sequencer.yaml):**
```yaml
sequencer:
  enabled: true
  role: "backup"
  block_submission_enabled: false
  backup_service_enabled: true
  backup_interval: "100ms"
  max_backup_latency: "2s"
  conductor_endpoint: "http://conductor:8080"
  node_address: "0x0987654321098765432109876543210987654321"
  priority: 50
  main_node: "http://main:8545"
  failover:
    enabled: true
    health_check_interval: "5s"
  monitoring:
    enabled: true
    metrics_port: 9091
```

### 1. 시퀀서 등록 및 관리

시퀀서 등록 시스템은 누가 시퀀서가 될 수 있는지를 관리하는 핵심 컴포넌트입니다. 이 시스템을 통해 신뢰할 수 있는 시퀀서만 시스템에 참여할 수 있도록 합니다.

#### 시퀀서 등록 시스템

시퀀서 등록 시스템은 다음과 같은 주요 기능을 제공합니다:

- **시퀀서 등록**: 새로운 시퀀서의 등록 및 검증
- **권한 관리**: 각 시퀀서의 블록 제출 권한 및 백업 권한 관리
- **상태 추적**: 시퀀서의 활성 상태, 성능 지표 등을 실시간 추적
- **우선순위 관리**: 시퀀서 간의 블록 제출 우선순위 관리

이러한 시스템을 통해 시스템에 참여하는 시퀀서들의 신뢰성과 성능을 보장할 수 있습니다.
```go
type SequencerRegistry struct {
    // 등록된 시퀀서 목록
    sequencers map[common.Address]*SequencerInfo

    // 현재 활성 시퀀서
    activeSequencer common.Address

    // 시퀀서 권한 관리
    permissions map[common.Address]SequencerPermission

    mu sync.RWMutex
    log log.Logger
}

type SequencerInfo struct {
    Address     common.Address
    PublicKey   []byte
    Endpoint    string
    Status      SequencerStatus
    Registered  time.Time
    LastActive  time.Time
    Performance SequencerPerformance
}

type SequencerStatus string

const (
    SequencerStatusInactive SequencerStatus = "inactive"
    SequencerStatusActive   SequencerStatus = "active"
    SequencerStatusBackup   SequencerStatus = "backup"
    SequencerStatusFaulty   SequencerStatus = "faulty"
)

type SequencerPermission struct {
    CanSubmitBlocks bool
    CanBackupTx     bool
    Priority        int // 높을수록 우선순위
    MaxLatency      time.Duration
}
```

#### 시퀀서 등록 프로세스
```go
// 새로운 시퀀서 등록
func (sr *SequencerRegistry) RegisterSequencer(
    address common.Address,
    publicKey []byte,
    endpoint string,
    permission SequencerPermission,
) error {
    sr.mu.Lock()
    defer sr.mu.Unlock()

    // 중복 등록 확인
    if _, exists := sr.sequencers[address]; exists {
        return fmt.Errorf("sequencer already registered: %s", address)
    }

    // 권한 검증
    if !sr.validatePermission(permission) {
        return fmt.Errorf("invalid permission for sequencer: %s", address)
    }

    // 시퀀서 정보 등록
    sr.sequencers[address] = &SequencerInfo{
        Address:    address,
        PublicKey:  publicKey,
        Endpoint:   endpoint,
        Status:     SequencerStatusInactive,
        Registered: time.Now(),
        LastActive: time.Now(),
    }

    sr.permissions[address] = permission
    sr.log.Info("Sequencer registered", "address", address, "permission", permission)

    return nil
}
```

### 2. 블록 제출 룰

블록 제출 룰은 여러 시퀀서가 동시에 블록을 제출하려고 할 때 어떤 블록을 선택할지를 결정하는 규칙입니다. 이 규칙을 통해 시스템의 일관성과 공정성을 보장할 수 있습니다.

#### 블록 제출 우선순위

블록 제출 우선순위는 다음과 같은 요소들을 종합적으로 고려하여 결정됩니다:

- **시퀀서 우선순위**: 등록 시 설정된 시퀀서의 우선순위
- **백업 완료 조건**: 백업이 완료된 거래의 수와 지연 시간
- **시퀀서 상태**: 시퀀서의 현재 상태 (활성, 백업, 장애 등)
- **제출 제한**: 블록 제출 간격 및 지연 시간 제한

이러한 우선순위 시스템을 통해 시스템의 안정성을 유지하면서도 공정한 블록 제출 기회를 제공할 수 있습니다.
```go
type BlockSubmissionRule struct {
    registry *SequencerRegistry

    // 블록 제출 룰
    submissionRules []SubmissionRule

    // 충돌 해결 규칙
    conflictResolution ConflictResolution

    log log.Logger
}

type SubmissionRule struct {
    // 시퀀서 우선순위
    Priority int

    // 블록 제출 조건
    Conditions []SubmissionCondition

    // 제출 제한
    MaxSubmissionDelay time.Duration
    MinBlockInterval   time.Duration
}

type SubmissionCondition struct {
    // 백업 완료 거래 수
    MinBackedUpTxs int

    // 백업 지연 시간
    MaxBackupLatency time.Duration

    // 시퀀서 상태
    RequiredStatus SequencerStatus
}

// 블록 제출 권한 확인
func (bsr *BlockSubmissionRule) CanSubmitBlock(
    sequencerAddr common.Address,
    block *types.Block,
    backupStats BackupStatistics,
) (bool, error) {
    // 시퀀서 등록 확인
    sequencer, exists := bsr.registry.GetSequencer(sequencerAddr)
    if !exists {
        return false, fmt.Errorf("sequencer not registered: %s", sequencerAddr)
    }

    // 권한 확인
    permission := bsr.registry.GetPermission(sequencerAddr)
    if !permission.CanSubmitBlocks {
        return false, fmt.Errorf("sequencer not authorized to submit blocks: %s", sequencerAddr)
    }

    // 조건 확인
    for _, rule := range bsr.submissionRules {
        if rule.Priority == permission.Priority {
            return bsr.checkSubmissionConditions(rule, backupStats, sequencer)
        }
    }

    return false, fmt.Errorf("no matching submission rule for sequencer: %s", sequencerAddr)
}
```

### 3. 충돌 해결 규칙

충돌 해결 규칙은 여러 시퀀서가 동시에 블록을 제출했을 때 어떤 블록을 최종적으로 선택할지를 결정하는 메커니즘입니다. 이 규칙을 통해 시스템의 일관성을 유지할 수 있습니다.

#### 동시 블록 제출 처리

동시 블록 제출 상황에서는 다음과 같은 전략을 사용하여 충돌을 해결합니다:

- **우선순위 기반 해결**: 시퀀서의 우선순위가 높은 블록을 선택
- **합의 기반 해결**: 여러 시퀀서의 투표를 통해 블록을 선택
- **시간 기반 해결**: 먼저 제출된 블록을 선택

이러한 충돌 해결 메커니즘을 통해 시스템의 안정성을 유지하면서도 공정한 블록 선택을 보장할 수 있습니다.
```go
type ConflictResolution struct {
    // 충돌 해결 전략
    strategy ConflictResolutionStrategy

    // 투표 기간
    votingPeriod time.Duration

    // 최소 동의 수
    minConsensus int

    log log.Logger
}

type ConflictResolutionStrategy string

const (
    StrategyPriorityBased    ConflictResolutionStrategy = "priority_based"
    StrategyConsensusBased   ConflictResolutionStrategy = "consensus_based"
    StrategyTimeBased        ConflictResolutionStrategy = "time_based"
)

// 충돌 해결
func (cr *ConflictResolution) ResolveConflict(
    blocks []*types.Block,
    sequencers []common.Address,
) (*types.Block, error) {
    switch cr.strategy {
    case StrategyPriorityBased:
        return cr.resolveByPriority(blocks, sequencers)
    case StrategyConsensusBased:
        return cr.resolveByConsensus(blocks, sequencers)
    case StrategyTimeBased:
        return cr.resolveByTime(blocks)
    default:
        return nil, fmt.Errorf("unknown conflict resolution strategy: %s", cr.strategy)
    }
}
```

## 백업 동기화 메커니즘

백업 동기화 메커니즘은 메인 노드와 백업 노드 간의 거래를 실시간으로 동기화하는 핵심 시스템입니다. 이 메커니즘을 통해 백업 노드가 메인 노드와 동일한 거래 상태를 유지할 수 있습니다.

### 1. 실시간 거래 백업

실시간 거래 백업은 메인 노드에서 수신된 거래를 즉시 백업 노드로 전파하는 시스템입니다. 이를 통해 백업 노드가 메인 노드와 최대한 동기화된 상태를 유지할 수 있습니다.

#### 백업 서비스 구현

백업 서비스는 다음과 같은 주요 기능을 제공합니다:

- **실시간 거래 전파**: 메인 노드에서 수신된 거래를 즉시 백업 노드로 전송
- **백업 상태 모니터링**: 백업 성공률, 지연 시간 등을 실시간으로 모니터링
- **재시도 메커니즘**: 백업 실패 시 자동으로 재시도하여 신뢰성 보장
- **성능 최적화**: 백업 지연 시간을 최소화하여 시스템 성능 향상

이러한 백업 서비스를 통해 메인 노드와 백업 노드 간의 거래 동기화를 안정적으로 유지할 수 있습니다.
```go
type TransactionBackupService struct {
    mainNode   *MainNode
    backupNode *BackupNode
    tracker    *TransactionBackupTracker
    log        log.Logger

    // 백업 설정
    backupInterval time.Duration
    maxRetries     int
}

// 백업 서비스 시작
func (tbs *TransactionBackupService) StartBackup() {
    tbs.log.Info("Starting transaction backup service")

    // 백업 모니터링 시작
    go tbs.backupNewTransactions()
    go tbs.MonitorBackupStatus()
}

// 새 거래 백업
func (tbs *TransactionBackupService) backupNewTransactions() {
    ticker := time.NewTicker(tbs.backupInterval)
    defer ticker.Stop()

    for range ticker.C {
        // 메인 노드에서 새 거래 가져오기
        newTxs := tbs.mainNode.GetNewTransactions()

        for _, tx := range newTxs {
            // 백업 노드로 거래 전송
            err := tbs.backupNode.SendTransaction(tx)
            if err != nil {
                tbs.log.Error("Failed to backup transaction",
                    "hash", tx.Hash(), "err", err)
                continue
            }

            // 백업 추적기 업데이트
            tbs.tracker.AddTransactionToBackup(tx)
        }
    }
}
```

### 2. 백업 상태 모니터링

백업 상태 모니터링은 백업 시스템의 성능과 안정성을 실시간으로 추적하는 시스템입니다. 이를 통해 백업 시스템의 문제를 조기에 발견하고 대응할 수 있습니다.

#### 백업 통계 모니터링

백업 통계 모니터링은 다음과 같은 주요 지표들을 추적합니다:

- **총 거래 수**: 메인 노드에서 수신된 총 거래 수
- **백업 완료 거래 수**: 백업 노드에서 성공적으로 수신된 거래 수
- **백업 성공률**: 백업 완료 거래 수 / 총 거래 수
- **평균 지연 시간**: 메인 수신부터 백업 완료까지의 평균 시간
- **최대/최소 지연 시간**: 백업 지연 시간의 범위

이러한 통계 정보를 통해 백업 시스템의 성능을 지속적으로 모니터링하고 개선할 수 있습니다.
```go
type BackupStatistics struct {
    TotalTransactions    int
    BackedUpTransactions int
    BackupRate           float64
    AverageLatency       time.Duration
    MaxLatency           time.Duration
    MinLatency           time.Duration
}

func (tbs *TransactionBackupService) MonitorBackupStatus() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats := tbs.tracker.GetBackupStatistics()
        tbs.log.Info("Backup statistics",
            "total_txs", stats.TotalTransactions,
            "backed_up_txs", stats.BackedUpTransactions,
            "backup_rate", fmt.Sprintf("%.2f%%", stats.BackupRate*100),
            "avg_latency", stats.AverageLatency,
            "max_latency", stats.MaxLatency,
            "min_latency", stats.MinLatency)
    }
}
```

## 장점 및 특징

### 1. 거래 순서 보장
- **결정적 거래 선택**: 백업 완료된 거래만 블록에 포함
- **기존 룰 유지**: Gas Price 우선순위 + Nonce + 도착 시간 순서
- **일관성**: 메인과 백업 시퀀서가 동일한 블록 생성

### 2. 안정성
- **백업 노드 장애 대응**: 백업 노드 장애 시에도 메인 노드 계속 운영
- **거래 손실 방지**: 백업 완료된 거래만 처리하여 일관성 보장
- **복구 능력**: 백업 노드 복구 후 자동 동기화

### 3. 성능
- **백업 지연 시간 제한**: 설정 가능한 백업 지연 시간으로 성능 보장
- **점진적 처리**: 백업 완료되지 않은 거래는 다음 블록에서 처리
- **병렬 처리**: 메인과 백업 노드가 병렬로 거래 처리

### 4. 모니터링 및 관찰성
- **실시간 모니터링**: 백업 상태 실시간 추적
- **지연 시간 분석**: 백업 지연 시간 상세 분석
- **통계 정보**: 백업 성공률, 지연 시간 통계 제공

### 5. 확장성
- **다중 백업 노드**: 여러 백업 노드 지원 가능
- **설정 가능**: 백업 지연 시간, 재시도 횟수 등 설정 가능
- **플러그인 구조**: 다양한 백업 전략 지원

## 구현 예시

### 1. Conductor 통합

#### Conductor 서비스 수정
```go
// op-conductor/conductor/service.go
func (oc *OpConductor) startSequencer() error {
    ctx := context.Background()

    // 백업 완료 거래 풀 초기화
    err := oc.initializeBackupPool()
    if err != nil {
        return fmt.Errorf("failed to initialize backup pool: %w", err)
    }

    // 백업 서비스 시작
    oc.backupService.StartBackup()

    // 시퀀서 시작
    err = oc.ctrl.StartSequencer(ctx, latestBlockHash)
    if err != nil {
        return fmt.Errorf("failed to start sequencer: %w", err)
    }

    oc.seqActive.Store(true)
    oc.log.Info("Sequencer started with backup pool")
    return nil
}

// 백업 풀 초기화
func (oc *OpConductor) initializeBackupPool() error {
    // 백업 추적기 생성
    oc.backupTracker = NewTransactionBackupTracker(
        oc.mainNode.GetTransactionPool(),
        oc.backupNode.GetTransactionPool(),
        oc.log,
    )

    // 백업 서비스 생성
    oc.backupService = NewTransactionBackupService(
        oc.mainNode,
        oc.backupNode,
        oc.backupTracker,
        oc.log,
    )

    return nil
}
```

### 2. 설정 파일

#### 백업 설정
```yaml
# backup-config.yaml
backup:
  # 백업 지연 시간 제한
  max_latency: "2s"

  # 백업 간격
  backup_interval: "100ms"

  # 최대 재시도 횟수
  max_retries: 3

  # 백업 노드 상태 확인 간격
  health_check_interval: "5s"

  # 백업 완료 조건
  completion_criteria:
    backup_received: true
    max_latency: "2s"
    backup_node_healthy: true
```

### 3. 메트릭스

#### 백업 메트릭스 정의
```go
type BackupMetrics interface {
    // 백업 성공률
    RecordBackupSuccess(txHash common.Hash, latency time.Duration)

    // 백업 실패
    RecordBackupFailure(txHash common.Hash, error string)

    // 백업 지연 시간
    RecordBackupLatency(latency time.Duration)

    // 백업 풀 크기
    SetBackupPoolSize(size int)

    // 백업 완료 거래 수
    SetBackedUpTransactionCount(count int)
}
```

## 결론

백업 완료 거래 풀 기반 블록 생성 방식은 멀티 시퀀서 환경에서 거래 순서를 보장하는 효과적인 솔루션입니다.

### 주요 장점
1. **거래 순서 보장**: 백업 완료된 거래만 처리하여 일관성 보장
2. **안정성**: 백업 노드 장애 시에도 메인 노드 운영 가능
3. **성능**: 설정 가능한 백업 지연 시간으로 성능 최적화
4. **관찰성**: 실시간 백업 상태 모니터링 및 통계 제공
5. **확장성**: 다중 백업 노드 및 다양한 설정 지원

### 적용 시나리오
- **고가용성 요구**: 24/7 서비스 운영이 필요한 환경
- **거래 순서 중요**: MEV 방지가 중요한 환경
- **안정성 우선**: 거래 손실을 최소화해야 하는 환경
- **모니터링 필요**: 백업 상태를 실시간으로 추적해야 하는 환경

이 방식을 통해 메인과 백업 시퀀서가 모두 운영되면서도 일관된 거래 순서를 보장할 수 있습니다.
