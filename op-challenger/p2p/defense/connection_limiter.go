package defense

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-challenger/p2p/types"
	"github.com/ethereum/go-ethereum/log"
)

// ConnectionLimiter implements connection limiting for P2P network (Phase 1: basic IP-based limiting)
type ConnectionLimiter struct {
	// Configuration
	config *types.ConnectionLimiterConfig // Connection limiter configuration
	logger log.Logger                     // Logger

	// Connection tracking
	connections      map[string]*ConnectionInfo // Active connections by IP
	ipConnections    map[string]int             // Connection count per IP
	totalConnections int                        // Total active connections
	mu               sync.RWMutex               // Concurrency control

	// Statistics
	totalAccepted int64     // Total connections accepted
	totalRejected int64     // Total connections rejected
	lastCleanup   time.Time // Last cleanup time

	// State
	isRunning bool // Whether the connection limiter is running
}

// ConnectionInfo represents information about an active connection
type ConnectionInfo struct {
	IP            string    `json:"ip"`             // IP address
	ConnectedAt   time.Time `json:"connected_at"`   // Connection time
	LastActivity  time.Time `json:"last_activity"`  // Last activity time
	BytesSent     int64     `json:"bytes_sent"`     // Bytes sent
	BytesReceived int64     `json:"bytes_received"` // Bytes received
}

// NewConnectionLimiter creates a new connection limiter
func NewConnectionLimiter(config *types.ConnectionLimiterConfig, logger log.Logger) *ConnectionLimiter {
	if config == nil {
		config = types.DefaultConnectionLimiterConfig()
	}

	return &ConnectionLimiter{
		config:        config,
		logger:        logger,
		connections:   make(map[string]*ConnectionInfo),
		ipConnections: make(map[string]int),
		mu:            sync.RWMutex{},
		lastCleanup:   time.Now(),
		isRunning:     false,
	}
}

// Start starts the connection limiter
func (cl *ConnectionLimiter) Start() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.isRunning {
		return fmt.Errorf("connection limiter is already running")
	}

	cl.logger.Info("Starting connection limiter",
		"max_connections", cl.config.MaxConnections,
		"max_connections_per_ip", cl.config.MaxConnectionsPerIP)

	cl.isRunning = true
	cl.lastCleanup = time.Now()

	// Start cleanup routine
	go cl.cleanupRoutine()

	cl.logger.Info("Connection limiter started successfully")
	return nil
}

// Stop stops the connection limiter
func (cl *ConnectionLimiter) Stop() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if !cl.isRunning {
		return
	}

	cl.logger.Info("Stopping connection limiter")
	cl.isRunning = false
	cl.logger.Info("Connection limiter stopped")
}

// CheckConnection checks if a new connection should be allowed
func (cl *ConnectionLimiter) CheckConnection(remoteAddr string) bool {
	cl.mu.RLock()
	if !cl.isRunning {
		cl.mu.RUnlock()
		return true // Allow all connections if not running
	}
	cl.mu.RUnlock()

	// Extract IP from address
	ip := extractIP(remoteAddr)
	if ip == "" {
		cl.logger.Warn("Failed to extract IP from address", "address", remoteAddr)
		return false
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Check total connection limit
	if cl.totalConnections >= cl.config.MaxConnections {
		cl.totalRejected++
		cl.logger.Debug("Connection rejected: total limit reached",
			"ip", ip, "total_connections", cl.totalConnections)
		return false
	}

	// Check per-IP connection limit
	ipCount := cl.ipConnections[ip]
	if ipCount >= cl.config.MaxConnectionsPerIP {
		cl.totalRejected++
		cl.logger.Debug("Connection rejected: per-IP limit reached",
			"ip", ip, "ip_connections", ipCount)
		return false
	}

	cl.totalAccepted++
	cl.logger.Debug("Connection allowed", "ip", ip)
	return true
}

// AddConnection registers a new connection
func (cl *ConnectionLimiter) AddConnection(connectionID, remoteAddr string) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if !cl.isRunning {
		return fmt.Errorf("connection limiter is not running")
	}

	// Extract IP from address
	ip := extractIP(remoteAddr)
	if ip == "" {
		return fmt.Errorf("failed to extract IP from address: %s", remoteAddr)
	}

	// Check if connection already exists
	if _, exists := cl.connections[connectionID]; exists {
		return fmt.Errorf("connection already exists: %s", connectionID)
	}

	// Create connection info
	connInfo := &ConnectionInfo{
		IP:            ip,
		ConnectedAt:   time.Now(),
		LastActivity:  time.Now(),
		BytesSent:     0,
		BytesReceived: 0,
	}

	// Add connection
	cl.connections[connectionID] = connInfo
	cl.ipConnections[ip]++
	cl.totalConnections++

	cl.logger.Debug("Connection added",
		"connection_id", connectionID, "ip", ip,
		"total_connections", cl.totalConnections)
	return nil
}

