# Phase 1 P2P 챌린저 네트워크 - 문제 해결 가이드

## 📋 목차

1. [🔧 자동 해결된 문제들](#-자동-해결된-문제들)
2. [🚀 일반적인 문제 해결](#-일반적인-문제-해결)
3. [📊 진단 및 모니터링](#-진단-및-모니터링)
4. [🔍 로그 분석](#-로그-분석)
5. [⚡ 성능 최적화](#-성능-최적화)
6. [🛠️ 고급 문제 해결](#-고급-문제-해결)

---

## 🔧 자동 해결된 문제들

### ✅ 자동화된 Devnet 빌더가 해결하는 문제들

우리의 자동화된 Devnet 빌더(`build-devnet.sh`)가 다음 문제들을 자동으로 해결합니다:

#### 1. GitHub API Rate Limit 문제
**문제**: `mise install just` 실행 시 GitHub API rate limit 초과
**자동 해결**: Docker 빌드 시 `just`를 직접 다운로드하여 설치
```dockerfile
# GitHub API rate limit 문제를 우회하기 위해 just를 직접 설치
RUN wget https://github.com/casey/just/releases/download/1.37.0/just-1.37.0-x86_64-unknown-linux-musl.tar.gz && \
    tar -xzf just-1.37.0-x86_64-unknown-linux-musl.tar.gz && \
    cp just /usr/local/bin/just && \
    chmod +x /usr/local/bin/just && \
    rm just-1.37.0-x86_64-unknown-linux-musl.tar.gz && \
    just --version
```

#### 2. Go-libp2p-mplex 호환성 문제
**문제**: `go-libp2p-mplex@v0.9.0`에서 `CloseWithError`, `ResetWithError` 메서드 누락
**자동 해결**: Docker 빌드 시 자동으로 메서드 패치
```dockerfile
# Fix go-libp2p-mplex compatibility issues in Docker build
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go mod download && \
    cd /go/pkg/mod/github.com/libp2p/go-libp2p-mplex@v0.9.0 && \
    if ! grep -q "func (c *conn) CloseWithError" conn.go; then \
        sed -i '/var _ network.MuxedConn = &conn{}/a\\nfunc (c *conn) CloseWithError(code network.ConnErrorCode) error {\n\treturn c.mplex().Close()\n}' conn.go; \
    fi && \
    if ! grep -q "func (s *stream) ResetWithError" stream.go; then \
        sed -i '/var _ network.MuxedStream = &stream{}/a\\nfunc (s *stream) ResetWithError(code network.StreamErrorCode) error {\n\treturn s.mplex().Reset()\n}' stream.go; \
    fi
```

#### 3. apk/apt-get 패키지 매니저 문제
**문제**: Ubuntu 이미지에서 Alpine의 `apk` 명령어 사용
**자동 해결**: Ubuntu 이미지에 맞게 `apt-get` 사용
```dockerfile
# Also produce an op-challenger loaded with kona and asterisc using ubuntu
FROM $UBUNTU_TARGET_BASE_IMAGE AS op-challenger-target
RUN apt-get update && apt-get install -y --no-install-recommends \
    musl-tools openssl ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```

#### 4. Git 정보 전달 문제
**문제**: `just` 명령어에서 환경 변수가 제대로 전달되지 않음
**자동 해결**: 올바른 환경 변수명 사용
```bash
# GITCOMMIT, GITDATE (대문자) 사용
if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-node-image >> "$BUILD_LOG" 2>&1; then
    log_success "✅ $service 빌드 성공"
else
    log_error "❌ $service 빌드 실패"
fi
```

#### 5. Docker 빌드 캐시 문제
**문제**: Docker 캐시로 인한 빌드 실패
**자동 해결**: 캐시 관리 및 중복 방지 로직
```dockerfile
# 중복 메서드 삽입 방지
if ! grep -q "func (c *conn) CloseWithError" conn.go; then
    # 메서드 추가
fi
```

---

## 🚀 일반적인 문제 해결

### 1. 포트 충돌 문제

#### 문제 진단
```bash
# 포트 사용 현황 확인
./check-ports.sh

# 특정 포트 확인
lsof -i :8545
lsof -i :9545
lsof -i :9876
```

#### 해결 방법
```bash
# 충돌하는 프로세스 종료
lsof -ti:8545 | xargs kill -9
lsof -ti:9545 | xargs kill -9
lsof -ti:9876 | xargs kill -9

# Devnet 완전 재시작
kurtosis enclave rm --force simple-devnet
./build-devnet.sh
```

### 2. Docker 리소스 부족

#### 문제 진단
```bash
# Docker 리소스 사용량 확인
docker system df

# 컨테이너 리소스 사용량 확인
docker stats

# 시스템 리소스 확인
free -h
df -h
```

#### 해결 방법
```bash
# Docker 리소스 정리
docker system prune -a -f

# Docker 빌더 캐시 정리
docker builder prune -f

# 불필요한 이미지 정리
docker image prune -a -f

# Docker Desktop 리소스 할당 증가 (macOS)
# Docker Desktop > Settings > Resources > Advanced
```

### 3. 빌드 실패 문제

#### 문제 진단
```bash
# 빌드 로그 확인
./build-devnet.sh 2>&1 | tee build.log

# 특정 서비스 빌드 로그 확인
grep -A 10 -B 10 "op-node" build.log
grep -A 10 -B 10 "op-challenger" build.log
```

#### 해결 방법
```bash
# 완전 재빌드
docker system prune -a -f
./build-devnet.sh

# 특정 서비스만 재빌드
cd kurtosis-devnet
GITCOMMIT=$(git rev-parse HEAD) GITDATE=$(git log -1 --format=%cd) just op-node-image
```

### 4. Devnet 시작 실패

#### 문제 진단
```bash
# Devnet 상태 확인
kurtosis enclave inspect simple-devnet

# Devnet 로그 확인
kurtosis enclave logs simple-devnet

# 컨테이너 상태 확인
docker ps | grep kurtosis
```

#### 해결 방법
```bash
# Devnet 완전 재시작
kurtosis enclave rm --force simple-devnet
./build-devnet.sh

# 또는 기존 Devnet 정리 후 재시작
kurtosis enclave rm --force --all
./build-devnet.sh
```

### 5. 네트워크 연결 문제

#### 문제 진단
```bash
# 인터넷 연결 확인
ping google.com

# DNS 확인
nslookup github.com

# 포트 연결 확인
telnet localhost 8545
telnet localhost 9545
```

#### 해결 방법
```bash
# 네트워크 재설정
sudo systemctl restart network-manager

# DNS 캐시 정리
sudo systemctl restart systemd-resolved

# 방화벽 확인
sudo ufw status
```

---

## 📊 진단 및 모니터링

### 시스템 상태 진단
```bash
# 시스템 요구사항 확인
./check-system.sh

# 포트 연결 상태 확인
./check-ports.sh

# Devnet 상태 모니터링
./monitor-devnet.sh
```

### 실시간 모니터링
```bash
# Devnet 실시간 로그
kurtosis enclave logs simple-devnet --follow

# 특정 서비스 로그
kurtosis enclave logs simple-devnet --service op-node --follow
kurtosis enclave logs simple-devnet --service op-challenger --follow

# 시스템 리소스 모니터링
htop
iotop
```

### 성능 메트릭 확인
```bash
# Docker 컨테이너 성능
docker stats

# 네트워크 성능
iftop

# 디스크 I/O 성능
iostat -x 1
```

---

## 🔍 로그 분석

### 로그 수집
```bash
# 전체 로그 수집
kurtosis enclave logs simple-devnet > devnet.log 2>&1

# 특정 서비스 로그 수집
kurtosis enclave logs simple-devnet --service op-node > op-node.log 2>&1
kurtosis enclave logs simple-devnet --service op-challenger > op-challenger.log 2>&1

# 빌드 로그 수집
./build-devnet.sh > build.log 2>&1
```

### 로그 분석 명령어
```bash
# 오류 로그 필터링
grep -i error devnet.log
grep -i fail devnet.log
grep -i exception devnet.log

# 경고 로그 필터링
grep -i warn devnet.log
grep -i warning devnet.log

# 특정 패턴 검색
grep -i "connection refused" devnet.log
grep -i "timeout" devnet.log
grep -i "out of memory" devnet.log
```

### 로그 시각화
```bash
# 로그 라인 수 통계
wc -l devnet.log

# 시간별 로그 분포
grep -o "2024-[0-9][0-9]-[0-9][0-9]" devnet.log | sort | uniq -c

# 오류 발생 빈도
grep -i error devnet.log | grep -o "2024-[0-9][0-9]-[0-9][0-9]" | sort | uniq -c
```

---

## ⚡ 성능 최적화

### Docker 성능 최적화
```bash
# Docker 데몬 설정 최적화
sudo tee /etc/docker/daemon.json <<EOF
{
  "storage-driver": "overlay2",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "default-ulimits": {
    "nofile": {
      "Hard": 64000,
      "Name": "nofile",
      "Soft": 64000
    }
  }
}
EOF

# Docker 재시작
sudo systemctl restart docker
```

### 시스템 성능 최적화
```bash
# 시스템 튜닝 (Linux)
echo 'vm.max_map_count=262144' | sudo tee -a /etc/sysctl.conf
echo 'fs.file-max=65536' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 사용자 제한 설정
echo '* soft nofile 65536' | sudo tee -a /etc/security/limits.conf
echo '* hard nofile 65536' | sudo tee -a /etc/security/limits.conf
```

### 네트워크 성능 최적화
```bash
# 네트워크 버퍼 크기 증가
echo 'net.core.rmem_max=16777216' | sudo tee -a /etc/sysctl.conf
echo 'net.core.wmem_max=16777216' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_rmem=4096 87380 16777216' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_wmem=4096 65536 16777216' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

---

## 🛠️ 고급 문제 해결

### 커스텀 빌드 문제
```bash
# 특정 브랜치에서 빌드
git checkout feature/custom-branch
./build-devnet.sh

# 특정 커밋에서 빌드
git checkout <commit-hash>
./build-devnet.sh
```

### 의존성 문제 해결
```bash
# Go 모듈 정리
cd op-challenger && go mod tidy
cd ../op-node && go mod tidy

# 의존성 업데이트
go get -u ./...
go mod download
```

### 환경 변수 문제 해결
```bash
# 환경 변수 확인
env | grep -i git
env | grep -i docker

# 환경 변수 설정
export GITCOMMIT=$(git rev-parse HEAD)
export GITDATE=$(git log -1 --format=%cd)
```

### 디버깅 모드 활성화
```bash
# 상세 로그 활성화
export OP_CHALLENGER_LOG_LEVEL=debug
export OP_NODE_LOG_LEVEL=debug

# 빌드 시 상세 출력
./build-devnet.sh 2>&1 | tee build-debug.log
```

---

## 📚 추가 리소스

- [Devnet 가이드](devnet-guide.md): 상세한 Devnet 설정 및 운영 가이드
- [설치 가이드](installation-guide.md): 시스템 요구사항 및 설치 방법
- [보안 가이드](security-guide.md): 보안 고려사항 및 모범 사례
- [모니터링 가이드](monitoring-guide.md): Devnet 모니터링 및 관리

---

## 🎯 문제 해결 체크리스트

### 기본 진단
- [ ] 시스템 요구사항 충족 확인
- [ ] 포트 충돌 확인
- [ ] Docker 리소스 확인
- [ ] 네트워크 연결 확인

### 로그 분석
- [ ] Devnet 로그 확인
- [ ] 빌드 로그 확인
- [ ] 오류 패턴 분석
- [ ] 경고 메시지 확인

### 해결 시도
- [ ] 자동 해결 스크립트 실행
- [ ] 완전 재시작 시도
- [ ] 리소스 정리 수행
- [ ] 설정 재검토

### 검증
- [ ] 문제 해결 확인
- [ ] 성능 테스트 수행
- [ ] 안정성 검증
- [ ] 문서 업데이트
