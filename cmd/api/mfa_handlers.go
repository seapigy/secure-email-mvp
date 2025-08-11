package main

import (
	"encoding/json"
	"log"
	"net/http"

	"secure-email-mvp/pkg/mfa"

	"github.com/gorilla/mux"
)

// validateMFAHandler handles POST /api/mfa/validate
// Validates MFA codes (TOTP or email-based) for email access
func (srv *Server) validateMFAHandler(w http.ResponseWriter, r *http.Request) {
	var req mfa.MFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmailID == "" || req.MFACode == "" {
		http.Error(w, `{"error":"Email ID and MFA code are required"}`, http.StatusBadRequest)
		return
	}

	// Get MFA service
	mfaService := mfa.NewMFAService(srv.db)

	// Check if MFA is locked due to too many failed attempts
	locked, _, err := mfaService.CheckMFALockout(req.EmailID)
	if err != nil {
		log.Printf("Failed to check MFA lockout: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if locked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(mfa.MFAResponse{
			Success: false,
			Message: "MFA is temporarily locked due to too many failed attempts",
			Code:    "mfa_locked",
		})
		return
	}

	// Get MFA configuration for the email
	config, err := mfaService.GetMFAConfig(req.EmailID)
	if err != nil {
		log.Printf("Failed to get MFA config: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if !config.RequireMFA {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(mfa.MFAResponse{
			Success: false,
			Message: "MFA is not required for this email",
			Code:    "mfa_not_required",
		})
		return
	}

	// Validate the MFA code based on type
	var valid bool
	switch config.MFAType {
	case mfa.MFATypeTOTP:
		valid, err = mfaService.ValidateTOTP(req.EmailID, req.MFACode)
	case mfa.MFATypeEmailCode:
		valid, err = mfaService.ValidateEmailCode(req.EmailID, req.MFACode)
	default:
		http.Error(w, `{"error":"Invalid MFA type"}`, http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("MFA validation error: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if valid {
		// Reset failed attempts on successful validation
		if err := mfaService.ResetFailedAttempts(req.EmailID); err != nil {
			log.Printf("Failed to reset MFA attempts: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mfa.MFAResponse{
			Success: true,
			Message: "MFA validation successful",
		})
	} else {
		// Increment failed attempts
		if err := mfaService.IncrementFailedAttempts(req.EmailID); err != nil {
			log.Printf("Failed to increment MFA attempts: %v", err)
		}

		// Check if we should lock the account
		locked, _, err := mfaService.CheckMFALockout(req.EmailID)
		if err != nil {
			log.Printf("Failed to check MFA lockout after failed attempt: %v", err)
		}

		if locked {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(mfa.MFAResponse{
				Success: false,
				Message: "Too many failed attempts. MFA is now locked.",
				Code:    "mfa_locked",
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(mfa.MFAResponse{
				Success: false,
				Message: "Invalid MFA code",
				Code:    "invalid_mfa_code",
			})
		}
	}
}

// generateEmailCodeHandler handles POST /api/mfa/email-code
// Generates and sends an email-based MFA code
func (srv *Server) generateEmailCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req mfa.EmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmailID == "" {
		http.Error(w, `{"error":"Email ID is required"}`, http.StatusBadRequest)
		return
	}

	// Get MFA service
	mfaService := mfa.NewMFAService(srv.db)

	// Get MFA configuration for the email
	config, err := mfaService.GetMFAConfig(req.EmailID)
	if err != nil {
		log.Printf("Failed to get MFA config: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if !config.RequireMFA {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(mfa.EmailCodeResponse{
			Success: false,
			Message: "MFA is not required for this email",
			Code:    "mfa_not_required",
		})
		return
	}

	if config.MFAType != mfa.MFATypeEmailCode {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(mfa.EmailCodeResponse{
			Success: false,
			Message: "Email-based MFA is not enabled for this email",
			Code:    "wrong_mfa_type",
		})
		return
	}

	// Check if MFA is locked
	locked, _, err := mfaService.CheckMFALockout(req.EmailID)
	if err != nil {
		log.Printf("Failed to check MFA lockout: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if locked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(mfa.EmailCodeResponse{
			Success: false,
			Message: "MFA is temporarily locked due to too many failed attempts",
			Code:    "mfa_locked",
		})
		return
	}

	// Generate a new email code
	code, err := mfaService.GenerateEmailCode()
	if err != nil {
		log.Printf("Failed to generate email code: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Store the code with expiration
	if err := mfaService.StoreEmailCode(req.EmailID, code); err != nil {
		log.Printf("Failed to store email code: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// In a real implementation, you would send this code via email
	// For now, we'll return it in the response for testing purposes
	// In production, this should be sent via email and not returned in the response
	log.Printf("Generated email code for email %s: %s", req.EmailID, code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mfa.EmailCodeResponse{
		Success: true,
		Message: "Email code generated successfully",
		Code:    code, // Remove this in production
	})
}

// getMFAConfigHandler handles GET /api/mfa/config/{emailID}
// Returns the MFA configuration for an email
func (srv *Server) getMFAConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["emailID"]
	if emailID == "" {
		http.Error(w, `{"error":"Email ID is required"}`, http.StatusBadRequest)
		return
	}

	// Get MFA service
	mfaService := mfa.NewMFAService(srv.db)

	// Get MFA configuration
	config, err := mfaService.GetMFAConfig(emailID)
	if err != nil {
		log.Printf("Failed to get MFA config: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
