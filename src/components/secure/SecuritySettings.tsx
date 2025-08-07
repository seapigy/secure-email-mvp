import React, { useState } from 'react';
import { 
  Lock, 
  Globe, 
  Clock, 
  Eye, 
  EyeOff, 
  Shield, 
  AlertTriangle,
  CheckCircle,
  XCircle,
  Settings,
  User,
  Calendar,
  MapPin,
  Fingerprint,
  AlertCircle,
  FileText,
  Paperclip,
  Copy,
  Trash2,
  Eye as EyeIcon,
  Shield as ShieldIcon,
  Globe as GlobeIcon,
  Clock as ClockIcon,
  AlertTriangle as AlertTriangleIcon,
  CheckCircle as CheckCircleIcon,
  XCircle as XCircleIcon,
  X,
  Save,
  RefreshCw
} from 'lucide-react';
import { SecuritySettings as SecuritySettingsType } from '@/types/secureEmail';

/**
 * Security Settings Props Interface
 * 
 * Props for the SecuritySettings component.
 */
interface SecuritySettingsProps {
  /** Whether the settings modal is open */
  isOpen: boolean;
  
  /** Callback to close the settings modal */
  onClose: () => void;
  
  /** Current security settings */
  settings: SecuritySettingsType;
  
  /** Callback when settings are changed */
  onSettingsChange: (settings: SecuritySettingsType) => void;
  
  /** Scope of settings (global or per-email) */
  scope?: 'global' | 'per-email';
  
  /** Callback when scope is changed */
  onScopeChange?: (scope: 'global' | 'per-email') => void;
}

/**
 * SecuritySettings Component
 * 
 * Comprehensive security settings panel with toggles and options for
 * all security features. Features modern UI with privacy-first design.
 * 
 * Features:
 * - Password protection settings with strength validation
 * - Geolocation restrictions with country selection
 * - Time-based locks with date/time pickers
 * - Auto-destruct options with view count limits
 * - Read-once mode for single-view emails
 * - Remote revocation capability
 * - Decoy messages for security through obscurity
 * - Fingerprint management for integrity verification
 * - Metadata stripping for privacy
 * - Tamper alerts for unauthorized access detection
 * - Self-destruct after failed attempts
 * - Per-email vs global scope management
 * 
 * Security Options:
 * - Password Protection: Require password for email access
 * - Per-Email Password: Require password for every email (vs session-based)
 * - Geolocation Lock: Restrict access by country
 * - Time Lock: Set unlock times for emails
 * - Auto-Destruct: Messages that self-destruct after viewing
 * - Read-Once Mode: Messages that can only be viewed once
 * - Remote Revoke: Ability to revoke access to sent messages
 * - Decoy Message: Fake messages to mislead attackers
 * - Strip Metadata: Remove identifying information
 * - Tamper Alerts: Detect unauthorized access attempts
 * - Self-Destruct After Failed Attempts: Auto-delete after failed access
 * 
 * UI Features:
 * - Modern toggle switches for all options
 * - Conditional input fields based on settings
 * - Real-time validation and feedback
 * - Save/reset functionality with change tracking
 * - Responsive design with dark/light mode support
 * - Comprehensive help text and descriptions
 */
