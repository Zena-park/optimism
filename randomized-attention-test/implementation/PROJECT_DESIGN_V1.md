# RAT (Randomized Attention Test) 프로젝트 설계 문서

## 🎯 프로젝트 개요

### 목표
Optimistic Rollups의 **Verifier's Dilemma**를 해결하는 실제 동작하는 RAT 시스템을 구현합니다.

### 핵심 기능
- **확률적 Attention Test**: 검증자를 무작위로 선택하여 주기적으로 테스트
- **사전 검증**: 사기 발생 전에 검증자의 활동성 확인
- **경제적 인센티브**: 실패 시 직접적인 경제적 페널티
- **실시간 모니터링**: 검증자의 준비 상태 지속적 확인

## 🏗️ 전체 아키텍처

### 시스템 구성도
```
┌─────────────────────────────────────────────────────────────────┐
│                        RAT 시스템 아키텍처                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐ │
│  │   L2 Rollup     │    │   L1 Ethereum   │    │   Validators    │ │
│  │                 │    │                 │    │                 │ │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │ │
│  │ │  Proposer   │ │    │ │ RAT Contract│ │    │ │ Validator 1 │ │ │
│  │ │             │ │    │ │             │ │    │ │             │ │ │
│  │ │ • State     │ │────┼─│ • Test      │ │────┼─│ • Monitor   │ │ │
│  │ │   Compute   │ │    │ │   Trigger   │ │    │ │ • Verify    │ │ │
│  │ │ • Submit    │ │    │ │ • Validator │ │    │ │ • Respond   │ │ │
│  │ │   Root      │ │    │ │   Select    │ │    │ │ • Solution  │ │ │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │ │
│  │                 │    │                 │    │                 │ │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │ │
│  │ │  State      │ │    │ │  Dispute    │ │    │ │ Validator N │ │ │
│  │ │  Tree       │ │    │ │  Resolution │ │    │ │             │ │ │
│  │ │             │ │    │ │             │ │    │ │ • Monitor   │ │ │
│  │ │ • Merkle    │ │    │ │ • Fraud     │ │    │ │ • Verify    │ │ │
│  │ │   Root      │ │    │ │   Proof     │ │    │ │ • Respond   │ │ │
│  │ │ • Children  │ │    │ │ • Penalty   │ │    │ │ • Solution  │ │ │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │ │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 데이터 흐름
```
1. Proposer가 L2 상태 전환 계산
   ↓
2. L1 스마트 컨트랙트에 상태 커밋먼트 제출
   ↓
3. 확률 πa로 Attention Test 트리거 결정
   ↓
4. 무작위 검증자 선택
   ↓
5. Attention Test 이벤트 발생
   ↓
6. 검증자가 상태 전환 재계산
   ↓
7. 검증자가 솔루션 제출
   ↓
8. 성공/실패 판정 및 페널티 적용
```

## 📁 프로젝트 구조

### 디렉토리 구조
```
implementation/
├── contracts/                    # 스마트 컨트랙트
│   └── RandomizedAttentionTest.sol
├── test/                        # 테스트 파일
│   ├── RAT.test.js
│   ├── ValidatorClient.test.js
│   └── ProposerClient.test.js
├── scripts/                     # 배포 및 실행 스크립트
│   ├── deploy.js
│   ├── validator-simulation.js
│   └── proposer-simulation.js
├── client/                      # 클라이언트 구현
│   ├── ValidatorClient.js
│   ├── ProposerClient.js
│   └── RATClient.js
├── utils/                       # 유틸리티 함수
│   ├── merkleTree.js
│   ├── stateComputation.js
│   └── randomness.js
├── config/                      # 설정 파일
│   ├── parameters.js
│   └── networks.js
├── docs/                        # 문서
│   ├── API.md
│   └── DEPLOYMENT.md
├── hardhat.config.js            # Hardhat 설정
├── package.json                 # 의존성 관리
└── PROJECT_DESIGN.md           # 이 문서
```

## 🔧 핵심 컴포넌트 설계

### 1. 스마트 컨트랙트 (RandomizedAttentionTest.sol)

#### 주요 구조체
```solidity
struct Validator {
    address validator;           // 검증자 주소
    uint256 deposit;            // 예치금
    bool isActive;              // 활성 상태
    uint256 lastTestTime;       // 마지막 테스트 시간
    uint256 totalTests;         // 총 테스트 수
    uint256 passedTests;        // 통과한 테스트 수
    uint256 failedTests;        // 실패한 테스트 수
}

