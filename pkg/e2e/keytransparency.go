package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// KeyTransparency handles public key verification and audit operations
type KeyTransparency struct {
	config KTConfig
	// db will be injected later when we integrate with the database
}

// PublicKeyEntry represents a public key entry in the transparency log
type PublicKeyEntry struct {
	ID          string     `json:"id"`
	UserUUID    string     `json:"user_uuid"`
	PublicKey   string     `json:"public_key"`
	KeyType     string     `json:"key_type"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	MerkleIndex int        `json:"merkle_index"`
	MerkleProof string     `json:"merkle_proof"`
	EntryHash   string     `json:"entry_hash"`
}

// LogEntry represents an entry in the transparency log
type LogEntry struct {
	ID          string    `json:"id"`
	EntryHash   string    `json:"entry_hash"`
	MerkleIndex int       `json:"merkle_index"`
	MerkleRoot  string    `json:"merkle_root"`
	Timestamp   time.Time `json:"timestamp"`
	Signature   string    `json:"signature"`
}

// MerkleProof represents a proof of inclusion in the Merkle tree
type MerkleProof struct {
	Index    int      `json:"index"`
	Path     []string `json:"path"`
	TreeSize int      `json:"tree_size"`
	TreeHead string   `json:"tree_head"`
}

// AuditResult represents the result of an audit operation
type AuditResult struct {
	Valid       bool      `json:"valid"`
	EntryHash   string    `json:"entry_hash"`
	MerkleProof string    `json:"merkle_proof"`
	TreeHead    string    `json:"tree_head"`
	Timestamp   time.Time `json:"timestamp"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// NewKeyTransparency creates a new Key Transparency instance
func NewKeyTransparency(config KTConfig) *KeyTransparency {
	return &KeyTransparency{
		config: config,
		// db will be injected later when we integrate with the database
	}
}

// RegisterPublicKey registers a public key in the transparency log
func (kt *KeyTransparency) RegisterPublicKey(userUUID, publicKey, keyType string) (*PublicKeyEntry, error) {
	if !kt.config.Enabled {
		return nil, fmt.Errorf("key transparency is disabled")
	}

	// Create the public key entry
	entry := &PublicKeyEntry{
		ID:        kt.generateEntryID(),
		UserUUID:  userUUID,
		PublicKey: publicKey,
		KeyType:   keyType,
		CreatedAt: time.Now(),
		ExpiresAt: kt.calculateKeyExpiry(),
	}

	// Calculate entry hash
	entryHash, err := kt.calculateEntryHash(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate entry hash: %w", err)
	}
	entry.EntryHash = entryHash

	// Get next Merkle index
	merkleIndex, err := kt.getNextMerkleIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get next Merkle index: %w", err)
	}
	entry.MerkleIndex = merkleIndex

	// Generate Merkle proof
	merkleProof, err := kt.generateMerkleProof(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Merkle proof: %w", err)
	}
	entry.MerkleProof = merkleProof

	// TODO: Store in database
	// kt.db.StorePublicKeyEntry(entry)

	// TODO: Create and store log entry in database
	// logEntry := &LogEntry{
	//     ID:          kt.generateLogEntryID(),
	//     EntryHash:   entryHash,
	//     MerkleIndex: merkleIndex,
	//     MerkleRoot:  kt.calculateMerkleRoot(merkleIndex),
	//     Timestamp:   time.Now(),
	//     Signature:   kt.signLogEntry(entryHash),
	// }
	// kt.db.StoreLogEntry(logEntry)

	return entry, nil
}

