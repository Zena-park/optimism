# RAT (Randomized Attention Test) 통합 계획

## 🔄 서비스 플로우

RAT 시스템의 전체 서비스 플로우는 다음과 같습니다:

1. **챌린저 네트워크 연결**: 챌린저들이 P2P 네트워크에 참여하여 서로를 발견하고 연결
2. **Dispute Game 모니터링 및 검증**: L1 체인을 모니터링하여 새로운 Dispute Game 발생 감지, Dispute Game 발생 시 Enhanced Verifier를 통한 검증 수행
3. **Attention Test 트리거**: Dispute Game이 발생하면 개인화된 비율에 따라 Attention Test 실행, 평판관리
4. **게임 참여**: 검증 결과가 틀렸을 경우 Dispute Game에 참여

```mermaid
flowchart TD
    A[챌린저 시작<br/>트리거 비율: 10%] --> B[P2P 네트워크 연결<br/>Discovery Service]
    B --> C[L1 체인 모니터링]
            C --> D[Dispute Game 생성 감지<br/>(프로포저가 생성)]

    D --> E{Dispute Game 발생?}
    E -->|Yes| F[동시 실행 시작]
    E -->|No| C

    F --> G[직접 검증<br/>Enhanced Verifier]
    F --> H[Attention Test<br/>트리거 판단]

    G --> I[상태 루트 검증<br/>배치 체커<br/>L1/L2 일치성 확인]
    I --> J[검증 결과]
    J --> K{검증 결과}
    K -->|이상 감지| L[Dispute Game 참여<br/>추가 보증금 지불]
    K -->|정상| M[정상 - 게임 참여 안함]

    H --> N[배치 번호 % 100<br/>< 트리거 비율?]
    N -->|Yes| O[챌린저 선택<br/>개인별 설정에 따라]
    N -->|No| P[트리거 안함]
    O --> Q[Attention 질문 전송<br/>다른 챌린저에게]

    Q --> R[Attention 응답 대기]
    R --> S[응답 수신]
    S --> T[응답 검증]
    T --> U[평판 업데이트<br/>질문자/응답자]

    L --> V[평판 관리<br/>양방향 평판 시스템]
    M --> V
    U --> V
    P --> V

    V --> DD[게임 최종화 대기<br/>(airgap period)]
    DD --> EE{게임 종료}
    EE -->|최종화 완료| FF[보증금 분배 모드 결정]
    EE -->|미완료| DD

    FF --> GG{게임 상태}
    GG -->|Proper Game| HH[NORMAL 모드<br/>게임 결과에 따른 분배]
    GG -->|Improper Game| II[REFUND 모드<br/>원래 보증금 환급]

    HH --> JJ[프로포저 보증금 환급<br/>claimCredit() 호출]
    II --> JJ

        V --> W[다른 챌린저로부터<br/>Attention 질문 수신]
    W --> X[Attention 질문 처리]
    X --> Y[Attention 응답 전송<br/>다른 챌린저에게]
    Y --> V

    V --> AA[정기적 L2 등록<br/>평판 데이터 백업]
    AA --> BB[L2 체인에 평판 기록<br/>챌린저 주소 + 평판점수]
    BB --> CC[등록자 보상 지급<br/>추후 이코노미용]
    CC --> V

    V --> Z[다음 Dispute Game 대기]
    Z --> C

    style A fill:#e1f5fe
    style F fill:#fff3e0
    style G fill:#f3e5f5
    style H fill:#f3e5f5
    style L fill:#ffebee
    style M fill:#e8f5e8
    style V fill:#f1f8e9
```

## 📋 개요

기존 P2P 챌린저 기본 구현에 RAT (Randomized Attention Test) 프로토콜을 확장하여 챌린저들이 서로에게 지속적으로 어텐션하고 있는지 확인하고, Optimistic Rollups의 검증자 무임승차 문제를 해결합니다.

## 🎯 Dispute Game의 attack()과 defend() 함수 이해

### 핵심 개념

**중요**: `attack()`과 `defend()`는 **이진 트리에서의 위치 이동**을 나타내는 함수입니다!

#### 1. 이진 트리 기반 분할 정복 시스템

```
        Root Claim (프로포저의 L2 상태)
       /                    \
   Attack (왼쪽)        Defend (오른쪽)
   /         \           /         \
Attack    Defend     Attack    Defend
```

#### 2. attack() 함수의 정확한 정의

```solidity
/// @notice Attack a disagreed upon `Claim`.
function attack(Claim _disputed, uint256 _parentIndex, Claim _claim) external payable {
    move(_disputed, _parentIndex, _claim, true);
}
```

