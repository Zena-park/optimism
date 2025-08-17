package defense

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum/go-ethereum/log"
)

// BasicMonitor implements basic traffic monitoring for P2P network (Phase 1: simple metrics collection)
type BasicMonitor struct {
	// Configuration
	config *types.BasicMonitorConfig // Monitor configuration
	logger log.Logger                // Logger

	// Traffic statistics
	stats     *TrafficStats      // Current traffic statistics
	history   []*TrafficSnapshot // Historical snapshots
	historyMu sync.RWMutex       // History mutex

	// Monitoring state
	isRunning      bool         // Whether the monitor is running
	monitorTicker  *time.Ticker // Monitoring ticker
	snapshotTicker *time.Ticker // Snapshot ticker
	mu             sync.RWMutex // Concurrency control

	// Alert thresholds (Phase 1: basic)
	alertThresholds *AlertThresholds // Alert thresholds
	lastAlert       time.Time        // Last alert time
}

// TrafficStats represents current traffic statistics
type TrafficStats struct {
	// Message statistics
	TotalMessages     int64            `json:"total_messages"`      // Total messages processed
	MessagesByType    map[string]int64 `json:"messages_by_type"`    // Messages by type
	MessagesPerSecond float64          `json:"messages_per_second"` // Current rate

	// Bandwidth statistics
	TotalBytesIn     int64   `json:"total_bytes_in"`     // Total bytes received
	TotalBytesOut    int64   `json:"total_bytes_out"`    // Total bytes sent
	BandwidthInMbps  float64 `json:"bandwidth_in_mbps"`  // Inbound bandwidth (Mbps)
	BandwidthOutMbps float64 `json:"bandwidth_out_mbps"` // Outbound bandwidth (Mbps)

	// Connection statistics
	ActiveConnections int            `json:"active_connections"` // Active connections
	ConnectionsByIP   map[string]int `json:"connections_by_ip"`  // Connections per IP

	// Error statistics
	TotalErrors  int64            `json:"total_errors"`   // Total errors
	ErrorsByType map[string]int64 `json:"errors_by_type"` // Errors by type
	ErrorRate    float64          `json:"error_rate"`     // Current error rate

	// Timestamps
	StartTime  time.Time `json:"start_time"`  // Monitoring start time
	LastUpdate time.Time `json:"last_update"` // Last update time

	// Mutex for thread safety
	mu sync.RWMutex `json:"-"` // Statistics mutex
}

// TrafficSnapshot represents a point-in-time snapshot of traffic statistics
type TrafficSnapshot struct {
	Timestamp         time.Time `json:"timestamp"`
	MessagesPerSecond float64   `json:"messages_per_second"`
	BandwidthInMbps   float64   `json:"bandwidth_in_mbps"`
	BandwidthOutMbps  float64   `json:"bandwidth_out_mbps"`
	ActiveConnections int       `json:"active_connections"`
	ErrorRate         float64   `json:"error_rate"`
}

// AlertThresholds defines thresholds for generating alerts
type AlertThresholds struct {
	MaxMessagesPerSecond float64 `json:"max_messages_per_second"` // Max messages per second
	MaxBandwidthMbps     float64 `json:"max_bandwidth_mbps"`      // Max bandwidth in Mbps
	MaxErrorRate         float64 `json:"max_error_rate"`          // Max error rate (%)
	MaxConnections       int     `json:"max_connections"`         // Max connections
}

// NewBasicMonitor creates a new basic traffic monitor
func NewBasicMonitor(config *types.BasicMonitorConfig, logger log.Logger) *BasicMonitor {
	if config == nil {
		config = types.DefaultBasicMonitorConfig()
	}

	stats := &TrafficStats{
		MessagesByType:  make(map[string]int64),
		ConnectionsByIP: make(map[string]int),
		ErrorsByType:    make(map[string]int64),
		StartTime:       time.Now(),
		LastUpdate:      time.Now(),
	}

	thresholds := &AlertThresholds{
		MaxMessagesPerSecond: float64(config.MaxMessagesPerSecond),
		MaxBandwidthMbps:     float64(config.MaxBandwidthMbps),
		MaxErrorRate:         config.MaxErrorRate,
		MaxConnections:       config.MaxConnections,
	}

	return &BasicMonitor{
		config:          config,
		logger:          logger,
		stats:           stats,
		history:         make([]*TrafficSnapshot, 0),
		alertThresholds: thresholds,
		isRunning:       false,
	}
}

