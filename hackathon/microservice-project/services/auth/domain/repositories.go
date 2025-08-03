package domain

import (
	"context"
	"time"
	"github.com/elotusteam/microservice-project/shared/domain"
	"github.com/elotusteam/microservice-project/shared/data"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	data.Repository
	
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*domain.User, error)
	
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	
	// Update updates a user
	Update(ctx context.Context, user *domain.User) error
	
	// UpdatePassword updates user password
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	
	// UpdateLastLogin updates user's last login time
	UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error
	
	// Delete soft deletes a user
	Delete(ctx context.Context, id string) error
	
	// ExistsByUsername checks if username exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	
	// ExistsByEmail checks if email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// List retrieves users with pagination
	List(ctx context.Context, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// Search searches users by criteria
	Search(ctx context.Context, criteria map[string]interface{}, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetActiveUserCount gets count of active users
	GetActiveUserCount(ctx context.Context) (int64, error)
	
	// GetUsersByRole gets users by role
	GetUsersByRole(ctx context.Context, role domain.UserRole, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetUsersByStatus gets users by status
	GetUsersByStatus(ctx context.Context, status domain.UserStatus, pagination *data.Pagination) (*data.PaginatedResult, error)
}

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	data.Repository
	
	// Create creates a new session
	Create(ctx context.Context, session *domain.Session) error
	
	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	
	// GetByTokenID retrieves a session by token ID
	GetByTokenID(ctx context.Context, tokenID string) (*domain.Session, error)
	
	// GetByUserID retrieves sessions by user ID
	GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	
	// Update updates a session
	Update(ctx context.Context, session *domain.Session) error
	
	// UpdateLastUsed updates session's last used time
	UpdateLastUsed(ctx context.Context, sessionID string, lastUsed time.Time) error
	
	// Delete deletes a session
	Delete(ctx context.Context, id string) error
	
	// DeleteByTokenID deletes a session by token ID
	DeleteByTokenID(ctx context.Context, tokenID string) error
	
	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error
	
	// DeleteExpired deletes expired sessions
	DeleteExpired(ctx context.Context) error
	
	// GetActiveSessions gets active sessions for a user
	GetActiveSessions(ctx context.Context, userID string) ([]*domain.Session, error)
	
	// GetActiveSessionCount gets count of active sessions
	GetActiveSessionCount(ctx context.Context) (int64, error)
	
	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, sessionID string) error
	
	// RevokeAllUserSessions revokes all sessions for a user
	RevokeAllUserSessions(ctx context.Context, userID string) error
}

// RevokedTokenRepository defines the interface for revoked token operations
type RevokedTokenRepository interface {
	data.Repository
	
	// Create creates a new revoked token record
	Create(ctx context.Context, token *domain.RevokedToken) error
	
	// IsRevoked checks if a token is revoked
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
	
	// GetByTokenID retrieves a revoked token by token ID
	GetByTokenID(ctx context.Context, tokenID string) (*domain.RevokedToken, error)
	
	// GetByUserID retrieves revoked tokens by user ID
	GetByUserID(ctx context.Context, userID string) ([]*domain.RevokedToken, error)
	
	// DeleteExpired deletes expired revoked tokens
	DeleteExpired(ctx context.Context) error
	
	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, tokenID, userID, reason string, expiresAt time.Time) error
	
	// RevokeAllUserTokens revokes all tokens for a user
	RevokeAllUserTokens(ctx context.Context, userID, reason string) error
	
	// GetRevokedTokenCount gets count of revoked tokens
	GetRevokedTokenCount(ctx context.Context) (int64, error)
	
	// CleanupExpired removes expired revoked tokens
	CleanupExpired(ctx context.Context) (int64, error)
}

// PasswordResetTokenRepository defines the interface for password reset token operations
type PasswordResetTokenRepository interface {
	data.Repository
	
	// Create creates a new password reset token
	Create(ctx context.Context, token *PasswordResetToken) error
	
	// GetByToken retrieves a password reset token by token value
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	
	// GetByUserID retrieves password reset tokens by user ID
	GetByUserID(ctx context.Context, userID string) ([]*PasswordResetToken, error)
	
	// MarkAsUsed marks a token as used
	MarkAsUsed(ctx context.Context, tokenID string) error
	
	// Delete deletes a password reset token
	Delete(ctx context.Context, id string) error
	
	// DeleteByUserID deletes all password reset tokens for a user
	DeleteByUserID(ctx context.Context, userID string) error
	
	// DeleteExpired deletes expired password reset tokens
	DeleteExpired(ctx context.Context) error
	
	// IsValidToken checks if a token is valid (not used and not expired)
	IsValidToken(ctx context.Context, token string) (bool, error)
	
	// GetActiveTokensCount gets count of active reset tokens for a user
	GetActiveTokensCount(ctx context.Context, userID string) (int64, error)
	
	// CleanupExpired removes expired tokens
	CleanupExpired(ctx context.Context) (int64, error)
}

