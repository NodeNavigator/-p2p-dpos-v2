package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"
)

// Store provides persistent blockchain storage using LevelDB
type Store struct {
	db     *leveldb.DB
	logger *zap.Logger
}

// NewStore creates a new blockchain store
func NewStore(dbPath string, logger *zap.Logger) (*Store, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &Store{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveBlock persists a block to storage
func (s *Store) SaveBlock(block *Block) error {
	key := []byte(fmt.Sprintf("block:%d", block.Height))
	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	if err := s.db.Put(key, data, nil); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Also save height -> hash index
	hashKey := []byte(fmt.Sprintf("height:%d", block.Height))
	if err := s.db.Put(hashKey, []byte(block.Hash), nil); err != nil {
		return fmt.Errorf("failed to save height index: %w", err)
	}

	s.logger.Debug("block saved", zap.Uint64("height", block.Height), zap.String("hash", block.Hash))
	return nil
}

// GetBlock retrieves a block by height
func (s *Store) GetBlock(height uint64) (*Block, error) {
	key := []byte(fmt.Sprintf("block:%d", height))
	data, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("block not found at height %d", height)
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &block, nil
}

// SaveState persists the chain state
func (s *Store) SaveState(state *ChainState) error {
	state.Mu.RLock()
	height := state.Height
	hash := state.LastBlockHash
	balances := make(map[string]uint64)
	for k, v := range state.Balances {
		balances[k] = v
	}
	state.Mu.RUnlock()

	key := []byte(fmt.Sprintf("state:%d", height))
	data := map[string]interface{}{
		"height":        height,
		"lastBlockHash": hash,
		"balances":      balances,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := s.db.Put(key, jsonData, nil); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	s.logger.Debug("state saved", zap.Uint64("height", height))
	return nil
}

// SaveTransaction persists a transaction
func (s *Store) SaveTransaction(tx *Transaction) error {
	key := []byte(fmt.Sprintf("tx:%s", tx.ID))
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	if err := s.db.Put(key, data, nil); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (s *Store) GetTransaction(txID string) (*Transaction, error) {
	key := []byte(fmt.Sprintf("tx:%s", txID))
	data, err := s.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, fmt.Errorf("transaction not found: %s", txID)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	var tx Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}