// Start starts the traffic monitor
func (bm *BasicMonitor) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.isRunning {
		return fmt.Errorf("basic monitor is already running")
	}

	bm.logger.Info("Starting basic traffic monitor",
		"snapshot_interval", bm.config.SnapshotInterval,
		"history_size", bm.config.HistorySize)

	// Reset statistics
	bm.stats.StartTime = time.Now()
	bm.stats.LastUpdate = time.Now()

	// Start monitoring routines
	bm.monitorTicker = time.NewTicker(bm.config.UpdateInterval)
	bm.snapshotTicker = time.NewTicker(bm.config.SnapshotInterval)

	go bm.monitoringLoop()
	go bm.snapshotLoop()

	bm.isRunning = true
	bm.logger.Info("Basic traffic monitor started successfully")
	return nil
}

// Stop stops the traffic monitor
func (bm *BasicMonitor) Stop() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.isRunning {
		return
	}

	bm.logger.Info("Stopping basic traffic monitor")

	// Stop tickers
	if bm.monitorTicker != nil {
		bm.monitorTicker.Stop()
	}
	if bm.snapshotTicker != nil {
		bm.snapshotTicker.Stop()
	}

	bm.isRunning = false
	bm.logger.Info("Basic traffic monitor stopped")
}

// RecordMessage records a processed message
func (bm *BasicMonitor) RecordMessage(messageType string, bytesIn, bytesOut int64) {
	if !bm.isRunning {
		return
	}

	bm.stats.mu.Lock()
	defer bm.stats.mu.Unlock()

	bm.stats.TotalMessages++
	bm.stats.MessagesByType[messageType]++
	bm.stats.TotalBytesIn += bytesIn
	bm.stats.TotalBytesOut += bytesOut
	bm.stats.LastUpdate = time.Now()
}

// RecordConnection records connection information
func (bm *BasicMonitor) RecordConnection(ip string, connected bool) {
	if !bm.isRunning {
		return
	}

	bm.stats.mu.Lock()
	defer bm.stats.mu.Unlock()

	if connected {
		bm.stats.ConnectionsByIP[ip]++
		bm.stats.ActiveConnections++
	} else {
		bm.stats.ConnectionsByIP[ip]--
		if bm.stats.ConnectionsByIP[ip] <= 0 {
			delete(bm.stats.ConnectionsByIP, ip)
		}
		bm.stats.ActiveConnections--
		if bm.stats.ActiveConnections < 0 {
			bm.stats.ActiveConnections = 0
		}
	}

	bm.stats.LastUpdate = time.Now()
}

// RecordError records an error occurrence
func (bm *BasicMonitor) RecordError(errorType string) {
	if !bm.isRunning {
		return
	}

	bm.stats.mu.Lock()
	defer bm.stats.mu.Unlock()

	bm.stats.TotalErrors++
	bm.stats.ErrorsByType[errorType]++
	bm.stats.LastUpdate = time.Now()
}

// GetCurrentStats returns current traffic statistics
func (bm *BasicMonitor) GetCurrentStats() *TrafficStats {
	bm.stats.mu.RLock()
	defer bm.stats.mu.RUnlock()

	// Create a copy to prevent external modification
	statsCopy := &TrafficStats{
		TotalMessages:     bm.stats.TotalMessages,
		MessagesByType:    make(map[string]int64),
		MessagesPerSecond: bm.stats.MessagesPerSecond,
		TotalBytesIn:      bm.stats.TotalBytesIn,
		TotalBytesOut:     bm.stats.TotalBytesOut,
		BandwidthInMbps:   bm.stats.BandwidthInMbps,
		BandwidthOutMbps:  bm.stats.BandwidthOutMbps,
		ActiveConnections: bm.stats.ActiveConnections,
		ConnectionsByIP:   make(map[string]int),
		TotalErrors:       bm.stats.TotalErrors,
		ErrorsByType:      make(map[string]int64),
		ErrorRate:         bm.stats.ErrorRate,
		StartTime:         bm.stats.StartTime,
		LastUpdate:        bm.stats.LastUpdate,
	}

	// Copy maps
	for k, v := range bm.stats.MessagesByType {
		statsCopy.MessagesByType[k] = v
	}
	for k, v := range bm.stats.ConnectionsByIP {
		statsCopy.ConnectionsByIP[k] = v
	}
	for k, v := range bm.stats.ErrorsByType {
		statsCopy.ErrorsByType[k] = v
	}

	return statsCopy
}

