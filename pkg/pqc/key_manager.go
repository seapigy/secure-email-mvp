package pqc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// KeyManager handles Kyber key generation, encapsulation, and HSM operations
type KeyManager struct {
	config     *PQCConfig
	keys       map[string]*KyberKeyPair
	currentKey string
	mu         sync.RWMutex
}

// KyberKeyPair represents a Kyber public/private key pair
type KyberKeyPair struct {
	ID               string            `json:"id"`
	PublicKey        []byte            `json:"public_key"`
	PrivateKey       []byte            `json:"private_key"`
	CreatedAt        time.Time         `json:"created_at"`
	ExpiresAt        time.Time         `json:"expires_at"`
	KyberLevel       int               `json:"kyber_level"`
	HSMKeyID         string            `json:"hsm_key_id,omitempty"`
	encapsulatedKeys map[string][]byte `json:"-"` // Maps ciphertext to symmetric key for simulation
}

// NewKeyManager creates a new key manager instance
func NewKeyManager(config *PQCConfig) (*KeyManager, error) {
	km := &KeyManager{
		config: config,
		keys:   make(map[string]*KyberKeyPair),
	}

	// Generate initial key pair
	if err := km.generateNewKeyPair(); err != nil {
		return nil, fmt.Errorf("failed to generate initial key pair: %w", err)
	}

	// Start key rotation timer if enabled
	if config.KeyRotationDays > 0 {
		go km.startKeyRotation()
	}

	return km, nil
}

// generateNewKeyPair creates a new Kyber key pair
func (km *KeyManager) generateNewKeyPair() error {
	log.Println("generateNewKeyPair: starting key generation")

	// Generate Kyber key pair (simulated for now - would use actual Kyber library)
	publicKey := make([]byte, km.getKyberPublicKeySize())
	privateKey := make([]byte, km.getKyberPrivateKeySize())

	if _, err := io.ReadFull(rand.Reader, publicKey); err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}

	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	keyID := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(km.config.KeyRotationDays) * 24 * time.Hour)

	keyPair := &KyberKeyPair{
		ID:               keyID,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		CreatedAt:        time.Now(),
		ExpiresAt:        expiresAt,
		KyberLevel:       km.config.KyberLevel,
		encapsulatedKeys: make(map[string][]byte),
	}

	log.Println("generateNewKeyPair: key pair created, ID:", keyID)

	// Store in HSM if enabled (but not while holding lock)
	if km.config.HSMEnabled {
		log.Println("generateNewKeyPair: HSM enabled, will store key")
		// Note: HSM storage will be handled outside the lock
	}

	km.keys[keyID] = keyPair
	km.currentKey = keyID

	log.Println("generateNewKeyPair: key pair stored, current key set to:", keyID)
	return nil
}

// getKyberPublicKeySize returns the size of Kyber public key based on level
func (km *KeyManager) getKyberPublicKeySize() int {
	switch km.config.KyberLevel {
	case 512:
		return 800
	case 768:
		return 1184
	case 1024:
		return 1568
	default:
		return 1184 // Default to Kyber-768
	}
}

// getKyberPrivateKeySize returns the size of Kyber private key based on level
func (km *KeyManager) getKyberPrivateKeySize() int {
	switch km.config.KyberLevel {
	case 512:
		return 1632
	case 768:
		return 2400
	case 1024:
		return 3168
	default:
		return 2400 // Default to Kyber-768
	}
}

// getKyberCiphertextSize returns the size of Kyber ciphertext based on level
func (km *KeyManager) getKyberCiphertextSize() int {
	switch km.config.KyberLevel {
	case 512:
		return 768
	case 768:
		return 1088
	case 1024:
		return 1568
	default:
		return 1088 // Default to Kyber-768
	}
}

