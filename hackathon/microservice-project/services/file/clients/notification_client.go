package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NotificationClient implements the NotificationService interface using HTTP calls
type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNotificationClient creates a new notification service HTTP client
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendFileUploadedNotificationRequest represents the request payload
type SendFileUploadedNotificationRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Filename string    `json:"filename"`
}

// SendFileUploadedNotification sends a notification when a file is uploaded
func (c *NotificationClient) SendFileUploadedNotification(ctx context.Context, userID uuid.UUID, filename string) error {
	req := SendFileUploadedNotificationRequest{
		UserID:   userID,
		Filename: filename,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/notifications/file-uploaded", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("send file uploaded notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SendFileSharedNotificationRequest represents the request payload
type SendFileSharedNotificationRequest struct {
	SharedWith uuid.UUID `json:"shared_with"`
	Filename   string    `json:"filename"`
	SharedBy   string    `json:"shared_by"`
}

// SendFileSharedNotification sends a notification when a file is shared
func (c *NotificationClient) SendFileSharedNotification(ctx context.Context, sharedWith uuid.UUID, filename string, sharedBy string) error {
	req := SendFileSharedNotificationRequest{
		SharedWith: sharedWith,
		Filename:   filename,
		SharedBy:   sharedBy,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/notifications/file-shared", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("send file shared notification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SendStorageQuotaNotificationRequest represents the request payload
type SendStorageQuotaNotificationRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	UsedSpace  int64     `json:"used_space"`
	TotalSpace int64     `json:"total_space"`
}

// SendStorageQuotaNotification sends a notification about storage quota
func (c *NotificationClient) SendStorageQuotaNotification(ctx context.Context, userID uuid.UUID, usedSpace, totalSpace int64) error {
	req := SendStorageQuotaNotificationRequest{
		UserID:     userID,
		UsedSpace:  usedSpace,
		TotalSpace: totalSpace,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/notifications/storage-quota", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("send storage quota notification failed with status: %d", resp.StatusCode)
	}

	return nil
}
