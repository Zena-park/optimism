# Thokamak-thanos Q&A

Repo : https://github.com/tokamak-network/tokamak-thanos


## Thokamak-thanos는 Dispute Game 의 type을 무엇으로 설정하고 배포하는가?


### 1. 정의된 Game Type들

`packages/tokamak/contracts-bedrock/src/dispute/lib/Types.sol`에서 다음과 같은 게임 타입들이 정의되어 있습니다:

```solidity
library GameTypes {
    /// @dev A dispute game type the uses the cannon vm.
    GameType internal constant CANNON = GameType.wrap(0);

    /// @dev A permissioned dispute game type the uses the cannon vm.
    GameType internal constant PERMISSIONED_CANNON = GameType.wrap(1);

    /// @notice A dispute game type the uses the asterisc VM
    GameType internal constant ASTERISC = GameType.wrap(2);

    /// @notice A dispute game type with short game duration for testing withdrawals.
    ///         Not intended for production use.
    GameType internal constant FAST = GameType.wrap(254);

    /// @notice A dispute game type that uses an alphabet vm.
    ///         Not intended for production use.
    GameType internal constant ALPHABET = GameType.wrap(255);
}
```

### 2. 배포에서 설정하는 Respected Game Type

**현재 배포 설정에서는 `respectedGameType`을 `0`으로 설정**하고 있습니다:

- **thanos-sepolia.json**: `"respectedGameType": 0`
- **thanos-stack-sepolia.json**: `"respectedGameType": 0`
- **hardhat.json**: `"respectedGameType": 0`

이는 **`GameTypes.CANNON` (값: 0)**을 의미합니다.

### 3. 배포 과정에서의 설정

#### OptimismPortal2 초기화 시:
```solidity
OptimismPortal2.initialize(
    DisputeGameFactory(disputeGameFactoryProxy),
    SystemConfig(systemConfigProxy),
    SuperchainConfig(superchainConfigProxy),
    GameType.wrap(uint32(cfg.respectedGameType())) // 0 = CANNON
)
```

#### DisputeGameFactory에 구현체 설정:
`Deploy.s.sol`에서 다음 게임 타입들에 대한 구현체를 설정합니다:

- `gameType: GameTypes.CANNON` (0)
- `gameType: GameTypes.PERMISSIONED_CANNON` (1)
- `gameType: GameTypes.ALPHABET` (255)
- `gameType: GameTypes.ASTERISC` (2)
- `gameType: GameTypes.FAST` (254)

### 4. 결론

**Tokamak Thanos에서는 Dispute Game의 타입을 `CANNON` (값: 0)으로 설정하여 배포**합니다.

- **Respected Game Type**: `CANNON` (GameType.wrap(0))
- **사용 VM**: Cannon VM
- **용도**: 프로덕션 환경에서 사용되는 메인 dispute game 타입
- **특징**: 권한 없이 누구나 dispute game을 생성할 수 있는 일반적인 타입

이 설정은 Optimism의 fault proof 시스템에서 Cannon VM을 기반으로 한 dispute resolution을 사용한다는 것을 의미합니다.

### 5. 관련 파일 위치

- **Game Type 정의**: `packages/tokamak/contracts-bedrock/src/dispute/lib/Types.sol`
- **배포 설정**:
  - `packages/tokamak/contracts-bedrock/deploy-config/thanos-sepolia.json`
  - `packages/tokamak/contracts-bedrock/deploy-config/thanos-stack-sepolia.json`
- **배포 스크립트**: `packages/tokamak/contracts-bedrock/scripts/Deploy.s.sol`
- **OptimismPortal2 컨트랙트**: `packages/tokamak/contracts-bedrock/src/L1/OptimismPortal2.sol`

---

## 시퀀서(Sequencer)와 프로포저(Proposer)의 역할

### 시퀀서(Sequencer)

#### 주요 역할
1. **트랜잭션 수집 및 순서 결정**: L2에서 발생하는 사용자 트랜잭션들을 수집하고 실행 순서를 결정
2. **블록 생성**: 트랜잭션들을 포함한 L2 블록을 생성
3. **실시간 블록 프로덕션**: 지속적으로 새로운 L2 블록을 생성하여 네트워크를 운영

#### 코드에서 확인한 구체적 동작
```go
// op-node/rollup/driver/sequencer.go
func (d *Sequencer) StartBuildingBlock(ctx context.Context) error {
    l2Head := d.engine.UnsafeL2Head()

    // L1 origin 블록 결정
    l1Origin, err := d.l1OriginSelector.FindL1Origin(ctx, l2Head)

    // 블록 속성 준비
    attrs, err := d.attrBuilder.PreparePayloadAttributes(fetchCtx, l2Head, l1Origin.ID())

    // 트랜잭션 풀에서 트랜잭션 포함 여부 결정
    attrs.NoTxPool = uint64(attrs.Timestamp) > l1Origin.Time+d.spec.MaxSequencerDrift(l1Origin.Time)

    // 엔진에 블록 생성 요청
    return d.engine.StartPayload(ctx, l2Head, attrs, async.NoOpGossiper{})
}
```