// GetHistory returns historical snapshots
func (bm *BasicMonitor) GetHistory() []*TrafficSnapshot {
	bm.historyMu.RLock()
	defer bm.historyMu.RUnlock()

	// Create a copy to prevent external modification
	history := make([]*TrafficSnapshot, len(bm.history))
	for i, snapshot := range bm.history {
		snapshotCopy := *snapshot
		history[i] = &snapshotCopy
	}

	return history
}

// monitoringLoop runs the main monitoring loop
func (bm *BasicMonitor) monitoringLoop() {
	defer func() {
		if r := recover(); r != nil {
			bm.logger.Error("Monitoring loop panic recovered", "error", r)
		}
	}()

	var lastStats *TrafficStats

	for {
		select {
		case <-bm.monitorTicker.C:
			if !bm.isRunning {
				return
			}
			bm.updateRates(lastStats)
			bm.checkAlerts()
			lastStats = bm.GetCurrentStats()
		}
	}
}

// updateRates updates calculated rates (messages/sec, bandwidth, error rate)
func (bm *BasicMonitor) updateRates(lastStats *TrafficStats) {
	bm.stats.mu.Lock()
	defer bm.stats.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bm.stats.LastUpdate).Seconds()

	if lastStats != nil && elapsed > 0 {
		// Calculate messages per second
		messageDiff := bm.stats.TotalMessages - lastStats.TotalMessages
		bm.stats.MessagesPerSecond = float64(messageDiff) / elapsed

		// Calculate bandwidth (convert bytes to Mbps)
		bytesInDiff := bm.stats.TotalBytesIn - lastStats.TotalBytesIn
		bytesOutDiff := bm.stats.TotalBytesOut - lastStats.TotalBytesOut
		bm.stats.BandwidthInMbps = float64(bytesInDiff) * 8 / (elapsed * 1024 * 1024)
		bm.stats.BandwidthOutMbps = float64(bytesOutDiff) * 8 / (elapsed * 1024 * 1024)

		// Calculate error rate
		if bm.stats.TotalMessages > 0 {
			bm.stats.ErrorRate = float64(bm.stats.TotalErrors) / float64(bm.stats.TotalMessages) * 100
		}
	}
}

// snapshotLoop runs the snapshot creation loop
func (bm *BasicMonitor) snapshotLoop() {
	defer func() {
		if r := recover(); r != nil {
			bm.logger.Error("Snapshot loop panic recovered", "error", r)
		}
	}()

	for {
		select {
		case <-bm.snapshotTicker.C:
			if !bm.isRunning {
				return
			}
			bm.createSnapshot()
		}
	}
}

// createSnapshot creates a new traffic snapshot
func (bm *BasicMonitor) createSnapshot() {
	bm.stats.mu.RLock()
	snapshot := &TrafficSnapshot{
		Timestamp:         time.Now(),
		MessagesPerSecond: bm.stats.MessagesPerSecond,
		BandwidthInMbps:   bm.stats.BandwidthInMbps,
		BandwidthOutMbps:  bm.stats.BandwidthOutMbps,
		ActiveConnections: bm.stats.ActiveConnections,
		ErrorRate:         bm.stats.ErrorRate,
	}
	bm.stats.mu.RUnlock()

	bm.historyMu.Lock()
	defer bm.historyMu.Unlock()

	// Add snapshot to history
	bm.history = append(bm.history, snapshot)

	// Limit history size
	if len(bm.history) > bm.config.HistorySize {
		bm.history = bm.history[len(bm.history)-bm.config.HistorySize:]
	}
}

