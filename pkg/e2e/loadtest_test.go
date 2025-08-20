package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoadTestSuite(t *testing.T) {
	config := DefaultLoadTestConfig()
	suite, err := NewLoadTestSuite(config)

	require.NoError(t, err)
	assert.NotNil(t, suite)
	assert.NotNil(t, suite.TestRunner)
	assert.NotNil(t, suite.MetricsCollector)
	assert.Equal(t, config.ConcurrentUsers, suite.Config.ConcurrentUsers)
}

func TestLoadTestConfig_Defaults(t *testing.T) {
	config := DefaultLoadTestConfig()

	assert.Equal(t, 100, config.ConcurrentUsers)
	assert.Equal(t, 5*time.Minute, config.TestDuration)
	assert.Equal(t, 10, config.MessageRate)
	assert.Equal(t, 30*time.Second, config.RampUpTime)
	assert.Equal(t, 30*time.Second, config.RampDownTime)
	assert.Equal(t, 5.0, config.FailureThreshold)
	assert.Equal(t, 5*time.Second, config.TimeoutThreshold)

	// Check message size range
	assert.Equal(t, 256, config.MessageSizeRange.MinSize)
	assert.Equal(t, 4096, config.MessageSizeRange.MaxSize)

	// Check think time config
	assert.Equal(t, 1*time.Second, config.ThinkTime.MinThinkTime)
	assert.Equal(t, 5*time.Second, config.ThinkTime.MaxThinkTime)
	assert.Equal(t, "uniform", config.ThinkTime.Distribution)

	// Check scenario weights
	expectedWeights := map[string]float64{
		"send_message":     0.60,
		"receive_message":  0.20,
		"key_rotation":     0.05,
		"key_verification": 0.10,
		"thread_creation":  0.05,
	}
	assert.Equal(t, expectedWeights, config.ScenarioWeights)

	// Check resource limits
	assert.Equal(t, int64(8192), config.ResourceLimits.MaxMemoryMB)
	assert.Equal(t, 80, config.ResourceLimits.MaxCPUPercent)
	assert.Equal(t, 10000, config.ResourceLimits.MaxConnections)
	assert.Equal(t, 100000, config.ResourceLimits.MaxGoroutines)
}

func TestLoadTestSuite_InitializeUsers(t *testing.T) {
	config := LoadTestConfig{
		ConcurrentUsers: 5,
		TestDuration:    1 * time.Second,
	}

	suite, err := NewLoadTestSuite(config)
	require.NoError(t, err)

	err = suite.initializeUsers()
	require.NoError(t, err)

	assert.Len(t, suite.TestRunner.clients, 5)
	assert.Len(t, suite.MetricsCollector.userMetrics, 5)

	// Verify all clients are initialized
	for i, client := range suite.TestRunner.clients {
		assert.NotNil(t, client)
		userID := fmt.Sprintf("load_test_user_%d", i)
		assert.Contains(t, suite.MetricsCollector.userMetrics, userID)
	}
}

