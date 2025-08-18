package geofencing

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"secure-email-mvp/pkg/geolocation"
)

// GeofencingService provides geofencing functionality for email access control
type GeofencingService struct {
	db             *sql.DB
	geolocationSvc geolocation.GeolocationService
}

// NewGeofencingService creates a new geofencing service
func NewGeofencingService(db *sql.DB, geolocationSvc geolocation.GeolocationService) *GeofencingService {
	return &GeofencingService{
		db:             db,
		geolocationSvc: geolocationSvc,
	}
}

// GeofenceResult represents the result of a geofence check
type GeofenceResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// CheckGeofenceAccess checks if a client can access an email based on geofencing rules
func (g *GeofencingService) CheckGeofenceAccess(emailID, clientIP string) (*GeofenceResult, error) {
	// Get geofencing settings for the email
	allowedCountries, allowedIPRanges, err := g.getGeofencingSettings(emailID)
	if err != nil {
		return nil, fmt.Errorf("failed to get geofencing settings: %w", err)
	}

	// If no restrictions are set, allow access
	if len(allowedCountries) == 0 && len(allowedIPRanges) == 0 {
		return &GeofenceResult{Allowed: true}, nil
	}

	// Check if access is allowed based on IP ranges OR country restrictions
	accessAllowed := false

	// Check IP range restrictions first
	if len(allowedIPRanges) > 0 {
		if g.geolocationSvc.IsIPInRange(clientIP, allowedIPRanges) {
			accessAllowed = true
		}
	}

	// If not allowed by IP ranges, check country restrictions
	if !accessAllowed && len(allowedCountries) > 0 {
		// Get client location
		location, err := g.geolocationSvc.GetLocation(clientIP)
		if err != nil {
			log.Printf("Failed to get location for IP %s: %v", clientIP, err)
			// If we can't determine location, deny access for security
			return &GeofenceResult{
				Allowed: false,
				Reason:  "geofence_location_unknown",
			}, nil
		}

		if g.geolocationSvc.IsCountryAllowed(location.Country, allowedCountries) {
			accessAllowed = true
		} else {
			log.Printf("Geofence violation: Country %s not in allowed countries %v for email %s",
				location.Country, allowedCountries, emailID)
		}
	}

	// If neither IP ranges nor countries are specified, allow access
	if len(allowedIPRanges) == 0 && len(allowedCountries) == 0 {
		accessAllowed = true
	}

	// If access is not allowed, increment violation counter and return
	if !accessAllowed {
		// Check if IP range violation occurred
		if len(allowedIPRanges) > 0 && !g.geolocationSvc.IsIPInRange(clientIP, allowedIPRanges) {
			log.Printf("Geofence violation: IP %s not in allowed ranges %v for email %s",
				clientIP, allowedIPRanges, emailID)
		}

		// Increment violation counter
		if err := g.incrementGeofenceViolations(emailID); err != nil {
			log.Printf("Failed to increment geofence violations: %v", err)
		}

		return &GeofenceResult{
			Allowed: false,
			Reason:  "geofence_blocked",
		}, nil
	}

	return &GeofenceResult{Allowed: true}, nil
}

// getGeofencingSettings retrieves geofencing settings for an email
func (g *GeofencingService) getGeofencingSettings(emailID string) ([]string, []string, error) {
	var allowedCountriesJSON, allowedIPRangesJSON sql.NullString

	err := g.db.QueryRow(`
		SELECT allowed_countries, allowed_ip_ranges 
		FROM emails 
		WHERE email_id = ?`, emailID).Scan(&allowedCountriesJSON, &allowedIPRangesJSON)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to query geofencing settings: %w", err)
	}

	// Parse allowed countries
	var allowedCountries []string
	if allowedCountriesJSON.Valid && allowedCountriesJSON.String != "" {
		allowedCountries, err = geolocation.ParseAllowedCountries(allowedCountriesJSON.String)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse allowed countries: %w", err)
		}
	}

	// Parse allowed IP ranges
	var allowedIPRanges []string
	if allowedIPRangesJSON.Valid && allowedIPRangesJSON.String != "" {
		allowedIPRanges, err = geolocation.ParseAllowedIPRanges(allowedIPRangesJSON.String)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse allowed IP ranges: %w", err)
		}
	}

	return allowedCountries, allowedIPRanges, nil
}

