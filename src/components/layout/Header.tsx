import React from 'react';
import { User, Theme } from '@/types';
import {
  SunIcon,
  MoonIcon,
  BellIcon,
  UserCircleIcon,
} from '@heroicons/react/24/outline';

interface HeaderProps {
  user: User | null;
  onLogout: () => void;
  onToggleTheme: () => void;
  theme: Theme;
}

/**
 * Header Component
 * 
 * Top navigation bar with user information, theme toggle, and logout.
 * Features responsive design and user-friendly interface.
 * 
 * Features:
 * - User profile display
 * - Theme toggle (dark/light mode)
 * - Notifications icon
 * - Logout functionality
 */
const Header: React.FC<HeaderProps> = ({ user, onLogout, onToggleTheme, theme }) => {
  return (
    <header className="bg-white dark:bg-secondary-800 border-b border-secondary-200 dark:border-secondary-700 px-6 py-4">
      <div className="flex items-center justify-between">
        {/* Left side - Page title */}
        <div className="flex items-center space-x-4">
          <h1 className="text-2xl font-bold text-secondary-900 dark:text-white">
            SecureMail
          </h1>
          <div className="hidden md:block">
            <div className="h-6 w-px bg-secondary-300 dark:bg-secondary-600"></div>
          </div>
          <div className="hidden md:block">
            <p className="text-sm text-secondary-600 dark:text-secondary-400">
              Your privacy-first email solution
            </p>
          </div>
        </div>

        {/* Right side - User actions */}
        <div className="flex items-center space-x-4">
          {/* Theme Toggle */}
          <button
            onClick={onToggleTheme}
            className="p-2 rounded-lg text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 transition-colors duration-200"
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {theme === 'dark' ? (
              <SunIcon className="w-5 h-5" />
            ) : (
              <MoonIcon className="w-5 h-5" />
            )}
          </button>

          {/* Notifications */}
          <button className="p-2 rounded-lg text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 transition-colors duration-200 relative">
            <BellIcon className="w-5 h-5" />
            <span className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full"></span>
          </button>

          {/* User Menu */}
          <div className="relative">
            <button className="flex items-center space-x-3 p-2 rounded-lg text-secondary-600 hover:bg-secondary-100 dark:text-secondary-400 dark:hover:bg-secondary-700 transition-colors duration-200">
              <UserCircleIcon className="w-8 h-8" />
              <div className="hidden md:block text-left">
                <p className="text-sm font-medium text-secondary-900 dark:text-white">
                  {user?.email || 'User'}
                </p>
                <p className="text-xs text-secondary-500 dark:text-secondary-400">
                  {user?.email || 'user@securesystem.email'}
                </p>
              </div>
            </button>

            {/* Dropdown Menu */}
            <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-secondary-800 rounded-lg shadow-lg border border-secondary-200 dark:border-secondary-700 py-1 z-50">
              <button
                onClick={onLogout}
                className="w-full text-left px-4 py-2 text-sm text-secondary-700 hover:bg-secondary-100 dark:text-secondary-300 dark:hover:bg-secondary-700 transition-colors duration-200"
              >
                Sign out
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header; 