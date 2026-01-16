package p2p

import (
	"encoding/json"

	"github.com/example/p2p-dpos/blockchain"
)

// MessageType represents the type of P2P message
type MessageType int

const (
	MessageTypeBlock MessageType = iota
	MessageTypeTransaction
	MessageTypeGetBlocks
	MessageTypeBlockResponse
)

// Message is the envelope for all P2P communication
type Message struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}

// BlockMessage wraps a block for broadcasting
type BlockMessage struct {
	Block *blockchain.Block `json:"block"`
}

// TransactionMessage wraps a transaction for broadcasting
type TransactionMessage struct {
	Transaction *blockchain.Transaction `json:"transaction"`
}

// GetBlocksMessage requests blocks in a range
type GetBlocksMessage struct {
	FromHeight uint64 `json:"fromHeight"`
	ToHeight   uint64 `json:"toHeight"`
}

// BlockResponseMessage responds with blocks
type BlockResponseMessage struct {
	Blocks []*blockchain.Block `json:"blocks"`
}

// NewBlockMessage creates a new block message
func NewBlockMessage(block *blockchain.Block) (*Message, error) {
	payload := &BlockMessage{Block: block}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      MessageTypeBlock,
		Payload:   data,
		Timestamp: block.Timestamp,
	}, nil
}

// NewTransactionMessage creates a new transaction message
func NewTransactionMessage(tx *blockchain.Transaction) (*Message, error) {
	payload := &TransactionMessage{Transaction: tx}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      MessageTypeTransaction,
		Payload:   data,
		Timestamp: tx.Timestamp,
	}, nil
}

// NewGetBlocksMessage creates a get blocks message
func NewGetBlocksMessage(fromHeight, toHeight uint64) (*Message, error) {
	payload := &GetBlocksMessage{
		FromHeight: fromHeight,
		ToHeight:   toHeight,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      MessageTypeGetBlocks,
		Payload:   data,
		Timestamp: 0,
	}, nil
}

// DecodeBlockMessage decodes a block message
func DecodeBlockMessage(msg *Message) (*BlockMessage, error) {
	var payload BlockMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeTransactionMessage decodes a transaction message
func DecodeTransactionMessage(msg *Message) (*TransactionMessage, error) {
	var payload TransactionMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeGetBlocksMessage decodes a get blocks message
func DecodeGetBlocksMessage(msg *Message) (*GetBlocksMessage, error) {
	var payload GetBlocksMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecodeBlockResponseMessage decodes a block response message
func DecodeBlockResponseMessage(msg *Message) (*BlockResponseMessage, error) {
	var payload BlockResponseMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
