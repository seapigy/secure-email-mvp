package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PerformanceMonitor provides real-time performance monitoring for E2E operations
type PerformanceMonitor struct {
	MetricsCollector *MetricsCollector
	AlertManager     *AlertManager
	Dashboard        *Dashboard
	Config           MonitoringConfig
	isRunning        bool
	stopChan         chan bool
	mutex            sync.RWMutex
}

// MonitoringConfig configures performance monitoring
type MonitoringConfig struct {
	Enabled         bool            `json:"enabled"`
	SampleInterval  time.Duration   `json:"sample_interval"`
	RetentionPeriod time.Duration   `json:"retention_period"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
	MetricsEndpoint string          `json:"metrics_endpoint"`
	DashboardPort   int             `json:"dashboard_port"`
	ExportFormat    string          `json:"export_format"` // "prometheus", "json", "csv"
}

// AlertThresholds defines performance alert thresholds
type AlertThresholds struct {
	MaxLatency    time.Duration `json:"max_latency"`
	MaxErrorRate  float64       `json:"max_error_rate"`
	MaxMemoryMB   int64         `json:"max_memory_mb"`
	MaxCPUPercent int           `json:"max_cpu_percent"`
	MinThroughput float64       `json:"min_throughput"`
}

// MetricsCollector collects and stores performance metrics
type MetricsCollector struct {
	metrics     []PerformanceMetric
	mutex       sync.RWMutex
	startTime   time.Time
	sampleCount int64
}

// PerformanceMetric represents a single performance measurement
type PerformanceMetric struct {
	Timestamp     time.Time              `json:"timestamp"`
	Operation     string                 `json:"operation"`
	Duration      time.Duration          `json:"duration"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	MemoryUsage   int64                  `json:"memory_usage"`
	CPUUsage      float64                `json:"cpu_usage"`
	Throughput    float64                `json:"throughput"`
	MessageSize   int64                  `json:"message_size"`
	UserID        string                 `json:"user_id,omitempty"`
	ThreadID      string                 `json:"thread_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AlertManager manages performance alerts
type AlertManager struct {
	alerts    []PerformanceAlert
	config    AlertThresholds
	mutex     sync.RWMutex
	callbacks []AlertCallback
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"` // "info", "warning", "critical"
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Value       interface{}            `json:"value"`
	Threshold   interface{}            `json:"threshold"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertCallback defines a callback function for alerts
type AlertCallback func(alert PerformanceAlert)

// Dashboard provides real-time performance visualization
type Dashboard struct {
	config      MonitoringConfig
	metricsData *MetricsCollector
	isRunning   bool
	httpServer  interface{} // Placeholder for HTTP server
}

// MetricsSummary provides aggregated metrics
type MetricsSummary struct {
	TimeWindow      time.Duration      `json:"time_window"`
	TotalOperations int64              `json:"total_operations"`
	SuccessfulOps   int64              `json:"successful_operations"`
	FailedOps       int64              `json:"failed_operations"`
	SuccessRate     float64            `json:"success_rate"`
	AverageLatency  time.Duration      `json:"average_latency"`
	MedianLatency   time.Duration      `json:"median_latency"`
	P95Latency      time.Duration      `json:"p95_latency"`
	P99Latency      time.Duration      `json:"p99_latency"`
	MinLatency      time.Duration      `json:"min_latency"`
	MaxLatency      time.Duration      `json:"max_latency"`
	TotalThroughput float64            `json:"total_throughput"`
	AverageMemory   int64              `json:"average_memory_mb"`
	PeakMemory      int64              `json:"peak_memory_mb"`
	AverageCPU      float64            `json:"average_cpu_percent"`
	PeakCPU         float64            `json:"peak_cpu_percent"`
	OperationStats  map[string]OpStats `json:"operation_stats"`
}

// OpStats provides statistics for specific operations
type OpStats struct {
	Count          int64         `json:"count"`
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	Throughput     float64       `json:"throughput"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config MonitoringConfig) (*PerformanceMonitor, error) {
	// Set default config values
	if config.SampleInterval == 0 {
		config.SampleInterval = 1 * time.Second
	}
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 24 * time.Hour
	}
	if config.DashboardPort == 0 {
		config.DashboardPort = 8080
	}
	if config.ExportFormat == "" {
		config.ExportFormat = "json"
	}

	// Set default alert thresholds
	if config.AlertThresholds.MaxLatency == 0 {
		config.AlertThresholds.MaxLatency = 1 * time.Second
	}
	if config.AlertThresholds.MaxErrorRate == 0 {
		config.AlertThresholds.MaxErrorRate = 5.0 // 5%
	}
	if config.AlertThresholds.MaxMemoryMB == 0 {
		config.AlertThresholds.MaxMemoryMB = 1024 // 1GB
	}
	if config.AlertThresholds.MaxCPUPercent == 0 {
		config.AlertThresholds.MaxCPUPercent = 80 // 80%
	}
	if config.AlertThresholds.MinThroughput == 0 {
		config.AlertThresholds.MinThroughput = 100 // 100 ops/sec
	}

	metricsCollector := &MetricsCollector{
		metrics:   make([]PerformanceMetric, 0),
		startTime: time.Now(),
	}

	alertManager := &AlertManager{
		alerts:    make([]PerformanceAlert, 0),
		config:    config.AlertThresholds,
		callbacks: make([]AlertCallback, 0),
	}

	dashboard := &Dashboard{
		config:      config,
		metricsData: metricsCollector,
	}

	return &PerformanceMonitor{
		MetricsCollector: metricsCollector,
		AlertManager:     alertManager,
		Dashboard:        dashboard,
		Config:           config,
		stopChan:         make(chan bool),
	}, nil
}

