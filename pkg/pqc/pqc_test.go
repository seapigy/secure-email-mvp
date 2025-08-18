package pqc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestPQCConfig(t *testing.T) {
	// Test default configuration
	config := DefaultPQCConfig()

	if config.EnablePQC {
		t.Error("PQC should be disabled by default")
	}

	if config.KyberLevel != 768 {
		t.Errorf("Expected Kyber level 768, got %d", config.KyberLevel)
	}

	if !config.HybridMode {
		t.Error("Hybrid mode should be enabled by default")
	}

	if config.KeyRotationDays != 30 {
		t.Errorf("Expected key rotation days 30, got %d", config.KeyRotationDays)
	}

	if config.HSMEnabled {
		t.Error("HSM should be disabled by default")
	}

	if config.PerformanceMode {
		t.Error("Performance mode should be disabled by default")
	}

	if !config.AuditLogging {
		t.Error("Audit logging should be enabled by default")
	}
}

func TestPQCServiceCreation(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true
	config.HSMEnabled = false // Disable HSM for testing to avoid delays

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	if service == nil {
		t.Fatal("PQC service should not be nil")
	}

	if !service.IsEnabled() {
		t.Error("PQC service should be enabled")
	}

	stats := service.GetStats()
	if stats["enabled"] != true {
		t.Error("Service stats should show enabled")
	}
}

func TestPQCEncryptionDecryption(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true
	config.HSMEnabled = false // Disable HSM for testing to avoid delays

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	// Test data
	plaintext := []byte("This is a test message for PQC encryption")
	context := "test_context"

	// Encrypt
	hybridData, err := service.EncryptHybrid(plaintext, context)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if hybridData == nil {
		t.Fatal("Encrypted data should not be nil")
	}

	if len(hybridData.KyberCiphertext) == 0 {
		t.Error("Kyber ciphertext should not be empty")
	}

	if hybridData.AES256GCMData == nil {
		t.Error("AES-256-GCM data should not be nil")
	}

	if hybridData.ChaCha20Data == nil {
		t.Error("ChaCha20 data should not be nil")
	}

	if hybridData.KyberLevel != 768 {
		t.Errorf("Expected Kyber level 768, got %d", hybridData.KyberLevel)
	}

	// Decrypt
	decrypted, err := service.DecryptHybrid(hybridData, context)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted data should match original plaintext")
	}
}

