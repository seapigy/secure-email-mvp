import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { toast } from 'react-toastify';
import ComposeModal from './ComposeModal';

// Mock dependencies
jest.mock('@/lib/api', () => ({
  sendSecureEmail: jest.fn(),
  isApiError: jest.fn(),
  getErrorMessage: jest.fn(),
}));

jest.mock('react-toastify', () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock('@/lib/geolocation', () => ({
  validateCountryCode: jest.fn(),
  validateCityName: jest.fn(),
  SUPPORTED_COUNTRIES: [
    { code: 'US', name: 'United States' },
    { code: 'CA', name: 'Canada' },
    { code: 'GB', name: 'United Kingdom' },
  ],
  GeoVerificationType: jest.fn(),
}));

const mockProps = {
  isOpen: true,
  onClose: jest.fn(),
};

describe('ComposeModal - Enhanced Geolocation Verification', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render geolocation verification section', () => {
    render(<ComposeModal {...mockProps} />);
    
    expect(screen.getByText('Enhanced Geolocation Verification')).toBeInTheDocument();
    expect(screen.getByText('Verification Type')).toBeInTheDocument();
    expect(screen.getByDisplayValue('None - No location restrictions')).toBeInTheDocument();
  });

  it('should show country dropdown when country verification is selected', async () => {
    render(<ComposeModal {...mockProps} />);
    
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'country' } });
    
    await waitFor(() => {
      expect(screen.getByText('Required Country')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Select a country (required)')).toBeInTheDocument();
    });
  });

  it('should show city input when city verification is selected', async () => {
    render(<ComposeModal {...mockProps} />);
    
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'city' } });
    
    await waitFor(() => {
      expect(screen.getByText('Required City')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Enter city name (e.g., New York, London, Tokyo)')).toBeInTheDocument();
    });
  });

  it('should show both city and country inputs when city_country verification is selected', async () => {
    render(<ComposeModal {...mockProps} />);
    
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'city_country' } });
    
    await waitFor(() => {
      expect(screen.getByText('Required City')).toBeInTheDocument();
      expect(screen.getByText('Required Country')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Enter city name (e.g., New York, London, Tokyo)')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Select a country (required)')).toBeInTheDocument();
    });
  });

  it('should show verification info when geolocation verification is enabled', async () => {
    render(<ComposeModal {...mockProps} />);
    
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'country' } });
    
    await waitFor(() => {
      expect(screen.getByText('Geolocation Verification')).toBeInTheDocument();
      expect(screen.getByText('Recipients must be in the specified country.')).toBeInTheDocument();
      expect(screen.getByText("Location is determined by the recipient's IP address.")).toBeInTheDocument();
    });
  });

  it('should disable submit button when required geolocation fields are missing', async () => {
    render(<ComposeModal {...mockProps} />);
    
    // Fill required fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Subject' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'Test message body' },
    });
    
    // Select country verification but don't select country
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'country' } });
    
    await waitFor(() => {
      const submitButton = screen.getByText('Send Securely');
      expect(submitButton).toBeDisabled();
    });
  });

  it('should enable submit button when all required geolocation fields are filled', async () => {
    render(<ComposeModal {...mockProps} />);
    
    // Fill required fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Subject' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'Test message body' },
    });
    
    // Select country verification and select country
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'country' } });
    
    await waitFor(() => {
      const countrySelect = screen.getByDisplayValue('Select a country (required)');
      fireEvent.change(countrySelect, { target: { value: 'US' } });
    });
    
    await waitFor(() => {
      const submitButton = screen.getByText('Send Securely');
      expect(submitButton).not.toBeDisabled();
    });
  });

  it('should include geolocation data in API request', async () => {
    const { sendSecureEmail } = require('@/lib/api');
    sendSecureEmail.mockResolvedValue({ status: 'success', blob_id: 'test-blob-id' });
    
    render(<ComposeModal {...mockProps} />);
    
    // Fill required fields
    fireEvent.change(screen.getByPlaceholderText('recipient@example.com'), {
      target: { value: 'test@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('Enter subject line'), {
      target: { value: 'Test Subject' },
    });
    fireEvent.change(screen.getByPlaceholderText('Type your secure message here...'), {
      target: { value: 'Test message body' },
    });
    
    // Select city_country verification and fill fields
    const verificationSelect = screen.getByDisplayValue('None - No location restrictions');
    fireEvent.change(verificationSelect, { target: { value: 'city_country' } });
    
    await waitFor(() => {
      const cityInput = screen.getByPlaceholderText('Enter city name (e.g., New York, London, Tokyo)');
      const countrySelect = screen.getByDisplayValue('Select a country (required)');
      
      fireEvent.change(cityInput, { target: { value: 'New York' } });
      fireEvent.change(countrySelect, { target: { value: 'US' } });
    });
    
    // Submit form
    const submitButton = screen.getByText('Send Securely');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(sendSecureEmail).toHaveBeenCalledWith(
        expect.objectContaining({
          geoVerificationType: 'city_country',
          geoCity: 'New York',
          geoCountry: 'US',
        })
      );
    });
  });
});