func TestLoadTestSuite_ShortLoadTest(t *testing.T) {
	// Use a very short test to verify the load test framework works
	config := LoadTestConfig{
		ConcurrentUsers:  2,
		TestDuration:     2 * time.Second,
		MessageRate:      5,
		RampUpTime:       500 * time.Millisecond,
		RampDownTime:     500 * time.Millisecond,
		FailureThreshold: 50.0, // High threshold for test
		TimeoutThreshold: 1 * time.Second,
		MessageSizeRange: MessageSizeRange{
			MinSize: 256,
			MaxSize: 512,
		},
		ThinkTime: ThinkTimeConfig{
			MinThinkTime: 10 * time.Millisecond,
			MaxThinkTime: 50 * time.Millisecond,
			Distribution: "uniform",
		},
		ScenarioWeights: map[string]float64{
			"send_message":     0.80,
			"receive_message":  0.20,
			"key_rotation":     0.00, // Disable expensive operations for test
			"key_verification": 0.00,
			"thread_creation":  0.00,
		},
	}

	suite, err := NewLoadTestSuite(config)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := suite.RunLoadTest(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify result structure
	assert.NotEmpty(t, result.TestID)
	assert.Equal(t, 2, result.TotalUsers)
	assert.True(t, result.TotalRequests > 0)
	assert.True(t, result.SuccessRate >= 0.0 && result.SuccessRate <= 100.0)
	assert.True(t, result.Throughput > 0.0)
	assert.True(t, result.Duration > 0)

	// Verify timing
	assert.True(t, result.StartTime.Before(result.EndTime))

	// Verify results are stored
	results := suite.GetResults()
	assert.Len(t, results, 1)
	assert.Equal(t, result.TestID, results[0].TestID)
}

func TestUserSimulator_ScenarioSelection(t *testing.T) {
	config := LoadTestConfig{
		ScenarioWeights: map[string]float64{
			"send_message":    0.50,
			"receive_message": 0.30,
			"key_rotation":    0.20,
		},
	}

	userID := "test_user"
	client, err := NewClient(*getTestConfig(), userID)
	require.NoError(t, err)

	userMetrics := &UserMetrics{
		UserID:    userID,
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	scenarios := []TestScenario{
		{Name: "send_message", Weight: 0.50},
		{Name: "receive_message", Weight: 0.30},
		{Name: "key_rotation", Weight: 0.20},
	}

	simulator := &UserSimulator{
		userID:    userID,
		client:    client,
		scenarios: scenarios,
		metrics:   userMetrics,
		config:    config,
		random:    rand.New(rand.NewSource(42)), // Fixed seed for reproducible test
	}

	// Test scenario selection multiple times
	selectedScenarios := make(map[string]int)
	for i := 0; i < 1000; i++ {
		scenario := simulator.selectScenario()
		selectedScenarios[scenario.Name]++
	}

	// Verify distribution is roughly correct (allowing for randomness)
	total := 1000
	assert.True(t, selectedScenarios["send_message"] > int(float64(total)*0.4))    // ~50% +/- 10%
	assert.True(t, selectedScenarios["receive_message"] > int(float64(total)*0.2)) // ~30% +/- 10%
	assert.True(t, selectedScenarios["key_rotation"] > int(float64(total)*0.1))    // ~20% +/- 10%
}

func TestUserSimulator_ThinkTimeDistributions(t *testing.T) {
	config := LoadTestConfig{
		ThinkTime: ThinkTimeConfig{
			MinThinkTime: 100 * time.Millisecond,
			MaxThinkTime: 500 * time.Millisecond,
			Distribution: "uniform",
		},
	}

	userID := "test_user"
	client, err := NewClient(*getTestConfig(), userID)
	require.NoError(t, err)

	userMetrics := &UserMetrics{
		UserID:    userID,
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	simulator := &UserSimulator{
		userID:  userID,
		client:  client,
		metrics: userMetrics,
		config:  config,
		random:  rand.New(rand.NewSource(42)),
	}

	// Test uniform distribution
	config.ThinkTime.Distribution = "uniform"
	simulator.config = config

	start := time.Now()
	simulator.applyThinkTime()
	elapsed := time.Since(start)

	// Should be within the configured range
	assert.True(t, elapsed >= config.ThinkTime.MinThinkTime)
	assert.True(t, elapsed <= config.ThinkTime.MaxThinkTime*2) // Allow some tolerance

	// Test exponential distribution
	config.ThinkTime.Distribution = "exponential"
	simulator.config = config

	start = time.Now()
	simulator.applyThinkTime()
	elapsed = time.Since(start)

	// Should be within reasonable bounds
	assert.True(t, elapsed >= config.ThinkTime.MinThinkTime)
	assert.True(t, elapsed <= config.ThinkTime.MaxThinkTime*3) // Exponential can have longer tail
}

func TestUserSimulator_MessageScenarios(t *testing.T) {
	config := LoadTestConfig{
		MessageSizeRange: MessageSizeRange{
			MinSize: 256,
			MaxSize: 1024,
		},
	}

	userID := "test_user"
	client, err := NewClient(*getTestConfig(), userID)
	require.NoError(t, err)

	userMetrics := &UserMetrics{
		UserID:    userID,
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	simulator := &UserSimulator{
		userID:  userID,
		client:  client,
		metrics: userMetrics,
		config:  config,
		random:  rand.New(rand.NewSource(42)),
	}

	// Test send message scenario
	err = simulator.sendMessage()
	assert.NoError(t, err)

	// Test receive message scenario
	err = simulator.receiveMessage()
	assert.NoError(t, err)

	// Test key rotation scenario
	err = simulator.rotateKeys()
	assert.NoError(t, err)

	// Test key verification scenario
	err = simulator.verifyKeys()
	assert.NoError(t, err)

	// Test thread creation scenario
	err = simulator.createThread()
	assert.NoError(t, err)
}

func TestUserMetrics_Recording(t *testing.T) {
	userMetrics := &UserMetrics{
		UserID:    "test_user",
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	config := LoadTestConfig{}
	simulator := &UserSimulator{
		metrics: userMetrics,
		config:  config,
	}

	// Test successful operation
	simulator.recordMetrics("test_scenario", 100*time.Millisecond, nil)

	assert.Equal(t, int64(1), userMetrics.TotalRequests)
	assert.Equal(t, int64(1), userMetrics.SuccessfulRequests)
	assert.Equal(t, int64(0), userMetrics.FailedRequests)
	assert.Len(t, userMetrics.Latencies, 1)
	assert.Equal(t, 100*time.Millisecond, userMetrics.Latencies[0])
	assert.Len(t, userMetrics.Errors, 0)

	// Test failed operation
	testError := fmt.Errorf("test error")
	simulator.recordMetrics("test_scenario", 50*time.Millisecond, testError)

	assert.Equal(t, int64(2), userMetrics.TotalRequests)
	assert.Equal(t, int64(1), userMetrics.SuccessfulRequests)
	assert.Equal(t, int64(1), userMetrics.FailedRequests)
	assert.Len(t, userMetrics.Latencies, 2)
	assert.Len(t, userMetrics.Errors, 1)
	assert.Equal(t, "test error", userMetrics.Errors[0])
}

func TestLoadTestSuite_LatencyCalculations(t *testing.T) {
	suite := &LoadTestSuite{}

	// Test empty latencies
	avg, median, p95, p99 := suite.calculateLatencyPercentiles([]time.Duration{})
	assert.Equal(t, time.Duration(0), avg)
	assert.Equal(t, time.Duration(0), median)
	assert.Equal(t, time.Duration(0), p95)
	assert.Equal(t, time.Duration(0), p99)

	// Test with sample latencies
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	avg, median, p95, p99 = suite.calculateLatencyPercentiles(latencies)

	assert.Equal(t, 30*time.Millisecond, avg)    // (10+20+30+40+50)/5 = 30
	assert.Equal(t, 30*time.Millisecond, median) // Middle value
	assert.True(t, p95 > 0)
	assert.True(t, p99 > 0)
}

func TestLoadTestConfig_Validation(t *testing.T) {
	// Test configuration with zero values gets defaults
	config := LoadTestConfig{}
	suite, err := NewLoadTestSuite(config)

	require.NoError(t, err)

	// Verify defaults were applied
	assert.Equal(t, 100, suite.Config.ConcurrentUsers)
	assert.Equal(t, 5*time.Minute, suite.Config.TestDuration)
	assert.Equal(t, 10, suite.Config.MessageRate)
	assert.NotEmpty(t, suite.Config.ScenarioWeights)
}

func TestResourceMonitoring(t *testing.T) {
	config := DefaultLoadTestConfig()
	suite, err := NewLoadTestSuite(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start resource monitoring
	resourceChan := suite.startResourceMonitoring(ctx)

	// Wait for context to complete
	<-ctx.Done()

	// Stop monitoring and get results
	usage := suite.stopResourceMonitoring(resourceChan)

	// Verify we got some resource usage data
	// Note: These are mock values in our implementation
	assert.True(t, usage.PeakMemoryMB >= int64(0))
	assert.True(t, usage.AverageCPU >= 0.0)
}

// Benchmark tests for load testing framework
func BenchmarkUserSimulator_SendMessage(b *testing.B) {
	config := DefaultLoadTestConfig()
	userID := "benchmark_user"
	client, err := NewClient(*getTestConfig(), userID)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	userMetrics := &UserMetrics{
		UserID:    userID,
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	simulator := &UserSimulator{
		userID:  userID,
		client:  client,
		metrics: userMetrics,
		config:  config,
		random:  rand.New(rand.NewSource(42)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := simulator.sendMessage()
		if err != nil {
			b.Fatalf("Send message failed: %v", err)
		}
	}
}

func BenchmarkUserSimulator_ReceiveMessage(b *testing.B) {
	config := DefaultLoadTestConfig()
	userID := "benchmark_user"
	client, err := NewClient(*getTestConfig(), userID)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	userMetrics := &UserMetrics{
		UserID:    userID,
		Latencies: make([]time.Duration, 0),
		Errors:    make([]string, 0),
	}

	simulator := &UserSimulator{
		userID:  userID,
		client:  client,
		metrics: userMetrics,
		config:  config,
		random:  rand.New(rand.NewSource(42)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := simulator.receiveMessage()
		if err != nil {
			b.Fatalf("Receive message failed: %v", err)
		}
	}
}
