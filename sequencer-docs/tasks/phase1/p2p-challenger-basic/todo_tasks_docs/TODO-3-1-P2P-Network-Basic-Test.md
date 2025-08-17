# TODO 3.1: P2P 네트워크 기본 테스트 설계

## 📋 작업 개요
- **작업 ID**: TODO 3.1
- **작업명**: P2P 네트워크 기본 테스트
- **Phase**: 1 - P2P 챌린저 네트워크 구축 및 관리
- **상태**: 📝 설계 중
- **예상 소요**: 1-2일

## 🎯 목표
구현된 P2P 챌린저 네트워크의 기본 기능들이 정상적으로 동작하는지 검증하고, 기존 `op-challenger` 시스템과의 통합 테스트를 수행한다.

## 📋 테스트 범위

### **1. 챌린저 노드 생성 및 연결 테스트**

#### **1.1 노드 생성 테스트**
```go
// 테스트 파일: op-challenger/p2p/network/tests/node_test.go
func TestChallengerNodeCreation(t *testing.T)
func TestChallengerNodeConfiguration(t *testing.T)
func TestChallengerNodeInitialization(t *testing.T)
```

**테스트 시나리오:**
- ✅ 챌린저 노드 생성 및 초기화
- ✅ 챌린저 정보 설정 및 검증
- ✅ 챌린저 역할(Role) 설정 테스트
- ✅ 공개키/개인키 생성 및 서명 검증
- ✅ 챌린저 ID 생성 및 유효성 검증

#### **1.2 노드 연결 테스트**
```go
// 테스트 파일: op-challenger/p2p/network/tests/connection_test.go
func TestChallengerNodeConnection(t *testing.T)
func TestMultipleNodeConnections(t *testing.T)
func TestConnectionFailureHandling(t *testing.T)
```

**테스트 시나리오:**
- ✅ 두 챌린저 노드 간 연결 설정
- ✅ 다중 노드 연결 (3-5개 노드)
- ✅ 연결 실패 시 재연결 메커니즘
- ✅ 연결 타임아웃 처리
- ✅ 연결 품질 모니터링

### **2. 챌린저 발견 프로토콜 테스트**

#### **2.1 기본 발견 테스트**
```go
// 테스트 파일: op-challenger/p2p/network/tests/discovery_test.go
func TestChallengerDiscovery(t *testing.T)
func TestChallengerDiscoveryByRole(t *testing.T)
func TestChallengerDiscoveryByProximity(t *testing.T)
```

**테스트 시나리오:**
- ✅ 챌린저 노드 자동 발견
- ✅ 역할 기반 챌린저 발견 (Challenger vs Sequencer)
- ✅ 근접성 기반 챌린저 발견
- ✅ 부트스트랩 노드를 통한 네트워크 참여
- ✅ DHT 기반 분산 발견

#### **2.2 발견 서비스 확장 테스트**
```go
func TestChallengerDiscoveryService(t *testing.T)
func TestDiscoveryServiceIntegration(t *testing.T)
```

**테스트 시나리오:**
- ✅ `DiscoveryService`의 챌린저 기능 확장
- ✅ 챌린저 노드 추가/제거 동작
- ✅ 챌린저 상태 업데이트 반영
- ✅ 오래된 챌린저 노드 정리

### **3. 네트워크 연결 안정성 테스트**

#### **3.1 연결 안정성 테스트**
```go
// 테스트 파일: op-challenger/p2p/network/tests/stability_test.go
func TestNetworkStability(t *testing.T)
func TestNodeFailureRecovery(t *testing.T)
func TestNetworkPartition(t *testing.T)
```

**테스트 시나리오:**
- ✅ 연결 안정성 테스트 (5분간 연결 유지)
- ✅ 노드 장애 시 자동 복구 (30초 내)
- ✅ 네트워크 분할 상황 처리 (1분 시나리오)
- ✅ 노드 재시작 시 자동 재연결 (30초 내)
- ✅ 메모리 누수 및 리소스 정리 (5분 모니터링)

#### **3.2 부하 테스트**
```go
func TestNetworkLoadHandling(t *testing.T)
func TestHighConnectionCount(t *testing.T)
```

**테스트 시나리오:**
- ✅ 다수 노드 동시 연결 (50-100개)
- ✅ 높은 메시지 처리량 테스트
- ✅ CPU 및 메모리 사용량 모니터링
- ✅ 네트워크 대역폭 사용량 측정

### **4. 메시지 전송 및 수신 테스트**

#### **4.1 기본 메시지 테스트**
```go
// 테스트 파일: op-challenger/p2p/network/tests/messages_test.go
func TestChallengerMessages(t *testing.T)
func TestMessageSerialization(t *testing.T)
func TestMessageValidation(t *testing.T)
```

**테스트 시나리오:**
- ✅ 챌린저 Hello/Goodbye 메시지
- ✅ 하트비트 메시지 전송/수신
- ✅ 상태 업데이트 메시지
- ✅ 피어 발견 메시지
- ✅ 메시지 직렬화/역직렬화

#### **4.2 메시지 신뢰성 테스트**
```go
func TestMessageReliability(t *testing.T)
func TestMessageOrdering(t *testing.T)
func TestMessageDuplication(t *testing.T)
```

**테스트 시나리오:**
- ✅ 메시지 전송 신뢰성 (100% 전달 보장)
- ✅ 메시지 순서 보장
- ✅ 중복 메시지 처리
- ✅ 메시지 손실 감지 및 재전송
- ✅ 메시지 타임아웃 처리

## 🔗 기존 시스템 통합 테스트

### **5. op-challenger 통합 테스트**

#### **5.1 기존 시스템 호환성**
```go
// 테스트 파일: op-challenger/p2p/tests/integration_test.go
func TestOpChallengerIntegration(t *testing.T)
func TestExistingP2PCompatibility(t *testing.T)
```

