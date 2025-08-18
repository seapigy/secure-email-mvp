package pqc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

// PQCConfig holds configuration for PQC operations
type PQCConfig struct {
	EnablePQC       bool `json:"enable_pqc"`
	KyberLevel      int  `json:"kyber_level"`       // 512, 768, or 1024
	HybridMode      bool `json:"hybrid_mode"`       // Use hybrid classical + PQC
	KeyRotationDays int  `json:"key_rotation_days"` // Key rotation interval
	HSMEnabled      bool `json:"hsm_enabled"`       // Use HSM for key operations
	PerformanceMode bool `json:"performance_mode"`  // Optimize for performance
	AuditLogging    bool `json:"audit_logging"`     // Enable detailed audit logging
}

// DefaultPQCConfig returns the default PQC configuration
func DefaultPQCConfig() *PQCConfig {
	return &PQCConfig{
		EnablePQC:       false, // Disabled by default, controlled by feature flag
		KyberLevel:      768,   // Kyber-768 (recommended security level)
		HybridMode:      true,  // Use hybrid classical + PQC
		KeyRotationDays: 30,    // Rotate keys every 30 days
		HSMEnabled:      false, // HSM disabled by default
		PerformanceMode: false, // Security over performance by default
		AuditLogging:    true,  // Enable audit logging
	}
}

// LoadPQCConfigFromEnv loads PQC configuration from environment variables
func LoadPQCConfigFromEnv() *PQCConfig {
	config := DefaultPQCConfig()

	// Check if PQC is enabled via feature flag
	if os.Getenv("ENABLE_PQC_LAYER") == "true" {
		config.EnablePQC = true
	}

	// Load other configuration values
	if hsmEnabled := os.Getenv("PQC_HSM_ENABLED"); hsmEnabled == "true" {
		config.HSMEnabled = true
	}

	if performanceMode := os.Getenv("PQC_PERFORMANCE_MODE"); performanceMode == "true" {
		config.PerformanceMode = true
	}

	return config
}

// HybridEncryptedData represents the hybrid PQC + symmetric encryption result
type HybridEncryptedData struct {
	// PQC Components
	KyberCiphertext []byte `json:"kyber_ciphertext"` // Kyber encapsulated key
	KyberLevel      int    `json:"kyber_level"`      // Kyber security level used

	// Symmetric Components (dual encryption)
	AES256GCMData *SymmetricEncryptedData `json:"aes256gcm_data"` // AES-256-GCM encrypted data
	ChaCha20Data  *SymmetricEncryptedData `json:"chacha20_data"`  // ChaCha20-Poly1305 encrypted data

	// Metadata
	EncryptionTime time.Time `json:"encryption_time"`
	HybridMode     bool      `json:"hybrid_mode"`
	KeyID          string    `json:"key_id"`  // HSM key identifier
	Version        string    `json:"version"` // PQC implementation version
}

// SymmetricEncryptedData represents AES-256-GCM or ChaCha20-Poly1305 encrypted data
type SymmetricEncryptedData struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	AuthTag    []byte `json:"auth_tag"`
	Algorithm  string `json:"algorithm"` // "AES-256-GCM" or "ChaCha20-Poly1305"
}

// PQCService provides post-quantum cryptography operations
type PQCService struct {
	config     *PQCConfig
	keyManager *KeyManager
	auditLog   *AuditLogger
	mu         sync.RWMutex
}

// NewPQCService creates a new PQC service instance
func NewPQCService(config *PQCConfig) (*PQCService, error) {
	if config == nil {
		config = DefaultPQCConfig()
	}

	keyManager, err := NewKeyManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create key manager: %w", err)
	}

	auditLog := NewAuditLogger(config.AuditLogging)

	service := &PQCService{
		config:     config,
		keyManager: keyManager,
		auditLog:   auditLog,
	}

	// Log service initialization
	service.auditLog.LogEvent("PQC_SERVICE_INIT", "Service initialized", map[string]interface{}{
		"kyber_level":      config.KyberLevel,
		"hybrid_mode":      config.HybridMode,
		"hsm_enabled":      config.HSMEnabled,
		"performance_mode": config.PerformanceMode,
	})

	return service, nil
}

// EncryptHybrid encrypts data using hybrid PQC + symmetric encryption
func (s *PQCService) EncryptHybrid(plaintext []byte, context string) (*HybridEncryptedData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.EnablePQC {
		return nil, fmt.Errorf("PQC layer is disabled")
	}

	startTime := time.Now()

	// Generate a random symmetric key for this encryption
	symmetricKey := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	// Encapsulate the symmetric key using Kyber
	kyberCiphertext, err := s.keyManager.EncapsulateKey(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encapsulate key with Kyber: %w", err)
	}

	// Encrypt data with AES-256-GCM
	aesData, err := s.encryptAES256GCM(plaintext, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with AES-256-GCM: %w", err)
	}

	// Encrypt data with ChaCha20-Poly1305
	chachaData, err := s.encryptChaCha20(plaintext, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with ChaCha20-Poly1305: %w", err)
	}

	// Create hybrid encrypted data
	hybridData := &HybridEncryptedData{
		KyberCiphertext: kyberCiphertext,
		KyberLevel:      s.config.KyberLevel,
		AES256GCMData:   aesData,
		ChaCha20Data:    chachaData,
		EncryptionTime:  time.Now(),
		HybridMode:      s.config.HybridMode,
		KeyID:           s.keyManager.GetCurrentKeyID(),
		Version:         "1.0.0",
	}

	// Log encryption event
	s.auditLog.LogEvent("HYBRID_ENCRYPT", "Data encrypted with hybrid PQC", map[string]interface{}{
		"context":         context,
		"plaintext_size":  len(plaintext),
		"kyber_level":     s.config.KyberLevel,
		"encryption_time": time.Since(startTime).Milliseconds(),
		"key_id":          hybridData.KeyID,
	})

	return hybridData, nil
}

