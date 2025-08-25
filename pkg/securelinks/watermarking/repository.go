package watermarking

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/models"
)

// WatermarkRepository defines the interface for watermarking data operations
type WatermarkRepository interface {
	// Watermark configuration operations
	SaveConfig(config *models.WatermarkConfig) error
	GetConfigByLinkID(linkID string) (*models.WatermarkConfig, error)
	GetConfigByAttachmentID(attachmentID string) (*models.WatermarkConfig, error)
	UpdateConfig(config *models.WatermarkConfig) error
	DeleteConfig(configID string) error

	// Watermark template operations
	SaveTemplate(template *models.WatermarkTemplate) error
	GetTemplateByID(id string) (*models.WatermarkTemplate, error)
	ListTemplates(watermarkType, contentType string) ([]*models.WatermarkTemplate, error)
	UpdateTemplate(template *models.WatermarkTemplate) error
	DeleteTemplate(templateID string) error

	// Audit log operations
	SaveAuditLog(log *models.WatermarkAuditLog) error
	GetAuditLogsByLinkID(linkID string) ([]*models.WatermarkAuditLog, error)
	GetAuditLogsByAttachmentID(attachmentID string) ([]*models.WatermarkAuditLog, error)
	GetAuditLogsByDateRange(startDate, endDate time.Time) ([]*models.WatermarkAuditLog, error)
}

// SQLiteWatermarkRepository implements WatermarkRepository using SQLite
type SQLiteWatermarkRepository struct {
	db *sql.DB
}

// NewSQLiteWatermarkRepository creates a new SQLite-based watermark repository
func NewSQLiteWatermarkRepository(db *sql.DB) *SQLiteWatermarkRepository {
	return &SQLiteWatermarkRepository{db: db}
}

// isSQLiteBusyError checks if the error is a SQLite busy error
func isSQLiteBusyError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite busy error typically contains "database is locked" or "SQLITE_BUSY"
	return err.Error() == "database is locked (5) (SQLITE_BUSY)" ||
		err.Error() == "database is locked" ||
		err.Error() == "SQLITE_BUSY"
}