// EncapsulateKey encapsulates a symmetric key using the current Kyber public key
func (km *KeyManager) EncapsulateKey(symmetricKey []byte) ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.currentKey == "" {
		return nil, fmt.Errorf("no current key available")
	}

	keyPair := km.keys[km.currentKey]
	if keyPair == nil {
		return nil, fmt.Errorf("current key not found")
	}

	// Check if key is expired
	if time.Now().After(keyPair.ExpiresAt) {
		return nil, fmt.Errorf("current key is expired")
	}

	// Simulate Kyber encapsulation (would use actual Kyber library)
	// In a real implementation, this would use the Kyber Encaps function
	ciphertext := make([]byte, km.getKyberCiphertextSize())

	// For simulation, we'll create a deterministic ciphertext based on the key
	// In reality, this would be the output of Kyber.Encaps(publicKey)
	hash := sha256.New()
	hash.Write(keyPair.PublicKey)
	hash.Write(symmetricKey)
	hash.Write([]byte("encapsulation"))

	// Use hash output to generate ciphertext
	hashOutput := hash.Sum(nil)
	copy(ciphertext, hashOutput)
	if len(ciphertext) > len(hashOutput) {
		// Fill remaining bytes with random data
		remaining := ciphertext[len(hashOutput):]
		if _, err := io.ReadFull(rand.Reader, remaining); err != nil {
			return nil, fmt.Errorf("failed to generate random ciphertext: %w", err)
		}
	}

	// Store the symmetric key for decapsulation (in a real implementation, this would be derived)
	keyPair.encapsulatedKeys[string(ciphertext)] = symmetricKey

	return ciphertext, nil
}

// DecapsulateKey decapsulates a symmetric key using the current Kyber private key
func (km *KeyManager) DecapsulateKey(ciphertext []byte) ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.currentKey == "" {
		return nil, fmt.Errorf("no current key available")
	}

	keyPair := km.keys[km.currentKey]
	if keyPair == nil {
		return nil, fmt.Errorf("current key not found")
	}

	// Check if key is expired
	if time.Now().After(keyPair.ExpiresAt) {
		return nil, fmt.Errorf("current key is expired")
	}

	// If HSM is enabled, use HSM for decapsulation
	if km.config.HSMEnabled && keyPair.HSMKeyID != "" {
		return km.decapsulateWithHSM(keyPair.HSMKeyID, ciphertext)
	}

	// For simulation, retrieve the stored symmetric key
	// In a real implementation, this would use the Kyber Decaps function
	symmetricKey, exists := keyPair.encapsulatedKeys[string(ciphertext)]
	if !exists {
		return nil, fmt.Errorf("ciphertext not found in encapsulated keys")
	}

	return symmetricKey, nil
}

// GetCurrentKeyID returns the ID of the current key
func (km *KeyManager) GetCurrentKeyID() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.currentKey
}

// GetCurrentPublicKey returns the current public key
func (km *KeyManager) GetCurrentPublicKey() ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.currentKey == "" {
		return nil, fmt.Errorf("no current key available")
	}

	keyPair := km.keys[km.currentKey]
	if keyPair == nil {
		return nil, fmt.Errorf("current key not found")
	}

	return keyPair.PublicKey, nil
}

// RotateKeys generates a new key pair and marks it as current
func (km *KeyManager) RotateKeys() error {
	log.Println("RotateKeys: starting key rotation")

	// Step 1: Create key under lock (fast operations only)
	km.mu.Lock()
	log.Println("RotateKeys: acquired write lock for key creation")

	if err := km.generateNewKeyPair(); err != nil {
		km.mu.Unlock()
		return fmt.Errorf("failed to generate new key pair: %w", err)
	}

	// Get the newly created key for HSM operations
	newKeyID := km.currentKey
	newKeyPair := km.keys[newKeyID]

	km.mu.Unlock()
	log.Println("RotateKeys: released write lock after key creation")

	// Step 2: Run slow HSM operations outside lock
	if km.config.HSMEnabled && newKeyPair != nil {
		log.Println("RotateKeys: starting HSM storage")
		hsmKeyID, err := km.storeInHSM(newKeyPair)
		if err != nil {
			return fmt.Errorf("failed to store key in HSM: %w", err)
		}

		// Update HSM key ID under lock
		km.mu.Lock()
		if keyPair, exists := km.keys[newKeyID]; exists {
			keyPair.HSMKeyID = hsmKeyID
		}
		km.mu.Unlock()
		log.Println("RotateKeys: HSM storage completed")
	}

	// Step 3: Clean up old keys
	log.Println("RotateKeys: starting cleanup")
	km.cleanupOldKeys()

	log.Println("RotateKeys: key rotation completed successfully")
	return nil
}