**코드 분석 결과**:
- **`move(_isAttack = true)`**: 이진 트리에서 **왼쪽 자식 노드**로 이동
- **위치 계산**: `Position nextPosition = parentPos.move(true)`
- **실제 동작**: 부모 클레임의 왼쪽 자식에 새로운 클레임 생성
- **순수한 위치 이동**: 동의/반대와 무관한 수학적 구조

#### 3. defend() 함수의 정확한 정의

```solidity
/// @notice Defend an agreed upon `Claim`.
function defend(Claim _disputed, uint256 _parentIndex, Claim _claim) external payable {
    move(_disputed, _parentIndex, _claim, false);
}
```

**코드 분석 결과**:
- **`move(_isAttack = false)`**: 이진 트리에서 **오른쪽 자식 노드**로 이동
- **위치 계산**: `Position nextPosition = parentPos.move(false)`
- **실제 동작**: 부모 클레임의 오른쪽 자식에 새로운 클레임 생성
- **순수한 위치 이동**: 동의/반대와 무관한 수학적 구조

#### 4. Position.move() 함수의 핵심

```solidity
// LibPosition.sol
function move(Position _position, bool _isAttack) internal pure returns (Position move_) {
    assembly {
        move_ := shl(1, or(iszero(_isAttack), _position))
    }
}
```

**수학적 의미**:
- **`_isAttack = true`**: 왼쪽 자식 노드 (index * 2)
- **`_isAttack = false`**: 오른쪽 자식 노드 (index * 2 + 1)
- **순수한 위치 이동**: 동의/반대와 무관한 이진 트리 구조

### 5. 챌린저 참여 전략

#### 핵심 원칙

**`attack()`과 `defend()`는 단순한 이진 트리 위치 이동**:
- **`attack()`**: 왼쪽 자식 노드로 이동
- **`defend()`**: 오른쪽 자식 노드로 이동

#### 챌린저의 실제 판단 기준

**Enhanced Verifier를 통한 검증 후**:
- **프로포저의 클레임이 올바른 경우**: 게임에 참여하지 않음
- **프로포저의 클레임이 잘못된 경우**: 게임에 참여하여 올바른 상태로 수정
- **게임 참여 시**: 이진 트리 구조에 따라 적절한 위치에 클레임 생성

#### 게임 참여 시 보증금 시스템

**모든 게임 참여자는 보증금 지불**:
- 게임 결과에 따라 승리한 쪽이 패배한 쪽의 보증금 획득
- 잘못된 판단을 한 챌린저는 보증금 손실
- 올바른 판단을 한 챌린저는 보증금 획득

### 6. 보증금 시스템의 핵심

**모든 참여자는 보증금 지불**:
```solidity
bond: uint128(msg.value),
```

**게임 결과에 따른 분배**:
- **DEFENDER_WINS**: 프로포저가 모든 챌린저 보증금 획득
- **CHALLENGER_WINS**: 공격한 챌린저가 프로포저 보증금 획득


### 7. 핵심 포인트

1. **`attack()`**: 이진 트리에서 왼쪽 자식 노드로 이동
2. **`defend()`**: 이진 트리에서 오른쪽 자식 노드로 이동
3. **분할 정복**: 복잡한 문제를 작은 단위로 나누어 해결
4. **경제적 인센티브**: 올바른 판단을 한 쪽이 보상받음
5. **중립적 시스템**: attack/defend 모두 동등한 지위
6. **프로포저가 올바르면 챌린저는 참여하지 않아야 함**
7. **순수한 위치 이동**: 동의/반대와 무관한 이진 트리 구조

### 🎮 현재 Optimism Dispute Game 설정

**현재 사용 중인 게임 타입:**
- **DisputeGameType = 1**: PERMISSIONED_CANNON (PermissionedDisputeGame)
- **설정 위치**: `op-deployer/pkg/deployer/standard/standard.go`의 `DisputeGameType = 1`

**게임 타입 매핑:**
- **GameType 0**: CANNON (Permissionless FaultDisputeGame)
- **GameType 1**: PERMISSIONED_CANNON (PermissionedDisputeGame) ← **현재 사용 중**
- **GameType 2**: ASTERISC
- **GameType 3**: ASTERISC_KONA
- **GameType 4**: SUPER_CANNON
- **GameType 5**: SUPER_PERMISSIONED_CANNON