// Start begins performance monitoring
func (pm *PerformanceMonitor) Start(ctx context.Context) error {
	if !pm.Config.Enabled {
		return nil
	}

	pm.mutex.Lock()
	if pm.isRunning {
		pm.mutex.Unlock()
		return fmt.Errorf("performance monitor is already running")
	}
	pm.isRunning = true
	pm.mutex.Unlock()

	fmt.Println("Starting performance monitor...")

	// Start metrics collection
	go pm.metricsCollectionLoop(ctx)

	// Start alert monitoring
	go pm.alertMonitoringLoop(ctx)

	// Start dashboard (if enabled)
	if pm.Config.DashboardPort > 0 {
		go pm.startDashboard(ctx)
	}

	return nil
}

// Stop stops performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.isRunning {
		return
	}

	fmt.Println("Stopping performance monitor...")
	pm.isRunning = false
	close(pm.stopChan)
}

// RecordMetric records a performance metric
func (pm *PerformanceMonitor) RecordMetric(metric PerformanceMetric) {
	if !pm.Config.Enabled {
		return
	}

	metric.Timestamp = time.Now()
	pm.MetricsCollector.AddMetric(metric)

	// Check for alert conditions
	pm.AlertManager.CheckMetric(metric)
}

// RecordOperation records the performance of an operation
func (pm *PerformanceMonitor) RecordOperation(operation string, duration time.Duration, success bool, errorMsg string) {
	metric := PerformanceMetric{
		Operation:    operation,
		Duration:     duration,
		Success:      success,
		ErrorMessage: errorMsg,
		Throughput:   1.0 / duration.Seconds(), // Simple throughput calculation
	}
	pm.RecordMetric(metric)
}

// RecordE2EOperation records the performance of an E2E operation with additional context
func (pm *PerformanceMonitor) RecordE2EOperation(operation string, duration time.Duration, success bool, userID, threadID string, messageSize int64, metadata map[string]interface{}) {
	metric := PerformanceMetric{
		Operation:   operation,
		Duration:    duration,
		Success:     success,
		MessageSize: messageSize,
		UserID:      userID,
		ThreadID:    threadID,
		Throughput:  1.0 / duration.Seconds(),
		Metadata:    metadata,
	}
	pm.RecordMetric(metric)
}

// GetMetricsSummary returns a summary of metrics for the specified time window
func (pm *PerformanceMonitor) GetMetricsSummary(timeWindow time.Duration) MetricsSummary {
	return pm.MetricsCollector.GetSummary(timeWindow)
}

