// =============================================================================
// SECURE EMAIL MVP - TYPE DEFINITIONS
// =============================================================================

// User-related types
export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

// Authentication types
export interface LoginCredentials {
  email: string;
  password: string;
  totpCode: string;
}

export interface SignupData {
  email: string;
  password: string;
  confirm_password: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

// Theme types
export type Theme = 'light' | 'dark' | 'system';

// Email types
export interface Email {
  id: string;
  from?: string;
  to?: string;
  subject: string;
  content: string;
  date?: string;
  read?: boolean;
  important?: boolean;
  attachments?: EmailAttachment[];
  // Additional properties used by components
  sender?: string;
  recipients?: string[];
  isRead?: boolean;
  isStarred?: boolean;
  isEncrypted?: boolean;
  hasAttachments?: boolean;
  receivedAt?: string;
  createdAt?: string;
  labels?: string[];
}

export interface EmailAttachment {
  id?: string;
  name: string;
  size: string | number;
  type?: string;
  isEncrypted?: boolean;
}

// API response types
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

import React from 'react';

// Navigation types
export interface NavigationItem {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  count?: number;
}

// Form types
export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'textarea' | 'select';
  placeholder?: string;
  required?: boolean;
  validation?: (value: string) => string | null;
}

// UI Component types
export interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export interface InputProps {
  id?: string;
  type?: 'text' | 'email' | 'password' | 'number';
  placeholder?: string;
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  error?: string;
  className?: string;
  required?: boolean;
  maxLength?: number;
}

// Layout types
export interface LayoutProps {
  children: React.ReactNode;
}

export interface HeaderProps {
  user: User | null;
  onLogout: () => void;
  onToggleTheme: () => void;
  theme: Theme;
}

export interface SidebarProps {
  // Add any specific sidebar props here
}

// Page component types
export interface PageProps {
  // Add any common page props here
}

// Hook types
export interface UseAuthReturn extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  signup: (data: SignupData) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  setUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
  updateUser: (updates: Partial<User>) => void;
}

export interface UseThemeReturn {
  theme: Theme;
  toggleTheme: () => void;
  setTheme: (theme: Theme) => void;
}

// Store types
export interface AuthStore extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  signup: (data: SignupData) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  setUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
  updateUser: (updates: Partial<User>) => void;
}

export interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  selectedEmails: string[];
  viewMode: 'list' | 'grid';
  sortOrder: 'asc' | 'desc';
  searchQuery: string;
  filters: EmailFilters;
  sidebarCollapsed: boolean;
  selectedFolder: string;
}

export interface EmailFilters {
  read?: boolean | null;
  starred?: boolean | null;
  encrypted?: boolean | null;
  hasAttachments?: boolean | null;
  dateRange?: {
    start: string | null;
    end: string | null;
  };
  labels?: string[];
}

export interface UIStore {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  // Additional properties used by components
  theme?: 'light' | 'dark';
  setTheme?: (theme: 'light' | 'dark') => void;
  selectedEmails?: string[];
  toggleEmailSelection?: (emailId: string) => void;
  clearEmailSelection?: () => void;
  viewMode?: 'list' | 'grid';
  sortOrder?: 'asc' | 'desc';
  searchQuery?: string;
  filters?: EmailFilters;
  sidebarCollapsed?: boolean;
  setSidebarCollapsed?: (collapsed: boolean) => void;
  selectedFolder?: string;
  setSelectedFolder?: (folderId: string) => void;
  setSelectedEmails?: (emailIds: string[]) => void;
  setViewMode?: (mode: 'list' | 'grid') => void;
  setSortBy?: (sortBy: string) => void;
  setSortOrder?: (order: 'asc' | 'desc') => void;
  setSearchQuery?: (query: string) => void;
  setFilters?: (filters: EmailFilters) => void;
}

// Utility types
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
export type MakeRequired<T, K extends keyof T> = T & { [P in K]-?: T[P] };

// ============================================================================
// ERROR TYPES
// ============================================================================

export interface APIError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, unknown>;
}

export interface ErrorResponse {
  error: string;
  code?: string;
  details?: Record<string, unknown>;
} 