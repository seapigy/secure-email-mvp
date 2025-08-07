package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptedData represents the result of AES-256-GCM encryption
// with all components separated for secure storage. This structure allows
// for flexible storage strategies where encryption components can be stored
// separately (e.g., key in secure key management, ciphertext in cloud storage).
type EncryptedData struct {
	Ciphertext []byte // The encrypted data (without authentication tag)
	Key        []byte // The encryption key (32 bytes for AES-256)
	Nonce      []byte // The nonce (12 bytes for AES-GCM)
	AuthTag    []byte // The GCM authentication tag (16 bytes)
}

// EncryptAES256GCM encrypts data using AES-256-GCM with a random key and nonce.
// This function implements a secure encryption scheme where:
// 1. A cryptographically secure random 32-byte key is generated
// 2. A random 12-byte nonce is generated for each encryption
// 3. The data is encrypted using AES-256-GCM mode
// 4. The authentication tag is extracted and stored separately
// 5. All components are returned for secure storage
//
// Returns the encrypted data with all components separated for secure storage.
// The caller is responsible for securely storing the key and nonce.
func EncryptAES256GCM(plaintext []byte) (*EncryptedData, error) {
	// Generate a cryptographically secure random 32-byte encryption key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create AES-256 cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode for authenticated encryption
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate a cryptographically secure random nonce
	// GCM requires a unique nonce for each encryption operation
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data using AES-256-GCM
	// This provides both confidentiality and authenticity
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Extract the authentication tag (last 16 bytes of ciphertext)
	// The auth tag provides integrity verification during decryption
	authTag := ciphertext[len(ciphertext)-16:]
	// Remove the auth tag from ciphertext for separate storage
	ciphertextOnly := ciphertext[:len(ciphertext)-16]

	return &EncryptedData{
		Ciphertext: ciphertextOnly,
		Key:        key,
		Nonce:      nonce,
		AuthTag:    authTag,
	}, nil
}

// DecryptAES256GCM decrypts data using AES-256-GCM with the provided components.
// This function reconstructs the original plaintext from the separated encryption
// components. It performs integrity verification using the authentication tag.
//
// Security considerations:
// - The function validates the authentication tag to detect tampering
// - If the auth tag doesn't match, the function returns an error
// - The nonce must match the one used during encryption
// - The key must be the same one used during encryption
func DecryptAES256GCM(encryptedData *EncryptedData) ([]byte, error) {
	// Validate the encrypted data structure before processing
	if err := ValidateEncryptedData(encryptedData); err != nil {
		return nil, fmt.Errorf("invalid encrypted data: %w", err)
	}

	// Create AES-256 cipher block using the stored key
	block, err := aes.NewCipher(encryptedData.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode for authenticated decryption
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Reconstruct the full ciphertext by combining ciphertext and auth tag
	// This is necessary because GCM.Open() expects the auth tag to be appended
	fullCiphertext := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

	// Decrypt and verify the data using AES-256-GCM
	// This will fail if the auth tag doesn't match (indicating tampering)
	plaintext, err := aesgcm.Open(nil, encryptedData.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data (possible tampering): %w", err)
	}

	return plaintext, nil
}

// ValidateEncryptedData validates that the encrypted data structure is properly formed
// and contains all required components with correct lengths. This function is called
// before decryption to ensure the data structure is valid.
//
// Validation checks:
// - Data structure is not nil
// - Key is exactly 32 bytes (AES-256 requirement)
// - Nonce is exactly 12 bytes (AES-GCM requirement)
// - Authentication tag is exactly 16 bytes (AES-GCM requirement)
// - Ciphertext can be empty (valid for AES-GCM)
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
	// as it will still produce a valid auth tag for empty plaintext
	return nil
}
