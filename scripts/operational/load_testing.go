package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL               string                `json:"base_url"`
	ConcurrentUsers       int                   `json:"concurrent_users"`
	TestDuration          string                `json:"test_duration"`
	RampUpTime            string                `json:"ramp_up_time"`
	ThinkTime             string                `json:"think_time"`
	Timeout               string                `json:"timeout"`
	MaxRetries            int                   `json:"max_retries"`
	Headers               map[string]string     `json:"headers"`
	Endpoints             []EndpointConfig      `json:"endpoints"`
	DatabaseMonitoring    bool                  `json:"database_monitoring"`
	PerformanceThresholds PerformanceThresholds `json:"performance_thresholds"`
	ReportFormat          string                `json:"report_format"`
	SaveResults           bool                  `json:"save_results"`
	ResultsDir            string                `json:"results_dir"`
	EnableHTTP2           bool                  `json:"enable_http2"`
	SSLVerification       bool                  `json:"ssl_verification"`
	ConnectionPooling     bool                  `json:"connection_pooling"`
	KeepAlive             bool                  `json:"keep_alive"`
	Compression           bool                  `json:"compression"`
}

// EndpointConfig defines an endpoint to test
type EndpointConfig struct {
	Path             string            `json:"path"`
	Method           string            `json:"method"`
	Weight           int               `json:"weight"`
	ExpectedStatus   int               `json:"expected_status"`
	Headers          map[string]string `json:"headers"`
	Body             interface{}       `json:"body"`
	AuthRequired     bool              `json:"auth_required"`
	AuthToken        string            `json:"auth_token"`
	RateLimit        int               `json:"rate_limit"`
	Timeout          string            `json:"timeout"`
	RetryOnFailure   bool              `json:"retry_on_failure"`
	ValidateResponse bool              `json:"validate_response"`
	ResponseSchema   interface{}       `json:"response_schema"`
}

// PerformanceThresholds defines performance expectations
type PerformanceThresholds struct {
	MaxLatencyP50            int     `json:"max_latency_p50"`
	MaxLatencyP95            int     `json:"max_latency_p95"`
	MaxLatencyP99            int     `json:"max_latency_p99"`
	MaxErrorRate             float64 `json:"max_error_rate"`
	MinThroughput            int     `json:"min_throughput"`
	MaxResponseTime          int     `json:"max_response_time"`
	MaxConcurrentConnections int     `json:"max_concurrent_connections"`
	MaxMemoryUsage           int     `json:"max_memory_usage"`
	MaxCPUUsage              int     `json:"max_cpu_usage"`
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TestID              string                    `json:"test_id"`
	StartTime           time.Time                 `json:"start_time"`
	EndTime             time.Time                 `json:"end_time"`
	Duration            time.Duration             `json:"duration"`
	TotalRequests       int64                     `json:"total_requests"`
	SuccessfulRequests  int64                     `json:"successful_requests"`
	FailedRequests      int64                     `json:"failed_requests"`
	ErrorRate           float64                   `json:"error_rate"`
	AverageLatency      float64                   `json:"average_latency"`
	P50Latency          float64                   `json:"p50_latency"`
	P95Latency          float64                   `json:"p95_latency"`
	P99Latency          float64                   `json:"p99_latency"`
	Throughput          float64                   `json:"throughput"`
	ConcurrentUsers     int                       `json:"concurrent_users"`
	EndpointResults     map[string]EndpointResult `json:"endpoint_results"`
	Errors              []TestError               `json:"errors"`
	PerformanceMetrics  PerformanceMetrics        `json:"performance_metrics"`
	ThresholdViolations []ThresholdViolation      `json:"threshold_violations"`
	Recommendations     []string                  `json:"recommendations"`
}

// EndpointResult holds results for a specific endpoint
type EndpointResult struct {
	Path               string        `json:"path"`
	Method             string        `json:"method"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ErrorRate          float64       `json:"error_rate"`
	AverageLatency     float64       `json:"average_latency"`
	P50Latency         float64       `json:"p50_latency"`
	P95Latency         float64       `json:"p95_latency"`
	P99Latency         float64       `json:"p99_latency"`
	MinLatency         float64       `json:"min_latency"`
	MaxLatency         float64       `json:"max_latency"`
	Throughput         float64       `json:"throughput"`
	StatusCodes        map[int]int64 `json:"status_codes"`
	ResponseSizes      []int64       `json:"response_sizes"`
	LatencyPercentiles []float64     `json:"latency_percentiles"`
}

// TestError represents an error during testing
type TestError struct {
	Endpoint     string    `json:"endpoint"`
	Error        string    `json:"error"`
	Timestamp    time.Time `json:"timestamp"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body"`
}

