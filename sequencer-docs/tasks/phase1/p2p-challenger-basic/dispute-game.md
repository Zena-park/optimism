
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