package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DisasterRecoveryConfig holds configuration for disaster recovery operations
type DisasterRecoveryConfig struct {
	BackupDir           string `json:"backup_dir"`
	EncryptionKey       string `json:"encryption_key"`
	RetentionDays       int    `json:"retention_days"`
	MaxBackupSize       int64  `json:"max_backup_size"`
	CompressionEnabled  bool   `json:"compression_enabled"`
	VerifyBackups       bool   `json:"verify_backups"`
	AutoBackupInterval  string `json:"auto_backup_interval"`
	RecoveryTimeout     string `json:"recovery_timeout"`
	ParallelRestore     bool   `json:"parallel_restore"`
	BackupEncryption    bool   `json:"backup_encryption"`
	IntegrityChecks     bool   `json:"integrity_checks"`
	NotificationEnabled bool   `json:"notification_enabled"`
}

// BackupMetadata contains metadata about a backup
type BackupMetadata struct {
	BackupID        string    `json:"backup_id"`
	Timestamp       time.Time `json:"timestamp"`
	Type            string    `json:"type"` // zkid, pqc, audit, admin, full
	Size            int64     `json:"size"`
	Checksum        string    `json:"checksum"`
	Encrypted       bool      `json:"encrypted"`
	Compressed      bool      `json:"compressed"`
	DatabaseVersion string    `json:"database_version"`
	ZKIDMappings    int       `json:"zkid_mappings"`
	PQCKeys         int       `json:"pqc_keys"`
	AuditLogs       int       `json:"audit_logs"`
	AdminSessions   int       `json:"admin_sessions"`
	Status          string    `json:"status"` // created, verified, failed
	Error           string    `json:"error,omitempty"`
}

// DisasterRecoveryManager handles all disaster recovery operations
type DisasterRecoveryManager struct {
	config     DisasterRecoveryConfig
	db         *sql.DB
	backupPath string
	logger     *log.Logger
}

