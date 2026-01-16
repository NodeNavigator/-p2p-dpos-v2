package node

import (
	"context"
	"fmt"

	"github.com/example/p2p-dpos/blockchain"
	"github.com/example/p2p-dpos/config"
	"github.com/example/p2p-dpos/consensus"
	"github.com/example/p2p-dpos/crypto"
	"github.com/example/p2p-dpos/p2p"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Node represents a full P2P-DPoS node
type Node struct {
	cfg       *config.Config
	logger    *zap.Logger
	keypair   *crypto.KeyPair
	host      *p2p.Host
	discovery *p2p.Discovery
	gossip    *p2p.Gossip
	store     *blockchain.Store
	state     *blockchain.ChainState
	consensus *consensus.DPoS
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewNode creates a new blockchain node
func NewNode(cfg *config.Config, logger *zap.Logger) (*Node, error) {
	// Generate or load keypair
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	logger.Info("node identity generated", zap.String("peerID", kp.PublicKeyHex()))

	// Initialize blockchain storage
	store, err := blockchain.NewStore(cfg.DBPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	// Initialize chain state
	state := blockchain.NewChainState(logger)

	// Set initial balances
	for account, balance := range cfg.InitialBalances {
		state.Balances[account] = balance
	}

	// Initialize consensus engine
	dpos := consensus.NewDPoS(cfg, state, store, kp, logger)

	ctx, cancel := context.WithCancel(context.Background())

	return &Node{
		cfg:       cfg,
		logger:    logger,
		keypair:   kp,
		store:     store,
		state:     state,
		consensus: dpos,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start initializes networking and starts the node
func (n *Node) Start() error {
	// Create libp2p host
	priv, err := crypto.PrivateKeyToLibp2p(n.keypair)
	if err != nil {
		// If conversion fails, create new libp2p key
		priv, _, err = crypto.GenerateLibp2pKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate libp2p keypair: %w", err)
		}
	}

	listenAddrs := []multiaddr.Multiaddr{}
	if len(n.cfg.ListenAddr) == 0 {
		// Default listen address
		addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
		listenAddrs = append(listenAddrs, addr)
	} else {
		listenAddrs = n.cfg.ListenAddr
	}

	host, err := p2p.NewHost(n.ctx, listenAddrs, priv, n.logger)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	n.host = host

	n.logger.Info("libp2p host started",
		zap.String("peerID", host.ID().String()),
		zap.Any("addresses", host.Addrs()),
	)

	// Initialize discovery
	n.discovery = p2p.NewDiscovery(host.GetHost(), n.cfg.BootPeers, n.logger)
	if err := n.discovery.Bootstrap(n.ctx); err != nil {
		n.logger.Warn("bootstrap failed", zap.Error(err))
	}

	// Initialize gossip
	gossip, err := p2p.NewGossip(n.ctx, host.GetHost(), n.logger)
	if err != nil {
		return fmt.Errorf("failed to create gossip: %w", err)
	}
	n.gossip = gossip

	// Start block production
	go func() {
		if err := n.consensus.StartBlockProduction(n.ctx); err != nil {
			n.logger.Debug("block production ended", zap.Error(err))
		}
	}()

	// Start block listener
	go n.listenBlocks()

	// Start transaction listener
	go n.listenTransactions()

	// Start block broadcaster
	go n.broadcastBlocks()

	n.logger.Info("node started successfully", zap.String("peerID", host.ID().String()))
	return nil
}

// Stop gracefully stops the node
func (n *Node) Stop() error {
	n.logger.Info("stopping node")
	n.cancel()

	if n.gossip != nil {
		n.gossip.Close()
	}

	if n.host != nil {
		n.host.Close()
	}

	if n.store != nil {
		n.store.Close()
	}

	n.logger.Info("node stopped")
	return nil
}

// listenBlocks listens for incoming blocks
func (n *Node) listenBlocks() {
	sub, err := n.gossip.Subscribe(n.ctx, "blocks")
	if err != nil {
		n.logger.Error("failed to subscribe to blocks", zap.Error(err))
		return
	}
	defer sub.Cancel()

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}

		msg, err := sub.Next(n.ctx)
		if err != nil {
			n.logger.Warn("error receiving block", zap.Error(err))
			continue
		}

		// msg is already a *Message from the channel
		blockMsg, err := p2p.DecodeBlockMessage(msg)
		if err != nil {
			n.logger.Warn("failed to decode block", zap.Error(err))
			continue
		}

		// Accept block
		if err := n.consensus.AcceptBlock(blockMsg.Block); err != nil {
			n.logger.Warn("failed to accept block", zap.Error(err))
		}
	}
}

// listenTransactions listens for incoming transactions
func (n *Node) listenTransactions() {
	sub, err := n.gossip.Subscribe(n.ctx, "transactions")
	if err != nil {
		n.logger.Error("failed to subscribe to transactions", zap.Error(err))
		return
	}
	defer sub.Cancel()

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}

		msg, err := sub.Next(n.ctx)
		if err != nil {
			n.logger.Warn("error receiving transaction", zap.Error(err))
			continue
		}

		// msg is already a *Message from the channel
		txMsg, err := p2p.DecodeTransactionMessage(msg)
		if err != nil {
			n.logger.Warn("failed to decode transaction", zap.Error(err))
			continue
		}

		// Add to pending
		if err := n.consensus.AddTransaction(txMsg.Transaction); err != nil {
			n.logger.Warn("failed to add transaction", zap.Error(err))
		}
	}
}

// broadcastBlocks broadcasts produced blocks
func (n *Node) broadcastBlocks() {
	blockChan := n.consensus.GetBlockProductionChannel()

	for {
		select {
		case <-n.ctx.Done():
			return

		case block := <-blockChan:
			msg, err := p2p.NewBlockMessage(block)
			if err != nil {
				n.logger.Error("failed to create block message", zap.Error(err))
				continue
			}

			if err := n.gossip.Publish(n.ctx, "blocks", msg); err != nil {
				n.logger.Error("failed to publish block", zap.Error(err))
			}
		}
	}
}

// GetState returns the current chain state
func (n *Node) GetState() *blockchain.ChainState {
	return n.state
}

// GetKeypair returns the node's keypair
func (n *Node) GetKeypair() *crypto.KeyPair {
	return n.keypair
}

// GetHost returns the libp2p host
func (n *Node) GetHost() *p2p.Host {
	return n.host
}

// BroadcastTransaction broadcasts a transaction to the network
func (n *Node) BroadcastTransaction(tx *blockchain.Transaction) error {
	msg, err := p2p.NewTransactionMessage(tx)
	if err != nil {
		return fmt.Errorf("failed to create transaction message: %w", err)
	}

	return n.gossip.Publish(n.ctx, "transactions", msg)
}

// AddPendingTransaction adds a transaction to the consensus engine
func (n *Node) AddPendingTransaction(tx *blockchain.Transaction) error {
	return n.consensus.AddTransaction(tx)
}
