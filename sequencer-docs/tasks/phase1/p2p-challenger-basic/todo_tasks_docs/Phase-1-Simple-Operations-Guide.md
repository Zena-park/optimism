# 🚀 P2P 챌린저 간단 운영 가이드

## 📋 개요

이 문서는 **기존 Optimism 시스템이 이미 동작 중**인 환경에서 **P2P 챌린저만 별도로 추가**하는 방법을 설명합니다.

### 🎯 사용 시나리오
- 기존 Optimism 노드들이 정상 동작 중
- P2P 챌린저 네트워크 기능 추가
- **Fault Proof 기능**으로 실제 검증 수행
- 복잡한 설정 없이 빠르게 시작하고 싶음

### 🔧 **실행 모드**
**완전한 챌린저 모드**: P2P 네트워킹 + Fault Proof 검증
- 다른 챌린저들과의 통신
- **실제 fault proof 검증 수행**
- 게임 참여 및 클레임 처리

### 📋 **필수 구성 요소**
- **Cannon 바이너리**: fault proof 실행 엔진
- **op-program 바이너리**: 프로그램 서버
- **prestate 파일**: fault proof의 초기 상태 (19MB)
- **L2 설정 파일**: 모니터링할 L2 체인의 설정

### 🔗 **사전 작업**
**이 가이드를 사용하기 전에 다음 문서를 먼저 확인하세요:**
**[🔧 P2P 챌린저 사전 작업 가이드](./Phase-1-Prerequisites-Guide.md)**

---

## 🛠️ 빠른 시작

### 📋 사전 요구사항

**이 가이드를 사용하기 전에 다음 사전 작업을 완료해야 합니다:**

**[🔧 P2P 챌린저 사전 작업 가이드](./Phase-1-Prerequisites-Guide.md)**

이 문서에서 다음을 확인하세요:
- ✅ 바이너리 빌드 (cannon, op-program, op-challenger)
- ✅ prestate 파일 생성
- ✅ L2 설정 파일 준비
- ✅ 모든 필수 파일 존재 확인

### 🎯 필수 파일 체크리스트

다음 파일들이 모두 준비되어 있는지 확인하세요:

- [ ] `cannon/bin/cannon` (Cannon 바이너리)
- [ ] `op-program/bin/op-program` (op-program 바이너리)
- [ ] `op-program/bin/prestate-mt64Next.bin.gz` (prestate 파일)
- [ ] `op-challenger/bin/op-challenger` (챌린저 바이너리)
- [ ] `packages/contracts-bedrock/deploy-config/sepolia.json` (L2 설정 파일)

**모든 파일이 준비되었다면 아래 실행 단계로 진행하세요!**

## 🚀 실행 단계

### 1. 기본 실행 (Fault Proof 포함)

```bash
# optimism 디렉토리에서 실행

# Optimism Sepolia 테스트넷 - 완전한 기능
./op-challenger/bin/op-challenger \
    --network "sepolia" \
    --trace-type cannon \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --cannon-bin "./cannon/bin/cannon" \
    --cannon-server "./op-program/bin/op-program" \
               --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
    --mnemonic "your mnemonic phrase here" \
    --hd-path "m/44'/60'/0'/0/0" \
    --num-confirmations 1 \
    --log.level "info"

# Base Sepolia 테스트넷 (예시)
# ./op-challenger/bin/op-challenger \
#     --network "base-sepolia" \
#     --trace-type cannon \
#     --p2p-enabled \
#     --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
#     --p2p-network-id "base-challenger-sepolia" \
#     --p2p-max-peers 50 \
#     --p2p-discovery-enabled \
#     --cannon-bin "./cannon/bin/cannon" \
#     --cannon-server "./op-program/bin/op-program" \
#     --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
#     --mnemonic "your mnemonic phrase here" \
#     --hd-path "m/44'/60'/0'/0/0" \
#     --num-confirmations 1 \
#     --log.level "info"

# Arbitrum Sepolia 테스트넷 (예시)
# ./op-challenger/bin/op-challenger \
#     --network "arbitrum-sepolia" \
#     --trace-type cannon \
#     --p2p-enabled \
#     --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
#     --p2p-network-id "arbitrum-challenger-sepolia" \
#     --p2p-max-peers 50 \
#     --p2p-discovery-enabled \
#     --cannon-bin "./cannon/bin/cannon" \
#     --cannon-server "./op-program/bin/op-program" \
#     --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
#     --mnemonic "your mnemonic phrase here" \
#     --hd-path "m/44'/60'/0'/0/0" \
#     --num-confirmations 1 \
#     --log.level "info"

# 로컬 devnet (예시)
# ./op-challenger/bin/op-challenger \
#     --network "internal-devnet" \
#     --trace-type cannon \
#     --p2p-enabled \
#     --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
#     --p2p-network-id "local-challenger-devnet" \
#     --p2p-max-peers 50 \
#     --p2p-discovery-enabled \
#     --cannon-bin "./cannon/bin/cannon" \
#     --cannon-server "./op-program/bin/op-program" \
#     --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
#     --mnemonic "your mnemonic phrase here" \
#     --hd-path "m/44'/60'/0'/0/0" \
#     --num-confirmations 1 \
#     --log.level "info"
```

