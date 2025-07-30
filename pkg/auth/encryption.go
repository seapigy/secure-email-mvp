package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptedData represents the result of AES-256-GCM encryption
// with all components separated for secure storage
type EncryptedData struct {
	Ciphertext []byte // The encrypted data
	Key        []byte // The encryption key (32 bytes)
	Nonce      []byte // The nonce (12 bytes)
	AuthTag    []byte // The GCM authentication tag (16 bytes)
}

// EncryptAES256GCM encrypts data using AES-256-GCM with a random key and nonce.
// Returns the encrypted data with all components separated for secure storage.
func EncryptAES256GCM(plaintext []byte) (*EncryptedData, error) {
	// Generate a random 32-byte encryption key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate a random 12-byte nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Extract the authentication tag (last 16 bytes of ciphertext)
	authTag := ciphertext[len(ciphertext)-16:]
	// Remove the auth tag from ciphertext
	ciphertextOnly := ciphertext[:len(ciphertext)-16]

	return &EncryptedData{
		Ciphertext: ciphertextOnly,
		Key:        key,
		Nonce:      nonce,
		AuthTag:    authTag,
	}, nil
}

// DecryptAES256GCM decrypts data using AES-256-GCM with the provided components.
// This function is for completeness but should be used carefully in production.
func DecryptAES256GCM(encryptedData *EncryptedData) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(encryptedData.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Combine ciphertext and auth tag for decryption
	fullCiphertext := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

	// Decrypt the data
	plaintext, err := aesgcm.Open(nil, encryptedData.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// ValidateEncryptedData validates that the encrypted data structure is properly formed
func ValidateEncryptedData(data *EncryptedData) error {
	if data == nil {
		return fmt.Errorf("encrypted data is nil")
	}
	if len(data.Key) != 32 {
		return fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(data.Key))
	}
	if len(data.Nonce) != 12 {
		return fmt.Errorf("invalid nonce length: expected 12 bytes, got %d", len(data.Nonce))
	}
	if len(data.AuthTag) != 16 {
		return fmt.Errorf("invalid auth tag length: expected 16 bytes, got %d", len(data.AuthTag))
	}
	// Note: Empty ciphertext is actually valid for AES-GCM encryption
	// as it will still produce a valid auth tag
	return nil
}
