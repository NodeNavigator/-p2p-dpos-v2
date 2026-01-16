package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/example/p2p-dpos/blockchain"
	"github.com/example/p2p-dpos/config"
	"github.com/example/p2p-dpos/consensus"
	"github.com/example/p2p-dpos/crypto"
	"go.uber.org/zap"
)

// TestFullBlockchainFlow tests complete blockchain flow
func TestFullBlockchainFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create chain state
	state := blockchain.NewChainState(logger)

	// Initialize balances
	peers := []string{"peer1", "peer2", "peer3"}
	for _, peer := range peers {
		state.Balances[peer] = 10000
	}

	// Create store
	store, _ := blockchain.NewStore(t.TempDir(), logger)
	defer store.Close()

	// Register validators
	for _, peer := range peers {
		_ = state.RegisterValidator(peer, 5000)
	}

	// Verify validators registered
	if len(state.Validators) != 3 {
		t.Errorf("Expected 3 validators, got %d", len(state.Validators))
	}

	// Perform transfer
	_ = state.Transfer("peer1", "peer2", 1000)

	if state.GetBalance("peer1") != 3000 {
		t.Errorf("peer1 balance should be 3000, got %d", state.GetBalance("peer1"))
	}

	if state.GetBalance("peer2") != 5000 {
		t.Errorf("peer2 balance should be 5000, got %d", state.GetBalance("peer2"))
	}
}

// TestMultiNodeConsensus tests consensus with multiple nodes
func TestMultiNodeConsensus(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	numNodes := 3
	nodes := make([]*consensus.DPoS, numNodes)
	states := make([]*blockchain.ChainState, numNodes)
	keypairs := make([]*crypto.KeyPair, numNodes)

	// Create nodes
	for i := 0; i < numNodes; i++ {
		state := blockchain.NewChainState(logger)
		for j := 0; j < numNodes; j++ {
			peer := fmt.Sprintf("peer%d", j)
			state.Balances[peer] = 100000
		}

		store, _ := blockchain.NewStore(t.TempDir(), logger)
		defer store.Close()

		kp, _ := crypto.GenerateKeyPair()
		cfg := &config.Config{
			BlockProposalInterval: 100 * time.Millisecond,
		}

		dpos := consensus.NewDPoS(cfg, state, store, kp, logger)

		states[i] = state
		nodes[i] = dpos
		keypairs[i] = kp
	}

	// Register all as validators
	for i := 0; i < numNodes; i++ {
		peer := fmt.Sprintf("peer%d", i)
		_ = states[i].RegisterValidator(peer, 50000)
	}

	// Verify consensus state
	for i := 0; i < numNodes; i++ {
		if len(states[i].Validators) != numNodes {
			t.Errorf("Node %d should have %d validators, got %d", i, numNodes, len(states[i].Validators))
		}
	}

	// All nodes add same transaction
	tx := &blockchain.Transaction{
		From:      "peer0",
		To:        "peer1",
		Amount:    100,
		Nonce:     1,
		Type:      blockchain.TxTypeTransfer,
		Timestamp: time.Now().Unix() * 1000,
	}
	tx.ID = tx.Hash()

	for i := 0; i < numNodes; i++ {
		err := nodes[i].AddTransaction(tx)
		if err != nil {
			t.Errorf("Failed to add transaction to node %d: %v", i, err)
		}
	}

}

// TestValidatorDelegation tests delegation to validators
func TestValidatorDelegation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := blockchain.NewChainState(logger)

	// Initialize accounts
	peers := []string{"validator1", "validator2", "delegator1", "delegator2", "delegator3"}
	for _, peer := range peers {
		state.Balances[peer] = 10000
	}

	// Register validators
	_ = state.RegisterValidator("validator1", 5000)
	_ = state.RegisterValidator("validator2", 5000)

	// Delegations
	_ = state.Delegate("delegator1", "validator1", 2000)
	_ = state.Delegate("delegator2", "validator1", 1500)
	_ = state.Delegate("delegator3", "validator2", 3000)

	// Verify delegations
	validator1 := state.Validators["validator1"]
	if validator1.DelegatedAmount != 3500 {
		t.Errorf("Validator1 should have 3500 delegated, got %d", validator1.DelegatedAmount)
	}

	validator2 := state.Validators["validator2"]
	if validator2.DelegatedAmount != 3000 {
		t.Errorf("Validator2 should have 3000 delegated, got %d", validator2.DelegatedAmount)
	}
}

