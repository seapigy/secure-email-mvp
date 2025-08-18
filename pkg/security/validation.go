package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Common attack patterns
var (
	// SQL Injection patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`),
		regexp.MustCompile(`(?i)(--|/\*|\*/|xp_|sp_|@@|char\(|nchar\(|varchar\(|nvarchar\(|cast\(|convert\()`),
		regexp.MustCompile(`(?i)(waitfor|delay|sleep|benchmark|load_file|into\s+outfile|into\s+dumpfile)`),
		regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|or\s+true|or\s+false|and\s+1\s*=\s*1|and\s+true|and\s+false)`),
		regexp.MustCompile(`(?i)(';|";|'--|"--|'/\*|"/\*|'union|"union|'select|"select)`),
	}

	// XSS patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload\s*=|onerror\s*=|onclick\s*=|onmouseover\s*=|onfocus\s*=|onblur\s*=)`),
		regexp.MustCompile(`(?i)(<iframe|<object|<embed|<form|<input|<textarea|<select|<button|<link|<meta|<style)`),
		regexp.MustCompile(`(?i)(alert\(|confirm\(|prompt\(|eval\(|setTimeout\(|setInterval\(|Function\(|document\.|window\.|location\.|history\.)`),
		regexp.MustCompile(`(?i)(data:text/html|data:application/x-javascript|vbscript:|javascript:|file:|ftp:|gopher:|mailto:|news:|telnet:)`),
	}

	// CSV Injection patterns
	csvInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^[=+\-@\t\r\n]`), // Starts with =, +, -, @, tab, newline
		regexp.MustCompile(`(?i)(=cmd|'cmd|"cmd|=powershell|'powershell|"powershell)`),
		regexp.MustCompile(`(?i)(=http|'http|"http|=ftp|'ftp|"ftp|=file|'file|"file)`),
		regexp.MustCompile(`(?i)(=javascript|'javascript|"javascript|=vbscript|'vbscript|"vbscript)`),
	}

	// Path traversal patterns
	pathTraversalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\.\./|\.\.\\|\.\.%2f|\.\.%5c|\.\.%2F|\.\.%5C)`),
		regexp.MustCompile(`(?i)(%2e%2e%2f|%2e%2e%5c|%2e%2e%2F|%2e%2e%5C)`),
		regexp.MustCompile(`(?i)(\.\.%252f|\.\.%255c|\.\.%252F|\.\.%255C)`),
	}

	// Command injection patterns
	commandInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile("(?i)(;|\\||&|\\$\\(|`|>|<|\\$IFS|\\$PATH|\\$PWD|\\$HOME|\\$USER)"),
		regexp.MustCompile(`(?i)(cmd|powershell|bash|sh|zsh|ksh|tcsh|fish|dash|ash|busybox)`),
		regexp.MustCompile(`(?i)(cat|ls|dir|pwd|whoami|id|uname|hostname|ps|top|kill|chmod|chown)`),
		regexp.MustCompile(`(?i)(wget|curl|nc|netcat|telnet|ssh|scp|rsync|ftp|sftp|ncftp)`),
	}

	// Valid UUID pattern
	uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Valid email pattern
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Valid organization name pattern (alphanumeric, spaces, hyphens, underscores)
	orgNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]{1,100}$`)

	// Valid action pattern for compliance logs
	actionPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid    bool     `json:"is_valid"`
	Errors     []string `json:"errors,omitempty"`
	Sanitized  string   `json:"sanitized,omitempty"`
	ThreatType string   `json:"threat_type,omitempty"`
}

// ValidateAndSanitizeInput validates and sanitizes input to prevent common attacks
func ValidateAndSanitizeInput(input string, inputType string, maxLength int) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Check for empty input
	if strings.TrimSpace(input) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Input cannot be empty")
		return result
	}

	// Check length
	if len(input) > maxLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input exceeds maximum length of %d characters", maxLength))
		return result
	}

	// Check for SQL injection
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			result.IsValid = false
			result.ThreatType = "sql_injection"
			result.Errors = append(result.Errors, "SQL injection attempt detected")
			return result
		}
	}

	// Check for XSS
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			result.IsValid = false
			result.ThreatType = "xss"
			result.Errors = append(result.Errors, "XSS attempt detected")
			return result
		}
	}

	// Check for CSV injection
	for _, pattern := range csvInjectionPatterns {
		if pattern.MatchString(input) {
			result.IsValid = false
			result.ThreatType = "csv_injection"
			result.Errors = append(result.Errors, "CSV injection attempt detected")
			return result
		}
	}

	// Check for path traversal
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			result.IsValid = false
			result.ThreatType = "path_traversal"
			result.Errors = append(result.Errors, "Path traversal attempt detected")
			return result
		}
	}

	// Check for command injection
	for _, pattern := range commandInjectionPatterns {
		if pattern.MatchString(input) {
			result.IsValid = false
			result.ThreatType = "command_injection"
			result.Errors = append(result.Errors, "Command injection attempt detected")
			return result
		}
	}

	// Type-specific validation
	switch inputType {
	case "email":
		if !emailPattern.MatchString(input) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid email format")
		}
	case "uuid":
		if !uuidPattern.MatchString(input) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid UUID format")
		}
	case "org_name":
		if !orgNamePattern.MatchString(input) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid organization name format")
		}
	case "action":
		if !actionPattern.MatchString(input) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid action format")
		}
	case "integer":
		if _, err := strconv.Atoi(input); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid integer format")
		}
	case "date":
		// Basic date format validation (YYYY-MM-DD)
		datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !datePattern.MatchString(input) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Invalid date format (expected YYYY-MM-DD)")
		}
	}

	// Sanitize the input if valid
	if result.IsValid {
		result.Sanitized = sanitizeInput(input)
	}

	return result
}

// sanitizeInput removes potentially dangerous characters and normalizes input
func sanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	var sanitized strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			continue
		}
		sanitized.WriteRune(r)
	}

	// Trim whitespace
	result := strings.TrimSpace(sanitized.String())

	// Normalize line endings
	result = strings.ReplaceAll(result, "\r\n", "\n")
	result = strings.ReplaceAll(result, "\r", "\n")

	return result
}

// ValidateUUID validates if a string is a valid UUID
func ValidateUUID(uuid string) bool {
	return uuidPattern.MatchString(uuid)
}

// ValidateEmail validates if a string is a valid email address
func ValidateEmail(email string) bool {
	return emailPattern.MatchString(email)
}

// ValidateOrganizationName validates if a string is a valid organization name
func ValidateOrganizationName(name string) bool {
	return orgNamePattern.MatchString(name)
}

// ValidateAction validates if a string is a valid action name
func ValidateAction(action string) bool {
	return actionPattern.MatchString(action)
}

// ValidateInteger validates if a string is a valid integer
func ValidateInteger(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

// ValidateDate validates if a string is a valid date in YYYY-MM-DD format
func ValidateDate(date string) bool {
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	return datePattern.MatchString(date)
}

// SanitizeCSVValue sanitizes a value for CSV export to prevent injection
func SanitizeCSVValue(value string) string {
	// Remove null bytes
	value = strings.ReplaceAll(value, "\x00", "")

	// Escape quotes
	value = strings.ReplaceAll(value, `"`, `""`)

	// Remove control characters
	var sanitized strings.Builder
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			continue
		}
		sanitized.WriteRune(r)
	}

	result := sanitized.String()

	// Check if the value needs to be quoted (contains comma, newline, or quote)
	if strings.ContainsAny(result, ",\"\n\r") {
		result = `"` + result + `"`
	}

	return result
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32 // Default length
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %v", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateJWTToken validates JWT token format and structure
func ValidateJWTToken(token string) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Check if token is empty
	if strings.TrimSpace(token) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "JWT token cannot be empty")
		return result
	}

	// Check if token has the correct format (3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		result.IsValid = false
		result.ThreatType = "jwt_tampering"
		result.Errors = append(result.Errors, "Invalid JWT format")
		return result
	}

	// Validate each part
	for i, part := range parts {
		if part == "" {
			result.IsValid = false
			result.ThreatType = "jwt_tampering"
			result.Errors = append(result.Errors, fmt.Sprintf("JWT part %d cannot be empty", i+1))
			return result
		}

		// Check for valid base64 characters
		for _, r := range part {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '=') {
				result.IsValid = false
				result.ThreatType = "jwt_tampering"
				result.Errors = append(result.Errors, fmt.Sprintf("JWT part %d contains invalid characters", i+1))
				return result
			}
		}
	}

	return result
}

