package geolocation

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// Location represents geographic location information
type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
	IP      string `json:"ip"`
}

// GeolocationService provides IP-to-location mapping functionality
type GeolocationService interface {
	GetLocation(ip string) (*Location, error)
	GetLocationByIP(ip string) (*Location, error) // Alias for GetLocation for compatibility
	IsCountryAllowed(clientCountry string, allowedCountries []string) bool
	IsIPInRange(clientIP string, allowedRanges []string) bool
}

// MockGeolocationService is a mock implementation for testing
type MockGeolocationService struct {
	locations map[string]*Location
}

// NewMockGeolocationService creates a new mock geolocation service
func NewMockGeolocationService() *MockGeolocationService {
	return &MockGeolocationService{
		locations: make(map[string]*Location),
	}
}

// SetLocation sets a mock location for a specific IP
func (m *MockGeolocationService) SetLocation(ip string, location *Location) {
	m.locations[ip] = location
}

// GetLocation returns the mock location for an IP
func (m *MockGeolocationService) GetLocation(ip string) (*Location, error) {
	if location, exists := m.locations[ip]; exists {
		return location, nil
	}

	// Default fallback for unknown IPs
	return &Location{
		Country: "US",
		City:    "Unknown",
		IP:      ip,
	}, nil
}

// GetLocationByIP is an alias for GetLocation for compatibility
func (m *MockGeolocationService) GetLocationByIP(ip string) (*Location, error) {
	return m.GetLocation(ip)
}

// IsCountryAllowed checks if a client country is in the allowed list
func (m *MockGeolocationService) IsCountryAllowed(clientCountry string, allowedCountries []string) bool {
	if len(allowedCountries) == 0 {
		return true // No restrictions
	}

	clientCountry = strings.ToUpper(strings.TrimSpace(clientCountry))
	for _, allowed := range allowedCountries {
		if strings.ToUpper(strings.TrimSpace(allowed)) == clientCountry {
			return true
		}
	}
	return false
}

// IsIPInRange checks if a client IP is in any of the allowed CIDR ranges
func (m *MockGeolocationService) IsIPInRange(clientIP string, allowedRanges []string) bool {
	if len(allowedRanges) == 0 {
		return true // No restrictions
	}

	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		return false // Invalid IP
	}

	for _, cidrRange := range allowedRanges {
		_, network, err := net.ParseCIDR(strings.TrimSpace(cidrRange))
		if err != nil {
			continue // Skip invalid CIDR ranges
		}

		if network.Contains(clientIPAddr) {
			return true
		}
	}
	return false
}

// SimpleGeolocationService is a simple implementation using a basic IP-to-country mapping
type SimpleGeolocationService struct{}

// NewSimpleGeolocationService creates a new simple geolocation service
func NewSimpleGeolocationService() *SimpleGeolocationService {
	return &SimpleGeolocationService{}
}

// NewGeolocationService creates a new geolocation service (alias for NewSimpleGeolocationService for compatibility)
func NewGeolocationService() *SimpleGeolocationService {
	return NewSimpleGeolocationService()
}

// GetLocation returns location information for an IP address
// This is a simplified implementation - in production, you'd use MaxMind GeoLite2 or similar
func (s *SimpleGeolocationService) GetLocation(ip string) (*Location, error) {
	// Parse IP to validate format
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Simple country detection based on IP ranges
	// This is a basic implementation - in production, use a proper geolocation database
	country := s.detectCountry(ipAddr)

	return &Location{
		Country: country,
		City:    "Unknown", // Would be populated by a real geolocation service
		IP:      ip,
	}, nil
}

// GetLocationByIP is an alias for GetLocation for compatibility
func (s *SimpleGeolocationService) GetLocationByIP(ip string) (*Location, error) {
	return s.GetLocation(ip)
}

