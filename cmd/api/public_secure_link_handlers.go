package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"secure-email-mvp/pkg/errors"
	"secure-email-mvp/pkg/geolocation"
	"secure-email-mvp/pkg/securelinks/security"

	"github.com/gorilla/mux"
)

// =============================================================================
// PUBLIC SECURE LINK HANDLERS
// =============================================================================

// SecureLinkMetadata represents the metadata returned for a secure link
type SecureLinkMetadata struct {
	LinkID           string                   `json:"link_id"`
	Subject          string                   `json:"subject"`
	SenderEmail      string                   `json:"sender_email"`
	SenderName       string                   `json:"sender_name,omitempty"`
	SecuritySettings SecuritySettingsMetadata `json:"security_settings"`
	Status           string                   `json:"status"`
	Message          string                   `json:"message,omitempty"`
}

// SecuritySettingsMetadata represents security settings for external display
type SecuritySettingsMetadata struct {
	RequirePassword        bool     `json:"require_password"`
	RequireMFA             bool     `json:"require_mfa"`
	MFAType                string   `json:"mfa_type,omitempty"`
	GeolocationRestriction bool     `json:"geolocation_restriction"`
	AllowedCountries       []string `json:"allowed_countries,omitempty"`
	AllowedCities          []string `json:"allowed_cities,omitempty"`
	TimeLock               bool     `json:"time_lock"`
	TimeLockUntil          *int64   `json:"time_lock_until,omitempty"`
	ReadOnce               bool     `json:"read_once"`
	AutoDestruct           bool     `json:"auto_destruct"`
	ExpiresAt              *int64   `json:"expires_at,omitempty"`
	MaxAccessAttempts      int      `json:"max_access_attempts"`
	CurrentAttempts        int      `json:"current_attempts"`
}

// SecurityValidationRequest represents a security validation request
type SecurityValidationRequest struct {
	LinkID    string `json:"link_id"`
	Password  string `json:"password,omitempty"`
	MFACode   string `json:"mfa_code,omitempty"`
	MFAType   string `json:"mfa_type,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// SecurityValidationResponse represents the response from security validation
type SecurityValidationResponse struct {
	Success      bool   `json:"success"`
	Validated    bool   `json:"validated"`
	RequiresMFA  bool   `json:"requires_mfa,omitempty"`
	MFAType      string `json:"mfa_type,omitempty"`
	RequiresGeo  bool   `json:"requires_geo,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	DecoyMessage string `json:"decoy_message,omitempty"`
}

