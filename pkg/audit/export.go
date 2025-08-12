// =============================================================================
// SECURE EMAIL MVP - AUDIT LOG EXPORT FUNCTIONALITY
// =============================================================================
// Package audit export provides CSV and JSON export capabilities for audit logs.
// =============================================================================

package audit

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExportService handles audit log export functionality
type ExportService struct {
	db           *sql.DB
	exportDir    string
	fileLifetime time.Duration
}

// NewExportService creates a new export service
func NewExportService(db *sql.DB, exportDir string) *ExportService {
	if exportDir == "" {
		exportDir = "/tmp/audit-exports"
	}

	// Ensure export directory exists
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		log.Printf("Warning: Could not create export directory %s: %v", exportDir, err)
	}

	return &ExportService{
		db:           db,
		exportDir:    exportDir,
		fileLifetime: 24 * time.Hour, // Default 24 hour lifetime
	}
}

// CreateExportRequest creates a new export request
func (s *ExportService) CreateExportRequest(ctx context.Context, userID string, exportType string, filter AuditLogFilter) (*ExportRequest, error) {
	exportID := uuid.New().String()
	now := time.Now().UTC()
	expiresAt := now.Add(s.fileLifetime)

	// Convert filter to JSON
	filterJSON := "{}"
	if filterJSONBytes, err := json.Marshal(filter); err == nil {
		filterJSON = string(filterJSONBytes)
	}

	// Convert event types to comma-separated string
	eventTypesStr := ""
	if len(filter.EventTypes) > 0 {
		eventTypeStrings := make([]string, len(filter.EventTypes))
		for i, et := range filter.EventTypes {
			eventTypeStrings[i] = string(et)
		}
		eventTypesStr = strings.Join(eventTypeStrings, ",")
	}

	query := `
		INSERT INTO audit_log_exports (
			export_id, user_id, export_type, date_from, date_to, event_types,
			filters, status, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		exportID, userID, exportType, filter.DateFrom, filter.DateTo,
		eventTypesStr, filterJSON, now, expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create export request: %w", err)
	}

	return &ExportRequest{
		ExportID:   exportID,
		UserID:     userID,
		ExportType: exportType,
		DateFrom:   filter.DateFrom,
		DateTo:     filter.DateTo,
		EventTypes: filter.EventTypes,
		Filters:    filter,
		Status:     "pending",
		CreatedAt:  now,
		ExpiresAt:  &expiresAt,
	}, nil
}

// ProcessExport processes an export request and generates the file
func (s *ExportService) ProcessExport(ctx context.Context, exportID string) error {
	// Get export request
	export, err := s.GetExportRequest(ctx, exportID)
	if err != nil {
		return fmt.Errorf("failed to get export request: %w", err)
	}

	if export.Status != "pending" {
		return fmt.Errorf("export request is not pending")
	}

	// Update status to processing
	err = s.updateExportStatus(ctx, exportID, "processing", nil)
	if err != nil {
		return fmt.Errorf("failed to update export status: %w", err)
	}

	// Generate file based on export type
	var filePath string
	var fileSize int64

	switch export.ExportType {
	case "csv":
		filePath, fileSize, err = s.generateCSVExport(ctx, export)
	case "json":
		filePath, fileSize, err = s.generateJSONExport(ctx, export)
	default:
		err = fmt.Errorf("unsupported export type: %s", export.ExportType)
	}

	if err != nil {
		errorMsg := err.Error()
		s.updateExportStatus(ctx, exportID, "failed", &errorMsg)
		return fmt.Errorf("failed to generate export: %w", err)
	}

	// Update status to completed
	now := time.Now().UTC()
	err = s.updateExportCompleted(ctx, exportID, filePath, fileSize, &now)
	if err != nil {
		return fmt.Errorf("failed to update export completion: %w", err)
	}

	return nil
}

// GetExportRequest retrieves an export request by ID
func (s *ExportService) GetExportRequest(ctx context.Context, exportID string) (*ExportRequest, error) {
	query := `
		SELECT export_id, user_id, export_type, date_from, date_to, event_types,
		       filters, file_path, file_size, status, error_message, created_at,
		       completed_at, expires_at
		FROM audit_log_exports
		WHERE export_id = ?
	`

	var export ExportRequest
	var eventTypesStr, filtersJSON sql.NullString
	var filePath, errorMessage sql.NullString
	var fileSize sql.NullInt64
	var completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, exportID).Scan(
		&export.ExportID, &export.UserID, &export.ExportType, &export.DateFrom, &export.DateTo,
		&eventTypesStr, &filtersJSON, &filePath, &fileSize, &export.Status, &errorMessage,
		&export.CreatedAt, &completedAt, &export.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get export request: %w", err)
	}

	if eventTypesStr.Valid && eventTypesStr.String != "" {
		eventTypeStrings := strings.Split(eventTypesStr.String, ",")
		export.EventTypes = make([]EventType, len(eventTypeStrings))
		for i, et := range eventTypeStrings {
			export.EventTypes[i] = EventType(strings.TrimSpace(et))
		}
	}

	if filtersJSON.Valid && filtersJSON.String != "" {
		if err := json.Unmarshal([]byte(filtersJSON.String), &export.Filters); err != nil {
			log.Printf("Failed to unmarshal export filters: %v", err)
		}
	}

	if filePath.Valid {
		export.FilePath = &filePath.String
	}
	if fileSize.Valid {
		export.FileSize = &fileSize.Int64
	}
	if errorMessage.Valid {
		export.ErrorMessage = &errorMessage.String
	}
	if completedAt.Valid {
		export.CompletedAt = &completedAt.Time
	}

	return &export, nil
}

// GetUserExports retrieves export requests for a user
func (s *ExportService) GetUserExports(ctx context.Context, userID string, limit int) ([]ExportRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT export_id, user_id, export_type, date_from, date_to, event_types,
		       filters, file_path, file_size, status, error_message, created_at,
		       completed_at, expires_at
		FROM audit_log_exports
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user exports: %w", err)
	}
	defer rows.Close()

	var exports []ExportRequest
	for rows.Next() {
		var export ExportRequest
		var eventTypesStr, filtersJSON sql.NullString
		var filePath, errorMessage sql.NullString
		var fileSize sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&export.ExportID, &export.UserID, &export.ExportType, &export.DateFrom, &export.DateTo,
			&eventTypesStr, &filtersJSON, &filePath, &fileSize, &export.Status, &errorMessage,
			&export.CreatedAt, &completedAt, &export.ExpiresAt,
		)
		if err != nil {
			log.Printf("Failed to scan export request: %v", err)
			continue
		}

		if eventTypesStr.Valid && eventTypesStr.String != "" {
			eventTypeStrings := strings.Split(eventTypesStr.String, ",")
			export.EventTypes = make([]EventType, len(eventTypeStrings))
			for i, et := range eventTypeStrings {
				export.EventTypes[i] = EventType(strings.TrimSpace(et))
			}
		}

		if filtersJSON.Valid && filtersJSON.String != "" {
			if err := json.Unmarshal([]byte(filtersJSON.String), &export.Filters); err != nil {
				log.Printf("Failed to unmarshal export filters: %v", err)
			}
		}

		if filePath.Valid {
			export.FilePath = &filePath.String
		}
		if fileSize.Valid {
			export.FileSize = &fileSize.Int64
		}
		if errorMessage.Valid {
			export.ErrorMessage = &errorMessage.String
		}
		if completedAt.Valid {
			export.CompletedAt = &completedAt.Time
		}

		exports = append(exports, export)
	}

	return exports, nil
}

// DeleteExport deletes an export request and its file
func (s *ExportService) DeleteExport(ctx context.Context, exportID string) error {
	// Get export to find file path
	export, err := s.GetExportRequest(ctx, exportID)
	if err != nil {
		return fmt.Errorf("failed to get export request: %w", err)
	}

	// Delete file if it exists
	if export.FilePath != nil && *export.FilePath != "" {
		if err := os.Remove(*export.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete export file %s: %v", *export.FilePath, err)
		}
	}

	// Delete from database
	query := "DELETE FROM audit_log_exports WHERE export_id = ?"
	_, err = s.db.ExecContext(ctx, query, exportID)
	if err != nil {
		return fmt.Errorf("failed to delete export request: %w", err)
	}

	return nil
}

// CleanupExpiredExports removes expired export files and records
func (s *ExportService) CleanupExpiredExports(ctx context.Context) error {
	query := `
		SELECT export_id, file_path
		FROM audit_log_exports
		WHERE expires_at < CURRENT_TIMESTAMP AND file_path IS NOT NULL
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query expired exports: %w", err)
	}
	defer rows.Close()

	var deletedCount int
	for rows.Next() {
		var exportID, filePath string
		if err := rows.Scan(&exportID, &filePath); err != nil {
			log.Printf("Failed to scan expired export: %v", err)
			continue
		}

		// Delete file
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete expired export file %s: %v", filePath, err)
		}

		// Delete database record
		if _, err := s.db.ExecContext(ctx, "DELETE FROM audit_log_exports WHERE export_id = ?", exportID); err != nil {
			log.Printf("Failed to delete expired export record %s: %v", exportID, err)
			continue
		}

		deletedCount++
	}

	if deletedCount > 0 {
		log.Printf("Cleaned up %d expired export files", deletedCount)
	}

	return nil
}