// IsCountryAllowed checks if a client country is in the allowed list
func (s *SimpleGeolocationService) IsCountryAllowed(clientCountry string, allowedCountries []string) bool {
	if len(allowedCountries) == 0 {
		return true // No restrictions
	}

	clientCountry = strings.ToUpper(strings.TrimSpace(clientCountry))
	for _, allowed := range allowedCountries {
		if strings.ToUpper(strings.TrimSpace(allowed)) == clientCountry {
			return true
		}
	}
	return false
}

// IsIPInRange checks if a client IP is in any of the allowed CIDR ranges
func (s *SimpleGeolocationService) IsIPInRange(clientIP string, allowedRanges []string) bool {
	if len(allowedRanges) == 0 {
		return true // No restrictions
	}

	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		return false // Invalid IP
	}

	for _, cidrRange := range allowedRanges {
		_, network, err := net.ParseCIDR(strings.TrimSpace(cidrRange))
		if err != nil {
			continue // Skip invalid CIDR ranges
		}

		if network.Contains(clientIPAddr) {
			return true
		}
	}
	return false
}

// detectCountry provides basic country detection based on IP ranges
// This is a simplified implementation - in production, use MaxMind GeoLite2
func (s *SimpleGeolocationService) detectCountry(ip net.IP) string {
	// Convert to string for easier comparison
	ipStr := ip.String()

	// Basic country detection based on common IP ranges
	// This is not comprehensive and should be replaced with a proper geolocation database

	// US ranges (simplified)
	if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "172.") {
		return "US" // Private IP ranges
	}

	// Example public ranges (this is just for demonstration)
	if strings.HasPrefix(ipStr, "8.8.") || strings.HasPrefix(ipStr, "1.1.") {
		return "US"
	}

	// Default to US for unknown ranges
	return "US"
}

// ParseAllowedCountries parses a JSON array of country codes
func ParseAllowedCountries(countriesJSON string) ([]string, error) {
	if countriesJSON == "" {
		return []string{}, nil
	}

	var countries []string
	err := json.Unmarshal([]byte(countriesJSON), &countries)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed countries JSON: %w", err)
	}

	return countries, nil
}

// ParseAllowedIPRanges parses a JSON array of CIDR ranges
func ParseAllowedIPRanges(rangesJSON string) ([]string, error) {
	if rangesJSON == "" {
		return []string{}, nil
	}

	var ranges []string
	err := json.Unmarshal([]byte(rangesJSON), &ranges)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed IP ranges JSON: %w", err)
	}

	// Validate CIDR ranges
	for _, cidr := range ranges {
		_, _, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR range %s: %w", cidr, err)
		}
	}

	return ranges, nil
}

// ValidateCountryCode validates if a country code is in ISO 3166-1 alpha-2 format
func ValidateCountryCode(countryCode string) bool {
	if len(countryCode) != 2 {
		return false
	}

	// Check if it's all uppercase letters
	for _, char := range countryCode {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	return true
}

// ValidateCIDRRange validates if a string is a valid CIDR range
func ValidateCIDRRange(cidr string) bool {
	_, _, err := net.ParseCIDR(strings.TrimSpace(cidr))
	return err == nil
}

// NormalizeCityName normalizes a city name for comparison
func NormalizeCityName(city string) string {
	if city == "" {
		return ""
	}

	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(city))

	// Remove common punctuation and special characters
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, ",", "")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")

	// Replace multiple spaces with single space
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	return strings.TrimSpace(normalized)
}

// ValidateCityName validates if a city name is valid
func ValidateCityName(city string) bool {
	if city == "" {
		return false
	}

	// Check if city name is too short or too long
	if len(strings.TrimSpace(city)) < 2 || len(city) > 100 {
		return false
	}

	// Check if city name contains only valid characters
	// Allow letters, spaces, hyphens, apostrophes, and periods
	for _, char := range city {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == ' ' || char == '-' || char == '\'' || char == '.') {
			return false
		}
	}

	return true
}
