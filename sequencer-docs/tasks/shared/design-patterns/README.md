# 공통 설계 패턴 가이드

이 폴더는 Optimism 시퀀서 시스템 개선 프로젝트에서 사용할 공통 설계 패턴들을 정의합니다.

## 🏗️ 아키텍처 패턴

### 1. 마이크로서비스 패턴
```go
// 서비스 인터페이스 정의
type Service interface {
    Start() error
    Stop() error
    Health() HealthStatus
}

// 서비스 구현
type BaseService struct {
    name    string
    config  *Config
    metrics *Metrics
}

func (bs *BaseService) Start() error {
    // 공통 시작 로직
    return nil
}

func (bs *BaseService) Stop() error {
    // 공통 정지 로직
    return nil
}
```

### 2. 이벤트 드리븐 패턴
```go
// 이벤트 정의
type Event struct {
    Type      EventType
    Data      interface{}
    Timestamp time.Time
    Source    string
}

// 이벤트 핸들러
type EventHandler interface {
    Handle(event *Event) error
}

// 이벤트 버스
type EventBus struct {
    handlers map[EventType][]EventHandler
    mu       sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event *Event) error {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    for _, handler := range eb.handlers[event.Type] {
        if err := handler.Handle(event); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. 팩토리 패턴
```go
// 팩토리 인터페이스
type Factory interface {
    Create(config *Config) (Service, error)
}

// 구체적인 팩토리 구현
type BackupPoolFactory struct{}

func (bpf *BackupPoolFactory) Create(config *Config) (Service, error) {
    return NewBackupPool(config)
}

type P2PChallengerFactory struct{}

func (pcf *P2PChallengerFactory) Create(config *Config) (Service, error) {
    return NewP2PChallenger(config)
}
```

## 🔄 동시성 패턴

### 1. 워커 풀 패턴
```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    wg         sync.WaitGroup
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()

    for job := range wp.jobQueue {
        result := wp.processJob(job)
        wp.resultChan <- result
    }
}

func (wp *WorkerPool) Submit(job Job) {
    wp.jobQueue <- job
}
```

### 2. 파이프라인 패턴
```go
type Pipeline struct {
    stages []Stage
}

type Stage interface {
    Process(input interface{}) (interface{}, error)
}

