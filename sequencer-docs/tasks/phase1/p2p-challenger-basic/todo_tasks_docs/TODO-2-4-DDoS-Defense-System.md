# TODO 2.4: DDoS 방어 시스템

## 📋 작업 개요
- **작업명**: DDoS 방어 시스템
- **상태**: 🔄 진행중
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
P2P 챌린저 네트워크를 DDoS 공격으로부터 보호하는 방어 시스템을 구현한다.

## 📐 구현 요구사항

### **🚀 Phase 1: 기본 DDoS 방어 (구현 대상)**
- ✅ 간단한 레이트 리미팅 (초당 메시지 수 제한)
- ✅ 기본 연결 수 제한
- ✅ IP 기반 차단
- ✅ 간단한 트래픽 모니터링

### **🔮 Phase 2-3: 고급 DDoS 방어 (향후 구현)**
- 🔄 ML 기반 공격 탐지
- 🔄 복잡한 트래픽 패턴 분석
- 🔄 적응형 방어 전략
- 🔄 공격 패턴 학습
- 🔄 분산 방어 네트워크
- 🔄 공격 정보 공유 시스템

## 🏗️ 시스템 설계

### 1. 통합 DDoS 방어 시스템

#### **1.1 DDoSDefenseSystem**
```go
// DDoSDefenseSystem provides comprehensive DDoS protection for P2P challenger network
type DDoSDefenseSystem struct {
    // 기본 정보
    nodeID          string
    challengerNode  *ChallengerP2PNode
    stateManager    *ChallengerStateManager

    // 핵심 방어 컴포넌트들
    trafficAnalyzer *TrafficAnalyzer       // 트래픽 분석기
    attackDetector  *AttackDetector        // 공격 탐지기
    defenseManager  *DefenseManager        // 방어 관리자
    blockingSystem  *BlockingSystem        // 차단 시스템

    // 지능형 방어 시스템
    patternLearner  *AttackPatternLearner  // 공격 패턴 학습기
    adaptiveDefense *AdaptiveDefense       // 적응형 방어
    reputationIntegration *ReputationIntegration // 평판 시스템 연동

    // 분산 협력 시스템
    defenseNetwork  *DefenseNetwork        // 방어 네트워크
    infoSharing     *AttackInfoSharing     // 공격 정보 공유

    // 모니터링 및 복구
    healthMonitor   *DefenseHealthMonitor  // 방어 시스템 건강 모니터
    recoveryManager *RecoveryManager       // 복구 관리자

    // 데이터 저장소
    attackHistory   *AttackHistory         // 공격 이력
    defenseMetrics  *DefenseMetrics        // 방어 메트릭스

    // 설정 및 제어
    config          *DDoSDefenseConfig     // 설정
    running         bool                   // 실행 상태

    // 동시성 제어
    mu              sync.RWMutex           // 뮤텍스
    ctx             context.Context        // 컨텍스트
    cancel          context.CancelFunc     // 취소 함수

    // 로깅 및 알림
    logger          log.Logger             // 로거
    alertManager    *AlertManager          // 알림 관리자
}

// DDoSDefenseConfig contains configuration for DDoS defense system
type DDoSDefenseConfig struct {
    // 기본 방어 설정
    DefenseEnabled          bool          `json:"defense_enabled"`           // 방어 활성화
    MonitoringInterval      time.Duration `json:"monitoring_interval"`       // 모니터링 간격
    AnalysisWindow          time.Duration `json:"analysis_window"`           // 분석 윈도우

    // 트래픽 임계값 설정
    MaxConnectionsPerSecond int           `json:"max_connections_per_second"` // 초당 최대 연결 수
    MaxMessagesPerSecond    int           `json:"max_messages_per_second"`    // 초당 최대 메시지 수
    MaxBytesPerSecond       uint64        `json:"max_bytes_per_second"`       // 초당 최대 바이트 수
    MaxConnectionsPerPeer   int           `json:"max_connections_per_peer"`   // 피어당 최대 연결 수

    // 공격 탐지 설정
    AttackDetectionEnabled  bool          `json:"attack_detection_enabled"`   // 공격 탐지 활성화
    SuspiciousThreshold     float64       `json:"suspicious_threshold"`       // 의심스러운 활동 임계값
    AttackThreshold         float64       `json:"attack_threshold"`           // 공격 임계값
    FalsePositiveRate       float64       `json:"false_positive_rate"`        // 허용 가능한 오탐률

    // 차단 설정
    AutoBlockingEnabled     bool          `json:"auto_blocking_enabled"`      // 자동 차단 활성화
    BlockDuration           time.Duration `json:"block_duration"`             // 차단 지속 시간
    MaxBlockDuration        time.Duration `json:"max_block_duration"`         // 최대 차단 시간
    BlockEscalationFactor   float64       `json:"block_escalation_factor"`    // 차단 에스컬레이션 팩터

    // 적응형 방어 설정
    AdaptiveDefenseEnabled  bool          `json:"adaptive_defense_enabled"`   // 적응형 방어 활성화
    LearningEnabled         bool          `json:"learning_enabled"`           // 학습 활성화
    PatternUpdateInterval   time.Duration `json:"pattern_update_interval"`    // 패턴 업데이트 간격

    // 분산 방어 설정
    DistributedDefenseEnabled bool        `json:"distributed_defense_enabled"` // 분산 방어 활성화
    InfoSharingEnabled      bool          `json:"info_sharing_enabled"`       // 정보 공유 활성화
    ReputationIntegrationEnabled bool     `json:"reputation_integration_enabled"` // 평판 연동 활성화

    // 복구 설정
    AutoRecoveryEnabled     bool          `json:"auto_recovery_enabled"`      // 자동 복구 활성화
    RecoveryCheckInterval   time.Duration `json:"recovery_check_interval"`    // 복구 체크 간격
    RecoveryThreshold       float64       `json:"recovery_threshold"`         // 복구 임계값
}

// AttackType represents different types of attacks
type AttackType int

const (
    AttackTypeUnknown AttackType = iota
    AttackTypeConnectionFlood                    // 연결 플러딩
    AttackTypeMessageFlood                       // 메시지 플러딩
    AttackTypeBandwidthExhaustion               // 대역폭 고갈
    AttackTypeSlowloris                         // 슬로우로리스
    AttackTypeAmplification                     // 증폭 공격
    AttackTypeDistributedFlood                  // 분산 플러딩
    AttackTypeProtocolExploit                   // 프로토콜 악용
    AttackTypeResourceExhaustion                // 리소스 고갈
)

// AttackSeverity represents attack severity levels
type AttackSeverity int

const (
    AttackSeverityLow AttackSeverity = iota
    AttackSeverityMedium
    AttackSeverityHigh
    AttackSeverityCritical
)

// DefenseAction represents defense actions that can be taken
type DefenseAction int

const (
    DefenseActionNone DefenseAction = iota
    DefenseActionMonitor                         // 모니터링만
    DefenseActionRateLimit                       // 속도 제한
    DefenseActionTemporaryBlock                  // 임시 차단
    DefenseActionPermanentBlock                  // 영구 차단
    DefenseActionQuarantine                      // 격리
    DefenseActionChallenge                       // 챌린지 요청
    DefenseActionReputationPenalty               // 평판 패널티
)

// AttackInfo represents information about a detected attack
type AttackInfo struct {
    // 기본 정보
    AttackID        string            `json:"attack_id"`
    AttackType      AttackType        `json:"attack_type"`
    Severity        AttackSeverity    `json:"severity"`

    // 공격원 정보
    SourceIPs       []string          `json:"source_ips"`
    SourcePeers     []string          `json:"source_peers"`
    AttackPattern   *AttackPattern    `json:"attack_pattern"`

    // 공격 특성
    StartTime       time.Time         `json:"start_time"`
    Duration        time.Duration     `json:"duration"`
    Volume          uint64            `json:"volume"`          // 공격 볼륨
    Frequency       float64           `json:"frequency"`       // 공격 빈도

    // 영향 정보
    ImpactLevel     ImpactLevel       `json:"impact_level"`
    AffectedServices []string         `json:"affected_services"`
    PerformanceImpact *PerformanceImpact `json:"performance_impact"`

    // 탐지 정보
    DetectionTime   time.Time         `json:"detection_time"`
    DetectionMethod string            `json:"detection_method"`
    Confidence      float64           `json:"confidence"`      // 탐지 신뢰도

    // 대응 정보
    DefenseActions  []DefenseAction   `json:"defense_actions"`
    ResponseTime    time.Duration     `json:"response_time"`
    Mitigated       bool              `json:"mitigated"`

    // 메타데이터
    ReportedBy      string            `json:"reported_by"`
    Verified        bool              `json:"verified"`
    SharedWithPeers bool              `json:"shared_with_peers"`
}

// AttackPattern represents a pattern of attack behavior
type AttackPattern struct {
    // 패턴 식별
    PatternID       string            `json:"pattern_id"`
    PatternType     AttackType        `json:"pattern_type"`
    Signature       string            `json:"signature"`       // 패턴 시그니처

    // 트래픽 특성
    TrafficProfile  *TrafficProfile   `json:"traffic_profile"`
    TimingPattern   *TimingPattern    `json:"timing_pattern"`
    VolumePattern   *VolumePattern    `json:"volume_pattern"`

    // 행동 특성
    BehaviorProfile *BehaviorProfile  `json:"behavior_profile"`

    // 통계 정보
    Frequency       int               `json:"frequency"`       // 발생 빈도
    LastSeen        time.Time         `json:"last_seen"`
    FirstSeen       time.Time         `json:"first_seen"`
    Effectiveness   float64           `json:"effectiveness"`   // 공격 효과성

    // 대응 정보
    CounterMeasures []DefenseAction   `json:"counter_measures"`
    SuccessRate     float64           `json:"success_rate"`    // 대응 성공률
}

// TrafficProfile represents traffic characteristics of an attack
type TrafficProfile struct {
    PacketSize      []int             `json:"packet_size"`     // 패킷 크기 분포
    PacketRate      float64           `json:"packet_rate"`     // 패킷 전송률
    Protocol        string            `json:"protocol"`        // 프로토콜
    Ports           []int             `json:"ports"`           // 사용 포트들
    Flags           []string          `json:"flags"`           // 프로토콜 플래그들
}

// TimingPattern represents timing characteristics of an attack
type TimingPattern struct {
    IntervalDistribution []time.Duration `json:"interval_distribution"` // 간격 분포
    BurstPattern        *BurstPattern   `json:"burst_pattern"`         // 버스트 패턴
    Periodicity         time.Duration   `json:"periodicity"`           // 주기성
    Randomness          float64         `json:"randomness"`            // 무작위성 정도
}

// BurstPattern represents burst attack patterns
type BurstPattern struct {
    BurstSize       int               `json:"burst_size"`       // 버스트 크기
    BurstDuration   time.Duration     `json:"burst_duration"`   // 버스트 지속시간
    BurstInterval   time.Duration     `json:"burst_interval"`   // 버스트 간격
    BurstIntensity  float64           `json:"burst_intensity"`  // 버스트 강도
}

// VolumePattern represents volume characteristics of an attack
type VolumePattern struct {
    TotalVolume     uint64            `json:"total_volume"`     // 총 볼륨
    PeakVolume      uint64            `json:"peak_volume"`      // 피크 볼륨
    AverageVolume   uint64            `json:"average_volume"`   // 평균 볼륨
    VolumeGrowthRate float64          `json:"volume_growth_rate"` // 볼륨 증가율
    VolumeDistribution []VolumePoint  `json:"volume_distribution"` // 볼륨 분포
}

// VolumePoint represents a point in volume distribution
type VolumePoint struct {
    Timestamp       time.Time         `json:"timestamp"`
    Volume          uint64            `json:"volume"`
    Rate            float64           `json:"rate"`
}

// BehaviorProfile represents behavioral characteristics of an attack
type BehaviorProfile struct {
    // 연결 행동
    ConnectionBehavior *ConnectionBehavior `json:"connection_behavior"`

    // 메시지 행동
    MessageBehavior   *MessageBehavior    `json:"message_behavior"`

    // 응답 행동
    ResponseBehavior  *ResponseBehavior   `json:"response_behavior"`

    // 지속성 행동
    PersistenceBehavior *PersistenceBehavior `json:"persistence_behavior"`
}

// ConnectionBehavior represents connection-related behavior
type ConnectionBehavior struct {
    ConnectionRate      float64           `json:"connection_rate"`       // 연결 생성률
    ConnectionDuration  time.Duration     `json:"connection_duration"`   // 연결 지속시간
    ReconnectionPattern *ReconnectionPattern `json:"reconnection_pattern"` // 재연결 패턴
    MultipleConnections bool              `json:"multiple_connections"`  // 다중 연결 시도
}

// ReconnectionPattern represents reconnection behavior patterns
type ReconnectionPattern struct {
    ReconnectionRate    float64           `json:"reconnection_rate"`     // 재연결률
    ReconnectionDelay   time.Duration     `json:"reconnection_delay"`    // 재연결 지연
    MaxReconnections    int               `json:"max_reconnections"`     // 최대 재연결 시도
    ReconnectionStrategy string           `json:"reconnection_strategy"` // 재연결 전략
}

// MessageBehavior represents message-related behavior
type MessageBehavior struct {
    MessageRate         float64           `json:"message_rate"`          // 메시지 전송률
    MessageSize         []int             `json:"message_size"`          // 메시지 크기 분포
    MessageTypes        []string          `json:"message_types"`         // 메시지 타입들
    PayloadPattern      string            `json:"payload_pattern"`       // 페이로드 패턴
    DuplicateMessages   bool              `json:"duplicate_messages"`    // 중복 메시지 여부
}

// ResponseBehavior represents response-related behavior
type ResponseBehavior struct {
    ResponseRate        float64           `json:"response_rate"`         // 응답률
    ResponseTime        time.Duration     `json:"response_time"`         // 응답 시간
    ResponsePattern     string            `json:"response_pattern"`      // 응답 패턴
    IgnoresRequests     bool              `json:"ignores_requests"`      // 요청 무시 여부
}

// PersistenceBehavior represents persistence-related behavior
type PersistenceBehavior struct {
    AttackDuration      time.Duration     `json:"attack_duration"`       // 공격 지속시간
    IntensityVariation  float64           `json:"intensity_variation"`   // 강도 변화
    AdaptationAbility   float64           `json:"adaptation_ability"`    // 적응 능력
    EvasionTechniques   []string          `json:"evasion_techniques"`    // 회피 기술들
}

// ImpactLevel represents the level of impact an attack has
type ImpactLevel int

const (
    ImpactLevelNone ImpactLevel = iota
    ImpactLevelMinor
    ImpactLevelModerate
    ImpactLevelMajor
    ImpactLevelSevere
    ImpactLevelCritical
)

// PerformanceImpact represents performance impact of an attack
type PerformanceImpact struct {
    CPUUsageIncrease    float64           `json:"cpu_usage_increase"`    // CPU 사용률 증가
    MemoryUsageIncrease float64           `json:"memory_usage_increase"` // 메모리 사용률 증가
    NetworkLatencyIncrease time.Duration  `json:"network_latency_increase"` // 네트워크 지연 증가
    ThroughputDecrease  float64           `json:"throughput_decrease"`   // 처리량 감소
    ConnectionDropRate  float64           `json:"connection_drop_rate"`  // 연결 끊김률
    ServiceAvailability float64           `json:"service_availability"`  // 서비스 가용성
}
```

