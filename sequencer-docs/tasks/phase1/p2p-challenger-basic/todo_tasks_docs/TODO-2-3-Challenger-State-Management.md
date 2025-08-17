# TODO 2.3: 챌린저 상태 관리

## 📋 작업 개요
- **작업명**: 챌린저 상태 관리
- **상태**: 🔄 진행중
- **담당**: Phase 1 - P2P 챌린저 네트워크 구축

## 🎯 작업 목표
챌린저의 상태를 체계적으로 추적하고 관리하는 시스템을 구현한다.

## 📐 구현 요구사항

### **🚀 Phase 1: 기본 상태 관리 (구현 대상)**
- ✅ 기본 온라인/오프라인 상태 관리
- ✅ 간단한 연결 상태 추적
- ✅ 기본 상태 동기화

### **🔮 Phase 2-3: 고급 상태 관리 (향후 구현)**
- 🔄 복잡한 활성도 모니터링 및 패턴 분석
- 🔄 성과 데이터 수집 및 트렌드 분석
- 🔄 지능형 상태 동기화 및 충돌 해결
- 🔄 상태 이력 관리 및 분석

## 🏗️ 시스템 설계

### 1. 통합 상태 관리 시스템

#### **1.1 ChallengerStateManager**
```go
// ChallengerStateManager manages comprehensive challenger state information
type ChallengerStateManager struct {
    // 기본 정보
    nodeID          string
    challengerID    string
    challengerNode  *ChallengerP2PNode

    // 상태 관리 컴포넌트들
    statusTracker   *StatusTracker         // 기본 상태 추적 (TODO 2.1에서 구현)
    activityMonitor *ActivityMonitor       // 활성도 모니터링
    performanceTracker *PerformanceTracker // 성과 데이터 수집
    stateSync       *StateSync             // 상태 동기화

    // 상태 저장소
    localState      *ChallengerState       // 로컬 상태
    remoteStates    map[string]*ChallengerState // 다른 챌린저들의 상태
    stateHistory    *StateHistory          // 상태 변경 이력

    // 동기화 관리
    syncManager     *StateSyncManager      // 상태 동기화 관리자
    conflictResolver *ConflictResolver     // 충돌 해결

    // 이벤트 및 알림
    eventBus        *EventBus              // 이벤트 버스
    notifications   *NotificationManager   // 알림 관리자

    // 설정 및 제어
    config          *StateManagerConfig    // 설정
    running         bool                   // 실행 상태

    // 동시성 제어
    mu              sync.RWMutex           // 읽기/쓰기 뮤텍스
    ctx             context.Context        // 컨텍스트
    cancel          context.CancelFunc     // 취소 함수

    // 로깅 및 메트릭스
    logger          log.Logger             // 로거
    metrics         *StateMetrics          // 상태 메트릭스
}

// ChallengerState represents comprehensive state of a challenger
type ChallengerState struct {
    // 기본 정보
    ChallengerID    string            `json:"challenger_id"`
    NodeID          string            `json:"node_id"`
    Address         string            `json:"address"`

    // 상태 정보
    Status          ChallengerStatus  `json:"status"`
    OnlineStatus    OnlineStatus      `json:"online_status"`
    ActivityLevel   ActivityLevel     `json:"activity_level"`

    // 성과 정보
    Performance     *PerformanceData  `json:"performance"`
    Reputation      *ReputationScore  `json:"reputation"`

    // 네트워크 정보
    NetworkInfo     *NetworkInfo      `json:"network_info"`

    // 게임 참여 정보
    GameParticipation *GameParticipation `json:"game_participation"`

    // 메타데이터
    Version         string            `json:"version"`
    LastUpdated     time.Time         `json:"last_updated"`
    UpdateSequence  uint64            `json:"update_sequence"`
    Signature       []byte            `json:"signature"`
}

// OnlineStatus represents detailed online status
type OnlineStatus struct {
    IsOnline        bool              `json:"is_online"`
    LastSeen        time.Time         `json:"last_seen"`
    SessionStart    time.Time         `json:"session_start"`
    TotalUptime     time.Duration     `json:"total_uptime"`
    UptimeToday     time.Duration     `json:"uptime_today"`
    ConnectionCount int               `json:"connection_count"`
}

// ActivityLevel represents challenger activity level
type ActivityLevel struct {
    Level           ActivityLevelType `json:"level"`
    Score           float64           `json:"score"`           // 0-100
    LastActivity    time.Time         `json:"last_activity"`
    RecentActions   []ActivityAction  `json:"recent_actions"`
    TrendDirection  TrendDirection    `json:"trend_direction"` // 증가/감소/안정
}

// ActivityLevelType represents activity level categories
type ActivityLevelType int

const (
    ActivityLevelInactive ActivityLevelType = iota // 비활성
    ActivityLevelLow                               // 낮음
    ActivityLevelModerate                          // 보통
    ActivityLevelHigh                              // 높음
    ActivityLevelVeryHigh                          // 매우 높음
)

// ActivityAction represents a single activity action
type ActivityAction struct {
    Type        ActivityType  `json:"type"`
    Timestamp   time.Time     `json:"timestamp"`
    Details     interface{}   `json:"details"`
    Impact      float64       `json:"impact"`     // 활동 점수에 미치는 영향
}

// ActivityType represents types of activities
type ActivityType int

const (
    ActivityTypeGameParticipation ActivityType = iota
    ActivityTypeAttentionTest
    ActivityTypeMessageSent
    ActivityTypeConnectionMade
    ActivityTypeValidationPerformed
    ActivityTypeStateUpdate
)

// TrendDirection represents trend direction
type TrendDirection int

const (
    TrendDirectionDecreasing TrendDirection = iota
    TrendDirectionStable
    TrendDirectionIncreasing
)

// PerformanceData represents performance metrics
type PerformanceData struct {
    // 테스트 성과
    TotalTests      uint64        `json:"total_tests"`
    PassedTests     uint64        `json:"passed_tests"`
    FailedTests     uint64        `json:"failed_tests"`
    SuccessRate     float64       `json:"success_rate"`

    // 응답 성과
    AvgResponseTime time.Duration `json:"avg_response_time"`
    MinResponseTime time.Duration `json:"min_response_time"`
    MaxResponseTime time.Duration `json:"max_response_time"`

    // 가용성 성과
    Uptime          time.Duration `json:"uptime"`
    Downtime        time.Duration `json:"downtime"`
    AvailabilityRate float64      `json:"availability_rate"`

    // 네트워크 성과
    MessagesSent    uint64        `json:"messages_sent"`
    MessagesReceived uint64       `json:"messages_received"`
    NetworkLatency  time.Duration `json:"network_latency"`

    // 시간 정보
    LastUpdated     time.Time     `json:"last_updated"`
    ReportingPeriod time.Duration `json:"reporting_period"`
}

// NetworkInfo represents network-related information
type NetworkInfo struct {
    // 연결 정보
    ActiveConnections   int               `json:"active_connections"`
    TotalConnections    int               `json:"total_connections"`
    PeerConnections     map[string]string `json:"peer_connections"` // challengerID -> address

    // 품질 정보
    AverageLatency      time.Duration     `json:"average_latency"`
    PacketLossRate      float64           `json:"packet_loss_rate"`
    Bandwidth           uint64            `json:"bandwidth"`

    // 지리적 정보
    Region              string            `json:"region"`
    Country             string            `json:"country"`

    // 네트워크 타입
    NetworkType         NetworkType       `json:"network_type"`

    LastUpdated         time.Time         `json:"last_updated"`
}

// NetworkType represents network connection type
type NetworkType int

const (
    NetworkTypeUnknown NetworkType = iota
    NetworkTypeEthernet
    NetworkTypeWiFi
    NetworkTypeMobile
    NetworkTypeVPN
)

// GameParticipation represents game participation information
type GameParticipation struct {
    // 현재 참여 게임들
    ActiveGames         []string          `json:"active_games"`

    // 게임 통계
    TotalGamesJoined    uint64            `json:"total_games_joined"`
    GamesWon            uint64            `json:"games_won"`
    GamesLost           uint64            `json:"games_lost"`
    WinRate             float64           `json:"win_rate"`

    // 최근 게임 활동
    RecentGames         []GameRecord      `json:"recent_games"`
    LastGameTime        time.Time         `json:"last_game_time"`

    // 선호도 및 패턴
    PreferredGameTypes  []string          `json:"preferred_game_types"`
    PlayingPattern      PlayingPattern    `json:"playing_pattern"`

    LastUpdated         time.Time         `json:"last_updated"`
}

// GameRecord represents a single game record
type GameRecord struct {
    GameID      string        `json:"game_id"`
    GameType    string        `json:"game_type"`
    Role        string        `json:"role"`
    Result      GameResult    `json:"result"`
    Duration    time.Duration `json:"duration"`
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
}

// GameResult represents the result of a game
type GameResult int

const (
    GameResultPending GameResult = iota
    GameResultWon
    GameResultLost
    GameResultDraw
    GameResultAborted
)

// PlayingPattern represents playing behavior patterns
type PlayingPattern struct {
    AverageSessionDuration time.Duration     `json:"average_session_duration"`
    PreferredPlayingHours  []int            `json:"preferred_playing_hours"` // 0-23
    PlayingFrequency       PlayingFrequency `json:"playing_frequency"`
    ConsistencyScore       float64          `json:"consistency_score"`       // 0-100
}

// PlayingFrequency represents how frequently a challenger plays
type PlayingFrequency int

const (
    PlayingFrequencyRare      PlayingFrequency = iota // 드물게
    PlayingFrequencyOccasional                        // 가끔
    PlayingFrequencyRegular                           // 정기적
    PlayingFrequencyFrequent                          // 자주
    PlayingFrequencyConstant                          // 지속적
)

// StateManagerConfig contains configuration for state manager
type StateManagerConfig struct {
    // 상태 업데이트 설정
    StateUpdateInterval     time.Duration `json:"state_update_interval"`      // 상태 업데이트 간격
    ActivityCheckInterval   time.Duration `json:"activity_check_interval"`    // 활동 체크 간격
    PerformanceUpdateInterval time.Duration `json:"performance_update_interval"` // 성과 업데이트 간격

    // 동기화 설정
    SyncEnabled            bool          `json:"sync_enabled"`               // 동기화 활성화
    SyncInterval           time.Duration `json:"sync_interval"`              // 동기화 간격
    SyncTimeout            time.Duration `json:"sync_timeout"`               // 동기화 타임아웃
    MaxSyncRetries         int           `json:"max_sync_retries"`           // 최대 동기화 재시도

    // 이력 관리 설정
    HistoryEnabled         bool          `json:"history_enabled"`            // 이력 저장 활성화
    HistoryRetentionPeriod time.Duration `json:"history_retention_period"`   // 이력 보관 기간
    MaxHistoryEntries      int           `json:"max_history_entries"`        // 최대 이력 항목 수

    // 성능 설정
    ActivityScoreDecayRate float64       `json:"activity_score_decay_rate"`  // 활동 점수 감소율
    PerformanceWindowSize  time.Duration `json:"performance_window_size"`    // 성과 측정 윈도우 크기

    // 충돌 해결 설정
    ConflictResolutionEnabled bool       `json:"conflict_resolution_enabled"` // 충돌 해결 활성화
    ConflictResolutionStrategy ConflictResolutionStrategy `json:"conflict_resolution_strategy"`
}

// ConflictResolutionStrategy represents strategy for resolving state conflicts
type ConflictResolutionStrategy int

const (
    ConflictResolutionLastWriter ConflictResolutionStrategy = iota // 마지막 쓰기 우선
    ConflictResolutionHighestSequence                              // 높은 시퀀스 번호 우선
    ConflictResolutionMajorityVote                                 // 다수결
    ConflictResolutionTrustedNode                                  // 신뢰 노드 우선
)
```

