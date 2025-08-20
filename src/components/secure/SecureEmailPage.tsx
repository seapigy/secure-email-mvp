import React, { useState } from 'react';
import { 
  Settings,
  Moon,
  Sun,
  Shield,
  Plus,
  Search,
  ArrowLeft
} from 'lucide-react';
import EmailInbox from './EmailInbox';
import EmailDetail from './EmailDetail';
import SecuritySettings from './SecuritySettings';
import ComposeModal from './ComposeModal';
import { SecureEmail, SecuritySettings as SecuritySettingsType } from '@/types/secureEmail';
import { useTheme } from '@/hooks/useTheme';

/**
 * SecureEmailPage Component
 * 
 * Split-view secure email interface inspired by ProtonMail and Tutanota.
 * Features a modern, privacy-first design with responsive layout.
 * 
 * Features:
 * - Split-view layout (desktop): Inbox list (30-40%) + Message detail (60-70%)
 * - Mobile responsive: Single panel with inbox/detail toggle
 * - Email inbox with filtering and sorting
 * - Detailed email view with security metadata
 * - Comprehensive security settings panel
 * - Dark/light mode toggle
 * - Mock data integration
 */
const SecureEmailPage: React.FC = () => {
  const [selectedEmail, setSelectedEmail] = useState<SecureEmail | null>(null);
  const [selectedAt, setSelectedAt] = useState<number>(Date.now());
  const [showSecuritySettings, setShowSecuritySettings] = useState(false);
  const [showComposeModal, setShowComposeModal] = useState(false);
  const [isMobileDetailView, setIsMobileDetailView] = useState(false);
  const { toggleTheme, isDark } = useTheme();
  const [securitySettings, setSecuritySettings] = useState<SecuritySettingsType>({
    passwordProtection: false,
    perEmailPassword: false, // New: Require password for every secure email
    geolocationLock: false,
    allowedCountries: [],
    timeLock: false,
    autoDestruct: false,
    readOnce: false,
    remoteRevoke: false,
    decoyMessage: false,
    fingerprintHash: '',
    stripMetadata: true,
    tamperAlerts: true,
    selfDestructAfterAttempts: false,
    maxFailedAttempts: 3,
  });

  // Handle email selection from inbox
  const handleEmailSelect = (email: SecureEmail) => {
    setSelectedEmail(email);
    setSelectedAt(Date.now()); // Update timestamp to trigger modal reset
    // On mobile, switch to detail view
    if (window.innerWidth < 768) {
      setIsMobileDetailView(true);
    }
  };

  // Handle back to inbox (mobile)
  const handleBackToInbox = () => {
    setIsMobileDetailView(false);
  };

  // Handle security settings change
  const handleSecuritySettingsChange = (newSettings: SecuritySettingsType) => {
    setSecuritySettings(newSettings);
    setShowSecuritySettings(false);
  };

  // Handle compose modal
  const handleOpenCompose = () => {
    setShowComposeModal(true);
  };

  const handleCloseCompose = () => {
    setShowComposeModal(false);
  };

  return (
    <div className="min-h-screen bg-secondary-50 dark:bg-secondary-900 w-full overflow-x-hidden">
      {/* Header */}
      <header className="bg-white dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700 w-full">
        <div className="px-4 sm:px-6 py-4">
          <div className="flex items-center justify-between min-w-0">
            <div className="flex items-center space-x-4 min-w-0">
              {/* Mobile back button */}
              {isMobileDetailView && (
                <button
                  onClick={handleBackToInbox}
                  className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200 md:hidden flex-shrink-0"
                >
                  <ArrowLeft className="w-5 h-5" />
                </button>
              )}
              
              <div className="flex items-center space-x-3 min-w-0">
                <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center flex-shrink-0">
                  <Shield className="w-5 h-5 text-white" />
                </div>
                <div className="min-w-0">
                  <h1 className="text-lg sm:text-xl font-bold text-secondary-900 dark:text-white truncate">
                    SecureMail
                  </h1>
                  <p className="text-xs text-secondary-600 dark:text-secondary-400 truncate">
                    Privacy-first encrypted messaging
                  </p>
                </div>
              </div>
            </div>

            <div className="flex items-center space-x-2 sm:space-x-3 flex-shrink-0">
              {/* Search - hidden on mobile detail view */}
              {!isMobileDetailView && (
                <div className="relative hidden sm:block">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-secondary-400" />
                  <input
                    type="text"
                    placeholder="Search secure emails..."
                    className="pl-10 pr-4 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm w-48 lg:w-64"
                  />
                </div>
              )}

              {/* Theme Toggle */}
              <button
                onClick={toggleTheme}
                className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
                title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
              </button>

              {/* Security Settings */}
              <button
                onClick={() => setShowSecuritySettings(true)}
                className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
                title="Security Settings"
              >
                <Settings className="w-5 h-5" />
              </button>

              {/* New Secure Email - hidden on mobile detail view */}
              {!isMobileDetailView && (
                <button 
                  onClick={handleOpenCompose}
                  className="flex items-center space-x-2 px-3 sm:px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200"
                >
                  <Plus className="w-4 h-4" />
                  <span className="hidden sm:inline">New Secure Email</span>
                  <span className="sm:hidden">New</span>
                </button>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Split View */}
      <main className="flex h-[calc(100vh-80px)] w-full">
        {/* Desktop Layout */}
        <div className="hidden md:flex w-full">
          {/* Left Panel - Inbox List (35% width) */}
          <div className="w-[35%] border-r border-secondary-200 dark:border-secondary-700 bg-white dark:bg-secondary-800 min-w-0">
            <EmailInbox onEmailSelect={handleEmailSelect} selectedEmail={selectedEmail} />
          </div>
          
          {/* Right Panel - Message Detail (65% width) */}
          <div className="w-[65%] bg-secondary-50 dark:bg-secondary-900 min-w-0 p-4 sm:p-6">
            {selectedEmail ? (
              <EmailDetail 
                email={selectedEmail} 
                onBack={() => setSelectedEmail(null)}
                isCompact={true}
                securitySettings={securitySettings}
                selectedAt={selectedAt}
              />
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <div className="w-16 h-16 bg-secondary-200 dark:bg-secondary-700 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Shield className="w-8 h-8 text-secondary-400 dark:text-secondary-500" />
                  </div>
                  <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                    Select a Message
                  </h3>
                  <p className="text-secondary-600 dark:text-secondary-400">
                    Choose an email from the inbox to view its details
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Mobile Layout */}
        <div className="md:hidden w-full">
          {isMobileDetailView ? (
            /* Mobile Detail View */
            <div className="h-full bg-secondary-50 dark:bg-secondary-900 w-full">
              {selectedEmail && (
                <EmailDetail 
                  email={selectedEmail} 
                  onBack={handleBackToInbox}
                  isCompact={false}
                  securitySettings={securitySettings}
                  selectedAt={selectedAt}
                />
              )}
            </div>
          ) : (
            /* Mobile Inbox View */
            <div className="h-full bg-white dark:bg-secondary-800 w-full">
              <EmailInbox onEmailSelect={handleEmailSelect} selectedEmail={selectedEmail} />
            </div>
          )}
        </div>
      </main>

      {/* Security Settings Modal */}
      <SecuritySettings
        isOpen={showSecuritySettings}
        onClose={() => setShowSecuritySettings(false)}
        settings={securitySettings}
        onSettingsChange={handleSecuritySettingsChange}
      />

      {/* Compose Modal */}
      <ComposeModal
        isOpen={showComposeModal}
        onClose={handleCloseCompose}
      />

      {/* Demo Notice */}
      <div className="fixed bottom-4 right-4 bg-yellow-100 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4 max-w-sm z-50">
        <div className="flex items-start space-x-3">
          <div className="w-5 h-5 bg-yellow-600 dark:bg-yellow-400 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
            <span className="text-xs text-white font-bold">!</span>
          </div>
          <div className="min-w-0">
            <h4 className="text-sm font-medium text-yellow-800 dark:text-yellow-300">
              Demo Mode
            </h4>
            <p className="text-xs text-yellow-700 dark:text-yellow-400 mt-1">
              This is a preview interface with mock data. All security features are simulated for demonstration purposes.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecureEmailPage; 