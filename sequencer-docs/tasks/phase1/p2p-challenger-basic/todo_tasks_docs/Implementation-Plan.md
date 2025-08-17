# Phase 1 구현 계획: P2P 챌린저 네트워크

## 📋 문서 개요
- **문서명**: Phase 1 구현 계획
- **상태**: 📝 구현 준비
- **목적**: 설계 완료된 P2P 챌린저 네트워크를 실제 Go 코드로 구현

## 📁 구현 폴더 구조 및 파일명

### **1. 기본 P2P 네트워크 확장** (`op-challenger/p2p/network/`)
기존 파일들이 이미 있으므로 확장:
```
op-challenger/p2p/network/
├── node.go          ← 확장 (기존 P2PNode 확장)
├── discovery.go     ← 확장 (기존 DiscoveryService 확장)
├── messages.go      ← 확장 (챌린저 메시지 추가)
├── transport.go     ← 확장 (기존 Transport 확장)
└── types.go         ← 확장 (기존 타입 확장)
```

**확장 내용:**
- `P2PNode` → 챌린저 특화 기능 추가
- `DiscoveryService` → 챌린저 발견 메커니즘 추가
- `MessageType` → RAT 관련 메시지 타입 추가
- `Transport` → 챌린저 통신 프로토콜 지원

### **2. 챌린저 특화 기능** (`op-challenger/p2p/challenger/`) - Phase 1 기본 기능
현재 빈 폴더이므로 새로 구현:
```
op-challenger/p2p/challenger/
├── node.go               ← 신규 (기본 ChallengerP2PNode)
├── network_manager.go    ← 신규 (기본 ChallengerNetworkManager)
├── discovery.go          ← 신규 (기본 ChallengerDiscovery)
├── registry.go           ← 신규 (기본 ChallengerRegistry)
└── tests/
    ├── node_test.go
    ├── network_manager_test.go
    ├── discovery_test.go
    └── registry_test.go
```

**Phase 1 구현 내용 (기본 기능):**
- `ChallengerP2PNode`: 기본 챌린저 노드 (P2PNode 확장)
- `ChallengerNetworkManager`: 기본 챌린저 관리 (발견, 등록만)
- `ChallengerDiscovery`: 기본 챌린저 발견 및 연결
- `ChallengerRegistry`: 기본 챌린저 등록 및 검증

**Phase 2-3에서 추가할 고급 기능:**
- `ChallengerMonitor`: 고급 건강 상태 모니터링
- `ChallengerAnalyzer`: 복잡한 챌린저 분석

### **3. 상태 관리 시스템** (`op-challenger/p2p/state/`) - Phase 1 기본 기능
새 폴더 생성:
```
op-challenger/p2p/state/
├── manager.go       ← 신규 (기본 상태 관리 - 온라인/오프라인만)
├── sync.go          ← 신규 (기본 StateSync)
└── tests/
    ├── manager_test.go
    └── sync_test.go
```

**Phase 1 구현 내용 (기본 기능):**
- `ChallengerStateManager`: 기본 상태 관리 (온라인/오프라인, 연결 상태만)
- `StateSync`: 기본 상태 동기화

**Phase 2-3에서 추가할 고급 기능:**
- `ActivityMonitor`: 복잡한 활성도 모니터링 및 패턴 분석
- `PerformanceTracker`: 성과 추적 및 트렌드 분석
- `StateHistory`: 상태 이력 관리
- `ConflictResolver`: 고급 충돌 해결

### **4. DDoS 방어 시스템** (`op-challenger/p2p/defense/`) - Phase 1 기본 기능
새 폴더 생성:
```
op-challenger/p2p/defense/
├── rate_limiter.go      ← 신규 (간단한 레이트 리미팅)
├── connection_limiter.go ← 신규 (연결 수 제한)
├── basic_monitor.go     ← 신규 (기본 트래픽 모니터링)
└── tests/
    ├── rate_limiter_test.go
    ├── connection_limiter_test.go
    └── basic_monitor_test.go
```

**Phase 1 구현 내용 (기본 기능):**
- `RateLimiter`: 간단한 레이트 리미팅 (초당 메시지 수 제한)
- `ConnectionLimiter`: 연결 수 제한
- `BasicMonitor`: 기본 트래픽 모니터링