// SaveConfig saves a watermark configuration to the database
func (r *SQLiteWatermarkRepository) SaveConfig(config *models.WatermarkConfig) error {
	query := `
		INSERT INTO watermark_configs (
			config_id, attachment_id, watermark_text, watermark_position, 
			watermark_opacity, watermark_font_size, watermark_color, watermark_rotation,
			applied_at, watermark_hash, created_by, recipient_email, recipient_id,
			watermark_type, content_type, watermark_data, is_recipient_specific
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Retry logic for database busy errors
	maxRetries := 3
	backoffMs := 50 // Start with 50ms backoff for writes

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := r.db.Exec(query,
			config.ConfigID, config.AttachmentID, config.WatermarkText, config.WatermarkPosition,
			config.WatermarkOpacity, config.WatermarkFontSize, config.WatermarkColor, config.WatermarkRotation,
			config.AppliedAt, config.WatermarkHash, config.CreatedBy, config.RecipientEmail, config.RecipientID,
			config.WatermarkType, config.ContentType, config.WatermarkData, config.IsRecipientSpecific,
		)
		if err != nil {
			// Check if it's a busy error
			if isSQLiteBusyError(err) {
				log.Printf("Database busy on SaveConfig attempt %d, retrying in %dms: %v", attempt+1, backoffMs, err)
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
				backoffMs *= 2 // Exponential backoff
				continue
			}
			return err
		}
		return nil // Success
	}

	return fmt.Errorf("failed to save config after %d retries due to database busy", maxRetries)
}

// GetConfigByLinkID retrieves watermark configuration by link ID
func (r *SQLiteWatermarkRepository) GetConfigByLinkID(linkID string) (*models.WatermarkConfig, error) {
	query := `
		SELECT config_id, attachment_id, watermark_text, watermark_position,
		       watermark_opacity, watermark_font_size, watermark_color, watermark_rotation,
		       applied_at, watermark_hash, created_by, recipient_email, recipient_id,
		       watermark_type, content_type, watermark_data, is_recipient_specific
		FROM watermark_configs 
		WHERE attachment_id IN (
			SELECT attachment_id FROM secure_attachments WHERE link_id = ?
		)
		ORDER BY applied_at DESC LIMIT 1
	`

	var config models.WatermarkConfig
	err := r.db.QueryRow(query, linkID).Scan(
		&config.ConfigID, &config.AttachmentID, &config.WatermarkText, &config.WatermarkPosition,
		&config.WatermarkOpacity, &config.WatermarkFontSize, &config.WatermarkColor, &config.WatermarkRotation,
		&config.AppliedAt, &config.WatermarkHash, &config.CreatedBy, &config.RecipientEmail, &config.RecipientID,
		&config.WatermarkType, &config.ContentType, &config.WatermarkData, &config.IsRecipientSpecific,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &config, err
}

// GetConfigByAttachmentID retrieves watermark configuration by attachment ID
func (r *SQLiteWatermarkRepository) GetConfigByAttachmentID(attachmentID string) (*models.WatermarkConfig, error) {
	query := `
		SELECT config_id, attachment_id, watermark_text, watermark_position,
		       watermark_opacity, watermark_font_size, watermark_color, watermark_rotation,
		       applied_at, watermark_hash, created_by, recipient_email, recipient_id,
		       watermark_type, content_type, watermark_data, is_recipient_specific
		FROM watermark_configs 
		WHERE attachment_id = ?
		ORDER BY applied_at DESC LIMIT 1
	`

	var config models.WatermarkConfig
	err := r.db.QueryRow(query, attachmentID).Scan(
		&config.ConfigID, &config.AttachmentID, &config.WatermarkText, &config.WatermarkPosition,
		&config.WatermarkOpacity, &config.WatermarkFontSize, &config.WatermarkColor, &config.WatermarkRotation,
		&config.AppliedAt, &config.WatermarkHash, &config.CreatedBy, &config.RecipientEmail, &config.RecipientID,
		&config.WatermarkType, &config.ContentType, &config.WatermarkData, &config.IsRecipientSpecific,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &config, err
}

// UpdateConfig updates an existing watermark configuration
func (r *SQLiteWatermarkRepository) UpdateConfig(config *models.WatermarkConfig) error {
	query := `
		UPDATE watermark_configs SET
			watermark_text = ?, watermark_position = ?, watermark_opacity = ?,
			watermark_font_size = ?, watermark_color = ?, watermark_rotation = ?,
			applied_at = ?, watermark_hash = ?, created_by = ?, recipient_email = ?,
			recipient_id = ?, watermark_type = ?, content_type = ?, watermark_data = ?,
			is_recipient_specific = ?
		WHERE config_id = ?
	`

	_, err := r.db.Exec(query,
		config.WatermarkText, config.WatermarkPosition, config.WatermarkOpacity,
		config.WatermarkFontSize, config.WatermarkColor, config.WatermarkRotation,
		config.AppliedAt, config.WatermarkHash, config.CreatedBy, config.RecipientEmail,
		config.RecipientID, config.WatermarkType, config.ContentType, config.WatermarkData,
		config.IsRecipientSpecific, config.ConfigID,
	)
	return err
}

// DeleteConfig deletes a watermark configuration
func (r *SQLiteWatermarkRepository) DeleteConfig(configID string) error {
	query := `DELETE FROM watermark_configs WHERE config_id = ?`
	_, err := r.db.Exec(query, configID)
	return err
}

// SaveTemplate saves a watermark template to the database
func (r *SQLiteWatermarkRepository) SaveTemplate(template *models.WatermarkTemplate) error {
	query := `
		INSERT INTO watermark_templates (
			template_id, template_name, template_description, watermark_type,
			content_types, default_config, is_recipient_specific, is_active,
			created_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		template.TemplateID, template.TemplateName, template.TemplateDescription, template.WatermarkType,
		template.ContentTypes, template.DefaultConfig, template.IsRecipientSpecific, template.IsActive,
		template.CreatedAt, template.CreatedBy,
	)
	return err
}

