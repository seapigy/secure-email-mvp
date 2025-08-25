import React from 'react';
import {
  AlertTriangle,
  Shield,
  Eye,
  XCircle,
  Info,
  CheckCircle,
  Clock,
  FileText,
  CreditCard,
  User,
  Phone,
  Mail
} from 'lucide-react';

interface DLPViolation {
  scan_id: string;
  rule_id: string;
  content_type: string;
  matched_content?: string;
  confidence_score: number;
  action_taken: string;
  scan_timestamp: string;
  // AI DLP specific fields
  severity_score?: number;
  risk_level?: string;
  classification?: {
    category: string;
    confidence: number;
    severity: string;
    risk_score: number;
    keywords: string[];
    entities: Array<{
      type: string;
      value: string;
      confidence: number;
    }>;
    context: string;
    model_version: string;
  };
}

interface DLPViolationDisplayProps {
  violations: DLPViolation[];
  actionTaken: string;
  onAcknowledge?: () => void;
  onOverride?: () => void;
  onCancel?: () => void;
  isOpen: boolean;
}

const DLPViolationDisplay: React.FC<DLPViolationDisplayProps> = ({
  violations,
  actionTaken,
  onAcknowledge,
  onOverride,
  onCancel,
  isOpen
}) => {
  if (!isOpen) return null;

  const getSeverityColor = (action: string) => {
    switch (action) {
      case 'blocked':
        return 'text-red-600 bg-red-50 border-red-200';
      case 'warned':
        return 'text-orange-600 bg-orange-50 border-orange-200';
      case 'allowed':
        return 'text-green-600 bg-green-50 border-green-200';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getSeverityIcon = (action: string) => {
    switch (action) {
      case 'blocked':
        return <XCircle className="h-5 w-5 text-red-600" />;
      case 'warned':
        return <AlertTriangle className="h-5 w-5 text-orange-600" />;
      case 'allowed':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      default:
        return <Info className="h-5 w-5 text-gray-600" />;
    }
  };

  const getRuleIcon = (ruleId: string) => {
    if (ruleId.includes('credit')) {
      return <CreditCard className="h-4 w-4" />;
    } else if (ruleId.includes('ssn') || ruleId.includes('social')) {
      return <User className="h-4 w-4" />;
    } else if (ruleId.includes('phone')) {
      return <Phone className="h-4 w-4" />;
    } else if (ruleId.includes('email')) {
      return <Mail className="h-4 w-4" />;
    } else {
      return <FileText className="h-4 w-4" />;
    }
  };

  const getRuleName = (ruleId: string) => {
    switch (ruleId) {
      case 'dlp_001':
        return 'Credit Card Numbers';
      case 'dlp_002':
        return 'Social Security Numbers';
      case 'dlp_003':
        return 'Email Addresses';
      case 'dlp_004':
        return 'Phone Numbers';
      case 'dlp_005':
        return 'Confidential Keywords';
      case 'dlp_006':
        return 'Financial Keywords';
      default:
        return 'Unknown Rule';
    }
  };

  const getActionDescription = (action: string) => {
    switch (action) {
      case 'blocked':
        return 'This content has been blocked due to sensitive information detected.';
      case 'warned':
        return 'Sensitive information detected. Please review before proceeding.';
      case 'allowed':
        return 'Content allowed with monitoring.';
      default:
        return 'Content processed normally.';
    }
  };

  const canProceed = actionTaken !== 'blocked';
  const hasWarnings = violations.some(v => v.action_taken === 'warned');

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className={`flex items-center justify-between p-6 border-b ${getSeverityColor(actionTaken)}`}>
          <div className="flex items-center space-x-3">
            {getSeverityIcon(actionTaken)}
            <div>
              <h2 className="text-xl font-semibold">Data Loss Prevention Alert</h2>
              <p className="text-sm opacity-80">
                {getActionDescription(actionTaken)}
              </p>
            </div>
          </div>
          <button
            onClick={onCancel}
            className="text-gray-400 hover:text-gray-600"
          >
            <XCircle className="h-6 w-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4 overflow-y-auto max-h-[calc(90vh-200px)]">
          {/* Summary */}
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-medium text-gray-900">
                  {violations.length} violation{violations.length !== 1 ? 's' : ''} detected
                </h3>
                <p className="text-sm text-gray-600">
                  Confidence: {Math.round((violations.reduce((sum, v) => sum + v.confidence_score, 0) / violations.length) * 100)}%
                  {violations.some(v => v.severity_score !== undefined) && (
                    <span className="ml-2">
                      | Risk Score: {Math.round((violations.reduce((sum, v) => sum + (v.severity_score || 0), 0) / violations.length) * 100)}%
                    </span>
                  )}
                </p>
              </div>
              <div className={`px-3 py-1 rounded-full text-sm font-medium ${getSeverityColor(actionTaken)}`}>
                {actionTaken.toUpperCase()}
              </div>
            </div>
          </div>

          {/* Violations List */}
          <div className="space-y-3">
            {violations.map((violation) => (
              <div
                key={violation.scan_id}
                className={`border rounded-lg p-4 ${getSeverityColor(violation.action_taken)}`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start space-x-3">
                    {getRuleIcon(violation.rule_id)}
                    <div className="flex-1">
                      <h4 className="font-medium text-gray-900">
                        {getRuleName(violation.rule_id)}
                      </h4>
                      <p className="text-sm text-gray-600">
                        Content Type: {violation.content_type.replace('_', ' ')}
                      </p>
                      {violation.matched_content && (
                        <div className="mt-2">
                          <p className="text-xs text-gray-500 mb-1">Matched Content:</p>
                          <div className="bg-gray-100 rounded px-2 py-1 text-xs font-mono break-all">
                            {violation.matched_content.length > 100
                              ? `${violation.matched_content.substring(0, 100)}...`
                              : violation.matched_content
                            }
                          </div>
                        </div>
                      )}
                      <div className="flex items-center space-x-4 mt-2 text-xs text-gray-500">
                        <span>Confidence: {Math.round(violation.confidence_score * 100)}%</span>
                        {violation.severity_score !== undefined && (
                          <span>Risk Score: {Math.round(violation.severity_score * 100)}%</span>
                        )}
                        {violation.risk_level && (
                          <span>Risk Level: {violation.risk_level}</span>
                        )}
                        <span>Action: {violation.action_taken}</span>
                        <span>
                          <Clock className="h-3 w-3 inline mr-1" />
                          {new Date(violation.scan_timestamp).toLocaleString()}
                        </span>
                      </div>
                      
                      {/* AI Classification Details */}
                      {violation.classification && (
                        <div className="mt-3 p-3 bg-blue-50 rounded-lg">
                          <div className="flex items-center space-x-2 mb-2">
                            <Shield className="h-4 w-4 text-blue-600" />
                            <span className="text-sm font-medium text-blue-800">AI Classification</span>
                          </div>
                          <div className="text-xs text-blue-700 space-y-1">
                            <div><strong>Category:</strong> {violation.classification.category}</div>
                            <div><strong>Context:</strong> {violation.classification.context}</div>
                            {violation.classification.keywords.length > 0 && (
                              <div><strong>Keywords:</strong> {violation.classification.keywords.join(', ')}</div>
                            )}
                            {violation.classification.entities.length > 0 && (
                              <div><strong>Entities:</strong> {violation.classification.entities.length} detected</div>
                            )}
                            <div><strong>Model:</strong> {violation.classification.model_version}</div>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className={`px-2 py-1 rounded text-xs font-medium ${getSeverityColor(violation.action_taken)}`}>
                    {violation.action_taken.toUpperCase()}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Security Notice */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start space-x-2">
              <Shield className="h-5 w-5 text-blue-600 mt-0.5 flex-shrink-0" />
              <div>
                <h4 className="text-sm font-medium text-blue-800">Security Notice</h4>
                <p className="text-sm text-blue-700 mt-1">
                  This system automatically scans content for sensitive information to prevent data loss. 
                  All violations are logged for compliance and security monitoring purposes.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200">
          <div className="flex items-center space-x-2">
            <Eye className="h-4 w-4 text-gray-400" />
            <span className="text-sm text-gray-500">
              All actions are logged for compliance
            </span>
          </div>
          <div className="flex space-x-3">
            {onCancel && (
              <button
                onClick={onCancel}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
            )}
            {onAcknowledge && canProceed && (
              <button
                onClick={onAcknowledge}
                className={`px-4 py-2 rounded-md transition-colors ${
                  hasWarnings
                    ? 'bg-orange-600 text-white hover:bg-orange-700'
                    : 'bg-green-600 text-white hover:bg-green-700'
                }`}
              >
                {hasWarnings ? 'Acknowledge & Continue' : 'Continue'}
              </button>
            )}
            {onOverride && !canProceed && (
              <button
                onClick={onOverride}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
              >
                Override (Admin Only)
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default DLPViolationDisplay;