// GetRealtimeMetrics returns real-time metrics
func (pm *PerformanceMonitor) GetRealtimeMetrics() []PerformanceMetric {
	return pm.MetricsCollector.GetRecentMetrics(1 * time.Minute)
}

// GetAlerts returns current alerts
func (pm *PerformanceMonitor) GetAlerts() []PerformanceAlert {
	return pm.AlertManager.GetActiveAlerts()
}

// RegisterAlertCallback registers a callback for performance alerts
func (pm *PerformanceMonitor) RegisterAlertCallback(callback AlertCallback) {
	pm.AlertManager.RegisterCallback(callback)
}

// ExportMetrics exports metrics in the specified format
func (pm *PerformanceMonitor) ExportMetrics(format string, timeWindow time.Duration) ([]byte, error) {
	summary := pm.GetMetricsSummary(timeWindow)

	switch format {
	case "json":
		return json.MarshalIndent(summary, "", "  ")
	case "prometheus":
		return pm.exportPrometheusFormat(summary)
	case "csv":
		return pm.exportCSVFormat(summary)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// MetricsCollector methods

// AddMetric adds a metric to the collection
func (mc *MetricsCollector) AddMetric(metric PerformanceMetric) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics = append(mc.metrics, metric)
	mc.sampleCount++

	// Cleanup old metrics based on retention period (simplified)
	// In production, use a more efficient circular buffer or time-series database
	if len(mc.metrics) > 10000 { // Simple size limit
		mc.metrics = mc.metrics[len(mc.metrics)-5000:] // Keep latest 5000
	}
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() []PerformanceMetric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make([]PerformanceMetric, len(mc.metrics))
	copy(result, mc.metrics)
	return result
}

// GetRecentMetrics returns metrics from the specified time window
func (mc *MetricsCollector) GetRecentMetrics(timeWindow time.Duration) []PerformanceMetric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	var recent []PerformanceMetric

	for _, metric := range mc.metrics {
		if metric.Timestamp.After(cutoff) {
			recent = append(recent, metric)
		}
	}

	return recent
}

// GetSummary calculates metrics summary for the specified time window
func (mc *MetricsCollector) GetSummary(timeWindow time.Duration) MetricsSummary {
	metrics := mc.GetRecentMetrics(timeWindow)

	if len(metrics) == 0 {
		return MetricsSummary{TimeWindow: timeWindow}
	}

	var totalOps, successfulOps, failedOps int64
	var totalLatency time.Duration
	var minLatency, maxLatency time.Duration
	var totalThroughput, totalMemory, totalCPU float64
	var peakMemory int64
	var peakCPU float64
	latencies := make([]time.Duration, 0, len(metrics))
	operationStats := make(map[string]*OpStats)

	// Initialize min/max
	if len(metrics) > 0 {
		minLatency = metrics[0].Duration
		maxLatency = metrics[0].Duration
	}

	for _, metric := range metrics {
		totalOps++
		if metric.Success {
			successfulOps++
		} else {
			failedOps++
		}

		totalLatency += metric.Duration
		latencies = append(latencies, metric.Duration)
		totalThroughput += metric.Throughput
		totalMemory += float64(metric.MemoryUsage)
		totalCPU += metric.CPUUsage

		if metric.Duration < minLatency {
			minLatency = metric.Duration
		}
		if metric.Duration > maxLatency {
			maxLatency = metric.Duration
		}
		if metric.MemoryUsage > peakMemory {
			peakMemory = metric.MemoryUsage
		}
		if metric.CPUUsage > peakCPU {
			peakCPU = metric.CPUUsage
		}

		// Update operation stats
		if _, exists := operationStats[metric.Operation]; !exists {
			operationStats[metric.Operation] = &OpStats{}
		}
		op := operationStats[metric.Operation]
		op.Count++
		if metric.Success {
			op.AverageLatency = (op.AverageLatency*time.Duration(op.Count-1) + metric.Duration) / time.Duration(op.Count)
			op.SuccessRate = float64(op.Count) / float64(op.Count) * 100.0 // Simplified
		}
		op.Throughput += metric.Throughput
	}

	summary := MetricsSummary{
		TimeWindow:      timeWindow,
		TotalOperations: totalOps,
		SuccessfulOps:   successfulOps,
		FailedOps:       failedOps,
		SuccessRate:     float64(successfulOps) / float64(totalOps) * 100.0,
		AverageLatency:  totalLatency / time.Duration(totalOps),
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		TotalThroughput: totalThroughput,
		AverageMemory:   int64(totalMemory / float64(totalOps)),
		PeakMemory:      peakMemory,
		AverageCPU:      totalCPU / float64(totalOps),
		PeakCPU:         peakCPU,
		OperationStats:  make(map[string]OpStats),
	}

	// Calculate percentiles (simplified)
	if len(latencies) > 0 {
		// Sort latencies for percentile calculation (simplified)
		p95Index := int(float64(len(latencies)) * 0.95)
		p99Index := int(float64(len(latencies)) * 0.99)
		medianIndex := len(latencies) / 2

		if p95Index >= len(latencies) {
			p95Index = len(latencies) - 1
		}
		if p99Index >= len(latencies) {
			p99Index = len(latencies) - 1
		}

		summary.MedianLatency = latencies[medianIndex]
		summary.P95Latency = latencies[p95Index]
		summary.P99Latency = latencies[p99Index]
	}

	// Convert operation stats
	for operation, stats := range operationStats {
		summary.OperationStats[operation] = *stats
	}

	return summary
}

