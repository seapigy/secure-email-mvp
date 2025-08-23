# =============================================================================
# Phase 3 Development Environment Setup
# =============================================================================
# Secure Link External Email Flow - Viewing & Reply Flow
# =============================================================================

Write-Host "🚀 Phase 3 Development Environment Setup" -ForegroundColor Green
Write-Host "Secure Link External Email Flow - Viewing & Reply Flow" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Gray

# =============================================================================
# PHASE 3 SPECIFIC SETUP
# =============================================================================
Write-Host "`n🔒 Phase 3 Viewing & Reply Flow Setup..." -ForegroundColor Yellow

# Create Phase 3 specific directories
$phase3Dirs = @(
    "pkg/securelinks/viewer",
    "pkg/securelinks/reply", 
    "pkg/securelinks/chains",
    "pkg/securelinks/attachments",
    "src/components/external",
    "src/components/secure-viewer",
    "src/pages/external",
    "tests/phase3",
    "docs/phase3"
)

foreach ($dir in $phase3Dirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "✓ Created directory: $dir" -ForegroundColor Green
    }
}

# Create Phase 3 implementation files
Write-Host "Creating Phase 3 implementation files..." -ForegroundColor Cyan

# Secure viewer service
$viewerContent = @"
package viewer

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

// =============================================================================
// SECURE VIEWER SERVICE
// =============================================================================

// ViewerService handles secure viewing of emails for external users
type ViewerService struct {
    db *sql.DB
}

// ViewSession represents a secure viewing session
type ViewSession struct {
    ID             string    `json:"id" db:"id"`
    LinkID         string    `json:"link_id" db:"link_id"`
    IPAddress      string    `json:"ip_address" db:"ip_address"`
    UserAgent      string    `json:"user_agent" db:"user_agent"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
    IsActive       bool      `json:"is_active" db:"is_active"`
    EmailViewed    bool      `json:"email_viewed" db:"email_viewed"`
    ViewedAt       *time.Time `json:"viewed_at,omitempty" db:"viewed_at"`
    SessionToken   string    `json:"session_token" db:"session_token"`
}

// EmailView represents the sanitized email content for external viewing
type EmailView struct {
    Subject       string      `json:"subject"`
    Body          string      `json:"body"`
    SenderName    string      `json:"sender_name"`
    SenderEmail   string      `json:"sender_email"`
    SentAt        time.Time   `json:"sent_at"`
    Attachments   []Attachment `json:"attachments"`
    SecurityInfo  SecurityInfo `json:"security_info"`
}

// Attachment represents a secure attachment
type Attachment struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Size        int64  `json:"size"`
    ContentType string `json:"content_type"`
    SecureURL   string `json:"secure_url"`
}

// SecurityInfo provides security context to external users
type SecurityInfo struct {
    IsSecure       bool   `json:"is_secure"`
    EncryptionType string `json:"encryption_type"`
    ExpiresAt      *time.Time `json:"expires_at,omitempty"`
    ReadOnce       bool   `json:"read_once"`
    AutoDestruct   bool   `json:"auto_destruct"`
}

// NewViewerService creates a new secure viewer service
func NewViewerService(db *sql.DB) *ViewerService {
    return &ViewerService{
        db: db,
    }
}

// CreateViewSession creates a new secure viewing session
func (v *ViewerService) CreateViewSession(ctx context.Context, linkID, ipAddress, userAgent string) (*ViewSession, error) {
    // TODO: Implement view session creation
    return nil, fmt.Errorf("not implemented")
}

// GetEmailView retrieves sanitized email content for external viewing
func (v *ViewerService) GetEmailView(ctx context.Context, sessionToken string) (*EmailView, error) {
    // TODO: Implement secure email viewing
    return nil, fmt.Errorf("not implemented")
}

// RecordEmailView records that an email has been viewed
func (v *ViewerService) RecordEmailView(ctx context.Context, sessionToken string) error {
    // TODO: Implement view recording
    return fmt.Errorf("not implemented")
}
"@

$viewerContent | Out-File -FilePath "pkg/securelinks/viewer/service.go" -Encoding UTF8

# Reply handling service
$replyContent = @"
package reply

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

// =============================================================================
// REPLY HANDLING SERVICE
// =============================================================================

// ReplyService handles secure replies from external users
type ReplyService struct {
    db *sql.DB
}

// ReplyRequest represents a reply from an external user
type ReplyRequest struct {
    LinkID      string `json:"link_id" validate:"required"`
    Subject     string `json:"subject" validate:"required"`
    Body        string `json:"body" validate:"required"`
    SenderEmail string `json:"sender_email" validate:"required"`
    IPAddress   string `json:"ip_address"`
    UserAgent   string `json:"user_agent"`
}

// ReplyResponse represents the response to a reply request
type ReplyResponse struct {
    Success   bool   `json:"success"`
    ReplyID   string `json:"reply_id,omitempty"`
    Message   string `json:"message,omitempty"`
    Error     string `json:"error,omitempty"`
    ChainID   string `json:"chain_id,omitempty"`
}

// SecureReply represents a secure reply in the database
type SecureReply struct {
    ID               string    `json:"id" db:"id"`
    LinkID           string    `json:"link_id" db:"link_id"`
    ChainID          string    `json:"chain_id" db:"chain_id"`
    Subject          string    `json:"subject" db:"subject"`
    Body             string    `json:"body" db:"body"`
    SenderEmail      string    `json:"sender_email" db:"sender_email"`
    RecipientEmail   string    `json:"recipient_email" db:"recipient_email"`
    IPAddress        string    `json:"ip_address" db:"ip_address"`
    UserAgent        string    `json:"user_agent" db:"user_agent"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    ProcessedAt      *time.Time `json:"processed_at,omitempty" db:"processed_at"`
    Status           string    `json:"status" db:"status"` // "pending", "processed", "failed"
    InternalEmailID  *string   `json:"internal_email_id,omitempty" db:"internal_email_id"`
}

