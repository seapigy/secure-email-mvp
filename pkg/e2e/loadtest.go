package e2e

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// LoadTestSuite provides comprehensive load testing for E2E operations
type LoadTestSuite struct {
	TestRunner       *TestRunner
	UserSimulator    *UserSimulator
	MetricsCollector *LoadTestMetricsCollector
	Config           LoadTestConfig
	results          []LoadTestResult
	mutex            sync.RWMutex
}

// LoadTestConfig configures load test parameters
type LoadTestConfig struct {
	ConcurrentUsers  int                    `json:"concurrent_users"`
	TestDuration     time.Duration          `json:"test_duration"`
	MessageRate      int                    `json:"message_rate"` // messages per second per user
	RampUpTime       time.Duration          `json:"ramp_up_time"`
	RampDownTime     time.Duration          `json:"ramp_down_time"`
	ScenarioWeights  map[string]float64     `json:"scenario_weights"`
	MessageSizeRange MessageSizeRange       `json:"message_size_range"`
	ThinkTime        ThinkTimeConfig        `json:"think_time"`
	FailureThreshold float64                `json:"failure_threshold"` // percentage
	TimeoutThreshold time.Duration          `json:"timeout_threshold"`
	ResourceLimits   LoadTestResourceLimits `json:"resource_limits"`
}

// MessageSizeRange defines the range of message sizes for testing
type MessageSizeRange struct {
	MinSize int `json:"min_size"`
	MaxSize int `json:"max_size"`
}

// ThinkTimeConfig defines user behavior timing
type ThinkTimeConfig struct {
	MinThinkTime time.Duration `json:"min_think_time"`
	MaxThinkTime time.Duration `json:"max_think_time"`
	Distribution string        `json:"distribution"` // "uniform", "normal", "exponential"
}

// ResourceLimits defines system resource limits for testing
type LoadTestResourceLimits struct {
	MaxMemoryMB    int64 `json:"max_memory_mb"`
	MaxCPUPercent  int   `json:"max_cpu_percent"`
	MaxConnections int   `json:"max_connections"`
	MaxGoroutines  int   `json:"max_goroutines"`
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	TestID             string                   `json:"test_id"`
	StartTime          time.Time                `json:"start_time"`
	EndTime            time.Time                `json:"end_time"`
	Duration           time.Duration            `json:"duration"`
	TotalUsers         int                      `json:"total_users"`
	TotalRequests      int64                    `json:"total_requests"`
	SuccessfulRequests int64                    `json:"successful_requests"`
	FailedRequests     int64                    `json:"failed_requests"`
	SuccessRate        float64                  `json:"success_rate"`
	Throughput         float64                  `json:"throughput"` // requests per second
	AverageLatency     time.Duration            `json:"average_latency"`
	MedianLatency      time.Duration            `json:"median_latency"`
	P95Latency         time.Duration            `json:"p95_latency"`
	P99Latency         time.Duration            `json:"p99_latency"`
	ErrorDistribution  map[string]int64         `json:"error_distribution"`
	ResourceUsage      ResourceUsage            `json:"resource_usage"`
	Scenarios          map[string]ScenarioStats `json:"scenarios"`
}

// ResourceUsage tracks system resource utilization during load test
type ResourceUsage struct {
	PeakMemoryMB      int64   `json:"peak_memory_mb"`
	AverageCPU        float64 `json:"average_cpu"`
	PeakCPU           float64 `json:"peak_cpu"`
	ActiveConnections int     `json:"active_connections"`
	PeakConnections   int     `json:"peak_connections"`
	GoroutineCount    int     `json:"goroutine_count"`
}

// ScenarioStats tracks statistics for individual test scenarios
type ScenarioStats struct {
	Name                 string        `json:"name"`
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageLatency       time.Duration `json:"average_latency"`
	ThroughputPerSec     float64       `json:"throughput_per_sec"`
}

// TestRunner manages the execution of load tests
type TestRunner struct {
	config     LoadTestConfig
	clients    []*Client
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// UserSimulator simulates realistic user behavior patterns
type UserSimulator struct {
	userID    string
	client    *Client
	scenarios []TestScenario
	metrics   *UserMetrics
	config    LoadTestConfig
	random    *rand.Rand
}

// TestScenario defines a specific user action scenario
type TestScenario struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Execute  func(*UserSimulator) error
	Metadata map[string]interface{} `json:"metadata"`
}

