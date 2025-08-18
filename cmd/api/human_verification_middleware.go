package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"secure-email-mvp/pkg/humanverification"
)

// HumanVerificationMiddleware wraps the human verification logic
type HumanVerificationMiddleware struct {
	verificationService humanverification.HumanVerificationService
	config              *humanverification.Config
}

// NewHumanVerificationMiddleware creates a new human verification middleware
func NewHumanVerificationMiddleware(verificationService humanverification.HumanVerificationService, config *humanverification.Config) *HumanVerificationMiddleware {
	return &HumanVerificationMiddleware{
		verificationService: verificationService,
		config:              config,
	}
}

// RequireHumanVerification middleware that requires human verification for protected endpoints
func (hvm *HumanVerificationMiddleware) RequireHumanVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hvm.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract email ID from URL
		vars := mux.Vars(r)
		emailID := vars["id"]
		if emailID == "" {
			http.Error(w, "Email ID not found", http.StatusBadRequest)
			return
		}

		// Get client information
		clientIP := getClientIP(r)
		userAgent := r.Header.Get("User-Agent")

		// Check for verification token in headers or query parameters
		verificationToken := r.Header.Get("X-Human-Verification-Token")
		if verificationToken == "" {
			verificationToken = r.URL.Query().Get("human_verification_token")
		}

		// Determine verification type
		verificationType := humanverification.VerificationType(hvm.config.VerificationType)

		// If no token provided, return challenge
		if verificationToken == "" {
			hvm.handleChallengeRequest(w, r, emailID, clientIP, userAgent, verificationType)
			return
		}

		// Verify the token
		ctx := r.Context()
		success, err := hvm.verificationService.VerifyResponse(ctx, emailID, verificationToken, verificationType)
		
		// Log the verification attempt
		result := "success"
		if !success {
			result = "failure"
		}
		
		logEntry := &humanverification.VerificationLog{
			EmailID:         emailID,
			IPAddress:       clientIP,
			UserAgent:       userAgent,
			VerificationType: verificationType,
			Result:          result,
		}
		
		if err := hvm.verificationService.LogVerification(ctx, logEntry); err != nil {
			log.Printf("Failed to log verification attempt: %v", err)
		}

		if err != nil {
			log.Printf("Human verification error: %v", err)
			hvm.handleVerificationFailure(w, r, emailID, clientIP)
			return
		}

		if !success {
			log.Printf("Human verification failed for email %s from IP %s", emailID, clientIP)
			hvm.handleVerificationFailure(w, r, emailID, clientIP)
			return
		}

		// Verification successful, proceed to next handler
		next.ServeHTTP(w, r)
	})
}

// handleChallengeRequest handles requests for verification challenges
func (hvm *HumanVerificationMiddleware) handleChallengeRequest(w http.ResponseWriter, r *http.Request, emailID, clientIP, userAgent string, verificationType humanverification.VerificationType) {
	ctx := r.Context()
	
	// Log the challenge request
	logEntry := &humanverification.VerificationLog{
		EmailID:         emailID,
		IPAddress:       clientIP,
		UserAgent:       userAgent,
		VerificationType: verificationType,
		Result:          "challenge_requested",
	}
	
	if err := hvm.verificationService.LogVerification(ctx, logEntry); err != nil {
		log.Printf("Failed to log challenge request: %v", err)
	}

	// Generate challenge based on verification type
	switch verificationType {
	case humanverification.VerificationTypeProofOfWork:
		challenge, err := hvm.verificationService.GenerateChallenge(ctx)
		if err != nil {
			log.Printf("Failed to generate proof-of-work challenge: %v", err)
			http.Error(w, "Failed to generate verification challenge", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"verification_required": true,
			"verification_type":     "proof_of_work",
			"challenge":             challenge,
			"message":               "Human verification required. Solve the proof-of-work challenge and include the solution in X-Human-Verification-Token header.",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)

	case humanverification.VerificationTypeCAPTCHA:
		response := map[string]interface{}{
			"verification_required": true,
			"verification_type":     "captcha",
			"captcha_site_key":      hvm.config.CAPTCHASiteKey,
			"message":               "Human verification required. Complete the CAPTCHA and include the token in X-Human-Verification-Token header.",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Unsupported verification type", http.StatusBadRequest)
	}
}

// handleVerificationFailure handles verification failures
func (hvm *HumanVerificationMiddleware) handleVerificationFailure(w http.ResponseWriter, r *http.Request, emailID, clientIP string) {
			// Check if we should increment self-destruct counter
		ctx := r.Context()
		stats, err := hvm.verificationService.GetVerificationStats(ctx, emailID, clientIP, time.Hour)
		if err != nil {
			log.Printf("Failed to get verification stats: %v", err)
		} else {
			// If failure rate is high or total attempts exceed threshold, log for monitoring
			if stats.FailureRate > 0.8 || stats.TotalAttempts > hvm.config.FailureThreshold {
				log.Printf("High verification failure rate detected for email %s from IP %s: %f%% (%d attempts)", 
					emailID, clientIP, stats.FailureRate*100, stats.TotalAttempts)
			}
		}

	// Return generic error message to prevent information leakage
	response := map[string]interface{}{
		"error":   "Email has been revoked or cannot be accessed",
		"message": "Access denied",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(response)
}

// GenerateChallengeHandler handles requests for new verification challenges
func (hvm *HumanVerificationMiddleware) GenerateChallengeHandler(w http.ResponseWriter, r *http.Request) {
	if !hvm.config.Enabled {
		http.Error(w, "Human verification is disabled", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	verificationType := humanverification.VerificationType(hvm.config.VerificationType)

	switch verificationType {
	case humanverification.VerificationTypeProofOfWork:
		challenge, err := hvm.verificationService.GenerateChallenge(ctx)
		if err != nil {
			log.Printf("Failed to generate proof-of-work challenge: %v", err)
			http.Error(w, "Failed to generate challenge", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"challenge": challenge,
			"type":      "proof_of_work",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case humanverification.VerificationTypeCAPTCHA:
		response := map[string]interface{}{
			"captcha_site_key": hvm.config.CAPTCHASiteKey,
			"type":             "captcha",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Unsupported verification type", http.StatusBadRequest)
	}
}


