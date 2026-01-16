package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Host wraps libp2p host and provides blockchain-specific networking
type Host struct {
	h      host.Host
	logger *zap.Logger
}

// NewHost creates and initializes a new libp2p host
func NewHost(ctx context.Context, listenAddrs []multiaddr.Multiaddr, priv crypto.PrivKey, logger *zap.Logger) (*Host, error) {
	// Create host with default settings
	opts := []libp2p.Option{
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.DefaultConnectionManager,
	}

	// Add listen addresses
	if len(listenAddrs) > 0 {
		opts = append(opts, libp2p.ListenAddrs(listenAddrs...))
	} else {
		// Default: listen on localhost TCP and QUIC
		opts = append(opts, libp2p.ListenAddrStrings(
			"/ip4/127.0.0.1/tcp/0",
		))
	}

	// Add identity
	opts = append(opts, libp2p.Identity(priv))

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	logger.Info("libp2p host created",
		zap.String("peerID", h.ID().String()),
		zap.Any("listenAddrs", h.Addrs()),
	)

	return &Host{
		h:      h,
		logger: logger,
	}, nil
}

// ID returns the peer ID of this host
func (host *Host) ID() peer.ID {
	return host.h.ID()
}

// Addrs returns the multiaddresses this peer is listening on
func (host *Host) Addrs() []multiaddr.Multiaddr {
	return host.h.Addrs()
}

// Connect connects to a peer
func (host *Host) Connect(ctx context.Context, addr multiaddr.Multiaddr) error {
	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	if err := host.h.Connect(ctx, *info); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	host.logger.Debug("connected to peer", zap.String("peerID", info.ID.String()))
	return nil
}

// GetPeers returns list of connected peers
func (host *Host) GetPeers() []peer.ID {
	return host.h.Network().Peers()
}

// PeerCount returns number of connected peers
func (host *Host) PeerCount() int {
	return len(host.GetPeers())
}

// Close closes the host and cleans up resources
func (host *Host) Close() error {
	return host.h.Close()
}

// GetConnections returns currently open connections
func (host *Host) GetConnections() []peer.ID {
	return host.h.Network().Peers()
}

// GetPeerAddrs returns addresses for a peer
func (host *Host) GetPeerAddrs(p peer.ID) []multiaddr.Multiaddr {
	return host.h.Peerstore().Addrs(p)
}

// GetHost returns the underlying libp2p host for internal use
func (host *Host) GetHost() host.Host {
	return host.h
}
