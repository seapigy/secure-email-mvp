import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import AdminLogin from '../../pages/admin/Login';

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock axios
vi.mock('axios', () => ({
  default: {
    post: vi.fn(),
  },
}));

// Mock react-toastify
vi.mock('react-toastify', () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

const renderWithRouter = (component: React.ReactElement) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('AdminLogin', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form', () => {
    renderWithRouter(<AdminLogin />);
    
    expect(screen.getByText('Admin Login')).toBeInTheDocument();
    expect(screen.getByText('Secure Email MVP Administration')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('handles valid login', async () => {
    const mockResponse = {
      data: {
        success: true,
        message: 'Login successful',
        token: 'test-token',
        admin: {
          id: 1,
          email: 'admin@test.com',
          role: 'admin',
        },
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.post).mockResolvedValueOnce(mockResponse);

    renderWithRouter(<AdminLogin />);

    // Fill form
    fireEvent.change(screen.getByLabelText('Email Address'), {
      target: { value: 'admin@test.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'admin123456' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(axios.default.post).toHaveBeenCalledWith('/api/admin/login', {
        email: 'admin@test.com',
        password: 'admin123456',
      });
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/admin/dashboard');
    });
  });

  it('handles invalid login', async () => {
    const mockError = {
      response: {
        data: {
          message: 'Invalid credentials',
        },
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.post).mockRejectedValueOnce(mockError);

    renderWithRouter(<AdminLogin />);

    // Fill form
    fireEvent.change(screen.getByLabelText('Email Address'), {
      target: { value: 'admin@test.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'wrongpassword' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(axios.default.post).toHaveBeenCalledWith('/api/admin/login', {
        email: 'admin@test.com',
        password: 'wrongpassword',
      });
    });

    // Should not navigate on error
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('validates email format', async () => {
    renderWithRouter(<AdminLogin />);

    // Fill form with invalid email
    fireEvent.change(screen.getByLabelText('Email Address'), {
      target: { value: 'invalid-email' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'admin123456' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    // Should not make API call with invalid email
    const axios = await import('axios');
    expect(axios.default.post).not.toHaveBeenCalled();
  });

  it('validates password length', async () => {
    renderWithRouter(<AdminLogin />);

    // Fill form with short password
    fireEvent.change(screen.getByLabelText('Email Address'), {
      target: { value: 'admin@test.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: '123' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    // Should not make API call with short password
    const axios = await import('axios');
    expect(axios.default.post).not.toHaveBeenCalled();
  });

  it('shows loading state during login', async () => {
    // Mock a slow response
    const axios = await import('axios');
    vi.mocked(axios.default.post).mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    renderWithRouter(<AdminLogin />);

    // Fill form
    fireEvent.change(screen.getByLabelText('Email Address'), {
      target: { value: 'admin@test.com' },
    });
    fireEvent.change(screen.getByLabelText('Password'), {
      target: { value: 'admin123456' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    // Should show loading state
    expect(screen.getByText('Signing In...')).toBeInTheDocument();
  });
});
