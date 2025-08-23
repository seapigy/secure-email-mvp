package pqc

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"
)

// BenchmarkPQCEncryption benchmarks PQC hybrid encryption performance
func BenchmarkPQCEncryption(b *testing.B) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		b.Fatalf("Failed to create PQC service: %v", err)
	}

	// Test different data sizes
	testSizes := []int{100, 500, 1000, 5000, 10000, 50000}

	for _, size := range testSizes {
		// Generate test data
		testData := make([]byte, size)
		rand.Read(testData)

		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := service.EncryptHybrid(testData, "benchmark")
				if err != nil {
					b.Fatalf("Encryption failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPQCDecryption benchmarks PQC hybrid decryption performance
func BenchmarkPQCDecryption(b *testing.B) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		b.Fatalf("Failed to create PQC service: %v", err)
	}

	// Test different data sizes
	testSizes := []int{100, 500, 1000, 5000, 10000, 50000}

	for _, size := range testSizes {
		// Generate test data and encrypt it once
		testData := make([]byte, size)
		rand.Read(testData)

		encryptedData, err := service.EncryptHybrid(testData, "benchmark")
		if err != nil {
			b.Fatalf("Failed to encrypt test data: %v", err)
		}

		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := service.DecryptHybrid(encryptedData, "benchmark")
				if err != nil {
					b.Fatalf("Decryption failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPQCSerialization benchmarks PQC data serialization/deserialization
func BenchmarkPQCSerialization(b *testing.B) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		b.Fatalf("Failed to create PQC service: %v", err)
	}

	// Generate test data
	testData := make([]byte, 1000)
	rand.Read(testData)

	// Encrypt data once
	encryptedData, err := service.EncryptHybrid(testData, "benchmark")
	if err != nil {
		b.Fatalf("Failed to encrypt test data: %v", err)
	}

	b.Run("Serialize", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.SerializeHybridData(encryptedData)
			if err != nil {
				b.Fatalf("Serialization failed: %v", err)
			}
		}
	})

	// Serialize once for deserialization benchmark
	serialized, err := service.SerializeHybridData(encryptedData)
	if err != nil {
		b.Fatalf("Failed to serialize test data: %v", err)
	}

	b.Run("Deserialize", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.DeserializeHybridData(serialized)
			if err != nil {
				b.Fatalf("Deserialization failed: %v", err)
			}
		}
	})
}

// BenchmarkPQCKeyGeneration benchmarks PQC key generation performance
func BenchmarkPQCKeyGeneration(b *testing.B) {
	b.Run("KeyGeneration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create a new key manager for each generation to simulate key generation
			newConfig := DefaultPQCConfig()
			newConfig.EnablePQC = true
			_, err := NewKeyManager(newConfig)
			if err != nil {
				b.Fatalf("Key generation failed: %v", err)
			}
		}
	})
}

// BenchmarkPQCKeyEncapsulation benchmarks PQC key encapsulation performance
func BenchmarkPQCKeyEncapsulation(b *testing.B) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		b.Fatalf("Failed to create PQC service: %v", err)
	}

	keyManager := service.GetKeyManager()

	// Generate symmetric key
	symmetricKey := make([]byte, 32)
	rand.Read(symmetricKey)

	b.Run("KeyEncapsulation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := keyManager.EncapsulateKey(symmetricKey)
			if err != nil {
				b.Fatalf("Key encapsulation failed: %v", err)
			}
		}
	})
}

// BenchmarkPQCKeyDecapsulation benchmarks PQC key decapsulation performance
func BenchmarkPQCKeyDecapsulation(b *testing.B) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		b.Fatalf("Failed to create PQC service: %v", err)
	}

	keyManager := service.GetKeyManager()

	// Generate symmetric key and encapsulate it once
	symmetricKey := make([]byte, 32)
	rand.Read(symmetricKey)

	encapsulatedKey, err := keyManager.EncapsulateKey(symmetricKey)
	if err != nil {
		b.Fatalf("Failed to encapsulate key: %v", err)
	}

	b.Run("KeyDecapsulation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := keyManager.DecapsulateKey(encapsulatedKey)
			if err != nil {
				b.Fatalf("Key decapsulation failed: %v", err)
			}
		}
	})
}