### 2. 트래픽 분석 시스템

#### **2.1 TrafficAnalyzer**
```go
// TrafficAnalyzer analyzes network traffic patterns to detect anomalies
type TrafficAnalyzer struct {
    // 기본 정보
    nodeID          string
    defenseSystem   *DDoSDefenseSystem

    // 트래픽 수집기들
    connectionTracker *ConnectionTracker     // 연결 추적기
    messageTracker   *MessageTracker        // 메시지 추적기
    bandwidthTracker *BandwidthTracker      // 대역폭 추적기

    // 분석 엔진들
    statisticalAnalyzer *StatisticalAnalyzer // 통계 분석기
    patternAnalyzer    *PatternAnalyzer     // 패턴 분석기
    anomalyDetector    *AnomalyDetector     // 이상 탐지기

    // 데이터 저장소
    trafficData     *TrafficData           // 트래픽 데이터
    baseline        *TrafficBaseline       // 기준선 데이터
    analysisResults *AnalysisResults       // 분석 결과

    // 설정
    config          *TrafficAnalyzerConfig // 설정

    // 상태 관리
    running         bool                   // 실행 상태
    lastAnalysis    time.Time              // 마지막 분석 시간

    // 동시성 제어
    mu              sync.RWMutex           // 뮤텍스
    ctx             context.Context        // 컨텍스트
    cancel          context.CancelFunc     // 취소 함수

    logger          log.Logger             // 로거
}

// TrafficData represents comprehensive traffic data
type TrafficData struct {
    // 연결 데이터
    TotalConnections    int               `json:"total_connections"`
    ActiveConnections   int               `json:"active_connections"`
    ConnectionsPerSecond float64          `json:"connections_per_second"`
    ConnectionsByPeer   map[string]int    `json:"connections_by_peer"`

    // 메시지 데이터
    TotalMessages       uint64            `json:"total_messages"`
    MessagesPerSecond   float64           `json:"messages_per_second"`
    MessagesByType      map[string]uint64 `json:"messages_by_type"`
    MessagesByPeer      map[string]uint64 `json:"messages_by_peer"`

    // 대역폭 데이터
    TotalBytesReceived  uint64            `json:"total_bytes_received"`
    TotalBytesSent      uint64            `json:"total_bytes_sent"`
    BytesPerSecond      float64           `json:"bytes_per_second"`
    BandwidthByPeer     map[string]uint64 `json:"bandwidth_by_peer"`

    // 시간 정보
    CollectionPeriod    time.Duration     `json:"collection_period"`
    Timestamp           time.Time         `json:"timestamp"`
}

// TrafficBaseline represents normal traffic baseline
type TrafficBaseline struct {
    // 기준 연결 메트릭스
    NormalConnectionsPerSecond float64    `json:"normal_connections_per_second"`
    MaxConnectionsPerPeer     int         `json:"max_connections_per_peer"`
    TypicalConnectionDuration time.Duration `json:"typical_connection_duration"`

    // 기준 메시지 메트릭스
    NormalMessagesPerSecond   float64     `json:"normal_messages_per_second"`
    TypicalMessageSize        int         `json:"typical_message_size"`
    MessageTypeDistribution   map[string]float64 `json:"message_type_distribution"`

    // 기준 대역폭 메트릭스
    NormalBytesPerSecond      float64     `json:"normal_bytes_per_second"`
    TypicalBandwidthPerPeer   uint64      `json:"typical_bandwidth_per_peer"`

    // 통계 정보
    StandardDeviations        *StandardDeviations `json:"standard_deviations"`
    ConfidenceIntervals       *ConfidenceIntervals `json:"confidence_intervals"`

    // 메타데이터
    BaselineCreated           time.Time   `json:"baseline_created"`
    BaselineUpdated           time.Time   `json:"baseline_updated"`
    DataPoints                int         `json:"data_points"`
}

// StandardDeviations represents standard deviations for various metrics
type StandardDeviations struct {
    ConnectionsPerSecond      float64     `json:"connections_per_second"`
    MessagesPerSecond         float64     `json:"messages_per_second"`
    BytesPerSecond           float64     `json:"bytes_per_second"`
    MessageSize              float64     `json:"message_size"`
}

// ConfidenceIntervals represents confidence intervals for metrics
type ConfidenceIntervals struct {
    ConnectionsPerSecond      *ConfidenceInterval `json:"connections_per_second"`
    MessagesPerSecond         *ConfidenceInterval `json:"messages_per_second"`
    BytesPerSecond           *ConfidenceInterval `json:"bytes_per_second"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
    Lower                    float64     `json:"lower"`
    Upper                    float64     `json:"upper"`
    Confidence               float64     `json:"confidence"` // 신뢰도 (예: 0.95)
}

