package defense

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestNewBasicMonitor tests basic monitor creation
func TestNewBasicMonitor(t *testing.T) {
	config := &types.BasicMonitorConfig{
		Enabled:              true,
		UpdateInterval:       time.Second,
		SnapshotInterval:     time.Minute,
		MaxMessagesPerSecond: 100,
		MaxBandwidthMbps:     10,
		MaxErrorRate:         0.1,
		MaxConnections:       1000,
		HistorySize:          100,
		AlertCooldown:        time.Minute,
	}

	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	require.NotNil(t, monitor, "Basic monitor should not be nil")
	assert.Equal(t, config, monitor.config, "Config should match")
	assert.Equal(t, logger, monitor.logger, "Logger should match")
}

// TestBasicMonitorStart tests starting the basic monitor
func TestBasicMonitorStart(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	// Test starting
	err := monitor.Start()
	assert.NoError(t, err, "Starting basic monitor should succeed")
	assert.True(t, monitor.IsRunning(), "Basic monitor should be running")

	// Test double start
	err = monitor.Start()
	assert.Error(t, err, "Double start should fail")

	// Cleanup
	monitor.Stop()
}

// TestBasicMonitorStop tests stopping the basic monitor
func TestBasicMonitorStop(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	// Test stopping when not running
	monitor.Stop()
	assert.False(t, monitor.IsRunning(), "Basic monitor should not be running")

	// Start and then stop
	err := monitor.Start()
	require.NoError(t, err)

	monitor.Stop()
	assert.False(t, monitor.IsRunning(), "Basic monitor should not be running")

	// Test double stop
	monitor.Stop()
	assert.False(t, monitor.IsRunning(), "Basic monitor should still not be running")
}

// TestBasicMonitorRecordMessage tests message recording
func TestBasicMonitorRecordMessage(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Record some messages
	monitor.RecordMessage("test", 100, 50)
	monitor.RecordMessage("test", 200, 100)
	monitor.RecordMessage("ping", 150, 75)

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Get stats
	stats := monitor.GetCurrentStats()
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Equal(t, int64(3), stats.TotalMessages, "Should have recorded 3 messages")
	assert.Equal(t, int64(450), stats.TotalBytesIn, "Should have recorded 450 bytes in")
	assert.Equal(t, int64(225), stats.TotalBytesOut, "Should have recorded 225 bytes out")
}

// TestBasicMonitorRecordConnection tests connection recording
func TestBasicMonitorRecordConnection(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Record connections
	monitor.RecordConnection("127.0.0.1", true)  // Connect
	monitor.RecordConnection("127.0.0.2", true)  // Connect
	monitor.RecordConnection("127.0.0.3", true)  // Connect
	monitor.RecordConnection("127.0.0.1", false) // Disconnect

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Get stats
	stats := monitor.GetCurrentStats()
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Equal(t, 2, stats.ActiveConnections, "Should have 2 active connections")
}

// TestBasicMonitorRecordError tests error recording
func TestBasicMonitorRecordError(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Record some errors
	monitor.RecordError("timeout")
	monitor.RecordError("invalid_message")
	monitor.RecordError("timeout")

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Get stats
	stats := monitor.GetCurrentStats()
	assert.NotNil(t, stats, "Stats should not be nil")
	assert.Equal(t, int64(3), stats.TotalErrors, "Should have recorded 3 errors")
	assert.Equal(t, int64(2), stats.ErrorsByType["timeout"], "Should have 2 timeout errors")
	assert.Equal(t, int64(1), stats.ErrorsByType["invalid_message"], "Should have 1 invalid_message error")
}

// TestBasicMonitorGetHistory tests history retrieval
func TestBasicMonitorGetHistory(t *testing.T) {
	config := &types.BasicMonitorConfig{
		Enabled:              true,
		UpdateInterval:       50 * time.Millisecond,  // Fast updates for testing
		SnapshotInterval:     100 * time.Millisecond, // Fast snapshots for testing
		MaxMessagesPerSecond: 100,
		MaxBandwidthMbps:     10,
		MaxErrorRate:         0.1,
		MaxConnections:       1000,
		HistorySize:          10,
		AlertCooldown:        time.Minute,
	}

	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Generate activity to create history
	monitor.RecordMessage("test", 100, 50)
	monitor.RecordMessage("test", 200, 100)

	// Wait for snapshots to be taken
	time.Sleep(250 * time.Millisecond)

	// Get history
	history := monitor.GetHistory()
	assert.NotNil(t, history, "History should not be nil")
	assert.Greater(t, len(history), 0, "History should contain entries")

	// Check that history entries have expected structure
	if len(history) > 0 {
		entry := history[0]
		assert.Greater(t, entry.Timestamp.Unix(), int64(0), "History entry should have valid timestamp")
		assert.GreaterOrEqual(t, entry.MessagesPerSecond, float64(0), "History entry should have valid message rate")
	}
}

