package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ErrorResponse represents the standardized error response schema
type ErrorResponse struct {
	Error     bool                   `json:"error"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Path      string                 `json:"path,omitempty"`
}

// ErrorCode represents predefined error codes for consistent error handling
type ErrorCode string

const (
	// Authentication Errors
	ErrorCodeAuthRequired           ErrorCode = "AUTH_REQUIRED"
	ErrorCodeAuthInvalidToken       ErrorCode = "AUTH_INVALID_TOKEN"
	ErrorCodeAuthExpiredToken       ErrorCode = "AUTH_EXPIRED_TOKEN"
	ErrorCodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrorCodeAuthAccountLocked      ErrorCode = "AUTH_ACCOUNT_LOCKED"
	ErrorCodeAuthMFARequired        ErrorCode = "AUTH_MFA_REQUIRED"
	ErrorCodeAuthMFAInvalid         ErrorCode = "AUTH_MFA_INVALID"

	// Authorization Errors
	ErrorCodeForbidden               ErrorCode = "FORBIDDEN"
	ErrorCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeAdminRequired           ErrorCode = "ADMIN_REQUIRED"

	// Validation Errors
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrorCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrorCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrorCodeFieldTooLong     ErrorCode = "FIELD_TOO_LONG"
	ErrorCodeFieldTooShort    ErrorCode = "FIELD_TOO_SHORT"

	// Resource Errors
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrorCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrorCodeEmailNotFound    ErrorCode = "EMAIL_NOT_FOUND"

	// Rate Limiting Errors
	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodeTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"

	// Server Errors
	ErrorCodeInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeStorageError       ErrorCode = "STORAGE_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Email Specific Errors
	ErrorCodeEmailExpired          ErrorCode = "EMAIL_EXPIRED"
	ErrorCodeEmailAlreadyRead      ErrorCode = "EMAIL_ALREADY_READ"
	ErrorCodeEmailSelfDestructed   ErrorCode = "EMAIL_SELF_DESTRUCTED"
	ErrorCodeEmailAccessDenied     ErrorCode = "EMAIL_ACCESS_DENIED"
	ErrorCodeEmailPasswordRequired ErrorCode = "EMAIL_PASSWORD_REQUIRED"
	ErrorCodeEmailPasswordInvalid  ErrorCode = "EMAIL_PASSWORD_INVALID"

	// Geolocation Errors
	ErrorCodeGeolocationBlocked ErrorCode = "GEOLOCATION_BLOCKED"
	ErrorCodeGeolocationInvalid ErrorCode = "GEOLOCATION_INVALID"

	// ZKID Errors
	ErrorCodeZKIDError           ErrorCode = "ZKID_ERROR"
	ErrorCodeZKIDMappingNotFound ErrorCode = "ZKID_MAPPING_NOT_FOUND"

	// PQC Errors
	ErrorCodePQCError            ErrorCode = "PQC_ERROR"
	ErrorCodePQCKeyNotFound      ErrorCode = "PQC_KEY_NOT_FOUND"
	ErrorCodePQCEncryptionFailed ErrorCode = "PQC_ENCRYPTION_FAILED"
)

// NewErrorResponse creates a new standardized error response
func NewErrorResponse(code ErrorCode, message string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error:     true,
		Code:      string(code),
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
}

// WriteErrorResponse writes a standardized error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details map[string]interface{}) {
	response := NewErrorResponse(code, message, details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(response)
}

// WriteErrorResponseWithPath writes a standardized error response with request path
func WriteErrorResponseWithPath(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details map[string]interface{}, path string) {
	response := NewErrorResponse(code, message, details)
	response.Path = path

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(response)
}

// ErrorMiddleware is middleware that standardizes error responses
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer that captures errors
		errorWriter := &ErrorResponseWriter{
			ResponseWriter: w,
			Request:        r,
		}

		next.ServeHTTP(errorWriter, r)
	})
}

// ErrorResponseWriter wraps http.ResponseWriter to capture and standardize errors
type ErrorResponseWriter struct {
	http.ResponseWriter
	Request *http.Request
	written bool
}

// WriteHeader overrides the default WriteHeader to standardize error responses
func (w *ErrorResponseWriter) WriteHeader(statusCode int) {
	if statusCode >= 400 && !w.written {
		// Check if we're in test mode for debug-friendly error messages
		testMode := os.Getenv("TEST_MODE") == "true"

		// Determine error code based on status code
		var code ErrorCode
		var message string

		switch statusCode {
		case http.StatusBadRequest:
			code = ErrorCodeInvalidRequest
			if testMode {
				message = "Bad Request - Check request format and required fields"
			} else {
				message = "Invalid request"
			}
		case http.StatusUnauthorized:
			code = ErrorCodeAuthRequired
			if testMode {
				message = "Authentication Failed - Check credentials, password, and TOTP code"
			} else {
				message = "Authentication required"
			}
		case http.StatusForbidden:
			code = ErrorCodeForbidden
			if testMode {
				message = "Access Forbidden - Insufficient permissions or blocked"
			} else {
				message = "Access forbidden"
			}
		case http.StatusNotFound:
			code = ErrorCodeNotFound
			if testMode {
				message = "Resource Not Found - Check URL and resource existence"
			} else {
				message = "Resource not found"
			}
		case http.StatusTooManyRequests:
			code = ErrorCodeRateLimitExceeded
			if testMode {
				message = "Rate Limit Exceeded - Too many requests, try again later"
			} else {
				message = "Rate limit exceeded"
			}
		case http.StatusInternalServerError:
			code = ErrorCodeInternalServer
			if testMode {
				message = "Internal Server Error - Check server logs for details"
			} else {
				message = "Internal server error"
			}
		case http.StatusServiceUnavailable:
			code = ErrorCodeServiceUnavailable
			if testMode {
				message = "Service Unavailable - Server temporarily unavailable"
			} else {
				message = "Service unavailable"
			}
		default:
			code = ErrorCodeInternalServer
			if testMode {
				message = fmt.Sprintf("HTTP %d Error - Unexpected status code", statusCode)
			} else {
				message = fmt.Sprintf("HTTP %d error", statusCode)
			}
		}

		// Add debug information in test mode
		var details map[string]interface{}
		if testMode {
			details = map[string]interface{}{
				"debug_info": map[string]interface{}{
					"status_code": statusCode,
					"path":        w.Request.URL.Path,
					"method":      w.Request.Method,
					"timestamp":   time.Now().UTC(),
					"test_mode":   true,
				},
			}
		}

		WriteErrorResponseWithPath(w.ResponseWriter, statusCode, code, message, details, w.Request.URL.Path)
		w.written = true
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

// Write overrides the default Write to prevent double writing
func (w *ErrorResponseWriter) Write(data []byte) (int, error) {
	if w.written {
		return 0, nil
	}
	w.written = true
	return w.ResponseWriter.Write(data)
}

// Helper functions for common error scenarios

// WriteAuthError writes authentication-related error responses
func WriteAuthError(w http.ResponseWriter, code ErrorCode, message string) {
	WriteErrorResponse(w, http.StatusUnauthorized, code, message, nil)
}

// WriteValidationError writes validation error responses
func WriteValidationError(w http.ResponseWriter, field string, message string) {
	details := map[string]interface{}{
		"field": field,
	}
	WriteErrorResponse(w, http.StatusBadRequest, ErrorCodeValidationFailed, message, details)
}

// WriteNotFoundError writes not found error responses
func WriteNotFoundError(w http.ResponseWriter, resource string) {
	message := fmt.Sprintf("%s not found", resource)
	details := map[string]interface{}{
		"resource": resource,
	}
	WriteErrorResponse(w, http.StatusNotFound, ErrorCodeResourceNotFound, message, details)
}

// WriteRateLimitError writes rate limiting error responses
func WriteRateLimitError(w http.ResponseWriter, retryAfter int) {
	message := "Rate limit exceeded. Please try again later."
	details := map[string]interface{}{
		"retry_after_seconds": retryAfter,
	}
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	WriteErrorResponse(w, http.StatusTooManyRequests, ErrorCodeRateLimitExceeded, message, details)
}

// WriteInternalError writes internal server error responses
func WriteInternalError(w http.ResponseWriter, err error) {
	message := "Internal server error"
	details := map[string]interface{}{
		"error": err.Error(),
	}
	WriteErrorResponse(w, http.StatusInternalServerError, ErrorCodeInternalServer, message, details)
}
