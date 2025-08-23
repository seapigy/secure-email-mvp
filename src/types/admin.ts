// Enhanced Multi-Admin Enterprise Dashboard Types
// Comprehensive monitoring for the entire Secure Email MVP system with multi-admin support

// ============================================================================
// AUTHENTICATION & AUTHORIZATION - MULTI-ADMIN
// ============================================================================

export interface AdminAuthConfig {
  enabled: boolean;
  require_mfa: boolean;
  mfa_type: 'TOTP' | 'HARDWARE_TOKEN';
  session_timeout_minutes: number;
  max_failed_attempts: number;
  lockout_duration_minutes: number;
  invitation_key_expiry_hours: number;
  max_secondary_admins: number;
  require_primary_approval: boolean;
}

export interface AdminUser {
  id: string;
  username: string;
  role: 'primary_admin' | 'full_admin' | 'read_only_admin';
  mfa_enabled: boolean;
  mfa_type?: 'TOTP' | 'HARDWARE_TOKEN';
  last_login?: string;
  failed_attempts: number;
  locked_until?: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
  is_active: boolean;
  permissions: AdminPermissions;
}

export interface AdminPermissions {
  can_manage_system: boolean;
  can_manage_admins: boolean;
  can_manage_organizations: boolean;
  can_manage_users: boolean;
  can_view_sensitive_data: boolean;
  can_export_data: boolean;
  can_modify_settings: boolean;
  can_approve_actions: boolean;
  can_access_audit_logs: boolean;
  can_manage_feature_flags: boolean;
}

export interface AdminLoginRequest {
  username: string;
  password: string;
  mfa_code?: string;
  invitation_key?: string;
}

export interface AdminLoginResponse {
  success: boolean;
  token?: string;
  requires_mfa?: boolean;
  mfa_type?: 'TOTP' | 'HARDWARE_TOKEN';
  message?: string;
  error_code?: string;
  user?: AdminUser;
}

export interface MFASetupRequest {
  mfa_type: 'TOTP' | 'HARDWARE_TOKEN';
  totp_secret?: string;
  hardware_token_id?: string;
}

export interface MFASetupResponse {
  success: boolean;
  qr_code_url?: string;
  backup_codes?: string[];
  message?: string;
}

export interface InvitationKey {
  key: string;
  created_by: string;
  expires_at: string;
  scope: 'dashboard_access' | 'full_access' | 'read_only_access';
  used: boolean;
  used_at?: string;
  used_by?: string;
  max_uses?: number;
  current_uses: number;
}

export interface AdminInvitationRequest {
  email: string;
  role: 'full_admin' | 'read_only_admin';
  permissions?: Partial<AdminPermissions>;
  expires_in_hours?: number;
  max_uses?: number;
  message?: string;
}

export interface AdminInvitationResponse {
  success: boolean;
  invitation_key: string;
  expires_at: string;
  message?: string;
}

export interface AdminActionApproval {
  id: string;
  action_type: string;
  requested_by: string;
  requested_at: string;
  details: Record<string, unknown>;
  status: 'pending' | 'approved' | 'rejected';
  approved_by?: string;
  approved_at?: string;
  rejection_reason?: string;
}

// ============================================================================
// SYSTEM MONITORING - ZKID LAYER
// ============================================================================

export interface ZKIDLayerMetrics {
  enabled: boolean;
  endpoint_health: {
    mapping_creation: EndpointMetric;
    email_retrieval: EndpointMetric;
    recovery_generation: EndpointMetric;
    recovery_validation: EndpointMetric;
    recovery_revocation: EndpointMetric;
  };
  recovery_operations: {
    total_generated: number;
    total_used: number;
    total_revoked: number;
    failed_attempts: number;
    recent_activity: RecoveryActivity[];
  };
  database_performance: {
    mapping_queries_per_second: number;
    recovery_queries_per_second: number;
    average_query_time_ms: number;
    encryption_overhead_ms: number;
  };
  security_events: {
    unauthorized_access_attempts: number;
    failed_recovery_attempts: number;
    encryption_errors: number;
    audit_log_entries: number;
  };
  
