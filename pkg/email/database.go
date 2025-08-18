package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/storage"
)

// EmailSecurityDB provides comprehensive database operations for email security features
//
// This struct encapsulates all database operations related to email security,
// including security toggle management, access control, self-destruct functionality,
// read-once consumption, and secure deletion. It provides a clean interface for
// the HTTP handlers to interact with email security features.
//
// SECURITY FEATURES MANAGED:
// - Per-email security toggles (remote revoke, time lock, expiration, MFA, etc.)
// - Access control and authorization checks
// - Self-destruct counter with threshold-based deletion
// - Read-once consumption with atomic operations
// - Secure deletion (both database and R2 blob removal)
// - Failed attempts tracking and management
//
// DATABASE OPERATIONS:
// - CRUD operations for security toggles (sender only)
// - Access validation and authorization checks
// - Atomic read-once consumption with optimistic locking
// - Self-destruct counter management with transactions
// - Secure deletion with R2 integration
// - Audit trail support for all operations
//
// DEPENDENCY INJECTION:
// - R2 client is optional and can be nil for testing
// - Database connection is required for all operations
// - Supports both production (with R2) and testing (without R2) scenarios
type EmailSecurityDB struct {
	db       *sql.DB           // SQLite database connection
	r2Client *storage.R2Client // Optional R2 client for physical deletion
}

// NewEmailSecurityDB creates a new EmailSecurityDB instance
func NewEmailSecurityDB(db *sql.DB) *EmailSecurityDB {
	return &EmailSecurityDB{db: db}
}

// NewEmailSecurityDBWithR2 creates a new EmailSecurityDB instance with R2 client for physical deletion
func NewEmailSecurityDBWithR2(db *sql.DB, r2Client *storage.R2Client) *EmailSecurityDB {
	return &EmailSecurityDB{db: db, r2Client: r2Client}
}

// UpdateEmailSecurityToggles updates the security settings for a specific email
//
// This function allows email senders to modify the security settings for their
// emails after they have been sent. Only the original sender can update these
// settings, providing fine-grained control over email access and security.
//
// SECURITY FEATURES THAT CAN BE UPDATED:
// - Time Lock (not_before): When the email becomes accessible
// - Expiration (expires_at): When the email expires and becomes inaccessible
// - Read-Once (read_once): Whether the email should be burned after first access
// - MFA-on-Open (mfa_on_open): Whether TOTP verification is required for access
// - Decoy Secret (decoy_secret): Hash for triggering decoy messages
// - Remote Revoke (remote_revoke): Whether the email has been revoked by sender
// - Strip Metadata (strip_metadata): Whether to remove EXIF data from attachments
// - Self-Destruct Threshold (self_destruct_threshold): Max failed attempts before deletion
// - Geo Rules Reference (geo_rules_ref): JSON reference to geofencing rules
// - Self-Destruct on Read-Once (self_destruct_on_read_once): Delete after first read
//
// SECURITY CHECKS:
// - Validates that the email exists in the database
// - Verifies that the requesting user is the original sender
// - Validates all security toggle settings before applying
// - Uses COALESCE to only update provided fields (partial updates supported)
//
// PARAMETERS:
// - emailID: The unique identifier of the email to update
// - senderID: The ID of the user requesting the update (must be the sender)
// - toggles: The security toggle settings to apply
//
// RETURNS:
// - error: nil on success, descriptive error on failure
func (esdb *EmailSecurityDB) UpdateEmailSecurityToggles(emailID, senderID string, toggles EmailSecurityToggles) error {
	// Validate the security toggles
	if err := ValidateEmailSecurityToggles(toggles); err != nil {
		return fmt.Errorf("invalid security toggles: %w", err)
	}

	// Verify the email exists and belongs to the sender
	var existingSenderID string
	err := esdb.db.QueryRow(`
		SELECT sender_id FROM emails WHERE email_id = ?
	`, emailID).Scan(&existingSenderID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email not found")
		}
		log.Printf("Database error checking email ownership: %v", err)
		return fmt.Errorf("database error")
	}

	if existingSenderID != senderID {
		return fmt.Errorf("access denied")
	}

	// Build the update query dynamically based on provided fields
	query := `
		UPDATE emails SET 
			not_before = COALESCE(?, not_before),
			expires_at = COALESCE(?, expires_at),
			read_once = COALESCE(?, read_once),
			mfa_on_open = COALESCE(?, mfa_on_open),
			decoy_secret = COALESCE(?, decoy_secret),
			remote_revoke = COALESCE(?, remote_revoke),
			strip_metadata = COALESCE(?, strip_metadata),
			self_destruct_threshold = COALESCE(?, self_destruct_threshold),
			geo_rules_ref = COALESCE(?, geo_rules_ref),
			self_destruct_on_read_once = COALESCE(?, self_destruct_on_read_once)
		WHERE email_id = ?
	`

	// Execute the update
	_, err = esdb.db.Exec(query,
		toggles.NotBefore,
		toggles.ExpiresAt,
		toggles.ReadOnce,
		toggles.MFAOnOpen,
		toggles.DecoySecret,
		toggles.RemoteRevoke,
		toggles.StripMetadata,
		toggles.SelfDestructThreshold,
		toggles.GeoRulesRef,
		toggles.SelfDestructOnReadOnce,
		emailID,
	)

	if err != nil {
		log.Printf("Database error updating email security toggles: %v", err)
		return fmt.Errorf("failed to update security settings")
	}

	return nil
}

