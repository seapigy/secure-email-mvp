package e2e

import (
	"testing"
	"time"
)

func TestCryptoProvider_GenerateKeyPair(t *testing.T) {
	config := DefaultCryptoConfig()
	provider, err := NewCryptoProvider(config)
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	tests := []struct {
		name      string
		algorithm string
		wantErr   bool
	}{
		{"Kyber512", "kyber512", false},
		{"Kyber768", "kyber768", false},
		{"Kyber1024", "kyber1024", false},
		{"Dilithium2", "dilithium2", false},
		{"Dilithium3", "dilithium3", false},
		{"Dilithium5", "dilithium5", false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair, err := provider.GenerateKeyPair(tt.algorithm)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateKeyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && keyPair == nil {
				t.Error("GenerateKeyPair() returned nil keyPair")
				return
			}
			if !tt.wantErr {
				if keyPair.Algorithm != tt.algorithm {
					t.Errorf("GenerateKeyPair() algorithm = %v, want %v", keyPair.Algorithm, tt.algorithm)
				}
				if len(keyPair.PublicKey) == 0 {
					t.Error("GenerateKeyPair() returned empty public key")
				}
				if len(keyPair.PrivateKey) == 0 {
					t.Error("GenerateKeyPair() returned empty private key")
				}
				if keyPair.CreatedAt.IsZero() {
					t.Error("GenerateKeyPair() returned zero creation time")
				}
			}
		})
	}
}