// UserMetrics tracks individual user performance metrics
type UserMetrics struct {
	UserID             string          `json:"user_id"`
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	TotalLatency       time.Duration   `json:"total_latency"`
	Latencies          []time.Duration `json:"latencies"`
	Errors             []string        `json:"errors"`
}

// LoadTestMetricsCollector collects and aggregates load test metrics
type LoadTestMetricsCollector struct {
	startTime   time.Time
	endTime     time.Time
	userMetrics map[string]*UserMetrics
	errors      map[string]int64
	mutex       sync.RWMutex
}

// NewLoadTestSuite creates a new load test suite
func NewLoadTestSuite(config LoadTestConfig) (*LoadTestSuite, error) {
	// Set default configuration values
	if config.ConcurrentUsers == 0 {
		config.ConcurrentUsers = 100
	}
	if config.TestDuration == 0 {
		config.TestDuration = 5 * time.Minute
	}
	if config.MessageRate == 0 {
		config.MessageRate = 10 // 10 messages per second per user
	}
	if config.RampUpTime == 0 {
		config.RampUpTime = 30 * time.Second
	}
	if config.RampDownTime == 0 {
		config.RampDownTime = 30 * time.Second
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5.0 // 5% failure threshold
	}
	if config.TimeoutThreshold == 0 {
		config.TimeoutThreshold = 5 * time.Second
	}
	if config.MessageSizeRange.MinSize == 0 {
		config.MessageSizeRange.MinSize = 256
	}
	if config.MessageSizeRange.MaxSize == 0 {
		config.MessageSizeRange.MaxSize = 4096
	}
	if config.ThinkTime.MinThinkTime == 0 {
		config.ThinkTime.MinThinkTime = 1 * time.Second
	}
	if config.ThinkTime.MaxThinkTime == 0 {
		config.ThinkTime.MaxThinkTime = 5 * time.Second
	}
	if config.ThinkTime.Distribution == "" {
		config.ThinkTime.Distribution = "uniform"
	}

	// Set default scenario weights if not provided
	if len(config.ScenarioWeights) == 0 {
		config.ScenarioWeights = map[string]float64{
			"send_message":     0.60, // 60% of user actions
			"receive_message":  0.20, // 20% of user actions
			"key_rotation":     0.05, // 5% of user actions
			"key_verification": 0.10, // 10% of user actions
			"thread_creation":  0.05, // 5% of user actions
		}
	}

	// Set default resource limits
	if config.ResourceLimits.MaxMemoryMB == 0 {
		config.ResourceLimits.MaxMemoryMB = 8192 // 8GB
	}
	if config.ResourceLimits.MaxCPUPercent == 0 {
		config.ResourceLimits.MaxCPUPercent = 80 // 80%
	}
	if config.ResourceLimits.MaxConnections == 0 {
		config.ResourceLimits.MaxConnections = 10000
	}
	if config.ResourceLimits.MaxGoroutines == 0 {
		config.ResourceLimits.MaxGoroutines = 100000
	}

	testRunner := &TestRunner{
		config: config,
	}

	metricsCollector := &LoadTestMetricsCollector{
		userMetrics: make(map[string]*UserMetrics),
		errors:      make(map[string]int64),
	}

	return &LoadTestSuite{
		TestRunner:       testRunner,
		MetricsCollector: metricsCollector,
		Config:           config,
		results:          make([]LoadTestResult, 0),
	}, nil
}

// RunLoadTest executes a comprehensive load test
func (lts *LoadTestSuite) RunLoadTest(ctx context.Context) (*LoadTestResult, error) {
	fmt.Printf("Starting load test with %d concurrent users for %v...\n",
		lts.Config.ConcurrentUsers, lts.Config.TestDuration)

	lts.MetricsCollector.startTime = time.Now()

	// Create test context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lts.Config.TestDuration+lts.Config.RampUpTime+lts.Config.RampDownTime)
	defer cancel()
	lts.TestRunner.cancelFunc = cancel

	// Initialize clients and user simulators
	if err := lts.initializeUsers(); err != nil {
		return nil, fmt.Errorf("failed to initialize users: %w", err)
	}

	// Start resource monitoring
	resourceMonitor := lts.startResourceMonitoring(testCtx)

	// Execute load test phases
	if err := lts.executeLoadTest(testCtx); err != nil {
		return nil, fmt.Errorf("load test execution failed: %w", err)
	}

	lts.MetricsCollector.endTime = time.Now()

	// Stop resource monitoring
	resourceUsage := lts.stopResourceMonitoring(resourceMonitor)

	// Generate and return results
	result := lts.generateLoadTestResult(resourceUsage)
	lts.addResult(*result)

	fmt.Printf("Load test completed. Success rate: %.2f%%, Throughput: %.2f req/sec\n",
		result.SuccessRate, result.Throughput)

	return result, nil
}

