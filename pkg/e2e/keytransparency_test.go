package e2e

import (
	"testing"
	"time"
)

func TestNewKeyTransparency(t *testing.T) {
	config := KTConfig{
		Enabled:      true,
		LogURL:       "https://kt.example.com",
		VerifyProofs: true,
		TreeHeight:   20,
	}

	kt := NewKeyTransparency(config)
	if kt == nil {
		t.Fatal("NewKeyTransparency() returned nil")
	}

	if kt.config.Enabled != config.Enabled {
		t.Errorf("Config enabled = %v, want %v", kt.config.Enabled, config.Enabled)
	}
}

func TestKeyTransparency_RegisterPublicKey(t *testing.T) {
	config := KTConfig{
		Enabled:      true,
		VerifyProofs: true,
	}
	kt := NewKeyTransparency(config)

	tests := []struct {
		name      string
		userUUID  string
		publicKey string
		keyType   string
		wantErr   bool
	}{
		{
			name:      "Valid kyber768 key",
			userUUID:  "user123",
			publicKey: "dGVzdF9wdWJsaWNfa2V5",
			keyType:   "kyber768",
			wantErr:   false,
		},
		{
			name:      "Valid dilithium3 key",
			userUUID:  "user456",
			publicKey: "dGVzdF9zaWduaW5nX2tleQ==",
			keyType:   "dilithium3",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := kt.RegisterPublicKey(tt.userUUID, tt.publicKey, tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if entry == nil {
					t.Error("RegisterPublicKey() returned nil entry")
					return
				}

				if entry.UserUUID != tt.userUUID {
					t.Errorf("Entry UserUUID = %v, want %v", entry.UserUUID, tt.userUUID)
				}

				if entry.PublicKey != tt.publicKey {
					t.Errorf("Entry PublicKey = %v, want %v", entry.PublicKey, tt.publicKey)
				}

				if entry.KeyType != tt.keyType {
					t.Errorf("Entry KeyType = %v, want %v", entry.KeyType, tt.keyType)
				}

				if entry.ID == "" {
					t.Error("Entry ID is empty")
				}

				if entry.EntryHash == "" {
					t.Error("Entry EntryHash is empty")
				}

				if entry.MerkleProof == "" {
					t.Error("Entry MerkleProof is empty")
				}

				if entry.CreatedAt.IsZero() {
					t.Error("Entry CreatedAt is zero")
				}
			}
		})
	}
}

func TestKeyTransparency_RegisterPublicKey_Disabled(t *testing.T) {
	config := KTConfig{
		Enabled: false,
	}
	kt := NewKeyTransparency(config)

	_, err := kt.RegisterPublicKey("user123", "test_key", "kyber768")
	if err == nil {
		t.Error("RegisterPublicKey() should fail when KT is disabled")
	}
}

func TestKeyTransparency_VerifyPublicKey(t *testing.T) {
	config := KTConfig{
		Enabled:      true,
		VerifyProofs: true,
	}
	kt := NewKeyTransparency(config)

	// Test verification
	result, err := kt.VerifyPublicKey("user123", "test_key", "kyber768")
	if err != nil {
		t.Fatalf("VerifyPublicKey() error = %v", err)
	}

	if result == nil {
		t.Fatal("VerifyPublicKey() returned nil result")
	}

	if !result.Valid {
		t.Error("VerifyPublicKey() result should be valid")
	}

	if result.EntryHash == "" {
		t.Error("VerifyPublicKey() result EntryHash is empty")
	}

	if result.TreeHead == "" {
		t.Error("VerifyPublicKey() result TreeHead is empty")
	}

	if result.Timestamp.IsZero() {
		t.Error("VerifyPublicKey() result Timestamp is zero")
	}
}

func TestKeyTransparency_VerifyPublicKey_Disabled(t *testing.T) {
	config := KTConfig{
		Enabled: false,
	}
	kt := NewKeyTransparency(config)

	result, err := kt.VerifyPublicKey("user123", "test_key", "kyber768")
	if err != nil {
		t.Fatalf("VerifyPublicKey() error = %v", err)
	}

	if result.Valid {
		t.Error("VerifyPublicKey() should return invalid when KT is disabled")
	}

	if result.ErrorMsg == "" {
		t.Error("VerifyPublicKey() should have error message when disabled")
	}
}

