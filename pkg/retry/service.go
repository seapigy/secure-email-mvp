package retry

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"time"
)

// =============================================================================
// RETRY LOGIC AND ERROR RECOVERY SERVICE
// =============================================================================

// RetryService handles automatic retry logic and error recovery
type RetryService struct {
	db *sql.DB
}

// RetryTask represents a task that can be retried
type RetryTask struct {
	ID             string                 `json:"id" db:"id"`
	TaskType       string                 `json:"task_type" db:"task_type"`
	EntityID       string                 `json:"entity_id" db:"entity_id"`
	Payload        map[string]interface{} `json:"payload" db:"payload"`
	MaxAttempts    int                    `json:"max_attempts" db:"max_attempts"`
	CurrentAttempt int                    `json:"current_attempt" db:"current_attempt"`
	NextRetryAt    time.Time              `json:"next_retry_at" db:"next_retry_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	Status         string                 `json:"status" db:"status"` // "pending", "running", "completed", "failed", "cancelled"
	LastError      *string                `json:"last_error,omitempty" db:"last_error"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// RetryAttempt represents a single retry attempt
type RetryAttempt struct {
	ID        string     `json:"id" db:"id"`
	TaskID    string     `json:"task_id" db:"task_id"`
	AttemptNo int        `json:"attempt_no" db:"attempt_no"`
	StartedAt time.Time  `json:"started_at" db:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	Success   bool       `json:"success" db:"success"`
	Error     *string    `json:"error,omitempty" db:"error"`
	Duration  *int64     `json:"duration_ms,omitempty" db:"duration_ms"` // in milliseconds
}

// RetryConfig represents retry configuration for different task types
type RetryConfig struct {
	TaskType          string        `json:"task_type"`
	MaxAttempts       int           `json:"max_attempts"`
	InitialDelay      time.Duration `json:"initial_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	EnableJitter      bool          `json:"enable_jitter"`
}

// TaskProcessor defines the interface for processing retry tasks
type TaskProcessor interface {
	ProcessTask(ctx context.Context, task *RetryTask) error
	GetTaskType() string
}

// NewRetryService creates a new retry service
func NewRetryService(db *sql.DB) *RetryService {
	return &RetryService{
		db: db,
	}
}

// ScheduleTask schedules a new task for execution with retry capability
func (r *RetryService) ScheduleTask(ctx context.Context, taskType, entityID string, payload map[string]interface{}, config *RetryConfig) (*RetryTask, error) {
	task := &RetryTask{
		ID:             r.generateTaskID(),
		TaskType:       taskType,
		EntityID:       entityID,
		Payload:        payload,
		MaxAttempts:    config.MaxAttempts,
		CurrentAttempt: 0,
		NextRetryAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         "pending",
	}

	// Serialize payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}

	query := `
		INSERT INTO retry_tasks (
			id, task_type, entity_id, payload, max_attempts, current_attempt,
			next_retry_at, created_at, updated_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		task.ID, task.TaskType, task.EntityID, string(payloadJSON),
		task.MaxAttempts, task.CurrentAttempt, task.NextRetryAt,
		task.CreatedAt, task.UpdatedAt, task.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to schedule retry task: %w", err)
	}

	return task, nil
}

// ProcessPendingTasks processes all pending tasks that are ready for execution
func (r *RetryService) ProcessPendingTasks(ctx context.Context, processors map[string]TaskProcessor) error {
	// Get all pending tasks that are ready for execution
	tasks, err := r.getPendingTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending tasks: %w", err)
	}

	for _, task := range tasks {
		// Find processor for this task type
		processor, exists := processors[task.TaskType]
		if !exists {
			log.Printf("No processor found for task type: %s", task.TaskType)
			continue
		}

		// Process the task
		if err := r.processTask(ctx, task, processor); err != nil {
			log.Printf("Failed to process task %s: %v", task.ID, err)
		}
	}

	return nil
}

// processTask processes a single task with retry logic
func (r *RetryService) processTask(ctx context.Context, task *RetryTask, processor TaskProcessor) error {
	// Update task status to running
	if err := r.updateTaskStatus(ctx, task.ID, "running"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Increment attempt counter
	task.CurrentAttempt++
	if err := r.updateCurrentAttempt(ctx, task.ID, task.CurrentAttempt); err != nil {
		return fmt.Errorf("failed to update attempt count: %w", err)
	}

	// Create retry attempt record
	attempt := &RetryAttempt{
		ID:        r.generateAttemptID(),
		TaskID:    task.ID,
		AttemptNo: task.CurrentAttempt,
		StartedAt: time.Now(),
		Success:   false,
	}

	// Record the attempt start
	if err := r.recordAttemptStart(ctx, attempt); err != nil {
		log.Printf("Warning: Failed to record attempt start: %v", err)
	}

	// Process the task
	startTime := time.Now()
	err := processor.ProcessTask(ctx, task)
	duration := time.Since(startTime)

	// Update attempt record
	attempt.EndedAt = &startTime
	durationMs := duration.Milliseconds()
	attempt.Duration = &durationMs

	if err != nil {
		// Task failed
		attempt.Success = false
		errorStr := err.Error()
		attempt.Error = &errorStr

		// Update attempt record
		if updateErr := r.recordAttemptEnd(ctx, attempt); updateErr != nil {
			log.Printf("Warning: Failed to record attempt end: %v", updateErr)
		}

		// Update task with error
		if updateErr := r.updateTaskError(ctx, task.ID, errorStr); updateErr != nil {
			log.Printf("Warning: Failed to update task error: %v", updateErr)
		}

		// Check if we should retry
		if task.CurrentAttempt >= task.MaxAttempts {
			// Max attempts reached, mark as failed
			if updateErr := r.updateTaskStatus(ctx, task.ID, "failed"); updateErr != nil {
				log.Printf("Warning: Failed to update task status to failed: %v", updateErr)
			}
			return fmt.Errorf("task failed after %d attempts: %w", task.MaxAttempts, err)
		} else {
			// Schedule next retry
			nextRetry := r.calculateNextRetry(task.CurrentAttempt)
			if updateErr := r.scheduleNextRetry(ctx, task.ID, nextRetry); updateErr != nil {
				log.Printf("Warning: Failed to schedule next retry: %v", updateErr)
			}
			if updateErr := r.updateTaskStatus(ctx, task.ID, "pending"); updateErr != nil {
				log.Printf("Warning: Failed to update task status to pending: %v", updateErr)
			}
		}

		return err
	}

	// Task succeeded
	attempt.Success = true

	// Update attempt record
	if err := r.recordAttemptEnd(ctx, attempt); err != nil {
		log.Printf("Warning: Failed to record attempt end: %v", err)
	}

	// Update task status to completed
	completedAt := time.Now()
	if err := r.updateTaskCompleted(ctx, task.ID, completedAt); err != nil {
		log.Printf("Warning: Failed to update task completion: %v", err)
	}

	return nil
}

// GetTaskStatus retrieves the current status of a task
func (r *RetryService) GetTaskStatus(ctx context.Context, taskID string) (*RetryTask, error) {
	query := `
		SELECT id, task_type, entity_id, payload, max_attempts, current_attempt,
		       next_retry_at, created_at, updated_at, status, last_error, completed_at
		FROM retry_tasks
		WHERE id = ?
	`

	var task RetryTask
	var payloadJSON string
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID, &task.TaskType, &task.EntityID, &payloadJSON,
		&task.MaxAttempts, &task.CurrentAttempt, &task.NextRetryAt,
		&task.CreatedAt, &task.UpdatedAt, &task.Status,
		&task.LastError, &task.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}

	// Deserialize payload
	if err := json.Unmarshal([]byte(payloadJSON), &task.Payload); err != nil {
		log.Printf("Warning: Failed to unmarshal task payload: %v", err)
		task.Payload = make(map[string]interface{})
	}

	return &task, nil
}

// GetTaskAttempts retrieves all attempts for a task
func (r *RetryService) GetTaskAttempts(ctx context.Context, taskID string) ([]*RetryAttempt, error) {
	query := `
		SELECT id, task_id, attempt_no, started_at, ended_at, success, error, duration_ms
		FROM retry_attempts
		WHERE task_id = ?
		ORDER BY attempt_no ASC
	`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*RetryAttempt
	for rows.Next() {
		var attempt RetryAttempt
		err := rows.Scan(
			&attempt.ID, &attempt.TaskID, &attempt.AttemptNo,
			&attempt.StartedAt, &attempt.EndedAt, &attempt.Success,
			&attempt.Error, &attempt.Duration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retry attempt: %w", err)
		}
		attempts = append(attempts, &attempt)
	}

	return attempts, nil
}

// CancelTask cancels a pending task
func (r *RetryService) CancelTask(ctx context.Context, taskID string) error {
	query := `UPDATE retry_tasks SET status = 'cancelled', updated_at = ? WHERE id = ? AND status IN ('pending', 'running')`
	result, err := r.db.ExecContext(ctx, query, time.Now(), taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get cancel result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found or not in cancellable state")
	}

	return nil
}

// CleanupCompletedTasks removes completed tasks older than the specified retention period
func (r *RetryService) CleanupCompletedTasks(ctx context.Context, retentionPeriod time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-retentionPeriod)

	// Start transaction to clean up both tasks and attempts
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get task IDs to delete
	taskQuery := `SELECT id FROM retry_tasks WHERE status IN ('completed', 'failed', 'cancelled') AND updated_at < ?`
	rows, err := tx.QueryContext(ctx, taskQuery, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to query old tasks: %w", err)
	}

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			rows.Close()
			return 0, fmt.Errorf("failed to scan task ID: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}
	rows.Close()

	if len(taskIDs) == 0 {
		return 0, nil
	}

	// Delete attempts for these tasks
	for _, taskID := range taskIDs {
		_, err = tx.ExecContext(ctx, `DELETE FROM retry_attempts WHERE task_id = ?`, taskID)
		if err != nil {
			return 0, fmt.Errorf("failed to delete attempts for task %s: %w", taskID, err)
		}
	}

	// Delete tasks
	deleteQuery := `DELETE FROM retry_tasks WHERE status IN ('completed', 'failed', 'cancelled') AND updated_at < ?`
	result, err := tx.ExecContext(ctx, deleteQuery, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old tasks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get cleanup result: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	return rowsAffected, nil
}

// Helper methods

// getPendingTasks retrieves all pending tasks ready for execution
func (r *RetryService) getPendingTasks(ctx context.Context) ([]*RetryTask, error) {
	query := `
		SELECT id, task_type, entity_id, payload, max_attempts, current_attempt,
		       next_retry_at, created_at, updated_at, status, last_error, completed_at
		FROM retry_tasks
		WHERE status = 'pending' AND next_retry_at <= ?
		ORDER BY next_retry_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*RetryTask
	for rows.Next() {
		var task RetryTask
		var payloadJSON string
		err := rows.Scan(
			&task.ID, &task.TaskType, &task.EntityID, &payloadJSON,
			&task.MaxAttempts, &task.CurrentAttempt, &task.NextRetryAt,
			&task.CreatedAt, &task.UpdatedAt, &task.Status,
			&task.LastError, &task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retry task: %w", err)
		}

		// Deserialize payload
		if err := json.Unmarshal([]byte(payloadJSON), &task.Payload); err != nil {
			log.Printf("Warning: Failed to unmarshal task payload: %v", err)
			task.Payload = make(map[string]interface{})
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// calculateNextRetry calculates the next retry time using exponential backoff
func (r *RetryService) calculateNextRetry(attempt int) time.Time {
	// Exponential backoff: 2^attempt seconds, max 1 hour
	delaySeconds := 1 << uint(attempt-1) // 2^(attempt-1)
	if delaySeconds > 3600 {             // max 1 hour
		delaySeconds = 3600
	}

	// Add jitter (±25%)
	jitter := float64(delaySeconds) * 0.25
	jitterSeconds := int(jitter * (2*mathrand.Float64() - 1))

	return time.Now().Add(time.Duration(delaySeconds+jitterSeconds) * time.Second)
}

// Database helper methods

func (r *RetryService) updateTaskStatus(ctx context.Context, taskID, status string) error {
	query := `UPDATE retry_tasks SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), taskID)
	return err
}

func (r *RetryService) updateCurrentAttempt(ctx context.Context, taskID string, attempt int) error {
	query := `UPDATE retry_tasks SET current_attempt = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, attempt, time.Now(), taskID)
	return err
}

func (r *RetryService) updateTaskError(ctx context.Context, taskID, errorMsg string) error {
	query := `UPDATE retry_tasks SET last_error = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, errorMsg, time.Now(), taskID)
	return err
}

func (r *RetryService) scheduleNextRetry(ctx context.Context, taskID string, nextRetry time.Time) error {
	query := `UPDATE retry_tasks SET next_retry_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, nextRetry, time.Now(), taskID)
	return err
}

func (r *RetryService) updateTaskCompleted(ctx context.Context, taskID string, completedAt time.Time) error {
	query := `UPDATE retry_tasks SET status = 'completed', completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, completedAt, time.Now(), taskID)
	return err
}

func (r *RetryService) recordAttemptStart(ctx context.Context, attempt *RetryAttempt) error {
	query := `
		INSERT INTO retry_attempts (id, task_id, attempt_no, started_at, success)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.TaskID, attempt.AttemptNo, attempt.StartedAt, attempt.Success,
	)
	return err
}

func (r *RetryService) recordAttemptEnd(ctx context.Context, attempt *RetryAttempt) error {
	query := `
		UPDATE retry_attempts 
		SET ended_at = ?, success = ?, error = ?, duration_ms = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		attempt.EndedAt, attempt.Success, attempt.Error, attempt.Duration, attempt.ID,
	)
	return err
}

func (r *RetryService) generateTaskID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (r *RetryService) generateAttemptID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// GetDefaultRetryConfigs returns default retry configurations for different task types
func GetDefaultRetryConfigs() map[string]*RetryConfig {
	return map[string]*RetryConfig{
		"email_send": {
			TaskType:          "email_send",
			MaxAttempts:       3,
			InitialDelay:      time.Second,
			MaxDelay:          time.Minute * 5,
			BackoffMultiplier: 2.0,
			EnableJitter:      true,
		},
		"link_creation": {
			TaskType:          "link_creation",
			MaxAttempts:       5,
			InitialDelay:      time.Second,
			MaxDelay:          time.Minute * 2,
			BackoffMultiplier: 1.5,
			EnableJitter:      true,
		},
		"notification_send": {
			TaskType:          "notification_send",
			MaxAttempts:       3,
			InitialDelay:      time.Second * 2,
			MaxDelay:          time.Minute * 10,
			BackoffMultiplier: 2.0,
			EnableJitter:      true,
		},
		"cleanup": {
			TaskType:          "cleanup",
			MaxAttempts:       2,
			InitialDelay:      time.Minute,
			MaxDelay:          time.Hour,
			BackoffMultiplier: 2.0,
			EnableJitter:      false,
		},
	}
}
