package audit

import (
	"fmt"
	"net"
	"strings"
)

// AnonymizeIP anonymizes an IP address by masking the last octet (IPv4) or last 64 bits (IPv6)
// This ensures privacy compliance while still providing useful location information
func AnonymizeIP(ipAddress string) (string, error) {
	if ipAddress == "" {
		return "", fmt.Errorf("empty IP address")
	}

	// Parse the IP address
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipAddress)
	}

	// Handle IPv4 addresses
	if ipv4 := ip.To4(); ipv4 != nil {
		// Mask the last octet (e.g., 192.168.1.100 -> 192.168.1.0/24)
		masked := net.IPv4(ipv4[0], ipv4[1], ipv4[2], 0)
		return fmt.Sprintf("%s/24", masked.String()), nil
	}

	// Handle IPv6 addresses
	if ipv6 := ip.To16(); ipv6 != nil {
		// For IPv6, mask the last 64 bits (e.g., 2001:db8::1 -> 2001:db8::/64)
		// This is a common practice for IPv6 privacy
		masked := make(net.IP, 16)
		copy(masked, ipv6)
		// Set the last 8 bytes to zero
		for i := 8; i < 16; i++ {
			masked[i] = 0
		}
		return fmt.Sprintf("%s/64", masked.String()), nil
	}

	return "", fmt.Errorf("unsupported IP address format: %s", ipAddress)
}

// IsPrivateIP checks if an IP address is in a private range
func IsPrivateIP(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// FormatIPForDisplay formats an IP address for display, handling both IPv4 and IPv6
func FormatIPForDisplay(ipAddress string) string {
	if ipAddress == "" {
		return "Unknown"
	}

	// Check if it's already in CIDR format
	if strings.Contains(ipAddress, "/") {
		return ipAddress
	}

	// Try to anonymize the IP
	anonymized, err := AnonymizeIP(ipAddress)
	if err != nil {
		// If anonymization fails, return the original IP
		return ipAddress
	}

	return anonymized
}