struct AttentionTest {
    bytes32 testId;             // 테스트 ID
    bytes32 stateRoot;          // 상태 루트
    uint256 l2BlockNumber;      // L2 블록 번호
    address targetValidator;    // 대상 검증자
    uint256 startTime;          // 시작 시간
    uint256 responseWindow;     // 응답 윈도우
    bool completed;             // 완료 여부
    bool passed;                // 통과 여부
    bytes32 leftChild;          // 왼쪽 자식 해시
    bytes32 rightChild;         // 오른쪽 자식 해시
}
```

#### 핵심 함수들
- `submitStateCommitment()`: 상태 커밋먼트 제출 및 Attention Test 트리거
- `submitAttentionSolution()`: Attention Test 솔루션 제출
- `registerValidator()`: 검증자 등록
- `unregisterValidator()`: 검증자 해제

### 2. 검증자 클라이언트 (ValidatorClient.js)

#### 주요 기능
```javascript
class ValidatorClient {
    // 이벤트 모니터링
    async startMonitoring()

    // Attention Test 처리
    async handleAttentionTest(testId, stateRoot)

    // 상태 전환 계산
    async computeStateTransition(l2BlockNumber, expectedStateRoot)

    // 솔루션 제출
    async submitSolution(testId, leftChild, rightChild)
}
```

### 3. 제안자 클라이언트 (ProposerClient.js)

#### 주요 기능
```javascript
class ProposerClient {
    // 상태 커밋먼트 제출
    async submitStateCommitment(stateRoot, l2BlockNumber)

    // 이벤트 모니터링
    async monitorEvents()

