// API Configuration for SecureMail Frontend
export const API_CONFIG = {
  // Backend API base URL - will be different for development/production
  BASE_URL: process.env.EXPO_PUBLIC_API_URL || (__DEV__ 
    ? 'http://localhost:8080'  // Development backend
    : 'https://api.securesystem.email'),  // Production backend
  
  // API endpoints
  ENDPOINTS: {
    SIGNUP: '/api/auth/signup',
    LOGIN: '/api/auth/login',
    VERIFY_EMAIL: '/api/auth/verify-email',
    RESEND_VERIFICATION: '/api/auth/resend-verification',
    SETUP_MFA: '/api/auth/setup-mfa',
    VALIDATE_MFA: '/api/auth/validate-mfa',
    RECOVER_ACCOUNT: '/api/account/recover',
    INBOX_FOLDERS: '/api/inbox/folders',
    INBOX_MESSAGES: '/api/inbox/messages',
    TRIAL_WARNING: '/api/trial/warning',
    EXTEND_TRIAL: '/api/trial/extend',
    UPGRADE_ACCOUNT: '/api/auth/upgrade-account',
    DOWNGRADE_ACCOUNT: '/api/auth/downgrade-account',
    CREATE_ORGANIZATION: '/api/org/create',
    ADD_USER_TO_ORG: '/api/org/add-user',
    LIST_ORG_USERS: '/api/org/list-users',
  },
  
  // Request timeout
  TIMEOUT: 10000,
  
  // Headers
  DEFAULT_HEADERS: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
} as const;

// Account types supported by the backend
export const ACCOUNT_TYPES = {
  FREE: 'free',
  PREMIUM: 'premium',
  ENTERPRISE: 'enterprise',
} as const;

export type AccountType = typeof ACCOUNT_TYPES[keyof typeof ACCOUNT_TYPES];

// Trial warning levels
export const TRIAL_WARNING_LEVELS = {
  INFO: 'info',
  WARNING: 'warning',
  CRITICAL: 'critical',
} as const;

export type TrialWarningLevel = typeof TRIAL_WARNING_LEVELS[keyof typeof TRIAL_WARNING_LEVELS];
