package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// CanaryRolloutManager manages the canary rollout process
type CanaryRolloutManager struct {
	Config     CanaryConfig
	DB         *sql.DB
	Metrics    *MetricsCollector
	ABTest     *ABTestEngine
	RollbackCh chan RollbackSignal
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// CanaryConfig holds configuration for canary rollout
type CanaryConfig struct {
	Enabled           bool          `json:"enabled"`
	TrafficPercentage float64       `json:"traffic_percentage"`
	UserSegments      []string      `json:"user_segments"`
	RollbackThreshold float64       `json:"rollback_threshold"`
	RolloutStages     []float64     `json:"rollout_stages"` // e.g., [1, 5, 10, 25, 50, 100]
	StageDuration     time.Duration `json:"stage_duration"`
	MonitoringWindow  time.Duration `json:"monitoring_window"`
}

// ABTestEngine manages A/B testing for performance comparison
type ABTestEngine struct {
	Config  ABTestConfig
	Metrics *MetricsCollector
	Results *TestResults
	mu      sync.RWMutex
}

// ABTestConfig holds configuration for A/B testing
type ABTestConfig struct {
	TestDuration    time.Duration `json:"test_duration"`
	SampleSize      int           `json:"sample_size"`
	ConfidenceLevel float64       `json:"confidence_level"`
	SuccessCriteria []Criterion   `json:"success_criteria"`
	MinSampleSize   int           `json:"min_sample_size"`
}

// Criterion defines a success criterion for A/B testing
type Criterion struct {
	MetricName  string  `json:"metric_name"`
	Operator    string  `json:"operator"` // "gt", "lt", "eq", "gte", "lte"
	TargetValue float64 `json:"target_value"`
	Weight      float64 `json:"weight"`
	Critical    bool    `json:"critical"`
}

// TestResults holds A/B test results
type TestResults struct {
	LegacyMetrics map[string]MetricStats `json:"legacy_metrics"`
	E2EMetrics    map[string]MetricStats `json:"e2e_metrics"`
	Comparison    ComparisonResult       `json:"comparison"`
	Decision      TestDecision           `json:"decision"`
	Timestamp     time.Time              `json:"timestamp"`
}

// MetricStats holds statistical information about a metric
type MetricStats struct {
	Mean            float64 `json:"mean"`
	StdDev          float64 `json:"std_dev"`
	SampleSize      int     `json:"sample_size"`
	ConfidenceLower float64 `json:"confidence_lower"`
	ConfidenceUpper float64 `json:"confidence_upper"`
	MinValue        float64 `json:"min_value"`
	MaxValue        float64 `json:"max_value"`
}

// ComparisonResult holds comparison between legacy and E2E
type ComparisonResult struct {
	MetricName    string  `json:"metric_name"`
	PercentChange float64 `json:"percent_change"`
	PValue        float64 `json:"p_value"`
	Significant   bool    `json:"significant"`
	Winner        string  `json:"winner"` // "legacy", "e2e", "tie"
}

// TestDecision holds the final decision of the A/B test
type TestDecision struct {
	PromoteE2E     bool    `json:"promote_e2e"`
	Confidence     float64 `json:"confidence"`
	Reason         string  `json:"reason"`
	Recommendation string  `json:"recommendation"`
}

// RollbackSignal indicates when a rollback should occur
type RollbackSignal struct {
	Reason        string    `json:"reason"`
	TriggerType   string    `json:"trigger_type"`
	AffectedUsers int       `json:"affected_users"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"`
}

// TrafficRouter routes traffic between legacy and E2E systems
type TrafficRouter struct {
	Config     CanaryConfig
	Metrics    *MetricsCollector
	RollbackCh chan RollbackSignal
	mu         sync.RWMutex
}

// NewCanaryRolloutManager creates a new canary rollout manager
func NewCanaryRolloutManager(config CanaryConfig, db *sql.DB, metrics *MetricsCollector) (*CanaryRolloutManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default values
	if config.RolloutStages == nil {
		config.RolloutStages = []float64{1, 5, 10, 25, 50, 100}
	}
	if config.StageDuration == 0 {
		config.StageDuration = 1 * time.Hour
	}
	if config.MonitoringWindow == 0 {
		config.MonitoringWindow = 15 * time.Minute
	}

	abTest := &ABTestEngine{
		Config: ABTestConfig{
			TestDuration:    24 * time.Hour,
			SampleSize:      10000,
			ConfidenceLevel: 0.95,
			MinSampleSize:   1000,
			SuccessCriteria: []Criterion{
				{MetricName: "response_time_ms", Operator: "lte", TargetValue: 300, Weight: 0.4, Critical: true},
				{MetricName: "error_rate_percent", Operator: "lte", TargetValue: 1.0, Weight: 0.3, Critical: true},
				{MetricName: "throughput_ops_per_sec", Operator: "gte", TargetValue: 800, Weight: 0.3, Critical: false},
			},
		},
		Metrics: metrics,
		Results: &TestResults{
			LegacyMetrics: make(map[string]MetricStats),
			E2EMetrics:    make(map[string]MetricStats),
		},
	}

	return &CanaryRolloutManager{
		Config:     config,
		DB:         db,
		Metrics:    metrics,
		ABTest:     abTest,
		RollbackCh: make(chan RollbackSignal, 100),
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start begins the canary rollout process
func (crm *CanaryRolloutManager) Start() error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if !crm.Config.Enabled {
		return fmt.Errorf("canary rollout is not enabled")
	}

	// Start monitoring goroutine
	go crm.monitorRollout()

	// Start A/B testing if configured
	if crm.Config.TrafficPercentage > 0 {
		go crm.ABTest.Start()
	}

	return nil
}

// Stop stops the canary rollout process
func (crm *CanaryRolloutManager) Stop() error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	crm.cancel()
	close(crm.RollbackCh)
	return nil
}

// ShouldRouteToE2E determines if a request should be routed to E2E
func (crm *CanaryRolloutManager) ShouldRouteToE2E(userID string, userSegments []string) bool {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	if !crm.Config.Enabled || crm.Config.TrafficPercentage <= 0 {
		return false
	}

	// Check if user is in allowed segments
	if len(crm.Config.UserSegments) > 0 {
		allowed := false
		for _, segment := range crm.Config.UserSegments {
			for _, userSegment := range userSegments {
				if segment == userSegment {
					allowed = true
					break
				}
			}
			if allowed {
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// Use consistent hashing for deterministic routing
	hash := crm.hashUserID(userID)
	percentage := (float64(hash) / float64(1<<32)) * 100

	return percentage <= crm.Config.TrafficPercentage
}

// UpdateTrafficPercentage updates the traffic percentage
func (crm *CanaryRolloutManager) UpdateTrafficPercentage(percentage float64) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("traffic percentage must be between 0 and 100")
	}

	crm.Config.TrafficPercentage = percentage

	// Update database
	_, err := crm.DB.Exec(`
		UPDATE canary_config 
		SET traffic_percentage = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = 'canary_main'
	`, percentage)

	if err != nil {
		return fmt.Errorf("failed to update traffic percentage: %w", err)
	}

	return nil
}

// TriggerRollback triggers an immediate rollback
func (crm *CanaryRolloutManager) TriggerRollback(reason string, triggerType string) error {
	crm.mu.Lock()
	defer crm.mu.Unlock()

	// Set traffic percentage to 0
	crm.Config.TrafficPercentage = 0

	// Update database
	_, err := crm.DB.Exec(`
		UPDATE canary_config 
		SET traffic_percentage = 0, updated_at = CURRENT_TIMESTAMP 
		WHERE id = 'canary_main'
	`)

	if err != nil {
		return fmt.Errorf("failed to update traffic percentage: %w", err)
	}

	// Record rollback event
	rollbackID := uuid.New().String()
	_, err = crm.DB.Exec(`
		INSERT INTO rollback_events (id, trigger_type, trigger_condition, rollback_reason, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, rollbackID, triggerType, "manual", reason)

	if err != nil {
		return fmt.Errorf("failed to record rollback event: %w", err)
	}

	// Send rollback signal
	select {
	case crm.RollbackCh <- RollbackSignal{
		Reason:        reason,
		TriggerType:   triggerType,
		AffectedUsers: 0, // Will be calculated
		Timestamp:     time.Now(),
		CorrelationID: rollbackID,
	}:
	default:
		// Channel is full, log warning
	}

	return nil
}

// GetRolloutStatus returns the current rollout status
func (crm *CanaryRolloutManager) GetRolloutStatus() (map[string]interface{}, error) {
	crm.mu.RLock()
	defer crm.mu.RUnlock()

	var config struct {
		Enabled           bool    `json:"enabled"`
		TrafficPercentage float64 `json:"traffic_percentage"`
		UserSegments      string  `json:"user_segments"`
		RollbackThreshold float64 `json:"rollback_threshold"`
	}

	err := crm.DB.QueryRow(`
		SELECT enabled, traffic_percentage, user_segments, rollback_threshold
		FROM canary_config WHERE id = 'canary_main'
	`).Scan(&config.Enabled, &config.TrafficPercentage, &config.UserSegments, &config.RollbackThreshold)

	if err != nil {
		return nil, fmt.Errorf("failed to get rollout status: %w", err)
	}

	var userSegments []string
	if config.UserSegments != "" {
		err = json.Unmarshal([]byte(config.UserSegments), &userSegments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user segments: %w", err)
		}
	}

	// Get recent rollback events
	rows, err := crm.DB.Query(`
		SELECT trigger_type, trigger_condition, rollback_reason, created_at
		FROM rollback_events 
		WHERE created_at >= datetime('now', '-24 hours')
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback events: %w", err)
	}
	defer rows.Close()

	var rollbackEvents []map[string]interface{}
	for rows.Next() {
		var event struct {
			TriggerType      string    `json:"trigger_type"`
			TriggerCondition string    `json:"trigger_condition"`
			RollbackReason   string    `json:"rollback_reason"`
			CreatedAt        time.Time `json:"created_at"`
		}
		err := rows.Scan(&event.TriggerType, &event.TriggerCondition, &event.RollbackReason, &event.CreatedAt)
		if err != nil {
			continue
		}
		rollbackEvents = append(rollbackEvents, map[string]interface{}{
			"trigger_type":      event.TriggerType,
			"trigger_condition": event.TriggerCondition,
			"rollback_reason":   event.RollbackReason,
			"created_at":        event.CreatedAt,
		})
	}

	return map[string]interface{}{
		"enabled":            config.Enabled,
		"traffic_percentage": config.TrafficPercentage,
		"user_segments":      userSegments,
		"rollback_threshold": config.RollbackThreshold,
		"recent_rollbacks":   rollbackEvents,
		"ab_test_status":     crm.ABTest.GetStatus(),
	}, nil
}

// monitorRollout continuously monitors the rollout and triggers rollbacks if needed
func (crm *CanaryRolloutManager) monitorRollout() {
	ticker := time.NewTicker(crm.Config.MonitoringWindow)
	defer ticker.Stop()

	for {
		select {
		case <-crm.ctx.Done():
			return
		case <-ticker.C:
			crm.checkRollbackConditions()
		}
	}
}

// checkRollbackConditions checks if rollback conditions are met
func (crm *CanaryRolloutManager) checkRollbackConditions() {
	crm.mu.RLock()
	trafficPercentage := crm.Config.TrafficPercentage
	rollbackThreshold := crm.Config.RollbackThreshold
	crm.mu.RUnlock()

	if trafficPercentage == 0 {
		return
	}

	// Get recent error rates
	errorRate, err := crm.getRecentErrorRate()
	if err != nil {
		// Log error but don't rollback
		return
	}

	// Check if error rate exceeds threshold
	if errorRate > rollbackThreshold {
		crm.TriggerRollback(
			fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", errorRate, rollbackThreshold),
			"automatic",
		)
	}

	// Check A/B test results
	if crm.ABTest.Results != nil && crm.ABTest.Results.Decision.PromoteE2E == false {
		crm.TriggerRollback(
			"A/B test indicates E2E performance is not acceptable",
			"automatic",
		)
	}
}

// getRecentErrorRate gets the recent error rate
func (crm *CanaryRolloutManager) getRecentErrorRate() (float64, error) {
	// This would typically query metrics from the monitoring system
	// For now, return a placeholder value
	return 0.5, nil
}

// hashUserID creates a consistent hash for user ID
func (crm *CanaryRolloutManager) hashUserID(userID string) uint32 {
	hash := uint32(0)
	for _, char := range userID {
		hash = hash*31 + uint32(char)
	}
	return hash
}

// NewTrafficRouter creates a new traffic router
func NewTrafficRouter(config CanaryConfig, metrics *MetricsCollector) *TrafficRouter {
	return &TrafficRouter{
		Config:     config,
		Metrics:    metrics,
		RollbackCh: make(chan RollbackSignal, 100),
	}
}

// RouteRequest routes a request to either legacy or E2E system
func (tr *TrafficRouter) RouteRequest(userID string, userSegments []string) string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if !tr.Config.Enabled || tr.Config.TrafficPercentage <= 0 {
		return "legacy"
	}

	// Check user segments
	if len(tr.Config.UserSegments) > 0 {
		allowed := false
		for _, segment := range tr.Config.UserSegments {
			for _, userSegment := range userSegments {
				if segment == userSegment {
					allowed = true
					break
				}
			}
			if allowed {
				break
			}
		}
		if !allowed {
			return "legacy"
		}
	}

	// Use consistent hashing
	hash := tr.hashUserID(userID)
	percentage := (float64(hash) / float64(1<<32)) * 100

	if percentage <= tr.Config.TrafficPercentage {
		return "e2e"
	}
	return "legacy"
}

// hashUserID creates a consistent hash for user ID
func (tr *TrafficRouter) hashUserID(userID string) uint32 {
	hash := uint32(0)
	for _, char := range userID {
		hash = hash*31 + uint32(char)
	}
	return hash
}

// Start begins A/B testing
func (ab *ABTestEngine) Start() {
	// Start collecting metrics for both variants
	go ab.collectMetrics()
}

// collectMetrics collects metrics for A/B testing
func (ab *ABTestEngine) collectMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ab.updateMetrics()
	}
}