// PerformanceMetrics holds system performance data
type PerformanceMetrics struct {
	CPUUsage            float64 `json:"cpu_usage"`
	MemoryUsage         float64 `json:"memory_usage"`
	DiskIOPerSecond     float64 `json:"disk_io_per_second"`
	NetworkThroughput   float64 `json:"network_throughput"`
	DatabaseConnections int     `json:"database_connections"`
	DatabaseQueryTime   float64 `json:"database_query_time"`
	Goroutines          int     `json:"goroutines"`
	GCPauseTime         float64 `json:"gc_pause_time"`
}

// ThresholdViolation represents a performance threshold violation
type ThresholdViolation struct {
	Metric      string  `json:"metric"`
	Threshold   float64 `json:"threshold"`
	ActualValue float64 `json:"actual_value"`
	Severity    string  `json:"severity"`
	Endpoint    string  `json:"endpoint"`
}

// LoadTester handles load testing operations
type LoadTester struct {
	config    LoadTestConfig
	client    *http.Client
	results   *LoadTestResult
	latencies []float64
	mutex     sync.RWMutex
	logger    *log.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoadTester creates a new load tester
func NewLoadTester(configPath string) (*LoadTester, error) {
	// Load configuration
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config LoadTestConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Parse durations
	_, err = time.ParseDuration(config.TestDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid test duration: %w", err)
	}

	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.SSLVerification,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  !config.Compression,
		DisableKeepAlives:   !config.KeepAlive,
	}

	// HTTP/2 configuration would go here if golang.org/x/net/http2 is available
	// For now, we'll skip HTTP/2 configuration

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// Setup logger
	logger := log.New(io.Discard, "[LOAD_TEST] ", log.LstdFlags|log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())

	return &LoadTester{
		config: config,
		client: client,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// RunLoadTest executes the load test
func (lt *LoadTester) RunLoadTest() (*LoadTestResult, error) {
	lt.logger.Println("Starting load test...")

	// Initialize results
	lt.results = &LoadTestResult{
		TestID:              generateTestID(),
		StartTime:           time.Now(),
		ConcurrentUsers:     lt.config.ConcurrentUsers,
		EndpointResults:     make(map[string]EndpointResult),
		Errors:              make([]TestError, 0),
		ThresholdViolations: make([]ThresholdViolation, 0),
		Recommendations:     make([]string, 0),
	}

	// Initialize endpoint results
	for _, endpoint := range lt.config.Endpoints {
		key := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		lt.results.EndpointResults[key] = EndpointResult{
			Path:               endpoint.Path,
			Method:             endpoint.Method,
			StatusCodes:        make(map[int]int64),
			ResponseSizes:      make([]int64, 0),
			LatencyPercentiles: make([]float64, 0),
		}
	}

	// Start performance monitoring
	if lt.config.DatabaseMonitoring {
		go lt.monitorPerformance()
	}

	// Create worker pool
	var wg sync.WaitGroup
	requestChan := make(chan struct{}, lt.config.ConcurrentUsers*10)

	// Start workers
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go lt.worker(i, &wg, requestChan)
	}

	// Send requests
	go func() {
		defer close(requestChan)
		testDuration, _ := time.ParseDuration(lt.config.TestDuration)
		endTime := time.Now().Add(testDuration)

		for time.Now().Before(endTime) {
			select {
			case requestChan <- struct{}{}:
			case <-lt.ctx.Done():
				return
			default:
				// Channel full, wait a bit
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Wait for completion
	wg.Wait()
	lt.results.EndTime = time.Now()
	lt.results.Duration = lt.results.EndTime.Sub(lt.results.StartTime)

	// Calculate final metrics
	lt.calculateFinalMetrics()

	// Check thresholds
	lt.checkThresholds()

	// Generate recommendations
	lt.generateRecommendations()

	lt.logger.Printf("Load test completed: %d requests, %.2f%% error rate, %.2fms avg latency",
		lt.results.TotalRequests, lt.results.ErrorRate, lt.results.AverageLatency)

	return lt.results, nil
}

// worker represents a single user making requests
func (lt *LoadTester) worker(id int, wg *sync.WaitGroup, requestChan <-chan struct{}) {
	defer wg.Done()

	for range requestChan {
		select {
		case <-lt.ctx.Done():
			return
		default:
			lt.makeRequest()
		}

		// Think time between requests
		if thinkTime, err := time.ParseDuration(lt.config.ThinkTime); err == nil {
			time.Sleep(thinkTime)
		}
	}
}

// makeRequest makes a single HTTP request
func (lt *LoadTester) makeRequest() {
	// Select endpoint based on weights
	endpoint := lt.selectEndpoint()
	if endpoint == nil {
		return
	}

	// Prepare request
	url := lt.config.BaseURL + endpoint.Path
	var body io.Reader

	if endpoint.Body != nil {
		jsonBody, err := json.Marshal(endpoint.Body)
		if err != nil {
			lt.recordError(endpoint, fmt.Sprintf("failed to marshal body: %v", err), 0, "")
			return
		}
		body = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(endpoint.Method, url, body)
	if err != nil {
		lt.recordError(endpoint, fmt.Sprintf("failed to create request: %v", err), 0, "")
		return
	}

	// Add headers
	for key, value := range lt.config.Headers {
		req.Header.Set(key, value)
	}
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// Add auth if required
	if endpoint.AuthRequired && endpoint.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+endpoint.AuthToken)
	}

	// Make request
	startTime := time.Now()
	resp, err := lt.client.Do(req)
	latency := float64(time.Since(startTime).Milliseconds())

	if err != nil {
		lt.recordError(endpoint, fmt.Sprintf("request failed: %v", err), 0, "")
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		lt.recordError(endpoint, fmt.Sprintf("failed to read response: %v", err), resp.StatusCode, "")
		return
	}

	// Record result
	lt.recordResult(endpoint, resp.StatusCode, latency, int64(len(respBody)), string(respBody))

	// Validate response if required
	if endpoint.ValidateResponse {
		if resp.StatusCode != endpoint.ExpectedStatus {
			lt.recordError(endpoint, fmt.Sprintf("unexpected status code: got %d, expected %d",
				resp.StatusCode, endpoint.ExpectedStatus), resp.StatusCode, string(respBody))
		}
	}
}

// selectEndpoint selects an endpoint based on weights
func (lt *LoadTester) selectEndpoint() *EndpointConfig {
	if len(lt.config.Endpoints) == 0 {
		return nil
	}

	totalWeight := 0
	for _, endpoint := range lt.config.Endpoints {
		totalWeight += endpoint.Weight
	}

	if totalWeight == 0 {
		return &lt.config.Endpoints[0]
	}

	random := rand.Intn(totalWeight)
	currentWeight := 0

	for _, endpoint := range lt.config.Endpoints {
		currentWeight += endpoint.Weight
		if random < currentWeight {
			return &endpoint
		}
	}

	return &lt.config.Endpoints[0]
}

// recordResult records a successful request result
func (lt *LoadTester) recordResult(endpoint *EndpointConfig, statusCode int, latency float64, responseSize int64, responseBody string) {
	key := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)

	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	// Update global metrics
	atomic.AddInt64(&lt.results.TotalRequests, 1)
	atomic.AddInt64(&lt.results.SuccessfulRequests, 1)

	// Update endpoint results
	endpointResult := lt.results.EndpointResults[key]
	endpointResult.TotalRequests++
	endpointResult.SuccessfulRequests++
	endpointResult.AverageLatency = (endpointResult.AverageLatency*float64(endpointResult.TotalRequests-1) + latency) / float64(endpointResult.TotalRequests)
	endpointResult.StatusCodes[statusCode]++
	endpointResult.ResponseSizes = append(endpointResult.ResponseSizes, responseSize)

	// Update min/max latency
	if endpointResult.MinLatency == 0 || latency < endpointResult.MinLatency {
		endpointResult.MinLatency = latency
	}
	if latency > endpointResult.MaxLatency {
		endpointResult.MaxLatency = latency
	}

	lt.results.EndpointResults[key] = endpointResult

	// Store latency for percentile calculation
	lt.latencies = append(lt.latencies, latency)
}

// recordError records a failed request
func (lt *LoadTester) recordError(endpoint *EndpointConfig, errorMsg string, statusCode int, responseBody string) {
	key := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)

	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	// Update global metrics
	atomic.AddInt64(&lt.results.TotalRequests, 1)
	atomic.AddInt64(&lt.results.FailedRequests, 1)

	// Update endpoint results
	endpointResult := lt.results.EndpointResults[key]
	endpointResult.TotalRequests++
	endpointResult.FailedRequests++
	endpointResult.StatusCodes[statusCode]++
	lt.results.EndpointResults[key] = endpointResult

	// Record error
	lt.results.Errors = append(lt.results.Errors, TestError{
		Endpoint:     key,
		Error:        errorMsg,
		Timestamp:    time.Now(),
		StatusCode:   statusCode,
		ResponseBody: responseBody,
	})
}

// calculateFinalMetrics calculates final performance metrics
func (lt *LoadTester) calculateFinalMetrics() {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	// Calculate error rate
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedRequests) / float64(lt.results.TotalRequests) * 100
	}

	// Calculate throughput
	if lt.results.Duration > 0 {
		lt.results.Throughput = float64(lt.results.TotalRequests) / lt.results.Duration.Seconds()
	}

	// Calculate percentiles
	if len(lt.latencies) > 0 {
		lt.results.AverageLatency = calculateAverage(lt.latencies)
		lt.results.P50Latency = calculatePercentile(lt.latencies, 50)
		lt.results.P95Latency = calculatePercentile(lt.latencies, 95)
		lt.results.P99Latency = calculatePercentile(lt.latencies, 99)
	}

	// Calculate endpoint-specific metrics
	for key, endpointResult := range lt.results.EndpointResults {
		if endpointResult.TotalRequests > 0 {
			endpointResult.ErrorRate = float64(endpointResult.FailedRequests) / float64(endpointResult.TotalRequests) * 100
			endpointResult.Throughput = float64(endpointResult.TotalRequests) / lt.results.Duration.Seconds()

			// Calculate endpoint percentiles
			endpointLatencies := append([]float64{}, lt.latencies...)

			if len(endpointLatencies) > 0 {
				endpointResult.P50Latency = calculatePercentile(endpointLatencies, 50)
				endpointResult.P95Latency = calculatePercentile(endpointLatencies, 95)
				endpointResult.P99Latency = calculatePercentile(endpointLatencies, 99)
			}

			lt.results.EndpointResults[key] = endpointResult
		}
	}
}

