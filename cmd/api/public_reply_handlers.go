package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/errors"
	"secure-email-mvp/pkg/email"

	"github.com/gorilla/mux"
)

// =============================================================================
// PUBLIC REPLY HANDLERS
// =============================================================================

// ReplyRequest represents an external reply request
type ReplyRequest struct {
	LinkID      string `json:"link_id"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	AccessToken string `json:"access_token,omitempty"` // Session token for validated access
}

// ReplyResponse represents the response from a reply request
type ReplyResponse struct {
	Success       bool   `json:"success"`
	ReplyID       string `json:"reply_id,omitempty"`
	Message       string `json:"message,omitempty"`
	Error         string `json:"error,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
}

// SecureReplyData represents the data stored for a secure reply
type SecureReplyData struct {
	ReplyID       string    `json:"reply_id"`
	LinkID        string    `json:"link_id"`
	EmailChainID  string    `json:"email_chain_id"`
	ParentEmailID string    `json:"parent_email_id"`
	Subject       string    `json:"subject"`
	Body          string    `json:"body"`
	SenderEmail   string    `json:"sender_email"`
	SenderName    string    `json:"sender_name,omitempty"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"status"`
}

// replyHandler handles POST /v/{linkID}/reply for external replies
func (srv *Server) replyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["linkID"]

	if linkID == "" {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Link ID is required", nil)
		return
	}

	var req ReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorCodeInvalidRequest, "Invalid request format", nil)
		return
	}

	req.LinkID = linkID
	req.IPAddress = srv.getClientIP(r)
	req.UserAgent = r.UserAgent()

	log.Printf("📧 External reply request for link: %s from IP: %s", linkID, req.IPAddress)

	// Process the reply
	response, err := srv.processExternalReply(r.Context(), req)
	if err != nil {
		log.Printf("❌ Failed to process external reply: %v", err)
		errors.WriteErrorResponse(w, http.StatusInternalServerError, errors.ErrorCodeInternalServer, "Failed to process reply", nil)
		return
	}

	// Log the reply attempt
	srv.logSecureLinkAccess(r.Context(), linkID, req.IPAddress, req.UserAgent, "reply_attempted", 
		fmt.Sprintf("Reply attempted: %t", response.Success), &response.TransactionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processExternalReply processes an external reply and forwards it to the original sender
func (srv *Server) processExternalReply(ctx context.Context, req ReplyRequest) (*ReplyResponse, error) {
	// Validate the secure link and get original email details
	originalEmail, err := srv.getOriginalEmailForReply(ctx, req.LinkID)
	if err != nil {
		return &ReplyResponse{
			Success:   false,
			Error:     "Secure link not found or expired",
			ErrorCode: "LINK_NOT_FOUND",
		}, nil
	}

	// Check if link is still valid for replies
	if err := srv.validateLinkForReply(ctx, req.LinkID); err != nil {
		return &ReplyResponse{
			Success:   false,
			Error:     err.Error(),
			ErrorCode: "LINK_INVALID",
		}, nil
	}

	// Generate reply ID
	replyID := fmt.Sprintf("reply_%s_%d", req.LinkID, time.Now().Unix())

	// Store the reply in the database
	replyData, err := srv.storeSecureReply(ctx, replyID, req, originalEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to store secure reply: %w", err)
	}

	// Forward the reply to the original sender
	transactionID, err := srv.forwardReplyToSender(ctx, replyData, originalEmail)
	if err != nil {
		log.Printf("⚠️ Failed to forward reply to sender: %v", err)
		// Don't fail the entire operation, just log the error
	}

	// Update reply status
	if err := srv.updateReplyStatus(ctx, replyID, "delivered"); err != nil {
		log.Printf("⚠️ Failed to update reply status: %v", err)
	}

	return &ReplyResponse{
		Success:       true,
		ReplyID:       replyID,
		Message:       "Reply sent successfully",
		TransactionID: transactionID,
	}, nil
}

// getOriginalEmailForReply retrieves the original email details for a reply
func (srv *Server) getOriginalEmailForReply(ctx context.Context, linkID string) (*SecureReplyData, error) {
	query := `
		SELECT 
			sl.email_id, sl.email_chain_id, sl.recipient_email, sl.recipient_name,
			e.subject, e.body, e.sender_id,
			u.email as sender_email, u.name as sender_name
		FROM secure_links sl
		JOIN emails e ON sl.email_id = e.email_id
		JOIN users u ON e.sender_id = u.id
		WHERE sl.link_id = ? AND sl.status = 'active'
	`

	var (
		emailID, emailChainID, recipientEmail, recipientName string
		subject, body, senderID, senderEmail, senderName     string
	)

	err := srv.db.QueryRowContext(ctx, query, linkID).Scan(
		&emailID, &emailChainID, &recipientEmail, &recipientName,
		&subject, &body, &senderID, &senderEmail, &senderName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("secure link not found or inactive")
		}
		return nil, fmt.Errorf("failed to query original email: %w", err)
	}

	return &SecureReplyData{
		LinkID:        linkID,
		EmailChainID:  emailChainID,
		ParentEmailID: emailID,
		SenderEmail:   recipientEmail,
		SenderName:    recipientName,
	}, nil
}

