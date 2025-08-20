package e2e

import (
	"testing"
	"time"
)

func TestNewMetadataMinimizer(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	if mm == nil {
		t.Fatal("NewMetadataMinimizer() returned nil")
	}

	if mm.config.Enabled != config.Enabled {
		t.Errorf("Config enabled = %v, want %v", mm.config.Enabled, config.Enabled)
	}
}

func TestMetadataMinimizer_CreatePrivacyPolicy(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	tests := []struct {
		name    string
		userID  string
		level   string
		wantErr bool
	}{
		{"Minimal level", "user123", "minimal", false},
		{"Standard level", "user456", "standard", false},
		{"Enhanced level", "user789", "enhanced", false},
		{"Invalid level", "user000", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := mm.CreatePrivacyPolicy(tt.userID, tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePrivacyPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if policy == nil {
					t.Error("CreatePrivacyPolicy() returned nil policy")
					return
				}

				if policy.UserID != tt.userID {
					t.Errorf("Policy UserID = %v, want %v", policy.UserID, tt.userID)
				}

				if policy.CreatedAt.IsZero() {
					t.Error("Policy CreatedAt is zero")
				}

				if policy.UpdatedAt.IsZero() {
					t.Error("Policy UpdatedAt is zero")
				}

				// Check level-specific settings
				switch tt.level {
				case "minimal":
					if policy.MinimizeTimestamps {
						t.Error("Minimal level should not minimize timestamps")
					}
					if policy.AnonymizeParticipants {
						t.Error("Minimal level should not anonymize participants")
					}
					if policy.PaddingStrategy != "none" {
						t.Errorf("Minimal level padding strategy = %v, want none", policy.PaddingStrategy)
					}

				case "standard":
					if !policy.MinimizeTimestamps {
						t.Error("Standard level should minimize timestamps")
					}
					if !policy.UseTimeBatching {
						t.Error("Standard level should use time batching")
					}
					if policy.PaddingStrategy != "standard" {
						t.Errorf("Standard level padding strategy = %v, want standard", policy.PaddingStrategy)
					}

				case "enhanced":
					if !policy.MinimizeTimestamps {
						t.Error("Enhanced level should minimize timestamps")
					}
					if !policy.AnonymizeParticipants {
						t.Error("Enhanced level should anonymize participants")
					}
					if policy.PaddingStrategy != "aggressive" {
						t.Errorf("Enhanced level padding strategy = %v, want aggressive", policy.PaddingStrategy)
					}
				}
			}
		})
	}
}

func TestMetadataMinimizer_MinimizeMetadata(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	// Create a standard privacy policy
	policy, err := mm.CreatePrivacyPolicy("sender123", "standard")
	if err != nil {
		t.Fatalf("Failed to create privacy policy: %v", err)
	}

	tests := []struct {
		name        string
		senderID    string
		recipientID string
		threadID    string
		messageSize int
	}{
		{"Small message", "sender123", "recipient456", "thread789", 100},
		{"Medium message", "sender123", "recipient456", "thread789", 5000},
		{"Large message", "sender123", "recipient456", "thread789", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := mm.MinimizeMetadata(
				tt.senderID,
				tt.recipientID,
				tt.threadID,
				tt.messageSize,
				policy,
			)
			if err != nil {
				t.Errorf("MinimizeMetadata() error = %v", err)
				return
			}

			if metadata == nil {
				t.Error("MinimizeMetadata() returned nil metadata")
				return
			}

			// Verify basic fields
			if metadata.Version == "" {
				t.Error("Metadata version is empty")
			}

			if metadata.EncryptionLevel == "" {
				t.Error("Metadata encryption level is empty")
			}

			if metadata.RoutingToken == "" {
				t.Error("Metadata routing token is empty")
			}

			if metadata.DeliveryToken == "" {
				t.Error("Metadata delivery token is empty")
			}

			if metadata.TimeWindow == "" {
				t.Error("Metadata time window is empty")
			}

			if metadata.PaddedSize <= 0 {
				t.Error("Metadata padded size should be positive")
			}

			if metadata.ContentClass == "" {
				t.Error("Metadata content class is empty")
			}

			if metadata.TTL <= 0 {
				t.Error("Metadata TTL should be positive")
			}

			// Verify policy-specific behavior
			if policy.MinimizeThreadInfo {
				if metadata.ThreadToken == tt.threadID {
					t.Error("Thread token should be minimized when policy requires it")
				}
				if metadata.SequenceHash == "" {
					t.Error("Sequence hash should be set when minimizing thread info")
				}
			}

			if policy.AnonymizeParticipants {
				if metadata.SenderToken == tt.senderID {
					t.Error("Sender token should be anonymized when policy requires it")
				}
				if metadata.RecipientToken == tt.recipientID {
					t.Error("Recipient token should be anonymized when policy requires it")
				}
			}

			if policy.UseTimeBatching {
				if metadata.BatchID == "" {
					t.Error("Batch ID should be set when using time batching")
				}
			}

			// Verify padding behavior
			if policy.PaddingStrategy != "none" {
				if metadata.PaddedSize < tt.messageSize {
					t.Error("Padded size should not be smaller than original size")
				}
			}
		})
	}
}

