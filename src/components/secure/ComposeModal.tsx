import React, { useState } from 'react';
import { 
  X, 
  Send, 
  Paperclip, 
  Lock, 
  Globe, 
  Clock, 
  Eye, 
  AlertTriangle,
  Shield,
  FileText,
  Copy,
  EyeOff,
  Calendar,
  MapPin,
  Trash2,
  AlertCircle,
  CheckCircle,
  XCircle
} from 'lucide-react';

/**
 * Compose Form Data Interface
 * 
 * Defines the structure of the email composition form data.
 * Includes all form fields and comprehensive security settings.
 */
interface ComposeFormData {
  /** Recipient email address */
  recipient: string;
  
  /** Email subject line */
  subject: string;
  
  /** Email body content */
  body: string;
  
  /** List of file attachments */
  attachments: File[];
  
  /** Comprehensive security settings for the email */
  securitySettings: {
    /** Enable password protection */
    passwordProtection: boolean;
    
    /** Password for protected emails */
    password?: string;
    
    /** Enable geolocation-based restrictions */
    geolocationLock: boolean;
    
    /** List of allowed countries */
    allowedCountries: string[];
    
    /** Enable time-based access restrictions */
    timeLock: boolean;
    
    /** Date/time when email becomes accessible */
    unlockAfter?: string;
    
    /** Enable auto-destruct after viewing */
    autoDestruct: boolean;
    
    /** Number of views before auto-destruction */
    destructAfterViews?: number;
    
    /** Enable read-once mode */
    readOnce: boolean;
    
    /** Enable remote revocation capability */
    remoteRevoke: boolean;
    
    /** Enable decoy message feature */
    decoyMessage: boolean;
    
    /** Strip metadata from email */
    stripMetadata: boolean;
    
    /** Enable tamper detection alerts */
    tamperAlerts: boolean;
    
    /** Enable self-destruct after failed attempts */
    selfDestructAfterAttempts: boolean;
    
    /** Maximum failed attempts before self-destruction */
    maxFailedAttempts?: number;
  };
}

/**
 * Compose Modal Props Interface
 * 
 * Props for the ComposeModal component.
 */
interface ComposeModalProps {
  /** Whether the modal is open */
  isOpen: boolean;
  
  /** Callback to close the modal */
  onClose: () => void;
}

/**
 * ComposeModal Component
 * 
 * Modal for composing new secure emails with comprehensive security options.
 * Features modern design with all security toggles and form validation.
 * 
 * Features:
 * - Complete email composition form with recipient, subject, and body
 * - Comprehensive security settings panel with all available options
 * - Attachment handling (UI only, no actual upload logic)
 * - Form validation including password length requirements
 * - Responsive design that works on desktop and mobile
 * - Dark/light mode support
 * - Loading states and error handling
 * - Mock submission that logs form data to console
 * 
 * Security Options:
 * - Password protection with minimum 6 characters
 * - Geolocation restrictions by country
 * - Time-based access controls
 * - Auto-destruct after viewing
 * - Read-once mode
 * - Remote revocation capability
 * - Decoy message feature
 * - Metadata stripping
 * - Tamper detection alerts
 * - Self-destruct after failed attempts
 */
