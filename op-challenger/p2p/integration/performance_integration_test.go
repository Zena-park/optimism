package integration

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/defense"
	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
)

// TestMessageThroughput tests message processing throughput
func TestMessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping throughput test in short mode")
	}

	// Create logger for testing
	logger := log.New()

	monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
	require.NoError(t, monitor.Start())
	defer monitor.Stop()

	t.Run("MessageProcessingThroughput", func(t *testing.T) {
		messageCount := 10000
		bytesIn := int64(100) // bytes
		bytesOut := int64(0)

		startTime := time.Now()

		// Send messages as fast as possible
		for i := 0; i < messageCount; i++ {
			messageType := fmt.Sprintf("throughput-test-message-%d", i%10)
			monitor.RecordMessage(messageType, bytesIn, bytesOut)
		}

		duration := time.Since(startTime)
		throughput := float64(messageCount) / duration.Seconds()

		// Verify throughput meets requirements (> 1,000 msg/sec)
		assert.Greater(t, throughput, 1000.0,
			"Message throughput should exceed 1,000 msg/sec, got %.2f", throughput)

		// Verify all messages were recorded
		time.Sleep(100 * time.Millisecond) // Allow processing
		stats := monitor.GetCurrentStats()
		assert.GreaterOrEqual(t, stats.TotalMessages, uint64(messageCount))

		t.Logf("Message throughput: %.2f messages/second", throughput)
	})

	t.Run("ChallengerRegistrationThroughput", func(t *testing.T) {
		registrationCount := 1000

		startTime := time.Now()

		// Simulate challenger registrations
		for i := 0; i < registrationCount; i++ {
			challengerID := GenerateValidHexID(fmt.Sprintf("reg-test-%d", i))
			challengerInfo := &types.ChallengerInfo{
				ID:          challengerID,
				NodeID:      GenerateValidHexID("node"),
				Address:     fmt.Sprintf("127.0.0.1:808%d", i%10),
				Role:        types.ChallengerRoleChallenger,
				Status:      types.ChallengerStatusOnline,
				IsConnected: true,
			}

			// Verify challenger creation is fast
			assert.NotNil(t, challengerInfo)
			assert.Equal(t, challengerID, challengerInfo.ID)
		}

		duration := time.Since(startTime)
		throughput := float64(registrationCount) / duration.Seconds()

		// Verify throughput meets requirements (> 100 registrations/sec)
		assert.Greater(t, throughput, 100.0,
			"Registration throughput should exceed 100/sec, got %.2f", throughput)

		t.Logf("Challenger registration throughput: %.2f registrations/second", throughput)
	})

	t.Run("StateSyncThroughput", func(t *testing.T) {
		stateUpdateCount := 5000

		startTime := time.Now()

		// Simulate state updates
		for i := 0; i < stateUpdateCount; i++ {
			challengerID := GenerateValidHexID(fmt.Sprintf("state-test-%d", i%100))
			challengerState := types.NewChallengerState(
				challengerID,
				GenerateValidHexID("node"),
				fmt.Sprintf("127.0.0.1:808%d", i%10),
			)

			// Perform state operations
			challengerState.SetOnline()
			challengerState.UpdateConnectionInfo(5, 50*time.Millisecond)

			// Verify state operations are fast
			assert.True(t, challengerState.IsOnline)
			assert.True(t, challengerState.IsHealthy())
		}

		duration := time.Since(startTime)
		throughput := float64(stateUpdateCount) / duration.Seconds()

		// State sync should be very fast for in-memory operations
		assert.Greater(t, throughput, 1000.0,
			"State sync throughput should exceed 1,000/sec, got %.2f", throughput)

		t.Logf("State sync throughput: %.2f updates/second", throughput)
	})

	t.Run("DefenseComponentThroughput", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)
		require.NoError(t, rateLimiter.Start())
		defer rateLimiter.Stop()

		operationCount := 10000
		clientID := "throughput-test-client"

		startTime := time.Now()

		// Test rate limiter throughput
		for i := 0; i < operationCount; i++ {
			rateLimiter.CheckLimit(clientID)
		}

		duration := time.Since(startTime)
		throughput := float64(operationCount) / duration.Seconds()

		// Rate limiter should be very fast
		assert.Greater(t, throughput, 10000.0,
			"Rate limiter throughput should exceed 10,000 ops/sec, got %.2f", throughput)

		t.Logf("Rate limiter throughput: %.2f operations/second", throughput)
	})
}

