// =============================================================================
// SECURE EMAIL MVP - TYPE DEFINITIONS
// =============================================================================

/**
 * Secure Email Interface
 * 
 * Represents a secure email with comprehensive security features.
 * This interface defines all properties and security settings
 * for individual secure emails in the system.
 */
export interface SecureEmail {
  /** Unique identifier for the email */
  id: string;
  
  /** Email subject line */
  subject: string;
  
  /** Sender email address */
  from: string;
  
  /** Recipient email address */
  to: string;
  
  /** Date when the email was sent */
  date: string;
  
  /** Expiration date for the email */
  expires: string;
  
  /** Optional expiration timestamp (ISO 8601 UTC format) */
  expiresAt?: string | null;
  
  /** Current status of the email */
  status: 'pending' | 'opened' | 'expired' | 'revoked';
  
  /** Number of access attempts made */
  accessAttempts: number;
  
  /** Maximum number of allowed access attempts */
  maxAttempts: number;
  
  /** Whether the email requires a password to access */
  passwordProtected: boolean;
  
  /** Whether the email has geolocation restrictions */
  geolocationRestricted: boolean;
  
  /** List of countries allowed to access the email */
  allowedCountries: string[];
  
  /** Whether the email can only be read once */
  readOnce: boolean;
  
  /** Whether the email will auto-destruct after viewing */
  autoDestruct: boolean;
  
  /** Number of views before auto-destruction */
  destructAfterAttempts: number | null;
  
  /** Whether decoy message is enabled */
  decoyEnabled: boolean;
  
  /** Cryptographic fingerprint for tamper detection */
  fingerprint: string;
  
  /** Whether metadata has been stripped */
  metadataStripped: boolean;
  
  /** Whether tamper alerts are enabled */
  tamperAlerts: boolean;
  
  /** Whether the email self-destructs after failed attempts */
  selfDestructAfterAttempts: boolean;
  
  /** Maximum number of failed attempts before self-destruction */
  maxFailedAttempts: number;
  
  /** Encrypted email content (base64 encoded) */
  content: string;
  
  /** Preview text for the email */
  preview: string;
  
  /** List of email attachments */
  attachments: EmailAttachment[];
}

/**
 * Email Attachment Interface
 * 
 * Represents an attachment to a secure email.
 */
export interface EmailAttachment {
  /** Name of the attachment file */
  name: string;
  
  /** Size of the attachment */
  size: string;
  
  /** Whether the attachment is encrypted */
  encrypted: boolean;
}

/**
 * Email Statistics Interface
 * 
 * Provides statistics about the user's email collection.
 */
export interface EmailStats {
  /** Total number of emails */
  total: number;
  
  /** Number of pending emails */
  pending: number;
  
  /** Number of opened emails */
  opened: number;
  
  /** Number of expired emails */
  expired: number;
  
  /** Number of password-protected emails */
  passwordProtected: number;
  
  /** Number of geolocation-restricted emails */
  geolocationRestricted: number;
  
  /** Number of read-once emails */
  readOnce: number;
  
  /** Number of auto-destruct emails */
  autoDestruct: number;
}

/**
 * Mock Data Interface
 * 
 * Structure for mock data used in development.
 */
export interface MockData {
  /** Array of mock emails */
  emails: SecureEmail[];
  
  /** Email statistics */
  stats: EmailStats;
}

/**
 * Security Settings Interface
 * 
 * Defines all available security settings for secure emails.
 * These settings can be applied globally or per-email.
 */
export interface SecuritySettings {
  /** Enable password protection for emails */
  passwordProtection: boolean;
  
  /** Password for protected emails */
  password?: string;
  
  /** Require password for every secure email (per-email mode) */
  perEmailPassword: boolean;
  
  /** Enable geolocation-based access restrictions */
  geolocationLock: boolean;
  
  /** List of countries allowed to access emails */
  allowedCountries: string[];
  
  /** Enable time-based access restrictions */
  timeLock: boolean;
  
  /** Date/time when emails become accessible */
  unlockAfter?: string;
  
  /** Enable auto-destruct after viewing */
  autoDestruct: boolean;
  
  /** Number of views before auto-destruction */
  destructAfterAttempts?: number;
  
  /** Enable read-once mode */
  readOnce: boolean;
  
  /** Enable remote revocation capability */
  remoteRevoke: boolean;
  
  /** Enable decoy message feature */
  decoyMessage: boolean;
  
  /** Cryptographic fingerprint for integrity */
  fingerprintHash: string;
  
  /** Strip metadata from emails */
  stripMetadata: boolean;
  
  /** Enable tamper detection alerts */
  tamperAlerts: boolean;
  
  /** Enable self-destruct after failed attempts */
  selfDestructAfterAttempts: boolean;
  
  /** Maximum failed attempts before self-destruction */
  maxFailedAttempts?: number;
}

/**
 * Status Type
 * 
 * Possible status values for secure emails.
 */
export type StatusType = 'pending' | 'opened' | 'expired' | 'revoked';

/**
 * Status Configuration Interface
 * 
 * Defines the visual appearance of status badges.
 */
export interface StatusConfig {
  /** Display label for the status */
  label: string;
  
  /** Text color for the status */
  color: string;
  
  /** Background color for the status */
  bgColor: string;
  
  /** Icon name for the status */
  icon: string;
}

/**
 * Email Filters Interface
 * 
 * Defines filter options for the email inbox.
 */
export interface EmailFilters {
  /** Filter by email status */
  status?: StatusType;
  
  /** Filter by password protection */
  passwordProtected?: boolean;
  
  /** Filter by geolocation restrictions */
  geolocationRestricted?: boolean;
  
  /** Filter by read-once mode */
  readOnce?: boolean;
  
  /** Filter by auto-destruct feature */
  autoDestruct?: boolean;
  
  /** Search term for filtering */
  search?: string;
}

/**
 * Sort Field Type
 * 
 * Available fields for sorting emails.
 */
export type SortField = 'date' | 'subject' | 'from' | 'expires' | 'status';

/**
 * Sort Direction Type
 * 
 * Available sort directions.
 */
export type SortDirection = 'asc' | 'desc';

/**
 * Sort Configuration Interface
 * 
 * Defines how emails should be sorted.
 */
export interface SortConfig {
  /** Field to sort by */
  field: SortField;
  
  /** Sort direction */
  direction: SortDirection;
} 