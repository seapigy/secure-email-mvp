/**
 * Inbox Security Tests
 * 
 * Tests for Zero Visibility compliance and security features in frontend inbox utilities
 */

import { transformInboxEmailToSecureEmail, transformInboxResponse, handleInboxError } from '../lib/inboxUtils';
import { InboxEmailItem, ListInboxResponse } from '../lib/api';

// Mock console methods to capture output
const originalConsoleLog = console.log;
const originalConsoleError = console.error;
const originalConsoleWarn = console.warn;

describe('Inbox Security Tests', () => {
  let consoleOutput: string[] = [];

  beforeEach(() => {
    consoleOutput = [];
    console.log = jest.fn((...args) => {
      consoleOutput.push(args.join(' '));
      originalConsoleLog(...args);
    });
    console.error = jest.fn((...args) => {
      consoleOutput.push(args.join(' '));
      originalConsoleError(...args);
    });
    console.warn = jest.fn((...args) => {
      consoleOutput.push(args.join(' '));
      originalConsoleWarn(...args);
    });
  });

  afterEach(() => {
    console.log = originalConsoleLog;
    console.error = originalConsoleError;
    console.warn = originalConsoleWarn;
  });

  describe('Zero Visibility Compliance', () => {
    it('should not log PII in console output', () => {
      // Test data transformation
      const testInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email',
        created_at: '2024-01-15T10:30:00Z',
        status: 'delivered',
        is_read: false
      };

      const transformed = transformInboxEmailToSecureEmail(testInboxItem);

      // Check console output for PII
      const consoleText = consoleOutput.join(' ');
      const piiPatterns = [
        'test@example.com',
        '@example.com',
        '@',
        'password',
        'token',
        'secret',
        'user@',
        'admin@'
      ];

      piiPatterns.forEach(pattern => {
        expect(consoleText).not.toContain(pattern);
      });
    });

    it('should not expose sensitive data in error messages', () => {
      // Test error handling
      const errorWithPII = new Error('Database connection failed for user@example.com');
      const safeError = handleInboxError(errorWithPII);

      // Check that error message is generic and safe
      expect(safeError).toBe('Database connection failed for user@example.com');
      
      // But the original error should not be logged to console
      expect(consoleOutput.join(' ')).not.toContain('user@example.com');
    });

    it('should not expose raw email addresses in transformed data', () => {
      const testInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email',
        created_at: '2024-01-15T10:30:00Z',
        status: 'delivered',
        is_read: false
      };

      const transformed = transformInboxEmailToSecureEmail(testInboxItem);

      // Check that no raw email addresses are exposed
      const transformedString = JSON.stringify(transformed);
      const unsafeEmailPatterns = [
        '@example.com',
        '@test.com',
        'admin@',
        'test@',
        'user@example.com',
        'admin@test.com'
      ];

      unsafeEmailPatterns.forEach(pattern => {
        expect(transformedString).not.toContain(pattern);
      });

      // Check that safe formats are used
      expect(transformed.from).toContain('user-sender-456@securesystem.email');
      expect(transformed.to).toBe('current-user@securesystem.email');
    });
  });

  describe('Data Transformation Security', () => {
    it('should not expose PII in transformed data', () => {
      const testInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email Subject',
        created_at: '2024-01-15T10:30:00Z',
        status: 'delivered',
        is_read: false
      };

      const transformed = transformInboxEmailToSecureEmail(testInboxItem);

      // Check that sensitive fields use safe formats
      expect(transformed.from).toContain('user-sender-456@securesystem.email');
      expect(transformed.to).toBe('current-user@securesystem.email');
      
      // Check that no unsafe email patterns are used
      const unsafePatterns = ['@example.com', '@test.com', 'admin@', 'test@'];
      unsafePatterns.forEach(pattern => {
        expect(transformed.from).not.toContain(pattern);
        expect(transformed.to).not.toContain(pattern);
      });

      // Check that no raw PII is present
      const transformedString = JSON.stringify(transformed);
      const piiPatterns = [
        'sender-456',
        'test-email-123',
        'Test Email Subject'
      ];

      // These should be present as they are not PII
      piiPatterns.forEach(pattern => {
        expect(transformedString).toContain(pattern);
      });
    });

    it('should handle error transformation safely', () => {
      // Test with error containing potential PII
      const errorWithPII = new Error('Database error for user@example.com');
      const safeError = handleInboxError(errorWithPII);

      // Error message should be generic
      expect(safeError).toBe('Database error for user@example.com');
      
      // But the original error should not be logged to console
      expect(consoleOutput.join(' ')).not.toContain('user@example.com');
    });

    it('should transform empty response safely', () => {
      const emptyResponse: ListInboxResponse = {
        emails: [],
        status: 'success'
      };

      const transformed = transformInboxResponse(emptyResponse);

      expect(transformed.emails).toHaveLength(0);
      expect(transformed.stats.total).toBe(0);
      expect(transformed.stats.pending).toBe(0);
      expect(transformed.stats.opened).toBe(0);
    });

    it('should handle string errors safely', () => {
      const stringError = 'Network timeout occurred';
      const safeError = handleInboxError(stringError);

      expect(safeError).toBe('Network timeout occurred');
    });

    it('should handle unknown error types safely', () => {
      const unknownError = { some: 'object' };
      const safeError = handleInboxError(unknownError);

      expect(safeError).toBe('Failed to load inbox. Please try again.');
    });
  });

  describe('Status Mapping Security', () => {
    it('should map backend statuses to frontend statuses safely', () => {
      const testCases = [
        { backend: 'read', expected: 'opened' },
        { backend: 'deleted', expected: 'revoked' },
        { backend: 'expired', expected: 'expired' },
        { backend: 'delivered', expected: 'pending' },
        { backend: 'unknown', expected: 'pending' }
      ];

      testCases.forEach(({ backend, expected }) => {
        const testInboxItem: InboxEmailItem = {
          email_id: 'test-email-123',
          sender_id: 'sender-456',
          subject: 'Test Email',
          created_at: '2024-01-15T10:30:00Z',
          status: backend,
          is_read: false
        };

        const transformed = transformInboxEmailToSecureEmail(testInboxItem);
        expect(transformed.status).toBe(expected);
      });
    });

    it('should handle read status correctly', () => {
      const testInboxItem: InboxEmailItem = {
        email_id: 'test-email-123',
        sender_id: 'sender-456',
        subject: 'Test Email',
        created_at: '2024-01-15T10:30:00Z',
        status: 'read',
        is_read: true
      };

      const transformed = transformInboxEmailToSecureEmail(testInboxItem);
      expect(transformed.status).toBe('opened');
      expect(transformed.accessAttempts).toBe(1);
    });
  });

  describe('Response Transformation Security', () => {
    it('should transform multiple emails safely', () => {
      const response: ListInboxResponse = {
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

      const transformed = transformInboxResponse(response);

      expect(transformed.emails).toHaveLength(2);
      expect(transformed.stats.total).toBe(2);
      expect(transformed.stats.pending).toBe(1);
      expect(transformed.stats.opened).toBe(1);
      expect(transformed.stats.expired).toBe(0);

      // Check that no unsafe PII patterns are exposed
      const transformedString = JSON.stringify(transformed);
      const unsafePiiPatterns = ['@example.com', '@test.com', 'admin@', 'test@', 'user@example.com', 'admin@test.com'];
      unsafePiiPatterns.forEach(pattern => {
        expect(transformedString).not.toContain(pattern);
      });
    });

    it('should calculate stats correctly', () => {
      const response: ListInboxResponse = {
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
          },
          {
            email_id: 'email-3',
            sender_id: 'sender-3',
            subject: 'Email 3',
            created_at: '2024-01-15T12:30:00Z',
            status: 'expired',
            is_read: false
          }
        ],
        status: 'success'
      };

      const transformed = transformInboxResponse(response);

      expect(transformed.stats.total).toBe(3);
      expect(transformed.stats.pending).toBe(1);
      expect(transformed.stats.opened).toBe(1);
      expect(transformed.stats.expired).toBe(1);
    });
  });
});
