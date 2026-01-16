package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Hash computes SHA-256 hash of data
func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashHex computes SHA-256 hash and returns hex string
func HashHex(data []byte) string {
	return hex.EncodeToString(Hash(data))
}

// MerkleRoot computes simple merkle root from hashes
func MerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return Hash([]byte{})
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	// Simple merkle tree: pair up hashes and hash them
	for len(hashes) > 1 {
		var next [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				next = append(next, Hash(combined))
			} else {
				next = append(next, hashes[i])
			}
		}
		hashes = next
	}

	return hashes[0]
}

// Timestamp returns current Unix timestamp in milliseconds
func Timestamp() int64 {
	return time.Now().UnixMilli()
}

// FormatTimestamp formats Unix milliseconds to readable string
func FormatTimestamp(ms int64) string {
	return time.UnixMilli(ms).Format(time.RFC3339)
}

// TruncateString truncates string to max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return s[:maxLen-3] + "..."
}

// MustPanic panics with formatted message if err != nil
func MustPanic(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}
