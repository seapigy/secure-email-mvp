package pqc

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/auth"
)

// PQCIntegration provides integration between PQC and existing encryption systems
type PQCIntegration struct {
	service *PQCService
	db      *sql.DB
}

// NewPQCIntegration creates a new PQC integration instance
func NewPQCIntegration(db *sql.DB, config *PQCConfig) (*PQCIntegration, error) {
	service, err := NewPQCService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PQC service: %w", err)
	}

	return &PQCIntegration{
		service: service,
		db:      db,
	}, nil
}

// EncryptEmailContent encrypts email content using PQC if enabled, otherwise falls back to AES
func (pi *PQCIntegration) EncryptEmailContent(plaintext []byte, emailID string) (*EncryptionResult, error) {
	if !pi.service.IsEnabled() {
		// Fall back to existing AES encryption
		return pi.encryptWithAES(plaintext, emailID)
	}

	// Use PQC hybrid encryption
	return pi.encryptWithPQC(plaintext, emailID)
}

// DecryptEmailContent decrypts email content using appropriate method based on encryption version
func (pi *PQCIntegration) DecryptEmailContent(emailID string) (*DecryptionResult, error) {
	// Get encryption metadata from database
	metadata, err := pi.getEmailEncryptionMetadata(emailID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption metadata: %w", err)
	}

	if metadata.PQCEnabled {
		return pi.decryptWithPQC(emailID, metadata)
	} else {
		return pi.decryptWithAES(emailID, metadata)
	}
}

// MigrateEmailToPQC migrates an existing AES-encrypted email to PQC encryption
func (pi *PQCIntegration) MigrateEmailToPQC(emailID string) error {
	if !pi.service.IsEnabled() {
		return fmt.Errorf("PQC is not enabled")
	}

	// Get current encryption metadata
	metadata, err := pi.getEmailEncryptionMetadata(emailID)
	if err != nil {
		return fmt.Errorf("failed to get encryption metadata: %w", err)
	}

	if metadata.PQCEnabled {
		return fmt.Errorf("email is already PQC encrypted")
	}

	// Decrypt with AES
	aesResult, err := pi.decryptWithAES(emailID, metadata)
	if err != nil {
		return fmt.Errorf("failed to decrypt with AES: %w", err)
	}

	// Re-encrypt with PQC
	pqcResult, err := pi.encryptWithPQC(aesResult.Plaintext, emailID)
	if err != nil {
		return fmt.Errorf("failed to re-encrypt with PQC: %w", err)
	}

	// Update database with PQC metadata
	err = pi.updateEmailEncryptionMetadata(emailID, &EmailEncryptionMetadata{
		PQCEnabled:        true,
		EncryptionVersion: "PQC-HYBRID",
		PQCKeyID:          pqcResult.KeyID,
		PQCEncryptionTime: time.Now(),
		PQCEncryptedData:  pqcResult.SerializedData,
	})
	if err != nil {
		return fmt.Errorf("failed to update encryption metadata: %w", err)
	}

	log.Printf("Successfully migrated email %s to PQC encryption", emailID)
	return nil
}

// BatchMigrateEmailsToPQC migrates multiple emails to PQC encryption
func (pi *PQCIntegration) BatchMigrateEmailsToPQC(batchSize int) (int, error) {
	if !pi.service.IsEnabled() {
		return 0, fmt.Errorf("PQC is not enabled")
	}

	// Get emails that need migration
	emails, err := pi.getEmailsForMigration(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get emails for migration: %w", err)
	}

	migratedCount := 0
	for _, emailID := range emails {
		if err := pi.MigrateEmailToPQC(emailID); err != nil {
			log.Printf("Failed to migrate email %s: %v", emailID, err)
			continue
		}
		migratedCount++
	}

	return migratedCount, nil
}

// EncryptionResult represents the result of an encryption operation
type EncryptionResult struct {
	EncryptedData    []byte // The encrypted data (for R2 storage)
	SerializedData   string // Serialized metadata (for database storage)
	KeyID            string // Key identifier
	EncryptionMethod string // "AES-256-GCM" or "PQC-HYBRID"
	Metadata         map[string]interface{}
}

// DecryptionResult represents the result of a decryption operation
type DecryptionResult struct {
	Plaintext        []byte
	EncryptionMethod string
	Metadata         map[string]interface{}
}

// EmailEncryptionMetadata represents encryption metadata stored in the database
type EmailEncryptionMetadata struct {
	PQCEnabled        bool
	EncryptionVersion string
	PQCKeyID          string
	PQCEncryptionTime time.Time
	PQCEncryptedData  string
	AESKey            string
	AESNonce          string
	AESAuthTag        string
}

