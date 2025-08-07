import React from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import {
  InboxIcon,
  PaperAirplaneIcon,
  DocumentTextIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import HealthStatusBanner from '@/components/ui/HealthStatusBanner';

/**
 * Dashboard Component
 * 
 * Main dashboard view showing email inbox and statistics.
 * This serves as the primary email management interface.
 * 
 * Features:
 * - Email list with preview
 * - Quick actions
 * - Email statistics
 * - Search functionality
 */
const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  // Mock email data for demonstration
  const emails = [
    {
      id: '1',
      from: 'john@securesystem.email',
      subject: 'Welcome to SecureMail',
      preview: 'Thank you for joining our secure email platform...',
      date: '2024-01-15T10:30:00Z',
      read: false,
      important: true,
    },
    {
      id: '2',
      from: 'support@securesystem.email',
      subject: 'Account Setup Complete',
      preview: 'Your account has been successfully configured...',
      date: '2024-01-15T09:15:00Z',
      read: true,
      important: false,
    },
    {
      id: '3',
      from: 'alice@securesystem.email',
      subject: 'Security Update',
      preview: 'We have implemented new security features...',
      date: '2024-01-14T16:45:00Z',
      read: true,
      important: true,
    },
  ];

  const stats = [
    { name: 'Unread', value: 5, icon: InboxIcon, color: 'text-blue-600' },
    { name: 'Sent', value: 12, icon: PaperAirplaneIcon, color: 'text-green-600' },
    { name: 'Drafts', value: 3, icon: DocumentTextIcon, color: 'text-yellow-600' },
    { name: 'Trash', value: 2, icon: TrashIcon, color: 'text-red-600' },
  ];

  return (
    <div className="p-6 space-y-6">
      {/* Health Status Banner */}
      <HealthStatusBanner />
      
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-secondary-900 dark:text-white">
            Inbox
          </h1>
          <p className="text-secondary-600 dark:text-secondary-400">
            Welcome back, {user?.email}
          </p>
        </div>
        <button 
          onClick={() => navigate('/send')}
          className="w-full sm:w-auto px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200"
        >
          Compose
        </button>
      </div>

      {/* Statistics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <div
            key={stat.name}
            className="bg-white dark:bg-secondary-800 p-6 rounded-lg border border-secondary-200 dark:border-secondary-700"
          >
            <div className="flex items-center">
              <div className={`p-2 rounded-lg bg-secondary-100 dark:bg-secondary-700`}>
                <stat.icon className={`w-6 h-6 ${stat.color}`} />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-secondary-600 dark:text-secondary-400">
                  {stat.name}
                </p>
                <p className="text-2xl font-bold text-secondary-900 dark:text-white">
                  {stat.value}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Email List */}
      <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700">
        <div className="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-white">
            Recent Emails
          </h2>
        </div>
        <div className="divide-y divide-secondary-200 dark:divide-secondary-700">
          {emails.map((email) => (
            <div
              key={email.id}
              onClick={() => navigate(`/view/${email.id}`)}
              className={`px-6 py-4 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors duration-200 cursor-pointer ${
                !email.read ? 'bg-blue-50 dark:bg-blue-900/10' : ''
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-3">
                    <div className="flex-shrink-0">
                      <div className="w-8 h-8 bg-secondary-300 dark:bg-secondary-600 rounded-full flex items-center justify-center">
                        <span className="text-sm font-medium text-secondary-700 dark:text-secondary-300">
                          {email.from.charAt(0).toUpperCase()}
                        </span>
                      </div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2">
                        <p className={`text-sm font-medium truncate ${
                          !email.read 
                            ? 'text-secondary-900 dark:text-white' 
                            : 'text-secondary-700 dark:text-secondary-300'
                        }`}>
                          {email.from}
                        </p>
                        {email.important && (
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400">
                            Important
                          </span>
                        )}
                      </div>
                      <p className={`text-sm truncate ${
                        !email.read 
                          ? 'text-secondary-900 dark:text-white font-medium' 
                          : 'text-secondary-600 dark:text-secondary-400'
                      }`}>
                        {email.subject}
                      </p>
                      <p className="text-sm text-secondary-500 dark:text-secondary-500 truncate">
                        {email.preview}
                      </p>
                    </div>
                  </div>
                </div>
                <div className="flex-shrink-0 ml-4">
                  <p className="text-xs text-secondary-500 dark:text-secondary-400">
                    {new Date(email.date).toLocaleDateString()}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Dashboard; 