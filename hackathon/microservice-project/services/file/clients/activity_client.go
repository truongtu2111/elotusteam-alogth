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

// ActivityClient implements the ActivityService interface using HTTP calls
type ActivityClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewActivityClient creates a new activity service HTTP client
func NewActivityClient(baseURL string) *ActivityClient {
	return &ActivityClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LogActivityRequest represents the request payload for logging activity
type LogActivityRequest struct {
	UserID       uuid.UUID              `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// LogActivity logs an activity to the analytics service
func (c *ActivityClient) LogActivity(ctx context.Context, userID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) error {
	req := LogActivityRequest{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Timestamp:    time.Now().UTC(),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/activities/log", bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("log activity failed with status: %d", resp.StatusCode)
	}

	return nil
}
