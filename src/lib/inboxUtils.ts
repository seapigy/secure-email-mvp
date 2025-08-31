/**
 * Inbox Utilities
 * 
 * Utility functions for transforming backend API responses to frontend data structures
 * and handling inbox-specific data operations.
 */

import { InboxEmailItem, ListInboxResponse } from './api';
import { SecureEmail, EmailStats, StatusType } from '@/types/secureEmail';

/**
 * Transform backend inbox email item to frontend SecureEmail format
 * @param inboxItem - Backend inbox email item
 * @returns Frontend SecureEmail object
 */
export const transformInboxEmailToSecureEmail = (inboxItem: InboxEmailItem): SecureEmail => {
  // Map backend status to frontend status
  const mapStatus = (backendStatus: string): StatusType => {
    switch (backendStatus) {
      case 'read':
        return 'opened';
      case 'deleted':
        return 'revoked';
      case 'expired':
        return 'expired';
      case 'delivered':
      default:
        return 'pending';
    }
  };

  return {
    id: inboxItem.email_id,
    subject: inboxItem.subject,
    from: `user-${inboxItem.sender_id}@securesystem.email`, // Transform sender_id to email format
    to: 'current-user@securesystem.email', // This will be the current user's email
    date: inboxItem.created_at,
    expires: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // Default 30 days
    status: mapStatus(inboxItem.status),
    accessAttempts: inboxItem.is_read ? 1 : 0,
    maxAttempts: 3,
    passwordProtected: false, // Default values - could be enhanced with more backend data
    geolocationRestricted: false,
    allowedCountries: [],
    readOnce: false,
    autoDestruct: false,
    destructAfterAttempts: null,
    decoyEnabled: false,
    fingerprint: `fingerprint-${inboxItem.email_id}`,
    metadataStripped: true,
    tamperAlerts: true,
    selfDestructAfterAttempts: false,
    maxFailedAttempts: 3,
    content: 'encrypted-content-placeholder',
    preview: inboxItem.subject,
    attachments: []
  };
};

/**
 * Transform backend inbox response to frontend email list and stats
 * @param response - Backend inbox response
 * @returns Object with emails array and stats
 */
export const transformInboxResponse = (response: ListInboxResponse): {
  emails: SecureEmail[];
  stats: EmailStats;
} => {
  const emails = response.emails.map(transformInboxEmailToSecureEmail);
  
  // Calculate stats from the emails
  const stats: EmailStats = {
    total: emails.length,
    pending: emails.filter(e => e.status === 'pending').length,
    opened: emails.filter(e => e.status === 'opened').length,
    expired: emails.filter(e => e.status === 'expired').length,
    passwordProtected: emails.filter(e => e.passwordProtected).length,
    geolocationRestricted: emails.filter(e => e.geolocationRestricted).length,
    readOnce: emails.filter(e => e.readOnce).length,
    autoDestruct: emails.filter(e => e.autoDestruct).length
  };

  return { emails, stats };
};

/**
 * Handle inbox API errors gracefully
 * @param error - API error
 * @returns User-friendly error message
 */
export const handleInboxError = (error: unknown): string => {
  if (typeof error === 'string') {
    return error;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'Failed to load inbox. Please try again.';
};
