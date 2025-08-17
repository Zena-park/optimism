package types

import (
	"time"
)

// ChallengerNetworkManagerConfig contains configuration for challenger network manager
type ChallengerNetworkManagerConfig struct {
	// 기본 설정
	MinStakeAmount    uint64        `json:"min_stake_amount"`   // 최소 스테이킹 금액 (Phase 2+)
	MaxPeers          int           `json:"max_peers"`          // 최대 피어 수
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // 하트비트 간격

	// 발견 설정
	DiscoveryEnabled    bool          `json:"discovery_enabled"`     // 발견 활성화
	DiscoveryInterval   time.Duration `json:"discovery_interval"`    // 발견 간격
	BootstrapNodes      []string      `json:"bootstrap_nodes"`       // 부트스트랩 노드들
	MaxDiscoveryRetries int           `json:"max_discovery_retries"` // 최대 발견 재시도

	// 연결 설정
	ConnectionTimeout      time.Duration `json:"connection_timeout"`       // 연결 타임아웃
	ReconnectionEnabled    bool          `json:"reconnection_enabled"`     // 재연결 활성화
	ReconnectionInterval   time.Duration `json:"reconnection_interval"`    // 재연결 간격
	MaxReconnectionRetries int           `json:"max_reconnection_retries"` // 최대 재연결 재시도

	// 상태 관리 설정
	StateUpdateInterval time.Duration `json:"state_update_interval"` // 상태 업데이트 간격
	StateSyncEnabled    bool          `json:"state_sync_enabled"`    // 상태 동기화 활성화
	StateSyncInterval   time.Duration `json:"state_sync_interval"`   // 상태 동기화 간격
	StaleStateDuration  time.Duration `json:"stale_state_duration"`  // 오래된 상태 판단 기간

	// 보안 설정 (Phase 1 기본)
	RateLimitEnabled      bool `json:"rate_limit_enabled"`       // 레이트 리미팅 활성화
	MaxMessagesPerSecond  int  `json:"max_messages_per_second"`  // 초당 최대 메시지 수
	MaxConnectionsPerPeer int  `json:"max_connections_per_peer"` // 피어당 최대 연결 수

	// 성능 설정
	MessageBufferSize int           `json:"message_buffer_size"` // 메시지 버퍼 크기
	WorkerPoolSize    int           `json:"worker_pool_size"`    // 워커 풀 크기
	ProcessingTimeout time.Duration `json:"processing_timeout"`  // 처리 타임아웃
}

// DefaultChallengerNetworkManagerConfig returns default configuration
func DefaultChallengerNetworkManagerConfig() *ChallengerNetworkManagerConfig {
	return &ChallengerNetworkManagerConfig{
		// 기본 설정
		MinStakeAmount:    1000000000000000000, // 1 ETH in wei (Phase 2+)
		MaxPeers:          100,
		HeartbeatInterval: 30 * time.Second,

		// 발견 설정
		DiscoveryEnabled:    true,
		DiscoveryInterval:   60 * time.Second,
		BootstrapNodes:      []string{}, // 실제 환경에서 설정
		MaxDiscoveryRetries: 3,

		// 연결 설정
		ConnectionTimeout:      10 * time.Second,
		ReconnectionEnabled:    true,
		ReconnectionInterval:   30 * time.Second,
		MaxReconnectionRetries: 5,

		// 상태 관리 설정
		StateUpdateInterval: 10 * time.Second,
		StateSyncEnabled:    true,
		StateSyncInterval:   60 * time.Second,
		StaleStateDuration:  300 * time.Second, // 5분

		// 보안 설정 (Phase 1 기본)
		RateLimitEnabled:      true,
		MaxMessagesPerSecond:  10, // 초당 10개 메시지
		MaxConnectionsPerPeer: 5,  // 피어당 최대 5개 연결

		// 성능 설정
		MessageBufferSize: 1000,
		WorkerPoolSize:    10,
		ProcessingTimeout: 5 * time.Second,
	}
}

