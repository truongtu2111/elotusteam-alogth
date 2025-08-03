package usecases

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/elotusteam/microservice-project/services/auth/domain"
	"github.com/elotusteam/microservice-project/shared/config"
	sharedDomain "github.com/elotusteam/microservice-project/shared/domain"
	"github.com/elotusteam/microservice-project/shared/utils"
)

// AuthService defines the interface for authentication use cases
type AuthService interface {
	// Authentication operations
	Login(ctx context.Context, req *domain.LoginRequest, ipAddress, userAgent string) (*domain.AuthResponse, error)
	Register(ctx context.Context, req *domain.RegisterRequest, ipAddress, userAgent string) (*domain.AuthResponse, error)
	Logout(ctx context.Context, tokenID string) error
	LogoutAll(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, req *domain.RefreshTokenRequest) (*domain.TokenPair, error)

	// Password operations
	ChangePassword(ctx context.Context, userID string, req *domain.ChangePasswordRequest) error
	RequestPasswordReset(ctx context.Context, req *domain.ResetPasswordRequest) error
	ConfirmPasswordReset(ctx context.Context, req *domain.ConfirmResetPasswordRequest) error

	// Token operations
	ValidateToken(ctx context.Context, token string) (*domain.AuthUser, error)
	RevokeToken(ctx context.Context, tokenID, userID, reason string) error
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)

	// User operations
	GetUserProfile(ctx context.Context, userID string) (*domain.AuthUser, error)
	UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) error
	DeactivateUser(ctx context.Context, userID string) error
	ActivateUser(ctx context.Context, userID string) error

	// Session operations
	GetUserSessions(ctx context.Context, userID string) ([]*sharedDomain.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	// Security operations
	CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error)
	RecordLoginAttempt(ctx context.Context, identifier string, success bool, ipAddress, userAgent string) error
	IsUserLockedOut(ctx context.Context, userID string) (bool, error)
	LockUser(ctx context.Context, userID string, duration time.Duration, reason string) error
	UnlockUser(ctx context.Context, userID string) error

	// Activity logging
	LogActivity(ctx context.Context, userID, action, resourceType, resourceID string, details map[string]interface{}, ipAddress, userAgent string) error

	// Health check
	Health(ctx context.Context) error
}

// authService implements the AuthService interface
type authService struct {
	repoManager         domain.RepositoryManager
	tokenService        TokenService
	config              *config.Config
	notificationService NotificationService
	rateLimitService    RateLimitService
	securityService     SecurityService
	activityService     ActivityService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	repoManager domain.RepositoryManager,
	tokenService TokenService,
	config *config.Config,
	notificationService NotificationService,
	rateLimitService RateLimitService,
	securityService SecurityService,
	activityService ActivityService,
) AuthService {
	return &authService{
		repoManager:         repoManager,
		tokenService:        tokenService,
		config:              config,
		notificationService: notificationService,
		rateLimitService:    rateLimitService,
		securityService:     securityService,
		activityService:     activityService,
	}
}