  // Enhanced ZKID Monitoring
  uuid_mapping_creations: number;
  uuid_mapping_retrievals: number;
  recovery_code_usage_count: number;
  active_recovery_codes: number;
  expired_recovery_codes: number;
  revoked_recovery_codes: number;
  mapping_creation_latency_ms: number;
  mapping_retrieval_latency_ms: number;
  recovery_code_generation_latency_ms: number;
  failed_uuid_lookups: number;
  side_channel_protection_status: 'active' | 'inactive';
  rate_limiting_status: 'enforced' | 'bypassed';
  zero_knowledge_compliance: boolean;
  audit_log_entries: number;
}

// ============================================================================
// SYSTEM MONITORING - PQC / ENCRYPTION LAYER
// ============================================================================

export interface PQCEncryptionMetrics {
  key_management: {
    keys_generated: number;
    keys_rotated: number;
    rotation_failures: number;
    hsm_operations: number;
    hsm_errors: number;
  };
  encryption_performance: {
    aes_256_gcm_operations: number;
    chacha20_operations: number;
    kyber_operations: number;
    average_encryption_time_ms: number;
    average_decryption_time_ms: number;
    encryption_errors: number;
    decryption_errors: number;
  };
  algorithm_usage: {
    aes_256_gcm_percentage: number;
    chacha20_percentage: number;
    kyber_percentage: number;
    hybrid_percentage: number;
  };
  security_status: {
    hsm_available: boolean;
    key_store_encrypted: boolean;
    rotation_schedule_compliant: boolean;
    post_quantum_ready: boolean;
  };
  
  // Enhanced PQC Monitoring
  key_rotation_schedule: {
    next_rotation: string;
    last_rotation: string;
    rotation_interval_hours: number;
    grace_period_hours: number;
  };
  key_health_status: {
    active_keys: number;
    expiring_keys: number;
    expired_keys: number;
    compromised_keys: number;
    key_strength_score: number;
  };
  aead_encryption_stats: {
    successful_encryptions: number;
    failed_encryptions: number;
    tag_verification_errors: number;
    nonce_reuse_attempts: number;
    ciphertext_tampering_attempts: number;
  };
  performance_metrics: {
    encryption_throughput_ops_per_sec: number;
    decryption_throughput_ops_per_sec: number;
    key_generation_time_ms: number;
    key_rotation_time_ms: number;
    hsm_latency_ms: number;
  };
}

// ============================================================================
// SYSTEM MONITORING - EMAIL DELIVERY & SYSTEM METRICS
// ============================================================================

export interface EmailDeliveryMetrics {
  queue_status: {
    pending_emails: number;
    processing_emails: number;
    failed_emails: number;
    queue_size_limit: number;
    queue_health_percentage: number;
  };
  delivery_performance: {
    total_sent: number;
    successful_deliveries: number;
    failed_deliveries: number;
    delivery_success_rate: number;
    average_processing_time_ms: number;
    retry_attempts: number;
  };
  storage_metrics: {
    total_storage_used_gb: number;
    storage_limit_gb: number;
    storage_usage_percentage: number;
    encrypted_blobs: number;
    average_blob_size_kb: number;
    storage_errors: number;
  };
  system_resources: {
    cpu_usage_percentage: number;
    memory_usage_percentage: number;
    disk_usage_percentage: number;
    network_bandwidth_mbps: number;
    active_connections: number;
    database_connections: number;
  };
  
  // Enhanced Email System Monitoring
  email_queue_monitoring: {
    queue_length: number;
    failed_delivery_attempts: number;
    storage_usage_gb: number;
    storage_limit_gb: number;
    queue_processing_rate_per_min: number;
    average_queue_wait_time_ms: number;
  };
  read_once_enforcement: {
    read_once_violations: number;
    self_destruct_triggers: number;
    burn_after_read_count: number;
    retention_policy_enforcement: number;
    email_expiration_count: number;
  };
  delivery_analytics: {
    delivery_success_rate_percent: number;
    average_delivery_time_ms: number;
    failed_delivery_reasons: Record<string, number>;
    retry_success_rate: number;
    dead_letter_queue_size: number;
  };
  storage_performance: {
    read_operations_per_sec: number;
    write_operations_per_sec: number;
    storage_latency_ms: number;
    encryption_overhead_ms: number;
    compression_ratio: number;
  };
}

