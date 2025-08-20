package e2e

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// E2EConfig holds the complete configuration for the E2E PQC system
type E2EConfig struct {
	// Feature Flags
	Enabled bool `json:"enabled"`

	// Crypto Configuration
	Crypto CryptoConfig `json:"crypto"`

	// Key Transparency Configuration
	KeyTransparency KTConfig `json:"key_transparency"`

	// HSM Configuration
	HSM HSMConfig `json:"hsm"`

	// Observability Configuration
	Observability ObservabilityConfig `json:"observability"`

	// Safety Configuration
	Safety SafetyConfig `json:"safety"`
}

// CryptoConfig holds cryptographic algorithm and parameter configuration
type CryptoConfig struct {
	// KEM (Key Encapsulation Mechanism)
	KEMAlgorithm string `json:"kem_algorithm"` // "kyber512", "kyber768", "kyber1024"
	KEMLevel     int    `json:"kem_level"`     // 512, 768, or 1024

	// DEM (Data Encapsulation Mechanism)
	DEMAlgorithm string `json:"dem_algorithm"` // "aes256gcm", "chacha20poly1305"

	// Signatures
	SignatureAlgorithm string `json:"signature_algorithm"` // "dilithium2", "dilithium3", "dilithium5"

	// Key Derivation
	KeyRotationDays int `json:"key_rotation_days"` // Key rotation interval in days

	// Performance
	PerformanceMode bool `json:"performance_mode"` // Optimize for performance over security
}

// KTConfig holds Key Transparency service configuration
type KTConfig struct {
	Enabled      bool   `json:"enabled"`
	LogURL       string `json:"log_url"`
	VerifyProofs bool   `json:"verify_proofs"`

	// Merkle Tree Configuration
	TreeHeight int `json:"tree_height"` // Height of the Merkle tree

	// Signing Configuration
	SigningKeyPath string `json:"signing_key_path"`
}

// HSMConfig holds Hardware Security Module configuration
type HSMConfig struct {
	Enabled    bool `json:"enabled"`
	ThresholdM int  `json:"threshold_m"` // M-of-N threshold
	ThresholdN int  `json:"threshold_n"` // Total number of operators

	// HSM Connection
	HSMType       string `json:"hsm_type"` // "software", "pkcs11", "cloudkms"
	HSMConfigPath string `json:"hsm_config_path"`

	// Key Management
	MasterKeyID        string `json:"master_key_id"`
	KeyRotationEnabled bool   `json:"key_rotation_enabled"`
}

// ObservabilityConfig holds monitoring and logging configuration
type ObservabilityConfig struct {
	Enabled bool `json:"enabled"`
	Debug   bool `json:"debug"`

	// Logging
	LogLevel     string `json:"log_level"`     // "debug", "info", "warn", "error"
	LogFormat    string `json:"log_format"`    // "json", "text"
	LogRedaction bool   `json:"log_redaction"` // Enable sensitive data redaction

	// Metrics
	MetricsEnabled bool `json:"metrics_enabled"`
	MetricsPort    int  `json:"metrics_port"`

	// Tracing
	TracingEnabled bool   `json:"tracing_enabled"`
	TracingURL     string `json:"tracing_url"`
}

// SafetyConfig holds safety and security controls
type SafetyConfig struct {
	// Plaintext Protection
	DemoPlaintextMode bool `json:"demo_plaintext_mode"` // NEVER true in production

	// Feature Flag Granularity
	GlobalEnabled bool            `json:"global_enabled"`
	OrgEnabled    map[string]bool `json:"org_enabled"`
	UserEnabled   map[string]bool `json:"user_enabled"`

	// Rollback Controls
	RollbackEnabled bool          `json:"rollback_enabled"`
	RollbackTimeout time.Duration `json:"rollback_timeout"`

	// Validation
	StrictValidation bool `json:"strict_validation"`
	FailClosed       bool `json:"fail_closed"` // Fail closed on errors
}

