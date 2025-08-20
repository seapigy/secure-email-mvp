package e2e

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// HardwareSecurityConfig defines configuration for hardware-backed security
type HardwareSecurityConfig struct {
	Enabled         bool              `json:"enabled"`
	PreferHardware  bool              `json:"prefer_hardware"`
	Platform        string            `json:"platform"`         // "windows", "darwin", "linux"
	Provider        string            `json:"provider"`         // "tpm", "secure_enclave", "pkcs11"
	FallbackMode    string            `json:"fallback_mode"`    // "software", "error", "disabled"
	AttestationMode string            `json:"attestation_mode"` // "required", "optional", "disabled"
	KeyProtection   KeyProtectionMode `json:"key_protection"`
	Timeouts        HSMTimeouts       `json:"timeouts"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// KeyProtectionMode defines how keys are protected in hardware
type KeyProtectionMode struct {
	NonExportable    bool `json:"non_exportable"`
	RequirePIN       bool `json:"require_pin"`
	RequireBiometric bool `json:"require_biometric"`
	SessionBound     bool `json:"session_bound"`
}

// HSMTimeouts defines timeout configuration for HSM operations
type HSMTimeouts struct {
	KeyGeneration time.Duration `json:"key_generation"`
	Signing       time.Duration `json:"signing"`
	Attestation   time.Duration `json:"attestation"`
	Discovery     time.Duration `json:"discovery"`
}

// HardwareSecurityProvider provides hardware-backed cryptographic operations
type HardwareSecurityProvider interface {
	// Platform support
	IsAvailable() bool
	GetCapabilities() HardwareCapabilities
	Initialize(config HardwareSecurityConfig) error
	Close() error

	// Key management
	GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error)
	LoadKey(keyID string) (*HardwareKey, error)
	DeleteKey(keyID string) error
	ListKeys() ([]*HardwareKeyInfo, error)

	// Cryptographic operations
	Sign(keyID string, data []byte) ([]byte, error)
	Verify(keyID string, data, signature []byte) error
	Encrypt(keyID string, plaintext []byte) ([]byte, error)
	Decrypt(keyID string, ciphertext []byte) ([]byte, error)

	// Attestation and security
	AttestKey(keyID string) (*AttestationReport, error)
	GetSecurityState() (*SecurityState, error)
}

// HardwareCapabilities describes hardware security capabilities
type HardwareCapabilities struct {
	Platform              string   `json:"platform"`
	Provider              string   `json:"provider"`
	Version               string   `json:"version"`
	SupportedAlgorithms   []string `json:"supported_algorithms"`
	SupportsAttestation   bool     `json:"supports_attestation"`
	SupportsSealing       bool     `json:"supports_sealing"`
	SupportsNonExportable bool     `json:"supports_non_exportable"`
	MaxKeys               int      `json:"max_keys"`
	Features              []string `json:"features"`
}

// HardwareKey represents a hardware-backed cryptographic key
type HardwareKey struct {
	ID               string                 `json:"id"`
	Algorithm        string                 `json:"algorithm"`
	PublicKey        []byte                 `json:"public_key"`
	KeyHandle        interface{}            `json:"-"` // Platform-specific key handle
	CreatedAt        time.Time              `json:"created_at"`
	Protection       KeyProtectionMode      `json:"protection"`
	Attestation      *AttestationReport     `json:"attestation,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	IsHardwareBacked bool                   `json:"is_hardware_backed"`
}

