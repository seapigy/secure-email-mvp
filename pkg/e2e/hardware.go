// Package e2e provides end-to-end encryption functionality with hardware-backed keys
package e2e

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"runtime"
	"time"
)

// HardwareKeyRef represents a reference to a key stored in hardware
type HardwareKeyRef struct {
	KeyID       string    `json:"key_id"`
	Algorithm   string    `json:"algorithm"`
	Platform    string    `json:"platform"`
	Handle      string    `json:"handle,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used,omitempty"`
	IsAvailable bool      `json:"is_available"`
}

// PlatformInfo contains information about the hardware platform
type PlatformInfo struct {
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	Capabilities []string `json:"capabilities"`
	IsAvailable bool   `json:"is_available"`
}

// HardwareKeyManager interface defines operations for hardware-backed key storage
type HardwareKeyManager interface {
	// Key Generation
	GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error)
	
	// Key Operations
	Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error)
	Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error)
	
	// Key Management
	DeleteKey(keyRef *HardwareKeyRef) error
	ListKeys() ([]*HardwareKeyRef, error)
	
	// Platform Info
	IsAvailable() bool
	GetPlatformInfo() *PlatformInfo
}

// HardwareKeyConfig contains configuration for hardware key management
type HardwareKeyConfig struct {
	Enabled          bool   `json:"enabled"`
	PreferredPlatform string `json:"preferred_platform"` // "tpm", "enclave", "pkcs11", "auto"
	FallbackToSoftware bool   `json:"fallback_to_software"`
	
	// Platform-specific configs
	TPMConfig     TPMConfig     `json:"tpm_config,omitempty"`
	EnclaveConfig EnclaveConfig `json:"enclave_config,omitempty"`
	HSMConfig     HardwareHSMConfig `json:"hsm_config,omitempty"`
}

// TPMConfig contains TPM-specific configuration
type TPMConfig struct {
	DevicePath    string `json:"device_path"`
	OwnerPassword string `json:"owner_password,omitempty"`
	EndorsementPassword string `json:"endorsement_password,omitempty"`
	KeyHierarchy  string `json:"key_hierarchy"` // "storage", "endorsement", "platform"
}

// EnclaveConfig contains Secure Enclave configuration
type EnclaveConfig struct {
	KeychainPath string `json:"keychain_path"`
	AccessControl string `json:"access_control"` // "user", "device", "biometric"
	KeySize      int    `json:"key_size"`
}

// HardwareHSMConfig contains PKCS#11 HSM configuration
type HardwareHSMConfig struct {
	ModulePath   string `json:"module_path"`
	TokenLabel   string `json:"token_label"`
	UserPIN      string `json:"user_pin,omitempty"`
	SlotID       int    `json:"slot_id"`
	KeyTemplate  string `json:"key_template"`
}

// SoftwareKeyManager provides fallback software implementation
type SoftwareKeyManager struct {
	keys map[string]*SoftwareKey
}

// SoftwareKey represents a software-stored key
type SoftwareKey struct {
	KeyID       string
	Algorithm   string
	PrivateKey  interface{}
	PublicKey   interface{}
	CreatedAt   time.Time
	LastUsed    time.Time
}

// NewSoftwareKeyManager creates a new software key manager
func NewSoftwareKeyManager() *SoftwareKeyManager {
	return &SoftwareKeyManager{
		keys: make(map[string]*SoftwareKey),
	}
}

