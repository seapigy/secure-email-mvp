package devicefingerprint

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// DeviceFingerprintService provides device fingerprinting functionality
type DeviceFingerprintService interface {
	GenerateFingerprint(userAgent, clientIP string, browserHints map[string]string) (string, error)
	HashFingerprint(fingerprint, emailID string) (string, error)
	IsDeviceTrusted(emailID, deviceHash string) (bool, error)
	TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP string) error
	UpdateDeviceAccess(emailID, deviceHash string) error
	GetDeviceInfo(emailID, deviceHash string) (*DeviceInfo, error)
	GetTrustedDevices(emailID string) ([]*DeviceInfo, error)
	RemoveTrustedDevice(emailID, deviceHash string) error
}

// DeviceInfo represents information about a trusted device
type DeviceInfo struct {
	ID                string    `json:"id"`
	EmailID           string    `json:"email_id"`
	DeviceHash        string    `json:"device_hash"`
	UserAgent         string    `json:"user_agent"`
	IPAddress         string    `json:"ip_address"`
	AddedAt           time.Time `json:"added_at"`
	LastUsedAt        time.Time `json:"last_used_at"`
	AccessCount       int       `json:"access_count"`
	DeviceFingerprint string    `json:"device_fingerprint,omitempty"` // Only for debugging
}

// DeviceFingerprintServiceImpl implements DeviceFingerprintService
type DeviceFingerprintServiceImpl struct {
	db *sql.DB
}

// NewDeviceFingerprintService creates a new device fingerprinting service
func NewDeviceFingerprintService(db *sql.DB) DeviceFingerprintService {
	return &DeviceFingerprintServiceImpl{
		db: db,
	}
}

// GenerateFingerprint creates a deterministic device fingerprint
func (d *DeviceFingerprintServiceImpl) GenerateFingerprint(userAgent, clientIP string, browserHints map[string]string) (string, error) {
	// Create a deterministic fingerprint based on device characteristics
	var components []string

	// Add User-Agent (normalized)
	if userAgent != "" {
		components = append(components, "ua:"+normalizeUserAgent(userAgent))
	}

	// Add IP subnet (for privacy, use /24 for IPv4, /64 for IPv6)
	if clientIP != "" {
		ipSubnet := getIPSubnet(clientIP)
		if ipSubnet != "" {
			components = append(components, "ip:"+ipSubnet)
		}
	}

	// Add browser hints if provided
	if browserHints != nil {
		for key, value := range browserHints {
			if value != "" {
				components = append(components, fmt.Sprintf("hint:%s:%s", key, value))
			}
		}
	}

	// Sort components for deterministic ordering
	components = sortComponents(components)

	// Create fingerprint string
	fingerprint := strings.Join(components, "|")

	// Hash the fingerprint for consistency
	hash := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(hash[:]), nil
}

// HashFingerprint creates an Argon2id hash of the fingerprint using emailID as salt
func (d *DeviceFingerprintServiceImpl) HashFingerprint(fingerprint, emailID string) (string, error) {
	// Use Argon2id with emailID as salt for additional security
	salt := []byte(emailID)
	hash := argon2.IDKey([]byte(fingerprint), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(hash), nil
}

// IsDeviceTrusted checks if a device hash is trusted for a specific email
func (d *DeviceFingerprintServiceImpl) IsDeviceTrusted(emailID, deviceHash string) (bool, error) {
	var exists int
	err := d.db.QueryRow(`
		SELECT 1 FROM trusted_devices 
		WHERE email_id = ? AND device_hash = ?`, emailID, deviceHash).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check device trust: %w", err)
	}

	return true, nil
}