// TestMessageLatency tests message processing latency
func TestMessageLatency(t *testing.T) {
	// Create logger for testing
	logger := log.New()

	monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
	require.NoError(t, monitor.Start())
	defer monitor.Stop()

	t.Run("MessageProcessingLatency", func(t *testing.T) {
		iterations := 1000
		latencies := make([]time.Duration, iterations)

		// Measure individual message processing latencies
		for i := 0; i < iterations; i++ {
			messageType := fmt.Sprintf("latency-test-message-%d", i%10)
			startTime := time.Now()
			monitor.RecordMessage(messageType, 50, 0)
			latencies[i] = time.Since(startTime)
		}

		// Calculate statistics
		var totalLatency time.Duration
		maxLatency := time.Duration(0)
		for _, latency := range latencies {
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
		}
		avgLatency := totalLatency / time.Duration(iterations)

		// Verify latency requirements (< 1ms average)
		assert.Less(t, avgLatency, 1*time.Millisecond,
			"Average message latency should be < 1ms, got %v", avgLatency)
		assert.Less(t, maxLatency, 10*time.Millisecond,
			"Max message latency should be < 10ms, got %v", maxLatency)

		t.Logf("Average message latency: %v, Max: %v", avgLatency, maxLatency)
	})

	t.Run("StateUpdateLatency", func(t *testing.T) {
		iterations := 1000
		latencies := make([]time.Duration, iterations)

		// Measure state update latencies
		for i := 0; i < iterations; i++ {
			challengerID := GenerateValidHexID(fmt.Sprintf("latency-test-%d", i))
			challengerState := types.NewChallengerState(
				challengerID,
				GenerateValidHexID("node"),
				"127.0.0.1:8080",
			)

			startTime := time.Now()
			challengerState.SetOnline()
			challengerState.UpdateConnectionInfo(5, 50*time.Millisecond)
			latencies[i] = time.Since(startTime)
		}

		// Calculate statistics
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		avgLatency := totalLatency / time.Duration(iterations)

		// State updates should be very fast (< 100μs)
		assert.Less(t, avgLatency, 100*time.Microsecond,
			"Average state update latency should be < 100μs, got %v", avgLatency)

		t.Logf("Average state update latency: %v", avgLatency)
	})

	t.Run("DefenseComponentLatency", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)
		require.NoError(t, rateLimiter.Start())
		defer rateLimiter.Stop()

		iterations := 1000
		latencies := make([]time.Duration, iterations)
		clientID := "latency-test-client"

		// Measure rate limiter latencies
		for i := 0; i < iterations; i++ {
			startTime := time.Now()
			rateLimiter.CheckLimit(clientID)
			latencies[i] = time.Since(startTime)
		}

		// Calculate average latency
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		avgLatency := totalLatency / time.Duration(iterations)

		// Rate limiter should be very fast (< 50μs)
		assert.Less(t, avgLatency, 50*time.Microsecond,
			"Average rate limiter latency should be < 50μs, got %v", avgLatency)

		t.Logf("Average rate limiter latency: %v", avgLatency)
	})
}

