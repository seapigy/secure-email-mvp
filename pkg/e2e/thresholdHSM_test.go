package e2e

import (
	"testing"
	"time"
)

func TestNewThresholdHSM(t *testing.T) {
	config := HSMConfig{
		Enabled:            true,
		ThresholdM:         3,
		ThresholdN:         5,
		HSMType:            "software",
		KeyRotationEnabled: true,
	}

	hsm := NewThresholdHSM(config)
	if hsm == nil {
		t.Fatal("NewThresholdHSM() returned nil")
	}

	if hsm.config.ThresholdM != config.ThresholdM {
		t.Errorf("Config ThresholdM = %v, want %v", hsm.config.ThresholdM, config.ThresholdM)
	}

	if hsm.config.ThresholdN != config.ThresholdN {
		t.Errorf("Config ThresholdN = %v, want %v", hsm.config.ThresholdN, config.ThresholdN)
	}

	if hsm.shares == nil {
		t.Error("HSM shares map is nil")
	}
}

func TestThresholdHSM_GenerateThresholdKey(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 3,
		ThresholdN: 5,
	}
	hsm := NewThresholdHSM(config)

	tests := []struct {
		name    string
		keyType string
		wantErr bool
	}{
		{"Valid dilithium3", "dilithium3", false},
		{"Valid dilithium5", "dilithium5", false},
		{"Valid falcon512", "falcon512", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := hsm.GenerateThresholdKey(tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateThresholdKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if key == nil {
					t.Error("GenerateThresholdKey() returned nil key")
					return
				}

				if key.ID == "" {
					t.Error("Threshold key ID is empty")
				}

				if key.KeyType != tt.keyType {
					t.Errorf("Key type = %v, want %v", key.KeyType, tt.keyType)
				}

				if key.Threshold != config.ThresholdM {
					t.Errorf("Key threshold = %v, want %v", key.Threshold, config.ThresholdM)
				}

				if key.Total != config.ThresholdN {
					t.Errorf("Key total = %v, want %v", key.Total, config.ThresholdN)
				}

				if len(key.PublicKey) == 0 {
					t.Error("Public key is empty")
				}

				if len(key.Shares) != config.ThresholdN {
					t.Errorf("Number of shares = %v, want %v", len(key.Shares), config.ThresholdN)
				}

				if key.Status != "active" {
					t.Errorf("Key status = %v, want active", key.Status)
				}

				if key.CreatedAt.IsZero() {
					t.Error("Key CreatedAt is zero")
				}

				if key.RotationID == "" {
					t.Error("Key RotationID is empty")
				}

				// Verify shares
				for i := 1; i <= config.ThresholdN; i++ {
					share, exists := key.Shares[i]
					if !exists {
						t.Errorf("Share %d does not exist", i)
						continue
					}

					if share.ShareID != i {
						t.Errorf("Share %d has wrong ShareID = %v", i, share.ShareID)
					}

					if share.KeyID != key.ID {
						t.Errorf("Share %d has wrong KeyID = %v", i, share.KeyID)
					}

					if len(share.ShareData) == 0 {
						t.Errorf("Share %d has empty ShareData", i)
					}

					if share.Threshold != config.ThresholdM {
						t.Errorf("Share %d threshold = %v, want %v", i, share.Threshold, config.ThresholdM)
					}

					if share.Total != config.ThresholdN {
						t.Errorf("Share %d total = %v, want %v", i, share.Total, config.ThresholdN)
					}
				}
			}
		})
	}
}

func TestThresholdHSM_GenerateThresholdKey_Disabled(t *testing.T) {
	config := HSMConfig{
		Enabled: false,
	}
	hsm := NewThresholdHSM(config)

	_, err := hsm.GenerateThresholdKey("dilithium3")
	if err == nil {
		t.Error("GenerateThresholdKey() should fail when HSM is disabled")
	}
}

func TestThresholdHSM_GenerateThresholdKey_InvalidThreshold(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 5,
		ThresholdN: 3, // M > N, should fail
	}
	hsm := NewThresholdHSM(config)

	_, err := hsm.GenerateThresholdKey("dilithium3")
	if err == nil {
		t.Error("GenerateThresholdKey() should fail when M > N")
	}
}

