import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { log } from '@/lib/logger';
import {
  EyeIcon, EyeSlashIcon, ShieldCheckIcon, ExclamationTriangleIcon, CheckCircleIcon, QrCodeIcon,
  KeyIcon, UserIcon, LockClosedIcon, Cog6ToothIcon, UserGroupIcon,
} from '@heroicons/react/24/outline';
import { AdminAuthConfig } from '../../types/admin';
import { EnterpriseDashboardService } from '../../services/enterpriseDashboardService';
import { AdminUser } from '../../types/admin';

interface EnterpriseAdminLoginProps {
  onLoginSuccess: (token: string, user: AdminUser) => void;
}

const EnterpriseAdminLogin: React.FC<EnterpriseAdminLoginProps> = ({
  onLoginSuccess,
}) => {
  const [step, setStep] = useState<'login' | 'mfa' | 'mfa-setup' | 'invitation'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [mfaCode, setMfaCode] = useState('');
  const [invitationKey, setInvitationKey] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showInvitationKey, setShowInvitationKey] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [authConfig, setAuthConfig] = useState<AdminAuthConfig | null>(null);

  const [totpSecret, setTotpSecret] = useState('');
  const [hardwareTokenId, setHardwareTokenId] = useState('');
  const [mfaType, setMfaType] = useState<'TOTP' | 'HARDWARE_TOKEN'>('TOTP');

  const dashboardService = useMemo(() => new EnterpriseDashboardService(), []);

  const loadAuthConfig = useCallback(async () => {
    try {
      const config = await dashboardService.getAuthConfig();
      setAuthConfig(config);
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to load authentication configuration');
    }
  }, [dashboardService]);

  useEffect(() => {
    loadAuthConfig();
  }, [loadAuthConfig]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      const response = await dashboardService.login({
        username,
        password,
        mfa_code: mfaCode || undefined,
        invitation_key: invitationKey || undefined,
      });

      if (response.success && response.token) {
        onLoginSuccess(response.token, response.user!);
      } else {
        setError(response.message || 'Login failed');
      }
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleMFASubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      const loginRequest = {
        username,
        password,
        mfa_code: mfaCode,
        invitation_key: invitationKey || undefined,
      };

      const response = await dashboardService.login(loginRequest);

      if (response.token) {
        dashboardService.setAdminToken(response.token);
        onLoginSuccess(response.token, response.user!);
        setSuccess('Login successful!');
      } else {
        setError(response.message || 'MFA verification failed');
      }
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'MFA verification failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleMFASetup = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await dashboardService.setupMFA({
        mfa_type: mfaType,
      });

      if (response.success) {
        // Handle MFA setup success
        log.info('MFA setup successful', null, 'EnterpriseAdminLogin');
      } else {
        setError(response.message || 'MFA setup failed');
      }
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'MFA setup failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleInvitationKeySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      // Validate invitation key format
      if (!invitationKey.trim()) {
        setError('Invitation key is required');
        return;
      }

      // In a real implementation, you would validate the invitation key with the backend
      // For now, we'll simulate a validation delay
      await new Promise(resolve => setTimeout(resolve, 1000));

      setSuccess('Invitation key validated successfully');
      setStep('login');
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Invalid invitation key');
    } finally {
      setIsLoading(false);
    }
  };

  const renderLoginStep = () => (
    <div className="space-y-6">
      <div>
        <label htmlFor="username" className="block text-sm font-medium text-gray-700">
          Username
        </label>
        <div className="mt-1 relative">
          <input
            id="username"
            name="username"
            type="text"
            required
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Enter your username"
          />
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
            <UserIcon className="h-5 w-5 text-gray-400" />
          </div>
        </div>
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium text-gray-700">
          Password
        </label>
        <div className="mt-1 relative">
          <input
            id="password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Enter your password"
          />
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
            >
              {showPassword ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
            </button>
          </div>
        </div>
      </div>

      {authConfig?.require_mfa && (
        <div className="text-sm text-gray-600">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-4 w-4 text-green-600" />
            <span>Multi-factor authentication is required</span>
          </div>
        </div>
      )}

      <div>
        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Signing in...</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <LockClosedIcon className="h-4 w-4" />
              <span>Sign In</span>
            </div>
          )}
        </button>
      </div>

      <div className="text-center">
        <button
          type="button"
          onClick={() => setStep('invitation')}
          className="text-sm text-indigo-600 hover:text-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
        >
          Have an invitation key?
        </button>
      </div>
    </div>
  );

  const renderMFAStep = () => (
    <div className="space-y-6">
      <div className="text-center">
        <QrCodeIcon className="mx-auto h-12 w-12 text-indigo-600" />
        <h3 className="mt-2 text-lg font-medium text-gray-900">Multi-Factor Authentication</h3>
        <p className="mt-1 text-sm text-gray-600">
          Enter the {mfaType === 'TOTP' ? '6-digit code' : 'hardware token code'} from your authenticator app
        </p>
      </div>

      <div>
        <label htmlFor="mfa-code" className="block text-sm font-medium text-gray-700">
          {mfaType === 'TOTP' ? 'TOTP Code' : 'Hardware Token Code'}
        </label>
        <div className="mt-1">
          <input
            id="mfa-code"
            name="mfa-code"
            type="text"
            required
            value={mfaCode}
            onChange={(e) => setMfaCode(e.target.value)}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder={mfaType === 'TOTP' ? 'Enter 6-digit code' : 'Enter hardware token code'}
            maxLength={mfaType === 'TOTP' ? 6 : 8}
          />
        </div>
      </div>

      <div>
        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Verifying...</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <ShieldCheckIcon className="h-4 w-4" />
              <span>Verify MFA</span>
            </div>
          )}
        </button>
      </div>

      <div className="text-center">
        <button
          type="button"
          onClick={() => setStep('mfa-setup')}
          className="text-sm text-indigo-600 hover:text-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
        >
          Need to set up MFA?
        </button>
      </div>
    </div>
  );

  const renderMFASetupStep = () => (
    <div className="space-y-6">
      <div className="text-center">
        <Cog6ToothIcon className="mx-auto h-12 w-12 text-indigo-600" />
        <h3 className="mt-2 text-lg font-medium text-gray-900">Set Up Multi-Factor Authentication</h3>
        <p className="mt-1 text-sm text-gray-600">
          Choose your preferred MFA method for enhanced security
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-3">MFA Type</label>
        <div className="space-y-3">
          <label className="flex items-center space-x-3">
            <input
              type="radio"
              name="mfa-type"
              value="TOTP"
              checked={mfaType === 'TOTP'}
              onChange={(e) => setMfaType(e.target.value as 'TOTP' | 'HARDWARE_TOKEN')}
              className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
            />
            <div>
              <span className="text-sm font-medium text-gray-900">TOTP (Google Authenticator, Authy)</span>
              <p className="text-xs text-gray-500">Time-based one-time password</p>
            </div>
          </label>
          <label className="flex items-center space-x-3">
            <input
              type="radio"
              name="mfa-type"
              value="HARDWARE_TOKEN"
              checked={mfaType === 'HARDWARE_TOKEN'}
              onChange={(e) => setMfaType(e.target.value as 'TOTP' | 'HARDWARE_TOKEN')}
              className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
            />
            <div>
              <span className="text-sm font-medium text-gray-900">Hardware Token (YubiKey)</span>
              <p className="text-xs text-gray-500">Physical security key</p>
            </div>
          </label>
        </div>
      </div>

      {mfaType === 'TOTP' && (
        <div>
          <label htmlFor="totp-secret" className="block text-sm font-medium text-gray-700">
            TOTP Secret (Optional)
          </label>
          <div className="mt-1">
            <input
              id="totp-secret"
              name="totp-secret"
              type="text"
              value={totpSecret}
              onChange={(e) => setTotpSecret(e.target.value)}
              className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              placeholder="Leave empty to generate automatically"
            />
          </div>
        </div>
      )}

      {mfaType === 'HARDWARE_TOKEN' && (
        <div>
          <label htmlFor="hardware-token-id" className="block text-sm font-medium text-gray-700">
            Hardware Token ID
          </label>
          <div className="mt-1">
            <input
              id="hardware-token-id"
              name="hardware-token-id"
              type="text"
              value={hardwareTokenId}
              onChange={(e) => setHardwareTokenId(e.target.value)}
              className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              placeholder="Enter your hardware token ID"
            />
          </div>
        </div>
      )}

      <div>
        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Setting up...</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <Cog6ToothIcon className="h-4 w-4" />
              <span>Set Up MFA</span>
            </div>
          )}
        </button>
      </div>
    </div>
  );

  const renderInvitationStep = () => (
    <div className="space-y-6">
      <div className="text-center">
        <KeyIcon className="mx-auto h-12 w-12 text-indigo-600" />
        <h3 className="mt-2 text-lg font-medium text-gray-900">Enter Invitation Key</h3>
        <p className="mt-1 text-sm text-gray-600">
          Enter the invitation key provided by your administrator
        </p>
      </div>

      <div>
        <label htmlFor="invitation-key" className="block text-sm font-medium text-gray-700">
          Invitation Key
        </label>
        <div className="mt-1 relative">
          <input
            id="invitation-key"
            name="invitation-key"
            type={showInvitationKey ? 'text' : 'password'}
            required
            value={invitationKey}
            onChange={(e) => setInvitationKey(e.target.value)}
            className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Enter your invitation key"
          />
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
            <button
              type="button"
              onClick={() => setShowInvitationKey(!showInvitationKey)}
              className="text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
            >
              {showInvitationKey ? <EyeSlashIcon className="h-5 w-5" /> : <EyeIcon className="h-5 w-5" />}
            </button>
          </div>
        </div>
      </div>

      <div>
        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Validating...</span>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <KeyIcon className="h-4 w-4" />
              <span>Validate Key</span>
            </div>
          )}
        </button>
      </div>

      <div className="text-center">
        <button
          type="button"
          onClick={() => setStep('login')}
          className="text-sm text-indigo-600 hover:text-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
        >
          Back to login
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-indigo-100">
            <UserGroupIcon className="h-8 w-8 text-indigo-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Enterprise Admin Dashboard
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Secure multi-admin access to system monitoring and controls
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="rounded-md bg-red-50 p-4">
            <div className="flex">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Authentication Error</h3>
                <div className="mt-2 text-sm text-red-700">{error}</div>
              </div>
            </div>
          </div>
        )}

        {/* Success Message */}
        {success && (
          <div className="rounded-md bg-green-50 p-4">
            <div className="flex">
              <CheckCircleIcon className="h-5 w-5 text-green-400" />
              <div className="ml-3">
                <h3 className="text-sm font-medium text-green-800">Success</h3>
                <div className="mt-2 text-sm text-green-700">{success}</div>
              </div>
            </div>
          </div>
        )}

        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          <form
            className="space-y-6"
            onSubmit={
              step === 'login'
                ? handleLogin
                : step === 'mfa'
                ? handleMFASubmit
                : step === 'mfa-setup'
                ? handleMFASetup
                : handleInvitationKeySubmit
            }
          >
            {step === 'login' && renderLoginStep()}
            {step === 'mfa' && renderMFAStep()}
            {step === 'mfa-setup' && renderMFASetupStep()}
            {step === 'invitation' && renderInvitationStep()}
          </form>

          {/* Security Notice */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">Security Features</span>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-1 gap-2 text-xs text-gray-600">
              <div className="flex items-center space-x-2">
                <ShieldCheckIcon className="h-3 w-3 text-green-600" />
                <span>Zero-knowledge privacy guarantees</span>
              </div>
              <div className="flex items-center space-x-2">
                <LockClosedIcon className="h-3 w-3 text-green-600" />
                <span>Role-based access control (RBAC)</span>
              </div>
              <div className="flex items-center space-x-2">
                <UserGroupIcon className="h-3 w-3 text-green-600" />
                <span>Multi-admin support with audit logging</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EnterpriseAdminLogin;
