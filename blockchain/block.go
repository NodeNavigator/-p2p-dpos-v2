package blockchain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/example/p2p-dpos/utils"
)

// TxType represents the type of transaction
type TxType uint8

const (
	TxTypeTransfer TxType = iota
	TxTypeValidatorRegister
	TxTypeDelegate
	TxTypeUndelegate
)

// Transaction represents a signed transaction
type Transaction struct {
	ID        string // Hash of (From, To, Amount, Nonce, Type, Data)
	From      string // Public key (peer ID)
	To        string // Recipient or validator ID
	Amount    uint64
	Nonce     uint64 // Sequence number to prevent replay
	Type      TxType
	Data      []byte // Extra data for validator/delegate ops
	Signature []byte // Ed25519 signature
	Timestamp int64  // milliseconds
}

// Hash computes the transaction hash (excludes Signature)
func (t *Transaction) Hash() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(t.Type))
	buf.WriteString(t.From)
	buf.WriteString(t.To)
	binary.Write(buf, binary.BigEndian, t.Amount)
	binary.Write(buf, binary.BigEndian, t.Nonce)
	buf.Write(t.Data)
	binary.Write(buf, binary.BigEndian, t.Timestamp)

	return utils.HashHex(buf.Bytes())
}

// ToBytes serializes transaction for signing (excludes Signature and ID)
func (t *Transaction) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(t.Type))
	buf.WriteString(t.From)
	buf.WriteString(t.To)
	binary.Write(buf, binary.BigEndian, t.Amount)
	binary.Write(buf, binary.BigEndian, t.Nonce)
	buf.Write(t.Data)
	binary.Write(buf, binary.BigEndian, t.Timestamp)
	return buf.Bytes()
}

// Block represents a block in the blockchain
type Block struct {
	Height     uint64          // Block number
	Timestamp  int64           // milliseconds
	Hash       string          // Block hash
	PrevHash   string          // Parent block hash
	Proposer   string          // Public key of proposer
	Signature  []byte          // Block signature by proposer
	Txs        []*Transaction  // Transactions in block
	TxRoot     string          // Merkle root of transactions
	StateRoot  string          // State root hash
	Validators map[string]bool // Current validators in this block
}

// CalculateHash computes the block hash (excludes Hash and Signature fields)
func (b *Block) CalculateHash() string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, b.Height)
	binary.Write(buf, binary.BigEndian, b.Timestamp)
	buf.WriteString(b.PrevHash)
	buf.WriteString(b.Proposer)
	buf.WriteString(b.TxRoot)
	buf.WriteString(b.StateRoot)
	binary.Write(buf, binary.BigEndian, uint32(len(b.Validators)))
	for validator := range b.Validators {
		buf.WriteString(validator)
	}
	return utils.HashHex(buf.Bytes())
}

// ToBytes serializes block for signing (excludes Hash and Signature)
func (b *Block) ToBytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, b.Height)
	binary.Write(buf, binary.BigEndian, b.Timestamp)
	buf.WriteString(b.PrevHash)
	buf.WriteString(b.Proposer)
	buf.WriteString(b.TxRoot)
	buf.WriteString(b.StateRoot)
	binary.Write(buf, binary.BigEndian, uint32(len(b.Validators)))
	for validator := range b.Validators {
		buf.WriteString(validator)
	}
	return buf.Bytes()
}

// NewBlock creates a new block
func NewBlock(height uint64, prevHash string, proposer string) *Block {
	return &Block{
		Height:     height,
		Timestamp:  time.Now().UnixMilli(),
		PrevHash:   prevHash,
		Proposer:   proposer,
		Txs:        make([]*Transaction, 0),
		Validators: make(map[string]bool),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	b.Txs = append(b.Txs, tx)
	return nil
}

// AddValidator adds a validator to the block's validator set
func (b *Block) AddValidator(validatorID string) {
	b.Validators[validatorID] = true
}

// ComputeTxRoot computes the merkle root of transactions
func (b *Block) ComputeTxRoot() string {
	if len(b.Txs) == 0 {
		return utils.HashHex([]byte{})
	}

	var hashes [][]byte
	for _, tx := range b.Txs {
		hashes = append(hashes, utils.Hash([]byte(tx.ID)))
	}
	return utils.HashHex(utils.MerkleRoot(hashes))
}

// Finalize prepares block for mining (computes txroot and hash)
func (b *Block) Finalize(stateRoot string) {
	b.TxRoot = b.ComputeTxRoot()
	b.StateRoot = stateRoot
	b.Hash = b.CalculateHash()
}
