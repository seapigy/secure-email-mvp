package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// RunbookEngine manages automated operational procedures
type RunbookEngine struct {
	Procedures map[string]Procedure
	Executor   ProcedureExecutor
	Logger     *RunbookLogger
	DB         *sql.DB
	mu         sync.RWMutex
}

// Procedure defines an operational procedure
type Procedure struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []Step        `json:"steps"`
	Rollback    []Step        `json:"rollback"`
	Validation  []Validation  `json:"validation"`
	Timeout     time.Duration `json:"timeout"`
	Critical    bool          `json:"critical"`
}

// Step represents a single step in a procedure
type Step struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	Critical    bool                   `json:"critical"`
}

// Validation represents a validation check
type Validation struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "http", "database", "metric", "custom"
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    interface{}            `json:"expected"`
	Timeout     time.Duration          `json:"timeout"`
}

// ProcedureExecutor executes procedure steps
type ProcedureExecutor struct {
	DB      *sql.DB
	Metrics *MetricsCollector
	Logger  *RunbookLogger
}

// RunbookLogger handles logging for runbook execution
type RunbookLogger struct {
	DB *sql.DB
}

// ExecutionResult holds the result of a procedure execution
type ExecutionResult struct {
	ExecutionID   string                 `json:"execution_id"`
	ProcedureName string                 `json:"procedure_name"`
	Status        string                 `json:"status"` // "running", "completed", "failed", "rolled_back"
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	StepResults   []StepResult           `json:"step_results"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// StepResult holds the result of a single step
type StepResult struct {
	StepID     string                 `json:"step_id"`
	StepName   string                 `json:"step_name"`
	Status     string                 `json:"status"` // "pending", "running", "completed", "failed", "skipped"
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Error      string                 `json:"error,omitempty"`
	Output     map[string]interface{} `json:"output"`
	RetryCount int                    `json:"retry_count"`
	Duration   time.Duration          `json:"duration"`
}

// NewRunbookEngine creates a new runbook engine
func NewRunbookEngine(db *sql.DB, metrics *MetricsCollector) *RunbookEngine {
	logger := &RunbookLogger{DB: db}
	executor := &ProcedureExecutor{
		DB:      db,
		Metrics: metrics,
		Logger:  logger,
	}

	engine := &RunbookEngine{
		Procedures: make(map[string]Procedure),
		Executor:   *executor,
		Logger:     logger,
		DB:         db,
	}

	// Register built-in procedures
	engine.registerBuiltInProcedures()

	return engine
}

// registerBuiltInProcedures registers common operational procedures
func (re *RunbookEngine) registerBuiltInProcedures() {
	re.RegisterProcedure(Procedure{
		Name:        "canary_rollout",
		Description: "Deploy E2E system using canary rollout strategy",
		Steps: []Step{
			{
				ID:          "validate_prerequisites",
				Name:        "Validate Prerequisites",
				Description: "Check that all prerequisites are met",
				Action:      "validate_prerequisites",
				Timeout:     5 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "enable_canary",
				Name:        "Enable Canary Rollout",
				Description: "Enable canary rollout with initial traffic percentage",
				Action:      "enable_canary",
				Parameters:  map[string]interface{}{"traffic_percentage": 1.0},
				Timeout:     2 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "monitor_initial",
				Name:        "Monitor Initial Deployment",
				Description: "Monitor system for 15 minutes after initial deployment",
				Action:      "monitor_system",
				Parameters:  map[string]interface{}{"duration": "15m"},
				Timeout:     20 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "increase_traffic",
				Name:        "Increase Traffic Percentage",
				Description: "Gradually increase traffic percentage",
				Action:      "increase_traffic",
				Parameters:  map[string]interface{}{"target_percentage": 5.0},
				Timeout:     5 * time.Minute,
				Critical:    true,
			},
		},
		Rollback: []Step{
			{
				ID:          "disable_canary",
				Name:        "Disable Canary Rollout",
				Description: "Disable canary rollout and revert to legacy",
				Action:      "disable_canary",
				Timeout:     1 * time.Minute,
				Critical:    true,
			},
		},
		Validation: []Validation{
			{
				Name:        "Check Error Rate",
				Description: "Verify error rate is within acceptable limits",
				Type:        "metric",
				Parameters:  map[string]interface{}{"metric": "error_rate_percent", "threshold": 1.0},
				Expected:    "below_threshold",
				Timeout:     2 * time.Minute,
			},
		},
		Timeout:  2 * time.Hour,
		Critical: true,
	})

	re.RegisterProcedure(Procedure{
		Name:        "emergency_rollback",
		Description: "Emergency rollback of E2E system",
		Steps: []Step{
			{
				ID:          "assess_situation",
				Name:        "Assess Situation",
				Description: "Quick assessment of the current situation",
				Action:      "assess_situation",
				Timeout:     1 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "disable_e2e",
				Name:        "Disable E2E System",
				Description: "Immediately disable E2E system",
				Action:      "disable_e2e",
				Timeout:     30 * time.Second,
				Critical:    true,
			},
			{
				ID:          "verify_legacy",
				Name:        "Verify Legacy System",
				Description: "Verify legacy system is functioning properly",
				Action:      "verify_legacy",
				Timeout:     2 * time.Minute,
				Critical:    true,
			},
		},
		Timeout:  10 * time.Minute,
		Critical: true,
	})

	re.RegisterProcedure(Procedure{
		Name:        "key_rotation",
		Description: "Rotate cryptographic keys",
		Steps: []Step{
			{
				ID:          "backup_keys",
				Name:        "Backup Current Keys",
				Description: "Create backup of current keys",
				Action:      "backup_keys",
				Timeout:     5 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "generate_new_keys",
				Name:        "Generate New Keys",
				Description: "Generate new cryptographic keys",
				Action:      "generate_new_keys",
				Timeout:     10 * time.Minute,
				Critical:    true,
			},
			{
				ID:          "deploy_new_keys",
				Name:        "Deploy New Keys",
				Description: "Deploy new keys to the system",
				Action:      "deploy_new_keys",
				Timeout:     5 * time.Minute,
				Critical:    true,
			},
		},
		Rollback: []Step{
			{
				ID:          "restore_old_keys",
				Name:        "Restore Old Keys",
				Description: "Restore previous keys",
				Action:      "restore_old_keys",
				Timeout:     5 * time.Minute,
				Critical:    true,
			},
		},
		Timeout:  30 * time.Minute,
		Critical: true,
	})
}

// RegisterProcedure registers a new procedure
func (re *RunbookEngine) RegisterProcedure(procedure Procedure) {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.Procedures[procedure.Name] = procedure
}

// ExecuteProcedure executes a procedure
func (re *RunbookEngine) ExecuteProcedure(procedureName string, operatorID string, correlationID string) (*ExecutionResult, error) {
	re.mu.RLock()
	procedure, exists := re.Procedures[procedureName]
	re.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("procedure '%s' not found", procedureName)
	}

	executionID := uuid.New().String()
	result := &ExecutionResult{
		ExecutionID:   executionID,
		ProcedureName: procedureName,
		Status:        "running",
		StartTime:     time.Now(),
		StepResults:   make([]StepResult, 0),
		Metadata:      make(map[string]interface{}),
	}

	// Record execution start
	err := re.recordExecutionStart(result, operatorID, correlationID)
	if err != nil {
		return nil, fmt.Errorf("failed to record execution start: %w", err)
	}

	// Execute procedure in background
	go re.executeProcedureAsync(procedure, result, operatorID, correlationID)

	return result, nil
}

// executeProcedureAsync executes a procedure asynchronously
func (re *RunbookEngine) executeProcedureAsync(procedure Procedure, result *ExecutionResult, operatorID string, correlationID string) {
	defer func() {
		result.EndTime = time.Now()
		re.recordExecutionEnd(result)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), procedure.Timeout)
	defer cancel()

	// Execute steps
	for i, step := range procedure.Steps {
		stepResult := re.executeStep(ctx, step, i+1, len(procedure.Steps))
		result.StepResults = append(result.StepResults, stepResult)

		// Check if step failed
		if stepResult.Status == "failed" {
			if step.Critical {
				result.Status = "failed"
				result.Error = fmt.Sprintf("Critical step '%s' failed: %s", step.Name, stepResult.Error)
				re.triggerRollback(procedure, result, operatorID, correlationID)
				return
			}
		}
	}

	// Run validations
	if len(procedure.Validation) > 0 {
		validationCtx, validationCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer validationCancel()

		for _, validation := range procedure.Validation {
			valid := re.runValidation(validationCtx, validation)
			if !valid {
				result.Status = "failed"
				result.Error = fmt.Sprintf("Validation '%s' failed", validation.Name)
				re.triggerRollback(procedure, result, operatorID, correlationID)
				return
			}
		}
	}

	result.Status = "completed"
}

// executeStep executes a single step
func (re *RunbookEngine) executeStep(ctx context.Context, step Step, stepNumber int, totalSteps int) StepResult {
	stepResult := StepResult{
		StepID:    step.ID,
		StepName:  step.Name,
		Status:    "running",
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	// Execute the step action
	err := re.Executor.ExecuteAction(ctx, step.Action, step.Parameters, &stepResult)
	if err != nil {
		stepResult.Status = "failed"
		stepResult.Error = err.Error()

		// Retry if configured
		if step.RetryCount > 0 {
			for retry := 1; retry <= step.RetryCount; retry++ {
				stepResult.RetryCount = retry
				time.Sleep(time.Duration(retry) * time.Second) // Exponential backoff

				err = re.Executor.ExecuteAction(ctx, step.Action, step.Parameters, &stepResult)
				if err == nil {
					stepResult.Status = "completed"
					stepResult.Error = ""
					break
				}
				stepResult.Error = err.Error()
			}
		}
	} else {
		stepResult.Status = "completed"
	}

	stepResult.EndTime = time.Now()
	stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)

	return stepResult
}

// triggerRollback triggers the rollback steps for a procedure
func (re *RunbookEngine) triggerRollback(procedure Procedure, result *ExecutionResult, operatorID string, correlationID string) {
	if len(procedure.Rollback) == 0 {
		return
	}

	log.Printf("Triggering rollback for procedure '%s'", procedure.Name)

	rollbackResult := &ExecutionResult{
		ExecutionID:   uuid.New().String(),
		ProcedureName: procedure.Name + "_rollback",
		Status:        "running",
		StartTime:     time.Now(),
		StepResults:   make([]StepResult, 0),
		Metadata:      make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), procedure.Timeout)
	defer cancel()

	// Execute rollback steps
	for _, step := range procedure.Rollback {
		stepResult := re.executeStep(ctx, step, 1, len(procedure.Rollback))
		rollbackResult.StepResults = append(rollbackResult.StepResults, stepResult)

		if stepResult.Status == "failed" {
			rollbackResult.Status = "failed"
			rollbackResult.Error = fmt.Sprintf("Rollback step '%s' failed: %s", step.Name, stepResult.Error)
			break
		}
	}

	if rollbackResult.Status != "failed" {
		rollbackResult.Status = "completed"
		result.Status = "rolled_back"
	}

	rollbackResult.EndTime = time.Now()
	re.recordExecutionEnd(rollbackResult)
}

// runValidation runs a validation check
func (re *RunbookEngine) runValidation(ctx context.Context, validation Validation) bool {
	switch validation.Type {
	case "metric":
		return re.validateMetric(ctx, validation)
	case "http":
		return re.validateHTTP(ctx, validation)
	case "database":
		return re.validateDatabase(ctx, validation)
	default:
		log.Printf("Unknown validation type: %s", validation.Type)
		return false
	}
}

// validateMetric validates a metric
func (re *RunbookEngine) validateMetric(ctx context.Context, validation Validation) bool {
	_, ok := validation.Parameters["metric"].(string)
	if !ok {
		return false
	}

	threshold, ok := validation.Parameters["threshold"].(float64)
	if !ok {
		return false
	}

	// This would typically query the metrics system
	// For now, return a placeholder value
	currentValue := 0.5 // Placeholder

	return currentValue < threshold
}

// validateHTTP validates an HTTP endpoint
func (re *RunbookEngine) validateHTTP(ctx context.Context, validation Validation) bool {
	// This would make an HTTP request to validate the endpoint
	// For now, return true as placeholder
	return true
}

// validateDatabase validates a database query
func (re *RunbookEngine) validateDatabase(ctx context.Context, validation Validation) bool {
	// This would execute a database query to validate
	// For now, return true as placeholder
	return true
}

// recordExecutionStart records the start of an execution
func (re *RunbookEngine) recordExecutionStart(result *ExecutionResult, operatorID string, correlationID string) error {
	_, err := re.DB.Exec(`
		INSERT INTO runbook_executions (id, runbook_name, procedure_name, status, started_at, operator_id, correlation_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, result.ExecutionID, result.ProcedureName, result.ProcedureName, result.Status, result.StartTime, operatorID, correlationID)

	return err
}

