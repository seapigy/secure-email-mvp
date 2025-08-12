// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION SYSTEM
// =============================================================================
// Package notification handles access notifications for secure emails.
// Provides email and SMS notifications with metadata about access attempts.
// Enhanced with delivery frequency controls and rate limiting (Micro-Iteration 4.18).
// =============================================================================

package notification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
)

// AccessEventType represents the type of access event
type AccessEventType string

const (
	AccessEventTypeSuccess AccessEventType = "success"
	AccessEventTypeFailure AccessEventType = "failure"
	AccessEventTypeBlocked AccessEventType = "blocked"
)

// DeliveryFrequency represents the frequency of notification delivery
type DeliveryFrequency string

const (
	DeliveryFrequencyImmediate        DeliveryFrequency = "immediate"
	DeliveryFrequencyDailyDigest      DeliveryFrequency = "daily_digest"
	DeliveryFrequencyFirstAttemptOnly DeliveryFrequency = "first_attempt_only"
	DeliveryFrequencyThresholdTrigger DeliveryFrequency = "threshold_trigger"
)

// SuppressionReason represents the reason for suppressing a notification
type SuppressionReason string

const (
	SuppressionReasonRateLimited         SuppressionReason = "rate_limited"
	SuppressionReasonFrequencyControlled SuppressionReason = "frequency_controlled"
	SuppressionReasonThresholdNotMet     SuppressionReason = "threshold_not_met"
	SuppressionReasonFirstAttemptOnly    SuppressionReason = "first_attempt_only"
)

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID                    string            `json:"user_id"`
	EmailNotifications        bool              `json:"email_notifications"`
	SMSNotifications          bool              `json:"sms_notifications"`
	NotifyOnSuccess           bool              `json:"notify_on_success"`
	NotifyOnFailure           bool              `json:"notify_on_failure"`
	NotifyOnBlocked           bool              `json:"notify_on_blocked"`
	IncludeGeolocation        bool              `json:"include_geolocation"`
	IncludeDeviceInfo         bool              `json:"include_device_info"`
	DeliveryFrequency         DeliveryFrequency `json:"delivery_frequency"`
	ThresholdAttempts         int               `json:"threshold_attempts"`
	RateLimitWindowMinutes    int               `json:"rate_limit_window_minutes"`
	RateLimitMaxNotifications int               `json:"rate_limit_max_notifications"`
	DigestDeliveryTime        string            `json:"digest_delivery_time"`
	DigestEmailEnabled        bool              `json:"digest_email_enabled"`
	DigestSMSEnabled          bool              `json:"digest_sms_enabled"`
	PushNotificationsEnabled  bool              `json:"push_notifications_enabled"`
	SignalEnabled             bool              `json:"signal_enabled"`
	MatrixEnabled             bool              `json:"matrix_enabled"`
	TelegramEnabled           bool              `json:"telegram_enabled"`
	DiscordEnabled            bool              `json:"discord_enabled"`
	PushDeviceToken           string            `json:"push_device_token"`
	SignalPhone               string            `json:"signal_phone"`
	MatrixUserID              string            `json:"matrix_user_id"`
	MatrixHomeserver          string            `json:"matrix_homeserver"`
	TelegramChatID            string            `json:"telegram_chat_id"`
	DiscordWebhookURL         string            `json:"discord_webhook_url"`
	HighRiskChannels          string            `json:"high_risk_channels"`
	HighRiskThreshold         int               `json:"high_risk_threshold"`
	HighRiskTimeoutMinutes    int               `json:"high_risk_timeout_minutes"`
	CreatedAt                 time.Time         `json:"created_at"`
	UpdatedAt                 time.Time         `json:"updated_at"`
}

// EmailNotificationPreferences represents per-email notification preferences
type EmailNotificationPreferences struct {
	EmailID                   string            `json:"email_id"`
	UserID                    string            `json:"user_id"`
	DeliveryFrequency         DeliveryFrequency `json:"delivery_frequency"`
	ThresholdAttempts         int               `json:"threshold_attempts"`
	RateLimitWindowMinutes    int               `json:"rate_limit_window_minutes"`
	RateLimitMaxNotifications int               `json:"rate_limit_max_notifications"`
	InheritGlobalSettings     bool              `json:"inherit_global_settings"`
	DigestDeliveryTime        string            `json:"digest_delivery_time"`
	DigestEmailEnabled        bool              `json:"digest_email_enabled"`
	DigestSMSEnabled          bool              `json:"digest_sms_enabled"`
	PushNotificationsEnabled  bool              `json:"push_notifications_enabled"`
	SignalEnabled             bool              `json:"signal_enabled"`
	MatrixEnabled             bool              `json:"matrix_enabled"`
	TelegramEnabled           bool              `json:"telegram_enabled"`
	DiscordEnabled            bool              `json:"discord_enabled"`
	PushDeviceToken           string            `json:"push_device_token"`
	SignalPhone               string            `json:"signal_phone"`
	MatrixUserID              string            `json:"matrix_user_id"`
	MatrixHomeserver          string            `json:"matrix_homeserver"`
	TelegramChatID            string            `json:"telegram_chat_id"`
	DiscordWebhookURL         string            `json:"discord_webhook_url"`
	HighRiskChannels          string            `json:"high_risk_channels"`
	HighRiskThreshold         int               `json:"high_risk_threshold"`
	HighRiskTimeoutMinutes    int               `json:"high_risk_timeout_minutes"`
	CreatedAt                 time.Time         `json:"created_at"`
	UpdatedAt                 time.Time         `json:"updated_at"`
}

// NotificationSuppression represents a suppressed notification
type NotificationSuppression struct {
	SuppressionID     string            `json:"suppression_id"`
	EmailID           string            `json:"email_id"`
	UserID            string            `json:"user_id"`
	EventID           string            `json:"event_id"`
	SuppressionReason SuppressionReason `json:"suppression_reason"`
	SuppressedAt      time.Time         `json:"suppressed_at"`
	OriginalEventType string            `json:"original_event_type"`
	IPAddress         string            `json:"ip_address,omitempty"`
	UserAgent         string            `json:"user_agent,omitempty"`
	Country           string            `json:"country,omitempty"`
	City              string            `json:"city,omitempty"`
	DeviceType        string            `json:"device_type,omitempty"`
	FailureReason     string            `json:"failure_reason,omitempty"`
}

