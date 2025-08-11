package geolocation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Location represents the geolocation data for an IP address
type Location struct {
	Country string `json:"country"` // ISO 3166-1 alpha-2 country code (e.g., "US")
	City    string `json:"city"`    // City name (e.g., "New York")
	IP      string `json:"ip"`      // The IP address that was geolocated
}

// GeolocationService provides methods for IP-based geolocation
type GeolocationService struct {
	client *http.Client
}

// NewGeolocationService creates a new geolocation service instance
func NewGeolocationService() *GeolocationService {
	return &GeolocationService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLocationByIP retrieves geolocation information for an IP address
// Uses ipapi.co as the geolocation provider (free tier: 1,000 requests/day)
func (g *GeolocationService) GetLocationByIP(ip string) (*Location, error) {
	// Clean and validate IP address
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil, fmt.Errorf("empty IP address")
	}

	// Use ipapi.co for geolocation (free service)
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,countryCode,city,query", ip)

	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geolocation service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var result struct {
		Status      string `json:"status"`
		Message     string `json:"message,omitempty"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
		Query       string `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse geolocation response: %w", err)
	}

	// Check if the request was successful
	if result.Status != "success" {
		return nil, fmt.Errorf("geolocation service error: %s", result.Message)
	}

	// Normalize the data
	location := &Location{
		Country: strings.ToLower(strings.TrimSpace(result.CountryCode)),
		City:    normalizeCityName(result.City),
		IP:      result.Query,
	}

	return location, nil
}

// normalizeCityName normalizes a city name for comparison
// Converts to lowercase, trims whitespace, and removes extra spaces
func normalizeCityName(city string) string {
	if city == "" {
		return ""
	}

	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(city))

	// Replace multiple spaces with single space
	normalized = strings.Join(strings.Fields(normalized), " ")

	return normalized
}

// NormalizeCityName is a public wrapper for normalizeCityName
func NormalizeCityName(city string) string {
	return normalizeCityName(city)
}

// ValidateCountryCode validates if a country code is a valid ISO 3166-1 alpha-2 code
func ValidateCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}

	// Convert to lowercase for validation
	code = strings.ToLower(code)

	// Check if it contains only letters
	for _, char := range code {
		if char < 'a' || char > 'z' {
			return false
		}
	}

	return true
}

// ValidateCityName validates if a city name is reasonable
func ValidateCityName(city string) bool {
	if city == "" {
		return false
	}

	// Check minimum and maximum length
	if len(city) < 2 || len(city) > 100 {
		return false
	}

	// Check if it contains only letters, spaces, hyphens, and apostrophes
	for _, char := range city {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			char == ' ' || char == '-' || char == '\'') {
			return false
		}
	}

	return true
}

// CheckGeolocationRestrictions checks if a location passes the geolocation restrictions
// Returns true if access is allowed, false if blocked
func CheckGeolocationRestrictions(location *Location, allowedCountries []string, allowedCities []string) (bool, string) {
	// If no restrictions are set, allow access
	if len(allowedCountries) == 0 && len(allowedCities) == 0 {
		return true, ""
	}

	// Check country restriction if set
	countryAllowed := true
	if len(allowedCountries) > 0 {
		countryAllowed = false
		for _, allowedCountry := range allowedCountries {
			if strings.ToLower(strings.TrimSpace(allowedCountry)) == location.Country {
				countryAllowed = true
				break
			}
		}
	}

	// Check city restriction if set
	cityAllowed := true
	if len(allowedCities) > 0 {
		cityAllowed = false
		for _, allowedCity := range allowedCities {
			if normalizeCityName(allowedCity) == location.City {
				cityAllowed = true
				break
			}
		}
	}

	// Both restrictions must pass (AND logic)
	if !countryAllowed || !cityAllowed {
		var reason string
		if !countryAllowed && !cityAllowed {
			reason = fmt.Sprintf("Access blocked: Your location (%s, %s) is not in the allowed countries or cities.", location.City, strings.ToUpper(location.Country))
		} else if !countryAllowed {
			reason = fmt.Sprintf("Access blocked: Your country (%s) is not in the allowed countries.", strings.ToUpper(location.Country))
		} else {
			reason = fmt.Sprintf("Access blocked: Your city (%s) is not in the allowed cities.", location.City)
		}
		return false, reason
	}

	return true, ""
}
