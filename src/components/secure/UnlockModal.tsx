import React, { useState, useEffect } from 'react';
import { 
  Lock, 
  X, 
  Eye, 
  EyeOff,
  AlertCircle
} from 'lucide-react';

/**
 * Unlock Modal Props Interface
 * 
 * Props for the UnlockModal component.
 */
interface UnlockModalProps {
  /** Whether the modal is open */
  isOpen: boolean;
  
  /** Callback to close the modal */
  onClose: () => void;
  
  /** Callback when password is submitted */
  onSubmit: (password: string) => void;
  
  /** Whether this is per-email mode (vs per-session mode) */
  isPerEmailMode?: boolean;
  
  /** Whether the unlock process is loading */
  isLoading?: boolean;
  
  /** Error message to display */
  error?: string;
  
  /** Whether self-destruct after failed attempts is enabled */
  selfDestructEnabled?: boolean;
  
  /** Maximum failed attempts before self-destruction */
  maxFailedAttempts?: number;
  
  /** Current number of failed attempts */
  failedAttempts?: number;
}

/**
 * UnlockModal Component
 * 
 * Modal for unlocking secure emails with password verification.
 * Features modern design with proper error handling and loading states.
 * 
 * Features:
 * - Password input with show/hide toggle for better UX
 * - Error message display for incorrect passwords
 * - Loading state during password verification
 * - Per-email vs per-session mode messaging
 * - Responsive design that works on all screen sizes
 * - Dark/light mode support
 * - Demo password hint for testing
 * - Proper form validation and submission handling
 * 
 * Usage:
 * - Used when accessing password-protected secure emails
 * - Supports both per-email and per-session password modes
 * - Integrates with session store for tracking unlocked emails
 * - Provides clear feedback for success/failure states
 */
const UnlockModal: React.FC<UnlockModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  isPerEmailMode = false,
  isLoading = false,
  error,
  selfDestructEnabled = false,
  maxFailedAttempts = 3,
  failedAttempts = 0
}) => {
  // Password input state
  const [password, setPassword] = useState('');
  
  // Password visibility toggle state
  const [showPassword, setShowPassword] = useState(false);

  /**
   * Handle escape key press to close modal
   */
  useEffect(() => {
    if (isOpen) {
      setPassword('');
      setShowPassword(false);
    }
  }, [isOpen]);

  /**
   * Handle form submission
   * @param e - Form submission event
   */
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (password.trim()) {
      onSubmit(password);
    }
  };

  /**
   * Handle modal close and reset state
   */
  const handleClose = () => {
    setPassword('');
    setShowPassword(false);
    onClose();
  };

  // Calculate remaining attempts
  const remainingAttempts = maxFailedAttempts - failedAttempts;
  const isLastAttempt = remainingAttempts === 1;

  // Don't render if modal is not open
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        {/* Background overlay */}
        <div 
          className="fixed inset-0 bg-secondary-900 bg-opacity-75 transition-opacity"
          onClick={handleClose}
        />

        {/* Modal */}
        <div className="inline-block align-bottom bg-white dark:bg-secondary-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-md sm:w-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-yellow-100 dark:bg-yellow-900/20 rounded-lg flex items-center justify-center">
                <Lock className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
              </div>
              <div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white">
                  Unlock Secure Email
                </h3>
                <p className="text-sm text-secondary-600 dark:text-secondary-400">
                  {isPerEmailMode 
                    ? 'Individual email verification required'
                    : 'Session-based verification'
                  }
                </p>
              </div>
            </div>
            <button
              onClick={handleClose}
              className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
              aria-label="Close modal"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Content */}
          <div className="px-6 py-6">
            <div className="text-center mb-6">
              <div className="w-16 h-16 bg-yellow-100 dark:bg-yellow-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <Lock className="w-8 h-8 text-yellow-600 dark:text-yellow-400" />
              </div>
              <h4 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                Password Required
              </h4>
              <p className="text-secondary-600 dark:text-secondary-400">
                {isPerEmailMode 
                  ? 'This email requires individual password verification for maximum security.'
                  : 'This message is encrypted and requires a password to decrypt.'
                }
              </p>
            </div>

            {/* Error Message */}
            {error && (
              <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <div className="flex items-center space-x-2">
                  <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400 flex-shrink-0" />
                  <p className="text-sm text-red-600 dark:text-red-400">
                    {error}
                  </p>
                </div>
              </div>
            )}

            {/* Self-Destruct Warning */}
            {selfDestructEnabled && (
              <div className={`mb-4 p-3 border rounded-lg ${
                isLastAttempt 
                  ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800' 
                  : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800'
              }`}>
                <div className="flex items-start space-x-2">
                  <AlertCircle className={`w-4 h-4 mt-0.5 flex-shrink-0 ${
                    isLastAttempt 
                      ? 'text-red-600 dark:text-red-400' 
                      : 'text-yellow-600 dark:text-yellow-400'
                  }`} />
                  <div className="text-sm">
                    <p className={`font-medium ${
                      isLastAttempt 
                        ? 'text-red-700 dark:text-red-300' 
                        : 'text-yellow-700 dark:text-yellow-300'
                    }`}>
                      {isLastAttempt 
                        ? '⚠️ Last Attempt!' 
                        : 'Self-Destruct Enabled'
                      }
                    </p>
                    <p className={`${
                      isLastAttempt 
                        ? 'text-red-600 dark:text-red-400' 
                        : 'text-yellow-600 dark:text-yellow-400'
                    }`}>
                      {isLastAttempt 
                        ? `This is your final attempt. The message will be permanently deleted if the password is incorrect.`
                        : `Failed attempts: ${failedAttempts}/${maxFailedAttempts}. Message will self-destruct after ${maxFailedAttempts} failed attempts.`
                      }
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Password Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                  Password
                </label>
                <div className="relative">
                  <input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    disabled={isLoading}
                    className="w-full pl-4 pr-12 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
                    autoFocus
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-200"
                  >
                    {showPassword ? (
                      <EyeOff className="w-4 h-4" />
                    ) : (
                      <Eye className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>

              {/* Demo Password Hint */}
              <div className="text-xs text-secondary-500 dark:text-secondary-400 text-center">
                Demo password: <code className="bg-secondary-100 dark:bg-secondary-700 px-1 py-0.5 rounded">demo123</code>
              </div>

              {/* Cancel Note */}
              <div className="text-xs text-secondary-500 dark:text-secondary-400 text-center">
                You can cancel and return to the email list if you don&apos;t want to unlock this email now.
              </div>

              {/* Action Buttons */}
              <div className="flex space-x-3 pt-4">
                <button
                  type="button"
                  onClick={handleClose}
                  disabled={isLoading}
                  className="flex-1 px-4 py-2 border border-secondary-300 dark:border-secondary-600 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isLoading || !password.trim()}
                  className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? (
                    <div className="flex items-center justify-center space-x-2">
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      <span>Unlocking...</span>
                    </div>
                  ) : (
                    'Unlock Email'
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UnlockModal;
