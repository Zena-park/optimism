# Optimism 배치 제출 시스템 분석

## 목차
1. [시스템 개요](#시스템-개요)
2. [배치 제출 아키텍처](#배치-제출-아키텍처)
3. [BatchInbox 컨트랙트 분석](#batchinbox-컨트랙트-분석)
4. [op-node 검증 프로세스](#op-node-검증-프로세스)
5. [보안 메커니즘](#보안-메커니즘)
6. [질문과 답변](#질문과-답변)
7. [기술적 세부사항](#기술적-세부사항)

## 시스템 개요

### 배치 제출 시스템의 목적
Optimism의 배치 제출 시스템은 L2 블록들을 L1에 효율적으로 전송하여 L2 상태를 L1에 기록하는 핵심 메커니즘입니다.

### 주요 구성 요소
- **배처(Batcher)**: L2 블록들을 수집하여 L1로 전송
- **BatchInbox**: L1에서 배치 데이터를 받는 주소
- **op-node**: L1의 배치 데이터를 읽어서 L2 블록 재생성
- **op-geth**: L2 블록 실행 및 상태 관리

## 배치 제출 아키텍처

### 전체 데이터 흐름
```
L2 거래들 → L2 블록 생성 (op-geth)
    ↓
L2 블록들 → 배처가 수집
    ↓
배처 → L1 BatchInbox 주소로 전송
    ↓
op-node (L2) → L1에서 배치 데이터 읽기
    ↓
op-node (L2) → 배치 데이터로 L2 블록 재생성
```

### 배처의 역할
```go
// op-batcher/batcher/driver.go:806-850
func (l *BatchSubmitter) publishTxToL1(ctx context.Context, queue *txmgr.Queue[txRef], receiptsCh chan txmgr.TxReceipt[txRef], daGroup *errgroup.Group) error {
    // 1. L2 블록들을 수집
    // 2. 배치 데이터 생성
    // 3. BatchInbox 주소로 전송 (배처 서명으로)
}
```

## BatchInbox 컨트랙트 분석

### BatchInbox의 정체
**중요**: BatchInbox는 실제 스마트 컨트랙트가 아니라 **계산된 주소**입니다.

### 주소 계산 방식
```solidity
// packages/contracts-bedrock/src/L1/OPContractsManager.sol:103-108
function chainIdToBatchInboxAddress(uint256 _l2ChainId) public pure returns (address) {
    bytes1 versionByte = 0x00;
    bytes32 hashedChainId = keccak256(bytes.concat(bytes32(_l2ChainId)));
    bytes19 first19Bytes = bytes19(hashedChainId);
    return address(uint160(bytes20(bytes.concat(versionByte, first19Bytes))));
}
```

### 주소 계산 공식
```
BatchInbox 주소 = versionByte || keccak256(bytes32(chainId))[:19]
```
- **versionByte**: `0x00`
- **chainId**: L2 체인 ID
- **예시**: Chain ID 901의 경우 → `0xff00000000000000000000000000000000000901`

### 전송되는 데이터 구조

#### Calldata 트랜잭션
```go
func (td *txData) CallData() []byte {
    data := make([]byte, 1, 1+td.Len())
    data[0] = params.DerivationVersion0  // 버전 바이트 (0x00)
    for _, f := range td.frames {
        data = append(data, f.data...)   // 프레임 데이터 연결
    }
    return data
}
```

#### Blob 트랜잭션
```go
func (td *txData) Blobs() ([]*eth.Blob, error) {
    blobs := make([]*eth.Blob, 0, len(td.frames))
    for _, f := range td.frames {
        var blob eth.Blob
        if err := blob.FromData(append([]byte{params.DerivationVersion0}, f.data...)); err != nil {
            return nil, err
        }
        blobs = append(blobs, &blob)
    }
    return blobs, nil
}
```

## op-node 검증 프로세스

### op-node의 역할
**중요**: op-node는 **L2 노드**이며, L1의 배치 데이터를 읽어서 L2 블록을 생성하는 역할을 합니다.

### 기본 검증 (`isValidBatchTx`)
```go
// op-node/rollup/derive/data_source.go:96-118
func isValidBatchTx(tx *types.Transaction, l1Signer types.Signer, batchInboxAddr, batcherAddr common.Address, logger log.Logger) bool {
    // 1. 트랜잭션 타입 검증
    if tx.Type() > types.BlobTxType && tx.Type() != types.DepositTxType {
        return false
    }

    // 2. 대상 주소 검증 (BatchInbox 주소)
    to := tx.To()
    if to == nil || *to != batchInboxAddr {
        return false
    }

    // 3. 배처 서명 검증
    seqDataSubmitter, err := l1Signer.Sender(tx)
    if err != nil || seqDataSubmitter != batcherAddr {
        logger.Warn("tx in inbox with unauthorized submitter", "addr", seqDataSubmitter, "hash", tx.Hash())
        return false
    }

    return true
}
```

### 데이터 추출 및 검증

#### Calldata 트랜잭션 처리
```go
// op-node/rollup/derive/calldata_source.go:67-87
func DataFromEVMTransactions(dsCfg DataSourceConfig, batcherAddr common.Address, txs types.Transactions, log log.Logger) []eth.Data {
    out := []eth.Data{}
    for _, tx := range txs {
        if isValidBatchTx(tx, dsCfg.l1Signer, dsCfg.batchInboxAddress, batcherAddr, log) {
            out = append(out, tx.Data())  // calldata 추출
        }
    }
    return out
}
```

#### Blob 트랜잭션 처리
```go
// op-node/rollup/derive/blob_data_source.go:120-154
func dataAndHashesFromTxs(txs types.Transactions, config *DataSourceConfig, batcherAddr common.Address, logger log.Logger) ([]blobOrCalldata, []eth.IndexedBlobHash) {
    for _, tx := range txs {
        if !isValidBatchTx(tx, config.l1Signer, config.batchInboxAddress, batcherAddr, logger) {
            continue
        }

        if tx.Type() != types.BlobTxType {
            // calldata 추출
            calldata := eth.Data(tx.Data())
            data = append(data, blobOrCalldata{nil, &calldata})
        } else {
            // blob 해시 추출
            for _, h := range tx.BlobHashes() {
                hashes = append(hashes, eth.IndexedBlobHash{...})
            }
        }
    }
}
```

### 프레임 파싱 및 검증
```go
// op-node/rollup/derive/frame.go:120-154
func ParseFrames(data []byte) ([]Frame, error) {
    // 1. 버전 바이트 검증
    if data[0] != params.DerivationVersion0 {
        return nil, fmt.Errorf("invalid derivation format byte: got %d", data[0])
    }

    buf := bytes.NewBuffer(data[1:])
    var frames []Frame

    // 2. 각 프레임 파싱
    for buf.Len() > 0 {
        var f Frame
        if err := f.UnmarshalBinary(buf); err != nil {
            return nil, fmt.Errorf("parsing frame %d: %w", len(frames), err)
        }
        frames = append(frames, f)
    }

    // 3. 완전한 소비 확인
    if buf.Len() != 0 {
        return nil, fmt.Errorf("did not fully consume data")
    }

    return frames, nil
}
```

### 프레임 구조 검증
```go
// op-node/rollup/derive/frame.go:60-95
func (f *Frame) UnmarshalBinary(r ByteReader) error {
    // 1. Channel ID (16 bytes)
    if _, err := io.ReadFull(r, f.ID[:]); err != nil {
        return fmt.Errorf("reading channel_id: %w", err)
    }

    // 2. Frame Number (2 bytes)
    if err := binary.Read(r, binary.BigEndian, &f.FrameNumber); err != nil {
        return fmt.Errorf("reading frame_number: %w", err)
    }

    // 3. Frame Data Length (4 bytes) + 크기 제한 검증
    var frameLength uint32
    if err := binary.Read(r, binary.BigEndian, &frameLength); err != nil {
        return fmt.Errorf("reading frame_data_length: %w", err)
    }
    if frameLength > MaxFrameLen { // 1MB 제한
        return fmt.Errorf("frame_data_length is too large: %d", frameLength)
    }

    // 4. Frame Data
    f.Data = make([]byte, int(frameLength))
    if _, err := io.ReadFull(r, f.Data); err != nil {
        return fmt.Errorf("reading frame_data: %w", err)
    }

    // 5. Is Last Flag (1 byte)
    if isLastByte, err := r.ReadByte(); err != nil {
        return fmt.Errorf("reading final byte (is_last): %w", err)
    } else if isLastByte == 0 {
        f.IsLast = false
    } else if isLastByte == 1 {
        f.IsLast = true
    } else {
        return errors.New("invalid byte as is_last")
    }

    return nil
}
```

## 보안 메커니즘

### 배처 서명 검증의 중요성

#### 문제 상황
누구나 BatchInbox 주소로 트랜잭션을 보낼 수 있습니다!

```solidity
// 악의적인 사용자가 BatchInbox 주소로 트랜잭션을 보낼 수 있음
maliciousTx = {
    to: "0xff00000000000000000000000000000000000901", // BatchInbox 주소
    from: "0x1234567890abcdef...", // 악의적 사용자
    data: "0x1234567890abcdef..." // 가짜 배치 데이터
}
```

#### 보안 위험
**만약 서명 검증이 없다면:**
- ❌ 악의적 사용자가 가짜 L2 블록 데이터를 L1에 전송
- ❌ op-node가 가짜 데이터를 L2 블록으로 처리
- ❌ L2 체인이 손상됨

#### 보안 효과
**서명 검증으로 보장되는 것:**
- ✅ **인증된 배처만** 배치 데이터 전송 가능
- ✅ **가짜 데이터 차단** - 권한 없는 사용자 차단
- ✅ **L2 체인 무결성** 보장

### 올바른 보안 시나리오
```
✅ 배처 → L1 BatchInbox → op-node(L2) → 정상 L2 블록
❌ 악의적 사용자 → L1 BatchInbox → op-node(L2) → 거부됨 (서명 검증)
```

## 질문과 답변

### Q1: BatchInbox 컨트랙트의 정확한 컨트랙트 이름이 뭐니? 컨트랙트 코드가 어디있어?

**A1**: BatchInbox는 별도의 Solidity 컨트랙트가 아닙니다. 단순한 **주소 계산 방식**으로 구현됩니다.

- **실제 구현**: `packages/contracts-bedrock/src/L1/OPContractsManager.sol`의 `chainIdToBatchInboxAddress` 함수
- **주소 계산**: `versionByte || keccak256(bytes32(chainId))[:19]`
- **관련 코드**:
  - `op-node/rollup/derive/data_source.go` - 배치 트랜잭션 검증
  - `op-batcher/batcher/driver.go` - 배치 제출 로직

### Q2: publishTxToL1에서 만들어진 트랜잭션 데이터는 어느 컨트랙트에 어떤 데이터를 전송하는 것인지?

**A2**:
- **대상**: BatchInbox 주소 (`l.RollupConfig.BatchInboxAddress`)
- **데이터**:
  - **Calldata**: 버전 바이트(0x00) + 프레임 데이터 연결
  - **Blob**: 버전 바이트(0x00) + 프레임 데이터를 Blob으로 변환
- **목적**: L2 블록 데이터를 L1에 저장하여 L2 상태를 L1에 기록

### Q3: op-node가 해당 주소의 트랜잭션을 검증한다는데 무엇을 검증하는 거지?

**A3**: op-node는 다음을 검증합니다:

1. **트랜잭션 타입**: Legacy, ACL, DynamicFee, Blob, Deposit만 허용
2. **대상 주소**: BatchInbox 주소와 일치
3. **서명 검증**: 배처 주소에서 전송된 트랜잭션만 허용
4. **프레임 구조**: ChannelID(16) + FrameNumber(2) + Length(4) + Data + IsLast(1)
5. **크기 제한**: 프레임당 최대 1MB
6. **완전성**: 모든 데이터가 소비되었는지 확인

### Q4: 유효한 배처 서명을 가진 트랜잭션만 처리한다는 의미가 뭐지? 이것은 L2에서 L1으로 보낸 배치인데. 이것으로 무엇을 처리한다는 거지?

**A4**:

**"유효한 배처 서명을 가진 트랜잭션만 처리한다"는 의미:**

1. **보안**: 권한 없는 사용자의 가짜 데이터 차단
2. **신뢰성**: 공식 배처가 전송한 데이터만 신뢰
3. **동기화**: L1의 배치 데이터를 기반으로 L2 블록 재생성
4. **무결성**: L2 체인의 상태 무결성 보장

**"처리한다"는 의미:**
- **L2 블록 재생성**: 검증된 배치 데이터로 L2 블록 생성
- **L2 상태 동기화**: 모든 L2 노드가 동일한 L2 상태를 가지도록
- **보안 모델**: L1(신뢰할 수 있는 레이어) → L2(실행 레이어)

### Q5: op-node가 가짜 데이터를 L2 블록으로 처리한다는 것은 L1의 op-node를 말하는 것인가?

**A5**: 아니요, **op-node는 L2 노드**입니다!

**정확한 아키텍처:**
```
L1 (Ethereum Mainnet)
├── BatchInbox 주소: 배치 데이터 저장소
└── 배처(Batcher): L2 블록들을 L1로 전송

L2 (Optimism)
├── op-node: L1의 배치 데이터를 읽어서 L2 블록 생성
├── op-geth: L2 블록 실행 및 상태 관리
└── 기타 L2 노드들
```

**op-node의 역할:**
1. **L1 모니터링**: L1의 BatchInbox 주소를 지속적으로 모니터링
2. **배치 데이터 읽기**: L1에서 배치 데이터를 읽어옴
3. **서명 검증**: 배처 서명을 검증하여 권한 있는 데이터만 처리
4. **L2 블록 생성**: 검증된 배치 데이터로 L2 블록을 생성/동기화

## 기술적 세부사항

### 검증 요약

#### 1단계: 트랜잭션 검증
- ✅ **트랜잭션 타입**: Legacy, ACL, DynamicFee, Blob, Deposit만 허용
- ✅ **대상 주소**: BatchInbox 주소와 일치
- ✅ **서명 검증**: 배처 주소에서 전송된 트랜잭션만 허용

#### 2단계: 데이터 추출
- ✅ **Calldata**: 트랜잭션 데이터 직접 추출
- ✅ **Blob**: Blob 해시를 통해 실제 데이터 다운로드

#### 3단계: 프레임 파싱
- ✅ **버전 검증**: DerivationVersion0 (0x00)
- ✅ **프레임 구조**: ChannelID(16) + FrameNumber(2) + Length(4) + Data + IsLast(1)
- ✅ **크기 제한**: 프레임당 최대 1MB
- ✅ **완전성**: 모든 데이터가 소비되었는지 확인

#### 4단계: 채널 어셈블리
- ✅ **순서 검증**: FrameNumber 순서대로
- ✅ **채널 크기**: 최대 RLP 바이트 제한
- ✅ **완성도**: IsLast 플래그로 채널 완성 확인

### 보안의 중요성
- **L1은 신뢰할 수 있지만**: 누구나 BatchInbox 주소로 트랜잭션을 보낼 수 있음
- **op-node(L2)가 보호해야 함**: 배처 서명 검증으로 가짜 데이터 차단
- **결과**: L2 체인의 무결성과 신뢰성 보장

이러한 다단계 검증을 통해 **안전하고 신뢰할 수 있는 L2 블록 생성**을 보장합니다.