### 2. 활성도 모니터링 시스템

#### **2.1 ActivityMonitor**
```go
// ActivityMonitor monitors challenger activity and calculates activity levels
type ActivityMonitor struct {
    // 기본 정보
    challengerID    string
    stateManager    *ChallengerStateManager

    // 활동 추적
    activityBuffer  *ActivityBuffer        // 활동 버퍼
    activityScore   *ActivityScore         // 활동 점수 계산기

    // 패턴 분석
    patternAnalyzer *ActivityPatternAnalyzer // 패턴 분석기

    // 설정
    config          *ActivityMonitorConfig  // 설정

    // 상태 관리
    running         bool                    // 실행 상태
    lastUpdate      time.Time               // 마지막 업데이트

    // 동시성 제어
    mu              sync.RWMutex            // 뮤텍스
    ctx             context.Context         // 컨텍스트
    cancel          context.CancelFunc      // 취소 함수

    logger          log.Logger              // 로거
}

// ActivityBuffer manages recent activity records
type ActivityBuffer struct {
    activities      []ActivityAction        // 활동 기록들
    maxSize         int                     // 최대 크기
    totalScore      float64                 // 총 점수
    mu              sync.RWMutex            // 뮤텍스
}

// ActivityScore calculates activity scores based on various factors
type ActivityScore struct {
    // 점수 가중치
    gameParticipationWeight float64         // 게임 참여 가중치
    attentionTestWeight     float64         // 어텐션 테스트 가중치
    communicationWeight     float64         // 커뮤니케이션 가중치
    uptimeWeight           float64         // 업타임 가중치

    // 점수 계산 설정
    decayRate              float64         // 감소율 (시간에 따른)
    maxScore               float64         // 최대 점수

    // 시간 윈도우
    evaluationWindow       time.Duration    // 평가 윈도우

    mu                     sync.RWMutex     // 뮤텍스
}

// ActivityPatternAnalyzer analyzes activity patterns and trends
type ActivityPatternAnalyzer struct {
    // 패턴 데이터
    hourlyActivity    [24]float64          // 시간별 활동 패턴
    dailyActivity     [7]float64           // 요일별 활동 패턴
    weeklyTrend       []float64            // 주간 트렌드

    // 분석 결과
    peakHours         []int                // 피크 시간대
    preferredDays     []time.Weekday       // 선호 요일
    consistencyScore  float64              // 일관성 점수
    trendDirection    TrendDirection       // 트렌드 방향

    // 설정
    analysisWindow    time.Duration        // 분석 윈도우
    minDataPoints     int                  // 최소 데이터 포인트

    mu                sync.RWMutex         // 뮤텍스
    logger            log.Logger           // 로거
}

// ActivityMonitorConfig contains configuration for activity monitor
type ActivityMonitorConfig struct {
    // 모니터링 설정
    MonitoringEnabled       bool          `json:"monitoring_enabled"`       // 모니터링 활성화
    UpdateInterval          time.Duration `json:"update_interval"`          // 업데이트 간격
    ActivityBufferSize      int           `json:"activity_buffer_size"`     // 활동 버퍼 크기

    // 점수 계산 설정
    ScoreDecayRate          float64       `json:"score_decay_rate"`         // 점수 감소율
    MaxActivityScore        float64       `json:"max_activity_score"`       // 최대 활동 점수
    EvaluationWindow        time.Duration `json:"evaluation_window"`        // 평가 윈도우

    // 가중치 설정
    GameParticipationWeight float64       `json:"game_participation_weight"` // 게임 참여 가중치
    AttentionTestWeight     float64       `json:"attention_test_weight"`     // 어텐션 테스트 가중치
    CommunicationWeight     float64       `json:"communication_weight"`     // 커뮤니케이션 가중치
    UptimeWeight           float64       `json:"uptime_weight"`            // 업타임 가중치

    // 패턴 분석 설정
    PatternAnalysisEnabled  bool          `json:"pattern_analysis_enabled"` // 패턴 분석 활성화
    PatternAnalysisWindow   time.Duration `json:"pattern_analysis_window"`  // 패턴 분석 윈도우
    MinDataPointsForAnalysis int          `json:"min_data_points_for_analysis"` // 분석용 최소 데이터 포인트
}

// RecordActivity records a new activity
func (am *ActivityMonitor) RecordActivity(activityType ActivityType, details interface{}) {
    am.mu.Lock()
    defer am.mu.Unlock()

    activity := ActivityAction{
        Type:      activityType,
        Timestamp: time.Now(),
        Details:   details,
        Impact:    am.calculateActivityImpact(activityType, details),
    }

    // 활동 버퍼에 추가
    am.activityBuffer.AddActivity(activity)

    // 활동 점수 업데이트
    am.updateActivityScore()

    // 패턴 분석 업데이트
    if am.config.PatternAnalysisEnabled {
        am.patternAnalyzer.UpdatePattern(activity)
    }

    am.logger.Debug("Activity recorded",
        "type", activityType, "impact", activity.Impact, "challenger", am.challengerID)
}

// calculateActivityImpact calculates the impact score of an activity
func (am *ActivityMonitor) calculateActivityImpact(activityType ActivityType, details interface{}) float64 {
    baseScore := 1.0

    switch activityType {
    case ActivityTypeGameParticipation:
        baseScore = 10.0 // 게임 참여는 높은 점수
    case ActivityTypeAttentionTest:
        baseScore = 5.0  // 어텐션 테스트는 중간 점수
    case ActivityTypeMessageSent:
        baseScore = 1.0  // 메시지 전송은 낮은 점수
    case ActivityTypeConnectionMade:
        baseScore = 2.0  // 연결 생성은 낮은 점수
    case ActivityTypeValidationPerformed:
        baseScore = 3.0  // 검증 수행은 중간 점수
    case ActivityTypeStateUpdate:
        baseScore = 1.0  // 상태 업데이트는 낮은 점수
    }

    // 세부 정보에 따른 추가 점수 계산
    if details != nil {
        // 구체적인 세부 정보에 따라 점수 조정
        // 예: 게임 결과, 테스트 성공/실패 등
    }

    return baseScore
}

// updateActivityScore updates the overall activity score
func (am *ActivityMonitor) updateActivityScore() {
    now := time.Now()

    // 시간 기반 감소 적용
    timeSinceLastUpdate := now.Sub(am.lastUpdate)
    decayFactor := math.Exp(-am.config.ScoreDecayRate * timeSinceLastUpdate.Hours())

    // 현재 활동 점수 계산
    currentScore := am.activityBuffer.GetTotalScore()

    // 감소 적용된 점수 계산
    adjustedScore := currentScore * decayFactor

    // 최대 점수 제한
    if adjustedScore > am.config.MaxActivityScore {
        adjustedScore = am.config.MaxActivityScore
    }

    // 활동 레벨 결정
    activityLevel := am.determineActivityLevel(adjustedScore)

    // 상태 매니저에 업데이트
    am.stateManager.UpdateActivityLevel(activityLevel, adjustedScore)

    am.lastUpdate = now
}

// determineActivityLevel determines activity level based on score
func (am *ActivityMonitor) determineActivityLevel(score float64) ActivityLevelType {
    maxScore := am.config.MaxActivityScore
    ratio := score / maxScore

    switch {
    case ratio >= 0.8:
        return ActivityLevelVeryHigh
    case ratio >= 0.6:
        return ActivityLevelHigh
    case ratio >= 0.4:
        return ActivityLevelModerate
    case ratio >= 0.2:
        return ActivityLevelLow
    default:
        return ActivityLevelInactive
    }
}

// GetActivityLevel returns current activity level
func (am *ActivityMonitor) GetActivityLevel() ActivityLevel {
    am.mu.RLock()
    defer am.mu.RUnlock()

    score := am.activityBuffer.GetTotalScore()
    level := am.determineActivityLevel(score)

    // 최근 활동들 가져오기
    recentActivities := am.activityBuffer.GetRecentActivities(10)

    // 트렌드 방향 계산
    trendDirection := am.patternAnalyzer.GetTrendDirection()

    return ActivityLevel{
        Level:          level,
        Score:          score,
        LastActivity:   am.getLastActivityTime(),
        RecentActions:  recentActivities,
        TrendDirection: trendDirection,
    }
}

// getLastActivityTime returns the timestamp of the last activity
func (am *ActivityMonitor) getLastActivityTime() time.Time {
    activities := am.activityBuffer.GetRecentActivities(1)
    if len(activities) > 0 {
        return activities[0].Timestamp
    }
    return time.Time{}
}
```

