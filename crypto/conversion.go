package crypto

import (
	"crypto/ed25519"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// PrivateKeyToLibp2p converts an Ed25519 private key to libp2p format
func PrivateKeyToLibp2p(kp *KeyPair) (crypto.PrivKey, error) {
	if len(kp.PrivateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	libp2pPriv, err := crypto.UnmarshalPrivateKey(append([]byte{0x01}, kp.PrivateKey[:]...))
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key: %w", err)
	}

	return libp2pPriv, nil
}

// GenerateLibp2pKeyPair generates a new Ed25519 keypair in libp2p format
func GenerateLibp2pKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	priv, pub, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate libp2p keypair: %w", err)
	}

	return priv, pub, nil
}

// LibKeyPairToEd25519 converts libp2p keypair to Ed25519 KeyPair
func LibKeyPairToEd25519(priv crypto.PrivKey, pub crypto.PubKey) (*KeyPair, error) {
	privBytes, err := priv.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key bytes: %w", err)
	}

	pubBytes, err := pub.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key bytes: %w", err)
	}

	// Ed25519 private key is 64 bytes (32 seed + 32 public), but libp2p Raw() returns just the seed
	fullPriv := ed25519.NewKeyFromSeed(privBytes)

	return &KeyPair{
		PublicKey:  ed25519.PublicKey(pubBytes),
		PrivateKey: fullPriv,
	}, nil
}
