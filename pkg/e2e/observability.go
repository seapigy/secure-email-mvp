package e2e

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// AdvancedObservabilityConfig defines configuration for advanced observability and monitoring
type AdvancedObservabilityConfig struct {
	Enabled           bool              `json:"enabled"`
	TracingEnabled    bool              `json:"tracing_enabled"`
	MetricsEnabled    bool              `json:"metrics_enabled"`
	AlertingEnabled   bool              `json:"alerting_enabled"`
	AnomalyDetection  bool              `json:"anomaly_detection"`
	LogLevel          string            `json:"log_level"`
	SamplingRate      float64           `json:"sampling_rate"`
	RetentionPeriod   time.Duration     `json:"retention_period"`
	ExportEndpoints   []ExportEndpoint  `json:"export_endpoints"`
	RedactionRules    []RedactionRule   `json:"redaction_rules"`
	CorrelationConfig CorrelationConfig `json:"correlation_config"`
	Metadata          map[string]string `json:"metadata,omitempty"`
}

// ExportEndpoint defines an observability data export endpoint
type ExportEndpoint struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // "jaeger", "prometheus", "elasticsearch", "datadog"
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Enabled  bool              `json:"enabled"`
}

// RedactionRule defines rules for redacting sensitive data in logs and traces
type RedactionRule struct {
	Name        string   `json:"name"`
	Pattern     string   `json:"pattern"`
	Replacement string   `json:"replacement"`
	Fields      []string `json:"fields,omitempty"`
}

// CorrelationConfig defines correlation ID configuration
type CorrelationConfig struct {
	Enabled       bool   `json:"enabled"`
	HeaderName    string `json:"header_name"`
	IDLength      int    `json:"id_length"`
	PropagateMode string `json:"propagate_mode"` // "always", "on_demand", "disabled"
}

// DistributedTracer provides distributed tracing capabilities
type DistributedTracer struct {
	config     AdvancedObservabilityConfig
	spans      map[string]*Span
	traces     map[string]*Trace
	mutex      sync.RWMutex
	exporter   TraceExporter
	redactor   *DataRedactor
	correlator *CorrelationManager
}

// Span represents a single operation in a distributed trace
type Span struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentID      string                 `json:"parent_id,omitempty"`
	OperationName string                 `json:"operation_name"`
	ServiceName   string                 `json:"service_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Status        SpanStatus             `json:"status"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
	Logs          []SpanLog              `json:"logs,omitempty"`
	Baggage       map[string]string      `json:"baggage,omitempty"`
}

// Trace represents a complete distributed trace
type Trace struct {
	TraceID       string        `json:"trace_id"`
	Spans         []*Span       `json:"spans"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       *time.Time    `json:"end_time,omitempty"`
	Duration      time.Duration `json:"duration"`
	ServiceCount  int           `json:"service_count"`
	ErrorCount    int           `json:"error_count"`
	CorrelationID string        `json:"correlation_id,omitempty"`
}

// SpanStatus represents the status of a span
type SpanStatus struct {
	Code    SpanStatusCode `json:"code"`
	Message string         `json:"message,omitempty"`
}

// SpanStatusCode represents span status codes
type SpanStatusCode int

const (
	SpanStatusOK SpanStatusCode = iota
	SpanStatusError
	SpanStatusTimeout
	SpanStatusCancelled
)

// SpanLog represents a log entry within a span
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// AdvancedAlertManager manages security and performance alerts
type AdvancedAlertManager struct {
	config        AdvancedObservabilityConfig
	rules         []*AlertRule
	notifications chan Alert
	db            *sql.DB
	mutex         sync.RWMutex
	processors    map[string]AlertProcessor
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Severity      AlertSeverity          `json:"severity"`
	Condition     AlertCondition         `json:"condition"`
	Actions       []AlertAction          `json:"actions"`
	Enabled       bool                   `json:"enabled"`
	Cooldown      time.Duration          `json:"cooldown"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertCondition defines the condition for triggering an alert
type AlertCondition struct {
	Metric    string        `json:"metric"`
	Operator  string        `json:"operator"` // "gt", "lt", "eq", "ne", "contains"
	Threshold interface{}   `json:"threshold"`
	Window    time.Duration `json:"window"`
	Query     string        `json:"query,omitempty"`
}

// AlertAction defines actions to take when an alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"` // "email", "webhook", "sms", "slack"
	Target     string                 `json:"target"`
	Template   string                 `json:"template,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Alert represents a triggered alert
type Alert struct {
	ID            string                 `json:"id"`
	RuleID        string                 `json:"rule_id"`
	Severity      AlertSeverity          `json:"severity"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Timestamp     time.Time              `json:"timestamp"`
	Value         interface{}            `json:"value"`
	Tags          map[string]string      `json:"tags,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Status        AlertStatus            `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// AnomalyDetector detects behavioral anomalies using ML techniques
type AnomalyDetector struct {
	config    AdvancedObservabilityConfig
	models    map[string]*AnomalyModel
	baselines map[string]*Baseline
	detector  *MLDetector
	mutex     sync.RWMutex
}

// AnomalyModel represents a machine learning model for anomaly detection
type AnomalyModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // "statistical", "ml", "threshold"
	Algorithm    string                 `json:"algorithm"`
	Parameters   map[string]interface{} `json:"parameters"`
	TrainingData []DataPoint            `json:"training_data"`
	Accuracy     float64                `json:"accuracy"`
	LastTrained  time.Time              `json:"last_trained"`
}

