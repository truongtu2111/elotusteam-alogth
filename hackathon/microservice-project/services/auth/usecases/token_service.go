package usecases

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elotusteam/microservice-project/services/auth/domain"
	"github.com/elotusteam/microservice-project/shared/config"
	sharedDomain "github.com/elotusteam/microservice-project/shared/domain"
	"github.com/elotusteam/microservice-project/shared/utils"
)

// tokenService implements the TokenService interface
type tokenService struct {
	config *config.Config
}

// NewTokenService creates a new token service
func NewTokenService(config *config.Config) TokenService {
	return &tokenService{
		config: config,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (s *tokenService) GenerateTokenPair(ctx context.Context, user *sharedDomain.User) (*domain.TokenPair, error) {
	// Generate access token
	accessToken, err := s.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("access token generation failed: %w", err)
	}
	
	// Generate refresh token
	refreshToken, err := s.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("refresh token generation failed: %w", err)
	}
	
	// Calculate expiration times
	accessExpiresAt := time.Now().Add(s.config.Security.JWT.AccessTokenTTL)
	refreshExpiresAt := time.Now().Add(s.config.Security.JWT.RefreshTokenTTL)
	
	return &domain.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// GenerateAccessToken generates an access token for a user
func (s *tokenService) GenerateAccessToken(ctx context.Context, user *sharedDomain.User) (string, error) {
	now := time.Now()
	tokenID := utils.GenerateID()
	
	claims := &domain.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		TokenID:   tokenID,
		TokenType: "access",
		ID:        tokenID,
		Subject:   user.ID,
		Audience:  []string{s.config.Security.JWT.Audience},
		Issuer:    s.config.Security.JWT.Issuer,
		IssuedAt:  domain.NewNumericDate(now),
		NotBefore: domain.NewNumericDate(now),
		ExpiresAt: domain.NewNumericDate(now.Add(s.config.Security.JWT.AccessTokenTTL)),
	}
	
	return s.signToken(claims)
}

// GenerateRefreshToken generates a refresh token for a user
func (s *tokenService) GenerateRefreshToken(ctx context.Context, user *sharedDomain.User) (string, error) {
	now := time.Now()
	tokenID := utils.GenerateID()
	
	claims := &domain.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		TokenID:   tokenID,
		TokenType: "refresh",
		ID:        tokenID,
		Subject:   user.ID,
		Audience:  []string{s.config.Security.JWT.Audience},
		Issuer:    s.config.Security.JWT.Issuer,
		IssuedAt:  domain.NewNumericDate(now),
		NotBefore: domain.NewNumericDate(now),
		ExpiresAt: domain.NewNumericDate(now.Add(s.config.Security.JWT.RefreshTokenTTL)),
	}
	
	return s.signToken(claims)
}

// ValidateAccessToken validates an access token and returns claims
func (s *tokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	claims, err := s.parseAndValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}
	
	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns claims
func (s *tokenService) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	claims, err := s.parseAndValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}
	
	return claims, nil
}

// ValidateToken validates any token and returns claims
func (s *tokenService) ValidateToken(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	return s.parseAndValidateToken(tokenString)
}

// ParseToken parses a token without validation
func (s *tokenService) ParseToken(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	return s.parseToken(tokenString)
}

// ExtractTokenID extracts the token ID from a token
func (s *tokenService) ExtractTokenID(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	
	return claims.TokenID, nil
}

// ExtractUserID extracts the user ID from a token
func (s *tokenService) ExtractUserID(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	
	return claims.UserID, nil
}

// GetTokenExpiration gets the expiration time of a token
func (s *tokenService) GetTokenExpiration(ctx context.Context, tokenString string) (time.Time, error) {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return time.Time{}, err
	}
	
	if claims.ExpiresAt == nil {
		return time.Time{}, fmt.Errorf("token has no expiration")
	}
	
	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (s *tokenService) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	expiresAt, err := s.GetTokenExpiration(ctx, tokenString)
	if err != nil {
		return true, err
	}
	
	return time.Now().After(expiresAt), nil
}

// GetTokenType gets the type of a token
func (s *tokenService) GetTokenType(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return "", err
	}
	
	return claims.TokenType, nil
}

// signToken signs a JWT token with HMAC-SHA256
func (s *tokenService) signToken(claims *domain.JWTClaims) (string, error) {
	// Create header
	header := map[string]interface{}{
		"typ": "JWT",
		"alg": "HS256",
	}
	
	// Encode header
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("header encoding failed: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	
	// Encode payload
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("payload encoding failed: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	// Create signature
	message := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, []byte(s.config.Security.JWT.SecretKey))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	
	return message + "." + signature, nil
}

// parseToken parses a JWT token
func (s *tokenService) parseToken(tokenString string) (*domain.JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	
	// Verify signature
	message := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, []byte(s.config.Security.JWT.SecretKey))
	h.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	
	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}
	
	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("payload decoding failed: %w", err)
	}
	
	var claims domain.JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("payload unmarshaling failed: %w", err)
	}
	
	return &claims, nil
}

// parseAndValidateToken is a helper method to parse and validate tokens
func (s *tokenService) parseAndValidateToken(tokenString string) (*domain.JWTClaims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	// Additional validation
	if err := s.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	
	return claims, nil
}

// validateClaims validates JWT claims
func (s *tokenService) validateClaims(claims *domain.JWTClaims) error {
	now := time.Now()
	
	// Check expiration
	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time) {
		return fmt.Errorf("token has expired")
	}
	
	// Check not before
	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time) {
		return fmt.Errorf("token not yet valid")
	}
	
	// Check issuer
	if claims.Issuer != s.config.Security.JWT.Issuer {
		return fmt.Errorf("invalid issuer")
	}
	
	// Check audience
	if len(claims.Audience) > 0 {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == s.config.Security.JWT.Audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience")
		}
	}
	
	// Check required fields
	if claims.UserID == "" {
		return fmt.Errorf("missing user ID")
	}
	
	if claims.TokenID == "" {
		return fmt.Errorf("missing token ID")
	}
	
	if claims.TokenType == "" {
		return fmt.Errorf("missing token type")
	}
	
	return nil
}