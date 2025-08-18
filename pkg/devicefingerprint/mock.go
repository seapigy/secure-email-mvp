package devicefingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// MockDeviceFingerprintService is a mock implementation for testing
type MockDeviceFingerprintService struct {
	trustedDevices map[string]map[string]*DeviceInfo // emailID -> deviceHash -> DeviceInfo
	fingerprints   map[string]string                 // deviceHash -> fingerprint
}

// NewMockDeviceFingerprintService creates a new mock device fingerprinting service
func NewMockDeviceFingerprintService() *MockDeviceFingerprintService {
	return &MockDeviceFingerprintService{
		trustedDevices: make(map[string]map[string]*DeviceInfo),
		fingerprints:   make(map[string]string),
	}
}

// GenerateFingerprint creates a mock deterministic device fingerprint
func (m *MockDeviceFingerprintService) GenerateFingerprint(userAgent, clientIP string, browserHints map[string]string) (string, error) {
	// Create a simple deterministic fingerprint for testing
	var components []string

	if userAgent != "" {
		components = append(components, "ua:"+strings.ToLower(strings.TrimSpace(userAgent)))
	}

	if clientIP != "" {
		// Use the full IP for mock testing to ensure different fingerprints
		components = append(components, "ip:"+clientIP)
	}

	if browserHints != nil {
		for key, value := range browserHints {
			if value != "" {
				components = append(components, fmt.Sprintf("hint:%s:%s", key, value))
			}
		}
	}

	// Sort components for deterministic ordering
	for i := 0; i < len(components)-1; i++ {
		for j := i + 1; j < len(components); j++ {
			if components[i] > components[j] {
				components[i], components[j] = components[j], components[i]
			}
		}
	}

	fingerprint := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(hash[:]), nil
}

// HashFingerprint creates a mock hash of the fingerprint
func (m *MockDeviceFingerprintService) HashFingerprint(fingerprint, emailID string) (string, error) {
	// Simple hash for testing - combine fingerprint and emailID
	combined := fingerprint + "|" + emailID
	hash := sha256.Sum256([]byte(combined))
	deviceHash := hex.EncodeToString(hash[:])

	// Store the fingerprint for later retrieval
	m.fingerprints[deviceHash] = fingerprint

	return deviceHash, nil
}

// IsDeviceTrusted checks if a device hash is trusted for a specific email
func (m *MockDeviceFingerprintService) IsDeviceTrusted(emailID, deviceHash string) (bool, error) {
	if emailDevices, exists := m.trustedDevices[emailID]; exists {
		_, trusted := emailDevices[deviceHash]
		return trusted, nil
	}
	return false, nil
}

// TrustDevice adds a device to the trusted devices list
func (m *MockDeviceFingerprintService) TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP string) error {
	if m.trustedDevices[emailID] == nil {
		m.trustedDevices[emailID] = make(map[string]*DeviceInfo)
	}

	deviceInfo := &DeviceInfo{
		ID:                fmt.Sprintf("mock-device-%s", deviceHash[:8]),
		EmailID:           emailID,
		DeviceHash:        deviceHash,
		UserAgent:         userAgent,
		IPAddress:         clientIP,
		AddedAt:           time.Now(),
		LastUsedAt:        time.Now(),
		AccessCount:       0,
		DeviceFingerprint: fingerprint,
	}

	m.trustedDevices[emailID][deviceHash] = deviceInfo
	return nil
}

// UpdateDeviceAccess updates the last access time and count for a trusted device
func (m *MockDeviceFingerprintService) UpdateDeviceAccess(emailID, deviceHash string) error {
	if emailDevices, exists := m.trustedDevices[emailID]; exists {
		if device, trusted := emailDevices[deviceHash]; trusted {
			device.LastUsedAt = time.Now()
			device.AccessCount++
			return nil
		}
	}
	return fmt.Errorf("device not found or not trusted")
}

// GetDeviceInfo retrieves information about a specific trusted device
func (m *MockDeviceFingerprintService) GetDeviceInfo(emailID, deviceHash string) (*DeviceInfo, error) {
	if emailDevices, exists := m.trustedDevices[emailID]; exists {
		if device, trusted := emailDevices[deviceHash]; trusted {
			return device, nil
		}
	}
	return nil, fmt.Errorf("device not found")
}

// GetTrustedDevices retrieves all trusted devices for an email
func (m *MockDeviceFingerprintService) GetTrustedDevices(emailID string) ([]*DeviceInfo, error) {
	if emailDevices, exists := m.trustedDevices[emailID]; exists {
		var devices []*DeviceInfo
		for _, device := range emailDevices {
			devices = append(devices, device)
		}
		return devices, nil
	}
	return []*DeviceInfo{}, nil
}

// RemoveTrustedDevice removes a device from the trusted devices list
func (m *MockDeviceFingerprintService) RemoveTrustedDevice(emailID, deviceHash string) error {
	if emailDevices, exists := m.trustedDevices[emailID]; exists {
		delete(emailDevices, deviceHash)
		return nil
	}
	return fmt.Errorf("device not found")
}

// SetTrustedDevice allows setting a trusted device for testing
func (m *MockDeviceFingerprintService) SetTrustedDevice(emailID, deviceHash string, deviceInfo *DeviceInfo) {
	if m.trustedDevices[emailID] == nil {
		m.trustedDevices[emailID] = make(map[string]*DeviceInfo)
	}
	m.trustedDevices[emailID][deviceHash] = deviceInfo
}

// ClearTrustedDevices clears all trusted devices for testing
func (m *MockDeviceFingerprintService) ClearTrustedDevices() {
	m.trustedDevices = make(map[string]map[string]*DeviceInfo)
	m.fingerprints = make(map[string]string)
}