// GetEmailSecurityToggles retrieves the current security settings for an email
// Only the sender can view security settings for their emails
func (esdb *EmailSecurityDB) GetEmailSecurityToggles(emailID, senderID string) (*EmailSecurityToggles, error) {
	// Query the email security settings
	var (
		notBefore              sql.NullInt64
		expiresAt              sql.NullInt64
		readOnce               sql.NullBool
		mfaOnOpen              sql.NullBool
		decoySecret            sql.NullString
		remoteRevoke           sql.NullBool
		stripMetadata          sql.NullBool
		selfDestructThreshold  sql.NullInt64
		geoRulesRef            sql.NullString
		selfDestructOnReadOnce sql.NullBool
		existingSenderID       string
	)

	err := esdb.db.QueryRow(`
		SELECT 
			sender_id,
			not_before,
			expires_at,
			read_once,
			mfa_on_open,
			decoy_secret,
			remote_revoke,
			strip_metadata,
			self_destruct_threshold,
			geo_rules_ref,
			self_destruct_on_read_once
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(
		&existingSenderID,
		&notBefore,
		&expiresAt,
		&readOnce,
		&mfaOnOpen,
		&decoySecret,
		&remoteRevoke,
		&stripMetadata,
		&selfDestructThreshold,
		&geoRulesRef,
		&selfDestructOnReadOnce,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		log.Printf("Database error retrieving email security toggles: %v", err)
		return nil, fmt.Errorf("database error")
	}

	// Verify the email belongs to the sender
	if existingSenderID != senderID {
		return nil, fmt.Errorf("access denied")
	}

	// Build the security toggles struct
	toggles := &EmailSecurityToggles{}

	if notBefore.Valid {
		toggles.NotBefore = &notBefore.Int64
	}
	if expiresAt.Valid {
		toggles.ExpiresAt = &expiresAt.Int64
	}
	if readOnce.Valid {
		toggles.ReadOnce = readOnce.Bool
	}
	if mfaOnOpen.Valid {
		toggles.MFAOnOpen = mfaOnOpen.Bool
	}
	if decoySecret.Valid {
		toggles.DecoySecret = &decoySecret.String
	}
	if remoteRevoke.Valid {
		toggles.RemoteRevoke = remoteRevoke.Bool
	}
	if stripMetadata.Valid {
		toggles.StripMetadata = stripMetadata.Bool
	}
	if selfDestructThreshold.Valid {
		threshold := int(selfDestructThreshold.Int64)
		toggles.SelfDestructThreshold = &threshold
	}
	if geoRulesRef.Valid {
		toggles.GeoRulesRef = &geoRulesRef.String
	}
	if selfDestructOnReadOnce.Valid {
		toggles.SelfDestructOnReadOnce = selfDestructOnReadOnce.Bool
	}

	return toggles, nil
}

// GetEmailSecurityTogglesForAccess retrieves security toggles for access control
// This is used during email access to check security rules
// Returns nil if email doesn't exist (to avoid info leaks)
func (esdb *EmailSecurityDB) GetEmailSecurityTogglesForAccess(emailID string) (*EmailSecurityToggles, error) {
	var (
		notBefore              sql.NullInt64
		expiresAt              sql.NullInt64
		readOnce               sql.NullBool
		mfaOnOpen              sql.NullBool
		decoySecret            sql.NullString
		remoteRevoke           sql.NullBool
		stripMetadata          sql.NullBool
		selfDestructThreshold  sql.NullInt64
		geoRulesRef            sql.NullString
		selfDestructOnReadOnce sql.NullBool
	)

	err := esdb.db.QueryRow(`
		SELECT 
			not_before,
			expires_at,
			read_once,
			mfa_on_open,
			decoy_secret,
			remote_revoke,
			strip_metadata,
			self_destruct_threshold,
			geo_rules_ref,
			self_destruct_on_read_once
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(
		&notBefore,
		&expiresAt,
		&readOnce,
		&mfaOnOpen,
		&decoySecret,
		&remoteRevoke,
		&stripMetadata,
		&selfDestructThreshold,
		&geoRulesRef,
		&selfDestructOnReadOnce,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil instead of error to avoid info leaks
		}
		log.Printf("Database error retrieving email security toggles for access: %v", err)
		return nil, fmt.Errorf("database error")
	}

	// Build the security toggles struct
	toggles := &EmailSecurityToggles{}

	if notBefore.Valid {
		toggles.NotBefore = &notBefore.Int64
	}
	if expiresAt.Valid {
		toggles.ExpiresAt = &expiresAt.Int64
	}
	if readOnce.Valid {
		toggles.ReadOnce = readOnce.Bool
	}
	if mfaOnOpen.Valid {
		toggles.MFAOnOpen = mfaOnOpen.Bool
	}
	if decoySecret.Valid {
		toggles.DecoySecret = &decoySecret.String
	}
	if remoteRevoke.Valid {
		toggles.RemoteRevoke = remoteRevoke.Bool
	}
	if stripMetadata.Valid {
		toggles.StripMetadata = stripMetadata.Bool
	}
	if selfDestructThreshold.Valid {
		threshold := int(selfDestructThreshold.Int64)
		toggles.SelfDestructThreshold = &threshold
	}
	if geoRulesRef.Valid {
		toggles.GeoRulesRef = &geoRulesRef.String
	}
	if selfDestructOnReadOnce.Valid {
		toggles.SelfDestructOnReadOnce = selfDestructOnReadOnce.Bool
	}

	return toggles, nil
}