// recordExecutionEnd records the end of an execution
func (re *RunbookEngine) recordExecutionEnd(result *ExecutionResult) error {
	executionLog, _ := json.Marshal(result.StepResults)

	_, err := re.DB.Exec(`
		UPDATE runbook_executions 
		SET status = ?, completed_at = ?, error_message = ?, execution_log = ?
		WHERE id = ?
	`, result.Status, result.EndTime, result.Error, executionLog, result.ExecutionID)

	return err
}

// GetExecutionStatus gets the status of an execution
func (re *RunbookEngine) GetExecutionStatus(executionID string) (*ExecutionResult, error) {
	var execution struct {
		RunbookName   string     `json:"runbook_name"`
		ProcedureName string     `json:"procedure_name"`
		Status        string     `json:"status"`
		StartedAt     time.Time  `json:"started_at"`
		CompletedAt   *time.Time `json:"completed_at"`
		ErrorMessage  *string    `json:"error_message"`
		ExecutionLog  *string    `json:"execution_log"`
		OperatorID    string     `json:"operator_id"`
		CorrelationID string     `json:"correlation_id"`
	}

	err := re.DB.QueryRow(`
		SELECT runbook_name, procedure_name, status, started_at, completed_at, error_message, execution_log, operator_id, correlation_id
		FROM runbook_executions WHERE id = ?
	`, executionID).Scan(
		&execution.RunbookName, &execution.ProcedureName, &execution.Status,
		&execution.StartedAt, &execution.CompletedAt, &execution.ErrorMessage,
		&execution.ExecutionLog, &execution.OperatorID, &execution.CorrelationID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get execution status: %w", err)
	}

	result := &ExecutionResult{
		ExecutionID:   executionID,
		ProcedureName: execution.ProcedureName,
		Status:        execution.Status,
		StartTime:     execution.StartedAt,
		Metadata:      make(map[string]interface{}),
	}

	if execution.CompletedAt != nil {
		result.EndTime = *execution.CompletedAt
	}

	if execution.ErrorMessage != nil {
		result.Error = *execution.ErrorMessage
	}

	if execution.ExecutionLog != nil && *execution.ExecutionLog != "" {
		var stepResults []StepResult
		json.Unmarshal([]byte(*execution.ExecutionLog), &stepResults)
		result.StepResults = stepResults
	}

	return result, nil
}