**Phase 2-3에서 추가할 고급 기능:**
- `TrafficAnalyzer`: 복잡한 트래픽 패턴 분석
- `AttackDetector`: ML 기반 공격 탐지
- `AdaptiveDefense`: 적응형 방어 시스템
- `AttackInfoSharing`: 공격 정보 공유 네트워크

### **5. 설정 및 타입 정의** (`op-challenger/p2p/types/`)
새 폴더 생성:
```
op-challenger/p2p/types/
├── challenger.go    ← 신규 (ChallengerInfo, ChallengerRole 등)
├── state.go         ← 신규 (ChallengerState, ActivityLevel 등)
├── defense.go       ← 신규 (AttackInfo, AttackPattern 등)
├── config.go        ← 신규 (모든 Config 구조체들)
├── events.go        ← 신규 (이벤트 타입들)
└── tests/
    ├── challenger_test.go
    ├── state_test.go
    └── defense_test.go
```

**구현 내용:**
- `ChallengerInfo`: 챌린저 메타데이터 구조체
- `ChallengerState`: 포괄적 챌린저 상태 정보
- `AttackInfo`: 공격 정보 및 패턴 구조체
- 모든 설정 구조체들 (Config)

### **6. 유틸리티 및 공통 기능** (`op-challenger/p2p/utils/`)
새 폴더 생성:
```
op-challenger/p2p/utils/
├── crypto.go        ← 신규 (암호화, 서명, 검증)
├── metrics.go       ← 신규 (메트릭스 수집)
├── logger.go        ← 신규 (로깅 유틸리티)
├── validation.go    ← 신규 (데이터 검증)
└── tests/
    ├── crypto_test.go
    ├── metrics_test.go
    └── validation_test.go
```

**구현 내용:**
- 암호화 및 서명 유틸리티
- 메트릭스 수집 및 모니터링
- 로깅 및 디버깅 도구
- 데이터 검증 함수들

## 🎯 구현 순서 및 우선순위

### **Phase 1.1: 기반 타입 및 설정** (1일) ✅ **완료**
```
1. op-challenger/p2p/types/ 기본 구현 ✅
   - challenger.go (기본 ChallengerInfo, ChallengerRole) ✅
   - state.go (기본 ChallengerState - 온라인/오프라인만) ✅
   - config.go (기본 Config 구조체들) ✅

2. op-challenger/p2p/utils/ 기본 구현 ✅
   - crypto.go (기본 암호화 함수) ✅
   - validation.go (기본 검증 함수) ✅
```

**📋 구현 완료 보고서**: [Phase-1-1-Implementation-Report.md](Phase-1-1-Implementation-Report.md)

### **Phase 1.2: 기본 네트워크 확장** (1-2일) ✅ **완료**
```
3. op-challenger/p2p/network/ 확장 ✅
   - node.go (P2PNode 기본 챌린저 기능 추가) ✅
   - messages.go (기본 챌린저 메시지 타입 추가) ✅
   - discovery.go (기본 챌린저 발견 기능 추가) ✅
```

**📋 구현 완료 보고서**: [Phase-1-2-Implementation-Report.md](Phase-1-2-Implementation-Report.md)

### **Phase 1.3: 챌린저 시스템 구현** (2일) ✅ **완료**
```
4. op-challenger/p2p/challenger/ 기본 구현 ✅
   - network_manager.go (기본 ChallengerNetworkManager) ✅
   - discovery.go (기본 ChallengerDiscovery) ✅
   - registry.go (기본 ChallengerRegistry) ✅
   - monitor.go (기본 ChallengerMonitor - Phase 1 기능만) ✅
```

**📋 구현 완료 보고서**: [Phase-1-3-Implementation-Report.md](Phase-1-3-Implementation-Report.md)

### **Phase 1.4: 기본 상태 관리** (1일) ✅ **완료**
```
5. op-challenger/p2p/state/ 기본 구현 ✅
   - manager.go (기본 상태 관리 - 온라인/오프라인만) ✅
   - synchronizer.go (기본 상태 동기화) ✅
   - types.go (상태 타입 및 설정) ✅
```

