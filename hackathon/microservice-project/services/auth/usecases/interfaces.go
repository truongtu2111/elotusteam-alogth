package usecases

import (
	"context"
	"time"

	"github.com/elotusteam/microservice-project/services/auth/domain"
	sharedDomain "github.com/elotusteam/microservice-project/shared/domain"
)

// TokenService defines the interface for JWT token operations
type TokenService interface {
	// Token generation
	GenerateTokenPair(ctx context.Context, user *sharedDomain.User) (*domain.TokenPair, error)
	GenerateAccessToken(ctx context.Context, user *sharedDomain.User) (string, error)
	GenerateRefreshToken(ctx context.Context, user *sharedDomain.User) (string, error)

	// Token validation
	ValidateAccessToken(ctx context.Context, token string) (*domain.JWTClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*domain.JWTClaims, error)
	ValidateToken(ctx context.Context, token string) (*domain.JWTClaims, error)

	// Token parsing
	ParseToken(ctx context.Context, token string) (*domain.JWTClaims, error)
	ExtractTokenID(ctx context.Context, token string) (string, error)
	ExtractUserID(ctx context.Context, token string) (string, error)

	// Token utilities
	GetTokenExpiration(ctx context.Context, token string) (time.Time, error)
	IsTokenExpired(ctx context.Context, token string) (bool, error)
	GetTokenType(ctx context.Context, token string) (string, error)
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Email notifications
	SendWelcomeEmail(ctx context.Context, email, firstName string) error
	SendPasswordResetEmail(ctx context.Context, email, resetToken string) error
	SendPasswordChangedEmail(ctx context.Context, email, firstName string) error
	SendAccountLockedEmail(ctx context.Context, email, firstName string) error
	SendLoginAlertEmail(ctx context.Context, email, firstName, ipAddress, userAgent string) error

	// SMS notifications
	SendSMS(ctx context.Context, phoneNumber, message string) error
	SendLoginAlertSMS(ctx context.Context, phoneNumber, message string) error

	// Push notifications
	SendPushNotification(ctx context.Context, userID, title, message string, data map[string]interface{}) error

	// In-app notifications
	CreateNotification(ctx context.Context, userID, title, message, notificationType string, data map[string]interface{}) error
	MarkAsRead(ctx context.Context, notificationID string) error
	GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*sharedDomain.Notification, error)

	// Notification preferences
	GetUserPreferences(ctx context.Context, userID string) (*NotificationPreferences, error)
	UpdateUserPreferences(ctx context.Context, userID string, preferences *NotificationPreferences) error
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID          string `json:"user_id"`
	EmailEnabled    bool   `json:"email_enabled"`
	SMSEnabled      bool   `json:"sms_enabled"`
	PushEnabled     bool   `json:"push_enabled"`
	LoginAlerts     bool   `json:"login_alerts"`
	SecurityAlerts  bool   `json:"security_alerts"`
	MarketingEmails bool   `json:"marketing_emails"`
	ProductUpdates  bool   `json:"product_updates"`
}

