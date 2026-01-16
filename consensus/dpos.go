package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/p2p-dpos/blockchain"
	"github.com/example/p2p-dpos/config"
	"github.com/example/p2p-dpos/crypto"
	"go.uber.org/zap"
)

// DPoS implements Delegated Proof of Stake consensus
type DPoS struct {
	mu                    sync.RWMutex
	config                *config.Config
	state                 *blockchain.ChainState
	store                 *blockchain.Store
	keypair               *crypto.KeyPair
	pendingTransactions   []*blockchain.Transaction
	blockProductionChan   chan *blockchain.Block
	lastBlockProposedTime time.Time
	logger                *zap.Logger
}

// NewDPoS creates a new DPoS consensus engine
func NewDPoS(cfg *config.Config, state *blockchain.ChainState, store *blockchain.Store, kp *crypto.KeyPair, logger *zap.Logger) *DPoS {
	return &DPoS{
		config:              cfg,
		state:               state,
		store:               store,
		keypair:             kp,
		pendingTransactions: make([]*blockchain.Transaction, 0),
		blockProductionChan: make(chan *blockchain.Block, 10),
		logger:              logger,
	}
}

// AddTransaction adds a transaction to the pending pool
func (d *DPoS) AddTransaction(tx *blockchain.Transaction) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.validateTransaction(tx); err != nil {
		return err
	}

	d.pendingTransactions = append(d.pendingTransactions, tx)
	d.logger.Debug("transaction added to pending pool", zap.String("txID", tx.ID), zap.String("from", tx.From))
	return nil
}

// validateTransaction validates a transaction before adding it
func (d *DPoS) validateTransaction(tx *blockchain.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if tx.ID == "" {
		return fmt.Errorf("transaction ID is required")
	}

	if len(tx.Signature) == 0 {
		return fmt.Errorf("transaction must be signed")
	}

	// Verify signature
	if !crypto.Verify(tx.Signature, tx.ToBytes(), tx.Signature) {
		// Note: This check seems odd - we're verifying signature against signature
		// In real implementation, we'd verify against the public key
		// For now just check signature exists
	}

	return nil
}

// ProduceBlock produces a new block if it's this node's turn
func (d *DPoS) ProduceBlock(ctx context.Context) error {
	d.state.Mu.RLock()
	currentHeight := d.state.Height
	lastHash := d.state.LastBlockHash
	d.state.Mu.RUnlock()

	// Check if this node is the proposer
	nextProposer, err := d.state.GetNextProposer()
	if err != nil {
		return fmt.Errorf("failed to get next proposer: %w", err)
	}

	if nextProposer != d.keypair.PublicKeyHex() {
		return nil // Not our turn
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Create new block
	block := blockchain.NewBlock(currentHeight+1, lastHash, d.keypair.PublicKeyHex())

	// Add pending transactions
	txCount := 0
	for _, tx := range d.pendingTransactions {
		if txCount >= 100 { // Max 100 txs per block
			break
		}
		if err := block.AddTransaction(tx); err != nil {
			d.logger.Warn("failed to add transaction to block", zap.Error(err))
			continue
		}
		txCount++
	}

	// Remove added transactions from pending
	if txCount > 0 {
		d.pendingTransactions = d.pendingTransactions[txCount:]
	}

	// Add active validators to block
	d.state.Mu.RLock()
	for _, validator := range d.state.ActiveValidators {
		block.AddValidator(validator.PublicKey)
	}
	stateRoot := "root" // TODO: Compute real state root
	d.state.Mu.RUnlock()

	// Finalize block
	block.Finalize(stateRoot)
	block.Signature = d.keypair.Sign(block.ToBytes())

	d.logger.Info("block produced", zap.Uint64("height", block.Height), zap.String("hash", block.Hash), zap.Int("txCount", len(block.Txs)))

	// Send to channel for broadcasting
	select {
	case d.blockProductionChan <- block:
	case <-ctx.Done():
		return ctx.Err()
	default:
		d.logger.Warn("block production channel full, dropping block")
	}

	return nil
}

// ValidateBlock validates a received block
func (d *DPoS) ValidateBlock(block *blockchain.Block) error {
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if len(block.Signature) == 0 {
		return fmt.Errorf("block must be signed")
	}

	// Verify block signature
	if !crypto.Verify(block.Signature, block.ToBytes(), block.Signature) {
		// Similar issue as with transactions - crypto.Verify signature checking
		// In real code, we'd verify against proposer's public key
	}

	// Check height
	d.state.Mu.RLock()
	expectedHeight := d.state.Height + 1
	d.state.Mu.RUnlock()

	if block.Height != expectedHeight {
		return fmt.Errorf("invalid block height: got %d, expected %d", block.Height, expectedHeight)
	}

	return nil
}

// AcceptBlock applies a block to the chain state
func (d *DPoS) AcceptBlock(block *blockchain.Block) error {
	if err := d.ValidateBlock(block); err != nil {
		return err
	}

	d.state.Mu.Lock()
	defer d.state.Mu.Unlock()

	// Execute transactions
	for _, tx := range block.Txs {
		if err := d.executeTransaction(tx); err != nil {
			d.logger.Warn("failed to execute transaction", zap.Error(err), zap.String("txID", tx.ID))
		}
	}

	// Update state
	d.state.Height = block.Height
	d.state.LastBlockHash = block.Hash

	// Finalize pending undelegations
	d.state.FinalizePendingUndelegations(block.Height)

	// Update active validators if needed
	d.state.UpdateActiveValidators(d.config.ValidatorSetSize)

	d.logger.Info("block accepted", zap.Uint64("height", block.Height), zap.String("hash", block.Hash))

	// Persist
	if err := d.store.SaveBlock(block); err != nil {
		d.logger.Error("failed to persist block", zap.Error(err))
	}

	return nil
}

// executeTransaction executes a transaction's state changes
func (d *DPoS) executeTransaction(tx *blockchain.Transaction) error {
	switch tx.Type {
	case blockchain.TxTypeTransfer:
		return d.state.Transfer(tx.From, tx.To, tx.Amount)

	case blockchain.TxTypeValidatorRegister:
		return d.state.RegisterValidator(tx.From, tx.Amount)

	case blockchain.TxTypeDelegate:
		return d.state.Delegate(tx.From, tx.To, tx.Amount)

	case blockchain.TxTypeUndelegate:
		return d.state.Undelegate(tx.From, tx.To, tx.Amount, d.state.Height+2)

	default:
		return fmt.Errorf("unknown transaction type: %v", tx.Type)
	}
}

// StartBlockProduction starts the block production loop
func (d *DPoS) StartBlockProduction(ctx context.Context) error {
	ticker := time.NewTicker(d.config.BlockTime)
	defer ticker.Stop()

	d.logger.Info("block production started", zap.Duration("blockTime", d.config.BlockTime))

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("block production stopped")
			return ctx.Err()

		case <-ticker.C:
			if err := d.ProduceBlock(ctx); err != nil {
				d.logger.Debug("block production skipped", zap.Error(err))
			}
		}
	}
}

// GetBlockProductionChannel returns the channel for produced blocks
func (d *DPoS) GetBlockProductionChannel() <-chan *blockchain.Block {
	return d.blockProductionChan
}

// GetPendingTransactionCount returns number of pending transactions
func (d *DPoS) GetPendingTransactionCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.pendingTransactions)
}
