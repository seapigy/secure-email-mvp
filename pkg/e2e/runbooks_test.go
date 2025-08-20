package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunbookEngine(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.Procedures)
	assert.NotNil(t, engine.Executor)
	assert.NotNil(t, engine.Logger)
	assert.NotNil(t, engine.DB)

	// Check that built-in procedures are registered
	assert.Contains(t, engine.Procedures, "canary_rollout")
	assert.Contains(t, engine.Procedures, "emergency_rollback")
	assert.Contains(t, engine.Procedures, "key_rotation")
}

func TestRunbookEngine_RegisterProcedure(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	// Register a custom procedure
	procedure := Procedure{
		Name:        "test_procedure",
		Description: "Test procedure for unit testing",
		Steps: []Step{
			{
				ID:          "step1",
				Name:        "Test Step",
				Description: "A test step",
				Action:      "test_action",
				Timeout:     1 * time.Minute,
				Critical:    true,
			},
		},
		Timeout:  5 * time.Minute,
		Critical: false,
	}

	engine.RegisterProcedure(procedure)

	// Verify procedure was registered
	registered, exists := engine.Procedures["test_procedure"]
	assert.True(t, exists)
	assert.Equal(t, procedure.Name, registered.Name)
	assert.Equal(t, procedure.Description, registered.Description)
	assert.Len(t, registered.Steps, 1)
}

func TestRunbookEngine_ExecuteProcedure(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	// Create required tables
	var err error
	_, err = db.Exec(`
		CREATE TABLE runbook_executions (
			id TEXT PRIMARY KEY,
			runbook_name TEXT NOT NULL,
			procedure_name TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			error_message TEXT,
			execution_log TEXT,
			operator_id TEXT,
			correlation_id TEXT
		)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	// Test executing a built-in procedure
	result, err := engine.ExecuteProcedure("canary_rollout", "test_operator", "test_correlation")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "canary_rollout", result.ProcedureName)
	assert.Equal(t, "running", result.Status)
	assert.NotEmpty(t, result.ExecutionID)

	// Test executing non-existent procedure
	_, err = engine.ExecuteProcedure("non_existent", "test_operator", "test_correlation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "procedure 'non_existent' not found")
}

func TestRunbookEngine_GetExecutionStatus(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	// Create required tables
	var err error
	_, err = db.Exec(`
		CREATE TABLE runbook_executions (
			id TEXT PRIMARY KEY,
			runbook_name TEXT NOT NULL,
			procedure_name TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			error_message TEXT,
			execution_log TEXT,
			operator_id TEXT,
			correlation_id TEXT
		)
	`)
	require.NoError(t, err)

	// Insert test execution
	executionID := "test_execution_123"
	_, err = db.Exec(`
		INSERT INTO runbook_executions (id, runbook_name, procedure_name, status, started_at, operator_id, correlation_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, executionID, "test_runbook", "test_procedure", "completed", time.Now(), "test_operator", "test_correlation")
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	// Test getting execution status
	result, err := engine.GetExecutionStatus(executionID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, executionID, result.ExecutionID)
	assert.Equal(t, "test_procedure", result.ProcedureName)
	assert.Equal(t, "completed", result.Status)

	// Test getting non-existent execution
	_, err = engine.GetExecutionStatus("non_existent")
	assert.Error(t, err)
}

func TestRunbookEngine_ListProcedures(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	procedures := engine.ListProcedures()
	assert.NotEmpty(t, procedures)

	// Check that built-in procedures are included
	procedureNames := make([]string, 0, len(procedures))
	for _, p := range procedures {
		procedureNames = append(procedureNames, p.Name)
	}

	assert.Contains(t, procedureNames, "canary_rollout")
	assert.Contains(t, procedureNames, "emergency_rollback")
	assert.Contains(t, procedureNames, "key_rotation")
}

func TestProcedureExecutor_ExecuteAction(t *testing.T) {
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

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', FALSE, 0.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  &RunbookLogger{DB: db},
	}

	// Test validate_prerequisites action
	stepResult := &StepResult{
		Output: make(map[string]interface{}),
	}
	err = executor.ExecuteAction(context.Background(), "validate_prerequisites", nil, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, "connected", stepResult.Output["database_status"])
	assert.Equal(t, "available", stepResult.Output["metrics_status"])

	// Test enable_canary action
	stepResult = &StepResult{
		Output: make(map[string]interface{}),
	}
	parameters := map[string]interface{}{"traffic_percentage": 5.0}
	err = executor.ExecuteAction(context.Background(), "enable_canary", parameters, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, 5.0, stepResult.Output["traffic_percentage"])
	assert.Equal(t, "enabled", stepResult.Output["status"])

	// Test invalid action
	stepResult = &StepResult{
		Output: make(map[string]interface{}),
	}
	err = executor.ExecuteAction(context.Background(), "invalid_action", nil, stepResult)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action")
}

