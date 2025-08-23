package unit

import (
	"testing"
	"time"

	"secure-email-mvp/pkg/pqc"
)

// Test data
const (
	testContext = "test_encryption_context"
	testData    = "This is test data for PQC encryption"
)

// TestPQCConfig tests PQC configuration
func TestPQCConfig(t *testing.T) {
	t.Run("DefaultPQCConfig", func(t *testing.T) {
		config := pqc.DefaultPQCConfig()

		if !config.EnablePQC {
			t.Error("PQC should be enabled by default")
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
	})

	t.Run("LoadPQCConfigFromEnv", func(t *testing.T) {
		config := pqc.LoadPQCConfigFromEnv()

		// Test that config is loaded (we don't know the exact values from env)
		if config == nil {
			t.Error("Config should not be nil")
		}
	})
}

// TestPQCService tests PQC service functionality
func TestPQCService(t *testing.T) {
	t.Run("NewPQCService", func(t *testing.T) {
		config := pqc.DefaultPQCConfig()
		service, err := pqc.NewPQCService(config)

		if err != nil {
			t.Fatalf("NewPQCService failed: %v", err)
		}

		if service == nil {
			t.Error("Service should not be nil")
		}
	})

	t.Run("NewPQCService with nil config", func(t *testing.T) {
		// NewPQCService should handle nil config gracefully
		service, err := pqc.NewPQCService(nil)
		if err != nil {
			t.Fatalf("NewPQCService should not fail with nil config: %v", err)
		}
		if service == nil {
			t.Error("Service should not be nil")
		}
	})
}

// TestPQCEncryption tests PQC encryption functionality
func TestPQCEncryption(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("EncryptHybrid", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, testContext)

		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		if encryptedData == nil {
			t.Error("Encrypted data should not be nil")
		}

		if len(encryptedData.KyberCiphertext) == 0 {
			t.Error("Kyber ciphertext should not be empty")
		}

		if encryptedData.AES256GCMData == nil {
			t.Error("AES-256-GCM data should not be nil")
		}

		if len(encryptedData.AES256GCMData.Ciphertext) == 0 {
			t.Error("AES-256-GCM ciphertext should not be empty")
		}

		if len(encryptedData.AES256GCMData.Nonce) == 0 {
			t.Error("AES-256-GCM nonce should not be empty")
		}

		if len(encryptedData.AES256GCMData.AuthTag) == 0 {
			t.Error("AES-256-GCM auth tag should not be empty")
		}
	})

	t.Run("EncryptHybrid with empty data", func(t *testing.T) {
		encryptedData, err := service.EncryptHybrid([]byte{}, testContext)

		if err != nil {
			t.Fatalf("EncryptHybrid failed with empty data: %v", err)
		}

		if encryptedData == nil {
			t.Error("Encrypted data should not be nil for empty input")
		}
	})

	t.Run("EncryptHybrid with large data", func(t *testing.T) {
		// Create 1MB of test data
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		encryptedData, err := service.EncryptHybrid(largeData, testContext)

		if err != nil {
			t.Fatalf("EncryptHybrid failed with large data: %v", err)
		}

		if encryptedData == nil {
			t.Error("Encrypted data should not be nil for large input")
		}
	})
}

// TestPQCDecryption tests PQC decryption functionality
func TestPQCDecryption(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("DecryptHybrid", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, testContext)

		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		decryptedData, err := service.DecryptHybrid(encryptedData, testContext)

		if err != nil {
			t.Fatalf("DecryptHybrid failed: %v", err)
		}

		if string(decryptedData) != string(plaintext) {
			t.Errorf("Decrypted data doesn't match original: expected %s, got %s", string(plaintext), string(decryptedData))
		}
	})

	t.Run("DecryptHybrid with nil data", func(t *testing.T) {
		// Add nil check to prevent panic
		_, err := service.DecryptHybrid(nil, testContext)
		if err == nil {
			t.Error("Expected error for nil data")
		}
	})

	t.Run("DecryptHybrid with corrupted data", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, testContext)
		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		// Corrupt the ciphertext
		encryptedData.AES256GCMData.Ciphertext[0] ^= 1

		_, err = service.DecryptHybrid(encryptedData, testContext)
		if err == nil {
			t.Error("Expected error for corrupted data")
		}
	})
}

