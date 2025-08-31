import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import SecurityValidationModal from './SecurityValidationModal';

const mockOnValidate = vi.fn();
const mockOnClose = vi.fn();

const defaultProps = {
  isOpen: true,
  onClose: mockOnClose,
  onValidate: mockOnValidate,
  securitySettings: {
    require_password: false,
    require_mfa: false,
    geolocation_restriction: false,
    time_lock: false,
    read_once: false,
    auto_destruct: false,
    max_access_attempts: 3,
    current_attempts: 0,
  },
  currentStep: 'password' as const,
  linkID: 'test-link-123',
};

describe('SecurityValidationModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders password step when password is required', () => {
    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_password: true,
      },
      currentStep: 'password' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    expect(screen.getByText(/Password Required/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Continue/i })).toBeInTheDocument();
  });

  it('renders MFA step when MFA is required', () => {
    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_mfa: true,
        mfa_type: 'totp',
      },
      currentStep: 'mfa' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    expect(screen.getByText(/Two-Factor Authentication/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Verification Code/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Continue/i })).toBeInTheDocument();
  });

  it('renders geolocation step when geolocation restriction is enabled', () => {
    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        geolocation_restriction: true,
        allowed_countries: ['United States'],
      },
      currentStep: 'geo' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    expect(screen.getByText(/Location Verification/i)).toBeInTheDocument();
    expect(screen.getByText(/confirm your current location/i)).toBeInTheDocument();
  });

  it('shows completion step when all validations pass', () => {
    const props = {
      ...defaultProps,
      currentStep: 'complete' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    expect(screen.getByText(/Access Granted/i)).toBeInTheDocument();
    expect(screen.getByText(/All security checks have been completed successfully/i)).toBeInTheDocument();
  });

  it('handles password submission correctly', async () => {
    mockOnValidate.mockResolvedValueOnce({ success: true });

    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_password: true,
      },
      currentStep: 'password' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /Continue/i });

    fireEvent.change(passwordInput, { target: { value: 'testpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnValidate).toHaveBeenCalledWith({
        link_id: 'test-link-123',
        password: 'testpassword',
      });
    });
  });

  it('handles MFA submission correctly', async () => {
    mockOnValidate.mockResolvedValueOnce({ success: true });

    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_mfa: true,
        mfa_type: 'totp',
      },
      currentStep: 'mfa' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    const mfaInput = screen.getByLabelText(/Verification Code/i);
    const verifyButton = screen.getByRole('button', { name: /Continue/i });

    fireEvent.change(mfaInput, { target: { value: '123456' } });
    fireEvent.click(verifyButton);

    await waitFor(() => {
      expect(mockOnValidate).toHaveBeenCalledWith({
        link_id: 'test-link-123',
        mfa_code: '123456',
        mfa_type: 'totp',
      });
    });
  });

  it('shows error message when validation fails', async () => {
    mockOnValidate.mockResolvedValueOnce({ 
      error: 'Invalid password',
      decoyMessage: 'This is a decoy message'
    });

    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_password: true,
      },
      currentStep: 'password' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /Continue/i });

    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/Invalid password/i)).toBeInTheDocument();
      expect(screen.getByText(/This is a decoy message/i)).toBeInTheDocument();
    });
  });

  it('transitions to MFA step when password is correct but MFA is required', async () => {
    mockOnValidate.mockResolvedValueOnce({ 
      requiresMFA: true,
      mfaType: 'totp'
    });

    const props = {
      ...defaultProps,
      securitySettings: {
        ...defaultProps.securitySettings,
        require_password: true,
        require_mfa: true,
        mfa_type: 'totp',
      },
      currentStep: 'password' as const,
    };

    render(<SecurityValidationModal {...props} />);
    
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /Continue/i });

    fireEvent.change(passwordInput, { target: { value: 'correctpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/Two-Factor Authentication/i)).toBeInTheDocument();
    });
  });

  it('closes modal when close button is clicked', () => {
    render(<SecurityValidationModal {...defaultProps} />);
    
    const closeButton = screen.getByRole('button', { name: '' }); // Close button has no text, just an X icon
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalled();
  });
});
