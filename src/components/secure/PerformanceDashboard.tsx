/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE PERFORMANCE DASHBOARD COMPONENT.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual layout or design of the dashboard
 * 2. NEVER modify the component structure or styling
 * 3. NEVER alter the color scheme or Tailwind classes
 * 4. ONLY add new functionality that doesn't affect the visual design
 * 5. ALWAYS maintain the exact same visual appearance
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE DESIGN, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 */

import React, { useState, useEffect } from 'react';
import { log } from '@/lib/logger';
import { 
  Activity, 
  Clock, 
  HardDrive, 
  Wifi, 
  AlertTriangle, 
  CheckCircle, 
  XCircle,
  Download,
  Trash2,
  RefreshCw
} from 'lucide-react';
import { performanceMonitor } from '@/lib/performance';

interface PerformanceDashboardProps {
  className?: string;
}

interface MetricCardProps {
  title: string;
  value: string;
  unit: string;
  icon: React.ReactNode;
  status: 'good' | 'warning' | 'error';
  description?: string;
}

const MetricCard: React.FC<MetricCardProps> = ({ 
  title, 
  value, 
  unit, 
  icon, 
  status, 
  description 
}) => {
  const getStatusColor = () => {
    switch (status) {
      case 'good': return 'text-green-600 dark:text-green-400';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400';
      case 'error': return 'text-red-600 dark:text-red-400';
      default: return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getStatusBg = () => {
    switch (status) {
      case 'good': return 'bg-green-50 dark:bg-green-900/20';
      case 'warning': return 'bg-yellow-50 dark:bg-yellow-900/20';
      case 'error': return 'bg-red-50 dark:bg-red-900/20';
      default: return 'bg-gray-50 dark:bg-gray-900/20';
    }
  };

  return (
    <div className={`p-4 rounded-lg border border-secondary-200 dark:border-secondary-700 ${getStatusBg()}`}>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          <div className={`p-2 rounded-lg ${getStatusColor()} bg-white dark:bg-secondary-800`}>
            {icon}
          </div>
          <h3 className="text-sm font-medium text-secondary-900 dark:text-secondary-100">
            {title}
          </h3>
        </div>
        <div className="flex items-center space-x-1">
          {status === 'good' && <CheckCircle className="w-4 h-4 text-green-600 dark:text-green-400" />}
          {status === 'warning' && <AlertTriangle className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />}
          {status === 'error' && <XCircle className="w-4 h-4 text-red-600 dark:text-red-400" />}
        </div>
      </div>
      <div className="flex items-baseline space-x-1">
        <span className="text-2xl font-bold text-secondary-900 dark:text-secondary-100">
          {value}
        </span>
        <span className="text-sm text-secondary-600 dark:text-secondary-400">
          {unit}
        </span>
      </div>
      {description && (
        <p className="text-xs text-secondary-600 dark:text-secondary-400 mt-1">
          {description}
        </p>
      )}
    </div>
  );
};

const PerformanceDashboard: React.FC<PerformanceDashboardProps> = ({ className = '' }) => {
  const [metrics, setMetrics] = useState<unknown | null>(null);
  const [events, setEvents] = useState<unknown[]>([]);
  const [report, setReport] = useState<unknown>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [autoRefresh] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [metricsData, reportData] = await Promise.all([
        performanceMonitor.getMetrics(),
        performanceMonitor.getReport()
      ]);

      setMetrics(metricsData);
      setEvents(reportData.events || []);
      setReport(reportData);
    } catch (err) {
      log.error('Error loading performance data', err, 'PerformanceDashboard');
      setError('Error loading performance data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();

    if (autoRefresh) {
      const interval = setInterval(loadData, 30000); // Refresh every 30 seconds
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  const handleClearEvents = async () => {
    try {
      await performanceMonitor.clearEvents();
      setEvents([]);
    } catch (err) {
      log.error('Error clearing events', err, 'PerformanceDashboard');
    }
  };

  const handleExportData = () => {
    try {
      const exportData = {
        timestamp: new Date().toISOString(),
        metrics,
        events,
        report
      };

      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json'
      });

      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `performance-data-${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      log.error('Error exporting data', err, 'PerformanceDashboard');
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'error': return 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/20';
      case 'info': return 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20';
      default: return 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-900/20';
    }
  };

  const filteredEvents = events.filter(event => 
    severityFilter === 'all' || (event as { severity: string }).severity === severityFilter
  );

  if (loading) {
    return (
      <div className={`p-6 ${className || ''}`}>
        <div className="flex items-center justify-center h-64">
          <div className="flex items-center space-x-2">
            <RefreshCw className="w-5 h-5 animate-spin text-primary-600" />
            <span className="text-secondary-600 dark:text-secondary-400">Loading metrics...</span>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`p-6 ${className || ''}`}>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <XCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <p className="text-red-600 dark:text-red-400">{error}</p>
            <button
              onClick={loadData}
              className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`p-6 space-y-6 ${className || ''}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-secondary-900 dark:text-secondary-100">
            Performance Dashboard
          </h1>
          <p className="text-secondary-600 dark:text-secondary-400">
            Real-time performance metrics and monitoring
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <button
            onClick={loadData}
            className="flex items-center space-x-2 px-3 py-2 bg-secondary-100 dark:bg-secondary-800 text-secondary-700 dark:text-secondary-300 rounded-lg hover:bg-secondary-200 dark:hover:bg-secondary-700 transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Refresh</span>
          </button>
          <button
            onClick={handleExportData}
            className="flex items-center space-x-2 px-3 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            <Download className="w-4 h-4" />
            <span>Export Data</span>
          </button>
        </div>
      </div>

      {/* Metrics Grid */}
      {metrics !== null && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricCard
            title="Render Time"
            value={(metrics as { renderTime?: number }).renderTime?.toFixed(1) || '0'}
            unit="ms"
            icon={<Activity className="w-5 h-5" />}
            status={((metrics as { renderTime?: number }).renderTime || 0) > 16 ? 'warning' : 'good'}
            description="Component render performance"
          />
          <MetricCard
            title="Memory Usage"
            value={(metrics as { memoryUsage?: number }).memoryUsage?.toFixed(1) || '0'}
            unit="MB"
            icon={<HardDrive className="w-5 h-5" />}
            status={((metrics as { memoryUsage?: number }).memoryUsage || 0) > 100 ? 'warning' : 'good'}
            description="Current memory consumption"
          />
          <MetricCard
            title="API Response Time"
            value={(metrics as { apiResponseTime?: number }).apiResponseTime?.toFixed(0) || '0'}
            unit="ms"
            icon={<Wifi className="w-5 h-5" />}
            status={((metrics as { apiResponseTime?: number }).apiResponseTime || 0) > 1000 ? 'error' : ((metrics as { apiResponseTime?: number }).apiResponseTime || 0) > 500 ? 'warning' : 'good'}
            description="Average API response time"
          />
          <MetricCard
            title="User Interaction"
            value={(metrics as { userInteractionTime?: number }).userInteractionTime?.toFixed(1) || '0'}
            unit="ms"
            icon={<Clock className="w-5 h-5" />}
            status={((metrics as { userInteractionTime?: number }).userInteractionTime || 0) > 100 ? 'warning' : 'good'}
            description="User interaction responsiveness"
          />
        </div>
      )}

      {/* Events Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
            Performance Events
          </h2>
          <div className="flex items-center space-x-3">
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              className="px-3 py-1 border border-secondary-300 dark:border-secondary-600 rounded-lg bg-white dark:bg-secondary-800 text-secondary-900 dark:text-secondary-100"
            >
              <option value="all">All Severities</option>
              <option value="error">Error</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
            <button
              onClick={handleClearEvents}
              className="flex items-center space-x-2 px-3 py-1 bg-red-100 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg hover:bg-red-200 dark:hover:bg-red-900/40 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
              <span>Clear Events</span>
            </button>
          </div>
        </div>

        <div className="space-y-2 max-h-64 overflow-y-auto">
          {filteredEvents.length === 0 ? (
            <p className="text-secondary-500 dark:text-secondary-400 text-center py-8">
              No performance events found
            </p>
          ) : (
            filteredEvents.map((event, index) => (
              <div
                key={index}
                className={`p-3 rounded-lg border border-secondary-200 dark:border-secondary-700 ${getSeverityColor((event as { severity: string }).severity)}`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <span className="text-xs font-medium uppercase tracking-wide">
                      {(event as { severity: string }).severity}
                    </span>
                    <span className="text-sm font-medium">
                      {(event as { component?: string; type?: string }).component || (event as { component?: string; type?: string }).type}
                    </span>
                  </div>
                  <span className="text-xs text-secondary-600 dark:text-secondary-400">
                    {new Date((event as { timestamp: number }).timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className="mt-1 text-sm">
                  <span className="font-medium">{(event as { duration?: number }).duration?.toFixed(2)}ms</span>
                  {(event as { metadata?: Record<string, unknown> }).metadata && (
                    <span className="text-secondary-600 dark:text-secondary-400 ml-2">
                      {Object.entries((event as { metadata?: Record<string, unknown> }).metadata || {}).map(([key, value]) => `${key}: ${value}`).join(', ')}
                    </span>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Recommendations */}
      {(report as { recommendations?: string[] })?.recommendations && ((report as { recommendations?: string[] }).recommendations?.length || 0) > 0 && (
        <div className="space-y-3">
          <h2 className="text-lg font-semibold text-secondary-900 dark:text-secondary-100">
            Recommendations
          </h2>
          <div className="space-y-2">
            {(report as { recommendations?: string[] }).recommendations?.map((recommendation: string, index: number) => (
              <div
                key={index}
                className="flex items-start space-x-3 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg"
              >
                <AlertTriangle className="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                <p className="text-sm text-blue-800 dark:text-blue-200">
                  {recommendation}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default PerformanceDashboard;
