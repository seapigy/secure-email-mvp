import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { toast } from 'react-toastify';
import ComposeModal from '../../components/secure/ComposeModal';
import { sendSecureEmail } from '../../lib/api';

// Mock the API module
vi.mock('../../lib/api', () => ({
  sendSecureEmail: vi.fn(),
}));

// Mock react-toastify
vi.mock('react-toastify', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('Email Pipeline Integration', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should render compose modal with basic form fields', async () => {
    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Check that basic form elements are present
    expect(screen.getByText('Compose Secure Email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('recipient@example.com')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter subject line')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Type your secure message here...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Send secure email/i })).toBeInTheDocument();
  });

  it('should fill in basic email fields', async () => {
    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Fill in basic email fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Email' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'This is a test email' },
    });

    // Verify the values were set
    expect(screen.getByDisplayValue('test@example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Test Email')).toBeInTheDocument();
    expect(screen.getByDisplayValue('This is a test email')).toBeInTheDocument();
  });

  it('should enable password protection', async () => {
    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Enable password protection
    const passwordProtectionToggle = screen.getByLabelText('Enable password protection for this secure email');
    fireEvent.click(passwordProtectionToggle);

    // Check that password field appears
    expect(screen.getByPlaceholderText('Enter password (min. 6 characters)')).toBeInTheDocument();
  });

  it('should handle form submission with basic fields', async () => {
    // Mock successful API response
    const mockApiResponse = {
      status: 'success',
      blob_id: 'test-blob-123',
      burn_after_read: true,
      access_count: 0,
      max_attempts: 3,
    };
    
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (sendSecureEmail as any).mockResolvedValue(mockApiResponse);

    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Fill in basic email fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Email' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'This is a test email' },
    });

    // Submit the form
    const submitButton = screen.getByRole('button', { name: /Send secure email/i });
    fireEvent.click(submitButton);

    // Wait for API call
    await waitFor(() => {
      expect(sendSecureEmail).toHaveBeenCalled();
    }, { timeout: 3000 });

    // Verify success toast
    expect(toast.success).toHaveBeenCalledWith('Secure email sent successfully!');
    
    // Verify modal closes
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should handle API errors gracefully', async () => {
    // Mock API error
    const mockApiError = new Error('Network error');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (sendSecureEmail as any).mockRejectedValue(mockApiError);

    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Fill in basic email fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Email' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'This is a test email' },
    });

    // Submit the form
    const submitButton = screen.getByRole('button', { name: /Send secure email/i });
    fireEvent.click(submitButton);

    // Wait for error handling
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Failed to prepare request. Please try again.');
    }, { timeout: 3000 });

    // Verify modal doesn't close on error
    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should validate required fields before submission', async () => {
    render(
      <ComposeModal 
        isOpen={true} 
        onClose={mockOnClose} 
      />
    );

    // Try to submit without filling required fields
    const submitButton = screen.getByRole('button', { name: /Send secure email/i });
    fireEvent.click(submitButton);

    // Should not call API due to validation
    expect(sendSecureEmail).not.toHaveBeenCalled();
  });
});
