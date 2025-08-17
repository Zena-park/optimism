# TODO 3.1: 단위 테스트 설계

## 📋 작업 개요
- **작업 ID**: TODO 3.1 Unit Tests
- **작업명**: P2P 챌린저 시스템 단위 테스트 설계
- **Phase**: 1 - P2P 챌린저 네트워크 구축 및 관리
- **상태**: 📝 설계 중
- **예상 소요**: 1일

## 🎯 목표
구현된 모든 P2P 챌린저 시스템 컴포넌트들의 개별 기능을 검증하는 단위 테스트를 설계하고 구현한다.

## 📋 단위 테스트 범위

### **1. Types 패키지 단위 테스트** (`op-challenger/p2p/types/`)

#### **1.1 ChallengerInfo 테스트** (`challenger_test.go`)
```go
func TestNewChallengerInfo(t *testing.T)
func TestChallengerInfoHasRole(t *testing.T)
func TestChallengerInfoRoleBitmask(t *testing.T)
func TestChallengerInfoSerialization(t *testing.T)
func TestChallengerInfoValidation(t *testing.T)
```

**테스트 시나리오:**
- ✅ ChallengerInfo 생성 및 초기화
- ✅ 역할(Role) 비트마스크 검증
- ✅ HasRole() 메서드 동작 확인
- ✅ 챌린저 + 시퀀서 복합 역할 테스트
- ✅ JSON 직렬화/역직렬화
- ✅ 필드 유효성 검증

#### **1.2 ChallengerState 테스트** (`state_test.go`)
```go
func TestChallengerStateCreation(t *testing.T)
func TestChallengerStateClone(t *testing.T)
func TestChallengerStateStatusTransition(t *testing.T)
func TestConnectionInfoValidation(t *testing.T)
```

**테스트 시나리오:**
- ✅ ChallengerState 생성 및 초기화
- ✅ Clone() 메서드 동작 확인
- ✅ 상태 전환 로직 (Offline → Online → Active)
- ✅ ConnectionInfo 구조체 검증
- ✅ 타임스탬프 관리

#### **1.3 Config 테스트** (`config_test.go`)
```go
func TestDefaultConfigurations(t *testing.T)
func TestConfigValidation(t *testing.T)
func TestConfigSerialization(t *testing.T)
```

**테스트 시나리오:**
- ✅ 모든 기본 설정값 검증
- ✅ 설정값 유효성 검증
- ✅ 설정 직렬화/역직렬화
- ✅ 설정 범위 검증 (시간, 숫자 등)

### **2. Utils 패키지 단위 테스트** (`op-challenger/p2p/utils/`)

#### **2.1 Crypto 테스트** (`crypto_test.go`)
```go
func TestGenerateKeyPair(t *testing.T)
func TestSignMessage(t *testing.T)
func TestVerifySignature(t *testing.T)
func TestPublicKeyConversion(t *testing.T)
func TestGenerateChallengerID(t *testing.T)
```

**테스트 시나리오:**
- ✅ ECDSA 키 쌍 생성
- ✅ 메시지 서명 생성
- ✅ 서명 검증 (유효/무효)
- ✅ 공개키 ↔ 바이트 변환
- ✅ 챌린저 ID 생성 (결정적, 고유성)
- ✅ 잘못된 입력에 대한 에러 처리

#### **2.2 Validation 테스트** (`validation_test.go`)
```go
func TestValidateNodeID(t *testing.T)
func TestValidateChallengerID(t *testing.T)
func TestValidateNetworkAddress(t *testing.T)
func TestValidateHexString(t *testing.T)
func TestGenerateRequestID(t *testing.T)
```

**테스트 시나리오:**
- ✅ 노드 ID 형식 검증
- ✅ 챌린저 ID 형식 검증
- ✅ 네트워크 주소 검증 (IP:Port)
- ✅ Hex 문자열 검증
- ✅ 요청 ID 생성 및 고유성
- ✅ 경계값 테스트 (빈 문자열, 너무 긴 문자열)

### **3. Defense 패키지 단위 테스트** (`op-challenger/p2p/defense/`)

#### **3.1 RateLimiter 테스트** (`rate_limiter_test.go`)
```go
func TestRateLimiterCreation(t *testing.T)
func TestTokenBucketAlgorithm(t *testing.T)
func TestRateLimitEnforcement(t *testing.T)
func TestBucketCleanup(t *testing.T)
func TestRateLimiterStats(t *testing.T)
```

**테스트 시나리오:**
- ✅ RateLimiter 생성 및 설정
- ✅ 토큰 버킷 알고리즘 동작
- ✅ 요청 제한 및 허용 로직
- ✅ 버스트 허용 테스트
- ✅ 토큰 보충 메커니즘
- ✅ 오래된 버킷 정리
- ✅ 통계 수집 및 조회

#### **3.2 ConnectionLimiter 테스트** (`connection_limiter_test.go`)
```go
func TestConnectionLimiterCreation(t *testing.T)
func TestConnectionLimits(t *testing.T)
func TestIPBasedLimiting(t *testing.T)
func TestConnectionTracking(t *testing.T)
func TestConnectionCleanup(t *testing.T)
```