// DefaultE2EConfig returns the default E2E configuration
func DefaultE2EConfig() *E2EConfig {
	return &E2EConfig{
		Enabled: false, // Disabled by default for safety

		Crypto: CryptoConfig{
			KEMAlgorithm:       "kyber768",
			KEMLevel:           768,
			DEMAlgorithm:       "aes256gcm",
			SignatureAlgorithm: "dilithium3",
			KeyRotationDays:    30,
			PerformanceMode:    false,
		},

		KeyTransparency: KTConfig{
			Enabled:      false,
			LogURL:       "",
			VerifyProofs: true,
			TreeHeight:   20,
		},

		HSM: HSMConfig{
			Enabled:            false,
			ThresholdM:         3,
			ThresholdN:         5,
			HSMType:            "software",
			KeyRotationEnabled: true,
		},

		Observability: ObservabilityConfig{
			Enabled:        true,
			Debug:          false,
			LogLevel:       "info",
			LogFormat:      "json",
			LogRedaction:   true,
			MetricsEnabled: true,
			MetricsPort:    9090,
			TracingEnabled: false,
		},

		Safety: SafetyConfig{
			DemoPlaintextMode: false, // CRITICAL: Never true in production
			GlobalEnabled:     false,
			OrgEnabled:        make(map[string]bool),
			UserEnabled:       make(map[string]bool),
			RollbackEnabled:   true,
			RollbackTimeout:   5 * time.Minute,
			StrictValidation:  true,
			FailClosed:        true,
		},
	}
}

// getTestConfig returns a test configuration with E2E enabled
func getTestConfig() *E2EConfig {
	config := DefaultE2EConfig()
	config.Enabled = true
	config.Safety.GlobalEnabled = true
	config.KeyTransparency.Enabled = true
	config.HSM.Enabled = true
	return config
}

// mustGetTestConfig returns a test configuration with validation
func mustGetTestConfig(t interface{}) *E2EConfig {
	cfg := getTestConfig()
	if !cfg.Enabled || !cfg.Safety.GlobalEnabled {
		// Handle both *testing.T and *testing.B interfaces
		if tb, ok := t.(interface{ Fatal(...interface{}) }); ok {
			tb.Fatal("test config not properly enabled")
		}
		panic("test config not properly enabled")
	}
	return cfg
}

// initTestDB initializes an in-memory SQLite database for testing
func initTestDB(t interface{}) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		if tb, ok := t.(interface{ Fatal(...interface{}) }); ok {
			tb.Fatal("Failed to create test database:", err)
		}
		panic("Failed to create test database: " + err.Error())
	}
	return db
}

// registerTestUsers is a helper to register test users for message routing
func registerTestUsers(users ...string) map[string]bool {
	userMap := make(map[string]bool)
	for _, user := range users {
		userMap[user] = true
	}
	return userMap
}