// checkAlerts checks if any alert thresholds are exceeded
func (bm *BasicMonitor) checkAlerts() {
	// Avoid too frequent alerts
	if time.Since(bm.lastAlert) < bm.config.AlertCooldown {
		return
	}

	bm.stats.mu.RLock()
	defer bm.stats.mu.RUnlock()

	alerts := []string{}

	// Check message rate
	if bm.stats.MessagesPerSecond > bm.alertThresholds.MaxMessagesPerSecond {
		alerts = append(alerts, fmt.Sprintf("High message rate: %.2f/s (threshold: %.2f/s)",
			bm.stats.MessagesPerSecond, bm.alertThresholds.MaxMessagesPerSecond))
	}

	// Check bandwidth
	totalBandwidth := bm.stats.BandwidthInMbps + bm.stats.BandwidthOutMbps
	if totalBandwidth > bm.alertThresholds.MaxBandwidthMbps {
		alerts = append(alerts, fmt.Sprintf("High bandwidth usage: %.2f Mbps (threshold: %.2f Mbps)",
			totalBandwidth, bm.alertThresholds.MaxBandwidthMbps))
	}

	// Check error rate
	if bm.stats.ErrorRate > bm.alertThresholds.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%% (threshold: %.2f%%)",
			bm.stats.ErrorRate, bm.alertThresholds.MaxErrorRate))
	}

	// Check connections
	if bm.stats.ActiveConnections > bm.alertThresholds.MaxConnections {
		alerts = append(alerts, fmt.Sprintf("High connection count: %d (threshold: %d)",
			bm.stats.ActiveConnections, bm.alertThresholds.MaxConnections))
	}

	// Log alerts
	if len(alerts) > 0 {
		for _, alert := range alerts {
			bm.logger.Warn("Traffic alert", "alert", alert)
		}
		bm.lastAlert = time.Now()
	}
}

// Reset resets all statistics
func (bm *BasicMonitor) Reset() {
	bm.stats.mu.Lock()
	bm.historyMu.Lock()
	defer bm.stats.mu.Unlock()
	defer bm.historyMu.Unlock()

	bm.stats.TotalMessages = 0
	bm.stats.MessagesByType = make(map[string]int64)
	bm.stats.MessagesPerSecond = 0
	bm.stats.TotalBytesIn = 0
	bm.stats.TotalBytesOut = 0
	bm.stats.BandwidthInMbps = 0
	bm.stats.BandwidthOutMbps = 0
	bm.stats.ActiveConnections = 0
	bm.stats.ConnectionsByIP = make(map[string]int)
	bm.stats.TotalErrors = 0
	bm.stats.ErrorsByType = make(map[string]int64)
	bm.stats.ErrorRate = 0
	bm.stats.StartTime = time.Now()
	bm.stats.LastUpdate = time.Now()

	bm.history = make([]*TrafficSnapshot, 0)

	bm.logger.Info("Basic monitor statistics reset")
}

// IsRunning returns whether the monitor is running
func (bm *BasicMonitor) IsRunning() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.isRunning
}

// GetAlertThresholds returns current alert thresholds
func (bm *BasicMonitor) GetAlertThresholds() *AlertThresholds {
	// Return a copy to prevent external modification
	return &AlertThresholds{
		MaxMessagesPerSecond: bm.alertThresholds.MaxMessagesPerSecond,
		MaxBandwidthMbps:     bm.alertThresholds.MaxBandwidthMbps,
		MaxErrorRate:         bm.alertThresholds.MaxErrorRate,
		MaxConnections:       bm.alertThresholds.MaxConnections,
	}
}

// UpdateAlertThresholds updates alert thresholds
func (bm *BasicMonitor) UpdateAlertThresholds(thresholds *AlertThresholds) {
	bm.alertThresholds = &AlertThresholds{
		MaxMessagesPerSecond: thresholds.MaxMessagesPerSecond,
		MaxBandwidthMbps:     thresholds.MaxBandwidthMbps,
		MaxErrorRate:         thresholds.MaxErrorRate,
		MaxConnections:       thresholds.MaxConnections,
	}

	bm.logger.Info("Alert thresholds updated",
		"max_messages_per_second", thresholds.MaxMessagesPerSecond,
		"max_bandwidth_mbps", thresholds.MaxBandwidthMbps,
		"max_error_rate", thresholds.MaxErrorRate,
		"max_connections", thresholds.MaxConnections)
}