// TestMemoryUsageUnderLoad tests memory usage under various load conditions
func TestMemoryUsageUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	t.Run("MemoryUsageWithManyChallengers", func(t *testing.T) {
		// Force garbage collection to get baseline
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Create many challengers
		challengerCount := 1000
		challengers := make([]*types.ChallengerInfo, challengerCount)
		states := make([]*types.ChallengerState, challengerCount)

		for i := 0; i < challengerCount; i++ {
			challengerID := GenerateValidHexID(fmt.Sprintf("mem-test-%d", i))
			challengers[i] = &types.ChallengerInfo{
				ID:          challengerID,
				NodeID:      GenerateValidHexID("node"),
				Address:     fmt.Sprintf("127.0.0.1:808%d", i%100),
				Role:        types.ChallengerRoleChallenger,
				Status:      types.ChallengerStatusOnline,
				IsConnected: true,
			}
			states[i] = types.NewChallengerState(challengerID, challengers[i].NodeID, challengers[i].Address)
			states[i].SetOnline()
		}

		// Force garbage collection and measure memory
		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memoryUsed := memAfter.Alloc - memBefore.Alloc
		memoryUsedMB := float64(memoryUsed) / (1024 * 1024)

		// Verify memory usage is reasonable (< 50MB for 1000 challengers)
		assert.Less(t, memoryUsedMB, 50.0,
			"Memory usage should be < 50MB for 1000 challengers, got %.2f MB", memoryUsedMB)

		t.Logf("Memory usage for %d challengers: %.2f MB", challengerCount, memoryUsedMB)

		// Clean up to prevent memory leaks in test
		challengers = nil
		states = nil
		runtime.GC()
	})

	t.Run("MemoryUsageWithHighMessageVolume", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
		require.NoError(t, monitor.Start())
		defer monitor.Stop()

		// Force garbage collection to get baseline
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Generate high message volume
		messageCount := 50000

		for i := 0; i < messageCount; i++ {
			messageType := fmt.Sprintf("memory-test-message-%d", i%100)
			monitor.RecordMessage(messageType, 100, 0)
			if i%1000 == 0 {
				monitor.RecordConnection(fmt.Sprintf("192.168.1.%d", i%255+1), true)
			}
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)

		// Force garbage collection and measure memory
		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memoryUsed := memAfter.Alloc - memBefore.Alloc
		memoryUsedMB := float64(memoryUsed) / (1024 * 1024)

		// Verify memory usage is reasonable (< 20MB for message processing)
		assert.Less(t, memoryUsedMB, 20.0,
			"Memory usage should be < 20MB for message processing, got %.2f MB", memoryUsedMB)

		t.Logf("Memory usage for %d messages: %.2f MB", messageCount, memoryUsedMB)
	})

	t.Run("MemoryUsageWithDefenseComponents", func(t *testing.T) {
		// Force garbage collection to get baseline
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Create multiple defense components
		numComponents := 10
		rateLimiters := make([]*defense.RateLimiter, numComponents)
		connLimiters := make([]*defense.ConnectionLimiter, numComponents)
		monitors := make([]*defense.BasicMonitor, numComponents)

		// Create logger for testing
		logger := log.New()

		for i := 0; i < numComponents; i++ {
			rateLimiters[i] = defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)
			connLimiters[i] = defense.NewConnectionLimiter(types.DefaultConnectionLimiterConfig(), logger)
			monitors[i] = defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)

			require.NoError(t, rateLimiters[i].Start())
			require.NoError(t, connLimiters[i].Start())
			require.NoError(t, monitors[i].Start())
		}

		// Use the components
		for i := 0; i < 1000; i++ {
			componentIndex := i % numComponents
			clientID := fmt.Sprintf("client-%d", i)

			rateLimiters[componentIndex].CheckLimit(clientID)
			monitors[componentIndex].RecordMessage("test-message", 10, 0)
		}

		// Clean up
		for i := 0; i < numComponents; i++ {
			rateLimiters[i].Stop()
			connLimiters[i].Stop()
			monitors[i].Stop()
		}

		// Force garbage collection and measure memory
		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memoryUsed := memAfter.Alloc - memBefore.Alloc
		memoryUsedMB := float64(memoryUsed) / (1024 * 1024)

		// Verify memory usage is reasonable (< 10MB for defense components)
		assert.Less(t, memoryUsedMB, 10.0,
			"Memory usage should be < 10MB for defense components, got %.2f MB", memoryUsedMB)

		t.Logf("Memory usage for %d defense components: %.2f MB", numComponents, memoryUsedMB)
	})
}

// TestCPUUsageUnderLoad tests CPU usage under various load conditions
func TestCPUUsageUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CPU test in short mode")
	}

	t.Run("CPUUsageNormalLoad", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
		rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)

		require.NoError(t, monitor.Start())
		require.NoError(t, rateLimiter.Start())

		defer func() {
			monitor.Stop()
			rateLimiter.Stop()
		}()

		// Simulate normal load
		startTime := time.Now()
		duration := 5 * time.Second
		operations := 0

		for time.Since(startTime) < duration {
			// Normal challenger operations
			monitor.RecordMessage("normal-message", 50, 0)
			rateLimiter.CheckLimit("normal-client")

			// Small delay to simulate normal operation rate
			time.Sleep(100 * time.Microsecond)
			operations++
		}

		operationsPerSecond := float64(operations) / duration.Seconds()

		// Verify system can handle normal load efficiently
		assert.Greater(t, operationsPerSecond, 1000.0,
			"Should handle > 1000 operations/sec under normal load, got %.2f", operationsPerSecond)

		t.Logf("Normal load: %.2f operations/second", operationsPerSecond)
	})

	t.Run("CPUUsageUnderDefense", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		// Create defense components with strict limits
		rateLimiterConfig := types.DefaultRateLimiterConfig()
		rateLimiterConfig.MaxRequestsPerSecond = 100
		rateLimiter := defense.NewRateLimiter(rateLimiterConfig, logger)

		connLimiterConfig := types.DefaultConnectionLimiterConfig()
		connLimiterConfig.MaxConnections = 10
		connLimiter := defense.NewConnectionLimiter(connLimiterConfig, logger)

		monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)

		require.NoError(t, rateLimiter.Start())
		require.NoError(t, connLimiter.Start())
		require.NoError(t, monitor.Start())

		defer func() {
			rateLimiter.Stop()
			connLimiter.Stop()
			monitor.Stop()
		}()

		// Simulate defense scenario with high attack load
		startTime := time.Now()
		duration := 3 * time.Second
		operations := 0

		for time.Since(startTime) < duration {
			// High frequency attack simulation
			rateLimiter.CheckLimit("attacker")
			connLimiter.CheckConnection(fmt.Sprintf("10.0.0.%d:8080", operations%255+1))
			monitor.RecordMessage("attack-message", 10, 0)
			monitor.RecordError("malicious-behavior")

			operations++
		}

		operationsPerSecond := float64(operations) / duration.Seconds()

		// System should still be responsive under attack
		assert.Greater(t, operationsPerSecond, 500.0,
			"Should handle > 500 operations/sec under defense, got %.2f", operationsPerSecond)

		t.Logf("Defense load: %.2f operations/second", operationsPerSecond)
	})
}

