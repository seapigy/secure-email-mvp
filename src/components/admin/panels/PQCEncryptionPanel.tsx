import React, { useState } from 'react';
import {
  KeyIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,

  EyeIcon,
  EyeSlashIcon,

  ClockIcon,
  ServerIcon,
  ShieldCheckIcon,
  CpuChipIcon,
  LockClosedIcon,
} from '@heroicons/react/24/outline';
import { PQCEncryptionMetrics } from '../../../types/admin';

interface PQCEncryptionPanelProps {
  metrics: PQCEncryptionMetrics | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const PQCEncryptionPanel: React.FC<PQCEncryptionPanelProps> = ({
  metrics,
  isLoading,
  onRefresh,
}) => {
  const [showDetails, setShowDetails] = useState(false);
  const [showAlgorithmDetails, setShowAlgorithmDetails] = useState(false);
  const [showEnhancedMetrics, setShowEnhancedMetrics] = useState(false);

  if (!metrics) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">PQC Encryption</h3>
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
          <p className="mt-2 text-sm text-gray-500">Loading PQC metrics...</p>
        </div>
      </div>
    );
  }



  const getStatusColor = (status: string) => {
    switch (status) {
      case 'secure':
        return 'text-green-600 bg-green-100';
      case 'insecure':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };



  const getAlgorithmColor = (percentage: number) => {
    if (percentage > 50) return 'text-blue-600';
    if (percentage > 25) return 'text-green-600';
    return 'text-gray-600';
  };

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <KeyIcon className="h-6 w-6 text-purple-600" />
            <h3 className="text-lg font-medium text-gray-900">PQC Encryption</h3>
            <div className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(metrics.security_status.post_quantum_ready ? 'secure' : 'insecure')}`}>
              {metrics.security_status.post_quantum_ready ? 'PQC Ready' : 'Classical Only'}
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
        {/* Key Management */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Key Management</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <KeyIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-medium text-blue-900">Keys Generated</span>
              </div>
              <div className="text-2xl font-bold text-blue-600">{metrics.key_management.keys_generated}</div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ArrowPathIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Keys Rotated</span>
              </div>
              <div className="text-2xl font-bold text-green-600">{metrics.key_management.keys_rotated}</div>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Rotation Failures</span>
              </div>
              <div className="text-2xl font-bold text-red-600">{metrics.key_management.rotation_failures}</div>
            </div>
            <div className="bg-purple-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ServerIcon className="h-4 w-4 text-purple-600" />
                <span className="text-xs font-medium text-purple-900">HSM Operations</span>
              </div>
              <div className="text-2xl font-bold text-purple-600">{metrics.key_management.hsm_operations}</div>
            </div>
          </div>
        </div>

        {/* Encryption Performance */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Encryption Performance</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CpuChipIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">AES-256-GCM Ops</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.encryption_performance.aes_256_gcm_operations.toLocaleString()}
              </div>
              <div className="text-xs text-gray-500 mt-1">
                Avg: {metrics.encryption_performance.average_encryption_time_ms}ms
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CpuChipIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">ChaCha20 Ops</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.encryption_performance.chacha20_operations.toLocaleString()}
              </div>
              <div className="text-xs text-gray-500 mt-1">
                Avg: {metrics.encryption_performance.average_encryption_time_ms}ms
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ShieldCheckIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">Kyber Ops</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.encryption_performance.kyber_operations.toLocaleString()}
              </div>
              <div className="text-xs text-gray-500 mt-1">
                Post-Quantum Ready
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ClockIcon className="h-4 w-4 text-gray-600" />
                <span className="text-xs font-medium text-gray-700">Decryption Time</span>
              </div>
              <div className="text-lg font-bold text-gray-900">
                {metrics.encryption_performance.average_decryption_time_ms}ms
              </div>
              <div className="text-xs text-gray-500 mt-1">
                Average
              </div>
            </div>
          </div>
        </div>

        {/* Algorithm Usage */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Algorithm Usage</h4>
            <button
              onClick={() => setShowAlgorithmDetails(!showAlgorithmDetails)}
              className="text-xs text-indigo-600 hover:text-indigo-500"
            >
              {showAlgorithmDetails ? 'Hide' : 'Show'} Details
            </button>
          </div>
          <div className="space-y-3">
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">AES-256-GCM</span>
                <span className={`text-sm font-bold ${getAlgorithmColor(metrics.algorithm_usage.aes_256_gcm_percentage)}`}>
                  {metrics.algorithm_usage.aes_256_gcm_percentage.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full"
                  style={{ width: `${metrics.algorithm_usage.aes_256_gcm_percentage}%` }}
                ></div>
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">ChaCha20</span>
                <span className={`text-sm font-bold ${getAlgorithmColor(metrics.algorithm_usage.chacha20_percentage)}`}>
                  {metrics.algorithm_usage.chacha20_percentage.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-green-600 h-2 rounded-full"
                  style={{ width: `${metrics.algorithm_usage.chacha20_percentage}%` }}
                ></div>
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">Kyber (PQC)</span>
                <span className={`text-sm font-bold ${getAlgorithmColor(metrics.algorithm_usage.kyber_percentage)}`}>
                  {metrics.algorithm_usage.kyber_percentage.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-purple-600 h-2 rounded-full"
                  style={{ width: `${metrics.algorithm_usage.kyber_percentage}%` }}
                ></div>
              </div>
            </div>
            <div className="bg-gray-50 rounded-lg p-3">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">Hybrid (Classical + PQC)</span>
                <span className={`text-sm font-bold ${getAlgorithmColor(metrics.algorithm_usage.hybrid_percentage)}`}>
                  {metrics.algorithm_usage.hybrid_percentage.toFixed(1)}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-indigo-600 h-2 rounded-full"
                  style={{ width: `${metrics.algorithm_usage.hybrid_percentage}%` }}
                ></div>
              </div>
            </div>
          </div>

          {/* Algorithm Details */}
          {showAlgorithmDetails && (
            <div className="mt-4 bg-indigo-50 rounded-lg p-3">
              <h5 className="text-xs font-medium text-indigo-900 mb-2">Algorithm Details</h5>
              <div className="space-y-2 text-xs text-indigo-700">
                <div className="flex justify-between">
                  <span>AES-256-GCM:</span>
                  <span>Classical encryption, high performance</span>
                </div>
                <div className="flex justify-between">
                  <span>ChaCha20:</span>
                  <span>Stream cipher, constant-time operations</span>
                </div>
                <div className="flex justify-between">
                  <span>Kyber:</span>
                  <span>Post-quantum key encapsulation</span>
                </div>
                <div className="flex justify-between">
                  <span>Hybrid:</span>
                  <span>Combines classical + PQC for security</span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Security Status */}
        <div>
          <h4 className="text-sm font-medium text-gray-900 mb-3">Security Status</h4>
          <div className="grid grid-cols-2 gap-4">
            <div className={`rounded-lg p-3 ${metrics.security_status.hsm_available ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <ServerIcon className={`h-4 w-4 ${metrics.security_status.hsm_available ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.security_status.hsm_available ? 'text-green-900' : 'text-red-900'}`}>
                  HSM Available
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.security_status.hsm_available ? 'text-green-600' : 'text-red-600'}`}>
                {metrics.security_status.hsm_available ? 'Yes' : 'No'}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.security_status.key_store_encrypted ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <LockClosedIcon className={`h-4 w-4 ${metrics.security_status.key_store_encrypted ? 'text-green-600' : 'text-red-600'}`} />
                <span className={`text-xs font-medium ${metrics.security_status.key_store_encrypted ? 'text-green-900' : 'text-red-900'}`}>
                  Key Store Encrypted
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.security_status.key_store_encrypted ? 'text-green-600' : 'text-red-600'}`}>
                {metrics.security_status.key_store_encrypted ? 'Yes' : 'No'}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.security_status.rotation_schedule_compliant ? 'bg-green-50' : 'bg-yellow-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <ArrowPathIcon className={`h-4 w-4 ${metrics.security_status.rotation_schedule_compliant ? 'text-green-600' : 'text-yellow-600'}`} />
                <span className={`text-xs font-medium ${metrics.security_status.rotation_schedule_compliant ? 'text-green-900' : 'text-yellow-900'}`}>
                  Rotation Compliant
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.security_status.rotation_schedule_compliant ? 'text-green-600' : 'text-yellow-600'}`}>
                {metrics.security_status.rotation_schedule_compliant ? 'Yes' : 'No'}
              </div>
            </div>
            <div className={`rounded-lg p-3 ${metrics.security_status.post_quantum_ready ? 'bg-green-50' : 'bg-orange-50'}`}>
              <div className="flex items-center space-x-2 mb-2">
                <ShieldCheckIcon className={`h-4 w-4 ${metrics.security_status.post_quantum_ready ? 'text-green-600' : 'text-orange-600'}`} />
                <span className={`text-xs font-medium ${metrics.security_status.post_quantum_ready ? 'text-green-900' : 'text-orange-900'}`}>
                  Post-Quantum Ready
                </span>
              </div>
              <div className={`text-lg font-bold ${metrics.security_status.post_quantum_ready ? 'text-green-600' : 'text-orange-600'}`}>
                {metrics.security_status.post_quantum_ready ? 'Yes' : 'No'}
              </div>
            </div>
          </div>
        </div>

        {/* Error Tracking */}
        {showDetails && (
          <div className="mt-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Error Tracking</h4>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-red-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                  <span className="text-xs font-medium text-red-900">Encryption Errors</span>
                </div>
                <div className="text-lg font-bold text-red-600">
                  {metrics.encryption_performance.encryption_errors}
                </div>
              </div>
              <div className="bg-red-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                  <span className="text-xs font-medium text-red-900">Decryption Errors</span>
                </div>
                <div className="text-lg font-bold text-red-600">
                  {metrics.encryption_performance.decryption_errors}
                </div>
              </div>
              <div className="bg-red-50 rounded-lg p-3">
                <div className="flex items-center space-x-2 mb-2">
                  <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                  <span className="text-xs font-medium text-red-900">HSM Errors</span>
                </div>
                <div className="text-lg font-bold text-red-600">
                  {metrics.key_management.hsm_errors}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Enhanced PQC Monitoring */}
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Enhanced PQC Monitoring</h4>
            <button
              onClick={() => setShowEnhancedMetrics(!showEnhancedMetrics)}
              className="text-xs text-purple-600 hover:text-purple-800 focus:outline-none"
            >
              {showEnhancedMetrics ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          {showEnhancedMetrics && (
            <div className="space-y-4">
              {/* Key Rotation Schedule */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Key Rotation Schedule</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Next Rotation</div>
                    <div className="text-sm font-bold text-gray-900">
                      {new Date(metrics.key_rotation_schedule.next_rotation).toLocaleString()}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Last Rotation</div>
                    <div className="text-sm font-bold text-gray-900">
                      {new Date(metrics.key_rotation_schedule.last_rotation).toLocaleString()}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Rotation Interval</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.key_rotation_schedule.rotation_interval_hours}h</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Grace Period</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.key_rotation_schedule.grace_period_hours}h</div>
                  </div>
                </div>
              </div>

              {/* Key Health Status */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Key Health Status</h5>
                <div className="grid grid-cols-5 gap-3">
                  <div className="bg-green-50 rounded-lg p-3">
                    <div className="text-xs text-green-600">Active Keys</div>
                    <div className="text-lg font-bold text-green-900">{metrics.key_health_status.active_keys}</div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Expiring</div>
                    <div className="text-lg font-bold text-yellow-900">{metrics.key_health_status.expiring_keys}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Expired</div>
                    <div className="text-lg font-bold text-red-900">{metrics.key_health_status.expired_keys}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Compromised</div>
                    <div className="text-lg font-bold text-red-900">{metrics.key_health_status.compromised_keys}</div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Strength Score</div>
                    <div className="text-lg font-bold text-blue-900">{metrics.key_health_status.key_strength_score}%</div>
                  </div>
                </div>
              </div>

              {/* AEAD Encryption Stats */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">AEAD Encryption Statistics</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-green-50 rounded-lg p-3">
                    <div className="text-xs text-green-600">Successful Encryptions</div>
                    <div className="text-lg font-bold text-green-900">{metrics.aead_encryption_stats.successful_encryptions.toLocaleString()}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Failed Encryptions</div>
                    <div className="text-lg font-bold text-red-900">{metrics.aead_encryption_stats.failed_encryptions}</div>
                  </div>
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Tag Verification Errors</div>
                    <div className="text-lg font-bold text-red-900">{metrics.aead_encryption_stats.tag_verification_errors}</div>
                  </div>
                  <div className="bg-orange-50 rounded-lg p-3">
                    <div className="text-xs text-orange-600">Nonce Reuse Attempts</div>
                    <div className="text-lg font-bold text-orange-900">{metrics.aead_encryption_stats.nonce_reuse_attempts}</div>
                  </div>
                  <div className="bg-orange-50 rounded-lg p-3">
                    <div className="text-xs text-orange-600">Ciphertext Tampering</div>
                    <div className="text-lg font-bold text-orange-900">{metrics.aead_encryption_stats.ciphertext_tampering_attempts}</div>
                  </div>
                </div>
              </div>

              {/* Performance Metrics */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Performance Metrics</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Encryption Throughput</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.performance_metrics.encryption_throughput_ops_per_sec} ops/sec</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Decryption Throughput</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.performance_metrics.decryption_throughput_ops_per_sec} ops/sec</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Key Generation Time</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.performance_metrics.key_generation_time_ms}ms</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Key Rotation Time</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.performance_metrics.key_rotation_time_ms}ms</div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">HSM Latency</div>
                    <div className="text-sm font-bold text-gray-900">{metrics.performance_metrics.hsm_latency_ms}ms</div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Post-Quantum Security Notice */}
        <div className="mt-6 p-3 bg-purple-50 rounded-lg border border-purple-200">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-5 w-5 text-purple-600" />
            <div>
              <h5 className="text-sm font-medium text-purple-900">Post-Quantum Cryptography</h5>
              <p className="text-xs text-purple-700">
                System supports both classical and post-quantum algorithms for future-proof security.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PQCEncryptionPanel;
