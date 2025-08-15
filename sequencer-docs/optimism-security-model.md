# Optimism 보안 모델: 배치 제출 vs Dispute Game

## 개요

Optimism에서는 두 가지 다른 L1 제출이 존재합니다: **Sequencer의 배치 제출**과 **Proposer의 Dispute Game 제출**. 이 문서는 왜 두 가지 제출이 모두 필요한지, 그리고 각각의 역할과 보안상의 중요성을 설명합니다.

## 1. 두 가지 제출 시스템

### 1.1 Sequencer (op-batcher) - 배치 제출

#### 목적
- L2 트랜잭션들을 L1에 압축하여 제출
- 데이터 가용성(Data Availability) 보장

#### 제출 내용
```go
// op-batcher가 L2 트랜잭션들을 L1에 제출
func (b *Batcher) loop() {
    // L2 트랜잭션들을 수집하여 배치로 압축
    // BatchInbox에 배치 데이터 제출
}
```

#### 특징
- **데이터 전송**: 실제 트랜잭션 데이터를 L1에 저장
- **압축**: 여러 트랜잭션을 하나의 배치로 압축하여 가스 비용 절약
- **가용성**: L2 데이터가 L1에서 접근 가능함을 보장

### 1.2 Proposer (op-proposer) - Dispute Game 제출

#### 목적
- L2 상태의 무결성 검증을 위한 Fault Proof 게임 생성
- 상태 무결성(State Integrity) 보장

#### 제출 내용
```go
// op-proposer가 L2 상태 루트를 L1에 제출
func (l *L2OutputSubmitter) sendTransaction(ctx context.Context, output source.Proposal) error {
    if l.Cfg.DisputeGameFactoryAddr != nil {
        candidate, err := l.ProposeL2OutputDGFTxCandidate(ctx, output)
        // DisputeGameFactory에 output root 제출
    }
}
```

#### 특징
- **상태 검증**: L2에서 계산된 상태 루트를 L1에 제출
- **챌린지 가능**: 잘못된 상태에 대해 챌린지 가능
- **무결성**: L2 상태의 올바름을 보장

## 2. 왜 두 가지 제출이 모두 필요한가?

### 2.1 배치 제출만으로는 부족한 이유

#### 배치 제출의 한계
```go
// 배치 제출만 있는 경우의 문제점
type BatchOnlyLimitation struct {
    DataAvailability    bool   // ✅ 트랜잭션 데이터는 L1에 있음
    StateIntegrity      bool   // ❌ 상태가 올바른지 검증 불가
    CalculationError    bool   // ❌ 계산 오류 감지 불가
    MaliciousBehavior   bool   // ❌ 악의적 행위 감지 불가
}
```

#### 실제 시나리오: 배치만 있는 경우
```
1. Sequencer가 트랜잭션 배치를 L1에 제출 ✅
2. L2에서 잘못된 계산으로 상태 변경 ❌
3. 배치 데이터는 올바르지만 상태는 잘못됨
4. 누구도 이를 감지할 수 없음 ❌
5. 사용자 자산 손실 발생 ❌
```

### 2.2 Dispute Game의 필요성

#### Dispute Game의 역할
```go
// Dispute Game이 있는 경우의 보안
type DisputeGameSecurity struct {
    DataAvailability    bool   // ✅ 배치 제출로 보장
    StateIntegrity      bool   // ✅ Dispute Game으로 보장
    CalculationError    bool   // ✅ 챌린지로 감지 가능
    MaliciousBehavior   bool   // ✅ 챌린지로 감지 가능
}
```

#### 실제 시나리오: Dispute Game이 있는 경우
```
1. Sequencer가 트랜잭션 배치를 L1에 제출 ✅
2. L2에서 잘못된 계산으로 상태 변경 ❌
3. Proposer가 잘못된 상태 루트를 L1에 제출
4. op-challenger가 즉시 감지하고 챌린지 ✅
5. 잘못된 상태가 롤백됨 ✅
6. 사용자 자산 보호됨 ✅
```

## 3. 비용 vs 보안의 트레이드오프

### 3.1 비용 분석

#### 배치 제출 비용
```go
type BatchSubmissionCost struct {
    GasCost     uint64 // 트랜잭션 데이터 전송 가스 비용
    Frequency   string // "매 배치마다"
    Purpose     string // "데이터 가용성"
}
```

#### Dispute Game 제출 비용
```go
type DisputeGameCost struct {
    GasCost     uint64 // 상태 루트 검증 가스 비용
    BondCost    uint64 // 보증금 비용
    Frequency   string // "설정된 간격마다"
    Purpose     string // "상태 무결성 검증"
}
```

#### 총 비용
```go
type TotalCost struct {
    BatchSubmission    uint64 // 배치 제출 비용
    DisputeGame        uint64 // Dispute Game 제출 비용
    Total              uint64 // 두 비용의 합
    SecurityLevel      string // "완전한 보안 보장"
}
```

### 3.2 보안 수준 비교

#### 배치만 있는 경우
```go
type BatchOnlySecurity struct {
    DataAvailability    bool   // ✅ 보장됨
    StateIntegrity      bool   // ❌ 보장되지 않음
    FaultDetection      bool   // ❌ 불가능
    UserProtection      bool   // ❌ 제한적
    SecurityLevel       string // "부분적 보안"
}
```

#### Dispute Game 포함
```go
type FullSecurity struct {
    DataAvailability    bool   // ✅ 보장됨
    StateIntegrity      bool   // ✅ 보장됨
    FaultDetection      bool   // ✅ 가능함
    UserProtection      bool   // ✅ 완전함
    SecurityLevel       string // "완전한 보안"
}
```