// NewReplyService creates a new reply service
func NewReplyService(db *sql.DB) *ReplyService {
    return &ReplyService{
        db: db,
    }
}

// ProcessReply processes a secure reply from an external user
func (r *ReplyService) ProcessReply(ctx context.Context, req ReplyRequest) (*ReplyResponse, error) {
    // TODO: Implement reply processing
    return nil, fmt.Errorf("not implemented")
}

// ForwardReplyToInternal forwards the reply to the internal email system
func (r *ReplyService) ForwardReplyToInternal(ctx context.Context, reply *SecureReply) error {
    // TODO: Implement reply forwarding to internal system
    return fmt.Errorf("not implemented")
}

// GetReplyHistory gets reply history for a secure link chain
func (r *ReplyService) GetReplyHistory(ctx context.Context, chainID string) ([]*SecureReply, error) {
    // TODO: Implement reply history retrieval
    return nil, fmt.Errorf("not implemented")
}
"@

$replyContent | Out-File -FilePath "pkg/securelinks/reply/service.go" -Encoding UTF8

# Email chains service
$chainsContent = @"
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
    ID              string    `json:"id" db:"id"`
    InitialLinkID   string    `json:"initial_link_id" db:"initial_link_id"`
    InternalUserID  string    `json:"internal_user_id" db:"internal_user_id"`
    ExternalEmail   string    `json:"external_email" db:"external_email"`
    Subject         string    `json:"subject" db:"subject"`
    Status          string    `json:"status" db:"status"` // "active", "closed", "expired"
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    LastActivity    time.Time `json:"last_activity" db:"last_activity"`
    MessageCount    int       `json:"message_count" db:"message_count"`
    ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// ChainMessage represents a message in an email chain
type ChainMessage struct {
    ID           string    `json:"id" db:"id"`
    ChainID      string    `json:"chain_id" db:"chain_id"`
    MessageType  string    `json:"message_type" db:"message_type"` // "initial", "reply", "forward"
    Subject      string    `json:"subject" db:"subject"`
    Body         string    `json:"body" db:"body"`
    SenderEmail  string    `json:"sender_email" db:"sender_email"`
    SenderType   string    `json:"sender_type" db:"sender_type"` // "internal", "external"
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    LinkID       *string   `json:"link_id,omitempty" db:"link_id"`
    EmailID      *string   `json:"email_id,omitempty" db:"email_id"`
}

// NewChainsService creates a new email chains service
func NewChainsService(db *sql.DB) *ChainsService {
    return &ChainsService{
        db: db,
    }
}

// CreateChain creates a new email chain
func (c *ChainsService) CreateChain(ctx context.Context, linkID, internalUserID, externalEmail, subject string) (*EmailChain, error) {
    // TODO: Implement chain creation
    return nil, fmt.Errorf("not implemented")
}

