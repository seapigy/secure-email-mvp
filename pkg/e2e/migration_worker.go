package e2e

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// MigrationWorker handles background migration of legacy data to E2E format
type MigrationWorker struct {
	db              *sql.DB
	jobQueue        chan MigrationJob
	progressTracker *ProgressTracker
	errorHandler    *ErrorHandler
	config          MigrationConfig
	mu              sync.RWMutex
	isRunning       bool
	currentJob      *MigrationJob
	ctx             context.Context
	cancel          context.CancelFunc
}

// MigrationJob represents a migration job
type MigrationJob struct {
	ID                string     `json:"id"`
	Type              string     `json:"type"`   // 'legacy_to_e2e', 'key_rotation', 'metadata_migration'
	Status            string     `json:"status"` // 'pending', 'running', 'completed', 'failed', 'paused'
	Progress          int        `json:"progress"`
	TotalItems        int        `json:"total_items"`
	ProcessedItems    int        `json:"processed_items"`
	FailedItems       int        `json:"failed_items"`
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	RollbackAvailable bool       `json:"rollback_available"`
	CreatedBy         string     `json:"created_by"`

	// Job-specific data
	BatchSize  int           `json:"batch_size"`
	RetryCount int           `json:"retry_count"`
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`

	// Migration-specific data
	SourceTable    string            `json:"source_table,omitempty"`
	TargetTable    string            `json:"target_table,omitempty"`
	Filters        map[string]string `json:"filters,omitempty"`
	TransformRules map[string]string `json:"transform_rules,omitempty"`
}

// MigrationConfig contains configuration for the migration worker
type MigrationConfig struct {
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	DefaultBatchSize  int           `json:"default_batch_size"`
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
	ProgressInterval  time.Duration `json:"progress_interval"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

// ProgressTracker tracks migration progress
type ProgressTracker struct {
	db *sql.DB
	mu sync.RWMutex
}

// ErrorHandler handles migration errors
type ErrorHandler struct {
	db *sql.DB
	mu sync.RWMutex
}

// MigrationProgress represents migration progress
type MigrationProgress struct {
	JobID          string    `json:"job_id"`
	Progress       int       `json:"progress"`
	TotalItems     int       `json:"total_items"`
	ProcessedItems int       `json:"processed_items"`
	FailedItems    int       `json:"failed_items"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	LastUpdated    time.Time `json:"last_updated"`
}

// TestMessage represents a test message for migration
type TestMessage struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recipient_id"`
	Content     string    `json:"content"`
	Subject     string    `json:"subject"`
	CreatedAt   time.Time `json:"created_at"`
	Type        string    `json:"type"`
}

// TestUser represents a test user for migration
type TestUser struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	E2EEnabled bool   `json:"e2e_enabled"`
}