// Login authenticates a user and returns tokens
func (s *authService) Login(ctx context.Context, req *domain.LoginRequest, ipAddress, userAgent string) (*domain.AuthResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check rate limiting
	if allowed, err := s.rateLimitService.CheckRateLimit(ctx, req.Username, 5, 15*time.Minute); err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	} else if !allowed {
		// Record failed attempt
		if err := s.securityService.RecordLoginAttempt(ctx, req.Username, false, ipAddress, userAgent); err != nil {
			fmt.Printf("Failed to record login attempt: %v\n", err)
		}
		return nil, domain.NewAuthError(domain.AuthErrTooManyAttempts, "Too many login attempts. Please try again later.")
	}

	// Get user by username or email
	userRepo := s.repoManager.GetUserRepository()
	var user *sharedDomain.User
	var err error

	if utils.ValidateEmail(req.Username) {
		user, err = userRepo.GetByEmail(ctx, req.Username)
	} else {
		user, err = userRepo.GetByUsername(ctx, req.Username)
	}

	if err != nil {
		// Record failed attempt
		if err := s.securityService.RecordLoginAttempt(ctx, req.Username, false, ipAddress, userAgent); err != nil {
			fmt.Printf("Failed to record login attempt: %v\n", err)
		}
		return nil, domain.NewAuthError(domain.AuthErrInvalidCredentials, "Invalid username or password")
	}

	// Check user status
	if user.Status != sharedDomain.UserStatusActive {
		// Record failed attempt
		if err := s.securityService.RecordLoginAttempt(ctx, req.Username, false, ipAddress, userAgent); err != nil {
			fmt.Printf("Failed to record login attempt: %v\n", err)
		}

		switch user.Status {
		case sharedDomain.UserStatusInactive:
			return nil, domain.NewAuthError(domain.AuthErrAccountInactive, "Account is inactive")
		case sharedDomain.UserStatusSuspended:
			return nil, domain.NewAuthError(domain.AuthErrAccountLocked, "Account is suspended")
		case sharedDomain.UserStatusDeleted:
			return nil, domain.NewAuthError(domain.AuthErrUserNotFound, "User not found")
		default:
			return nil, domain.NewAuthError(domain.AuthErrAccountInactive, "Account is not available")
		}
	}

	// Check if user is locked out
	if locked, err := s.securityService.IsUserLockedOut(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("lockout check failed: %w", err)
	} else if locked {
		return nil, domain.NewAuthError(domain.AuthErrAccountLocked, "Account is temporarily locked")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Record failed attempt
		if err := s.securityService.RecordLoginAttempt(ctx, req.Username, false, ipAddress, userAgent); err != nil {
			fmt.Printf("Failed to record login attempt: %v\n", err)
		}

		// Check if we should lock the user
		failedAttempts, _ := s.repoManager.GetLoginAttemptRepository().CountFailedAttempts(ctx, req.Username, time.Now().Add(-15*time.Minute))
		if failedAttempts >= 4 { // Lock after 5 failed attempts
			if err := s.securityService.LockUser(ctx, user.ID, 30*time.Minute, "Too many failed login attempts"); err != nil {
				fmt.Printf("Failed to lock user: %v\n", err)
			}
		}

		return nil, domain.NewAuthError(domain.AuthErrInvalidCredentials, "Invalid username or password")
	}

	// Generate tokens
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Create session
	session := &sharedDomain.Session{
		ID:         utils.GenerateID(),
		UserID:     user.ID,
		TokenID:    tokens.AccessToken, // This should be the JTI from the token
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
		ExpiresAt:  tokens.ExpiresAt,
		LastUsedAt: time.Now(),
		Status:     sharedDomain.SessionStatusActive,
	}

	sessionRepo := s.repoManager.GetSessionRepository()
	if err := sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("session creation failed: %w", err)
	}

	// Update last login time
	if err := userRepo.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Record successful attempt
	if err := s.securityService.RecordLoginAttempt(ctx, req.Username, true, ipAddress, userAgent); err != nil {
		fmt.Printf("Failed to record successful login attempt: %v\n", err)
	}

	// Reset login attempts counter
	if err := s.repoManager.GetCacheRepository().ResetLoginAttempts(ctx, req.Username); err != nil {
		fmt.Printf("Failed to reset login attempts: %v\n", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, user.ID, "user.login", "user", user.ID, map[string]interface{}{
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}, ipAddress, userAgent); err != nil {
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	// Create auth user response
	authUser := &domain.AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	}

	// Cache user session
	if err := s.repoManager.GetCacheRepository().SetUserSession(ctx, session.ID, authUser, s.config.Security.JWT.AccessTokenTTL); err != nil {
		fmt.Printf("Failed to cache user session: %v\n", err)
	}

	return &domain.AuthResponse{
		User:   authUser,
		Tokens: tokens,
	}, nil
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest, ipAddress, userAgent string) (*domain.AuthResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate email format
	if !utils.ValidateEmail(req.Email) {
		return nil, domain.NewAuthErrorWithField(sharedDomain.ErrorCodeValidation, "Invalid email format", "email")
	}

	// Validate username format
	if !utils.ValidateUsername(req.Username) {
		return nil, domain.NewAuthErrorWithField(sharedDomain.ErrorCodeValidation, "Invalid username format", "username")
	}

	// Validate password strength
	passwordErrors := utils.ValidatePassword(
		req.Password,
		s.config.Security.Password.MinLength,
		s.config.Security.Password.RequireUppercase,
		s.config.Security.Password.RequireLowercase,
		s.config.Security.Password.RequireNumbers,
		s.config.Security.Password.RequireSymbols,
	)
	if len(passwordErrors) > 0 {
		return nil, domain.NewAuthErrorWithField(domain.AuthErrWeakPassword, passwordErrors[0], "password")
	}

	userRepo := s.repoManager.GetUserRepository()

	// Check if username already exists
	if exists, err := userRepo.ExistsByUsername(ctx, req.Username); err != nil {
		return nil, fmt.Errorf("username check failed: %w", err)
	} else if exists {
		return nil, domain.NewAuthErrorWithField(domain.AuthErrUsernameAlreadyExists, "Username already exists", "username")
	}

	// Check if email already exists
	if exists, err := userRepo.ExistsByEmail(ctx, req.Email); err != nil {
		return nil, fmt.Errorf("email check failed: %w", err)
	} else if exists {
		return nil, domain.NewAuthErrorWithField(domain.AuthErrEmailAlreadyExists, "Email already exists", "email")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.Security.Password.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// Create user
	user := &sharedDomain.User{
		ID:           utils.GenerateID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         sharedDomain.UserRoleUser,
		Status:       sharedDomain.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Create user in database
	if err := userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	// Generate tokens
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Create session
	session := &sharedDomain.Session{
		ID:         utils.GenerateID(),
		UserID:     user.ID,
		TokenID:    tokens.AccessToken, // This should be the JTI from the token
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
		ExpiresAt:  tokens.ExpiresAt,
		LastUsedAt: time.Now(),
		Status:     sharedDomain.SessionStatusActive,
	}

	sessionRepo := s.repoManager.GetSessionRepository()
	if err := sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("session creation failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, user.ID, "user.register", "user", user.ID, map[string]interface{}{
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}, ipAddress, userAgent); err != nil {
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	// Send welcome notification
	if s.notificationService != nil {
		go func() {
			if err := s.notificationService.SendWelcomeEmail(context.Background(), user.Email, user.FirstName); err != nil {
				fmt.Printf("Failed to send welcome email: %v\n", err)
			}
		}()
	}

	// Create auth user response
	authUser := &domain.AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	}

	// Cache user session
	if err := s.repoManager.GetCacheRepository().SetUserSession(ctx, session.ID, authUser, s.config.Security.JWT.AccessTokenTTL); err != nil {
		fmt.Printf("Failed to cache user session: %v\n", err)
	}

	return &domain.AuthResponse{
		User:   authUser,
		Tokens: tokens,
	}, nil
}

// Logout invalidates a user session
func (s *authService) Logout(ctx context.Context, tokenID string) error {
	// Revoke the token
	if err := s.RevokeToken(ctx, tokenID, "", "user_logout"); err != nil {
		return fmt.Errorf("token revocation failed: %w", err)
	}

	// Delete session
	sessionRepo := s.repoManager.GetSessionRepository()
	if err := sessionRepo.DeleteByTokenID(ctx, tokenID); err != nil {
		return fmt.Errorf("session deletion failed: %w", err)
	}

	// Remove from cache
	if err := s.repoManager.GetCacheRepository().DeleteUserSession(ctx, tokenID); err != nil {
		fmt.Printf("Failed to delete user session from cache: %v\n", err)
	}

	return nil
}

// LogoutAll invalidates all user sessions
func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	// Revoke all user tokens
	if err := s.repoManager.GetRevokedTokenRepository().RevokeAllUserTokens(ctx, userID, "logout_all"); err != nil {
		return fmt.Errorf("token revocation failed: %w", err)
	}

	// Delete all user sessions
	sessionRepo := s.repoManager.GetSessionRepository()
	if err := sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("session deletion failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, userID, "user.logout_all", "user", userID, nil, "", ""); err != nil {
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	return nil
}

// RefreshToken generates new tokens using a refresh token
func (s *authService) RefreshToken(ctx context.Context, req *domain.RefreshTokenRequest) (*domain.TokenPair, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate refresh token
	claims, err := s.tokenService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, domain.NewAuthError(domain.AuthErrInvalidToken, "Invalid refresh token")
	}

	// Check if token is revoked
	if revoked, err := s.IsTokenRevoked(ctx, claims.TokenID); err != nil {
		return nil, fmt.Errorf("token revocation check failed: %w", err)
	} else if revoked {
		return nil, domain.NewAuthError(domain.AuthErrRevokedToken, "Token has been revoked")
	}

	// Get user
	userRepo := s.repoManager.GetUserRepository()
	user, err := userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.NewAuthError(domain.AuthErrUserNotFound, "User not found")
	}

	// Check user status
	if user.Status != sharedDomain.UserStatusActive {
		return nil, domain.NewAuthError(domain.AuthErrAccountInactive, "Account is not active")
	}

	// Generate new token pair
	tokens, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	// Revoke old refresh token
	if err := s.RevokeToken(ctx, claims.TokenID, user.ID, "token_refresh"); err != nil {
		fmt.Printf("Failed to revoke old refresh token: %v\n", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, user.ID, "token.refresh", "token", claims.TokenID, nil, "", ""); err != nil {
		fmt.Printf("Failed to log activity: %v\n", err)
	}

	return tokens, nil
}