// ============================================================================
// SYSTEM MONITORING - SECURITY & COMPLIANCE
// ============================================================================

export interface SecurityComplianceMetrics {
  authentication_security: {
    failed_login_attempts: number;
    successful_logins: number;
    brute_force_attempts: number;
    account_lockouts: number;
    password_resets: number;
  };
  access_control: {
    rbac_violations: number;
    unauthorized_access_attempts: number;
    privilege_escalation_attempts: number;
    session_timeouts: number;
    concurrent_sessions: number;
  };
  audit_compliance: {
    audit_log_entries: number;
    compliance_events: number;
    gdpr_requests: number;
    data_retention_events: number;
    privacy_violations: number;
  };
  feature_flags: {
    zkid_enabled: boolean;
    pqc_enabled: boolean;
    enterprise_enabled: boolean;
    mfa_enabled: boolean;
    geo_restrictions_enabled: boolean;
    recent_rollbacks: number;
  };
  geolocation_compliance: {
    geo_restriction_violations: number;
    vpn_detections: number;
    suspicious_locations: number;
    compliance_checks: number;
  };
}

// ============================================================================
// SYSTEM MONITORING - PERFORMANCE & OPERATIONAL METRICS
// ============================================================================

export interface PerformanceOperationalMetrics {
  api_performance: {
    total_requests: number;
    successful_requests: number;
    failed_requests: number;
    average_response_time_ms: number;
    p95_response_time_ms: number;
    p99_response_time_ms: number;
    requests_per_second: number;
  };
  endpoint_metrics: {
    [endpoint: string]: {
      requests: number;
      errors: number;
      average_latency_ms: number;
      success_rate: number;
    };
  };
  error_tracking: {
    total_errors: number;
    error_rate_percentage: number;
    critical_errors: number;
    error_trends: ErrorTrend[];
    recent_errors: ErrorEvent[];
  };
  load_testing: {
    concurrent_users: number;
    max_concurrent_users: number;
    throughput_requests_per_second: number;
  };
  
  // Enhanced Performance Metrics
  real_time_api_latency: {
    [endpoint: string]: {
      current_latency_ms: number;
      p50_latency_ms: number;
      p95_latency_ms: number;
      p99_latency_ms: number;
      requests_per_minute: number;
      error_rate_percent: number;
    };
  };
  session_metrics: {
    concurrent_user_sessions: number;
    active_admin_sessions: number;
    session_creation_rate_per_min: number;
    session_timeout_count: number;
    average_session_duration_min: number;
  };
  database_performance: {
    encrypted_mappings_queries_per_sec: number;
    average_query_time_ms: number;
    slow_queries_count: number;
    connection_pool_usage_percent: number;
    database_latency_ms: number;
    encryption_overhead_ms: number;
  };
  system_health: {
    cpu_usage_percent: number;
    memory_usage_percent: number;
    disk_io_operations_per_sec: number;
    network_throughput_mbps: number;
    active_goroutines: number;
    gc_pause_time_ms: number;
    uptime_percentage: number;
    last_restart: string;
    health_check_status: 'healthy' | 'degraded' | 'critical';
    dependency_status: {
      database: 'healthy' | 'degraded' | 'critical';
      storage: 'healthy' | 'degraded' | 'critical';
      encryption: 'healthy' | 'degraded' | 'critical';
    };
  };
  load_test_results: LoadTestResult[];
}

// ============================================================================
// ALERTS & NOTIFICATIONS
// ============================================================================

