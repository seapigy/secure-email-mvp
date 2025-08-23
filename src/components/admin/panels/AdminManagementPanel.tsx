import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  UserGroupIcon, KeyIcon, ClockIcon, CheckCircleIcon, PlusIcon, TrashIcon,
  EyeIcon, EyeSlashIcon, ShieldCheckIcon, ExclamationTriangleIcon, UserIcon, ArrowPathIcon,
} from '@heroicons/react/24/outline';
import {
  AdminUser, InvitationKey, AdminActionApproval, AdminInvitationRequest,
} from '../../../types/admin';
import { EnterpriseDashboardService } from '../../../services/enterpriseDashboardService';

interface AdminManagementPanelProps {
  isLoading: boolean;
  onRefresh: () => void;
}

const AdminManagementPanel: React.FC<AdminManagementPanelProps> = ({
  isLoading,
  onRefresh,
}) => {
  const [admins, setAdmins] = useState<AdminUser[]>([]);
  const [invitations, setInvitations] = useState<InvitationKey[]>([]);
  const [pendingApprovals, setPendingApprovals] = useState<AdminActionApproval[]>([]);
  const [showInviteForm, setShowInviteForm] = useState(false);
  const [showDetails, setShowDetails] = useState(false);
  const [currentUser, setCurrentUser] = useState<AdminUser | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const dashboardService = useMemo(() => new EnterpriseDashboardService(), []);

  const loadData = useCallback(async () => {
    try {
      const [adminsData, invitationsData, approvalsData] = await Promise.all([
        dashboardService.listAdmins(),
        dashboardService.listInvitationKeys(),
        dashboardService.getPendingApprovals(),
      ]);
      setAdmins(adminsData);
      setInvitations(invitationsData);
      setPendingApprovals(approvalsData);
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to load admin data');
    }
  }, [dashboardService]);

  useEffect(() => {
    loadData();
    const user = dashboardService.getCurrentUser();
    setCurrentUser(user);
  }, [dashboardService, loadData]);

  const handleCreateInvitation = async (request: AdminInvitationRequest) => {
    try {
      setError(null);
      await dashboardService.createInvitationKey(request);
      await loadData();
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to create invitation');
    }
  };

  const handleRevokeInvitation = async (key: string) => {
    try {
      setError(null);
      await dashboardService.revokeInvitationKey(key);
      await loadData();
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to revoke invitation');
    }
  };

  const handleDeactivateAdmin = async (adminId: string) => {
    try {
      await dashboardService.deactivateAdmin(adminId);
      setSuccess('Admin deactivated successfully');
      loadData();
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to deactivate admin');
    }
  };

  const handleApproveAction = async (approvalId: string, approved: boolean, reason?: string) => {
    try {
      await dashboardService.approveAction(approvalId, approved, reason);
      setSuccess(`Action ${approved ? 'approved' : 'rejected'} successfully`);
      loadData();
    } catch (error: unknown) {
      const err = error as Error;
      setError(err.message || 'Failed to process approval');
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'primary_admin':
        return 'bg-purple-100 text-purple-800';
      case 'full_admin':
        return 'bg-blue-100 text-blue-800';
      case 'read_only_admin':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'primary_admin':
        return <ShieldCheckIcon className="h-4 w-4" />;
      case 'full_admin':
        return <UserGroupIcon className="h-4 w-4" />;
      case 'read_only_admin':
        return <UserIcon className="h-4 w-4" />;
      default:
        return <UserIcon className="h-4 w-4" />;
    }
  };

  const canManageAdmins = currentUser && dashboardService.canManageAdmins();
  const isPrimaryAdmin = currentUser && dashboardService.isPrimaryAdmin();

  if (!canManageAdmins) {
    return (
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <UserGroupIcon className="h-6 w-6 text-gray-600" />
              <h3 className="text-lg font-medium text-gray-900">Admin Management</h3>
            </div>
          </div>
        </div>
        <div className="p-6">
          <div className="text-center py-8">
            <ExclamationTriangleIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">Access Restricted</h3>
            <p className="mt-1 text-sm text-gray-500">
              You don&apos;t have permission to manage admin accounts.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <UserGroupIcon className="h-6 w-6 text-indigo-600" />
            <h3 className="text-lg font-medium text-gray-900">Admin Management</h3>
            <div className="px-2 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
              {admins.length} Admins
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button onClick={() => setShowDetails(!showDetails)} className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded">
              {showDetails ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
            </button>
            <button onClick={onRefresh} disabled={isLoading} className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded disabled:opacity-50">
              <ArrowPathIcon className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* Error/Success Messages */}
      {error && (
        <div className="px-6 py-3 bg-red-50 border-b border-red-200">
          <div className="flex items-center space-x-2">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
            <span className="text-sm text-red-700">{error}</span>
          </div>
        </div>
      )}

      {success && (
        <div className="px-6 py-3 bg-green-50 border-b border-green-200">
          <div className="flex items-center space-x-2">
            <CheckCircleIcon className="h-5 w-5 text-green-400" />
            <span className="text-sm text-green-700">{success}</span>
          </div>
        </div>
      )}

      {/* Content */}
      <div className="p-6">
        {/* Admin Summary */}
        <div className="mb-6">
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-blue-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <ShieldCheckIcon className="h-5 w-5 text-blue-600" />
                <div>
                  <p className="text-sm font-medium text-blue-900">Primary Admin</p>
                  <p className="text-lg font-bold text-blue-600">
                    {admins.filter(a => a.role === 'primary_admin').length}
                  </p>
                </div>
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <UserGroupIcon className="h-5 w-5 text-green-600" />
                <div>
                  <p className="text-sm font-medium text-green-900">Full Admins</p>
                  <p className="text-lg font-bold text-green-600">
                    {admins.filter(a => a.role === 'full_admin').length}
                  </p>
                </div>
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <UserIcon className="h-5 w-5 text-gray-600" />
                <div>
                  <p className="text-sm font-medium text-gray-900">Read-Only</p>
                  <p className="text-lg font-bold text-gray-600">
                    {admins.filter(a => a.role === 'read_only_admin').length}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Active Admins */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Active Admins</h4>
            {isPrimaryAdmin && (
              <button
                onClick={() => setShowInviteForm(true)}
                className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                <PlusIcon className="h-3 w-3 mr-1" />
                Invite Admin
              </button>
            )}
          </div>
          <div className="space-y-3">
            {admins.map((admin) => (
              <div key={admin.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center space-x-3">
                  <div className={`p-2 rounded-full ${getRoleColor(admin.role)}`}>
                    {getRoleIcon(admin.role)}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">{admin.username}</p>
                    <p className="text-xs text-gray-500">
                      {admin.role.replace('_', ' ')} • Last login: {admin.last_login ? new Date(admin.last_login).toLocaleDateString() : 'Never'}
                    </p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  {admin.mfa_enabled && (
                    <div className="px-2 py-1 rounded-full text-xs bg-green-100 text-green-800">
                      MFA Enabled
                    </div>
                  )}
                  {admin.id === currentUser?.id && (
                    <div className="px-2 py-1 rounded-full text-xs bg-blue-100 text-blue-800">
                      You
                    </div>
                  )}
                  {isPrimaryAdmin && admin.id !== currentUser?.id && (
                    <button
                      onClick={() => handleDeactivateAdmin(admin.id)}
                      className="p-1 text-red-400 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 rounded"
                      title="Deactivate Admin"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Pending Invitations */}
        {invitations.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Pending Invitations</h4>
            <div className="space-y-3">
              {invitations.filter(inv => !inv.used).map((invitation) => (
                <div key={invitation.key} className="flex items-center justify-between p-3 bg-yellow-50 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <KeyIcon className="h-5 w-5 text-yellow-600" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {invitation.scope.replace('_', ' ')} Access
                      </p>
                      <p className="text-xs text-gray-500">
                        Created by {invitation.created_by} • Expires: {new Date(invitation.expires_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className="text-xs text-gray-500">
                      {invitation.current_uses}/{invitation.max_uses || 1} uses
                    </span>
                    <button
                      onClick={() => handleRevokeInvitation(invitation.key)}
                      className="p-1 text-red-400 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 rounded"
                      title="Revoke Invitation"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Pending Approvals */}
        {pendingApprovals.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Pending Approvals</h4>
            <div className="space-y-3">
              {pendingApprovals.map((approval) => (
                <div key={approval.id} className="p-3 bg-orange-50 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <ClockIcon className="h-4 w-4 text-orange-600" />
                      <span className="text-sm font-medium text-gray-900">
                        {approval.action_type.replace('_', ' ')}
                      </span>
                    </div>
                    <span className="text-xs text-gray-500">
                      {new Date(approval.requested_at).toLocaleDateString()}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 mb-3">
                    Requested by: {approval.requested_by}
                  </p>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => handleApproveAction(approval.id, true)}
                      className="px-3 py-1 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded focus:outline-none focus:ring-2 focus:ring-green-500"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => handleApproveAction(approval.id, false)}
                      className="px-3 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded focus:outline-none focus:ring-2 focus:ring-red-500"
                    >
                      Reject
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Admin Permissions Summary */}
        {showDetails && (
          <div className="mt-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Permission Summary</h4>
            <div className="grid grid-cols-2 gap-4 text-xs">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">System Management:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_manage_system).length} admins
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Admin Management:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_manage_admins).length} admins
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Sensitive Data Access:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_view_sensitive_data).length} admins
                  </span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Data Export:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_export_data).length} admins
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Settings Modification:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_modify_settings).length} admins
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Audit Log Access:</span>
                  <span className="font-medium">
                    {admins.filter(a => a.permissions.can_access_audit_logs).length} admins
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Invite Admin Modal */}
      {showInviteForm && (
        <InviteAdminModal
          onClose={() => setShowInviteForm(false)}
          onSubmit={handleCreateInvitation}
        />
      )}
    </div>
  );
};

// Invite Admin Modal Component
interface InviteAdminModalProps {
  onClose: () => void;
  onSubmit: (data: AdminInvitationRequest) => void;
}

const InviteAdminModal: React.FC<InviteAdminModalProps> = ({ onClose, onSubmit }) => {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<'full_admin' | 'read_only_admin'>('read_only_admin');
  const [expiresInHours, setExpiresInHours] = useState(24);
  const [maxUses, setMaxUses] = useState(1);
  const [message, setMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      await onSubmit({
        email,
        role,
        expires_in_hours: expiresInHours,
        max_uses: maxUses,
        message: message || undefined,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="mt-3">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Invite New Admin</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Email Address</label>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="admin@example.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Role</label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value as 'full_admin' | 'read_only_admin')}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              >
                <option value="read_only_admin">Read-Only Admin</option>
                <option value="full_admin">Full Admin</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Expires In (hours)</label>
              <input
                type="number"
                min="1"
                max="168"
                value={expiresInHours}
                onChange={(e) => setExpiresInHours(parseInt(e.target.value))}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Max Uses</label>
              <input
                type="number"
                min="1"
                max="10"
                value={maxUses}
                onChange={(e) => setMaxUses(parseInt(e.target.value))}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Message (Optional)</label>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                rows={3}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="Optional message for the invitee..."
              />
            </div>

            <div className="flex items-center justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
              >
                {isSubmitting ? 'Sending...' : 'Send Invitation'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default AdminManagementPanel;
