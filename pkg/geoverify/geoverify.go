package geoverify

import (
	"fmt"
	"strings"

	"secure-email-mvp/pkg/geolocation"
)

// VerificationType represents the type of geolocation verification required
type VerificationType string

const (
	VerificationTypeNone        VerificationType = "none"
	VerificationTypeCountry     VerificationType = "country"
	VerificationTypeCity        VerificationType = "city"
	VerificationTypeCityCountry VerificationType = "city_country"
)

// VerificationResult represents the result of a geolocation verification check
type VerificationResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// GeolocationVerifier provides methods for verifying geolocation-based access
type GeolocationVerifier struct{}

// NewGeolocationVerifier creates a new geolocation verifier instance
func NewGeolocationVerifier() *GeolocationVerifier {
	return &GeolocationVerifier{}
}

// VerifyLocation checks if a location passes the specified verification requirements
func (gv *GeolocationVerifier) VerifyLocation(
	verificationType VerificationType,
	clientLocation *geolocation.Location,
	requiredCity string,
	requiredCountry string,
) *VerificationResult {
	// If no verification is required, allow access
	if verificationType == VerificationTypeNone {
		return &VerificationResult{Allowed: true}
	}

	// Validate inputs
	if clientLocation == nil {
		return &VerificationResult{
			Allowed: false,
			Reason:  "Unable to determine your location",
		}
	}

	// Check country verification if required
	if verificationType == VerificationTypeCountry || verificationType == VerificationTypeCityCountry {
		if requiredCountry == "" {
			return &VerificationResult{
				Allowed: false,
				Reason:  "Country verification required but no country specified",
			}
		}

		normalizedClientCountry := strings.ToLower(strings.TrimSpace(clientLocation.Country))
		normalizedRequiredCountry := strings.ToLower(strings.TrimSpace(requiredCountry))

		if normalizedClientCountry != normalizedRequiredCountry {
			return &VerificationResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Country verification failed: expected %s, got %s", strings.ToUpper(requiredCountry), strings.ToUpper(clientLocation.Country)),
			}
		}
	}

	// Check city verification if required
	if verificationType == VerificationTypeCity || verificationType == VerificationTypeCityCountry {
		if requiredCity == "" {
			return &VerificationResult{
				Allowed: false,
				Reason:  "City verification required but no city specified",
			}
		}

		normalizedClientCity := geolocation.NormalizeCityName(clientLocation.City)
		normalizedRequiredCity := geolocation.NormalizeCityName(requiredCity)

		if normalizedClientCity != normalizedRequiredCity {
			return &VerificationResult{
				Allowed: false,
				Reason:  fmt.Sprintf("City verification failed: expected %s, got %s", requiredCity, clientLocation.City),
			}
		}
	}

	// All checks passed
	return &VerificationResult{Allowed: true}
}

// ValidateVerificationType validates if a verification type is valid
func (gv *GeolocationVerifier) ValidateVerificationType(verificationType string) error {
	switch VerificationType(verificationType) {
	case VerificationTypeNone, VerificationTypeCountry, VerificationTypeCity, VerificationTypeCityCountry:
		return nil
	default:
		return fmt.Errorf("invalid verification type: %s. Must be 'none', 'country', 'city', or 'city_country'", verificationType)
	}
}

// ValidateVerificationFields validates the verification fields based on the verification type
func (gv *GeolocationVerifier) ValidateVerificationFields(
	verificationType VerificationType,
	city string,
	country string,
) error {
	switch verificationType {
	case VerificationTypeNone:
		// No validation needed
		return nil

	case VerificationTypeCountry:
		// Country is required, city is ignored
		if country == "" {
			return fmt.Errorf("country is required when verification type is 'country'")
		}
		if !geolocation.ValidateCountryCode(country) {
			return fmt.Errorf("invalid country code: %s. Must be ISO 3166-1 alpha-2 format (e.g., US, CA, GB)", country)
		}
		return nil

	case VerificationTypeCity:
		// City is required, country is ignored
		if city == "" {
			return fmt.Errorf("city is required when verification type is 'city'")
		}
		if !geolocation.ValidateCityName(city) {
			return fmt.Errorf("invalid city name: %s. Must be 2-100 characters, letters, spaces, hyphens, and apostrophes only", city)
		}
		return nil

	case VerificationTypeCityCountry:
		// Both city and country are required
		if city == "" {
			return fmt.Errorf("city is required when verification type is 'city_country'")
		}
		if country == "" {
			return fmt.Errorf("country is required when verification type is 'city_country'")
		}
		if !geolocation.ValidateCityName(city) {
			return fmt.Errorf("invalid city name: %s. Must be 2-100 characters, letters, spaces, hyphens, and apostrophes only", city)
		}
		if !geolocation.ValidateCountryCode(country) {
			return fmt.Errorf("invalid country code: %s. Must be ISO 3166-1 alpha-2 format (e.g., US, CA, GB)", country)
		}
		return nil

	default:
		return fmt.Errorf("invalid verification type: %s", verificationType)
	}
}

// NormalizeVerificationFields normalizes the verification fields for storage
func (gv *GeolocationVerifier) NormalizeVerificationFields(
	verificationType VerificationType,
	city string,
	country string,
) (string, string) {
	var normalizedCity, normalizedCountry string

	switch verificationType {
	case VerificationTypeCountry:
		normalizedCity = ""
		normalizedCountry = strings.ToLower(strings.TrimSpace(country))

	case VerificationTypeCity:
		normalizedCity = geolocation.NormalizeCityName(city)
		normalizedCountry = ""

	case VerificationTypeCityCountry:
		normalizedCity = geolocation.NormalizeCityName(city)
		normalizedCountry = strings.ToLower(strings.TrimSpace(country))

	default:
		normalizedCity = ""
		normalizedCountry = ""
	}

	return normalizedCity, normalizedCountry
}

// GetVerificationDescription returns a human-readable description of the verification requirements
func (gv *GeolocationVerifier) GetVerificationDescription(
	verificationType VerificationType,
	city string,
	country string,
) string {
	switch verificationType {
	case VerificationTypeNone:
		return "No geolocation verification required"

	case VerificationTypeCountry:
		return fmt.Sprintf("Access restricted to country: %s", strings.ToUpper(country))

	case VerificationTypeCity:
		return fmt.Sprintf("Access restricted to city: %s", city)

	case VerificationTypeCityCountry:
		return fmt.Sprintf("Access restricted to city: %s, country: %s", city, strings.ToUpper(country))

	default:
		return "Unknown verification type"
	}
}