#### L1에서의 역할
시퀀서는 **직접적으로 L1과 상호작용하지 않습니다**. 대신 다음과 같이 동작:
- L1 블록을 참조하여 L2 블록의 순서와 타임스탬프 결정
- 생성된 L2 블록 데이터는 **Batcher**가 L1에 제출

### 배처(Batcher)

#### 주요 역할
1. **L2 트랜잭션 데이터 수집**: 시퀀서가 생성한 L2 블록들의 트랜잭션 데이터 수집
2. **데이터 압축 및 배치**: 여러 L2 블록의 트랜잭션 데이터를 압축하고 배치로 묶음
3. **L1 제출**: 배치된 데이터를 L1의 BatchInbox에 제출

#### 코드에서 확인한 구체적 동작
```go
// op-batcher/batcher/driver.go
func (l *BatchSubmitter) sendTransaction(ctx context.Context, txdata txData, queue *txmgr.Queue[txID], receiptsCh chan txmgr.TxReceipt[txID]) error {
    var candidate *txmgr.TxCandidate
    if l.Config.UseBlobs {
        // Blob 트랜잭션으로 제출
        candidate, err = l.blobTxCandidate(txdata)
    } else {
        // Calldata 트랜잭션으로 제출
        data := txdata.CallData()
        candidate = l.calldataTxCandidate(data)
    }

    // L1 BatchInbox에 트랜잭션 전송
    queue.Send(txdata.ID(), *candidate, receiptsCh)
    return nil
}
```

#### L1에서의 역할
- **BatchInbox 컨트랙트에 데이터 제출**: `RollupConfig.BatchInboxAddress`로 트랜잭션 데이터 전송
- **데이터 가용성(DA) 보장**: L2 트랜잭션 데이터가 L1에서 누구나 접근 가능하도록 보장

### 프로포저(Proposer)

#### 주요 역할
1. **L2 상태 루트 계산**: 주기적으로 L2 체인의 상태 루트를 계산
2. **Output Root 제출**: 계산된 상태 루트를 L1에 제출
3. **Dispute Game 생성**: Fault Proof 시스템에서 Dispute Game을 생성

#### 코드에서 확인한 구체적 동작
```go
// op-proposer/proposer/driver.go
func (l *L2OutputSubmitter) proposeOutput(ctx context.Context, output *eth.OutputResponse) {
    if err := l.sendTransaction(cCtx, output); err != nil {
        l.Log.Error("Failed to send proposal transaction", "err", err)
        return
    }
    l.Metr.RecordL2BlocksProposed(output.BlockRef)
}

// L2OutputOracle에 제출하는 경우
func proposeL2OutputTxData(abi *abi.ABI, output *eth.OutputResponse) ([]byte, error) {
    return abi.Pack("proposeL2Output",
        output.OutputRoot,                              // L2 상태 루트
        new(big.Int).SetUint64(output.BlockRef.Number), // L2 블록 번호
        output.Status.CurrentL1.Hash,                   // 참조 L1 블록 해시
        new(big.Int).SetUint64(output.Status.CurrentL1.Number)) // 참조 L1 블록 번호
}

// DisputeGameFactory에 제출하는 경우 (Fault Proof)
func proposeL2OutputDGFTxData(abi *abi.ABI, gameType uint32, output *eth.OutputResponse) ([]byte, error) {
    return abi.Pack("create",
        gameType,           // Dispute Game 타입 (CANNON = 0)
        output.OutputRoot,  // L2 상태 루트 (Claim)
        math.U256Bytes(new(big.Int).SetUint64(output.BlockRef.Number))) // L2 블록 번호
}
```

#### L1에서의 역할
1. **L2OutputOracle 컨트랙트 상호작용**:
   - `proposeL2Output()` 함수 호출
   - L2 상태 루트, L2 블록 번호, L1 참조 블록 정보 제출

2. **DisputeGameFactory 컨트랙트 상호작용** (Fault Proof 활성화 시):
   - `create()` 함수 호출
   - Dispute Game 생성 (게임 타입: CANNON)
   - 본드(Bond) 예치 필요

### 요약

