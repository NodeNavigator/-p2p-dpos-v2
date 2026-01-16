package p2p

import "github.com/libp2p/go-libp2p/core/host"

// Unexported accessor for internal host - allows node.go to access the libp2p host
func (h *Host) H() host.Host {
	return h.h
}
