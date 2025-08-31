import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import '@testing-library/jest-dom';
import AuditLogTable from './AuditLogTable';
import { fetchAuditLogs } from '../../lib/adminApi';

// Mock the adminApi module
vi.mock('../../lib/adminApi', () => ({
  fetchAuditLogs: vi.fn(),
}));

const mockFetchAuditLogs = vi.mocked(fetchAuditLogs);

describe('AuditLogTable', () => {
  const mockAuditLogs = [
    {
      id: 1,
      timestamp: '2024-01-15T10:30:00Z',
      user_id: 'admin@test.com',
      action: 'login',
      entity: 'admin',
      details: 'Admin login successful',
      severity: 'low',
    },
    {
      id: 2,
      timestamp: '2024-01-15T11:15:00Z',
      user_id: 'user@example.com',
      action: 'dlp_scan',
      entity: 'secure_link',
      details: 'Detected sensitive data',
      severity: 'high',
    },
    {
      id: 3,
      timestamp: '2024-01-15T12:00:00Z',
      user_id: 'admin@test.com',
      action: 'update_policy',
      entity: 'system_security_policy',
      details: 'Updated password policy',
      severity: 'high',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders audit logs table with data', async () => {
    mockFetchAuditLogs.mockResolvedValueOnce({
      success: true,
      logs: mockAuditLogs,
      total: 3,
      page: 1,
      limit: 20,
      filters: {},
    });

    render(<AuditLogTable />);

    // Wait for data to load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs
      expect(screen.getByText('user@example.com')).toBeInTheDocument();
    });

    // Check that all logs are displayed
    expect(screen.getByText('login')).toBeInTheDocument();
    expect(screen.getByText('dlp_scan')).toBeInTheDocument();
    expect(screen.getByText('update_policy')).toBeInTheDocument();

    // Check severity badges - use getAllByText since there are multiple
    const lowSeverityElements = screen.getAllByText('low');
    const highSeverityElements = screen.getAllByText('high');
    expect(lowSeverityElements).toHaveLength(1);
    expect(highSeverityElements).toHaveLength(2);

    // Check table headers - use getAllByText since they appear in both filter labels and table headers
    const timestampElements = screen.getAllByText('Timestamp');
    const userElements = screen.getAllByText('User');
    const actionElements = screen.getAllByText('Action');
    const entityElements = screen.getAllByText('Entity');
    const detailsElements = screen.getAllByText('Details');
    const severityElements = screen.getAllByText('Severity');
    
    expect(timestampElements).toHaveLength(1); // Only in table header
    expect(userElements).toHaveLength(1); // Only in table header
    expect(actionElements).toHaveLength(2); // In filter label and table header
    expect(entityElements).toHaveLength(2); // In filter label and table header
    expect(detailsElements).toHaveLength(1); // Only in table header
    expect(severityElements).toHaveLength(2); // In filter label and table header
  });

  it('shows loading state initially', () => {
    // Don't mock anything to see loading state
    render(<AuditLogTable />);
    expect(screen.getByText('Loading audit logs...')).toBeInTheDocument();
  });

  it('shows error state when API call fails', async () => {
    mockFetchAuditLogs.mockRejectedValueOnce(new Error('API Error'));

    render(<AuditLogTable />);

    // Wait for error state to be rendered
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByText('Error: API Error')).toBeInTheDocument();
      expect(screen.getByText('Retry')).toBeInTheDocument();
    }, { timeout: 10000 });
  });

  it('filters by user ID', async () => {
    // Mock initial load and filtered response
    mockFetchAuditLogs
      .mockResolvedValueOnce({
        success: true,
        logs: mockAuditLogs,
        total: 3,
        page: 1,
        limit: 20,
        filters: {},
      })
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[0]], // Only admin@test.com
        total: 1,
        page: 1,
        limit: 20,
        filters: { user_id: 'admin@test.com' },
      });

    render(<AuditLogTable />);

    // Wait for initial load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs initially
    });

    // Enter filter value
    const userFilter = screen.getByPlaceholderText('Filter by user ID...');
    fireEvent.change(userFilter, { target: { value: 'admin@test.com' } });

    // Wait for filtered results
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(1); // Should have 1 admin log after filtering
      expect(screen.queryByText('user@example.com')).not.toBeInTheDocument();
    });
  });

  it('filters by severity', async () => {
    // Mock initial load and filtered response
    mockFetchAuditLogs
      .mockResolvedValueOnce({
        success: true,
        logs: mockAuditLogs,
        total: 3,
        page: 1,
        limit: 20,
        filters: {},
      })
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[1], mockAuditLogs[2]], // Only high severity
        total: 2,
        page: 1,
        limit: 20,
        filters: { severity: 'high' },
      });

    render(<AuditLogTable />);

    // Wait for initial load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs initially
    });

    // Select severity filter
    const severityFilter = screen.getByDisplayValue('All Severities');
    fireEvent.change(severityFilter, { target: { value: 'high' } });

    // Wait for filtered results
    await waitFor(() => {
      expect(screen.getByText('user@example.com')).toBeInTheDocument();
      expect(screen.queryByText('low')).not.toBeInTheDocument();
    });
  });

  it('filters by action', async () => {
    // Mock initial load and filtered response
    mockFetchAuditLogs
      .mockResolvedValueOnce({
        success: true,
        logs: mockAuditLogs,
        total: 3,
        page: 1,
        limit: 20,
        filters: {},
      })
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[0]], // Only login action
        total: 1,
        page: 1,
        limit: 20,
        filters: { action: 'login' },
      });

    render(<AuditLogTable />);

    // Wait for initial load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs initially
    });

    // Enter action filter
    const actionFilter = screen.getByPlaceholderText('Filter by action...');
    fireEvent.change(actionFilter, { target: { value: 'login' } });

    // Wait for filtered results
    await waitFor(() => {
      expect(screen.getByText('login')).toBeInTheDocument();
      expect(screen.queryByText('dlp_scan')).not.toBeInTheDocument();
    });
  });

  it('filters by entity', async () => {
    // Mock initial load and filtered response
    mockFetchAuditLogs
      .mockResolvedValueOnce({
        success: true,
        logs: mockAuditLogs,
        total: 3,
        page: 1,
        limit: 20,
        filters: {},
      })
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[0]], // Only admin entity
        total: 1,
        page: 1,
        limit: 20,
        filters: { entity: 'admin' },
      });

    render(<AuditLogTable />);

    // Wait for initial load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs initially
    });

    // Enter entity filter
    const entityFilter = screen.getByPlaceholderText('Filter by entity...');
    fireEvent.change(entityFilter, { target: { value: 'admin' } });

    // Wait for filtered results
    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
      expect(screen.queryByText('secure_link')).not.toBeInTheDocument();
    });
  });

  it('handles pagination', async () => {
    // Mock initial load and second page - use limit: 20 to match component's pageSize
    mockFetchAuditLogs
      .mockResolvedValueOnce({
        success: true,
        logs: mockAuditLogs,
        total: 25, // More than pageSize to trigger pagination
        page: 1,
        limit: 20,
        filters: {},
      })
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[0]], // Second page data
        total: 25,
        page: 2,
        limit: 20,
        filters: {},
      });

    render(<AuditLogTable />);

    // Wait for initial load
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs initially
    });

    // Check pagination info - use regex to match text that spans multiple elements
    expect(screen.getByText((content, _element) => content.includes('Page 1 of 2'))).toBeInTheDocument();

    // Click next page
    const nextButton = screen.getByText('Next');
    fireEvent.click(nextButton);

    // Wait for page change
    await waitFor(() => {
      expect(screen.getByText((content, _element) => content.includes('Page 2 of 2'))).toBeInTheDocument();
    });
  });

  it('displays results count correctly', async () => {
    mockFetchAuditLogs.mockResolvedValueOnce({
      success: true,
      logs: mockAuditLogs,
      total: 3,
      page: 1,
      limit: 20,
      filters: {},
    });

    render(<AuditLogTable />);

    await waitFor(() => {
      // Use regex to match text that spans multiple elements
      expect(screen.getByText((content, _element) => content.includes('Showing 3 of 3 audit logs'))).toBeInTheDocument();
    });
  });

  it('formats timestamps correctly', async () => {
    mockFetchAuditLogs.mockResolvedValueOnce({
      success: true,
      logs: mockAuditLogs,
      total: 3,
      page: 1,
      limit: 20,
      filters: {},
    });

    render(<AuditLogTable />);

    await waitFor(() => {
      // Check that timestamps are formatted (should contain date/time)
      const timestamps = screen.getAllByText(/1\/15\/2024/);
      expect(timestamps).toHaveLength(3); // Should have 3 timestamps
      expect(timestamps[0]).toBeInTheDocument();
    });
  });

  it('applies correct severity colors', async () => {
    mockFetchAuditLogs.mockResolvedValueOnce({
      success: true,
      logs: mockAuditLogs,
      total: 3,
      page: 1,
      limit: 20,
      filters: {},
    });

    render(<AuditLogTable />);

    // Wait for all logs to be rendered
    await waitFor(() => {
      const adminElements = screen.getAllByText('admin@test.com');
      expect(adminElements).toHaveLength(2); // Should have 2 admin logs
      expect(screen.getByText('user@example.com')).toBeInTheDocument();
      expect(screen.getByText('login')).toBeInTheDocument();
      expect(screen.getByText('dlp_scan')).toBeInTheDocument();
      expect(screen.getByText('update_policy')).toBeInTheDocument();
    });

    // Check that severity badges have the correct classes
    const lowSeverity = screen.getByText('low');
    const highSeverityElements = screen.getAllByText('high');
    expect(highSeverityElements).toHaveLength(2); // Should have 2 high severity logs

    expect(lowSeverity).toHaveClass('bg-green-100', 'text-green-800');
    expect(highSeverityElements[0]).toHaveClass('bg-orange-100', 'text-orange-800');
    expect(highSeverityElements[1]).toHaveClass('bg-orange-100', 'text-orange-800');
  });

  it('handles retry on error', async () => {
    // Mock error and then successful retry
    mockFetchAuditLogs
      .mockRejectedValueOnce(new Error('API Error'))
      .mockResolvedValueOnce({
        success: true,
        logs: [mockAuditLogs[0]], // Only one log to avoid multiple elements
        total: 1,
        page: 1,
        limit: 20,
        filters: {},
      });

    render(<AuditLogTable />);

    // Wait for error state to be rendered
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
      expect(screen.getByText('Error: API Error')).toBeInTheDocument();
      expect(screen.getByText('Retry')).toBeInTheDocument();
    }, { timeout: 10000 });

    // Click retry button
    const retryButton = screen.getByText('Retry');
    fireEvent.click(retryButton);

    // Wait for successful retry - check for the specific log content
    await waitFor(() => {
      expect(screen.getByText('login')).toBeInTheDocument();
      expect(screen.getByText('admin@test.com')).toBeInTheDocument();
    }, { timeout: 10000 });
  });
});