**현재 시스템의 특징:**
- **제한된 참여**: 오직 지정된 PROPOSER와 CHALLENGER 주소만 게임에 참여 가능
- **단일 챌린저**: 하나의 지정된 챌린저만 게임에 참여할 수 있음
- **권한 기반**: `onlyAuthorized` 모디파이어로 접근 제어

**RAT 시스템의 역할:**
현재는 단일 챌린저만 참여할 수 있지만, RAT 시스템은 향후 더 많은 챌린저가 참여할 수 있는 환경을 위한 기반을 마련하고, 현재 단일 챌린저의 주의력을 테스트하여 시스템의 안정성을 높이는 역할을 합니다.

**RAT 시스템을 위한 권장 게임 타입:**
- **GameType 0**: CANNON (Permissionless FaultDisputeGame) - **RAT 시스템에 적합**
  - 누구나 게임에 참여 가능
  - 여러 챌린저가 동시에 참여 가능
  - 분산화된 검증 환경 구축 가능
- **현재 GameType 1**: PERMISSIONED_CANNON - **RAT 시스템에 제한적**
  - 지정된 챌린저만 참여 가능
  - 단일 챌린저 제한
  - 권한 기반 접근 제어

**결론:** RAT 시스템이 제대로 작동하려면 GameType 0 (Permissionless)으로 전환하거나, 현재 GameType 1 환경에서 RAT를 보조적인 모니터링 도구로 활용하는 방안을 고려해야 합니다.

### 🔧 GameType 0으로 전환하기 위한 설정 변경사항

**1. 핵심 설정 파일 변경:**
- **`op-deployer/pkg/deployer/standard/standard.go`**:
  ```go
  // 현재: DisputeGameType = 1 (PERMISSIONED)
  // 변경: DisputeGameType = 0 (CANNON - Permissionless)
  DisputeGameType = 0
  ```

**2. 배포 시 설정 변경:**
- **환경 변수 설정**:
  ```bash
  export DISPUTE_GAME_TYPE=0
  export PERMISSIONLESS=true
  ```
- **CLI 플래그 설정**:
  ```bash
  --dispute-game-type=0 --permissionless
  ```

**3. 관련 컴포넌트 설정:**
- **Proposer 설정** (`op-proposer`):
  ```bash
  --game-type=0
  ```
- **Challenger 설정** (`op-challenger`):
  ```bash
  --trace-type=cannon
  ```

**4. 배포 파이프라인 변경:**
- **`op-deployer/pkg/deployer/pipeline/dispute_games.go`**: GameKind를 "FaultDisputeGame"으로 설정
- **`op-deployer/pkg/deployer/opcm/add_game_type.go`**: Permissioned를 false로 설정

**5. 컨트랙트 배포 변경:**
- **FaultDisputeGame** (Permissionless) 배포
- **PermissionedDisputeGame** 대신 사용
- **Proposer/Challenger 역할 제거** (누구나 참여 가능)

**6. 영향받는 시스템:**
- **OptimismPortal**: Dispute Game 생성 로직 변경
- **SystemConfig**: 기본 게임 타입 설정 변경
- **AnchorStateRegistry**: Respected Game Type 변경
- **모든 L2 체인**: 새로운 게임 타입으로 재배포 필요

**⚠️ 주의사항:**
- **기존 체인 영향**: 이미 배포된 L2 체인들은 새로운 게임 타입으로 마이그레이션 필요
- **보안 고려사항**: Permissionless 시스템은 더 많은 보안 위험을 가짐
- **테스트 필요**: 충분한 테스트 후 프로덕션 적용 권장

### 📝 수정해야 하는 파일 목록

**1. 핵심 설정 파일:**
- **`op-deployer/pkg/deployer/standard/standard.go`**:
  ```go
  // Line 30: DisputeGameType = 1 → DisputeGameType = 0
  DisputeGameType = 0 // CANNON (Permissionless)
  ```

**2. 배포 관련 파일:**
- **`op-deployer/pkg/deployer/pipeline/dispute_games.go`**:
  ```go
  // Line 107: GameKind를 "FaultDisputeGame"으로 설정
  GameKind: "FaultDisputeGame", // Permissionless
  ```
- **`op-deployer/pkg/deployer/opcm/add_game_type.go`**:
  ```go
  // Line 290: Permissioned를 false로 설정
  Permissioned: false, // Permissionless 모드
  ```

**3. 컨트랙트 배포 스크립트:**
- **`packages/contracts-bedrock/scripts/deploy/DeployDisputeGame.s.sol`**:
  ```solidity
  // Line 291: FaultDisputeGame 배포 로직 사용
  if (LibString.eq(_dgi.gameKind(), "FaultDisputeGame")) {
      // Permissionless 배포
  }
  ```
