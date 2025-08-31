/**
 * ⚠️ CRITICAL WARNING - API PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE FRONTEND API CLIENT FUNCTIONS.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change API function signatures that the frontend components depend on
 * 2. NEVER modify request/response types that could break the UI
 * 3. NEVER alter error handling that affects the user experience
 * 4. NEVER change authentication flows that affect the design
 * 5. ONLY add new API functions that don't affect existing functionality
 * 6. ALWAYS maintain backward compatibility with existing components
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * The frontend design was restored from commit e291daf and represents the "perfect" design.
 * Any changes to the API that affect the frontend will result in immediate user dissatisfaction.
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE API, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: API PRESERVATION CRITICAL
 * @user_feedback: "This is the perfect design, never change it"
 */

import axios from 'axios';
import { log } from './logger';

// API base URL - adjust based on your environment
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// Create axios instance with default config
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized - redirect to login
      localStorage.removeItem('authToken');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Types for API requests and responses
export interface SendEmailRequest {
  recipient: string;
  subject: string;
  body: string;
  password?: string;
  requirePasswordForEveryEmail?: boolean;
  passwordPerEmail?: boolean;
  burnAfterRead?: boolean;
  selfDestructAfterAttempts?: boolean;
  maxFailedAttempts?: number;
  timeLock?: boolean;
  unlockAfter?: string;
  expiresAt?: string;
  geolocationLock?: boolean;
  geoVerificationType?: string;
  geoCity?: string;
  geoCountry?: string;
  allowedCountries?: string[];
  autoDestruct?: boolean;
  destructAfterViews?: number;
  readOnce?: boolean;
  requireMFA?: boolean;
  mfaType?: string;
  remoteRevoke?: boolean;
  decoyMessage?: boolean;
  decoySecret?: string;
  stripMetadata?: boolean;
  tamperAlerts?: boolean;
  generateFingerprintHash?: boolean;
  fingerprintHash?: string;
}

export interface SendEmailResponse {
  blob_id?: string;
  status: string;
  error?: string;
  burn_after_read?: boolean;
  access_count?: number;
  max_attempts?: number;
  secure_link_url?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
  totp_code: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    email: string;
  };
}

// API Functions

/**
 * Send a secure email with all security features
 */
export const sendSecureEmail = async (emailData: SendEmailRequest): Promise<SendEmailResponse> => {
  try {
    const response = await api.post<SendEmailResponse>('/api/email/send', emailData);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Login user with email, password, and TOTP
 */
export const loginUser = async (credentials: LoginRequest): Promise<LoginResponse> => {
  try {
    const response = await api.post<LoginResponse>('/api/auth/login', credentials);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get current user info
 */
export const getCurrentUser = async () => {
  try {
    const response = await api.get('/api/auth/me');
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Logout user
 */
export const logoutUser = async () => {
  try {
    await api.post('/api/auth/logout');
    localStorage.removeItem('authToken');
  } catch (error) {
          log.error('Logout error:', error, 'api');
    // Still remove token even if API call fails
    localStorage.removeItem('authToken');
  }
};

/**
 * Upload attachment
 */
export const uploadAttachment = async (file: File): Promise<{ url: string }> => {
  try {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await api.post('/api/attachments/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Validate email format
 */
export const validateEmail = async (email: string): Promise<{ valid: boolean }> => {
  try {
    const response = await api.post('/api/email/validate', { email });
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get secure link information
 */
export const getSecureLink = async (linkId: string): Promise<{ link_id: string; status: string }> => {
  try {
    const response = await api.get(`/api/secure-links/${linkId}`);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Check if error is an API error
 */
export const isApiError = (error: unknown): boolean => {
  const apiError = error as { response?: unknown; request?: unknown };
  return Boolean(error && (apiError.response || apiError.request));
};

/**
 * Extract error message from API error
 */
export const getErrorMessage = (error: unknown): string => {
  const apiError = error as { response?: { data?: { error?: string } }; message?: string };
  if (apiError.response?.data?.error) {
    return apiError.response.data.error;
  }
  if (apiError.message) {
    return apiError.message;
  }
  return 'An unexpected error occurred';
};

export interface HealthCheckResponse {
  status: string;
  timestamp: string;
  version?: string;
  uptime?: number;
}

// Inbox API Types
export interface InboxEmailItem {
  email_id: string;
  sender_id: string;
  subject: string;
  created_at: string;
  status: string;
  is_read: boolean;
}

export interface ListInboxResponse {
  emails: InboxEmailItem[];
  status: string;
  error?: string;
}

export interface GetInboxEmailResponse {
  email: InboxEmailItem;
  status: string;
  error?: string;
}

export interface DeleteInboxEmailResponse {
  status: string;
  error?: string;
}

/**
 * Send email (legacy function)
 */
export const sendEmail = async (emailData: SendEmailRequest): Promise<SendEmailResponse> => {
  return sendSecureEmail(emailData);
};

/**
 * Health check endpoint
 */
export const healthCheck = async (): Promise<HealthCheckResponse> => {
  try {
    const response = await api.get('/api/health');
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get notification preferences
 */
export const getNotificationPreferences = async () => {
  try {
    const response = await api.get('/api/notifications/preferences');
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Update notification preferences
 */
export const updateNotificationPreferences = async (preferences: unknown) => {
  try {
    const response = await api.put('/api/notifications/preferences', preferences);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get access event history
 */
export const getAccessEventHistory = async () => {
  try {
    const response = await api.get('/api/access-events');
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get inbox emails list
 */
export const getInboxEmails = async (): Promise<ListInboxResponse> => {
  try {
    const response = await api.get<ListInboxResponse>('/api/inbox/list');
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Get single inbox email
 */
export const getInboxEmail = async (emailId: string): Promise<GetInboxEmailResponse> => {
  try {
    const response = await api.get<GetInboxEmailResponse>(`/api/inbox/${emailId}`);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

/**
 * Delete inbox email (soft delete)
 */
export const deleteInboxEmail = async (emailId: string): Promise<DeleteInboxEmailResponse> => {
  try {
    const response = await api.delete<DeleteInboxEmailResponse>(`/api/inbox/${emailId}`);
    return response.data;
  } catch (error) {
    throw new Error(getErrorMessage(error));
  }
};

// Export default for backward compatibility
export default {
  sendSecureEmail,
  loginUser,
  getCurrentUser,
  logoutUser,
  uploadAttachment,
  validateEmail,
  getSecureLink,
  isApiError,
  getErrorMessage,
  sendEmail,
  healthCheck,
  getNotificationPreferences,
  updateNotificationPreferences,
  getAccessEventHistory,
  getInboxEmails,
  getInboxEmail,
  deleteInboxEmail,
}; 