    // 상태 전환 계산
    async computeStateTransition(transactions)
}
```

## ⚙️ 핵심 메커니즘

### 1. Attention Test 트리거 메커니즘

#### 확률적 결정
```solidity
function shouldTriggerAttentionTest(bytes32 stateRoot) internal view returns (bool) {
    bytes32 randomness = keccak256(
        abi.encodePacked(
            stateRoot,
            blockhash(block.number - 1),
            block.timestamp
        )
    );

    uint256 threshold = (2**256 - 1) * attentionTestProbability / 10000;
    return uint256(randomness) <= threshold;
}
```

#### 검증자 선택
```solidity
function selectTargetValidator(bytes32 stateRoot) internal view returns (address) {
    bytes32 seed = keccak256(
        abi.encodePacked(
            stateRoot,
            blockhash(block.number - 1),
            block.timestamp
        )
    );

    uint256 selectedIndex = uint256(seed) % validatorList.length;
    return validatorList[selectedIndex];
}
```

### 2. Attention Puzzle 메커니즘

#### 솔루션 검증
```solidity
function verifyAttentionSolution(
    bytes32 leftChild,
    bytes32 rightChild,
    bytes32 expectedStateRoot
) internal pure returns (bool) {
    bytes32 computedRoot = hashTree(leftChild, rightChild);
    return computedRoot == expectedStateRoot;
}
```

#### 해시 트리 함수
```solidity
function hashTree(bytes32 left, bytes32 right) internal pure returns (bytes32) {
    return keccak256(abi.encodePacked(left, right));
}
```

### 3. 페널티 메커니즘

#### 페널티 적용
```solidity
function applyPenalty(address validator) internal {
    uint256 penalty = Math.min(penaltyAmount, validators[validator].deposit);

    if (penalty > 0) {
        validators[validator].deposit -= penalty;
        (bool success, ) = owner().call{value: penalty}("");
        require(success, "Penalty transfer failed");

        emit PenaltyApplied(validator, penalty);
    }
}
```

## 📊 매개변수 설계

### 기본 매개변수
```javascript
const RAT_PARAMETERS = {
    attentionTestProbability: 28,    // 0.28% (basis points)
    responseWindow: 300,             // 5분 (초)
    minDeposit: ethers.utils.parseEther("1"),    // 1 ETH
    penaltyAmount: ethers.utils.parseEther("0.1") // 0.1 ETH
};
```

### 매개변수 계산 근거
- **attentionTestProbability**: 논문 분석 결과 0.28%가 최적
- **responseWindow**: 5분은 충분한 응답 시간
- **minDeposit**: 검증자의 진지한 참여 보장
- **penaltyAmount**: 검증자 예치금의 10% 수준

## 🔄 프로토콜 흐름

### 1. 초기 설정
```
1. 컨트랙트 배포
2. 검증자 등록 (예치금과 함께)
3. 매개변수 설정
4. 모니터링 시작
```

### 2. 정상 운영
```
1. Proposer가 L2 상태 전환 계산
2. 상태 커밋먼트 제출
3. 확률적으로 Attention Test 트리거
4. 무작위 검증자 선택
5. 검증자가 상태 재계산
6. 솔루션 제출 및 검증
7. 결과에 따른 페널티/보상
```

### 3. 예외 처리
```
- 검증자 응답 시간 초과 → 자동 페널티
- 잘못된 솔루션 → 페널티 적용
- 검증자 등록 해제 → 예치금 반환
- 매개변수 조정 → 관리자 권한으로 수정
```

## 🧪 테스트 전략

### 1. 단위 테스트
- 스마트 컨트랙트 함수별 테스트
- 검증자 클라이언트 기능 테스트
- 제안자 클라이언트 기능 테스트

### 2. 통합 테스트
- 전체 프로토콜 흐름 테스트
- 다중 검증자 시나리오 테스트
- 페널티 메커니즘 테스트

### 3. 시뮬레이션 테스트
- 대규모 검증자 네트워크 시뮬레이션
- 다양한 공격 시나리오 테스트
- 성능 및 확장성 테스트

## 🚀 배포 전략

### 1. 개발 환경
- Hardhat 네트워크에서 로컬 테스트
- Ganache를 사용한 개인 네트워크 테스트

### 2. 테스트넷 배포
- Sepolia 테스트넷에서 검증
- 실제 가스비와 네트워크 조건에서 테스트

### 3. 메인넷 배포
- 점진적 배포 (소규모 → 대규모)
- 모니터링 및 매개변수 조정

## 📈 모니터링 및 분석

### 1. 실시간 모니터링
- Attention Test 트리거 빈도
- 검증자 응답 시간
- 성공/실패 비율
- 페널티 적용 현황

### 2. 성능 분석
- 가스 비용 분석
- 처리량 영향 측정
- 확장성 테스트 결과

### 3. 보안 분석
- 공격 시나리오 시뮬레이션
- 경제적 인센티브 분석
- 랜덤니스 조작 저항성 테스트

## 🔮 향후 발전 방향

### 1. 단기 개선
- 더 정교한 랜덤니스 생성
- 동적 매개변수 조정
- 개선된 이벤트 시스템

### 2. 중기 확장
- 다중 체인 지원
- 크로스체인 검증
- AI 기반 최적화

### 3. 장기 비전
- 완전 탈중앙화된 RAT 네트워크
- 양자 내성 암호화
- 블록체인 생태계 전체 적용

## 💡 구현 우선순위

### Phase 1: 기본 구현 (현재)
- [x] 스마트 컨트랙트 구현
- [ ] 기본 클라이언트 구현
- [ ] 단위 테스트 작성
- [ ] 로컬 환경 테스트

### Phase 2: 고급 기능
- [ ] 다중 검증자 시뮬레이션
- [ ] 성능 최적화
- [ ] 보안 강화
- [ ] 테스트넷 배포

### Phase 3: 프로덕션 준비
- [ ] 대규모 테스트
- [ ] 감사 및 검증
- [ ] 메인넷 배포
- [ ] 모니터링 시스템 구축

이 설계 문서는 RAT 프로젝트의 전체적인 구조와 구현 방향을 제시하며, 단계별로 발전시켜 나갈 수 있는 로드맵을 제공합니다.