// TestResourceLeakDetection tests for resource leaks
func TestResourceLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource leak test in short mode")
	}

	t.Run("MemoryLeakDetection", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
		require.NoError(t, monitor.Start())
		defer monitor.Stop()

		// Measure memory before operations
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Perform many operations in cycles
		cycles := 10
		operationsPerCycle := 1000

		for cycle := 0; cycle < cycles; cycle++ {
			for i := 0; i < operationsPerCycle; i++ {
				messageType := fmt.Sprintf("leak-test-message-%d", i%10)
				monitor.RecordMessage(messageType, 50, 0)
				monitor.RecordConnection(fmt.Sprintf("192.168.1.%d", i%255+1), true)
			}

			// Force garbage collection between cycles
			runtime.GC()
			time.Sleep(10 * time.Millisecond)
		}

		// Final memory measurement
		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memoryGrowth := float64(memAfter.Alloc-memBefore.Alloc) / (1024 * 1024)

		// Memory growth should be minimal (< 10MB) after many cycles
		assert.Less(t, memoryGrowth, 10.0,
			"Memory growth should be < 10MB after %d cycles, got %.2f MB", cycles, memoryGrowth)

		t.Logf("Memory growth after %d cycles: %.2f MB", cycles, memoryGrowth)
	})

	t.Run("GoroutineLeakDetection", func(t *testing.T) {
		// Create logger for testing
		logger := log.New()

		// Measure goroutines before
		goroutinesBefore := runtime.NumGoroutine()

		// Start and stop components multiple times
		for i := 0; i < 5; i++ {
			monitor := defense.NewBasicMonitor(types.DefaultBasicMonitorConfig(), logger)
			rateLimiter := defense.NewRateLimiter(types.DefaultRateLimiterConfig(), logger)

			require.NoError(t, monitor.Start())
			require.NoError(t, rateLimiter.Start())

			// Do some work
			time.Sleep(10 * time.Millisecond)

			// Stop components
			monitor.Stop()
			rateLimiter.Stop()
		}

		// Allow time for cleanup
		time.Sleep(100 * time.Millisecond)
		runtime.GC()

		// Measure goroutines after
		goroutinesAfter := runtime.NumGoroutine()
		goroutineGrowth := goroutinesAfter - goroutinesBefore

		// Goroutine growth should be minimal (< 5)
		assert.LessOrEqual(t, goroutineGrowth, 5,
			"Goroutine growth should be <= 5, got %d (before: %d, after: %d)",
			goroutineGrowth, goroutinesBefore, goroutinesAfter)

		t.Logf("Goroutine growth: %d (before: %d, after: %d)",
			goroutineGrowth, goroutinesBefore, goroutinesAfter)
	})

	t.Run("ChallengerStateLeakDetection", func(t *testing.T) {
		// Test for leaks in challenger state management
		runtime.GC()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		// Create and destroy many challenger states
		for cycle := 0; cycle < 10; cycle++ {
			states := make([]*types.ChallengerState, 100)

			for i := 0; i < 100; i++ {
				challengerID := GenerateValidHexID(fmt.Sprintf("leak-test-%d-%d", cycle, i))
				states[i] = types.NewChallengerState(challengerID, GenerateValidHexID("node"), "127.0.0.1:8080")
				states[i].SetOnline()
				states[i].UpdateConnectionInfo(5, 50*time.Millisecond)
			}

			// Clear references
			states = nil
			runtime.GC()
		}

		// Final memory measurement
		runtime.GC()
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		memoryGrowth := float64(memAfter.Alloc-memBefore.Alloc) / (1024 * 1024)

		// Memory growth should be minimal
		assert.Less(t, memoryGrowth, 5.0,
			"Challenger state memory growth should be < 5MB, got %.2f MB", memoryGrowth)

		t.Logf("Challenger state memory growth: %.2f MB", memoryGrowth)
	})
}
