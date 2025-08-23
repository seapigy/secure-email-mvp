package chains

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// =============================================================================
// EMAIL CHAINS SERVICE
// =============================================================================

// ChainsService manages email chains between internal and external users
type ChainsService struct {
	db *sql.DB
}

// EmailChain represents a conversation chain
type EmailChain struct {
	ID             string     `json:"id" db:"id"`
	InitialLinkID  string     `json:"initial_link_id" db:"initial_link_id"`
	InternalUserID string     `json:"internal_user_id" db:"internal_user_id"`
	ExternalEmail  string     `json:"external_email" db:"external_email"`
	Subject        string     `json:"subject" db:"subject"`
	Status         string     `json:"status" db:"status"` // "active", "closed", "expired"
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	LastActivity   time.Time  `json:"last_activity" db:"last_activity"`
	MessageCount   int        `json:"message_count" db:"message_count"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// ChainMessage represents a message in an email chain
type ChainMessage struct {
	ID          string    `json:"id" db:"id"`
	ChainID     string    `json:"chain_id" db:"chain_id"`
	MessageType string    `json:"message_type" db:"message_type"` // "initial", "reply", "forward"
	Subject     string    `json:"subject" db:"subject"`
	Body        string    `json:"body" db:"body"`
	SenderEmail string    `json:"sender_email" db:"sender_email"`
	SenderType  string    `json:"sender_type" db:"sender_type"` // "internal", "external"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	LinkID      *string   `json:"link_id,omitempty" db:"link_id"`
	EmailID     *string   `json:"email_id,omitempty" db:"email_id"`
}

// NewChainsService creates a new email chains service
func NewChainsService(db *sql.DB) *ChainsService {
	return &ChainsService{
		db: db,
	}
}

// CreateChain creates a new email chain
func (c *ChainsService) CreateChain(ctx context.Context, linkID, internalUserID, externalEmail, subject string) (*EmailChain, error) {
	chainID := c.generateChainID()

	query := `
		INSERT INTO email_chains (
			id, initial_link_id, internal_user_id, external_email, subject,
			status, created_at, last_activity, message_count
		) VALUES (?, ?, ?, ?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 1)
	`

	_, err := c.db.ExecContext(ctx, query, chainID, linkID, internalUserID, externalEmail, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain: %w", err)
	}

	return &EmailChain{
		ID:             chainID,
		InitialLinkID:  linkID,
		InternalUserID: internalUserID,
		ExternalEmail:  externalEmail,
		Subject:        subject,
		Status:         "active",
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
		MessageCount:   1,
	}, nil
}

// AddMessageToChain adds a message to an existing chain
func (c *ChainsService) AddMessageToChain(ctx context.Context, chainID string, message *ChainMessage) error {
	query := `
		INSERT INTO chain_messages (
			id, chain_id, message_type, subject, body, sender_email, sender_type,
			created_at, link_id, email_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?)
	`

	messageID := c.generateMessageID()
	_, err := c.db.ExecContext(ctx, query,
		messageID, chainID, message.MessageType, message.Subject, message.Body,
		message.SenderEmail, message.SenderType, message.LinkID, message.EmailID,
	)

	if err != nil {
		return fmt.Errorf("failed to add message to chain: %w", err)
	}

	// Update chain activity and message count
	updateQuery := `
		UPDATE email_chains 
		SET last_activity = CURRENT_TIMESTAMP, message_count = message_count + 1 
		WHERE id = ?
	`
	_, err = c.db.ExecContext(ctx, updateQuery, chainID)

	return err
}

// GetChainMessages retrieves all messages in a chain
func (c *ChainsService) GetChainMessages(ctx context.Context, chainID string) ([]*ChainMessage, error) {
	query := `
		SELECT id, chain_id, message_type, subject, body, sender_email, sender_type,
		       created_at, link_id, email_id
		FROM chain_messages
		WHERE chain_id = ?
		ORDER BY created_at ASC
	`

	rows, err := c.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chain messages: %w", err)
	}
	defer rows.Close()

	var messages []*ChainMessage
	for rows.Next() {
		var message ChainMessage
		err := rows.Scan(
			&message.ID, &message.ChainID, &message.MessageType, &message.Subject,
			&message.Body, &message.SenderEmail, &message.SenderType, &message.CreatedAt,
			&message.LinkID, &message.EmailID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chain message: %w", err)
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// CreateReplyLink creates a new secure link for reply continuation
func (c *ChainsService) CreateReplyLink(ctx context.Context, chainID, internalUserID, externalEmail string) (string, error) {
	// Generate new link ID
	linkID := c.generateLinkID()

	// Create a basic secure link for chain continuation
	query := `
		INSERT INTO secure_links (
			link_id, email_id, recipient_email, sender_id, security_settings,
			created_at, access_count, status, failed_attempts
		) VALUES (?, 'chain_continuation', ?, ?, '{}', CURRENT_TIMESTAMP, 0, 'active', 0)
	`

	_, err := c.db.ExecContext(ctx, query, linkID, externalEmail, internalUserID)
	if err != nil {
		return "", fmt.Errorf("failed to create reply link: %w", err)
	}

	return linkID, nil
}

// Helper methods

// generateChainID generates a unique chain ID
func (c *ChainsService) generateChainID() string {
	return fmt.Sprintf("chain_%d", time.Now().UnixNano())
}

// generateMessageID generates a unique message ID
func (c *ChainsService) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// generateLinkID generates a unique link ID
func (c *ChainsService) generateLinkID() string {
	return fmt.Sprintf("link_%d", time.Now().UnixNano())
}