- **`packages/contracts-bedrock/scripts/deploy/SetDisputeGameImpl.s.sol`**:
  ```solidity
  // Line 60: FaultDisputeGame 구현 설정
  factory.setImplementation(gameType, impl);
  ```

**4. 시스템 설정 파일:**
- **`packages/contracts-bedrock/src/L1/OPContractsManager.sol`**:
  ```solidity
  // Line 402-436: GameType 0에 대한 처리 로직 확인
  if (gameConfig.disputeGameType.raw() == GameTypes.CANNON.raw()) {
      // Permissionless 처리
  }
  ```

**5. 개발 환경 설정:**
- **`op-devstack/presets/proof.go`**:
  ```go
  // Line 12: WithProposerGameType 설정
  func WithProposerGameType(gameType faultTypes.GameType) stack.CommonOption {
      // GameType 0 설정
  }
  ```

**6. 테스트 설정:**
- **`op-e2e/config/init.go`**:
  ```go
  // Line 436-512: 테스트용 게임 타입 설정
  AdditionalDisputeGames: []state.AdditionalDisputeGame{
      {
          DisputeGameType: 0, // CANNON
          // ...
      },
  }
  ```

**7. CLI 플래그 설정:**
- **`op-deployer/pkg/deployer/manage/flags.go`**:
  ```go
  // Line 47: 기본값 변경
  DisputeGameTypeFlag = &cli.Uint64Flag{
      Value: 0, // 기본값을 0으로 변경
  }
  ```

**8. 환경 변수 설정:**
- **`.env` 파일 또는 환경 변수**:
  ```bash
  DISPUTE_GAME_TYPE=0
  PERMISSIONLESS=true
  ```

**9. 배포 스크립트:**
- **`op-deployer/cmd/` 디렉토리의 모든 배포 스크립트**:
  - `deploy.go`
  - `add_game_type.go`
  - `upgrade.go`

**10. 문서 업데이트:**
- **`docs/` 디렉토리의 모든 관련 문서**
- **배포 가이드 및 설정 문서**
- **API 문서 및 인터페이스 설명**

**11. 테스트 파일:**
- **`packages/contracts-bedrock/test/dispute/`**:
  - `FaultDisputeGame.t.sol`
  - `DisputeGameFactory.t.sol`
- **`op-challenger/` 테스트 파일들**

**12. 설정 검증:**
- **`op-deployer/pkg/deployer/manage/add_game_type.go`**:
  ```go
  // Line 50: 설정 검증 로직 확인
  func (c *AddGameTypeConfig) Check() error {
      // GameType 0 검증 로직 추가
  }
  ```

### 🎮 Dispute Game 참여 방법 및 보증금 구조

**게임 생성 및 참여 플로우:**

**1. 게임 생성 (프로포저):**
- **함수**: `DisputeGameFactory.create(GameType _gameType, Claim _rootClaim, bytes memory _extraData)`
- **보증금**: 게임 타입별 초기 보증금 지불 (`initBonds[_gameType]`)
- **결과**: 새로운 DisputeGame 컨트랙트 배포 및 초기화

**2. 게임 감지 (챌린저):**
- **모니터링**: L1 블록 헤드 구독으로 `DisputeGameCreated` 이벤트 감지
- **게임 목록 조회**: `DisputeGameFactory.gameAtIndex()` 또는 `findLatestGames()` 사용
- **자동 스케줄링**: 새로운 게임을 발견하면 자동으로 스케줄러에 등록

**3. 게임 참여 (챌린저):**

**기본 참여 함수:**
```solidity
// 공격 (반대 의견 제시)
function attack(Claim _disputed, uint256 _parentIndex, Claim _claim) external payable

// 방어 (동의 의견 제시)
function defend(Claim _disputed, uint256 _parentIndex, Claim _claim) external payable
```

**고급 참여 함수:**
```solidity
// 최대 깊이에서의 단계 실행 (VM 검증)
function step(uint256 _claimIndex, bool _isAttack, bytes calldata _stateData, bytes calldata _proof) public virtual
```

**내부 공통 함수:**
```solidity
// 공격/방어의 공통 로직 처리
function move(Claim _disputed, uint256 _challengeIndex, Claim _claim, bool _isAttack) public payable virtual
```

**보증금 구조:**

**1. 프로포저 보증금:**
- **지불 시점**: 게임 생성 시 (`DisputeGameFactory.create()`)
- **금액**: 게임 타입별 초기 보증금 (`initBonds[_gameType]`)
- **목적**: 게임 생성 비용 및 기본 보증금