// LoginAttemptRepository defines the interface for login attempt tracking
type LoginAttemptRepository interface {
	data.Repository
	
	// Create creates a new login attempt record
	Create(ctx context.Context, attempt *LoginAttempt) error
	
	// GetRecentAttempts gets recent login attempts for an identifier
	GetRecentAttempts(ctx context.Context, identifier string, since time.Time) ([]*LoginAttempt, error)
	
	// GetFailedAttempts gets failed login attempts for an identifier
	GetFailedAttempts(ctx context.Context, identifier string, since time.Time) ([]*LoginAttempt, error)
	
	// GetSuccessfulAttempts gets successful login attempts for an identifier
	GetSuccessfulAttempts(ctx context.Context, identifier string, since time.Time) ([]*LoginAttempt, error)
	
	// CountFailedAttempts counts failed login attempts for an identifier
	CountFailedAttempts(ctx context.Context, identifier string, since time.Time) (int64, error)
	
	// CountAttemptsByIP counts login attempts by IP address
	CountAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error)
	
	// DeleteOldAttempts deletes old login attempts
	DeleteOldAttempts(ctx context.Context, before time.Time) error
	
	// GetAttemptsByTimeRange gets attempts within a time range
	GetAttemptsByTimeRange(ctx context.Context, start, end time.Time, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetSuspiciousActivity gets suspicious login activity
	GetSuspiciousActivity(ctx context.Context, threshold int, since time.Time) ([]*LoginAttempt, error)
	
	// CleanupOldAttempts removes old login attempts
	CleanupOldAttempts(ctx context.Context, retentionPeriod time.Duration) (int64, error)
}

// ActivityLogRepository defines the interface for activity logging
type ActivityLogRepository interface {
	data.Repository
	
	// Create creates a new activity log entry
	Create(ctx context.Context, log *domain.ActivityLog) error
	
	// GetByUserID retrieves activity logs by user ID
	GetByUserID(ctx context.Context, userID string, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetByAction retrieves activity logs by action
	GetByAction(ctx context.Context, action string, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetByTimeRange retrieves activity logs within a time range
	GetByTimeRange(ctx context.Context, start, end time.Time, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetByResourceType retrieves activity logs by resource type
	GetByResourceType(ctx context.Context, resourceType string, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// Search searches activity logs by criteria
	Search(ctx context.Context, criteria map[string]interface{}, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// DeleteOldLogs deletes old activity logs
	DeleteOldLogs(ctx context.Context, before time.Time) error
	
	// GetSecurityEvents gets security-related events
	GetSecurityEvents(ctx context.Context, since time.Time, pagination *data.Pagination) (*data.PaginatedResult, error)
	
	// GetUserActivity gets user activity summary
	GetUserActivity(ctx context.Context, userID string, since time.Time) (map[string]int64, error)
	
	// CleanupOldLogs removes old activity logs
	CleanupOldLogs(ctx context.Context, retentionPeriod time.Duration) (int64, error)
}

// CacheRepository defines the interface for authentication caching
type AuthCacheRepository interface {
	data.CacheRepository
	
	// SetUserSession caches user session data
	SetUserSession(ctx context.Context, sessionID string, user *AuthUser, ttl time.Duration) error
	
	// GetUserSession retrieves cached user session data
	GetUserSession(ctx context.Context, sessionID string) (*AuthUser, error)
	
	// DeleteUserSession removes cached user session data
	DeleteUserSession(ctx context.Context, sessionID string) error
	
	// SetLoginAttempts caches login attempt count
	SetLoginAttempts(ctx context.Context, identifier string, count int, ttl time.Duration) error
	
	// GetLoginAttempts retrieves cached login attempt count
	GetLoginAttempts(ctx context.Context, identifier string) (int, error)
	
	// IncrementLoginAttempts increments login attempt count
	IncrementLoginAttempts(ctx context.Context, identifier string, ttl time.Duration) (int, error)
	
	// ResetLoginAttempts resets login attempt count
	ResetLoginAttempts(ctx context.Context, identifier string) error
	
	// SetPasswordResetToken caches password reset token
	SetPasswordResetToken(ctx context.Context, token string, userID string, ttl time.Duration) error
	
	// GetPasswordResetToken retrieves cached password reset token
	GetPasswordResetToken(ctx context.Context, token string) (string, error)
	
	// DeletePasswordResetToken removes cached password reset token
	DeletePasswordResetToken(ctx context.Context, token string) error
	
	// SetRevokedToken caches revoked token
	SetRevokedToken(ctx context.Context, tokenID string, ttl time.Duration) error
	
	// IsTokenRevoked checks if token is revoked from cache
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
	
	// SetUserLockout sets user lockout status
	SetUserLockout(ctx context.Context, userID string, ttl time.Duration) error
	
	// IsUserLockedOut checks if user is locked out
	IsUserLockedOut(ctx context.Context, userID string) (bool, error)
	
	// RemoveUserLockout removes user lockout status
	RemoveUserLockout(ctx context.Context, userID string) error
	
	// SetRateLimitCounter sets rate limit counter
	SetRateLimitCounter(ctx context.Context, key string, count int, ttl time.Duration) error
	
	// GetRateLimitCounter gets rate limit counter
	GetRateLimitCounter(ctx context.Context, key string) (int, error)
	
	// IncrementRateLimitCounter increments rate limit counter
	IncrementRateLimitCounter(ctx context.Context, key string, ttl time.Duration) (int, error)
}

// RepositoryManager defines the interface for managing all repositories
type RepositoryManager interface {
	// GetUserRepository returns the user repository
	GetUserRepository() UserRepository
	
	// GetSessionRepository returns the session repository
	GetSessionRepository() SessionRepository
	
	// GetRevokedTokenRepository returns the revoked token repository
	GetRevokedTokenRepository() RevokedTokenRepository
	
	// GetPasswordResetTokenRepository returns the password reset token repository
	GetPasswordResetTokenRepository() PasswordResetTokenRepository
	
	// GetLoginAttemptRepository returns the login attempt repository
	GetLoginAttemptRepository() LoginAttemptRepository
	
	// GetActivityLogRepository returns the activity log repository
	GetActivityLogRepository() ActivityLogRepository
	
	// GetCacheRepository returns the cache repository
	GetCacheRepository() AuthCacheRepository
	
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (data.Transaction, error)
	
	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(tx data.Transaction) error) error
	
	// Close closes all repository connections
	Close() error
	
	// Health checks the health of all repositories
	Health(ctx context.Context) error
}