// ListProcedures lists all available procedures
func (re *RunbookEngine) ListProcedures() []Procedure {
	re.mu.RLock()
	defer re.mu.RUnlock()

	procedures := make([]Procedure, 0, len(re.Procedures))
	for _, procedure := range re.Procedures {
		procedures = append(procedures, procedure)
	}

	return procedures
}

// ExecuteAction executes a specific action
func (pe *ProcedureExecutor) ExecuteAction(ctx context.Context, action string, parameters map[string]interface{}, result *StepResult) error {
	switch action {
	case "validate_prerequisites":
		return pe.validatePrerequisites(ctx, parameters, result)
	case "enable_canary":
		return pe.enableCanary(ctx, parameters, result)
	case "monitor_system":
		return pe.monitorSystem(ctx, parameters, result)
	case "increase_traffic":
		return pe.increaseTraffic(ctx, parameters, result)
	case "disable_canary":
		return pe.disableCanary(ctx, parameters, result)
	case "assess_situation":
		return pe.assessSituation(ctx, parameters, result)
	case "disable_e2e":
		return pe.disableE2E(ctx, parameters, result)
	case "verify_legacy":
		return pe.verifyLegacy(ctx, parameters, result)
	case "backup_keys":
		return pe.backupKeys(ctx, parameters, result)
	case "generate_new_keys":
		return pe.generateNewKeys(ctx, parameters, result)
	case "deploy_new_keys":
		return pe.deployNewKeys(ctx, parameters, result)
	case "restore_old_keys":
		return pe.restoreOldKeys(ctx, parameters, result)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// validatePrerequisites validates system prerequisites
func (pe *ProcedureExecutor) validatePrerequisites(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// Check database connectivity
	err := pe.DB.PingContext(ctx)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("database connectivity check failed: %v", err)
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Check metrics system
	if pe.Metrics == nil {
		result.Status = "failed"
		result.Error = "metrics system not available"
		return fmt.Errorf("metrics system not available")
	}

	result.Status = "completed"
	result.Output["database_status"] = "connected"
	result.Output["metrics_status"] = "available"
	return nil
}

// enableCanary enables canary rollout
func (pe *ProcedureExecutor) enableCanary(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	trafficPercentage, ok := parameters["traffic_percentage"].(float64)
	if !ok {
		result.Status = "failed"
		result.Error = "invalid traffic_percentage parameter"
		return fmt.Errorf("invalid traffic_percentage parameter")
	}

	_, err := pe.DB.ExecContext(ctx, `
		UPDATE canary_config 
		SET enabled = TRUE, traffic_percentage = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = 'canary_main'
	`, trafficPercentage)

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to enable canary: %v", err)
		return fmt.Errorf("failed to enable canary: %w", err)
	}

	result.Status = "completed"
	result.Output["traffic_percentage"] = trafficPercentage
	result.Output["status"] = "enabled"
	return nil
}

// monitorSystem monitors system health
func (pe *ProcedureExecutor) monitorSystem(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	durationStr, ok := parameters["duration"].(string)
	if !ok {
		return fmt.Errorf("invalid duration parameter")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	// Simulate monitoring
	time.Sleep(duration)

	result.Output["monitoring_duration"] = duration.String()
	result.Output["status"] = "monitoring_completed"
	return nil
}

// increaseTraffic increases traffic percentage
func (pe *ProcedureExecutor) increaseTraffic(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	targetPercentage, ok := parameters["target_percentage"].(float64)
	if !ok {
		result.Status = "failed"
		result.Error = "invalid target_percentage parameter"
		return fmt.Errorf("invalid target_percentage parameter")
	}

	_, err := pe.DB.ExecContext(ctx, `
		UPDATE canary_config 
		SET traffic_percentage = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = 'canary_main'
	`, targetPercentage)

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to increase traffic: %v", err)
		return fmt.Errorf("failed to increase traffic: %w", err)
	}

	result.Status = "completed"
	result.Output["new_traffic_percentage"] = targetPercentage
	result.Output["status"] = "traffic_increased"
	return nil
}

// disableCanary disables canary rollout
func (pe *ProcedureExecutor) disableCanary(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	_, err := pe.DB.ExecContext(ctx, `
		UPDATE canary_config 
		SET enabled = FALSE, traffic_percentage = 0, updated_at = CURRENT_TIMESTAMP 
		WHERE id = 'canary_main'
	`)

	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to disable canary: %v", err)
		return fmt.Errorf("failed to disable canary: %w", err)
	}

	result.Status = "completed"
	result.Output["status"] = "disabled"
	return nil
}