// validateLinkForReply validates that a link is still valid for replies
func (srv *Server) validateLinkForReply(ctx context.Context, linkID string) error {
	query := `
		SELECT status, expires_at, revoked_at, read_once, auto_destruct,
		       current_attempts, max_access_attempts
		FROM secure_links
		WHERE link_id = ?
	`

	var (
		status, revokedAt                                                                  string
		expiresAt                                                                          *int64
		readOnce, autoDestruct                                                             bool
		currentAttempts, maxAccessAttempts                                                 int
	)

	err := srv.db.QueryRowContext(ctx, query, linkID).Scan(
		&status, &expiresAt, &revokedAt, &readOnce, &autoDestruct,
		&currentAttempts, &maxAccessAttempts,
	)

	if err != nil {
		return fmt.Errorf("failed to validate link: %w", err)
	}

	// Check if link is active
	if status != "active" {
		return fmt.Errorf("secure link is not active")
	}

	// Check if link is revoked
	if revokedAt != "" {
		return fmt.Errorf("secure link has been revoked")
	}

	// Check if link is expired
	if expiresAt != nil && *expiresAt < time.Now().Unix() {
		return fmt.Errorf("secure link has expired")
	}

	// Check if link has been destroyed
	if currentAttempts >= maxAccessAttempts {
		return fmt.Errorf("secure link has been destroyed")
	}

	// Check if read-once link has been read
	if readOnce {
		var readCount int
		err := srv.db.QueryRowContext(ctx, "SELECT read_count FROM secure_links WHERE link_id = ?", linkID).Scan(&readCount)
		if err == nil && readCount > 0 {
			return fmt.Errorf("secure link has been read and cannot be replied to")
		}
	}

	return nil
}