// AnalysisResults represents traffic analysis results
type AnalysisResults struct {
    // 이상 탐지 결과
    Anomalies               []TrafficAnomaly      `json:"anomalies"`
    AnomalyScore            float64               `json:"anomaly_score"`

    // 패턴 분석 결과
    DetectedPatterns        []TrafficPattern      `json:"detected_patterns"`
    PatternMatchScore       float64               `json:"pattern_match_score"`

    // 통계 분석 결과
    StatisticalDeviations   *StatisticalDeviations `json:"statistical_deviations"`

    // 위협 평가
    ThreatLevel             ThreatLevel           `json:"threat_level"`
    ThreatScore             float64               `json:"threat_score"`
    RecommendedActions      []DefenseAction       `json:"recommended_actions"`

    // 분석 메타데이터
    AnalysisTimestamp       time.Time             `json:"analysis_timestamp"`
    AnalysisDuration        time.Duration         `json:"analysis_duration"`
    DataQuality             float64               `json:"data_quality"`
    Confidence              float64               `json:"confidence"`
}

// TrafficAnomaly represents a detected traffic anomaly
type TrafficAnomaly struct {
    AnomalyID               string                `json:"anomaly_id"`
    AnomalyType             AnomalyType           `json:"anomaly_type"`
    Severity                AnomalySeverity       `json:"severity"`

    // 이상 데이터
    MetricName              string                `json:"metric_name"`
    ExpectedValue           float64               `json:"expected_value"`
    ActualValue             float64               `json:"actual_value"`
    Deviation               float64               `json:"deviation"`

    // 시간 정보
    StartTime               time.Time             `json:"start_time"`
    Duration                time.Duration         `json:"duration"`

    // 관련 정보
    AffectedPeers           []string              `json:"affected_peers"`
    RelatedAnomalies        []string              `json:"related_anomalies"`

    // 탐지 정보
    DetectionMethod         string                `json:"detection_method"`
    Confidence              float64               `json:"confidence"`
}

// AnomalyType represents types of traffic anomalies
type AnomalyType int

const (
    AnomalyTypeUnknown AnomalyType = iota
    AnomalyTypeConnectionSpike                   // 연결 급증
    AnomalyTypeMessageFlood                      // 메시지 플러딩
    AnomalyTypeBandwidthSpike                    // 대역폭 급증
    AnomalyTypeUnusualPattern                    // 비정상 패턴
    AnomalyTypeFrequencyAnomaly                  // 빈도 이상
    AnomalyTypeSizeAnomaly                       // 크기 이상
    AnomalyTypeTimingAnomaly                     // 타이밍 이상
)

// AnomalySeverity represents anomaly severity levels
type AnomalySeverity int

const (
    AnomalySeverityLow AnomalySeverity = iota
    AnomalySeverityMedium
    AnomalySeverityHigh
    AnomalySeverityCritical
)

// ThreatLevel represents overall threat level
type ThreatLevel int

const (
    ThreatLevelNone ThreatLevel = iota
    ThreatLevelLow
    ThreatLevelModerate
    ThreatLevelHigh
    ThreatLevelCritical
    ThreatLevelEmergency
)

// TrafficPattern represents a detected traffic pattern
type TrafficPattern struct {
    PatternID               string                `json:"pattern_id"`
    PatternType             PatternType           `json:"pattern_type"`
    PatternName             string                `json:"pattern_name"`

    // 패턴 특성
    Characteristics         *PatternCharacteristics `json:"characteristics"`

    // 매칭 정보
    MatchStrength           float64               `json:"match_strength"`
    MatchedDataPoints       int                   `json:"matched_data_points"`

    // 시간 정보
    FirstDetected           time.Time             `json:"first_detected"`
    LastSeen                time.Time             `json:"last_seen"`
    Frequency               int                   `json:"frequency"`

    // 위험도 평가
    RiskLevel               RiskLevel             `json:"risk_level"`
    ThreatIndicator         bool                  `json:"threat_indicator"`
}