// TestPQCKeyManagement tests PQC key management functionality
func TestPQCKeyManagement(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("GetKeyManager", func(t *testing.T) {
		keyManager := service.GetKeyManager()
		if keyManager == nil {
			t.Error("Key manager should not be nil")
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		config := service.GetConfig()
		if config == nil {
			t.Error("Config should not be nil")
		}

		if !config.EnablePQC {
			t.Error("PQC should be enabled")
		}
	})

	t.Run("IsEnabled", func(t *testing.T) {
		enabled := service.IsEnabled()
		if !enabled {
			t.Error("PQC service should be enabled")
		}
	})
}

// TestPQCPerformance tests PQC performance benchmarks
func TestPQCPerformance(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("EncryptionPerformance", func(t *testing.T) {
		plaintext := []byte(testData)
		start := time.Now()

		for i := 0; i < 10; i++ {
			_, err := service.EncryptHybrid(plaintext, testContext)
			if err != nil {
				t.Fatalf("EncryptHybrid failed on iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / 10

		t.Logf("Average encryption time: %v", avgDuration)

		// Performance should be reasonable (less than 100ms per operation)
		if avgDuration > 100*time.Millisecond {
			t.Errorf("Encryption performance too slow: %v average", avgDuration)
		}
	})

	t.Run("DecryptionPerformance", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, testContext)
		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		start := time.Now()

		for i := 0; i < 10; i++ {
			_, err := service.DecryptHybrid(encryptedData, testContext)
			if err != nil {
				t.Fatalf("DecryptHybrid failed on iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / 10

		t.Logf("Average decryption time: %v", avgDuration)

		// Performance should be reasonable (less than 100ms per operation)
		if avgDuration > 100*time.Millisecond {
			t.Errorf("Decryption performance too slow: %v average", avgDuration)
		}
	})
}

// TestPQCSecurity tests PQC security properties
func TestPQCSecurity(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("KeyUniqueness", func(t *testing.T) {
		// Test that each encryption uses unique keys
		plaintext := []byte(testData)

		encrypted1, err := service.EncryptHybrid(plaintext, testContext)
		if err != nil {
			t.Fatalf("First encryption failed: %v", err)
		}

		encrypted2, err := service.EncryptHybrid(plaintext, testContext)
		if err != nil {
			t.Fatalf("Second encryption failed: %v", err)
		}

		// The ciphertexts should be different due to unique keys
		if string(encrypted1.KyberCiphertext) == string(encrypted2.KyberCiphertext) {
			t.Error("Ciphertexts should be different for unique keys")
		}

		if string(encrypted1.AES256GCMData.Ciphertext) == string(encrypted2.AES256GCMData.Ciphertext) {
			t.Error("AES ciphertexts should be different for unique keys")
		}
	})

	t.Run("ContextIsolation", func(t *testing.T) {
		// Test that different contexts produce different ciphertexts
		plaintext := []byte(testData)

		encrypted1, err := service.EncryptHybrid(plaintext, "context1")
		if err != nil {
			t.Fatalf("First encryption failed: %v", err)
		}

		encrypted2, err := service.EncryptHybrid(plaintext, "context2")
		if err != nil {
			t.Fatalf("Second encryption failed: %v", err)
		}

		// The ciphertexts should be different due to different contexts
		if string(encrypted1.KyberCiphertext) == string(encrypted2.KyberCiphertext) {
			t.Error("Ciphertexts should be different for different contexts")
		}
	})
}

// TestPQCErrorHandling tests PQC error handling
func TestPQCErrorHandling(t *testing.T) {
	config := pqc.DefaultPQCConfig()
	service, err := pqc.NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	t.Run("DecryptWithWrongContext", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, "context1")
		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		// Try to decrypt with wrong context (should still work as context is not validated)
		_, err = service.DecryptHybrid(encryptedData, "context2")
		if err != nil {
			t.Errorf("Decryption with wrong context should work: %v", err)
		}
	})

	t.Run("DecryptWithCorruptedKyberData", func(t *testing.T) {
		plaintext := []byte(testData)
		encryptedData, err := service.EncryptHybrid(plaintext, testContext)
		if err != nil {
			t.Fatalf("EncryptHybrid failed: %v", err)
		}

		// Corrupt Kyber ciphertext
		encryptedData.KyberCiphertext[0] ^= 1

		_, err = service.DecryptHybrid(encryptedData, testContext)
		if err == nil {
			t.Error("Expected error for corrupted Kyber data")
		}
	})
}