// AddMessageToChain adds a message to an existing chain
func (c *ChainsService) AddMessageToChain(ctx context.Context, chainID string, message *ChainMessage) error {
    // TODO: Implement message addition to chain
    return fmt.Errorf("not implemented")
}

// GetChainMessages retrieves all messages in a chain
func (c *ChainsService) GetChainMessages(ctx context.Context, chainID string) ([]*ChainMessage, error) {
    // TODO: Implement chain message retrieval
    return nil, fmt.Errorf("not implemented")
}

// CreateReplyLink creates a new secure link for reply continuation
func (c *ChainsService) CreateReplyLink(ctx context.Context, chainID, internalUserID, externalEmail string) (string, error) {
    // TODO: Implement reply link creation for chain continuation
    return "", fmt.Errorf("not implemented")
}
"@

$chainsContent | Out-File -FilePath "pkg/securelinks/chains/service.go" -Encoding UTF8

# Attachments service
$attachmentsContent = @"
package attachments

import (
    "context"
    "database/sql"
    "fmt"
    "io"
)

// =============================================================================
// SECURE ATTACHMENTS SERVICE
// =============================================================================

// AttachmentsService handles secure attachment management for external users
type AttachmentsService struct {
    db *sql.DB
}

// SecureAttachment represents a secure attachment
type SecureAttachment struct {
    ID            string `json:"id" db:"id"`
    LinkID        string `json:"link_id" db:"link_id"`
    OriginalName  string `json:"original_name" db:"original_name"`
    SecureName    string `json:"secure_name" db:"secure_name"`
    ContentType   string `json:"content_type" db:"content_type"`
    Size          int64  `json:"size" db:"size"`
    StoragePath   string `json:"storage_path" db:"storage_path"`
    EncryptionKey string `json:"encryption_key" db:"encryption_key"`
    Checksum      string `json:"checksum" db:"checksum"`
    CreatedAt     string `json:"created_at" db:"created_at"`
    ExpiresAt     string `json:"expires_at" db:"expires_at"`
    DownloadCount int    `json:"download_count" db:"download_count"`
    MaxDownloads  int    `json:"max_downloads" db:"max_downloads"`
}

// AttachmentDownloadRequest represents a request to download an attachment
type AttachmentDownloadRequest struct {
    AttachmentID string `json:"attachment_id" validate:"required"`
    SessionToken string `json:"session_token" validate:"required"`
    IPAddress    string `json:"ip_address"`
    UserAgent    string `json:"user_agent"`
}

// NewAttachmentsService creates a new secure attachments service
func NewAttachmentsService(db *sql.DB) *AttachmentsService {
    return &AttachmentsService{
        db: db,
    }
}

// GetAttachmentMetadata retrieves attachment metadata for external viewing
func (a *AttachmentsService) GetAttachmentMetadata(ctx context.Context, linkID string) ([]*SecureAttachment, error) {
    // TODO: Implement attachment metadata retrieval
    return nil, fmt.Errorf("not implemented")
}

// DownloadAttachment handles secure attachment downloads for external users
func (a *AttachmentsService) DownloadAttachment(ctx context.Context, req AttachmentDownloadRequest) (io.ReadCloser, string, error) {
    // TODO: Implement secure attachment download
    return nil, "", fmt.Errorf("not implemented")
}

// ValidateAttachmentAccess validates that an external user can access an attachment
func (a *AttachmentsService) ValidateAttachmentAccess(ctx context.Context, attachmentID, sessionToken string) error {
    // TODO: Implement attachment access validation
    return fmt.Errorf("not implemented")
}
"@

$attachmentsContent | Out-File -FilePath "pkg/securelinks/attachments/service.go" -Encoding UTF8

# Create Phase 3 API handlers template
$handlersContent = @"
package main

import (
    "encoding/json"
    "net/http"
    "log"
    
    "secure-email-mvp/pkg/securelinks/viewer"
    "secure-email-mvp/pkg/securelinks/reply"
    "secure-email-mvp/pkg/securelinks/chains"
    "secure-email-mvp/pkg/securelinks/attachments"
)

// =============================================================================
// PHASE 3 VIEWING & REPLY API HANDLERS
// =============================================================================

// TODO: Implement Phase 3 API handlers for:
// - Secure viewer endpoints
// - Reply handling endpoints  
// - Email chain management endpoints
// - Secure attachment endpoints