// HardwareKeyInfo provides summary information about a hardware key
type HardwareKeyInfo struct {
	ID          string            `json:"id"`
	Algorithm   string            `json:"algorithm"`
	PublicKey   []byte            `json:"public_key"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsed    time.Time         `json:"last_used"`
	UsageCount  int64             `json:"usage_count"`
	IsAvailable bool              `json:"is_available"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// KeyParameters defines parameters for hardware key generation
type KeyParameters struct {
	Algorithm  string                 `json:"algorithm"`
	KeySize    int                    `json:"key_size"`
	Usage      []string               `json:"usage"` // "sign", "encrypt", "decrypt", "verify"
	Protection KeyProtectionMode      `json:"protection"`
	Exportable bool                   `json:"exportable"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AttestationReport provides cryptographic proof of key authenticity
type AttestationReport struct {
	KeyID            string            `json:"key_id"`
	Platform         string            `json:"platform"`
	Provider         string            `json:"provider"`
	Timestamp        time.Time         `json:"timestamp"`
	Nonce            []byte            `json:"nonce"`
	Signature        []byte            `json:"signature"`
	Certificate      []byte            `json:"certificate,omitempty"`
	TrustChain       [][]byte          `json:"trust_chain,omitempty"`
	SecurityLevel    string            `json:"security_level"`
	IsHardwareBacked bool              `json:"is_hardware_backed"`
	Claims           map[string]string `json:"claims,omitempty"`
}

// SecurityState represents the current security state of the hardware provider
type SecurityState struct {
	IsSecure          bool              `json:"is_secure"`
	TamperDetected    bool              `json:"tamper_detected"`
	DebugMode         bool              `json:"debug_mode"`
	SecureBootEnabled bool              `json:"secure_boot_enabled"`
	Measurements      map[string]string `json:"measurements,omitempty"`
	LastChecked       time.Time         `json:"last_checked"`
}

// HardwareSecurityManager manages hardware security providers
type HardwareSecurityManager struct {
	config   HardwareSecurityConfig
	provider HardwareSecurityProvider
	fallback *SoftwareFallbackProvider
	keyCache map[string]*HardwareKey
	mutex    sync.RWMutex
	metrics  *HSMMetrics
}

// HSMMetrics tracks hardware security metrics
type HSMMetrics struct {
	KeyGenerations      int64         `json:"key_generations"`
	SignOperations      int64         `json:"sign_operations"`
	VerifyOperations    int64         `json:"verify_operations"`
	EncryptOperations   int64         `json:"encrypt_operations"`
	DecryptOperations   int64         `json:"decrypt_operations"`
	AttestationRequests int64         `json:"attestation_requests"`
	HardwareFailures    int64         `json:"hardware_failures"`
	FallbackUsage       int64         `json:"fallback_usage"`
	AverageLatency      time.Duration `json:"average_latency"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// SoftwareFallbackProvider provides software-based fallback when hardware is unavailable
type SoftwareFallbackProvider struct {
	keys  map[string]*HardwareKey
	mutex sync.RWMutex
}

// NewHardwareSecurityManager creates a new hardware security manager
func NewHardwareSecurityManager(config HardwareSecurityConfig) (*HardwareSecurityManager, error) {
	if !config.Enabled {
		return &HardwareSecurityManager{
			config:   config,
			keyCache: make(map[string]*HardwareKey),
			metrics:  &HSMMetrics{LastUpdated: time.Now()},
		}, nil
	}

	// Detect platform if not specified
	if config.Platform == "" {
		config.Platform = runtime.GOOS
	}

	// Create appropriate provider
	provider, err := createHardwareProvider(config)
	if err != nil && config.FallbackMode == "error" {
		return nil, fmt.Errorf("failed to create hardware provider: %w", err)
	}

	// Create fallback provider
	fallback := &SoftwareFallbackProvider{
		keys: make(map[string]*HardwareKey),
	}

	manager := &HardwareSecurityManager{
		config:   config,
		provider: provider,
		fallback: fallback,
		keyCache: make(map[string]*HardwareKey),
		metrics:  &HSMMetrics{LastUpdated: time.Now()},
	}

	// Initialize provider if available
	if provider != nil && provider.IsAvailable() {
		if err := provider.Initialize(config); err != nil {
			if config.FallbackMode == "error" {
				return nil, fmt.Errorf("failed to initialize hardware provider: %w", err)
			}
		}
	}

	return manager, nil
}

// GenerateKey generates a new hardware-backed key
func (hsm *HardwareSecurityManager) GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error) {
	start := time.Now()
	defer func() {
		hsm.updateMetrics("key_generation", time.Since(start))
	}()

	// Try hardware provider first
	if hsm.provider != nil && hsm.provider.IsAvailable() && hsm.config.PreferHardware {
		key, err := hsm.provider.GenerateKey(algorithm, params)
		if err == nil {
			hsm.cacheKey(key)
			return key, nil
		}

		hsm.metrics.HardwareFailures++
		if hsm.config.FallbackMode == "error" {
			return nil, fmt.Errorf("hardware key generation failed: %w", err)
		}
	}

	// Fallback to software
	if hsm.config.FallbackMode == "software" || hsm.config.FallbackMode == "" {
		hsm.metrics.FallbackUsage++
		return hsm.fallback.GenerateKey(algorithm, params)
	}

	return nil, fmt.Errorf("hardware key generation failed and fallback disabled")
}

// LoadKey loads an existing key by ID
func (hsm *HardwareSecurityManager) LoadKey(keyID string) (*HardwareKey, error) {
	// Check cache first
	hsm.mutex.RLock()
	if key, exists := hsm.keyCache[keyID]; exists {
		hsm.mutex.RUnlock()
		return key, nil
	}
	hsm.mutex.RUnlock()

	// Try hardware provider
	if hsm.provider != nil && hsm.provider.IsAvailable() {
		key, err := hsm.provider.LoadKey(keyID)
		if err == nil {
			hsm.cacheKey(key)
			return key, nil
		}
	}

	// Try fallback
	return hsm.fallback.LoadKey(keyID)
}

// Sign signs data with the specified key
func (hsm *HardwareSecurityManager) Sign(keyID string, data []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		hsm.updateMetrics("sign", time.Since(start))
	}()

	key, err := hsm.LoadKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	}

	if key.IsHardwareBacked && hsm.provider != nil {
		signature, err := hsm.provider.Sign(keyID, data)
		if err == nil {
			hsm.metrics.SignOperations++
			return signature, nil
		}
		hsm.metrics.HardwareFailures++
	}

	// Fallback to software
	hsm.metrics.FallbackUsage++
	return hsm.fallback.Sign(keyID, data)
}