// GenerateKey creates a new software key
func (s *SoftwareKeyManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
	var privateKey interface{}
	var publicKey interface{}
	
	switch algorithm {
	case "RSA-2048":
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		privateKey = rsaKey
		publicKey = &rsaKey.PublicKey
		
	case "RSA-4096":
		rsaKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		privateKey = rsaKey
		publicKey = &rsaKey.PublicKey
		
	case "EC-P256":
		// Placeholder for EC key generation
		return nil, fmt.Errorf("EC key generation not implemented")
		
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	
	key := &SoftwareKey{
		KeyID:      keyID,
		Algorithm:  algorithm,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CreatedAt:  time.Now(),
	}
	
	s.keys[keyID] = key
	
	return &HardwareKeyRef{
		KeyID:       keyID,
		Algorithm:   algorithm,
		Platform:    "software",
		Handle:      keyID,
		CreatedAt:   key.CreatedAt,
		IsAvailable: true,
	}, nil
}

// Sign signs data using a software key
func (s *SoftwareKeyManager) Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error) {
	key, exists := s.keys[keyRef.KeyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyRef.KeyID)
	}
	
	key.LastUsed = time.Now()
	
	switch key.Algorithm {
	case "RSA-2048", "RSA-4096":
		rsaKey, ok := key.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type for RSA signing")
		}
		
		hash := sha256.Sum256(data)
		signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hash[:])
		if err != nil {
			return nil, fmt.Errorf("RSA signing failed: %w", err)
		}
		
		return signature, nil
		
	default:
		return nil, fmt.Errorf("unsupported algorithm for signing: %s", key.Algorithm)
	}
}

// Decrypt decrypts data using a software key
func (s *SoftwareKeyManager) Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error) {
	key, exists := s.keys[keyRef.KeyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyRef.KeyID)
	}
	
	key.LastUsed = time.Now()
	
	switch key.Algorithm {
	case "RSA-2048", "RSA-4096":
		rsaKey, ok := key.PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("invalid key type for RSA decryption")
		}
		
		plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, rsaKey, ciphertext)
		if err != nil {
			return nil, fmt.Errorf("RSA decryption failed: %w", err)
		}
		
		return plaintext, nil
		
	default:
		return nil, fmt.Errorf("unsupported algorithm for decryption: %s", key.Algorithm)
	}
}

// DeleteKey removes a software key
func (s *SoftwareKeyManager) DeleteKey(keyRef *HardwareKeyRef) error {
	if _, exists := s.keys[keyRef.KeyID]; !exists {
		return fmt.Errorf("key not found: %s", keyRef.KeyID)
	}
	
	delete(s.keys, keyRef.KeyID)
	return nil
}

// ListKeys returns all software keys
func (s *SoftwareKeyManager) ListKeys() ([]*HardwareKeyRef, error) {
	var refs []*HardwareKeyRef
	
	for _, key := range s.keys {
		refs = append(refs, &HardwareKeyRef{
			KeyID:       key.KeyID,
			Algorithm:   key.Algorithm,
			Platform:    "software",
			Handle:      key.KeyID,
			CreatedAt:   key.CreatedAt,
			LastUsed:    key.LastUsed,
			IsAvailable: true,
		})
	}
	
	return refs, nil
}

// IsAvailable always returns true for software implementation
func (s *SoftwareKeyManager) IsAvailable() bool {
	return true
}

// GetPlatformInfo returns software platform information
func (s *SoftwareKeyManager) GetPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		Platform:     "software",
		Version:      "1.0.0",
		Capabilities: []string{"RSA-2048", "RSA-4096"},
		IsAvailable:  true,
	}
}

// WindowsTPMManager provides TPM 2.0 integration for Windows
type WindowsTPMManager struct {
	config TPMConfig
	// tpmDevice *tpm2.TPMDevice // Placeholder for actual TPM implementation
}

// NewWindowsTPMManager creates a new Windows TPM manager
func NewWindowsTPMManager(config TPMConfig) *WindowsTPMManager {
	return &WindowsTPMManager{
		config: config,
	}
}

// GenerateKey creates a key in Windows TPM (placeholder implementation)
func (w *WindowsTPMManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
	// Placeholder implementation - would use Windows TPM 2.0 API
	return &HardwareKeyRef{
		KeyID:       keyID,
		Algorithm:   algorithm,
		Platform:    "windows_tpm",
		Handle:      fmt.Sprintf("tpm_%s", keyID),
		CreatedAt:   time.Now(),
		IsAvailable: true,
	}, nil
}