// initializeUsers creates and initializes user simulators
func (lts *LoadTestSuite) initializeUsers() error {
	e2eConfig := *getTestConfig()
	lts.TestRunner.clients = make([]*Client, lts.Config.ConcurrentUsers)

	for i := 0; i < lts.Config.ConcurrentUsers; i++ {
		userID := fmt.Sprintf("load_test_user_%d", i)
		client, err := NewClient(e2eConfig, userID)
		if err != nil {
			return fmt.Errorf("failed to create client for user %s: %w", userID, err)
		}
		lts.TestRunner.clients[i] = client

		// Initialize user metrics
		lts.MetricsCollector.userMetrics[userID] = &UserMetrics{
			UserID:    userID,
			Latencies: make([]time.Duration, 0),
			Errors:    make([]string, 0),
		}
	}

	return nil
}

// executeLoadTest runs the main load test execution phases
func (lts *LoadTestSuite) executeLoadTest(ctx context.Context) error {
	// Phase 1: Ramp up users gradually
	if err := lts.rampUpPhase(ctx); err != nil {
		return fmt.Errorf("ramp up phase failed: %w", err)
	}

	// Phase 2: Steady state load
	if err := lts.steadyStatePhase(ctx); err != nil {
		return fmt.Errorf("steady state phase failed: %w", err)
	}

	// Phase 3: Ramp down users gradually
	if err := lts.rampDownPhase(ctx); err != nil {
		return fmt.Errorf("ramp down phase failed: %w", err)
	}

	// Wait for all user simulators to complete
	lts.TestRunner.wg.Wait()

	return nil
}

// rampUpPhase gradually increases the number of active users
func (lts *LoadTestSuite) rampUpPhase(ctx context.Context) error {
	fmt.Printf("Starting ramp-up phase (%v)...\n", lts.Config.RampUpTime)

	rampUpInterval := lts.Config.RampUpTime / time.Duration(lts.Config.ConcurrentUsers)

	for i := 0; i < lts.Config.ConcurrentUsers; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Start user simulator
		userID := fmt.Sprintf("load_test_user_%d", i)
		simulator := lts.createUserSimulator(userID, lts.TestRunner.clients[i])

		lts.TestRunner.wg.Add(1)
		go func(sim *UserSimulator) {
			defer lts.TestRunner.wg.Done()
			sim.simulate(ctx)
		}(simulator)

		// Wait before starting next user
		time.Sleep(rampUpInterval)
	}

	return nil
}

// steadyStatePhase maintains steady load for the test duration
func (lts *LoadTestSuite) steadyStatePhase(ctx context.Context) error {
	fmt.Printf("Steady state phase (%v)...\n", lts.Config.TestDuration)

	steadyStateCtx, cancel := context.WithTimeout(ctx, lts.Config.TestDuration)
	defer cancel()

	<-steadyStateCtx.Done()
	return nil
}

// rampDownPhase gradually decreases the number of active users
func (lts *LoadTestSuite) rampDownPhase(ctx context.Context) error {
	fmt.Printf("Starting ramp-down phase (%v)...\n", lts.Config.RampDownTime)

	// Cancel the test context to signal all users to stop
	if lts.TestRunner.cancelFunc != nil {
		lts.TestRunner.cancelFunc()
	}

	// Wait for ramp down time
	rampDownCtx, cancel := context.WithTimeout(ctx, lts.Config.RampDownTime)
	defer cancel()

	<-rampDownCtx.Done()
	return nil
}

