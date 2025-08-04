package infrastructure

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elotusteam/microservice-project/services/file/usecases"
	"github.com/elotusteam/microservice-project/shared/config"
)

// CDNProvider represents different CDN providers
type CDNProvider string

const (
	CDNProviderPrimary   CDNProvider = "primary"
	CDNProviderSecondary CDNProvider = "secondary"
	CDNProviderTertiary  CDNProvider = "tertiary"
)

// ImageVariant represents different image sizes and qualities
type ImageVariant struct {
	Name    string
	Width   int
	Height  int
	Quality int
	Format  string
}

// MultiCDNStorageService implements usecases.StorageService with multi-CDN support
type MultiCDNStorageService struct {
	config        *config.Config
	imageVariants []ImageVariant
	primaryPath   string
	secondaryPath string
	tertiaryPath  string
	mu            sync.RWMutex
}

// NewMultiCDNStorageService creates a new multi-CDN storage service
func NewMultiCDNStorageService(cfg *config.Config) (usecases.StorageService, error) {
	// Create storage directories for different CDNs
	primaryPath := filepath.Join(cfg.Storage.LocalPath, "primary")
	secondaryPath := filepath.Join(cfg.Storage.LocalPath, "secondary")
	tertiaryPath := filepath.Join(cfg.Storage.LocalPath, "tertiary")

	// Ensure directories exist
	for _, path := range []string{primaryPath, secondaryPath, tertiaryPath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory %s: %w", path, err)
		}
	}

	// Define image variants for different qualities and sizes
	imageVariants := []ImageVariant{
		{Name: "thumbnail", Width: 150, Height: 150, Quality: 80, Format: "jpeg"},
		{Name: "small", Width: 300, Height: 300, Quality: 85, Format: "jpeg"},
		{Name: "medium", Width: 600, Height: 600, Quality: 90, Format: "jpeg"},
		{Name: "large", Width: 1200, Height: 1200, Quality: 95, Format: "jpeg"},
		{Name: "webp_small", Width: 300, Height: 300, Quality: 80, Format: "webp"},
		{Name: "webp_medium", Width: 600, Height: 600, Quality: 85, Format: "webp"},
		{Name: "webp_large", Width: 1200, Height: 1200, Quality: 90, Format: "webp"},
	}

	return &MultiCDNStorageService{
		config:        cfg,
		imageVariants: imageVariants,
		primaryPath:   primaryPath,
		secondaryPath: secondaryPath,
		tertiaryPath:  tertiaryPath,
	}, nil
}

// Store uploads a file to multiple CDNs and generates image variants if it's an image
func (s *MultiCDNStorageService) Store(ctx context.Context, path string, content io.Reader, contentType string) error {
	// Read content into buffer for multiple uploads
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, content)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	contentBytes := buf.Bytes()

	// Upload to primary CDN
	err = s.uploadToCDN(ctx, s.primaryPath, path, contentBytes)
	if err != nil {
		return fmt.Errorf("failed to upload to primary CDN: %w", err)
	}

	// Upload to secondary and tertiary CDNs asynchronously
	go func() {
		if err := s.uploadToCDN(context.Background(), s.secondaryPath, path, contentBytes); err != nil {
			fmt.Printf("Warning: Failed to upload to secondary CDN: %v\n", err)
		}
	}()

	go func() {
		if err := s.uploadToCDN(context.Background(), s.tertiaryPath, path, contentBytes); err != nil {
			fmt.Printf("Warning: Failed to upload to tertiary CDN: %v\n", err)
		}
	}()

	// Generate image variants if it's an image
	if s.isImageContent(contentType) {
		go func() {
			if err := s.generateImageVariants(context.Background(), path, contentBytes); err != nil {
				fmt.Printf("Warning: Failed to generate image variants: %v\n", err)
			}
		}()
	}

	return nil
}

