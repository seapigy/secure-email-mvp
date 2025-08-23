package geolocation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// =============================================================================
// ENHANCED GEOLOCATION VERIFICATION SERVICE
// =============================================================================

// GeolocationVerificationService handles enhanced geolocation verification
type GeolocationVerificationService struct {
	apiURL string
	client *http.Client
}

// GeolocationData represents geolocation information from IP
type GeolocationData struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	City        string `json:"city"`
	Region      string `json:"region"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	Timezone    string `json:"timezone"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
}

// GeolocationRestriction represents geolocation restrictions
type GeolocationRestriction struct {
	Enabled          bool     `json:"enabled"`
	AllowedCountries []string `json:"allowed_countries,omitempty"`
	AllowedCities    []string `json:"allowed_cities,omitempty"`
	BlockedCountries []string `json:"blocked_countries,omitempty"`
	BlockedCities    []string `json:"blocked_cities,omitempty"`
}

// GeolocationVerificationResult represents the result of geolocation verification
type GeolocationVerificationResult struct {
	Allowed     bool                    `json:"allowed"`
	Reason      string                  `json:"reason,omitempty"`
	Location    *GeolocationData        `json:"location,omitempty"`
	Restriction *GeolocationRestriction `json:"restriction,omitempty"`
}

// NewGeolocationVerificationService creates a new geolocation verification service
func NewGeolocationVerificationService() *GeolocationVerificationService {
	return &GeolocationVerificationService{
		apiURL: "http://ip-api.com/json/",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyLocation verifies if an IP address is allowed based on geolocation restrictions
func (g *GeolocationVerificationService) VerifyLocation(ctx context.Context, ipAddress string, restriction GeolocationRestriction) (*GeolocationVerificationResult, error) {
	// If geolocation restrictions are not enabled, allow access
	if !restriction.Enabled {
		return &GeolocationVerificationResult{
			Allowed: true,
			Reason:  "Geolocation restrictions not enabled",
		}, nil
	}

	// Get geolocation data for the IP address
	location, err := g.getGeolocationData(ctx, ipAddress)
	if err != nil {
		return &GeolocationVerificationResult{
			Allowed: false,
			Reason:  "Failed to determine location",
		}, fmt.Errorf("failed to get geolocation data: %w", err)
	}

	// Check if the location is blocked
	if g.isLocationBlocked(location, restriction) {
		return &GeolocationVerificationResult{
			Allowed:     false,
			Reason:      "Access denied from your location",
			Location:    location,
			Restriction: &restriction,
		}, nil
	}

	// Check if the location is allowed
	if g.isLocationAllowed(location, restriction) {
		return &GeolocationVerificationResult{
			Allowed:     true,
			Reason:      "Location access granted",
			Location:    location,
			Restriction: &restriction,
		}, nil
	}

	// If we have allowlists but the location is not in them, deny access
	if len(restriction.AllowedCountries) > 0 || len(restriction.AllowedCities) > 0 {
		return &GeolocationVerificationResult{
			Allowed:     false,
			Reason:      "Access denied from your location",
			Location:    location,
			Restriction: &restriction,
		}, nil
	}

	// Default: allow access if no specific restrictions
	return &GeolocationVerificationResult{
		Allowed:     true,
		Reason:      "No specific location restrictions",
		Location:    location,
		Restriction: &restriction,
	}, nil
}

// getGeolocationData retrieves geolocation data for an IP address
func (g *GeolocationVerificationService) getGeolocationData(ctx context.Context, ipAddress string) (*GeolocationData, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", g.apiURL+ipAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "SecureEmail-Geolocation/1.0")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var location GeolocationData
	if err := json.Unmarshal(body, &location); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Set IP address
	location.IP = ipAddress

	return &location, nil
}

// isLocationBlocked checks if a location is blocked
func (g *GeolocationVerificationService) isLocationBlocked(location *GeolocationData, restriction GeolocationRestriction) bool {
	// Check blocked countries
	for _, blockedCountry := range restriction.BlockedCountries {
		if strings.EqualFold(location.CountryCode, blockedCountry) {
			return true
		}
	}

	// Check blocked cities
	for _, blockedCity := range restriction.BlockedCities {
		if strings.EqualFold(normalizeCityName(location.City), normalizeCityName(blockedCity)) {
			return true
		}
	}

	return false
}

// isLocationAllowed checks if a location is allowed
func (g *GeolocationVerificationService) isLocationAllowed(location *GeolocationData, restriction GeolocationRestriction) bool {
	// If no allowlists are specified, allow access
	if len(restriction.AllowedCountries) == 0 && len(restriction.AllowedCities) == 0 {
		return true
	}

	// Check allowed countries
	countryAllowed := false
	if len(restriction.AllowedCountries) == 0 {
		countryAllowed = true // No country restrictions
	} else {
		for _, allowedCountry := range restriction.AllowedCountries {
			if strings.EqualFold(location.CountryCode, allowedCountry) {
				countryAllowed = true
				break
			}
		}
	}

	// Check allowed cities
	cityAllowed := false
	if len(restriction.AllowedCities) == 0 {
		cityAllowed = true // No city restrictions
	} else {
		for _, allowedCity := range restriction.AllowedCities {
			if strings.EqualFold(normalizeCityName(location.City), normalizeCityName(allowedCity)) {
				cityAllowed = true
				break
			}
		}
	}

	// Both country and city must be allowed if both are specified
	if len(restriction.AllowedCountries) > 0 && len(restriction.AllowedCities) > 0 {
		return countryAllowed && cityAllowed
	}

	// If only one type is specified, that one must be allowed
	return countryAllowed || cityAllowed
}

// normalizeCityName normalizes city names for comparison
func normalizeCityName(city string) string {
	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(city))

	// Remove common suffixes and prefixes
	normalized = strings.ReplaceAll(normalized, " city", "")
	normalized = strings.ReplaceAll(normalized, " town", "")
	normalized = strings.ReplaceAll(normalized, " village", "")

	// Remove special characters
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")

	// Remove extra spaces
	normalized = strings.Join(strings.Fields(normalized), " ")

	return normalized
}

// ValidateGeolocationRestriction validates a geolocation restriction configuration
func (g *GeolocationVerificationService) ValidateGeolocationRestriction(restriction GeolocationRestriction) error {
	// Check for conflicting allowlist and blocklist
	for _, allowedCountry := range restriction.AllowedCountries {
		for _, blockedCountry := range restriction.BlockedCountries {
			if strings.EqualFold(allowedCountry, blockedCountry) {
				return fmt.Errorf("country %s appears in both allowlist and blocklist", allowedCountry)
			}
		}
	}

	for _, allowedCity := range restriction.AllowedCities {
		for _, blockedCity := range restriction.BlockedCities {
			if strings.EqualFold(normalizeCityName(allowedCity), normalizeCityName(blockedCity)) {
				return fmt.Errorf("city %s appears in both allowlist and blocklist", allowedCity)
			}
		}
	}

	return nil
}

// GetGeolocationDataForIP retrieves geolocation data for an IP address without restrictions
func (g *GeolocationVerificationService) GetGeolocationDataForIP(ctx context.Context, ipAddress string) (*GeolocationData, error) {
	return g.getGeolocationData(ctx, ipAddress)
}