func TestMetadataMinimizer_CreateTimeBatch(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	timeWindow := "2024-01-15T10:00:00Z_1h"
	messageIDs := []string{"msg1", "msg2", "msg3"}

	batch, err := mm.CreateTimeBatch(timeWindow, messageIDs)
	if err != nil {
		t.Fatalf("CreateTimeBatch() error = %v", err)
	}

	if batch == nil {
		t.Fatal("CreateTimeBatch() returned nil batch")
	}

	if batch.ID == "" {
		t.Error("Batch ID is empty")
	}

	if batch.TimeWindow != timeWindow {
		t.Errorf("Batch TimeWindow = %v, want %v", batch.TimeWindow, timeWindow)
	}

	if len(batch.MessageIDs) != len(messageIDs) {
		t.Errorf("Batch has %d message IDs, want %d", len(batch.MessageIDs), len(messageIDs))
	}

	for i, id := range messageIDs {
		if batch.MessageIDs[i] != id {
			t.Errorf("Message ID %d = %v, want %v", i, batch.MessageIDs[i], id)
		}
	}

	if batch.BatchSize != len(messageIDs) {
		t.Errorf("Batch size = %v, want %v", batch.BatchSize, len(messageIDs))
	}

	if batch.PaddingCount < 0 {
		t.Error("Padding count should not be negative")
	}

	if batch.CreatedAt.IsZero() {
		t.Error("Batch CreatedAt is zero")
	}
}

func TestMetadataMinimizer_ReleaseBatch(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	err := mm.ReleaseBatch("test_batch_id")
	if err != nil {
		t.Errorf("ReleaseBatch() error = %v", err)
	}
}

func TestMetadataMinimizer_ResolveAnonymousToken(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	resolvedID, err := mm.ResolveAnonymousToken("test_token")
	if err != nil {
		t.Errorf("ResolveAnonymousToken() error = %v", err)
	}

	if resolvedID == "" {
		t.Error("ResolveAnonymousToken() returned empty ID")
	}
}

func TestMetadataMinimizer_PadMessageSize(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	tests := []struct {
		name     string
		size     int
		strategy string
		wantMin  int
		wantMax  int
	}{
		{"None strategy", 100, "none", 100, 100},
		{"Standard small", 100, "standard", 1024, 1024},
		{"Standard medium", 5000, "standard", 16384, 16384},
		{"Standard large", 100000, "standard", 262144, 262144},
		{"Aggressive small", 100, "aggressive", 4096, 4096},
		{"Aggressive medium", 5000, "aggressive", 8192, 8192},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := mm.padMessageSize(tt.size, tt.strategy)

			if padded < tt.wantMin {
				t.Errorf("Padded size %d < minimum %d", padded, tt.wantMin)
			}

			if padded > tt.wantMax && tt.wantMax > 0 {
				t.Errorf("Padded size %d > maximum %d", padded, tt.wantMax)
			}

			if tt.strategy != "none" && padded < tt.size {
				t.Errorf("Padded size %d should not be smaller than original %d", padded, tt.size)
			}
		})
	}
}

func TestMetadataMinimizer_ClassifyContent(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	tests := []struct {
		name string
		size int
		want string
	}{
		{"Text message", 100, "text"},
		{"Small media", 5000, "media"},
		{"Large media", 500000, "media"},
		{"File", 2000000, "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class := mm.classifyContent(tt.size)
			if class != tt.want {
				t.Errorf("classifyContent() = %v, want %v", class, tt.want)
			}
		})
	}
}