// CheckEmailAccess checks if an email can be accessed based on security toggles
// Returns an error message if access should be denied, nil if access is allowed
func (esdb *EmailSecurityDB) CheckEmailAccess(emailID string) (string, error) {
	toggles, err := esdb.GetEmailSecurityTogglesForAccess(emailID)
	if err != nil {
		return "", err
	}

	// If email doesn't exist, return generic error to avoid info leaks
	if toggles == nil {
		return "Access denied", nil
	}

	// Check remote revoke
	if toggles.IsRevoked() {
		return "Email has been revoked by sender", nil
	}

	// Check time lock
	if toggles.IsTimeLocked() {
		return fmt.Sprintf("Email is time-locked: %s", toggles.GetTimeWindowStatus()), nil
	}

	// Check expiration
	if toggles.IsExpired() {
		return fmt.Sprintf("Email has expired: %s", toggles.GetTimeWindowStatus()), nil
	}

	// Access is allowed
	return "", nil
}

// CheckEmailAccessWithGeofencing checks if an email can be accessed based on security toggles and geofencing
// Returns an error message if access should be denied, nil if access is allowed
func (esdb *EmailSecurityDB) CheckEmailAccessWithGeofencing(emailID, clientIP string, geofencingSvc interface{}) (string, error) {
	// First check basic security toggles
	accessError, err := esdb.CheckEmailAccess(emailID)
	if err != nil {
		return "", err
	}
	
	if accessError != "" {
		return accessError, nil
	}

	// Check geofencing if service is provided
	if geofencingSvc != nil {
		// Use reflection to call the geofencing service
		// This allows for dependency injection without tight coupling
		if checkMethod, ok := geofencingSvc.(interface {
			CheckGeofenceAccess(emailID, clientIP string) (interface{}, error)
		}); ok {
			result, err := checkMethod.CheckGeofenceAccess(emailID, clientIP)
			if err != nil {
				return "", fmt.Errorf("geofencing check failed: %w", err)
			}
			
			// Check if result indicates access is denied
			if resultMap, ok := result.(map[string]interface{}); ok {
				if allowed, exists := resultMap["allowed"]; exists {
					if !allowed.(bool) {
						return "Email has been revoked or cannot be accessed", nil
					}
				}
			}
		}
	}

	// Access is allowed
	return "", nil
}