// RateLimitService defines the interface for rate limiting operations
type RateLimitService interface {
	// Rate limiting
	CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error)
	IncrementCounter(ctx context.Context, identifier string, window time.Duration) error
	GetCounter(ctx context.Context, identifier string) (int, error)
	ResetCounter(ctx context.Context, identifier string) error

	// Advanced rate limiting
	CheckSlidingWindowRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error)
	CheckTokenBucketRateLimit(ctx context.Context, identifier string, capacity, refillRate int, window time.Duration) (bool, error)

	// Rate limit info
	GetRateLimitInfo(ctx context.Context, identifier string) (*RateLimitInfo, error)
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Identifier string        `json:"identifier"`
	Limit      int           `json:"limit"`
	Remaining  int           `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	Window     time.Duration `json:"window"`
	Blocked    bool          `json:"blocked"`
}

// SecurityService defines the interface for security operations
type SecurityService interface {
	// User lockout management
	IsUserLockedOut(ctx context.Context, userID string) (bool, error)
	LockUser(ctx context.Context, userID string, duration time.Duration, reason string) error
	UnlockUser(ctx context.Context, userID string) error
	GetLockoutInfo(ctx context.Context, userID string) (*LockoutInfo, error)

	// Login attempt tracking
	RecordLoginAttempt(ctx context.Context, identifier string, success bool, ipAddress, userAgent string) error
	GetLoginAttempts(ctx context.Context, identifier string, since time.Time) ([]*domain.LoginAttempt, error)
	ClearLoginAttempts(ctx context.Context, identifier string) error

	// Suspicious activity detection
	DetectSuspiciousActivity(ctx context.Context, userID, ipAddress, userAgent string) (bool, error)
	RecordSuspiciousActivity(ctx context.Context, userID, activityType, description, ipAddress, userAgent string) error

	// IP and device management
	IsIPBlocked(ctx context.Context, ipAddress string) (bool, error)
	BlockIP(ctx context.Context, ipAddress string, duration time.Duration, reason string) error
	UnblockIP(ctx context.Context, ipAddress string) error
	IsTrustedDevice(ctx context.Context, userID, deviceFingerprint string) (bool, error)
	AddTrustedDevice(ctx context.Context, userID, deviceFingerprint, deviceName string) error

	// Password security
	CheckPasswordHistory(ctx context.Context, userID, newPasswordHash string) (bool, error)
	AddPasswordToHistory(ctx context.Context, userID, passwordHash string) error
	IsPasswordCompromised(ctx context.Context, password string) (bool, error)
}

// LockoutInfo represents user lockout information
type LockoutInfo struct {
	UserID    string    `json:"user_id"`
	Locked    bool      `json:"locked"`
	LockedAt  time.Time `json:"locked_at"`
	UnlocksAt time.Time `json:"unlocks_at"`
	Reason    string    `json:"reason"`
	Attempts  int       `json:"attempts"`
}

// ActivityService defines the interface for activity logging
type ActivityService interface {
	// Activity logging
	LogActivity(ctx context.Context, userID, action, resourceType, resourceID string, details map[string]interface{}, ipAddress, userAgent string) error
	LogSecurityEvent(ctx context.Context, userID, eventType, description string, severity string, ipAddress, userAgent string) error

	// Activity retrieval
	GetUserActivity(ctx context.Context, userID string, limit, offset int) ([]*sharedDomain.ActivityLog, error)
	GetActivityByType(ctx context.Context, activityType string, since time.Time, limit, offset int) ([]*sharedDomain.ActivityLog, error)
	GetSecurityEvents(ctx context.Context, since time.Time, severity string, limit, offset int) ([]*sharedDomain.ActivityLog, error)

	// Activity analytics
	GetActivityStats(ctx context.Context, userID string, since time.Time) (*ActivityStats, error)
	GetSystemActivityStats(ctx context.Context, since time.Time) (*SystemActivityStats, error)
}

// ActivityStats represents user activity statistics
type ActivityStats struct {
	UserID          string         `json:"user_id"`
	TotalActivities int            `json:"total_activities"`
	ByType          map[string]int `json:"by_type"`
	ByDay           map[string]int `json:"by_day"`
	LastActivity    time.Time      `json:"last_activity"`
	MostActive      string         `json:"most_active"`
}

// SystemActivityStats represents system-wide activity statistics
type SystemActivityStats struct {
	TotalActivities int            `json:"total_activities"`
	TotalUsers      int            `json:"total_users"`
	ActiveUsers     int            `json:"active_users"`
	ByType          map[string]int `json:"by_type"`
	ByHour          map[string]int `json:"by_hour"`
	TopUsers        []string       `json:"top_users"`
	PeakHour        string         `json:"peak_hour"`
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	// Password hashing
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error

	// Password validation
	ValidatePasswordStrength(password string) []string
	CheckPasswordPolicy(password string, policy *PasswordPolicy) []string

	// Password reset
	GenerateResetToken(userID string) (string, error)
	ValidateResetToken(token string) (string, error)
	InvalidateResetToken(token string) error

	// Password history
	CheckPasswordHistory(userID, newPasswordHash string, historyCount int) (bool, error)
	AddPasswordToHistory(userID, passwordHash string) error
}

// PasswordPolicy represents password policy configuration
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	MaxLength        int  `json:"max_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	PreventReuse     int  `json:"prevent_reuse"`
	MaxAge           int  `json:"max_age"`
}