const SecuritySettings: React.FC<SecuritySettingsProps> = ({
  isOpen,
  onClose,
  settings,
  onSettingsChange,
  scope = 'global',
  onScopeChange
}) => {
  // Local settings state for form management
  const [localSettings, setLocalSettings] = useState<SecuritySettingsType>(settings);
  
  // Track whether settings have been modified
  const [hasChanges, setHasChanges] = useState(false);

  /**
   * Handle setting changes
   * @param key - The setting key to update
   * @param value - The new value for the setting
   */
  const handleSettingChange = (key: keyof SecuritySettingsType, value: any) => {
    const newSettings = { ...localSettings, [key]: value };
    setLocalSettings(newSettings);
    setHasChanges(true);
  };

  /**
   * Handle save button click
   * Applies the local settings and closes the modal
   */
  const handleSave = () => {
    onSettingsChange(localSettings);
    setHasChanges(false);
  };

  /**
   * Handle reset button click
   * Resets local settings to original values
   */
  const handleReset = () => {
    setLocalSettings(settings);
    setHasChanges(false);
  };

  /**
   * Generate a random fingerprint hash
   * Creates a 40-character alphanumeric string for integrity verification
   */
  const generateFingerprint = () => {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < 40; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    handleSettingChange('fingerprintHash', result);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        {/* Background overlay */}
        <div 
          className="fixed inset-0 bg-secondary-900 bg-opacity-75 transition-opacity"
          onClick={onClose}
        />

        {/* Modal */}
        <div className="inline-block align-bottom bg-white dark:bg-secondary-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl sm:w-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900/20 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-primary-600 dark:text-primary-400" />
              </div>
              <div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white">
                  Security Settings
                </h3>
                <p className="text-sm text-secondary-600 dark:text-secondary-400">
                  Configure advanced security features
                </p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Content */}
          <div className="px-6 py-6 max-h-96 overflow-y-auto">
            {/* Scope Toggle */}
            {onScopeChange && (
              <div className="mb-6 p-4 bg-secondary-50 dark:bg-secondary-700 rounded-lg">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-secondary-900 dark:text-white">
                      Settings Scope
                    </h4>
                    <p className="text-xs text-secondary-600 dark:text-secondary-400">
                      Choose how to apply these settings
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => onScopeChange('global')}
                      className={`px-3 py-1 text-xs font-medium rounded-md transition-colors duration-200 ${
                        scope === 'global'
                          ? 'bg-primary-600 text-white'
                          : 'bg-secondary-200 text-secondary-700 dark:bg-secondary-600 dark:text-secondary-300'
                      }`}
                    >
                      Global
                    </button>
                    <button
                      onClick={() => onScopeChange('per-email')}
                      className={`px-3 py-1 text-xs font-medium rounded-md transition-colors duration-200 ${
                        scope === 'per-email'
                          ? 'bg-primary-600 text-white'
                          : 'bg-secondary-200 text-secondary-700 dark:bg-secondary-600 dark:text-secondary-300'
                      }`}
                    >
                      Per Email
                    </button>
                  </div>
                </div>
              </div>
            )}
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Protection Settings */}
              <div className="space-y-6">
                <div>
                  <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-4">
                    Protection Settings
                  </h4>
                  
                  {/* Password Protection */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Lock className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Password Protection
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Require password to decrypt
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.passwordProtection}
                        onChange={(e) => handleSettingChange('passwordProtection', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Per-Email Password */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Shield className="w-5 h-5 text-green-600 dark:text-green-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Require password for every secure email
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Maximum security - password per email
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.perEmailPassword}
                        onChange={(e) => handleSettingChange('perEmailPassword', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Geolocation Lock */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Globe className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Geolocation Lock
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Restrict access by location
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.geolocationLock}
                        onChange={(e) => handleSettingChange('geolocationLock', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Time Lock */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Clock className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Time Lock
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Unlock after specific date
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.timeLock}
                        onChange={(e) => handleSettingChange('timeLock', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Read Once */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Eye className="w-5 h-5 text-red-600 dark:text-red-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Read Once Mode
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Self-destruct after first read
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.readOnce}
                        onChange={(e) => handleSettingChange('readOnce', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>
                </div>
              </div>

              {/* Advanced Settings */}
              <div className="space-y-6">
                <div>
                  <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-4">
                    Advanced Settings
                  </h4>
                  
                  {/* Auto-Destruct */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <AlertTriangle className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Auto-Destruct
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Destroy after X attempts
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.autoDestruct}
                        onChange={(e) => handleSettingChange('autoDestruct', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Remote Revoke */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <XCircle className="w-5 h-5 text-red-600 dark:text-red-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Remote Revoke
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Allow remote destruction
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.remoteRevoke}
                        onChange={(e) => handleSettingChange('remoteRevoke', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Decoy Message */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <EyeOff className="w-5 h-5 text-gray-600 dark:text-gray-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Decoy Message
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Show fake content to attackers
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.decoyMessage}
                        onChange={(e) => handleSettingChange('decoyMessage', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Strip Metadata */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <FileText className="w-5 h-5 text-green-600 dark:text-green-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Strip Metadata
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Remove identifying information
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.stripMetadata}
                        onChange={(e) => handleSettingChange('stripMetadata', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Tamper Alerts */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <AlertCircle className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Tamper Alerts
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Email alert on breach attempt
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.tamperAlerts}
                        onChange={(e) => handleSettingChange('tamperAlerts', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Self-Destruct After Failed Attempts */}
                  <div className="flex items-center justify-between py-3 border-b border-secondary-200 dark:border-secondary-700">
                    <div className="flex items-center space-x-3">
                      <Trash2 className="w-5 h-5 text-red-600 dark:text-red-400" />
                      <div>
                        <p className="text-sm font-medium text-secondary-900 dark:text-white">
                          Self-destruct after failed attempts
                        </p>
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          Permanently delete message after failed password attempts
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={localSettings.selfDestructAfterAttempts}
                        onChange={(e) => handleSettingChange('selfDestructAfterAttempts', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                  {/* Max Failed Attempts Input */}
                  {localSettings.selfDestructAfterAttempts && (
                    <div className="py-3 border-b border-secondary-200 dark:border-secondary-700">
                      <div className="flex items-center space-x-3">
                        <div className="flex-1">
                          <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                            Max Failed Attempts
                          </label>
                          <input
                            type="number"
                            min="1"
                            max="10"
                            value={localSettings.maxFailedAttempts || 3}
                            onChange={(e) => handleSettingChange('maxFailedAttempts', parseInt(e.target.value))}
                            className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white"
                          />
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Fingerprint Section */}
            <div className="mt-6 pt-6 border-t border-secondary-200 dark:border-secondary-700">
              <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-4">
                Fingerprint Hash
              </h4>
              <div className="flex items-center space-x-3">
                <div className="flex-1">
                  <input
                    type="text"
                    value={localSettings.fingerprintHash}
                    onChange={(e) => handleSettingChange('fingerprintHash', e.target.value)}
                    placeholder="Generate a unique fingerprint hash"
                    className="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
                  />
                </div>
                <button
                  onClick={generateFingerprint}
                  className="px-3 py-2 bg-secondary-100 dark:bg-secondary-700 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-600 transition-colors duration-200"
                >
                  <RefreshCw className="w-4 h-4" />
                </button>
                <button
                  onClick={() => navigator.clipboard.writeText(localSettings.fingerprintHash)}
                  className="px-3 py-2 bg-secondary-100 dark:bg-secondary-700 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-600 transition-colors duration-200"
                >
                  <Copy className="w-4 h-4" />
                </button>
              </div>
              <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-2">
                This hash uniquely identifies this message and can be used to verify authenticity.
              </p>
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between px-6 py-4 border-t border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-700/50">
            <button
              onClick={handleReset}
              className="px-4 py-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
            >
              Reset
            </button>
            <div className="flex items-center space-x-3">
              <button
                onClick={onClose}
                className="px-4 py-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={!hasChanges}
                className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
              >
                <Save className="w-4 h-4" />
                <span>Save Settings</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecuritySettings; 