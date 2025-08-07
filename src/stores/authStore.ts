import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, AuthState, LoginCredentials, SignupData, AuthTokens } from '@/types';

interface AuthStore extends AuthState {
  // Actions
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

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (credentials: LoginCredentials) => {
        set({ isLoading: true, error: null });
        try {
          // Demo login for development - bypass API
          if (credentials.email === 'demo@securesystem.email' && credentials.password === 'demo123' && credentials.totpCode === '123456') {
            const mockUser: User = {
              id: 'demo-user-1',
              email: credentials.email,
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            };
            
            const mockTokens: AuthTokens = {
              accessToken: 'demo-access-token',
              refreshToken: 'demo-refresh-token',
            };
            
            // Store tokens securely
            sessionStorage.setItem('accessToken', mockTokens.accessToken);
            sessionStorage.setItem('refreshToken', mockTokens.refreshToken);
            
            set({
              user: mockUser,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            return;
          }
          
          // Use the API utility for non-demo login
          try {
            const { login: apiLogin, getUserProfile } = await import('@/lib/api');
            const data = await apiLogin(credentials);
            
            // Store tokens securely
            sessionStorage.setItem('accessToken', data.accessToken);
            sessionStorage.setItem('refreshToken', data.refreshToken);

            // Get user info
            const user = await getUserProfile();
            set({
              user,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          } catch (apiError) {
            console.error('API login failed:', apiError);
            set({
              isLoading: false,
              error: 'API connection failed. Please try again later.',
            });
          }
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Login failed',
          });
        }
      },

      signup: async (data: SignupData) => {
        set({ isLoading: true, error: null });
        try {
          // Note: This is a placeholder for the actual signup API call
          const response = await fetch('/api/auth/signup', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
          });

          if (!response.ok) {
            throw new Error('Signup failed');
          }

          set({ isLoading: false, error: null });
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Signup failed',
          });
        }
      },

      logout: () => {
        // Clear tokens
        sessionStorage.removeItem('accessToken');
        sessionStorage.removeItem('refreshToken');
        
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false,
          error: null,
        });
      },

      refreshToken: async () => {
        const refreshToken = sessionStorage.getItem('refreshToken');
        if (!refreshToken) {
          get().logout();
          return;
        }

        try {
          const { refreshToken: apiRefreshToken } = await import('@/lib/api');
          const data = await apiRefreshToken(refreshToken);
          sessionStorage.setItem('accessToken', data.accessToken);
        } catch (error) {
          get().logout();
        }
      },

      setUser: (user: User) => {
        set({ user, isAuthenticated: true });
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },

      updateUser: (updates: Partial<User>) => {
        const { user } = get();
        if (user) {
          set({ user: { ...user, ...updates } });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
); 