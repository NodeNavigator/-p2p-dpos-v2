package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Discovery manages peer discovery and bootstrap
type Discovery struct {
	h         host.Host
	bootPeers []multiaddr.Multiaddr
	logger    *zap.Logger
}

// NewDiscovery creates a new discovery service
func NewDiscovery(h host.Host, bootPeers []multiaddr.Multiaddr, logger *zap.Logger) *Discovery {
	return &Discovery{
		h:         h,
		bootPeers: bootPeers,
		logger:    logger,
	}
}

// Bootstrap connects to bootstrap peers
func (d *Discovery) Bootstrap(ctx context.Context) error {
	for _, addr := range d.bootPeers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			d.logger.Warn("invalid bootstrap peer address", zap.String("addr", addr.String()), zap.Error(err))
			continue
		}

		if err := d.h.Connect(ctx, *info); err != nil {
			d.logger.Warn("failed to connect to bootstrap peer", zap.String("peerID", info.ID.String()), zap.Error(err))
			continue
		}

		d.logger.Info("bootstrapped to peer", zap.String("peerID", info.ID.String()))
	}

	return nil
}

// FindPeers searches for peers (can be enhanced with DHT lookup)
func (d *Discovery) FindPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	// For now, return connected peers
	peers := d.h.Network().Peers()
	var result []peer.AddrInfo

	for _, p := range peers {
		addrs := d.h.Peerstore().Addrs(p)
		if len(addrs) > 0 {
			result = append(result, peer.AddrInfo{
				ID:    p,
				Addrs: addrs,
			})
		}
	}

	return result, nil
}

// AdvertiseService announces this node as a service provider (simple version)
func (d *Discovery) AdvertiseService(ctx context.Context, namespace string) error {
	d.logger.Info("advertising service", zap.String("namespace", namespace))
	return nil
}

// GetPeerInfo returns information about a specific peer
func (d *Discovery) GetPeerInfo(p peer.ID) (peer.AddrInfo, error) {
	addrs := d.h.Peerstore().Addrs(p)
	if len(addrs) == 0 {
		return peer.AddrInfo{}, fmt.Errorf("no addresses known for peer")
	}

	return peer.AddrInfo{
		ID:    p,
		Addrs: addrs,
	}, nil
}

// ConnectPeer connects to a specific peer
func (d *Discovery) ConnectPeer(ctx context.Context, p peer.AddrInfo) error {
	if err := d.h.Connect(ctx, p); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}
	d.logger.Debug("connected to peer", zap.String("peerID", p.ID.String()))
	return nil
}
