# Phase 1: 기본 P2P 챌린저 네트워크

## 🎯 목표

완전한 분산화된 챌린저 네트워크의 기반을 구축하여 중앙 권한 없이 챌린저들이 상호 관리할 수 있는 시스템을 개발합니다.

## 📋 작업 개요

### 핵심 기능
- **P2P 네트워크**: 챌린저 발견 및 연결
- **개인화된 어텐션 테스트**: 무임승차 방지를 위한 개별 테스트
- **Enhanced Verifier**: 이중 검증 시스템 (상태 루트 + 배치 데이터)
- **L2 체인 기록**: 평판 데이터 백업 및 보상 이코노미
- **온체인 검증 우선**: 다수결 최소화 원칙

### 기술적 요구사항
- P2P 네트워크 자동 발견
- 실시간 챌린저 상태 동기화
- 온체인 검증 결과 공유
- 기본 보안 및 인증 메커니즘
- 개인화된 어텐션 테스트 트리거
- 이중 검증 시스템 (상태 루트 + 배치 데이터)

## 🏗️ 아키텍처 설계

### 시스템 구성
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   챌린저 A      │    │   챌린저 B      │    │   챌린저 C      │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ P2P 노드    │ │◄──►│ │ P2P 노드    │ │◄──►│ │ P2P 노드    │ │
│ │ 관리        │ │    │ │ 관리        │ │    │ │ 관리        │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│         │       │    │         │       │    │         │       │
│         ▼       │    │         ▼       │    │         ▼       │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ 어텐션 테스트│ │    │ │ 어텐션 테스트│ │    │ │ 어텐션 테스트│ │
│ │ - 개인화된  │ │    │ │ - 개인화된  │ │    │ │ - 개인화된  │ │
│ │   테스트    │ │    │ │   테스트    │ │    │ │   테스트    │ │
│ │ - 평판 기반 │ │    │ │ - 평판 기반 │ │    │ │ - 평판 기반 │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│         │       │    │         │       │    │         │       │
│         ▼       │    │         ▼       │    │         ▼       │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Enhanced    │ │    │ │ Enhanced    │ │    │ │ Enhanced    │ │
│ │ Verifier    │ │    │ │ Verifier    │ │    │ │ Verifier    │ │
│ │ (이중 검증)  │ │    │ │ (이중 검증)  │ │    │ │ (이중 검증)  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────────────┐
                    │      L2 Chain           │
                    │   - 평판 데이터 기록     │
                    │   - 보상 이코노미       │
                    └─────────────────────────┘
```

### 데이터 흐름
1. **노드 발견**: 챌린저 노드 자동 발견
2. **연결 수립**: P2P 프로토콜을 통한 직접 연결
3. **평판 관리**: 평판 시스템 및 L2 체인 기록
4. **이중 검증**: Enhanced Verifier를 통한 검증
5. **상태 동기화**: 챌린저 상태 및 검증 결과 공유

## 🚀 실행 가이드

### 로컬 Devnet에서 P2P 챌린저 실행

#### 사전 요구사항
- Kurtosis Devnet이 실행 중이어야 함
- op-challenger Docker 이미지가 빌드되어 있어야 함
- Cannon 및 op-program 바이너리가 준비되어 있어야 함

#### 자동 실행 스크립트 사용
```bash
# P2P 챌린저 실행 (자동 설정)
./op-challenger/scripts/p2p/run-challenger-devnet.sh
```

#### 스크립트 기능
- ✅ **Devnet 상태 자동 확인**
- ✅ **포트 정보 자동 감지**
- ✅ **게임 팩토리 주소 자동 감지**
- ✅ **필수 바이너리 파일 확인**
- ✅ **P2P 네트워크 설정**
- ✅ **서비스 상태 검증**

#### 수동 실행 (고급 사용자)
```bash
# 게임 팩토리 주소 확인
kurtosis files download simple-devnet op-deployer-configs /tmp/configs
GAME_FACTORY=$(grep -i "DisputeGameFactoryProxy" /tmp/configs/state.json | sed 's/.*"DisputeGameFactoryProxy": *"\([^"]*\)".*/\1/')

