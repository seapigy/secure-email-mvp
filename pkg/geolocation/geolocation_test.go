package geolocation

import (
	"strings"
	"testing"
)

func TestNormalizeCityName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"New York", "new york"},
		{"  Los Angeles  ", "los angeles"},
		{"San Francisco", "san francisco"},
		{"  New   York  ", "new york"},
		{"", ""},
		{"   ", ""},
		{"Toronto", "toronto"},
		{"London", "london"},
		{"Paris", "paris"},
		{"Tokyo", "tokyo"},
	}

	for _, test := range tests {
		result := NormalizeCityName(test.input)
		if result != test.expected {
			t.Errorf("NormalizeCityName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"US", true},
		{"CA", true},
		{"GB", true},
		{"DE", true},
		{"FR", true},
		{"JP", true},
		{"AU", true},
		{"BR", true},
		{"IN", true},
		{"CN", true},
		{"us", true},   // lowercase should be valid
		{"ca", true},   // lowercase should be valid
		{"USA", false}, // too long
		{"U", false},   // too short
		{"12", false},  // numbers not allowed
		{"", false},    // empty
		{"A1", false},  // mixed
		{"1A", false},  // mixed
	}

	for _, test := range tests {
		result := ValidateCountryCode(test.input)
		if result != test.expected {
			t.Errorf("ValidateCountryCode(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestValidateCityName(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"New York", true},
		{"Los Angeles", true},
		{"San Francisco", true},
		{"Toronto", true},
		{"London", true},
		{"Paris", true},
		{"Tokyo", true},
		{"New-York", true},     // hyphen allowed
		{"O'Connor", true},     // apostrophe allowed
		{"A", false},           // too short
		{"", false},            // empty
		{"New York123", false}, // numbers not allowed
		{"New@York", false},    // special characters not allowed
		{"New#York", false},    // special characters not allowed
		{"New$York", false},    // special characters not allowed
		{"New%York", false},    // special characters not allowed
		{"New^York", false},    // special characters not allowed
		{"New&York", false},    // special characters not allowed
		{"New*York", false},    // special characters not allowed
		{"New(York", false},    // special characters not allowed
		{"New)York", false},    // special characters not allowed
		{"New[York", false},    // special characters not allowed
		{"New]York", false},    // special characters not allowed
		{"New{York", false},    // special characters not allowed
		{"New}York", false},    // special characters not allowed
		{"New|York", false},    // special characters not allowed
		{"New\\York", false},   // special characters not allowed
		{"New/York", false},    // special characters not allowed
		{"New?York", false},    // special characters not allowed
		{"New!York", false},    // special characters not allowed
		{"New:York", false},    // special characters not allowed
		{"New;York", false},    // special characters not allowed
		{"New\"York", false},   // special characters not allowed
		{"New'York", true},     // apostrophe allowed
		{"'New York'", true},   // apostrophe allowed at start/end
	}

	for _, test := range tests {
		result := ValidateCityName(test.input)
		if result != test.expected {
			t.Errorf("ValidateCityName(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestExactGeolocationMatching(t *testing.T) {
	// Test exact matching logic (this would be used in the view_email_handler)
	tests := []struct {
		clientCountry  string
		clientCity     string
		allowedCountry string
		allowedCity    string
		expected       bool
	}{
		// No restrictions
		{"us", "new york", "", "", true},

		// Country only restrictions
		{"us", "new york", "us", "", true},
		{"us", "new york", "ca", "", false},
		{"ca", "toronto", "us", "", false},
		{"ca", "toronto", "ca", "", true},

		// City only restrictions
		{"us", "new york", "", "new york", true},
		{"us", "new york", "", "los angeles", false},
		{"ca", "toronto", "", "new york", false},
		{"ca", "toronto", "", "toronto", true},

		// Both restrictions
		{"us", "new york", "us", "new york", true},
		{"us", "new york", "us", "los angeles", false},
		{"us", "new york", "ca", "new york", false},
		{"us", "new york", "ca", "los angeles", false},
		{"ca", "toronto", "us", "new york", false},
		{"ca", "toronto", "ca", "toronto", true},

		// Case sensitivity tests
		{"US", "NEW YORK", "us", "new york", true},
		{"us", "New York", "US", "NEW YORK", true},
		{"CA", "TORONTO", "ca", "toronto", true},

		// Whitespace tests
		{" us ", " new york ", "us", "new york", true},
		{"us", "  new york  ", "us", "new york", true},
	}

	for _, test := range tests {
		// Simulate the exact matching logic from view_email_handler
		accessAllowed := true

		// Check country restriction if set
		if test.allowedCountry != "" {
			normalizedClientCountry := strings.ToLower(strings.TrimSpace(test.clientCountry))
			normalizedAllowedCountry := strings.ToLower(strings.TrimSpace(test.allowedCountry))
			if normalizedClientCountry != normalizedAllowedCountry {
				accessAllowed = false
			}
		}

		// Check city restriction if set
		if test.allowedCity != "" {
			normalizedClientCity := NormalizeCityName(test.clientCity)
			normalizedAllowedCity := NormalizeCityName(test.allowedCity)
			if normalizedClientCity != normalizedAllowedCity {
				accessAllowed = false
			}
		}

		if accessAllowed != test.expected {
			t.Errorf("Geolocation matching failed: client(%s, %s) vs allowed(%s, %s) = %v, expected %v",
				test.clientCountry, test.clientCity, test.allowedCountry, test.allowedCity, accessAllowed, test.expected)
		}
	}
}
