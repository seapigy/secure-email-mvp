package humanverification

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// VerificationType represents the type of human verification
type VerificationType string

const (
	VerificationTypeCAPTCHA     VerificationType = "captcha"
	VerificationTypeProofOfWork VerificationType = "proof_of_work"
)

// Challenge represents a proof-of-work challenge
type Challenge struct {
	ID       string `json:"id"`
	Prefix   string `json:"prefix"`
	Target   string `json:"target"`
	MaxNonce int64  `json:"max_nonce"`
}

// VerificationResult represents the result of a verification attempt
type VerificationResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// VerificationLog represents a verification log entry
type VerificationLog struct {
	ID               string           `json:"id"`
	EmailID          string           `json:"email_id"`
	IPAddress        string           `json:"ip_address"`
	UserAgent        string           `json:"user_agent"`
	VerificationType VerificationType `json:"verification_type"`
	ChallengeID      string           `json:"challenge_id"`
	Result           string           `json:"result"`
	Timestamp        time.Time        `json:"timestamp"`
	Details          string           `json:"details"`
}

// HumanVerificationService defines the interface for human verification
type HumanVerificationService interface {
	// VerifyResponse verifies a human verification response
	VerifyResponse(ctx context.Context, emailID, token string, verificationType VerificationType) (bool, error)

	// GenerateChallenge generates a proof-of-work challenge
	GenerateChallenge(ctx context.Context) (*Challenge, error)

	// LogVerification logs a verification attempt
	LogVerification(ctx context.Context, log *VerificationLog) error

	// GetVerificationStats gets verification statistics for abuse detection
	GetVerificationStats(ctx context.Context, emailID, ipAddress string, duration time.Duration) (*VerificationStats, error)
}

// VerificationStats represents verification statistics
type VerificationStats struct {
	TotalAttempts   int     `json:"total_attempts"`
	FailedAttempts  int     `json:"failed_attempts"`
	SuccessAttempts int     `json:"success_attempts"`
	FailureRate     float64 `json:"failure_rate"`
}

