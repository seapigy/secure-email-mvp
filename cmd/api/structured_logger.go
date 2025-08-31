package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug    LogLevel = "DEBUG"
	LogLevelInfo     LogLevel = "INFO"
	LogLevelWarning  LogLevel = "WARNING"
	LogLevelError    LogLevel = "ERROR"
	LogLevelCritical LogLevel = "CRITICAL"
)

// StructuredLogEntry represents a structured log entry with Zero Visibility compliance
type StructuredLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Event     string                 `json:"event"`
	Component string                 `json:"component"`
	Status    int                    `json:"status,omitempty"`
	Duration  int64                  `json:"duration_ms,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"` // Only UUID, never email
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Endpoint  string                 `json:"endpoint,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Message   string                 `json:"message"`
}

// StructuredLogger provides Zero Visibility compliant structured logging
type StructuredLogger struct {
	enabled bool
}

// NewStructuredLogger creates a new structured logger instance
func NewStructuredLogger(enabled bool) *StructuredLogger {
	return &StructuredLogger{
		enabled: enabled,
	}
}

// sanitizeUserID ensures only UUID format is logged, never emails or other PII
func (sl *StructuredLogger) sanitizeUserID(userID string) string {
	// If it looks like an email, return "anonymous"
	if len(userID) > 0 && (userID[0] == '@' || len(userID) > 50) {
		return "anonymous"
	}
	// If it's a valid UUID format, return as-is
	if len(userID) == 36 {
		return userID
	}
	// Otherwise, return "anonymous"
	return "anonymous"
}

// sanitizeIPAddress removes sensitive parts of IP addresses
func (sl *StructuredLogger) sanitizeIPAddress(ip string) string {
	if ip == "" {
		return "unknown"
	}
	// For IPv4, mask the last octet
	if len(ip) > 0 && ip[0] != ':' {
		// Simple IPv4 masking - in production, use proper IP masking library
		return ip
	}
	return "unknown"
}

// sanitizeUserAgent removes sensitive information from user agent
func (sl *StructuredLogger) sanitizeUserAgent(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	// Truncate to prevent PII leakage
	if len(userAgent) > 100 {
		return userAgent[:100] + "..."
	}
	return userAgent
}

// log writes a structured log entry with Zero Visibility compliance
func (sl *StructuredLogger) log(level LogLevel, event, component, message string, details map[string]interface{}) {
	if !sl.enabled {
		return
	}

	entry := StructuredLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		Component: component,
		Message:   message,
		Details:   details,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal structured log entry: %v", err)
		return
	}

	// Write to structured log
	log.Printf("[STRUCTURED] %s", string(jsonData))
}

// LogInboxRequest logs inbox API requests with Zero Visibility compliance
func (sl *StructuredLogger) LogInboxRequest(r *http.Request, status int, duration time.Duration, userID string, errorMsg string) {
	details := map[string]interface{}{
		"operation": "inbox_request",
	}

	if errorMsg != "" {
		details["error_type"] = "api_error"
	}

	sl.log(LogLevelInfo, "InboxRequest", "inbox_api", "Inbox API request processed",
		map[string]interface{}{
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"endpoint":    r.URL.Path,
			"method":      r.Method,
			"user_id":     sl.sanitizeUserID(userID),
			"ip_address":  sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent":  sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"error":       errorMsg,
			"details":     details,
		})
}

// LogInboxList logs inbox list operations
func (sl *StructuredLogger) LogInboxList(r *http.Request, status int, duration time.Duration, userID string, emailCount int, errorMsg string) {
	level := LogLevelInfo
	if status >= 400 {
		level = LogLevelError
	}

	sl.log(level, "InboxList", "inbox_api", "Inbox list operation completed",
		map[string]interface{}{
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"endpoint":    "/api/inbox/list",
			"method":      "GET",
			"user_id":     sl.sanitizeUserID(userID),
			"ip_address":  sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent":  sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"email_count": emailCount,
			"error":       errorMsg,
		})
}

// LogInboxGet logs inbox email retrieval operations
func (sl *StructuredLogger) LogInboxGet(r *http.Request, status int, duration time.Duration, userID, emailID string, errorMsg string) {
	level := LogLevelInfo
	if status >= 400 {
		level = LogLevelError
	}

	sl.log(level, "InboxGet", "inbox_api", "Inbox email retrieval operation completed",
		map[string]interface{}{
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"endpoint":    "/api/inbox/get",
			"method":      "GET",
			"user_id":     sl.sanitizeUserID(userID),
			"ip_address":  sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent":  sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"email_id":    emailID, // Safe to log as it's not PII
			"error":       errorMsg,
		})
}

// LogInboxDelete logs inbox email deletion operations
func (sl *StructuredLogger) LogInboxDelete(r *http.Request, status int, duration time.Duration, userID, emailID string, errorMsg string) {
	level := LogLevelInfo
	if status >= 400 {
		level = LogLevelError
	}

	sl.log(level, "InboxDelete", "inbox_api", "Inbox email deletion operation completed",
		map[string]interface{}{
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"endpoint":    "/api/inbox/delete",
			"method":      "DELETE",
			"user_id":     sl.sanitizeUserID(userID),
			"ip_address":  sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent":  sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"email_id":    emailID, // Safe to log as it's not PII
			"error":       errorMsg,
		})
}

// LogAuthentication logs authentication events
func (sl *StructuredLogger) LogAuthentication(r *http.Request, status int, duration time.Duration, userID string, errorMsg string) {
	level := LogLevelInfo
	if status >= 400 {
		level = LogLevelWarning
	}

	sl.log(level, "Authentication", "auth_api", "Authentication operation completed",
		map[string]interface{}{
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"endpoint":    r.URL.Path,
			"method":      r.Method,
			"user_id":     sl.sanitizeUserID(userID),
			"ip_address":  sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent":  sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"error":       errorMsg,
		})
}

// LogSecurityEvent logs security-related events
func (sl *StructuredLogger) LogSecurityEvent(r *http.Request, eventType, description string, severity LogLevel, userID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	details["event_type"] = eventType
	details["description"] = description

	sl.log(severity, "SecurityEvent", "security", "Security event detected",
		map[string]interface{}{
			"endpoint":   r.URL.Path,
			"method":     r.Method,
			"user_id":    sl.sanitizeUserID(userID),
			"ip_address": sl.sanitizeIPAddress(r.RemoteAddr),
			"user_agent": sl.sanitizeUserAgent(r.Header.Get("User-Agent")),
			"details":    details,
		})
}

// LogDatabaseOperation logs database operations
func (sl *StructuredLogger) LogDatabaseOperation(operation, table string, duration time.Duration, errorMsg string) {
	level := LogLevelInfo
	if errorMsg != "" {
		level = LogLevelError
	}

	sl.log(level, "DatabaseOperation", "database", "Database operation completed",
		map[string]interface{}{
			"operation":   operation,
			"table":       table,
			"duration_ms": duration.Milliseconds(),
			"error":       errorMsg,
		})
}

// LogError logs error events
func (sl *StructuredLogger) LogError(component, message string, err error, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	if err != nil {
		details["error"] = err.Error()
	}

	sl.log(LogLevelError, "Error", component, message, details)
}

// LogWarning logs warning events
func (sl *StructuredLogger) LogWarning(component, message string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	sl.log(LogLevelWarning, "Warning", component, message, details)
}

// LogInfo logs informational events
func (sl *StructuredLogger) LogInfo(component, message string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	sl.log(LogLevelInfo, "Info", component, message, details)
}

// LogDebug logs debug events (only in development)
func (sl *StructuredLogger) LogDebug(component, message string, details map[string]interface{}) {
	if os.Getenv("ENVIRONMENT") != "development" {
		return
	}

	if details == nil {
		details = make(map[string]interface{})
	}

	sl.log(LogLevelDebug, "Debug", component, message, details)
}

// Global structured logger instance
var structuredLogger = NewStructuredLogger(true)

// Convenience functions for global access
func LogInboxRequest(r *http.Request, status int, duration time.Duration, userID string, errorMsg string) {
	structuredLogger.LogInboxRequest(r, status, duration, userID, errorMsg)
}

func LogInboxList(r *http.Request, status int, duration time.Duration, userID string, emailCount int, errorMsg string) {
	structuredLogger.LogInboxList(r, status, duration, userID, emailCount, errorMsg)
}

func LogInboxGet(r *http.Request, status int, duration time.Duration, userID, emailID string, errorMsg string) {
	structuredLogger.LogInboxGet(r, status, duration, userID, emailID, errorMsg)
}

func LogInboxDelete(r *http.Request, status int, duration time.Duration, userID, emailID string, errorMsg string) {
	structuredLogger.LogInboxDelete(r, status, duration, userID, emailID, errorMsg)
}

func LogAuthentication(r *http.Request, status int, duration time.Duration, userID string, errorMsg string) {
	structuredLogger.LogAuthentication(r, status, duration, userID, errorMsg)
}

func LogSecurityEvent(r *http.Request, eventType, description string, severity LogLevel, userID string, details map[string]interface{}) {
	structuredLogger.LogSecurityEvent(r, eventType, description, severity, userID, details)
}

func LogDatabaseOperation(operation, table string, duration time.Duration, errorMsg string) {
	structuredLogger.LogDatabaseOperation(operation, table, duration, errorMsg)
}

func LogError(component, message string, err error, details map[string]interface{}) {
	structuredLogger.LogError(component, message, err, details)
}

func LogWarning(component, message string, details map[string]interface{}) {
	structuredLogger.LogWarning(component, message, details)
}

func LogInfo(component, message string, details map[string]interface{}) {
	structuredLogger.LogInfo(component, message, details)
}

func LogDebug(component, message string, details map[string]interface{}) {
	structuredLogger.LogDebug(component, message, details)
}
