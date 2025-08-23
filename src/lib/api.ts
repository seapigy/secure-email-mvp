// =============================================================================
// SECURE EMAIL MVP - API UTILITY
// =============================================================================
// Centralized API client for making HTTP requests to the backend.
// Handles authentication, error handling, and request/response interceptors.
// =============================================================================

import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { toast } from 'react-toastify';

// API Configuration
const API_HOST = import.meta.env.VITE_API_HOST || 'http://129.146.245.203:8080';

// Debug logging
console.log('API Configuration:', {
  VITE_API_HOST: import.meta.env.VITE_API_HOST,
  API_HOST,
  NODE_ENV: import.meta.env.MODE,
});

// Response types
export interface HealthCheckResponse {
  status: 'ok' | 'error';
  message: string;
  timestamp: string;
  version?: string;
}

export interface ApiError {
  message: string;
  status?: number;
  code?: string;
}

// Secure Email Request Interface
export interface SendSecureEmailRequest {
  recipient: string;
  subject: string;
  body: string;
  // Security settings
  selfDestructAfterAttempts?: boolean;
  maxFailedAttempts?: number;
  passwordProtection?: boolean;
  password?: string;
  // Enhanced geolocation verification (Micro-Iteration 4.15)
  geoVerificationType?: string; // "none", "country", "city", "city_country"
  geoCity?: string;
  geoCountry?: string;
  // Legacy geolocation fields (for backward compatibility)
  geolocationLock?: boolean;
  allowedCountry?: string;
  allowedCity?: string;
  timeLock?: boolean;
  unlockAfter?: string;
  autoDestruct?: boolean;
  destructAfterViews?: number;
  readOnce?: boolean;
  remoteRevoke?: boolean;
  decoyMessage?: boolean;
  stripMetadata?: boolean;
  tamperAlerts?: boolean;
  expiresAt?: string; // ISO 8601 UTC format
  requireMFA?: boolean;
  mfaType?: string;
}

// Secure Email Response Interface
export interface SendSecureEmailResponse {
  blob_id?: string;
  status: string;
  error?: string;
}

// Notification Interfaces
export interface NotificationPreferences {
  user_id: string;
  email_notifications: boolean;
  sms_notifications: boolean;
  notify_on_success: boolean;
  notify_on_failure: boolean;
  notify_on_blocked: boolean;
  include_geolocation: boolean;
  include_device_info: boolean;
  created_at: string;
  updated_at: string;
}

export interface AccessEvent {
  event_id: string;
  email_id: string;
  user_id: string;
  event_type: 'success' | 'failure' | 'blocked';
  ip_address: string;
  user_agent: string;
  country?: string;
  city?: string;
  device_type?: string;
  failure_reason?: string;
  timestamp: string;
}

export interface NotificationResponse {
  success: boolean;
  message?: string;
  data?: Record<string, unknown>;
}

// API Client Configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: API_HOST,
    timeout: 10000, // 10 second timeout
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor for authentication
  client.interceptors.request.use(
    (config) => {
      const token = sessionStorage.getItem('accessToken');
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
  client.interceptors.response.use(
    (response: AxiosResponse) => {
      return response;
    },
    (error) => {
      const message = error.response?.data?.message || error.message || 'An error occurred';
      const status = error.response?.status;

      // Handle different error types
      if (status === 401) {
        // Unauthorized - clear tokens and redirect to login
        sessionStorage.removeItem('accessToken');
        sessionStorage.removeItem('refreshToken');
        window.location.href = '/login';
      } else if (status === 403) {
        toast.error('Access denied. Please check your permissions.');
      } else if (status >= 500) {
        toast.error('Server error. Please try again later.');
      } else if (error.code === 'ECONNABORTED') {
        toast.error('Request timeout. Please check your connection.');
      } else if (error.code === 'ERR_NETWORK') {
        toast.error('Network error. Please check your connection.');
      }

      return Promise.reject({
        message,
        status,
        code: error.code,
      } as ApiError);
    }
  );

  return client;
};

// Create API client instance
const apiClient = createApiClient();

// Health Check API
export const healthCheck = async (): Promise<HealthCheckResponse> => {
  try {
    const response = await apiClient.get('/health');
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

// Authentication APIs
export const login = async (credentials: {
  email: string;
  password: string;
  totpCode: string;
}) => {
  try {
    const response = await apiClient.post('/api/auth/login', credentials);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const signup = async (data: {
  email: string;
  password: string;
  confirm_password: string;
}) => {
  try {
    const response = await apiClient.post('/api/auth/signup', data);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const verifyTotp = async (data: { totpCode: string }) => {
  try {
    const response = await apiClient.post('/api/auth/verify-totp', data);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const refreshToken = async (refreshToken: string) => {
  try {
    const response = await apiClient.post('/api/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const getUserProfile = async () => {
  try {
    const response = await apiClient.get('/api/auth/me');
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

// Secure Email APIs
export const sendSecureEmail = async (emailData: SendSecureEmailRequest): Promise<SendSecureEmailResponse> => {
  try {
    const response = await apiClient.post('/api/email/send', emailData);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

// Legacy Email APIs (for backward compatibility)
export const getEmails = async (params?: {
  folder?: string;
  page?: number;
  limit?: number;
}) => {
  try {
    const response = await apiClient.get('/api/emails', { params });
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const sendEmail = async (emailData: {
  to: string;
  subject: string;
  content: string;
  cc?: string;
  bcc?: string;
}) => {
  try {
    const response = await apiClient.post('/api/emails/send', emailData);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

export const getEmail = async (id: string) => {
  try {
    const response = await apiClient.get(`/api/emails/${id}`);
    return response.data;
  } catch (error) {
    throw error as ApiError;
  }
};

// Utility functions
export const isApiError = (error: unknown): error is ApiError => {
  return typeof error === 'object' && error !== null && 'message' in error && typeof (error as ApiError).message === 'string';
};

export const getErrorMessage = (error: unknown): string => {
  if (isApiError(error)) {
    return error.message;
  }
  return 'An unexpected error occurred';
};

// Notification API Functions
export const getNotificationPreferences = async (): Promise<NotificationPreferences> => {
  try {
    const response = await apiClient.get('/api/notifications/preferences');
    return response.data.data;
  } catch (error) {
    console.error('Error fetching notification preferences:', error);
    throw error;
  }
};

export const updateNotificationPreferences = async (
  preferences: Partial<NotificationPreferences>
): Promise<NotificationPreferences> => {
  try {
    const response = await apiClient.put('/api/notifications/preferences', preferences);
    return response.data.data;
  } catch (error) {
    console.error('Error updating notification preferences:', error);
    throw error;
  }
};

export const getAccessEventHistory = async (limit?: number): Promise<AccessEvent[]> => {
  try {
    const response = await apiClient.get('/api/notifications/history', {
      params: { limit },
    });
    return response.data.data;
  } catch (error) {
    console.error('Error fetching access event history:', error);
    throw error;
  }
};

// Export the API client for direct use if needed
export { apiClient }; 