// GetTemplateByID retrieves a watermark template by ID
func (r *SQLiteWatermarkRepository) GetTemplateByID(id string) (*models.WatermarkTemplate, error) {
	query := `
		SELECT template_id, template_name, template_description, watermark_type,
		       content_types, default_config, is_recipient_specific, is_active,
		       created_at, created_by
		FROM watermark_templates 
		WHERE template_id = ?
	`

	var template models.WatermarkTemplate
	err := r.db.QueryRow(query, id).Scan(
		&template.TemplateID, &template.TemplateName, &template.TemplateDescription, &template.WatermarkType,
		&template.ContentTypes, &template.DefaultConfig, &template.IsRecipientSpecific, &template.IsActive,
		&template.CreatedAt, &template.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &template, err
}

// ListTemplates retrieves watermark templates with optional filtering
func (r *SQLiteWatermarkRepository) ListTemplates(watermarkType, contentType string) ([]*models.WatermarkTemplate, error) {
	log.Println("Executing ListTemplates query")
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Panic in ListTemplates: %v", rec)
		}
	}()

	query := `
		SELECT template_id, template_name, template_description, watermark_type,
		       content_types, default_config, is_recipient_specific, is_active,
		       created_at, created_by
		FROM watermark_templates 
		WHERE is_active = 1
	`

	var args []interface{}
	if watermarkType != "" {
		query += " AND watermark_type = ?"
		args = append(args, watermarkType)
	}
	if contentType != "" {
		query += " AND content_types LIKE ?"
		args = append(args, "%"+contentType+"%")
	}

	query += " ORDER BY template_name"

	log.Printf("Executing query: %s with args: %v", query, args)

	// Retry logic for database busy errors
	maxRetries := 5
	backoffMs := 100 // Start with 100ms backoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		rows, err := r.db.Query(query, args...)
		if err != nil {
			// Check if it's a busy error
			if isSQLiteBusyError(err) {
				log.Printf("Database busy on attempt %d, retrying in %dms: %v", attempt+1, backoffMs, err)
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
				backoffMs *= 2 // Exponential backoff
				continue
			}
			log.Printf("Query failed: %v", err)
			return nil, fmt.Errorf("failed to query templates: %w", err)
		}
		defer rows.Close()

		var templates []*models.WatermarkTemplate
		for rows.Next() {
			var template models.WatermarkTemplate
			err := rows.Scan(
				&template.TemplateID, &template.TemplateName, &template.TemplateDescription, &template.WatermarkType,
				&template.ContentTypes, &template.DefaultConfig, &template.IsRecipientSpecific, &template.IsActive,
				&template.CreatedAt, &template.CreatedBy,
			)
			if err != nil {
				log.Printf("Row scan failed: %v", err)
				return nil, fmt.Errorf("failed to scan template: %w", err)
			}
			templates = append(templates, &template)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Rows error: %v", err)
			return nil, fmt.Errorf("rows error: %w", err)
		}

		log.Printf("ListTemplates returned %d templates", len(templates))
		return templates, nil
	}

	return nil, fmt.Errorf("failed to query templates after %d retries due to database busy", maxRetries)
}

// UpdateTemplate updates an existing watermark template
func (r *SQLiteWatermarkRepository) UpdateTemplate(template *models.WatermarkTemplate) error {
	query := `
		UPDATE watermark_templates SET
			template_name = ?, template_description = ?, watermark_type = ?,
			content_types = ?, default_config = ?, is_recipient_specific = ?,
			is_active = ?, created_by = ?
		WHERE template_id = ?
	`

	_, err := r.db.Exec(query,
		template.TemplateName, template.TemplateDescription, template.WatermarkType,
		template.ContentTypes, template.DefaultConfig, template.IsRecipientSpecific,
		template.IsActive, template.CreatedBy, template.TemplateID,
	)
	return err
}

// DeleteTemplate deletes a watermark template
func (r *SQLiteWatermarkRepository) DeleteTemplate(templateID string) error {
	query := `DELETE FROM watermark_templates WHERE template_id = ?`
	_, err := r.db.Exec(query, templateID)
	return err
}

