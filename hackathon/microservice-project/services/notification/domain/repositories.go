package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)

	// GetByUserID retrieves notifications for a user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)

	// GetUnreadByUserID retrieves unread notifications for a user
	GetUnreadByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)

	// GetPendingNotifications retrieves pending notifications for sending
	GetPendingNotifications(ctx context.Context, limit int) ([]*Notification, error)

	// UpdateStatus updates the status of a notification
	UpdateStatus(ctx context.Context, id uuid.UUID, status NotificationStatus) error

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, id uuid.UUID, readAt time.Time) error

	// MarkAllAsRead marks all notifications for a user as read
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, readAt time.Time) error

	// Delete deletes a notification
	Delete(ctx context.Context, id uuid.UUID) error

	// GetCount gets the total count of notifications for a user
	GetCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetUnreadCount gets the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

// NotificationTemplateRepository defines the interface for notification template operations
type NotificationTemplateRepository interface {
	// Create creates a new notification template
	Create(ctx context.Context, template *NotificationTemplate) error

	// GetByID retrieves a template by ID
	GetByID(ctx context.Context, id uuid.UUID) (*NotificationTemplate, error)

	// GetByName retrieves a template by name
	GetByName(ctx context.Context, name string) (*NotificationTemplate, error)

	// GetByType retrieves templates by type
	GetByType(ctx context.Context, notificationType NotificationType) ([]*NotificationTemplate, error)

	// GetActive retrieves all active templates
	GetActive(ctx context.Context) ([]*NotificationTemplate, error)

	// Update updates a notification template
	Update(ctx context.Context, template *NotificationTemplate) error

	// Delete deletes a notification template
	Delete(ctx context.Context, id uuid.UUID) error
}

// NotificationPreferenceRepository defines the interface for notification preference operations
type NotificationPreferenceRepository interface {
	// Create creates a new notification preference
	Create(ctx context.Context, preference *NotificationPreference) error

	// GetByUserID retrieves notification preferences for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)

	// GetByUserIDAndType retrieves a specific preference for a user and type
	GetByUserIDAndType(ctx context.Context, userID uuid.UUID, notificationType NotificationType) (*NotificationPreference, error)

	// Update updates a notification preference
	Update(ctx context.Context, preference *NotificationPreference) error

	// Delete deletes a notification preference
	Delete(ctx context.Context, id uuid.UUID) error

	// CreateDefaultPreferences creates default preferences for a new user
	CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error
}

// RepositoryManager defines the interface for managing all repositories
type RepositoryManager interface {
	// Notification returns the notification repository
	Notification() NotificationRepository

	// Template returns the notification template repository
	Template() NotificationTemplateRepository

	// Preference returns the notification preference repository
	Preference() NotificationPreferenceRepository

	// BeginTx begins a database transaction
	BeginTx(ctx context.Context) (RepositoryManager, error)

	// Commit commits the current transaction
	Commit() error

	// Rollback rolls back the current transaction
	Rollback() error
}
