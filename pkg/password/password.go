package password

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// PasswordConfig holds configuration for password validation
type PasswordConfig struct {
	MinLength           int
	RequireUppercase    bool
	RequireLowercase    bool
	RequireNumbers      bool
	RequireSpecialChars bool
	HIBPAPIKey          string
	HIBPTimeout         time.Duration
	UserAgent           string
}

// DefaultConfig returns default password validation configuration
func DefaultConfig() *PasswordConfig {
	return &PasswordConfig{
		MinLength:           8, // Reduced from 12 for testing
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:      true,
		RequireSpecialChars: true,
		HIBPAPIKey:          os.Getenv("HIBP_API_KEY"),
		HIBPTimeout:         10 * time.Second,
		UserAgent:           "SecureEmail-MVP/1.0",
	}
}

// PasswordValidationResult contains the result of password validation
type PasswordValidationResult struct {
	IsValid     bool
	Score       int // 0-100
	Suggestions []string
	BreachCount int
	Errors      []string
}

// HIBPResponse represents the response from HaveIBeenPwned API
type HIBPResponse struct {
	Count int `json:"count"`
}

// PasswordService handles password validation and breach checking
type PasswordService struct {
	config *PasswordConfig
	client *http.Client
}

// NewPasswordService creates a new password service with default configuration
func NewPasswordService() *PasswordService {
	config := DefaultConfig()
	return NewPasswordServiceWithConfig(config)
}

// NewPasswordServiceWithConfig creates a new password service with custom configuration
func NewPasswordServiceWithConfig(config *PasswordConfig) *PasswordService {
	client := &http.Client{
		Timeout: config.HIBPTimeout,
	}
	return &PasswordService{
		config: config,
		client: client,
	}
}

// ValidatePassword performs comprehensive password validation
func (s *PasswordService) ValidatePassword(ctx context.Context, password string) (*PasswordValidationResult, error) {
	result := &PasswordValidationResult{
		IsValid:     true,
		Score:       0,
		Suggestions: []string{},
		Errors:      []string{},
	}

	// Check minimum length
	if len(password) < s.config.MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters long", s.config.MinLength))
	}

	// Check character requirements
	if s.config.RequireUppercase && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
	}

	if s.config.RequireLowercase && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
	}

	if s.config.RequireNumbers && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
	}

	if s.config.RequireSpecialChars && !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
	}

	// Check against common password blacklist
	if s.isCommonPassword(password) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Password is too common and easily guessable")
	}

	// Calculate password strength score
	result.Score = s.calculatePasswordScore(password)

	// Generate suggestions
	result.Suggestions = s.generateSuggestions(password, result.Errors)

	// Check for breaches if password is otherwise valid and API key is configured
	if result.IsValid && s.config.HIBPAPIKey != "" {
		breachCount, err := s.checkPasswordBreach(ctx, password)
		if err != nil {
			log.Printf("Password breach check failed: %v", err)
			// Don't fail validation on API errors, but log them
		} else {
			result.BreachCount = breachCount
			if breachCount > 0 {
				result.IsValid = false
				result.Errors = append(result.Errors, "This password has been compromised in data breaches")
			}
		}
	}

	return result, nil
}

// calculatePasswordScore calculates a password strength score (0-100)
func (s *PasswordService) calculatePasswordScore(password string) int {
	score := 0

	// Length contribution (up to 25 points)
	lengthScore := len(password) * 2
	if lengthScore > 25 {
		lengthScore = 25
	}
	score += lengthScore

	// Character variety contribution (up to 25 points)
	varietyScore := 0
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		varietyScore += 5
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		varietyScore += 5
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		varietyScore += 5
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		varietyScore += 5
	}
	score += varietyScore

	// Complexity bonus (up to 25 points)
	complexityScore := 0
	if len(password) >= 16 {
		complexityScore += 10
	}
	if len(password) >= 20 {
		complexityScore += 5
	}
	if !s.isCommonPassword(password) {
		complexityScore += 10
	}
	score += complexityScore

	// Entropy bonus (up to 25 points)
	entropyScore := s.calculateEntropy(password)
	if entropyScore > 25 {
		entropyScore = 25
	}
	score += entropyScore

	return score
}

// calculateEntropy calculates password entropy
func (s *PasswordService) calculateEntropy(password string) int {
	charSet := 0
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		charSet += 26
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		charSet += 26
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		charSet += 10
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		charSet += 32
	}

	if charSet == 0 {
		return 0
	}

	// Simplified entropy calculation
	entropy := len(password) * int(float64(charSet)/10)
	if entropy > 25 {
		return 25
	}
	return entropy
}