// AttestKey provides cryptographic attestation of a key
func (hsm *HardwareSecurityManager) AttestKey(keyID string) (*AttestationReport, error) {
	if hsm.provider == nil || !hsm.provider.IsAvailable() {
		return nil, fmt.Errorf("hardware attestation not available")
	}

	start := time.Now()
	defer func() {
		hsm.updateMetrics("attestation", time.Since(start))
	}()

	report, err := hsm.provider.AttestKey(keyID)
	if err != nil {
		hsm.metrics.HardwareFailures++
		return nil, fmt.Errorf("attestation failed: %w", err)
	}

	hsm.metrics.AttestationRequests++
	return report, nil
}

// GetSecurityState returns the current security state
func (hsm *HardwareSecurityManager) GetSecurityState() (*SecurityState, error) {
	if hsm.provider == nil || !hsm.provider.IsAvailable() {
		return &SecurityState{
			IsSecure:    false,
			LastChecked: time.Now(),
		}, nil
	}

	return hsm.provider.GetSecurityState()
}

// GetMetrics returns current HSM metrics
func (hsm *HardwareSecurityManager) GetMetrics() *HSMMetrics {
	hsm.mutex.RLock()
	defer hsm.mutex.RUnlock()

	metrics := *hsm.metrics
	return &metrics
}

// Close closes the hardware security manager
func (hsm *HardwareSecurityManager) Close() error {
	if hsm.provider != nil {
		return hsm.provider.Close()
	}
	return nil
}

// Helper methods

func (hsm *HardwareSecurityManager) cacheKey(key *HardwareKey) {
	hsm.mutex.Lock()
	defer hsm.mutex.Unlock()
	hsm.keyCache[key.ID] = key
}

