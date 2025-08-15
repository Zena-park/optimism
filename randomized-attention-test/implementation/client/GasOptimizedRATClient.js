const { ethers } = require('ethers');
const { MerkleTree } = require('merkletreejs');

/**
 * 가스비 최적화된 RAT 클라이언트
 * 논문 기반으로 기존 ORU 상태 커밋먼트 제출 과정에 RAT 통합
 */
class GasOptimizedRATClient {
    constructor(provider, contractAddress, privateKey) {
        this.provider = provider;
        this.wallet = new ethers.Wallet(privateKey, provider);
        this.contract = new ethers.Contract(
            contractAddress,
            this.getContractABI(),
            this.wallet
        );

        // 가스비 최적화 설정
        this.gasOptimization = {
            maxGasPrice: ethers.utils.parseUnits('20', 'gwei'),
            maxPriorityFee: ethers.utils.parseUnits('2', 'gwei'),
            batchSize: 10,  // 배치 처리 크기
            retryAttempts: 3
        };

        // 상태 추적
        this.pendingStateCommitments = [];
        this.lastProcessedBlock = 0;
    }

    /**
     * 컨트랙트 ABI (핵심 함수만 포함)
     */
    getContractABI() {
        return [
            "function submitStateCommitment(bytes32 stateRoot, uint256 l2BlockNumber) external returns (bytes32 testId)",
            "function submitAttentionSolution(bytes32 testId, bytes32 leftChild, bytes32 rightChild) external",
            "function registerValidator() external payable",
            "function unregisterValidator() external",
            "function getStateCommitment(uint256 l2BlockNumber) external view returns (bytes32)",
            "function getAttentionTest(bytes32 testId) external view returns (tuple(bytes32 stateRoot, address targetValidator, uint256 startTime, uint256 responseWindow, bool completed, bool passed))",
            "function getValidator(address validator) external view returns (tuple(address validator, uint256 deposit, bool isActive, uint256 score, uint256 totalPenalties))",
            "function getAllValidators() external view returns (address[])",
            "function getValidatorCount() external view returns (uint256)",
            "event StateCommitmentSubmitted(bytes32 indexed stateRoot, uint256 indexed l2BlockNumber, bytes32 testId)",
            "event AttentionTestTriggered(bytes32 indexed testId, bytes32 indexed stateRoot, address indexed targetValidator)",
            "event AttentionTestPassed(bytes32 indexed testId, address indexed validator)",
            "event AttentionTestFailed(bytes32 indexed testId, address indexed validator)"
        ];
    }

    /**
     * 가스비 최적화된 상태 커밋먼트 제출
     */
    async submitStateCommitmentOptimized(stateRoot, l2BlockNumber) {
        try {
            console.log(`🔄 Submitting state commitment: ${stateRoot} for block ${l2BlockNumber}`);

            // 가스비 최적화된 트랜잭션 설정
            const gasEstimate = await this.contract.estimateGas.submitStateCommitment(
                stateRoot,
                l2BlockNumber
            );

            const tx = await this.contract.submitStateCommitment(
                stateRoot,
                l2BlockNumber,
                {
                    gasLimit: gasEstimate.mul(120).div(100), // 20% 버퍼
                    maxFeePerGas: this.gasOptimization.maxGasPrice,
                    maxPriorityFeePerGas: this.gasOptimization.maxPriorityFee
                }
            );

            console.log(`✅ State commitment submitted: ${tx.hash}`);

            // 트랜잭션 완료 대기
            const receipt = await tx.wait();

            // Attention Test 트리거 확인
            const events = receipt.events?.filter(e => e.event === 'AttentionTestTriggered');
            if (events && events.length > 0) {
                const testId = events[0].args.testId;
                console.log(`🎯 Attention Test triggered: ${testId}`);
                return { success: true, testId, txHash: tx.hash };
            }

            return { success: true, testId: null, txHash: tx.hash };

        } catch (error) {
            console.error(`❌ Error submitting state commitment: ${error.message}`);
            return { success: false, error: error.message };
        }
    }