| 컴포넌트 | 주요 역할 | L1 상호작용 |
|---------|----------|------------|
| **시퀀서** | L2 트랜잭션 순서 결정 및 블록 생성 | 없음 (L1 블록 참조만) |
| **배처** | L2 트랜잭션 데이터 수집 및 L1 제출 | BatchInbox에 데이터 제출 |
| **프로포저** | L2 상태 루트 계산 및 제출 | L2OutputOracle 또는 DisputeGameFactory에 상태 루트 제출 |

---

## L2OutputOracle에 proposeL2Output() 호출 시점

### 호출 조건 및 시점

#### 1. **주기적 폴링 (PollInterval)**
```go
// op-proposer/proposer/driver.go - loopL2OO()
func (l *L2OutputSubmitter) loopL2OO(ctx context.Context) {
    ticker := time.NewTicker(l.Cfg.PollInterval)  // 기본값: 설정에 따라
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            output, shouldPropose, err := l.FetchNextOutputInfo(ctx)
            if err != nil || !shouldPropose {
                break
            }
            l.proposeOutput(ctx, output)  // 실제 제출
        case <-l.done:
            return
        }
    }
}
```

#### 2. **제출 조건 확인 (FetchNextOutputInfo)**
다음 모든 조건을 만족해야 제출:

**A. 제출 간격 조건**
```go
// L2OutputOracle 컨트랙트에서 다음 제출해야 할 블록 번호 확인
nextCheckpointBlock, err := l.l2ooContract.NextBlockNumber(callOpts)

// 현재 L2 블록 번호가 제출해야 할 블록 번호에 도달했는지 확인
if currentBlockNumber.Cmp(nextCheckpointBlock) < 0 {
    return nil, false, nil  // 아직 제출 시점이 아님
}
```

**B. L2 블록 준비 상태 확인**
```go
// L2 블록이 finalized 되었거나, AllowNonFinalized=true인 경우 safe 상태여야 함
if output.BlockRef.Number > output.Status.FinalizedL2.Number &&
   (!l.Cfg.AllowNonFinalized || output.BlockRef.Number > output.Status.SafeL2.Number) {
    return nil, false, nil  // 블록이 아직 준비되지 않음
}
```

#### 3. **L2OutputOracle 컨트랙트의 제출 간격 계산**
```solidity
// packages/tokamak/contracts-bedrock/src/L1/L2OutputOracle.sol
function nextBlockNumber() public view returns (uint256) {
    return latestBlockNumber() + submissionInterval;  // 마지막 제출된 블록 + 제출 간격
}

function computeL2Timestamp(uint256 _l2BlockNumber) public view returns (uint256) {
    return startingTimestamp + ((_l2BlockNumber - startingBlockNumber) * l2BlockTime);
}
```

#### 4. **실제 제출 조건 (proposeL2Output)**
```solidity
function proposeL2Output(
    bytes32 _outputRoot,
    uint256 _l2BlockNumber,
    bytes32 _l1BlockHash,
    uint256 _l1BlockNumber
) external payable {
    require(msg.sender == proposer, "L2OutputOracle: only the proposer address can propose new outputs");

    require(
        _l2BlockNumber == nextBlockNumber(),  // 정확한 블록 번호여야 함
        "L2OutputOracle: block number must be equal to next expected block number"
    );

    require(
        computeL2Timestamp(_l2BlockNumber) < block.timestamp,  // 미래 블록 제출 불가
        "L2OutputOracle: cannot propose L2 output in the future"
    );

    require(_outputRoot != bytes32(0), "L2OutputOracle: L2 output proposal cannot be the zero hash");
    // ... L1 블록 해시 검증 등
}
```

### 요약

**L2OutputOracle에 `proposeL2Output()` 호출 시점:**

1. **주기**: `PollInterval` 마다 확인 (설정 가능한 주기)
2. **블록 번호 조건**: `현재 L2 블록 번호 >= (마지막 제출 블록 + submissionInterval)`
3. **블록 상태 조건**: L2 블록이 finalized 상태이거나 (AllowNonFinalized=true인 경우) safe 상태
4. **시간 조건**: 해당 L2 블록의 타임스탬프가 현재 시간보다 이전
5. **권한 조건**: 지정된 proposer 주소만 제출 가능

**핵심**: `submissionInterval` (제출 간격)에 따라 정기적으로, 그리고 L2 블록이 충분히 안정화된 후에 제출됩니다.

---

## L2 운영비 메커니즘

### 1. **시퀀서 운영비 (Sequencer Fees)**

#### 수익 메커니즘
시퀀서는 **L2 트랜잭션 수수료**를 통해 운영비를 확보합니다:

```solidity
// packages/tokamak/contracts-bedrock/src/L2/SequencerFeeVault.sol
contract SequencerFeeVault is FeeVault {
    constructor(
        address _recipient,           // 시퀀서 수익 수취 주소
        uint256 _minWithdrawalAmount, // 최소 인출 금액
        WithdrawalNetwork _withdrawalNetwork // L1 또는 L2 네트워크
    )
}
```