// incrementGeofenceViolations increments the geofence violation counter
func (g *GeofencingService) incrementGeofenceViolations(emailID string) error {
	_, err := g.db.Exec(`
		UPDATE emails 
		SET geofence_violations = geofence_violations + 1,
		    geofence_last_violation = CURRENT_TIMESTAMP
		WHERE email_id = ?`, emailID)

	if err != nil {
		return fmt.Errorf("failed to increment geofence violations: %w", err)
	}

	return nil
}

// GetGeofenceViolations gets the current geofence violation count
func (g *GeofencingService) GetGeofenceViolations(emailID string) (int, error) {
	var violations int
	err := g.db.QueryRow(`
		SELECT geofence_violations 
		FROM emails 
		WHERE email_id = ?`, emailID).Scan(&violations)

	if err != nil {
		return 0, fmt.Errorf("failed to get geofence violations: %w", err)
	}

	return violations, nil
}

// ResetGeofenceViolations resets the geofence violation counter
func (g *GeofencingService) ResetGeofenceViolations(emailID string) error {
	_, err := g.db.Exec(`
		UPDATE emails 
		SET geofence_violations = 0,
		    geofence_last_violation = NULL
		WHERE email_id = ?`, emailID)

	if err != nil {
		return fmt.Errorf("failed to reset geofence violations: %w", err)
	}

	return nil
}

// SetGeofencingSettings sets geofencing settings for an email
func (g *GeofencingService) SetGeofencingSettings(emailID string, allowedCountries, allowedIPRanges []string) error {
	// Validate country codes
	for _, country := range allowedCountries {
		if !geolocation.ValidateCountryCode(country) {
			return fmt.Errorf("invalid country code: %s", country)
		}
	}

	// Validate CIDR ranges
	for _, cidr := range allowedIPRanges {
		if !geolocation.ValidateCIDRRange(cidr) {
			return fmt.Errorf("invalid CIDR range: %s", cidr)
		}
	}

	// Convert to JSON
	var allowedCountriesJSON, allowedIPRangesJSON sql.NullString

	if len(allowedCountries) > 0 {
		jsonData, err := json.Marshal(allowedCountries)
		if err != nil {
			return fmt.Errorf("failed to marshal allowed countries: %w", err)
		}
		allowedCountriesJSON.String = string(jsonData)
		allowedCountriesJSON.Valid = true
	}

	if len(allowedIPRanges) > 0 {
		jsonData, err := json.Marshal(allowedIPRanges)
		if err != nil {
			return fmt.Errorf("failed to marshal allowed IP ranges: %w", err)
		}
		allowedIPRangesJSON.String = string(jsonData)
		allowedIPRangesJSON.Valid = true
	}

	// Update database
	_, err := g.db.Exec(`
		UPDATE emails 
		SET allowed_countries = ?,
		    allowed_ip_ranges = ?
		WHERE email_id = ?`, allowedCountriesJSON, allowedIPRangesJSON, emailID)

	if err != nil {
		return fmt.Errorf("failed to update geofencing settings: %w", err)
	}

	return nil
}

// GetGeofencingSettings retrieves geofencing settings for an email
func (g *GeofencingService) GetGeofencingSettings(emailID string) ([]string, []string, error) {
	return g.getGeofencingSettings(emailID)
}

// ValidateGeofencingSettings validates geofencing settings
func (g *GeofencingService) ValidateGeofencingSettings(allowedCountries, allowedIPRanges []string) error {
	// Validate country codes
	for _, country := range allowedCountries {
		if !geolocation.ValidateCountryCode(country) {
			return fmt.Errorf("invalid country code: %s (must be ISO 3166-1 alpha-2 format, e.g., US, CA, GB)", country)
		}
	}

	// Validate CIDR ranges
	for _, cidr := range allowedIPRanges {
		if !geolocation.ValidateCIDRRange(cidr) {
			return fmt.Errorf("invalid CIDR range: %s (must be valid CIDR notation, e.g., 192.168.1.0/24)", cidr)
		}
	}

	return nil
}

// FormatGeofencingDescription returns a human-readable description of geofencing settings
func (g *GeofencingService) FormatGeofencingDescription(allowedCountries, allowedIPRanges []string) string {
	var parts []string

	if len(allowedCountries) > 0 {
		parts = append(parts, fmt.Sprintf("Countries: %s", strings.Join(allowedCountries, ", ")))
	}

	if len(allowedIPRanges) > 0 {
		parts = append(parts, fmt.Sprintf("IP Ranges: %s", strings.Join(allowedIPRanges, ", ")))
	}

	if len(parts) == 0 {
		return "No geofencing restrictions"
	}

	return strings.Join(parts, "; ")
}