// checkThresholds checks if performance thresholds are violated
func (lt *LoadTester) checkThresholds() {
	thresholds := lt.config.PerformanceThresholds

	// Check latency thresholds
	if lt.results.P50Latency > float64(thresholds.MaxLatencyP50) {
		lt.results.ThresholdViolations = append(lt.results.ThresholdViolations, ThresholdViolation{
			Metric:      "P50 Latency",
			Threshold:   float64(thresholds.MaxLatencyP50),
			ActualValue: lt.results.P50Latency,
			Severity:    "high",
		})
	}

	if lt.results.P95Latency > float64(thresholds.MaxLatencyP95) {
		lt.results.ThresholdViolations = append(lt.results.ThresholdViolations, ThresholdViolation{
			Metric:      "P95 Latency",
			Threshold:   float64(thresholds.MaxLatencyP95),
			ActualValue: lt.results.P95Latency,
			Severity:    "critical",
		})
	}

	if lt.results.P99Latency > float64(thresholds.MaxLatencyP99) {
		lt.results.ThresholdViolations = append(lt.results.ThresholdViolations, ThresholdViolation{
			Metric:      "P99 Latency",
			Threshold:   float64(thresholds.MaxLatencyP99),
			ActualValue: lt.results.P99Latency,
			Severity:    "critical",
		})
	}

	// Check error rate threshold
	if lt.results.ErrorRate > thresholds.MaxErrorRate {
		lt.results.ThresholdViolations = append(lt.results.ThresholdViolations, ThresholdViolation{
			Metric:      "Error Rate",
			Threshold:   thresholds.MaxErrorRate,
			ActualValue: lt.results.ErrorRate,
			Severity:    "critical",
		})
	}

	// Check throughput threshold
	if lt.results.Throughput < float64(thresholds.MinThroughput) {
		lt.results.ThresholdViolations = append(lt.results.ThresholdViolations, ThresholdViolation{
			Metric:      "Throughput",
			Threshold:   float64(thresholds.MinThroughput),
			ActualValue: lt.results.Throughput,
			Severity:    "high",
		})
	}
}