### 3. 성과 추적 시스템

#### **3.1 PerformanceTracker**
```go
// PerformanceTracker tracks and analyzes challenger performance metrics
type PerformanceTracker struct {
    // 기본 정보
    challengerID    string
    stateManager    *ChallengerStateManager

    // 성과 데이터 수집기들
    testTracker     *TestPerformanceTracker     // 테스트 성과 추적
    responseTracker *ResponsePerformanceTracker // 응답 성과 추적
    uptimeTracker   *UptimeTracker             // 가용성 추적
    networkTracker  *NetworkPerformanceTracker  // 네트워크 성과 추적

    // 성과 분석기
    analyzer        *PerformanceAnalyzer       // 성과 분석기

    // 데이터 저장소
    performanceData *PerformanceData           // 현재 성과 데이터
    historicalData  *PerformanceHistory        // 이력 데이터

    // 설정
    config          *PerformanceTrackerConfig  // 설정

    // 상태 관리
    running         bool                       // 실행 상태
    lastUpdate      time.Time                  // 마지막 업데이트

    // 동시성 제어
    mu              sync.RWMutex               // 뮤텍스
    ctx             context.Context            // 컨텍스트
    cancel          context.CancelFunc         // 취소 함수

    logger          log.Logger                 // 로거
}

// TestPerformanceTracker tracks test-related performance
type TestPerformanceTracker struct {
    totalTests    uint64                       // 총 테스트 수
    passedTests   uint64                       // 통과한 테스트 수
    failedTests   uint64                       // 실패한 테스트 수
    testHistory   []TestResult                 // 테스트 결과 이력

    mu            sync.RWMutex                 // 뮤텍스
}

// TestResult represents a single test result
type TestResult struct {
    TestID        string        `json:"test_id"`
    TestType      string        `json:"test_type"`
    Success       bool          `json:"success"`
    ResponseTime  time.Duration `json:"response_time"`
    Timestamp     time.Time     `json:"timestamp"`
    ErrorMessage  string        `json:"error_message,omitempty"`
}

// ResponsePerformanceTracker tracks response time performance
type ResponsePerformanceTracker struct {
    responseTimes    []time.Duration            // 응답 시간들
    minResponseTime  time.Duration              // 최소 응답 시간
    maxResponseTime  time.Duration              // 최대 응답 시간
    avgResponseTime  time.Duration              // 평균 응답 시간

    // 응답 시간 분포
    responseDistribution map[time.Duration]int  // 응답 시간 분포

    mu               sync.RWMutex               // 뮤텍스
}

// UptimeTracker tracks uptime and availability
type UptimeTracker struct {
    sessionStart     time.Time                 // 세션 시작 시간
    totalUptime      time.Duration             // 총 업타임
    totalDowntime    time.Duration             // 총 다운타임
    uptimeHistory    []UptimeRecord            // 업타임 기록

    // 가용성 계산
    availabilityRate float64                   // 가용성 비율

    mu               sync.RWMutex               // 뮤텍스
}

// UptimeRecord represents an uptime/downtime record
type UptimeRecord struct {
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
    Duration    time.Duration `json:"duration"`
    IsUptime    bool          `json:"is_uptime"`
    Reason      string        `json:"reason,omitempty"`
}

// NetworkPerformanceTracker tracks network-related performance
type NetworkPerformanceTracker struct {
    messagesSent     uint64                    // 전송된 메시지 수
    messagesReceived uint64                    // 수신된 메시지 수
    bytesTransferred uint64                    // 전송된 바이트 수
    networkLatency   time.Duration             // 네트워크 지연시간

    // 네트워크 품질 메트릭스
    packetLossRate   float64                   // 패킷 손실률
    throughput       uint64                    // 처리량 (bytes/sec)

    mu               sync.RWMutex               // 뮤텍스
}

// PerformanceAnalyzer analyzes performance data and generates insights
type PerformanceAnalyzer struct {
    // 분석 결과
    overallScore     float64                   // 전체 성과 점수
    trends          *PerformanceTrends         // 성과 트렌드
    recommendations []PerformanceRecommendation // 개선 권장사항

    // 분석 설정
    analysisWindow   time.Duration             // 분석 윈도우
    benchmarks      *PerformanceBenchmarks     // 성과 벤치마크

    mu              sync.RWMutex               // 뮤텍스
    logger          log.Logger                 // 로거
}

// PerformanceTrends represents performance trend analysis
type PerformanceTrends struct {
    TestSuccessRate    TrendAnalysis `json:"test_success_rate"`
    ResponseTime       TrendAnalysis `json:"response_time"`
    Availability       TrendAnalysis `json:"availability"`
    NetworkPerformance TrendAnalysis `json:"network_performance"`
}

// TrendAnalysis represents trend analysis for a specific metric
type TrendAnalysis struct {
    Direction       TrendDirection `json:"direction"`        // 트렌드 방향
    Magnitude       float64        `json:"magnitude"`        // 변화 크기
    Confidence      float64        `json:"confidence"`       // 신뢰도
    PredictedValue  float64        `json:"predicted_value"`  // 예측 값
    LastUpdated     time.Time      `json:"last_updated"`     // 마지막 업데이트
}

// PerformanceRecommendation represents a performance improvement recommendation
type PerformanceRecommendation struct {
    Category    RecommendationCategory `json:"category"`
    Priority    RecommendationPriority `json:"priority"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Action      string                 `json:"action"`
    Impact      string                 `json:"impact"`
    CreatedAt   time.Time              `json:"created_at"`
}