// RemoveConnection unregisters a connection
func (cl *ConnectionLimiter) RemoveConnection(connectionID string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	connInfo, exists := cl.connections[connectionID]
	if !exists {
		return
	}

	// Remove connection
	delete(cl.connections, connectionID)
	cl.ipConnections[connInfo.IP]--
	if cl.ipConnections[connInfo.IP] <= 0 {
		delete(cl.ipConnections, connInfo.IP)
	}
	cl.totalConnections--

	cl.logger.Debug("Connection removed",
		"connection_id", connectionID, "ip", connInfo.IP,
		"total_connections", cl.totalConnections)
}

// UpdateConnectionActivity updates the last activity time for a connection
func (cl *ConnectionLimiter) UpdateConnectionActivity(connectionID string, bytesSent, bytesReceived int64) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	connInfo, exists := cl.connections[connectionID]
	if !exists {
		return
	}

	connInfo.LastActivity = time.Now()
	connInfo.BytesSent += bytesSent
	connInfo.BytesReceived += bytesReceived
}

// GetConnectionInfo returns information about a specific connection
func (cl *ConnectionLimiter) GetConnectionInfo(connectionID string) *ConnectionInfo {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	connInfo, exists := cl.connections[connectionID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	infoCopy := *connInfo
	return &infoCopy
}

// GetConnectionsByIP returns all connections from a specific IP
func (cl *ConnectionLimiter) GetConnectionsByIP(ip string) []*ConnectionInfo {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	var connections []*ConnectionInfo
	for _, connInfo := range cl.connections {
		if connInfo.IP == ip {
			infoCopy := *connInfo
			connections = append(connections, &infoCopy)
		}
	}

	return connections
}

// GetAllConnections returns all active connections
func (cl *ConnectionLimiter) GetAllConnections() map[string]*ConnectionInfo {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	connections := make(map[string]*ConnectionInfo)
	for id, connInfo := range cl.connections {
		infoCopy := *connInfo
		connections[id] = &infoCopy
	}

	return connections
}

// extractIP extracts IP address from a network address string
func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// Try parsing as just IP
		if ip := net.ParseIP(addr); ip != nil {
			return ip.String()
		}
		return ""
	}

	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}

	return host // Return hostname if not an IP
}

// cleanupRoutine periodically cleans up stale connections
func (cl *ConnectionLimiter) cleanupRoutine() {
	ticker := time.NewTicker(cl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !cl.isRunning {
				return
			}
			cl.cleanup()
		}
	}
}

// cleanup removes stale connections
func (cl *ConnectionLimiter) cleanup() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-cl.config.ConnectionTimeout)
	removed := 0

	for connectionID, connInfo := range cl.connections {
		if connInfo.LastActivity.Before(cutoff) {
			// Remove stale connection
			delete(cl.connections, connectionID)
			cl.ipConnections[connInfo.IP]--
			if cl.ipConnections[connInfo.IP] <= 0 {
				delete(cl.ipConnections, connInfo.IP)
			}
			cl.totalConnections--
			removed++
		}
	}

	cl.lastCleanup = now

	if removed > 0 {
		cl.logger.Debug("Cleaned up stale connections", "removed", removed)
	}
}

// GetStats returns connection limiter statistics
func (cl *ConnectionLimiter) GetStats() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	rejectionRate := 0.0
	totalAttempts := cl.totalAccepted + cl.totalRejected
	if totalAttempts > 0 {
		rejectionRate = float64(cl.totalRejected) / float64(totalAttempts) * 100
	}

	return map[string]interface{}{
		"is_running":        cl.isRunning,
		"total_connections": cl.totalConnections,
		"total_accepted":    cl.totalAccepted,
		"total_rejected":    cl.totalRejected,
		"rejection_rate":    rejectionRate,
		"unique_ips":        len(cl.ipConnections),
		"last_cleanup":      cl.lastCleanup,
		"config": map[string]interface{}{
			"max_connections":        cl.config.MaxConnections,
			"max_connections_per_ip": cl.config.MaxConnectionsPerIP,
			"connection_timeout":     cl.config.ConnectionTimeout,
			"cleanup_interval":       cl.config.CleanupInterval,
		},
	}
}

// GetIPStats returns statistics per IP
func (cl *ConnectionLimiter) GetIPStats() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	ipStats := make(map[string]interface{})
	for ip, count := range cl.ipConnections {
		ipStats[ip] = map[string]interface{}{
			"connection_count": count,
		}
	}

	return ipStats
}

// IsIPBlocked checks if an IP should be blocked (Phase 1: basic check)
func (cl *ConnectionLimiter) IsIPBlocked(ip string) bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	// Check if IP has reached connection limit
	ipCount := cl.ipConnections[ip]
	return ipCount >= cl.config.MaxConnectionsPerIP
}

// Reset resets the connection limiter statistics
func (cl *ConnectionLimiter) Reset() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.connections = make(map[string]*ConnectionInfo)
	cl.ipConnections = make(map[string]int)
	cl.totalConnections = 0
	cl.totalAccepted = 0
	cl.totalRejected = 0
	cl.lastCleanup = time.Now()

	cl.logger.Info("Connection limiter statistics reset")
}

// IsRunning returns whether the connection limiter is running
func (cl *ConnectionLimiter) IsRunning() bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.isRunning
}

// GetTotalConnections returns the current total connection count
func (cl *ConnectionLimiter) GetTotalConnections() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.totalConnections
}
