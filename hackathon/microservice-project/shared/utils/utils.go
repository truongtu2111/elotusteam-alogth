package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// GenerateID generates a new UUID
func GenerateID() string {
	return uuid.New().String()
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashString creates a SHA256 hash of the input string
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// HashFile creates a SHA256 hash of the file content
func HashFile(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername validates a username
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernameRegex.MatchString(username)
}

// String validation helpers
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

func IsValidUsername(username string) bool {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	return usernameRegex.MatchString(username)
}

func IsValidPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

// Password strength validation helpers
func HasUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func HasLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func HasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func HasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		for _, special := range specialChars {
			if r == special {
				return true
			}
		}
	}
	return false
}

// Legacy aliases for backward compatibility
func ContainsUppercase(s string) bool   { return HasUppercase(s) }
func ContainsLowercase(s string) bool   { return HasLowercase(s) }
func ContainsNumber(s string) bool      { return HasDigit(s) }
func ContainsSpecialChar(s string) bool { return HasSpecialChar(s) }

// ValidatePassword validates a password based on security requirements
func ValidatePassword(password string, minLength int, requireUppercase, requireLowercase, requireNumbers, requireSymbols bool) []string {
	var errors []string

	if len(password) < minLength {
		errors = append(errors, fmt.Sprintf("Password must be at least %d characters long", minLength))
	}

	if requireUppercase && !hasUppercase(password) {
		errors = append(errors, "Password must contain at least one uppercase letter")
	}

	if requireLowercase && !hasLowercase(password) {
		errors = append(errors, "Password must contain at least one lowercase letter")
	}

	if requireNumbers && !hasNumbers(password) {
		errors = append(errors, "Password must contain at least one number")
	}

	if requireSymbols && !hasSymbols(password) {
		errors = append(errors, "Password must contain at least one symbol")
	}

	return errors
}

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasNumbers(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSymbols(s string) bool {
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}

// ValidateFileExtension validates if the file extension is allowed
func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	return false
}

// ValidateMimeType validates if the MIME type is allowed
func ValidateMimeType(mimeType string, allowedMimeTypes []string) bool {
	for _, allowed := range allowedMimeTypes {
		if matched, _ := filepath.Match(allowed, mimeType); matched {
			return true
		}
	}
	return false
}

// DetectMimeType detects the MIME type of a file
func DetectMimeType(filename string, content []byte) string {
	// First try to detect from content
	if len(content) > 0 {
		if mimeType := http.DetectContentType(content); mimeType != "application/octet-stream" {
			return mimeType
		}
	}

	// Fallback to extension-based detection
	ext := filepath.Ext(filename)
	return mime.TypeByExtension(ext)
}

// SanitizeFilename sanitizes a filename by removing or replacing invalid characters
func SanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	sanitized := invalidChars.ReplaceAllString(filename, "_")

	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")

	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "file"
	}

	// Limit length
	if len(sanitized) > 255 {
		ext := filepath.Ext(sanitized)
		name := sanitized[:255-len(ext)]
		sanitized = name + ext
	}

	return sanitized
}

// GenerateUniqueFilename generates a unique filename with timestamp and random suffix
func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)
	name = SanitizeFilename(name)

	timestamp := time.Now().Unix()
	randomSuffix, _ := GenerateRandomString(8)

	return fmt.Sprintf("%s_%d_%s%s", name, timestamp, randomSuffix, ext)
}

// FormatFileSize formats file size in human-readable format
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ParseFileSize parses human-readable file size to bytes
func ParseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Extract number and unit
	var numStr string
	var unit string

	for i, r := range sizeStr {
		if unicode.IsDigit(r) || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid size format")
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %v", err)
	}

	unit = strings.TrimSpace(unit)
	if unit == "" || unit == "B" {
		return int64(num), nil
	}

	multipliers := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
		"PB": 1024 * 1024 * 1024 * 1024 * 1024,
	}

	if multiplier, exists := multipliers[unit]; exists {
		return int64(num * float64(multiplier)), nil
	}

	return 0, fmt.Errorf("unknown unit: %s", unit)
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// SliceContains checks if a slice contains a specific item
func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveFromSlice removes an item from a slice
func RemoveFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueSlice removes duplicates from a slice
func UniqueSlice(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// MergeMaps merges multiple maps into one
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// GetMapValue safely gets a value from a map with a default
func GetMapValue(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, exists := m[key]; exists {
		return value
	}
	return defaultValue
}

// StringToInt converts string to int with default value
func StringToInt(s string, defaultValue int) int {
	if value, err := strconv.Atoi(s); err == nil {
		return value
	}
	return defaultValue
}

// StringToInt64 converts string to int64 with default value
func StringToInt64(s string, defaultValue int64) int64 {
	if value, err := strconv.ParseInt(s, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// StringToBool converts string to bool with default value
func StringToBool(s string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(s); err == nil {
		return value
	}
	return defaultValue
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// NormalizeString normalizes a string by trimming whitespace and converting to lowercase
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ExtractIPAddress extracts IP address from various sources
func ExtractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}

	// Fallback to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}

	return r.RemoteAddr
}

// GetUserAgent extracts user agent from request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// TimePtr returns a pointer to a time.Time value
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int value
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to a bool value
func BoolPtr(b bool) *bool {
	return &b
}

// DerefString safely dereferences a string pointer
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefInt safely dereferences an int pointer
func DerefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// DerefBool safely dereferences a bool pointer
func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// DerefTime safely dereferences a time pointer
func DerefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// CalculatePagination calculates pagination parameters
func CalculatePagination(page, pageSize int, total int64) (offset int, limit int, totalPages int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset = (page - 1) * pageSize
	limit = pageSize
	totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))

	return offset, limit, totalPages
}

// IsImageFile checks if a file is an image based on MIME type
func IsImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsVideoFile checks if a file is a video based on MIME type
func IsVideoFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// IsAudioFile checks if a file is an audio based on MIME type
func IsAudioFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}

// IsDocumentFile checks if a file is a document based on MIME type
func IsDocumentFile(mimeType string) bool {
	documentTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
		"application/rtf",
	}

	for _, docType := range documentTypes {
		if mimeType == docType {
			return true
		}
	}

	return false
}

// GetFileCategory categorizes a file based on its MIME type
func GetFileCategory(mimeType string) string {
	switch {
	case IsImageFile(mimeType):
		return "image"
	case IsVideoFile(mimeType):
		return "video"
	case IsAudioFile(mimeType):
		return "audio"
	case IsDocumentFile(mimeType):
		return "document"
	default:
		return "other"
	}
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int, initialDelay time.Duration) error {
	var err error
	delay := initialDelay

	for i := 0; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i < maxRetries {
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return err
}

// Debounce creates a debounced function that delays invoking func until after wait duration
func Debounce(fn func(), wait time.Duration) func() {
	var timer *time.Timer

	return func() {
		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(wait, fn)
	}
}

// Throttle creates a throttled function that only invokes func at most once per every wait duration
func Throttle(fn func(), wait time.Duration) func() {
	var lastCall time.Time

	return func() {
		now := time.Now()
		if now.Sub(lastCall) >= wait {
			lastCall = now
			fn()
		}
	}
}

// Context helpers
func GetIPFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value("ip").(string); ok {
		return ip
	}
	return "unknown"
}

func GetUserAgentFromContext(ctx context.Context) string {
	if ua, ok := ctx.Value("user_agent").(string); ok {
		return ua
	}
	return "unknown"
}