// Sign signs data using TPM (placeholder implementation)
func (w *WindowsTPMManager) Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error) {
	// Placeholder implementation - would use TPM signing
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// Decrypt decrypts data using TPM (placeholder implementation)
func (w *WindowsTPMManager) Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error) {
	// Placeholder implementation - would use TPM decryption
	return ciphertext, nil
}

// DeleteKey removes a key from TPM (placeholder implementation)
func (w *WindowsTPMManager) DeleteKey(keyRef *HardwareKeyRef) error {
	// Placeholder implementation - would use TPM key deletion
	return nil
}

// ListKeys lists TPM keys (placeholder implementation)
func (w *WindowsTPMManager) ListKeys() ([]*HardwareKeyRef, error) {
	// Placeholder implementation - would enumerate TPM keys
	return []*HardwareKeyRef{}, nil
}

// IsAvailable checks if TPM is available
func (w *WindowsTPMManager) IsAvailable() bool {
	// Placeholder implementation - would check TPM availability
	return runtime.GOOS == "windows"
}

// GetPlatformInfo returns TPM platform information
func (w *WindowsTPMManager) GetPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		Platform:     "windows_tpm",
		Version:      "2.0",
		Capabilities: []string{"RSA-2048", "RSA-4096", "EC-P256", "EC-P384"},
		IsAvailable:  w.IsAvailable(),
	}
}

// MacOSEnclaveManager provides Secure Enclave integration for macOS
type MacOSEnclaveManager struct {
	config EnclaveConfig
	// keychainRef SecKeychainRef // Placeholder for actual Security Framework
}

// NewMacOSEnclaveManager creates a new macOS Secure Enclave manager
func NewMacOSEnclaveManager(config EnclaveConfig) *MacOSEnclaveManager {
	return &MacOSEnclaveManager{
		config: config,
	}
}

// GenerateKey creates a key in Secure Enclave (placeholder implementation)
func (m *MacOSEnclaveManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
	// Placeholder implementation - would use Security Framework
	return &HardwareKeyRef{
		KeyID:       keyID,
		Algorithm:   algorithm,
		Platform:    "macos_enclave",
		Handle:      fmt.Sprintf("enclave_%s", keyID),
		CreatedAt:   time.Now(),
		IsAvailable: true,
	}, nil
}

// Sign signs data using Secure Enclave (placeholder implementation)
func (m *MacOSEnclaveManager) Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error) {
	// Placeholder implementation - would use Secure Enclave signing
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// Decrypt decrypts data using Secure Enclave (placeholder implementation)
func (m *MacOSEnclaveManager) Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error) {
	// Placeholder implementation - would use Secure Enclave decryption
	return ciphertext, nil
}

// DeleteKey removes a key from Secure Enclave (placeholder implementation)
func (m *MacOSEnclaveManager) DeleteKey(keyRef *HardwareKeyRef) error {
	// Placeholder implementation - would use Security Framework
	return nil
}

// ListKeys lists Secure Enclave keys (placeholder implementation)
func (m *MacOSEnclaveManager) ListKeys() ([]*HardwareKeyRef, error) {
	// Placeholder implementation - would enumerate Secure Enclave keys
	return []*HardwareKeyRef{}, nil
}

// IsAvailable checks if Secure Enclave is available
func (m *MacOSEnclaveManager) IsAvailable() bool {
	// Placeholder implementation - would check Secure Enclave availability
	return runtime.GOOS == "darwin"
}

// GetPlatformInfo returns Secure Enclave platform information
func (m *MacOSEnclaveManager) GetPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		Platform:     "macos_enclave",
		Version:      "1.0",
		Capabilities: []string{"RSA-2048", "RSA-4096", "EC-P256", "EC-P384"},
		IsAvailable:  m.IsAvailable(),
	}
}