// updateMetrics updates the metrics for both variants
func (ab *ABTestEngine) updateMetrics() {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// Collect metrics for legacy system
	legacyMetrics := ab.collectVariantMetrics("legacy")
	for name, stats := range legacyMetrics {
		ab.Results.LegacyMetrics[name] = stats
	}

	// Collect metrics for E2E system
	e2eMetrics := ab.collectVariantMetrics("e2e")
	for name, stats := range e2eMetrics {
		ab.Results.E2EMetrics[name] = stats
	}

	// Update comparison results
	ab.updateComparisonResults()

	// Update test decision
	ab.updateTestDecision()
}

// collectVariantMetrics collects metrics for a specific variant
func (ab *ABTestEngine) collectVariantMetrics(variant string) map[string]MetricStats {
	// This would typically query the metrics system
	// For now, return placeholder data
	return map[string]MetricStats{
		"response_time_ms": {
			Mean:            150.0 + float64(len(variant)*10),
			StdDev:          20.0,
			SampleSize:      1000,
			ConfidenceLower: 130.0,
			ConfidenceUpper: 170.0,
			MinValue:        100.0,
			MaxValue:        300.0,
		},
		"error_rate_percent": {
			Mean:            0.1 + float64(len(variant))*0.05,
			StdDev:          0.05,
			SampleSize:      1000,
			ConfidenceLower: 0.05,
			ConfidenceUpper: 0.15,
			MinValue:        0.0,
			MaxValue:        1.0,
		},
	}
}