// SaveAuditLog saves a watermark audit log entry
func (r *SQLiteWatermarkRepository) SaveAuditLog(auditLog *models.WatermarkAuditLog) error {
	query := `
		INSERT INTO advanced_watermark_audit_log (
			audit_id, link_id, attachment_id, watermark_type, content_type,
			recipient_email, recipient_id, watermark_config, applied_at,
			success, error_message, created_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Retry logic for database busy errors
	maxRetries := 3
	backoffMs := 50 // Start with 50ms backoff for writes

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := r.db.Exec(query,
			auditLog.AuditID, auditLog.LinkID, auditLog.AttachmentID, auditLog.WatermarkType, auditLog.ContentType,
			auditLog.RecipientEmail, auditLog.RecipientID, auditLog.WatermarkData, auditLog.CreatedAt,
			auditLog.Success, auditLog.ErrorMessage, auditLog.CreatedAt, auditLog.CreatedBy,
		)
		if err != nil {
			// Check if it's a busy error
			if isSQLiteBusyError(err) {
				log.Printf("Database busy on SaveAuditLog attempt %d, retrying in %dms: %v", attempt+1, backoffMs, err)
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
				backoffMs *= 2 // Exponential backoff
				continue
			}
			return err
		}
		return nil // Success
	}

	return fmt.Errorf("failed to save audit log after %d retries due to database busy", maxRetries)
}

// GetAuditLogsByLinkID retrieves audit logs for a specific link
func (r *SQLiteWatermarkRepository) GetAuditLogsByLinkID(linkID string) ([]*models.WatermarkAuditLog, error) {
	query := `
		SELECT audit_id, link_id, attachment_id, watermark_type, content_type,
		       recipient_email, recipient_id, watermark_config, applied_at,
		       success, error_message, created_at, created_by
		FROM advanced_watermark_audit_log 
		WHERE link_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.WatermarkAuditLog
	for rows.Next() {
		var log models.WatermarkAuditLog
		var appliedAt time.Time
		err := rows.Scan(
			&log.AuditID, &log.LinkID, &log.AttachmentID, &log.WatermarkType, &log.ContentType,
			&log.RecipientEmail, &log.RecipientID, &log.WatermarkData, &appliedAt,
			&log.Success, &log.ErrorMessage, &log.CreatedAt, &log.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		// Map applied_at to ProcessingTime (using duration since creation)
		log.ProcessingTime = float64(time.Since(appliedAt).Milliseconds())
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetAuditLogsByAttachmentID retrieves audit logs for a specific attachment
func (r *SQLiteWatermarkRepository) GetAuditLogsByAttachmentID(attachmentID string) ([]*models.WatermarkAuditLog, error) {
	query := `
		SELECT audit_id, link_id, attachment_id, watermark_type, content_type,
		       recipient_email, recipient_id, watermark_config, applied_at,
		       success, error_message, created_at, created_by
		FROM advanced_watermark_audit_log 
		WHERE attachment_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, attachmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.WatermarkAuditLog
	for rows.Next() {
		var log models.WatermarkAuditLog
		var appliedAt time.Time
		err := rows.Scan(
			&log.AuditID, &log.LinkID, &log.AttachmentID, &log.WatermarkType, &log.ContentType,
			&log.RecipientEmail, &log.RecipientID, &log.WatermarkData, &appliedAt,
			&log.Success, &log.ErrorMessage, &log.CreatedAt, &log.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		// Map applied_at to ProcessingTime (using duration since creation)
		log.ProcessingTime = float64(time.Since(appliedAt).Milliseconds())
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetAuditLogsByDateRange retrieves audit logs within a date range
func (r *SQLiteWatermarkRepository) GetAuditLogsByDateRange(startDate, endDate time.Time) ([]*models.WatermarkAuditLog, error) {
	query := `
		SELECT audit_id, link_id, attachment_id, watermark_type, content_type,
		       recipient_email, recipient_id, watermark_config, applied_at,
		       success, error_message, created_at, created_by
		FROM advanced_watermark_audit_log 
		WHERE created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.WatermarkAuditLog
	for rows.Next() {
		var log models.WatermarkAuditLog
		var appliedAt time.Time
		err := rows.Scan(
			&log.AuditID, &log.LinkID, &log.AttachmentID, &log.WatermarkType, &log.ContentType,
			&log.RecipientEmail, &log.RecipientID, &log.WatermarkData, &appliedAt,
			&log.Success, &log.ErrorMessage, &log.CreatedAt, &log.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		// Map applied_at to ProcessingTime (using duration since creation)
		log.ProcessingTime = float64(time.Since(appliedAt).Milliseconds())
		logs = append(logs, &log)
	}

	return logs, nil
}
