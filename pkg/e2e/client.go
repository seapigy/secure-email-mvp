package e2e

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Client represents an E2E client for encryption operations
type Client struct {
	cryptoProvider *CryptoProvider
	config         E2EConfig
	userID         string
	keyPair        *KeyPair
}

// Message represents an E2E encrypted message
type Message struct {
	ID          string            `json:"id"`
	ThreadID    string            `json:"thread_id"`
	SenderID    string            `json:"sender_id"`
	RecipientID string            `json:"recipient_id"`
	Envelope    *Envelope         `json:"envelope"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewClient creates a new E2E client
func NewClient(config E2EConfig, userID string) (*Client, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("E2E system is disabled")
	}

	cryptoProvider, err := NewCryptoProvider(config.Crypto)
	if err != nil {
		return nil, fmt.Errorf("failed to create crypto provider: %w", err)
	}

	// Generate key pair for the user (using signature algorithm for signing)
	keyPair, err := cryptoProvider.GenerateKeyPair(config.Crypto.SignatureAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Client{
		cryptoProvider: cryptoProvider,
		config:         config,
		userID:         userID,
		keyPair:        keyPair,
	}, nil
}

// GetPublicKey returns the client's public key
func (c *Client) GetPublicKey() []byte {
	return c.keyPair.PublicKey
}

// GetPublicKeyBase64 returns the client's public key as base64
func (c *Client) GetPublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(c.keyPair.PublicKey)
}

// EncryptMessage encrypts a message for a recipient
func (c *Client) EncryptMessage(plaintext []byte, recipientPublicKey []byte, threadID string, recipientID string) (*Message, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Encrypt the message
	envelope, err := c.cryptoProvider.EncryptMessage(plaintext, recipientPublicKey, c.keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Create the message
	message := &Message{
		ID:          c.generateMessageID(),
		ThreadID:    threadID,
		SenderID:    c.userID,
		RecipientID: recipientID,
		Envelope:    envelope,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}

	return message, nil
}

// DecryptMessage decrypts a message
func (c *Client) DecryptMessage(message *Message, senderPublicKey []byte) ([]byte, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Verify this message is for us
	if message.RecipientID != c.userID {
		return nil, fmt.Errorf("message is not for this user")
	}

	// Decrypt the message
	plaintext, err := c.cryptoProvider.DecryptMessage(message.Envelope, c.keyPair.PrivateKey, senderPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	return plaintext, nil
}

// CreateThread creates a new encrypted thread
func (c *Client) CreateThread(participantIDs []string) (*Thread, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Generate thread key
	threadKey := make([]byte, 32)
	if _, err := rand.Read(threadKey); err != nil {
		return nil, fmt.Errorf("failed to generate thread key: %w", err)
	}

	// Create thread
	thread := &Thread{
		ID:             c.generateThreadID(),
		CreatorID:      c.userID,
		ParticipantIDs: participantIDs,
		ThreadKey:      threadKey,
		CreatedAt:      time.Now(),
		Metadata:       make(map[string]string),
	}

	return thread, nil
}

// EncryptThreadMessage encrypts a message for a thread
func (c *Client) EncryptThreadMessage(plaintext []byte, thread *Thread) (*Message, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Verify user is a participant
	if !c.isParticipant(thread, c.userID) {
		return nil, fmt.Errorf("user is not a participant in this thread")
	}

	// Encrypt with thread key using dedicated thread message function
	envelope, err := c.cryptoProvider.EncryptThreadMessage(plaintext, thread.ThreadKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt thread message: %w", err)
	}

	// Create the message
	message := &Message{
		ID:          c.generateMessageID(),
		ThreadID:    thread.ID,
		SenderID:    c.userID,
		RecipientID: "", // Thread messages don't have a specific recipient
		Envelope:    envelope,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}

	return message, nil
}

// DecryptThreadMessage decrypts a thread message
func (c *Client) DecryptThreadMessage(message *Message, thread *Thread) ([]byte, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Verify user is a participant
	if !c.isParticipant(thread, c.userID) {
		return nil, fmt.Errorf("user is not a participant in this thread")
	}

	// Decrypt with thread key using the dedicated thread message function
	plaintext, err := c.cryptoProvider.DecryptThreadMessage(message.Envelope, thread.ThreadKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt thread message: %w", err)
	}

	return plaintext, nil
}

// RotateKeys rotates the client's key pair
func (c *Client) RotateKeys() error {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Generate new key pair
	newKeyPair, err := c.cryptoProvider.GenerateKeyPair(c.config.Crypto.KEMAlgorithm)
	if err != nil {
		return fmt.Errorf("failed to generate new key pair: %w", err)
	}

	// Update key pair
	c.keyPair = newKeyPair

	return nil
}

// ExportKeyPair exports the key pair for backup
func (c *Client) ExportKeyPair() ([]byte, error) {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return nil, fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Export key pair (in production, this should be encrypted)
	exportData := map[string]interface{}{
		"user_id":     c.userID,
		"key_pair":    c.keyPair,
		"algorithm":   c.config.Crypto.KEMAlgorithm,
		"exported_at": time.Now(),
	}

	return json.Marshal(exportData)
}

// ImportKeyPair imports a key pair from backup
func (c *Client) ImportKeyPair(exportData []byte) error {
	// Check if E2E is enabled for this user
	if !c.config.IsEnabledForUser(c.userID) {
		return fmt.Errorf("E2E is not enabled for user: %s", c.userID)
	}

	// Parse export data
	var export map[string]interface{}
	if err := json.Unmarshal(exportData, &export); err != nil {
		return fmt.Errorf("failed to parse export data: %w", err)
	}

	// Verify user ID matches
	if export["user_id"] != c.userID {
		return fmt.Errorf("export data is for different user")
	}

	// Import key pair
	keyPairData, err := json.Marshal(export["key_pair"])
	if err != nil {
		return fmt.Errorf("failed to marshal key pair data: %w", err)
	}

	var keyPair KeyPair
	if err := json.Unmarshal(keyPairData, &keyPair); err != nil {
		return fmt.Errorf("failed to unmarshal key pair: %w", err)
	}

	c.keyPair = &keyPair

	return nil
}

// GetKeyInfo returns information about the current key pair
func (c *Client) GetKeyInfo() map[string]interface{} {
	return map[string]interface{}{
		"algorithm":  c.keyPair.Algorithm,
		"created_at": c.keyPair.CreatedAt,
		"expires_at": c.keyPair.ExpiresAt,
		"public_key": base64.StdEncoding.EncodeToString(c.keyPair.PublicKey),
		"key_length": len(c.keyPair.PublicKey),
	}
}

// isParticipant checks if a user is a participant in a thread
func (c *Client) isParticipant(thread *Thread, userID string) bool {
	for _, participantID := range thread.ParticipantIDs {
		if participantID == userID {
			return true
		}
	}
	return false
}

// generateMessageID generates a unique message ID
func (c *Client) generateMessageID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("msg_%x", id)
}

// generateThreadID generates a unique thread ID
func (c *Client) generateThreadID() string {
	id := make([]byte, 16)
	rand.Read(id)
	return fmt.Sprintf("thread_%x", id)
}

// Thread represents an encrypted thread
type Thread struct {
	ID             string            `json:"id"`
	CreatorID      string            `json:"creator_id"`
	ParticipantIDs []string          `json:"participant_ids"`
	ThreadKey      []byte            `json:"thread_key"`
	CreatedAt      time.Time         `json:"created_at"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// GetThreadKeyBase64 returns the thread key as base64
func (t *Thread) GetThreadKeyBase64() string {
	return base64.StdEncoding.EncodeToString(t.ThreadKey)
}

// AddParticipant adds a participant to the thread
func (t *Thread) AddParticipant(participantID string) {
	// Check if participant already exists
	for _, id := range t.ParticipantIDs {
		if id == participantID {
			return
		}
	}
	t.ParticipantIDs = append(t.ParticipantIDs, participantID)
}

// RemoveParticipant removes a participant from the thread
func (t *Thread) RemoveParticipant(participantID string) {
	for i, id := range t.ParticipantIDs {
		if id == participantID {
			t.ParticipantIDs = append(t.ParticipantIDs[:i], t.ParticipantIDs[i+1:]...)
			break
		}
	}
}

// IsParticipant checks if a user is a participant in the thread
func (t *Thread) IsParticipant(userID string) bool {
	for _, id := range t.ParticipantIDs {
		if id == userID {
			return true
		}
	}
	return false
}