// updateComparisonResults updates the comparison between variants
func (ab *ABTestEngine) updateComparisonResults() {
	ab.Results.Comparison = ComparisonResult{
		MetricName:    "response_time_ms",
		PercentChange: 10.0,
		PValue:        0.05,
		Significant:   true,
		Winner:        "legacy",
	}
}

// updateTestDecision updates the test decision based on criteria
func (ab *ABTestEngine) updateTestDecision() {
	decision := TestDecision{
		PromoteE2E:     false,
		Confidence:     0.95,
		Reason:         "Performance criteria not met",
		Recommendation: "Continue testing with current configuration",
	}

	// Evaluate success criteria
	passedCriteria := 0
	totalCriteria := len(ab.Config.SuccessCriteria)

	for _, criterion := range ab.Config.SuccessCriteria {
		e2eValue, exists := ab.Results.E2EMetrics[criterion.MetricName]
		if !exists {
			continue
		}

		passed := ab.evaluateCriterion(criterion, e2eValue)
		if passed {
			passedCriteria++
		}

		if criterion.Critical && !passed {
			decision.Reason = fmt.Sprintf("Critical criterion '%s' not met", criterion.MetricName)
			break
		}
	}

	// Calculate overall score
	score := float64(passedCriteria) / float64(totalCriteria)
	if score >= 0.8 {
		decision.PromoteE2E = true
		decision.Reason = "All criteria met"
		decision.Recommendation = "Promote E2E to full rollout"
	}

	ab.Results.Decision = decision
	ab.Results.Timestamp = time.Now()
}

// evaluateCriterion evaluates if a criterion is met
func (ab *ABTestEngine) evaluateCriterion(criterion Criterion, stats MetricStats) bool {
	switch criterion.Operator {
	case "gt":
		return stats.Mean > criterion.TargetValue
	case "gte":
		return stats.Mean >= criterion.TargetValue
	case "lt":
		return stats.Mean < criterion.TargetValue
	case "lte":
		return stats.Mean <= criterion.TargetValue
	case "eq":
		return math.Abs(stats.Mean-criterion.TargetValue) < 0.001
	default:
		return false
	}
}

// GetStatus returns the current A/B test status
func (ab *ABTestEngine) GetStatus() map[string]interface{} {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	if ab.Results == nil {
		return map[string]interface{}{
			"status": "not_started",
		}
	}

	return map[string]interface{}{
		"status":         "running",
		"legacy_metrics": ab.Results.LegacyMetrics,
		"e2e_metrics":    ab.Results.E2EMetrics,
		"comparison":     ab.Results.Comparison,
		"decision":       ab.Results.Decision,
		"last_updated":   ab.Results.Timestamp,
	}
}