// SecureViewerHandler handles secure email viewing for external users
func (srv *Server) SecureViewerHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement secure viewer handler
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// SecureReplyHandler handles replies from external users
func (srv *Server) SecureReplyHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement secure reply handler
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// EmailChainHandler handles email chain management
func (srv *Server) EmailChainHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement email chain handler
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// SecureAttachmentHandler handles secure attachment downloads
func (srv *Server) SecureAttachmentHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement secure attachment handler
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
"@

$handlersContent | Out-File -FilePath "cmd/api/phase3_viewing_handlers.go" -Encoding UTF8

# Create frontend components
$externalViewerContent = @"
import React, { useState, useEffect } from 'react';
import { Shield, Mail, Download, Reply, AlertTriangle } from 'lucide-react';

// =============================================================================
// EXTERNAL SECURE EMAIL VIEWER
// =============================================================================

interface SecureEmailViewerProps {
  linkId: string;
  sessionToken?: string;
}

interface EmailData {
  subject: string;
  body: string;
  senderName: string;
  senderEmail: string;
  sentAt: string;
  attachments: Attachment[];
  securityInfo: SecurityInfo;
}

interface Attachment {
  id: string;
  name: string;
  size: number;
  contentType: string;
  secureUrl: string;
}

interface SecurityInfo {
  isSecure: boolean;
  encryptionType: string;
  expiresAt?: string;
  readOnce: boolean;
  autoDestruct: boolean;
}

