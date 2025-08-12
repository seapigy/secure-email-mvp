package georestriction

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"secure-email-mvp/pkg/geolocation"
)

// RestrictionType defines the type of geo-restriction
type RestrictionType string

const (
	RestrictionTypeAllow RestrictionType = "allow"
	RestrictionTypeBlock RestrictionType = "block"
)

// GeoRestrictionRule represents a single geo-restriction rule
type GeoRestrictionRule struct {
	ID          string         `json:"id"`
	EmailID     string         `json:"email_id"`
	Type        RestrictionType `json:"type"`
	Countries   []string       `json:"countries,omitempty"`
	Cities      []string       `json:"cities,omitempty"`
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// GeoRestrictionConfig represents the configuration for geo-restrictions
type GeoRestrictionConfig struct {
	Enabled           bool          `json:"enabled"`
	DefaultAction     RestrictionType `json:"default_action"` // "allow" or "block"
	StrictMode        bool          `json:"strict_mode"`      // If true, both country and city must match
	LogViolations     bool          `json:"log_violations"`
	BlockOnGeolocationFailure bool   `json:"block_on_geolocation_failure"`
}

// GeoRestrictionService provides methods for managing and enforcing geo-restrictions
type GeoRestrictionService struct {
	config GeoRestrictionConfig
}

// NewGeoRestrictionService creates a new geo-restriction service with default configuration
func NewGeoRestrictionService() *GeoRestrictionService {
	return &GeoRestrictionService{
		config: GeoRestrictionConfig{
			Enabled:                    true,
			DefaultAction:              RestrictionTypeAllow,
			StrictMode:                 false,
			LogViolations:              true,
			BlockOnGeolocationFailure:  true,
		},
	}
}

// NewGeoRestrictionServiceWithConfig creates a new geo-restriction service with custom configuration
func NewGeoRestrictionServiceWithConfig(config GeoRestrictionConfig) *GeoRestrictionService {
	return &GeoRestrictionService{
		config: config,
	}
}

// CheckAccess determines if access should be allowed based on geo-restriction rules
func (grs *GeoRestrictionService) CheckAccess(
	clientLocation *geolocation.Location,
	rules []GeoRestrictionRule,
) (bool, string, error) {
	if !grs.config.Enabled {
		return true, "", nil
	}

	if clientLocation == nil {
		if grs.config.BlockOnGeolocationFailure {
			return false, "Access denied: Unable to determine your location", nil
		}
		return true, "", nil
	}

	// If no rules are defined, use default action
	if len(rules) == 0 {
		if grs.config.DefaultAction == RestrictionTypeAllow {
			return true, "", nil
		} else {
			return false, "Access blocked: No geo-restriction rules configured", nil
		}
	}

	// Check each rule in order
	hasAllowRules := false
	hasBlockRules := false
	allowRuleMatched := false
	blockRuleMatched := false

	for _, rule := range rules {
		matches, reason := grs.checkRule(clientLocation, rule)
		
		if rule.Type == RestrictionTypeAllow {
			hasAllowRules = true
			if matches {
				allowRuleMatched = true
			}
		}
		
		if rule.Type == RestrictionTypeBlock {
			hasBlockRules = true
			if matches {
				blockRuleMatched = true
				return false, fmt.Sprintf("Access blocked: %s", reason), nil
			}
		}
	}

	// If we have allow rules and one matched, allow access
	if hasAllowRules && allowRuleMatched {
		return true, "", nil
	}

	// If we have allow rules but none matched, deny access
	if hasAllowRules && !allowRuleMatched {
		return false, "Access blocked: Your location is not in the allowed areas", nil
	}

	// If we only have block rules and none matched, allow access
	if hasBlockRules && !blockRuleMatched {
		return true, "", nil
	}

	// Default to deny if no rules matched
	return false, "Access blocked: No matching geo-restriction rules", nil
}

// checkRule checks if a client location matches a specific rule
func (grs *GeoRestrictionService) checkRule(
	clientLocation *geolocation.Location,
	rule GeoRestrictionRule,
) (bool, string) {
	countryMatch := false
	cityMatch := false

	// Check country restrictions
	if len(rule.Countries) > 0 {
		clientCountry := strings.ToLower(strings.TrimSpace(clientLocation.Country))
		for _, allowedCountry := range rule.Countries {
			if strings.ToLower(strings.TrimSpace(allowedCountry)) == clientCountry {
				countryMatch = true
				break
			}
		}
	} else {
		// No country restrictions means country always matches
		countryMatch = true
	}

	// Check city restrictions
	if len(rule.Cities) > 0 {
		clientCity := geolocation.NormalizeCityName(clientLocation.City)
		for _, allowedCity := range rule.Cities {
			if geolocation.NormalizeCityName(allowedCity) == clientCity {
				cityMatch = true
				break
			}
		}
	} else {
		// No city restrictions means city always matches
		cityMatch = true
	}

	// Determine if rule matches based on strict mode
	if grs.config.StrictMode {
		// Both country and city must match
		if countryMatch && cityMatch {
			return true, ""
		}
		return false, fmt.Sprintf("Location (%s, %s) does not match strict requirements", 
			clientLocation.City, strings.ToUpper(clientLocation.Country))
	} else {
		// Either country or city can match
		if countryMatch || cityMatch {
			return true, ""
		}
		return false, fmt.Sprintf("Location (%s, %s) is not in allowed areas", 
			clientLocation.City, strings.ToUpper(clientLocation.Country))
	}
}

// ValidateRule validates a geo-restriction rule
func (grs *GeoRestrictionService) ValidateRule(rule GeoRestrictionRule) error {
	// Validate rule type
	if rule.Type != RestrictionTypeAllow && rule.Type != RestrictionTypeBlock {
		return fmt.Errorf("invalid restriction type: %s", rule.Type)
	}

	// Validate that at least one restriction is specified
	if len(rule.Countries) == 0 && len(rule.Cities) == 0 {
		return fmt.Errorf("at least one country or city must be specified")
	}

	// Validate countries
	for i, country := range rule.Countries {
		if !geolocation.ValidateCountryCode(country) {
			return fmt.Errorf("invalid country code at index %d: %s", i, country)
		}
	}

	// Validate cities
	for i, city := range rule.Cities {
		if !geolocation.ValidateCityName(city) {
			return fmt.Errorf("invalid city name at index %d: %s", i, city)
		}
	}

	return nil
}

// NormalizeRule normalizes a geo-restriction rule (converts to lowercase, trims whitespace)
func (grs *GeoRestrictionService) NormalizeRule(rule GeoRestrictionRule) GeoRestrictionRule {
	normalized := rule

	// Normalize countries
	normalized.Countries = make([]string, len(rule.Countries))
	for i, country := range rule.Countries {
		normalized.Countries[i] = strings.ToLower(strings.TrimSpace(country))
	}

	// Normalize cities
	normalized.Cities = make([]string, len(rule.Cities))
	for i, city := range rule.Cities {
		normalized.Cities[i] = geolocation.NormalizeCityName(city)
	}

	return normalized
}

// GetDefaultConfig returns the default configuration for geo-restrictions
func (grs *GeoRestrictionService) GetDefaultConfig() GeoRestrictionConfig {
	return GeoRestrictionConfig{
		Enabled:                    true,
		DefaultAction:              RestrictionTypeAllow,
		StrictMode:                 false,
		LogViolations:              true,
		BlockOnGeolocationFailure:  true,
	}
}

// GetConfig returns the current configuration
func (grs *GeoRestrictionService) GetConfig() GeoRestrictionConfig {
	return grs.config
}

// SetConfig updates the configuration
func (grs *GeoRestrictionService) SetConfig(config GeoRestrictionConfig) {
	grs.config = config
}

// ParseRulesFromJSON parses geo-restriction rules from JSON
func (grs *GeoRestrictionService) ParseRulesFromJSON(jsonData []byte) ([]GeoRestrictionRule, error) {
	var rules []GeoRestrictionRule
	if err := json.Unmarshal(jsonData, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse geo-restriction rules: %w", err)
	}

	// Validate each rule
	for i, rule := range rules {
		if err := grs.ValidateRule(rule); err != nil {
			return nil, fmt.Errorf("invalid rule at index %d: %w", i, err)
		}
		rules[i] = grs.NormalizeRule(rule)
	}

	return rules, nil
}

// SerializeRulesToJSON serializes geo-restriction rules to JSON
func (grs *GeoRestrictionService) SerializeRulesToJSON(rules []GeoRestrictionRule) ([]byte, error) {
	return json.Marshal(rules)
}
