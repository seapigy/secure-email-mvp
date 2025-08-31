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

describe('SecureEmailViewer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
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

    renderWithRouter(<SecureEmailViewer />);
    
    expect(screen.getByText(/Secure Email System/i)).toBeInTheDocument();
  });

  it('shows loading state initially', () => {
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

    renderWithRouter(<SecureEmailViewer />);
    
    expect(screen.getByText(/Verifying Access/i)).toBeInTheDocument();
  });

  it('displays error when link is not found', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ message: 'Link not found' }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      expect(screen.getByText(/Link not found/i)).toBeInTheDocument();
    });
  });

  it('shows security modal when password is required', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: true,
          require_mfa: false,
          geolocation_restriction: false,
        },
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      expect(screen.getByText(/Password Required/i)).toBeInTheDocument();
    });
  });

  it('shows MFA prompt when MFA is required', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        status: 'active',
        security_settings: {
          require_password: false,
          require_mfa: true,
          mfa_type: 'totp',
          geolocation_restriction: false,
        },
      }),
    } as Response);

    renderWithRouter(<SecureEmailViewer />);
    
    await waitFor(() => {
      // The component should show the security modal with MFA step
      expect(screen.getByText(/Secure Access/i)).toBeInTheDocument();
    });
  });
});