// Retrieve downloads a file from the CDNs with fallback
func (s *MultiCDNStorageService) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	// Try primary CDN first
	content, err := s.retrieveFromCDN(ctx, s.primaryPath, path)
	if err == nil {
		return content, nil
	}

	// Fallback to secondary CDN
	content, err = s.retrieveFromCDN(ctx, s.secondaryPath, path)
	if err == nil {
		return content, nil
	}

	// Fallback to tertiary CDN
	content, err = s.retrieveFromCDN(ctx, s.tertiaryPath, path)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from all CDNs: %w", err)
	}

	return content, nil
}

// Delete removes a file from all CDNs
func (s *MultiCDNStorageService) Delete(ctx context.Context, path string) error {
	// Delete from primary CDN
	err := s.deleteFromCDN(ctx, s.primaryPath, path)
	if err != nil {
		return fmt.Errorf("failed to delete from primary CDN: %w", err)
	}

	// Delete from secondary and tertiary CDNs asynchronously
	go func() {
		if err := s.deleteFromCDN(context.Background(), s.secondaryPath, path); err != nil {
			fmt.Printf("Warning: Failed to delete from secondary CDN: %v\n", err)
		}
	}()

	go func() {
		if err := s.deleteFromCDN(context.Background(), s.tertiaryPath, path); err != nil {
			fmt.Printf("Warning: Failed to delete from tertiary CDN: %v\n", err)
		}
	}()

	// Delete image variants if they exist
	go func() {
		if err := s.deleteImageVariants(context.Background(), path); err != nil {
			fmt.Printf("Warning: Failed to delete image variants: %v\n", err)
		}
	}()

	return nil
}

// Exists checks if a file exists in the primary CDN
func (s *MultiCDNStorageService) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.primaryPath, path)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetURL generates a CDN URL for the file with optional image variant
func (s *MultiCDNStorageService) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// For images, try to get the best variant based on request
	if s.isImagePath(path) {
		variantPath := s.selectBestImageVariant(path, "medium") // Default to medium
		if exists, _ := s.Exists(ctx, variantPath); exists {
			path = variantPath
		}
	}

	// Generate CDN URL based on configuration
	if s.config.Storage.CDN.Enabled {
		return fmt.Sprintf("%s/%s", s.config.Storage.CDN.BaseURL, path), nil
	}

	// Fallback to local URL
	return fmt.Sprintf("/files/%s", path), nil
}

