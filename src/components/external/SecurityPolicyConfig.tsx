import React, { useState, useEffect, useCallback } from 'react';
import {
  Shield,
  Eye,
  Download,
  Upload,
  Clock,
  Lock,
  AlertTriangle,
  CheckCircle,
  Save,
  X
} from 'lucide-react';
import { log } from '@/lib/logger';

interface SecurityPolicy {
  policy_id?: string;
  link_id: string;
  dlp_enabled: boolean;
  ai_dlp_enabled?: boolean; // AI-powered DLP
  watermark_enabled: boolean;
  // Advanced watermarking options (Iteration 8)
  advanced_watermarking_enabled?: boolean;
  watermark_type?: string; // 'text', 'image', 'audio', 'video', 'inline'
  recipient_specific_watermark?: boolean;
  watermark_template_id?: string;
  download_disabled: boolean;
  forwarding_disabled: boolean;
  auto_revoke_after_reply: boolean;
  max_views?: number;
  expires_at?: string;
  expires_after_views?: number;
  notify_on_expiry: boolean;
  notify_on_revoke: boolean;
}

interface SecurityPolicyTemplate {
  template_id: string;
  template_name: string;
  template_description?: string;
  dlp_enabled: boolean;
  watermark_enabled: boolean;
  download_disabled: boolean;
  forwarding_disabled: boolean;
  auto_revoke_after_reply: boolean;
  max_views?: number;
  default_expiry_hours?: number;
  notify_on_expiry: boolean;
  notify_on_revoke: boolean;
  is_default: boolean;
}

interface SecurityPolicyConfigProps {
  linkID: string;
  onPolicySaved?: (policy: SecurityPolicy) => void;
  onCancel?: () => void;
  isOpen: boolean;
}