func TestThresholdHSM_Sign(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 3,
		ThresholdN: 5,
	}
	hsm := NewThresholdHSM(config)

	keyID := "test_key_123"
	message := []byte("test message for signing")

	signature, err := hsm.Sign(keyID, message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if signature == nil {
		t.Fatal("Sign() returned nil signature")
	}

	if signature.KeyID != keyID {
		t.Errorf("Signature KeyID = %v, want %v", signature.KeyID, keyID)
	}

	if len(signature.Message) != len(message) {
		t.Errorf("Signature message length = %v, want %v", len(signature.Message), len(message))
	}

	if len(signature.Shares) != config.ThresholdM {
		t.Errorf("Number of signature shares = %v, want %v", len(signature.Shares), config.ThresholdM)
	}

	if len(signature.Signature) == 0 {
		t.Error("Combined signature is empty")
	}

	if !signature.Valid {
		t.Error("Signature should be valid")
	}

	if signature.Timestamp.IsZero() {
		t.Error("Signature timestamp is zero")
	}

	// Verify signature shares
	for i, share := range signature.Shares {
		if share.ShareID == 0 {
			t.Errorf("Share %d has zero ShareID", i)
		}

		if share.NodeID == "" {
			t.Errorf("Share %d has empty NodeID", i)
		}

		if len(share.Signature) == 0 {
			t.Errorf("Share %d has empty signature", i)
		}

		if share.Timestamp.IsZero() {
			t.Errorf("Share %d has zero timestamp", i)
		}
	}
}

func TestThresholdHSM_Sign_Disabled(t *testing.T) {
	config := HSMConfig{
		Enabled: false,
	}
	hsm := NewThresholdHSM(config)

	_, err := hsm.Sign("test_key", []byte("test"))
	if err == nil {
		t.Error("Sign() should fail when HSM is disabled")
	}
}

func TestThresholdHSM_Verify(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 3,
		ThresholdN: 5,
	}
	hsm := NewThresholdHSM(config)

	// Create a test signature
	signature := &ThresholdSignature{
		KeyID:   "test_key",
		Message: []byte("test message"),
		Shares: []SignatureShare{
			{ShareID: 1, NodeID: "node1", Signature: []byte("sig1"), Timestamp: time.Now()},
			{ShareID: 2, NodeID: "node2", Signature: []byte("sig2"), Timestamp: time.Now()},
			{ShareID: 3, NodeID: "node3", Signature: []byte("sig3"), Timestamp: time.Now()},
		},
		Signature: []byte("combined_signature"),
		Timestamp: time.Now(),
		Valid:     true,
	}

	publicKey := []byte("test_public_key")

	valid, err := hsm.Verify(signature, publicKey)
	if err != nil {
		t.Errorf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Verify() should return valid for properly constructed signature")
	}
}

func TestThresholdHSM_Verify_InsufficientShares(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 3,
		ThresholdN: 5,
	}
	hsm := NewThresholdHSM(config)

	// Create signature with insufficient shares
	signature := &ThresholdSignature{
		KeyID:   "test_key",
		Message: []byte("test message"),
		Shares: []SignatureShare{
			{ShareID: 1, NodeID: "node1", Signature: []byte("sig1"), Timestamp: time.Now()},
		}, // Only 1 share, need 3
		Signature: []byte("combined_signature"),
		Timestamp: time.Now(),
		Valid:     true,
	}

	publicKey := []byte("test_public_key")

	_, err := hsm.Verify(signature, publicKey)
	if err == nil {
		t.Error("Verify() should fail with insufficient shares")
	}
}

func TestThresholdHSM_RotateKey(t *testing.T) {
	config := HSMConfig{
		Enabled:            true,
		KeyRotationEnabled: true,
		ThresholdM:         3,
		ThresholdN:         5,
	}
	hsm := NewThresholdHSM(config)

	event, err := hsm.RotateKey("old_key_123", "scheduled_rotation", "admin")
	if err != nil {
		t.Fatalf("RotateKey() error = %v", err)
	}

	if event == nil {
		t.Fatal("RotateKey() returned nil event")
	}

	if event.ID == "" {
		t.Error("Rotation event ID is empty")
	}

	if event.OldKeyID != "old_key_123" {
		t.Errorf("Old key ID = %v, want old_key_123", event.OldKeyID)
	}

	if event.Reason != "scheduled_rotation" {
		t.Errorf("Reason = %v, want scheduled_rotation", event.Reason)
	}

	if event.InitiatedBy != "admin" {
		t.Errorf("InitiatedBy = %v, want admin", event.InitiatedBy)
	}

	if event.Status != "pending" {
		t.Errorf("Status = %v, want pending", event.Status)
	}

	if len(event.Participants) != config.ThresholdN {
		t.Errorf("Number of participants = %v, want %v", len(event.Participants), config.ThresholdN)
	}

	if event.Timestamp.IsZero() {
		t.Error("Rotation event timestamp is zero")
	}
}

