package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/securelinks/decoy"
	"secure-email-mvp/pkg/securelinks/geolocation"
	"secure-email-mvp/pkg/securelinks/mfa"
)

// =============================================================================
// PHASE 2 SECURITY API HANDLERS
// =============================================================================

// GeolocationVerificationHandler handles geolocation verification requests
func (srv *Server) GeolocationVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		LinkID           string   `json:"link_id" validate:"required"`
		IPAddress        string   `json:"ip_address" validate:"required"`
		AllowedCountries []string `json:"allowed_countries,omitempty"`
		AllowedCities    []string `json:"allowed_cities,omitempty"`
		BlockedCountries []string `json:"blocked_countries,omitempty"`
		BlockedCities    []string `json:"blocked_cities,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create geolocation service
	geoService := geolocation.NewGeolocationVerificationService()

	// Create geolocation restriction
	restriction := geolocation.GeolocationRestriction{
		Enabled:          true,
		AllowedCountries: req.AllowedCountries,
		AllowedCities:    req.AllowedCities,
		BlockedCountries: req.BlockedCountries,
		BlockedCities:    req.BlockedCities,
	}

	// Verify location
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := geoService.VerifyLocation(ctx, req.IPAddress, restriction)
	if err != nil {
		log.Printf("Geolocation verification failed: %v", err)
		http.Error(w, "Geolocation verification failed", http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MFAInitiationHandler handles MFA initiation requests
func (srv *Server) MFAInitiationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req mfa.MFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create MFA service
	mfaService := mfa.NewExternalMFAService(srv.db)

	// Initiate MFA
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := mfaService.InitiateMFA(ctx, req)
	if err != nil {
		log.Printf("MFA initiation failed: %v", err)
		http.Error(w, "MFA initiation failed", http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MFAVerificationHandler handles MFA verification requests
func (srv *Server) MFAVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req mfa.MFAVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create MFA service
	mfaService := mfa.NewExternalMFAService(srv.db)

	// Verify MFA
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := mfaService.VerifyMFA(ctx, req)
	if err != nil {
		log.Printf("MFA verification failed: %v", err)
		http.Error(w, "MFA verification failed", http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DecoyMessageHandler handles decoy message requests
func (srv *Server) DecoyMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req decoy.DecoyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create decoy service
	decoyService := decoy.NewDecoyMessageService(srv.db)

	// Get decoy message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := decoyService.GetDecoyMessage(ctx, req)
	if err != nil {
		log.Printf("Decoy message retrieval failed: %v", err)
		http.Error(w, "Decoy message retrieval failed", http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateDecoyMessageHandler handles decoy message creation requests
func (srv *Server) CreateDecoyMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var decoyMessage decoy.DecoyMessage
	if err := json.NewDecoder(r.Body).Decode(&decoyMessage); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create decoy service
	decoyService := decoy.NewDecoyMessageService(srv.db)

	// Create decoy message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := decoyService.CreateDecoyMessage(ctx, &decoyMessage)
	if err != nil {
		log.Printf("Decoy message creation failed: %v", err)
		http.Error(w, "Decoy message creation failed", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Decoy message created successfully",
		"id":      decoyMessage.ID,
	})
}

// GetDecoyTemplatesHandler returns available decoy message templates
func (srv *Server) GetDecoyTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create decoy service
	decoyService := decoy.NewDecoyMessageService(srv.db)

	// Get templates
	templates := decoyService.GetDecoyMessageTemplates()

	// Return templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"templates": templates,
	})
}

// GeolocationDataHandler returns geolocation data for an IP address
func (srv *Server) GeolocationDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get IP address from query parameter
	ipAddress := r.URL.Query().Get("ip")
	if ipAddress == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	// Create geolocation service
	geoService := geolocation.NewGeolocationVerificationService()

	// Get geolocation data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	location, err := geoService.GetGeolocationDataForIP(ctx, ipAddress)
	if err != nil {
		log.Printf("Geolocation data retrieval failed: %v", err)
		http.Error(w, "Geolocation data retrieval failed", http.StatusInternalServerError)
		return
	}

	// Return location data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"location": location,
	})
}

// ValidateGeolocationRestrictionHandler validates geolocation restriction configuration
func (srv *Server) ValidateGeolocationRestrictionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var restriction geolocation.GeolocationRestriction
	if err := json.NewDecoder(r.Body).Decode(&restriction); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create geolocation service
	geoService := geolocation.NewGeolocationVerificationService()

	// Validate restriction
	err := geoService.ValidateGeolocationRestriction(restriction)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Geolocation restriction configuration is valid",
	})
}
