package auth

// DO NOT EDIT EXISTING CODE - new file added
// Domain management handlers: add, verify, and remove custom domains

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type addDomainRequest struct {
	DomainName string `json:"domain_name"`
}

type addDomainResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	VerificationCode string `json:"verification_code,omitempty"`
	DNSRecord string `json:"dns_record,omitempty"`
}

type verifyDomainRequest struct {
	DomainName string `json:"domain_name"`
}

type verifyDomainResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type removeDomainRequest struct {
	DomainName string `json:"domain_name"`
}

type removeDomainResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// POST /api/domain/add
func AddDomainHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req addDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate domain name
	if req.DomainName == "" {
		http.Error(w, "domain name required", http.StatusBadRequest)
		return
	}

	// Normalize domain name
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	if !isValidDomain(domainName) {
		http.Error(w, "invalid domain format", http.StatusBadRequest)
		return
	}

	// Check if user has permission to add domains (premium or enterprise)
	var accountType string
	err := DB.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
	if err != nil {
		log.Printf("ERROR getting user account type: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if accountType == "free" {
		http.Error(w, "domain management requires premium or enterprise account", http.StatusForbidden)
		return
	}

	// Check if domain already exists
	var existingDomain string
	err = DB.QueryRow("SELECT domain_name FROM domains WHERE domain_name = ?", domainName).Scan(&existingDomain)
	if err == nil {
		http.Error(w, "domain already exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("ERROR checking domain existence: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Generate verification code
	verificationCode, err := GenerateRandomToken(16)
	if err != nil {
		log.Printf("ERROR generating verification code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Encrypt verification code
	encryptedCode, err := encryptSubscriptionData(verificationCode)
	if err != nil {
		log.Printf("ERROR encrypting verification code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Create DNS record value
	dnsRecordValue := fmt.Sprintf("securemail-verification=%s", verificationCode)
	encryptedDNSRecord, err := encryptSubscriptionData(dnsRecordValue)
	if err != nil {
		log.Printf("ERROR encrypting DNS record: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Insert domain into database
	domainID := uuid.New().String()
	now := time.Now().UTC()

	_, err = DB.Exec(`
		INSERT INTO domains (id, user_id, domain_name, verified, verification_code, verification_method, dns_record_value, created_at, updated_at)
		VALUES (?, ?, ?, FALSE, ?, 'TXT', ?, ?, ?)
	`, domainID, userID, domainName, encryptedCode, encryptedDNSRecord, now, now)

	if err != nil {
		log.Printf("ERROR inserting domain: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log domain addition (non-sensitive)
	log.Printf("INFO domain_added user_id=%s domain=%s", userID, domainName)

	resp := addDomainResponse{
		Success: true,
		Message: "Domain added successfully. Please add the DNS record to verify ownership.",
		VerificationCode: verificationCode,
		DNSRecord: dnsRecordValue,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/domain/verify
func VerifyDomainHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req verifyDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Normalize domain name
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))

	// Get domain from database
	var domainID string
	var encryptedCode string
	var encryptedDNSRecord string
	var verificationAttempts int
	var lastAttempt sql.NullTime

	err := DB.QueryRow(`
		SELECT id, verification_code, dns_record_value, verification_attempts, last_verification_attempt
		FROM domains 
		WHERE domain_name = ? AND user_id = ?
	`, domainName, userID).Scan(&domainID, &encryptedCode, &encryptedDNSRecord, &verificationAttempts, &lastAttempt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "domain not found", http.StatusNotFound)
			return
		}
		log.Printf("ERROR getting domain: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Check verification attempts (rate limiting)
	if verificationAttempts >= 5 {
		http.Error(w, "too many verification attempts", http.StatusTooManyRequests)
		return
	}

	// Check if domain is already verified
	var verified bool
	err = DB.QueryRow("SELECT verified FROM domains WHERE id = ?", domainID).Scan(&verified)
	if err != nil {
		log.Printf("ERROR checking domain verification status: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if verified {
		http.Error(w, "domain already verified", http.StatusConflict)
		return
	}

	// Decrypt DNS record value
	_, err = decryptSubscriptionData(encryptedDNSRecord)
	if err != nil {
		log.Printf("ERROR decrypting DNS record: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// TODO: Implement actual DNS verification
	// For now, we'll simulate verification success
	// In production, this would query DNS records to verify the TXT record exists
	verificationSuccess := true // simulateDNSVerification(domainName, expectedDNSRecord)

	// Update verification attempt
	now := time.Now().UTC()
	_, err = DB.Exec(`
		UPDATE domains 
		SET verification_attempts = verification_attempts + 1, 
		    last_verification_attempt = ?, 
		    updated_at = ?
		WHERE id = ?
	`, now, now, domainID)

	if err != nil {
		log.Printf("ERROR updating verification attempt: %v", err)
	}

	if verificationSuccess {
		// Mark domain as verified
		_, err = DB.Exec(`
			UPDATE domains 
			SET verified = TRUE, 
			    verification_code = NULL, 
			    dns_record_value = NULL,
			    updated_at = ?
			WHERE id = ?
		`, now, domainID)

		if err != nil {
			log.Printf("ERROR marking domain as verified: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// Log successful verification (non-sensitive)
		log.Printf("INFO domain_verified user_id=%s domain=%s", userID, domainName)

		resp := verifyDomainResponse{
			Success: true,
			Message: "Domain verified successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else {
		http.Error(w, "domain verification failed", http.StatusBadRequest)
	}
}

// POST /api/domain/remove
func RemoveDomainHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req removeDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Normalize domain name
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))

	// Check if domain exists and belongs to user
	var domainID string
	err := DB.QueryRow("SELECT id FROM domains WHERE domain_name = ? AND user_id = ?", domainName, userID).Scan(&domainID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "domain not found", http.StatusNotFound)
			return
		}
		log.Printf("ERROR getting domain: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Remove domain
	_, err = DB.Exec("DELETE FROM domains WHERE id = ?", domainID)
	if err != nil {
		log.Printf("ERROR removing domain: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log domain removal (non-sensitive)
	log.Printf("INFO domain_removed user_id=%s domain=%s", userID, domainName)

	resp := removeDomainResponse{
		Success: true,
		Message: "Domain removed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper function to validate domain name
func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	
	// Basic domain validation
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Check for valid characters
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
				return false
			}
		}
		// Cannot start or end with hyphen
		if part[0] == '-' || part[len(part)-1] == '-' {
			return false
		}
	}
	
	return true
}