export interface Alert {
  id: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: 'security' | 'performance' | 'system' | 'compliance' | 'zkid' | 'pqc' | 'email' | 'database';
  title: string;
  description: string;
  timestamp: string;
  acknowledged: boolean;
  acknowledged_by?: string;
  acknowledged_at?: string;
  resolved: boolean;
  resolved_at?: string;
  metadata?: Record<string, unknown>;
  
  // Enhanced Alert Properties
  source_component?: 'zkid' | 'pqc' | 'email' | 'database' | 'authentication' | 'api';
  threshold_value?: number;
  current_value?: number;
  alert_type: 'threshold_exceeded' | 'anomaly_detected' | 'system_failure' | 'security_incident';
  notification_channels: ('email' | 'webhook' | 'dashboard' | 'slack')[];
  escalation_level: number;
  auto_resolve_after_hours?: number;
}

export interface AlertRule {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  condition: string;
  threshold: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: 'security' | 'performance' | 'system' | 'compliance';
  actions: string[];
}

// ============================================================================
// AUDIT LOGGING
// ============================================================================

export interface AuditLogEntry {
  id: string;
  timestamp: string;
  user_id: string;
  action: string;
  resource: string;
  details: Record<string, unknown>;
  ip_address: string;
  user_agent: string;
  success: boolean;
  error_message?: string;
}

export interface AuditLogFilter {
  start_date?: string;
  end_date?: string;
  user_id?: string;
  action?: string;
  resource?: string;
  success?: boolean;
  limit?: number;
  offset?: number;
}

// ============================================================================
// REAL-TIME MONITORING & WEBSOCKET
// ============================================================================

export interface RealTimeUpdate {
  id: string;
  timestamp: string;
  type: 'metric_update' | 'alert' | 'status_change' | 'performance_data';
  component: 'zkid' | 'pqc' | 'email' | 'database' | 'authentication' | 'api';
  data: Record<string, unknown>;
  severity?: 'info' | 'warning' | 'error' | 'critical';
}

export interface WebSocketMessage {
  type: 'metric_update' | 'alert' | 'status_change' | 'heartbeat';
  timestamp: string;
  data: Record<string, unknown>;
  component?: string;
}

export interface MonitoringThreshold {
  component: string;
  metric: string;
  threshold: number;
  operator: 'gt' | 'lt' | 'eq' | 'gte' | 'lte';
  severity: 'low' | 'medium' | 'high' | 'critical';
  enabled: boolean;
}

// ============================================================================
// DASHBOARD CONFIGURATION
// ============================================================================

export interface DashboardConfig {
  refresh_interval_seconds: number;
  auto_refresh_enabled: boolean;
  theme: 'light' | 'dark' | 'auto';
  layout: 'compact' | 'comfortable' | 'spacious';
  panels: {
    [panel_id: string]: {
      enabled: boolean;
      position: number;
      size: 'small' | 'medium' | 'large';
    };
  };
  alerts: {
    enabled: boolean;
    sound_enabled: boolean;
    desktop_notifications: boolean;
    email_notifications: boolean;
  };
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

export interface EndpointMetric {
  requests: number;
  errors: number;
  average_latency_ms: number;
  success_rate: number;
  last_updated: string;
}

export interface RecoveryActivity {
  id: string;
  user_id: string;
  action: 'generated' | 'used' | 'revoked';
  timestamp: string;
  ip_address: string;
  success: boolean;
}

export interface ErrorTrend {
  timestamp: string;
  error_count: number;
  error_rate: number;
}

export interface ErrorEvent {
  id: string;
  timestamp: string;
  error_type: string;
  message: string;
  stack_trace?: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
}

export interface LoadTestResult {
  id: string;
  timestamp: string;
  concurrent_users: number;
  requests_per_second: number;
  average_response_time_ms: number;
  error_rate: number;
  success: boolean;
}

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  error_code?: string;
  timestamp: string;
}

export interface SystemEvent {
  type: 'metric_update' | 'alert' | 'log_entry' | 'system_event';
  data: Record<string, unknown>;
  timestamp: string;
}

