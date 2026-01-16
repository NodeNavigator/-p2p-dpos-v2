package config

import (
	"path/filepath"
	"time"

	"github.com/multiformats/go-multiaddr"
)

// Config holds all node configuration
type Config struct {
	// Node identity and network
	DataDir    string
	ListenAddr []multiaddr.Multiaddr
	BootPeers  []multiaddr.Multiaddr

	// Network timing
	NetworkTimeout    time.Duration
	DiscoveryInterval time.Duration

	// Blockchain parameters
	BlockTime           time.Duration
	MaxValidators       int
	MinValidatorStake   uint64
	BlockProductionTime time.Duration

	// DPoS parameters
	ValidatorSetSize  int
	ProposerRotation  bool
	FinalizationStake uint64 // % of stake needed to finalize

	// Storage
	DBPath string

	// Logging
	LogLevel string

	// Initial balances for testing
	InitialBalances map[string]uint64
}

// DefaultConfig returns sensible defaults
func DefaultConfig(dataDir string) *Config {
	return &Config{
		DataDir:           dataDir,
		ListenAddr:        []multiaddr.Multiaddr{},
		BootPeers:         []multiaddr.Multiaddr{},
		NetworkTimeout:    30 * time.Second,
		DiscoveryInterval: 5 * time.Second,
		BlockTime:         10 * time.Second,
		MaxValidators:     10,
		MinValidatorStake: 100,
		ValidatorSetSize:  3,
		ProposerRotation:  true,
		FinalizationStake: 67, // 2/3 + 1
		DBPath:            filepath.Join(dataDir, "blockchain.db"),
		LogLevel:          "info",
		InitialBalances:   make(map[string]uint64),
	}
}
