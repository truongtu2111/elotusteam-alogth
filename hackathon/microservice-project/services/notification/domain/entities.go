package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
	NotificationTypeInApp NotificationType = "in_app"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRead    NotificationStatus = "read"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// Notification represents a notification entity
type Notification struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      uuid.UUID              `json:"user_id" db:"user_id"`
	Type        NotificationType       `json:"type" db:"type"`
	Title       string                 `json:"title" db:"title"`
	Message     string                 `json:"message" db:"message"`
	Data        map[string]interface{} `json:"data" db:"data"`
	Status      NotificationStatus     `json:"status" db:"status"`
	Priority    NotificationPriority   `json:"priority" db:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at" db:"scheduled_at"`
	SentAt      *time.Time             `json:"sent_at" db:"sent_at"`
	ReadAt      *time.Time             `json:"read_at" db:"read_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	Name      string           `json:"name" db:"name"`
	Type      NotificationType `json:"type" db:"type"`
	Subject   string           `json:"subject" db:"subject"`
	Body      string           `json:"body" db:"body"`
	Variables []string         `json:"variables" db:"variables"`
	IsActive  bool             `json:"is_active" db:"is_active"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	UserID          uuid.UUID        `json:"user_id" db:"user_id"`
	Type            NotificationType `json:"type" db:"type"`
	Enabled         bool             `json:"enabled" db:"enabled"`
	QuietHoursStart *time.Time       `json:"quiet_hours_start" db:"quiet_hours_start"`
	QuietHoursEnd   *time.Time       `json:"quiet_hours_end" db:"quiet_hours_end"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}