// cleanupOldKeys removes expired keys and keeps only recent ones
func (km *KeyManager) cleanupOldKeys() {
	log.Println("cleanupOldKeys: starting cleanup")

	// Step 1: Gather expired keys under read lock
	km.mu.RLock()
	var expired []string
	var keysToKeep []string

	for id, keyPair := range km.keys {
		// Never remove the current key, even if it appears expired
		if id == km.currentKey {
			keysToKeep = append(keysToKeep, id)
			continue
		}

		if time.Now().After(keyPair.ExpiresAt) {
			expired = append(expired, id)
		} else {
			keysToKeep = append(keysToKeep, id)
		}
	}
	km.mu.RUnlock()

	log.Printf("cleanupOldKeys: found %d expired keys, %d keys to keep", len(expired), len(keysToKeep))

	// Step 2: Remove expired keys under write lock
	if len(expired) > 0 {
		km.mu.Lock()
		for _, id := range expired {
			delete(km.keys, id)
			log.Printf("cleanupOldKeys: removed expired key %s", id)
		}
		km.mu.Unlock()
	}

	log.Println("cleanupOldKeys: cleanup completed")
}

// startKeyRotation starts the automatic key rotation timer
func (km *KeyManager) startKeyRotation() {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for range ticker.C {
		km.mu.RLock()
		needsRotation := false

		if km.currentKey != "" {
			keyPair := km.keys[km.currentKey]
			if keyPair != nil && time.Now().After(keyPair.ExpiresAt) {
				needsRotation = true
			}
		}
		km.mu.RUnlock()

		if needsRotation {
			if err := km.RotateKeys(); err != nil {
				// Log error but continue
				fmt.Printf("Failed to rotate keys: %v\n", err)
			}
		}
	}
}

// storeInHSM stores a key pair in the HSM (simulated)
func (km *KeyManager) storeInHSM(keyPair *KyberKeyPair) (string, error) {
	log.Println("storeInHSM: starting HSM storage for key", keyPair.ID)

	// Simulate HSM storage
	// In a real implementation, this would use actual HSM APIs
	hsmKeyID := fmt.Sprintf("hsm_%s_%d", keyPair.ID, time.Now().Unix())

	// Simulate HSM operation delay
	time.Sleep(10 * time.Millisecond)

	log.Println("storeInHSM: completed HSM storage, ID:", hsmKeyID)
	return hsmKeyID, nil
}

// decapsulateWithHSM performs decapsulation using HSM (simulated)
func (km *KeyManager) decapsulateWithHSM(hsmKeyID string, ciphertext []byte) ([]byte, error) {
	// Simulate HSM decapsulation
	// In a real implementation, this would use actual HSM APIs

	// Simulate HSM operation delay
	time.Sleep(5 * time.Millisecond)

	// For simulation, derive key from HSM key ID and ciphertext
	hash := sha256.New()
	hash.Write([]byte(hsmKeyID))
	hash.Write(ciphertext)
	hash.Write([]byte("hsm_decapsulation"))

	symmetricKey := hash.Sum(nil)[:32] // Take first 32 bytes for 256-bit key

	return symmetricKey, nil
}

// GetKeyStats returns statistics about the key manager
func (km *KeyManager) GetKeyStats() map[string]interface{} {
	km.mu.RLock()
	defer km.mu.RUnlock()

	activeKeys := 0
	totalKeys := len(km.keys)

	for _, keyPair := range km.keys {
		if time.Now().Before(keyPair.ExpiresAt) {
			activeKeys++
		}
	}

	return map[string]interface{}{
		"total_keys":     totalKeys,
		"active_keys":    activeKeys,
		"current_key_id": km.currentKey,
		"hsm_enabled":    km.config.HSMEnabled,
		"kyber_level":    km.config.KyberLevel,
	}
}

// ExportPublicKey exports the current public key in base64 format
func (km *KeyManager) ExportPublicKey() (string, error) {
	publicKey, err := km.GetCurrentPublicKey()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(publicKey), nil
}