// PatternType represents types of traffic patterns
type PatternType int

const (
    PatternTypeNormal PatternType = iota
    PatternTypeSuspicious
    PatternTypeMalicious
    PatternTypeAttack
    PatternTypeReconnaissance
    PatternTypeEvasion
)

// PatternCharacteristics represents characteristics of a traffic pattern
type PatternCharacteristics struct {
    // 시간적 특성
    TemporalFeatures        *TemporalFeatures     `json:"temporal_features"`

    // 볼륨 특성
    VolumeFeatures          *VolumeFeatures       `json:"volume_features"`

    // 행동 특성
    BehavioralFeatures      *BehavioralFeatures   `json:"behavioral_features"`

    // 통계적 특성
    StatisticalFeatures     *StatisticalFeatures  `json:"statistical_features"`
}

// TemporalFeatures represents temporal characteristics
type TemporalFeatures struct {
    Periodicity             time.Duration         `json:"periodicity"`
    Regularity              float64               `json:"regularity"`
    Burstiness              float64               `json:"burstiness"`
    TimeDistribution        []TimeDistributionPoint `json:"time_distribution"`
}

// TimeDistributionPoint represents a point in time distribution
type TimeDistributionPoint struct {
    Hour                    int                   `json:"hour"`
    Frequency               float64               `json:"frequency"`
}

// VolumeFeatures represents volume characteristics
type VolumeFeatures struct {
    AverageVolume           float64               `json:"average_volume"`
    PeakVolume              float64               `json:"peak_volume"`
    VolumeVariability       float64               `json:"volume_variability"`
    GrowthRate              float64               `json:"growth_rate"`
}

// BehavioralFeatures represents behavioral characteristics
type BehavioralFeatures struct {
    Aggressiveness          float64               `json:"aggressiveness"`
    Persistence             float64               `json:"persistence"`
    Adaptability            float64               `json:"adaptability"`
    Coordination            float64               `json:"coordination"`
}

// StatisticalFeatures represents statistical characteristics
type StatisticalFeatures struct {
    Mean                    float64               `json:"mean"`
    StandardDeviation       float64               `json:"standard_deviation"`
    Skewness                float64               `json:"skewness"`
    Kurtosis                float64               `json:"kurtosis"`
    Entropy                 float64               `json:"entropy"`
}

// RiskLevel represents risk level of a pattern
type RiskLevel int

const (
    RiskLevelVeryLow RiskLevel = iota
    RiskLevelLow
    RiskLevelModerate
    RiskLevelHigh
    RiskLevelVeryHigh
    RiskLevelExtreme
)

// StatisticalDeviations represents statistical deviations from baseline
type StatisticalDeviations struct {
    ConnectionDeviations    *MetricDeviation      `json:"connection_deviations"`
    MessageDeviations       *MetricDeviation      `json:"message_deviations"`
    BandwidthDeviations     *MetricDeviation      `json:"bandwidth_deviations"`
}

// MetricDeviation represents deviation of a specific metric
type MetricDeviation struct {
    MetricName              string                `json:"metric_name"`
    BaselineValue           float64               `json:"baseline_value"`
    CurrentValue            float64               `json:"current_value"`
    AbsoluteDeviation       float64               `json:"absolute_deviation"`
    RelativeDeviation       float64               `json:"relative_deviation"`
    StandardDeviations      float64               `json:"standard_deviations"`
    Significance            SignificanceLevel     `json:"significance"`
}

// SignificanceLevel represents statistical significance level
type SignificanceLevel int

const (
    SignificanceLevelNone SignificanceLevel = iota
    SignificanceLevelLow
    SignificanceLevelModerate
    SignificanceLevelHigh
    SignificanceLevelVeryHigh
)

// TrafficAnalyzerConfig contains configuration for traffic analyzer
type TrafficAnalyzerConfig struct {
    // 분석 설정
    AnalysisEnabled         bool          `json:"analysis_enabled"`         // 분석 활성화
    AnalysisInterval        time.Duration `json:"analysis_interval"`        // 분석 간격
    AnalysisWindow          time.Duration `json:"analysis_window"`          // 분석 윈도우

    // 기준선 설정
    BaselineEnabled         bool          `json:"baseline_enabled"`         // 기준선 활성화
    BaselineUpdateInterval  time.Duration `json:"baseline_update_interval"` // 기준선 업데이트 간격
    BaselineLearningPeriod  time.Duration `json:"baseline_learning_period"` // 기준선 학습 기간

    // 이상 탐지 설정
    AnomalyDetectionEnabled bool          `json:"anomaly_detection_enabled"` // 이상 탐지 활성화
    AnomalyThreshold        float64       `json:"anomaly_threshold"`        // 이상 임계값
    SensitivityLevel        float64       `json:"sensitivity_level"`        // 민감도 수준

    // 패턴 분석 설정
    PatternAnalysisEnabled  bool          `json:"pattern_analysis_enabled"` // 패턴 분석 활성화
    PatternMatchThreshold   float64       `json:"pattern_match_threshold"`  // 패턴 매칭 임계값
    PatternLearningEnabled  bool          `json:"pattern_learning_enabled"` // 패턴 학습 활성화

    // 성능 설정
    MaxDataPoints           int           `json:"max_data_points"`          // 최대 데이터 포인트
    DataRetentionPeriod     time.Duration `json:"data_retention_period"`    // 데이터 보관 기간
    AnalysisTimeout         time.Duration `json:"analysis_timeout"`         // 분석 타임아웃
}

// AnalyzeTraffic performs comprehensive traffic analysis
func (ta *TrafficAnalyzer) AnalyzeTraffic() *AnalysisResults {
    ta.mu.Lock()
    defer ta.mu.Unlock()

    startTime := time.Now()

    // 현재 트래픽 데이터 수집
    currentData := ta.collectCurrentTrafficData()

    // 통계 분석 수행
    statisticalResults := ta.statisticalAnalyzer.Analyze(currentData, ta.baseline)

    // 패턴 분석 수행
    patternResults := ta.patternAnalyzer.AnalyzePatterns(currentData)

    // 이상 탐지 수행
    anomalies := ta.anomalyDetector.DetectAnomalies(currentData, ta.baseline)

    // 위협 평가 수행
    threatLevel, threatScore := ta.evaluateThreat(statisticalResults, patternResults, anomalies)

    // 권장 조치 결정
    recommendedActions := ta.determineRecommendedActions(threatLevel, anomalies)

    // 분석 결과 구성
    results := &AnalysisResults{
        Anomalies:             anomalies,
        AnomalyScore:          ta.calculateAnomalyScore(anomalies),
        DetectedPatterns:      patternResults.Patterns,
        PatternMatchScore:     patternResults.OverallMatchScore,
        StatisticalDeviations: statisticalResults,
        ThreatLevel:          threatLevel,
        ThreatScore:          threatScore,
        RecommendedActions:   recommendedActions,
        AnalysisTimestamp:    startTime,
        AnalysisDuration:     time.Since(startTime),
        DataQuality:          ta.assessDataQuality(currentData),
        Confidence:           ta.calculateAnalysisConfidence(statisticalResults, patternResults, anomalies),
    }

    // 결과 저장
    ta.analysisResults = results
    ta.lastAnalysis = startTime

    ta.logger.Info("Traffic analysis completed",
        "threat_level", threatLevel, "anomalies", len(anomalies),
        "patterns", len(patternResults.Patterns), "duration", results.AnalysisDuration)

    return results
}

// collectCurrentTrafficData collects current traffic data from all trackers
func (ta *TrafficAnalyzer) collectCurrentTrafficData() *TrafficData {
    // 각 추적기에서 데이터 수집
    connectionData := ta.connectionTracker.GetCurrentData()
    messageData := ta.messageTracker.GetCurrentData()
    bandwidthData := ta.bandwidthTracker.GetCurrentData()

    return &TrafficData{
        TotalConnections:    connectionData.TotalConnections,
        ActiveConnections:   connectionData.ActiveConnections,
        ConnectionsPerSecond: connectionData.ConnectionsPerSecond,
        ConnectionsByPeer:   connectionData.ConnectionsByPeer,
        TotalMessages:       messageData.TotalMessages,
        MessagesPerSecond:   messageData.MessagesPerSecond,
        MessagesByType:      messageData.MessagesByType,
        MessagesByPeer:      messageData.MessagesByPeer,
        TotalBytesReceived:  bandwidthData.TotalBytesReceived,
        TotalBytesSent:      bandwidthData.TotalBytesSent,
        BytesPerSecond:      bandwidthData.BytesPerSecond,
        BandwidthByPeer:     bandwidthData.BandwidthByPeer,
        CollectionPeriod:    ta.config.AnalysisWindow,
        Timestamp:           time.Now(),
    }
}

