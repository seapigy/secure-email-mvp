// API Service for SecureMail Frontend
import { API_CONFIG, AccountType } from '../config/api';
import { 
  LoginRequest, 
  LoginResponse, 
  SignupRequest, 
  SignupResponse, 
  TrialWarning, 
  InboxFolder, 
  EmailMessage,
  ApiError 
} from '../types';

class ApiService {
  private baseURL: string;
  private timeout: number;

  constructor() {
    this.baseURL = API_CONFIG.BASE_URL;
    this.timeout = API_CONFIG.TIMEOUT;
  }

  private async makeRequest<T>(
    endpoint: string, 
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const config: RequestInit = {
      ...options,
      headers: {
        ...API_CONFIG.DEFAULT_HEADERS,
        ...options.headers,
      },
      signal: AbortSignal.timeout(this.timeout),
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new ApiError(
          errorData.message || `HTTP ${response.status}: ${response.statusText}`,
          errorData.code,
          response.status
        );
      }

      return await response.json();
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      // Handle network errors
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new ApiError('Network error. Please check your connection.');
      }
      
      // Handle timeout errors
      if (error instanceof Error && error.name === 'AbortError') {
        throw new ApiError('Request timeout. Please try again.');
      }
      
      throw new ApiError('An unexpected error occurred.');
    }
  }

  private async makeAuthenticatedRequest<T>(
    endpoint: string, 
    token: string,
    options: RequestInit = {}
  ): Promise<T> {
    return this.makeRequest<T>(endpoint, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${token}`,
      },
    });
  }

  // Authentication methods
  async signup(data: SignupRequest): Promise<SignupResponse> {
    return this.makeRequest<SignupResponse>(API_CONFIG.ENDPOINTS.SIGNUP, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async login(data: LoginRequest): Promise<LoginResponse> {
    return this.makeRequest<LoginResponse>(API_CONFIG.ENDPOINTS.LOGIN, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async verifyEmail(code: string): Promise<{ success: boolean }> {
    return this.makeRequest<{ success: boolean }>(API_CONFIG.ENDPOINTS.VERIFY_EMAIL, {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async resendVerification(): Promise<{ success: boolean }> {
    return this.makeRequest<{ success: boolean }>(API_CONFIG.ENDPOINTS.RESEND_VERIFICATION, {
      method: 'POST',
    });
  }

  // Inbox methods
  async getInboxFolders(token: string): Promise<{ folders: InboxFolder[] }> {
    return this.makeAuthenticatedRequest<{ folders: InboxFolder[] }>(
      API_CONFIG.ENDPOINTS.INBOX_FOLDERS,
      token
    );
  }

  async getInboxMessages(token: string, folderId: string): Promise<{ messages: EmailMessage[] }> {
    return this.makeAuthenticatedRequest<{ messages: EmailMessage[] }>(
      `${API_CONFIG.ENDPOINTS.INBOX_MESSAGES}?folder_id=${folderId}`,
      token
    );
  }

  // Trial management
  async getTrialWarning(token: string): Promise<{ has_warning: boolean; warning?: TrialWarning }> {
    return this.makeAuthenticatedRequest<{ has_warning: boolean; warning?: TrialWarning }>(
      API_CONFIG.ENDPOINTS.TRIAL_WARNING,
      token
    );
  }

  async extendTrial(token: string): Promise<{ success: boolean; message: string }> {
    return this.makeAuthenticatedRequest<{ success: boolean; message: string }>(
      API_CONFIG.ENDPOINTS.EXTEND_TRIAL,
      token,
      { method: 'POST' }
    );
  }

  // Account management
  async upgradeAccount(token: string, plan: AccountType): Promise<{ success: boolean }> {
    return this.makeAuthenticatedRequest<{ success: boolean }>(
      API_CONFIG.ENDPOINTS.UPGRADE_ACCOUNT,
      token,
      {
        method: 'POST',
        body: JSON.stringify({ plan }),
      }
    );
  }

  async downgradeAccount(token: string): Promise<{ success: boolean }> {
    return this.makeAuthenticatedRequest<{ success: boolean }>(
      API_CONFIG.ENDPOINTS.DOWNGRADE_ACCOUNT,
      token,
      { method: 'POST' }
    );
  }

  // Organization management
  async createOrganization(token: string, name: string): Promise<{ success: boolean; organizationId: string }> {
    return this.makeAuthenticatedRequest<{ success: boolean; organizationId: string }>(
      API_CONFIG.ENDPOINTS.CREATE_ORGANIZATION,
      token,
      {
        method: 'POST',
        body: JSON.stringify({ name }),
      }
    );
  }

  async addUserToOrganization(token: string, email: string, role: string): Promise<{ success: boolean }> {
    return this.makeAuthenticatedRequest<{ success: boolean }>(
      API_CONFIG.ENDPOINTS.ADD_USER_TO_ORG,
      token,
      {
        method: 'POST',
        body: JSON.stringify({ email, role }),
      }
    );
  }

  async listOrganizationUsers(token: string): Promise<{ users: any[] }> {
    return this.makeAuthenticatedRequest<{ users: any[] }>(
      API_CONFIG.ENDPOINTS.LIST_ORG_USERS,
      token
    );
  }
}

// Export singleton instance
export const apiService = new ApiService();
export default apiService;
