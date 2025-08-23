package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// BenchmarkSuite provides comprehensive performance benchmarking for E2E operations
type BenchmarkSuite struct {
	CryptoProvider  *CryptoProvider
	Client          *Client
	KeyTransparency *KeyTransparency
	ThresholdHSM    *ThresholdHSM
	Results         []BenchmarkResult
	Config          BenchmarkConfig
	mutex           sync.RWMutex
}

// BenchmarkResult represents the result of a single benchmark operation
type BenchmarkResult struct {
	Operation    string                 `json:"operation"`
	Duration     time.Duration          `json:"duration"`
	Throughput   float64                `json:"throughput"`   // operations per second
	MemoryUsage  int64                  `json:"memory_usage"` // bytes
	CPUUsage     float64                `json:"cpu_usage"`    // percentage
	MessageSize  int64                  `json:"message_size"` // bytes
	Timestamp    time.Time              `json:"timestamp"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BenchmarkConfig configures benchmark parameters
type BenchmarkConfig struct {
	Iterations        int           `json:"iterations"`
	WarmupRuns        int           `json:"warmup_runs"`
	Timeout           time.Duration `json:"timeout"`
	MessageSizes      []int         `json:"message_sizes"` // bytes
	ConcurrencyLevels []int         `json:"concurrency_levels"`
	KEMAlgorithms     []string      `json:"kem_algorithms"`
	DEMAlgorithms     []string      `json:"dem_algorithms"`
	EnableGC          bool          `json:"enable_gc"`
	CollectMemStats   bool          `json:"collect_mem_stats"`
}

// BenchmarkStats provides statistical analysis of benchmark results
type BenchmarkStats struct {
	Mean       time.Duration `json:"mean"`
	Median     time.Duration `json:"median"`
	Min        time.Duration `json:"min"`
	Max        time.Duration `json:"max"`
	StdDev     time.Duration `json:"std_dev"`
	P95        time.Duration `json:"p95"`
	P99        time.Duration `json:"p99"`
	Throughput float64       `json:"throughput"`
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(config BenchmarkConfig) (*BenchmarkSuite, error) {
	cryptoProvider, err := NewCryptoProvider(DefaultE2EConfig().Crypto)
	if err != nil {
		return nil, fmt.Errorf("failed to create crypto provider: %w", err)
	}

	client, err := NewClient(*getTestConfig(), "benchmark_user")
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	ktConfig := DefaultE2EConfig().KeyTransparency
	ktConfig.Enabled = true
	kt := NewKeyTransparency(ktConfig)

	hsmConfig := DefaultE2EConfig().HSM
	hsmConfig.Enabled = true
	hsm := NewThresholdHSM(hsmConfig)

	// Set default config values
	if config.Iterations == 0 {
		config.Iterations = 1000
	}
	if config.WarmupRuns == 0 {
		config.WarmupRuns = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if len(config.MessageSizes) == 0 {
		config.MessageSizes = []int{256, 1024, 4096, 16384, 65536} // 256B to 64KB
	}
	if len(config.ConcurrencyLevels) == 0 {
		config.ConcurrencyLevels = []int{1, 4, 8, 16, 32}
	}
	if len(config.KEMAlgorithms) == 0 {
		config.KEMAlgorithms = []string{"kyber512", "kyber768", "kyber1024"}
	}
	if len(config.DEMAlgorithms) == 0 {
		config.DEMAlgorithms = []string{"aes256gcm", "chacha20poly1305"}
	}

	return &BenchmarkSuite{
		CryptoProvider:  cryptoProvider,
		Client:          client,
		KeyTransparency: kt,
		ThresholdHSM:    hsm,
		Results:         make([]BenchmarkResult, 0),
		Config:          config,
	}, nil
}

// RunAllBenchmarks runs the complete benchmark suite
func (bs *BenchmarkSuite) RunAllBenchmarks(ctx context.Context) error {
	fmt.Println("Starting E2E Performance Benchmark Suite...")

	// Warmup
	fmt.Printf("Running %d warmup iterations...\n", bs.Config.WarmupRuns)
	for i := 0; i < bs.Config.WarmupRuns; i++ {
		bs.benchmarkKeyGeneration(ctx, "kyber768", 1)
	}

	// Core crypto benchmarks
	if err := bs.runCryptoBenchmarks(ctx); err != nil {
		return fmt.Errorf("crypto benchmarks failed: %w", err)
	}

	// E2E message flow benchmarks
	if err := bs.runMessageFlowBenchmarks(ctx); err != nil {
		return fmt.Errorf("message flow benchmarks failed: %w", err)
	}

	// Key management benchmarks
	if err := bs.runKeyManagementBenchmarks(ctx); err != nil {
		return fmt.Errorf("key management benchmarks failed: %w", err)
	}

	// Concurrency benchmarks
	if err := bs.runConcurrencyBenchmarks(ctx); err != nil {
		return fmt.Errorf("concurrency benchmarks failed: %w", err)
	}

	fmt.Printf("Benchmark suite completed. Total results: %d\n", len(bs.Results))
	return nil
}

// runCryptoBenchmarks runs core cryptographic operation benchmarks
func (bs *BenchmarkSuite) runCryptoBenchmarks(ctx context.Context) error {
	fmt.Println("Running cryptographic benchmarks...")

	// Key generation benchmarks
	for _, kemAlg := range bs.Config.KEMAlgorithms {
		if err := bs.benchmarkKeyGeneration(ctx, kemAlg, bs.Config.Iterations); err != nil {
			return err
		}
	}

	// Encryption/decryption benchmarks
	for _, kemAlg := range bs.Config.KEMAlgorithms {
		for _, demAlg := range bs.Config.DEMAlgorithms {
			for _, size := range bs.Config.MessageSizes {
				if err := bs.benchmarkEncryption(ctx, kemAlg, demAlg, size, bs.Config.Iterations); err != nil {
					return err
				}
				if err := bs.benchmarkDecryption(ctx, kemAlg, demAlg, size, bs.Config.Iterations); err != nil {
					return err
				}
			}
		}
	}

	// Signature benchmarks
	for _, kemAlg := range bs.Config.KEMAlgorithms {
		for _, size := range bs.Config.MessageSizes {
			if err := bs.benchmarkSigning(ctx, kemAlg, size, bs.Config.Iterations); err != nil {
				return err
			}
			if err := bs.benchmarkVerification(ctx, kemAlg, size, bs.Config.Iterations); err != nil {
				return err
			}
		}
	}

	return nil
}

// runMessageFlowBenchmarks runs end-to-end message flow benchmarks
func (bs *BenchmarkSuite) runMessageFlowBenchmarks(ctx context.Context) error {
	fmt.Println("Running message flow benchmarks...")

	for _, size := range bs.Config.MessageSizes {
		// E2E message encryption
		if err := bs.benchmarkE2EMessageEncryption(ctx, size, bs.Config.Iterations); err != nil {
			return err
		}

		// E2E message decryption
		if err := bs.benchmarkE2EMessageDecryption(ctx, size, bs.Config.Iterations); err != nil {
			return err
		}

		// Thread message encryption
		if err := bs.benchmarkThreadMessageEncryption(ctx, size, bs.Config.Iterations); err != nil {
			return err
		}

		// Thread message decryption
		if err := bs.benchmarkThreadMessageDecryption(ctx, size, bs.Config.Iterations); err != nil {
			return err
		}
	}

	return nil
}

// runKeyManagementBenchmarks runs key management operation benchmarks
func (bs *BenchmarkSuite) runKeyManagementBenchmarks(ctx context.Context) error {
	fmt.Println("Running key management benchmarks...")

	// Key transparency benchmarks
	if err := bs.benchmarkKeyRegistration(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	if err := bs.benchmarkKeyVerification(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	if err := bs.benchmarkKeyTransparencyAudit(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	// Threshold HSM benchmarks
	if err := bs.benchmarkThresholdSigning(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	if err := bs.benchmarkThresholdVerification(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	// Key rotation benchmarks
	if err := bs.benchmarkKeyRotation(ctx, bs.Config.Iterations); err != nil {
		return err
	}

	return nil
}

// runConcurrencyBenchmarks runs concurrent operation benchmarks
func (bs *BenchmarkSuite) runConcurrencyBenchmarks(ctx context.Context) error {
	fmt.Println("Running concurrency benchmarks...")

	for _, concurrency := range bs.Config.ConcurrencyLevels {
		// Concurrent encryption
		if err := bs.benchmarkConcurrentEncryption(ctx, concurrency, 1024, 100); err != nil {
			return err
		}

		// Concurrent key operations
		if err := bs.benchmarkConcurrentKeyOperations(ctx, concurrency, 100); err != nil {
			return err
		}
	}

	return nil
}

// Individual benchmark methods

func (bs *BenchmarkSuite) benchmarkKeyGeneration(ctx context.Context, algorithm string, iterations int) error {
	return bs.runBenchmark(ctx, fmt.Sprintf("key_generation_%s", algorithm), iterations, func() error {
		_, err := bs.CryptoProvider.GenerateKeyPair(algorithm)
		return err
	}, map[string]interface{}{
		"algorithm": algorithm,
	})
}

func (bs *BenchmarkSuite) benchmarkEncryption(ctx context.Context, kemAlg, demAlg string, messageSize, iterations int) error {
	// Setup
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Generate KEM key pair for encryption
	recipientKeys, err := bs.CryptoProvider.GenerateKeyPair(kemAlg)
	if err != nil {
		return err
	}

	// Generate signature key pair for signing
	signatureKeys, err := bs.CryptoProvider.GenerateKeyPair("dilithium3")
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("encryption_%s_%s_%db", kemAlg, demAlg, messageSize), iterations, func() error {
		_, err := bs.CryptoProvider.EncryptMessage(plaintext, recipientKeys.PublicKey, signatureKeys.PrivateKey)
		return err
	}, map[string]interface{}{
		"kem_algorithm":  kemAlg,
		"dem_algorithm":  demAlg,
		"message_size":   messageSize,
		"recipient_keys": recipientKeys.Algorithm,
	})
}

func (bs *BenchmarkSuite) benchmarkDecryption(ctx context.Context, kemAlg, demAlg string, messageSize, iterations int) error {
	// Setup
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Generate KEM key pair for encryption/decryption
	recipientKeys, err := bs.CryptoProvider.GenerateKeyPair(kemAlg)
	if err != nil {
		return err
	}

	// Generate signature key pair for signing/verification
	signatureKeys, err := bs.CryptoProvider.GenerateKeyPair("dilithium3")
	if err != nil {
		return err
	}

	envelope, err := bs.CryptoProvider.EncryptMessage(plaintext, recipientKeys.PublicKey, signatureKeys.PrivateKey)
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("decryption_%s_%s_%db", kemAlg, demAlg, messageSize), iterations, func() error {
		_, err := bs.CryptoProvider.DecryptMessage(envelope, recipientKeys.PrivateKey, signatureKeys.PublicKey)
		return err
	}, map[string]interface{}{
		"kem_algorithm": kemAlg,
		"dem_algorithm": demAlg,
		"message_size":  messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkSigning(ctx context.Context, algorithm string, messageSize, iterations int) error {
	message := make([]byte, messageSize)
	for i := range message {
		message[i] = byte(i % 256)
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("signing_%s_%db", algorithm, messageSize), iterations, func() error {
		// Placeholder for signing benchmark - actual implementation would use proper signing
		return nil
	}, map[string]interface{}{
		"algorithm":    algorithm,
		"message_size": messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkVerification(ctx context.Context, algorithm string, messageSize, iterations int) error {
	message := make([]byte, messageSize)
	for i := range message {
		message[i] = byte(i % 256)
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("verification_%s_%db", algorithm, messageSize), iterations, func() error {
		// Placeholder for verification benchmark - actual implementation would use proper verification
		return nil
	}, map[string]interface{}{
		"algorithm":    algorithm,
		"message_size": messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkE2EMessageEncryption(ctx context.Context, messageSize, iterations int) error {
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Use the client's own key pair for consistent encryption/decryption
	clientPublicKey := bs.Client.GetPublicKey()

	return bs.runBenchmark(ctx, fmt.Sprintf("e2e_message_encryption_%db", messageSize), iterations, func() error {
		_, err := bs.Client.EncryptMessage(plaintext, clientPublicKey, "test_thread", bs.Client.userID)
		return err
	}, map[string]interface{}{
		"operation_type": "e2e_message_encryption",
		"message_size":   messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkE2EMessageDecryption(ctx context.Context, messageSize, iterations int) error {
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Use the client's own key pair for consistent encryption/decryption
	clientPublicKey := bs.Client.GetPublicKey()

	message, err := bs.Client.EncryptMessage(plaintext, clientPublicKey, "test_thread", bs.Client.userID)
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("e2e_message_decryption_%db", messageSize), iterations, func() error {
		_, err := bs.Client.DecryptMessage(message, clientPublicKey)
		return err
	}, map[string]interface{}{
		"operation_type": "e2e_message_decryption",
		"message_size":   messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkThreadMessageEncryption(ctx context.Context, messageSize, iterations int) error {
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	thread, err := bs.Client.CreateThread([]string{"participant1", "participant2"})
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("thread_message_encryption_%db", messageSize), iterations, func() error {
		_, err := bs.Client.EncryptThreadMessage(plaintext, thread)
		return err
	}, map[string]interface{}{
		"operation_type": "thread_message_encryption",
		"message_size":   messageSize,
		"thread_id":      thread.ID,
	})
}

func (bs *BenchmarkSuite) benchmarkThreadMessageDecryption(ctx context.Context, messageSize, iterations int) error {
	plaintext := make([]byte, messageSize)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	thread, err := bs.Client.CreateThread([]string{"participant1", "participant2"})
	if err != nil {
		return err
	}

	message, err := bs.Client.EncryptThreadMessage(plaintext, thread)
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, fmt.Sprintf("thread_message_decryption_%db", messageSize), iterations, func() error {
		_, err := bs.Client.DecryptThreadMessage(message, thread)
		return err
	}, map[string]interface{}{
		"operation_type": "thread_message_decryption",
		"message_size":   messageSize,
		"thread_id":      thread.ID,
	})
}

func (bs *BenchmarkSuite) benchmarkKeyRegistration(ctx context.Context, iterations int) error {
	return bs.runBenchmark(ctx, "key_registration", iterations, func() error {
		userUUID := fmt.Sprintf("user_%d", time.Now().UnixNano())
		publicKey := "mock_public_key_data"
		_, err := bs.KeyTransparency.RegisterPublicKey(userUUID, publicKey, "kyber768")
		return err
	}, map[string]interface{}{
		"operation_type": "key_transparency_registration",
	})
}

func (bs *BenchmarkSuite) benchmarkKeyVerification(ctx context.Context, iterations int) error {
	// Setup: Register a key first
	userUUID := "test_user_verification"
	publicKey := "mock_public_key_data"
	entry, err := bs.KeyTransparency.RegisterPublicKey(userUUID, publicKey, "kyber768")
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, "key_verification", iterations, func() error {
		_, err := bs.KeyTransparency.VerifyPublicKey(userUUID, entry.PublicKey, entry.EntryHash)
		return err
	}, map[string]interface{}{
		"operation_type": "key_transparency_verification",
		"user_uuid":      userUUID,
	})
}

func (bs *BenchmarkSuite) benchmarkKeyTransparencyAudit(ctx context.Context, iterations int) error {
	return bs.runBenchmark(ctx, "key_transparency_audit", iterations, func() error {
		results, err := bs.KeyTransparency.AuditLog(0, 10)
		if err != nil {
			return err
		}
		// Use results to avoid unused variable warning
		_ = results
		return nil
	}, map[string]interface{}{
		"operation_type": "key_transparency_audit",
		"audit_range":    "0-10",
	})
}

func (bs *BenchmarkSuite) benchmarkThresholdSigning(ctx context.Context, iterations int) error {
	message := []byte("test message for threshold signing benchmark")

	return bs.runBenchmark(ctx, "threshold_signing", iterations, func() error {
		_, err := bs.ThresholdHSM.Sign("test_key", message)
		return err
	}, map[string]interface{}{
		"operation_type": "threshold_hsm_signing",
		"message_size":   len(message),
	})
}

func (bs *BenchmarkSuite) benchmarkThresholdVerification(ctx context.Context, iterations int) error {
	message := []byte("test message for threshold verification benchmark")
	signature, err := bs.ThresholdHSM.Sign("test_key", message)
	if err != nil {
		return err
	}

	return bs.runBenchmark(ctx, "threshold_verification", iterations, func() error {
		_, err := bs.ThresholdHSM.Verify(signature, message)
		return err
	}, map[string]interface{}{
		"operation_type": "threshold_hsm_verification",
		"message_size":   len(message),
	})
}

func (bs *BenchmarkSuite) benchmarkKeyRotation(ctx context.Context, iterations int) error {
	return bs.runBenchmark(ctx, "key_rotation", iterations, func() error {
		return bs.Client.RotateKeys()
	}, map[string]interface{}{
		"operation_type": "key_rotation",
	})
}

func (bs *BenchmarkSuite) benchmarkConcurrentEncryption(ctx context.Context, concurrency, messageSize, iterations int) error {
	plaintext := make([]byte, messageSize)

	// Generate proper PQC key pair for benchmarking
	keyPair, err := bs.CryptoProvider.GenerateKeyPair("kyber768")
	if err != nil {
		return fmt.Errorf("failed to generate key pair for benchmark: %w", err)
	}
	recipientPublicKey := keyPair.PublicKey

	return bs.runConcurrentBenchmark(ctx, fmt.Sprintf("concurrent_encryption_%d_threads", concurrency), concurrency, iterations, func() error {
		_, err := bs.Client.EncryptMessage(plaintext, recipientPublicKey, "test_thread", "test_recipient")
		return err
	}, map[string]interface{}{
		"operation_type": "concurrent_encryption",
		"concurrency":    concurrency,
		"message_size":   messageSize,
	})
}

func (bs *BenchmarkSuite) benchmarkConcurrentKeyOperations(ctx context.Context, concurrency, iterations int) error {
	return bs.runConcurrentBenchmark(ctx, fmt.Sprintf("concurrent_key_operations_%d_threads", concurrency), concurrency, iterations, func() error {
		_, err := bs.CryptoProvider.GenerateKeyPair("kyber768")
		return err
	}, map[string]interface{}{
		"operation_type": "concurrent_key_operations",
		"concurrency":    concurrency,
	})
}

// Helper methods

func (bs *BenchmarkSuite) runBenchmark(ctx context.Context, name string, iterations int, operation func() error, metadata map[string]interface{}) error {
	return bs.runConcurrentBenchmark(ctx, name, 1, iterations, operation, metadata)
}

func (bs *BenchmarkSuite) runConcurrentBenchmark(ctx context.Context, name string, concurrency, iterations int, operation func() error, metadata map[string]interface{}) error {
	fmt.Printf("  Running %s (%d iterations, %d concurrency)...\n", name, iterations, concurrency)

	var memBefore, memAfter runtime.MemStats
	if bs.Config.CollectMemStats {
		runtime.GC()
		runtime.ReadMemStats(&memBefore)
	}

	start := time.Now()

	// Run concurrent operations
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)
	iterationsPerWorker := iterations / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerWorker; j++ {
				if err := operation(); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// Check for errors
	select {
	case err := <-errChan:
		bs.addResult(BenchmarkResult{
			Operation:    name,
			Duration:     duration,
			Timestamp:    time.Now(),
			Success:      false,
			ErrorMessage: err.Error(),
			Metadata:     metadata,
		})
		return fmt.Errorf("benchmark failed: %w", err)
	default:
	}

	if bs.Config.CollectMemStats {
		runtime.ReadMemStats(&memAfter)
	}

	// Avoid division by zero for very fast operations
	var throughput float64
	if duration.Seconds() > 0 {
		throughput = float64(iterations) / duration.Seconds()
	} else {
		throughput = float64(iterations) // For operations that complete instantly
	}
	memoryUsage := int64(memAfter.Alloc - memBefore.Alloc)

	result := BenchmarkResult{
		Operation:   name,
		Duration:    duration,
		Throughput:  throughput,
		MemoryUsage: memoryUsage,
		Timestamp:   time.Now(),
		Success:     true,
		Metadata:    metadata,
	}

	bs.addResult(result)

	fmt.Printf("    Completed in %v (%.2f ops/sec)\n", duration, throughput)
	return nil
}

func (bs *BenchmarkSuite) addResult(result BenchmarkResult) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.Results = append(bs.Results, result)
}

// GetResults returns all benchmark results
func (bs *BenchmarkSuite) GetResults() []BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	results := make([]BenchmarkResult, len(bs.Results))
	copy(results, bs.Results)
	return results
}

// GetResultsByOperation returns results filtered by operation name
func (bs *BenchmarkSuite) GetResultsByOperation(operation string) []BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	var filtered []BenchmarkResult
	for _, result := range bs.Results {
		if result.Operation == operation {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// GenerateReport generates a comprehensive performance report
func (bs *BenchmarkSuite) GenerateReport() ([]byte, error) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	report := map[string]interface{}{
		"timestamp":   time.Now(),
		"config":      bs.Config,
		"total_tests": len(bs.Results),
		"successful_tests": func() int {
			count := 0
			for _, r := range bs.Results {
				if r.Success {
					count++
				}
			}
			return count
		}(),
		"results": bs.Results,
		"summary": bs.generateSummary(),
	}

	return json.MarshalIndent(report, "", "  ")
}

func (bs *BenchmarkSuite) generateSummary() map[string]interface{} {
	operationGroups := make(map[string][]BenchmarkResult)

	for _, result := range bs.Results {
		if result.Success {
			operationGroups[result.Operation] = append(operationGroups[result.Operation], result)
		}
	}

	summary := make(map[string]interface{})
	for operation, results := range operationGroups {
		if len(results) > 0 {
			stats := bs.calculateStats(results)
			summary[operation] = stats
		}
	}

	return summary
}

func (bs *BenchmarkSuite) calculateStats(results []BenchmarkResult) BenchmarkStats {
	if len(results) == 0 {
		return BenchmarkStats{}
	}

	// Calculate basic stats
	var totalDuration time.Duration
	var totalThroughput float64
	minDuration := results[0].Duration
	maxDuration := results[0].Duration

	for _, result := range results {
		totalDuration += result.Duration
		totalThroughput += result.Throughput

		if result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	mean := totalDuration / time.Duration(len(results))
	avgThroughput := totalThroughput / float64(len(results))

	// Calculate percentiles (simplified)
	// Note: For production use, consider using a proper percentile calculation library
	p95Index := int(float64(len(results)) * 0.95)
	p99Index := int(float64(len(results)) * 0.99)

	if p95Index >= len(results) {
		p95Index = len(results) - 1
	}
	if p99Index >= len(results) {
		p99Index = len(results) - 1
	}

	return BenchmarkStats{
		Mean:       mean,
		Median:     results[len(results)/2].Duration, // Simplified median
		Min:        minDuration,
		Max:        maxDuration,
		P95:        results[p95Index].Duration,
		P99:        results[p99Index].Duration,
		Throughput: avgThroughput,
	}
}

// DefaultBenchmarkConfig returns a default benchmark configuration
func DefaultBenchmarkConfig() BenchmarkConfig {
	return BenchmarkConfig{
		Iterations:        1000,
		WarmupRuns:        10,
		Timeout:           30 * time.Second,
		MessageSizes:      []int{256, 1024, 4096, 16384, 65536},
		ConcurrencyLevels: []int{1, 4, 8, 16, 32},
		KEMAlgorithms:     []string{"kyber512", "kyber768", "kyber1024"},
		DEMAlgorithms:     []string{"aes256gcm", "chacha20poly1305"},
		EnableGC:          true,
		CollectMemStats:   true,
	}
}
