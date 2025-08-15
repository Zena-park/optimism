const { ethers } = require("ethers");

// RAT 시스템 기본 매개변수
const RAT_PARAMETERS = {
    // Attention Test 확률 (basis points, 100 = 1%)
    attentionTestProbability: 28,  // 0.28%

    // 응답 윈도우 (초)
    responseWindow: 300,  // 5분

    // 최소 검증자 예치금 (ETH)
    minDeposit: ethers.utils.parseEther("1"),  // 1 ETH

    // 페널티 금액 (ETH)
    penaltyAmount: ethers.utils.parseEther("0.1"),  // 0.1 ETH

    // L2 에포크 간격 (초)
    epochInterval: 600,  // 10분

    // 모니터링 간격 (밀리초)
    monitoringInterval: 30000,  // 30초
};

// 네트워크별 설정
const NETWORK_CONFIG = {
    hardhat: {
        chainId: 1337,
        url: "http://127.0.0.1:8545",
        gasPrice: ethers.utils.parseUnits("20", "gwei"),
    },
    sepolia: {
        chainId: 11155111,
        url: process.env.SEPOLIA_URL || "https://sepolia.infura.io/v3/YOUR_PROJECT_ID",
        gasPrice: ethers.utils.parseUnits("20", "gwei"),
    },
    mainnet: {
        chainId: 1,
        url: process.env.MAINNET_URL || "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        gasPrice: ethers.utils.parseUnits("20", "gwei"),
    }
};

// 시뮬레이션 설정
const SIMULATION_CONFIG = {
    // 시뮬레이션 검증자 수
    validatorCount: 10,

    // 시뮬레이션 기간 (초)
    duration: 3600,  // 1시간

    // 상태 커밋먼트 빈도 (초)
    commitmentInterval: 60,  // 1분

    // 검증자 응답 확률 (0-1)
    responseProbability: 0.9,  // 90%

    // 검증자 실패 확률 (0-1)
    failureProbability: 0.05,  // 5%
};

// 성능 모니터링 설정
const MONITORING_CONFIG = {
    // 메트릭 수집 간격 (초)
    metricsInterval: 60,

    // 로그 레벨
    logLevel: "info",  // debug, info, warn, error

    // 알림 설정
    alerts: {
        highFailureRate: 0.1,  // 10% 이상 실패 시 알림
        lowResponseRate: 0.8,  // 80% 미만 응답 시 알림
        highGasUsage: 100000,  // 10만 가스 이상 사용 시 알림
    }
};

// 보안 설정
const SECURITY_CONFIG = {
    // 최대 검증자 수
    maxValidators: 1000,

    // 최소 검증자 수
    minValidators: 3,

    // 검증자 등록 쿨다운 (초)
    registrationCooldown: 3600,  // 1시간

    // 최대 페널티 비율 (예치금 대비)
    maxPenaltyRatio: 0.5,  // 50%

    // 랜덤니스 보안
    randomness: {
        // 블록 해시 지연
        blockHashDelay: 1,

        // 추가 엔트로피 소스
        additionalEntropy: true,
    }
};

module.exports = {
    RAT_PARAMETERS,
    NETWORK_CONFIG,
    SIMULATION_CONFIG,
    MONITORING_CONFIG,
    SECURITY_CONFIG
};