func TestCryptoProvider_EncryptDecryptMessage(t *testing.T) {
	config := DefaultCryptoConfig()
	provider, err := NewCryptoProvider(config)
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate KEM key pairs for encryption/decryption
	recipientKEMKeyPair, err := provider.GenerateKeyPair(config.KEMAlgorithm)
	if err != nil {
		t.Fatalf("Failed to generate recipient KEM key pair: %v", err)
	}
	if err != nil {
		t.Fatalf("Failed to generate recipient KEM key pair: %v", err)
	}

	// Generate signature key pair for signing
	senderSigKeyPair, err := provider.GenerateKeyPair(config.SignatureAlgorithm)
	if err != nil {
		t.Fatalf("Failed to generate sender signature key pair: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"Empty", ""},
		{"Short", "Hello, World!"},
		{"Long", "This is a much longer message that should test the encryption and decryption capabilities of the system. It contains multiple sentences and should be properly encrypted and decrypted."},
		{"SpecialChars", "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"Unicode", "Unicode: 🚀🔐📧✨"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext := []byte(tt.plaintext)

			// Encrypt message using KEM keys for encryption and signature key for signing
			envelope, err := provider.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, senderSigKeyPair.PrivateKey)
			if err != nil {
				t.Fatalf("EncryptMessage() error = %v", err)
			}

			// Verify envelope structure
			if envelope.ID == "" {
				t.Error("Envelope ID is empty")
			}
			if envelope.Version != "1.0" {
				t.Errorf("Envelope version = %v, want 1.0", envelope.Version)
			}
			if envelope.KEMAlgorithm != config.KEMAlgorithm {
				t.Errorf("Envelope KEM algorithm = %v, want %v", envelope.KEMAlgorithm, config.KEMAlgorithm)
			}
			if envelope.DEMAlgorithm != config.DEMAlgorithm {
				t.Errorf("Envelope DEM algorithm = %v, want %v", envelope.DEMAlgorithm, config.DEMAlgorithm)
			}
			if envelope.SignatureAlgorithm != config.SignatureAlgorithm {
				t.Errorf("Envelope signature algorithm = %v, want %v", envelope.SignatureAlgorithm, config.SignatureAlgorithm)
			}
			if envelope.EncryptedKey == "" {
				t.Error("Envelope encrypted key is empty")
			}
			if envelope.EncryptedData == "" {
				t.Error("Envelope encrypted data is empty")
			}
			if envelope.Signature == "" {
				t.Error("Envelope signature is empty")
			}
			if envelope.KeyRotationID == "" {
				t.Error("Envelope key rotation ID is empty")
			}
			if envelope.CreatedAt.IsZero() {
				t.Error("Envelope creation time is zero")
			}
			if envelope.ExpiresAt == nil {
				t.Error("Envelope expiry time is nil")
			}

			// Decrypt message using KEM keys for decryption and signature key for verification
			decrypted, err := provider.DecryptMessage(envelope, recipientKEMKeyPair.PrivateKey, senderSigKeyPair.PublicKey)
			if err != nil {
				t.Fatalf("DecryptMessage() error = %v", err)
			}

			// Verify decrypted content
			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypted content = %v, want %v", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestCryptoProvider_EncryptDecryptWithDifferentAlgorithms(t *testing.T) {
	algorithms := []struct {
		kem string
		dem string
	}{
		{"kyber768", "aes256gcm"},
		{"kyber768", "chacha20poly1305"},
		{"kyber1024", "aes256gcm"},
		{"kyber1024", "chacha20poly1305"},
	}

	for _, alg := range algorithms {
		t.Run(alg.kem+"_"+alg.dem, func(t *testing.T) {
			config := CryptoConfig{
				KEMAlgorithm:       alg.kem,
				DEMAlgorithm:       alg.dem,
				SignatureAlgorithm: "dilithium3",
				KeyRotationDays:    30,
				PerformanceMode:    false,
			}
			provider, err := NewCryptoProvider(config)
			if err != nil {
				t.Fatalf("Failed to create crypto provider: %v", err)
			}

			// Generate KEM key pairs for encryption/decryption
			recipientKEMKeyPair, err := provider.GenerateKeyPair(config.KEMAlgorithm)
			if err != nil {
				t.Fatalf("Failed to generate recipient KEM key pair: %v", err)
			}

			// Generate signature key pairs for signing/verification
			senderSigKeyPair, err := provider.GenerateKeyPair(config.SignatureAlgorithm)
			if err != nil {
				t.Fatalf("Failed to generate sender signature key pair: %v", err)
			}

			plaintext := []byte("Test message for " + alg.kem + " + " + alg.dem)

			// Encrypt message (use KEM public key for encryption, signature private key for signing)
			envelope, err := provider.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, senderSigKeyPair.PrivateKey)
			if err != nil {
				t.Fatalf("EncryptMessage() error = %v", err)
			}

			// Decrypt message (use KEM private key for decryption, signature public key for verification)
			decrypted, err := provider.DecryptMessage(envelope, recipientKEMKeyPair.PrivateKey, senderSigKeyPair.PublicKey)
			if err != nil {
				t.Fatalf("DecryptMessage() error = %v", err)
			}

			// Verify decrypted content
			if string(decrypted) != string(plaintext) {
				t.Errorf("Decrypted content = %v, want %v", string(decrypted), string(plaintext))
			}
		})
	}
}

func TestCryptoProvider_SignatureVerification(t *testing.T) {
	config := DefaultCryptoConfig()
	provider, err := NewCryptoProvider(config)
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate KEM key pair for encryption/decryption
	recipientKEMKeyPair, err := provider.GenerateKeyPair(config.KEMAlgorithm)
	if err != nil {
		t.Fatalf("Failed to generate recipient KEM key pair: %v", err)
	}

	// Generate signature key pair for signing/verification
	senderSigKeyPair, err := provider.GenerateKeyPair(config.SignatureAlgorithm)
	if err != nil {
		t.Fatalf("Failed to generate sender signature key pair: %v", err)
	}

	plaintext := []byte("Test message for signature verification")

	// Encrypt message (use KEM public key for encryption, signature private key for signing)
	envelope, err := provider.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, senderSigKeyPair.PrivateKey)
	if err != nil {
		t.Fatalf("EncryptMessage() error = %v", err)
	}

	// Test valid signature (use signature public key for verification)
	err = provider.verifyEnvelopeSignature(envelope, senderSigKeyPair.PublicKey)
	if err != nil {
		t.Errorf("Valid signature verification failed: %v", err)
	}

	// Test invalid signature (tampered)
	originalSignature := envelope.Signature
	envelope.Signature = "invalid_signature"
	err = provider.verifyEnvelopeSignature(envelope, senderSigKeyPair.PublicKey)
	if err == nil {
		t.Error("Invalid signature verification should have failed")
	}
	envelope.Signature = originalSignature

	// Test invalid signature (wrong public key)
	wrongKeyPair, err := provider.GenerateKeyPair(config.KEMAlgorithm)
	if err != nil {
		t.Fatalf("Failed to generate wrong key pair: %v", err)
	}
	err = provider.verifyEnvelopeSignature(envelope, wrongKeyPair.PublicKey)
	if err == nil {
		t.Error("Signature verification with wrong public key should have failed")
	}
}

