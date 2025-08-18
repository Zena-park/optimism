# Phase 1 P2P 챌린저 네트워크 - 시스템 운영 가이드

## 📋 목차

1. [개요](#-개요)
2. [시스템 요구사항](#-시스템-요구사항)
3. [챌린저 P2P 노드 설치 및 구성](#-챌린저-p2p-노드-설치-및-구성)
4. [챌린저 P2P 노드 실행](#-챌린저-p2p-노드-실행)
5. [P2P 네트워크 설정 및 구성](#-p2p-네트워크-설정-및-구성)
6. [테스트 및 검증](#-테스트-및-검증)
7. [모니터링 및 디버깅](#-모니터링-및-디버깅)
8. [트러블슈팅](#-트러블슈팅)
9. [프로덕션 배포 가이드](#-프로덕션-배포-가이드)
10. [보안 및 베스트 프랙티스](#-보안-및-베스트-프랙티스)

---

## 🎯 개요

이 문서는 **Phase 1 P2P 챌린저 네트워크**의 실제 운영을 위한 완전한 가이드입니다. LibP2P 기반의 분산 P2P 네트워크에서 챌린저 노드를 설치, 구성, 실행, 모니터링하는 모든 과정을 다룹니다.

### 주요 기능
- ✅ **LibP2P 기반 P2P 네트워킹**: 실제 분산 네트워크 통신
- ✅ **DHT 기반 피어 발견**: Kademlia DHT를 통한 자동 피어 발견
- ✅ **CLI 기반 설정**: 9개 P2P 관련 CLI 플래그 지원
- ✅ **환경변수 지원**: `OP_CHALLENGER_P2P_*` 환경변수 완전 지원
- ✅ **실시간 모니터링**: 성능 지표 및 네트워크 상태 모니터링
- ✅ **보안 방어**: Rate limiting, 연결 제한, DDoS 방어

---

## 💻 시스템 요구사항

### 최소 요구사항
- **OS**: Linux, macOS, Windows (WSL2)
- **메모리**: 4GB RAM
- **디스크**: 20GB 여유 공간
- **네트워크**: 안정적인 인터넷 연결
- **포트**: P2P 통신을 위한 TCP 포트 (기본: 9876)

### 권장 요구사항
- **OS**: Ubuntu 20.04+ / macOS 12+
- **메모리**: 8GB+ RAM
- **디스크**: 100GB+ SSD
- **네트워크**: 100Mbps+ 대역폭
- **CPU**: 4코어+ 

### 필수 의존성
- **Go**: 1.21+ 
- **Git**: 2.0+
- **Make**: 최신 버전

---

## 🏗️ 챌린저 P2P 노드 설치 및 구성

### 1. 소스코드 다운로드 및 빌드

```bash
# Optimism 저장소 클론
git clone https://github.com/ethereum-optimism/optimism.git
cd optimism

# op-challenger 빌드
cd op-challenger
go build -o bin/op-challenger ./cmd/

# 빌드 확인
./bin/op-challenger --version
```

### 2. 필수 바이너리 설치

```bash
# Cannon (Fault Proof 시스템)
git clone https://github.com/ethereum-optimism/cannon.git
cd cannon
make cannon
cp bin/cannon /usr/local/bin/

# op-program (프로그램 서버)
cd ../op-program
go build -o bin/op-program ./cmd/
cp bin/op-program /usr/local/bin/
```

### 3. 데이터 디렉토리 설정

```bash
# 챌린저 데이터 디렉토리 생성
sudo mkdir -p /opt/optimism/challenger
sudo chown $USER:$USER /opt/optimism/challenger

# 로그 디렉토리 생성
sudo mkdir -p /var/log/optimism/challenger
sudo chown $USER:$USER /var/log/optimism/challenger

# P2P 키 디렉토리 생성
mkdir -p ~/.optimism/challenger/keys
```

---

## 🚀 챌린저 P2P 노드 실행

### 1. 기본 P2P 노드 시작

#### Sepolia 테스트넷에서 P2P 활성화
```bash
./bin/op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/opt/optimism/challenger" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --cannon-bin "/usr/local/bin/cannon" \
    --cannon-server "/usr/local/bin/op-program" \
    --log.level "info"
```

#### 메인넷에서 P2P 활성화
```bash
./bin/op-challenger \
    --network "mainnet" \
    --l1-eth-rpc "https://ethereum-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://mainnet.optimism.io" \
    --rollup-rpc "https://mainnet.optimism.io" \
    --datadir "/opt/optimism/challenger" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-bootnodes "/ip4/bootnode1.optimism.io/tcp/9876/p2p/12D3KooW..." \
    --p2p-network-id "optimism-challenger-mainnet" \
    --p2p-max-peers 100 \
    --p2p-discovery-enabled \
    --cannon-bin "/usr/local/bin/cannon" \
    --cannon-server "/usr/local/bin/op-program" \
    --metrics-enabled \
    --metrics-addr "0.0.0.0" \
    --metrics-port 7300 \
    --log.level "info"
```

### 2. 환경변수를 통한 설정

#### .env 파일 생성
```bash
# /opt/optimism/challenger/.env
export OP_CHALLENGER_NETWORK="sepolia"
export OP_CHALLENGER_L1_ETH_RPC="https://ethereum-sepolia-rpc.publicnode.com"
export OP_CHALLENGER_L1_BEACON="https://ethereum-sepolia-beacon-api.publicnode.com"
export OP_CHALLENGER_L2_ETH_RPC="https://sepolia.optimism.io"
export OP_CHALLENGER_ROLLUP_RPC="https://sepolia.optimism.io"
export OP_CHALLENGER_DATADIR="/opt/optimism/challenger"

# P2P 설정
export OP_CHALLENGER_P2P_ENABLED=true
export OP_CHALLENGER_P2P_LISTEN_ADDR="/ip4/0.0.0.0/tcp/9876"
export OP_CHALLENGER_P2P_NETWORK_ID="optimism-challenger-sepolia"
export OP_CHALLENGER_P2P_MAX_PEERS=50
export OP_CHALLENGER_P2P_DISCOVERY_ENABLED=true
export OP_CHALLENGER_P2P_RATE_LIMIT=1000
export OP_CHALLENGER_P2P_CONNECTION_LIMIT=100

# 바이너리 경로
export OP_CHALLENGER_CANNON_BIN="/usr/local/bin/cannon"
export OP_CHALLENGER_CANNON_SERVER="/usr/local/bin/op-program"

# 로깅 및 메트릭
export OP_CHALLENGER_LOG_LEVEL="info"
export OP_CHALLENGER_METRICS_ENABLED=true
export OP_CHALLENGER_METRICS_ADDR="0.0.0.0"
export OP_CHALLENGER_METRICS_PORT=7300
```

#### 환경변수로 실행
```bash
# 환경변수 로드
source /opt/optimism/challenger/.env

# 챌린저 실행 (환경변수 사용)
./bin/op-challenger
```

### 3. P2P 없이 실행 (기본 모드)

```bash
# P2P 비활성화 (기본값)
./bin/op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/opt/optimism/challenger" \
    --cannon-bin "/usr/local/bin/cannon" \
    --cannon-server "/usr/local/bin/op-program"
    # --p2p-enabled=false는 기본값이므로 생략 가능
```

---

## ⚙️ P2P 네트워크 설정 및 구성

### 1. 지원되는 P2P CLI 플래그

| 플래그 | 기본값 | 설명 | 환경변수 |
|--------|--------|------|----------|
| `--p2p-enabled` | `false` | P2P 네트워킹 활성화 | `OP_CHALLENGER_P2P_ENABLED` |
| `--p2p-listen-addr` | `/ip4/0.0.0.0/tcp/9876` | LibP2P 멀티어드레스 | `OP_CHALLENGER_P2P_LISTEN_ADDR` |
| `--p2p-bootnodes` | `[]` | 부트스트랩 피어 목록 | `OP_CHALLENGER_P2P_BOOTNODES` |
| `--p2p-max-peers` | `50` | 최대 피어 연결 수 | `OP_CHALLENGER_P2P_MAX_PEERS` |
| `--p2p-network-id` | `optimism-challenger` | 네트워크 식별자 | `OP_CHALLENGER_P2P_NETWORK_ID` |
| `--p2p-private-key` | `""` | 프라이빗 키 파일 경로 | `OP_CHALLENGER_P2P_PRIVATE_KEY` |
| `--p2p-discovery-enabled` | `true` | DHT 피어 발견 활성화 | `OP_CHALLENGER_P2P_DISCOVERY_ENABLED` |
| `--p2p-rate-limit` | `1000` | 초당 메시지 제한 | `OP_CHALLENGER_P2P_RATE_LIMIT` |
| `--p2p-connection-limit` | `100` | 동시 연결 제한 | `OP_CHALLENGER_P2P_CONNECTION_LIMIT` |

### 2. 멀티어드레스 형식 가이드

```bash
# IPv4 TCP 주소
--p2p-listen-addr "/ip4/0.0.0.0/tcp/9876"
--p2p-listen-addr "/ip4/192.168.1.100/tcp/9876"

# IPv6 TCP 주소  
--p2p-listen-addr "/ip6/::/tcp/9876"
--p2p-listen-addr "/ip6/::1/tcp/9876"

# 다중 주소 (환경변수 사용)
export OP_CHALLENGER_P2P_LISTEN_ADDR="/ip4/0.0.0.0/tcp/9876,/ip6/::/tcp/9876"
```

### 3. 부트스트랩 피어 설정

```bash
# 단일 부트스트랩 피어
--p2p-bootnodes "/ip4/127.0.0.1/tcp/9877/p2p/12D3KooWExample..."

# 다중 부트스트랩 피어
--p2p-bootnodes "/ip4/bootnode1.optimism.io/tcp/9876/p2p/12D3KooW...,/ip4/bootnode2.optimism.io/tcp/9876/p2p/12D3KooW..."

# 환경변수로 설정
export OP_CHALLENGER_P2P_BOOTNODES="/ip4/bootnode1.optimism.io/tcp/9876/p2p/12D3KooW...,/ip4/bootnode2.optimism.io/tcp/9876/p2p/12D3KooW..."
```

### 4. 프라이빗 키 관리

```bash
# 키 파일 생성 (자동)
./bin/op-challenger --p2p-enabled --p2p-private-key "/opt/optimism/challenger/keys/p2p-key.pem"

# 기존 키 파일 사용
--p2p-private-key "/path/to/existing/key.pem"

# 키 파일 권한 설정
chmod 600 /opt/optimism/challenger/keys/p2p-key.pem
```

---

## 🧪 테스트 및 검증

### 1. CLI 플래그 확인

```bash
# 모든 P2P 플래그 확인
./bin/op-challenger --help | grep p2p

# 실제 출력 예시:
#   --p2p-bootnodes value                                                  ($OP_CHALLENGER_P2P_BOOTNODES)
#          List of P2P bootstrap nodes in multiaddr format
#   
#   --p2p-connection-limit value        (default: 100)                     ($OP_CHALLENGER_P2P_CONNECTION_LIMIT)
#          Maximum number of concurrent connections
#   
#   --p2p-discovery-enabled             (default: true)                    ($OP_CHALLENGER_P2P_DISCOVERY_ENABLED)
#          Enable DHT-based peer discovery
#   
#   --p2p-enabled                       (default: false)                   ($OP_CHALLENGER_P2P_ENABLED)
#          Enable P2P networking for challenger coordination
```

### 2. P2P 네트워크 테스트

#### 단일 노드 P2P 테스트
```bash
# 터미널 1: 첫 번째 노드 (부트스트랩)
./bin/op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/tmp/challenger-node1" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "test-network" \
    --cannon-bin "/usr/local/bin/cannon" \
    --cannon-server "/usr/local/bin/op-program" \
    --log.level "debug"

# 터미널 2: 두 번째 노드 (피어)
./bin/op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/tmp/challenger-node2" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9877" \
    --p2p-bootnodes "/ip4/127.0.0.1/tcp/9876/p2p/[NODE1_PEER_ID]" \
    --p2p-network-id "test-network" \
    --cannon-bin "/usr/local/bin/cannon" \
    --cannon-server "/usr/local/bin/op-program" \
    --log.level "debug"
```

### 3. 개발 테스트 스위트

#### LibP2P 통합 테스트
```bash
cd op-challenger

# LibP2P 기본 통합 테스트
go test ./p2p/network/ -v -run TestLibP2P

# 예상 출력:
# === RUN   TestLibP2PNodeBasicIntegration
# --- PASS: TestLibP2PNodeBasicIntegration (1.05s)
# === RUN   TestLibP2PNodeMessageRouter
# --- PASS: TestLibP2PNodeMessageRouter (0.01s)
# === RUN   TestLibP2PNodeDiscovery
# --- PASS: TestLibP2PNodeDiscovery (1.01s)
```

#### P2P 성능 테스트
```bash
# 메시지 처리량 테스트
go test ./p2p/integration/ -v -run TestMessageThroughput

# 예상 출력:
# === RUN   TestMessageThroughput/MessageProcessingThroughput
#     performance_integration_test.go:55: Message throughput: 2,710,332.66 messages/second
# === RUN   TestMessageThroughput/ChallengerRegistrationThroughput
#     performance_integration_test.go:87: Challenger registration throughput: 571,877.90 registrations/second
```

#### 전체 테스트 스위트
```bash
# 모든 P2P 관련 테스트 실행
go test ./p2p/types ./p2p/utils ./p2p/defense ./p2p/challenger ./p2p/state ./p2p/network ./p2p/integration -v

# E2E 테스트 (시간이 오래 걸림)
go test ./p2p/integration/ -v -timeout 10m
```

---

## 📊 모니터링 및 디버깅

### 1. 메트릭 모니터링 설정

#### 메트릭 서버 활성화
```bash
./bin/op-challenger \
    [기본 플래그들...] \
    --metrics-enabled \
    --metrics-addr "0.0.0.0" \
    --metrics-port 7300 \
    --p2p-enabled
```

#### 메트릭 엔드포인트 확인
```bash
# 기본 메트릭
curl http://localhost:7300/metrics

# P2P 관련 메트릭 필터링
curl http://localhost:7300/metrics | grep p2p

# 디버그 정보
curl http://localhost:7300/debug/vars
```

### 2. P2P 네트워크 진단

#### 피어 연결 상태 확인
```bash
# 연결된 피어 목록
curl http://localhost:7300/debug/peers

# DHT 라우팅 테이블
curl http://localhost:7300/debug/dht

# 메시지 통계
curl http://localhost:7300/debug/messages

# 방어 시스템 상태
curl http://localhost:7300/debug/defense
```

### 3. 로깅 및 디버깅

#### 상세 로그 활성화
```bash
# LibP2P 디버그 로그
export GOLOG_LOG_LEVEL=debug
export GOLOG_LOG_FILE=/var/log/optimism/challenger/libp2p.log

# 챌린저 디버그 로그
./bin/op-challenger \
    [기본 플래그들...] \
    --log.level debug \
    --log.format json \
    > /var/log/optimism/challenger/challenger.log 2>&1
```

#### 로그 분석
```bash
# P2P 관련 로그 필터링
grep -i "p2p\|libp2p" /var/log/optimism/challenger/challenger.log

# 에러 로그 확인
grep -i "error\|failed" /var/log/optimism/challenger/challenger.log

# 연결 로그 확인
grep -i "peer\|connect\|disconnect" /var/log/optimism/challenger/libp2p.log
```

### 4. 성능 프로파일링

#### 메모리 프로파일링
```bash
# 메모리 사용량 프로파일
go test ./p2p/integration/ -memprofile=challenger-mem.prof

# 프로파일 분석
go tool pprof challenger-mem.prof
```

#### CPU 프로파일링
```bash
# CPU 사용량 프로파일
go test ./p2p/integration/ -cpuprofile=challenger-cpu.prof

# 프로파일 분석
go tool pprof challenger-cpu.prof
```

---

## 🔧 트러블슈팅

### 1. 일반적인 문제 및 해결책

#### 문제: P2P 노드가 시작되지 않음
```bash
# 에러: "failed to create libp2p host"
# 해결: 포트 충돌 확인
sudo netstat -tulpn | grep 9876

# 다른 포트 사용
--p2p-listen-addr "/ip4/0.0.0.0/tcp/9877"
```

#### 문제: 피어 발견이 안됨
```bash
# 에러: "no peers found in DHT"
# 해결: 부트스트랩 피어 확인
--p2p-bootnodes "/ip4/valid-bootnode.example.com/tcp/9876/p2p/12D3KooW..."

# DHT 비활성화 후 수동 피어 연결
--p2p-discovery-enabled=false
```

#### 문제: 연결이 자주 끊어짐
```bash
# 해결: 연결 제한 증가
--p2p-max-peers 100
--p2p-connection-limit 200

# 네트워크 안정성 확인
ping bootnode.example.com
```

### 2. 로그 에러 해석

#### LibP2P 관련 에러
```bash
# "failed to dial peer": 피어 연결 실패
# -> 네트워크 연결 또는 피어 주소 확인

# "context deadline exceeded": 타임아웃
# -> 네트워크 지연 또는 피어 응답 없음

# "resource limit exceeded": 리소스 한계
# -> --p2p-max-peers, --p2p-connection-limit 조정
```

#### DHT 관련 에러
```bash
# "bootstrap timeout": 부트스트랩 실패
# -> 부트스트랩 피어 주소 및 연결 확인

# "routing table empty": 라우팅 테이블 비어있음
# -> 충분한 시간 대기 또는 부트스트랩 피어 추가
```

### 3. 네트워크 진단 도구

#### 연결성 테스트
```bash
# 포트 연결 테스트
telnet bootnode.example.com 9876

# LibP2P 피어 ID 확인
./bin/op-challenger --p2p-enabled --datadir /tmp/test 2>&1 | grep "peer ID"
```

#### 성능 테스트
```bash
# 네트워크 대역폭 테스트
iperf3 -c bootnode.example.com -p 9876

# 지연시간 테스트
ping -c 10 bootnode.example.com
```

---

## 🚀 프로덕션 배포 가이드

### 1. 시스템 서비스 설정

#### systemd 서비스 파일 생성
```bash
# /etc/systemd/system/op-challenger.service
[Unit]
Description=Optimism Challenger with P2P
After=network.target

[Service]
Type=simple
User=optimism
Group=optimism
WorkingDirectory=/opt/optimism/challenger
ExecStart=/opt/optimism/challenger/bin/op-challenger
EnvironmentFile=/opt/optimism/challenger/.env
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# 리소스 제한
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

#### 서비스 활성화
```bash
# 서비스 파일 리로드
sudo systemctl daemon-reload

# 서비스 활성화 및 시작
sudo systemctl enable op-challenger
sudo systemctl start op-challenger

# 상태 확인
sudo systemctl status op-challenger

# 로그 확인
sudo journalctl -u op-challenger -f
```

### 2. 프로덕션 환경변수

#### /opt/optimism/challenger/.env (프로덕션)
```bash
# 네트워크 설정 (메인넷)
OP_CHALLENGER_NETWORK="mainnet"
OP_CHALLENGER_L1_ETH_RPC="https://ethereum-rpc.publicnode.com"
OP_CHALLENGER_L1_BEACON="https://ethereum-beacon-api.publicnode.com"
OP_CHALLENGER_L2_ETH_RPC="https://mainnet.optimism.io"
OP_CHALLENGER_ROLLUP_RPC="https://mainnet.optimism.io"

# 데이터 및 키 경로
OP_CHALLENGER_DATADIR="/opt/optimism/challenger/data"
OP_CHALLENGER_P2P_PRIVATE_KEY="/opt/optimism/challenger/keys/p2p-mainnet.pem"

# P2P 프로덕션 설정
OP_CHALLENGER_P2P_ENABLED=true
OP_CHALLENGER_P2P_LISTEN_ADDR="/ip4/0.0.0.0/tcp/9876"
OP_CHALLENGER_P2P_BOOTNODES="[MAINNET_BOOTNODES]"
OP_CHALLENGER_P2P_NETWORK_ID="optimism-challenger-mainnet"
OP_CHALLENGER_P2P_MAX_PEERS=100
OP_CHALLENGER_P2P_DISCOVERY_ENABLED=true
OP_CHALLENGER_P2P_RATE_LIMIT=2000
OP_CHALLENGER_P2P_CONNECTION_LIMIT=200

# 바이너리 경로
OP_CHALLENGER_CANNON_BIN="/usr/local/bin/cannon"
OP_CHALLENGER_CANNON_SERVER="/usr/local/bin/op-program"

# 모니터링
OP_CHALLENGER_METRICS_ENABLED=true
OP_CHALLENGER_METRICS_ADDR="127.0.0.1"
OP_CHALLENGER_METRICS_PORT=7300

# 로깅
OP_CHALLENGER_LOG_LEVEL="info"
OP_CHALLENGER_LOG_FORMAT="json"
```

### 3. 방화벽 설정

#### iptables 규칙
```bash
# P2P 포트 열기
sudo iptables -A INPUT -p tcp --dport 9876 -j ACCEPT

# 메트릭 포트 (로컬만)
sudo iptables -A INPUT -p tcp --dport 7300 -s 127.0.0.1 -j ACCEPT

# 변경사항 저장
sudo iptables-save > /etc/iptables/rules.v4
```

#### ufw 설정
```bash
# P2P 포트 허용
sudo ufw allow 9876/tcp

# 메트릭 포트 (로컬만)
sudo ufw allow from 127.0.0.1 to any port 7300
```

### 4. 로그 로테이션

#### /etc/logrotate.d/op-challenger
```bash
/var/log/optimism/challenger/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 optimism optimism
    postrotate
        systemctl reload op-challenger
    endscript
}
```

---

## 🔐 보안 및 베스트 프랙티스

### 1. 보안 설정

#### 프라이빗 키 보안
```bash
# 키 파일 권한 제한
chmod 600 /opt/optimism/challenger/keys/*.pem
chown optimism:optimism /opt/optimism/challenger/keys/*.pem

# 키 백업 (암호화)
gpg --symmetric --cipher-algo AES256 p2p-mainnet.pem
```

#### 네트워크 보안
```bash
# Rate limiting 활성화
--p2p-rate-limit 1000

# 연결 제한
--p2p-connection-limit 100

# 알려진 악의적 피어 차단 (설정 파일 사용)
```

### 2. 모니터링 베스트 프랙티스

#### Prometheus 메트릭 수집
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'op-challenger'
    static_configs:
      - targets: ['localhost:7300']
    scrape_interval: 15s
```

#### 알람 설정
```yaml
# alertmanager 규칙
groups:
  - name: op-challenger
    rules:
      - alert: ChallengerDown
        expr: up{job="op-challenger"} == 0
        for: 1m
        annotations:
          summary: "Challenger is down"
      
      - alert: LowPeerCount  
        expr: libp2p_peers_connected < 5
        for: 5m
        annotations:
          summary: "Low peer count: {{ $value }}"
```

### 3. 백업 및 복구

#### 데이터 백업
```bash
# 일일 백업 스크립트
#!/bin/bash
BACKUP_DIR="/backup/optimism/challenger"
DATA_DIR="/opt/optimism/challenger/data"

# 챌린저 정지
systemctl stop op-challenger

# 데이터 백업
rsync -av $DATA_DIR/ $BACKUP_DIR/$(date +%Y%m%d)/

# 챌린저 재시작
systemctl start op-challenger
```

#### 복구 절차
```bash
# 서비스 정지
systemctl stop op-challenger

# 데이터 복구
rsync -av /backup/optimism/challenger/20240101/ /opt/optimism/challenger/data/

# 권한 복구
chown -R optimism:optimism /opt/optimism/challenger/

# 서비스 재시작
systemctl start op-challenger
```

### 4. 성능 최적화

#### 시스템 튜닝
```bash
# /etc/sysctl.d/99-optimism.conf
# 네트워크 버퍼 크기 증가
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 65536 134217728
net.ipv4.tcp_wmem = 4096 65536 134217728

# 파일 디스크립터 한계 증가
fs.file-max = 1000000

# 설정 적용
sudo sysctl -p /etc/sysctl.d/99-optimism.conf
```

#### Go 런타임 최적화
```bash
# 환경변수 추가 (.env)
GOMAXPROCS=4
GOMEMLIMIT=6GB
GOGC=100
```

---

## 📞 지원 및 문의

### 문제 신고
- **GitHub Issues**: https://github.com/ethereum-optimism/optimism/issues
- **Discord**: Optimism 개발자 커뮤니티

### 추가 자료
- **Optimism 문서**: https://docs.optimism.io/
- **LibP2P 문서**: https://docs.libp2p.io/
- **Phase 1 구현 보고서**: `Phase-1-Implementation.md`

---

**Phase 1 P2P 챌린저 네트워크**는 이제 완전히 운영 준비가 완료되었습니다! 🎉

이 가이드를 통해 실제 프로덕션 환경에서 LibP2P 기반의 분산 챌린저 네트워크를 성공적으로 운영할 수 있습니다.