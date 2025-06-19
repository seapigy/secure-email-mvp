import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import AuthCard from '../components/AuthCard';
import * as api from '../lib/api';

// Mock the API functions
jest.mock('../lib/api', () => ({
  login: jest.fn(),
  signup: jest.fn(),
  verifyTotp: jest.fn(),
}));

// Mock react-toastify
jest.mock('react-toastify', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
    info: jest.fn(),
  },
  ToastContainer: () => <div data-testid="toast-container" />,
}));

// Mock GSAP
jest.mock('gsap', () => ({
  fromTo: jest.fn(),
}));

const renderWithRouter = (component) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('AuthCard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders login form by default', () => {
    renderWithRouter(<AuthCard />);
    
    expect(screen.getByText('Login')).toBeInTheDocument();
    expect(screen.getByText('Sign Up')).toBeInTheDocument();
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('TOTP code')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sign In' })).toBeInTheDocument();
  });

  test('switches to sign-up form when sign-up tab is clicked', () => {
    renderWithRouter(<AuthCard />);
    
    const signUpTab = screen.getByText('Sign Up');
    fireEvent.click(signUpTab);
    
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sign Up' })).toBeInTheDocument();
  });

  test('has proper accessibility attributes', () => {
    renderWithRouter(<AuthCard />);
    
    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Password');
    const totpInput = screen.getByLabelText('TOTP code');
    
    expect(emailInput).toHaveAttribute('type', 'email');
    expect(passwordInput).toHaveAttribute('type', 'password');
    expect(totpInput).toHaveAttribute('maxLength', '6');
    expect(totpInput).toHaveAttribute('required');
  });

  test('applies correct CSS classes for neumorphic design', () => {
    renderWithRouter(<AuthCard />);
    
    const authCard = screen.getByRole('form').closest('.auth-card');
    expect(authCard).toHaveClass('auth-card');
    expect(authCard).toHaveClass('bg-white');
    expect(authCard).toHaveClass('rounded-xl');
  });
});

test('shows error toast on invalid sign-up', async () => {
  render(<AuthCard />, { wrapper: MemoryRouter });
  fireEvent.click(screen.getByText(/sign up/i));
  fireEvent.change(screen.getByLabelText(/email input/i), { target: { value: 'invalid@example.com' } });
  fireEvent.change(screen.getByLabelText(/password input/i), { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText(/confirm password input/i), { target: { value: 'password123' } });
  fireEvent.click(screen.getByText(/sign up/i));
  expect(await screen.findByText(/invalid email format/i)).toBeInTheDocument();
});

test('submits sign-up and shows QR code', async () => {
  api.signup.mockResolvedValue({ data: { temp_id: 'uuid', totp_qr: 'base64_png' } });
  render(<AuthCard />, { wrapper: MemoryRouter });
  fireEvent.click(screen.getByText(/sign up/i));
  fireEvent.change(screen.getByLabelText(/email input/i), { target: { value: 'test@securesystem.email' } });
  fireEvent.change(screen.getByLabelText(/password input/i), { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText(/confirm password input/i), { target: { value: 'password123' } });
  fireEvent.click(screen.getByText(/sign up/i));
  expect(api.signup).toHaveBeenCalledWith({
    email: 'test@securesystem.email',
    password: 'password123',
    confirm_password: 'password123',
  });
  expect(await screen.findByText(/scan this QR code/i)).toBeInTheDocument();
}); 