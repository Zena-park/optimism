# Phase 1: 기본 P2P 챌린저 네트워크 + RAT 시스템

## 🎯 목표

완전한 분산화된 챌린저 네트워크의 기반을 구축하여 중앙 권한 없이 챌린저들이 상호 관리할 수 있는 시스템을 개발하고, RAT (Randomized Attention Test) 프로토콜을 통합하여 챌린저들의 무임승차 문제를 해결합니다.

## 📋 작업 개요

### 핵심 기능
- **P2P 네트워크**: 챌린저 발견 및 연결
- **RAT 시스템**: 개인화된 Attention Test로 무임승차 방지
- **Enhanced Verifier**: 이중 검증 시스템 (상태 루트 + 배치 데이터)
- **L2 체인 기록**: 평판 데이터 백업 및 보상 이코노미
- **온체인 검증 우선**: 다수결 최소화 원칙

> 📖 **자세한 내용**: [RAT_INTEGRATION_PLAN.md](./RAT_INTEGRATION_PLAN.md)에서 RAT 시스템 설계 및 플로우 확인

### 기술적 요구사항
- P2P 네트워크 자동 발견
- 실시간 챌린저 상태 동기화
- 온체인 검증 결과 공유
- 기본 보안 및 인증 메커니즘
- 개인화된 Attention Test 트리거
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
│ │ RAT 시스템  │ │    │ │ RAT 시스템  │ │    │ │ RAT 시스템  │ │
│ │ - Attention │ │    │ │ - Attention │ │    │ │ - Attention │ │
│ │   Question  │ │    │ │   Question  │ │    │ │   Question  │ │
│ │ - Reputation│ │    │ │ - Reputation│ │    │ │ - Reputation│ │
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
3. **RAT 시스템**: Attention Test 트리거 및 응답
4. **이중 검증**: Enhanced Verifier를 통한 검증
5. **평판 관리**: 평판 시스템 및 L2 체인 기록
6. **상태 동기화**: 챌린저 상태 및 검증 결과 공유

> 📖 **자세한 플로우**: [RAT_INTEGRATION_PLAN.md](./RAT_INTEGRATION_PLAN.md)의 Mermaid 다이어그램 참조

## 📁 구현 파일 구조

```
op-challenger/p2p/
├── network/
│   ├── node.go              # P2P 노드 핵심 구현
│   ├── discovery.go         # DHT 기반 노드 발견
│   ├── connection.go        # 연결 관리
│   └── protocol.go          # P2P 프로토콜
├── attention/
│   ├── question.go          # Distributed Attention Question System
│   ├── validator.go         # Peer Attention Validator
│   └── tests/               # Attention 관련 테스트
├── challenger/
│   ├── engine.go            # 챌린저 엔진
│   ├── state.go             # 상태 관리
│   ├── validation.go        # 검증 로직
│   ├── attention.go         # Attention Test Handler
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

- [RAT_INTEGRATION_PLAN.md](./RAT_INTEGRATION_PLAN.md) - RAT 시스템 통합 계획
- [TODO.md](./TODO.md) - 구체적인 구현 TODO 리스트
- [optimism-challenger-systems.md](../../../optimism-challenger-systems.md)
- [optimism-challenger-management.md](../../../optimism-challenger-management.md)
- [backup-sequencer-challenger-analysis.md](../../../backup-sequencer-challenger-analysis.md)