export const SecureEmailViewer: React.FC<SecureEmailViewerProps> = ({ linkId, sessionToken }) => {
  const [emailData, setEmailData] = useState<EmailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showReply, setShowReply] = useState(false);
  const [replyText, setReplyText] = useState('');

  useEffect(() => {
    loadEmailData();
  }, [linkId, sessionToken]);

  const loadEmailData = async () => {
    try {
      setLoading(true);
      // TODO: Implement API call to load secure email data
      console.log('Loading email data for link:', linkId);
      
      // Mock data for now
      const mockData: EmailData = {
        subject: 'Secure Email Subject',
        body: 'This is a secure email body content.',
        senderName: 'John Doe',
        senderEmail: 'john@company.com',
        sentAt: new Date().toISOString(),
        attachments: [],
        securityInfo: {
          isSecure: true,
          encryptionType: 'AES-256-GCM',
          readOnce: false,
          autoDestruct: false
        }
      };
      
      setEmailData(mockData);
    } catch (err) {
      setError('Failed to load email data');
      console.error('Error loading email:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleReply = async () => {
    try {
      // TODO: Implement secure reply functionality
      console.log('Sending reply:', replyText);
      setShowReply(false);
      setReplyText('');
    } catch (err) {
      console.error('Error sending reply:', err);
    }
  };

  const downloadAttachment = async (attachment: Attachment) => {
    try {
      // TODO: Implement secure attachment download
      console.log('Downloading attachment:', attachment.name);
    } catch (err) {
      console.error('Error downloading attachment:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertTriangle className="mx-auto h-12 w-12 text-red-600" />
          <h2 className="mt-4 text-xl font-semibold text-gray-900">Error Loading Email</h2>
          <p className="mt-2 text-gray-600">{error}</p>
        </div>
      </div>
    );
  }

  if (!emailData) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Mail className="mx-auto h-12 w-12 text-gray-400" />
          <h2 className="mt-4 text-xl font-semibold text-gray-900">No Email Data</h2>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Security Banner */}
      <div className="bg-green-600 text-white px-4 py-3">
        <div className="flex items-center justify-center">
          <Shield className="h-5 w-5 mr-2" />
          <span className="text-sm font-medium">
            This email is secured with {emailData.securityInfo.encryptionType} encryption
          </span>
        </div>
      </div>

      {/* Email Content */}
      <div className="max-w-4xl mx-auto py-8 px-4">
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          {/* Email Header */}
          <div className="border-b border-gray-200 px-6 py-4">
            <h1 className="text-2xl font-bold text-gray-900">{emailData.subject}</h1>
            <div className="mt-2 flex items-center text-sm text-gray-600">
              <span>From: {emailData.senderName} &lt;{emailData.senderEmail}&gt;</span>
              <span className="mx-2">•</span>
              <span>{new Date(emailData.sentAt).toLocaleString()}</span>
            </div>
          </div>

          {/* Security Info */}
          {(emailData.securityInfo.readOnce || emailData.securityInfo.autoDestruct || emailData.securityInfo.expiresAt) && (
            <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
              <div className="flex">
                <AlertTriangle className="h-5 w-5 text-yellow-400" />
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-yellow-800">Security Notice</h3>
                  <div className="mt-2 text-sm text-yellow-700">
                    {emailData.securityInfo.readOnce && <p>• This email can only be viewed once</p>}
                    {emailData.securityInfo.autoDestruct && <p>• This email will self-destruct after viewing</p>}
                    {emailData.securityInfo.expiresAt && (
                      <p>• This email expires on {new Date(emailData.securityInfo.expiresAt).toLocaleString()}</p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Email Body */}
          <div className="px-6 py-6">
            <div className="prose max-w-none">
              <div dangerouslySetInnerHTML={{ __html: emailData.body.replace(/\n/g, '<br />') }} />
            </div>
          </div>

          {/* Attachments */}
          {emailData.attachments.length > 0 && (
            <div className="border-t border-gray-200 px-6 py-4">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Attachments</h3>
              <div className="space-y-2">
                {emailData.attachments.map((attachment) => (
                  <div
                    key={attachment.id}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center">
                      <Download className="h-5 w-5 text-gray-400 mr-3" />
                      <div>
                        <p className="text-sm font-medium text-gray-900">{attachment.name}</p>
                        <p className="text-xs text-gray-500">
                          {(attachment.size / 1024 / 1024).toFixed(2)} MB
                        </p>
                      </div>
                    </div>
                    <button
                      onClick={() => downloadAttachment(attachment)}
                      className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700"
                    >
                      Download
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Reply Section */}
          <div className="border-t border-gray-200 px-6 py-4">
            {!showReply ? (
              <button
                onClick={() => setShowReply(true)}
                className="flex items-center bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
              >
                <Reply className="h-4 w-4 mr-2" />
                Reply Securely
              </button>
            ) : (
              <div className="space-y-4">
                <h3 className="text-lg font-medium text-gray-900">Secure Reply</h3>
                <textarea
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  placeholder="Type your secure reply here..."
                  className="w-full h-32 p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
                <div className="flex space-x-3">
                  <button
                    onClick={handleReply}
                    className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                  >
                    Send Reply
                  </button>
                  <button
                    onClick={() => setShowReply(false)}
                    className="bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecureEmailViewer;
"@

$externalViewerContent | Out-File -FilePath "src/components/external/SecureEmailViewer.tsx" -Encoding UTF8

# Create database schema for Phase 3
$schemaContent = @"
-- =============================================================================
-- PHASE 3 DATABASE SCHEMA - VIEWING & REPLY FLOW
-- =============================================================================

-- View sessions for external users
CREATE TABLE IF NOT EXISTS link_view_sessions (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    email_viewed BOOLEAN NOT NULL DEFAULT 0,
    viewed_at DATETIME,
    session_token TEXT NOT NULL UNIQUE,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Secure replies from external users
CREATE TABLE IF NOT EXISTS secure_replies (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    chain_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, processed, failed
    internal_email_id TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Email chains for conversation management
CREATE TABLE IF NOT EXISTS email_chains (
    id TEXT PRIMARY KEY,
    initial_link_id TEXT NOT NULL,
    internal_user_id TEXT NOT NULL,
    external_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active', -- active, closed, expired
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message_count INTEGER NOT NULL DEFAULT 1,
    expires_at DATETIME,
    FOREIGN KEY (initial_link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (internal_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Chain messages for conversation history
CREATE TABLE IF NOT EXISTS chain_messages (
    id TEXT PRIMARY KEY,
    chain_id TEXT NOT NULL,
    message_type TEXT NOT NULL, -- initial, reply, forward
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    sender_type TEXT NOT NULL, -- internal or external
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    link_id TEXT,
    email_id TEXT,
    FOREIGN KEY (chain_id) REFERENCES email_chains(id) ON DELETE CASCADE,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE SET NULL,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE SET NULL
);

-- Secure attachments for external access
CREATE TABLE IF NOT EXISTS secure_attachments (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    original_name TEXT NOT NULL,
    secure_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    storage_path TEXT NOT NULL,
    encryption_key TEXT NOT NULL,
    checksum TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    download_count INTEGER NOT NULL DEFAULT 0,
    max_downloads INTEGER NOT NULL DEFAULT -1, -- -1 for unlimited
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Attachment downloads tracking
CREATE TABLE IF NOT EXISTS attachment_downloads (
    id TEXT PRIMARY KEY,
    attachment_id TEXT NOT NULL,
    session_token TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    downloaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_view_sessions_link_id ON link_view_sessions(link_id);
CREATE INDEX IF NOT EXISTS idx_view_sessions_token ON link_view_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_view_sessions_expires ON link_view_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_secure_replies_link_id ON secure_replies(link_id);
CREATE INDEX IF NOT EXISTS idx_secure_replies_chain_id ON secure_replies(chain_id);
CREATE INDEX IF NOT EXISTS idx_secure_replies_status ON secure_replies(status);

CREATE INDEX IF NOT EXISTS idx_email_chains_initial_link ON email_chains(initial_link_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_internal_user ON email_chains(internal_user_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_external_email ON email_chains(external_email);
CREATE INDEX IF NOT EXISTS idx_email_chains_status ON email_chains(status);

CREATE INDEX IF NOT EXISTS idx_chain_messages_chain_id ON chain_messages(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_messages_type ON chain_messages(message_type);

CREATE INDEX IF NOT EXISTS idx_secure_attachments_link_id ON secure_attachments(link_id);
CREATE INDEX IF NOT EXISTS idx_attachment_downloads_attachment_id ON attachment_downloads(attachment_id);
"@

$schemaContent | Out-File -FilePath "schema/migrate_add_phase3_viewing.sql" -Encoding UTF8

# Create Phase 3 test script template
$testContent = @"
# =============================================================================
# PHASE 3 VIEWING & REPLY FLOW TEST SCRIPT
# =============================================================================
# Tests all Phase 3 viewing and reply functionality
# =============================================================================

Write-Host "📧 Testing Phase 3 Viewing & Reply Flow" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan

# Configuration
`$API_BASE_URL = "http://localhost:8080"
`$TEST_EMAIL = "test@example.com"
`$TEST_PASSWORD = "test123"

Write-Host "`n📋 Phase 3 Test Scenarios:" -ForegroundColor Magenta
Write-Host "1. Secure email viewing for external users" -ForegroundColor Yellow
Write-Host "2. Reply handling from external users" -ForegroundColor Yellow
Write-Host "3. Email chain management" -ForegroundColor Yellow
Write-Host "4. Secure attachment downloads" -ForegroundColor Yellow
Write-Host "5. View session management" -ForegroundColor Yellow

# TODO: Implement comprehensive Phase 3 testing
Write-Host "`n⚠️ Phase 3 tests will be implemented during development" -ForegroundColor Yellow
"@

$testContent | Out-File -FilePath "tests/phase3/test_phase3_viewing_reply.ps1" -Encoding UTF8

Write-Host "`n✅ Phase 3 Development Environment Setup Complete!" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green

Write-Host "`n📁 Created directories:" -ForegroundColor Cyan
foreach ($dir in $phase3Dirs) {
    Write-Host "  • $dir" -ForegroundColor Gray
}

Write-Host "`n📄 Created implementation files:" -ForegroundColor Cyan
Write-Host "  • pkg/securelinks/viewer/service.go - Secure viewer service" -ForegroundColor Gray
Write-Host "  • pkg/securelinks/reply/service.go - Reply handling service" -ForegroundColor Gray
Write-Host "  • pkg/securelinks/chains/service.go - Email chains service" -ForegroundColor Gray
Write-Host "  • pkg/securelinks/attachments/service.go - Secure attachments service" -ForegroundColor Gray
Write-Host "  • cmd/api/phase3_viewing_handlers.go - API handlers template" -ForegroundColor Gray
Write-Host "  • src/components/external/SecureEmailViewer.tsx - React component" -ForegroundColor Gray
Write-Host "  • schema/migrate_add_phase3_viewing.sql - Database schema" -ForegroundColor Gray
Write-Host "  • tests/phase3/test_phase3_viewing_reply.ps1 - Test script template" -ForegroundColor Gray

Write-Host "`n🚀 Next Steps:" -ForegroundColor Magenta
Write-Host "1. Implement secure viewer service" -ForegroundColor Yellow
Write-Host "2. Create reply handling system" -ForegroundColor Yellow
Write-Host "3. Build email chain management" -ForegroundColor Yellow
Write-Host "4. Add secure attachment support" -ForegroundColor Yellow
Write-Host "5. Integrate frontend components" -ForegroundColor Yellow
Write-Host "6. Create comprehensive test suite" -ForegroundColor Yellow

Write-Host "`n📊 Phase 3 Development Environment Ready!" -ForegroundColor Green
