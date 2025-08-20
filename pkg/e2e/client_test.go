package e2e

import (
	"encoding/json"
	"testing"
	"time"
)

// Helper function removed - now using mustGetTestConfig() from config.go

func TestNewClient(t *testing.T) {
	config := *mustGetTestConfig(t)

	// Test successful client creation
	client, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}

	if client.userID != "user123" {
		t.Errorf("Client userID = %v, want user123", client.userID)
	}

	if client.keyPair == nil {
		t.Error("Client keyPair is nil")
	}

	// Test with disabled E2E
	config.Enabled = false
	_, err = NewClient(config, "user123")
	if err == nil {
		t.Error("NewClient() should fail when E2E is disabled")
	}
}

func TestClient_GetPublicKey(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	publicKey := client.GetPublicKey()
	if len(publicKey) == 0 {
		t.Error("GetPublicKey() returned empty public key")
	}

	publicKeyBase64 := client.GetPublicKeyBase64()
	if publicKeyBase64 == "" {
		t.Error("GetPublicKeyBase64() returned empty string")
	}
}

func TestClient_EncryptDecryptMessage(t *testing.T) {
	config := *mustGetTestConfig(t)

	// Create two clients
	alice, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("Failed to create Alice client: %v", err)
	}

	bob, err := NewClient(config, "bob")
	if err != nil {
		t.Fatalf("Failed to create Bob client: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"Empty", ""},
		{"Short", "Hello, Bob!"},
		{"Long", "This is a much longer message that should test the encryption and decryption capabilities of the client SDK. It contains multiple sentences and should be properly encrypted and decrypted."},
		{"SpecialChars", "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"Unicode", "Unicode: 🚀🔐📧✨"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext := []byte(tt.plaintext)

			// Alice encrypts message for Bob
			message, err := alice.EncryptMessage(plaintext, bob.GetPublicKey(), "thread123", "bob")
			if err != nil {
				t.Fatalf("EncryptMessage() error = %v", err)
			}

			// Verify message structure
			if message.ID == "" {
				t.Error("Message ID is empty")
			}
			if message.ThreadID != "thread123" {
				t.Errorf("Message ThreadID = %v, want thread123", message.ThreadID)
			}
			if message.SenderID != "alice" {
				t.Errorf("Message SenderID = %v, want alice", message.SenderID)
			}
			if message.RecipientID != "bob" {
				t.Errorf("Message RecipientID = %v, want bob", message.RecipientID)
			}
			if message.Envelope == nil {
				t.Error("Message Envelope is nil")
			}
			if message.CreatedAt.IsZero() {
				t.Error("Message CreatedAt is zero")
			}

			// Bob decrypts message
			decrypted, err := bob.DecryptMessage(message, alice.GetPublicKey())
			if err != nil {
				t.Fatalf("DecryptMessage() error = %v", err)
			}

			// Verify decrypted content
			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypted content = %v, want %v", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestClient_EncryptDecryptMessage_WrongRecipient(t *testing.T) {
	config := *mustGetTestConfig(t)

	alice, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("Failed to create Alice client: %v", err)
	}

	bob, err := NewClient(config, "bob")
	if err != nil {
		t.Fatalf("Failed to create Bob client: %v", err)
	}

	charlie, err := NewClient(config, "charlie")
	if err != nil {
		t.Fatalf("Failed to create Charlie client: %v", err)
	}

	// Alice encrypts message for Bob
	message, err := alice.EncryptMessage([]byte("Hello, Bob!"), bob.GetPublicKey(), "thread123", "bob")
	if err != nil {
		t.Fatalf("EncryptMessage() error = %v", err)
	}

	// Charlie tries to decrypt message (should fail)
	_, err = charlie.DecryptMessage(message, alice.GetPublicKey())
	if err == nil {
		t.Error("DecryptMessage() should fail for wrong recipient")
	}
}

func TestClient_CreateThread(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	participantIDs := []string{"alice", "bob", "charlie"}
	thread, err := client.CreateThread(participantIDs)
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	if thread.ID == "" {
		t.Error("Thread ID is empty")
	}
	if thread.CreatorID != "alice" {
		t.Errorf("Thread CreatorID = %v, want alice", thread.CreatorID)
	}
	if len(thread.ParticipantIDs) != 3 {
		t.Errorf("Thread has %v participants, want 3", len(thread.ParticipantIDs))
	}
	if len(thread.ThreadKey) == 0 {
		t.Error("Thread key is empty")
	}
	if thread.CreatedAt.IsZero() {
		t.Error("Thread CreatedAt is zero")
	}

	// Verify all participants are included
	for _, expectedID := range participantIDs {
		found := false
		for _, actualID := range thread.ParticipantIDs {
			if actualID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Participant %v not found in thread", expectedID)
		}
	}
}

func TestClient_EncryptDecryptThreadMessage(t *testing.T) {
	config := *mustGetTestConfig(t)

	alice, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("Failed to create Alice client: %v", err)
	}

	bob, err := NewClient(config, "bob")
	if err != nil {
		t.Fatalf("Failed to create Bob client: %v", err)
	}

	// Alice creates a thread
	participantIDs := []string{"alice", "bob"}
	thread, err := alice.CreateThread(participantIDs)
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	// Alice encrypts a thread message
	plaintext := []byte("Hello, thread!")
	message, err := alice.EncryptThreadMessage(plaintext, thread)
	if err != nil {
		t.Fatalf("EncryptThreadMessage() error = %v", err)
	}

	// Verify message structure
	if message.ThreadID != thread.ID {
		t.Errorf("Message ThreadID = %v, want %v", message.ThreadID, thread.ID)
	}
	if message.SenderID != "alice" {
		t.Errorf("Message SenderID = %v, want alice", message.SenderID)
	}
	if message.RecipientID != "" {
		t.Errorf("Message RecipientID = %v, want empty", message.RecipientID)
	}

	// Bob decrypts the thread message
	decrypted, err := bob.DecryptThreadMessage(message, thread)
	if err != nil {
		t.Fatalf("DecryptThreadMessage() error = %v", err)
	}

	// Verify decrypted content
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted content = %v, want %v", string(decrypted), string(plaintext))
	}
}

