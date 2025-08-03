package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/elotusteam/microservice-project/services/auth/domain"
	"github.com/elotusteam/microservice-project/shared/config"
	sharedDomain "github.com/elotusteam/microservice-project/shared/domain"
	"github.com/elotusteam/microservice-project/shared/utils"
)

// passwordService implements PasswordService interface
type passwordService struct {
	config                    *config.Config
	userRepo                  domain.UserRepository
	passwordResetTokenRepo    domain.PasswordResetTokenRepository
	activityLogRepo           domain.ActivityLogRepository
	notificationService       NotificationService
}

// NewPasswordService creates a new password service
func NewPasswordService(
	config *config.Config,
	userRepo domain.UserRepository,
	passwordResetTokenRepo domain.PasswordResetTokenRepository,
	activityLogRepo domain.ActivityLogRepository,
	notificationService NotificationService,
) PasswordService {
	return &passwordService{
		config:                    config,
		userRepo:                  userRepo,
		passwordResetTokenRepo:    passwordResetTokenRepo,
		activityLogRepo:           activityLogRepo,
		notificationService:       notificationService,
	}
}

// HashPassword hashes a password using bcrypt
func (s *passwordService) HashPassword(password string) (string, error) {
	if len(password) < s.config.Security.Password.MinLength {
		return "", fmt.Errorf("password too short: minimum %d characters required", s.config.Security.Password.MinLength)
	}

	cost := s.config.Security.Password.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword validates a password against its hash
func (s *passwordService) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePasswordStrength validates password strength requirements
func (s *passwordService) ValidatePasswordStrength(password string) []string {
	var errors []string

	if len(password) < s.config.Security.Password.MinLength {
		errors = append(errors, fmt.Sprintf("password too short: minimum %d characters required", s.config.Security.Password.MinLength))
	}

	if s.config.Security.Password.RequireUppercase && !utils.ContainsUppercase(password) {
		errors = append(errors, "password must contain at least one uppercase letter")
	}

	if s.config.Security.Password.RequireLowercase && !utils.ContainsLowercase(password) {
		errors = append(errors, "password must contain at least one lowercase letter")
	}

	if s.config.Security.Password.RequireNumbers && !utils.ContainsNumber(password) {
		errors = append(errors, "password must contain at least one number")
	}

	if s.config.Security.Password.RequireSpecialChars && !utils.ContainsSpecialChar(password) {
		errors = append(errors, "password must contain at least one special character")
	}

	return errors
}

// CheckPasswordPolicy validates password against policy
func (s *passwordService) CheckPasswordPolicy(password string, policy *PasswordPolicy) []string {
	var errors []string

	if len(password) < policy.MinLength {
		errors = append(errors, fmt.Sprintf("password too short: minimum %d characters required", policy.MinLength))
	}

	if policy.MaxLength > 0 && len(password) > policy.MaxLength {
		errors = append(errors, fmt.Sprintf("password too long: maximum %d characters allowed", policy.MaxLength))
	}

	if policy.RequireUppercase && !utils.ContainsUppercase(password) {
		errors = append(errors, "password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !utils.ContainsLowercase(password) {
		errors = append(errors, "password must contain at least one lowercase letter")
	}

	if policy.RequireNumbers && !utils.ContainsNumber(password) {
		errors = append(errors, "password must contain at least one number")
	}

	if policy.RequireSymbols && !utils.ContainsSpecialChar(password) {
		errors = append(errors, "password must contain at least one special character")
	}

	return errors
}

// GenerateResetToken generates a password reset token
func (s *passwordService) GenerateResetToken(userID string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("token generation failed: %w", err)
	}
	tokenString := hex.EncodeToString(tokenBytes)

	// Create reset token
	resetToken := &domain.PasswordResetToken{
		ID:        utils.GenerateID(),
		UserID:    userID,
		Token:     tokenString,
		Used:      false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.Security.Password.ResetTokenTTL),
	}

	// Save token
	ctx := context.Background()
	if err := s.passwordResetTokenRepo.Create(ctx, resetToken); err != nil {
		return "", fmt.Errorf("failed to save reset token: %w", err)
	}

	// Log activity
	s.LogActivity(ctx, userID, "password_reset_requested", "user", userID, "Password reset token generated", utils.GetIPFromContext(ctx), utils.GetUserAgentFromContext(ctx))

	return tokenString, nil
}