func (hsm *HardwareSecurityManager) updateMetrics(operation string, duration time.Duration) {
	hsm.mutex.Lock()
	defer hsm.mutex.Unlock()

	switch operation {
	case "key_generation":
		hsm.metrics.KeyGenerations++
	case "sign":
		hsm.metrics.SignOperations++
	case "verify":
		hsm.metrics.VerifyOperations++
	case "encrypt":
		hsm.metrics.EncryptOperations++
	case "decrypt":
		hsm.metrics.DecryptOperations++
	case "attestation":
		hsm.metrics.AttestationRequests++
	}

	// Update average latency
	totalOps := hsm.metrics.KeyGenerations + hsm.metrics.SignOperations +
		hsm.metrics.VerifyOperations + hsm.metrics.EncryptOperations +
		hsm.metrics.DecryptOperations + hsm.metrics.AttestationRequests

	if totalOps > 0 {
		hsm.metrics.AverageLatency = time.Duration(
			(int64(hsm.metrics.AverageLatency)*totalOps + int64(duration)) / (totalOps + 1))
	} else {
		hsm.metrics.AverageLatency = duration
	}

	hsm.metrics.LastUpdated = time.Now()
}

// SoftwareFallbackProvider implementation

// GenerateKey generates a software-based key
func (sfp *SoftwareFallbackProvider) GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error) {
	sfp.mutex.Lock()
	defer sfp.mutex.Unlock()

	// Generate key ID
	keyID := generateKeyID()

	// Generate mock key material (simplified)
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	key := &HardwareKey{
		ID:               keyID,
		Algorithm:        algorithm,
		PublicKey:        publicKey,
		CreatedAt:        time.Now(),
		Protection:       params.Protection,
		IsHardwareBacked: false,
		Metadata:         params.Metadata,
	}

	sfp.keys[keyID] = key
	return key, nil
}

// LoadKey loads a software key
func (sfp *SoftwareFallbackProvider) LoadKey(keyID string) (*HardwareKey, error) {
	sfp.mutex.RLock()
	defer sfp.mutex.RUnlock()

	key, exists := sfp.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	return key, nil
}

// Sign signs data with a software key
func (sfp *SoftwareFallbackProvider) Sign(keyID string, data []byte) ([]byte, error) {
	key, err := sfp.LoadKey(keyID)
	if err != nil {
		return nil, err
	}

	// Simplified signing (in production, use proper cryptographic signing)
	hash := sha256.Sum256(append(data, key.PublicKey...))
	return hash[:], nil
}

// Platform-specific provider creation

func createHardwareProvider(config HardwareSecurityConfig) (HardwareSecurityProvider, error) {
	switch config.Platform {
	case "windows":
		return createTPMProvider(config)
	case "darwin":
		return createSecureEnclaveProvider(config)
	case "linux":
		return createPKCS11Provider(config)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", config.Platform)
	}
}

// Mock implementations for different platforms

func createTPMProvider(config HardwareSecurityConfig) (HardwareSecurityProvider, error) {
	return &MockTPMProvider{config: config}, nil
}

func createSecureEnclaveProvider(config HardwareSecurityConfig) (HardwareSecurityProvider, error) {
	return &MockSecureEnclaveProvider{config: config}, nil
}

func createPKCS11Provider(config HardwareSecurityConfig) (HardwareSecurityProvider, error) {
	return &MockPKCS11Provider{config: config}, nil
}

// Mock provider implementations (in production, these would integrate with actual hardware)

type MockTPMProvider struct {
	config HardwareSecurityConfig
}

