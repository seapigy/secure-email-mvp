package auth

// DO NOT EDIT EXISTING CODE - new file added
// Inbox management handlers for creating default folders and welcome messages

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Inbox encryption key (32 bytes for AES-256)
var inboxEncryptionKey = []byte("your-32-byte-inbox-encryption-ke")

// CreateDefaultInbox creates default folders and welcome message for new user
func CreateDefaultInbox(userID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create default folders
	defaultFolders := []struct {
		name       string
		folderType string
		sortOrder  int
	}{
		{"Inbox", "inbox", 1},
		{"Sent", "sent", 2},
		{"Trash", "trash", 3},
		{"Drafts", "drafts", 4},
	}

	for _, folder := range defaultFolders {
		_, err = tx.Exec(`
			INSERT INTO mailbox_folders (id, user_id, name, folder_type, sort_order, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, uuid.New().String(), userID, folder.name, folder.folderType, folder.sortOrder, time.Now(), time.Now())
		if err != nil {
			return err
		}
	}

	// Get the inbox folder ID
	var inboxFolderID string
	err = tx.QueryRow("SELECT id FROM mailbox_folders WHERE user_id = ? AND folder_type = 'inbox'", userID).Scan(&inboxFolderID)
	if err != nil {
		return err
	}

	// Create welcome message
	welcomeMessage := map[string]interface{}{
		"subject": "Welcome to Secure Email!",
		"body": `Welcome to Secure Email!

Thank you for choosing our secure email service. Your account has been successfully created and your inbox is ready to use.

Key Features:
• End-to-end encryption for all your emails
• Zero-knowledge architecture - we can't read your messages
• Advanced security with quantum-resistant cryptography
• Custom domains for Premium and Enterprise users

Getting Started:
1. Verify your email address using the verification code sent to you
2. Set up two-factor authentication for enhanced security
3. Start sending and receiving encrypted emails

Need help? Check out our documentation or contact support.

Best regards,
The Secure Email Team`,
		"from": "noreply@securesystem.email",
		"to":   "your-email@example.com", // This will be replaced with actual user email
	}

	// Encrypt welcome message
	encryptedBody, err := encryptInboxData(welcomeMessage["body"].(string))
	if err != nil {
		return err
	}

	encryptedSubject, err := encryptInboxData(welcomeMessage["subject"].(string))
	if err != nil {
		return err
	}

	encryptedFrom, err := encryptInboxData(welcomeMessage["from"].(string))
	if err != nil {
		return err
	}

	// Insert welcome message
	_, err = tx.Exec(`
		INSERT INTO email_messages (
			id, user_id, folder_id, message_id, from_address, to_addresses,
			subject, body_encrypted, body_type, size_bytes, is_read,
			received_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 
		uuid.New().String(), userID, inboxFolderID, uuid.New().String(),
		encryptedFrom, `["encrypted-user-email"]`, // Placeholder, will be replaced
		encryptedSubject, encryptedBody, "text/plain", 
		len(welcomeMessage["body"].(string)), false,
		time.Now(), time.Now(), time.Now())

	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetInboxFolders returns all folders for a user
func GetInboxFolders(userID string) ([]map[string]interface{}, error) {
	rows, err := DB.Query(`
		SELECT id, name, folder_type, sort_order, created_at
		FROM mailbox_folders 
		WHERE user_id = ? 
		ORDER BY sort_order, name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []map[string]interface{}
	for rows.Next() {
		var id, name, folderType string
		var sortOrder int
		var createdAt time.Time

		err := rows.Scan(&id, &name, &folderType, &sortOrder, &createdAt)
		if err != nil {
			return nil, err
		}

		folders = append(folders, map[string]interface{}{
			"id":          id,
			"name":        name,
			"folder_type": folderType,
			"sort_order":  sortOrder,
			"created_at":  createdAt.Format(time.RFC3339),
		})
	}

	return folders, nil
}

// GetInboxMessages returns messages in a folder
func GetInboxMessages(userID, folderID string, limit, offset int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(`
		SELECT id, message_id, from_address, to_addresses, subject, 
		       body_encrypted, body_type, size_bytes, is_read, is_important, 
		       is_starred, received_at, created_at
		FROM email_messages 
		WHERE user_id = ? AND folder_id = ?
		ORDER BY received_at DESC
		LIMIT ? OFFSET ?
	`, userID, folderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, messageID, fromAddress, toAddresses, subject string
		var bodyEncrypted []byte
		var bodyType string
		var sizeBytes int
		var isRead, isImportant, isStarred bool
		var receivedAt, createdAt time.Time

		err := rows.Scan(&id, &messageID, &fromAddress, &toAddresses, &subject,
			&bodyEncrypted, &bodyType, &sizeBytes, &isRead, &isImportant,
			&isStarred, &receivedAt, &createdAt)
		if err != nil {
			return nil, err
		}

		// Decrypt message data
		decryptedSubject, _ := decryptInboxData(subject)
		decryptedFrom, _ := decryptInboxData(fromAddress)

		messages = append(messages, map[string]interface{}{
			"id":           id,
			"message_id":   messageID,
			"from":         decryptedFrom,
			"to":           toAddresses, // Keep encrypted for now
			"subject":      decryptedSubject,
			"body_type":    bodyType,
			"size_bytes":   sizeBytes,
			"is_read":      isRead,
			"is_important": isImportant,
			"is_starred":   isStarred,
			"received_at":  receivedAt.Format(time.RFC3339),
			"created_at":   createdAt.Format(time.RFC3339),
		})
	}

	return messages, nil
}

// InboxFoldersHandler returns user's inbox folders
func InboxFoldersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	folders, err := GetInboxFolders(userID)
	if err != nil {
		log.Printf("ERROR getting inbox folders: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"folders": folders,
	})
}

// InboxMessagesHandler returns messages in a folder
func InboxMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	folderID := r.URL.Query().Get("folder_id")
	if folderID == "" {
		http.Error(w, "folder_id required", http.StatusBadRequest)
		return
	}

	messages, err := GetInboxMessages(userID, folderID, 50, 0) // Default limit/offset
	if err != nil {
		log.Printf("ERROR getting inbox messages: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

// encryptInboxData encrypts sensitive inbox data using AES-256-GCM
func encryptInboxData(data string) (string, error) {
	block, err := aes.NewCipher(inboxEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return string(ciphertext), nil
}

// decryptInboxData decrypts sensitive inbox data using AES-256-GCM
func decryptInboxData(encryptedData string) (string, error) {
	block, err := aes.NewCipher(inboxEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data := []byte(encryptedData)
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