// ValidateToken validates an access token and returns user info
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.AuthUser, error) {
	// Validate token
	claims, err := s.tokenService.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, domain.NewAuthError(domain.AuthErrInvalidToken, "Invalid token")
	}

	// Check if token is revoked
	if revoked, err := s.IsTokenRevoked(ctx, claims.TokenID); err != nil {
		return nil, fmt.Errorf("token revocation check failed: %w", err)
	} else if revoked {
		return nil, domain.NewAuthError(domain.AuthErrRevokedToken, "Token has been revoked")
	}

	// Try to get user from cache first
	cacheRepo := s.repoManager.GetCacheRepository()
	if authUser, err := cacheRepo.GetUserSession(ctx, claims.TokenID); err == nil && authUser != nil {
		return authUser, nil
	}

	// Get user from database
	userRepo := s.repoManager.GetUserRepository()
	user, err := userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.NewAuthError(domain.AuthErrUserNotFound, "User not found")
	}

	// Check user status
	if user.Status != sharedDomain.UserStatusActive {
		return nil, domain.NewAuthError(domain.AuthErrAccountInactive, "Account is not active")
	}

	// Create auth user
	authUser := &domain.AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	}

	// Cache user session
	if err := cacheRepo.SetUserSession(ctx, claims.TokenID, authUser, s.config.Security.JWT.AccessTokenTTL); err != nil {
		fmt.Printf("Failed to cache user session: %v\n", err)
	}

	return authUser, nil
}

