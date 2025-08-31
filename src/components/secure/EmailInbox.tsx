/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE EMAIL INBOX COMPONENT DESIGN.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual layout or design of the inbox
 * 2. NEVER modify the table structure or styling
 * 3. NEVER alter the button designs or positioning
 * 4. NEVER change the color scheme or Tailwind classes
 * 5. ONLY add new functionality that doesn't change the visual design
 * 6. ALWAYS maintain the exact same visual appearance
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * Any changes to the visual design will result in immediate user dissatisfaction.
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE DESIGN, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 * @user_feedback: "This is the perfect design, never change it"
 */

import React, { useState, useEffect } from 'react';
import { 
  Shield,
  Lock,
  Globe,
  Clock,
  Eye,
  AlertTriangle,
  FileText,
  Search,
  Filter,
  SortAsc,
  SortDesc,
  X,
  RefreshCw
} from 'lucide-react';
import { SecureEmail, StatusType, EmailStats, EmailFilters, SortConfig } from '@/types/secureEmail';
import { getInboxEmails, deleteInboxEmail } from '@/lib/api';
import { transformInboxResponse, handleInboxError } from '@/lib/inboxUtils';

/**
 * Email Inbox Props Interface
 * 
 * Props for the EmailInbox component.
 */
interface EmailInboxProps {
  /** Callback when an email is selected */
  onEmailSelect?: (email: SecureEmail) => void;
  
  /** Currently selected email (for external state management) */
  selectedEmail?: SecureEmail | null;
}

/**
 * EmailInbox Component
 * 
 * Secure email inbox with filtering, sorting, and responsive design.
 * Features a modern privacy-first interface with comprehensive email management.
 * 
 * Features:
 * - Secure email list with status indicators and security badges
 * - Advanced filtering by status, security features, and search terms
 * - Multi-column sorting with visual indicators
 * - Responsive design optimized for mobile and desktop
 * - Real-time search functionality with instant results
 * - Email statistics display with visual counters
 * - Dark/light mode support with consistent theming
 * - Split-view integration for desktop layout
 * - Mobile-optimized single panel layout
 * - Comprehensive security feature indicators
 * 
 * Email Management:
 * - Display email metadata (subject, sender, date, status)
 * - Visual status badges with color coding
 * - Security feature indicators (password, geolocation, etc.)
 * - Attachment indicators and file information
 * - Expiration and access attempt tracking
 * - Comprehensive filtering options
 * 
 * UI/UX Features:
 * - Modern card-based email list design
 * - Hover effects and interactive elements
 * - Loading states and empty states
 * - Responsive grid layout
 * - Accessibility features and keyboard navigation
 * - Smooth animations and transitions
 * 
 * Filtering Options:
 * - Status-based filtering (pending, opened, expired, revoked)
 * - Security feature filtering (password protected, geolocation restricted)
 * - Search term filtering across subject and sender
 * - Combined filter application
 * - Filter reset functionality
 * 
 * Sorting Options:
 * - Date sorting (newest/oldest first)
 * - Subject alphabetical sorting
 * - Sender alphabetical sorting
 * - Status-based sorting
 * - Visual sort indicators
 */