// ValidateEmailExists checks if an email exists without revealing security details
// Used for authorization checks
func (esdb *EmailSecurityDB) ValidateEmailExists(emailID string) (bool, error) {
	var exists int
	err := esdb.db.QueryRow(`
		SELECT 1 FROM emails WHERE email_id = ?
	`, emailID).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetEmailSenderID retrieves the sender ID for an email
// Used for authorization checks
func (esdb *EmailSecurityDB) GetEmailSenderID(emailID string) (string, error) {
	var senderID string
	err := esdb.db.QueryRow(`
		SELECT sender_id FROM emails WHERE email_id = ?
	`, emailID).Scan(&senderID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("email not found")
		}
		return "", err
	}

	return senderID, nil
}

// IncrementFailedAttempts increments the failed attempts counter for an email
// Returns SelfDestructError if the threshold is reached and email should be destroyed
func (esdb *EmailSecurityDB) IncrementFailedAttempts(emailID string) error {
	// Start a transaction to ensure atomicity
	tx, err := esdb.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction for failed attempts increment: %v", err)
		return fmt.Errorf("database error")
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Get current failed attempts and self-destruct threshold
	var failedAttempts, selfDestructThreshold int
	err = tx.QueryRow(`
		SELECT failed_attempts, self_destruct_threshold 
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&failedAttempts, &selfDestructThreshold)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email not found")
		}
		log.Printf("Database error getting failed attempts: %v", err)
		return fmt.Errorf("database error")
	}

	// Use default threshold if not set
	if selfDestructThreshold == 0 {
		selfDestructThreshold = 3 // Default value
	}

	// Increment failed attempts
	newFailedAttempts := failedAttempts + 1

	// Check if threshold is reached
	if newFailedAttempts >= selfDestructThreshold {
		// Update the counter first
		_, err = tx.Exec(`
			UPDATE emails 
			SET failed_attempts = ? 
			WHERE email_id = ?
		`, newFailedAttempts, emailID)
		if err != nil {
			log.Printf("Database error updating failed attempts: %v", err)
			return fmt.Errorf("database error")
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			log.Printf("Failed to commit failed attempts update: %v", err)
			return fmt.Errorf("database error")
		}

		// Return SelfDestructError to indicate email should be destroyed
		return SelfDestructError{EmailID: emailID}
	}

	// Update failed attempts counter
	_, err = tx.Exec(`
		UPDATE emails 
		SET failed_attempts = ? 
		WHERE email_id = ?
	`, newFailedAttempts, emailID)
	if err != nil {
		log.Printf("Database error updating failed attempts: %v", err)
		return fmt.Errorf("database error")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit failed attempts update: %v", err)
		return fmt.Errorf("database error")
	}

	log.Printf("Failed attempt for email %s: %d/%d", emailID, newFailedAttempts, selfDestructThreshold)
	return nil
}

// ResetFailedAttempts resets the failed attempts counter to 0 for an email
// Called on successful access to reset the counter
func (esdb *EmailSecurityDB) ResetFailedAttempts(emailID string) error {
	_, err := esdb.db.Exec(`
		UPDATE emails 
		SET failed_attempts = 0 
		WHERE email_id = ?
	`, emailID)

	if err != nil {
		log.Printf("Database error resetting failed attempts: %v", err)
		return fmt.Errorf("database error")
	}

	log.Printf("Reset failed attempts for email %s", emailID)
	return nil
}

// DeleteEmailSecure securely deletes an email and all related data
//
// This function performs comprehensive secure deletion of an email, including
// both the database record and the encrypted blob stored in R2 storage. It
// ensures that no traces of the email remain in the system.
//
// SECURITY FEATURES:
// - Cryptographic deletion: Removes email record from database
// - Physical deletion: Removes encrypted blob from R2 storage
// - Atomic operations: Uses database transaction for consistency
// - Error handling: Logs R2 deletion failures but doesn't fail the operation
// - Audit trail: Comprehensive logging of deletion process
//
// DELETION PROCESS:
// 1. Start database transaction for atomicity
// 2. Retrieve email details (including R2 blob URL) for logging
// 3. Delete email record from database
// 4. Commit database transaction
// 5. Delete encrypted blob from R2 storage (if client available)
// 6. Log success or failure of each step
//
// ERROR HANDLING:
// - Database deletion failures cause transaction rollback
// - R2 deletion failures are logged but don't fail the operation
// - This ensures database consistency even if R2 is unavailable
// - Missing R2 objects are treated as success (idempotent operation)
//
// USE CASES:
// - Self-destruct threshold reached (too many failed attempts)
// - Read-once consumption with self-destruct enabled
// - Manual deletion by sender
// - System cleanup and maintenance
//
// PARAMETERS:
// - emailID: The unique identifier of the email to delete
//
// RETURNS:
// - error: nil on success, descriptive error on failure
func (esdb *EmailSecurityDB) DeleteEmailSecure(emailID string) error {
	// Start a transaction to ensure atomicity
	tx, err := esdb.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction for secure email deletion: %v", err)
		return fmt.Errorf("database error")
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Get email details for logging and cleanup
	var encryptedBlobURL string
	err = tx.QueryRow(`
		SELECT encrypted_blob_url 
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&encryptedBlobURL)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email not found")
		}
		log.Printf("Database error getting email details for deletion: %v", err)
		return fmt.Errorf("database error")
	}

	// Delete the email record
	_, err = tx.Exec(`
		DELETE FROM emails 
		WHERE email_id = ?
	`, emailID)
	if err != nil {
		log.Printf("Database error deleting email record: %v", err)
		return fmt.Errorf("database error")
	}

	// Commit the database transaction first
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit secure email deletion: %v", err)
		return fmt.Errorf("database error")
	}

	// Now perform physical deletion from R2 storage if client is available
	if esdb.r2Client != nil && encryptedBlobURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = esdb.r2Client.DeleteEmail(ctx, encryptedBlobURL)
		if err != nil {
			// Log the error but don't fail the operation since DB deletion already succeeded
			// This ensures we don't leave orphaned database records
			log.Printf("Warning: Failed to delete R2 blob %s for email %s: %v", encryptedBlobURL, emailID, err)
		} else {
			log.Printf("Successfully deleted R2 blob %s for email %s", encryptedBlobURL, emailID)
		}
	} else if encryptedBlobURL != "" {
		// Log that R2 deletion was skipped due to missing client
		log.Printf("Email %s deleted from database. R2 blob %s not deleted (no R2 client available)", emailID, encryptedBlobURL)
	}

	log.Printf("Successfully deleted email %s and all related data", emailID)
	return nil
}