func (mtp *MockTPMProvider) IsAvailable() bool { return true }
func (mtp *MockTPMProvider) GetCapabilities() HardwareCapabilities {
	return HardwareCapabilities{
		Platform:              "windows",
		Provider:              "tpm",
		Version:               "2.0",
		SupportedAlgorithms:   []string{"rsa2048", "ecc256"},
		SupportsAttestation:   true,
		SupportsSealing:       true,
		SupportsNonExportable: true,
		MaxKeys:               100,
		Features:              []string{"secure_key_storage", "attestation", "sealing"},
	}
}
func (mtp *MockTPMProvider) Initialize(config HardwareSecurityConfig) error { return nil }
func (mtp *MockTPMProvider) Close() error                                   { return nil }
func (mtp *MockTPMProvider) GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error) {
	return mtp.mockGenerateKey(algorithm, params, true)
}
func (mtp *MockTPMProvider) LoadKey(keyID string) (*HardwareKey, error) {
	return nil, fmt.Errorf("not implemented")
}
func (mtp *MockTPMProvider) DeleteKey(keyID string) error          { return nil }
func (mtp *MockTPMProvider) ListKeys() ([]*HardwareKeyInfo, error) { return nil, nil }
func (mtp *MockTPMProvider) Sign(keyID string, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return hash[:], nil
}
func (mtp *MockTPMProvider) Verify(keyID string, data, signature []byte) error { return nil }
func (mtp *MockTPMProvider) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	return plaintext, nil
}
func (mtp *MockTPMProvider) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
func (mtp *MockTPMProvider) AttestKey(keyID string) (*AttestationReport, error) {
	return &AttestationReport{
		KeyID:            keyID,
		Platform:         "windows",
		Provider:         "tpm",
		Timestamp:        time.Now(),
		SecurityLevel:    "hardware",
		IsHardwareBacked: true,
	}, nil
}
func (mtp *MockTPMProvider) GetSecurityState() (*SecurityState, error) {
	return &SecurityState{
		IsSecure:          true,
		TamperDetected:    false,
		DebugMode:         false,
		SecureBootEnabled: true,
		LastChecked:       time.Now(),
	}, nil
}

func (mtp *MockTPMProvider) mockGenerateKey(algorithm string, params KeyParameters, hardwareBacked bool) (*HardwareKey, error) {
	keyID := generateKeyID()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	return &HardwareKey{
		ID:               keyID,
		Algorithm:        algorithm,
		PublicKey:        publicKey,
		CreatedAt:        time.Now(),
		Protection:       params.Protection,
		IsHardwareBacked: hardwareBacked,
		Metadata:         params.Metadata,
	}, nil
}

type MockSecureEnclaveProvider struct {
	config HardwareSecurityConfig
}

func (msep *MockSecureEnclaveProvider) IsAvailable() bool { return true }
func (msep *MockSecureEnclaveProvider) GetCapabilities() HardwareCapabilities {
	return HardwareCapabilities{
		Platform:              "darwin",
		Provider:              "secure_enclave",
		Version:               "1.0",
		SupportedAlgorithms:   []string{"ecc256"},
		SupportsAttestation:   true,
		SupportsSealing:       false,
		SupportsNonExportable: true,
		MaxKeys:               50,
		Features:              []string{"biometric_auth", "secure_key_storage"},
	}
}
func (msep *MockSecureEnclaveProvider) Initialize(config HardwareSecurityConfig) error { return nil }
func (msep *MockSecureEnclaveProvider) Close() error                                   { return nil }
func (msep *MockSecureEnclaveProvider) GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error) {
	return msep.mockGenerateKey(algorithm, params, true)
}
func (msep *MockSecureEnclaveProvider) LoadKey(keyID string) (*HardwareKey, error) {
	return nil, fmt.Errorf("not implemented")
}
func (msep *MockSecureEnclaveProvider) DeleteKey(keyID string) error          { return nil }
func (msep *MockSecureEnclaveProvider) ListKeys() ([]*HardwareKeyInfo, error) { return nil, nil }
func (msep *MockSecureEnclaveProvider) Sign(keyID string, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return hash[:], nil
}
func (msep *MockSecureEnclaveProvider) Verify(keyID string, data, signature []byte) error { return nil }
func (msep *MockSecureEnclaveProvider) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	return plaintext, nil
}
func (msep *MockSecureEnclaveProvider) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
func (msep *MockSecureEnclaveProvider) AttestKey(keyID string) (*AttestationReport, error) {
	return &AttestationReport{
		KeyID:            keyID,
		Platform:         "darwin",
		Provider:         "secure_enclave",
		Timestamp:        time.Now(),
		SecurityLevel:    "hardware",
		IsHardwareBacked: true,
	}, nil
}
func (msep *MockSecureEnclaveProvider) GetSecurityState() (*SecurityState, error) {
	return &SecurityState{
		IsSecure:          true,
		TamperDetected:    false,
		DebugMode:         false,
		SecureBootEnabled: true,
		LastChecked:       time.Now(),
	}, nil
}