```bash
# 메인넷 - 완전한 기능
./op-challenger/bin/op-challenger \
    --network "mainnet" \
    --trace-type cannon \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-mainnet" \
    --p2p-max-peers 100 \
    --p2p-discovery-enabled \
    --cannon-bin "./cannon/bin/cannon" \
    --cannon-server "./op-program/bin/op-program" \
    --cannon-prestate "/opt/optimism/config/prestate-mainnet.bin.gz" \
    --mnemonic "your mnemonic phrase here" \
    --hd-path "m/44'/60'/0'/0/0" \
    --num-confirmations 1 \
    --log.level "info"
```

---

## ⚙️ 고급 설정 (선택사항)

### 1. P2P 부트노드 추가

```bash
./bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-bootnodes "/ip4/bootnode1.optimism.io/tcp/9876/p2p/12D3KooW..." \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "info"
```

### 2. 메트릭스 활성화

```bash
./bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --metrics-enabled \
    --metrics-addr "0.0.0.0" \
    --metrics-port 7300 \
    --log.level "info"
```

### 3. 데이터 디렉토리 지정

```bash
./bin/op-challenger \
    --network "sepolia" \
    --datadir "/tmp/challenger-data" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "info"
```

---

## 🔍 상태 확인

### 1. P2P 노드 상태 확인

```bash
# 로그에서 P2P 상태 확인
./bin/op-challenger --network "sepolia" --p2p-enabled --log.level "debug" 2>&1 | grep -i p2p
```

### 2. 연결된 피어 확인

```bash
# 메트릭스에서 피어 수 확인 (메트릭스 활성화 시)
curl http://localhost:7300/metrics | grep p2p_peers
```

### 3. 네트워크 상태 확인

```bash
# 네트워크 ID 및 포트 확인
netstat -tlnp | grep 9876
```

---

## 🐛 문제 해결

### 1. 포트 충돌

```bash
# 다른 포트 사용
./bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9877" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "info"
```

### 2. 네트워크 연결 실패

```bash
# 디버그 로그로 상세 확인
./bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "debug"
```

### 3. 권한 문제

```bash
# 실행 권한 확인
chmod +x ./bin/op-challenger

# 포트 권한 확인 (1024 이하 포트 사용 시)
sudo ./bin/op-challenger --p2p-listen-addr "/ip4/0.0.0.0/tcp/80" ...
```

---

## 📊 모니터링

### 1. 로그 모니터링

```bash
# 실시간 로그 확인
./bin/op-challenger --network "sepolia" --p2p-enabled --log.level "info" 2>&1 | tee challenger.log
```

### 2. 프로세스 모니터링

```bash
# 프로세스 상태 확인
ps aux | grep op-challenger

# 포트 사용 확인
lsof -i :9876
```

### 3. 메트릭스 모니터링 (활성화 시)

```bash
# 메트릭스 엔드포인트 확인
curl http://localhost:7300/metrics

# 특정 메트릭 확인
curl http://localhost:7300/metrics | grep -E "(p2p|challenger)"
```

---

## 🚀 백그라운드 실행

### 1. nohup 사용

```bash
nohup ./bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "info" > challenger.log 2>&1 &
```

### 2. systemd 서비스 (Linux)

```bash
# /etc/systemd/system/optimism-challenger.service
[Unit]
Description=Optimism P2P Challenger
After=network.target

[Service]
Type=simple
User=optimism
WorkingDirectory=/opt/optimism/op-challenger
ExecStart=/opt/optimism/op-challenger/bin/op-challenger \
    --network "sepolia" \
    --p2p-enabled \
    --p2p-listen-addr "/ip4/0.0.0.0/tcp/9876" \
    --p2p-network-id "optimism-challenger-sepolia" \
    --p2p-max-peers 50 \
    --p2p-discovery-enabled \
    --log.level "info"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 서비스 시작
sudo systemctl enable optimism-challenger
sudo systemctl start optimism-challenger
sudo systemctl status optimism-challenger
```

---

## 📝 요약

### ✅ **장점**
- **완전한 기능**: P2P 네트워킹 + Fault Proof 검증
- **간단함**: 복잡한 설정 불필요
- **빠른 시작**: 몇 줄의 명령어로 시작
- **독립성**: 기존 시스템에 영향 없음
- **실제 검증**: 실제 게임 참여 및 클레임 처리

### 🎯 **핵심 명령어**
```bash
# 기본 실행 (Fault Proof 포함)
./op-challenger/bin/op-challenger \
    --network "sepolia" \
    --trace-type cannon \
    --p2p-enabled \
    --cannon-bin "./cannon/bin/cannon" \
    --cannon-server "./op-program/bin/op-program" \
    --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
    --mnemonic "your mnemonic phrase here"

# 메트릭스 포함
./op-challenger/bin/op-challenger \
    --network "sepolia" \
    --trace-type cannon \
    --p2p-enabled \
    --metrics-enabled \
    --cannon-bin "./cannon/bin/cannon" \
    --cannon-server "./op-program/bin/op-program" \
    --cannon-prestate "./op-program/bin/prestate-mt64Next.bin.gz" \
    --mnemonic "your mnemonic phrase here"
```

### 📋 **사전 요구사항**
모든 사전 작업은 다음 문서에서 확인하세요:
**[🔧 P2P 챌린저 사전 작업 가이드](./Phase-1-Prerequisites-Guide.md)**

이제 **기존 Optimism 시스템에 P2P 챌린저를 간단하게 추가**할 수 있습니다! 🚀
