import React, { useState } from 'react';
import {
  ShieldCheckIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeIcon,
  EyeSlashIcon,

  ClockIcon,
  ServerIcon,
  KeyIcon,

  DocumentTextIcon,
} from '@heroicons/react/24/outline';
import { ZKIDLayerMetrics } from '../../../types/admin';

interface ZKIDLayerPanelProps {
  metrics: ZKIDLayerMetrics | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const ZKIDLayerPanel: React.FC<ZKIDLayerPanelProps> = ({
  metrics,
  isLoading,
  onRefresh,
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [showRecoveryActivity, setShowRecoveryActivity] = useState(false);
  const [showEnhancedMetrics, setShowEnhancedMetrics] = useState(false);

  if (!metrics) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">ZKID Layer</h3>
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
          <p className="mt-2 text-sm text-gray-500">Loading ZKID metrics...</p>
        </div>
      </div>
    );
  }

  const getEndpointStatus = (metric: { success_rate: number }) => {
    if (metric.success_rate >= 99) return 'healthy';
    if (metric.success_rate >= 95) return 'warning';
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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircleIcon className="h-4 w-4" />;
      case 'warning':
        return <ExclamationTriangleIcon className="h-4 w-4" />;
      case 'critical':
        return <ExclamationTriangleIcon className="h-4 w-4" />;
      default:
        return <ClockIcon className="h-4 w-4" />;
    }
  };

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-6 w-6 text-indigo-600" />
            <h3 className="text-lg font-medium text-gray-900">ZKID Layer</h3>
            <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(metrics.enabled ? 'healthy' : 'critical')}`}>
              {metrics.enabled ? 'Active' : 'Inactive'}
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
        {/* Endpoint Health Summary */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Endpoint Health</h4>
          <div className="grid grid-cols-2 gap-3">
            {Object.entries(metrics.endpoint_health).map(([endpoint, metric]) => {
              const status = getEndpointStatus(metric);
              return (
                <div key={endpoint} className="bg-gray-50 rounded-lg p-3">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-medium text-gray-700 capitalize">
                      {endpoint.replace(/_/g, ' ')}
                    </span>
                    <div className={`flex items-center space-x-1 px-2 py-1 rounded-full text-xs ${getStatusColor(status)}`}>
                      {getStatusIcon(status)}
                      <span className="capitalize">{status}</span>
                    </div>
                  </div>
                  <div className="space-y-1">
                    <div className="flex justify-between text-xs">
                      <span className="text-gray-500">Success Rate:</span>
                      <span className="font-medium">{metric.success_rate.toFixed(1)}%</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-gray-500">Latency:</span>
                      <span className="font-medium">{metric.average_latency_ms}ms</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-gray-500">Requests:</span>
                      <span className="font-medium">{metric.requests.toLocaleString()}</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Recovery Operations */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Recovery Operations</h4>
            <button
              onClick={() => setShowRecoveryActivity(!showRecoveryActivity)}
              className="text-xs text-indigo-600 hover:text-indigo-500"
            >
              {showRecoveryActivity ? 'Hide' : 'Show'} Activity
            </button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <KeyIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Generated</span>
              </div>
              <div className="text-2xl font-bold text-blue-600">{metrics.recovery_operations.total_generated}</div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Used</span>
              </div>
              <div className="text-2xl font-bold text-green-600">{metrics.recovery_operations.total_used}</div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Revoked</span>
              </div>
              <div className="text-2xl font-bold text-red-600">{metrics.recovery_operations.total_revoked}</div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">Failed</span>
              </div>
              <div className="text-2xl font-bold text-yellow-600">{metrics.recovery_operations.failed_attempts}</div>
            </div>
          </div>

          {/* Recent Recovery Activity */}
          {showRecoveryActivity && (
            <div className="mt-4 bg-gray-50 rounded-lg p-3">
              <h5 className="text-xs font-medium text-gray-700 mb-2">Recent Activity</h5>
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {metrics.recovery_operations.recent_activity.map((activity) => (
                  <div key={activity.id} className="flex items-center justify-between text-xs">
                    <div className="flex items-center space-x-2">
                      <div
                        className={`w-2 h-2 rounded-full ${
                          activity.success ? 'bg-green-400' : 'bg-red-400'
                        }`}
                      ></div>
                      <span className="text-gray-600 capitalize">{activity.action}</span>
                      <span className="text-gray-400">by {activity.user_id.slice(0, 8)}...</span>
                    </div>
                    <span className="text-gray-400">
                      {new Date(activity.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Database Performance */}
        {showDetails && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Database Performance</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ServerIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Mapping QPS</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.database_performance.mapping_queries_per_second.toFixed(1)}
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ServerIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Recovery QPS</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.database_performance.recovery_queries_per_second.toFixed(1)}
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ClockIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Avg Query Time</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.database_performance.average_query_time_ms}ms
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <KeyIcon className="h-4 w-4 text-gray-600" />
                  <span className="text-xs font-medium text-gray-700">Encryption Overhead</span>
                </div>
                <div className="text-lg font-bold text-gray-900">
                  {metrics.database_performance.encryption_overhead_ms}ms
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Security Events */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-3">Security Events</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Unauthorized Access</span>
              </div>
              <div className="text-lg font-bold text-red-600">
                {metrics.security_events.unauthorized_access_attempts}
              </div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">Failed Recovery</span>
              </div>
              <div className="text-lg font-bold text-yellow-600">
                {metrics.security_events.failed_recovery_attempts}
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Encryption Errors</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.security_events.encryption_errors}
              </div>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <DocumentTextIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Audit Logs</span>
              </div>
              <div className="text-lg font-bold text-blue-600">
                {metrics.security_events.audit_log_entries.toLocaleString()}
              </div>
            </div>
          </div>
        </div>

        {/* Enhanced ZKID Monitoring */}
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Enhanced Monitoring</h4>
            <button
              onClick={() => setShowEnhancedMetrics(!showEnhancedMetrics)}
              className="text-xs text-indigo-600 hover:text-indigo-800 focus:outline-none"
            >
              {showEnhancedMetrics ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          {showEnhancedMetrics && (
            <div className="space-y-4">
              {/* UUID Mapping Operations */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">UUID Mapping Operations</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Mapping Creations</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.uuid_mapping_creations.toLocaleString()}</div>
                    <div className="text-xs text-gray-500">{metrics.mapping_creation_latency_ms}ms avg</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Mapping Retrievals</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.uuid_mapping_retrievals.toLocaleString()}</div>
                    <div className="text-xs text-gray-500">{metrics.mapping_retrieval_latency_ms}ms avg</div>
                  </div>
                </div>
              </div>

              {/* Recovery Code Status */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Recovery Code Status</h5>
                <div className="grid grid-cols-4 gap-3">
                  <div className="bg-green-50 rounded-lg p-3">
                    <div className="text-xs text-green-600">Active</div>
                    <div className="text-lg font-bold text-green-900">{metrics.active_recovery_codes}</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Expired</div>
                    <div className="text-lg font-bold text-yellow-900">{metrics.expired_recovery_codes}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Revoked</div>
                    <div className="text-lg font-bold text-red-900">{metrics.revoked_recovery_codes}</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Used</div>
                    <div className="text-lg font-bold text-blue-900">{metrics.recovery_code_usage_count}</div>
                  </div>
                </div>
              </div>

              {/* Security Status */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Security Status</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className={`rounded-lg p-3 ${metrics.side_channel_protection_status === 'active' ? 'bg-green-50' : 'bg-red-50'}`}>
                    <div className="text-xs text-gray-600">Side-Channel Protection</div>
                    <div className={`text-sm font-medium ${metrics.side_channel_protection_status === 'active' ? 'text-green-900' : 'text-red-900'}`}>
                      {metrics.side_channel_protection_status === 'active' ? 'Active' : 'Inactive'}
                    </div>
                  </div>
                  <div className={`rounded-lg p-3 ${metrics.rate_limiting_status === 'enforced' ? 'bg-green-50' : 'bg-red-50'}`}>
                    <div className="text-xs text-gray-600">Rate Limiting</div>
                    <div className={`text-sm font-medium ${metrics.rate_limiting_status === 'enforced' ? 'text-green-900' : 'text-red-900'}`}>
                      {metrics.rate_limiting_status === 'enforced' ? 'Enforced' : 'Bypassed'}
                    </div>
                  </div>
                </div>
              </div>

              {/* Performance Metrics */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Performance Metrics</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Failed UUID Lookups</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.failed_uuid_lookups}</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Recovery Generation Latency</div>
                    <div className="text-lg font-bold text-gray-900">{metrics.recovery_code_generation_latency_ms}ms</div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Zero-Knowledge Guarantee */}
        <div className="mt-6 p-3 bg-indigo-50 rounded-lg border border-indigo-200">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-5 w-5 text-indigo-600" />
            <div>
              <h5 className="text-sm font-medium text-indigo-900">Zero-Knowledge Privacy</h5>
              <p className="text-xs text-indigo-700">
                All operations use UUID-only identifiers. No external emails are visible to administrators.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ZKIDLayerPanel;