// Baseline represents normal behavior patterns
type Baseline struct {
	Metric      string             `json:"metric"`
	Mean        float64            `json:"mean"`
	StdDev      float64            `json:"std_dev"`
	Min         float64            `json:"min"`
	Max         float64            `json:"max"`
	Percentiles map[int]float64    `json:"percentiles"`
	Seasonality map[string]float64 `json:"seasonality,omitempty"`
	LastUpdated time.Time          `json:"last_updated"`
}

// DataPoint represents a single data point for analysis
type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Severity      float64                `json:"severity"`
	Metric        string                 `json:"metric"`
	Value         float64                `json:"value"`
	Expected      float64                `json:"expected"`
	Deviation     float64                `json:"deviation"`
	Timestamp     time.Time              `json:"timestamp"`
	Description   string                 `json:"description"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewDistributedTracer creates a new distributed tracer
func NewDistributedTracer(config AdvancedObservabilityConfig) (*DistributedTracer, error) {
	if !config.Enabled || !config.TracingEnabled {
		return &DistributedTracer{config: config}, nil
	}

	exporter, err := createTraceExporter(config.ExportEndpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	redactor := NewDataRedactor(config.RedactionRules)
	correlator := NewCorrelationManager(config.CorrelationConfig)

	tracer := &DistributedTracer{
		config:     config,
		spans:      make(map[string]*Span),
		traces:     make(map[string]*Trace),
		exporter:   exporter,
		redactor:   redactor,
		correlator: correlator,
	}

	return tracer, nil
}

// StartSpan starts a new span
func (dt *DistributedTracer) StartSpan(ctx context.Context, operationName, serviceName string) (*Span, context.Context) {
	if !dt.config.Enabled || !dt.config.TracingEnabled {
		return nil, ctx
	}

	// Generate IDs
	traceID := dt.correlator.GetOrCreateTraceID(ctx)
	spanID := generateSpanID()

	span := &Span{
		TraceID:       traceID,
		SpanID:        spanID,
		OperationName: operationName,
		ServiceName:   serviceName,
		StartTime:     time.Now(),
		Status:        SpanStatus{Code: SpanStatusOK},
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		Baggage:       make(map[string]string),
	}

	// Set parent if exists
	if parentSpan := dt.getActiveSpan(ctx); parentSpan != nil {
		span.ParentID = parentSpan.SpanID
	}

	dt.mutex.Lock()
	dt.spans[spanID] = span
	dt.mutex.Unlock()

	// Add span to context
	ctx = dt.setActiveSpan(ctx, span)

	return span, ctx
}

// FinishSpan completes a span
func (dt *DistributedTracer) FinishSpan(span *Span) {
	if span == nil {
		return
	}

	now := time.Now()
	span.EndTime = &now
	span.Duration = now.Sub(span.StartTime)

	// Redact sensitive data
	if dt.redactor != nil {
		dt.redactor.RedactSpan(span)
	}

	// Export span
	if dt.exporter != nil {
		dt.exporter.ExportSpan(span)
	}

	// Update trace
	dt.updateTrace(span)
}

// AddSpanTag adds a tag to a span
func (dt *DistributedTracer) AddSpanTag(span *Span, key string, value interface{}) {
	if span != nil && span.Tags != nil {
		span.Tags[key] = value
	}
}

// LogSpanEvent logs an event to a span
func (dt *DistributedTracer) LogSpanEvent(span *Span, level, message string, fields map[string]interface{}) {
	if span == nil {
		return
	}

	log := SpanLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	span.Logs = append(span.Logs, log)
}

// NewAdvancedAlertManager creates a new advanced alert manager
func NewAdvancedAlertManager(config AdvancedObservabilityConfig) (*AdvancedAlertManager, error) {
	if !config.Enabled || !config.AlertingEnabled {
		return &AdvancedAlertManager{config: config}, nil
	}

	// Initialize database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create alert database: %w", err)
	}

	if err := createAlertTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create alert tables: %w", err)
	}

	manager := &AdvancedAlertManager{
		config:        config,
		rules:         make([]*AlertRule, 0),
		notifications: make(chan Alert, 100),
		db:            db,
		processors:    make(map[string]AlertProcessor),
	}

	// Start alert processing
	go manager.processAlerts()

	// Load default alert rules
	manager.loadDefaultRules()

	return manager, nil
}

// AddRule adds a new alert rule
func (am *AdvancedAlertManager) AddRule(rule *AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.rules = append(am.rules, rule)
	return am.saveRuleToDB(rule)
}

// TriggerAlert triggers an alert based on a rule
func (am *AdvancedAlertManager) TriggerAlert(ruleID string, value interface{}, correlationID string) error {
	rule := am.findRule(ruleID)
	if rule == nil {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	// Check cooldown
	if rule.LastTriggered != nil && time.Since(*rule.LastTriggered) < rule.Cooldown {
		return nil // Skip due to cooldown
	}

	alert := Alert{
		ID:            generateAdvancedAlertID(),
		RuleID:        ruleID,
		Severity:      rule.Severity,
		Title:         rule.Name,
		Description:   rule.Description,
		Timestamp:     time.Now(),
		Value:         value,
		CorrelationID: correlationID,
		Status:        AlertStatusActive,
		Tags:          make(map[string]string),
		Metadata:      make(map[string]interface{}),
	}

	// Send alert
	select {
	case am.notifications <- alert:
		now := time.Now()
		rule.LastTriggered = &now
	default:
		return fmt.Errorf("alert queue full")
	}

	return nil
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(config AdvancedObservabilityConfig) (*AnomalyDetector, error) {
	if !config.Enabled || !config.AnomalyDetection {
		return &AnomalyDetector{config: config}, nil
	}

	detector := &AnomalyDetector{
		config:    config,
		models:    make(map[string]*AnomalyModel),
		baselines: make(map[string]*Baseline),
		detector:  NewMLDetector(),
	}

	// Initialize default models
	detector.initializeDefaultModels()

	return detector, nil
}

// DetectAnomalies detects anomalies in the provided data points
func (ad *AnomalyDetector) DetectAnomalies(metric string, dataPoints []DataPoint) ([]*Anomaly, error) {
	if !ad.config.Enabled || !ad.config.AnomalyDetection {
		return nil, nil
	}

	model, exists := ad.models[metric]
	if !exists {
		return nil, fmt.Errorf("no model found for metric: %s", metric)
	}

	baseline, exists := ad.baselines[metric]
	if !exists {
		return nil, fmt.Errorf("no baseline found for metric: %s", metric)
	}

	var anomalies []*Anomaly
	for _, point := range dataPoints {
		if ad.isAnomaly(point, baseline, model) {
			anomaly := &Anomaly{
				ID:          generateAnomalyID(),
				Type:        "statistical",
				Metric:      metric,
				Value:       point.Value,
				Expected:    baseline.Mean,
				Deviation:   abs(point.Value - baseline.Mean),
				Timestamp:   point.Timestamp,
				Description: fmt.Sprintf("Anomalous value detected for %s", metric),
				Metadata:    point.Metadata,
			}

			// Calculate severity
			anomaly.Severity = ad.calculateSeverity(point.Value, baseline)
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies, nil
}

// Helper methods and interfaces

// TraceExporter exports traces to external systems
type TraceExporter interface {
	ExportSpan(span *Span) error
	ExportTrace(trace *Trace) error
	Close() error
}

// AlertProcessor processes triggered alerts
type AlertProcessor interface {
	ProcessAlert(alert Alert) error
}

// MLDetector provides machine learning-based anomaly detection
type MLDetector struct {
	// Simplified ML detector implementation
}

// DataRedactor redacts sensitive data from spans and logs
type DataRedactor struct {
	rules      []*RedactionRule
	regexCache map[string]*regexp.Regexp
	mutex      sync.RWMutex
}

// CorrelationManager manages correlation IDs across requests
type CorrelationManager struct {
	config CorrelationConfig
}

// Implementations

func NewDataRedactor(rules []RedactionRule) *DataRedactor {
	redactor := &DataRedactor{
		rules:      make([]*RedactionRule, len(rules)),
		regexCache: make(map[string]*regexp.Regexp),
	}

	for i, rule := range rules {
		redactor.rules[i] = &rule
	}

	return redactor
}

func (dr *DataRedactor) RedactSpan(span *Span) {
	// Redact span tags
	for key, value := range span.Tags {
		if str, ok := value.(string); ok {
			span.Tags[key] = dr.redactString(str)
		}
	}

	// Redact span logs
	for i := range span.Logs {
		span.Logs[i].Message = dr.redactString(span.Logs[i].Message)
		for key, value := range span.Logs[i].Fields {
			if str, ok := value.(string); ok {
				span.Logs[i].Fields[key] = dr.redactString(str)
			}
		}
	}
}

func (dr *DataRedactor) redactString(input string) string {
	output := input
	for _, rule := range dr.rules {
		regex := dr.getRegex(rule.Pattern)
		if regex != nil {
			output = regex.ReplaceAllString(output, rule.Replacement)
		}
	}
	return output
}

func (dr *DataRedactor) getRegex(pattern string) *regexp.Regexp {
	dr.mutex.RLock()
	if regex, exists := dr.regexCache[pattern]; exists {
		dr.mutex.RUnlock()
		return regex
	}
	dr.mutex.RUnlock()

	dr.mutex.Lock()
	defer dr.mutex.Unlock()

	// Double-check after acquiring write lock
	if regex, exists := dr.regexCache[pattern]; exists {
		return regex
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	dr.regexCache[pattern] = regex
	return regex
}

func NewCorrelationManager(config CorrelationConfig) *CorrelationManager {
	return &CorrelationManager{config: config}
}

func (cm *CorrelationManager) GetOrCreateTraceID(ctx context.Context) string {
	if !cm.config.Enabled {
		return generateTraceID()
	}

	// In a real implementation, extract from context or headers
	return generateTraceID()
}

func NewMLDetector() *MLDetector {
	return &MLDetector{}
}

// Helper functions and implementations

func (dt *DistributedTracer) updateTrace(span *Span) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	trace, exists := dt.traces[span.TraceID]
	if !exists {
		trace = &Trace{
			TraceID:   span.TraceID,
			Spans:     make([]*Span, 0),
			StartTime: span.StartTime,
		}
		dt.traces[span.TraceID] = trace
	}

	trace.Spans = append(trace.Spans, span)

	// Update trace statistics
	if span.Status.Code == SpanStatusError {
		trace.ErrorCount++
	}

	if span.EndTime != nil {
		if trace.EndTime == nil || span.EndTime.After(*trace.EndTime) {
			trace.EndTime = span.EndTime
			if trace.EndTime != nil {
				trace.Duration = trace.EndTime.Sub(trace.StartTime)
			}
		}
	}
}

func (dt *DistributedTracer) getActiveSpan(ctx context.Context) *Span {
	// In a real implementation, extract from context
	return nil
}

func (dt *DistributedTracer) setActiveSpan(ctx context.Context, span *Span) context.Context {
	// In a real implementation, add to context
	return ctx
}

func (am *AdvancedAlertManager) processAlerts() {
	for alert := range am.notifications {
		am.handleAlert(alert)
	}
}

func (am *AdvancedAlertManager) handleAlert(alert Alert) {
	// Save to database
	am.saveAlertToDB(alert)

	// Process with registered processors
	for _, processor := range am.processors {
		go processor.ProcessAlert(alert)
	}
}

func (am *AdvancedAlertManager) findRule(ruleID string) *AlertRule {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	for _, rule := range am.rules {
		if rule.ID == ruleID {
			return rule
		}
	}
	return nil
}

func (am *AdvancedAlertManager) loadDefaultRules() {
	// Load default security and performance alert rules
	defaultRules := []*AlertRule{
		{
			ID:          "high_error_rate",
			Name:        "High Error Rate",
			Description: "Error rate exceeds threshold",
			Severity:    AlertSeverityHigh,
			Condition: AlertCondition{
				Metric:    "error_rate",
				Operator:  "gt",
				Threshold: 0.05,
				Window:    5 * time.Minute,
			},
			Enabled:  true,
			Cooldown: 10 * time.Minute,
		},
		{
			ID:          "authentication_failures",
			Name:        "Multiple Authentication Failures",
			Description: "Multiple failed authentication attempts detected",
			Severity:    AlertSeverityCritical,
			Condition: AlertCondition{
				Metric:    "auth_failures",
				Operator:  "gt",
				Threshold: 10,
				Window:    1 * time.Minute,
			},
			Enabled:  true,
			Cooldown: 5 * time.Minute,
		},
	}

	for _, rule := range defaultRules {
		am.AddRule(rule)
	}
}

func (ad *AnomalyDetector) initializeDefaultModels() {
	// Initialize default statistical models
	models := []*AnomalyModel{
		{
			ID:        "response_time_model",
			Name:      "Response Time Anomaly Detection",
			Type:      "statistical",
			Algorithm: "z_score",
			Parameters: map[string]interface{}{
				"threshold": 3.0,
				"window":    "5m",
			},
			LastTrained: time.Now(),
		},
		{
			ID:        "throughput_model",
			Name:      "Throughput Anomaly Detection",
			Type:      "statistical",
			Algorithm: "iqr",
			Parameters: map[string]interface{}{
				"multiplier": 1.5,
				"window":     "10m",
			},
			LastTrained: time.Now(),
		},
	}

	for _, model := range models {
		ad.models[model.ID] = model
	}
}

func (ad *AnomalyDetector) isAnomaly(point DataPoint, baseline *Baseline, model *AnomalyModel) bool {
	// Simplified anomaly detection using z-score
	if baseline.StdDev == 0 {
		return false
	}

	zScore := abs(point.Value-baseline.Mean) / baseline.StdDev
	threshold := 3.0 // Default threshold

	if thresholdVal, exists := model.Parameters["threshold"]; exists {
		if t, ok := thresholdVal.(float64); ok {
			threshold = t
		}
	}

	return zScore > threshold
}

func (ad *AnomalyDetector) calculateSeverity(value float64, baseline *Baseline) float64 {
	if baseline.StdDev == 0 {
		return 0.5
	}

	zScore := abs(value-baseline.Mean) / baseline.StdDev
	return minFloat(1.0, zScore/5.0) // Normalize to 0-1 range
}

// Database operations and utility functions

func createTraceExporter(endpoints []ExportEndpoint) (TraceExporter, error) {
	// Create mock exporter for testing
	return &MockTraceExporter{}, nil
}

func createAlertTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS alert_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		severity TEXT NOT NULL,
		condition TEXT NOT NULL,
		actions TEXT NOT NULL,
		enabled BOOLEAN DEFAULT true,
		cooldown INTEGER DEFAULT 0,
		last_triggered DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		rule_id TEXT NOT NULL,
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		timestamp DATETIME NOT NULL,
		value TEXT,
		correlation_id TEXT,
		status TEXT DEFAULT 'active',
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (rule_id) REFERENCES alert_rules (id)
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp);
	CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
	CREATE INDEX IF NOT EXISTS idx_alerts_correlation_id ON alerts(correlation_id);
	`

	_, err := db.Exec(schema)
	return err
}

