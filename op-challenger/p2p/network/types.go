package network

import (
	"crypto/ecdsa"
	"math/big"
	"time"
)

// NodeInfo represents information about a node in the network
type NodeInfo struct {
	ID          string
	Address     string
	PublicKey   *ecdsa.PublicKey
	LastSeen    time.Time
	Distance    *big.Int            // XOR distance
}

// GameInfo represents information about a game
type GameInfo struct {
	GameAddress     string
	Status          GameStatus
	LastMove        time.Time
	CurrentClaim    uint64
	MyRole          PlayerRole
	LastValidation  time.Time
}

// GameStatus represents the status of a game
type GameStatus int

const (
	GameStatusActive GameStatus = iota
	GameStatusCompleted
	GameStatusFailed
)

// PlayerRole represents the role of a player in a game
type PlayerRole int

const (
	PlayerRoleChallenger PlayerRole = iota
	PlayerRoleDefender
)

// DiscoveryConfig contains configuration for discovery service
type DiscoveryConfig struct {
	K           int           // Bucket size (usually 20)
	Alpha       int           // Parallel requests (usually 3)
	Timeout     time.Duration // Discovery timeout
	Bootstrap   []string      // Bootstrap nodes
}

// DefaultDiscoveryConfig returns default discovery configuration
func DefaultDiscoveryConfig() DiscoveryConfig {
	return DiscoveryConfig{
		K:         20,
		Alpha:     3,
		Timeout:   30 * time.Second,
		Bootstrap: []string{},
	}
}