// LoadE2EConfigFromEnv loads E2E configuration from environment variables
func LoadE2EConfigFromEnv() *E2EConfig {
	config := DefaultE2EConfig()

	// Global Feature Flags
	if enabled := getEnvBool("E2E_ENABLED", false); enabled {
		config.Enabled = enabled
		config.Safety.GlobalEnabled = enabled
	}

	// Crypto Configuration
	if kemAlgo := os.Getenv("PQC_KEM_ALGORITHM"); kemAlgo != "" {
		config.Crypto.KEMAlgorithm = kemAlgo
	}

	if kemLevel := getEnvInt("PQC_KEM_LEVEL", 768); kemLevel > 0 {
		config.Crypto.KEMLevel = kemLevel
	}

	if demAlgo := os.Getenv("PQC_DEM_ALGORITHM"); demAlgo != "" {
		config.Crypto.DEMAlgorithm = demAlgo
	}

	if sigAlgo := os.Getenv("PQC_SIGNATURE_ALGORITHM"); sigAlgo != "" {
		config.Crypto.SignatureAlgorithm = sigAlgo
	}

	if rotationDays := getEnvInt("E2E_KEY_ROTATION_DAYS", 30); rotationDays > 0 {
		config.Crypto.KeyRotationDays = rotationDays
	}

	config.Crypto.PerformanceMode = getEnvBool("E2E_PERFORMANCE_MODE", false)

	// Key Transparency Configuration
	config.KeyTransparency.Enabled = getEnvBool("KT_ENABLED", false)
	if ktURL := os.Getenv("KT_LOG_URL"); ktURL != "" {
		config.KeyTransparency.LogURL = ktURL
	}
	config.KeyTransparency.VerifyProofs = getEnvBool("KT_VERIFY_PROOFS", true)

	if treeHeight := getEnvInt("KT_TREE_HEIGHT", 20); treeHeight > 0 {
		config.KeyTransparency.TreeHeight = treeHeight
	}

	// HSM Configuration
	config.HSM.Enabled = getEnvBool("HSM_ENABLED", false)
	if hsmType := os.Getenv("HSM_TYPE"); hsmType != "" {
		config.HSM.HSMType = hsmType
	}

	if thresholdM := getEnvInt("HSM_THRESHOLD_M", 3); thresholdM > 0 {
		config.HSM.ThresholdM = thresholdM
	}

	if thresholdN := getEnvInt("HSM_THRESHOLD_N", 5); thresholdN > 0 {
		config.HSM.ThresholdN = thresholdN
	}

	if masterKeyID := os.Getenv("HSM_MASTER_KEY_ID"); masterKeyID != "" {
		config.HSM.MasterKeyID = masterKeyID
	}

	config.HSM.KeyRotationEnabled = getEnvBool("HSM_KEY_ROTATION_ENABLED", true)

	// Observability Configuration
	config.Observability.Enabled = getEnvBool("E2E_OBSERVABILITY", true)
	config.Observability.Debug = getEnvBool("E2E_DEBUG", false)

	if logLevel := os.Getenv("E2E_LOG_LEVEL"); logLevel != "" {
		config.Observability.LogLevel = logLevel
	}

	if logFormat := os.Getenv("E2E_LOG_FORMAT"); logFormat != "" {
		config.Observability.LogFormat = logFormat
	}

	config.Observability.LogRedaction = getEnvBool("E2E_LOG_REDACTION", true)
	config.Observability.MetricsEnabled = getEnvBool("E2E_METRICS_ENABLED", true)

	if metricsPort := getEnvInt("E2E_METRICS_PORT", 9090); metricsPort > 0 {
		config.Observability.MetricsPort = metricsPort
	}

	config.Observability.TracingEnabled = getEnvBool("E2E_TRACING_ENABLED", false)
	if tracingURL := os.Getenv("E2E_TRACING_URL"); tracingURL != "" {
		config.Observability.TracingURL = tracingURL
	}

	// Safety Configuration
	// CRITICAL: Demo plaintext mode should NEVER be true in production
	config.Safety.DemoPlaintextMode = getEnvBool("DEMO_PLAINTEXT_MODE", false)
	if config.Safety.DemoPlaintextMode {
		panic("CRITICAL SECURITY ERROR: DEMO_PLAINTEXT_MODE is enabled! This should NEVER be true in production!")
	}

	config.Safety.RollbackEnabled = getEnvBool("E2E_ROLLBACK_ENABLED", true)
	if rollbackTimeout := getEnvInt("E2E_ROLLBACK_TIMEOUT_MINUTES", 5); rollbackTimeout > 0 {
		config.Safety.RollbackTimeout = time.Duration(rollbackTimeout) * time.Minute
	}

	config.Safety.StrictValidation = getEnvBool("E2E_STRICT_VALIDATION", true)
	config.Safety.FailClosed = getEnvBool("E2E_FAIL_CLOSED", true)

	// Load per-org and per-user feature flags
	config.loadOrgAndUserFlags()

	return config
}

// loadOrgAndUserFlags loads organization and user-specific feature flags
func (c *E2EConfig) loadOrgAndUserFlags() {
	// Load organization flags
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "E2E_ORG_ENABLED_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				orgID := strings.TrimPrefix(parts[0], "E2E_ORG_ENABLED_")
				enabled := parts[1] == "true"
				c.Safety.OrgEnabled[orgID] = enabled
			}
		}
	}

	// Load user flags
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "E2E_USER_ENABLED_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				userID := strings.TrimPrefix(parts[0], "E2E_USER_ENABLED_")
				enabled := parts[1] == "true"
				c.Safety.UserEnabled[userID] = enabled
			}
		}
	}
}

// IsEnabledForUser checks if E2E is enabled for a specific user
func (c *E2EConfig) IsEnabledForUser(userID string) bool {
	// Check global flag first
	if !c.Safety.GlobalEnabled {
		return false
	}

	// Check user-specific flag
	if userEnabled, exists := c.Safety.UserEnabled[userID]; exists {
		return userEnabled
	}

	// Default to global setting
	return c.Safety.GlobalEnabled
}