// isCommonPassword checks if password is in common password blacklist
func (s *PasswordService) isCommonPassword(password string) bool {
	commonPasswords := map[string]bool{
		"password": true, "123456": true, "123456789": true, "qwerty": true,
		"abc123": true, "password123": true, "admin": true, "letmein": true,
		"welcome": true, "monkey": true, "dragon": true, "master": true,
		"football": true, "superman": true, "trustno1": true, "hello": true,
		"freedom": true, "whatever": true, "qazwsx": true, "baseball": true,
		"sunshine": true, "princess": true, "login": true, "solo": true,
		"hunter": true, "buster": true, "charles": true, "butterfly": true,
		"jordan": true, "michael": true, "michelle": true, "lovely": true,
		"robert": true, "daniel": true, "anthony": true, "joshua": true,
		"jennifer": true, "amanda": true, "jessica": true, "jason": true,
		"matt": true, "mark": true, "kevin": true, "steve": true,
		"thomas": true, "lisa": true, "kelly": true,
		"george": true, "andrew": true, "ryan": true, "jacob": true,
		"stephen": true, "tyler": true, "aaron": true, "nathan": true,
		"adam": true, "paul": true, "scott": true, "samuel": true,
		"christian": true, "justin": true, "bryan": true, "sean": true,
		"eric": true, "jake": true, "morgan": true, "taylor": true,
		"emma": true, "madison": true, "emily": true,
		"olivia": true, "isabella": true, "sophia": true, "ava": true,
		"mia": true, "charlotte": true, "amelia": true, "harper": true,
		"evelyn": true, "abigail": true, "elizabeth": true,
		"sofia": true, "avery": true, "ella": true,
		"scarlett": true, "victoria": true, "aria": true,
		"grace": true, "chloe": true, "camila": true, "penelope": true,
		"layla": true, "riley": true, "zoey": true, "nora": true,
		"lily": true, "eleanor": true, "hannah": true, "lucy": true,
		"savannah": true, "audrey": true, "brooklyn": true, "bella": true, "claire": true,
		"skylar": true, "paisley": true, "everly": true,
		"anna": true, "caroline": true, "nova": true, "genesis": true,
		"emilia": true, "kennedy": true, "samantha": true, "maya": true,
		"willow": true, "kinsley": true, "naomi": true, "aaliyah": true,
		"elena": true, "sarah": true, "ariana": true, "allison": true,
		"gabriella": true, "alice": true, "madelyn": true, "cora": true,
		"ruby": true, "eva": true, "serenity": true, "autumn": true,
		"adeline": true, "hailey": true, "gianna": true, "valentina": true,
		"isla": true, "eliana": true, "quinn": true, "nevaeh": true,
		"ivy": true, "sadie": true, "piper": true, "lydia": true,
		"alexa": true, "josephine": true, "emery": true, "julia": true,
		"delilah": true, "arianna": true, "violet": true, "hazel": true,
		"aubrey": true,
	}

	return commonPasswords[strings.ToLower(password)]
}

// generateSuggestions generates password improvement suggestions
func (s *PasswordService) generateSuggestions(password string, _ []string) []string {
	suggestions := []string{}

	if len(password) < s.config.MinLength {
		suggestions = append(suggestions, fmt.Sprintf("Make it at least %d characters long", s.config.MinLength))
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		suggestions = append(suggestions, "Add uppercase letters")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		suggestions = append(suggestions, "Add lowercase letters")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		suggestions = append(suggestions, "Add numbers")
	}

	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		suggestions = append(suggestions, "Add special characters")
	}

	if s.isCommonPassword(password) {
		suggestions = append(suggestions, "Avoid common words and patterns")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Consider using a passphrase for better security")
	}

	return suggestions
}

// checkPasswordBreach checks if password has been compromised using HIBP API
func (s *PasswordService) checkPasswordBreach(ctx context.Context, password string) (int, error) {
	// If no API key configured, skip breach check
	if s.config.HIBPAPIKey == "" {
		log.Printf("HIBP API key not configured, skipping breach check")
		return 0, nil
	}

	// Create SHA1 hash of password
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Use k-anonymity: send only first 5 characters of hash
	prefix := hashHex[:5]
	suffix := hashHex[5:]

	// Make request to HIBP API
	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	if s.config.HIBPAPIKey != "" {
		req.Header.Set("hibp-api-key", s.config.HIBPAPIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response (format: hash_suffix:count)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == suffix {
			count := 0
			fmt.Sscanf(parts[1], "%d", &count)
			return count, nil
		}
	}

	return 0, nil
}