// RecommendationCategory represents recommendation categories
type RecommendationCategory int

const (
    RecommendationCategoryResponse RecommendationCategory = iota
    RecommendationCategoryAvailability
    RecommendationCategoryNetwork
    RecommendationCategoryTesting
    RecommendationCategoryGeneral
)

// RecommendationPriority represents recommendation priority levels
type RecommendationPriority int

const (
    RecommendationPriorityLow RecommendationPriority = iota
    RecommendationPriorityMedium
    RecommendationPriorityHigh
    RecommendationPriorityCritical
)

// PerformanceBenchmarks contains performance benchmarks for comparison
type PerformanceBenchmarks struct {
    TargetSuccessRate      float64       `json:"target_success_rate"`      // 목표 성공률
    TargetResponseTime     time.Duration `json:"target_response_time"`     // 목표 응답 시간
    TargetAvailability     float64       `json:"target_availability"`      // 목표 가용성
    TargetNetworkLatency   time.Duration `json:"target_network_latency"`   // 목표 네트워크 지연시간

    // 업계 평균 값들
    IndustryAvgSuccessRate float64       `json:"industry_avg_success_rate"`
    IndustryAvgResponseTime time.Duration `json:"industry_avg_response_time"`
    IndustryAvgAvailability float64       `json:"industry_avg_availability"`
}