const ComposeModal: React.FC<ComposeModalProps> = ({ isOpen, onClose }) => {
  // Initialize form data with default values
  const [formData, setFormData] = useState<ComposeFormData>({
    recipient: '',
    subject: '',
    body: '',
    attachments: [],
    securitySettings: {
      passwordProtection: false,
      password: '',
      geolocationLock: false,
      allowedCountries: [],
      timeLock: false,
      unlockAfter: '',
      autoDestruct: false,
      destructAfterViews: 1,
      readOnce: false,
      remoteRevoke: false,
      decoyMessage: false,
      stripMetadata: true,
      tamperAlerts: true,
      selfDestructAfterAttempts: false,
      maxFailedAttempts: 3
    }
  });

  // Form state management
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  /**
   * Handle form field changes
   * @param field - The field name to update
   * @param value - The new value for the field
   */
  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  /**
   * Handle security setting changes
   * @param setting - The security setting name
   * @param value - The new value for the setting
   */
  const handleSecurityChange = (setting: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      securitySettings: {
        ...prev.securitySettings,
        [setting]: value
      }
    }));
  };

  /**
   * Handle file attachment selection
   * @param event - File input change event
   */
  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    setFormData(prev => ({
      ...prev,
      attachments: [...prev.attachments, ...files]
    }));
  };

  /**
   * Remove an attachment from the list
   * @param index - Index of the attachment to remove
   */
  const removeAttachment = (index: number) => {
    setFormData(prev => ({
      ...prev,
      attachments: prev.attachments.filter((_, i) => i !== index)
    }));
  };

  /**
   * Handle form submission
   * @param e - Form submission event
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate password length if password protection is enabled
    if (formData.securitySettings.passwordProtection && 
        (!formData.securitySettings.password || 
         formData.securitySettings.password.length < 6)) {
      alert('Password must be at least 6 characters long');
      return;
    }

    setIsSubmitting(true);
    
    try {
      // Simulate API call delay
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Log form data to console (mock submission)
      console.log('Compose Form Data:', formData);
      
      // Show success message
      alert('Secure email composed successfully! (Mock submission)');
      
      // Close modal and reset form
      handleClose();
    } catch (error) {
      console.error('Error composing email:', error);
      alert('Error composing email. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  /**
   * Handle modal close and form reset
   */
  const handleClose = () => {
    // Reset form data to initial state
    setFormData({
      recipient: '',
      subject: '',
      body: '',
      attachments: [],
      securitySettings: {
        passwordProtection: false,
        password: '',
        geolocationLock: false,
        allowedCountries: [],
        timeLock: false,
        unlockAfter: '',
        autoDestruct: false,
        destructAfterViews: 1,
        readOnce: false,
        remoteRevoke: false,
        decoyMessage: false,
        stripMetadata: true,
        tamperAlerts: true,
        selfDestructAfterAttempts: false,
        maxFailedAttempts: 3
      }
    });
    
    // Reset form state
    setIsSubmitting(false);
    setShowPassword(false);
    
    // Close modal
    onClose();
  };

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
        <div className="inline-block align-bottom bg-white dark:bg-secondary-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl sm:w-full">
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900/20 rounded-lg flex items-center justify-center">
                <Send className="w-5 h-5 text-primary-600 dark:text-primary-400" />
              </div>
              <div>
                <h3 className="text-lg font-medium text-secondary-900 dark:text-white">
                  Compose Secure Email
                </h3>
                <p className="text-sm text-secondary-600 dark:text-secondary-400">
                  Create a new encrypted message with advanced security options
                </p>
              </div>
            </div>
            <button
              onClick={handleClose}
              className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Content */}
          <form onSubmit={handleSubmit} className="px-6 py-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Main Form */}
              <div className="lg:col-span-2 space-y-6">
                {/* Recipient */}
                <div>
                  <label htmlFor="recipient" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                    Recipient Email
                  </label>
                  <input
                    id="recipient"
                    type="email"
                    placeholder="recipient@example.com"
                    value={formData.recipient}
                    onChange={(e) => handleInputChange('recipient', e.target.value)}
                    required
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white"
                  />
                </div>

                {/* Subject */}
                <div>
                  <label htmlFor="subject" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                    Subject
                  </label>
                  <input
                    id="subject"
                    type="text"
                    placeholder="Enter subject line"
                    value={formData.subject}
                    onChange={(e) => handleInputChange('subject', e.target.value)}
                    required
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white"
                  />
                </div>

                {/* Body */}
                <div>
                  <label htmlFor="body" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                    Message Body
                  </label>
                  <textarea
                    id="body"
                    rows={12}
                    placeholder="Type your secure message here..."
                    value={formData.body}
                    onChange={(e) => handleInputChange('body', e.target.value)}
                    required
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white resize-none"
                  />
                </div>

                {/* Attachments */}
                <div>
                  <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                    Attachments
                  </label>
                  <div className="border-2 border-dashed border-secondary-300 dark:border-secondary-600 rounded-lg p-6 text-center">
                    <Paperclip className="w-8 h-8 text-secondary-400 mx-auto mb-2" />
                    <p className="text-sm text-secondary-600 dark:text-secondary-400 mb-2">
                      Drag and drop files here, or click to browse
                    </p>
                    <input
                      type="file"
                      multiple
                      onChange={handleFileSelect}
                      className="hidden"
                      id="file-upload"
                    />
                    <label
                      htmlFor="file-upload"
                      className="inline-flex items-center px-4 py-2 bg-secondary-100 dark:bg-secondary-700 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-600 transition-colors duration-200 cursor-pointer"
                    >
                      Choose Files
                    </label>
                  </div>

                  {/* Attachment List */}
                  {formData.attachments.length > 0 && (
                    <div className="mt-4 space-y-2">
                      {formData.attachments.map((file, index) => (
                        <div key={index} className="flex items-center justify-between p-3 bg-secondary-50 dark:bg-secondary-700 rounded-lg">
                          <div className="flex items-center space-x-3 min-w-0">
                            <FileText className="w-4 h-4 text-secondary-600 dark:text-secondary-400 flex-shrink-0" />
                            <div className="min-w-0 flex-1">
                              <p className="text-sm font-medium text-secondary-900 dark:text-white truncate">
                                {file.name}
                              </p>
                              <p className="text-xs text-secondary-500 dark:text-secondary-400">
                                {(file.size / 1024 / 1024).toFixed(2)} MB
                              </p>
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() => removeAttachment(index)}
                            className="p-1 text-secondary-400 hover:text-red-600 dark:hover:text-red-400"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Security Settings */}
              <div className="lg:col-span-1">
                <div className="bg-secondary-50 dark:bg-secondary-700 rounded-lg p-4">
                  <h4 className="text-lg font-medium text-secondary-900 dark:text-white mb-4">
                    Security Settings
                  </h4>
                  
                  <div className="space-y-4">
                    {/* Password Protection */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Lock className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Password Protection</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.passwordProtection}
                          onChange={(e) => handleSecurityChange('passwordProtection', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Password Input */}
                    {formData.securitySettings.passwordProtection && (
                      <div className="ml-6">
                        <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                          Password
                        </label>
                                                 <div className="relative">
                           <input
                             type={showPassword ? 'text' : 'password'}
                             placeholder="Enter password (min. 6 characters)"
                             value={formData.securitySettings.password}
                             onChange={(e) => handleSecurityChange('password', e.target.value)}
                             minLength={6}
                             className="w-full px-3 py-2 pr-10 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                           />
                           <button
                             type="button"
                             onClick={() => setShowPassword(!showPassword)}
                             className="absolute right-3 top-1/2 transform -translate-y-1/2 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-200"
                           >
                             {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                           </button>
                         </div>
                      </div>
                    )}

                    {/* Geolocation Lock */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Globe className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Geolocation Lock</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.geolocationLock}
                          onChange={(e) => handleSecurityChange('geolocationLock', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Time Lock */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Clock className="w-4 h-4 text-green-600 dark:text-green-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Time Lock</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.timeLock}
                          onChange={(e) => handleSecurityChange('timeLock', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Time Lock Input */}
                    {formData.securitySettings.timeLock && (
                      <div className="ml-6">
                        <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                          Unlock After
                        </label>
                        <input
                          type="datetime-local"
                          value={formData.securitySettings.unlockAfter}
                          onChange={(e) => handleSecurityChange('unlockAfter', e.target.value)}
                          className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                        />
                      </div>
                    )}

                    {/* Auto-Destruct */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <AlertTriangle className="w-4 h-4 text-orange-600 dark:text-orange-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Auto-Destruct</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.autoDestruct}
                          onChange={(e) => handleSecurityChange('autoDestruct', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Auto-Destruct Input */}
                    {formData.securitySettings.autoDestruct && (
                      <div className="ml-6">
                        <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                          Destruct After Views
                        </label>
                        <input
                          type="number"
                          min="1"
                          value={formData.securitySettings.destructAfterViews}
                          onChange={(e) => handleSecurityChange('destructAfterViews', parseInt(e.target.value))}
                          className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                        />
                      </div>
                    )}

                    {/* Read Once */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Eye className="w-4 h-4 text-red-600 dark:text-red-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Read Once</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.readOnce}
                          onChange={(e) => handleSecurityChange('readOnce', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Remote Revoke */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Shield className="w-4 h-4 text-purple-600 dark:text-purple-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Remote Revoke</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.remoteRevoke}
                          onChange={(e) => handleSecurityChange('remoteRevoke', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Decoy Message */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <AlertCircle className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Decoy Message</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.decoyMessage}
                          onChange={(e) => handleSecurityChange('decoyMessage', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                    {/* Strip Metadata */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Copy className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Strip Metadata</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.securitySettings.stripMetadata}
                          onChange={(e) => handleSecurityChange('stripMetadata', e.target.checked)}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                      </label>
                    </div>

                                         {/* Tamper Alerts */}
                     <div className="flex items-center justify-between">
                       <div className="flex items-center space-x-2">
                         <AlertTriangle className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
                         <span className="text-sm text-secondary-700 dark:text-secondary-300">Tamper Alerts</span>
                       </div>
                       <label className="relative inline-flex items-center cursor-pointer">
                         <input
                           type="checkbox"
                           checked={formData.securitySettings.tamperAlerts}
                           onChange={(e) => handleSecurityChange('tamperAlerts', e.target.checked)}
                           className="sr-only peer"
                         />
                         <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                       </label>
                     </div>

                     {/* Self-Destruct After Failed Attempts */}
                     <div className="flex items-center justify-between">
                       <div className="flex items-center space-x-2">
                         <Trash2 className="w-4 h-4 text-red-600 dark:text-red-400" />
                         <span className="text-sm text-secondary-700 dark:text-secondary-300">Self-destruct after failed attempts</span>
                       </div>
                       <label className="relative inline-flex items-center cursor-pointer">
                         <input
                           type="checkbox"
                           checked={formData.securitySettings.selfDestructAfterAttempts}
                           onChange={(e) => handleSecurityChange('selfDestructAfterAttempts', e.target.checked)}
                           className="sr-only peer"
                         />
                         <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                       </label>
                     </div>

                     {/* Failed Attempts Input */}
                     {formData.securitySettings.selfDestructAfterAttempts && (
                       <div className="ml-6">
                         <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                           Max Failed Attempts
                         </label>
                         <input
                           type="number"
                           min="1"
                           max="10"
                           value={formData.securitySettings.maxFailedAttempts}
                           onChange={(e) => handleSecurityChange('maxFailedAttempts', parseInt(e.target.value))}
                           className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                         />
                       </div>
                     )}
                  </div>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end space-x-3 pt-6 border-t border-secondary-200 dark:border-secondary-700">
              <button
                type="button"
                onClick={handleClose}
                disabled={isSubmitting}
                className="px-4 py-2 border border-secondary-300 dark:border-secondary-600 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cancel
              </button>
                             <button
                 type="submit"
                 disabled={
                   isSubmitting || 
                   !formData.recipient || 
                   !formData.subject || 
                   !formData.body ||
                   (formData.securitySettings.passwordProtection && (!formData.securitySettings.password || formData.securitySettings.password.length < 6))
                 }
                 className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
               >
                {isSubmitting ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                    <span>Sending...</span>
                  </>
                ) : (
                  <>
                    <Send className="w-4 h-4" />
                    <span>Send Securely</span>
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default ComposeModal;
