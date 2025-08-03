package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string                 `json:"id" db:"id"`
	Username     string                 `json:"username" db:"username"`
	Email        string                 `json:"email" db:"email"`
	PasswordHash string                 `json:"-" db:"password_hash"` // Hidden from JSON
	FirstName    string                 `json:"first_name" db:"first_name"`
	LastName     string                 `json:"last_name" db:"last_name"`
	Role         UserRole               `json:"role" db:"role"`
	Status       UserStatus             `json:"status" db:"status"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time             `json:"last_login_at,omitempty" db:"last_login_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// UserRole represents user roles
type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleModerator UserRole = "moderator"
	UserRoleUser      UserRole = "user"
	UserRoleGuest     UserRole = "guest"
)

// UserStatus represents user status
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// File represents a file in the system
type File struct {
	ID           string                 `json:"id" db:"id"`
	OwnerID      string                 `json:"owner_id" db:"owner_id"`
	Filename     string                 `json:"filename" db:"filename"`
	OriginalName string                 `json:"original_name" db:"original_name"`
	ContentType  string                 `json:"content_type" db:"content_type"`
	Size         int64                  `json:"size" db:"size"`
	Path         string                 `json:"path" db:"path"`
	StoragePath  string                 `json:"storage_path" db:"storage_path"`
	Checksum     string                 `json:"checksum" db:"checksum"`
	Status       FileStatus             `json:"status" db:"status"`
	Visibility   FileVisibility         `json:"visibility" db:"visibility"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Tags         []string               `json:"tags,omitempty" db:"tags"`
}

// FileStatus represents file status
type FileStatus string

const (
	FileStatusUploading  FileStatus = "uploading"
	FileStatusProcessing FileStatus = "processing"
	FileStatusActive     FileStatus = "active"
	FileStatusArchived   FileStatus = "archived"
	FileStatusDeleted    FileStatus = "deleted"
	FileStatusCorrupted  FileStatus = "corrupted"
)

// FileVisibility represents file visibility
type FileVisibility string

const (
	FileVisibilityPrivate FileVisibility = "private"
	FileVisibilityPublic  FileVisibility = "public"
	FileVisibilityShared  FileVisibility = "shared"
)

// UserGroup represents a user group
type UserGroup struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	OwnerID     string                 `json:"owner_id" db:"owner_id"`
	Type        GroupType              `json:"type" db:"type"`
	Status      GroupStatus            `json:"status" db:"status"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// GroupType represents group types
type GroupType string

const (
	GroupTypePublic  GroupType = "public"
	GroupTypePrivate GroupType = "private"
	GroupTypeSystem  GroupType = "system"
)

// GroupStatus represents group status
type GroupStatus string

const (
	GroupStatusActive   GroupStatus = "active"
	GroupStatusInactive GroupStatus = "inactive"
	GroupStatusDeleted  GroupStatus = "deleted"
)

// GroupMember represents a group member
type GroupMember struct {
	ID        string            `json:"id" db:"id"`
	GroupID   string            `json:"group_id" db:"group_id"`
	UserID    string            `json:"user_id" db:"user_id"`
	Role      GroupMemberRole   `json:"role" db:"role"`
	Status    GroupMemberStatus `json:"status" db:"status"`
	JoinedAt  time.Time         `json:"joined_at" db:"joined_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`
}

// GroupMemberRole represents group member roles
type GroupMemberRole string

const (
	GroupMemberRoleOwner     GroupMemberRole = "owner"
	GroupMemberRoleAdmin     GroupMemberRole = "admin"
	GroupMemberRoleModerator GroupMemberRole = "moderator"
	GroupMemberRoleMember    GroupMemberRole = "member"
	GroupMemberRoleViewer    GroupMemberRole = "viewer"
)

// GroupMemberStatus represents group member status
type GroupMemberStatus string

const (
	GroupMemberStatusActive   GroupMemberStatus = "active"
	GroupMemberStatusInactive GroupMemberStatus = "inactive"
	GroupMemberStatusPending  GroupMemberStatus = "pending"
	GroupMemberStatusBanned   GroupMemberStatus = "banned"
)

// FilePermission represents file permissions
type FilePermission struct {
	ID         string                 `json:"id" db:"id"`
	FileID     string                 `json:"file_id" db:"file_id"`
	UserID     *string                `json:"user_id,omitempty" db:"user_id"`
	GroupID    *string                `json:"group_id,omitempty" db:"group_id"`
	Permission PermissionType         `json:"permission" db:"permission"`
	GrantedBy  string                 `json:"granted_by" db:"granted_by"`
	GrantedAt  time.Time              `json:"granted_at" db:"granted_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	Status     PermissionStatus       `json:"status" db:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// PermissionType represents permission types
type PermissionType string

const (
	PermissionTypeRead   PermissionType = "read"
	PermissionTypeWrite  PermissionType = "write"
	PermissionTypeDelete PermissionType = "delete"
	PermissionTypeShare  PermissionType = "share"
	PermissionTypeAdmin  PermissionType = "admin"
)

// PermissionStatus represents permission status
type PermissionStatus string

const (
	PermissionStatusActive  PermissionStatus = "active"
	PermissionStatusRevoked PermissionStatus = "revoked"
	PermissionStatusExpired PermissionStatus = "expired"
	PermissionStatusPending PermissionStatus = "pending"
)

// ImageVariant represents image variants/thumbnails
type ImageVariant struct {
	ID          string             `json:"id" db:"id"`
	FileID      string             `json:"file_id" db:"file_id"`
	VariantType string             `json:"variant_type" db:"variant_type"` // thumbnail, small, medium, large, etc.
	Width       int                `json:"width" db:"width"`
	Height      int                `json:"height" db:"height"`
	Size        int64              `json:"size" db:"size"`
	Path        string             `json:"path" db:"path"`
	StoragePath string             `json:"storage_path" db:"storage_path"`
	Format      string             `json:"format" db:"format"`
	Quality     int                `json:"quality" db:"quality"`
	Status      ImageVariantStatus `json:"status" db:"status"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// ImageVariantStatus represents image variant status
type ImageVariantStatus string

const (
	ImageVariantStatusProcessing ImageVariantStatus = "processing"
	ImageVariantStatusReady      ImageVariantStatus = "ready"
	ImageVariantStatusFailed     ImageVariantStatus = "failed"
	ImageVariantStatusDeleted    ImageVariantStatus = "deleted"
)

// ActivityLog represents system activity logs
type ActivityLog struct {
	ID           string                 `json:"id" db:"id"`
	UserID       *string                `json:"user_id,omitempty" db:"user_id"`
	Action       string                 `json:"action" db:"action"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty" db:"resource_id"`
	Details      map[string]interface{} `json:"details,omitempty" db:"details"`
	IPAddress    string                 `json:"ip_address" db:"ip_address"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	Status       ActivityStatus         `json:"status" db:"status"`
}

// ActivityStatus represents activity status
type ActivityStatus string

const (
	ActivityStatusSuccess ActivityStatus = "success"
	ActivityStatusFailure ActivityStatus = "failure"
	ActivityStatusPending ActivityStatus = "pending"
)

// RevokedToken represents revoked JWT tokens
type RevokedToken struct {
	ID        string    `json:"id" db:"id"`
	TokenID   string    `json:"token_id" db:"token_id"` // JTI claim
	UserID    string    `json:"user_id" db:"user_id"`
	Reason    string    `json:"reason" db:"reason"`
	RevokedAt time.Time `json:"revoked_at" db:"revoked_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	Used      bool       `json:"used" db:"used"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
}

// Session represents user sessions
type Session struct {
	ID         string        `json:"id" db:"id"`
	UserID     string        `json:"user_id" db:"user_id"`
	TokenID    string        `json:"token_id" db:"token_id"`
	IPAddress  string        `json:"ip_address" db:"ip_address"`
	UserAgent  string        `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time     `json:"expires_at" db:"expires_at"`
	LastUsedAt time.Time     `json:"last_used_at" db:"last_used_at"`
	Status     SessionStatus `json:"status" db:"status"`
}

// SessionStatus represents session status
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

// UploadSession represents file upload sessions for chunked uploads
type UploadSession struct {
	ID             string                 `json:"id" db:"id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	Filename       string                 `json:"filename" db:"filename"`
	ContentType    string                 `json:"content_type" db:"content_type"`
	TotalSize      int64                  `json:"total_size" db:"total_size"`
	ChunkSize      int64                  `json:"chunk_size" db:"chunk_size"`
	TotalChunks    int                    `json:"total_chunks" db:"total_chunks"`
	UploadedChunks []int                  `json:"uploaded_chunks" db:"uploaded_chunks"`
	Status         UploadSessionStatus    `json:"status" db:"status"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	ExpiresAt      time.Time              `json:"expires_at" db:"expires_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// UploadSessionStatus represents upload session status
type UploadSessionStatus string

const (
	UploadSessionStatusActive    UploadSessionStatus = "active"
	UploadSessionStatusCompleted UploadSessionStatus = "completed"
	UploadSessionStatusFailed    UploadSessionStatus = "failed"
	UploadSessionStatusExpired   UploadSessionStatus = "expired"
	UploadSessionStatusCancelled UploadSessionStatus = "cancelled"
)

// Notification represents system notifications
type Notification struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Type      NotificationType       `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Message   string                 `json:"message" db:"message"`
	Data      map[string]interface{} `json:"data,omitempty" db:"data"`
	Status    NotificationStatus     `json:"status" db:"status"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty" db:"read_at"`
}

// NotificationType represents notification types
type NotificationType string

const (
	NotificationTypeFileUploaded      NotificationType = "file_uploaded"
	NotificationTypeFileShared        NotificationType = "file_shared"
	NotificationTypeFileDeleted       NotificationType = "file_deleted"
	NotificationTypePermissionGranted NotificationType = "permission_granted"
	NotificationTypePermissionRevoked NotificationType = "permission_revoked"
	NotificationTypeGroupInvite       NotificationType = "group_invite"
	NotificationTypeSystemAlert       NotificationType = "system_alert"
	NotificationTypeSecurityAlert     NotificationType = "security_alert"
)

// NotificationStatus represents notification status
type NotificationStatus string

const (
	NotificationStatusUnread  NotificationStatus = "unread"
	NotificationStatusRead    NotificationStatus = "read"
	NotificationStatusDeleted NotificationStatus = "deleted"
)

// APIKey represents API keys for external access
type APIKey struct {
	ID          string       `json:"id" db:"id"`
	UserID      string       `json:"user_id" db:"user_id"`
	Name        string       `json:"name" db:"name"`
	KeyHash     string       `json:"-" db:"key_hash"` // Hidden from JSON
	Prefix      string       `json:"prefix" db:"prefix"`
	Permissions []string     `json:"permissions" db:"permissions"`
	Status      APIKeyStatus `json:"status" db:"status"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	LastUsedAt  *time.Time   `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty" db:"expires_at"`
}

// APIKeyStatus represents API key status
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusInactive APIKeyStatus = "inactive"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
	APIKeyStatusExpired  APIKeyStatus = "expired"
)

// SystemConfig represents system configuration
type SystemConfig struct {
	ID          string    `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Type        string    `json:"type" db:"type"` // string, int, bool, json
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// HealthCheck represents health check results
type HealthCheck struct {
	Service   string                 `json:"service"`
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

// HealthStatus represents health check status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Metrics represents system metrics
type Metrics struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Event represents domain events
type Event struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	Subject       string                 `json:"subject"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       string                 `json:"version"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// Common event types
const (
	EventTypeUserCreated              = "user.created"
	EventTypeUserUpdated              = "user.updated"
	EventTypeUserDeleted              = "user.deleted"
	EventTypeFileUploaded             = "file.uploaded"
	EventTypeFileDeleted              = "file.deleted"
	EventTypeFileShared               = "file.shared"
	EventTypePermissionGranted        = "permission.granted"
	EventTypePermissionRevoked        = "permission.revoked"
	EventTypeGroupCreated             = "group.created"
	EventTypeGroupMemberAdded         = "group.member.added"
	EventTypeGroupMemberRemoved       = "group.member.removed"
	EventTypeImageProcessingStarted   = "image.processing.started"
	EventTypeImageProcessingCompleted = "image.processing.completed"
	EventTypeImageProcessingFailed    = "image.processing.failed"
)

// Error represents domain errors
type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrorCodeValidation              = "VALIDATION_ERROR"
	ErrorCodeNotFound                = "NOT_FOUND"
	ErrorCodeUnauthorized            = "UNAUTHORIZED"
	ErrorCodeForbidden               = "FORBIDDEN"
	ErrorCodeConflict                = "CONFLICT"
	ErrorCodeInternalError           = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable      = "SERVICE_UNAVAILABLE"
	ErrorCodeRateLimitExceeded       = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInvalidToken            = "INVALID_TOKEN"
	ErrorCodeExpiredToken            = "EXPIRED_TOKEN"
	ErrorCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeFileTooLarge            = "FILE_TOO_LARGE"
	ErrorCodeUnsupportedFileType     = "UNSUPPORTED_FILE_TYPE"
	ErrorCodeStorageQuotaExceeded    = "STORAGE_QUOTA_EXCEEDED"
)