**2. 챌린저 보증금:**
- **지불 시점**: 게임 참여 시 (`attack()` 또는 `defend()`)
- **금액**: 위치별 필요 보증금 (`getRequiredBond(_movePos)`)
- **목적**: 게임 참여 및 의견 제시에 대한 보증금

**3. 보증금 관리:**
```solidity
// 보증금 예치
refundModeCredit[msg.sender] += msg.value;
WETH.deposit{ value: msg.value }();

// 보증금 환급 (게임 종료 후)
function claimCredit() external returns (uint256 credit_)
```

**GameType 0 (Permissionless) vs GameType 1 (Permissioned):**

**GameType 0 (CANNON - Permissionless):**
- **참여 제한**: 없음 (누구나 참여 가능)
- **동시 참여**: 여러 챌린저가 같은 게임에 참여 가능
- **의견 자유**: 각 챌린저가 독립적으로 공격/방어 선택
- **분산화**: 완전한 분산화된 검증 환경
- **RAT 시스템 적합성**: 높음 (여러 챌린저 주의력 테스트 가능)

**GameType 1 (PERMISSIONED_CANNON):**
- **참여 제한**: 지정된 PROPOSER와 CHALLENGER만 참여
- **동시 참여**: 단일 챌린저만 참여 가능
- **의견 제한**: 지정된 챌린저만 의견 제시
- **중앙화**: 권한 기반 접근 제어
- **RAT 시스템 적합성**: 제한적 (단일 챌린저만 테스트 가능)

**게임 진행 과정:**

**1. 초기 상태:**
- 프로포저가 루트 클레임 제시
- 게임 상태: `IN_PROGRESS`

**2. 챌린저 참여:**
- 챌린저들이 `attack()` 또는 `defend()` 함수로 참여
- 각 참여마다 새로운 클레임 추가
- 보증금 지불 및 WETH 예치

**3. 게임 진행:**
- 챌린저들이 번갈아가며 `move()` 함수로 게임 진행
- 최대 깊이에 도달하면 `step()` 함수로 VM 검증
- 게임 트리 구조로 의견 분기

**4. 게임 종료 및 보증금 분배:**

**게임 해결 과정:**
- **서브게임 해결**: `resolveClaim()` 함수로 각 서브게임 해결
- **게임 해결**: `resolve()` 함수로 전체 게임 해결
- **최종화 대기**: AnchorStateRegistry에서 게임 최종화 대기 (약 3.5일)
- **게임 종료**: `closeGame()` 함수로 게임 종료 및 보증금 분배 모드 결정

**보증금 분배 모드:**

**중요: NORMAL/REFUND 모드는 "누가 이겼는지"가 아니라 "게임이 올바르게 진행되었는지"를 판단합니다!**

**1. NORMAL 모드 (Proper Game):**
```solidity
// 게임이 올바르게 진행된 경우
if (properGame) {
    bondDistributionMode = BondDistributionMode.NORMAL;
}
```

**"Proper Game"의 조건:**
- 게임이 DisputeGameFactory에 등록되어 있음
- 게임이 블랙리스트에 없음
- 게임이 은퇴 타임스탬프 이후에 생성됨
- 시스템이 일시정지되지 않음

**NORMAL 모드에서의 보증금 분배:**

**게임 결과에 따른 정상 분배:**
- **DEFENDER_WINS**: 프로포저(방어자)가 승리 → 프로포저가 모든 챌린저들의 보증금 획득
- **CHALLENGER_WINS**: 챌린저(도전자)가 승리 → 챌린저가 프로포저의 보증금 획득

**프로포저를 방어한 챌린저의 경우:**
- **방어 챌린저**: 프로포저의 클레임을 `defend()` 함수로 방어 (보증금 지불)
- **게임 결과**:
  - 프로포저가 올바른 상태를 제출했다면 → **DEFENDER_WINS** → 프로포저가 모든 챌린저 보증금 획득
  - 프로포저가 잘못된 상태를 제출했다면 → **CHALLENGER_WINS** → 공격한 챌린저가 프로포저 보증금 획득

**보증금 환급 과정:**

**1. WETH 잠금 해제:**
```solidity
// 첫 번째 claimCredit 호출 시 잠금 해제
if (!hasUnlockedCredit[_recipient]) {
    hasUnlockedCredit[_recipient] = true;
    WETH.unlock(_recipient, recipientCredit);
    return;
}
```

