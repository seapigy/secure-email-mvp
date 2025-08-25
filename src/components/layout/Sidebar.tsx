import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import {
  InboxIcon,
  PaperAirplaneIcon,
  DocumentTextIcon,
  TrashIcon,
  Cog6ToothIcon,
  PlusIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';

/**
 * Sidebar Component
 * 
 * Navigation sidebar with email-related links and actions.
 * Features responsive design and active state indicators.
 * 
 * Navigation Items:
 * - Inbox: View received emails
 * - Send: Compose and send new emails
 * - Drafts: View saved drafts
 * - Trash: View deleted emails
 * - Settings: User preferences
 */
const Sidebar: React.FC = () => {
  const navigate = useNavigate();

  const navigation = [
    {
      name: 'Dashboard',
      href: '/dashboard',
      icon: InboxIcon,
      count: 12, // Mock count - will be replaced with actual email count from API
    },
    {
      name: 'Send',
      href: '/send',
      icon: PaperAirplaneIcon,
    },
    {
      name: 'Drafts',
      href: '/drafts',
      icon: DocumentTextIcon,
      count: 3, // Mock count - will be replaced with actual draft count from API
    },
    {
      name: 'Monitoring',
      href: '/monitoring',
      icon: ChartBarIcon,
    },
    {
      name: 'Trash',
      href: '/trash',
      icon: TrashIcon,
    },
    {
      name: 'Settings',
      href: '/settings',
      icon: Cog6ToothIcon,
    },
  ];

  return (
    <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white dark:bg-secondary-800 border-r border-secondary-200 dark:border-secondary-700">
      {/* Logo/Brand */}
      <div className="flex items-center justify-center h-16 px-4 border-b border-secondary-200 dark:border-secondary-700">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
            <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <span className="text-xl font-bold text-secondary-900 dark:text-white">
            SecureMail
          </span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="mt-6 px-4">
        <div className="space-y-2">
          {/* Compose Button */}
          <button 
            onClick={() => navigate('/send')}
            className="w-full flex items-center justify-center px-4 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors duration-200"
          >
            <PlusIcon className="w-5 h-5 mr-2" />
            Compose
          </button>

          {/* Navigation Links */}
          <div className="mt-6 space-y-1">
            {navigation.map((item) => {
              return (
                <NavLink
                  key={item.name}
                  to={item.href}
                  className={({ isActive }) =>
                    `flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors duration-200 ${
                      isActive
                        ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
                        : 'text-secondary-700 hover:bg-secondary-100 dark:text-secondary-300 dark:hover:bg-secondary-700'
                    }`
                  }
                >
                  <item.icon className="w-5 h-5 mr-3" />
                  <span className="flex-1">{item.name}</span>
                  {item.count && (
                    <span className="ml-auto bg-secondary-200 dark:bg-secondary-600 text-secondary-700 dark:text-secondary-300 text-xs font-medium px-2 py-1 rounded-full">
                      {item.count}
                    </span>
                  )}
                </NavLink>
              );
            })}
          </div>
        </div>
      </nav>

      {/* Footer */}
      <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-secondary-200 dark:border-secondary-700">
        <div className="text-xs text-secondary-500 dark:text-secondary-400 text-center">
          Secure Email MVP v0.1.0
        </div>
      </div>
    </div>
  );
};

export default Sidebar; 