    /**
     * 배치 상태 커밋먼트 제출 (가스비 최적화)
     */
    async submitBatchStateCommitments(stateCommitments) {
        if (stateCommitments.length === 0) return [];

        console.log(`📦 Submitting batch of ${stateCommitments.length} state commitments`);

        const results = [];
        const batchSize = this.gasOptimization.batchSize;

        // 배치로 나누어 처리
        for (let i = 0; i < stateCommitments.length; i += batchSize) {
            const batch = stateCommitments.slice(i, i + batchSize);

            console.log(`Processing batch ${Math.floor(i/batchSize) + 1}/${Math.ceil(stateCommitments.length/batchSize)}`);

            // 배치 내에서 순차 처리 (가스비 최적화)
            for (const commitment of batch) {
                const result = await this.submitStateCommitmentOptimized(
                    commitment.stateRoot,
                    commitment.l2BlockNumber
                );
                results.push(result);

                // 가스비 최적화를 위한 짧은 대기
                await this.delay(100);
            }
        }

        console.log(`✅ Batch submission completed: ${results.filter(r => r.success).length}/${results.length} successful`);
        return results;
    }

    /**
     * 가스비 최적화된 Attention Test 응답
     */
    async submitAttentionSolutionOptimized(testId, leftChild, rightChild) {
        try {
            console.log(`🎯 Submitting attention solution for test: ${testId}`);

            // 솔루션 검증 (오프체인)
            const isValid = this.verifySolutionOffchain(leftChild, rightChild);
            if (!isValid) {
                console.warn(`⚠️ Invalid solution detected, skipping submission`);
                return { success: false, error: 'Invalid solution' };
            }

            // 가스비 최적화된 트랜잭션
            const gasEstimate = await this.contract.estimateGas.submitAttentionSolution(
                testId,
                leftChild,
                rightChild
            );

            const tx = await this.contract.submitAttentionSolution(
                testId,
                leftChild,
                rightChild,
                {
                    gasLimit: gasEstimate.mul(110).div(100), // 10% 버퍼
                    maxFeePerGas: this.gasOptimization.maxGasPrice,
                    maxPriorityFeePerGas: this.gasOptimization.maxPriorityFee
                }
            );

            console.log(`✅ Attention solution submitted: ${tx.hash}`);

            const receipt = await tx.wait();

            // 결과 확인
            const passedEvent = receipt.events?.find(e => e.event === 'AttentionTestPassed');
            const failedEvent = receipt.events?.find(e => e.event === 'AttentionTestFailed');

            if (passedEvent) {
                console.log(`🎉 Attention test passed!`);
                return { success: true, passed: true, txHash: tx.hash };
            } else if (failedEvent) {
                console.log(`❌ Attention test failed!`);
                return { success: true, passed: false, txHash: tx.hash };
            }

            return { success: true, passed: null, txHash: tx.hash };

        } catch (error) {
            console.error(`❌ Error submitting attention solution: ${error.message}`);
            return { success: false, error: error.message };
        }
    }

    /**
     * 오프체인 솔루션 검증 (가스비 절약)
     */
    verifySolutionOffchain(leftChild, rightChild) {
        try {
            // 간단한 해시 검증
            const computedHash = ethers.utils.keccak256(
                ethers.utils.defaultAbiCoder.encode(
                    ['bytes32', 'bytes32'],
                    [leftChild, rightChild]
                )
            );

            // 실제 검증은 컨트랙트에서 수행되지만, 기본 검증은 오프체인에서
            return leftChild !== ethers.constants.HashZero &&
                   rightChild !== ethers.constants.HashZero;

        } catch (error) {
            console.error(`Error in offchain verification: ${error.message}`);
            return false;
        }
    }

    /**
     * Attention Test 모니터링 (가스비 최적화)
     */
    async startAttentionTestMonitoring() {
        console.log(`👀 Starting attention test monitoring...`);

        // 이벤트 리스너 설정
        this.contract.on('AttentionTestTriggered', async (testId, stateRoot, targetValidator) => {
            console.log(`🎯 Attention test triggered: ${testId}`);
            console.log(`   State root: ${stateRoot}`);
            console.log(`   Target validator: ${targetValidator}`);

            // 현재 검증자인지 확인
            if (targetValidator.toLowerCase() === this.wallet.address.toLowerCase()) {
                console.log(`🎯 You are the target validator!`);
                await this.handleAttentionTest(testId, stateRoot);
            }
        });

        // 기존 미완료 테스트 확인
        await this.checkPendingTests();
    }