// TestPQCPerformanceMetrics runs performance tests and outputs metrics
func TestPQCPerformanceMetrics(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	fmt.Println("🔐 PQC Performance Metrics")
	fmt.Println("==========================")

	// Test different data sizes
	testSizes := []int{100, 500, 1000, 5000, 10000, 50000}

	for _, size := range testSizes {
		// Generate test data
		testData := make([]byte, size)
		rand.Read(testData)

		// Measure encryption time
		startTime := time.Now()
		encryptedData, err := service.EncryptHybrid(testData, "performance_test")
		encryptTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Encryption failed for size %d: %v", size, err)
		}

		// Measure decryption time
		startTime = time.Now()
		_, err = service.DecryptHybrid(encryptedData, "performance_test")
		decryptTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("Decryption failed for size %d: %v", size, err)
		}

		// Calculate throughput
		encryptThroughput := float64(size) / encryptTime.Seconds()
		decryptThroughput := float64(size) / decryptTime.Seconds()

		fmt.Printf("Size %5d bytes: Encrypt %6dms (%8.0f bytes/sec), Decrypt %6dms (%8.0f bytes/sec)\n",
			size,
			encryptTime.Milliseconds(),
			encryptThroughput,
			decryptTime.Milliseconds(),
			decryptThroughput,
		)
	}

	// Test key generation performance
	keyManager := service.GetKeyManager()

	startTime := time.Now()
	newConfig := DefaultPQCConfig()
	newConfig.EnablePQC = true
	_, err = NewKeyManager(newConfig)
	keyGenTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Key generation failed: %v", err)
	}

	fmt.Printf("\nKey Generation: %6dms\n", keyGenTime.Milliseconds())

	// Test key encapsulation/decapsulation
	symmetricKey := make([]byte, 32)
	rand.Read(symmetricKey)

	startTime = time.Now()
	encapsulatedKey, err := keyManager.EncapsulateKey(symmetricKey)
	encapsulateTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Key encapsulation failed: %v", err)
	}

	startTime = time.Now()
	_, err = keyManager.DecapsulateKey(encapsulatedKey)
	decapsulateTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Key decapsulation failed: %v", err)
	}

	fmt.Printf("Key Encapsulation: %6dms\n", encapsulateTime.Milliseconds())
	fmt.Printf("Key Decapsulation: %6dms\n", decapsulateTime.Milliseconds())

	// Test serialization performance with a sample encrypted data
	testData := make([]byte, 1000)
	rand.Read(testData)
	encryptedData, err := service.EncryptHybrid(testData, "performance_test")
	if err != nil {
		t.Fatalf("Failed to encrypt test data for serialization: %v", err)
	}

	startTime = time.Now()
	serialized, err := service.SerializeHybridData(encryptedData)
	serializeTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	startTime = time.Now()
	_, err = service.DeserializeHybridData(serialized)
	deserializeTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	fmt.Printf("Serialization: %6dms\n", serializeTime.Milliseconds())
	fmt.Printf("Deserialization: %6dms\n", deserializeTime.Milliseconds())

	fmt.Println("\n✅ PQC Performance Metrics Completed")
}

// TestPQCStressTest runs stress tests with multiple concurrent operations
func TestPQCStressTest(t *testing.T) {
	config := DefaultPQCConfig()
	config.EnablePQC = true

	service, err := NewPQCService(config)
	if err != nil {
		t.Fatalf("Failed to create PQC service: %v", err)
	}

	fmt.Println("🚀 PQC Stress Test")
	fmt.Println("==================")

	// Test concurrent encryption/decryption
	numGoroutines := 10
	numOperations := 100

	startTime := time.Now()

	// Create channels for coordination
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				// Generate test data
				testData := make([]byte, 1000)
				rand.Read(testData)

				// Encrypt
				encryptedData, err := service.EncryptHybrid(testData, fmt.Sprintf("stress_test_%d_%d", goroutineID, j))
				if err != nil {
					t.Errorf("Encryption failed in goroutine %d, operation %d: %v", goroutineID, j, err)
					return
				}

				// Decrypt
				_, err = service.DecryptHybrid(encryptedData, fmt.Sprintf("stress_test_%d_%d", goroutineID, j))
				if err != nil {
					t.Errorf("Decryption failed in goroutine %d, operation %d: %v", goroutineID, j, err)
					return
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	totalTime := time.Since(startTime)
	totalOperations := numGoroutines * numOperations

	fmt.Printf("Completed %d operations in %6dms\n", totalOperations, totalTime.Milliseconds())
	fmt.Printf("Throughput: %6.2f operations/second\n", float64(totalOperations)/totalTime.Seconds())
	fmt.Printf("Average time per operation: %6dms\n", totalTime.Milliseconds()/int64(totalOperations))

	fmt.Println("✅ PQC Stress Test Completed")
}
