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

// PermissionClient implements the PermissionService interface using HTTP calls
type PermissionClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPermissionClient creates a new permission service HTTP client
func NewPermissionClient(baseURL string) *PermissionClient {
	return &PermissionClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckFilePermissionRequest represents the request payload
type CheckFilePermissionRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	FileID     uuid.UUID `json:"file_id"`
	Permission string    `json:"permission"`
}

// CheckFilePermissionResponse represents the response payload
type CheckFilePermissionResponse struct {
	HasPermission bool `json:"has_permission"`
}

// CheckFilePermission checks if a user has permission for a file
func (c *PermissionClient) CheckFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) (bool, error) {
	req := CheckFilePermissionRequest{
		UserID:     userID,
		FileID:     fileID,
		Permission: permission,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/permissions/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("permission check failed with status: %d", resp.StatusCode)
	}

	var response CheckFilePermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.HasPermission, nil
}

// GrantFilePermissionRequest represents the request payload
type GrantFilePermissionRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	FileID     uuid.UUID `json:"file_id"`
	Permission string    `json:"permission"`
}

// GrantFilePermission grants a permission to a user for a file
func (c *PermissionClient) GrantFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) error {
	req := GrantFilePermissionRequest{
		UserID:     userID,
		FileID:     fileID,
		Permission: permission,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/permissions/grant", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("grant permission failed with status: %d", resp.StatusCode)
	}

	return nil
}

// RevokeFilePermission revokes a permission from a user for a file
func (c *PermissionClient) RevokeFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) error {
	req := GrantFilePermissionRequest{ // Same structure as grant
		UserID:     userID,
		FileID:     fileID,
		Permission: permission,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/permissions/revoke", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke permission failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ListFilePermissionsResponse represents the response payload
type ListFilePermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

// ListFilePermissions lists all permissions for a user and file
func (c *PermissionClient) ListFilePermissions(ctx context.Context, userID, fileID uuid.UUID) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/permissions/list?user_id=%s&file_id=%s", c.baseURL, userID.String(), fileID.String())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list permissions failed with status: %d", resp.StatusCode)
	}

	var response ListFilePermissionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Permissions, nil
}