**📋 구현 완료 보고서**: [Phase-1-4-Implementation-Report.md](Phase-1-4-Implementation-Report.md)

### **Phase 1.5: 기본 보안 시스템** (1일) ✅ **완료**
```
6. op-challenger/p2p/defense/ 기본 구현 ✅
   - rate_limiter.go (간단한 레이트 리미팅) ✅
   - connection_limiter.go (연결 수 제한) ✅
   - basic_monitor.go (기본 트래픽 모니터링) ✅
```

**📋 구현 완료 보고서**: [Phase-1-5-Implementation-Report.md](Phase-1-5-Implementation-Report.md)

### **Phase 1.6: 통합 및 테스트** (1-2일)
```
7. 기본 시스템 통합
   - 기본 기능 연동 테스트
   - 단위 테스트 작성
   - 기본 통합 테스트
   - 성능 검증
```

## 🔮 Phase 2-3에서 구현할 고급 기능들
```
- ActivityMonitor (복잡한 활성도 분석)
- PerformanceTracker (성과 트렌드 분석)
- TrafficAnalyzer (복잡한 트래픽 분석)
- AttackDetector (ML 기반 탐지)
- AdaptiveDefense (적응형 방어)
- ChallengerMonitor (고급 모니터링)
```

## 📋 구현 시 고려사항

### **1. 기존 코드와의 호환성**
- 기존 `op-challenger/p2p/network/` 파일들을 확장할 때 기존 기능 보존
- 기존 인터페이스와 구조체 시그니처 유지
- 하위 호환성 보장

### **2. 의존성 관리**
- 기존 Optimism 패키지 활용 (`op-service`, `op-node` 등)
- 외부 라이브러리 최소화
- Go 모듈 의존성 정리

### **3. 성능 목표 (Phase 1 기본 기능 기준)**
- 메모리 사용량: < 50MB (10,000 챌린저 기준)
- CPU 사용률: < 3% (정상 상태), < 8% (기본 방어 시)
- 동기화 지연: < 1초
- 전환 시간: < 5초

### **4. Phase 1 기능 범위 제한**
**포함되는 기능 (필수):**
- 기본 P2P 네트워크 통신
- 챌린저 발견 및 등록
- 기본 상태 관리 (온라인/오프라인)
- 간단한 레이트 리미팅 (초당 메시지 수 제한)
- 기본 연결 관리

**Phase 2-3로 연기되는 기능 (고급):**
- ML 기반 DDoS 탐지
- 복잡한 패턴 분석
- 고급 상태 분석 (활성도 패턴, 성과 트렌드)
- 적응형 방어 시스템
- 공격 정보 공유 네트워크

### **5. 테스트 전략**
- 단위 테스트: 각 함수별 기능 테스트
- 통합 테스트: 컴포넌트 간 연동 테스트
- E2E 테스트: 전체 시스템 동작 테스트
- 성능 테스트: 부하 테스트 및 벤치마크

### **5. 문서화**
- 각 패키지별 README.md 작성
- API 문서 생성 (godoc)
- 사용 예제 및 튜토리얼
- 아키텍처 다이어그램

## 🔗 관련 설계 문서

1. [TODO-1-1-P2P-Infrastructure-Analysis.md](TODO-1-1-P2P-Infrastructure-Analysis.md)
2. [TODO-1-2-Challenger-Management-System-Design.md](TODO-1-2-Challenger-Management-System-Design.md)
3. [TODO-2-1-Challenger-Node-Extension.md](TODO-2-1-Challenger-Node-Extension.md)
4. [TODO-2-2-Challenger-Discovery-Connection.md](TODO-2-2-Challenger-Discovery-Connection.md)
5. [TODO-2-3-Challenger-State-Management.md](TODO-2-3-Challenger-State-Management.md)
6. [TODO-2-4-DDoS-Defense-System.md](TODO-2-4-DDoS-Defense-System.md)

## 🚀 다음 단계

구현 계획이 확정되면 **Phase 1.1: 기반 타입 및 설정**부터 시작하여 순차적으로 구현을 진행합니다.

각 Phase별로 구현 완료 후 테스트를 거쳐 다음 단계로 진행하며, 전체 구현 완료 후 Phase 2 (기본 어텐션 테스트 구현)로 이어집니다.