func TestCryptoProvider_KeyDerivation(t *testing.T) {
	provider, err := NewCryptoProvider(DefaultCryptoConfig())
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	secret := []byte("test_secret")
	salt := []byte("test_salt")
	info := []byte("test_info")
	length := 32

	// Derive key
	key1, err := provider.DeriveKey(secret, salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if len(key1) != length {
		t.Errorf("Derived key length = %v, want %v", len(key1), length)
	}

	// Derive key again with same parameters (should be deterministic)
	key2, err := provider.DeriveKey(secret, salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("Key derivation is not deterministic")
	}

	// Derive key with different salt (should be different)
	key3, err := provider.DeriveKey(secret, []byte("different_salt"), info, length)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if string(key1) == string(key3) {
		t.Error("Key derivation with different salt should produce different keys")
	}
}

func TestCryptoProvider_EnvelopeExpiry(t *testing.T) {
	provider, err := NewCryptoProvider(DefaultCryptoConfig())
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate KEM key pair for encryption/decryption
	recipientKEMKeyPair, err := provider.GenerateKeyPair("kyber768")
	if err != nil {
		t.Fatalf("Failed to generate recipient KEM key pair: %v", err)
	}

	// Generate signature key pair for signing/verification
	senderSigKeyPair, err := provider.GenerateKeyPair("dilithium3")
	if err != nil {
		t.Fatalf("Failed to generate sender signature key pair: %v", err)
	}

	plaintext := []byte("Test message")

	// Encrypt message (use KEM public key for encryption, signature private key for signing)
	envelope, err := provider.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, senderSigKeyPair.PrivateKey)
	if err != nil {
		t.Fatalf("EncryptMessage() error = %v", err)
	}

	// Check expiry time
	if envelope.ExpiresAt == nil {
		t.Error("Envelope expiry time is nil")
	}

	// Expiry should be in the future
	if envelope.ExpiresAt.Before(time.Now()) {
		t.Error("Envelope expiry time is in the past")
	}

	// Expiry should be approximately 30 days from now
	expectedExpiry := time.Now().AddDate(0, 0, 30)
	diff := envelope.ExpiresAt.Sub(expectedExpiry)
	if diff < -time.Hour || diff > time.Hour {
		t.Errorf("Envelope expiry time is not approximately 30 days from now: %v", diff)
	}
}

func TestCryptoProvider_KeyExpiry(t *testing.T) {
	config := DefaultCryptoConfig()
	config.KeyRotationDays = 7
	provider, err := NewCryptoProvider(config)
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate key pair
	keyPair, err := provider.GenerateKeyPair("kyber768")
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Check expiry time
	if keyPair.ExpiresAt == nil {
		t.Error("Key expiry time is nil")
	}

	// Expiry should be in the future
	if keyPair.ExpiresAt.Before(time.Now()) {
		t.Error("Key expiry time is in the past")
	}

	// Expiry should be approximately 7 days from now
	expectedExpiry := time.Now().AddDate(0, 0, 7)
	diff := keyPair.ExpiresAt.Sub(expectedExpiry)
	if diff < -time.Hour || diff > time.Hour {
		t.Errorf("Key expiry time is not approximately 7 days from now: %v", diff)
	}
}

func TestCryptoProvider_NoKeyExpiry(t *testing.T) {
	config := DefaultCryptoConfig()
	config.KeyRotationDays = 0
	provider, err := NewCryptoProvider(config)
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate key pair
	keyPair, err := provider.GenerateKeyPair("kyber768")
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Check expiry time (should be nil)
	if keyPair.ExpiresAt != nil {
		t.Error("Key expiry time should be nil when KeyRotationDays is 0")
	}
}

func TestCryptoProvider_EnvelopeIDGeneration(t *testing.T) {
	provider, err := NewCryptoProvider(DefaultCryptoConfig())
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate multiple envelope IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := provider.generateEnvelopeID()
		if ids[id] {
			t.Errorf("Duplicate envelope ID generated: %v", id)
		}
		ids[id] = true

		// Check format
		if len(id) < 20 {
			t.Errorf("Envelope ID too short: %v", id)
		}
		if id[:4] != "env_" {
			t.Errorf("Envelope ID doesn't start with 'env_': %v", id)
		}
	}
}

func TestCryptoProvider_KeyRotationIDGeneration(t *testing.T) {
	provider, err := NewCryptoProvider(DefaultCryptoConfig())
	if err != nil {
		t.Fatalf("Failed to create crypto provider: %v", err)
	}

	// Generate multiple key rotation IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := provider.generateKeyRotationID()
		if ids[id] {
			t.Errorf("Duplicate key rotation ID generated: %v", id)
		}
		ids[id] = true

		// Check format
		if len(id) < 12 {
			t.Errorf("Key rotation ID too short: %v", id)
		}
		if id[:4] != "rot_" {
			t.Errorf("Key rotation ID doesn't start with 'rot_': %v", id)
		}
	}
}

// Helper function to create a default crypto config for testing
func DefaultCryptoConfig() CryptoConfig {
	return CryptoConfig{
		PQCImplementation:  "circl", // Use real PQC implementation
		KEMAlgorithm:       "kyber768",
		DEMAlgorithm:       "aes256gcm",
		SignatureAlgorithm: "dilithium3",
		KeyRotationDays:    30,
		PerformanceMode:    false,
	}
}
