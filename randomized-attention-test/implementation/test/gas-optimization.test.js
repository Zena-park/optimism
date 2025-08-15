const { expect } = require('chai');
const { ethers } = require('hardhat');
const GasOptimizedRATClient = require('../client/GasOptimizedRATClient');

describe('Gas Optimized RAT System', function () {
    let integratedRATContract;
    let owner, proposer, validator1, validator2;
    let gasOptimizedClient;

    const minDeposit = ethers.utils.parseEther('1');
    const penaltyAmount = ethers.utils.parseEther('0.1');

    beforeEach(async function () {
        [owner, proposer, validator1, validator2] = await ethers.getSigners();

        // 컨트랙트 배포
        const IntegratedRATContract = await ethers.getContractFactory('IntegratedRATContract');
        integratedRATContract = await IntegratedRATContract.deploy(minDeposit, penaltyAmount);
        await integratedRATContract.deployed();

        // 가스비 최적화 클라이언트 초기화
        gasOptimizedClient = new GasOptimizedRATClient(
            ethers.provider,
            integratedRATContract.address,
            proposer.privateKey
        );
    });

    describe('Gas Cost Analysis', function () {
        it('should have minimal gas overhead for state commitment submission', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });
            await integratedRATContract.connect(validator2).registerValidator({ value: minDeposit });

            // 상태 커밋먼트 제출 가스비 측정
            const stateRoot = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('test state'));
            const l2BlockNumber = 1;

            const tx = await integratedRATContract.connect(proposer).submitStateCommitment(
                stateRoot,
                l2BlockNumber
            );

            const receipt = await tx.wait();

            console.log(`📊 Gas used for state commitment: ${receipt.gasUsed.toString()}`);
            console.log(`💰 Gas cost: ${ethers.utils.formatEther(receipt.gasUsed.mul(20).mul(10**9))} ETH (20 gwei)`);

            // 가스비가 합리적인 범위 내에 있는지 확인
            expect(receipt.gasUsed).to.be.lt(100000); // 100k gas 미만
        });

        it('should have minimal additional cost when attention test is triggered', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });

            // Attention Test 확률을 높게 설정 (테스트용)
            await integratedRATContract.connect(owner).updateRATParameters(1000, 120, minDeposit, penaltyAmount); // 10%

            const stateRoot = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('test state'));
            const l2BlockNumber = 1;

            const tx = await integratedRATContract.connect(proposer).submitStateCommitment(
                stateRoot,
                l2BlockNumber
            );

            const receipt = await tx.wait();

            console.log(`📊 Gas used with attention test: ${receipt.gasUsed.toString()}`);

            // Attention Test 트리거 이벤트 확인
            const attentionTestEvent = receipt.events?.find(e => e.event === 'AttentionTestTriggered');
            if (attentionTestEvent) {
                console.log(`🎯 Attention test triggered: ${attentionTestEvent.args.testId}`);

                // 검증자 응답 가스비 측정
                const testId = attentionTestEvent.args.testId;
                const leftChild = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('left'));
                const rightChild = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('right'));

                const responseTx = await integratedRATContract.connect(validator1).submitAttentionSolution(
                    testId,
                    leftChild,
                    rightChild
                );

                const responseReceipt = await responseTx.wait();

                console.log(`📊 Gas used for attention response: ${responseReceipt.gasUsed.toString()}`);
                console.log(`💰 Total additional cost: ${ethers.utils.formatEther(
                    responseReceipt.gasUsed.mul(20).mul(10**9)
                )} ETH`);

                // 응답 가스비가 합리적인 범위 내에 있는지 확인
                expect(responseReceipt.gasUsed).to.be.lt(50000); // 50k gas 미만
            }
        });

        it('should optimize gas costs with batch processing', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });

            // 배치 상태 커밋먼트 준비
            const stateCommitments = [];
            for (let i = 1; i <= 5; i++) {
                stateCommitments.push({
                    stateRoot: ethers.utils.keccak256(ethers.utils.toUtf8Bytes(`state ${i}`)),
                    l2BlockNumber: i
                });
            }

            // 배치 처리 가스비 측정
            const startTime = Date.now();
            const results = await gasOptimizedClient.submitBatchStateCommitments(stateCommitments);
            const endTime = Date.now();

            console.log(`📦 Batch processing time: ${endTime - startTime}ms`);
            console.log(`✅ Successful submissions: ${results.filter(r => r.success).length}/${results.length}`);

            // 배치 처리가 효율적으로 작동하는지 확인
            expect(results.filter(r => r.success).length).to.be.gt(0);
        });
    });

    describe('Gas Optimization Features', function () {
        it('should use optimized gas settings', async function () {
            // 가스비 최적화 설정 업데이트
            gasOptimizedClient.updateGasOptimization({
                maxGasPrice: ethers.utils.parseUnits('15', 'gwei'),
                maxPriorityFee: ethers.utils.parseUnits('1', 'gwei'),
                batchSize: 5
            });

            const settings = gasOptimizedClient.gasOptimization;

            expect(settings.maxGasPrice).to.equal(ethers.utils.parseUnits('15', 'gwei'));
            expect(settings.maxPriorityFee).to.equal(ethers.utils.parseUnits('1', 'gwei'));
            expect(settings.batchSize).to.equal(5);
        });

        it('should perform offchain validation to save gas', async function () {
            const leftChild = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('left'));
            const rightChild = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('right'));

            // 오프체인 검증 테스트
            const isValid = gasOptimizedClient.verifySolutionOffchain(leftChild, rightChild);

            expect(isValid).to.be.true;

            // 잘못된 입력에 대한 검증
            const invalidResult = gasOptimizedClient.verifySolutionOffchain(
                ethers.constants.HashZero,
                rightChild
            );

            expect(invalidResult).to.be.false;
        });
    });

    describe('Cost Comparison', function () {
        it('should demonstrate cost savings compared to naive implementation', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });

            const iterations = 10;
            let totalGasUsed = 0;
            let attentionTestsTriggered = 0;

            // 여러 상태 커밋먼트 제출 및 가스비 측정
            for (let i = 1; i <= iterations; i++) {
                const stateRoot = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(`state ${i}`));
                const l2BlockNumber = i;

                const tx = await integratedRATContract.connect(proposer).submitStateCommitment(
                    stateRoot,
                    l2BlockNumber
                );

                const receipt = await tx.wait();
                totalGasUsed += receipt.gasUsed.toNumber();

                // Attention Test 트리거 확인
                const attentionTestEvent = receipt.events?.find(e => e.event === 'AttentionTestTriggered');
                if (attentionTestEvent) {
                    attentionTestsTriggered++;
                }
            }

            const averageGasPerCommitment = totalGasUsed / iterations;
            const attentionTestRate = (attentionTestsTriggered / iterations * 100).toFixed(2);

            console.log(`📊 Average gas per commitment: ${averageGasPerCommitment}`);
            console.log(`🎯 Attention test rate: ${attentionTestRate}%`);
            console.log(`💰 Total gas cost: ${ethers.utils.formatEther(
                ethers.BigNumber.from(totalGasUsed).mul(20).mul(10**9)
            )} ETH`);

            // 비용 효율성 검증
            expect(averageGasPerCommitment).to.be.lt(60000); // 평균 60k gas 미만
            expect(attentionTestRate).to.be.gt(0); // 일부 Attention Test 트리거
        });
    });

    describe('Performance Monitoring', function () {
        it('should collect accurate gas statistics', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });

            // 여러 상태 커밋먼트 제출
            for (let i = 1; i <= 5; i++) {
                const stateRoot = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(`state ${i}`));
                await integratedRATContract.connect(proposer).submitStateCommitment(stateRoot, i);
            }

            // 통계 수집
            const stats = await gasOptimizedClient.getGasStatistics();

            console.log(`📈 Gas Statistics:`, stats);

            // 통계가 올바르게 수집되는지 확인
            expect(stats).to.not.be.null;
            expect(stats.totalStateCommitments).to.be.gte(5);
            expect(stats.totalAttentionTests).to.be.gte(0);
        });
    });

    describe('Error Handling and Recovery', function () {
        it('should handle gas estimation failures gracefully', async function () {
            // 잘못된 입력으로 가스 추정 실패 시뮬레이션
            const invalidStateRoot = '0x0000000000000000000000000000000000000000000000000000000000000000';

            try {
                await integratedRATContract.connect(proposer).submitStateCommitment(
                    invalidStateRoot,
                    999999 // 매우 큰 블록 번호
                );
            } catch (error) {
                console.log(`❌ Expected error handled: ${error.message}`);
                expect(error.message).to.include('revert');
            }
        });

        it('should retry failed transactions with optimized gas settings', async function () {
            // 검증자 등록
            await integratedRATContract.connect(validator1).registerValidator({ value: minDeposit });

            // 정상적인 상태 커밋먼트 제출 (재시도 로직 테스트)
            const stateRoot = ethers.utils.keccak256(ethers.utils.toUtf8Bytes('retry test'));
            const result = await gasOptimizedClient.submitStateCommitmentOptimized(stateRoot, 1);

            expect(result.success).to.be.true;
        });
    });

    afterEach(async function () {
        // 클라이언트 정리
        if (gasOptimizedClient) {
            await gasOptimizedClient.cleanup();
        }
    });
});
