package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/geolocation"
	"secure-email-mvp/pkg/georestriction"
	"secure-email-mvp/pkg/reputation"

	"github.com/gorilla/mux"
)

// GeoRestrictionRuleRequest represents a request to create or update a geo-restriction rule
type GeoRestrictionRuleRequest struct {
	EmailID     string                           `json:"email_id"`
	Type        georestriction.RestrictionType   `json:"type"`
	Countries   []string                         `json:"countries,omitempty"`
	Cities      []string                         `json:"cities,omitempty"`
	Description string                           `json:"description,omitempty"`
}

// GeoRestrictionConfigRequest represents a request to update geo-restriction configuration
type GeoRestrictionConfigRequest struct {
	EmailID                    string                           `json:"email_id"`
	Enabled                    bool                             `json:"enabled"`
	DefaultAction              georestriction.RestrictionType   `json:"default_action"`
	StrictMode                 bool                             `json:"strict_mode"`
	LogViolations              bool                             `json:"log_violations"`
	BlockOnGeolocationFailure  bool                             `json:"block_on_geolocation_failure"`
}

// GeoRestrictionResponse represents the response for geo-restriction operations
type GeoRestrictionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Rules   []georestriction.GeoRestrictionRule `json:"rules,omitempty"`
	Config  georestriction.GeoRestrictionConfig  `json:"config,omitempty"`
}

// GeoRestrictionStatusResponse represents the status of geo-restrictions for an email
type GeoRestrictionStatusResponse struct {
	EmailID                    string                           `json:"email_id"`
	Enabled                    bool                             `json:"enabled"`
	RulesCount                 int                              `json:"rules_count"`
	ViolationsCount            int                              `json:"violations_count"`
	LastViolation              *time.Time                       `json:"last_violation,omitempty"`
	Config                     georestriction.GeoRestrictionConfig `json:"config"`
	CurrentLocation            *geolocation.Location            `json:"current_location,omitempty"`
	AccessAllowed              bool                             `json:"access_allowed"`
	AccessReason               string                           `json:"access_reason,omitempty"`
}

// geoRestrictionHandlers provides HTTP handlers for geo-restriction management
type geoRestrictionHandlers struct {
	db *sql.DB
}

// NewGeoRestrictionHandlers creates a new geo-restriction handlers instance
func NewGeoRestrictionHandlers(db *sql.DB) *geoRestrictionHandlers {
	return &geoRestrictionHandlers{db: db}
}