// evaluateThreat evaluates overall threat level based on analysis results
func (ta *TrafficAnalyzer) evaluateThreat(statistical *StatisticalDeviations, patterns *PatternAnalysisResults, anomalies []TrafficAnomaly) (ThreatLevel, float64) {
    threatScore := 0.0

    // 이상 기반 위협 점수
    for _, anomaly := range anomalies {
        switch anomaly.Severity {
        case AnomalySeverityLow:
            threatScore += 10.0
        case AnomalySeverityMedium:
            threatScore += 25.0
        case AnomalySeverityHigh:
            threatScore += 50.0
        case AnomalySeverityCritical:
            threatScore += 100.0
        }
    }

    // 패턴 기반 위협 점수
    for _, pattern := range patterns.Patterns {
        switch pattern.RiskLevel {
        case RiskLevelLow:
            threatScore += 5.0
        case RiskLevelModerate:
            threatScore += 15.0
        case RiskLevelHigh:
            threatScore += 30.0
        case RiskLevelVeryHigh:
            threatScore += 60.0
        case RiskLevelExtreme:
            threatScore += 120.0
        }
    }

    // 통계적 편차 기반 위협 점수
    if statistical != nil {
        if statistical.ConnectionDeviations != nil && statistical.ConnectionDeviations.StandardDeviations > 3.0 {
            threatScore += 20.0
        }
        if statistical.MessageDeviations != nil && statistical.MessageDeviations.StandardDeviations > 3.0 {
            threatScore += 20.0
        }
        if statistical.BandwidthDeviations != nil && statistical.BandwidthDeviations.StandardDeviations > 3.0 {
            threatScore += 20.0
        }
    }

    // 위협 레벨 결정
    var threatLevel ThreatLevel
    switch {
    case threatScore >= 200.0:
        threatLevel = ThreatLevelEmergency
    case threatScore >= 100.0:
        threatLevel = ThreatLevelCritical
    case threatScore >= 50.0:
        threatLevel = ThreatLevelHigh
    case threatScore >= 20.0:
        threatLevel = ThreatLevelModerate
    case threatScore >= 5.0:
        threatLevel = ThreatLevelLow
    default:
        threatLevel = ThreatLevelNone
    }

    return threatLevel, threatScore
}

// determineRecommendedActions determines recommended defense actions based on threat assessment
func (ta *TrafficAnalyzer) determineRecommendedActions(threatLevel ThreatLevel, anomalies []TrafficAnomaly) []DefenseAction {
    var actions []DefenseAction

    switch threatLevel {
    case ThreatLevelEmergency:
        actions = append(actions, DefenseActionPermanentBlock, DefenseActionQuarantine)
    case ThreatLevelCritical:
        actions = append(actions, DefenseActionTemporaryBlock, DefenseActionReputationPenalty)
    case ThreatLevelHigh:
        actions = append(actions, DefenseActionRateLimit, DefenseActionChallenge)
    case ThreatLevelModerate:
        actions = append(actions, DefenseActionMonitor, DefenseActionRateLimit)
    case ThreatLevelLow:
        actions = append(actions, DefenseActionMonitor)
    default:
        actions = append(actions, DefenseActionNone)
    }

    // 특정 이상에 대한 추가 조치
    for _, anomaly := range anomalies {
        switch anomaly.AnomalyType {
        case AnomalyTypeConnectionSpike:
            if !contains(actions, DefenseActionRateLimit) {
                actions = append(actions, DefenseActionRateLimit)
            }
        case AnomalyTypeMessageFlood:
            if !contains(actions, DefenseActionTemporaryBlock) {
                actions = append(actions, DefenseActionTemporaryBlock)
            }
        case AnomalyTypeBandwidthSpike:
            if !contains(actions, DefenseActionQuarantine) {
                actions = append(actions, DefenseActionQuarantine)
            }
        }
    }

    return actions
}

// calculateAnomalyScore calculates overall anomaly score
func (ta *TrafficAnalyzer) calculateAnomalyScore(anomalies []TrafficAnomaly) float64 {
    if len(anomalies) == 0 {
        return 0.0
    }

    totalScore := 0.0
    for _, anomaly := range anomalies {
        score := anomaly.Deviation * anomaly.Confidence
        switch anomaly.Severity {
        case AnomalySeverityCritical:
            score *= 4.0
        case AnomalySeverityHigh:
            score *= 2.0
        case AnomalySeverityMedium:
            score *= 1.5
        }
        totalScore += score
    }

    return totalScore / float64(len(anomalies))
}

// assessDataQuality assesses the quality of collected traffic data
func (ta *TrafficAnalyzer) assessDataQuality(data *TrafficData) float64 {
    quality := 1.0

    // 데이터 완성도 체크
    if data.TotalConnections < 0 || data.TotalMessages < 0 || data.TotalBytesReceived < 0 {
        quality -= 0.3
    }

    // 데이터 일관성 체크
    if data.ActiveConnections > data.TotalConnections {
        quality -= 0.2
    }

    // 시간 일관성 체크
    if time.Since(data.Timestamp) > ta.config.AnalysisWindow*2 {
        quality -= 0.2
    }

    // 데이터 범위 체크
    if data.ConnectionsPerSecond < 0 || data.MessagesPerSecond < 0 || data.BytesPerSecond < 0 {
        quality -= 0.3
    }

    if quality < 0 {
        quality = 0
    }

    return quality
}

// calculateAnalysisConfidence calculates confidence in analysis results
func (ta *TrafficAnalyzer) calculateAnalysisConfidence(statistical *StatisticalDeviations, patterns *PatternAnalysisResults, anomalies []TrafficAnomaly) float64 {
    confidence := 0.0
    factors := 0

    // 통계 분석 신뢰도
    if statistical != nil {
        confidence += 0.8 // 통계 분석은 일반적으로 신뢰도가 높음
        factors++
    }

    // 패턴 분석 신뢰도
    if patterns != nil && len(patterns.Patterns) > 0 {
        confidence += patterns.OverallMatchScore
        factors++
    }

    // 이상 탐지 신뢰도
    if len(anomalies) > 0 {
        anomalyConfidence := 0.0
        for _, anomaly := range anomalies {
            anomalyConfidence += anomaly.Confidence
        }
        confidence += anomalyConfidence / float64(len(anomalies))
        factors++
    }

    if factors == 0 {
        return 0.0
    }

    return confidence / float64(factors)
}

// contains checks if a slice contains a specific defense action
func contains(actions []DefenseAction, action DefenseAction) bool {
    for _, a := range actions {
        if a == action {
            return true
        }
    }
    return false
}

// PatternAnalysisResults represents results of pattern analysis
type PatternAnalysisResults struct {
    Patterns           []TrafficPattern `json:"patterns"`
    OverallMatchScore  float64          `json:"overall_match_score"`
    NewPatternsDetected int             `json:"new_patterns_detected"`
    KnownPatternsMatched int            `json:"known_patterns_matched"`
}
```

### 3. 공격 탐지 시스템

#### **3.1 AttackDetector**
```go
// AttackDetector detects various types of DDoS attacks
type AttackDetector struct {
    // 기본 정보
    nodeID          string
    defenseSystem   *DDoSDefenseSystem
    trafficAnalyzer *TrafficAnalyzer

    // 탐지 엔진들
    signatureDetector *SignatureDetector     // 시그니처 기반 탐지
    behaviorDetector  *BehaviorDetector      // 행동 기반 탐지
    statisticalDetector *StatisticalDetector // 통계 기반 탐지
    mlDetector        *MLDetector            // 머신러닝 기반 탐지

    // 탐지 규칙 및 패턴
    attackSignatures  *AttackSignatures      // 공격 시그니처들
    detectionRules    *DetectionRules        // 탐지 규칙들

    // 탐지 결과
    detectedAttacks   []AttackInfo           // 탐지된 공격들
    activeAttacks     map[string]*AttackInfo // 진행 중인 공격들

    // 설정
    config            *AttackDetectorConfig  // 설정

    // 상태 관리
    running           bool                   // 실행 상태
    lastDetection     time.Time              // 마지막 탐지 시간

    // 동시성 제어
    mu                sync.RWMutex           // 뮤텍스
    ctx               context.Context        // 컨텍스트
    cancel            context.CancelFunc     // 취소 함수

    logger            log.Logger             // 로거
}