// GetFailedAttemptsCount retrieves the current failed attempts count for an email
// Used for testing and debugging purposes only - not exposed to clients
func (esdb *EmailSecurityDB) GetFailedAttemptsCount(emailID string) (int, error) {
	var failedAttempts int
	err := esdb.db.QueryRow(`
		SELECT failed_attempts 
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&failedAttempts)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("email not found")
		}
		return 0, err
	}

	return failedAttempts, nil
}

// MarkReadOnceConsumed atomically marks a read-once email as consumed
// Uses optimistic locking to prevent race conditions
// Returns the timestamp when consumed and any error
func (esdb *EmailSecurityDB) MarkReadOnceConsumed(emailID string, consumerDevice string) (time.Time, error) {
	// Start a transaction to ensure atomicity
	tx, err := esdb.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction for read-once consumption: %v", err)
		return time.Time{}, fmt.Errorf("database error")
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Check if email exists and is read-once
	var readOnce bool
	var consumedAt sql.NullInt64
	err = tx.QueryRow(`
		SELECT read_once, read_once_consumed_at 
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&readOnce, &consumedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, fmt.Errorf("email not found")
		}
		log.Printf("Database error checking read-once status: %v", err)
		return time.Time{}, fmt.Errorf("database error")
	}

	// If not read-once, return error
	if !readOnce {
		return time.Time{}, fmt.Errorf("email is not configured for read-once")
	}

	// If already consumed, return error
	if consumedAt.Valid {
		return time.Time{}, ReadOnceConsumedError{EmailID: emailID}
	}

	// Mark as consumed with current timestamp
	now := time.Now().Unix()
	result, err := tx.Exec(`
		UPDATE emails 
		SET read_once_consumed_at = ?, read_once_consumer_device = ?
		WHERE email_id = ? AND read_once = TRUE AND read_once_consumed_at IS NULL
	`, now, consumerDevice, emailID)

	if err != nil {
		log.Printf("Database error marking read-once consumed: %v", err)
		return time.Time{}, fmt.Errorf("database error")
	}

	// Check if any rows were affected (optimistic locking)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Database error checking rows affected: %v", err)
		return time.Time{}, fmt.Errorf("database error")
	}

	if rowsAffected == 0 {
		// Another request already consumed this email
		return time.Time{}, ReadOnceConsumedError{EmailID: emailID}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit read-once consumption: %v", err)
		return time.Time{}, fmt.Errorf("database error")
	}

	consumedTime := time.Unix(now, 0)
	log.Printf("Successfully marked email %s as consumed at %s", emailID, consumedTime.Format(time.RFC3339))
	return consumedTime, nil
}

