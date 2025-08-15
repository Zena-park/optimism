// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Math.sol";

/**
 * @title LightRATContract
 * @dev 가스비를 최소화한 RAT 시스템의 L1 컨트랙트
 * L2에서 대부분의 Attention Test를 처리하고, L1에는 배치 결과만 기록
 */
contract LightRATContract is ReentrancyGuard, Ownable {
    using Math for uint256;

    // ============ STRUCTS ============

    struct BatchResult {
        bytes32 batchId;
        uint256 startBlock;
        uint256 endBlock;
        uint256 totalTests;
        uint256 passedTests;
        uint256 failedTests;
        bytes32 merkleRoot;  // L2 결과의 Merkle 루트
        uint256 timestamp;
        bool submitted;
    }

    struct Validator {
        address validator;
        uint256 deposit;
        bool isActive;
        uint256 totalPenalties;
        uint256 lastBatchId;
    }

    // ============ STATE VARIABLES ============

    mapping(bytes32 => BatchResult) public batchResults;
    mapping(address => Validator) public validators;
    mapping(address => bool) public isRegisteredValidator;
    mapping(address => uint256) public penalties;

    address[] public validatorList;

    // RAT Parameters (가스비 최적화)
    uint256 public batchSize = 100;  // 배치 크기
    uint256 public minDeposit;       // 최소 예치금
    uint256 public penaltyAmount;    // 페널티 금액
    uint256 public verificationThreshold = 10; // 10% 이상 실패 시 L1 검증

    // Events
    event ValidatorRegistered(address indexed validator, uint256 deposit);
    event ValidatorUnregistered(address indexed validator);
    event BatchResultSubmitted(
        bytes32 indexed batchId,
        uint256 totalTests,
        uint256 passedTests,
        uint256 failedTests,
        bytes32 merkleRoot
    );
    event PenaltiesApplied(
        address[] failedValidators,
        uint256[] penaltyAmounts
    );
    event L1VerificationTriggered(
        bytes32 indexed batchId,
        uint256 failureRate
    );

    // ============ CONSTRUCTOR ============

    constructor(
        uint256 _minDeposit,
        uint256 _penaltyAmount
    ) {
        minDeposit = _minDeposit;
        penaltyAmount = _penaltyAmount;
    }

    // ============ MODIFIERS ============

    modifier onlyValidator() {
        require(isRegisteredValidator[msg.sender], "Not a registered validator");
        _;
    }

    modifier onlyProposer() {
        require(msg.sender == owner() || isRegisteredValidator[msg.sender], "Not authorized");
        _;
    }

    // ============ VALIDATOR MANAGEMENT ============

    /**
     * @dev 검증자 등록 (가스비 최적화)
     */
    function registerValidator() external payable {
        require(msg.value >= minDeposit, "Insufficient deposit");
        require(!isRegisteredValidator[msg.sender], "Already registered");

        validators[msg.sender] = Validator({
            validator: msg.sender,
            deposit: msg.value,
            isActive: true,
            totalPenalties: 0,
            lastBatchId: 0
        });

        isRegisteredValidator[msg.sender] = true;
        validatorList.push(msg.sender);

        emit ValidatorRegistered(msg.sender, msg.value);
    }

    /**
     * @dev 검증자 해제 (가스비 최적화)
     */
    function unregisterValidator() external onlyValidator nonReentrant {
        Validator storage validator = validators[msg.sender];
        require(validator.isActive, "Validator not active");

        uint256 deposit = validator.deposit - validator.totalPenalties;
        validator.isActive = false;
        validator.deposit = 0;

        // 검증자 목록에서 제거 (가스비 최적화)
        _removeValidatorFromList(msg.sender);
        isRegisteredValidator[msg.sender] = false;

        // 예치금 반환
        if (deposit > 0) {
            (bool success, ) = msg.sender.call{value: deposit}("");
            require(success, "Deposit transfer failed");
        }

        emit ValidatorUnregistered(msg.sender);
    }

    // ============ BATCH PROCESSING ============

    /**
     * @dev 배치 결과 제출 (가스비 최소화)
     * L2에서 처리된 Attention Test 결과를 배치로 제출
     */
    function submitBatchResult(
        bytes32 batchId,
        uint256 startBlock,
        uint256 endBlock,
        uint256 totalTests,
        uint256 passedTests,
        uint256 failedTests,
        bytes32 merkleRoot
    ) external onlyProposer {
        require(!batchResults[batchId].submitted, "Batch already submitted");
        require(totalTests == passedTests + failedTests, "Invalid test counts");

        batchResults[batchId] = BatchResult({
            batchId: batchId,
            startBlock: startBlock,
            endBlock: endBlock,
            totalTests: totalTests,
            passedTests: passedTests,
            failedTests: failedTests,
            merkleRoot: merkleRoot,
            timestamp: block.timestamp,
            submitted: true
        });

        emit BatchResultSubmitted(batchId, totalTests, passedTests, failedTests, merkleRoot);

        // 실패율이 높으면 L1 검증 트리거
        if (failedTests > 0) {
            uint256 failureRate = (failedTests * 100) / totalTests;
            if (failureRate >= verificationThreshold) {
                emit L1VerificationTriggered(batchId, failureRate);
            }
        }
    }

    /**
     * @dev 페널티 일괄 적용 (가스비 최소화)
     * 실패한 검증자들에 대해 배치로 페널티 적용
     */
    function applyPenalties(
        address[] calldata failedValidators,
        uint256[] calldata penaltyAmounts
    ) external onlyProposer {
        require(failedValidators.length == penaltyAmounts.length, "Array length mismatch");

        for (uint i = 0; i < failedValidators.length; i++) {
            address validator = failedValidators[i];
            uint256 penalty = penaltyAmounts[i];

            if (isRegisteredValidator[validator] && validators[validator].isActive) {
                uint256 actualPenalty = Math.min(penalty, validators[validator].deposit);

                if (actualPenalty > 0) {
                    validators[validator].deposit -= actualPenalty;
                    validators[validator].totalPenalties += actualPenalty;
                    penalties[validator] += actualPenalty;
                }
            }
        }

        emit PenaltiesApplied(failedValidators, penaltyAmounts);
    }

    // ============ SELECTIVE VERIFICATION ============

    /**
     * @dev L1 검증 필요 여부 확인
     * 실패율이 임계값을 초과하면 L1에서 상세 검증 수행
     */
    function shouldVerifyOnL1(bytes32 batchId) public view returns (bool) {
        BatchResult storage batch = batchResults[batchId];
        if (!batch.submitted) return false;

        uint256 failureRate = (batch.failedTests * 100) / batch.totalTests;
        return failureRate >= verificationThreshold;
    }

    /**
     * @dev L1 상세 검증 수행 (필요한 경우에만)
     */
    function performL1Verification(
        bytes32 batchId,
        bytes32[] calldata proofs
    ) external onlyProposer {
        require(shouldVerifyOnL1(batchId), "L1 verification not needed");

        // L1에서 상세 검증 로직 수행
        // 여기서는 간단한 구조만 구현
        BatchResult storage batch = batchResults[batchId];

        // 실제 구현에서는 Merkle 증명 검증 등을 수행
        // verifyMerkleProofs(batch.merkleRoot, proofs);
    }

    // ============ VIEW FUNCTIONS ============

    /**
     * @dev 배치 결과 조회
     */
    function getBatchResult(bytes32 batchId) external view returns (BatchResult memory) {
        return batchResults[batchId];
    }

    /**
     * @dev 검증자 정보 조회
     */
    function getValidator(address validator) external view returns (Validator memory) {
        return validators[validator];
    }

    /**
     * @dev 모든 검증자 목록 조회
     */
    function getAllValidators() external view returns (address[] memory) {
        return validatorList;
    }

    /**
     * @dev 검증자 수 조회
     */
    function getValidatorCount() external view returns (uint256) {
        return validatorList.length;
    }

    /**
     * @dev 페널티 정보 조회
     */
    function getPenalty(address validator) external view returns (uint256) {
        return penalties[validator];
    }

    // ============ ADMIN FUNCTIONS ============

    /**
     * @dev 매개변수 업데이트 (가스비 최적화)
     */
    function updateParameters(
        uint256 _batchSize,
        uint256 _minDeposit,
        uint256 _penaltyAmount,
        uint256 _verificationThreshold
    ) external onlyOwner {
        batchSize = _batchSize;
        minDeposit = _minDeposit;
        penaltyAmount = _penaltyAmount;
        verificationThreshold = _verificationThreshold;
    }

    /**
     * @dev 누적된 페널티 출금
     */
    function withdrawPenalties() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No penalties to withdraw");

        (bool success, ) = owner().call{value: balance}("");
        require(success, "Withdrawal failed");
    }

    // ============ INTERNAL FUNCTIONS ============

    /**
     * @dev 검증자 목록에서 제거 (가스비 최적화)
     */
    function _removeValidatorFromList(address validator) internal {
        for (uint i = 0; i < validatorList.length; i++) {
            if (validatorList[i] == validator) {
                // 마지막 요소를 현재 위치로 이동
                validatorList[i] = validatorList[validatorList.length - 1];
                validatorList.pop();
                break;
            }
        }
    }

    /**
     * @dev 배치 ID 생성 (가스비 최적화)
     */
    function _generateBatchId(
        uint256 startBlock,
        uint256 endBlock
    ) internal view returns (bytes32) {
        return keccak256(
            abi.encodePacked(
                startBlock,
                endBlock,
                block.timestamp,
                blockhash(block.number - 1)
            )
        );
    }
}
