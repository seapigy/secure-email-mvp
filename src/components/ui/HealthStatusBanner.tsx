import React from 'react';
import { 
  CheckCircleIcon, 
  ExclamationTriangleIcon, 
  XCircleIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import { useHealthCheck } from '@/hooks/useHealthCheck';

/**
 * HealthStatusBanner Component
 * 
 * Displays the backend connection status with appropriate styling and actions.
 * Shows loading, success, and error states with user-friendly messages.
 * 
 * Features:
 * - Real-time health status display
 * - Manual refresh capability
 * - Responsive design
 * - Accessibility support
 */
const HealthStatusBanner: React.FC = () => {
  const { isHealthy, isLoading, error, lastChecked, checkHealth } = useHealthCheck();

  const getStatusIcon = () => {
    if (isLoading) {
      return <ArrowPathIcon className="w-5 h-5 animate-spin" />;
    }
    if (isHealthy) {
      return <CheckCircleIcon className="w-5 h-5" />;
    }
    return <XCircleIcon className="w-5 h-5" />;
  };

  const getStatusColor = () => {
    if (isLoading) {
      return 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-300';
    }
    if (isHealthy) {
      return 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300';
    }
    return 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300';
  };

  const getStatusText = () => {
    if (isLoading) {
      return 'Checking backend connection...';
    }
    if (isHealthy) {
      return 'Backend connection healthy';
    }
    return error || 'Backend connection failed';
  };

  const getStatusDetails = () => {
    if (isLoading) {
      return 'Connecting to backend server...';
    }
    if (isHealthy && lastChecked) {
      return `Server is responding normally. Last checked: ${lastChecked.toLocaleTimeString()}`;
    }
    if (error) {
      return `Error: ${error}`;
    }
    return 'Unable to connect to backend server';
  };

  const getLastCheckedText = () => {
    if (!lastChecked) return '';
    return `Last checked: ${lastChecked.toLocaleTimeString()}`;
  };

  return (
    <div className={`border rounded-lg p-4 mb-6 ${getStatusColor()}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          {getStatusIcon()}
          <div>
            <p className="font-medium">{getStatusText()}</p>
            <p className="text-sm opacity-75">{getStatusDetails()}</p>
          </div>
        </div>
        
        <button
          onClick={checkHealth}
          disabled={isLoading}
          className="flex items-center space-x-2 px-3 py-1 text-sm font-medium rounded-md bg-white/20 hover:bg-white/30 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
          title="Refresh health check"
        >
          <ArrowPathIcon className="w-4 h-4" />
          <span className="hidden sm:inline">Refresh</span>
        </button>
      </div>
    </div>
  );
};

export default HealthStatusBanner; 