package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// KeyPair represents a node's cryptographic identity
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair generates a new Ed25519 keypair
func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// Sign signs a message with the private key
func (kp *KeyPair) Sign(message []byte) []byte {
	return ed25519.Sign(kp.PrivateKey, message)
}

// PublicKeyHex returns the hex-encoded public key (used as peer ID)
func (kp *KeyPair) PublicKeyHex() string {
	return hex.EncodeToString(kp.PublicKey)
}

// PublicKeyShort returns shortened public key for display
func (kp *KeyPair) PublicKeyShort() string {
	hex := kp.PublicKeyHex()
	if len(hex) > 8 {
		return hex[:8]
	}
	return hex
}

// Verify verifies a signature with a public key
func Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// Hash256 computes SHA-256 hash of data
func Hash256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Hash256Hex returns hex-encoded SHA-256 hash
func Hash256Hex(data []byte) string {
	return hex.EncodeToString(Hash256(data))
}
