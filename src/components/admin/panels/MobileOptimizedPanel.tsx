import React, { useState } from 'react';
import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';

interface MobileOptimizedPanelProps {
  title: string;
  children: React.ReactNode;
  defaultExpanded?: boolean;
  className?: string;
  icon?: React.ComponentType<{ className?: string }>;
  badge?: string;
  badgeColor?: 'green' | 'red' | 'yellow' | 'blue' | 'gray';
}

const MobileOptimizedPanel: React.FC<MobileOptimizedPanelProps> = ({
  title,
  children,
  defaultExpanded = true,
  className = '',
  icon: Icon,
  badge,
  badgeColor = 'gray'
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  const badgeColors = {
    green: 'bg-green-100 text-green-800',
    red: 'bg-red-100 text-red-800',
    yellow: 'bg-yellow-100 text-yellow-800',
    blue: 'bg-blue-100 text-blue-800',
    gray: 'bg-gray-100 text-gray-800'
  };

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 ${className}`}>
      {/* Mobile-optimized header with large tap target */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-4 py-4 sm:px-6 sm:py-4 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
        aria-expanded={isExpanded}
        aria-controls={`panel-${title.toLowerCase().replace(/\s+/g, '-')}`}
      >
        <div className="flex items-center space-x-3 min-w-0 flex-1">
          {Icon && (
            <Icon className="h-6 w-6 text-gray-400 flex-shrink-0" />
          )}
          <div className="min-w-0 flex-1">
            <h3 className="text-base sm:text-lg font-medium text-gray-900 dark:text-white truncate">
              {title}
            </h3>
          </div>
        </div>
        
        <div className="flex items-center space-x-2 flex-shrink-0">
          {badge && (
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${badgeColors[badgeColor]}`}>
              {badge}
            </span>
          )}
          {isExpanded ? (
            <ChevronUpIcon className="h-5 w-5 text-gray-400" />
          ) : (
            <ChevronDownIcon className="h-5 w-5 text-gray-400" />
          )}
        </div>
      </button>

      {/* Collapsible content */}
      <div
        id={`panel-${title.toLowerCase().replace(/\s+/g, '-')}`}
        className={`transition-all duration-200 ease-in-out ${
          isExpanded ? 'max-h-screen opacity-100' : 'max-h-0 opacity-0 overflow-hidden'
        }`}
      >
        <div className="px-4 pb-4 sm:px-6 sm:pb-6">
          {children}
        </div>
      </div>
    </div>
  );
};

// Mobile-optimized metric card component
interface MobileMetricCardProps {
  label: string;
  value: string | number;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon?: React.ComponentType<{ className?: string }>;
  className?: string;
}

export const MobileMetricCard: React.FC<MobileMetricCardProps> = ({
  label,
  value,
  change,
  changeType = 'neutral',
  icon: Icon,
  className = ''
}) => {
  const changeColors = {
    positive: 'text-green-600',
    negative: 'text-red-600',
    neutral: 'text-gray-600'
  };

  const changeIcons = {
    positive: '↗',
    negative: '↘',
    neutral: '→'
  };

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 ${className}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          {Icon && (
            <Icon className="h-5 w-5 text-gray-400" />
          )}
          <div>
            <p className="text-sm font-medium text-gray-600 dark:text-gray-400 truncate">
              {label}
            </p>
            <p className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">
              {value}
            </p>
          </div>
        </div>
        
        {change && (
          <div className={`flex items-center space-x-1 text-sm font-medium ${changeColors[changeType]}`}>
            <span>{changeIcons[changeType]}</span>
            <span>{change}</span>
          </div>
        )}
      </div>
    </div>
  );
};

// Mobile-optimized data table component
interface MobileDataTableProps {
  data: Array<Record<string, any>>;
  columns: Array<{
    key: string;
    label: string;
    render?: (value: any, row: Record<string, any>) => React.ReactNode;
  }>;
  className?: string;
  maxRows?: number;
}

export const MobileDataTable: React.FC<MobileDataTableProps> = ({
  data,
  columns,
  className = '',
  maxRows = 5
}) => {
  const [showAll, setShowAll] = useState(false);
  const displayData = showAll ? data : data.slice(0, maxRows);

  return (
    <div className={className}>
      {/* Mobile-optimized table */}
      <div className="space-y-3">
        {displayData.map((row, index) => (
          <div
            key={index}
            className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 space-y-2"
          >
            {columns.map((column) => (
              <div key={column.key} className="flex justify-between items-center">
                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                  {column.label}:
                </span>
                <span className="text-sm text-gray-900 dark:text-white text-right">
                  {column.render ? column.render(row[column.key], row) : row[column.key]}
                </span>
              </div>
            ))}
          </div>
        ))}
      </div>

      {/* Show more/less button */}
      {data.length > maxRows && (
        <button
          onClick={() => setShowAll(!showAll)}
          className="mt-4 w-full px-4 py-2 text-sm font-medium text-blue-600 hover:text-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 rounded-md transition-colors duration-200"
        >
          {showAll ? 'Show Less' : `Show ${data.length - maxRows} More`}
        </button>
      )}
    </div>
  );
};

// Mobile-optimized chart container
interface MobileChartContainerProps {
  title: string;
  children: React.ReactNode;
  className?: string;
  height?: string;
}

export const MobileChartContainer: React.FC<MobileChartContainerProps> = ({
  title,
  children,
  className = '',
  height = 'h-64 sm:h-80'
}) => {
  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 ${className}`}>
      <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-4">
        {title}
      </h4>
      <div className={`${height} w-full`}>
        {children}
      </div>
    </div>
  );
};

// Mobile-optimized action button
interface MobileActionButtonProps {
  onClick: () => void;
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  className?: string;
  icon?: React.ComponentType<{ className?: string }>;
}

export const MobileActionButton: React.FC<MobileActionButtonProps> = ({
  onClick,
  children,
  variant = 'primary',
  size = 'md',
  disabled = false,
  className = '',
  icon: Icon
}) => {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2';
  
  const variantClasses = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 focus:ring-gray-500',
    danger: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500'
  };

  const sizeClasses = {
    sm: 'px-3 py-2 text-sm',
    md: 'px-4 py-3 text-base',
    lg: 'px-6 py-4 text-lg'
  };

  const disabledClasses = disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer';

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${disabledClasses} ${className}`}
    >
      {Icon && (
        <Icon className="h-4 w-4 mr-2" />
      )}
      {children}
    </button>
  );
};

export default MobileOptimizedPanel;
