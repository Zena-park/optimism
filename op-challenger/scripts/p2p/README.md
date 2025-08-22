# Phase 1 P2P 챌린저 네트워크

## 🚀 빠른 시작 가이드

### 🎯 로컬 Devnet 개발 환경 (가장 추천)
```bash
# 1단계: Devnet 환경 구축
cd op-challenger/scripts/p2p
./build-devnet.sh

# 2단계: P2P 챌린저 실행
./run-challenger-devnet.sh
```

### 📋 다중 네트워크 지원 (테스트넷/메인넷)
```bash
# Sepolia, Mainnet, Goerli 등 다양한 네트워크 지원
cd op-challenger/scripts/p2p
./run-challenger-multi.sh
```

### 📋 시스템 도구 설치
```bash
# 시스템 상태 확인
cd op-challenger/scripts/p2p
./check-system.sh

# 도구 자동 설치
./install-tools.sh
```

## 📁 스크립트 목록

### 🔧 시스템 관리 스크립트
- **`check-system.sh`**: 시스템 요구사항 확인 및 진단
- **`install-tools.sh`**: Go, Docker, Mise, Kurtosis, Just 자동 설치

### 🚀 Devnet 환경 스크립트
- **`build-devnet.sh`**: **Devnet 환경 구축** - Optimism의 simple-devnet을 대체하는 강화된 Devnet 빌더 (의존성 문제 자동 해결)

### 🎯 챌린저 실행 스크립트
- **`run-challenger-devnet.sh`**: **로컬 Devnet 전용** - 개발 및 테스트 목적의 P2P 챌린저 실행
- **`run-challenger-multi.sh`**: **다중 네트워크 지원** - Sepolia, Mainnet, Goerli 등 다양한 네트워크 지원

### 📊 관리 스크립트
- **`manage.sh`**: 챌린저 관리 (시작/중지/상태확인)
- **`quick-test.sh`**: 빠른 테스트

## 🔄 워크플로우

### 시나리오 1: 로컬 Devnet 개발 (가장 추천)
```bash
# 1. Devnet 환경 구축
./build-devnet.sh

# 2. P2P 챌린저 실행
./run-challenger-devnet.sh
```

### 시나리오 2: 다중 네트워크 지원
```bash
# Sepolia, Mainnet, Goerli 등 다양한 네트워크
./run-challenger-multi.sh
```

### 시나리오 3: 시스템 도구 설치
```bash
# 시스템 도구 설치
./install-tools.sh
```

## ✅ 해결된 문제들

### 🔧 자동 해결된 기술적 문제들
- **GitHub API Rate Limit**: `just` 직접 설치로 우회
- **Go-libp2p-mplex 호환성**: Docker 빌드 시 자동 패치
- **apk/apt-get 문제**: Ubuntu 이미지에 맞게 자동 수정
- **Git 정보 전달**: 환경 변수 자동 설정
- **Docker 빌드 캐시**: 자동 캐시 관리

### 🚀 현재 환경의 장점
- **완전 자동화**: 수동 설정 불필요
- **의존성 자동 해결**: 호환성 문제 자동 패치
- **개발 환경 연동**: 포크한 저장소 코드가 Devnet에 자동 반영
- **빠른 배포**: 5-15분 내 완전한 Devnet 구축

## ⚠️ 주의사항

### Devnet 상태 확인
```bash
# Devnet 실행 상태 확인
kurtosis enclave inspect simple-devnet

# 챌린저 상태 확인
docker ps | grep challenger

# 챌린저 로그 확인
docker logs op-challenger
```

### 개발 환경 연동
- **포크한 저장소**: `https://github.com/Zena-park/optimism`
- **Docker 빌드**: 로컬 코드가 자동으로 Devnet에 반영
- **실시간 테스트**: 코드 변경사항이 즉시 Devnet에서 테스트 가능

## 📚 추가 문서
- [Devnet 가이드](docs/devnet-guide.md): 상세한 Devnet 설정 및 운영 가이드
- [설치 가이드](docs/installation-guide.md): 시스템 요구사항 및 설치 방법
- [문제 해결 가이드](docs/troubleshooting-guide.md): 일반적인 문제 해결 방법
- [보안 가이드](docs/security-guide.md): 보안 고려사항 및 모범 사례
- [모니터링 가이드](docs/monitoring-guide.md): Devnet 모니터링 및 관리