// IsEnabledForOrg checks if E2E is enabled for a specific organization
func (c *E2EConfig) IsEnabledForOrg(orgID string) bool {
	// Check global flag first
	if !c.Safety.GlobalEnabled {
		return false
	}

	// Check org-specific flag
	if orgEnabled, exists := c.Safety.OrgEnabled[orgID]; exists {
		return orgEnabled
	}

	// Default to global setting
	return c.Safety.GlobalEnabled
}

// ValidateConfig validates the E2E configuration for safety and correctness
func (c *E2EConfig) ValidateConfig() error {
	var errors []string

	// Critical safety checks
	if c.Safety.DemoPlaintextMode {
		errors = append(errors, "CRITICAL: DEMO_PLAINTEXT_MODE is enabled - this should NEVER be true in production")
	}

	// Crypto validation
	if c.Crypto.KEMLevel != 512 && c.Crypto.KEMLevel != 768 && c.Crypto.KEMLevel != 1024 {
		errors = append(errors, fmt.Sprintf("Invalid KEM level: %d (must be 512, 768, or 1024)", c.Crypto.KEMLevel))
	}

	if c.Crypto.KeyRotationDays < 1 {
		errors = append(errors, "Key rotation days must be at least 1")
	}

	// HSM validation
	if c.HSM.Enabled {
		if c.HSM.ThresholdM > c.HSM.ThresholdN {
			errors = append(errors, "HSM threshold M cannot be greater than N")
		}

		if c.HSM.ThresholdM < 1 {
			errors = append(errors, "HSM threshold M must be at least 1")
		}

		if c.HSM.ThresholdN < 1 {
			errors = append(errors, "HSM threshold N must be at least 1")
		}
	}

	// KT validation
	if c.KeyTransparency.Enabled {
		if c.KeyTransparency.LogURL == "" {
			errors = append(errors, "KT log URL is required when KT is enabled")
		}

		if c.KeyTransparency.TreeHeight < 1 {
			errors = append(errors, "KT tree height must be at least 1")
		}
	}

	// Observability validation
	if c.Observability.MetricsPort < 1 || c.Observability.MetricsPort > 65535 {
		errors = append(errors, "Metrics port must be between 1 and 65535")
	}

	if len(errors) > 0 {
		return fmt.Errorf("E2E configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ToJSON returns the configuration as a JSON string
func (c *E2EConfig) ToJSON() (string, error) {
	// Create a safe copy for JSON serialization (exclude sensitive fields)
	safeConfig := *c
	safeConfig.HSM.MasterKeyID = "[REDACTED]"
	safeConfig.KeyTransparency.SigningKeyPath = "[REDACTED]"

	data, err := json.MarshalIndent(safeConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(data), nil
}

// GetFeatureFlagStatus returns the current status of all feature flags
func (c *E2EConfig) GetFeatureFlagStatus() map[string]interface{} {
	status := map[string]interface{}{
		"global_enabled": c.Safety.GlobalEnabled,
		"e2e_enabled":    c.Enabled,
		"kt_enabled":     c.KeyTransparency.Enabled,
		"hsm_enabled":    c.HSM.Enabled,
		"debug_enabled":  c.Observability.Debug,
		"org_flags":      c.Safety.OrgEnabled,
		"user_flags":     c.Safety.UserEnabled,
	}

	return status
}

// Helper functions for environment variable parsing

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int(parsed)
		}
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GlobalE2EConfig is the global E2E configuration instance
var GlobalE2EConfig = LoadE2EConfigFromEnv()

// InitializeE2EConfig initializes the global E2E configuration
func InitializeE2EConfig() error {
	// Validate the configuration
	if err := GlobalE2EConfig.ValidateConfig(); err != nil {
		return fmt.Errorf("E2E configuration validation failed: %w", err)
	}

	// Log configuration status (without sensitive data)
	if GlobalE2EConfig.Observability.Debug {
		configJSON, err := GlobalE2EConfig.ToJSON()
		if err == nil {
			fmt.Printf("E2E Configuration loaded:\n%s\n", configJSON)
		}
	}

	return nil
}
