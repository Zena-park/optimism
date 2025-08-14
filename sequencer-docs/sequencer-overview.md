# Optimism 시퀀서 개요

## 목차
1. [시퀀서의 역할](#시퀀서의-역할)
2. [주요 코드 위치](#주요-코드-위치)
3. [거래 순서 결정 규칙](#거래-순서-결정-규칙)
4. [시퀀서 아키텍처](#시퀀서-아키텍처)
5. [거래 순서 보존 제약사항](#거래-순서-보존-제약사항)

## 시퀀서의 역할

Optimism L2 체인에서 시퀀서는 다음과 같은 핵심 역할을 담당합니다:

### 1. 블록 생성 및 순서 결정
- L2 블록을 생성하고 거래의 순서를 결정
- 정해진 블록 시간에 맞춰 블록 생성
- 거래 풀(mempool)에서 거래를 선택하여 블록에 포함

### 2. L1과의 동기화
- L1 블록을 참조하여 L2 블록 생성
- L1의 상태 변화를 L2에 반영
- L1과 L2 간의 일관성 유지

### 3. 거래 처리
- 사용자 거래를 수집하고 검증
- 거래 풀 관리 및 최적화
- 블록 가스 한계 내에서 거래 포함

### 4. 네트워크 안정성
- 블록 생성의 연속성 보장
- 네트워크 상태 모니터링
- 장애 상황에서의 복구 처리

## 주요 코드 위치

### 핵심 시퀀서 구현
```
op-node/rollup/sequencing/
├── sequencer.go          # 메인 시퀀서 구현체
├── iface.go             # 시퀀서 인터페이스 정의
└── sequencer_test.go    # 시퀀서 테스트
```

### 시퀀서 관련 컴포넌트
```
op-conductor/                    # 고가용성 시퀀서 관리
├── conductor/
│   ├── service.go              # 시퀀서 상태 관리
│   └── conductor.go            # 리더 선출 로직
├── client/
│   └── sequencer.go            # 시퀀서 제어 인터페이스
└── README.md                   # Conductor 문서

op-test-sequencer/              # 테스트용 시퀀서 구현
├── sequencer/
│   ├── backend/
│   │   └── work/
│   │       ├── iface.go        # 시퀀서 인터페이스
│   │       └── sequencers/
│   │           └── fullseq/    # 전체 시퀀서 구현
│   └── service.go              # 시퀀서 서비스
```

### 드라이버 통합
```
op-node/rollup/driver/
└── driver.go                   # 시퀀서를 드라이버에 통합
```

### 거래 처리 관련
```
op-service/
├── txmgr/                      # 거래 관리자
│   ├── txmgr.go               # 메인 거래 관리 로직
│   ├── queue.go               # 거래 큐 관리
│   └── send_state.go          # 거래 전송 상태 관리
└── txinclude/                 # 거래 포함 로직
    ├── persistent.go          # 지속적 거래 포함
    ├── resubmitter.go         # 거래 재전송
    └── nonce_manager.go       # Nonce 관리
```

## 거래 순서 결정 규칙

### 1. 기본 거래 순서 결정 방식

시퀀서는 **거래 풀(Transaction Pool)에서 거래를 선택하는 방식**으로 거래 순서를 결정합니다.

#### 핵심 설정: `NoTxPool` 플래그
```go
// op-node/rollup/sequencing/sequencer.go:538
attrs.NoTxPool = uint64(attrs.Timestamp) > l1Origin.Time+d.spec.MaxSequencerDrift(l1Origin.Time)
```

- **`NoTxPool = true`**: 거래 풀에서 거래를 가져오지 않음 (빈 블록 또는 예약된 거래만)
- **`NoTxPool = false`**: 거래 풀에서 거래를 가져와서 블록에 포함

### 2. 거래 순서 결정 규칙

#### A. 시퀀서 드리프트 제한
```go
// 시퀀서가 L1보다 너무 앞서가면 거래 풀 사용 안함
attrs.NoTxPool = uint64(attrs.Timestamp) > l1Origin.Time+d.spec.MaxSequencerDrift(l1Origin.Time)
```

#### B. 업그레이드 블록 처리
```go
// 특정 업그레이드 블록에서는 거래 풀 사용 안함
if d.rollupCfg.IsEcotoneActivationBlock(uint64(attrs.Timestamp)) {
    attrs.NoTxPool = true
}
```

#### C. 복구 모드
```go
if recoverMode {
    attrs.NoTxPool = true
    d.log.Warn("Sequencing temporarily without user transactions, in recover mode")
}
```

### 3. 거래 풀에서의 거래 선택

시퀀서가 `NoTxPool = false`일 때, **거래 풀의 기본 정렬 규칙**을 따릅니다:

#### Ethereum 거래 풀 정렬 규칙
1. **Gas Price 우선순위**: 높은 gas price를 가진 거래가 우선
2. **Nonce 순서**: 같은 계정 내에서는 nonce 순서대로
3. **거래 타입**: Blob 거래와 일반 거래는 별도 처리

## 시퀀서 아키텍처

### 1. 시퀀서 상태 관리
```go
type Sequencer struct {
    active atomic.Bool                    // 시퀀서 활성화 상태
    latest BuildingState                  // 현재 빌딩 상태
    latestHead eth.L2BlockRef            // 최신 L2 헤드
    rollupCfg *rollup.Config             // 롤업 설정
    attrBuilder derive.AttributesBuilder  // 속성 빌더
    l1OriginSelector L1OriginSelectorIface // L1 원본 선택자
    conductor conductor.SequencerConductor // Conductor 인터페이스
}
```

### 2. 블록 빌딩 프로세스
```go
func (d *Sequencer) startBuildingBlock() {
    // 1. L1 원본 블록 찾기
    l1Origin, err := d.l1OriginSelector.FindL1Origin(ctx, l2Head)

    // 2. 페이로드 속성 준비
    attrs, err := d.attrBuilder.PreparePayloadAttributes(fetchCtx, l2Head, l1Origin.ID())

    // 3. NoTxPool 설정 결정
    attrs.NoTxPool = uint64(attrs.Timestamp) > l1Origin.Time+d.spec.MaxSequencerDrift(l1Origin.Time)

    // 4. 블록 빌딩 시작
    d.emitter.Emit(d.ctx, engine.BuildStartEvent{
        Attributes: withParent,
    })
}
```

### 3. Conductor 시스템
Conductor는 고가용성 시퀀서 관리를 위한 시스템입니다:

- **리더 선출**: 여러 시퀀서 중 하나를 리더로 선출
- **상태 동기화**: 시퀀서 간 상태 동기화
- **장애 복구**: 리더 장애 시 자동 전환

## 거래 순서 보존 제약사항

### 1. 현재 제약
- **Nonce 순서**: 같은 계정의 거래는 nonce 순서대로만 처리
- **거래 타입 제한**: Blob 거래와 일반 거래는 별도 처리
- **Gas Limit**: 블록당 gas limit 제한

### 2. 부족한 제약
- **전체 거래 순서 보존**: 시퀀서가 거래 순서를 임의로 변경 가능
- **MEV 방지**: 시퀀서가 수익을 위해 거래 순서 조작 가능

### 3. 거래 순서 변경 가능한 경우
1. **MEV 추출**: 높은 수수료 거래를 우선 처리
2. **거래 검열**: 특정 거래를 지연시키거나 제외
3. **블록 공간 최적화**: 수익성 있는 거래만 선택

### 4. 거래 순서 보존 구현 방향

시퀀서가 거래 순서를 바꾸지 못하게 하려면:

1. **FIFO 큐 구현**: 거래가 도착한 순서대로 처리
2. **거래 순서 검증**: 블록에 포함된 거래 순서 검증
3. **MEV 방지 메커니즘**: 거래 순서 조작 방지
4. **거래 풀 정렬 규칙 수정**: Gas price 대신 도착 시간 기준 정렬

## 결론

현재 Optimism의 시퀀서는 **거래 풀의 기본 정렬 규칙을 따르므로, 시퀀서가 거래 순서를 자유롭게 변경할 수 있는 구조**입니다.

거래 순서 보존을 위해서는 시퀀서의 거래 선택 로직을 수정하여 FIFO(First-In-First-Out) 방식으로 거래를 처리하도록 변경해야 합니다. 이는 MEV 추출을 방지하고 공정한 거래 처리를 보장하는 데 중요한 역할을 합니다.

## 관련 파일들

### 핵심 시퀀서 파일
- `op-node/rollup/sequencing/sequencer.go` - 메인 시퀀서 구현
- `op-node/rollup/sequencing/iface.go` - 시퀀서 인터페이스
- `op-conductor/conductor/service.go` - Conductor 서비스
- `op-conductor/client/sequencer.go` - 시퀀서 제어 클라이언트

### 거래 처리 파일
- `op-service/txmgr/txmgr.go` - 거래 관리자
- `op-service/txinclude/persistent.go` - 거래 포함 로직
- `op-node/rollup/derive/attributes.go` - 페이로드 속성 생성

### 설정 및 문서
- `op-conductor/README.md` - Conductor 문서
- `op-node/rollup/driver/driver.go` - 드라이버 통합
- `op-test-sequencer/` - 테스트용 시퀀서
