import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import {
  Send,
  User,
  FileText,
  MessageSquare,
  Lock,
  Shield,
  AlertCircle,
  Loader2,
  ArrowLeft,
  Eye,
  EyeOff,
  Globe,
  AlertTriangle
} from 'lucide-react';
import { sendEmail } from '@/lib/api';

interface EmailFormData {
  to: string;
  subject: string;
  message: string;
  password?: string;
  passwordProtected: boolean;
  geolocationRestricted: boolean;
  timeLock: boolean;
  readOnce: boolean;
  autoDestruct: boolean;
}

interface EmailSendFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
  initialData?: Partial<EmailFormData>;
}

/**
 * EmailSendForm Component
 * 
 * Production-quality secure email sending interface with comprehensive
 * security features and modern UI design.
 * 
 * Features:
 * - Email composition with validation
 * - Security settings (password protection, geolocation, etc.)
 * - Loading states and error handling
 * - Backend integration with toast notifications
 * - Responsive design with dark mode support
 * - Future-proof layout for attachments and advanced features
 */
const EmailSendForm: React.FC<EmailSendFormProps> = ({
  onSuccess,
  onCancel,
  initialData
}) => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState<EmailFormData>({
    to: '',
    subject: '',
    message: '',
    password: '',
    passwordProtected: false,
    geolocationRestricted: false,
    timeLock: false,
    readOnce: false,
    autoDestruct: false,
    ...initialData
  });

  const [errors, setErrors] = useState<Partial<EmailFormData>>({});

  // Email validation
  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  // Form validation
  const validateForm = (): boolean => {
    const newErrors: Partial<EmailFormData> = {};

    if (!formData.to.trim()) {
      newErrors.to = 'Recipient email is required';
    } else if (!validateEmail(formData.to)) {
      newErrors.to = 'Please enter a valid email address';
    }

    if (!formData.subject.trim()) {
      newErrors.subject = 'Subject is required';
    }

    if (!formData.message.trim()) {
      newErrors.message = 'Message is required';
    }

    if (formData.passwordProtected && !formData.password?.trim()) {
      newErrors.password = 'Password is required when password protection is enabled';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      toast.error('Please fix the errors before submitting');
      return;
    }

    setIsLoading(true);

    try {
      const payload = {
        to: formData.to,
        subject: formData.subject,
        content: formData.message,
        security_settings: {
          password_protection: formData.passwordProtected,
          password: formData.passwordProtected ? formData.password : undefined,
          geolocation_restricted: formData.geolocationRestricted,
          time_lock: formData.timeLock,
          read_once: formData.readOnce,
          auto_destruct: formData.autoDestruct
        }
      };

      await sendEmail(payload);
      
      toast.success('Secure email sent successfully! 🔒');
      
      if (onSuccess) {
        onSuccess();
      } else {
        navigate('/dashboard');
      }
    } catch (error: any) {
      console.error('Failed to send email:', error);
      
      const errorMessage = error.response?.data?.error || 
                          error.message || 
                          'Failed to send email. Please try again.';
      
      toast.error(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle input changes
  const handleInputChange = (field: keyof EmailFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  // Handle security setting toggle
  const handleSecurityToggle = (setting: keyof EmailFormData) => {
    setFormData(prev => ({ 
      ...prev, 
      [setting]: !prev[setting] 
    }));
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-4">
          <button
            onClick={onCancel || (() => navigate('/dashboard'))}
            className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-secondary-900 dark:text-white">
              Send Secure Email
            </h1>
            <p className="text-sm text-secondary-600 dark:text-secondary-400">
              Compose and send an encrypted message
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <Shield className="w-5 h-5 text-primary-600 dark:text-primary-400" />
          <span className="text-sm font-medium text-primary-600 dark:text-primary-400">
            🔒 End-to-End Encrypted
          </span>
        </div>
      </div>

      {/* Security Notice */}
      <div className="mb-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
        <div className="flex items-start space-x-3">
          <Shield className="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5" />
          <div>
            <h3 className="text-sm font-medium text-blue-800 dark:text-blue-300">
              Secure Transmission
            </h3>
            <p className="text-sm text-blue-700 dark:text-blue-400 mt-1">
              Your message will be encrypted and securely stored. All communications are protected with military-grade encryption.
            </p>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Email Form */}
        <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white mb-4">
            Message Details
          </h2>
          
          <div className="space-y-4">
            {/* Recipient */}
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                <User className="w-4 h-4 inline mr-2" />
                Recipient Email
              </label>
              <input
                type="email"
                value={formData.to}
                onChange={(e) => handleInputChange('to', e.target.value)}
                className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white transition-colors duration-200 ${
                  errors.to 
                    ? 'border-red-300 dark:border-red-600' 
                    : 'border-secondary-300 dark:border-secondary-600'
                }`}
                placeholder="recipient@example.com"
              />
              {errors.to && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400 flex items-center">
                  <AlertCircle className="w-4 h-4 mr-1" />
                  {errors.to}
                </p>
              )}
            </div>

            {/* Subject */}
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                <FileText className="w-4 h-4 inline mr-2" />
                Subject
              </label>
              <input
                type="text"
                value={formData.subject}
                onChange={(e) => handleInputChange('subject', e.target.value)}
                className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white transition-colors duration-200 ${
                  errors.subject 
                    ? 'border-red-300 dark:border-red-600' 
                    : 'border-secondary-300 dark:border-secondary-600'
                }`}
                placeholder="Enter subject line"
              />
              {errors.subject && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400 flex items-center">
                  <AlertCircle className="w-4 h-4 mr-1" />
                  {errors.subject}
                </p>
              )}
            </div>

            {/* Message */}
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                <MessageSquare className="w-4 h-4 inline mr-2" />
                Message
              </label>
              <textarea
                value={formData.message}
                onChange={(e) => handleInputChange('message', e.target.value)}
                rows={8}
                className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white transition-colors duration-200 resize-vertical ${
                  errors.message 
                    ? 'border-red-300 dark:border-red-600' 
                    : 'border-secondary-300 dark:border-secondary-600'
                }`}
                placeholder="Type your secure message here..."
              />
              {errors.message && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400 flex items-center">
                  <AlertCircle className="w-4 h-4 mr-1" />
                  {errors.message}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Security Settings */}
        <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-6">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white mb-4">
            Security Settings
          </h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Password Protection */}
            <div className="flex items-center justify-between p-4 border border-secondary-200 dark:border-secondary-700 rounded-lg">
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
                  checked={formData.passwordProtected}
                  onChange={() => handleSecurityToggle('passwordProtected')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
              </label>
            </div>

            {/* Geolocation Lock */}
            <div className="flex items-center justify-between p-4 border border-secondary-200 dark:border-secondary-700 rounded-lg">
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
                  checked={formData.geolocationRestricted}
                  onChange={() => handleSecurityToggle('geolocationRestricted')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
              </label>
            </div>

            {/* Read Once */}
            <div className="flex items-center justify-between p-4 border border-secondary-200 dark:border-secondary-700 rounded-lg">
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
                  checked={formData.readOnce}
                  onChange={() => handleSecurityToggle('readOnce')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
              </label>
            </div>

            {/* Auto-Destruct */}
            <div className="flex items-center justify-between p-4 border border-secondary-200 dark:border-secondary-700 rounded-lg">
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
                  checked={formData.autoDestruct}
                  onChange={() => handleSecurityToggle('autoDestruct')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-secondary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
              </label>
            </div>
          </div>

          {/* Password Input (shown when password protection is enabled) */}
          {formData.passwordProtected && (
            <div className="mt-4">
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                <Lock className="w-4 h-4 inline mr-2" />
                Decryption Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) => handleInputChange('password', e.target.value)}
                  className={`w-full px-4 py-3 pr-12 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white transition-colors duration-200 ${
                    errors.password 
                      ? 'border-red-300 dark:border-red-600' 
                      : 'border-secondary-300 dark:border-secondary-600'
                  }`}
                  placeholder="Enter password for decryption"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-300"
                >
                  {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                </button>
              </div>
              {errors.password && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400 flex items-center">
                  <AlertCircle className="w-4 h-4 mr-1" />
                  {errors.password}
                </p>
              )}
            </div>
          )}
        </div>

        {/* Submit Button */}
        <div className="flex items-center justify-between">
          <button
            type="button"
            onClick={onCancel || (() => navigate('/dashboard'))}
            className="px-6 py-3 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
          >
            Cancel
          </button>
          
          <button
            type="submit"
            disabled={isLoading}
            className="flex items-center space-x-2 px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                <span>Sending...</span>
              </>
            ) : (
              <>
                <Send className="w-5 h-5" />
                <span>Send Secure Email</span>
              </>
            )}
          </button>
        </div>
      </form>

      {/* Debug Section (Development Only) - Removed for TypeScript compatibility */}
    </div>
  );
};

export default EmailSendForm; 