package reputation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewReputationService(t *testing.T) {
	service := NewReputationService()
	if service == nil {
		t.Fatal("NewReputationService returned nil")
	}
	if service.config == nil {
		t.Fatal("ReputationService config is nil")
	}
	if service.client == nil {
		t.Fatal("ReputationService client is nil")
	}
}

func TestNewReputationServiceWithConfig(t *testing.T) {
	config := &ReputationConfig{
		APIKey:    "test-key",
		Threshold: 50,
		Timeout:   5 * time.Second,
		BaseURL:   "https://test.api.com",
		UserAgent: "TestAgent/1.0",
	}

	service := NewReputationServiceWithConfig(config)
	if service == nil {
		t.Fatal("NewReputationServiceWithConfig returned nil")
	}
	if service.config != config {
		t.Fatal("ReputationService config does not match provided config")
	}
}

func TestIsValidIP(t *testing.T) {
	service := NewReputationService()

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid public ipv4", "8.8.8.8", true},
		{"valid public ipv6", "2001:4860:4860::8888", true},
		{"private ipv4", "192.168.1.1", false},
		{"loopback", "127.0.0.1", false},
		{"link local", "169.254.1.1", false},
		{"invalid ip", "invalid", false},
		{"empty string", "", false},
		{"ip with port", "8.8.8.8:8080", true},
		{"ipv6 with port", "[2001:4860:4860::8888]:8080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isValidIP(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteAddr string
		expected string
	}{
		{
			name: "x-forwarded-for",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 10.0.0.1",
			},
			expected: "203.0.113.1",
		},
		{
			name: "x-real-ip",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			expected: "203.0.113.2",
		},
		{
			name: "cf-connecting-ip",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.3",
			},
			expected: "203.0.113.3",
		},
		{
			name: "remote addr",
			remoteAddr: "203.0.113.4:12345",
			expected: "203.0.113.4",
		},
		{
			name: "remote addr without port",
			remoteAddr: "203.0.113.5",
			expected: "203.0.113.5",
		},
		{
			name: "invalid headers fallback",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
			},
			remoteAddr: "203.0.113.6:12345",
			expected: "203.0.113.6",
		},
		{
			name: "no valid ip",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
			},
			remoteAddr: "invalid-addr",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			result := GetClientIP(req)
			if result != tt.expected {
				t.Errorf("GetClientIP() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestCheckIPReputation_AllowedIP(t *testing.T) {
	// Create a test server that returns a clean IP response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"ipAddress": "8.8.8.8",
				"countryCode": "US",
				"usageType": "Residential",
				"isp": "Google LLC",
				"domain": "google.com",
				"hostnames": "",
				"totalReports": 0,
				"numDistinctUsers": 0,
				"lastReportedAt": null,
				"abuseConfidenceScore": 0
			}
		}`))
	}))
	defer server.Close()

	config := &ReputationConfig{
		APIKey:    "test-key",
		Threshold: 25,
		Timeout:   5 * time.Second,
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
	}

	service := NewReputationServiceWithConfig(config)
	ctx := context.Background()

	isMalicious, err := service.CheckIPReputation(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("CheckIPReputation failed: %v", err)
	}

	if isMalicious {
		t.Error("Expected IP to be allowed, but it was flagged as malicious")
	}
}

func TestCheckIPReputation_BlockedIP(t *testing.T) {
	// Create a test server that returns a malicious IP response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"ipAddress": "203.0.113.1",
				"countryCode": "XX",
				"usageType": "Hosting",
				"isp": "Malicious ISP",
				"domain": "malicious.com",
				"hostnames": "",
				"totalReports": 150,
				"numDistinctUsers": 45,
				"lastReportedAt": "2024-01-15T10:30:00Z",
				"abuseConfidenceScore": 85
			}
		}`))
	}))
	defer server.Close()

	config := &ReputationConfig{
		APIKey:    "test-key",
		Threshold: 25,
		Timeout:   5 * time.Second,
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
	}

	service := NewReputationServiceWithConfig(config)
	ctx := context.Background()

	isMalicious, err := service.CheckIPReputation(ctx, "203.0.113.1")
	if err != nil {
		t.Fatalf("CheckIPReputation failed: %v", err)
	}

	if !isMalicious {
		t.Error("Expected IP to be blocked, but it was allowed")
	}
}

func TestCheckIPReputation_APIFailureFallback(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	config := &ReputationConfig{
		APIKey:    "test-key",
		Threshold: 25,
		Timeout:   5 * time.Second,
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
	}

	service := NewReputationServiceWithConfig(config)
	ctx := context.Background()

	isMalicious, err := service.CheckIPReputation(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("CheckIPReputation failed: %v", err)
	}

	// Should allow access on API failure
	if isMalicious {
		t.Error("Expected IP to be allowed on API failure, but it was blocked")
	}
}

func TestCheckIPReputation_TimeoutFallback(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &ReputationConfig{
		APIKey:    "test-key",
		Threshold: 25,
		Timeout:   1 * time.Second, // Short timeout
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
	}

	service := NewReputationServiceWithConfig(config)
	ctx := context.Background()

	isMalicious, err := service.CheckIPReputation(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("CheckIPReputation failed: %v", err)
	}

	// Should allow access on timeout
	if isMalicious {
		t.Error("Expected IP to be allowed on timeout, but it was blocked")
	}
}

func TestCheckIPReputation_NoAPIKey(t *testing.T) {
	// Clear API key environment variable
	originalKey := os.Getenv("IP_REPUTATION_API_KEY")
	os.Unsetenv("IP_REPUTATION_API_KEY")
	defer os.Setenv("IP_REPUTATION_API_KEY", originalKey)

	service := NewReputationService()
	ctx := context.Background()

	isMalicious, err := service.CheckIPReputation(ctx, "8.8.8.8")
	if err != nil {
		t.Fatalf("CheckIPReputation failed: %v", err)
	}

	// Should allow access when no API key is configured
	if isMalicious {
		t.Error("Expected IP to be allowed when no API key configured, but it was blocked")
	}
}

func TestCheckIPReputation_InvalidIP(t *testing.T) {
	service := NewReputationService()
	ctx := context.Background()

	_, err := service.CheckIPReputation(ctx, "invalid-ip")
	if err == nil {
		t.Error("Expected error for invalid IP, but got none")
	}

	_, err = service.CheckIPReputation(ctx, "192.168.1.1")
	if err == nil {
		t.Error("Expected error for private IP, but got none")
	}
}

func TestGetIntEnv(t *testing.T) {
	// Test with valid environment variable
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getIntEnv("TEST_INT", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with invalid environment variable
	result = getIntEnv("TEST_INVALID", 10)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}

	// Test with non-existent environment variable
	result = getIntEnv("TEST_MISSING", 10)
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}
}