// ValidateResetToken validates a password reset token
func (s *passwordService) ValidateResetToken(token string) (string, error) {
	ctx := context.Background()
	resetToken, err := s.passwordResetTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("invalid reset token: %w", err)
	}

	// Check if token is used
	if resetToken.Used {
		return "", fmt.Errorf("reset token already used")
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return "", fmt.Errorf("reset token expired")
	}

	return resetToken.UserID, nil
}

// InvalidateResetToken invalidates a password reset token
func (s *passwordService) InvalidateResetToken(token string) error {
	ctx := context.Background()
	resetToken, err := s.passwordResetTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid reset token: %w", err)
	}

	// Mark token as used
	return s.passwordResetTokenRepo.MarkAsUsed(ctx, resetToken.ID)
}

// ResetPassword resets a user's password using a reset token
func (s *passwordService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate reset token
	userID, err := s.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// Validate new password strength
	if errors := s.ValidatePasswordStrength(newPassword); len(errors) > 0 {
		return fmt.Errorf("password validation failed: %v", errors)
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update user password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate token
	if err := s.InvalidateResetToken(token); err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	// Log activity
	s.LogActivity(ctx, user.ID, "password_reset_completed", "user", user.ID, "Password reset successfully", utils.GetIPFromContext(ctx), utils.GetUserAgentFromContext(ctx))

	// Send notification
	if err := s.notificationService.SendPasswordChangedEmail(ctx, user.Email, user.FirstName); err != nil {
		// Log error but don't fail the operation
		s.LogActivity(ctx, user.ID, "notification_failed", "user", user.ID, fmt.Sprintf("Failed to send password reset notification: %v", err), utils.GetIPFromContext(ctx), utils.GetUserAgentFromContext(ctx))
	}

	return nil
}

// ChangePassword changes a user's password
func (s *passwordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate current password
	if err := s.VerifyPassword(currentPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	if errors := s.ValidatePasswordStrength(newPassword); len(errors) > 0 {
		return fmt.Errorf("password validation failed: %v", errors)
	}

	// Check if new password is different from current
	if err := s.VerifyPassword(newPassword, user.PasswordHash); err == nil {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log activity
	s.LogActivity(ctx, userID, "password_changed", "user", userID, "Password changed successfully", utils.GetIPFromContext(ctx), utils.GetUserAgentFromContext(ctx))

	// Send notification
	if err := s.notificationService.SendPasswordChangedEmail(ctx, user.Email, user.FirstName); err != nil {
		// Log error but don't fail the operation
		s.LogActivity(ctx, userID, "notification_failed", "user", userID, fmt.Sprintf("Failed to send password change notification: %v", err), utils.GetIPFromContext(ctx), utils.GetUserAgentFromContext(ctx))
	}

	return nil
}

// CheckPasswordHistory checks if password was used recently
func (s *passwordService) CheckPasswordHistory(userID, newPasswordHash string, historyCount int) (bool, error) {
	// This would require a password history table
	// For now, we'll just check against current password
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Check if new password hash matches current password hash
	if user.PasswordHash == newPasswordHash {
		return true, nil // Password was used recently
	}

	return false, nil // Password not found in history
}

// AddPasswordToHistory adds password to history
func (s *passwordService) AddPasswordToHistory(userID, passwordHash string) error {
	// This would require a password history table
	// For now, this is a no-op since we don't have password history storage
	return nil
}

// CleanupExpiredTokens removes expired password reset tokens
func (s *passwordService) CleanupExpiredTokens(ctx context.Context) error {
	return s.passwordResetTokenRepo.DeleteExpired(ctx)
}

// LogActivity logs an activity
func (s *passwordService) LogActivity(ctx context.Context, userID, action, resourceType, resourceID, details, ipAddress, userAgent string) {
	activityLog := &sharedDomain.ActivityLog{
		ID:           utils.GenerateID(),
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   &resourceID,
		Details:      map[string]interface{}{"description": details},
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Timestamp:    time.Now(),
		Status:       sharedDomain.ActivityStatusSuccess,
	}

	if err := s.activityLogRepo.Create(ctx, activityLog); err != nil {
		fmt.Printf("Failed to log activity: %v\n", err)
	}
}