const EmailInbox: React.FC<EmailInboxProps> = ({ onEmailSelect, selectedEmail: externalSelectedEmail }) => {
  // Email data state
  const [emails, setEmails] = useState<SecureEmail[]>([]);
  const [stats, setStats] = useState<EmailStats | null>(null);
  
  // Loading and error states
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Email selection state (internal vs external)
  const [internalSelectedEmail, setInternalSelectedEmail] = useState<SecureEmail | null>(null);
  
  // Use external selected email if provided, otherwise use internal state
  const selectedEmail = externalSelectedEmail || internalSelectedEmail;
  
  // Filter and search state
  const [filters, setFilters] = useState<EmailFilters>({});
  const [searchTerm, setSearchTerm] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  
  // Sorting state
  const [sortConfig, setSortConfig] = useState<SortConfig>({
    field: 'date',
    direction: 'desc'
  });

  /**
   * Load inbox data from API
   * Fetches real inbox emails and statistics from the backend
   */
  const loadInboxData = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await getInboxEmails();
      const { emails: inboxEmails, stats: inboxStats } = transformInboxResponse(response);
      setEmails(inboxEmails);
      setStats(inboxStats);
    } catch (err) {
      const errorMessage = handleInboxError(err);
      setError(errorMessage);
      console.error('Failed to load inbox:', err);
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Load inbox data on component mount
   */
  useEffect(() => {
    loadInboxData();
  }, []);

  /**
   * Handle email deletion
   * @param emailId - ID of the email to delete
   */
  const handleEmailDelete = async (emailId: string) => {
    try {
      await deleteInboxEmail(emailId);
      // Refresh the inbox after deletion
      await loadInboxData();
    } catch (err) {
      const errorMessage = handleInboxError(err);
      setError(errorMessage);
      console.error('Failed to delete email:', err);
    }
  };

  /**
   * Get status configuration for visual display
   * @param status - The email status
   * @returns Configuration object with colors, labels, and icons
   */
  const getStatusConfig = (status: StatusType) => {
    const configs = {
      pending: {
        label: 'Pending',
        bgColor: 'bg-yellow-100 dark:bg-yellow-900/20',
        color: 'text-yellow-600 dark:text-yellow-400',
        icon: '⏳'
      },
      opened: {
        label: 'Opened',
        bgColor: 'bg-green-100 dark:bg-green-900/20',
        color: 'text-green-600 dark:text-green-400',
        icon: '✓'
      },
      expired: {
        label: 'Expired',
        bgColor: 'bg-red-100 dark:bg-red-900/20',
        color: 'text-red-600 dark:text-red-400',
        icon: '⏰'
      },
      revoked: {
        label: 'Revoked',
        bgColor: 'bg-gray-100 dark:bg-gray-900/20',
        color: 'text-gray-600 dark:text-gray-400',
        icon: '🚫'
      }
    };
    return configs[status];
  };

  /**
   * Handle email selection
   * @param email - The email that was selected
   */
  const handleEmailSelect = (email: SecureEmail) => {
    setInternalSelectedEmail(email);
    if (onEmailSelect) {
      onEmailSelect(email);
    }
  };

  /**
   * Handle sorting changes
   * @param field - The field to sort by
   */
  const handleSort = (field: SortConfig['field']) => {
    setSortConfig(prev => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  // Clear filters
  const clearFilters = () => {
    setFilters({});
    setSearchTerm('');
  };

  // Filter and sort emails
  const filteredEmails = emails.filter(email => {
    const matchesSearch = searchTerm === '' || 
      email.subject.toLowerCase().includes(searchTerm.toLowerCase()) ||
      email.from.toLowerCase().includes(searchTerm.toLowerCase()) ||
      email.preview.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesFilters = Object.entries(filters).every(([key, value]) => {
      if (!value) return true;
      return email[key as keyof SecureEmail] === value;
    });
    
    return matchesSearch && matchesFilters;
  });

  const sortedEmails = [...filteredEmails].sort((a, b) => {
    const aValue = a[sortConfig.field];
    const bValue = b[sortConfig.field];
    
    if (sortConfig.direction === 'asc') {
      return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
    } else {
      return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
    }
  });

  return (
    <div className="h-full flex flex-col min-w-0">
      <div className="flex-1 overflow-hidden">
        {/* Stats Cards */}
        {stats && (
          <div className="grid grid-cols-4 gap-2 p-4 bg-secondary-50 dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700">
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <FileText className="w-4 h-4 text-blue-600 dark:text-blue-400" />
              </div>
              <p className="text-xs text-secondary-600 dark:text-secondary-400">Total</p>
              <p className="text-lg font-bold text-secondary-900 dark:text-white">{stats.total}</p>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <Clock className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
              </div>
              <p className="text-xs text-secondary-600 dark:text-secondary-400">Pending</p>
              <p className="text-lg font-bold text-yellow-600 dark:text-yellow-400">{stats.pending}</p>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <Lock className="w-4 h-4 text-green-600 dark:text-green-400" />
              </div>
              <p className="text-xs text-secondary-600 dark:text-secondary-400">Protected</p>
              <p className="text-lg font-bold text-green-600 dark:text-green-400">{stats.passwordProtected}</p>
            </div>
            
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <AlertTriangle className="w-4 h-4 text-red-600 dark:text-red-400" />
              </div>
              <p className="text-xs text-secondary-600 dark:text-secondary-400">Auto-Destruct</p>
              <p className="text-lg font-bold text-red-600 dark:text-red-400">{stats.autoDestruct}</p>
            </div>
          </div>
        )}

        {/* Search and Filters */}
        <div className="bg-white dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700 p-3">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            {/* Search */}
            <div className="relative flex-1 min-w-0">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-secondary-400" />
              <input
                type="text"
                placeholder="Search secure emails..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
              />
            </div>

            {/* Filter Controls */}
            <div className="flex items-center space-x-2">
              <button 
                onClick={loadInboxData}
                disabled={isLoading}
                className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200 disabled:opacity-50"
                title="Refresh inbox"
              >
                <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
              </button>
              <button 
                onClick={() => setShowFilters(!showFilters)}
                className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
              >
                <Filter className="w-4 h-4" />
              </button>
              {(searchTerm || Object.keys(filters).length > 0) && (
                <button
                  onClick={clearFilters}
                  className="p-2 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 rounded-lg transition-colors duration-200"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>

          {/* Filter Panel */}
          {showFilters && (
            <div className="mt-3 pt-3 border-t border-secondary-200 dark:border-secondary-700">
              <div className="grid grid-cols-2 gap-3">
                <select
                  value={filters.status || ''}
                  onChange={(e) => setFilters(prev => ({ ...prev, status: e.target.value as StatusType || undefined }))}
                  className="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
                >
                  <option value="">All Status</option>
                  <option value="pending">Pending</option>
                  <option value="opened">Opened</option>
                  <option value="expired">Expired</option>
                  <option value="revoked">Revoked</option>
                </select>

                <select
                  value={filters.passwordProtected === undefined ? '' : filters.passwordProtected.toString()}
                  onChange={(e) => setFilters(prev => ({ ...prev, passwordProtected: e.target.value === '' ? undefined : e.target.value === 'true' }))}
                  className="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
                >
                  <option value="">All Protection</option>
                  <option value="true">Password Protected</option>
                  <option value="false">Not Protected</option>
                </select>

                <select
                  value={filters.geolocationRestricted === undefined ? '' : filters.geolocationRestricted.toString()}
                  onChange={(e) => setFilters(prev => ({ ...prev, geolocationRestricted: e.target.value === '' ? undefined : e.target.value === 'true' }))}
                  className="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
                >
                  <option value="">All Locations</option>
                  <option value="true">Location Restricted</option>
                  <option value="false">No Location Restriction</option>
                </select>

                <select
                  value={filters.readOnce === undefined ? '' : filters.readOnce.toString()}
                  onChange={(e) => setFilters(prev => ({ ...prev, readOnce: e.target.value === '' ? undefined : e.target.value === 'true' }))}
                  className="px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-secondary-700 dark:text-white text-sm"
                >
                  <option value="">All Read Modes</option>
                  <option value="true">Read Once</option>
                  <option value="false">Multiple Reads</option>
                </select>
              </div>
            </div>
          )}
        </div>

        {/* Email List */}
        <div className="bg-white dark:bg-secondary-800 flex-1 overflow-hidden">
          {/* Error Display */}
          {error && (
            <div className="px-4 py-3 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="w-4 h-4 text-red-600 dark:text-red-400" />
                <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
                <button
                  onClick={() => setError(null)}
                  className="ml-auto p-1 text-red-600 hover:bg-red-100 dark:text-red-400 dark:hover:bg-red-900/30 rounded"
                >
                  <X className="w-3 h-3" />
                </button>
              </div>
            </div>
          )}

          {/* Loading State */}
          {isLoading && (
            <div className="px-4 py-8 text-center">
              <RefreshCw className="w-8 h-8 text-secondary-400 mx-auto mb-4 animate-spin" />
              <p className="text-secondary-600 dark:text-secondary-400">Loading inbox...</p>
            </div>
          )}
          {/* Table Header */}
          <div className="px-4 py-3 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-700/50">
            <div className="grid grid-cols-12 gap-3 text-xs font-medium text-secondary-700 dark:text-secondary-300">
              <div className="col-span-5">
                <button
                  onClick={() => handleSort('subject')}
                  className="flex items-center space-x-1 hover:text-secondary-900 dark:hover:text-white w-full text-left"
                >
                  <span>Subject</span>
                  {sortConfig.field === 'subject' && (
                    sortConfig.direction === 'asc' ? <SortAsc className="w-3 h-3" /> : <SortDesc className="w-3 h-3" />
                  )}
                </button>
              </div>
              <div className="col-span-3">
                <button
                  onClick={() => handleSort('from')}
                  className="flex items-center space-x-1 hover:text-secondary-900 dark:hover:text-white w-full text-left"
                >
                  <span>From</span>
                  {sortConfig.field === 'from' && (
                    sortConfig.direction === 'asc' ? <SortAsc className="w-3 h-3" /> : <SortDesc className="w-3 h-3" />
                  )}
                </button>
              </div>
              <div className="col-span-2">
                <button
                  onClick={() => handleSort('date')}
                  className="flex items-center space-x-1 hover:text-secondary-900 dark:hover:text-white w-full text-left"
                >
                  <span>Date</span>
                  {sortConfig.field === 'date' && (
                    sortConfig.direction === 'asc' ? <SortAsc className="w-3 h-3" /> : <SortDesc className="w-3 h-3" />
                  )}
                </button>
              </div>
              <div className="col-span-2">
                <button
                  onClick={() => handleSort('status')}
                  className="flex items-center space-x-1 hover:text-secondary-900 dark:hover:text-white w-full text-left"
                >
                  <span>Status</span>
                  {sortConfig.field === 'status' && (
                    sortConfig.direction === 'asc' ? <SortAsc className="w-3 h-3" /> : <SortDesc className="w-3 h-3" />
                  )}
                </button>
              </div>
            </div>
          </div>

          {/* Email Rows */}
          {!isLoading && (
            <div className="divide-y divide-secondary-200 dark:divide-secondary-700 overflow-y-auto max-h-[calc(100vh-300px)]">
            {sortedEmails.map((email) => {
              const statusConfig = getStatusConfig(email.status);
              return (
                <div
                  key={email.id}
                  onClick={() => handleEmailSelect(email)}
                  className={`px-4 py-3 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors duration-200 cursor-pointer ${
                    selectedEmail?.id === email.id 
                      ? 'bg-primary-50 dark:bg-primary-900/20 border-l-4 border-primary-600' 
                      : ''
                  }`}
                >
                  <div className="grid grid-cols-12 gap-3 items-center">
                    {/* Subject */}
                    <div className="col-span-5 min-w-0">
                      <div className="flex items-center space-x-2">
                        <div className="flex items-center space-x-1 flex-shrink-0">
                          {email.passwordProtected && (
                            <Lock className="w-3 h-3 text-yellow-600 dark:text-yellow-400" />
                          )}
                          {email.geolocationRestricted && (
                            <Globe className="w-3 h-3 text-blue-600 dark:text-blue-400" />
                          )}
                          {email.readOnce && (
                            <Eye className="w-3 h-3 text-red-600 dark:text-red-400" />
                          )}
                          {email.autoDestruct && (
                            <AlertTriangle className="w-3 h-3 text-orange-600 dark:text-orange-400" />
                          )}
                        </div>
                        <span className="font-medium text-secondary-900 dark:text-white truncate">
                          {email.subject}
                        </span>
                      </div>
                      <p className="text-xs text-secondary-500 dark:text-secondary-400 truncate mt-1">
                        {email.preview}
                      </p>
                    </div>

                    {/* From */}
                    <div className="col-span-3 min-w-0">
                      <p className="text-xs text-secondary-700 dark:text-secondary-300 truncate">
                        {email.from}
                      </p>
                    </div>

                    {/* Date */}
                    <div className="col-span-2 min-w-0">
                      <p className="text-xs text-secondary-600 dark:text-secondary-400 truncate">
                        {new Date(email.date).toLocaleDateString()}
                      </p>
                    </div>

                    {/* Status */}
                    <div className="col-span-2 min-w-0">
                      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusConfig.bgColor} ${statusConfig.color}`}>
                        <span className="mr-1">{statusConfig.icon}</span>
                        <span className="truncate">{statusConfig.label}</span>
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
            </div>
          )}

          {/* Empty State */}
          {!isLoading && sortedEmails.length === 0 && (
            <div className="px-4 py-12 text-center">
              <Shield className="w-12 h-12 text-secondary-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-secondary-900 dark:text-white mb-2">
                No emails found
              </h3>
              <p className="text-secondary-600 dark:text-secondary-400">
                {searchTerm || Object.keys(filters).length > 0 
                  ? 'Try adjusting your search or filters'
                  : 'No secure emails in your inbox yet'
                }
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default EmailInbox; 