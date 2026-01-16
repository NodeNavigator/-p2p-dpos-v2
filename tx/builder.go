package tx

import (
	"time"

	"crypto/ed25519"

	"github.com/example/p2p-dpos/blockchain"
	"github.com/example/p2p-dpos/crypto"
)

// Builder provides a fluent interface for building transactions
type Builder struct {
	from      string
	to        string
	amount    uint64
	nonce     uint64
	txType    blockchain.TxType
	data      []byte
	timestamp int64
}

// NewBuilder creates a new transaction builder
func NewBuilder() *Builder {
	return &Builder{
		timestamp: time.Now().UnixMilli(),
		data:      []byte{},
	}
}

// From sets the sender (public key)
func (b *Builder) From(publicKey string) *Builder {
	b.from = publicKey
	return b
}

// To sets the recipient or target (for validators, this is the validator ID)
func (b *Builder) To(recipient string) *Builder {
	b.to = recipient
	return b
}

// Amount sets the transaction amount
func (b *Builder) Amount(amount uint64) *Builder {
	b.amount = amount
	return b
}

// Nonce sets the transaction nonce (sequence number)
func (b *Builder) Nonce(nonce uint64) *Builder {
	b.nonce = nonce
	return b
}

// Type sets the transaction type
func (b *Builder) Type(txType blockchain.TxType) *Builder {
	b.txType = txType
	return b
}

// Data sets additional data
func (b *Builder) Data(data []byte) *Builder {
	b.data = data
	return b
}

// Build creates the transaction (unsigned)
func (b *Builder) Build() *blockchain.Transaction {
	return &blockchain.Transaction{
		From:      b.from,
		To:        b.to,
		Amount:    b.amount,
		Nonce:     b.nonce,
		Type:      b.txType,
		Data:      b.data,
		Timestamp: b.timestamp,
	}
}

// SignAndBuild creates and signs the transaction
func (b *Builder) SignAndBuild(kp *crypto.KeyPair) (*blockchain.Transaction, error) {
	tx := b.Build()

	// Compute transaction ID before signing
	tx.ID = tx.Hash()

	// Sign the transaction
	tx.Signature = kp.Sign(tx.ToBytes())

	return tx, nil
}

// Transfer creates a simple transfer transaction
func Transfer(from, to string, amount uint64, nonce uint64, kp *crypto.KeyPair) (*blockchain.Transaction, error) {
	tx := NewBuilder().
		From(from).
		To(to).
		Amount(amount).
		Nonce(nonce).
		Type(blockchain.TxTypeTransfer).
		Build()

	tx.ID = tx.Hash()
	tx.Signature = kp.Sign(tx.ToBytes())

	return tx, nil
}

// RegisterValidator creates a validator registration transaction
func RegisterValidator(from string, stakeAmount uint64, nonce uint64, kp *crypto.KeyPair) (*blockchain.Transaction, error) {
	tx := NewBuilder().
		From(from).
		To(from). // Validator ID is own public key
		Amount(stakeAmount).
		Nonce(nonce).
		Type(blockchain.TxTypeValidatorRegister).
		Build()

	tx.ID = tx.Hash()
	tx.Signature = kp.Sign(tx.ToBytes())

	return tx, nil
}

// Delegate creates a delegation transaction
func Delegate(delegator, validator string, amount uint64, nonce uint64, kp *crypto.KeyPair) (*blockchain.Transaction, error) {
	tx := NewBuilder().
		From(delegator).
		To(validator).
		Amount(amount).
		Nonce(nonce).
		Type(blockchain.TxTypeDelegate).
		Build()

	tx.ID = tx.Hash()
	tx.Signature = kp.Sign(tx.ToBytes())

	return tx, nil
}

// Undelegate creates an undelegation transaction
func Undelegate(delegator, validator string, amount uint64, nonce uint64, kp *crypto.KeyPair) (*blockchain.Transaction, error) {
	tx := NewBuilder().
		From(delegator).
		To(validator).
		Amount(amount).
		Nonce(nonce).
		Type(blockchain.TxTypeUndelegate).
		Build()

	tx.ID = tx.Hash()
	tx.Signature = kp.Sign(tx.ToBytes())

	return tx, nil
}

// VerifyTransactionSignature verifies a transaction's signature
func VerifyTransactionSignature(tx *blockchain.Transaction, pubKey ed25519.PublicKey) bool {
	return crypto.Verify(pubKey, tx.ToBytes(), tx.Signature)
}