// PerformanceTrackerConfig contains configuration for performance tracker
type PerformanceTrackerConfig struct {
    // 추적 설정
    TrackingEnabled        bool          `json:"tracking_enabled"`         // 추적 활성화
    UpdateInterval         time.Duration `json:"update_interval"`          // 업데이트 간격
    DataRetentionPeriod    time.Duration `json:"data_retention_period"`    // 데이터 보관 기간

    // 분석 설정
    AnalysisEnabled        bool          `json:"analysis_enabled"`         // 분석 활성화
    AnalysisInterval       time.Duration `json:"analysis_interval"`        // 분석 간격
    AnalysisWindow         time.Duration `json:"analysis_window"`          // 분석 윈도우

    // 벤치마크 설정
    TargetSuccessRate      float64       `json:"target_success_rate"`      // 목표 성공률
    TargetResponseTime     time.Duration `json:"target_response_time"`     // 목표 응답 시간
    TargetAvailability     float64       `json:"target_availability"`      // 목표 가용성

    // 알림 설정
    AlertEnabled           bool          `json:"alert_enabled"`            // 알림 활성화
    PerformanceThreshold   float64       `json:"performance_threshold"`    // 성과 임계값
    AlertCooldown          time.Duration `json:"alert_cooldown"`           // 알림 쿨다운
}

