package challenger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum/go-ethereum/log"
)

// ChallengerMonitor handles basic challenger monitoring (Phase 1 features only)
type ChallengerMonitor struct {
	manager *ChallengerNetworkManager // Reference to parent manager
	logger  log.Logger                // Logger

	// Monitor state
	isRunning     bool               // Whether monitor is running
	monitorTicker *time.Ticker       // Monitor ticker
	ctx           context.Context    // Context
	cancel        context.CancelFunc // Cancel function

	// Monitoring data (Phase 1: basic tracking only)
	challengerHealth map[string]*healthEntry // Challenger health tracking
	mu               sync.RWMutex            // Concurrency control

	// Phase 1: Basic monitoring stats
	lastMonitorRun  time.Time // Last monitor run time
	monitorRunCount int       // Number of monitor runs
	alertCount      int       // Number of alerts generated
}

// healthEntry represents basic health information for a challenger
type healthEntry struct {
	ChallengerID        string                 `json:"challenger_id"`
	LastSeen            time.Time              `json:"last_seen"`
	Status              types.ChallengerStatus `json:"status"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	LastHealthCheck     time.Time              `json:"last_health_check"`
	IsHealthy           bool                   `json:"is_healthy"`

	// Phase 1: Basic metrics only
	PeerCount      int           `json:"peer_count"`
	NetworkLatency time.Duration `json:"network_latency"`
	LastUpdate     time.Time     `json:"last_update"`
}

// NewChallengerMonitor creates a new challenger monitor service
func NewChallengerMonitor(manager *ChallengerNetworkManager, logger log.Logger) *ChallengerMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ChallengerMonitor{
		manager:          manager,
		logger:           logger,
		isRunning:        false,
		challengerHealth: make(map[string]*healthEntry),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start starts the challenger monitor service
func (cm *ChallengerMonitor) Start() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		return fmt.Errorf("challenger monitor is already running")
	}

	cm.logger.Info("Starting challenger monitor service")

	// Start monitoring loop (Phase 1: basic health checking)
	cm.monitorTicker = time.NewTicker(30 * time.Second) // Check every 30 seconds
	go cm.monitorLoop()

	cm.isRunning = true
	cm.lastMonitorRun = time.Now()

	cm.logger.Info("Challenger monitor service started")
	return nil
}

// Stop stops the challenger monitor service
func (cm *ChallengerMonitor) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isRunning {
		return
	}

	cm.logger.Info("Stopping challenger monitor service")

	// Stop monitor ticker
	if cm.monitorTicker != nil {
		cm.monitorTicker.Stop()
	}

	// Cancel context
	cm.cancel()

	cm.isRunning = false
	cm.logger.Info("Challenger monitor service stopped")
}

// UpdateChallengerHealth updates health information for a challenger
func (cm *ChallengerMonitor) UpdateChallengerHealth(challengerID string, challengerInfo *types.ChallengerInfo) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isRunning {
		return
	}

	// Get or create health entry
	health, exists := cm.challengerHealth[challengerID]
	if !exists {
		health = &healthEntry{
			ChallengerID:        challengerID,
			ConsecutiveFailures: 0,
			IsHealthy:           true,
		}
		cm.challengerHealth[challengerID] = health
	}

	// Update health information
	health.LastSeen = challengerInfo.LastSeen
	health.Status = challengerInfo.Status
	health.PeerCount = challengerInfo.PeerCount
	health.NetworkLatency = challengerInfo.NetworkLatency
	health.LastUpdate = time.Now()
	health.LastHealthCheck = time.Now()

	// Phase 1: Simple health determination
	health.IsHealthy = cm.isHealthyBasic(challengerInfo)

	if !health.IsHealthy {
		health.ConsecutiveFailures++
		if health.ConsecutiveFailures == 1 { // First failure
			cm.alertCount++
			cm.logger.Warn("Challenger health degraded",
				"challenger_id", challengerID,
				"status", challengerInfo.Status.String())
		}
	} else {
		if health.ConsecutiveFailures > 0 {
			cm.logger.Info("Challenger health recovered",
				"challenger_id", challengerID,
				"previous_failures", health.ConsecutiveFailures)
		}
		health.ConsecutiveFailures = 0
	}

	cm.logger.Debug("Challenger health updated",
		"challenger_id", challengerID,
		"is_healthy", health.IsHealthy,
		"consecutive_failures", health.ConsecutiveFailures)
}

// GetChallengerHealth returns health information for a challenger
func (cm *ChallengerMonitor) GetChallengerHealth(challengerID string) *healthEntry {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	health, exists := cm.challengerHealth[challengerID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	healthCopy := *health
	return &healthCopy
}

// GetAllChallengerHealth returns health information for all challengers
func (cm *ChallengerMonitor) GetAllChallengerHealth() map[string]*healthEntry {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	healthMap := make(map[string]*healthEntry)
	for id, health := range cm.challengerHealth {
		// Return a copy to prevent external modification
		healthCopy := *health
		healthMap[id] = &healthCopy
	}

	return healthMap
}

// GetHealthyChallengers returns a list of healthy challengers
func (cm *ChallengerMonitor) GetHealthyChallengers() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var healthy []string
	for id, health := range cm.challengerHealth {
		if health.IsHealthy {
			healthy = append(healthy, id)
		}
	}

	return healthy
}

// GetUnhealthyChallengers returns a list of unhealthy challengers
func (cm *ChallengerMonitor) GetUnhealthyChallengers() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var unhealthy []string
	for id, health := range cm.challengerHealth {
		if !health.IsHealthy {
			unhealthy = append(unhealthy, id)
		}
	}

	return unhealthy
}

// RemoveChallengerHealth removes health tracking for a challenger
func (cm *ChallengerMonitor) RemoveChallengerHealth(challengerID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.challengerHealth, challengerID)
	cm.logger.Debug("Challenger health tracking removed", "challenger_id", challengerID)
}

// monitorLoop runs the monitoring loop
func (cm *ChallengerMonitor) monitorLoop() {
	defer func() {
		if r := recover(); r != nil {
			cm.logger.Error("Monitor loop panic recovered", "error", r)
		}
	}()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.monitorTicker.C:
			cm.performMonitoring()
		}
	}
}

// performMonitoring performs the monitoring checks
func (cm *ChallengerMonitor) performMonitoring() {
	cm.mu.Lock()
	cm.lastMonitorRun = time.Now()
	cm.monitorRunCount++
	cm.mu.Unlock()

	// Get all known challengers from manager
	challengers := cm.manager.GetAllChallengers()

	// Update health for each challenger
	for _, challengerInfo := range challengers {
		cm.UpdateChallengerHealth(challengerInfo.ID, challengerInfo)
	}

	// Phase 1: Basic cleanup - remove health entries for unknown challengers
	cm.cleanupUnknownChallengers(challengers)

	cm.logger.Debug("Monitoring check completed",
		"monitored_challengers", len(challengers),
		"run_count", cm.monitorRunCount)
}

// cleanupUnknownChallengers removes health entries for challengers no longer known
func (cm *ChallengerMonitor) cleanupUnknownChallengers(knownChallengers map[string]*types.ChallengerInfo) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var toRemove []string
	for challengerID := range cm.challengerHealth {
		if _, exists := knownChallengers[challengerID]; !exists {
			toRemove = append(toRemove, challengerID)
		}
	}

	for _, challengerID := range toRemove {
		delete(cm.challengerHealth, challengerID)
		cm.logger.Debug("Removed health tracking for unknown challenger", "challenger_id", challengerID)
	}
}

// isHealthyBasic performs basic health check (Phase 1)
func (cm *ChallengerMonitor) isHealthyBasic(challengerInfo *types.ChallengerInfo) bool {
	// Phase 1: Simple health criteria

	// Check if online
	if !challengerInfo.IsOnline() {
		return false
	}

	// Check if recently seen (within last 5 minutes)
	if time.Since(challengerInfo.LastSeen) > 5*time.Minute {
		return false
	}

	// Phase 1: Basic checks only
	// In Phase 2-3, we could add more sophisticated health checks:
	// - Network latency thresholds
	// - Peer count requirements
	// - Performance metrics
	// - Response time analysis

	return true
}

// GetMonitoringStats returns monitoring statistics
func (cm *ChallengerMonitor) GetMonitoringStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	healthyCount := 0
	unhealthyCount := 0

	for _, health := range cm.challengerHealth {
		if health.IsHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	return map[string]interface{}{
		"is_running":        cm.isRunning,
		"monitored_count":   len(cm.challengerHealth),
		"healthy_count":     healthyCount,
		"unhealthy_count":   unhealthyCount,
		"monitor_run_count": cm.monitorRunCount,
		"alert_count":       cm.alertCount,
		"last_monitor_run":  cm.lastMonitorRun,
	}
}

// GetHealthSummary returns a summary of challenger health
func (cm *ChallengerMonitor) GetHealthSummary() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalCount := len(cm.challengerHealth)
	healthyCount := 0
	unhealthyCount := 0
	avgFailures := 0.0

	if totalCount > 0 {
		totalFailures := 0
		for _, health := range cm.challengerHealth {
			if health.IsHealthy {
				healthyCount++
			} else {
				unhealthyCount++
			}
			totalFailures += health.ConsecutiveFailures
		}
		avgFailures = float64(totalFailures) / float64(totalCount)
	}

	healthPercentage := 0.0
	if totalCount > 0 {
		healthPercentage = float64(healthyCount) / float64(totalCount) * 100
	}

	return map[string]interface{}{
		"total_challengers":        totalCount,
		"healthy_challengers":      healthyCount,
		"unhealthy_challengers":    unhealthyCount,
		"health_percentage":        healthPercentage,
		"avg_consecutive_failures": avgFailures,
	}
}

// IsRunning returns whether the monitor service is running
func (cm *ChallengerMonitor) IsRunning() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.isRunning
}