**테스트 시나리오:**
- ✅ 기존 `op-challenger` 시작 시 P2P 시스템 자동 활성화
- ✅ 기존 P2P 네트워크와의 호환성
- ✅ 설정 파일 통합 및 검증
- ✅ 로깅 시스템 통합

#### **5.2 생명주기 통합**
```go
func TestLifecycleIntegration(t *testing.T)
func TestGracefulShutdown(t *testing.T)
```

**테스트 시나리오:**
- ✅ `op-challenger` 시작 시 P2P 챌린저 시스템 시작
- ✅ `op-challenger` 종료 시 P2P 시스템 정상 종료
- ✅ 설정 변경 시 동적 재로드
- ✅ 에러 상황에서 안전한 종료

## 🧪 테스트 구현 계획

### **테스트 파일 구조**
```
op-challenger/p2p/network/tests/
├── node_test.go              # 노드 생성 및 설정 테스트
├── connection_test.go        # 연결 관리 테스트
├── discovery_test.go         # 발견 프로토콜 테스트
├── messages_test.go          # 메시지 시스템 테스트
├── stability_test.go         # 안정성 및 부하 테스트
├── integration_test.go       # 기존 시스템 통합 테스트
├── mock_transport.go         # 테스트용 모의 전송 계층
└── test_utils.go            # 테스트 유틸리티 함수
```

### **테스트 환경 설정**
```go
// 테스트 설정 구조체
type TestConfig struct {
    NodeCount        int           // 테스트할 노드 수 (기본: 5개)
    TestDuration     time.Duration // 테스트 지속 시간 (기본: 5분)
    MessageCount     int           // 전송할 메시지 수 (기본: 1000개)
    NetworkLatency   time.Duration // 네트워크 지연 시뮬레이션 (기본: 50ms)
    FailureRate      float64       // 장애 발생률 (기본: 0.1%)
    EnableLogging    bool          // 상세 로깅 활성화
}

// 빠른 테스트용 설정
func QuickTestConfig() *TestConfig {
    return &TestConfig{
        NodeCount:      3,
        TestDuration:   30 * time.Second,
        MessageCount:   100,
        NetworkLatency: 10 * time.Millisecond,
        FailureRate:    0.0,
        EnableLogging:  false,
    }
}

// 표준 테스트용 설정
func StandardTestConfig() *TestConfig {
    return &TestConfig{
        NodeCount:      5,
        TestDuration:   5 * time.Minute,
        MessageCount:   1000,
        NetworkLatency: 50 * time.Millisecond,
        FailureRate:    0.1,
        EnableLogging:  true,
    }
}
```

### **모의 객체 (Mock Objects)**
```go
// 테스트용 모의 전송 계층
type MockTransport struct {
    connections map[string]*MockConnection
    latency     time.Duration
    dropRate    float64
}

// 테스트용 모의 연결
type MockConnection struct {
    localAddr  string
    remoteAddr string
    messages   chan []byte
    closed     bool
}
```

## 📊 성능 벤치마크

### **성능 목표**
- **메모리 사용량**: < 50MB (100개 노드 기준)
- **CPU 사용률**: < 3% (정상 상태)
- **연결 설정 시간**: < 1초
- **메시지 전송 지연**: < 100ms
- **노드 발견 시간**: < 5초

### **벤치마크 테스트**
```go
// 벤치마크 테스트 파일: op-challenger/p2p/network/tests/benchmark_test.go
func BenchmarkNodeCreation(b *testing.B)
func BenchmarkMessageSending(b *testing.B)
func BenchmarkDiscoveryPerformance(b *testing.B)
func BenchmarkConnectionHandling(b *testing.B)
```

## 🔍 테스트 검증 기준

### **기능적 검증**
- ✅ 모든 챌린저 메시지 타입이 정상 전송/수신
- ✅ 노드 발견 및 연결이 예상 시간 내 완료
- ✅ 네트워크 장애 상황에서 자동 복구
- ✅ 메시지 손실률 < 0.1%

### **성능적 검증**
- ✅ 메모리 사용량이 목표치 이하 유지
- ✅ CPU 사용률이 목표치 이하 유지
- ✅ 네트워크 처리량이 요구사항 충족
- ✅ 응답 시간이 허용 범위 내

### **안정성 검증**
- ✅ 10분간 연속 실행 시 메모리 누수 없음
- ✅ 노드 재시작 후 정상 복구 (30초 내)
- ✅ 네트워크 분할 후 자동 재연결 (1분 내)
- ✅ 예외 상황에서 시스템 크래시 없음

## 🚀 구현 순서

### **1단계: 기본 테스트 구현** (1일)
1. 노드 생성 및 연결 테스트
2. 기본 메시지 전송/수신 테스트
3. 발견 프로토콜 기본 테스트

### **2단계: 고급 테스트 구현** (1일)
1. 안정성 및 부하 테스트
2. 기존 시스템 통합 테스트
3. 성능 벤치마크 테스트

### **3단계: 테스트 자동화** (0.5일)
1. CI/CD 파이프라인 통합
2. 자동화된 테스트 리포트
3. 성능 회귀 테스트

## 📝 결론

TODO 3.1에서는 구현된 P2P 챌린저 네트워크의 모든 기본 기능을 체계적으로 테스트하여 안정성과 성능을 검증한다. 특히 기존 `op-challenger` 시스템과의 통합을 중점적으로 테스트하여 실제 운영 환경에서의 호환성을 보장한다.

이 테스트를 통해 Phase 1의 P2P 네트워크 구축이 성공적으로 완료되었음을 검증하고, Phase 2의 어텐션 테스트 구현을 위한 안정적인 기반을 제공할 수 있다.
