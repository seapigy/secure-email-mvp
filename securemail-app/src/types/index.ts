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
  fallback_email: string;
  setupMFA?: boolean;
}

export interface SignupResponse {
  id: string;
  username: string;
  email: string;
  accountType: AccountType;
  createdAt: string;
  mfa?: {
    secret: string;
    qr_code_url: string;
    backup_codes: string[];
  };
  // recovery_key: string; // Will be sent after email verification
}

export interface VerifyEmailRequest {
  user_id: string;
  code: string;
}

export interface VerifyEmailResponse {
  success: boolean;
  message: string;
  recovery_key: string;
}

export interface RecoveryRequest {
  fallback_email: string;
  recovery_key: string;
  action: 'reset_password' | 'reset_email';
  new_password?: string;
  new_email?: string;
}

export interface RecoveryResponse {
  success: boolean;
  message: string;
  user_id?: string;
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
  Signup: { selectedPlan?: AccountType };
  Login: undefined;
  ForgotPassword: undefined;
  PlanSelection: undefined;
  EmailVerification: { userId: string; username: string; email: string };
  RecoveryKey: { recoveryKey: string; username: string; email: string };
  AccountRecovery: undefined;
};

export type MainTabParamList = {
  Dashboard: undefined;
  Inbox: undefined;
  Settings: undefined;
};

// Import types from config
import { AccountType, TrialWarningLevel } from '../config/api';
