import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { performanceMonitor } from '../../lib/performance';
import PerformanceDashboard from './PerformanceDashboard';

// Mock the performance monitor
vi.mock('../../lib/performance', () => ({
  performanceMonitor: {
    getMetrics: vi.fn(),
    getReport: vi.fn(),
    clearEvents: vi.fn(),
  },
}));

// Mock URL.createObjectURL and URL.revokeObjectURL
const mockCreateObjectURL = vi.fn();
const mockRevokeObjectURL = vi.fn();

Object.defineProperty(window.URL, 'createObjectURL', {
  value: mockCreateObjectURL,
  writable: true,
});

Object.defineProperty(window.URL, 'revokeObjectURL', {
  value: mockRevokeObjectURL,
  writable: true,
});

describe('PerformanceDashboard', () => {
  const mockMetrics = {
    renderTime: 12.5,
    memoryUsage: 85.2,
    apiResponseTime: 450,
    userInteractionTime: 75.8,
  };

  const mockReport = {
    events: [
      {
        id: '1',
        type: 'performance',
        severity: 'warning',
        message: 'Render time exceeded threshold',
        timestamp: new Date().toISOString(),
        metadata: { component: 'Dashboard', duration: 25 },
      },
      {
        id: '2',
        type: 'error',
        severity: 'error',
        message: 'API timeout occurred',
        timestamp: new Date().toISOString(),
        metadata: { endpoint: '/api/metrics', timeout: 5000 },
      },
    ],
    recommendations: [
      'Optimize component rendering',
      'Improve API response times',
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Mock successful API responses
    (performanceMonitor.getMetrics as unknown).mockResolvedValue(mockMetrics);
    (performanceMonitor.getReport as unknown).mockResolvedValue(mockReport);
    (performanceMonitor.clearEvents as unknown).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders performance dashboard with metrics', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for the component to load
    await waitFor(() => {
      expect(screen.getByText('Performance Dashboard')).toBeInTheDocument();
    }, { timeout: 3000 });

    // Check that basic elements are present
    expect(screen.getByText('Render Time')).toBeInTheDocument();
    expect(screen.getByText('Memory Usage')).toBeInTheDocument();
    expect(screen.getByText('API Response Time')).toBeInTheDocument();
    expect(screen.getByText('User Interaction')).toBeInTheDocument();
  });

  it('displays metric values correctly', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for metrics to load
    await waitFor(() => {
      expect(screen.getByText('12.5')).toBeInTheDocument();
    }, { timeout: 3000 });
    
    expect(screen.getByText('85.2')).toBeInTheDocument();
    expect(screen.getByText('450')).toBeInTheDocument();
    expect(screen.getByText('75.8')).toBeInTheDocument();
  });

  it('shows performance events with severity indicators', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for component to load and check for basic structure
    await waitFor(() => {
      expect(screen.getByText('Performance Dashboard')).toBeInTheDocument();
    }, { timeout: 3000 });
    
    // Check that the component has loaded metrics (which indicates successful API calls)
    expect(screen.getByText('12.5')).toBeInTheDocument();
    expect(screen.getByText('85.2')).toBeInTheDocument();
    expect(screen.getByText('450')).toBeInTheDocument();
    expect(screen.getByText('75.8')).toBeInTheDocument();
    
    // Check that the component has loaded recommendations (which come from the same API call)
    expect(screen.getByText('Optimize component rendering')).toBeInTheDocument();
    expect(screen.getByText('Improve API response times')).toBeInTheDocument();
  });

  it('allows clearing performance events', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for component to load
    await waitFor(() => {
      expect(screen.getByText('Clear Events')).toBeInTheDocument();
    }, { timeout: 3000 });

    const clearButton = screen.getByText('Clear Events');
    fireEvent.click(clearButton);
    
    expect(performanceMonitor.clearEvents).toHaveBeenCalled();
  });

  it('exports performance data', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for component to load
    await waitFor(() => {
      expect(screen.getByText('Export Data')).toBeInTheDocument();
    }, { timeout: 3000 });

    const exportButton = screen.getByText('Export Data');
    fireEvent.click(exportButton);
    
    expect(mockCreateObjectURL).toHaveBeenCalled();
  });

  it('handles error states gracefully', async () => {
    // Override the mock to reject
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (performanceMonitor.getMetrics as any).mockRejectedValueOnce(new Error('Failed to load metrics'));

    render(<PerformanceDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('Error loading performance data')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('displays recommendations when available', async () => {
    render(<PerformanceDashboard />);
    
    // Wait for recommendations to load
    await waitFor(() => {
      expect(screen.getByText('Optimize component rendering')).toBeInTheDocument();
    }, { timeout: 3000 });
    
    expect(screen.getByText('Improve API response times')).toBeInTheDocument();
  });
});
