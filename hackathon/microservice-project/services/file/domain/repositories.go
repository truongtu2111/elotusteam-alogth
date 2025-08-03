package domain

import (
	"context"
	"github.com/google/uuid"
)

// FileRepository defines the interface for file data operations
type FileRepository interface {
	Create(ctx context.Context, file *File) error
	GetByID(ctx context.Context, id uuid.UUID) (*File, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*File, error)
	Update(ctx context.Context, file *File) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByChecksum(ctx context.Context, checksum string) (*File, error)
	Search(ctx context.Context, query string, userID uuid.UUID, limit, offset int) ([]*File, error)
	GetTotalSize(ctx context.Context, userID uuid.UUID) (int64, error)
	GetFileCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

// UploadSessionRepository defines the interface for upload session operations
type UploadSessionRepository interface {
	Create(ctx context.Context, session *UploadSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*UploadSession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*UploadSession, error)
	Update(ctx context.Context, session *UploadSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*UploadSession, error)
}

// FileShareRepository defines the interface for file sharing operations
type FileShareRepository interface {
	Create(ctx context.Context, share *FileShare) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileShare, error)
	GetByToken(ctx context.Context, token string) (*FileShare, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileShare, error)
	GetBySharedBy(ctx context.Context, userID uuid.UUID) ([]*FileShare, error)
	GetBySharedWith(ctx context.Context, userID uuid.UUID) ([]*FileShare, error)
	Update(ctx context.Context, share *FileShare) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// FileVersionRepository defines the interface for file version operations
type FileVersionRepository interface {
	Create(ctx context.Context, version *FileVersion) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileVersion, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileVersion, error)
	GetLatestByFileID(ctx context.Context, fileID uuid.UUID) (*FileVersion, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
}

// RepositoryManager aggregates all file-related repositories
type RepositoryManager interface {
	File() FileRepository
	UploadSession() UploadSessionRepository
	FileShare() FileShareRepository
	FileVersion() FileVersionRepository
	BeginTx(ctx context.Context) (RepositoryManager, error)
	Commit() error
	Rollback() error
}
