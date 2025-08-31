/**
 * Inbox Utilities Tests
 * 
 * Tests for inbox utility functions that transform backend API responses
 * to frontend data structures.
 */

import { transformInboxEmailToSecureEmail, transformInboxResponse, handleInboxError } from '../lib/inboxUtils';
import { InboxEmailItem, ListInboxResponse } from '../lib/api';
import { StatusType } from '@/types/secureEmail';

describe('inboxUtils', () => {
  describe('transformInboxEmailToSecureEmail', () => {
    it('should transform backend inbox item to frontend SecureEmail format', () => {
      const mockInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email Subject',
        created_at: '2024-01-15T10:30:00Z',
        status: 'delivered',
        is_read: false
      };

      const result = transformInboxEmailToSecureEmail(mockInboxItem);

      expect(result.id).toBe('test-email-123');
      expect(result.subject).toBe('Test Email Subject');
      expect(result.from).toBe('user-sender-456@securesystem.email');
      expect(result.status).toBe('pending');
      expect(result.accessAttempts).toBe(0);
    });

    it('should map backend status "read" to frontend status "opened"', () => {
      const mockInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email Subject',
        created_at: '2024-01-15T10:30:00Z',
        status: 'read',
        is_read: true
      };

      const result = transformInboxEmailToSecureEmail(mockInboxItem);

      expect(result.status).toBe('opened');
      expect(result.accessAttempts).toBe(1);
    });

    it('should map backend status "deleted" to frontend status "revoked"', () => {
      const mockInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email Subject',
        created_at: '2024-01-15T10:30:00Z',
        status: 'deleted',
        is_read: false
      };

      const result = transformInboxEmailToSecureEmail(mockInboxItem);

      expect(result.status).toBe('revoked');
    });

    it('should map backend status "expired" to frontend status "expired"', () => {
      const mockInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email Subject',
        created_at: '2024-01-15T10:30:00Z',
        status: 'expired',
        is_read: false
      };

      const result = transformInboxEmailToSecureEmail(mockInboxItem);

      expect(result.status).toBe('expired');
    });
  });

  describe('transformInboxResponse', () => {
    it('should transform backend response to frontend format with stats', () => {
      const mockResponse: ListInboxResponse = {
        emails: [
          {
            email_id: 'email-1',
            sender_id: 'sender-1',
            subject: 'Email 1',
            created_at: '2024-01-15T10:30:00Z',
            status: 'delivered',
            is_read: false
          },
          {
            email_id: 'email-2',
            sender_id: 'sender-2',
            subject: 'Email 2',
            created_at: '2024-01-15T11:30:00Z',
            status: 'read',
            is_read: true
          }
        ],
        status: 'success'
      };

      const result = transformInboxResponse(mockResponse);

      expect(result.emails).toHaveLength(2);
      expect(result.stats.total).toBe(2);
      expect(result.stats.pending).toBe(1);
      expect(result.stats.opened).toBe(1);
      expect(result.stats.expired).toBe(0);
    });

    it('should handle empty response', () => {
      const mockResponse: ListInboxResponse = {
        emails: [],
        status: 'success'
      };

      const result = transformInboxResponse(mockResponse);

      expect(result.emails).toHaveLength(0);
      expect(result.stats.total).toBe(0);
      expect(result.stats.pending).toBe(0);
      expect(result.stats.opened).toBe(0);
    });
  });

  describe('handleInboxError', () => {
    it('should return error message for string errors', () => {
      const error = 'Network error occurred';
      const result = handleInboxError(error);
      expect(result).toBe('Network error occurred');
    });

    it('should return error message for Error objects', () => {
      const error = new Error('API request failed');
      const result = handleInboxError(error);
      expect(result).toBe('API request failed');
    });

    it('should return default message for unknown error types', () => {
      const error = { some: 'object' };
      const result = handleInboxError(error);
      expect(result).toBe('Failed to load inbox. Please try again.');
    });
  });
});
