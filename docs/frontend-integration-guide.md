# Frontend Integration Guide - Admin Authentication System

## Overview

This guide provides step-by-step instructions for integrating the Secure Email MVP Admin Authentication System with the React frontend components.

## 🔧 **Integration Steps**

### **1. Update Environment Configuration**

Add the admin authentication API endpoints to your environment configuration:

```javascript
// src/config/api.js
export const API_CONFIG = {
  // Existing endpoints...
  
  // Admin Authentication Endpoints
  ADMIN: {
    CHECK_SETUP: '/admin/check-setup',
    SETUP: '/admin/setup',
    LOGIN: '/admin/login',
    LOGOUT: '/admin/logout',
    SESSION: '/admin/session',
    AUDIT_LOGS: '/admin/audit-logs'
  }
};
```

### **2. Create Admin Authentication Service**

Create a dedicated service for admin authentication:

```typescript
// src/services/adminAuthService.ts
import axios from 'axios';
import { API_CONFIG } from '../config/api';

export interface AdminSetupRequest {
  email: string;
  password: string;
}

export interface AdminLoginRequest {
  email: string;
  password: string;
  totp_code?: string;
}

export interface AdminUser {
  id: string;
  email: string;
  role: string;
  totp_enabled: boolean;
  is_active: boolean;
  created_at: string;
}

export interface AdminSession {
  session_token: string;
  admin: AdminUser;
}

export class AdminAuthService {
  private baseURL: string;
  private sessionToken: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    this.sessionToken = localStorage.getItem('adminSessionToken');
  }

  // Check if admin setup is required
  async checkSetup(): Promise<{ setup_required: boolean; root_admin_email: string }> {
    const response = await axios.get(`${this.baseURL}${API_CONFIG.ADMIN.CHECK_SETUP}`);
    return response.data;
  }

  // Create root admin (first-time setup)
  async setupAdmin(request: AdminSetupRequest): Promise<{ success: boolean; admin_id: string }> {
    const response = await axios.post(`${this.baseURL}${API_CONFIG.ADMIN.SETUP}`, request);
    return response.data;
  }

  // Admin login
  async login(request: AdminLoginRequest): Promise<AdminSession> {
    const response = await axios.post(`${this.baseURL}${API_CONFIG.ADMIN.LOGIN}`, request);
    const data = response.data;
    
    if (data.success) {
      this.sessionToken = data.session_token;
      localStorage.setItem('adminSessionToken', data.session_token);
    }
    
    return data;
  }

  // Validate admin session
  async validateSession(): Promise<AdminUser | null> {
    if (!this.sessionToken) {
      return null;
    }

    try {
      const response = await axios.get(`${this.baseURL}${API_CONFIG.ADMIN.SESSION}`, {
        headers: {
          'Authorization': `Bearer ${this.sessionToken}`
        }
      });
      return response.data.admin;
    } catch (error) {
      this.logout();
      return null;
    }
  }

  // Admin logout
  async logout(): Promise<void> {
    if (this.sessionToken) {
      try {
        await axios.post(`${this.baseURL}${API_CONFIG.ADMIN.LOGOUT}`, {
          session_token: this.sessionToken
        });
      } catch (error) {
        // Ignore logout errors
      }
    }
    
    this.sessionToken = null;
    localStorage.removeItem('adminSessionToken');
  }

  // Get audit logs
  async getAuditLogs(limit: number = 100): Promise<any[]> {
    if (!this.sessionToken) {
      throw new Error('No active session');
    }

    const response = await axios.get(`${this.baseURL}${API_CONFIG.ADMIN.AUDIT_LOGS}?limit=${limit}`, {
      headers: {
        'Authorization': `Bearer ${this.sessionToken}`
      }
    });
    return response.data.logs;
  }

  // Get current session token
  getSessionToken(): string | null {
    return this.sessionToken;
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!this.sessionToken;
  }
}

// Create singleton instance
export const adminAuthService = new AdminAuthService(process.env.REACT_APP_API_URL || 'http://localhost:8080');
```

### **3. Create Admin Setup Component**

