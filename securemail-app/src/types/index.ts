// TypeScript types for SecureMail Frontend

export interface User {
  id: string;
  username: string;
  email: string;
  accountType: AccountType;
  emailVerified: boolean;
  mfaEnabled: boolean;
  organizationId?: string;
  organizationRole?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expiresAt: string;
  accountType: AccountType;
  organizationId?: string;
  organizationRole?: string;
}

export interface SignupRequest {
  username: string;
  email: string;
  password: string;
  accountType: AccountType;
}

export interface SignupResponse {
  id: string;
  username: string;
  email: string;
  accountType: AccountType;
  createdAt: string;
}

export interface TrialWarning {
  subscriptionId: string;
  plan: string;
  daysRemaining: number;
  expiryDate: string;
  warningLevel: TrialWarningLevel;
}

export interface InboxFolder {
  id: string;
  name: string;
  folderType: string;
  sortOrder: number;
  createdAt: string;
}

export interface EmailMessage {
  id: string;
  messageId: string;
  from: string;
  to: string;
  subject: string;
  bodyType: string;
  sizeBytes: number;
  isRead: boolean;
  isImportant: boolean;
  isStarred: boolean;
  receivedAt: string;
  createdAt: string;
}

export class ApiError extends Error {
  public code?: string;
  public status?: number;

  constructor(message: string, code?: string, status?: number) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

export interface AppState {
  auth: AuthState;
  trialWarning: TrialWarning | null;
  inboxFolders: InboxFolder[];
  inboxMessages: EmailMessage[];
}

// Navigation types
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  Signup: undefined;
  Login: undefined;
  ForgotPassword: undefined;
  Dashboard: undefined;
  Inbox: undefined;
  Settings: undefined;
};

export type AuthStackParamList = {
  Signup: undefined;
  Login: undefined;
  ForgotPassword: undefined;
};

export type MainTabParamList = {
  Dashboard: undefined;
  Inbox: undefined;
  Settings: undefined;
};

// Import types from config
import { AccountType, TrialWarningLevel } from '../config/api';