// NewMigrationWorker creates a new migration worker
func NewMigrationWorker(db *sql.DB, config MigrationConfig) *MigrationWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &MigrationWorker{
		db:              db,
		jobQueue:        make(chan MigrationJob, config.MaxConcurrentJobs),
		progressTracker: NewProgressTracker(db),
		errorHandler:    NewErrorHandler(db),
		config:          config,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(db *sql.DB) *ProgressTracker {
	return &ProgressTracker{
		db: db,
	}
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(db *sql.DB) *ErrorHandler {
	return &ErrorHandler{
		db: db,
	}
}

// Start starts the migration worker
func (w *MigrationWorker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return fmt.Errorf("migration worker is already running")
	}

	w.isRunning = true

	// Start worker goroutines
	for i := 0; i < w.config.MaxConcurrentJobs; i++ {
		go w.worker()
	}

	// Start progress tracker
	go w.progressTrackerLoop()

	// Start heartbeat
	go w.heartbeat()

	log.Printf("Migration worker started with %d workers", w.config.MaxConcurrentJobs)
	return nil
}

// Stop stops the migration worker
func (w *MigrationWorker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isRunning {
		return fmt.Errorf("migration worker is not running")
	}

	w.isRunning = false
	w.cancel()

	log.Printf("Migration worker stopped")
	return nil
}

// SubmitJob submits a migration job
func (w *MigrationWorker) SubmitJob(job MigrationJob) error {
	// Validate job
	if err := w.validateJob(&job); err != nil {
		return fmt.Errorf("invalid job: %w", err)
	}

	// Set default values
	if job.BatchSize == 0 {
		job.BatchSize = w.config.DefaultBatchSize
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = w.config.MaxRetries
	}
	if job.RetryDelay == 0 {
		job.RetryDelay = w.config.RetryDelay
	}

	// Store job in database
	if err := w.storeJob(&job); err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	// Submit to queue
	select {
	case w.jobQueue <- job:
		log.Printf("Job %s submitted for migration type %s", job.ID, job.Type)
		return nil
	case <-w.ctx.Done():
		return fmt.Errorf("migration worker is stopped")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetJobStatus gets the status of a migration job
func (w *MigrationWorker) GetJobStatus(jobID string) (*MigrationProgress, error) {
	return w.progressTracker.GetProgress(jobID)
}

// PauseJob pauses a migration job
func (w *MigrationWorker) PauseJob(jobID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentJob != nil && w.currentJob.ID == jobID {
		w.currentJob.Status = "paused"
		return w.progressTracker.UpdateStatus(jobID, "paused", "")
	}

	return fmt.Errorf("job %s not found or not running", jobID)
}

// ResumeJob resumes a migration job
func (w *MigrationWorker) ResumeJob(jobID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentJob != nil && w.currentJob.ID == jobID {
		w.currentJob.Status = "running"
		return w.progressTracker.UpdateStatus(jobID, "running", "")
	}

	return fmt.Errorf("job %s not found or not paused", jobID)
}

// RollbackJob rolls back a migration job
func (w *MigrationWorker) RollbackJob(jobID string) error {
	// Get job details
	job, err := w.getJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if !job.RollbackAvailable {
		return fmt.Errorf("rollback not available for job %s", jobID)
	}

	// Perform rollback based on job type
	switch job.Type {
	case "legacy_to_e2e":
		return w.rollbackLegacyToE2E(job)
	case "key_rotation":
		return w.rollbackKeyRotation(job)
	case "metadata_migration":
		return w.rollbackMetadataMigration(job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

// worker processes migration jobs
func (w *MigrationWorker) worker() {
	for {
		select {
		case job := <-w.jobQueue:
			w.processJob(&job)
		case <-w.ctx.Done():
			return
		}
	}
}

// processJob processes a single migration job
func (w *MigrationWorker) processJob(job *MigrationJob) {
	w.mu.Lock()
	w.currentJob = job
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.currentJob = nil
		w.mu.Unlock()
	}()

	// Update job status to running
	job.Status = "running"
	job.StartedAt = time.Now()
	w.progressTracker.UpdateStatus(job.ID, "running", "")

	log.Printf("Processing migration job %s of type %s", job.ID, job.Type)

	// Process job based on type
	var err error
	switch job.Type {
	case "legacy_to_e2e":
		err = w.migrateLegacyToE2E(job)
	case "key_rotation":
		err = w.migrateKeyRotation(job)
	case "metadata_migration":
		err = w.migrateMetadata(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Handle job completion
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		w.progressTracker.UpdateStatus(job.ID, "failed", err.Error())
		log.Printf("Migration job %s failed: %v", job.ID, err)
	} else {
		job.Status = "completed"
		now := time.Now()
		job.CompletedAt = &now
		w.progressTracker.UpdateStatus(job.ID, "completed", "")
		log.Printf("Migration job %s completed successfully", job.ID)
	}
}

// migrateLegacyToE2E migrates legacy messages to E2E format
func (w *MigrationWorker) migrateLegacyToE2E(job *MigrationJob) error {
	// Get total count of legacy messages
	totalCount, err := w.getLegacyMessageCount(job.Filters)
	if err != nil {
		return fmt.Errorf("failed to get legacy message count: %w", err)
	}

	job.TotalItems = totalCount
	w.progressTracker.UpdateProgress(job.ID, 0, 0, 0)

	// Process in batches
	offset := 0
	for offset < totalCount {
		// Check if job is paused
		if job.Status == "paused" {
			time.Sleep(time.Second)
			continue
		}

		// Get batch of legacy messages
		messages, err := w.getLegacyMessages(job.Filters, job.BatchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to get legacy messages: %w", err)
		}

		// Process each message in the batch
		for _, msg := range messages {
			// Check if user has E2E enabled
			if !w.isE2EEnabledForUser(msg.SenderID) {
				w.recordProcessedItem(job, msg.ID)
				continue
			}

			// Re-encrypt with E2E
			e2eEnvelope, err := w.reencryptWithE2E(msg)
			if err != nil {
				w.recordFailedItem(job, msg.ID, err)
				continue
			}

			// Store E2E message
			err = w.storeE2EMessage(e2eEnvelope)
			if err != nil {
				w.recordFailedItem(job, msg.ID, err)
				continue
			}

			// Mark legacy message as migrated
			err = w.markLegacyMessageMigrated(msg.ID)
			if err != nil {
				w.recordFailedItem(job, msg.ID, err)
				continue
			}

			w.recordProcessedItem(job, msg.ID)
		}

		offset += job.BatchSize

		// Update progress
		progress := int(float64(offset) / float64(totalCount) * 100)
		w.progressTracker.UpdateProgress(job.ID, progress, job.ProcessedItems, job.FailedItems)
	}

	return nil
}

// migrateKeyRotation performs key rotation migration
func (w *MigrationWorker) migrateKeyRotation(job *MigrationJob) error {
	// Get users requiring key rotation
	users, err := w.getUsersForKeyRotation(job.Filters)
	if err != nil {
		return fmt.Errorf("failed to get users for key rotation: %w", err)
	}

	job.TotalItems = len(users)
	w.progressTracker.UpdateProgress(job.ID, 0, 0, 0)

	// Process each user
	for i, user := range users {
		// Check if job is paused
		if job.Status == "paused" {
			time.Sleep(time.Second)
			continue
		}

		// Rotate user keys
		err := w.rotateUserKeys(user.ID)
		if err != nil {
			w.recordFailedItem(job, user.ID, err)
			continue
		}

		w.recordProcessedItem(job, user.ID)

		// Update progress
		progress := int(float64(i+1) / float64(len(users)) * 100)
		w.progressTracker.UpdateProgress(job.ID, progress, job.ProcessedItems, job.FailedItems)
	}

	return nil
}

// migrateMetadata performs metadata migration
func (w *MigrationWorker) migrateMetadata(job *MigrationJob) error {
	// Get messages requiring metadata migration
	messages, err := w.getMessagesForMetadataMigration(job.Filters)
	if err != nil {
		return fmt.Errorf("failed to get messages for metadata migration: %w", err)
	}

	job.TotalItems = len(messages)
	w.progressTracker.UpdateProgress(job.ID, 0, 0, 0)

	// Process each message
	for i, msg := range messages {
		// Check if job is paused
		if job.Status == "paused" {
			time.Sleep(time.Second)
			continue
		}

		// Apply metadata minimization
		minimizedMetadata, err := w.minimizeMessageMetadata(msg)
		if err != nil {
			w.recordFailedItem(job, msg.ID, err)
			continue
		}

		// Update message with minimized metadata
		err = w.updateMessageMetadata(msg.ID, minimizedMetadata)
		if err != nil {
			w.recordFailedItem(job, msg.ID, err)
			continue
		}

		w.recordProcessedItem(job, msg.ID)

		// Update progress
		progress := int(float64(i+1) / float64(len(messages)) * 100)
		w.progressTracker.UpdateProgress(job.ID, progress, job.ProcessedItems, job.FailedItems)
	}

	return nil
}

// progressTrackerLoop updates progress periodically
func (w *MigrationWorker) progressTrackerLoop() {
	ticker := time.NewTicker(w.config.ProgressInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.RLock()
			if w.currentJob != nil {
				w.progressTracker.UpdateProgress(
					w.currentJob.ID,
					w.currentJob.Progress,
					w.currentJob.ProcessedItems,
					w.currentJob.FailedItems,
				)
			}
			w.mu.RUnlock()
		case <-w.ctx.Done():
			return
		}
	}
}

// heartbeat sends heartbeat signals
func (w *MigrationWorker) heartbeat() {
	ticker := time.NewTicker(w.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.RLock()
			isRunning := w.isRunning
			w.mu.RUnlock()

			if isRunning {
				log.Printf("Migration worker heartbeat - running")
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// Helper methods

func (w *MigrationWorker) validateJob(job *MigrationJob) error {
	if job.ID == "" {
		job.ID = w.generateJobID()
	}

	if job.Type == "" {
		return fmt.Errorf("job type is required")
	}

	validTypes := []string{"legacy_to_e2e", "key_rotation", "metadata_migration"}
	valid := false
	for _, t := range validTypes {
		if job.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid job type: %s", job.Type)
	}

	return nil
}

func (w *MigrationWorker) generateJobID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("migration_%x", id)
}

func (w *MigrationWorker) storeJob(job *MigrationJob) error {
	query := `
		INSERT INTO e2e_migrations (
			id, migration_type, status, progress, total_items, processed_items, failed_items,
			started_at, error_message, rollback_available, created_by, batch_size, max_retries, retry_delay
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := w.db.Exec(query,
		job.ID,
		job.Type,
		job.Status,
		job.Progress,
		job.TotalItems,
		job.ProcessedItems,
		job.FailedItems,
		job.StartedAt,
		job.ErrorMessage,
		job.RollbackAvailable,
		job.CreatedBy,
		job.BatchSize,
		job.MaxRetries,
		job.RetryDelay,
	)

	return err
}

func (w *MigrationWorker) getJob(jobID string) (*MigrationJob, error) {
	query := `SELECT * FROM e2e_migrations WHERE id = ?`
	row := w.db.QueryRow(query, jobID)

	var job MigrationJob
	err := row.Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.Progress,
		&job.TotalItems,
		&job.ProcessedItems,
		&job.FailedItems,
		&job.StartedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
		&job.RollbackAvailable,
		&job.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (w *MigrationWorker) recordProcessedItem(job *MigrationJob, itemID string) {
	job.ProcessedItems++
	job.Progress = int(float64(job.ProcessedItems) / float64(job.TotalItems) * 100)
}

func (w *MigrationWorker) recordFailedItem(job *MigrationJob, itemID string, err error) {
	job.FailedItems++
	w.errorHandler.RecordError(job.ID, itemID, err)
}

// ProgressTracker methods

func (pt *ProgressTracker) UpdateProgress(jobID string, progress int, processed, failed int) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	query := `
		UPDATE e2e_migrations 
		SET progress = ?, processed_items = ?, failed_items = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := pt.db.Exec(query, progress, processed, failed, jobID)
	return err
}

func (pt *ProgressTracker) UpdateStatus(jobID string, status string, errorMessage string) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	query := `
		UPDATE e2e_migrations 
		SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := pt.db.Exec(query, status, errorMessage, jobID)
	return err
}

func (pt *ProgressTracker) GetProgress(jobID string) (*MigrationProgress, error) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	query := `
		SELECT progress, total_items, processed_items, failed_items, status, error_message, updated_at
		FROM e2e_migrations 
		WHERE id = ?
	`

	var progress MigrationProgress
	err := pt.db.QueryRow(query, jobID).Scan(
		&progress.Progress,
		&progress.TotalItems,
		&progress.ProcessedItems,
		&progress.FailedItems,
		&progress.Status,
		&progress.ErrorMessage,
		&progress.LastUpdated,
	)

	if err != nil {
		return nil, err
	}

	progress.JobID = jobID
	return &progress, nil
}

// ErrorHandler methods

func (eh *ErrorHandler) RecordError(jobID string, itemID string, err error) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	// Store error in database
	query := `
		INSERT INTO e2e_migration_errors (job_id, item_id, error_message, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, dbErr := eh.db.Exec(query, jobID, itemID, err.Error(), time.Now())
	if dbErr != nil {
		log.Printf("Failed to record migration error: %v", dbErr)
	}
}

// Placeholder methods for database operations

func (w *MigrationWorker) getLegacyMessageCount(filters map[string]string) (int, error) {
	// Placeholder implementation
	return 1000, nil
}

func (w *MigrationWorker) getLegacyMessages(filters map[string]string, batchSize, offset int) ([]*TestMessage, error) {
	// Placeholder implementation
	return []*TestMessage{}, nil
}

func (w *MigrationWorker) isE2EEnabledForUser(userID string) bool {
	// Placeholder implementation
	return true
}

func (w *MigrationWorker) reencryptWithE2E(msg *TestMessage) (interface{}, error) {
	// Placeholder implementation
	return nil, nil
}

func (w *MigrationWorker) storeE2EMessage(envelope interface{}) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) markLegacyMessageMigrated(msgID string) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) getUsersForKeyRotation(filters map[string]string) ([]*TestUser, error) {
	// Placeholder implementation
	return []*TestUser{}, nil
}

func (w *MigrationWorker) rotateUserKeys(userID string) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) getMessagesForMetadataMigration(filters map[string]string) ([]*TestMessage, error) {
	// Placeholder implementation
	return []*TestMessage{}, nil
}

func (w *MigrationWorker) minimizeMessageMetadata(msg *TestMessage) (interface{}, error) {
	// Placeholder implementation
	return nil, nil
}

func (w *MigrationWorker) updateMessageMetadata(msgID string, metadata interface{}) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) rollbackLegacyToE2E(job *MigrationJob) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) rollbackKeyRotation(job *MigrationJob) error {
	// Placeholder implementation
	return nil
}

func (w *MigrationWorker) rollbackMetadataMigration(job *MigrationJob) error {
	// Placeholder implementation
	return nil
}
