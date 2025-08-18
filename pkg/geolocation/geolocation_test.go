package geolocation

import (
	"testing"
)

// TestMockGeolocationService tests the mock geolocation service
func TestMockGeolocationService(t *testing.T) {
	// Create mock service
	mockSvc := NewMockGeolocationService()

	// Test setting and getting location
	testLocation := &Location{
		Country: "US",
		City:    "New York",
		IP:      "192.168.1.1",
	}

	mockSvc.SetLocation("192.168.1.1", testLocation)

	// Test getting the location
	location, err := mockSvc.GetLocation("192.168.1.1")
	if err != nil {
		t.Errorf("GetLocation() error = %v", err)
		return
	}

	if location.Country != testLocation.Country {
		t.Errorf("GetLocation() country = %v, want %v", location.Country, testLocation.Country)
	}

	if location.City != testLocation.City {
		t.Errorf("GetLocation() city = %v, want %v", location.City, testLocation.City)
	}

	if location.IP != testLocation.IP {
		t.Errorf("GetLocation() IP = %v, want %v", location.IP, testLocation.IP)
	}
}

// TestMockGeolocationService_IsCountryAllowed tests country allowance checking
func TestMockGeolocationService_IsCountryAllowed(t *testing.T) {
	mockSvc := NewMockGeolocationService()

	tests := []struct {
		name             string
		clientCountry    string
		allowedCountries []string
		expected         bool
	}{
		{
			name:             "No restrictions",
			clientCountry:    "US",
			allowedCountries: []string{},
			expected:         true,
		},
		{
			name:             "Country allowed",
			clientCountry:    "US",
			allowedCountries: []string{"US", "CA"},
			expected:         true,
		},
		{
			name:             "Country not allowed",
			clientCountry:    "CA",
			allowedCountries: []string{"US"},
			expected:         false,
		},
		{
			name:             "Case insensitive",
			clientCountry:    "us",
			allowedCountries: []string{"US"},
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockSvc.IsCountryAllowed(tt.clientCountry, tt.allowedCountries)
			if result != tt.expected {
				t.Errorf("IsCountryAllowed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMockGeolocationService_IsIPInRange tests IP range checking
func TestMockGeolocationService_IsIPInRange(t *testing.T) {
	mockSvc := NewMockGeolocationService()

	tests := []struct {
		name           string
		clientIP       string
		allowedRanges  []string
		expected       bool
	}{
		{
			name:           "No restrictions",
			clientIP:       "192.168.1.1",
			allowedRanges:  []string{},
			expected:       true,
		},
		{
			name:           "IP in range",
			clientIP:       "192.168.1.100",
			allowedRanges:  []string{"192.168.1.0/24"},
			expected:       true,
		},
		{
			name:           "IP not in range",
			clientIP:       "10.0.0.1",
			allowedRanges:  []string{"192.168.1.0/24"},
			expected:       false,
		},
		{
			name:           "Invalid IP",
			clientIP:       "invalid-ip",
			allowedRanges:  []string{"192.168.1.0/24"},
			expected:       false,
		},
		{
			name:           "Invalid CIDR range",
			clientIP:       "192.168.1.1",
			allowedRanges:  []string{"invalid-cidr"},
			expected:       false, // Invalid ranges should be treated as no match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockSvc.IsIPInRange(tt.clientIP, tt.allowedRanges)
			if result != tt.expected {
				t.Errorf("IsIPInRange() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSimpleGeolocationService tests the simple geolocation service
func TestSimpleGeolocationService(t *testing.T) {
	simpleSvc := NewSimpleGeolocationService()

	// Test valid IP
	location, err := simpleSvc.GetLocation("192.168.1.1")
	if err != nil {
		t.Errorf("GetLocation() error = %v", err)
		return
	}

	if location == nil {
		t.Errorf("GetLocation() returned nil location")
		return
	}

	if location.IP != "192.168.1.1" {
		t.Errorf("GetLocation() IP = %v, want %v", location.IP, "192.168.1.1")
	}

	// Test invalid IP
	_, err = simpleSvc.GetLocation("invalid-ip")
	if err == nil {
		t.Errorf("GetLocation() expected error for invalid IP")
	}
}

// TestParseAllowedCountries tests parsing of allowed countries JSON
func TestParseAllowedCountries(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected []string
		hasError bool
	}{
		{
			name:     "Empty string",
			jsonStr:  "",
			expected: []string{},
			hasError: false,
		},
		{
			name:     "Valid JSON array",
			jsonStr:  `["US", "CA", "GB"]`,
			expected: []string{"US", "CA", "GB"},
			hasError: false,
		},
		{
			name:     "Invalid JSON",
			jsonStr:  `["US", "CA"`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAllowedCountries(tt.jsonStr)
			
			if tt.hasError && err == nil {
				t.Errorf("ParseAllowedCountries() expected error but got none")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("ParseAllowedCountries() unexpected error: %v", err)
			}
			
			if !tt.hasError {
				if len(result) != len(tt.expected) {
					t.Errorf("ParseAllowedCountries() length = %v, want %v", len(result), len(tt.expected))
				}
				
				for i, country := range result {
					if country != tt.expected[i] {
						t.Errorf("ParseAllowedCountries()[%d] = %v, want %v", i, country, tt.expected[i])
					}
				}
			}
		})
	}
}

// TestParseAllowedIPRanges tests parsing of allowed IP ranges JSON
func TestParseAllowedIPRanges(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected []string
		hasError bool
	}{
		{
			name:     "Empty string",
			jsonStr:  "",
			expected: []string{},
			hasError: false,
		},
		{
			name:     "Valid JSON array",
			jsonStr:  `["192.168.1.0/24", "10.0.0.0/8"]`,
			expected: []string{"192.168.1.0/24", "10.0.0.0/8"},
			hasError: false,
		},
		{
			name:     "Invalid JSON",
			jsonStr:  `["192.168.1.0/24"`,
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid CIDR range",
			jsonStr:  `["192.168.1.0/24", "invalid-cidr"]`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAllowedIPRanges(tt.jsonStr)
			
			if tt.hasError && err == nil {
				t.Errorf("ParseAllowedIPRanges() expected error but got none")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("ParseAllowedIPRanges() unexpected error: %v", err)
			}
			
			if !tt.hasError {
				if len(result) != len(tt.expected) {
					t.Errorf("ParseAllowedIPRanges() length = %v, want %v", len(result), len(tt.expected))
				}
				
				for i, cidr := range result {
					if cidr != tt.expected[i] {
						t.Errorf("ParseAllowedIPRanges()[%d] = %v, want %v", i, cidr, tt.expected[i])
					}
				}
			}
		})
	}
}

// TestValidateCountryCode tests country code validation
func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "Valid country code",
			code:     "US",
			expected: true,
		},
		{
			name:     "Valid country code lowercase",
			code:     "us",
			expected: false, // Should be uppercase
		},
		{
			name:     "Too short",
			code:     "U",
			expected: false,
		},
		{
			name:     "Too long",
			code:     "USA",
			expected: false,
		},
		{
			name:     "Contains numbers",
			code:     "U1",
			expected: false,
		},
		{
			name:     "Empty string",
			code:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCountryCode(tt.code)
			if result != tt.expected {
				t.Errorf("ValidateCountryCode(%s) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

// TestValidateCIDRRange tests CIDR range validation
func TestValidateCIDRRange(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		expected bool
	}{
		{
			name:     "Valid CIDR range",
			cidr:     "192.168.1.0/24",
			expected: true,
		},
		{
			name:     "Valid CIDR range with spaces",
			cidr:     " 192.168.1.0/24 ",
			expected: true,
		},
		{
			name:     "Invalid CIDR range",
			cidr:     "192.168.1.0",
			expected: false,
		},
		{
			name:     "Invalid IP",
			cidr:     "invalid/24",
			expected: false,
		},
		{
			name:     "Invalid mask",
			cidr:     "192.168.1.0/invalid",
			expected: false,
		},
		{
			name:     "Empty string",
			cidr:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCIDRRange(tt.cidr)
			if result != tt.expected {
				t.Errorf("ValidateCIDRRange(%s) = %v, want %v", tt.cidr, result, tt.expected)
			}
		})
	}
}