**테스트 시나리오:**
- ✅ ConnectionLimiter 생성 및 설정
- ✅ 총 연결 수 제한
- ✅ IP별 연결 수 제한
- ✅ 연결 정보 추적
- ✅ 비활성 연결 정리
- ✅ IP 추출 로직 (IPv4, IPv6)

#### **3.3 BasicMonitor 테스트** (`basic_monitor_test.go`)
```go
func TestBasicMonitorCreation(t *testing.T)
func TestTrafficRecording(t *testing.T)
func TestStatisticsCalculation(t *testing.T)
func TestAlertThresholds(t *testing.T)
func TestHistoryManagement(t *testing.T)
```

**테스트 시나리오:**
- ✅ BasicMonitor 생성 및 설정
- ✅ 메시지/연결/에러 기록
- ✅ 통계 계산 (msg/sec, 대역폭, 에러율)
- ✅ 알림 임계값 감지
- ✅ 히스토리 스냅샷 관리
- ✅ 메모리 사용량 제한

### **4. Challenger 패키지 단위 테스트** (`op-challenger/p2p/challenger/`)

#### **4.1 NetworkManager 테스트** (`network_manager_test.go`)
```go
func TestNetworkManagerCreation(t *testing.T)
func TestChallengerRegistration(t *testing.T)
func TestChallengerStatusUpdate(t *testing.T)
func TestHeartbeatProcessing(t *testing.T)
func TestPeerManagement(t *testing.T)
```

**테스트 시나리오:**
- ✅ NetworkManager 생성 및 초기화
- ✅ 챌린저 등록/해제
- ✅ 챌린저 상태 업데이트
- ✅ 하트비트 처리
- ✅ 피어 추가/제거
- ✅ 설정 검증

#### **4.2 Discovery 테스트** (`discovery_test.go`)
```go
func TestChallengerDiscoveryCreation(t *testing.T)
func TestRoleBasedDiscovery(t *testing.T)
func TestProximityBasedDiscovery(t *testing.T)
func TestDiscoveryFiltering(t *testing.T)
```

**테스트 시나리오:**
- ✅ Discovery 서비스 생성
- ✅ 역할 기반 챌린저 검색
- ✅ 근접성 기반 검색
- ✅ 검색 결과 필터링
- ✅ 검색 결과 캐싱

#### **4.3 Registry 테스트** (`registry_test.go`)
```go
func TestChallengerRegistryCreation(t *testing.T)
func TestStakeValidation(t *testing.T)
func TestChallengerValidation(t *testing.T)
func TestRegistryQueries(t *testing.T)
```

**테스트 시나리오:**
- ✅ Registry 서비스 생성
- ✅ 스테이킹 금액 검증
- ✅ 챌린저 신원 검증
- ✅ 등록 성공/실패 처리
- ✅ 등록된 챌린저 조회

#### **4.4 Monitor 테스트** (`monitor_test.go`)
```go
func TestChallengerMonitorCreation(t *testing.T)
func TestHealthMonitoring(t *testing.T)
func TestConnectionFailureDetection(t *testing.T)
func TestMonitoringStats(t *testing.T)
```

**테스트 시나리오:**
- ✅ Monitor 서비스 생성
- ✅ 챌린저 건강 상태 모니터링
- ✅ 연결 장애 감지
- ✅ 모니터링 통계 수집
- ✅ Phase 1 기본 기능만 테스트

### **5. State 패키지 단위 테스트** (`op-challenger/p2p/state/`)

#### **5.1 StateManager 테스트** (`manager_test.go`)
```go
func TestStateManagerCreation(t *testing.T)
func TestLocalStateManagement(t *testing.T)
func TestRemoteStateTracking(t *testing.T)
func TestStateConflictDetection(t *testing.T)
func TestStateCleanup(t *testing.T)
```

**테스트 시나리오:**
- ✅ StateManager 생성 및 초기화
- ✅ 로컬 상태 관리
- ✅ 원격 상태 추적
- ✅ 상태 충돌 감지
- ✅ 오래된 상태 정리
- ✅ 동시성 안전성

#### **5.2 StateSynchronizer 테스트** (`synchronizer_test.go`)
```go
func TestStateSynchronizerCreation(t *testing.T)
func TestSyncRequestHandling(t *testing.T)
func TestSyncResponseProcessing(t *testing.T)
func TestStateUpdateValidation(t *testing.T)
func TestSyncRequestCleanup(t *testing.T)
```

**테스트 시나리오:**
- ✅ StateSynchronizer 생성
- ✅ 동기화 요청 처리
- ✅ 동기화 응답 처리
- ✅ 상태 업데이트 검증
- ✅ 오래된 요청 정리
- ✅ 타임아웃 처리

#### **5.3 StateTypes 테스트** (`types_test.go`)
```go
func TestStateConflictCreation(t *testing.T)
func TestStateUpdateSerialization(t *testing.T)
func TestSyncRequestResponse(t *testing.T)
```

