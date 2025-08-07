/**
 * Authentication utilities for the Secure Email MVP frontend
 * Handles JWT token management, automatic refresh, and authenticated API calls
 */

const API_BASE_URL = '/api';
const TOKEN_REFRESH_THRESHOLD = 5 * 60 * 1000; // 5 minutes before expiry

/**
 * Token storage utilities
 */
const TokenStorage = {
  // Store tokens in sessionStorage (cleared when browser closes)
  setTokens(accessToken, refreshToken) {
    sessionStorage.setItem('access_token', accessToken);
    sessionStorage.setItem('refresh_token', refreshToken);
    sessionStorage.setItem('token_expiry', Date.now() + (15 * 60 * 1000)); // 15 minutes
  },

  getAccessToken() {
    return sessionStorage.getItem('access_token');
  },

  getRefreshToken() {
    return sessionStorage.getItem('refresh_token');
  },

  getTokenExpiry() {
    const expiry = sessionStorage.getItem('token_expiry');
    return expiry ? parseInt(expiry) : 0;
  },

  clearTokens() {
    sessionStorage.removeItem('access_token');
    sessionStorage.removeItem('refresh_token');
    sessionStorage.removeItem('token_expiry');
  },

  isTokenExpired() {
    const expiry = this.getTokenExpiry();
    return Date.now() >= expiry - TOKEN_REFRESH_THRESHOLD;
  }
};

/**
 * Authentication API functions
 */
const AuthAPI = {
  /**
   * Login user with email, password, and TOTP code
   * @param {string} email - User's email
   * @param {string} password - User's password
   * @param {string} totpCode - TOTP code from authenticator app
   * @returns {Promise<Object>} Login response with tokens
   */
  async login(email, password, totpCode) {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password, totp_code: totpCode }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Login failed');
      }

      const data = await response.json();
      TokenStorage.setTokens(data.access_token, data.refresh_token);
      return data;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  },

  /**
   * Refresh access token using refresh token
   * @returns {Promise<Object>} New access token
   */
  async refreshToken() {
    try {
      const refreshToken = TokenStorage.getRefreshToken();
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      TokenStorage.setTokens(data.access_token, refreshToken); // Keep same refresh token
      return data;
    } catch (error) {
      console.error('Token refresh error:', error);
      TokenStorage.clearTokens();
      throw error;
    }
  },

  /**
   * Logout user and revoke refresh token
   * @returns {Promise<void>}
   */
  async logout() {
    try {
      const refreshToken = TokenStorage.getRefreshToken();
      if (refreshToken) {
        await fetch(`${API_BASE_URL}/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      TokenStorage.clearTokens();
    }
  },

  /**
   * Get current user information
   * @returns {Promise<Object>} User information
   */
  async getCurrentUser() {
    try {
      const response = await this.authFetch(`${API_BASE_URL}/auth/me`);
      return await response.json();
    } catch (error) {
      console.error('Get current user error:', error);
      throw error;
    }
  }
};

/**
 * Authenticated fetch wrapper
 * Automatically handles token refresh and adds Authorization header
 * @param {string} url - API endpoint URL
 * @param {Object} options - Fetch options
 * @returns {Promise<Response>} Fetch response
 */
async function authFetch(url, options = {}) {
  // Check if token needs refresh
  if (TokenStorage.isTokenExpired()) {
    try {
      await AuthAPI.refreshToken();
    } catch (error) {
      // If refresh fails, redirect to login
      window.location.href = '/login';
      throw new Error('Authentication required');
    }
  }

  // Get current access token
  const accessToken = TokenStorage.getAccessToken();
  if (!accessToken) {
    window.location.href = '/login';
    throw new Error('No access token available');
  }

  // Add Authorization header
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${accessToken}`,
    ...options.headers,
  };

  // Make the request
  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Handle 401 responses (token expired)
  if (response.status === 401) {
    try {
      await AuthAPI.refreshToken();
      // Retry the request with new token
      const newAccessToken = TokenStorage.getAccessToken();
      headers.Authorization = `Bearer ${newAccessToken}`;
      return await fetch(url, {
        ...options,
        headers,
      });
    } catch (error) {
      // If refresh fails, redirect to login
      TokenStorage.clearTokens();
      window.location.href = '/login';
      throw new Error('Authentication required');
    }
  }

  return response;
}

/**
 * Utility functions for components
 */
const AuthUtils = {
  /**
   * Check if user is authenticated
   * @returns {boolean} True if user has valid tokens
   */
  isAuthenticated() {
    const accessToken = TokenStorage.getAccessToken();
    return accessToken && !TokenStorage.isTokenExpired();
  },

  /**
   * Get user information from tokens (without API call)
   * @returns {Object|null} User info or null if not available
   */
  getUserFromToken() {
    const accessToken = TokenStorage.getAccessToken();
    if (!accessToken) return null;

    try {
      // Decode JWT payload (base64 decode the second part)
      const payload = accessToken.split('.')[1];
      const decoded = JSON.parse(atob(payload));
      return {
        user_id: decoded.user_id,
        email: decoded.email,
      };
    } catch (error) {
      console.error('Error decoding token:', error);
      return null;
    }
  },

  /**
   * Redirect to login if not authenticated
   */
  requireAuth() {
    if (!this.isAuthenticated()) {
      window.location.href = '/login';
    }
  }
};

// Export for use in components
export { AuthAPI, authFetch, TokenStorage, AuthUtils }; 