// AttackSignatures contains known attack signatures
type AttackSignatures struct {
    Signatures        map[string]*AttackSignature `json:"signatures"`
    LastUpdated       time.Time                   `json:"last_updated"`
    Version           string                      `json:"version"`
}

// AttackSignature represents a specific attack signature
type AttackSignature struct {
    SignatureID       string                      `json:"signature_id"`
    SignatureName     string                      `json:"signature_name"`
    AttackType        AttackType                  `json:"attack_type"`
    Severity          AttackSeverity              `json:"severity"`

    // 시그니처 패턴
    TrafficPattern    *TrafficSignaturePattern    `json:"traffic_pattern"`
    BehaviorPattern   *BehaviorSignaturePattern   `json:"behavior_pattern"`
    TimingPattern     *TimingSignaturePattern     `json:"timing_pattern"`

    // 매칭 조건
    MatchingCriteria  *MatchingCriteria           `json:"matching_criteria"`

    // 메타데이터
    Created           time.Time                   `json:"created"`
    LastSeen          time.Time                   `json:"last_seen"`
    Frequency         int                         `json:"frequency"`
    Effectiveness     float64                     `json:"effectiveness"`
}

// TrafficSignaturePattern represents traffic-based signature pattern
type TrafficSignaturePattern struct {
    ConnectionRate    *RangePattern               `json:"connection_rate"`
    MessageRate       *RangePattern               `json:"message_rate"`
    BandwidthUsage    *RangePattern               `json:"bandwidth_usage"`
    PacketSize        *RangePattern               `json:"packet_size"`
    ProtocolFlags     []string                    `json:"protocol_flags"`
}

// BehaviorSignaturePattern represents behavior-based signature pattern
type BehaviorSignaturePattern struct {
    ConnectionBehavior *ConnectionBehaviorPattern `json:"connection_behavior"`
    MessageBehavior   *MessageBehaviorPattern     `json:"message_behavior"`
    ResponseBehavior  *ResponseBehaviorPattern    `json:"response_behavior"`
    PersistenceBehavior *PersistenceBehaviorPattern `json:"persistence_behavior"`
}

// TimingSignaturePattern represents timing-based signature pattern
type TimingSignaturePattern struct {
    IntervalPattern   *IntervalPattern            `json:"interval_pattern"`
    BurstPattern      *BurstSignaturePattern      `json:"burst_pattern"`
    PeriodicityPattern *PeriodicityPattern        `json:"periodicity_pattern"`
}

// RangePattern represents a range-based pattern
type RangePattern struct {
    MinValue          float64                     `json:"min_value"`
    MaxValue          float64                     `json:"max_value"`
    TypicalValue      float64                     `json:"typical_value"`
    Tolerance         float64                     `json:"tolerance"`
}

// ConnectionBehaviorPattern represents connection behavior pattern
type ConnectionBehaviorPattern struct {
    MultipleConnections bool                      `json:"multiple_connections"`
    ShortLivedConnections bool                    `json:"short_lived_connections"`
    ReconnectionAttempts *RangePattern            `json:"reconnection_attempts"`
    ConnectionSpread    bool                      `json:"connection_spread"`
}

// MessageBehaviorPattern represents message behavior pattern
type MessageBehaviorPattern struct {
    HighFrequency     bool                        `json:"high_frequency"`
    LargeMessages     bool                        `json:"large_messages"`
    DuplicateMessages bool                        `json:"duplicate_messages"`
    InvalidMessages   bool                        `json:"invalid_messages"`
    MessageTypeSpread bool                        `json:"message_type_spread"`
}

// ResponseBehaviorPattern represents response behavior pattern
type ResponseBehaviorPattern struct {
    NoResponses       bool                        `json:"no_responses"`
    DelayedResponses  bool                        `json:"delayed_responses"`
    InvalidResponses  bool                        `json:"invalid_responses"`
    IgnoresChallenges bool                        `json:"ignores_challenges"`
}

// PersistenceBehaviorPattern represents persistence behavior pattern
type PersistenceBehaviorPattern struct {
    LongDuration      bool                        `json:"long_duration"`
    IntensityVariation bool                       `json:"intensity_variation"`
    AdaptesToDefense  bool                        `json:"adapts_to_defense"`
    UseEvasion        bool                        `json:"use_evasion"`
}

// IntervalPattern represents interval-based pattern
type IntervalPattern struct {
    RegularIntervals  bool                        `json:"regular_intervals"`
    ShortIntervals    bool                        `json:"short_intervals"`
    RandomIntervals   bool                        `json:"random_intervals"`
    IntervalRange     *RangePattern               `json:"interval_range"`
}

// BurstSignaturePattern represents burst signature pattern
type BurstSignaturePattern struct {
    HasBursts         bool                        `json:"has_bursts"`
    BurstSize         *RangePattern               `json:"burst_size"`
    BurstDuration     *RangePattern               `json:"burst_duration"`
    BurstInterval     *RangePattern               `json:"burst_interval"`
}

// PeriodicityPattern represents periodicity pattern
type PeriodicityPattern struct {
    IsPeriodic        bool                        `json:"is_periodic"`
    Period            *RangePattern               `json:"period"`
    Regularity        float64                     `json:"regularity"`
}

// MatchingCriteria represents criteria for signature matching
type MatchingCriteria struct {
    MinMatchScore     float64                     `json:"min_match_score"`
    RequiredPatterns  []string                    `json:"required_patterns"`
    OptionalPatterns  []string                    `json:"optional_patterns"`
    ExclusionPatterns []string                    `json:"exclusion_patterns"`
    TimeWindow        time.Duration               `json:"time_window"`
}

// DetectionRules contains detection rules for various attack types
type DetectionRules struct {
    Rules             map[string]*DetectionRule   `json:"rules"`
    LastUpdated       time.Time                   `json:"last_updated"`
    Version           string                      `json:"version"`
}

// DetectionRule represents a detection rule
type DetectionRule struct {
    RuleID            string                      `json:"rule_id"`
    RuleName          string                      `json:"rule_name"`
    AttackType        AttackType                  `json:"attack_type"`
    Severity          AttackSeverity              `json:"severity"`

    // 규칙 조건들
    Conditions        []RuleCondition             `json:"conditions"`
    LogicalOperator   LogicalOperator             `json:"logical_operator"`

    // 임계값들
    Thresholds        *RuleThresholds             `json:"thresholds"`

    // 시간 조건
    TimeConstraints   *TimeConstraints            `json:"time_constraints"`

    // 액션
    Actions           []DefenseAction             `json:"actions"`

    // 메타데이터
    Enabled           bool                        `json:"enabled"`
    Created           time.Time                   `json:"created"`
    LastTriggered     time.Time                   `json:"last_triggered"`
    TriggerCount      int                         `json:"trigger_count"`
}

// RuleCondition represents a condition in a detection rule
type RuleCondition struct {
    ConditionID       string                      `json:"condition_id"`
    MetricName        string                      `json:"metric_name"`
    Operator          ComparisonOperator          `json:"operator"`
    Value             float64                     `json:"value"`
    TimeWindow        time.Duration               `json:"time_window"`
    Weight            float64                     `json:"weight"`
}

// LogicalOperator represents logical operators for combining conditions
type LogicalOperator int

