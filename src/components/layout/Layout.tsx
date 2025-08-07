import React from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { useTheme } from '@/hooks/useTheme';
import Sidebar from './Sidebar';
import Header from './Header';

/**
 * Layout Component
 * 
 * The main layout wrapper that provides the application structure.
 * Includes a sidebar, top navigation bar, and main content area.
 * 
 * Features:
 * - Responsive sidebar with navigation
 * - Top header with user info and actions
 * - Main content area with outlet for routes
 * - Dark/light theme support
 * - Mobile-responsive design
 */
const Layout: React.FC = () => {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-secondary-50 dark:bg-secondary-900">
      {/* Sidebar */}
      <Sidebar />
      
      {/* Main Content Area */}
      <div className="lg:ml-64">
        {/* Top Navigation Bar */}
        <Header 
          user={user}
          onLogout={handleLogout}
          onToggleTheme={toggleTheme}
          theme={theme}
        />
        
        {/* Main Content */}
        <main className="p-6">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout; 