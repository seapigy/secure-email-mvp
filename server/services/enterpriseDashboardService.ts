import axios, { AxiosInstance, AxiosResponse } from 'axios';
import {
  AdminAuthConfig, AdminUser, AdminLoginRequest, AdminLoginResponse, MFASetupRequest, MFASetupResponse,
  InvitationKey, ZKIDLayerMetrics, PQCEncryptionMetrics, EmailDeliveryMetrics, SecurityComplianceMetrics,
  PerformanceOperationalMetrics, Alert, AuditLogEntry, AuditLogFilter, DashboardConfig,
  APIResponse, RealTimeUpdate, AdminPermissions, AdminInvitationRequest, AdminInvitationResponse,
  AdminActionApproval, AnalyticsFilter, AnalyticsDashboard, AnalyticsTimeRange, EmailUsageAnalytics,
  SecurityEventAnalytics, ZKIDPQCAnalytics, AnalyticsExportRequest, AnalyticsExportResponse,
  ThreatIntelligence, ThreatAlert, ThreatFeed, ThreatRule, ThreatAwarenessConfig,
} from '../types/admin';
import { ErrorResponse } from '../types';
import { log } from '@/lib/logger';

export class EnterpriseDashboardService {
  private api: AxiosInstance;
  private adminToken: string | null = null;
  private currentUser: AdminUser | null = null;
  private refreshInterval: NodeJS.Timeout | null = null;

