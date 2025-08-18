package devicefingerprint

import (
	"testing"
	"time"
)

func TestMockDeviceFingerprintService_GenerateFingerprint(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	tests := []struct {
		name         string
		userAgent    string
		clientIP     string
		browserHints map[string]string
		expected     string
	}{
		{
			name:      "Basic fingerprint",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:  "192.168.1.100",
			expected:  "", // We'll check it's not empty
		},
		{
			name:      "Same device should generate same fingerprint",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:  "192.168.1.100",
			expected:  "", // We'll check it matches the first one
		},
		{
			name:      "Different IP should generate different fingerprint",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:  "192.168.1.200",
			expected:  "", // We'll check it's different
		},
		{
			name:         "With browser hints",
			userAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:     "192.168.1.100",
			browserHints: map[string]string{"screen": "1920x1080", "timezone": "UTC"},
			expected:     "", // We'll check it's not empty
		},
	}

	var firstFingerprint string
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fingerprint, err := service.GenerateFingerprint(tt.userAgent, tt.clientIP, tt.browserHints)
			if err != nil {
				t.Fatalf("GenerateFingerprint() error = %v", err)
			}

			if fingerprint == "" {
				t.Error("GenerateFingerprint() returned empty fingerprint")
			}

			if i == 0 {
				firstFingerprint = fingerprint
			} else if i == 1 {
				// Same device should generate same fingerprint
				if fingerprint != firstFingerprint {
					t.Errorf("Same device generated different fingerprints: %s vs %s", firstFingerprint, fingerprint)
				}
			} else if i == 2 {
				// Different IP should generate different fingerprint
				if fingerprint == firstFingerprint {
					t.Errorf("Different IP generated same fingerprint: %s", fingerprint)
				}
			}
		})
	}
}

func TestMockDeviceFingerprintService_HashFingerprint(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	fingerprint := "test-fingerprint"
	emailID := "test-email-123"

	hash1, err := service.HashFingerprint(fingerprint, emailID)
	if err != nil {
		t.Fatalf("HashFingerprint() error = %v", err)
	}

	if hash1 == "" {
		t.Error("HashFingerprint() returned empty hash")
	}

	// Same fingerprint and email should generate same hash
	hash2, err := service.HashFingerprint(fingerprint, emailID)
	if err != nil {
		t.Fatalf("HashFingerprint() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Same fingerprint and email generated different hashes: %s vs %s", hash1, hash2)
	}

	// Different email should generate different hash
	hash3, err := service.HashFingerprint(fingerprint, "different-email")
	if err != nil {
		t.Fatalf("HashFingerprint() error = %v", err)
	}

	if hash1 == hash3 {
		t.Errorf("Different email generated same hash: %s", hash1)
	}
}

func TestMockDeviceFingerprintService_TrustDevice(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"
	deviceHash := "test-device-hash"
	fingerprint := "test-fingerprint"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	clientIP := "192.168.1.100"

	// Initially, device should not be trusted
	trusted, err := service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if trusted {
		t.Error("Device should not be trusted initially")
	}

	// Trust the device
	err = service.TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP)
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Now device should be trusted
	trusted, err = service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if !trusted {
		t.Error("Device should be trusted after TrustDevice()")
	}
}

func TestMockDeviceFingerprintService_UpdateDeviceAccess(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"
	deviceHash := "test-device-hash"
	fingerprint := "test-fingerprint"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	clientIP := "192.168.1.100"

	// Trust the device first
	err := service.TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP)
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Get initial device info
	deviceInfo, err := service.GetDeviceInfo(emailID, deviceHash)
	if err != nil {
		t.Fatalf("GetDeviceInfo() error = %v", err)
	}

	initialAccessCount := deviceInfo.AccessCount
	initialLastUsed := deviceInfo.LastUsedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update device access
	err = service.UpdateDeviceAccess(emailID, deviceHash)
	if err != nil {
		t.Fatalf("UpdateDeviceAccess() error = %v", err)
	}

	// Get updated device info
	deviceInfo, err = service.GetDeviceInfo(emailID, deviceHash)
	if err != nil {
		t.Fatalf("GetDeviceInfo() error = %v", err)
	}

	// Check that access count increased
	if deviceInfo.AccessCount != initialAccessCount+1 {
		t.Errorf("Access count should have increased from %d to %d, got %d",
			initialAccessCount, initialAccessCount+1, deviceInfo.AccessCount)
	}

	// Check that last used time was updated
	if !deviceInfo.LastUsedAt.After(initialLastUsed) {
		t.Error("Last used time should have been updated")
	}
}

