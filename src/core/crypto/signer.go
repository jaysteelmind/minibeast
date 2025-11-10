package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"os"
)

// Signer provides cryptographic signing operations
type Signer struct {
	keyPair *KeyPair
}

// NewSigner creates a new signer with the given key pair
// Complexity: O(1)
func NewSigner(keyPair *KeyPair) *Signer {
	return &Signer{keyPair: keyPair}
}

// Sign creates an Ed25519 signature over the SHA-256 hash of data
// Mathematical specification:
//  1. Hash: h = SHA256(data)
//  2. Sign: signature = Ed25519.Sign(privateKey, h)
//
// Security: 2^128 computational hardness (collision resistance: 2^256)
// Complexity: O(n) where n = len(data)
func (s *Signer) Sign(data []byte) (Signature, error) {
	if s.keyPair == nil || s.keyPair.PrivateKey == nil {
		return nil, fmt.Errorf("no private key available")
	}

	// Step 1: Hash the data with SHA-256
	hash := sha256.Sum256(data)

	// Step 2: Sign the hash with Ed25519
	signature := ed25519.Sign(s.keyPair.PrivateKey, hash[:])

	return Signature(signature), nil
}

// SignFile signs the contents of a file
// Complexity: O(n) where n = file size
func (s *Signer) SignFile(filePath string) (Signature, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return s.Sign(data)
}

// Verify checks an Ed25519 signature against data
// Mathematical specification:
//  1. Hash: h = SHA256(data)
//  2. Verify: Ed25519.Verify(publicKey, h, signature) = true/false
//
// Complexity: O(n) where n = len(data)
func Verify(publicKey ed25519.PublicKey, data []byte, signature Signature) bool {
	if len(signature) != SignatureSize {
		return false
	}

	// Step 1: Hash the data
	hash := sha256.Sum256(data)

	// Step 2: Verify signature
	return ed25519.Verify(publicKey, hash[:], signature)
}

// VerifyFile verifies a signature against file contents
// Complexity: O(n) where n = file size
func VerifyFile(publicKey ed25519.PublicKey, filePath string, signature Signature) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	return Verify(publicKey, data, signature), nil
}

// SaveSignature writes signature to binary file
// Complexity: O(1)
func SaveSignature(signature Signature, path string) error {
	if len(signature) != SignatureSize {
		return fmt.Errorf("invalid signature size: %d bytes", len(signature))
	}

	// Atomic write pattern
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, signature, 0644); err != nil {
		return fmt.Errorf("failed to write temp signature: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename signature: %w", err)
	}

	return nil
}

// LoadSignature reads signature from binary file
// Complexity: O(1)
func LoadSignature(path string) (Signature, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature file: %w", err)
	}

	if len(data) != SignatureSize {
		return nil, fmt.Errorf("invalid signature size: %d bytes", len(data))
	}

	return Signature(data), nil
}
