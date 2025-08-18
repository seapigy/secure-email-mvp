import React, { useState } from 'react';
import {
  ChartBarIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeIcon,
  EyeSlashIcon,
  ClockIcon,
  ServerIcon,
  ArrowTrendingUpIcon,
} from '@heroicons/react/24/outline';
import { PerformanceOperationalMetrics } from '../../../types/admin';

interface PerformanceOperationalPanelProps {
  metrics: PerformanceOperationalMetrics | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const PerformanceOperationalPanel: React.FC<PerformanceOperationalPanelProps> = ({
  metrics,
  isLoading,
  onRefresh,
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [showEnhancedMetrics, setShowEnhancedMetrics] = useState(false);

  if (!metrics) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Performance & Operational</h3>
          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded-md disabled:opacity-50"
          >
            <ArrowPathIcon className={`h-5 w-5 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
        <div className="text-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-500">Loading performance metrics...</p>
        </div>
      </div>
    );
  }



  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-green-600 bg-green-100';
      case 'degraded':
        return 'text-yellow-600 bg-yellow-100';
      case 'critical':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };



  const getPerformanceColor = (value: number, threshold: number) => {
    if (value <= threshold * 0.7) return 'text-green-600';
    if (value <= threshold) return 'text-yellow-600';
    return 'text-red-600';
  };



  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-6 w-6 text-blue-600" />
            <h3 className="text-lg font-medium text-gray-900">Performance & Operational</h3>
            <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(metrics.system_health.health_check_status)}`}>
              {metrics.system_health.health_check_status.toUpperCase()}
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowDetails(!showDetails)}
              className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
            >
              {showDetails ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
            </button>
            <button
              onClick={onRefresh}
              disabled={isLoading}
              className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded disabled:opacity-50"
            >
              <ArrowPathIcon className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-6">
        {/* API Performance */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">API Performance</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ChartBarIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Total Requests</span>
              </div>
              <div className="text-2xl font-bold text-blue-600">
                {metrics.api_performance.total_requests.toLocaleString()}
              </div>
              <div className="text-xs text-blue-500 mt-1">
                {metrics.api_performance.requests_per_second.toFixed(1)} req/s
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Success Rate</span>
              </div>
              <div className="text-2xl font-bold text-green-600">
                {((metrics.api_performance.successful_requests / metrics.api_performance.total_requests) * 100).toFixed(1)}%
              </div>
              <div className="text-xs text-green-500 mt-1">
                {metrics.api_performance.successful_requests.toLocaleString()} successful
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">Avg Response Time</span>
              </div>
              <div className={`text-lg font-bold ${getPerformanceColor(metrics.api_performance.average_response_time_ms, 500)}`}>
                {metrics.api_performance.average_response_time_ms}ms
              </div>
              <div className="text-xs text-gray-500 mt-1">
                P95: {metrics.api_performance.p95_response_time_ms}ms
              </div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Failed Requests</span>
              </div>
              <div className="text-lg font-bold text-red-600">
                {metrics.api_performance.failed_requests.toLocaleString()}
              </div>
              <div className="text-xs text-red-500 mt-1">
                {((metrics.api_performance.failed_requests / metrics.api_performance.total_requests) * 100).toFixed(2)}% error rate
              </div>
            </div>
          </div>
        </div>

        {/* Endpoint Metrics */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Endpoint Performance</h4>
          <div className="space-y-3">
            {Object.entries(metrics.endpoint_metrics).map(([endpoint, metric]) => (
              <div key={endpoint} className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-gray-700">{endpoint}</span>
                  <div className={`px-2 py-1 rounded-full text-xs font-medium ${
                    metric.success_rate >= 99 ? 'bg-green-100 text-green-800' :
                    metric.success_rate >= 95 ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'
                  }`}>
                    {metric.success_rate.toFixed(1)}%
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4 text-xs">
                  <div>
                    <span className="text-gray-500">Requests:</span>
                    <span className="font-medium ml-1">{metric.requests.toLocaleString()}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Errors:</span>
                    <span className="font-medium ml-1">{metric.errors}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Latency:</span>
                    <span className={`font-medium ml-1 ${getPerformanceColor(metric.average_latency_ms, 500)}`}>
                      {metric.average_latency_ms}ms
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Error Tracking */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Error Tracking</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Total Errors</span>
              </div>
              <div className="text-2xl font-bold text-red-600">
                {metrics.error_tracking.total_errors.toLocaleString()}
              </div>
              <div className="text-xs text-red-500 mt-1">
                {metrics.error_tracking.error_rate_percentage.toFixed(2)}% error rate
              </div>
            </div>
            <div className="bg-orange-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-orange-600" />
                <span className="text-xs font-medium text-orange-900">Critical Errors</span>
              </div>
              <div className="text-2xl font-bold text-orange-600">
                {metrics.error_tracking.critical_errors}
              </div>
            </div>
          </div>

          {/* Recent Errors */}
          {showDetails && metrics.error_tracking.recent_errors.length > 0 && (
            <div className="mt-4 bg-gray-50 rounded-lg p-3">
              <h5 className="text-xs font-medium text-gray-700 mb-2">Recent Errors</h5>
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {metrics.error_tracking.recent_errors.slice(0, 3).map((error) => (
                  <div key={error.id} className="text-xs">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-gray-700">{error.error_type}</span>
                      <span className={`px-2 py-1 rounded text-xs ${
                        error.severity === 'critical' ? 'bg-red-100 text-red-800' :
                        error.severity === 'high' ? 'bg-orange-100 text-orange-800' :
                        error.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' : 'bg-blue-100 text-blue-800'
                      }`}>
                        {error.severity}
                      </span>
                    </div>
                    <p className="text-gray-600 mt-1">{error.message}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Load Testing */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Load Testing</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-purple-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ServerIcon className="h-4 w-4 text-purple-600" />
                <span className="text-xs font-medium text-purple-900">Concurrent Users</span>
              </div>
              <div className="text-lg font-bold text-purple-600">
                {metrics.load_testing.concurrent_users}
              </div>
              <div className="text-xs text-purple-500 mt-1">
                Max: {metrics.load_testing.max_concurrent_users}
              </div>
            </div>
            <div className="bg-indigo-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ArrowTrendingUpIcon className="h-4 w-4 text-indigo-600" />
                <span className="text-xs font-medium text-indigo-900">Throughput</span>
              </div>
              <div className="text-lg font-bold text-indigo-600">
                {metrics.load_testing.throughput_requests_per_second.toFixed(1)} req/s
              </div>
            </div>
          </div>

          {/* Load Test Results */}
          {showDetails && metrics.load_test_results.length > 0 && (
            <div className="mt-4 bg-gray-50 rounded-lg p-3">
              <h5 className="text-xs font-medium text-gray-700 mb-2">Recent Load Tests</h5>
              <div className="space-y-2">
                {metrics.load_test_results.slice(0, 2).map((result: any) => (
                  <div key={result.id} className="text-xs bg-white rounded p-2">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-gray-700">
                        {result.concurrent_users} users
                      </span>
                      <div className={`px-2 py-1 rounded text-xs ${
                        result.success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                      }`}>
                        {result.success ? 'Passed' : 'Failed'}
                      </div>
                    </div>
                    <div className="grid grid-cols-3 gap-2 mt-1 text-gray-600">
                      <div>{result.requests_per_second.toFixed(1)} req/s</div>
                      <div>{result.average_response_time_ms}ms</div>
                      <div>{(result.error_rate * 100).toFixed(1)}% errors</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* System Health */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-3">System Health</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Uptime</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.system_health.uptime_percentage.toFixed(2)}%
              </div>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Last Restart</span>
              </div>
              <div className="text-xs font-medium text-blue-600">
                {new Date(metrics.system_health.last_restart).toLocaleDateString()}
              </div>
            </div>
          </div>

          {/* Dependency Status */}
          <div className="mt-4">
            <h5 className="text-xs font-medium text-gray-700 mb-2">Dependency Status</h5>
            <div className="space-y-2">
              {Object.entries(metrics.system_health.dependency_status).map(([dependency, status]) => (
                <div key={dependency} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <span className="text-xs font-medium text-gray-700 capitalize">{dependency}</span>
                  <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(status)}`}>
                    {status}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Enhanced Performance Monitoring */}
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Enhanced Performance Monitoring</h4>
            <button
              onClick={() => setShowEnhancedMetrics(!showEnhancedMetrics)}
              className="text-xs text-blue-600 hover:text-blue-800 focus:outline-none"
            >
              {showEnhancedMetrics ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          {showEnhancedMetrics && (
            <div className="space-y-4">
              {/* Real-Time API Latency */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Real-Time API Latency</h5>
                <div className="grid grid-cols-2 gap-3">
                  {Object.entries(metrics.real_time_api_latency).map(([endpoint, latency]) => (
                    <div key={endpoint} className="bg-gray-50 rounded-lg p-3">
                      <div className="text-xs text-gray-600 font-medium mb-2">{endpoint}</div>
                      <div className="grid grid-cols-2 gap-2 text-xs">
                        <div>
                          <span className="text-gray-500">Current:</span>
                          <span className="font-bold text-gray-900 ml-1">{latency.current_latency_ms}ms</span>
                        </div>
                        <div>
                          <span className="text-gray-500">P95:</span>
                          <span className="font-bold text-gray-900 ml-1">{latency.p95_latency_ms}ms</span>
                        </div>
                        <div>
                          <span className="text-gray-500">P99:</span>
                          <span className="font-bold text-gray-900 ml-1">{latency.p99_latency_ms}ms</span>
                        </div>
                        <div>
                          <span className="text-gray-500">Req/min:</span>
                          <span className="font-bold text-gray-900 ml-1">{latency.requests_per_minute}</span>
                        </div>
                      </div>
                      <div className="mt-2">
                        <span className="text-gray-500 text-xs">Error Rate:</span>
                        <span className={`font-bold text-xs ml-1 ${
                          latency.error_rate_percent > 5 ? 'text-red-600' : 
                          latency.error_rate_percent > 1 ? 'text-yellow-600' : 'text-green-600'
                        }`}>
                          {latency.error_rate_percent}%
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Session Metrics */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Session Metrics</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Concurrent User Sessions</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.session_metrics.concurrent_user_sessions}</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Active Admin Sessions</div>
                    <div className="text-lg font-bold text-blue-900">{metrics.session_metrics.active_admin_sessions}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Session Creation Rate</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.session_metrics.session_creation_rate_per_min}/min</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Session Timeouts</div>
                    <div className="text-lg font-bold text-yellow-900">{metrics.session_metrics.session_timeout_count}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Avg Session Duration</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.session_metrics.average_session_duration_min}min</div>
                  </div>
                </div>
              </div>

              {/* Database Performance */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Database Performance</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Encrypted Mappings QPS</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.database_performance.encrypted_mappings_queries_per_sec.toFixed(1)}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Average Query Time</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.database_performance.average_query_time_ms}ms</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Slow Queries</div>
                    <div className="text-lg font-bold text-red-900">{metrics.database_performance.slow_queries_count}</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Connection Pool Usage</div>
                    <div className="text-lg font-bold text-blue-900">{metrics.database_performance.connection_pool_usage_percent}%</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Database Latency</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.database_performance.database_latency_ms}ms</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Encryption Overhead</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.database_performance.encryption_overhead_ms}ms</div>
                  </div>
                </div>
              </div>

              {/* System Health Details */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">System Health Details</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">CPU Usage</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.cpu_usage_percent}%</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Memory Usage</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.memory_usage_percent}%</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Disk I/O Operations</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.disk_io_operations_per_sec}/sec</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Network Throughput</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.network_throughput_mbps} Mbps</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Active Goroutines</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.active_goroutines}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">GC Pause Time</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.system_health.gc_pause_time_ms}ms</div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Performance Summary */}
        <div className="mt-6 bg-blue-50 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <ChartBarIcon className="h-5 w-5 text-blue-600" />
            <div>
              <h5 className="text-sm font-medium text-blue-900">Performance Summary</h5>
              <p className="text-xs text-blue-700">
                Avg response: {metrics.api_performance.average_response_time_ms}ms | 
                Success rate: {((metrics.api_performance.successful_requests / metrics.api_performance.total_requests) * 100).toFixed(1)}% | 
                Error rate: {metrics.error_tracking.error_rate_percentage.toFixed(2)}% | 
                Uptime: {metrics.system_health.uptime_percentage.toFixed(2)}%
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PerformanceOperationalPanel;
