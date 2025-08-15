// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Math.sol";

/**
 * @title IntegratedRATContract
 * @dev 논문 기반 RAT 통합 컨트랙트
 * 기존 ORU 상태 커밋먼트 제출 과정에 RAT 로직을 통합
 */
contract IntegratedRATContract is ReentrancyGuard, Ownable {
    using Math for uint256;

    // ============ STRUCTS ============

    struct Validator {
        address validator;
        uint256 deposit;
        bool isActive;
        uint256 score;
        uint256 totalPenalties;
    }

    struct AttentionTest {
        bytes32 stateRoot;
        address targetValidator;
        uint256 startTime;
        uint256 responseWindow;
        bool completed;
        bool passed;
    }

    // ============ STATE VARIABLES ============

    // ORU 상태 커밋먼트 관리
    mapping(uint256 => bytes32) public stateCommitments;
    mapping(uint256 => uint256) public commitmentTimestamps;

    // RAT 관련 상태
    mapping(bytes32 => AttentionTest) public activeTests;
    mapping(address => Validator) public validators;
    mapping(address => bool) public isRegisteredValidator;
    mapping(address => uint256) public penalties;

    address[] public validatorList;

    // RAT Parameters (논문 기반)
    uint256 public attentionTestProbability = 28; // 0.28% (28/10000)
    uint256 public responseWindow = 120; // 2분 (120초)
    uint256 public minDeposit;
    uint256 public penaltyAmount;

    // Events
    event StateCommitmentSubmitted(
        bytes32 indexed stateRoot,
        uint256 indexed l2BlockNumber,
        bytes32 testId
    );
    event AttentionTestTriggered(
        bytes32 indexed testId,
        bytes32 indexed stateRoot,
        address indexed targetValidator
    );
    event AttentionTestPassed(
        bytes32 indexed testId,
        address indexed validator
    );
    event AttentionTestFailed(
        bytes32 indexed testId,
        address indexed validator
    );
    event ValidatorRegistered(address indexed validator, uint256 deposit);
    event ValidatorUnregistered(address indexed validator);
    event PenaltyApplied(address indexed validator, uint256 amount);

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

    // ============ CORE ORU + RAT INTEGRATION ============

    /**
     * @dev 상태 커밋먼트 제출 및 RAT 통합 (논문 기반)
     * 기존 ORU 상태 커밋먼트 제출 과정에 RAT 로직을 통합
     */
    function submitStateCommitment(
        bytes32 stateRoot,
        uint256 l2BlockNumber
    ) external onlyProposer returns (bytes32 testId) {
        // 1. 기존 ORU 로직: 상태 커밋먼트 기록
        stateCommitments[l2BlockNumber] = stateRoot;
        commitmentTimestamps[l2BlockNumber] = block.timestamp;

        // 2. RAT 로직: Attention Test 트리거 (확률적)
        if (shouldTriggerAttentionTest(stateRoot)) {
            testId = triggerAttentionTest(stateRoot, l2BlockNumber);
        }

        // 3. 이벤트 발생
        emit StateCommitmentSubmitted(stateRoot, l2BlockNumber, testId);

        return testId;
    }

    /**
     * @dev Attention Test 트리거 결정 (논문 기반)
     */
    function shouldTriggerAttentionTest(
        bytes32 stateRoot
    ) internal view returns (bool) {
        // 논문의 확률적 결정: πa 확률로 테스트 트리거
        bytes32 randomness = keccak256(
            abi.encodePacked(stateRoot, blockhash(block.number - 1))
        );

        uint256 threshold = (2**256 - 1) * attentionTestProbability / 10000;
        return uint256(randomness) <= threshold;
    }

    /**
     * @dev Attention Test 트리거 (논문 기반)
     */
    function triggerAttentionTest(
        bytes32 stateRoot,
        uint256 l2BlockNumber
    ) internal returns (bytes32 testId) {
        require(validatorList.length > 0, "No validators registered");

        // 무작위 검증자 선택
        address targetValidator = selectTargetValidator(stateRoot);

        // 테스트 ID 생성
        testId = keccak256(
            abi.encodePacked(stateRoot, targetValidator, block.timestamp)
        );

        // Attention Test 기록
        activeTests[testId] = AttentionTest({
            stateRoot: stateRoot,
            targetValidator: targetValidator,
            startTime: block.timestamp,
            responseWindow: responseWindow,
            completed: false,
            passed: false
        });

        emit AttentionTestTriggered(testId, stateRoot, targetValidator);
    }

    /**
     * @dev 무작위 검증자 선택 (논문 기반)
     */
    function selectTargetValidator(
        bytes32 stateRoot
    ) internal view returns (address) {
        bytes32 seed = keccak256(
            abi.encodePacked(stateRoot, blockhash(block.number - 1))
        );

        uint256 selectedIndex = uint256(seed) % validatorList.length;
        return validatorList[selectedIndex];
    }

    // ============ VALIDATOR RESPONSE ============

    /**
     * @dev Attention Test 솔루션 제출 (논문 기반)
     */
    function submitAttentionSolution(
        bytes32 testId,
        bytes32 leftChild,
        bytes32 rightChild
    ) external onlyValidator {
        AttentionTest storage test = activeTests[testId];
        require(test.targetValidator == msg.sender, "Not target validator");
        require(!test.completed, "Test already completed");
        require(
            block.timestamp <= test.startTime + test.responseWindow,
            "Response window expired"
        );

        // 솔루션 검증
        bool isValid = verifyAttentionSolution(leftChild, rightChild, test.stateRoot);

        // 결과 처리
        test.completed = true;
        test.passed = isValid;

        if (isValid) {
            // 성공: 검증자 점수 증가
            validators[msg.sender].score++;
            emit AttentionTestPassed(testId, msg.sender);
        } else {
            // 실패: 페널티 적용
            applyPenalty(msg.sender);
            emit AttentionTestFailed(testId, msg.sender);
        }
    }

    /**
     * @dev Attention Puzzle 검증 (논문 기반)
     */
    function verifyAttentionSolution(
        bytes32 leftChild,
        bytes32 rightChild,
        bytes32 expectedStateRoot
    ) internal pure returns (bool) {
        bytes32 computedRoot = hashTree(leftChild, rightChild);
        return computedRoot == expectedStateRoot;
    }

    /**
     * @dev Merkle 트리 해시 계산 (논문 기반)
     */
    function hashTree(
        bytes32 left,
        bytes32 right
    ) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(left, right));
    }

    // ============ VALIDATOR MANAGEMENT ============

    /**
     * @dev 검증자 등록
     */
    function registerValidator() external payable {
        require(msg.value >= minDeposit, "Insufficient deposit");
        require(!isRegisteredValidator[msg.sender], "Already registered");

        validators[msg.sender] = Validator({
            validator: msg.sender,
            deposit: msg.value,
            isActive: true,
            score: 0,
            totalPenalties: 0
        });

        isRegisteredValidator[msg.sender] = true;
        validatorList.push(msg.sender);

        emit ValidatorRegistered(msg.sender, msg.value);
    }

    /**
     * @dev 검증자 해제
     */
    function unregisterValidator() external onlyValidator nonReentrant {
        Validator storage validator = validators[msg.sender];
        require(validator.isActive, "Validator not active");

        uint256 deposit = validator.deposit - validator.totalPenalties;
        validator.isActive = false;
        validator.deposit = 0;

        // 검증자 목록에서 제거
        _removeValidatorFromList(msg.sender);
        isRegisteredValidator[msg.sender] = false;

        // 예치금 반환
        if (deposit > 0) {
            (bool success, ) = msg.sender.call{value: deposit}("");
            require(success, "Deposit transfer failed");
        }

        emit ValidatorUnregistered(msg.sender);
    }

    // ============ PENALTY MECHANISM ============

    /**
     * @dev 페널티 적용
     */
    function applyPenalty(address validator) internal {
        Validator storage val = validators[validator];
        if (val.isActive && val.deposit > 0) {
            uint256 actualPenalty = Math.min(penaltyAmount, val.deposit);

            if (actualPenalty > 0) {
                val.deposit -= actualPenalty;
                val.totalPenalties += actualPenalty;
                penalties[validator] += actualPenalty;

                emit PenaltyApplied(validator, actualPenalty);
            }
        }
    }

    // ============ VIEW FUNCTIONS ============

    /**
     * @dev 상태 커밋먼트 조회
     */
    function getStateCommitment(uint256 l2BlockNumber) external view returns (bytes32) {
        return stateCommitments[l2BlockNumber];
    }

    /**
     * @dev Attention Test 조회
     */
    function getAttentionTest(bytes32 testId) external view returns (AttentionTest memory) {
        return activeTests[testId];
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
     * @dev RAT 매개변수 업데이트
     */
    function updateRATParameters(
        uint256 _attentionTestProbability,
        uint256 _responseWindow,
        uint256 _minDeposit,
        uint256 _penaltyAmount
    ) external onlyOwner {
        attentionTestProbability = _attentionTestProbability;
        responseWindow = _responseWindow;
        minDeposit = _minDeposit;
        penaltyAmount = _penaltyAmount;
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
     * @dev 검증자 목록에서 제거
     */
    function _removeValidatorFromList(address validator) internal {
        for (uint i = 0; i < validatorList.length; i++) {
            if (validatorList[i] == validator) {
                validatorList[i] = validatorList[validatorList.length - 1];
                validatorList.pop();
                break;
            }
        }
    }
}
