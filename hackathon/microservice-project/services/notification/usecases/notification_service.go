package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/notification/domain"
	"github.com/elotusteam/microservice-project/shared/config"
)

// notificationService implements the NotificationService interface
type notificationService struct {
	repoManager         domain.RepositoryManager
	templateService     NotificationTemplateService
	preferenceService   NotificationPreferenceService
	emailService        EmailService
	smsService          SMSService
	pushService         PushService
	activityService     ActivityService
	config              *config.Config
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(
	repoManager domain.RepositoryManager,
	templateService NotificationTemplateService,
	preferenceService NotificationPreferenceService,
	emailService EmailService,
	smsService SMSService,
	pushService PushService,
	activityService ActivityService,
	config *config.Config,
) NotificationService {
	return &notificationService{
		repoManager:         repoManager,
		templateService:     templateService,
		preferenceService:   preferenceService,
		emailService:        emailService,
		smsService:          smsService,
		pushService:         pushService,
		activityService:     activityService,
		config:              config,
	}
}

// SendNotification sends a notification to a user
func (s *notificationService) SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error) {
	// Check if user can receive this type of notification
	canSend, err := s.preferenceService.CanSendNotification(ctx, req.UserID, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to check notification preferences: %w", err)
	}
	
	if !canSend {
		return &SendNotificationResponse{
			Status:  "skipped",
			Message: "User has disabled this type of notification",
		}, nil
	}

	// Create notification entity
	notification := &domain.Notification{
		ID:       uuid.New(),
		UserID:   req.UserID,
		Type:     req.Type,
		Title:    req.Title,
		Message:  req.Message,
		Data:     req.Data,
		Status:   domain.NotificationStatusPending,
		Priority: req.Priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Priority == "" {
		notification.Priority = domain.NotificationPriorityNormal
	}

	if req.ScheduledAt != nil {
		notification.ScheduledAt = req.ScheduledAt
	}

	// Save notification to database
	err = s.repoManager.Notification().Create(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// If not scheduled, send immediately
	if req.ScheduledAt == nil {
		err = s.sendNotificationNow(ctx, notification)
		if err != nil {
			// Update status to failed
			s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusFailed)
			return nil, fmt.Errorf("failed to send notification: %w", err)
		}
	}

	// Log activity
	s.activityService.LogActivity(ctx, req.UserID, "notification_sent", fmt.Sprintf("Notification sent: %s", req.Title))

	return &SendNotificationResponse{
		NotificationID: notification.ID,
		Status:         "sent",
		Message:        "Notification sent successfully",
	}, nil
}

// SendBulkNotifications sends notifications to multiple users
func (s *notificationService) SendBulkNotifications(ctx context.Context, userIDs []uuid.UUID, req *SendNotificationRequest) error {
	for _, userID := range userIDs {
		bulkReq := *req
		bulkReq.UserID = userID
		
		_, err := s.SendNotification(ctx, &bulkReq)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}
	return nil
}

// GetNotifications retrieves notifications for a user
func (s *notificationService) GetNotifications(ctx context.Context, req *GetNotificationsRequest) (*GetNotificationsResponse, error) {
	var notifications []*domain.Notification
	var err error

	if req.Limit == 0 {
		req.Limit = 20
	}

	if req.Unread {
		notifications, err = s.repoManager.Notification().GetUnreadByUserID(ctx, req.UserID, req.Limit, req.Offset)
	} else {
		notifications, err = s.repoManager.Notification().GetByUserID(ctx, req.UserID, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Get total count
	total, err := s.repoManager.Notification().GetCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification count: %w", err)
	}

	// Get unread count
	unreadCount, err := s.repoManager.Notification().GetUnreadCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	return &GetNotificationsResponse{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	}, nil
}

// GetNotificationByID retrieves a specific notification
func (s *notificationService) GetNotificationByID(ctx context.Context, userID, notificationID uuid.UUID) (*domain.Notification, error) {
	notification, err := s.repoManager.Notification().GetByID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Verify ownership
	if notification.UserID != userID {
		return nil, fmt.Errorf("notification not found")
	}

	return notification, nil
}

// MarkAsRead marks notifications as read
func (s *notificationService) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) error {
	readAt := time.Now()

	if req.MarkAll {
		err := s.repoManager.Notification().MarkAllAsRead(ctx, req.UserID, readAt)
		if err != nil {
			return fmt.Errorf("failed to mark all notifications as read: %w", err)
		}
	} else {
		for _, notificationID := range req.NotificationIDs {
			// Verify ownership before marking as read
			notification, err := s.GetNotificationByID(ctx, req.UserID, notificationID)
			if err != nil {
				continue // Skip if not found or not owned
			}

			if notification.Status != domain.NotificationStatusRead {
				err = s.repoManager.Notification().MarkAsRead(ctx, notificationID, readAt)
				if err != nil {
					return fmt.Errorf("failed to mark notification as read: %w", err)
				}
			}
		}
	}

	// Log activity
	s.activityService.LogActivity(ctx, req.UserID, "notifications_read", "Marked notifications as read")

	return nil
}

// DeleteNotification deletes a notification
func (s *notificationService) DeleteNotification(ctx context.Context, userID, notificationID uuid.UUID) error {
	// Verify ownership
	_, err := s.GetNotificationByID(ctx, userID, notificationID)
	if err != nil {
		return err
	}

	err = s.repoManager.Notification().Delete(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	// Log activity
	s.activityService.LogActivity(ctx, userID, "notification_deleted", "Deleted notification")

	return nil
}

// GetUnreadCount gets the count of unread notifications for a user
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := s.repoManager.Notification().GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// ProcessPendingNotifications processes pending notifications for sending
func (s *notificationService) ProcessPendingNotifications(ctx context.Context) error {
	notifications, err := s.repoManager.Notification().GetPendingNotifications(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		// Check if it's time to send scheduled notifications
		if notification.ScheduledAt != nil && notification.ScheduledAt.After(time.Now()) {
			continue
		}

		err = s.sendNotificationNow(ctx, notification)
		if err != nil {
			// Update status to failed
			s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusFailed)
			fmt.Printf("Failed to send notification %s: %v\n", notification.ID, err)
		}
	}

	return nil
}

// sendNotificationNow sends a notification immediately based on its type
func (s *notificationService) sendNotificationNow(ctx context.Context, notification *domain.Notification) error {
	switch notification.Type {
	case domain.NotificationTypeEmail:
		return s.sendEmailNotification(ctx, notification)
	case domain.NotificationTypeSMS:
		return s.sendSMSNotification(ctx, notification)
	case domain.NotificationTypePush:
		return s.sendPushNotification(ctx, notification)
	case domain.NotificationTypeInApp:
		return s.sendInAppNotification(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}
}

// sendEmailNotification sends an email notification
func (s *notificationService) sendEmailNotification(ctx context.Context, notification *domain.Notification) error {
	// For now, we'll use a placeholder email address
	// In a real implementation, you'd get the user's email from the user service
	email := "user@example.com"
	
	err := s.emailService.SendEmail(ctx, email, notification.Title, notification.Message)
	if err != nil {
		return err
	}

	// Update notification status
	now := time.Now()
	notification.Status = domain.NotificationStatusSent
	notification.SentAt = &now
	return s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent)
}

// sendSMSNotification sends an SMS notification
func (s *notificationService) sendSMSNotification(ctx context.Context, notification *domain.Notification) error {
	// For now, we'll use a placeholder phone number
	phone := "+1234567890"
	
	err := s.smsService.SendSMS(ctx, phone, notification.Message)
	if err != nil {
		return err
	}

	// Update notification status
	now := time.Now()
	notification.Status = domain.NotificationStatusSent
	notification.SentAt = &now
	return s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent)
}

// sendPushNotification sends a push notification
func (s *notificationService) sendPushNotification(ctx context.Context, notification *domain.Notification) error {
	// For now, we'll use a placeholder device token
	deviceToken := "device_token_placeholder"
	
	err := s.pushService.SendPushNotification(ctx, deviceToken, notification.Title, notification.Message, notification.Data)
	if err != nil {
		return err
	}

	// Update notification status
	now := time.Now()
	notification.Status = domain.NotificationStatusSent
	notification.SentAt = &now
	return s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent)
}

// sendInAppNotification sends an in-app notification
func (s *notificationService) sendInAppNotification(ctx context.Context, notification *domain.Notification) error {
	// In-app notifications are just stored in the database and marked as sent
	now := time.Now()
	notification.Status = domain.NotificationStatusSent
	notification.SentAt = &now
	return s.repoManager.Notification().UpdateStatus(ctx, notification.ID, domain.NotificationStatusSent)
}