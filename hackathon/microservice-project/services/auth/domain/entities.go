package domain

import (
	"github.com/elotusteam/microservice-project/shared/domain"
	"time"
)

// AuthUser represents an authenticated user
type AuthUser struct {
	ID       string            `json:"id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Role     domain.UserRole   `json:"role"`
	Status   domain.UserStatus `json:"status"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmResetPasswordRequest represents a password reset confirmation
type ConfirmResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User   *AuthUser  `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenID   string `json:"token_id"`
	TokenType string `json:"token_type"`

	// Standard JWT claims
	Issuer    string       `json:"iss,omitempty"`
	Subject   string       `json:"sub,omitempty"`
	Audience  []string     `json:"aud,omitempty"`
	ExpiresAt *NumericDate `json:"exp,omitempty"`
	NotBefore *NumericDate `json:"nbf,omitempty"`
	IssuedAt  *NumericDate `json:"iat,omitempty"`
	ID        string       `json:"jti,omitempty"`
}

// NumericDate represents a JSON numeric date value
type NumericDate struct {
	Time time.Time
}

// NewNumericDate creates a new NumericDate from time.Time
func NewNumericDate(t time.Time) *NumericDate {
	return &NumericDate{Time: t}
}

// Unix returns the Unix timestamp
func (date *NumericDate) Unix() int64 {
	if date == nil {
		return 0
	}
	return date.Time.Unix()
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginAttempt represents a login attempt for rate limiting
type LoginAttempt struct {
	ID         string    `json:"id"`
	Identifier string    `json:"identifier"` // username, email, or IP
	Success    bool      `json:"success"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Timestamp  time.Time `json:"timestamp"`
}

// AuthError represents authentication-specific errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *AuthError) Error() string {
	return e.Message
}

// Common authentication error codes - these are error identifiers, not credentials
// #nosec G101 -- These are error codes, not hardcoded credentials
const (
	// AuthErrInvalidCredentials indicates authentication failed due to wrong credentials
	AuthErrInvalidCredentials    = "INVALID_CREDENTIALS"
	AuthErrUserNotFound          = "USER_NOT_FOUND"
	AuthErrUserAlreadyExists     = "USER_ALREADY_EXISTS"
	AuthErrEmailAlreadyExists    = "EMAIL_ALREADY_EXISTS"
	AuthErrUsernameAlreadyExists = "USERNAME_ALREADY_EXISTS"
	AuthErrInvalidToken          = "INVALID_TOKEN"
	AuthErrExpiredToken          = "EXPIRED_TOKEN"
	AuthErrRevokedToken          = "REVOKED_TOKEN"
	AuthErrInvalidPassword       = "INVALID_PASSWORD"
	AuthErrWeakPassword          = "WEAK_PASSWORD"
	AuthErrAccountLocked         = "ACCOUNT_LOCKED"
	AuthErrAccountInactive       = "ACCOUNT_INACTIVE"
	AuthErrTooManyAttempts       = "TOO_MANY_ATTEMPTS"
	AuthErrPasswordResetRequired = "PASSWORD_RESET_REQUIRED"
	AuthErrEmailNotVerified      = "EMAIL_NOT_VERIFIED"
)

// NewAuthError creates a new authentication error
func NewAuthError(code, message string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
	}
}

// NewAuthErrorWithField creates a new authentication error with field
func NewAuthErrorWithField(code, message, field string) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// Validation helper functions
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Username is required", "username")
	}
	if r.Password == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Password is required", "password")
	}
	return nil
}

func (r *RegisterRequest) Validate() error {
	if r.Username == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Username is required", "username")
	}
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Username must be between 3 and 50 characters", "username")
	}
	if r.Email == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Email is required", "email")
	}
	if r.Password == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Password is required", "password")
	}
	if len(r.Password) < 8 {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Password must be at least 8 characters", "password")
	}
	if r.FirstName == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "First name is required", "first_name")
	}
	if r.LastName == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Last name is required", "last_name")
	}
	return nil
}

func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Current password is required", "current_password")
	}
	if r.NewPassword == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "New password is required", "new_password")
	}
	if len(r.NewPassword) < 8 {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "New password must be at least 8 characters", "new_password")
	}
	if r.CurrentPassword == r.NewPassword {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "New password must be different from current password", "new_password")
	}
	return nil
}

func (r *ResetPasswordRequest) Validate() error {
	if r.Email == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Email is required", "email")
	}
	return nil
}

func (r *ConfirmResetPasswordRequest) Validate() error {
	if r.Token == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Reset token is required", "token")
	}
	if r.NewPassword == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "New password is required", "new_password")
	}
	if len(r.NewPassword) < 8 {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "New password must be at least 8 characters", "new_password")
	}
	return nil
}

func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return NewAuthErrorWithField(domain.ErrorCodeValidation, "Refresh token is required", "refresh_token")
	}
	return nil
}
