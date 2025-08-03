package usecases

import (
	"context"
	"io"
	"mime/multipart"
	"time"

	fileDomain "github.com/elotusteam/microservice-project/services/file/domain"
	"github.com/google/uuid"
)

// FileService defines the interface for file management operations
type FileService interface {
	UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error)
	GetFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*fileDomain.File, error)
	GetFileContent(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error
	ListFiles(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error)
	SearchFiles(ctx context.Context, userID uuid.UUID, req *SearchFilesRequest) (*SearchFilesResponse, error)
	GetFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*FileMetadata, error)
	UpdateFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, metadata map[string]interface{}) error
	GetUserStorageStats(ctx context.Context, userID uuid.UUID) (*StorageStats, error)
}

// ChunkedUploadService defines the interface for chunked upload operations
type ChunkedUploadService interface {
	InitiateUpload(ctx context.Context, req *InitiateUploadRequest) (*InitiateUploadResponse, error)
	UploadChunk(ctx context.Context, req *UploadChunkRequest) (*UploadChunkResponse, error)
	CompleteUpload(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*CompleteUploadResponse, error)
	CancelUpload(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error
	GetUploadStatus(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*UploadStatusResponse, error)
	ListUploadSessions(ctx context.Context, userID uuid.UUID) ([]*fileDomain.UploadSession, error)
}

// FileShareService defines the interface for file sharing operations
type FileShareService interface {
	ShareFile(ctx context.Context, req *ShareFileRequest) (*ShareFileResponse, error)
	GetSharedFile(ctx context.Context, token string) (*SharedFileResponse, error)
	RevokeShare(ctx context.Context, shareID uuid.UUID, userID uuid.UUID) error
	ListSharedFiles(ctx context.Context, userID uuid.UUID) ([]*fileDomain.FileShare, error)
	ListFilesSharedWithMe(ctx context.Context, userID uuid.UUID) ([]*fileDomain.FileShare, error)
	UpdateSharePermissions(ctx context.Context, shareID uuid.UUID, userID uuid.UUID, permissions []string) error
}

// FileVersionService defines the interface for file versioning operations
type FileVersionService interface {
	CreateVersion(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, content io.Reader) (*fileDomain.FileVersion, error)
	GetVersions(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) ([]*fileDomain.FileVersion, error)
	GetVersion(ctx context.Context, versionID uuid.UUID, userID uuid.UUID) (*fileDomain.FileVersion, error)
	GetVersionContent(ctx context.Context, versionID uuid.UUID, userID uuid.UUID) (io.ReadCloser, error)
	DeleteVersion(ctx context.Context, versionID uuid.UUID, userID uuid.UUID) error
	RestoreVersion(ctx context.Context, versionID uuid.UUID, userID uuid.UUID) error
}

// StorageService defines the interface for storage operations
type StorageService interface {
	Store(ctx context.Context, path string, content io.Reader, contentType string) error
	Retrieve(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
	Copy(ctx context.Context, srcPath, destPath string) error
	Move(ctx context.Context, srcPath, destPath string) error
	GetSize(ctx context.Context, path string) (int64, error)
}

// PermissionService defines the interface for permission checking
type PermissionService interface {
	CheckFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) (bool, error)
	GrantFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) error
	RevokeFilePermission(ctx context.Context, userID, fileID uuid.UUID, permission string) error
	ListFilePermissions(ctx context.Context, userID, fileID uuid.UUID) ([]string, error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendFileUploadedNotification(ctx context.Context, userID uuid.UUID, filename string) error
	SendFileSharedNotification(ctx context.Context, sharedWith uuid.UUID, filename string, sharedBy string) error
	SendStorageQuotaNotification(ctx context.Context, userID uuid.UUID, usedSpace, totalSpace int64) error
}

// ActivityService defines the interface for logging activities
type ActivityService interface {
	LogActivity(ctx context.Context, userID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) error
}

// Request/Response DTOs
type UploadFileRequest struct {
	UserID     uuid.UUID
	File       multipart.File
	Header     *multipart.FileHeader
	Metadata   map[string]interface{}
	Overwrite  bool
	Versioning bool
}

type UploadFileResponse struct {
	File     *fileDomain.File `json:"file"`
	URL      string           `json:"url"`
	Checksum string           `json:"checksum"`
}

type ListFilesRequest struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
	Filter   string `json:"filter"`
}

type ListFilesResponse struct {
	Files   []*fileDomain.File `json:"files"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasMore bool               `json:"has_more"`
}

type SearchFilesRequest struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	MimeType string `json:"mime_type,omitempty"`
}

type SearchFilesResponse struct {
	Files   []*fileDomain.File `json:"files"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasMore bool               `json:"has_more"`
}

type FileMetadata struct {
	ID           uuid.UUID              `json:"id"`
	Filename     string                 `json:"filename"`
	OriginalName string                 `json:"original_name"`
	MimeType     string                 `json:"mime_type"`
	Size         int64                  `json:"size"`
	Checksum     string                 `json:"checksum"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type StorageStats struct {
	UsedSpace  int64   `json:"used_space"`
	TotalSpace int64   `json:"total_space"`
	FileCount  int64   `json:"file_count"`
	QuotaUsed  float64 `json:"quota_used"`
}

type InitiateUploadRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	Filename  string    `json:"filename"`
	TotalSize int64     `json:"total_size"`
	ChunkSize int64     `json:"chunk_size"`
	MimeType  string    `json:"mime_type"`
}

type InitiateUploadResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	ChunkSize int64     `json:"chunk_size"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UploadChunkRequest struct {
	SessionID   uuid.UUID `json:"session_id"`
	UserID      uuid.UUID `json:"user_id"`
	ChunkNumber int       `json:"chunk_number"`
	ChunkData   []byte    `json:"chunk_data"`
	Checksum    string    `json:"checksum"`
}

type UploadChunkResponse struct {
	ChunkNumber     int     `json:"chunk_number"`
	UploadedSize    int64   `json:"uploaded_size"`
	TotalSize       int64   `json:"total_size"`
	Percentage      float64 `json:"percentage"`
	RemainingChunks int     `json:"remaining_chunks"`
}

type CompleteUploadResponse struct {
	File     *fileDomain.File `json:"file"`
	URL      string           `json:"url"`
	Checksum string           `json:"checksum"`
}

type UploadStatusResponse struct {
	SessionID       uuid.UUID               `json:"session_id"`
	Status          fileDomain.UploadStatus `json:"status"`
	UploadedSize    int64                   `json:"uploaded_size"`
	TotalSize       int64                   `json:"total_size"`
	Percentage      float64                 `json:"percentage"`
	RemainingChunks int                     `json:"remaining_chunks"`
	ExpiresAt       time.Time               `json:"expires_at"`
}

type ShareFileRequest struct {
	FileID      uuid.UUID  `json:"file_id"`
	UserID      uuid.UUID  `json:"user_id"`
	SharedWith  *uuid.UUID `json:"shared_with,omitempty"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Public      bool       `json:"public"`
}

type ShareFileResponse struct {
	Share      *fileDomain.FileShare `json:"share"`
	ShareToken string                `json:"share_token"`
	ShareURL   string                `json:"share_url"`
}

type SharedFileResponse struct {
	File        *fileDomain.File      `json:"file"`
	Share       *fileDomain.FileShare `json:"share"`
	Permissions []string              `json:"permissions"`
	CanDownload bool                  `json:"can_download"`
	CanView     bool                  `json:"can_view"`
}
