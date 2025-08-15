// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/utils/Math.sol";

contract RandomizedAttentionTest is ReentrancyGuard, Ownable {
    using Counters for Counters.Counter;
    using Math for uint256;

    // ============ STRUCTS ============

    struct Validator {
        address validator;
        uint256 deposit;
        bool isActive;
        uint256 lastTestTime;
        uint256 totalTests;
        uint256 passedTests;
        uint256 failedTests;
    }

    struct AttentionTest {
        bytes32 testId;
        bytes32 stateRoot;
        uint256 l2BlockNumber;
        address targetValidator;
        uint256 startTime;
        uint256 responseWindow;
        bool completed;
        bool passed;
        bytes32 leftChild;
        bytes32 rightChild;
    }

    // ============ STATE VARIABLES ============

    mapping(address => Validator) public validators;
    mapping(bytes32 => AttentionTest) public activeTests;
    mapping(address => bool) public isRegisteredValidator;

    address[] public validatorList;
    Counters.Counter private _testCounter;

    // RAT Parameters
    uint256 public attentionTestProbability; // πa (in basis points, 100 = 1%)
    uint256 public responseWindow; // Tresponse in seconds
    uint256 public minDeposit; // Minimum validator deposit
    uint256 public penaltyAmount; // coff

    // Events
    event ValidatorRegistered(address indexed validator, uint256 deposit);
    event ValidatorUnregistered(address indexed validator);
    event StateCommitmentSubmitted(
        bytes32 indexed stateRoot,
        uint256 l2BlockNumber,
        bytes32 testId
    );
    event AttentionTestTriggered(
        bytes32 indexed testId,
        bytes32 stateRoot,
        address indexed targetValidator
    );
    event AttentionTestCompleted(
        bytes32 indexed testId,
        address indexed validator,
        bool passed
    );
    event PenaltyApplied(address indexed validator, uint256 amount);

    // ============ CONSTRUCTOR ============

    constructor(
        uint256 _attentionTestProbability,
        uint256 _responseWindow,
        uint256 _minDeposit,
        uint256 _penaltyAmount
    ) {
        attentionTestProbability = _attentionTestProbability;
        responseWindow = _responseWindow;
        minDeposit = _minDeposit;
        penaltyAmount = _penaltyAmount;
    }

    // ============ MODIFIERS ============

    modifier onlyValidator() {
        require(isRegisteredValidator[msg.sender], "Not a registered validator");
        _;
    }

    modifier onlyActiveValidator() {
        require(
            isRegisteredValidator[msg.sender] && validators[msg.sender].isActive,
            "Not an active validator"
        );
        _;
    }

    // ============ VALIDATOR MANAGEMENT ============

    /**
     * @dev Register as a validator with deposit
     */
    function registerValidator() external payable {
        require(msg.value >= minDeposit, "Insufficient deposit");
        require(!isRegisteredValidator[msg.sender], "Already registered");

        validators[msg.sender] = Validator({
            validator: msg.sender,
            deposit: msg.value,
            isActive: true,
            lastTestTime: 0,
            totalTests: 0,
            passedTests: 0,
            failedTests: 0
        });

        isRegisteredValidator[msg.sender] = true;
        validatorList.push(msg.sender);

        emit ValidatorRegistered(msg.sender, msg.value);
    }

    /**
     * @dev Unregister validator and withdraw deposit
     */
    function unregisterValidator() external onlyValidator nonReentrant {
        Validator storage validator = validators[msg.sender];
        require(validator.isActive, "Validator not active");

        uint256 deposit = validator.deposit;
        validator.isActive = false;
        validator.deposit = 0;

        // Remove from validator list
        for (uint i = 0; i < validatorList.length; i++) {
            if (validatorList[i] == msg.sender) {
                validatorList[i] = validatorList[validatorList.length - 1];
                validatorList.pop();
                break;
            }
        }

        isRegisteredValidator[msg.sender] = false;

        // Transfer deposit back
        (bool success, ) = msg.sender.call{value: deposit}("");
        require(success, "Deposit transfer failed");

        emit ValidatorUnregistered(msg.sender);
    }

    // ============ CORE RAT FUNCTIONS ============

    /**
     * @dev Submit state commitment and potentially trigger attention test
     * @param stateRoot The L2 state root
     * @param l2BlockNumber The L2 block number
     */
    function submitStateCommitment(
        bytes32 stateRoot,
        uint256 l2BlockNumber
    ) external returns (bytes32 testId) {
        // Check if attention test should be triggered
        if (shouldTriggerAttentionTest(stateRoot)) {
            testId = triggerAttentionTest(stateRoot, l2BlockNumber);
        }

        // Emit state commitment event
        emit StateCommitmentSubmitted(stateRoot, l2BlockNumber, testId);
    }

    /**
     * @dev Determine if attention test should be triggered
     * @param stateRoot The state root to use for randomness
     * @return bool True if test should be triggered
     */
    function shouldTriggerAttentionTest(
        bytes32 stateRoot
    ) internal view returns (bool) {
        if (validatorList.length == 0) return false;

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

    /**
     * @dev Trigger attention test for a randomly selected validator
     * @param stateRoot The state root being tested
     * @param l2BlockNumber The L2 block number
     * @return testId The unique test identifier
     */
    function triggerAttentionTest(
        bytes32 stateRoot,
        uint256 l2BlockNumber
    ) internal returns (bytes32 testId) {
        require(validatorList.length > 0, "No validators registered");

        // Generate test ID
        _testCounter.increment();
        testId = keccak256(
            abi.encodePacked(
                stateRoot,
                l2BlockNumber,
                _testCounter.current(),
                block.timestamp
            )
        );

        // Select target validator
        address targetValidator = selectTargetValidator(stateRoot);

        // Create attention test
        activeTests[testId] = AttentionTest({
            testId: testId,
            stateRoot: stateRoot,
            l2BlockNumber: l2BlockNumber,
            targetValidator: targetValidator,
            startTime: block.timestamp,
            responseWindow: responseWindow,
            completed: false,
            passed: false,
            leftChild: bytes32(0),
            rightChild: bytes32(0)
        });

        // Update validator stats
        validators[targetValidator].totalTests++;
        validators[targetValidator].lastTestTime = block.timestamp;

        emit AttentionTestTriggered(testId, stateRoot, targetValidator);
    }

    /**
     * @dev Select target validator using deterministic randomness
     * @param stateRoot The state root for randomness
     * @return address The selected validator address
     */
    function selectTargetValidator(
        bytes32 stateRoot
    ) internal view returns (address) {
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

    /**
     * @dev Submit attention test solution
     * @param testId The test identifier
     * @param leftChild The left child hash
     * @param rightChild The right child hash
     */
    function submitAttentionSolution(
        bytes32 testId,
        bytes32 leftChild,
        bytes32 rightChild
    ) external onlyActiveValidator nonReentrant {
        AttentionTest storage test = activeTests[testId];

        require(test.testId != bytes32(0), "Test does not exist");
        require(test.targetValidator == msg.sender, "Not the target validator");
        require(!test.completed, "Test already completed");
        require(
            block.timestamp <= test.startTime + test.responseWindow,
            "Response window expired"
        );

        // Verify solution
        bool isValid = verifyAttentionSolution(
            leftChild,
            rightChild,
            test.stateRoot
        );

        // Update test status
        test.completed = true;
        test.passed = isValid;
        test.leftChild = leftChild;
        test.rightChild = rightChild;

        // Update validator stats
        if (isValid) {
            validators[msg.sender].passedTests++;
        } else {
            validators[msg.sender].failedTests++;
            applyPenalty(msg.sender);
        }

        emit AttentionTestCompleted(testId, msg.sender, isValid);
    }

    /**
     * @dev Verify attention test solution
     * @param leftChild The left child hash
     * @param rightChild The right child hash
     * @param expectedStateRoot The expected state root
     * @return bool True if solution is valid
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
     * @dev Hash tree function (Merkle tree hash)
     * @param left The left child
     * @param right The right child
     * @return bytes32 The computed hash
     */
    function hashTree(
        bytes32 left,
        bytes32 right
    ) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(left, right));
    }

    /**
     * @dev Apply penalty to validator
     * @param validator The validator address
     */
    function applyPenalty(address validator) internal {
        uint256 penalty = Math.min(penaltyAmount, validators[validator].deposit);

        if (penalty > 0) {
            validators[validator].deposit -= penalty;

            // Transfer penalty to treasury
            (bool success, ) = owner().call{value: penalty}("");
            require(success, "Penalty transfer failed");

            emit PenaltyApplied(validator, penalty);
        }
    }

    // ============ VIEW FUNCTIONS ============

    /**
     * @dev Get validator statistics
     * @param validator The validator address
     * @return Validator struct
     */
    function getValidator(address validator) external view returns (Validator memory) {
        return validators[validator];
    }

    /**
     * @dev Get active test details
     * @param testId The test identifier
     * @return AttentionTest struct
     */
    function getActiveTest(bytes32 testId) external view returns (AttentionTest memory) {
        return activeTests[testId];
    }

    /**
     * @dev Get all registered validators
     * @return Array of validator addresses
     */
    function getAllValidators() external view returns (address[] memory) {
        return validatorList;
    }

    /**
     * @dev Get validator count
     * @return Number of registered validators
     */
    function getValidatorCount() external view returns (uint256) {
        return validatorList.length;
    }

    // ============ ADMIN FUNCTIONS ============

    /**
     * @dev Update RAT parameters (only owner)
     */
    function updateParameters(
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
     * @dev Withdraw accumulated penalties (only owner)
     */
    function withdrawPenalties() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "No penalties to withdraw");

        (bool success, ) = owner().call{value: balance}("");
        require(success, "Withdrawal failed");
    }
}