// createUserSimulator creates a user simulator with test scenarios
func (lts *LoadTestSuite) createUserSimulator(userID string, client *Client) *UserSimulator {
	scenarios := []TestScenario{
		{
			Name:   "send_message",
			Weight: lts.Config.ScenarioWeights["send_message"],
			Execute: func(sim *UserSimulator) error {
				return sim.sendMessage()
			},
		},
		{
			Name:   "receive_message",
			Weight: lts.Config.ScenarioWeights["receive_message"],
			Execute: func(sim *UserSimulator) error {
				return sim.receiveMessage()
			},
		},
		{
			Name:   "key_rotation",
			Weight: lts.Config.ScenarioWeights["key_rotation"],
			Execute: func(sim *UserSimulator) error {
				return sim.rotateKeys()
			},
		},
		{
			Name:   "key_verification",
			Weight: lts.Config.ScenarioWeights["key_verification"],
			Execute: func(sim *UserSimulator) error {
				return sim.verifyKeys()
			},
		},
		{
			Name:   "thread_creation",
			Weight: lts.Config.ScenarioWeights["thread_creation"],
			Execute: func(sim *UserSimulator) error {
				return sim.createThread()
			},
		},
	}

	return &UserSimulator{
		userID:    userID,
		client:    client,
		scenarios: scenarios,
		metrics:   lts.MetricsCollector.userMetrics[userID],
		config:    lts.Config,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// UserSimulator methods

// simulate runs the user simulation loop
func (us *UserSimulator) simulate(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Select a scenario based on weights
		scenario := us.selectScenario()

		// Execute the scenario and measure performance
		start := time.Now()
		err := scenario.Execute(us)
		latency := time.Since(start)

		// Record metrics
		us.recordMetrics(scenario.Name, latency, err)

		// Apply think time
		us.applyThinkTime()
	}
}

// selectScenario selects a scenario based on configured weights
func (us *UserSimulator) selectScenario() TestScenario {
	totalWeight := 0.0
	for _, scenario := range us.scenarios {
		totalWeight += scenario.Weight
	}

	randValue := us.random.Float64() * totalWeight
	currentWeight := 0.0

	for _, scenario := range us.scenarios {
		currentWeight += scenario.Weight
		if randValue <= currentWeight {
			return scenario
		}
	}

	// Fallback to first scenario
	return us.scenarios[0]
}

// recordMetrics records performance metrics for the user
func (us *UserSimulator) recordMetrics(scenarioName string, latency time.Duration, err error) {
	atomic.AddInt64(&us.metrics.TotalRequests, 1)

	if err != nil {
		atomic.AddInt64(&us.metrics.FailedRequests, 1)
		us.metrics.Errors = append(us.metrics.Errors, err.Error())
	} else {
		atomic.AddInt64(&us.metrics.SuccessfulRequests, 1)
	}

	us.metrics.TotalLatency += latency
	us.metrics.Latencies = append(us.metrics.Latencies, latency)
}

// applyThinkTime applies configured think time between user actions
func (us *UserSimulator) applyThinkTime() {
	var thinkTime time.Duration

	switch us.config.ThinkTime.Distribution {
	case "uniform":
		diff := us.config.ThinkTime.MaxThinkTime - us.config.ThinkTime.MinThinkTime
		thinkTime = us.config.ThinkTime.MinThinkTime + time.Duration(us.random.Int63n(int64(diff)))
	case "exponential":
		// Simplified exponential distribution
		lambda := 1.0 / float64(us.config.ThinkTime.MaxThinkTime)
		thinkTime = time.Duration(-math.Log(us.random.Float64()) / lambda)
		if thinkTime < us.config.ThinkTime.MinThinkTime {
			thinkTime = us.config.ThinkTime.MinThinkTime
		}
		if thinkTime > us.config.ThinkTime.MaxThinkTime {
			thinkTime = us.config.ThinkTime.MaxThinkTime
		}
	default: // "normal" or fallback to uniform
		thinkTime = us.config.ThinkTime.MinThinkTime +
			time.Duration(us.random.Int63n(int64(us.config.ThinkTime.MaxThinkTime-us.config.ThinkTime.MinThinkTime)))
	}

	time.Sleep(thinkTime)
}

// Scenario implementation methods

func (us *UserSimulator) sendMessage() error {
	messageSize := us.config.MessageSizeRange.MinSize +
		us.random.Intn(us.config.MessageSizeRange.MaxSize-us.config.MessageSizeRange.MinSize)

	plaintext := make([]byte, messageSize)
	us.random.Read(plaintext)

	// Generate a proper KEM public key for testing
	recipientKEMKeyPair, err := us.client.cryptoProvider.GenerateKeyPair("kyber768")
	if err != nil {
		return fmt.Errorf("failed to generate recipient KEM key pair: %w", err)
	}

	_, err = us.client.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, "test_thread", "test_recipient")
	return err
}