// storeSecureReply stores a secure reply in the database
func (srv *Server) storeSecureReply(ctx context.Context, replyID string, req ReplyRequest, originalEmail *SecureReplyData) (*SecureReplyData, error) {
	query := `
		INSERT INTO secure_replies (
			reply_id, link_id, email_chain_id, parent_email_id,
			subject, body, sender_email, sender_name,
			ip_address, user_agent, created_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := srv.db.ExecContext(ctx, query,
		replyID,
		req.LinkID,
		originalEmail.EmailChainID,
		originalEmail.ParentEmailID,
		req.Subject,
		req.Body,
		req.IPAddress,
		req.UserAgent,
		now,
		"pending",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert secure reply: %w", err)
	}

	replyData := &SecureReplyData{
		ReplyID:       replyID,
		LinkID:        req.LinkID,
		EmailChainID:  originalEmail.EmailChainID,
		ParentEmailID: originalEmail.ParentEmailID,
		Subject:       req.Subject,
		Body:          req.Body,
		SenderEmail:   originalEmail.SenderEmail,
		SenderName:    originalEmail.SenderName,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		CreatedAt:     now,
		Status:        "pending",
	}

	return replyData, nil
}

// forwardReplyToSender forwards a reply to the original sender
func (srv *Server) forwardReplyToSender(ctx context.Context, replyData *SecureReplyData, originalEmail *SecureReplyData) (string, error) {
	// Get the original sender's email
	var senderEmail string
	err := srv.db.QueryRowContext(ctx, 
		"SELECT u.email FROM users u JOIN emails e ON u.id = e.sender_id WHERE e.email_id = ?", 
		originalEmail.ParentEmailID).Scan(&senderEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get original sender email: %w", err)
	}

	// Create email content for the reply
	emailSubject := fmt.Sprintf("Re: %s", replyData.Subject)
	emailBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background-color: #f8f9fa; padding: 20px; border-left: 4px solid #007bff; margin-bottom: 20px;">
				<h3 style="margin: 0; color: #007bff;">🔒 Secure Reply Received</h3>
				<p style="margin: 10px 0 0 0; color: #6c757d;">
					You have received a secure reply from <strong>%s</strong> (%s)
				</p>
			</div>
			
			<div style="background-color: #ffffff; padding: 20px; border: 1px solid #dee2e6; border-radius: 4px;">
				<h4 style="margin: 0 0 15px 0; color: #212529;">%s</h4>
				<div style="line-height: 1.6; color: #212529;">
					%s
				</div>
			</div>
			
			<div style="margin-top: 20px; padding: 15px; background-color: #f8f9fa; border-radius: 4px; font-size: 12px; color: #6c757d;">
				<p style="margin: 0;"><strong>Reply Details:</strong></p>
				<p style="margin: 5px 0;">• From: %s</p>
				<p style="margin: 5px 0;">• Date: %s</p>
				<p style="margin: 5px 0;">• IP Address: %s</p>
				<p style="margin: 5px 0;">• Secure Link ID: %s</p>
			</div>
		</div>
	`,
		replyData.SenderName, replyData.SenderEmail,
		replyData.Subject,
		replyData.Body,
		replyData.SenderEmail,
		replyData.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		replyData.IPAddress,
		replyData.LinkID,
	)

	// Send the email using SES
	transaction, err := srv.sesHandler.SendEmailViaSES(ctx, 
		replyData.ReplyID, 
		"noreply@securemail.com", // System sender
		senderEmail, 
		emailSubject, 
		emailBody,
	)

	if err != nil {
		return "", fmt.Errorf("failed to send reply email: %w", err)
	}

	// Log the SES transaction
	if err := srv.logSESTransaction(ctx, transaction, replyData.ReplyID, emailSubject); err != nil {
		log.Printf("⚠️ Failed to log SES transaction: %v", err)
	}

	return transaction.TransactionID, nil
}

// updateReplyStatus updates the status of a secure reply
func (srv *Server) updateReplyStatus(ctx context.Context, replyID, status string) error {
	query := `
		UPDATE secure_replies 
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE reply_id = ?
	`
	
	_, err := srv.db.ExecContext(ctx, query, status, replyID)
	return err
}

// logSESTransaction logs an SES transaction for a reply
func (srv *Server) logSESTransaction(ctx context.Context, transaction *email.SESTransaction, replyID string, subject string) error {
	query := `
		INSERT INTO ses_transactions (
			transaction_id, message_id, email_id, sender_id, recipient,
			subject, status, sent_at, reply_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := srv.db.ExecContext(ctx, query,
		transaction.TransactionID,
		transaction.MessageID,
		replyID, // Use reply ID as email ID
		"system", // System sender
		transaction.Recipient,
		subject,
		transaction.Status,
		transaction.Timestamp,
		replyID,
	)

	return err
}