func TestPQCServiceSerialization(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true
	config.HSMEnabled = false // Disable HSM for testing to avoid delays

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	// Create test data
	plaintext := []byte("Test message for serialization")
	hybridData, err := service.EncryptHybrid(plaintext, "test")
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Serialize
	serialized, err := service.SerializeHybridData(hybridData)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if serialized == "" {
		t.Error("Serialized data should not be empty")
	}

	// Deserialize
	deserialized, err := service.DeserializeHybridData(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if deserialized == nil {
		t.Fatal("Deserialized data should not be nil")
	}

	// Verify deserialized data matches original
	if !bytes.Equal(hybridData.KyberCiphertext, deserialized.KyberCiphertext) {
		t.Error("Deserialized Kyber ciphertext should match original")
	}

	if hybridData.KyberLevel != deserialized.KyberLevel {
		t.Error("Deserialized Kyber level should match original")
	}

	if hybridData.HybridMode != deserialized.HybridMode {
		t.Error("Deserialized hybrid mode should match original")
	}

	// Test decryption with deserialized data
	decrypted, err := service.DecryptHybrid(deserialized, "test")
	if err != nil {
		t.Fatalf("Failed to decrypt deserialized data: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted deserialized data should match original plaintext")
	}
}

func TestKeyManager(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true
	config.HSMEnabled = false   // Disable HSM for testing to avoid delays
	config.KeyRotationDays = 30 // Set reasonable expiration for testing

	keyManager, err := NewKeyManager(config)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	// Test key generation
	currentKeyID := keyManager.GetCurrentKeyID()
	if currentKeyID == "" {
		t.Error("Current key ID should not be empty")
	}

	// Test public key export
	publicKey, err := keyManager.GetCurrentPublicKey()
	if err != nil {
		t.Fatalf("Failed to get public key: %v", err)
	}

	if len(publicKey) == 0 {
		t.Error("Public key should not be empty")
	}

	// Test key encapsulation and decapsulation
	symmetricKey := []byte("test-symmetric-key-32-bytes-long")

	encapsulated, err := keyManager.EncapsulateKey(symmetricKey)
	if err != nil {
		t.Fatalf("Failed to encapsulate key: %v", err)
	}

	if len(encapsulated) == 0 {
		t.Error("Encapsulated key should not be empty")
	}

	decapsulated, err := keyManager.DecapsulateKey(encapsulated)
	if err != nil {
		t.Fatalf("Failed to decapsulate key: %v", err)
	}

	if !bytes.Equal(symmetricKey, decapsulated) {
		t.Error("Decapsulated key should match original symmetric key")
	}

	// Test key stats
	stats := keyManager.GetKeyStats()
	if stats["total_keys"] == nil {
		t.Error("Key stats should contain total_keys")
	}

	if stats["active_keys"] == nil {
		t.Error("Key stats should contain active_keys")
	}
}

func runKeyRotationTestLogic(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true
	config.HSMEnabled = false   // Disable HSM for testing to avoid delays
	config.KeyRotationDays = 30 // Set reasonable expiration for testing

	keyManager, err := NewKeyManager(config)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	// Get initial key ID
	initialKeyID := keyManager.GetCurrentKeyID()
	if initialKeyID == "" {
		t.Fatal("Initial key ID should not be empty")
	}

	// Test basic key operations before rotation
	symmetricKey := []byte("test-symmetric-key-32-bytes-long")
	encapsulated, err := keyManager.EncapsulateKey(symmetricKey)
	if err != nil {
		t.Fatalf("Failed to encapsulate key before rotation: %v", err)
	}

	decapsulated, err := keyManager.DecapsulateKey(encapsulated)
	if err != nil {
		t.Fatalf("Failed to decapsulate key before rotation: %v", err)
	}

	if !bytes.Equal(symmetricKey, decapsulated) {
		t.Error("Decapsulated key should match original before rotation")
	}

	// Rotate keys
	err = keyManager.RotateKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Get new key ID
	newKeyID := keyManager.GetCurrentKeyID()
	if newKeyID == "" {
		t.Fatal("New key ID should not be empty")
	}

	if initialKeyID == newKeyID {
		t.Error("Key ID should change after rotation")
	}

	// Test that new encapsulated keys work
	newSymmetricKey := []byte("new-symmetric-key-32-bytes-long")
	newEncapsulated, err := keyManager.EncapsulateKey(newSymmetricKey)
	if err != nil {
		t.Fatalf("Failed to encapsulate key after rotation: %v", err)
	}

	newDecapsulated, err := keyManager.DecapsulateKey(newEncapsulated)
	if err != nil {
		t.Fatalf("Failed to decapsulate key after rotation: %v", err)
	}

	if !bytes.Equal(newSymmetricKey, newDecapsulated) {
		t.Error("Decapsulated key should match original after rotation")
	}
}

func TestKeyRotation(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		runKeyRotationTestLogic(t)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("TestKeyRotation timed out after 5 seconds")
	}
}