func TestMockDeviceFingerprintService_GetTrustedDevices(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"

	// Initially, no trusted devices
	devices, err := service.GetTrustedDevices(emailID)
	if err != nil {
		t.Fatalf("GetTrustedDevices() error = %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("Expected 0 trusted devices, got %d", len(devices))
	}

	// Trust a device
	deviceHash1 := "device-hash-1"
	err = service.TrustDevice(emailID, deviceHash1, "fingerprint-1", "user-agent-1", "192.168.1.100")
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Trust another device
	deviceHash2 := "device-hash-2"
	err = service.TrustDevice(emailID, deviceHash2, "fingerprint-2", "user-agent-2", "192.168.1.200")
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Now should have 2 trusted devices
	devices, err = service.GetTrustedDevices(emailID)
	if err != nil {
		t.Fatalf("GetTrustedDevices() error = %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("Expected 2 trusted devices, got %d", len(devices))
	}

	// Check that both devices are in the list
	deviceHashes := make(map[string]bool)
	for _, device := range devices {
		deviceHashes[device.DeviceHash] = true
	}

	if !deviceHashes[deviceHash1] {
		t.Error("First device not found in trusted devices list")
	}
	if !deviceHashes[deviceHash2] {
		t.Error("Second device not found in trusted devices list")
	}
}

func TestMockDeviceFingerprintService_RemoveTrustedDevice(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"
	deviceHash := "test-device-hash"
	fingerprint := "test-fingerprint"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	clientIP := "192.168.1.100"

	// Trust the device
	err := service.TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP)
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Verify device is trusted
	trusted, err := service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if !trusted {
		t.Error("Device should be trusted after TrustDevice()")
	}

	// Remove the device
	err = service.RemoveTrustedDevice(emailID, deviceHash)
	if err != nil {
		t.Fatalf("RemoveTrustedDevice() error = %v", err)
	}

	// Verify device is no longer trusted
	trusted, err = service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if trusted {
		t.Error("Device should not be trusted after RemoveTrustedDevice()")
	}
}

func TestMockDeviceFingerprintService_GetDeviceInfo(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"
	deviceHash := "test-device-hash"
	fingerprint := "test-fingerprint"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	clientIP := "192.168.1.100"

	// Try to get info for non-existent device
	_, err := service.GetDeviceInfo(emailID, deviceHash)
	if err == nil {
		t.Error("GetDeviceInfo() should return error for non-existent device")
	}

	// Trust the device
	err = service.TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP)
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Get device info
	deviceInfo, err := service.GetDeviceInfo(emailID, deviceHash)
	if err != nil {
		t.Fatalf("GetDeviceInfo() error = %v", err)
	}

	// Verify device info
	if deviceInfo.EmailID != emailID {
		t.Errorf("Expected EmailID %s, got %s", emailID, deviceInfo.EmailID)
	}
	if deviceInfo.DeviceHash != deviceHash {
		t.Errorf("Expected DeviceHash %s, got %s", deviceHash, deviceInfo.DeviceHash)
	}
	if deviceInfo.UserAgent != userAgent {
		t.Errorf("Expected UserAgent %s, got %s", userAgent, deviceInfo.UserAgent)
	}
	if deviceInfo.IPAddress != clientIP {
		t.Errorf("Expected IPAddress %s, got %s", clientIP, deviceInfo.IPAddress)
	}
	if deviceInfo.DeviceFingerprint != fingerprint {
		t.Errorf("Expected DeviceFingerprint %s, got %s", fingerprint, deviceInfo.DeviceFingerprint)
	}
	if deviceInfo.AccessCount != 0 {
		t.Errorf("Expected AccessCount 0, got %d", deviceInfo.AccessCount)
	}
}

func TestMockDeviceFingerprintService_ClearTrustedDevices(t *testing.T) {
	service := NewMockDeviceFingerprintService()

	emailID := "test-email-123"
	deviceHash := "test-device-hash"
	fingerprint := "test-fingerprint"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
	clientIP := "192.168.1.100"

	// Trust a device
	err := service.TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP)
	if err != nil {
		t.Fatalf("TrustDevice() error = %v", err)
	}

	// Verify device is trusted
	trusted, err := service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if !trusted {
		t.Error("Device should be trusted after TrustDevice()")
	}

	// Clear all trusted devices
	service.ClearTrustedDevices()

	// Verify device is no longer trusted
	trusted, err = service.IsDeviceTrusted(emailID, deviceHash)
	if err != nil {
		t.Fatalf("IsDeviceTrusted() error = %v", err)
	}
	if trusted {
		t.Error("Device should not be trusted after ClearTrustedDevices()")
	}

	// Verify no trusted devices exist
	devices, err := service.GetTrustedDevices(emailID)
	if err != nil {
		t.Fatalf("GetTrustedDevices() error = %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("Expected 0 trusted devices after clear, got %d", len(devices))
	}
}
