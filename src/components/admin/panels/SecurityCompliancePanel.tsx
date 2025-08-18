import React, { useState } from 'react';
import {
  ShieldCheckIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeIcon,
  EyeSlashIcon,
  LockClosedIcon,
  UserGroupIcon,
  DocumentTextIcon,
  FlagIcon,
  GlobeAltIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';
import { SecurityComplianceMetrics } from '../../../types/admin';

interface SecurityCompliancePanelProps {
  metrics: SecurityComplianceMetrics | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const SecurityCompliancePanel: React.FC<SecurityCompliancePanelProps> = ({
  metrics,
  isLoading,
  onRefresh,
}) => {
  const [showDetails, setShowDetails] = useState(false);

  if (!metrics) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Security & Compliance</h3>
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
          <p className="mt-2 text-sm text-gray-500">Loading security metrics...</p>
        </div>
      </div>
    );
  }

  const getSecurityStatus = (violations: number) => {
    if (violations === 0) return 'secure';
    if (violations <= 5) return 'warning';
    return 'critical';
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'secure':
        return 'text-green-600 bg-green-100';
      case 'warning':
        return 'text-yellow-600 bg-yellow-100';
      case 'critical':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };



  const getFeatureFlagStatus = (enabled: boolean) => {
    return enabled ? 'enabled' : 'disabled';
  };

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-6 w-6 text-green-600" />
            <h3 className="text-lg font-medium text-gray-900">Security & Compliance</h3>
            <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(getSecurityStatus(metrics.access_control.rbac_violations))}`}>
              {getSecurityStatus(metrics.access_control.rbac_violations).toUpperCase()}
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
        {/* Authentication Security */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Authentication Security</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Failed Logins</span>
              </div>
              <div className="text-2xl font-bold text-red-600">
                {metrics.authentication_security.failed_login_attempts}
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Successful Logins</span>
              </div>
              <div className="text-2xl font-bold text-green-600">
                {metrics.authentication_security.successful_logins}
              </div>
            </div>
            <div className="bg-orange-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-orange-600" />
                <span className="text-xs font-medium text-orange-900">Brute Force</span>
              </div>
              <div className="text-2xl font-bold text-orange-600">
                {metrics.authentication_security.brute_force_attempts}
              </div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <LockClosedIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">Account Lockouts</span>
              </div>
              <div className="text-2xl font-bold text-yellow-600">
                {metrics.authentication_security.account_lockouts}
              </div>
            </div>
          </div>
        </div>

        {/* Access Control */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Access Control</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className={`rounded-lg p-3 ${metrics.access_control.rbac_violations === 0 ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <UserGroupIcon className={`h-4 w-4 ${metrics.access_control.rbac_violations === 0 ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.access_control.rbac_violations === 0 ? 'text-green-900' : 'text-red-900'}`}>
                  RBAC Violations
                </span>
              </div>
              <div className={`text-2xl font-bold ${metrics.access_control.rbac_violations === 0 ? 'text-green-600' : 'text-red-600'}`}>
                {metrics.access_control.rbac_violations}
              </div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Unauthorized Access</span>
              </div>
              <div className="text-2xl font-bold text-red-600">
                {metrics.access_control.unauthorized_access_attempts}
              </div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Privilege Escalation</span>
              </div>
              <div className="text-2xl font-bold text-red-600">
                {metrics.access_control.privilege_escalation_attempts}
              </div>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Session Timeouts</span>
              </div>
              <div className="text-2xl font-bold text-blue-600">
                {metrics.access_control.session_timeouts}
              </div>
            </div>
          </div>
        </div>

        {/* Audit Compliance */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Audit Compliance</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <DocumentTextIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Audit Logs</span>
              </div>
              <div className="text-lg font-bold text-blue-600">
                {metrics.audit_compliance.audit_log_entries.toLocaleString()}
              </div>
            </div>
            <div className="bg-indigo-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <DocumentTextIcon className="h-4 w-4 text-indigo-600" />
                <span className="text-xs font-medium text-indigo-900">Compliance Events</span>
              </div>
              <div className="text-lg font-bold text-indigo-600">
                {metrics.audit_compliance.compliance_events}
              </div>
            </div>
            <div className="bg-purple-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <DocumentTextIcon className="h-4 w-4 text-purple-600" />
                <span className="text-xs font-medium text-purple-900">GDPR Requests</span>
              </div>
              <div className="text-lg font-bold text-purple-600">
                {metrics.audit_compliance.gdpr_requests}
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Privacy Violations</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.audit_compliance.privacy_violations}
              </div>
            </div>
          </div>
        </div>

        {/* Feature Flags */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Feature Flags</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className={`rounded-lg p-3 ${metrics.feature_flags.zkid_enabled ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <FlagIcon className={`h-4 w-4 ${metrics.feature_flags.zkid_enabled ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.feature_flags.zkid_enabled ? 'text-green-900' : 'text-red-900'}`}>
                  ZKID Layer
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.feature_flags.zkid_enabled ? 'text-green-600' : 'text-red-600'}`}>
                {getFeatureFlagStatus(metrics.feature_flags.zkid_enabled)}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.feature_flags.pqc_enabled ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <FlagIcon className={`h-4 w-4 ${metrics.feature_flags.pqc_enabled ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.feature_flags.pqc_enabled ? 'text-green-900' : 'text-red-900'}`}>
                  PQC Encryption
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.feature_flags.pqc_enabled ? 'text-green-600' : 'text-red-600'}`}>
                {getFeatureFlagStatus(metrics.feature_flags.pqc_enabled)}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.feature_flags.enterprise_enabled ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <FlagIcon className={`h-4 w-4 ${metrics.feature_flags.enterprise_enabled ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.feature_flags.enterprise_enabled ? 'text-green-900' : 'text-red-900'}`}>
                  Enterprise
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.feature_flags.enterprise_enabled ? 'text-green-600' : 'text-red-600'}`}>
                {getFeatureFlagStatus(metrics.feature_flags.enterprise_enabled)}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.feature_flags.mfa_enabled ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <FlagIcon className={`h-4 w-4 ${metrics.feature_flags.mfa_enabled ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.feature_flags.mfa_enabled ? 'text-green-900' : 'text-red-900'}`}>
                  MFA
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.feature_flags.mfa_enabled ? 'text-green-600' : 'text-red-600'}`}>
                {getFeatureFlagStatus(metrics.feature_flags.mfa_enabled)}
              </div>
            </div>
          </div>
        </div>

        {/* Geolocation Compliance */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Geolocation Compliance</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <GlobeAltIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Geo Violations</span>
              </div>
              <div className="text-lg font-bold text-red-600">
                {metrics.geolocation_compliance.geo_restriction_violations}
              </div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <GlobeAltIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">VPN Detections</span>
              </div>
              <div className="text-lg font-bold text-yellow-600">
                {metrics.geolocation_compliance.vpn_detections}
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Suspicious Locations</span>
              </div>
              <div className="text-lg font-bold text-green-600">
                {metrics.geolocation_compliance.suspicious_locations}
              </div>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ShieldCheckIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Compliance Checks</span>
              </div>
              <div className="text-lg font-bold text-blue-600">
                {metrics.geolocation_compliance.compliance_checks.toLocaleString()}
              </div>
            </div>
          </div>
        </div>

        {/* Security Summary */}
        <div className="bg-green-50 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-5 w-5 text-green-600" />
            <div>
              <h5 className="text-sm font-medium text-green-900">Security Summary</h5>
              <p className="text-xs text-green-700">
                RBAC violations: {metrics.access_control.rbac_violations} | 
                Privacy violations: {metrics.audit_compliance.privacy_violations} | 
                Geo violations: {metrics.geolocation_compliance.geo_restriction_violations}
              </p>
            </div>
          </div>
        </div>

        {/* Compliance Status */}
        {showDetails && (
          <div className="mt-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Compliance Status</h4>
            <div className="space-y-3">
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center space-x-2">
                  <CheckCircleIcon className="h-4 w-4 text-green-600" />
                  <span className="text-sm font-medium text-gray-700">GDPR Compliance</span>
                </div>
                <span className="text-sm text-green-600 font-medium">Compliant</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center space-x-2">
                  <CheckCircleIcon className="h-4 w-4 text-green-600" />
                  <span className="text-sm font-medium text-gray-700">SOC 2 Compliance</span>
                </div>
                <span className="text-sm text-green-600 font-medium">Compliant</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center space-x-2">
                  <CheckCircleIcon className="h-4 w-4 text-green-600" />
                  <span className="text-sm font-medium text-gray-700">Zero-Knowledge Privacy</span>
                </div>
                <span className="text-sm text-green-600 font-medium">Active</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SecurityCompliancePanel;