func TestAuditLogger(t *testing.T) {
	auditLogger := NewAuditLogger(true)

	if !auditLogger.IsEnabled() {
		t.Error("Audit logger should be enabled")
	}

	// Test basic event logging
	auditLogger.LogEvent("TEST_EVENT", "Test event description", map[string]interface{}{
		"test_key": "test_value",
	})

	// Test security event logging
	auditLogger.LogSecurityEvent("TEST_SECURITY", "Test security event", map[string]interface{}{
		"severity": "INFO",
	}, "INFO")

	// Test key operation logging
	auditLogger.LogKeyOperation("GENERATE", "test-key-id", map[string]interface{}{
		"kyber_level": 768,
	})

	// Test encryption operation logging
	auditLogger.LogEncryptionOperation("ENCRYPT", "test_context", 1024, map[string]interface{}{
		"algorithm": "PQC-HYBRID",
	})

	// Test decryption operation logging
	auditLogger.LogDecryptionOperation("DECRYPT", "test_context", 1024, "AES-256-GCM", map[string]interface{}{
		"success": true,
	})

	// Test performance event logging
	auditLogger.LogPerformanceEvent("ENCRYPT", 50*time.Millisecond, map[string]interface{}{
		"data_size": 1024,
	})

	// Test error logging
	auditLogger.LogError("TEST_ERROR", "Test error event", fmt.Errorf("test error"), map[string]interface{}{
		"error_code": "TEST_001",
	})

	// Test warning logging
	auditLogger.LogWarning("TEST_WARNING", "Test warning event", map[string]interface{}{
		"warning_code": "TEST_002",
	})

	// Test critical logging
	auditLogger.LogCritical("TEST_CRITICAL", "Test critical event", map[string]interface{}{
		"critical_code": "TEST_003",
	})

	// Get stats
	stats := auditLogger.GetStats()
	if stats["enabled"] != true {
		t.Error("Audit logger stats should show enabled")
	}

	// Close logger
	err := auditLogger.Close()
	if err != nil {
		t.Errorf("Failed to close audit logger: %v", err)
	}
}

func TestPQCConfigurationUpdate(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	// Test configuration update
	newConfig := DefaultPQCConfig()
	newConfig.EnablePQC = true
	newConfig.KyberLevel = 1024
	newConfig.HSMEnabled = true
	newConfig.PerformanceMode = true

	err = service.UpdateConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to update configuration: %v", err)
	}

	updatedConfig := service.GetConfig()
	if updatedConfig.KyberLevel != 1024 {
		t.Errorf("Expected Kyber level 1024, got %d", updatedConfig.KyberLevel)
	}

	if !updatedConfig.HSMEnabled {
		t.Error("HSM should be enabled after update")
	}

	if !updatedConfig.PerformanceMode {
		t.Error("Performance mode should be enabled after update")
	}
}

func TestPQCDisabledMode(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = false

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	if service.IsEnabled() {
		t.Error("PQC service should be disabled")
	}

	// Test that encryption fails when disabled
	plaintext := []byte("test message")
	_, err = service.EncryptHybrid(plaintext, "test")
	if err == nil {
		t.Error("Encryption should fail when PQC is disabled")
	}

	// Test that decryption fails when disabled
	hybridData := &HybridEncryptedData{}
	_, err = service.DecryptHybrid(hybridData, "test")
	if err == nil {
		t.Error("Decryption should fail when PQC is disabled")
	}
}

func TestPQCJSONSerialization(t *testing.T) {
	// Test that HybridEncryptedData can be marshaled to JSON
	hybridData := &HybridEncryptedData{
		KyberCiphertext: []byte("test-kyber-ciphertext"),
		KyberLevel:      768,
		AES256GCMData: &SymmetricEncryptedData{
			Ciphertext: []byte("test-aes-ciphertext"),
			Nonce:      []byte("test-nonce"),
			AuthTag:    []byte("test-auth-tag"),
			Algorithm:  "AES-256-GCM",
		},
		ChaCha20Data: &SymmetricEncryptedData{
			Ciphertext: []byte("test-chacha-ciphertext"),
			Nonce:      []byte("test-chacha-nonce"),
			AuthTag:    []byte("test-chacha-auth-tag"),
			Algorithm:  "ChaCha20-Poly1305",
		},
		EncryptionTime: time.Now(),
		HybridMode:     true,
		KeyID:          "test-key-id",
		Version:        "1.0.0",
	}

	jsonData, err := json.Marshal(hybridData)
	if err != nil {
		t.Fatalf("Failed to marshal HybridEncryptedData: %v", err)
	}

	var unmarshaled HybridEncryptedData
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal HybridEncryptedData: %v", err)
	}

	if unmarshaled.KyberLevel != hybridData.KyberLevel {
		t.Error("Unmarshaled Kyber level should match original")
	}

	if unmarshaled.HybridMode != hybridData.HybridMode {
		t.Error("Unmarshaled hybrid mode should match original")
	}

	if unmarshaled.KeyID != hybridData.KeyID {
		t.Error("Unmarshaled key ID should match original")
	}

	if unmarshaled.Version != hybridData.Version {
		t.Error("Unmarshaled version should match original")
	}
}
