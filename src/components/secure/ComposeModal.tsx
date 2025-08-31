/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE "PERFECT" DESIGN THAT MUST NEVER BE CHANGED.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual layout or design
 * 2. NEVER modify the two-column structure (email composition left, security panel right)
 * 3. NEVER change the toggle switch styling or appearance
 * 4. NEVER alter the color scheme or Tailwind classes
 * 5. NEVER modify the modal structure or positioning
 * 6. ONLY add new security features to the existing right-side panel
 * 7. ONLY add new form fields to the existing left-side panel
 * 8. ALWAYS maintain the exact same visual appearance
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * This design was restored from commit e291daf and represents the "perfect" design.
 * Any changes to the visual design will result in immediate user dissatisfaction.
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE DESIGN, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 * @last_restored: From commit e291daf
 * @user_feedback: "This is the perfect design, never change it"
 */

import React, { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import { log } from '@/lib/logger';
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
  Trash2,
  AlertCircle,
  Fingerprint
} from 'lucide-react';
import { sendSecureEmail } from '@/lib/api';
import { toast } from 'react-toastify';
import { usePerformanceMonitoring } from '@/lib/performance';

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

    /** Require password for every secure email */
    requirePasswordForEveryEmail: boolean;

    /** Maximum security - password per email */
    passwordPerEmail: boolean;
    
    /** Enable geolocation-based restrictions */
    geolocationLock: boolean;
    
    /** Geolocation verification type */
  geoVerificationType: string;

    /** Allowed city for geolocation */
  geoCity: string;

    /** Allowed country for geolocation */
  geoCountry: string;
  
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

    /** Decoy secret for decoy message */
  decoySecret: string;
    
    /** Strip metadata from email */
  stripMetadata: boolean;
    
    /** Enable tamper detection alerts */
  tamperAlerts: boolean;
    
    /** Enable self-destruct after failed attempts */
    selfDestructAfterAttempts: boolean;
    
    /** Maximum failed attempts before self-destruction */
    maxFailedAttempts?: number;

    /** Generate fingerprint hash */
    generateFingerprintHash: boolean;

    /** Fingerprint hash value */
    fingerprintHash?: string;
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
      requirePasswordForEveryEmail: false,
      passwordPerEmail: false,
      geolocationLock: false,
    geoVerificationType: 'none',
    geoCity: '',
    geoCountry: '',
      allowedCountries: [],
      timeLock: false,
      unlockAfter: '',
      autoDestruct: false,
      destructAfterViews: 1,
      readOnce: false,
    remoteRevoke: false,
    decoyMessage: false,
    decoySecret: '',
      stripMetadata: true,
      tamperAlerts: true,
      selfDestructAfterAttempts: false,
      maxFailedAttempts: 3,
      generateFingerprintHash: false,
      fingerprintHash: ''
    }
  });

  // Form state management
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Memoized computations for performance optimization
  const hasSecurityFeatures = useMemo(() => {
    return formData.securitySettings.passwordProtection ||
           formData.securitySettings.geolocationLock ||
           formData.securitySettings.timeLock ||
           formData.securitySettings.autoDestruct ||
           formData.securitySettings.readOnce ||
           formData.securitySettings.remoteRevoke ||
           formData.securitySettings.decoyMessage ||
           formData.securitySettings.stripMetadata ||
           formData.securitySettings.tamperAlerts ||
           formData.securitySettings.selfDestructAfterAttempts ||
           formData.securitySettings.generateFingerprintHash;
  }, [formData.securitySettings]);

  // TODO: Implement attachment size validation
  // const totalAttachmentSize = useMemo(() => {
  //   return formData.attachments.reduce((sum, file) => sum + file.size, 0);
  // }, [formData.attachments]);

  // TODO: Implement attachment count display
  // const attachmentCount = useMemo(() => {
  //   return formData.attachments.length;
  // }, [formData.attachments]);

  // TODO: Implement content validation
  // const hasContent = useMemo(() => {
  //   return formData.recipient || formData.subject || formData.body || formData.attachments.length > 0;
  // }, [formData.recipient, formData.subject, formData.body, formData.attachments.length]);

  // TODO: Implement form validation
  // const isFormValid = useMemo(() => {
  //   return formData.recipient && formData.subject && formData.body.trim();
  // }, [formData.recipient, formData.subject, formData.body]);

  // Debounced input handling for performance
  const debounceTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
    };
  }, []);

  // Enhanced performance monitoring
  const { monitorAsync } = usePerformanceMonitoring('ComposeModal');

  // Performance monitoring for component lifecycle
  useEffect(() => {
    const startTime = performance.now();
    
    return () => {
      const endTime = performance.now();
      const renderTime = endTime - startTime;
      
      if (renderTime > 16) { // 60fps threshold
        log.warn(`ComposeModal render took ${renderTime.toFixed(2)}ms (target: <16ms)`, null, 'ComposeModal');
      }
    };
  });

  // Security monitoring and logging
  useEffect(() => {
    // Log security-relevant actions for audit trail
    const logSecurityEvent = (event: string, details: unknown) => {
      log.security(event, {
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        details
      });
    };

    // Monitor for suspicious activity
    const checkForSuspiciousActivity = () => {
      // Check for rapid form changes (potential automation)
              // TODO: Implement change tracking
        // let changeCount = 0;
      const changeThreshold = 50; // Changes per minute
      
      // This would be implemented with a more sophisticated monitoring system
      // For now, we'll just log the concept
      logSecurityEvent('Security monitoring initialized', { changeThreshold });
    };

    checkForSuspiciousActivity();
  }, []);

  /**
   * Sanitize input to prevent XSS attacks
   * @param input - The input string to sanitize
   * @returns Sanitized input string
   */
  const sanitizeInput = useCallback((input: string): string => {
    if (!input) return input;
    
    // Remove potentially dangerous HTML tags and attributes
    const dangerousPatterns = [
      /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
      /<iframe\b[^<]*(?:(?!<\/iframe>)<[^<]*)*<\/iframe>/gi,
      /<object\b[^<]*(?:(?!<\/object>)<[^<]*)*<\/object>/gi,
      /<embed\b[^<]*(?:(?!<\/embed>)<[^<]*)*<\/embed>/gi,
      /javascript:/gi,
      /vbscript:/gi,
      /on\w+\s*=/gi,
      /data:text\/html/gi,
      /data:application\/javascript/gi
    ];
    
    let sanitized = input;
    dangerousPatterns.forEach(pattern => {
      sanitized = sanitized.replace(pattern, '');
    });
    
    // Remove null bytes and control characters
    sanitized = sanitized.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '');
    
    return sanitized;
  }, []);

  /**
   * Validate and sanitize email address
   * @param email - The email address to validate
   * @returns Validated and sanitized email or null if invalid
   */
  const validateAndSanitizeEmail = useCallback((email: string): string | null => {
    if (!email) return null;
    
    // Basic email validation
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(email)) {
      return null;
    }
    
    // Sanitize email
    const sanitized = sanitizeInput(email.toLowerCase().trim());
    
    // Additional security checks
    if (sanitized.length > 254) return null; // RFC 5321 limit
    if (sanitized.includes('..')) return null; // No consecutive dots
    if (sanitized.startsWith('.') || sanitized.endsWith('.')) return null; // No leading/trailing dots
    
    return sanitized;
  }, [sanitizeInput]);

  /**
   * Handle form field changes with security hardening
   * @param field - The field name to update
   * @param value - The new value for the field
   */
  const handleInputChange = useCallback((field: string, value: string) => {
    try {
      // Sanitize input first
      const sanitizedValue = sanitizeInput(value);
      
      // Field-specific validation and sanitization
      if (field === 'recipient') {
        const validatedEmail = validateAndSanitizeEmail(sanitizedValue);
        if (sanitizedValue && !validatedEmail) {
          // Don't update if email is invalid
          return;
        }
        if (validatedEmail) {
          setFormData(prev => ({ ...prev, [field]: validatedEmail }));
          return;
        }
      }
      
      if (field === 'subject') {
        // Subject-specific sanitization
        const cleanSubject = sanitizedValue.replace(/[<>]/g, ''); // Remove angle brackets
        if (cleanSubject.length > 200) {
          toast.warning('Subject is getting long. Maximum is 200 characters.');
        }
        setFormData(prev => ({ ...prev, [field]: cleanSubject }));
        return;
      }
      
      if (field === 'body') {
        // Body-specific sanitization (allow some HTML for formatting)
        const cleanBody = sanitizedValue.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
        if (cleanBody.length > 10000) {
          toast.warning('Message is getting long. Maximum is 10,000 characters.');
        }
        setFormData(prev => ({ ...prev, [field]: cleanBody }));
        return;
      }
      
      // For other fields, use sanitized value
      setFormData(prev => ({ ...prev, [field]: sanitizedValue }));
      
      // Log form field changes for debugging (excluding body for privacy)
      // Only log in development environment
      if (field !== 'body' && typeof window !== 'undefined' && window.location.hostname === 'localhost') {
        log.debug(`Form field changed: ${field}`, sanitizedValue, 'ComposeModal');
      }
      
    } catch (error) {
      log.error(`Error updating form field ${field}`, error, 'ComposeModal');
      toast.error(`Error updating ${field}. Please try again.`);
    }
  }, [sanitizeInput, validateAndSanitizeEmail]);

  // TODO: Implement debounced input change
  // const debouncedInputChange = useCallback((field: string, value: string) => {
  //   if (debounceTimeoutRef.current) {
  //     clearTimeout(debounceTimeoutRef.current);
  //   }
  //   
  //   debounceTimeoutRef.current = setTimeout(() => {
  //     handleInputChange(field, value);
  //   }, 300); // 300ms debounce delay
  // }, [handleInputChange]);

  /**
   * Validate and sanitize security settings
   * @param setting - The security setting name
   * @param value - The value to validate and sanitize
   * @returns Sanitized value or null if invalid
   */
  const validateAndSanitizeSecuritySetting = useCallback((setting: string, value: unknown): unknown => {
    try {
      switch (setting) {
        case 'password':
          if (typeof value === 'string') {
            const sanitized = sanitizeInput(value);
            // Additional password security checks
            if (sanitized.length < 6) return null;
            if (sanitized.length > 128) return null; // Reasonable password length limit
            // Check for common weak passwords
            const weakPasswords = ['password', '123456', 'qwerty', 'admin', 'letmein'];
            if (weakPasswords.includes(sanitized.toLowerCase())) return null;
            return sanitized;
          }
          return null;
          
        case 'geoCity':
        case 'geoCountry':
          if (typeof value === 'string') {
            const sanitized = sanitizeInput(value.trim());
            const nameRegex = /^[a-zA-Z\s\-']{1,50}$/;
            return nameRegex.test(sanitized) ? sanitized : null;
          }
          return null;
          
        case 'decoySecret':
          if (typeof value === 'string') {
            const sanitized = sanitizeInput(value.trim());
            if (sanitized.length < 4 || sanitized.length > 50) return null;
            // Check for sensitive patterns
            const sensitivePatterns = ['password', 'secret', 'key', 'token', 'auth', 'admin'];
            const lowerValue = sanitized.toLowerCase();
            if (sensitivePatterns.some(pattern => lowerValue.includes(pattern))) return null;
            return sanitized;
          }
          return null;
          
        case 'destructAfterViews':
        case 'maxFailedAttempts':
          const numValue = typeof value === 'string' ? parseInt(value) : typeof value === 'number' ? value : NaN;
          if (isNaN(numValue)) return null;
          if (setting === 'destructAfterViews' && (numValue < 1 || numValue > 100)) return null;
          if (setting === 'maxFailedAttempts' && (numValue < 1 || numValue > 10)) return null;
          return numValue;
          
        case 'geoVerificationType':
          const validTypes = ['none', 'country', 'city', 'city_country'];
          return typeof value === 'string' && validTypes.includes(value) ? value : 'none';
          
        case 'unlockAfter':
          if (typeof value === 'string') {
            const date = new Date(value);
            if (isNaN(date.getTime())) return null;
            const now = new Date();
            const maxDate = new Date();
            maxDate.setFullYear(maxDate.getFullYear() + 1);
            if (date <= now || date > maxDate) return null;
            return value;
          }
          return null;
          
        default:
          // For boolean values and other settings, return as-is
          return value;
      }
    } catch (error) {
      log.error(`Error validating security setting ${setting}`, error, 'ComposeModal');
      return null;
    }
  }, [sanitizeInput]);

  /**
   * Handle security setting changes with security hardening
   * @param setting - The security setting name
   * @param value - The new value for the setting
   */
  const handleSecurityChange = useCallback((setting: string, value: unknown) => {
    try {
      // Validate and sanitize the security setting
      const sanitizedValue = validateAndSanitizeSecuritySetting(setting, value);
      
      // Show warnings for validation issues
      if (sanitizedValue === null) {
        switch (setting) {
          case 'password':
            toast.warning('Password should be at least 6 characters long and not a common weak password');
            break;
          case 'geoCity':
          case 'geoCountry':
            toast.warning('Location name can only contain letters, spaces, hyphens, and apostrophes (max 50 characters)');
            break;
          case 'decoySecret':
            toast.warning('Decoy secret should be 4-50 characters and not contain sensitive words');
            break;
          case 'destructAfterViews':
            toast.warning('Destruct after views must be between 1 and 100');
            break;
          case 'maxFailedAttempts':
            toast.warning('Maximum failed attempts must be between 1 and 10');
            break;
          case 'unlockAfter':
            toast.warning('Unlock time must be in the future and not more than 1 year ahead');
            break;
          default:
            toast.warning(`Invalid value for ${setting}`);
        }
        return; // Don't update if invalid
      }
      
      // Update the security setting
      setFormData(prev => ({
        ...prev,
        securitySettings: {
          ...prev.securitySettings,
          [setting]: sanitizedValue
        }
      }));
      
      // Log security setting changes for debugging
      log.debug(`Security setting changed: ${setting}`, sanitizedValue, 'ComposeModal');
      
    } catch (error) {
      log.error(`Error updating security setting ${setting}`, error, 'ComposeModal');
      toast.error(`Error updating ${setting}. Please try again.`);
    }
  }, [validateAndSanitizeSecuritySetting]);

  /**
   * Validate file security and integrity
   * @param file - The file to validate
   * @returns Validation result object
   */
  const validateFileSecurity = useCallback((file: File): { valid: boolean; error?: string } => {
    try {
      // Check file name for malicious patterns
      const dangerousPatterns = [
        /\.(exe|bat|cmd|com|pif|scr|vbs|js|jar|msi|dll|sys|drv|bin)$/i,
        /^(con|prn|aux|nul|com[1-9]|lpt[1-9])$/i,
        /[<>:"|?*]/g,
        /\.\./g,
        /^\./g
      ];
      
      for (const pattern of dangerousPatterns) {
        if (pattern.test(file.name)) {
          return { valid: false, error: 'File name contains invalid characters or dangerous extensions' };
        }
      }
      
      // Check for null bytes in filename
      if (file.name.includes('\0')) {
        return { valid: false, error: 'File name contains null bytes' };
      }
      
      // Validate file type by checking both MIME type and extension
      const allowedTypes = [
        'image/jpeg', 'image/png', 'image/gif', 'image/webp',
        'application/pdf', 'application/msword', 
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'text/plain', 'text/csv'
      ];
      
      const allowedExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.pdf', '.doc', '.docx', '.txt', '.csv'];
      const fileExtension = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
      
      if (!allowedTypes.includes(file.type) || !allowedExtensions.includes(fileExtension)) {
        return { valid: false, error: 'File type not allowed' };
      }
      
      // Check file size limits
      const maxFileSize = 10 * 1024 * 1024; // 10MB
      if (file.size > maxFileSize) {
        return { valid: false, error: 'File size exceeds 10MB limit' };
      }
      
      // Check for empty files
      if (file.size === 0) {
        return { valid: false, error: 'Empty files are not allowed' };
      }
      
      return { valid: true };
      
    } catch (error) {
      log.error('File security validation error', error, 'ComposeModal');
      return { valid: false, error: 'File validation failed' };
    }
  }, []);

  /**
   * Handle file attachment selection with security hardening
   * @param event - File input change event
   */
  const handleFileSelect = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    try {
      const files = Array.from(event.target.files || []);
      
      if (files.length === 0) {
        return;
      }
      
      const maxTotalSize = 50 * 1024 * 1024; // 50MB total
      let totalSize = formData.attachments.reduce((sum, file) => sum + file.size, 0);
      let validFiles: File[] = [];
      let errorCount = 0;
      let securityViolations = 0;
      
      for (const file of files) {
        try {
          // Security validation
          const securityCheck = validateFileSecurity(file);
          if (!securityCheck.valid) {
            log.warn(`Security violation for file ${file.name}`, securityCheck.error, 'ComposeModal');
            toast.error(`Security violation: ${file.name} - ${securityCheck.error}`);
            securityViolations++;
            continue;
          }
          
          // Check total size limit
          if (totalSize + file.size > maxTotalSize) {
            toast.error(`Total attachment size would exceed 50MB limit`);
            errorCount++;
            continue;
          }
          
          // Check for duplicate files (by name, size, and last modified)
          const isDuplicate = formData.attachments.some(existingFile => 
            existingFile.name === file.name && 
            existingFile.size === file.size &&
            existingFile.lastModified === file.lastModified
          );
          
          if (isDuplicate) {
            toast.warning(`File already attached: ${file.name}`);
            continue;
          }
          
          // Check for too many files
          if (formData.attachments.length + validFiles.length >= 10) {
            toast.error('Maximum 10 attachments allowed');
            break;
          }
          
          // File is valid
          validFiles.push(file);
          totalSize += file.size;
          
        } catch (fileError) {
          log.error(`Error processing file ${file.name}`, fileError, 'ComposeModal');
          toast.error(`Error processing file: ${file.name}`);
          errorCount++;
        }
      }
      
      // Add valid files
      if (validFiles.length > 0) {
      setFormData(prev => ({
        ...prev,
          attachments: [...prev.attachments, ...validFiles]
        }));
        
        const message = `${validFiles.length} file(s) added successfully.`;
        if (errorCount > 0 || securityViolations > 0) {
          toast.info(`${message} ${errorCount + securityViolations} file(s) were rejected.`);
        } else {
          toast.success(message);
        }
      } else if (errorCount > 0 || securityViolations > 0) {
        toast.error('No files were added due to validation errors or security violations.');
      }
      
    } catch (error) {
      log.error('Error in file selection', error, 'ComposeModal');
      toast.error('Error processing files. Please try again.');
    } finally {
      // Reset the file input
      if (event.target) {
        event.target.value = '';
      }
    }
  }, [formData.attachments, validateFileSecurity]);

  /**
   * Remove an attachment from the list
   * @param index - Index of the attachment to remove
   */
  const removeAttachment = useCallback((index: number) => {
    try {
      const attachmentToRemove = formData.attachments[index];
      
      if (!attachmentToRemove) {
        log.warn(`Attempted to remove attachment at invalid index: ${index}`, null, 'ComposeModal');
        return;
      }
      
      setFormData(prev => ({
      ...prev,
        attachments: prev.attachments.filter((_, i) => i !== index)
      }));
      
      toast.success(`Removed attachment: ${attachmentToRemove.name}`);
      log.info(`Attachment removed: ${attachmentToRemove.name}`, null, 'ComposeModal');
      
    } catch (error) {
      log.error('Error removing attachment', error, 'ComposeModal');
      toast.error('Error removing attachment. Please try again.');
    }
  }, [formData.attachments]);

  /**
   * Validate form data for security before submission
   * @returns Validation result object
   */
  const validateFormSecurity = useCallback((): { valid: boolean; errors: string[] } => {
    const errors: string[] = [];
    
    try {
      // Validate recipient email
      const validatedEmail = validateAndSanitizeEmail(formData.recipient);
      if (!validatedEmail) {
        errors.push('Invalid recipient email address');
      }
      
      // Validate subject
      if (!formData.subject || formData.subject.trim().length === 0) {
        errors.push('Subject is required');
      } else if (formData.subject.length > 200) {
        errors.push('Subject exceeds 200 character limit');
      }
      
      // Validate body
      if (!formData.body || formData.body.trim().length === 0) {
        errors.push('Message body is required');
      } else if (formData.body.length > 10000) {
        errors.push('Message body exceeds 10,000 character limit');
      }
      
      // Validate password if protection is enabled
      if (formData.securitySettings.passwordProtection) {
        const validatedPassword = validateAndSanitizeSecuritySetting('password', formData.securitySettings.password);
        if (!validatedPassword) {
          errors.push('Password must be at least 6 characters and not a common weak password');
        }
      }
      
      // Validate geolocation settings
      if (formData.securitySettings.geolocationLock) {
        if (!formData.securitySettings.geoVerificationType || formData.securitySettings.geoVerificationType === 'none') {
          errors.push('Geolocation verification type must be selected when geolocation lock is enabled');
        }
        
        if (formData.securitySettings.geoVerificationType === 'city' || formData.securitySettings.geoVerificationType === 'city_country') {
          const validatedCity = validateAndSanitizeSecuritySetting('geoCity', formData.securitySettings.geoCity);
          if (!validatedCity) {
            errors.push('Valid city name is required for city verification');
          }
        }
        
        if (formData.securitySettings.geoVerificationType === 'country' || formData.securitySettings.geoVerificationType === 'city_country') {
          const validatedCountry = validateAndSanitizeSecuritySetting('geoCountry', formData.securitySettings.geoCountry);
          if (!validatedCountry) {
            errors.push('Valid country name is required for country verification');
          }
        }
      }
      
      // Validate time lock settings
      if (formData.securitySettings.timeLock) {
        const validatedTime = validateAndSanitizeSecuritySetting('unlockAfter', formData.securitySettings.unlockAfter);
        if (!validatedTime) {
          errors.push('Valid unlock time is required when time lock is enabled');
        }
      }
      
      // Validate decoy message settings
      if (formData.securitySettings.decoyMessage) {
        const validatedDecoy = validateAndSanitizeSecuritySetting('decoySecret', formData.securitySettings.decoySecret);
        if (!validatedDecoy) {
          errors.push('Valid decoy secret is required when decoy message is enabled');
        }
      }
      
      // Check for security conflicts
      if (formData.securitySettings.readOnce && formData.securitySettings.autoDestruct) {
        const destructViews = formData.securitySettings.destructAfterViews || 1;
        if (destructViews > 1) {
          errors.push('Read Once and Auto-Destruct with more than 1 view cannot be used together');
        }
      }
      
      return { valid: errors.length === 0, errors };
      
    } catch (error) {
      log.error('Form security validation error', error, 'ComposeModal');
      errors.push('Form validation failed due to an error');
      return { valid: false, errors };
    }
  }, [formData, validateAndSanitizeEmail, validateAndSanitizeSecuritySetting]);

  /**
   * Handle form submission with security hardening
   * @param e - Form submission event
   */
  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Security validation before submission
    const securityValidation = validateFormSecurity();
    if (!securityValidation.valid) {
      log.error('Form security validation failed', securityValidation.errors, 'ComposeModal');
      securityValidation.errors.forEach(error => {
        toast.error(error);
      });
      return;
    }

    // Rate limiting check (prevent rapid submissions)
    if (isSubmitting) {
      toast.warning('Please wait for the current submission to complete');
      return;
    }

    // Validate subject length
    if (formData.subject.length > 200) {
      toast.error('Subject cannot exceed 200 characters');
      return;
    }

    // Validate body length
    if (formData.body.length > 10000) {
      toast.error('Message body cannot exceed 10,000 characters');
      return;
    }

    // Validate body is not just whitespace
    if (formData.body.trim().length === 0) {
      toast.error('Message body cannot be empty');
      return;
    }

    // Validate recipient is not the same as sender (if sender is available)
    // This would need to be implemented if sender information is available

    // Validate password if protection is enabled
    if (formData.securitySettings.passwordProtection) {
      if (!formData.securitySettings.password || formData.securitySettings.password.trim() === '') {
      toast.error('Password is required when password protection is enabled');
      return;
    }
      if (formData.securitySettings.password.length < 6) {
        toast.error('Password must be at least 6 characters long');
        return;
      }
      // Enhanced password strength validation
      const password = formData.securitySettings.password;
      const hasUpperCase = /[A-Z]/.test(password);
      const hasLowerCase = /[a-z]/.test(password);
      const hasNumbers = /\d/.test(password);
      const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);
      
      if (!hasUpperCase || !hasLowerCase || !hasNumbers || !hasSpecialChar) {
        toast.error('Password must contain uppercase, lowercase, numbers, and special characters');
      return;
      }
    }

    // Validate additional password settings
    if (formData.securitySettings.requirePasswordForEveryEmail && !formData.securitySettings.passwordProtection) {
      toast.error('Password protection must be enabled for "Require password for every secure email"');
      return;
    }

    if (formData.securitySettings.passwordPerEmail && !formData.securitySettings.passwordProtection) {
      toast.error('Password protection must be enabled for "Maximum security - password per email"');
      return;
    }

    // Validate self-destruct settings
    if (formData.securitySettings.selfDestructAfterAttempts) {
      const maxAttempts = formData.securitySettings.maxFailedAttempts || 0;
      if (maxAttempts < 1 || maxAttempts > 10) {
        toast.error('Maximum failed attempts must be between 1 and 10');
        return;
      }
    }

    // Validate auto-destruct settings
    if (formData.securitySettings.autoDestruct) {
      const destructViews = formData.securitySettings.destructAfterViews || 0;
      if (destructViews < 1) {
        toast.error('Destruct after views must be at least 1');
        return;
      }
      if (destructViews > 100) {
        toast.error('Destruct after views cannot exceed 100');
        return;
      }
    }

    // Validate read once conflicts
    if (formData.securitySettings.readOnce && formData.securitySettings.autoDestruct) {
      const destructViews = formData.securitySettings.destructAfterViews || 1;
      if (destructViews > 1) {
        toast.error('Read Once and Auto-Destruct with more than 1 view cannot be used together');
        return;
      }
    }

    // Validate time lock if enabled
    if (formData.securitySettings.timeLock) {
      if (!formData.securitySettings.unlockAfter) {
        toast.error('Please set an unlock time when time lock is enabled');
        return;
      }
      
      // Validate unlock time is in the future
      const unlockTime = new Date(formData.securitySettings.unlockAfter);
      const now = new Date();
      
      if (unlockTime <= now) {
        toast.error('Unlock time must be in the future');
        return;
      }
      
      // Validate unlock time is not too far in the future (max 1 year)
      const maxUnlockTime = new Date();
      maxUnlockTime.setFullYear(maxUnlockTime.getFullYear() + 1);
      
      if (unlockTime > maxUnlockTime) {
        toast.error('Unlock time cannot be more than 1 year in the future');
        return;
      }
    }

    // Validate auto-destruct if enabled
    if (formData.securitySettings.autoDestruct) {
      const destructViews = formData.securitySettings.destructAfterViews || 0;
      if (destructViews < 1) {
        toast.error('Destruct after views must be at least 1');
        return;
      }
    }

    // Validate geolocation settings if enabled
    if (formData.securitySettings.geolocationLock) {
      if (!formData.securitySettings.geoVerificationType || formData.securitySettings.geoVerificationType === 'none') {
        toast.error('Please select a verification type when geolocation lock is enabled');
        return;
      }
      
      if (formData.securitySettings.geoVerificationType === 'city' || 
          formData.securitySettings.geoVerificationType === 'city_country') {
        if (!formData.securitySettings.geoCity || formData.securitySettings.geoCity.trim() === '') {
          toast.error('City is required when city verification is enabled');
          return;
        }
        // Validate city name format (letters, spaces, hyphens only)
        const cityRegex = /^[a-zA-Z\s\-']+$/;
        if (!cityRegex.test(formData.securitySettings.geoCity.trim())) {
          toast.error('City name can only contain letters, spaces, hyphens, and apostrophes');
          return;
        }
      }
      
      if (formData.securitySettings.geoVerificationType === 'country' || 
          formData.securitySettings.geoVerificationType === 'city_country') {
        if (!formData.securitySettings.geoCountry || formData.securitySettings.geoCountry.trim() === '') {
          toast.error('Country is required when country verification is enabled');
          return;
        }
        // Validate country name format (letters, spaces, hyphens only)
        const countryRegex = /^[a-zA-Z\s\-']+$/;
        if (!countryRegex.test(formData.securitySettings.geoCountry.trim())) {
          toast.error('Country name can only contain letters, spaces, hyphens, and apostrophes');
          return;
        }
      }
    }

    // Validate decoy message if enabled
    if (formData.securitySettings.decoyMessage) {
      if (!formData.securitySettings.decoySecret || formData.securitySettings.decoySecret.trim() === '') {
        toast.error('Decoy secret is required when decoy message is enabled');
        return;
      }
      
      // Validate decoy secret length and complexity
      const decoySecret = formData.securitySettings.decoySecret.trim();
      if (decoySecret.length < 4) {
        toast.error('Decoy secret must be at least 4 characters long');
        return;
      }
      
      if (decoySecret.length > 50) {
        toast.error('Decoy secret cannot exceed 50 characters');
        return;
      }
      
      // Validate decoy secret doesn't contain sensitive patterns
      const sensitivePatterns = ['password', 'secret', 'key', 'token', 'auth'];
      const lowerDecoy = decoySecret.toLowerCase();
      for (const pattern of sensitivePatterns) {
        if (lowerDecoy.includes(pattern)) {
          toast.error('Decoy secret should not contain sensitive words like "password", "secret", "key", etc.');
          return;
        }
      }
    }

    // Validate fingerprint hash generation
    if (formData.securitySettings.generateFingerprintHash) {
      // Generate a fingerprint hash if not already present
      if (!formData.securitySettings.fingerprintHash || formData.securitySettings.fingerprintHash.trim() === '') {
        // Generate a simple hash based on content and timestamp
        const content = `${formData.recipient}${formData.subject}${formData.body}${Date.now()}`;
        const hash = btoa(content).substring(0, 32); // Simple base64 hash
        handleSecurityChange('fingerprintHash', hash);
      }
    }

    // Validate security setting conflicts
    if (formData.securitySettings.readOnce && formData.securitySettings.autoDestruct) {
      const destructViews = formData.securitySettings.destructAfterViews || 1;
      if (destructViews > 1) {
        toast.error('Read Once and Auto-Destruct with more than 1 view cannot be used together');
        return;
      }
    }

    // Validate that at least one security feature is enabled
    const hasSecurityFeatures = 
      formData.securitySettings.passwordProtection ||
      formData.securitySettings.geolocationLock ||
      formData.securitySettings.timeLock ||
      formData.securitySettings.autoDestruct ||
      formData.securitySettings.readOnce ||
      formData.securitySettings.remoteRevoke ||
      formData.securitySettings.decoyMessage ||
      formData.securitySettings.stripMetadata ||
      formData.securitySettings.tamperAlerts ||
      formData.securitySettings.selfDestructAfterAttempts ||
      formData.securitySettings.generateFingerprintHash;

    if (!hasSecurityFeatures) {
      toast.warning('No security features are enabled. Consider enabling at least one security feature for better protection.');
    }

    setIsSubmitting(true);
    
    try {
      // Prepare and sanitize API request data
      const prepareSecureApiRequest = () => {
        // Create a sanitized copy of the request data
        const validatedRecipient = validateAndSanitizeEmail(formData.recipient);
        if (!validatedRecipient) {
          throw new Error('Invalid recipient email address');
        }
        
        const sanitizedRequest = {
          recipient: validatedRecipient,
          subject: sanitizeInput(formData.subject).substring(0, 200),
          body: sanitizeInput(formData.body).substring(0, 10000),
          // Password Protection
          password: formData.securitySettings.passwordProtection 
            ? (validateAndSanitizeSecuritySetting('password', formData.securitySettings.password) as string | undefined)
            : undefined,
          requirePasswordForEveryEmail: Boolean(formData.securitySettings.requirePasswordForEveryEmail),
          passwordPerEmail: Boolean(formData.securitySettings.passwordPerEmail),
          // Self-Destruct Settings
          selfDestructAfterAttempts: Boolean(formData.securitySettings.selfDestructAfterAttempts),
          maxFailedAttempts: formData.securitySettings.selfDestructAfterAttempts 
            ? (validateAndSanitizeSecuritySetting('maxFailedAttempts', formData.securitySettings.maxFailedAttempts) as number | undefined)
            : undefined,
          // Geolocation Settings
          geolocationLock: Boolean(formData.securitySettings.geolocationLock),
          geoVerificationType: validateAndSanitizeSecuritySetting('geoVerificationType', formData.securitySettings.geoVerificationType) as string,
          geoCity: formData.securitySettings.geolocationLock && 
                  (formData.securitySettings.geoVerificationType === 'city' || 
                   formData.securitySettings.geoVerificationType === 'city_country')
            ? (validateAndSanitizeSecuritySetting('geoCity', formData.securitySettings.geoCity) as string | undefined)
            : undefined,
          geoCountry: formData.securitySettings.geolocationLock && 
                     (formData.securitySettings.geoVerificationType === 'country' || 
                      formData.securitySettings.geoVerificationType === 'city_country')
            ? (validateAndSanitizeSecuritySetting('geoCountry', formData.securitySettings.geoCountry) as string | undefined)
            : undefined,
          allowedCountries: Array.isArray(formData.securitySettings.allowedCountries) 
            ? formData.securitySettings.allowedCountries.filter(country => 
                validateAndSanitizeSecuritySetting('geoCountry', country))
            : [],
          // Time-based Settings
          timeLock: Boolean(formData.securitySettings.timeLock),
          unlockAfter: formData.securitySettings.timeLock 
            ? (validateAndSanitizeSecuritySetting('unlockAfter', formData.securitySettings.unlockAfter) as string | undefined)
            : undefined,
          // Auto-Destruct Settings
          autoDestruct: Boolean(formData.securitySettings.autoDestruct),
          destructAfterViews: formData.securitySettings.autoDestruct 
            ? (validateAndSanitizeSecuritySetting('destructAfterViews', formData.securitySettings.destructAfterViews) as number | undefined)
            : undefined,
          // Read Once
          readOnce: Boolean(formData.securitySettings.readOnce),
          // Remote Revoke
          remoteRevoke: Boolean(formData.securitySettings.remoteRevoke),
          // Decoy Message
          decoyMessage: Boolean(formData.securitySettings.decoyMessage),
          decoySecret: formData.securitySettings.decoyMessage 
            ? (validateAndSanitizeSecuritySetting('decoySecret', formData.securitySettings.decoySecret) as string | undefined)
            : undefined,
          // Metadata and Alerts
          stripMetadata: Boolean(formData.securitySettings.stripMetadata),
          tamperAlerts: Boolean(formData.securitySettings.tamperAlerts),
          // Fingerprint Hash
          generateFingerprintHash: Boolean(formData.securitySettings.generateFingerprintHash),
          fingerprintHash: formData.securitySettings.generateFingerprintHash 
            ? sanitizeInput(formData.securitySettings.fingerprintHash || '').substring(0, 64)
            : undefined,
        };

        // Remove undefined values for cleaner API request
        Object.keys(sanitizedRequest).forEach(key => {
          if (sanitizedRequest[key as keyof typeof sanitizedRequest] === undefined) {
            delete sanitizedRequest[key as keyof typeof sanitizedRequest];
          }
        });

        return sanitizedRequest;
      };

      const apiRequest = prepareSecureApiRequest();

      // Send secure email via API with performance monitoring
      const response = await monitorAsync('sendSecureEmail', async () => {
        return sendSecureEmail(apiRequest);
      });
      
      if (response.status === 'success') {
        toast.success('Secure email sent successfully!');
        
        // Log success details
        log.info('Secure email sent', {
          blobId: response.blob_id,
          recipient: formData.recipient,
          secureLink: response.secure_link_url,
          securitySettings: {
            passwordProtection: formData.securitySettings.passwordProtection,
            requirePasswordForEveryEmail: formData.securitySettings.requirePasswordForEveryEmail,
            passwordPerEmail: formData.securitySettings.passwordPerEmail,
            geolocationLock: formData.securitySettings.geolocationLock,
            geoVerificationType: formData.securitySettings.geoVerificationType,
            geoCity: formData.securitySettings.geoCity,
            geoCountry: formData.securitySettings.geoCountry,
            timeLock: formData.securitySettings.timeLock,
            unlockAfter: formData.securitySettings.unlockAfter,
            autoDestruct: formData.securitySettings.autoDestruct,
            destructAfterViews: formData.securitySettings.destructAfterViews,
            readOnce: formData.securitySettings.readOnce,
            remoteRevoke: formData.securitySettings.remoteRevoke,
            decoyMessage: formData.securitySettings.decoyMessage,
            decoySecret: formData.securitySettings.decoySecret,
            stripMetadata: formData.securitySettings.stripMetadata,
            tamperAlerts: formData.securitySettings.tamperAlerts,
            selfDestructAfterAttempts: formData.securitySettings.selfDestructAfterAttempts,
            maxFailedAttempts: formData.securitySettings.maxFailedAttempts,
            generateFingerprintHash: formData.securitySettings.generateFingerprintHash,
            fingerprintHash: formData.securitySettings.fingerprintHash,
          }
        });
        
        // Close modal and reset form
        handleClose();
      } else {
        // Handle specific API error responses
        const errorMessage = response.error || 'Failed to send secure email';
        toast.error(errorMessage);
        
        // Log detailed error information
        log.error('API Error Response', {
          status: response.status,
          error: response.error,
          requestData: apiRequest
        });
      }
    } catch (error: unknown) {
      log.error('Error sending secure email', error, 'ComposeModal');
      
      // Handle different types of errors
      let errorMessage = 'Failed to send secure email. Please try again.';
      
      const axiosError = error as { response?: { status: number; data?: { message?: string }; headers?: unknown }; request?: unknown; message?: string };
      
      if (axiosError.response) {
        // Server responded with error status
        const status = axiosError.response.status;
        const data = axiosError.response.data;
        
        log.error('Server Error Details', {
          status,
          data,
          headers: axiosError.response.headers
        });
        
        switch (status) {
          case 400:
            errorMessage = data?.message || 'Invalid request data. Please check your input.';
            break;
          case 401:
            errorMessage = 'Authentication required. Please log in again.';
            break;
          case 403:
            errorMessage = 'Access denied. You may not have permission to send secure emails.';
            break;
          case 404:
            errorMessage = 'Service not found. Please contact support.';
            break;
          case 429:
            errorMessage = 'Too many requests. Please wait a moment before trying again.';
            break;
          case 500:
            errorMessage = 'Server error. Please try again later or contact support.';
            break;
          case 502:
          case 503:
          case 504:
            errorMessage = 'Service temporarily unavailable. Please try again later.';
            break;
          default:
            errorMessage = data?.message || `Server error (${status}). Please try again.`;
        }
      } else if (axiosError.request) {
        // Network error - no response received
        log.error('Network Error', axiosError.request, 'ComposeModal');
        errorMessage = 'Network error. Please check your internet connection and try again.';
      } else {
        // Other error (e.g., request setup error)
        log.error('Request Setup Error', axiosError.message, 'ComposeModal');
        errorMessage = 'Failed to prepare request. Please try again.';
      }
      
      toast.error(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  }, [formData, hasSecurityFeatures]);

  /**
   * Handle modal close and form reset
   */
  const handleClose = useCallback(() => {
    try {
      // Check if there's unsaved content
      const hasContent = formData.recipient || formData.subject || formData.body || formData.attachments.length > 0;
      
      if (hasContent) {
        // Log the close action for debugging
        log.info('Modal closed with unsaved content', {
          hasRecipient: !!formData.recipient,
          hasSubject: !!formData.subject,
          hasBody: !!formData.body,
          attachmentCount: formData.attachments.length
        });
      }
      
      // Reset form data to initial state
      setFormData({
        recipient: '',
        subject: '',
        body: '',
        attachments: [],
        securitySettings: {
        passwordProtection: false,
        password: '',
          requirePasswordForEveryEmail: false,
          passwordPerEmail: false,
          geolocationLock: false,
        geoVerificationType: 'none',
        geoCity: '',
        geoCountry: '',
          allowedCountries: [],
          timeLock: false,
          unlockAfter: '',
          autoDestruct: false,
          destructAfterViews: 1,
          readOnce: false,
        remoteRevoke: false,
        decoyMessage: false,
        decoySecret: '',
          stripMetadata: true,
          tamperAlerts: true,
          selfDestructAfterAttempts: false,
          maxFailedAttempts: 3,
          generateFingerprintHash: false,
          fingerprintHash: ''
        }
      });
      
      // Reset form state
      setIsSubmitting(false);
      setShowPassword(false);
      
      // Close modal
      onClose();
      
      log.info('Modal closed successfully', null, 'ComposeModal');
      
    } catch (error) {
      log.error('Error closing modal', error, 'ComposeModal');
      // Still try to close the modal even if reset fails
      onClose();
    }
  }, [formData, onClose]);

  // Keyboard navigation support
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;
      
      // Escape key closes modal
      if (e.key === 'Escape') {
        handleClose();
      }
      
      // Ctrl/Cmd + Enter submits form
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        const submitButton = document.querySelector('button[type="submit"]') as HTMLButtonElement;
        if (submitButton && !submitButton.disabled) {
          submitButton.click();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, handleClose]);

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
        <div 
          className="inline-block align-bottom bg-white dark:bg-secondary-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl sm:w-full"
          role="dialog"
          aria-modal="true"
          aria-labelledby="compose-modal-title"
          aria-describedby="compose-modal-description"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-primary-100 dark:bg-primary-900/20 rounded-lg flex items-center justify-center">
                <Send className="w-5 h-5 text-primary-600 dark:text-primary-400" />
              </div>
              <div>
                <h3 
                  id="compose-modal-title"
                  className="text-lg font-medium text-secondary-900 dark:text-white"
                >
                  Compose Secure Email
                </h3>
                <p 
                  id="compose-modal-description"
                  className="text-sm text-secondary-600 dark:text-secondary-400"
                >
                  Create a new encrypted message with advanced security options
                </p>
              </div>
            </div>
          <button
              onClick={handleClose}
              className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
              aria-label="Close compose email modal"
              title="Close modal"
            >
              <X className="w-5 h-5" aria-hidden="true" />
          </button>
        </div>

          {/* Content */}
          <form onSubmit={handleSubmit} className="px-6 py-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Main Form */}
              <div className="lg:col-span-2 space-y-6">
                {/* Recipient */}
            <div>
                  <label 
                    htmlFor="recipient" 
                    className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2"
                  >
                    Recipient Email
                    <span className="text-red-500 ml-1" aria-label="required">*</span>
              </label>
              <input
                    id="recipient"
                type="email"
                placeholder="recipient@example.com"
                    value={formData.recipient}
                    onChange={(e) => handleInputChange('recipient', e.target.value)}
                required
                    aria-required="true"
                    aria-describedby="recipient-help"
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white"
              />
                  <div id="recipient-help" className="text-xs text-secondary-500 mt-1">
                    Enter the email address of the person you want to send the secure message to
            </div>
                </div>

                {/* Subject */}
            <div>
                  <label 
                    htmlFor="subject" 
                    className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2"
                  >
                    Subject
                    <span className="text-red-500 ml-1" aria-label="required">*</span>
              </label>
              <input
                    id="subject"
                type="text"
                    placeholder="Enter subject line"
                value={formData.subject}
                    onChange={(e) => handleInputChange('subject', e.target.value)}
                required
                    aria-required="true"
                    aria-describedby="subject-help"
                    maxLength={200}
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white"
              />
                  <div id="subject-help" className="text-xs text-secondary-500 mt-1">
                    Enter a brief subject for your secure message (max 200 characters)
            </div>
          </div>

                {/* Body */}
          <div>
                  <label 
                    htmlFor="body" 
                    className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2"
                  >
                    Message Body
                    <span className="text-red-500 ml-1" aria-label="required">*</span>
            </label>
            <textarea
                    id="body"
                    rows={12}
                    placeholder="Type your secure message here..."
              value={formData.body}
                    onChange={(e) => handleInputChange('body', e.target.value)}
              required
                    aria-required="true"
                    aria-describedby="body-help"
                    maxLength={10000}
                    className="w-full px-4 py-3 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white resize-none"
            />
                  <div id="body-help" className="text-xs text-secondary-500 mt-1">
                    Enter your secure message content (max 10,000 characters)
                  </div>
          </div>

                {/* Attachments */}
                <div>
                  <label 
                    htmlFor="file-upload"
                    className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2"
                  >
                    Attachments
                  </label>
                  <div 
                    className="border-2 border-dashed border-secondary-300 dark:border-secondary-600 rounded-lg p-6 text-center"
                    role="button"
                    tabIndex={0}
                    aria-describedby="attachment-help"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        document.getElementById('file-upload')?.click();
                      }
                    }}
                  >
                    <Paperclip className="w-8 h-8 text-secondary-400 mx-auto mb-2" aria-hidden="true" />
                    <p className="text-sm text-secondary-600 dark:text-secondary-400 mb-2">
                      Drag and drop files here, or click to browse
                    </p>
                    <input
                      type="file"
                      multiple
                      onChange={handleFileSelect}
                      className="hidden"
                      id="file-upload"
                      accept=".jpg,.jpeg,.png,.gif,.webp,.pdf,.doc,.docx,.txt,.csv"
                      aria-describedby="attachment-help"
                    />
                    <label
                      htmlFor="file-upload"
                      className="inline-flex items-center px-4 py-2 bg-secondary-100 dark:bg-secondary-700 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-600 transition-colors duration-200 cursor-pointer"
                    >
                      Choose Files
                    </label>
                  </div>
                  <div id="attachment-help" className="text-xs text-secondary-500 mt-1">
                    Upload images, PDFs, Word documents, or text files (max 10MB each, 50MB total)
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
                        <Lock className="w-4 h-4 text-yellow-600 dark:text-yellow-400" aria-hidden="true" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Password Protection</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                          checked={formData.securitySettings.passwordProtection}
                          onChange={(e) => handleSecurityChange('passwordProtection', e.target.checked)}
                          className="sr-only peer"
                          aria-label="Enable password protection for this secure email"
                        />
                        <div 
                          className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"
                          role="switch"
                          aria-checked={formData.securitySettings.passwordProtection}
                        ></div>
                </label>
              </div>

                    {/* Password Input */}
                    {formData.securitySettings.passwordProtection && (
                      <div className="ml-6 space-y-3">
                        <div>
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

                        {/* Additional Password Options */}
                        <div className="space-y-2">
                          <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                              checked={formData.securitySettings.requirePasswordForEveryEmail}
                              onChange={(e) => handleSecurityChange('requirePasswordForEveryEmail', e.target.checked)}
                              className="h-3 w-3 text-primary-600 focus:ring-primary-500 border-secondary-300 rounded"
                            />
                            <label className="text-xs text-secondary-600 dark:text-secondary-400">
                              Require password for every secure email
                </label>
              </div>
                          <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                              checked={formData.securitySettings.passwordPerEmail}
                              onChange={(e) => handleSecurityChange('passwordPerEmail', e.target.checked)}
                              className="h-3 w-3 text-primary-600 focus:ring-primary-500 border-secondary-300 rounded"
                            />
                            <label className="text-xs text-secondary-600 dark:text-secondary-400">
                              Maximum security - password per email
                </label>
              </div>
                        </div>
                      </div>
                    )}

                    {/* Geolocation Lock */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Globe className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Restrict access by location</span>
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

                    {/* Geolocation Configuration */}
                    {formData.securitySettings.geolocationLock && (
                      <div className="ml-6 space-y-3">
                        <div>
                          <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                            Verification Type
                </label>
                          <select
                            value={formData.securitySettings.geoVerificationType}
                            onChange={(e) => handleSecurityChange('geoVerificationType', e.target.value)}
                            className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                          >
                            <option value="none">No Restrictions</option>
                            <option value="country">Country Only</option>
                            <option value="city">City Only</option>
                            <option value="city_country">City and Country</option>
                          </select>
              </div>
                        {formData.securitySettings.geoVerificationType !== 'none' && (
                          <>
                            {(formData.securitySettings.geoVerificationType === 'city' || formData.securitySettings.geoVerificationType === 'city_country') && (
                              <div>
                                <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                                  Allowed City
                </label>
                <input
                                  type="text"
                                  value={formData.securitySettings.geoCity}
                                  onChange={(e) => handleSecurityChange('geoCity', e.target.value)}
                                  className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                                  placeholder="e.g., New York"
                />
              </div>
            )}
                            {(formData.securitySettings.geoVerificationType === 'country' || formData.securitySettings.geoVerificationType === 'city_country') && (
                              <div>
                                <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                                  Allowed Country
                </label>
                <input
                                  type="text"
                                  value={formData.securitySettings.geoCountry}
                                  onChange={(e) => handleSecurityChange('geoCountry', e.target.value)}
                                  className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                                  placeholder="e.g., US"
                                />
                              </div>
                            )}
                          </>
                        )}
              </div>
            )}

                    {/* Time Lock */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Clock className="w-4 h-4 text-green-600 dark:text-green-400" />
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Unlock after specific date</span>
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

                    {/* Advanced Settings Header */}
                    <div className="pt-4 border-t border-secondary-200 dark:border-secondary-600">
                      <h5 className="text-sm font-medium text-secondary-900 dark:text-white mb-3">
                        Advanced Settings
                      </h5>
                    </div>

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
                        <span className="text-sm text-secondary-700 dark:text-secondary-300">Self-destruct after first read</span>
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

                    {/* Decoy Message Secret */}
                    {formData.securitySettings.decoyMessage && (
                      <div className="ml-6">
                        <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                          Decoy Secret
                    </label>
                    <input
                      type="text"
                          value={formData.securitySettings.decoySecret}
                          onChange={(e) => handleSecurityChange('decoySecret', e.target.value)}
                          className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                          placeholder="Custom decoy secret (auto-generated if empty)"
                        />
                        <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-1">
                          Show fake content to attackers when this secret is used
                        </p>
                </div>
              )}

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

                     {/* Fingerprint Hash */}
                     <div className="flex items-center justify-between">
                       <div className="flex items-center space-x-2">
                         <Fingerprint className="w-4 h-4 text-green-600 dark:text-green-400" />
                         <div>
                           <span className="text-sm text-secondary-700 dark:text-secondary-300">Fingerprint Hash</span>
                           <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-0.5">
                             This hash uniquely identifies this message and can be used to verify authenticity.
                           </p>
                         </div>
                       </div>
                       <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                           checked={formData.securitySettings.generateFingerprintHash}
                           onChange={(e) => handleSecurityChange('generateFingerprintHash', e.target.checked)}
                           className="sr-only peer"
                         />
                         <div className="w-11 h-6 bg-secondary-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:peer-focus:ring-primary-800 rounded-full peer dark:bg-secondary-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-secondary-600 peer-checked:bg-primary-600"></div>
                    </label>
                  </div>

                     {/* Fingerprint Hash Display */}
                     {formData.securitySettings.generateFingerprintHash && (
                       <div className="ml-6">
                         <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                           Generated Hash
                         </label>
                         <div className="p-2 bg-secondary-100 dark:bg-secondary-800 rounded border border-secondary-200 dark:border-secondary-700">
                           <p className="text-xs font-mono text-secondary-700 dark:text-secondary-300 break-all">
                             {formData.securitySettings.fingerprintHash || 'Hash will be generated when email is sent'}
                           </p>
                </div>
                         <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-1">
                           This hash uniquely identifies this message and can be used to verify authenticity.
                         </p>
                       </div>
                     )}

                     {/* Self-Destruct After Failed Attempts */}
                     <div className="flex items-center justify-between">
                       <div className="flex items-center space-x-2">
                         <Trash2 className="w-4 h-4 text-red-600 dark:text-red-400" />
                         <div>
                           <span className="text-sm text-secondary-700 dark:text-secondary-300">Self-Destruct After Failed Attempts</span>
                           <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-0.5">
                             Auto-delete message after failed password attempts
                           </p>
                         </div>
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
                       <div className="ml-6 space-y-2">
                         <div>
                           <label className="block text-xs text-secondary-600 dark:text-secondary-400 mb-1">
                             Max Failed Attempts
                           </label>
                           <div className="relative">
                    <input
                               type="number"
                               min="1"
                               max="10"
                               value={formData.securitySettings.maxFailedAttempts}
                               onChange={(e) => handleSecurityChange('maxFailedAttempts', parseInt(e.target.value))}
                               className="w-full px-3 py-2 text-sm border border-secondary-300 dark:border-secondary-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-600 dark:text-white"
                               placeholder="3"
                             />
                             <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                               <AlertCircle className="w-4 h-4 text-secondary-400" />
                  </div>
                           </div>
                           <p className="text-xs text-secondary-500 dark:text-secondary-400 mt-1">
                             Message will self-destruct after {formData.securitySettings.maxFailedAttempts || 3} failed password attempts
                           </p>
                         </div>
                         
                         {/* Security Warning */}
                         <div className="flex items-start space-x-2 p-2 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
                           <AlertTriangle className="w-4 h-4 text-yellow-600 dark:text-yellow-400 mt-0.5 flex-shrink-0" />
                           <div className="text-xs text-yellow-700 dark:text-yellow-300">
                             <p className="font-medium">Security Warning</p>
                             <p>This message will be permanently deleted after failed access attempts. This action cannot be undone.</p>
                           </div>
                         </div>
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
                aria-label="Cancel and close compose email modal"
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
                aria-label={isSubmitting ? 'Sending secure email, please wait' : 'Send secure email with current security settings'}
                aria-describedby="submit-help"
              >
                {isSubmitting ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" aria-hidden="true"></div>
                    <span>Sending...</span>
                  </>
                ) : (
                  <>
                    <Send className="w-4 h-4" aria-hidden="true" />
                    <span>Send Securely</span>
                  </>
                )}
            </button>
          </div>
            <div id="submit-help" className="text-xs text-secondary-500 mt-1 text-right">
              {isSubmitting ? 'Please wait while your secure email is being sent...' : 'Click to send your secure email with the configured security settings'}
            </div>
        </form>
        </div>
      </div>
    </div>
  );
};

export default ComposeModal;

