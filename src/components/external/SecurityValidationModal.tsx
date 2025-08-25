import React, { useState, useEffect } from 'react';
import { X, Lock, Shield, MapPin, AlertTriangle, CheckCircle } from 'lucide-react';

interface SecuritySettings {
  require_password: boolean;
  require_mfa: boolean;
  mfa_type?: string;
  geolocation_restriction: boolean;
  allowed_countries?: string[];
  allowed_cities?: string[];
  time_lock: boolean;
  time_lock_until?: number;
  read_once: boolean;
  auto_destruct: boolean;
  expires_at?: number;
  max_access_attempts: number;
  current_attempts: number;
}

interface SecurityValidationRequest {
  link_id: string;
  password?: string;
  mfa_code?: string;
  mfa_type?: string;
  ip_address?: string;
  user_agent?: string;
}

interface SecurityValidationModalProps {
  isOpen: boolean;
  onClose: () => void;
  onValidate: (request: SecurityValidationRequest) => Promise<any>;
  securitySettings: SecuritySettings;
  currentStep: 'password' | 'mfa' | 'geo' | 'complete';
  linkID: string;
}

const SecurityValidationModal: React.FC<SecurityValidationModalProps> = ({
  isOpen,
  onClose,
  onValidate,
  securitySettings,
  currentStep,
  linkID,
}) => {
  const [step, setStep] = useState<'password' | 'mfa' | 'geo' | 'complete'>(currentStep);
  const [password, setPassword] = useState('');
  const [mfaCode, setMfaCode] = useState('');
  const [geoLocation, setGeoLocation] = useState<{ country: string; city: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [decoyMessage, setDecoyMessage] = useState<string | null>(null);
  const [validationData, setValidationData] = useState<SecurityValidationRequest>({
    link_id: linkID,
  });

  useEffect(() => {
    setStep(currentStep);
    setError(null);
    setDecoyMessage(null);
  }, [currentStep]);

  useEffect(() => {
    if (isOpen && securitySettings.geolocation_restriction) {
      getCurrentLocation();
    }
  }, [isOpen, securitySettings.geolocation_restriction]);

  const getCurrentLocation = async () => {
    try {
      // In a real implementation, this would call a geolocation service
      // For now, we'll use a mock location
      setGeoLocation({
        country: 'United States',
        city: 'New York',
      });
    } catch (err) {
      console.error('Error getting location:', err);
      setError('Unable to determine your location');
    }
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!password.trim()) {
      setError('Password is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: SecurityValidationRequest = {
        ...validationData,
        password: password,
      };

      const result = await onValidate(request);

      if (result?.error) {
        setError(result.error);
        if (result.decoyMessage) {
          setDecoyMessage(result.decoyMessage);
        }
        return;
      }

      if (result?.success) {
        // All validation complete
        setStep('complete');
        return;
      }

      if (result?.requiresMFA) {
        setStep('mfa');
        setValidationData(request);
        return;
      }

      if (result?.requiresGeo) {
        setStep('geo');
        setValidationData(request);
        return;
      }
    } catch (err) {
      setError('Password validation failed');
      console.error('Password validation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleMFASubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!mfaCode.trim()) {
      setError('Verification code is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: SecurityValidationRequest = {
        ...validationData,
        mfa_code: mfaCode,
        mfa_type: securitySettings.mfa_type,
      };

      const result = await onValidate(request);

      if (result?.error) {
        setError(result.error);
        if (result.decoyMessage) {
          setDecoyMessage(result.decoyMessage);
        }
        return;
      }

      if (result?.success) {
        setStep('complete');
        return;
      }

      if (result?.requiresGeo) {
        setStep('geo');
        setValidationData(request);
        return;
      }
    } catch (err) {
      setError('MFA validation failed');
      console.error('MFA validation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleGeoSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!geoLocation) {
      setError('Location verification is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: SecurityValidationRequest = {
        ...validationData,
        ip_address: '127.0.0.1', // In real implementation, get from request
        user_agent: navigator.userAgent,
      };

      const result = await onValidate(request);

      if (result?.error) {
        setError(result.error);
        if (result.decoyMessage) {
          setDecoyMessage(result.decoyMessage);
        }
        return;
      }

      if (result?.success) {
        setStep('complete');
        return;
      }
    } catch (err) {
      setError('Location validation failed');
      console.error('Location validation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const renderPasswordStep = () => (
    <div className="space-y-4">
      <div className="text-center">
        <Lock className="h-12 w-12 text-blue-500 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900">Password Required</h3>
        <p className="text-sm text-gray-600 mt-2">
          This secure message is protected with a password. Please enter the password provided by the sender.
        </p>
      </div>

      <form onSubmit={handlePasswordSubmit} className="space-y-4">
        <div>
          <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
            Password
          </label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="Enter password"
            disabled={loading}
          />
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-3">
            <div className="flex">
              <AlertTriangle className="h-5 w-5 text-red-400" />
              <div className="ml-3">
                <p className="text-sm text-red-800">{error}</p>
                {decoyMessage && (
                  <p className="text-sm text-red-700 mt-1">{decoyMessage}</p>
                )}
              </div>
            </div>
          </div>
        )}

        <div className="flex justify-between items-center">
          <button
            type="button"
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 text-sm"
            disabled={loading}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading || !password.trim()}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Verifying...' : 'Continue'}
          </button>
        </div>
      </form>
    </div>
  );

  const renderMFAStep = () => (
    <div className="space-y-4">
      <div className="text-center">
        <Shield className="h-12 w-12 text-green-500 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900">Two-Factor Authentication</h3>
        <p className="text-sm text-gray-600 mt-2">
          Please enter the {securitySettings.mfa_type} verification code to access this secure message.
        </p>
      </div>

      <form onSubmit={handleMFASubmit} className="space-y-4">
        <div>
          <label htmlFor="mfaCode" className="block text-sm font-medium text-gray-700 mb-1">
            Verification Code
          </label>
          <input
            type="text"
            id="mfaCode"
            value={mfaCode}
            onChange={(e) => setMfaCode(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder={`Enter ${securitySettings.mfa_type} code`}
            disabled={loading}
            maxLength={6}
          />
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-3">
            <div className="flex">
              <AlertTriangle className="h-5 w-5 text-red-400" />
              <div className="ml-3">
                <p className="text-sm text-red-800">{error}</p>
                {decoyMessage && (
                  <p className="text-sm text-red-700 mt-1">{decoyMessage}</p>
                )}
              </div>
            </div>
          </div>
        )}

        <div className="flex justify-between items-center">
          <button
            type="button"
            onClick={() => setStep('password')}
            className="text-gray-500 hover:text-gray-700 text-sm"
            disabled={loading}
          >
            Back
          </button>
          <button
            type="submit"
            disabled={loading || !mfaCode.trim()}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Verifying...' : 'Continue'}
          </button>
        </div>
      </form>
    </div>
  );

  const renderGeoStep = () => (
    <div className="space-y-4">
      <div className="text-center">
        <MapPin className="h-12 w-12 text-purple-500 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900">Location Verification</h3>
        <p className="text-sm text-gray-600 mt-2">
          This secure message is restricted to specific locations. Please confirm your current location.
        </p>
      </div>

      <div className="bg-gray-50 rounded-md p-4">
        <div className="space-y-2">
          <div className="flex justify-between">
            <span className="text-sm text-gray-600">Current Location:</span>
            <span className="text-sm font-medium text-gray-900">
              {geoLocation ? `${geoLocation.city}, ${geoLocation.country}` : 'Detecting...'}
            </span>
          </div>
          {securitySettings.allowed_countries && securitySettings.allowed_countries.length > 0 && (
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Allowed Countries:</span>
              <span className="text-sm font-medium text-gray-900">
                {securitySettings.allowed_countries.join(', ')}
              </span>
            </div>
          )}
          {securitySettings.allowed_cities && securitySettings.allowed_cities.length > 0 && (
            <div className="flex justify-between">
              <span className="text-sm text-gray-600">Allowed Cities:</span>
              <span className="text-sm font-medium text-gray-900">
                {securitySettings.allowed_cities.join(', ')}
              </span>
            </div>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-3">
          <div className="flex">
            <AlertTriangle className="h-5 w-5 text-red-400" />
            <div className="ml-3">
              <p className="text-sm text-red-800">{error}</p>
              {decoyMessage && (
                <p className="text-sm text-red-700 mt-1">{decoyMessage}</p>
              )}
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-between items-center">
        <button
          type="button"
          onClick={() => setStep(securitySettings.require_mfa ? 'mfa' : 'password')}
          className="text-gray-500 hover:text-gray-700 text-sm"
          disabled={loading}
        >
          Back
        </button>
        <button
          onClick={handleGeoSubmit}
          disabled={loading || !geoLocation}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Verifying...' : 'Confirm Location'}
        </button>
      </div>
    </div>
  );

  const renderCompleteStep = () => (
    <div className="text-center space-y-4">
      <CheckCircle className="h-12 w-12 text-green-500 mx-auto" />
      <h3 className="text-lg font-medium text-gray-900">Access Granted</h3>
      <p className="text-sm text-gray-600">
        All security checks have been completed successfully. Loading your secure message...
      </p>
    </div>
  );

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Secure Access</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            disabled={loading}
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {step === 'password' && renderPasswordStep()}
          {step === 'mfa' && renderMFAStep()}
          {step === 'geo' && renderGeoStep()}
          {step === 'complete' && renderCompleteStep()}
        </div>

        {/* Footer */}
        <div className="bg-gray-50 px-6 py-4 rounded-b-lg">
          <div className="flex items-center justify-between text-sm text-gray-600">
            <span>Access attempts: {securitySettings.current_attempts} / {securitySettings.max_access_attempts}</span>
            {securitySettings.auto_destruct && (
              <span className="text-red-600 flex items-center">
                <AlertTriangle className="h-4 w-4 mr-1" />
                Auto-destruct enabled
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecurityValidationModal;
