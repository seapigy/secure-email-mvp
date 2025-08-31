import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import AdminDashboard from '../../pages/admin/Dashboard';

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
    get: vi.fn(),
    put: vi.fn(),
  },
}));

// Mock react-toastify
vi.mock('react-toastify', () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

// Mock localStorage
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage,
});

const renderWithRouter = (component: React.ReactElement) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('AdminDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock authenticated admin user
    mockLocalStorage.getItem.mockImplementation((key: string) => {
      if (key === 'admin_token') return 'test-token';
      if (key === 'admin_user') return JSON.stringify({
        id: 1,
        email: 'admin@test.com',
        role: 'admin',
      });
      return null;
    });
  });

  it('renders dashboard with admin user info', () => {
    renderWithRouter(<AdminDashboard />);
    
    expect(screen.getByText('Admin Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Welcome, admin@test.com (admin)')).toBeInTheDocument();
    expect(screen.getByText('Logout')).toBeInTheDocument();
  });

  it('renders navigation tabs', () => {
    renderWithRouter(<AdminDashboard />);
    
    expect(screen.getByText('Security Policies')).toBeInTheDocument();
    expect(screen.getByText('DLP Logs')).toBeInTheDocument();
    expect(screen.getByText('User Management')).toBeInTheDocument();
  });

  it('shows security policies tab by default', async () => {
    // Mock policies response
    const mockPoliciesResponse = {
      data: {
        success: true,
        policies: [
          {
            id: '1',
            name: 'Password Policy',
            type: 'password',
            category: 'authentication',
            enabled: true,
            severity: 'high',
            enforcement_level: 'strict',
          },
        ],
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.get).mockResolvedValueOnce(mockPoliciesResponse);

    renderWithRouter(<AdminDashboard />);

    // Wait for the content to load
    await waitFor(() => {
      expect(screen.getByText('Manage system-wide security policies')).toBeInTheDocument();
    });
  });

  it('switches to DLP logs tab when clicked', async () => {
    const mockDLPResponse = {
      data: {
        success: true,
        logs: [
          {
            id: '1',
            timestamp: '2024-01-15T10:30:00Z',
            email: 'user@example.com',
            scan_type: 'content_scan',
            findings: ['credit_card', 'ssn'],
            action: 'blocked',
            admin: 'admin@test.com',
          },
        ],
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.get).mockResolvedValueOnce(mockDLPResponse);

    renderWithRouter(<AdminDashboard />);

    // Click on DLP Logs tab button (use getAllByText to get the button specifically)
    const dlpButtons = screen.getAllByText('DLP Logs');
    const dlpTabButton = dlpButtons.find(button => button.tagName === 'BUTTON');
    fireEvent.click(dlpTabButton!);

    await waitFor(() => {
      expect(screen.getByText('Data Loss Prevention scan results and actions')).toBeInTheDocument();
    });

    // Should have made API call to load DLP logs
    expect(axios.default.get).toHaveBeenCalledWith('/api/admin/dlp/logs', expect.any(Object));
  });

  it('switches to user management tab when clicked', async () => {
    renderWithRouter(<AdminDashboard />);

    // Click on User Management tab button
    const userManagementButtons = screen.getAllByText('User Management');
    const userManagementTabButton = userManagementButtons.find(button => button.tagName === 'BUTTON');
    fireEvent.click(userManagementTabButton!);

    await waitFor(() => {
      expect(screen.getByText('User management functionality coming soon...')).toBeInTheDocument();
    });
  });

  it('handles logout', () => {
    renderWithRouter(<AdminDashboard />);

    // Click logout button
    fireEvent.click(screen.getByText('Logout'));

    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('admin_token');
    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('admin_user');
    expect(mockNavigate).toHaveBeenCalledWith('/admin/login');
  });

  it('redirects to login if no token', () => {
    // Mock no token
    mockLocalStorage.getItem.mockReturnValue(null);

    renderWithRouter(<AdminDashboard />);

    expect(mockNavigate).toHaveBeenCalledWith('/admin/login');
  });

  it('loads security policies on mount', async () => {
    const mockPoliciesResponse = {
      data: {
        success: true,
        policies: [
          {
            id: '1',
            name: 'Password Policy',
            type: 'password',
            category: 'authentication',
            enabled: true,
            severity: 'high',
            enforcement_level: 'strict',
          },
        ],
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.get).mockResolvedValueOnce(mockPoliciesResponse);

    renderWithRouter(<AdminDashboard />);

    await waitFor(() => {
      expect(axios.default.get).toHaveBeenCalledWith('/api/security/policies', expect.any(Object));
    });
  });

  it('handles security policy update', async () => {
    const mockUpdateResponse = {
      data: {
        success: true,
        message: 'Policy updated successfully',
      },
    };
    const axios = await import('axios');
    vi.mocked(axios.default.put).mockResolvedValueOnce(mockUpdateResponse);

    renderWithRouter(<AdminDashboard />);

    // Wait for policies to load
    await waitFor(() => {
      expect(axios.default.get).toHaveBeenCalledWith('/api/security/policies', expect.any(Object));
    });

    // Mock policies data
    const mockPoliciesResponse = {
      data: {
        success: true,
        policies: [
          {
            id: '1',
            name: 'Password Policy',
            type: 'password',
            category: 'authentication',
            enabled: true,
            severity: 'high',
            enforcement_level: 'strict',
          },
        ],
      },
    };
    vi.mocked(axios.default.get).mockResolvedValueOnce(mockPoliciesResponse);

    // This would test the actual update functionality
    // For now, just verify the component renders correctly
    await waitFor(() => {
      const securityPolicyHeaders = screen.getAllByText('Security Policies');
      expect(securityPolicyHeaders.length).toBeGreaterThan(0);
    });
  });

  it('shows loading state', async () => {
    // Mock a slow response
    const axios = await import('axios');
    vi.mocked(axios.default.get).mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    renderWithRouter(<AdminDashboard />);

    // Should show loading state initially
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });
});
