import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { fetchAuditLogs, AuditLog, AuditLogResponse } from '../../lib/adminApi';

interface AuditLogTableProps {
  className?: string;
}

const AuditLogTable: React.FC<AuditLogTableProps> = ({ className = '' }) => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  
  // Separate state for each filter to avoid object recreation
  const [userFilter, setUserFilter] = useState('');
  const [actionFilter, setActionFilter] = useState('');
  const [entityFilter, setEntityFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [pageSize] = useState(20);

  // Memoize filters to prevent unnecessary re-renders - only change when actual values change
  const memoizedFilters = useMemo(() => ({
    user_id: userFilter || undefined,
    action: actionFilter || undefined,
    entity: entityFilter || undefined,
    severity: severityFilter || undefined,
    page: currentPage,
    limit: pageSize,
  }), [userFilter, actionFilter, entityFilter, severityFilter, currentPage, pageSize]);

  // Memoize the load function to prevent recreation
  const loadLogs = useCallback(async () => {
    let isMounted = true;
    setLoading(true);
    setError(null);

    try {
      const response: AuditLogResponse = await fetchAuditLogs(memoizedFilters);
      if (isMounted) {
        setLogs(response.logs || []); // Replace with fresh logs, not overwrite mid-loop
        setTotal(response.total);
        setError(null);
      }
    } catch (err: unknown) {
      if (isMounted) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to load audit logs';
        setError(errorMessage);
        setLogs([]);
        setTotal(0);
      }
    } finally {
      if (isMounted) {
        setLoading(false);
      }
    }
  }, [memoizedFilters]);

  // Load logs whenever filters change - this will only fire when memoizedFilters actually changes
  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  const handleFilterChange = (filterType: string, value: string) => {
    // Reset to first page when filtering
    setCurrentPage(1);
    
    switch (filterType) {
      case 'user_id':
        setUserFilter(value);
        break;
      case 'action':
        setActionFilter(value);
        break;
      case 'entity':
        setEntityFilter(value);
        break;
      case 'severity':
        setSeverityFilter(value);
        break;
    }
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'bg-red-100 text-red-800';
      case 'high':
        return 'bg-orange-100 text-orange-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const totalPages = Math.ceil(total / pageSize);

  if (loading) {
    return (
      <div className={`flex justify-center items-center py-8 ${className}`}>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <span className="ml-2 text-gray-600">Loading audit logs...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`text-center py-8 ${className}`} role="alert">
        <div className="text-red-600">Error: {error}</div>
        <button
          onClick={loadLogs}
          className="mt-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className={className}>
      {/* Filters */}
      <div className="mb-6 grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            User ID
          </label>
          <input
            type="text"
            placeholder="Filter by user ID..."
            value={userFilter}
            onChange={(e) => handleFilterChange('user_id', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Action
          </label>
          <input
            type="text"
            placeholder="Filter by action..."
            value={actionFilter}
            onChange={(e) => handleFilterChange('action', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Entity
          </label>
          <input
            type="text"
            placeholder="Filter by entity..."
            value={entityFilter}
            onChange={(e) => handleFilterChange('entity', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Severity
          </label>
          <select
            value={severityFilter}
            onChange={(e) => handleFilterChange('severity', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Severities</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </div>
      </div>

      {/* Results count */}
      <div className="mb-4 text-sm text-gray-600">
        Showing {logs.length} of {total} audit logs
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Timestamp
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                User
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Action
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Entity
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Details
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Severity
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {logs.map((log) => (
              <tr key={log.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {formatTimestamp(log.timestamp)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {log.user_id || 'System'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {log.action}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                  {log.entity}
                </td>
                <td className="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
                  {log.details}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getSeverityColor(log.severity)}`}>
                    {log.severity}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-6 flex items-center justify-between">
          <div className="text-sm text-gray-700">
            Page {currentPage} of {totalPages}
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage <= 1}
              className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
              className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}

      {logs.length === 0 && !loading && (
        <div className="text-center py-8 text-gray-500">
          No audit logs found matching the current filters.
        </div>
      )}
    </div>
  );
};

export default AuditLogTable;