**2. WETH 출금 및 ETH 전송:**
```solidity
// WETH에서 ETH로 출금
WETH.withdraw(_recipient, recipientCredit);

// ETH를 수신자에게 전송
(bool success,) = _recipient.call{ value: recipientCredit }(hex"");
if (!success) revert BondTransferFailed();
```

**보증금 분배 규칙:**

**1. 서브게임 해결 시:**
- **무반박 클레임**: 클레임한 사람이 보증금 획득
- **반박된 클레임**: 반박한 사람이 보증금 획득
- **루트 클레임**: 게임 전체 결과에 따라 분배

**2. 게임 최종 결과:**
- **DEFENDER_WINS**: 방어자가 모든 보증금 획득
- **CHALLENGER_WINS**: 도전자가 모든 보증금 획득
- **IN_PROGRESS**: 아직 진행 중 (분배 불가)

**자동 보증금 청구:**
```go
// op-challenger의 자동 보증금 청구 시스템
func (c *Claimer) ClaimBonds(ctx context.Context, games []types.GameMetadata) error {
    for _, game := range games {
        for _, claimant := range c.claimants {
            err = errors.Join(err, c.claimBond(ctx, game, claimant))
        }
    }
    return err
}
```

**보증금 분배 모니터링:**
- **op-dispute-mon**: 보증금 분배 상태 모니터링
- **예상 크레딧 vs 실제 크레딧**: 분배 정확성 검증
- **WETH 지연 시간**: 출금 지연 기간 모니터링

**RAT 시스템과의 연동:**

**현재 GameType 1 환경:**
- RAT는 단일 챌린저의 주의력 테스트 도구
- 챌린저의 평판 관리 및 모니터링
- 향후 확장을 위한 기반 마련

**GameType 0 환경 (권장):**
- RAT는 여러 챌린저의 주의력 테스트
- 분산화된 검증 환경에서의 품질 관리
- 챌린저 간 경쟁 및 협력 모니터링
- 전체 시스템의 안정성 증대

### 🔄 **최근 설계 변경사항**
- **개인화된 트리거 시스템**: 기존 0.28% 고정 확률에서 각 챌린저별 트리거 비율 설정으로 변경
- **배치 번호 기준 비율**: 배치 번호 % 100 < 트리거 비율로 간단하고 정확한 트리거
- **개인 최적화**: 각 챌린저가 자신의 평판점수, 리소스 상황, 목표에 맞게 트리거 비율 조정 가능

## 🎯 RAT 프로토콜 기본 플로우 (컨트랙트 없이, RAT Controller 없이)

### **1단계: 챌린저 네트워크 참여**
- 챌린저들이 P2P 네트워크에 참여
- 서로를 발견하고 연결 (Discovery Service)
- 챌린저 등록 및 상태 공유

### **2단계: 어텐션 테스트 트리거**
- **각 챌린저**가 개인화된 설정으로 배치 번호 기준 트리거 결정 (개인별 트리거 비율 설정)
- **간단한 비율 계산**: 배치 번호 % 100 < 트리거 비율 → 트리거 실행
- **정확한 비율 달성**: 10% 설정 시 배치 번호 끝자리가 0-9일 때 정확히 트리거
- **트리거 실행 방법**:
  - **L2 배치 이벤트 모니터링 (주된 동작)**: L2 배치가 L1에 제출될 때 이벤트를 모니터링하여 자동 트리거
  - **운영자 수동 트리거 (보조 동작)**: 운영자가 특정 챌린저에게 수동으로 트리거 발생
- 선택된 챌린저에게 **다른 챌린저들의 어텐션 확인 테스트** 출제
- 테스트에는 **다른 챌린저들의 상태, 평판, 활동성**에 대한 질문 포함
- **다른 챌린저들에게 어텐션하고 있지 않으면 답할 수 없는 문제**로 무임승차 방지

### **3단계: 챌린저 응답**
- 선택된 챌린저가 **다른 챌린저들 어텐션 테스트**에 응답
- **다른 챌린저들의 상태, 평판, 활동성 정보**를 기반으로 답변
- **다른 챌린저들에게 어텐션하고 있다면** 올바른 답변 제출 가능
- **다른 챌린저들에게 어텐션하지 않았다면** 답할 수 없어 평판 하락

### **4단계: 결과 처리 및 양방향 평판 관리**
- **응답자 평판**: 응답하지 않거나 잘못된 응답 시 점수 하락, 정확한 응답 시 점수 상승
- **질문자 평판**: 적극적으로 테스트를 수행하고 질문을 할수록 점수 상승
- **종합 평판**: 질문자와 응답자 점수를 종합하여 네트워크 내 지위 결정
- **공정성 보장**: 모든 챌린저가 동등한 기회로 테스트를 받고 수행