// AccessEvent represents an email access event
type AccessEvent struct {
	EventID       string          `json:"event_id"`
	EmailID       string          `json:"email_id"`
	UserID        string          `json:"user_id"`
	EventType     AccessEventType `json:"event_type"`
	IPAddress     string          `json:"ip_address"`
	UserAgent     string          `json:"user_agent"`
	Country       string          `json:"country,omitempty"`
	City          string          `json:"city,omitempty"`
	DeviceType    string          `json:"device_type,omitempty"`
	FailureReason string          `json:"failure_reason,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// NotificationService handles notification operations
type NotificationService struct {
	db *sql.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{
		db: db,
	}
}

// RecordAccessEvent records an access event in the database
func (ns *NotificationService) RecordAccessEvent(ctx context.Context, event *AccessEvent) error {
	query := `
		INSERT INTO access_events (
			event_id, email_id, user_id, event_type, ip_address, user_agent,
			country, city, device_type, failure_reason, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ns.db.ExecContext(ctx, query,
		event.EventID,
		event.EmailID,
		event.UserID,
		event.EventType,
		event.IPAddress,
		event.UserAgent,
		event.Country,
		event.City,
		event.DeviceType,
		event.FailureReason,
		event.Timestamp,
	)

	if err != nil {
		log.Printf("Failed to record access event: %v", err)
		return err
	}

	return nil
}

// GetNotificationPreferences retrieves notification preferences for a user
func (ns *NotificationService) GetNotificationPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
	query := `
		SELECT user_id, email_notifications, sms_notifications, notify_on_success,
			   notify_on_failure, notify_on_blocked, include_geolocation, include_device_info,
			   delivery_frequency, threshold_attempts, rate_limit_window_minutes, rate_limit_max_notifications,
			   digest_delivery_time, digest_email_enabled, digest_sms_enabled,
			   push_notifications_enabled, signal_enabled, matrix_enabled, telegram_enabled, discord_enabled,
			   push_device_token, signal_phone, matrix_user_id, matrix_homeserver, telegram_chat_id, discord_webhook_url,
			   high_risk_channels, high_risk_threshold, high_risk_timeout_minutes,
			   created_at, updated_at
		FROM notification_preferences
		WHERE user_id = ?
	`

	var prefs NotificationPreferences
	err := ns.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID,
		&prefs.EmailNotifications,
		&prefs.SMSNotifications,
		&prefs.NotifyOnSuccess,
		&prefs.NotifyOnFailure,
		&prefs.NotifyOnBlocked,
		&prefs.IncludeGeolocation,
		&prefs.IncludeDeviceInfo,
		&prefs.DeliveryFrequency,
		&prefs.ThresholdAttempts,
		&prefs.RateLimitWindowMinutes,
		&prefs.RateLimitMaxNotifications,
		&prefs.DigestDeliveryTime,
		&prefs.DigestEmailEnabled,
		&prefs.DigestSMSEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences if none exist
		return &NotificationPreferences{
			UserID:                    userID,
			EmailNotifications:        true,
			SMSNotifications:          false,
			NotifyOnSuccess:           true,
			NotifyOnFailure:           true,
			NotifyOnBlocked:           true,
			IncludeGeolocation:        true,
			IncludeDeviceInfo:         true,
			DeliveryFrequency:         DeliveryFrequencyImmediate,
			ThresholdAttempts:         3,
			RateLimitWindowMinutes:    15,
			RateLimitMaxNotifications: 5,
			DigestDeliveryTime:        "08:00",
			DigestEmailEnabled:        true,
			DigestSMSEnabled:          false,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}, nil
	}

	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		return nil, err
	}

	return &prefs, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (ns *NotificationService) UpdateNotificationPreferences(ctx context.Context, prefs *NotificationPreferences) error {
	query := `
		INSERT OR REPLACE INTO notification_preferences (
			user_id, email_notifications, sms_notifications, notify_on_success,
			notify_on_failure, notify_on_blocked, include_geolocation, include_device_info,
			delivery_frequency, threshold_attempts, rate_limit_window_minutes, rate_limit_max_notifications,
			digest_delivery_time, digest_email_enabled, digest_sms_enabled,
			push_notifications_enabled, signal_enabled, matrix_enabled, telegram_enabled, discord_enabled,
			push_device_token, signal_phone, matrix_user_id, matrix_homeserver, telegram_chat_id, discord_webhook_url,
			high_risk_channels, high_risk_threshold, high_risk_timeout_minutes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if prefs.CreatedAt.IsZero() {
		prefs.CreatedAt = now
	}
	prefs.UpdatedAt = now

	_, err := ns.db.ExecContext(ctx, query,
		prefs.UserID,
		prefs.EmailNotifications,
		prefs.SMSNotifications,
		prefs.NotifyOnSuccess,
		prefs.NotifyOnFailure,
		prefs.NotifyOnBlocked,
		prefs.IncludeGeolocation,
		prefs.IncludeDeviceInfo,
		prefs.DeliveryFrequency,
		prefs.ThresholdAttempts,
		prefs.RateLimitWindowMinutes,
		prefs.RateLimitMaxNotifications,
		prefs.DigestDeliveryTime,
		prefs.DigestEmailEnabled,
		prefs.DigestSMSEnabled,
		prefs.PushNotificationsEnabled,
		prefs.SignalEnabled,
		prefs.MatrixEnabled,
		prefs.TelegramEnabled,
		prefs.DiscordEnabled,
		prefs.PushDeviceToken,
		prefs.SignalPhone,
		prefs.MatrixUserID,
		prefs.MatrixHomeserver,
		prefs.TelegramChatID,
		prefs.DiscordWebhookURL,
		prefs.HighRiskChannels,
		prefs.HighRiskThreshold,
		prefs.HighRiskTimeoutMinutes,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to update notification preferences: %v", err)
		return err
	}

	return nil
}

// GetEmailNotificationPreferences retrieves per-email notification preferences
func (ns *NotificationService) GetEmailNotificationPreferences(ctx context.Context, emailID string) (*EmailNotificationPreferences, error) {
	query := `
		SELECT email_id, user_id, delivery_frequency, threshold_attempts,
			   rate_limit_window_minutes, rate_limit_max_notifications, inherit_global_settings,
			   digest_delivery_time, digest_email_enabled, digest_sms_enabled,
			   push_notifications_enabled, signal_enabled, matrix_enabled, telegram_enabled, discord_enabled,
			   push_device_token, signal_phone, matrix_user_id, matrix_homeserver, telegram_chat_id, discord_webhook_url,
			   high_risk_channels, high_risk_threshold, high_risk_timeout_minutes,
			   created_at, updated_at
		FROM email_notification_preferences
		WHERE email_id = ?
	`

	var prefs EmailNotificationPreferences
	err := ns.db.QueryRowContext(ctx, query, emailID).Scan(
		&prefs.EmailID,
		&prefs.UserID,
		&prefs.DeliveryFrequency,
		&prefs.ThresholdAttempts,
		&prefs.RateLimitWindowMinutes,
		&prefs.RateLimitMaxNotifications,
		&prefs.InheritGlobalSettings,
		&prefs.DigestDeliveryTime,
		&prefs.DigestEmailEnabled,
		&prefs.DigestSMSEnabled,
		&prefs.PushNotificationsEnabled,
		&prefs.SignalEnabled,
		&prefs.MatrixEnabled,
		&prefs.TelegramEnabled,
		&prefs.DiscordEnabled,
		&prefs.PushDeviceToken,
		&prefs.SignalPhone,
		&prefs.MatrixUserID,
		&prefs.MatrixHomeserver,
		&prefs.TelegramChatID,
		&prefs.DiscordWebhookURL,
		&prefs.HighRiskChannels,
		&prefs.HighRiskThreshold,
		&prefs.HighRiskTimeoutMinutes,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return nil to indicate no email-specific preferences
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get email notification preferences: %v", err)
		return nil, err
	}

	return &prefs, nil
}

// UpdateEmailNotificationPreferences updates per-email notification preferences
func (ns *NotificationService) UpdateEmailNotificationPreferences(ctx context.Context, prefs *EmailNotificationPreferences) error {
	query := `
		INSERT OR REPLACE INTO email_notification_preferences (
			email_id, user_id, delivery_frequency, threshold_attempts,
			rate_limit_window_minutes, rate_limit_max_notifications, inherit_global_settings,
			digest_delivery_time, digest_email_enabled, digest_sms_enabled,
			push_notifications_enabled, signal_enabled, matrix_enabled, telegram_enabled, discord_enabled,
			push_device_token, signal_phone, matrix_user_id, matrix_homeserver, telegram_chat_id, discord_webhook_url,
			high_risk_channels, high_risk_threshold, high_risk_timeout_minutes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if prefs.CreatedAt.IsZero() {
		prefs.CreatedAt = now
	}
	prefs.UpdatedAt = now

	_, err := ns.db.ExecContext(ctx, query,
		prefs.EmailID,
		prefs.UserID,
		prefs.DeliveryFrequency,
		prefs.ThresholdAttempts,
		prefs.RateLimitWindowMinutes,
		prefs.RateLimitMaxNotifications,
		prefs.InheritGlobalSettings,
		prefs.DigestDeliveryTime,
		prefs.DigestEmailEnabled,
		prefs.DigestSMSEnabled,
		prefs.PushNotificationsEnabled,
		prefs.SignalEnabled,
		prefs.MatrixEnabled,
		prefs.TelegramEnabled,
		prefs.DiscordEnabled,
		prefs.PushDeviceToken,
		prefs.SignalPhone,
		prefs.MatrixUserID,
		prefs.MatrixHomeserver,
		prefs.TelegramChatID,
		prefs.DiscordWebhookURL,
		prefs.HighRiskChannels,
		prefs.HighRiskThreshold,
		prefs.HighRiskTimeoutMinutes,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to update email notification preferences: %v", err)
		return err
	}

	return nil
}

// ShouldSendNotification determines if a notification should be sent based on preferences and delivery controls
func (ns *NotificationService) ShouldSendNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences, emailPrefs *EmailNotificationPreferences) (bool, SuppressionReason, error) {
	// Use email-specific preferences if available and not inheriting global settings
	effectivePrefs := prefs
	if emailPrefs != nil && !emailPrefs.InheritGlobalSettings {
		// Create effective preferences from email-specific settings
		effectivePrefs = &NotificationPreferences{
			UserID:                    prefs.UserID,
			EmailNotifications:        prefs.EmailNotifications,
			SMSNotifications:          prefs.SMSNotifications,
			NotifyOnSuccess:           prefs.NotifyOnSuccess,
			NotifyOnFailure:           prefs.NotifyOnFailure,
			NotifyOnBlocked:           prefs.NotifyOnBlocked,
			IncludeGeolocation:        prefs.IncludeGeolocation,
			IncludeDeviceInfo:         prefs.IncludeDeviceInfo,
			DeliveryFrequency:         emailPrefs.DeliveryFrequency,
			ThresholdAttempts:         emailPrefs.ThresholdAttempts,
			RateLimitWindowMinutes:    emailPrefs.RateLimitWindowMinutes,
			RateLimitMaxNotifications: emailPrefs.RateLimitMaxNotifications,
			DigestDeliveryTime:        emailPrefs.DigestDeliveryTime,
			DigestEmailEnabled:        emailPrefs.DigestEmailEnabled,
			DigestSMSEnabled:          emailPrefs.DigestSMSEnabled,
			// Multi-channel preferences
			PushNotificationsEnabled: emailPrefs.PushNotificationsEnabled,
			SignalEnabled:            emailPrefs.SignalEnabled,
			MatrixEnabled:            emailPrefs.MatrixEnabled,
			TelegramEnabled:          emailPrefs.TelegramEnabled,
			DiscordEnabled:           emailPrefs.DiscordEnabled,
			PushDeviceToken:          emailPrefs.PushDeviceToken,
			SignalPhone:              emailPrefs.SignalPhone,
			MatrixUserID:             emailPrefs.MatrixUserID,
			MatrixHomeserver:         emailPrefs.MatrixHomeserver,
			TelegramChatID:           emailPrefs.TelegramChatID,
			DiscordWebhookURL:        emailPrefs.DiscordWebhookURL,
			HighRiskChannels:         emailPrefs.HighRiskChannels,
			HighRiskThreshold:        emailPrefs.HighRiskThreshold,
			HighRiskTimeoutMinutes:   emailPrefs.HighRiskTimeoutMinutes,
		}
	}

	// Check basic notification preferences
	if !ns.shouldNotifyForEventType(effectivePrefs, event.EventType) {
		return false, SuppressionReasonFrequencyControlled, nil
	}

	// Check delivery frequency controls
	shouldSend, reason, err := ns.checkDeliveryFrequency(ctx, event, effectivePrefs)
	if err != nil {
		return false, "", err
	}
	if !shouldSend {
		return false, reason, nil
	}

	// Check rate limiting
	shouldSend, reason, err = ns.checkRateLimiting(ctx, event, effectivePrefs)
	if err != nil {
		return false, "", err
	}
	if !shouldSend {
		return false, reason, nil
	}

	return true, "", nil
}