func TestClient_EncryptThreadMessage_NonParticipant(t *testing.T) {
	config := *mustGetTestConfig(t)

	alice, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("Failed to create Alice client: %v", err)
	}

	charlie, err := NewClient(config, "charlie")
	if err != nil {
		t.Fatalf("Failed to create Charlie client: %v", err)
	}

	// Alice creates a thread without Charlie
	participantIDs := []string{"alice", "bob"}
	thread, err := alice.CreateThread(participantIDs)
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	// Charlie tries to encrypt a thread message (should fail)
	_, err = charlie.EncryptThreadMessage([]byte("Hello!"), thread)
	if err == nil {
		t.Error("EncryptThreadMessage() should fail for non-participant")
	}
}

func TestClient_RotateKeys(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	originalPublicKey := client.GetPublicKey()

	// Rotate keys
	err = client.RotateKeys()
	if err != nil {
		t.Fatalf("RotateKeys() error = %v", err)
	}

	newPublicKey := client.GetPublicKey()

	// Keys should be different
	if string(originalPublicKey) == string(newPublicKey) {
		t.Error("Keys should be different after rotation")
	}
}

func TestClient_ExportImportKeyPair(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	originalPublicKey := client.GetPublicKey()

	// Export key pair
	exportData, err := client.ExportKeyPair()
	if err != nil {
		t.Fatalf("ExportKeyPair() error = %v", err)
	}

	// Verify export data structure
	var export map[string]interface{}
	if err := json.Unmarshal(exportData, &export); err != nil {
		t.Fatalf("Failed to unmarshal export data: %v", err)
	}

	if export["user_id"] != "user123" {
		t.Errorf("Export user_id = %v, want user123", export["user_id"])
	}

	// Create new client and import key pair
	newClient, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = newClient.ImportKeyPair(exportData)
	if err != nil {
		t.Fatalf("ImportKeyPair() error = %v", err)
	}

	// Keys should be the same
	if string(originalPublicKey) != string(newClient.GetPublicKey()) {
		t.Error("Imported key pair should match original")
	}
}

func TestClient_ImportKeyPair_WrongUser(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "alice")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Export key pair for Alice
	exportData, err := client.ExportKeyPair()
	if err != nil {
		t.Fatalf("ExportKeyPair() error = %v", err)
	}

	// Try to import into Bob's client (should fail)
	bob, err := NewClient(config, "bob")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = bob.ImportKeyPair(exportData)
	if err == nil {
		t.Error("ImportKeyPair() should fail for wrong user")
	}
}

func TestClient_GetKeyInfo(t *testing.T) {
	config := *mustGetTestConfig(t)
	client, err := NewClient(config, "user123")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	keyInfo := client.GetKeyInfo()

	// Verify key info structure
	if keyInfo["algorithm"] != config.Crypto.KEMAlgorithm {
		t.Errorf("Key info algorithm = %v, want %v", keyInfo["algorithm"], config.Crypto.KEMAlgorithm)
	}
	if keyInfo["created_at"] == nil {
		t.Error("Key info created_at is nil")
	}
	if keyInfo["public_key"] == "" {
		t.Error("Key info public_key is empty")
	}
	if keyInfo["key_length"] == 0 {
		t.Error("Key info key_length is zero")
	}
}

func TestThread_AddRemoveParticipant(t *testing.T) {
	thread := &Thread{
		ID:             "thread123",
		CreatorID:      "alice",
		ParticipantIDs: []string{"alice", "bob"},
		ThreadKey:      []byte("thread_key"),
		CreatedAt:      time.Now(),
		Metadata:       make(map[string]string),
	}

	// Test IsParticipant
	if !thread.IsParticipant("alice") {
		t.Error("Alice should be a participant")
	}
	if !thread.IsParticipant("bob") {
		t.Error("Bob should be a participant")
	}
	if thread.IsParticipant("charlie") {
		t.Error("Charlie should not be a participant")
	}

	// Test AddParticipant
	thread.AddParticipant("charlie")
	if !thread.IsParticipant("charlie") {
		t.Error("Charlie should be a participant after adding")
	}

	// Test adding existing participant (should not duplicate)
	originalLength := len(thread.ParticipantIDs)
	thread.AddParticipant("alice")
	if len(thread.ParticipantIDs) != originalLength {
		t.Error("Adding existing participant should not change length")
	}

	// Test RemoveParticipant
	thread.RemoveParticipant("bob")
	if thread.IsParticipant("bob") {
		t.Error("Bob should not be a participant after removing")
	}
	if !thread.IsParticipant("alice") {
		t.Error("Alice should still be a participant")
	}
	if !thread.IsParticipant("charlie") {
		t.Error("Charlie should still be a participant")
	}
}

func TestThread_GetThreadKeyBase64(t *testing.T) {
	threadKey := []byte("test_thread_key_32_bytes_long")
	thread := &Thread{
		ID:             "thread123",
		CreatorID:      "alice",
		ParticipantIDs: []string{"alice", "bob"},
		ThreadKey:      threadKey,
		CreatedAt:      time.Now(),
		Metadata:       make(map[string]string),
	}

	base64Key := thread.GetThreadKeyBase64()
	if base64Key == "" {
		t.Error("GetThreadKeyBase64() returned empty string")
	}
}