// RecordTestResult records a test result
func (pt *PerformanceTracker) RecordTestResult(result TestResult) {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    // 테스트 추적기에 기록
    pt.testTracker.RecordTest(result)

    // 응답 시간 추적기에 기록
    pt.responseTracker.RecordResponseTime(result.ResponseTime)

    // 성과 데이터 업데이트
    pt.updatePerformanceData()

    pt.logger.Debug("Test result recorded",
        "test_id", result.TestID, "success", result.Success,
        "response_time", result.ResponseTime, "challenger", pt.challengerID)
}

// RecordNetworkActivity records network activity
func (pt *PerformanceTracker) RecordNetworkActivity(sent, received uint64, latency time.Duration) {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    // 네트워크 추적기에 기록
    pt.networkTracker.RecordActivity(sent, received, latency)

    // 성과 데이터 업데이트
    pt.updatePerformanceData()
}

// RecordUptimeEvent records an uptime/downtime event
func (pt *PerformanceTracker) RecordUptimeEvent(isUptime bool, duration time.Duration, reason string) {
    pt.mu.Lock()
    defer pt.mu.Unlock()

    // 업타임 추적기에 기록
    pt.uptimeTracker.RecordEvent(isUptime, duration, reason)

    // 성과 데이터 업데이트
    pt.updatePerformanceData()
}