func TestKeyTransparency_AuditLog(t *testing.T) {
	config := KTConfig{
		Enabled:      true,
		VerifyProofs: true,
	}
	kt := NewKeyTransparency(config)

	results, err := kt.AuditLog(1, 5)
	if err != nil {
		t.Fatalf("AuditLog() error = %v", err)
	}

	if len(results) != 5 {
		t.Errorf("AuditLog() returned %d results, want 5", len(results))
	}

	for i, result := range results {
		if !result.Valid {
			t.Errorf("AuditLog() result %d should be valid", i)
		}

		if result.EntryHash == "" {
			t.Errorf("AuditLog() result %d EntryHash is empty", i)
		}

		if result.TreeHead == "" {
			t.Errorf("AuditLog() result %d TreeHead is empty", i)
		}
	}
}

func TestKeyTransparency_RevokePublicKey(t *testing.T) {
	config := KTConfig{
		Enabled: true,
	}
	kt := NewKeyTransparency(config)

	err := kt.RevokePublicKey("user123", "kyber768")
	if err != nil {
		t.Errorf("RevokePublicKey() error = %v", err)
	}
}

func TestKeyTransparency_GetPublicKeys(t *testing.T) {
	config := KTConfig{
		Enabled: true,
	}
	kt := NewKeyTransparency(config)

	keys, err := kt.GetPublicKeys("user123")
	if err != nil {
		t.Errorf("GetPublicKeys() error = %v", err)
	}

	if keys == nil {
		t.Error("GetPublicKeys() returned nil")
	}
}

func TestKeyTransparency_ValidatePublicKey(t *testing.T) {
	config := KTConfig{Enabled: true}
	kt := NewKeyTransparency(config)

	tests := []struct {
		name      string
		publicKey string
		keyType   string
		wantErr   bool
	}{
		{"Valid kyber768", "test_key", "kyber768", false},
		{"Valid dilithium3", "test_key", "dilithium3", false},
		{"Empty key", "", "kyber768", true},
		{"Empty type", "test_key", "", true},
		{"Invalid type", "test_key", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kt.ValidatePublicKey(tt.publicKey, tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeyTransparency_ValidateUserUUID(t *testing.T) {
	config := KTConfig{Enabled: true}
	kt := NewKeyTransparency(config)

	tests := []struct {
		name     string
		userUUID string
		wantErr  bool
	}{
		{"Valid UUID", "user123456", false},
		{"Short UUID", "user123", true},
		{"Empty UUID", "", true},
		{"Too short", "user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := kt.ValidateUserUUID(tt.userUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeyTransparency_HelperFunctions(t *testing.T) {
	config := KTConfig{Enabled: true}
	kt := NewKeyTransparency(config)

	// Test ID generation
	entryID := kt.generateEntryID()
	if entryID == "" {
		t.Error("generateEntryID() returned empty string")
	}

	logID := kt.generateLogEntryID()
	if logID == "" {
		t.Error("generateLogEntryID() returned empty string")
	}

	// Test hashing
	hash1 := kt.hashString("test")
	hash2 := kt.hashString("test")
	if hash1 != hash2 {
		t.Error("hashString() should be deterministic")
	}

	hash3 := kt.hashString("different")
	if hash1 == hash3 {
		t.Error("hashString() should produce different hashes for different inputs")
	}

	// Test Merkle root calculation
	root1 := kt.calculateMerkleRoot(1)
	root2 := kt.calculateMerkleRoot(2)
	if root1 == root2 {
		t.Error("calculateMerkleRoot() should produce different roots for different indices")
	}

	// Test key expiry calculation
	expiry := kt.calculateKeyExpiry()
	if expiry == nil {
		t.Error("calculateKeyExpiry() returned nil")
		return
	}

	if expiry.Before(time.Now()) {
		t.Error("calculateKeyExpiry() should return future time")
	}
}