// NewDisasterRecoveryManager creates a new disaster recovery manager
func NewDisasterRecoveryManager(configPath string) (*DisasterRecoveryManager, error) {
	// Load configuration
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config DisasterRecoveryConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Initialize database connection
	db, err := sql.Open("sqlite", "secure_email_mvp.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Setup logger
	logger := log.New(os.Stdout, "[DISASTER_RECOVERY] ", log.LstdFlags|log.Lshortfile)

	return &DisasterRecoveryManager{
		config:     config,
		db:         db,
		backupPath: config.BackupDir,
		logger:     logger,
	}, nil
}

// CreateFullBackup creates a complete system backup
func (dr *DisasterRecoveryManager) CreateFullBackup() (*BackupMetadata, error) {
	dr.logger.Println("Starting full system backup...")

	backupID := generateBackupID()
	backupDir := filepath.Join(dr.backupPath, backupID)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	metadata := &BackupMetadata{
		BackupID:        backupID,
		Timestamp:       time.Now(),
		Type:            "full",
		Encrypted:       dr.config.BackupEncryption,
		Compressed:      dr.config.CompressionEnabled,
		DatabaseVersion: "1.0.0", // This should be retrieved from actual DB
		Status:          "created",
	}

	// Backup ZKID mappings
	if err := dr.backupZKIDMappings(backupDir, metadata); err != nil {
		metadata.Status = "failed"
		metadata.Error = fmt.Sprintf("ZKID backup failed: %v", err)
		return metadata, err
	}

	// Backup PQC keys
	if err := dr.backupPQCKeys(backupDir, metadata); err != nil {
		metadata.Status = "failed"
		metadata.Error = fmt.Sprintf("PQC backup failed: %v", err)
		return metadata, err
	}

	// Backup audit logs
	if err := dr.backupAuditLogs(backupDir, metadata); err != nil {
		metadata.Status = "failed"
		metadata.Error = fmt.Sprintf("Audit backup failed: %v", err)
		return metadata, err
	}

	// Backup admin sessions
	if err := dr.backupAdminSessions(backupDir, metadata); err != nil {
		metadata.Status = "failed"
		metadata.Error = fmt.Sprintf("Admin sessions backup failed: %v", err)
		return metadata, err
	}

	// Create metadata file
	if err := dr.saveBackupMetadata(backupDir, metadata); err != nil {
		metadata.Status = "failed"
		metadata.Error = fmt.Sprintf("Metadata save failed: %v", err)
		return metadata, err
	}

	// Verify backup if enabled
	if dr.config.VerifyBackups {
		if err := dr.verifyBackup(backupDir, metadata); err != nil {
			metadata.Status = "failed"
			metadata.Error = fmt.Sprintf("Backup verification failed: %v", err)
			return metadata, err
		}
		metadata.Status = "verified"
	} else {
		metadata.Status = "created"
	}

	dr.logger.Printf("Full backup completed successfully: %s", backupID)
	return metadata, nil
}

// backupZKIDMappings backs up ZKID encrypted mappings
func (dr *DisasterRecoveryManager) backupZKIDMappings(backupDir string, metadata *BackupMetadata) error {
	dr.logger.Println("Backing up ZKID mappings...")

	// Query all ZKID mappings
	rows, err := dr.db.Query(`
		SELECT uuid, encrypted_email, encrypted_mapping, created_at, expires_at, is_active
		FROM zkid_mappings
		WHERE is_active = 1
	`)
	if err != nil {
		return fmt.Errorf("failed to query ZKID mappings: %w", err)
	}
	defer rows.Close()

	var mappings []map[string]interface{}
	count := 0

	for rows.Next() {
		var uuid, encryptedEmail, encryptedMapping string
		var createdAt, expiresAt time.Time
		var isActive bool

		if err := rows.Scan(&uuid, &encryptedEmail, &encryptedMapping, &createdAt, &expiresAt, &isActive); err != nil {
			return fmt.Errorf("failed to scan ZKID mapping: %w", err)
		}

		mapping := map[string]interface{}{
			"uuid":              uuid,
			"encrypted_email":   encryptedEmail,
			"encrypted_mapping": encryptedMapping,
			"created_at":        createdAt,
			"expires_at":        expiresAt,
			"is_active":         isActive,
		}

		mappings = append(mappings, mapping)
		count++
	}

	metadata.ZKIDMappings = count

	// Save to file
	zkidFile := filepath.Join(backupDir, "zkid_mappings.json")
	if err := dr.saveEncryptedData(zkidFile, mappings); err != nil {
		return fmt.Errorf("failed to save ZKID mappings: %w", err)
	}

	dr.logger.Printf("ZKID mappings backup completed: %d mappings", count)
	return nil
}

// backupPQCKeys backs up PQC encryption keys
func (dr *DisasterRecoveryManager) backupPQCKeys(backupDir string, metadata *BackupMetadata) error {
	dr.logger.Println("Backing up PQC keys...")

	// Query all PQC keys
	rows, err := dr.db.Query(`
		SELECT key_id, key_type, encrypted_key_data, key_strength, created_at, expires_at, is_active
		FROM pqc_keys
		WHERE is_active = 1
	`)
	if err != nil {
		return fmt.Errorf("failed to query PQC keys: %w", err)
	}
	defer rows.Close()

	var keys []map[string]interface{}
	count := 0

	for rows.Next() {
		var keyID, keyType, encryptedKeyData string
		var keyStrength int
		var createdAt, expiresAt time.Time
		var isActive bool

		if err := rows.Scan(&keyID, &keyType, &encryptedKeyData, &keyStrength, &createdAt, &expiresAt, &isActive); err != nil {
			return fmt.Errorf("failed to scan PQC key: %w", err)
		}

		key := map[string]interface{}{
			"key_id":             keyID,
			"key_type":           keyType,
			"encrypted_key_data": encryptedKeyData,
			"key_strength":       keyStrength,
			"created_at":         createdAt,
			"expires_at":         expiresAt,
			"is_active":          isActive,
		}

		keys = append(keys, key)
		count++
	}

	metadata.PQCKeys = count

	// Save to file
	pqcFile := filepath.Join(backupDir, "pqc_keys.json")
	if err := dr.saveEncryptedData(pqcFile, keys); err != nil {
		return fmt.Errorf("failed to save PQC keys: %w", err)
	}

	dr.logger.Printf("PQC keys backup completed: %d keys", count)
	return nil
}

// backupAuditLogs backs up audit logs
func (dr *DisasterRecoveryManager) backupAuditLogs(backupDir string, metadata *BackupMetadata) error {
	dr.logger.Println("Backing up audit logs...")

	// Query recent audit logs (last 30 days)
	rows, err := dr.db.Query(`
		SELECT log_id, user_uuid, action, resource_type, resource_id, ip_address, user_agent, timestamp, details
		FROM audit_logs
		WHERE timestamp >= datetime('now', '-30 days')
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []map[string]interface{}
	count := 0

	for rows.Next() {
		var logID, userUUID, action, resourceType, resourceID, ipAddress, userAgent, details string
		var timestamp time.Time

		if err := rows.Scan(&logID, &userUUID, &action, &resourceType, &resourceID, &ipAddress, &userAgent, &timestamp, &details); err != nil {
			return fmt.Errorf("failed to scan audit log: %w", err)
		}

		log := map[string]interface{}{
			"log_id":        logID,
			"user_uuid":     userUUID,
			"action":        action,
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
			"timestamp":     timestamp,
			"details":       details,
		}

		logs = append(logs, log)
		count++
	}

	metadata.AuditLogs = count

	// Save to file
	auditFile := filepath.Join(backupDir, "audit_logs.json")
	if err := dr.saveEncryptedData(auditFile, logs); err != nil {
		return fmt.Errorf("failed to save audit logs: %w", err)
	}

	dr.logger.Printf("Audit logs backup completed: %d logs", count)
	return nil
}

// backupAdminSessions backs up admin session data
func (dr *DisasterRecoveryManager) backupAdminSessions(backupDir string, metadata *BackupMetadata) error {
	dr.logger.Println("Backing up admin sessions...")

	// Query active admin sessions
	rows, err := dr.db.Query(`
		SELECT session_id, admin_uuid, role, created_at, expires_at, last_activity, ip_address, user_agent
		FROM admin_sessions
		WHERE expires_at > datetime('now')
	`)
	if err != nil {
		return fmt.Errorf("failed to query admin sessions: %w", err)
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	count := 0

	for rows.Next() {
		var sessionID, adminUUID, role, ipAddress, userAgent string
		var createdAt, expiresAt, lastActivity time.Time

		if err := rows.Scan(&sessionID, &adminUUID, &role, &createdAt, &expiresAt, &lastActivity, &ipAddress, &userAgent); err != nil {
			return fmt.Errorf("failed to scan admin session: %w", err)
		}

		session := map[string]interface{}{
			"session_id":    sessionID,
			"admin_uuid":    adminUUID,
			"role":          role,
			"created_at":    createdAt,
			"expires_at":    expiresAt,
			"last_activity": lastActivity,
			"ip_address":    ipAddress,
			"user_agent":    userAgent,
		}

		sessions = append(sessions, session)
		count++
	}

	metadata.AdminSessions = count

	// Save to file
	sessionsFile := filepath.Join(backupDir, "admin_sessions.json")
	if err := dr.saveEncryptedData(sessionsFile, sessions); err != nil {
		return fmt.Errorf("failed to save admin sessions: %w", err)
	}

	dr.logger.Printf("Admin sessions backup completed: %d sessions", count)
	return nil
}

// RestoreFromBackup restores the system from a backup
func (dr *DisasterRecoveryManager) RestoreFromBackup(backupID string) error {
	dr.logger.Printf("Starting restore from backup: %s", backupID)

	backupDir := filepath.Join(dr.backupPath, backupID)

	// Verify backup exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	// Load backup metadata
	metadata, err := dr.loadBackupMetadata(backupDir)
	if err != nil {
		return fmt.Errorf("failed to load backup metadata: %w", err)
	}

	dr.logger.Printf("Restoring backup: %s (Type: %s, Created: %s)",
		metadata.BackupID, metadata.Type, metadata.Timestamp.Format(time.RFC3339))

	// Begin transaction
	tx, err := dr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Restore ZKID mappings
	if err := dr.restoreZKIDMappings(backupDir, tx); err != nil {
		return fmt.Errorf("failed to restore ZKID mappings: %w", err)
	}

	// Restore PQC keys
	if err := dr.restorePQCKeys(backupDir, tx); err != nil {
		return fmt.Errorf("failed to restore PQC keys: %w", err)
	}

	// Restore audit logs
	if err := dr.restoreAuditLogs(backupDir, tx); err != nil {
		return fmt.Errorf("failed to restore audit logs: %w", err)
	}

	// Restore admin sessions
	if err := dr.restoreAdminSessions(backupDir, tx); err != nil {
		return fmt.Errorf("failed to restore admin sessions: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit restore transaction: %w", err)
	}

	dr.logger.Printf("Restore completed successfully from backup: %s", backupID)
	return nil
}

// restoreZKIDMappings restores ZKID mappings from backup
func (dr *DisasterRecoveryManager) restoreZKIDMappings(backupDir string, tx *sql.Tx) error {
	dr.logger.Println("Restoring ZKID mappings...")

	zkidFile := filepath.Join(backupDir, "zkid_mappings.json")
	var mappings []map[string]interface{}

	if err := dr.loadEncryptedData(zkidFile, &mappings); err != nil {
		return fmt.Errorf("failed to load ZKID mappings: %w", err)
	}

	// Clear existing mappings (optional - depends on restore strategy)
	_, err := tx.Exec("DELETE FROM zkid_mappings")
	if err != nil {
		return fmt.Errorf("failed to clear existing ZKID mappings: %w", err)
	}

	// Restore mappings
	stmt, err := tx.Prepare(`
		INSERT INTO zkid_mappings (uuid, encrypted_email, encrypted_mapping, created_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare ZKID restore statement: %w", err)
	}
	defer stmt.Close()

	for _, mapping := range mappings {
		_, err := stmt.Exec(
			mapping["uuid"],
			mapping["encrypted_email"],
			mapping["encrypted_mapping"],
			mapping["created_at"],
			mapping["expires_at"],
			mapping["is_active"],
		)
		if err != nil {
			return fmt.Errorf("failed to restore ZKID mapping: %w", err)
		}
	}

	dr.logger.Printf("ZKID mappings restored: %d mappings", len(mappings))
	return nil
}

// restorePQCKeys restores PQC keys from backup
func (dr *DisasterRecoveryManager) restorePQCKeys(backupDir string, tx *sql.Tx) error {
	dr.logger.Println("Restoring PQC keys...")

	pqcFile := filepath.Join(backupDir, "pqc_keys.json")
	var keys []map[string]interface{}

	if err := dr.loadEncryptedData(pqcFile, &keys); err != nil {
		return fmt.Errorf("failed to load PQC keys: %w", err)
	}

	// Clear existing keys (optional - depends on restore strategy)
	_, err := tx.Exec("DELETE FROM pqc_keys")
	if err != nil {
		return fmt.Errorf("failed to clear existing PQC keys: %w", err)
	}

	// Restore keys
	stmt, err := tx.Prepare(`
		INSERT INTO pqc_keys (key_id, key_type, encrypted_key_data, key_strength, created_at, expires_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare PQC restore statement: %w", err)
	}
	defer stmt.Close()

	for _, key := range keys {
		_, err := stmt.Exec(
			key["key_id"],
			key["key_type"],
			key["encrypted_key_data"],
			key["key_strength"],
			key["created_at"],
			key["expires_at"],
			key["is_active"],
		)
		if err != nil {
			return fmt.Errorf("failed to restore PQC key: %w", err)
		}
	}

	dr.logger.Printf("PQC keys restored: %d keys", len(keys))
	return nil
}

// restoreAuditLogs restores audit logs from backup
func (dr *DisasterRecoveryManager) restoreAuditLogs(backupDir string, tx *sql.Tx) error {
	dr.logger.Println("Restoring audit logs...")

	auditFile := filepath.Join(backupDir, "audit_logs.json")
	var logs []map[string]interface{}

	if err := dr.loadEncryptedData(auditFile, &logs); err != nil {
		return fmt.Errorf("failed to load audit logs: %w", err)
	}

	// Clear existing logs (optional - depends on restore strategy)
	_, err := tx.Exec("DELETE FROM audit_logs")
	if err != nil {
		return fmt.Errorf("failed to clear existing audit logs: %w", err)
	}

	// Restore logs
	stmt, err := tx.Prepare(`
		INSERT INTO audit_logs (log_id, user_uuid, action, resource_type, resource_id, ip_address, user_agent, timestamp, details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare audit restore statement: %w", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		_, err := stmt.Exec(
			log["log_id"],
			log["user_uuid"],
			log["action"],
			log["resource_type"],
			log["resource_id"],
			log["ip_address"],
			log["user_agent"],
			log["timestamp"],
			log["details"],
		)
		if err != nil {
			return fmt.Errorf("failed to restore audit log: %w", err)
		}
	}

	dr.logger.Printf("Audit logs restored: %d logs", len(logs))
	return nil
}

// restoreAdminSessions restores admin sessions from backup
func (dr *DisasterRecoveryManager) restoreAdminSessions(backupDir string, tx *sql.Tx) error {
	dr.logger.Println("Restoring admin sessions...")

	sessionsFile := filepath.Join(backupDir, "admin_sessions.json")
	var sessions []map[string]interface{}

	if err := dr.loadEncryptedData(sessionsFile, &sessions); err != nil {
		return fmt.Errorf("failed to load admin sessions: %w", err)
	}

	// Clear existing sessions (optional - depends on restore strategy)
	_, err := tx.Exec("DELETE FROM admin_sessions")
	if err != nil {
		return fmt.Errorf("failed to clear existing admin sessions: %w", err)
	}

	// Restore sessions
	stmt, err := tx.Prepare(`
		INSERT INTO admin_sessions (session_id, admin_uuid, role, created_at, expires_at, last_activity, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare admin sessions restore statement: %w", err)
	}
	defer stmt.Close()

	for _, session := range sessions {
		_, err := stmt.Exec(
			session["session_id"],
			session["admin_uuid"],
			session["role"],
			session["created_at"],
			session["expires_at"],
			session["last_activity"],
			session["ip_address"],
			session["user_agent"],
		)
		if err != nil {
			return fmt.Errorf("failed to restore admin session: %w", err)
		}
	}

	dr.logger.Printf("Admin sessions restored: %d sessions", len(sessions))
	return nil
}

// ListBackups lists all available backups
func (dr *DisasterRecoveryManager) ListBackups() ([]BackupMetadata, error) {
	var backups []BackupMetadata

	entries, err := os.ReadDir(dr.backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			backupDir := filepath.Join(dr.backupPath, entry.Name())
			metadata, err := dr.loadBackupMetadata(backupDir)
			if err != nil {
				dr.logger.Printf("Warning: failed to load metadata for backup %s: %v", entry.Name(), err)
				continue
			}
			backups = append(backups, *metadata)
		}
	}

	return backups, nil
}

// CleanupOldBackups removes backups older than retention period
func (dr *DisasterRecoveryManager) CleanupOldBackups() error {
	dr.logger.Printf("Cleaning up backups older than %d days", dr.config.RetentionDays)

	backups, err := dr.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -dr.config.RetentionDays)
	deletedCount := 0

	for _, backup := range backups {
		if backup.Timestamp.Before(cutoffTime) {
			backupDir := filepath.Join(dr.backupPath, backup.BackupID)
			if err := os.RemoveAll(backupDir); err != nil {
				dr.logger.Printf("Warning: failed to delete old backup %s: %v", backup.BackupID, err)
				continue
			}
			deletedCount++
			dr.logger.Printf("Deleted old backup: %s", backup.BackupID)
		}
	}

	dr.logger.Printf("Cleanup completed: %d backups deleted", deletedCount)
	return nil
}

// Helper functions

func generateBackupID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("backup_%x_%d", b, time.Now().Unix())
}

func (dr *DisasterRecoveryManager) saveBackupMetadata(backupDir string, metadata *BackupMetadata) error {
	metadataFile := filepath.Join(backupDir, "metadata.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func (dr *DisasterRecoveryManager) loadBackupMetadata(backupDir string) (*BackupMetadata, error) {
	metadataFile := filepath.Join(backupDir, "metadata.json")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

func (dr *DisasterRecoveryManager) saveEncryptedData(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if dr.config.BackupEncryption {
		encryptedData, err := dr.encryptData(jsonData)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		return os.WriteFile(filename, encryptedData, 0644)
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func (dr *DisasterRecoveryManager) loadEncryptedData(filename string, data interface{}) error {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var jsonData []byte
	if dr.config.BackupEncryption {
		decryptedData, err := dr.decryptData(fileData)
		if err != nil {
			return fmt.Errorf("failed to decrypt data: %w", err)
		}
		jsonData = decryptedData
	} else {
		jsonData = fileData
	}

	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func (dr *DisasterRecoveryManager) encryptData(data []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(dr.config.EncryptionKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (dr *DisasterRecoveryManager) decryptData(data []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(dr.config.EncryptionKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (dr *DisasterRecoveryManager) verifyBackup(backupDir string, metadata *BackupMetadata) error {
	dr.logger.Println("Verifying backup integrity...")

	// Verify all required files exist
	requiredFiles := []string{"zkid_mappings.json", "pqc_keys.json", "audit_logs.json", "admin_sessions.json", "metadata.json"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(backupDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required backup file missing: %s", file)
		}
	}

	// Verify data integrity by attempting to load each file
	var testData interface{}

	// Test ZKID mappings
	if err := dr.loadEncryptedData(filepath.Join(backupDir, "zkid_mappings.json"), &testData); err != nil {
		return fmt.Errorf("zkid mappings integrity check failed: %w", err)
	}

	// Test PQC keys
	if err := dr.loadEncryptedData(filepath.Join(backupDir, "pqc_keys.json"), &testData); err != nil {
		return fmt.Errorf("pqc keys integrity check failed: %w", err)
	}

	// Test audit logs
	if err := dr.loadEncryptedData(filepath.Join(backupDir, "audit_logs.json"), &testData); err != nil {
		return fmt.Errorf("audit logs integrity check failed: %w", err)
	}

	// Test admin sessions
	if err := dr.loadEncryptedData(filepath.Join(backupDir, "admin_sessions.json"), &testData); err != nil {
		return fmt.Errorf("admin sessions integrity check failed: %w", err)
	}

	dr.logger.Println("Backup verification completed successfully")
	return nil
}

// Close closes the disaster recovery manager
func (dr *DisasterRecoveryManager) Close() error {
	return dr.db.Close()
}

func main() {
	// Example usage
	configPath := "scripts/operational/disaster_recovery_config.json"

	dr, err := NewDisasterRecoveryManager(configPath)
	if err != nil {
		log.Fatalf("Failed to create disaster recovery manager: %v", err)
	}
	defer dr.Close()

	// Create a full backup
	metadata, err := dr.CreateFullBackup()
	if err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}

	fmt.Printf("Backup created successfully: %s\n", metadata.BackupID)
	fmt.Printf("ZKID mappings: %d\n", metadata.ZKIDMappings)
	fmt.Printf("PQC keys: %d\n", metadata.PQCKeys)
	fmt.Printf("Audit logs: %d\n", metadata.AuditLogs)
	fmt.Printf("Admin sessions: %d\n", metadata.AdminSessions)

	// List all backups
	backups, err := dr.ListBackups()
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}

	fmt.Printf("\nAvailable backups:\n")
	for _, backup := range backups {
		fmt.Printf("- %s (%s) - %s\n", backup.BackupID, backup.Type, backup.Timestamp.Format(time.RFC3339))
	}

	// Cleanup old backups
	if err := dr.CleanupOldBackups(); err != nil {
		log.Printf("Warning: failed to cleanup old backups: %v", err)
	}
}