// updatePerformanceData updates the current performance data
func (pt *PerformanceTracker) updatePerformanceData() {
    // 각 추적기에서 데이터 수집
    testData := pt.testTracker.GetSummary()
    responseData := pt.responseTracker.GetSummary()
    uptimeData := pt.uptimeTracker.GetSummary()
    networkData := pt.networkTracker.GetSummary()

    // 성과 데이터 업데이트
    pt.performanceData = &PerformanceData{
        TotalTests:       testData.TotalTests,
        PassedTests:      testData.PassedTests,
        FailedTests:      testData.FailedTests,
        SuccessRate:      testData.SuccessRate,
        AvgResponseTime:  responseData.AvgResponseTime,
        MinResponseTime:  responseData.MinResponseTime,
        MaxResponseTime:  responseData.MaxResponseTime,
        Uptime:          uptimeData.TotalUptime,
        Downtime:        uptimeData.TotalDowntime,
        AvailabilityRate: uptimeData.AvailabilityRate,
        MessagesSent:     networkData.MessagesSent,
        MessagesReceived: networkData.MessagesReceived,
        NetworkLatency:   networkData.NetworkLatency,
        LastUpdated:      time.Now(),
        ReportingPeriod:  pt.config.AnalysisWindow,
    }

    // 상태 매니저에 업데이트
    pt.stateManager.UpdatePerformanceData(pt.performanceData)
}

