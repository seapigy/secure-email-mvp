import React from 'react';
import { AlertTriangle, Lock, Shield, Clock, XCircle, ArrowLeft, RefreshCw } from 'lucide-react';

interface ErrorPageProps {
  errorType: 'expired' | 'revoked' | 'invalid_access' | 'failed_validation' | 'not_found' | 'system_error';
  errorCode?: string;
  errorMessage?: string;
  linkID?: string;
  onRetry?: () => void;
  onGoBack?: () => void;
  showDecoyMessage?: boolean;
  decoyMessage?: string;
}

const ErrorPage: React.FC<ErrorPageProps> = ({
  errorType,
  errorCode,
  errorMessage,
  linkID,
  onRetry,
  onGoBack,
  showDecoyMessage = false,
  decoyMessage,
}) => {
  const getErrorConfig = () => {
    switch (errorType) {
      case 'expired':
        return {
          icon: Clock,
          iconColor: 'text-orange-500',
          bgColor: 'bg-orange-50',
          borderColor: 'border-orange-200',
          title: 'Link Expired',
          message: 'This secure link has expired and is no longer accessible.',
          description: 'Secure links have a limited lifespan for security purposes. Please contact the sender for a new link.',
          severity: 'warning' as const,
        };
      case 'revoked':
        return {
          icon: XCircle,
          iconColor: 'text-red-500',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
          title: 'Link Revoked',
          message: 'This secure link has been revoked by the sender.',
          description: 'The sender has cancelled access to this secure message. Please contact them directly if you need this information.',
          severity: 'error' as const,
        };
      case 'invalid_access':
        return {
          icon: Shield,
          iconColor: 'text-blue-500',
          bgColor: 'bg-blue-50',
          borderColor: 'border-blue-200',
          title: 'Access Denied',
          message: 'You do not have permission to access this secure message.',
          description: 'This message may be restricted to specific users or locations. Please verify your access credentials.',
          severity: 'error' as const,
        };
      case 'failed_validation':
        return {
          icon: AlertTriangle,
          iconColor: 'text-yellow-500',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200',
          title: 'Security Validation Failed',
          message: 'Unable to verify your identity or location.',
          description: 'For security reasons, we could not validate your access. This may be due to location restrictions or invalid credentials.',
          severity: 'warning' as const,
        };
      case 'not_found':
        return {
          icon: Lock,
          iconColor: 'text-gray-500',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200',
          title: 'Link Not Found',
          message: 'This secure link could not be found or has been removed.',
          description: 'The link you are trying to access does not exist or may have been deleted for security reasons.',
          severity: 'info' as const,
        };
      case 'system_error':
        return {
          icon: AlertTriangle,
          iconColor: 'text-red-500',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
          title: 'System Error',
          message: 'We encountered an unexpected error while processing your request.',
          description: 'Our system is experiencing technical difficulties. Please try again later or contact support if the problem persists.',
          severity: 'error' as const,
        };
      default:
        return {
          icon: AlertTriangle,
          iconColor: 'text-gray-500',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200',
          title: 'Error',
          message: 'An error occurred while processing your request.',
          description: 'Please try again or contact support if the problem persists.',
          severity: 'info' as const,
        };
    }
  };

  const config = getErrorConfig();
  const IconComponent = config.icon;

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        {/* Security Branding Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-white rounded-full shadow-lg mb-4">
            <Lock className="h-8 w-8 text-blue-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Secure Email System</h1>
          <p className="text-sm text-gray-600">Enterprise-grade secure messaging</p>
        </div>

        {/* Error Card */}
        <div className={`bg-white rounded-lg shadow-xl border ${config.borderColor} overflow-hidden`}>
          {/* Error Header */}
          <div className={`${config.bgColor} px-6 py-4 border-b ${config.borderColor}`}>
            <div className="flex items-center">
              <IconComponent className={`h-6 w-6 ${config.iconColor} mr-3`} />
              <div>
                <h2 className="text-lg font-semibold text-gray-900">{config.title}</h2>
                {errorCode && (
                  <p className="text-sm text-gray-600">Error Code: {errorCode}</p>
                )}
              </div>
            </div>
          </div>

          {/* Error Content */}
          <div className="px-6 py-6">
            <div className="mb-6">
              <p className="text-gray-900 font-medium mb-2">{config.message}</p>
              <p className="text-sm text-gray-600 leading-relaxed">{config.description}</p>
            </div>

            {/* Custom Error Message */}
            {errorMessage && (
              <div className="mb-6 p-3 bg-gray-50 rounded-md">
                <p className="text-sm text-gray-700">{errorMessage}</p>
              </div>
            )}

            {/* Decoy Message (if enabled) */}
            {showDecoyMessage && decoyMessage && (
              <div className="mb-6 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                <div className="flex items-start">
                  <AlertTriangle className="h-4 w-4 text-yellow-600 mt-0.5 mr-2 flex-shrink-0" />
                  <div>
                    <p className="text-sm font-medium text-yellow-800">Security Notice</p>
                    <p className="text-sm text-yellow-700 mt-1">{decoyMessage}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Link ID (for debugging) */}
            {linkID && (
              <div className="mb-6 p-3 bg-gray-50 rounded-md">
                <p className="text-xs text-gray-500">
                  <span className="font-medium">Link ID:</span> {linkID}
                </p>
              </div>
            )}

            {/* Security Information */}
            <div className="mb-6 p-4 bg-blue-50 rounded-md">
              <div className="flex items-start">
                <Shield className="h-4 w-4 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium text-blue-800">Security Information</p>
                  <p className="text-sm text-blue-700 mt-1">
                    All access attempts are logged for security monitoring. If you believe this is an error, 
                    please contact the sender directly.
                  </p>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col space-y-3">
              {onRetry && (
                <button
                  onClick={onRetry}
                  className="w-full flex items-center justify-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Try Again
                </button>
              )}
              
              {onGoBack && (
                <button
                  onClick={onGoBack}
                  className="w-full flex items-center justify-center px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                >
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Go Back
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center mt-8">
          <p className="text-xs text-gray-500">
            © 2024 Secure Email System. All rights reserved.
          </p>
          <div className="flex items-center justify-center mt-2 space-x-4">
            <div className="flex items-center text-xs text-gray-400">
              <Shield className="h-3 w-3 mr-1" />
              Enterprise Security
            </div>
            <div className="flex items-center text-xs text-gray-400">
              <Lock className="h-3 w-3 mr-1" />
              End-to-End Encryption
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ErrorPage;