const (
    LogicalOperatorAND LogicalOperator = iota
    LogicalOperatorOR
    LogicalOperatorNOT
    LogicalOperatorXOR
)

// ComparisonOperator represents comparison operators
type ComparisonOperator int

const (
    ComparisonOperatorEqual ComparisonOperator = iota
    ComparisonOperatorNotEqual
    ComparisonOperatorGreater
    ComparisonOperatorGreaterEqual
    ComparisonOperatorLess
    ComparisonOperatorLessEqual
    ComparisonOperatorBetween
    ComparisonOperatorNotBetween
)

// RuleThresholds represents thresholds for a detection rule
type RuleThresholds struct {
    TriggerThreshold  float64                     `json:"trigger_threshold"`
    ClearThreshold    float64                     `json:"clear_threshold"`
    EscalationThreshold float64                   `json:"escalation_threshold"`
    MaxThreshold      float64                     `json:"max_threshold"`
}

// TimeConstraints represents time-based constraints
type TimeConstraints struct {
    MinDuration       time.Duration               `json:"min_duration"`
    MaxDuration       time.Duration               `json:"max_duration"`
    EvaluationWindow  time.Duration               `json:"evaluation_window"`
    CooldownPeriod    time.Duration               `json:"cooldown_period"`
}

// AttackDetectorConfig contains configuration for attack detector
type AttackDetectorConfig struct {
    // 탐지 설정
    DetectionEnabled        bool          `json:"detection_enabled"`         // 탐지 활성화
    DetectionInterval       time.Duration `json:"detection_interval"`        // 탐지 간격
    DetectionWindow         time.Duration `json:"detection_window"`          // 탐지 윈도우

    // 시그니처 탐지 설정
    SignatureDetectionEnabled bool        `json:"signature_detection_enabled"` // 시그니처 탐지 활성화
    SignatureMatchThreshold float64       `json:"signature_match_threshold"`  // 시그니처 매칭 임계값
    SignatureUpdateInterval time.Duration `json:"signature_update_interval"`  // 시그니처 업데이트 간격

    // 행동 탐지 설정
    BehaviorDetectionEnabled bool         `json:"behavior_detection_enabled"` // 행동 탐지 활성화
    BehaviorAnalysisWindow  time.Duration `json:"behavior_analysis_window"`   // 행동 분석 윈도우
    BehaviorAnomalyThreshold float64      `json:"behavior_anomaly_threshold"` // 행동 이상 임계값

    // 통계 탐지 설정
    StatisticalDetectionEnabled bool      `json:"statistical_detection_enabled"` // 통계 탐지 활성화
    StatisticalSignificanceLevel float64  `json:"statistical_significance_level"` // 통계적 유의성 수준
    StatisticalConfidenceLevel float64    `json:"statistical_confidence_level"`   // 통계적 신뢰 수준

    // ML 탐지 설정
    MLDetectionEnabled      bool          `json:"ml_detection_enabled"`       // ML 탐지 활성화
    MLModelUpdateInterval   time.Duration `json:"ml_model_update_interval"`   // ML 모델 업데이트 간격
    MLPredictionThreshold   float64       `json:"ml_prediction_threshold"`    // ML 예측 임계값

    // 성능 설정
    MaxConcurrentDetections int           `json:"max_concurrent_detections"`  // 최대 동시 탐지 수
    DetectionTimeout        time.Duration `json:"detection_timeout"`          // 탐지 타임아웃
    MaxAttackHistory        int           `json:"max_attack_history"`         // 최대 공격 이력 수
}

// DetectAttacks performs comprehensive attack detection
func (ad *AttackDetector) DetectAttacks() []AttackInfo {
    ad.mu.Lock()
    defer ad.mu.Unlock()

    var detectedAttacks []AttackInfo

    // 현재 트래픽 분석 결과 가져오기
    analysisResults := ad.trafficAnalyzer.GetLatestAnalysisResults()
    if analysisResults == nil {
        return detectedAttacks
    }

    // 시그니처 기반 탐지
    if ad.config.SignatureDetectionEnabled {
        signatureAttacks := ad.signatureDetector.DetectBySignature(analysisResults)
        detectedAttacks = append(detectedAttacks, signatureAttacks...)
    }

    // 행동 기반 탐지
    if ad.config.BehaviorDetectionEnabled {
        behaviorAttacks := ad.behaviorDetector.DetectByBehavior(analysisResults)
        detectedAttacks = append(detectedAttacks, behaviorAttacks...)
    }

    // 통계 기반 탐지
    if ad.config.StatisticalDetectionEnabled {
        statisticalAttacks := ad.statisticalDetector.DetectByStatistics(analysisResults)
        detectedAttacks = append(detectedAttacks, statisticalAttacks...)
    }

    // ML 기반 탐지
    if ad.config.MLDetectionEnabled {
        mlAttacks := ad.mlDetector.DetectByML(analysisResults)
        detectedAttacks = append(detectedAttacks, mlAttacks...)
    }

    // 탐지 결과 후처리
    processedAttacks := ad.postProcessDetections(detectedAttacks)

    // 활성 공격 업데이트
    ad.updateActiveAttacks(processedAttacks)

    // 탐지 이력 업데이트
    ad.detectedAttacks = append(ad.detectedAttacks, processedAttacks...)

    // 이력 크기 제한
    if len(ad.detectedAttacks) > ad.config.MaxAttackHistory {
        ad.detectedAttacks = ad.detectedAttacks[len(ad.detectedAttacks)-ad.config.MaxAttackHistory:]
    }

    ad.lastDetection = time.Now()

    ad.logger.Info("Attack detection completed",
        "detected_attacks", len(processedAttacks),
        "active_attacks", len(ad.activeAttacks))

    return processedAttacks
}

// postProcessDetections performs post-processing on detected attacks
func (ad *AttackDetector) postProcessDetections(attacks []AttackInfo) []AttackInfo {
    var processedAttacks []AttackInfo

    // 중복 제거
    uniqueAttacks := ad.removeDuplicateAttacks(attacks)

    // 공격 병합 (관련된 공격들을 하나로 병합)
    mergedAttacks := ad.mergeRelatedAttacks(uniqueAttacks)

    // 신뢰도 검증
    for _, attack := range mergedAttacks {
        if ad.verifyAttackConfidence(attack) {
            processedAttacks = append(processedAttacks, attack)
        }
    }

    // 심각도 재평가
    for i := range processedAttacks {
        processedAttacks[i].Severity = ad.reevaluateSeverity(processedAttacks[i])
    }

    return processedAttacks
}

// removeDuplicateAttacks removes duplicate attack detections
func (ad *AttackDetector) removeDuplicateAttacks(attacks []AttackInfo) []AttackInfo {
    seen := make(map[string]bool)
    var unique []AttackInfo

    for _, attack := range attacks {
        // 공격의 고유 키 생성 (타입, 소스, 시간 기반)
        key := fmt.Sprintf("%d_%s_%d", attack.AttackType,
            strings.Join(attack.SourceIPs, ","),
            attack.StartTime.Unix()/60) // 분 단위로 그룹화

        if !seen[key] {
            seen[key] = true
            unique = append(unique, attack)
        }
    }

    return unique
}

// mergeRelatedAttacks merges related attacks into single attack instances
func (ad *AttackDetector) mergeRelatedAttacks(attacks []AttackInfo) []AttackInfo {
    var merged []AttackInfo
    processed := make(map[int]bool)

    for i, attack := range attacks {
        if processed[i] {
            continue
        }

        mergedAttack := attack
        processed[i] = true

        // 관련된 공격들 찾기
        for j := i + 1; j < len(attacks); j++ {
            if processed[j] {
                continue
            }

            if ad.areAttacksRelated(attack, attacks[j]) {
                mergedAttack = ad.mergeAttacks(mergedAttack, attacks[j])
                processed[j] = true
            }
        }

        merged = append(merged, mergedAttack)
    }

    return merged
}

