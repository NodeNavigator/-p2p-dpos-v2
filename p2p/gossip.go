package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// Subscription wraps a channel for receiving messages
type Subscription struct {
	ch     chan *Message
	cancel context.CancelFunc
}

// Next returns the next message or error
func (s *Subscription) Next(ctx context.Context) (*Message, error) {
	select {
	case msg := <-s.ch:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Cancel stops the subscription
func (s *Subscription) Cancel() {
	s.cancel()
}

// Gossip provides simple in-memory pubsub for testing
type Gossip struct {
	host        host.Host
	subscribers map[string][]*Subscription
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewGossip creates a new gossip service
func NewGossip(ctx context.Context, h host.Host, logger *zap.Logger) (*Gossip, error) {
	return &Gossip{
		host:        h,
		subscribers: make(map[string][]*Subscription),
		logger:      logger,
	}, nil
}

// Subscribe subscribes to a topic and returns a subscription channel
func (g *Gossip) Subscribe(ctx context.Context, topicName string) (*Subscription, error) {
	ctx, cancel := context.WithCancel(ctx)
	sub := &Subscription{
		ch:     make(chan *Message, 100),
		cancel: cancel,
	}

	g.mu.Lock()
	g.subscribers[topicName] = append(g.subscribers[topicName], sub)
	g.mu.Unlock()

	g.logger.Debug("subscribed to topic", zap.String("topic", topicName))
	return sub, nil
}

// Publish publishes a message to a topic
func (g *Gossip) Publish(ctx context.Context, topicName string, msg *Message) error {
	g.mu.RLock()
	subs := g.subscribers[topicName]
	g.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			g.logger.Warn("subscription channel full, dropping message", zap.String("topic", topicName))
		}
	}

	return nil
}

// PublishBlock publishes a block
func (g *Gossip) PublishBlock(ctx context.Context, block *interface{}) error {
	var msg *Message
	switch v := (*block).(type) {
	case *Message:
		msg = v
	default:
		return fmt.Errorf("invalid block message type")
	}
	return g.Publish(ctx, "blocks", msg)
}

// PublishTransaction publishes a transaction
func (g *Gossip) PublishTransaction(ctx context.Context, txMsg *interface{}) error {
	var msg *Message
	switch v := (*txMsg).(type) {
	case *Message:
		msg = v
	default:
		return fmt.Errorf("invalid transaction message type")
	}
	return g.Publish(ctx, "transactions", msg)
}

// Close closes all subscriptions
func (g *Gossip) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, subs := range g.subscribers {
		for _, sub := range subs {
			sub.Cancel()
		}
	}
	g.subscribers = make(map[string][]*Subscription)
	return nil
}

// ListPeers returns peers (stub for compatibility)
func (g *Gossip) ListPeers(topicName string) []peer.ID {
	return g.host.Network().Peers()
}

// TopicPeerCount returns number of peers on a topic
func (g *Gossip) TopicPeerCount(topicName string) int {
	return len(g.ListPeers(topicName))
}