// TrustDevice adds a device to the trusted devices list
func (d *DeviceFingerprintServiceImpl) TrustDevice(emailID, deviceHash, fingerprint, userAgent, clientIP string) error {
	// Generate UUID for the trusted device record
	deviceID := generateUUID()

	_, err := d.db.Exec(`
		INSERT INTO trusted_devices (
			id, email_id, device_hash, device_fingerprint, 
			user_agent, ip_address, added_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		deviceID, emailID, deviceHash, fingerprint, userAgent, clientIP)

	if err != nil {
		return fmt.Errorf("failed to trust device: %w", err)
	}

	log.Printf("Device trusted for email %s: %s", emailID, deviceHash[:16]+"...")
	return nil
}

// UpdateDeviceAccess updates the last access time and count for a trusted device
func (d *DeviceFingerprintServiceImpl) UpdateDeviceAccess(emailID, deviceHash string) error {
	_, err := d.db.Exec(`
		UPDATE trusted_devices 
		SET last_used_at = CURRENT_TIMESTAMP,
		    access_count = access_count + 1
		WHERE email_id = ? AND device_hash = ?`, emailID, deviceHash)

	if err != nil {
		return fmt.Errorf("failed to update device access: %w", err)
	}

	return nil
}

// GetDeviceInfo retrieves information about a specific trusted device
func (d *DeviceFingerprintServiceImpl) GetDeviceInfo(emailID, deviceHash string) (*DeviceInfo, error) {
	var device DeviceInfo
	var lastUsedAt sql.NullTime
	err := d.db.QueryRow(`
		SELECT id, email_id, device_hash, user_agent, ip_address, 
		       added_at, last_used_at, access_count, device_fingerprint
		FROM trusted_devices 
		WHERE email_id = ? AND device_hash = ?`, emailID, deviceHash).Scan(
		&device.ID, &device.EmailID, &device.DeviceHash, &device.UserAgent, &device.IPAddress,
		&device.AddedAt, &lastUsedAt, &device.AccessCount, &device.DeviceFingerprint)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// Handle NULL last_used_at
	if lastUsedAt.Valid {
		device.LastUsedAt = lastUsedAt.Time
	} else {
		device.LastUsedAt = time.Time{} // Zero time
	}

	return &device, nil
}

// GetTrustedDevices retrieves all trusted devices for an email
func (d *DeviceFingerprintServiceImpl) GetTrustedDevices(emailID string) ([]*DeviceInfo, error) {
	rows, err := d.db.Query(`
		SELECT id, email_id, device_hash, user_agent, ip_address, 
		       added_at, last_used_at, access_count, device_fingerprint
		FROM trusted_devices 
		WHERE email_id = ? 
		ORDER BY added_at DESC`, emailID)

	if err != nil {
		return nil, fmt.Errorf("failed to get trusted devices: %w", err)
	}
	defer rows.Close()

	var devices []*DeviceInfo
	for rows.Next() {
		var device DeviceInfo
		var lastUsedAt sql.NullTime
		err := rows.Scan(
			&device.ID, &device.EmailID, &device.DeviceHash, &device.UserAgent, &device.IPAddress,
			&device.AddedAt, &lastUsedAt, &device.AccessCount, &device.DeviceFingerprint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device info: %w", err)
		}

		// Handle NULL last_used_at
		if lastUsedAt.Valid {
			device.LastUsedAt = lastUsedAt.Time
		} else {
			device.LastUsedAt = time.Time{} // Zero time
		}

		devices = append(devices, &device)
	}

	return devices, nil
}

// RemoveTrustedDevice removes a device from the trusted devices list
func (d *DeviceFingerprintServiceImpl) RemoveTrustedDevice(emailID, deviceHash string) error {
	_, err := d.db.Exec(`
		DELETE FROM trusted_devices 
		WHERE email_id = ? AND device_hash = ?`, emailID, deviceHash)

	if err != nil {
		return fmt.Errorf("failed to remove trusted device: %w", err)
	}

	log.Printf("Device removed from trusted list for email %s: %s", emailID, deviceHash[:16]+"...")
	return nil
}

// Helper functions

// normalizeUserAgent normalizes the User-Agent string for consistent fingerprinting
func normalizeUserAgent(userAgent string) string {
	// Convert to lowercase and remove extra whitespace
	normalized := strings.ToLower(strings.TrimSpace(userAgent))

	// Remove version numbers for more stable fingerprinting
	// This helps with minor browser updates
	normalized = removeVersionNumbers(normalized)

	return normalized
}

// removeVersionNumbers removes version numbers from User-Agent strings
func removeVersionNumbers(ua string) string {
	// Simple regex-like replacement for common version patterns
	// This is a simplified approach - in production, you might want more sophisticated parsing

	// Remove version numbers like "Chrome/91.0.4472.124" -> "Chrome"
	ua = strings.ReplaceAll(ua, "/", " ")
	ua = strings.ReplaceAll(ua, ".", " ")

	// Remove common version patterns
	patterns := []string{
		"version", "v ", " ver ", " build ", " rv:",
	}

	for _, pattern := range patterns {
		ua = strings.ReplaceAll(ua, pattern, " ")
	}

	// Clean up multiple spaces
	for strings.Contains(ua, "  ") {
		ua = strings.ReplaceAll(ua, "  ", " ")
	}

	return strings.TrimSpace(ua)
}

// getIPSubnet returns the IP subnet for privacy-preserving fingerprinting
func getIPSubnet(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}

	if parsedIP.To4() != nil {
		// IPv4: use /24 subnet
		ipv4 := parsedIP.To4()
		return fmt.Sprintf("%d.%d.%d.0/24", ipv4[0], ipv4[1], ipv4[2])
	} else {
		// IPv6: use /64 subnet
		ipv6 := parsedIP.To16()
		if ipv6 == nil {
			return ""
		}
		// Simplified IPv6 subnet - in production, you'd want proper IPv6 handling
		return fmt.Sprintf("%x:%x:%x:%x::/64", ipv6[0], ipv6[1], ipv6[2], ipv6[3])
	}
}

// sortComponents sorts fingerprint components for deterministic ordering
func sortComponents(components []string) []string {
	// Simple string sorting for deterministic ordering
	// In production, you might want more sophisticated sorting
	for i := 0; i < len(components)-1; i++ {
		for j := i + 1; j < len(components); j++ {
			if components[i] > components[j] {
				components[i], components[j] = components[j], components[i]
			}
		}
	}
	return components
}

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	// Simple UUID generation - in production, use a proper UUID library
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return hex.EncodeToString(hash[:16])
}