# 챌린저 실행
docker run -d --name op-challenger --network host \
  -v challenger-data:/data \
  -v "$(pwd)/cannon/bin:/cannon-bin:ro" \
  -v "$(pwd)/op-program/bin:/op-program-bin:ro" \
  op-challenger:devnet \
  op-challenger \
  --network=op-sepolia \
  --datadir=/data \
  --l1-eth-rpc=http://localhost:54352 \
  --l1-beacon=http://localhost:54357 \
  --l2-eth-rpc=http://localhost:54405 \
  --rollup-rpc=http://localhost:54411 \
  --game-factory-address=$GAME_FACTORY \
  --trace-type=cannon \
  --cannon-bin=/cannon-bin/cannon \
  --cannon-server=/op-program-bin/op-program \
  --cannon-prestate=/op-program-bin/prestate-mt64Next.bin.gz \
  --mnemonic="test test test test test test test test test test test junk" \
  --hd-path="m/44'/60'/0'/0/0" \
  --p2p-enabled \
  --p2p-listen-addr=/ip4/0.0.0.0/tcp/9876 \
  --p2p-network-id=optimism-challenger-devnet \
  --p2p-max-peers=20 \
  --p2p-discovery-enabled \
  --log.level=INFO
```

#### 연결 정보
- **P2P Port**: 9876
- **Network ID**: optimism-challenger-devnet
- **Max Peers**: 20
- **Discovery**: Enabled

#### 관리 명령어
```bash
# 상태 확인
docker ps | grep challenger

# 로그 확인
docker logs op-challenger

# 중지
docker stop op-challenger

# 제거
docker rm op-challenger

# 데이터 볼륨 확인
docker volume ls | grep challenger
```

## 📁 구현 파일 구조

```
op-challenger/p2p/
├── network/
│   ├── node.go              # P2P 노드 핵심 구현
│   ├── discovery.go         # DHT 기반 노드 발견
│   ├── connection.go        # 연결 관리
│   └── protocol.go          # P2P 프로토콜
├── challenger/
│   ├── engine.go            # 챌린저 엔진
│   ├── state.go             # 상태 관리
│   ├── validation.go        # 검증 로직
│   └── reputation.go        # 평판 시스템
├── reputation/
│   ├── manager.go           # Reputation Manager
│   ├── hybrid_store.go      # Hybrid Reputation Store
│   └── tests/               # 평판 관련 테스트
├── sync/
│   ├── sync.go              # 상태 동기화
│   ├── message.go           # 메시지 정의
│   └── handler.go           # 메시지 핸들러
├── l2/
│   ├── recorder.go          # L2 Chain Recorder
│   └── tests/               # L2 관련 테스트
└── security/
    ├── auth.go              # 인증 메커니즘
    ├── encryption.go        # 암호화
    └── verification.go      # 검증

op-challenger/
├── verification/
│   └── enhanced.go          # Enhanced Verifier (이중 검증)
├── batch/
│   ├── checker.go           # Batch Checker
│   └── optimization.go      # 배치 검증 성능 최적화
└── ...
```

## 🔧 구현 단계

### Phase 1: 핵심 기능 구현

#### Step 1: Distributed Attention Question System
- 기본 구조 설계 및 P2P 네트워크 구축
- 개인화된 질문 로직 및 Attention 시스템 구현
- 단위 테스트

#### Step 2: Enhanced Verifier 통합
- 이중 검증 시스템 통합
- 단위 테스트

### Phase 2: 고급 기능 구현

#### Step 3: 성능 최적화
- 병렬 검증 및 캐싱 전략
- 단위 테스트

> 📋 **구체적인 TODO**: [TODO.md](./TODO.md)에서 상세한 구현 항목 확인

## 🧪 테스트 전략

### 단위 테스트
- P2P 네트워크 컴포넌트 테스트
- 챌린저 엔진 테스트
- 동기화 시스템 테스트

### 통합 테스트
- 전체 P2P 시스템 테스트
- 다중 노드 시뮬레이션

### 네트워크 테스트
- 실제 네트워크 환경 테스트

## 📊 성능 지표

### 목표 성능
- **노드 발견 시간**: < 30초
- **상태 동기화 지연**: < 5초
- **연결 수립 시간**: < 10초
- **메시지 전송 지연**: < 1초


## 📚 참고 자료

- [RAT_INTEGRATION_PLAN.md](./RAT_INTEGRATION_PLAN.md) - Attention Test 시스템 통합 계획
- [TODO.md](./TODO.md) - 구체적인 구현 TODO 리스트
- [optimism-challenger-systems.md](../../../optimism-challenger-systems.md)
- [optimism-challenger-management.md](../../../optimism-challenger-management.md)
- [backup-sequencer-challenger-analysis.md](../../../backup-sequencer-challenger-analysis.md)