const SecurityPolicyConfig: React.FC<SecurityPolicyConfigProps> = ({
  linkID,
  onPolicySaved,
  onCancel,
  isOpen
}) => {
  const [policy, setPolicy] = useState<SecurityPolicy>({
    link_id: linkID,
    dlp_enabled: true,
    ai_dlp_enabled: true, // Enable AI DLP by default
    watermark_enabled: true,
    advanced_watermarking_enabled: false, // Disable advanced watermarking by default
    watermark_type: 'text',
    recipient_specific_watermark: false,
    download_disabled: false,
    forwarding_disabled: false,
    auto_revoke_after_reply: false,
    notify_on_expiry: true,
    notify_on_revoke: true
  });

  const [templates, setTemplates] = useState<SecurityPolicyTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const loadTemplates = useCallback(async () => {
    try {
      const response = await fetch('/api/security/templates');
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          setTemplates(data.templates || []);
        }
      }
    } catch (error) {
      log.error('Error loading templates:', error, 'SecurityPolicyConfig');
      setError('Failed to load policy templates');
    }
  }, []);

  const loadExistingPolicy = useCallback(async () => {
    try {
      const response = await fetch(`/api/v/${linkID}/security/policy`);
      if (response.ok) {
        const data = await response.json();
        if (data.success && data.policy) {
          setPolicy(data.policy);
        }
      }
    } catch (error) {
      log.error('Error loading existing policy:', error, 'SecurityPolicyConfig');
      setError('Failed to load existing policy');
    }
  }, [linkID]);

  // Load security policy templates
  useEffect(() => {
    if (isOpen) {
      loadTemplates();
      loadExistingPolicy();
    }
  }, [isOpen, linkID, loadTemplates, loadExistingPolicy]);

  const applyTemplate = (template: SecurityPolicyTemplate) => {
    setPolicy({
      ...policy,
      dlp_enabled: template.dlp_enabled,
      watermark_enabled: template.watermark_enabled,
      download_disabled: template.download_disabled,
      forwarding_disabled: template.forwarding_disabled,
      auto_revoke_after_reply: template.auto_revoke_after_reply,
      max_views: template.max_views,
      notify_on_expiry: template.notify_on_expiry,
      notify_on_revoke: template.notify_on_revoke
    });
    setSelectedTemplate(template.template_id);
  };

  const handleSave = async () => {
    setLoading(true);
    setError(null);

    try {
      const method = policy.policy_id ? 'PUT' : 'POST';
      const response = await fetch(`/api/v/${linkID}/security/policy`, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(policy),
      });

      const data = await response.json();

      if (response.ok && data.success) {
        setSuccess(true);
        if (onPolicySaved) {
          onPolicySaved(data.policy || policy);
        }
        setTimeout(() => {
          setSuccess(false);
          if (onCancel) onCancel();
        }, 2000);
      } else {
        setError(data.error || 'Failed to save security policy');
      }
    } catch (error) {
      log.error('Error saving policy:', error, 'SecurityPolicyConfig');
      setError('Failed to save policy');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    if (!loading && onCancel) {
      onCancel();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <Shield className="h-6 w-6 text-blue-600" />
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Security Policy Configuration</h2>
              <p className="text-sm text-gray-600">Configure advanced security controls for this secure link</p>
            </div>
          </div>
          <button
            onClick={handleCancel}
            className="text-gray-400 hover:text-gray-600"
            disabled={loading}
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6 overflow-y-auto max-h-[calc(90vh-140px)]">
          {/* Template Selection */}
          <div>
            <h3 className="text-lg font-medium text-gray-900 mb-3">Quick Templates</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
              {templates.map((template) => (
                <button
                  key={template.template_id}
                  onClick={() => applyTemplate(template)}
                  className={`p-3 border rounded-lg text-left transition-colors ${
                    selectedTemplate === template.template_id
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-300 hover:border-gray-400'
                  }`}
                  disabled={loading}
                >
                  <div className="font-medium text-sm text-gray-900">{template.template_name}</div>
                  {template.template_description && (
                    <div className="text-xs text-gray-500 mt-1">{template.template_description}</div>
                  )}
                  {template.is_default && (
                    <div className="text-xs text-blue-600 mt-1 font-medium">Default</div>
                  )}
                </button>
              ))}
            </div>
          </div>

          {/* DLP Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5 text-orange-600" />
                <h3 className="text-lg font-medium text-gray-900">Data Loss Prevention (DLP)</h3>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.dlp_enabled}
                  onChange={(e) => setPolicy({ ...policy, dlp_enabled: e.target.checked })}
                  className="sr-only peer"
                  disabled={loading}
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
            <p className="text-sm text-gray-600">
              Automatically scan content for sensitive information like credit card numbers, 
              social security numbers, and confidential keywords.
            </p>
          </div>

          {/* AI DLP Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <Shield className="h-5 w-5 text-purple-600" />
                <h3 className="text-lg font-medium text-gray-900">AI-Powered DLP</h3>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.ai_dlp_enabled || false}
                  onChange={(e) => setPolicy({ ...policy, ai_dlp_enabled: e.target.checked })}
                  className="sr-only peer"
                  disabled={loading}
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
            <p className="text-sm text-gray-600">
              Advanced AI-powered content classification with severity scoring, entity extraction, 
              and contextual analysis for enhanced data loss prevention.
            </p>
            {policy.ai_dlp_enabled && (
              <div className="mt-3 p-3 bg-purple-50 rounded-lg">
                <div className="text-xs text-purple-700 space-y-1">
                  <div><strong>Features:</strong> NLP classification, entity extraction, severity scoring</div>
                  <div><strong>Categories:</strong> PII, Financial, Healthcare, Legal, Confidential</div>
                  <div><strong>Override:</strong> Available for authorized roles</div>
                </div>
              </div>
            )}
          </div>

          {/* Watermarking Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <Upload className="h-5 w-5 text-purple-600" />
                <h3 className="text-lg font-medium text-gray-900">Document Watermarking</h3>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.watermark_enabled}
                  onChange={(e) => setPolicy({ ...policy, watermark_enabled: e.target.checked })}
                  className="sr-only peer"
                  disabled={loading}
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
            <p className="text-sm text-gray-600">
              Apply watermarks to downloaded documents with recipient information and timestamps.
            </p>
            
            {/* Advanced Watermarking Options */}
            {policy.watermark_enabled && (
              <div className="mt-4 p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <Upload className="h-4 w-4 text-purple-600" />
                    <span className="text-sm font-medium text-gray-900">Advanced Watermarking</span>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={policy.advanced_watermarking_enabled || false}
                      onChange={(e) => setPolicy({ ...policy, advanced_watermarking_enabled: e.target.checked })}
                      className="sr-only peer"
                      disabled={loading}
                    />
                    <div className="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-purple-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-purple-600"></div>
                  </label>
                </div>
                
                {policy.advanced_watermarking_enabled && (
                  <div className="space-y-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">
                        Watermark Type
                      </label>
                      <select
                        value={policy.watermark_type || 'text'}
                        onChange={(e) => setPolicy({ ...policy, watermark_type: e.target.value })}
                        className="w-full text-xs border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-purple-500"
                        disabled={loading}
                      >
                        <option value="text">Text Watermark</option>
                        <option value="image">Image Watermark</option>
                        <option value="audio">Audio Watermark</option>
                        <option value="video">Video Overlay</option>
                        <option value="inline">Inline Content</option>
                      </select>
                    </div>
                    
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Shield className="h-3 w-3 text-purple-600" />
                        <span className="text-xs text-gray-700">Recipient-Specific</span>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={policy.recipient_specific_watermark || false}
                          onChange={(e) => setPolicy({ ...policy, recipient_specific_watermark: e.target.checked })}
                          className="sr-only peer"
                          disabled={loading}
                        />
                        <div className="w-7 h-4 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-purple-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[1px] after:left-[1px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-3 after:w-3 after:transition-all peer-checked:bg-purple-600"></div>
                      </label>
                    </div>
                    
                    <div className="text-xs text-purple-700 space-y-1">
                      <div><strong>Features:</strong> Multi-format support, recipient identification, audit trails</div>
                      <div><strong>Formats:</strong> PDF, Images, Audio, Video, Inline content</div>
                      <div><strong>Security:</strong> Tamper-resistant, timestamped, traceable</div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Access Control Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Access Controls</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Download className="h-5 w-5 text-green-600" />
                  <span className="text-sm font-medium text-gray-900">Allow Downloads</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={!policy.download_disabled}
                    onChange={(e) => setPolicy({ ...policy, download_disabled: !e.target.checked })}
                    className="sr-only peer"
                    disabled={loading}
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Eye className="h-5 w-5 text-blue-600" />
                  <span className="text-sm font-medium text-gray-900">Allow Forwarding</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={!policy.forwarding_disabled}
                    onChange={(e) => setPolicy({ ...policy, forwarding_disabled: !e.target.checked })}
                    className="sr-only peer"
                    disabled={loading}
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Lock className="h-5 w-5 text-red-600" />
                  <span className="text-sm font-medium text-gray-900">Auto-Revoke After Reply</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={policy.auto_revoke_after_reply}
                    onChange={(e) => setPolicy({ ...policy, auto_revoke_after_reply: e.target.checked })}
                    className="sr-only peer"
                    disabled={loading}
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>
            </div>
          </div>

          {/* Expiration Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-4">
              <Clock className="h-5 w-5 text-yellow-600" />
              <h3 className="text-lg font-medium text-gray-900">Expiration Settings</h3>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Maximum Views
                </label>
                <input
                  type="number"
                  min="1"
                  value={policy.max_views || ''}
                  onChange={(e) => setPolicy({ 
                    ...policy, 
                    max_views: e.target.value ? parseInt(e.target.value) : undefined 
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="No limit"
                  disabled={loading}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Link will expire after this many views (leave empty for no limit)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Expires After Views
                </label>
                <input
                  type="number"
                  min="1"
                  value={policy.expires_after_views || ''}
                  onChange={(e) => setPolicy({ 
                    ...policy, 
                    expires_after_views: e.target.value ? parseInt(e.target.value) : undefined 
                  })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="No limit"
                  disabled={loading}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Link will expire after this many views (alternative to max_views)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Expires At
                </label>
                <input
                  type="datetime-local"
                  value={policy.expires_at || ''}
                  onChange={(e) => setPolicy({ ...policy, expires_at: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  disabled={loading}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Link will expire at this specific date and time
                </p>
              </div>
            </div>
          </div>

          {/* Notification Settings */}
          <div className="border border-gray-200 rounded-lg p-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Notifications</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <AlertTriangle className="h-5 w-5 text-orange-600" />
                  <span className="text-sm font-medium text-gray-900">Notify on Expiry</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={policy.notify_on_expiry}
                    onChange={(e) => setPolicy({ ...policy, notify_on_expiry: e.target.checked })}
                    className="sr-only peer"
                    disabled={loading}
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Lock className="h-5 w-5 text-red-600" />
                  <span className="text-sm font-medium text-gray-900">Notify on Revoke</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={policy.notify_on_revoke}
                    onChange={(e) => setPolicy({ ...policy, notify_on_revoke: e.target.checked })}
                    className="sr-only peer"
                    disabled={loading}
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-gray-200">
          <div className="flex items-center space-x-2">
            {success && (
              <div className="flex items-center space-x-2 text-green-600">
                <CheckCircle className="h-5 w-5" />
                <span className="text-sm font-medium">Policy saved successfully!</span>
              </div>
            )}
            {error && (
              <div className="flex items-center space-x-2 text-red-600">
                <AlertTriangle className="h-5 w-5" />
                <span className="text-sm font-medium">{error}</span>
              </div>
            )}
          </div>
          <div className="flex space-x-3">
            <button
              onClick={handleCancel}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={loading}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Policy
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SecurityPolicyConfig;