// Config represents the configuration for human verification
type Config struct {
	Enabled               bool          `json:"enabled"`
	VerificationType      string        `json:"verification_type"` // "captcha" or "proof_of_work"
	CAPTCHASecretKey      string        `json:"captcha_secret_key"`
	CAPTCHASiteKey        string        `json:"captcha_site_key"`
	CAPTCHAEndpoint       string        `json:"captcha_endpoint"`
	ProofOfWorkDifficulty int           `json:"proof_of_work_difficulty"`
	MaxNonce              int64         `json:"max_nonce"`
	FailureThreshold      int           `json:"failure_threshold"`
	BanDuration           time.Duration `json:"ban_duration"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:               true,
		VerificationType:      "proof_of_work",
		CAPTCHAEndpoint:       "https://www.google.com/recaptcha/api/siteverify",
		ProofOfWorkDifficulty: 4,
		MaxNonce:              1000000,
		FailureThreshold:      5,
		BanDuration:           15 * time.Minute,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	if enabled := os.Getenv("HUMAN_VERIFICATION_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true"
	}

	if verificationType := os.Getenv("HUMAN_VERIFICATION_TYPE"); verificationType != "" {
		config.VerificationType = verificationType
	}

	if secretKey := os.Getenv("CAPTCHA_SECRET_KEY"); secretKey != "" {
		config.CAPTCHASecretKey = secretKey
	}

	if siteKey := os.Getenv("CAPTCHA_SITE_KEY"); siteKey != "" {
		config.CAPTCHASiteKey = siteKey
	}

	if endpoint := os.Getenv("CAPTCHA_ENDPOINT"); endpoint != "" {
		config.CAPTCHAEndpoint = endpoint
	}

	if difficulty := os.Getenv("PROOF_OF_WORK_DIFFICULTY"); difficulty != "" {
		if d, err := strconv.Atoi(difficulty); err == nil {
			config.ProofOfWorkDifficulty = d
		}
	}

	if maxNonce := os.Getenv("PROOF_OF_WORK_MAX_NONCE"); maxNonce != "" {
		if n, err := strconv.ParseInt(maxNonce, 10, 64); err == nil {
			config.MaxNonce = n
		}
	}

	if threshold := os.Getenv("HUMAN_VERIFICATION_FAILURE_THRESHOLD"); threshold != "" {
		if t, err := strconv.Atoi(threshold); err == nil {
			config.FailureThreshold = t
		}
	}

	if banDuration := os.Getenv("HUMAN_VERIFICATION_BAN_DURATION"); banDuration != "" {
		if d, err := time.ParseDuration(banDuration); err == nil {
			config.BanDuration = d
		}
	}

	return config
}

// HumanVerificationServiceImpl implements the HumanVerificationService interface
type HumanVerificationServiceImpl struct {
	db     *sql.DB
	config *Config
	client *http.Client
	// Add in-memory challenge storage
	challenges     map[string]*Challenge
	challengeMutex sync.RWMutex
}

// NewHumanVerificationService creates a new human verification service
func NewHumanVerificationService(db *sql.DB, config *Config) HumanVerificationService {
	if config == nil {
		config = DefaultConfig()
	}

	return &HumanVerificationServiceImpl{
		db:     db,
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		challenges: make(map[string]*Challenge),
	}
}

// VerifyResponse verifies a human verification response
func (hvs *HumanVerificationServiceImpl) VerifyResponse(ctx context.Context, emailID, token string, verificationType VerificationType) (bool, error) {
	if !hvs.config.Enabled {
		return true, nil
	}

	switch verificationType {
	case VerificationTypeCAPTCHA:
		return hvs.verifyCAPTCHA(ctx, token)
	case VerificationTypeProofOfWork:
		return hvs.verifyProofOfWork(ctx, token)
	default:
		return false, fmt.Errorf("unsupported verification type: %s", verificationType)
	}
}

// verifyCAPTCHA verifies a CAPTCHA response
func (hvs *HumanVerificationServiceImpl) verifyCAPTCHA(ctx context.Context, token string) (bool, error) {
	if hvs.config.CAPTCHASecretKey == "" {
		log.Printf("CAPTCHA secret key not configured, skipping verification")
		return true, nil
	}

	// Prepare the verification request
	data := url.Values{}
	data.Set("secret", hvs.config.CAPTCHASecretKey)
	data.Set("response", token)

	// Make the verification request
	req, err := http.NewRequestWithContext(ctx, "POST", hvs.config.CAPTCHAEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create CAPTCHA verification request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := hvs.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("CAPTCHA verification request failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Success     bool     `json:"success"`
		Score       float64  `json:"score"`
		Action      string   `json:"action"`
		ChallengeTS string   `json:"challenge_ts"`
		Hostname    string   `json:"hostname"`
		ErrorCodes  []string `json:"error-codes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode CAPTCHA response: %v", err)
	}

	if !result.Success {
		log.Printf("CAPTCHA verification failed: %v", result.ErrorCodes)
		return false, nil
	}

	// For reCAPTCHA v3, check the score (0.0 is very likely a bot, 1.0 is very likely a human)
	if result.Score < 0.5 {
		log.Printf("CAPTCHA score too low: %f", result.Score)
		return false, nil
	}

	return true, nil
}

// verifyProofOfWork verifies a proof-of-work solution
func (hvs *HumanVerificationServiceImpl) verifyProofOfWork(ctx context.Context, solution string) (bool, error) {
	// Parse the solution (format: challenge_id:nonce)
	parts := strings.Split(solution, ":")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid proof-of-work solution format")
	}

	challengeID := parts[0]
	nonceStr := parts[1]

	nonce, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid nonce: %v", err)
	}

	if nonce < 0 || nonce > hvs.config.MaxNonce {
		return false, fmt.Errorf("nonce out of range")
	}

	// Get the challenge from the database
	challenge, err := hvs.getChallenge(ctx, challengeID)
	if err != nil {
		return false, fmt.Errorf("failed to get challenge: %v", err)
	}

	// Verify the solution
	prefix := challenge.Prefix
	target := challenge.Target

	// Calculate hash: SHA256(prefix + nonce)
	data := fmt.Sprintf("%s%d", prefix, nonce)
	hash := sha256.Sum256([]byte(data))
	hashHex := hex.EncodeToString(hash[:])

	// Check if the hash starts with the target
	if !strings.HasPrefix(hashHex, target) {
		return false, nil
	}

	// Clean up the used challenge
	hvs.cleanupChallenge(ctx, challengeID)

	return true, nil
}

