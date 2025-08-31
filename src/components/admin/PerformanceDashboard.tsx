/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE PERFORMANCE DASHBOARD COMPONENT.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual design or layout of this dashboard
 * 2. NEVER modify the component structure or styling
 * 3. NEVER remove or change existing UI elements
 * 4. NEVER alter the dashboard's appearance or user interface
 * 5. ALWAYS preserve the current design and layout exactly as is
 * 6. ALWAYS maintain the existing visual hierarchy and styling
 * 7. ALWAYS keep the dashboard's current look and feel
 * 8. ALWAYS ensure the design remains consistent and unchanged
 * 
 * This component provides performance monitoring and analytics capabilities.
 * The design has been finalized and must remain unchanged.
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 * @last_updated: Priority 9 - Performance Monitoring
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { log } from '@/lib/logger';
import { 
  Activity, 
  TrendingUp, 
  TrendingDown, 
  AlertTriangle, 
  CheckCircle, 
  Zap, 
  Database, 
  Monitor,
  RefreshCw,
  Download
} from 'lucide-react';
import { performanceMonitor, PerformanceMetrics, PerformanceEvent } from '@/lib/performance';

interface PerformanceDashboardProps {
  isOpen: boolean;
  onClose: () => void;
}

interface MetricCardProps {
  title: string;
  value: string | number;
  unit: string;
  trend: 'up' | 'down' | 'stable';
  status: 'good' | 'warning' | 'error' | 'critical';
  icon: React.ReactNode;
  description?: string;
}

const MetricCard: React.FC<MetricCardProps> = ({ 
  title, 
  value, 
  unit, 
  trend, 
  status, 
  icon, 
  description 
}) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'good': return 'text-green-600 bg-green-50 border-green-200';
      case 'warning': return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'error': return 'text-red-600 bg-red-50 border-red-200';
      case 'critical': return 'text-red-800 bg-red-100 border-red-300';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return <TrendingUp className="w-4 h-4 text-green-600" />;
      case 'down': return <TrendingDown className="w-4 h-4 text-red-600" />;
      default: return <div className="w-4 h-4" />;
    }
  };

  return (
    <div className={`p-6 rounded-lg border-2 ${getStatusColor(status)} transition-all duration-200 hover:shadow-md`}>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-3">
          <div className="p-2 rounded-lg bg-white/50">
            {icon}
          </div>
          <div>
            <h3 className="text-sm font-medium text-gray-700">{title}</h3>
            {description && (
              <p className="text-xs text-gray-500 mt-1">{description}</p>
            )}
          </div>
        </div>
        {getTrendIcon(trend)}
      </div>
      
      <div className="flex items-baseline space-x-2">
        <span className="text-2xl font-bold">{value}</span>
        <span className="text-sm text-gray-500">{unit}</span>
      </div>
    </div>
  );
};