func (msep *MockSecureEnclaveProvider) mockGenerateKey(algorithm string, params KeyParameters, hardwareBacked bool) (*HardwareKey, error) {
	keyID := generateKeyID()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	return &HardwareKey{
		ID:               keyID,
		Algorithm:        algorithm,
		PublicKey:        publicKey,
		CreatedAt:        time.Now(),
		Protection:       params.Protection,
		IsHardwareBacked: hardwareBacked,
		Metadata:         params.Metadata,
	}, nil
}

type MockPKCS11Provider struct {
	config HardwareSecurityConfig
}

func (mp11p *MockPKCS11Provider) IsAvailable() bool { return true }
func (mp11p *MockPKCS11Provider) GetCapabilities() HardwareCapabilities {
	return HardwareCapabilities{
		Platform:              "linux",
		Provider:              "pkcs11",
		Version:               "2.40",
		SupportedAlgorithms:   []string{"rsa2048", "rsa4096", "ecc256", "ecc384"},
		SupportsAttestation:   false,
		SupportsSealing:       false,
		SupportsNonExportable: true,
		MaxKeys:               1000,
		Features:              []string{"hsm_backed", "high_performance"},
	}
}
func (mp11p *MockPKCS11Provider) Initialize(config HardwareSecurityConfig) error { return nil }
func (mp11p *MockPKCS11Provider) Close() error                                   { return nil }
func (mp11p *MockPKCS11Provider) GenerateKey(algorithm string, params KeyParameters) (*HardwareKey, error) {
	return mp11p.mockGenerateKey(algorithm, params, true)
}
func (mp11p *MockPKCS11Provider) LoadKey(keyID string) (*HardwareKey, error) {
	return nil, fmt.Errorf("not implemented")
}
func (mp11p *MockPKCS11Provider) DeleteKey(keyID string) error          { return nil }
func (mp11p *MockPKCS11Provider) ListKeys() ([]*HardwareKeyInfo, error) { return nil, nil }
func (mp11p *MockPKCS11Provider) Sign(keyID string, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	return hash[:], nil
}
func (mp11p *MockPKCS11Provider) Verify(keyID string, data, signature []byte) error { return nil }
func (mp11p *MockPKCS11Provider) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	return plaintext, nil
}
func (mp11p *MockPKCS11Provider) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
func (mp11p *MockPKCS11Provider) AttestKey(keyID string) (*AttestationReport, error) {
	return nil, fmt.Errorf("attestation not supported")
}
func (mp11p *MockPKCS11Provider) GetSecurityState() (*SecurityState, error) {
	return &SecurityState{
		IsSecure:    true,
		LastChecked: time.Now(),
	}, nil
}

func (mp11p *MockPKCS11Provider) mockGenerateKey(algorithm string, params KeyParameters, hardwareBacked bool) (*HardwareKey, error) {
	keyID := generateKeyID()
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	return &HardwareKey{
		ID:               keyID,
		Algorithm:        algorithm,
		PublicKey:        publicKey,
		CreatedAt:        time.Now(),
		Protection:       params.Protection,
		IsHardwareBacked: hardwareBacked,
		Metadata:         params.Metadata,
	}, nil
}

// Utility functions

func generateKeyID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("hwkey_%s", hex.EncodeToString(id))
}

// DefaultHardwareSecurityConfig returns a default hardware security configuration
func DefaultHardwareSecurityConfig() HardwareSecurityConfig {
	return HardwareSecurityConfig{
		Enabled:         false, // Disabled by default
		PreferHardware:  true,
		Platform:        runtime.GOOS,
		FallbackMode:    "software",
		AttestationMode: "optional",
		KeyProtection: KeyProtectionMode{
			NonExportable:    true,
			RequirePIN:       false,
			RequireBiometric: false,
			SessionBound:     false,
		},
		Timeouts: HSMTimeouts{
			KeyGeneration: 30 * time.Second,
			Signing:       5 * time.Second,
			Attestation:   10 * time.Second,
			Discovery:     10 * time.Second,
		},
		Metadata: make(map[string]string),
	}
}
