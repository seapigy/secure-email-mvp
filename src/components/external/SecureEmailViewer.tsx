import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import SecurityValidationModal from './SecurityValidationModal';
import ReplyComposer from './ReplyComposer';
import { Lock, Shield, Clock, MapPin, Eye, AlertTriangle, CheckCircle, Reply } from 'lucide-react';

// Types for secure link data
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

interface SecureLinkMetadata {
  link_id: string;
  subject: string;
  sender_email: string;
  sender_name?: string;
  security_settings: SecuritySettings;
  status: string;
  message?: string;
}

interface SecureEmailContent {
  link_id: string;
  subject: string;
  body: string;
  sender_email: string;
  sender_name?: string;
  read_once: boolean;
  auto_destruct: boolean;
}

interface SecurityValidationRequest {
  link_id: string;
  password?: string;
  mfa_code?: string;
  mfa_type?: string;
  ip_address?: string;
  user_agent?: string;
}

interface SecurityValidationResponse {
  success: boolean;
  validated: boolean;
  requires_mfa?: boolean;
  mfa_type?: string;
  requires_geo?: boolean;
  error?: string;
  error_code?: string;
  decoy_message?: string;
}

const SecureEmailViewer: React.FC = () => {
  const { linkID } = useParams<{ linkID: string }>();
  const navigate = useNavigate();
  
  const [metadata, setMetadata] = useState<SecureLinkMetadata | null>(null);
  const [content, setContent] = useState<SecureEmailContent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showSecurityModal, setShowSecurityModal] = useState(false);
  const [showReplyComposer, setShowReplyComposer] = useState(false);
  const [securityStep, setSecurityStep] = useState<'password' | 'mfa' | 'geo' | 'complete'>('password');
  const [validationData, setValidationData] = useState<SecurityValidationRequest>({
    link_id: linkID || '',
  });

  useEffect(() => {
    if (linkID) {
      loadSecureLinkMetadata();
    }
  }, [linkID]);

  const loadSecureLinkMetadata = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v/${linkID}`);
      
      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.message || 'Failed to load secure link');
        return;
      }

      const data: SecureLinkMetadata = await response.json();
      setMetadata(data);

      // Check if link is accessible
      if (data.status !== 'active') {
        setError(data.message || 'This secure link is not accessible');
        return;
      }

      // Check if security validation is required
      if (data.security_settings.require_password || 
          data.security_settings.require_mfa || 
          data.security_settings.geolocation_restriction) {
        setShowSecurityModal(true);
        setSecurityStep('password');
      } else {
        // No security required, load content directly
        await loadSecureEmailContent();
      }
    } catch (err) {
      setError('Failed to load secure link metadata');
      console.error('Error loading metadata:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadSecureEmailContent = async () => {
    try {
      const response = await fetch(`/api/v/${linkID}/content`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(validationData),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setError(errorData.message || 'Failed to load email content');
        return;
      }

      const data: SecureEmailContent = await response.json();
      setContent(data);
      setShowSecurityModal(false);
    } catch (err) {
      setError('Failed to load email content');
      console.error('Error loading content:', err);
    }
  };

  const handleSecurityValidation = async (validationRequest: SecurityValidationRequest) => {
    try {
      const response = await fetch(`/api/v/${linkID}/validate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(validationRequest),
      });

      const data: SecurityValidationResponse = await response.json();

      if (!data.success) {
        // Handle validation errors
        if (data.error_code === 'LINK_DESTROYED') {
          setError('This secure link has been destroyed due to too many failed attempts');
          setShowSecurityModal(false);
          return;
        }
        
        // Show error in modal
        return { error: data.error, decoyMessage: data.decoy_message };
      }

      if (data.validated) {
        // All security checks passed
        setValidationData(validationRequest);
        await loadSecureEmailContent();
        return { success: true };
      } else {
        // More security steps required
        if (data.requires_mfa) {
          setSecurityStep('mfa');
          return { requiresMFA: true, mfaType: data.mfa_type };
        } else if (data.requires_geo) {
          setSecurityStep('geo');
          return { requiresGeo: true };
        }
      }
    } catch (err) {
      console.error('Error during security validation:', err);
      return { error: 'Security validation failed' };
    }
  };

  const handleModalClose = () => {
    setShowSecurityModal(false);
    navigate('/');
  };

  const handleReplySent = (replyID: string) => {
    setShowReplyComposer(false);
    // Optionally show a success message or update the UI
    console.log('Reply sent successfully:', replyID);
  };

  if (loading) {
    return (
      <div
        className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center"
        style={{ backgroundImage: 'url("data:image/svg+xml,%3Csvg width="60" height="60" viewBox="0 0 60 60" xmlns="http://www.w3.org/2000/svg"%3E%3Cg fill="none" fill-rule="evenodd"%3E%3Cg fill="%239C92AC" fill-opacity="0.1"%3E%3Ccircle cx="30" cy="30" r="4"/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")' }}
      >
        <div className="text-center max-w-md mx-auto px-4">
          {/* Security Branding */}
          <div className="inline-flex items-center justify-center w-16 h-16 bg-white rounded-full shadow-lg mb-6">
            <Lock className="h-8 w-8 text-blue-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Secure Email System</h1>
          <p className="text-sm text-gray-600 mb-8">Enterprise-grade secure messaging</p>
          
          {/* Loading Animation */}
          <div className="bg-white rounded-lg shadow-xl p-8">
            <div className="flex items-center justify-center mb-6">
              <div className="relative">
                <div className="animate-spin rounded-full h-12 w-12 border-4 border-blue-200 border-t-blue-600"></div>
                <div className="absolute inset-0 flex items-center justify-center">
                  <Shield className="h-6 w-6 text-blue-600" />
                </div>
              </div>
            </div>
            <h2 className="text-lg font-semibold text-gray-900 mb-2">Verifying Access</h2>
            <p className="text-sm text-gray-600 mb-4">Please wait while we securely validate your access...</p>
            
            {/* Progress Steps */}
            <div className="space-y-3">
              <div className="flex items-center text-sm text-gray-600">
                <div className="w-2 h-2 bg-blue-600 rounded-full mr-3"></div>
                <span>Validating secure link</span>
              </div>
              <div className="flex items-center text-sm text-gray-500">
                <div className="w-2 h-2 bg-gray-300 rounded-full mr-3"></div>
                <span>Checking security requirements</span>
              </div>
              <div className="flex items-center text-sm text-gray-500">
                <div className="w-2 h-2 bg-gray-300 rounded-full mr-3"></div>
                <span>Loading secure content</span>
              </div>
            </div>
          </div>
          
          {/* Security Notice */}
          <div className="mt-6 p-4 bg-blue-50 rounded-lg">
            <div className="flex items-start">
              <Shield className="h-4 w-4 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
              <div>
                <p className="text-sm font-medium text-blue-800">Security Information</p>
                <p className="text-sm text-blue-700 mt-1">
                  All access attempts are logged for security monitoring. Your privacy and security are our top priority.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6 text-center">
          <AlertTriangle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Access Denied</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            onClick={() => navigate('/')}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
          >
            Return Home
          </button>
        </div>
      </div>
    );
  }

  if (content) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4">
          {/* Header */}
          <div className="bg-white rounded-lg shadow-md p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{content.subject}</h1>
                <p className="text-gray-600">
                  From: {content.sender_name || content.sender_email}
                </p>
              </div>
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-500" />
                <span className="text-sm text-green-600 font-medium">Verified</span>
              </div>
            </div>

            {/* Security Indicators */}
            <div className="flex flex-wrap gap-2">
              {content.read_once && (
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                  <Eye className="h-3 w-3 mr-1" />
                  One-time viewing
                </span>
              )}
              {content.auto_destruct && (
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                  <AlertTriangle className="h-3 w-3 mr-1" />
                  Auto-destruct enabled
                </span>
              )}
            </div>
          </div>

          {/* Email Content */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <div className="prose max-w-none">
              <div dangerouslySetInnerHTML={{ __html: content.body }} />
            </div>
          </div>

          {/* Reply Section */}
          <div className="mt-6 border-t border-gray-200 pt-6">
            <div className="flex justify-between items-center">
              <div className="text-sm text-gray-600">
                <p>This secure message was delivered using SecureMail's encrypted email system.</p>
                {content.read_once && (
                  <p className="mt-2 text-red-600">
                    ⚠️ This message has been destroyed after viewing for security.
                  </p>
                )}
              </div>
              <button
                onClick={() => setShowReplyComposer(true)}
                className="flex items-center bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
              >
                <Reply className="h-4 w-4 mr-2" />
                Reply Securely
              </button>
            </div>
          </div>

          {/* Reply Composer Modal */}
          {showReplyComposer && (
            <ReplyComposer
              isOpen={showReplyComposer}
              linkID={linkID || ''}
              originalSubject={content.subject}
              originalSenderEmail={content.sender_email}
              originalSenderName={content.sender_name}
              onReplySent={handleReplySent}
              onCancel={() => setShowReplyComposer(false)}
            />
          )}
        </div>
      </div>
    );
  }

  if (metadata) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4">
          {/* Header */}
          <div className="bg-white rounded-lg shadow-md p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{metadata.subject}</h1>
                <p className="text-gray-600">
                  From: {metadata.sender_name || metadata.sender_email}
                </p>
              </div>
              <div className="flex items-center space-x-2">
                <Shield className="h-5 w-5 text-blue-500" />
                <span className="text-sm text-blue-600 font-medium">Secure Message</span>
              </div>
            </div>

            {/* Security Features */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              {metadata.security_settings.require_password && (
                <div className="flex items-center space-x-2">
                  <Lock className="h-4 w-4 text-gray-500" />
                  <span className="text-sm text-gray-600">Password Protected</span>
                </div>
              )}
              {metadata.security_settings.require_mfa && (
                <div className="flex items-center space-x-2">
                  <Shield className="h-4 w-4 text-gray-500" />
                  <span className="text-sm text-gray-600">
                    {metadata.security_settings.mfa_type} Verification Required
                  </span>
                </div>
              )}
              {metadata.security_settings.geolocation_restriction && (
                <div className="flex items-center space-x-2">
                  <MapPin className="h-4 w-4 text-gray-500" />
                  <span className="text-sm text-gray-600">Location Restricted</span>
                </div>
              )}
              {metadata.security_settings.time_lock && (
                <div className="flex items-center space-x-2">
                  <Clock className="h-4 w-4 text-gray-500" />
                  <span className="text-sm text-gray-600">Time-locked Access</span>
                </div>
              )}
            </div>

            {/* Access Attempts */}
            <div className="bg-gray-50 rounded-md p-3">
              <p className="text-sm text-gray-600">
                Access attempts: {metadata.security_settings.current_attempts} / {metadata.security_settings.max_access_attempts}
              </p>
            </div>
          </div>

                     {/* Security Validation Modal */}
           {showSecurityModal && (
             <SecurityValidationModal
               isOpen={showSecurityModal}
               onClose={handleModalClose}
               onValidate={handleSecurityValidation}
               securitySettings={metadata.security_settings}
               currentStep={securityStep}
               linkID={linkID || ''}
             />
           )}


        </div>
      </div>
    );
  }

  return null;
};

export default SecureEmailViewer;