// ============================================================================
// EXPORT ALL TYPES
// ============================================================================

// All types are exported inline above

// ============================================================================
// ANALYTICS & TREND ANALYSIS - ITERATION 6
// ============================================================================

export interface AnalyticsTimeRange {
  start: string;
  end: string;
  granularity: 'hour' | 'day' | 'week' | 'month';
}

export interface EmailUsageAnalytics {
  total_emails_sent: number;
  total_emails_received: number;
  emails_by_type: {
    secure: number;
    read_once: number;
    self_destruct: number;
    burn_after_read: number;
    password_protected: number;
  };
  emails_by_status: {
    delivered: number;
    failed: number;
    pending: number;
    expired: number;
    destroyed: number;
  };
  top_senders: Array<{
    email: string;
    count: number;
    percentage: number;
  }>;
  top_recipients: Array<{
    email: string;
    count: number;
    percentage: number;
  }>;
  delivery_times: {
    average_seconds: number;
    p50_seconds: number;
    p95_seconds: number;
    p99_seconds: number;
  };
  storage_usage: {
    total_bytes: number;
    encrypted_bytes: number;
    compression_ratio: number;
    growth_rate_percent: number;
  };
  trend_data: Array<{
    timestamp: string;
    emails_sent: number;
    emails_received: number;
    storage_used: number;
    delivery_success_rate: number;
  }>;
}

export interface SecurityEventAnalytics {
  total_security_events: number;
  events_by_severity: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  events_by_type: {
    failed_login_attempts: number;
    brute_force_attempts: number;
    unauthorized_access: number;
    suspicious_activity: number;
    encryption_failures: number;
    key_rotation_failures: number;
    audit_log_tampering: number;
    compliance_violations: number;
  };
  events_by_component: {
    zkid_layer: number;
    pqc_encryption: number;
    email_pipeline: number;
    admin_dashboard: number;
    authentication: number;
    realtime_monitoring: number;
  };
  threat_indicators: {
    ioc_count: number;
    suspicious_ips: number;
    anomalous_patterns: number;
    potential_attacks: number;
  };
  trend_data: Array<{
    timestamp: string;
    total_events: number;
    critical_events: number;
    high_events: number;
    medium_events: number;
    low_events: number;
  }>;
}

export interface ZKIDPQCAnalytics {
  zkid_operations: {
    total_mappings_created: number;
    total_mappings_retrieved: number;
    total_recovery_codes_generated: number;
    total_recovery_codes_used: number;
    total_recovery_codes_revoked: number;
    active_mappings: number;
    expired_mappings: number;
    revoked_mappings: number;
  };
  pqc_operations: {
    total_keys_generated: number;
    total_keys_rotated: number;
    total_encryptions: number;
    total_decryptions: number;
    hybrid_encryptions: number;
    classical_encryptions: number;
    encryption_failures: number;
    decryption_failures: number;
  };
  performance_metrics: {
    zkid_creation_latency_ms: number;
    zkid_retrieval_latency_ms: number;
    pqc_encryption_latency_ms: number;
    pqc_decryption_latency_ms: number;
    key_rotation_duration_ms: number;
  };
  security_metrics: {
    recovery_code_usage_rate: number;
    key_rotation_success_rate: number;
    encryption_success_rate: number;
    decryption_success_rate: number;
    zero_knowledge_compliance: number;
  };
  trend_data: Array<{
    timestamp: string;
    zkid_operations: number;
    pqc_operations: number;
    avg_zkid_latency: number;
    avg_pqc_latency: number;
    security_score: number;
  }>;
}

