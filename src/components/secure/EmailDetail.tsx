import React, { useState } from 'react';
import { 
  ArrowLeft, 
  Lock, 
  Globe, 
  Eye, 
  AlertTriangle,
  Download,
  Trash2,
  RefreshCw,
  User,
  FileText,
  Copy,
  Shield as ShieldIcon,
  Globe as GlobeIcon,
  Eye as EyeIcon,
  AlertTriangle as AlertTriangleIcon,
  Clock
} from 'lucide-react';
import { SecureEmail, StatusType, SecuritySettings } from '@/types/secureEmail';
import { useSessionStore } from '@/stores/sessionStore';
import UnlockModal from './UnlockModal';

/**
 * Email Detail Props Interface
 * 
 * Props for the EmailDetail component.
 */
interface EmailDetailProps {
  /** The secure email to display */
  email: SecureEmail;
  
  /** Callback for back navigation */
  onBack: () => void;
  
  /** Whether to use compact layout */
  isCompact?: boolean;
  
  /** Security settings for the current session */
  securitySettings?: SecuritySettings;
}

/**
 * EmailDetail Component
 * 
 * Detailed view of a secure email with encrypted content and security metadata.
 * Features comprehensive security information display and modern privacy-first design.
 * 
 * Features:
 * - Encrypted content display with password protection
 * - Security metadata visualization (fingerprint, status, etc.)
 * - Attachment handling and display
 * - Security status indicators with visual badges
 * - Responsive design with dark/light mode support
 * - Per-email password protection with session tracking
 * - Self-destruct after failed attempts feature
 * - Comprehensive security information display
 * - Copy-to-clipboard functionality for metadata
 * - Mobile-responsive layout
 * 
 * Security Features:
 * - Password protection with unlock modal
 * - Session-based email unlocking
 * - Failed attempt tracking and self-destruction
 * - Security status visualization
 * - Cryptographic fingerprint display
 * - Access attempt monitoring
 * 
 * Layout Modes:
 * - Compact mode for split-view desktop layout
 * - Full mode for mobile detail view
 * - Responsive design that adapts to screen size
 */
