import React, { useState } from 'react';
import {
  EnvelopeIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeIcon,
  EyeSlashIcon,
  ChartBarIcon,
  ClockIcon,
  ServerIcon,
  CloudIcon,
  CpuChipIcon,
  WifiIcon,
} from '@heroicons/react/24/outline';
import { EmailDeliveryMetrics } from '../../../types/admin';

interface EmailDeliveryPanelProps {
  metrics: EmailDeliveryMetrics | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const EmailDeliveryPanel: React.FC<EmailDeliveryPanelProps> = ({
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
          <h3 className="text-lg font-medium text-gray-900">Email Delivery</h3>
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
          <p className="mt-2 text-sm text-gray-500">Loading email delivery metrics...</p>
        </div>
      </div>
    );
  }

  const getQueueHealthColor = (percentage: number) => {
    if (percentage >= 90) return 'text-green-600';
    if (percentage >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getQueueHealthStatus = (percentage: number) => {
    if (percentage >= 90) return 'healthy';
    if (percentage >= 70) return 'warning';
    return 'critical';
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-green-600 bg-green-100';
      case 'warning':
        return 'text-yellow-600 bg-yellow-100';
      case 'critical':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };



  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <EnvelopeIcon className="h-6 w-6 text-blue-600" />
            <h3 className="text-lg font-medium text-gray-900">Email Delivery</h3>
            <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(getQueueHealthStatus(metrics.queue_status.queue_health_percentage))}`}>
              {getQueueHealthStatus(metrics.queue_status.queue_health_percentage).toUpperCase()}
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
        {/* Queue Status */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Queue Status</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Pending</span>
              </div>
              <div className="text-2xl font-bold text-blue-600">{metrics.queue_status.pending_emails}</div>
              <div className="text-xs text-blue-500 mt-1">
                of {metrics.queue_status.queue_size_limit} limit
              </div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ServerIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">Processing</span>
              </div>
              <div className="text-2xl font-bold text-yellow-600">{metrics.queue_status.processing_emails}</div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Failed</span>
              </div>
              <div className="text-2xl font-bold text-red-600">{metrics.queue_status.failed_emails}</div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Health</span>
              </div>
              <div className={`text-2xl font-bold ${getQueueHealthColor(metrics.queue_status.queue_health_percentage)}`}>
                {metrics.queue_status.queue_health_percentage.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>

        {/* Delivery Performance */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Delivery Performance</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Successful</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.delivery_performance.successful_deliveries.toLocaleString()}
              </div>
              <div className="text-xs text-green-500 mt-1">
                {metrics.delivery_performance.delivery_success_rate.toFixed(1)}% success rate
              </div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Failed</span>
              </div>
              <div className="text-lg font-bold text-red-600">
                {metrics.delivery_performance.failed_deliveries.toLocaleString()}
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">Avg Processing</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.delivery_performance.average_processing_time_ms}ms
              </div>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ArrowPathIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Retry Attempts</span>
              </div>
              <div className="text-lg font-bold text-blue-600">
                {metrics.delivery_performance.retry_attempts.toLocaleString()}
              </div>
            </div>
          </div>
        </div>

        {/* Storage Metrics */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Storage Metrics</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-purple-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CloudIcon className="h-4 w-4 text-purple-600" />
                <span className="text-xs font-medium text-purple-900">Storage Used</span>
              </div>
              <div className="text-lg font-bold text-purple-600">
                {metrics.storage_metrics.total_storage_used_gb.toFixed(1)} GB
              </div>
              <div className="text-xs text-purple-500 mt-1">
                of {metrics.storage_metrics.storage_limit_gb} GB limit
              </div>
            </div>
            <div className="bg-indigo-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <EnvelopeIcon className="h-4 w-4 text-indigo-600" />
                <span className="text-xs font-medium text-indigo-900">Encrypted Blobs</span>
              </div>
              <div className="text-lg font-bold text-indigo-600">
                {metrics.storage_metrics.encrypted_blobs.toLocaleString()}
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ChartBarIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">Avg Blob Size</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.storage_metrics.average_blob_size_kb.toFixed(1)} KB
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Storage Errors</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.storage_metrics.storage_errors}
              </div>
            </div>
          </div>

          {/* Storage Usage Bar */}
          <div className="mt-4 bg-gray-200 rounded-full h-2">
            <div
              className="bg-purple-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${metrics.storage_metrics.storage_usage_percentage}%` }}
            ></div>
          </div>
          <div className="flex justify-between text-xs text-gray-500 mt-1">
            <span>0 GB</span>
            <span>{metrics.storage_metrics.storage_usage_percentage.toFixed(1)}% used</span>
            <span>{metrics.storage_metrics.storage_limit_gb} GB</span>
          </div>
        </div>

        {/* System Resources */}
        {showDetails && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">System Resources</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <CpuChipIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">CPU Usage</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.system_resources.cpu_usage_percentage.toFixed(1)}%
                </div>
                <div className="w-full bg-gray-200 rounded-full h-1 mt-2">
                  <div
                    className={`h-1 rounded-full ${
                      metrics.system_resources.cpu_usage_percentage > 80 ? 'bg-red-500' :
                      metrics.system_resources.cpu_usage_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                    style={{ width: `${metrics.system_resources.cpu_usage_percentage}%` }}
                  ></div>
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ServerIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Memory Usage</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.system_resources.memory_usage_percentage.toFixed(1)}%
                </div>
                <div className="w-full bg-gray-200 rounded-full h-1 mt-2">
                  <div
                    className={`h-1 rounded-full ${
                      metrics.system_resources.memory_usage_percentage > 80 ? 'bg-red-500' :
                      metrics.system_resources.memory_usage_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                    style={{ width: `${metrics.system_resources.memory_usage_percentage}%` }}
                  ></div>
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <CloudIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Disk Usage</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.system_resources.disk_usage_percentage.toFixed(1)}%
                </div>
                <div className="w-full bg-gray-200 rounded-full h-1 mt-2">
                  <div
                    className={`h-1 rounded-full ${
                      metrics.system_resources.disk_usage_percentage > 80 ? 'bg-red-500' :
                      metrics.system_resources.disk_usage_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                    style={{ width: `${metrics.system_resources.disk_usage_percentage}%` }}
                  ></div>
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <WifiIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Network</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.system_resources.network_bandwidth_mbps.toFixed(1)} Mbps
                </div>
                <div className="text-xs text-gray-500 mt-1">
                  {metrics.system_resources.active_connections} active connections
                </div>
              </div>
            </div>

            {/* Database Connections */}
            <div className="mt-4 bg-blue-50 rounded-lg p-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <ServerIcon className="h-4 w-4 text-blue-600" />
                  <span className="text-sm font-medium text-blue-900">Database Connections</span>
                </div>
                <div className="text-lg font-bold text-blue-600">
                  {metrics.system_resources.database_connections}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Enhanced Email System Monitoring */}
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Enhanced Email Monitoring</h4>
            <button
              onClick={() => setShowEnhancedMetrics(!showEnhancedMetrics)}
              className="text-xs text-blue-600 hover:text-blue-800 focus:outline-none"
            >
              {showEnhancedMetrics ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          {showEnhancedMetrics && (
            <div className="space-y-4">
              {/* Email Queue Monitoring */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Email Queue Monitoring</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Queue Length</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.email_queue_monitoring.queue_length}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Failed Delivery Attempts</div>
                    <div className="text-lg font-bold text-red-900">{metrics.email_queue_monitoring.failed_delivery_attempts}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Processing Rate</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.email_queue_monitoring.queue_processing_rate_per_min}/min</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Storage Usage</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.email_queue_monitoring.storage_usage_gb.toFixed(1)} GB</div>
                    <div className="text-xs text-gray-500">of {metrics.email_queue_monitoring.storage_limit_gb} GB</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Average Queue Wait</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.email_queue_monitoring.average_queue_wait_time_ms}ms</div>
                  </div>
                </div>
              </div>

              {/* Read-Once Enforcement */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Read-Once Enforcement</h5>
                <div className="grid grid-cols-5 gap-3">
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Read-Once Violations</div>
                    <div className="text-lg font-bold text-red-900">{metrics.read_once_enforcement.read_once_violations}</div>
                  </div>
                  <div className="bg-orange-50 rounded-lg p-3">
                    <div className="text-xs text-orange-600">Self-Destruct Triggers</div>
                    <div className="text-lg font-bold text-orange-900">{metrics.read_once_enforcement.self_destruct_triggers}</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Burn-After-Read</div>
                    <div className="text-lg font-bold text-yellow-900">{metrics.read_once_enforcement.burn_after_read_count}</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Retention Policy</div>
                    <div className="text-lg font-bold text-blue-900">{metrics.read_once_enforcement.retention_policy_enforcement}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Email Expiration</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.read_once_enforcement.email_expiration_count}</div>
                  </div>
                </div>
              </div>

              {/* Delivery Analytics */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Delivery Analytics</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-green-50 rounded-lg p-3">
                    <div className="text-xs text-green-600">Delivery Success Rate</div>
                    <div className="text-lg font-bold text-green-900">{metrics.delivery_analytics.delivery_success_rate_percent}%</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Average Delivery Time</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.delivery_analytics.average_delivery_time_ms}ms</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Retry Success Rate</div>
                    <div className="text-lg font-bold text-yellow-900">{metrics.delivery_analytics.retry_success_rate}%</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Dead Letter Queue</div>
                    <div className="text-lg font-bold text-red-900">{metrics.delivery_analytics.dead_letter_queue_size}</div>
                  </div>
                </div>
                
                {/* Failed Delivery Reasons */}
                <div className="mt-3">
                  <h6 className="text-xs font-medium text-gray-600 mb-2">Failed Delivery Reasons</h6>
                  <div className="grid grid-cols-2 gap-2">
                    {Object.entries(metrics.delivery_analytics.failed_delivery_reasons).map(([reason, count]) => (
                      <div key={reason} className="bg-gray-50 rounded p-2">
                        <div className="text-xs text-gray-600 capitalize">{reason.replace('_', ' ')}</div>
                        <div className="text-sm font-bold text-gray-900">{count}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Storage Performance */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Storage Performance</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Read Operations</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.storage_performance.read_operations_per_sec}/sec</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Write Operations</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.storage_performance.write_operations_per_sec}/sec</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Storage Latency</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.storage_performance.storage_latency_ms}ms</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Encryption Overhead</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.storage_performance.encryption_overhead_ms}ms</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Compression Ratio</div>
                    <div className="text-lg font-bold text-gray-900">{(metrics.storage_performance.compression_ratio * 100).toFixed(1)}%</div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Delivery Summary */}
        <div className="bg-blue-50 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <EnvelopeIcon className="h-5 w-5 text-blue-600" />
            <div>
              <h5 className="text-sm font-medium text-blue-900">Email Delivery Summary</h5>
              <p className="text-xs text-blue-700">
                Total sent: {metrics.delivery_performance.total_sent.toLocaleString()} | 
                Success rate: {metrics.delivery_performance.delivery_success_rate.toFixed(1)}% | 
                Queue health: {metrics.queue_status.queue_health_percentage.toFixed(1)}%
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmailDeliveryPanel;