// RevokeToken revokes a specific token
func (s *authService) RevokeToken(ctx context.Context, tokenID, userID, reason string) error {
	// Add to revoked tokens
	revokedRepo := s.repoManager.GetRevokedTokenRepository()
	if err := revokedRepo.RevokeToken(ctx, tokenID, userID, reason, time.Now().Add(24*time.Hour)); err != nil {
		return fmt.Errorf("token revocation failed: %w", err)
	}

	// Add to cache for fast lookup
	cacheRepo := s.repoManager.GetCacheRepository()
	if err := cacheRepo.SetRevokedToken(ctx, tokenID, 24*time.Hour); err != nil {
		fmt.Printf("Warning: Failed to cache revoked token: %v\n", err)
	}

	return nil
}

// IsTokenRevoked checks if a token is revoked
func (s *authService) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	// Check cache first
	cacheRepo := s.repoManager.GetCacheRepository()
	if revoked, err := cacheRepo.IsTokenRevoked(ctx, tokenID); err == nil {
		return revoked, nil
	}

	// Check database
	revokedRepo := s.repoManager.GetRevokedTokenRepository()
	return revokedRepo.IsRevoked(ctx, tokenID)
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(ctx context.Context, userID string, req *domain.ChangePasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	userRepo := s.repoManager.GetUserRepository()
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.NewAuthError(domain.AuthErrUserNotFound, "User not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return domain.NewAuthError(domain.AuthErrInvalidCredentials, "Current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.Security.Password.BcryptCost)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Update password
	if err := userRepo.UpdatePassword(ctx, userID, string(newPasswordHash)); err != nil {
		return fmt.Errorf("password update failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, userID, "user.password_changed", "user", userID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log password change activity: %v\n", err)
	}

	return nil
}