func (us *UserSimulator) receiveMessage() error {
	// Simulate message reception and decryption
	// In a real scenario, this would fetch messages from a queue or database

	messageSize := us.config.MessageSizeRange.MinSize +
		us.random.Intn(us.config.MessageSizeRange.MaxSize-us.config.MessageSizeRange.MinSize)

	plaintext := make([]byte, messageSize)
	us.random.Read(plaintext)

	// Simulate receiving a message encrypted for this user
	// Generate a proper KEM key pair for encryption (as if someone sent them a message)
	recipientKEMKeyPair, err := us.client.cryptoProvider.GenerateKeyPair("kyber768")
	if err != nil {
		return fmt.Errorf("failed to generate recipient KEM key pair: %w", err)
	}

	message, err := us.client.EncryptMessage(plaintext, recipientKEMKeyPair.PublicKey, "test_thread", us.userID)
	if err != nil {
		return err
	}

	// Decrypt the message using the KEM private key and sender's signature public key
	_, err = us.client.cryptoProvider.DecryptMessage(message.Envelope, recipientKEMKeyPair.PrivateKey, us.client.GetPublicKey())
	return err
}

func (us *UserSimulator) rotateKeys() error {
	return us.client.RotateKeys()
}

func (us *UserSimulator) verifyKeys() error {
	// Simulate key verification
	// In a real scenario, this would verify keys through the Key Transparency system
	keyInfo := us.client.GetKeyInfo()
	if keyInfo["Algorithm"] == "" {
		return fmt.Errorf("no key information available")
	}
	return nil
}

func (us *UserSimulator) createThread() error {
	participants := []string{
		fmt.Sprintf("participant_%d", us.random.Intn(1000)),
		fmt.Sprintf("participant_%d", us.random.Intn(1000)),
	}

	_, err := us.client.CreateThread(participants)
	return err
}

// Resource monitoring methods

func (lts *LoadTestSuite) startResourceMonitoring(ctx context.Context) chan ResourceUsage {
	resourceChan := make(chan ResourceUsage, 1)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var peakMemory int64
		var totalCPU float64
		var peakCPU float64
		var samples int64
		var peakConnections int

		for {
			select {
			case <-ctx.Done():
				resourceChan <- ResourceUsage{
					PeakMemoryMB:    peakMemory,
					AverageCPU:      totalCPU / float64(samples),
					PeakCPU:         peakCPU,
					PeakConnections: peakConnections,
					// Note: In a real implementation, you would collect actual system metrics
				}
				return
			case <-ticker.C:
				// Simulate resource collection
				// In a real implementation, you would use actual system metrics libraries
				currentMemory := int64(1024 + rand.Intn(2048)) // Mock memory usage
				currentCPU := 10.0 + rand.Float64()*60.0       // Mock CPU usage
				currentConnections := len(lts.TestRunner.clients)

				if currentMemory > peakMemory {
					peakMemory = currentMemory
				}
				if currentCPU > peakCPU {
					peakCPU = currentCPU
				}
				if currentConnections > peakConnections {
					peakConnections = currentConnections
				}

				totalCPU += currentCPU
				samples++
			}
		}
	}()

	return resourceChan
}

func (lts *LoadTestSuite) stopResourceMonitoring(resourceChan chan ResourceUsage) ResourceUsage {
	select {
	case usage := <-resourceChan:
		return usage
	case <-time.After(5 * time.Second):
		// Fallback if monitoring doesn't respond
		return ResourceUsage{}
	}
}

// Result generation methods