    /**
     * Attention Test 처리
     */
    async handleAttentionTest(testId, stateRoot) {
        try {
            console.log(`🔍 Processing attention test: ${testId}`);

            // L2 상태 재계산 (시뮬레이션)
            const { leftChild, rightChild } = await this.computeAttentionPuzzle(stateRoot);

            // 솔루션 제출
            const result = await this.submitAttentionSolutionOptimized(testId, leftChild, rightChild);

            if (result.success) {
                console.log(`✅ Attention test completed: ${result.passed ? 'PASSED' : 'FAILED'}`);
            } else {
                console.error(`❌ Attention test failed: ${result.error}`);
            }

        } catch (error) {
            console.error(`❌ Error handling attention test: ${error.message}`);
        }
    }

    /**
     * Attention Puzzle 계산 (시뮬레이션)
     */
    async computeAttentionPuzzle(stateRoot) {
        // 실제 구현에서는 L2 상태를 재계산하여 자식 해시를 찾아야 함
        // 여기서는 시뮬레이션용 더미 데이터 생성

        const leftChild = ethers.utils.keccak256(
            ethers.utils.defaultAbiCoder.encode(
                ['bytes32', 'uint256'],
                [stateRoot, Date.now()]
            )
        );

        const rightChild = ethers.utils.keccak256(
            ethers.utils.defaultAbiCoder.encode(
                ['bytes32', 'uint256'],
                [stateRoot, Date.now() + 1]
            )
        );

        return { leftChild, rightChild };
    }

    /**
     * 미완료 테스트 확인
     */
    async checkPendingTests() {
        try {
            const validatorCount = await this.contract.getValidatorCount();
            console.log(`📊 Total validators: ${validatorCount}`);

            // 최근 100개 블록에서 미완료 테스트 확인
            const currentBlock = await this.provider.getBlockNumber();
            const fromBlock = Math.max(0, currentBlock - 100);

            const events = await this.contract.queryFilter(
                this.contract.filters.AttentionTestTriggered(),
                fromBlock,
                currentBlock
            );

            console.log(`🔍 Found ${events.length} recent attention tests`);

            for (const event of events) {
                const testId = event.args.testId;
                const targetValidator = event.args.targetValidator;

                // 현재 검증자가 대상이고 아직 완료되지 않은 테스트 확인
                if (targetValidator.toLowerCase() === this.wallet.address.toLowerCase()) {
                    const test = await this.contract.getAttentionTest(testId);
                    if (!test.completed) {
                        console.log(`⚠️ Found pending test: ${testId}`);
                        await this.handleAttentionTest(testId, test.stateRoot);
                    }
                }
            }

        } catch (error) {
            console.error(`❌ Error checking pending tests: ${error.message}`);
        }
    }

    /**
     * 가스비 통계 수집
     */
    async getGasStatistics() {
        try {
            const currentBlock = await this.provider.getBlockNumber();
            const fromBlock = Math.max(0, currentBlock - 1000); // 최근 1000 블록

            const stateCommitmentEvents = await this.contract.queryFilter(
                this.contract.filters.StateCommitmentSubmitted(),
                fromBlock,
                currentBlock
            );

            const attentionTestEvents = await this.contract.queryFilter(
                this.contract.filters.AttentionTestTriggered(),
                fromBlock,
                currentBlock
            );

            const passedEvents = await this.contract.queryFilter(
                this.contract.filters.AttentionTestPassed(),
                fromBlock,
                currentBlock
            );

            const failedEvents = await this.contract.queryFilter(
                this.contract.filters.AttentionTestFailed(),
                fromBlock,
                currentBlock
            );

            return {
                totalStateCommitments: stateCommitmentEvents.length,
                totalAttentionTests: attentionTestEvents.length,
                passedTests: passedEvents.length,
                failedTests: failedEvents.length,
                successRate: attentionTestEvents.length > 0 ?
                    (passedEvents.length / attentionTestEvents.length * 100).toFixed(2) + '%' : '0%',
                testFrequency: (attentionTestEvents.length / stateCommitmentEvents.length * 100).toFixed(2) + '%'
            };

        } catch (error) {
            console.error(`❌ Error getting gas statistics: ${error.message}`);
            return null;
        }
    }

    /**
     * 가스비 최적화 설정 업데이트
     */
    updateGasOptimization(settings) {
        this.gasOptimization = { ...this.gasOptimization, ...settings };
        console.log(`⚙️ Gas optimization settings updated:`, this.gasOptimization);
    }

    /**
     * 유틸리티: 지연 함수
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * 클라이언트 정리
     */
    async cleanup() {
        // 이벤트 리스너 제거
        this.contract.removeAllListeners();
        console.log(`🧹 RAT client cleaned up`);
    }
}

module.exports = GasOptimizedRATClient;
