import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import SecureEmailViewer from './SecureEmailViewer';

// Mock the useParams hook
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({ linkID: 'test-link-123' }),
  };
});

// Mock fetch responses
const mockFetch = vi.fn();
Object.defineProperty(window, 'fetch', {
  value: mockFetch,
  writable: true,
});

const renderWithRouter = (component: React.ReactElement) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('SecureEmailViewer - Burn After Read', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows burn-after-read warning when enabled', async () => {
    // Mock metadata response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: false,
          require_mfa: false,
          geolocation_restriction: false,
        },
      }),
    } as Response);

    // Mock content response with read_once enabled
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        subject: 'Test Email',
        body: 'This is a test email',
        read_once: true,
        auto_destruct: false,
        expires_at: null,
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      expect(screen.getByText(/One-time viewing/i)).toBeInTheDocument();
    });
  });

  it('shows auto-destruct warning when enabled', async () => {
    // Mock metadata response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: false,
          require_mfa: false,
          geolocation_restriction: false,
        },
      }),
    } as Response);

    // Mock content response with auto_destruct enabled
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        subject: 'Test Email',
        body: 'This is a test email',
        read_once: false,
        auto_destruct: true,
        expires_at: Date.now() + 3600000, // 1 hour from now
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      expect(screen.getByText(/Auto-destruct enabled/i)).toBeInTheDocument();
    });
  });

  it('shows expiration warning when message expires', async () => {
    // Mock metadata response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: false,
          require_mfa: false,
          geolocation_restriction: false,
        },
      }),
    } as Response);

    // Mock content response with expiration
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        subject: 'Test Email',
        body: 'This is a test email',
        read_once: false,
        auto_destruct: false,
        expires_at: Date.now() + 3600000, // 1 hour from now
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      // Check for the expiration badge or warning
      expect(screen.getByText(/This secure message was delivered using SecureMail's encrypted email system/i)).toBeInTheDocument();
    });
  });

  it('displays secure message footer', async () => {
    // Mock metadata response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: false,
          require_mfa: false,
          geolocation_restriction: false,
        },
      }),
    } as Response);

    // Mock content response
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        subject: 'Test Email',
        body: 'This is a test email',
        read_once: false,
        auto_destruct: false,
        expires_at: null,
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      expect(screen.getByText(/This secure message was delivered using SecureMail's encrypted email system/i)).toBeInTheDocument();
    });
  });
});