// shouldNotifyForEventType checks if notifications are enabled for the event type
func (ns *NotificationService) shouldNotifyForEventType(prefs *NotificationPreferences, eventType AccessEventType) bool {
	switch eventType {
	case AccessEventTypeSuccess:
		return prefs.NotifyOnSuccess
	case AccessEventTypeFailure:
		return prefs.NotifyOnFailure
	case AccessEventTypeBlocked:
		return prefs.NotifyOnBlocked
	default:
		return false
	}
}

// checkDeliveryFrequency checks if notification should be sent based on delivery frequency
func (ns *NotificationService) checkDeliveryFrequency(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) (bool, SuppressionReason, error) {
	switch prefs.DeliveryFrequency {
	case DeliveryFrequencyImmediate:
		return true, "", nil

	case DeliveryFrequencyFirstAttemptOnly:
		// Check if this is the first attempt for this email/IP combination
		query := `
			SELECT COUNT(*) FROM access_events 
			WHERE email_id = ? AND ip_address = ? AND event_type = ?
		`
		var count int
		err := ns.db.QueryRowContext(ctx, query, event.EmailID, event.IPAddress, event.EventType).Scan(&count)
		if err != nil {
			return false, "", err
		}
		if count > 0 {
			return false, SuppressionReasonFirstAttemptOnly, nil
		}
		return true, "", nil

	case DeliveryFrequencyThresholdTrigger:
		// Only send notification after threshold number of failed attempts
		if event.EventType != AccessEventTypeFailure {
			return false, SuppressionReasonThresholdNotMet, nil
		}
		query := `
			SELECT COUNT(*) FROM access_events 
			WHERE email_id = ? AND event_type = 'failure'
		`
		var count int
		err := ns.db.QueryRowContext(ctx, query, event.EmailID).Scan(&count)
		if err != nil {
			return false, "", err
		}
		// Add 1 to count because we're checking if this current event would be the threshold
		if (count + 1) < prefs.ThresholdAttempts {
			return false, SuppressionReasonThresholdNotMet, nil
		}
		return true, "", nil

	case DeliveryFrequencyDailyDigest:
		// For daily digest, we'll track events but not send immediate notifications
		// The digest will be sent by a separate process
		return false, SuppressionReasonFrequencyControlled, nil

	default:
		return true, "", nil
	}
}