// RequestPasswordReset initiates a password reset process
func (s *authService) RequestPasswordReset(ctx context.Context, req *domain.ResetPasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	userRepo := s.repoManager.GetUserRepository()
	user, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Generate reset token
	resetToken := utils.GenerateID()
	passwordResetToken := &domain.PasswordResetToken{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		Used:      false,
	}

	resetRepo := s.repoManager.GetPasswordResetTokenRepository()
	if err := resetRepo.Create(ctx, passwordResetToken); err != nil {
		return fmt.Errorf("reset token creation failed: %w", err)
	}

	// Send reset email
	if s.notificationService != nil {
		go func() {
			if err := s.notificationService.SendPasswordResetEmail(context.Background(), user.Email, resetToken); err != nil {
				fmt.Printf("Warning: Failed to send password reset email: %v\n", err)
			}
		}()
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, user.ID, "user.password_reset_requested", "user", user.ID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log password reset request activity: %v\n", err)
	}

	return nil
}

// ConfirmPasswordReset confirms and completes a password reset
func (s *authService) ConfirmPasswordReset(ctx context.Context, req *domain.ConfirmResetPasswordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	resetRepo := s.repoManager.GetPasswordResetTokenRepository()
	resetToken, err := resetRepo.GetByToken(ctx, req.Token)
	if err != nil {
		return domain.NewAuthError(domain.AuthErrInvalidToken, "Invalid reset token")
	}

	if resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		return domain.NewAuthError(domain.AuthErrExpiredToken, "Reset token has expired")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.Security.Password.BcryptCost)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Update password
	userRepo := s.repoManager.GetUserRepository()
	if err := userRepo.UpdatePassword(ctx, resetToken.UserID, string(newPasswordHash)); err != nil {
		return fmt.Errorf("password update failed: %w", err)
	}

	// Mark token as used
	if err := resetRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return fmt.Errorf("token update failed: %w", err)
	}

	// Revoke all user tokens
	if err := s.LogoutAll(ctx, resetToken.UserID); err != nil {
		fmt.Printf("Warning: Failed to logout all user sessions: %v\n", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, resetToken.UserID, "user.password_reset_completed", "user", resetToken.UserID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log password reset completion activity: %v\n", err)
	}

	return nil
}

// GetUserProfile retrieves user profile information
func (s *authService) GetUserProfile(ctx context.Context, userID string) (*domain.AuthUser, error) {
	userRepo := s.repoManager.GetUserRepository()
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.NewAuthError(domain.AuthErrUserNotFound, "User not found")
	}

	return &domain.AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
	}, nil
}