#### 수수료 수집 과정
1. **L2 트랜잭션 처리 시**: 사용자가 지불한 가스비 중 일부가 SequencerFeeVault로 전송
2. **Fee Vault 누적**: `0x4200000000000000000000000000000000000011` 주소에 수수료 누적
3. **인출 메커니즘**:
   ```solidity
   function withdraw() external {
       require(address(this).balance >= MIN_WITHDRAWAL_AMOUNT);
       uint256 value = address(this).balance;

       if (WITHDRAWAL_NETWORK == WithdrawalNetwork.L2) {
           // L2에서 직접 수취
           (bool success,) = RECIPIENT.call{value: value}(hex"");
       } else {
           // L1으로 브릿지하여 수취
           L2StandardBridge.bridgeNativeTokenTo{value: value}(RECIPIENT, ...);
       }
   }
   ```

### 2. **배처 운영비 (Batcher Costs)**

#### 비용 구조
배처는 **L1 가스비를 지불**하여 L2 데이터를 제출하므로 **직접적인 수익 메커니즘이 없습니다**:

```go
// op-batcher/batcher/driver.go
func (l *BatchSubmitter) sendTransaction(ctx context.Context, txdata txData, ...) error {
    // L1에 트랜잭션 제출 - 가스비 지불 필요
    queue.Send(txdata.ID(), *candidate, receiptsCh)
    return nil
}
```

#### 보상 메커니즘
- **L1FeeVault**: L2 사용자들이 지불하는 L1 데이터 가용성 비용이 L1FeeVault (`0x420000000000000000000000000000000000001a`)에 누적
- **운영자 보상**: L1FeeVault에서 배처 운영자에게 보상 지급 (별도 설정 필요)

### 3. **프로포저 운영비 (Proposer Costs)**

#### 비용 구조
프로포저도 **L1 가스비를 지불**하여 상태 루트를 제출:

```go
// op-proposer/proposer/driver.go
func (l *L2OutputSubmitter) sendTransaction(ctx context.Context, output *eth.OutputResponse) error {
    if l.Cfg.DisputeGameFactoryAddr != nil {
        // DisputeGameFactory에 제출 시 본드도 필요
        receipt, err = l.Txmgr.Send(ctx, txmgr.TxCandidate{
            TxData:   data,
            To:       l.Cfg.DisputeGameFactoryAddr,
            Value:    bond,  // 본드 예치
        })
    } else {
        // L2OutputOracle에 제출
        receipt, err = l.Txmgr.Send(ctx, txmgr.TxCandidate{
            TxData:   data,
            To:       l.Cfg.L2OutputOracleAddr,
        })
    }
}
```

#### 보상 메커니즘
- **직접 보상 없음**: 프로포저는 일반적으로 직접적인 온체인 보상이 없음
- **Dispute Game 승리 시**: Fault Proof 시스템에서 올바른 제안 시 상대방 본드 획득 가능

### 4. **Fee Vault 시스템**

#### 3가지 Fee Vault
```go
// op-service/predeploys/addresses.go
const (
    SequencerFeeVault = "0x4200000000000000000000000000000000000011"  // 시퀀서 수수료
    BaseFeeVault      = "0x4200000000000000000000000000000000000019"  // Base Fee
    L1FeeVault        = "0x420000000000000000000000000000000000001a"  // L1 DA 비용
)
```

#### 수수료 분배 구조
1. **SequencerFeeVault**: L2 트랜잭션의 Priority Fee → 시퀀서
2. **BaseFeeVault**: L2 트랜잭션의 Base Fee → 프로토콜/DAO
3. **L1FeeVault**: L1 데이터 가용성 비용 → 배처 보상

### 5. **운영비 요약**

| 컴포넌트 | 수익원 | 비용 | 보상 메커니즘 |
|---------|-------|------|-------------|
| **시퀀서** | L2 트랜잭션 수수료 | 서버 운영비 | SequencerFeeVault에서 자동 수취 |
| **배처** | 없음 | L1 가스비 | L1FeeVault에서 보상 (설정 필요) |
| **프로포저** | 없음 | L1 가스비 + 본드 | 일반적으로 없음 (Dispute 승리 시 본드) |

### 6. **결론**

- **시퀀서만** 직접적인 수익 메커니즘 보유
- **배처와 프로포저**는 비용만 발생하고 별도 보상 체계 필요
- **L2 운영자**는 일반적으로 시퀀서 수익으로 전체 인프라 운영비 충당
- **Fee Vault 시스템**을 통해 수수료가 자동으로 분배되고 누적됨
