package georestriction

import (
	"encoding/json"
	"testing"
	"time"

	"secure-email-mvp/pkg/geolocation"
)

func TestNewGeoRestrictionService(t *testing.T) {
	service := NewGeoRestrictionService()
	if service == nil {
		t.Fatal("NewGeoRestrictionService returned nil")
	}

	config := service.GetConfig()
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}
	if config.DefaultAction != RestrictionTypeAllow {
		t.Error("Default action should be allow")
	}
	if config.StrictMode {
		t.Error("Default strict mode should be false")
	}
}

func TestNewGeoRestrictionServiceWithConfig(t *testing.T) {
	customConfig := GeoRestrictionConfig{
		Enabled:                   false,
		DefaultAction:             RestrictionTypeBlock,
		StrictMode:                true,
		LogViolations:             false,
		BlockOnGeolocationFailure: false,
	}

	service := NewGeoRestrictionServiceWithConfig(customConfig)
	if service == nil {
		t.Fatal("NewGeoRestrictionServiceWithConfig returned nil")
	}

	config := service.GetConfig()
	if config.Enabled != customConfig.Enabled {
		t.Error("Config should match custom config")
	}
	if config.DefaultAction != customConfig.DefaultAction {
		t.Error("Default action should match custom config")
	}
	if config.StrictMode != customConfig.StrictMode {
		t.Error("Strict mode should match custom config")
	}
}

func TestCheckAccess_Disabled(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled: false,
	})

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error when disabled: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when service is disabled")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_NoRules_DefaultAllow(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:       true,
		DefaultAction: RestrictionTypeAllow,
	})

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when no rules and default is allow")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_NoRules_DefaultBlock(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:       true,
		DefaultAction: RestrictionTypeBlock,
	})

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when no rules and default is block")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_AllowRule_Match(t *testing.T) {
	service := NewGeoRestrictionService()

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when rule matches")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_AllowRule_NoMatch(t *testing.T) {
	service := NewGeoRestrictionService()

	location := &geolocation.Location{Country: "ca", City: "toronto"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when rule does not match")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_BlockRule_Match(t *testing.T) {
	service := NewGeoRestrictionService()

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeBlock,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when block rule matches")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_BlockRule_NoMatch(t *testing.T) {
	service := NewGeoRestrictionService()

	location := &geolocation.Location{Country: "ca", City: "toronto"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeBlock,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when block rule does not match")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_StrictMode_BothMatch(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:    true,
		StrictMode: true,
	})

	location := &geolocation.Location{Country: "us", City: "new york"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when both country and city match in strict mode")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_StrictMode_OnlyCountryMatch(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:    true,
		StrictMode: true,
	})

	location := &geolocation.Location{Country: "us", City: "los angeles"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when only country matches in strict mode")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_StrictMode_OnlyCityMatch(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:    true,
		StrictMode: true,
	})

	location := &geolocation.Location{Country: "ca", City: "new york"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when only city matches in strict mode")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_NonStrictMode_EitherMatch(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:    true,
		StrictMode: false,
	})

	location := &geolocation.Location{Country: "ca", City: "new york"}
	rules := []GeoRestrictionRule{
		{
			ID:        "rule1",
			EmailID:   "email1",
			Type:      RestrictionTypeAllow,
			Countries: []string{"us"},
			Cities:    []string{"new york"},
		},
	}

	allowed, reason, err := service.CheckAccess(location, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when city matches in non-strict mode")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestCheckAccess_NilLocation_BlockOnFailure(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:                   true,
		BlockOnGeolocationFailure: true,
	})

	rules := []GeoRestrictionRule{}

	allowed, reason, err := service.CheckAccess(nil, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if allowed {
		t.Error("Access should be blocked when location is nil and block on failure is true")
	}
	if reason == "" {
		t.Error("Reason should not be empty when access is blocked")
	}
}

func TestCheckAccess_NilLocation_AllowOnFailure(t *testing.T) {
	service := NewGeoRestrictionServiceWithConfig(GeoRestrictionConfig{
		Enabled:                   true,
		BlockOnGeolocationFailure: false,
	})

	rules := []GeoRestrictionRule{}

	allowed, reason, err := service.CheckAccess(nil, rules)
	if err != nil {
		t.Fatalf("CheckAccess should not return error: %v", err)
	}
	if !allowed {
		t.Error("Access should be allowed when location is nil and block on failure is false")
	}
	if reason != "" {
		t.Error("Reason should be empty when access is allowed")
	}
}

func TestValidateRule_ValidAllowRule(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      RestrictionTypeAllow,
		Countries: []string{"US"},
		Cities:    []string{"new york"},
	}

	err := service.ValidateRule(rule)
	if err != nil {
		t.Fatalf("ValidateRule should not return error for valid rule: %v", err)
	}
}

func TestValidateRule_ValidBlockRule(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      RestrictionTypeBlock,
		Countries: []string{"US"},
		Cities:    []string{"new york"},
	}

	err := service.ValidateRule(rule)
	if err != nil {
		t.Fatalf("ValidateRule should not return error for valid rule: %v", err)
	}
}