### 추후 고려사항.
- 챌린저들의 주소정보와 평판점수가 주기적으로 기록이 되어야 함.
- 모니터링하고 있는 L2에 기록하는 것이 어떤지.. (가격저렴)
- 누가 기록할것이면. 그것이 정확한지에 대한 합의 필요.
- 온체인에 기록된 정보를 바탕으로 추후 보상 지급에 대한 이코노미 설계 (L2에 트랜잭션 실행해서 등록한 등록자에게 추가 보상고려)


## 🔧 구현해야 할 모듈들

### **1. Distributed Attention Question System** (`p2p/attention/question.go`)
       ```go
       // 핵심 기능:
       - ShouldAskAttentionQuestion() // Dispute Game 번호 기반 질문 비율 계산
       - SelectRandomChallenger() // 결정적 무작위성 기반 챌린저 선택
       - BroadcastAttentionQuestion() // P2P 네트워크로 어텐션 질문 전파
       - CollectQuestionConsensus() // 질문 합의 수집 (51% 이상 동의 필요)
       - UpdatePeerList() // 챌린저 목록 관리 (추가/제거)
       - GetAvailableChallengerCount() // 사용 가능한 챌린저 수 조회
       - MonitorDisputeGameEvents() // Dispute Game 이벤트 모니터링 (자동 질문)
       - ShouldAskManualQuestion() // 수동 질문 가능 여부 확인
       - AskManualQuestion() // 운영자 수동 질문 실행 (특정 챌린저 지정)
       - ExecuteAttentionQuestion() // 어텐션 질문 실행
       - ValidateAttentionResponse() // 어텐션 응답 검증
       ```

### **2. Enhanced Verifier** (`op-challenger/verification/enhanced.go`)
       ```go
       // 핵심 기능:
       - ValidateL2StateRoot() // L2 상태 루트 검증
       - ValidateBatchData() // 배치 데이터 검증
       - ValidateL1L2Consistency() // L1/L2 일치성 검증
       - GetVerificationResult() // 통합 검증 결과 반환
       - BatchChecker() // L1/L2 배치 데이터 일치성 확인
       - ValidateBatchHash() // 배치 해시 검증
       - ValidateBatchOrder() // 배치 순서 검증
       - ClassifyVerificationResult() // 검증 결과 분류 (VALID/INVALID_STATE/INVALID_BATCH/INVALID_BOTH)
       ```

### **3. Hybrid Reputation Store** (`p2p/reputation/hybrid_store.go`)
       ```go
       // 핵심 기능:
       - StoreLocalReputation() // 로컬 LevelDB에 평판 데이터 저장
       - SyncPeerReputations() // P2P 네트워크를 통한 실시간 평판 동기화
       - BackupToL2Node() // L2 노드에 백업 (저비용, 높은 가용성)
       - BackupToIPFS() // 주기적 IPFS 백업 (장기 보관)
       - ProposeReputationUpdate() // 평판 변경 제안
       - CollectReputationConsensus() // 평판 변경 합의 수집
       - ValidateReputationBlock() // 평판 블록 검증
       - GetConsensusReputation() // 합의된 평판 데이터 조회
       - RestoreFromL2Backup() // L2 노드 백업에서 복원
       - RestoreFromIPFSBackup() // IPFS 백업에서 복원
       - RecordToL2Chain() // L2 체인에 평판 정보 기록 (추후 보상용)
       ```

### **4. Attention Test Handler** (`p2p/challenger/attention.go`)
       ```go
       // 핵심 기능:
       - HandleIncomingQuestion() // 수신된 어텐션 질문 처리
       - GetPeerAttentionInfo() // 다른 챌린저들의 어텐션 정보 수집
       - SubmitQuestionResponse() // 어텐션 질문 응답 제출
       - ValidateOwnResponse() // 자신의 응답 검증
       - CollectPeerStatus() // 다른 챌린저들의 상태 정보 수집
       - CollectPeerReputation() // 다른 챌린저들의 평판 정보 수집
       - CollectPeerActivity() // 다른 챌린저들의 활동성 정보 수집
       ```

