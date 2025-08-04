package infrastructure

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	fileDomain "github.com/elotusteam/microservice-project/services/file/domain"
	fileUsecases "github.com/elotusteam/microservice-project/services/file/usecases"
	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/google/uuid"
)

// ImageProcessingService implements fileUsecases.ImageProcessingService
type ImageProcessingService struct {
	repoManager    fileDomain.RepositoryManager
	storageService fileUsecases.StorageService
	config         *config.Config
}

// NewImageProcessingService creates a new image processing service
func NewImageProcessingService(
	repoManager fileDomain.RepositoryManager,
	storageService fileUsecases.StorageService,
	config *config.Config,
) fileUsecases.ImageProcessingService {
	return &ImageProcessingService{
		repoManager:    repoManager,
		storageService: storageService,
		config:         config,
	}
}

// ImageVariantConfig defines configuration for image variants
type ImageVariantConfig struct {
	Type    string
	Width   int
	Height  int
	Quality int
	Format  string
}

// getVariantConfigs returns the predefined image variant configurations
func (s *ImageProcessingService) getVariantConfigs() []ImageVariantConfig {
	return []ImageVariantConfig{
		{Type: "thumbnail", Width: 150, Height: 150, Quality: 80, Format: "jpeg"},
		{Type: "small", Width: 300, Height: 300, Quality: 85, Format: "jpeg"},
		{Type: "medium", Width: 600, Height: 600, Quality: 90, Format: "jpeg"},
		{Type: "large", Width: 1200, Height: 1200, Quality: 95, Format: "jpeg"},
		{Type: "webp_small", Width: 300, Height: 300, Quality: 80, Format: "webp"},
		{Type: "webp_medium", Width: 600, Height: 600, Quality: 85, Format: "webp"},
	}
}

// GenerateVariants generates image variants for a given file
func (s *ImageProcessingService) GenerateVariants(ctx context.Context, fileID uuid.UUID, originalPath string) error {
	// Check if the file is an image
	if !s.isImageFile(originalPath) {
		return fmt.Errorf("file is not an image: %s", originalPath)
	}

	// Load the original image
	originalImage, err := s.loadImage(originalPath)
	if err != nil {
		return fmt.Errorf("failed to load original image: %w", err)
	}

	// Generate variants
	configs := s.getVariantConfigs()
	for _, config := range configs {
		variant, err := s.generateVariant(ctx, fileID, originalImage, config, originalPath)
		if err != nil {
			// Log error but continue with other variants
			fmt.Printf("Failed to generate variant %s: %v\n", config.Type, err)
			continue
		}

		// Save variant to database
		if err := s.repoManager.ImageVariant().Create(ctx, variant); err != nil {
			fmt.Printf("Failed to save variant %s to database: %v\n", config.Type, err)
		}
	}

	return nil
}

// GetVariants retrieves all variants for a file
func (s *ImageProcessingService) GetVariants(ctx context.Context, fileID uuid.UUID) ([]*fileDomain.ImageVariant, error) {
	return s.repoManager.ImageVariant().GetByFileID(ctx, fileID)
}

// DeleteVariants deletes all variants for a file
func (s *ImageProcessingService) DeleteVariants(ctx context.Context, fileID uuid.UUID) error {
	// Get all variants first
	variants, err := s.repoManager.ImageVariant().GetByFileID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get variants: %w", err)
	}

	// Delete files from storage
	for _, variant := range variants {
		if err := s.storageService.Delete(ctx, variant.Path); err != nil {
			fmt.Printf("Failed to delete variant file %s: %v\n", variant.Path, err)
		}
	}

	// Delete from database
	return s.repoManager.ImageVariant().DeleteByFileID(ctx, fileID)
}

// generateVariant creates a single image variant
func (s *ImageProcessingService) generateVariant(
	ctx context.Context,
	fileID uuid.UUID,
	originalImage image.Image,
	config ImageVariantConfig,
	originalPath string,
) (*fileDomain.ImageVariant, error) {
	// Resize image
	resizedImage := s.resizeImage(originalImage, config.Width, config.Height)

	// Generate variant path
	variantPath := s.generateVariantPath(originalPath, config.Type, config.Format)

	// Save resized image
	if err := s.saveImage(resizedImage, variantPath, config.Format, config.Quality); err != nil {
		return nil, fmt.Errorf("failed to save variant image: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(variantPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant file info: %w", err)
	}

	// Create variant entity
	variant := &fileDomain.ImageVariant{
		ID:          fmt.Sprintf("%s_%s", fileID.String(), config.Type),
		FileID:      fileID,
		VariantType: config.Type,
		Width:       config.Width,
		Height:      config.Height,
		Size:        fileInfo.Size(),
		Path:        variantPath,
		Format:      config.Format,
		Quality:     config.Quality,
		Status:      fileDomain.ImageVariantStatusReady,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return variant, nil
}

// isImageFile checks if the file is an image based on its extension
func (s *ImageProcessingService) isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// loadImage loads an image from file
func (s *ImageProcessingService) loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// resizeImage resizes an image to the specified dimensions using nearest neighbor
func (s *ImageProcessingService) resizeImage(img image.Image, width, height int) image.Image {
	// Calculate aspect ratio preserving dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate scaling factor to fit within target dimensions
	scaleX := float64(width) / float64(originalWidth)
	scaleY := float64(height) / float64(originalHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Calculate new dimensions
	newWidth := int(float64(originalWidth) * scale)
	newHeight := int(float64(originalHeight) * scale)

	// Create new image using simple nearest neighbor scaling
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Map destination pixel to source pixel
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)

			// Ensure we don't go out of bounds
			if srcX >= originalWidth {
				srcX = originalWidth - 1
			}
			if srcY >= originalHeight {
				srcY = originalHeight - 1
			}

			// Copy pixel
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

// saveImage saves an image to file with specified format and quality
func (s *ImageProcessingService) saveImage(img image.Image, path, format string, quality int) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode based on format
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(file, img)
	case "webp":
		// For now, save as JPEG since webp encoding requires additional library
		return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// generateVariantPath generates a path for the variant file
func (s *ImageProcessingService) generateVariantPath(originalPath, variantType, format string) string {
	dir := filepath.Dir(originalPath)
	filename := filepath.Base(originalPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// Create variant filename
	variantFilename := fmt.Sprintf("%s_%s.%s", name, variantType, format)
	return filepath.Join(dir, "variants", variantFilename)
}