// ChallengerStateManagerConfig contains configuration for challenger state manager
type ChallengerStateManagerConfig struct {
	// 상태 업데이트 설정
	StateUpdateInterval time.Duration `json:"state_update_interval"` // 상태 업데이트 간격
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`    // 하트비트 간격

	// 동기화 설정
	SyncEnabled    bool          `json:"sync_enabled"`     // 동기화 활성화
	SyncInterval   time.Duration `json:"sync_interval"`    // 동기화 간격
	SyncTimeout    time.Duration `json:"sync_timeout"`     // 동기화 타임아웃
	MaxSyncRetries int           `json:"max_sync_retries"` // 최대 동기화 재시도

	// 데이터 관리 설정
	MaxStates            int           `json:"max_states"`             // 최대 저장 상태 수
	StateRetentionPeriod time.Duration `json:"state_retention_period"` // 상태 보관 기간
	CleanupInterval      time.Duration `json:"cleanup_interval"`       // 정리 간격

	// 성능 설정
	BatchSize         int           `json:"batch_size"`         // 배치 크기
	ProcessingTimeout time.Duration `json:"processing_timeout"` // 처리 타임아웃
}

// DefaultChallengerStateManagerConfig returns default configuration
func DefaultChallengerStateManagerConfig() *ChallengerStateManagerConfig {
	return &ChallengerStateManagerConfig{
		// 상태 업데이트 설정
		StateUpdateInterval: 10 * time.Second,
		HeartbeatInterval:   30 * time.Second,

		// 동기화 설정
		SyncEnabled:    true,
		SyncInterval:   60 * time.Second,
		SyncTimeout:    10 * time.Second,
		MaxSyncRetries: 3,

		// 데이터 관리 설정
		MaxStates:            10000,            // 최대 10,000개 상태 저장
		StateRetentionPeriod: 24 * time.Hour,   // 24시간 보관
		CleanupInterval:      60 * time.Minute, // 1시간마다 정리

		// 성능 설정
		BatchSize:         100,
		ProcessingTimeout: 5 * time.Second,
	}
}

// RateLimiterConfig contains configuration for rate limiter
type RateLimiterConfig struct {
	// 기본 설정
	Enabled               bool `json:"enabled"`                  // 활성화 여부
	MaxMessagesPerSecond  int  `json:"max_messages_per_second"`  // 초당 최대 메시지 수
	MaxRequestsPerSecond  int  `json:"max_requests_per_second"`  // 초당 최대 요청 수 (Phase 1)
	MaxConnectionsPerPeer int  `json:"max_connections_per_peer"` // 피어당 최대 연결 수

	// 버스트 설정
	BurstSize     int           `json:"burst_size"`     // 버스트 크기
	BurstDuration time.Duration `json:"burst_duration"` // 버스트 지속시간

	// 차단 설정
	BlockEnabled  bool          `json:"block_enabled"`  // 차단 활성화
	BlockDuration time.Duration `json:"block_duration"` // 차단 지속시간
	MaxViolations int           `json:"max_violations"` // 최대 위반 횟수

	// 정리 설정
	CleanupInterval time.Duration `json:"cleanup_interval"` // 정리 간격
	EntryTTL        time.Duration `json:"entry_ttl"`        // 엔트리 TTL
	BucketTTL       time.Duration `json:"bucket_ttl"`       // 버킷 TTL (Phase 1)
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		// 기본 설정
		Enabled:               true,
		MaxMessagesPerSecond:  10,
		MaxRequestsPerSecond:  100, // Phase 1: 100 requests per second
		MaxConnectionsPerPeer: 5,

		// 버스트 설정
		BurstSize:     20, // 20개까지 버스트 허용
		BurstDuration: 5 * time.Second,

		// 차단 설정
		BlockEnabled:  false,             // Phase 1: 차단 비활성화
		BlockDuration: 300 * time.Second, // 5분 차단
		MaxViolations: 3,

		// 정리 설정
		CleanupInterval: 60 * time.Second,  // 1분마다 정리
		EntryTTL:        600 * time.Second, // 10분 TTL
		BucketTTL:       600 * time.Second, // Phase 1: 10분 버킷 TTL
	}
}

// ConnectionLimiterConfig contains configuration for connection limiter
type ConnectionLimiterConfig struct {
	// 기본 설정
	Enabled             bool `json:"enabled"`                // 활성화 여부
	MaxTotalConnections int  `json:"max_total_connections"`  // 총 최대 연결 수
	MaxConnections      int  `json:"max_connections"`        // 최대 연결 수 (Phase 1)
	MaxConnectionsPerIP int  `json:"max_connections_per_ip"` // IP당 최대 연결 수

	// 연결 관리
	ConnectionTimeout time.Duration `json:"connection_timeout"`  // 연결 타임아웃
	IdleTimeout       time.Duration `json:"idle_timeout"`        // 유휴 타임아웃
	KeepAliveInterval time.Duration `json:"keep_alive_interval"` // 킵얼라이브 간격

	// 정리 설정
	CleanupInterval time.Duration `json:"cleanup_interval"` // 정리 간격
}

// DefaultConnectionLimiterConfig returns default configuration
func DefaultConnectionLimiterConfig() *ConnectionLimiterConfig {
	return &ConnectionLimiterConfig{
		// 기본 설정
		Enabled:             true,
		MaxTotalConnections: 1000, // 총 1000개 연결
		MaxConnections:      1000, // Phase 1: 1000개 연결
		MaxConnectionsPerIP: 10,   // IP당 10개 연결

		// 연결 관리
		ConnectionTimeout: 30 * time.Second,
		IdleTimeout:       300 * time.Second, // 5분 유휴 타임아웃
		KeepAliveInterval: 60 * time.Second,  // 1분마다 킵얼라이브

		// 정리 설정
		CleanupInterval: 60 * time.Second, // 1분마다 정리
	}
}

// BasicMonitorConfig contains configuration for basic traffic monitor
type BasicMonitorConfig struct {
	// 기본 설정
	Enabled            bool          `json:"enabled"`             // 활성화 여부
	MonitoringInterval time.Duration `json:"monitoring_interval"` // 모니터링 간격
	ReportingInterval  time.Duration `json:"reporting_interval"`  // 리포팅 간격
	UpdateInterval     time.Duration `json:"update_interval"`     // 업데이트 간격 (Phase 1)
	SnapshotInterval   time.Duration `json:"snapshot_interval"`   // 스냅샷 간격 (Phase 1)

	// 임계값 설정
	HighTrafficThreshold       float64 `json:"high_traffic_threshold"`       // 높은 트래픽 임계값 (msg/sec)
	SuspiciousPatternThreshold float64 `json:"suspicious_pattern_threshold"` // 의심스러운 패턴 임계값
	MaxMessagesPerSecond       int     `json:"max_messages_per_second"`      // 최대 메시지/초 (Phase 1)
	MaxBandwidthMbps           int     `json:"max_bandwidth_mbps"`           // 최대 대역폭 (Phase 1)
	MaxErrorRate               float64 `json:"max_error_rate"`               // 최대 에러율 (Phase 1)
	MaxConnections             int     `json:"max_connections"`              // 최대 연결 수 (Phase 1)

	// 데이터 보관
	DataRetentionPeriod time.Duration `json:"data_retention_period"` // 데이터 보관 기간
	MaxDataPoints       int           `json:"max_data_points"`       // 최대 데이터 포인트
	HistorySize         int           `json:"history_size"`          // 히스토리 크기 (Phase 1)
	AlertCooldown       time.Duration `json:"alert_cooldown"`        // 알림 쿨다운 (Phase 1)
}

// DefaultBasicMonitorConfig returns default configuration
func DefaultBasicMonitorConfig() *BasicMonitorConfig {
	return &BasicMonitorConfig{
		// 기본 설정
		Enabled:            true,
		MonitoringInterval: 10 * time.Second,
		ReportingInterval:  60 * time.Second,
		UpdateInterval:     10 * time.Second, // Phase 1: 10초 업데이트
		SnapshotInterval:   1 * time.Minute,  // Phase 1: 1분 스냅샷

		// 임계값 설정
		HighTrafficThreshold:       50.0,  // 초당 50개 메시지
		SuspiciousPatternThreshold: 100.0, // 초당 100개 메시지
		MaxMessagesPerSecond:       1000,  // Phase 1: 1000 msg/sec
		MaxBandwidthMbps:           100,   // Phase 1: 100 Mbps
		MaxErrorRate:               5.0,   // Phase 1: 5% 에러율
		MaxConnections:             800,   // Phase 1: 800 연결

		// 데이터 보관
		DataRetentionPeriod: 24 * time.Hour,  // 24시간
		MaxDataPoints:       8640,            // 24시간 * 60분 * 6 (10초 간격)
		HistorySize:         60,              // Phase 1: 60개 스냅샷 (1시간)
		AlertCooldown:       5 * time.Minute, // Phase 1: 5분 쿨다운
	}
}
