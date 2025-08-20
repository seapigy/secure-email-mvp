package e2e

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// CryptoProvider handles all cryptographic operations for the E2E system
type CryptoProvider struct {
	config CryptoConfig
}

// Envelope represents the encrypted message envelope
type Envelope struct {
	ID                 string            `json:"id"`
	Version            string            `json:"version"`
	KEMAlgorithm       string            `json:"kem_algorithm"`
	DEMAlgorithm       string            `json:"dem_algorithm"`
	SignatureAlgorithm string            `json:"signature_algorithm"`
	EncryptedKey       string            `json:"encrypted_key"`
	EncryptedData      string            `json:"encrypted_data"`
	Signature          string            `json:"signature"`
	KeyRotationID      string            `json:"key_rotation_id"`
	CreatedAt          time.Time         `json:"created_at"`
	ExpiresAt          *time.Time        `json:"expires_at,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	PublicKey  []byte     `json:"public_key"`
	PrivateKey []byte     `json:"private_key"`
	Algorithm  string     `json:"algorithm"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// NewCryptoProvider creates a new cryptographic provider
func NewCryptoProvider(config CryptoConfig) *CryptoProvider {
	return &CryptoProvider{
		config: config,
	}
}

// GenerateKeyPair generates a new key pair for the specified algorithm
func (cp *CryptoProvider) GenerateKeyPair(algorithm string) (*KeyPair, error) {
	switch algorithm {
	case "kyber512", "kyber768", "kyber1024":
		return cp.generateKyberKeyPair(algorithm)
	case "dilithium2", "dilithium3", "dilithium5":
		return cp.generateDilithiumKeyPair(algorithm)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// generateKyberKeyPair generates a Kyber key pair (placeholder implementation)
func (cp *CryptoProvider) generateKyberKeyPair(algorithm string) (*KeyPair, error) {
	// TODO: Implement actual Kyber key generation
	// For now, generate random keys as placeholders with a deterministic relationship
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Derive public key from private key for placeholder consistency
	publicKey := make([]byte, 32)
	for i := range privateKey {
		publicKey[i] = privateKey[i] ^ 0xAA // Simple XOR transformation for placeholder
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Algorithm:  algorithm,
		CreatedAt:  time.Now(),
		ExpiresAt:  cp.calculateKeyExpiry(),
	}, nil
}

// generateDilithiumKeyPair generates a Dilithium key pair (placeholder implementation)
func (cp *CryptoProvider) generateDilithiumKeyPair(algorithm string) (*KeyPair, error) {
	// TODO: Implement actual Dilithium key generation
	// For now, generate random keys as placeholders with a deterministic relationship
	privateKey := make([]byte, 128)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Derive public key from private key for placeholder consistency
	publicKey := make([]byte, 64)
	for i := range publicKey {
		if i < len(privateKey) {
			publicKey[i] = privateKey[i] ^ 0xBB // Simple XOR transformation for placeholder
		}
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Algorithm:  algorithm,
		CreatedAt:  time.Now(),
		ExpiresAt:  cp.calculateKeyExpiry(),
	}, nil
}

// EncryptMessage encrypts a message using the E2E envelope format
func (cp *CryptoProvider) EncryptMessage(plaintext []byte, recipientPublicKey []byte, senderPrivateKey []byte) (*Envelope, error) {
	// Generate a random symmetric key for DEM
	symmetricKey := make([]byte, 32)
	if _, err := rand.Read(symmetricKey); err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	// Encrypt the symmetric key using KEM (Key Encapsulation)
	encryptedKey, err := cp.encapsulateKey(symmetricKey, recipientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encapsulate key: %w", err)
	}

	// Encrypt the plaintext using DEM (Data Encapsulation)
	encryptedData, err := cp.encryptData(plaintext, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Create the envelope
	envelope := &Envelope{
		ID:                 cp.generateEnvelopeID(),
		Version:            "1.0",
		KEMAlgorithm:       cp.config.KEMAlgorithm,
		DEMAlgorithm:       cp.config.DEMAlgorithm,
		SignatureAlgorithm: cp.config.SignatureAlgorithm,
		EncryptedKey:       base64.StdEncoding.EncodeToString(encryptedKey),
		EncryptedData:      base64.StdEncoding.EncodeToString(encryptedData),
		KeyRotationID:      cp.generateKeyRotationID(),
		CreatedAt:          time.Now(),
		ExpiresAt:          cp.calculateEnvelopeExpiry(),
		Metadata:           make(map[string]string),
	}

	// Sign the envelope
	signature, err := cp.signEnvelope(envelope, senderPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign envelope: %w", err)
	}
	envelope.Signature = base64.StdEncoding.EncodeToString(signature)

	return envelope, nil
}

// DecryptMessage decrypts a message from the E2E envelope format
func (cp *CryptoProvider) DecryptMessage(envelope *Envelope, recipientPrivateKey []byte, senderPublicKey []byte) ([]byte, error) {
	// Verify the signature
	if err := cp.verifyEnvelopeSignature(envelope, senderPublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Decode the encrypted key and data
	encryptedKey, err := base64.StdEncoding.DecodeString(envelope.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(envelope.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Decapsulate the symmetric key
	symmetricKey, err := cp.decapsulateKey(encryptedKey, recipientPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decapsulate key: %w", err)
	}

	// Decrypt the data
	plaintext, err := cp.decryptData(encryptedData, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// encapsulateKey encapsulates a symmetric key using KEM
func (cp *CryptoProvider) encapsulateKey(symmetricKey []byte, publicKey []byte) ([]byte, error) {
	// TODO: Implement actual KEM encapsulation
	// For now, use a simple XOR-based placeholder
	encapsulated := make([]byte, len(symmetricKey))
	for i := range symmetricKey {
		encapsulated[i] = symmetricKey[i] ^ publicKey[i%len(publicKey)]
	}
	return encapsulated, nil
}

// decapsulateKey decapsulates a symmetric key using KEM
func (cp *CryptoProvider) decapsulateKey(encapsulatedKey []byte, privateKey []byte) ([]byte, error) {
	// TODO: Implement actual KEM decapsulation
	// For now, use a simple XOR-based placeholder

	// The private key should be the same length as the public key used for encapsulation
	// For the placeholder, we need to derive the public key from the private key to match encapsulation
	publicKey := make([]byte, len(privateKey))
	for i := range privateKey {
		if len(privateKey) == 32 {
			// Kyber key transformation
			publicKey[i] = privateKey[i] ^ 0xAA
		} else {
			// Dilithium key transformation
			publicKey[i] = privateKey[i] ^ 0xBB
		}
	}

	decapsulated := make([]byte, len(encapsulatedKey))
	for i := range encapsulatedKey {
		decapsulated[i] = encapsulatedKey[i] ^ publicKey[i%len(publicKey)]
	}
	return decapsulated, nil
}

// encryptData encrypts data using DEM
func (cp *CryptoProvider) encryptData(plaintext []byte, key []byte) ([]byte, error) {
	switch cp.config.DEMAlgorithm {
	case "aes256gcm":
		return cp.encryptAES256GCM(plaintext, key)
	case "chacha20poly1305":
		return cp.encryptChaCha20Poly1305(plaintext, key)
	default:
		return nil, fmt.Errorf("unsupported DEM algorithm: %s", cp.config.DEMAlgorithm)
	}
}

// decryptData decrypts data using DEM
func (cp *CryptoProvider) decryptData(ciphertext []byte, key []byte) ([]byte, error) {
	switch cp.config.DEMAlgorithm {
	case "aes256gcm":
		return cp.decryptAES256GCM(ciphertext, key)
	case "chacha20poly1305":
		return cp.decryptChaCha20Poly1305(ciphertext, key)
	default:
		return nil, fmt.Errorf("unsupported DEM algorithm: %s", cp.config.DEMAlgorithm)
	}
}

// encryptAES256GCM encrypts data using AES-256-GCM
func (cp *CryptoProvider) encryptAES256GCM(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES256GCM decrypts data using AES-256-GCM
func (cp *CryptoProvider) decryptAES256GCM(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20Poly1305 encrypts data using ChaCha20-Poly1305
func (cp *CryptoProvider) encryptChaCha20Poly1305(plaintext []byte, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptChaCha20Poly1305 decrypts data using ChaCha20-Poly1305
func (cp *CryptoProvider) decryptChaCha20Poly1305(ciphertext []byte, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// signEnvelope signs an envelope
func (cp *CryptoProvider) signEnvelope(envelope *Envelope, privateKey []byte) ([]byte, error) {
	// Create a canonical representation of the envelope for signing
	signatureData := cp.createSignatureData(envelope)

	// TODO: Implement actual signature using Dilithium
	// For now, use a simple HMAC-based placeholder
	signature := cp.createHMACSignature(signatureData, privateKey)
	return signature, nil
}

// verifyEnvelopeSignature verifies an envelope signature
func (cp *CryptoProvider) verifyEnvelopeSignature(envelope *Envelope, publicKey []byte) error {
	signature, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	signatureData := cp.createSignatureData(envelope)

	// For the placeholder implementation, derive the signing key from the public key
	// This reverses the transformation used in key generation
	signingKey := make([]byte, len(publicKey))
	for i := range publicKey {
		if len(publicKey) == 32 {
			// Kyber key transformation
			signingKey[i] = publicKey[i] ^ 0xAA
		} else {
			// Dilithium key transformation
			signingKey[i] = publicKey[i] ^ 0xBB
		}
	}

	expectedSignature := cp.createHMACSignature(signatureData, signingKey)

	// Debug: Print signature lengths and first few bytes for debugging
	if len(signature) != len(expectedSignature) {
		return fmt.Errorf("signature verification failed: length mismatch (got %d, expected %d)", len(signature), len(expectedSignature))
	}

	if !cp.constantTimeCompare(signature, expectedSignature) {
		return fmt.Errorf("signature verification failed: signature mismatch")
	}

	return nil
}

// createSignatureData creates canonical data for signing
func (cp *CryptoProvider) createSignatureData(envelope *Envelope) []byte {
	// Create a canonical representation excluding the signature field and time-based fields
	// for deterministic signing in the placeholder implementation
	tempEnvelope := *envelope
	tempEnvelope.Signature = ""
	tempEnvelope.CreatedAt = time.Time{} // Zero time for deterministic signing
	tempEnvelope.ExpiresAt = nil         // Exclude expiry for deterministic signing

	data, _ := json.Marshal(tempEnvelope)
	return data
}

// createHMACSignature creates a simple HMAC-based signature (placeholder)
func (cp *CryptoProvider) createHMACSignature(data []byte, key []byte) []byte {
	// TODO: Replace with actual Dilithium signature
	// For now, use a simple hash-based approach
	h := sha256.New()
	h.Write(data)
	h.Write(key)
	return h.Sum(nil)
}

// constantTimeCompare performs constant-time comparison
func (cp *CryptoProvider) constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// generateEnvelopeID generates a unique envelope ID
func (cp *CryptoProvider) generateEnvelopeID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("env_%x", id)
}

// generateKeyRotationID generates a key rotation identifier
func (cp *CryptoProvider) generateKeyRotationID() string {
	id := make([]byte, 8)
	rand.Read(id)
	return fmt.Sprintf("rot_%x", id)
}

// calculateKeyExpiry calculates when a key should expire
func (cp *CryptoProvider) calculateKeyExpiry() *time.Time {
	if cp.config.KeyRotationDays <= 0 {
		return nil
	}
	expiry := time.Now().AddDate(0, 0, cp.config.KeyRotationDays)
	return &expiry
}

// calculateEnvelopeExpiry calculates when an envelope should expire
func (cp *CryptoProvider) calculateEnvelopeExpiry() *time.Time {
	// Default envelope expiry: 30 days
	expiry := time.Now().AddDate(0, 0, 30)
	return &expiry
}

// DeriveKey derives a key using HKDF
func (cp *CryptoProvider) DeriveKey(secret []byte, salt []byte, info []byte, length int) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	_, err := io.ReadFull(hkdf, key)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	return key, nil
}

// DecryptThreadMessage decrypts a thread message without signature verification
func (cp *CryptoProvider) DecryptThreadMessage(envelope *Envelope, threadKey []byte) ([]byte, error) {
	// Decode the encrypted key and data
	encryptedKey, err := base64.StdEncoding.DecodeString(envelope.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(envelope.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// For thread messages, the thread key is used directly for decapsulation
	// This is symmetric encryption, so the same key is used for both encapsulation and decapsulation
	symmetricKey := make([]byte, len(encryptedKey))
	for i := range encryptedKey {
		symmetricKey[i] = encryptedKey[i] ^ threadKey[i%len(threadKey)]
	}

	// Decrypt the data
	plaintext, err := cp.decryptData(encryptedData, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