func (am *AdvancedAlertManager) saveRuleToDB(rule *AlertRule) error {
	conditionJSON, _ := json.Marshal(rule.Condition)
	actionsJSON, _ := json.Marshal(rule.Actions)

	_, err := am.db.Exec(`
		INSERT OR REPLACE INTO alert_rules 
		(id, name, description, severity, condition, actions, enabled, cooldown, last_triggered)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Name, rule.Description, string(rule.Severity),
		string(conditionJSON), string(actionsJSON), rule.Enabled,
		int64(rule.Cooldown.Seconds()), rule.LastTriggered)

	return err
}

func (am *AdvancedAlertManager) saveAlertToDB(alert Alert) error {
	valueJSON, _ := json.Marshal(alert.Value)
	metadataJSON, _ := json.Marshal(alert.Metadata)

	_, err := am.db.Exec(`
		INSERT INTO alerts 
		(id, rule_id, severity, title, description, timestamp, value, correlation_id, status, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		alert.ID, alert.RuleID, string(alert.Severity), alert.Title,
		alert.Description, alert.Timestamp, string(valueJSON),
		alert.CorrelationID, string(alert.Status), string(metadataJSON))

	return err
}

// Mock implementations

type MockTraceExporter struct{}

func (mte *MockTraceExporter) ExportSpan(span *Span) error    { return nil }
func (mte *MockTraceExporter) ExportTrace(trace *Trace) error { return nil }
func (mte *MockTraceExporter) Close() error                   { return nil }

// ID generation functions

func generateSpanID() string {
	id := make([]byte, 8)
	rand.Read(id)
	return fmt.Sprintf("%x", id)
}

func generateTraceID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("%x", id)
}

