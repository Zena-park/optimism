# Phase 1 P2P 챌린저 네트워크 - 설치 가이드

## 📋 목차

1. [🚀 빠른 설치 (추천)](#-빠른-설치-추천)
2. [💻 시스템 요구사항](#-시스템-요구사항)
3. [🔧 자동화된 설치 과정](#-자동화된-설치-과정)
4. [🎯 완전 자동화 워크플로우](#-완전-자동화-워크플로우)
5. [📊 설치 확인 및 검증](#-설치-확인-및-검증)
6. [🔧 수동 설치 (고급 사용자)](#-수동-설치-고급-사용자)

---

## 🚀 빠른 설치 (추천)

### 로컬 Devnet 개발 환경
```bash
# 1단계: 시스템 도구 설치
cd op-challenger/scripts/p2p
./install-tools.sh

# 2단계: Devnet 환경 구축
./build-devnet.sh

# 3단계: P2P 챌린저 실행
./run-challenger-devnet.sh
```

### 시스템 도구 자동 설치
```bash
# 시스템 도구만 설치
cd op-challenger/scripts/p2p
./install-tools.sh
```

**자동 설치 스크립트가 하는 일:**
- ✅ 시스템 상태 자동 확인
- ✅ 누락된 도구들 자동 설치
- ✅ Go 1.24.6+ 자동 설치
- ✅ Docker, Mise, Kurtosis, Just 자동 설치
- ✅ 설치 순서 최적화 (의존성 고려)
- ✅ 사용자 확인 후 설치 진행

---

## 💻 시스템 요구사항

### 최소 요구사항
- **OS**: Linux (Ubuntu 20.04+ 권장), macOS 12+
- **Go**: 1.24.6+ (자동 설치)
- **Docker**: 최신 버전 (자동 설치)
- **Git**: 2.0+
- **메모리**: 8GB+ RAM
- **디스크**: 50GB+ 여유 공간
- **네트워크**: 안정적인 인터넷 연결

### 권장 요구사항
- **메모리**: 16GB+ RAM
- **디스크**: 100GB+ SSD
- **네트워크**: 100Mbps+ 대역폭
- **CPU**: 8코어+

### Devnet 요구사항 (로컬 개발환경)
- **Docker**: 최신 버전 (자동 설치)
- **Docker Compose**: 최신 버전 (자동 설치)
- **Mise**: 도구 버전 관리 (자동 설치)
- **Kurtosis**: Devnet 관리 (자동 설치)
- **Just**: 빌드 도구 (자동 설치)
- **포트**: 동적 할당 (Kurtosis가 자동 관리)
- **메모리**: 8GB+ RAM (devnet + P2P 챌린저)
- **디스크**: 50GB+ 여유 공간 (devnet 데이터 포함)

---

## 🔧 자동화된 설치 과정

### 1. 시스템 요구사항 확인
자동 설치 스크립트가 다음을 확인합니다:
- OS 호환성
- 메모리 및 디스크 공간
- 기존 도구 설치 상태
- 네트워크 연결 상태

### 2. 도구 자동 설치
```bash
# Go 1.24.6+ 설치
mise install go@1.24.6
mise use go@1.24.6

# Docker 설치 (macOS)
brew install --cask docker

# Mise 설치
curl https://mise.run | sh

# Kurtosis 설치
mise install kurtosis@latest

# Just 설치
mise install just@latest
```

### 3. Devnet 자동 생성
```bash
# 우리만의 Devnet 빌더 사용
./build-devnet.sh
```

**자동 해결되는 문제들:**
- ✅ GitHub API Rate Limit: `just` 직접 설치로 우회
- ✅ Go-libp2p-mplex 호환성: Docker 빌드 시 자동 패치
- ✅ apk/apt-get 문제: Ubuntu 이미지에 맞게 자동 수정
- ✅ Git 정보 전달: 환경 변수 자동 설정
- ✅ Docker 빌드 캐시: 자동 캐시 관리

### 4. P2P 챌린저 자동 설정
```bash
# P2P 챌린저 네트워크 설정
./install-and-run.sh
```

---

## 🎯 완전 자동화 워크플로우

### 시나리오 1: 원클릭 완전 자동화 (가장 추천)
```bash
# 모든 과정을 한 번에 처리
cd op-challenger/scripts/p2p
./install-and-run.sh
```

**처리되는 과정:**
1. 시스템 도구 설치 (Go, Docker, Mise, Kurtosis, Just)
2. Devnet 빌드 및 배포
3. P2P 챌린저 설정
4. 상태 확인 및 모니터링

### 시나리오 2: 단계별 진행 (상세 제어)
```bash
# 1. 시스템 도구 설치
./install-tools.sh

# 2. 새로운 Devnet 생성
./build-devnet.sh

# 3. P2P 챌린저 설정
./install-and-run.sh
```

### 시나리오 3: 기존 Devnet 활용
```bash
# 1. 시스템 도구 설치
./install-tools.sh

# 2. 기존 Devnet에 P2P 챌린저 연결
./install-and-run.sh
```

---

## 📊 설치 확인 및 검증

### 시스템 도구 확인
```bash
# Go 버전 확인
go version
# 예상 출력: go version go1.24.6 darwin/arm64

# Docker 확인
docker --version
# 예상 출력: Docker version 24.0.7

# Mise 확인
mise --version
# 예상 출력: mise 2025.8.16 macos-arm64

# Kurtosis 확인
kurtosis version
# 예상 출력: CLI Version: 1.8.1

# Just 확인
just --version
# 예상 출력: just 1.37.0
```

### Devnet 상태 확인
```bash
# Devnet 실행 상태 확인
kurtosis enclave inspect simple-devnet

# 포트 연결 확인
./check-ports.sh

# 로그 확인
kurtosis enclave logs simple-devnet
```

### P2P 챌린저 확인
```bash
# P2P 챌린저 상태 확인
./monitor-devnet.sh

# 특정 서비스 로그 확인
kurtosis enclave logs simple-devnet --service op-challenger
```

---

## 🔧 수동 설치 (고급 사용자)

### Go 수동 설치
```bash
# Go 1.24.6+ 설치
wget https://go.dev/dl/go1.24.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Docker 수동 설치
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# macOS
brew install --cask docker
```

### Mise 수동 설치
```bash
# Mise 설치
curl https://mise.run | sh

# 환경 변수 설정
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc
source ~/.zshrc
```

### Kurtosis 수동 설치
```bash
# Kurtosis 설치
mise install kurtosis@latest

# 또는 직접 설치
curl -L https://github.com/kurtosis-tech/kurtosis-cli-release-artifacts/releases/latest/download/kurtosis-cli_linux_amd64.tar.gz | tar -xz
sudo mv kurtosis /usr/local/bin/
```

### Just 수동 설치
```bash
# Just 설치
mise install just@latest

# 또는 직접 설치
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash
```

---

## 🎯 설치 완료 후 다음 단계

### 1. P2P 챌린저 네트워크 개발 시작
```bash
# Devnet이 성공적으로 실행되면
# P2P 챌린저 네트워크 개발을 시작할 수 있습니다
```

### 2. 새로운 챌린저 노드 추가 테스트
```bash
# 여러 챌린저 노드를 추가하여
# P2P 네트워크 동작을 테스트할 수 있습니다
```

### 3. 네트워크 토폴로지 구성
```bash
# 다양한 네트워크 토폴로지를 구성하여
# 성능과 안정성을 테스트할 수 있습니다
```

### 4. 성능 테스트 및 최적화
```bash
# 네트워크 성능을 테스트하고
# 최적화 작업을 진행할 수 있습니다
```

### 5. 보안 테스트 및 강화
```bash
# 보안 테스트를 수행하고
# 네트워크 보안을 강화할 수 있습니다
```

---

## 🔧 문제 해결

### 일반적인 설치 문제

#### 1. 권한 문제
```bash
# Docker 권한 문제 해결
sudo usermod -aG docker $USER
newgrp docker
```

#### 2. 포트 충돌
```bash
# 포트 확인
./check-ports.sh

# 충돌하는 프로세스 종료
lsof -ti:8545 | xargs kill -9
```

#### 3. 리소스 부족
```bash
# 메모리 확인
free -h

# 디스크 공간 확인
df -h

# Docker 리소스 정리
docker system prune -a -f
```

#### 4. 네트워크 문제
```bash
# 인터넷 연결 확인
ping google.com

# DNS 확인
nslookup github.com
```

### 설치 로그 확인
```bash
# 설치 로그 확인
./install-tools.sh 2>&1 | tee install.log

# 오류 로그 확인
grep -i error install.log
```

---

## 📚 추가 리소스

- [Devnet 가이드](devnet-guide.md): 상세한 Devnet 설정 및 운영 가이드
- [문제 해결 가이드](troubleshooting-guide.md): 일반적인 문제 해결 방법
- [보안 가이드](security-guide.md): 보안 고려사항 및 모범 사례
- [모니터링 가이드](monitoring-guide.md): Devnet 모니터링 및 관리

---

## 🎯 개발 환경 연동

### 포크한 저장소 연동
- **저장소**: `https://github.com/Zena-park/optimism`
- **Docker 빌드**: 로컬 코드가 자동으로 Devnet에 반영
- **실시간 테스트**: 코드 변경사항이 즉시 Devnet에서 테스트 가능

### 자동화된 개발 워크플로우
1. **코드 수정**: 로컬에서 코드 변경
2. **자동 빌드**: Docker 빌드 시 자동으로 반영
3. **Devnet 테스트**: 변경사항이 즉시 Devnet에서 테스트됨
4. **반복 개발**: 빠른 반복 개발 가능