func TestMetadataMinimizer_ValidatePrivacyPolicy(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	tests := []struct {
		name    string
		policy  *PrivacyPolicy
		wantErr bool
	}{
		{
			name: "Valid policy",
			policy: &PrivacyPolicy{
				UserID:          "user123",
				BatchingWindow:  15 * time.Minute,
				PaddingStrategy: "standard",
			},
			wantErr: false,
		},
		{
			name: "Empty user ID",
			policy: &PrivacyPolicy{
				UserID:          "",
				BatchingWindow:  15 * time.Minute,
				PaddingStrategy: "standard",
			},
			wantErr: true,
		},
		{
			name: "Negative batching window",
			policy: &PrivacyPolicy{
				UserID:          "user123",
				BatchingWindow:  -1 * time.Minute,
				PaddingStrategy: "standard",
			},
			wantErr: true,
		},
		{
			name: "Too long batching window",
			policy: &PrivacyPolicy{
				UserID:          "user123",
				BatchingWindow:  25 * time.Hour,
				PaddingStrategy: "standard",
			},
			wantErr: true,
		},
		{
			name: "Invalid padding strategy",
			policy: &PrivacyPolicy{
				UserID:          "user123",
				BatchingWindow:  15 * time.Minute,
				PaddingStrategy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mm.ValidatePrivacyPolicy(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrivacyPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadataMinimizer_GetPaddingProfile(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	tests := []struct {
		name      string
		strategy  string
		wantSizes int
	}{
		{"None strategy", "none", 0},
		{"Standard strategy", "standard", 5},
		{"Aggressive strategy", "aggressive", 7},
		{"Invalid strategy", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := mm.GetPaddingProfile(tt.strategy)
			if profile == nil {
				t.Error("GetPaddingProfile() returned nil")
				return
			}

			if profile.Strategy != tt.strategy && tt.strategy != "invalid" {
				t.Errorf("Profile strategy = %v, want %v", profile.Strategy, tt.strategy)
			}

			if len(profile.StandardSizes) != tt.wantSizes {
				t.Errorf("Profile has %d standard sizes, want %d", len(profile.StandardSizes), tt.wantSizes)
			}

			// Verify strategy-specific properties
			switch tt.strategy {
			case "standard":
				if profile.RandomPadding {
					t.Error("Standard strategy should not use random padding")
				}
			case "aggressive":
				if !profile.RandomPadding {
					t.Error("Aggressive strategy should use random padding")
				}
				if profile.MinPadding == 0 {
					t.Error("Aggressive strategy should have non-zero min padding")
				}
			case "none":
				if profile.MaxPadding != 0 {
					t.Error("None strategy should have zero max padding")
				}
			}
		})
	}
}

func TestMetadataMinimizer_HelperFunctions(t *testing.T) {
	config := mustGetTestConfig(t)
	mm := NewMetadataMinimizer(*config)

	// Test timestamp obfuscation
	timestamp := time.Now()
	window := time.Hour
	obfuscated := mm.obfuscateTimestamp(timestamp, window)
	if obfuscated == "" {
		t.Error("obfuscateTimestamp() returned empty string")
	}

	// Test batch ID generation
	batchID1 := mm.generateBatchID("window1")
	batchID2 := mm.generateBatchID("window1")
	if batchID1 != batchID2 {
		t.Error("generateBatchID() should be deterministic for same input")
	}

	batchID3 := mm.generateBatchID("window2")
	if batchID1 == batchID3 {
		t.Error("generateBatchID() should produce different IDs for different inputs")
	}

	// Test nonce generation
	nonce1 := mm.generateNonce()
	nonce2 := mm.generateNonce()
	if nonce1 == "" || nonce2 == "" {
		t.Error("generateNonce() returned empty string")
	}
	if nonce1 == nonce2 {
		t.Error("generateNonce() should produce different nonces")
	}

	// Test routing token generation
	token, err := mm.generateRoutingToken("recipient123")
	if err != nil {
		t.Errorf("generateRoutingToken() error = %v", err)
	}
	if token == "" {
		t.Error("generateRoutingToken() returned empty token")
	}

	// Test delivery token generation
	deliveryToken, err := mm.generateDeliveryToken()
	if err != nil {
		t.Errorf("generateDeliveryToken() error = %v", err)
	}
	if deliveryToken == "" {
		t.Error("generateDeliveryToken() returned empty token")
	}

	// Test thread token generation
	threadToken, err := mm.generateThreadToken("thread123")
	if err != nil {
		t.Errorf("generateThreadToken() error = %v", err)
	}
	if threadToken == "" {
		t.Error("generateThreadToken() returned empty token")
	}

	// Test sequence hash
	sequenceHash := mm.hashSequence("thread123", timestamp)
	if sequenceHash == "" {
		t.Error("hashSequence() returned empty hash")
	}

	// Test anonymous token generation
	anonymousToken, err := mm.generateAnonymousToken("user123", "sender")
	if err != nil {
		t.Errorf("generateAnonymousToken() error = %v", err)
	}
	if anonymousToken == "" {
		t.Error("generateAnonymousToken() returned empty token")
	}

	// Test padding count calculation
	paddingCount := mm.calculatePaddingCount(10)
	if paddingCount < 0 {
		t.Error("calculatePaddingCount() returned negative count")
	}
	if paddingCount > 5 { // Should be 10-50% of 10
		t.Errorf("calculatePaddingCount() returned too high count: %d", paddingCount)
	}
}
