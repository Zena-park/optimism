# L1FeeVault 보상 메커니즘 분석

## 1. **L1FeeVault 구조 분석**

### 코드 구조
```solidity
// packages/tokamak/contracts-bedrock/src/L2/L1FeeVault.sol
contract L1FeeVault is FeeVault {
    constructor(
        address _recipient,           // 수수료 수취자 주소
        uint256 _minWithdrawalAmount, // 최소 인출 금액
        WithdrawalNetwork _withdrawalNetwork // L1 또는 L2 네트워크
    )
}
```

### 상속받은 FeeVault의 핵심 기능
```solidity
// packages/tokamak/contracts-bedrock/src/universal/FeeVault.sol
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

## 2. **현재 Tokamak Thanos 설정**

### thanos-sepolia.json 설정
```json
{
  "l1FeeVaultRecipient": "0x7cf8aD47c1B5E18cd827633F9e25538A08dB5448",
  "l1FeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000", // 10 ETH
  "l1FeeVaultWithdrawalNetwork": 0  // L1 네트워크
}
```

### thanos-stack-sepolia.json 설정
```json
{
  "l1FeeVaultRecipient": "0xC9402d34Cd9B375d1eD19c92c9Ac325738301FA5",
  "l1FeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000", // 10 ETH
  "l1FeeVaultWithdrawalNetwork": 0  // L1 네트워크
}
```

## 3. **보상 메커니즘 동작 방식**

### A. **수수료 누적**
1. L2 사용자가 트랜잭션 실행 시 L1 데이터 가용성 비용 지불
2. 해당 비용이 L1FeeVault (`0x420000000000000000000000000000000000001a`)에 자동 누적
3. 누적된 수수료는 컨트랙트 잔액으로 저장

### B. **인출 과정**
1. **인출 조건**: 컨트랙트 잔액이 최소 인출 금액(10 ETH) 이상
2. **인출 실행**: 누구나 `withdraw()` 함수 호출 가능
3. **수취자**: `l1FeeVaultRecipient`로 설정된 주소가 모든 수수료 수취

### C. **배처 보상 설정 방법**

#### 방법 1: 직접 배처 주소 설정
```json
{
  "l1FeeVaultRecipient": "0x[BATCHER_ADDRESS]"  // 배처 운영자 주소로 직접 설정
}
```

#### 방법 2: 멀티시그/DAO 주소 설정 후 분배
```json
{
  "l1FeeVaultRecipient": "0x[MULTISIG_ADDRESS]"  // 멀티시그 주소로 설정
}
```
- 멀티시그에서 배처 운영자에게 적절한 비율로 분배

## 4. **실제 보상 가능성 검증**

### 수수료 발생량
- **L1 데이터 비용**: L2 트랜잭션마다 L1 가스비 상당액이 L1FeeVault에 누적
- **배처 비용**: L1에 배치 제출 시 실제 가스비 지출
- **수익성**: L1FeeVault 누적액 > 배처 실제 비용이면 보상 가능

### 현재 설정의 문제점
- **현재**: 모든 Fee Vault의 수취자가 동일한 주소로 설정됨
- **결과**: 시퀀서, Base Fee, L1 Fee가 모두 같은 주소로 집중
- **개선**: 각 역할별로 다른 주소 설정 또는 분배 로직 구현

## 5. **권장 설정 방안**

### A. **역할별 분리**
```json
{
  "sequencerFeeVaultRecipient": "0x[SEQUENCER_ADDRESS]",
  "l1FeeVaultRecipient": "0x[BATCHER_ADDRESS]",
  "baseFeeVaultRecipient": "0x[PROTOCOL_TREASURY]"
}
```

### B. **중앙 집중 후 분배**
```json
{
  "l1FeeVaultRecipient": "0x[DISTRIBUTION_CONTRACT]"  // 자동 분배 컨트랙트
}
```

## 6. **결론**

### ✅ **보상 가능함**
- L1FeeVault는 **이미 동작하는 보상 시스템**
- `l1FeeVaultRecipient` 설정만 변경하면 **즉시 보상 수취 가능**
- 별도 컨트랙트 배포나 복잡한 설정 불필요

### 🔧 **설정 방법**
1. **배포 설정 파일**에서 `l1FeeVaultRecipient`를 배처 주소로 변경
2. **최소 인출 금액** 조정 (필요시)
3. **인출 네트워크** 설정 (L1 또는 L2)

### 💰 **경제성**
- L2 사용량에 비례하여 L1FeeVault 누적
- 배처 운영 비용과 비교하여 수익성 판단 가능
- **실제 많은 L2 프로젝트에서 이 방식으로 배처 보상 운영 중**