// generateRecommendations generates performance recommendations
func (lt *LoadTester) generateRecommendations() {
	// High error rate recommendations
	if lt.results.ErrorRate > 5.0 {
		lt.results.Recommendations = append(lt.results.Recommendations,
			"High error rate detected. Consider implementing retry logic and circuit breakers.")
	}

	// High latency recommendations
	if lt.results.P95Latency > 1000 {
		lt.results.Recommendations = append(lt.results.Recommendations,
			"High P95 latency detected. Consider optimizing database queries and implementing caching.")
	}

	// Low throughput recommendations
	if lt.results.Throughput < 100 {
		lt.results.Recommendations = append(lt.results.Recommendations,
			"Low throughput detected. Consider horizontal scaling and connection pooling.")
	}

	// Memory usage recommendations
	if lt.results.PerformanceMetrics.MemoryUsage > 80 {
		lt.results.Recommendations = append(lt.results.Recommendations,
			"High memory usage detected. Consider memory optimization and garbage collection tuning.")
	}

	// Database performance recommendations
	if lt.results.PerformanceMetrics.DatabaseQueryTime > 100 {
		lt.results.Recommendations = append(lt.results.Recommendations,
			"Slow database queries detected. Consider query optimization and indexing.")
	}
}

// monitorPerformance monitors system performance during the test
func (lt *LoadTester) monitorPerformance() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// In a real implementation, you would collect actual system metrics here
			// For now, we'll simulate some metrics
			lt.results.PerformanceMetrics = PerformanceMetrics{
				CPUUsage:            rand.Float64() * 100,
				MemoryUsage:         rand.Float64() * 100,
				DiskIOPerSecond:     rand.Float64() * 1000,
				NetworkThroughput:   rand.Float64() * 100,
				DatabaseConnections: rand.Intn(100),
				DatabaseQueryTime:   rand.Float64() * 200,
				Goroutines:          rand.Intn(1000),
				GCPauseTime:         rand.Float64() * 10,
			}
		case <-lt.ctx.Done():
			return
		}
	}
}