// TestBasicMonitorReset tests reset functionality
func TestBasicMonitorReset(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Generate some activity
	monitor.RecordMessage("test", 100, 50)
	monitor.RecordError("timeout")
	monitor.RecordConnection("127.0.0.1", true)

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Get stats before reset
	statsBefore := monitor.GetCurrentStats()
	assert.Greater(t, statsBefore.TotalMessages, int64(0), "Should have processed messages")

	// Reset
	monitor.Reset()

	// Get stats after reset
	statsAfter := monitor.GetCurrentStats()
	assert.Equal(t, int64(0), statsAfter.TotalMessages, "Total messages should be reset to 0")
	assert.Equal(t, int64(0), statsAfter.TotalErrors, "Total errors should be reset to 0")
	assert.Equal(t, 0, statsAfter.ActiveConnections, "Active connections should be reset to 0")
}

// TestBasicMonitorConcurrency tests concurrent access
func TestBasicMonitorConcurrency(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				messageType := fmt.Sprintf("type-%d", id)
				ip := fmt.Sprintf("127.0.%d.%d", id%255, j%255)

				monitor.RecordMessage(messageType, 100, 50)
				monitor.RecordError("test_error")
				monitor.RecordConnection(ip, true)
				monitor.RecordConnection(ip, false)
				monitor.GetCurrentStats()
				monitor.GetHistory()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Get final stats
	stats := monitor.GetCurrentStats()
	assert.NotNil(t, stats, "Stats should be available after concurrent operations")
}

// TestBasicMonitorEdgeCases tests edge cases
func TestBasicMonitorEdgeCases(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Test empty message type
	monitor.RecordMessage("", 100, 50)
	// Should not panic

	// Test very long message type
	longMessageType := strings.Repeat("a", 1000)
	monitor.RecordMessage(longMessageType, 100, 50)
	// Should not panic

	// Test zero and negative byte counts
	monitor.RecordMessage("test", 0, 0)
	monitor.RecordMessage("test", -100, -50)
	// Should handle gracefully

	// Test very large byte counts
	monitor.RecordMessage("test", 1000000000, 1000000000)
	// Should handle gracefully

	// Test empty error type
	monitor.RecordError("")
	// Should handle gracefully

	// Test very long error type
	longErrorType := strings.Repeat("error", 1000)
	monitor.RecordError(longErrorType)
	// Should handle gracefully

	// Test empty IP
	monitor.RecordConnection("", true)
	monitor.RecordConnection("", false)
	// Should handle gracefully

	// Test invalid IP
	monitor.RecordConnection("invalid-ip", true)
	// Should handle gracefully

	// Get stats after edge cases
	stats := monitor.GetCurrentStats()
	assert.NotNil(t, stats, "Stats should be available after edge cases")
}

// TestBasicMonitorNilConfig tests behavior with nil config
func TestBasicMonitorNilConfig(t *testing.T) {
	logger := log.New()

	// This should not panic but use default config
	monitor := NewBasicMonitor(nil, logger)
	assert.NotNil(t, monitor, "Basic monitor should handle nil config")

	// Should be able to start with default config
	err := monitor.Start()
	assert.NoError(t, err, "Should start with default config")

	monitor.Stop()
}

// TestBasicMonitorAlertThresholds tests alert threshold management
func TestBasicMonitorAlertThresholds(t *testing.T) {
	config := types.DefaultBasicMonitorConfig()
	logger := log.New()
	monitor := NewBasicMonitor(config, logger)

	err := monitor.Start()
	require.NoError(t, err)
	defer monitor.Stop()

	// Get initial thresholds
	thresholds := monitor.GetAlertThresholds()
	assert.NotNil(t, thresholds, "Alert thresholds should not be nil")

	// Update thresholds
	newThresholds := &AlertThresholds{
		MaxMessagesPerSecond: 200,
		MaxBandwidthMbps:     20,
		MaxErrorRate:         0.2,
		MaxConnections:       2000,
	}

	monitor.UpdateAlertThresholds(newThresholds)

	// Get updated thresholds
	updatedThresholds := monitor.GetAlertThresholds()
	assert.Equal(t, float64(200), updatedThresholds.MaxMessagesPerSecond, "Message threshold should be updated")
	assert.Equal(t, float64(20), updatedThresholds.MaxBandwidthMbps, "Bandwidth threshold should be updated")
	assert.Equal(t, 0.2, updatedThresholds.MaxErrorRate, "Error rate threshold should be updated")
	assert.Equal(t, 2000, updatedThresholds.MaxConnections, "Connection threshold should be updated")
}