```typescript
// src/components/admin/AdminSetup.tsx
import React, { useState } from 'react';
import { adminAuthService } from '../../services/adminAuthService';

interface AdminSetupProps {
  onSetupComplete: () => void;
  rootAdminEmail: string;
}

export const AdminSetup: React.FC<AdminSetupProps> = ({ onSetupComplete, rootAdminEmail }) => {
  const [email, setEmail] = useState(rootAdminEmail);
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const validatePassword = (password: string): string[] => {
    const errors: string[] = [];
    
    if (password.length < 16) {
      errors.push('Password must be at least 16 characters long');
    }
    
    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter');
    }
    
    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter');
    }
    
    if (!/\d/.test(password)) {
      errors.push('Password must contain at least one digit');
    }
    
    if (!/[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]/.test(password)) {
      errors.push('Password must contain at least one special character');
    }
    
    return errors;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    // Validate email
    if (email !== rootAdminEmail) {
      setError('Invalid root admin email');
      setLoading(false);
      return;
    }

    // Validate password
    const passwordErrors = validatePassword(password);
    if (passwordErrors.length > 0) {
      setError(passwordErrors.join(', '));
      setLoading(false);
      return;
    }

    // Validate password confirmation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      setLoading(false);
      return;
    }

    try {
      await adminAuthService.setupAdmin({ email, password });
      onSetupComplete();
    } catch (error: any) {
      setError(error.response?.data || 'Failed to create admin account');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Secure Admin Setup
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Create the initial root admin account for the Secure Email MVP system
          </p>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="email" className="sr-only">Email address</label>
              <input
                id="email"
                name="email"
                type="email"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Root admin email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled
              />
            </div>
            <div>
              <label htmlFor="password" className="sr-only">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Password (min 16 chars)"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="confirm-password" className="sr-only">Confirm Password</label>
              <input
                id="confirm-password"
                name="confirm-password"
                type="password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Confirm password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>
          </div>

          {error && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="text-sm text-red-700">{error}</div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
            >
              {loading ? 'Creating Admin Account...' : 'Create Root Admin'}
            </button>
          </div>
        </form>

        <div className="mt-6">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-gray-50 text-gray-500">Security Requirements</span>
            </div>
          </div>
          <div className="mt-4 text-xs text-gray-600">
            <ul className="list-disc list-inside space-y-1">
              <li>Password must be at least 16 characters long</li>
              <li>Must contain uppercase and lowercase letters</li>
              <li>Must contain at least one digit</li>
              <li>Must contain at least one special character</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};
```

### **4. Update Admin Login Component**

```typescript
// src/components/admin/AdminLogin.tsx
import React, { useState } from 'react';
import { adminAuthService } from '../../services/adminAuthService';

interface AdminLoginProps {
  onLoginSuccess: (admin: any) => void;
}

export const AdminLogin: React.FC<AdminLoginProps> = ({ onLoginSuccess }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await adminAuthService.login({
        email,
        password,
        totp_code: totpCode
      });
      
      onLoginSuccess(response.admin);
    } catch (error: any) {
      setError(error.response?.data || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Admin Login
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Sign in to the Secure Email MVP Admin Dashboard
          </p>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <label htmlFor="email" className="sr-only">Email address</label>
              <input
                id="email"
                name="email"
                type="email"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Admin email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="password" className="sr-only">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="totp" className="sr-only">TOTP Code</label>
              <input
                id="totp"
                name="totp"
                type="text"
                className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="TOTP Code (optional)"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
              />
            </div>
          </div>

          {error && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="text-sm text-red-700">{error}</div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
```

### **5. Create Session Validation Hook**

```typescript
// src/hooks/useAdminAuth.ts
import { useState, useEffect } from 'react';
import { adminAuthService } from '../services/adminAuthService';

export const useAdminAuth = () => {
  const [admin, setAdmin] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    validateSession();
  }, []);

  const validateSession = async () => {
    try {
      setLoading(true);
      const adminUser = await adminAuthService.validateSession();
      setAdmin(adminUser);
      setError(null);
    } catch (err) {
      setAdmin(null);
      setError('Session validation failed');
    } finally {
      setLoading(false);
    }
  };

  const login = async (email: string, password: string, totpCode?: string) => {
    try {
      setLoading(true);
      const response = await adminAuthService.login({ email, password, totp_code: totpCode });
      setAdmin(response.admin);
      setError(null);
      return response;
    } catch (err: any) {
      setError(err.response?.data || 'Login failed');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      await adminAuthService.logout();
    } catch (err) {
      // Ignore logout errors
    } finally {
      setAdmin(null);
      setError(null);
    }
  };

  return {
    admin,
    loading,
    error,
    login,
    logout,
    validateSession,
    isAuthenticated: !!admin
  };
};
```

### **6. Update Main Admin App Component**