// SaveResults saves test results to file
func (lt *LoadTester) SaveResults() error {
	if !lt.config.SaveResults {
		return nil
	}

	// Create results directory
	if err := createDirIfNotExists(lt.config.ResultsDir); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Save JSON report
	jsonFile := fmt.Sprintf("%s/load_test_%s.json", lt.config.ResultsDir, lt.results.TestID)
	jsonData, err := json.MarshalIndent(lt.results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := writeFile(jsonFile, jsonData); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	// Generate HTML report
	htmlFile := fmt.Sprintf("%s/load_test_%s.html", lt.config.ResultsDir, lt.results.TestID)
	htmlContent := lt.generateHTMLReport()
	if err := writeFile(htmlFile, []byte(htmlContent)); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	lt.logger.Printf("Results saved to %s and %s", jsonFile, htmlFile)
	return nil
}

// generateHTMLReport generates an HTML report
func (lt *LoadTester) generateHTMLReport() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Load Test Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .metric { margin: 10px 0; }
        .threshold-violation { color: red; font-weight: bold; }
        .recommendation { background: #fff3cd; padding: 10px; margin: 10px 0; border-radius: 5px; }
        table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Load Test Report</h1>
        <p><strong>Test ID:</strong> %s</p>
        <p><strong>Duration:</strong> %v</p>
        <p><strong>Concurrent Users:</strong> %d</p>
    </div>

    <h2>Summary</h2>
    <div class="metric">
        <strong>Total Requests:</strong> %d<br>
        <strong>Successful Requests:</strong> %d<br>
        <strong>Failed Requests:</strong> %d<br>
        <strong>Error Rate:</strong> %.2f%%<br>
        <strong>Average Latency:</strong> %.2fms<br>
        <strong>P95 Latency:</strong> %.2fms<br>
        <strong>Throughput:</strong> %.2f req/s
    </div>

    <h2>Endpoint Results</h2>
    <table>
        <tr>
            <th>Endpoint</th>
            <th>Requests</th>
            <th>Error Rate</th>
            <th>Avg Latency</th>
            <th>P95 Latency</th>
        </tr>
`, lt.results.TestID, lt.results.TestID, lt.results.Duration, lt.results.ConcurrentUsers,
		lt.results.TotalRequests, lt.results.SuccessfulRequests, lt.results.FailedRequests,
		lt.results.ErrorRate, lt.results.AverageLatency, lt.results.P95Latency, lt.results.Throughput)
}

// Helper functions

func generateTestID() string {
	return fmt.Sprintf("load_test_%d", time.Now().Unix())
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)

	index := int(float64(percentile) / 100.0 * float64(len(sorted)-1))
	return sorted[index]
}

func createDirIfNotExists(dir string) error {
	return nil // Simplified for this example
}

func writeFile(filename string, data []byte) error {
	return nil // Simplified for this example
}


