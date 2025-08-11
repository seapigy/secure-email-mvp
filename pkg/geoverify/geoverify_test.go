package geoverify

import (
	"testing"

	"secure-email-mvp/pkg/geolocation"
)

func TestNewGeolocationVerifier(t *testing.T) {
	verifier := NewGeolocationVerifier()
	if verifier == nil {
		t.Fatal("NewGeolocationVerifier returned nil")
	}
}

func TestVerifyLocation_NoVerification(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test with no verification required
	result := verifier.VerifyLocation(
		VerificationTypeNone,
		&geolocation.Location{Country: "us", City: "new york"},
		"",
		"",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for no verification, got: %v", result)
	}
	if result.Reason != "" {
		t.Errorf("Expected no reason for allowed access, got: %s", result.Reason)
	}
}

func TestVerifyLocation_CityOnly_Success(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city-only verification with matching city
	result := verifier.VerifyLocation(
		VerificationTypeCity,
		&geolocation.Location{Country: "us", City: "New York"},
		"new york",
		"",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for matching city, got: %v", result)
	}
}

func TestVerifyLocation_CityOnly_Failure(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city-only verification with non-matching city
	result := verifier.VerifyLocation(
		VerificationTypeCity,
		&geolocation.Location{Country: "us", City: "Los Angeles"},
		"new york",
		"",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for non-matching city, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CountryOnly_Success(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test country-only verification with matching country
	result := verifier.VerifyLocation(
		VerificationTypeCountry,
		&geolocation.Location{Country: "us", City: "New York"},
		"",
		"us",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for matching country, got: %v", result)
	}
}

func TestVerifyLocation_CountryOnly_Failure(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test country-only verification with non-matching country
	result := verifier.VerifyLocation(
		VerificationTypeCountry,
		&geolocation.Location{Country: "ca", City: "Toronto"},
		"",
		"us",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for non-matching country, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CountryOnly_EmptyRequiredCountry(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test country-only verification with empty required country
	result := verifier.VerifyLocation(
		VerificationTypeCountry,
		&geolocation.Location{Country: "us", City: "New York"},
		"",
		"",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for empty required country, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CityCountry_Success(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city+country verification with matching values
	result := verifier.VerifyLocation(
		VerificationTypeCityCountry,
		&geolocation.Location{Country: "us", City: "New York"},
		"new york",
		"us",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for matching city and country, got: %v", result)
	}
}

func TestVerifyLocation_CityCountry_CityMismatch(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city+country verification with city mismatch
	result := verifier.VerifyLocation(
		VerificationTypeCityCountry,
		&geolocation.Location{Country: "us", City: "Los Angeles"},
		"new york",
		"us",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for city mismatch, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CityCountry_CountryMismatch(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city+country verification with country mismatch
	result := verifier.VerifyLocation(
		VerificationTypeCityCountry,
		&geolocation.Location{Country: "ca", City: "New York"},
		"new york",
		"us",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for country mismatch, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_NilLocation(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test with nil location
	result := verifier.VerifyLocation(
		VerificationTypeCity,
		nil,
		"new york",
		"",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for nil location, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CityOnly_EmptyRequiredCity(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city-only verification with empty required city
	result := verifier.VerifyLocation(
		VerificationTypeCity,
		&geolocation.Location{Country: "us", City: "New York"},
		"",
		"",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for empty required city, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestVerifyLocation_CityCountry_EmptyRequiredCountry(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test city+country verification with empty required country
	result := verifier.VerifyLocation(
		VerificationTypeCityCountry,
		&geolocation.Location{Country: "us", City: "New York"},
		"new york",
		"",
	)

	if result.Allowed {
		t.Errorf("Expected access to be denied for empty required country, got: %v", result)
	}
	if result.Reason == "" {
		t.Error("Expected reason for denied access")
	}
}

func TestValidateVerificationType(t *testing.T) {
	verifier := NewGeolocationVerifier()

	tests := []struct {
		verificationType string
		expectedError    bool
	}{
		{"none", false},
		{"country", false},
		{"city", false},
		{"city_country", false},
		{"invalid", true},
		{"", true},
		{"CITY", true},
	}

	for _, test := range tests {
		err := verifier.ValidateVerificationType(test.verificationType)
		if test.expectedError && err == nil {
			t.Errorf("Expected error for verification type '%s', got none", test.verificationType)
		}
		if !test.expectedError && err != nil {
			t.Errorf("Expected no error for verification type '%s', got: %v", test.verificationType, err)
		}
	}
}

func TestValidateVerificationFields(t *testing.T) {
	verifier := NewGeolocationVerifier()

	tests := []struct {
		verificationType string
		city             string
		country          string
		expectedError    bool
		errorContains    string
	}{
		// Valid cases
		{"none", "", "", false, ""},
		{"country", "", "US", false, ""},
		{"city", "New York", "", false, ""},
		{"city_country", "New York", "US", false, ""},

		// Invalid cases
		{"country", "", "", true, "country is required"},
		{"country", "", "USA", true, "invalid country code"},
		{"city", "", "", true, "city is required"},
		{"city", "N", "", true, "invalid city name"},
		{"city_country", "", "US", true, "city is required"},
		{"city_country", "New York", "", true, "country is required"},
		{"city_country", "New York", "USA", true, "invalid country code"},
		{"invalid", "New York", "US", true, "invalid verification type"},
	}

	for _, test := range tests {
		err := verifier.ValidateVerificationFields(
			VerificationType(test.verificationType),
			test.city,
			test.country,
		)

		if test.expectedError {
			if err == nil {
				t.Errorf("Expected error for %s/%s/%s, got none", test.verificationType, test.city, test.country)
			} else if test.errorContains != "" && !contains(err.Error(), test.errorContains) {
				t.Errorf("Expected error to contain '%s' for %s/%s/%s, got: %v", test.errorContains, test.verificationType, test.city, test.country, err)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for %s/%s/%s, got: %v", test.verificationType, test.city, test.country, err)
			}
		}
	}
}

func TestNormalizeVerificationFields(t *testing.T) {
	verifier := NewGeolocationVerifier()

	tests := []struct {
		verificationType string
		city             string
		country          string
		expectedCity     string
		expectedCountry  string
	}{
		{"none", "New York", "US", "", ""},
		{"country", "New York", "US", "", "us"},
		{"country", "New York", "  US  ", "", "us"},
		{"city", "New York", "US", "new york", ""},
		{"city", "  NEW YORK  ", "US", "new york", ""},
		{"city_country", "New York", "US", "new york", "us"},
		{"city_country", "  NEW YORK  ", "  US  ", "new york", "us"},
	}

	for _, test := range tests {
		normalizedCity, normalizedCountry := verifier.NormalizeVerificationFields(
			VerificationType(test.verificationType),
			test.city,
			test.country,
		)

		if normalizedCity != test.expectedCity {
			t.Errorf("Expected normalized city '%s' for %s/%s/%s, got: %s", test.expectedCity, test.verificationType, test.city, test.country, normalizedCity)
		}
		if normalizedCountry != test.expectedCountry {
			t.Errorf("Expected normalized country '%s' for %s/%s/%s, got: %s", test.expectedCountry, test.verificationType, test.city, test.country, normalizedCountry)
		}
	}
}

func TestGetVerificationDescription(t *testing.T) {
	verifier := NewGeolocationVerifier()

	tests := []struct {
		verificationType string
		city             string
		country          string
		expected         string
	}{
		{"none", "", "", "No geolocation verification required"},
		{"country", "", "US", "Access restricted to country: US"},
		{"city", "New York", "", "Access restricted to city: New York"},
		{"city_country", "New York", "US", "Access restricted to city: New York, country: US"},
		{"invalid", "New York", "US", "Unknown verification type"},
	}

	for _, test := range tests {
		description := verifier.GetVerificationDescription(
			VerificationType(test.verificationType),
			test.city,
			test.country,
		)

		if description != test.expected {
			t.Errorf("Expected description '%s' for %s/%s/%s, got: %s", test.expected, test.verificationType, test.city, test.country, description)
		}
	}
}

func TestVerifyLocation_CaseInsensitive(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test case-insensitive matching
	result := verifier.VerifyLocation(
		VerificationTypeCityCountry,
		&geolocation.Location{Country: "US", City: "NEW YORK"},
		"new york",
		"us",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for case-insensitive matching, got: %v", result)
	}
}

func TestVerifyLocation_WhitespaceHandling(t *testing.T) {
	verifier := NewGeolocationVerifier()

	// Test whitespace handling
	result := verifier.VerifyLocation(
		VerificationTypeCity,
		&geolocation.Location{Country: "us", City: "  New York  "},
		"new york",
		"",
	)

	if !result.Allowed {
		t.Errorf("Expected access to be allowed for whitespace-normalized matching, got: %v", result)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
