import React, { useState } from 'react';
import {
  DocumentTextIcon,
  ArrowPathIcon,
  MagnifyingGlassIcon,
  FunnelIcon,
  EyeIcon,
  EyeSlashIcon,
  ArrowDownTrayIcon,
  ClockIcon,
  UserIcon,
  GlobeAltIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';
import { AuditLogEntry } from '../../../types/admin';

interface AuditLogsPanelProps {
  logs: AuditLogEntry[];
  isLoading: boolean;
  onRefresh: () => void;
}

const AuditLogsPanel: React.FC<AuditLogsPanelProps> = ({
  logs,
  isLoading,
  onRefresh,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedAction, setSelectedAction] = useState<string>('all');
  const [selectedSuccess, setSelectedSuccess] = useState<string>('all');
  const [showDetails, setShowDetails] = useState(false);

  // Get unique actions and success states for filtering
  const actions = Array.from(new Set(logs.map(log => log.action)));


  // Filter logs based on search and filters
  const filteredLogs = logs.filter(log => {
    const matchesSearch = searchTerm === '' || 
      log.action.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.resource.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.user_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      log.ip_address.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesAction = selectedAction === 'all' || log.action === selectedAction;
    const matchesSuccess = selectedSuccess === 'all' || 
      (selectedSuccess === 'true' && log.success) ||
      (selectedSuccess === 'false' && !log.success);

    return matchesSearch && matchesAction && matchesSuccess;
  });

  const getActionColor = (action: string) => {
    switch (action) {
      case 'LOGIN':
        return 'bg-blue-50 text-blue-700';
      case 'LOGOUT':
        return 'bg-gray-50 text-gray-700';
      case 'EMAIL_SENT':
        return 'bg-green-50 text-green-700';
      case 'EMAIL_RETRIEVED':
        return 'bg-purple-50 text-purple-700';
      case 'RECOVERY_CODE_GENERATED':
        return 'bg-yellow-50 text-yellow-700';
      case 'RECOVERY_CODE_USED':
        return 'bg-orange-50 text-orange-700';
      case 'ADMIN_ACCESS':
        return 'bg-red-50 text-red-700';
      default:
        return 'bg-gray-50 text-gray-700';
    }
  };

  const getSuccessIcon = (success: boolean) => {
    return success ? 
      <CheckCircleIcon className="h-4 w-4 text-green-600" /> : 
      <XCircleIcon className="h-4 w-4 text-red-600" />;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const truncateUUID = (uuid: string) => {
    return uuid.length > 8 ? `${uuid.slice(0, 8)}...` : uuid;
  };

  const exportLogs = () => {
    const csvContent = [
      'Timestamp,User ID,Action,Resource,IP Address,Success,User Agent',
      ...filteredLogs.map(log => 
        `"${log.timestamp}","${log.user_id}","${log.action}","${log.resource}","${log.ip_address}","${log.success}","${log.user_agent}"`
      )
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <DocumentTextIcon className="h-6 w-6 text-indigo-600" />
            <h3 className="text-lg font-medium text-gray-900">Audit Logs</h3>
            <div className="px-2 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
              {logs.length} Entries
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={exportLogs}
              className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded-md"
              title="Export to CSV"
            >
              <ArrowDownTrayIcon className="h-4 w-4" />
            </button>
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
        {/* Search and Filters */}
        <div className="mb-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Search */}
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search logs..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm"
              />
            </div>

            {/* Action Filter */}
            <div className="relative">
              <FunnelIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <select
                value={selectedAction}
                onChange={(e) => setSelectedAction(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm appearance-none bg-white"
              >
                <option value="all">All Actions</option>
                {actions.map(action => (
                  <option key={action} value={action}>{action}</option>
                ))}
              </select>
            </div>

            {/* Success Filter */}
            <div className="relative">
              <FunnelIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <select
                value={selectedSuccess}
                onChange={(e) => setSelectedSuccess(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm appearance-none bg-white"
              >
                <option value="all">All Results</option>
                <option value="true">Successful</option>
                <option value="false">Failed</option>
              </select>
            </div>

            {/* Results Count */}
            <div className="flex items-center justify-end text-sm text-gray-500">
              {filteredLogs.length} of {logs.length} entries
            </div>
          </div>
        </div>

        {/* Logs Table */}
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Timestamp
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  User ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Action
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Resource
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  IP Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                {showDetails && (
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Details
                  </th>
                )}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredLogs.length === 0 ? (
                <tr>
                  <td colSpan={showDetails ? 7 : 6} className="px-6 py-4 text-center text-sm text-gray-500">
                    No audit logs found matching the current filters
                  </td>
                </tr>
              ) : (
                filteredLogs.map((log) => (
                  <tr key={log.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center space-x-2">
                        <ClockIcon className="h-4 w-4 text-gray-400" />
                        <span>{formatTimestamp(log.timestamp)}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center space-x-2">
                        <UserIcon className="h-4 w-4 text-gray-400" />
                        <span className="font-mono">{truncateUUID(log.user_id)}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getActionColor(log.action)}`}>
                        {log.action}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                        {log.resource}
                      </code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center space-x-2">
                        <GlobeAltIcon className="h-4 w-4 text-gray-400" />
                        <span>{log.ip_address}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center space-x-2">
                        {getSuccessIcon(log.success)}
                        <span className={`text-sm font-medium ${log.success ? 'text-green-600' : 'text-red-600'}`}>
                          {log.success ? 'Success' : 'Failed'}
                        </span>
                      </div>
                    </td>
                    {showDetails && (
                      <td className="px-6 py-4 text-sm text-gray-900">
                        <div className="max-w-xs">
                          <details className="text-xs">
                            <summary className="cursor-pointer text-indigo-600 hover:text-indigo-800">
                              View Details
                            </summary>
                            <div className="mt-2 p-2 bg-gray-50 rounded text-xs">
                              <div className="space-y-1">
                                <div><strong>User Agent:</strong> {log.user_agent}</div>
                                {log.details && Object.entries(log.details).map(([key, value]) => (
                                  <div key={key}>
                                    <strong>{key}:</strong> {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                                  </div>
                                ))}
                              </div>
                            </div>
                          </details>
                        </div>
                      </td>
                    )}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Zero-Knowledge Notice */}
        <div className="mt-6 p-4 bg-indigo-50 rounded-lg border border-indigo-200">
          <div className="flex items-center space-x-2">
            <DocumentTextIcon className="h-5 w-5 text-indigo-600" />
            <div>
              <h5 className="text-sm font-medium text-indigo-900">Zero-Knowledge Audit Logging</h5>
              <p className="text-xs text-indigo-700">
                All audit logs use UUID-only identifiers. No external emails or personal data are visible to administrators.
                All operations are logged for compliance and security monitoring while maintaining privacy.
              </p>
            </div>
          </div>
        </div>

        {/* Log Summary */}
        <div className="mt-4 grid grid-cols-4 gap-4 text-center">
          <div className="bg-gray-50 rounded-lg p-3">
            <div className="text-lg font-bold text-gray-900">{logs.length}</div>
            <div className="text-xs text-gray-500">Total Entries</div>
          </div>
          <div className="bg-green-50 rounded-lg p-3">
            <div className="text-lg font-bold text-green-600">
              {logs.filter(log => log.success).length}
            </div>
            <div className="text-xs text-green-500">Successful</div>
          </div>
          <div className="bg-red-50 rounded-lg p-3">
            <div className="text-lg font-bold text-red-600">
              {logs.filter(log => !log.success).length}
            </div>
            <div className="text-xs text-red-500">Failed</div>
          </div>
          <div className="bg-blue-50 rounded-lg p-3">
            <div className="text-lg font-bold text-blue-600">
              {Array.from(new Set(logs.map(log => log.user_id))).length}
            </div>
            <div className="text-xs text-blue-500">Unique Users</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuditLogsPanel;