// ValidateTOTPCode validates TOTP code format
func ValidateTOTPCode(code string) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Check if code is empty
	if strings.TrimSpace(code) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "TOTP code cannot be empty")
		return result
	}

	// Check length (typically 6 digits)
	if len(code) != 6 {
		result.IsValid = false
		result.Errors = append(result.Errors, "TOTP code must be 6 digits")
		return result
	}

	// Check if all characters are digits
	for _, r := range code {
		if r < '0' || r > '9' {
			result.IsValid = false
			result.Errors = append(result.Errors, "TOTP code must contain only digits")
			return result
		}
	}

	return result
}

// ValidatePassword validates password strength
func ValidatePassword(password string) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Check minimum length
	if len(password) < 8 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must be at least 8 characters long")
	}

	// Check maximum length
	if len(password) > 128 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must be no more than 128 characters long")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123", "password123",
		"admin", "root", "user", "test", "guest", "welcome", "login",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			result.IsValid = false
			result.Errors = append(result.Errors, "Password is too common")
			break
		}
	}

	// Check for control characters
	for _, r := range password {
		if unicode.IsControl(r) {
			result.IsValid = false
			result.Errors = append(result.Errors, "Password contains invalid characters")
			break
		}
	}

	return result
}

// ValidateRateLimit validates rate limiting parameters
func ValidateRateLimit(limit, window int) ValidationResult {
	result := ValidationResult{
		IsValid: true,
		Errors:  []string{},
	}

	// Validate limit
	if limit <= 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Rate limit must be greater than 0")
	}

	if limit > 10000 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Rate limit cannot exceed 10000")
	}

	// Validate window (in seconds)
	if window <= 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Rate limit window must be greater than 0")
	}

	if window > 86400 { // 24 hours
		result.IsValid = false
		result.Errors = append(result.Errors, "Rate limit window cannot exceed 24 hours")
	}

	return result
}

// IsSuspiciousInput checks if input contains suspicious patterns
func IsSuspiciousInput(input string) (bool, string) {
	// Check for SQL injection
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true, "sql_injection"
		}
	}

	// Check for XSS
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true, "xss"
		}
	}

	// Check for CSV injection
	for _, pattern := range csvInjectionPatterns {
		if pattern.MatchString(input) {
			return true, "csv_injection"
		}
	}

	// Check for path traversal
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			return true, "path_traversal"
		}
	}

	// Check for command injection
	for _, pattern := range commandInjectionPatterns {
		if pattern.MatchString(input) {
			return true, "command_injection"
		}
	}

	return false, ""
}