// checkRateLimiting checks if notification should be sent based on rate limiting
func (ns *NotificationService) checkRateLimiting(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) (bool, SuppressionReason, error) {
	// Clean up old rate limit records
	cleanupQuery := `
		DELETE FROM notification_rate_limits 
		WHERE window_start < datetime('now', '-' || ? || ' minutes')
	`
	_, err := ns.db.ExecContext(ctx, cleanupQuery, prefs.RateLimitWindowMinutes)
	if err != nil {
		log.Printf("Failed to cleanup rate limit records: %v", err)
	}

	// Check current rate limit for this email/IP combination
	query := `
		SELECT notification_count, window_start 
		FROM notification_rate_limits 
		WHERE email_id = ? AND ip_address = ?
	`
	var count int
	var windowStart time.Time
	err = ns.db.QueryRowContext(ctx, query, event.EmailID, event.IPAddress).Scan(&count, &windowStart)

	if err == sql.ErrNoRows {
		// No existing rate limit record, create one
		insertQuery := `
			INSERT INTO notification_rate_limits (rate_limit_id, email_id, user_id, ip_address, notification_count, window_start, last_notification_at)
			VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))
		`
		_, err = ns.db.ExecContext(ctx, insertQuery, uuid.New().String(), event.EmailID, event.UserID, event.IPAddress)
		if err != nil {
			return false, "", err
		}
		return true, "", nil
	}

	if err != nil {
		return false, "", err
	}

	// Check if we're within the rate limit
	if count >= prefs.RateLimitMaxNotifications {
		return false, SuppressionReasonRateLimited, nil
	}

	// Update the rate limit record
	updateQuery := `
		UPDATE notification_rate_limits 
		SET notification_count = notification_count + 1, last_notification_at = datetime('now')
		WHERE email_id = ? AND ip_address = ?
	`
	_, err = ns.db.ExecContext(ctx, updateQuery, event.EmailID, event.IPAddress)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

// RecordNotificationSuppression records a suppressed notification for audit purposes
func (ns *NotificationService) RecordNotificationSuppression(ctx context.Context, event *AccessEvent, reason SuppressionReason) error {
	query := `
		INSERT INTO notification_suppressions (
			suppression_id, email_id, user_id, event_id, suppression_reason, suppressed_at,
			original_event_type, ip_address, user_agent, country, city, device_type, failure_reason
		) VALUES (?, ?, ?, ?, ?, datetime('now'), ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ns.db.ExecContext(ctx, query,
		uuid.New().String(),
		event.EmailID,
		event.UserID,
		event.EventID,
		reason,
		event.EventType,
		event.IPAddress,
		event.UserAgent,
		event.Country,
		event.City,
		event.DeviceType,
		event.FailureReason,
	)

	if err != nil {
		log.Printf("Failed to record notification suppression: %v", err)
		return err
	}

	return nil
}

// SendNotification sends a notification based on the event and preferences
func (ns *NotificationService) SendNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	// Get email-specific preferences
	emailPrefs, err := ns.GetEmailNotificationPreferences(ctx, event.EmailID)
	if err != nil {
		log.Printf("Failed to get email notification preferences: %v", err)
		// Continue with global preferences
	}

	// Check if notification should be sent
	shouldSend, reason, err := ns.ShouldSendNotification(ctx, event, prefs, emailPrefs)
	if err != nil {
		log.Printf("Failed to determine if notification should be sent: %v", err)
		return err
	}

	if !shouldSend {
		// Record the suppression for audit purposes
		if err := ns.RecordNotificationSuppression(ctx, event, reason); err != nil {
			log.Printf("Failed to record notification suppression: %v", err)
		}
		log.Printf("Notification suppressed for event %s: %s", event.EventID, reason)
		return nil
	}

	// Send email notification if enabled
	if prefs.EmailNotifications {
		if err := ns.sendEmailNotification(ctx, event, prefs); err != nil {
			log.Printf("Failed to send email notification: %v", err)
			// Continue with SMS notification even if email fails
		}
	}

	// Send SMS notification if enabled
	if prefs.SMSNotifications {
		if err := ns.sendSMSNotification(ctx, event, prefs); err != nil {
			log.Printf("Failed to send SMS notification: %v", err)
			return err
		}
	}

	return nil
}

// sendEmailNotification sends an email notification
func (ns *NotificationService) sendEmailNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	// Get sender email from database
	var senderEmail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, event.UserID).Scan(&senderEmail)
	if err != nil {
		log.Printf("Failed to get sender email: %v", err)
		return err
	}

	// Build notification content
	subject := ns.buildEmailSubject(event)
	body := ns.buildEmailBody(event, prefs)

	// For now, log the email notification
	// In production, this would integrate with an email service like SendGrid, AWS SES, etc.
	log.Printf("EMAIL NOTIFICATION - To: %s, Subject: %s, Body: %s", senderEmail, subject, body)

	// TODO: Integrate with actual email service
	// Example with SendGrid:
	// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	// response, err := client.Send(message)

	return nil
}

// sendSMSNotification sends an SMS notification
func (ns *NotificationService) sendSMSNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	// Get sender phone number from database
	var phoneNumber string
	query := `SELECT phone_number FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, event.UserID).Scan(&phoneNumber)
	if err != nil {
		log.Printf("Failed to get sender phone number: %v", err)
		return err
	}

	// Build SMS content
	message := ns.buildSMSMessage(event, prefs)

	// For now, log the SMS notification
	// In production, this would integrate with Twilio or similar service
	log.Printf("SMS NOTIFICATION - To: %s, Message: %s", phoneNumber, message)

	// TODO: Integrate with actual SMS service
	// Example with Twilio:
	// message := client.CreateMessage(&twilio.CreateMessageParams{
	//     To:   &phoneNumber,
	//     From: &fromNumber,
	//     Body: &message,
	// })

	return nil
}