// LinuxHSMManager provides PKCS#11 HSM integration for Linux
type LinuxHSMManager struct {
	config HardwareHSMConfig
	// pkcs11Module *pkcs11.Module // Placeholder for actual PKCS#11 implementation
}

// NewLinuxHSMManager creates a new Linux HSM manager
func NewLinuxHSMManager(config HardwareHSMConfig) *LinuxHSMManager {
	return &LinuxHSMManager{
		config: config,
	}
}

// GenerateKey creates a key in HSM (placeholder implementation)
func (l *LinuxHSMManager) GenerateKey(algorithm string, keyID string) (*HardwareKeyRef, error) {
	// Placeholder implementation - would use PKCS#11
	return &HardwareKeyRef{
		KeyID:       keyID,
		Algorithm:   algorithm,
		Platform:    "linux_hsm",
		Handle:      fmt.Sprintf("hsm_%s", keyID),
		CreatedAt:   time.Now(),
		IsAvailable: true,
	}, nil
}

// Sign signs data using HSM (placeholder implementation)
func (l *LinuxHSMManager) Sign(keyRef *HardwareKeyRef, data []byte) ([]byte, error) {
	// Placeholder implementation - would use PKCS#11 signing
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// Decrypt decrypts data using HSM (placeholder implementation)
func (l *LinuxHSMManager) Decrypt(keyRef *HardwareKeyRef, ciphertext []byte) ([]byte, error) {
	// Placeholder implementation - would use PKCS#11 decryption
	return ciphertext, nil
}

// DeleteKey removes a key from HSM (placeholder implementation)
func (l *LinuxHSMManager) DeleteKey(keyRef *HardwareKeyRef) error {
	// Placeholder implementation - would use PKCS#11
	return nil
}

// ListKeys lists HSM keys (placeholder implementation)
func (l *LinuxHSMManager) ListKeys() ([]*HardwareKeyRef, error) {
	// Placeholder implementation - would enumerate HSM keys
	return []*HardwareKeyRef{}, nil
}

// IsAvailable checks if HSM is available
func (l *LinuxHSMManager) IsAvailable() bool {
	// Placeholder implementation - would check HSM availability
	return runtime.GOOS == "linux"
}

// GetPlatformInfo returns HSM platform information
func (l *LinuxHSMManager) GetPlatformInfo() *PlatformInfo {
	return &PlatformInfo{
		Platform:     "linux_hsm",
		Version:      "PKCS#11",
		Capabilities: []string{"RSA-2048", "RSA-4096", "EC-P256", "EC-P384", "AES-256"},
		IsAvailable:  l.IsAvailable(),
	}
}

// CreateHardwareKeyManager creates the appropriate hardware key manager for the platform
func CreateHardwareKeyManager(config HardwareKeyConfig) (HardwareKeyManager, error) {
	if !config.Enabled {
		return NewSoftwareKeyManager(), nil
	}
	
	platform := config.PreferredPlatform
	if platform == "auto" {
		platform = detectPlatform()
	}
	
	switch platform {
	case "tpm":
		if runtime.GOOS == "windows" {
			return NewWindowsTPMManager(config.TPMConfig), nil
		}
		fallthrough
		
	case "enclave":
		if runtime.GOOS == "darwin" {
			return NewMacOSEnclaveManager(config.EnclaveConfig), nil
		}
		fallthrough
		
	case "pkcs11":
		if runtime.GOOS == "linux" {
			return NewLinuxHSMManager(config.HSMConfig), nil
		}
		fallthrough
		
	default:
		if config.FallbackToSoftware {
			return NewSoftwareKeyManager(), nil
		}
		return nil, fmt.Errorf("no suitable hardware key manager available for platform: %s", platform)
	}
}

// detectPlatform automatically detects the best available platform
func detectPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "tpm"
	case "darwin":
		return "enclave"
	case "linux":
		return "pkcs11"
	default:
		return "software"
	}
}