```typescript
// src/components/admin/AdminApp.tsx
import React, { useState, useEffect } from 'react';
import { AdminSetup } from './AdminSetup';
import { AdminLogin } from './AdminLogin';
import { EnterpriseDashboard } from './EnterpriseDashboard';
import { adminAuthService } from '../../services/adminAuthService';
import { useAdminAuth } from '../../hooks/useAdminAuth';

export const AdminApp: React.FC = () => {
  const [setupRequired, setSetupRequired] = useState<boolean | null>(null);
  const [rootAdminEmail, setRootAdminEmail] = useState('');
  const { admin, loading, login, logout, isAuthenticated } = useAdminAuth();

  useEffect(() => {
    checkSetupStatus();
  }, []);

  const checkSetupStatus = async () => {
    try {
      const status = await adminAuthService.checkSetup();
      setSetupRequired(status.setup_required);
      setRootAdminEmail(status.root_admin_email);
    } catch (error) {
      console.error('Failed to check setup status:', error);
    }
  };

  const handleSetupComplete = () => {
    setSetupRequired(false);
    // Redirect to login
  };

  const handleLoginSuccess = (adminUser: any) => {
    // Login handled by useAdminAuth hook
  };

  const handleLogout = async () => {
    await logout();
  };

  if (loading || setupRequired === null) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (setupRequired) {
    return (
      <AdminSetup
        onSetupComplete={handleSetupComplete}
        rootAdminEmail={rootAdminEmail}
      />
    );
  }

  if (!isAuthenticated) {
    return (
      <AdminLogin onLoginSuccess={handleLoginSuccess} />
    );
  }

  return (
    <EnterpriseDashboard
      admin={admin}
      onLogout={handleLogout}
    />
  );
};
```

### **7. Update Audit Logs Panel**

```typescript
// src/components/admin/panels/AuditLogsPanel.tsx
import React, { useState, useEffect } from 'react';
import { adminAuthService } from '../../../services/adminAuthService';

export const AuditLogsPanel: React.FC = () => {
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAuditLogs();
  }, []);

  const loadAuditLogs = async () => {
    try {
      setLoading(true);
      const auditLogs = await adminAuthService.getAuditLogs(100);
      setLogs(auditLogs);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load audit logs');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Audit Logs</h3>
        <div className="text-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading audit logs...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Audit Logs</h3>
        <div className="text-center py-8">
          <p className="text-red-600">{error}</p>
          <button
            onClick={loadAuditLogs}
            className="mt-2 px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium text-gray-900">Audit Logs</h3>
        <button
          onClick={loadAuditLogs}
          className="px-3 py-1 bg-indigo-600 text-white rounded text-sm hover:bg-indigo-700"
        >
          Refresh
        </button>
      </div>
      
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Action
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                IP Address
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Timestamp
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {logs.map((log) => (
              <tr key={log.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {log.action}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                    log.success 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {log.success ? 'Success' : 'Failed'}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {log.ip_address}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {new Date(log.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      
      {logs.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No audit logs found
        </div>
      )}
    </div>
  );
};
```

## 🔒 **Security Considerations**

### **HttpOnly Cookies**
For production, configure the backend to set HttpOnly secure cookies:

```typescript
// Backend configuration (Go)
http.SetCookie(w, &http.Cookie{
    Name:     "admin_session",
    Value:    sessionToken,
    HttpOnly: true,
    Secure:   true, // HTTPS only
    SameSite: http.SameSiteStrictMode,
    MaxAge:   1800, // 30 minutes
})
```

### **CSRF Protection**
Implement CSRF protection for admin endpoints:

```typescript
// Add CSRF token to requests
const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');

axios.defaults.headers.common['X-CSRF-Token'] = csrfToken;
```

### **Session Management**
Implement automatic session refresh and logout on inactivity:

```typescript
// Auto-logout after 30 minutes of inactivity
useEffect(() => {
  const timeout = setTimeout(() => {
    logout();
  }, 30 * 60 * 1000); // 30 minutes

  const resetTimeout = () => {
    clearTimeout(timeout);
    setTimeout(() => {
      logout();
    }, 30 * 60 * 1000);
  };

  window.addEventListener('mousemove', resetTimeout);
  window.addEventListener('keypress', resetTimeout);

  return () => {
    clearTimeout(timeout);
    window.removeEventListener('mousemove', resetTimeout);
    window.removeEventListener('keypress', resetTimeout);
  };
}, [logout]);
```

## 🚀 **Deployment Checklist**

- [ ] Environment variables configured
- [ ] API endpoints updated
- [ ] Admin authentication service implemented
- [ ] Admin setup component created
- [ ] Admin login component updated
- [ ] Session validation hook implemented
- [ ] Audit logs panel integrated
- [ ] Security features configured
- [ ] Error handling implemented
- [ ] Loading states added
- [ ] Responsive design verified
- [ ] Accessibility features added
- [ ] Unit tests written
- [ ] Integration tests performed

## 📚 **Additional Resources**

- [Admin Authentication System Documentation](../admin-authentication-system.md)
- [API Endpoint Reference](../admin-authentication-system.md#api-endpoints)
- [Security Considerations](../admin-authentication-system.md#security-considerations)
- [Troubleshooting Guide](../admin-authentication-system.md#troubleshooting)