const EmailDetail: React.FC<EmailDetailProps> = ({ email, onBack, isCompact = false, securitySettings }) => {
  // State for decrypted content display
  const [showDecrypted, setShowDecrypted] = useState(false);
  
  // Unlock modal state management
  const [showUnlockModal, setShowUnlockModal] = useState(false);
  const [isUnlocking, setIsUnlocking] = useState(false);
  const [unlockError, setUnlockError] = useState<string>('');
  const [userManuallyClosed, setUserManuallyClosed] = useState(false);
  
  // Self-destruct feature state
  const [failedAttempts, setFailedAttempts] = useState(0);
  const [isSelfDestructed, setIsSelfDestructed] = useState(false);
  
  // Session store for tracking unlocked emails
  const { unlockEmail, isEmailUnlocked } = useSessionStore();

  /**
   * Check if the email has expired based on expiresAt timestamp
   * @returns true if the email has expired, false otherwise
   */
  const isEmailExpired = (): boolean => {
    if (!email.expiresAt) return false;
    const expirationDate = new Date(email.expiresAt);
    const now = new Date();
    return now > expirationDate;
  };

  /**
   * Get time remaining until expiration
   * @returns string with time remaining or null if no expiration
   */
  const getTimeRemaining = (): string | null => {
    if (!email.expiresAt) return null;
    const expirationDate = new Date(email.expiresAt);
    const now = new Date();
    const diff = expirationDate.getTime() - now.getTime();
    
    if (diff <= 0) return null;
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 24) {
      const days = Math.floor(hours / 24);
      return `${days} day${days > 1 ? 's' : ''}`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  };

  /**
   * Check if we need to show password prompt
   * Determines whether to show unlock modal based on:
   * - Email password protection status
   * - Per-email vs per-session mode
   * - Whether email is already unlocked
   * - Whether email has been self-destructed
   * - Whether user has manually closed the modal
   * - Whether email has expired
   */
  const needsPasswordPrompt = email.passwordProtected && !isSelfDestructed && !isEmailExpired() && !userManuallyClosed && (
    securitySettings?.perEmailPassword ? !isEmailUnlocked(email.id) : !showDecrypted
  );

  /**
   * Show unlock modal when email is selected and needs password
   * Automatically triggers unlock modal when email requires password verification
   * Only shows if user hasn't manually closed it
   */
  React.useEffect(() => {
    if (needsPasswordPrompt && !showUnlockModal && !userManuallyClosed) {
      setShowUnlockModal(true);
    }
  }, [needsPasswordPrompt, showUnlockModal, userManuallyClosed]);

  /**
   * Get status configuration for visual display
   * @param status - The email status
   * @returns Configuration object with colors, labels, and icons
   */
  const getStatusConfig = (status: StatusType) => {
    const configs = {
      pending: {
        label: 'Pending Access',
        color: 'text-yellow-600 dark:text-yellow-400',
        bgColor: 'bg-yellow-100 dark:bg-yellow-900/20',
        icon: '⏳'
      },
      opened: {
        label: 'Opened',
        color: 'text-green-600 dark:text-green-400',
        bgColor: 'bg-green-100 dark:bg-green-900/20',
        icon: '✓'
      },
      expired: {
        label: 'Expired',
        color: 'text-red-600 dark:text-red-400',
        bgColor: 'bg-red-100 dark:bg-red-900/20',
        icon: '⏰'
      },
      revoked: {
        label: 'Revoked',
        color: 'text-gray-600 dark:text-gray-400',
        bgColor: 'bg-gray-100 dark:bg-gray-900/20',
        icon: '🚫'
      }
    };
    return configs[status];
  };

  // Handle unlock modal submission
  const handleUnlockSubmit = async (password: string) => {
    setIsUnlocking(true);
    setUnlockError('');
    
    // Simulate API call delay
    await new Promise(resolve => setTimeout(resolve, 500));
    
    if (password === 'demo123') { // Mock password for demo
      if (securitySettings?.perEmailPassword) {
        // Per-email mode: unlock this specific email
        unlockEmail(email.id);
      } else {
        // Per-session mode: unlock all emails for this session
        setShowDecrypted(true);
      }
      setShowUnlockModal(false);
      setFailedAttempts(0); // Reset failed attempts on success
    } else {
      // Increment failed attempts
      const newFailedAttempts = failedAttempts + 1;
      setFailedAttempts(newFailedAttempts);
      
      // Check if self-destruct should be triggered
      if (email.selfDestructAfterAttempts && newFailedAttempts >= email.maxFailedAttempts) {
        setIsSelfDestructed(true);
        setShowUnlockModal(false);
        setUnlockError('');
      } else {
        setUnlockError(`Incorrect password. Please try again. (${newFailedAttempts}/${email.maxFailedAttempts} attempts)`);
      }
    }
    
    setIsUnlocking(false);
  };

  // Handle unlock modal close
  const handleUnlockClose = () => {
    setShowUnlockModal(false);
    setUnlockError('');
    setIsUnlocking(false);
    setUserManuallyClosed(true); // Mark that user manually closed
  };

  // Handle back navigation - allow going back even if not unlocked
  const handleBack = () => {
    if (needsPasswordPrompt && !userManuallyClosed) {
      // If modal is open and user hasn't manually closed, close it first
      setShowUnlockModal(false);
      setUserManuallyClosed(true);
    } else {
      // Otherwise proceed with normal back navigation
      onBack();
    }
  };

  // Copy to clipboard
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const statusConfig = getStatusConfig(email.status);

  return (
    <div className={`${isCompact ? 'h-full' : 'min-h-screen bg-secondary-50 dark:bg-secondary-900'}`}>
      <div className={`${isCompact ? 'h-full flex flex-col' : 'max-w-4xl mx-auto p-6'}`}>
        {/* Header */}
        <div className={`flex items-center justify-between ${isCompact ? 'mb-4' : 'mb-6'}`}>
          <div className="flex items-center space-x-4 min-w-0">
            {!isCompact && (
              <button
                onClick={handleBack}
                className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200 flex-shrink-0"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
            )}
            <div className="min-w-0 flex-1">
              <h1 className={`${isCompact ? 'text-lg' : 'text-2xl'} font-bold text-secondary-900 dark:text-white truncate`}>
                {email.subject}
              </h1>
              <p className="text-sm text-secondary-600 dark:text-secondary-400 truncate">
                Secure Email Details
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-3 flex-shrink-0">
            <button className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
              <RefreshCw className="w-5 h-5" />
            </button>
            <button className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200">
              <Trash2 className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Email Content */}
        <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 shadow-sm overflow-hidden">
          {/* Email Header */}
          <div className="p-4 sm:p-6 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-4">
              <div className="flex items-center space-x-3 min-w-0">
                <div className="w-10 h-10 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center flex-shrink-0">
                  <User className="w-5 h-5 text-primary-600 dark:text-primary-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="font-medium text-secondary-900 dark:text-white truncate">
                    {email.from}
                  </p>
                  <p className="text-sm text-secondary-600 dark:text-secondary-400 truncate">
                    to {email.to}
                  </p>
                </div>
              </div>
              
              <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${statusConfig.bgColor} ${statusConfig.color} flex-shrink-0`}>
                <span className="mr-1">{statusConfig.icon}</span>
                <span className="truncate">{statusConfig.label}</span>
              </span>
            </div>

            {/* Security Indicators */}
            <div className="flex flex-wrap items-center gap-4 mb-4">
              {email.passwordProtected && (
                <div className="flex items-center space-x-2 text-yellow-600 dark:text-yellow-400">
                  <Lock className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">Password Protected</span>
                </div>
              )}
              {email.geolocationRestricted && (
                <div className="flex items-center space-x-2 text-blue-600 dark:text-blue-400">
                  <Globe className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">Location Restricted</span>
                </div>
              )}
              {email.readOnce && (
                <div className="flex items-center space-x-2 text-red-600 dark:text-red-400">
                  <Eye className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">Read Once</span>
                </div>
              )}
              {email.autoDestruct && (
                <div className="flex items-center space-x-2 text-orange-600 dark:text-orange-400">
                  <AlertTriangle className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">Auto-Destruct</span>
                </div>
              )}
              {email.selfDestructAfterAttempts && (
                <div className="flex items-center space-x-2 text-red-600 dark:text-red-400">
                  <Trash2 className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">Self-Destruct After {email.maxFailedAttempts} Attempts</span>
                </div>
              )}
              {email.expiresAt && (
                <div className={`flex items-center space-x-2 ${isEmailExpired() ? 'text-red-600 dark:text-red-400' : 'text-purple-600 dark:text-purple-400'}`}>
                  <Clock className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium truncate">
                    {isEmailExpired() ? 'Expired' : `Expires in ${getTimeRemaining()}`}
                  </span>
                </div>
              )}
            </div>

            {/* Metadata */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
              <div className="min-w-0">
                <p className="text-secondary-600 dark:text-secondary-400 truncate">Created</p>
                <p className="font-medium text-secondary-900 dark:text-white truncate">
                  {new Date(email.date).toLocaleString()}
                </p>
              </div>
              <div className="min-w-0">
                <p className="text-secondary-600 dark:text-secondary-400 truncate">Expires</p>
                <p className="font-medium text-secondary-900 dark:text-white truncate">
                  {email.expiresAt 
                    ? new Date(email.expiresAt).toLocaleString()
                    : new Date(email.expires).toLocaleString()
                  }
                </p>
              </div>
              <div className="min-w-0">
                <p className="text-secondary-600 dark:text-secondary-400 truncate">Access Attempts</p>
                <p className="font-medium text-secondary-900 dark:text-white truncate">
                  {email.accessAttempts} / {email.maxAttempts}
                </p>
              </div>
              <div className="min-w-0">
                <p className="text-secondary-600 dark:text-secondary-400 truncate">Status</p>
                <p className="font-medium text-secondary-900 dark:text-white capitalize truncate">
                  {isEmailExpired() ? 'expired' : email.status}
                </p>
              </div>
            </div>
          </div>

          {/* Email Content */}
          <div className="p-4 sm:p-6">
            {isEmailExpired() ? (
              /* Expired Email Message */
              <div className="text-center py-8 sm:py-12">
                <div className="w-16 h-16 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Clock className="w-8 h-8 text-red-600 dark:text-red-400" />
                </div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                  Message Has Expired
                </h3>
                <p className="text-secondary-600 dark:text-secondary-400 mb-6">
                  ⏰ This message has expired and is no longer accessible.
                </p>
                <div className="bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 rounded-lg p-4">
                  <p className="text-sm text-red-700 dark:text-red-300">
                    The message content has been permanently deleted and cannot be recovered.
                  </p>
                  {email.expiresAt && (
                    <p className="text-xs text-red-600 dark:text-red-400 mt-2">
                      Expired on: {new Date(email.expiresAt).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>
            ) : isSelfDestructed ? (
              /* Self-Destruct Message */
              <div className="text-center py-8 sm:py-12">
                <div className="w-16 h-16 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Trash2 className="w-8 h-8 text-red-600 dark:text-red-400" />
                </div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                  Message Permanently Deleted
                </h3>
                <p className="text-secondary-600 dark:text-secondary-400 mb-6">
                  🔒 This message has been permanently deleted after {email.maxFailedAttempts} failed access attempts.
                </p>
                <div className="bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 rounded-lg p-4">
                  <p className="text-sm text-red-700 dark:text-red-300">
                    The message content has been permanently destroyed and cannot be recovered.
                  </p>
                </div>
              </div>
            ) : needsPasswordPrompt ? (
              /* Password Protection Screen */
              <div className="text-center py-8 sm:py-12">
                <div className="w-16 h-16 bg-yellow-100 dark:bg-yellow-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Lock className="w-8 h-8 text-yellow-600 dark:text-yellow-400" />
                </div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                  Password Protected
                </h3>
                <p className="text-secondary-600 dark:text-secondary-400 mb-6">
                  {securitySettings?.perEmailPassword 
                    ? 'This email requires individual password verification for maximum security.'
                    : 'This message is encrypted and requires a password to decrypt.'
                  }
                </p>
                
                <button
                  onClick={() => setShowUnlockModal(true)}
                  className="px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200"
                >
                  Unlock Email
                </button>
                
                <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-4">
                  Demo password: demo123
                </p>
              </div>
            ) : (
              /* Decrypted Content */
              <div className="space-y-6">
                <div className="prose prose-sm max-w-none dark:prose-invert">
                  <div className="bg-secondary-50 dark:bg-secondary-700 p-4 rounded-lg">
                    <p className="text-sm text-secondary-600 dark:text-secondary-400 mb-2">
                      Decrypted Content:
                    </p>
                    <div className="text-secondary-900 dark:text-white break-words">
                      {atob(email.content)}
                    </div>
                  </div>
                </div>

                {/* Attachments */}
                {email.attachments && email.attachments.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-3">
                      Attachments ({email.attachments.length})
                    </h4>
                    <div className="space-y-2">
                      {email.attachments.map((attachment, index) => (
                        <div
                          key={index}
                          className="flex items-center justify-between p-3 bg-secondary-50 dark:bg-secondary-700 rounded-lg"
                        >
                          <div className="flex items-center space-x-3 min-w-0">
                            <div className="w-8 h-8 bg-secondary-200 dark:bg-secondary-600 rounded flex items-center justify-center flex-shrink-0">
                              <FileText className="w-4 h-4 text-secondary-600 dark:text-secondary-400" />
                            </div>
                            <div className="min-w-0 flex-1">
                              <p className="text-sm font-medium text-secondary-900 dark:text-white truncate">
                                {attachment.name}
                              </p>
                              <p className="text-xs text-secondary-500 dark:text-secondary-400 truncate">
                                {attachment.size} {attachment.encrypted && '• Encrypted'}
                              </p>
                            </div>
                          </div>
                          <button className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-600 rounded-lg transition-colors duration-200 flex-shrink-0">
                            <Download className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Security Details */}
        <div className="mt-6 bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 shadow-sm overflow-hidden">
          <div className="p-4 sm:p-6">
            <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-4">
              Security Details
            </h3>
            
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Security Features */}
              <div className="min-w-0">
                <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-3">
                  Security Features
                </h4>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <ShieldIcon className="w-4 h-4 text-green-600 dark:text-green-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Password Protection</span>
                    </div>
                    <span className={`text-sm font-medium ${email.passwordProtected ? 'text-green-600 dark:text-green-400' : 'text-gray-400'} flex-shrink-0`}>
                      {email.passwordProtected ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  
                  {/* Self-Destruct After Failed Attempts */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <Trash2 className="w-4 h-4 text-red-600 dark:text-red-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Self-Destruct After Failed Attempts</span>
                    </div>
                    <span className={`text-sm font-medium ${email.selfDestructAfterAttempts ? 'text-red-600 dark:text-red-400' : 'text-gray-400'} flex-shrink-0`}>
                      {email.selfDestructAfterAttempts ? `Enabled (${email.maxFailedAttempts} attempts)` : 'Disabled'}
                    </span>
                  </div>
                  
                  {/* Access Attempts Counter */}
                  {email.selfDestructAfterAttempts && (
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2 min-w-0">
                        <AlertTriangle className="w-4 h-4 text-yellow-600 dark:text-yellow-400 flex-shrink-0" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Failed Attempts</span>
                      </div>
                      <span className={`text-sm font-medium ${failedAttempts > 0 ? 'text-yellow-600 dark:text-yellow-400' : 'text-gray-400'} flex-shrink-0`}>
                        {failedAttempts} / {email.maxFailedAttempts}
                      </span>
                    </div>
                  )}
                  
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <GlobeIcon className="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Geolocation Lock</span>
                    </div>
                    <span className={`text-sm font-medium ${email.geolocationRestricted ? 'text-blue-600 dark:text-blue-400' : 'text-gray-400'} flex-shrink-0`}>
                      {email.geolocationRestricted ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <EyeIcon className="w-4 h-4 text-red-600 dark:text-red-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Read Once</span>
                    </div>
                    <span className={`text-sm font-medium ${email.readOnce ? 'text-red-600 dark:text-red-400' : 'text-gray-400'} flex-shrink-0`}>
                      {email.readOnce ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <AlertTriangleIcon className="w-4 h-4 text-orange-600 dark:text-orange-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Auto-Destruct</span>
                    </div>
                    <span className={`text-sm font-medium ${email.autoDestruct ? 'text-orange-600 dark:text-orange-400' : 'text-gray-400'} flex-shrink-0`}>
                      {email.autoDestruct ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  
                  {/* Email Expiration */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 min-w-0">
                      <Clock className="w-4 h-4 text-purple-600 dark:text-purple-400 flex-shrink-0" />
                      <span className="text-sm text-secondary-700 dark:text-secondary-300 truncate">Email Expiration</span>
                    </div>
                    <span className={`text-sm font-medium ${email.expiresAt ? (isEmailExpired() ? 'text-red-600 dark:text-red-400' : 'text-purple-600 dark:text-purple-400') : 'text-gray-400'} flex-shrink-0`}>
                      {email.expiresAt 
                        ? (isEmailExpired() ? 'Expired' : `Expires in ${getTimeRemaining()}`)
                        : 'Disabled'
                      }
                    </span>
                  </div>
                </div>
              </div>

              {/* Technical Details */}
              <div className="min-w-0">
                <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-3">
                  Technical Details
                </h4>
                <div className="space-y-3">
                  <div className="min-w-0">
                    <p className="text-xs text-secondary-500 dark:text-secondary-400">Fingerprint Hash</p>
                    <div className="flex items-center space-x-2 mt-1">
                      <code className="text-xs bg-secondary-100 dark:bg-secondary-700 px-2 py-1 rounded truncate flex-1">
                        {email.fingerprint}
                      </code>
                      <button
                        onClick={() => copyToClipboard(email.fingerprint)}
                        className="p-1 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-200 flex-shrink-0"
                      >
                        <Copy className="w-3 h-3" />
                      </button>
                    </div>
                  </div>
                  
                  <div className="min-w-0">
                    <p className="text-xs text-secondary-500 dark:text-secondary-400">Allowed Countries</p>
                    <p className="text-sm text-secondary-700 dark:text-secondary-300 truncate">
                      {email.allowedCountries.length > 0 
                        ? email.allowedCountries.join(', ') 
                        : 'No restrictions'
                      }
                    </p>
                  </div>
                  
                  <div className="min-w-0">
                    <p className="text-xs text-secondary-500 dark:text-secondary-400">Metadata Stripped</p>
                    <p className="text-sm text-secondary-700 dark:text-secondary-300 truncate">
                      {email.metadataStripped ? 'Yes' : 'No'}
                    </p>
                  </div>
                  
                  <div className="min-w-0">
                    <p className="text-xs text-secondary-500 dark:text-secondary-400">Tamper Alerts</p>
                    <p className="text-sm text-secondary-700 dark:text-secondary-300 truncate">
                      {email.tamperAlerts ? 'Enabled' : 'Disabled'}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Unlock Modal */}
      <UnlockModal
        isOpen={showUnlockModal}
        onClose={handleUnlockClose}
        onSubmit={handleUnlockSubmit}
        isPerEmailMode={securitySettings?.perEmailPassword}
        isLoading={isUnlocking}
        error={unlockError}
        selfDestructEnabled={email.selfDestructAfterAttempts}
        maxFailedAttempts={email.maxFailedAttempts}
        failedAttempts={failedAttempts}
      />
    </div>
  );
};

export default EmailDetail; 