// buildEmailSubject builds the email subject line
func (ns *NotificationService) buildEmailSubject(event *AccessEvent) string {
	switch event.EventType {
	case AccessEventTypeSuccess:
		return "Secure Email Accessed Successfully"
	case AccessEventTypeFailure:
		return "Secure Email Access Attempt Failed"
	case AccessEventTypeBlocked:
		return "Secure Email Access Blocked"
	default:
		return "Secure Email Access Notification"
	}
}

// buildEmailBody builds the email body content
func (ns *NotificationService) buildEmailBody(event *AccessEvent, prefs *NotificationPreferences) string {
	var body strings.Builder

	body.WriteString("Hello,\n\n")
	body.WriteString("Your secure email has been accessed.\n\n")

	// Event details
	body.WriteString("Event Details:\n")
	body.WriteString(fmt.Sprintf("- Event Type: %s\n", event.EventType))
	body.WriteString(fmt.Sprintf("- Timestamp: %s\n", event.Timestamp.Format("2006-01-02 15:04:05 UTC")))

	// Include geolocation if enabled
	if prefs.IncludeGeolocation && event.Country != "" {
		body.WriteString(fmt.Sprintf("- Location: %s", event.Country))
		if event.City != "" {
			body.WriteString(fmt.Sprintf(", %s", event.City))
		}
		body.WriteString("\n")
	}

	// Include device info if enabled
	if prefs.IncludeDeviceInfo && event.DeviceType != "" {
		body.WriteString(fmt.Sprintf("- Device: %s\n", event.DeviceType))
	}

	// Include failure reason if applicable
	if event.FailureReason != "" {
		body.WriteString(fmt.Sprintf("- Reason: %s\n", event.FailureReason))
	}

	body.WriteString("\n")
	body.WriteString("If you did not expect this access, please review your security settings.\n\n")
	body.WriteString("Best regards,\nSecure Email MVP Team")

	return body.String()
}

// buildSMSMessage builds the SMS message content
func (ns *NotificationService) buildSMSMessage(event *AccessEvent, prefs *NotificationPreferences) string {
	var message strings.Builder

	switch event.EventType {
	case AccessEventTypeSuccess:
		message.WriteString("Secure email accessed successfully")
	case AccessEventTypeFailure:
		message.WriteString("Secure email access failed")
	case AccessEventTypeBlocked:
		message.WriteString("Secure email access blocked")
	}

	message.WriteString(fmt.Sprintf(" at %s", event.Timestamp.Format("15:04 UTC")))

	// Include location if enabled and available
	if prefs.IncludeGeolocation && event.Country != "" {
		message.WriteString(fmt.Sprintf(" from %s", event.Country))
	}

	return message.String()
}