## 4. Optimism의 보안 모델

### 4.1 전체 보안 아키텍처

```go
type OptimismSecurityModel struct {
    // 데이터 계층
    DataLayer struct {
        Sequencer    string // "배치 제출"
        Batcher      string // "트랜잭션 압축"
        L1Storage    string // "데이터 가용성"
    }

    // 검증 계층
    ValidationLayer struct {
        Proposer     string // "상태 루트 제출"
        Challenger   string // "잘못된 상태 챌린지"
        DisputeGame  string // "Fault Proof 게임"
    }

    // 보안 계층
    SecurityLayer struct {
        DataAvailability    bool   // ✅ 배치로 보장
        StateIntegrity      bool   // ✅ Dispute Game으로 보장
        FaultProof          bool   // ✅ 챌린지 시스템으로 보장
    }
}
```

### 4.2 데이터 흐름

```
L2 트랜잭션들
    ↓
Sequencer (op-batcher) → L1 BatchInbox에 배치 제출
    ↓
L2 블록 생성 및 상태 계산
    ↓
Proposer (op-proposer) → L1 DisputeGameFactory에 output root 제출
    ↓
op-challenger가 L1 헤드 구독으로 감지
    ↓
잘못된 output root에 대해 챌린지
    ↓
Fault Proof 게임 진행
    ↓
올바른 상태로 복구
```

## 5. 다른 L2 솔루션과의 비교

### 5.1 ZK-Rollup

#### 특징
```go
type ZKRollup struct {
    Submission      string // "증명(proof)만 제출"
    Finality        string // "즉시 최종성"
    Computation     string // "복잡한 증명 생성"
    SupportedOps    string // "특정 연산만 지원"
    Cost            string // "높은 증명 생성 비용"
}
```

#### 장단점
- **장점**: 즉시 최종성 보장, 높은 보안
- **단점**: 복잡한 증명 생성, 특정 연산만 지원, 높은 비용

### 5.2 Optimistic Rollup (Optimism)

#### 특징
```go
type OptimisticRollup struct {
    Submission      string // "배치 + 상태 루트 제출"
    Finality        string // "챌린지 기간 후 최종성"
    Computation     string // "간단한 구조"
    SupportedOps    string // "모든 연산 지원"
    Cost            string // "두 가지 제출 비용"
}
```

#### 장단점
- **장점**: 모든 연산 지원, 간단한 구조, 낮은 비용
- **단점**: 챌린지 기간 필요, 두 가지 제출 비용

## 6. 실제 위험 시나리오

### 6.1 시나리오 1: 악의적인 시퀀서

#### 상황
```
1. 악의적인 시퀀서가 L2에서 잘못된 계산 수행
2. 트랜잭션 배치는 올바르게 L1에 제출됨
3. 하지만 L2 상태는 잘못됨
```

#### 배치만 있는 경우
```
❌ 누구도 잘못된 상태를 감지할 수 없음
❌ 사용자 자산 손실 발생
❌ 복구 불가능
```

#### Dispute Game이 있는 경우
```
✅ op-challenger가 잘못된 상태 감지
✅ 즉시 챌린지 시작
✅ Fault Proof 게임으로 올바른 상태 증명
✅ 잘못된 상태 롤백
✅ 사용자 자산 보호
```

### 6.2 시나리오 2: 소프트웨어 버그

#### 상황
```
1. L2 노드 소프트웨어에 버그 발생
2. 특정 조건에서 잘못된 상태 계산
3. 배치 데이터는 정상이지만 상태는 오류
```

#### 배치만 있는 경우
```
❌ 버그로 인한 잘못된 상태 감지 불가
❌ 모든 사용자에게 영향
❌ 수동 개입 필요
```

#### Dispute Game이 있는 경우
```
✅ 자동으로 잘못된 상태 감지
✅ 즉시 챌린지 및 복구
✅ 사용자 영향 최소화
✅ 자동화된 복구 시스템
```

## 7. 결론

### 7.1 Dispute Game 제출이 필요한 핵심 이유

1. **상태 무결성 검증**: 배치 데이터만으로는 L2 상태의 올바름을 검증할 수 없음
2. **계산 오류 감지**: L2에서 발생할 수 있는 계산 오류를 감지하고 복구
3. **보안 보장**: 악의적인 행위에 대한 완전한 보호
4. **사용자 보호**: 잘못된 상태로 인한 사용자 자산 손실 방지
5. **자동화된 복구**: 수동 개입 없이 자동으로 시스템 복구

### 7.2 비용 vs 보안의 균형

```go
type CostSecurityBalance struct {
    HigherCost     bool   // ✅ 두 가지 제출 비용
    CompleteSecurity bool   // ✅ 완전한 보안 보장
    UserProtection  bool   // ✅ 사용자 자산 완전 보호
    Automation      bool   // ✅ 자동화된 복구 시스템
    Tradeoff        string // "비용은 높지만 보안은 완전"
}
```

### 7.3 Optimism의 선택

Optimism은 **"Optimistic Rollup"** 모델을 선택하여:

- **모든 연산 지원**: EVM 호환성으로 모든 스마트 컨트랙트 실행 가능
- **간단한 구조**: 복잡한 증명 생성 없이 간단한 구조
- **완전한 보안**: 두 가지 제출을 통한 완전한 보안 보장
- **사용자 친화적**: 높은 처리량과 낮은 비용

**결론적으로, Dispute Game 제출은 비용이 높지만 Optimism의 보안을 완전히 보장하는 필수적인 요소입니다.**