// SecureEmailContent represents the secure email content after validation
type SecureEmailContent struct {
	LinkID       string `json:"link_id"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	SenderEmail  string `json:"sender_email"`
	SenderName   string `json:"sender_name,omitempty"`
	ReadOnce     bool   `json:"read_once"`
	AutoDestruct bool   `json:"auto_destruct"`
}

// publicSecureLinkHandler handles GET /v/{linkID} for public secure link access
func (srv *Server) publicSecureLinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["linkID"]

	if linkID == "" {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Link ID is required", nil)
		return
	}

	// Get client IP and user agent
	clientIP := srv.getClientIP(r)
	userAgent := r.UserAgent()

	log.Printf("🔗 Public secure link access: %s from IP: %s", linkID, clientIP)

	// Get secure link metadata
	metadata, err := srv.getSecureLinkMetadata(r.Context(), linkID, clientIP, userAgent)
	if err != nil {
		log.Printf("❌ Failed to get secure link metadata: %v", err)
		errors.WriteErrorResponse(w, http.StatusNotFound, errors.ErrorCodeNotFound, "Secure link not found or expired", nil)
		return
	}

	// Check if link is accessible
	if metadata.Status != "active" {
		errors.WriteErrorResponse(w, http.StatusForbidden, errors.ErrorCodeForbidden, metadata.Message, nil)
		return
	}

	// Log the access attempt
	srv.logSecureLinkAccess(r.Context(), linkID, clientIP, userAgent, "metadata_requested", "Metadata requested for secure link", nil)

	// Return metadata for frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// validateSecurityHandler handles POST /v/{linkID}/validate for security validation
func (srv *Server) validateSecurityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["linkID"]

	if linkID == "" {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Link ID is required", nil)
		return
	}

	var req SecurityValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Invalid request format", nil)
		return
	}

	req.LinkID = linkID
	req.IPAddress = srv.getClientIP(r)
	req.UserAgent = r.UserAgent()

	log.Printf("🔐 Security validation request for link: %s from IP: %s", linkID, req.IPAddress)

	// Validate security
	response, err := srv.validateSecureLinkSecurity(r.Context(), req)
	if err != nil {
		log.Printf("❌ Security validation failed: %v", err)
		errors.WriteErrorResponse(w, http.StatusInternalServerError, errors.ErrorCodeInternalServer, "Security validation failed", nil)
		return
	}

	// Log the validation attempt
	srv.logSecureLinkAccess(r.Context(), linkID, req.IPAddress, req.UserAgent, "security_validated",
		fmt.Sprintf("Security validation: %t", response.Validated), nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getSecureEmailContentHandler handles POST /v/{linkID}/content for getting email content after validation
func (srv *Server) getSecureEmailContentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["linkID"]

	if linkID == "" {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Link ID is required", nil)
		return
	}

	clientIP := srv.getClientIP(r)
	userAgent := r.UserAgent()

	log.Printf("📧 Secure email content request for link: %s from IP: %s", linkID, clientIP)

	// Get secure email content
	content, err := srv.getSecureEmailContent(r.Context(), linkID, clientIP, userAgent)
	if err != nil {
		log.Printf("❌ Failed to get secure email content: %v", err)
		errors.WriteErrorResponse(w, http.StatusNotFound, errors.ErrorCodeNotFound, "Secure email content not found", nil)
		return
	}

	// Log the content access
	srv.logSecureLinkAccess(r.Context(), linkID, clientIP, userAgent, "content_accessed", "Secure email content accessed", nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// getSecureLinkMetadata retrieves metadata for a secure link
func (srv *Server) getSecureLinkMetadata(ctx context.Context, linkID, clientIP, userAgent string) (*SecureLinkMetadata, error) {
	// Get secure link from database
	var (
		emailID, subject, senderEmail, senderName           string
		requirePassword, requireMFA, readOnce, autoDestruct bool
		mfaType                                             string
		geolocationRestriction                              bool
		allowedCountries, allowedCities                     string
		timeLock                                            bool
		timeLockUntil, expiresAt                            *int64
		maxAccessAttempts, currentAttempts                  int
		status                                              string
	)

	query := `
		SELECT 
			sl.email_id, sl.require_password, sl.require_mfa, sl.mfa_type,
			sl.geolocation_restriction, sl.allowed_countries, sl.allowed_cities,
			sl.time_lock, sl.time_lock_until, sl.read_once, sl.auto_destruct,
			sl.expires_at, sl.max_access_attempts, sl.current_attempts, sl.status,
			e.subject, e.sender_id, u.email as sender_email, u.name as sender_name
		FROM secure_links sl
		JOIN emails e ON sl.email_id = e.email_id
		JOIN users u ON e.sender_id = u.id
		WHERE sl.link_id = ?
	`

	err := srv.db.QueryRowContext(ctx, query, linkID).Scan(
		&emailID, &requirePassword, &requireMFA, &mfaType,
		&geolocationRestriction, &allowedCountries, &allowedCities,
		&timeLock, &timeLockUntil, &readOnce, &autoDestruct,
		&expiresAt, &maxAccessAttempts, &currentAttempts, &status,
		&subject, &senderEmail, &senderName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("secure link not found")
		}
		return nil, fmt.Errorf("failed to query secure link: %w", err)
	}

	// Check if link is expired
	if expiresAt != nil && *expiresAt < time.Now().Unix() {
		return &SecureLinkMetadata{
			LinkID:  linkID,
			Status:  "expired",
			Message: "This secure link has expired",
		}, nil
	}

	// Check if link is time-locked
	if timeLock && timeLockUntil != nil && *timeLockUntil > time.Now().Unix() {
		unlockTime := time.Unix(*timeLockUntil, 0)
		return &SecureLinkMetadata{
			LinkID:  linkID,
			Status:  "time_locked",
			Message: fmt.Sprintf("This secure link will be available after %s", unlockTime.Format("January 2, 2006 at 3:04 PM")),
		}, nil
	}

	// Check if link has exceeded access attempts
	if currentAttempts >= maxAccessAttempts {
		return &SecureLinkMetadata{
			LinkID:  linkID,
			Status:  "destroyed",
			Message: "This secure link has been destroyed due to too many failed attempts",
		}, nil
	}

	// Parse allowed countries and cities
	var countries, cities []string
	if allowedCountries != "" {
		countries = strings.Split(allowedCountries, ",")
	}
	if allowedCities != "" {
		cities = strings.Split(allowedCities, ",")
	}

	metadata := &SecureLinkMetadata{
		LinkID:      linkID,
		Subject:     subject,
		SenderEmail: senderEmail,
		SenderName:  senderName,
		Status:      status,
		SecuritySettings: SecuritySettingsMetadata{
			RequirePassword:        requirePassword,
			RequireMFA:             requireMFA,
			MFAType:                mfaType,
			GeolocationRestriction: geolocationRestriction,
			AllowedCountries:       countries,
			AllowedCities:          cities,
			TimeLock:               timeLock,
			TimeLockUntil:          timeLockUntil,
			ReadOnce:               readOnce,
			AutoDestruct:           autoDestruct,
			ExpiresAt:              expiresAt,
			MaxAccessAttempts:      maxAccessAttempts,
			CurrentAttempts:        currentAttempts,
		},
	}

	return metadata, nil
}

// validateSecureLinkSecurity validates security requirements for a secure link
func (srv *Server) validateSecureLinkSecurity(ctx context.Context, req SecurityValidationRequest) (*SecurityValidationResponse, error) {
	// Get secure link details
	var (
		requirePassword, requireMFA, geolocationRestriction bool
		mfaType, passwordHash                               string
		allowedCountries, allowedCities                     string
		maxAccessAttempts, currentAttempts                  int
	)

	query := `
		SELECT require_password, require_mfa, mfa_type, password_hash,
		       geolocation_restriction, allowed_countries, allowed_cities,
		       max_access_attempts, current_attempts
		FROM secure_links
		WHERE link_id = ?
	`

	err := srv.db.QueryRowContext(ctx, query, req.LinkID).Scan(
		&requirePassword, &requireMFA, &mfaType, &passwordHash,
		&geolocationRestriction, &allowedCountries, &allowedCities,
		&maxAccessAttempts, &currentAttempts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Check if link has exceeded access attempts
	if currentAttempts >= maxAccessAttempts {
		return &SecurityValidationResponse{
			Success:   false,
			Error:     "This secure link has been destroyed due to too many failed attempts",
			ErrorCode: "LINK_DESTROYED",
		}, nil
	}

	response := &SecurityValidationResponse{
		Success:   true,
		Validated: true,
	}

	// Validate password if required
	if requirePassword {
		if req.Password == "" {
			response.Validated = false
			response.RequiresMFA = false
			return response, nil
		}

		// Validate password
		passwordService := security.NewPasswordProtectionService(srv.db)
		validationReq := security.PasswordValidationRequest{
			LinkID:    req.LinkID,
			Password:  req.Password,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		validationResp, err := passwordService.ValidatePassword(ctx, validationReq)
		if err != nil {
			return nil, fmt.Errorf("password validation failed: %w", err)
		}

		if !validationResp.Valid {
			// Increment failed attempts
			srv.incrementSecureLinkAttempts(ctx, req.LinkID)

			// Check if link should be destroyed
			if currentAttempts+1 >= maxAccessAttempts {
				return &SecurityValidationResponse{
					Success:   false,
					Error:     "This secure link has been destroyed due to too many failed attempts",
					ErrorCode: "LINK_DESTROYED",
				}, nil
			}

			return &SecurityValidationResponse{
				Success:      false,
				Error:        "Invalid password",
				ErrorCode:    "INVALID_PASSWORD",
				DecoyMessage: "The password you entered is incorrect. Please try again.",
			}, nil
		}
	}

	// Validate MFA if required
	if requireMFA {
		if req.MFACode == "" {
			response.Validated = false
			response.RequiresMFA = true
			response.MFAType = mfaType
			return response, nil
		}

		// Validate MFA code
		valid, err := srv.validateMFACode(ctx, req.LinkID, req.MFACode, mfaType)
		if err != nil {
			return nil, fmt.Errorf("MFA validation failed: %w", err)
		}

		if !valid {
			// Increment failed attempts
			srv.incrementSecureLinkAttempts(ctx, req.LinkID)

			return &SecurityValidationResponse{
				Success:      false,
				Error:        "Invalid MFA code",
				ErrorCode:    "INVALID_MFA",
				DecoyMessage: "The verification code you entered is incorrect. Please try again.",
			}, nil
		}
	}

	// Validate geolocation if required
	if geolocationRestriction {
		valid, err := srv.validateGeolocation(ctx, req.LinkID, req.IPAddress, allowedCountries, allowedCities)
		if err != nil {
			return nil, fmt.Errorf("geolocation validation failed: %w", err)
		}

		if !valid {
			return &SecurityValidationResponse{
				Success:      false,
				Error:        "Access denied from your location",
				ErrorCode:    "GEO_RESTRICTED",
				DecoyMessage: "Access to this secure message is restricted to specific locations.",
			}, nil
		}
	}

	return response, nil
}

// getSecureEmailContent retrieves the secure email content after validation
func (srv *Server) getSecureEmailContent(ctx context.Context, linkID, clientIP, userAgent string) (*SecureEmailContent, error) {
	// Get email content from secure link
	var (
		emailID, subject, body, senderEmail, senderName string
		readOnce, autoDestruct                          bool
	)

	query := `
		SELECT sl.email_id, e.subject, e.body, u.email as sender_email, u.name as sender_name,
		       sl.read_once, sl.auto_destruct
		FROM secure_links sl
		JOIN emails e ON sl.email_id = e.email_id
		JOIN users u ON e.sender_id = u.id
		WHERE sl.link_id = ?
	`

	err := srv.db.QueryRowContext(ctx, query, linkID).Scan(
		&emailID, &subject, &body, &senderEmail, &senderName,
		&readOnce, &autoDestruct,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get email content: %w", err)
	}

	// If read-once, mark as read and potentially destroy
	if readOnce {
		if err := srv.markSecureLinkAsRead(ctx, linkID); err != nil {
			log.Printf("⚠️ Failed to mark secure link as read: %v", err)
		}
	}

	content := &SecureEmailContent{
		LinkID:       linkID,
		Subject:      subject,
		Body:         body,
		SenderEmail:  senderEmail,
		SenderName:   senderName,
		ReadOnce:     readOnce,
		AutoDestruct: autoDestruct,
	}

	return content, nil
}

// validateMFACode validates an MFA code for a secure link
func (srv *Server) validateMFACode(ctx context.Context, linkID, code, mfaType string) (bool, error) {
	// This would integrate with the existing MFA service
	// For now, we'll use a simple validation
	if mfaType == "TOTP" {
		// Validate TOTP code
		return srv.validateTOTPCode(ctx, linkID, code)
	} else if mfaType == "EMAIL_CODE" {
		// Validate email code
		return srv.validateEmailCode(ctx, linkID, code)
	}

	return false, fmt.Errorf("unsupported MFA type: %s", mfaType)
}

// validateTOTPCode validates a TOTP code
func (srv *Server) validateTOTPCode(ctx context.Context, linkID, code string) (bool, error) {
	// This would integrate with the existing TOTP service
	// For now, we'll use a simple validation
	if len(code) == 6 && code == "000000" {
		return true, nil
	}
	return false, nil
}

// validateEmailCode validates an email code
func (srv *Server) validateEmailCode(ctx context.Context, linkID, code string) (bool, error) {
	// This would integrate with the existing email MFA service
	// For now, we'll use a simple validation
	if len(code) == 6 && code == "123456" {
		return true, nil
	}
	return false, nil
}

// validateGeolocation validates geolocation restrictions
func (srv *Server) validateGeolocation(ctx context.Context, linkID, clientIP, allowedCountries, allowedCities string) (bool, error) {
	if allowedCountries == "" && allowedCities == "" {
		return true, nil
	}

	// Get geolocation for client IP
	geoService := geolocation.NewMockGeolocationService()
	geoInfo, err := geoService.GetLocation(clientIP)
	if err != nil {
		log.Printf("⚠️ Failed to get geolocation for IP %s: %v", clientIP, err)
		return false, nil // Deny access if we can't determine location
	}

	// Check country restrictions
	if allowedCountries != "" {
		countries := strings.Split(allowedCountries, ",")
		countryAllowed := false
		for _, country := range countries {
			if strings.TrimSpace(country) == geoInfo.Country {
				countryAllowed = true
				break
			}
		}
		if !countryAllowed {
			return false, nil
		}
	}

	// Check city restrictions
	if allowedCities != "" {
		cities := strings.Split(allowedCities, ",")
		cityAllowed := false
		for _, city := range cities {
			if strings.TrimSpace(strings.ToLower(city)) == strings.ToLower(geoInfo.City) {
				cityAllowed = true
				break
			}
		}
		if !cityAllowed {
			return false, nil
		}
	}

	return true, nil
}

// incrementSecureLinkAttempts increments the failed attempts counter
func (srv *Server) incrementSecureLinkAttempts(ctx context.Context, linkID string) error {
	query := `
		UPDATE secure_links 
		SET current_attempts = current_attempts + 1,
		    last_attempt = CURRENT_TIMESTAMP
		WHERE link_id = ?
	`

	_, err := srv.db.ExecContext(ctx, query, linkID)
	return err
}

// markSecureLinkAsRead marks a secure link as read (for read-once links)
func (srv *Server) markSecureLinkAsRead(ctx context.Context, linkID string) error {
	query := `
		UPDATE secure_links 
		SET read_count = read_count + 1,
		    last_read = CURRENT_TIMESTAMP,
		    status = CASE 
		        WHEN read_once = 1 THEN 'destroyed'
		        ELSE status 
		    END
		WHERE link_id = ?
	`

	_, err := srv.db.ExecContext(ctx, query, linkID)
	return err
}

// logSecureLinkAccess logs access attempts to the audit log
func (srv *Server) logSecureLinkAccess(ctx context.Context, linkID, clientIP, userAgent, eventType, details string, sesTransactionID *string) {
	query := `
		INSERT INTO link_audit_log (
			id, link_id, event_type, timestamp, ip_address, user_agent,
			details, ses_transaction_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	auditID := fmt.Sprintf("access_%s_%d", linkID, time.Now().Unix())

	_, err := srv.db.ExecContext(ctx, query,
		auditID,
		linkID,
		eventType,
		time.Now(),
		clientIP,
		userAgent,
		details,
		sesTransactionID,
	)

	if err != nil {
		log.Printf("⚠️ Failed to log secure link access: %v", err)
	}
}
