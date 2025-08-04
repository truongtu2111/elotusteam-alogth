package usecases

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	fileDomain "github.com/elotusteam/microservice-project/services/file/domain"
	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/elotusteam/microservice-project/shared/utils"
	"github.com/google/uuid"
)

// fileService implements the FileService interface
type fileService struct {
	repoManager            fileDomain.RepositoryManager
	storageService         StorageService
	permissionService      PermissionService
	notificationService    NotificationService
	activityService        ActivityService
	imageProcessingService ImageProcessingService
	config                 *config.Config
}

// NewFileService creates a new file service instance
func NewFileService(
	repoManager fileDomain.RepositoryManager,
	storageService StorageService,
	permissionService PermissionService,
	notificationService NotificationService,
	activityService ActivityService,
	imageProcessingService ImageProcessingService,
	config *config.Config,
) FileService {
	return &fileService{
		repoManager:            repoManager,
		storageService:         storageService,
		permissionService:      permissionService,
		notificationService:    notificationService,
		activityService:        activityService,
		imageProcessingService: imageProcessingService,
		config:                 config,
	}
}

// UploadFile handles file upload operations
func (s *fileService) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	// Validate file size
	if req.Header.Size > s.config.FileUpload.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.config.FileUpload.MaxFileSize)
	}

	// Check user storage quota
	stats, err := s.GetUserStorageStats(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check storage quota: %w", err)
	}

	if stats.UsedSpace+req.Header.Size > stats.TotalSpace {
		return nil, fmt.Errorf("insufficient storage space")
	}

	// Generate file ID and path
	fileID := uuid.New()
	filename := s.generateUniqueFilename(req.Header.Filename)
	filePath := s.generateFilePath(req.UserID, filename)

	// Calculate checksum
	checksum, err := s.calculateChecksum(req.File)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Reset file reader
	if _, err := req.File.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Check for duplicate files
	if !req.Overwrite {
		existingFile, err := s.repoManager.File().GetByChecksum(ctx, checksum)
		if err == nil && existingFile != nil && existingFile.UserID == req.UserID {
			return &UploadFileResponse{
				File:     existingFile,
				URL:      existingFile.URL,
				Checksum: existingFile.Checksum,
			}, nil
		}
	}

	// Store file in storage
	err = s.storageService.Store(ctx, filePath, req.File, req.Header.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	// Generate file URL
	fileURL, err := s.storageService.GetURL(ctx, filePath, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file URL: %w", err)
	}

	// Create file record
	file := &fileDomain.File{
		ID:           fileID,
		UserID:       req.UserID,
		Filename:     filename,
		OriginalName: req.Header.Filename,
		MimeType:     req.Header.Header.Get("Content-Type"),
		Size:         req.Header.Size,
		Path:         filePath,
		URL:          fileURL,
		Checksum:     checksum,
		Status:       fileDomain.FileStatusReady,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save file to database
	err = s.repoManager.File().Create(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	// Log activity
	ipAddress := utils.GetIPFromContext(ctx)
	userAgent := utils.GetUserAgentFromContext(ctx)
	details := map[string]interface{}{
		"filename":      file.Filename,
		"original_name": file.OriginalName,
		"size":          file.Size,
		"mime_type":     file.MimeType,
	}

	err = s.activityService.LogActivity(ctx, req.UserID, "file_upload", "file", &fileID, details, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	// Send notification
	err = s.notificationService.SendFileUploadedNotification(ctx, req.UserID, file.Filename)
	if err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Failed to send notification: %v\n", err)
	}

	// Generate image variants if the file is an image
	if s.isImageFile(file.MimeType) {
		go func() {
			// Use background context for async processing
			bgCtx := context.Background()
			if err := s.imageProcessingService.GenerateVariants(bgCtx, file.ID, file.Path); err != nil {
				fmt.Printf("Failed to generate image variants for file %s: %v\n", file.ID, err)
			}
		}()
	}

	return &UploadFileResponse{
		File:     file,
		URL:      fileURL,
		Checksum: checksum,
	}, nil
}

// GetFile retrieves a file by ID
func (s *fileService) GetFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*fileDomain.File, error) {
	// Check permissions
	hasPermission, err := s.permissionService.CheckFilePermission(ctx, userID, fileID, "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("access denied")
	}

	file, err := s.repoManager.File().GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

// GetFileContent retrieves file content
func (s *fileService) GetFileContent(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (io.ReadCloser, error) {
	// Check permissions
	hasPermission, err := s.permissionService.CheckFilePermission(ctx, userID, fileID, "read")
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("access denied")
	}

	file, err := s.repoManager.File().GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	content, err := s.storageService.Retrieve(ctx, file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file content: %w", err)
	}

	// Log activity
	ipAddress := utils.GetIPFromContext(ctx)
	userAgent := utils.GetUserAgentFromContext(ctx)
	details := map[string]interface{}{
		"filename": file.Filename,
		"size":     file.Size,
	}

	err = s.activityService.LogActivity(ctx, userID, "file_download", "file", &fileID, details, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail the download
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	return content, nil
}

// DeleteFile deletes a file
func (s *fileService) DeleteFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	// Check permissions
	hasPermission, err := s.permissionService.CheckFilePermission(ctx, userID, fileID, "delete")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("access denied")
	}

	file, err := s.repoManager.File().GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Delete from storage
	err = s.storageService.Delete(ctx, file.Path)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Delete from database
	err = s.repoManager.File().Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// Log activity
	ipAddress := utils.GetIPFromContext(ctx)
	userAgent := utils.GetUserAgentFromContext(ctx)
	details := map[string]interface{}{
		"filename": file.Filename,
		"size":     file.Size,
	}

	err = s.activityService.LogActivity(ctx, userID, "file_delete", "file", &fileID, details, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail the deletion
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	return nil
}

// ListFiles lists files for a user
func (s *fileService) ListFiles(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error) {
	files, err := s.repoManager.File().GetByUserID(ctx, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	total, err := s.repoManager.File().GetFileCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	return &ListFilesResponse{
		Files:   files,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+req.Limit) < total,
	}, nil
}

// SearchFiles searches files for a user
func (s *fileService) SearchFiles(ctx context.Context, userID uuid.UUID, req *SearchFilesRequest) (*SearchFilesResponse, error) {
	files, err := s.repoManager.File().Search(ctx, req.Query, userID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	// For simplicity, we'll use the file count as total for search results
	total, err := s.repoManager.File().GetFileCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	return &SearchFilesResponse{
		Files:   files,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: int64(req.Offset+req.Limit) < total,
	}, nil
}

// GetFileMetadata retrieves file metadata
func (s *fileService) GetFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*FileMetadata, error) {
	file, err := s.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		ID:           file.ID,
		Filename:     file.Filename,
		OriginalName: file.OriginalName,
		MimeType:     file.MimeType,
		Size:         file.Size,
		Checksum:     file.Checksum,
		Metadata:     file.Metadata,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
	}, nil
}

// UpdateFileMetadata updates file metadata
func (s *fileService) UpdateFileMetadata(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, metadata map[string]interface{}) error {
	// Check permissions
	hasPermission, err := s.permissionService.CheckFilePermission(ctx, userID, fileID, "write")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("access denied")
	}

	file, err := s.repoManager.File().GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Update metadata
	file.Metadata = metadata
	file.UpdatedAt = time.Now()

	err = s.repoManager.File().Update(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
}

// isImageFile checks if the file is an image based on its MIME type
func (s *fileService) isImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// GetUserStorageStats retrieves user storage statistics
func (s *fileService) GetUserStorageStats(ctx context.Context, userID uuid.UUID) (*StorageStats, error) {
	usedSpace, err := s.repoManager.File().GetTotalSize(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get used space: %w", err)
	}

	fileCount, err := s.repoManager.File().GetFileCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file count: %w", err)
	}

	totalSpace := s.config.FileUpload.MaxTotalSize
	quotaUsed := float64(usedSpace) / float64(totalSpace) * 100

	return &StorageStats{
		UsedSpace:  usedSpace,
		TotalSpace: totalSpace,
		FileCount:  fileCount,
		QuotaUsed:  quotaUsed,
	}, nil
}

// Helper functions

func (s *fileService) generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%d_%s%s", name, timestamp, uuid, ext)
}

func (s *fileService) generateFilePath(userID uuid.UUID, filename string) string {
	return fmt.Sprintf("users/%s/files/%s", userID.String(), filename)
}

func (s *fileService) calculateChecksum(file io.Reader) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
