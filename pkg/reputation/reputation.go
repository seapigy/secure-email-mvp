package reputation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ReputationConfig holds configuration for IP reputation checking
type ReputationConfig struct {
	APIKey     string        // AbuseIPDB API key
	Threshold  int           // Malicious threshold (0-100)
	Timeout    time.Duration // API request timeout
	BaseURL    string        // API base URL
	UserAgent  string        // User agent for API requests
}

// DefaultConfig returns the default reputation configuration
func DefaultConfig() *ReputationConfig {
	return &ReputationConfig{
		APIKey:    os.Getenv("IP_REPUTATION_API_KEY"),
		Threshold: getIntEnv("IP_REPUTATION_THRESHOLD", 25), // Default 25% threshold
		Timeout:   10 * time.Second,
		BaseURL:   "https://api.abuseipdb.com/api/v2",
		UserAgent: "SecureEmailMVP/1.0",
	}
}

// AbuseIPDBResponse represents the response from AbuseIPDB API
type AbuseIPDBResponse struct {
	Data struct {
		IPAddress     string  `json:"ipAddress"`
		CountryCode   string  `json:"countryCode"`
		UsageType     string  `json:"usageType"`
		ISP           string  `json:"isp"`
		Domain        string  `json:"domain"`
		Hostnames     string  `json:"hostnames"`
		TotalReports  int     `json:"totalReports"`
		NumDistinctUsers int  `json:"numDistinctUsers"`
		LastReportedAt string `json:"lastReportedAt"`
		AbuseConfidenceScore int `json:"abuseConfidenceScore"`
	} `json:"data"`
}

// ReputationService provides methods for checking IP reputation
type ReputationService struct {
	config *ReputationConfig
	client *http.Client
}

// NewReputationService creates a new reputation service
func NewReputationService() *ReputationService {
	config := DefaultConfig()
	
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	return &ReputationService{
		config: config,
		client: client,
	}
}

// NewReputationServiceWithConfig creates a new reputation service with custom configuration
func NewReputationServiceWithConfig(config *ReputationConfig) *ReputationService {
	client := &http.Client{
		Timeout: config.Timeout,
	}
	
	return &ReputationService{
		config: config,
		client: client,
	}
}

// CheckIPReputation checks the reputation of an IP address using AbuseIPDB
// Returns true if IP is malicious (above threshold), false if clean or API failure
func (s *ReputationService) CheckIPReputation(ctx context.Context, ipAddress string) (bool, error) {
	// Validate and sanitize IP address
	if !s.isValidIP(ipAddress) {
		return false, fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	// Check if API key is configured
	if s.config.APIKey == "" {
		log.Printf("IP reputation API key not configured, allowing access for IP: %s", ipAddress)
		return false, nil
	}

	// Create API request
	url := fmt.Sprintf("%s/check", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("ipAddress", ipAddress)
	q.Add("maxAgeInDays", "90") // Check last 90 days
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.config.UserAgent)

	// Make request
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("IP reputation API request failed for IP %s: %v", ipAddress, err)
		return false, nil // Allow access on API failure
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read IP reputation response for IP %s: %v", ipAddress, err)
		return false, nil // Allow access on read failure
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		log.Printf("IP reputation API returned status %d for IP %s: %s", resp.StatusCode, ipAddress, string(body))
		return false, nil // Allow access on API error
	}

	// Parse response
	var apiResp AbuseIPDBResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Printf("Failed to parse IP reputation response for IP %s: %v", ipAddress, err)
		return false, nil // Allow access on parse failure
	}

	// Check abuse confidence score
	score := apiResp.Data.AbuseConfidenceScore
	isMalicious := score >= s.config.Threshold

	if isMalicious {
		log.Printf("IP %s flagged as malicious (score: %d%%, threshold: %d%%)", ipAddress, score, s.config.Threshold)
	} else {
		log.Printf("IP %s reputation check passed (score: %d%%, threshold: %d%%)", ipAddress, score, s.config.Threshold)
	}

	return isMalicious, nil
}

// isValidIP validates and sanitizes an IP address
func (s *ReputationService) isValidIP(ipAddress string) bool {
	// Handle IPv6 addresses with brackets and ports
	if strings.HasPrefix(ipAddress, "[") && strings.Contains(ipAddress, "]:") {
		// Extract IP from [ip]:port format
		endBracket := strings.Index(ipAddress, "]")
		if endBracket > 0 {
			ipAddress = ipAddress[1:endBracket]
		}
	} else if strings.Contains(ipAddress, ":") && !strings.Contains(ipAddress, "::") {
		// Handle IPv4 with port (simple case)
		ipAddress = strings.Split(ipAddress, ":")[0]
	}

	// Parse IP address
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	// Check if it's a private IP (we might want to skip these)
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	return true
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetClientIP extracts the real client IP from HTTP request headers
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For for proxies
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		if net.ParseIP(cfip) != nil {
			return cfip
		}
	}

	// Fallback to RemoteAddr
	if r.RemoteAddr != "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && net.ParseIP(ip) != nil {
			return ip
		}
		// If SplitHostPort fails, try parsing as IP
		if net.ParseIP(r.RemoteAddr) != nil {
			return r.RemoteAddr
		}
	}

	return "unknown"
}
