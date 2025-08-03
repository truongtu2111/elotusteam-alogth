package usecases

import (
	"context"
	"time"

	"github.com/elotusteam/microservice-project/services/notification/domain"
	"github.com/google/uuid"
)

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	UserID      uuid.UUID                   `json:"user_id" validate:"required"`
	Type        domain.NotificationType     `json:"type" validate:"required"`
	Title       string                      `json:"title" validate:"required,max=255"`
	Message     string                      `json:"message" validate:"required"`
	Data        map[string]interface{}      `json:"data,omitempty"`
	Priority    domain.NotificationPriority `json:"priority,omitempty"`
	ScheduledAt *time.Time                  `json:"scheduled_at,omitempty"`
}

// SendNotificationResponse represents the response after sending a notification
type SendNotificationResponse struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
}

// GetNotificationsRequest represents a request to get notifications
type GetNotificationsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Unread bool      `json:"unread,omitempty"`
	Limit  int       `json:"limit,omitempty" validate:"min=1,max=100"`
	Offset int       `json:"offset,omitempty" validate:"min=0"`
}

// GetNotificationsResponse represents the response with notifications
type GetNotificationsResponse struct {
	Notifications []*domain.Notification `json:"notifications"`
	Total         int64                  `json:"total"`
	UnreadCount   int64                  `json:"unread_count"`
}

// MarkAsReadRequest represents a request to mark notifications as read
type MarkAsReadRequest struct {
	UserID          uuid.UUID   `json:"user_id" validate:"required"`
	NotificationIDs []uuid.UUID `json:"notification_ids,omitempty"`
	MarkAll         bool        `json:"mark_all,omitempty"`
}

// UpdatePreferencesRequest represents a request to update notification preferences
type UpdatePreferencesRequest struct {
	UserID      uuid.UUID                        `json:"user_id" validate:"required"`
	Preferences []*domain.NotificationPreference `json:"preferences" validate:"required"`
}

// CreateTemplateRequest represents a request to create a notification template
type CreateTemplateRequest struct {
	Name      string                  `json:"name" validate:"required,max=255"`
	Type      domain.NotificationType `json:"type" validate:"required"`
	Subject   string                  `json:"subject" validate:"required,max=255"`
	Body      string                  `json:"body" validate:"required"`
	Variables []string                `json:"variables,omitempty"`
}

// NotificationService defines the interface for notification business logic
type NotificationService interface {
	// SendNotification sends a notification to a user
	SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error)

	// SendBulkNotifications sends notifications to multiple users
	SendBulkNotifications(ctx context.Context, userIDs []uuid.UUID, req *SendNotificationRequest) error

	// GetNotifications retrieves notifications for a user
	GetNotifications(ctx context.Context, req *GetNotificationsRequest) (*GetNotificationsResponse, error)

	// GetNotificationByID retrieves a specific notification
	GetNotificationByID(ctx context.Context, userID, notificationID uuid.UUID) (*domain.Notification, error)

	// MarkAsRead marks notifications as read
	MarkAsRead(ctx context.Context, req *MarkAsReadRequest) error

	// DeleteNotification deletes a notification
	DeleteNotification(ctx context.Context, userID, notificationID uuid.UUID) error

	// GetUnreadCount gets the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// ProcessPendingNotifications processes pending notifications for sending
	ProcessPendingNotifications(ctx context.Context) error
}

// NotificationTemplateService defines the interface for template management
type NotificationTemplateService interface {
	// CreateTemplate creates a new notification template
	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*domain.NotificationTemplate, error)

	// GetTemplate retrieves a template by ID
	GetTemplate(ctx context.Context, id uuid.UUID) (*domain.NotificationTemplate, error)

	// GetTemplateByName retrieves a template by name
	GetTemplateByName(ctx context.Context, name string) (*domain.NotificationTemplate, error)

	// UpdateTemplate updates a notification template
	UpdateTemplate(ctx context.Context, template *domain.NotificationTemplate) error

	// DeleteTemplate deletes a notification template
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// RenderTemplate renders a template with provided data
	RenderTemplate(ctx context.Context, templateName string, data map[string]interface{}) (string, string, error)
}

// NotificationPreferenceService defines the interface for preference management
type NotificationPreferenceService interface {
	// GetPreferences retrieves notification preferences for a user
	GetPreferences(ctx context.Context, userID uuid.UUID) ([]*domain.NotificationPreference, error)

	// UpdatePreferences updates notification preferences for a user
	UpdatePreferences(ctx context.Context, req *UpdatePreferencesRequest) error

	// CreateDefaultPreferences creates default preferences for a new user
	CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error

	// CanSendNotification checks if a notification can be sent based on user preferences
	CanSendNotification(ctx context.Context, userID uuid.UUID, notificationType domain.NotificationType) (bool, error)
}

// EmailService defines the interface for email operations
type EmailService interface {
	// SendEmail sends an email
	SendEmail(ctx context.Context, to, subject, body string) error

	// SendBulkEmail sends emails to multiple recipients
	SendBulkEmail(ctx context.Context, recipients []string, subject, body string) error
}

// SMSService defines the interface for SMS operations
type SMSService interface {
	// SendSMS sends an SMS
	SendSMS(ctx context.Context, to, message string) error

	// SendBulkSMS sends SMS to multiple recipients
	SendBulkSMS(ctx context.Context, recipients []string, message string) error
}

// PushService defines the interface for push notification operations
type PushService interface {
	// SendPushNotification sends a push notification
	SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) error

	// SendBulkPushNotification sends push notifications to multiple devices
	SendBulkPushNotification(ctx context.Context, deviceTokens []string, title, body string, data map[string]interface{}) error
}

// ActivityService defines the interface for activity logging
type ActivityService interface {
	// LogActivity logs a notification activity
	LogActivity(ctx context.Context, userID uuid.UUID, action, details string) error
}
