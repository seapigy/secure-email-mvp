package email

import (
	"testing"
	"time"
)

func TestValidateEmailSecurityToggles(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid empty toggles",
			toggles: EmailSecurityToggles{},
			wantErr: false,
		},
		{
			name: "Valid time window",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(time.Now().Unix()),
				ExpiresAt: int64Ptr(time.Now().Add(time.Hour).Unix()),
			},
			wantErr: false,
		},
		{
			name: "Invalid time window - not_before after expires_at",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(time.Now().Add(time.Hour).Unix()),
				ExpiresAt: int64Ptr(time.Now().Unix()),
			},
			wantErr: true,
			errMsg:  "not_before must be before expires_at",
		},
		{
			name: "Invalid time window - not_before equals expires_at",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(time.Now().Unix()),
				ExpiresAt: int64Ptr(time.Now().Unix()),
			},
			wantErr: true,
			errMsg:  "not_before must be before expires_at",
		},
		{
			name: "Valid self-destruct threshold",
			toggles: EmailSecurityToggles{
				SelfDestructThreshold: intPtr(5),
			},
			wantErr: false,
		},
		{
			name: "Invalid self-destruct threshold - too low",
			toggles: EmailSecurityToggles{
				SelfDestructThreshold: intPtr(0),
			},
			wantErr: true,
			errMsg:  "self_destruct_threshold must be at least 1",
		},
		{
			name: "Invalid self-destruct threshold - too high",
			toggles: EmailSecurityToggles{
				SelfDestructThreshold: intPtr(101),
			},
			wantErr: true,
			errMsg:  "self_destruct_threshold cannot exceed 100",
		},
		{
			name: "Valid geo rules JSON",
			toggles: EmailSecurityToggles{
				GeoRulesRef: stringPtr(`{"type": "circle", "lat": 40.7128, "lng": -74.0060, "radius": 1000}`),
			},
			wantErr: false,
		},
		{
			name: "Invalid geo rules JSON",
			toggles: EmailSecurityToggles{
				GeoRulesRef: stringPtr(`{"type": "circle", "lat": 40.7128, "lng": -74.0060, "radius": 1000`),
			},
			wantErr: true,
			errMsg:  "geo_rules_ref must be valid JSON: unexpected end of JSON input",
		},
		{
			name: "Empty geo rules JSON",
			toggles: EmailSecurityToggles{
				GeoRulesRef: stringPtr(""),
			},
			wantErr: false,
		},
		{
			name: "Empty decoy secret",
			toggles: EmailSecurityToggles{
				DecoySecret: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "decoy_secret cannot be empty if provided",
		},
		{
			name: "Valid decoy secret",
			toggles: EmailSecurityToggles{
				DecoySecret: stringPtr("valid-secret-hash"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmailSecurityToggles(tt.toggles)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateEmailSecurityToggles() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ValidateEmailSecurityToggles() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateEmailSecurityToggles() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestEmailSecurityToggles_IsTimeLocked(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "No time lock",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Time locked - future",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(future.Unix()),
			},
			want: true,
		},
		{
			name: "Not time locked - past",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(past.Unix()),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.IsTimeLocked(); got != tt.want {
				t.Errorf("EmailSecurityToggles.IsTimeLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_IsExpired(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "No expiration",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Not expired - future",
			toggles: EmailSecurityToggles{
				ExpiresAt: int64Ptr(future.Unix()),
			},
			want: false,
		},
		{
			name: "Expired - past",
			toggles: EmailSecurityToggles{
				ExpiresAt: int64Ptr(past.Unix()),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.IsExpired(); got != tt.want {
				t.Errorf("EmailSecurityToggles.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_IsRevoked(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "Not revoked",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Revoked",
			toggles: EmailSecurityToggles{
				RemoteRevoke: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.IsRevoked(); got != tt.want {
				t.Errorf("EmailSecurityToggles.IsRevoked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_GetSelfDestructThreshold(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    int
	}{
		{
			name:    "Default threshold",
			toggles: EmailSecurityToggles{},
			want:    3,
		},
		{
			name: "Custom threshold",
			toggles: EmailSecurityToggles{
				SelfDestructThreshold: intPtr(5),
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.GetSelfDestructThreshold(); got != tt.want {
				t.Errorf("EmailSecurityToggles.GetSelfDestructThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_RequiresMFA(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "No MFA required",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "MFA required",
			toggles: EmailSecurityToggles{
				MFAOnOpen: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.RequiresMFA(); got != tt.want {
				t.Errorf("EmailSecurityToggles.RequiresMFA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_IsReadOnce(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "Not read once",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Read once",
			toggles: EmailSecurityToggles{
				ReadOnce: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.IsReadOnce(); got != tt.want {
				t.Errorf("EmailSecurityToggles.IsReadOnce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_ShouldStripMetadata(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "Don't strip metadata",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Strip metadata",
			toggles: EmailSecurityToggles{
				StripMetadata: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.ShouldStripMetadata(); got != tt.want {
				t.Errorf("EmailSecurityToggles.ShouldStripMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_HasDecoySecret(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "No decoy secret",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Has decoy secret",
			toggles: EmailSecurityToggles{
				DecoySecret: stringPtr("secret-hash"),
			},
			want: true,
		},
		{
			name: "Empty decoy secret",
			toggles: EmailSecurityToggles{
				DecoySecret: stringPtr(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.HasDecoySecret(); got != tt.want {
				t.Errorf("EmailSecurityToggles.HasDecoySecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_HasGeoRules(t *testing.T) {
	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    bool
	}{
		{
			name:    "No geo rules",
			toggles: EmailSecurityToggles{},
			want:    false,
		},
		{
			name: "Has geo rules",
			toggles: EmailSecurityToggles{
				GeoRulesRef: stringPtr(`{"type": "circle"}`),
			},
			want: true,
		},
		{
			name: "Empty geo rules",
			toggles: EmailSecurityToggles{
				GeoRulesRef: stringPtr(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.toggles.HasGeoRules(); got != tt.want {
				t.Errorf("EmailSecurityToggles.HasGeoRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailSecurityToggles_GetTimeWindowStatus(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name    string
		toggles EmailSecurityToggles
		want    string
	}{
		{
			name:    "No time restrictions",
			toggles: EmailSecurityToggles{},
			want:    "No time restrictions",
		},
		{
			name: "Time locked",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(future.Unix()),
			},
			want: "Time-locked until",
		},
		{
			name: "Expired",
			toggles: EmailSecurityToggles{
				ExpiresAt: int64Ptr(past.Unix()),
			},
			want: "Expired at",
		},
		{
			name: "Time window",
			toggles: EmailSecurityToggles{
				NotBefore: int64Ptr(past.Unix()),
				ExpiresAt: int64Ptr(future.Unix()),
			},
			want: "Available from",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.toggles.GetTimeWindowStatus()
			if tt.want != "No time restrictions" && !contains(got, tt.want) {
				t.Errorf("EmailSecurityToggles.GetTimeWindowStatus() = %v, want to contain %v", got, tt.want)
			}
		})
	}
}

// Helper functions
func int64Ptr(v int64) *int64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
