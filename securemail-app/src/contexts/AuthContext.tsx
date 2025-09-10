// Authentication Context for SecureMail Frontend
import React, { createContext, useContext, useReducer, useEffect, ReactNode } from 'react';
import { AuthState, User, LoginRequest, SignupRequest, SignupResponse } from '../types';
import { apiService } from '../services/api';
import { storageService } from '../services/storage';

// Auth actions
type AuthAction =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: { user: User; token: string } }
  | { type: 'AUTH_FAILURE'; payload: string }
  | { type: 'AUTH_LOGOUT' }
  | { type: 'AUTH_CLEAR_ERROR' }
  | { type: 'AUTH_UPDATE_USER'; payload: User };

// Initial state
const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,
};

// Auth reducer
function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'AUTH_START':
      return {
        ...state,
        isLoading: true,
        error: null,
      };
    case 'AUTH_SUCCESS':
      return {
        ...state,
        user: action.payload.user,
        token: action.payload.token,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      };
    case 'AUTH_FAILURE':
      return {
        ...state,
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: action.payload,
      };
    case 'AUTH_LOGOUT':
      return {
        ...state,
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      };
    case 'AUTH_CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };
    case 'AUTH_UPDATE_USER':
      return {
        ...state,
        user: action.payload,
      };
    default:
      return state;
  }
}

// Auth context type
interface AuthContextType {
  state: AuthState;
  signup: (data: SignupRequest) => Promise<SignupResponse>;
  login: (data: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
  updateUser: (user: User) => void;
  checkAuthStatus: () => Promise<void>;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Check authentication status on app start
  useEffect(() => {
    checkAuthStatus();
  }, []);

  // Check if user is already authenticated
  const checkAuthStatus = async () => {
    try {
      dispatch({ type: 'AUTH_START' });
      
      const { token, user } = await storageService.getSession();
      
      if (token && user) {
        dispatch({ 
          type: 'AUTH_SUCCESS', 
          payload: { user, token } 
        });
      } else {
        dispatch({ type: 'AUTH_FAILURE', payload: 'No valid session found' });
      }
    } catch (error) {
      console.error('Error checking auth status:', error);
      dispatch({ type: 'AUTH_FAILURE', payload: 'Failed to check authentication status' });
    }
  };

  // Signup function
  const signup = async (data: SignupRequest) => {
    try {
      dispatch({ type: 'AUTH_START' });
      
      const response = await apiService.signup(data);
      
      // Create user object from signup response
      const user: User = {
        id: response.id,
        username: response.username,
        email: response.email,
        accountType: response.accountType,
        emailVerified: false,
        mfaEnabled: false,
      };
      
      // Note: We don't store session yet since email verification is required
      dispatch({ type: 'AUTH_SUCCESS', payload: { user, token: '' } });
      
      // Return the response with recovery key for navigation
      return response;
      
    } catch (error: any) {
      console.error('Signup error:', error);
      dispatch({ 
        type: 'AUTH_FAILURE', 
        payload: error.message || 'Signup failed' 
      });
      throw error;
    }
  };

  // Login function
  const login = async (data: LoginRequest) => {
    try {
      dispatch({ type: 'AUTH_START' });
      
      const response = await apiService.login(data);
      
      // Create user object from login response
      const user: User = {
        id: '', // Will be populated from token validation
        username: '', // Will be populated from token validation
        email: data.email,
        accountType: response.accountType,
        emailVerified: true, // If login succeeded, email is verified
        mfaEnabled: false, // Will be updated based on MFA status
        organizationId: response.organizationId,
        organizationRole: response.organizationRole,
      };
      
      // Store session data
      await storageService.storeSession(response.token, user);
      
      dispatch({ 
        type: 'AUTH_SUCCESS', 
        payload: { user, token: response.token } 
      });
      
    } catch (error: any) {
      console.error('Login error:', error);
      dispatch({ 
        type: 'AUTH_FAILURE', 
        payload: error.message || 'Login failed' 
      });
      throw error;
    }
  };

  // Logout function
  const logout = async () => {
    try {
      await storageService.clearSession();
      dispatch({ type: 'AUTH_LOGOUT' });
    } catch (error) {
      console.error('Logout error:', error);
      // Still dispatch logout even if storage clear fails
      dispatch({ type: 'AUTH_LOGOUT' });
    }
  };

  // Clear error function
  const clearError = () => {
    dispatch({ type: 'AUTH_CLEAR_ERROR' });
  };

  // Update user function
  const updateUser = (user: User) => {
    dispatch({ type: 'AUTH_UPDATE_USER', payload: user });
  };

  const value: AuthContextType = {
    state,
    signup,
    login,
    logout,
    clearError,
    updateUser,
    checkAuthStatus,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

// Custom hook to use auth context
export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