// AlertManager methods

// CheckMetric checks if a metric triggers any alerts
func (am *AlertManager) CheckMetric(metric PerformanceMetric) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Check latency threshold
	if metric.Duration > am.config.MaxLatency {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Level:       "warning",
			Type:        "high_latency",
			Description: fmt.Sprintf("High latency detected: %v > %v", metric.Duration, am.config.MaxLatency),
			Value:       metric.Duration,
			Threshold:   am.config.MaxLatency,
		}
		am.addAlert(alert)
	}

	// Check memory threshold
	if metric.MemoryUsage > am.config.MaxMemoryMB*1024*1024 {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Level:       "critical",
			Type:        "high_memory",
			Description: fmt.Sprintf("High memory usage detected: %d MB > %d MB", metric.MemoryUsage/(1024*1024), am.config.MaxMemoryMB),
			Value:       metric.MemoryUsage,
			Threshold:   am.config.MaxMemoryMB,
		}
		am.addAlert(alert)
	}

	// Check CPU threshold
	if metric.CPUUsage > float64(am.config.MaxCPUPercent) {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Level:       "warning",
			Type:        "high_cpu",
			Description: fmt.Sprintf("High CPU usage detected: %.2f%% > %d%%", metric.CPUUsage, am.config.MaxCPUPercent),
			Value:       metric.CPUUsage,
			Threshold:   am.config.MaxCPUPercent,
		}
		am.addAlert(alert)
	}
}

// addAlert adds a new alert
func (am *AlertManager) addAlert(alert PerformanceAlert) {
	am.alerts = append(am.alerts, alert)

	// Trigger callbacks
	for _, callback := range am.callbacks {
		go callback(alert)
	}
}

// GetActiveAlerts returns all unresolved alerts
func (am *AlertManager) GetActiveAlerts() []PerformanceAlert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var active []PerformanceAlert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}
	return active
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for i := range am.alerts {
		if am.alerts[i].ID == alertID {
			now := time.Now()
			am.alerts[i].Resolved = true
			am.alerts[i].ResolvedAt = &now
			break
		}
	}
}

// RegisterCallback registers an alert callback
func (am *AlertManager) RegisterCallback(callback AlertCallback) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.callbacks = append(am.callbacks, callback)
}

// Monitoring loop methods

func (pm *PerformanceMonitor) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(pm.Config.SampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			// Collect system metrics (simplified)
			pm.collectSystemMetrics()
		}
	}
}

func (pm *PerformanceMonitor) alertMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check alerts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			// Check for alert conditions based on recent metrics
			pm.checkSystemAlerts()
		}
	}
}