// Copy copies a file within the storage
func (s *MultiCDNStorageService) Copy(ctx context.Context, srcPath, destPath string) error {
	// Read from source
	content, err := s.Retrieve(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	defer content.Close()

	// Store to destination
	return s.Store(ctx, destPath, content, "application/octet-stream")
}

// Move moves a file within the storage
func (s *MultiCDNStorageService) Move(ctx context.Context, srcPath, destPath string) error {
	// Copy first
	err := s.Copy(ctx, srcPath, destPath)
	if err != nil {
		return err
	}

	// Then delete original
	return s.Delete(ctx, srcPath)
}

// GetSize returns the size of a file
func (s *MultiCDNStorageService) GetSize(ctx context.Context, path string) (int64, error) {
	fullPath := filepath.Join(s.primaryPath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	return info.Size(), nil
}

// Private helper methods

func (s *MultiCDNStorageService) uploadToCDN(ctx context.Context, cdnPath, filePath string, content []byte) error {
	fullPath := filepath.Join(cdnPath, filePath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *MultiCDNStorageService) retrieveFromCDN(ctx context.Context, cdnPath, filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(cdnPath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (s *MultiCDNStorageService) deleteFromCDN(ctx context.Context, cdnPath, filePath string) error {
	fullPath := filepath.Join(cdnPath, filePath)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *MultiCDNStorageService) generateImageVariants(ctx context.Context, originalPath string, content []byte) error {
	// Decode the original image
	img, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	baseDir := filepath.Dir(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))

	// Generate variants concurrently
	var wg sync.WaitGroup
	errorChan := make(chan error, len(s.imageVariants))

	for _, variant := range s.imageVariants {
		wg.Add(1)
		go func(v ImageVariant) {
			defer wg.Done()

			// Skip WebP for now (would need additional library)
			if v.Format == "webp" {
				return
			}

			// Simple resize using basic image operations
			resizedImg := s.resizeImage(img, v.Width, v.Height)

			// Encode to buffer
			var buf bytes.Buffer
			switch v.Format {
			case "jpeg":
				err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: v.Quality})
			case "png":
				err = png.Encode(&buf, resizedImg)
			default:
				err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: v.Quality})
			}

			if err != nil {
				errorChan <- fmt.Errorf("failed to encode %s variant: %w", v.Name, err)
				return
			}

			// Upload variant to all CDNs
			variantPath := fmt.Sprintf("%s/%s_%s.%s", baseDir, baseName, v.Name, v.Format)

			// Upload to primary CDN
			err = s.uploadToCDN(ctx, s.primaryPath, variantPath, buf.Bytes())
			if err != nil {
				errorChan <- fmt.Errorf("failed to upload %s variant: %w", v.Name, err)
				return
			}

			// Upload to secondary and tertiary CDNs asynchronously
			go func() {
				if err := s.uploadToCDN(context.Background(), s.secondaryPath, variantPath, buf.Bytes()); err != nil {
					fmt.Printf("Warning: Failed to upload %s variant to secondary CDN: %v\n", v.Name, err)
				}
			}()

			go func() {
				if err := s.uploadToCDN(context.Background(), s.tertiaryPath, variantPath, buf.Bytes()); err != nil {
					fmt.Printf("Warning: Failed to upload %s variant to tertiary CDN: %v\n", v.Name, err)
				}
			}()

		}(variant)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MultiCDNStorageService) deleteImageVariants(ctx context.Context, originalPath string) error {
	if !s.isImagePath(originalPath) {
		return nil
	}

	baseDir := filepath.Dir(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))

	for _, variant := range s.imageVariants {
		variantPath := fmt.Sprintf("%s/%s_%s.%s", baseDir, baseName, variant.Name, variant.Format)

		// Delete from all CDNs
		for _, cdnPath := range []string{s.primaryPath, s.secondaryPath, s.tertiaryPath} {
			if err := s.deleteFromCDN(ctx, cdnPath, variantPath); err != nil {
				fmt.Printf("Warning: Failed to delete variant %s from CDN %s: %v\n", variantPath, cdnPath, err)
			}
		}
	}

	return nil
}

func (s *MultiCDNStorageService) selectBestImageVariant(originalPath, preferredSize string) string {
	baseDir := filepath.Dir(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))

	// Try to find the preferred size first
	for _, variant := range s.imageVariants {
		if variant.Name == preferredSize {
			return fmt.Sprintf("%s/%s_%s.%s", baseDir, baseName, variant.Name, variant.Format)
		}
	}

	// Fallback to medium if preferred not found
	return fmt.Sprintf("%s/%s_medium.jpeg", baseDir, baseName)
}

func (s *MultiCDNStorageService) resizeImage(src image.Image, width, height int) image.Image {
	// Simple resize implementation using nearest neighbor
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	// Calculate aspect ratio
	aspectRatio := float64(srcW) / float64(srcH)

	// Adjust dimensions to maintain aspect ratio
	if float64(width)/float64(height) > aspectRatio {
		width = int(float64(height) * aspectRatio)
	} else {
		height = int(float64(width) / aspectRatio)
	}

	// Create new image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Simple nearest neighbor scaling
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

func (s *MultiCDNStorageService) isImageContent(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

func (s *MultiCDNStorageService) isImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"
}

// generateChecksum generates MD5 checksum for content
func (s *MultiCDNStorageService) generateChecksum(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash)
}
