import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { toast } from 'react-toastify';
import Login from '../components/Login';
import axios from 'axios';

// Mock dependencies
jest.mock('axios');
jest.mock('react-toastify', () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
  ToastContainer: () => <div data-testid="toast-container" />,
}));
jest.mock('gsap', () => ({
  fromTo: jest.fn(),
  to: jest.fn(),
}));

// Mock sessionStorage
const mockSessionStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'sessionStorage', {
  value: mockSessionStorage,
});

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage,
});

const renderLogin = () => {
  return render(
    <MemoryRouter>
      <Login />
    </MemoryRouter>
  );
};

describe('Login Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockSessionStorage.getItem.mockReturnValue(null);
    mockLocalStorage.getItem.mockReturnValue('light');
  });

  test('renders login form with all fields', () => {
    renderLogin();
    
    expect(screen.getByLabelText(/email input/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password input/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/totp code input/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  test('shows error toast on invalid email format', async () => {
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'invalid@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Invalid email format. Must be user@securesystem.email',
        expect.any(Object)
      );
    });
  });

  test('shows error toast on invalid password length', async () => {
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'short' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Password must be 8–128 characters',
        expect.any(Object)
      );
    });
  });

  test('shows error toast on invalid TOTP format', async () => {
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '12345' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'TOTP code must be 6 digits',
        expect.any(Object)
      );
    });
  });

  test('submits form with valid inputs', async () => {
    const mockResponse = { data: { token: 'jwt-token', user_id: 'user-123' } };
    axios.post.mockResolvedValue(mockResponse);
    
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(axios.post).toHaveBeenCalledWith(
        'https://api.securesystem.email/api/auth/login',
        {
          email: 'test@securesystem.email',
          password: 'password123',
          totp_code: '123456',
        }
      );
    });

    expect(mockSessionStorage.setItem).toHaveBeenCalledWith('token', 'jwt-token');
    expect(mockSessionStorage.setItem).toHaveBeenCalledWith('user_id', 'user-123');
    expect(toast.success).toHaveBeenCalledWith(
      'Login successful! Redirecting...',
      expect.any(Object)
    );
  });

  test('handles API error responses', async () => {
    const mockError = {
      response: {
        status: 401,
        data: { error: 'Invalid credentials' }
      }
    };
    axios.post.mockRejectedValue(mockError);
    
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Invalid credentials. Please check your email, password, and TOTP code.',
        expect.any(Object)
      );
    });
  });

  test('handles network errors', async () => {
    const mockError = { request: {} };
    axios.post.mockRejectedValue(mockError);
    
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Unable to connect to server. Please check your internet connection.',
        expect.any(Object)
      );
    });
  });

  test('filters non-numeric characters from TOTP input', () => {
    renderLogin();
    
    const totpInput = screen.getByLabelText(/totp code input/i);
    
    fireEvent.change(totpInput, { target: { value: '123abc456' } });
    
    expect(totpInput.value).toBe('123456');
  });

  test('shows loading state during submission', async () => {
    axios.post.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    
    renderLogin();
    
    const emailInput = screen.getByLabelText(/email input/i);
    const passwordInput = screen.getByLabelText(/password input/i);
    const totpInput = screen.getByLabelText(/totp code input/i);
    const submitButton = screen.getByRole('button', { name: /sign in/i });

    fireEvent.change(emailInput, { target: { value: 'test@securesystem.email' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(totpInput, { target: { value: '123456' } });
    fireEvent.click(submitButton);

    expect(screen.getByText('Signing In...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
  });
}); 