// encryptWithPQC encrypts data using PQC hybrid encryption
func (pi *PQCIntegration) encryptWithPQC(plaintext []byte, emailID string) (*EncryptionResult, error) {
	startTime := time.Now()

	// Encrypt with PQC hybrid
	hybridData, err := pi.service.EncryptHybrid(plaintext, fmt.Sprintf("email_%s", emailID))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with PQC: %w", err)
	}

	// Serialize hybrid data
	serializedData, err := pi.service.SerializeHybridData(hybridData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize hybrid data: %w", err)
	}

	// Use AES-256-GCM data for R2 storage (primary method)
	encryptedData := append(hybridData.AES256GCMData.Ciphertext, hybridData.AES256GCMData.AuthTag...)

	// Log performance metric
	pi.logPerformanceMetric("ENCRYPT", time.Since(startTime), len(plaintext), hybridData.KyberLevel, true, nil, emailID)

	return &EncryptionResult{
		EncryptedData:    encryptedData,
		SerializedData:   serializedData,
		KeyID:            hybridData.KeyID,
		EncryptionMethod: "PQC-HYBRID",
		Metadata: map[string]interface{}{
			"kyber_level":     hybridData.KyberLevel,
			"hybrid_mode":     hybridData.HybridMode,
			"encryption_time": hybridData.EncryptionTime,
			"version":         hybridData.Version,
		},
	}, nil
}

// decryptWithPQC decrypts data using PQC hybrid decryption
func (pi *PQCIntegration) decryptWithPQC(emailID string, metadata *EmailEncryptionMetadata) (*DecryptionResult, error) {
	startTime := time.Now()

	// Deserialize hybrid data
	hybridData, err := pi.service.DeserializeHybridData(metadata.PQCEncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize hybrid data: %w", err)
	}

	// Decrypt with PQC hybrid
	plaintext, err := pi.service.DecryptHybrid(hybridData, fmt.Sprintf("email_%s", emailID))
	if err != nil {
		// Log error
		pi.logPerformanceMetric("DECRYPT", time.Since(startTime), 0, hybridData.KyberLevel, false, err, emailID)
		return nil, fmt.Errorf("failed to decrypt with PQC: %w", err)
	}

	// Log performance metric
	pi.logPerformanceMetric("DECRYPT", time.Since(startTime), len(plaintext), hybridData.KyberLevel, true, nil, emailID)

	return &DecryptionResult{
		Plaintext:        plaintext,
		EncryptionMethod: "PQC-HYBRID",
		Metadata: map[string]interface{}{
			"kyber_level":     hybridData.KyberLevel,
			"hybrid_mode":     hybridData.HybridMode,
			"encryption_time": hybridData.EncryptionTime,
			"version":         hybridData.Version,
		},
	}, nil
}

// encryptWithAES encrypts data using existing AES-256-GCM encryption
func (pi *PQCIntegration) encryptWithAES(plaintext []byte, emailID string) (*EncryptionResult, error) {
	startTime := time.Now()

	// Use existing AES encryption
	encryptedData, err := auth.EncryptAES256GCM(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with AES: %w", err)
	}

	// Create metadata for database storage
	metadata := map[string]interface{}{
		"key":      base64.StdEncoding.EncodeToString(encryptedData.Key),
		"nonce":    base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		"auth_tag": base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
	}

	serializedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Combine ciphertext and auth tag for R2 storage
	encryptedBlob := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

	// Log performance metric
	pi.logPerformanceMetric("ENCRYPT", time.Since(startTime), len(plaintext), 0, true, nil, emailID)

	return &EncryptionResult{
		EncryptedData:    encryptedBlob,
		SerializedData:   string(serializedMetadata),
		KeyID:            "",
		EncryptionMethod: "AES-256-GCM",
		Metadata:         metadata,
	}, nil
}

