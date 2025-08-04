package domain

import (
	"github.com/google/uuid"
	"time"
)

// File represents a file entity in the system
type File struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       uuid.UUID              `json:"user_id" db:"user_id"`
	Filename     string                 `json:"filename" db:"filename"`
	OriginalName string                 `json:"original_name" db:"original_name"`
	MimeType     string                 `json:"mime_type" db:"mime_type"`
	Size         int64                  `json:"size" db:"size"`
	Path         string                 `json:"path" db:"path"`
	URL          string                 `json:"url" db:"url"`
	Checksum     string                 `json:"checksum" db:"checksum"`
	Status       FileStatus             `json:"status" db:"status"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// FileStatus represents the status of a file
type FileStatus string

const (
	FileStatusUploading  FileStatus = "uploading"
	FileStatusProcessing FileStatus = "processing"
	FileStatusReady      FileStatus = "ready"
	FileStatusError      FileStatus = "error"
	FileStatusDeleted    FileStatus = "deleted"
)

// UploadSession represents an upload session for chunked uploads
type UploadSession struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	UserID       uuid.UUID    `json:"user_id" db:"user_id"`
	Filename     string       `json:"filename" db:"filename"`
	TotalSize    int64        `json:"total_size" db:"total_size"`
	ChunkSize    int64        `json:"chunk_size" db:"chunk_size"`
	UploadedSize int64        `json:"uploaded_size" db:"uploaded_size"`
	Status       UploadStatus `json:"status" db:"status"`
	ExpiresAt    time.Time    `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// UploadStatus represents the status of an upload session
type UploadStatus string

const (
	UploadStatusActive    UploadStatus = "active"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusExpired   UploadStatus = "expired"
	UploadStatusCancelled UploadStatus = "cancelled"
)

// FileShare represents file sharing permissions
type FileShare struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	FileID      uuid.UUID  `json:"file_id" db:"file_id"`
	SharedBy    uuid.UUID  `json:"shared_by" db:"shared_by"`
	SharedWith  *uuid.UUID `json:"shared_with,omitempty" db:"shared_with"`
	ShareToken  string     `json:"share_token" db:"share_token"`
	Permissions []string   `json:"permissions" db:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// FileVersion represents file versioning
type FileVersion struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FileID    uuid.UUID `json:"file_id" db:"file_id"`
	Version   int       `json:"version" db:"version"`
	Path      string    `json:"path" db:"path"`
	Size      int64     `json:"size" db:"size"`
	Checksum  string    `json:"checksum" db:"checksum"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ImageVariant represents different image sizes and qualities
type ImageVariant struct {
	ID          string             `json:"id" db:"id"`
	FileID      uuid.UUID          `json:"file_id" db:"file_id"`
	VariantType string             `json:"variant_type" db:"variant_type"`
	Width       int                `json:"width" db:"width"`
	Height      int                `json:"height" db:"height"`
	Size        int64              `json:"size" db:"size"`
	Path        string             `json:"path" db:"path"`
	Format      string             `json:"format" db:"format"`
	Quality     int                `json:"quality" db:"quality"`
	Status      ImageVariantStatus `json:"status" db:"status"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// ImageVariantStatus represents the status of an image variant
type ImageVariantStatus string

const (
	ImageVariantStatusProcessing ImageVariantStatus = "processing"
	ImageVariantStatusReady      ImageVariantStatus = "ready"
	ImageVariantStatusError      ImageVariantStatus = "error"
)