func TestValidateRule_InvalidType(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      "invalid",
		Countries: []string{"us"},
		Cities:    []string{"new york"},
	}

	err := service.ValidateRule(rule)
	if err == nil {
		t.Fatal("ValidateRule should return error for invalid type")
	}
}

func TestValidateRule_NoRestrictions(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:      "rule1",
		EmailID: "email1",
		Type:    RestrictionTypeAllow,
	}

	err := service.ValidateRule(rule)
	if err == nil {
		t.Fatal("ValidateRule should return error when no restrictions specified")
	}
}

func TestValidateRule_InvalidCountry(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      RestrictionTypeAllow,
		Countries: []string{"invalid"},
		Cities:    []string{"new york"},
	}

	err := service.ValidateRule(rule)
	if err == nil {
		t.Fatal("ValidateRule should return error for invalid country")
	}
}

func TestValidateRule_InvalidCity(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      RestrictionTypeAllow,
		Countries: []string{"us"},
		Cities:    []string{""}, // Empty city
	}

	err := service.ValidateRule(rule)
	if err == nil {
		t.Fatal("ValidateRule should return error for invalid city")
	}
}

func TestNormalizeRule(t *testing.T) {
	service := NewGeoRestrictionService()

	rule := GeoRestrictionRule{
		ID:        "rule1",
		EmailID:   "email1",
		Type:      RestrictionTypeAllow,
		Countries: []string{"US", " CA "},
		Cities:    []string{" New York ", "Los Angeles"},
	}

	normalized := service.NormalizeRule(rule)

	expectedCountries := []string{"us", "ca"}
	for i, country := range normalized.Countries {
		if country != expectedCountries[i] {
			t.Errorf("Expected country %s, got %s", expectedCountries[i], country)
		}
	}

	expectedCities := []string{"new york", "los angeles"}
	for i, city := range normalized.Cities {
		if city != expectedCities[i] {
			t.Errorf("Expected city %s, got %s", expectedCities[i], city)
		}
	}
}

func TestParseRulesFromJSON_Valid(t *testing.T) {
	service := NewGeoRestrictionService()

	jsonData := `[
		{
			"id": "rule1",
			"email_id": "email1",
			"type": "allow",
			"countries": ["US", "CA"],
			"cities": ["new york", "toronto"],
			"description": "Test rule"
		}
	]`

	rules, err := service.ParseRulesFromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseRulesFromJSON should not return error: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if rule.ID != "rule1" {
		t.Errorf("Expected ID rule1, got %s", rule.ID)
	}
	if rule.Type != RestrictionTypeAllow {
		t.Errorf("Expected type allow, got %s", rule.Type)
	}
	if len(rule.Countries) != 2 {
		t.Errorf("Expected 2 countries, got %d", len(rule.Countries))
	}
	if len(rule.Cities) != 2 {
		t.Errorf("Expected 2 cities, got %d", len(rule.Cities))
	}
}

func TestParseRulesFromJSON_Invalid(t *testing.T) {
	service := NewGeoRestrictionService()

	jsonData := `[
		{
			"id": "rule1",
			"email_id": "email1",
			"type": "invalid",
			"countries": ["invalid"],
			"cities": [""]
		}
	]`

	_, err := service.ParseRulesFromJSON([]byte(jsonData))
	if err == nil {
		t.Fatal("ParseRulesFromJSON should return error for invalid rules")
	}
}

func TestSerializeRulesToJSON(t *testing.T) {
	service := NewGeoRestrictionService()

	rules := []GeoRestrictionRule{
		{
			ID:          "rule1",
			EmailID:     "email1",
			Type:        RestrictionTypeAllow,
			Countries:   []string{"US"},
			Cities:      []string{"new york"},
			Description: "Test rule",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	jsonData, err := service.SerializeRulesToJSON(rules)
	if err != nil {
		t.Fatalf("SerializeRulesToJSON should not return error: %v", err)
	}

	// Verify the JSON can be parsed back
	var parsedRules []GeoRestrictionRule
	if err := json.Unmarshal(jsonData, &parsedRules); err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	if len(parsedRules) != 1 {
		t.Fatalf("Expected 1 rule after parsing, got %d", len(parsedRules))
	}

	rule := parsedRules[0]
	if rule.ID != "rule1" {
		t.Errorf("Expected ID rule1, got %s", rule.ID)
	}
	if rule.Type != RestrictionTypeAllow {
		t.Errorf("Expected type allow, got %s", rule.Type)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	service := NewGeoRestrictionService()

	config := service.GetDefaultConfig()
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}
	if config.DefaultAction != RestrictionTypeAllow {
		t.Error("Default action should be allow")
	}
	if config.StrictMode {
		t.Error("Default strict mode should be false")
	}
	if !config.LogViolations {
		t.Error("Default log violations should be true")
	}
	if !config.BlockOnGeolocationFailure {
		t.Error("Default block on geolocation failure should be true")
	}
}