// UpdateUserProfile updates user profile information
func (s *authService) UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	userRepo := s.repoManager.GetUserRepository()

	// Get current user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Apply updates
	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}

	user.UpdatedAt = time.Now()

	if err := userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("profile update failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, userID, "user.profile_updated", "user", userID, updates, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log profile update activity: %v\n", err)
	}

	return nil
}

// DeactivateUser deactivates a user account
func (s *authService) DeactivateUser(ctx context.Context, userID string) error {
	userRepo := s.repoManager.GetUserRepository()

	// Get current user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update status
	user.Status = sharedDomain.UserStatusInactive
	user.UpdatedAt = time.Now()

	if err := userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("user deactivation failed: %w", err)
	}

	// Revoke all user tokens
	if err := s.LogoutAll(ctx, userID); err != nil {
		fmt.Printf("Warning: Failed to logout all user sessions: %v\n", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, userID, "user.deactivated", "user", userID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log user deactivation activity: %v\n", err)
	}

	return nil
}

// ActivateUser activates a user account
func (s *authService) ActivateUser(ctx context.Context, userID string) error {
	userRepo := s.repoManager.GetUserRepository()

	// Get current user
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update status
	user.Status = sharedDomain.UserStatusActive
	user.UpdatedAt = time.Now()

	if err := userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("user activation failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, userID, "user.activated", "user", userID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log user activation activity: %v\n", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *authService) GetUserSessions(ctx context.Context, userID string) ([]*sharedDomain.Session, error) {
	sessionRepo := s.repoManager.GetSessionRepository()
	return sessionRepo.GetByUserID(ctx, userID)
}

// RevokeSession revokes a specific session
func (s *authService) RevokeSession(ctx context.Context, sessionID string) error {
	sessionRepo := s.repoManager.GetSessionRepository()
	session, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Revoke the token
	if err := s.RevokeToken(ctx, session.TokenID, session.UserID, "session_revoked"); err != nil {
		fmt.Printf("Warning: Failed to revoke token: %v\n", err)
	}

	// Delete session
	if err := sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("session deletion failed: %w", err)
	}

	// Log activity
	if err := s.activityService.LogActivity(ctx, session.UserID, "session.revoked", "session", sessionID, nil, "", ""); err != nil {
		fmt.Printf("Warning: Failed to log session revocation activity: %v\n", err)
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *authService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return s.LogoutAll(ctx, userID)
}

// CheckRateLimit checks if an operation is rate limited
func (s *authService) CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	return s.rateLimitService.CheckRateLimit(ctx, identifier, limit, window)
}

// RecordLoginAttempt records a login attempt
func (s *authService) RecordLoginAttempt(ctx context.Context, identifier string, success bool, ipAddress, userAgent string) error {
	return s.securityService.RecordLoginAttempt(ctx, identifier, success, ipAddress, userAgent)
}

// IsUserLockedOut checks if a user is locked out
func (s *authService) IsUserLockedOut(ctx context.Context, userID string) (bool, error) {
	return s.securityService.IsUserLockedOut(ctx, userID)
}

// LockUser locks a user account
func (s *authService) LockUser(ctx context.Context, userID string, duration time.Duration, reason string) error {
	return s.securityService.LockUser(ctx, userID, duration, reason)
}

// UnlockUser unlocks a user account
func (s *authService) UnlockUser(ctx context.Context, userID string) error {
	return s.securityService.UnlockUser(ctx, userID)
}

// LogActivity logs user activity
func (s *authService) LogActivity(ctx context.Context, userID, action, resourceType, resourceID string, details map[string]interface{}, ipAddress, userAgent string) error {
	return s.activityService.LogActivity(ctx, userID, action, resourceType, resourceID, details, ipAddress, userAgent)
}

// Health checks the health of the authentication service
func (s *authService) Health(ctx context.Context) error {
	return s.repoManager.Health(ctx)
}