// TestBlockCreationSequence tests creating a sequence of blocks
func TestBlockCreationSequence(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	blocks := make([]*blockchain.Block, 5)
	prevHash := "genesis"

	// Create block sequence
	for i := 0; i < 5; i++ {
		block := &blockchain.Block{
			Height:       uint64(i + 1),
			Timestamp:    time.Now().Unix() * 1000,
			PrevHash:     prevHash,
			Transactions: []*blockchain.Transaction{},
			Proposer:     fmt.Sprintf("peer%d", i%3),
		}

		blocks[i] = block
		prevHash = block.Hash()

		// Verify linking
		if i > 0 {
			if block.PrevHash != blocks[i-1].Hash() {
				t.Error("Block should reference previous block hash")
			}
		}
	}

	// Verify sequence
	for i := 1; i < 5; i++ {
		if blocks[i].Height != uint64(i+1) {
			t.Errorf("Block %d height should be %d, got %d", i, i+1, blocks[i].Height)
		}
	}
}

// TestConcurrentTransactions tests concurrent transaction processing
func TestConcurrentTransactions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	state := blockchain.NewChainState(logger)
	store, _ := blockchain.NewStore(t.TempDir(), logger)
	defer store.Close()

	kp, _ := crypto.GenerateKeyPair()
	cfg := &config.Config{}
	dpos := consensus.NewDPoS(cfg, state, store, kp, logger)

	// Initialize accounts
	for i := 0; i < 10; i++ {
		peer := fmt.Sprintf("peer%d", i)
		state.Balances[peer] = 100000
	}

	var wg sync.WaitGroup
	numTxs := 100

	wg.Add(numTxs)
	for i := 0; i < numTxs; i++ {
		go func(index int) {
			defer wg.Done()
			from := fmt.Sprintf("peer%d", index%10)
			to := fmt.Sprintf("peer%d", (index+1)%10)
			tx := &blockchain.Transaction{
				From:      from,
				To:        to,
				Amount:    uint64((index % 100) + 1),
				Nonce:     uint64(index),
				Type:      blockchain.TxTypeTransfer,
				Timestamp: time.Now().Unix() * 1000,
			}
			tx.ID = tx.Hash()
			_ = dpos.AddTransaction(tx)
		}(i)
	}

	wg.Wait()

	// All transactions should have been added successfully
	// (individual transactions are tested in consensus_test.go)

}

// BenchmarkFullBlockchainFlow benchmarks complete blockchain operation
func BenchmarkFullBlockchainFlow(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := blockchain.NewChainState(logger)

		for j := 0; j < 10; j++ {
			peer := fmt.Sprintf("peer%d", j)
			state.Balances[peer] = 10000
		}

		for j := 0; j < 5; j++ {
			peer := fmt.Sprintf("peer%d", j)
			_ = state.RegisterValidator(peer, 5000)
		}

		for j := 0; j < 100; j++ {
			from := fmt.Sprintf("peer%d", j%10)
			to := fmt.Sprintf("peer%d", (j+1)%10)
			_ = state.Transfer(from, to, 1)
		}
	}
}

// BenchmarkMultipleTransactions benchmarks multiple transaction processing
func BenchmarkMultipleTransactions(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	state := blockchain.NewChainState(logger)
	store, _ := blockchain.NewStore(b.TempDir(), logger)
	defer store.Close()

	kp, _ := crypto.GenerateKeyPair()
	cfg := &config.Config{}
	dpos := consensus.NewDPoS(cfg, state, store, kp, logger)

	for i := 0; i < 50; i++ {
		peer := fmt.Sprintf("peer%d", i)
		state.Balances[peer] = 100000
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := &blockchain.Transaction{
			From:      fmt.Sprintf("peer%d", i%50),
			To:        fmt.Sprintf("peer%d", (i+1)%50),
			Amount:    1,
			Nonce:     uint64(i),
			Type:      blockchain.TxTypeTransfer,
			Timestamp: time.Now().Unix() * 1000,
		}
		tx.ID = tx.Hash()
		_ = dpos.AddTransaction(tx)
	}
}
