package crypto

import (
	"crypto/ed25519"
)

// KeyPair represents an Ed25519 key pair
// Mathematical security: 2^128 computational hardness
type KeyPair struct {
	PublicKey  ed25519.PublicKey  // 32 bytes
	PrivateKey ed25519.PrivateKey // 64 bytes (seed + public key)
}

// Signature represents a 64-byte Ed25519 signature
type Signature []byte

// SignatureSize is the byte length of Ed25519 signatures
const SignatureSize = ed25519.SignatureSize // 64 bytes

// PublicKeySize is the byte length of Ed25519 public keys
const PublicKeySize = ed25519.PublicKeySize // 32 bytes

// PrivateKeySize is the byte length of Ed25519 private keys
const PrivateKeySize = ed25519.PrivateKeySize // 64 bytes
