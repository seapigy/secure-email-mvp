import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import axios from 'axios';
import AuditLogTable from '../../components/admin/AuditLogTable';
import { log } from '@/lib/logger';

interface SecurityPolicy {
  id: string;
  name: string;
  type: string;
  category: string;
  enabled: boolean;
  severity: string;
  enforcement_level: string;
  created_at: string;
  updated_at: string;
}

interface DLPLog {
  id: string;
  timestamp: string;
  email: string;
  scan_type: string;
  findings: string[];
  action: string;
  admin: string;
}

interface AdminUser {
  id: number;
  email: string;
  role: string;
  created_at: string;
}

const AdminDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'policies' | 'dlp' | 'users' | 'audit'>('policies');
  const [isLoading, setIsLoading] = useState(false);
  const [securityPolicies, setSecurityPolicies] = useState<SecurityPolicy[]>([]);
  const [dlpLogs, setDlpLogs] = useState<DLPLog[]>([]);
  // TODO: Implement admin users management
  // const [adminUsers, setAdminUsers] = useState<AdminUser[]>([]);
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);

  useEffect(() => {
    // Check if admin is authenticated
    const token = localStorage.getItem('admin_token');
    const userStr = localStorage.getItem('admin_user');
    
    if (!token || !userStr) {
      toast.error('Please log in to access the admin dashboard', {
        position: 'top-right',
        autoClose: 3000,
      });
      navigate('/admin/login');
      return;
    }

    try {
      const user = JSON.parse(userStr);
      setAdminUser(user);
    } catch (error) {
      log.error('Failed to parse admin user:', error, 'AdminDashboard');
      navigate('/admin/login');
      return;
    }

    // Load initial data
    loadSecurityPolicies();
  }, [navigate]);

  const loadSecurityPolicies = async () => {
    setIsLoading(true);
    try {
      const token = localStorage.getItem('admin_token');
      const response = await axios.get('/api/security/policies', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.data.success) {
        setSecurityPolicies(response.data.policies || []);
      }
    } catch (error: unknown) {
      log.error('Failed to load security policies:', error, 'AdminDashboard');
      toast.error('Failed to load security policies', {
        position: 'top-right',
        autoClose: 5000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const loadDLPLogs = async () => {
    setIsLoading(true);
    try {
      const token = localStorage.getItem('admin_token');
      const response = await axios.get('/api/admin/dlp/logs', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.data.success) {
        setDlpLogs(response.data.logs || []);
      }
    } catch (error: unknown) {
      log.error('Failed to load DLP logs:', error, 'AdminDashboard');
      toast.error('Failed to load DLP logs', {
        position: 'top-right',
        autoClose: 5000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');
    toast.success('Logged out successfully', {
      position: 'top-right',
      autoClose: 3000,
    });
    navigate('/admin/login');
  };

  const handleTabChange = (tab: 'policies' | 'dlp' | 'users' | 'audit') => {
    setActiveTab(tab);
    if (tab === 'dlp') {
      loadDLPLogs();
    }
  };

  const updateSecurityPolicy = async (policyId: string, updates: Partial<SecurityPolicy>) => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await axios.put(`/api/security/policies/${policyId}`, updates, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (response.data.success) {
        toast.success('Security policy updated successfully', {
          position: 'top-right',
          autoClose: 3000,
        });
        loadSecurityPolicies(); // Reload policies
      }
    } catch (error: unknown) {
      log.error('Failed to update security policy:', error, 'AdminDashboard');
      toast.error('Failed to update security policy', {
        position: 'top-right',
        autoClose: 5000,
      });
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
              <p className="text-sm text-gray-600">
                Welcome, {adminUser?.email} ({adminUser?.role})
              </p>
            </div>
            <button
              onClick={handleLogout}
              className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium"
            >
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => handleTabChange('policies')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'policies'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Security Policies
            </button>
            <button
              onClick={() => handleTabChange('dlp')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'dlp'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              DLP Logs
            </button>
            <button
              onClick={() => handleTabChange('users')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'users'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              User Management
            </button>
            <button
              onClick={() => handleTabChange('audit')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'audit'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Audit Logs
            </button>
          </nav>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {isLoading && (
          <div className="flex justify-center items-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span className="ml-2 text-gray-600">Loading...</span>
          </div>
        )}

        {/* Security Policies Tab */}
        {activeTab === 'policies' && !isLoading && (
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">Security Policies</h2>
              <p className="text-sm text-gray-600">Manage system-wide security policies</p>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Policy
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Severity
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {securityPolicies.map((policy) => (
                    <tr key={policy.id}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">{policy.name}</div>
                        <div className="text-sm text-gray-500">{policy.category}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {policy.type}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            policy.enabled
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {policy.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {policy.severity}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <button
                          onClick={() => updateSecurityPolicy(policy.id, { enabled: !policy.enabled })}
                          className="text-blue-600 hover:text-blue-900"
                        >
                          {policy.enabled ? 'Disable' : 'Enable'}
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* DLP Logs Tab */}
        {activeTab === 'dlp' && !isLoading && (
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">DLP Logs</h2>
              <p className="text-sm text-gray-600">Data Loss Prevention scan results and actions</p>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Timestamp
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Email
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Scan Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Findings
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {dlpLogs.map((log) => (
                    <tr key={log.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {new Date(log.timestamp).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {log.email}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {log.scan_type}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {log.findings.length > 0 ? (
                          <div className="flex flex-wrap gap-1">
                            {log.findings.map((finding, index) => (
                              <span
                                key={index}
                                className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800"
                              >
                                {finding}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <span className="text-sm text-gray-500">None</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            log.action === 'blocked'
                              ? 'bg-red-100 text-red-800'
                              : 'bg-green-100 text-green-800'
                          }`}
                        >
                          {log.action}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* User Management Tab */}
        {activeTab === 'users' && !isLoading && (
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">User Management</h2>
              <p className="text-sm text-gray-600">Manage system users and permissions</p>
            </div>
            <div className="p-6 text-center text-gray-500">
              <p>User management functionality coming soon...</p>
              <p className="text-sm mt-2">This will include user creation, role management, and permission controls.</p>
            </div>
          </div>
        )}

        {/* Audit Logs Tab */}
        {activeTab === 'audit' && !isLoading && (
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">Audit Logs</h2>
              <p className="text-sm text-gray-600">System audit trail and activity monitoring</p>
            </div>
            <div className="p-6">
              <AuditLogTable />
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminDashboard;