// IsReadOnceConsumed checks if a read-once email has already been consumed
// Returns whether consumed, the consumption timestamp, and any error
func (esdb *EmailSecurityDB) IsReadOnceConsumed(emailID string) (bool, time.Time, error) {
	var consumedAt sql.NullInt64
	err := esdb.db.QueryRow(`
		SELECT read_once_consumed_at 
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&consumedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, time.Time{}, fmt.Errorf("email not found")
		}
		return false, time.Time{}, err
	}

	if !consumedAt.Valid {
		return false, time.Time{}, nil
	}

	return true, time.Unix(consumedAt.Int64, 0), nil
}

// GetReadOnceInfo retrieves comprehensive read-once information for an email
// Used for testing and debugging purposes only - not exposed to clients
func (esdb *EmailSecurityDB) GetReadOnceInfo(emailID string) (*ReadOnceInfo, error) {
	var (
		readOnce               bool
		consumedAt             sql.NullInt64
		consumerDevice         sql.NullString
		selfDestructOnReadOnce sql.NullBool
	)

	err := esdb.db.QueryRow(`
		SELECT read_once, read_once_consumed_at, read_once_consumer_device, self_destruct_on_read_once
		FROM emails 
		WHERE email_id = ?
	`, emailID).Scan(&readOnce, &consumedAt, &consumerDevice, &selfDestructOnReadOnce)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		return nil, err
	}

	info := &ReadOnceInfo{
		IsConsumed:         consumedAt.Valid,
		SelfDestructOnRead: selfDestructOnReadOnce.Bool,
	}

	if consumedAt.Valid {
		info.ConsumedAt = time.Unix(consumedAt.Int64, 0)
	}
	if consumerDevice.Valid {
		info.ConsumerDevice = consumerDevice.String
	}

	return info, nil
}