### **5. Reputation Manager** (`p2p/reputation/manager.go`)
       ```go
       // 핵심 기능:
       - UpdateQuestionerScore() // 질문자 평판 점수 업데이트 (적극성, 질문 빈도)
       - UpdateResponderScore() // 응답자 평판 점수 업데이트 (응답성, 정확성, 속도)
       - GetCombinedScore() // 종합 평판 점수 조회
       - ApplyPenalty() // 페널티 적용 (응답 실패, 부정확한 응답)
       - ApplyReward() // 보상 적용 (정확한 응답, 적극적 참여)
       - CalculateFairnessScore() // 공정성 점수 계산
       - DetectCollusion() // 담합 탐지 및 위험도 계산
       - CalculateNodeAffinity() // 노드 간 밀접성 계산
       - SeparateColludingGroups() // 담합 그룹 분리
       - CalculateHonestyScore() // 정직성 점수 계산
       - DistributeSynergy() // 시뇨리지 분배 (정직성/성과 기반)
       - ValidateWithMultiLayer() // 다중 검증 계층을 통한 검증
       - RecordToL2ForReward() // L2에 기록하여 추후 보상 지급 준비
       ```

### **6. Peer Attention Validator** (`p2p/attention/validator.go`)
       ```go
       // 핵심 기능:
       - ValidatePeerAttention() // 다른 챌린저들의 어텐션 검증
       - CollectPeerInfo() // 다른 챌린저들의 정보 수집
       - VerifyAttentionResponse() // 어텐션 응답 검증
       - GetAttentionData() // 어텐션 데이터 조회
       - ValidatePeerStatus() // 다른 챌린저들의 상태 검증
       - ValidatePeerReputation() // 다른 챌린저들의 평판 검증
       - ValidatePeerActivity() // 다른 챌린저들의 활동성 검증
       ```

### **7. L2 Chain Recorder** (`p2p/l2/recorder.go`)
       ```go
       // 핵심 기능:
       - RecordChallengerInfo() // 챌린저 주소정보와 평판점수를 L2에 기록
       - ValidateRecordAccuracy() // 기록된 정보의 정확성 검증
       - CollectRecordingConsensus() // 기록에 대한 합의 수집
       - PrepareRewardEconomy() // 온체인 기록 기반 보상 이코노미 준비
       - DistributeRecordingReward() // L2 트랜잭션 실행 등록자에게 추가 보상
       ```

## 🚀 구현 순서

### **Phase 1: 핵심 기능 구현 (1-2개월)**
1. **Distributed Attention Question System** (완전 분산화된 어텐션 질문 시스템)
   - TODO 1.1: 기본 구조 설계
   - TODO 1.2: P2P 챌린저 네트워크 구축
   - TODO 1.3: 개인화된 질문 로직 구현
   - TODO 1.4: 챌린저 간 Attention 질문 시스템
   - TODO 1.5: 질문 수신 및 응답 처리 시스템
   - TODO 1.6: 수동 Attention 질문 실행
   - TODO 1.7: 단위 테스트

2. **Enhanced Verifier** (L2 상태 및 배치 데이터 통합 검증)
   - TODO 2.1: 통합 검증기 구조
   - TODO 2.2: 검증 프로세스
   - TODO 2.3: 결과 분류
   - TODO 2.4: 단위 테스트

### **Phase 2: 고급 기능 구현 (2-3개월)**
3. **배치 검증 성능 최적화** (병렬 검증 및 캐싱 전략)
   - TODO 4.1: 병렬 검증
   - TODO 4.2: 캐싱 전략
   - TODO 4.3: 단위 테스트

### **테스트 및 최적화**
4. **통합 테스트** (TODO 5.1)
   - 전체 RAT 플로우 테스트
   - 배치 데이터 검증 시스템 통합 테스트
   - 성능 테스트

5. **성능 최적화** (TODO 5.2)
   - 메모리 사용량 최적화 (< 500MB)
   - CPU 사용률 최적화 (< 5%)
   - 배치 검증 성능 최적화 (< 1초)


## 🔗 기존 P2P 챌린저 시스템과의 통합

### **Discovery Service 확장**
- 기존 노드 발견 기능에 챌린저 등록 기능 추가
- 챌린저 상태 모니터링 및 업데이트

### **메시지 시스템 확장**
- RAT 관련 메시지 타입 추가 완료
- 어텐션 테스트 메시지 처리 로직 구현 필요

### **네트워크 관리 확장**
- 챌린저 평판 기반 네트워크 관리
- 비활성 챌린저 자동 제거

### **L2 가스비 최적화**
- 낮은 가스비로 빈번한 평판 업데이트 가능
- 실시간 평판 동기화 및 검증
- 비용 효율적인 RAT 테스트 실행
