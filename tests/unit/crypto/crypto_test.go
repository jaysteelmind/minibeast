package crypto_test

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"

	"github.com/minibeast/usb-agent/src/core/crypto"
)

// TestGenerateKeyPair verifies key pair generation
func TestGenerateKeyPair(t *testing.T) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	if keyPair == nil {
		t.Fatal("GenerateKeyPair() returned nil")
	}

	if len(keyPair.PublicKey) != crypto.PublicKeySize {
		t.Errorf("PublicKey size = %d, want %d", len(keyPair.PublicKey), crypto.PublicKeySize)
	}

	if len(keyPair.PrivateKey) != crypto.PrivateKeySize {
		t.Errorf("PrivateKey size = %d, want %d", len(keyPair.PrivateKey), crypto.PrivateKeySize)
	}
}

// TestSaveLoadPrivateKey verifies private key persistence
func TestSaveLoadPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	// Generate key
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	// Save key
	if err := crypto.SavePrivateKey(keyPair.PrivateKey, keyPath); err != nil {
		t.Fatalf("SavePrivateKey() failed: %v", err)
	}

	// Verify file exists with correct permissions
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Key file not created: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Key permissions = %o, want 0600", info.Mode().Perm())
	}

	// Load key
	loadedKey, err := crypto.LoadPrivateKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey() failed: %v", err)
	}

	// Verify keys match
	if !ed25519.PrivateKey(keyPair.PrivateKey).Equal(loadedKey) {
		t.Error("Loaded key does not match original")
	}
}

// TestSaveLoadPublicKey verifies public key persistence
func TestSaveLoadPublicKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.pub")

	// Generate key
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	// Save key
	if err := crypto.SavePublicKey(keyPair.PublicKey, keyPath); err != nil {
		t.Fatalf("SavePublicKey() failed: %v", err)
	}

	// Load key
	loadedKey, err := crypto.LoadPublicKey(keyPath)
	if err != nil {
		t.Fatalf("LoadPublicKey() failed: %v", err)
	}

	// Verify keys match
	if !ed25519.PublicKey(keyPair.PublicKey).Equal(loadedKey) {
		t.Error("Loaded key does not match original")
	}
}

// TestSignVerify verifies signature creation and verification
func TestSignVerify(t *testing.T) {
	// Generate key pair
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	// Create signer
	signer := crypto.NewSigner(keyPair)

	// Test data
	testData := []byte("Hello, MiniBeast!")

	// Sign data
	signature, err := signer.Sign(testData)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	if len(signature) != crypto.SignatureSize {
		t.Errorf("Signature size = %d, want %d", len(signature), crypto.SignatureSize)
	}

	// Verify signature
	if !crypto.Verify(keyPair.PublicKey, testData, signature) {
		t.Error("Verify() failed for valid signature")
	}
}

// TestVerify_InvalidSignature verifies rejection of invalid signatures
func TestVerify_InvalidSignature(t *testing.T) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	testData := []byte("Hello, MiniBeast!")
	invalidSig := make([]byte, crypto.SignatureSize)

	if crypto.Verify(keyPair.PublicKey, testData, invalidSig) {
		t.Error("Verify() accepted invalid signature")
	}
}

// TestVerify_ModifiedData verifies detection of data tampering
func TestVerify_ModifiedData(t *testing.T) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	signer := crypto.NewSigner(keyPair)
	originalData := []byte("Hello, MiniBeast!")
	modifiedData := []byte("Hello, MiniBeast?")

	signature, err := signer.Sign(originalData)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	// Signature should not verify with modified data
	if crypto.Verify(keyPair.PublicKey, modifiedData, signature) {
		t.Error("Verify() accepted signature for modified data")
	}
}

// TestSignFile verifies file signing
func TestSignFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test file
	testData := []byte("Test file content")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Generate key and sign
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	signer := crypto.NewSigner(keyPair)
	signature, err := signer.SignFile(testFile)
	if err != nil {
		t.Fatalf("SignFile() failed: %v", err)
	}

	// Verify signature
	if !crypto.Verify(keyPair.PublicKey, testData, signature) {
		t.Error("File signature verification failed")
	}
}

// TestVerifyFile verifies file signature verification
func TestVerifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test file
	testData := []byte("Test file content")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Generate key and sign
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	signer := crypto.NewSigner(keyPair)
	signature, err := signer.SignFile(testFile)
	if err != nil {
		t.Fatalf("SignFile() failed: %v", err)
	}

	// Verify file signature
	valid, err := crypto.VerifyFile(keyPair.PublicKey, testFile, signature)
	if err != nil {
		t.Fatalf("VerifyFile() failed: %v", err)
	}
	if !valid {
		t.Error("VerifyFile() returned false for valid signature")
	}
}

// TestSaveLoadSignature verifies signature persistence
func TestSaveLoadSignature(t *testing.T) {
	tmpDir := t.TempDir()
	sigPath := filepath.Join(tmpDir, "test.sig")

	// Generate key and create signature
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() failed: %v", err)
	}

	signer := crypto.NewSigner(keyPair)
	testData := []byte("Test data")
	signature, err := signer.Sign(testData)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	// Save signature
	if err := crypto.SaveSignature(signature, sigPath); err != nil {
		t.Fatalf("SaveSignature() failed: %v", err)
	}

	// Load signature
	loadedSig, err := crypto.LoadSignature(sigPath)
	if err != nil {
		t.Fatalf("LoadSignature() failed: %v", err)
	}

	// Verify signatures match
	if len(loadedSig) != len(signature) {
		t.Errorf("Signature size mismatch: got %d, want %d", len(loadedSig), len(signature))
	}

	for i := range signature {
		if signature[i] != loadedSig[i] {
			t.Error("Loaded signature does not match original")
			break
		}
	}

	// Verify loaded signature still works
	if !crypto.Verify(keyPair.PublicKey, testData, loadedSig) {
		t.Error("Loaded signature failed verification")
	}
}

// TestSign_NoPrivateKey verifies error when signing without key
func TestSign_NoPrivateKey(t *testing.T) {
	signer := crypto.NewSigner(&crypto.KeyPair{})
	_, err := signer.Sign([]byte("test"))
	if err == nil {
		t.Error("Sign() should fail without private key")
	}
}

// TestLoadPrivateKey_InvalidFile verifies error handling for invalid files
func TestLoadPrivateKey_InvalidFile(t *testing.T) {
	_, err := crypto.LoadPrivateKey("/nonexistent/key")
	if err == nil {
		t.Error("LoadPrivateKey() should fail for nonexistent file")
	}
}

// TestLoadPrivateKey_InvalidPEM verifies error handling for invalid PEM
func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid.key")

	// Write invalid PEM
	if err := os.WriteFile(keyPath, []byte("not a PEM file"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := crypto.LoadPrivateKey(keyPath)
	if err == nil {
		t.Error("LoadPrivateKey() should fail for invalid PEM")
	}
}

// TestSaveSignature_InvalidSize verifies error for wrong signature size
func TestSaveSignature_InvalidSize(t *testing.T) {
	tmpDir := t.TempDir()
	sigPath := filepath.Join(tmpDir, "invalid.sig")

	invalidSig := crypto.Signature([]byte("too short"))
	err := crypto.SaveSignature(invalidSig, sigPath)
	if err == nil {
		t.Error("SaveSignature() should fail for invalid signature size")
	}
}
