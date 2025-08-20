package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCanaryRolloutManager(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 5.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, config.Enabled, manager.Config.Enabled)
	assert.Equal(t, config.TrafficPercentage, manager.Config.TrafficPercentage)
	assert.Equal(t, config.UserSegments, manager.Config.UserSegments)
	assert.Equal(t, config.RollbackThreshold, manager.Config.RollbackThreshold)
}

func TestCanaryRolloutManager_ShouldRouteToE2E(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 100.0, // 100% to ensure all users get routed
		UserSegments:      []string{"beta_users", "internal_testers"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test with user in allowed segments - should always be routed with 100% traffic
	result := manager.ShouldRouteToE2E("test_user_1", []string{"beta_users"})
	assert.True(t, result, "User should be routed to E2E with 100% traffic")

	// Test with user not in allowed segments
	result = manager.ShouldRouteToE2E("user456", []string{"regular_users"})
	assert.False(t, result)

	// Test with disabled canary
	manager.Config.Enabled = false
	result = manager.ShouldRouteToE2E("user123", []string{"beta_users"})
	assert.False(t, result)

	// Test with 0% traffic
	manager.Config.Enabled = true
	manager.Config.TrafficPercentage = 0.0
	result = manager.ShouldRouteToE2E("user123", []string{"beta_users"})
	assert.False(t, result)
}

func TestCanaryRolloutManager_UpdateTrafficPercentage(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	// Create canary_config table
	var err error
	_, err = db.Exec(`
		CREATE TABLE canary_config (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			traffic_percentage REAL DEFAULT 0.0,
			user_segments TEXT,
			rollback_threshold REAL DEFAULT 5.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', TRUE, 1.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 1.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test valid percentage update
	err = manager.UpdateTrafficPercentage(25.0)
	assert.NoError(t, err)
	assert.Equal(t, 25.0, manager.Config.TrafficPercentage)

	// Test invalid percentage (negative)
	err = manager.UpdateTrafficPercentage(-5.0)
	assert.Error(t, err)

	// Test invalid percentage (over 100)
	err = manager.UpdateTrafficPercentage(150.0)
	assert.Error(t, err)
}

func TestCanaryRolloutManager_TriggerRollback(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	// Create required tables
	var err error
	_, err = db.Exec(`
		CREATE TABLE canary_config (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			traffic_percentage REAL DEFAULT 0.0,
			user_segments TEXT,
			rollback_threshold REAL DEFAULT 5.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE rollback_events (
			id TEXT PRIMARY KEY,
			trigger_type TEXT NOT NULL,
			trigger_condition TEXT NOT NULL,
			rollback_reason TEXT,
			affected_users INTEGER,
			duration_seconds INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', TRUE, 10.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 10.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test rollback trigger
	err = manager.TriggerRollback("High error rate detected", "automatic")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, manager.Config.TrafficPercentage)

	// Verify rollback event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM rollback_events").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCanaryRolloutManager_GetRolloutStatus(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	// Create required tables
	var err error
	_, err = db.Exec(`
		CREATE TABLE canary_config (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			traffic_percentage REAL DEFAULT 0.0,
			user_segments TEXT,
			rollback_threshold REAL DEFAULT 5.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE rollback_events (
			id TEXT PRIMARY KEY,
			trigger_type TEXT NOT NULL,
			trigger_condition TEXT NOT NULL,
			rollback_reason TEXT,
			affected_users INTEGER,
			duration_seconds INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', TRUE, 15.0, '["beta_users", "internal_testers"]', 5.0)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO rollback_events (id, trigger_type, trigger_condition, rollback_reason, created_at)
		VALUES ('rollback1', 'automatic', 'error_rate_threshold', 'High error rate', CURRENT_TIMESTAMP)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 15.0,
		UserSegments:      []string{"beta_users", "internal_testers"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test getting rollout status
	status, err := manager.GetRolloutStatus()
	assert.NoError(t, err)
	assert.NotNil(t, status)

	// Verify status fields
	assert.Equal(t, true, status["enabled"])
	assert.Equal(t, 15.0, status["traffic_percentage"])
	assert.Equal(t, []string{"beta_users", "internal_testers"}, status["user_segments"])
	assert.Equal(t, 5.0, status["rollback_threshold"])

	// Verify rollback events
	rollbackEvents, ok := status["recent_rollbacks"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, rollbackEvents, 1)
	assert.Equal(t, "automatic", rollbackEvents[0]["trigger_type"])
}

func TestTrafficRouter_RouteRequest(t *testing.T) {
	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 20.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	router := NewTrafficRouter(config, metrics)

	// Test routing with user in allowed segments
	result := router.RouteRequest("user123", []string{"beta_users"})
	assert.Contains(t, []string{"legacy", "e2e"}, result)

	// Test routing with user not in allowed segments
	result = router.RouteRequest("user456", []string{"regular_users"})
	assert.Equal(t, "legacy", result)

	// Test routing with disabled canary
	router.Config.Enabled = false
	result = router.RouteRequest("user123", []string{"beta_users"})
	assert.Equal(t, "legacy", result)

	// Test routing with 0% traffic
	router.Config.Enabled = true
	router.Config.TrafficPercentage = 0.0
	result = router.RouteRequest("user123", []string{"beta_users"})
	assert.Equal(t, "legacy", result)
}

func TestABTestEngine_GetStatus(t *testing.T) {
	metrics := &MetricsCollector{}
	config := ABTestConfig{
		TestDuration:    24 * time.Hour,
		SampleSize:      10000,
		ConfidenceLevel: 0.95,
		MinSampleSize:   1000,
		SuccessCriteria: []Criterion{
			{MetricName: "response_time_ms", Operator: "lte", TargetValue: 300, Weight: 0.4, Critical: true},
		},
	}

	abTest := &ABTestEngine{
		Config:  config,
		Metrics: metrics,
		Results: nil, // Not started yet
	}

	// Test status when not started
	status := abTest.GetStatus()
	assert.Equal(t, "not_started", status["status"])

	// Test status when running - initialize Results first
	abTest.Results = &TestResults{
		LegacyMetrics: make(map[string]MetricStats),
		E2EMetrics:    make(map[string]MetricStats),
		Timestamp:     time.Now(),
		Decision: TestDecision{
			PromoteE2E:     false,
			Confidence:     0.95,
			Reason:         "Testing in progress",
			Recommendation: "Continue monitoring",
		},
	}

	status = abTest.GetStatus()
	assert.Equal(t, "running", status["status"])
	assert.NotNil(t, status["decision"])
	assert.NotNil(t, status["last_updated"])
}

func TestABTestEngine_EvaluateCriterion(t *testing.T) {
	metrics := &MetricsCollector{}
	config := ABTestConfig{
		TestDuration:    24 * time.Hour,
		SampleSize:      10000,
		ConfidenceLevel: 0.95,
		MinSampleSize:   1000,
	}

	abTest := &ABTestEngine{
		Config:  config,
		Metrics: metrics,
		Results: &TestResults{
			LegacyMetrics: make(map[string]MetricStats),
			E2EMetrics:    make(map[string]MetricStats),
		},
	}

	stats := MetricStats{
		Mean:       250.0,
		StdDev:     20.0,
		SampleSize: 1000,
	}

	// Test less than or equal criterion
	criterion := Criterion{MetricName: "response_time_ms", Operator: "lte", TargetValue: 300}
	result := abTest.evaluateCriterion(criterion, stats)
	assert.True(t, result)

	// Test greater than criterion
	criterion = Criterion{MetricName: "response_time_ms", Operator: "gt", TargetValue: 200}
	result = abTest.evaluateCriterion(criterion, stats)
	assert.True(t, result)

	// Test equal criterion
	criterion = Criterion{MetricName: "response_time_ms", Operator: "eq", TargetValue: 250}
	result = abTest.evaluateCriterion(criterion, stats)
	assert.True(t, result)

	// Test failed criterion
	criterion = Criterion{MetricName: "response_time_ms", Operator: "lte", TargetValue: 200}
	result = abTest.evaluateCriterion(criterion, stats)
	assert.False(t, result)

	// Test unknown operator
	criterion = Criterion{MetricName: "response_time_ms", Operator: "unknown", TargetValue: 200}
	result = abTest.evaluateCriterion(criterion, stats)
	assert.False(t, result)
}

func TestCanaryRolloutManager_StartStop(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           true,
		TrafficPercentage: 5.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test start
	err = manager.Start()
	assert.NoError(t, err)

	// Test stop
	err = manager.Stop()
	assert.NoError(t, err)
}

func TestCanaryRolloutManager_StartDisabled(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	config := CanaryConfig{
		Enabled:           false,
		TrafficPercentage: 5.0,
		UserSegments:      []string{"beta_users"},
		RollbackThreshold: 5.0,
	}

	manager, err := NewCanaryRolloutManager(config, db, metrics)
	require.NoError(t, err)

	// Test start with disabled config
	err = manager.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "canary rollout is not enabled")
}