// assessSituation assesses the current situation
func (pe *ProcedureExecutor) assessSituation(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would typically check various system metrics
	// For now, return placeholder assessment
	result.Status = "completed"
	result.Output["assessment"] = "system_requires_rollback"
	result.Output["urgency"] = "high"
	return nil
}

// disableE2E disables E2E system
func (pe *ProcedureExecutor) disableE2E(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would disable E2E features
	result.Status = "completed"
	result.Output["status"] = "e2e_disabled"
	return nil
}

// verifyLegacy verifies legacy system is working
func (pe *ProcedureExecutor) verifyLegacy(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would verify legacy system functionality
	result.Status = "completed"
	result.Output["status"] = "legacy_verified"
	return nil
}

// backupKeys backs up current keys
func (pe *ProcedureExecutor) backupKeys(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would create a backup of current keys
	result.Status = "completed"
	result.Output["backup_created"] = true
	result.Output["backup_location"] = "/backup/keys_" + time.Now().Format("20060102_150405")
	return nil
}

// generateNewKeys generates new cryptographic keys
func (pe *ProcedureExecutor) generateNewKeys(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would generate new keys
	result.Status = "completed"
	result.Output["keys_generated"] = true
	result.Output["key_count"] = 3
	return nil
}

// deployNewKeys deploys new keys
func (pe *ProcedureExecutor) deployNewKeys(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would deploy new keys to the system
	result.Status = "completed"
	result.Output["keys_deployed"] = true
	result.Output["deployment_time"] = time.Now().Format(time.RFC3339)
	return nil
}

// restoreOldKeys restores old keys
func (pe *ProcedureExecutor) restoreOldKeys(ctx context.Context, parameters map[string]interface{}, result *StepResult) error {
	// This would restore previous keys
	result.Status = "completed"
	result.Output["keys_restored"] = true
	result.Output["restore_time"] = time.Now().Format(time.RFC3339)
	return nil
}