func TestThresholdHSM_RotateKey_Disabled(t *testing.T) {
	config := HSMConfig{
		Enabled:            true,
		KeyRotationEnabled: false,
	}
	hsm := NewThresholdHSM(config)

	_, err := hsm.RotateKey("test_key", "test", "admin")
	if err == nil {
		t.Error("RotateKey() should fail when rotation is disabled")
	}
}

func TestThresholdHSM_GetKeyStatus(t *testing.T) {
	config := HSMConfig{Enabled: true}
	hsm := NewThresholdHSM(config)

	status, err := hsm.GetKeyStatus("test_key")
	if err != nil {
		t.Errorf("GetKeyStatus() error = %v", err)
	}

	if status == "" {
		t.Error("GetKeyStatus() returned empty status")
	}
}

func TestThresholdHSM_ListActiveKeys(t *testing.T) {
	config := HSMConfig{Enabled: true}
	hsm := NewThresholdHSM(config)

	keys, err := hsm.ListActiveKeys()
	if err != nil {
		t.Errorf("ListActiveKeys() error = %v", err)
	}

	if keys == nil {
		t.Error("ListActiveKeys() returned nil")
	}
}

func TestThresholdHSM_RevokeKey(t *testing.T) {
	config := HSMConfig{Enabled: true}
	hsm := NewThresholdHSM(config)

	err := hsm.RevokeKey("test_key", "compromised")
	if err != nil {
		t.Errorf("RevokeKey() error = %v", err)
	}
}

func TestThresholdHSM_ValidateThresholdParams(t *testing.T) {
	config := HSMConfig{Enabled: true}
	hsm := NewThresholdHSM(config)

	tests := []struct {
		name    string
		m       int
		n       int
		wantErr bool
	}{
		{"Valid 3 of 5", 3, 5, false},
		{"Valid 2 of 3", 2, 3, false},
		{"Valid 1 of 1", 1, 1, false},
		{"Invalid M > N", 5, 3, true},
		{"Invalid zero M", 0, 5, true},
		{"Invalid zero N", 3, 0, true},
		{"Invalid negative M", -1, 5, true},
		{"Invalid large N", 5, 15, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hsm.ValidateThresholdParams(tt.m, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThresholdParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThresholdHSM_ValidateKeyType(t *testing.T) {
	config := HSMConfig{Enabled: true}
	hsm := NewThresholdHSM(config)

	tests := []struct {
		name    string
		keyType string
		wantErr bool
	}{
		{"Valid dilithium2", "dilithium2", false},
		{"Valid dilithium3", "dilithium3", false},
		{"Valid dilithium5", "dilithium5", false},
		{"Valid falcon512", "falcon512", false},
		{"Valid falcon1024", "falcon1024", false},
		{"Invalid kyber", "kyber768", true},
		{"Invalid empty", "", true},
		{"Invalid random", "invalid_type", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hsm.ValidateKeyType(tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKeyType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThresholdHSM_HelperFunctions(t *testing.T) {
	config := HSMConfig{
		Enabled:    true,
		ThresholdM: 3,
		ThresholdN: 5,
	}
	hsm := NewThresholdHSM(config)

	// Test ID generation
	keyID := hsm.generateKeyID()
	if keyID == "" {
		t.Error("generateKeyID() returned empty string")
	}

	shareID := hsm.generateShareID()
	if shareID == "" {
		t.Error("generateShareID() returned empty string")
	}

	rotationID := hsm.generateRotationID()
	if rotationID == "" {
		t.Error("generateRotationID() returned empty string")
	}

	// Test key expiry
	expiry := hsm.calculateKeyExpiry()
	if expiry == nil {
		t.Error("calculateKeyExpiry() returned nil")
		return
	}

	if expiry.Before(time.Now()) {
		t.Error("calculateKeyExpiry() should return future time")
	}

	// Test active nodes
	nodes := hsm.getActiveNodes()
	if len(nodes) != config.ThresholdN {
		t.Errorf("getActiveNodes() returned %d nodes, want %d", len(nodes), config.ThresholdN)
	}

	// Test key derivation
	masterKey := []byte("test_master_key")
	publicKey, err := hsm.derivePublicKey(masterKey, "dilithium3")
	if err != nil {
		t.Errorf("derivePublicKey() error = %v", err)
	}

	if len(publicKey) == 0 {
		t.Error("derivePublicKey() returned empty key")
	}

	// Test different keys produce different results
	publicKey2, err := hsm.derivePublicKey([]byte("different_key"), "dilithium3")
	if err != nil {
		t.Errorf("derivePublicKey() error = %v", err)
	}

	if string(publicKey) == string(publicKey2) {
		t.Error("derivePublicKey() should produce different keys for different inputs")
	}
}