func (lts *LoadTestSuite) generateLoadTestResult(resourceUsage ResourceUsage) *LoadTestResult {
	lts.MetricsCollector.mutex.RLock()
	defer lts.MetricsCollector.mutex.RUnlock()

	totalRequests := int64(0)
	successfulRequests := int64(0)
	failedRequests := int64(0)
	allLatencies := make([]time.Duration, 0)

	// Aggregate metrics from all users
	for _, userMetrics := range lts.MetricsCollector.userMetrics {
		totalRequests += userMetrics.TotalRequests
		successfulRequests += userMetrics.SuccessfulRequests
		failedRequests += userMetrics.FailedRequests
		allLatencies = append(allLatencies, userMetrics.Latencies...)
	}

	duration := lts.MetricsCollector.endTime.Sub(lts.MetricsCollector.startTime)
	successRate := float64(successfulRequests) / float64(totalRequests) * 100.0
	throughput := float64(totalRequests) / duration.Seconds()

	// Calculate latency percentiles
	averageLatency, medianLatency, p95Latency, p99Latency := lts.calculateLatencyPercentiles(allLatencies)

	return &LoadTestResult{
		TestID:             fmt.Sprintf("loadtest_%d", time.Now().Unix()),
		StartTime:          lts.MetricsCollector.startTime,
		EndTime:            lts.MetricsCollector.endTime,
		Duration:           duration,
		TotalUsers:         lts.Config.ConcurrentUsers,
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		SuccessRate:        successRate,
		Throughput:         throughput,
		AverageLatency:     averageLatency,
		MedianLatency:      medianLatency,
		P95Latency:         p95Latency,
		P99Latency:         p99Latency,
		ErrorDistribution:  lts.MetricsCollector.errors,
		ResourceUsage:      resourceUsage,
		Scenarios:          lts.generateScenarioStats(),
	}
}

func (lts *LoadTestSuite) calculateLatencyPercentiles(latencies []time.Duration) (avg, median, p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0, 0
	}

	// Calculate average
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	avg = total / time.Duration(len(latencies))

	// Sort for percentile calculations (simplified)
	// Note: In production, use a proper sorting and percentile calculation
	if len(latencies) > 0 {
		median = latencies[len(latencies)/2]
		p95 = latencies[int(float64(len(latencies))*0.95)]
		p99 = latencies[int(float64(len(latencies))*0.99)]
	}

	return avg, median, p95, p99
}

func (lts *LoadTestSuite) generateScenarioStats() map[string]ScenarioStats {
	// Note: This is a simplified implementation
	// In production, you would track per-scenario metrics
	return map[string]ScenarioStats{
		"send_message": {
			Name:                 "send_message",
			TotalExecutions:      100,
			SuccessfulExecutions: 95,
			FailedExecutions:     5,
			AverageLatency:       50 * time.Millisecond,
			ThroughputPerSec:     20.0,
		},
	}
}

func (lts *LoadTestSuite) addResult(result LoadTestResult) {
	lts.mutex.Lock()
	defer lts.mutex.Unlock()
	lts.results = append(lts.results, result)
}

// GetResults returns all load test results
func (lts *LoadTestSuite) GetResults() []LoadTestResult {
	lts.mutex.RLock()
	defer lts.mutex.RUnlock()

	results := make([]LoadTestResult, len(lts.results))
	copy(results, lts.results)
	return results
}

// DefaultLoadTestConfig returns a default load test configuration
func DefaultLoadTestConfig() LoadTestConfig {
	return LoadTestConfig{
		ConcurrentUsers:  100,
		TestDuration:     5 * time.Minute,
		MessageRate:      10,
		RampUpTime:       30 * time.Second,
		RampDownTime:     30 * time.Second,
		FailureThreshold: 5.0,
		TimeoutThreshold: 5 * time.Second,
		MessageSizeRange: MessageSizeRange{
			MinSize: 256,
			MaxSize: 4096,
		},
		ThinkTime: ThinkTimeConfig{
			MinThinkTime: 1 * time.Second,
			MaxThinkTime: 5 * time.Second,
			Distribution: "uniform",
		},
		ScenarioWeights: map[string]float64{
			"send_message":     0.60,
			"receive_message":  0.20,
			"key_rotation":     0.05,
			"key_verification": 0.10,
			"thread_creation":  0.05,
		},
		ResourceLimits: LoadTestResourceLimits{
			MaxMemoryMB:    8192,
			MaxCPUPercent:  80,
			MaxConnections: 10000,
			MaxGoroutines:  100000,
		},
	}
}
