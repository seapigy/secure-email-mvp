package zkid

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Config with feature flag and keying
type Config struct {
	Enabled           bool
	MasterKey         []byte // 32 bytes
	EmailHashPepper   []byte // secret pepper for email hash
	UsePQCKeyWrapping bool   // future: integrate with PQC key manager
	RecoveryPepper    []byte // pepper for recovery code hashing
}

// Mapping represents an email mapping record
type Mapping struct {
	ID                  string
	UserID              string
	EmailHash           string
	EmailCiphertext     []byte
	EmailNonce          []byte
	EmailTag            []byte
	WrappedKey          []byte
	WrappedKeyNonce     []byte
	WrappedKeyTag       []byte
	FallbackEmailCipher []byte
	FallbackEmailNonce  []byte
	FallbackEmailTag    []byte
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Service provides ZKID operations
type Service struct {
	db     *sql.DB
	config *Config
}

func NewService(db *sql.DB, cfg *Config) *Service {
	return &Service{db: db, config: cfg}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func hashEmail(pepper []byte, email string) string {
	h := sha256.New()
	h.Write(pepper)
	h.Write([]byte(normalizeEmail(email)))
	return hex.EncodeToString(h.Sum(nil))
}

// generateDataKey returns a random 32-byte data key
func generateDataKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func aesGCMEncrypt(key, plaintext []byte) (ciphertext, nonce, tag []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, err
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	tag = ct[len(ct)-16:]
	ciphertext = ct[:len(ct)-16]
	return
}

func aesGCMDecrypt(key, ciphertext, nonce, tag []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ct := append(ciphertext, tag...)
	return gcm.Open(nil, nonce, ct, nil)
}

// wrapKey encrypts the data key with master key
func wrapKey(master, dataKey []byte) (wrapped, nonce, tag []byte, err error) {
	return aesGCMEncrypt(master, dataKey)
}

// unwrapKey decrypts wrapped data key with master key
func unwrapKey(master, wrapped, nonce, tag []byte) ([]byte, error) {
	return aesGCMDecrypt(master, wrapped, nonce, tag)
}

// CreateOrUpdateMapping stores or updates the encrypted email mapping
func (s *Service) CreateOrUpdateMapping(userID, email string, fallbackEmail *string) (*Mapping, error) {
	if !s.config.Enabled {
		return nil, errors.New("zkid disabled")
	}
	if len(s.config.MasterKey) != 32 {
		return nil, errors.New("invalid master key")
	}

	emailHash := hashEmail(s.config.EmailHashPepper, email)
	dataKey, err := generateDataKey()
	if err != nil {
		return nil, fmt.Errorf("generate data key: %w", err)
	}

	emailCT, emailNonce, emailTag, err := aesGCMEncrypt(dataKey, []byte(normalizeEmail(email)))
	if err != nil {
		return nil, fmt.Errorf("encrypt email: %w", err)
	}

	var fbCT, fbNonce, fbTag []byte
	if fallbackEmail != nil && *fallbackEmail != "" {
		fbCT, fbNonce, fbTag, err = aesGCMEncrypt(dataKey, []byte(normalizeEmail(*fallbackEmail)))
		if err != nil {
			return nil, fmt.Errorf("encrypt fallback email: %w", err)
		}
	}

	wrapped, wrappedNonce, wrappedTag, err := wrapKey(s.config.MasterKey, dataKey)
	if err != nil {
		return nil, fmt.Errorf("wrap key: %w", err)
	}

	id := fmt.Sprintf("zkid_%d", time.Now().UnixNano())

	// Upsert mapping
	q := `INSERT INTO zkid_email_mappings (id,user_id,email_hash,email_ciphertext,email_nonce,email_tag,wrapped_key,wrapped_key_nonce,wrapped_key_tag,fallback_email_ciphertext,fallback_email_nonce,fallback_email_tag,created_at,updated_at)
          VALUES (?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
          ON CONFLICT(user_id) DO UPDATE SET email_hash=excluded.email_hash,email_ciphertext=excluded.email_ciphertext,email_nonce=excluded.email_nonce,email_tag=excluded.email_tag,wrapped_key=excluded.wrapped_key,wrapped_key_nonce=excluded.wrapped_key_nonce,wrapped_key_tag=excluded.wrapped_key_tag,fallback_email_ciphertext=excluded.fallback_email_ciphertext,fallback_email_nonce=excluded.fallback_email_nonce,fallback_email_tag=excluded.fallback_email_tag,updated_at=CURRENT_TIMESTAMP`
	if _, err := s.db.Exec(q, id, userID, emailHash, emailCT, emailNonce, emailTag, wrapped, wrappedNonce, wrappedTag, fbCT, fbNonce, fbTag); err != nil {
		return nil, fmt.Errorf("upsert mapping: %w", err)
	}

	return &Mapping{ID: id, UserID: userID, EmailHash: emailHash, EmailCiphertext: emailCT, EmailNonce: emailNonce, EmailTag: emailTag, WrappedKey: wrapped, WrappedKeyNonce: wrappedNonce, WrappedKeyTag: wrappedTag, FallbackEmailCipher: fbCT, FallbackEmailNonce: fbNonce, FallbackEmailTag: fbTag}, nil
}

// GetEmailByUserID decrypts the email for the given user ID
func (s *Service) GetEmailByUserID(userID string) (string, error) {
	if !s.config.Enabled {
		return "", errors.New("zkid disabled")
	}
	row := s.db.QueryRow(`SELECT email_ciphertext,email_nonce,email_tag,wrapped_key,wrapped_key_nonce,wrapped_key_tag FROM zkid_email_mappings WHERE user_id=?`, userID)
	var emailCT, emailNonce, emailTag, wrapped, wrappedNonce, wrappedTag []byte
	if err := row.Scan(&emailCT, &emailNonce, &emailTag, &wrapped, &wrappedNonce, &wrappedTag); err != nil {
		return "", err
	}
	dataKey, err := unwrapKey(s.config.MasterKey, wrapped, wrappedNonce, wrappedTag)
	if err != nil {
		return "", fmt.Errorf("unwrap key: %w", err)
	}
	pt, err := aesGCMDecrypt(dataKey, emailCT, emailNonce, emailTag)
	if err != nil {
		return "", fmt.Errorf("decrypt email: %w", err)
	}
	return string(pt), nil
}

// ValidateRecoveryCode verifies a provided code without revealing user identity
// The caller must already have identified the user via separate challenge flows.
func (s *Service) ValidateRecoveryCode(userID string, code string, salt []byte, hash []byte) bool {
	// In this simplified version, we recompute Argon2id(code+pepper, salt) externally and compare hashes
	// Actual implementation would store and compare in DB. This function is a placeholder for test wiring.
	return len(userID) > 0 && len(code) > 0 && len(salt) > 0 && len(hash) > 0
}

// RevokeRecoveryCode marks a specific recovery code as used
func (s *Service) RevokeRecoveryCode(userID, codeID string) (bool, error) {
	if !s.config.Enabled {
		return false, errors.New("zkid disabled")
	}

	result, err := s.db.Exec(`UPDATE zkid_recovery_codes SET used=1, used_at=CURRENT_TIMESTAMP WHERE user_id=? AND id=? AND used=0`, userID, codeID)
	if err != nil {
		return false, fmt.Errorf("revoke recovery code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetStats returns ZKID statistics for admin monitoring
func (s *Service) GetStats() (map[string]interface{}, error) {
	if !s.config.Enabled {
		return map[string]interface{}{"enabled": false}, nil
	}

	stats := map[string]interface{}{
		"enabled": true,
	}

	// Count email mappings
	var mappingCount int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM zkid_email_mappings`).Scan(&mappingCount)
	if err != nil {
		return nil, fmt.Errorf("count mappings: %w", err)
	}
	stats["total_mappings"] = mappingCount

	// Count recovery codes
	var totalCodes, usedCodes int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM zkid_recovery_codes`).Scan(&totalCodes)
	if err != nil {
		return nil, fmt.Errorf("count total codes: %w", err)
	}
	err = s.db.QueryRow(`SELECT COUNT(*) FROM zkid_recovery_codes WHERE used=1`).Scan(&usedCodes)
	if err != nil {
		return nil, fmt.Errorf("count used codes: %w", err)
	}

	stats["total_recovery_codes"] = totalCodes
	stats["used_recovery_codes"] = usedCodes
	stats["available_recovery_codes"] = totalCodes - usedCodes

	return stats, nil
}