func generateAdvancedAlertID() string {
	id := make([]byte, 8)
	rand.Read(id)
	return fmt.Sprintf("advanced_alert_%x", id)
}

func generateAnomalyID() string {
	id := make([]byte, 8)
	rand.Read(id)
	return fmt.Sprintf("anomaly_%x", id)
}

// Utility functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// DefaultAdvancedObservabilityConfig returns a default advanced observability configuration
func DefaultAdvancedObservabilityConfig() AdvancedObservabilityConfig {
	return AdvancedObservabilityConfig{
		Enabled:          false, // Disabled by default
		TracingEnabled:   true,
		MetricsEnabled:   true,
		AlertingEnabled:  true,
		AnomalyDetection: false,
		LogLevel:         "info",
		SamplingRate:     0.1,
		RetentionPeriod:  24 * time.Hour,
		ExportEndpoints:  make([]ExportEndpoint, 0),
		RedactionRules: []RedactionRule{
			{
				Name:        "email_redaction",
				Pattern:     `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
				Replacement: "[EMAIL_REDACTED]",
			},
			{
				Name:        "key_redaction",
				Pattern:     `(?i)(key|token|secret|password)[\s]*[=:][\s]*[\w\-/+]+`,
				Replacement: "[KEY_REDACTED]",
			},
		},
		CorrelationConfig: CorrelationConfig{
			Enabled:       true,
			HeaderName:    "X-Correlation-ID",
			IDLength:      16,
			PropagateMode: "always",
		},
		Metadata: make(map[string]string),
	}
}
