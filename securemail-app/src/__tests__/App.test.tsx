// Basic App Test for SecureMail
import React from 'react';
import { render } from '@testing-library/react-native';
import App from '../../App';

// Mock the navigation
jest.mock('@react-navigation/native', () => ({
  NavigationContainer: ({ children }: { children: React.ReactNode }) => children,
}));

// Mock the auth context
jest.mock('../contexts/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => children,
  useAuth: () => ({
    state: {
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    },
    signup: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    clearError: jest.fn(),
    updateUser: jest.fn(),
    checkAuthStatus: jest.fn(),
  }),
}));

describe('App', () => {
  it('renders without crashing', () => {
    const { getByText } = render(<App />);
    // The app should render without throwing an error
    expect(getByText).toBeDefined();
  });
});