// getGeoRestrictionRulesHandler handles GET /api/email/{id}/geo-restrictions
// Returns the current geo-restriction rules for an email
func (h *geoRestrictionHandlers) getGeoRestrictionRulesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, "Missing email_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get geo-restriction rules from database
	rules, err := h.getGeoRestrictionRules(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction rules for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction rules", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction rules retrieved successfully",
		Rules:   rules,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// createGeoRestrictionRuleHandler handles POST /api/email/{id}/geo-restrictions
// Creates a new geo-restriction rule for an email
func (h *geoRestrictionHandlers) createGeoRestrictionRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, "Missing email_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req GeoRestrictionRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the rule
	rule := georestriction.GeoRestrictionRule{
		ID:          generateUUID(),
		EmailID:     emailID,
		Type:        req.Type,
		Countries:   req.Countries,
		Cities:      req.Cities,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate the rule
	service := georestriction.NewGeoRestrictionService()
	if err := service.ValidateRule(rule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid rule: %v", err), http.StatusBadRequest)
		return
	}

	// Normalize the rule
	rule = service.NormalizeRule(rule)

	// Save the rule to database
	if err := h.saveGeoRestrictionRule(rule); err != nil {
		log.Printf("Failed to save geo-restriction rule for email %s: %v", emailID, err)
		http.Error(w, "Failed to create geo-restriction rule", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction rule created successfully",
		Rules:   []georestriction.GeoRestrictionRule{rule},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateGeoRestrictionRuleHandler handles PUT /api/email/{id}/geo-restrictions/{ruleId}
// Updates an existing geo-restriction rule
func (h *geoRestrictionHandlers) updateGeoRestrictionRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	ruleID := vars["ruleId"]
	if emailID == "" || ruleID == "" {
		http.Error(w, "Missing email_id or rule_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req GeoRestrictionRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing rule
	existingRules, err := h.getGeoRestrictionRules(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction rules for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction rules", http.StatusInternalServerError)
		return
	}

	// Find the rule to update
	var ruleToUpdate *georestriction.GeoRestrictionRule
	for i, rule := range existingRules {
		if rule.ID == ruleID {
			ruleToUpdate = &existingRules[i]
			break
		}
	}

	if ruleToUpdate == nil {
		http.Error(w, "Geo-restriction rule not found", http.StatusNotFound)
		return
	}

	// Update the rule
	ruleToUpdate.Type = req.Type
	ruleToUpdate.Countries = req.Countries
	ruleToUpdate.Cities = req.Cities
	ruleToUpdate.Description = req.Description
	ruleToUpdate.UpdatedAt = time.Now()

	// Validate the updated rule
	service := georestriction.NewGeoRestrictionService()
	if err := service.ValidateRule(*ruleToUpdate); err != nil {
		http.Error(w, fmt.Sprintf("Invalid rule: %v", err), http.StatusBadRequest)
		return
	}

	// Normalize the rule
	*ruleToUpdate = service.NormalizeRule(*ruleToUpdate)

	// Save the updated rules to database
	if err := h.saveGeoRestrictionRules(emailID, existingRules); err != nil {
		log.Printf("Failed to update geo-restriction rule for email %s: %v", emailID, err)
		http.Error(w, "Failed to update geo-restriction rule", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction rule updated successfully",
		Rules:   []georestriction.GeoRestrictionRule{*ruleToUpdate},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deleteGeoRestrictionRuleHandler handles DELETE /api/email/{id}/geo-restrictions/{ruleId}
// Deletes a geo-restriction rule
func (h *geoRestrictionHandlers) deleteGeoRestrictionRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	ruleID := vars["ruleId"]
	if emailID == "" || ruleID == "" {
		http.Error(w, "Missing email_id or rule_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get existing rules
	existingRules, err := h.getGeoRestrictionRules(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction rules for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction rules", http.StatusInternalServerError)
		return
	}

	// Filter out the rule to delete
	var updatedRules []georestriction.GeoRestrictionRule
	ruleFound := false
	for _, rule := range existingRules {
		if rule.ID != ruleID {
			updatedRules = append(updatedRules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		http.Error(w, "Geo-restriction rule not found", http.StatusNotFound)
		return
	}

	// Save the updated rules to database
	if err := h.saveGeoRestrictionRules(emailID, updatedRules); err != nil {
		log.Printf("Failed to delete geo-restriction rule for email %s: %v", emailID, err)
		http.Error(w, "Failed to delete geo-restriction rule", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction rule deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getGeoRestrictionConfigHandler handles GET /api/email/{id}/geo-restrictions/config
// Returns the geo-restriction configuration for an email
func (h *geoRestrictionHandlers) getGeoRestrictionConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, "Missing email_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get geo-restriction configuration from database
	config, err := h.getGeoRestrictionConfig(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction config for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction configuration", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction configuration retrieved successfully",
		Config:  config,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateGeoRestrictionConfigHandler handles PUT /api/email/{id}/geo-restrictions/config
// Updates the geo-restriction configuration for an email
func (h *geoRestrictionHandlers) updateGeoRestrictionConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, "Missing email_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req GeoRestrictionConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create configuration
	config := georestriction.GeoRestrictionConfig{
		Enabled:                    req.Enabled,
		DefaultAction:              req.DefaultAction,
		StrictMode:                 req.StrictMode,
		LogViolations:              req.LogViolations,
		BlockOnGeolocationFailure:  req.BlockOnGeolocationFailure,
	}

	// Save configuration to database
	if err := h.saveGeoRestrictionConfig(emailID, config); err != nil {
		log.Printf("Failed to save geo-restriction config for email %s: %v", emailID, err)
		http.Error(w, "Failed to update geo-restriction configuration", http.StatusInternalServerError)
		return
	}

	response := GeoRestrictionResponse{
		Success: true,
		Message: "Geo-restriction configuration updated successfully",
		Config:  config,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getGeoRestrictionStatusHandler handles GET /api/email/{id}/geo-restrictions/status
// Returns the current status of geo-restrictions for an email
func (h *geoRestrictionHandlers) getGeoRestrictionStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, "Missing email_id", http.StatusBadRequest)
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user owns this email
	if err := h.verifyEmailOwnership(emailID, userID); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get geo-restriction data from database
	rules, err := h.getGeoRestrictionRules(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction rules for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction status", http.StatusInternalServerError)
		return
	}

	config, err := h.getGeoRestrictionConfig(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction config for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction status", http.StatusInternalServerError)
		return
	}

	// Get violation data
	violationsCount, lastViolation, err := h.getGeoRestrictionViolations(emailID)
	if err != nil {
		log.Printf("Failed to get geo-restriction violations for email %s: %v", emailID, err)
		http.Error(w, "Failed to retrieve geo-restriction status", http.StatusInternalServerError)
		return
	}

	// Get current client location
	clientIP := reputation.GetClientIP(r)
	geoService := geolocation.NewGeolocationService()
	location, err := geoService.GetLocationByIP(clientIP)
	if err != nil {
		log.Printf("Failed to get geolocation for IP %s: %v", clientIP, err)
		// Don't fail the request, just don't include location
	}

	// Check access
	service := georestriction.NewGeoRestrictionServiceWithConfig(config)
	accessAllowed, accessReason, err := service.CheckAccess(location, rules)
	if err != nil {
		log.Printf("Failed to check geo-restriction access for email %s: %v", emailID, err)
		accessAllowed = false
		accessReason = "Error checking access"
	}

	response := GeoRestrictionStatusResponse{
		EmailID:         emailID,
		Enabled:         config.Enabled,
		RulesCount:      len(rules),
		ViolationsCount: violationsCount,
		LastViolation:   lastViolation,
		Config:          config,
		CurrentLocation: location,
		AccessAllowed:   accessAllowed,
		AccessReason:    accessReason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods for database operations

func (h *geoRestrictionHandlers) verifyEmailOwnership(emailID, userID string) error {
	var senderID string
	err := h.db.QueryRow("SELECT sender_id FROM emails WHERE email_id = ?", emailID).Scan(&senderID)
	if err != nil {
		return err
	}
	if senderID != userID {
		return fmt.Errorf("user does not own this email")
	}
	return nil
}

func (h *geoRestrictionHandlers) getGeoRestrictionRules(emailID string) ([]georestriction.GeoRestrictionRule, error) {
	var rulesJSON sql.NullString
	err := h.db.QueryRow("SELECT geo_restriction_rules FROM emails WHERE email_id = ?", emailID).Scan(&rulesJSON)
	if err != nil {
		return nil, err
	}

	if !rulesJSON.Valid || rulesJSON.String == "" {
		return []georestriction.GeoRestrictionRule{}, nil
	}

	service := georestriction.NewGeoRestrictionService()
	return service.ParseRulesFromJSON([]byte(rulesJSON.String))
}

func (h *geoRestrictionHandlers) saveGeoRestrictionRule(rule georestriction.GeoRestrictionRule) error {
	// Get existing rules
	existingRules, err := h.getGeoRestrictionRules(rule.EmailID)
	if err != nil {
		return err
	}

	// Add new rule
	existingRules = append(existingRules, rule)

	// Save all rules
	return h.saveGeoRestrictionRules(rule.EmailID, existingRules)
}

func (h *geoRestrictionHandlers) saveGeoRestrictionRules(emailID string, rules []georestriction.GeoRestrictionRule) error {
	service := georestriction.NewGeoRestrictionService()
	rulesJSON, err := service.SerializeRulesToJSON(rules)
	if err != nil {
		return err
	}

	_, err = h.db.Exec("UPDATE emails SET geo_restriction_rules = ? WHERE email_id = ?", string(rulesJSON), emailID)
	return err
}

func (h *geoRestrictionHandlers) getGeoRestrictionConfig(emailID string) (georestriction.GeoRestrictionConfig, error) {
	var configJSON sql.NullString
	err := h.db.QueryRow("SELECT geo_restriction_config FROM emails WHERE email_id = ?", emailID).Scan(&configJSON)
	if err != nil {
		return georestriction.GeoRestrictionConfig{}, err
	}

	if !configJSON.Valid || configJSON.String == "" {
		// Return default config
		service := georestriction.NewGeoRestrictionService()
		return service.GetDefaultConfig(), nil
	}

	var config georestriction.GeoRestrictionConfig
	err = json.Unmarshal([]byte(configJSON.String), &config)
	return config, err
}

func (h *geoRestrictionHandlers) saveGeoRestrictionConfig(emailID string, config georestriction.GeoRestrictionConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = h.db.Exec("UPDATE emails SET geo_restriction_config = ? WHERE email_id = ?", string(configJSON), emailID)
	return err
}

func (h *geoRestrictionHandlers) getGeoRestrictionViolations(emailID string) (int, *time.Time, error) {
	var violationsCount int
	var lastViolation sql.NullTime

	err := h.db.QueryRow("SELECT geo_restriction_violations, geo_restriction_last_violation FROM emails WHERE email_id = ?", emailID).Scan(&violationsCount, &lastViolation)
	if err != nil {
		return 0, nil, err
	}

	if lastViolation.Valid {
		return violationsCount, &lastViolation.Time, nil
	}
	return violationsCount, nil, nil
}

func (h *geoRestrictionHandlers) incrementGeoRestrictionViolations(emailID string) error {
	_, err := h.db.Exec("UPDATE emails SET geo_restriction_violations = geo_restriction_violations + 1, geo_restriction_last_violation = CURRENT_TIMESTAMP WHERE email_id = ?", emailID)
	return err
}



// generateUUID generates a simple UUID-like string
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