// decryptWithAES decrypts data using existing AES-256-GCM decryption
func (pi *PQCIntegration) decryptWithAES(emailID string, metadata *EmailEncryptionMetadata) (*DecryptionResult, error) {
	startTime := time.Now()

	// Parse metadata
	var aesMetadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadata.PQCEncryptedData), &aesMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse AES metadata: %w", err)
	}

	// Decode components
	key, err := base64.StdEncoding.DecodeString(aesMetadata["key"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES key: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(aesMetadata["nonce"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES nonce: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(aesMetadata["auth_tag"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES auth tag: %w", err)
	}

	// Create EncryptedData struct
	encryptedData := &auth.EncryptedData{
		Key:     key,
		Nonce:   nonce,
		AuthTag: authTag,
	}

	// Decrypt
	plaintext, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		// Log error
		pi.logPerformanceMetric("DECRYPT", time.Since(startTime), 0, 0, false, err, emailID)
		return nil, fmt.Errorf("failed to decrypt with AES: %w", err)
	}

	// Log performance metric
	pi.logPerformanceMetric("DECRYPT", time.Since(startTime), len(plaintext), 0, true, nil, emailID)

	return &DecryptionResult{
		Plaintext:        plaintext,
		EncryptionMethod: "AES-256-GCM",
		Metadata:         aesMetadata,
	}, nil
}

// getEmailEncryptionMetadata retrieves encryption metadata from the database
func (pi *PQCIntegration) getEmailEncryptionMetadata(emailID string) (*EmailEncryptionMetadata, error) {
	query := `
		SELECT 
			pqc_enabled,
			encryption_version,
			pqc_key_id,
			pqc_encryption_time,
			pqc_encrypted_data,
			encrypted_key,
			encryption_nonce,
			encryption_auth_tag
		FROM emails 
		WHERE id = ?
	`

	var metadata EmailEncryptionMetadata
	var pqcEncryptionTime sql.NullTime

	err := pi.db.QueryRow(query, emailID).Scan(
		&metadata.PQCEnabled,
		&metadata.EncryptionVersion,
		&metadata.PQCKeyID,
		&pqcEncryptionTime,
		&metadata.PQCEncryptedData,
		&metadata.AESKey,
		&metadata.AESNonce,
		&metadata.AESAuthTag,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query encryption metadata: %w", err)
	}

	if pqcEncryptionTime.Valid {
		metadata.PQCEncryptionTime = pqcEncryptionTime.Time
	}

	return &metadata, nil
}

// updateEmailEncryptionMetadata updates encryption metadata in the database
func (pi *PQCIntegration) updateEmailEncryptionMetadata(emailID string, metadata *EmailEncryptionMetadata) error {
	query := `
		UPDATE emails 
		SET 
			pqc_enabled = ?,
			encryption_version = ?,
			pqc_key_id = ?,
			pqc_encryption_time = ?,
			pqc_encrypted_data = ?
		WHERE id = ?
	`

	_, err := pi.db.Exec(query,
		metadata.PQCEnabled,
		metadata.EncryptionVersion,
		metadata.PQCKeyID,
		metadata.PQCEncryptionTime,
		metadata.PQCEncryptedData,
		emailID,
	)

	if err != nil {
		return fmt.Errorf("failed to update encryption metadata: %w", err)
	}

	return nil
}

// getEmailsForMigration retrieves emails that need migration to PQC
func (pi *PQCIntegration) getEmailsForMigration(batchSize int) ([]string, error) {
	query := `
		SELECT id 
		FROM emails 
		WHERE pqc_enabled = FALSE 
		AND encryption_version = 'AES-256-GCM'
		LIMIT ?
	`

	rows, err := pi.db.Query(query, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails for migration: %w", err)
	}
	defer rows.Close()

	var emailIDs []string
	for rows.Next() {
		var emailID string
		if err := rows.Scan(&emailID); err != nil {
			return nil, fmt.Errorf("failed to scan email ID: %w", err)
		}
		emailIDs = append(emailIDs, emailID)
	}

	return emailIDs, nil
}

// logPerformanceMetric logs a performance metric to the database
func (pi *PQCIntegration) logPerformanceMetric(operation string, duration time.Duration, dataSize int, kyberLevel int, success bool, err error, context string) {
	query := `
		INSERT INTO pqc_performance_metrics (
			id, operation, duration_ms, data_size, kyber_level, 
			hsm_enabled, success, error_message, context, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	_, logErr := pi.db.Exec(query,
		generateUUID(),
		operation,
		duration.Milliseconds(),
		dataSize,
		kyberLevel,
		pi.service.GetConfig().HSMEnabled,
		success,
		errorMessage,
		context,
		time.Now(),
	)

	if logErr != nil {
		log.Printf("Failed to log performance metric: %v", logErr)
	}
}

// generateUUID generates a simple UUID (in production, use proper UUID library)
func generateUUID() string {
	return fmt.Sprintf("pqc_%d", time.Now().UnixNano())
}

// GetMigrationStats returns statistics about PQC migration
func (pi *PQCIntegration) GetMigrationStats() (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_emails,
			COUNT(CASE WHEN pqc_enabled = TRUE THEN 1 END) as pqc_emails,
			COUNT(CASE WHEN pqc_enabled = FALSE THEN 1 END) as aes_emails,
			COUNT(CASE WHEN encryption_version = 'PQC-HYBRID' THEN 1 END) as hybrid_emails
		FROM emails
	`

	var stats struct {
		TotalEmails  int
		PQCEmails    int
		AESEmails    int
		HybridEmails int
	}

	err := pi.db.QueryRow(query).Scan(
		&stats.TotalEmails,
		&stats.PQCEmails,
		&stats.AESEmails,
		&stats.HybridEmails,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get migration stats: %w", err)
	}

	adoptionRate := 0.0
	if stats.TotalEmails > 0 {
		adoptionRate = float64(stats.PQCEmails) / float64(stats.TotalEmails) * 100
	}

	return map[string]interface{}{
		"total_emails":     stats.TotalEmails,
		"pqc_emails":       stats.PQCEmails,
		"aes_emails":       stats.AESEmails,
		"hybrid_emails":    stats.HybridEmails,
		"adoption_rate":    adoptionRate,
		"pqc_enabled":      pi.service.IsEnabled(),
		"migration_needed": stats.AESEmails,
	}, nil
}
