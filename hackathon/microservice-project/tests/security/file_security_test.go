package security

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileUploadSecurity tests file upload security measures
func TestFileUploadSecurity(t *testing.T) {
	maliciousFiles := []struct {
		name        string
		content     []byte
		mimeType    string
		filename    string
		expectError bool
		desc        string
	}{
		{
			name:        "Valid image file",
			content:     createValidImageContent(),
			mimeType:    "image/jpeg",
			filename:    "test.jpg",
			expectError: false,
			desc:        "Should accept valid image files",
		},
		{
			name:        "Executable disguised as image",
			content:     createExecutableContent(),
			mimeType:    "image/jpeg",
			filename:    "malware.jpg",
			expectError: true,
			desc:        "Should reject executable files disguised as images",
		},
		{
			name:        "Script injection in filename",
			content:     createValidImageContent(),
			mimeType:    "image/png",
			filename:    "<script>alert('xss')</script>.png",
			expectError: true,
			desc:        "Should reject files with script injection in filename",
		},
		{
			name:        "Oversized file",
			content:     createOversizedContent(),
			mimeType:    "image/jpeg",
			filename:    "large.jpg",
			expectError: true,
			desc:        "Should reject files exceeding size limit",
		},
		{
			name:        "Path traversal in filename",
			content:     createValidImageContent(),
			mimeType:    "image/png",
			filename:    "../../../etc/passwd.png",
			expectError: true,
			desc:        "Should reject files with path traversal attempts",
		},
		{
			name:        "Null byte injection",
			content:     createValidImageContent(),
			mimeType:    "image/png",
			filename:    "test.png\x00.exe",
			expectError: true,
			desc:        "Should reject files with null byte injection",
		},
		{
			name:        "Double extension",
			content:     createExecutableContent(),
			mimeType:    "application/octet-stream",
			filename:    "malware.jpg.exe",
			expectError: true,
			desc:        "Should reject files with double extensions",
		},
	}

	for _, tt := range maliciousFiles {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileUpload(tt.content, tt.mimeType, tt.filename)
			if tt.expectError {
				assert.Error(t, err, tt.desc)
			} else {
				assert.NoError(t, err, tt.desc)
			}
		})
	}
}

// TestFileTypeValidation tests file type validation
func TestFileTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		mimeType    string
		filename    string
		expectValid bool
		desc        string
	}{
		{
			name:        "Valid JPEG",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG magic bytes
			mimeType:    "image/jpeg",
			filename:    "test.jpg",
			expectValid: true,
			desc:        "Should accept valid JPEG files",
		},
		{
			name:        "Valid PNG",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG magic bytes
			mimeType:    "image/png",
			filename:    "test.png",
			expectValid: true,
			desc:        "Should accept valid PNG files",
		},
		{
			name:        "Mismatched content and extension",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG content
			mimeType:    "image/png",                    // PNG MIME type
			filename:    "test.png",
			expectValid: false,
			desc:        "Should reject files with mismatched content and MIME type",
		},
		{
			name:        "Executable file",
			content:     []byte{0x4D, 0x5A}, // PE executable magic bytes
			mimeType:    "application/octet-stream",
			filename:    "malware.exe",
			expectValid: false,
			desc:        "Should reject executable files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateFileType(tt.content, tt.mimeType, tt.filename)
			assert.Equal(t, tt.expectValid, valid, tt.desc)
		})
	}
}

// TestFileContentScanning tests malware and content scanning
func TestFileContentScanning(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		expectClean bool
		desc        string
	}{
		{
			name:        "Clean image content",
			content:     createValidImageContent(),
			expectClean: true,
			desc:        "Should pass clean image content",
		},
		{
			name:        "Embedded script in image",
			content:     createImageWithEmbeddedScript(),
			expectClean: false,
			desc:        "Should detect embedded scripts in images",
		},
		{
			name:        "Suspicious binary patterns",
			content:     createSuspiciousBinaryContent(),
			expectClean: false,
			desc:        "Should detect suspicious binary patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clean := scanFileContent(tt.content)
			assert.Equal(t, tt.expectClean, clean, tt.desc)
		})
	}
}