// VerifyPublicKey verifies a public key against the transparency log
func (kt *KeyTransparency) VerifyPublicKey(userUUID, publicKey, keyType string) (*AuditResult, error) {
	if !kt.config.Enabled {
		return &AuditResult{
			Valid:     false,
			ErrorMsg:  "key transparency is disabled",
			Timestamp: time.Now(),
		}, nil
	}

	// TODO: Query database for the public key entry
	// entry, err := kt.db.GetPublicKeyEntry(userUUID, keyType)
	// For now, simulate a positive verification
	entry := &PublicKeyEntry{
		UserUUID:    userUUID,
		PublicKey:   publicKey,
		KeyType:     keyType,
		MerkleIndex: 1,
		EntryHash:   kt.hashString(fmt.Sprintf("%s:%s:%s", userUUID, publicKey, keyType)),
	}

	// Verify the Merkle proof
	valid, err := kt.verifyMerkleProof(entry)
	if err != nil {
		return &AuditResult{
			Valid:     false,
			ErrorMsg:  fmt.Sprintf("proof verification failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &AuditResult{
		Valid:       valid,
		EntryHash:   entry.EntryHash,
		MerkleProof: entry.MerkleProof,
		TreeHead:    kt.calculateMerkleRoot(entry.MerkleIndex),
		Timestamp:   time.Now(),
	}, nil
}

// AuditLog performs an audit of the transparency log
func (kt *KeyTransparency) AuditLog(fromIndex, toIndex int) ([]*AuditResult, error) {
	if !kt.config.Enabled {
		return nil, fmt.Errorf("key transparency is disabled")
	}

	if !kt.config.VerifyProofs {
		return nil, fmt.Errorf("proof verification is disabled")
	}

	var results []*AuditResult

	// TODO: Query database for log entries in the range
	// For now, simulate audit results
	for i := fromIndex; i <= toIndex; i++ {
		result := &AuditResult{
			Valid:       true,
			EntryHash:   kt.hashString(fmt.Sprintf("entry_%d", i)),
			MerkleProof: kt.hashString(fmt.Sprintf("proof_%d", i)),
			TreeHead:    kt.calculateMerkleRoot(i),
			Timestamp:   time.Now(),
		}
		results = append(results, result)
	}

	return results, nil
}

// TODO: Temporary adapter to resolve compiler cache issue
// Remove this wrapper once the cache issue is resolved
func (kt *KeyTransparency) AuditLogWrapper(fromIndex, toIndex int) error {
	_, err := kt.AuditLog(fromIndex, toIndex)
	return err
}

// GetPublicKeys retrieves all public keys for a user
func (kt *KeyTransparency) GetPublicKeys(userUUID string) ([]*PublicKeyEntry, error) {
	if !kt.config.Enabled {
		return nil, fmt.Errorf("key transparency is disabled")
	}

	// TODO: Query database for user's public keys
	// For now, return empty slice
	return []*PublicKeyEntry{}, nil
}

// RevokePublicKey revokes a public key in the transparency log
func (kt *KeyTransparency) RevokePublicKey(userUUID, keyType string) error {
	if !kt.config.Enabled {
		return fmt.Errorf("key transparency is disabled")
	}

	// TODO: Create revocation entry and store in database
	// revocationEntry := &PublicKeyEntry{
	//     ID:        kt.generateEntryID(),
	//     UserUUID:  userUUID,
	//     PublicKey: "", // Empty for revocation
	//     KeyType:   keyType + "_REVOKED",
	//     CreatedAt: time.Now(),
	// }
	// entryHash, err := kt.calculateEntryHash(revocationEntry)
	// if err != nil {
	//     return fmt.Errorf("failed to calculate revocation entry hash: %w", err)
	// }
	// kt.db.StoreRevocation(revocationEntry)

	return nil
}

// Helper functions

// calculateEntryHash calculates the hash of a public key entry
func (kt *KeyTransparency) calculateEntryHash(entry *PublicKeyEntry) (string, error) {
	// Create canonical representation
	canonical := map[string]interface{}{
		"user_uuid":  entry.UserUUID,
		"public_key": entry.PublicKey,
		"key_type":   entry.KeyType,
		"created_at": entry.CreatedAt.Unix(),
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal entry: %w", err)
	}

	return kt.hashString(string(data)), nil
}

// generateMerkleProof generates a Merkle proof for an entry
func (kt *KeyTransparency) generateMerkleProof(entry *PublicKeyEntry) (string, error) {
	// TODO: Implement actual Merkle tree proof generation
	// For now, create a simple proof based on the entry hash
	proof := MerkleProof{
		Index:    entry.MerkleIndex,
		Path:     []string{entry.EntryHash, kt.hashString("sibling_hash")},
		TreeSize: entry.MerkleIndex + 1,
		TreeHead: kt.calculateMerkleRoot(entry.MerkleIndex),
	}

	proofData, err := json.Marshal(proof)
	if err != nil {
		return "", fmt.Errorf("failed to marshal proof: %w", err)
	}

	return base64.StdEncoding.EncodeToString(proofData), nil
}

// verifyMerkleProof verifies a Merkle proof
func (kt *KeyTransparency) verifyMerkleProof(entry *PublicKeyEntry) (bool, error) {
	if !kt.config.VerifyProofs {
		return true, nil // Skip verification if disabled
	}

	// TODO: Implement actual Merkle proof verification
	// For now, always return true for placeholder implementation
	return true, nil
}

// calculateMerkleRoot calculates the Merkle root at a given index
func (kt *KeyTransparency) calculateMerkleRoot(index int) string {
	// TODO: Implement actual Merkle root calculation
	// For now, create a deterministic root based on index
	return kt.hashString(fmt.Sprintf("merkle_root_%d", index))
}

// getNextMerkleIndex gets the next available Merkle index
func (kt *KeyTransparency) getNextMerkleIndex() (int, error) {
	// TODO: Query database for the next index
	// For now, return a simple incrementing counter
	return int(time.Now().Unix()) % 1000000, nil
}

// signLogEntry signs a log entry
func (kt *KeyTransparency) signLogEntry(entryHash string) string {
	// TODO: Implement actual signing
	// For now, create a simple HMAC-based signature
	return kt.hashString(fmt.Sprintf("signature_%s", entryHash))
}

// calculateKeyExpiry calculates when a key should expire
func (kt *KeyTransparency) calculateKeyExpiry() *time.Time {
	// Keys expire after 1 year by default
	expiry := time.Now().AddDate(1, 0, 0)
	return &expiry
}

// generateEntryID generates a unique entry ID
func (kt *KeyTransparency) generateEntryID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("kt_entry_%d", timestamp)
}

// generateLogEntryID generates a unique log entry ID
func (kt *KeyTransparency) generateLogEntryID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("kt_log_%d", timestamp)
}

// hashString creates a SHA-256 hash of a string
func (kt *KeyTransparency) hashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// Validation helpers

// ValidatePublicKey validates a public key format
func (kt *KeyTransparency) ValidatePublicKey(publicKey, keyType string) error {
	if publicKey == "" {
		return fmt.Errorf("public key cannot be empty")
	}

	if keyType == "" {
		return fmt.Errorf("key type cannot be empty")
	}

	// Validate key type
	validKeyTypes := []string{"kyber512", "kyber768", "kyber1024", "dilithium2", "dilithium3", "dilithium5"}
	for _, validType := range validKeyTypes {
		if keyType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid key type: %s", keyType)
}

// ValidateUserUUID validates a user UUID format
func (kt *KeyTransparency) ValidateUserUUID(userUUID string) error {
	if userUUID == "" {
		return fmt.Errorf("user UUID cannot be empty")
	}

	if len(userUUID) < 8 {
		return fmt.Errorf("user UUID too short")
	}

	return nil
}