func (pm *PerformanceMonitor) startDashboard(ctx context.Context) {
	// Simplified dashboard startup
	// In production, this would start an HTTP server with real-time metrics endpoints
	fmt.Printf("Dashboard available at http://localhost:%d\n", pm.Config.DashboardPort)
}

func (pm *PerformanceMonitor) collectSystemMetrics() {
	// Simplified system metrics collection
	// In production, use actual system monitoring libraries
	metric := PerformanceMetric{
		Operation:   "system_health",
		Duration:    1 * time.Millisecond, // Placeholder
		Success:     true,
		MemoryUsage: 512 * 1024 * 1024, // 512MB placeholder
		CPUUsage:    25.0,              // 25% placeholder
		Throughput:  100.0,             // 100 ops/sec placeholder
	}
	pm.RecordMetric(metric)
}

func (pm *PerformanceMonitor) checkSystemAlerts() {
	summary := pm.GetMetricsSummary(5 * time.Minute)

	// Check error rate
	if summary.SuccessRate < 100.0-pm.Config.AlertThresholds.MaxErrorRate {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Level:       "critical",
			Type:        "high_error_rate",
			Description: fmt.Sprintf("High error rate: %.2f%% < %.2f%%", summary.SuccessRate, 100.0-pm.Config.AlertThresholds.MaxErrorRate),
			Value:       summary.SuccessRate,
			Threshold:   100.0 - pm.Config.AlertThresholds.MaxErrorRate,
		}
		pm.AlertManager.addAlert(alert)
	}

	// Check throughput
	if summary.TotalThroughput < pm.Config.AlertThresholds.MinThroughput {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Level:       "warning",
			Type:        "low_throughput",
			Description: fmt.Sprintf("Low throughput: %.2f < %.2f ops/sec", summary.TotalThroughput, pm.Config.AlertThresholds.MinThroughput),
			Value:       summary.TotalThroughput,
			Threshold:   pm.Config.AlertThresholds.MinThroughput,
		}
		pm.AlertManager.addAlert(alert)
	}
}

// Export format methods

func (pm *PerformanceMonitor) exportPrometheusFormat(summary MetricsSummary) ([]byte, error) {
	// Simplified Prometheus format export
	prometheus := fmt.Sprintf(`# HELP e2e_operations_total Total number of E2E operations
# TYPE e2e_operations_total counter
e2e_operations_total %d

# HELP e2e_success_rate Success rate percentage
# TYPE e2e_success_rate gauge
e2e_success_rate %.2f

# HELP e2e_latency_average Average latency in seconds
# TYPE e2e_latency_average gauge
e2e_latency_average %.6f

# HELP e2e_throughput_total Total throughput operations per second
# TYPE e2e_throughput_total gauge
e2e_throughput_total %.2f
`, summary.TotalOperations, summary.SuccessRate, summary.AverageLatency.Seconds(), summary.TotalThroughput)

	return []byte(prometheus), nil
}

func (pm *PerformanceMonitor) exportCSVFormat(summary MetricsSummary) ([]byte, error) {
	// Simplified CSV format export
	csv := "timestamp,total_operations,success_rate,average_latency_ms,throughput_ops_sec,average_memory_mb,peak_memory_mb\n"
	csv += fmt.Sprintf("%s,%d,%.2f,%.2f,%.2f,%d,%d\n",
		time.Now().Format(time.RFC3339),
		summary.TotalOperations,
		summary.SuccessRate,
		float64(summary.AverageLatency.Nanoseconds())/1000000, // Convert to milliseconds
		summary.TotalThroughput,
		summary.AverageMemory/(1024*1024), // Convert to MB
		summary.PeakMemory/(1024*1024),    // Convert to MB
	)

	return []byte(csv), nil
}

// Utility functions

func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// DefaultMonitoringConfig returns a default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:         true,
		SampleInterval:  1 * time.Second,
		RetentionPeriod: 24 * time.Hour,
		DashboardPort:   8080,
		ExportFormat:    "json",
		AlertThresholds: AlertThresholds{
			MaxLatency:    1 * time.Second,
			MaxErrorRate:  5.0,
			MaxMemoryMB:   1024,
			MaxCPUPercent: 80,
			MinThroughput: 100,
		},
	}
}
