package auth

import (
	"bytes"
	"testing"
)

func TestEncryptAES256GCM(t *testing.T) {
	// Test data
	plaintext := []byte("This is a test message for AES-256-GCM encryption")

	// Encrypt the data
	encryptedData, err := EncryptAES256GCM(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Validate the encrypted data structure
	if err := ValidateEncryptedData(encryptedData); err != nil {
		t.Fatalf("Encrypted data validation failed: %v", err)
	}

	// Verify key length (32 bytes for AES-256)
	if len(encryptedData.Key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(encryptedData.Key))
	}

	// Verify nonce length (12 bytes for GCM)
	if len(encryptedData.Nonce) != 12 {
		t.Errorf("Expected nonce length 12, got %d", len(encryptedData.Nonce))
	}

	// Verify auth tag length (16 bytes for GCM)
	if len(encryptedData.AuthTag) != 16 {
		t.Errorf("Expected auth tag length 16, got %d", len(encryptedData.AuthTag))
	}

	// Verify ciphertext is not empty
	if len(encryptedData.Ciphertext) == 0 {
		t.Error("Ciphertext should not be empty")
	}

	// Decrypt and verify the result matches the original
	decrypted, err := DecryptAES256GCM(encryptedData)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted data does not match original plaintext")
	}
}

func TestEncryptAES256GCM_EmptyData(t *testing.T) {
	// Test with empty data
	plaintext := []byte{}

	encryptedData, err := EncryptAES256GCM(plaintext)
	if err != nil {
		t.Fatalf("Encryption of empty data failed: %v", err)
	}

	// Validate the structure
	if err := ValidateEncryptedData(encryptedData); err != nil {
		t.Fatalf("Encrypted data validation failed: %v", err)
	}

	// Decrypt and verify
	decrypted, err := DecryptAES256GCM(encryptedData)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted empty data does not match original")
	}
}

func TestEncryptAES256GCM_LargeData(t *testing.T) {
	// Test with larger data (simulating compressed email content)
	plaintext := make([]byte, 1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	encryptedData, err := EncryptAES256GCM(plaintext)
	if err != nil {
		t.Fatalf("Encryption of large data failed: %v", err)
	}

	// Validate the structure
	if err := ValidateEncryptedData(encryptedData); err != nil {
		t.Fatalf("Encrypted data validation failed: %v", err)
	}

	// Decrypt and verify
	decrypted, err := DecryptAES256GCM(encryptedData)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted large data does not match original")
	}
}

func TestValidateEncryptedData(t *testing.T) {
	// Test valid data
	validData := &EncryptedData{
		Ciphertext: []byte{1, 2, 3, 4},
		Key:        make([]byte, 32),
		Nonce:      make([]byte, 12),
		AuthTag:    make([]byte, 16),
	}

	if err := ValidateEncryptedData(validData); err != nil {
		t.Errorf("Valid data should not fail validation: %v", err)
	}

	// Test nil data
	if err := ValidateEncryptedData(nil); err == nil {
		t.Error("Nil data should fail validation")
	}

	// Test invalid key length
	invalidKeyData := &EncryptedData{
		Ciphertext: []byte{1, 2, 3, 4},
		Key:        make([]byte, 16), // Wrong length
		Nonce:      make([]byte, 12),
		AuthTag:    make([]byte, 16),
	}

	if err := ValidateEncryptedData(invalidKeyData); err == nil {
		t.Error("Invalid key length should fail validation")
	}

	// Test invalid nonce length
	invalidNonceData := &EncryptedData{
		Ciphertext: []byte{1, 2, 3, 4},
		Key:        make([]byte, 32),
		Nonce:      make([]byte, 8), // Wrong length
		AuthTag:    make([]byte, 16),
	}

	if err := ValidateEncryptedData(invalidNonceData); err == nil {
		t.Error("Invalid nonce length should fail validation")
	}

	// Test invalid auth tag length
	invalidAuthTagData := &EncryptedData{
		Ciphertext: []byte{1, 2, 3, 4},
		Key:        make([]byte, 32),
		Nonce:      make([]byte, 12),
		AuthTag:    make([]byte, 8), // Wrong length
	}

	if err := ValidateEncryptedData(invalidAuthTagData); err == nil {
		t.Error("Invalid auth tag length should fail validation")
	}

}
