package pqc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// AuditLogger handles PQC operation logging for compliance and security monitoring
type AuditLogger struct {
	enabled bool
	mu      sync.Mutex
	logFile *os.File
}

// AuditEvent represents a PQC audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Severity    string                 `json:"severity"` // INFO, WARN, ERROR, CRITICAL
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
}

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(enabled bool) *AuditLogger {
	al := &AuditLogger{
		enabled: enabled,
	}

	if enabled {
		// Open audit log file
		logFile, err := os.OpenFile("/var/log/pqc_audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to stderr if file cannot be opened
			log.Printf("Warning: Could not open PQC audit log file: %v", err)
		} else {
			al.logFile = logFile
		}
	}

	return al
}

// LogEvent logs a PQC audit event
func (al *AuditLogger) LogEvent(eventType, description string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	event := &AuditEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Details:     details,
		Severity:    al.determineSeverity(eventType),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	// Write to log file if available
	if al.logFile != nil {
		if _, err := al.logFile.Write(append(jsonData, '\n')); err != nil {
			log.Printf("Failed to write to audit log file: %v", err)
		}
	}

	// Also log to standard logger for immediate visibility
	log.Printf("[PQC_AUDIT] %s: %s", eventType, description)
}

// LogEventWithContext logs a PQC audit event with user context
func (al *AuditLogger) LogEventWithContext(eventType, description string, details map[string]interface{}, userID, ipAddress, sessionID string) {
	if !al.enabled {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	event := &AuditEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Details:     details,
		Severity:    al.determineSeverity(eventType),
		UserID:      userID,
		IPAddress:   ipAddress,
		SessionID:   sessionID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	// Write to log file if available
	if al.logFile != nil {
		if _, err := al.logFile.Write(append(jsonData, '\n')); err != nil {
			log.Printf("Failed to write to audit log file: %v", err)
		}
	}

	// Also log to standard logger for immediate visibility
	log.Printf("[PQC_AUDIT] %s: %s (User: %s, IP: %s)", eventType, description, userID, ipAddress)
}

// LogSecurityEvent logs a security-related PQC event
func (al *AuditLogger) LogSecurityEvent(eventType, description string, details map[string]interface{}, severity string) {
	if !al.enabled {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	event := &AuditEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Details:     details,
		Severity:    severity,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal security event: %v", err)
		return
	}

	// Write to log file if available
	if al.logFile != nil {
		if _, err := al.logFile.Write(append(jsonData, '\n')); err != nil {
			log.Printf("Failed to write to audit log file: %v", err)
		}
	}

	// Also log to standard logger for immediate visibility
	log.Printf("[PQC_SECURITY] %s: %s (Severity: %s)", eventType, description, severity)
}

// LogKeyOperation logs a key management operation
func (al *AuditLogger) LogKeyOperation(operation, keyID string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	eventType := fmt.Sprintf("KEY_%s", operation)
	description := fmt.Sprintf("Key operation: %s for key ID: %s", operation, keyID)

	al.LogEvent(eventType, description, details)
}

// LogEncryptionOperation logs an encryption operation
func (al *AuditLogger) LogEncryptionOperation(operation, context string, dataSize int, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	eventType := fmt.Sprintf("ENCRYPTION_%s", operation)
	description := fmt.Sprintf("Encryption operation: %s for context: %s (data size: %d bytes)", operation, context, dataSize)

	if details == nil {
		details = make(map[string]interface{})
	}
	details["data_size"] = dataSize
	details["context"] = context

	al.LogEvent(eventType, description, details)
}

// LogDecryptionOperation logs a decryption operation
func (al *AuditLogger) LogDecryptionOperation(operation, context string, dataSize int, method string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	eventType := fmt.Sprintf("DECRYPTION_%s", operation)
	description := fmt.Sprintf("Decryption operation: %s for context: %s using %s (data size: %d bytes)", operation, context, method, dataSize)

	if details == nil {
		details = make(map[string]interface{})
	}
	details["data_size"] = dataSize
	details["context"] = context
	details["method"] = method

	al.LogEvent(eventType, description, details)
}

// LogHSMOperation logs an HSM operation
func (al *AuditLogger) LogHSMOperation(operation, hsmKeyID string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	eventType := fmt.Sprintf("HSM_%s", operation)
	description := fmt.Sprintf("HSM operation: %s for HSM key ID: %s", operation, hsmKeyID)

	if details == nil {
		details = make(map[string]interface{})
	}
	details["hsm_key_id"] = hsmKeyID

	al.LogEvent(eventType, description, details)
}

// LogPerformanceEvent logs a performance-related event
func (al *AuditLogger) LogPerformanceEvent(operation string, duration time.Duration, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	eventType := fmt.Sprintf("PERFORMANCE_%s", operation)
	description := fmt.Sprintf("Performance event: %s took %v", operation, duration)

	if details == nil {
		details = make(map[string]interface{})
	}
	details["duration_ms"] = duration.Milliseconds()
	details["duration_ns"] = duration.Nanoseconds()

	al.LogEvent(eventType, description, details)
}

// LogError logs an error event
func (al *AuditLogger) LogError(eventType, description string, err error, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()

	al.LogSecurityEvent(eventType, description, details, "ERROR")
}

// LogWarning logs a warning event
func (al *AuditLogger) LogWarning(eventType, description string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	al.LogSecurityEvent(eventType, description, details, "WARN")
}

// LogCritical logs a critical security event
func (al *AuditLogger) LogCritical(eventType, description string, details map[string]interface{}) {
	if !al.enabled {
		return
	}

	al.LogSecurityEvent(eventType, description, details, "CRITICAL")
}

// determineSeverity determines the severity level based on event type
func (al *AuditLogger) determineSeverity(eventType string) string {
	switch {
	case al.isCriticalEvent(eventType):
		return "CRITICAL"
	case al.isErrorEvent(eventType):
		return "ERROR"
	case al.isWarningEvent(eventType):
		return "WARN"
	default:
		return "INFO"
	}
}

// isCriticalEvent checks if an event type is critical
func (al *AuditLogger) isCriticalEvent(eventType string) bool {
	criticalEvents := []string{
		"KEY_COMPROMISE",
		"HSM_FAILURE",
		"DECRYPTION_FAILURE",
		"UNAUTHORIZED_ACCESS",
		"KEY_ROTATION_FAILURE",
	}

	for _, event := range criticalEvents {
		if eventType == event {
			return true
		}
	}
	return false
}

// isErrorEvent checks if an event type is an error
func (al *AuditLogger) isErrorEvent(eventType string) bool {
	errorEvents := []string{
		"ENCRYPTION_FAILURE",
		"KEY_GENERATION_FAILURE",
		"HSM_OPERATION_FAILURE",
		"CONFIGURATION_ERROR",
	}

	for _, event := range errorEvents {
		if eventType == event {
			return true
		}
	}
	return false
}

// isWarningEvent checks if an event type is a warning
func (al *AuditLogger) isWarningEvent(eventType string) bool {
	warningEvents := []string{
		"KEY_EXPIRING_SOON",
		"PERFORMANCE_DEGRADATION",
		"HSM_SLOW_RESPONSE",
		"KEY_ROTATION_DUE",
	}

	for _, event := range warningEvents {
		if eventType == event {
			return true
		}
	}
	return false
}

// Close closes the audit logger and its file handle
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.logFile != nil {
		return al.logFile.Close()
	}
	return nil
}

// IsEnabled returns whether audit logging is enabled
func (al *AuditLogger) IsEnabled() bool {
	return al.enabled
}

// GetStats returns audit logger statistics
func (al *AuditLogger) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": al.enabled,
		"log_file": func() string {
			if al.logFile != nil {
				return al.logFile.Name()
			}
			return "stderr"
		}(),
	}
}