// GenerateChallenge generates a new proof-of-work challenge
func (hvs *HumanVerificationServiceImpl) GenerateChallenge(ctx context.Context) (*Challenge, error) {
	// Generate a random prefix
	prefixBytes := make([]byte, 16)
	if _, err := rand.Read(prefixBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random prefix: %v", err)
	}
	prefix := hex.EncodeToString(prefixBytes)

	// Generate target based on difficulty
	target := strings.Repeat("0", hvs.config.ProofOfWorkDifficulty)

	// Generate challenge ID
	challengeID := uuid.New().String()

	challenge := &Challenge{
		ID:       challengeID,
		Prefix:   prefix,
		Target:   target,
		MaxNonce: hvs.config.MaxNonce,
	}

	// Store the challenge in the database
	if err := hvs.storeChallenge(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %v", err)
	}

	return challenge, nil
}

// LogVerification logs a verification attempt
func (hvs *HumanVerificationServiceImpl) LogVerification(ctx context.Context, logEntry *VerificationLog) error {
	if logEntry.ID == "" {
		logEntry.ID = uuid.New().String()
	}

	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now()
	}

	detailsJSON, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log details: %v", err)
	}

	_, err = hvs.db.ExecContext(ctx, `
		INSERT INTO human_verification_logs (
			id, email_id, ip_address, user_agent, verification_type,
			challenge_id, result, timestamp, details
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, logEntry.ID, logEntry.EmailID, logEntry.IPAddress, logEntry.UserAgent,
		logEntry.VerificationType, logEntry.ChallengeID, logEntry.Result,
		logEntry.Timestamp, string(detailsJSON))

	if err != nil {
		return fmt.Errorf("failed to insert verification log: %v", err)
	}

	return nil
}

// GetVerificationStats gets verification statistics for abuse detection
func (hvs *HumanVerificationServiceImpl) GetVerificationStats(ctx context.Context, emailID, ipAddress string, duration time.Duration) (*VerificationStats, error) {
	since := time.Now().Add(-duration)

	query := `
		SELECT 
			COUNT(*) as total_attempts,
			SUM(CASE WHEN result = 'failure' THEN 1 ELSE 0 END) as failed_attempts,
			SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END) as success_attempts
		FROM human_verification_logs
		WHERE timestamp >= ? AND (email_id = ? OR ip_address = ?)
	`

	var stats VerificationStats
	err := hvs.db.QueryRowContext(ctx, query, since, emailID, ipAddress).Scan(
		&stats.TotalAttempts, &stats.FailedAttempts, &stats.SuccessAttempts)

	if err != nil {
		return nil, fmt.Errorf("failed to get verification stats: %v", err)
	}

	if stats.TotalAttempts > 0 {
		stats.FailureRate = float64(stats.FailedAttempts) / float64(stats.TotalAttempts)
	}

	return &stats, nil
}

// storeChallenge stores a challenge in the database
func (hvs *HumanVerificationServiceImpl) storeChallenge(ctx context.Context, challenge *Challenge) error {
	hvs.challengeMutex.Lock()
	defer hvs.challengeMutex.Unlock()

	hvs.challenges[challenge.ID] = challenge
	return nil
}

// getChallenge retrieves a challenge from the database
func (hvs *HumanVerificationServiceImpl) getChallenge(ctx context.Context, challengeID string) (*Challenge, error) {
	hvs.challengeMutex.RLock()
	defer hvs.challengeMutex.RUnlock()

	challenge, exists := hvs.challenges[challengeID]
	if !exists {
		return nil, fmt.Errorf("challenge not found: %s", challengeID)
	}

	return challenge, nil
}

// cleanupChallenge removes a used challenge
func (hvs *HumanVerificationServiceImpl) cleanupChallenge(ctx context.Context, challengeID string) error {
	hvs.challengeMutex.Lock()
	defer hvs.challengeMutex.Unlock()

	delete(hvs.challenges, challengeID)
	log.Printf("Cleaned up challenge: %s", challengeID)
	return nil
}