// areAttacksRelated determines if two attacks are related
func (ad *AttackDetector) areAttacksRelated(attack1, attack2 AttackInfo) bool {
    // 같은 타입의 공격인지 확인
    if attack1.AttackType != attack2.AttackType {
        return false
    }

    // 시간적으로 가까운지 확인 (5분 이내)
    timeDiff := attack2.StartTime.Sub(attack1.StartTime)
    if timeDiff < 0 {
        timeDiff = -timeDiff
    }
    if timeDiff > 5*time.Minute {
        return false
    }

    // 공통 소스가 있는지 확인
    sourceOverlap := ad.calculateSourceOverlap(attack1.SourceIPs, attack2.SourceIPs)
    if sourceOverlap > 0.3 { // 30% 이상 겹치면 관련 있다고 판단
        return true
    }

    return false
}

// calculateSourceOverlap calculates overlap ratio between two source lists
func (ad *AttackDetector) calculateSourceOverlap(sources1, sources2 []string) float64 {
    if len(sources1) == 0 || len(sources2) == 0 {
        return 0.0
    }

    set1 := make(map[string]bool)
    for _, source := range sources1 {
        set1[source] = true
    }

    overlap := 0
    for _, source := range sources2 {
        if set1[source] {
            overlap++
        }
    }

    minLength := len(sources1)
    if len(sources2) < minLength {
        minLength = len(sources2)
    }

    return float64(overlap) / float64(minLength)
}

// mergeAttacks merges two related attacks into one
func (ad *AttackDetector) mergeAttacks(attack1, attack2 AttackInfo) AttackInfo {
    merged := attack1

    // 시작 시간을 더 이른 시간으로 설정
    if attack2.StartTime.Before(attack1.StartTime) {
        merged.StartTime = attack2.StartTime
    }

    // 지속 시간 업데이트
    endTime1 := attack1.StartTime.Add(attack1.Duration)
    endTime2 := attack2.StartTime.Add(attack2.Duration)
    if endTime2.After(endTime1) {
        merged.Duration = endTime2.Sub(merged.StartTime)
    }

    // 소스 IP 병합
    sourceSet := make(map[string]bool)
    for _, ip := range attack1.SourceIPs {
        sourceSet[ip] = true
    }
    for _, ip := range attack2.SourceIPs {
        sourceSet[ip] = true
    }

    merged.SourceIPs = make([]string, 0, len(sourceSet))
    for ip := range sourceSet {
        merged.SourceIPs = append(merged.SourceIPs, ip)
    }

    // 볼륨 합계
    merged.Volume = attack1.Volume + attack2.Volume

    // 심각도를 더 높은 것으로 설정
    if attack2.Severity > attack1.Severity {
        merged.Severity = attack2.Severity
    }

    // 신뢰도 평균
    merged.Confidence = (attack1.Confidence + attack2.Confidence) / 2.0

    return merged
}

// verifyAttackConfidence verifies the confidence of an attack detection
func (ad *AttackDetector) verifyAttackConfidence(attack AttackInfo) bool {
    // 최소 신뢰도 임계값 확인
    minConfidence := 0.7 // 70% 이상의 신뢰도 요구
    if attack.Confidence < minConfidence {
        return false
    }

    // 심각도에 따른 신뢰도 요구사항
    switch attack.Severity {
    case AttackSeverityCritical:
        return attack.Confidence >= 0.9 // 90% 이상
    case AttackSeverityHigh:
        return attack.Confidence >= 0.8 // 80% 이상
    case AttackSeverityMedium:
        return attack.Confidence >= 0.7 // 70% 이상
    case AttackSeverityLow:
        return attack.Confidence >= 0.6 // 60% 이상
    default:
        return attack.Confidence >= 0.5 // 50% 이상
    }
}

// reevaluateSeverity reevaluates attack severity based on comprehensive analysis
func (ad *AttackDetector) reevaluateSeverity(attack AttackInfo) AttackSeverity {
    score := 0.0

    // 볼륨 기반 점수
    if attack.Volume > 1000000 { // 1MB 이상
        score += 30.0
    } else if attack.Volume > 100000 { // 100KB 이상
        score += 20.0
    } else if attack.Volume > 10000 { // 10KB 이상
        score += 10.0
    }

    // 지속 시간 기반 점수
    if attack.Duration > 10*time.Minute {
        score += 25.0
    } else if attack.Duration > 5*time.Minute {
        score += 15.0
    } else if attack.Duration > 1*time.Minute {
        score += 10.0
    }

    // 소스 수 기반 점수
    sourceCount := len(attack.SourceIPs)
    if sourceCount > 100 {
        score += 25.0
    } else if sourceCount > 10 {
        score += 15.0
    } else if sourceCount > 1 {
        score += 10.0
    }

    // 공격 타입 기반 점수
    switch attack.AttackType {
    case AttackTypeDistributedFlood:
        score += 30.0
    case AttackTypeBandwidthExhaustion:
        score += 25.0
    case AttackTypeResourceExhaustion:
        score += 25.0
    case AttackTypeConnectionFlood:
        score += 20.0
    case AttackTypeMessageFlood:
        score += 20.0
    default:
        score += 10.0
    }

    // 영향 수준 기반 점수
    if attack.PerformanceImpact != nil {
        if attack.PerformanceImpact.ServiceAvailability < 0.5 {
            score += 30.0
        } else if attack.PerformanceImpact.ServiceAvailability < 0.8 {
            score += 20.0
        } else if attack.PerformanceImpact.ServiceAvailability < 0.95 {
            score += 10.0
        }
    }

    // 점수에 따른 심각도 결정
    switch {
    case score >= 80.0:
        return AttackSeverityCritical
    case score >= 60.0:
        return AttackSeverityHigh
    case score >= 40.0:
        return AttackSeverityMedium
    case score >= 20.0:
        return AttackSeverityLow
    default:
        return AttackSeverityLow
    }
}

// updateActiveAttacks updates the list of active attacks
func (ad *AttackDetector) updateActiveAttacks(newAttacks []AttackInfo) {
    now := time.Now()

    // 새로운 공격들을 활성 공격 목록에 추가
    for _, attack := range newAttacks {
        ad.activeAttacks[attack.AttackID] = &attack
    }

    // 만료된 공격들 제거 (10분 이상 지난 공격)
    for id, attack := range ad.activeAttacks {
        if now.Sub(attack.StartTime.Add(attack.Duration)) > 10*time.Minute {
            delete(ad.activeAttacks, id)
        }
    }
}

// GetActiveAttacks returns currently active attacks
func (ad *AttackDetector) GetActiveAttacks() []AttackInfo {
    ad.mu.RLock()
    defer ad.mu.RUnlock()

    var attacks []AttackInfo
    for _, attack := range ad.activeAttacks {
        attacks = append(attacks, *attack)
    }

    return attacks
}

// GetAttackHistory returns attack detection history
func (ad *AttackDetector) GetAttackHistory() []AttackInfo {
    ad.mu.RLock()
    defer ad.mu.RUnlock()

    // 복사본 반환
    history := make([]AttackInfo, len(ad.detectedAttacks))
    copy(history, ad.detectedAttacks)

    return history
}
```

## ✅ Phase 1 구현 체크리스트 (기본 기능)
- [x] 기본 레이트 리미팅 시스템 설계
- [x] 연결 수 제한 메커니즘 설계
- [x] IP 기반 차단 시스템 설계
- [x] 간단한 트래픽 모니터링 설계

## 🔮 Phase 2-3 구현 예정 (고급 기능)
- [ ] ML 기반 공격 탐지 시스템
- [ ] 복잡한 트래픽 패턴 분석
- [ ] 적응형 방어 전략
- [ ] 공격 패턴 학습 시스템
- [ ] 분산 방어 네트워크
- [ ] 공격 정보 공유 시스템

## 🚀 다음 단계
1. **Phase 1 기본 방어 기능 코드 구현**
2. **Phase 2에서 지능형 탐지 기능 추가**
3. **Phase 3에서 분산 협력 방어 완성**

## 📝 설계 결론
Phase 1에서는 기본적인 DDoS 방어에 집중하여 필수 보안을 확보하고, Phase 2-3에서 지능형 탐지 및 적응형 방어 기능을 점진적으로 추가합니다.