// GetAccessEventHistory retrieves access event history for a user
func (ns *NotificationService) GetAccessEventHistory(ctx context.Context, userID string, limit int) ([]*AccessEvent, error) {
	query := `
		SELECT event_id, email_id, user_id, event_type, ip_address, user_agent,
			   country, city, device_type, failure_reason, timestamp
		FROM access_events
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := ns.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		log.Printf("Failed to get access event history: %v", err)
		return nil, err
	}
	defer rows.Close()

	var events []*AccessEvent
	for rows.Next() {
		var event AccessEvent
		err := rows.Scan(
			&event.EventID,
			&event.EmailID,
			&event.UserID,
			&event.EventType,
			&event.IPAddress,
			&event.UserAgent,
			&event.Country,
			&event.City,
			&event.DeviceType,
			&event.FailureReason,
			&event.Timestamp,
		)
		if err != nil {
			log.Printf("Failed to scan access event: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetNotificationSuppressions retrieves suppressed notifications for a user
func (ns *NotificationService) GetNotificationSuppressions(ctx context.Context, userID string, limit int) ([]*NotificationSuppression, error) {
	query := `
		SELECT suppression_id, email_id, user_id, event_id, suppression_reason, suppressed_at,
			   original_event_type, ip_address, user_agent, country, city, device_type, failure_reason
		FROM notification_suppressions
		WHERE user_id = ?
		ORDER BY suppressed_at DESC
		LIMIT ?
	`

	rows, err := ns.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		log.Printf("Failed to get notification suppressions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var suppressions []*NotificationSuppression
	for rows.Next() {
		var suppression NotificationSuppression
		err := rows.Scan(
			&suppression.SuppressionID,
			&suppression.EmailID,
			&suppression.UserID,
			&suppression.EventID,
			&suppression.SuppressionReason,
			&suppression.SuppressedAt,
			&suppression.OriginalEventType,
			&suppression.IPAddress,
			&suppression.UserAgent,
			&suppression.Country,
			&suppression.City,
			&suppression.DeviceType,
			&suppression.FailureReason,
		)
		if err != nil {
			log.Printf("Failed to scan notification suppression: %v", err)
			continue
		}
		suppressions = append(suppressions, &suppression)
	}

	return suppressions, nil
}

// GetSuppressionStats retrieves suppression statistics for a user
func (ns *NotificationService) GetSuppressionStats(ctx context.Context, userID string) (map[string]int, error) {
	query := `
		SELECT suppression_reason, COUNT(*) as count
		FROM notification_suppressions
		WHERE user_id = ?
		GROUP BY suppression_reason
	`

	rows, err := ns.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Failed to get suppression stats: %v", err)
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var reason string
		var count int
		err := rows.Scan(&reason, &count)
		if err != nil {
			log.Printf("Failed to scan suppression stat: %v", err)
			continue
		}
		stats[reason] = count
	}

	return stats, nil
}

// GenerateDailyDigest generates a daily digest summary for a user
func (ns *NotificationService) GenerateDailyDigest(ctx context.Context, userID string, digestDate time.Time) (*DigestSummary, error) {
	// Get user notification preferences
	prefs, err := ns.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	// Check if user has daily digest enabled
	if prefs.DeliveryFrequency != DeliveryFrequencyDailyDigest {
		return nil, fmt.Errorf("user does not have daily digest enabled")
	}

	// Calculate time range for the digest (24 hours from digest date)
	startTime := digestDate.Truncate(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	// Get all access events for the user in the time range
	query := `
		SELECT ae.event_id, ae.email_id, ae.user_id, ae.event_type, ae.ip_address, ae.user_agent,
			   ae.country, ae.city, ae.device_type, ae.failure_reason, ae.timestamp,
			   e.subject, e.recipient
		FROM access_events ae
		JOIN emails e ON ae.email_id = e.email_id
		WHERE ae.user_id = ? AND ae.timestamp >= ? AND ae.timestamp < ?
		ORDER BY ae.timestamp DESC
	`

	rows, err := ns.db.QueryContext(ctx, query, userID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query access events: %w", err)
	}
	defer rows.Close()

	// Group events by email
	emailEvents := make(map[string][]*AccessEvent)
	emailInfo := make(map[string]struct {
		subject   string
		recipient string
	})

	for rows.Next() {
		var event AccessEvent
		var subject, recipient string
		err := rows.Scan(
			&event.EventID,
			&event.EmailID,
			&event.UserID,
			&event.EventType,
			&event.IPAddress,
			&event.UserAgent,
			&event.Country,
			&event.City,
			&event.DeviceType,
			&event.FailureReason,
			&event.Timestamp,
			&subject,
			&recipient,
		)
		if err != nil {
			log.Printf("Failed to scan access event: %v", err)
			continue
		}

		emailEvents[event.EmailID] = append(emailEvents[event.EmailID], &event)
		emailInfo[event.EmailID] = struct {
			subject   string
			recipient string
		}{subject: subject, recipient: recipient}
	}

	// Get suppression counts for the time range
	suppressionQuery := `
		SELECT email_id, COUNT(*) as suppression_count
		FROM notification_suppressions
		WHERE user_id = ? AND suppressed_at >= ? AND suppressed_at < ?
		GROUP BY email_id
	`

	suppressionRows, err := ns.db.QueryContext(ctx, suppressionQuery, userID, startTime, endTime)
	if err != nil {
		log.Printf("Failed to query suppressions: %v", err)
	}
	defer suppressionRows.Close()

	suppressionCounts := make(map[string]int)
	for suppressionRows.Next() {
		var emailID string
		var count int
		err := suppressionRows.Scan(&emailID, &count)
		if err != nil {
			log.Printf("Failed to scan suppression count: %v", err)
			continue
		}
		suppressionCounts[emailID] = count
	}

	// Build digest summary
	summary := &DigestSummary{
		UserID:         userID,
		DigestDate:     digestDate,
		EmailSummaries: []*DigestEmailSummary{},
	}

	// Process each email's events
	for emailID, events := range emailEvents {
		if len(events) == 0 {
			continue
		}

		emailSummary := &DigestEmailSummary{
			EmailID:      emailID,
			EmailSubject: emailInfo[emailID].subject,
			Recipient:    emailInfo[emailID].recipient,
		}

		// Count event types
		for _, event := range events {
			summary.TotalEvents++
			switch event.EventType {
			case AccessEventTypeSuccess:
				emailSummary.SuccessCount++
				summary.SuccessCount++
			case AccessEventTypeFailure:
				emailSummary.FailureCount++
				summary.FailureCount++
			case AccessEventTypeBlocked:
				emailSummary.BlockedCount++
				summary.BlockedCount++
			}

			// Track last access details
			if emailSummary.LastAccessAt == nil || event.Timestamp.After(*emailSummary.LastAccessAt) {
				emailSummary.LastAccessAt = &event.Timestamp
				emailSummary.LastIPAddress = event.IPAddress
				emailSummary.LastDeviceType = event.DeviceType
				emailSummary.LastCountry = event.Country
				emailSummary.LastCity = event.City
			}
		}

		// Add suppression count
		emailSummary.SuppressionCount = suppressionCounts[emailID]
		summary.SuppressionCount += suppressionCounts[emailID]

		summary.EmailSummaries = append(summary.EmailSummaries, emailSummary)
	}

	summary.TotalEmails = len(summary.EmailSummaries)

	return summary, nil
}

// SendDailyDigest sends a daily digest to a user
func (ns *NotificationService) SendDailyDigest(ctx context.Context, userID string, digestDate time.Time) error {
	// Generate digest summary
	summary, err := ns.GenerateDailyDigest(ctx, userID, digestDate)
	if err != nil {
		return fmt.Errorf("failed to generate digest: %w", err)
	}

	// Check if there are any events to report
	if summary.TotalEvents == 0 {
		log.Printf("No events to report in digest for user %s on %s", userID, digestDate.Format("2006-01-02"))
		return nil
	}

	// Get user preferences
	prefs, err := ns.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get notification preferences: %w", err)
	}

	// Create delivery record
	deliveryID := uuid.New().String()
	delivery := &DailyDigestDelivery{
		DeliveryID:       deliveryID,
		UserID:           userID,
		DigestDate:       digestDate,
		EventCount:       summary.TotalEvents,
		EmailCount:       summary.TotalEmails,
		SuccessCount:     summary.SuccessCount,
		FailureCount:     summary.FailureCount,
		BlockedCount:     summary.BlockedCount,
		SuppressionCount: summary.SuppressionCount,
		CreatedAt:        time.Now(),
	}

	// Store delivery record
	if err := ns.storeDigestDelivery(ctx, delivery); err != nil {
		return fmt.Errorf("failed to store digest delivery: %w", err)
	}

	// Store digest content for audit
	if err := ns.storeDigestContent(ctx, deliveryID, summary); err != nil {
		log.Printf("Failed to store digest content: %v", err)
	}

	// Send email digest if enabled
	if prefs.DigestEmailEnabled {
		if err := ns.sendDigestEmail(ctx, summary, prefs); err != nil {
			log.Printf("Failed to send digest email: %v", err)
		} else {
			delivery.EmailSent = true
			now := time.Now()
			delivery.EmailSentAt = &now
		}
	}

	// Send SMS digest if enabled
	if prefs.DigestSMSEnabled {
		if err := ns.sendDigestSMS(ctx, summary, prefs); err != nil {
			log.Printf("Failed to send digest SMS: %v", err)
		} else {
			delivery.SMSSent = true
			now := time.Now()
			delivery.SMSSentAt = &now
		}
	}

	// Update delivery record with sent status
	if err := ns.updateDigestDelivery(ctx, delivery); err != nil {
		log.Printf("Failed to update digest delivery status: %v", err)
	}

	return nil
}

// storeDigestDelivery stores a digest delivery record
func (ns *NotificationService) storeDigestDelivery(ctx context.Context, delivery *DailyDigestDelivery) error {
	query := `
		INSERT INTO daily_digest_deliveries (
			delivery_id, user_id, digest_date, email_sent, sms_sent,
			email_sent_at, sms_sent_at, event_count, email_count,
			success_count, failure_count, blocked_count, suppression_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ns.db.ExecContext(ctx, query,
		delivery.DeliveryID,
		delivery.UserID,
		delivery.DigestDate.Format("2006-01-02"),
		delivery.EmailSent,
		delivery.SMSSent,
		delivery.EmailSentAt,
		delivery.SMSSentAt,
		delivery.EventCount,
		delivery.EmailCount,
		delivery.SuccessCount,
		delivery.FailureCount,
		delivery.BlockedCount,
		delivery.SuppressionCount,
		delivery.CreatedAt,
	)

	return err
}

// updateDigestDelivery updates a digest delivery record
func (ns *NotificationService) updateDigestDelivery(ctx context.Context, delivery *DailyDigestDelivery) error {
	query := `
		UPDATE daily_digest_deliveries SET
			email_sent = ?, sms_sent = ?, email_sent_at = ?, sms_sent_at = ?
		WHERE delivery_id = ?
	`

	_, err := ns.db.ExecContext(ctx, query,
		delivery.EmailSent,
		delivery.SMSSent,
		delivery.EmailSentAt,
		delivery.SMSSentAt,
		delivery.DeliveryID,
	)

	return err
}

// storeDigestContent stores digest content for audit purposes
func (ns *NotificationService) storeDigestContent(ctx context.Context, deliveryID string, summary *DigestSummary) error {
	for _, emailSummary := range summary.EmailSummaries {
		content := &DailyDigestContent{
			ContentID:        uuid.New().String(),
			DeliveryID:       deliveryID,
			EmailID:          emailSummary.EmailID,
			EmailSubject:     emailSummary.EmailSubject,
			Recipient:        emailSummary.Recipient,
			SuccessCount:     emailSummary.SuccessCount,
			FailureCount:     emailSummary.FailureCount,
			BlockedCount:     emailSummary.BlockedCount,
			LastAccessAt:     emailSummary.LastAccessAt,
			LastIPAddress:    emailSummary.LastIPAddress,
			LastDeviceType:   emailSummary.LastDeviceType,
			LastCountry:      emailSummary.LastCountry,
			LastCity:         emailSummary.LastCity,
			SuppressionCount: emailSummary.SuppressionCount,
			CreatedAt:        time.Now(),
		}

		query := `
			INSERT INTO daily_digest_content (
				content_id, delivery_id, email_id, email_subject, recipient,
				success_count, failure_count, blocked_count, last_access_at,
				last_ip_address, last_device_type, last_country, last_city,
				suppression_count, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := ns.db.ExecContext(ctx, query,
			content.ContentID,
			content.DeliveryID,
			content.EmailID,
			content.EmailSubject,
			content.Recipient,
			content.SuccessCount,
			content.FailureCount,
			content.BlockedCount,
			content.LastAccessAt,
			content.LastIPAddress,
			content.LastDeviceType,
			content.LastCountry,
			content.LastCity,
			content.SuppressionCount,
			content.CreatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to store digest content for email %s: %w", content.EmailID, err)
		}
	}

	return nil
}

// sendDigestEmail sends a daily digest email
func (ns *NotificationService) sendDigestEmail(ctx context.Context, summary *DigestSummary, prefs *NotificationPreferences) error {
	// Get sender email from database
	var senderEmail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, summary.UserID).Scan(&senderEmail)
	if err != nil {
		return fmt.Errorf("failed to get sender email: %w", err)
	}

	// Build digest email content
	subject := fmt.Sprintf("Secure Email Daily Digest - %s", summary.DigestDate.Format("January 2, 2006"))
	body := ns.buildDigestEmailBody(summary, prefs)

	// For now, log the digest email
	// In production, this would integrate with an email service
	log.Printf("DAILY DIGEST EMAIL - To: %s, Subject: %s, Body: %s", senderEmail, subject, body)

	// TODO: Integrate with actual email service
	// Example with SendGrid:
	// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	// response, err := client.Send(message)

	return nil
}

// sendDigestSMS sends a daily digest SMS
func (ns *NotificationService) sendDigestSMS(ctx context.Context, summary *DigestSummary, prefs *NotificationPreferences) error {
	// Get sender phone number from database
	var phoneNumber string
	query := `SELECT phone_number FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, summary.UserID).Scan(&phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to get sender phone number: %w", err)
	}

	// Build digest SMS content
	message := ns.buildDigestSMSMessage(summary, prefs)

	// For now, log the digest SMS
	// In production, this would integrate with Twilio or similar service
	log.Printf("DAILY DIGEST SMS - To: %s, Message: %s", phoneNumber, message)

	// TODO: Integrate with actual SMS service
	// Example with Twilio:
	// message := client.CreateMessage(&twilio.CreateMessageParams{
	//     To:   &phoneNumber,
	//     From: &fromNumber,
	//     Body: &message,
	// })

	return nil
}

// buildDigestEmailBody builds the digest email body content
func (ns *NotificationService) buildDigestEmailBody(summary *DigestSummary, prefs *NotificationPreferences) string {
	var body strings.Builder

	body.WriteString("Hello,\n\n")
	body.WriteString(fmt.Sprintf("Here's your secure email activity summary for %s:\n\n", summary.DigestDate.Format("January 2, 2006")))

	// Summary statistics
	body.WriteString("Summary:\n")
	body.WriteString(fmt.Sprintf("- Total events: %d\n", summary.TotalEvents))
	body.WriteString(fmt.Sprintf("- Emails accessed: %d\n", summary.TotalEmails))
	body.WriteString(fmt.Sprintf("- Successful accesses: %d\n", summary.SuccessCount))
	body.WriteString(fmt.Sprintf("- Failed attempts: %d\n", summary.FailureCount))
	body.WriteString(fmt.Sprintf("- Blocked attempts: %d\n", summary.BlockedCount))
	if summary.SuppressionCount > 0 {
		body.WriteString(fmt.Sprintf("- Suppressed notifications: %d\n", summary.SuppressionCount))
	}
	body.WriteString("\n")

	// Email details
	if len(summary.EmailSummaries) > 0 {
		body.WriteString("Email Details:\n")
		for _, emailSummary := range summary.EmailSummaries {
			body.WriteString(fmt.Sprintf("\nEmail: %s\n", emailSummary.EmailSubject))
			body.WriteString(fmt.Sprintf("Recipient: %s\n", emailSummary.Recipient))
			body.WriteString(fmt.Sprintf("Events: %d successful, %d failed, %d blocked\n",
				emailSummary.SuccessCount, emailSummary.FailureCount, emailSummary.BlockedCount))

			if emailSummary.LastAccessAt != nil {
				body.WriteString(fmt.Sprintf("Last access: %s\n", emailSummary.LastAccessAt.Format("2006-01-02 15:04:05 UTC")))
			}

			if prefs.IncludeGeolocation && emailSummary.LastCountry != "" {
				body.WriteString(fmt.Sprintf("Last location: %s", emailSummary.LastCountry))
				if emailSummary.LastCity != "" {
					body.WriteString(fmt.Sprintf(", %s", emailSummary.LastCity))
				}
				body.WriteString("\n")
			}

			if prefs.IncludeDeviceInfo && emailSummary.LastDeviceType != "" {
				body.WriteString(fmt.Sprintf("Last device: %s\n", emailSummary.LastDeviceType))
			}

			if emailSummary.SuppressionCount > 0 {
				body.WriteString(fmt.Sprintf("Suppressed notifications: %d\n", emailSummary.SuppressionCount))
			}
		}
	}

	body.WriteString("\n")
	body.WriteString("If you notice any suspicious activity, please review your security settings.\n\n")
	body.WriteString("Best regards,\nSecure Email MVP Team")

	return body.String()
}

// buildDigestSMSMessage builds the digest SMS message content
func (ns *NotificationService) buildDigestSMSMessage(summary *DigestSummary, prefs *NotificationPreferences) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("Secure Email Digest %s: ", summary.DigestDate.Format("1/2")))
	message.WriteString(fmt.Sprintf("%d events, %d emails, ", summary.TotalEvents, summary.TotalEmails))
	message.WriteString(fmt.Sprintf("%d success, %d failed", summary.SuccessCount, summary.FailureCount))

	if summary.BlockedCount > 0 {
		message.WriteString(fmt.Sprintf(", %d blocked", summary.BlockedCount))
	}

	if summary.SuppressionCount > 0 {
		message.WriteString(fmt.Sprintf(", %d suppressed", summary.SuppressionCount))
	}

	return message.String()
}