  constructor(adminToken?: string) {
    this.adminToken = adminToken || null;
    this.api = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || '',
      timeout: 30000,
    });

    // Add request interceptor to include auth token
    this.api.interceptors.request.use((config) => {
      if (this.adminToken) {
        config.headers.Authorization = `Bearer ${this.adminToken}`;
      }
      return config;
    });
  }

  /**
   * Helper function to extract error message from caught error
   */
  private extractErrorMessage(error: unknown, defaultMessage: string): string {
    const apiError = error as { response?: { data?: ErrorResponse } };
    return apiError.response?.data?.error || defaultMessage;
  }

  // ============================================================================
  // AUTHENTICATION & AUTHORIZATION - MULTI-ADMIN
  // ============================================================================

  setAdminToken(token: string) {
    this.adminToken = token;
  }

  getCurrentUser(): AdminUser | null {
    return this.currentUser;
  }

  async login(credentials: AdminLoginRequest): Promise<AdminLoginResponse> {
    try {
      const response: AxiosResponse<APIResponse<AdminLoginResponse>> = await this.api.post(
        '/api/admin/auth/login',
        credentials
      );
      
      if (response.data.data?.user) {
        this.currentUser = response.data.data.user;
      }
      
      return response.data.data!;
    } catch {
      throw new Error('Login failed');
    }
  }

  async logout(): Promise<void> {
    try {
      await this.api.post('/api/admin/auth/logout');
    } catch (error) {
      log.error('Logout error:', error, 'enterpriseDashboardService');
    } finally {
      this.adminToken = null;
      this.currentUser = null;
    }
  }

  async setupMFA(request: MFASetupRequest): Promise<MFASetupResponse> {
    try {
      const response: AxiosResponse<APIResponse<MFASetupResponse>> = await this.api.post(
        '/api/admin/auth/mfa/setup',
        request
      );
      return response.data.data!;
    } catch {
      throw new Error('MFA setup failed');
    }
  }

  async verifyMFA(code: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.post(
        '/api/admin/auth/mfa/verify',
        { code }
      );
      return response.data.data!.success;
    } catch {
      throw new Error('MFA verification failed');
    }
  }

  async getAuthConfig(): Promise<AdminAuthConfig> {
    try {
      const response: AxiosResponse<APIResponse<AdminAuthConfig>> = await this.api.get(
        '/api/admin/auth/config'
      );
      return response.data.data!;
    } catch {
      // Return default config if API not available
      return {
        enabled: true,
        require_mfa: true,
        mfa_type: 'TOTP',
        session_timeout_minutes: 1440, // 24 hours
        max_failed_attempts: 5,
        lockout_duration_minutes: 30,
        invitation_key_expiry_hours: 24,
        max_secondary_admins: 10,
        require_primary_approval: true,
      };
    }
  }

  // ============================================================================
  // ADMIN MANAGEMENT - MULTI-ADMIN
  // ============================================================================

  async createInvitationKey(request: AdminInvitationRequest): Promise<AdminInvitationResponse> {
    try {
      const response: AxiosResponse<APIResponse<AdminInvitationResponse>> = await this.api.post(
        '/api/admin/invitations',
        request
      );
      return response.data.data!;
    } catch {
      throw new Error('Failed to create invitation');
    }
  }

  async listInvitationKeys(): Promise<InvitationKey[]> {
    try {
      const response: AxiosResponse<APIResponse<InvitationKey[]>> = await this.api.get(
        '/api/admin/invitations'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return [
        {
          key: 'inv-test-123',
          created_by: 'primary-admin',
          expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
          scope: 'full_access',
          used: false,
          max_uses: 1,
          current_uses: 0,
        },
      ];
    }
  }

  async revokeInvitationKey(key: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.delete(
        `/api/admin/invitations/${key}`
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to revoke invitation');
    }
  }

  async listAdmins(): Promise<AdminUser[]> {
    try {
      const response: AxiosResponse<APIResponse<AdminUser[]>> = await this.api.get(
        '/api/admin/users'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return [
        {
          id: 'admin-1',
          username: 'primary-admin',
          role: 'primary_admin',
          mfa_enabled: true,
          mfa_type: 'TOTP',
          last_login: new Date().toISOString(),
          failed_attempts: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          is_active: true,
          permissions: {
            can_manage_system: true,
            can_manage_admins: true,
            can_manage_organizations: true,
            can_manage_users: true,
            can_view_sensitive_data: true,
            can_export_data: true,
            can_modify_settings: true,
            can_approve_actions: true,
            can_access_audit_logs: true,
            can_manage_feature_flags: true,
          },
        },
        {
          id: 'admin-2',
          username: 'secondary-admin',
          role: 'full_admin',
          mfa_enabled: true,
          mfa_type: 'TOTP',
          last_login: new Date().toISOString(),
          failed_attempts: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          created_by: 'primary-admin',
          is_active: true,
          permissions: {
            can_manage_system: false,
            can_manage_admins: false,
            can_manage_organizations: true,
            can_manage_users: true,
            can_view_sensitive_data: true,
            can_export_data: true,
            can_modify_settings: false,
            can_approve_actions: false,
            can_access_audit_logs: true,
            can_manage_feature_flags: false,
          },
        },
      ];
    }
  }

  async updateAdminRole(adminId: string, role: string, permissions: Partial<AdminPermissions>): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.put(
        `/api/admin/users/${adminId}`,
        { role, permissions }
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to update admin role');
    }
  }

  async deactivateAdmin(adminId: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.delete(
        `/api/admin/users/${adminId}`
      );
      return response.data.data!.success;
    } catch (error: unknown) {
      throw new Error(this.extractErrorMessage(error, 'Failed to deactivate admin'));
    }
  }

  async getPendingApprovals(): Promise<AdminActionApproval[]> {
    try {
      const response: AxiosResponse<APIResponse<AdminActionApproval[]>> = await this.api.get(
        '/api/admin/approvals'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return [
        {
          id: 'approval-1',
          action_type: 'feature_flag_change',
          requested_by: 'secondary-admin',
          requested_at: new Date().toISOString(),
          details: { feature: 'zkid_enabled', new_value: false },
          status: 'pending',
        },
      ];
    }
  }

  async approveAction(approvalId: string, approved: boolean, reason?: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.post(
        `/api/admin/approvals/${approvalId}`,
        { approved, reason }
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to process approval');
    }
  }

  // ============================================================================
  // SYSTEM MONITORING - ZKID LAYER
  // ============================================================================

  async getZKIDMetrics(): Promise<ZKIDLayerMetrics> {
    try {
      const response: AxiosResponse<APIResponse<ZKIDLayerMetrics>> = await this.api.get(
        '/api/admin/metrics/zkid'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return {
        enabled: true,
        endpoint_health: {
          mapping_creation: {
            requests: 1250,
            errors: 12,
            average_latency_ms: 45,
            success_rate: 99.04,
            last_updated: new Date().toISOString(),
          },
          email_retrieval: {
            requests: 890,
            errors: 8,
            average_latency_ms: 32,
            success_rate: 99.10,
            last_updated: new Date().toISOString(),
          },
          recovery_generation: {
            requests: 156,
            errors: 3,
            average_latency_ms: 78,
            success_rate: 98.08,
            last_updated: new Date().toISOString(),
          },
          recovery_validation: {
            requests: 89,
            errors: 2,
            average_latency_ms: 65,
            success_rate: 97.75,
            last_updated: new Date().toISOString(),
          },
          recovery_revocation: {
            requests: 23,
            errors: 1,
            average_latency_ms: 42,
            success_rate: 95.65,
            last_updated: new Date().toISOString(),
          },
        },
        recovery_operations: {
          total_generated: 156,
          total_used: 89,
          total_revoked: 23,
          failed_attempts: 5,
          recent_activity: [
            {
              id: 'activity-1',
              user_id: 'user-123',
              action: 'generated',
              timestamp: new Date().toISOString(),
              ip_address: '192.168.1.100',
              success: true,
            },
            {
              id: 'activity-2',
              user_id: 'user-456',
              action: 'used',
              timestamp: new Date(Date.now() - 300000).toISOString(),
              ip_address: '192.168.1.101',
              success: true,
            },
          ],
        },
        database_performance: {
          mapping_queries_per_second: 45.2,
          recovery_queries_per_second: 12.8,
          average_query_time_ms: 15.3,
          encryption_overhead_ms: 8.7,
        },
        security_events: {
          unauthorized_access_attempts: 3,
          failed_recovery_attempts: 5,
          encryption_errors: 1,
          audit_log_entries: 1247,
        },
        
        // Enhanced ZKID Monitoring
        uuid_mapping_creations: 1250,
        uuid_mapping_retrievals: 890,
        recovery_code_usage_count: 89,
        active_recovery_codes: 67,
        expired_recovery_codes: 23,
        revoked_recovery_codes: 12,
        mapping_creation_latency_ms: 45,
        mapping_retrieval_latency_ms: 32,
        recovery_code_generation_latency_ms: 78,
        failed_uuid_lookups: 5,
        side_channel_protection_status: 'active',
        rate_limiting_status: 'enforced',
        zero_knowledge_compliance: true,
        audit_log_entries: 1247,
      };
    }
  }

  // ============================================================================
  // SYSTEM MONITORING - PQC / ENCRYPTION LAYER
  // ============================================================================

  async getPQCEncryptionMetrics(): Promise<PQCEncryptionMetrics> {
    try {
      const response: AxiosResponse<APIResponse<PQCEncryptionMetrics>> = await this.api.get(
        '/api/admin/metrics/pqc'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return {
        key_management: {
          keys_generated: 1250,
          keys_rotated: 89,
          rotation_failures: 2,
          hsm_operations: 15600,
          hsm_errors: 0,
        },
        encryption_performance: {
          aes_256_gcm_operations: 8900,
          chacha20_operations: 3400,
          kyber_operations: 200,
          average_encryption_time_ms: 12.5,
          average_decryption_time_ms: 8.3,
          encryption_errors: 3,
          decryption_errors: 1,
        },
        algorithm_usage: {
          aes_256_gcm_percentage: 67.5,
          chacha20_percentage: 25.8,
          kyber_percentage: 1.5,
          hybrid_percentage: 5.2,
        },
        security_status: {
          hsm_available: true,
          key_store_encrypted: true,
          rotation_schedule_compliant: true,
          post_quantum_ready: true,
        },
        
        // Enhanced PQC Monitoring
        key_rotation_schedule: {
          next_rotation: new Date(Date.now() + 86400000).toISOString(), // 24 hours from now
          last_rotation: new Date(Date.now() - 86400000).toISOString(), // 24 hours ago
          rotation_interval_hours: 24,
          grace_period_hours: 2,
        },
        key_health_status: {
          active_keys: 5,
          expiring_keys: 1,
          expired_keys: 0,
          compromised_keys: 0,
          key_strength_score: 95,
        },
        aead_encryption_stats: {
          successful_encryptions: 12400,
          failed_encryptions: 3,
          tag_verification_errors: 1,
          nonce_reuse_attempts: 0,
          ciphertext_tampering_attempts: 0,
        },
        performance_metrics: {
          encryption_throughput_ops_per_sec: 1250,
          decryption_throughput_ops_per_sec: 1350,
          key_generation_time_ms: 45,
          key_rotation_time_ms: 120,
          hsm_latency_ms: 8,
        },
      };
    }
  }

  // ============================================================================
  // SYSTEM MONITORING - EMAIL DELIVERY & SYSTEM METRICS
  // ============================================================================

  async getEmailDeliveryMetrics(): Promise<EmailDeliveryMetrics> {
    try {
      const response: AxiosResponse<APIResponse<EmailDeliveryMetrics>> = await this.api.get(
        '/api/admin/metrics/email-delivery'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return {
        queue_status: {
          pending_emails: 45,
          processing_emails: 12,
          failed_emails: 3,
          queue_size_limit: 1000,
          queue_health_percentage: 94.2,
        },
        delivery_performance: {
          total_sent: 12500,
          successful_deliveries: 12450,
          failed_deliveries: 50,
          delivery_success_rate: 99.6,
          average_processing_time_ms: 1250,
          retry_attempts: 89,
        },
        storage_metrics: {
          total_storage_used_gb: 45.7,
          storage_limit_gb: 100,
          storage_usage_percentage: 45.7,
          encrypted_blobs: 12500,
          average_blob_size_kb: 3.7,
          storage_errors: 0,
        },
        system_resources: {
          cpu_usage_percentage: 23.5,
          memory_usage_percentage: 67.8,
          disk_usage_percentage: 45.7,
          network_bandwidth_mbps: 125.3,
          active_connections: 89,
          database_connections: 12,
        },
        
        // Enhanced Email System Monitoring
        email_queue_monitoring: {
          queue_length: 45,
          failed_delivery_attempts: 3,
          storage_usage_gb: 45.7,
          storage_limit_gb: 100,
          queue_processing_rate_per_min: 125,
          average_queue_wait_time_ms: 850,
        },
        read_once_enforcement: {
          read_once_violations: 0,
          self_destruct_triggers: 12,
          burn_after_read_count: 8,
          retention_policy_enforcement: 156,
          email_expiration_count: 23,
        },
        delivery_analytics: {
          delivery_success_rate_percent: 99.6,
          average_delivery_time_ms: 1250,
          failed_delivery_reasons: {
            'invalid_recipient': 25,
            'network_timeout': 15,
            'quota_exceeded': 8,
            'spam_filter': 2,
          },
          retry_success_rate: 85.2,
          dead_letter_queue_size: 5,
        },
        storage_performance: {
          read_operations_per_sec: 450,
          write_operations_per_sec: 380,
          storage_latency_ms: 12,
          encryption_overhead_ms: 3,
          compression_ratio: 0.75,
        },
      };
    }
  }

  // ============================================================================
  // SYSTEM MONITORING - SECURITY & COMPLIANCE
  // ============================================================================

  async getSecurityComplianceMetrics(): Promise<SecurityComplianceMetrics> {
    try {
      const response: AxiosResponse<APIResponse<SecurityComplianceMetrics>> = await this.api.get(
        '/api/admin/metrics/security-compliance'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return {
        authentication_security: {
          failed_login_attempts: 23,
          successful_logins: 456,
          brute_force_attempts: 5,
          account_lockouts: 2,
          password_resets: 12,
        },
        access_control: {
          rbac_violations: 1,
          unauthorized_access_attempts: 8,
          privilege_escalation_attempts: 0,
          session_timeouts: 45,
          concurrent_sessions: 89,
        },
        audit_compliance: {
          audit_log_entries: 1247,
          compliance_events: 23,
          gdpr_requests: 2,
          data_retention_events: 5,
          privacy_violations: 0,
        },
        feature_flags: {
          zkid_enabled: true,
          pqc_enabled: true,
          enterprise_enabled: true,
          mfa_enabled: true,
          geo_restrictions_enabled: true,
          recent_rollbacks: 0,
        },
        geolocation_compliance: {
          geo_restriction_violations: 3,
          vpn_detections: 12,
          suspicious_locations: 2,
          compliance_checks: 1250,
        },
      };
    }
  }

  // ============================================================================
  // SYSTEM MONITORING - PERFORMANCE & OPERATIONAL METRICS
  // ============================================================================

  async getPerformanceOperationalMetrics(): Promise<PerformanceOperationalMetrics> {
    try {
      const response: AxiosResponse<APIResponse<PerformanceOperationalMetrics>> = await this.api.get(
        '/api/admin/metrics/performance-operational'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return {
        api_performance: {
          total_requests: 15600,
          successful_requests: 15520,
          failed_requests: 80,
          average_response_time_ms: 125,
          p95_response_time_ms: 450,
          p99_response_time_ms: 890,
          requests_per_second: 45.2,
        },
        endpoint_metrics: {
          '/api/auth/login': {
            requests: 456,
            errors: 23,
            average_latency_ms: 89,
            success_rate: 94.96,
          },
          '/api/email/send': {
            requests: 1250,
            errors: 12,
            average_latency_ms: 234,
            success_rate: 99.04,
          },
          '/api/zkid/mapping': {
            requests: 890,
            errors: 8,
            average_latency_ms: 67,
            success_rate: 99.10,
          },
        },
        error_tracking: {
          total_errors: 80,
          error_rate_percentage: 0.51,
          critical_errors: 3,
          error_trends: [
            {
              timestamp: new Date().toISOString(),
              error_count: 12,
              error_rate: 0.08,
            },
          ],
          recent_errors: [
            {
              id: 'error-1',
              timestamp: new Date().toISOString(),
              error_type: 'database_connection',
              message: 'Database connection timeout',
              severity: 'high',
            },
          ],
        },
        load_testing: {
          concurrent_users: 150,
          max_concurrent_users: 500,
          throughput_requests_per_second: 45.2,
        },
        
        // Enhanced Performance Metrics
        real_time_api_latency: {
          '/api/auth/login': {
            current_latency_ms: 89,
            p50_latency_ms: 75,
            p95_latency_ms: 125,
            p99_latency_ms: 180,
            requests_per_minute: 456,
            error_rate_percent: 5.04,
          },
          '/api/email/send': {
            current_latency_ms: 234,
            p50_latency_ms: 200,
            p95_latency_ms: 350,
            p99_latency_ms: 450,
            requests_per_minute: 1250,
            error_rate_percent: 0.96,
          },
          '/api/zkid/mapping': {
            current_latency_ms: 67,
            p50_latency_ms: 55,
            p95_latency_ms: 95,
            p99_latency_ms: 120,
            requests_per_minute: 890,
            error_rate_percent: 0.90,
          },
        },
        session_metrics: {
          concurrent_user_sessions: 89,
          active_admin_sessions: 3,
          session_creation_rate_per_min: 12,
          session_timeout_count: 5,
          average_session_duration_min: 45,
        },
        database_performance: {
          encrypted_mappings_queries_per_sec: 45.2,
          average_query_time_ms: 15.3,
          slow_queries_count: 3,
          connection_pool_usage_percent: 65,
          database_latency_ms: 8.7,
          encryption_overhead_ms: 2.1,
        },
        system_health: {
          cpu_usage_percent: 23.5,
          memory_usage_percent: 67.8,
          disk_io_operations_per_sec: 1250,
          network_throughput_mbps: 125.3,
          active_goroutines: 156,
          gc_pause_time_ms: 2.5,
          uptime_percentage: 99.95,
          last_restart: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
          health_check_status: 'healthy',
          dependency_status: {
            database: 'healthy',
            storage: 'healthy',
            encryption: 'healthy',
          },
        },
        load_test_results: [
          {
            id: 'load-test-1',
            timestamp: new Date().toISOString(),
            concurrent_users: 100,
            requests_per_second: 45.2,
            average_response_time_ms: 125,
            error_rate: 0.51,
            success: true,
          },
        ],
      };
    }
  }

  // ============================================================================
  // ALERTS & NOTIFICATIONS
  // ============================================================================

  async getAlerts(): Promise<Alert[]> {
    try {
      const response: AxiosResponse<APIResponse<Alert[]>> = await this.api.get(
        '/api/admin/alerts'
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return [
        {
          id: 'alert-1',
          severity: 'medium',
          category: 'performance',
          title: 'High Response Time',
          description: 'API response time exceeded 500ms threshold',
          timestamp: new Date().toISOString(),
          acknowledged: false,
          resolved: false,
          alert_type: 'threshold_exceeded',
          notification_channels: ['email', 'dashboard'],
          escalation_level: 1,
          source_component: 'api',
          threshold_value: 500,
          current_value: 650,
        },
        {
          id: 'alert-2',
          severity: 'low',
          category: 'security',
          title: 'Failed Login Attempts',
          description: 'Multiple failed login attempts detected',
          timestamp: new Date(Date.now() - 300000).toISOString(),
          acknowledged: true,
          acknowledged_by: 'secondary-admin',
          acknowledged_at: new Date(Date.now() - 240000).toISOString(),
          resolved: false,
          alert_type: 'anomaly_detected',
          notification_channels: ['email', 'dashboard', 'slack'],
          escalation_level: 2,
          source_component: 'authentication',
          threshold_value: 5,
          current_value: 8,
        },
      ];
    }
  }

  async acknowledgeAlert(alertId: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.post(
        `/api/admin/alerts/${alertId}/acknowledge`
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to acknowledge alert');
    }
  }

  async resolveAlert(alertId: string): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.post(
        `/api/admin/alerts/${alertId}/resolve`
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to resolve alert');
    }
  }

  // ============================================================================
  // AUDIT LOGGING
  // ============================================================================

  async getAuditLogs(filter?: AuditLogFilter): Promise<AuditLogEntry[]> {
    try {
      const response: AxiosResponse<APIResponse<AuditLogEntry[]>> = await this.api.get(
        '/api/admin/audit-logs',
        { params: filter }
      );
      return response.data.data!;
    } catch {
      // Return mock data if API not available
      return [
        {
          id: 'log-1',
          timestamp: new Date().toISOString(),
          user_id: 'admin-1',
          action: 'login',
          resource: 'dashboard',
          details: { ip_address: '192.168.1.100' },
          ip_address: '192.168.1.100',
          user_agent: 'Mozilla/5.0...',
          success: true,
        },
        {
          id: 'log-2',
          timestamp: new Date(Date.now() - 300000).toISOString(),
          user_id: 'admin-2',
          action: 'view_metrics',
          resource: 'zkid_layer',
          details: { panel: 'ZKIDLayerPanel' },
          ip_address: '192.168.1.101',
          user_agent: 'Mozilla/5.0...',
          success: true,
        },
      ];
    }
  }

  // ============================================================================
  // DASHBOARD CONFIGURATION
  // ============================================================================

  async getDashboardConfig(): Promise<DashboardConfig> {
    try {
      const response: AxiosResponse<APIResponse<DashboardConfig>> = await this.api.get(
        '/api/admin/dashboard/config'
      );
      return response.data.data!;
    } catch {
      // Return default config if API not available
      return {
        refresh_interval_seconds: 30,
        auto_refresh_enabled: true,
        theme: 'light',
        layout: 'comfortable',
        panels: {
          zkid_layer: { enabled: true, position: 1, size: 'medium' },
          pqc_encryption: { enabled: true, position: 2, size: 'medium' },
          email_delivery: { enabled: true, position: 3, size: 'medium' },
          security_compliance: { enabled: true, position: 4, size: 'medium' },
          performance_operational: { enabled: true, position: 5, size: 'medium' },
          alerts: { enabled: true, position: 6, size: 'medium' },
          audit_logs: { enabled: true, position: 7, size: 'large' },
        },
        alerts: {
          enabled: true,
          sound_enabled: false,
          desktop_notifications: true,
          email_notifications: false,
        },
      };
    }
  }

  async updateDashboardConfig(config: Partial<DashboardConfig>): Promise<boolean> {
    try {
      const response: AxiosResponse<APIResponse<{ success: boolean }>> = await this.api.put(
        '/api/admin/dashboard/config',
        config
      );
      return response.data.data!.success;
    } catch {
      throw new Error('Failed to update dashboard config');
    }
  }

  // ============================================================================
  // REAL-TIME UPDATES
  // ============================================================================

  startRealTimeUpdates(callback: (update: RealTimeUpdate) => void): void {
    // In a real implementation, this would establish WebSocket connection
    // For now, simulate real-time updates with polling
    this.refreshInterval = setInterval(async () => {
      try {
        // Simulate real-time updates
        const update: RealTimeUpdate = {
          id: `update-${Date.now()}`,
          type: 'metric_update',
          component: 'api',
          data: { timestamp: new Date().toISOString() },
          timestamp: new Date().toISOString(),
          severity: 'info',
        };
        callback(update);
      } catch (error) {
        log.error('Real-time update error:', error, 'enterpriseDashboardService');
      }
    }, 30000); // 30 seconds
  }

  stopRealTimeUpdates(): void {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
      this.refreshInterval = null;
    }
  }

  // ============================================================================
  // UTILITY METHODS
  // ============================================================================

  hasPermission(permission: keyof AdminPermissions): boolean {
    if (!this.currentUser) return false;
    return this.currentUser.permissions[permission] || false;
  }

  canManageAdmins(): boolean {
    return this.hasPermission('can_manage_admins');
  }

  canViewSensitiveData(): boolean {
    return this.hasPermission('can_view_sensitive_data');
  }

  canExportData(): boolean {
    return this.hasPermission('can_export_data');
  }

  isPrimaryAdmin(): boolean {
    return this.currentUser?.role === 'primary_admin';
  }

  isReadOnlyAdmin(): boolean {
    return this.currentUser?.role === 'read_only_admin';
  }

  // ============================================================================
  // ANALYTICS & TREND ANALYSIS - ITERATION 6
  // ============================================================================

  async getAnalyticsDashboard(filter?: AnalyticsFilter): Promise<AnalyticsDashboard> {
    try {
      const response: AxiosResponse<APIResponse<AnalyticsDashboard>> = await this.api.get(
        '/api/admin/analytics/dashboard',
        { params: filter }
      );
      return response.data.data!;
    } catch {
      // Return mock data for development
      return this.getMockAnalyticsDashboard();
    }
  }

  async getEmailUsageAnalytics(timeRange?: AnalyticsTimeRange): Promise<EmailUsageAnalytics> {
    try {
      const response: AxiosResponse<APIResponse<EmailUsageAnalytics>> = await this.api.get(
        '/api/admin/analytics/email-usage',
        { params: timeRange }
      );
      return response.data.data!;
    } catch {
      return this.getMockEmailUsageAnalytics();
    }
  }

  async getSecurityEventAnalytics(timeRange?: AnalyticsTimeRange): Promise<SecurityEventAnalytics> {
    try {
      const response: AxiosResponse<APIResponse<SecurityEventAnalytics>> = await this.api.get(
        '/api/admin/analytics/security-events',
        { params: timeRange }
      );
      return response.data.data!;
    } catch {
      return this.getMockSecurityEventAnalytics();
    }
  }

  async getZKIDPQCAnalytics(timeRange?: AnalyticsTimeRange): Promise<ZKIDPQCAnalytics> {
    try {
      const response: AxiosResponse<APIResponse<ZKIDPQCAnalytics>> = await this.api.get(
        '/api/admin/analytics/zkid-pqc',
        { params: timeRange }
      );
      return response.data.data!;
    } catch {
      return this.getMockZKIDPQCAnalytics();
    }
  }

  async exportAnalytics(request: AnalyticsExportRequest): Promise<AnalyticsExportResponse> {
    try {
      const response: AxiosResponse<APIResponse<AnalyticsExportResponse>> = await this.api.post(
        '/api/admin/analytics/export',
        request
      );
      return response.data.data!;
    } catch {
      throw new Error('Export failed');
    }
  }

  // ============================================================================
  // THREAT AWARENESS MODULE - ITERATION 6
  // ============================================================================

  async getThreatIntelligence(): Promise<ThreatIntelligence> {
    try {
      const response: AxiosResponse<APIResponse<ThreatIntelligence>> = await this.api.get(
        '/api/admin/threat/intelligence'
      );
      return response.data.data!;
    } catch {
      return this.getMockThreatIntelligence();
    }
  }

  async getThreatAlerts(): Promise<ThreatAlert[]> {
    try {
      const response: AxiosResponse<APIResponse<ThreatAlert[]>> = await this.api.get(
        '/api/admin/threat/alerts'
      );
      return response.data.data!;
    } catch {
      return this.getMockThreatAlerts();
    }
  }

  async getThreatFeeds(): Promise<ThreatFeed[]> {
    try {
      const response: AxiosResponse<APIResponse<ThreatFeed[]>> = await this.api.get(
        '/api/admin/threat/feeds'
      );
      return response.data.data!;
    } catch {
      return this.getMockThreatFeeds();
    }
  }

  async getThreatRules(): Promise<ThreatRule[]> {
    try {
      const response: AxiosResponse<APIResponse<ThreatRule[]>> = await this.api.get(
        '/api/admin/threat/rules'
      );
      return response.data.data!;
    } catch {
      return this.getMockThreatRules();
    }
  }

  async getThreatAwarenessConfig(): Promise<ThreatAwarenessConfig> {
    try {
      const response: AxiosResponse<APIResponse<ThreatAwarenessConfig>> = await this.api.get(
        '/api/admin/threat/config'
      );
      return response.data.data!;
    } catch {
      return this.getMockThreatAwarenessConfig();
    }
  }

  async updateThreatAlert(alertId: string, updates: Partial<ThreatAlert>): Promise<ThreatAlert> {
    try {
      const response: AxiosResponse<APIResponse<ThreatAlert>> = await this.api.put(
        `/api/admin/threat/alerts/${alertId}`,
        updates
      );
      return response.data.data!;
    } catch (error: unknown) {
      throw new Error(this.extractErrorMessage(error, 'Failed to update threat alert'));
    }
  }

  async createThreatRule(rule: Omit<ThreatRule, 'id' | 'created_at' | 'updated_at'>): Promise<ThreatRule> {
    try {
      const response: AxiosResponse<APIResponse<ThreatRule>> = await this.api.post(
        '/api/admin/threat/rules',
        rule
      );
      return response.data.data!;
    } catch (error: unknown) {
      throw new Error(this.extractErrorMessage(error, 'Failed to create threat rule'));
    }
  }

  async updateThreatAwarenessConfig(config: Partial<ThreatAwarenessConfig>): Promise<ThreatAwarenessConfig> {
    try {
      const response: AxiosResponse<APIResponse<ThreatAwarenessConfig>> = await this.api.put(
        '/api/admin/threat/config',
        config
      );
      return response.data.data!;
    } catch (error: unknown) {
      throw new Error(this.extractErrorMessage(error, 'Failed to update threat awareness config'));
    }
  }

  // ============================================================================
  // MOCK DATA GENERATORS - ITERATION 6
  // ============================================================================

  private getMockAnalyticsDashboard(): AnalyticsDashboard {
    const now = new Date();
    const timeRange: AnalyticsTimeRange = {
      start: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString(),
      end: now.toISOString(),
      granularity: 'day'
    };

    return {
      time_range: timeRange,
      email_usage: this.getMockEmailUsageAnalytics(),
      security_events: this.getMockSecurityEventAnalytics(),
      zkid_pqc: this.getMockZKIDPQCAnalytics(),
      threat_intelligence: this.getMockThreatIntelligence(),
      last_updated: now.toISOString(),
      refresh_interval: 300
    };
  }

  private getMockEmailUsageAnalytics(): EmailUsageAnalytics {
    const trendData = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        timestamp: date.toISOString(),
        emails_sent: Math.floor(Math.random() * 1000) + 500,
        emails_received: Math.floor(Math.random() * 800) + 400,
        storage_used: Math.floor(Math.random() * 1000000) + 500000,
        delivery_success_rate: Math.random() * 0.1 + 0.95
      };
    });

    return {
      total_emails_sent: 12500,
      total_emails_received: 9800,
      emails_by_type: {
        secure: 8500,
        read_once: 1200,
        self_destruct: 800,
        burn_after_read: 600,
        password_protected: 1400
      },
      emails_by_status: {
        delivered: 11800,
        failed: 400,
        pending: 200,
        expired: 50,
        destroyed: 50
      },
      top_senders: [
        { email: 'admin@company.com', count: 1250, percentage: 10.0 },
        { email: 'support@company.com', count: 980, percentage: 7.8 },
        { email: 'alerts@company.com', count: 750, percentage: 6.0 }
      ],
      top_recipients: [
        { email: 'user1@client.com', count: 890, percentage: 9.1 },
        { email: 'user2@client.com', count: 720, percentage: 7.3 },
        { email: 'user3@client.com', count: 650, percentage: 6.6 }
      ],
      delivery_times: {
        average_seconds: 2.5,
        p50_seconds: 1.8,
        p95_seconds: 4.2,
        p99_seconds: 8.1
      },
      storage_usage: {
        total_bytes: 5242880000, // 5GB
        encrypted_bytes: 5242880000,
        compression_ratio: 0.75,
        growth_rate_percent: 12.5
      },
      trend_data: trendData
    };
  }

  private getMockSecurityEventAnalytics(): SecurityEventAnalytics {
    const trendData = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        timestamp: date.toISOString(),
        total_events: Math.floor(Math.random() * 50) + 10,
        critical_events: Math.floor(Math.random() * 5) + 1,
        high_events: Math.floor(Math.random() * 10) + 2,
        medium_events: Math.floor(Math.random() * 15) + 5,
        low_events: Math.floor(Math.random() * 20) + 10
      };
    });

    return {
      total_security_events: 245,
      events_by_severity: {
        critical: 8,
        high: 25,
        medium: 67,
        low: 145
      },
      events_by_type: {
        failed_login_attempts: 89,
        brute_force_attempts: 23,
        unauthorized_access: 12,
        suspicious_activity: 45,
        encryption_failures: 8,
        key_rotation_failures: 3,
        audit_log_tampering: 2,
        compliance_violations: 63
      },
      events_by_component: {
        zkid_layer: 15,
        pqc_encryption: 12,
        email_pipeline: 45,
        admin_dashboard: 23,
        authentication: 89,
        realtime_monitoring: 61
      },
      threat_indicators: {
        ioc_count: 34,
        suspicious_ips: 12,
        anomalous_patterns: 8,
        potential_attacks: 5
      },
      trend_data: trendData
    };
  }

  private getMockZKIDPQCAnalytics(): ZKIDPQCAnalytics {
    const trendData = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        timestamp: date.toISOString(),
        zkid_operations: Math.floor(Math.random() * 500) + 200,
        pqc_operations: Math.floor(Math.random() * 800) + 400,
        avg_zkid_latency: Math.random() * 10 + 5,
        avg_pqc_latency: Math.random() * 20 + 10,
        security_score: Math.random() * 20 + 80
      };
    });

    return {
      zkid_operations: {
        total_mappings_created: 12500,
        total_mappings_retrieved: 11800,
        total_recovery_codes_generated: 1250,
        total_recovery_codes_used: 89,
        total_recovery_codes_revoked: 12,
        active_mappings: 11200,
        expired_mappings: 800,
        revoked_mappings: 500
      },
      pqc_operations: {
        total_keys_generated: 1250,
        total_keys_rotated: 89,
        total_encryptions: 12500,
        total_decryptions: 11800,
        hybrid_encryptions: 11200,
        classical_encryptions: 1300,
        encryption_failures: 8,
        decryption_failures: 12
      },
      performance_metrics: {
        zkid_creation_latency_ms: 8.5,
        zkid_retrieval_latency_ms: 5.2,
        pqc_encryption_latency_ms: 15.8,
        pqc_decryption_latency_ms: 12.3,
        key_rotation_duration_ms: 2500
      },
      security_metrics: {
        recovery_code_usage_rate: 0.071,
        key_rotation_success_rate: 0.987,
        encryption_success_rate: 0.999,
        decryption_success_rate: 0.998,
        zero_knowledge_compliance: 1.0
      },
      trend_data: trendData
    };
  }

  private getMockThreatIntelligence(): ThreatIntelligence {
    const threatTrends = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        timestamp: date.toISOString(),
        threat_level: ['low', 'medium', 'high', 'critical'][Math.floor(Math.random() * 4)] as 'low' | 'medium' | 'high' | 'critical',
        threat_score: Math.floor(Math.random() * 40) + 20,
        active_threats: Math.floor(Math.random() * 10) + 2,
        emerging_threats: Math.floor(Math.random() * 5) + 1
      };
    });

    return {
      threat_level: 'medium',
      threat_score: 45,
      active_threats: 8,
      emerging_threats: 3,
      threat_categories: {
        email_attacks: 12,
        cryptographic_attacks: 5,
        infrastructure_attacks: 8,
        social_engineering: 15,
        insider_threats: 3
      },
      threat_indicators: [
        {
          id: 'ti-001',
          type: 'ip',
          value: '192.168.1.100',
          confidence: 0.85,
          first_seen: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
          last_seen: new Date().toISOString(),
          threat_level: 'high',
          description: 'Suspicious login attempts from unknown IP'
        },
        {
          id: 'ti-002',
          type: 'domain',
          value: 'malicious-domain.com',
          confidence: 0.92,
          first_seen: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
          last_seen: new Date().toISOString(),
          threat_level: 'critical',
          description: 'Known phishing domain attempting email delivery'
        }
      ],
      threat_trends: threatTrends,
      recommendations: [
        {
          id: 'rec-001',
          priority: 'high',
          category: 'authentication',
          title: 'Implement additional MFA requirements',
          description: 'Recent suspicious login patterns suggest need for enhanced authentication',
          action_required: true,
          estimated_effort: '2-3 hours'
        },
        {
          id: 'rec-002',
          priority: 'medium',
          category: 'monitoring',
          title: 'Review and update threat detection rules',
          description: 'Current rules may not be catching all emerging threat patterns',
          action_required: false,
          estimated_effort: '4-6 hours'
        }
      ]
    };
  }

  private getMockThreatAlerts(): ThreatAlert[] {
    return [
      {
        id: 'ta-001',
        timestamp: new Date().toISOString(),
        threat_level: 'high',
        category: 'email',
        title: 'Suspicious Email Delivery Pattern Detected',
        description: 'Multiple failed delivery attempts from suspicious IP addresses',
        indicators: ['192.168.1.100', 'malicious-domain.com'],
        affected_components: ['email_pipeline', 'authentication'],
        recommended_actions: ['Block suspicious IPs', 'Review email filtering rules'],
        status: 'active',
        assigned_to: 'admin@company.com',
        notes: 'Investigating source of suspicious activity',
        created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
        updated_at: new Date().toISOString()
      },
      {
        id: 'ta-002',
        timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        threat_level: 'medium',
        category: 'cryptographic',
        title: 'PQC Key Rotation Failure',
        description: 'Failed to rotate PQC keys during scheduled maintenance',
        indicators: ['key_rotation_failure', 'pqc_service_error'],
        affected_components: ['pqc_encryption'],
        recommended_actions: ['Check PQC service logs', 'Verify key backup integrity'],
        status: 'resolved',
        assigned_to: 'admin@company.com',
        notes: 'Issue resolved after service restart',
        created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        updated_at: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString()
      }
    ];
  }

  private getMockThreatFeeds(): ThreatFeed[] {
    return [
      {
        id: 'tf-001',
        name: 'CISCO Talos Intelligence',
        description: 'Real-time threat intelligence from Cisco Talos',
        url: 'https://talosintelligence.com/feeds/ip-filter',
        format: 'json',
        update_frequency: 'hourly',
        last_update: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
        next_update: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
        status: 'active',
        threat_indicators_count: 15420,
        confidence_score: 0.95
      },
      {
        id: 'tf-002',
        name: 'AbuseIPDB',
        description: 'IP reputation database for threat detection',
        url: 'https://api.abuseipdb.com/api/v2/blacklist',
        format: 'json',
        update_frequency: 'daily',
        last_update: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(),
        next_update: new Date(Date.now() + 18 * 60 * 60 * 1000).toISOString(),
        status: 'active',
        threat_indicators_count: 8920,
        confidence_score: 0.88
      }
    ];
  }

  private getMockThreatRules(): ThreatRule[] {
    return [
      {
        id: 'tr-001',
        name: 'High Volume Login Attempts',
        description: 'Detect multiple failed login attempts from same IP',
        enabled: true,
        priority: 'high',
        conditions: [
          { field: 'event_type', operator: 'equals', value: 'failed_login' },
          { field: 'attempts', operator: 'greater_than', value: 5 },
          { field: 'time_window_minutes', operator: 'less_than', value: 10 }
        ],
        actions: [
          { type: 'alert', parameters: { severity: 'high' } },
          { type: 'block', parameters: { duration_minutes: 30 } }
        ],
        created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
      },
      {
        id: 'tr-002',
        name: 'Suspicious Email Patterns',
        description: 'Detect potential phishing or malicious email patterns',
        enabled: true,
        priority: 'medium',
        conditions: [
          { field: 'email_subject', operator: 'contains', value: 'urgent' },
          { field: 'sender_domain', operator: 'regex', value: '.*\\.suspicious\\.com$' }
        ],
        actions: [
          { type: 'alert', parameters: { severity: 'medium' } },
          { type: 'log', parameters: { level: 'warning' } }
        ],
        created_at: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
        updated_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString()
      }
    ];
  }

  private getMockThreatAwarenessConfig(): ThreatAwarenessConfig {
    return {
      enabled: true,
      threat_feeds: this.getMockThreatFeeds(),
      threat_rules: this.getMockThreatRules(),
      alert_thresholds: {
        low_threshold: 10,
        medium_threshold: 25,
        high_threshold: 50,
        critical_threshold: 75
      },
      auto_blocking: true,
      notification_channels: ['email', 'webhook', 'dashboard'],
      update_interval_minutes: 15
    };
  }
}