export interface ThreatIntelligence {
  threat_level: 'low' | 'medium' | 'high' | 'critical';
  threat_score: number; // 0-100
  active_threats: number;
  emerging_threats: number;
  threat_categories: {
    email_attacks: number;
    cryptographic_attacks: number;
    infrastructure_attacks: number;
    social_engineering: number;
    insider_threats: number;
  };
  threat_indicators: Array<{
    id: string;
    type: 'ip' | 'domain' | 'hash' | 'pattern';
    value: string;
    confidence: number;
    first_seen: string;
    last_seen: string;
    threat_level: 'low' | 'medium' | 'high' | 'critical';
    description: string;
  }>;
  threat_trends: Array<{
    timestamp: string;
    threat_level: 'low' | 'medium' | 'high' | 'critical';
    threat_score: number;
    active_threats: number;
    emerging_threats: number;
  }>;
  recommendations: Array<{
    id: string;
    priority: 'low' | 'medium' | 'high' | 'critical';
    category: string;
    title: string;
    description: string;
    action_required: boolean;
    estimated_effort: string;
  }>;
}

export interface AnalyticsDashboard {
  time_range: AnalyticsTimeRange;
  email_usage: EmailUsageAnalytics;
  security_events: SecurityEventAnalytics;
  zkid_pqc: ZKIDPQCAnalytics;
  threat_intelligence: ThreatIntelligence;
  last_updated: string;
  refresh_interval: number;
}

export interface ChartDataPoint {
  timestamp: string;
  value: number;
  label?: string;
  metadata?: Record<string, unknown>;
}

export interface ChartConfig {
  type: 'line' | 'bar' | 'pie' | 'doughnut' | 'area' | 'scatter';
  title: string;
  description?: string;
  x_axis_label?: string;
  y_axis_label?: string;
  data: ChartDataPoint[];
  options?: {
    show_legend?: boolean;
    show_grid?: boolean;
    animate?: boolean;
    responsive?: boolean;
    maintain_aspect_ratio?: boolean;
  };
}

export interface AnalyticsFilter {
  time_range?: AnalyticsTimeRange;
  component?: 'zkid' | 'pqc' | 'email' | 'security' | 'performance';
  severity?: 'low' | 'medium' | 'high' | 'critical';
  event_type?: string;
  user_id?: string;
  organization_id?: string;
}

export interface AnalyticsExportRequest {
  filter: AnalyticsFilter;
  format: 'json' | 'csv' | 'pdf';
  include_charts?: boolean;
  include_raw_data?: boolean;
}

export interface AnalyticsExportResponse {
  download_url: string;
  expires_at: string;
  file_size: number;
  record_count: number;
}

// ============================================================================
// THREAT AWARENESS MODULE - ITERATION 6
// ============================================================================

export interface ThreatAlert {
  id: string;
  timestamp: string;
  threat_level: 'low' | 'medium' | 'high' | 'critical';
  category: 'email' | 'cryptographic' | 'infrastructure' | 'social_engineering' | 'insider';
  title: string;
  description: string;
  indicators: string[];
  affected_components: string[];
  recommended_actions: string[];
  status: 'active' | 'investigating' | 'mitigated' | 'resolved';
  assigned_to?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface ThreatFeed {
  id: string;
  name: string;
  description: string;
  url: string;
  format: 'json' | 'xml' | 'csv';
  update_frequency: 'hourly' | 'daily' | 'weekly';
  last_update: string;
  next_update: string;
  status: 'active' | 'inactive' | 'error';
  threat_indicators_count: number;
  confidence_score: number;
}

export interface ThreatRule {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: 'low' | 'medium' | 'high' | 'critical';
  conditions: Array<{
    field: string;
    operator: 'equals' | 'contains' | 'greater_than' | 'less_than' | 'regex';
    value: string | number;
  }>;
  actions: Array<{
    type: 'alert' | 'block' | 'log' | 'email' | 'webhook';
    parameters: Record<string, unknown>;
  }>;
  created_at: string;
  updated_at: string;
}

export interface ThreatAwarenessConfig {
  enabled: boolean;
  threat_feeds: ThreatFeed[];
  threat_rules: ThreatRule[];
  alert_thresholds: {
    low_threshold: number;
    medium_threshold: number;
    high_threshold: number;
    critical_threshold: number;
  };
  auto_blocking: boolean;
  notification_channels: string[];
  update_interval_minutes: number;
}
