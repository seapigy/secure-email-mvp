/**
 * Simplified Signup Flow Test Suite
 * 
 * Focused on core functionality without complex multi-step flows
 * Tests the essential signup validation and API integration
 */

import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import SignupPage from '../../pages/SignupPage';

// Mock fetch for API calls
const mockFetch = vi.fn();
Object.defineProperty(window, 'fetch', {
  value: mockFetch,
  writable: true,
});

// Mock logger to prevent console noise during tests
vi.mock('@/lib/logger', () => ({
  log: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock navigation
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Helper function to render SignupPage with router
const renderSignupPage = () => {
  return render(
    <BrowserRouter>
      <SignupPage />
    </BrowserRouter>
  );
};

describe('Signup Flow - Simplified Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  // ------------------------------
  // Test 1: Form Validation
  // ------------------------------
  it('should show validation errors for empty form submission', async () => {
    renderSignupPage();

    // Select Free Account
    const freePlanButton = screen.getByText(/free plan/i);
    fireEvent.click(freePlanButton);

    // Wait for form to be rendered
    await waitFor(() => {
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    // Submit empty form
    const form = document.querySelector('form');
    expect(form).toBeInTheDocument();
    fireEvent.submit(form!);

    // Check for validation errors (use getAllByText to handle multiple instances)
    await waitFor(() => {
      const emailErrors = screen.getAllByText(/email is required/i);
      const passwordErrors = screen.getAllByText(/password is required/i);
      expect(emailErrors.length).toBeGreaterThan(0);
      expect(passwordErrors.length).toBeGreaterThan(0);
    });
  });

  // ------------------------------
  // Test 2: Free Account Signup
  // ------------------------------
  it('should complete free account signup successfully', async () => {
    // Mock successful API response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'success',
        user_id: 'test-user-id-123',
        next_step: 'verify_email',
      }),
    });

    renderSignupPage();

    // Select Free Account
    const freePlanButton = screen.getByText(/free plan/i);
    fireEvent.click(freePlanButton);

    // Fill form
    await waitFor(() => {
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const fallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(emailInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(fallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Wait for review page
    await waitFor(() => {
      expect(screen.getByText(/review your account/i)).toBeInTheDocument();
    });

    // Click create account
    const createAccountButton = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(createAccountButton);

    // Verify API call
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/signup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          plan: 'free',
          email: 'testuser@securesystem.email',
          password: 'SecurePass123!',
          company_code: undefined,
        }),
      });
    });
  });

  // ------------------------------
  // Test 3: Company Account Signup
  // ------------------------------
  it('should complete company account signup successfully', async () => {
    // Mock successful API response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'success',
        user_id: 'test-user-id-789',
        next_step: 'verify_email',
      }),
    });

    renderSignupPage();

    // Select Company Account
    const companyPlanButton = screen.getByText(/enterprise plan/i);
    fireEvent.click(companyPlanButton);

    // Fill company form
    await waitFor(() => {
      expect(screen.getByText(/company account setup/i)).toBeInTheDocument();
    });

    const companyNameInput = screen.getByLabelText(/company name/i);
    const adminEmailInput = screen.getByLabelText(/admin email/i);
    const adminPasswordInput = screen.getByLabelText(/admin password/i);
    const adminConfirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const adminFallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(companyNameInput, { target: { value: 'Test Company' } });
    fireEvent.change(adminEmailInput, { target: { value: 'admin@testcompany.com' } });
    fireEvent.change(adminPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(adminConfirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(adminFallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Wait for review page
    await waitFor(() => {
      expect(screen.getByText(/review your account/i)).toBeInTheDocument();
    });

    // Click create account
    const createAccountButton = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(createAccountButton);

    // Verify API call
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/signup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          plan: 'company',
          email: 'admin@testcompany.com',
          password: 'SecurePass123!',
          company_code: 'Test Company',
        }),
      });
    });
  });

  // ------------------------------
  // Test 4: Company Validation
  // ------------------------------
  it('should validate company name requirement', async () => {
    renderSignupPage();

    // Select Company Account
    const companyPlanButton = screen.getByText(/enterprise plan/i);
    fireEvent.click(companyPlanButton);

    // Wait for company form
    await waitFor(() => {
      expect(screen.getByText(/company account setup/i)).toBeInTheDocument();
    });

    // Fill all fields except company name
    const adminEmailInput = screen.getByLabelText(/admin email/i);
    const adminPasswordInput = screen.getByLabelText(/admin password/i);
    const adminConfirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const adminFallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(adminEmailInput, { target: { value: 'admin@testcompany.com' } });
    fireEvent.change(adminPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(adminConfirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(adminFallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form without company name
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Check for validation error - use simple text matching
    await waitFor(() => {
      expect(screen.getByText('Company name is required')).toBeInTheDocument();
    });
  });

  // ------------------------------
  // Test 5: API Error Handling
  // ------------------------------
  it('should handle API errors gracefully', async () => {
    // Mock API error response
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({
        error: 'User already exists',
      }),
    });

    renderSignupPage();

    // Complete free signup flow
    const freePlanButton = screen.getByText(/free plan/i);
    fireEvent.click(freePlanButton);

    await waitFor(() => {
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const fallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(emailInput, { target: { value: 'existinguser' } });
    fireEvent.change(passwordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(fallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Wait for review page
    await waitFor(() => {
      expect(screen.getByText(/review your account/i)).toBeInTheDocument();
    });

    // Click create account
    const createAccountButton = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(createAccountButton);

    // Check for error message - use simple text matching
    await waitFor(() => {
      expect(screen.getByText('User already exists')).toBeInTheDocument();
    });

    // Should not navigate to login
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  // ------------------------------
  // Test 6: Network Error Handling
  // ------------------------------
  it('should handle network errors gracefully', async () => {
    // Mock network error
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    renderSignupPage();

    // Complete free signup flow
    const freePlanButton = screen.getByText(/free plan/i);
    fireEvent.click(freePlanButton);

    await waitFor(() => {
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const fallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(emailInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(fallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Wait for review page
    await waitFor(() => {
      expect(screen.getByText(/review your account/i)).toBeInTheDocument();
    });

    // Click create account
    const createAccountButton = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(createAccountButton);

    // Check for generic error message - use simple text matching
    await waitFor(() => {
      expect(screen.getByText('Signup failed')).toBeInTheDocument();
    });
  });

  // ------------------------------
  // Test 7: Privacy Compliance
  // ------------------------------
  it('should not log sensitive information', async () => {
    const { log } = await import('../../lib/logger');
    
    // Mock successful API response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'success',
        user_id: 'test-user-id-123',
        next_step: 'verify_email',
      }),
    });

    renderSignupPage();

    // Complete free signup flow
    const freePlanButton = screen.getByText(/free plan/i);
    fireEvent.click(freePlanButton);

    await waitFor(() => {
      expect(screen.getByLabelText(/email address/i)).toBeInTheDocument();
    });

    const emailInput = screen.getByLabelText(/email address/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
    const fallbackEmailInput = screen.getByLabelText(/fallback email/i);

    fireEvent.change(emailInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'SecurePass123!' } });
    fireEvent.change(fallbackEmailInput, { target: { value: 'backup@example.com' } });

    // Submit form
    const form = document.querySelector('form');
    fireEvent.submit(form!);

    // Wait for review page
    await waitFor(() => {
      expect(screen.getByText(/review your account/i)).toBeInTheDocument();
    });

    // Click create account
    const createAccountButton = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(createAccountButton);

    // Verify that sensitive data is not logged
    expect(log.info).not.toHaveBeenCalledWith(
      expect.stringContaining('testuser'),
      expect.anything(),
      expect.anything()
    );
    expect(log.info).not.toHaveBeenCalledWith(
      expect.stringContaining('SecurePass123!'),
      expect.anything(),
      expect.anything()
    );
    expect(log.info).not.toHaveBeenCalledWith(
      expect.stringContaining('backup@example.com'),
      expect.anything(),
      expect.anything()
    );

    // Verify that only safe data is logged
    expect(log.info).toHaveBeenCalledWith(
      expect.stringContaining('Account type selected'),
      expect.objectContaining({ accountType: 'free' }),
      expect.anything()
    );
  });
});