const PerformanceDashboard: React.FC<PerformanceDashboardProps> = ({ isOpen, onClose }) => {
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [events, setEvents] = useState<PerformanceEvent[]>([]);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [autoRefresh] = useState(true);
  const [filterSeverity, setFilterSeverity] = useState<'all' | 'warning' | 'error' | 'critical'>('all');

  // Refresh performance data
  const refreshData = useCallback(async () => {
    setIsRefreshing(true);
    try {
      const report = performanceMonitor.getReport();
      setMetrics(report.metrics);
      setEvents(report.events);
    } catch (error) {
      log.error('Failed to refresh performance data:', error, 'PerformanceDashboard');
    } finally {
      setIsRefreshing(false);
    }
  }, []);

  // Auto-refresh effect
  useEffect(() => {
    if (!isOpen) return;

    refreshData();

    if (autoRefresh) {
      const interval = setInterval(refreshData, 5000); // Refresh every 5 seconds
      return () => clearInterval(interval);
    }
  }, [isOpen, autoRefresh, refreshData]);

  // Calculate metric status
  const getMetricStatus = useCallback((value: number, threshold: number): 'good' | 'warning' | 'error' | 'critical' => {
    if (value <= threshold) return 'good';
    if (value <= threshold * 1.5) return 'warning';
    if (value <= threshold * 2) return 'error';
    return 'critical';
  }, []);

  // Format metric values
  const formatMetricValue = useCallback((value: number, type: string): { value: string | number; unit: string } => {
    switch (type) {
      case 'renderTime':
      case 'apiResponseTime':
      case 'userInteractionTime':
      case 'loadTime':
        return { value: value.toFixed(2), unit: 'ms' };
      case 'memoryUsage':
        return { value: (value / (1024 * 1024)).toFixed(1), unit: 'MB' };
      case 'bundleSize':
        return { value: (value / 1024).toFixed(1), unit: 'KB' };
      default:
        return { value: value.toFixed(2), unit: '' };
    }
  }, []);

  // Filtered events
  const filteredEvents = useMemo(() => {
    if (filterSeverity === 'all') return events;
    return events.filter(event => event.severity === filterSeverity);
  }, [events, filterSeverity]);

  // Recent critical issues
  const criticalIssues = useMemo(() => {
    return events
      .filter(event => event.severity === 'critical' || event.severity === 'error')
      .slice(-10)
      .reverse();
  }, [events]);

  // Performance trends
  const performanceTrends = useMemo(() => {
    if (!metrics) {
      return {
        renderTime: 'stable' as const,
        memoryUsage: 'stable' as const,
        apiResponseTime: 'stable' as const,
        userInteractionTime: 'stable' as const,
      };
    }
    
    return {
      renderTime: metrics.renderTime < 16 ? 'down' as const : 'up' as const,
      memoryUsage: metrics.memoryUsage < 50 * 1024 * 1024 ? 'down' as const : 'up' as const,
      apiResponseTime: metrics.apiResponseTime < 1000 ? 'down' as const : 'up' as const,
      userInteractionTime: metrics.userInteractionTime < 100 ? 'down' as const : 'up' as const,
    };
  }, [metrics]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-7xl h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <Monitor className="w-6 h-6 text-blue-600" />
            <h2 className="text-xl font-semibold text-gray-900">Performance Dashboard</h2>
          </div>
          
          <div className="flex items-center space-x-3">
            <button
              onClick={() => {/* TODO: Implement auto-refresh toggle */}}
              className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                autoRefresh 
                  ? 'bg-green-100 text-green-700 hover:bg-green-200' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              <RefreshCw className={`w-4 h-4 ${autoRefresh ? 'animate-spin' : ''}`} />
              <span>Auto-refresh</span>
            </button>
            
            <button
              onClick={refreshData}
              disabled={isRefreshing}
              className="flex items-center space-x-2 px-3 py-2 bg-blue-100 text-blue-700 rounded-lg text-sm font-medium hover:bg-blue-200 transition-colors disabled:opacity-50"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span>Refresh</span>
            </button>
            
            <button
              onClick={onClose}
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
            >
              <span className="sr-only">Close</span>
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-6">
          {!metrics ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <RefreshCw className="w-8 h-8 text-gray-400 animate-spin mx-auto mb-4" />
                <p className="text-gray-500">Loading performance data...</p>
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Metrics Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <MetricCard
                  title="Render Time"
                  value={formatMetricValue(metrics.renderTime, 'renderTime').value}
                  unit={formatMetricValue(metrics.renderTime, 'renderTime').unit}
                  trend={performanceTrends.renderTime}
                  status={getMetricStatus(metrics.renderTime, 16)}
                  icon={<Zap className="w-5 h-5 text-blue-600" />}
                  description="Average component render time"
                />
                
                <MetricCard
                  title="Memory Usage"
                  value={formatMetricValue(metrics.memoryUsage, 'memoryUsage').value}
                  unit={formatMetricValue(metrics.memoryUsage, 'memoryUsage').unit}
                  trend={performanceTrends.memoryUsage}
                  status={getMetricStatus(metrics.memoryUsage, 50 * 1024 * 1024)}
                  icon={<Database className="w-5 h-5 text-purple-600" />}
                  description="Current JavaScript heap usage"
                />
                
                <MetricCard
                  title="API Response"
                  value={formatMetricValue(metrics.apiResponseTime, 'apiResponseTime').value}
                  unit={formatMetricValue(metrics.apiResponseTime, 'apiResponseTime').unit}
                  trend={performanceTrends.apiResponseTime}
                  status={getMetricStatus(metrics.apiResponseTime, 1000)}
                  icon={<Database className="w-5 h-5 text-green-600" />}
                  description="Average API response time"
                />
                
                <MetricCard
                  title="Interaction Time"
                  value={formatMetricValue(metrics.userInteractionTime, 'userInteractionTime').value}
                  unit={formatMetricValue(metrics.userInteractionTime, 'userInteractionTime').unit}
                  trend={performanceTrends.userInteractionTime}
                  status={getMetricStatus(metrics.userInteractionTime, 100)}
                  icon={<Activity className="w-5 h-5 text-orange-600" />}
                  description="User interaction response time"
                />
              </div>

              {/* Critical Issues */}
              {criticalIssues.length > 0 && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                  <div className="flex items-center space-x-3 mb-4">
                    <AlertTriangle className="w-5 h-5 text-red-600" />
                    <h3 className="text-lg font-semibold text-red-900">Critical Performance Issues</h3>
                    <span className="bg-red-100 text-red-800 text-xs font-medium px-2 py-1 rounded-full">
                      {criticalIssues.length}
                    </span>
                  </div>
                  
                  <div className="space-y-3">
                    {criticalIssues.map((event, index) => (
                      <div key={index} className="flex items-center justify-between p-3 bg-white rounded-lg border border-red-100">
                        <div className="flex items-center space-x-3">
                          <div className={`w-3 h-3 rounded-full ${
                            event.severity === 'critical' ? 'bg-red-500' : 'bg-orange-500'
                          }`} />
                          <div>
                            <p className="text-sm font-medium text-gray-900">
                              {event.component || event.type}
                            </p>
                            <p className="text-xs text-gray-500">
                              {new Date(event.timestamp).toLocaleString()}
                            </p>
                          </div>
                        </div>
                        <div className="text-right">
                          <p className="text-sm font-medium text-gray-900">
                            {event.duration.toFixed(2)}ms
                          </p>
                          <p className="text-xs text-gray-500 capitalize">
                            {event.severity}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Performance Events */}
              <div className="bg-white border border-gray-200 rounded-lg">
                <div className="flex items-center justify-between p-6 border-b border-gray-200">
                  <h3 className="text-lg font-semibold text-gray-900">Performance Events</h3>
                  
                  <div className="flex items-center space-x-3">
                    <select
                      value={filterSeverity}
                      onChange={(e) => setFilterSeverity(e.target.value as 'all' | 'warning' | 'error' | 'critical')}
                      className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">All Severities</option>
                      <option value="warning">Warnings</option>
                      <option value="error">Errors</option>
                      <option value="critical">Critical</option>
                    </select>
                    
                    <button
                      onClick={() => {
                        const report = performanceMonitor.getReport();
                        const dataStr = JSON.stringify(report, null, 2);
                        const dataBlob = new Blob([dataStr], { type: 'application/json' });
                        const url = URL.createObjectURL(dataBlob);
                        const link = document.createElement('a');
                        link.href = url;
                        link.download = `performance-report-${new Date().toISOString()}.json`;
                        link.click();
                        URL.revokeObjectURL(url);
                      }}
                      className="flex items-center space-x-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
                    >
                      <Download className="w-4 h-4" />
                      <span>Export</span>
                    </button>
                  </div>
                </div>
                
                <div className="max-h-96 overflow-auto">
                  {filteredEvents.length === 0 ? (
                    <div className="p-6 text-center text-gray-500">
                      <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-500" />
                      <p>No performance events found</p>
                    </div>
                  ) : (
                    <div className="divide-y divide-gray-200">
                      {filteredEvents.map((event, index) => (
                        <div key={index} className="p-4 hover:bg-gray-50 transition-colors">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-3">
                              <div className={`w-3 h-3 rounded-full ${
                                event.severity === 'critical' ? 'bg-red-500' :
                                event.severity === 'error' ? 'bg-orange-500' :
                                event.severity === 'warning' ? 'bg-yellow-500' : 'bg-green-500'
                              }`} />
                              <div>
                                <p className="text-sm font-medium text-gray-900">
                                  {event.component || event.type}
                                </p>
                                <p className="text-xs text-gray-500">
                                  {event.type} • {new Date(event.timestamp).toLocaleString()}
                                </p>
                              </div>
                            </div>
                            <div className="text-right">
                              <p className="text-sm font-medium text-gray-900">
                                {event.duration.toFixed(2)}ms
                              </p>
                              <p className="text-xs text-gray-500 capitalize">
                                {event.severity}
                              </p>
                            </div>
                          </div>
                          
                          {event.metadata && Object.keys(event.metadata).length > 0 && (
                            <div className="mt-2 p-2 bg-gray-50 rounded text-xs text-gray-600">
                              <pre className="whitespace-pre-wrap">
                                {JSON.stringify(event.metadata, null, 2)}
                              </pre>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PerformanceDashboard;