// GetPerformanceData returns current performance data
func (pt *PerformanceTracker) GetPerformanceData() *PerformanceData {
    pt.mu.RLock()
    defer pt.mu.RUnlock()

    // 복사본 반환
    data := *pt.performanceData
    return &data
}

// AnalyzePerformance performs comprehensive performance analysis
func (pt *PerformanceTracker) AnalyzePerformance() *PerformanceAnalysis {
    pt.mu.RLock()
    defer pt.mu.RUnlock()

    return pt.analyzer.AnalyzePerformance(pt.performanceData, pt.historicalData)
}
```

## ✅ Phase 1 구현 체크리스트 (기본 기능)
- [x] 기본 ChallengerState 구조 정의
- [x] 온라인/오프라인 상태 관리 설계
- [x] 간단한 연결 상태 추적 설계
- [x] 기본 상태 동기화 메커니즘 설계

## 🔮 Phase 2-3 구현 예정 (고급 기능)
- [ ] ActivityMonitor 활성도 모니터링 시스템
- [ ] PerformanceTracker 성과 추적 시스템
- [ ] 고급 상태 동기화 및 충돌 해결
- [ ] 상태 이력 관리 및 분석
- [ ] 이벤트 기반 아키텍처

## 🚀 다음 단계
1. **Phase 1 기본 기능 코드 구현**
2. **Phase 2에서 고급 기능 추가**
3. **전체 시스템 통합 및 최적화**

## 📝 설계 결론
Phase 1에서는 기본적인 챌린저 상태 관리에 집중하여 안정적인 기반을 구축하고, Phase 2-3에서 고급 모니터링 및 분석 기능을 점진적으로 추가합니다.