// generateCSVExport generates a CSV export file
func (s *ExportService) generateCSVExport(ctx context.Context, export *ExportRequest) (string, int64, error) {
	// Query events using the filter
	auditService := NewAuditService(s.db)
	query, err := auditService.QueryEvents(ctx, export.Filters, 1, 10000) // Get up to 10k events
	if err != nil {
		return "", 0, fmt.Errorf("failed to query events: %w", err)
	}

	// Create file
	filename := fmt.Sprintf("audit_log_%s_%s.csv", export.ExportID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.exportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Log ID", "Timestamp", "Event Type", "User ID", "IP Address", "User Agent",
		"Related Email ID", "Outcome", "Severity", "Session ID", "Request ID",
		"Country", "City", "Device Type", "Details", "Created At",
	}
	if err := writer.Write(header); err != nil {
		return "", 0, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, event := range query.Events {
		details := "{}"
		if event.Details != nil {
			if detailsBytes, err := json.Marshal(event.Details); err == nil {
				details = string(detailsBytes)
			}
		}

		row := []string{
			event.LogID,
			event.Timestamp.Format(time.RFC3339),
			string(event.EventType),
			safeString(event.UserID),
			safeString(event.IPAddress),
			safeString(event.UserAgent),
			safeString(event.RelatedEmailID),
			string(event.Outcome),
			string(event.Severity),
			safeString(event.SessionID),
			safeString(event.RequestID),
			safeString(event.Country),
			safeString(event.City),
			safeString(event.DeviceType),
			details,
			event.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return "", 0, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return filePath, fileInfo.Size(), nil
}

// generateJSONExport generates a JSON export file
func (s *ExportService) generateJSONExport(ctx context.Context, export *ExportRequest) (string, int64, error) {
	// Query events using the filter
	auditService := NewAuditService(s.db)
	query, err := auditService.QueryEvents(ctx, export.Filters, 1, 10000) // Get up to 10k events
	if err != nil {
		return "", 0, fmt.Errorf("failed to query events: %w", err)
	}

	// Create export data structure
	exportData := map[string]interface{}{
		"export_info": map[string]interface{}{
			"export_id":    export.ExportID,
			"export_type":  export.ExportType,
			"created_at":   export.CreatedAt,
			"date_from":    export.DateFrom,
			"date_to":      export.DateTo,
			"event_types":  export.EventTypes,
			"total_events": len(query.Events),
		},
		"events": query.Events,
	}

	// Create file
	filename := fmt.Sprintf("audit_log_%s_%s.json", export.ExportID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(s.exportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Write JSON data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		return "", 0, fmt.Errorf("failed to write JSON data: %w", err)
	}

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return filePath, fileInfo.Size(), nil
}

// updateExportStatus updates the status of an export request
func (s *ExportService) updateExportStatus(ctx context.Context, exportID, status string, errorMessage *string) error {
	query := "UPDATE audit_log_exports SET status = ?, error_message = ? WHERE export_id = ?"
	_, err := s.db.ExecContext(ctx, query, status, errorMessage, exportID)
	return err
}

// updateExportCompleted updates the completion status of an export request
func (s *ExportService) updateExportCompleted(ctx context.Context, exportID, filePath string, fileSize int64, completedAt *time.Time) error {
	query := `
		UPDATE audit_log_exports 
		SET status = 'completed', file_path = ?, file_size = ?, completed_at = ?
		WHERE export_id = ?
	`
	_, err := s.db.ExecContext(ctx, query, filePath, fileSize, completedAt, exportID)
	return err
}

// safeString safely converts a pointer to string
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