// GetDailyDigestHistory retrieves daily digest delivery history for a user
func (ns *NotificationService) GetDailyDigestHistory(ctx context.Context, userID string, limit int) ([]*DailyDigestDelivery, error) {
	query := `
		SELECT delivery_id, user_id, digest_date, email_sent, sms_sent,
			   email_sent_at, sms_sent_at, event_count, email_count,
			   success_count, failure_count, blocked_count, suppression_count, created_at
		FROM daily_digest_deliveries
		WHERE user_id = ?
		ORDER BY digest_date DESC
		LIMIT ?
	`

	rows, err := ns.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query digest history: %w", err)
	}
	defer rows.Close()

	var deliveries []*DailyDigestDelivery
	for rows.Next() {
		var delivery DailyDigestDelivery
		var digestDateStr string
		var emailSentAt, smsSentAt sql.NullTime

		err := rows.Scan(
			&delivery.DeliveryID,
			&delivery.UserID,
			&digestDateStr,
			&delivery.EmailSent,
			&delivery.SMSSent,
			&emailSentAt,
			&smsSentAt,
			&delivery.EventCount,
			&delivery.EmailCount,
			&delivery.SuccessCount,
			&delivery.FailureCount,
			&delivery.BlockedCount,
			&delivery.SuppressionCount,
			&delivery.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan digest delivery: %v", err)
			continue
		}

		// Parse digest date - handle both date-only and datetime formats
		var digestDate time.Time
		if strings.Contains(digestDateStr, "T") {
			// Handle datetime format (e.g., "2025-08-11T00:00:00Z")
			digestDate, err = time.Parse(time.RFC3339, digestDateStr)
			if err != nil {
				log.Printf("Failed to parse digest date %s: %v", digestDateStr, err)
				continue
			}
		} else {
			// Handle date-only format (e.g., "2025-08-11")
			digestDate, err = time.Parse("2006-01-02", digestDateStr)
			if err != nil {
				log.Printf("Failed to parse digest date %s: %v", digestDateStr, err)
				continue
			}
		}
		delivery.DigestDate = digestDate

		// Handle nullable timestamps
		if emailSentAt.Valid {
			delivery.EmailSentAt = &emailSentAt.Time
		}
		if smsSentAt.Valid {
			delivery.SMSSentAt = &smsSentAt.Time
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

// GetUsersWithDailyDigestEnabled returns all users who have daily digest enabled
func (ns *NotificationService) GetUsersWithDailyDigestEnabled(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT user_id
		FROM notification_preferences
		WHERE delivery_frequency = 'daily_digest'
	`

	rows, err := ns.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users with daily digest: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			log.Printf("Failed to scan user ID: %v", err)
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// DailyDigestDelivery represents a daily digest delivery record
type DailyDigestDelivery struct {
	DeliveryID       string     `json:"delivery_id"`
	UserID           string     `json:"user_id"`
	DigestDate       time.Time  `json:"digest_date"`
	EmailSent        bool       `json:"email_sent"`
	SMSSent          bool       `json:"sms_sent"`
	EmailSentAt      *time.Time `json:"email_sent_at,omitempty"`
	SMSSentAt        *time.Time `json:"sms_sent_at,omitempty"`
	EventCount       int        `json:"event_count"`
	EmailCount       int        `json:"email_count"`
	SuccessCount     int        `json:"success_count"`
	FailureCount     int        `json:"failure_count"`
	BlockedCount     int        `json:"blocked_count"`
	SuppressionCount int        `json:"suppression_count"`
	CreatedAt        time.Time  `json:"created_at"`
}

// DailyDigestContent represents digest content for a specific email
type DailyDigestContent struct {
	ContentID        string     `json:"content_id"`
	DeliveryID       string     `json:"delivery_id"`
	EmailID          string     `json:"email_id"`
	EmailSubject     string     `json:"email_subject"`
	Recipient        string     `json:"recipient"`
	SuccessCount     int        `json:"success_count"`
	FailureCount     int        `json:"failure_count"`
	BlockedCount     int        `json:"blocked_count"`
	LastAccessAt     *time.Time `json:"last_access_at,omitempty"`
	LastIPAddress    string     `json:"last_ip_address,omitempty"`
	LastDeviceType   string     `json:"last_device_type,omitempty"`
	LastCountry      string     `json:"last_country,omitempty"`
	LastCity         string     `json:"last_city,omitempty"`
	SuppressionCount int        `json:"suppression_count"`
	CreatedAt        time.Time  `json:"created_at"`
}

// DigestEmailSummary represents a summary of email activity for digest
type DigestEmailSummary struct {
	EmailID          string     `json:"email_id"`
	EmailSubject     string     `json:"email_subject"`
	Recipient        string     `json:"recipient"`
	SuccessCount     int        `json:"success_count"`
	FailureCount     int        `json:"failure_count"`
	BlockedCount     int        `json:"blocked_count"`
	LastAccessAt     *time.Time `json:"last_access_at,omitempty"`
	LastIPAddress    string     `json:"last_ip_address,omitempty"`
	LastDeviceType   string     `json:"last_device_type,omitempty"`
	LastCountry      string     `json:"last_country,omitempty"`
	LastCity         string     `json:"last_city,omitempty"`
	SuppressionCount int        `json:"suppression_count"`
}

// DigestSummary represents the complete daily digest summary
type DigestSummary struct {
	UserID           string                `json:"user_id"`
	DigestDate       time.Time             `json:"digest_date"`
	TotalEvents      int                   `json:"total_events"`
	TotalEmails      int                   `json:"total_emails"`
	SuccessCount     int                   `json:"success_count"`
	FailureCount     int                   `json:"failure_count"`
	BlockedCount     int                   `json:"blocked_count"`
	SuppressionCount int                   `json:"suppression_count"`
	EmailSummaries   []*DigestEmailSummary `json:"email_summaries"`
}

// DetectDeviceType attempts to detect device type from user agent
func DetectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "Mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "Tablet"
	}
	if strings.Contains(ua, "desktop") || strings.Contains(ua, "windows") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "linux") {
		return "Desktop"
	}

	return "Unknown"
}