**테스트 시나리오:**
- ✅ StateConflict 구조체 생성
- ✅ StateUpdate 직렬화/역직렬화
- ✅ SyncRequest/Response 메시지
- ✅ 상태 업데이트 타입 검증

## 🧪 테스트 구현 전략

### **테스트 파일 구조**
```
op-challenger/p2p/
├── types/
│   ├── challenger_test.go
│   ├── state_test.go
│   └── config_test.go
├── utils/
│   ├── crypto_test.go
│   └── validation_test.go
├── defense/
│   ├── rate_limiter_test.go
│   ├── connection_limiter_test.go
│   └── basic_monitor_test.go
├── challenger/
│   ├── network_manager_test.go
│   ├── discovery_test.go
│   ├── registry_test.go
│   └── monitor_test.go
└── state/
    ├── manager_test.go
    ├── synchronizer_test.go
    └── types_test.go
```

### **테스트 유틸리티**
```go
// 공통 테스트 헬퍼 함수들
func createTestChallengerInfo() *types.ChallengerInfo
func createTestKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey)
func createTestConfig() *types.Config
func assertNoError(t *testing.T, err error, msg string)
func assertError(t *testing.T, err error, msg string)
```

### **모의 객체 (Mocks)**
```go
// 테스트용 모의 객체들
type MockNetworkManager struct {
    challengers map[string]*types.ChallengerInfo
    calls       []string
}

type MockLogger struct {
    logs []string
}

type MockTimer struct {
    currentTime time.Time
}
```

## 📊 테스트 커버리지 목표

### **패키지별 커버리지 목표**
- **types**: 95% 이상 (기본 구조체, 중요도 높음)
- **utils**: 90% 이상 (유틸리티 함수, 안정성 중요)
- **defense**: 85% 이상 (보안 기능, 다양한 시나리오)
- **challenger**: 80% 이상 (복잡한 로직, 통합 의존성)
- **state**: 80% 이상 (상태 관리, 동시성 복잡)

### **테스트 메트릭스**
- **총 테스트 수**: ~100개
- **테스트 실행 시간**: < 30초
- **메모리 사용량**: < 100MB
- **실패율**: 0% (모든 테스트 통과)

## 🔍 테스트 검증 기준

### **기능적 검증**
- ✅ 모든 공개 메서드/함수 테스트
- ✅ 정상 케이스 및 에러 케이스 모두 테스트
- ✅ 경계값 테스트 (빈 값, 최대값, 최소값)
- ✅ 입력 검증 및 에러 처리

### **품질 검증**
- ✅ 테스트 독립성 (테스트 간 의존성 없음)
- ✅ 결정적 테스트 (매번 같은 결과)
- ✅ 빠른 실행 (각 테스트 < 1초)
- ✅ 명확한 테스트 이름 및 문서화

### **보안 검증**
- ✅ 암호화 함수 정확성
- ✅ 입력 검증 우회 시도
- ✅ 메모리 누수 없음
- ✅ 예외 상황 안전 처리

## 🚀 구현 순서

### **1단계: 기본 패키지 테스트** (0.5일)
1. types 패키지 테스트
2. utils 패키지 테스트

### **2단계: 보안 시스템 테스트** (0.3일)
1. defense 패키지 테스트

### **3단계: 핵심 시스템 테스트** (0.2일)
1. challenger 패키지 테스트
2. state 패키지 테스트

## 📝 테스트 실행 및 검증

### **테스트 실행 명령어**

**프로젝트 루트에서**
```bash
cd ./optimism

# 완료된 패키지 테스트
go test ./op-challenger/p2p/types ./op-challenger/p2p/utils ./op-challenger/p2p/defense -v

# 커버리지 확인
go test -cover ./op-challenger/p2p/types ./op-challenger/p2p/utils ./op-challenger/p2p/defense

# 커버리지 리포트 생성
go test -coverprofile=coverage.out ./op-challenger/p2p/types ./op-challenger/p2p/utils ./op-challenger/p2p/defense
go tool cover -html=coverage.out
```

**p2p 디렉토리에서:**
```bash
# 완료된 패키지 테스트
go test ./types ./utils ./defense -v

# 커버리지 확인
go test -cover ./types ./utils ./defense

# 특정 패키지만 테스트
go test ./types -v
go test ./utils -v
go test ./defense -v

# 커버리지 리포트 생성
go test -coverprofile=coverage.out ./types ./utils ./defense
go tool cover -html=coverage.out
```



### **CI/CD 통합**
```yaml
# GitHub Actions 예시
- name: Run Unit Tests
  run: |
    cd op-challenger/p2p
    go test -v -cover ./types ./utils ./defense
```

## 📝 결론

이 단위 테스트 설계는 P2P 챌린저 시스템의 모든 핵심 컴포넌트를 개별적으로 검증하여 시스템의 안정성과 신뢰성을 보장한다.

각 패키지별로 체계적인 테스트를 통해:
- **기능 정확성** 검증
- **에러 처리** 검증
- **성능 및 메모리** 검증
- **보안 및 암호화** 검증

이를 통해 Phase 1의 P2P 챌린저 네트워크가 견고한 기반 위에 구축되었음을 확인할 수 있다.