func TestProcedureExecutor_EnableCanary(t *testing.T) {
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

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', FALSE, 0.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  &RunbookLogger{DB: db},
	}

	// Test valid parameters
	stepResult := &StepResult{
		Output: make(map[string]interface{}),
	}
	parameters := map[string]interface{}{"traffic_percentage": 10.0}
	err = executor.enableCanary(context.Background(), parameters, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, 10.0, stepResult.Output["traffic_percentage"])

	// Test invalid parameters
	stepResult = &StepResult{
		Output: make(map[string]interface{}),
	}
	invalidParameters := map[string]interface{}{"invalid_param": "value"}
	err = executor.enableCanary(context.Background(), invalidParameters, stepResult)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid traffic_percentage parameter")
}

func TestProcedureExecutor_DisableCanary(t *testing.T) {
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

	// Insert test data with enabled canary
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', TRUE, 15.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  &RunbookLogger{DB: db},
	}

	// Test disable canary
	stepResult := &StepResult{
		Output: make(map[string]interface{}),
	}
	err = executor.disableCanary(context.Background(), nil, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, "disabled", stepResult.Output["status"])

	// Verify database was updated
	var enabled bool
	var trafficPercentage float64
	err = db.QueryRow("SELECT enabled, traffic_percentage FROM canary_config WHERE id = 'canary_main'").Scan(&enabled, &trafficPercentage)
	assert.NoError(t, err)
	assert.False(t, enabled)
	assert.Equal(t, 0.0, trafficPercentage)
}

func TestProcedureExecutor_IncreaseTraffic(t *testing.T) {
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

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold)
		VALUES ('canary_main', 'Main E2E Canary', TRUE, 5.0, '["beta_users"]', 5.0)
	`)
	require.NoError(t, err)

	metrics := &MetricsCollector{}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  &RunbookLogger{DB: db},
	}

	// Test valid traffic increase
	stepResult := &StepResult{
		Output: make(map[string]interface{}),
	}
	parameters := map[string]interface{}{"target_percentage": 25.0}
	err = executor.increaseTraffic(context.Background(), parameters, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, 25.0, stepResult.Output["new_traffic_percentage"])
	assert.Equal(t, "traffic_increased", stepResult.Output["status"])

	// Test invalid parameters
	stepResult = &StepResult{
		Output: make(map[string]interface{}),
	}
	invalidParameters := map[string]interface{}{"invalid_param": "value"}
	err = executor.increaseTraffic(context.Background(), invalidParameters, stepResult)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid target_percentage parameter")
}

func TestProcedureExecutor_ValidatePrerequisites(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  &RunbookLogger{DB: db},
	}

	// Test valid prerequisites
	stepResult := &StepResult{
		Output: make(map[string]interface{}),
	}
	var err error
	err = executor.validatePrerequisites(context.Background(), nil, stepResult)
	assert.NoError(t, err)
	assert.Equal(t, "completed", stepResult.Status)
	assert.Equal(t, "connected", stepResult.Output["database_status"])
	assert.Equal(t, "available", stepResult.Output["metrics_status"])

	// Test with nil metrics
	executor.Metrics = nil
	stepResult = &StepResult{
		Output: make(map[string]interface{}),
	}
	err = executor.validatePrerequisites(context.Background(), nil, stepResult)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metrics system not available")
}

func TestRunbookEngine_ValidateMetric(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	// Test valid metric validation
	validation := Validation{
		Name:        "Test Metric",
		Description: "Test metric validation",
		Type:        "metric",
		Parameters: map[string]interface{}{
			"metric":    "error_rate_percent",
			"threshold": 1.0,
		},
		Expected: "below_threshold",
		Timeout:  2 * time.Minute,
	}

	result := engine.validateMetric(context.Background(), validation)
	assert.True(t, result)

	// Test invalid parameters
	invalidValidation := Validation{
		Name:        "Test Metric",
		Description: "Test metric validation",
		Type:        "metric",
		Parameters: map[string]interface{}{
			"invalid_param": "value",
		},
		Expected: "below_threshold",
		Timeout:  2 * time.Minute,
	}

	result = engine.validateMetric(context.Background(), invalidValidation)
	assert.False(t, result)
}

func TestRunbookEngine_ValidateHTTP(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	validation := Validation{
		Name:        "Test HTTP",
		Description: "Test HTTP validation",
		Type:        "http",
		Parameters:  make(map[string]interface{}),
		Expected:    "success",
		Timeout:     2 * time.Minute,
	}

	result := engine.validateHTTP(context.Background(), validation)
	assert.True(t, result) // Placeholder implementation returns true
}

func TestRunbookEngine_ValidateDatabase(t *testing.T) {
	db := initTestDB(t)
	defer db.Close()

	metrics := &MetricsCollector{}
	engine := NewRunbookEngine(db, metrics)

	validation := Validation{
		Name:        "Test Database",
		Description: "Test database validation",
		Type:        "database",
		Parameters:  make(map[string]interface{}),
		Expected:    "success",
		Timeout:     2 * time.Minute,
	}

	result := engine.validateDatabase(context.Background(), validation)
	assert.True(t, result) // Placeholder implementation returns true
}