// TestFileUploadEndpoint tests the actual file upload endpoint security
func TestFileUploadEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		content        []byte
		mimeType       string
		authorization  string
		expectedStatus int
		desc           string
	}{
		{
			name:           "Valid upload with auth",
			filename:       "test.jpg",
			content:        createValidImageContent(),
			mimeType:       "image/jpeg",
			authorization:  "Bearer " + generateValidToken(t),
			expectedStatus: http.StatusOK,
			desc:           "Should accept valid file upload with authentication",
		},
		{
			name:           "Upload without auth",
			filename:       "test.jpg",
			content:        createValidImageContent(),
			mimeType:       "image/jpeg",
			authorization:  "",
			expectedStatus: http.StatusUnauthorized,
			desc:           "Should reject file upload without authentication",
		},
		{
			name:           "Malicious file upload",
			filename:       "malware.jpg",
			content:        createExecutableContent(),
			mimeType:       "image/jpeg",
			authorization:  "Bearer " + generateValidToken(t),
			expectedStatus: http.StatusBadRequest,
			desc:           "Should reject malicious file uploads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createFileUploadRequest(t, tt.filename, tt.content, tt.mimeType)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			rr := httptest.NewRecorder()
			// Mock handler would go here
			mockFileUploadHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, tt.desc)
		})
	}
}

// Helper functions for creating test content
func createValidImageContent() []byte {
	// JPEG header with minimal valid structure
	return []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48,
		0x00, 0x48, 0x00, 0x00, 0xFF, 0xD9, // End of image
	}
}

func createExecutableContent() []byte {
	// PE executable header (Windows)
	return []byte{
		0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00,
		0x04, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00,
	}
}

func createOversizedContent() []byte {
	// Create content larger than typical size limit (10MB)
	content := make([]byte, 11*1024*1024) // 11MB
	// Add JPEG header
	copy(content[:4], []byte{0xFF, 0xD8, 0xFF, 0xE0})
	return content
}

func createImageWithEmbeddedScript() []byte {
	content := createValidImageContent()
	// Embed script-like content
	script := []byte("<script>alert('xss')</script>")
	return append(content, script...)
}

func createSuspiciousBinaryContent() []byte {
	// Content with suspicious patterns that might indicate malware
	return []byte{
		0xFF, 0xD8, 0xFF, 0xE0, // JPEG header
		0x4D, 0x5A, 0x90, 0x00, // PE header embedded
		0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, // "Hello Wo"
		0x72, 0x6C, 0x64, // "rld"
	}
}

func createFileUploadRequest(t *testing.T, filename string, content []byte, mimeType string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// Mock validation functions (replace with actual implementations)
func validateFileUpload(content []byte, mimeType, filename string) error {
	// Check file size
	if len(content) > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("file too large")
	}

	// Check filename for malicious patterns
	if strings.Contains(filename, "<script>") ||
		strings.Contains(filename, "../") ||
		strings.Contains(filename, "\x00") ||
		strings.HasSuffix(filename, ".exe") {
		return fmt.Errorf("invalid filename")
	}

	// Check file type
	if !validateFileType(content, mimeType, filename) {
		return fmt.Errorf("invalid file type")
	}

	// Scan content
	if !scanFileContent(content) {
		return fmt.Errorf("malicious content detected")
	}

	return nil
}

func validateFileType(content []byte, mimeType, filename string) bool {
	// Check magic bytes
	if len(content) < 4 {
		return false
	}

	// JPEG validation
	if mimeType == "image/jpeg" {
		return content[0] == 0xFF && content[1] == 0xD8
	}

	// PNG validation
	if mimeType == "image/png" {
		return len(content) >= 8 &&
			content[0] == 0x89 && content[1] == 0x50 &&
			content[2] == 0x4E && content[3] == 0x47
	}

	// Reject executables
	if content[0] == 0x4D && content[1] == 0x5A { // PE header
		return false
	}

	return true
}

func scanFileContent(content []byte) bool {
	// Simple content scanning for suspicious patterns
	suspiciousPatterns := [][]byte{
		[]byte("<script>"),
		[]byte("</script>"),
		{0x4D, 0x5A}, // PE header
		[]byte("eval("),
	}

	for _, pattern := range suspiciousPatterns {
		if bytes.Contains(content, pattern) {
			return false
		}
	}

	return true
}

// Mock handler for testing
func mockFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Check authorization
	auth := r.Header.Get("Authorization")
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate file
	err = validateFileUpload(content, header.Header.Get("Content-Type"), header.Filename)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