func (p *Pipeline) Execute(input interface{}) (interface{}, error) {
    result := input

    for _, stage := range p.stages {
        var err error
        result, err = stage.Process(result)
        if err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

### 3. 컨텍스트 패턴
```go
type Context struct {
    ctx    context.Context
    cancel context.CancelFunc
    data   map[string]interface{}
    mu     sync.RWMutex
}

func NewContext(parent context.Context) *Context {
    ctx, cancel := context.WithCancel(parent)
    return &Context{
        ctx:    ctx,
        cancel: cancel,
        data:   make(map[string]interface{}),
    }
}

func (c *Context) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    value, exists := c.data[key]
    return value, exists
}
```

## 🛡️ 보안 패턴

### 1. 인증 패턴
```go
type Authenticator interface {
    Authenticate(credentials *Credentials) (*Token, error)
    Validate(token *Token) (*Claims, error)
}

type JWTAuthenticator struct {
    secretKey []byte
    duration  time.Duration
}

func (ja *JWTAuthenticator) Authenticate(credentials *Credentials) (*Token, error) {
    // JWT 토큰 생성 로직
    return &Token{}, nil
}

func (ja *JWTAuthenticator) Validate(token *Token) (*Claims, error) {
    // JWT 토큰 검증 로직
    return &Claims{}, nil
}
```

### 2. 암호화 패턴
```go
type Encryptor interface {
    Encrypt(data []byte) ([]byte, error)
    Decrypt(data []byte) ([]byte, error)
}

type AESEncryptor struct {
    key []byte
}

func (ae *AESEncryptor) Encrypt(data []byte) ([]byte, error) {
    // AES 암호화 로직
    return data, nil
}

func (ae *AESEncryptor) Decrypt(data []byte) ([]byte, error) {
    // AES 복호화 로직
    return data, nil
}
```

## 📊 모니터링 패턴

### 1. 메트릭 패턴
```go
type Metrics interface {
    IncrementCounter(name string, labels map[string]string)
    SetGauge(name string, value float64, labels map[string]string)
    RecordHistogram(name string, value float64, labels map[string]string)
}

type PrometheusMetrics struct {
    registry *prometheus.Registry
}

func (pm *PrometheusMetrics) IncrementCounter(name string, labels map[string]string) {
    // Prometheus 카운터 증가
}

func (pm *PrometheusMetrics) SetGauge(name string, value float64, labels map[string]string) {
    // Prometheus 게이지 설정
}
```

### 2. 로깅 패턴
```go
type Logger interface {
    Debug(msg string, fields map[string]interface{})
    Info(msg string, fields map[string]interface{})
    Warn(msg string, fields map[string]interface{})
    Error(msg string, fields map[string]interface{})
}

type StructuredLogger struct {
    logger *zap.Logger
}

func (sl *StructuredLogger) Info(msg string, fields map[string]interface{}) {
    sl.logger.Info(msg, zap.Any("fields", fields))
}
```

## 🧪 테스트 패턴

### 1. 모킹 패턴
```go
type MockService struct {
    mock.Mock
}

func (ms *MockService) Start() error {
    args := ms.Called()
    return args.Error(0)
}

func (ms *MockService) Stop() error {
    args := ms.Called()
    return args.Error(0)
}

// 테스트에서 사용
func TestService(t *testing.T) {
    mockService := &MockService{}
    mockService.On("Start").Return(nil)
    mockService.On("Stop").Return(nil)

    // 테스트 로직
    mockService.AssertExpectations(t)
}
```

### 2. 테스트 헬퍼 패턴
```go
type TestHelper struct {
    t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
    return &TestHelper{t: t}
}

func (th *TestHelper) CreateTestConfig() *Config {
    return &Config{
        // 테스트용 설정
    }
}

func (th *TestHelper) AssertNoError(err error) {
    if err != nil {
        th.t.Fatalf("Expected no error, got: %v", err)
    }
}
```

## 🔧 설정 패턴

### 1. 설정 관리 패턴
```go
type Config struct {
    Environment string            `json:"environment"`
    Services    map[string]Config `json:"services"`
    Database    DatabaseConfig    `json:"database"`
    Network     NetworkConfig     `json:"network"`
}

type ConfigManager struct {
    config *Config
    mu     sync.RWMutex
}

func (cm *ConfigManager) Load(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return err
    }

    cm.mu.Lock()
    cm.config = &config
    cm.mu.Unlock()

    return nil
}

func (cm *ConfigManager) Get() *Config {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.config
}
```

### 2. 환경 변수 패턴
```go
type Environment struct {
    vars map[string]string
}

func LoadEnvironment() *Environment {
    env := &Environment{
        vars: make(map[string]string),
    }

    for _, e := range os.Environ() {
        pair := strings.SplitN(e, "=", 2)
        env.vars[pair[0]] = pair[1]
    }

    return env
}

func (e *Environment) Get(key, defaultValue string) string {
    if value, exists := e.vars[key]; exists {
        return value
    }
    return defaultValue
}
```

## 📝 사용 가이드

### 패턴 선택 기준
1. **마이크로서비스 패턴**: 독립적인 서비스 개발 시
2. **이벤트 드리븐 패턴**: 느슨한 결합이 필요한 경우
3. **워커 풀 패턴**: CPU 집약적 작업 처리 시
4. **파이프라인 패턴**: 순차적 데이터 처리 시
5. **컨텍스트 패턴**: 요청별 상태 관리 시

### 패턴 적용 순서
1. 요구사항 분석
2. 적절한 패턴 선택
3. 패턴 구현
4. 테스트 작성
5. 문서화

### 주의사항
- 패턴은 도구일 뿐, 과도한 사용은 복잡성 증가
- 프로젝트 규모에 맞는 패턴 선택
- 일관성 있는 패턴 사용
- 정기적인 패턴 리뷰 및 개선

---

**참고**: 이 패턴들은 Optimism 시퀀서 시스템 개선 프로젝트에서 실제 사용되는 패턴들입니다.