// DecryptHybrid decrypts data using hybrid PQC + symmetric decryption
func (s *PQCService) DecryptHybrid(hybridData *HybridEncryptedData, context string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.EnablePQC {
		return nil, fmt.Errorf("PQC layer is disabled")
	}

	startTime := time.Now()

	// Decapsulate the symmetric key using Kyber
	symmetricKey, err := s.keyManager.DecapsulateKey(hybridData.KyberCiphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decapsulate key with Kyber: %w", err)
	}

	// Try to decrypt with AES-256-GCM first (primary method)
	plaintext, err := s.decryptAES256GCM(hybridData.AES256GCMData, symmetricKey)
	if err == nil {
		// Log successful decryption
		s.auditLog.LogEvent("HYBRID_DECRYPT", "Data decrypted with AES-256-GCM", map[string]interface{}{
			"context":         context,
			"plaintext_size":  len(plaintext),
			"decryption_time": time.Since(startTime).Milliseconds(),
			"key_id":          hybridData.KeyID,
			"method":          "AES-256-GCM",
		})
		return plaintext, nil
	}

	// Fallback to ChaCha20-Poly1305 if AES-256-GCM fails
	plaintext, err = s.decryptChaCha20(hybridData.ChaCha20Data, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("both AES-256-GCM and ChaCha20-Poly1305 decryption failed: %w", err)
	}

	// Log successful fallback decryption
	s.auditLog.LogEvent("HYBRID_DECRYPT", "Data decrypted with ChaCha20-Poly1305 (fallback)", map[string]interface{}{
		"context":         context,
		"plaintext_size":  len(plaintext),
		"decryption_time": time.Since(startTime).Milliseconds(),
		"key_id":          hybridData.KeyID,
		"method":          "ChaCha20-Poly1305",
		"fallback":        true,
	})

	return plaintext, nil
}

// encryptAES256GCM encrypts data using AES-256-GCM
func (s *PQCService) encryptAES256GCM(plaintext, key []byte) (*SymmetricEncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertextOnly := ciphertext[:len(ciphertext)-16]

	return &SymmetricEncryptedData{
		Ciphertext: ciphertextOnly,
		Nonce:      nonce,
		AuthTag:    authTag,
		Algorithm:  "AES-256-GCM",
	}, nil
}

// decryptAES256GCM decrypts data using AES-256-GCM
func (s *PQCService) decryptAES256GCM(data *SymmetricEncryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// For AES-256-GCM, the ciphertext should already include the auth tag
	// Combine ciphertext and auth tag (nonce is separate)
	ciphertext := append(data.Ciphertext, data.AuthTag...)

	plaintext, err := aesGCM.Open(nil, data.Nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with AES-256-GCM: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20 encrypts data using ChaCha20-Poly1305
func (s *PQCService) encryptChaCha20(plaintext, key []byte) (*SymmetricEncryptedData, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertextOnly := ciphertext[:len(ciphertext)-16]

	return &SymmetricEncryptedData{
		Ciphertext: ciphertextOnly,
		Nonce:      nonce,
		AuthTag:    authTag,
		Algorithm:  "ChaCha20-Poly1305",
	}, nil
}

// decryptChaCha20 decrypts data using ChaCha20-Poly1305
func (s *PQCService) decryptChaCha20(data *SymmetricEncryptedData, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	// For ChaCha20-Poly1305, the ciphertext should already include the auth tag
	// Combine ciphertext and auth tag (nonce is separate)
	ciphertext := append(data.Ciphertext, data.AuthTag...)

	plaintext, err := aead.Open(nil, data.Nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with ChaCha20-Poly1305: %w", err)
	}

	return plaintext, nil
}

// SerializeHybridData serializes hybrid encrypted data to JSON
func (s *PQCService) SerializeHybridData(data *HybridEncryptedData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal hybrid data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// DeserializeHybridData deserializes hybrid encrypted data from JSON
func (s *PQCService) DeserializeHybridData(serialized string) (*HybridEncryptedData, error) {
	jsonData, err := base64.StdEncoding.DecodeString(serialized)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var data HybridEncryptedData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hybrid data: %w", err)
	}

	return &data, nil
}

// GetConfig returns the current PQC configuration
func (s *PQCService) GetConfig() *PQCConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig updates the PQC configuration
func (s *PQCService) UpdateConfig(newConfig *PQCConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Log configuration change
	s.auditLog.LogEvent("PQC_CONFIG_UPDATE", "Configuration updated", map[string]interface{}{
		"old_kyber_level": s.config.KyberLevel,
		"new_kyber_level": newConfig.KyberLevel,
		"old_hybrid_mode": s.config.HybridMode,
		"new_hybrid_mode": newConfig.HybridMode,
		"old_hsm_enabled": s.config.HSMEnabled,
		"new_hsm_enabled": newConfig.HSMEnabled,
	})

	s.config = newConfig
	return nil
}

// IsEnabled returns whether PQC is currently enabled
func (s *PQCService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnablePQC
}

// GetStats returns PQC service statistics
func (s *PQCService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"enabled":           s.config.EnablePQC,
		"kyber_level":       s.config.KyberLevel,
		"hybrid_mode":       s.config.HybridMode,
		"hsm_enabled":       s.config.HSMEnabled,
		"performance_mode":  s.config.PerformanceMode,
		"audit_logging":     s.config.AuditLogging,
		"key_rotation_days": s.config.KeyRotationDays,
	}
}

// GetKeyManager returns the key manager instance
func (s *PQCService) GetKeyManager() *KeyManager {
	return s.keyManager
}
