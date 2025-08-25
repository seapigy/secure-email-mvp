package audit

import (
	"testing"
)

func TestAnonymizeIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "IPv4 address",
			input:    "192.168.1.100",
			expected: "192.168.1.0/24",
			wantErr:  false,
		},
		{
			name:     "IPv4 address with different octets",
			input:    "10.0.0.1",
			expected: "10.0.0.0/24",
			wantErr:  false,
		},
		{
			name:     "IPv6 address",
			input:    "2001:db8::1",
			expected: "2001:db8::/64",
			wantErr:  false,
		},
		{
			name:     "IPv6 address with more segments",
			input:    "2001:db8:1:2:3:4:5:6",
			expected: "2001:db8:1:2::/64",
			wantErr:  false,
		},
		{
			name:     "Empty IP address",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid IP address",
			input:    "invalid-ip",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AnonymizeIP(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnonymizeIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("AnonymizeIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Private IPv4 - Class A",
			input:    "10.0.0.1",
			expected: true,
		},
		{
			name:     "Private IPv4 - Class B",
			input:    "172.16.0.1",
			expected: true,
		},
		{
			name:     "Private IPv4 - Class C",
			input:    "192.168.0.1",
			expected: true,
		},
		{
			name:     "Loopback IPv4",
			input:    "127.0.0.1",
			expected: true,
		},
		{
			name:     "Link-local IPv4",
			input:    "169.254.0.1",
			expected: true,
		},
		{
			name:     "Public IPv4",
			input:    "8.8.8.8",
			expected: false,
		},
		{
			name:     "IPv6 loopback",
			input:    "::1",
			expected: true,
		},
		{
			name:     "IPv6 link-local",
			input:    "fe80::1",
			expected: true,
		},
		{
			name:     "Public IPv6",
			input:    "2001:4860:4860::8888",
			expected: false,
		},
		{
			name:     "Invalid IP",
			input:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrivateIP(tt.input)
			if result != tt.expected {
				t.Errorf("IsPrivateIP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatIPForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 address",
			input:    "192.168.1.100",
			expected: "192.168.1.0/24",
		},
		{
			name:     "IPv6 address",
			input:    "2001:db8::1",
			expected: "2001:db8::/64",
		},
		{
			name:     "Already anonymized IPv4",
			input:    "192.168.1.0/24",
			expected: "192.168.1.0/24",
		},
		{
			name:     "Empty IP",
			input:    "",
			expected: "Unknown",
		},
		{
			name:     "Invalid IP",
			input:    "invalid-ip",
			expected: "invalid-ip", // Should return original if anonymization fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIPForDisplay(tt.input)
			if result != tt.expected {
				t.Errorf("FormatIPForDisplay() = %v, want %v", result, tt.expected)
			}
		})
	}
}









